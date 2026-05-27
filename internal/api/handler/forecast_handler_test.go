package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/new-energy-monitoring/pkg/ai/forecast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockForecastService struct {
	mock.Mock
}

func (m *mockForecastService) PowerForecast(ctx context.Context, stationID string, forecastType entity.ForecastType) ([]*entity.ForecastResult, error) {
	args := m.Called(ctx, stationID, forecastType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ForecastResult), args.Error(1)
}
func (m *mockForecastService) GetResults(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, page, pageSize int) ([]*entity.ForecastResult, int64, error) {
	args := m.Called(ctx, stationID, forecastType, startTime, endTime, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.ForecastResult), args.Get(1).(int64), args.Error(2)
}
func (m *mockForecastService) GetAccuracy(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time) (*repository.ForecastAccuracyStats, error) {
	args := m.Called(ctx, stationID, forecastType, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ForecastAccuracyStats), args.Error(1)
}
func (m *mockForecastService) EvaluateModel(ctx context.Context, stationID string, forecastType *entity.ForecastType, start, end time.Time, installedCapacity float64) (*forecast.EvaluationReport, error) {
	args := m.Called(ctx, stationID, forecastType, start, end, installedCapacity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*forecast.EvaluationReport), args.Error(1)
}
func (m *mockForecastService) AttributionAnalysis(ctx context.Context, stationID string, targetTime string, predictedPower, actualPower float64) (*forecast.AttributionResult, error) {
	args := m.Called(ctx, stationID, targetTime, predictedPower, actualPower)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*forecast.AttributionResult), args.Error(1)
}

func setupForecastHandler(svc *mockForecastService) (*ForecastHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	handler := NewForecastHandler(svc)
	r := gin.New()
	return handler, r
}

func TestForecastHandler_PowerForecast_Success(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.POST("/forecast/power", handler.PowerForecast)

	results := []*entity.ForecastResult{
		{StationID: "station-001", ForecastType: entity.ForecastTypeShortTerm, PredictedPower: 500.0},
	}
	mockSvc.On("PowerForecast", mock.Anything, "station-001", entity.ForecastTypeShortTerm).Return(results, nil)

	body := map[string]string{"station_id": "station-001", "forecast_type": "short_term"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/forecast/power", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestForecastHandler_PowerForecast_InvalidJSON(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.POST("/forecast/power", handler.PowerForecast)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/forecast/power", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestForecastHandler_PowerForecast_Error(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.POST("/forecast/power", handler.PowerForecast)

	mockSvc.On("PowerForecast", mock.Anything, "station-001", entity.ForecastTypeShortTerm).Return(nil, assert.AnError)

	body := map[string]string{"station_id": "station-001", "forecast_type": "short_term"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/forecast/power", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestForecastHandler_GetResults_Success(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.GET("/forecast/results", handler.GetResults)

	mockSvc.On("GetResults", mock.Anything, "station-001", (*entity.ForecastType)(nil), (*time.Time)(nil), (*time.Time)(nil), 1, 20).Return([]*entity.ForecastResult{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/forecast/results?station_id=station-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestForecastHandler_GetResults_MissingStationID(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.GET("/forecast/results", handler.GetResults)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/forecast/results", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestForecastHandler_GetResults_WithTimeRange(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.GET("/forecast/results", handler.GetResults)

	mockSvc.On("GetResults", mock.Anything, "station-001", (*entity.ForecastType)(nil), mock.Anything, mock.Anything, 1, 20).Return([]*entity.ForecastResult{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/forecast/results?station_id=station-001&start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestForecastHandler_GetAccuracy_Success(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.GET("/forecast/accuracy", handler.GetAccuracy)

	stats := &repository.ForecastAccuracyStats{StationID: "station-001", AvgAccuracy: 0.95}
	mockSvc.On("GetAccuracy", mock.Anything, "station-001", (*entity.ForecastType)(nil), mock.Anything, mock.Anything).Return(stats, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/forecast/accuracy?station_id=station-001&start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestForecastHandler_GetAccuracy_MissingStationID(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.GET("/forecast/accuracy", handler.GetAccuracy)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/forecast/accuracy?start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestForecastHandler_GetAccuracy_MissingTimes(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.GET("/forecast/accuracy", handler.GetAccuracy)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/forecast/accuracy?station_id=station-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestForecastHandler_GetAccuracy_InvalidStartTime(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.GET("/forecast/accuracy", handler.GetAccuracy)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/forecast/accuracy?station_id=station-001&start_time=invalid&end_time=2024-01-02T00:00:00Z", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestForecastHandler_GetAccuracy_InvalidEndTime(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.GET("/forecast/accuracy", handler.GetAccuracy)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/forecast/accuracy?station_id=station-001&start_time=2024-01-01T00:00:00Z&end_time=invalid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestForecastHandler_EvaluateModel_Success(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.POST("/forecast/evaluate", handler.EvaluateModel)

	report := &forecast.EvaluationReport{}
	mockSvc.On("EvaluateModel", mock.Anything, "station-001", (*entity.ForecastType)(nil), mock.Anything, mock.Anything, 100.0).Return(report, nil)

	body := map[string]interface{}{
		"station_id":         "station-001",
		"start_time":         "2024-01-01T00:00:00Z",
		"end_time":           "2024-01-02T00:00:00Z",
		"installed_capacity": 100.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/forecast/evaluate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestForecastHandler_EvaluateModel_InvalidJSON(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.POST("/forecast/evaluate", handler.EvaluateModel)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/forecast/evaluate", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestForecastHandler_EvaluateModel_InvalidStartTime(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.POST("/forecast/evaluate", handler.EvaluateModel)

	body := map[string]interface{}{
		"station_id":         "station-001",
		"start_time":         "invalid",
		"end_time":           "2024-01-02T00:00:00Z",
		"installed_capacity": 100.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/forecast/evaluate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestForecastHandler_AttributionAnalysis_Success(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.POST("/forecast/attribution", handler.AttributionAnalysis)

	result := &forecast.AttributionResult{}
	mockSvc.On("AttributionAnalysis", mock.Anything, "station-001", "2024-01-01T00:00:00Z", 500.0, 480.0).Return(result, nil)

	body := map[string]interface{}{
		"station_id":      "station-001",
		"target_time":     "2024-01-01T00:00:00Z",
		"predicted_power": 500.0,
		"actual_power":    480.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/forecast/attribution", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestForecastHandler_AttributionAnalysis_InvalidJSON(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler, r := setupForecastHandler(mockSvc)

	r.POST("/forecast/attribution", handler.AttributionAnalysis)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/forecast/attribution", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestForecastHandler_NewForecastHandler(t *testing.T) {
	mockSvc := new(mockForecastService)
	handler := NewForecastHandler(mockSvc)
	assert.NotNil(t, handler)
}
