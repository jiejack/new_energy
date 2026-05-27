package query

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRateLimiterConfig(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "token_bucket", cfg.Algorithm)
	assert.Equal(t, 100, cfg.DefaultRate)
	assert.Equal(t, 200, cfg.DefaultBurst)
	assert.Equal(t, 30*time.Second, cfg.MaxWaitTime)
	assert.Equal(t, 4, cfg.PriorityQueues)
	assert.True(t, cfg.EnableAdaptive)
}

func TestNewQueryRateLimiter(t *testing.T) {
	limiter := NewQueryRateLimiter(DefaultRateLimiterConfig())
	require.NotNil(t, limiter)
}

func TestQueryRateLimiter_Allow_TokenBucket(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.Algorithm = "token_bucket"
	cfg.DefaultRate = 100
	cfg.DefaultBurst = 5
	limiter := NewQueryRateLimiter(cfg)

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		result, err := limiter.Allow(ctx, "key1", PriorityNormal)
		require.NoError(t, err)
		assert.True(t, result.Allowed)
	}
}

func TestQueryRateLimiter_Allow_SlidingWindow(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.Algorithm = "sliding_window"
	cfg.DefaultRate = 5
	limiter := NewQueryRateLimiter(cfg)

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		result, err := limiter.Allow(ctx, "key1", PriorityNormal)
		require.NoError(t, err)
		assert.True(t, result.Allowed)
	}

	result, err := limiter.Allow(ctx, "key1", PriorityNormal)
	require.NoError(t, err)
	assert.False(t, result.Allowed)
}

func TestQueryRateLimiter_Allow_LeakyBucket(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.Algorithm = "leaky_bucket"
	cfg.DefaultRate = 5
	limiter := NewQueryRateLimiter(cfg)

	ctx := context.Background()
	result, err := limiter.Allow(ctx, "key1", PriorityNormal)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestQueryRateLimiter_Disabled(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.Enabled = false
	limiter := NewQueryRateLimiter(cfg)

	ctx := context.Background()
	result, err := limiter.Allow(ctx, "key1", PriorityNormal)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, RateLimitStatusAllowed, result.Status)
}

func TestQueryRateLimiter_GetStats(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.DefaultBurst = 100
	limiter := NewQueryRateLimiter(cfg)

	ctx := context.Background()
	limiter.Allow(ctx, "key1", PriorityNormal)
	limiter.Allow(ctx, "key1", PriorityNormal)

	stats := limiter.GetStats()
	assert.Equal(t, int64(2), stats.TotalRequests)
	assert.Equal(t, int64(2), stats.AllowedRequests)
}

func TestQueryRateLimiter_SetRate(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	limiter := NewQueryRateLimiter(cfg)

	ctx := context.Background()
	limiter.Allow(ctx, "key1", PriorityNormal)
	limiter.SetRate("key1", 200)
}

func TestQueryRateLimiter_Reset(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	limiter := NewQueryRateLimiter(cfg)

	ctx := context.Background()
	limiter.Allow(ctx, "key1", PriorityNormal)
	limiter.Reset("key1")
}

func TestTokenBucket(t *testing.T) {
	bucket := NewTokenBucket(10, 5)
	assert.True(t, bucket.Take())
	assert.True(t, bucket.Take())
	assert.True(t, bucket.Tokens() > 0)
}

func TestTokenBucket_TakeN(t *testing.T) {
	bucket := NewTokenBucket(10, 10)
	assert.True(t, bucket.TakeN(5))
	assert.False(t, bucket.TakeN(10))
}

func TestTokenBucket_Exhaust(t *testing.T) {
	bucket := NewTokenBucket(1, 2)
	assert.True(t, bucket.Take())
	assert.True(t, bucket.Take())
	assert.False(t, bucket.Take())
}

func TestTokenBucket_SetRate(t *testing.T) {
	bucket := NewTokenBucket(10, 10)
	bucket.SetRate(100)
}

func TestTokenBucket_LastRefill(t *testing.T) {
	bucket := NewTokenBucket(10, 10)
	refill := bucket.LastRefill()
	assert.False(t, refill.IsZero())
}

func TestSlidingWindow(t *testing.T) {
	window := NewSlidingWindow(3, time.Second)
	assert.True(t, window.Allow())
	assert.True(t, window.Allow())
	assert.True(t, window.Allow())
	assert.False(t, window.Allow())
	assert.Equal(t, 3, window.Count())
}

func TestSlidingWindow_SetLimit(t *testing.T) {
	window := NewSlidingWindow(3, time.Second)
	window.SetLimit(10)
}

func TestSlidingWindow_LastAccess(t *testing.T) {
	window := NewSlidingWindow(3, time.Second)
	window.Allow()
	la := window.LastAccess()
	assert.False(t, la.IsZero())
}

func TestPriorityQueue(t *testing.T) {
	q := NewPriorityQueue(10)
	assert.Equal(t, 0, q.Size())

	pos := q.Add("item1")
	assert.Equal(t, 0, pos)

	pos = q.Add("item2")
	assert.Equal(t, 1, pos)

	assert.True(t, q.IsNext("item1"))
	assert.False(t, q.IsNext("item2"))

	q.Remove("item1")
	assert.Equal(t, 1, q.Size())
	assert.True(t, q.IsNext("item2"))
}

func TestPriorityQueue_DuplicateAdd(t *testing.T) {
	q := NewPriorityQueue(10)
	q.Add("item1")
	pos := q.Add("item1")
	assert.Equal(t, 0, pos)
}

func TestPriorityQueue_Empty(t *testing.T) {
	q := NewPriorityQueue(10)
	assert.False(t, q.IsNext("anything"))
}

func TestPriorityQueue_RemoveNonExistent(t *testing.T) {
	q := NewPriorityQueue(10)
	q.Remove("nonexistent")
}

func TestAdaptiveRateLimiter(t *testing.T) {
	ar := NewAdaptiveRateLimiter(100)
	assert.Equal(t, 100, ar.GetCurrentRate())

	ar.RecordSuccess()
	ar.RecordFailure()
}

func TestRateLimitMiddleware(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.DefaultBurst = 100
	limiter := NewQueryRateLimiter(cfg)
	mw := NewRateLimitMiddleware(limiter)
	require.NotNil(t, mw)

	ctx := context.Background()
	req := NewQueryBuilder().Table("devices").Build()
	result, err := mw.Process(ctx, req)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestRateLimitMiddleware_CustomKeyFunc(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.DefaultBurst = 100
	limiter := NewQueryRateLimiter(cfg)
	mw := NewRateLimitMiddleware(limiter)

	mw.SetKeyFunc(func(req *QueryRequest) string {
		return "custom:" + req.Table
	})

	ctx := context.Background()
	req := NewQueryBuilder().Table("devices").Build()
	result, err := mw.Process(ctx, req)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestRateLimitMiddleware_CustomOnLimit(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.DefaultRate = 1
	cfg.DefaultBurst = 1
	limiter := NewQueryRateLimiter(cfg)
	mw := NewRateLimitMiddleware(limiter)

	mw.SetOnLimit(func(req *QueryRequest, result *RateLimitResult) error {
		_ = req
		_ = result
		return ErrQueryRateLimited
	})

	ctx := context.Background()
	req := NewQueryBuilder().Table("devices").Build()
	mw.Process(ctx, req)
	_, err := mw.Process(ctx, req)
	assert.Error(t, err)
}

func TestDefaultKeyFunc(t *testing.T) {
	req := &QueryRequest{Database: "db1", Table: "table1"}
	key := DefaultKeyFunc(req)
	assert.Equal(t, "query:db1:table1", key)
}

func TestDistributedRateLimiter(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.DefaultBurst = 100
	limiter := NewQueryRateLimiter(cfg)

	mockRedis := &mockRedisRateLimitClient{}
	distLimiter := NewDistributedRateLimiter(limiter, mockRedis, DistributedRateLimitConfig{
		GlobalRate: 100,
		LocalRate:  50,
		KeyPrefix:  "ratelimit:",
	})

	ctx := context.Background()
	result, err := distLimiter.Allow(ctx, "key1", PriorityNormal)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestDistributedRateLimiter_GlobalLimit(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.DefaultBurst = 100
	limiter := NewQueryRateLimiter(cfg)

	mockRedis := &mockRedisRateLimitClient{count: 101}
	distLimiter := NewDistributedRateLimiter(limiter, mockRedis, DistributedRateLimitConfig{
		GlobalRate: 100,
		LocalRate:  50,
		KeyPrefix:  "ratelimit:",
	})

	ctx := context.Background()
	result, err := distLimiter.Allow(ctx, "key1", PriorityNormal)
	require.NoError(t, err)
	assert.False(t, result.Allowed)
}

func TestDistributedRateLimiter_RedisError(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.DefaultBurst = 100
	limiter := NewQueryRateLimiter(cfg)

	mockRedis := &mockRedisRateLimitClient{err: assert.AnError}
	distLimiter := NewDistributedRateLimiter(limiter, mockRedis, DistributedRateLimitConfig{
		GlobalRate: 100,
		LocalRate:  50,
		KeyPrefix:  "ratelimit:",
	})

	ctx := context.Background()
	result, err := distLimiter.Allow(ctx, "key1", PriorityNormal)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestRateLimitGroup(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.DefaultBurst = 100
	group := NewRateLimitGroup("test", cfg)
	require.NotNil(t, group)

	ctx := context.Background()
	result, err := group.Allow(ctx, "group1", "key1", PriorityNormal)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestRateLimitGroup_Get(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	group := NewRateLimitGroup("test", cfg)

	l1 := group.Get("key1")
	l2 := group.Get("key1")
	assert.Equal(t, l1, l2)
}

func TestRateLimitRuleManager(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	limiter := NewQueryRateLimiter(cfg)
	mgr := NewRateLimitRuleManager(limiter)
	require.NotNil(t, mgr)

	rule := &RateLimitRule{
		ID:         "rule1",
		Name:       "Test Rule",
		KeyPattern: "query:*",
		Rate:       50,
		Burst:      100,
		Priority:   PriorityNormal,
		Enabled:    true,
	}
	mgr.AddRule(rule)

	got, exists := mgr.GetRule("rule1")
	assert.True(t, exists)
	assert.Equal(t, "Test Rule", got.Name)

	_, exists = mgr.GetRule("nonexistent")
	assert.False(t, exists)

	matched := mgr.MatchRule("query:db:table")
	assert.NotNil(t, matched)
	assert.Equal(t, "rule1", matched.ID)

	matched = mgr.MatchRule("other:db:table")
	assert.Nil(t, matched)

	mgr.RemoveRule("rule1")
	_, exists = mgr.GetRule("rule1")
	assert.False(t, exists)
}

func TestMatchPattern(t *testing.T) {
	assert.True(t, matchPattern("*", "anything"))
	assert.True(t, matchPattern("query:*", "query:db:table"))
	assert.True(t, matchPattern("exact", "exact"))
	assert.False(t, matchPattern("exact", "different"))
	assert.False(t, matchPattern("prefix*", "other"))
}

func TestRateLimitResult_Struct(t *testing.T) {
	result := &RateLimitResult{
		Status:     RateLimitStatusAllowed,
		Allowed:    true,
		RetryAfter: 0,
		Limit:      100,
		Remaining:  99,
		Key:        "test",
		Priority:   PriorityNormal,
	}
	assert.True(t, result.Allowed)
	assert.Equal(t, 100, result.Limit)
}

type mockRedisRateLimitClient struct {
	count int64
	err   error
}

func (m *mockRedisRateLimitClient) Incr(ctx context.Context, key string) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	m.count++
	return m.count, nil
}

func (m *mockRedisRateLimitClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return m.err
}

func (m *mockRedisRateLimitClient) Get(ctx context.Context, key string) (string, error) {
	return "", m.err
}

func (m *mockRedisRateLimitClient) Del(ctx context.Context, keys ...string) error {
	return m.err
}
