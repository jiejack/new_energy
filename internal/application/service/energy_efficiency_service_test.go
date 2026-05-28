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

type mockEERepoSvc struct {
	mock.Mock
}

func (m *mockEERepoSvc) CreateRecord(ctx context.Context, record *entity.EnergyEfficiencyRecord) error {
	return m.Called(ctx, record).Error(0)
}

func (m *mockEERepoSvc) BatchCreateRecords(ctx context.Context, records []*entity.EnergyEfficiencyRecord) error {
	return m.Called(ctx, records).Error(0)
}

func (m *mockEERepoSvc) GetRecordByID(ctx context.Context, id string) (*entity.EnergyEfficiencyRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EnergyEfficiencyRecord), args.Error(1)
}

func (m *mockEERepoSvc) ListRecords(ctx context.Context, query *repository.EnergyEfficiencyQuery) ([]*entity.EnergyEfficiencyRecord, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.EnergyEfficiencyRecord), args.Get(1).(int64), args.Error(2)
}

func (m *mockEERepoSvc) GetRecordsByTimeRange(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType, startTime, endTime time.Time) ([]*entity.EnergyEfficiencyRecord, error) {
	args := m.Called(ctx, targetID, eeType, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.EnergyEfficiencyRecord), args.Error(1)
}

func (m *mockEERepoSvc) CreateAnalysis(ctx context.Context, analysis *entity.EnergyEfficiencyAnalysis) error {
	return m.Called(ctx, analysis).Error(0)
}

func (m *mockEERepoSvc) GetAnalysisByID(ctx context.Context, id string) (*entity.EnergyEfficiencyAnalysis, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EnergyEfficiencyAnalysis), args.Error(1)
}

func (m *mockEERepoSvc) ListAnalyses(ctx context.Context, query *repository.EnergyEfficiencyAnalysisQuery) ([]*entity.EnergyEfficiencyAnalysis, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.EnergyEfficiencyAnalysis), args.Get(1).(int64), args.Error(2)
}

func (m *mockEERepoSvc) GetLatestAnalysis(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType) (*entity.EnergyEfficiencyAnalysis, error) {
	args := m.Called(ctx, targetID, eeType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EnergyEfficiencyAnalysis), args.Error(1)
}

func (m *mockEERepoSvc) GetStatistics(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType, period string, startTime, endTime time.Time) (*repository.EnergyEfficiencyStatistics, error) {
	args := m.Called(ctx, targetID, eeType, period, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.EnergyEfficiencyStatistics), args.Error(1)
}

func (m *mockEERepoSvc) GetBenchmark(ctx context.Context, eeType entity.EnergyEfficiencyType, targetID string) (float64, error) {
	args := m.Called(ctx, eeType, targetID)
	return args.Get(0).(float64), args.Error(1)
}

var _ repository.EnergyEfficiencyRepository = (*mockEERepoSvc)(nil)

func TestEEService_CreateRecord_Success(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	repo.On("GetBenchmark", mock.Anything, mock.Anything, "target-001").Return(0.0, assert.AnError)
	repo.On("CreateRecord", mock.Anything, mock.AnythingOfType("*entity.EnergyEfficiencyRecord")).Return(nil)

	req := &CreateEnergyEfficiencyRecordRequest{
		RecordTime: time.Now(), Type: entity.EnergyEfficiencyTypeDevice, TargetID: "target-001",
		TargetName: "Test", InputEnergy: 100, OutputEnergy: 85, Period: "monthly",
	}
	record, err := svc.CreateRecord(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestEEService_CreateRecord_WithBenchmark(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	repo.On("GetBenchmark", mock.Anything, mock.Anything, "target-001").Return(0.80, nil)
	repo.On("CreateRecord", mock.Anything, mock.AnythingOfType("*entity.EnergyEfficiencyRecord")).Return(nil)

	req := &CreateEnergyEfficiencyRecordRequest{
		RecordTime: time.Now(), Type: entity.EnergyEfficiencyTypeDevice, TargetID: "target-001",
		TargetName: "Test", InputEnergy: 100, OutputEnergy: 85, Period: "monthly",
	}
	record, err := svc.CreateRecord(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, 0.80, record.BenchmarkEfficiency)
}

func TestEEService_CreateRecord_RepoError(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	repo.On("GetBenchmark", mock.Anything, mock.Anything, "target-001").Return(0.0, assert.AnError)
	repo.On("CreateRecord", mock.Anything, mock.AnythingOfType("*entity.EnergyEfficiencyRecord")).Return(assert.AnError)

	req := &CreateEnergyEfficiencyRecordRequest{
		RecordTime: time.Now(), Type: entity.EnergyEfficiencyTypeDevice, TargetID: "target-001",
		TargetName: "Test", InputEnergy: 100, OutputEnergy: 85, Period: "monthly",
	}
	record, err := svc.CreateRecord(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestEEService_BatchCreateRecords_Success(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	repo.On("BatchCreateRecords", mock.Anything, mock.Anything).Return(nil)

	reqs := []*CreateEnergyEfficiencyRecordRequest{
		{RecordTime: time.Now(), Type: entity.EnergyEfficiencyTypeDevice, TargetID: "t1", TargetName: "T1", InputEnergy: 100, OutputEnergy: 85, Period: "monthly"},
	}
	records, err := svc.BatchCreateRecords(context.Background(), reqs)
	assert.NoError(t, err)
	assert.Len(t, records, 1)
}

func TestEEService_BatchCreateRecords_Error(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	repo.On("BatchCreateRecords", mock.Anything, mock.Anything).Return(assert.AnError)

	reqs := []*CreateEnergyEfficiencyRecordRequest{
		{RecordTime: time.Now(), Type: entity.EnergyEfficiencyTypeDevice, TargetID: "t1", TargetName: "T1", InputEnergy: 100, OutputEnergy: 85, Period: "monthly"},
	}
	records, err := svc.BatchCreateRecords(context.Background(), reqs)
	assert.Error(t, err)
	assert.Nil(t, records)
}

func TestEEService_GetRecord_Success(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	repo.On("GetRecordByID", mock.Anything, "rec-001").Return(&entity.EnergyEfficiencyRecord{ID: "rec-001"}, nil)

	record, err := svc.GetRecord(context.Background(), "rec-001")
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestEEService_ListRecords_Success(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	repo.On("ListRecords", mock.Anything, mock.AnythingOfType("*repository.EnergyEfficiencyQuery")).Return([]*entity.EnergyEfficiencyRecord{{ID: "rec-001"}}, int64(1), nil)

	records, total, err := svc.ListRecords(context.Background(), &QueryEnergyEfficiencyRecordsRequest{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, records, 1)
}

func TestEEService_GetTrendData_Success(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	now := time.Now()
	repo.On("GetRecordsByTimeRange", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, mock.Anything, mock.Anything).Return([]*entity.EnergyEfficiencyRecord{
		{RecordTime: now, Efficiency: 0.85, InputEnergy: 100, OutputEnergy: 85},
	}, nil)

	trend, err := svc.GetTrendData(context.Background(), "target-001", entity.EnergyEfficiencyTypeDevice, now.AddDate(0, -1, 0), now)
	assert.NoError(t, err)
	assert.Len(t, trend, 1)
	assert.Equal(t, 0.85, trend[0].Efficiency)
}

func TestEEService_GetTrendData_Error(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	now := time.Now()
	repo.On("GetRecordsByTimeRange", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	trend, err := svc.GetTrendData(context.Background(), "target-001", entity.EnergyEfficiencyTypeDevice, now.AddDate(0, -1, 0), now)
	assert.Error(t, err)
	assert.Nil(t, trend)
}

func TestEEService_GetStatistics_Success(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	now := time.Now()
	stats := &repository.EnergyEfficiencyStatistics{AvgEfficiency: 0.85}
	repo.On("GetStatistics", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, "monthly", mock.Anything, mock.Anything).Return(stats, nil)

	result, err := svc.GetStatistics(context.Background(), "target-001", entity.EnergyEfficiencyTypeDevice, "monthly", now.AddDate(0, -1, 0), now)
	assert.NoError(t, err)
	assert.Equal(t, 0.85, result.AvgEfficiency)
}

func TestEEService_GetComparisonData_Success(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	now := time.Now()
	stats := &repository.EnergyEfficiencyStatistics{AvgEfficiency: 0.85}
	repo.On("GetStatistics", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, "monthly", mock.Anything, mock.Anything).Return(stats, nil)

	result, err := svc.GetComparisonData(context.Background(), "target-001", entity.EnergyEfficiencyTypeDevice, "monthly", now.AddDate(0, -1, 0), now)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestEEService_GetComparisonData_CurrentStatsError(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	now := time.Now()
	repo.On("GetStatistics", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, "monthly", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	result, err := svc.GetComparisonData(context.Background(), "target-001", entity.EnergyEfficiencyTypeDevice, "monthly", now.AddDate(0, -1, 0), now)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestEEService_CreateAnalysis_Success(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	now := time.Now()
	stats := &repository.EnergyEfficiencyStatistics{AvgEfficiency: 0.92, MaxEfficiency: 0.95, MinEfficiency: 0.88, StdDevEfficiency: 0.02, TotalRecords: 5, TotalInputEnergy: 1000}
	repo.On("GetStatistics", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, "month", mock.Anything, mock.Anything).Return(stats, nil)
	repo.On("GetRecordsByTimeRange", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, mock.Anything, mock.Anything).Return([]*entity.EnergyEfficiencyRecord{
		{Efficiency: 0.90}, {Efficiency: 0.91}, {Efficiency: 0.92}, {Efficiency: 0.93}, {Efficiency: 0.94},
	}, nil)
	repo.On("CreateAnalysis", mock.Anything, mock.AnythingOfType("*entity.EnergyEfficiencyAnalysis")).Return(nil)

	analysis, err := svc.CreateAnalysis(context.Background(), "target-001", entity.EnergyEfficiencyTypeDevice, "Target 1", now.AddDate(0, -1, 0), now)
	assert.NoError(t, err)
	assert.NotNil(t, analysis)
}

func TestEEService_CreateAnalysis_StatsError(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	now := time.Now()
	repo.On("GetStatistics", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice, "month", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	analysis, err := svc.CreateAnalysis(context.Background(), "target-001", entity.EnergyEfficiencyTypeDevice, "Target 1", now.AddDate(0, -1, 0), now)
	assert.Error(t, err)
	assert.Nil(t, analysis)
}

func TestEEService_GetAnalysis_Success(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	repo.On("GetAnalysisByID", mock.Anything, "analysis-001").Return(&entity.EnergyEfficiencyAnalysis{ID: "analysis-001"}, nil)

	analysis, err := svc.GetAnalysis(context.Background(), "analysis-001")
	assert.NoError(t, err)
	assert.NotNil(t, analysis)
}

func TestEEService_ListAnalyses_Success(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	repo.On("ListAnalyses", mock.Anything, mock.AnythingOfType("*repository.EnergyEfficiencyAnalysisQuery")).Return([]*entity.EnergyEfficiencyAnalysis{{ID: "analysis-001"}}, int64(1), nil)

	analyses, total, err := svc.ListAnalyses(context.Background(), &QueryEnergyEfficiencyAnalysesRequest{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, analyses, 1)
}

func TestEEService_GetLatestAnalysis_Success(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)

	repo.On("GetLatestAnalysis", mock.Anything, "target-001", entity.EnergyEfficiencyTypeDevice).Return(&entity.EnergyEfficiencyAnalysis{ID: "analysis-001"}, nil)

	analysis, err := svc.GetLatestAnalysis(context.Background(), "target-001", entity.EnergyEfficiencyTypeDevice)
	assert.NoError(t, err)
	assert.NotNil(t, analysis)
}

func TestNewEnergyEfficiencyService(t *testing.T) {
	repo := new(mockEERepoSvc)
	svc := NewEnergyEfficiencyService(repo)
	assert.NotNil(t, svc)
}
