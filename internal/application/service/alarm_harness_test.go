package service

import (
	"context"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewAlarmHarness(t *testing.T) {
	harness := NewAlarmHarness()
	assert.NotNil(t, harness)
	assert.NotNil(t, harness.harness)
}

func TestAlarmHarness_ValidateCreateAlarm(t *testing.T) {
	ctx := context.Background()
	harness := NewAlarmHarness()

	t.Run("有效的创建告警请求", func(t *testing.T) {
		req := &CreateAlarmRequest{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Type:      entity.AlarmTypeLimit,
			Level:     entity.AlarmLevelWarning,
			Title:     "电压高限告警",
			Message:   "电压超过上限阈值",
			Value:     450.0,
			Threshold: 400.0,
		}

		err := harness.ValidateCreateAlarm(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("无效的告警级别", func(t *testing.T) {
		req := &CreateAlarmRequest{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Type:      entity.AlarmTypeLimit,
			Level:     entity.AlarmLevel(99), // 无效级别
			Title:     "电压高限告警",
		}

		err := harness.ValidateCreateAlarm(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid alarm level")
	})

	t.Run("无效的告警类型", func(t *testing.T) {
		req := &CreateAlarmRequest{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Type:      entity.AlarmType("invalid_type"), // 无效类型
			Level:     entity.AlarmLevelWarning,
			Title:     "电压高限告警",
		}

		err := harness.ValidateCreateAlarm(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid alarm type")
	})

	t.Run("标题为空", func(t *testing.T) {
		req := &CreateAlarmRequest{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Type:      entity.AlarmTypeLimit,
			Level:     entity.AlarmLevelWarning,
			Title:     "", // 空标题
		}

		err := harness.ValidateCreateAlarm(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title cannot be empty")
	})

	t.Run("限值告警值等于阈值", func(t *testing.T) {
		req := &CreateAlarmRequest{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Type:      entity.AlarmTypeLimit,
			Level:     entity.AlarmLevelWarning,
			Title:     "电压高限告警",
			Value:     400.0,
			Threshold: 400.0, // 值等于阈值
		}

		err := harness.ValidateCreateAlarm(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value should deviate from threshold")
	})

	t.Run("限值告警值超过阈值", func(t *testing.T) {
		req := &CreateAlarmRequest{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Type:      entity.AlarmTypeLimit,
			Level:     entity.AlarmLevelWarning,
			Title:     "电压高限告警",
			Value:     450.0,
			Threshold: 400.0,
		}

		err := harness.ValidateCreateAlarm(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("信息级别告警", func(t *testing.T) {
		req := &CreateAlarmRequest{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Type:      entity.AlarmTypeStatus,
			Level:     entity.AlarmLevelInfo,
			Title:     "设备状态变化",
			Message:   "设备从停机转为运行",
		}

		err := harness.ValidateCreateAlarm(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("严重级别告警", func(t *testing.T) {
		req := &CreateAlarmRequest{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Type:      entity.AlarmTypeDevice,
			Level:     entity.AlarmLevelCritical,
			Title:     "设备故障",
			Message:   "逆变器严重故障",
		}

		err := harness.ValidateCreateAlarm(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("通信告警", func(t *testing.T) {
		req := &CreateAlarmRequest{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Type:      entity.AlarmTypeComm,
			Level:     entity.AlarmLevelMajor,
			Title:     "通信中断",
			Message:   "设备通信中断超过5分钟",
		}

		err := harness.ValidateCreateAlarm(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("系统告警", func(t *testing.T) {
		req := &CreateAlarmRequest{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Type:      entity.AlarmTypeSystem,
			Level:     entity.AlarmLevelWarning,
			Title:     "系统资源不足",
			Message:   "CPU使用率超过90%",
		}

		err := harness.ValidateCreateAlarm(ctx, req)
		assert.NoError(t, err)
	})
}

func TestAlarmHarness_ValidateAcknowledgeAlarm(t *testing.T) {
	ctx := context.Background()
	harness := NewAlarmHarness()

	t.Run("有效的确认请求", func(t *testing.T) {
		err := harness.ValidateAcknowledgeAlarm(ctx, "alarm001", "user001")
		assert.NoError(t, err)
	})

	t.Run("告警ID为空", func(t *testing.T) {
		err := harness.ValidateAcknowledgeAlarm(ctx, "", "user001")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "alarm ID cannot be empty")
	})

	t.Run("操作者为空", func(t *testing.T) {
		err := harness.ValidateAcknowledgeAlarm(ctx, "alarm001", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operator cannot be empty")
	})
}

func TestAlarmHarness_ValidateClearAlarm(t *testing.T) {
	ctx := context.Background()
	harness := NewAlarmHarness()

	t.Run("有效的清除请求", func(t *testing.T) {
		err := harness.ValidateClearAlarm(ctx, "alarm001")
		assert.NoError(t, err)
	})

	t.Run("告警ID为空", func(t *testing.T) {
		err := harness.ValidateClearAlarm(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "alarm ID cannot be empty")
	})
}

func TestAlarmHarness_ValidateAlarmQuery(t *testing.T) {
	ctx := context.Background()
	harness := NewAlarmHarness()

	t.Run("有效的查询请求", func(t *testing.T) {
		startTime := int64(1709414400000)
		endTime := int64(1709500800000)
		err := harness.ValidateAlarmQuery(ctx, nil, startTime, endTime)
		assert.NoError(t, err)
	})

	t.Run("开始时间大于结束时间", func(t *testing.T) {
		startTime := int64(1709500800000)
		endTime := int64(1709414400000)
		err := harness.ValidateAlarmQuery(ctx, nil, startTime, endTime)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start time cannot be greater than end time")
	})

	t.Run("时间为零值", func(t *testing.T) {
		err := harness.ValidateAlarmQuery(ctx, nil, 0, 0)
		assert.NoError(t, err)
	})
}

func TestAlarmHarness_VerifyAlarmOutput(t *testing.T) {
	ctx := context.Background()
	harness := NewAlarmHarness()

	t.Run("输出验证成功", func(t *testing.T) {
		expected := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警", "测试告警")
		actual := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警", "测试告警")

		match, err := harness.VerifyAlarmOutput(ctx, expected, actual)
		assert.NoError(t, err)
		assert.True(t, match)
	})

	t.Run("输出验证失败", func(t *testing.T) {
		expected := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警1", "测试告警")
		actual := entity.NewAlarm("point002", "device002", "station002", entity.AlarmTypeDevice, entity.AlarmLevelCritical, "告警2", "测试告警2")

		match, err := harness.VerifyAlarmOutput(ctx, expected, actual)
		assert.NoError(t, err)
		assert.False(t, match)
	})
}

func TestAlarmHarness_CreateAlarmSnapshot(t *testing.T) {
	ctx := context.Background()
	harness := NewAlarmHarness()

	alarm := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警", "测试告警")

	snapshot, err := harness.CreateAlarmSnapshot(ctx, alarm)
	assert.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.True(t, len(snapshot) > 0)
}

func TestAlarmHarness_GetHarness(t *testing.T) {
	harness := NewAlarmHarness()

	h := harness.GetHarness()
	assert.NotNil(t, h)
}

func TestIsValidAlarmLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    entity.AlarmLevel
		expected bool
	}{
		{"信息级别", entity.AlarmLevelInfo, true},
		{"警告级别", entity.AlarmLevelWarning, true},
		{"重要级别", entity.AlarmLevelMajor, true},
		{"严重级别", entity.AlarmLevelCritical, true},
		{"无效级别", entity.AlarmLevel(0), false},
		{"无效级别99", entity.AlarmLevel(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidAlarmLevel(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidAlarmType(t *testing.T) {
	tests := []struct {
		name      string
		alarmType entity.AlarmType
		expected  bool
	}{
		{"限值告警", entity.AlarmTypeLimit, true},
		{"状态告警", entity.AlarmTypeStatus, true},
		{"通信告警", entity.AlarmTypeComm, true},
		{"系统告警", entity.AlarmTypeSystem, true},
		{"设备告警", entity.AlarmTypeDevice, true},
		{"无效类型", entity.AlarmType("invalid"), false},
		{"空类型", entity.AlarmType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidAlarmType(tt.alarmType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
