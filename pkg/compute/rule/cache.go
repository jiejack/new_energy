package rule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ErrCacheNotFound = errors.New("cache not found")
	ErrCacheExpired  = errors.New("cache expired")
)

// Prometheus指标
var (
	cacheHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compute_cache_hits_total",
		Help: "Total number of cache hits",
	}, []string{"cache_type"})

	cacheMissesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compute_cache_misses_total",
		Help: "Total number of cache misses",
	}, []string{"cache_type"})

	cacheSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "compute_cache_size",
		Help: "Current cache size",
	}, []string{"cache_type"})

	cacheEvictionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compute_cache_evictions_total",
		Help: "Total number of cache evictions",
	}, []string{"cache_type", "reason"})
)

// CachePolicy 缓存策略
type CachePolicy string

const (
	CachePolicyLRU     CachePolicy = "lru"     // 最近最少使用
	CachePolicyLFU     CachePolicy = "lfu"     // 最不经常使用
	CachePolicyFIFO    CachePolicy = "fifo"    // 先进先出
	CachePolicyTTL     CachePolicy = "ttl"     // 基于时间过期
	CachePolicyNone    CachePolicy = "none"    // 不缓存
)

// ComputeCache 计算结果缓存结构
type ComputeCache struct {
	localCache  *LocalCache
	redisCache  *RedisCache
	config      *CacheConfig
	metrics     *CacheMetrics
	mu          sync.RWMutex
}

// CacheConfig 缓存配置
type CacheConfig struct {
	EnableLocalCache  bool          `json:"enableLocalCache"`  // 启用本地缓存
	EnableRedisCache  bool          `json:"enableRedisCache"`  // 启用Redis缓存
	LocalCacheSize    int           `json:"localCacheSize"`    // 本地缓存大小
	DefaultTTL        time.Duration `json:"defaultTTL"`        // 默认过期时间
	CleanupInterval   time.Duration `json:"cleanupInterval"`   // 清理间隔
	Policy            CachePolicy   `json:"policy"`            // 缓存策略
	EnableMetrics     bool          `json:"enableMetrics"`     // 启用指标
	RedisKeyPrefix    string        `json:"redisKeyPrefix"`    // Redis键前缀
}

// CacheMetrics 缓存指标
type CacheMetrics struct {
	Hits         int64 `json:"hits"`         // 命中次数
	Misses       int64 `json:"misses"`       // 未命中次数
	Evictions    int64 `json:"evictions"`    // 驱逐次数
	Size         int64 `json:"size"`         // 当前大小
	LocalHits    int64 `json:"localHits"`    // 本地缓存命中
	RedisHits    int64 `json:"redisHits"`    // Redis缓存命中
	TotalQueries int64 `json:"totalQueries"` // 总查询次数
}

// ComputeResult 计算结果
type ComputeResult struct {
	PointID    string                 `json:"pointId"`
	Value      float64                `json:"value"`
	Quality    int                    `json:"quality"`
	Timestamp  time.Time              `json:"timestamp"`
	ComputeTime time.Duration         `json:"computeTime"`
	Formula    string                 `json:"formula"`
	Inputs     map[string]float64     `json:"inputs"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// NewComputeCache 创建计算结果缓存
func NewComputeCache(config *CacheConfig, redisClient *redis.Client) *ComputeCache {
	if config == nil {
		config = &CacheConfig{
			EnableLocalCache: true,
			EnableRedisCache: false,
			LocalCacheSize:   10000,
			DefaultTTL:       5 * time.Minute,
			CleanupInterval:  1 * time.Minute,
			Policy:           CachePolicyLRU,
			EnableMetrics:    true,
			RedisKeyPrefix:   "compute:cache:",
		}
	}

	cache := &ComputeCache{
		config:  config,
		metrics: &CacheMetrics{},
	}

	// 创建本地缓存
	if config.EnableLocalCache {
		cache.localCache = NewLocalCache(config.LocalCacheSize, config.DefaultTTL, config.Policy)
	}

	// 创建Redis缓存
	if config.EnableRedisCache && redisClient != nil {
		cache.redisCache = NewRedisCache(redisClient, config.RedisKeyPrefix, config.DefaultTTL)
	}

	return cache
}

// Set 设置缓存
func (c *ComputeCache) Set(ctx context.Context, key string, result *ComputeResult) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 设置本地缓存
	if c.localCache != nil {
		c.localCache.Set(key, result)
	}

	// 设置Redis缓存
	if c.redisCache != nil {
		if err := c.redisCache.Set(ctx, key, result); err != nil {
			return fmt.Errorf("failed to set redis cache: %w", err)
		}
	}

	// 更新指标
	atomic.AddInt64(&c.metrics.Size, 1)
	if c.config.EnableMetrics {
		cacheSize.WithLabelValues("total").Set(float64(atomic.LoadInt64(&c.metrics.Size)))
	}

	return nil
}

// Get 获取缓存
func (c *ComputeCache) Get(ctx context.Context, key string) (*ComputeResult, error) {
	atomic.AddInt64(&c.metrics.TotalQueries, 1)

	// 先查本地缓存
	if c.localCache != nil {
		if result, exists := c.localCache.Get(key); exists {
			atomic.AddInt64(&c.metrics.Hits, 1)
			atomic.AddInt64(&c.metrics.LocalHits, 1)
			if c.config.EnableMetrics {
				cacheHitsTotal.WithLabelValues("local").Inc()
			}
			return result, nil
		}
	}

	// 再查Redis缓存
	if c.redisCache != nil {
		result, err := c.redisCache.Get(ctx, key)
		if err == nil {
			atomic.AddInt64(&c.metrics.Hits, 1)
			atomic.AddInt64(&c.metrics.RedisHits, 1)
			if c.config.EnableMetrics {
				cacheHitsTotal.WithLabelValues("redis").Inc()
			}

			// 回填本地缓存
			if c.localCache != nil {
				c.localCache.Set(key, result)
			}

			return result, nil
		}
	}

	// 缓存未命中
	atomic.AddInt64(&c.metrics.Misses, 1)
	if c.config.EnableMetrics {
		cacheMissesTotal.WithLabelValues("total").Inc()
	}

	return nil, ErrCacheNotFound
}

// Delete 删除缓存
func (c *ComputeCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 删除本地缓存
	if c.localCache != nil {
		c.localCache.Delete(key)
	}

	// 删除Redis缓存
	if c.redisCache != nil {
		if err := c.redisCache.Delete(ctx, key); err != nil {
			return fmt.Errorf("failed to delete redis cache: %w", err)
		}
	}

	atomic.AddInt64(&c.metrics.Size, -1)
	atomic.AddInt64(&c.metrics.Evictions, 1)

	if c.config.EnableMetrics {
		cacheEvictionsTotal.WithLabelValues("total", "delete").Inc()
	}

	return nil
}

// Clear 清空缓存
func (c *ComputeCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 清空本地缓存
	if c.localCache != nil {
		c.localCache.Clear()
	}

	// 清空Redis缓存
	if c.redisCache != nil {
		if err := c.redisCache.Clear(ctx); err != nil {
			return fmt.Errorf("failed to clear redis cache: %w", err)
		}
	}

	atomic.StoreInt64(&c.metrics.Size, 0)

	return nil
}

// Invalidate 失效缓存
func (c *ComputeCache) Invalidate(ctx context.Context, pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 失效本地缓存
	if c.localCache != nil {
		c.localCache.Invalidate(pattern)
	}

	// 失效Redis缓存
	if c.redisCache != nil {
		if err := c.redisCache.Invalidate(ctx, pattern); err != nil {
			return fmt.Errorf("failed to invalidate redis cache: %w", err)
		}
	}

	return nil
}

// InvalidateByPoint 根据计算点失效缓存
func (c *ComputeCache) InvalidateByPoint(ctx context.Context, pointID string) error {
	pattern := fmt.Sprintf("*%s*", pointID)
	return c.Invalidate(ctx, pattern)
}

// InvalidateByDependency 根据依赖关系失效缓存
func (c *ComputeCache) InvalidateByDependency(ctx context.Context, pointIDs []string) error {
	for _, pointID := range pointIDs {
		if err := c.InvalidateByPoint(ctx, pointID); err != nil {
			return err
		}
	}
	return nil
}

// GetMetrics 获取缓存指标
func (c *ComputeCache) GetMetrics() *CacheMetrics {
	return &CacheMetrics{
		Hits:         atomic.LoadInt64(&c.metrics.Hits),
		Misses:       atomic.LoadInt64(&c.metrics.Misses),
		Evictions:    atomic.LoadInt64(&c.metrics.Evictions),
		Size:         atomic.LoadInt64(&c.metrics.Size),
		LocalHits:    atomic.LoadInt64(&c.metrics.LocalHits),
		RedisHits:    atomic.LoadInt64(&c.metrics.RedisHits),
		TotalQueries: atomic.LoadInt64(&c.metrics.TotalQueries),
	}
}

// GetHitRate 获取命中率
func (c *ComputeCache) GetHitRate() float64 {
	hits := atomic.LoadInt64(&c.metrics.Hits)
	total := atomic.LoadInt64(&c.metrics.TotalQueries)
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total)
}

// LocalCache 本地内存缓存
type LocalCache struct {
	items    map[string]*cacheItem
	size     int
	maxSize  int
	ttl      time.Duration
	policy   CachePolicy
	mu       sync.RWMutex
}

type cacheItem struct {
	value      *ComputeResult
	expiration time.Time
	accessTime time.Time
	accessCount int64
}

// NewLocalCache 创建本地缓存
func NewLocalCache(maxSize int, ttl time.Duration, policy CachePolicy) *LocalCache {
	cache := &LocalCache{
		items:   make(map[string]*cacheItem),
		maxSize: maxSize,
		ttl:     ttl,
		policy:  policy,
	}

	// 启动清理协程
	go cache.cleanup()

	return cache
}

// Set 设置缓存
func (lc *LocalCache) Set(key string, value *ComputeResult) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// 检查是否需要驱逐
	if len(lc.items) >= lc.maxSize {
		lc.evict()
	}

	lc.items[key] = &cacheItem{
		value:       value,
		expiration:  time.Now().Add(lc.ttl),
		accessTime:  time.Now(),
		accessCount: 0,
	}
	lc.size = len(lc.items)
}

// Get 获取缓存
func (lc *LocalCache) Get(key string) (*ComputeResult, bool) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	item, exists := lc.items[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.expiration) {
		return nil, false
	}

	// 更新访问信息
	item.accessTime = time.Now()
	item.accessCount++

	return item.value, true
}

// Delete 删除缓存
func (lc *LocalCache) Delete(key string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	delete(lc.items, key)
	lc.size = len(lc.items)
}

// Clear 清空缓存
func (lc *LocalCache) Clear() {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.items = make(map[string]*cacheItem)
	lc.size = 0
}

// Invalidate 失效缓存
func (lc *LocalCache) Invalidate(pattern string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// 简单实现：删除包含pattern的所有key
	for key := range lc.items {
		if containsPattern(key, pattern) {
			delete(lc.items, key)
		}
	}
	lc.size = len(lc.items)
}

// evict 驱逐缓存
func (lc *LocalCache) evict() {
	if len(lc.items) == 0 {
		return
	}

	var evictKey string

	switch lc.policy {
	case CachePolicyLRU:
		// 驱逐最近最少使用的
		var oldest time.Time
		for key, item := range lc.items {
			if oldest.IsZero() || item.accessTime.Before(oldest) {
				oldest = item.accessTime
				evictKey = key
			}
		}

	case CachePolicyLFU:
		// 驱逐最不经常使用的
		var minCount int64 = -1
		for key, item := range lc.items {
			if minCount == -1 || item.accessCount < minCount {
				minCount = item.accessCount
				evictKey = key
			}
		}

	case CachePolicyFIFO:
		// 驱逐最早进入的
		for key := range lc.items {
			evictKey = key
			break
		}

	default:
		// 默认LRU
		var oldest time.Time
		for key, item := range lc.items {
			if oldest.IsZero() || item.accessTime.Before(oldest) {
				oldest = item.accessTime
				evictKey = key
			}
		}
	}

	if evictKey != "" {
		delete(lc.items, evictKey)
		lc.size = len(lc.items)
		cacheEvictionsTotal.WithLabelValues("local", "policy").Inc()
	}
}

// cleanup 定期清理过期缓存
func (lc *LocalCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		lc.mu.Lock()
		now := time.Now()
		for key, item := range lc.items {
			if now.After(item.expiration) {
				delete(lc.items, key)
				cacheEvictionsTotal.WithLabelValues("local", "expired").Inc()
			}
		}
		lc.size = len(lc.items)
		lc.mu.Unlock()
	}
}

// RedisCache Redis缓存
type RedisCache struct {
	client    *redis.Client
	prefix    string
	ttl       time.Duration
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(client *redis.Client, prefix string, ttl time.Duration) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
		ttl:    ttl,
	}
}

// Set 设置缓存
func (rc *RedisCache) Set(ctx context.Context, key string, value *ComputeResult) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	fullKey := rc.prefix + key
	return rc.client.Set(ctx, fullKey, data, rc.ttl).Err()
}

// Get 获取缓存
func (rc *RedisCache) Get(ctx context.Context, key string) (*ComputeResult, error) {
	fullKey := rc.prefix + key
	data, err := rc.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheNotFound
		}
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	var result ComputeResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	return &result, nil
}

// Delete 删除缓存
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := rc.prefix + key
	return rc.client.Del(ctx, fullKey).Err()
}

// Clear 清空缓存
func (rc *RedisCache) Clear(ctx context.Context) error {
	pattern := rc.prefix + "*"
	keys, err := rc.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	if len(keys) > 0 {
		return rc.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Invalidate 失效缓存
func (rc *RedisCache) Invalidate(ctx context.Context, pattern string) error {
	fullPattern := rc.prefix + pattern
	keys, err := rc.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	if len(keys) > 0 {
		return rc.client.Del(ctx, keys...).Err()
	}

	return nil
}

// SetWithTTL 设置缓存（自定义TTL）
func (c *ComputeCache) SetWithTTL(ctx context.Context, key string, result *ComputeResult, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 设置本地缓存
	if c.localCache != nil {
		// 本地缓存使用默认TTL，因为已经有定期清理机制
		c.localCache.Set(key, result)
	}

	// 设置Redis缓存
	if c.redisCache != nil {
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal cache value: %w", err)
		}
		fullKey := c.config.RedisKeyPrefix + key
		if err := c.redisCache.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
			return fmt.Errorf("failed to set redis cache: %w", err)
		}
	}

	atomic.AddInt64(&c.metrics.Size, 1)

	return nil
}

// GetOrCompute 获取缓存或计算
func (c *ComputeCache) GetOrCompute(ctx context.Context, key string, computeFunc func() (*ComputeResult, error)) (*ComputeResult, error) {
	// 先尝试从缓存获取
	result, err := c.Get(ctx, key)
	if err == nil {
		return result, nil
	}

	// 缓存未命中，执行计算
	result, err = computeFunc()
	if err != nil {
		return nil, fmt.Errorf("failed to compute: %w", err)
	}

	// 设置缓存
	if err := c.Set(ctx, key, result); err != nil {
		// 缓存设置失败不影响返回结果
		// 只记录日志
	}

	return result, nil
}

// containsPattern 检查key是否匹配pattern
func containsPattern(key, pattern string) bool {
	// 简单实现：检查是否包含
	// 实际可以使用正则表达式或通配符匹配
	return len(pattern) == 0 ||
		(pattern[0] == '*' && pattern[len(pattern)-1] == '*' && contains(key, pattern[1:len(pattern)-1])) ||
		key == pattern
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BatchGet 批量获取缓存
func (c *ComputeCache) BatchGet(ctx context.Context, keys []string) (map[string]*ComputeResult, error) {
	results := make(map[string]*ComputeResult)
	var missedKeys []string

	for _, key := range keys {
		result, err := c.Get(ctx, key)
		if err == nil {
			results[key] = result
		} else {
			missedKeys = append(missedKeys, key)
		}
	}

	return results, nil
}

// BatchSet 批量设置缓存
func (c *ComputeCache) BatchSet(ctx context.Context, items map[string]*ComputeResult) error {
	for key, result := range items {
		if err := c.Set(ctx, key, result); err != nil {
			return fmt.Errorf("failed to set cache for key %s: %w", key, err)
		}
	}
	return nil
}

// GetSize 获取缓存大小
func (c *ComputeCache) GetSize() int {
	return int(atomic.LoadInt64(&c.metrics.Size))
}

// IsEnabled 检查缓存是否启用
func (c *ComputeCache) IsEnabled() bool {
	return c.localCache != nil || c.redisCache != nil
}
