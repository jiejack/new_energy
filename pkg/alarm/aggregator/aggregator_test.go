package aggregator

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func createTestAlarm(id, deviceID, stationID string, level entity.AlarmLevel, triggeredAt time.Time) *entity.Alarm {
	return &entity.Alarm{
		ID:          id,
		DeviceID:    deviceID,
		StationID:   stationID,
		Type:        entity.AlarmTypeDevice,
		Level:       level,
		Title:       "Test Alarm",
		Message:     "Test alarm message",
		TriggeredAt: triggeredAt,
		Status:      entity.AlarmStatusActive,
	}
}

func TestAggregationStrategy_String(t *testing.T) {
	tests := []struct {
		strategy AggregationStrategy
		expected string
	}{
		{StrategyNone, "none"},
		{StrategyByDevice, "by_device"},
		{StrategyByStation, "by_station"},
		{StrategyByType, "by_type"},
		{StrategyByLevel, "by_level"},
		{StrategyByTimeWindow, "by_time_window"},
		{StrategyByDeviceAndType, "by_device_and_type"},
		{StrategyByStationAndLevel, "by_station_and_level"},
		{AggregationStrategy(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.strategy.String())
		})
	}
}

func TestDefaultAggregationConfig(t *testing.T) {
	config := DefaultAggregationConfig()

	assert.Equal(t, StrategyByDevice, config.Strategy)
	assert.Equal(t, 5*time.Minute, config.WindowDuration)
	assert.Equal(t, 100, config.MaxGroupSize)
	assert.Equal(t, 1, config.MinGroupSize)
	assert.Equal(t, 30*time.Second, config.FlushInterval)
	assert.True(t, config.EnableAutoFlush)
}

func TestNewAggregatedAlarm(t *testing.T) {
	strategy := StrategyByDevice
	groupKey := "device-001"

	agg := NewAggregatedAlarm(strategy, groupKey)

	assert.NotNil(t, agg)
	assert.Equal(t, strategy, agg.Strategy)
	assert.Equal(t, groupKey, agg.GroupKey)
	assert.NotNil(t, agg.Alarms)
	assert.Equal(t, 0, agg.Count)
	assert.False(t, agg.FirstTriggeredAt.IsZero())
	assert.False(t, agg.LastTriggeredAt.IsZero())
	assert.Equal(t, entity.AlarmLevelInfo, agg.MaxLevel)
	assert.NotNil(t, agg.Metadata)
}

func TestAggregatedAlarm_Add(t *testing.T) {
	now := time.Now()
	agg := NewAggregatedAlarm(StrategyByDevice, "device-001")

	alarm1 := createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now.Add(-1*time.Hour))
	alarm2 := createTestAlarm("alarm-002", "device-001", "station-001", entity.AlarmLevelCritical, now)
	alarm3 := createTestAlarm("alarm-003", "device-001", "station-001", entity.AlarmLevelMajor, now.Add(-30*time.Minute))

	agg.Add(alarm1)
	assert.Equal(t, 1, agg.Count)
	// FirstTriggeredAt 和 LastTriggeredAt 应该被更新为 alarm1.TriggeredAt
	assert.True(t, agg.FirstTriggeredAt.Equal(alarm1.TriggeredAt) || agg.FirstTriggeredAt.Before(alarm1.TriggeredAt))
	assert.True(t, agg.LastTriggeredAt.Equal(alarm1.TriggeredAt) || agg.LastTriggeredAt.After(alarm1.TriggeredAt))
	assert.Equal(t, entity.AlarmLevelWarning, agg.MaxLevel)

	agg.Add(alarm2)
	assert.Equal(t, 2, agg.Count)
	// FirstTriggeredAt 应该是最早的时间
	assert.True(t, agg.FirstTriggeredAt.Equal(alarm1.TriggeredAt) || agg.FirstTriggeredAt.Before(alarm1.TriggeredAt))
	// LastTriggeredAt 应该是最晚的时间
	assert.True(t, agg.LastTriggeredAt.Equal(alarm2.TriggeredAt) || agg.LastTriggeredAt.After(alarm2.TriggeredAt))
	assert.Equal(t, entity.AlarmLevelCritical, agg.MaxLevel)

	agg.Add(alarm3)
	assert.Equal(t, 3, agg.Count)
	// FirstTriggeredAt 应该是最早的时间（alarm1）
	assert.True(t, agg.FirstTriggeredAt.Equal(alarm1.TriggeredAt) || agg.FirstTriggeredAt.Before(alarm1.TriggeredAt))
	// LastTriggeredAt 应该是最晚的时间（alarm2）
	assert.True(t, agg.LastTriggeredAt.Equal(alarm2.TriggeredAt) || agg.LastTriggeredAt.After(alarm2.TriggeredAt))
	assert.Equal(t, entity.AlarmLevelCritical, agg.MaxLevel)
}

func TestNewAggregator(t *testing.T) {
	config := AggregationConfig{
		Strategy:        StrategyByDevice,
		WindowDuration:  5 * time.Minute,
		MaxGroupSize:    100,
		MinGroupSize:    1,
		FlushInterval:   30 * time.Second,
		EnableAutoFlush: true,
	}

	agg := NewAggregator(config)

	assert.NotNil(t, agg)
	assert.NotNil(t, agg.groups)
	assert.False(t, agg.windowStart.IsZero())
	assert.False(t, agg.windowEnd.IsZero())
}

func TestNewAggregator_DefaultValues(t *testing.T) {
	config := AggregationConfig{
		Strategy: StrategyByDevice,
		// WindowDuration 和 FlushInterval 为 0
	}

	agg := NewAggregator(config)

	assert.NotNil(t, agg)
	assert.Equal(t, 5*time.Minute, agg.config.WindowDuration)
	assert.Equal(t, 30*time.Second, agg.config.FlushInterval)
}

func TestAggregator_Aggregate_ByDevice(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)
	ctx := context.Background()

	now := time.Now()
	alarm1 := createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now)
	alarm2 := createTestAlarm("alarm-002", "device-001", "station-001", entity.AlarmLevelMajor, now)
	alarm3 := createTestAlarm("alarm-003", "device-002", "station-001", entity.AlarmLevelCritical, now)

	// 第一个告警
	result, isNew := agg.Aggregate(ctx, alarm1)
	assert.NotNil(t, result)
	assert.True(t, isNew)
	assert.Equal(t, 1, result.Count)

	// 同一设备的第二个告警
	result, isNew = agg.Aggregate(ctx, alarm2)
	assert.NotNil(t, result)
	assert.False(t, isNew)
	assert.Equal(t, 2, result.Count)

	// 不同设备的告警
	result, isNew = agg.Aggregate(ctx, alarm3)
	assert.NotNil(t, result)
	assert.True(t, isNew)
	assert.Equal(t, 1, result.Count)

	// 验证分组数量
	assert.Equal(t, 2, agg.GetGroupCount())
	assert.Equal(t, 3, agg.GetTotalAlarmCount())
}

func TestAggregator_Aggregate_ByStation(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByStation,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)
	ctx := context.Background()

	now := time.Now()
	alarm1 := createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now)
	alarm2 := createTestAlarm("alarm-002", "device-002", "station-001", entity.AlarmLevelMajor, now)
	alarm3 := createTestAlarm("alarm-003", "device-003", "station-002", entity.AlarmLevelCritical, now)

	agg.Aggregate(ctx, alarm1)
	agg.Aggregate(ctx, alarm2)
	agg.Aggregate(ctx, alarm3)

	assert.Equal(t, 2, agg.GetGroupCount())
	assert.Equal(t, 3, agg.GetTotalAlarmCount())
}

func TestAggregator_Aggregate_ByType(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByType,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)
	ctx := context.Background()

	now := time.Now()
	alarm1 := createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now)
	alarm1.Type = entity.AlarmTypeDevice
	alarm2 := createTestAlarm("alarm-002", "device-002", "station-001", entity.AlarmLevelMajor, now)
	alarm2.Type = entity.AlarmTypeDevice
	alarm3 := createTestAlarm("alarm-003", "device-003", "station-002", entity.AlarmLevelCritical, now)
	alarm3.Type = entity.AlarmTypeSystem

	agg.Aggregate(ctx, alarm1)
	agg.Aggregate(ctx, alarm2)
	agg.Aggregate(ctx, alarm3)

	assert.Equal(t, 2, agg.GetGroupCount())
}

func TestAggregator_Aggregate_ByLevel(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByLevel,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)
	ctx := context.Background()

	now := time.Now()
	alarm1 := createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now)
	alarm2 := createTestAlarm("alarm-002", "device-002", "station-001", entity.AlarmLevelWarning, now)
	alarm3 := createTestAlarm("alarm-003", "device-003", "station-002", entity.AlarmLevelCritical, now)

	agg.Aggregate(ctx, alarm1)
	agg.Aggregate(ctx, alarm2)
	agg.Aggregate(ctx, alarm3)

	assert.Equal(t, 2, agg.GetGroupCount())
}

func TestAggregator_Aggregate_MaxGroupSize(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
		MaxGroupSize: 2,
	}

	agg := NewAggregator(config)
	ctx := context.Background()

	now := time.Now()

	// 添加第一个告警
	alarm1 := createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now)
	agg.Aggregate(ctx, alarm1)

	// 添加第二个告警
	alarm2 := createTestAlarm("alarm-002", "device-001", "station-001", entity.AlarmLevelMajor, now)
	agg.Aggregate(ctx, alarm2)

	// 添加第三个告警，应该触发刷新
	alarm3 := createTestAlarm("alarm-003", "device-001", "station-001", entity.AlarmLevelCritical, now)
	agg.Aggregate(ctx, alarm3)

	// 验证分组已刷新并重新创建
	group := agg.GetGroup("device-001")
	assert.NotNil(t, group)
	assert.Equal(t, 1, group.Count)
}

func TestAggregator_AggregateBatch(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)
	ctx := context.Background()

	now := time.Now()
	alarms := []*entity.Alarm{
		createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now),
		createTestAlarm("alarm-002", "device-001", "station-001", entity.AlarmLevelMajor, now),
		createTestAlarm("alarm-003", "device-002", "station-001", entity.AlarmLevelCritical, now),
	}

	results := agg.AggregateBatch(ctx, alarms)

	assert.Len(t, results, 3)
	assert.Equal(t, 2, agg.GetGroupCount())
}

func TestAggregator_GetGroup(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)
	ctx := context.Background()

	now := time.Now()
	alarm := createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now)
	agg.Aggregate(ctx, alarm)

	group := agg.GetGroup("device-001")
	assert.NotNil(t, group)
	assert.Equal(t, 1, group.Count)

	// 不存在的分组
	group = agg.GetGroup("device-999")
	assert.Nil(t, group)
}

func TestAggregator_GetAllGroups(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)
	ctx := context.Background()

	now := time.Now()
	agg.Aggregate(ctx, createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now))
	agg.Aggregate(ctx, createTestAlarm("alarm-002", "device-002", "station-001", entity.AlarmLevelMajor, now))

	groups := agg.GetAllGroups()
	assert.Len(t, groups, 2)
}

func TestAggregator_GetStats(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)
	ctx := context.Background()

	now := time.Now()
	agg.Aggregate(ctx, createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now))
	agg.Aggregate(ctx, createTestAlarm("alarm-002", "device-001", "station-001", entity.AlarmLevelMajor, now))
	agg.Aggregate(ctx, createTestAlarm("alarm-003", "device-002", "station-001", entity.AlarmLevelCritical, now))

	stats := agg.GetStats()

	assert.Equal(t, 2, stats.TotalGroups)
	assert.Equal(t, 3, stats.TotalAlarms)
	assert.Equal(t, 2, stats.MaxGroupSize)
	assert.Equal(t, 1, stats.MinGroupSize)
	assert.Equal(t, 1.5, stats.AvgGroupSize)
	assert.False(t, stats.WindowStartTime.IsZero())
	assert.False(t, stats.WindowEndTime.IsZero())
}

func TestAggregator_GetStats_Empty(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)

	stats := agg.GetStats()

	assert.Equal(t, 0, stats.TotalGroups)
	assert.Equal(t, 0, stats.TotalAlarms)
}

func TestAggregator_Flush(t *testing.T) {
	flushedGroups := make([]*AggregatedAlarm, 0)
	var mu sync.Mutex

	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
		MinGroupSize: 1,
	}

	agg := NewAggregator(config)
	agg.AddHandler(func(ctx context.Context, aggregated *AggregatedAlarm) error {
		mu.Lock()
		defer mu.Unlock()
		flushedGroups = append(flushedGroups, aggregated)
		return nil
	})

	ctx := context.Background()
	now := time.Now()

	agg.Aggregate(ctx, createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now))
	agg.Aggregate(ctx, createTestAlarm("alarm-002", "device-002", "station-001", entity.AlarmLevelMajor, now))

	err := agg.Flush(ctx)
	assert.NoError(t, err)

	mu.Lock()
	assert.Len(t, flushedGroups, 2)
	mu.Unlock()

	// 验证分组已清空
	assert.Equal(t, 0, agg.GetGroupCount())
}

func TestAggregator_Clear(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)
	ctx := context.Background()

	now := time.Now()
	agg.Aggregate(ctx, createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now))
	agg.Aggregate(ctx, createTestAlarm("alarm-002", "device-002", "station-001", entity.AlarmLevelMajor, now))

	assert.Equal(t, 2, agg.GetGroupCount())

	agg.Clear()

	assert.Equal(t, 0, agg.GetGroupCount())
}

func TestAggregator_RemoveGroup(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)
	ctx := context.Background()

	now := time.Now()
	agg.Aggregate(ctx, createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now))
	agg.Aggregate(ctx, createTestAlarm("alarm-002", "device-002", "station-001", entity.AlarmLevelMajor, now))

	assert.Equal(t, 2, agg.GetGroupCount())

	agg.RemoveGroup("device-001")

	assert.Equal(t, 1, agg.GetGroupCount())
	assert.Nil(t, agg.GetGroup("device-001"))
	assert.NotNil(t, agg.GetGroup("device-002"))
}

func TestAggregator_SetStrategy(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)

	agg.SetStrategy(StrategyByStation)

	assert.Equal(t, StrategyByStation, agg.GetConfig().Strategy)
}

func TestAggregator_SetWindowDuration(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)

	newDuration := 10 * time.Minute
	agg.SetWindowDuration(newDuration)

	assert.Equal(t, newDuration, agg.GetConfig().WindowDuration)
}

func TestAggregator_StartStop(t *testing.T) {
	config := AggregationConfig{
		Strategy:        StrategyByDevice,
		WindowDuration:  5 * time.Minute,
		FlushInterval:   100 * time.Millisecond,
		EnableAutoFlush: true,
	}

	agg := NewAggregator(config)
	ctx := context.Background()

	agg.Start(ctx)

	// 等待一段时间
	time.Sleep(150 * time.Millisecond)

	agg.Stop()
}

func TestAggregator_TriggerFlush(t *testing.T) {
	flushed := false

	config := AggregationConfig{
		Strategy:        StrategyByDevice,
		WindowDuration:  5 * time.Minute,
		FlushInterval:   1 * time.Hour, // 很长的间隔
		EnableAutoFlush: true,
		MinGroupSize:    1,
	}

	agg := NewAggregator(config)
	agg.AddHandler(func(ctx context.Context, aggregated *AggregatedAlarm) error {
		flushed = true
		return nil
	})

	ctx := context.Background()
	agg.Start(ctx)
	defer agg.Stop()

	now := time.Now()
	agg.Aggregate(ctx, createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now))

	// 触发刷新
	agg.TriggerFlush()

	// 等待刷新完成
	time.Sleep(100 * time.Millisecond)

	assert.True(t, flushed)
}

func TestAggregator_AddHandler(t *testing.T) {
	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
	}

	agg := NewAggregator(config)

	handler1 := func(ctx context.Context, aggregated *AggregatedAlarm) error { return nil }
	handler2 := func(ctx context.Context, aggregated *AggregatedAlarm) error { return nil }

	agg.AddHandler(handler1)
	agg.AddHandler(handler2)

	assert.Len(t, agg.handlers, 2)
}

func TestNewMultiStrategyAggregator(t *testing.T) {
	configs := map[AggregationStrategy]AggregationConfig{
		StrategyByDevice: {
			WindowDuration: 5 * time.Minute,
			MaxGroupSize:   100,
		},
		StrategyByStation: {
			WindowDuration: 10 * time.Minute,
			MaxGroupSize:   200,
		},
	}

	multi := NewMultiStrategyAggregator(configs)

	assert.NotNil(t, multi)
	assert.Len(t, multi.aggregators, 2)
}

func TestMultiStrategyAggregator_Aggregate(t *testing.T) {
	configs := map[AggregationStrategy]AggregationConfig{
		StrategyByDevice: {WindowDuration: 5 * time.Minute},
		StrategyByStation: {WindowDuration: 5 * time.Minute},
	}

	multi := NewMultiStrategyAggregator(configs)
	ctx := context.Background()

	now := time.Now()
	alarm := createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now)

	// 按设备聚合
	result, ok := multi.Aggregate(ctx, StrategyByDevice, alarm)
	assert.NotNil(t, result)
	assert.True(t, ok)

	// 按站点聚合
	result, ok = multi.Aggregate(ctx, StrategyByStation, alarm)
	assert.NotNil(t, result)
	assert.True(t, ok)

	// 不存在的策略
	result, ok = multi.Aggregate(ctx, StrategyByLevel, alarm)
	assert.Nil(t, result)
	assert.False(t, ok)
}

func TestMultiStrategyAggregator_AggregateAll(t *testing.T) {
	configs := map[AggregationStrategy]AggregationConfig{
		StrategyByDevice: {WindowDuration: 5 * time.Minute},
		StrategyByStation: {WindowDuration: 5 * time.Minute},
		StrategyByType: {WindowDuration: 5 * time.Minute},
	}

	multi := NewMultiStrategyAggregator(configs)
	ctx := context.Background()

	now := time.Now()
	alarm := createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now)

	results := multi.AggregateAll(ctx, alarm)

	assert.Len(t, results, 3)
	assert.NotNil(t, results[StrategyByDevice])
	assert.NotNil(t, results[StrategyByStation])
	assert.NotNil(t, results[StrategyByType])
}

func TestMultiStrategyAggregator_GetAggregator(t *testing.T) {
	configs := map[AggregationStrategy]AggregationConfig{
		StrategyByDevice: {WindowDuration: 5 * time.Minute},
	}

	multi := NewMultiStrategyAggregator(configs)

	agg := multi.GetAggregator(StrategyByDevice)
	assert.NotNil(t, agg)

	agg = multi.GetAggregator(StrategyByStation)
	assert.Nil(t, agg)
}

func TestMultiStrategyAggregator_StartStop(t *testing.T) {
	configs := map[AggregationStrategy]AggregationConfig{
		StrategyByDevice: {
			WindowDuration:  5 * time.Minute,
			FlushInterval:   100 * time.Millisecond,
			EnableAutoFlush: true,
		},
		StrategyByStation: {
			WindowDuration:  5 * time.Minute,
			FlushInterval:   100 * time.Millisecond,
			EnableAutoFlush: true,
		},
	}

	multi := NewMultiStrategyAggregator(configs)
	ctx := context.Background()

	multi.Start(ctx)

	time.Sleep(150 * time.Millisecond)

	multi.Stop()
}

func TestMultiStrategyAggregator_FlushAll(t *testing.T) {
	flushCount := 0
	var mu sync.Mutex

	handler := func(ctx context.Context, aggregated *AggregatedAlarm) error {
		mu.Lock()
		defer mu.Unlock()
		flushCount++
		return nil
	}

	configs := map[AggregationStrategy]AggregationConfig{
		StrategyByDevice: {WindowDuration: 5 * time.Minute, MinGroupSize: 1},
		StrategyByStation: {WindowDuration: 5 * time.Minute, MinGroupSize: 1},
	}

	multi := NewMultiStrategyAggregator(configs)

	for _, agg := range multi.aggregators {
		agg.AddHandler(handler)
	}

	ctx := context.Background()
	now := time.Now()
	alarm := createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now)

	multi.AggregateAll(ctx, alarm)

	multi.FlushAll(ctx)

	mu.Lock()
	assert.Equal(t, 2, flushCount)
	mu.Unlock()
}

func TestAggregator_MinGroupSize(t *testing.T) {
	flushedGroups := make([]*AggregatedAlarm, 0)
	var mu sync.Mutex

	config := AggregationConfig{
		Strategy:     StrategyByDevice,
		WindowDuration: 5 * time.Minute,
		MinGroupSize: 2, // 至少需要2个告警才刷新
	}

	agg := NewAggregator(config)
	agg.AddHandler(func(ctx context.Context, aggregated *AggregatedAlarm) error {
		mu.Lock()
		defer mu.Unlock()
		flushedGroups = append(flushedGroups, aggregated)
		return nil
	})

	ctx := context.Background()
	now := time.Now()

	// 只添加1个告警
	agg.Aggregate(ctx, createTestAlarm("alarm-001", "device-001", "station-001", entity.AlarmLevelWarning, now))

	// 刷新
	agg.Flush(ctx)

	// 不应该刷新（不满足最小分组大小）
	mu.Lock()
	assert.Len(t, flushedGroups, 0)
	mu.Unlock()

	// 添加第二个告警
	agg.Aggregate(ctx, createTestAlarm("alarm-002", "device-001", "station-001", entity.AlarmLevelMajor, now))

	// 再次刷新
	agg.Flush(ctx)

	// 现在应该刷新了
	mu.Lock()
	assert.Len(t, flushedGroups, 1)
	mu.Unlock()
}

func TestAggregator_GetConfig(t *testing.T) {
	config := AggregationConfig{
		Strategy:        StrategyByDevice,
		WindowDuration:  5 * time.Minute,
		MaxGroupSize:    100,
		MinGroupSize:    1,
		FlushInterval:   30 * time.Second,
		EnableAutoFlush: true,
	}

	agg := NewAggregator(config)

	retrievedConfig := agg.GetConfig()

	assert.Equal(t, config.Strategy, retrievedConfig.Strategy)
	assert.Equal(t, config.WindowDuration, retrievedConfig.WindowDuration)
	assert.Equal(t, config.MaxGroupSize, retrievedConfig.MaxGroupSize)
	assert.Equal(t, config.MinGroupSize, retrievedConfig.MinGroupSize)
	assert.Equal(t, config.FlushInterval, retrievedConfig.FlushInterval)
	assert.Equal(t, config.EnableAutoFlush, retrievedConfig.EnableAutoFlush)
}
