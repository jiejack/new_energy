package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultCacheConfig(t *testing.T) {
	cfg := DefaultCacheConfig()
	assert.True(t, cfg.API.Enabled)
	assert.Equal(t, 5*time.Minute, cfg.API.DefaultTTL)
	assert.Equal(t, "api:", cfg.API.KeyPrefix)
	assert.True(t, cfg.API.CacheQueryParams)
	assert.Equal(t, int64(10*1024*1024), cfg.API.MaxCacheSize)
	assert.Equal(t, "nem:cache:", cfg.Redis.KeyPrefix)
	assert.Equal(t, 5*time.Minute, cfg.Redis.DefaultExpiration)
	assert.False(t, cfg.API.Warmup.Enabled)
}

func TestAPICacheConfig_GetTTLForPath(t *testing.T) {
	cfg := DefaultCacheConfig()

	ttl := cfg.API.GetTTLForPath("/api/v1/devices")
	assert.Equal(t, 5*time.Minute, ttl)

	ttl = cfg.API.GetTTLForPath("/api/v1/regions")
	assert.Equal(t, 10*time.Minute, ttl)

	ttl = cfg.API.GetTTLForPath("/api/v1/alarms")
	assert.Equal(t, 1*time.Minute, ttl)

	ttl = cfg.API.GetTTLForPath("/api/v1/nonexistent")
	assert.Equal(t, time.Duration(0), ttl)
}

func TestAPICacheConfig_ShouldCachePath(t *testing.T) {
	cfg := DefaultCacheConfig()

	assert.True(t, cfg.API.ShouldCachePath("/api/v1/devices"))
	assert.True(t, cfg.API.ShouldCachePath("/api/v1/stations"))
	assert.False(t, cfg.API.ShouldCachePath("/health"))
	assert.False(t, cfg.API.ShouldCachePath("/ready"))
	assert.False(t, cfg.API.ShouldCachePath("/metrics"))
	assert.False(t, cfg.API.ShouldCachePath("/api/v1/nonexistent"))
}

func TestAPICacheConfig_GetInvalidatePaths(t *testing.T) {
	cfg := DefaultCacheConfig()

	paths := cfg.API.GetInvalidatePaths("POST", "/api/v1/devices")
	assert.Equal(t, 1, len(paths))
	assert.Equal(t, "/api/v1/devices", paths[0])

	paths = cfg.API.GetInvalidatePaths("GET", "/api/v1/devices")
	assert.Equal(t, 0, len(paths))

	paths = cfg.API.GetInvalidatePaths("DELETE", "/api/v1/alarms")
	assert.Equal(t, 1, len(paths))

	paths = cfg.API.GetInvalidatePaths("POST", "/api/v1/nonexistent")
	assert.Equal(t, 0, len(paths))
}

func TestCachePathPattern(t *testing.T) {
	pattern := CachePathPattern{
		Pattern:      "/api/v1/test",
		TTL:          10 * time.Minute,
		KeyStrategy:  "path_query",
		Enabled:      true,
		InvalidateOn: []string{"POST", "PUT", "DELETE"},
	}
	assert.Equal(t, "/api/v1/test", pattern.Pattern)
	assert.Equal(t, 10*time.Minute, pattern.TTL)
	assert.True(t, pattern.Enabled)
}

func TestWarmupConfig(t *testing.T) {
	cfg := WarmupConfig{
		Enabled:  true,
		Schedule: "0 3 * * *",
		URLs: []WarmupURL{
			{Path: "/api/v1/devices", Method: "GET"},
		},
	}
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "0 3 * * *", cfg.Schedule)
	assert.Equal(t, 1, len(cfg.URLs))
}

func TestWarmupURL(t *testing.T) {
	url := WarmupURL{
		Path:    "/api/v1/stations",
		Query:   "region=east",
		Method:  "GET",
		Headers: map[string]string{"Authorization": "Bearer token"},
		Delay:   5 * time.Second,
	}
	assert.Equal(t, "/api/v1/stations", url.Path)
	assert.Equal(t, "region=east", url.Query)
	assert.Equal(t, "GET", url.Method)
	assert.Equal(t, 5*time.Second, url.Delay)
}

func TestRedisCacheConfig(t *testing.T) {
	cfg := RedisCacheConfig{
		Addrs:           []string{"localhost:6379"},
		Password:        "secret",
		DB:              1,
		PoolSize:        50,
		MinIdleConns:    5,
		MaxRetries:      3,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolTimeout:     4 * time.Second,
		IdleTimeout:     5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		KeyPrefix:       "test:",
		DefaultExpiration: 10 * time.Minute,
	}
	assert.Equal(t, []string{"localhost:6379"}, cfg.Addrs)
	assert.Equal(t, "secret", cfg.Password)
	assert.Equal(t, 1, cfg.DB)
	assert.Equal(t, "test:", cfg.KeyPrefix)
}
