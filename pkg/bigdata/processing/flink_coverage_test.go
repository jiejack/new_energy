package processing

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/infrastructure/mq"
	"github.com/new-energy-monitoring/pkg/alarm/notifier"
	"github.com/new-energy-monitoring/pkg/ai/fault"
	"github.com/new-energy-monitoring/pkg/bigdata/types"
	"github.com/new-energy-monitoring/pkg/bigdata/visualization"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBasicProcessor_Process_TemperatureHigh(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "temperature", Value: 200, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	assert.NoError(t, err)
	assert.Equal(t, 25.0, result.DataPoints[0].Value)
}

func TestBasicProcessor_Process_VoltageHigh(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "voltage", Value: 15000, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, result.DataPoints[0].Value)
}

func TestBasicProcessor_Process_PowerNoTransform(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "power", Value: 500, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	assert.NoError(t, err)
	assert.Equal(t, 500.0, result.DataPoints[0].Value)
	assert.Equal(t, "power", result.DataPoints[0].Metric)
}

func TestBasicProcessor_Process_EnergyNoTransform(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "energy", Value: 500, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	assert.NoError(t, err)
	assert.Equal(t, 500.0, result.DataPoints[0].Value)
	assert.Equal(t, "energy", result.DataPoints[0].Metric)
}

func TestBasicProcessor_Process_CurrentNormal(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "current", Value: 100, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, result.DataPoints[0].Value)
}

func TestBasicProcessor_Process_EnhanceData_Features(t *testing.T) {
	p := NewBasicProcessor()
	now := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "power", Value: 100, Timestamp: now},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: now, RecordCount: 1},
	}
	result, err := p.Process(bd)
	assert.NoError(t, err)
	assert.Equal(t, 14, result.DataPoints[0].Attributes["hour"])
	assert.Equal(t, time.Saturday, result.DataPoints[0].Attributes["day_of_week"])
	assert.Equal(t, time.June, result.DataPoints[0].Attributes["month"])
	assert.Equal(t, true, result.DataPoints[0].Attributes["is_weekend"])
}

func TestBasicProcessor_Process_EnhanceData_Weekday(t *testing.T) {
	p := NewBasicProcessor()
	now := time.Date(2024, 6, 13, 10, 0, 0, 0, time.UTC)
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "power", Value: 100, Timestamp: now},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: now, RecordCount: 1},
	}
	result, err := p.Process(bd)
	assert.NoError(t, err)
	assert.Equal(t, false, result.DataPoints[0].Attributes["is_weekend"])
}

func TestFlinkProcessor_WriteToKafka_NilProducer(t *testing.T) {
	f := NewFlinkProcessor()

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()},
		},
	}
	err := f.writeToKafka(bd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kafka producer not initialized")
}

func TestFlinkProcessor_StartKafkaConsumer_NilConsumer(t *testing.T) {
	f := NewFlinkProcessor()
	f.kafkaConsumer = nil
	f.startKafkaConsumer()
}

func TestFlinkProcessor_UpdateVisualization_Disabled(t *testing.T) {
	f := NewFlinkProcessor()
	f.visualizationEnabled = false

	stats := map[string]interface{}{"anomalies": 0}
	dps := []*types.DataPoint{{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()}}
	f.updateVisualization(stats, dps)
}

func TestFlinkProcessor_CreateDefaultDashboard_Enabled(t *testing.T) {
	f := NewFlinkProcessor()
	f.visualizationEnabled = true
	f.visualizer = nil
	err := f.createDefaultDashboard()
	assert.NoError(t, err)
}

func TestFlinkProcessor_DetectAnomalies_NoDetectors(t *testing.T) {
	f := NewFlinkProcessor()
	f.anomalyDetectors = map[string]fault.FaultDetector{}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()}
	anomalies := f.detectAnomalies(dp)
	assert.Nil(t, anomalies)
}

func TestFlinkProcessor_DetectAnomalies_NoMatchedDetector(t *testing.T) {
	f := NewFlinkProcessor()

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "unknown_metric", Value: 9999.0, Timestamp: time.Now()}
	anomalies := f.detectAnomalies(dp)
	assert.Nil(t, anomalies)
}

func TestFlinkProcessor_GetWindowAggregations_NoWindow(t *testing.T) {
	f := NewFlinkProcessor()

	agg := f.GetWindowAggregations("dev1", "nonexistent_metric")
	assert.Nil(t, agg)
}

func TestFlinkProcessor_GetWindowAggregations_NoMetricInWindow(t *testing.T) {
	f := NewFlinkProcessor()
	f.windows = map[string]*WindowState{
		"dev1_temp": {Aggregations: map[string]map[AggregationType]float64{"temp": {}}},
	}

	agg := f.GetWindowAggregations("dev1", "voltage")
	assert.Nil(t, agg)
}

func TestFlinkProcessor_StopJob_EmptyJobID(t *testing.T) {
	f := NewFlinkProcessor()
	f.isRunning = true
	f.jobID = ""

	err := f.StopJob()
	assert.NoError(t, err)
}

func TestFlinkProcessor_Close_ClearsWindows(t *testing.T) {
	f := NewFlinkProcessor()
	f.started = true
	f.isRunning = true
	f.jobID = "test-job"
	f.windows = map[string]*WindowState{
		"w1": {},
	}

	err := f.Close()
	assert.NoError(t, err)
	assert.Empty(t, f.windows)
}

func TestFlinkProcessor_Process_StreamMode_Default(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.enabled = true
	f.isRunning = true

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()},
		},
	}
	result, err := f.Process(bd)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "flink_stream", result.Metadata.Source)
}

func TestFlinkProcessor_ProcessDataPoint_AnomalyDetection(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.operators["anomaly_detection"] = true

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0, "windowed": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temperature", Value: 50.0, Timestamp: time.Now()}
	result := f.processDataPoint(dp, "stream", &stats)
	assert.NotNil(t, result)
}

func TestFlinkProcessor_ProcessDataPoint_QualityValidationEnabled(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.operators["quality_validation"] = true
	f.operators["data_cleaning"] = false

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: -5.0, Timestamp: time.Now()}
	result := f.processDataPoint(dp, "stream", &stats)
	assert.NotNil(t, result)
	assert.Equal(t, 1, stats["cleaned"])
}

func TestFlinkProcessor_CalculateAggregations_MultipleMetrics(t *testing.T) {
	f := NewFlinkProcessor()

	window := &WindowState{
		DataPoints:   make([]*types.DataPoint, 0),
		Aggregations: make(map[string]map[AggregationType]float64),
	}

	f.calculateAggregations(window, &types.DataPoint{Metric: "temp", Value: 20.0, Timestamp: time.Now()})
	f.calculateAggregations(window, &types.DataPoint{Metric: "humidity", Value: 50.0, Timestamp: time.Now()})

	assert.NotNil(t, window.Aggregations["temp"])
	assert.NotNil(t, window.Aggregations["humidity"])
	assert.Equal(t, 20.0, window.Aggregations["temp"][AggMin])
	assert.Equal(t, 50.0, window.Aggregations["humidity"][AggMin])
}

func TestFlinkProcessor_UpdateWindows_MultipleMetricsSameWindow(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	stats := map[string]int{"total": 2, "cleaned": 0, "invalid": 0, "windowed": 0, "anomalies": 0}

	now := time.Now()
	f.updateWindows(&types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: now}, &stats)
	f.updateWindows(&types.DataPoint{DeviceID: "dev1", Metric: "humidity", Value: 60.0, Timestamp: now}, &stats)

	assert.Equal(t, 2, stats["windowed"])

	tempAgg := f.GetWindowAggregations("dev1", "temp")
	humidityAgg := f.GetWindowAggregations("dev1", "humidity")
	assert.NotNil(t, tempAgg)
	assert.NotNil(t, humidityAgg)
}

func TestFilterOperator_Struct(t *testing.T) {
	op := FilterOperator{
		Condition: func(dp *types.DataPoint) bool { return dp.Value > 0 },
	}
	assert.NotNil(t, op.Condition)
}

func TestMapOperator_Struct(t *testing.T) {
	op := MapOperator{
		Transform: func(dp *types.DataPoint) *types.DataPoint { return dp },
	}
	assert.NotNil(t, op.Transform)
}

func TestAggregationOperator_Struct(t *testing.T) {
	op := AggregationOperator{
		GroupBy:    "device_id",
		AggType:    AggAvg,
		MetricName: "temperature",
	}
	assert.Equal(t, "device_id", op.GroupBy)
	assert.Equal(t, AggAvg, op.AggType)
}

func TestWindowState_Struct(t *testing.T) {
	now := time.Now()
	ws := WindowState{
		DataPoints:  []*types.DataPoint{},
		StartTime:   now,
		EndTime:     now.Add(time.Hour),
		Aggregations: map[string]map[AggregationType]float64{},
	}
	assert.True(t, now.Before(ws.EndTime))
}

func TestBasicProcessor_Process_MetadataPreservation(t *testing.T) {
	p := NewBasicProcessor()
	now := time.Now()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "power", Value: 100, Timestamp: now},
		},
		Metadata: types.Metadata{
			Source:      "sensor-network",
			BatchID:     "batch-123",
			Timestamp:   now,
			RecordCount: 1,
			Properties: map[string]interface{}{"region": "north"},
		},
	}
	result, err := p.Process(bd)
	assert.NoError(t, err)
	assert.Contains(t, result.Metadata.BatchID, "-processed")
	assert.Equal(t, 1, result.Metadata.RecordCount)
	assert.Equal(t, 1, result.Metadata.Properties["original_count"])
}

func TestBasicProcessor_Process_ZeroValue(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "power", Value: 0, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, result.DataPoints[0].Value)
}

func TestFlinkProcessor_Init_CustomParallelism(t *testing.T) {
	f := NewFlinkProcessor()
	err := f.Init(types.ProcessingConfig{
		Type:        "flink",
		Parallelism: 16,
	})
	assert.NoError(t, err)
	assert.Equal(t, 16, f.parallelism)
	f.Close()
}

func TestFlinkProcessor_Init_CustomWindowSizeAndSlide(t *testing.T) {
	f := NewFlinkProcessor()
	err := f.Init(types.ProcessingConfig{
		Type:       "flink",
		WindowSize: "300",
		SlideSize:  "150",
	})
	assert.NoError(t, err)
	assert.Equal(t, 300*time.Second, f.windowSize)
	assert.Equal(t, 150*time.Second, f.slideSize)
	f.Close()
}

func TestFlinkProcessor_ProcessDataPoint_SetsTags(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()}
	result := f.processDataPoint(dp, "batch", &stats)
	assert.NotNil(t, result)
	assert.Equal(t, "flink", result.Tags["processor"])
	assert.Equal(t, "batch", result.Tags["mode"])
}

func TestFlinkProcessor_GetStats_AfterInit(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.enabled = true
	f.isRunning = true
	f.jobName = "test-flink-job"

	stats := f.GetStats()
	assert.Equal(t, TumblingWindow, stats["window_type"])
	assert.Equal(t, float64(defaultWindowSize), stats["window_size"])
	assert.Equal(t, defaultParallelism, stats["parallelism"])
}

func TestFlinkProcessor_SendAlert_WithNotifier(t *testing.T) {
	f := NewFlinkProcessor()
	f.alertNotifier = nil

	notification := &notifier.Notification{
		ID:       "test-alert",
		Subject:  "Test Subject",
		Content:  "Test Content",
		Priority: notifier.PriorityNormal,
	}
	f.sendAlert(notification)
}

func TestBasicProcessor_Process_UnknownDeviceType(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "XYZ-001", Metric: "power", Value: 100, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	assert.NoError(t, err)
	assert.Equal(t, "unknown", result.DataPoints[0].Tags["device_type"])
}

func TestBasicProcessor_CalculateDataQuality_UnknownMetric(t *testing.T) {
	p := NewBasicProcessor()
	dp := &types.DataPoint{Metric: "unknown_metric", Value: 42.0, Timestamp: time.Now()}
	quality := p.calculateDataQuality(dp)
	assert.Equal(t, 1.0, quality)
}

func TestFlinkProcessor_ContextCancelled(t *testing.T) {
	f := NewFlinkProcessor()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f.ctx = ctx
	f.cancel = cancel
	f.Close()
}

func TestFlinkProcessor_StartFlinkJob_WithOperators(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.operators["data_cleaning"] = true
	f.operators["anomaly_detection"] = true
	f.sources = []map[string]interface{}{{"type": "kafka", "topic": "input", "consumer_group": "cg1"}}
	f.sinks = []map[string]interface{}{{"type": "kafka", "topic": "output"}}

	err := f.startFlinkJob()
	assert.NoError(t, err)
	assert.True(t, f.isRunning)
	assert.NotEmpty(t, f.jobID)
}

func TestFlinkProcessor_Process_BatchModeWithDataCleaning(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "batch"})
	defer f.Close()

	f.enabled = true
	f.isRunning = true
	f.operators["data_cleaning"] = true

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: -5.0, Timestamp: time.Now()},
			{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()},
		},
	}
	result, err := f.Process(bd)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "flink_batch", result.Metadata.Source)
}

func TestFlinkProcessor_Process_StreamModeWithDataCleaning(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.enabled = true
	f.isRunning = true
	f.operators["data_cleaning"] = true

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: -5.0, Timestamp: time.Now()},
			{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()},
		},
	}
	result, err := f.Process(bd)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "flink_stream", result.Metadata.Source)
}

func TestFlinkProcessor_ProcessDataPoint_AnomalyDetectionWithDetector(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.operators["anomaly_detection"] = true

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temperature", Value: 50.0, Timestamp: time.Now()}
	result := f.processDataPoint(dp, "stream", &stats)
	assert.NotNil(t, result)
}

func TestFlinkProcessor_ProcessDataPoint_NoAnomaly(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.operators["anomaly_detection"] = true

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "humidity", Value: 50.0, Timestamp: time.Now()}
	result := f.processDataPoint(dp, "batch", &stats)
	assert.NotNil(t, result)
	assert.Equal(t, false, result.Attributes["has_anomaly"])
}

func TestFlinkProcessor_ProcessDataPoint_DataCleaningPositive(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.operators["data_cleaning"] = true

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()}
	result := f.processDataPoint(dp, "stream", &stats)
	assert.NotNil(t, result)
	assert.Equal(t, 25.0, result.Value)
}

func TestFlinkProcessor_ProcessDataPoint_InfValueWithCleaning(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.operators["data_cleaning"] = true

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: math.Inf(1), Timestamp: time.Now()}
	result := f.processDataPoint(dp, "stream", &stats)
	assert.Nil(t, result)
}

func TestFlinkProcessor_DetectAnomalies_WithDetector(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temperature", Value: 50.0, Timestamp: time.Now()}
	_ = f.detectAnomalies(dp)
}

func TestFlinkProcessor_DetectAnomalies_UnmatchedMetric(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "custom_metric", Value: 50.0, Timestamp: time.Now()}
	anomalies := f.detectAnomalies(dp)
	assert.Nil(t, anomalies)
}

func TestFlinkProcessor_WriteToKafka_NilProducer_Error(t *testing.T) {
	f := NewFlinkProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()},
		},
	}
	err := f.writeToKafka(bd)
	assert.Error(t, err)
}

func TestFlinkProcessor_StartKafkaConsumer_Nil(t *testing.T) {
	f := NewFlinkProcessor()
	f.kafkaConsumer = nil
	f.startKafkaConsumer()
}

func TestFlinkProcessor_UpdateVisualization_Disabled_NilVisualizer(t *testing.T) {
	f := NewFlinkProcessor()
	f.visualizationEnabled = true
	f.visualizer = nil

	stats := map[string]interface{}{"anomalies": 0}
	dps := []*types.DataPoint{{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()}}
	f.updateVisualization(stats, dps)
}

func TestFlinkProcessor_Process_EmptyData(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.enabled = true
	f.isRunning = true

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{},
	}
	result, err := f.Process(bd)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestFlinkProcessor_Process_NilData(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.enabled = false

	result, err := f.Process(nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestFlinkProcessor_Process_BatchModeEmptyData(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "batch"})
	defer f.Close()

	f.enabled = true
	f.isRunning = true

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{},
	}
	result, err := f.Process(bd)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestFlinkProcessor_Close_WithVisualizer(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	f.started = true
	f.visualizer = nil

	err := f.Close()
	assert.NoError(t, err)
}

func TestFlinkProcessor_InitKafka_NoKafkaSinks(t *testing.T) {
	f := NewFlinkProcessor()
	f.sinks = []map[string]interface{}{
		{"type": "file", "path": "/tmp/output"},
	}
	f.sources = []map[string]interface{}{
		{"type": "file", "path": "/tmp/input"},
	}
	err := f.initKafka()
	assert.NoError(t, err)
}

func TestFlinkProcessor_LoadFlinkConfig(t *testing.T) {
	f := NewFlinkProcessor()
	f.loadFlinkConfig()
}

func TestFlinkProcessor_LoadKafkaConfig(t *testing.T) {
	f := NewFlinkProcessor()
	f.loadKafkaConfig()
}

func TestFlinkProcessor_InitAlerting(t *testing.T) {
	f := NewFlinkProcessor()
	f.initAlerting()
	time.Sleep(10 * time.Millisecond)
}

func TestFlinkProcessor_InitVisualization(t *testing.T) {
	f := NewFlinkProcessor()
	f.initVisualization()
}

func TestFlinkProcessor_InitAnomalyDetection(t *testing.T) {
	f := NewFlinkProcessor()
	f.initAnomalyDetection()
	assert.Equal(t, 6, len(f.anomalyDetectors))
}

func TestFlinkProcessor_ProcessDataPoint_AllOperators(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.operators["data_cleaning"] = true
	f.operators["quality_validation"] = true
	f.operators["anomaly_detection"] = true

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temperature", Value: 25.0, Timestamp: time.Now()}
	result := f.processDataPoint(dp, "stream", &stats)
	assert.NotNil(t, result)
	assert.Equal(t, true, result.Attributes["processed"])
	assert.Equal(t, "flink", result.Attributes["processor"])
	assert.Equal(t, "stream", result.Attributes["mode"])
}

func TestFlinkProcessor_StreamProcessing_WithWindowUpdates(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.enabled = true
	f.isRunning = true

	now := time.Now()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: now},
			{DeviceID: "dev1", Metric: "temp", Value: 30.0, Timestamp: now.Add(time.Minute)},
			{DeviceID: "dev2", Metric: "temp", Value: 20.0, Timestamp: now},
		},
	}
	result, err := f.Process(bd)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.DataPoints, 3)

	agg := f.GetWindowAggregations("dev1", "temp")
	assert.NotNil(t, agg)
	assert.Equal(t, 25.0, agg[AggMin])
	assert.Equal(t, 30.0, agg[AggMax])
}

func TestFlinkProcessor_BatchProcessing_WithOperators(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "batch"})
	defer f.Close()

	f.enabled = true
	f.isRunning = true
	f.operators["data_cleaning"] = true
	f.operators["anomaly_detection"] = true

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temperature", Value: 25.0, Timestamp: time.Now()},
			{DeviceID: "dev1", Metric: "temperature", Value: -5.0, Timestamp: time.Now()},
		},
	}
	result, err := f.Process(bd)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestFlinkProcessor_GetStats_AfterProcessing(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.enabled = true
	f.isRunning = true
	f.jobName = "test-job"

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()},
		},
	}
	f.Process(bd)

	stats := f.GetStats()
	assert.Equal(t, "test-job", stats["job_name"])
	assert.True(t, stats["is_running"].(bool))
}

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) Channel() notifier.NotificationChannel {
	args := m.Called()
	return args.Get(0).(notifier.NotificationChannel)
}

func (m *MockNotifier) Send(ctx context.Context, notification *notifier.Notification) (*notifier.NotificationResult, error) {
	args := m.Called(ctx, notification)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*notifier.NotificationResult), args.Error(1)
}

func (m *MockNotifier) SendBatch(ctx context.Context, notifications []*notifier.Notification) ([]*notifier.NotificationResult, error) {
	args := m.Called(ctx, notifications)
	return args.Get(0).([]*notifier.NotificationResult), args.Error(1)
}

func (m *MockNotifier) Validate(notification *notifier.Notification) error {
	args := m.Called(notification)
	return args.Error(0)
}

func (m *MockNotifier) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockNotifier) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestFlinkProcessor_SendAlert_WithMockNotifier(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	mockNotifier := new(MockNotifier)
	mockNotifier.On("Send", f.ctx, mock.AnythingOfType("*notifier.Notification")).Return(
		&notifier.NotificationResult{Status: "sent"}, nil,
	)
	f.alertNotifier = mockNotifier

	notification := &notifier.Notification{
		ID:       "test-alert",
		Subject:  "Test Subject",
		Content:  "Test Content",
		Priority: notifier.PriorityNormal,
	}
	f.sendAlert(notification)
	mockNotifier.AssertCalled(t, "Send", f.ctx, notification)
}

func TestFlinkProcessor_SendAlert_WithMockNotifier_Error(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	mockNotifier := new(MockNotifier)
	mockNotifier.On("Send", f.ctx, mock.AnythingOfType("*notifier.Notification")).Return(
		nil, fmt.Errorf("send error"),
	)
	f.alertNotifier = mockNotifier

	notification := &notifier.Notification{
		ID:       "test-alert",
		Subject:  "Error Subject",
		Content:  "Error Content",
		Priority: notifier.PriorityHigh,
	}
	f.sendAlert(notification)
	mockNotifier.AssertCalled(t, "Send", f.ctx, notification)
}

func TestFlinkProcessor_UpdateVisualization_WithVisualizer(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	vis := visualization.NewBasicVisualizer()
	vis.Init(types.VisualizationConfig{Type: "basic", Host: "localhost", Port: 8080})
	vis.CreateDashboard("flink-processing", []types.Panel{
		{ID: "flink-metrics", Title: "Metrics", Type: "metrics", Data: map[string]interface{}{}, Options: map[string]interface{}{}},
		{ID: "anomaly-detection", Title: "Anomaly", Type: "anomaly", Data: map[string]interface{}{}, Options: map[string]interface{}{}},
		{ID: "window-aggregations", Title: "Window", Type: "aggregations", Data: map[string]interface{}{}, Options: map[string]interface{}{}},
	})
	f.visualizer = vis
	f.visualizationEnabled = true

	stats := map[string]interface{}{"anomalies": 0, "cleaned": 5}
	dps := []*types.DataPoint{{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()}}
	f.updateVisualization(stats, dps)
}

func TestFlinkProcessor_UpdateVisualization_EmptyDataPoints(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	vis := visualization.NewBasicVisualizer()
	vis.Init(types.VisualizationConfig{Type: "basic", Host: "localhost", Port: 8080})
	vis.CreateDashboard("flink-processing", []types.Panel{
		{ID: "flink-metrics", Title: "Metrics", Type: "metrics", Data: map[string]interface{}{}, Options: map[string]interface{}{}},
		{ID: "anomaly-detection", Title: "Anomaly", Type: "anomaly", Data: map[string]interface{}{}, Options: map[string]interface{}{}},
		{ID: "window-aggregations", Title: "Window", Type: "aggregations", Data: map[string]interface{}{}, Options: map[string]interface{}{}},
	})
	f.visualizer = vis
	f.visualizationEnabled = true

	stats := map[string]interface{}{"anomalies": 0}
	f.updateVisualization(stats, nil)
}

func TestFlinkProcessor_InitVisualization_WithViper(t *testing.T) {
	viper.Set("flink.visualization", map[string]interface{}{
		"enabled": true,
	})
	defer viper.Reset()

	f := NewFlinkProcessor()
	f.initVisualization()
	assert.True(t, f.visualizationEnabled)
	assert.NotNil(t, f.visualizer)
}

func TestFlinkProcessor_InitVisualization_DisabledInViper(t *testing.T) {
	viper.Set("flink.visualization", map[string]interface{}{
		"enabled": false,
	})
	defer viper.Reset()

	f := NewFlinkProcessor()
	f.initVisualization()
	assert.False(t, f.visualizationEnabled)
}

func TestFlinkProcessor_CreateDefaultDashboard_WithVisualizer(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	vis := visualization.NewBasicVisualizer()
	vis.Init(types.VisualizationConfig{Type: "basic", Host: "localhost", Port: 8080})
	f.visualizer = vis
	f.visualizationEnabled = true

	err := f.createDefaultDashboard()
	assert.NoError(t, err)
}

func TestFlinkProcessor_LoadFlinkConfig_WithViper(t *testing.T) {
	viper.Set("flink.enabled", true)
	viper.Set("flink.job_name", "test-viper-job")
	viper.Set("flink.parallelism", 8)
	viper.Set("flink.window.type", "sliding")
	viper.Set("flink.window.size", 120)
	viper.Set("flink.window.slide", 60)
	viper.Set("flink.operators", []interface{}{
		map[string]interface{}{"name": "data_cleaning", "enabled": true},
		map[string]interface{}{"name": "anomaly_detection", "enabled": false},
	})
	viper.Set("flink.sinks", []map[string]interface{}{{"type": "kafka", "topic": "output"}})
	viper.Set("flink.sources", []map[string]interface{}{{"type": "kafka", "topic": "input"}})
	viper.Set("flink.metrics", map[string]interface{}{"reporter": "prometheus"})
	viper.Set("flink.checkpoint", map[string]interface{}{"interval": 1000})
	defer viper.Reset()

	f := NewFlinkProcessor()
	f.loadFlinkConfig()
}

func TestFlinkProcessor_LoadFlinkConfig_SessionWindow(t *testing.T) {
	viper.Set("flink.enabled", true)
	viper.Set("flink.window.type", "session")
	defer viper.Reset()

	f := NewFlinkProcessor()
	f.loadFlinkConfig()
}

func TestFlinkProcessor_InitKafka_WithKafkaSink(t *testing.T) {
	f := NewFlinkProcessor()
	f.kafkaConfig = mq.KafkaConfig{Brokers: []string{}, TopicPrefix: "test"}
	f.sinks = []map[string]interface{}{
		{"type": "kafka", "topic": "output-topic"},
	}
	f.sources = []map[string]interface{}{}
	err := f.initKafka()
	assert.NoError(t, err)
	assert.NotNil(t, f.kafkaProducer)
}

func TestFlinkProcessor_InitAlerting_WithViper(t *testing.T) {
	viper.Set("flink.alerting", map[string]interface{}{
		"enabled": true,
		"channels": []string{"email", "sms"},
	})
	defer viper.Reset()

	f := NewFlinkProcessor()
	f.initAlerting()
	time.Sleep(10 * time.Millisecond)
}

func TestFlinkProcessor_DetectAnomalies_WithValueOutOfRange(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	dp := &types.DataPoint{DeviceID: "default", Metric: "temperature", Value: 200.0, Timestamp: time.Now()}
	anomalies := f.detectAnomalies(dp)
	assert.NotNil(t, anomalies)
	assert.GreaterOrEqual(t, len(anomalies), 1)
}

func TestFlinkProcessor_DetectAnomalies_AllMetrics(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	metrics := []string{"temperature", "humidity", "pressure", "current", "voltage", "power"}
	for _, metric := range metrics {
		dp := &types.DataPoint{DeviceID: "default", Metric: metric, Value: 200.0, Timestamp: time.Now()}
		anomalies := f.detectAnomalies(dp)
		assert.NotNil(t, anomalies, "Expected anomalies for metric: %s", metric)
		assert.GreaterOrEqual(t, len(anomalies), 1, "Expected at least 1 anomaly for metric: %s", metric)
	}
}

func TestFlinkProcessor_ProcessDataPoint_AnomalyDetected(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.operators["anomaly_detection"] = true

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "default", Metric: "temperature", Value: 200.0, Timestamp: time.Now()}
	result := f.processDataPoint(dp, "stream", &stats)
	assert.NotNil(t, result)
	assert.Equal(t, true, result.Attributes["has_anomaly"])
	assert.Equal(t, 1, stats["anomalies"])
}

func TestFlinkProcessor_ProcessAlerts_ContextCancelled(t *testing.T) {
	f := NewFlinkProcessor()
	ctx, cancel := context.WithCancel(context.Background())
	f.ctx = ctx
	f.cancel = cancel

	cancel()
	f.processAlerts()
}

func TestFlinkProcessor_Close_WithNotifier(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	f.started = true
	f.alertNotifier = nil

	err := f.Close()
	assert.NoError(t, err)
}

func TestFlinkProcessor_Close_WithKafkaProducer(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	f.started = true
	f.kafkaProducer = nil
	f.kafkaConsumer = nil

	err := f.Close()
	assert.NoError(t, err)
}

func TestFlinkProcessor_InitAnomalyDetection_WithViper(t *testing.T) {
	viper.Set("flink.anomaly_detection", map[string]interface{}{
		"enabled": true,
		"method":  "threshold",
	})
	defer viper.Reset()

	f := NewFlinkProcessor()
	f.initAnomalyDetection()
	assert.Equal(t, 6, len(f.anomalyDetectors))
}
