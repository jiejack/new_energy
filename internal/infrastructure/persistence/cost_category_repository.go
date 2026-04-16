package persistence

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

var _ repository.CostCategoryRepository = (*CostCategoryRepository)(nil)

// CostCategoryRepository 成本类别仓储实现
type CostCategoryRepository struct {
	db *Database
}

// NewCostCategoryRepository 创建成本类别仓储实例
func NewCostCategoryRepository(db *Database) repository.CostCategoryRepository {
	return &CostCategoryRepository{db: db}
}

// Create 创建成本类别
func (r *CostCategoryRepository) Create(ctx context.Context, category *entity.CostCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// Update 更新成本类别
func (r *CostCategoryRepository) Update(ctx context.Context, category *entity.CostCategory) error {
	return r.db.WithContext(ctx).Save(category).Error
}

// Delete 删除成本类别
func (r *CostCategoryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.CostCategory{}, "id = ?", id).Error
}

// GetByID 根据ID获取成本类别
func (r *CostCategoryRepository) GetByID(ctx context.Context, id string) (*entity.CostCategory, error) {
	var category entity.CostCategory
	err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetByCode 根据编码获取成本类别
func (r *CostCategoryRepository) GetByCode(ctx context.Context, code string) (*entity.CostCategory, error) {
	var category entity.CostCategory
	err := r.db.WithContext(ctx).First(&category, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// List 列出成本类别
func (r *CostCategoryRepository) List(ctx context.Context, parentID *string, status *string) ([]*entity.CostCategory, error) {
	var categories []*entity.CostCategory
	query := r.db.WithContext(ctx)

	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Find(&categories).Error
	return categories, err
}

// GetTree 获取成本类别树
func (r *CostCategoryRepository) GetTree(ctx context.Context) ([]*entity.CostCategory, error) {
	var categories []*entity.CostCategory
	err := r.db.WithContext(ctx).Where("parent_id IS NULL").Find(&categories).Error
	if err != nil {
		return nil, err
	}

	// 递归获取子类别
	for _, category := range categories {
		if err := r.loadChildren(ctx, category); err != nil {
			return nil, err
		}
	}

	return categories, nil
}

// loadChildren 递归加载子类别
func (r *CostCategoryRepository) loadChildren(ctx context.Context, parent *entity.CostCategory) error {
	var children []*entity.CostCategory
	err := r.db.WithContext(ctx).Where("parent_id = ?", parent.ID).Find(&children).Error
	if err != nil {
		return err
	}

	for _, child := range children {
		if err := r.loadChildren(ctx, child); err != nil {
			return err
		}
	}

	return nil
}
