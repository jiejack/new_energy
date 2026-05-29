package lifecycle

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCovTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&CleanupPolicy{}, &CleanupTask{}, &CleanupLog{},
		&ArchivePolicy{}, &ArchiveTask{}, &ArchiveRecord{},
		&BackupPolicy{}, &BackupRecord{}, &BackupTableRecord{}, &RestoreRecord{},
		&WarmData{}, &ColdData{},
	))
	return db
}

func newCovTestArchiver(t *testing.T) (*DataArchiver, string, func()) {
	db := setupCovTestDB(t)
	dir := t.TempDir()
	cfg := DefaultArchiveConfig()
	cfg.StoragePath = filepath.Join(dir, "archive")
	os.MkdirAll(cfg.StoragePath, 0755)
	cfg.TempPath = filepath.Join(dir, "tmp")
	os.MkdirAll(cfg.TempPath, 0755)
	archiver := NewDataArchiver(cfg, db, zap.L().Named("test-archiver"))
	cleanup := func() { sqlDB, _ := db.DB(); sqlDB.Close() }
	return archiver, dir, cleanup
}

func newCovTestBackupManager(t *testing.T) (*BackupManager, string, func()) {
	db := setupCovTestDB(t)
	dir := t.TempDir()
	cfg := DefaultBackupConfig()
	cfg.StoragePath = filepath.Join(dir, "backup")
	os.MkdirAll(cfg.StoragePath, 0755)
	cfg.TempPath = filepath.Join(dir, "tmp")
	os.MkdirAll(cfg.TempPath, 0755)
	cfg.EnableVerify = false
	bm := NewBackupManager(cfg, db, zap.L().Named("test-backup"))
	cleanup := func() { sqlDB, _ := db.DB(); sqlDB.Close() }
	return bm, dir, cleanup
}

func newCovTestCleaner(t *testing.T) (*DataCleaner, func()) {
	db := setupCovTestDB(t)
	cfg := DefaultCleanupConfig()
	cfg.MaxRetryCount = 1
	cfg.RetryDelay = 0
	cleaner := NewDataCleaner(cfg, db, zap.L().Named("test-cleaner"))
	cleanup := func() { sqlDB, _ := db.DB(); sqlDB.Close() }
	return cleaner, cleanup
}

func newCovTestTieredStorage(t *testing.T) (*TieredStorage, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	db := setupCovTestDB(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cfg := TierConfig{HotTTL: time.Hour, WarmTTL: 24 * time.Hour, MigrationInterval: time.Minute, StatsInterval: time.Minute}
	ts := NewTieredStorage(cfg, rdb, db, zap.L().Named("test-tiered"))
	cleanup := func() {
		mr.Close()
		rdb.Close()
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}
	return ts, cleanup
}

func covSha256Sum(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func TestCov_Archiver_CreatePolicy(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		Name: "test-policy", Description: "test", DataType: "archive_policies",
		RetentionDays: 30, ArchiveAfter: 7 * 24 * time.Hour,
		Format: ArchiveFormatJSON, Compression: true, BatchSize: 100, Enabled: true,
	}
	err := da.CreatePolicy(ctx, policy)
	require.NoError(t, err)
	assert.NotEmpty(t, policy.ID)

	policy2 := &ArchivePolicy{
		ID: "explicit-id", Name: "test-policy-2", DataType: "archive_tasks",
		RetentionDays: 60, ArchiveAfter: 14 * 24 * time.Hour,
		Format: ArchiveFormatCSV, Enabled: false,
	}
	err = da.CreatePolicy(ctx, policy2)
	require.NoError(t, err)
	assert.Equal(t, "explicit-id", policy2.ID)
}

func TestCov_Archiver_GetPolicy(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{ID: "get-test-id", Name: "get-test", DataType: "archive_policies", Enabled: true}
	da.CreatePolicy(ctx, policy)

	got, err := da.GetPolicy(ctx, "get-test-id")
	require.NoError(t, err)
	assert.Equal(t, "get-test", got.Name)

	_, err = da.GetPolicy(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestCov_Archiver_UpdatePolicy(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{ID: "upd-test-id", Name: "before", DataType: "archive_policies", Enabled: true}
	da.CreatePolicy(ctx, policy)

	policy.Name = "after"
	err := da.UpdatePolicy(ctx, policy)
	require.NoError(t, err)

	got, _ := da.GetPolicy(ctx, "upd-test-id")
	assert.Equal(t, "after", got.Name)
}

func TestCov_Archiver_DeletePolicy(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{ID: "del-test-id", Name: "del-test", DataType: "archive_policies", Enabled: true}
	da.CreatePolicy(ctx, policy)

	err := da.DeletePolicy(ctx, "del-test-id")
	require.NoError(t, err)

	_, err = da.GetPolicy(ctx, "del-test-id")
	assert.Error(t, err)
}

func TestCov_Archiver_ListTasks(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{ID: "lt-policy", Name: "lt-test", DataType: "archive_policies", Enabled: true}
	da.CreatePolicy(ctx, policy)

	da.db.Create(&ArchiveTask{ID: "lt-task-1", PolicyID: "lt-policy", Status: ArchiveStatusPending, CreatedAt: time.Now()})
	da.db.Create(&ArchiveTask{ID: "lt-task-2", PolicyID: "lt-policy", Status: ArchiveStatusCompleted, CreatedAt: time.Now()})
	da.db.Create(&ArchiveTask{ID: "lt-task-3", PolicyID: "other", Status: ArchiveStatusPending, CreatedAt: time.Now()})

	tasks, err := da.ListTasks(ctx, "lt-policy", 0)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)

	tasks, err = da.ListTasks(ctx, "", 2)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(tasks), 2)

	tasks, err = da.ListTasks(ctx, "", 0)
	require.NoError(t, err)
	assert.Len(t, tasks, 3)
}

func TestCov_Archiver_GetArchiveStats(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{ID: "stats-p1", Name: "stats", DataType: "archive_policies", Enabled: true}
	da.CreatePolicy(ctx, policy)

	now := time.Now()
	da.db.Create(&ArchiveTask{
		ID: "stats-t1", PolicyID: "stats-p1", Status: ArchiveStatusCompleted,
		RecordsCount: 100, DataSize: 10240, FilePath: "/tmp/f1", CreatedAt: now, StartTime: &now, EndTime: &now,
	})
	da.db.Create(&ArchiveTask{
		ID: "stats-t2", PolicyID: "stats-p1", Status: ArchiveStatusCompleted,
		RecordsCount: 200, DataSize: 20480, FilePath: "/tmp/f2", CreatedAt: now, StartTime: &now, EndTime: &now,
	})
	da.db.Create(&ArchiveTask{
		ID: "stats-t3", PolicyID: "other", Status: ArchiveStatusCompleted,
		RecordsCount: 50, DataSize: 5120, FilePath: "/tmp/f3", CreatedAt: now, StartTime: &now, EndTime: &now,
	})

	stats, err := da.GetArchiveStats(ctx, "stats-p1")
	require.NoError(t, err)
	assert.NotNil(t, stats)

	stats, err = da.GetArchiveStats(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestCov_Archiver_RestoreArchive(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	records := []map[string]interface{}{
		{"id": "restore-1", "name": "r1"},
		{"id": "restore-2", "name": "r2"},
	}
	filePath, err := da.ArchiveData(ctx, "archive_policies", records)
	require.NoError(t, err)

	now := time.Now()
	task := &ArchiveTask{
		ID: "restore-task", PolicyID: "p1", Status: ArchiveStatusCompleted,
		RecordsCount: 2, FilePath: filePath, CreatedAt: now, StartTime: &now, EndTime: &now,
	}
	da.db.Create(task)

	err = da.RestoreArchive(ctx, "restore-task", "archive_policies")
	assert.NoError(t, err)

	nonCompletedTask := &ArchiveTask{
		ID: "restore-task-2", PolicyID: "p1", Status: ArchiveStatusPending,
		RecordsCount: 0, FilePath: filePath, CreatedAt: now,
	}
	da.db.Create(nonCompletedTask)
	err = da.RestoreArchive(ctx, "restore-task-2", "archive_policies")
	assert.Error(t, err)

	_, err = da.GetTask(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestCov_Archiver_RestoreArchive_Compressed(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "test_restore.json.gz")
	file, err := os.Create(filePath)
	require.NoError(t, err)
	gzWriter := gzip.NewWriter(file)
	encoder := json.NewEncoder(gzWriter)
	encoder.Encode(map[string]interface{}{"id": "gz-1", "name": "compressed"})
	encoder.Encode(map[string]interface{}{"id": "gz-2", "name": "compressed2"})
	gzWriter.Close()
	file.Close()

	now := time.Now()
	task := &ArchiveTask{
		ID: "gz-restore-task", PolicyID: "p1", Status: ArchiveStatusCompleted,
		RecordsCount: 2, FilePath: filePath, CreatedAt: now, StartTime: &now, EndTime: &now,
	}
	da.db.Create(task)

	err = da.RestoreArchive(ctx, "gz-restore-task", "archive_policies")
	assert.NoError(t, err)
}

func TestCov_Archiver_CompactArchives(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		ID: "compact-p1", Name: "compact", DataType: "archive_policies",
		Format: ArchiveFormatJSON, Compression: false, Enabled: true,
	}
	da.CreatePolicy(ctx, policy)

	records1 := []map[string]interface{}{{"id": "c1", "name": "n1"}}
	records2 := []map[string]interface{}{{"id": "c2", "name": "n2"}}
	fp1, _ := da.ArchiveData(ctx, "archive_policies", records1)
	fp2, _ := da.ArchiveData(ctx, "archive_policies", records2)

	now := time.Now()
	da.db.Create(&ArchiveTask{
		ID: "compact-t1", PolicyID: "compact-p1", Status: ArchiveStatusCompleted,
		RecordsCount: 1, DataSize: 100, FilePath: fp1, CreatedAt: now, StartTime: &now, EndTime: &now,
	})
	da.db.Create(&ArchiveTask{
		ID: "compact-t2", PolicyID: "compact-p1", Status: ArchiveStatusCompleted,
		RecordsCount: 1, DataSize: 100, FilePath: fp2, CreatedAt: now, StartTime: &now, EndTime: &now,
	})

	err := da.CompactArchives(ctx, "compact-p1")
	require.NoError(t, err)

	var tasks []ArchiveTask
	da.db.Where("policy_id = ?", "compact-p1").Find(&tasks)
	assert.Len(t, tasks, 1)
}

func TestCov_Archiver_CompactArchives_NotEnough(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		ID: "compact-p2", Name: "compact2", DataType: "archive_policies",
		Format: ArchiveFormatJSON, Compression: false, Enabled: true,
	}
	da.CreatePolicy(ctx, policy)

	now := time.Now()
	da.db.Create(&ArchiveTask{
		ID: "compact-t3", PolicyID: "compact-p2", Status: ArchiveStatusCompleted,
		RecordsCount: 1, DataSize: 100, FilePath: "/tmp/single", CreatedAt: now, StartTime: &now, EndTime: &now,
	})

	err := da.CompactArchives(ctx, "compact-p2")
	assert.NoError(t, err)
}

func TestCov_Archiver_CompactArchives_Compressed(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		ID: "compact-p3", Name: "compact3", DataType: "archive_policies",
		Format: ArchiveFormatJSON, Compression: true, Enabled: true,
	}
	da.CreatePolicy(ctx, policy)

	dir := da.config.StoragePath
	f1 := filepath.Join(dir, "f1.json.gz")
	f2 := filepath.Join(dir, "f2.json.gz")
	for _, fp := range []string{f1, f2} {
		file, _ := os.Create(fp)
		gzWriter := gzip.NewWriter(file)
		encoder := json.NewEncoder(gzWriter)
		encoder.Encode(map[string]interface{}{"id": "1", "name": "test"})
		gzWriter.Close()
		file.Close()
	}

	now := time.Now()
	da.db.Create(&ArchiveTask{
		ID: "compact-gz1", PolicyID: "compact-p3", Status: ArchiveStatusCompleted,
		RecordsCount: 1, DataSize: 50, FilePath: f1, CreatedAt: now, StartTime: &now, EndTime: &now,
	})
	da.db.Create(&ArchiveTask{
		ID: "compact-gz2", PolicyID: "compact-p3", Status: ArchiveStatusCompleted,
		RecordsCount: 1, DataSize: 50, FilePath: f2, CreatedAt: now, StartTime: &now, EndTime: &now,
	})

	err := da.CompactArchives(ctx, "compact-p3")
	require.NoError(t, err)
}

func TestCov_Archiver_doArchive(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		ID: "doarch-p1", Name: "doarch", DataType: "archive_policies",
		ArchiveAfter: 1 * time.Hour, Format: ArchiveFormatJSON,
		Compression: false, BatchSize: 100, Enabled: true,
	}
	da.CreatePolicy(ctx, policy)

	da.db.Create(&ArchivePolicy{ID: "old-rec-1", Name: "old", DataType: "test", Enabled: true})
	da.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-48*time.Hour), "old-rec-1")

	task := &ArchiveTask{
		ID: "doarch-task-1", PolicyID: "doarch-p1", Status: ArchiveStatusRunning,
		CreatedAt: time.Now(),
	}
	da.db.Create(task)

	err := da.doArchive(ctx, task, policy)
	require.NoError(t, err)
	assert.Greater(t, task.RecordsCount, int64(0))
	assert.NotEmpty(t, task.FilePath)
	assert.NotEmpty(t, task.Checksum)
}

func TestCov_Archiver_doArchive_Compressed(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		ID: "doarch-p2", Name: "doarch-gz", DataType: "archive_policies",
		ArchiveAfter: 1 * time.Hour, Format: ArchiveFormatJSON,
		Compression: true, BatchSize: 100, Enabled: true,
	}
	da.CreatePolicy(ctx, policy)

	da.db.Create(&ArchivePolicy{ID: "old-rec-2", Name: "old2", DataType: "test", Enabled: true})
	da.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-48*time.Hour), "old-rec-2")

	task := &ArchiveTask{
		ID: "doarch-task-2", PolicyID: "doarch-p2", Status: ArchiveStatusRunning,
		CreatedAt: time.Now(),
	}
	da.db.Create(task)

	err := da.doArchive(ctx, task, policy)
	require.NoError(t, err)
	assert.NotEmpty(t, task.FilePath)
}

func TestCov_Archiver_doArchive_NoData(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		ID: "doarch-p3", Name: "doarch-nodata", DataType: "archive_policies",
		ArchiveAfter: 1 * time.Hour, Format: ArchiveFormatJSON,
		Compression: false, BatchSize: 100, Enabled: true,
	}
	da.CreatePolicy(ctx, policy)

	task := &ArchiveTask{
		ID: "doarch-task-3", PolicyID: "doarch-p3", Status: ArchiveStatusRunning,
		CreatedAt: time.Now(),
	}
	da.db.Create(task)

	err := da.doArchive(ctx, task, policy)
	assert.NoError(t, err)
}

func TestCov_Archiver_verifyArchive(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		ID: "verify-p1", Name: "verify", DataType: "archive_policies",
		ArchiveAfter: 1 * time.Hour, Format: ArchiveFormatJSON,
		Compression: false, BatchSize: 100, Enabled: true,
	}
	da.CreatePolicy(ctx, policy)

	da.db.Create(&ArchivePolicy{ID: "verify-rec-1", Name: "v1", DataType: "test", Enabled: true})
	da.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-48*time.Hour), "verify-rec-1")

	task := &ArchiveTask{
		ID: "verify-task-1", PolicyID: "verify-p1", Status: ArchiveStatusRunning,
		CreatedAt: time.Now(),
	}
	da.db.Create(task)

	da.doArchive(ctx, task, policy)

	err := da.verifyArchive(task)
	assert.NoError(t, err)
}

func TestCov_Archiver_verifyArchive_Compressed(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		ID: "verify-p2", Name: "verify-gz", DataType: "archive_policies",
		ArchiveAfter: 1 * time.Hour, Format: ArchiveFormatJSON,
		Compression: true, BatchSize: 100, Enabled: true,
	}
	da.CreatePolicy(ctx, policy)

	da.db.Create(&ArchivePolicy{ID: "verify-rec-2", Name: "v2", DataType: "test", Enabled: true})
	da.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-48*time.Hour), "verify-rec-2")

	task := &ArchiveTask{
		ID: "verify-task-2", PolicyID: "verify-p2", Status: ArchiveStatusRunning,
		CreatedAt: time.Now(),
	}
	da.db.Create(task)

	da.doArchive(ctx, task, policy)

	err := da.verifyArchive(task)
	assert.NoError(t, err)
}

func TestCov_Archiver_verifyArchive_FileNotFound(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()

	task := &ArchiveTask{
		ID: "verify-task-nf", PolicyID: "p1", Status: ArchiveStatusRunning,
		FilePath: "/nonexistent/file.json", RecordsCount: 1, CreatedAt: time.Now(),
	}
	err := da.verifyArchive(task)
	assert.Error(t, err)
}

func TestCov_Archiver_executeArchiveTask(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		ID: "exec-p1", Name: "exec", DataType: "archive_policies",
		ArchiveAfter: 1 * time.Hour, Format: ArchiveFormatJSON,
		Compression: false, BatchSize: 100, Enabled: true,
	}
	da.CreatePolicy(ctx, policy)

	da.db.Create(&ArchivePolicy{ID: "exec-rec-1", Name: "old", DataType: "test", Enabled: true})
	da.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-48*time.Hour), "exec-rec-1")

	task := &ArchiveTask{
		ID: "exec-task-1", PolicyID: "exec-p1", Status: ArchiveStatusPending,
		CreatedAt: time.Now(),
	}
	da.db.Create(task)

	da.executeArchiveTask(ctx, task)

	var updated ArchiveTask
	da.db.First(&updated, "id = ?", "exec-task-1")
	assert.Equal(t, ArchiveStatusCompleted, updated.Status)
}

func TestCov_Archiver_executeArchiveTask_PolicyNotFound(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	task := &ArchiveTask{
		ID: "exec-task-nf", PolicyID: "nonexistent", Status: ArchiveStatusPending,
		CreatedAt: time.Now(),
	}
	da.db.Create(task)

	da.executeArchiveTask(ctx, task)

	var updated ArchiveTask
	da.db.First(&updated, "id = ?", "exec-task-nf")
	assert.Equal(t, ArchiveStatusFailed, updated.Status)
}

func TestCov_Archiver_checkScheduledArchives(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		ID: "sched-p1", Name: "sched", DataType: "archive_policies",
		ArchiveAfter: 1 * time.Hour, Format: ArchiveFormatJSON,
		Compression: false, BatchSize: 100, Enabled: true,
	}
	da.CreatePolicy(ctx, policy)

	da.db.Create(&ArchivePolicy{ID: "sched-rec-1", Name: "old", DataType: "test", Enabled: true})
	da.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-48*time.Hour), "sched-rec-1")

	da.checkScheduledArchives(ctx)

	var tasks []ArchiveTask
	da.db.Where("policy_id = ?", "sched-p1").Find(&tasks)
	assert.GreaterOrEqual(t, len(tasks), 1)
}

func TestCov_Archiver_TriggerArchive(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		ID: "trigger-p1", Name: "trigger", DataType: "archive_policies",
		ArchiveAfter: 1 * time.Hour, Format: ArchiveFormatJSON,
		Compression: false, BatchSize: 100, Enabled: true,
	}
	da.CreatePolicy(ctx, policy)

	task, err := da.TriggerArchive(ctx, "trigger-p1")
	require.NoError(t, err)
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, ArchiveStatusPending, task.Status)

	disabledPolicy := &ArchivePolicy{
		ID: "trigger-p2", Name: "disabled", DataType: "archive_policies",
		ArchiveAfter: 1 * time.Hour, Format: ArchiveFormatJSON,
		Compression: false, BatchSize: 100, Enabled: false,
	}
	da.CreatePolicy(ctx, disabledPolicy)

	_, err = da.TriggerArchive(ctx, "trigger-p2")
	assert.Error(t, err)

	_, err = da.TriggerArchive(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestCov_Archiver_StartStop(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := da.Start(ctx)
	require.NoError(t, err)

	err = da.Stop()
	require.NoError(t, err)
}

func TestCov_Archiver_ValidateArchiveFile_Gz(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()

	dir := t.TempDir()
	fp := filepath.Join(dir, "validate.json.gz")
	file, _ := os.Create(fp)
	gzWriter := gzip.NewWriter(file)
	encoder := json.NewEncoder(gzWriter)
	encoder.Encode(map[string]interface{}{"id": "1", "name": "test"})
	gzWriter.Close()
	file.Close()

	err := da.ValidateArchiveFile(fp)
	assert.NoError(t, err)
}

func TestCov_Archiver_CalculateChecksum_Gz(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()

	dir := t.TempDir()
	fp := filepath.Join(dir, "checksum.json.gz")
	file, _ := os.Create(fp)
	gzWriter := gzip.NewWriter(file)
	gzWriter.Write([]byte(`{"key":"value"}`))
	gzWriter.Close()
	file.Close()

	checksum, err := da.CalculateChecksum(fp)
	require.NoError(t, err)
	assert.NotEmpty(t, checksum)
}

func TestCov_Archiver_ReadArchive_Gz(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	dir := t.TempDir()
	fp := filepath.Join(dir, "read.json.gz")
	file, _ := os.Create(fp)
	gzWriter := gzip.NewWriter(file)
	encoder := json.NewEncoder(gzWriter)
	encoder.Encode(map[string]interface{}{"id": "1", "value": 10})
	encoder.Encode(map[string]interface{}{"id": "2", "value": 20})
	gzWriter.Close()
	file.Close()

	records, err := da.ReadArchive(ctx, fp)
	require.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestCov_Archiver_StreamArchive_Gz(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	dir := t.TempDir()
	fp := filepath.Join(dir, "stream.json.gz")
	file, _ := os.Create(fp)
	gzWriter := gzip.NewWriter(file)
	encoder := json.NewEncoder(gzWriter)
	encoder.Encode(map[string]interface{}{"id": "1"})
	gzWriter.Close()
	file.Close()

	count := 0
	err := da.StreamArchive(ctx, fp, func(record map[string]interface{}) error {
		count++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCov_Archiver_ArchiveBuffer_Concurrent(t *testing.T) {
	ab := NewArchiveBuffer()
	done := make(chan bool, 2)
	go func() {
		for i := 0; i < 100; i++ {
			ab.Write([]byte("a"))
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 100; i++ {
			ab.Len()
		}
		done <- true
	}()
	<-done
	<-done
	assert.Equal(t, 100, ab.Len())
}

func TestCov_Backup_CreatePolicy(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{
		Name: "test-bp", Type: BackupTypeFull,
		Tables: []string{"archive_policies"}, Enabled: true,
	}
	err := bm.CreatePolicy(ctx, policy)
	require.NoError(t, err)
	assert.NotEmpty(t, policy.ID)
	assert.Greater(t, policy.RetentionDays, 0)
	assert.Greater(t, policy.MaxBackups, 0)

	policy2 := &BackupPolicy{
		ID: "explicit-bp-id", Name: "test-bp-2", Type: BackupTypeIncrement,
		Tables: []string{"archive_policies"}, Enabled: true,
	}
	err = bm.CreatePolicy(ctx, policy2)
	require.NoError(t, err)
	assert.Equal(t, "explicit-bp-id", policy2.ID)
}

func TestCov_Backup_GetPolicy(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{ID: "get-bp-id", Name: "get-test", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	got, err := bm.GetPolicy(ctx, "get-bp-id")
	require.NoError(t, err)
	assert.Equal(t, "get-test", got.Name)

	_, err = bm.GetPolicy(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestCov_Backup_UpdatePolicy(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{ID: "upd-bp-id", Name: "before", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	policy.Name = "after"
	err := bm.UpdatePolicy(ctx, policy)
	require.NoError(t, err)

	got, _ := bm.GetPolicy(ctx, "upd-bp-id")
	assert.Equal(t, "after", got.Name)
}

func TestCov_Backup_DeletePolicy(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{ID: "del-bp-id", Name: "del-test", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	err := bm.DeletePolicy(ctx, "del-bp-id")
	require.NoError(t, err)

	_, err = bm.GetPolicy(ctx, "del-bp-id")
	assert.Error(t, err)
}

func TestCov_Backup_TriggerBackup(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{ID: "trig-bp-id", Name: "trig-test", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	record, err := bm.TriggerBackup(ctx, "trig-bp-id")
	require.NoError(t, err)
	assert.NotEmpty(t, record.ID)
	assert.Equal(t, BackupStatusPending, record.Status)

	disabledPolicy := &BackupPolicy{ID: "trig-bp-dis", Name: "disabled", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Enabled: false}
	bm.CreatePolicy(ctx, disabledPolicy)

	_, err = bm.TriggerBackup(ctx, "trig-bp-dis")
	assert.Error(t, err)
}

func TestCov_Backup_TriggerBackup_IncrementNoBase(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{ID: "inc-bp-id", Name: "inc-test", Type: BackupTypeIncrement, Tables: []string{"archive_policies"}, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	record, err := bm.TriggerBackup(ctx, "inc-bp-id")
	require.NoError(t, err)
	assert.Equal(t, BackupTypeFull, record.Type)
}

func TestCov_Backup_CreateIncrementBackup(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	fullPolicy := &BackupPolicy{ID: "inc-err-bp", Name: "full-only", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Enabled: true}
	bm.CreatePolicy(ctx, fullPolicy)

	_, err := bm.CreateIncrementBackup(ctx, "inc-err-bp")
	assert.Error(t, err)

	incPolicy := &BackupPolicy{ID: "inc-ok-bp", Name: "inc-ok", Type: BackupTypeIncrement, Tables: []string{"archive_policies"}, Enabled: true}
	bm.CreatePolicy(ctx, incPolicy)

	record, err := bm.CreateIncrementBackup(ctx, "inc-ok-bp")
	require.NoError(t, err)
	assert.NotNil(t, record)
}

func TestCov_Backup_CreateFullBackup(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	record, err := bm.CreateFullBackup(ctx, []string{"archive_policies"}, "manual full")
	require.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, BackupTypeFull, record.Type)
}

func TestCov_Backup_ListBackups(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{ID: "lb-bp-id", Name: "lb-test", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	now := time.Now()
	bm.db.Create(&BackupRecord{ID: "lb-r1", PolicyID: "lb-bp-id", Type: BackupTypeFull, Status: BackupStatusCompleted, CreatedAt: now})
	bm.db.Create(&BackupRecord{ID: "lb-r2", PolicyID: "lb-bp-id", Type: BackupTypeFull, Status: BackupStatusCompleted, CreatedAt: now})
	bm.db.Create(&BackupRecord{ID: "lb-r3", PolicyID: "other", Type: BackupTypeFull, Status: BackupStatusCompleted, CreatedAt: now})

	records, err := bm.ListBackups(ctx, "lb-bp-id", 0)
	require.NoError(t, err)
	assert.Len(t, records, 2)

	records, err = bm.ListBackups(ctx, "", 2)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(records), 2)
}

func TestCov_Backup_ListRestores(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	now := time.Now()
	bm.db.Create(&RestoreRecord{ID: "lr-r1", BackupID: "b1", Status: BackupStatusCompleted, CreatedAt: now})
	bm.db.Create(&RestoreRecord{ID: "lr-r2", BackupID: "b1", Status: BackupStatusCompleted, CreatedAt: now})
	bm.db.Create(&RestoreRecord{ID: "lr-r3", BackupID: "b2", Status: BackupStatusCompleted, CreatedAt: now})

	records, err := bm.ListRestores(ctx, "b1", 0)
	require.NoError(t, err)
	assert.Len(t, records, 2)

	records, err = bm.ListRestores(ctx, "", 1)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(records), 1)
}

func TestCov_Backup_DeleteBackup(t *testing.T) {
	bm, dir, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	fp := filepath.Join(dir, "backup", "del-backup.json")
	os.WriteFile(fp, []byte(`{"test":true}`), 0644)

	bm.db.Create(&BackupRecord{ID: "del-r1", PolicyID: "p1", Type: BackupTypeFull, Status: BackupStatusCompleted, FilePath: fp, CreatedAt: time.Now()})
	bm.db.Create(&BackupTableRecord{ID: "del-btr1", BackupID: "del-r1", TableName: "test", CreatedAt: time.Now()})

	err := bm.DeleteBackup(ctx, "del-r1")
	require.NoError(t, err)

	var count int64
	bm.db.Model(&BackupRecord{}).Where("id = ?", "del-r1").Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestCov_Backup_DeleteBackup_NoFile(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "del-r2", PolicyID: "p1", Type: BackupTypeFull, Status: BackupStatusCompleted, FilePath: "/nonexistent/file.json", CreatedAt: time.Now()})

	err := bm.DeleteBackup(ctx, "del-r2")
	require.NoError(t, err)
}

func TestCov_Backup_VerifyBackup(t *testing.T) {
	bm, dir, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	fp := filepath.Join(dir, "backup", "verify-backup.json")
	data := []byte(`{"metadata":{"backup_id":"test"}}`)
	os.WriteFile(fp, data, 0644)

	hash := covSha256Sum(data)

	bm.db.Create(&BackupRecord{
		ID: "verify-r1", PolicyID: "p1", Type: BackupTypeFull,
		Status: BackupStatusCompleted, FilePath: fp, Checksum: hash, CreatedAt: time.Now(),
	})

	err := bm.VerifyBackup(ctx, "verify-r1")
	assert.NoError(t, err)

	bm.db.Create(&BackupRecord{
		ID: "verify-r2", PolicyID: "p1", Type: BackupTypeFull,
		Status: BackupStatusPending, FilePath: fp, CreatedAt: time.Now(),
	})
	err = bm.VerifyBackup(ctx, "verify-r2")
	assert.Error(t, err)
}

func TestCov_Backup_VerifyBackup_ChecksumMismatch(t *testing.T) {
	bm, dir, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	fp := filepath.Join(dir, "backup", "mismatch-backup.json")
	os.WriteFile(fp, []byte(`{"data":1}`), 0644)

	bm.db.Create(&BackupRecord{
		ID: "mismatch-r1", PolicyID: "p1", Type: BackupTypeFull,
		Status: BackupStatusCompleted, FilePath: fp, Checksum: "wrongchecksum", CreatedAt: time.Now(),
	})

	err := bm.VerifyBackup(ctx, "mismatch-r1")
	assert.Error(t, err)
}

func TestCov_Backup_VerifyBackup_FileNotFound(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{
		ID: "fnf-r1", PolicyID: "p1", Type: BackupTypeFull,
		Status: BackupStatusCompleted, FilePath: "/nonexistent/file.json", CreatedAt: time.Now(),
	})

	err := bm.VerifyBackup(ctx, "fnf-r1")
	assert.Error(t, err)
}

func TestCov_Backup_VerifyBackup_Gz(t *testing.T) {
	bm, dir, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	fp := filepath.Join(dir, "backup", "verify-backup.json.gz")
	plainData := []byte(`{"metadata":{"backup_id":"test"}}`)
	file, _ := os.Create(fp)
	gzWriter := gzip.NewWriter(file)
	gzWriter.Write(plainData)
	gzWriter.Close()
	file.Close()

	hash := covSha256Sum(plainData)

	bm.db.Create(&BackupRecord{
		ID: "verify-gz-r1", PolicyID: "p1", Type: BackupTypeFull,
		Status: BackupStatusCompleted, FilePath: fp, Checksum: hash, CreatedAt: time.Now(),
	})

	err := bm.VerifyBackup(ctx, "verify-gz-r1")
	assert.NoError(t, err)
}

func TestCov_Backup_GetBackupStats(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{ID: "stats-bp", Name: "stats", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	now := time.Now()
	bm.db.Create(&BackupRecord{
		ID: "stats-r1", PolicyID: "stats-bp", Type: BackupTypeFull,
		Status: BackupStatusCompleted, Size: 1024, CompressedSize: 512, CreatedAt: now,
	})
	bm.db.Create(&BackupRecord{
		ID: "stats-r2", PolicyID: "stats-bp", Type: BackupTypeFull,
		Status: BackupStatusCompleted, Size: 2048, CompressedSize: 1024, CreatedAt: now,
	})

	stats, err := bm.GetBackupStats(ctx, "stats-bp")
	require.NoError(t, err)
	assert.NotNil(t, stats)

	stats, err = bm.GetBackupStats(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestCov_Backup_ExportBackup(t *testing.T) {
	bm, dir, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	srcDir := filepath.Join(dir, "backup")
	fp := filepath.Join(srcDir, "export-src.json")
	os.WriteFile(fp, []byte(`{"export":"data"}`), 0644)

	bm.db.Create(&BackupRecord{
		ID: "export-r1", PolicyID: "p1", Type: BackupTypeFull,
		Status: BackupStatusCompleted, FilePath: fp, CreatedAt: time.Now(),
	})

	destPath := filepath.Join(dir, "exported.json")
	err := bm.ExportBackup(ctx, "export-r1", destPath)
	require.NoError(t, err)
	assert.FileExists(t, destPath)

	bm.db.Create(&BackupRecord{
		ID: "export-r2", PolicyID: "p1", Type: BackupTypeFull,
		Status: BackupStatusPending, FilePath: fp, CreatedAt: time.Now(),
	})
	err = bm.ExportBackup(ctx, "export-r2", filepath.Join(dir, "exported2.json"))
	assert.Error(t, err)
}

func TestCov_Backup_ImportBackup(t *testing.T) {
	bm, dir, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	srcPath := filepath.Join(dir, "import-src.json")
	os.WriteFile(srcPath, []byte(`{"import":"data"}`), 0644)

	record, err := bm.ImportBackup(ctx, srcPath, "some-policy")
	require.NoError(t, err)
	assert.NotEmpty(t, record.ID)
	assert.NotEmpty(t, record.Checksum)
	assert.Equal(t, BackupTypeFull, record.Type)
}

func TestCov_Backup_CompareBackups(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	now := time.Now()
	bm.db.Create(&BackupRecord{
		ID: "cmp-r1", PolicyID: "p1", Type: BackupTypeFull,
		Status: BackupStatusCompleted, Size: 1024, RecordsCount: 100, CreatedAt: now,
	})
	bm.db.Create(&BackupRecord{
		ID: "cmp-r2", PolicyID: "p1", Type: BackupTypeFull,
		Status: BackupStatusCompleted, Size: 2048, RecordsCount: 200, CreatedAt: now.Add(time.Hour),
	})

	result, err := bm.CompareBackups(ctx, "cmp-r1", "cmp-r2")
	require.NoError(t, err)
	assert.NotNil(t, result)

	_, err = bm.CompareBackups(ctx, "cmp-r1", "nonexistent")
	assert.Error(t, err)
}

func TestCov_Backup_ScheduleBackup(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{ID: "sched-bp", Name: "sched", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	err := bm.ScheduleBackup(ctx, "sched-bp", time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	err = bm.ScheduleBackup(ctx, "nonexistent", time.Now())
	assert.Error(t, err)
}

func TestCov_Backup_CancelScheduledBackup(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "cancel-r1", PolicyID: "p1", Type: BackupTypeFull, Status: BackupStatusPending, CreatedAt: time.Now()})

	err := bm.CancelScheduledBackup(ctx, "cancel-r1")
	require.NoError(t, err)

	bm.db.Create(&BackupRecord{ID: "cancel-r2", PolicyID: "p1", Type: BackupTypeFull, Status: BackupStatusCompleted, CreatedAt: time.Now()})
	err = bm.CancelScheduledBackup(ctx, "cancel-r2")
	assert.Error(t, err)

	err = bm.CancelScheduledBackup(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestCov_Backup_GetBackupTables(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupTableRecord{ID: "bt-r1", BackupID: "b1", TableName: "table1", RecordsCount: 10, Size: 100, CreatedAt: time.Now()})
	bm.db.Create(&BackupTableRecord{ID: "bt-r2", BackupID: "b1", TableName: "table2", RecordsCount: 20, Size: 200, CreatedAt: time.Now()})

	tables, err := bm.GetBackupTables(ctx, "b1")
	require.NoError(t, err)
	assert.Len(t, tables, 2)
}

func TestCov_Backup_doBackup(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{
		ID: "dobk-p1", Name: "dobk", Type: BackupTypeFull,
		Tables: []string{"archive_policies"}, Compression: true, Enabled: true,
	}
	bm.CreatePolicy(ctx, policy)

	bm.db.Create(&ArchivePolicy{ID: "dobk-rec-1", Name: "test", DataType: "test", Enabled: true})

	record := &BackupRecord{
		ID: "dobk-r1", PolicyID: "dobk-p1", Type: BackupTypeFull,
		Status: BackupStatusRunning, CreatedAt: time.Now(),
	}
	bm.db.Create(record)

	err := bm.doBackup(ctx, record, policy)
	require.NoError(t, err)
	assert.NotEmpty(t, record.FilePath)
	assert.NotEmpty(t, record.Checksum)
	assert.Greater(t, record.TablesCount, 0)
}

func TestCov_Backup_doBackup_NoCompression(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{
		ID: "dobk-p2", Name: "dobk-nocomp", Type: BackupTypeFull,
		Tables: []string{"archive_policies"}, Compression: false, Enabled: true,
	}
	bm.CreatePolicy(ctx, policy)

	bm.db.Create(&ArchivePolicy{ID: "dobk-rec-2", Name: "test2", DataType: "test", Enabled: true})

	record := &BackupRecord{
		ID: "dobk-r2", PolicyID: "dobk-p2", Type: BackupTypeFull,
		Status: BackupStatusRunning, CreatedAt: time.Now(),
	}
	bm.db.Create(record)

	err := bm.doBackup(ctx, record, policy)
	require.NoError(t, err)
	assert.NotEmpty(t, record.FilePath)
}

func TestCov_Backup_backupTable(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&ArchivePolicy{ID: "bt-rec-1", Name: "test", DataType: "test", Enabled: true})
	bm.db.Create(&ArchivePolicy{ID: "bt-rec-2", Name: "test2", DataType: "test", Enabled: true})

	dir := t.TempDir()
	fp := filepath.Join(dir, "table_backup.json")
	file, _ := os.Create(fp)
	defer file.Close()

	record := &BackupRecord{ID: "bt-r1", PolicyID: "p1", Type: BackupTypeFull, CreatedAt: time.Now()}

	totalRecords, totalSize, err := bm.backupTable(ctx, "archive_policies", file, record)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, totalRecords, int64(2))
	assert.GreaterOrEqual(t, totalSize, int64(0))
}

func TestCov_Backup_executeBackup(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{
		ID: "execbk-p1", Name: "execbk", Type: BackupTypeFull,
		Tables: []string{"archive_policies"}, Compression: false, Enabled: true,
	}
	bm.CreatePolicy(ctx, policy)

	bm.db.Create(&ArchivePolicy{ID: "execbk-rec-1", Name: "test", DataType: "test", Enabled: true})

	record := &BackupRecord{
		ID: "execbk-r1", PolicyID: "execbk-p1", Type: BackupTypeFull,
		Status: BackupStatusPending, CreatedAt: time.Now(),
	}
	bm.db.Create(record)

	bm.executeBackup(ctx, record)

	var updated BackupRecord
	bm.db.First(&updated, "id = ?", "execbk-r1")
	assert.Equal(t, BackupStatusCompleted, updated.Status)
}

func TestCov_Backup_executeBackup_PolicyNotFound(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	record := &BackupRecord{
		ID: "execbk-r2", PolicyID: "nonexistent", Type: BackupTypeFull,
		Status: BackupStatusPending, CreatedAt: time.Now(),
	}
	bm.db.Create(record)

	bm.executeBackup(ctx, record)

	var updated BackupRecord
	bm.db.First(&updated, "id = ?", "execbk-r2")
	assert.Equal(t, BackupStatusFailed, updated.Status)
}

func TestCov_Backup_doRestore(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{
		ID: "restore-p1", Name: "restore", Type: BackupTypeFull,
		Tables: []string{"archive_policies"}, Compression: false, Enabled: true,
	}
	bm.CreatePolicy(ctx, policy)

	bm.db.Create(&ArchivePolicy{ID: "restore-src-1", Name: "src", DataType: "test", Enabled: true})

	backupRecord := &BackupRecord{
		ID: "restore-br1", PolicyID: "restore-p1", Type: BackupTypeFull,
		Status: BackupStatusRunning, CreatedAt: time.Now(),
	}
	bm.db.Create(backupRecord)
	bm.doBackup(ctx, backupRecord, policy)
	backupRecord.Status = BackupStatusCompleted
	bm.db.Save(backupRecord)

	restoreRecord := &RestoreRecord{
		ID: "restore-rr1", BackupID: "restore-br1", Status: BackupStatusPending,
		CreatedAt: time.Now(),
	}
	bm.db.Create(restoreRecord)

	err := bm.doRestore(ctx, backupRecord, restoreRecord)
	assert.NoError(t, err)
}

func TestCov_Backup_doRestore_Compressed(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{
		ID: "restore-p2", Name: "restore-gz", Type: BackupTypeFull,
		Tables: []string{"archive_policies"}, Compression: true, Enabled: true,
	}
	bm.CreatePolicy(ctx, policy)

	bm.db.Create(&ArchivePolicy{ID: "restore-src-2", Name: "src2", DataType: "test", Enabled: true})

	backupRecord := &BackupRecord{
		ID: "restore-br2", PolicyID: "restore-p2", Type: BackupTypeFull,
		Status: BackupStatusRunning, CreatedAt: time.Now(),
	}
	bm.db.Create(backupRecord)
	bm.doBackup(ctx, backupRecord, policy)
	backupRecord.Status = BackupStatusCompleted
	bm.db.Save(backupRecord)

	restoreRecord := &RestoreRecord{
		ID: "restore-rr2", BackupID: "restore-br2", Status: BackupStatusPending,
		CreatedAt: time.Now(),
	}
	bm.db.Create(restoreRecord)

	err := bm.doRestore(ctx, backupRecord, restoreRecord)
	assert.NoError(t, err)
}

func TestCov_Backup_RestoreBackup(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{
		ID: "rbk-r1", PolicyID: "p1", Type: BackupTypeFull,
		Status: BackupStatusCompleted, FilePath: "/tmp/test", CreatedAt: time.Now(),
	})

	restoreRecord, err := bm.RestoreBackup(ctx, "rbk-r1", []string{"archive_policies"})
	require.NoError(t, err)
	assert.NotEmpty(t, restoreRecord.ID)

	bm.db.Create(&BackupRecord{
		ID: "rbk-r2", PolicyID: "p1", Type: BackupTypeFull,
		Status: BackupStatusPending, FilePath: "/tmp/test", CreatedAt: time.Now(),
	})
	_, err = bm.RestoreBackup(ctx, "rbk-r2", nil)
	assert.Error(t, err)
}

func TestCov_Backup_GetRestore(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&RestoreRecord{ID: "gr-r1", BackupID: "b1", Status: BackupStatusCompleted, CreatedAt: time.Now()})

	record, err := bm.GetRestore(ctx, "gr-r1")
	require.NoError(t, err)
	assert.Equal(t, "b1", record.BackupID)

	_, err = bm.GetRestore(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestCov_Backup_checkScheduledBackups(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{ID: "schedbk-p1", Name: "schedbk", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	bm.checkScheduledBackups(ctx)

	var records []BackupRecord
	bm.db.Where("policy_id = ?", "schedbk-p1").Find(&records)
	assert.GreaterOrEqual(t, len(records), 1)
}

func TestCov_Backup_checkScheduledBackups_RecentBackup(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{ID: "schedbk-p2", Name: "schedbk-recent", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	bm.db.Create(&BackupRecord{
		ID: "schedbk-r1", PolicyID: "schedbk-p2", Type: BackupTypeFull,
		Status: BackupStatusCompleted, CreatedAt: time.Now(),
	})

	bm.checkScheduledBackups(ctx)
}

func TestCov_Backup_cleanupOldBackups(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{
		ID: "cleanbk-p1", Name: "cleanbk", Type: BackupTypeFull,
		Tables: []string{"archive_policies"}, Enabled: true,
		RetentionDays: 1, MaxBackups: 1,
	}
	bm.CreatePolicy(ctx, policy)

	now := time.Now()
	oldTime := now.Add(-30 * 24 * time.Hour)
	bm.db.Create(&BackupRecord{ID: "cleanbk-r1", PolicyID: "cleanbk-p1", Type: BackupTypeFull, Status: BackupStatusCompleted, FilePath: "/tmp/old", CreatedAt: oldTime})
	bm.db.Create(&BackupRecord{ID: "cleanbk-r2", PolicyID: "cleanbk-p1", Type: BackupTypeFull, Status: BackupStatusCompleted, FilePath: "/tmp/new", CreatedAt: now})
	bm.db.Create(&BackupRecord{ID: "cleanbk-r3", PolicyID: "cleanbk-p1", Type: BackupTypeFull, Status: BackupStatusCompleted, FilePath: "/tmp/new2", CreatedAt: now})

	bm.cleanupOldBackups(ctx)

	var remaining []BackupRecord
	bm.db.Where("policy_id = ?", "cleanbk-p1").Find(&remaining)
	assert.LessOrEqual(t, len(remaining), 1)
}

func TestCov_Backup_StartStop(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bm.Start(ctx)
	require.NoError(t, err)

	err = bm.Stop()
	require.NoError(t, err)
}

func TestCov_Backup_GetMetrics(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()

	metrics := bm.GetMetrics()
	require.NotNil(t, metrics)
}

func TestCov_Cleaner_CreatePolicy(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test-cp", DataType: "archive_policies", Enabled: true}
	err := dc.CreatePolicy(ctx, policy)
	require.NoError(t, err)
	assert.NotEmpty(t, policy.ID)
	assert.Greater(t, policy.BatchSize, 0)
	assert.Greater(t, policy.RetentionDays, 0)

	policy2 := &CleanupPolicy{ID: "explicit-cp-id", Name: "test-cp-2", DataType: "archive_tasks", Enabled: true}
	err = dc.CreatePolicy(ctx, policy2)
	require.NoError(t, err)
	assert.Equal(t, "explicit-cp-id", policy2.ID)
}

func TestCov_Cleaner_GetPolicy(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{ID: "get-cp-id", Name: "get-test", DataType: "archive_policies", Enabled: true}
	dc.CreatePolicy(ctx, policy)

	got, err := dc.GetPolicy(ctx, "get-cp-id")
	require.NoError(t, err)
	assert.Equal(t, "get-test", got.Name)

	_, err = dc.GetPolicy(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestCov_Cleaner_UpdatePolicy(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{ID: "upd-cp-id", Name: "before", DataType: "archive_policies", Enabled: true}
	dc.CreatePolicy(ctx, policy)

	policy.Name = "after"
	err := dc.UpdatePolicy(ctx, policy)
	require.NoError(t, err)

	got, _ := dc.GetPolicy(ctx, "upd-cp-id")
	assert.Equal(t, "after", got.Name)
}

func TestCov_Cleaner_DeletePolicy(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{ID: "del-cp-id", Name: "del-test", DataType: "archive_policies", Enabled: true}
	dc.CreatePolicy(ctx, policy)

	err := dc.DeletePolicy(ctx, "del-cp-id")
	require.NoError(t, err)

	_, err = dc.GetPolicy(ctx, "del-cp-id")
	assert.Error(t, err)
}

func TestCov_Cleaner_TriggerCleanup(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{ID: "trig-cp-id", Name: "trig-test", DataType: "archive_policies", Enabled: true}
	dc.CreatePolicy(ctx, policy)

	task, err := dc.TriggerCleanup(ctx, "trig-cp-id")
	require.NoError(t, err)
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, CleanupStatusPending, task.Status)

	disabledPolicy := &CleanupPolicy{ID: "trig-cp-dis", Name: "disabled", DataType: "archive_policies", Enabled: false}
	dc.CreatePolicy(ctx, disabledPolicy)

	_, err = dc.TriggerCleanup(ctx, "trig-cp-dis")
	assert.Error(t, err)
}

func TestCov_Cleaner_CancelTask(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	dc.db.Create(&CleanupTask{ID: "cancel-ct-1", PolicyID: "p1", Status: CleanupStatusPending, CreatedAt: time.Now()})

	err := dc.CancelTask(ctx, "cancel-ct-1")
	require.NoError(t, err)

	dc.db.Create(&CleanupTask{ID: "cancel-ct-2", PolicyID: "p1", Status: CleanupStatusCompleted, CreatedAt: time.Now()})
	err = dc.CancelTask(ctx, "cancel-ct-2")
	assert.Error(t, err)

	err = dc.CancelTask(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestCov_Cleaner_PreviewCleanup(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{ID: "preview-cp", Name: "preview", DataType: "archive_policies", RetentionDays: 30, Enabled: true}
	dc.CreatePolicy(ctx, policy)

	dc.db.Create(&ArchivePolicy{ID: "preview-old-1", Name: "old", DataType: "test", Enabled: true})
	dc.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-100*24*time.Hour), "preview-old-1")

	count, err := dc.PreviewCleanup(ctx, "preview-cp")
	require.NoError(t, err)
	assert.Greater(t, count, int64(0))

	_, err = dc.PreviewCleanup(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestCov_Cleaner_EstimateCleanupSize(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{ID: "est-cp", Name: "estimate", DataType: "archive_policies", RetentionDays: 30, Enabled: true}
	dc.CreatePolicy(ctx, policy)

	dc.db.Create(&ArchivePolicy{ID: "est-old-1", Name: "old", DataType: "test", Enabled: true})
	dc.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-100*24*time.Hour), "est-old-1")

	result, err := dc.EstimateCleanupSize(ctx, "est-cp")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "record_count")
}

func TestCov_Cleaner_CleanupByQuery(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	dc.db.Create(&ArchivePolicy{ID: "cq-1", Name: "del-me", DataType: "test", Enabled: true})
	dc.db.Create(&ArchivePolicy{ID: "cq-2", Name: "keep-me", DataType: "test", Enabled: true})

	deleted, err := dc.CleanupByQuery(ctx, "archive_policies", "name = ?", "del-me")
	require.NoError(t, err)
	assert.Greater(t, deleted, int64(0))
}

func TestCov_Cleaner_CleanupByDate(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	dc.db.Create(&ArchivePolicy{ID: "cbd-1", Name: "old", DataType: "test", Enabled: true})
	dc.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-100*24*time.Hour), "cbd-1")

	dc.db.Create(&ArchivePolicy{ID: "cbd-2", Name: "new", DataType: "test", Enabled: true})

	deleted, err := dc.CleanupByDate(ctx, "archive_policies", "created_at", time.Now().Add(-50*24*time.Hour))
	require.NoError(t, err)
	assert.Greater(t, deleted, int64(0))
}

func TestCov_Cleaner_CleanupBatch(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	dc.db.Create(&ArchivePolicy{ID: "cb-1", Name: "del1", DataType: "test", Enabled: true})
	dc.db.Create(&ArchivePolicy{ID: "cb-2", Name: "del2", DataType: "test", Enabled: true})

	deleted, err := dc.CleanupBatch(ctx, "archive_policies", []string{"cb-1", "cb-2"})
	require.NoError(t, err)
	assert.Greater(t, deleted, int64(0))

	deleted, err = dc.CleanupBatch(ctx, "archive_policies", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(0), deleted)
}

func TestCov_Cleaner_CleanupOrphanedRecords(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	dc.db.Create(&ArchivePolicy{ID: "orphan-parent", Name: "parent", DataType: "test", Enabled: true})

	_, err := dc.CleanupOrphanedRecords(ctx, "archive_records", "archive_policies", "task_id")
	require.NoError(t, err)
}

func TestCov_Cleaner_CleanupDuplicates(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := dc.CleanupDuplicates(ctx, "archive_policies", []string{}, true)
	assert.Error(t, err)

	_, err = dc.CleanupDuplicates(ctx, "archive_policies", []string{"name"}, true)
	assert.Error(t, err)
}

func TestCov_Cleaner_GetStorageStats(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	dc.db.Create(&ArchivePolicy{ID: "ss-1", Name: "stat1", DataType: "test", Enabled: true})
	dc.db.Create(&ArchivePolicy{ID: "ss-2", Name: "stat2", DataType: "test", Enabled: true})

	stats, err := dc.GetStorageStats(ctx, "archive_policies")
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "total_count")
	assert.Contains(t, stats, "last_24h")
}

func TestCov_Cleaner_VacuumTable(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	err := dc.VacuumTable(ctx, "invalid; table")
	assert.Error(t, err)

	err = dc.VacuumTable(ctx, "archive_policies")
	assert.Error(t, err)
}

func TestCov_Cleaner_ReindexTable(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	err := dc.ReindexTable(ctx, "invalid; table")
	assert.Error(t, err)

	err = dc.ReindexTable(ctx, "archive_policies")
	assert.Error(t, err)
}

func TestCov_Cleaner_AnalyzeTable(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	err := dc.AnalyzeTable(ctx, "invalid; table")
	assert.Error(t, err)

	err = dc.AnalyzeTable(ctx, "archive_policies")
	assert.NoError(t, err)
}

func TestCov_Cleaner_GetTableSize(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := dc.GetTableSize(ctx, "archive_policies")
	assert.Error(t, err)
}

func TestCov_Cleaner_ListTasks(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	dc.db.Create(&CleanupTask{ID: "lt-ct-1", PolicyID: "p1", Status: CleanupStatusPending, CreatedAt: time.Now()})
	dc.db.Create(&CleanupTask{ID: "lt-ct-2", PolicyID: "p1", Status: CleanupStatusCompleted, CreatedAt: time.Now()})
	dc.db.Create(&CleanupTask{ID: "lt-ct-3", PolicyID: "p2", Status: CleanupStatusPending, CreatedAt: time.Now()})

	tasks, err := dc.ListTasks(ctx, "p1", 0)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)

	tasks, err = dc.ListTasks(ctx, "", 2)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(tasks), 2)

	tasks, err = dc.ListTasks(ctx, "", 0)
	require.NoError(t, err)
	assert.Len(t, tasks, 3)
}

func TestCov_Cleaner_GetCleanupLogs(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	dc.db.Create(&CleanupLog{ID: "gcl-1", TaskID: "t1", Level: "info", Message: "test log 1", CreatedAt: time.Now()})
	dc.db.Create(&CleanupLog{ID: "gcl-2", TaskID: "t1", Level: "warn", Message: "test log 2", CreatedAt: time.Now()})
	dc.db.Create(&CleanupLog{ID: "gcl-3", TaskID: "t2", Level: "info", Message: "test log 3", CreatedAt: time.Now()})

	logs, err := dc.GetCleanupLogs(ctx, "t1", 0)
	require.NoError(t, err)
	assert.Len(t, logs, 2)

	logs, err = dc.GetCleanupLogs(ctx, "t1", 1)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(logs), 1)
}

func TestCov_Cleaner_doCleanup_Real(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{
		ID: "docl-p1", Name: "docl-real", DataType: "archive_policies",
		RetentionDays: 30, BatchSize: 100, Enabled: true,
	}
	dc.CreatePolicy(ctx, policy)

	dc.db.Create(&ArchivePolicy{ID: "docl-old-1", Name: "old", DataType: "test", Enabled: true})
	dc.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-100*24*time.Hour), "docl-old-1")

	task := &CleanupTask{
		ID: "docl-task-1", PolicyID: "docl-p1", Status: CleanupStatusRunning,
		DryRun: false, CreatedAt: time.Now(),
	}
	dc.db.Create(task)

	err := dc.doCleanup(ctx, task, policy)
	require.NoError(t, err)
	assert.Greater(t, task.RecordsScanned, int64(0))
	assert.Greater(t, task.RecordsDeleted, int64(0))
}

func TestCov_Cleaner_doCleanup_DryRun(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{
		ID: "docl-p2", Name: "docl-dry", DataType: "archive_policies",
		RetentionDays: 30, BatchSize: 100, Enabled: true, DryRun: true,
	}
	dc.CreatePolicy(ctx, policy)

	dc.db.Create(&ArchivePolicy{ID: "docl-old-2", Name: "old2", DataType: "test", Enabled: true})
	dc.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-100*24*time.Hour), "docl-old-2")

	task := &CleanupTask{
		ID: "docl-task-2", PolicyID: "docl-p2", Status: CleanupStatusRunning,
		DryRun: true, CreatedAt: time.Now(),
	}
	dc.db.Create(task)

	err := dc.doCleanup(ctx, task, policy)
	require.NoError(t, err)
	assert.Greater(t, task.RecordsScanned, int64(0))

	var remaining []ArchivePolicy
	dc.db.Where("id = ?", "docl-old-2").Find(&remaining)
	assert.NotEmpty(t, remaining)
}

func TestCov_Cleaner_doCleanup_Cancelled(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{
		ID: "docl-p3", Name: "docl-cancel", DataType: "archive_policies",
		RetentionDays: 30, BatchSize: 100, Enabled: true,
	}
	dc.CreatePolicy(ctx, policy)

	dc.db.Create(&ArchivePolicy{ID: "docl-old-3", Name: "old3", DataType: "test", Enabled: true})
	dc.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-100*24*time.Hour), "docl-old-3")

	task := &CleanupTask{
		ID: "docl-task-3", PolicyID: "docl-p3", Status: CleanupStatusCancelled,
		DryRun: false, CreatedAt: time.Now(),
	}
	dc.db.Create(task)

	dc.lastCancelCheck = time.Now().Add(-10 * time.Second)

	err := dc.doCleanup(ctx, task, policy)
	assert.Error(t, err)
}

func TestCov_Cleaner_checkScheduledCleanups(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{
		ID: "schedcl-p1", Name: "schedcl", DataType: "archive_policies",
		RetentionDays: 30, BatchSize: 100, Enabled: true,
	}
	dc.CreatePolicy(ctx, policy)

	dc.db.Create(&ArchivePolicy{ID: "schedcl-old-1", Name: "old", DataType: "test", Enabled: true})
	dc.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-100*24*time.Hour), "schedcl-old-1")

	dc.checkScheduledCleanups(ctx)

	var tasks []CleanupTask
	dc.db.Where("policy_id = ?", "schedcl-p1").Find(&tasks)
	assert.GreaterOrEqual(t, len(tasks), 1)
}

func TestCov_Cleaner_checkScheduledCleanups_RunningTask(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{
		ID: "schedcl-p2", Name: "schedcl-running", DataType: "archive_policies",
		RetentionDays: 30, BatchSize: 100, Enabled: true,
	}
	dc.CreatePolicy(ctx, policy)

	dc.db.Create(&CleanupTask{ID: "schedcl-rt-1", PolicyID: "schedcl-p2", Status: CleanupStatusRunning, CreatedAt: time.Now()})

	dc.checkScheduledCleanups(ctx)
}

func TestCov_Cleaner_executeCleanupTask(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{
		ID: "execcl-p1", Name: "execcl", DataType: "archive_policies",
		RetentionDays: 30, BatchSize: 100, Enabled: true,
	}
	dc.CreatePolicy(ctx, policy)

	dc.db.Create(&ArchivePolicy{ID: "execcl-old-1", Name: "old", DataType: "test", Enabled: true})
	dc.db.Exec("UPDATE archive_policies SET created_at = ? WHERE id = ?", time.Now().Add(-100*24*time.Hour), "execcl-old-1")

	task := &CleanupTask{
		ID: "execcl-task-1", PolicyID: "execcl-p1", Status: CleanupStatusPending,
		CreatedAt: time.Now(),
	}
	dc.db.Create(task)

	dc.executeCleanupTask(ctx, task)

	var updated CleanupTask
	dc.db.First(&updated, "id = ?", "execcl-task-1")
	assert.Equal(t, CleanupStatusCompleted, updated.Status)
}

func TestCov_Cleaner_executeCleanupTask_PolicyNotFound(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	task := &CleanupTask{
		ID: "execcl-task-nf", PolicyID: "nonexistent", Status: CleanupStatusPending,
		CreatedAt: time.Now(),
	}
	dc.db.Create(task)

	dc.executeCleanupTask(ctx, task)

	var updated CleanupTask
	dc.db.First(&updated, "id = ?", "execcl-task-nf")
	assert.Equal(t, CleanupStatusFailed, updated.Status)
}

func TestCov_Cleaner_StartStop(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := dc.Start(ctx)
	require.NoError(t, err)

	err = dc.Stop()
	require.NoError(t, err)
}

func TestCov_Cleaner_GetMetrics(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()

	metrics := dc.GetMetrics()
	require.NotNil(t, metrics)
}

func TestCov_Cleaner_ListPolicies(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	dc.CreatePolicy(ctx, &CleanupPolicy{Name: "lp-1", DataType: "archive_policies", Enabled: true})
	dc.CreatePolicy(ctx, &CleanupPolicy{Name: "lp-2", DataType: "archive_tasks", Enabled: true})

	policies, err := dc.ListPolicies(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(policies), 2)
}

func TestCov_Tiered_Store(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.Store(ctx, "test-key", map[string]interface{}{"value": 42}, time.Hour)
	require.NoError(t, err)
}

func TestCov_Tiered_Get_Hot(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"value": 42}
	err := ts.Store(ctx, "hot-key", data, time.Hour)
	require.NoError(t, err)

	var result map[string]interface{}
	err = ts.Get(ctx, "hot-key", &result)
	require.NoError(t, err)
}

func TestCov_Tiered_Get_Warm(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	jsonData, _ := json.Marshal(map[string]interface{}{"value": 99})
	ts.db.Exec("INSERT INTO warm_data (key, data, created_at, updated_at) VALUES (?, ?, ?, ?)",
		"warm-key", jsonData, time.Now(), time.Now())

	var result map[string]interface{}
	err := ts.Get(ctx, "warm-key", &result)
	require.NoError(t, err)
}

func TestCov_Tiered_Get_Cold(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	jsonData, _ := json.Marshal(map[string]interface{}{"value": 77})
	ts.db.Exec("INSERT INTO cold_data (key, data, created_at) VALUES (?, ?, ?)",
		"cold-key", jsonData, time.Now())

	var result map[string]interface{}
	err := ts.Get(ctx, "cold-key", &result)
	require.NoError(t, err)
}

func TestCov_Tiered_Get_Miss(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	var result map[string]interface{}
	err := ts.Get(ctx, "nonexistent-key", &result)
	assert.Error(t, err)
}

func TestCov_Tiered_GetTier(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.Store(ctx, "tier-hot", map[string]interface{}{"v": 1}, time.Hour)
	require.NoError(t, err)

	tier, err := ts.GetTier(ctx, "tier-hot")
	require.NoError(t, err)
	assert.Equal(t, TierHot, tier)

	jsonData, _ := json.Marshal(map[string]interface{}{"v": 2})
	ts.db.Exec("INSERT INTO warm_data (key, data, created_at, updated_at) VALUES (?, ?, ?, ?)",
		"tier-warm", jsonData, time.Now(), time.Now())

	tier, err = ts.GetTier(ctx, "tier-warm")
	require.NoError(t, err)
	assert.Equal(t, TierWarm, tier)

	ts.db.Exec("INSERT INTO cold_data (key, data, created_at) VALUES (?, ?, ?)",
		"tier-cold", jsonData, time.Now())

	tier, err = ts.GetTier(ctx, "tier-cold")
	require.NoError(t, err)
	assert.Equal(t, TierCold, tier)

	_, err = ts.GetTier(ctx, "nonexistent-tier")
	assert.Error(t, err)
}

func TestCov_Tiered_Delete(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.Store(ctx, "del-key", map[string]interface{}{"v": 1}, time.Hour)
	require.NoError(t, err)

	err = ts.Delete(ctx, "del-key")
	require.NoError(t, err)

	var result map[string]interface{}
	err = ts.Get(ctx, "del-key", &result)
	assert.Error(t, err)
}

func TestCov_Tiered_Migrate_InvalidTiers(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.Store(ctx, "migrate-key", map[string]interface{}{"v": 1}, time.Hour)
	require.NoError(t, err)

	err = ts.Migrate(ctx, "migrate-key", DataTier(99), TierHot)
	assert.Error(t, err)

	err = ts.Migrate(ctx, "migrate-key", TierHot, DataTier(99))
	assert.Error(t, err)

	err = ts.Migrate(ctx, "nonexistent-migrate", TierHot, TierWarm)
	assert.Error(t, err)
}

func TestCov_Tiered_Migrate_HotToWarm(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.Store(ctx, "hot-warm-key", map[string]interface{}{"v": 1}, time.Hour)
	require.NoError(t, err)

	err = ts.Migrate(ctx, "hot-warm-key", TierHot, TierWarm)
	assert.Error(t, err)
}

func TestCov_Tiered_Migrate_HotToCold(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.Store(ctx, "hot-cold-key", map[string]interface{}{"v": 1}, time.Hour)
	require.NoError(t, err)

	err = ts.Migrate(ctx, "hot-cold-key", TierHot, TierCold)
	assert.Error(t, err)
}

func TestCov_Tiered_GetMetrics(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()

	metrics := ts.GetMetrics()
	require.NotNil(t, metrics)
}

func TestCov_Tiered_recordHit(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()

	ts.initHotStats("hit-key")
	ts.recordHit("hit-key", TierHot)

	ts.metrics.mu.Lock()
	hr := ts.metrics.HitRate
	ts.metrics.mu.Unlock()
	assert.Greater(t, hr, float64(0))

	ts.recordHit("nonexistent-key", TierHot)
}

func TestCov_Tiered_recordMiss(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()

	ts.recordMiss("miss-key")

	ts.metrics.mu.Lock()
	mr := ts.metrics.MissRate
	ts.metrics.mu.Unlock()
	assert.Greater(t, mr, float64(0))
}

func TestCov_Tiered_promoteToHot(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data, _ := json.Marshal(map[string]interface{}{"v": "promoted"})
	ts.promoteToHot(ctx, "promote-key", data)

	val, err := ts.redis.Get(ctx, "promote-key").Bytes()
	require.NoError(t, err)
	assert.NotEmpty(t, val)
}

func TestCov_Tiered_StartStop(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := ts.Start(ctx)
	require.NoError(t, err)

	err = ts.Stop()
	require.NoError(t, err)
}

func TestCov_Tiered_AutoMigrate(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()

	err := ts.AutoMigrate()
	require.NoError(t, err)
}

func TestCov_Archiver_ListPolicies(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	da.CreatePolicy(ctx, &ArchivePolicy{ID: "lp-a1", Name: "a1", DataType: "archive_policies", Enabled: true})
	da.CreatePolicy(ctx, &ArchivePolicy{ID: "lp-a2", Name: "a2", DataType: "archive_tasks", Enabled: true})

	policies, err := da.ListPolicies(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(policies), 2)
}

func TestCov_Backup_ListPolicies(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.CreatePolicy(ctx, &BackupPolicy{ID: "lp-b1", Name: "b1", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Enabled: true})
	bm.CreatePolicy(ctx, &BackupPolicy{ID: "lp-b2", Name: "b2", Type: BackupTypeIncrement, Tables: []string{"archive_tasks"}, Enabled: true})

	policies, err := bm.ListPolicies(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(policies), 2)
}

func TestCov_Backup_executeRestore_NonexistentBackup(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	restoreRecord := &RestoreRecord{
		ID: "exec-restore-nf", BackupID: "nonexistent-backup",
		Status: BackupStatusPending, CreatedAt: time.Now(),
	}
	bm.db.Create(restoreRecord)

	bm.executeRestore(ctx, restoreRecord)

	var updated RestoreRecord
	bm.db.First(&updated, "id = ?", "exec-restore-nf")
	assert.Equal(t, BackupStatusFailed, updated.Status)
}

func TestCov_Cleaner_cleanupOldLogs(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	dc.db.Create(&CleanupLog{ID: "col-old-1", TaskID: "t1", Level: "info", Message: "old log", CreatedAt: time.Now().Add(-100 * 24 * time.Hour)})
	dc.db.Create(&CleanupLog{ID: "col-new-1", TaskID: "t2", Level: "info", Message: "new log", CreatedAt: time.Now()})

	dc.cleanupOldLogs(ctx)

	var logs []CleanupLog
	dc.db.Find(&logs)
	assert.NotEmpty(t, logs)
}

func TestCov_Archiver_updateMetrics(t *testing.T) {
	da, _, cleanup := newCovTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	da.db.Create(&ArchiveTask{ID: "um-at-1", Status: ArchiveStatusCompleted, CreatedAt: time.Now()})
	da.db.Create(&ArchiveTask{ID: "um-at-2", Status: ArchiveStatusFailed, CreatedAt: time.Now()})

	da.updateMetrics(ctx)

	metrics := da.GetMetrics()
	assert.Equal(t, int64(2), metrics.TotalTasks)
	assert.Equal(t, int64(1), metrics.CompletedTasks)
}

func TestCov_Backup_updateMetrics(t *testing.T) {
	bm, _, cleanup := newCovTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "um-br-1", Status: BackupStatusCompleted, CreatedAt: time.Now()})
	bm.db.Create(&BackupRecord{ID: "um-br-2", Status: BackupStatusFailed, CreatedAt: time.Now()})

	bm.updateMetrics(ctx)

	metrics := bm.GetMetrics()
	assert.Equal(t, int64(2), metrics.TotalBackups)
	assert.Equal(t, int64(1), metrics.CompletedBackups)
}

func TestCov_Cleaner_updateMetrics(t *testing.T) {
	dc, cleanup := newCovTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	dc.db.Create(&CleanupTask{ID: "um-ct-1", Status: CleanupStatusCompleted, CreatedAt: time.Now()})
	dc.db.Create(&CleanupTask{ID: "um-ct-2", Status: CleanupStatusFailed, CreatedAt: time.Now()})

	dc.updateMetrics(ctx)

	metrics := dc.GetMetrics()
	assert.Equal(t, int64(2), metrics.TotalTasks)
	assert.Equal(t, int64(1), metrics.CompletedTasks)
}

func TestCov_Tiered_Delete_WarmOnly(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	jsonData, _ := json.Marshal(map[string]interface{}{"v": 1})
	ts.db.Exec("INSERT INTO warm_data (key, data, created_at, updated_at) VALUES (?, ?, ?, ?)",
		"del-warm-key", jsonData, time.Now(), time.Now())

	err := ts.Delete(ctx, "del-warm-key")
	require.NoError(t, err)
}

func TestCov_Tiered_Delete_ColdOnly(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	jsonData, _ := json.Marshal(map[string]interface{}{"v": 1})
	ts.db.Exec("INSERT INTO cold_data (key, data, created_at) VALUES (?, ?, ?)",
		"del-cold-key", jsonData, time.Now())

	err := ts.Delete(ctx, "del-cold-key")
	require.NoError(t, err)
}

func TestCov_Tiered_Migrate_WarmToCold(t *testing.T) {
	ts, cleanup := newCovTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	jsonData, _ := json.Marshal(map[string]interface{}{"v": 1})
	ts.db.Exec("INSERT INTO warm_data (key, data, created_at, updated_at) VALUES (?, ?, ?, ?)",
		"warm-cold-key", jsonData, time.Now(), time.Now())

	err := ts.Migrate(ctx, "warm-cold-key", TierWarm, TierCold)
	assert.Error(t, err)
}
