package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockCache 缓存Mock
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockCache) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCache) Incr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCache) Decr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCache) HSet(ctx context.Context, key string, field string, value interface{}) error {
	args := m.Called(ctx, key, field, value)
	return args.Error(0)
}

func (m *MockCache) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

func (m *MockCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockCache) HDel(ctx context.Context, key string, fields ...string) error {
	args := m.Called(ctx, key, fields)
	return args.Error(0)
}

func (m *MockCache) LPush(ctx context.Context, key string, values ...interface{}) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}

func (m *MockCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}

func (m *MockCache) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	args := m.Called(ctx, key, start, stop)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCache) LPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) RPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Publish(ctx context.Context, channel string, message interface{}) error {
	args := m.Called(ctx, channel, message)
	return args.Error(0)
}

func (m *MockCache) ZAdd(ctx context.Context, key string, score float64, member string) error {
	args := m.Called(ctx, key, score, member)
	return args.Error(0)
}

func (m *MockCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	args := m.Called(ctx, key, start, stop)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCache) ZRangeByScore(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error) {
	args := m.Called(ctx, key, min, max, offset, count)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCache) ZRem(ctx context.Context, key string, members ...interface{}) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewCacheService(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix:        "test:",
			DefaultExpiration: 5 * time.Minute,
		},
	}
	logger := zap.NewNop()

	service := NewCacheService(mockCache, cfg, logger)
	assert.NotNil(t, service)
	assert.Equal(t, mockCache, service.cache)
	assert.Equal(t, cfg, service.config)
	assert.Equal(t, logger, service.logger)
}

func TestCacheService_Get(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	mockCache.On("Get", mock.Anything, "test:key").Return("value", nil)

	value, err := service.Get(context.Background(), "key")
	assert.NoError(t, err)
	assert.Equal(t, "value", value)
	mockCache.AssertExpectations(t)
}

func TestCacheService_Set(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix:        "test:",
			DefaultExpiration: 5 * time.Minute,
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	mockCache.On("Set", mock.Anything, "test:key", "value", 5*time.Minute).Return(nil)

	err := service.Set(context.Background(), "key", "value")
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCacheService_SetWithCustomExpiration(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix:        "test:",
			DefaultExpiration: 5 * time.Minute,
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	customTTL := 10 * time.Minute
	mockCache.On("Set", mock.Anything, "test:key", "value", customTTL).Return(nil)

	err := service.Set(context.Background(), "key", "value", customTTL)
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCacheService_GetJSON(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	var dest map[string]string
	mockCache.On("GetJSON", mock.Anything, "test:key", &dest).Return(nil)

	err := service.GetJSON(context.Background(), "key", &dest)
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCacheService_SetJSON(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix:        "test:",
			DefaultExpiration: 5 * time.Minute,
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	value := map[string]string{"key": "value"}
	mockCache.On("SetJSON", mock.Anything, "test:key", value, 5*time.Minute).Return(nil)

	err := service.SetJSON(context.Background(), "key", value)
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCacheService_Delete(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	mockCache.On("Del", mock.Anything, []string{"test:key1", "test:key2"}).Return(nil)

	err := service.Delete(context.Background(), "key1", "key2")
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCacheService_Exists(t *testing.T) {
	tests := []struct {
		name      string
		mockCount int64
		mockErr   error
		want      bool
		wantErr   bool
	}{
		{
			name:      "键存在",
			mockCount: 1,
			mockErr:   nil,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "键不存在",
			mockCount: 0,
			mockErr:   nil,
			want:      false,
			wantErr:   false,
		},
		{
			name:      "错误",
			mockCount: 0,
			mockErr:   errors.New("redis error"),
			want:      false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCache := new(MockCache)
			cfg := &config.CacheConfig{
				Redis: config.RedisCacheConfig{
					KeyPrefix: "test:",
				},
			}
			service := NewCacheService(mockCache, cfg, zap.NewNop())

			mockCache.On("Exists", mock.Anything, []string{"test:key"}).Return(tt.mockCount, tt.mockErr)

			exists, err := service.Exists(context.Background(), "key")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, exists)
			}
			mockCache.AssertExpectations(t)
		})
	}
}

func TestCacheService_Expire(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	expiration := 10 * time.Minute
	mockCache.On("Expire", mock.Anything, "test:key", expiration).Return(nil)

	err := service.Expire(context.Background(), "key", expiration)
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCacheService_TTL(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	expectedTTL := 5 * time.Minute
	mockCache.On("TTL", mock.Anything, "test:key").Return(expectedTTL, nil)

	ttl, err := service.TTL(context.Background(), "key")
	assert.NoError(t, err)
	assert.Equal(t, expectedTTL, ttl)
	mockCache.AssertExpectations(t)
}

func TestCacheService_Increment(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	mockCache.On("Incr", mock.Anything, "test:counter").Return(int64(1), nil)

	result, err := service.Increment(context.Background(), "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result)
	mockCache.AssertExpectations(t)
}

func TestCacheService_IncrementBy(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		times int
	}{
		{
			name:  "增加1",
			value: 1,
			times: 1,
		},
		{
			name:  "增加5",
			value: 5,
			times: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCache := new(MockCache)
			cfg := &config.CacheConfig{
				Redis: config.RedisCacheConfig{
					KeyPrefix: "test:",
				},
			}
			service := NewCacheService(mockCache, cfg, zap.NewNop())

			for i := 0; i < tt.times; i++ {
				mockCache.On("Incr", mock.Anything, "test:counter").Return(int64(i+1), nil).Once()
			}

			result, err := service.IncrementBy(context.Background(), "counter", tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.value, result)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestCacheService_Decrement(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	mockCache.On("Decr", mock.Anything, "test:counter").Return(int64(9), nil)

	result, err := service.Decrement(context.Background(), "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(9), result)
	mockCache.AssertExpectations(t)
}

func TestCacheService_DecrementBy(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	mockCache.On("Decr", mock.Anything, "test:counter").Return(int64(9), nil).Once()
	mockCache.On("Decr", mock.Anything, "test:counter").Return(int64(8), nil).Once()

	result, err := service.DecrementBy(context.Background(), "counter", 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(8), result)
	mockCache.AssertExpectations(t)
}

func TestCacheService_HashSet(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	mockCache.On("HSet", mock.Anything, "test:hash", "field", "value").Return(nil)

	err := service.HashSet(context.Background(), "hash", "field", "value")
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCacheService_HashGet(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	mockCache.On("HGet", mock.Anything, "test:hash", "field").Return("value", nil)

	value, err := service.HashGet(context.Background(), "hash", "field")
	assert.NoError(t, err)
	assert.Equal(t, "value", value)
	mockCache.AssertExpectations(t)
}

func TestCacheService_HashGetAll(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	expected := map[string]string{"field1": "value1", "field2": "value2"}
	mockCache.On("HGetAll", mock.Anything, "test:hash").Return(expected, nil)

	result, err := service.HashGetAll(context.Background(), "hash")
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockCache.AssertExpectations(t)
}

func TestCacheService_HashDelete(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	mockCache.On("HDel", mock.Anything, "test:hash", []string{"field1", "field2"}).Return(nil)

	err := service.HashDelete(context.Background(), "hash", "field1", "field2")
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCacheService_ListPush(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	values := []interface{}{"value1", "value2"}
	mockCache.On("LPush", mock.Anything, "test:list", values).Return(nil)

	err := service.ListPush(context.Background(), "list", values...)
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCacheService_ListPop(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	mockCache.On("LPop", mock.Anything, "test:list").Return("value", nil)

	value, err := service.ListPop(context.Background(), "list")
	assert.NoError(t, err)
	assert.Equal(t, "value", value)
	mockCache.AssertExpectations(t)
}

func TestCacheService_ListRange(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	expected := []string{"value1", "value2", "value3"}
	mockCache.On("LRange", mock.Anything, "test:list", int64(0), int64(-1)).Return(expected, nil)

	result, err := service.ListRange(context.Background(), "list", 0, -1)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockCache.AssertExpectations(t)
}

func TestCacheService_SortedSetAdd(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	mockCache.On("ZAdd", mock.Anything, "test:zset", 1.0, "member").Return(nil)

	err := service.SortedSetAdd(context.Background(), "zset", 1.0, "member")
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCacheService_SortedSetRange(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	expected := []string{"member1", "member2"}
	mockCache.On("ZRange", mock.Anything, "test:zset", int64(0), int64(-1)).Return(expected, nil)

	result, err := service.SortedSetRange(context.Background(), "zset", 0, -1)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockCache.AssertExpectations(t)
}

func TestCacheService_SortedSetRemove(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	members := []interface{}{"member1", "member2"}
	mockCache.On("ZRem", mock.Anything, "test:zset", members).Return(nil)

	err := service.SortedSetRemove(context.Background(), "zset", members...)
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCacheService_GetOrSet(t *testing.T) {
	tests := []struct {
		name         string
		cacheValue   string
		cacheErr     error
		fnValue      interface{}
		fnErr        error
		expected     interface{}
		wantErr      bool
		expectSet    bool
	}{
		{
			name:       "缓存命中",
			cacheValue: "cached_value",
			cacheErr:   nil,
			expected:   "cached_value",
			wantErr:    false,
			expectSet:  false,
		},
		{
			name:       "缓存未命中",
			cacheValue: "",
			cacheErr:   errors.New("not found"),
			fnValue:    "new_value",
			expected:   "new_value",
			wantErr:    false,
			expectSet:  true,
		},
		{
			name:      "函数错误",
			cacheErr:  errors.New("not found"),
			fnErr:     errors.New("function error"),
			wantErr:   true,
			expectSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCache := new(MockCache)
			cfg := &config.CacheConfig{
				Redis: config.RedisCacheConfig{
					KeyPrefix:        "test:",
					DefaultExpiration: 5 * time.Minute,
				},
			}
			service := NewCacheService(mockCache, cfg, zap.NewNop())

			mockCache.On("Get", mock.Anything, "test:key").Return(tt.cacheValue, tt.cacheErr)

			if tt.expectSet {
				mockCache.On("Set", mock.Anything, "test:key", tt.fnValue, 5*time.Minute).Return(nil)
			}

			result, err := service.GetOrSet(context.Background(), "key", func() (interface{}, error) {
				return tt.fnValue, tt.fnErr
			})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockCache.AssertExpectations(t)
		})
	}
}

func TestCacheService_Remember(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix:        "test:",
			DefaultExpiration: 5 * time.Minute,
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	// 缓存未命中
	mockCache.On("Get", mock.Anything, "test:key").Return("", errors.New("not found"))
	mockCache.On("Set", mock.Anything, "test:key", "value", 5*time.Minute).Return(nil)
	mockCache.On("RPush", mock.Anything, "test:tag:mytag", mock.AnythingOfType("[]interface {}")).Return(nil)
	mockCache.On("Expire", mock.Anything, "test:tag:mytag", 5*time.Minute).Return(nil)

	result, err := service.Remember(context.Background(), "key", []string{"mytag"}, func() (interface{}, error) {
		return "value", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "value", result)
	mockCache.AssertExpectations(t)
}

func TestCacheService_Forget(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	keys := []string{"test:key1", "test:key2"}
	mockCache.On("LRange", mock.Anything, "test:tag:mytag", int64(0), int64(-1)).Return(keys, nil)
	mockCache.On("Del", mock.Anything, keys).Return(nil)
	mockCache.On("Del", mock.Anything, []string{"test:tag:mytag"}).Return(nil)

	err := service.Forget(context.Background(), "mytag")
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestCacheService_Flush(t *testing.T) {
	mockCache := new(MockCache)
	cfg := &config.CacheConfig{
		Redis: config.RedisCacheConfig{
			KeyPrefix: "test:",
		},
	}
	service := NewCacheService(mockCache, cfg, zap.NewNop())

	mockCache.On("Close").Return(nil)

	err := service.Flush(context.Background())
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}
