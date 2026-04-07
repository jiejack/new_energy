package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_Set_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Skipf("Failed to connect to redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	testData := &TestStruct{
		Name:  "test",
		Value: 123,
	}

	err = cache.Set(ctx, "test_key", testData, time.Minute)
	require.NoError(t, err)

	var result TestStruct
	err = cache.Get(ctx, "test_key", &result)
	require.NoError(t, err)

	assert.Equal(t, testData.Name, result.Name)
	assert.Equal(t, testData.Value, result.Value)

	err = cache.Delete(ctx, "test_key")
	require.NoError(t, err)
}

func TestCache_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Skipf("Failed to connect to redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	err = cache.Set(ctx, "delete_test_key", "test_value", time.Minute)
	require.NoError(t, err)

	err = cache.Delete(ctx, "delete_test_key")
	require.NoError(t, err)

	var result string
	err = cache.Get(ctx, "delete_test_key", &result)
	assert.Error(t, err)
}

func TestCache_Exists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Skipf("Failed to connect to redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	err = cache.Set(ctx, "exists_test_key", "test_value", time.Minute)
	require.NoError(t, err)

	count, err := cache.Exists(ctx, "exists_test_key")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = cache.Exists(ctx, "nonexistent_key")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	cache.Delete(ctx, "exists_test_key")
}

func TestCache_Increment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Skipf("Failed to connect to redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	result, err := cache.Increment(ctx, "incr_test_key")
	require.NoError(t, err)
	assert.Equal(t, int64(1), result)

	result, err = cache.Increment(ctx, "incr_test_key")
	require.NoError(t, err)
	assert.Equal(t, int64(2), result)

	cache.Delete(ctx, "incr_test_key")
}

func TestCache_HashSet_HashGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Skipf("Failed to connect to redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	values := map[string]interface{}{
		"field1": "value1",
		"field2": 123,
	}

	err = cache.HashSet(ctx, "hash_test_key", values)
	require.NoError(t, err)

	var result string
	err = cache.HashGet(ctx, "hash_test_key", "field1", &result)
	require.NoError(t, err)
	assert.Equal(t, "value1", result)

	var intResult int
	err = cache.HashGet(ctx, "hash_test_key", "field2", &intResult)
	require.NoError(t, err)
	assert.Equal(t, 123, intResult)

	cache.Delete(ctx, "hash_test_key")
}

func TestCache_TTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Skipf("Failed to connect to redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	err = cache.Set(ctx, "ttl_test_key", "test_value", 10*time.Second)
	require.NoError(t, err)

	ttl, err := cache.TTL(ctx, "ttl_test_key")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, ttl.Seconds(), float64(9))
	assert.LessOrEqual(t, ttl.Seconds(), float64(10))

	cache.Delete(ctx, "ttl_test_key")
}

func TestCache_SetMulti_GetMulti(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Skipf("Failed to connect to redis: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	items := map[string]interface{}{
		"multi_key1": "value1",
		"multi_key2": "value2",
		"multi_key3": "value3",
	}

	err = cache.SetMulti(ctx, items, time.Minute)
	require.NoError(t, err)

	results, err := cache.GetMulti(ctx, []string{"multi_key1", "multi_key2", "multi_key3"})
	require.NoError(t, err)
	assert.Len(t, results, 3)

	cache.Delete(ctx, "multi_key1", "multi_key2", "multi_key3")
}
