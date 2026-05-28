package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	e := &Error{Code: ErrCodeInvalidConfig, Message: "test error"}
	assert.Equal(t, "test error", e.Error())
}

func TestAcquireDataPoint(t *testing.T) {
	dp := AcquireDataPoint()
	assert.NotNil(t, dp)
	assert.NotNil(t, dp.Tags)
	assert.NotNil(t, dp.Attributes)
}

func TestReleaseDataPoint_Nil(t *testing.T) {
	ReleaseDataPoint(nil)
}

func TestReleaseDataPoint_Valid(t *testing.T) {
	dp := AcquireDataPoint()
	dp.Timestamp = time.Now()
	dp.DeviceID = "dev1"
	dp.Metric = "temp"
	dp.Value = 25.5
	dp.Tags["key"] = "value"
	dp.Attributes["attr"] = "val"
	ReleaseDataPoint(dp)
}

func TestAcquireBatchData(t *testing.T) {
	bd := AcquireBatchData()
	assert.NotNil(t, bd)
	assert.NotNil(t, bd.DataPoints)
	assert.NotNil(t, bd.Metadata.Properties)
}

func TestReleaseBatchData_Nil(t *testing.T) {
	ReleaseBatchData(nil)
}

func TestReleaseBatchData_Valid(t *testing.T) {
	bd := AcquireBatchData()
	bd.Metadata.Source = "test"
	bd.Metadata.BatchID = "batch-1"
	bd.Metadata.Timestamp = time.Now()
	bd.Metadata.RecordCount = 5
	bd.Metadata.Properties["key"] = "value"
	bd.DataPoints = append(bd.DataPoints, AcquireDataPoint())
	ReleaseBatchData(bd)
}

func TestDataPoint_Fields(t *testing.T) {
	dp := &DataPoint{
		Timestamp:  time.Now(),
		DeviceID:   "dev1",
		Metric:     "temperature",
		Value:      25.5,
		Tags:       map[string]string{"location": "roof"},
		Attributes: map[string]interface{}{"quality": 0.95},
	}
	assert.Equal(t, "dev1", dp.DeviceID)
	assert.Equal(t, "temperature", dp.Metric)
	assert.InDelta(t, 25.5, dp.Value, 0.001)
}

func TestBatchData_Fields(t *testing.T) {
	bd := &BatchData{
		DataPoints: []*DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
		},
		Metadata: Metadata{
			Source:      "test",
			BatchID:     "b1",
			Timestamp:   time.Now(),
			RecordCount: 1,
			Properties:  map[string]interface{}{"test": true},
		},
	}
	assert.Len(t, bd.DataPoints, 1)
	assert.Equal(t, "test", bd.Metadata.Source)
}

func TestStorageConfig_Fields(t *testing.T) {
	cfg := StorageConfig{
		Type:          "clickhouse",
		Host:          "localhost",
		Port:          9000,
		Database:      "test_db",
		Table:         "test_table",
		Username:      "default",
		Password:      "secret",
		BatchSize:     1000,
		FlushInterval: 5,
		Options:       map[string]interface{}{"secure": true},
	}
	assert.Equal(t, "clickhouse", cfg.Type)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 9000, cfg.Port)
}

func TestAnalysisConfig_Fields(t *testing.T) {
	cfg := AnalysisConfig{
		Type:     "spark",
		Master:   "local",
		AppName:  "test",
		Executor: 4,
		Memory:   "2g",
	}
	assert.Equal(t, "spark", cfg.Type)
}

func TestVisualizationConfig_Fields(t *testing.T) {
	cfg := VisualizationConfig{
		Type:   "grafana",
		Host:   "localhost",
		Port:   3000,
		APIKey: "key",
	}
	assert.Equal(t, "grafana", cfg.Type)
}

func TestProcessingConfig_Fields(t *testing.T) {
	cfg := ProcessingConfig{
		Type:        "stream",
		WindowSize:  "60s",
		SlideSize:   "30s",
		Parallelism: 4,
	}
	assert.Equal(t, "stream", cfg.Type)
}

func TestIngestionConfig_Fields(t *testing.T) {
	cfg := IngestionConfig{
		Type:       "kafka",
		Topic:      "test-topic",
		Broker:     "localhost:9092",
		ConsumerID: "consumer-1",
		BatchSize:  100,
	}
	assert.Equal(t, "kafka", cfg.Type)
}

func TestPanel_Fields(t *testing.T) {
	p := Panel{
		ID:      "panel-1",
		Title:   "Test Panel",
		Type:    "graph",
		Data:    map[string]interface{}{"value": 42},
		Options: map[string]interface{}{"refresh": 10},
	}
	assert.Equal(t, "panel-1", p.ID)
	assert.Equal(t, "graph", p.Type)
}

func TestErrorCodes(t *testing.T) {
	assert.Equal(t, "INVALID_CONFIG", ErrCodeInvalidConfig)
	assert.Equal(t, "STORAGE_ERROR", ErrCodeStorageError)
	assert.Equal(t, "ANALYSIS_ERROR", ErrCodeAnalysisError)
	assert.Equal(t, "VISUALIZATION_ERROR", ErrCodeVisualizationError)
	assert.Equal(t, "PROCESSING_ERROR", ErrCodeProcessingError)
	assert.Equal(t, "INGESTION_ERROR", ErrCodeIngestionError)
}
