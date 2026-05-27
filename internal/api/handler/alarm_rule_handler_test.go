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

type mockAlarmRuleRepoForHandler struct {
	mock.Mock
}

func (m *mockAlarmRuleRepoForHandler) Create(ctx context.Context, rule *entity.AlarmRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *mockAlarmRuleRepoForHandler) Update(ctx context.Context, rule *entity.AlarmRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *mockAlarmRuleRepoForHandler) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAlarmRuleRepoForHandler) GetByID(ctx context.Context, id string) (*entity.AlarmRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AlarmRule), args.Error(1)
}

func (m *mockAlarmRuleRepoForHandler) GetByName(ctx context.Context, name string) (*entity.AlarmRule, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AlarmRule), args.Error(1)
}

func (m *mockAlarmRuleRepoForHandler) List(ctx context.Context, query *repository.AlarmRuleQuery) ([]*entity.AlarmRule, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Get(1).(int64), args.Error(2)
}

func (m *mockAlarmRuleRepoForHandler) GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func (m *mockAlarmRuleRepoForHandler) GetRulesByPointID(ctx context.Context, pointID string) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx, pointID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func (m *mockAlarmRuleRepoForHandler) GetRulesByDeviceID(ctx context.Context, deviceID string) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func (m *mockAlarmRuleRepoForHandler) GetRulesByStationID(ctx context.Context, stationID string) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func TestAlarmRuleHandler_CreateAlarmRule_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)

	r := gin.New()
	r.POST("/alarm-rules", func(c *gin.Context) {
		c.Set("user_id", "admin")
		handler.CreateAlarmRule(c)
	})

	mockRepo.On("GetByName", mock.Anything, "Test Rule").Return(nil, assert.AnError)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

	body := service.CreateAlarmRuleRequest{
		Name:      "Test Rule",
		Type:      entity.AlarmRuleTypeLimit,
		Level:     entity.AlarmLevelWarning,
		Condition: "value > threshold",
		Threshold: 85.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/alarm-rules", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAlarmRuleHandler_CreateAlarmRule_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)

	r := gin.New()
	r.POST("/alarm-rules", handler.CreateAlarmRule)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/alarm-rules", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlarmRuleHandler_CreateAlarmRule_DuplicateName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)

	r := gin.New()
	r.POST("/alarm-rules", func(c *gin.Context) {
		c.Set("user_id", "admin")
		handler.CreateAlarmRule(c)
	})

	existing := entity.NewAlarmRule("Test Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
	mockRepo.On("GetByName", mock.Anything, "Test Rule").Return(existing, nil)

	body := service.CreateAlarmRuleRequest{
		Name:      "Test Rule",
		Type:      entity.AlarmRuleTypeLimit,
		Level:     entity.AlarmLevelWarning,
		Condition: "value > threshold",
		Threshold: 85.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/alarm-rules", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAlarmRuleHandler_GetAlarmRule_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)

	r := gin.New()
	r.GET("/alarm-rules/:id", handler.GetAlarmRule)

	rule := entity.NewAlarmRule("Test Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
	mockRepo.On("GetByID", mock.Anything, "rule-001").Return(rule, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/alarm-rules/rule-001", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlarmRuleHandler_GetAlarmRule_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)

	r := gin.New()
	r.GET("/alarm-rules/:id", handler.GetAlarmRule)

	mockRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/alarm-rules/nonexistent", nil)

	r.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestAlarmRuleHandler_ListAlarmRules(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)

	r := gin.New()
	r.GET("/alarm-rules", handler.ListAlarmRules)

	rules := []*entity.AlarmRule{
		entity.NewAlarmRule("Rule 1", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "c1"),
	}
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*repository.AlarmRuleQuery")).Return(rules, int64(1), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/alarm-rules?page=1&page_size=10", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlarmRuleHandler_UpdateAlarmRule_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)

	r := gin.New()
	r.PUT("/alarm-rules/:id", func(c *gin.Context) {
		c.Set("user_id", "admin")
		handler.UpdateAlarmRule(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/alarm-rules/rule-001", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlarmRuleHandler_DeleteAlarmRule(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)

	r := gin.New()
	r.DELETE("/alarm-rules/:id", handler.DeleteAlarmRule)

	mockRepo.On("Delete", mock.Anything, "rule-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/alarm-rules/rule-001", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlarmRuleHandler_EnableAlarmRule(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)

	r := gin.New()
	r.PUT("/alarm-rules/:id/enable", func(c *gin.Context) {
		c.Set("user_id", "admin")
		handler.EnableAlarmRule(c)
	})

	rule := entity.NewAlarmRule("Test Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "c1")
	rule.Status = entity.AlarmRuleStatusDisabled
	mockRepo.On("GetByID", mock.Anything, "rule-001").Return(rule, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/alarm-rules/rule-001/enable", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlarmRuleHandler_DisableAlarmRule(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)

	r := gin.New()
	r.PUT("/alarm-rules/:id/disable", func(c *gin.Context) {
		c.Set("user_id", "admin")
		handler.DisableAlarmRule(c)
	})

	rule := entity.NewAlarmRule("Test Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "c1")
	mockRepo.On("GetByID", mock.Anything, "rule-001").Return(rule, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/alarm-rules/rule-001/disable", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlarmRuleHandler_GetRulesByPoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)

	r := gin.New()
	r.GET("/points/:point_id/rules", handler.GetRulesByPoint)

	rules := []*entity.AlarmRule{
		entity.NewAlarmRule("Point Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "c1"),
	}
	mockRepo.On("GetRulesByPointID", mock.Anything, "point-001").Return(rules, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/points/point-001/rules", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlarmRuleHandler_GetRulesByDevice(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)

	r := gin.New()
	r.GET("/devices/:device_id/rules", handler.GetRulesByDevice)

	mockRepo.On("GetRulesByDeviceID", mock.Anything, "device-001").Return([]*entity.AlarmRule{}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/devices/device-001/rules", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlarmRuleHandler_GetRulesByStation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)

	r := gin.New()
	r.GET("/stations/:station_id/rules", handler.GetRulesByStation)

	mockRepo.On("GetRulesByStationID", mock.Anything, "station-001").Return([]*entity.AlarmRule{}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stations/station-001/rules", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlarmRuleHandler_NewAlarmRuleHandler(t *testing.T) {
	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)
	assert.NotNil(t, handler)
}
