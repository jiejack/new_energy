package lifecycle

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.AutoMigrate(&ArchivePolicy{}, &ArchiveTask{}, &ArchiveRecord{})
	db.AutoMigrate(&BackupPolicy{}, &BackupRecord{}, &BackupTableRecord{}, &RestoreRecord{})
	db.AutoMigrate(&CleanupPolicy{}, &CleanupTask{}, &CleanupLog{})
	return db
}

func newTestArchiver(t *testing.T) (*DataArchiver, string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "lifecycle_test_*")
	assert.NoError(t, err)
	db := setupTestDB(t)
	logger := zap.NewNop()
	config := DefaultArchiveConfig()
	config.StoragePath = filepath.Join(dir, "archive")
	config.TempPath = filepath.Join(dir, "tmp")
	config.ParallelWorkers = 1
	config.RetryCount = 1
	config.RetryDelay = 10 * time.Millisecond
	config.EnableVerify = false
	archiver := NewDataArchiver(config, db, logger)
	return archiver, dir, func() { os.RemoveAll(dir) }
}

func newTestBackupManager(t *testing.T) (*BackupManager, string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "lifecycle_test_*")
	assert.NoError(t, err)
	db := setupTestDB(t)
	logger := zap.NewNop()
	config := DefaultBackupConfig()
	config.StoragePath = filepath.Join(dir, "backup")
	config.TempPath = filepath.Join(dir, "tmp")
	config.ParallelWorkers = 1
	config.RetryCount = 1
	config.RetryDelay = 10 * time.Millisecond
	config.EnableVerify = false
	config.ScheduleInterval = 1 * time.Hour
	bm := NewBackupManager(config, db, logger)
	return bm, dir, func() { os.RemoveAll(dir) }
}

func newTestCleaner(t *testing.T) (*DataCleaner, func()) {
	t.Helper()
	db := setupTestDB(t)
	logger := zap.NewNop()
	config := DefaultCleanupConfig()
	config.ParallelWorkers = 1
	config.MaxRetryCount = 1
	config.RetryDelay = 10 * time.Millisecond
	config.ScheduleInterval = 1 * time.Hour
	config.EnableAutoCleanup = false
	cleaner := NewDataCleaner(config, db, logger)
	return cleaner, func() {}
}

func TestDataArchiver_StartStop(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := archiver.Start(ctx)
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	err = archiver.Stop()
	assert.NoError(t, err)
}

func TestDataArchiver_CreatePolicy(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		Name: "test-policy", Description: "test", DataType: "test_data",
		RetentionDays: 30, Format: ArchiveFormatJSON, Compression: true, BatchSize: 1000, Enabled: true,
	}
	err := archiver.CreatePolicy(ctx, policy)
	assert.NoError(t, err)
	assert.NotEmpty(t, policy.ID)
	assert.False(t, policy.CreatedAt.IsZero())
}

func TestDataArchiver_CreatePolicy_WithID(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{ID: "custom-id", Name: "test", DataType: "test_data", Enabled: true}
	err := archiver.CreatePolicy(ctx, policy)
	assert.NoError(t, err)
	assert.Equal(t, "custom-id", policy.ID)
}

func TestDataArchiver_UpdatePolicy(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{Name: "test", DataType: "test_data", Enabled: true}
	archiver.CreatePolicy(ctx, policy)
	policy.Name = "updated"
	err := archiver.UpdatePolicy(ctx, policy)
	assert.NoError(t, err)
}

func TestDataArchiver_DeletePolicy(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{Name: "test", DataType: "test_data", Enabled: true}
	archiver.CreatePolicy(ctx, policy)
	err := archiver.DeletePolicy(ctx, policy.ID)
	assert.NoError(t, err)
	_, err = archiver.GetPolicy(ctx, policy.ID)
	assert.Error(t, err)
}

func TestDataArchiver_GetPolicy(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{Name: "test", DataType: "test_data", Enabled: true}
	archiver.CreatePolicy(ctx, policy)
	result, err := archiver.GetPolicy(ctx, policy.ID)
	assert.NoError(t, err)
	assert.Equal(t, "test", result.Name)
}

func TestDataArchiver_GetPolicy_NotFound(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	_, err := archiver.GetPolicy(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get policy")
}

func TestDataArchiver_ListPolicies(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	archiver.CreatePolicy(ctx, &ArchivePolicy{Name: "p1", DataType: "d1", Enabled: true})
	archiver.CreatePolicy(ctx, &ArchivePolicy{Name: "p2", DataType: "d2", Enabled: true})
	policies, err := archiver.ListPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, policies, 2)
}

func TestDataArchiver_TriggerArchive(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{Name: "test", DataType: "test_data", Enabled: true, ArchiveAfter: 24 * time.Hour, BatchSize: 100}
	archiver.CreatePolicy(ctx, policy)
	task, err := archiver.TriggerArchive(ctx, policy.ID)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, ArchiveStatusPending, task.Status)
}

func TestDataArchiver_TriggerArchive_Disabled(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{Name: "test", DataType: "test_data", Enabled: false}
	archiver.CreatePolicy(ctx, policy)
	_, err := archiver.TriggerArchive(ctx, policy.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestDataArchiver_GetTask(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	archiver.db.Create(&ArchiveTask{ID: "task-1", PolicyID: "p1", Status: ArchiveStatusPending, CreatedAt: time.Now()})
	result, err := archiver.GetTask(ctx, "task-1")
	assert.NoError(t, err)
	assert.Equal(t, "task-1", result.ID)
}

func TestDataArchiver_GetTask_NotFound(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	_, err := archiver.GetTask(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestDataArchiver_ListTasks(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	archiver.db.Create(&ArchiveTask{ID: "t1", PolicyID: "p1", Status: ArchiveStatusPending, CreatedAt: time.Now()})
	archiver.db.Create(&ArchiveTask{ID: "t2", PolicyID: "p1", Status: ArchiveStatusCompleted, CreatedAt: time.Now()})
	tasks, err := archiver.ListTasks(ctx, "p1", 10)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestDataArchiver_ListTasks_All(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	archiver.db.Create(&ArchiveTask{ID: "t1", PolicyID: "p1", Status: ArchiveStatusPending, CreatedAt: time.Now()})
	tasks, err := archiver.ListTasks(ctx, "", 0)
	assert.NoError(t, err)
	assert.Len(t, tasks, 1)
}

func TestDataArchiver_GetArchiveStats(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	archiver.db.Create(&ArchiveTask{ID: "t1", PolicyID: "p1", Status: ArchiveStatusCompleted, RecordsCount: 100, DataSize: 1024, CreatedAt: time.Now()})
	stats, err := archiver.GetArchiveStats(ctx, "p1")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestDataArchiver_GetArchiveStats_All(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	archiver.db.Create(&ArchiveTask{ID: "t1", PolicyID: "p1", Status: ArchiveStatusCompleted, RecordsCount: 100, DataSize: 1024, CreatedAt: time.Now()})
	stats, err := archiver.GetArchiveStats(ctx, "")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestArchiveBuffer_Concurrent(t *testing.T) {
	buf := NewArchiveBuffer()
	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			buf.Write([]byte("a"))
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 100; i++ {
			buf.Len()
		}
		done <- true
	}()
	<-done
	<-done
	assert.Equal(t, 100, buf.Len())
}

func TestMin(t *testing.T) {
	assert.Equal(t, 3, min(3, 5))
	assert.Equal(t, 5, min(10, 5))
	assert.Equal(t, 0, min(0, 5))
}

func TestWarmData_TableName(t *testing.T) {
	assert.Equal(t, "warm_data", WarmData{}.TableName())
}

func TestColdData_TableName(t *testing.T) {
	assert.Equal(t, "cold_data", ColdData{}.TableName())
}

func TestBackupManager_StartStop(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bm.Start(ctx)
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	err = bm.Stop()
	assert.NoError(t, err)
}

func TestBackupManager_CreatePolicy(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Tables: []string{"t1"}, Enabled: true}
	err := bm.CreatePolicy(ctx, policy)
	assert.NoError(t, err)
	assert.NotEmpty(t, policy.ID)
	assert.Equal(t, 30, policy.RetentionDays)
	assert.Equal(t, 10, policy.MaxBackups)
}

func TestBackupManager_CreatePolicy_WithID(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{ID: "custom-id", Name: "test", Type: BackupTypeFull, Enabled: true}
	err := bm.CreatePolicy(ctx, policy)
	assert.NoError(t, err)
	assert.Equal(t, "custom-id", policy.ID)
}

func TestBackupManager_UpdatePolicy(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Enabled: true}
	bm.CreatePolicy(ctx, policy)
	policy.Name = "updated"
	err := bm.UpdatePolicy(ctx, policy)
	assert.NoError(t, err)
}

func TestBackupManager_DeletePolicy(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Enabled: true}
	bm.CreatePolicy(ctx, policy)
	err := bm.DeletePolicy(ctx, policy.ID)
	assert.NoError(t, err)
}

func TestBackupManager_GetPolicy(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Enabled: true}
	bm.CreatePolicy(ctx, policy)
	result, err := bm.GetPolicy(ctx, policy.ID)
	assert.NoError(t, err)
	assert.Equal(t, "test", result.Name)
}

func TestBackupManager_GetPolicy_NotFound(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	_, err := bm.GetPolicy(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestBackupManager_ListPolicies(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.CreatePolicy(ctx, &BackupPolicy{Name: "p1", Type: BackupTypeFull, Enabled: true})
	bm.CreatePolicy(ctx, &BackupPolicy{Name: "p2", Type: BackupTypeIncrement, Enabled: true})
	policies, err := bm.ListPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, policies, 2)
}

func TestBackupManager_TriggerBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Tables: []string{"t1"}, Compression: false, Enabled: true}
	bm.CreatePolicy(ctx, policy)
	record, err := bm.TriggerBackup(ctx, policy.ID)
	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, BackupStatusPending, record.Status)
}

func TestBackupManager_TriggerBackup_Disabled(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Enabled: false}
	bm.CreatePolicy(ctx, policy)
	_, err := bm.TriggerBackup(ctx, policy.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestBackupManager_GetBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted})
	result, err := bm.GetBackup(ctx, "b1")
	assert.NoError(t, err)
	assert.Equal(t, "b1", result.ID)
}

func TestBackupManager_GetBackup_NotFound(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	_, err := bm.GetBackup(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestBackupManager_ListBackups(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", PolicyID: "p1", Status: BackupStatusCompleted, CreatedAt: time.Now()})
	bm.db.Create(&BackupRecord{ID: "b2", PolicyID: "p1", Status: BackupStatusCompleted, CreatedAt: time.Now()})
	records, err := bm.ListBackups(ctx, "p1", 10)
	assert.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestBackupManager_RestoreBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted})
	restore, err := bm.RestoreBackup(ctx, "b1", []string{"t1"})
	assert.NoError(t, err)
	assert.Equal(t, BackupStatusPending, restore.Status)
}

func TestBackupManager_RestoreBackup_NotCompleted(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusPending})
	_, err := bm.RestoreBackup(ctx, "b1", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not completed")
}

func TestBackupManager_GetRestore(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&RestoreRecord{ID: "r1", BackupID: "b1", Status: BackupStatusCompleted})
	result, err := bm.GetRestore(ctx, "r1")
	assert.NoError(t, err)
	assert.Equal(t, "r1", result.ID)
}

func TestBackupManager_GetRestore_NotFound(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	_, err := bm.GetRestore(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestBackupManager_ListRestores(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&RestoreRecord{ID: "r1", BackupID: "b1", Status: BackupStatusCompleted, CreatedAt: time.Now()})
	bm.db.Create(&RestoreRecord{ID: "r2", BackupID: "b1", Status: BackupStatusCompleted, CreatedAt: time.Now()})
	records, err := bm.ListRestores(ctx, "b1", 10)
	assert.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestBackupManager_DeleteBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted})
	err := bm.DeleteBackup(ctx, "b1")
	assert.NoError(t, err)
}

func TestBackupManager_DeleteBackup_WithFile(t *testing.T) {
	bm, dir, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	filePath := filepath.Join(dir, "backup", "test.json")
	os.MkdirAll(filepath.Dir(filePath), 0755)
	os.WriteFile(filePath, []byte("test"), 0644)

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted, FilePath: filePath})
	err := bm.DeleteBackup(ctx, "b1")
	assert.NoError(t, err)
	assert.NoFileExists(t, filePath)
}

func TestBackupManager_GetBackupStats(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", PolicyID: "p1", Status: BackupStatusCompleted, Size: 2048, CompressedSize: 1024, RecordsCount: 100, CreatedAt: time.Now()})
	stats, err := bm.GetBackupStats(ctx, "p1")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestBackupManager_CompareBackups(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Size: 1024, RecordsCount: 100, CreatedAt: time.Now().Add(-24 * time.Hour)})
	bm.db.Create(&BackupRecord{ID: "b2", Size: 2048, RecordsCount: 200, CreatedAt: time.Now()})
	result, err := bm.CompareBackups(ctx, "b1", "b2")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestBackupManager_ScheduleBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Enabled: true}
	bm.CreatePolicy(ctx, policy)
	err := bm.ScheduleBackup(ctx, policy.ID, time.Now().Add(24*time.Hour))
	assert.NoError(t, err)
}

func TestBackupManager_CancelScheduledBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusPending})
	err := bm.CancelScheduledBackup(ctx, "b1")
	assert.NoError(t, err)
}

func TestBackupManager_CancelScheduledBackup_NotPending(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted})
	err := bm.CancelScheduledBackup(ctx, "b1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not pending")
}

func TestBackupManager_GetBackupTables(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupTableRecord{ID: "t1", BackupID: "b1", TableName: "table1", RecordsCount: 100, Size: 1024})
	tables, err := bm.GetBackupTables(ctx, "b1")
	assert.NoError(t, err)
	assert.Len(t, tables, 1)
}

func TestDataCleaner_StartStop(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cleaner.Start(ctx)
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	err = cleaner.Stop()
	assert.NoError(t, err)
}

func TestDataCleaner_CreatePolicy(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "test_data", RetentionDays: 30, Enabled: true}
	err := cleaner.CreatePolicy(ctx, policy)
	assert.NoError(t, err)
	assert.NotEmpty(t, policy.ID)
	assert.Equal(t, 1000, policy.BatchSize)
}

func TestDataCleaner_CreatePolicy_WithID(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{ID: "custom-id", Name: "test", DataType: "test_data", Enabled: true}
	err := cleaner.CreatePolicy(ctx, policy)
	assert.NoError(t, err)
	assert.Equal(t, "custom-id", policy.ID)
}

func TestDataCleaner_UpdatePolicy(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "test_data", Enabled: true}
	cleaner.CreatePolicy(ctx, policy)
	policy.Name = "updated"
	err := cleaner.UpdatePolicy(ctx, policy)
	assert.NoError(t, err)
}

func TestDataCleaner_DeletePolicy(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "test_data", Enabled: true}
	cleaner.CreatePolicy(ctx, policy)
	err := cleaner.DeletePolicy(ctx, policy.ID)
	assert.NoError(t, err)
}

func TestDataCleaner_GetPolicy(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "test_data", Enabled: true}
	cleaner.CreatePolicy(ctx, policy)
	result, err := cleaner.GetPolicy(ctx, policy.ID)
	assert.NoError(t, err)
	assert.Equal(t, "test", result.Name)
}

func TestDataCleaner_GetPolicy_NotFound(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.GetPolicy(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestDataCleaner_ListPolicies(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	cleaner.CreatePolicy(ctx, &CleanupPolicy{Name: "p1", DataType: "d1", Enabled: true})
	cleaner.CreatePolicy(ctx, &CleanupPolicy{Name: "p2", DataType: "d2", Enabled: true})
	policies, err := cleaner.ListPolicies(ctx)
	assert.NoError(t, err)
	assert.Len(t, policies, 2)
}

func TestDataCleaner_TriggerCleanup(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "test_data", Enabled: true, BatchSize: 100}
	cleaner.CreatePolicy(ctx, policy)
	task, err := cleaner.TriggerCleanup(ctx, policy.ID)
	assert.NoError(t, err)
	assert.Equal(t, CleanupStatusPending, task.Status)
}

func TestDataCleaner_TriggerCleanup_Disabled(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "test_data", Enabled: false}
	cleaner.CreatePolicy(ctx, policy)
	_, err := cleaner.TriggerCleanup(ctx, policy.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestDataCleaner_GetTask(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	cleaner.db.Create(&CleanupTask{ID: "t1", PolicyID: "p1", Status: CleanupStatusPending, CreatedAt: time.Now()})
	result, err := cleaner.GetTask(ctx, "t1")
	assert.NoError(t, err)
	assert.Equal(t, "t1", result.ID)
}

func TestDataCleaner_GetTask_NotFound(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.GetTask(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestDataCleaner_ListTasks(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	cleaner.db.Create(&CleanupTask{ID: "t1", PolicyID: "p1", Status: CleanupStatusPending, CreatedAt: time.Now()})
	cleaner.db.Create(&CleanupTask{ID: "t2", PolicyID: "p1", Status: CleanupStatusCompleted, CreatedAt: time.Now()})
	tasks, err := cleaner.ListTasks(ctx, "p1", 10)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestDataCleaner_CancelTask(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	cleaner.db.Create(&CleanupTask{ID: "t1", PolicyID: "p1", Status: CleanupStatusPending, CreatedAt: time.Now()})
	err := cleaner.CancelTask(ctx, "t1")
	assert.NoError(t, err)
	result, _ := cleaner.GetTask(ctx, "t1")
	assert.Equal(t, CleanupStatusCancelled, result.Status)
}

func TestDataCleaner_CancelTask_CannotCancel(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	cleaner.db.Create(&CleanupTask{ID: "t1", PolicyID: "p1", Status: CleanupStatusCompleted, CreatedAt: time.Now()})
	err := cleaner.CancelTask(ctx, "t1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be cancelled")
}

func TestDataCleaner_GetCleanupLogs(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	cleaner.db.Create(&CleanupLog{ID: "l1", TaskID: "t1", Level: "info", Message: "test", CreatedAt: time.Now()})
	logs, err := cleaner.GetCleanupLogs(ctx, "t1", 10)
	assert.NoError(t, err)
	assert.Len(t, logs, 1)
}

func TestDataCleaner_PreviewCleanup(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "archive_policies", RetentionDays: 30, Enabled: true}
	cleaner.CreatePolicy(ctx, policy)
	count, err := cleaner.PreviewCleanup(ctx, policy.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestDataCleaner_CleanupBatch_Empty(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	deleted, err := cleaner.CleanupBatch(ctx, "test_table", []string{})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), deleted)
}

func TestDataCleaner_CleanupDuplicates_EmptyFields(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.CleanupDuplicates(ctx, "test_table", []string{}, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unique fields cannot be empty")
}

func TestDataCleaner_EstimateCleanupSize(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "archive_policies", RetentionDays: 30, Enabled: true}
	cleaner.CreatePolicy(ctx, policy)
	result, err := cleaner.EstimateCleanupSize(ctx, policy.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDataArchiver_StreamArchive_HandlerError(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	records := []map[string]interface{}{{"id": "1"}}
	filePath, err := archiver.ArchiveData(ctx, "test_type", records)
	assert.NoError(t, err)
	err = archiver.StreamArchive(ctx, filePath, func(record map[string]interface{}) error {
		return fmt.Errorf("handler error")
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler error")
}

func TestDataArchiver_ReadArchive_FileNotFound(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	_, err := archiver.ReadArchive(ctx, "/nonexistent/file.json")
	assert.Error(t, err)
}

func TestDataArchiver_CalculateChecksum_FileNotFound(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()

	_, err := archiver.CalculateChecksum("/nonexistent/file.json")
	assert.Error(t, err)
}

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	return mr, func() { mr.Close() }
}

func newTestTieredStorage(t *testing.T) (*TieredStorage, func()) {
	t.Helper()
	mr, closeRedis := setupTestRedis(t)
	db := setupTestDB(t)
	db.AutoMigrate(&WarmData{}, &ColdData{})
	logger := zap.NewNop()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
		DB:   0,
	})

	config := DefaultTierConfig()
	config.MigrationInterval = 100 * time.Millisecond
	config.StatsInterval = 100 * time.Millisecond

	ts := NewTieredStorage(config, rdb, db, logger)
	return ts, func() {
		ts.Stop()
		rdb.Close()
		closeRedis()
	}
}

func TestNewTieredStorage(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	assert.NotNil(t, ts)
	assert.NotNil(t, ts.stopCh)
}

func TestTieredStorage_StoreAndGet(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{
		"id":   "stat-1",
		"type": "metric",
	}
	err := ts.Store(ctx, "key1", data, 1*time.Hour)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = ts.Get(ctx, "key1", &result)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestTieredStorage_StoreHot(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "stat-1", "type": "metric"}
	err := ts.Store(ctx, "hot-key", data, 1*time.Hour)
	assert.NoError(t, err)
	tier, _ := ts.GetTier(ctx, "hot-key")
	assert.Equal(t, TierHot, tier)
}

func TestTieredStorage_StoreWarm(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := []byte(`{"id":"stat-2","type":"metric"}`)
	err := ts.storeWarm(ctx, "warm-key", data)
	if err != nil {
		assert.Contains(t, err.Error(), "NOW")
	}
}

func TestTieredStorage_StoreCold(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := []byte(`{"id":"stat-3","type":"metric"}`)
	err := ts.storeCold(ctx, "cold-key", data)
	if err != nil {
		assert.Contains(t, err.Error(), "NOW")
	}
}

func TestTieredStorage_Get_NotFound(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	var result map[string]interface{}
	err := ts.Get(ctx, "nonexistent", &result)
	assert.Error(t, err)
}

func TestTieredStorage_Delete(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "stat-1"}
	ts.Store(ctx, "del-key", data, 1*time.Hour)

	err := ts.Delete(ctx, "del-key")
	assert.NoError(t, err)

	var result map[string]interface{}
	err = ts.Get(ctx, "del-key", &result)
	assert.Error(t, err)
}

func TestTieredStorage_Delete_NotFound(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.Delete(ctx, "nonexistent")
	assert.NoError(t, err)
}

func TestTieredStorage_GetTier_Unknown(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	tier, err := ts.GetTier(ctx, "unknown-key")
	assert.Error(t, err)
	assert.Equal(t, DataTier(0), tier)
}

func TestTieredStorage_Migrate(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "stat-1"}
	ts.Store(ctx, "migrate-key", data, 1*time.Hour)

	err := ts.Migrate(ctx, "migrate-key", TierHot, TierWarm)
	if err != nil {
		assert.Contains(t, err.Error(), "NOW")
	}
}

func TestTieredStorage_Migrate_NotFound(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.Migrate(ctx, "nonexistent", TierHot, TierWarm)
	assert.Error(t, err)
}

func TestTieredStorage_StartStop(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := ts.Start(ctx)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
}

func TestTieredStorage_AutoMigrate(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()

	err := ts.AutoMigrate()
	assert.NoError(t, err)
}

func TestTieredStorage_GetMetrics(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()

	metrics := ts.GetMetrics()
	assert.NotNil(t, metrics)
}

func TestDataArchiver_ArchiveData_EmptyRecords(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	_, err := archiver.ArchiveData(ctx, "test_type", []map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no records to archive")
}

func TestDataArchiver_ArchiveData_ReadArchive_RoundTrip(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	records := []map[string]interface{}{
		{"id": "1", "name": "test1"},
		{"id": "2", "name": "test2"},
	}

	filePath, err := archiver.ArchiveData(ctx, "test_type", records)
	assert.NoError(t, err)
	assert.FileExists(t, filePath)

	readRecords, err := archiver.ReadArchive(ctx, filePath)
	assert.NoError(t, err)
	assert.Len(t, readRecords, 2)
}

func TestDataArchiver_CalculateChecksum_Deterministic(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	records := []map[string]interface{}{{"id": "1"}}
	filePath, _ := archiver.ArchiveData(ctx, "test_type", records)

	c1, _ := archiver.CalculateChecksum(filePath)
	c2, _ := archiver.CalculateChecksum(filePath)
	assert.Equal(t, c1, c2)
}

func TestBackupManager_GetMetrics(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()

	metrics := bm.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.TotalBackups)
}

func TestDataCleaner_GetMetrics(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()

	metrics := cleaner.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.TotalTasks)
}

func TestDataCleaner_CleanupByQuery(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	deleted, err := cleaner.CleanupByQuery(ctx, "archive_policies", "1=0")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), deleted)
}

func TestDataCleaner_CleanupByDate(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	deleted, err := cleaner.CleanupByDate(ctx, "archive_policies", "created_at", time.Now().Add(-30*24*time.Hour))
	assert.NoError(t, err)
	assert.Equal(t, int64(0), deleted)
}

func TestDataCleaner_CleanupOrphanedRecords(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	deleted, err := cleaner.CleanupOrphanedRecords(ctx, "archive_records", "archive_tasks", "task_id")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), deleted)
}

func TestDataCleaner_GetTableSize(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.GetTableSize(ctx, "archive_policies")
	if err != nil {
		assert.Contains(t, err.Error(), "pg_total_relation_size")
	}
}

func TestDataCleaner_AnalyzeTable(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	err := cleaner.AnalyzeTable(ctx, "archive_policies")
	assert.NoError(t, err)
}

func TestDataCleaner_GetStorageStats(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	stats, err := cleaner.GetStorageStats(ctx, "archive_policies")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestDataCleaner_VacuumTable(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	err := cleaner.VacuumTable(ctx, "archive_policies")
	if err != nil {
		assert.Contains(t, err.Error(), "syntax error")
	}
}

func TestDataCleaner_ReindexTable(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	err := cleaner.ReindexTable(ctx, "archive_policies")
	if err != nil {
		assert.Contains(t, err.Error(), "syntax error")
	}
}

func TestDataCleaner_CleanupBatch_WithIDs(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "archive_policies", Enabled: true}
	cleaner.CreatePolicy(ctx, policy)

	deleted, err := cleaner.CleanupBatch(ctx, "archive_policies", []string{"nonexistent-id"})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), deleted)
}

func TestDataCleaner_CleanupDuplicates_WithFields(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.CleanupDuplicates(ctx, "archive_policies", []string{"name"}, false)
	if err != nil {
		assert.Contains(t, err.Error(), "ON")
	}
}

func TestDataArchiver_RestoreArchive(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	err := archiver.RestoreArchive(ctx, "/nonexistent/file.json", "test_type")
	assert.Error(t, err)
}

func TestDataArchiver_CompactArchives(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{Name: "test", DataType: "test_data", Enabled: true, Format: ArchiveFormatJSON}
	archiver.CreatePolicy(ctx, policy)

	err := archiver.CompactArchives(ctx, policy.ID)
	if err != nil {
		assert.Contains(t, err.Error(), "failed")
	}
}

func TestBackupManager_CreateFullBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	record, err := bm.CreateFullBackup(ctx, []string{"archive_policies"}, "test full backup")
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestBackupManager_CreateIncrementBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeIncrement, Tables: []string{"archive_policies"}, Compression: false, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	record, err := bm.CreateIncrementBackup(ctx, policy.ID)
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestBackupManager_VerifyBackup(t *testing.T) {
	bm, dir, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(bm.config.StoragePath, 0755)
	filePath := filepath.Join(dir, "backup", "test.json")
	os.MkdirAll(filepath.Dir(filePath), 0755)
	os.WriteFile(filePath, []byte("test"), 0644)

	archiver, _, _ := newTestArchiver(t)
	checksum, _ := archiver.CalculateChecksum(filePath)
	bm.db.Create(&BackupRecord{ID: "b1", PolicyID: "p1", Status: BackupStatusCompleted, FilePath: filePath, Checksum: checksum})
	err := bm.VerifyBackup(ctx, "b1")
	assert.NoError(t, err)
}

func TestBackupManager_VerifyBackup_NotFound(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	err := bm.VerifyBackup(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestBackupManager_ExportBackup(t *testing.T) {
	bm, dir, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(bm.config.StoragePath, 0755)
	filePath := filepath.Join(dir, "backup", "test.json")
	os.MkdirAll(filepath.Dir(filePath), 0755)
	os.WriteFile(filePath, []byte("test"), 0644)

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted, FilePath: filePath})

	exportPath := filepath.Join(dir, "export.json")
	err := bm.ExportBackup(ctx, "b1", exportPath)
	assert.NoError(t, err)
	assert.FileExists(t, exportPath)
}

func TestBackupManager_ImportBackup(t *testing.T) {
	bm, dir, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	os.MkdirAll(bm.config.StoragePath, 0755)
	srcPath := filepath.Join(dir, "import.json")
	os.WriteFile(srcPath, []byte(`{"id":"b1","policy_id":"`+policy.ID+`","status":"completed"}`), 0644)

	record, err := bm.ImportBackup(ctx, srcPath, policy.ID)
	if err != nil {
		assert.Contains(t, err.Error(), "failed")
	} else {
		assert.NotNil(t, record)
	}
}

func TestBackupManager_ImportBackup_FileNotFound(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	_, err := bm.ImportBackup(ctx, "/nonexistent/file.json", "policy-id")
	assert.Error(t, err)
}

func TestDataArchiver_ExecuteArchiveTask_NoPolicy(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	task := &ArchiveTask{ID: "t1", PolicyID: "nonexistent", Status: ArchiveStatusPending, CreatedAt: time.Now()}
	archiver.db.Create(task)

	archiver.executeArchiveTask(ctx, task)

	var updated ArchiveTask
	archiver.db.First(&updated, "id = ?", "t1")
	assert.Equal(t, ArchiveStatusFailed, updated.Status)
}

func TestDataArchiver_ExecuteArchiveTask_NoData(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{Name: "test", DataType: "archive_policies", Enabled: true, ArchiveAfter: 0, BatchSize: 100, Format: ArchiveFormatJSON, Compression: false}
	archiver.CreatePolicy(ctx, policy)

	task := &ArchiveTask{ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusPending, CreatedAt: time.Now()}
	archiver.db.Create(task)

	archiver.executeArchiveTask(ctx, task)

	var updated ArchiveTask
	archiver.db.First(&updated, "id = ?", "t1")
	assert.NotEqual(t, ArchiveStatusPending, updated.Status)
}

func TestDataArchiver_FailTask(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()

	task := &ArchiveTask{ID: "t1", PolicyID: "p1", Status: ArchiveStatusRunning, CreatedAt: time.Now()}
	archiver.db.Create(task)

	archiver.failTask(task, "test error")

	var updated ArchiveTask
	archiver.db.First(&updated, "id = ?", "t1")
	assert.Equal(t, ArchiveStatusFailed, updated.Status)
	assert.Equal(t, "test error", updated.Error)
}

func TestDataArchiver_VerifyArchive(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	records := []map[string]interface{}{{"id": "1"}, {"id": "2"}}
	filePath, _ := archiver.ArchiveData(ctx, "test_type", records)

	checksum, _ := archiver.CalculateChecksum(filePath)

	task := &ArchiveTask{
		ID:           "t1",
		PolicyID:     "p1",
		Status:       ArchiveStatusValidating,
		FilePath:     filePath,
		Checksum:     checksum,
		RecordsCount: 2,
		CreatedAt:    time.Now(),
	}

	err := archiver.verifyArchive(task)
	assert.NoError(t, err)
}

func TestDataArchiver_VerifyArchive_FileNotFound(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()

	task := &ArchiveTask{
		ID:        "t1",
		FilePath:  "/nonexistent/file.json",
		Checksum:  "abc",
		CreatedAt: time.Now(),
	}

	err := archiver.verifyArchive(task)
	assert.Error(t, err)
}

func TestDataArchiver_CheckScheduledArchives(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: 0, BatchSize: 100, Schedule: "0 0 * * *",
	}
	archiver.CreatePolicy(ctx, policy)

	archiver.checkScheduledArchives(ctx)
}

func TestDataArchiver_UpdateMetrics(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	archiver.updateMetrics(ctx)
}

func TestBackupManager_ExecuteBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Compression: false, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	record := &BackupRecord{
		ID:       "b1",
		PolicyID: policy.ID,
		Type:     BackupTypeFull,
		Status:   BackupStatusPending,
	}
	bm.db.Create(record)

	bm.executeBackup(ctx, record)

	var updated BackupRecord
	bm.db.First(&updated, "id = ?", "b1")
	assert.NotEqual(t, BackupStatusPending, updated.Status)
}

func TestBackupManager_FailBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()

	record := &BackupRecord{ID: "b1", Status: BackupStatusRunning}
	bm.db.Create(record)

	bm.failBackup(record, "test error")

	var updated BackupRecord
	bm.db.First(&updated, "id = ?", "b1")
	assert.Equal(t, BackupStatusFailed, updated.Status)
	assert.Equal(t, "test error", updated.Error)
}

func TestBackupManager_ExecuteRestore(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	restore := &RestoreRecord{
		ID:       "r1",
		BackupID: "b1",
		Status:   BackupStatusPending,
	}
	bm.db.Create(restore)

	bm.executeRestore(ctx, restore)

	var updated RestoreRecord
	bm.db.First(&updated, "id = ?", "r1")
	assert.NotEqual(t, BackupStatusPending, updated.Status)
}

func TestBackupManager_FailRestore(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()

	restore := &RestoreRecord{ID: "r1", BackupID: "b1", Status: BackupStatusRunning}
	bm.db.Create(restore)

	bm.failRestore(restore, "test error")

	var updated RestoreRecord
	bm.db.First(&updated, "id = ?", "r1")
	assert.Equal(t, BackupStatusFailed, updated.Status)
}

func TestBackupManager_CheckScheduledBackups(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Schedule: "0 0 * * *", Enabled: true}
	bm.CreatePolicy(ctx, policy)

	bm.checkScheduledBackups(ctx)
}

func TestBackupManager_CleanupOldBackups(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, MaxBackups: 2, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	bm.cleanupOldBackups(ctx)
}

func TestBackupManager_UpdateMetrics(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.updateMetrics(ctx)
}

func TestDataCleaner_ExecuteCleanupTask(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "archive_policies", RetentionDays: 30, Enabled: true, BatchSize: 100}
	cleaner.CreatePolicy(ctx, policy)

	task := &CleanupTask{ID: "t1", PolicyID: policy.ID, Status: CleanupStatusPending, CreatedAt: time.Now()}
	cleaner.db.Create(task)

	cleaner.executeCleanupTask(ctx, task)

	var updated CleanupTask
	cleaner.db.First(&updated, "id = ?", "t1")
	assert.Equal(t, CleanupStatusCompleted, updated.Status)
}

func TestDataCleaner_ExecuteCleanupTask_NoPolicy(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	task := &CleanupTask{ID: "t1", PolicyID: "nonexistent", Status: CleanupStatusPending, CreatedAt: time.Now()}
	cleaner.db.Create(task)

	cleaner.executeCleanupTask(ctx, task)

	var updated CleanupTask
	cleaner.db.First(&updated, "id = ?", "t1")
	assert.Equal(t, CleanupStatusFailed, updated.Status)
}

func TestDataCleaner_FailTask(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()

	task := &CleanupTask{ID: "t1", PolicyID: "p1", Status: CleanupStatusRunning, CreatedAt: time.Now()}
	cleaner.db.Create(task)

	cleaner.failTask(task, "test error")

	var updated CleanupTask
	cleaner.db.First(&updated, "id = ?", "t1")
	assert.Equal(t, CleanupStatusFailed, updated.Status)
}

func TestDataCleaner_CheckTaskCancelled(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()

	task := &CleanupTask{ID: "t1", PolicyID: "p1", Status: CleanupStatusCancelled, CreatedAt: time.Now()}
	cleaner.db.Create(task)

	cancelled := cleaner.checkTaskCancelled("t1")
	assert.True(t, cancelled)
}

func TestDataCleaner_CheckTaskCancelled_NotCancelled(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()

	task := &CleanupTask{ID: "t1", PolicyID: "p1", Status: CleanupStatusRunning, CreatedAt: time.Now()}
	cleaner.db.Create(task)

	cancelled := cleaner.checkTaskCancelled("t1")
	assert.False(t, cancelled)
}

func TestDataCleaner_LogTask(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()

	cleaner.logTask("task-1", "info", "test message", "rec-1", "")
}

func TestDataCleaner_CheckScheduledCleanups(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "archive_policies", Schedule: "0 0 * * *", Enabled: true}
	cleaner.CreatePolicy(ctx, policy)

	cleaner.checkScheduledCleanups(ctx)
}

func TestDataCleaner_CleanupOldLogs(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	cleaner.cleanupOldLogs(ctx)
}

func TestDataCleaner_UpdateMetrics(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	cleaner.updateMetrics(ctx)
}

func TestDataCleaner_DoCleanup(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "archive_policies", RetentionDays: 30, Enabled: true, BatchSize: 100}
	cleaner.CreatePolicy(ctx, policy)

	task := &CleanupTask{ID: "t1", PolicyID: policy.ID, Status: CleanupStatusRunning, CreatedAt: time.Now()}
	cleaner.db.Create(task)

	err := cleaner.doCleanup(ctx, task, policy)
	assert.NoError(t, err)
}

func TestTieredStorage_PromoteToHot(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "stat-1"}
	ts.Store(ctx, "promote-key", data, 1*time.Hour)

	ts.promoteToHot(ctx, "promote-key", []byte(`{"id":"stat-1"}`))
}

func TestTieredStorage_MigrateKey(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "stat-1"}
	ts.Store(ctx, "migrate-key2", data, 1*time.Hour)

	ts.migrateKey(ctx, "migrate-key2")
}

func TestTieredStorage_UpdateStats(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	ts.updateStats(ctx)
}

func TestTieredStorage_UpdateMetrics(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	ts.updateMetrics(ctx)
}

func TestTieredStorage_GetHotKeys(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "1"}
	ts.Store(ctx, "key1", data, 1*time.Hour)
	ts.Store(ctx, "key2", data, 1*time.Hour)

	keys, err := ts.getHotKeys(ctx)
	assert.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestTieredStorage_CheckAndMigrate(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "1"}
	ts.Store(ctx, "key1", data, 1*time.Hour)

	ts.hotStats.Store("key1", &HotStats{
		Key:        "key1",
		HitCount:   0,
		LastAccess: time.Now().Add(-48 * time.Hour),
		FirstSeen:  time.Now().Add(-48 * time.Hour),
		WindowHits: 0,
		UpdatedAt:  time.Now().Add(-48 * time.Hour),
	})

	ts.checkAndMigrate(ctx)
}

func TestTieredStorage_CheckHotToWarmMigration(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "1"}
	ts.Store(ctx, "key1", data, 1*time.Hour)

	ts.hotStats.Store("key1", &HotStats{
		Key:        "key1",
		HitCount:   0,
		LastAccess: time.Now().Add(-48 * time.Hour),
		FirstSeen:  time.Now().Add(-48 * time.Hour),
		WindowHits: 0,
		UpdatedAt:  time.Now().Add(-48 * time.Hour),
	})

	ts.checkHotToWarmMigration(ctx)
}

func TestTieredStorage_CheckWarmToColdMigration(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	ts.checkWarmToColdMigration(ctx)
}

func TestTieredStorage_MigrateKey_HotToWarm(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "1"}
	ts.Store(ctx, "migrate-key3", data, 1*time.Hour)

	ts.migrateKey(ctx, "migrate-key3")
}

func TestTieredStorage_MigrateKey_NotFound(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	ts.migrateKey(ctx, "nonexistent-key")
}

func TestBackupManager_DoRestore_FileNotFound(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	backup := &BackupRecord{ID: "b1", FilePath: "/nonexistent/file.json", Status: BackupStatusCompleted}
	restore := &RestoreRecord{ID: "r1", BackupID: "b1", Status: BackupStatusRunning}
	bm.db.Create(backup)
	bm.db.Create(restore)

	err := bm.doRestore(ctx, backup, restore)
	assert.Error(t, err)
}

func TestBackupManager_BackupTable(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(bm.config.StoragePath, 0755)
	filePath := filepath.Join(bm.config.StoragePath, "test_backup.json")
	file, err := os.Create(filePath)
	assert.NoError(t, err)
	defer file.Close()

	record := &BackupRecord{ID: "b1", Type: BackupTypeFull}
	records, size, err := bm.backupTable(ctx, "archive_policies", file, record)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, records, int64(0))
	assert.GreaterOrEqual(t, size, int64(0))
}

func TestDataArchiver_CheckScheduledArchives_WithSchedule(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: 0, BatchSize: 100, Format: ArchiveFormatJSON, Compression: false,
	}
	archiver.CreatePolicy(ctx, policy)

	policy2 := &ArchivePolicy{
		Name: "test2", DataType: "archive_tasks", Enabled: true,
		ArchiveAfter: 0, BatchSize: 100, Format: ArchiveFormatJSON, Compression: false,
	}
	archiver.CreatePolicy(ctx, policy2)

	task := &ArchiveTask{ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusPending, CreatedAt: time.Now()}
	archiver.db.Create(task)

	archiver.executeArchiveTask(ctx, task)

	var updated ArchiveTask
	archiver.db.First(&updated, "id = ?", "t1")
	assert.NotEqual(t, ArchiveStatusPending, updated.Status)
}

func TestDataArchiver_DoArchive(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: 0, BatchSize: 100, Format: ArchiveFormatJSON, Compression: false,
	}
	archiver.CreatePolicy(ctx, policy)

	task := &ArchiveTask{ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusRunning, CreatedAt: time.Now()}
	archiver.db.Create(task)

	err := archiver.doArchive(ctx, task, policy)
	if err != nil {
		assert.Contains(t, err.Error(), "failed")
	}
}

func TestDataArchiver_ExecuteArchiveTask_WithData(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: 0, BatchSize: 100, Format: ArchiveFormatJSON, Compression: false,
	}
	archiver.CreatePolicy(ctx, policy)

	task := &ArchiveTask{ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusPending, CreatedAt: time.Now()}
	archiver.db.Create(task)

	archiver.executeArchiveTask(ctx, task)

	var updated ArchiveTask
	archiver.db.First(&updated, "id = ?", "t1")
	assert.NotEqual(t, ArchiveStatusPending, updated.Status)
}

func TestBackupManager_CheckScheduledBackups_WithSchedule(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Schedule: "* * * * *", Enabled: true}
	bm.CreatePolicy(ctx, policy)

	bm.checkScheduledBackups(ctx)
}

func TestDataCleaner_CheckScheduledCleanups_WithSchedule(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "archive_policies", Schedule: "* * * * *", Enabled: true}
	cleaner.CreatePolicy(ctx, policy)

	cleaner.checkScheduledCleanups(ctx)
}

func TestDataCleaner_CleanupOldLogs_WithOldLogs(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	cleaner.db.Create(&CleanupLog{ID: "l1", TaskID: "t1", Level: "info", Message: "old", CreatedAt: time.Now().Add(-100 * 24 * time.Hour)})
	cleaner.db.Create(&CleanupLog{ID: "l2", TaskID: "t1", Level: "info", Message: "new", CreatedAt: time.Now()})

	cleaner.cleanupOldLogs(ctx)
}

func TestBackupManager_CleanupOldBackups_WithOldBackups(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, MaxBackups: 1, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	bm.db.Create(&BackupRecord{ID: "b1", PolicyID: policy.ID, Status: BackupStatusCompleted, CreatedAt: time.Now().Add(-48 * time.Hour)})
	bm.db.Create(&BackupRecord{ID: "b2", PolicyID: policy.ID, Status: BackupStatusCompleted, CreatedAt: time.Now()})

	bm.cleanupOldBackups(ctx)
}

func TestTieredStorage_PromoteToHot_WithData(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "stat-1"}
	ts.Store(ctx, "promote-key", data, 1*time.Hour)

	ts.promoteToHot(ctx, "promote-key", []byte(`{"id":"stat-1"}`))

	var result map[string]interface{}
	err := ts.Get(ctx, "promote-key", &result)
	assert.NoError(t, err)
}

func TestTieredStorage_Migrate_HotToWarm(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "1"}
	ts.Store(ctx, "migrate-hw", data, 1*time.Hour)

	err := ts.Migrate(ctx, "migrate-hw", TierHot, TierWarm)
	if err != nil {
		assert.Contains(t, err.Error(), "NOW")
	}
}

func TestTieredStorage_Migrate_InvalidFromTier(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.Migrate(ctx, "key1", DataTier(99), TierHot)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid source tier")
}

func TestTieredStorage_Migrate_InvalidToTier(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "1"}
	ts.Store(ctx, "migrate-it", data, 1*time.Hour)

	err := ts.Migrate(ctx, "migrate-it", TierHot, DataTier(99))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid target tier")
}

func TestTieredStorage_Migrate_NotFoundInSource(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.Migrate(ctx, "nonexistent", TierWarm, TierCold)
	assert.Error(t, err)
}

func TestBackupManager_ExecuteRestore_NoBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	restore := &RestoreRecord{ID: "r1", BackupID: "nonexistent", Status: BackupStatusPending}
	bm.db.Create(restore)

	bm.executeRestore(ctx, restore)

	var updated RestoreRecord
	bm.db.First(&updated, "id = ?", "r1")
	assert.Equal(t, BackupStatusFailed, updated.Status)
}

func TestBackupManager_ExecuteRestore_WithBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted, FilePath: "/nonexistent/file.json"})
	restore := &RestoreRecord{ID: "r1", BackupID: "b1", Status: BackupStatusPending}
	bm.db.Create(restore)

	bm.executeRestore(ctx, restore)

	var updated RestoreRecord
	bm.db.First(&updated, "id = ?", "r1")
	assert.NotEqual(t, BackupStatusPending, updated.Status)
}

func TestDataArchiver_ExecuteArchiveTask_WithCompression(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: 0, BatchSize: 100, Format: ArchiveFormatJSON, Compression: true,
	}
	archiver.CreatePolicy(ctx, policy)

	task := &ArchiveTask{ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusPending, CreatedAt: time.Now()}
	archiver.db.Create(task)

	archiver.executeArchiveTask(ctx, task)

	var updated ArchiveTask
	archiver.db.First(&updated, "id = ?", "t1")
	assert.NotEqual(t, ArchiveStatusPending, updated.Status)
}

func TestDataArchiver_DoArchive_WithArchiveAfter(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: 1 * time.Hour, BatchSize: 100, Format: ArchiveFormatJSON, Compression: false,
	}
	archiver.CreatePolicy(ctx, policy)

	task := &ArchiveTask{ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusRunning, CreatedAt: time.Now()}
	archiver.db.Create(task)

	err := archiver.doArchive(ctx, task, policy)
	if err != nil {
		assert.Contains(t, err.Error(), "failed")
	}
}

func TestDataCleaner_DoCleanup_WithDryRun(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "archive_policies", RetentionDays: 30, Enabled: true, BatchSize: 100, DryRun: true}
	cleaner.CreatePolicy(ctx, policy)

	task := &CleanupTask{ID: "t1", PolicyID: policy.ID, Status: CleanupStatusRunning, CreatedAt: time.Now()}
	cleaner.db.Create(task)

	err := cleaner.doCleanup(ctx, task, policy)
	assert.NoError(t, err)
}

func TestDataCleaner_DoCleanup_WithArchiveFirst(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "archive_policies", RetentionDays: 30, Enabled: true, BatchSize: 100, ArchiveFirst: true}
	cleaner.CreatePolicy(ctx, policy)

	task := &CleanupTask{ID: "t1", PolicyID: policy.ID, Status: CleanupStatusRunning, CreatedAt: time.Now()}
	cleaner.db.Create(task)

	err := cleaner.doCleanup(ctx, task, policy)
	assert.NoError(t, err)
}

func TestDataCleaner_ExecuteCleanupTask_WithCancelledTask(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "archive_policies", RetentionDays: 30, Enabled: true, BatchSize: 100}
	cleaner.CreatePolicy(ctx, policy)

	task := &CleanupTask{ID: "t1", PolicyID: policy.ID, Status: CleanupStatusCancelled, CreatedAt: time.Now()}
	cleaner.db.Create(task)

	cleaner.executeCleanupTask(ctx, task)

	var updated CleanupTask
	cleaner.db.First(&updated, "id = ?", "t1")
	assert.NotEqual(t, CleanupStatusPending, updated.Status)
}

func TestDataArchiver_VerifyArchive_ChecksumMismatch(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	records := []map[string]interface{}{{"id": "1"}}
	filePath, _ := archiver.ArchiveData(ctx, "test_type", records)

	task := &ArchiveTask{
		ID:           "t1",
		FilePath:     filePath,
		Checksum:     "wrong-checksum",
		RecordsCount: 1,
		CreatedAt:    time.Now(),
	}

	err := archiver.verifyArchive(task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestDataArchiver_VerifyArchive_RecordCountMismatch(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	records := []map[string]interface{}{{"id": "1"}}
	filePath, _ := archiver.ArchiveData(ctx, "test_type", records)

	checksum, _ := archiver.CalculateChecksum(filePath)

	task := &ArchiveTask{
		ID:           "t1",
		FilePath:     filePath,
		Checksum:     checksum,
		RecordsCount: 999,
		CreatedAt:    time.Now(),
	}

	err := archiver.verifyArchive(task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record count mismatch")
}

func TestDataArchiver_CompactArchives_WithPolicy(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{Name: "test", DataType: "archive_policies", Enabled: true, Format: ArchiveFormatJSON}
	archiver.CreatePolicy(ctx, policy)

	os.MkdirAll(archiver.config.StoragePath, 0755)
	records := []map[string]interface{}{{"id": "1"}}
	filePath, _ := archiver.ArchiveData(ctx, "test_type", records)
	checksum, _ := archiver.CalculateChecksum(filePath)

	archiver.db.Create(&ArchiveTask{
		ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusCompleted,
		FilePath: filePath, Checksum: checksum, RecordsCount: 1, DataSize: 100,
		CreatedAt: time.Now(),
	})

	err := archiver.CompactArchives(ctx, policy.ID)
	if err != nil {
		assert.Contains(t, err.Error(), "failed")
	}
}

func TestTieredStorage_Delete_FromWarm(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.deleteWarm(ctx, "nonexistent-warm-key")
	assert.NoError(t, err)
}

func TestTieredStorage_Delete_FromCold(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.deleteCold(ctx, "nonexistent-cold-key")
	assert.NoError(t, err)
}

func TestTieredStorage_Get_FromWarm(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	_, err := ts.getWarm(ctx, "nonexistent-warm-key")
	assert.Error(t, err)
}

func TestTieredStorage_Get_FromCold(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	_, err := ts.getCold(ctx, "nonexistent-cold-key")
	assert.Error(t, err)
}

func TestTieredStorage_CheckColdExists(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	exists, err := ts.checkColdExists(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestDataArchiver_RestoreArchive_WithCompletedTask(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	records := []map[string]interface{}{{"id": "1", "name": "test1"}}
	filePath, _ := archiver.ArchiveData(ctx, "test_type", records)

	task := &ArchiveTask{
		ID: "t1", PolicyID: "p1", Status: ArchiveStatusCompleted,
		FilePath: filePath, RecordsCount: 1,
		CreatedAt: time.Now(),
	}
	archiver.db.Create(task)

	err := archiver.RestoreArchive(ctx, "t1", "archive_policies")
	if err != nil {
		assert.Contains(t, err.Error(), "failed")
	}
}

func TestDataArchiver_RestoreArchive_NotCompletedTask(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	task := &ArchiveTask{
		ID: "t1", PolicyID: "p1", Status: ArchiveStatusPending,
		CreatedAt: time.Now(),
	}
	archiver.db.Create(task)

	err := archiver.RestoreArchive(ctx, "t1", "archive_policies")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not completed")
}

func TestDataArchiver_DoArchive_WithOldData(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: 0, BatchSize: 100, Format: ArchiveFormatJSON, Compression: false,
	}
	archiver.CreatePolicy(ctx, policy)

	archiver.db.Create(&ArchivePolicy{Name: "old-data", DataType: "test", CreatedAt: time.Now().Add(-48 * time.Hour)})

	task := &ArchiveTask{ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusRunning, CreatedAt: time.Now()}
	archiver.db.Create(task)

	err := archiver.doArchive(ctx, task, policy)
	if err != nil {
		assert.Contains(t, err.Error(), "failed")
	}
}

func TestBackupManager_ExecuteBackup_WithPolicy(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Compression: false, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	record := &BackupRecord{
		ID:       "b1",
		PolicyID: policy.ID,
		Type:     BackupTypeFull,
		Status:   BackupStatusPending,
	}
	bm.db.Create(record)

	bm.executeBackup(ctx, record)

	var updated BackupRecord
	bm.db.First(&updated, "id = ?", "b1")
	assert.NotEqual(t, BackupStatusPending, updated.Status)
}

func TestBackupManager_ExecuteBackup_NoPolicy(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	record := &BackupRecord{
		ID:       "b1",
		PolicyID: "nonexistent",
		Type:     BackupTypeFull,
		Status:   BackupStatusPending,
	}
	bm.db.Create(record)

	bm.executeBackup(ctx, record)

	var updated BackupRecord
	bm.db.First(&updated, "id = ?", "b1")
	assert.Equal(t, BackupStatusFailed, updated.Status)
}

func TestBackupManager_DoRestore_WithValidFile(t *testing.T) {
	bm, dir, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(bm.config.StoragePath, 0755)
	filePath := filepath.Join(dir, "backup", "test_restore.json")
	os.MkdirAll(filepath.Dir(filePath), 0755)

	file, err := os.Create(filePath)
	assert.NoError(t, err)
	json.NewEncoder(file).Encode(map[string]interface{}{"metadata": map[string]interface{}{"backup_id": "b1"}})
	json.NewEncoder(file).Encode(map[string]interface{}{"table_start": "archive_policies"})
	json.NewEncoder(file).Encode(map[string]interface{}{"row": map[string]interface{}{"name": "test"}})
	json.NewEncoder(file).Encode(map[string]interface{}{"table_end": true})
	file.Close()

	backup := &BackupRecord{ID: "b1", FilePath: filePath, Status: BackupStatusCompleted}
	restore := &RestoreRecord{ID: "r1", BackupID: "b1", Status: BackupStatusRunning}
	bm.db.Create(backup)
	bm.db.Create(restore)

	err = bm.doRestore(ctx, backup, restore)
	assert.NoError(t, err)
}

func TestDataCleaner_DoCleanup_WithQuery(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{
		Name: "test", DataType: "archive_policies", RetentionDays: 30,
		Enabled: true, BatchSize: 100,
	}
	cleaner.CreatePolicy(ctx, policy)

	task := &CleanupTask{ID: "t1", PolicyID: policy.ID, Status: CleanupStatusRunning, CreatedAt: time.Now()}
	cleaner.db.Create(task)

	err := cleaner.doCleanup(ctx, task, policy)
	assert.NoError(t, err)
}

func TestTieredStorage_UpdateStats_WithKeys(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "1"}
	ts.Store(ctx, "key1", data, 1*time.Hour)
	ts.Store(ctx, "key2", data, 1*time.Hour)

	ts.hotStats.Store("key1", &HotStats{
		Key: "key1", HitCount: 10, WindowHits: 5, UpdatedAt: time.Now(),
	})
	ts.hotStats.Store("key2", &HotStats{
		Key: "key2", HitCount: 20, WindowHits: 8, UpdatedAt: time.Now(),
	})

	ts.updateStats(ctx)

	s1, ok1 := ts.hotStats.Load("key1")
	assert.True(t, ok1)
	assert.Equal(t, int64(0), s1.(*HotStats).WindowHits)
}

func TestDataArchiver_CompactArchives_WithMultipleTasks(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{Name: "test", DataType: "archive_policies", Enabled: true, Format: ArchiveFormatJSON}
	archiver.CreatePolicy(ctx, policy)

	os.MkdirAll(archiver.config.StoragePath, 0755)

	records1 := []map[string]interface{}{{"id": "1"}}
	filePath1, _ := archiver.ArchiveData(ctx, "type1", records1)
	checksum1, _ := archiver.CalculateChecksum(filePath1)

	records2 := []map[string]interface{}{{"id": "2"}}
	filePath2, _ := archiver.ArchiveData(ctx, "type2", records2)
	checksum2, _ := archiver.CalculateChecksum(filePath2)

	archiver.db.Create(&ArchiveTask{
		ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusCompleted,
		FilePath: filePath1, Checksum: checksum1, RecordsCount: 1, DataSize: 100,
		CreatedAt: time.Now().Add(-48 * time.Hour),
	})
	archiver.db.Create(&ArchiveTask{
		ID: "t2", PolicyID: policy.ID, Status: ArchiveStatusCompleted,
		FilePath: filePath2, Checksum: checksum2, RecordsCount: 1, DataSize: 100,
		CreatedAt: time.Now(),
	})

	err := archiver.CompactArchives(ctx, policy.ID)
	if err != nil {
		assert.Contains(t, err.Error(), "failed")
	}
}

func TestDataArchiver_DoArchive_WithData(t *testing.T) {
	archiver, dir, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(filepath.Join(dir, "archive"), 0755)

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: time.Hour, BatchSize: 100, Format: ArchiveFormatJSON, Compression: false,
	}
	archiver.CreatePolicy(ctx, policy)

	archiver.db.Create(&ArchivePolicy{ID: "old-1", Name: "old-data-1", DataType: "test"})
	archiver.db.Create(&ArchivePolicy{ID: "old-2", Name: "old-data-2", DataType: "test"})
	archiver.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id IN (?, ?)", time.Now().Add(-48*time.Hour), "old-1", "old-2")

	task := &ArchiveTask{ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusRunning, CreatedAt: time.Now()}
	archiver.db.Create(task)

	err := archiver.doArchive(ctx, task, policy)
	assert.NoError(t, err)

	var updated ArchiveTask
	archiver.db.First(&updated, "id = ?", "t1")
	if updated.RecordsCount > 0 {
		assert.NotEmpty(t, updated.FilePath)
		assert.NotEmpty(t, updated.Checksum)
	}
}

func TestDataArchiver_DoArchive_WithCompression(t *testing.T) {
	archiver, dir, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(filepath.Join(dir, "archive"), 0755)

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: time.Hour, BatchSize: 100, Format: ArchiveFormatJSON, Compression: true,
	}
	archiver.CreatePolicy(ctx, policy)

	archiver.db.Create(&ArchivePolicy{ID: "old-1", Name: "old-data", DataType: "test"})
	archiver.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-48*time.Hour), "old-1")

	task := &ArchiveTask{ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusRunning, CreatedAt: time.Now()}
	archiver.db.Create(task)

	err := archiver.doArchive(ctx, task, policy)
	assert.NoError(t, err)

	var updated ArchiveTask
	archiver.db.First(&updated, "id = ?", "t1")
	if updated.RecordsCount > 0 && updated.FilePath != "" {
		assert.Contains(t, updated.FilePath, ".gz")
	}
}

func TestDataArchiver_ExecuteArchiveTask_FullFlow(t *testing.T) {
	archiver, dir, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(filepath.Join(dir, "archive"), 0755)

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: time.Hour, BatchSize: 100, Format: ArchiveFormatJSON, Compression: false,
	}
	archiver.CreatePolicy(ctx, policy)

	archiver.db.Create(&ArchivePolicy{ID: "old-1", Name: "data1", DataType: "test"})
	archiver.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-48*time.Hour), "old-1")

	task := &ArchiveTask{ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusPending, CreatedAt: time.Now()}
	archiver.db.Create(task)

	archiver.executeArchiveTask(ctx, task)

	var updated ArchiveTask
	archiver.db.First(&updated, "id = ?", "t1")
	assert.NotEqual(t, ArchiveStatusPending, updated.Status)
}

func TestDataCleaner_DoCleanup_WithOldData(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{
		Name: "test", DataType: "archive_policies", RetentionDays: 30,
		Enabled: true, BatchSize: 100,
	}
	cleaner.CreatePolicy(ctx, policy)

	cleaner.db.Create(&ArchivePolicy{Name: "old-policy", DataType: "test", CreatedAt: time.Now().Add(-100 * 24 * time.Hour)})

	task := &CleanupTask{ID: "t1", PolicyID: policy.ID, Status: CleanupStatusRunning, CreatedAt: time.Now()}
	cleaner.db.Create(task)

	err := cleaner.doCleanup(ctx, task, policy)
	assert.NoError(t, err)

	var updated CleanupTask
	cleaner.db.First(&updated, "id = ?", "t1")
	assert.True(t, updated.RecordsScanned > 0)
	assert.True(t, updated.RecordsDeleted > 0)
}

func TestDataCleaner_ExecuteCleanupTask_WithOldData(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{
		Name: "test", DataType: "archive_policies", RetentionDays: 30,
		Enabled: true, BatchSize: 100,
	}
	cleaner.CreatePolicy(ctx, policy)

	cleaner.db.Create(&ArchivePolicy{Name: "old-policy", DataType: "test", CreatedAt: time.Now().Add(-100 * 24 * time.Hour)})

	task := &CleanupTask{ID: "t1", PolicyID: policy.ID, Status: CleanupStatusPending, CreatedAt: time.Now()}
	cleaner.db.Create(task)

	cleaner.executeCleanupTask(ctx, task)

	var updated CleanupTask
	cleaner.db.First(&updated, "id = ?", "t1")
	assert.Equal(t, CleanupStatusCompleted, updated.Status)
}

func TestBackupManager_ExecuteBackup_WithCompression(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Compression: true, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	record := &BackupRecord{
		ID:       "b1",
		PolicyID: policy.ID,
		Type:     BackupTypeFull,
		Status:   BackupStatusPending,
	}
	bm.db.Create(record)

	bm.executeBackup(ctx, record)

	var updated BackupRecord
	bm.db.First(&updated, "id = ?", "b1")
	assert.NotEqual(t, BackupStatusPending, updated.Status)
}

func TestBackupManager_ExecuteBackup_IncrementType(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeIncrement, Tables: []string{"archive_policies"}, Compression: false, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	record := &BackupRecord{
		ID:            "b1",
		PolicyID:      policy.ID,
		Type:          BackupTypeIncrement,
		Status:        BackupStatusPending,
		BaseBackupID:  "",
	}
	bm.db.Create(record)

	bm.executeBackup(ctx, record)

	var updated BackupRecord
	bm.db.First(&updated, "id = ?", "b1")
	assert.NotEqual(t, BackupStatusPending, updated.Status)
}

func TestTieredStorage_StoreAndGet_MultipleKeys(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		data := map[string]interface{}{"id": fmt.Sprintf("stat-%d", i)}
		err := ts.Store(ctx, fmt.Sprintf("key-%d", i), data, 1*time.Hour)
		assert.NoError(t, err)
	}

	for i := 0; i < 10; i++ {
		var result map[string]interface{}
		err := ts.Get(ctx, fmt.Sprintf("key-%d", i), &result)
		assert.NoError(t, err)
	}
}

func TestTieredStorage_Delete_MultipleKeys(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		data := map[string]interface{}{"id": fmt.Sprintf("stat-%d", i)}
		ts.Store(ctx, fmt.Sprintf("del-key-%d", i), data, 1*time.Hour)
	}

	for i := 0; i < 5; i++ {
		err := ts.Delete(ctx, fmt.Sprintf("del-key-%d", i))
		assert.NoError(t, err)
	}
}

func TestDataArchiver_ArchiveData_LargeRecords(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)

	records := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		records[i] = map[string]interface{}{
			"id":    fmt.Sprintf("%d", i),
			"name":  fmt.Sprintf("test-%d", i),
			"value": i * 10,
		}
	}

	filePath, err := archiver.ArchiveData(ctx, "test_type", records)
	assert.NoError(t, err)
	assert.FileExists(t, filePath)

	readRecords, err := archiver.ReadArchive(ctx, filePath)
	assert.NoError(t, err)
	assert.Len(t, readRecords, 100)
}
