package calculator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStationCalculator_CalculateStationStatistics(t *testing.T) {
	now := time.Now()
	lastOnline := now.Add(-1 * time.Hour)
	provider := &MockDataProvider{
		stations: []StationInfo{
			{ID: "station-001", Code: "ST001", Name: "Test Station", Type: "solar", Capacity: 1000.0, Status: 1},
		},
		devices: []DeviceInfo{
			{ID: "dev-1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "station-001", RatedPower: 100, Status: 1, LastOnline: &lastOnline},
		},
		points: []PointInfo{
			{ID: "pt-gen-1", Code: "P001", Name: "Gen", Type: "generation", DeviceID: "dev-1"},
			{ID: "pt-eff-1", Code: "system_eff", Name: "Eff", Type: "efficiency", DeviceID: "dev-1"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"pt-gen-1": {{Timestamp: now.Add(-1 * time.Hour), Value: 500, Quality: 1}, {Timestamp: now, Value: 600, Quality: 1}},
			"pt-eff-1": {{Timestamp: now.Add(-1 * time.Hour), Value: 95, Quality: 1}, {Timestamp: now, Value: 96, Quality: 1}},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewStationCalculator(StationCalculatorConfig{DataProvider: provider}, storage)
	stats, err := calc.CalculateStationStatistics(context.Background(), "station-001", PeriodTypeDay, now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "station-001", stats.StationID)
}

func TestStationCalculator_CalculateStationStatistics_WithCache(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{
			{ID: "station-001", Code: "ST001", Name: "Test", Type: "solar", Capacity: 1000, Status: 1},
		},
		devices: []DeviceInfo{
			{ID: "dev-1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "station-001", Status: 1},
		},
		points: []PointInfo{
			{ID: "pt-gen-1", Code: "P001", Type: "generation", DeviceID: "dev-1"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"pt-gen-1": {{Timestamp: now, Value: 100, Quality: 1}},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewStationCalculator(StationCalculatorConfig{DataProvider: provider, CacheEnabled: true, CacheTTL: 5 * time.Minute}, storage)
	stats1, err := calc.CalculateStationStatistics(context.Background(), "station-001", PeriodTypeDay, now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	stats2, err := calc.CalculateStationStatistics(context.Background(), "station-001", PeriodTypeDay, now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	assert.Equal(t, stats1.StationID, stats2.StationID)
}

func TestStationCalculator_CalculateAllStations(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{
			{ID: "s1", Code: "ST1", Name: "Station1", Type: "solar", Capacity: 500, Status: 1},
			{ID: "s2", Code: "ST2", Name: "Station2", Type: "wind", Capacity: 800, Status: 1},
		},
		devices: []DeviceInfo{
			{ID: "d1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "s1", Status: 1},
			{ID: "d2", Code: "D2", Name: "Dev2", Type: "inverter", StationID: "s2", Status: 1},
		},
		points: []PointInfo{
			{ID: "pg1", Code: "P1", Type: "generation", DeviceID: "d1"},
			{ID: "pg2", Code: "P2", Type: "generation", DeviceID: "d2"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"pg1": {{Timestamp: now, Value: 100, Quality: 1}},
			"pg2": {{Timestamp: now, Value: 200, Quality: 1}},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewStationCalculator(StationCalculatorConfig{DataProvider: provider, ParallelWorkers: 2}, storage)
	results, err := calc.CalculateAllStations(context.Background(), PeriodTypeDay, now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestStationCalculator_SaveStatistics(t *testing.T) {
	now := time.Now()
	storage := &MockStatisticsStorage{}
	calc := NewStationCalculator(StationCalculatorConfig{DataProvider: &MockDataProvider{}}, storage)
	stats := &StationStatistics{
		StationID:   "s1",
		StationCode: "ST1",
		StationName: "Test",
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
	}
	err := calc.SaveStatistics(context.Background(), "task-1", stats)
	assert.NoError(t, err)
}

func TestStationCalculator_CalculateEfficiency_DefaultMethod(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{
			{ID: "s1", Code: "ST1", Name: "Test", Type: "solar", Capacity: 1000, Status: 1},
		},
		points: []PointInfo{
			{ID: "inp1", Code: "input_power", Type: "input_power", DeviceID: "d1"},
			{ID: "out1", Code: "output_power", Type: "output_power", DeviceID: "d1"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"inp1": {{Timestamp: now, Value: 200, Quality: 1}},
			"out1": {{Timestamp: now, Value: 180, Quality: 1}},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewStationCalculator(StationCalculatorConfig{DataProvider: provider}, storage)
	stats, err := calc.CalculateEfficiency(context.Background(), "s1", now.Add(-1*time.Hour), now)
	require.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestDeviceCalculator_CalculateDeviceStatus(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "s1", Status: 1},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider}, storage)
	stats, err := calc.CalculateDeviceStatus(context.Background(), "d1", now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "d1", stats.DeviceID)
}

func TestDeviceCalculator_CalculateDeviceStatus_NotFound(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{devices: []DeviceInfo{}}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider}, storage)
	_, err := calc.CalculateDeviceStatus(context.Background(), "nonexistent", now.Add(-24*time.Hour), now)
	assert.Error(t, err)
}

func TestDeviceCalculator_CalculateDeviceStatus_Offline(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "s1", Status: 0},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider}, storage)
	stats, err := calc.CalculateDeviceStatus(context.Background(), "d1", now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestDeviceCalculator_CalculateDeviceStatus_Fault(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "s1", Status: 2},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider}, storage)
	stats, err := calc.CalculateDeviceStatus(context.Background(), "d1", now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestDeviceCalculator_CalculateDeviceStatus_Maintain(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "s1", Status: 3},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider}, storage)
	stats, err := calc.CalculateDeviceStatus(context.Background(), "d1", now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestDeviceCalculator_CalculateDeviceAvailability(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "s1", Status: 1},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider}, storage)
	stats, err := calc.CalculateDeviceAvailability(context.Background(), "d1", now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.True(t, stats.Availability > 0)
}

func TestDeviceCalculator_CalculateAllDeviceTypes(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "s1", Status: 1},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider, ParallelWorkers: 2}, storage)
	results, err := calc.CalculateAllDeviceTypes(context.Background(), "s1", now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	assert.NotNil(t, results)
}

func TestDeviceCalculator_SaveDeviceTypeStatistics(t *testing.T) {
	now := time.Now()
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: &MockDataProvider{}}, storage)
	stats := &DeviceTypeStatistics{
		DeviceType:  "inverter",
		StationID:   "s1",
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
	}
	err := calc.SaveDeviceTypeStatistics(context.Background(), "task-1", stats)
	assert.NoError(t, err)
}

func TestDeviceCalculator_SaveDevicePerformanceStatistics(t *testing.T) {
	now := time.Now()
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: &MockDataProvider{}}, storage)
	stats := &DevicePerformanceStatistics{
		DeviceID:    "d1",
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
	}
	err := calc.SaveDevicePerformanceStatistics(context.Background(), "task-1", stats)
	assert.NoError(t, err)
}

func TestDeviceCalculator_SaveDeviceFaultStatistics(t *testing.T) {
	now := time.Now()
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: &MockDataProvider{}}, storage)
	stats := &DeviceFaultStatistics{
		DeviceID:             "d1",
		PeriodStart:          now.Add(-24 * time.Hour),
		PeriodEnd:            now,
		FaultTypeDistribution: map[string]int64{"fault": 1},
	}
	err := calc.SaveDeviceFaultStatistics(context.Background(), "task-1", stats)
	assert.NoError(t, err)
}

func TestDeviceCalculator_SaveDeviceAvailabilityStatistics(t *testing.T) {
	now := time.Now()
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: &MockDataProvider{}}, storage)
	stats := &DeviceAvailabilityStatistics{
		DeviceID:     "d1",
		PeriodStart:  now.Add(-24 * time.Hour),
		PeriodEnd:    now,
		TotalHours:   24,
		OutageCount:  2,
	}
	err := calc.SaveDeviceAvailabilityStatistics(context.Background(), "task-1", stats)
	assert.NoError(t, err)
}

func TestDeviceCalculator_CalculateDeviceTypeStatistics_NoStation(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "s1", Status: 1},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider}, storage)
	stats, err := calc.CalculateDeviceTypeStatistics(context.Background(), "inverter", "", now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestDeviceCalculator_CalculateDeviceTypeStatistics_WithCache(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "s1", Status: 1},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider, CacheEnabled: true, CacheTTL: 5 * time.Minute}, storage)
	stats1, err := calc.CalculateDeviceTypeStatistics(context.Background(), "inverter", "s1", now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	stats2, err := calc.CalculateDeviceTypeStatistics(context.Background(), "inverter", "s1", now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	assert.Equal(t, stats1.DeviceType, stats2.DeviceType)
}

func TestCustomCalculator_Calculate(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		points: []PointInfo{
			{ID: "pt-1", Code: "P1", Type: "generation", DeviceID: "d1"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"pt-1": {{Timestamp: now, Value: 100, Quality: 1}},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: provider}, storage)
	cfg := &CustomStatisticsConfig{
		ID:          "cfg-1",
		Name:        "test",
		PeriodType:  PeriodTypeDay,
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		Metrics: []CustomMetric{
			{Name: "gen", Aggregation: AggregationSum, PointIDs: []string{"pt-1"}},
		},
	}
	calc.RegisterConfig(cfg)
	results, err := calc.Calculate(context.Background(), "cfg-1")
	require.NoError(t, err)
	assert.NotNil(t, results)
}

func TestCustomCalculator_Calculate_NotFound(t *testing.T) {
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: provider}, storage)
	_, err := calc.Calculate(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestCustomCalculator_CalculateWithConfig_NoPoints(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: provider}, storage)
	cfg := &CustomStatisticsConfig{
		ID:          "cfg-np",
		Name:        "no points",
		PeriodType:  PeriodTypeDay,
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		Metrics:     []CustomMetric{},
	}
	results, err := calc.CalculateWithConfig(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, results)
}

func TestCustomCalculator_CalculateWithConfig_WithCache(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		points: []PointInfo{
			{ID: "pt-1", Code: "P1", Type: "generation", DeviceID: "d1"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"pt-1": {{Timestamp: now, Value: 50, Quality: 1}},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: provider, CacheEnabled: true, CacheTTL: 5 * time.Minute}, storage)
	cfg := &CustomStatisticsConfig{
		ID:          "cfg-cache",
		Name:        "cached",
		PeriodType:  PeriodTypeDay,
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		Metrics: []CustomMetric{
			{Name: "gen", Aggregation: AggregationAvg, PointIDs: []string{"pt-1"}},
		},
	}
	r1, err := calc.CalculateWithConfig(context.Background(), cfg)
	require.NoError(t, err)
	r2, err := calc.CalculateWithConfig(context.Background(), cfg)
	require.NoError(t, err)
	assert.Equal(t, r1.ConfigID, r2.ConfigID)
}

func TestCustomCalculator_CalculateMultiDimension(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: provider}, storage)
	cfg := &CustomStatisticsConfig{
		ID:          "cfg-md",
		Name:        "multi",
		PeriodType:  PeriodTypeDay,
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		Metrics:     []CustomMetric{},
	}
	results, err := calc.CalculateMultiDimension(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, results)
}

func TestCustomCalculator_CalculateTimeSeries(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: provider}, storage)
	cfg := &CustomStatisticsConfig{
		ID:          "cfg-ts",
		Name:        "timeseries",
		PeriodType:  PeriodTypeDay,
		PeriodStart: now.Add(-3 * time.Hour),
		PeriodEnd:   now,
		Metrics:     []CustomMetric{},
	}
	results, err := calc.CalculateTimeSeries(context.Background(), cfg, 1*time.Hour)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestCustomCalculator_SaveResults(t *testing.T) {
	now := time.Now()
	storage := &MockStatisticsStorage{}
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, storage)
	results := &CustomStatisticsResults{
		ConfigID:    "cfg-1",
		ConfigName:  "test",
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		Results: []*CustomStatisticsResult{
			{
				ConfigID:    "cfg-1",
				ConfigName:  "test",
				PeriodStart: now.Add(-24 * time.Hour),
				PeriodEnd:   now,
				Dimensions:  map[string]string{"station": "s1"},
				Metrics:     map[string]float64{"gen": 1000},
				Metadata:    map[string]interface{}{"quality": 95.0},
			},
		},
	}
	err := calc.SaveResults(context.Background(), "task-1", results)
	assert.NoError(t, err)
}

func TestCustomCalculator_CreatePresetConfig_DeviceStatus(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	cfg, err := calc.CreatePresetConfig("device_status_summary", nil)
	require.NoError(t, err)
	assert.Equal(t, "设备状态汇总", cfg.Name)
}

func TestCustomCalculator_CreatePresetConfig_AlarmStatistics(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	cfg, err := calc.CreatePresetConfig("alarm_statistics", nil)
	require.NoError(t, err)
	assert.Equal(t, "告警统计", cfg.Name)
}

func TestCustomCalculator_CreatePresetConfig_EfficiencyAnalysis(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	cfg, err := calc.CreatePresetConfig("efficiency_analysis", nil)
	require.NoError(t, err)
	assert.Equal(t, "效率分析", cfg.Name)
}

func TestCustomCalculator_CreatePresetConfig_Unknown(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	_, err := calc.CreatePresetConfig("unknown_type", nil)
	assert.Error(t, err)
}

func TestCustomCalculator_RegisterConfig_AutoID(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	cfg := &CustomStatisticsConfig{Name: "auto-id"}
	err := calc.RegisterConfig(cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.ID)
}

func TestCustomCalculator_GetConfig_NotFound(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	_, err := calc.GetConfig("nonexistent")
	assert.Error(t, err)
}

func TestCustomCalculator_ApplyFilters(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	data := []map[string]interface{}{
		{"name": "a", "value": 10.0},
		{"name": "b", "value": 20.0},
		{"name": "c", "value": 30.0},
	}
	filtered := calc.applyFilters(data, FilterGroup{
		Conditions: []FilterCondition{
			{Field: "value", Operator: "gt", Value: 15.0},
		},
		Logic: "and",
	})
	assert.Len(t, filtered, 2)
}

func TestCustomCalculator_ApplyFilters_Empty(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	data := []map[string]interface{}{{"name": "a"}}
	filtered := calc.applyFilters(data, FilterGroup{})
	assert.Len(t, filtered, 1)
}

func TestCustomCalculator_EvaluateFilterGroup_Or(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	data := []map[string]interface{}{
		{"name": "a", "value": 5.0},
		{"name": "b", "value": 25.0},
	}
	filtered := calc.applyFilters(data, FilterGroup{
		Conditions: []FilterCondition{
			{Field: "value", Operator: "lt", Value: 10.0},
			{Field: "value", Operator: "gt", Value: 20.0},
		},
		Logic: "or",
	})
	assert.Len(t, filtered, 2)
}

func TestCustomCalculator_EvaluateCondition_Operators(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	record := map[string]interface{}{"value": 50.0, "name": "test"}
	tests := []struct {
		op    string
		val   interface{}
		expect bool
	}{
		{"eq", 50.0, true},
		{"ne", 40.0, true},
		{"gt", 40.0, true},
		{"lt", 60.0, true},
		{"gte", 50.0, true},
		{"lte", 50.0, true},
		{"in", []interface{}{50.0, 60.0}, true},
		{"not_in", []interface{}{30.0, 40.0}, true},
		{"like", "es", true},
	}
	for _, tt := range tests {
		result := calc.evaluateCondition(record, FilterCondition{Field: "value", Operator: tt.op, Value: tt.val})
		if tt.op == "like" {
			result = calc.evaluateCondition(record, FilterCondition{Field: "name", Operator: tt.op, Value: tt.val})
		}
		assert.Equal(t, tt.expect, result, "operator %s with value %v", tt.op, tt.val)
	}
}

func TestCustomCalculator_EvaluateCondition_MissingField(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	record := map[string]interface{}{"value": 50.0}
	result := calc.evaluateCondition(record, FilterCondition{Field: "missing", Operator: "eq", Value: 50})
	assert.False(t, result)
}

func TestCustomCalculator_EvaluateCondition_InNotIn(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	record := map[string]interface{}{"name": "test"}
	result := calc.evaluateCondition(record, FilterCondition{Field: "name", Operator: "in", Value: []interface{}{"test", "other"}})
	assert.True(t, result)
	result = calc.evaluateCondition(record, FilterCondition{Field: "name", Operator: "not_in", Value: []interface{}{"test", "other"}})
	assert.False(t, result)
}

func TestCustomCalculator_ApplySorting(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	results := []*CustomStatisticsResult{
		{Metrics: map[string]float64{"val": 30}},
		{Metrics: map[string]float64{"val": 10}},
		{Metrics: map[string]float64{"val": 20}},
	}
	sorted := calc.applySorting(results, []OrderByField{{Field: "val", Desc: false}})
	assert.Equal(t, 10.0, sorted[0].Metrics["val"])
	assert.Equal(t, 30.0, sorted[2].Metrics["val"])
}

func TestCustomCalculator_ApplySorting_Desc(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	results := []*CustomStatisticsResult{
		{Metrics: map[string]float64{"val": 10}},
		{Metrics: map[string]float64{"val": 30}},
	}
	sorted := calc.applySorting(results, []OrderByField{{Field: "val", Desc: true}})
	assert.Equal(t, 30.0, sorted[0].Metrics["val"])
}

func TestCustomCalculator_ApplyPagination(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	results := []*CustomStatisticsResult{
		{Metrics: map[string]float64{"val": 1}},
		{Metrics: map[string]float64{"val": 2}},
		{Metrics: map[string]float64{"val": 3}},
		{Metrics: map[string]float64{"val": 4}},
		{Metrics: map[string]float64{"val": 5}},
	}
	paged := calc.applyPagination(results, 2, 1)
	assert.Len(t, paged, 2)
}

func TestCustomCalculator_ApplyPagination_OffsetExceeds(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	results := []*CustomStatisticsResult{{Metrics: map[string]float64{"val": 1}}}
	paged := calc.applyPagination(results, 10, 5)
	assert.Empty(t, paged)
}

func TestCustomCalculator_CalculateMetric_AllAggregations(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	data := []map[string]interface{}{
		{"value": 10.0},
		{"value": 20.0},
		{"value": 30.0},
	}
	aggs := []AggregationType{
		AggregationSum, AggregationAvg, AggregationMin, AggregationMax,
		AggregationCount, AggregationFirst, AggregationLast,
		AggregationStdDev, AggregationVariance, AggregationMedian,
		AggregationP95, AggregationP99,
	}
	for _, agg := range aggs {
		metric := CustomMetric{Name: string(agg), Aggregation: agg, ScaleFactor: 1.0}
		val := calc.calculateMetric(data, metric)
		assert.True(t, val >= 0, "aggregation %s should produce non-negative result", agg)
	}
}

func TestCustomCalculator_CalculateMetric_EmptyData(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	metric := CustomMetric{Name: "test", Aggregation: AggregationAvg}
	val := calc.calculateMetric([]map[string]interface{}{}, metric)
	assert.Equal(t, 0.0, val)
}

func TestCustomCalculator_CalculateMetric_ScaleFactor(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	data := []map[string]interface{}{{"value": 100.0}}
	metric := CustomMetric{Name: "scaled", Aggregation: AggregationSum, ScaleFactor: 0.5, Offset: 10}
	val := calc.calculateMetric(data, metric)
	assert.Equal(t, 60.0, val)
}

func TestCustomCalculator_CalculateMetric_DefaultAggregation(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	data := []map[string]interface{}{{"value": 10.0}, {"value": 20.0}}
	metric := CustomMetric{Name: "default", Aggregation: "unknown_agg"}
	val := calc.calculateMetric(data, metric)
	assert.Equal(t, 0.0, val)
}

func TestCustomCalculator_CalculateSummary(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	results := []*CustomStatisticsResult{
		{Metrics: map[string]float64{"val": 10}},
		{Metrics: map[string]float64{"val": 20}},
	}
	summary := calc.calculateSummary(results)
	assert.NotNil(t, summary)
	assert.Contains(t, summary, "val")
}

func TestCustomCalculator_CalculateSummary_Empty(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	summary := calc.calculateSummary([]*CustomStatisticsResult{})
	assert.Empty(t, summary)
}

func TestStatisticsCache_Expired(t *testing.T) {
	cache := NewStatisticsCache(1 * time.Millisecond)
	cache.Set("key1", "value1")
	time.Sleep(10 * time.Millisecond)
	_, ok := cache.Get("key1")
	assert.False(t, ok)
}

func TestSqrt(t *testing.T) {
	assert.Equal(t, 0.0, sqrt(0))
	assert.Equal(t, 0.0, sqrt(-1))
	result := sqrt(4)
	assert.InDelta(t, 2.0, result, 0.001)
}

func TestMathAbs(t *testing.T) {
	assert.Equal(t, 5.0, mathAbs(-5.0))
	assert.Equal(t, 5.0, mathAbs(5.0))
	assert.Equal(t, 0.0, mathAbs(0.0))
}

func TestMathPow(t *testing.T) {
	result := mathPow(2, 3)
	assert.InDelta(t, 8.0, result, 0.001)
}

func TestMathSqrt(t *testing.T) {
	result := mathSqrt(9)
	assert.InDelta(t, 3.0, result, 0.001)
}

func TestMedianValues_Empty(t *testing.T) {
	assert.Equal(t, 0.0, medianValues([]float64{}))
}

func TestMinValues_Empty(t *testing.T) {
	assert.Equal(t, 0.0, minValues([]float64{}))
}

func TestMaxValues_Empty(t *testing.T) {
	assert.Equal(t, 0.0, maxValues([]float64{}))
}

func TestAvgValues_Empty(t *testing.T) {
	assert.Equal(t, 0.0, avgValues([]float64{}))
}

func TestPercentileValues_Empty(t *testing.T) {
	assert.Equal(t, 0.0, percentileValues([]float64{}, 95))
}

func TestToFloat64_Float32(t *testing.T) {
	val, ok := toFloat64(float32(10.5))
	assert.True(t, ok)
	assert.InDelta(t, 10.5, val, 0.01)
}

func TestToFloat64_Int32(t *testing.T) {
	val, ok := toFloat64(int32(10))
	assert.True(t, ok)
	assert.Equal(t, 10.0, val)
}

func TestSplitString_NoSep(t *testing.T) {
	result := splitString("abc", "|")
	assert.Len(t, result, 1)
	assert.Equal(t, "abc", result[0])
}

func TestContains_Empty(t *testing.T) {
	assert.True(t, contains("abc", "abc"))
	assert.False(t, contains("abc", "xyz"))
	assert.False(t, contains("", "a"))
}

func TestDeviceCalculator_CalculateDevicePerformance_WithEfficiency(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "s1", RatedPower: 100, Status: 1},
		},
		points: []PointInfo{
			{ID: "pw1", Code: "P1", Type: "power", DeviceID: "d1"},
			{ID: "ef1", Code: "E1", Type: "efficiency", DeviceID: "d1"},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"pw1": {{Timestamp: now.Add(-1 * time.Hour), Value: 50, Quality: 1}, {Timestamp: now, Value: 80, Quality: 1}},
			"ef1": {{Timestamp: now.Add(-1 * time.Hour), Value: 95, Quality: 1}, {Timestamp: now, Value: 97, Quality: 1}},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider}, storage)
	stats, err := calc.CalculateDevicePerformance(context.Background(), "d1", now.Add(-2*time.Hour), now)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.True(t, stats.AvgEfficiency > 0)
}

func TestDeviceCalculator_CalculateDeviceFaultStats_NoFaults(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "d1", Code: "D1", Name: "Dev1", Type: "inverter", StationID: "s1", Status: 1},
		},
	}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider}, storage)
	stats, err := calc.CalculateDeviceFaultStats(context.Background(), "d1", now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalFaultCount)
	assert.True(t, stats.MTBF > 0)
}

func TestDeviceCalculator_CalculateDevicePerformance_NotFound(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{devices: []DeviceInfo{}}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider}, storage)
	_, err := calc.CalculateDevicePerformance(context.Background(), "nonexistent", now.Add(-24*time.Hour), now)
	assert.Error(t, err)
}

func TestDeviceCalculator_CalculateDeviceFaultStats_NotFound(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{devices: []DeviceInfo{}}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider}, storage)
	_, err := calc.CalculateDeviceFaultStats(context.Background(), "nonexistent", now.Add(-24*time.Hour), now)
	assert.Error(t, err)
}

func TestDeviceCalculator_CalculateDeviceAvailability_NotFound(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{devices: []DeviceInfo{}}
	storage := &MockStatisticsStorage{}
	calc := NewDeviceCalculator(DeviceCalculatorConfig{DataProvider: provider}, storage)
	_, err := calc.CalculateDeviceAvailability(context.Background(), "nonexistent", now.Add(-24*time.Hour), now)
	assert.Error(t, err)
}

func TestNewDataCompressor(t *testing.T) {
	c := NewDataCompressor(30)
	assert.NotNil(t, c)
	assert.Equal(t, 30, c.compressionDays)
}

func TestNewDataArchiver(t *testing.T) {
	a := NewDataArchiver(90)
	assert.NotNil(t, a)
	assert.Equal(t, 90, a.archiveDays)
}

func TestStatisticsData_TableName(t *testing.T) {
	s := &StatisticsData{}
	assert.Equal(t, "statistics_data", s.TableName())
}

func TestStatisticsTask_TableName(t *testing.T) {
	s := &StatisticsTask{}
	assert.Equal(t, "statistics_tasks", s.TableName())
}

func TestCustomCalculator_GenerateGroupKey(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	record := map[string]interface{}{
		"station_id": "s1",
		"timestamp":  time.Now(),
	}
	key := calc.generateGroupKey(record, []GroupByField{
		{Field: "station_id", Alias: "station"},
	})
	assert.Equal(t, "s1", key)
}

func TestCustomCalculator_GenerateGroupKey_TimeFormat(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	record := map[string]interface{}{
		"timestamp": ts,
	}
	key := calc.generateGroupKey(record, []GroupByField{
		{Field: "timestamp", Alias: "hour", TimeFormat: "2006-01-02 15:00"},
	})
	assert.Contains(t, key, "2024-01-15 10:00")
}

func TestCustomCalculator_ParseGroupKey(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	dims := calc.parseGroupKey("s1|inverter", []GroupByField{
		{Field: "station_id", Alias: "station"},
		{Field: "device_type", Alias: "type"},
	})
	assert.Equal(t, "s1", dims["station"])
	assert.Equal(t, "inverter", dims["type"])
}

func TestCustomCalculator_ParseGroupKey_NoAlias(t *testing.T) {
	calc := NewCustomCalculator(CustomCalculatorConfig{DataProvider: &MockDataProvider{}}, &MockStatisticsStorage{})
	dims := calc.parseGroupKey("s1", []GroupByField{
		{Field: "station_id"},
	})
	assert.Equal(t, "s1", dims["station_id"])
}
