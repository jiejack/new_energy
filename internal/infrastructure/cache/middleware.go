package cache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/infrastructure/config"
	"go.uber.org/zap"
)

// CacheMiddleware 缓存中间件
type CacheMiddleware struct {
	redis  *RedisClient
	config *config.APICacheConfig
	logger *zap.Logger
}

// NewCacheMiddleware 创建缓存中间件
func NewCacheMiddleware(redis *RedisClient, cfg *config.APICacheConfig, logger *zap.Logger) *CacheMiddleware {
	return &CacheMiddleware{
		redis:  redis,
		config: cfg,
		logger: logger,
	}
}

// Handler 返回缓存中间件处理函数
func (m *CacheMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用缓存
		if !m.config.Enabled {
			c.Next()
			return
		}

		// 只缓存GET请求
		if c.Request.Method != http.MethodGet {
			// 对于写操作，失效相关缓存
			m.invalidateCache(c)
			c.Next()
			return
		}

		// 检查路径是否应该缓存
		path := c.Request.URL.Path
		if !m.config.ShouldCachePath(path) {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := m.generateCacheKey(c)

		// 尝试从缓存获取
		cachedResponse, err := m.getFromCache(c.Request.Context(), cacheKey)
		if err == nil && cachedResponse != nil {
			// 缓存命中
			m.logger.Debug("Cache hit",
				zap.String("key", cacheKey),
				zap.String("path", path),
			)

			// 设置响应头
			for key, values := range cachedResponse.Headers {
				for _, value := range values {
					c.Header(key, value)
				}
			}

			// 设置缓存标记
			c.Header("X-Cache", "HIT")
			c.Header("X-Cache-Key", cacheKey)

			// 返回缓存的响应
			c.Data(cachedResponse.StatusCode, cachedResponse.ContentType, cachedResponse.Body)
			return
		}

		// 缓存未命中，继续处理请求
		m.logger.Debug("Cache miss",
			zap.String("key", cacheKey),
			zap.String("path", path),
		)

		// 使用响应写入器捕获响应
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		c.Next()

		// 只缓存成功的响应
		if writer.status >= 200 && writer.status < 300 {
			// 获取缓存时间
			ttl := m.config.GetTTLForPath(path)
			if ttl > 0 {
				// 保存到缓存
				response := &CachedResponse{
					StatusCode:  writer.status,
					Headers:     writer.Header(),
					Body:        writer.body.Bytes(),
					ContentType: writer.Header().Get("Content-Type"),
					CachedAt:    time.Now(),
				}

				if err := m.saveToCache(c.Request.Context(), cacheKey, response, ttl); err != nil {
					m.logger.Error("Failed to save cache",
						zap.Error(err),
						zap.String("key", cacheKey),
					)
				} else {
					c.Header("X-Cache", "MISS")
					c.Header("X-Cache-Key", cacheKey)
				}
			}
		}
	}
}

// CachedResponse 缓存的响应
type CachedResponse struct {
	StatusCode  int              `json:"status_code"`
	Headers     http.Header      `json:"headers"`
	Body        []byte           `json:"body"`
	ContentType string           `json:"content_type"`
	CachedAt    time.Time        `json:"cached_at"`
}

// responseWriter 响应写入器
type responseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// generateCacheKey 生成缓存键
func (m *CacheMiddleware) generateCacheKey(c *gin.Context) string {
	var keyParts []string

	// 添加方法
	keyParts = append(keyParts, c.Request.Method)

	// 添加路径
	keyParts = append(keyParts, c.Request.URL.Path)

	// 根据配置添加查询参数
	if m.config.CacheQueryParams && c.Request.URL.RawQuery != "" {
		// 对查询参数排序以保证一致性
		query := c.Request.URL.Query()
		sortedQuery := make([]string, 0, len(query))
		for key, values := range query {
			sort.Strings(values)
			for _, value := range values {
				sortedQuery = append(sortedQuery, fmt.Sprintf("%s=%s", key, value))
			}
		}
		sort.Strings(sortedQuery)
		keyParts = append(keyParts, strings.Join(sortedQuery, "&"))
	}

	// 根据配置添加请求头
	if len(m.config.CacheHeaders) > 0 {
		for _, headerKey := range m.config.CacheHeaders {
			if value := c.GetHeader(headerKey); value != "" {
				keyParts = append(keyParts, fmt.Sprintf("%s:%s", headerKey, value))
			}
		}
	}

	// 生成键
	key := strings.Join(keyParts, "|")

	// 使用SHA256生成固定长度的键
	hash := sha256.Sum256([]byte(key))
	hashKey := hex.EncodeToString(hash[:])

	// 添加前缀
	return m.config.KeyPrefix + hashKey
}

// getFromCache 从缓存获取响应
func (m *CacheMiddleware) getFromCache(ctx context.Context, key string) (*CachedResponse, error) {
	data, err := m.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var response CachedResponse
	if err := json.Unmarshal([]byte(data), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached response: %w", err)
	}

	return &response, nil
}

// saveToCache 保存响应到缓存
func (m *CacheMiddleware) saveToCache(ctx context.Context, key string, response *CachedResponse, ttl time.Duration) error {
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	return m.redis.Set(ctx, key, data, ttl)
}

// invalidateCache 失效缓存
func (m *CacheMiddleware) invalidateCache(c *gin.Context) {
	method := c.Request.Method
	path := c.Request.URL.Path

	// 获取需要失效的路径
	paths := m.config.GetInvalidatePaths(method, path)
	if len(paths) == 0 {
		return
	}

	ctx := c.Request.Context()

	// 失效相关缓存
	for _, path := range paths {
		// 构建匹配模式
		cachePattern := m.config.KeyPrefix + path + "*"

		// 查找所有匹配的键
		matchingKeys, err := m.redis.Keys(ctx, cachePattern)
		if err != nil {
			m.logger.Error("Failed to find cache keys",
				zap.Error(err),
				zap.String("pattern", cachePattern),
			)
			continue
		}

		if len(matchingKeys) > 0 {
			// 删除匹配的键
			if err := m.redis.Del(ctx, matchingKeys...); err != nil {
				m.logger.Error("Failed to invalidate cache",
					zap.Error(err),
					zap.Strings("keys", matchingKeys),
				)
			} else {
				m.logger.Info("Cache invalidated",
					zap.String("method", method),
					zap.String("path", path),
					zap.Int("count", len(matchingKeys)),
				)
			}
		}
	}
}

// InvalidateByPattern 按模式失效缓存
func (m *CacheMiddleware) InvalidateByPattern(ctx context.Context, pattern string) error {
	cachePattern := m.config.KeyPrefix + pattern

	keys, err := m.redis.Keys(ctx, cachePattern)
	if err != nil {
		return fmt.Errorf("failed to find cache keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	return m.redis.Del(ctx, keys...)
}

// InvalidateByTags 按标签失效缓存
func (m *CacheMiddleware) InvalidateByTags(ctx context.Context, tags ...string) error {
	for _, tag := range tags {
		tagKey := m.config.KeyPrefix + "tag:" + tag

		// 获取该标签下的所有键
		keys, err := m.redis.LRange(ctx, tagKey, 0, -1)
		if err != nil {
			m.logger.Error("Failed to get tagged keys",
				zap.Error(err),
				zap.String("tag", tag),
			)
			continue
		}

		if len(keys) > 0 {
			// 删除缓存键
			if err := m.redis.Del(ctx, keys...); err != nil {
				m.logger.Error("Failed to delete tagged cache",
					zap.Error(err),
					zap.String("tag", tag),
				)
			}

			// 删除标签键
			if err := m.redis.Del(ctx, tagKey); err != nil {
				m.logger.Error("Failed to delete tag key",
					zap.Error(err),
					zap.String("tag", tag),
				)
			}
		}
	}

	return nil
}

// CacheStatsMiddleware 缓存统计中间件
func (m *CacheMiddleware) CacheStatsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 在响应头中添加缓存统计信息
		stats := m.redis.GetStats()
		c.Header("X-Cache-Hits", fmt.Sprintf("%d", stats.Hits))
		c.Header("X-Cache-Misses", fmt.Sprintf("%d", stats.Misses))
		c.Header("X-Cache-Hit-Rate", fmt.Sprintf("%.2f", m.redis.GetHitRate()))
	}
}

// WarmupCache 预热缓存
func (m *CacheMiddleware) WarmupCache(ctx context.Context, urls []config.WarmupURL) error {
	if !m.config.Warmup.Enabled {
		return nil
	}

	m.logger.Info("Starting cache warmup", zap.Int("urls", len(urls)))

	for _, url := range urls {
		// 构建完整URL
		fullURL := url.Path
		if url.Query != "" {
			fullURL += "?" + url.Query
		}

		// 创建请求
		req, err := http.NewRequestWithContext(ctx, url.Method, fullURL, nil)
		if err != nil {
			m.logger.Error("Failed to create warmup request",
				zap.Error(err),
				zap.String("url", fullURL),
			)
			continue
		}

		// 添加请求头
		for key, value := range url.Headers {
			req.Header.Set(key, value)
		}

		// 发送请求
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			m.logger.Error("Failed to warmup cache",
				zap.Error(err),
				zap.String("url", fullURL),
			)
			continue
		}

		// 读取并丢弃响应体
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		m.logger.Info("Cache warmed up",
			zap.String("url", fullURL),
			zap.Int("status", resp.StatusCode),
		)
	}

	m.logger.Info("Cache warmup completed")
	return nil
}

// GetCacheStats 获取缓存统计信息
func (m *CacheMiddleware) GetCacheStats() CacheStats {
	return m.redis.GetStats()
}

// ResetCacheStats 重置缓存统计信息
func (m *CacheMiddleware) ResetCacheStats() {
	m.redis.ResetStats()
}
