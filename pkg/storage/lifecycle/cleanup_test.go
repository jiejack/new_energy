package lifecycle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCleanupStatus_Constants(t *testing.T) {
	assert.Equal(t, CleanupStatus("pending"), CleanupStatusPending)
	assert.Equal(t, CleanupStatus("running"), CleanupStatusRunning)
	assert.Equal(t, CleanupStatus("completed"), CleanupStatusCompleted)
	assert.Equal(t, CleanupStatus("failed"), CleanupStatusFailed)
	assert.Equal(t, CleanupStatus("cancelled"), CleanupStatusCancelled)
}

func TestDefaultCleanupConfig(t *testing.T) {
	cfg := DefaultCleanupConfig()
	assert.Equal(t, 1000, cfg.DefaultBatchSize)
	assert.Equal(t, 10000, cfg.MaxBatchSize)
	assert.Equal(t, 90, cfg.DefaultRetention)
	assert.Equal(t, 4, cfg.ParallelWorkers)
	assert.Equal(t, 1*time.Hour, cfg.ScheduleInterval)
	assert.True(t, cfg.EnableAutoCleanup)
	assert.Equal(t, 30, cfg.LogRetentionDays)
	assert.Equal(t, 3, cfg.MaxRetryCount)
	assert.Equal(t, 5*time.Second, cfg.RetryDelay)
}

func TestCleanupPolicy_Struct(t *testing.T) {
	policy := CleanupPolicy{
		ID:              "cp1",
		Name:            "Test Cleanup",
		Description:     "A test cleanup policy",
		DataType:        "metrics",
		RetentionDays:   30,
		BatchSize:       1000,
		Schedule:        "0 3 * * *",
		Enabled:         true,
		DryRun:          false,
		ArchiveFirst:    true,
		NotifyOnComplete: true,
	}
	assert.Equal(t, "cp1", policy.ID)
	assert.Equal(t, "metrics", policy.DataType)
	assert.True(t, policy.ArchiveFirst)
	assert.True(t, policy.NotifyOnComplete)
}

func TestCleanupTask_Struct(t *testing.T) {
	task := CleanupTask{
		ID:             "ct1",
		PolicyID:       "cp1",
		Status:         CleanupStatusPending,
		RecordsScanned: 1000,
		RecordsDeleted: 500,
		DataFreed:      1024 * 1024,
		DryRun:         false,
	}
	assert.Equal(t, "ct1", task.ID)
	assert.Equal(t, CleanupStatusPending, task.Status)
	assert.Equal(t, int64(1000), task.RecordsScanned)
	assert.Equal(t, int64(500), task.RecordsDeleted)
}

func TestCleanupLog_Struct(t *testing.T) {
	log := CleanupLog{
		ID:       "cl1",
		TaskID:   "ct1",
		Level:    "info",
		Message:  "Cleanup completed",
		RecordID: "rec1",
		Details:  `{"key": "value"}`,
	}
	assert.Equal(t, "cl1", log.ID)
	assert.Equal(t, "info", log.Level)
}

func TestCleanupConfig_Struct(t *testing.T) {
	cfg := CleanupConfig{
		DefaultBatchSize:  500,
		MaxBatchSize:      5000,
		DefaultRetention:  60,
		ParallelWorkers:   2,
		ScheduleInterval:  30 * time.Minute,
		EnableAutoCleanup: false,
		LogRetentionDays:  14,
		MaxRetryCount:     5,
		RetryDelay:        10 * time.Second,
	}
	assert.Equal(t, 500, cfg.DefaultBatchSize)
	assert.False(t, cfg.EnableAutoCleanup)
}

func TestCleanupMetrics_Struct(t *testing.T) {
	metrics := &CleanupMetrics{
		TotalTasks:         10,
		CompletedTasks:     8,
		FailedTasks:        1,
		TotalRecordsDeleted: 5000,
		TotalDataFreed:     100 * 1024 * 1024,
		LastCleanupTime:    time.Now(),
		AverageDuration:    3 * time.Minute,
	}
	assert.Equal(t, int64(10), metrics.TotalTasks)
	assert.Equal(t, int64(5000), metrics.TotalRecordsDeleted)
}
