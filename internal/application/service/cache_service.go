package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/new-energy-monitoring/internal/domain/cache"
	"github.com/new-energy-monitoring/internal/infrastructure/config"
	"go.uber.org/zap"
)

// CacheService 缓存服务
type CacheService struct {
	cache  cache.Cache
	config *config.CacheConfig
	logger *zap.Logger
}

// NewCacheService 创建缓存服务
func NewCacheService(cache cache.Cache, cfg *config.CacheConfig, logger *zap.Logger) *CacheService {
	return &CacheService{
		cache:  cache,
		config: cfg,
		logger: logger,
	}
}

// Get 获取缓存
func (s *CacheService) Get(ctx context.Context, key string) (string, error) {
	fullKey := s.buildKey(key)
	return s.cache.Get(ctx, fullKey)
}

// Set 设置缓存
func (s *CacheService) Set(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error {
	fullKey := s.buildKey(key)
	ttl := s.getTTL(expiration...)
	return s.cache.Set(ctx, fullKey, value, ttl)
}

// GetJSON 获取JSON缓存
func (s *CacheService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	fullKey := s.buildKey(key)
	return s.cache.GetJSON(ctx, fullKey, dest)
}

// SetJSON 设置JSON缓存
func (s *CacheService) SetJSON(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error {
	fullKey := s.buildKey(key)
	ttl := s.getTTL(expiration...)
	return s.cache.SetJSON(ctx, fullKey, value, ttl)
}

// Delete 删除缓存
func (s *CacheService) Delete(ctx context.Context, keys ...string) error {
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = s.buildKey(key)
	}
	return s.cache.Del(ctx, fullKeys...)
}

// Exists 检查缓存是否存在
func (s *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := s.buildKey(key)
	count, err := s.cache.Exists(ctx, fullKey)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Expire 设置缓存过期时间
func (s *CacheService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	fullKey := s.buildKey(key)
	return s.cache.Expire(ctx, fullKey, expiration)
}

// TTL 获取缓存剩余过期时间
func (s *CacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := s.buildKey(key)
	return s.cache.TTL(ctx, fullKey)
}

// Increment 自增
func (s *CacheService) Increment(ctx context.Context, key string) (int64, error) {
	fullKey := s.buildKey(key)
	return s.cache.Incr(ctx, fullKey)
}

// IncrementBy 自增指定值
func (s *CacheService) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	fullKey := s.buildKey(key)
	// 使用Incr实现
	result, err := s.cache.Incr(ctx, fullKey)
	if err != nil {
		return 0, err
	}
	// 如果value > 1，需要额外增加
	if value > 1 {
		for i := int64(1); i < value; i++ {
			_, err = s.cache.Incr(ctx, fullKey)
			if err != nil {
				return 0, err
			}
		}
		result += value - 1
	}
	return result, nil
}

// Decrement 自减
func (s *CacheService) Decrement(ctx context.Context, key string) (int64, error) {
	fullKey := s.buildKey(key)
	return s.cache.Decr(ctx, fullKey)
}

// DecrementBy 自减指定值
func (s *CacheService) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	fullKey := s.buildKey(key)
	result, err := s.cache.Decr(ctx, fullKey)
	if err != nil {
		return 0, err
	}
	if value > 1 {
		for i := int64(1); i < value; i++ {
			_, err = s.cache.Decr(ctx, fullKey)
			if err != nil {
				return 0, err
			}
		}
		result -= value - 1
	}
	return result, nil
}

// HashSet 设置Hash字段
func (s *CacheService) HashSet(ctx context.Context, key string, field string, value interface{}) error {
	fullKey := s.buildKey(key)
	return s.cache.HSet(ctx, fullKey, field, value)
}

// HashGet 获取Hash字段
func (s *CacheService) HashGet(ctx context.Context, key string, field string) (string, error) {
	fullKey := s.buildKey(key)
	return s.cache.HGet(ctx, fullKey, field)
}

// HashGetAll 获取所有Hash字段
func (s *CacheService) HashGetAll(ctx context.Context, key string) (map[string]string, error) {
	fullKey := s.buildKey(key)
	return s.cache.HGetAll(ctx, fullKey)
}

// HashDelete 删除Hash字段
func (s *CacheService) HashDelete(ctx context.Context, key string, fields ...string) error {
	fullKey := s.buildKey(key)
	return s.cache.HDel(ctx, fullKey, fields...)
}

// ListPush 左侧插入列表
func (s *CacheService) ListPush(ctx context.Context, key string, values ...interface{}) error {
	fullKey := s.buildKey(key)
	return s.cache.LPush(ctx, fullKey, values...)
}

// ListPop 左侧弹出列表元素
func (s *CacheService) ListPop(ctx context.Context, key string) (string, error) {
	fullKey := s.buildKey(key)
	return s.cache.LPop(ctx, fullKey)
}

// ListRange 获取列表范围
func (s *CacheService) ListRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	fullKey := s.buildKey(key)
	return s.cache.LRange(ctx, fullKey, start, stop)
}

// SortedSetAdd 添加有序集合成员
func (s *CacheService) SortedSetAdd(ctx context.Context, key string, score float64, member string) error {
	fullKey := s.buildKey(key)
	return s.cache.ZAdd(ctx, fullKey, score, member)
}

// SortedSetRange 获取有序集合范围
func (s *CacheService) SortedSetRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	fullKey := s.buildKey(key)
	return s.cache.ZRange(ctx, fullKey, start, stop)
}

// SortedSetRemove 删除有序集合成员
func (s *CacheService) SortedSetRemove(ctx context.Context, key string, members ...interface{}) error {
	fullKey := s.buildKey(key)
	return s.cache.ZRem(ctx, fullKey, members...)
}

// GetOrSet 获取缓存，如果不存在则设置
func (s *CacheService) GetOrSet(ctx context.Context, key string, fn func() (interface{}, error), expiration ...time.Duration) (interface{}, error) {
	// 尝试获取缓存
	value, err := s.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	// 缓存不存在，执行函数获取数据
	data, err := fn()
	if err != nil {
		return nil, err
	}

	// 设置缓存
	if err := s.Set(ctx, key, data, expiration...); err != nil {
		s.logger.Error("Failed to set cache",
			zap.Error(err),
			zap.String("key", key),
		)
	}

	return data, nil
}

// GetOrSetJSON 获取JSON缓存，如果不存在则设置
func (s *CacheService) GetOrSetJSON(ctx context.Context, key string, dest interface{}, fn func() (interface{}, error), expiration ...time.Duration) error {
	// 尝试获取缓存
	err := s.GetJSON(ctx, key, dest)
	if err == nil {
		return nil
	}

	// 缓存不存在，执行函数获取数据
	data, err := fn()
	if err != nil {
		return err
	}

	// 设置缓存
	if err := s.SetJSON(ctx, key, data, expiration...); err != nil {
		s.logger.Error("Failed to set JSON cache",
			zap.Error(err),
			zap.String("key", key),
		)
	}

	// 将数据赋值给dest
	if err := copyData(data, dest); err != nil {
		return err
	}

	return nil
}

// Remember 记住缓存（带标签）
func (s *CacheService) Remember(ctx context.Context, key string, tags []string, fn func() (interface{}, error), expiration ...time.Duration) (interface{}, error) {
	// 尝试获取缓存
	value, err := s.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	// 缓存不存在，执行函数获取数据
	data, err := fn()
	if err != nil {
		return nil, err
	}

	// 设置缓存
	ttl := s.getTTL(expiration...)
	if err := s.Set(ctx, key, data, ttl); err != nil {
		s.logger.Error("Failed to set cache",
			zap.Error(err),
			zap.String("key", key),
		)
	}

	// 为缓存添加标签
	for _, tag := range tags {
		tagKey := s.buildTagKey(tag)
		if err := s.cache.RPush(ctx, tagKey, s.buildKey(key)); err != nil {
			s.logger.Error("Failed to add tag to cache",
				zap.Error(err),
				zap.String("tag", tag),
				zap.String("key", key),
			)
		}
		// 设置标签过期时间
		if err := s.cache.Expire(ctx, tagKey, ttl); err != nil {
			s.logger.Error("Failed to set tag expiration",
				zap.Error(err),
				zap.String("tag", tag),
			)
		}
	}

	return data, nil
}

// Forget 按标签删除缓存
func (s *CacheService) Forget(ctx context.Context, tags ...string) error {
	for _, tag := range tags {
		tagKey := s.buildTagKey(tag)

		// 获取该标签下的所有键
		keys, err := s.cache.LRange(ctx, tagKey, 0, -1)
		if err != nil {
			s.logger.Error("Failed to get tagged keys",
				zap.Error(err),
				zap.String("tag", tag),
			)
			continue
		}

		if len(keys) > 0 {
			// 删除缓存键
			if err := s.cache.Del(ctx, keys...); err != nil {
				s.logger.Error("Failed to delete tagged cache",
					zap.Error(err),
					zap.String("tag", tag),
				)
			}

			// 删除标签键
			if err := s.cache.Del(ctx, tagKey); err != nil {
				s.logger.Error("Failed to delete tag key",
					zap.Error(err),
					zap.String("tag", tag),
				)
			}
		}
	}

	return nil
}

// Flush 清空所有缓存
func (s *CacheService) Flush(ctx context.Context) error {
	// 注意：这会清空整个数据库，谨慎使用
	return s.cache.Close()
}

// buildKey 构建完整的缓存键
func (s *CacheService) buildKey(key string) string {
	return s.config.Redis.KeyPrefix + key
}

// buildTagKey 构建标签键
func (s *CacheService) buildTagKey(tag string) string {
	return s.config.Redis.KeyPrefix + "tag:" + tag
}

// getTTL 获取过期时间
func (s *CacheService) getTTL(expiration ...time.Duration) time.Duration {
	if len(expiration) > 0 && expiration[0] > 0 {
		return expiration[0]
	}
	return s.config.Redis.DefaultExpiration
}

// copyData 复制数据
func copyData(src, dest interface{}) error {
	// 使用JSON序列化和反序列化进行深拷贝
	data, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("failed to marshal source data: %w", err)
	}
	return json.Unmarshal(data, dest)
}

// CacheItem 缓存项
type CacheItem struct {
	Key        string      `json:"key"`
	Value      interface{} `json:"value"`
	Expiration time.Time   `json:"expiration"`
	Tags       []string    `json:"tags,omitempty"`
}

// CacheStats 缓存统计
type CacheStats struct {
	TotalKeys      int64         `json:"total_keys"`
	TotalSize      int64         `json:"total_size"`
	HitRate        float64       `json:"hit_rate"`
	AverageTTL     time.Duration `json:"average_ttl"`
	KeysByPrefix   map[string]int64 `json:"keys_by_prefix"`
	ExpirationInfo ExpirationInfo `json:"expiration_info"`
}

// ExpirationInfo 过期信息
type ExpirationInfo struct {
	ExpiredKeys     int64 `json:"expired_keys"`
	ExpiringIn1Min  int64 `json:"expiring_in_1min"`
	ExpiringIn5Min  int64 `json:"expiring_in_5min"`
	ExpiringIn1Hour int64 `json:"expiring_in_1hour"`
	ExpiringIn1Day  int64 `json:"expiring_in_1day"`
}
