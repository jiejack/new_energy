package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// Loader 配置加载器
type Loader struct {
	mu sync.RWMutex

	env           Environment
	configCenter  string
	localConfig   string
	configName    string
	fallback      bool
	watchEnabled  bool

	viper         *viper.Viper
	watcher       *Watcher
	loaded        bool

	// 配置变更回调
	onChangeCallbacks map[string][]func(key string, value interface{})
}

// LoaderOption 加载器选项
type LoaderOption func(*Loader)

// NewLoader 创建配置加载器
func NewLoader(opts ...LoaderOption) *Loader {
	l := &Loader{
		env:               EnvDev,
		configName:        "config",
		fallback:          true,
		watchEnabled:      false,
		onChangeCallbacks: make(map[string][]func(key string, value interface{})),
		viper:             viper.New(),
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// WithEnv 设置环境
func WithEnv(env Environment) LoaderOption {
	return func(l *Loader) {
		l.env = env
	}
}

// WithConfigCenter 设置配置中心地址
func WithConfigCenter(addr string) LoaderOption {
	return func(l *Loader) {
		l.configCenter = addr
	}
}

// WithLocalConfig 设置本地配置目录
func WithLocalConfig(path string) LoaderOption {
	return func(l *Loader) {
		l.localConfig = path
	}
}

// WithConfigName 设置配置文件名称
func WithConfigName(name string) LoaderOption {
	return func(l *Loader) {
		l.configName = name
	}
}

// WithFallback 设置是否启用本地配置兜底
func WithFallback(fallback bool) LoaderOption {
	return func(l *Loader) {
		l.fallback = fallback
	}
}

// WithWatch 设置是否启用配置监听
func WithWatch(enabled bool) LoaderOption {
	return func(l *Loader) {
		l.watchEnabled = enabled
	}
}

// Load 加载配置
// 加载优先级：命令行参数 > 环境变量 > 配置中心 > 本地配置文件 > 默认值
func (l *Loader) Load(ctx context.Context) (*Config, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 1. 设置默认值
	l.setDefaults()

	// 2. 加载本地配置文件
	if err := l.loadLocalConfig(); err != nil && !l.fallback {
		return nil, fmt.Errorf("failed to load local config: %w", err)
	}

	// 3. 加载环境变量
	l.loadEnvVars()

	// 4. 加载命令行参数
	l.loadCommandLine()

	// 5. 尝试从配置中心加载
	if l.configCenter != "" {
		if err := l.loadFromConfigCenter(ctx); err != nil {
			if !l.fallback {
				return nil, fmt.Errorf("failed to load from config center: %w", err)
			}
			// 配置中心加载失败，使用本地配置
		}
	}

	// 6. 解析配置
	var cfg Config
	if err := l.viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 7. 启动配置监听
	if l.watchEnabled {
		if err := l.startWatcher(); err != nil {
			return nil, fmt.Errorf("failed to start watcher: %w", err)
		}
	}

	l.loaded = true
	return &cfg, nil
}

// setDefaults 设置默认值
func (l *Loader) setDefaults() {
	defaultCfg := DefaultConfig()

	// 服务器配置
	l.viper.SetDefault("server.name", defaultCfg.Server.Name)
	l.viper.SetDefault("server.port", defaultCfg.Server.Port)
	l.viper.SetDefault("server.mode", defaultCfg.Server.Mode)

	// 数据库配置
	l.viper.SetDefault("database.type", defaultCfg.Database.Type)
	l.viper.SetDefault("database.host", defaultCfg.Database.Host)
	l.viper.SetDefault("database.port", defaultCfg.Database.Port)
	l.viper.SetDefault("database.user", defaultCfg.Database.User)
	l.viper.SetDefault("database.password", defaultCfg.Database.Password)
	l.viper.SetDefault("database.dbname", defaultCfg.Database.DBName)
	l.viper.SetDefault("database.sslmode", defaultCfg.Database.SSLMode)
	l.viper.SetDefault("database.max_open_conns", defaultCfg.Database.MaxOpenConns)
	l.viper.SetDefault("database.max_idle_conns", defaultCfg.Database.MaxIdleConns)
	l.viper.SetDefault("database.conn_max_lifetime", defaultCfg.Database.ConnMaxLifetime)

	// Redis配置
	l.viper.SetDefault("redis.addrs", defaultCfg.Redis.Addrs)
	l.viper.SetDefault("redis.password", defaultCfg.Redis.Password)
	l.viper.SetDefault("redis.db", defaultCfg.Redis.DB)
	l.viper.SetDefault("redis.pool_size", defaultCfg.Redis.PoolSize)

	// Kafka配置
	l.viper.SetDefault("kafka.brokers", defaultCfg.Kafka.Brokers)
	l.viper.SetDefault("kafka.topic_prefix", defaultCfg.Kafka.TopicPrefix)

	// 时序数据库配置
	l.viper.SetDefault("timeseries.type", defaultCfg.TimeSeries.Type)
	l.viper.SetDefault("timeseries.host", defaultCfg.TimeSeries.Host)
	l.viper.SetDefault("timeseries.port", defaultCfg.TimeSeries.Port)
	l.viper.SetDefault("timeseries.user", defaultCfg.TimeSeries.User)
	l.viper.SetDefault("timeseries.password", defaultCfg.TimeSeries.Password)
	l.viper.SetDefault("timeseries.database", defaultCfg.TimeSeries.Database)

	// Nacos配置
	l.viper.SetDefault("nacos.enabled", defaultCfg.Nacos.Enabled)
	l.viper.SetDefault("nacos.namespace", defaultCfg.Nacos.Namespace)
	l.viper.SetDefault("nacos.group", defaultCfg.Nacos.Group)
	l.viper.SetDefault("nacos.service_name", defaultCfg.Nacos.ServiceName)
	l.viper.SetDefault("nacos.weight", defaultCfg.Nacos.Weight)
	l.viper.SetDefault("nacos.metadata", defaultCfg.Nacos.Metadata)

	// 配置中心配置
	l.viper.SetDefault("config_center.enabled", defaultCfg.ConfigCenter.Enabled)
	l.viper.SetDefault("config_center.provider", defaultCfg.ConfigCenter.Provider)
	l.viper.SetDefault("config_center.namespace", defaultCfg.ConfigCenter.Namespace)
	l.viper.SetDefault("config_center.group", defaultCfg.ConfigCenter.Group)
	l.viper.SetDefault("config_center.refresh_interval", defaultCfg.ConfigCenter.RefreshInterval)

	// 日志配置
	l.viper.SetDefault("logging.level", defaultCfg.Logging.Level)
	l.viper.SetDefault("logging.format", defaultCfg.Logging.Format)
	l.viper.SetDefault("logging.output", defaultCfg.Logging.Output)

	// 链路追踪配置
	l.viper.SetDefault("tracing.enabled", defaultCfg.Tracing.Enabled)
	l.viper.SetDefault("tracing.endpoint", defaultCfg.Tracing.Endpoint)
	l.viper.SetDefault("tracing.sampler_ratio", defaultCfg.Tracing.SamplerRatio)

	// 指标配置
	l.viper.SetDefault("metrics.enabled", defaultCfg.Metrics.Enabled)
	l.viper.SetDefault("metrics.port", defaultCfg.Metrics.Port)

	// 认证配置
	l.viper.SetDefault("auth.jwt.secret", defaultCfg.Auth.JWT.Secret)
	l.viper.SetDefault("auth.jwt.access_expire", defaultCfg.Auth.JWT.AccessExpire)
	l.viper.SetDefault("auth.jwt.refresh_expire", defaultCfg.Auth.JWT.RefreshExpire)
	l.viper.SetDefault("auth.password.min_length", defaultCfg.Auth.Password.MinLength)
	l.viper.SetDefault("auth.password.require_uppercase", defaultCfg.Auth.Password.RequireUppercase)
	l.viper.SetDefault("auth.password.require_lowercase", defaultCfg.Auth.Password.RequireLowercase)
	l.viper.SetDefault("auth.password.require_digit", defaultCfg.Auth.Password.RequireDigit)
	l.viper.SetDefault("auth.login.max_attempts", defaultCfg.Auth.Login.MaxAttempts)
	l.viper.SetDefault("auth.login.lock_duration", defaultCfg.Auth.Login.LockDuration)
}

// loadLocalConfig 加载本地配置文件
func (l *Loader) loadLocalConfig() error {
	configPath := l.localConfig
	if configPath == "" {
		configPath = "./configs"
	}

	// 检查配置目录是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config directory not found: %s", configPath)
	}

	// 设置配置文件搜索路径
	l.viper.AddConfigPath(configPath)
	l.viper.SetConfigName(l.configName)
	l.viper.SetConfigType("yaml")

	// 1. 先加载基础配置文件
	if err := l.viper.MergeInConfig(); err != nil {
		// 基础配置文件可能不存在，忽略错误
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read base config: %w", err)
		}
	}

	// 2. 加载环境特定配置文件
	envConfigFile := filepath.Join(configPath, fmt.Sprintf("%s-%s.yaml", l.configName, l.env))
	if _, err := os.Stat(envConfigFile); err == nil {
		l.viper.SetConfigFile(envConfigFile)
		if err := l.viper.MergeInConfig(); err != nil {
			return fmt.Errorf("failed to merge env config: %w", err)
		}
	}

	return nil
}

// loadEnvVars 加载环境变量
func (l *Loader) loadEnvVars() {
	// 设置环境变量前缀
	l.viper.SetEnvPrefix("NEM")
	
	// 自动绑定环境变量
	l.viper.AutomaticEnv()
	
	// 设置环境变量键替换规则（支持嵌套配置）
	l.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 从环境变量获取当前环境
	if env := os.Getenv("NEM_ENV"); env != "" {
		l.env = Environment(env)
	}
}

// loadCommandLine 加载命令行参数
func (l *Loader) loadCommandLine() {
	// 解析命令行参数
	// 例如：--config=/path/to/config.yaml
	if configPath := l.viper.GetString("config"); configPath != "" {
		l.viper.SetConfigFile(configPath)
		if err := l.viper.MergeInConfig(); err != nil {
			// 命令行指定的配置文件加载失败，记录警告
			fmt.Printf("Warning: failed to load config from command line: %v\n", err)
		}
	}

	// 从环境变量获取当前环境
	if env := l.viper.GetString("env"); env != "" {
		l.env = Environment(env)
	}
}

// loadFromConfigCenter 从配置中心加载配置
func (l *Loader) loadFromConfigCenter(ctx context.Context) error {
	// TODO: 实现从Nacos/Apollo/Consul等配置中心加载配置
	// 这里需要根据provider类型选择不同的实现
	provider := l.viper.GetString("config_center.provider")
	
	switch ConfigProvider(provider) {
	case ProviderNacos:
		return l.loadFromNacos(ctx)
	case ProviderApollo:
		return l.loadFromApollo(ctx)
	case ProviderConsul:
		return l.loadFromConsul(ctx)
	case ProviderEtcd:
		return l.loadFromEtcd(ctx)
	default:
		return fmt.Errorf("unsupported config provider: %s", provider)
	}
}

// loadFromNacos 从Nacos加载配置
func (l *Loader) loadFromNacos(ctx context.Context) error {
	// TODO: 实现Nacos配置加载
	// 需要引入Nacos SDK
	return fmt.Errorf("nacos config center not implemented yet")
}

// loadFromApollo 从Apollo加载配置
func (l *Loader) loadFromApollo(ctx context.Context) error {
	// TODO: 实现Apollo配置加载
	return fmt.Errorf("apollo config center not implemented yet")
}

// loadFromConsul 从Consul加载配置
func (l *Loader) loadFromConsul(ctx context.Context) error {
	// TODO: 实现Consul配置加载
	return fmt.Errorf("consul config center not implemented yet")
}

// loadFromEtcd 从Etcd加载配置
func (l *Loader) loadFromEtcd(ctx context.Context) error {
	// TODO: 实现Etcd配置加载
	return fmt.Errorf("etcd config center not implemented yet")
}

// startWatcher 启动配置监听器
func (l *Loader) startWatcher() error {
	watcher := NewWatcher(l.viper, l.onChangeCallbacks)
	l.watcher = watcher
	return watcher.Start()
}

// Watch 监听配置变更
func (l *Loader) Watch(key string, callback func(key string, value interface{})) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.onChangeCallbacks[key] = append(l.onChangeCallbacks[key], callback)

	// 如果已经启动监听，动态添加回调
	if l.watcher != nil {
		l.watcher.AddCallback(key, callback)
	}
}

// Unwatch 取消配置监听
func (l *Loader) Unwatch(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.onChangeCallbacks, key)

	// 如果已经启动监听，移除回调
	if l.watcher != nil {
		l.watcher.RemoveCallback(key)
	}
}

// Get 获取配置值
func (l *Loader) Get(key string) interface{} {
	return l.viper.Get(key)
}

// GetString 获取字符串配置值
func (l *Loader) GetString(key string) string {
	return l.viper.GetString(key)
}

// GetInt 获取整数配置值
func (l *Loader) GetInt(key string) int {
	return l.viper.GetInt(key)
}

// GetBool 获取布尔配置值
func (l *Loader) GetBool(key string) bool {
	return l.viper.GetBool(key)
}

// GetFloat64 获取浮点数配置值
func (l *Loader) GetFloat64(key string) float64 {
	return l.viper.GetFloat64(key)
}

// GetStringSlice 获取字符串数组配置值
func (l *Loader) GetStringSlice(key string) []string {
	return l.viper.GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置值
func (l *Loader) GetStringMap(key string) map[string]interface{} {
	return l.viper.GetStringMap(key)
}

// GetStringMapString 获取字符串映射字符串配置值
func (l *Loader) GetStringMapString(key string) map[string]string {
	return l.viper.GetStringMapString(key)
}

// Set 设置配置值
func (l *Loader) Set(key string, value interface{}) {
	l.viper.Set(key, value)
}

// IsSet 判断配置键是否存在
func (l *Loader) IsSet(key string) bool {
	return l.viper.IsSet(key)
}

// AllSettings 获取所有配置
func (l *Loader) AllSettings() map[string]interface{} {
	return l.viper.AllSettings()
}

// GetEnv 获取当前环境
func (l *Loader) GetEnv() Environment {
	return l.env
}

// Close 关闭配置加载器
func (l *Loader) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.watcher != nil {
		if err := l.watcher.Stop(); err != nil {
			return err
		}
	}

	return nil
}

// Reload 重新加载配置
func (l *Loader) Reload(ctx context.Context) (*Config, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 停止监听器
	if l.watcher != nil {
		_ = l.watcher.Stop()
	}

	// 重置viper
	l.viper = viper.New()
	l.loaded = false

	// 重新加载
	return l.Load(ctx)
}
