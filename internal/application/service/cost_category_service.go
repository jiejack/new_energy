package service

import (
	"context"
	"errors"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// CostCategoryService 成本类别服务
type CostCategoryService struct {
	costCategoryRepo repository.CostCategoryRepository
}

// NewCostCategoryService 创建成本类别服务实例
func NewCostCategoryService(
	costCategoryRepo repository.CostCategoryRepository,
) *CostCategoryService {
	return &CostCategoryService{
		costCategoryRepo: costCategoryRepo,
	}
}

// CreateCostCategory 创建成本类别
func (s *CostCategoryService) CreateCostCategory(ctx context.Context, category *entity.CostCategory) error {
	// 验证成本类别编码是否已存在
	existing, err := s.costCategoryRepo.GetByCode(ctx, category.Code)
	if err == nil && existing != nil {
		return errors.New("成本类别编码已存在")
	}

	// 验证父类别是否存在
	if category.ParentID != nil {
		_, err := s.costCategoryRepo.GetByID(ctx, *category.ParentID)
		if err != nil {
			return errors.New("父类别不存在")
		}
	}

	return s.costCategoryRepo.Create(ctx, category)
}

// UpdateCostCategory 更新成本类别
func (s *CostCategoryService) UpdateCostCategory(ctx context.Context, category *entity.CostCategory) error {
	// 验证成本类别是否存在
	existing, err := s.costCategoryRepo.GetByID(ctx, category.ID)
	if err != nil {
		return errors.New("成本类别不存在")
	}

	// 验证编码是否已被其他成本类别使用
	if category.Code != existing.Code {
		existingByCode, err := s.costCategoryRepo.GetByCode(ctx, category.Code)
		if err == nil && existingByCode != nil && existingByCode.ID != category.ID {
			return errors.New("成本类别编码已被使用")
		}
	}

	// 验证父类别是否存在
	if category.ParentID != nil {
		_, err := s.costCategoryRepo.GetByID(ctx, *category.ParentID)
		if err != nil {
			return errors.New("父类别不存在")
		}

		// 防止循环引用
		if *category.ParentID == category.ID {
			return errors.New("父类别不能是自身")
		}
	}

	return s.costCategoryRepo.Update(ctx, category)
}

// DeleteCostCategory 删除成本类别
func (s *CostCategoryService) DeleteCostCategory(ctx context.Context, id string) error {
	// 验证成本类别是否存在
	_, err := s.costCategoryRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("成本类别不存在")
	}

	// 检查是否有子类别
	children, err := s.costCategoryRepo.List(ctx, &id, nil)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return errors.New("该成本类别下有子类别，无法删除")
	}

	return s.costCategoryRepo.Delete(ctx, id)
}

// GetCostCategoryByID 根据ID获取成本类别
func (s *CostCategoryService) GetCostCategoryByID(ctx context.Context, id string) (*entity.CostCategory, error) {
	return s.costCategoryRepo.GetByID(ctx, id)
}

// GetCostCategoryByCode 根据编码获取成本类别
func (s *CostCategoryService) GetCostCategoryByCode(ctx context.Context, code string) (*entity.CostCategory, error) {
	return s.costCategoryRepo.GetByCode(ctx, code)
}

// ListCostCategories 列出成本类别
func (s *CostCategoryService) ListCostCategories(ctx context.Context, parentID *string, status *string) ([]*entity.CostCategory, error) {
	return s.costCategoryRepo.List(ctx, parentID, status)
}

// GetCostCategoryTree 获取成本类别树
func (s *CostCategoryService) GetCostCategoryTree(ctx context.Context) ([]*entity.CostCategory, error) {
	return s.costCategoryRepo.GetTree(ctx)
}
