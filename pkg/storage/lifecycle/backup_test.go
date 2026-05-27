package lifecycle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBackupType_Constants(t *testing.T) {
	assert.Equal(t, BackupType("full"), BackupTypeFull)
	assert.Equal(t, BackupType("increment"), BackupTypeIncrement)
	assert.Equal(t, BackupType("differential"), BackupTypeDifferential)
}

func TestBackupStatus_Constants(t *testing.T) {
	assert.Equal(t, BackupStatus("pending"), BackupStatusPending)
	assert.Equal(t, BackupStatus("running"), BackupStatusRunning)
	assert.Equal(t, BackupStatus("completed"), BackupStatusCompleted)
	assert.Equal(t, BackupStatus("failed"), BackupStatusFailed)
	assert.Equal(t, BackupStatus("validating"), BackupStatusValidating)
	assert.Equal(t, BackupStatus("restoring"), BackupStatusRestoring)
}

func TestDefaultBackupConfig(t *testing.T) {
	cfg := DefaultBackupConfig()
	assert.Equal(t, "/data/backup", cfg.StoragePath)
	assert.Equal(t, "/tmp/backup", cfg.TempPath)
	assert.Equal(t, 30, cfg.DefaultRetention)
	assert.Equal(t, 10, cfg.MaxBackups)
	assert.Equal(t, -1, cfg.CompressionLevel)
	assert.False(t, cfg.EnableEncryption)
	assert.Equal(t, 4, cfg.ParallelWorkers)
	assert.Equal(t, 10000, cfg.BatchSize)
	assert.True(t, cfg.EnableVerify)
	assert.Equal(t, 3, cfg.RetryCount)
	assert.Equal(t, 5*time.Second, cfg.RetryDelay)
}

func TestBackupPolicy_Struct(t *testing.T) {
	policy := BackupPolicy{
		ID:            "bp1",
		Name:          "Daily Full Backup",
		Description:   "Daily full backup policy",
		Type:          BackupTypeFull,
		Schedule:      "0 2 * * *",
		RetentionDays: 30,
		MaxBackups:    10,
		Compression:   true,
		Encryption:    false,
		Destination:   "/data/backup",
		Tables:        []string{"devices", "metrics", "alarms"},
		Enabled:       true,
	}
	assert.Equal(t, "bp1", policy.ID)
	assert.Equal(t, BackupTypeFull, policy.Type)
	assert.True(t, policy.Compression)
	assert.Equal(t, 3, len(policy.Tables))
}

func TestBackupRecord_Struct(t *testing.T) {
	record := BackupRecord{
		ID:             "br1",
		PolicyID:       "bp1",
		Type:           BackupTypeFull,
		Status:         BackupStatusCompleted,
		Size:           1024 * 1024,
		CompressedSize: 512 * 1024,
		FilePath:       "/data/backup/full_20240101.json.gz",
		Checksum:       "abc123",
		TablesCount:    3,
		RecordsCount:   10000,
	}
	assert.Equal(t, "br1", record.ID)
	assert.Equal(t, BackupTypeFull, record.Type)
	assert.Equal(t, BackupStatusCompleted, record.Status)
	assert.Equal(t, int64(1024*1024), record.Size)
}

func TestBackupTableRecord_Struct(t *testing.T) {
	record := BackupTableRecord{
		ID:           "btr1",
		BackupID:     "br1",
		TableName:    "devices",
		RecordsCount: 5000,
		Size:         256 * 1024,
		Checksum:     "def456",
	}
	assert.Equal(t, "btr1", record.ID)
	assert.Equal(t, "devices", record.TableName)
	assert.Equal(t, int64(5000), record.RecordsCount)
}

func TestRestoreRecord_Struct(t *testing.T) {
	record := RestoreRecord{
		ID:              "rr1",
		BackupID:        "br1",
		Status:          BackupStatusCompleted,
		TablesRestored:  3,
		RecordsRestored: 10000,
	}
	assert.Equal(t, "rr1", record.ID)
	assert.Equal(t, BackupStatusCompleted, record.Status)
	assert.Equal(t, 3, record.TablesRestored)
}

func TestBackupConfig_Struct(t *testing.T) {
	cfg := BackupConfig{
		StoragePath:      "/data/backup",
		TempPath:         "/tmp/backup",
		DefaultRetention: 30,
		MaxBackups:       10,
		CompressionLevel: 6,
		EnableEncryption: true,
		EncryptionKey:    "secret-key",
		ParallelWorkers:  4,
		BatchSize:        1000,
		EnableVerify:     true,
		RetryCount:       3,
		RetryDelay:       5 * time.Second,
	}
	assert.Equal(t, "/data/backup", cfg.StoragePath)
	assert.True(t, cfg.EnableEncryption)
	assert.Equal(t, "secret-key", cfg.EncryptionKey)
}

func TestBackupMetrics_Struct(t *testing.T) {
	metrics := &BackupMetrics{
		TotalBackups:     10,
		CompletedBackups: 8,
		FailedBackups:    1,
		TotalSize:        1024 * 1024 * 1024,
		CompressedSize:   512 * 1024 * 1024,
		TotalRecords:     100000,
		LastBackupTime:   time.Now(),
		AverageDuration:  10 * time.Minute,
		CompressionRatio: 0.5,
	}
	assert.Equal(t, int64(10), metrics.TotalBackups)
	assert.Equal(t, int64(512*1024*1024), metrics.CompressedSize)
}
