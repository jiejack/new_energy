package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFaultService struct {
	mock.Mock
}

func (m *mockFaultService) DetectFaults(ctx context.Context, deviceID string) ([]*entity.FaultDetectionResult, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.FaultDetectionResult), args.Error(1)
}
func (m *mockFaultService) GetDetections(ctx context.Context, deviceID string, severity *entity.FaultSeverity, status *entity.FaultDetectionStatus, page, pageSize int) ([]*entity.FaultDetectionResult, int64, error) {
	args := m.Called(ctx, deviceID, severity, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.FaultDetectionResult), args.Get(1).(int64), args.Error(2)
}
func (m *mockFaultService) GetDetectionByID(ctx context.Context, id string) (*entity.FaultDetectionResult, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.FaultDetectionResult), args.Error(1)
}
func (m *mockFaultService) UpdateDetectionStatus(ctx context.Context, id string, status entity.FaultDetectionStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}
func (m *mockFaultService) GetDeviceHealth(ctx context.Context, deviceID string) (int, error) {
	args := m.Called(ctx, deviceID)
	return args.Int(0), args.Error(1)
}
func (m *mockFaultService) AnalyzeRootCause(ctx context.Context, detectionID string) (*entity.FaultDetectionResult, error) {
	args := m.Called(ctx, detectionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.FaultDetectionResult), args.Error(1)
}
func (m *mockFaultService) CreateWorkOrderFromDetection(ctx context.Context, detectionID string) (string, error) {
	args := m.Called(ctx, detectionID)
	return args.String(0), args.Error(1)
}

func setupFaultHandler(svc *mockFaultService) (*FaultHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	handler := NewFaultHandler(svc)
	r := gin.New()
	return handler, r
}

func TestFaultHandler_DetectFaults_Success(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.POST("/faults/detect", handler.DetectFaults)

	results := []*entity.FaultDetectionResult{
		entity.NewFaultDetectionResult("device-001", "temperature", entity.FaultSeverityWarning, 0.9, "test", "1.0"),
	}
	mockSvc.On("DetectFaults", mock.Anything, "device-001").Return(results, nil)

	body := map[string]string{"device_id": "device-001"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/faults/detect", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFaultHandler_DetectFaults_InvalidJSON(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.POST("/faults/detect", handler.DetectFaults)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/faults/detect", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFaultHandler_DetectFaults_Error(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.POST("/faults/detect", handler.DetectFaults)

	mockSvc.On("DetectFaults", mock.Anything, "device-001").Return(nil, assert.AnError)

	body := map[string]string{"device_id": "device-001"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/faults/detect", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFaultHandler_GetDetections_Success(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.GET("/faults/detections", handler.GetDetections)

	mockSvc.On("GetDetections", mock.Anything, "device-001", (*entity.FaultSeverity)(nil), (*entity.FaultDetectionStatus)(nil), 1, 20).Return([]*entity.FaultDetectionResult{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/faults/detections?device_id=device-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFaultHandler_GetDetections_MissingDeviceID(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.GET("/faults/detections", handler.GetDetections)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/faults/detections", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFaultHandler_GetDetections_WithFilters(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.GET("/faults/detections", handler.GetDetections)

	severity := entity.FaultSeverityWarning
	status := entity.FaultDetectionStatusPending
	mockSvc.On("GetDetections", mock.Anything, "device-001", &severity, &status, 1, 20).Return([]*entity.FaultDetectionResult{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/faults/detections?device_id=device-001&severity=warning&status=1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFaultHandler_GetDetections_Error(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.GET("/faults/detections", handler.GetDetections)

	mockSvc.On("GetDetections", mock.Anything, "device-001", (*entity.FaultSeverity)(nil), (*entity.FaultDetectionStatus)(nil), 1, 20).Return(nil, int64(0), assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/faults/detections?device_id=device-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFaultHandler_GetDetectionByID_Success(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.GET("/faults/detections/:id", handler.GetDetectionByID)

	result := entity.NewFaultDetectionResult("device-001", "temp", entity.FaultSeverityWarning, 0.9, "test", "1.0")
	mockSvc.On("GetDetectionByID", mock.Anything, "det-001").Return(result, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/faults/detections/det-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFaultHandler_GetDetectionByID_NotFound(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.GET("/faults/detections/:id", handler.GetDetectionByID)

	mockSvc.On("GetDetectionByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/faults/detections/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFaultHandler_UpdateDetectionStatus_Success(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.PUT("/faults/detections/:id/status", handler.UpdateDetectionStatus)

	mockSvc.On("UpdateDetectionStatus", mock.Anything, "det-001", entity.FaultDetectionStatus(2)).Return(nil)

	body := map[string]int{"status": 2}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/faults/detections/det-001/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFaultHandler_UpdateDetectionStatus_InvalidJSON(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.PUT("/faults/detections/:id/status", handler.UpdateDetectionStatus)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/faults/detections/det-001/status", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFaultHandler_GetDeviceHealth_Success(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.GET("/faults/health/:device_id", handler.GetDeviceHealth)

	mockSvc.On("GetDeviceHealth", mock.Anything, "device-001").Return(85, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/faults/health/device-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFaultHandler_GetDeviceHealth_Error(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.GET("/faults/health/:device_id", handler.GetDeviceHealth)

	mockSvc.On("GetDeviceHealth", mock.Anything, "device-001").Return(0, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/faults/health/device-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFaultHandler_AnalyzeRootCause_Success(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.POST("/faults/analyze", handler.AnalyzeRootCause)

	result := entity.NewFaultDetectionResult("device-001", "temp", entity.FaultSeverityWarning, 0.9, "test", "1.0")
	mockSvc.On("AnalyzeRootCause", mock.Anything, "det-001").Return(result, nil)

	body := map[string]string{"detection_id": "det-001"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/faults/analyze", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFaultHandler_AnalyzeRootCause_InvalidJSON(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.POST("/faults/analyze", handler.AnalyzeRootCause)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/faults/analyze", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFaultHandler_CreateWorkOrderFromDetection_Success(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.POST("/faults/work-order", handler.CreateWorkOrderFromDetection)

	mockSvc.On("CreateWorkOrderFromDetection", mock.Anything, "det-001").Return("wo-001", nil)

	body := map[string]string{"detection_id": "det-001"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/faults/work-order", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFaultHandler_CreateWorkOrderFromDetection_InvalidJSON(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.POST("/faults/work-order", handler.CreateWorkOrderFromDetection)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/faults/work-order", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFaultHandler_CreateWorkOrderFromDetection_Error(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler, r := setupFaultHandler(mockSvc)

	r.POST("/faults/work-order", handler.CreateWorkOrderFromDetection)

	mockSvc.On("CreateWorkOrderFromDetection", mock.Anything, "det-001").Return("", assert.AnError)

	body := map[string]string{"detection_id": "det-001"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/faults/work-order", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFaultHandler_NewFaultHandler(t *testing.T) {
	mockSvc := new(mockFaultService)
	handler := NewFaultHandler(mockSvc)
	assert.NotNil(t, handler)
}


