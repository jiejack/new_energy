package service

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAlarmRuleRepository struct {
	mock.Mock
}

func (m *MockAlarmRuleRepository) Create(ctx context.Context, rule *entity.AlarmRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockAlarmRuleRepository) Update(ctx context.Context, rule *entity.AlarmRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockAlarmRuleRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlarmRuleRepository) GetByID(ctx context.Context, id string) (*entity.AlarmRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AlarmRule), args.Error(1)
}

func (m *MockAlarmRuleRepository) GetByName(ctx context.Context, name string) (*entity.AlarmRule, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AlarmRule), args.Error(1)
}

func (m *MockAlarmRuleRepository) List(ctx context.Context, query *repository.AlarmRuleQuery) ([]*entity.AlarmRule, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Get(1).(int64), args.Error(2)
}

func (m *MockAlarmRuleRepository) GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func (m *MockAlarmRuleRepository) GetRulesByPointID(ctx context.Context, pointID string) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx, pointID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func (m *MockAlarmRuleRepository) GetRulesByDeviceID(ctx context.Context, deviceID string) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func (m *MockAlarmRuleRepository) GetRulesByStationID(ctx context.Context, stationID string) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func TestAlarmRuleService_CreateRule(t *testing.T) {
	ctx := context.Background()

	t.Run("成功创建告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		req := &CreateAlarmRuleRequest{
			Name:        "温度过高告警",
			Description: "逆变器温度超过阈值告警",
			Type:        entity.AlarmRuleTypeLimit,
			Level:       entity.AlarmLevelWarning,
			Condition:   "value > threshold",
			Threshold:   85.0,
			Duration:    60,
			NotifyChannels: []string{"email", "sms"},
			NotifyUsers:    []string{"user001", "user002"},
		}

		mockRepo.On("GetByName", ctx, "温度过高告警").Return(nil, errors.New("not found"))
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

		rule, err := service.CreateRule(ctx, req, "admin")

		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, "温度过高告警", rule.Name)
		assert.Equal(t, entity.AlarmRuleTypeLimit, rule.Type)
		assert.Equal(t, entity.AlarmLevelWarning, rule.Level)
		assert.Len(t, rule.NotifyChannels, 2)
		assert.Len(t, rule.NotifyUsers, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("创建带关联对象的告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		pointID := "point001"
		deviceID := "device001"
		stationID := "station001"

		req := &CreateAlarmRuleRequest{
			Name:      "电压异常告警",
			Type:      entity.AlarmRuleTypeTrend,
			Level:     entity.AlarmLevelMajor,
			Condition: "value < threshold",
			PointID:   &pointID,
			DeviceID:  &deviceID,
			StationID: &stationID,
		}

		mockRepo.On("GetByName", ctx, "电压异常告警").Return(nil, errors.New("not found"))
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

		rule, err := service.CreateRule(ctx, req, "admin")

		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, &pointID, rule.PointID)
		assert.Equal(t, &deviceID, rule.DeviceID)
		assert.Equal(t, &stationID, rule.StationID)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleService_UpdateRule(t *testing.T) {
	ctx := context.Background()

	t.Run("成功更新告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		existingRule := entity.NewAlarmRule("原始规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
		existingRule.Threshold = 80.0

		name := "更新后的规则"
		desc := "更新后的描述"
		level := entity.AlarmLevelMajor
		cond := "value > threshold"
		threshold := 90.0
		duration := 120

		req := &UpdateAlarmRuleRequest{
			Name:        &name,
			Description: &desc,
			Level:       &level,
			Condition:   &cond,
			Threshold:   &threshold,
			Duration:    &duration,
		}

		mockRepo.On("GetByID", ctx, existingRule.ID).Return(existingRule, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

		rule, err := service.UpdateRule(ctx, existingRule.ID, req, "admin")

		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, "更新后的规则", rule.Name)
		assert.Equal(t, entity.AlarmLevelMajor, rule.Level)
		assert.Equal(t, 90.0, rule.Threshold)
		assert.Equal(t, 120, rule.Duration)
		mockRepo.AssertExpectations(t)
	})

	t.Run("告警规则不存在", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		name := "更新后的规则"
		level := entity.AlarmLevelMajor
		req := &UpdateAlarmRuleRequest{
			Name:  &name,
			Level: &level,
		}

		mockRepo.On("GetByID", ctx, "non-existent-id").Return(nil, errors.New("not found"))

		rule, err := service.UpdateRule(ctx, "non-existent-id", req, "admin")

		assert.Error(t, err)
		assert.Nil(t, rule)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleService_DeleteRule(t *testing.T) {
	ctx := context.Background()

	t.Run("成功删除告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		mockRepo.On("Delete", ctx, "rule-001").Return(nil)

		err := service.DeleteRule(ctx, "rule-001")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleService_GetRule(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		expectedRule := entity.NewAlarmRule("测试规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
		mockRepo.On("GetByID", ctx, expectedRule.ID).Return(expectedRule, nil)

		rule, err := service.GetRule(ctx, expectedRule.ID)

		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, expectedRule.ID, rule.ID)
		assert.Equal(t, "测试规则", rule.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("告警规则不存在", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		mockRepo.On("GetByID", ctx, "non-existent-id").Return(nil, errors.New("not found"))

		rule, err := service.GetRule(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, rule)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleService_ListRules(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取告警规则列表", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		expectedRules := []*entity.AlarmRule{
			entity.NewAlarmRule("规则1", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold"),
			entity.NewAlarmRule("规则2", entity.AlarmRuleTypeTrend, entity.AlarmLevelMajor, "value < threshold"),
		}

		query := &repository.AlarmRuleQuery{
			Page:     1,
			PageSize: 20,
		}

		mockRepo.On("List", ctx, query).Return(expectedRules, int64(2), nil)

		rules, total, err := service.ListRules(ctx, query)

		assert.NoError(t, err)
		assert.Len(t, rules, 2)
		assert.Equal(t, int64(2), total)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleService_GetEnabledRules(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockAlarmRuleRepository)
	service := NewAlarmRuleService(mockRepo)

	expectedRules := []*entity.AlarmRule{
		entity.NewAlarmRule("启用规则1", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold"),
		entity.NewAlarmRule("启用规则2", entity.AlarmRuleTypeTrend, entity.AlarmLevelMajor, "value < threshold"),
	}

	mockRepo.On("GetEnabledRules", ctx).Return(expectedRules, nil)

	rules, err := service.GetEnabledRules(ctx)

	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	mockRepo.AssertExpectations(t)
}
