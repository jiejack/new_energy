package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestCache(t *testing.T) (*Cache, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cache := &Cache{
		client: client,
		cfg: &Config{
			Addr: mr.Addr(),
		},
	}

	t.Cleanup(func() {
		cache.Close()
		mr.Close()
	})

	return cache, mr
}

func TestNewCache_ConnectionError(t *testing.T) {
	cfg := &Config{
		Addr: "localhost:19999",
	}
	_, err := NewCache(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to redis")
}

func TestNewCache_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	cfg := &Config{
		Addr: mr.Addr(),
	}
	cache, err := NewCache(cfg)
	require.NoError(t, err)
	defer cache.Close()
	assert.NotNil(t, cache)
}

func TestCache_Close(t *testing.T) {
	cache, _ := setupTestCache(t)
	err := cache.Close()
	assert.NoError(t, err)
}

func TestCache_Set_Get(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	testData := &TestStruct{Name: "test", Value: 123}
	err := cache.Set(ctx, "test_key", testData, time.Minute)
	require.NoError(t, err)

	var result TestStruct
	err = cache.Get(ctx, "test_key", &result)
	require.NoError(t, err)
	assert.Equal(t, testData.Name, result.Name)
	assert.Equal(t, testData.Value, result.Value)
}

func TestCache_Get_NonExistent(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	var result string
	err := cache.Get(ctx, "nonexistent_key", &result)
	assert.Error(t, err)
}

func TestCache_Delete(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	err := cache.Set(ctx, "delete_key", "value", time.Minute)
	require.NoError(t, err)

	err = cache.Delete(ctx, "delete_key")
	require.NoError(t, err)

	var result string
	err = cache.Get(ctx, "delete_key", &result)
	assert.Error(t, err)
}

func TestCache_Delete_Multiple(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	err := cache.Set(ctx, "del_key1", "val1", time.Minute)
	require.NoError(t, err)
	err = cache.Set(ctx, "del_key2", "val2", time.Minute)
	require.NoError(t, err)

	err = cache.Delete(ctx, "del_key1", "del_key2")
	require.NoError(t, err)
}

func TestCache_Exists(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	err := cache.Set(ctx, "exists_key", "value", time.Minute)
	require.NoError(t, err)

	count, err := cache.Exists(ctx, "exists_key")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = cache.Exists(ctx, "nonexistent_key")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestCache_Expire(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	err := cache.Set(ctx, "expire_key", "value", 0)
	require.NoError(t, err)

	err = cache.Expire(ctx, "expire_key", 10*time.Second)
	require.NoError(t, err)

	ttl, err := cache.TTL(ctx, "expire_key")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, ttl.Seconds(), float64(9))
	assert.LessOrEqual(t, ttl.Seconds(), float64(10))
}

func TestCache_TTL(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	err := cache.Set(ctx, "ttl_key", "value", 10*time.Second)
	require.NoError(t, err)

	ttl, err := cache.TTL(ctx, "ttl_key")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, ttl.Seconds(), float64(9))
	assert.LessOrEqual(t, ttl.Seconds(), float64(10))
}

func TestCache_Increment(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	result, err := cache.Increment(ctx, "incr_key")
	require.NoError(t, err)
	assert.Equal(t, int64(1), result)

	result, err = cache.Increment(ctx, "incr_key")
	require.NoError(t, err)
	assert.Equal(t, int64(2), result)
}

func TestCache_Decrement(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	_, _ = cache.Increment(ctx, "decr_key")
	_, _ = cache.Increment(ctx, "decr_key")

	result, err := cache.Decrement(ctx, "decr_key")
	require.NoError(t, err)
	assert.Equal(t, int64(1), result)
}

func TestCache_IncrementBy(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	result, err := cache.IncrementBy(ctx, "incrby_key", 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), result)

	result, err = cache.IncrementBy(ctx, "incrby_key", 3)
	require.NoError(t, err)
	assert.Equal(t, int64(8), result)
}

func TestCache_HashSet_HashGet(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	values := map[string]interface{}{
		"field1": "value1",
		"field2": 123,
	}

	err := cache.HashSet(ctx, "hash_key", values)
	require.NoError(t, err)

	var result string
	err = cache.HashGet(ctx, "hash_key", "field1", &result)
	require.NoError(t, err)
	assert.Equal(t, "value1", result)

	var intResult int
	err = cache.HashGet(ctx, "hash_key", "field2", &intResult)
	require.NoError(t, err)
	assert.Equal(t, 123, intResult)
}

func TestCache_HashGet_NonExistent(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	var result string
	err := cache.HashGet(ctx, "nonexistent_hash", "field1", &result)
	assert.Error(t, err)
}

func TestCache_HashGetAll(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	values := map[string]interface{}{
		"field1": "value1",
		"field2": 42,
	}

	err := cache.HashSet(ctx, "hashall_key", values)
	require.NoError(t, err)

	dest := make(map[string]interface{})
	err = cache.HashGetAll(ctx, "hashall_key", dest)
	require.NoError(t, err)
	assert.Equal(t, 2, len(dest))
}

func TestCache_HashDelete(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	values := map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}

	err := cache.HashSet(ctx, "hashdel_key", values)
	require.NoError(t, err)

	err = cache.HashDelete(ctx, "hashdel_key", "field1")
	require.NoError(t, err)

	exists, err := cache.HashExists(ctx, "hashdel_key", "field1")
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = cache.HashExists(ctx, "hashdel_key", "field2")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestCache_HashExists(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	values := map[string]interface{}{
		"field1": "value1",
	}

	err := cache.HashSet(ctx, "hashexists_key", values)
	require.NoError(t, err)

	exists, err := cache.HashExists(ctx, "hashexists_key", "field1")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = cache.HashExists(ctx, "hashexists_key", "nonexistent")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestCache_HashLen(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	values := map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
	}

	err := cache.HashSet(ctx, "hashlen_key", values)
	require.NoError(t, err)

	length, err := cache.HashLen(ctx, "hashlen_key")
	require.NoError(t, err)
	assert.Equal(t, int64(3), length)
}

func TestCache_SetMulti_GetMulti(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	items := map[string]interface{}{
		"multi_key1": "value1",
		"multi_key2": "value2",
		"multi_key3": "value3",
	}

	err := cache.SetMulti(ctx, items, time.Minute)
	require.NoError(t, err)

	results, err := cache.GetMulti(ctx, []string{"multi_key1", "multi_key2", "multi_key3"})
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestCache_GetMulti_WithMissing(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	err := cache.Set(ctx, "existing_key", "value", time.Minute)
	require.NoError(t, err)

	results, err := cache.GetMulti(ctx, []string{"existing_key", "missing_key"})
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.NotNil(t, results[0])
	assert.Nil(t, results[1])
}

func TestCache_Publish(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	err := cache.Publish(ctx, "test_channel", map[string]string{"msg": "hello"})
	require.NoError(t, err)
}

func TestCache_Subscribe(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	sub := cache.Subscribe(ctx, "test_channel")
	require.NotNil(t, sub)
	sub.Close()
}

func TestCache_Set_UnmarshalableValue(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	ch := make(chan int)
	err := cache.Set(ctx, "bad_key", ch, time.Minute)
	assert.Error(t, err)
}

func TestCache_HashSet_UnmarshalableValue(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	ch := make(chan int)
	values := map[string]interface{}{
		"field1": ch,
	}
	err := cache.HashSet(ctx, "bad_hash_key", values)
	assert.Error(t, err)
}

func TestCache_SetMulti_UnmarshalableValue(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	ch := make(chan int)
	items := map[string]interface{}{
		"bad_key": ch,
	}
	err := cache.SetMulti(ctx, items, time.Minute)
	assert.Error(t, err)
}

func TestCache_Publish_UnmarshalableValue(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	ch := make(chan int)
	err := cache.Publish(ctx, "bad_channel", ch)
	assert.Error(t, err)
}
