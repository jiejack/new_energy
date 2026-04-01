package nacos

import (
	"time"
)

// Options Nacos客户端配置选项
type Options struct {
	// ServerConfigs Nacos服务器地址列表
	ServerConfigs []ServerConfig

	// ClientConfig 客户端配置
	ClientConfig ClientConfig

	// Namespace 命名空间ID
	Namespace string

	// Group 分组名称
	Group string

	// ServiceName 服务名称
	ServiceName string

	// ServicePort 服务端口
	ServicePort uint64

	// ServiceWeight 服务权重
	ServiceWeight float64

	// ServiceMetadata 服务元数据
	ServiceMetadata map[string]string

	// ClusterName 集群名称
	ClusterName string

	// EnableHeartbeat 是否启用心跳
	EnableHeartbeat bool

	// HeartbeatInterval 心跳间隔
	HeartbeatInterval time.Duration

	// ConfigListenInterval 配置监听间隔
	ConfigListenInterval time.Duration
}

// ServerConfig Nacos服务器配置
type ServerConfig struct {
	// IpAddr 服务器IP地址
	IpAddr string

	// Port 服务器端口
	Port uint64

	// ContextPath 上下文路径
	ContextPath string

	// Scheme 协议类型 http/https
	Scheme string
}

// ClientConfig Nacos客户端配置
type ClientConfig struct {
	// NamespaceId 命名空间ID
	NamespaceId string

	// TimeoutMs 请求超时时间(毫秒)
	TimeoutMs uint64

	// NotLoadCacheAtStart 启动时不加载缓存
	NotLoadCacheAtStart bool

	// UpdateCacheWhenEmpty 当服务列表为空时更新缓存
	UpdateCacheWhenEmpty bool

	// Username 用户名
	Username string

	// Password 密码
	Password string

	// LogLevel 日志级别
	LogLevel string

	// LogDir 日志目录
	LogDir string

	// CacheDir 缓存目录
	CacheDir string

	// RotateTime 日志轮转时间
	RotateTime string

	// MaxAge 日志最大保留天数
	MaxAge int64
}

// ServiceInstance 服务实例配置
type ServiceInstance struct {
	// ServiceName 服务名称
	ServiceName string

	// Ip 实例IP
	Ip string

	// Port 实例端口
	Port uint64

	// Weight 权重
	Weight float64

	// Enable 是否启用
	Enable bool

	// Healthy 是否健康
	Healthy bool

	// Metadata 元数据
	Metadata map[string]string

	// ClusterName 集群名称
	ClusterName string

	// GroupName 分组名称
	GroupName string

	// Ephemeral 是否临时实例
	Ephemeral bool
}

// ConfigOptions 配置中心选项
type ConfigOptions struct {
	// DataId 配置ID
	DataId string

	// Group 分组
	Group string

	// Namespace 命名空间
	Namespace string

	// Content 配置内容
	Content string

	// Tag 配置标签
	Tag string

	// AppName 应用名称
	AppName string

	// EncryptedDataKey 加密数据key
	EncryptedDataKey string
}

// RegistryOptions 注册中心选项
type RegistryOptions struct {
	// ServiceName 服务名称
	ServiceName string

	// Group 分组
	Group string

	// Namespace 命名空间
	Namespace string

	// Clusters 集群列表
	Clusters []string

	// HealthyOnly 是否只返回健康实例
	HealthyOnly bool

	// Enable 是否启用
	Enable bool
}

// DefaultOptions 返回默认配置选项
func DefaultOptions() *Options {
	return &Options{
		ServerConfigs: []ServerConfig{
			{
				IpAddr:      "127.0.0.1",
				Port:        8848,
				ContextPath: "/nacos",
				Scheme:      "http",
			},
		},
		ClientConfig: ClientConfig{
			TimeoutMs:           5000,
			NotLoadCacheAtStart: true,
			UpdateCacheWhenEmpty: true,
			LogLevel:            "info",
		},
		Namespace:            "public",
		Group:               "DEFAULT_GROUP",
		ServiceWeight:       1.0,
		ClusterName:         "DEFAULT",
		EnableHeartbeat:     true,
		HeartbeatInterval:   5 * time.Second,
		ConfigListenInterval: 3 * time.Second,
		ServiceMetadata:     make(map[string]string),
	}
}

// Option 配置选项函数
type Option func(*Options)

// WithServerConfigs 设置服务器配置
func WithServerConfigs(servers []ServerConfig) Option {
	return func(o *Options) {
		o.ServerConfigs = servers
	}
}

// WithClientConfig 设置客户端配置
func WithClientConfig(config ClientConfig) Option {
	return func(o *Options) {
		o.ClientConfig = config
	}
}

// WithNamespace 设置命名空间
func WithNamespace(namespace string) Option {
	return func(o *Options) {
		o.Namespace = namespace
		o.ClientConfig.NamespaceId = namespace
	}
}

// WithGroup 设置分组
func WithGroup(group string) Option {
	return func(o *Options) {
		o.Group = group
	}
}

// WithServiceName 设置服务名称
func WithServiceName(name string) Option {
	return func(o *Options) {
		o.ServiceName = name
	}
}

// WithServicePort 设置服务端口
func WithServicePort(port uint64) Option {
	return func(o *Options) {
		o.ServicePort = port
	}
}

// WithServiceWeight 设置服务权重
func WithServiceWeight(weight float64) Option {
	return func(o *Options) {
		o.ServiceWeight = weight
	}
}

// WithServiceMetadata 设置服务元数据
func WithServiceMetadata(metadata map[string]string) Option {
	return func(o *Options) {
		o.ServiceMetadata = metadata
	}
}

// WithClusterName 设置集群名称
func WithClusterName(cluster string) Option {
	return func(o *Options) {
		o.ClusterName = cluster
	}
}

// WithHeartbeat 设置心跳配置
func WithHeartbeat(enable bool, interval time.Duration) Option {
	return func(o *Options) {
		o.EnableHeartbeat = enable
		o.HeartbeatInterval = interval
	}
}

// WithConfigListenInterval 设置配置监听间隔
func WithConfigListenInterval(interval time.Duration) Option {
	return func(o *Options) {
		o.ConfigListenInterval = interval
	}
}

// WithUsername 设置用户名
func WithUsername(username string) Option {
	return func(o *Options) {
		o.ClientConfig.Username = username
	}
}

// WithPassword 设置密码
func WithPassword(password string) Option {
	return func(o *Options) {
		o.ClientConfig.Password = password
	}
}

// ApplyOptions 应用配置选项
func ApplyOptions(opts ...Option) *Options {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	return options
}
