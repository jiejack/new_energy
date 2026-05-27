package persistence

import (
	"context"
	"embed"
	"io/fs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_migrations_fs/migrations/*.sql
var testMigrationsRawFS embed.FS

func getTestMigrationsFS() fs.FS {
	sub, _ := fs.Sub(testMigrationsRawFS, "test_migrations_fs")
	return sub
}

func TestMigrationManager(t *testing.T) {
	db := setupTestDB(t)

	manager := NewMigrationManager(db)
	require.NotNil(t, manager, "Migration manager should not be nil")
}

func TestCreateMigrationsTable(t *testing.T) {
	db := setupTestDB(t)

	manager := NewMigrationManager(db)
	ctx := context.Background()

	err := manager.createMigrationsTable(ctx)
	require.NoError(t, err, "Failed to create migrations table")

	var count int64
	err = db.WithContext(ctx).Table("schema_migrations").Count(&count).Error
	require.NoError(t, err)
}

func TestGetAppliedMigrations(t *testing.T) {
	db := setupTestDB(t)

	manager := NewMigrationManager(db)
	ctx := context.Background()

	err := manager.createMigrationsTable(ctx)
	require.NoError(t, err)

	applied, err := manager.getAppliedMigrations(ctx)
	require.NoError(t, err, "Failed to get applied migrations")
	assert.NotNil(t, applied, "Applied migrations map should not be nil")
}

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

func TestIsMigrationApplied(t *testing.T) {
	db := setupTestDB(t)

	manager := NewMigrationManager(db)
	ctx := context.Background()

	err := manager.createMigrationsTable(ctx)
	require.NoError(t, err)

	applied, err := manager.IsMigrationApplied(ctx, "999_nonexistent")
	require.NoError(t, err)
	assert.False(t, applied, "Nonexistent migration should not be applied")

	err = db.WithContext(ctx).Exec(
		"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
		"001_test",
		time.Now(),
	).Error
	require.NoError(t, err)

	applied, err = manager.IsMigrationApplied(ctx, "001_test")
	require.NoError(t, err)
	assert.True(t, applied, "Existing migration should be applied")
}

func TestGetMigrationStatusSummary(t *testing.T) {
	db := setupTestDB(t)

	manager := NewMigrationManager(db)
	ctx := context.Background()

	status, err := manager.GetMigrationStatusSummary(ctx, getTestMigrationsFS())
	require.NoError(t, err, "Failed to get migration status summary")
	require.NotNil(t, status, "Status should not be nil")

	assert.GreaterOrEqual(t, status.Total, 0, "Total should be >= 0")
	assert.GreaterOrEqual(t, status.Applied, 0, "Applied should be >= 0")
	assert.GreaterOrEqual(t, status.Pending, 0, "Pending should be >= 0")
	assert.Equal(t, status.Total, status.Applied+status.Pending, "Total should equal Applied + Pending")
}

func TestGetPendingMigrations(t *testing.T) {
	db := setupTestDB(t)

	manager := NewMigrationManager(db)
	ctx := context.Background()

	pending, err := manager.GetPendingMigrations(ctx, getTestMigrationsFS())
	require.NoError(t, err, "Failed to get pending migrations")
	assert.NotNil(t, pending, "Pending migrations should not be nil")
}

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

func TestMigration(t *testing.T) {
	now := time.Now()
	migration := Migration{
		Version:   "001_init",
		AppliedAt: now,
	}

	assert.Equal(t, "001_init", migration.Version)
	assert.Equal(t, now, migration.AppliedAt)
}
