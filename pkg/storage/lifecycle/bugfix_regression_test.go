package lifecycle

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupFixTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&CleanupPolicy{}, &CleanupTask{}, &CleanupLog{}, &ArchivePolicy{}))
	return db
}

func newFixTestCleaner(t *testing.T) (*DataCleaner, func()) {
	db := setupFixTestDB(t)
	cfg := DefaultCleanupConfig()
	cleaner := NewDataCleaner(cfg, db, zap.L().Named("test-cleaner"))
	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}
	return cleaner, cleanup
}

func TestBUG001_DryRun_NoInfiniteLoop(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{
		Name: "dryrun-test", DataType: "archive_policies", RetentionDays: 30,
		BatchSize: 100, Enabled: true, DryRun: true,
	}
	err := cleaner.CreatePolicy(ctx, policy)
	require.NoError(t, err)

	cleaner.db.Create(&ArchivePolicy{ID: "old-1", Name: "old-data", DataType: "test", Enabled: true})
	cleaner.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-100*24*time.Hour), "old-1")

	cleaner.db.Create(&ArchivePolicy{ID: "old-2", Name: "old-data-2", DataType: "test", Enabled: true})
	cleaner.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-200*24*time.Hour), "old-2")

	task := &CleanupTask{
		ID: "dryrun-task-1", PolicyID: policy.ID, Status: CleanupStatusRunning,
		DryRun: true, CreatedAt: time.Now(),
	}
	cleaner.db.Create(task)

	done := make(chan bool, 1)
	go func() {
		cleaner.doCleanup(ctx, task, policy)
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("DryRun mode caused infinite loop - BUG-001 not fixed")
	}

	var updated CleanupTask
	cleaner.db.First(&updated, "id = ?", "dryrun-task-1")
	assert.Greater(t, updated.RecordsScanned, int64(0), "Should have scanned records")
	assert.Greater(t, updated.RecordsDeleted, int64(0), "Should have counted would-be-deleted records")
}

func TestSEC001_ValidateIdentifier_ValidNames(t *testing.T) {
	assert.NoError(t, validateIdentifier("archive_policies", "table"))
	assert.NoError(t, validateIdentifier("created_at", "column"))
	assert.NoError(t, validateIdentifier("_private", "table"))
	assert.NoError(t, validateIdentifier("Table1", "table"))
	assert.NoError(t, validateIdentifier("col_123", "column"))
}

func TestSEC001_ValidateIdentifier_InvalidNames(t *testing.T) {
	assert.Error(t, validateIdentifier("users; DROP TABLE accounts--", "table"))
	assert.Error(t, validateIdentifier("1invalid", "table"))
	assert.Error(t, validateIdentifier("has space", "column"))
	assert.Error(t, validateIdentifier("has'dquote", "column"))
	assert.Error(t, validateIdentifier("", "table"))
	assert.Error(t, validateIdentifier("a.b", "table"))
}

func TestSEC001_CleanupByDate_InvalidIdentifier(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.CleanupByDate(ctx, "valid_table; DROP TABLE users--", "created_at", time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table identifier")
}

func TestSEC001_CleanupByDate_InvalidColumn(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.CleanupByDate(ctx, "archive_policies", "1=1 OR ", time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid column identifier")
}

func TestSEC001_CleanupOrphanedRecords_InvalidIdentifier(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.CleanupOrphanedRecords(ctx, "child; DROP TABLE x--", "parent", "fk")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table identifier")
}

func TestSEC001_CleanupDuplicates_InvalidIdentifier(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.CleanupDuplicates(ctx, "table; DROP TABLE x--", []string{"name"}, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table identifier")
}

func TestSEC001_CleanupDuplicates_InvalidField(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.CleanupDuplicates(ctx, "archive_policies", []string{"name; DROP TABLE x--"}, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid column identifier")
}

func TestSEC001_VacuumTable_InvalidIdentifier(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	err := cleaner.VacuumTable(ctx, "table; DROP TABLE x--")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table identifier")
}

func TestSEC001_ReindexTable_InvalidIdentifier(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	err := cleaner.ReindexTable(ctx, "table; DROP TABLE x--")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table identifier")
}

func TestSEC001_AnalyzeTable_InvalidIdentifier(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	err := cleaner.AnalyzeTable(ctx, "table; DROP TABLE x--")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table identifier")
}

func TestSEC001_CleanupByQuery_InvalidIdentifier(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.CleanupByQuery(ctx, "table; DROP TABLE x--", "1=1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table identifier")
}

func TestBUG002_AutoCreateTime_ManualValue(t *testing.T) {
	db := setupFixTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	pastTime := time.Now().Add(-365 * 24 * time.Hour)
	policy := &CleanupPolicy{
		ID:        "manual-time-test",
		Name:      "Manual Time Test",
		DataType:  "test_data",
		CreatedAt: pastTime,
	}
	result := db.Create(policy)
	require.NoError(t, result.Error)

	var loaded CleanupPolicy
	db.First(&loaded, "id = ?", "manual-time-test")

	assert.WithinDuration(t, pastTime, loaded.CreatedAt, 2*time.Second,
		"BUG-002: autoCreateTime should respect manually set CreatedAt value")
}

func TestBUG003_TransactionSupport(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{
		Name: "txn-test", DataType: "archive_policies", RetentionDays: 30,
		BatchSize: 100, Enabled: true,
	}
	err := cleaner.CreatePolicy(ctx, policy)
	require.NoError(t, err)

	cleaner.db.Create(&ArchivePolicy{ID: "txn-old-1", Name: "old-data", DataType: "test", Enabled: true})
	cleaner.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-100*24*time.Hour), "txn-old-1")

	task := &CleanupTask{
		ID: "txn-task-1", PolicyID: policy.ID, Status: CleanupStatusRunning,
		CreatedAt: time.Now(),
	}
	cleaner.db.Create(task)

	err = cleaner.doCleanup(ctx, task, policy)
	require.NoError(t, err)

	var remaining []ArchivePolicy
	cleaner.db.Where("id = ?", "txn-old-1").Find(&remaining)
	assert.Empty(t, remaining, "Record should be deleted after cleanup")
}

func TestBUG004_CheckTaskCancelled_Throttling(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()

	cleaner.lastCancelCheck = time.Now()
	result := cleaner.checkTaskCancelled("nonexistent")
	assert.False(t, result, "Should return false when within throttle interval")
}

func TestBUG004_CheckTaskCancelled_AfterInterval(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()

	cleaner.lastCancelCheck = time.Now().Add(-10 * time.Second)
	result := cleaner.checkTaskCancelled("nonexistent")
	assert.False(t, result, "Should query DB after interval elapsed")
}

func TestBUG004_CheckTaskCancelled_CancelledTask(t *testing.T) {
	cleaner, cleanup := newFixTestCleaner(t)
	defer cleanup()

	cleaner.db.Create(&CleanupTask{
		ID: "cancel-check-task", PolicyID: "p1", Status: CleanupStatusCancelled,
		CreatedAt: time.Now(),
	})

	cleaner.lastCancelCheck = time.Now().Add(-10 * time.Second)
	result := cleaner.checkTaskCancelled("cancel-check-task")
	assert.True(t, result, "Should return true for cancelled task")
}
