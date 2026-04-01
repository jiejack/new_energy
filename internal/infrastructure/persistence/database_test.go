package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 测试数据库配置
func getTestDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		DBName:          "nem_test",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// TestNewDatabase 测试创建数据库连接
func TestNewDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestDatabaseConfig()
	db, err := NewDatabase(cfg)

	require.NoError(t, err, "Failed to create database connection")
	require.NotNil(t, db, "Database should not be nil")

	// 清理
	err = db.Close()
	assert.NoError(t, err, "Failed to close database connection")
}

// TestDatabasePing 测试数据库连接
func TestDatabasePing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestDatabaseConfig()
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	err = db.Ping(ctx)
	assert.NoError(t, err, "Failed to ping database")
}

// TestDatabaseIsReady 测试数据库就绪检查
func TestDatabaseIsReady(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestDatabaseConfig()
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	ready := db.IsReady(ctx)
	assert.True(t, ready, "Database should be ready")
}

// TestDatabaseHealthCheck 测试健康检查
func TestDatabaseHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestDatabaseConfig()
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	status, err := db.HealthCheck(ctx)

	require.NoError(t, err, "Health check should not return error")
	require.NotNil(t, status, "Health status should not be nil")

	assert.Equal(t, "healthy", status.Status, "Database should be healthy")
	assert.NotEmpty(t, status.Details, "Health details should not be empty")
	assert.Contains(t, status.Details, "database_version", "Should contain database version")
	assert.Contains(t, status.Details, "open_connections", "Should contain connection stats")
}

// TestDatabaseGetStats 测试获取连接池统计
func TestDatabaseGetStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestDatabaseConfig()
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	stats := db.GetStats()
	require.NotNil(t, stats, "Stats should not be nil")

	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0, "MaxOpenConnections should be >= 0")
	assert.GreaterOrEqual(t, stats.OpenConnections, 0, "OpenConnections should be >= 0")
	assert.GreaterOrEqual(t, stats.InUse, 0, "InUse should be >= 0")
	assert.GreaterOrEqual(t, stats.Idle, 0, "Idle should be >= 0")
}

// TestDatabaseReconnect 测试重连机制
func TestDatabaseReconnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestDatabaseConfig()
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// 先关闭连接
	err = db.Close()
	require.NoError(t, err)

	// 测试重连
	err = db.Reconnect(ctx)
	require.NoError(t, err, "Reconnect should succeed")

	// 验证连接可用
	err = db.Ping(ctx)
	assert.NoError(t, err, "Ping should succeed after reconnect")
}

// TestDatabaseConfig 测试数据库配置
func TestDatabaseConfig(t *testing.T) {
	cfg := DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "test",
		Password:        "test123",
		DBName:          "testdb",
		SSLMode:         "disable",
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "test", cfg.User)
	assert.Equal(t, "test123", cfg.Password)
	assert.Equal(t, "testdb", cfg.DBName)
	assert.Equal(t, "disable", cfg.SSLMode)
	assert.Equal(t, 100, cfg.MaxOpenConns)
	assert.Equal(t, 10, cfg.MaxIdleConns)
	assert.Equal(t, time.Hour, cfg.ConnMaxLifetime)
	assert.Equal(t, 10*time.Minute, cfg.ConnMaxIdleTime)
}
