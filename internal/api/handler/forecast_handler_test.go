package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/new-energy-monitoring/pkg/ai/forecast"
)

type mockForecastService struct {
	powerForecastFunc     func(ctx context.Context, stationID string, forecastType entity.ForecastType) ([]*entity.ForecastResult, error)
	getResultsFunc        func(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, page, pageSize int) ([]*entity.ForecastResult, int64, error)
	getAccuracyFunc       func(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time) (*repository.ForecastAccuracyStats, error)
	evaluateModelFunc     func(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time, installedCapacity float64) (*forecast.EvaluationReport, error)
	attributionAnalysisFunc func(ctx context.Context, stationID string, targetTime string, predictedPower, actualPower float64) (*forecast.AttributionResult, error)
}

func (m *mockForecastService) PowerForecast(ctx context.Context, stationID string, forecastType entity.ForecastType) ([]*entity.ForecastResult, error) {
	if m.powerForecastFunc != nil {
		return m.powerForecastFunc(ctx, stationID, forecastType)
	}
	confidence := 0.9
	return []*entity.ForecastResult{
		{
			ID:             "fc-001",
			StationID:      stationID,
			ForecastType:   forecastType,
			TargetTime:     time.Now().Add(1 * time.Hour),
			PredictedPower: 500.0,
			Confidence:     &confidence,
			ModelVersion:   "1.0.0",
		},
	}, nil
}

func (m *mockForecastService) GetResults(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, page, pageSize int) ([]*entity.ForecastResult, int64, error) {
	if m.getResultsFunc != nil {
		return m.getResultsFunc(ctx, stationID, forecastType, startTime, endTime, page, pageSize)
	}
	confidence := 0.9
	return []*entity.ForecastResult{
		{
			ID:             "fc-001",
			StationID:      stationID,
			ForecastType:   entity.ForecastTypeShortTerm,
			TargetTime:     time.Now(),
			PredictedPower: 800.0,
			Confidence:     &confidence,
			ModelVersion:   "1.0.0",
		},
	}, 1, nil
}

func (m *mockForecastService) GetAccuracy(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time) (*repository.ForecastAccuracyStats, error) {
	if m.getAccuracyFunc != nil {
		return m.getAccuracyFunc(ctx, stationID, forecastType, start, end)
	}
	return &repository.ForecastAccuracyStats{
		StationID:    stationID,
		ForecastType: "short_term",
		TotalPoints:  100,
		AvgAccuracy:  0.92,
		RMSE:         15.5,
		MAE:          12.3,
	}, nil
}

func (m *mockForecastService) EvaluateModel(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time, installedCapacity float64) (*forecast.EvaluationReport, error) {
	if m.evaluateModelFunc != nil {
		return m.evaluateModelFunc(ctx, stationID, forecastType, start, end, installedCapacity)
	}
	return &forecast.EvaluationReport{
		TotalPoints:  100,
		RMSE:         15.5,
		MAE:          12.3,
		Bias:         2.1,
		Accuracy:     0.92,
		InstalledCap: installedCapacity,
		ByPeriod:     map[string]forecast.Stats{},
	}, nil
}

func (m *mockForecastService) AttributionAnalysis(ctx context.Context, stationID string, targetTime string, predictedPower, actualPower float64) (*forecast.AttributionResult, error) {
	if m.attributionAnalysisFunc != nil {
		return m.attributionAnalysisFunc(ctx, stationID, targetTime, predictedPower, actualPower)
	}
	return &forecast.AttributionResult{
		StationID:    stationID,
		PrimaryCause: "weather",
		Confidence:   0.8,
		Factors: []forecast.AttributionFactor{
			{Name: "cloud_cover", Contribution: 0.5, Description: "云量变化"},
		},
		Suggestion: "偏差在正常范围内，持续监控",
		Deviation:   5.0,
	}, nil
}

func setupForecastRouter(h *ForecastHandler) *gin.Engine {
	r := gin.New()
	r.POST("/api/v1/forecast/power", h.PowerForecast)
	r.GET("/api/v1/forecast/results", h.GetResults)
	r.GET("/api/v1/forecast/accuracy", h.GetAccuracy)
	r.POST("/api/v1/forecast/evaluate", h.EvaluateModel)
	r.POST("/api/v1/forecast/attribution", h.AttributionAnalysis)
	return r
}

func TestForecastHandler_PowerForecast_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockForecastService{}
	handler := NewForecastHandler(mockSvc)
	router := setupForecastRouter(handler)

	body := `{"station_id":"station-001","forecast_type":"short_term"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forecast/power", strings.NewReader(body))
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
	if resp.Message != "success" {
		t.Fatalf("expected message 'success', got '%s'", resp.Message)
	}
}

func TestForecastHandler_PowerForecast_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockForecastService{}
	handler := NewForecastHandler(mockSvc)
	router := setupForecastRouter(handler)

	body := `{"forecast_type":"short_term"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forecast/power", strings.NewReader(body))
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

func TestForecastHandler_GetResults_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockForecastService{}
	handler := NewForecastHandler(mockSvc)
	router := setupForecastRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/forecast/results?station_id=station-001", nil)
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

func TestForecastHandler_GetResults_BadRequest_MissingStationID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockForecastService{}
	handler := NewForecastHandler(mockSvc)
	router := setupForecastRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/forecast/results", nil)
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

func TestForecastHandler_GetAccuracy_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockForecastService{}
	handler := NewForecastHandler(mockSvc)
	router := setupForecastRouter(handler)

	now := time.Now()
	url := "/api/v1/forecast/accuracy?station_id=station-001&start_time=" + now.Add(-24*time.Hour).Format(time.RFC3339) + "&end_time=" + now.Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, url, nil)
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

func TestForecastHandler_GetAccuracy_BadRequest_MissingTimes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockForecastService{}
	handler := NewForecastHandler(mockSvc)
	router := setupForecastRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/forecast/accuracy?station_id=station-001", nil)
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

func TestForecastHandler_EvaluateModel_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockForecastService{}
	handler := NewForecastHandler(mockSvc)
	router := setupForecastRouter(handler)

	now := time.Now()
	body := `{"station_id":"station-001","forecast_type":"short_term","start_time":"` + now.Add(-24*time.Hour).Format(time.RFC3339) + `","end_time":"` + now.Format(time.RFC3339) + `","installed_capacity":1000}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forecast/evaluate", strings.NewReader(body))
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

func TestForecastHandler_EvaluateModel_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockForecastService{}
	handler := NewForecastHandler(mockSvc)
	router := setupForecastRouter(handler)

	body := `{"station_id":"station-001"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forecast/evaluate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestForecastHandler_AttributionAnalysis_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockForecastService{}
	handler := NewForecastHandler(mockSvc)
	router := setupForecastRouter(handler)

	body := `{"station_id":"station-001","target_time":"2025-01-01T00:00:00Z","predicted_power":500.0,"actual_power":480.0}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forecast/attribution", strings.NewReader(body))
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

func TestForecastHandler_AttributionAnalysis_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockForecastService{}
	handler := NewForecastHandler(mockSvc)
	router := setupForecastRouter(handler)

	body := `{"station_id":"station-001"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forecast/attribution", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
