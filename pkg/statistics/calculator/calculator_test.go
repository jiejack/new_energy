package calculator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockDataProvider 模拟数据提供者
type MockDataProvider struct {
	stations      []StationInfo
	devices       []DeviceInfo
	alarms        []AlarmInfo
	points        []PointInfo
	timeSeriesData map[string][]TimeSeriesPoint
}

func (m *MockDataProvider) GetTimeSeriesData(ctx context.Context, pointIDs []string, start, end time.Time) (map[string][]TimeSeriesPoint, error) {
	result := make(map[string][]TimeSeriesPoint)
	for _, pointID := range pointIDs {
		if data, ok := m.timeSeriesData[pointID]; ok {
			result[pointID] = data
		}
	}
	return result, nil
}

func (m *MockDataProvider) GetDevices(ctx context.Context, stationID string) ([]DeviceInfo, error) {
	var devices []DeviceInfo
	for _, d := range m.devices {
		if d.StationID == stationID {
			devices = append(devices, d)
		}
	}
	return devices, nil
}

func (m *MockDataProvider) GetAllDevices(ctx context.Context) ([]DeviceInfo, error) {
	return m.devices, nil
}

func (m *MockDataProvider) GetStation(ctx context.Context, stationID string) (*StationInfo, error) {
	for _, s := range m.stations {
		if s.ID == stationID {
			return &s, nil
		}
	}
	return nil, nil
}

func (m *MockDataProvider) GetAllStations(ctx context.Context) ([]StationInfo, error) {
	return m.stations, nil
}

func (m *MockDataProvider) GetAlarms(ctx context.Context, stationID string, start, end time.Time) ([]AlarmInfo, error) {
	var alarms []AlarmInfo
	for _, a := range m.alarms {
		if a.StationID == stationID && a.TriggeredAt.After(start) && a.TriggeredAt.Before(end) {
			alarms = append(alarms, a)
		}
	}
	return alarms, nil
}

func (m *MockDataProvider) GetPoints(ctx context.Context, stationID string, pointType string) ([]PointInfo, error) {
	var points []PointInfo
	for _, p := range m.points {
		if p.Type == pointType {
			points = append(points, p)
		}
	}
	return points, nil
}

// MockStatisticsStorage 模拟统计存储
type MockStatisticsStorage struct {
	data []*StatisticsData
}

func (m *MockStatisticsStorage) Save(ctx context.Context, data *StatisticsData) error {
	m.data = append(m.data, data)
	return nil
}

func (m *MockStatisticsStorage) SaveBatch(ctx context.Context, data []*StatisticsData) error {
	m.data = append(m.data, data...)
	return nil
}

func (m *MockStatisticsStorage) Query(ctx context.Context, query *StatisticsQuery) ([]*StatisticsData, error) {
	return m.data, nil
}

func (m *MockStatisticsStorage) QueryLatest(ctx context.Context, dimension, dimensionValue, metricName string) (*StatisticsData, error) {
	for i := len(m.data) - 1; i >= 0; i-- {
		if m.data[i].Dimension == dimension && m.data[i].DimensionValue == dimensionValue && m.data[i].MetricName == metricName {
			return m.data[i], nil
		}
	}
	return nil, nil
}

func (m *MockStatisticsStorage) SaveTimeSeries(ctx context.Context, data *TimeSeriesData) error {
	return nil
}

func (m *MockStatisticsStorage) QueryTimeSeries(ctx context.Context, pointID string, start, end time.Time) (*TimeSeriesData, error) {
	return &TimeSeriesData{PointID: pointID}, nil
}

func (m *MockStatisticsStorage) SaveTask(ctx context.Context, task *StatisticsTask) error {
	return nil
}

func (m *MockStatisticsStorage) GetTask(ctx context.Context, taskID string) (*StatisticsTask, error) {
	return nil, nil
}

func (m *MockStatisticsStorage) ListTasks(ctx context.Context, enabled *bool) ([]*StatisticsTask, error) {
	return nil, nil
}

func (m *MockStatisticsStorage) UpdateTaskRunTime(ctx context.Context, taskID string, lastRun, nextRun time.Time) error {
	return nil
}

func (m *MockStatisticsStorage) CompressData(ctx context.Context, before time.Time) error {
	return nil
}

func (m *MockStatisticsStorage) ArchiveData(ctx context.Context, before time.Time) error {
	return nil
}

func (m *MockStatisticsStorage) Ping(ctx context.Context) error {
	return nil
}

func (m *MockStatisticsStorage) Close() error {
	return nil
}

func TestNewStationCalculator(t *testing.T) {
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}

	config := StationCalculatorConfig{
		ParallelWorkers: 5,
		BatchSize:       100,
		CacheEnabled:    true,
		CacheTTL:        5 * time.Minute,
		DataProvider:    provider,
	}

	calc := NewStationCalculator(config, storage)

	assert.NotNil(t, calc)
	assert.Equal(t, 5, config.ParallelWorkers)
	assert.True(t, config.CacheEnabled)
}

func TestStationCalculator_CalculateGeneration(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{
			{
				ID:       "station-001",
				Code:     "ST001",
				Name:     "测试厂站",
				Type:     "solar",
				Capacity: 1000.0,
				Status:   1,
			},
		},
		points: []PointInfo{
			{
				ID:       "point-001",
				Code:     "P001",
				Name:     "发电量",
				Type:     "generation",
				DeviceID: "device-001",
				Unit:     "kWh",
			},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {
				{Timestamp: now.Add(-2 * time.Hour), Value: 100.0, Quality: 1},
				{Timestamp: now.Add(-1 * time.Hour), Value: 150.0, Quality: 1},
				{Timestamp: now, Value: 200.0, Quality: 1},
			},
		},
	}

	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{
		DataProvider: provider,
	}

	calc := NewStationCalculator(config, storage)

	start := now.Add(-3 * time.Hour)
	end := now

	stats, err := calc.CalculateGeneration(context.Background(), "station-001", PeriodTypeHour, start, end)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "station-001", stats.StationID)
	assert.Equal(t, PeriodTypeHour, stats.PeriodType)
	assert.True(t, stats.PeakPower > 0)
}

func TestStationCalculator_CalculateDeviceRunRate(t *testing.T) {
	now := time.Now()
	lastOnline := now.Add(-1 * time.Hour)

	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "设备1", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 1, LastOnline: &lastOnline},
			{ID: "device-002", Code: "D002", Name: "设备2", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 1, LastOnline: &lastOnline},
			{ID: "device-003", Code: "D003", Name: "设备3", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 2}, // Fault
			{ID: "device-004", Code: "D004", Name: "设备4", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 3}, // Maintain
		},
	}

	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{
		DataProvider: provider,
	}

	calc := NewStationCalculator(config, storage)

	start := now.Add(-24 * time.Hour)
	end := now

	stats, err := calc.CalculateDeviceRunRate(context.Background(), "station-001", start, end)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 4, stats.TotalDevices)
	assert.Equal(t, 2, stats.OnlineDevices)
	assert.Equal(t, 1, stats.FaultDevices)
	assert.Equal(t, 1, stats.MaintainDevices)
	assert.Equal(t, 50.0, stats.RunRate)
}

func TestStationCalculator_CalculateAlarmStats(t *testing.T) {
	now := time.Now()
	acknowledgedAt := now.Add(-30 * time.Minute)
	clearedAt := now.Add(-15 * time.Minute)

	provider := &MockDataProvider{
		alarms: []AlarmInfo{
			{ID: "alarm-001", StationID: "station-001", DeviceID: "device-001", Type: "overload", Level: 4, Status: 1, TriggeredAt: now.Add(-2 * time.Hour)},
			{ID: "alarm-002", StationID: "station-001", DeviceID: "device-002", Type: "fault", Level: 3, Status: 2, TriggeredAt: now.Add(-1 * time.Hour), AcknowledgedAt: &acknowledgedAt},
			{ID: "alarm-003", StationID: "station-001", DeviceID: "device-003", Type: "warning", Level: 2, Status: 3, TriggeredAt: now.Add(-3 * time.Hour), ClearedAt: &clearedAt},
		},
	}

	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{
		DataProvider: provider,
	}

	calc := NewStationCalculator(config, storage)

	start := now.Add(-24 * time.Hour)
	end := now

	stats, err := calc.CalculateAlarmStats(context.Background(), "station-001", start, end)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(3), stats.TotalCount)
	assert.Equal(t, int64(1), stats.ActiveCount)
	assert.Equal(t, int64(1), stats.AcknowledgedCount)
	assert.Equal(t, int64(1), stats.ClearedCount)
	assert.Equal(t, int64(1), stats.CriticalCount)
	assert.Equal(t, int64(1), stats.MajorCount)
	assert.Equal(t, int64(1), stats.WarningCount)
}

func TestStationCalculator_CalculateEfficiency(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{
			{
				ID:       "station-001",
				Code:     "ST001",
				Name:     "测试厂站",
				Type:     "solar",
				Capacity: 1000.0,
				Status:   1,
			},
		},
		points: []PointInfo{
			{
				ID:       "point-eff-001",
				Code:     "system_eff",
				Name:     "系统效率",
				Type:     "efficiency",
				DeviceID: "device-001",
				Unit:     "%",
			},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-eff-001": {
				{Timestamp: now.Add(-2 * time.Hour), Value: 95.0, Quality: 1},
				{Timestamp: now.Add(-1 * time.Hour), Value: 96.0, Quality: 1},
				{Timestamp: now, Value: 97.0, Quality: 1},
			},
		},
	}

	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{
		DataProvider: provider,
	}

	calc := NewStationCalculator(config, storage)

	start := now.Add(-3 * time.Hour)
	end := now

	stats, err := calc.CalculateEfficiency(context.Background(), "station-001", start, end)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "station-001", stats.StationID)
	assert.True(t, stats.SystemEfficiency > 0)
}

func TestStationCalculator_CalculateEquivalentHours(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		stations: []StationInfo{
			{
				ID:       "station-001",
				Code:     "ST001",
				Name:     "测试厂站",
				Type:     "solar",
				Capacity: 1000.0,
				Status:   1,
			},
		},
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "设备1", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 1},
		},
		points: []PointInfo{
			{
				ID:       "point-001",
				Code:     "P001",
				Name:     "发电量",
				Type:     "generation",
				DeviceID: "device-001",
				Unit:     "kWh",
			},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {
				{Timestamp: now.Add(-2 * time.Hour), Value: 500.0, Quality: 1},
				{Timestamp: now.Add(-1 * time.Hour), Value: 600.0, Quality: 1},
				{Timestamp: now, Value: 700.0, Quality: 1},
			},
		},
	}

	storage := &MockStatisticsStorage{}
	config := StationCalculatorConfig{
		DataProvider: provider,
	}

	calc := NewStationCalculator(config, storage)

	start := now.Add(-24 * time.Hour)
	end := now

	stats, err := calc.CalculateEquivalentHours(context.Background(), "station-001", start, end)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "station-001", stats.StationID)
	assert.Equal(t, 24.0, stats.TotalHours)
}

func TestStatisticsCache(t *testing.T) {
	cache := NewStatisticsCache(5 * time.Minute)

	// 测试设置和获取
	cache.Set("key1", "value1")
	value, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", value)

	// 测试不存在的键
	_, ok = cache.Get("key2")
	assert.False(t, ok)
}

func TestNewDeviceCalculator(t *testing.T) {
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}

	config := DeviceCalculatorConfig{
		ParallelWorkers: 4,
		BatchSize:       100,
		CacheEnabled:    true,
		CacheTTL:        5 * time.Minute,
		DataProvider:    provider,
	}

	calc := NewDeviceCalculator(config, storage)

	assert.NotNil(t, calc)
	assert.Equal(t, 4, config.ParallelWorkers)
	assert.True(t, config.CacheEnabled)
}

func TestDeviceCalculator_CalculateDeviceTypeStatistics(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "逆变器1", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 1},
			{ID: "device-002", Code: "D002", Name: "逆变器2", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 1},
			{ID: "device-003", Code: "D003", Name: "逆变器3", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 2},
		},
	}

	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{
		DataProvider: provider,
	}

	calc := NewDeviceCalculator(config, storage)

	start := now.Add(-24 * time.Hour)
	end := now

	stats, err := calc.CalculateDeviceTypeStatistics(context.Background(), "inverter", "station-001", start, end)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "inverter", stats.DeviceType)
	assert.Equal(t, 3, stats.TotalDevices)
	assert.Equal(t, 2, stats.OnlineDevices)
	assert.Equal(t, 1, stats.FaultDevices)
}

func TestDeviceCalculator_CalculateDevicePerformance(t *testing.T) {
	now := time.Now()
	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "逆变器1", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 1},
		},
		points: []PointInfo{
			{
				ID:       "point-001",
				Code:     "P001",
				Name:     "功率",
				Type:     "power",
				DeviceID: "device-001",
				Unit:     "kW",
			},
		},
		timeSeriesData: map[string][]TimeSeriesPoint{
			"point-001": {
				{Timestamp: now.Add(-2 * time.Hour), Value: 50.0, Quality: 1},
				{Timestamp: now.Add(-1 * time.Hour), Value: 75.0, Quality: 1},
				{Timestamp: now, Value: 90.0, Quality: 1},
			},
		},
	}

	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{
		DataProvider: provider,
	}

	calc := NewDeviceCalculator(config, storage)

	start := now.Add(-3 * time.Hour)
	end := now

	stats, err := calc.CalculateDevicePerformance(context.Background(), "device-001", start, end)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "device-001", stats.DeviceID)
	assert.Equal(t, 100.0, stats.RatedPower)
	assert.True(t, stats.MaxPower > 0)
}

func TestDeviceCalculator_CalculateDeviceFaultStats(t *testing.T) {
	now := time.Now()
	clearedAt := now.Add(-30 * time.Minute)

	provider := &MockDataProvider{
		devices: []DeviceInfo{
			{ID: "device-001", Code: "D001", Name: "逆变器1", Type: "inverter", StationID: "station-001", RatedPower: 100.0, Status: 1},
		},
		alarms: []AlarmInfo{
			{ID: "alarm-001", StationID: "station-001", DeviceID: "device-001", Type: "fault", Level: 4, Status: 3, TriggeredAt: now.Add(-2 * time.Hour), ClearedAt: &clearedAt},
			{ID: "alarm-002", StationID: "station-001", DeviceID: "device-001", Type: "warning", Level: 2, Status: 3, TriggeredAt: now.Add(-1 * time.Hour), ClearedAt: &clearedAt},
		},
	}

	storage := &MockStatisticsStorage{}
	config := DeviceCalculatorConfig{
		DataProvider: provider,
	}

	calc := NewDeviceCalculator(config, storage)

	start := now.Add(-24 * time.Hour)
	end := now

	stats, err := calc.CalculateDeviceFaultStats(context.Background(), "device-001", start, end)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "device-001", stats.DeviceID)
	assert.Equal(t, int64(2), stats.TotalFaultCount)
	assert.Equal(t, int64(1), stats.CriticalFaultCount)
}

func TestNewCustomCalculator(t *testing.T) {
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}

	config := CustomCalculatorConfig{
		ParallelWorkers: 4,
		BatchSize:       100,
		CacheEnabled:    true,
		CacheTTL:        5 * time.Minute,
		DataProvider:    provider,
	}

	calc := NewCustomCalculator(config, storage)

	assert.NotNil(t, calc)
	assert.Equal(t, 4, config.ParallelWorkers)
	assert.True(t, config.CacheEnabled)
}

func TestCustomCalculator_RegisterConfig(t *testing.T) {
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}

	config := CustomCalculatorConfig{
		DataProvider: provider,
	}

	calc := NewCustomCalculator(config, storage)

	customConfig := &CustomStatisticsConfig{
		ID:          "config-001",
		Name:        "测试统计",
		Description: "测试统计配置",
		PeriodType:  PeriodTypeDay,
	}

	err := calc.RegisterConfig(customConfig)
	assert.NoError(t, err)

	// 验证配置已注册
	retrieved, err := calc.GetConfig("config-001")
	assert.NoError(t, err)
	assert.Equal(t, "测试统计", retrieved.Name)
}

func TestCustomCalculator_ListConfigs(t *testing.T) {
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}

	config := CustomCalculatorConfig{
		DataProvider: provider,
	}

	calc := NewCustomCalculator(config, storage)

	// 注册多个配置
	calc.RegisterConfig(&CustomStatisticsConfig{ID: "config-001", Name: "统计1"})
	calc.RegisterConfig(&CustomStatisticsConfig{ID: "config-002", Name: "统计2"})
	calc.RegisterConfig(&CustomStatisticsConfig{ID: "config-003", Name: "统计3"})

	configs := calc.ListConfigs()
	assert.Len(t, configs, 3)
}

func TestCustomCalculator_CreatePresetConfig(t *testing.T) {
	provider := &MockDataProvider{}
	storage := &MockStatisticsStorage{}

	config := CustomCalculatorConfig{
		DataProvider: provider,
	}

	calc := NewCustomCalculator(config, storage)

	// 测试创建预设配置
	presetConfig, err := calc.CreatePresetConfig("generation_by_hour", map[string]interface{}{
		"station_id": "station-001",
	})

	assert.NoError(t, err)
	assert.NotNil(t, presetConfig)
	assert.Equal(t, "按小时发电量统计", presetConfig.Name)
	assert.Equal(t, PeriodTypeHour, presetConfig.PeriodType)
}

func TestCalculateAggregated(t *testing.T) {
	values := []float64{10.0, 20.0, 30.0, 40.0, 50.0}

	stats := CalculateAggregated(values)

	assert.Equal(t, 150.0, stats.Sum)
	assert.Equal(t, 30.0, stats.Avg)
	assert.Equal(t, 10.0, stats.Min)
	assert.Equal(t, 50.0, stats.Max)
	assert.Equal(t, int64(5), stats.Count)
	assert.True(t, stats.StdDev > 0)
	assert.True(t, stats.Variance > 0)
}

func TestCalculateAggregated_Empty(t *testing.T) {
	values := []float64{}

	stats := CalculateAggregated(values)

	assert.Equal(t, 0.0, stats.Sum)
	assert.Equal(t, 0.0, stats.Avg)
	assert.Equal(t, int64(0), stats.Count)
}

func TestStatisticsResult_ToStatisticsData(t *testing.T) {
	now := time.Now()

	result := &StatisticsResult{
		Dimension:      "station",
		DimensionValue: "station-001",
		Metrics: map[string]float64{
			"daily_generation": 1000.0,
			"efficiency":       95.5,
		},
		Metadata: map[string]interface{}{
			"station_name": "测试厂站",
		},
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		PeriodType:  PeriodTypeDay,
	}

	data := result.ToStatisticsData("task-001")

	assert.Len(t, data, 2)
	assert.Equal(t, "task-001", data[0].TaskID)
	assert.Equal(t, "station", data[0].Dimension)
	assert.Equal(t, "station-001", data[0].DimensionValue)
}

func TestSumValues(t *testing.T) {
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	result := sumValues(values)
	assert.Equal(t, 15.0, result)
}

func TestAvgValues(t *testing.T) {
	values := []float64{10.0, 20.0, 30.0}
	result := avgValues(values)
	assert.Equal(t, 20.0, result)
}

func TestMinValues(t *testing.T) {
	values := []float64{5.0, 2.0, 8.0, 1.0, 9.0}
	result := minValues(values)
	assert.Equal(t, 1.0, result)
}

func TestMaxValues(t *testing.T) {
	values := []float64{5.0, 2.0, 8.0, 1.0, 9.0}
	result := maxValues(values)
	assert.Equal(t, 9.0, result)
}

func TestMedianValues(t *testing.T) {
	// 奇数个元素
	values1 := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	result1 := medianValues(values1)
	assert.Equal(t, 3.0, result1)

	// 偶数个元素
	values2 := []float64{1.0, 2.0, 3.0, 4.0}
	result2 := medianValues(values2)
	assert.Equal(t, 2.5, result2)
}

func TestPercentileValues(t *testing.T) {
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0}

	p95 := percentileValues(values, 95)
	assert.True(t, p95 >= 9.0 && p95 <= 10.0)

	p99 := percentileValues(values, 99)
	assert.True(t, p99 >= 9.0 && p99 <= 10.0)
}

func TestToFloat64(t *testing.T) {
	// 测试各种类型转换
	val1, ok1 := toFloat64(float64(10.5))
	assert.True(t, ok1)
	assert.Equal(t, 10.5, val1)

	val2, ok2 := toFloat64(int(10))
	assert.True(t, ok2)
	assert.Equal(t, 10.0, val2)

	val3, ok3 := toFloat64(int64(10))
	assert.True(t, ok3)
	assert.Equal(t, 10.0, val3)

	_, ok4 := toFloat64("invalid")
	assert.False(t, ok4)
}

func TestSplitString(t *testing.T) {
	result := splitString("a|b|c", "|")
	assert.Len(t, result, 3)
	assert.Equal(t, "a", result[0])
	assert.Equal(t, "b", result[1])
	assert.Equal(t, "c", result[2])
}

func TestContains(t *testing.T) {
	assert.True(t, contains("hello world", "world"))
	assert.True(t, contains("hello world", "hello"))
	assert.False(t, contains("hello world", "xyz"))
}
