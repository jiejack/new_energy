package persistence

import (
	"context"
	"embed"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_migrations/*.sql
var testMigrationsFS embed.FS

// TestMigrationManager 测试迁移管理器创建
func TestMigrationManager(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestDatabaseConfig()
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	manager := NewMigrationManager(db)
	require.NotNil(t, manager, "Migration manager should not be nil")
}

// TestCreateMigrationsTable 测试创建迁移表
func TestCreateMigrationsTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestDatabaseConfig()
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	manager := NewMigrationManager(db)
	ctx := context.Background()

	err = manager.createMigrationsTable(ctx)
	require.NoError(t, err, "Failed to create migrations table")

	// 验证表存在
	var exists bool
	err = db.WithContext(ctx).Raw(
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations')",
	).Scan(&exists).Error
	require.NoError(t, err)
	assert.True(t, exists, "Migrations table should exist")
}

// TestGetAppliedMigrations 测试获取已应用的迁移
func TestGetAppliedMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestDatabaseConfig()
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	manager := NewMigrationManager(db)
	ctx := context.Background()

	// 创建迁移表
	err = manager.createMigrationsTable(ctx)
	require.NoError(t, err)

	// 获取已应用的迁移
	applied, err := manager.getAppliedMigrations(ctx)
	require.NoError(t, err, "Failed to get applied migrations")
	assert.NotNil(t, applied, "Applied migrations map should not be nil")
}

// TestValidateMigration 测试迁移验证
func TestValidateMigration(t *testing.T) {
	manager := &MigrationManager{}

	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{
			name:    "Valid SQL",
			sql:     "CREATE TABLE test (id INT);",
			wantErr: false,
		},
		{
			name:    "Empty SQL",
			sql:     "",
			wantErr: true,
		},
		{
			name:    "Whitespace only",
			sql:     "   \n\t  ",
			wantErr: true,
		},
		{
			name:    "Valid SQL with comments",
			sql:     "-- This is a comment\nCREATE TABLE test (id INT);",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.validateMigration(tt.sql)
			if tt.wantErr {
				assert.Error(t, err, "Should return error for %s", tt.name)
			} else {
				assert.NoError(t, err, "Should not return error for %s", tt.name)
			}
		})
	}
}

// TestIsMigrationApplied 测试检查迁移是否已应用
func TestIsMigrationApplied(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestDatabaseConfig()
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	manager := NewMigrationManager(db)
	ctx := context.Background()

	// 创建迁移表
	err = manager.createMigrationsTable(ctx)
	require.NoError(t, err)

	// 检查不存在的迁移
	applied, err := manager.IsMigrationApplied(ctx, "999_nonexistent")
	require.NoError(t, err)
	assert.False(t, applied, "Nonexistent migration should not be applied")

	// 插入一个迁移记录
	err = db.WithContext(ctx).Exec(
		"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
		"001_test",
		time.Now(),
	).Error
	require.NoError(t, err)

	// 检查已存在的迁移
	applied, err = manager.IsMigrationApplied(ctx, "001_test")
	require.NoError(t, err)
	assert.True(t, applied, "Existing migration should be applied")
}

// TestGetMigrationStatusSummary 测试获取迁移状态摘要
func TestGetMigrationStatusSummary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestDatabaseConfig()
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	manager := NewMigrationManager(db)
	ctx := context.Background()

	// 获取状态摘要
	status, err := manager.GetMigrationStatusSummary(ctx, testMigrationsFS)
	require.NoError(t, err, "Failed to get migration status summary")
	require.NotNil(t, status, "Status should not be nil")

	assert.GreaterOrEqual(t, status.Total, 0, "Total should be >= 0")
	assert.GreaterOrEqual(t, status.Applied, 0, "Applied should be >= 0")
	assert.GreaterOrEqual(t, status.Pending, 0, "Pending should be >= 0")
	assert.Equal(t, status.Total, status.Applied+status.Pending, "Total should equal Applied + Pending")
}

// TestGetPendingMigrations 测试获取待执行的迁移
func TestGetPendingMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestDatabaseConfig()
	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	manager := NewMigrationManager(db)
	ctx := context.Background()

	// 获取待执行的迁移
	pending, err := manager.GetPendingMigrations(ctx, testMigrationsFS)
	require.NoError(t, err, "Failed to get pending migrations")
	assert.NotNil(t, pending, "Pending migrations should not be nil")
}

// TestMigrationStatus 测试迁移状态结构
func TestMigrationStatus(t *testing.T) {
	status := MigrationStatus{
		Total:   10,
		Applied: 7,
		Pending: 3,
		LastApplied: &Migration{
			Version:   "005_test",
			AppliedAt: time.Now(),
		},
	}

	assert.Equal(t, 10, status.Total)
	assert.Equal(t, 7, status.Applied)
	assert.Equal(t, 3, status.Pending)
	assert.NotNil(t, status.LastApplied)
	assert.Equal(t, "005_test", status.LastApplied.Version)
}

// TestMigration 测试迁移记录结构
func TestMigration(t *testing.T) {
	now := time.Now()
	migration := Migration{
		Version:   "001_init",
		AppliedAt: now,
	}

	assert.Equal(t, "001_init", migration.Version)
	assert.Equal(t, now, migration.AppliedAt)
}
