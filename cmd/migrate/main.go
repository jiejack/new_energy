package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/new-energy-monitoring/internal/infrastructure/config"
	"github.com/new-energy-monitoring/internal/infrastructure/persistence"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	// 加载配置
	cfg, err := config.Load("../../configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建数据库连接
	dbConfig := persistence.DatabaseConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
	}

	db, err := persistence.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	// 创建迁移管理器
	migrationManager := persistence.NewMigrationManager(db)

	// 获取迁移状态摘要
	ctx := context.Background()
	status, err := migrationManager.GetMigrationStatusSummary(ctx, migrationsFS)
	if err != nil {
		log.Printf("Failed to get migration status: %v", err)
	} else {
		fmt.Printf("\nMigration Status:\n")
		fmt.Printf("  Total:   %d\n", status.Total)
		fmt.Printf("  Applied: %d\n", status.Applied)
		fmt.Printf("  Pending: %d\n", status.Pending)
		if status.LastApplied != nil {
			fmt.Printf("  Last Applied: %s at %s\n", 
				status.LastApplied.Version, 
				status.LastApplied.AppliedAt.Format(time.RFC3339))
		}
	}

	// 执行迁移
	fmt.Printf("\nRunning migrations...\n")
	if err := migrationManager.RunMigrations(ctx, migrationsFS); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	// 健康检查
	healthStatus, err := db.HealthCheck(ctx)
	if err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		fmt.Printf("\nHealth Check:\n")
		fmt.Printf("  Status: %s\n", healthStatus.Status)
		fmt.Printf("  Time: %s\n", healthStatus.Time.Format(time.RFC3339))
		if version, ok := healthStatus.Details["database_version"].(string); ok {
			fmt.Printf("  Database Version: %s\n", version)
		}
		if usage, ok := healthStatus.Details["connection_usage"].(string); ok {
			fmt.Printf("  Connection Usage: %s\n", usage)
		}
	}

	// 获取连接池统计
	stats := db.GetStats()
	if stats != nil {
		fmt.Printf("\nConnection Pool Stats:\n")
		fmt.Printf("  Max Open Connections: %d\n", stats.MaxOpenConnections)
		fmt.Printf("  Open Connections: %d\n", stats.OpenConnections)
		fmt.Printf("  In Use: %d\n", stats.InUse)
		fmt.Printf("  Idle: %d\n", stats.Idle)
		fmt.Printf("  Wait Count: %d\n", stats.WaitCount)
		fmt.Printf("  Wait Duration: %s\n", stats.WaitDuration)
	}

	// 监控连接池（每10秒打印一次统计信息）
	go monitorConnectionPool(db)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
}

func monitorConnectionPool(db *persistence.Database) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := db.GetStats()
		if stats != nil {
			log.Printf("[Monitor] Connections - Open: %d, InUse: %d, Idle: %d, WaitCount: %d",
				stats.OpenConnections,
				stats.InUse,
				stats.Idle,
				stats.WaitCount,
			)
		}

		// 检查连接池使用率
		if stats.MaxOpenConnections > 0 {
			usage := float64(stats.InUse) / float64(stats.MaxOpenConnections)
			if usage > 0.8 {
				log.Printf("[Warning] Connection pool usage is high: %.2f%%", usage*100)
			}
		}
	}
}
