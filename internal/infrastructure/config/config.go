package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Kafka      KafkaConfig      `mapstructure:"kafka"`
	TimeSeries TimeSeriesConfig `mapstructure:"timeseries"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Tracing    TracingConfig    `mapstructure:"tracing"`
	Metrics    MetricsConfig    `mapstructure:"metrics"`
	Auth       AuthConfig       `mapstructure:"auth"`
}

type ServerConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Type            string        `mapstructure:"type"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

type RedisConfig struct {
	Addrs     []string `mapstructure:"addrs"`
	Password  string   `mapstructure:"password"`
	DB        int      `mapstructure:"db"`
	PoolSize  int      `mapstructure:"pool_size"`
}

type KafkaConfig struct {
	Brokers     []string `mapstructure:"brokers"`
	TopicPrefix string   `mapstructure:"topic_prefix"`
}

type TimeSeriesConfig struct {
	Type       string             `mapstructure:"type"`
	Doris      DorisConfig        `mapstructure:"doris"`
	ClickHouse ClickHouseConfig   `mapstructure:"clickhouse"`
}

// DorisConfig Doris配置
type DorisConfig struct {
	Hosts        []string      `mapstructure:"hosts"`
	Database     string        `mapstructure:"database"`
	User         string        `mapstructure:"user"`
	Password     string        `mapstructure:"password"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	ConnTimeout  time.Duration `mapstructure:"conn_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	QueryTimeout time.Duration `mapstructure:"query_timeout"`
	BatchSize    int           `mapstructure:"batch_size"`
}

// ClickHouseConfig ClickHouse配置
type ClickHouseConfig struct {
	Addr         []string      `mapstructure:"addr"`
	Database     string        `mapstructure:"database"`
	User         string        `mapstructure:"user"`
	Password     string        `mapstructure:"password"`
	Compression  string        `mapstructure:"compression"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	ConnTimeout  time.Duration `mapstructure:"conn_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	QueryTimeout time.Duration `mapstructure:"query_timeout"`
	BatchSize    int           `mapstructure:"batch_size"`
	BlockSize    int           `mapstructure:"block_size"`
	Debug        bool          `mapstructure:"debug"`
}

func (c *TimeSeriesConfig) DSN() string {
	// 兼容旧配置格式
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.Doris.User, c.Doris.Password, c.Doris.Hosts[0], 9030, c.Doris.Database)
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type TracingConfig struct {
	Enabled      bool    `mapstructure:"enabled"`
	Endpoint     string  `mapstructure:"endpoint"`
	SamplerRatio float64 `mapstructure:"sampler_ratio"`
}

type MetricsConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
}

type AuthConfig struct {
	JWT      JWTConfig      `mapstructure:"jwt"`
	Password PasswordConfig `mapstructure:"password"`
	Login    LoginConfig    `mapstructure:"login"`
}

type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	AccessExpire  int64  `mapstructure:"access_expire"`
	RefreshExpire int64  `mapstructure:"refresh_expire"`
}

type PasswordConfig struct {
	MinLength       int  `mapstructure:"min_length"`
	RequireUppercase bool `mapstructure:"require_uppercase"`
	RequireLowercase bool `mapstructure:"require_lowercase"`
	RequireDigit    bool `mapstructure:"require_digit"`
}

type LoginConfig struct {
	MaxAttempts  int `mapstructure:"max_attempts"`
	LockDuration int `mapstructure:"lock_duration"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return &cfg, nil
}

func LoadFromEnv() (*Config, error) {
	viper.AutomaticEnv()
	
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return &cfg, nil
}
