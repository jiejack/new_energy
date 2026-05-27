package lifecycle

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestArchiveFormat_Constants(t *testing.T) {
	assert.Equal(t, ArchiveFormat("parquet"), ArchiveFormatParquet)
	assert.Equal(t, ArchiveFormat("arrow"), ArchiveFormatArrow)
	assert.Equal(t, ArchiveFormat("json"), ArchiveFormatJSON)
	assert.Equal(t, ArchiveFormat("csv"), ArchiveFormatCSV)
}

func TestArchiveStatus_Constants(t *testing.T) {
	assert.Equal(t, ArchiveStatus("pending"), ArchiveStatusPending)
	assert.Equal(t, ArchiveStatus("running"), ArchiveStatusRunning)
	assert.Equal(t, ArchiveStatus("completed"), ArchiveStatusCompleted)
	assert.Equal(t, ArchiveStatus("failed"), ArchiveStatusFailed)
	assert.Equal(t, ArchiveStatus("validating"), ArchiveStatusValidating)
}

func TestDefaultArchiveConfig(t *testing.T) {
	cfg := DefaultArchiveConfig()
	assert.Equal(t, "/data/archive", cfg.StoragePath)
	assert.Equal(t, "/tmp/archive", cfg.TempPath)
	assert.Equal(t, int64(1024*1024*1024), cfg.MaxFileSize)
	assert.Equal(t, ArchiveFormatJSON, cfg.DefaultFormat)
	assert.True(t, cfg.EnableVerify)
	assert.Equal(t, 4, cfg.ParallelWorkers)
	assert.Equal(t, 3, cfg.RetryCount)
	assert.Equal(t, 5*time.Second, cfg.RetryDelay)
}

func TestArchivePolicy_Struct(t *testing.T) {
	policy := ArchivePolicy{
		ID:            "p1",
		Name:          "Test Policy",
		Description:   "A test policy",
		DataType:      "metrics",
		RetentionDays: 30,
		ArchiveAfter:  7 * 24 * time.Hour,
		Format:        ArchiveFormatJSON,
		Compression:   true,
		BatchSize:     1000,
		Schedule:      "0 2 * * *",
		Destination:   "/archive/metrics",
		Enabled:       true,
	}
	assert.Equal(t, "p1", policy.ID)
	assert.Equal(t, "metrics", policy.DataType)
	assert.True(t, policy.Compression)
	assert.True(t, policy.Enabled)
}

func TestArchiveTask_Struct(t *testing.T) {
	task := ArchiveTask{
		ID:           "t1",
		PolicyID:     "p1",
		Status:       ArchiveStatusPending,
		RecordsCount: 0,
		DataSize:     0,
		FilePath:     "",
		Checksum:     "",
	}
	assert.Equal(t, "t1", task.ID)
	assert.Equal(t, ArchiveStatusPending, task.Status)
}

func TestArchiveRecord_Struct(t *testing.T) {
	record := ArchiveRecord{
		ID:         "r1",
		TaskID:     "t1",
		OriginalID: "orig1",
		DataType:   "metrics",
		FilePath:   "/archive/metrics/file.json",
		Offset:     0,
		Size:       256,
		Checksum:   "abc123",
	}
	assert.Equal(t, "r1", record.ID)
	assert.Equal(t, int64(256), record.Size)
}

func TestNewDataArchiver(t *testing.T) {
	da := NewDataArchiver(DefaultArchiveConfig(), nil, zap.NewNop())
	require.NotNil(t, da)
}

func TestDataArchiver_GetMetrics(t *testing.T) {
	da := NewDataArchiver(DefaultArchiveConfig(), nil, zap.NewNop())
	metrics := da.GetMetrics()
	require.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.TotalTasks)
	assert.Equal(t, int64(0), metrics.CompletedTasks)
	assert.Equal(t, int64(0), metrics.FailedTasks)
}

func TestDataArchiver_ArchiveData(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultArchiveConfig()
	cfg.StoragePath = tmpDir
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	records := []map[string]interface{}{
		{"id": "1", "value": 42.5},
		{"id": "2", "value": 37.8},
	}

	filePath, err := da.ArchiveData(context.Background(), "metrics", records)
	require.NoError(t, err)
	assert.NotEmpty(t, filePath)
	assert.FileExists(t, filePath)
}

func TestDataArchiver_ArchiveData_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultArchiveConfig()
	cfg.StoragePath = tmpDir
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	_, err := da.ArchiveData(context.Background(), "metrics", nil)
	assert.Error(t, err)
}

func TestDataArchiver_ReadArchive(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultArchiveConfig()
	cfg.StoragePath = tmpDir
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	records := []map[string]interface{}{
		{"id": "1", "value": 42.5},
		{"id": "2", "value": 37.8},
	}

	filePath, _ := da.ArchiveData(context.Background(), "metrics", records)

	readRecords, err := da.ReadArchive(context.Background(), filePath)
	require.NoError(t, err)
	assert.Equal(t, 2, len(readRecords))
}

func TestDataArchiver_ReadArchive_NotFound(t *testing.T) {
	cfg := DefaultArchiveConfig()
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	_, err := da.ReadArchive(context.Background(), "/nonexistent/file.json")
	assert.Error(t, err)
}

func TestDataArchiver_StreamArchive(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultArchiveConfig()
	cfg.StoragePath = tmpDir
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	records := []map[string]interface{}{
		{"id": "1", "value": 42.5},
		{"id": "2", "value": 37.8},
	}

	filePath, _ := da.ArchiveData(context.Background(), "metrics", records)

	count := 0
	err := da.StreamArchive(context.Background(), filePath, func(record map[string]interface{}) error {
		count++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestDataArchiver_CalculateChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")
	os.WriteFile(filePath, []byte(`{"key":"value"}`), 0644)

	cfg := DefaultArchiveConfig()
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	checksum, err := da.CalculateChecksum(filePath)
	require.NoError(t, err)
	assert.NotEmpty(t, checksum)
}

func TestDataArchiver_CalculateChecksum_NotFound(t *testing.T) {
	cfg := DefaultArchiveConfig()
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	_, err := da.CalculateChecksum("/nonexistent/file.json")
	assert.Error(t, err)
}

func TestDataArchiver_CompareChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")
	os.WriteFile(filePath, []byte(`{"key":"value"}`), 0644)

	cfg := DefaultArchiveConfig()
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	checksum, _ := da.CalculateChecksum(filePath)

	match, err := da.CompareChecksum(filePath, checksum)
	require.NoError(t, err)
	assert.True(t, match)

	match, err = da.CompareChecksum(filePath, "wrong_checksum")
	require.NoError(t, err)
	assert.False(t, match)
}

func TestDataArchiver_ValidateArchiveFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultArchiveConfig()
	cfg.StoragePath = tmpDir
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	records := []map[string]interface{}{
		{"id": "1", "value": 42.5},
	}

	filePath, _ := da.ArchiveData(context.Background(), "metrics", records)
	err := da.ValidateArchiveFile(filePath)
	require.NoError(t, err)
}

func TestDataArchiver_ValidateArchiveFile_NotFound(t *testing.T) {
	cfg := DefaultArchiveConfig()
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	err := da.ValidateArchiveFile("/nonexistent/file.json")
	assert.Error(t, err)
}

func TestNewArchiveBuffer(t *testing.T) {
	ab := NewArchiveBuffer()
	require.NotNil(t, ab)
	assert.Equal(t, 0, ab.Len())
}

func TestArchiveBuffer_Write_Read(t *testing.T) {
	ab := NewArchiveBuffer()
	n, err := ab.Write([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 5, ab.Len())
	assert.Equal(t, []byte("hello"), ab.Bytes())
}

func TestArchiveBuffer_Reset(t *testing.T) {
	ab := NewArchiveBuffer()
	ab.Write([]byte("hello"))
	ab.Reset()
	assert.Equal(t, 0, ab.Len())
	assert.Equal(t, []byte{}, ab.Bytes())
}

func TestArchiveBuffer_MultipleWrites(t *testing.T) {
	ab := NewArchiveBuffer()
	ab.Write([]byte("hello"))
	ab.Write([]byte(" world"))
	assert.Equal(t, 11, ab.Len())
	assert.Equal(t, []byte("hello world"), ab.Bytes())
}

func TestArchiveMetrics_Struct(t *testing.T) {
	metrics := &ArchiveMetrics{
		TotalTasks:       10,
		CompletedTasks:   8,
		FailedTasks:      2,
		TotalRecords:     1000,
		TotalDataSize:    1024 * 1024,
		LastArchiveTime:  time.Now(),
		AverageDuration:  5 * time.Minute,
		CompressionRatio: 0.5,
	}
	assert.Equal(t, int64(10), metrics.TotalTasks)
	assert.Equal(t, int64(8), metrics.CompletedTasks)
}

func TestArchiveConfig_Struct(t *testing.T) {
	cfg := ArchiveConfig{
		StoragePath:      "/data/archive",
		TempPath:         "/tmp/archive",
		MaxFileSize:      1024,
		DefaultFormat:    ArchiveFormatJSON,
		CompressionLevel: 6,
		EnableVerify:     true,
		ParallelWorkers:  4,
		RetryCount:       3,
		RetryDelay:       5 * time.Second,
	}
	assert.Equal(t, "/data/archive", cfg.StoragePath)
	assert.Equal(t, ArchiveFormatJSON, cfg.DefaultFormat)
}

func TestArchiveData_WithCompression(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultArchiveConfig()
	cfg.StoragePath = tmpDir
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	records := []map[string]interface{}{
		{"id": "1", "value": 42.5},
	}

	filePath, err := da.ArchiveData(context.Background(), "metrics", records)
	require.NoError(t, err)
	assert.FileExists(t, filePath)
}

func TestArchiveData_JSON_Roundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultArchiveConfig()
	cfg.StoragePath = tmpDir
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	original := []map[string]interface{}{
		{"id": "1", "name": "test", "value": 42.5},
		{"id": "2", "name": "test2", "value": 37.8},
	}

	filePath, _ := da.ArchiveData(context.Background(), "metrics", original)

	readBack, err := da.ReadArchive(context.Background(), filePath)
	require.NoError(t, err)
	assert.Equal(t, len(original), len(readBack))

	for i, record := range readBack {
		assert.Equal(t, original[i]["id"], record["id"])
	}
}

func TestStreamArchive_HandlerError(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultArchiveConfig()
	cfg.StoragePath = tmpDir
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	records := []map[string]interface{}{
		{"id": "1", "value": 42.5},
	}

	filePath, _ := da.ArchiveData(context.Background(), "metrics", records)

	err := da.StreamArchive(context.Background(), filePath, func(record map[string]interface{}) error {
		return assert.AnError
	})
	assert.Error(t, err)
}

func TestArchiveData_InvalidPath(t *testing.T) {
	cfg := DefaultArchiveConfig()
	cfg.StoragePath = "/nonexistent/path/that/does/not/exist"
	da := NewDataArchiver(cfg, nil, zap.NewNop())

	records := []map[string]interface{}{{"id": "1"}}
	_, err := da.ArchiveData(context.Background(), "metrics", records)
	assert.Error(t, err)
}

func TestArchivePolicy_JSON(t *testing.T) {
	policy := ArchivePolicy{
		ID:            "p1",
		Name:          "Test",
		DataType:      "metrics",
		RetentionDays: 30,
		Format:        ArchiveFormatJSON,
		Enabled:       true,
	}

	data, err := json.Marshal(policy)
	require.NoError(t, err)

	var parsed ArchivePolicy
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "p1", parsed.ID)
	assert.Equal(t, ArchiveFormatJSON, parsed.Format)
}
