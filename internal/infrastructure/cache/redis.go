package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient Redis客户端封装
type RedisClient struct {
	client *redis.Client
	config RedisConfig

	// 统计信息
	stats      CacheStats
	statsMutex sync.RWMutex
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addrs           []string      `mapstructure:"addrs"`
	Password        string        `mapstructure:"password"`
	DB              int           `mapstructure:"db"`
	PoolSize        int           `mapstructure:"pool_size"`
	MinIdleConns    int           `mapstructure:"min_idle_conns"`
	MaxRetries      int           `mapstructure:"max_retries"`
	DialTimeout     time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	PoolTimeout     time.Duration `mapstructure:"pool_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	MaxConnAge      time.Duration `mapstructure:"max_conn_age"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Hits       int64 // 缓存命中次数
	Misses     int64 // 缓存未命中次数
	Sets       int64 // 设置次数
	Gets       int64 // 获取次数
	Dels       int64 // 删除次数
	Errors     int64 // 错误次数
	TotalTime  int64 // 总耗时（纳秒）
	SlowCount  int64 // 慢查询次数
	SlowTime   int64 // 慢查询总耗时（纳秒）
	LastReset  time.Time
}

// SlowQueryThreshold 慢查询阈值（默认100ms）
const SlowQueryThreshold = 100 * time.Millisecond

// NewRedisClient 创建Redis客户端
func NewRedisClient(cfg RedisConfig) (*RedisClient, error) {
	// 设置默认值
	if cfg.PoolSize == 0 {
		cfg.PoolSize = 100
	}
	if cfg.MinIdleConns == 0 {
		cfg.MinIdleConns = 10
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 5 * time.Second
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 3 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 3 * time.Second
	}
	if cfg.PoolTimeout == 0 {
		cfg.PoolTimeout = 4 * time.Second
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 5 * time.Minute
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = 5 * time.Minute
	}

	client := redis.NewClient(&redis.Options{
		Addr:            cfg.Addrs[0],
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxRetries:      cfg.MaxRetries,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolTimeout:     cfg.PoolTimeout,
		IdleTimeout:     cfg.IdleTimeout,
		MaxConnAge:      cfg.MaxConnAge,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisClient{
		client: client,
		config: cfg,
		stats: CacheStats{
			LastReset: time.Now(),
		},
	}, nil
}

// Close 关闭连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Get 获取缓存
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	result, err := r.client.Get(ctx, key).Result()
	duration := time.Since(start)

	r.recordGet(err == nil, duration)

	return result, err
}

// Set 设置缓存
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	err := r.client.Set(ctx, key, value, expiration).Err()
	duration := time.Since(start)

	r.recordSet(duration)

	return err
}

// SetJSON 设置JSON缓存
func (r *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.Set(ctx, key, data, expiration)
}

// GetJSON 获取JSON缓存
func (r *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Del 删除缓存
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	start := time.Now()
	err := r.client.Del(ctx, keys...).Err()
	duration := time.Since(start)

	r.recordDel(duration)

	return err
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Incr 自增
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Decr 自减
func (r *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// HSet 设置Hash字段
func (r *RedisClient) HSet(ctx context.Context, key string, field string, value interface{}) error {
	return r.client.HSet(ctx, key, field, value).Err()
}

// HGet 获取Hash字段
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有Hash字段
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel 删除Hash字段
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// LPush 左侧插入列表
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPush 右侧插入列表
func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

// LRange 获取列表范围
func (r *RedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, key, start, stop).Result()
}

// LPop 左侧弹出列表元素
func (r *RedisClient) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

// RPop 右侧弹出列表元素
func (r *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// Publish 发布消息
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Subscribe 订阅频道
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// ZAdd 添加有序集合成员
func (r *RedisClient) ZAdd(ctx context.Context, key string, score float64, member string) error {
	return r.client.ZAdd(ctx, key, &redis.Z{Score: score, Member: member}).Err()
}

// ZRange 获取有序集合范围
func (r *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeByScore 按分数获取有序集合范围
func (r *RedisClient) ZRangeByScore(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error) {
	return r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}).Result()
}

// ZRem 删除有序集合成员
func (r *RedisClient) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

// Keys 按模式查找键
func (r *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// Scan 扫描键
func (r *RedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return r.client.Scan(ctx, cursor, match, count).Result()
}

// FlushDB 清空当前数据库
func (r *RedisClient) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Ping 测试连接
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// GetPoolStats 获取连接池统计信息
func (r *RedisClient) GetPoolStats() *redis.PoolStats {
	return r.client.PoolStats()
}

// GetStats 获取缓存统计信息
func (r *RedisClient) GetStats() CacheStats {
	r.statsMutex.RLock()
	defer r.statsMutex.RUnlock()
	return r.stats
}

// ResetStats 重置统计信息
func (r *RedisClient) ResetStats() {
	r.statsMutex.Lock()
	defer r.statsMutex.Unlock()
	r.stats = CacheStats{
		LastReset: time.Now(),
	}
}

// GetHitRate 获取缓存命中率
func (r *RedisClient) GetHitRate() float64 {
	r.statsMutex.RLock()
	defer r.statsMutex.RUnlock()

	total := r.stats.Hits + r.stats.Misses
	if total == 0 {
		return 0
	}
	return float64(r.stats.Hits) / float64(total)
}

// recordGet 记录Get操作统计
func (r *RedisClient) recordGet(hit bool, duration time.Duration) {
	r.statsMutex.Lock()
	defer r.statsMutex.Unlock()

	r.stats.Gets++
	r.stats.TotalTime += int64(duration)

	if hit {
		r.stats.Hits++
	} else {
		r.stats.Misses++
	}

	if duration > SlowQueryThreshold {
		r.stats.SlowCount++
		r.stats.SlowTime += int64(duration)
	}
}

// recordSet 记录Set操作统计
func (r *RedisClient) recordSet(duration time.Duration) {
	r.statsMutex.Lock()
	defer r.statsMutex.Unlock()

	r.stats.Sets++
	r.stats.TotalTime += int64(duration)

	if duration > SlowQueryThreshold {
		r.stats.SlowCount++
		r.stats.SlowTime += int64(duration)
	}
}

// recordDel 记录Del操作统计
func (r *RedisClient) recordDel(duration time.Duration) {
	r.statsMutex.Lock()
	defer r.statsMutex.Unlock()

	r.stats.Dels++
	r.stats.TotalTime += int64(duration)

	if duration > SlowQueryThreshold {
		r.stats.SlowCount++
		r.stats.SlowTime += int64(duration)
	}
}

// recordError 记录错误统计
func (r *RedisClient) recordError() {
	atomic.AddInt64(&r.stats.Errors, 1)
}

const (
	RealtimeDataPrefix = "nem:realtime:"
	AlarmActivePrefix  = "nem:alarm:active:"
	AlarmCountKey      = "nem:alarm:count"
	DeviceStatusPrefix = "nem:device:status:"
)

func (r *RedisClient) SetRealtimeData(ctx context.Context, pointID string, value float64, timestamp int64) error {
	key := RealtimeDataPrefix + pointID
	err := r.client.HSet(ctx, key, "value", value, "timestamp", timestamp).Err()
	return err
}

func (r *RedisClient) GetRealtimeData(ctx context.Context, pointID string) (float64, int64, error) {
	key := RealtimeDataPrefix + pointID
	data, err := r.HGetAll(ctx, key)
	if err != nil {
		return 0, 0, err
	}
	
	var value float64
	var timestamp int64
	fmt.Sscanf(data["value"], "%f", &value)
	fmt.Sscanf(data["timestamp"], "%d", &timestamp)
	
	return value, timestamp, nil
}

func (r *RedisClient) SetDeviceStatus(ctx context.Context, deviceID string, status int) error {
	key := DeviceStatusPrefix + deviceID
	return r.Set(ctx, key, status, 0)
}

func (r *RedisClient) GetDeviceStatus(ctx context.Context, deviceID string) (int, error) {
	key := DeviceStatusPrefix + deviceID
	result, err := r.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	
	var status int
	fmt.Sscanf(result, "%d", &status)
	return status, nil
}
