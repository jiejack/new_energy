package cache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupMiddleware(t *testing.T) (*miniredis.Miniredis, *RedisClient, *CacheMiddleware) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mr := miniredis.RunT(t)
	client, err := NewRedisClient(RedisConfig{
		Addrs: []string{mr.Addr()},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		client.Close()
	})

	cfg := config.DefaultCacheConfig().API
	mw := NewCacheMiddleware(client, &cfg, zap.NewNop())
	return mr, client, mw
}

func TestNewCacheMiddleware(t *testing.T) {
	_, _, mw := setupMiddleware(t)
	require.NotNil(t, mw)
}

func TestCacheMiddleware_Handler_Disabled(t *testing.T) {
	_, _, mw := setupMiddleware(t)
	mw.config.Enabled = false

	r := gin.New()
	r.Use(mw.Handler())
	r.GET("/api/v1/devices", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/devices", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestCacheMiddleware_Handler_CacheMiss(t *testing.T) {
	_, _, mw := setupMiddleware(t)

	r := gin.New()
	r.Use(mw.Handler())
	r.GET("/api/v1/devices", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/devices", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestCacheMiddleware_Handler_CacheHit(t *testing.T) {
	mr, _, mw := setupMiddleware(t)

	r := gin.New()
	r.Use(mw.Handler())
	r.GET("/api/v1/devices", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/devices", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	mr.FastForward(0)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/devices", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
}

func TestCacheMiddleware_Handler_PostRequest(t *testing.T) {
	_, _, mw := setupMiddleware(t)

	r := gin.New()
	r.Use(mw.Handler())
	r.POST("/api/v1/devices", func(c *gin.Context) {
		c.JSON(201, gin.H{"status": "created"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/devices", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)
}

func TestCacheMiddleware_Handler_ExcludedPath(t *testing.T) {
	_, _, mw := setupMiddleware(t)

	r := gin.New()
	r.Use(mw.Handler())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestCacheMiddleware_InvalidateByPattern(t *testing.T) {
	_, client, mw := setupMiddleware(t)
	ctx := context.Background()

	client.Set(ctx, "api:test1", "value1", 0)
	client.Set(ctx, "api:test2", "value2", 0)

	err := mw.InvalidateByPattern(ctx, "test*")
	require.NoError(t, err)
}

func TestCacheMiddleware_InvalidateByTags(t *testing.T) {
	_, client, mw := setupMiddleware(t)
	ctx := context.Background()

	client.LPush(ctx, "api:tag:devices", "api:device1")
	client.Set(ctx, "api:device1", "value1", 0)

	err := mw.InvalidateByTags(ctx, "devices")
	require.NoError(t, err)
}

func TestCacheMiddleware_CacheStatsMiddleware(t *testing.T) {
	_, _, mw := setupMiddleware(t)

	r := gin.New()
	r.Use(mw.CacheStatsMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestCacheMiddleware_WarmupCache_Disabled(t *testing.T) {
	_, _, mw := setupMiddleware(t)
	mw.config.Warmup.Enabled = false

	err := mw.WarmupCache(context.Background(), []config.WarmupURL{
		{Path: "http://localhost:8080/api/v1/devices", Method: "GET"},
	})
	require.NoError(t, err)
}

func TestCacheMiddleware_WarmupCache_Enabled(t *testing.T) {
	_, _, mw := setupMiddleware(t)
	mw.config.Warmup.Enabled = true

	err := mw.WarmupCache(context.Background(), []config.WarmupURL{
		{Path: "http://localhost:19999/nonexistent", Method: "GET"},
	})
	require.NoError(t, err)
}

func TestCacheMiddleware_GetCacheStats(t *testing.T) {
	_, _, mw := setupMiddleware(t)
	stats := mw.GetCacheStats()
	assert.NotNil(t, stats)
}

func TestCacheMiddleware_ResetCacheStats(t *testing.T) {
	_, _, mw := setupMiddleware(t)
	mw.ResetCacheStats()
	stats := mw.GetCacheStats()
	assert.Equal(t, int64(0), stats.Sets)
}

func TestCachedResponse(t *testing.T) {
	resp := &CachedResponse{
		StatusCode:  200,
		Headers:     http.Header{"Content-Type": []string{"application/json"}},
		Body:        []byte(`{"status":"ok"}`),
		ContentType: "application/json",
	}
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/json", resp.ContentType)
}
