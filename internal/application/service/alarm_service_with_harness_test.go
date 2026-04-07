package service

import (
	"context"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAlarmServiceWithHarness_CreateAlarm(t *testing.T) {
	ctx := context.Background()

	t.Run("成功创建告警（带验证）", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmServiceWithHarness(mockAlarmRepo)

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

		mockAlarmRepo.On("Create", ctx, mock.AnythingOfType("*entity.Alarm")).Return(nil)

		alarm, err := service.CreateAlarm(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, alarm)
		assert.Equal(t, "point001", alarm.PointID)
		assert.Equal(t, entity.AlarmLevelWarning, alarm.Level)
		assert.Equal(t, entity.AlarmStatusActive, alarm.Status)
		mockAlarmRepo.AssertExpectations(t)
	})

	t.Run("验证失败 - 无效告警级别", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmServiceWithHarness(mockAlarmRepo)

		req := &CreateAlarmRequest{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Type:      entity.AlarmTypeLimit,
			Level:     entity.AlarmLevel(99), // 无效级别
			Title:     "电压高限告警",
		}

		alarm, err := service.CreateAlarm(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, alarm)
		assert.Contains(t, err.Error(), "validation failed")
		mockAlarmRepo.AssertNotCalled(t, "Create")
	})

	t.Run("验证失败 - 标题为空", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmServiceWithHarness(mockAlarmRepo)

		req := &CreateAlarmRequest{
			PointID:   "point001",
			DeviceID:  "device001",
			StationID: "station001",
			Type:      entity.AlarmTypeLimit,
			Level:     entity.AlarmLevelWarning,
			Title:     "", // 空标题
		}

		alarm, err := service.CreateAlarm(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, alarm)
		assert.Contains(t, err.Error(), "validation failed")
		mockAlarmRepo.AssertNotCalled(t, "Create")
	})
}

func TestAlarmServiceWithHarness_AcknowledgeAlarm(t *testing.T) {
	ctx := context.Background()

	t.Run("成功确认告警（带验证）", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmServiceWithHarness(mockAlarmRepo)

		alarm := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警", "测试告警")
		mockAlarmRepo.On("GetByID", ctx, "alarm001").Return(alarm, nil)
		mockAlarmRepo.On("Acknowledge", ctx, "alarm001", "user001").Return(nil)

		err := service.AcknowledgeAlarm(ctx, "alarm001", "user001")

		assert.NoError(t, err)
		mockAlarmRepo.AssertExpectations(t)
	})

	t.Run("验证失败 - 告警ID为空", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmServiceWithHarness(mockAlarmRepo)

		err := service.AcknowledgeAlarm(ctx, "", "user001")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "alarm ID cannot be empty")
		mockAlarmRepo.AssertNotCalled(t, "GetByID")
		mockAlarmRepo.AssertNotCalled(t, "Acknowledge")
	})

	t.Run("验证失败 - 操作者为空", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmServiceWithHarness(mockAlarmRepo)

		err := service.AcknowledgeAlarm(ctx, "alarm001", "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operator cannot be empty")
		mockAlarmRepo.AssertNotCalled(t, "GetByID")
		mockAlarmRepo.AssertNotCalled(t, "Acknowledge")
	})
}

func TestAlarmServiceWithHarness_ClearAlarm(t *testing.T) {
	ctx := context.Background()

	t.Run("成功清除告警（带验证）", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmServiceWithHarness(mockAlarmRepo)

		alarm := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警", "测试告警")
		mockAlarmRepo.On("GetByID", ctx, "alarm001").Return(alarm, nil)
		mockAlarmRepo.On("Clear", ctx, "alarm001").Return(nil)

		err := service.ClearAlarm(ctx, "alarm001")

		assert.NoError(t, err)
		mockAlarmRepo.AssertExpectations(t)
	})

	t.Run("验证失败 - 告警ID为空", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmServiceWithHarness(mockAlarmRepo)

		err := service.ClearAlarm(ctx, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "alarm ID cannot be empty")
		mockAlarmRepo.AssertNotCalled(t, "GetByID")
		mockAlarmRepo.AssertNotCalled(t, "Clear")
	})
}

func TestAlarmServiceWithHarness_GetHistoryAlarms(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取历史告警（带验证）", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmServiceWithHarness(mockAlarmRepo)

		startTime := int64(1709414400000)
		endTime := int64(1709500800000)
		stationID := "station001"

		expectedAlarms := []*entity.Alarm{
			entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "历史告警1", "测试"),
		}

		mockAlarmRepo.On("GetHistoryAlarms", ctx, &stationID, startTime, endTime).Return(expectedAlarms, nil)

		alarms, err := service.GetHistoryAlarms(ctx, &stationID, startTime, endTime)

		assert.NoError(t, err)
		assert.Len(t, alarms, 1)
		mockAlarmRepo.AssertExpectations(t)
	})

	t.Run("验证失败 - 时间范围无效", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmServiceWithHarness(mockAlarmRepo)

		startTime := int64(1709500800000)
		endTime := int64(1709414400000)
		stationID := "station001"

		alarms, err := service.GetHistoryAlarms(ctx, &stationID, startTime, endTime)

		assert.Error(t, err)
		assert.Nil(t, alarms)
		assert.Contains(t, err.Error(), "start time cannot be greater than end time")
		mockAlarmRepo.AssertNotCalled(t, "GetHistoryAlarms")
	})
}

func TestAlarmServiceWithHarness_VerifyAlarmState(t *testing.T) {
	ctx := context.Background()

	t.Run("验证告警状态成功", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmServiceWithHarness(mockAlarmRepo)

		alarm := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警", "测试告警")
		mockAlarmRepo.On("GetByID", ctx, "alarm001").Return(alarm, nil)

		match, err := service.VerifyAlarmState(ctx, "alarm001", entity.AlarmStatusActive)

		assert.NoError(t, err)
		assert.True(t, match)
		mockAlarmRepo.AssertExpectations(t)
	})

	t.Run("验证告警状态失败", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmServiceWithHarness(mockAlarmRepo)

		alarm := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警", "测试告警")
		alarm.Acknowledge("user001")
		mockAlarmRepo.On("GetByID", ctx, "alarm001").Return(alarm, nil)

		match, err := service.VerifyAlarmState(ctx, "alarm001", entity.AlarmStatusActive)

		assert.NoError(t, err)
		assert.False(t, match)
		mockAlarmRepo.AssertExpectations(t)
	})
}

func TestAlarmServiceWithHarness_GetHarness(t *testing.T) {
	mockAlarmRepo := new(MockAlarmRepository)
	service := NewAlarmServiceWithHarness(mockAlarmRepo)

	h := service.GetHarness()
	assert.NotNil(t, h)
}
