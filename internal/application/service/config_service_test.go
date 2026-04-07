package service

import (
	"context"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockSystemConfigRepository 系统配置仓储Mock
type MockSystemConfigRepository struct {
	mock.Mock
	configs map[string]*entity.SystemConfig
}

func NewMockSystemConfigRepository() *MockSystemConfigRepository {
	return &MockSystemConfigRepository{
		configs: make(map[string]*entity.SystemConfig),
	}
}

func (m *MockSystemConfigRepository) Create(ctx context.Context, config *entity.SystemConfig) error {
	args := m.Called(ctx, config)
	if args.Error(0) == nil {
		m.configs[config.ID] = config
	}
	return args.Error(0)
}

func (m *MockSystemConfigRepository) Update(ctx context.Context, config *entity.SystemConfig) error {
	args := m.Called(ctx, config)
	if args.Error(0) == nil {
		m.configs[config.ID] = config
	}
	return args.Error(0)
}

func (m *MockSystemConfigRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if args.Error(0) == nil {
		delete(m.configs, id)
	}
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
	return args.Get(0).([]*entity.SystemConfig), args.Get(1).(int64), args.Error(2)
}

func (m *MockSystemConfigRepository) BatchUpdate(ctx context.Context, configs []*entity.SystemConfig) error {
	args := m.Called(ctx, configs)
	return args.Error(0)
}

func (m *MockSystemConfigRepository) ExistsByKey(ctx context.Context, category, key string) (bool, error) {
	args := m.Called(ctx, category, key)
	return args.Get(0).(bool), args.Error(1)
}

// MockOperationLogRepositoryForConfigService 操作日志仓储Mock
type MockOperationLogRepositoryForConfigService struct {
	mock.Mock
}

func (m *MockOperationLogRepositoryForConfigService) Create(ctx context.Context, log *entity.OperationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockOperationLogRepositoryForConfigService) GetByID(ctx context.Context, id string) (*entity.OperationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OperationLog), args.Error(1)
}

func (m *MockOperationLogRepositoryForConfigService) List(ctx context.Context, userID *string, action *string, startTime, endTime int64, page, pageSize int) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, userID, action, startTime, endTime, page, pageSize)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockOperationLogRepositoryForConfigService) GetByUserID(ctx context.Context, userID string, page, pageSize int) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, userID, page, pageSize)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func TestConfigService_CreateConfig_Success(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 设置Mock期望
	mockConfigRepo.On("ExistsByKey", ctx, "basic", "system_name").Return(false, nil)
	mockConfigRepo.On("Create", ctx, mock.AnythingOfType("*entity.SystemConfig")).Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	req := &CreateConfigRequest{
		Category:    "basic",
		Key:         "system_name",
		Value:       "新能源监控系统",
		ValueType:   "string",
		Description: "系统名称",
	}
	config, err := service.CreateConfig(ctx, req, "admin-001")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "basic", config.Category)
	assert.Equal(t, "system_name", config.Key)
	assert.Equal(t, "新能源监控系统", config.Value)
	assert.Equal(t, entity.SystemConfigValueTypeString, config.ValueType)
	assert.Equal(t, "系统名称", config.Description)

	mockConfigRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestConfigService_CreateConfig_KeyExists(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 设置Mock期望 - 键已存在
	mockConfigRepo.On("ExistsByKey", ctx, "basic", "system_name").Return(true, nil)

	// 执行测试
	req := &CreateConfigRequest{
		Category:  "basic",
		Key:       "system_name",
		Value:     "新能源监控系统",
		ValueType: "string",
	}
	config, err := service.CreateConfig(ctx, req, "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrConfigKeyExists, err)
	assert.Nil(t, config)

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_CreateConfig_InvalidValueType(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 设置Mock期望
	mockConfigRepo.On("ExistsByKey", ctx, "basic", "test_key").Return(false, nil)

	// 执行测试 - 使用无效的值类型
	req := &CreateConfigRequest{
		Category:  "basic",
		Key:       "test_key",
		Value:     "test_value",
		ValueType: "invalid_type",
	}
	config, err := service.CreateConfig(ctx, req, "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidValueType, err)
	assert.Nil(t, config)

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_CreateConfig_InvalidIntValue(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 设置Mock期望
	mockConfigRepo.On("ExistsByKey", ctx, "basic", "test_int").Return(false, nil)

	// 执行测试 - 使用无效的整数值
	req := &CreateConfigRequest{
		Category:  "basic",
		Key:       "test_int",
		Value:     "not_a_number",
		ValueType: "int",
	}
	config, err := service.CreateConfig(ctx, req, "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "value conversion failed")
	assert.Nil(t, config)

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_CreateConfig_InvalidBoolValue(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 设置Mock期望
	mockConfigRepo.On("ExistsByKey", ctx, "basic", "test_bool").Return(false, nil)

	// 执行测试 - 使用无效的布尔值
	req := &CreateConfigRequest{
		Category:  "basic",
		Key:       "test_bool",
		Value:     "not_a_bool",
		ValueType: "bool",
	}
	config, err := service.CreateConfig(ctx, req, "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "value conversion failed")
	assert.Nil(t, config)

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_CreateConfig_InvalidJSONValue(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 设置Mock期望
	mockConfigRepo.On("ExistsByKey", ctx, "basic", "test_json").Return(false, nil)

	// 执行测试 - 使用无效的JSON值
	req := &CreateConfigRequest{
		Category:  "basic",
		Key:       "test_json",
		Value:     "{invalid json}",
		ValueType: "json",
	}
	config, err := service.CreateConfig(ctx, req, "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "value conversion failed")
	assert.Nil(t, config)

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_UpdateConfig_Success(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 准备测试数据
	existingConfig := entity.NewSystemConfig("basic", "system_name", "旧系统名称", entity.SystemConfigValueTypeString, "系统名称")
	existingConfig.ID = "config-001"

	// 设置Mock期望
	mockConfigRepo.On("GetByKey", ctx, "basic", "system_name").Return(existingConfig, nil)
	mockConfigRepo.On("Update", ctx, mock.AnythingOfType("*entity.SystemConfig")).Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	req := &UpdateConfigRequest{
		Value:       "新系统名称",
		Description: "更新后的系统名称",
	}
	config, err := service.UpdateConfig(ctx, "basic", "system_name", req, "admin-001")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "新系统名称", config.Value)
	assert.Equal(t, "更新后的系统名称", config.Description)

	mockConfigRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestConfigService_UpdateConfig_NotFound(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 设置Mock期望 - 配置不存在
	mockConfigRepo.On("GetByKey", ctx, "basic", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	// 执行测试
	req := &UpdateConfigRequest{
		Value: "新值",
	}
	config, err := service.UpdateConfig(ctx, "basic", "nonexistent", req, "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrConfigNotFound, err)
	assert.Nil(t, config)

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_DeleteConfig_Success(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 准备测试数据
	existingConfig := entity.NewSystemConfig("basic", "system_name", "系统名称", entity.SystemConfigValueTypeString, "系统名称")
	existingConfig.ID = "config-001"

	// 设置Mock期望
	mockConfigRepo.On("GetByKey", ctx, "basic", "system_name").Return(existingConfig, nil)
	mockConfigRepo.On("Delete", ctx, "config-001").Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	err := service.DeleteConfig(ctx, "basic", "system_name", "admin-001")

	// 验证结果
	assert.NoError(t, err)

	mockConfigRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestConfigService_DeleteConfig_NotFound(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 设置Mock期望 - 配置不存在
	mockConfigRepo.On("GetByKey", ctx, "basic", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	// 执行测试
	err := service.DeleteConfig(ctx, "basic", "nonexistent", "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrConfigNotFound, err)

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_GetConfig_Success(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 准备测试数据
	existingConfig := entity.NewSystemConfig("basic", "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称")

	// 设置Mock期望
	mockConfigRepo.On("GetByKey", ctx, "basic", "system_name").Return(existingConfig, nil)

	// 执行测试
	config, err := service.GetConfig(ctx, "basic", "system_name")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "basic", config.Category)
	assert.Equal(t, "system_name", config.Key)

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_GetConfigsByCategory_Success(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 准备测试数据
	configs := []*entity.SystemConfig{
		entity.NewSystemConfig("basic", "system_name", "系统名称", entity.SystemConfigValueTypeString, "系统名称"),
		entity.NewSystemConfig("basic", "logo", "logo.png", entity.SystemConfigValueTypeString, "系统Logo"),
	}

	// 设置Mock期望
	mockConfigRepo.On("GetByCategory", ctx, "basic").Return(configs, nil)

	// 执行测试
	resp, err := service.GetConfigsByCategory(ctx, "basic")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "basic", resp.Category)
	assert.Len(t, resp.Configs, 2)

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_GetAllConfigs_Success(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 准备测试数据
	configs := []*entity.SystemConfig{
		entity.NewSystemConfig("basic", "system_name", "系统名称", entity.SystemConfigValueTypeString, "系统名称"),
		entity.NewSystemConfig("alarm", "default_level", "3", entity.SystemConfigValueTypeInt, "默认告警级别"),
	}

	// 设置Mock期望
	mockConfigRepo.On("GetAll", ctx).Return(configs, nil)

	// 执行测试
	result, err := service.GetAllConfigs(ctx)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_ListConfigs_Success(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 准备测试数据
	configs := []*entity.SystemConfig{
		entity.NewSystemConfig("basic", "system_name", "系统名称", entity.SystemConfigValueTypeString, "系统名称"),
		entity.NewSystemConfig("basic", "logo", "logo.png", entity.SystemConfigValueTypeString, "系统Logo"),
	}

	// 设置Mock期望
	filter := &entity.SystemConfigFilter{
		Category: strPtr("basic"),
		Page:     1,
		PageSize: 20,
	}
	mockConfigRepo.On("List", ctx, filter).Return(configs, int64(2), nil)

	// 执行测试
	resp, err := service.ListConfigs(ctx, filter)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Configs, 2)
	assert.Equal(t, int64(2), resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 20, resp.PageSize)

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_BatchUpdateConfigs_Success(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 准备测试数据
	config1 := entity.NewSystemConfig("basic", "system_name", "旧名称", entity.SystemConfigValueTypeString, "系统名称")
	config2 := entity.NewSystemConfig("basic", "logo", "old_logo.png", entity.SystemConfigValueTypeString, "系统Logo")

	// 设置Mock期望
	mockConfigRepo.On("GetByKey", ctx, "basic", "system_name").Return(config1, nil)
	mockConfigRepo.On("GetByKey", ctx, "basic", "logo").Return(config2, nil)
	mockConfigRepo.On("BatchUpdate", ctx, mock.Anything).Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	// 执行测试
	req := &BatchUpdateConfigRequest{
		Configs: []ConfigUpdateItem{
			{Category: "basic", Key: "system_name", Value: "新名称"},
			{Category: "basic", Key: "logo", Value: "new_logo.png"},
		},
	}
	err := service.BatchUpdateConfigs(ctx, req, "admin-001")

	// 验证结果
	assert.NoError(t, err)

	mockConfigRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestConfigService_BatchUpdateConfigs_ConfigNotFound(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 设置Mock期望 - 配置不存在
	mockConfigRepo.On("GetByKey", ctx, "basic", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	// 执行测试
	req := &BatchUpdateConfigRequest{
		Configs: []ConfigUpdateItem{
			{Category: "basic", Key: "nonexistent", Value: "新值"},
		},
	}
	err := service.BatchUpdateConfigs(ctx, req, "admin-001")

	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config not found")

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_GetConfigAsInt_Success(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 准备测试数据
	config := entity.NewSystemConfig("alarm", "default_level", "3", entity.SystemConfigValueTypeInt, "默认告警级别")

	// 设置Mock期望
	mockConfigRepo.On("GetByKey", ctx, "alarm", "default_level").Return(config, nil)

	// 执行测试
	val, err := service.GetConfigAsInt(ctx, "alarm", "default_level")

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, 3, val)

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_GetConfigAsBool_Success(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 准备测试数据
	config := entity.NewSystemConfig("alarm", "sound_enabled", "true", entity.SystemConfigValueTypeBool, "声音告警")

	// 设置Mock期望
	mockConfigRepo.On("GetByKey", ctx, "alarm", "sound_enabled").Return(config, nil)

	// 执行测试
	val, err := service.GetConfigAsBool(ctx, "alarm", "sound_enabled")

	// 验证结果
	assert.NoError(t, err)
	assert.True(t, val)

	mockConfigRepo.AssertExpectations(t)
}

func TestConfigService_GetConfigAsJSON_Success(t *testing.T) {
	ctx := context.Background()

	mockConfigRepo := NewMockSystemConfigRepository()
	mockLogRepo := new(MockOperationLogRepositoryForConfigService)

	service := NewConfigService(mockConfigRepo, mockLogRepo)

	// 准备测试数据
	jsonValue := `{"key1": "value1", "key2": 123}`
	config := entity.NewSystemConfig("display", "theme_config", jsonValue, entity.SystemConfigValueTypeJSON, "主题配置")

	// 设置Mock期望
	mockConfigRepo.On("GetByKey", ctx, "display", "theme_config").Return(config, nil)

	// 执行测试
	var result map[string]interface{}
	err := service.GetConfigAsJSON(ctx, "display", "theme_config", &result)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, float64(123), result["key2"])

	mockConfigRepo.AssertExpectations(t)
}
