package persistence

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type modelVersionRepository struct {
	db *Database
}

func NewModelVersionRepository(db *Database) repository.ModelVersionRepository {
	return &modelVersionRepository{db: db}
}

func (r *modelVersionRepository) Create(ctx context.Context, model *entity.ModelVersion) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *modelVersionRepository) GetByID(ctx context.Context, id string) (*entity.ModelVersion, error) {
	var model entity.ModelVersion
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *modelVersionRepository) GetProductionModel(ctx context.Context, modelName string) (*entity.ModelVersion, error) {
	var model entity.ModelVersion
	if err := r.db.WithContext(ctx).
		Where("model_name = ? AND status = ?", modelName, entity.ModelStatusProd).
		Order("created_at DESC").
		First(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *modelVersionRepository) ListByModel(ctx context.Context, modelName string) ([]*entity.ModelVersion, error) {
	var models []*entity.ModelVersion
	err := r.db.WithContext(ctx).
		Where("model_name = ?", modelName).
		Order("created_at DESC").
		Find(&models).Error
	return models, err
}

func (r *modelVersionRepository) UpdateStatus(ctx context.Context, id string, status entity.ModelStatus) error {
	return r.db.WithContext(ctx).
		Model(&entity.ModelVersion{}).
		Where("id = ?", id).
		Update("status", status).Error
}
