package analysis

import (
	"testing"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBasicAnalyzer(t *testing.T) {
	a := NewBasicAnalyzer()
	assert.NotNil(t, a)
}

func TestBasicAnalyzer_Init_Valid(t *testing.T) {
	a := NewBasicAnalyzer()
	err := a.Init(types.AnalysisConfig{Type: "basic"})
	require.NoError(t, err)
}

func TestBasicAnalyzer_Init_Invalid(t *testing.T) {
	a := NewBasicAnalyzer()
	err := a.Init(types.AnalysisConfig{Type: "invalid"})
	assert.Error(t, err)
}

func TestBasicAnalyzer_Execute(t *testing.T) {
	a := NewBasicAnalyzer()
	result, err := a.Execute("SELECT * FROM test")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestBasicAnalyzer_Process_Nil(t *testing.T) {
	a := NewBasicAnalyzer()
	_, err := a.Process(nil)
	assert.Error(t, err)
}

func TestBasicAnalyzer_Process_Empty(t *testing.T) {
	a := NewBasicAnalyzer()
	_, err := a.Process(&types.BatchData{DataPoints: []*types.DataPoint{}})
	assert.Error(t, err)
}

func TestBasicAnalyzer_Process_Valid(t *testing.T) {
	a := NewBasicAnalyzer()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now().Add(-2 * time.Hour)},
			{DeviceID: "dev1", Metric: "temp", Value: 30.0, Timestamp: time.Now().Add(-1 * time.Hour)},
			{DeviceID: "dev1", Metric: "temp", Value: 28.0, Timestamp: time.Now()},
			{DeviceID: "dev2", Metric: "voltage", Value: 220.0, Timestamp: time.Now().Add(-1 * time.Hour)},
			{DeviceID: "dev2", Metric: "voltage", Value: 225.0, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 5},
	}
	result, err := a.Process(bd)
	require.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.NotNil(t, resultMap["summary"])
	assert.NotNil(t, resultMap["grouped_statistics"])
	assert.NotNil(t, resultMap["time_series_analysis"])
}

func TestBasicAnalyzer_CalculateStats(t *testing.T) {
	a := NewBasicAnalyzer()
	points := []*types.DataPoint{
		{Value: 10.0, Timestamp: time.Now()},
		{Value: 20.0, Timestamp: time.Now()},
		{Value: 30.0, Timestamp: time.Now()},
	}
	stats := a.calculateStats(points)
	assert.Equal(t, 60.0, stats["sum"])
	assert.Equal(t, 20.0, stats["mean"])
	assert.Equal(t, 10.0, stats["min"])
	assert.Equal(t, 30.0, stats["max"])
	assert.Equal(t, 20.0, stats["median"])
}

func TestBasicAnalyzer_CalculateStats_Empty(t *testing.T) {
	a := NewBasicAnalyzer()
	stats := a.calculateStats([]*types.DataPoint{})
	assert.Nil(t, stats)
}

func TestBasicAnalyzer_GroupByDeviceAndMetric(t *testing.T) {
	a := NewBasicAnalyzer()
	points := []*types.DataPoint{
		{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now()},
		{DeviceID: "dev1", Metric: "temp", Value: 30.0, Timestamp: time.Now()},
		{DeviceID: "dev2", Metric: "voltage", Value: 220.0, Timestamp: time.Now()},
	}
	grouped := a.groupByDeviceAndMetric(points)
	assert.Contains(t, grouped, "dev1")
	assert.Contains(t, grouped, "dev2")
	assert.Contains(t, grouped["dev1"], "temp")
	assert.Contains(t, grouped["dev2"], "voltage")
	assert.Equal(t, 2.0, grouped["dev1"]["temp"]["count"])
	assert.Equal(t, 55.0, grouped["dev1"]["temp"]["sum"])
	assert.Equal(t, 27.5, grouped["dev1"]["temp"]["mean"])
}

func TestBasicAnalyzer_AnalyzeTimeSeries(t *testing.T) {
	a := NewBasicAnalyzer()
	points := []*types.DataPoint{
		{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: time.Now().Add(-4 * time.Hour)},
		{DeviceID: "dev1", Metric: "temp", Value: 26.0, Timestamp: time.Now().Add(-3 * time.Hour)},
		{DeviceID: "dev1", Metric: "temp", Value: 27.0, Timestamp: time.Now().Add(-2 * time.Hour)},
		{DeviceID: "dev1", Metric: "temp", Value: 28.0, Timestamp: time.Now().Add(-1 * time.Hour)},
		{DeviceID: "dev1", Metric: "temp", Value: 29.0, Timestamp: time.Now()},
	}
	result := a.analyzeTimeSeries(points)
	assert.NotNil(t, result)
	assert.Contains(t, result, "dev1")
}

func TestBasicAnalyzer_CalculateTrend(t *testing.T) {
	a := NewBasicAnalyzer()
	points := []*types.DataPoint{
		{Value: 10.0, Timestamp: time.Now()},
		{Value: 20.0, Timestamp: time.Now()},
		{Value: 30.0, Timestamp: time.Now()},
	}
	trend := a.calculateTrend(points)
	assert.Equal(t, 10.0, trend)
}

func TestBasicAnalyzer_CalculateTrend_SinglePoint(t *testing.T) {
	a := NewBasicAnalyzer()
	points := []*types.DataPoint{{Value: 10.0, Timestamp: time.Now()}}
	trend := a.calculateTrend(points)
	assert.Equal(t, 0.0, trend)
}

func TestBasicAnalyzer_CalculateVolatility(t *testing.T) {
	a := NewBasicAnalyzer()
	points := []*types.DataPoint{
		{Value: 100.0, Timestamp: time.Now()},
		{Value: 110.0, Timestamp: time.Now()},
		{Value: 105.0, Timestamp: time.Now()},
	}
	volatility := a.calculateVolatility(points)
	assert.True(t, volatility >= 0)
}

func TestBasicAnalyzer_CalculateVolatility_SinglePoint(t *testing.T) {
	a := NewBasicAnalyzer()
	points := []*types.DataPoint{{Value: 100.0, Timestamp: time.Now()}}
	volatility := a.calculateVolatility(points)
	assert.Equal(t, 0.0, volatility)
}

func TestBasicAnalyzer_CalculateVolatility_ZeroPrevValue(t *testing.T) {
	a := NewBasicAnalyzer()
	points := []*types.DataPoint{
		{Value: 0.0, Timestamp: time.Now()},
		{Value: 10.0, Timestamp: time.Now()},
	}
	volatility := a.calculateVolatility(points)
	assert.Equal(t, 0.0, volatility)
}

func TestBasicAnalyzer_DetectAnomalies(t *testing.T) {
	a := NewBasicAnalyzer()
	points := []*types.DataPoint{
		{Value: 10.0, Timestamp: time.Now()},
		{Value: 10.1, Timestamp: time.Now()},
		{Value: 10.2, Timestamp: time.Now()},
		{Value: 10.1, Timestamp: time.Now()},
		{Value: 10.3, Timestamp: time.Now()},
	}
	anomalies := a.detectAnomalies(points)
	_ = anomalies
}

func TestBasicAnalyzer_DetectAnomalies_FewPoints(t *testing.T) {
	a := NewBasicAnalyzer()
	points := []*types.DataPoint{
		{Value: 10.0, Timestamp: time.Now()},
		{Value: 20.0, Timestamp: time.Now()},
	}
	anomalies := a.detectAnomalies(points)
	assert.Empty(t, anomalies)
}

func TestBasicAnalyzer_Close(t *testing.T) {
	a := NewBasicAnalyzer()
	err := a.Close()
	assert.NoError(t, err)
}
