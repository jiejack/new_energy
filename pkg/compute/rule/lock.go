package rule

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ErrLockNotHeld = errors.New("lock not held by this client")
)

// RedisDistributedLock Redis分布式锁实现
type RedisDistributedLock struct {
	client *redis.Client
}

// NewRedisDistributedLock 创建Redis分布式锁
func NewRedisDistributedLock(client *redis.Client) *RedisDistributedLock {
	return &RedisDistributedLock{
		client: client,
	}
}

// Acquire 获取锁
func (l *RedisDistributedLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	// 使用SET NX EX命令原子性地获取锁
	result, err := l.client.SetNX(ctx, key, "locked", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return result, nil
}

// Release 释放锁
func (l *RedisDistributedLock) Release(ctx context.Context, key string) error {
	// 使用Lua脚本确保只有持有锁的客户端才能释放
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{key}, "locked").Int()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result == 0 {
		return ErrLockNotHeld
	}

	return nil
}

// IsHeld 检查锁是否被持有
func (l *RedisDistributedLock) IsHeld(ctx context.Context, key string) (bool, error) {
	result, err := l.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check lock: %w", err)
	}

	return result > 0, nil
}

// TryAcquire 尝试获取锁（带重试）
func (l *RedisDistributedLock) TryAcquire(ctx context.Context, key string, ttl time.Duration, maxRetries int, retryInterval time.Duration) (bool, error) {
	for i := 0; i < maxRetries; i++ {
		acquired, err := l.Acquire(ctx, key, ttl)
		if err != nil {
			return false, err
		}

		if acquired {
			return true, nil
		}

		// 等待一段时间后重试
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(retryInterval):
			continue
		}
	}

	return false, nil
}

// Extend 延长锁的过期时间
func (l *RedisDistributedLock) Extend(ctx context.Context, key string, ttl time.Duration) error {
	// 使用Lua脚本确保只有持有锁的客户端才能延长
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{key}, "locked", ttl.Milliseconds()).Int()
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	if result == 0 {
		return ErrLockNotHeld
	}

	return nil
}

// LocalLock 本地锁实现（用于单机部署）
type LocalLock struct {
	locks map[string]*lockEntry
	mu    chan struct{}
}

type lockEntry struct {
	holder    string
	expiresAt time.Time
}

// NewLocalLock 创建本地锁
func NewLocalLock() *LocalLock {
	return &LocalLock{
		locks: make(map[string]*lockEntry),
		mu:    make(chan struct{}, 1),
	}
}

// Acquire 获取锁
func (l *LocalLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	select {
	case l.mu <- struct{}{}:
		defer func() { <-l.mu }()

		// 检查锁是否存在
		if entry, exists := l.locks[key]; exists {
			// 检查是否过期
			if time.Now().Before(entry.expiresAt) {
				return false, nil
			}
		}

		// 获取锁
		l.locks[key] = &lockEntry{
			holder:    "local",
			expiresAt: time.Now().Add(ttl),
		}

		return true, nil

	case <-ctx.Done():
		return false, ctx.Err()
	}
}

// Release 释放锁
func (l *LocalLock) Release(ctx context.Context, key string) error {
	select {
	case l.mu <- struct{}{}:
		defer func() { <-l.mu }()

		delete(l.locks, key)
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsHeld 检查锁是否被持有
func (l *LocalLock) IsHeld(ctx context.Context, key string) (bool, error) {
	select {
	case l.mu <- struct{}{}:
		defer func() { <-l.mu }()

		entry, exists := l.locks[key]
		if !exists {
			return false, nil
		}

		// 检查是否过期
		if time.Now().After(entry.expiresAt) {
			delete(l.locks, key)
			return false, nil
		}

		return true, nil

	case <-ctx.Done():
		return false, ctx.Err()
	}
}

// LockManager 锁管理器
type LockManager struct {
	redisLock *RedisDistributedLock
	localLock *LocalLock
	useRedis  bool
}

// NewLockManager 创建锁管理器
func NewLockManager(redisClient *redis.Client) *LockManager {
	lm := &LockManager{
		localLock: NewLocalLock(),
		useRedis:  redisClient != nil,
	}

	if redisClient != nil {
		lm.redisLock = NewRedisDistributedLock(redisClient)
	}

	return lm
}

// Acquire 获取锁
func (lm *LockManager) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if lm.useRedis {
		return lm.redisLock.Acquire(ctx, key, ttl)
	}
	return lm.localLock.Acquire(ctx, key, ttl)
}

// Release 释放锁
func (lm *LockManager) Release(ctx context.Context, key string) error {
	if lm.useRedis {
		return lm.redisLock.Release(ctx, key)
	}
	return lm.localLock.Release(ctx, key)
}

// IsHeld 检查锁是否被持有
func (lm *LockManager) IsHeld(ctx context.Context, key string) (bool, error) {
	if lm.useRedis {
		return lm.redisLock.IsHeld(ctx, key)
	}
	return lm.localLock.IsHeld(ctx, key)
}

// TryAcquire 尝试获取锁
func (lm *LockManager) TryAcquire(ctx context.Context, key string, ttl time.Duration, maxRetries int, retryInterval time.Duration) (bool, error) {
	if lm.useRedis {
		return lm.redisLock.TryAcquire(ctx, key, ttl, maxRetries, retryInterval)
	}

	// 本地锁直接尝试获取
	for i := 0; i < maxRetries; i++ {
		acquired, err := lm.localLock.Acquire(ctx, key, ttl)
		if err != nil {
			return false, err
		}

		if acquired {
			return true, nil
		}

		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(retryInterval):
			continue
		}
	}

	return false, nil
}

// WithLock 使用锁执行函数
func (lm *LockManager) WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	// 获取锁
	acquired, err := lm.Acquire(ctx, key, ttl)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !acquired {
		return ErrLockFailed
	}

	// 确保释放锁
	defer lm.Release(ctx, key)

	// 执行函数
	return fn()
}
