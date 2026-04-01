package nacos

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// Registry Nacos服务注册中心
type Registry struct {
	options     *Options
	namingCli   naming_client.INamingClient
	instances   map[string]*ServiceInstance
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	registered  bool
}

// NewRegistry 创建新的注册中心客户端
func NewRegistry(opts ...Option) (*Registry, error) {
	options := ApplyOptions(opts...)

	// 构建服务器配置
	serverConfigs := make([]constant.ServerConfig, 0, len(options.ServerConfigs))
	for _, sc := range options.ServerConfigs {
		serverConfigs = append(serverConfigs, *constant.NewServerConfig(
			sc.IpAddr,
			sc.Port,
			constant.WithContextPath(sc.ContextPath),
			constant.WithScheme(sc.Scheme),
		))
	}

	// 构建客户端配置
	clientConfig := constant.NewClientConfig(
		constant.WithNamespaceId(options.ClientConfig.NamespaceId),
		constant.WithTimeoutMs(options.ClientConfig.TimeoutMs),
		constant.WithNotLoadCacheAtStart(options.ClientConfig.NotLoadCacheAtStart),
		constant.WithUpdateCacheWhenEmpty(options.ClientConfig.UpdateCacheWhenEmpty),
		constant.WithUsername(options.ClientConfig.Username),
		constant.WithPassword(options.ClientConfig.Password),
		constant.WithLogLevel(options.ClientConfig.LogLevel),
	)

	// 创建命名客户端
	namingCli, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create naming client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Registry{
		options:   options,
		namingCli: namingCli,
		instances: make(map[string]*ServiceInstance),
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Register 注册服务实例
func (r *Registry) Register(instance *ServiceInstance) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}

	// 获取本机IP
	ip := instance.Ip
	if ip == "" {
		localIP, err := getLocalIP()
		if err != nil {
			return fmt.Errorf("failed to get local IP: %w", err)
		}
		ip = localIP
	}

	// 注册服务
	success, err := r.namingCli.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        instance.Port,
		ServiceName: instance.ServiceName,
		Weight:      instance.Weight,
		Enable:      instance.Enable,
		Healthy:     instance.Healthy,
		Metadata:    instance.Metadata,
		ClusterName: instance.ClusterName,
		GroupName:   instance.GroupName,
		Ephemeral:   instance.Ephemeral,
	})

	if err != nil {
		return fmt.Errorf("failed to register instance: %w", err)
	}

	if !success {
		return fmt.Errorf("failed to register instance: operation returned false")
	}

	// 保存实例信息
	r.mu.Lock()
	instance.Ip = ip
	r.instances[instance.ServiceName] = instance
	r.registered = true
	r.mu.Unlock()

	return nil
}

// Deregister 注销服务实例
func (r *Registry) Deregister(serviceName string) error {
	r.mu.RLock()
	instance, exists := r.instances[serviceName]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	success, err := r.namingCli.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          instance.Ip,
		Port:        instance.Port,
		ServiceName: instance.ServiceName,
		ClusterName: instance.ClusterName,
		GroupName:   instance.GroupName,
		Ephemeral:   instance.Ephemeral,
	})

	if err != nil {
		return fmt.Errorf("failed to deregister instance: %w", err)
	}

	if !success {
		return fmt.Errorf("failed to deregister instance: operation returned false")
	}

	r.mu.Lock()
	delete(r.instances, serviceName)
	if len(r.instances) == 0 {
		r.registered = false
	}
	r.mu.Unlock()

	return nil
}

// Discover 发现服务实例
func (r *Registry) Discover(serviceName string, opts ...DiscoveryOption) ([]*ServiceInstance, error) {
	options := &DiscoveryOptions{
		Group:       r.options.Group,
		Clusters:    []string{r.options.ClusterName},
		HealthyOnly: true,
	}

	for _, opt := range opts {
		opt(options)
	}

	instances, err := r.namingCli.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   options.Group,
		Clusters:    options.Clusters,
		HealthyOnly: options.HealthyOnly,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover service %s: %w", serviceName, err)
	}

	result := make([]*ServiceInstance, 0, len(instances))
	for _, inst := range instances {
		result = append(result, &ServiceInstance{
			ServiceName: inst.ServiceName,
			Ip:          inst.Ip,
			Port:        inst.Port,
			Weight:      inst.Weight,
			Enable:      inst.Enable,
			Healthy:     inst.Healthy,
			Metadata:    inst.Metadata,
			ClusterName: inst.ClusterName,
			GroupName:   inst.GroupName,
			Ephemeral:   inst.Ephemeral,
		})
	}

	return result, nil
}

// DiscoverOne 发现单个服务实例(带负载均衡)
func (r *Registry) DiscoverOne(serviceName string, opts ...DiscoveryOption) (*ServiceInstance, error) {
	options := &DiscoveryOptions{
		Group:       r.options.Group,
		Clusters:    []string{r.options.ClusterName},
		HealthyOnly: true,
	}

	for _, opt := range opts {
		opt(options)
	}

	instance, err := r.namingCli.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   options.Group,
		Clusters:    options.Clusters,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover one instance for service %s: %w", serviceName, err)
	}

	return &ServiceInstance{
		ServiceName: instance.ServiceName,
		Ip:          instance.Ip,
		Port:        instance.Port,
		Weight:      instance.Weight,
		Enable:      instance.Enable,
		Healthy:     instance.Healthy,
		Metadata:    instance.Metadata,
		ClusterName: instance.ClusterName,
		GroupName:   instance.GroupName,
		Ephemeral:   instance.Ephemeral,
	}, nil
}

// Subscribe 订阅服务变更
func (r *Registry) Subscribe(serviceName string, callback func(instances []*ServiceInstance), opts ...DiscoveryOption) error {
	options := &DiscoveryOptions{
		Group:       r.options.Group,
		Clusters:    []string{r.options.ClusterName},
		HealthyOnly: true,
	}

	for _, opt := range opts {
		opt(options)
	}

	err := r.namingCli.Subscribe(&vo.SubscribeParam{
		ServiceName: serviceName,
		GroupName:   options.Group,
		Clusters:    options.Clusters,
		SubscribeCallback: func(services []interface{}, err error) {
			if err != nil {
				return
			}

			instances := make([]*ServiceInstance, 0, len(services))
			for _, svc := range services {
				if inst, ok := svc.(vo.Instance); ok {
					instances = append(instances, &ServiceInstance{
						ServiceName: inst.ServiceName,
						Ip:          inst.Ip,
						Port:        inst.Port,
						Weight:      inst.Weight,
						Enable:      inst.Enable,
						Healthy:     inst.Healthy,
						Metadata:    inst.Metadata,
						ClusterName: inst.ClusterName,
						GroupName:   inst.GroupName,
						Ephemeral:   inst.Ephemeral,
					})
				}
			}

			callback(instances)
		},
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe service %s: %w", serviceName, err)
	}

	return nil
}

// Unsubscribe 取消订阅服务变更
func (r *Registry) Unsubscribe(serviceName string, opts ...DiscoveryOption) error {
	options := &DiscoveryOptions{
		Group:       r.options.Group,
		Clusters:    []string{r.options.ClusterName},
	}

	for _, opt := range opts {
		opt(options)
	}

	err := r.namingCli.Unsubscribe(&vo.SubscribeParam{
		ServiceName: serviceName,
		GroupName:   options.Group,
		Clusters:    options.Clusters,
	})

	if err != nil {
		return fmt.Errorf("failed to unsubscribe service %s: %w", serviceName, err)
	}

	return nil
}

// GetAllServices 获取所有服务名称
func (r *Registry) GetAllServices(opts ...DiscoveryOption) ([]string, error) {
	options := &DiscoveryOptions{
		Group: r.options.Group,
	}

	for _, opt := range opts {
		opt(options)
	}

	services, err := r.namingCli.GetAllServiceInfo(vo.GetAllServiceInfoParam{
		GroupName: options.Group,
		PageNo:    1,
		PageSize:  1000,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get all services: %w", err)
	}

	return services.Doms, nil
}

// Close 关闭注册中心客户端
func (r *Registry) Close() error {
	r.cancel()

	// 注销所有已注册的服务
	r.mu.Lock()
	instances := make([]*ServiceInstance, 0, len(r.instances))
	for _, inst := range r.instances {
		instances = append(instances, inst)
	}
	r.mu.Unlock()

	for _, inst := range instances {
		_ = r.Deregister(inst.ServiceName)
	}

	return nil
}

// IsRegistered 检查是否已注册
func (r *Registry) IsRegistered() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.registered
}

// DiscoveryOptions 发现选项
type DiscoveryOptions struct {
	Group       string
	Clusters    []string
	HealthyOnly bool
}

// DiscoveryOption 发现选项函数
type DiscoveryOption func(*DiscoveryOptions)

// WithDiscoveryGroup 设置发现分组
func WithDiscoveryGroup(group string) DiscoveryOption {
	return func(o *DiscoveryOptions) {
		o.Group = group
	}
}

// WithDiscoveryClusters 设置发现集群
func WithDiscoveryClusters(clusters []string) DiscoveryOption {
	return func(o *DiscoveryOptions) {
		o.Clusters = clusters
	}
}

// WithDiscoveryHealthyOnly 设置是否只发现健康实例
func WithDiscoveryHealthyOnly(healthyOnly bool) DiscoveryOption {
	return func(o *DiscoveryOptions) {
		o.HealthyOnly = healthyOnly
	}
}

// getLocalIP 获取本机IP地址
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no valid local IP address found")
}
