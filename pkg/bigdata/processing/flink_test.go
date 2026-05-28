package processing

import (
	"math"
	"testing"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
	"github.com/stretchr/testify/assert"
)

func TestFlinkProcessor_New(t *testing.T) {
	f := NewFlinkProcessor()
	assert.NotNil(t, f)
	assert.False(t, f.enabled)
	assert.False(t, f.isRunning)
	assert.False(t, f.started)
}

func TestFlinkProcessor_Init(t *testing.T) {
	f := NewFlinkProcessor()
	err := f.Init(types.ProcessingConfig{
		Type:        "flink",
		Parallelism: 8,
		WindowSize:  "120",
		SlideSize:   "60",
	})
	assert.NoError(t, err)
	assert.Equal(t, 8, f.parallelism)
	assert.Equal(t, 120*time.Second, f.windowSize)
	assert.Equal(t, 60*time.Second, f.slideSize)
	assert.True(t, f.started)
	f.Close()
}

func TestFlinkProcessor_Init_Defaults(t *testing.T) {
	f := NewFlinkProcessor()
	err := f.Init(types.ProcessingConfig{Type: "flink"})
	assert.NoError(t, err)
	assert.Equal(t, defaultParallelism, f.parallelism)
	assert.Equal(t, time.Duration(defaultWindowSize)*time.Second, f.windowSize)
	assert.Equal(t, time.Duration(defaultSlideSize)*time.Second, f.slideSize)
	f.Close()
}

func TestFlinkProcessor_Process_Disabled(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()},
		},
	}
	result, err := f.Process(bd)
	assert.NoError(t, err)
	assert.Equal(t, bd, result)
}

func TestFlinkProcessor_Process_Enabled(t *testing.T) {
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
}

func TestFlinkProcessor_Process_BatchMode(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "batch"})
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
	assert.Equal(t, "flink_batch", result.Metadata.Source)
}

func TestFlinkProcessor_Process_Enabled_StartJob(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.enabled = true

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()},
		},
	}
	result, err := f.Process(bd)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, f.isRunning)
}

func TestFlinkProcessor_ProcessDataPoint(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()}
	result := f.processDataPoint(dp, "stream", &stats)
	assert.NotNil(t, result)
	assert.Equal(t, true, result.Attributes["processed"])
	assert.Equal(t, "flink", result.Attributes["processor"])
}

func TestFlinkProcessor_ProcessDataPoint_NilAttributes(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()}
	result := f.processDataPoint(dp, "batch", &stats)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Attributes)
	assert.NotNil(t, result.Tags)
}

func TestFlinkProcessor_ProcessDataPoint_DataCleaning(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.operators["data_cleaning"] = true

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: -5.0, Timestamp: time.Now()}
	result := f.processDataPoint(dp, "stream", &stats)
	assert.NotNil(t, result)
	assert.Equal(t, 0.0, result.Value)
}

func TestFlinkProcessor_ProcessDataPoint_InvalidValue(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.operators["data_cleaning"] = true

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: math.NaN(), Timestamp: time.Now()}
	result := f.processDataPoint(dp, "stream", &stats)
	assert.Nil(t, result)
	assert.Equal(t, 1, stats["invalid"])
}

func TestFlinkProcessor_ProcessDataPoint_InfValue(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.operators["data_cleaning"] = true

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: math.Inf(1), Timestamp: time.Now()}
	result := f.processDataPoint(dp, "stream", &stats)
	assert.Nil(t, result)
}

func TestFlinkProcessor_ProcessDataPoint_QualityValidation(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.operators["quality_validation"] = true

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()}
	result := f.processDataPoint(dp, "stream", &stats)
	assert.NotNil(t, result)
}

func TestFlinkProcessor_UpdateWindows(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	stats := map[string]int{"total": 1, "cleaned": 0, "invalid": 0, "anomalies": 0, "windowed": 0}

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()}
	f.updateWindows(dp, &stats)
	assert.Equal(t, 1, stats["windowed"])

	agg := f.GetWindowAggregations("dev1", "temp")
	assert.NotNil(t, agg)
	assert.Equal(t, 25.0, agg[AggMin])
	assert.Equal(t, 25.0, agg[AggMax])
	assert.Equal(t, 25.0, agg[AggAvg])
}

func TestFlinkProcessor_UpdateWindows_MultiplePoints(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	stats := map[string]int{"total": 3, "cleaned": 0, "invalid": 0, "anomalies": 0, "windowed": 0}

	for _, val := range []float64{10, 20, 30} {
		dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: val, Timestamp: time.Now()}
		f.updateWindows(dp, &stats)
	}

	agg := f.GetWindowAggregations("dev1", "temp")
	assert.NotNil(t, agg)
	assert.Equal(t, 10.0, agg[AggMin])
	assert.Equal(t, 30.0, agg[AggMax])
	assert.Equal(t, 20.0, agg[AggAvg])
	assert.Equal(t, 60.0, agg[AggSum])
	assert.Equal(t, 3.0, agg[AggCount])
}

func TestFlinkProcessor_GetWindowAggregations_NotFound(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	agg := f.GetWindowAggregations("nonexistent", "temp")
	assert.Nil(t, agg)
}

func TestFlinkProcessor_GetStats(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	stats := f.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, "flink", stats["processor"])
	assert.Equal(t, false, stats["is_running"])
	assert.Equal(t, false, stats["enabled"])
}

func TestFlinkProcessor_StopJob(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.isRunning = true
	f.jobID = "test-job"

	err := f.StopJob()
	assert.NoError(t, err)
	assert.False(t, f.isRunning)
}

func TestFlinkProcessor_StopJob_NotRunning(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	err := f.StopJob()
	assert.NoError(t, err)
}

func TestFlinkProcessor_Close(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})

	err := f.Close()
	assert.NoError(t, err)
	assert.False(t, f.isRunning)
	assert.False(t, f.started)
}

func TestFlinkProcessor_Close_NotStarted(t *testing.T) {
	f := NewFlinkProcessor()
	err := f.Close()
	assert.NoError(t, err)
}

func TestFlinkProcessor_Close_WithRunningJob(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	f.isRunning = true
	f.jobID = "test-job"

	err := f.Close()
	assert.NoError(t, err)
}

func TestFlinkProcessor_CalculateAggregations(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	window := &WindowState{
		DataPoints:   make([]*types.DataPoint, 0),
		Aggregations: make(map[string]map[AggregationType]float64),
	}

	dp := &types.DataPoint{Metric: "temp", Value: 25.0, Timestamp: time.Now()}
	f.calculateAggregations(window, dp)

	assert.Equal(t, 25.0, window.Aggregations["temp"][AggMin])
	assert.Equal(t, 25.0, window.Aggregations["temp"][AggMax])
	assert.Equal(t, 25.0, window.Aggregations["temp"][AggAvg])
	assert.Equal(t, 1.0, window.Aggregations["temp"][AggCount])

	dp2 := &types.DataPoint{Metric: "temp", Value: 30.0, Timestamp: time.Now()}
	f.calculateAggregations(window, dp2)

	assert.Equal(t, 25.0, window.Aggregations["temp"][AggMin])
	assert.Equal(t, 30.0, window.Aggregations["temp"][AggMax])
	assert.Equal(t, 27.5, window.Aggregations["temp"][AggAvg])
	assert.Equal(t, 2.0, window.Aggregations["temp"][AggCount])
}

func TestFlinkProcessor_WindowTypes(t *testing.T) {
	assert.Equal(t, WindowType("tumbling"), TumblingWindow)
	assert.Equal(t, WindowType("sliding"), SlidingWindow)
	assert.Equal(t, WindowType("session"), SessionWindow)
}

func TestFlinkProcessor_AggregationTypes(t *testing.T) {
	assert.Equal(t, AggregationType("sum"), AggSum)
	assert.Equal(t, AggregationType("avg"), AggAvg)
	assert.Equal(t, AggregationType("min"), AggMin)
	assert.Equal(t, AggregationType("max"), AggMax)
	assert.Equal(t, AggregationType("count"), AggCount)
}

func TestFlinkProcessor_StreamProcessing(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	f.enabled = true
	f.isRunning = true

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()},
			{DeviceID: "dev1", Metric: "temp", Value: 30.0, Timestamp: time.Now()},
		},
	}
	result, err := f.Process(bd)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "flink_stream", result.Metadata.Source)
}

func TestFlinkProcessor_CreateDefaultDashboard_Disabled(t *testing.T) {
	f := NewFlinkProcessor()
	f.Init(types.ProcessingConfig{Type: "flink"})
	defer f.Close()

	err := f.createDefaultDashboard()
	assert.NoError(t, err)
}
