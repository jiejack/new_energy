package config_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/new-energy-monitoring/pkg/config"
)

func main() {
	// 示例1: 基本配置加载
	basicConfigLoad()

	// 示例2: 配置变更监听
	configWatch()

	// 示例3: 多环境配置
	multiEnvConfig()
}

// basicConfigLoad 基本配置加载示例
func basicConfigLoad() {
	fmt.Println("=== 基本配置加载示例 ===")

	// 创建配置加载器
	loader := config.NewLoader(
		config.WithEnv(config.EnvDev),              // 指定环境
		config.WithLocalConfig("./configs"),        // 本地配置目录
		config.WithFallback(true),                  // 启用本地配置兜底
	)

	// 加载配置
	cfg, err := loader.Load(context.Background())
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 使用配置
	fmt.Printf("Server Name: %s\n", cfg.Server.Name)
	fmt.Printf("Server Port: %d\n", cfg.Server.Port)
	fmt.Printf("Database Host: %s\n", cfg.Database.Host)
	fmt.Printf("Database Port: %d\n", cfg.Database.Port)
	fmt.Printf("Redis Addrs: %v\n", cfg.Redis.Addrs)
	fmt.Printf("Current Environment: %s\n", loader.GetEnv())

	// 获取单个配置值
	dbHost := loader.GetString("database.host")
	dbPort := loader.GetInt("database.port")
	fmt.Printf("Database: %s:%d\n", dbHost, dbPort)

	fmt.Println()
}

// configWatch 配置变更监听示例
func configWatch() {
	fmt.Println("=== 配置变更监听示例 ===")

	// 创建配置加载器（启用监听）
	loader := config.NewLoader(
		config.WithEnv(config.EnvDev),
		config.WithLocalConfig("./configs"),
		config.WithWatch(true), // 启用配置监听
	)

	// 加载配置
	_, err := loader.Load(context.Background())
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 监听配置变更
	loader.Watch("database.host", func(key string, value interface{}) {
		fmt.Printf("配置 %s 已变更: %v\n", key, value)
		// 执行配置变更回调，例如重新建立数据库连接
	})

	loader.Watch("redis.addrs", func(key string, value interface{}) {
		fmt.Printf("配置 %s 已变更: %v\n", key, value)
		// 执行配置变更回调，例如重新建立Redis连接
	})

	fmt.Println("配置监听已启动，按Ctrl+C退出...")

	// 模拟运行一段时间
	time.Sleep(5 * time.Second)

	// 关闭加载器
	if err := loader.Close(); err != nil {
		log.Printf("Failed to close loader: %v", err)
	}

	fmt.Println()
}

// multiEnvConfig 多环境配置示例
func multiEnvConfig() {
	fmt.Println("=== 多环境配置示例 ===")

	environments := []config.Environment{
		config.EnvDev,
		config.EnvTest,
		config.EnvProd,
		config.EnvStandalone,
	}

	for _, env := range environments {
		loader := config.NewLoader(
			config.WithEnv(env),
			config.WithLocalConfig("./configs"),
			config.WithFallback(true),
		)

		cfg, err := loader.Load(context.Background())
		if err != nil {
			log.Printf("Failed to load config for env %s: %v", env, err)
			continue
		}

		fmt.Printf("Environment: %s\n", env)
		fmt.Printf("  Database: %s@%s:%d/%s\n",
			cfg.Database.User,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DBName,
		)
		fmt.Printf("  Kafka Topic Prefix: %s\n", cfg.Kafka.TopicPrefix)
		fmt.Printf("  Log Level: %s\n", cfg.Logging.Level)
		fmt.Println()
	}
}

// configCenterExample 配置中心示例（需要配置中心服务）
func configCenterExample() {
	fmt.Println("=== 配置中心示例 ===")

	// 创建配置加载器（启用配置中心）
	loader := config.NewLoader(
		config.WithEnv(config.EnvProd),
		config.WithConfigCenter("nacos:8848"), // Nacos配置中心地址
		config.WithLocalConfig("./configs"),   // 本地配置作为兜底
		config.WithFallback(true),
		config.WithWatch(true),
	)

	// 加载配置
	cfg, err := loader.Load(context.Background())
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Config loaded from config center: %s\n", cfg.Server.Name)

	// 监听配置变更（配置中心会实时推送变更）
	loader.Watch("database.host", func(key string, value interface{}) {
		fmt.Printf("配置 %s 已从配置中心变更: %v\n", key, value)
	})

	// 保持运行
	select {}
}

// advancedWatcherExample 高级配置监听示例
func advancedWatcherExample() {
	fmt.Println("=== 高级配置监听示例 ===")

	loader := config.NewLoader(
		config.WithEnv(config.EnvDev),
		config.WithLocalConfig("./configs"),
		config.WithWatch(true),
	)

	_, err := loader.Load(context.Background())
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 使用高级监听器（支持变更事件）
	// watcher := config.NewConfigWatcher(loader.viper)
	// watcher.OnChange("database.host", func(event config.ConfigChangeEvent) {
	// 	fmt.Printf("配置变更事件:\n")
	// 	fmt.Printf("  Key: %s\n", event.Key)
	// 	fmt.Printf("  Old Value: %v\n", event.OldValue)
	// 	fmt.Printf("  New Value: %v\n", event.NewValue)
	// 	fmt.Printf("  Time: %s\n", event.Timestamp)
	// })

	// if err := watcher.Start(); err != nil {
	// 	log.Fatalf("Failed to start watcher: %v", err)
	// }

	fmt.Println("高级配置监听已启动...")

	// 保持运行
	select {}
}
