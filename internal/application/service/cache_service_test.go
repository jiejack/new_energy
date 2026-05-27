package service

import (
	"context"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/cache"
	"github.com/new-energy-monitoring/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockCache struct {
	mock.Mock
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}
func (m *mockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}
func (m *mockCache) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}
func (m *mockCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}
func (m *mockCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}
func (m *mockCache) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}
func (m *mockCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}
func (m *mockCache) Incr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockCache) Decr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockCache) HSet(ctx context.Context, key string, field string, value interface{}) error {
	args := m.Called(ctx, key, field, value)
	return args.Error(0)
}
func (m *mockCache) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}
func (m *mockCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}
func (m *mockCache) HDel(ctx context.Context, key string, fields ...string) error {
	args := m.Called(ctx, key, fields)
	return args.Error(0)
}
func (m *mockCache) LPush(ctx context.Context, key string, values ...interface{}) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}
func (m *mockCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}
func (m *mockCache) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	args := m.Called(ctx, key, start, stop)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockCache) LPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}
func (m *mockCache) RPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}
func (m *mockCache) Publish(ctx context.Context, channel string, message interface{}) error {
	args := m.Called(ctx, channel, message)
	return args.Error(0)
}
func (m *mockCache) ZAdd(ctx context.Context, key string, score float64, member string) error {
	args := m.Called(ctx, key, score, member)
	return args.Error(0)
}
func (m *mockCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	args := m.Called(ctx, key, start, stop)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockCache) ZRangeByScore(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error) {
	args := m.Called(ctx, key, min, max, offset, count)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockCache) ZRem(ctx context.Context, key string, members ...interface{}) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}
func (m *mockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

func setupCacheService(c *mockCache) *CacheService {
	cfg := config.DefaultCacheConfig()
	return NewCacheService(c, &cfg, zap.NewNop())
}

func TestCacheService_Get_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Get", mock.Anything, "nem:cache:test-key").Return("test-value", nil)
	val, err := svc.Get(context.Background(), "test-key")
	assert.NoError(t, err)
	assert.Equal(t, "test-value", val)
}

func TestCacheService_Set_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Set", mock.Anything, "nem:cache:test-key", "test-value", 5*time.Minute).Return(nil)
	err := svc.Set(context.Background(), "test-key", "test-value")
	assert.NoError(t, err)
}

func TestCacheService_Set_WithExpiration(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Set", mock.Anything, "nem:cache:test-key", "test-value", 10*time.Minute).Return(nil)
	err := svc.Set(context.Background(), "test-key", "test-value", 10*time.Minute)
	assert.NoError(t, err)
}

func TestCacheService_Delete_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Del", mock.Anything, []string{"nem:cache:key1", "nem:cache:key2"}).Return(nil)
	err := svc.Delete(context.Background(), "key1", "key2")
	assert.NoError(t, err)
}

func TestCacheService_Exists_True(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Exists", mock.Anything, []string{"nem:cache:test-key"}).Return(int64(1), nil)
	exists, err := svc.Exists(context.Background(), "test-key")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestCacheService_Exists_False(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Exists", mock.Anything, []string{"nem:cache:test-key"}).Return(int64(0), nil)
	exists, err := svc.Exists(context.Background(), "test-key")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestCacheService_Expire_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Expire", mock.Anything, "nem:cache:test-key", 10*time.Minute).Return(nil)
	err := svc.Expire(context.Background(), "test-key", 10*time.Minute)
	assert.NoError(t, err)
}

func TestCacheService_TTL_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("TTL", mock.Anything, "nem:cache:test-key").Return(5*time.Minute, nil)
	ttl, err := svc.TTL(context.Background(), "test-key")
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Minute, ttl)
}

func TestCacheService_Increment_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Incr", mock.Anything, "nem:cache:counter").Return(int64(1), nil)
	val, err := svc.Increment(context.Background(), "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)
}

func TestCacheService_IncrementBy_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Incr", mock.Anything, "nem:cache:counter").Return(int64(1), nil)
	c.On("Incr", mock.Anything, "nem:cache:counter").Return(int64(2), nil)
	c.On("Incr", mock.Anything, "nem:cache:counter").Return(int64(3), nil)
	val, err := svc.IncrementBy(context.Background(), "counter", 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), val)
}

func TestCacheService_Decrement_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Decr", mock.Anything, "nem:cache:counter").Return(int64(9), nil)
	val, err := svc.Decrement(context.Background(), "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(9), val)
}

func TestCacheService_DecrementBy_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Decr", mock.Anything, "nem:cache:counter").Return(int64(9), nil)
	c.On("Decr", mock.Anything, "nem:cache:counter").Return(int64(8), nil)
	c.On("Decr", mock.Anything, "nem:cache:counter").Return(int64(7), nil)
	val, err := svc.DecrementBy(context.Background(), "counter", 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(7), val)
}

func TestCacheService_HashSet_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("HSet", mock.Anything, "nem:cache:hash-key", "field1", "value1").Return(nil)
	err := svc.HashSet(context.Background(), "hash-key", "field1", "value1")
	assert.NoError(t, err)
}

func TestCacheService_HashGet_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("HGet", mock.Anything, "nem:cache:hash-key", "field1").Return("value1", nil)
	val, err := svc.HashGet(context.Background(), "hash-key", "field1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)
}

func TestCacheService_HashGetAll_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("HGetAll", mock.Anything, "nem:cache:hash-key").Return(map[string]string{"f1": "v1"}, nil)
	val, err := svc.HashGetAll(context.Background(), "hash-key")
	assert.NoError(t, err)
	assert.Equal(t, "v1", val["f1"])
}

func TestCacheService_HashDelete_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("HDel", mock.Anything, "nem:cache:hash-key", []string{"field1"}).Return(nil)
	err := svc.HashDelete(context.Background(), "hash-key", "field1")
	assert.NoError(t, err)
}

func TestCacheService_ListPush_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("LPush", mock.Anything, "nem:cache:list-key", mock.Anything).Return(nil)
	err := svc.ListPush(context.Background(), "list-key", "value1")
	assert.NoError(t, err)
}

func TestCacheService_ListPop_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("LPop", mock.Anything, "nem:cache:list-key").Return("value1", nil)
	val, err := svc.ListPop(context.Background(), "list-key")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)
}

func TestCacheService_ListRange_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("LRange", mock.Anything, "nem:cache:list-key", int64(0), int64(-1)).Return([]string{"v1", "v2"}, nil)
	val, err := svc.ListRange(context.Background(), "list-key", 0, -1)
	assert.NoError(t, err)
	assert.Len(t, val, 2)
}

func TestCacheService_SortedSetAdd_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("ZAdd", mock.Anything, "nem:cache:zset-key", 1.0, "member1").Return(nil)
	err := svc.SortedSetAdd(context.Background(), "zset-key", 1.0, "member1")
	assert.NoError(t, err)
}

func TestCacheService_SortedSetRange_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("ZRange", mock.Anything, "nem:cache:zset-key", int64(0), int64(-1)).Return([]string{"m1"}, nil)
	val, err := svc.SortedSetRange(context.Background(), "zset-key", 0, -1)
	assert.NoError(t, err)
	assert.Len(t, val, 1)
}

func TestCacheService_SortedSetRemove_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("ZRem", mock.Anything, "nem:cache:zset-key", mock.Anything).Return(nil)
	err := svc.SortedSetRemove(context.Background(), "zset-key", "member1")
	assert.NoError(t, err)
}

func TestCacheService_GetOrSet_CacheHit(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Get", mock.Anything, "nem:cache:test-key").Return("cached-value", nil)
	val, err := svc.GetOrSet(context.Background(), "test-key", func() (interface{}, error) {
		return "fresh-value", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "cached-value", val)
}

func TestCacheService_GetOrSet_CacheMiss(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Get", mock.Anything, "nem:cache:test-key").Return("", assert.AnError)
	c.On("Set", mock.Anything, "nem:cache:test-key", "fresh-value", 5*time.Minute).Return(nil)
	val, err := svc.GetOrSet(context.Background(), "test-key", func() (interface{}, error) {
		return "fresh-value", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "fresh-value", val)
}

func TestCacheService_GetOrSet_FnError(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Get", mock.Anything, "nem:cache:test-key").Return("", assert.AnError)
	val, err := svc.GetOrSet(context.Background(), "test-key", func() (interface{}, error) {
		return nil, assert.AnError
	})
	assert.Error(t, err)
	assert.Nil(t, val)
}

func TestCacheService_Flush_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Close").Return(nil)
	err := svc.Flush(context.Background())
	assert.NoError(t, err)
}

func TestCacheService_GetJSON_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("GetJSON", mock.Anything, "nem:cache:test-key", mock.Anything).Return(nil)
	err := svc.GetJSON(context.Background(), "test-key", &map[string]interface{}{})
	assert.NoError(t, err)
}

func TestCacheService_SetJSON_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("SetJSON", mock.Anything, "nem:cache:test-key", mock.Anything, 5*time.Minute).Return(nil)
	err := svc.SetJSON(context.Background(), "test-key", map[string]string{"k": "v"})
	assert.NoError(t, err)
}

func TestCopyData_Success(t *testing.T) {
	src := map[string]string{"key": "value"}
	dest := make(map[string]string)
	err := copyData(src, &dest)
	assert.NoError(t, err)
	assert.Equal(t, "value", dest["key"])
}

func TestCacheService_Remember_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Get", mock.Anything, "nem:cache:test-key").Return("", assert.AnError)
	c.On("Set", mock.Anything, "nem:cache:test-key", "fresh-value", 5*time.Minute).Return(nil)
	c.On("RPush", mock.Anything, "nem:cache:tag:mytag", mock.Anything).Return(nil)
	c.On("Expire", mock.Anything, "nem:cache:tag:mytag", 5*time.Minute).Return(nil)
	val, err := svc.Remember(context.Background(), "test-key", []string{"mytag"}, func() (interface{}, error) {
		return "fresh-value", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "fresh-value", val)
}

func TestCacheService_Remember_CacheHit(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("Get", mock.Anything, "nem:cache:test-key").Return("cached-value", nil)
	val, err := svc.Remember(context.Background(), "test-key", []string{"mytag"}, func() (interface{}, error) {
		return "fresh-value", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "cached-value", val)
}

func TestCacheService_Forget_Success(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("LRange", mock.Anything, "nem:cache:tag:mytag", int64(0), int64(-1)).Return([]string{"nem:cache:key1"}, nil)
	c.On("Del", mock.Anything, []string{"nem:cache:key1"}).Return(nil)
	c.On("Del", mock.Anything, []string{"nem:cache:tag:mytag"}).Return(nil)
	err := svc.Forget(context.Background(), "mytag")
	assert.NoError(t, err)
}

func TestCacheService_Forget_EmptyTag(t *testing.T) {
	c := new(mockCache)
	svc := setupCacheService(c)
	c.On("LRange", mock.Anything, "nem:cache:tag:empty", int64(0), int64(-1)).Return([]string{}, nil)
	err := svc.Forget(context.Background(), "empty")
	assert.NoError(t, err)
}

var _ cache.Cache = (*mockCache)(nil)
