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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockNotificationConfigRepo struct {
	mock.Mock
}

func (m *mockNotificationConfigRepo) Create(ctx context.Context, config *entity.NotificationConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *mockNotificationConfigRepo) Update(ctx context.Context, config *entity.NotificationConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *mockNotificationConfigRepo) GetByType(ctx context.Context, notifType entity.NotificationType) (*entity.NotificationConfig, error) {
	args := m.Called(ctx, notifType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.NotificationConfig), args.Error(1)
}

func (m *mockNotificationConfigRepo) GetAll(ctx context.Context) ([]*entity.NotificationConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.NotificationConfig), args.Error(1)
}

func TestNotificationConfigHandler_GetAllConfigs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	configs := []*entity.NotificationConfig{
		{ID: "1", Type: entity.NotificationTypeEmail, Name: "Email", Enabled: true},
	}
	mockRepo.On("GetAll", mock.Anything).Return(configs, nil)

	r := gin.New()
	r.GET("/notification-configs", handler.GetAllConfigs)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notification-configs", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotificationConfigHandler_GetAllConfigs_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	mockRepo.On("GetAll", mock.Anything).Return(nil, assert.AnError)

	r := gin.New()
	r.GET("/notification-configs", handler.GetAllConfigs)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notification-configs", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNotificationConfigHandler_GetConfigByType_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	config := &entity.NotificationConfig{ID: "1", Type: entity.NotificationTypeEmail, Name: "Email", Enabled: true}
	mockRepo.On("GetByType", mock.Anything, entity.NotificationTypeEmail).Return(config, nil)

	r := gin.New()
	r.GET("/notification-configs/:type", handler.GetConfigByType)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notification-configs/email", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotificationConfigHandler_GetConfigByType_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	mockRepo.On("GetByType", mock.Anything, entity.NotificationTypeEmail).Return(nil, assert.AnError)

	r := gin.New()
	r.GET("/notification-configs/:type", handler.GetConfigByType)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notification-configs/email", nil)

	r.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestNotificationConfigHandler_UpdateConfig_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	r := gin.New()
	r.PUT("/notification-configs/:type", handler.UpdateConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notification-configs/email", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNotificationConfigHandler_UpdateConfig_CreateNew(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	mockRepo.On("GetByType", mock.Anything, entity.NotificationTypeEmail).Return(nil, assert.AnError)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.NotificationConfig")).Return(nil)

	r := gin.New()
	r.PUT("/notification-configs/:type", handler.UpdateConfig)

	body := service.UpdateNotificationConfigRequest{
		Name:   "Email Config",
		Config: map[string]interface{}{"smtp_host": "smtp.example.com"},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notification-configs/email", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotificationConfigHandler_UpdateConfig_UpdateExisting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	existing := &entity.NotificationConfig{ID: "1", Type: entity.NotificationTypeEmail, Name: "Old", Enabled: true}
	mockRepo.On("GetByType", mock.Anything, entity.NotificationTypeEmail).Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.NotificationConfig")).Return(nil)

	r := gin.New()
	r.PUT("/notification-configs/:type", handler.UpdateConfig)

	body := service.UpdateNotificationConfigRequest{
		Name:   "Updated Email Config",
		Config: map[string]interface{}{"smtp_host": "smtp.new.com"},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notification-configs/email", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotificationConfigHandler_EnableConfig_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	config := &entity.NotificationConfig{ID: "1", Type: entity.NotificationTypeEmail, Name: "Email", Enabled: false}
	mockRepo.On("GetByType", mock.Anything, entity.NotificationTypeEmail).Return(config, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.NotificationConfig")).Return(nil)

	r := gin.New()
	r.PUT("/notification-configs/:type/enable", handler.EnableConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notification-configs/email/enable", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotificationConfigHandler_EnableConfig_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	mockRepo.On("GetByType", mock.Anything, entity.NotificationTypeEmail).Return(nil, assert.AnError)

	r := gin.New()
	r.PUT("/notification-configs/:type/enable", handler.EnableConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notification-configs/email/enable", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNotificationConfigHandler_DisableConfig_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	config := &entity.NotificationConfig{ID: "1", Type: entity.NotificationTypeEmail, Name: "Email", Enabled: true}
	mockRepo.On("GetByType", mock.Anything, entity.NotificationTypeEmail).Return(config, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.NotificationConfig")).Return(nil)

	r := gin.New()
	r.PUT("/notification-configs/:type/disable", handler.DisableConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notification-configs/email/disable", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotificationConfigHandler_TestConfig_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	config := &entity.NotificationConfig{
		ID:     "1",
		Type:   entity.NotificationTypeEmail,
		Name:   "Email",
		Config: entity.JSONMap{"smtp_host": "smtp.example.com"},
	}
	mockRepo.On("GetByType", mock.Anything, entity.NotificationTypeEmail).Return(config, nil)

	r := gin.New()
	r.POST("/notification-configs/:type/test", handler.TestConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notification-configs/email/test", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotificationConfigHandler_TestConfig_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	mockRepo.On("GetByType", mock.Anything, entity.NotificationTypeSMS).Return(nil, assert.AnError)

	r := gin.New()
	r.POST("/notification-configs/:type/test", handler.TestConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notification-configs/sms/test", nil)

	r.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestNotificationConfigHandler_NewHandler(t *testing.T) {
	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)
	assert.NotNil(t, handler)
}

func TestNotificationConfigHandler_GetConfigByType_AllTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	types := []entity.NotificationType{
		entity.NotificationTypeEmail,
		entity.NotificationTypeSMS,
		entity.NotificationTypeWebhook,
		entity.NotificationTypeWeChat,
	}

	for _, notifType := range types {
		mockRepo := new(mockNotificationConfigRepo)
		svc := service.NewNotificationConfigService(mockRepo)
		handler := NewNotificationConfigHandler(svc)

		config := &entity.NotificationConfig{ID: "1", Type: notifType, Name: string(notifType), Enabled: true}
		mockRepo.On("GetByType", mock.Anything, notifType).Return(config, nil)

		r := gin.New()
		r.GET("/notification-configs/:type", handler.GetConfigByType)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/notification-configs/"+string(notifType), nil)

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestNotificationConfigHandler_DisableConfig_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	mockRepo.On("GetByType", mock.Anything, entity.NotificationTypeWeChat).Return(nil, assert.AnError)

	r := gin.New()
	r.PUT("/notification-configs/:type/disable", handler.DisableConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notification-configs/wechat/disable", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNotificationConfigHandler_TestConfig_MissingSMTPHost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockNotificationConfigRepo)
	svc := service.NewNotificationConfigService(mockRepo)
	handler := NewNotificationConfigHandler(svc)

	config := &entity.NotificationConfig{
		ID:     "1",
		Type:   entity.NotificationTypeEmail,
		Name:   "Email",
		Config: entity.JSONMap{},
	}
	mockRepo.On("GetByType", mock.Anything, entity.NotificationTypeEmail).Return(config, nil)

	r := gin.New()
	r.POST("/notification-configs/:type/test", handler.TestConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notification-configs/email/test", nil)

	r.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
}
