package persistence

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDB 模拟数据库
type MockDB struct {
	mock.Mock
}

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

func TestAlarmRuleRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("成功创建告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		rule := entity.NewAlarmRule("温度过高告警", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)
		rule.Description = "逆变器温度超过阈值"
		rule.Threshold = 85.0
		rule.Duration = 60

		mockRepo.On("Create", ctx, rule).Return(nil)

		err := mockRepo.Create(ctx, rule)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("创建失败", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		rule := entity.NewAlarmRule("测试规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)

		mockRepo.On("Create", ctx, rule).Return(errors.New("database error"))

		err := mockRepo.Create(ctx, rule)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("成功更新告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		rule := entity.NewAlarmRule("测试规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)
		rule.Threshold = 90.0

		mockRepo.On("Update", ctx, rule).Return(nil)

		err := mockRepo.Update(ctx, rule)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("更新失败", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		rule := entity.NewAlarmRule("测试规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)

		mockRepo.On("Update", ctx, rule).Return(errors.New("database error"))

		err := mockRepo.Update(ctx, rule)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("成功删除告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		ruleID := "rule-001"

		mockRepo.On("Delete", ctx, ruleID).Return(nil)

		err := mockRepo.Delete(ctx, ruleID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("删除不存在的规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		ruleID := "nonexistent"

		mockRepo.On("Delete", ctx, ruleID).Return(gorm.ErrRecordNotFound)

		err := mockRepo.Delete(ctx, ruleID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleRepository_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		expectedRule := entity.NewAlarmRule("测试规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)

		mockRepo.On("GetByID", ctx, expectedRule.ID).Return(expectedRule, nil)

		rule, err := mockRepo.GetByID(ctx, expectedRule.ID)

		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, expectedRule.ID, rule.ID)
		assert.Equal(t, "测试规则", rule.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("告警规则不存在", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		mockRepo.On("GetByID", ctx, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		rule, err := mockRepo.GetByID(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, rule)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleRepository_List(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取告警规则列表", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

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

		rules, total, err := mockRepo.List(ctx, query)

		assert.NoError(t, err)
		assert.Len(t, rules, 2)
		assert.Equal(t, int64(2), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("带过滤条件查询", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

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

		rules, total, err := mockRepo.List(ctx, query)

		assert.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, int64(1), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("空列表", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		query := &repository.AlarmRuleQuery{
			Page:     1,
			PageSize: 20,
		}

		mockRepo.On("List", ctx, query).Return([]*entity.AlarmRule{}, int64(0), nil)

		rules, total, err := mockRepo.List(ctx, query)

		assert.NoError(t, err)
		assert.NotNil(t, rules)
		assert.Len(t, rules, 0)
		assert.Equal(t, int64(0), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("查询失败", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		query := &repository.AlarmRuleQuery{
			Page:     1,
			PageSize: 20,
		}

		mockRepo.On("List", ctx, query).Return(nil, int64(0), errors.New("database error"))

		rules, total, err := mockRepo.List(ctx, query)

		assert.Error(t, err)
		assert.Nil(t, rules)
		assert.Equal(t, int64(0), total)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleRepository_GetEnabledRules(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取启用的告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		expectedRules := []*entity.AlarmRule{
			entity.NewAlarmRule("启用规则1", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
			entity.NewAlarmRule("启用规则2", entity.AlarmRuleTypeTrend, entity.AlarmLevelMajor),
		}

		mockRepo.On("GetEnabledRules", ctx).Return(expectedRules, nil)

		rules, err := mockRepo.GetEnabledRules(ctx)

		assert.NoError(t, err)
		assert.Len(t, rules, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("没有启用的规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		mockRepo.On("GetEnabledRules", ctx).Return([]*entity.AlarmRule{}, nil)

		rules, err := mockRepo.GetEnabledRules(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, rules)
		assert.Len(t, rules, 0)
		mockRepo.AssertExpectations(t)
	})

	t.Run("查询失败", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		mockRepo.On("GetEnabledRules", ctx).Return(nil, errors.New("database error"))

		rules, err := mockRepo.GetEnabledRules(ctx)

		assert.Error(t, err)
		assert.Nil(t, rules)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleRepository_GetRulesByPointID(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取采集点告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		pointID := "point-001"
		expectedRules := []*entity.AlarmRule{
			entity.NewAlarmRule("采集点规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
		}

		mockRepo.On("GetRulesByPointID", ctx, pointID).Return(expectedRules, nil)

		rules, err := mockRepo.GetRulesByPointID(ctx, pointID)

		assert.NoError(t, err)
		assert.Len(t, rules, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("采集点没有告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		pointID := "point-002"

		mockRepo.On("GetRulesByPointID", ctx, pointID).Return([]*entity.AlarmRule{}, nil)

		rules, err := mockRepo.GetRulesByPointID(ctx, pointID)

		assert.NoError(t, err)
		assert.NotNil(t, rules)
		assert.Len(t, rules, 0)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleRepository_GetRulesByDeviceID(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取设备告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		deviceID := "device-001"
		expectedRules := []*entity.AlarmRule{
			entity.NewAlarmRule("设备规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
		}

		mockRepo.On("GetRulesByDeviceID", ctx, deviceID).Return(expectedRules, nil)

		rules, err := mockRepo.GetRulesByDeviceID(ctx, deviceID)

		assert.NoError(t, err)
		assert.Len(t, rules, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("设备没有告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		deviceID := "device-002"

		mockRepo.On("GetRulesByDeviceID", ctx, deviceID).Return([]*entity.AlarmRule{}, nil)

		rules, err := mockRepo.GetRulesByDeviceID(ctx, deviceID)

		assert.NoError(t, err)
		assert.NotNil(t, rules)
		assert.Len(t, rules, 0)
		mockRepo.AssertExpectations(t)
	})
}

func TestAlarmRuleRepository_GetRulesByStationID(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取厂站告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		stationID := "station-001"
		expectedRules := []*entity.AlarmRule{
			entity.NewAlarmRule("厂站规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
		}

		mockRepo.On("GetRulesByStationID", ctx, stationID).Return(expectedRules, nil)

		rules, err := mockRepo.GetRulesByStationID(ctx, stationID)

		assert.NoError(t, err)
		assert.Len(t, rules, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("厂站没有告警规则", func(t *testing.T) {
		mockRepo := new(MockAlarmRuleRepository)

		stationID := "station-002"

		mockRepo.On("GetRulesByStationID", ctx, stationID).Return([]*entity.AlarmRule{}, nil)

		rules, err := mockRepo.GetRulesByStationID(ctx, stationID)

		assert.NoError(t, err)
		assert.NotNil(t, rules)
		assert.Len(t, rules, 0)
		mockRepo.AssertExpectations(t)
	})
}
