package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/redis/go-redis/v9"
)

type TwoLevelCache struct {
	localCache  *bigcache.BigCache
	redisClient *redis.Client
}

func NewTwoLevelCache(localConfig *bigcache.Config, redisOptions *redis.Options) (*TwoLevelCache, error) {
	localCache, err := bigcache.New(context.Background(), *localConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create local cache: %w", err)
	}

	var redisClient *redis.Client
	if redisOptions != nil {
		redisClient = redis.NewClient(redisOptions)
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("failed to connect to redis: %w", err)
		}
	}

	return &TwoLevelCache{
		localCache:  localCache,
		redisClient: redisClient,
	}, nil
}

func (c *TwoLevelCache) GenerateCacheKey(modelID, version string, inputs map[string]interface{}) string {
	hash := sha256.New()
	sortedInputs := sortMap(inputs)
	jsonData, _ := json.Marshal(sortedInputs)
	hash.Write(jsonData)
	featureHash := hex.EncodeToString(hash.Sum(nil))
	return fmt.Sprintf("ai:predict:%s:%s:%s", modelID, version, featureHash)
}

func sortMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		result[k] = m[k]
	}
	return result
}

func (c *TwoLevelCache) Get(ctx context.Context, key string) (*PredictResponse, error) {
	data, err := c.localCache.Get(key)
	if err == nil {
		var resp PredictResponse
		if err := json.Unmarshal(data, &resp); err == nil {
			return &resp, nil
		}
	}

	if c.redisClient != nil {
		data, err := c.redisClient.Get(ctx, key).Bytes()
		if err == nil {
			var resp PredictResponse
			if err := json.Unmarshal(data, &resp); err == nil {
				c.localCache.Set(key, data)
				return &resp, nil
			}
		}
	}

	return nil, fmt.Errorf("cache miss")
}

func (c *TwoLevelCache) Set(ctx context.Context, key string, resp *PredictResponse, ttl time.Duration) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	if err := c.localCache.Set(key, data); err != nil {
		return err
	}

	if c.redisClient != nil {
		if err := c.redisClient.Set(ctx, key, data, ttl).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (c *TwoLevelCache) Delete(ctx context.Context, key string) error {
	c.localCache.Delete(key)
	if c.redisClient != nil {
		return c.redisClient.Del(ctx, key).Err()
	}
	return nil
}

func (c *TwoLevelCache) DeleteByPattern(ctx context.Context, pattern string) error {
	if c.redisClient != nil {
		iter := c.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			c.redisClient.Del(ctx, iter.Val())
		}
		return iter.Err()
	}
	return nil
}

func (c *TwoLevelCache) Close() error {
	if c.redisClient != nil {
		c.redisClient.Close()
	}
	return c.localCache.Close()
}

type SimpleCache struct {
	cache map[string]*cacheItem
}

type cacheItem struct {
	value      *PredictResponse
	expiration time.Time
}

func NewSimpleCache() *SimpleCache {
	return &SimpleCache{
		cache: make(map[string]*cacheItem),
	}
}

func (c *SimpleCache) Get(ctx context.Context, key string) (*PredictResponse, error) {
	item, ok := c.cache[key]
	if !ok {
		return nil, fmt.Errorf("cache miss")
	}
	if time.Now().After(item.expiration) {
		delete(c.cache, key)
		return nil, fmt.Errorf("cache expired")
	}
	return item.value, nil
}

func (c *SimpleCache) Set(ctx context.Context, key string, resp *PredictResponse, ttl time.Duration) error {
	c.cache[key] = &cacheItem{
		value:      resp,
		expiration: time.Now().Add(ttl),
	}
	return nil
}

func (c *SimpleCache) Delete(ctx context.Context, key string) error {
	delete(c.cache, key)
	return nil
}

func (c *SimpleCache) DeleteByPattern(ctx context.Context, pattern string) error {
	for key := range c.cache {
		if matchPattern(key, pattern) {
			delete(c.cache, key)
		}
	}
	return nil
}

func matchPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}
	return len(key) >= len(pattern) && key[:len(pattern)] == pattern
}
