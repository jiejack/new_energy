package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "admin",
		Password: "secret",
		DBName:   "testdb",
		SSLMode:  "disable",
	}
	expected := "host=localhost port=5432 user=admin password=secret dbname=testdb sslmode=disable"
	assert.Equal(t, expected, cfg.DSN())
}

func TestTimeSeriesConfig_DSN(t *testing.T) {
	cfg := TimeSeriesConfig{
		Doris: DorisConfig{
			Hosts:    []string{"doris-host"},
			Database: "testdb",
			User:     "admin",
			Password: "secret",
		},
	}
	dsn := cfg.DSN()
	assert.Contains(t, dsn, "admin")
	assert.Contains(t, dsn, "secret")
	assert.Contains(t, dsn, "doris-host")
	assert.Contains(t, dsn, "testdb")
}

func TestConfig_Struct(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{
			Name: "test-server",
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "admin",
			Password: "secret",
			DBName:   "testdb",
			SSLMode:  "disable",
		},
		Redis: RedisConfig{
			Addrs:    []string{"localhost:6379"},
			Password: "",
			DB:       0,
			PoolSize: 100,
		},
		Kafka: KafkaConfig{
			Brokers:     []string{"localhost:9092"},
			TopicPrefix: "test",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Tracing: TracingConfig{
			Enabled:      true,
			Endpoint:     "localhost:4317",
			SamplerRatio: 0.5,
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Port:    9090,
		},
		Auth: AuthConfig{
			JWT: JWTConfig{
				Secret:        "test-secret",
				AccessExpire:  3600,
				RefreshExpire: 86400,
			},
			Password: PasswordConfig{
				MinLength:       8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:    true,
			},
			Login: LoginConfig{
				MaxAttempts:  5,
				LockDuration: 30,
			},
		},
	}
	assert.Equal(t, "test-server", cfg.Server.Name)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "localhost:6379", cfg.Redis.Addrs[0])
	assert.Equal(t, "localhost:9092", cfg.Kafka.Brokers[0])
	assert.True(t, cfg.Tracing.Enabled)
	assert.True(t, cfg.Metrics.Enabled)
	assert.Equal(t, "test-secret", cfg.Auth.JWT.Secret)
}
