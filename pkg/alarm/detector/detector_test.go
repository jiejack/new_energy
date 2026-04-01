package detector

import (
	"context"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewRule(t *testing.T) {
	rule := NewRule("rule001", "电压高限告警", RuleTypeThreshold)

	assert.Equal(t, "rule001", rule.ID)
	assert.Equal(t, "电压高限告警", rule.Name)
	assert.Equal(t, RuleTypeThreshold, rule.Type)
	assert.True(t, rule.Enabled)
	assert.NotNil(t, rule.Metadata)
	assert.NotZero(t, rule.CreatedAt)
	assert.NotZero(t, rule.UpdatedAt)
}

func TestRuleType_String(t *testing.T) {
	tests := []struct {
		ruleType RuleType
		expected string
	}{
		{RuleTypeThreshold, "threshold"},
		{RuleTypeRange, "range"},
		{RuleTypeRate, "rate"},
		{RuleTypeDeviation, "deviation"},
		{RuleTypeDuration, "duration"},
		{RuleTypeExpression, "expression"},
		{RuleType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ruleType.String())
		})
	}
}

func TestOperator_String(t *testing.T) {
	tests := []struct {
		operator Operator
		expected string
	}{
		{OpEqual, "=="},
		{OpNotEqual, "!="},
		{OpGreaterThan, ">"},
		{OpGreaterEqual, ">="},
		{OpLessThan, "<"},
		{OpLessEqual, "<="},
		{Operator(999), "?"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.operator.String())
		})
	}
}

func TestNewSlidingWindow(t *testing.T) {
	window := NewSlidingWindow("point001", 5*time.Minute, 100)

	assert.Equal(t, "point001", window.pointID)
	assert.Equal(t, 5*time.Minute, window.duration)
	assert.Equal(t, 100, window.maxCount)
	assert.NotNil(t, window.data)
}

func TestSlidingWindow_Add(t *testing.T) {
	window := NewSlidingWindow("point001", 5*time.Minute, 100)

	now := time.Now()
	window.Add(10.5, now)
	window.Add(20.3, now.Add(1*time.Second))
	window.Add(30.1, now.Add(2*time.Second))

	values := window.GetValues()
	assert.Len(t, values, 3)
	assert.Equal(t, 10.5, values[0])
	assert.Equal(t, 20.3, values[1])
	assert.Equal(t, 30.1, values[2])
}

func TestSlidingWindow_GetStats(t *testing.T) {
	window := NewSlidingWindow("point001", 5*time.Minute, 100)

	now := time.Now()
	window.Add(10.0, now)
	window.Add(20.0, now.Add(1*time.Second))
	window.Add(30.0, now.Add(2*time.Second))
	window.Add(40.0, now.Add(3*time.Second))
	window.Add(50.0, now.Add(4*time.Second))

	stats := window.GetStats()

	assert.Equal(t, 5, stats.Count)
	assert.Equal(t, 150.0, stats.Sum)
	assert.Equal(t, 30.0, stats.Avg)
	assert.Equal(t, 10.0, stats.Min)
	assert.Equal(t, 50.0, stats.Max)
	assert.Greater(t, stats.StdDev, 0.0)
}

func TestSlidingWindow_GetStats_Empty(t *testing.T) {
	window := NewSlidingWindow("point001", 5*time.Minute, 100)

	stats := window.GetStats()

	assert.Equal(t, 0, stats.Count)
	assert.Equal(t, 0.0, stats.Sum)
	assert.Equal(t, 0.0, stats.Avg)
}

func TestSlidingWindow_MaxCount(t *testing.T) {
	window := NewSlidingWindow("point001", 5*time.Minute, 3)

	now := time.Now()
	window.Add(10.0, now)
	window.Add(20.0, now.Add(1*time.Second))
	window.Add(30.0, now.Add(2*time.Second))
	window.Add(40.0, now.Add(3*time.Second)) // 应该挤出第一个

	values := window.GetValues()
	assert.Len(t, values, 3)
	assert.Equal(t, 20.0, values[0])
	assert.Equal(t, 30.0, values[1])
	assert.Equal(t, 40.0, values[2])
}

func TestNewDetector(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewDetector(config)

	assert.NotNil(t, detector)
	assert.Equal(t, config, detector.config)
	assert.NotNil(t, detector.rules)
	assert.NotNil(t, detector.windows)
	assert.NotNil(t, detector.dataChan)
	assert.NotNil(t, detector.resultChan)
}

func TestDetector_AddRule(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "测试规则", RuleTypeThreshold)

	detector.AddRule(rule)

	assert.Equal(t, 1, detector.GetRuleCount())
	assert.Equal(t, rule, detector.GetRule("rule001"))
}

func TestDetector_RemoveRule(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "测试规则", RuleTypeThreshold)

	detector.AddRule(rule)
	assert.Equal(t, 1, detector.GetRuleCount())

	detector.RemoveRule("rule001")
	assert.Equal(t, 0, detector.GetRuleCount())
	assert.Nil(t, detector.GetRule("rule001"))
}

func TestDetector_GetAllRules(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule1 := NewRule("rule001", "规则1", RuleTypeThreshold)
	rule2 := NewRule("rule002", "规则2", RuleTypeRange)

	detector.AddRule(rule1)
	detector.AddRule(rule2)

	rules := detector.GetAllRules()
	assert.Len(t, rules, 2)
}

func TestDetector_DetectThreshold(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "电压高限", RuleTypeThreshold)
	rule.Operator = OpGreaterThan
	rule.Threshold = 400.0
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit
	rule.Title = "电压高限告警"
	rule.PointIDs = []string{"point001"}

	detector.AddRule(rule)

	point := &DataPoint{
		PointID:   "point001",
		DeviceID:  "device001",
		StationID: "station001",
		Value:     450.0,
		Timestamp: time.Now(),
	}

	// 同步检测
	result := detector.detectRule(context.Background(), rule, point)

	assert.NotNil(t, result)
	assert.True(t, result.Triggered)
	assert.Equal(t, 450.0, result.Value)
	assert.Equal(t, 400.0, result.Threshold)
	assert.NotNil(t, result.Alarm)
}

func TestDetector_DetectRange(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "电压范围", RuleTypeRange)
	rule.MinValue = 200.0
	rule.MaxValue = 400.0
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit
	rule.Title = "电压越限告警"
	rule.PointIDs = []string{"point001"}

	detector.AddRule(rule)

	tests := []struct {
		name      string
		value     float64
		triggered bool
	}{
		{"值在范围内", 300.0, false},
		{"值低于下限", 150.0, true},
		{"值高于上限", 450.0, true},
		{"值等于下限", 200.0, false},
		{"值等于上限", 400.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			point := &DataPoint{
				PointID:   "point001",
				DeviceID:  "device001",
				StationID: "station001",
				Value:     tt.value,
				Timestamp: time.Now(),
			}

			result := detector.detectRule(context.Background(), rule, point)

			assert.NotNil(t, result)
			assert.Equal(t, tt.triggered, result.Triggered)
		})
	}
}

func TestDetector_DetectRate(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "变化率告警", RuleTypeRate)
	rule.RateThreshold = 10.0 // 每秒变化率阈值
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit
	rule.Title = "变化率过大"
	rule.PointIDs = []string{"point001"}

	detector.AddRule(rule)

	// 先添加一些数据到窗口
	now := time.Now()
	for i := 0; i < 10; i++ {
		point := &DataPoint{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Value:     float64(i * 100), // 快速增长
			Timestamp: now.Add(time.Duration(i) * time.Second),
		}
		detector.updateWindow(point)
	}

	// 检测变化率
	testPoint := &DataPoint{
		PointID:   "point001",
		DeviceID:  "device001",
		StationID: "station001",
		Value:     1000.0,
		Timestamp: now.Add(10 * time.Second),
	}

	result := detector.detectRule(context.Background(), rule, testPoint)

	assert.NotNil(t, result)
	// 变化率应该很大，触发告警
	assert.True(t, result.Triggered)
}

func TestDetector_DetectDeviation(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "偏差告警", RuleTypeDeviation)
	rule.DeviationThreshold = 50.0
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit
	rule.Title = "偏差过大"
	rule.PointIDs = []string{"point001"}

	detector.AddRule(rule)

	// 添加稳定的数据
	now := time.Now()
	for i := 0; i < 10; i++ {
		point := &DataPoint{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Value:     100.0 + float64(i%3), // 在100附近波动
			Timestamp: now.Add(time.Duration(i) * time.Second),
		}
		detector.updateWindow(point)
	}

	// 添加一个偏差很大的点
	testPoint := &DataPoint{
		PointID:   "point001",
		DeviceID:  "device001",
		StationID: "station001",
		Value:     200.0, // 偏差约100
		Timestamp: now.Add(10 * time.Second),
	}

	result := detector.detectRule(context.Background(), rule, testPoint)

	assert.NotNil(t, result)
	assert.True(t, result.Triggered)
}

func TestDetector_GetStats(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())

	stats := detector.GetStats()

	assert.Equal(t, int64(0), stats.TotalProcessed)
	assert.Equal(t, int64(0), stats.TotalTriggered)
	assert.Equal(t, int64(0), stats.TotalSuppressed)
}

func TestDetector_ClearWindows(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())

	// 添加一些窗口
	point := &DataPoint{
		PointID:   "point001",
		Value:     100.0,
		Timestamp: time.Now(),
	}
	detector.updateWindow(point)

	assert.Equal(t, 1, detector.GetWindowCount())

	detector.ClearWindows()

	assert.Equal(t, 0, detector.GetWindowCount())
}

func TestDetector_MatchingRules(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())

	// 添加多个规则
	rule1 := NewRule("rule001", "规则1", RuleTypeThreshold)
	rule1.PointIDs = []string{"point001"}
	rule1.Enabled = true

	rule2 := NewRule("rule002", "规则2", RuleTypeThreshold)
	rule2.PointIDs = []string{"point002"}
	rule2.Enabled = true

	rule3 := NewRule("rule003", "规则3", RuleTypeThreshold)
	rule3.PointIDs = []string{"point001", "point002"}
	rule3.Enabled = true

	rule4 := NewRule("rule004", "禁用规则", RuleTypeThreshold)
	rule4.PointIDs = []string{"point001"}
	rule4.Enabled = false

	detector.AddRule(rule1)
	detector.AddRule(rule2)
	detector.AddRule(rule3)
	detector.AddRule(rule4)

	point := &DataPoint{
		PointID:   "point001",
		DeviceID:  "device001",
		StationID: "station001",
		Value:     100.0,
		Timestamp: time.Now(),
	}

	matching := detector.getMatchingRules(point)

	// 应该匹配rule1和rule3，不匹配rule2（测点不匹配）和rule4（禁用）
	assert.Len(t, matching, 2)
}

func TestDefaultDetectorConfig(t *testing.T) {
	config := DefaultDetectorConfig()

	assert.Equal(t, 8, config.WorkerCount)
	assert.Equal(t, 10000, config.BufferSize)
	assert.Equal(t, 5*time.Minute, config.WindowDuration)
	assert.Equal(t, 100, config.MaxWindowsPerPoint)
	assert.True(t, config.EnableDedup)
	assert.True(t, config.EnableAggregator)
}
