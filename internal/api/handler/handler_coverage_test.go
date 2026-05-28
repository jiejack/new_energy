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
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockExportSvc struct {
	mock.Mock
}

func (m *mockExportSvc) Export(ctx context.Context, req *service.ExportRequest) (*service.ExportResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ExportResult), args.Error(1)
}

var _ service.ExportServiceInterface = (*mockExportSvc)(nil)

type mockCostEntryRepoForCov struct {
	mock.Mock
}

func (m *mockCostEntryRepoForCov) Create(ctx context.Context, entry *entity.CostEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *mockCostEntryRepoForCov) Update(ctx context.Context, entry *entity.CostEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *mockCostEntryRepoForCov) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockCostEntryRepoForCov) GetByID(ctx context.Context, id string) (*entity.CostEntry, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostEntry), args.Error(1)
}

func (m *mockCostEntryRepoForCov) GetByCode(ctx context.Context, code string) (*entity.CostEntry, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostEntry), args.Error(1)
}

func (m *mockCostEntryRepoForCov) List(ctx context.Context, categoryID *string, startDate, endDate *time.Time, status *string, offset, limit int) ([]*entity.CostEntry, int64, error) {
	args := m.Called(ctx, categoryID, startDate, endDate, status, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.CostEntry), args.Get(1).(int64), args.Error(2)
}

func (m *mockCostEntryRepoForCov) GetTotalByCategory(ctx context.Context, categoryID string, startDate, endDate *time.Time) (float64, error) {
	args := m.Called(ctx, categoryID, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

func (m *mockCostEntryRepoForCov) GetTotalByPeriod(ctx context.Context, startDate, endDate *time.Time) (float64, error) {
	args := m.Called(ctx, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

type mockCostCategoryRepoForCov struct {
	mock.Mock
}

func (m *mockCostCategoryRepoForCov) Create(ctx context.Context, category *entity.CostCategory) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *mockCostCategoryRepoForCov) Update(ctx context.Context, category *entity.CostCategory) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *mockCostCategoryRepoForCov) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockCostCategoryRepoForCov) GetByID(ctx context.Context, id string) (*entity.CostCategory, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostCategory), args.Error(1)
}

func (m *mockCostCategoryRepoForCov) GetByCode(ctx context.Context, code string) (*entity.CostCategory, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostCategory), args.Error(1)
}

func (m *mockCostCategoryRepoForCov) List(ctx context.Context, parentID *string, status *string) ([]*entity.CostCategory, error) {
	args := m.Called(ctx, parentID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CostCategory), args.Error(1)
}

func (m *mockCostCategoryRepoForCov) GetTree(ctx context.Context) ([]*entity.CostCategory, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CostCategory), args.Error(1)
}

type mockCostAllocationRepoForCov struct {
	mock.Mock
}

func (m *mockCostAllocationRepoForCov) Create(ctx context.Context, allocation *entity.CostAllocation) error {
	args := m.Called(ctx, allocation)
	return args.Error(0)
}

func (m *mockCostAllocationRepoForCov) Update(ctx context.Context, allocation *entity.CostAllocation) error {
	args := m.Called(ctx, allocation)
	return args.Error(0)
}

func (m *mockCostAllocationRepoForCov) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockCostAllocationRepoForCov) GetByID(ctx context.Context, id string) (*entity.CostAllocation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostAllocation), args.Error(1)
}

func (m *mockCostAllocationRepoForCov) ListByCostEntryID(ctx context.Context, costEntryID string) ([]*entity.CostAllocation, error) {
	args := m.Called(ctx, costEntryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CostAllocation), args.Error(1)
}

func (m *mockCostAllocationRepoForCov) ListByAllocated(ctx context.Context, allocatedTo, allocatedID string) ([]*entity.CostAllocation, error) {
	args := m.Called(ctx, allocatedTo, allocatedID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CostAllocation), args.Error(1)
}

func (m *mockCostAllocationRepoForCov) GetTotalByAllocated(ctx context.Context, allocatedTo, allocatedID string, startDate, endDate *time.Time) (float64, error) {
	args := m.Called(ctx, allocatedTo, allocatedID, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

type mockCostReportRepoForCov struct {
	mock.Mock
}

func (m *mockCostReportRepoForCov) Create(ctx context.Context, report *entity.CostReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *mockCostReportRepoForCov) Update(ctx context.Context, report *entity.CostReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *mockCostReportRepoForCov) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockCostReportRepoForCov) GetByID(ctx context.Context, id string) (*entity.CostReport, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostReport), args.Error(1)
}

func (m *mockCostReportRepoForCov) GetByCode(ctx context.Context, code string) (*entity.CostReport, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostReport), args.Error(1)
}

func (m *mockCostReportRepoForCov) List(ctx context.Context, reportType *string, status *string, startDate, endDate *time.Time, offset, limit int) ([]*entity.CostReport, int64, error) {
	args := m.Called(ctx, reportType, status, startDate, endDate, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.CostReport), args.Get(1).(int64), args.Error(2)
}

func (m *mockCostReportRepoForCov) GetByPeriod(ctx context.Context, reportType string, periodStart, periodEnd time.Time) (*entity.CostReport, error) {
	args := m.Called(ctx, reportType, periodStart, periodEnd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CostReport), args.Error(1)
}

var _ repository.CostEntryRepository = (*mockCostEntryRepoForCov)(nil)
var _ repository.CostCategoryRepository = (*mockCostCategoryRepoForCov)(nil)
var _ repository.CostAllocationRepository = (*mockCostAllocationRepoForCov)(nil)
var _ repository.CostReportRepository = (*mockCostReportRepoForCov)(nil)

func TestExportHandler_Export_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockExportSvc)
	handler := NewExportHandler(mockSvc)

	r := gin.New()
	r.POST("/export", handler.Export)

	buf := bytes.NewBuffer([]byte("test data"))
	mockSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(&service.ExportResult{
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Filename:    "alarms.xlsx",
		Buffer:      buf,
	}, nil)

	body := service.ExportRequest{Type: service.ExportTypeAlarm, Format: service.ExportFormatExcel}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/export", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExportHandler_Export_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockExportSvc)
	handler := NewExportHandler(mockSvc)

	r := gin.New()
	r.POST("/export", handler.Export)

	mockSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, assert.AnError)

	body := service.ExportRequest{Type: service.ExportTypeAlarm, Format: service.ExportFormatExcel}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/export", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestExportHandler_ExportAlarms_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockExportSvc)
	handler := NewExportHandler(mockSvc)

	r := gin.New()
	r.GET("/export/alarms", handler.ExportAlarms)

	buf := bytes.NewBuffer([]byte("alarm data"))
	mockSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(&service.ExportResult{
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Filename:    "alarms.xlsx",
		Buffer:      buf,
	}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/export/alarms?format=excel&start_time=1709500800000&end_time=1709587200000&station_id=st-001&level=3", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExportHandler_ExportAlarms_DefaultFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockExportSvc)
	handler := NewExportHandler(mockSvc)

	r := gin.New()
	r.GET("/export/alarms", handler.ExportAlarms)

	buf := bytes.NewBuffer([]byte("alarm data"))
	mockSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(&service.ExportResult{
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Filename:    "alarms.xlsx",
		Buffer:      buf,
	}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/export/alarms", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExportHandler_ExportAlarms_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockExportSvc)
	handler := NewExportHandler(mockSvc)

	r := gin.New()
	r.GET("/export/alarms", handler.ExportAlarms)

	mockSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/export/alarms", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestExportHandler_ExportDevices_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockExportSvc)
	handler := NewExportHandler(mockSvc)

	r := gin.New()
	r.GET("/export/devices", handler.ExportDevices)

	buf := bytes.NewBuffer([]byte("device data"))
	mockSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(&service.ExportResult{
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Filename:    "devices.xlsx",
		Buffer:      buf,
	}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/export/devices?format=csv&station_id=st-001&type=inverter", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExportHandler_ExportDevices_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockExportSvc)
	handler := NewExportHandler(mockSvc)

	r := gin.New()
	r.GET("/export/devices", handler.ExportDevices)

	mockSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/export/devices", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestExportHandler_ExportStations_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockExportSvc)
	handler := NewExportHandler(mockSvc)

	r := gin.New()
	r.GET("/export/stations", handler.ExportStations)

	buf := bytes.NewBuffer([]byte("station data"))
	mockSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(&service.ExportResult{
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Filename:    "stations.xlsx",
		Buffer:      buf,
	}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/export/stations?format=excel&sub_region_id=reg-001&type=pv", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExportHandler_ExportStations_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockExportSvc)
	handler := NewExportHandler(mockSvc)

	r := gin.New()
	r.GET("/export/stations", handler.ExportStations)

	mockSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/export/stations", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEEHandler_GetEnergyEfficiencyTrend_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	r.GET("/ee/trend", handler.GetEnergyEfficiencyTrend)

	eeRepo.On("GetRecordsByTimeRange", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, mock.Anything, mock.Anything).Return([]*entity.EnergyEfficiencyRecord{}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/trend?target_id=target-001&type=device", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEEHandler_GetEnergyEfficiencyTrend_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	r.GET("/ee/trend", handler.GetEnergyEfficiencyTrend)

	eeRepo.On("GetRecordsByTimeRange", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/trend?target_id=target-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEEHandler_GetEnergyEfficiencyStatistics_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	r.GET("/ee/statistics", handler.GetEnergyEfficiencyStatistics)

	eeRepo.On("GetStatistics", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, "month", mock.Anything, mock.Anything).Return(&repository.EnergyEfficiencyStatistics{TotalRecords: 10}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/statistics?target_id=target-001&type=device&period=month", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEEHandler_GetEnergyEfficiencyStatistics_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	r.GET("/ee/statistics", handler.GetEnergyEfficiencyStatistics)

	eeRepo.On("GetStatistics", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, "month", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/statistics?target_id=target-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEEHandler_GetEnergyEfficiencyComparison_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	r.GET("/ee/comparison", handler.GetEnergyEfficiencyComparison)

	stats := &repository.EnergyEfficiencyStatistics{TotalRecords: 10, AvgEfficiency: 0.85}
	eeRepo.On("GetStatistics", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, "month", mock.Anything, mock.Anything).Return(stats, nil).Times(3)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/comparison?target_id=target-001&type=device&period=month", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEEHandler_GetEnergyEfficiencyComparison_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	r.GET("/ee/comparison", handler.GetEnergyEfficiencyComparison)

	eeRepo.On("GetStatistics", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, "month", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/comparison?target_id=target-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEEHandler_CreateEnergyEfficiencyAnalysis_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	r.POST("/ee/analyses", handler.CreateEnergyEfficiencyAnalysis)

	stats := &repository.EnergyEfficiencyStatistics{TotalRecords: 2, AvgEfficiency: 0.85, MaxEfficiency: 0.9, MinEfficiency: 0.8, TotalInputEnergy: 1000}
	eeRepo.On("GetStatistics", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, mock.Anything, mock.Anything, mock.Anything).Return(stats, nil)
	eeRepo.On("GetBenchmark", mock.Anything, entity.EnergyEfficiencyTypeDevice, "target-001").Return(0.9, nil)
	eeRepo.On("CreateAnalysis", mock.Anything, mock.AnythingOfType("*entity.EnergyEfficiencyAnalysis")).Return(nil)

	body := map[string]interface{}{
		"target_id": "target-001", "target_name": "Test Target", "type": "device",
		"time_range_start": "2024-01-01T00:00:00Z", "time_range_end": "2024-01-31T00:00:00Z",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ee/analyses", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEEHandler_CreateEnergyEfficiencyAnalysis_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	r.POST("/ee/analyses", handler.CreateEnergyEfficiencyAnalysis)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ee/analyses", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEEHandler_CreateEnergyEfficiencyAnalysis_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	r.POST("/ee/analyses", handler.CreateEnergyEfficiencyAnalysis)

	eeRepo.On("GetStatistics", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	body := map[string]interface{}{
		"target_id": "target-001", "target_name": "Test Target", "type": "device",
		"time_range_start": "2024-01-01T00:00:00Z", "time_range_end": "2024-01-31T00:00:00Z",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ee/analyses", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEEHandler_ListEnergyEfficiencyRecords_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	r.GET("/ee/records", handler.ListEnergyEfficiencyRecords)

	eeRepo.On("ListRecords", mock.Anything, mock.AnythingOfType("*repository.EnergyEfficiencyQuery")).Return([]*entity.EnergyEfficiencyRecord{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/records?type=device&level=excellent&target_id=t-1&period=2024-01&start_time=2024-01-01T00:00:00Z&end_time=2024-01-31T00:00:00Z", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEEHandler_ListEnergyEfficiencyAnalyses_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	r.GET("/ee/analyses", handler.ListEnergyEfficiencyAnalyses)

	eeRepo.On("ListAnalyses", mock.Anything, mock.AnythingOfType("*repository.EnergyEfficiencyAnalysisQuery")).Return([]*entity.EnergyEfficiencyAnalysis{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/analyses?type=device&target_id=t-1&start_time=2024-01-01T00:00:00Z&end_time=2024-01-31T00:00:00Z", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEEHandler_GetLatestEnergyEfficiencyAnalysis_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	r.GET("/ee/analyses/latest", handler.GetLatestEnergyEfficiencyAnalysis)

	eeRepo.On("GetLatestAnalysis", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice).Return(&entity.EnergyEfficiencyAnalysis{ID: "a-001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/analyses/latest?target_id=target-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEEHandler_GetLatestEnergyEfficiencyAnalysis_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	r.GET("/ee/analyses/latest", handler.GetLatestEnergyEfficiencyAnalysis)

	eeRepo.On("GetLatestAnalysis", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/analyses/latest?target_id=target-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAlarmRuleHandler_ListAlarmRules_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)
	r := gin.New()
	r.GET("/alarm-rules", handler.ListAlarmRules)

	rules := []*entity.AlarmRule{entity.NewAlarmRule("R1", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "c1")}
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*repository.AlarmRuleQuery")).Return(rules, int64(1), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/alarm-rules?type=limit&level=2&status=1&station_id=st-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlarmRuleHandler_UpdateAlarmRule_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)
	r := gin.New()
	r.PUT("/alarm-rules/:id", func(c *gin.Context) {
		c.Set("user_id", "admin")
		handler.UpdateAlarmRule(c)
	})

	rule := entity.NewAlarmRule("Test Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "c1")
	mockRepo.On("GetByID", mock.Anything, "rule-001").Return(rule, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

	body := service.UpdateAlarmRuleRequest{Name: stringPtr("Updated Rule")}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/alarm-rules/rule-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlarmRuleHandler_DeleteAlarmRule_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)
	r := gin.New()
	r.DELETE("/alarm-rules/:id", handler.DeleteAlarmRule)

	mockRepo.On("Delete", mock.Anything, "nonexistent").Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/alarm-rules/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAlarmRuleHandler_GetRulesByPoint_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)
	r := gin.New()
	r.GET("/points/:point_id/rules", handler.GetRulesByPoint)

	mockRepo.On("GetRulesByPointID", mock.Anything, "point-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/points/point-001/rules", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAlarmRuleHandler_GetRulesByDevice_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)
	r := gin.New()
	r.GET("/devices/:device_id/rules", handler.GetRulesByDevice)

	mockRepo.On("GetRulesByDeviceID", mock.Anything, "device-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/devices/device-001/rules", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAlarmRuleHandler_GetRulesByStation_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mockAlarmRuleRepoForHandler)
	svc := service.NewAlarmRuleService(mockRepo)
	handler := NewAlarmRuleHandler(svc)
	r := gin.New()
	r.GET("/stations/:station_id/rules", handler.GetRulesByStation)

	mockRepo.On("GetRulesByStationID", mock.Anything, "station-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stations/station-001/rules", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAlarmRuleHandler_CreateAlarmRule_Error(t *testing.T) {
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
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AlarmRule")).Return(assert.AnError)

	body := service.CreateAlarmRuleRequest{
		Name: "Test Rule", Type: entity.AlarmRuleTypeLimit, Level: entity.AlarmLevelWarning,
		Condition: "value > threshold", Threshold: 85.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/alarm-rules", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostEntryHandler_ListCostEntries_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	entryRepo := new(mockCostEntryRepoForCov)
	catRepo := new(mockCostCategoryRepoForCov)
	svc := service.NewCostEntryService(entryRepo, catRepo)
	handler := NewCostEntryHandler(svc)
	r := gin.New()
	r.GET("/cost-entries", handler.ListCostEntries)

	entryRepo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entity.CostEntry{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-entries?category_id=cat-1&start_date=2024-01-01&end_date=2024-12-31&status=pending&page=2&page_size=20", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostEntryHandler_RejectCostEntry_Cov(t *testing.T) {
	gin.SetMode(gin.TestMode)
	entryRepo := new(mockCostEntryRepoForCov)
	catRepo := new(mockCostCategoryRepoForCov)
	svc := service.NewCostEntryService(entryRepo, catRepo)
	handler := NewCostEntryHandler(svc)
	r := gin.New()
	r.POST("/cost-entries/:id/reject", handler.RejectCostEntry)

	pendingEntry := &entity.CostEntry{ID: "ce-001", ApprovalStatus: "pending"}
	entryRepo.On("GetByID", mock.Anything, "ce-001").Return(pendingEntry, nil)
	entryRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostEntry")).Return(nil)

	body := map[string]string{"approved_by": "admin"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cost-entries/ce-001/reject", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostEntryHandler_GetTotalByCategory_Cov(t *testing.T) {
	gin.SetMode(gin.TestMode)
	entryRepo := new(mockCostEntryRepoForCov)
	catRepo := new(mockCostCategoryRepoForCov)
	svc := service.NewCostEntryService(entryRepo, catRepo)
	handler := NewCostEntryHandler(svc)
	r := gin.New()
	r.GET("/cost-categories/:category_id/total", handler.GetTotalByCategory)

	catRepo.On("GetByID", mock.Anything, "cat-001").Return(&entity.CostCategory{ID: "cat-001"}, nil)
	entryRepo.On("GetTotalByCategory", mock.Anything, "cat-001", mock.Anything, mock.Anything).Return(10000.0, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-categories/cat-001/total?start_date=2024-01-01&end_date=2024-12-31", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostEntryHandler_GetTotalByPeriod_Cov(t *testing.T) {
	gin.SetMode(gin.TestMode)
	entryRepo := new(mockCostEntryRepoForCov)
	catRepo := new(mockCostCategoryRepoForCov)
	svc := service.NewCostEntryService(entryRepo, catRepo)
	handler := NewCostEntryHandler(svc)
	r := gin.New()
	r.GET("/cost-entries/total", handler.GetTotalByPeriod)

	entryRepo.On("GetTotalByPeriod", mock.Anything, mock.Anything, mock.Anything).Return(50000.0, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-entries/total?start_date=2024-01-01&end_date=2024-12-31", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostAllocationHandler_GetTotalByAllocated_WithDates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	allocRepo := new(mockCostAllocationRepoForCov)
	entryRepo := new(mockCostEntryRepoForCov)
	svc := service.NewCostAllocationService(allocRepo, entryRepo)
	handler := NewCostAllocationHandler(svc)
	r := gin.New()
	r.GET("/cost-allocations/total", handler.GetTotalByAllocated)

	allocRepo.On("GetTotalByAllocated", mock.Anything, "project", "proj-001", mock.Anything, mock.Anything).Return(5000.0, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-allocations/total?allocated_to=project&allocated_id=proj-001&start_date=2024-01-01&end_date=2024-12-31", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostAllocationHandler_ListByAllocated_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	allocRepo := new(mockCostAllocationRepoForCov)
	entryRepo := new(mockCostEntryRepoForCov)
	svc := service.NewCostAllocationService(allocRepo, entryRepo)
	handler := NewCostAllocationHandler(svc)
	r := gin.New()
	r.GET("/cost-allocations/by-allocated", handler.ListCostAllocationsByAllocated)

	allocRepo.On("ListByAllocated", mock.Anything, "project", "proj-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-allocations/by-allocated?allocated_to=project&allocated_id=proj-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostReportHandler_ListCostReports_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reportRepo := new(mockCostReportRepoForCov)
	entryRepo := new(mockCostEntryRepoForCov)
	svc := service.NewCostReportService(reportRepo, entryRepo)
	handler := NewCostReportHandler(svc)
	r := gin.New()
	r.GET("/cost-reports", handler.ListCostReports)

	reportRepo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entity.CostReport{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-reports?report_type=monthly&status=draft&start_date=2024-01-01&end_date=2024-12-31&page=2&page_size=20", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostReportHandler_RejectCostReport_Cov(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reportRepo := new(mockCostReportRepoForCov)
	entryRepo := new(mockCostEntryRepoForCov)
	svc := service.NewCostReportService(reportRepo, entryRepo)
	handler := NewCostReportHandler(svc)
	r := gin.New()
	r.POST("/cost-reports/:id/reject", handler.RejectCostReport)

	report := &entity.CostReport{ID: "cr-001", Status: "generated"}
	reportRepo.On("GetByID", mock.Anything, "cr-001").Return(report, nil)
	reportRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostReport")).Return(nil)

	body := map[string]string{"approved_by": "admin"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cost-reports/cr-001/reject", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostEntryHandler_GetTotalByPeriod_NoDates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	entryRepo := new(mockCostEntryRepoForCov)
	catRepo := new(mockCostCategoryRepoForCov)
	svc := service.NewCostEntryService(entryRepo, catRepo)
	handler := NewCostEntryHandler(svc)
	r := gin.New()
	r.GET("/cost-entries/total", handler.GetTotalByPeriod)

	entryRepo.On("GetTotalByPeriod", mock.Anything, (*time.Time)(nil), (*time.Time)(nil)).Return(50000.0, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-entries/total", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func stringPtr(s string) *string {
	return &s
}
