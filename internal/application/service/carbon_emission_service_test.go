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

type mockCERepo struct {
	mock.Mock
}

func (m *mockCERepo) CreateFactor(ctx context.Context, factor *entity.CarbonEmissionFactor) error {
	return m.Called(ctx, factor).Error(0)
}

func (m *mockCERepo) UpdateFactor(ctx context.Context, factor *entity.CarbonEmissionFactor) error {
	return m.Called(ctx, factor).Error(0)
}

func (m *mockCERepo) GetFactorByID(ctx context.Context, id string) (*entity.CarbonEmissionFactor, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CarbonEmissionFactor), args.Error(1)
}

func (m *mockCERepo) GetFactorByCode(ctx context.Context, code string) (*entity.CarbonEmissionFactor, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CarbonEmissionFactor), args.Error(1)
}

func (m *mockCERepo) ListFactors(ctx context.Context, query *repository.CarbonEmissionFactorQuery) ([]*entity.CarbonEmissionFactor, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.CarbonEmissionFactor), args.Get(1).(int64), args.Error(2)
}

func (m *mockCERepo) GetActiveFactors(ctx context.Context, scope *entity.CarbonEmissionScope) ([]*entity.CarbonEmissionFactor, error) {
	args := m.Called(ctx, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CarbonEmissionFactor), args.Error(1)
}

func (m *mockCERepo) CreateRecord(ctx context.Context, record *entity.CarbonEmissionRecord) error {
	return m.Called(ctx, record).Error(0)
}

func (m *mockCERepo) BatchCreateRecords(ctx context.Context, records []*entity.CarbonEmissionRecord) error {
	return m.Called(ctx, records).Error(0)
}

func (m *mockCERepo) UpdateRecord(ctx context.Context, record *entity.CarbonEmissionRecord) error {
	return m.Called(ctx, record).Error(0)
}

func (m *mockCERepo) GetRecordByID(ctx context.Context, id string) (*entity.CarbonEmissionRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CarbonEmissionRecord), args.Error(1)
}

func (m *mockCERepo) ListRecords(ctx context.Context, query *repository.CarbonEmissionRecordQuery) ([]*entity.CarbonEmissionRecord, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.CarbonEmissionRecord), args.Get(1).(int64), args.Error(2)
}

func (m *mockCERepo) GetRecordsByTimeRange(ctx context.Context, targetID string, scope *entity.CarbonEmissionScope, startTime, endTime time.Time) ([]*entity.CarbonEmissionRecord, error) {
	args := m.Called(ctx, targetID, scope, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CarbonEmissionRecord), args.Error(1)
}

func (m *mockCERepo) GetTotalEmissionByScope(ctx context.Context, targetID string, scope entity.CarbonEmissionScope, period string, startTime, endTime time.Time) (float64, error) {
	args := m.Called(ctx, targetID, scope, period, startTime, endTime)
	return args.Get(0).(float64), args.Error(1)
}

func (m *mockCERepo) CreateSummary(ctx context.Context, summary *entity.CarbonEmissionSummary) error {
	return m.Called(ctx, summary).Error(0)
}

func (m *mockCERepo) UpdateSummary(ctx context.Context, summary *entity.CarbonEmissionSummary) error {
	return m.Called(ctx, summary).Error(0)
}

func (m *mockCERepo) GetSummaryByID(ctx context.Context, id string) (*entity.CarbonEmissionSummary, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CarbonEmissionSummary), args.Error(1)
}

func (m *mockCERepo) GetLatestSummary(ctx context.Context, targetID string, period string) (*entity.CarbonEmissionSummary, error) {
	args := m.Called(ctx, targetID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CarbonEmissionSummary), args.Error(1)
}

func (m *mockCERepo) ListSummaries(ctx context.Context, query *repository.CarbonEmissionSummaryQuery) ([]*entity.CarbonEmissionSummary, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.CarbonEmissionSummary), args.Get(1).(int64), args.Error(2)
}

func (m *mockCERepo) CreateTarget(ctx context.Context, target *entity.CarbonReductionTarget) error {
	return m.Called(ctx, target).Error(0)
}

func (m *mockCERepo) UpdateTarget(ctx context.Context, target *entity.CarbonReductionTarget) error {
	return m.Called(ctx, target).Error(0)
}

func (m *mockCERepo) GetTargetByID(ctx context.Context, id string) (*entity.CarbonReductionTarget, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CarbonReductionTarget), args.Error(1)
}

func (m *mockCERepo) ListTargets(ctx context.Context, query *repository.CarbonReductionTargetQuery) ([]*entity.CarbonReductionTarget, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.CarbonReductionTarget), args.Get(1).(int64), args.Error(2)
}

func (m *mockCERepo) GetActiveTargets(ctx context.Context, targetID string) ([]*entity.CarbonReductionTarget, error) {
	args := m.Called(ctx, targetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CarbonReductionTarget), args.Error(1)
}

var _ repository.CarbonEmissionRepository = (*mockCERepo)(nil)

func TestCarbonEmissionService_CreateFactor_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("GetFactorByCode", mock.Anything, "CEF001").Return(nil, assert.AnError)
	repo.On("CreateFactor", mock.Anything, mock.AnythingOfType("*entity.CarbonEmissionFactor")).Return(nil)

	req := &CreateCarbonEmissionFactorRequest{
		Name: "Grid Emission", Code: "CEF001", Scope: entity.CarbonEmissionScope2,
		Source: "IEA", Value: 0.5, Unit: "kgCO2/kWh", Version: "1.0",
		EffectiveAt: time.Now(),
	}
	factor, err := svc.CreateFactor(context.Background(), req, "admin")
	assert.NoError(t, err)
	assert.NotNil(t, factor)
	assert.Equal(t, "CEF001", factor.Code)
}

func TestCarbonEmissionService_CreateFactor_DuplicateCode(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("GetFactorByCode", mock.Anything, "CEF001").Return(&entity.CarbonEmissionFactor{Code: "CEF001"}, nil)

	req := &CreateCarbonEmissionFactorRequest{
		Name: "Grid Emission", Code: "CEF001", Scope: entity.CarbonEmissionScope2,
		Source: "IEA", Value: 0.5, Unit: "kgCO2/kWh", Version: "1.0",
		EffectiveAt: time.Now(),
	}
	factor, err := svc.CreateFactor(context.Background(), req, "admin")
	assert.Error(t, err)
	assert.Nil(t, factor)
}

func TestCarbonEmissionService_UpdateFactor_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	existing := &entity.CarbonEmissionFactor{ID: "f-001", Name: "Old Name", Value: 0.5}
	repo.On("GetFactorByID", mock.Anything, "f-001").Return(existing, nil)
	repo.On("UpdateFactor", mock.Anything, mock.AnythingOfType("*entity.CarbonEmissionFactor")).Return(nil)

	newName := "New Name"
	newValue := 0.6
	req := &UpdateCarbonEmissionFactorRequest{Name: &newName, Value: &newValue}
	factor, err := svc.UpdateFactor(context.Background(), "f-001", req, "admin")
	assert.NoError(t, err)
	assert.Equal(t, "New Name", factor.Name)
	assert.Equal(t, 0.6, factor.Value)
}

func TestCarbonEmissionService_UpdateFactor_NotFound(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("GetFactorByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	req := &UpdateCarbonEmissionFactorRequest{}
	factor, err := svc.UpdateFactor(context.Background(), "nonexistent", req, "admin")
	assert.Error(t, err)
	assert.Nil(t, factor)
}

func TestCarbonEmissionService_GetFactor_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("GetFactorByID", mock.Anything, "f-001").Return(&entity.CarbonEmissionFactor{ID: "f-001"}, nil)

	factor, err := svc.GetFactor(context.Background(), "f-001")
	assert.NoError(t, err)
	assert.NotNil(t, factor)
}

func TestCarbonEmissionService_ListFactors_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("ListFactors", mock.Anything, mock.AnythingOfType("*repository.CarbonEmissionFactorQuery")).Return([]*entity.CarbonEmissionFactor{{ID: "f-001"}}, int64(1), nil)

	factors, total, err := svc.ListFactors(context.Background(), &QueryCarbonEmissionFactorsRequest{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, factors, 1)
}

func TestCarbonEmissionService_GetActiveFactors_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	scope := entity.CarbonEmissionScope2
	repo.On("GetActiveFactors", mock.Anything, &scope).Return([]*entity.CarbonEmissionFactor{{ID: "f-001"}}, nil)

	factors, err := svc.GetActiveFactors(context.Background(), &scope)
	assert.NoError(t, err)
	assert.Len(t, factors, 1)
}

func TestCarbonEmissionService_CreateRecord_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("CreateRecord", mock.Anything, mock.AnythingOfType("*entity.CarbonEmissionRecord")).Return(nil)

	req := &CreateCarbonEmissionRecordRequest{
		RecordTime: time.Now(), Scope: entity.CarbonEmissionScope1,
		TargetID: "station-001", TargetName: "Station 1",
		FactorCode: "CEF001", FactorValue: 0.5,
		ActivityData: 1000, ActivityUnit: "kWh", Period: "monthly",
	}
	record, err := svc.CreateRecord(context.Background(), req, "admin")
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestCarbonEmissionService_BatchCreateRecords_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("BatchCreateRecords", mock.Anything, mock.Anything).Return(nil)

	reqs := []*CreateCarbonEmissionRecordRequest{
		{RecordTime: time.Now(), Scope: entity.CarbonEmissionScope1, TargetID: "s1", TargetName: "S1",
			FactorCode: "CEF001", FactorValue: 0.5, ActivityData: 1000, ActivityUnit: "kWh", Period: "monthly"},
	}
	records, err := svc.BatchCreateRecords(context.Background(), reqs, "admin")
	assert.NoError(t, err)
	assert.Len(t, records, 1)
}

func TestCarbonEmissionService_GetRecord_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("GetRecordByID", mock.Anything, "r-001").Return(&entity.CarbonEmissionRecord{ID: "r-001"}, nil)

	record, err := svc.GetRecord(context.Background(), "r-001")
	assert.NoError(t, err)
	assert.NotNil(t, record)
}

func TestCarbonEmissionService_ListRecords_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("ListRecords", mock.Anything, mock.AnythingOfType("*repository.CarbonEmissionRecordQuery")).Return([]*entity.CarbonEmissionRecord{{ID: "r-001"}}, int64(1), nil)

	records, total, err := svc.ListRecords(context.Background(), &QueryCarbonEmissionRecordsRequest{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, records, 1)
}

func TestCarbonEmissionService_GetTrendData_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	now := time.Now()
	records := []*entity.CarbonEmissionRecord{
		{RecordTime: now, Scope: entity.CarbonEmissionScope1, EmissionValue: 100},
		{RecordTime: now, Scope: entity.CarbonEmissionScope2, EmissionValue: 200},
	}
	repo.On("GetRecordsByTimeRange", mock.Anything, "station-001", (*entity.CarbonEmissionScope)(nil), mock.Anything, mock.Anything).Return(records, nil)

	trend, err := svc.GetTrendData(context.Background(), "station-001", "monthly", now.AddDate(0, -1, 0), now)
	assert.NoError(t, err)
	assert.Len(t, trend, 1)
	assert.Equal(t, 300.0, trend[0].TotalEmission)
}

func TestCarbonEmissionService_CreateSummary_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	now := time.Now()
	repo.On("GetTotalEmissionByScope", mock.Anything, "station-001", entity.CarbonEmissionScope1, "monthly", mock.Anything, mock.Anything).Return(100.0, nil)
	repo.On("GetTotalEmissionByScope", mock.Anything, "station-001", entity.CarbonEmissionScope2, "monthly", mock.Anything, mock.Anything).Return(200.0, nil)
	repo.On("GetTotalEmissionByScope", mock.Anything, "station-001", entity.CarbonEmissionScope3, "monthly", mock.Anything, mock.Anything).Return(50.0, nil)
	repo.On("GetLatestSummary", mock.Anything, "station-001", "monthly").Return(nil, assert.AnError)
	repo.On("CreateSummary", mock.Anything, mock.AnythingOfType("*entity.CarbonEmissionSummary")).Return(nil)

	summary, err := svc.CreateSummary(context.Background(), "station-001", "Station 1", "monthly", now.AddDate(0, -1, 0), now, "admin")
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 350.0, summary.TotalEmission)
}

func TestCarbonEmissionService_GetSummary_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("GetSummaryByID", mock.Anything, "s-001").Return(&entity.CarbonEmissionSummary{ID: "s-001"}, nil)

	summary, err := svc.GetSummary(context.Background(), "s-001")
	assert.NoError(t, err)
	assert.NotNil(t, summary)
}

func TestCarbonEmissionService_ListSummaries_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("ListSummaries", mock.Anything, mock.AnythingOfType("*repository.CarbonEmissionSummaryQuery")).Return([]*entity.CarbonEmissionSummary{{ID: "s-001"}}, int64(1), nil)

	summaries, total, err := svc.ListSummaries(context.Background(), &QueryCarbonEmissionSummariesRequest{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, summaries, 1)
}

func TestCarbonEmissionService_CreateTarget_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("CreateTarget", mock.Anything, mock.AnythingOfType("*entity.CarbonReductionTarget")).Return(nil)

	req := &CreateCarbonReductionTargetRequest{
		Name: "30% Reduction", TargetID: "station-001", TargetName: "Station 1",
		BaseYear: 2023, BaseEmission: 1000, TargetYear: 2030, TargetReduction: 30,
		StartDate: time.Now(), EndDate: time.Now().AddDate(7, 0, 0),
	}
	target, err := svc.CreateTarget(context.Background(), req, "admin")
	assert.NoError(t, err)
	assert.NotNil(t, target)
	assert.Equal(t, 700.0, target.TargetEmission)
}

func TestCarbonEmissionService_UpdateTarget_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	existing := &entity.CarbonReductionTarget{ID: "t-001", Name: "Old Target"}
	repo.On("GetTargetByID", mock.Anything, "t-001").Return(existing, nil)
	repo.On("UpdateTarget", mock.Anything, mock.AnythingOfType("*entity.CarbonReductionTarget")).Return(nil)

	newName := "New Target"
	req := &UpdateCarbonReductionTargetRequest{Name: &newName}
	target, err := svc.UpdateTarget(context.Background(), "t-001", req, "admin")
	assert.NoError(t, err)
	assert.Equal(t, "New Target", target.Name)
}

func TestCarbonEmissionService_GetTarget_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("GetTargetByID", mock.Anything, "t-001").Return(&entity.CarbonReductionTarget{ID: "t-001"}, nil)

	target, err := svc.GetTarget(context.Background(), "t-001")
	assert.NoError(t, err)
	assert.NotNil(t, target)
}

func TestCarbonEmissionService_ListTargets_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("ListTargets", mock.Anything, mock.AnythingOfType("*repository.CarbonReductionTargetQuery")).Return([]*entity.CarbonReductionTarget{{ID: "t-001"}}, int64(1), nil)

	targets, total, err := svc.ListTargets(context.Background(), &QueryCarbonReductionTargetsRequest{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, targets, 1)
}

func TestCarbonEmissionService_GetActiveTargets_Success(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("GetActiveTargets", mock.Anything, "station-001").Return([]*entity.CarbonReductionTarget{{ID: "t-001"}}, nil)

	targets, err := svc.GetActiveTargets(context.Background(), "station-001")
	assert.NoError(t, err)
	assert.Len(t, targets, 1)
}

func TestNewCarbonEmissionService(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)
	assert.NotNil(t, svc)
}
