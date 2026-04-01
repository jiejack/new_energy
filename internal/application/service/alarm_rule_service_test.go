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

// MockAlarmRuleRepository 告警规则仓储Mock
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

func TestAlarmRuleService_CreateAlarmRule(t *testing.T) {
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
			CreatedBy:      "admin",
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

		rule, err := service.CreateAlarmRule(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, "温度过高告警", rule.Name)
		assert.Equal(t, entity.AlarmRuleTypeLimit, rule.Type)
		assert.Equal(t, entity.AlarmLevelWarning, rule.Level)
		assert.Equal(t, entity.AlarmRuleStatusEnabled, rule.Status)
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
			Name:        "电压异常告警",
			Type:        entity.AlarmRuleTypeTrend,
			Level:       entity.AlarmLevelMajor,
			PointID:     &pointID,
			DeviceID:    &deviceID,
			StationID:   &stationID,
			CreatedBy:   "admin",
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

		rule, err := service.CreateAlarmRule(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, &pointID, rule.PointID)
		assert.Equal(t, &deviceID, rule.DeviceID)
		assert.Equal(t, &stationID, rule.StationID)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleService_UpdateAlarmRule(t *testing.T) {
	ctx := context.Background()

	t.Run("成功更新告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		existingRule := entity.NewAlarmRule("原始规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)
		existingRule.Threshold = 80.0

		req := &UpdateAlarmRuleRequest{
			Name:        "更新后的规则",
			Description: "更新后的描述",
			Level:       entity.AlarmLevelMajor,
			Condition:   "value > threshold",
			Threshold:   90.0,
			Duration:    120,
			UpdatedBy:   "admin",
		}

		mockRepo.On("GetByID", ctx, existingRule.ID).Return(existingRule, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

		rule, err := service.UpdateAlarmRule(ctx, existingRule.ID, req)

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

		req := &UpdateAlarmRuleRequest{
			Name:      "更新后的规则",
			Level:     entity.AlarmLevelMajor,
			UpdatedBy: "admin",
		}

		mockRepo.On("GetByID", ctx, "non-existent-id").Return(nil, errors.New("not found"))

		rule, err := service.UpdateAlarmRule(ctx, "non-existent-id", req)

		assert.Error(t, err)
		assert.Nil(t, rule)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleService_DeleteAlarmRule(t *testing.T) {
	ctx := context.Background()

	t.Run("成功删除告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		existingRule := entity.NewAlarmRule("测试规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)

		mockRepo.On("GetByID", ctx, existingRule.ID).Return(existingRule, nil)
		mockRepo.On("Delete", ctx, existingRule.ID).Return(nil)

		err := service.DeleteAlarmRule(ctx, existingRule.ID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("删除不存在的告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		mockRepo.On("GetByID", ctx, "non-existent-id").Return(nil, errors.New("not found"))

		err := service.DeleteAlarmRule(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleService_GetAlarmRule(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		expectedRule := entity.NewAlarmRule("测试规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)
		mockRepo.On("GetByID", ctx, expectedRule.ID).Return(expectedRule, nil)

		rule, err := service.GetAlarmRule(ctx, expectedRule.ID)

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

		rule, err := service.GetAlarmRule(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, rule)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleService_ListAlarmRules(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取告警规则列表", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		expectedRules := []*entity.AlarmRule{
			entity.NewAlarmRule("规则1", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
			entity.NewAlarmRule("规则2", entity.AlarmRuleTypeTrend, entity.AlarmLevelMajor),
		}

		query := &repository.AlarmRuleQuery{
			Page:     1,
			PageSize: 20,
			OrderBy:  "created_at",
			Order:    "desc",
		}

		mockRepo.On("List", ctx, query).Return(expectedRules, int64(2), nil)

		rules, total, err := service.ListAlarmRules(ctx, query)

		assert.NoError(t, err)
		assert.Len(t, rules, 2)
		assert.Equal(t, int64(2), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("使用过滤条件查询", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		ruleType := entity.AlarmRuleTypeLimit
		level := entity.AlarmLevelWarning

		expectedRules := []*entity.AlarmRule{
			entity.NewAlarmRule("限值告警", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
		}

		query := &repository.AlarmRuleQuery{
			Page:     1,
			PageSize: 20,
			OrderBy:  "created_at",
			Order:    "desc",
			Type:     &ruleType,
			Level:    &level,
		}

		mockRepo.On("List", ctx, query).Return(expectedRules, int64(1), nil)

		rules, total, err := service.ListAlarmRules(ctx, query)

		assert.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, int64(1), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("使用默认分页参数", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		query := &repository.AlarmRuleQuery{} // 空查询

		expectedRules := []*entity.AlarmRule{}

		// 期望使用默认值
		expectedQuery := &repository.AlarmRuleQuery{
			Page:     1,
			PageSize: 20,
			OrderBy:  "created_at",
			Order:    "desc",
		}

		mockRepo.On("List", ctx, expectedQuery).Return(expectedRules, int64(0), nil)

		rules, total, err := service.ListAlarmRules(ctx, query)

		assert.NoError(t, err)
		assert.NotNil(t, rules)
		assert.Equal(t, int64(0), total)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleService_EnableAlarmRule(t *testing.T) {
	ctx := context.Background()

	t.Run("成功启用告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		rule := entity.NewAlarmRule("测试规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)
		rule.Disable()

		mockRepo.On("GetByID", ctx, rule.ID).Return(rule, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

		err := service.EnableAlarmRule(ctx, rule.ID, "admin")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("启用不存在的告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		mockRepo.On("GetByID", ctx, "non-existent-id").Return(nil, errors.New("not found"))

		err := service.EnableAlarmRule(ctx, "non-existent-id", "admin")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleService_DisableAlarmRule(t *testing.T) {
	ctx := context.Background()

	t.Run("成功禁用告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		rule := entity.NewAlarmRule("测试规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)

		mockRepo.On("GetByID", ctx, rule.ID).Return(rule, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

		err := service.DisableAlarmRule(ctx, rule.ID, "admin")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("禁用不存在的告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)
		service := NewAlarmRuleService(mockRepo)

		mockRepo.On("GetByID", ctx, "non-existent-id").Return(nil, errors.New("not found"))

		err := service.DisableAlarmRule(ctx, "non-existent-id", "admin")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleService_GetEnabledRules(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockAlarmRuleRepository)
	service := NewAlarmRuleService(mockRepo)

	expectedRules := []*entity.AlarmRule{
		entity.NewAlarmRule("启用规则1", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
		entity.NewAlarmRule("启用规则2", entity.AlarmRuleTypeTrend, entity.AlarmLevelMajor),
	}

	mockRepo.On("GetEnabledRules", ctx).Return(expectedRules, nil)

	rules, err := service.GetEnabledRules(ctx)

	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	mockRepo.AssertExpectations(t)
}

func TestAlarmRuleService_GetRulesByPointID(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockAlarmRuleRepository)
	service := NewAlarmRuleService(mockRepo)

	pointID := "point001"
	expectedRules := []*entity.AlarmRule{
		entity.NewAlarmRule("采集点规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
	}

	mockRepo.On("GetRulesByPointID", ctx, pointID).Return(expectedRules, nil)

	rules, err := service.GetRulesByPointID(ctx, pointID)

	assert.NoError(t, err)
	assert.Len(t, rules, 1)
	mockRepo.AssertExpectations(t)
}

func TestAlarmRuleService_GetRulesByDeviceID(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockAlarmRuleRepository)
	service := NewAlarmRuleService(mockRepo)

	deviceID := "device001"
	expectedRules := []*entity.AlarmRule{
		entity.NewAlarmRule("设备规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
	}

	mockRepo.On("GetRulesByDeviceID", ctx, deviceID).Return(expectedRules, nil)

	rules, err := service.GetRulesByDeviceID(ctx, deviceID)

	assert.NoError(t, err)
	assert.Len(t, rules, 1)
	mockRepo.AssertExpectations(t)
}

func TestAlarmRuleService_GetRulesByStationID(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockAlarmRuleRepository)
	service := NewAlarmRuleService(mockRepo)

	stationID := "station001"
	expectedRules := []*entity.AlarmRule{
		entity.NewAlarmRule("厂站规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
	}

	mockRepo.On("GetRulesByStationID", ctx, stationID).Return(expectedRules, nil)

	rules, err := service.GetRulesByStationID(ctx, stationID)

	assert.NoError(t, err)
	assert.Len(t, rules, 1)
	mockRepo.AssertExpectations(t)
}
