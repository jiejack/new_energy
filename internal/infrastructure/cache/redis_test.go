package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMiniredis(t *testing.T) (*miniredis.Miniredis, *RedisClient) {
	t.Helper()
	mr := miniredis.RunT(t)
	client, err := NewRedisClient(RedisConfig{
		Addrs: []string{mr.Addr()},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		client.Close()
	})
	return mr, client
}

func TestNewRedisClient(t *testing.T) {
	mr := miniredis.RunT(t)
	client, err := NewRedisClient(RedisConfig{
		Addrs: []string{mr.Addr()},
	})
	require.NoError(t, err)
	require.NotNil(t, client)
	client.Close()
}

func TestNewRedisClient_ConnectionError(t *testing.T) {
	_, err := NewRedisClient(RedisConfig{
		Addrs:       []string{"localhost:19999"},
		DialTimeout: 1 * time.Second,
	})
	assert.Error(t, err)
}

func TestRedisClient_Ping(t *testing.T) {
	_, client := setupMiniredis(t)
	err := client.Ping(context.Background())
	require.NoError(t, err)
}

func TestRedisClient_SetGet(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	err := client.Set(ctx, "key1", "value1", 0)
	require.NoError(t, err)

	val, err := client.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)

	_, err = client.Get(ctx, "nonexistent")
	assert.Equal(t, redis.Nil, err)
}

func TestRedisClient_SetJSON_GetJSON(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := TestStruct{Name: "test", Value: 42}
	err := client.SetJSON(ctx, "json_key", data, 0)
	require.NoError(t, err)

	var result TestStruct
	err = client.GetJSON(ctx, "json_key", &result)
	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, 42, result.Value)
}

func TestRedisClient_Del(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	client.Set(ctx, "key1", "value1", 0)
	client.Set(ctx, "key2", "value2", 0)

	err := client.Del(ctx, "key1", "key2")
	require.NoError(t, err)

	_, err = client.Get(ctx, "key1")
	assert.Equal(t, redis.Nil, err)
}

func TestRedisClient_Exists(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	client.Set(ctx, "key1", "value1", 0)

	count, err := client.Exists(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = client.Exists(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRedisClient_Expire_TTL(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	client.Set(ctx, "key1", "value1", 0)

	err := client.Expire(ctx, "key1", 10*time.Minute)
	require.NoError(t, err)

	ttl, err := client.TTL(ctx, "key1")
	require.NoError(t, err)
	assert.True(t, ttl > 0)
}

func TestRedisClient_Incr_Decr(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	val, err := client.Incr(ctx, "counter")
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)

	val, err = client.Incr(ctx, "counter")
	require.NoError(t, err)
	assert.Equal(t, int64(2), val)

	val, err = client.Decr(ctx, "counter")
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)
}

func TestRedisClient_HSet_HGet_HGetAll_HDel(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	err := client.HSet(ctx, "hash1", "field1", "value1")
	require.NoError(t, err)

	val, err := client.HGet(ctx, "hash1", "field1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)

	client.HSet(ctx, "hash1", "field2", "value2")

	all, err := client.HGetAll(ctx, "hash1")
	require.NoError(t, err)
	assert.Equal(t, 2, len(all))

	err = client.HDel(ctx, "hash1", "field1")
	require.NoError(t, err)

	all, err = client.HGetAll(ctx, "hash1")
	require.NoError(t, err)
	assert.Equal(t, 1, len(all))
}

func TestRedisClient_ListOps(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	err := client.LPush(ctx, "list1", "a", "b")
	require.NoError(t, err)

	err = client.RPush(ctx, "list1", "c")
	require.NoError(t, err)

	items, err := client.LRange(ctx, "list1", 0, -1)
	require.NoError(t, err)
	assert.Equal(t, 3, len(items))

	val, err := client.LPop(ctx, "list1")
	require.NoError(t, err)
	assert.NotEmpty(t, val)

	val, err = client.RPop(ctx, "list1")
	require.NoError(t, err)
	assert.NotEmpty(t, val)
}

func TestRedisClient_ZAdd_ZRange_ZRangeByScore_ZRem(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	err := client.ZAdd(ctx, "zset1", 1.0, "member1")
	require.NoError(t, err)

	err = client.ZAdd(ctx, "zset1", 2.0, "member2")
	require.NoError(t, err)

	err = client.ZAdd(ctx, "zset1", 3.0, "member3")
	require.NoError(t, err)

	members, err := client.ZRange(ctx, "zset1", 0, -1)
	require.NoError(t, err)
	assert.Equal(t, 3, len(members))

	members, err = client.ZRangeByScore(ctx, "zset1", "1", "2", 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, len(members))

	err = client.ZRem(ctx, "zset1", "member1")
	require.NoError(t, err)

	members, err = client.ZRange(ctx, "zset1", 0, -1)
	require.NoError(t, err)
	assert.Equal(t, 2, len(members))
}

func TestRedisClient_Keys_Scan(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	client.Set(ctx, "nem:test:1", "v1", 0)
	client.Set(ctx, "nem:test:2", "v2", 0)

	keys, err := client.Keys(ctx, "nem:test:*")
	require.NoError(t, err)
	assert.Equal(t, 2, len(keys))

	var allKeys []string
	var cursor uint64
	for {
		var batch []string
		batch, cursor, err = client.Scan(ctx, cursor, "nem:test:*", 10)
		require.NoError(t, err)
		allKeys = append(allKeys, batch...)
		if cursor == 0 {
			break
		}
	}
	assert.Equal(t, 2, len(allKeys))
}

func TestRedisClient_FlushDB(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	client.Set(ctx, "key1", "value1", 0)
	err := client.FlushDB(ctx)
	require.NoError(t, err)

	_, err = client.Get(ctx, "key1")
	assert.Equal(t, redis.Nil, err)
}

func TestRedisClient_Publish_Subscribe(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	sub := client.Subscribe(ctx, "test_channel")
	defer sub.Close()

	err := client.Publish(ctx, "test_channel", "hello")
	require.NoError(t, err)
}

func TestRedisClient_GetPoolStats(t *testing.T) {
	_, client := setupMiniredis(t)
	stats := client.GetPoolStats()
	assert.NotNil(t, stats)
}

func TestRedisClient_Stats(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	client.Set(ctx, "key1", "value1", 0)
	client.Get(ctx, "key1")
	client.Get(ctx, "nonexistent")

	stats := client.GetStats()
	assert.Equal(t, int64(1), stats.Sets)
	assert.Equal(t, int64(2), stats.Gets)
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)

	hitRate := client.GetHitRate()
	assert.InDelta(t, 0.5, hitRate, 0.01)

	client.ResetStats()
	stats = client.GetStats()
	assert.Equal(t, int64(0), stats.Sets)
	assert.Equal(t, int64(0), stats.Gets)
}

func TestRedisClient_GetHitRate_NoOps(t *testing.T) {
	_, client := setupMiniredis(t)
	hitRate := client.GetHitRate()
	assert.Equal(t, float64(0), hitRate)
}

func TestRedisClient_RealtimeData(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	err := client.SetRealtimeData(ctx, "point1", 123.45, 1700000000)
	require.NoError(t, err)

	value, timestamp, err := client.GetRealtimeData(ctx, "point1")
	require.NoError(t, err)
	assert.InDelta(t, 123.45, value, 0.01)
	assert.Equal(t, int64(1700000000), timestamp)
}

func TestRedisClient_DeviceStatus(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	err := client.SetDeviceStatus(ctx, "device1", 1)
	require.NoError(t, err)

	status, err := client.GetDeviceStatus(ctx, "device1")
	require.NoError(t, err)
	assert.Equal(t, 1, status)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "nem:realtime:", RealtimeDataPrefix)
	assert.Equal(t, "nem:alarm:active:", AlarmActivePrefix)
	assert.Equal(t, "nem:alarm:count", AlarmCountKey)
	assert.Equal(t, "nem:device:status:", DeviceStatusPrefix)
}

func TestCacheStats_Struct(t *testing.T) {
	stats := CacheStats{
		Hits:      10,
		Misses:    5,
		Sets:      3,
		Gets:      15,
		Dels:      2,
		Errors:    0,
		TotalTime: 1000,
	}
	assert.Equal(t, int64(10), stats.Hits)
	assert.Equal(t, int64(5), stats.Misses)
}

func TestRedisConfig_Defaults(t *testing.T) {
	mr := miniredis.RunT(t)
	client, err := NewRedisClient(RedisConfig{
		Addrs: []string{mr.Addr()},
	})
	require.NoError(t, err)
	client.Close()
}
