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

type mockEERepo struct {
	mock.Mock
}

func (m *mockEERepo) CreateRecord(ctx context.Context, record *entity.EnergyEfficiencyRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *mockEERepo) BatchCreateRecords(ctx context.Context, records []*entity.EnergyEfficiencyRecord) error {
	args := m.Called(ctx, records)
	return args.Error(0)
}

func (m *mockEERepo) GetRecordByID(ctx context.Context, id string) (*entity.EnergyEfficiencyRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EnergyEfficiencyRecord), args.Error(1)
}

func (m *mockEERepo) ListRecords(ctx context.Context, query *repository.EnergyEfficiencyQuery) ([]*entity.EnergyEfficiencyRecord, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.EnergyEfficiencyRecord), args.Get(1).(int64), args.Error(2)
}

func (m *mockEERepo) GetRecordsByTimeRange(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType, startTime, endTime time.Time) ([]*entity.EnergyEfficiencyRecord, error) {
	args := m.Called(ctx, targetID, eeType, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.EnergyEfficiencyRecord), args.Error(1)
}

func (m *mockEERepo) CreateAnalysis(ctx context.Context, analysis *entity.EnergyEfficiencyAnalysis) error {
	args := m.Called(ctx, analysis)
	return args.Error(0)
}

func (m *mockEERepo) GetAnalysisByID(ctx context.Context, id string) (*entity.EnergyEfficiencyAnalysis, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EnergyEfficiencyAnalysis), args.Error(1)
}

func (m *mockEERepo) ListAnalyses(ctx context.Context, query *repository.EnergyEfficiencyAnalysisQuery) ([]*entity.EnergyEfficiencyAnalysis, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.EnergyEfficiencyAnalysis), args.Get(1).(int64), args.Error(2)
}

func (m *mockEERepo) GetLatestAnalysis(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType) (*entity.EnergyEfficiencyAnalysis, error) {
	args := m.Called(ctx, targetID, eeType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EnergyEfficiencyAnalysis), args.Error(1)
}

func (m *mockEERepo) GetStatistics(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType, period string, startTime, endTime time.Time) (*repository.EnergyEfficiencyStatistics, error) {
	args := m.Called(ctx, targetID, eeType, period, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.EnergyEfficiencyStatistics), args.Error(1)
}

func (m *mockEERepo) GetBenchmark(ctx context.Context, eeType entity.EnergyEfficiencyType, targetID string) (float64, error) {
	args := m.Called(ctx, eeType, targetID)
	return args.Get(0).(float64), args.Error(1)
}

func setupEEHandler(eeRepo *mockEERepo) (*EnergyEfficiencyHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	r := gin.New()
	return handler, r
}

func TestEEHandler_CreateEnergyEfficiencyRecord_Success(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.POST("/ee/records", handler.CreateEnergyEfficiencyRecord)

	eeRepo.On("GetBenchmark", mock.Anything, mock.Anything, "target-001").Return(0.0, assert.AnError)
	eeRepo.On("CreateRecord", mock.Anything, mock.AnythingOfType("*entity.EnergyEfficiencyRecord")).Return(nil)

	body := map[string]interface{}{
		"record_time": "2024-01-01T00:00:00Z", "type": "device", "target_id": "target-001",
		"target_name": "Test Target", "input_energy": 100.0, "output_energy": 85.0, "period": "monthly",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ee/records", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEEHandler_CreateEnergyEfficiencyRecord_BadRequest(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.POST("/ee/records", handler.CreateEnergyEfficiencyRecord)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ee/records", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEEHandler_BatchCreateEnergyEfficiencyRecords_Success(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.POST("/ee/records/batch", handler.BatchCreateEnergyEfficiencyRecords)

	eeRepo.On("BatchCreateRecords", mock.Anything, mock.Anything).Return(nil)

	body := []map[string]interface{}{
		{"record_time": "2024-01-01T00:00:00Z", "type": "device", "target_id": "target-001",
			"target_name": "Test", "input_energy": 100.0, "output_energy": 85.0, "period": "monthly"},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ee/records/batch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEEHandler_BatchCreateEnergyEfficiencyRecords_BadRequest(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.POST("/ee/records/batch", handler.BatchCreateEnergyEfficiencyRecords)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ee/records/batch", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEEHandler_GetEnergyEfficiencyRecord_Success(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.GET("/ee/records/:id", handler.GetEnergyEfficiencyRecord)

	eeRepo.On("GetRecordByID", mock.Anything, "rec-001").Return(&entity.EnergyEfficiencyRecord{ID: "rec-001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/records/rec-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEEHandler_GetEnergyEfficiencyRecord_NotFound(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.GET("/ee/records/:id", handler.GetEnergyEfficiencyRecord)

	eeRepo.On("GetRecordByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/records/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestEEHandler_ListEnergyEfficiencyRecords_Success(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.GET("/ee/records", handler.ListEnergyEfficiencyRecords)

	eeRepo.On("ListRecords", mock.Anything, mock.AnythingOfType("*repository.EnergyEfficiencyQuery")).Return([]*entity.EnergyEfficiencyRecord{{ID: "rec-001"}}, int64(1), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/records", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEEHandler_ListEnergyEfficiencyRecords_Error(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.GET("/ee/records", handler.ListEnergyEfficiencyRecords)

	eeRepo.On("ListRecords", mock.Anything, mock.AnythingOfType("*repository.EnergyEfficiencyQuery")).Return(([]*entity.EnergyEfficiencyRecord)(nil), int64(0), assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/records", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEEHandler_GetEnergyEfficiencyTrend_MissingTargetID(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.GET("/ee/trend", handler.GetEnergyEfficiencyTrend)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/trend", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEEHandler_GetEnergyEfficiencyStatistics_MissingTargetID(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.GET("/ee/statistics", handler.GetEnergyEfficiencyStatistics)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/statistics", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEEHandler_GetEnergyEfficiencyComparison_MissingTargetID(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.GET("/ee/comparison", handler.GetEnergyEfficiencyComparison)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/comparison", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEEHandler_GetLatestEnergyEfficiencyAnalysis_MissingTargetID(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.GET("/ee/analyses/latest", handler.GetLatestEnergyEfficiencyAnalysis)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/analyses/latest", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEEHandler_GetEnergyEfficiencyAnalysis_Success(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.GET("/ee/analyses/:id", handler.GetEnergyEfficiencyAnalysis)

	eeRepo.On("GetAnalysisByID", mock.Anything, "analysis-001").Return(&entity.EnergyEfficiencyAnalysis{ID: "analysis-001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/analyses/analysis-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEEHandler_GetEnergyEfficiencyAnalysis_NotFound(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.GET("/ee/analyses/:id", handler.GetEnergyEfficiencyAnalysis)

	eeRepo.On("GetAnalysisByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/analyses/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestEEHandler_ListEnergyEfficiencyAnalyses_Success(t *testing.T) {
	eeRepo := new(mockEERepo)
	handler, r := setupEEHandler(eeRepo)

	r.GET("/ee/analyses", handler.ListEnergyEfficiencyAnalyses)

	eeRepo.On("ListAnalyses", mock.Anything, mock.AnythingOfType("*repository.EnergyEfficiencyAnalysisQuery")).Return([]*entity.EnergyEfficiencyAnalysis{{ID: "analysis-001"}}, int64(1), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ee/analyses", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewEnergyEfficiencyHandler(t *testing.T) {
	eeRepo := new(mockEERepo)
	svc := service.NewEnergyEfficiencyService(eeRepo)
	handler := NewEnergyEfficiencyHandler(svc)
	assert.NotNil(t, handler)
}

var _ repository.EnergyEfficiencyRepository = (*mockEERepo)(nil)
