package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

type mockFaultService struct {
	detectFaultsFunc             func(ctx context.Context, deviceID string) ([]*entity.FaultDetectionResult, error)
	getDetectionsFunc            func(ctx context.Context, deviceID string, severity *entity.FaultSeverity, status *entity.FaultDetectionStatus, page, pageSize int) ([]*entity.FaultDetectionResult, int64, error)
	getDetectionByIDFunc         func(ctx context.Context, id string) (*entity.FaultDetectionResult, error)
	updateDetectionStatusFunc    func(ctx context.Context, id string, status entity.FaultDetectionStatus) error
	getDeviceHealthFunc          func(ctx context.Context, deviceID string) (int, error)
	analyzeRootCauseFunc         func(ctx context.Context, detectionID string) (*entity.FaultDetectionResult, error)
	createWorkOrderFromDetectionFunc func(ctx context.Context, detectionID string) (string, error)
}

func (m *mockFaultService) DetectFaults(ctx context.Context, deviceID string) ([]*entity.FaultDetectionResult, error) {
	if m.detectFaultsFunc != nil {
		return m.detectFaultsFunc(ctx, deviceID)
	}
	return []*entity.FaultDetectionResult{
		{
			ID:          "fd-001",
			DeviceID:    deviceID,
			FaultType:   "temperature",
			Severity:    entity.FaultSeverityWarning,
			Confidence:  0.85,
			Description: "检测到异常: temperature",
			Status:      entity.FaultDetectionStatusPending,
			ModelVersion: "1.0.0",
		},
	}, nil
}

func (m *mockFaultService) GetDetections(ctx context.Context, deviceID string, severity *entity.FaultSeverity, status *entity.FaultDetectionStatus, page, pageSize int) ([]*entity.FaultDetectionResult, int64, error) {
	if m.getDetectionsFunc != nil {
		return m.getDetectionsFunc(ctx, deviceID, severity, status, page, pageSize)
	}
	return []*entity.FaultDetectionResult{
		{
			ID:          "fd-001",
			DeviceID:    deviceID,
			FaultType:   "temperature",
			Severity:    entity.FaultSeverityWarning,
			Confidence:  0.85,
			Description: "检测到异常: temperature",
			Status:      entity.FaultDetectionStatusPending,
			ModelVersion: "1.0.0",
		},
	}, 1, nil
}

func (m *mockFaultService) GetDetectionByID(ctx context.Context, id string) (*entity.FaultDetectionResult, error) {
	if m.getDetectionByIDFunc != nil {
		return m.getDetectionByIDFunc(ctx, id)
	}
	if id == "not-found" {
		return nil, fmt.Errorf("not found")
	}
	return &entity.FaultDetectionResult{
		ID:          id,
		DeviceID:    "device-001",
		FaultType:   "temperature",
		Severity:    entity.FaultSeverityWarning,
		Confidence:  0.85,
		Description: "检测到异常: temperature",
		Status:      entity.FaultDetectionStatusPending,
		ModelVersion: "1.0.0",
	}, nil
}

func (m *mockFaultService) UpdateDetectionStatus(ctx context.Context, id string, status entity.FaultDetectionStatus) error {
	if m.updateDetectionStatusFunc != nil {
		return m.updateDetectionStatusFunc(ctx, id, status)
	}
	return nil
}

func (m *mockFaultService) GetDeviceHealth(ctx context.Context, deviceID string) (int, error) {
	if m.getDeviceHealthFunc != nil {
		return m.getDeviceHealthFunc(ctx, deviceID)
	}
	return 95, nil
}

func (m *mockFaultService) AnalyzeRootCause(ctx context.Context, detectionID string) (*entity.FaultDetectionResult, error) {
	if m.analyzeRootCauseFunc != nil {
		return m.analyzeRootCauseFunc(ctx, detectionID)
	}
	rootCause := "AI分析: temperature 可能由设备老化或环境因素导致"
	return &entity.FaultDetectionResult{
		ID:          detectionID,
		DeviceID:    "device-001",
		FaultType:   "temperature",
		Severity:    entity.FaultSeverityWarning,
		Confidence:  0.85,
		Description: "检测到异常: temperature",
		RootCause:   &rootCause,
		Status:      entity.FaultDetectionStatusPending,
		ModelVersion: "1.0.0",
	}, nil
}

func (m *mockFaultService) CreateWorkOrderFromDetection(ctx context.Context, detectionID string) (string, error) {
	if m.createWorkOrderFromDetectionFunc != nil {
		return m.createWorkOrderFromDetectionFunc(ctx, detectionID)
	}
	return "wo-001", nil
}

func setupFaultRouter(h *FaultHandler) *gin.Engine {
	r := gin.New()
	r.POST("/api/v1/fault/detect", h.DetectFaults)
	r.GET("/api/v1/fault/detections", h.GetDetections)
	r.GET("/api/v1/fault/detections/:id", h.GetDetectionByID)
	r.PUT("/api/v1/fault/detections/:id/status", h.UpdateDetectionStatus)
	r.GET("/api/v1/fault/devices/:device_id/health", h.GetDeviceHealth)
	r.POST("/api/v1/fault/root-cause", h.AnalyzeRootCause)
	r.POST("/api/v1/fault/work-order", h.CreateWorkOrderFromDetection)
	return r
}

func TestFaultHandler_DetectFaults_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	body := `{"device_id":"device-001"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fault/detect", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp dto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
}

func TestFaultHandler_DetectFaults_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fault/detect", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != 400 {
		t.Fatalf("expected error code 400, got %d", resp.Code)
	}
}

func TestFaultHandler_GetDetections_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fault/detections?device_id=device-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp dto.PagedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
	if resp.Total != 1 {
		t.Fatalf("expected total 1, got %d", resp.Total)
	}
}

func TestFaultHandler_GetDetections_BadRequest_MissingDeviceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fault/detections", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != 400 {
		t.Fatalf("expected error code 400, got %d", resp.Code)
	}
}

func TestFaultHandler_GetDetectionByID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fault/detections/fd-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp dto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
}

func TestFaultHandler_GetDetectionByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fault/detections/not-found", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestFaultHandler_UpdateDetectionStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	body := `{"status":2}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/fault/detections/fd-001/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestFaultHandler_UpdateDetectionStatus_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	body := `{}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/fault/detections/fd-001/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestFaultHandler_GetDeviceHealth_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fault/devices/device-001/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp dto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
}

func TestFaultHandler_AnalyzeRootCause_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	body := `{"detection_id":"fd-001"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fault/root-cause", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp dto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
}

func TestFaultHandler_AnalyzeRootCause_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fault/root-cause", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestFaultHandler_CreateWorkOrderFromDetection_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	body := `{"detection_id":"fd-001"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fault/work-order", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp dto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
}

func TestFaultHandler_CreateWorkOrderFromDetection_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fault/work-order", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestFaultHandler_DetectFaults_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockFaultService{
		detectFaultsFunc: func(ctx context.Context, deviceID string) ([]*entity.FaultDetectionResult, error) {
			return nil, fmt.Errorf("internal error")
		},
	}
	handler := NewFaultHandler(mockSvc)
	router := setupFaultRouter(handler)

	body := `{"device_id":"device-001"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fault/detect", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func init() {
	_ = time.Now
}
