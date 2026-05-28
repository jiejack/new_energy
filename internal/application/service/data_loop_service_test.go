package service

import (
	"context"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockForecastResultRepoDL struct {
	mock.Mock
}

func (m *mockForecastResultRepoDL) Create(ctx context.Context, result *entity.ForecastResult) error {
	return m.Called(ctx, result).Error(0)
}

func (m *mockForecastResultRepoDL) GetByID(ctx context.Context, id string) (*entity.ForecastResult, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ForecastResult), args.Error(1)
}

func (m *mockForecastResultRepoDL) ListByStation(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime *time.Time, offset, limit int) ([]*entity.ForecastResult, int64, error) {
	args := m.Called(ctx, stationID, forecastType, startTime, endTime, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.ForecastResult), args.Get(1).(int64), args.Error(2)
}

func (m *mockForecastResultRepoDL) UpdateActualPower(ctx context.Context, id string, actualPower float64) error {
	return m.Called(ctx, id, actualPower).Error(0)
}

func (m *mockForecastResultRepoDL) GetAccuracyStats(ctx context.Context, stationID string, forecastType *entity.ForecastType, startTime, endTime time.Time) (*repository.ForecastAccuracyStats, error) {
	args := m.Called(ctx, stationID, forecastType, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ForecastAccuracyStats), args.Error(1)
}

var _ repository.ForecastResultRepository = (*mockForecastResultRepoDL)(nil)

type mockModelSvcDL struct {
	mock.Mock
}

func (m *mockModelSvcDL) RegisterModel(ctx context.Context, modelName, version, artifactPath string, accuracy float64) (*entity.ModelVersion, error) {
	args := m.Called(ctx, modelName, version, artifactPath, accuracy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ModelVersion), args.Error(1)
}

func (m *mockModelSvcDL) GetModel(ctx context.Context, id string) (*entity.ModelVersion, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ModelVersion), args.Error(1)
}

func (m *mockModelSvcDL) GetProductionModel(ctx context.Context, modelName string) (*entity.ModelVersion, error) {
	args := m.Called(ctx, modelName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ModelVersion), args.Error(1)
}

func (m *mockModelSvcDL) ListModels(ctx context.Context, modelName string) ([]*entity.ModelVersion, error) {
	args := m.Called(ctx, modelName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ModelVersion), args.Error(1)
}

func (m *mockModelSvcDL) PromoteToProduction(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockModelSvcDL) RetireModel(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

var _ ModelService = (*mockModelSvcDL)(nil)

func TestDataLoopService_CollectFeedback_Success(t *testing.T) {
	forecastRepo := new(mockForecastResultRepoDL)
	modelSvc := new(mockModelSvcDL)
	svc := NewDataLoopService(forecastRepo, modelSvc)

	stats := &repository.ForecastAccuracyStats{TotalPoints: 100, RMSE: 0.05, AvgAccuracy: 0.95}
	forecastRepo.On("GetAccuracyStats", mock.Anything, "station-001", mock.Anything, mock.Anything, mock.Anything).Return(stats, nil)

	report, err := svc.CollectFeedback(context.Background(), "station-001", entity.ForecastTypeShortTerm, time.Now())
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "station-001", report.StationID)
	assert.Equal(t, 100, report.TotalPoints)
	assert.False(t, report.DriftDetected)
}

func TestDataLoopService_CollectFeedback_DriftDetected(t *testing.T) {
	forecastRepo := new(mockForecastResultRepoDL)
	modelSvc := new(mockModelSvcDL)
	svc := NewDataLoopService(forecastRepo, modelSvc)

	stats := &repository.ForecastAccuracyStats{TotalPoints: 100, RMSE: 0.25, AvgAccuracy: 0.70}
	forecastRepo.On("GetAccuracyStats", mock.Anything, "station-001", mock.Anything, mock.Anything, mock.Anything).Return(stats, nil)

	report, err := svc.CollectFeedback(context.Background(), "station-001", entity.ForecastTypeShortTerm, time.Now())
	assert.NoError(t, err)
	assert.True(t, report.DriftDetected)
}

func TestDataLoopService_CollectFeedback_Error(t *testing.T) {
	forecastRepo := new(mockForecastResultRepoDL)
	modelSvc := new(mockModelSvcDL)
	svc := NewDataLoopService(forecastRepo, modelSvc)

	forecastRepo.On("GetAccuracyStats", mock.Anything, "station-001", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	report, err := svc.CollectFeedback(context.Background(), "station-001", entity.ForecastTypeShortTerm, time.Now())
	assert.Error(t, err)
	assert.Nil(t, report)
}

func TestDataLoopService_EvaluateAndTrigger_ShouldRetrain(t *testing.T) {
	forecastRepo := new(mockForecastResultRepoDL)
	modelSvc := new(mockModelSvcDL)
	svc := NewDataLoopService(forecastRepo, modelSvc)

	stats := &repository.ForecastAccuracyStats{AvgAccuracy: 0.75}
	forecastRepo.On("GetAccuracyStats", mock.Anything, "station-001", mock.Anything, mock.Anything, mock.Anything).Return(stats, nil)

	decision, err := svc.EvaluateAndTrigger(context.Background(), "station-001")
	assert.NoError(t, err)
	assert.True(t, decision.ShouldRetrain)
}

func TestDataLoopService_EvaluateAndTrigger_NoRetrain(t *testing.T) {
	forecastRepo := new(mockForecastResultRepoDL)
	modelSvc := new(mockModelSvcDL)
	svc := NewDataLoopService(forecastRepo, modelSvc)

	stats := &repository.ForecastAccuracyStats{AvgAccuracy: 0.90}
	forecastRepo.On("GetAccuracyStats", mock.Anything, "station-001", mock.Anything, mock.Anything, mock.Anything).Return(stats, nil)

	decision, err := svc.EvaluateAndTrigger(context.Background(), "station-001")
	assert.NoError(t, err)
	assert.False(t, decision.ShouldRetrain)
}

func TestDataLoopService_EvaluateAndTrigger_Error(t *testing.T) {
	forecastRepo := new(mockForecastResultRepoDL)
	modelSvc := new(mockModelSvcDL)
	svc := NewDataLoopService(forecastRepo, modelSvc)

	forecastRepo.On("GetAccuracyStats", mock.Anything, "station-001", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	decision, err := svc.EvaluateAndTrigger(context.Background(), "station-001")
	assert.NoError(t, err)
	assert.False(t, decision.ShouldRetrain)
}

func TestDataLoopService_GetLoopStatus_Success(t *testing.T) {
	forecastRepo := new(mockForecastResultRepoDL)
	modelSvc := new(mockModelSvcDL)
	svc := NewDataLoopService(forecastRepo, modelSvc)

	stats := &repository.ForecastAccuracyStats{AvgAccuracy: 0.92}
	forecastRepo.On("GetAccuracyStats", mock.Anything, "station-001", mock.Anything, mock.Anything, mock.Anything).Return(stats, nil)

	status, err := svc.GetLoopStatus(context.Background(), "station-001")
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, 0.92, status.CurrentAccuracy)
}

func TestDataLoopService_GetLoopStatus_Error(t *testing.T) {
	forecastRepo := new(mockForecastResultRepoDL)
	modelSvc := new(mockModelSvcDL)
	svc := NewDataLoopService(forecastRepo, modelSvc)

	forecastRepo.On("GetAccuracyStats", mock.Anything, "station-001", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	status, err := svc.GetLoopStatus(context.Background(), "station-001")
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, 0.0, status.CurrentAccuracy)
}

func TestNewDataLoopService(t *testing.T) {
	forecastRepo := new(mockForecastResultRepoDL)
	modelSvc := new(mockModelSvcDL)
	svc := NewDataLoopService(forecastRepo, modelSvc)
	assert.NotNil(t, svc)
}
