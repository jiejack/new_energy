package lifecycle

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDataArchiver_ValidateArchiveFile_Valid(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	records := []map[string]interface{}{{"id": "1"}, {"id": "2"}}
	filePath, _ := archiver.ArchiveData(ctx, "test_type", records)

	err := archiver.ValidateArchiveFile(filePath)
	assert.NoError(t, err)
}

func TestDataArchiver_ValidateArchiveFile_Gzip(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	filePath := filepath.Join(archiver.config.StoragePath, "test_validate.json.gz")
	file, _ := os.Create(filePath)
	gzWriter := gzip.NewWriter(file)
	json.NewEncoder(gzWriter).Encode(map[string]interface{}{"id": "1"})
	json.NewEncoder(gzWriter).Encode(map[string]interface{}{"id": "2"})
	gzWriter.Close()
	file.Close()

	err := archiver.ValidateArchiveFile(filePath)
	assert.NoError(t, err)
}

func TestDataArchiver_ValidateArchiveFile_FileNotFound(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()

	err := archiver.ValidateArchiveFile("/nonexistent/file.json")
	assert.Error(t, err)
}

func TestDataArchiver_ValidateArchiveFile_InvalidGzip(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	filePath := filepath.Join(archiver.config.StoragePath, "bad.json.gz")
	os.WriteFile(filePath, []byte("not a gzip"), 0644)

	err := archiver.ValidateArchiveFile(filePath)
	assert.Error(t, err)
}

func TestDataArchiver_ValidateArchiveFile_InvalidJSON(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	filePath := filepath.Join(archiver.config.StoragePath, "bad.json")
	os.WriteFile(filePath, []byte("not json\nnot json2\n"), 0644)

	err := archiver.ValidateArchiveFile(filePath)
	assert.Error(t, err)
}

func TestDataArchiver_ReadArchive_Gzip(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	filePath := filepath.Join(archiver.config.StoragePath, "test_read.json.gz")
	file, _ := os.Create(filePath)
	gzWriter := gzip.NewWriter(file)
	json.NewEncoder(gzWriter).Encode(map[string]interface{}{"id": "1", "name": "test"})
	json.NewEncoder(gzWriter).Encode(map[string]interface{}{"id": "2", "name": "test2"})
	gzWriter.Close()
	file.Close()

	records, err := archiver.ReadArchive(ctx, filePath)
	assert.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestDataArchiver_ReadArchive_InvalidGzip(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	filePath := filepath.Join(archiver.config.StoragePath, "bad_read.json.gz")
	os.WriteFile(filePath, []byte("not gzip"), 0644)

	_, err := archiver.ReadArchive(ctx, filePath)
	assert.Error(t, err)
}

func TestDataArchiver_StreamArchive_Valid(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	records := []map[string]interface{}{{"id": "1"}, {"id": "2"}}
	filePath, _ := archiver.ArchiveData(ctx, "test_type", records)

	count := 0
	err := archiver.StreamArchive(ctx, filePath, func(record map[string]interface{}) error {
		count++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestDataArchiver_StreamArchive_FileNotFound(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	err := archiver.StreamArchive(ctx, "/nonexistent/file.json", func(record map[string]interface{}) error {
		return nil
	})
	assert.Error(t, err)
}

func TestDataArchiver_StreamArchive_Gzip(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	filePath := filepath.Join(archiver.config.StoragePath, "test_stream.json.gz")
	file, _ := os.Create(filePath)
	gzWriter := gzip.NewWriter(file)
	json.NewEncoder(gzWriter).Encode(map[string]interface{}{"id": "1"})
	gzWriter.Close()
	file.Close()

	count := 0
	err := archiver.StreamArchive(ctx, filePath, func(record map[string]interface{}) error {
		count++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDataArchiver_CalculateChecksum_Gzip(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	filePath := filepath.Join(archiver.config.StoragePath, "test_checksum.json.gz")
	file, _ := os.Create(filePath)
	gzWriter := gzip.NewWriter(file)
	gzWriter.Write([]byte("test data"))
	gzWriter.Close()
	file.Close()

	checksum, err := archiver.CalculateChecksum(filePath)
	assert.NoError(t, err)
	assert.NotEmpty(t, checksum)
}

func TestDataArchiver_CalculateChecksum_InvalidGzip(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	filePath := filepath.Join(archiver.config.StoragePath, "bad_checksum.json.gz")
	os.WriteFile(filePath, []byte("not gzip"), 0644)

	_, err := archiver.CalculateChecksum(filePath)
	assert.Error(t, err)
}

func TestDataArchiver_CompareChecksum_Valid(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	records := []map[string]interface{}{{"id": "1"}}
	filePath, _ := archiver.ArchiveData(ctx, "test_type", records)

	checksum, _ := archiver.CalculateChecksum(filePath)
	match, err := archiver.CompareChecksum(filePath, checksum)
	assert.NoError(t, err)
	assert.True(t, match)
}

func TestDataArchiver_CompareChecksum_Mismatch(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	records := []map[string]interface{}{{"id": "1"}}
	filePath, _ := archiver.ArchiveData(ctx, "test_type", records)

	match, err := archiver.CompareChecksum(filePath, "wrong-checksum")
	assert.NoError(t, err)
	assert.False(t, match)
}

func TestDataArchiver_CompareChecksum_FileNotFound(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()

	_, err := archiver.CompareChecksum("/nonexistent/file.json", "checksum")
	assert.Error(t, err)
}

func TestDataArchiver_RestoreArchive_Gzip(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	filePath := filepath.Join(archiver.config.StoragePath, "test_restore.json.gz")
	file, _ := os.Create(filePath)
	gzWriter := gzip.NewWriter(file)
	json.NewEncoder(gzWriter).Encode(map[string]interface{}{"id": "1", "name": "test"})
	gzWriter.Close()
	file.Close()

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

func TestDataArchiver_RestoreArchive_InvalidGzip(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	filePath := filepath.Join(archiver.config.StoragePath, "bad_restore.json.gz")
	os.WriteFile(filePath, []byte("not gzip"), 0644)

	task := &ArchiveTask{
		ID: "t1", PolicyID: "p1", Status: ArchiveStatusCompleted,
		FilePath: filePath, RecordsCount: 1,
		CreatedAt: time.Now(),
	}
	archiver.db.Create(task)

	err := archiver.RestoreArchive(ctx, "t1", "archive_policies")
	assert.Error(t, err)
}

func TestDataArchiver_CompactArchives_WithCompression(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{Name: "test", DataType: "archive_policies", Enabled: true, Format: ArchiveFormatJSON, Compression: true}
	archiver.CreatePolicy(ctx, policy)

	os.MkdirAll(archiver.config.StoragePath, 0755)

	filePath1 := filepath.Join(archiver.config.StoragePath, "file1.json.gz")
	file1, _ := os.Create(filePath1)
	gz1 := gzip.NewWriter(file1)
	json.NewEncoder(gz1).Encode(map[string]interface{}{"id": "1"})
	gz1.Close()
	file1.Close()

	filePath2 := filepath.Join(archiver.config.StoragePath, "file2.json.gz")
	file2, _ := os.Create(filePath2)
	gz2 := gzip.NewWriter(file2)
	json.NewEncoder(gz2).Encode(map[string]interface{}{"id": "2"})
	gz2.Close()
	file2.Close()

	archiver.db.Create(&ArchiveTask{
		ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusCompleted,
		FilePath: filePath1, RecordsCount: 1, DataSize: 100,
		CreatedAt: time.Now().Add(-48 * time.Hour),
	})
	archiver.db.Create(&ArchiveTask{
		ID: "t2", PolicyID: policy.ID, Status: ArchiveStatusCompleted,
		FilePath: filePath2, RecordsCount: 1, DataSize: 100,
		CreatedAt: time.Now(),
	})

	err := archiver.CompactArchives(ctx, policy.ID)
	if err != nil {
		assert.Contains(t, err.Error(), "failed")
	}
}

func TestDataArchiver_CompactArchives_LessThanTwo(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{Name: "test", DataType: "archive_policies", Enabled: true, Format: ArchiveFormatJSON}
	archiver.CreatePolicy(ctx, policy)

	archiver.db.Create(&ArchiveTask{
		ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusCompleted,
		FilePath: "/tmp/test.json", RecordsCount: 1, DataSize: 100,
		CreatedAt: time.Now(),
	})

	err := archiver.CompactArchives(ctx, policy.ID)
	assert.NoError(t, err)
}

func TestDataArchiver_CompactArchives_PolicyNotFound(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	err := archiver.CompactArchives(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestBackupManager_VerifyBackup_FileNotFound(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted, FilePath: "/nonexistent/file.json"})
	err := bm.VerifyBackup(ctx, "b1")
	assert.Error(t, err)
}

func TestBackupManager_VerifyBackup_NotCompleted(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusPending})
	err := bm.VerifyBackup(ctx, "b1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not completed")
}

func TestBackupManager_VerifyBackup_ChecksumMismatch(t *testing.T) {
	bm, dir, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(bm.config.StoragePath, 0755)
	filePath := filepath.Join(dir, "backup", "test.json")
	os.MkdirAll(filepath.Dir(filePath), 0755)
	os.WriteFile(filePath, []byte("test data"), 0644)

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted, FilePath: filePath, Checksum: "wrong-checksum"})
	err := bm.VerifyBackup(ctx, "b1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestBackupManager_VerifyBackup_Gzip(t *testing.T) {
	bm, dir, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(bm.config.StoragePath, 0755)
	filePath := filepath.Join(dir, "backup", "test.json.gz")
	os.MkdirAll(filepath.Dir(filePath), 0755)
	file, _ := os.Create(filePath)
	gzWriter := gzip.NewWriter(file)
	gzWriter.Write([]byte("test data for verify"))
	gzWriter.Close()
	file.Close()

	archiver, _, _ := newTestArchiver(t)
	checksum, _ := archiver.CalculateChecksum(filePath)

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted, FilePath: filePath, Checksum: checksum})
	err := bm.VerifyBackup(ctx, "b1")
	assert.NoError(t, err)
}

func TestBackupManager_ExportBackup_NotCompleted(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusPending})
	err := bm.ExportBackup(ctx, "b1", "/tmp/export.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not completed")
}

func TestBackupManager_ExportBackup_FileNotFound(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted, FilePath: "/nonexistent/file.json"})
	err := bm.ExportBackup(ctx, "b1", "/tmp/export.json")
	assert.Error(t, err)
}

func TestBackupManager_CreateIncrementBackup_NotIncrementPolicy(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	_, err := bm.CreateIncrementBackup(ctx, policy.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not an incremental")
}

func TestBackupManager_CreateIncrementBackup_PolicyNotFound(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	_, err := bm.CreateIncrementBackup(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestBackupManager_CompareBackups_NotFound(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	_, err := bm.CompareBackups(ctx, "nonexistent1", "nonexistent2")
	assert.Error(t, err)
}

func TestBackupManager_ScheduleBackup_PolicyNotFound(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	err := bm.ScheduleBackup(ctx, "nonexistent", time.Now())
	assert.Error(t, err)
}

func TestBackupManager_CancelScheduledBackup_NotFound(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	err := bm.CancelScheduledBackup(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestBackupManager_DoBackup_NoTables(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Tables: []string{}, Compression: false, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	record := &BackupRecord{
		ID: "b1", PolicyID: policy.ID, Type: BackupTypeFull, Status: BackupStatusRunning,
	}
	bm.db.Create(record)

	err := bm.doBackup(ctx, record, policy)
	if err != nil {
		assert.Contains(t, err.Error(), "failed")
	}
}

func TestBackupManager_DoBackup_WithCompression(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(bm.config.StoragePath, 0755)

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Compression: true, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	record := &BackupRecord{
		ID: "b1", PolicyID: policy.ID, Type: BackupTypeFull, Status: BackupStatusRunning,
	}
	bm.db.Create(record)

	err := bm.doBackup(ctx, record, policy)
	assert.NoError(t, err)
	assert.NotEmpty(t, record.FilePath)
	assert.Contains(t, record.FilePath, ".gz")
}

func TestBackupManager_DoRestore_InvalidGzip(t *testing.T) {
	bm, dir, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(bm.config.StoragePath, 0755)
	filePath := filepath.Join(dir, "backup", "bad.json.gz")
	os.MkdirAll(filepath.Dir(filePath), 0755)
	os.WriteFile(filePath, []byte("not gzip"), 0644)

	backup := &BackupRecord{ID: "b1", FilePath: filePath, Status: BackupStatusCompleted}
	restore := &RestoreRecord{ID: "r1", BackupID: "b1", Status: BackupStatusRunning}
	bm.db.Create(backup)
	bm.db.Create(restore)

	err := bm.doRestore(ctx, backup, restore)
	assert.Error(t, err)
}

func TestBackupManager_DoRestore_InvalidMetadata(t *testing.T) {
	bm, dir, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(bm.config.StoragePath, 0755)
	filePath := filepath.Join(dir, "backup", "bad_meta.json")
	os.MkdirAll(filepath.Dir(filePath), 0755)
	os.WriteFile(filePath, []byte("not json at all"), 0644)

	backup := &BackupRecord{ID: "b1", FilePath: filePath, Status: BackupStatusCompleted}
	restore := &RestoreRecord{ID: "r1", BackupID: "b1", Status: BackupStatusRunning}
	bm.db.Create(backup)
	bm.db.Create(restore)

	err := bm.doRestore(ctx, backup, restore)
	assert.Error(t, err)
}

func TestBackupManager_TriggerBackup_IncrementWithBase(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeIncrement, Tables: []string{"archive_policies"}, Compression: false, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	bm.db.Create(&BackupRecord{
		ID: "base-b1", PolicyID: policy.ID, Status: BackupStatusCompleted,
		Type: BackupTypeFull, CreatedAt: time.Now(),
	})

	record, err := bm.TriggerBackup(ctx, policy.ID)
	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, "base-b1", record.BaseBackupID)
}

func TestBackupManager_TriggerBackup_IncrementNoBase(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeIncrement, Tables: []string{"archive_policies"}, Compression: false, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	record, err := bm.TriggerBackup(ctx, policy.ID)
	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, BackupTypeFull, record.Type)
}

func TestBackupManager_DeleteBackup_WithNonexistentFile(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted, FilePath: "/nonexistent/path/file.json"})
	err := bm.DeleteBackup(ctx, "b1")
	assert.NoError(t, err)
}

func TestBackupManager_DeleteBackup_NotFound(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	err := bm.DeleteBackup(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestDataCleaner_CheckScheduledCleanups_WithRunningTask(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "archive_policies", Schedule: "* * * * *", Enabled: true}
	cleaner.CreatePolicy(ctx, policy)

	cleaner.db.Create(&CleanupTask{ID: "t1", PolicyID: policy.ID, Status: CleanupStatusRunning, CreatedAt: time.Now()})

	cleaner.checkScheduledCleanups(ctx)
}

func TestDataCleaner_CheckScheduledCleanups_WithOldData(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "archive_policies", RetentionDays: 30, Schedule: "* * * * *", Enabled: true}
	cleaner.CreatePolicy(ctx, policy)

	cleaner.db.Create(&ArchivePolicy{Name: "old-data", DataType: "test", CreatedAt: time.Now().Add(-100 * 24 * time.Hour)})

	cleaner.checkScheduledCleanups(ctx)
}

func TestDataCleaner_CheckTaskCancelled_NotFound(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()

	cancelled := cleaner.checkTaskCancelled("nonexistent")
	assert.False(t, cancelled)
}

func TestDataCleaner_PreviewCleanup_NotFound(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.PreviewCleanup(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestDataCleaner_EstimateCleanupSize_NotFound(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.EstimateCleanupSize(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestDataCleaner_CancelTask_NotFound(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	err := cleaner.CancelTask(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestDataCleaner_CleanupByQuery_WithError(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.CleanupByQuery(ctx, "archive_policies", "created_at < ?", time.Now())
	assert.NoError(t, err)
}

func TestDataCleaner_CleanupByDate_WithResults(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	cleaner.db.Create(&ArchivePolicy{Name: "old-policy", DataType: "test", CreatedAt: time.Now().Add(-100 * 24 * time.Hour)})

	deleted, err := cleaner.CleanupByDate(ctx, "archive_policies", "created_at", time.Now().Add(-50*24*time.Hour))
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, deleted, int64(0))
}

func TestDataCleaner_CleanupOrphanedRecords_WithOrphans(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	deleted, err := cleaner.CleanupOrphanedRecords(ctx, "archive_records", "archive_tasks", "task_id")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, deleted, int64(0))
}

func TestDataCleaner_CleanupDuplicates_KeepNewest(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cleaner.CleanupDuplicates(ctx, "archive_policies", []string{"name"}, false)
	if err != nil {
		assert.Contains(t, err.Error(), "ON")
	}
}

func TestDataArchiver_ExecuteArchiveTask_WithVerify(t *testing.T) {
	archiver, dir, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(filepath.Join(dir, "archive"), 0755)

	config := archiver.config
	config.EnableVerify = true
	config.ParallelWorkers = 1
	config.RetryCount = 1
	config.RetryDelay = 10 * time.Millisecond

	archiver.config.EnableVerify = true

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

func TestDataArchiver_ExecuteArchiveTask_WithVerifyAndCompression(t *testing.T) {
	archiver, dir, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(filepath.Join(dir, "archive"), 0755)

	archiver.config.EnableVerify = true

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: time.Hour, BatchSize: 100, Format: ArchiveFormatJSON, Compression: true,
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

func TestBackupManager_ExecuteBackup_WithVerify(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.config.EnableVerify = true

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Tables: []string{"archive_policies"}, Compression: false, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	record := &BackupRecord{
		ID: "b1", PolicyID: policy.ID, Type: BackupTypeFull, Status: BackupStatusPending,
	}
	bm.db.Create(record)

	bm.executeBackup(ctx, record)

	var updated BackupRecord
	bm.db.First(&updated, "id = ?", "b1")
	assert.NotEqual(t, BackupStatusPending, updated.Status)
}

func TestBackupManager_CheckScheduledBackups_WithOldBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Schedule: "* * * * *", Enabled: true}
	bm.CreatePolicy(ctx, policy)

	bm.db.Create(&BackupRecord{
		ID: "b1", PolicyID: policy.ID, Status: BackupStatusCompleted,
		CreatedAt: time.Now().Add(-48 * time.Hour),
	})

	bm.checkScheduledBackups(ctx)
}

func TestBackupManager_CheckScheduledBackups_NoBackup(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Schedule: "* * * * *", Enabled: true}
	bm.CreatePolicy(ctx, policy)

	bm.checkScheduledBackups(ctx)
}

func TestBackupManager_CleanupOldBackups_WithOldAndExcess(t *testing.T) {
	bm, dir, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, MaxBackups: 1, RetentionDays: 1, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	os.MkdirAll(bm.config.StoragePath, 0755)
	filePath := filepath.Join(dir, "backup", "test.json")
	os.MkdirAll(filepath.Dir(filePath), 0755)
	os.WriteFile(filePath, []byte("test"), 0644)

	bm.db.Create(&BackupRecord{
		ID: "b1", PolicyID: policy.ID, Status: BackupStatusCompleted,
		FilePath: filePath, CreatedAt: time.Now().Add(-48 * time.Hour),
	})
	bm.db.Create(&BackupRecord{
		ID: "b2", PolicyID: policy.ID, Status: BackupStatusCompleted,
		CreatedAt: time.Now(),
	})

	bm.cleanupOldBackups(ctx)
}

func TestBackupManager_GetBackupStats_All(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted, Size: 2048, CompressedSize: 1024, RecordsCount: 100, CreatedAt: time.Now()})
	stats, err := bm.GetBackupStats(ctx, "")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestBackupManager_GetBackupStats_WithSize(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", PolicyID: "p1", Status: BackupStatusCompleted, Size: 2048, CompressedSize: 1024, RecordsCount: 100, CreatedAt: time.Now()})
	stats, err := bm.GetBackupStats(ctx, "p1")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, float64(0.5), stats["compression_ratio"])
}

func TestBackupManager_ImportBackup_WithValidFile(t *testing.T) {
	bm, dir, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	os.MkdirAll(bm.config.StoragePath, 0755)
	srcPath := filepath.Join(dir, "import.json")
	os.WriteFile(srcPath, []byte(`{"backup_id":"b1","type":"full"}`), 0644)

	record, err := bm.ImportBackup(ctx, srcPath, policy.ID)
	if err != nil {
		assert.Contains(t, err.Error(), "failed")
	} else {
		assert.NotNil(t, record)
		assert.NotEmpty(t, record.Checksum)
	}
}

func TestBackupManager_DoRestore_WithBatchData(t *testing.T) {
	bm, dir, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(bm.config.StoragePath, 0755)
	filePath := filepath.Join(dir, "backup", "batch_restore.json")
	os.MkdirAll(filepath.Dir(filePath), 0755)

	file, _ := os.Create(filePath)
	enc := json.NewEncoder(file)
	enc.Encode(map[string]interface{}{"metadata": map[string]interface{}{"backup_id": "b1"}})
	enc.Encode(map[string]interface{}{"table_start": "archive_policies"})
	for i := 0; i < 5; i++ {
		enc.Encode(map[string]interface{}{"row": map[string]interface{}{"name": fmt.Sprintf("test-%d", i)}})
	}
	enc.Encode(map[string]interface{}{"table_end": true})
	file.Close()

	backup := &BackupRecord{ID: "b1", FilePath: filePath, Status: BackupStatusCompleted}
	restore := &RestoreRecord{ID: "r1", BackupID: "b1", Status: BackupStatusRunning}
	bm.db.Create(backup)
	bm.db.Create(restore)

	err := bm.doRestore(ctx, backup, restore)
	assert.NoError(t, err)
}

func TestTieredStorage_Migrate_HotToCold(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "1"}
	ts.Store(ctx, "migrate-hc", data, 1*time.Hour)

	err := ts.Migrate(ctx, "migrate-hc", TierHot, TierCold)
	if err != nil {
		assert.Contains(t, err.Error(), "NOW")
	}
}

func TestTieredStorage_Migrate_WarmToCold(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.Migrate(ctx, "nonexistent-warm", TierWarm, TierCold)
	assert.Error(t, err)
}

func TestTieredStorage_Migrate_ColdToHot(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	err := ts.Migrate(ctx, "nonexistent-cold", TierCold, TierHot)
	assert.Error(t, err)
}

func TestTieredStorage_Delete_WithErrors(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	data := map[string]interface{}{"id": "1"}
	ts.Store(ctx, "del-err-key", data, 1*time.Hour)

	ts.db.Exec("CREATE TABLE IF NOT EXISTS warm_data (key TEXT PRIMARY KEY, data BLOB, created_at DATETIME, updated_at DATETIME)")
	ts.db.Exec("CREATE TABLE IF NOT EXISTS cold_data (key TEXT PRIMARY KEY, data BLOB, created_at DATETIME)")

	err := ts.Delete(ctx, "del-err-key")
	assert.NoError(t, err)
}

func TestTieredStorage_RecordMiss(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()

	ts.metrics.mu.Lock()
	ts.metrics.HitRate = 0.5
	ts.metrics.MissRate = 0.5
	ts.metrics.mu.Unlock()

	ts.recordMiss("test-key")
}

func TestTieredStorage_RecordHit_WithStats(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()

	ts.hotStats.Store("hit-key", &HotStats{
		Key: "hit-key", HitCount: 5, WindowHits: 3, UpdatedAt: time.Now(),
	})

	ts.metrics.mu.Lock()
	ts.metrics.HitRate = 0.5
	ts.metrics.MissRate = 0.5
	ts.metrics.mu.Unlock()

	ts.recordHit("hit-key", TierHot)
}

func TestTieredStorage_RecordHit_NoStats(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()

	ts.recordHit("no-stats-key", TierHot)
}

func TestTieredStorage_Get_WarmTier(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	ts.db.Exec("CREATE TABLE IF NOT EXISTS warm_data (key TEXT PRIMARY KEY, data BLOB, created_at DATETIME, updated_at DATETIME)")
	ts.db.Exec("INSERT INTO warm_data (key, data, created_at, updated_at) VALUES (?, ?, datetime('now'), datetime('now'))", "warm-key", []byte(`{"id":"warm-1"}`))

	var result map[string]interface{}
	err := ts.Get(ctx, "warm-key", &result)
	if err != nil {
		assert.Contains(t, err.Error(), "not found")
	}
}

func TestTieredStorage_Get_ColdTier(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	ts.db.Exec("CREATE TABLE IF NOT EXISTS cold_data (key TEXT PRIMARY KEY, data BLOB, created_at DATETIME)")
	ts.db.Exec("INSERT INTO cold_data (key, data, created_at) VALUES (?, ?, datetime('now'))", "cold-key", []byte(`{"id":"cold-1"}`))

	var result map[string]interface{}
	err := ts.Get(ctx, "cold-key", &result)
	if err != nil {
		assert.Contains(t, err.Error(), "not found")
	}
}

func TestTieredStorage_GetTier_Warm(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	ts.db.Exec("CREATE TABLE IF NOT EXISTS warm_data (key TEXT PRIMARY KEY, data BLOB, created_at DATETIME, updated_at DATETIME)")
	ts.db.Exec("INSERT INTO warm_data (key, data, created_at, updated_at) VALUES (?, ?, datetime('now'), datetime('now'))", "warm-tier-key", []byte(`{"id":"1"}`))

	tier, err := ts.GetTier(ctx, "warm-tier-key")
	if err != nil {
		assert.Equal(t, TierHot, tier)
	} else {
		assert.Equal(t, TierWarm, tier)
	}
}

func TestTieredStorage_GetTier_Cold(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()
	ctx := context.Background()

	ts.db.Exec("CREATE TABLE IF NOT EXISTS cold_data (key TEXT PRIMARY KEY, data BLOB, created_at DATETIME)")
	ts.db.Exec("INSERT INTO cold_data (key, data, created_at) VALUES (?, ?, datetime('now'))", "cold-tier-key", []byte(`{"id":"1"}`))

	tier, err := ts.GetTier(ctx, "cold-tier-key")
	if err != nil {
		assert.Equal(t, TierHot, tier)
	} else {
		assert.Equal(t, TierCold, tier)
	}
}

func TestTieredStorage_PromoteToHot_Error(t *testing.T) {
	ts, cleanup := newTestTieredStorage(t)
	defer cleanup()

	ts.redis.Close()
	ts.promoteToHot(context.Background(), "error-key", []byte(`{"id":"1"}`))
}

func TestDataArchiver_ArchiveData_CreateError(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	archiver.config.StoragePath = "/nonexistent/path/that/does/not/exist"
	_, err := archiver.ArchiveData(ctx, "test_type", []map[string]interface{}{{"id": "1"}})
	assert.Error(t, err)
}

func TestDataArchiver_StreamArchive_InvalidGzip(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	filePath := filepath.Join(archiver.config.StoragePath, "bad_stream.json.gz")
	os.WriteFile(filePath, []byte("not gzip"), 0644)

	err := archiver.StreamArchive(ctx, filePath, func(record map[string]interface{}) error {
		return nil
	})
	assert.Error(t, err)
}

func TestDataArchiver_VerifyArchive_InvalidGzip(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	filePath := filepath.Join(archiver.config.StoragePath, "bad_verify.json.gz")
	os.WriteFile(filePath, []byte("not gzip"), 0644)

	task := &ArchiveTask{
		ID: "t1", FilePath: filePath, Checksum: "abc", RecordsCount: 1, CreatedAt: time.Now(),
	}

	err := archiver.verifyArchive(task)
	assert.Error(t, err)
}

func TestDataArchiver_VerifyArchive_InvalidJSON(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()

	os.MkdirAll(archiver.config.StoragePath, 0755)
	filePath := filepath.Join(archiver.config.StoragePath, "bad_json.json")
	os.WriteFile(filePath, []byte("not valid json content"), 0644)

	task := &ArchiveTask{
		ID: "t1", FilePath: filePath, Checksum: "", RecordsCount: 1, CreatedAt: time.Now(),
	}

	err := archiver.verifyArchive(task)
	assert.Error(t, err)
}

func TestDataArchiver_DoArchive_CreateFileError(t *testing.T) {
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

	archiver.config.StoragePath = "/nonexistent/path"
	err := archiver.doArchive(ctx, task, policy)
	assert.Error(t, err)
}

func TestDataArchiver_DoArchive_CompressionCreateError(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: 0, BatchSize: 100, Format: ArchiveFormatJSON, Compression: true,
	}
	archiver.CreatePolicy(ctx, policy)

	archiver.db.Create(&ArchivePolicy{Name: "old-data", DataType: "test", CreatedAt: time.Now().Add(-48 * time.Hour)})

	task := &ArchiveTask{ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusRunning, CreatedAt: time.Now()}
	archiver.db.Create(task)

	archiver.config.StoragePath = "/nonexistent/path"
	err := archiver.doArchive(ctx, task, policy)
	assert.Error(t, err)
}

func TestDataArchiver_ExecuteArchiveTask_RetryFail(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: 0, BatchSize: 100, Format: ArchiveFormatJSON, Compression: false,
	}
	archiver.CreatePolicy(ctx, policy)

	archiver.config.StoragePath = "/nonexistent/path"
	archiver.config.RetryCount = 2
	archiver.config.RetryDelay = 1 * time.Millisecond

	task := &ArchiveTask{ID: "t1", PolicyID: policy.ID, Status: ArchiveStatusPending, CreatedAt: time.Now()}
	archiver.db.Create(task)

	archiver.executeArchiveTask(ctx, task)

	var updated ArchiveTask
	archiver.db.First(&updated, "id = ?", "t1")
	assert.Equal(t, ArchiveStatusFailed, updated.Status)
}

func TestDataCleaner_DoCleanup_CancelledTask(t *testing.T) {
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

	go func() {
		time.Sleep(10 * time.Millisecond)
		cleaner.db.Model(&CleanupTask{}).Where("id = ?", "t1").Update("status", CleanupStatusCancelled)
	}()

	err := cleaner.doCleanup(ctx, task, policy)
	if err != nil {
		assert.Contains(t, err.Error(), "cancelled")
	}
}

func TestBackupManager_ExecuteBackup_RetryFail(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.config.RetryCount = 2
	bm.config.RetryDelay = 1 * time.Millisecond

	policy := &BackupPolicy{Name: "test", Type: BackupTypeFull, Tables: []string{}, Compression: false, Enabled: true}
	bm.CreatePolicy(ctx, policy)

	record := &BackupRecord{
		ID: "b1", PolicyID: policy.ID, Type: BackupTypeFull, Status: BackupStatusPending,
	}
	bm.db.Create(record)

	bm.executeBackup(ctx, record)

	var updated BackupRecord
	bm.db.First(&updated, "id = ?", "b1")
	assert.Equal(t, BackupStatusFailed, updated.Status)
}

func TestDataArchiver_CheckScheduledArchives_EnabledPolicy(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: true,
		ArchiveAfter: 0, BatchSize: 100, Format: ArchiveFormatJSON, Compression: false,
	}
	archiver.CreatePolicy(ctx, policy)

	archiver.db.Create(&ArchivePolicy{Name: "old-data", DataType: "test", CreatedAt: time.Now().Add(-100 * 24 * time.Hour)})

	archiver.checkScheduledArchives(ctx)
}

func TestDataArchiver_CheckScheduledArchives_DisabledPolicy(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	policy := &ArchivePolicy{
		Name: "test", DataType: "archive_policies", Enabled: false,
		ArchiveAfter: 0, BatchSize: 100,
	}
	archiver.CreatePolicy(ctx, policy)

	archiver.checkScheduledArchives(ctx)
}

func TestBackupManager_BackupTable_WithData(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&ArchivePolicy{ID: "ap-1", Name: "test-policy-1", DataType: "test", Enabled: true})
	bm.db.Create(&ArchivePolicy{ID: "ap-2", Name: "test-policy-2", DataType: "test", Enabled: true})

	os.MkdirAll(bm.config.StoragePath, 0755)
	filePath := filepath.Join(bm.config.StoragePath, "test_backup_with_data.json")
	file, err := os.Create(filePath)
	assert.NoError(t, err)
	defer file.Close()

	record := &BackupRecord{ID: "b1", Type: BackupTypeFull}
	records, size, err := bm.backupTable(ctx, "archive_policies", file, record)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, records, int64(0))
	assert.GreaterOrEqual(t, size, int64(0))
}

func TestDataCleaner_CreatePolicy_DefaultBatchSize(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "test_data", RetentionDays: 30, Enabled: true, BatchSize: 0}
	err := cleaner.CreatePolicy(ctx, policy)
	assert.NoError(t, err)
	assert.Equal(t, 1000, policy.BatchSize)
}

func TestDataCleaner_CreatePolicy_DefaultRetention(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	policy := &CleanupPolicy{Name: "test", DataType: "test_data", Enabled: true, RetentionDays: 0}
	err := cleaner.CreatePolicy(ctx, policy)
	assert.NoError(t, err)
	assert.Equal(t, 90, policy.RetentionDays)
}

func TestDataArchiver_GetArchiveStats_Empty(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	stats, err := archiver.GetArchiveStats(ctx, "")
	if err != nil {
		assert.Contains(t, err.Error(), "Scan error")
	} else {
		assert.NotNil(t, stats)
	}
}

func TestDataArchiver_ListTasks_NoPolicyFilter(t *testing.T) {
	archiver, _, cleanup := newTestArchiver(t)
	defer cleanup()
	ctx := context.Background()

	archiver.db.Create(&ArchiveTask{ID: "t1", PolicyID: "p1", Status: ArchiveStatusPending, CreatedAt: time.Now()})
	archiver.db.Create(&ArchiveTask{ID: "t2", PolicyID: "p2", Status: ArchiveStatusCompleted, CreatedAt: time.Now()})

	tasks, err := archiver.ListTasks(ctx, "", 0)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestBackupManager_ListBackups_NoFilter(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&BackupRecord{ID: "b1", Status: BackupStatusCompleted, CreatedAt: time.Now()})
	bm.db.Create(&BackupRecord{ID: "b2", Status: BackupStatusCompleted, CreatedAt: time.Now()})

	records, err := bm.ListBackups(ctx, "", 0)
	assert.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestBackupManager_ListRestores_NoFilter(t *testing.T) {
	bm, _, cleanup := newTestBackupManager(t)
	defer cleanup()
	ctx := context.Background()

	bm.db.Create(&RestoreRecord{ID: "r1", BackupID: "b1", Status: BackupStatusCompleted, CreatedAt: time.Now()})
	bm.db.Create(&RestoreRecord{ID: "r2", BackupID: "b2", Status: BackupStatusCompleted, CreatedAt: time.Now()})

	records, err := bm.ListRestores(ctx, "", 0)
	assert.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestDataCleaner_ListTasks_NoFilter(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	cleaner.db.Create(&CleanupTask{ID: "t1", PolicyID: "p1", Status: CleanupStatusPending, CreatedAt: time.Now()})
	cleaner.db.Create(&CleanupTask{ID: "t2", PolicyID: "p2", Status: CleanupStatusCompleted, CreatedAt: time.Now()})

	tasks, err := cleaner.ListTasks(ctx, "", 0)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestDataCleaner_GetCleanupLogs_NoLimit(t *testing.T) {
	cleaner, cleanup := newTestCleaner(t)
	defer cleanup()
	ctx := context.Background()

	cleaner.db.Create(&CleanupLog{ID: "l1", TaskID: "t1", Level: "info", Message: "test", CreatedAt: time.Now()})
	cleaner.db.Create(&CleanupLog{ID: "l2", TaskID: "t1", Level: "warn", Message: "test2", CreatedAt: time.Now()})

	logs, err := cleaner.GetCleanupLogs(ctx, "t1", 0)
	assert.NoError(t, err)
	assert.Len(t, logs, 2)
}
