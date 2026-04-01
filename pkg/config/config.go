package config

import (
	"fmt"
	"time"
)

// Environment 环境类型
type Environment string

const (
	EnvDev        Environment = "dev"        // 开发环境
	EnvTest       Environment = "test"       // 测试环境
	EnvProd       Environment = "prod"       // 生产环境
	EnvStandalone Environment = "standalone" // 单机模式
)

// ConfigProvider 配置中心提供者
type ConfigProvider string

const (
	ProviderNacos ConfigProvider = "nacos" // Nacos配置中心
	ProviderApollo ConfigProvider = "apollo" // Apollo配置中心
	ProviderConsul ConfigProvider = "consul" // Consul配置中心
	ProviderEtcd   ConfigProvider = "etcd"   // Etcd配置中心
)

// ValueType 配置值类型
type ValueType string

const (
	ValueTypeString  ValueType = "string"  // 字符串类型
	ValueTypeInt     ValueType = "int"     // 整数类型
	ValueTypeFloat   ValueType = "float"   // 浮点数类型
	ValueTypeBool    ValueType = "bool"    // 布尔类型
	ValueTypeJSON    ValueType = "json"    // JSON类型
	ValueTypeYAML    ValueType = "yaml"    // YAML类型
	ValueTypeList    ValueType = "list"    // 列表类型
)

// ReleaseType 发布类型
type ReleaseType string

const (
	ReleaseTypeFull    ReleaseType = "full"    // 全量发布
	ReleaseTypeGray    ReleaseType = "gray"    // 灰度发布
	ReleaseTypeRollback ReleaseType = "rollback" // 回滚发布
)

// ReleaseStatus 发布状态
type ReleaseStatus string

const (
	ReleaseStatusPending   ReleaseStatus = "pending"   // 待发布
	ReleaseStatusReleasing ReleaseStatus = "releasing" // 发布中
	ReleaseStatusSuccess   ReleaseStatus = "success"   // 发布成功
	ReleaseStatusFailed    ReleaseStatus = "failed"    // 发布失败
)

// AuditAction 审计操作类型
type AuditAction string

const (
	AuditActionCreate AuditAction = "create" // 创建
	AuditActionUpdate AuditAction = "update" // 更新
	AuditActionDelete AuditAction = "delete" // 删除
	AuditActionRollback AuditAction = "rollback" // 回滚
	AuditActionRelease AuditAction = "release" // 发布
)

// Config 应用配置主结构
type Config struct {
	Server       ServerConfig       `mapstructure:"server" json:"server" yaml:"server"`
	Database     DatabaseConfig     `mapstructure:"database" json:"database" yaml:"database"`
	Redis        RedisConfig        `mapstructure:"redis" json:"redis" yaml:"redis"`
	Kafka        KafkaConfig        `mapstructure:"kafka" json:"kafka" yaml:"kafka"`
	TimeSeries   TimeSeriesConfig   `mapstructure:"timeseries" json:"timeseries" yaml:"timeseries"`
	Nacos        NacosConfig        `mapstructure:"nacos" json:"nacos" yaml:"nacos"`
	ConfigCenter ConfigCenterConfig `mapstructure:"config_center" json:"config_center" yaml:"config_center"`
	Logging      LoggingConfig      `mapstructure:"logging" json:"logging" yaml:"logging"`
	Tracing      TracingConfig      `mapstructure:"tracing" json:"tracing" yaml:"tracing"`
	Metrics      MetricsConfig      `mapstructure:"metrics" json:"metrics" yaml:"metrics"`
	Auth         AuthConfig         `mapstructure:"auth" json:"auth" yaml:"auth"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name string `mapstructure:"name" json:"name" yaml:"name"` // 服务名称
	Port int    `mapstructure:"port" json:"port" yaml:"port"` // 服务端口
	Mode string `mapstructure:"mode" json:"mode" yaml:"mode"` // 运行模式: debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type            string        `mapstructure:"type" json:"type" yaml:"type"`                         // 数据库类型
	Host            string        `mapstructure:"host" json:"host" yaml:"host"`                         // 主机地址
	Port            int           `mapstructure:"port" json:"port" yaml:"port"`                         // 端口
	User            string        `mapstructure:"user" json:"user" yaml:"user"`                         // 用户名
	Password        string        `mapstructure:"password" json:"password" yaml:"password"`             // 密码
	DBName          string        `mapstructure:"dbname" json:"dbname" yaml:"dbname"`                   // 数据库名
	SSLMode         string        `mapstructure:"sslmode" json:"sslmode" yaml:"sslmode"`               // SSL模式
	MaxOpenConns    int           `mapstructure:"max_open_conns" json:"max_open_conns" yaml:"max_open_conns"`       // 最大打开连接数
	MaxIdleConns    int           `mapstructure:"max_idle_conns" json:"max_idle_conns" yaml:"max_idle_conns"`       // 最大空闲连接数
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" json:"conn_max_lifetime" yaml:"conn_max_lifetime"` // 连接最大生命周期
}

// DSN 返回数据库连接字符串
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addrs    []string `mapstructure:"addrs" json:"addrs" yaml:"addrs"`       // 集群地址列表
	Password string   `mapstructure:"password" json:"password" yaml:"password"` // 密码
	DB       int      `mapstructure:"db" json:"db" yaml:"db"`               // 数据库索引
	PoolSize int      `mapstructure:"pool_size" json:"pool_size" yaml:"pool_size"` // 连接池大小
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers     []string `mapstructure:"brokers" json:"brokers" yaml:"brokers"`           // Broker地址列表
	TopicPrefix string   `mapstructure:"topic_prefix" json:"topic_prefix" yaml:"topic_prefix"` // Topic前缀
}

// TimeSeriesConfig 时序数据库配置
type TimeSeriesConfig struct {
	Type     string `mapstructure:"type" json:"type" yaml:"type"`         // 数据库类型
	Host     string `mapstructure:"host" json:"host" yaml:"host"`         // 主机地址
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`         // 端口
	User     string `mapstructure:"user" json:"user" yaml:"user"`         // 用户名
	Password string `mapstructure:"password" json:"password" yaml:"password"` // 密码
	Database string `mapstructure:"database" json:"database" yaml:"database"` // 数据库名
}

// DSN 返回时序数据库连接字符串
func (c *TimeSeriesConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.User, c.Password, c.Host, c.Port, c.Database)
}

// NacosConfig Nacos配置
type NacosConfig struct {
	Enabled     bool              `mapstructure:"enabled" json:"enabled" yaml:"enabled"`             // 是否启用
	ServerAddr  string            `mapstructure:"server_addr" json:"server_addr" yaml:"server_addr"` // 服务地址
	Namespace   string            `mapstructure:"namespace" json:"namespace" yaml:"namespace"`       // 命名空间
	Group       string            `mapstructure:"group" json:"group" yaml:"group"`                   // 分组
	ServiceName string            `mapstructure:"service_name" json:"service_name" yaml:"service_name"` // 服务名称
	Weight      float64           `mapstructure:"weight" json:"weight" yaml:"weight"`               // 权重
	Metadata    map[string]string `mapstructure:"metadata" json:"metadata" yaml:"metadata"`          // 元数据
}

// ConfigCenterConfig 配置中心配置
type ConfigCenterConfig struct {
	Enabled         bool           `mapstructure:"enabled" json:"enabled" yaml:"enabled"`                     // 是否启用
	Provider        ConfigProvider `mapstructure:"provider" json:"provider" yaml:"provider"`                  // 配置中心提供者
	ServerAddr      string         `mapstructure:"server_addr" json:"server_addr" yaml:"server_addr"`         // 服务地址
	Namespace       string         `mapstructure:"namespace" json:"namespace" yaml:"namespace"`               // 命名空间
	Group           string         `mapstructure:"group" json:"group" yaml:"group"`                           // 分组
	DataID          string         `mapstructure:"data_id" json:"data_id" yaml:"data_id"`                     // 配置ID
	RefreshInterval int            `mapstructure:"refresh_interval" json:"refresh_interval" yaml:"refresh_interval"` // 刷新间隔(秒)
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `mapstructure:"level" json:"level" yaml:"level"`     // 日志级别
	Format string `mapstructure:"format" json:"format" yaml:"format"` // 输出格式: console, json
	Output string `mapstructure:"output" json:"output" yaml:"output"` // 输出位置: stdout, stderr, file
}

// TracingConfig 链路追踪配置
type TracingConfig struct {
	Enabled      bool    `mapstructure:"enabled" json:"enabled" yaml:"enabled"`               // 是否启用
	Endpoint     string  `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`           // 采集端点
	SamplerRatio float64 `mapstructure:"sampler_ratio" json:"sampler_ratio" yaml:"sampler_ratio"` // 采样率
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"` // 是否启用
	Port    int  `mapstructure:"port" json:"port" yaml:"port"`         // 指标端口
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWT     JWTConfig     `mapstructure:"jwt" json:"jwt" yaml:"jwt"`         // JWT配置
	Password PasswordConfig `mapstructure:"password" json:"password" yaml:"password"` // 密码配置
	Login   LoginConfig   `mapstructure:"login" json:"login" yaml:"login"`   // 登录配置
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret        string `mapstructure:"secret" json:"secret" yaml:"secret"`                     // 密钥
	AccessExpire  int    `mapstructure:"access_expire" json:"access_expire" yaml:"access_expire"`   // 访问令牌过期时间(秒)
	RefreshExpire int    `mapstructure:"refresh_expire" json:"refresh_expire" yaml:"refresh_expire"` // 刷新令牌过期时间(秒)
}

// PasswordConfig 密码配置
type PasswordConfig struct {
	MinLength       int  `mapstructure:"min_length" json:"min_length" yaml:"min_length"`                   // 最小长度
	RequireUppercase bool `mapstructure:"require_uppercase" json:"require_uppercase" yaml:"require_uppercase"` // 是否要求大写字母
	RequireLowercase bool `mapstructure:"require_lowercase" json:"require_lowercase" yaml:"require_lowercase"` // 是否要求小写字母
	RequireDigit    bool `mapstructure:"require_digit" json:"require_digit" yaml:"require_digit"`          // 是否要求数字
}

// LoginConfig 登录配置
type LoginConfig struct {
	MaxAttempts  int `mapstructure:"max_attempts" json:"max_attempts" yaml:"max_attempts"`     // 最大尝试次数
	LockDuration int `mapstructure:"lock_duration" json:"lock_duration" yaml:"lock_duration"` // 锁定时长(秒)
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Name: "nem-server",
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Type:            "postgres",
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Password:        "postgres",
			DBName:          "nem_system",
			SSLMode:         "disable",
			MaxOpenConns:    100,
			MaxIdleConns:    10,
			ConnMaxLifetime: 3600 * time.Second,
		},
		Redis: RedisConfig{
			Addrs:    []string{"localhost:6379"},
			Password: "",
			DB:       0,
			PoolSize: 100,
		},
		Kafka: KafkaConfig{
			Brokers:     []string{"localhost:9092"},
			TopicPrefix: "nem",
		},
		TimeSeries: TimeSeriesConfig{
			Type:     "doris",
			Host:     "localhost",
			Port:     9030,
			User:     "root",
			Password: "",
			Database: "nem_ts",
		},
		Nacos: NacosConfig{
			Enabled:     false,
			Namespace:   "public",
			Group:       "DEFAULT_GROUP",
			ServiceName: "nem-server",
			Weight:      1,
			Metadata:    map[string]string{"version": "1.0.0"},
		},
		ConfigCenter: ConfigCenterConfig{
			Enabled:         false,
			Provider:        ProviderNacos,
			Namespace:       "public",
			Group:           "DEFAULT_GROUP",
			RefreshInterval: 30,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "console",
			Output: "stdout",
		},
		Tracing: TracingConfig{
			Enabled:      false,
			Endpoint:     "localhost:4317",
			SamplerRatio: 1.0,
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Port:    9090,
		},
		Auth: AuthConfig{
			JWT: JWTConfig{
				Secret:        "change-me-in-production",
				AccessExpire:  7200,
				RefreshExpire: 604800,
			},
			Password: PasswordConfig{
				MinLength:       8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:    true,
			},
			Login: LoginConfig{
				MaxAttempts:  5,
				LockDuration: 1800,
			},
		},
	}
}
