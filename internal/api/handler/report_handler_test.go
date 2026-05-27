package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockReportRepoForHandler struct {
	mock.Mock
}

func (m *mockReportRepoForHandler) GetStationPowerStats(ctx context.Context, stationID string, startTime, endTime time.Time) (*repository.StationPowerStats, error) {
	args := m.Called(ctx, stationID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.StationPowerStats), args.Error(1)
}

func (m *mockReportRepoForHandler) GetStationAlarmStats(ctx context.Context, stationID string, startTime, endTime time.Time) (*repository.StationAlarmStats, error) {
	args := m.Called(ctx, stationID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.StationAlarmStats), args.Error(1)
}

func (m *mockReportRepoForHandler) GetStationOnlineStats(ctx context.Context, stationID string, startTime, endTime time.Time) (*repository.StationOnlineStats, error) {
	args := m.Called(ctx, stationID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.StationOnlineStats), args.Error(1)
}

func (m *mockReportRepoForHandler) GetAllStationPowerStats(ctx context.Context, startTime, endTime time.Time) ([]*repository.StationPowerStats, error) {
	args := m.Called(ctx, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.StationPowerStats), args.Error(1)
}

func (m *mockReportRepoForHandler) GetAllStationAlarmStats(ctx context.Context, startTime, endTime time.Time) ([]*repository.StationAlarmStats, error) {
	args := m.Called(ctx, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.StationAlarmStats), args.Error(1)
}

func (m *mockReportRepoForHandler) GetAllStationOnlineStats(ctx context.Context, startTime, endTime time.Time) ([]*repository.StationOnlineStats, error) {
	args := m.Called(ctx, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.StationOnlineStats), args.Error(1)
}

func TestReportHandler_GenerateReport_MissingType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockReportRepoForHandler)
	svc := service.NewReportService(mockRepo)
	handler := NewReportHandler(svc)

	r := gin.New()
	r.GET("/reports/generate", handler.GenerateReport)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reports/generate?start_time=2024-01-01&end_time=2024-01-31", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReportHandler_GenerateReport_InvalidStartTime(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockReportRepoForHandler)
	svc := service.NewReportService(mockRepo)
	handler := NewReportHandler(svc)

	r := gin.New()
	r.GET("/reports/generate", handler.GenerateReport)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reports/generate?type=daily&start_time=invalid&end_time=2024-01-31", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReportHandler_GenerateReport_InvalidEndTime(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockReportRepoForHandler)
	svc := service.NewReportService(mockRepo)
	handler := NewReportHandler(svc)

	r := gin.New()
	r.GET("/reports/generate", handler.GenerateReport)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reports/generate?type=daily&start_time=2024-01-01&end_time=invalid", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReportHandler_GenerateReport_InvalidType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockReportRepoForHandler)
	svc := service.NewReportService(mockRepo)
	handler := NewReportHandler(svc)

	r := gin.New()
	r.GET("/reports/generate", handler.GenerateReport)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reports/generate?type=invalid&start_time=2024-01-01&end_time=2024-01-31", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReportHandler_GenerateReport_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockReportRepoForHandler)
	svc := service.NewReportService(mockRepo)
	handler := NewReportHandler(svc)

	powerStats := &repository.StationPowerStats{StationID: "s1", StationName: "Station 1", TotalPower: 1000.0}
	alarmStats := &repository.StationAlarmStats{StationID: "s1", AlarmCount: 5}
	onlineStats := &repository.StationOnlineStats{StationID: "s1", OnlineRate: 99.0}

	mockRepo.On("GetStationPowerStats", mock.Anything, "s1", mock.Anything, mock.Anything).Return(powerStats, nil)
	mockRepo.On("GetStationAlarmStats", mock.Anything, "s1", mock.Anything, mock.Anything).Return(alarmStats, nil)
	mockRepo.On("GetStationOnlineStats", mock.Anything, "s1", mock.Anything, mock.Anything).Return(onlineStats, nil)

	r := gin.New()
	r.GET("/reports/generate", handler.GenerateReport)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reports/generate?type=daily&start_time=2024-01-01&end_time=2024-01-31&station_id=s1", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_ExportReport_MissingType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockReportRepoForHandler)
	svc := service.NewReportService(mockRepo)
	handler := NewReportHandler(svc)

	r := gin.New()
	r.GET("/reports/export", handler.ExportReport)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reports/export?start_time=2024-01-01&end_time=2024-01-31&format=excel", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReportHandler_ExportReport_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockReportRepoForHandler)
	svc := service.NewReportService(mockRepo)
	handler := NewReportHandler(svc)

	r := gin.New()
	r.GET("/reports/export", handler.ExportReport)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reports/export?type=daily&start_time=2024-01-01&end_time=2024-01-31&format=invalid", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReportHandler_GetReportTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockReportRepoForHandler)
	svc := service.NewReportService(mockRepo)
	handler := NewReportHandler(svc)

	r := gin.New()
	r.GET("/reports/types", handler.GetReportTypes)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reports/types", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_NewHandler(t *testing.T) {
	mockRepo := new(mockReportRepoForHandler)
	svc := service.NewReportService(mockRepo)
	handler := NewReportHandler(svc)
	assert.NotNil(t, handler)
}
