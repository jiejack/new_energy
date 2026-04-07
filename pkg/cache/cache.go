package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	cfg    *Config
}

type Config struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
}

func NewCache(cfg *Config) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Cache{
		client: client,
		cfg:    cfg,
	}, nil
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, data, expiration).Err()
}

func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func (c *Cache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

func (c *Cache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

func (c *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

func (c *Cache) GetMulti(ctx context.Context, keys []string) ([]interface{}, error) {
	results, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(results))
	for i, result := range results {
		if result == nil {
			values[i] = nil
			continue
		}

		var dest interface{}
		if err := json.Unmarshal([]byte(result.(string)), &dest); err != nil {
			return nil, err
		}
		values[i] = dest
	}

	return values, nil
}

func (c *Cache) SetMulti(ctx context.Context, items map[string]interface{}, expiration time.Duration) error {
	pipe := c.client.Pipeline()

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		pipe.Set(ctx, key, data, expiration)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (c *Cache) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *Cache) Decrement(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

func (c *Cache) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

func (c *Cache) HashSet(ctx context.Context, key string, values map[string]interface{}) error {
	pipe := c.client.Pipeline()

	for field, value := range values {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for field %s: %w", field, err)
		}
		pipe.HSet(ctx, key, field, data)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (c *Cache) HashGet(ctx context.Context, key, field string, dest interface{}) error {
	data, err := c.client.HGet(ctx, key, field).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

func (c *Cache) HashGetAll(ctx context.Context, key string, dest map[string]interface{}) error {
	results, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return err
	}

	for field, value := range results {
		var valueDest interface{}
		if err := json.Unmarshal([]byte(value), &valueDest); err != nil {
			return err
		}
		dest[field] = valueDest
	}

	return nil
}

func (c *Cache) HashDelete(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

func (c *Cache) HashExists(ctx context.Context, key, field string) (bool, error) {
	return c.client.HExists(ctx, key, field).Result()
}

func (c *Cache) HashLen(ctx context.Context, key string) (int64, error) {
	return c.client.HLen(ctx, key).Result()
}

func (c *Cache) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.client.Publish(ctx, channel, data).Err()
}

func (c *Cache) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return c.client.Subscribe(ctx, channel)
}
