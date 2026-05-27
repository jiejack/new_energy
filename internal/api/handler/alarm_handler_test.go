package handler

import (
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

type mockAlarmRepoForHandler struct {
	mock.Mock
}

func (m *mockAlarmRepoForHandler) Create(ctx context.Context, alarm *entity.Alarm) error {
	args := m.Called(ctx, alarm)
	return args.Error(0)
}

func (m *mockAlarmRepoForHandler) Update(ctx context.Context, alarm *entity.Alarm) error {
	args := m.Called(ctx, alarm)
	return args.Error(0)
}

func (m *mockAlarmRepoForHandler) GetByID(ctx context.Context, id string) (*entity.Alarm, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Alarm), args.Error(1)
}

func (m *mockAlarmRepoForHandler) GetActiveAlarms(ctx context.Context, stationID *string, level *entity.AlarmLevel) ([]*entity.Alarm, error) {
	args := m.Called(ctx, stationID, level)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Alarm), args.Error(1)
}

func (m *mockAlarmRepoForHandler) GetHistoryAlarms(ctx context.Context, stationID *string, startTime, endTime int64) ([]*entity.Alarm, error) {
	args := m.Called(ctx, stationID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Alarm), args.Error(1)
}

func (m *mockAlarmRepoForHandler) Acknowledge(ctx context.Context, id, by string) error {
	args := m.Called(ctx, id, by)
	return args.Error(0)
}

func (m *mockAlarmRepoForHandler) Clear(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAlarmRepoForHandler) CountByLevel(ctx context.Context, stationID *string) (map[entity.AlarmLevel]int64, error) {
	args := m.Called(ctx, stationID)
	return args.Get(0).(map[entity.AlarmLevel]int64), args.Error(1)
}

func setupAlarmHandler(mockRepo *mockAlarmRepoForHandler) (*AlarmHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewAlarmService(mockRepo)
	handler := NewAlarmHandler(svc)
	r := gin.New()
	return handler, r
}

func TestAlarmHandler_GetAlarm_Success(t *testing.T) {
	mockRepo := new(mockAlarmRepoForHandler)
	handler, r := setupAlarmHandler(mockRepo)

	r.GET("/alarms/:id", handler.GetAlarm)

	alarm := entity.NewAlarm("p1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test Alarm", "msg")
	mockRepo.On("GetByID", mock.Anything, "alarm-001").Return(alarm, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/alarms/alarm-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAlarmHandler_GetAlarm_NotFound(t *testing.T) {
	mockRepo := new(mockAlarmRepoForHandler)
	handler, r := setupAlarmHandler(mockRepo)

	r.GET("/alarms/:id", handler.GetAlarm)

	mockRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/alarms/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAlarmHandler_ListAlarms_Success(t *testing.T) {
	mockRepo := new(mockAlarmRepoForHandler)
	handler, r := setupAlarmHandler(mockRepo)

	r.GET("/alarms", handler.ListAlarms)

	alarms := []*entity.Alarm{
		entity.NewAlarm("p1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Alarm 1", "msg1"),
	}
	mockRepo.On("GetActiveAlarms", mock.Anything, (*string)(nil), (*entity.AlarmLevel)(nil)).Return(alarms, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/alarms", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlarmHandler_ListAlarms_Error(t *testing.T) {
	mockRepo := new(mockAlarmRepoForHandler)
	handler, r := setupAlarmHandler(mockRepo)

	r.GET("/alarms", handler.ListAlarms)

	mockRepo.On("GetActiveAlarms", mock.Anything, (*string)(nil), (*entity.AlarmLevel)(nil)).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/alarms", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAlarmHandler_AcknowledgeAlarm_Success(t *testing.T) {
	mockRepo := new(mockAlarmRepoForHandler)
	handler, r := setupAlarmHandler(mockRepo)

	r.PUT("/alarms/:id/ack", handler.AcknowledgeAlarm)

	alarm := entity.NewAlarm("p1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Alarm", "msg")
	mockRepo.On("GetByID", mock.Anything, "alarm-001").Return(alarm, nil)
	mockRepo.On("Acknowledge", mock.Anything, "alarm-001", "").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/alarms/alarm-001/ack", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlarmHandler_AcknowledgeAlarm_Error(t *testing.T) {
	mockRepo := new(mockAlarmRepoForHandler)
	handler, r := setupAlarmHandler(mockRepo)

	r.PUT("/alarms/:id/ack", handler.AcknowledgeAlarm)

	alarm := entity.NewAlarm("p1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Alarm", "msg")
	mockRepo.On("GetByID", mock.Anything, "alarm-001").Return(alarm, nil)
	mockRepo.On("Acknowledge", mock.Anything, "alarm-001", "").Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/alarms/alarm-001/ack", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlarmHandler_ClearAlarm_Success(t *testing.T) {
	mockRepo := new(mockAlarmRepoForHandler)
	handler, r := setupAlarmHandler(mockRepo)

	r.PUT("/alarms/:id/clear", handler.ClearAlarm)

	alarm := entity.NewAlarm("p1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Alarm", "msg")
	alarm.Acknowledge("admin")
	mockRepo.On("GetByID", mock.Anything, "alarm-001").Return(alarm, nil)
	mockRepo.On("Clear", mock.Anything, "alarm-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/alarms/alarm-001/clear", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlarmHandler_ClearAlarm_Error(t *testing.T) {
	mockRepo := new(mockAlarmRepoForHandler)
	handler, r := setupAlarmHandler(mockRepo)

	r.PUT("/alarms/:id/clear", handler.ClearAlarm)

	alarm := entity.NewAlarm("p1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Alarm", "msg")
	alarm.Acknowledge("admin")
	mockRepo.On("GetByID", mock.Anything, "alarm-001").Return(alarm, nil)
	mockRepo.On("Clear", mock.Anything, "alarm-001").Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/alarms/alarm-001/clear", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlarmHandler_NewAlarmHandler(t *testing.T) {
	mockRepo := new(mockAlarmRepoForHandler)
	svc := service.NewAlarmService(mockRepo)
	handler := NewAlarmHandler(svc)
	assert.NotNil(t, handler)
}

func TestAlarmHandler_AcknowledgeAlarm_NotFound(t *testing.T) {
	mockRepo := new(mockAlarmRepoForHandler)
	handler, r := setupAlarmHandler(mockRepo)

	r.PUT("/alarms/:id/ack", handler.AcknowledgeAlarm)

	mockRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/alarms/nonexistent/ack", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlarmHandler_ClearAlarm_NotFound(t *testing.T) {
	mockRepo := new(mockAlarmRepoForHandler)
	handler, r := setupAlarmHandler(mockRepo)

	r.PUT("/alarms/:id/clear", handler.ClearAlarm)

	mockRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/alarms/nonexistent/clear", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlarmHandler_ListAlarms_Empty(t *testing.T) {
	mockRepo := new(mockAlarmRepoForHandler)
	handler, r := setupAlarmHandler(mockRepo)

	r.GET("/alarms", handler.ListAlarms)

	mockRepo.On("GetActiveAlarms", mock.Anything, (*string)(nil), (*entity.AlarmLevel)(nil)).Return([]*entity.Alarm{}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/alarms", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}
