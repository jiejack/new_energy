package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	
	"github.com/new-energy-monitoring/internal/domain/entity"
)

// Database 数据库连接封装
type Database struct {
	*gorm.DB
	config DatabaseConfig
	mu     sync.RWMutex
}

// NewDatabase 创建数据库连接
func NewDatabase(cfg DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)
	
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
	
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	
	// 设置连接池参数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	
	database := &Database{
		DB:     db,
		config: cfg,
	}
	
	// 验证连接
	if err := database.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	return database, nil
}

// Close 关闭数据库连接
func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping 检查数据库连接
func (db *Database) Ping(ctx context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// IsReady 检查数据库是否就绪
func (db *Database) IsReady(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	return db.Ping(ctx) == nil
}

// Reconnect 重新连接数据库
func (db *Database) Reconnect(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	// 关闭旧连接
	if err := db.Close(); err != nil {
		// 记录错误但继续
	}
	
	// 创建新连接
	newDB, err := NewDatabase(db.config)
	if err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}
	
	db.DB = newDB.DB
	return nil
}

// GetStats 获取连接池统计信息
func (db *Database) GetStats() *sql.DBStats {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil
	}
	stats := sqlDB.Stats()
	return &stats
}

// HealthCheck 健康检查
func (db *Database) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	status := &HealthStatus{
		Status:  "healthy",
		Time:    time.Now(),
		Details: make(map[string]interface{}),
	}
	
	// 检查连接
	if err := db.Ping(ctx); err != nil {
		status.Status = "unhealthy"
		status.Error = err.Error()
		return status, err
	}
	
	// 获取连接池统计
	stats := db.GetStats()
	if stats != nil {
		status.Details["max_open_connections"] = stats.MaxOpenConnections
		status.Details["open_connections"] = stats.OpenConnections
		status.Details["in_use"] = stats.InUse
		status.Details["idle"] = stats.Idle
		status.Details["wait_count"] = stats.WaitCount
		status.Details["wait_duration"] = stats.WaitDuration.String()
		
		// 检查连接池使用率
		if stats.MaxOpenConnections > 0 {
			usage := float64(stats.InUse) / float64(stats.MaxOpenConnections)
			status.Details["connection_usage"] = fmt.Sprintf("%.2f%%", usage*100)
			
			// 如果使用率超过80%，标记为警告
			if usage > 0.8 {
				status.Status = "warning"
				status.Details["warning"] = "connection pool usage is high"
			}
		}
	}
	
	// 检查数据库版本
	var version string
	if err := db.DB.WithContext(ctx).Raw("SELECT version()").Scan(&version).Error; err == nil {
		status.Details["database_version"] = version
	}
	
	return status, nil
}

// AutoMigrate 自动迁移数据库表结构
func (db *Database) AutoMigrate() error {
	return db.DB.AutoMigrate(
		&entity.Region{},
		&entity.SubRegion{},
		&entity.Station{},
		&entity.Device{},
		&entity.Point{},
		&entity.Alarm{},
		&entity.AlarmRule{},
		&entity.QASession{},
		&entity.QAMessage{},
		&entity.SystemConfig{},
	)
}

// HealthStatus 健康检查状态
type HealthStatus struct {
	Status  string                 `json:"status"`
	Time    time.Time              `json:"time"`
	Error   string                 `json:"error,omitempty"`
	Details map[string]interface{} `json:"details"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}
