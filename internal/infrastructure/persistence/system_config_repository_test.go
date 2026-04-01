package persistence

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockSystemConfigRepository 系统配置仓储Mock
type MockSystemConfigRepository struct {
	mock.Mock
}

func (m *MockSystemConfigRepository) Create(ctx context.Context, config *entity.SystemConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockSystemConfigRepository) Update(ctx context.Context, config *entity.SystemConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockSystemConfigRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSystemConfigRepository) GetByID(ctx context.Context, id string) (*entity.SystemConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SystemConfig), args.Error(1)
}

func (m *MockSystemConfigRepository) GetByKey(ctx context.Context, category, key string) (*entity.SystemConfig, error) {
	args := m.Called(ctx, category, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SystemConfig), args.Error(1)
}

func (m *MockSystemConfigRepository) GetByCategory(ctx context.Context, category string) ([]*entity.SystemConfig, error) {
	args := m.Called(ctx, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.SystemConfig), args.Error(1)
}

func (m *MockSystemConfigRepository) GetAll(ctx context.Context) ([]*entity.SystemConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.SystemConfig), args.Error(1)
}

func (m *MockSystemConfigRepository) List(ctx context.Context, filter *entity.SystemConfigFilter) ([]*entity.SystemConfig, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.SystemConfig), args.Get(1).(int64), args.Error(2)
}

func (m *MockSystemConfigRepository) BatchUpdate(ctx context.Context, configs []*entity.SystemConfig) error {
	args := m.Called(ctx, configs)
	return args.Error(0)
}

func (m *MockSystemConfigRepository) ExistsByKey(ctx context.Context, category, key string) (bool, error) {
	args := m.Called(ctx, category, key)
	return args.Bool(0), args.Error(1)
}

func TestSystemConfigRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("成功创建配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		config := entity.NewSystemConfig("basic", "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称")

		mockRepo.On("Create", ctx, config).Return(nil)

		err := mockRepo.Create(ctx, config)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("创建配置失败", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		config := entity.NewSystemConfig("basic", "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称")

		mockRepo.On("Create", ctx, config).Return(errors.New("database error"))

		err := mockRepo.Create(ctx, config)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestSystemConfigRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("成功更新配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		config := entity.NewSystemConfig("basic", "system_name", "新能源监控系统V2", entity.SystemConfigValueTypeString, "系统名称")

		mockRepo.On("Update", ctx, config).Return(nil)

		err := mockRepo.Update(ctx, config)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("更新配置失败", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		config := entity.NewSystemConfig("basic", "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称")

		mockRepo.On("Update", ctx, config).Return(errors.New("database error"))

		err := mockRepo.Update(ctx, config)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestSystemConfigRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("成功删除配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		configID := "config-001"

		mockRepo.On("Delete", ctx, configID).Return(nil)

		err := mockRepo.Delete(ctx, configID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("删除不存在的配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		configID := "nonexistent"

		mockRepo.On("Delete", ctx, configID).Return(gorm.ErrRecordNotFound)

		err := mockRepo.Delete(ctx, configID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestSystemConfigRepository_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		expectedConfig := entity.NewSystemConfig("basic", "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称")

		mockRepo.On("GetByID", ctx, expectedConfig.ID).Return(expectedConfig, nil)

		config, err := mockRepo.GetByID(ctx, expectedConfig.ID)

		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, expectedConfig.ID, config.ID)
		assert.Equal(t, "basic", config.Category)
		assert.Equal(t, "system_name", config.Key)
		mockRepo.AssertExpectations(t)
	})

	t.Run("配置不存在", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		mockRepo.On("GetByID", ctx, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		config, err := mockRepo.GetByID(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, config)
		mockRepo.AssertExpectations(t)
	})
}

func TestSystemConfigRepository_GetByKey(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		category := "basic"
		key := "system_name"
		expectedConfig := entity.NewSystemConfig(category, key, "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称")

		mockRepo.On("GetByKey", ctx, category, key).Return(expectedConfig, nil)

		config, err := mockRepo.GetByKey(ctx, category, key)

		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, category, config.Category)
		assert.Equal(t, key, config.Key)
		mockRepo.AssertExpectations(t)
	})

	t.Run("配置不存在", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		category := "basic"
		key := "nonexistent"

		mockRepo.On("GetByKey", ctx, category, key).Return(nil, gorm.ErrRecordNotFound)

		config, err := mockRepo.GetByKey(ctx, category, key)

		assert.Error(t, err)
		assert.Nil(t, config)
		mockRepo.AssertExpectations(t)
	})
}

func TestSystemConfigRepository_GetByCategory(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取分类配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		category := "basic"
		expectedConfigs := []*entity.SystemConfig{
			entity.NewSystemConfig(category, "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称"),
			entity.NewSystemConfig(category, "system_version", "1.0.0", entity.SystemConfigValueTypeString, "系统版本"),
		}

		mockRepo.On("GetByCategory", ctx, category).Return(expectedConfigs, nil)

		configs, err := mockRepo.GetByCategory(ctx, category)

		assert.NoError(t, err)
		assert.Len(t, configs, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("分类没有配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		category := "empty"

		mockRepo.On("GetByCategory", ctx, category).Return([]*entity.SystemConfig{}, nil)

		configs, err := mockRepo.GetByCategory(ctx, category)

		assert.NoError(t, err)
		assert.NotNil(t, configs)
		assert.Len(t, configs, 0)
		mockRepo.AssertExpectations(t)
	})

	t.Run("查询失败", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		category := "basic"

		mockRepo.On("GetByCategory", ctx, category).Return(nil, errors.New("database error"))

		configs, err := mockRepo.GetByCategory(ctx, category)

		assert.Error(t, err)
		assert.Nil(t, configs)
		mockRepo.AssertExpectations(t)
	})
}

func TestSystemConfigRepository_GetAll(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取所有配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		expectedConfigs := []*entity.SystemConfig{
			entity.NewSystemConfig("basic", "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称"),
			entity.NewSystemConfig("basic", "system_version", "1.0.0", entity.SystemConfigValueTypeString, "系统版本"),
			entity.NewSystemConfig("alarm", "alarm_enabled", "true", entity.SystemConfigValueTypeBool, "告警开关"),
		}

		mockRepo.On("GetAll", ctx).Return(expectedConfigs, nil)

		configs, err := mockRepo.GetAll(ctx)

		assert.NoError(t, err)
		assert.Len(t, configs, 3)
		mockRepo.AssertExpectations(t)
	})

	t.Run("没有配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		mockRepo.On("GetAll", ctx).Return([]*entity.SystemConfig{}, nil)

		configs, err := mockRepo.GetAll(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, configs)
		assert.Len(t, configs, 0)
		mockRepo.AssertExpectations(t)
	})

	t.Run("查询失败", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		mockRepo.On("GetAll", ctx).Return(nil, errors.New("database error"))

		configs, err := mockRepo.GetAll(ctx)

		assert.Error(t, err)
		assert.Nil(t, configs)
		mockRepo.AssertExpectations(t)
	})
}

func TestSystemConfigRepository_List(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取配置列表", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		expectedConfigs := []*entity.SystemConfig{
			entity.NewSystemConfig("basic", "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称"),
		}

		filter := &entity.SystemConfigFilter{
			Page:     1,
			PageSize: 20,
		}

		mockRepo.On("List", ctx, filter).Return(expectedConfigs, int64(1), nil)

		configs, total, err := mockRepo.List(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, configs, 1)
		assert.Equal(t, int64(1), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("带分类过滤", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		category := "basic"
		expectedConfigs := []*entity.SystemConfig{
			entity.NewSystemConfig(category, "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称"),
		}

		filter := &entity.SystemConfigFilter{
			Category: &category,
			Page:     1,
			PageSize: 20,
		}

		mockRepo.On("List", ctx, filter).Return(expectedConfigs, int64(1), nil)

		configs, total, err := mockRepo.List(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, configs, 1)
		assert.Equal(t, int64(1), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("带键过滤", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		key := "system"
		expectedConfigs := []*entity.SystemConfig{
			entity.NewSystemConfig("basic", "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称"),
			entity.NewSystemConfig("basic", "system_version", "1.0.0", entity.SystemConfigValueTypeString, "系统版本"),
		}

		filter := &entity.SystemConfigFilter{
			Key:      &key,
			Page:     1,
			PageSize: 20,
		}

		mockRepo.On("List", ctx, filter).Return(expectedConfigs, int64(2), nil)

		configs, total, err := mockRepo.List(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, configs, 2)
		assert.Equal(t, int64(2), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("空列表", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		filter := &entity.SystemConfigFilter{
			Page:     1,
			PageSize: 20,
		}

		mockRepo.On("List", ctx, filter).Return([]*entity.SystemConfig{}, int64(0), nil)

		configs, total, err := mockRepo.List(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, configs)
		assert.Len(t, configs, 0)
		assert.Equal(t, int64(0), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("查询失败", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		filter := &entity.SystemConfigFilter{
			Page:     1,
			PageSize: 20,
		}

		mockRepo.On("List", ctx, filter).Return(nil, int64(0), errors.New("database error"))

		configs, total, err := mockRepo.List(ctx, filter)

		assert.Error(t, err)
		assert.Nil(t, configs)
		assert.Equal(t, int64(0), total)
		mockRepo.AssertExpectations(t)
	})
}

func TestSystemConfigRepository_BatchUpdate(t *testing.T) {
	ctx := context.Background()

	t.Run("成功批量更新配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		configs := []*entity.SystemConfig{
			entity.NewSystemConfig("basic", "system_name", "新能源监控系统V2", entity.SystemConfigValueTypeString, "系统名称"),
			entity.NewSystemConfig("basic", "system_version", "2.0.0", entity.SystemConfigValueTypeString, "系统版本"),
		}

		mockRepo.On("BatchUpdate", ctx, configs).Return(nil)

		err := mockRepo.BatchUpdate(ctx, configs)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("批量更新失败", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		configs := []*entity.SystemConfig{
			entity.NewSystemConfig("basic", "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称"),
		}

		mockRepo.On("BatchUpdate", ctx, configs).Return(errors.New("database error"))

		err := mockRepo.BatchUpdate(ctx, configs)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestSystemConfigRepository_ExistsByKey(t *testing.T) {
	ctx := context.Background()

	t.Run("配置键存在", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		category := "basic"
		key := "system_name"

		mockRepo.On("ExistsByKey", ctx, category, key).Return(true, nil)

		exists, err := mockRepo.ExistsByKey(ctx, category, key)

		assert.NoError(t, err)
		assert.True(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("配置键不存在", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		category := "basic"
		key := "nonexistent"

		mockRepo.On("ExistsByKey", ctx, category, key).Return(false, nil)

		exists, err := mockRepo.ExistsByKey(ctx, category, key)

		assert.NoError(t, err)
		assert.False(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("查询失败", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)

		category := "basic"
		key := "system_name"

		mockRepo.On("ExistsByKey", ctx, category, key).Return(false, errors.New("database error"))

		exists, err := mockRepo.ExistsByKey(ctx, category, key)

		assert.Error(t, err)
		assert.False(t, exists)
		mockRepo.AssertExpectations(t)
	})
}
