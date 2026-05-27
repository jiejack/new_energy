package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockConfigRepoForHandler struct {
	mock.Mock
}

func (m *mockConfigRepoForHandler) Create(ctx context.Context, config *entity.SystemConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}
func (m *mockConfigRepoForHandler) Update(ctx context.Context, config *entity.SystemConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}
func (m *mockConfigRepoForHandler) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockConfigRepoForHandler) GetByID(ctx context.Context, id string) (*entity.SystemConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SystemConfig), args.Error(1)
}
func (m *mockConfigRepoForHandler) GetByKey(ctx context.Context, category, key string) (*entity.SystemConfig, error) {
	args := m.Called(ctx, category, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SystemConfig), args.Error(1)
}
func (m *mockConfigRepoForHandler) GetByCategory(ctx context.Context, category string) ([]*entity.SystemConfig, error) {
	args := m.Called(ctx, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.SystemConfig), args.Error(1)
}
func (m *mockConfigRepoForHandler) GetAll(ctx context.Context) ([]*entity.SystemConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.SystemConfig), args.Error(1)
}
func (m *mockConfigRepoForHandler) List(ctx context.Context, filter *entity.SystemConfigFilter) ([]*entity.SystemConfig, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.SystemConfig), args.Get(1).(int64), args.Error(2)
}
func (m *mockConfigRepoForHandler) BatchUpdate(ctx context.Context, configs []*entity.SystemConfig) error {
	args := m.Called(ctx, configs)
	return args.Error(0)
}
func (m *mockConfigRepoForHandler) ExistsByKey(ctx context.Context, category, key string) (bool, error) {
	args := m.Called(ctx, category, key)
	return args.Bool(0), args.Error(1)
}

type mockLogRepoForConfig struct {
	mock.Mock
}

func (m *mockLogRepoForConfig) Create(ctx context.Context, log *entity.OperationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}
func (m *mockLogRepoForConfig) GetByID(ctx context.Context, id string) (*entity.OperationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OperationLog), args.Error(1)
}
func (m *mockLogRepoForConfig) List(ctx context.Context, query *repository.OperationLogQuery) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}
func (m *mockLogRepoForConfig) DeleteBefore(ctx context.Context, before int64) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func setupConfigHandler(configRepo *mockConfigRepoForHandler, logRepo *mockLogRepoForConfig) (*ConfigHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewConfigService(configRepo, logRepo)
	handler := NewConfigHandler(svc)
	r := gin.New()
	return handler, r
}

func TestConfigHandler_GetAllConfigs_Success(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.GET("/configs", handler.GetAllConfigs)

	configRepo.On("GetAll", mock.Anything).Return([]*entity.SystemConfig{}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/configs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConfigHandler_GetAllConfigs_Error(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.GET("/configs", handler.GetAllConfigs)

	configRepo.On("GetAll", mock.Anything).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/configs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestConfigHandler_GetConfigsByCategory_Success(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.GET("/configs/:category", handler.GetConfigsByCategory)

	configRepo.On("GetByCategory", mock.Anything, "basic").Return([]*entity.SystemConfig{}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/configs/basic", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConfigHandler_GetConfigsByCategory_Error(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.GET("/configs/:category", handler.GetConfigsByCategory)

	configRepo.On("GetByCategory", mock.Anything, "basic").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/configs/basic", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestConfigHandler_GetConfig_Success(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.GET("/configs/:category/:key", handler.GetConfig)

	cfg := entity.NewSystemConfig("basic", "system_name", "test", entity.SystemConfigValueTypeString, "desc")
	configRepo.On("GetByKey", mock.Anything, "basic", "system_name").Return(cfg, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/configs/basic/system_name", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConfigHandler_GetConfig_NotFound(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.GET("/configs/:category/:key", handler.GetConfig)

	configRepo.On("GetByKey", mock.Anything, "basic", "nonexistent").Return(nil, service.ErrConfigNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/configs/basic/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestConfigHandler_UpdateConfig_Success(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.PUT("/configs/:category/:key", handler.UpdateConfig)

	cfg := entity.NewSystemConfig("basic", "system_name", "old", entity.SystemConfigValueTypeString, "desc")
	configRepo.On("GetByKey", mock.Anything, "basic", "system_name").Return(cfg, nil)
	configRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.SystemConfig")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	body := service.UpdateConfigRequest{Value: "new_value", ValueType: "string"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/configs/basic/system_name", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConfigHandler_UpdateConfig_InvalidJSON(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.PUT("/configs/:category/:key", handler.UpdateConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/configs/basic/system_name", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfigHandler_UpdateConfig_NotFound(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.PUT("/configs/:category/:key", handler.UpdateConfig)

	configRepo.On("GetByKey", mock.Anything, "basic", "nonexistent").Return(nil, assert.AnError)

	body := service.UpdateConfigRequest{Value: "new_value"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/configs/basic/nonexistent", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestConfigHandler_CreateConfig_Success(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.POST("/configs", handler.CreateConfig)

	configRepo.On("ExistsByKey", mock.Anything, "basic", "new_key").Return(false, nil)
	configRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.SystemConfig")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	body := service.CreateConfigRequest{Category: "basic", Key: "new_key", Value: "val", ValueType: "string"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/configs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestConfigHandler_CreateConfig_InvalidJSON(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.POST("/configs", handler.CreateConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/configs", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfigHandler_CreateConfig_KeyExists(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.POST("/configs", handler.CreateConfig)

	configRepo.On("ExistsByKey", mock.Anything, "basic", "existing_key").Return(true, nil)

	body := service.CreateConfigRequest{Category: "basic", Key: "existing_key", Value: "val", ValueType: "string"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/configs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfigHandler_DeleteConfig_Success(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.DELETE("/configs/:category/:key", handler.DeleteConfig)

	cfg := entity.NewSystemConfig("basic", "system_name", "val", entity.SystemConfigValueTypeString, "desc")
	configRepo.On("GetByKey", mock.Anything, "basic", "system_name").Return(cfg, nil)
	configRepo.On("Delete", mock.Anything, cfg.ID).Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/configs/basic/system_name", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestConfigHandler_DeleteConfig_NotFound(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.DELETE("/configs/:category/:key", handler.DeleteConfig)

	configRepo.On("GetByKey", mock.Anything, "basic", "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/configs/basic/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestConfigHandler_BatchUpdateConfigs_Success(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.POST("/configs/batch", handler.BatchUpdateConfigs)

	cfg := entity.NewSystemConfig("basic", "k1", "v1", entity.SystemConfigValueTypeString, "desc")
	configRepo.On("GetByKey", mock.Anything, "basic", "k1").Return(cfg, nil)
	configRepo.On("BatchUpdate", mock.Anything, mock.Anything).Return(nil)
	logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	body := service.BatchUpdateConfigRequest{
		Configs: []service.ConfigUpdateItem{{Category: "basic", Key: "k1", Value: "v2"}},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/configs/batch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConfigHandler_BatchUpdateConfigs_InvalidJSON(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.POST("/configs/batch", handler.BatchUpdateConfigs)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/configs/batch", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfigHandler_ListConfigs_Success(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.GET("/configs/list", handler.ListConfigs)

	configRepo.On("List", mock.Anything, mock.AnythingOfType("*entity.SystemConfigFilter")).Return([]*entity.SystemConfig{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/configs/list", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConfigHandler_ListConfigs_WithFilters(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	handler, r := setupConfigHandler(configRepo, logRepo)

	r.GET("/configs/list", handler.ListConfigs)

	configRepo.On("List", mock.Anything, mock.AnythingOfType("*entity.SystemConfigFilter")).Return([]*entity.SystemConfig{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/configs/list?category=basic&page=2&page_size=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConfigHandler_NewConfigHandler(t *testing.T) {
	configRepo := new(mockConfigRepoForHandler)
	logRepo := new(mockLogRepoForConfig)
	svc := service.NewConfigService(configRepo, logRepo)
	handler := NewConfigHandler(svc)
	assert.NotNil(t, handler)
}
