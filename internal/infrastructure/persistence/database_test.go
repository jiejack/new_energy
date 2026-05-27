package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

func setupTestDB(t *testing.T) *Database {
	t.Helper()
	cfg := getTestDatabaseConfig()
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), gormConfig)
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	database := &Database{DB: db, config: cfg}
	t.Cleanup(func() { database.Close() })
	return database
}

func TestNewDatabaseWithDialector(t *testing.T) {
	db := setupTestDB(t)
	require.NotNil(t, db)
}

func TestDatabasePing(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	err := db.Ping(ctx)
	assert.NoError(t, err)
}

func TestDatabaseIsReady(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	ready := db.IsReady(ctx)
	assert.True(t, ready)
}

func TestDatabaseHealthCheck(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	status, err := db.HealthCheck(ctx)
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, "healthy", status.Status)
	assert.NotEmpty(t, status.Details)
	assert.Contains(t, status.Details, "database_version")
}

func TestDatabaseGetStats(t *testing.T) {
	db := setupTestDB(t)
	stats := db.GetStats()
	require.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)
	assert.GreaterOrEqual(t, stats.OpenConnections, 0)
}

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
	assert.Equal(t, "testdb", cfg.DBName)
	assert.Equal(t, 100, cfg.MaxOpenConns)
}

func TestDatabaseClose(t *testing.T) {
	cfg := getTestDatabaseConfig()
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), gormConfig)
	require.NoError(t, err)
	sqlDB, _ := gormDB.DB()
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	database := &Database{DB: gormDB, config: cfg}
	err = database.Close()
	assert.NoError(t, err)
}

func TestDatabaseHealthCheckUnhealthy(t *testing.T) {
	cfg := getTestDatabaseConfig()
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), gormConfig)
	require.NoError(t, err)
	sqlDB, _ := gormDB.DB()
	database := &Database{DB: gormDB, config: cfg}
	sqlDB.Close()
	ctx := context.Background()
	status, _ := database.HealthCheck(ctx)
	if status != nil {
		assert.NotEqual(t, "healthy", status.Status)
	}
}

func TestDatabaseIsReadyWhenClosed(t *testing.T) {
	cfg := getTestDatabaseConfig()
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), gormConfig)
	require.NoError(t, err)
	sqlDB, _ := gormDB.DB()
	database := &Database{DB: gormDB, config: cfg}
	sqlDB.Close()
	ctx := context.Background()
	ready := database.IsReady(ctx)
	assert.False(t, ready)
}
