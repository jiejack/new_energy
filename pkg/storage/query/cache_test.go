package query

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCacheConfig(t *testing.T) {
	cfg := DefaultCacheConfig()
	assert.True(t, cfg.Enabled)
	assert.Equal(t, 5*time.Minute, cfg.DefaultTTL)
	assert.Equal(t, int64(100*1024*1024), cfg.MaxSize)
	assert.Equal(t, 10000, cfg.MaxEntries)
	assert.Equal(t, "LRU", cfg.EvictionPolicy)
	assert.Equal(t, "nem:query:cache:", cfg.KeyPrefix)
	assert.True(t, cfg.EnableStats)
}

func TestCacheEntry_IsExpired(t *testing.T) {
	entry := &CacheEntry{
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	assert.False(t, entry.IsExpired())

	entry.ExpiresAt = time.Now().Add(-1 * time.Hour)
	assert.True(t, entry.IsExpired())
}

func TestNewQueryCache(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	require.NotNil(t, cache)
}

func TestQueryCache_SetGet(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{
		QueryID: "test",
		Status:  QueryStatusCompleted,
		Data:    []map[string]interface{}{{"id": 1, "name": "test"}},
		Total:   1,
	}

	err := cache.Set(ctx, req, result)
	require.NoError(t, err)

	got, status, err := cache.Get(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, CacheStatusHit, status)
	require.NotNil(t, got)
	assert.Equal(t, "test", got.QueryID)
	assert.True(t, got.Cached)
}

func TestQueryCache_Get_Miss(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	got, status, err := cache.Get(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, CacheStatusMiss, status)
	assert.Nil(t, got)
}

func TestQueryCache_Disabled(t *testing.T) {
	cfg := DefaultCacheConfig()
	cfg.Enabled = false
	cache := NewQueryCache(nil, cfg)
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted}

	err := cache.Set(ctx, req, result)
	require.NoError(t, err)

	got, status, err := cache.Get(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, CacheStatusMiss, status)
	assert.Nil(t, got)
}

func TestQueryCache_Delete(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted}
	cache.Set(ctx, req, result)

	err := cache.Delete(ctx, req)
	require.NoError(t, err)

	got, status, _ := cache.Get(ctx, req)
	assert.Equal(t, CacheStatusMiss, status)
	assert.Nil(t, got)
}

func TestQueryCache_DeleteByTags(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted}
	cache.Set(ctx, req, result)

	err := cache.DeleteByTags(ctx, []string{"table:devices"})
	require.NoError(t, err)

	got, status, _ := cache.Get(ctx, req)
	assert.Equal(t, CacheStatusMiss, status)
	assert.Nil(t, got)
}

func TestQueryCache_Clear(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted}
	cache.Set(ctx, req, result)

	err := cache.Clear(ctx)
	require.NoError(t, err)

	got, status, _ := cache.Get(ctx, req)
	assert.Equal(t, CacheStatusMiss, status)
	assert.Nil(t, got)
}

func TestQueryCache_Stats(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted}
	cache.Set(ctx, req, result)

	cache.Get(ctx, req)

	stats := cache.GetStats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.TotalRequests)
	assert.True(t, stats.EntryCount > 0)
}

func TestQueryCache_ResetStats(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted}
	cache.Set(ctx, req, result)
	cache.Get(ctx, req)

	cache.ResetStats()
	stats := cache.GetStats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.TotalRequests)
}

func TestQueryCache_GenerateKey(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	req1 := NewQueryBuilder().Table("devices").Where("status", "=", "active").Build()
	req2 := NewQueryBuilder().Table("devices").Where("status", "=", "inactive").Build()

	key1 := cache.GenerateKey(req1)
	key2 := cache.GenerateKey(req2)
	assert.NotEqual(t, key1, key2)

	key3 := cache.GenerateKey(req1)
	assert.Equal(t, key1, key3)
}

func TestQueryCache_GenerateHash(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	req := NewQueryBuilder().Table("devices").Build()
	hash := cache.GenerateHash(req)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64)
}

func TestQueryCache_Warmup(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	ctx := context.Background()

	tasks := []*WarmupTask{
		{Name: "warmup1", Query: NewQueryBuilder().Table("devices").Build(), Enabled: true},
		{Name: "warmup2", Query: NewQueryBuilder().Table("stations").Build(), Enabled: false},
	}
	err := cache.Warmup(ctx, tasks)
	require.NoError(t, err)
}

func TestQueryCache_Eviction(t *testing.T) {
	cfg := DefaultCacheConfig()
	cfg.MaxEntries = 2
	cfg.MaxSize = 1024 * 1024
	cache := NewQueryCache(nil, cfg)
	ctx := context.Background()

	req1 := NewQueryBuilder().Table("table1").Build()
	req2 := NewQueryBuilder().Table("table2").Build()
	req3 := NewQueryBuilder().Table("table3").Build()

	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted, Data: []map[string]interface{}{{"id": 1}}}

	cache.Set(ctx, req1, result)
	cache.Set(ctx, req2, result)
	cache.Set(ctx, req3, result)

	stats := cache.GetStats()
	assert.True(t, stats.Evictions > 0)
}

func TestQueryCache_Expired(t *testing.T) {
	cfg := DefaultCacheConfig()
	cfg.DefaultTTL = 1 * time.Nanosecond
	cache := NewQueryCache(nil, cfg)
	ctx := context.Background()

	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted}
	cache.Set(ctx, req, result)

	time.Sleep(10 * time.Millisecond)
	got, status, _ := cache.Get(ctx, req)
	assert.Equal(t, CacheStatusExpired, status)
	assert.Nil(t, got)
}

func TestLRUList(t *testing.T) {
	l := NewLRUList()
	require.NotNil(t, l)

	l.PushFront("a")
	l.PushFront("b")
	l.PushFront("c")

	back, ok := l.Back()
	assert.True(t, ok)
	assert.Equal(t, "a", back)

	l.MoveToFront("a")
	back, ok = l.Back()
	assert.True(t, ok)
	assert.Equal(t, "b", back)

	l.Remove("b")
	back, ok = l.Back()
	assert.True(t, ok)
	assert.Equal(t, "c", back)
}

func TestLRUList_Empty(t *testing.T) {
	l := NewLRUList()
	_, ok := l.Back()
	assert.False(t, ok)
}

func TestLRUList_DuplicatePush(t *testing.T) {
	l := NewLRUList()
	l.PushFront("a")
	l.PushFront("a")
	back, ok := l.Back()
	assert.True(t, ok)
	assert.Equal(t, "a", back)
}

func TestLRUList_MoveNonExistent(t *testing.T) {
	l := NewLRUList()
	l.MoveToFront("nonexistent")
}

func TestLRUList_RemoveNonExistent(t *testing.T) {
	l := NewLRUList()
	l.Remove("nonexistent")
}

func TestCacheInvalidator(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	invalidator := NewCacheInvalidator(cache)
	require.NotNil(t, invalidator)

	invalidator.AddRule(InvalidationRule{
		ID:      "rule1",
		Name:    "test rule",
		Trigger: "event",
		Tables:  []string{"devices"},
		Enabled: true,
	})

	invalidator.Notify(InvalidationEvent{
		Type:  "update",
		Table: "devices",
	})

	invalidator.RemoveRule("rule1")
}

func TestCacheKeyBuilder(t *testing.T) {
	b := NewCacheKeyBuilder()
	key := b.Add("devices").AddInt(42).AddInt64(100).AddTime(time.Now()).Build()
	assert.NotEmpty(t, key)
}

func TestMultiLevelCache(t *testing.T) {
	cfg := MultiLevelCacheConfig{
		Levels: []CacheLevelConfig{
			{Name: "L1", Priority: 1, Config: DefaultCacheConfig()},
			{Name: "L2", Priority: 2, Config: DefaultCacheConfig()},
		},
	}
	mlc := NewMultiLevelCache(nil, cfg)
	require.NotNil(t, mlc)

	ctx := context.Background()
	req := NewQueryBuilder().Table("devices").Build()
	result := &QueryResult{QueryID: "test", Status: QueryStatusCompleted}

	err := mlc.Set(ctx, req, result)
	require.NoError(t, err)

	got, status, err := mlc.Get(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, CacheStatusHit, status)
	require.NotNil(t, got)

	err = mlc.Delete(ctx, req)
	require.NoError(t, err)
}

func TestCacheStatus_Constants(t *testing.T) {
	assert.Equal(t, CacheStatus("hit"), CacheStatusHit)
	assert.Equal(t, CacheStatus("miss"), CacheStatusMiss)
	assert.Equal(t, CacheStatus("expired"), CacheStatusExpired)
	assert.Equal(t, CacheStatus("evicted"), CacheStatusEvicted)
}

func TestCacheWarmer(t *testing.T) {
	cache := NewQueryCache(nil, DefaultCacheConfig())
	warmer := NewCacheWarmer(cache, nil)
	require.NotNil(t, warmer)

	warmer.AddTask(&WarmupTask{
		Name:    "test",
		Query:   NewQueryBuilder().Table("devices").Build(),
		Enabled: true,
	})

	warmer.Stop()
}
