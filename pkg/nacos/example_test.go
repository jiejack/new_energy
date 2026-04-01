package nacos

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ExampleUsage 展示Nacos集成的使用示例
func ExampleUsage() {
	// 1. 创建服务注册中心客户端
	registry, err := NewRegistry(
		WithServerConfigs([]ServerConfig{
			{
				IpAddr:      "127.0.0.1",
				Port:        8848,
				ContextPath: "/nacos",
				Scheme:      "http",
			},
		}),
		WithNamespace("dev"),
		WithGroup("DEFAULT_GROUP"),
		WithUsername("nacos"),
		WithPassword("nacos"),
	)
	if err != nil {
		log.Fatalf("Failed to create registry: %v", err)
	}
	defer registry.Close()

	// 2. 注册服务实例
	instance := &ServiceInstance{
		ServiceName: "new-energy-monitoring-service",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Metadata: map[string]string{
			"version": "1.0.0",
			"env":     "dev",
		},
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true, // 临时实例，会自动发送心跳
	}

	err = registry.Register(instance)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}
	fmt.Println("Service registered successfully")

	// 3. 创建健康检查器
	healthChecker, err := NewHealthChecker(registry)
	if err != nil {
		log.Fatalf("Failed to create health checker: %v", err)
	}
	defer healthChecker.Close()

	// 4. 启动心跳上报（对于持久化实例需要手动启动）
	if !instance.Ephemeral {
		err = healthChecker.StartHeartbeat(instance)
		if err != nil {
			log.Fatalf("Failed to start heartbeat: %v", err)
		}
		fmt.Println("Heartbeat started")
	}

	// 5. 服务发现
	instances, err := registry.Discover("new-energy-monitoring-service")
	if err != nil {
		log.Printf("Failed to discover service: %v", err)
	} else {
		fmt.Printf("Discovered %d instances\n", len(instances))
		for _, inst := range instances {
			fmt.Printf("  - %s:%d (weight: %.1f)\n", inst.Ip, inst.Port, inst.Weight)
		}
	}

	// 6. 负载均衡选择一个实例
	selectedInstance, err := registry.DiscoverOne("new-energy-monitoring-service")
	if err != nil {
		log.Printf("Failed to discover one instance: %v", err)
	} else {
		fmt.Printf("Selected instance: %s:%d\n", selectedInstance.Ip, selectedInstance.Port)
	}

	// 7. 订阅服务变更
	err = registry.Subscribe("new-energy-monitoring-service", func(instances []*ServiceInstance) {
		fmt.Printf("Service changed, now has %d instances\n", len(instances))
	})
	if err != nil {
		log.Printf("Failed to subscribe service: %v", err)
	}

	// 8. 创建配置中心客户端
	configClient, err := NewConfigClient(
		WithServerConfigs([]ServerConfig{
			{
				IpAddr:      "127.0.0.1",
				Port:        8848,
				ContextPath: "/nacos",
				Scheme:      "http",
			},
		}),
		WithNamespace("dev"),
		WithGroup("DEFAULT_GROUP"),
		WithUsername("nacos"),
		WithPassword("nacos"),
	)
	if err != nil {
		log.Fatalf("Failed to create config client: %v", err)
	}
	defer configClient.Close()

	// 9. 获取配置
	config, err := configClient.GetConfig("application.yaml", "DEFAULT_GROUP")
	if err != nil {
		log.Printf("Failed to get config: %v", err)
	} else {
		fmt.Printf("Config content:\n%s\n", config)
	}

	// 10. 监听配置变更
	err = configClient.ListenConfig("application.yaml", "DEFAULT_GROUP",
		func(namespace, group, dataId, content string) {
			fmt.Printf("Config changed [namespace: %s, group: %s, dataId: %s]\n", namespace, group, dataId)
			fmt.Printf("New content:\n%s\n", content)
		})
	if err != nil {
		log.Printf("Failed to listen config: %v", err)
	}

	// 11. 发布配置
	success, err := configClient.PublishConfig("test-config.yaml", "DEFAULT_GROUP", "key: value\nname: test")
	if err != nil {
		log.Printf("Failed to publish config: %v", err)
	} else if success {
		fmt.Println("Config published successfully")
	}

	// 12. 健康状态检查
	status, err := healthChecker.CheckInstance("new-energy-monitoring-service")
	if err != nil {
		log.Printf("Failed to check instance: %v", err)
	} else {
		fmt.Printf("Instance health status: %v\n", status.Healthy)
	}

	// 13. 更新实例权重
	err = healthChecker.SetInstanceWeight("new-energy-monitoring-service", 2.0)
	if err != nil {
		log.Printf("Failed to update instance weight: %v", err)
	} else {
		fmt.Println("Instance weight updated to 2.0")
	}

	// 14. 更新实例元数据
	err = healthChecker.SetInstanceMetadata("new-energy-monitoring-service", map[string]string{
		"version": "1.0.1",
		"region":  "cn-east-1",
	})
	if err != nil {
		log.Printf("Failed to update instance metadata: %v", err)
	} else {
		fmt.Println("Instance metadata updated")
	}

	// 等待一段时间以观察服务运行
	time.Sleep(30 * time.Second)

	// 15. 取消订阅
	err = registry.Unsubscribe("new-energy-monitoring-service")
	if err != nil {
		log.Printf("Failed to unsubscribe service: %v", err)
	}

	// 16. 取消配置监听
	err = configClient.CancelListenConfig("application.yaml", "DEFAULT_GROUP")
	if err != nil {
		log.Printf("Failed to cancel listen config: %v", err)
	}

	// 17. 注销服务
	err = registry.Deregister("new-energy-monitoring-service")
	if err != nil {
		log.Printf("Failed to deregister service: %v", err)
	} else {
		fmt.Println("Service deregistered successfully")
	}
}

// ExampleClusterUsage 展示集群模式的使用示例
func ExampleClusterUsage() {
	// 创建支持多集群的注册中心
	registry, err := NewRegistry(
		WithServerConfigs([]ServerConfig{
			// Nacos集群节点1
			{
				IpAddr:      "nacos1.example.com",
				Port:        8848,
				ContextPath: "/nacos",
				Scheme:      "http",
			},
			// Nacos集群节点2
			{
				IpAddr:      "nacos2.example.com",
				Port:        8848,
				ContextPath: "/nacos",
				Scheme:      "http",
			},
			// Nacos集群节点3
			{
				IpAddr:      "nacos3.example.com",
				Port:        8848,
				ContextPath: "/nacos",
				Scheme:      "http",
			},
		}),
		WithNamespace("production"),
		WithGroup("ENERGY_GROUP"),
		WithClusterName("SHANGHAI"),
	)
	if err != nil {
		log.Fatalf("Failed to create registry: %v", err)
	}
	defer registry.Close()

	// 注册服务到指定集群
	instance := &ServiceInstance{
		ServiceName: "energy-monitoring-api",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		ClusterName: "SHANGHAI",
		GroupName:   "ENERGY_GROUP",
		Ephemeral:   true,
	}

	err = registry.Register(instance)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 从多个集群发现服务
	instances, err := registry.Discover("energy-monitoring-api",
		WithDiscoveryClusters([]string{"SHANGHAI", "BEIJING"}),
		WithDiscoveryHealthyOnly(true),
	)
	if err != nil {
		log.Printf("Failed to discover service: %v", err)
	} else {
		fmt.Printf("Discovered %d instances from multiple clusters\n", len(instances))
	}
}

// ExampleMultiNamespaceUsage 展示多命名空间的使用示例
func ExampleMultiNamespaceUsage() {
	// 创建开发环境的客户端
	devRegistry, err := NewRegistry(
		WithServerConfigs([]ServerConfig{
			{
				IpAddr:      "127.0.0.1",
				Port:        8848,
				ContextPath: "/nacos",
				Scheme:      "http",
			},
		}),
		WithNamespace("dev"),
	)
	if err != nil {
		log.Fatalf("Failed to create dev registry: %v", err)
	}
	defer devRegistry.Close()

	// 创建生产环境的客户端
	prodRegistry, err := NewRegistry(
		WithServerConfigs([]ServerConfig{
			{
				IpAddr:      "127.0.0.1",
				Port:        8848,
				ContextPath: "/nacos",
				Scheme:      "http",
			},
		}),
		WithNamespace("production"),
	)
	if err != nil {
		log.Fatalf("Failed to create prod registry: %v", err)
	}
	defer prodRegistry.Close()

	// 从不同命名空间获取配置
	configClient, err := NewConfigClient(WithNamespace("dev"))
	if err != nil {
		log.Fatalf("Failed to create config client: %v", err)
	}
	defer configClient.Close()

	// 获取开发环境配置
	devConfig, err := configClient.GetConfig("application.yaml", "DEFAULT_GROUP")
	if err != nil {
		log.Printf("Failed to get dev config: %v", err)
	}

	// 获取生产环境配置（需要指定命名空间）
	prodConfig, err := configClient.GetConfigWithNamespace("application.yaml", "DEFAULT_GROUP", "production")
	if err != nil {
		log.Printf("Failed to get prod config: %v", err)
	}

	fmt.Printf("Dev config: %s\n", devConfig)
	fmt.Printf("Prod config: %s\n", prodConfig)
}

// ExampleConfigHotReload 展示配置热更新的使用示例
func ExampleConfigHotReload(ctx context.Context) {
	configClient, err := NewConfigClient(
		WithNamespace("dev"),
		WithGroup("DEFAULT_GROUP"),
	)
	if err != nil {
		log.Fatalf("Failed to create config client: %v", err)
	}
	defer configClient.Close()

	// 定义配置结构
	type AppConfig struct {
		DatabaseURL string
		RedisURL    string
		LogLevel    string
	}

	currentConfig := &AppConfig{}

	// 获取初始配置并监听变更
	initialConfig, err := configClient.GetConfigAndListen("app-config.yaml", "DEFAULT_GROUP",
		func(namespace, group, dataId, content string) {
			// 配置变更回调
			fmt.Printf("Config updated, reloading...\n")

			// 这里可以解析新的配置内容并更新应用状态
			// 例如：解析YAML，更新数据库连接池，重置日志级别等
			// currentConfig.DatabaseURL = parseConfig(content, "database_url")
			// currentConfig.RedisURL = parseConfig(content, "redis_url")
			// currentConfig.LogLevel = parseConfig(content, "log_level")

			fmt.Printf("Config reloaded successfully\n")
		})

	if err != nil {
		log.Fatalf("Failed to get and listen config: %v", err)
	}

	fmt.Printf("Initial config: %s\n", initialConfig)

	// 等待上下文取消
	<-ctx.Done()
}
