package service

import (
	"context"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCarbonEmissionService_UpdateFactor_AllFields(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	existing := &entity.CarbonEmissionFactor{ID: "f-001", Name: "Old", Scope: entity.CarbonEmissionScope1, Source: "Old", Value: 0.5, Unit: "kg", Description: "desc", Version: "1.0"}
	repo.On("GetFactorByID", mock.Anything, "f-001").Return(existing, nil)
	repo.On("UpdateFactor", mock.Anything, mock.AnythingOfType("*entity.CarbonEmissionFactor")).Return(nil)

	name := "New"
	scope := entity.CarbonEmissionScope2
	source := "New"
	value := 0.8
	unit := "t"
	desc := "new desc"
	version := "2.0"
	effectiveAt := time.Now()
	expiresAt := time.Now().AddDate(1, 0, 0)
	isActive := false
	req := &UpdateCarbonEmissionFactorRequest{
		Name: &name, Scope: &scope, Source: &source, Value: &value,
		Unit: &unit, Description: &desc, Version: &version,
		EffectiveAt: &effectiveAt, ExpiresAt: &expiresAt, IsActive: &isActive,
	}
	factor, err := svc.UpdateFactor(context.Background(), "f-001", req, "admin")
	assert.NoError(t, err)
	assert.Equal(t, "New", factor.Name)
	assert.Equal(t, entity.CarbonEmissionScope2, factor.Scope)
	assert.Equal(t, false, factor.IsActive)
}

func TestCarbonEmissionService_CreateSummary_WithIncreasingTrend(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	now := time.Now()
	repo.On("GetTotalEmissionByScope", mock.Anything, "station-001", entity.CarbonEmissionScope1, "monthly", mock.Anything, mock.Anything).Return(100.0, nil)
	repo.On("GetTotalEmissionByScope", mock.Anything, "station-001", entity.CarbonEmissionScope2, "monthly", mock.Anything, mock.Anything).Return(200.0, nil)
	repo.On("GetTotalEmissionByScope", mock.Anything, "station-001", entity.CarbonEmissionScope3, "monthly", mock.Anything, mock.Anything).Return(50.0, nil)
	repo.On("GetLatestSummary", mock.Anything, "station-001", "monthly").Return(&entity.CarbonEmissionSummary{TotalEmission: 300.0}, nil)
	repo.On("CreateSummary", mock.Anything, mock.AnythingOfType("*entity.CarbonEmissionSummary")).Return(nil)

	summary, err := svc.CreateSummary(context.Background(), "station-001", "Station 1", "monthly", now.AddDate(0, -1, 0), now, "admin")
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, "increasing", summary.Trend)
}

func TestCarbonEmissionService_CreateSummary_WithDecreasingTrend(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	now := time.Now()
	repo.On("GetTotalEmissionByScope", mock.Anything, "station-001", entity.CarbonEmissionScope1, "monthly", mock.Anything, mock.Anything).Return(50.0, nil)
	repo.On("GetTotalEmissionByScope", mock.Anything, "station-001", entity.CarbonEmissionScope2, "monthly", mock.Anything, mock.Anything).Return(50.0, nil)
	repo.On("GetTotalEmissionByScope", mock.Anything, "station-001", entity.CarbonEmissionScope3, "monthly", mock.Anything, mock.Anything).Return(50.0, nil)
	repo.On("GetLatestSummary", mock.Anything, "station-001", "monthly").Return(&entity.CarbonEmissionSummary{TotalEmission: 200.0}, nil)
	repo.On("CreateSummary", mock.Anything, mock.AnythingOfType("*entity.CarbonEmissionSummary")).Return(nil)

	summary, err := svc.CreateSummary(context.Background(), "station-001", "Station 1", "monthly", now.AddDate(0, -1, 0), now, "admin")
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, "decreasing", summary.Trend)
}

func TestCarbonEmissionService_UpdateTarget_NotFound(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	repo.On("GetTargetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	req := &UpdateCarbonReductionTargetRequest{}
	target, err := svc.UpdateTarget(context.Background(), "nonexistent", req, "admin")
	assert.Error(t, err)
	assert.Nil(t, target)
}

func TestCarbonEmissionService_UpdateTarget_AllFields(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	existing := &entity.CarbonReductionTarget{ID: "t-001", Name: "Old", Status: "active"}
	repo.On("GetTargetByID", mock.Anything, "t-001").Return(existing, nil)
	repo.On("UpdateTarget", mock.Anything, mock.AnythingOfType("*entity.CarbonReductionTarget")).Return(nil)

	name := "New"
	status := "completed"
	targetEmission := 600.0
	currentProgress := 25.0
	req := &UpdateCarbonReductionTargetRequest{
		Name: &name, Status: &status, TargetEmission: &targetEmission, CurrentProgress: &currentProgress,
	}
	target, err := svc.UpdateTarget(context.Background(), "t-001", req, "admin")
	assert.NoError(t, err)
	assert.Equal(t, "New", target.Name)
	assert.Equal(t, "completed", target.Status)
}

func TestCarbonEmissionService_GetTrendData_WithScope3(t *testing.T) {
	repo := new(mockCERepo)
	svc := NewCarbonEmissionService(repo)

	now := time.Now()
	records := []*entity.CarbonEmissionRecord{
		{RecordTime: now, Scope: entity.CarbonEmissionScope3, EmissionValue: 50},
	}
	repo.On("GetRecordsByTimeRange", mock.Anything, "station-001", (*entity.CarbonEmissionScope)(nil), mock.Anything, mock.Anything).Return(records, nil)

	trend, err := svc.GetTrendData(context.Background(), "station-001", "monthly", now.AddDate(0, -1, 0), now)
	assert.NoError(t, err)
	assert.Len(t, trend, 1)
	assert.Equal(t, 50.0, trend[0].Scope3Emission)
}
