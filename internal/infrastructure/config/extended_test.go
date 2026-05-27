package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_ValidConfig(t *testing.T) {
	content := `
server:
  name: test-server
  port: 8080
  mode: debug
database:
  type: postgres
  host: localhost
  port: 5432
  user: admin
  password: secret
  dbname: testdb
  sslmode: disable
redis:
  addrs:
    - localhost:6379
  password: ""
  db: 0
  pool_size: 100
kafka:
  brokers:
    - localhost:9092
  topic_prefix: test
logging:
  level: info
  format: json
  output: stdout
tracing:
  enabled: true
  endpoint: localhost:4317
  sampler_ratio: 0.5
metrics:
  enabled: true
  port: 9090
auth:
  jwt:
    secret: test-secret
    access_expire: 3600
    refresh_expire: 86400
  password:
    min_length: 8
    require_uppercase: true
    require_lowercase: true
    require_digit: true
  login:
    max_attempts: 5
    lock_duration: 30
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test-server", cfg.Server.Name)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Server.Mode)
	assert.Equal(t, "postgres", cfg.Database.Type)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "admin", cfg.Database.User)
	assert.Equal(t, "secret", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.DBName)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, []string{"localhost:6379"}, cfg.Redis.Addrs)
	assert.Equal(t, []string{"localhost:9092"}, cfg.Kafka.Brokers)
	assert.Equal(t, "test", cfg.Kafka.TopicPrefix)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.True(t, cfg.Tracing.Enabled)
	assert.True(t, cfg.Metrics.Enabled)
	assert.Equal(t, "test-secret", cfg.Auth.JWT.Secret)
	assert.Equal(t, int64(3600), cfg.Auth.JWT.AccessExpire)
	assert.Equal(t, int64(86400), cfg.Auth.JWT.RefreshExpire)
	assert.Equal(t, 8, cfg.Auth.Password.MinLength)
	assert.True(t, cfg.Auth.Password.RequireUppercase)
	assert.Equal(t, 5, cfg.Auth.Login.MaxAttempts)
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config")
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte("invalid: [yaml: content"), 0644)
	require.NoError(t, err)

	_, err = Load(configPath)
	assert.Error(t, err)
}

func TestLoadFromEnv(t *testing.T) {
	cfg, err := LoadFromEnv()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestMatchPath_Exact(t *testing.T) {
	assert.True(t, matchPath("/api/v1/devices", "/api/v1/devices"))
	assert.False(t, matchPath("/api/v1/devices", "/api/v1/stations"))
}

func TestMatchPath_Wildcard(t *testing.T) {
	assert.True(t, matchPath("/api/v1/*", "/api/v1/devices"))
	assert.True(t, matchPath("/api/*", "/api/v1/stations"))
	assert.False(t, matchPath("/api/v1/*", "/api/v2/devices"))
}

func TestMatchPath_Empty(t *testing.T) {
	assert.False(t, matchPath("", "/api/v1/devices"))
	assert.True(t, matchPath("/api/v1/devices", "/api/v1/devices"))
}

func TestAPICacheConfig_GetTTLForPath_DisabledPattern(t *testing.T) {
	cfg := DefaultCacheConfig()
	cfg.API.PathPatterns = []CachePathPattern{
		{Pattern: "/api/v1/test", TTL: 10 * time.Minute, Enabled: false},
	}
	ttl := cfg.API.GetTTLForPath("/api/v1/test")
	assert.Equal(t, time.Duration(0), ttl)
}

func TestAPICacheConfig_ShouldCachePath_DisabledPattern(t *testing.T) {
	cfg := DefaultCacheConfig()
	cfg.API.PathPatterns = []CachePathPattern{
		{Pattern: "/api/v1/test", Enabled: false},
	}
	assert.False(t, cfg.API.ShouldCachePath("/api/v1/test"))
}

func TestAPICacheConfig_GetInvalidatePaths_MultipleMethods(t *testing.T) {
	cfg := DefaultCacheConfig()
	paths := cfg.API.GetInvalidatePaths("PUT", "/api/v1/devices")
	assert.Equal(t, 1, len(paths))
	paths = cfg.API.GetInvalidatePaths("PATCH", "/api/v1/devices")
	assert.Equal(t, 1, len(paths))
}
