package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAlarmRepository 告警仓储Mock
type MockAlarmRepository struct {
	mock.Mock
}

func (m *MockAlarmRepository) Create(ctx context.Context, alarm *entity.Alarm) error {
	args := m.Called(ctx, alarm)
	return args.Error(0)
}

func (m *MockAlarmRepository) Update(ctx context.Context, alarm *entity.Alarm) error {
	args := m.Called(ctx, alarm)
	return args.Error(0)
}

func (m *MockAlarmRepository) GetByID(ctx context.Context, id string) (*entity.Alarm, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Alarm), args.Error(1)
}

func (m *MockAlarmRepository) GetActiveAlarms(ctx context.Context, stationID *string, level *entity.AlarmLevel) ([]*entity.Alarm, error) {
	args := m.Called(ctx, stationID, level)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Alarm), args.Error(1)
}

func (m *MockAlarmRepository) GetHistoryAlarms(ctx context.Context, stationID *string, startTime, endTime int64) ([]*entity.Alarm, error) {
	args := m.Called(ctx, stationID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Alarm), args.Error(1)
}

func (m *MockAlarmRepository) Acknowledge(ctx context.Context, id, by string) error {
	args := m.Called(ctx, id, by)
	return args.Error(0)
}

func (m *MockAlarmRepository) Clear(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlarmRepository) CountByLevel(ctx context.Context, stationID *string) (map[entity.AlarmLevel]int64, error) {
	args := m.Called(ctx, stationID)
	return args.Get(0).(map[entity.AlarmLevel]int64), args.Error(1)
}

func TestAlarmService_CreateAlarm(t *testing.T) {
	ctx := context.Background()

	mockAlarmRepo := new(MockAlarmRepository)
	service := NewAlarmService(mockAlarmRepo)

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
}

func TestAlarmService_GetAlarm(t *testing.T) {
	ctx := context.Background()

	mockAlarmRepo := new(MockAlarmRepository)
	service := NewAlarmService(mockAlarmRepo)

	expectedAlarm := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警", "测试告警")
	mockAlarmRepo.On("GetByID", ctx, "alarm001").Return(expectedAlarm, nil)

	alarm, err := service.GetAlarm(ctx, "alarm001")

	assert.NoError(t, err)
	assert.NotNil(t, alarm)
	assert.Equal(t, "point001", alarm.PointID)
	mockAlarmRepo.AssertExpectations(t)
}

func TestAlarmService_GetActiveAlarms(t *testing.T) {
	ctx := context.Background()

	mockAlarmRepo := new(MockAlarmRepository)
	service := NewAlarmService(mockAlarmRepo)

	expectedAlarms := []*entity.Alarm{
		entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警1", "测试告警1"),
		entity.NewAlarm("point002", "device002", "station001", entity.AlarmTypeDevice, entity.AlarmLevelCritical, "告警2", "测试告警2"),
	}
	stationID := "station001"
	level := entity.AlarmLevelWarning

	mockAlarmRepo.On("GetActiveAlarms", ctx, &stationID, &level).Return(expectedAlarms, nil)

	alarms, err := service.GetActiveAlarms(ctx, &stationID, &level)

	assert.NoError(t, err)
	assert.Len(t, alarms, 2)
	mockAlarmRepo.AssertExpectations(t)
}

func TestAlarmService_GetHistoryAlarms(t *testing.T) {
	ctx := context.Background()

	mockAlarmRepo := new(MockAlarmRepository)
	service := NewAlarmService(mockAlarmRepo)

	now := time.Now()
	startTime := now.Add(-24 * time.Hour).Unix()
	endTime := now.Unix()

	expectedAlarms := []*entity.Alarm{
		entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "历史告警1", "测试"),
	}
	stationID := "station001"

	mockAlarmRepo.On("GetHistoryAlarms", ctx, &stationID, startTime, endTime).Return(expectedAlarms, nil)

	alarms, err := service.GetHistoryAlarms(ctx, &stationID, startTime, endTime)

	assert.NoError(t, err)
	assert.Len(t, alarms, 1)
	mockAlarmRepo.AssertExpectations(t)
}

func TestAlarmService_AcknowledgeAlarm(t *testing.T) {
	ctx := context.Background()

	t.Run("成功确认告警", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmService(mockAlarmRepo)

		alarm := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警", "测试告警")
		mockAlarmRepo.On("GetByID", ctx, "alarm001").Return(alarm, nil)
		mockAlarmRepo.On("Acknowledge", ctx, "alarm001", "user001").Return(nil)

		err := service.AcknowledgeAlarm(ctx, "alarm001", "user001")

		assert.NoError(t, err)
		mockAlarmRepo.AssertExpectations(t)
	})

	t.Run("告警不存在", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmService(mockAlarmRepo)

		mockAlarmRepo.On("GetByID", ctx, "alarm001").Return(nil, errors.New("not found"))

		err := service.AcknowledgeAlarm(ctx, "alarm001", "user001")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockAlarmRepo.AssertExpectations(t)
	})

	t.Run("告警状态不正确", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmService(mockAlarmRepo)

		alarm := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警", "测试告警")
		alarm.Acknowledge("user001") // 已经确认过

		mockAlarmRepo.On("GetByID", ctx, "alarm001").Return(alarm, nil)

		err := service.AcknowledgeAlarm(ctx, "alarm001", "user002")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in active state")
		mockAlarmRepo.AssertExpectations(t)
	})
}

func TestAlarmService_ClearAlarm(t *testing.T) {
	ctx := context.Background()

	t.Run("成功清除活动告警", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmService(mockAlarmRepo)

		alarm := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警", "测试告警")
		mockAlarmRepo.On("GetByID", ctx, "alarm001").Return(alarm, nil)
		mockAlarmRepo.On("Clear", ctx, "alarm001").Return(nil)

		err := service.ClearAlarm(ctx, "alarm001")

		assert.NoError(t, err)
		mockAlarmRepo.AssertExpectations(t)
	})

	t.Run("成功清除已确认告警", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmService(mockAlarmRepo)

		alarm := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警", "测试告警")
		alarm.Acknowledge("user001")
		mockAlarmRepo.On("GetByID", ctx, "alarm001").Return(alarm, nil)
		mockAlarmRepo.On("Clear", ctx, "alarm001").Return(nil)

		err := service.ClearAlarm(ctx, "alarm001")

		assert.NoError(t, err)
		mockAlarmRepo.AssertExpectations(t)
	})

	t.Run("告警状态不正确", func(t *testing.T) {
		mockAlarmRepo := new(MockAlarmRepository)
		service := NewAlarmService(mockAlarmRepo)

		alarm := entity.NewAlarm("point001", "device001", "station001", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "告警", "测试告警")
		alarm.Clear() // 已经清除过

		mockAlarmRepo.On("GetByID", ctx, "alarm001").Return(alarm, nil)

		err := service.ClearAlarm(ctx, "alarm001")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be cleared")
		mockAlarmRepo.AssertExpectations(t)
	})
}

func TestAlarmService_CountAlarmsByLevel(t *testing.T) {
	ctx := context.Background()

	mockAlarmRepo := new(MockAlarmRepository)
	service := NewAlarmService(mockAlarmRepo)

	expectedCounts := map[entity.AlarmLevel]int64{
		entity.AlarmLevelInfo:     5,
		entity.AlarmLevelWarning:  10,
		entity.AlarmLevelMajor:    3,
		entity.AlarmLevelCritical: 1,
	}
	stationID := "station001"

	mockAlarmRepo.On("CountByLevel", ctx, &stationID).Return(expectedCounts, nil)

	counts, err := service.CountAlarmsByLevel(ctx, &stationID)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), counts[entity.AlarmLevelInfo])
	assert.Equal(t, int64(10), counts[entity.AlarmLevelWarning])
	assert.Equal(t, int64(3), counts[entity.AlarmLevelMajor])
	assert.Equal(t, int64(1), counts[entity.AlarmLevelCritical])
	mockAlarmRepo.AssertExpectations(t)
}
