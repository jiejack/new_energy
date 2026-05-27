package service

import (
	"context"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockModelVersionRepo struct {
	mock.Mock
}

func (m *mockModelVersionRepo) Create(ctx context.Context, model *entity.ModelVersion) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}
func (m *mockModelVersionRepo) Update(ctx context.Context, model *entity.ModelVersion) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}
func (m *mockModelVersionRepo) UpdateStatus(ctx context.Context, id string, status entity.ModelStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}
func (m *mockModelVersionRepo) GetByID(ctx context.Context, id string) (*entity.ModelVersion, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ModelVersion), args.Error(1)
}
func (m *mockModelVersionRepo) GetProductionModel(ctx context.Context, modelName string) (*entity.ModelVersion, error) {
	args := m.Called(ctx, modelName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ModelVersion), args.Error(1)
}
func (m *mockModelVersionRepo) ListByModel(ctx context.Context, modelName string) ([]*entity.ModelVersion, error) {
	args := m.Called(ctx, modelName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ModelVersion), args.Error(1)
}

func TestModelService_RegisterModel_Success(t *testing.T) {
	repo := new(mockModelVersionRepo)
	svc := NewModelService(repo)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.ModelVersion")).Return(nil)
	model, err := svc.RegisterModel(context.Background(), "solar_v2", "1.0.0", "/path/to/model", 0.95)
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, "solar_v2", model.ModelName)
	assert.Equal(t, entity.ModelStatusStaging, model.Status)
}

func TestModelService_RegisterModel_Error(t *testing.T) {
	repo := new(mockModelVersionRepo)
	svc := NewModelService(repo)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.ModelVersion")).Return(assert.AnError)
	model, err := svc.RegisterModel(context.Background(), "solar_v2", "1.0.0", "/path/to/model", 0.95)
	assert.Error(t, err)
	assert.Nil(t, model)
}

func TestModelService_GetModel_Success(t *testing.T) {
	repo := new(mockModelVersionRepo)
	svc := NewModelService(repo)
	accuracy := 0.95
	model := &entity.ModelVersion{ID: "model-001", ModelName: "solar_v2", Accuracy: &accuracy}
	repo.On("GetByID", mock.Anything, "model-001").Return(model, nil)
	result, err := svc.GetModel(context.Background(), "model-001")
	assert.NoError(t, err)
	assert.Equal(t, "solar_v2", result.ModelName)
}

func TestModelService_GetProductionModel_Success(t *testing.T) {
	repo := new(mockModelVersionRepo)
	svc := NewModelService(repo)
	accuracy := 0.95
	model := &entity.ModelVersion{ID: "model-001", ModelName: "solar_v2", Status: entity.ModelStatusProd, Accuracy: &accuracy}
	repo.On("GetProductionModel", mock.Anything, "solar_v2").Return(model, nil)
	result, err := svc.GetProductionModel(context.Background(), "solar_v2")
	assert.NoError(t, err)
	assert.Equal(t, entity.ModelStatusProd, result.Status)
}

func TestModelService_ListModels_Success(t *testing.T) {
	repo := new(mockModelVersionRepo)
	svc := NewModelService(repo)
	models := []*entity.ModelVersion{{ID: "model-001", ModelName: "solar_v2"}}
	repo.On("ListByModel", mock.Anything, "solar_v2").Return(models, nil)
	result, err := svc.ListModels(context.Background(), "solar_v2")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestModelService_PromoteToProduction_Success(t *testing.T) {
	repo := new(mockModelVersionRepo)
	svc := NewModelService(repo)
	accuracy := 0.95
	model := &entity.ModelVersion{ID: "model-001", ModelName: "solar_v2", Accuracy: &accuracy}
	repo.On("GetByID", mock.Anything, "model-001").Return(model, nil)
	repo.On("GetProductionModel", mock.Anything, "solar_v2").Return(nil, assert.AnError)
	repo.On("UpdateStatus", mock.Anything, "model-001", entity.ModelStatusProd).Return(nil)
	err := svc.PromoteToProduction(context.Background(), "model-001")
	assert.NoError(t, err)
}

func TestModelService_PromoteToProduction_WithExistingProd(t *testing.T) {
	repo := new(mockModelVersionRepo)
	svc := NewModelService(repo)
	accuracy := 0.95
	model := &entity.ModelVersion{ID: "model-002", ModelName: "solar_v2", Accuracy: &accuracy}
	oldProd := &entity.ModelVersion{ID: "model-001", ModelName: "solar_v2", Status: entity.ModelStatusProd}
	repo.On("GetByID", mock.Anything, "model-002").Return(model, nil)
	repo.On("GetProductionModel", mock.Anything, "solar_v2").Return(oldProd, nil)
	repo.On("UpdateStatus", mock.Anything, "model-001", entity.ModelStatusRetired).Return(nil)
	repo.On("UpdateStatus", mock.Anything, "model-002", entity.ModelStatusProd).Return(nil)
	err := svc.PromoteToProduction(context.Background(), "model-002")
	assert.NoError(t, err)
}

func TestModelService_PromoteToProduction_NotFound(t *testing.T) {
	repo := new(mockModelVersionRepo)
	svc := NewModelService(repo)
	repo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)
	err := svc.PromoteToProduction(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestModelService_RetireModel_Success(t *testing.T) {
	repo := new(mockModelVersionRepo)
	svc := NewModelService(repo)
	repo.On("UpdateStatus", mock.Anything, "model-001", entity.ModelStatusRetired).Return(nil)
	err := svc.RetireModel(context.Background(), "model-001")
	assert.NoError(t, err)
}

func TestSeverityToPriority(t *testing.T) {
	assert.Equal(t, "urgent", severityToPriority(entity.FaultSeverityFatal))
	assert.Equal(t, "high", severityToPriority(entity.FaultSeverityCritical))
	assert.Equal(t, "medium", severityToPriority(entity.FaultSeverityWarning))
	assert.Equal(t, "low", severityToPriority(entity.FaultSeverityInfo))
}

var _ repository.ModelVersionRepository = (*mockModelVersionRepo)(nil)
