package detector

import (
	"context"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	window.Add(40.0, now.Add(3*time.Second))
	values := window.GetValues()
	assert.Len(t, values, 3)
	assert.Equal(t, 20.0, values[0])
	assert.Equal(t, 30.0, values[1])
	assert.Equal(t, 40.0, values[2])
}

func TestSlidingWindow_ExpiredData(t *testing.T) {
	window := NewSlidingWindow("point001", 10*time.Second, 100)
	now := time.Now()
	window.Add(10.0, now)
	window.Add(20.0, now.Add(1*time.Second))
	window.Add(30.0, now.Add(2*time.Second))
	window.Add(40.0, now.Add(15*time.Second))
	values := window.GetValues()
	assert.Len(t, values, 1)
	assert.Equal(t, 40.0, values[0])
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

func TestNewDetector_ZeroConfig(t *testing.T) {
	config := DetectorConfig{}
	detector := NewDetector(config)
	assert.NotNil(t, detector)
	assert.Equal(t, 8, detector.config.WorkerCount)
	assert.Equal(t, 10000, detector.config.BufferSize)
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

func TestDetector_DetectThreshold_AllOperators(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "阈值规则", RuleTypeThreshold)
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit
	rule.Title = "阈值告警"
	rule.PointIDs = []string{"point001"}

	tests := []struct {
		name      string
		operator  Operator
		threshold float64
		value     float64
		triggered bool
	}{
		{"OpEqual_triggered", OpEqual, 100.0, 100.0, true},
		{"OpEqual_not_triggered", OpEqual, 100.0, 99.0, false},
		{"OpNotEqual_triggered", OpNotEqual, 100.0, 99.0, true},
		{"OpNotEqual_not_triggered", OpNotEqual, 100.0, 100.0, false},
		{"OpGreaterThan_triggered", OpGreaterThan, 100.0, 101.0, true},
		{"OpGreaterThan_not_triggered", OpGreaterThan, 100.0, 100.0, false},
		{"OpGreaterEqual_triggered_eq", OpGreaterEqual, 100.0, 100.0, true},
		{"OpGreaterEqual_triggered_gt", OpGreaterEqual, 100.0, 101.0, true},
		{"OpGreaterEqual_not_triggered", OpGreaterEqual, 100.0, 99.0, false},
		{"OpLessThan_triggered", OpLessThan, 100.0, 99.0, true},
		{"OpLessThan_not_triggered", OpLessThan, 100.0, 100.0, false},
		{"OpLessEqual_triggered_eq", OpLessEqual, 100.0, 100.0, true},
		{"OpLessEqual_triggered_lt", OpLessEqual, 100.0, 99.0, true},
		{"OpLessEqual_not_triggered", OpLessEqual, 100.0, 101.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule.Operator = tt.operator
			rule.Threshold = tt.threshold
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

func TestDetector_DetectThreshold_UnknownOperator(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "阈值规则", RuleTypeThreshold)
	rule.Operator = Operator(999)
	rule.Threshold = 100.0
	rule.PointIDs = []string{"point001"}
	point := &DataPoint{
		PointID:   "point001",
		Value:     200.0,
		Timestamp: time.Now(),
	}
	triggered, threshold := detector.detectThreshold(rule, point)
	assert.False(t, triggered)
	assert.Equal(t, 100.0, threshold)
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

func TestDetector_DetectRange_ThresholdValues(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "范围规则", RuleTypeRange)
	rule.MinValue = 200.0
	rule.MaxValue = 400.0
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit
	rule.Title = "范围告警"
	rule.PointIDs = []string{"point001"}

	point := &DataPoint{PointID: "point001", Value: 150.0, Timestamp: time.Now()}
	triggered, threshold := detector.detectRange(rule, point)
	assert.True(t, triggered)
	assert.Equal(t, 200.0, threshold)

	point.Value = 450.0
	triggered, threshold = detector.detectRange(rule, point)
	assert.True(t, triggered)
	assert.Equal(t, 400.0, threshold)
}

func TestDetector_DetectRate(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "变化率告警", RuleTypeRate)
	rule.RateThreshold = 10.0
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit
	rule.Title = "变化率过大"
	rule.PointIDs = []string{"point001"}
	detector.AddRule(rule)

	now := time.Now()
	for i := 0; i < 10; i++ {
		point := &DataPoint{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Value:     float64(i * 100),
			Timestamp: now.Add(time.Duration(i) * time.Second),
		}
		detector.updateWindow(point)
	}

	testPoint := &DataPoint{
		PointID:   "point001",
		DeviceID:  "device001",
		StationID: "station001",
		Value:     1000.0,
		Timestamp: now.Add(10 * time.Second),
	}
	result := detector.detectRule(context.Background(), rule, testPoint)
	assert.NotNil(t, result)
	assert.True(t, result.Triggered)
}

func TestDetector_DetectRate_NoWindow(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "变化率告警", RuleTypeRate)
	rule.RateThreshold = 10.0
	rule.PointIDs = []string{"point001"}

	point := &DataPoint{PointID: "point001", Value: 100.0, Timestamp: time.Now()}
	triggered, threshold := detector.detectRate(rule, point)
	assert.False(t, triggered)
	assert.Equal(t, 0.0, threshold)
}

func TestDetector_DetectRate_InsufficientData(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "变化率告警", RuleTypeRate)
	rule.RateThreshold = 10.0
	rule.PointIDs = []string{"point001"}

	point := &DataPoint{PointID: "point001", Value: 100.0, Timestamp: time.Now()}
	detector.updateWindow(point)

	testPoint := &DataPoint{PointID: "point001", Value: 200.0, Timestamp: time.Now()}
	triggered, _ := detector.detectRate(rule, testPoint)
	assert.False(t, triggered)
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

	now := time.Now()
	for i := 0; i < 10; i++ {
		point := &DataPoint{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Value:     100.0 + float64(i%3),
			Timestamp: now.Add(time.Duration(i) * time.Second),
		}
		detector.updateWindow(point)
	}

	testPoint := &DataPoint{
		PointID:   "point001",
		DeviceID:  "device001",
		StationID: "station001",
		Value:     200.0,
		Timestamp: now.Add(10 * time.Second),
	}
	result := detector.detectRule(context.Background(), rule, testPoint)
	assert.NotNil(t, result)
	assert.True(t, result.Triggered)
}

func TestDetector_DetectDeviation_NoWindow(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "偏差告警", RuleTypeDeviation)
	rule.DeviationThreshold = 50.0
	rule.PointIDs = []string{"point001"}

	point := &DataPoint{PointID: "point001", Value: 200.0, Timestamp: time.Now()}
	triggered, _ := detector.detectDeviation(rule, point)
	assert.False(t, triggered)
}

func TestDetector_DetectDeviation_InsufficientData(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "偏差告警", RuleTypeDeviation)
	rule.DeviationThreshold = 50.0
	rule.PointIDs = []string{"point001"}

	point := &DataPoint{PointID: "point001", Value: 100.0, Timestamp: time.Now()}
	detector.updateWindow(point)

	testPoint := &DataPoint{PointID: "point001", Value: 200.0, Timestamp: time.Now()}
	triggered, _ := detector.detectDeviation(rule, testPoint)
	assert.False(t, triggered)
}

func TestDetector_DetectDuration(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "持续时间告警", RuleTypeDuration)
	rule.DurationThreshold = 5 * time.Second
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit
	rule.Title = "持续时间过长"
	rule.PointIDs = []string{"point001"}
	detector.AddRule(rule)

	now := time.Now()
	for i := 0; i < 10; i++ {
		point := &DataPoint{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Value:     100.0,
			Timestamp: now.Add(time.Duration(i) * time.Second),
		}
		detector.updateWindow(point)
	}

	testPoint := &DataPoint{
		PointID:   "point001",
		DeviceID:  "device001",
		StationID: "station001",
		Value:     100.0,
		Timestamp: now.Add(10 * time.Second),
	}
	result := detector.detectRule(context.Background(), rule, testPoint)
	assert.NotNil(t, result)
	assert.True(t, result.Triggered)
}

func TestDetector_DetectDuration_NoWindow(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "持续时间告警", RuleTypeDuration)
	rule.DurationThreshold = 5 * time.Second
	rule.PointIDs = []string{"point001"}

	point := &DataPoint{PointID: "point001", Value: 100.0, Timestamp: time.Now()}
	triggered, _ := detector.detectDuration(rule, point)
	assert.False(t, triggered)
}

func TestDetector_DetectDuration_InsufficientData(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "持续时间告警", RuleTypeDuration)
	rule.DurationThreshold = 5 * time.Second
	rule.PointIDs = []string{"point001"}

	point := &DataPoint{PointID: "point001", Value: 100.0, Timestamp: time.Now()}
	detector.updateWindow(point)

	testPoint := &DataPoint{PointID: "point001", Value: 100.0, Timestamp: time.Now()}
	triggered, _ := detector.detectDuration(rule, testPoint)
	assert.False(t, triggered)
}

func TestDetector_DetectRule_ExpressionType(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "表达式规则", RuleTypeExpression)
	rule.PointIDs = []string{"point001"}
	point := &DataPoint{PointID: "point001", Value: 100.0, Timestamp: time.Now()}
	result := detector.detectRule(context.Background(), rule, point)
	assert.Nil(t, result)
}

func TestDetector_GenerateMessage_Custom(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "测试", RuleTypeThreshold)
	rule.Message = "自定义消息"
	point := &DataPoint{PointID: "point001", Value: 100.0, Timestamp: time.Now()}
	msg := detector.generateMessage(rule, point, 50.0)
	assert.Equal(t, "自定义消息", msg)
}

func TestDetector_GenerateMessage_Default(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "电压高限", RuleTypeThreshold)
	point := &DataPoint{PointID: "point001", Value: 450.0, Timestamp: time.Now()}
	msg := detector.generateMessage(rule, point, 400.0)
	assert.Contains(t, msg, "电压高限")
	assert.Contains(t, msg, "point001")
}

func TestDetector_GetStats(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	stats := detector.GetStats()
	assert.Equal(t, int64(0), stats.TotalProcessed)
	assert.Equal(t, int64(0), stats.TotalTriggered)
	assert.Equal(t, int64(0), stats.TotalSuppressed)
	assert.NotNil(t, stats.RuleMatches)
}

func TestDetector_ClearWindows(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	point := &DataPoint{PointID: "point001", Value: 100.0, Timestamp: time.Now()}
	detector.updateWindow(point)
	assert.Equal(t, 1, detector.GetWindowCount())
	detector.ClearWindows()
	assert.Equal(t, 0, detector.GetWindowCount())
}

func TestDetector_GetWindowStats(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	now := time.Now()
	point := &DataPoint{PointID: "point001", Value: 100.0, Timestamp: now}
	detector.updateWindow(point)
	stats := detector.GetWindowStats("point001")
	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats.Count)
}

func TestDetector_GetWindowStats_NonExistent(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	stats := detector.GetWindowStats("nonexistent")
	assert.Nil(t, stats)
}

func TestDetector_MatchingRules(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
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
	point := &DataPoint{PointID: "point001", DeviceID: "device001", StationID: "station001", Value: 100.0, Timestamp: time.Now()}
	matching := detector.getMatchingRules(point)
	assert.Len(t, matching, 2)
}

func TestDetector_MatchingRules_DeviceFilter(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "设备规则", RuleTypeThreshold)
	rule.DeviceIDs = []string{"device001"}
	rule.Enabled = true
	detector.AddRule(rule)

	point := &DataPoint{PointID: "point001", DeviceID: "device001", StationID: "station001", Value: 100.0, Timestamp: time.Now()}
	matching := detector.getMatchingRules(point)
	assert.Len(t, matching, 1)

	point.DeviceID = "device002"
	matching = detector.getMatchingRules(point)
	assert.Len(t, matching, 0)
}

func TestDetector_MatchingRules_StationFilter(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "站点规则", RuleTypeThreshold)
	rule.StationIDs = []string{"station001"}
	rule.Enabled = true
	detector.AddRule(rule)

	point := &DataPoint{PointID: "point001", DeviceID: "device001", StationID: "station001", Value: 100.0, Timestamp: time.Now()}
	matching := detector.getMatchingRules(point)
	assert.Len(t, matching, 1)

	point.StationID = "station002"
	matching = detector.getMatchingRules(point)
	assert.Len(t, matching, 0)
}

func TestDetector_MatchingRules_NoFilters(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "无过滤规则", RuleTypeThreshold)
	rule.Enabled = true
	detector.AddRule(rule)

	point := &DataPoint{PointID: "point001", DeviceID: "device001", StationID: "station001", Value: 100.0, Timestamp: time.Now()}
	matching := detector.getMatchingRules(point)
	assert.Len(t, matching, 1)
}

func TestDetector_Detect(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	detector.Start(ctx)

	rule := NewRule("rule001", "阈值规则", RuleTypeThreshold)
	rule.Operator = OpGreaterThan
	rule.Threshold = 100.0
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit
	rule.Title = "测试告警"
	rule.PointIDs = []string{"point001"}
	detector.AddRule(rule)

	point := &DataPoint{PointID: "point001", DeviceID: "device001", StationID: "station001", Value: 200.0, Timestamp: time.Now()}
	err := detector.Detect(ctx, point)
	assert.NoError(t, err)
}

func TestDetector_DetectBatch(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	detector.Start(ctx)

	points := []*DataPoint{
		{PointID: "point001", Value: 100.0, Timestamp: time.Now()},
		{PointID: "point002", Value: 200.0, Timestamp: time.Now()},
	}
	err := detector.DetectBatch(ctx, points)
	assert.NoError(t, err)
}

func TestDetector_AddHandler(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	detector.AddHandler(func(ctx context.Context, result *DetectionResult) error {
		return nil
	})
	assert.Len(t, detector.handlers, 1)
}

func TestDetector_StartStop(t *testing.T) {
	detector := NewDetector(DetectorConfig{WorkerCount: 2, BufferSize: 100})
	ctx, cancel := context.WithCancel(context.Background())
	detector.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()
	detector.Stop()
}

func TestDetector_SetStateMachine(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	detector.SetStateMachine(nil)
	assert.Nil(t, detector.stateMachine)
}

func TestDetector_SetDeduplicator(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	detector.SetDeduplicator(nil)
	assert.Nil(t, detector.deduplicator)
}

func TestDetector_SetAggregator(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	detector.SetAggregator(nil)
	assert.Nil(t, detector.aggregator)
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

func TestRealtimeDetector(t *testing.T) {
	rd := NewRealtimeDetector(DefaultDetectorConfig())
	assert.NotNil(t, rd)
	assert.NotNil(t, rd.GetDetector())

	rule := NewRule("rule001", "测试规则", RuleTypeThreshold)
	rule.Operator = OpGreaterThan
	rule.Threshold = 100.0
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit
	rule.Title = "测试"
	rule.PointIDs = []string{"point001"}
	rd.AddRule(rule)
	assert.Equal(t, 1, rd.GetDetector().GetRuleCount())

	rd.RemoveRule("rule001")
	assert.Equal(t, 0, rd.GetDetector().GetRuleCount())
}

func TestRealtimeDetector_StartStop(t *testing.T) {
	rd := NewRealtimeDetector(DetectorConfig{WorkerCount: 2, BufferSize: 100})
	ctx, cancel := context.WithCancel(context.Background())
	rd.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()
	rd.Stop()
}

func TestRealtimeDetector_Detect(t *testing.T) {
	rd := NewRealtimeDetector(DetectorConfig{WorkerCount: 2, BufferSize: 100})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rd.Start(ctx)

	rule := NewRule("rule001", "阈值规则", RuleTypeThreshold)
	rule.Operator = OpGreaterThan
	rule.Threshold = 100.0
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit
	rule.Title = "测试"
	rule.PointIDs = []string{"point001"}
	rd.AddRule(rule)

	point := &DataPoint{PointID: "point001", DeviceID: "device001", StationID: "station001", Value: 200.0, Timestamp: time.Now()}
	err := rd.Detect(ctx, point)
	require.NoError(t, err)
}

func TestRealtimeDetector_DetectBatch(t *testing.T) {
	rd := NewRealtimeDetector(DetectorConfig{WorkerCount: 2, BufferSize: 100})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rd.Start(ctx)

	points := []*DataPoint{
		{PointID: "point001", Value: 100.0, Timestamp: time.Now()},
	}
	err := rd.DetectBatch(ctx, points)
	require.NoError(t, err)
}

func TestRealtimeDetector_AddHandler(t *testing.T) {
	rd := NewRealtimeDetector(DefaultDetectorConfig())
	rd.AddHandler(func(ctx context.Context, result *DetectionResult) error {
		return nil
	})
	assert.Len(t, rd.GetDetector().handlers, 1)
}

func TestRealtimeDetector_GetStats(t *testing.T) {
	rd := NewRealtimeDetector(DefaultDetectorConfig())
	stats := rd.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalProcessed)
}

func TestDetector_DetectRate_ZeroDuration(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "变化率告警", RuleTypeRate)
	rule.RateThreshold = 10.0
	rule.PointIDs = []string{"point001"}

	now := time.Now()
	point1 := &DataPoint{PointID: "point001", Value: 100.0, Timestamp: now}
	detector.updateWindow(point1)
	point2 := &DataPoint{PointID: "point001", Value: 200.0, Timestamp: now}
	detector.updateWindow(point2)

	triggered, _ := detector.detectRate(rule, point2)
	assert.False(t, triggered)
}

func TestDetector_HandleResult_NotTriggered(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	result := &DetectionResult{Triggered: false}
	detector.handleResult(context.Background(), result)
	stats := detector.GetStats()
	assert.Equal(t, int64(0), stats.TotalTriggered)
}

func TestDetector_HandleResult_Triggered(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "测试", RuleTypeThreshold)
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit
	result := &DetectionResult{
		Triggered: true,
		Rule:      rule,
		Alarm:     entity.NewAlarm("p1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "test", "msg"),
	}
	detector.handleResult(context.Background(), result)
	stats := detector.GetStats()
	assert.Equal(t, int64(1), stats.TotalTriggered)
}

func TestDetector_HandleResult_WithHandler(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "测试", RuleTypeThreshold)
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit

	called := false
	detector.AddHandler(func(ctx context.Context, result *DetectionResult) error {
		called = true
		return nil
	})

	result := &DetectionResult{
		Triggered: true,
		Rule:      rule,
		Alarm:     entity.NewAlarm("p1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "test", "msg"),
	}
	detector.handleResult(context.Background(), result)
	assert.True(t, called)
}

func TestDetector_HandleResult_HandlerError(t *testing.T) {
	detector := NewDetector(DefaultDetectorConfig())
	rule := NewRule("rule001", "测试", RuleTypeThreshold)
	rule.Level = entity.AlarmLevelWarning
	rule.AlarmType = entity.AlarmTypeLimit

	detector.AddHandler(func(ctx context.Context, result *DetectionResult) error {
		return assert.AnError
	})

	result := &DetectionResult{
		Triggered: true,
		Rule:      rule,
		Alarm:     entity.NewAlarm("p1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "test", "msg"),
	}
	detector.handleResult(context.Background(), result)
	stats := detector.GetStats()
	assert.Equal(t, int64(1), stats.TotalTriggered)
}
