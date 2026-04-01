package config

import "time"

// CacheConfig 缓存配置
type CacheConfig struct {
	// Redis配置
	Redis RedisCacheConfig `mapstructure:"redis"`

	// API缓存配置
	API APICacheConfig `mapstructure:"api"`
}

// RedisCacheConfig Redis缓存配置
type RedisCacheConfig struct {
	// 连接配置
	Addrs           []string      `mapstructure:"addrs"`
	Password        string        `mapstructure:"password"`
	DB              int           `mapstructure:"db"`
	PoolSize        int           `mapstructure:"pool_size"`
	MinIdleConns    int           `mapstructure:"min_idle_conns"`
	MaxRetries      int           `mapstructure:"max_retries"`
	DialTimeout     time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	PoolTimeout     time.Duration `mapstructure:"pool_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	MaxConnAge      time.Duration `mapstructure:"max_conn_age"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`

	// 缓存键前缀
	KeyPrefix string `mapstructure:"key_prefix"`

	// 默认过期时间
	DefaultExpiration time.Duration `mapstructure:"default_expiration"`
}

// APICacheConfig API缓存配置
type APICacheConfig struct {
	// 是否启用缓存
	Enabled bool `mapstructure:"enabled"`

	// 默认缓存时间
	DefaultTTL time.Duration `mapstructure:"default_ttl"`

	// 缓存键前缀
	KeyPrefix string `mapstructure:"key_prefix"`

	// 需要缓存的路径模式
	PathPatterns []CachePathPattern `mapstructure:"path_patterns"`

	// 需要排除缓存的路径
	ExcludePaths []string `mapstructure:"exclude_paths"`

	// 是否缓存查询参数
	CacheQueryParams bool `mapstructure:"cache_query_params"`

	// 是否缓存请求头
	CacheHeaders []string `mapstructure:"cache_headers"`

	// 最大缓存大小（字节）
	MaxCacheSize int64 `mapstructure:"max_cache_size"`

	// 缓存预热配置
	Warmup WarmupConfig `mapstructure:"warmup"`
}

// CachePathPattern 缓存路径模式配置
type CachePathPattern struct {
	// 路径模式（支持通配符）
	Pattern string `mapstructure:"pattern"`

	// 缓存时间（0表示使用默认值）
	TTL time.Duration `mapstructure:"ttl"`

	// 缓存键生成策略
	// - "path": 仅使用路径
	// - "path_query": 使用路径+查询参数
	// - "custom": 自定义键生成函数
	KeyStrategy string `mapstructure:"key_strategy"`

	// 自定义缓存键前缀
	KeyPrefix string `mapstructure:"key_prefix"`

	// 是否启用
	Enabled bool `mapstructure:"enabled"`

	// 失效策略
	InvalidateOn []string `mapstructure:"invalidate_on"` // POST, PUT, DELETE, PATCH
}

// WarmupConfig 缓存预热配置
type WarmupConfig struct {
	// 是否启用预热
	Enabled bool `mapstructure:"enabled"`

	// 预热时间点（cron表达式）
	Schedule string `mapstructure:"schedule"`

	// 需要预热的URL列表
	URLs []WarmupURL `mapstructure:"urls"`
}

// WarmupURL 预热URL配置
type WarmupURL struct {
	// URL路径
	Path string `mapstructure:"path"`

	// 查询参数
	Query string `mapstructure:"query"`

	// 请求方法
	Method string `mapstructure:"method"`

	// 请求头
	Headers map[string]string `mapstructure:"headers"`

	// 预热时间（相对于启动时间的延迟）
	Delay time.Duration `mapstructure:"delay"`
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Redis: RedisCacheConfig{
			PoolSize:         100,
			MinIdleConns:     10,
			MaxRetries:       3,
			DialTimeout:      5 * time.Second,
			ReadTimeout:      3 * time.Second,
			WriteTimeout:     3 * time.Second,
			PoolTimeout:      4 * time.Second,
			IdleTimeout:      5 * time.Minute,
			ConnMaxIdleTime:  5 * time.Minute,
			KeyPrefix:        "nem:cache:",
			DefaultExpiration: 5 * time.Minute,
		},
		API: APICacheConfig{
			Enabled:           true,
			DefaultTTL:        5 * time.Minute,
			KeyPrefix:         "api:",
			CacheQueryParams:  true,
			MaxCacheSize:      10 * 1024 * 1024, // 10MB
			ExcludePaths:      []string{"/health", "/ready", "/metrics"},
			PathPatterns:      getDefaultPathPatterns(),
			CacheHeaders:      []string{"Authorization", "Accept-Language"},
			Warmup: WarmupConfig{
				Enabled: false,
			},
		},
	}
}

// getDefaultPathPatterns 获取默认路径模式
func getDefaultPathPatterns() []CachePathPattern {
	return []CachePathPattern{
		{
			Pattern:      "/api/v1/devices",
			TTL:          5 * time.Minute,
			KeyStrategy:  "path_query",
			Enabled:      true,
			InvalidateOn: []string{"POST", "PUT", "DELETE", "PATCH"},
		},
		{
			Pattern:      "/api/v1/stations",
			TTL:          5 * time.Minute,
			KeyStrategy:  "path_query",
			Enabled:      true,
			InvalidateOn: []string{"POST", "PUT", "DELETE", "PATCH"},
		},
		{
			Pattern:      "/api/v1/regions",
			TTL:          10 * time.Minute,
			KeyStrategy:  "path_query",
			Enabled:      true,
			InvalidateOn: []string{"POST", "PUT", "DELETE", "PATCH"},
		},
		{
			Pattern:      "/api/v1/alarms",
			TTL:          1 * time.Minute,
			KeyStrategy:  "path_query",
			Enabled:      true,
			InvalidateOn: []string{"POST", "PUT", "DELETE", "PATCH"},
		},
		{
			Pattern:      "/api/v1/users",
			TTL:          10 * time.Minute,
			KeyStrategy:  "path_query",
			Enabled:      true,
			InvalidateOn: []string{"POST", "PUT", "DELETE", "PATCH"},
		},
		{
			Pattern:      "/api/v1/points",
			TTL:          5 * time.Minute,
			KeyStrategy:  "path_query",
			Enabled:      true,
			InvalidateOn: []string{"POST", "PUT", "DELETE", "PATCH"},
		},
	}
}

// GetTTLForPath 根据路径获取缓存时间
func (c *APICacheConfig) GetTTLForPath(path string) time.Duration {
	for _, pattern := range c.PathPatterns {
		if matchPath(pattern.Pattern, path) && pattern.Enabled {
			if pattern.TTL > 0 {
				return pattern.TTL
			}
			return c.DefaultTTL
		}
	}
	return 0
}

// ShouldCachePath 判断是否应该缓存该路径
func (c *APICacheConfig) ShouldCachePath(path string) bool {
	// 检查排除路径
	for _, excludePath := range c.ExcludePaths {
		if path == excludePath {
			return false
		}
	}

	// 检查匹配的模式
	for _, pattern := range c.PathPatterns {
		if matchPath(pattern.Pattern, path) && pattern.Enabled {
			return true
		}
	}

	return false
}

// GetInvalidatePaths 获取需要失效缓存的路径
func (c *APICacheConfig) GetInvalidatePaths(method, path string) []string {
	var paths []string

	for _, pattern := range c.PathPatterns {
		if matchPath(pattern.Pattern, path) {
			for _, m := range pattern.InvalidateOn {
				if m == method {
					paths = append(paths, pattern.Pattern)
					break
				}
			}
		}
	}

	return paths
}

// matchPath 匹配路径（支持简单的通配符）
func matchPath(pattern, path string) bool {
	// 简单实现：支持 * 通配符
	if pattern == path {
		return true
	}

	// 支持 /* 匹配所有子路径
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}
