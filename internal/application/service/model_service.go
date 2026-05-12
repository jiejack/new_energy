package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type ModelService interface {
	RegisterModel(ctx context.Context, modelName, version, artifactPath string, accuracy float64) (*entity.ModelVersion, error)
	GetModel(ctx context.Context, id string) (*entity.ModelVersion, error)
	GetProductionModel(ctx context.Context, modelName string) (*entity.ModelVersion, error)
	ListModels(ctx context.Context, modelName string) ([]*entity.ModelVersion, error)
	PromoteToProduction(ctx context.Context, id string) error
	RetireModel(ctx context.Context, id string) error
}

type modelService struct {
	repo repository.ModelVersionRepository
}

func NewModelService(repo repository.ModelVersionRepository) ModelService {
	return &modelService{repo: repo}
}

func (s *modelService) RegisterModel(ctx context.Context, modelName, version, artifactPath string, accuracy float64) (*entity.ModelVersion, error) {
	model := &entity.ModelVersion{
		ID:           uuid.New().String(),
		ModelName:    modelName,
		Version:      version,
		Status:       entity.ModelStatusStaging,
		Accuracy:     &accuracy,
		ArtifactPath: artifactPath,
	}
	if err := s.repo.Create(ctx, model); err != nil {
		return nil, fmt.Errorf("register model failed: %w", err)
	}
	return model, nil
}

func (s *modelService) GetModel(ctx context.Context, id string) (*entity.ModelVersion, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *modelService) GetProductionModel(ctx context.Context, modelName string) (*entity.ModelVersion, error) {
	return s.repo.GetProductionModel(ctx, modelName)
}

func (s *modelService) ListModels(ctx context.Context, modelName string) ([]*entity.ModelVersion, error) {
	return s.repo.ListByModel(ctx, modelName)
}

func (s *modelService) PromoteToProduction(ctx context.Context, id string) error {
	model, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("model not found: %w", err)
	}
	current, err := s.repo.GetProductionModel(ctx, model.ModelName)
	if err == nil && current != nil {
		if err := s.repo.UpdateStatus(ctx, current.ID, entity.ModelStatusRetired); err != nil {
			return fmt.Errorf("retire current production model failed: %w", err)
		}
	}
	now := time.Now()
	model.DeployedAt = &now
	model.Status = entity.ModelStatusProd
	return s.repo.UpdateStatus(ctx, id, entity.ModelStatusProd)
}

func (s *modelService) RetireModel(ctx context.Context, id string) error {
	return s.repo.UpdateStatus(ctx, id, entity.ModelStatusRetired)
}
