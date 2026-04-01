package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// MockOperationLogRepository 操作日志仓储Mock
type MockOperationLogRepository struct {
	mock.Mock
}

func (m *MockOperationLogRepository) Create(ctx context.Context, log *entity.OperationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockOperationLogRepository) GetByID(ctx context.Context, id string) (*entity.OperationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OperationLog), args.Error(1)
}

func (m *MockOperationLogRepository) List(ctx context.Context, userID *string, action *string, startTime, endTime int64, page, pageSize int) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, userID, action, startTime, endTime, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockOperationLogRepository) GetByUserID(ctx context.Context, userID string, page, pageSize int) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestConfigHandler_GetAllConfigs(t *testing.T) {
	t.Run("成功获取所有配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)
		mockLogRepo := new(MockOperationLogRepository)
		configService := service.NewConfigService(mockRepo, mockLogRepo)
		handler := NewConfigHandler(configService)

		expectedConfigs := []*entity.SystemConfig{
			entity.NewSystemConfig("basic", "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称"),
			entity.NewSystemConfig("basic", "system_version", "1.0.0", entity.SystemConfigValueTypeString, "系统版本"),
		}

		mockRepo.On("GetAll", mock.Anything).Return(expectedConfigs, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/configs", nil)

		handler.GetAllConfigs(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("服务返回错误", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)
		mockLogRepo := new(MockOperationLogRepository)
		configService := service.NewConfigService(mockRepo, mockLogRepo)
		handler := NewConfigHandler(configService)

		mockRepo.On("GetAll", mock.Anything).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/configs", nil)

		handler.GetAllConfigs(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestConfigHandler_GetConfigsByCategory(t *testing.T) {
	t.Run("成功获取分类配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)
		mockLogRepo := new(MockOperationLogRepository)
		configService := service.NewConfigService(mockRepo, mockLogRepo)
		handler := NewConfigHandler(configService)

		category := "basic"
		expectedConfigs := []*entity.SystemConfig{
			entity.NewSystemConfig(category, "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称"),
		}

		mockRepo.On("GetByCategory", mock.Anything, category).Return(expectedConfigs, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/configs/"+category, nil)
		c.Params = gin.Params{{Key: "category", Value: category}}

		handler.GetConfigsByCategory(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestConfigHandler_GetConfig(t *testing.T) {
	t.Run("成功获取配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)
		mockLogRepo := new(MockOperationLogRepository)
		configService := service.NewConfigService(mockRepo, mockLogRepo)
		handler := NewConfigHandler(configService)

		category := "basic"
		key := "system_name"
		expectedConfig := entity.NewSystemConfig(category, key, "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称")

		mockRepo.On("GetByKey", mock.Anything, category, key).Return(expectedConfig, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/configs/"+category+"/"+key, nil)
		c.Params = gin.Params{{Key: "category", Value: category}, {Key: "key", Value: key}}

		handler.GetConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("配置不存在", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)
		mockLogRepo := new(MockOperationLogRepository)
		configService := service.NewConfigService(mockRepo, mockLogRepo)
		handler := NewConfigHandler(configService)

		category := "basic"
		key := "nonexistent"
		mockRepo.On("GetByKey", mock.Anything, category, key).Return(nil, errors.New("not found"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/configs/"+category+"/"+key, nil)
		c.Params = gin.Params{{Key: "category", Value: category}, {Key: "key", Value: key}}

		handler.GetConfig(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestConfigHandler_UpdateConfig(t *testing.T) {
	t.Run("成功更新配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)
		mockLogRepo := new(MockOperationLogRepository)
		configService := service.NewConfigService(mockRepo, mockLogRepo)
		handler := NewConfigHandler(configService)

		category := "basic"
		key := "system_name"
		req := map[string]interface{}{
			"value":       "新能源监控系统V2",
			"value_type":  "string",
			"description": "系统名称",
		}

		expectedConfig := entity.NewSystemConfig(category, key, "新能源监控系统V2", entity.SystemConfigValueTypeString, "系统名称")

		mockRepo.On("GetByKey", mock.Anything, category, key).Return(expectedConfig, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.SystemConfig")).Return(nil)
		mockLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/configs/"+category+"/"+key, bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "category", Value: category}, {Key: "key", Value: key}}

		handler.UpdateConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("无效的请求参数", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)
		mockLogRepo := new(MockOperationLogRepository)
		configService := service.NewConfigService(mockRepo, mockLogRepo)
		handler := NewConfigHandler(configService)

		category := "basic"
		key := "system_name"

		body := []byte(`{"invalid": "data"}`)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/configs/"+category+"/"+key, bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "category", Value: category}, {Key: "key", Value: key}}

		handler.UpdateConfig(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestConfigHandler_CreateConfig(t *testing.T) {
	t.Run("成功创建配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)
		mockLogRepo := new(MockOperationLogRepository)
		configService := service.NewConfigService(mockRepo, mockLogRepo)
		handler := NewConfigHandler(configService)

		req := map[string]interface{}{
			"category":    "basic",
			"key":         "new_config",
			"value":       "test value",
			"value_type":  "string",
			"description": "测试配置",
		}

		mockRepo.On("ExistsByKey", mock.Anything, "basic", "new_config").Return(false, nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.SystemConfig")).Return(nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/configs", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateConfig(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("无效的请求参数", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)
		mockLogRepo := new(MockOperationLogRepository)
		configService := service.NewConfigService(mockRepo, mockLogRepo)
		handler := NewConfigHandler(configService)

		body := []byte(`{"invalid": "data"}`)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/configs", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateConfig(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestConfigHandler_DeleteConfig(t *testing.T) {
	t.Run("成功删除配置", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)
		mockLogRepo := new(MockOperationLogRepository)
		configService := service.NewConfigService(mockRepo, mockLogRepo)
		handler := NewConfigHandler(configService)

		category := "basic"
		key := "test_config"
		expectedConfig := entity.NewSystemConfig(category, key, "test", entity.SystemConfigValueTypeString, "测试配置")

		mockRepo.On("GetByKey", mock.Anything, category, key).Return(expectedConfig, nil)
		mockRepo.On("Delete", mock.Anything, expectedConfig.ID).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/configs/"+category+"/"+key, nil)
		c.Params = gin.Params{{Key: "category", Value: category}, {Key: "key", Value: key}}

		handler.DeleteConfig(c)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("配置不存在", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)
		mockLogRepo := new(MockOperationLogRepository)
		configService := service.NewConfigService(mockRepo, mockLogRepo)
		handler := NewConfigHandler(configService)

		category := "basic"
		key := "nonexistent"

		mockRepo.On("GetByKey", mock.Anything, category, key).Return(nil, errors.New("not found"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/configs/"+category+"/"+key, nil)
		c.Params = gin.Params{{Key: "category", Value: category}, {Key: "key", Value: key}}

		handler.DeleteConfig(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestConfigHandler_ListConfigs(t *testing.T) {
	t.Run("成功获取配置列表", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)
		mockLogRepo := new(MockOperationLogRepository)
		configService := service.NewConfigService(mockRepo, mockLogRepo)
		handler := NewConfigHandler(configService)

		expectedConfigs := []*entity.SystemConfig{
			entity.NewSystemConfig("basic", "system_name", "新能源监控系统", entity.SystemConfigValueTypeString, "系统名称"),
		}

		mockRepo.On("List", mock.Anything, mock.AnythingOfType("*entity.SystemConfigFilter")).Return(expectedConfigs, int64(1), nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/configs/list?page=1&page_size=20", nil)

		handler.ListConfigs(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("服务返回错误", func(t *testing.T) {
		mockRepo := new(MockSystemConfigRepository)
		mockLogRepo := new(MockOperationLogRepository)
		configService := service.NewConfigService(mockRepo, mockLogRepo)
		handler := NewConfigHandler(configService)

		mockRepo.On("List", mock.Anything, mock.AnythingOfType("*entity.SystemConfigFilter")).Return(nil, int64(0), errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/configs/list", nil)

		handler.ListConfigs(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
