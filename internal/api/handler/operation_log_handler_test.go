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

type mockOperationLogRepoForHandler struct {
	mock.Mock
}

func (m *mockOperationLogRepoForHandler) Create(ctx context.Context, log *entity.OperationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *mockOperationLogRepoForHandler) GetByID(ctx context.Context, id string) (*entity.OperationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OperationLog), args.Error(1)
}

func (m *mockOperationLogRepoForHandler) List(ctx context.Context, query *repository.OperationLogQuery) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func (m *mockOperationLogRepoForHandler) DeleteBefore(ctx context.Context, before int64) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func TestOperationLogHandler_CreateLog_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockOperationLogRepoForHandler)
	svc := service.NewOperationLogService(mockRepo)
	handler := NewOperationLogHandler(svc)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	r := gin.New()
	r.POST("/operation-logs", handler.CreateLog)

	body := service.CreateOperationLogRequest{
		UserID:       "user-001",
		Username:     "admin",
		Action:       entity.ActionLogin,
		ResourceType: "system",
		ResourceID:   "sys-001",
		Details:      map[string]interface{}{"ip": "192.168.1.1"},
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/operation-logs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationLogHandler_CreateLog_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockOperationLogRepoForHandler)
	svc := service.NewOperationLogService(mockRepo)
	handler := NewOperationLogHandler(svc)

	r := gin.New()
	r.POST("/operation-logs", handler.CreateLog)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/operation-logs", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOperationLogHandler_CreateLog_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockOperationLogRepoForHandler)
	svc := service.NewOperationLogService(mockRepo)
	handler := NewOperationLogHandler(svc)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(assert.AnError)

	r := gin.New()
	r.POST("/operation-logs", handler.CreateLog)

	body := service.CreateOperationLogRequest{
		UserID:   "user-001",
		Username: "admin",
		Action:   entity.ActionLogin,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/operation-logs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestOperationLogHandler_GetLog_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockOperationLogRepoForHandler)
	svc := service.NewOperationLogService(mockRepo)
	handler := NewOperationLogHandler(svc)

	log := entity.NewOperationLog("user-001", "admin", entity.ActionLogin)
	mockRepo.On("GetByID", mock.Anything, "log-001").Return(log, nil)

	r := gin.New()
	r.GET("/operation-logs/:id", handler.GetLog)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/operation-logs/log-001", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationLogHandler_GetLog_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockOperationLogRepoForHandler)
	svc := service.NewOperationLogService(mockRepo)
	handler := NewOperationLogHandler(svc)

	mockRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	r := gin.New()
	r.GET("/operation-logs/:id", handler.GetLog)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/operation-logs/nonexistent", nil)

	r.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestOperationLogHandler_ListLogs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockOperationLogRepoForHandler)
	svc := service.NewOperationLogService(mockRepo)
	handler := NewOperationLogHandler(svc)

	logs := []*entity.OperationLog{
		entity.NewOperationLog("user-001", "admin", entity.ActionLogin),
	}
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*repository.OperationLogQuery")).Return(logs, int64(1), nil)

	r := gin.New()
	r.GET("/operation-logs", handler.ListLogs)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/operation-logs?page=1&page_size=10&user_id=user-001&action=login&start_time=1000000&end_time=2000000", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationLogHandler_ListLogs_WithUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockOperationLogRepoForHandler)
	svc := service.NewOperationLogService(mockRepo)
	handler := NewOperationLogHandler(svc)

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*repository.OperationLogQuery")).Return([]*entity.OperationLog{}, int64(0), nil)

	r := gin.New()
	r.GET("/operation-logs", handler.ListLogs)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/operation-logs?username=admin", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationLogHandler_DeleteOldLogs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockOperationLogRepoForHandler)
	svc := service.NewOperationLogService(mockRepo)
	handler := NewOperationLogHandler(svc)

	mockRepo.On("DeleteBefore", mock.Anything, mock.AnythingOfType("int64")).Return(int64(5), nil)

	r := gin.New()
	r.DELETE("/operation-logs", handler.DeleteOldLogs)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/operation-logs?days=30", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationLogHandler_DeleteOldLogs_DefaultDays(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockOperationLogRepoForHandler)
	svc := service.NewOperationLogService(mockRepo)
	handler := NewOperationLogHandler(svc)

	mockRepo.On("DeleteBefore", mock.Anything, mock.AnythingOfType("int64")).Return(int64(0), nil)

	r := gin.New()
	r.DELETE("/operation-logs", handler.DeleteOldLogs)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/operation-logs", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationLogHandler_DeleteOldLogs_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockOperationLogRepoForHandler)
	svc := service.NewOperationLogService(mockRepo)
	handler := NewOperationLogHandler(svc)

	mockRepo.On("DeleteBefore", mock.Anything, mock.AnythingOfType("int64")).Return(int64(0), assert.AnError)

	r := gin.New()
	r.DELETE("/operation-logs", handler.DeleteOldLogs)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/operation-logs?days=30", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestOperationLogHandler_NewHandler(t *testing.T) {
	mockRepo := new(mockOperationLogRepoForHandler)
	svc := service.NewOperationLogService(mockRepo)
	handler := NewOperationLogHandler(svc)
	assert.NotNil(t, handler)
}
