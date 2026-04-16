package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// CostEntryService 成本条目服务
type CostEntryService struct {
	costEntryRepo   repository.CostEntryRepository
	costCategoryRepo repository.CostCategoryRepository
}

// NewCostEntryService 创建成本条目服务实例
func NewCostEntryService(
	costEntryRepo repository.CostEntryRepository,
	costCategoryRepo repository.CostCategoryRepository,
) *CostEntryService {
	return &CostEntryService{
		costEntryRepo:   costEntryRepo,
		costCategoryRepo: costCategoryRepo,
	}
}

// CreateCostEntry 创建成本条目
func (s *CostEntryService) CreateCostEntry(ctx context.Context, entry *entity.CostEntry) error {
	// 验证成本类别是否存在
	_, err := s.costCategoryRepo.GetByID(ctx, entry.CostCategoryID)
	if err != nil {
		return errors.New("成本类别不存在")
	}

	// 生成成本条目编码
	entry.Code = fmt.Sprintf("CE%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
	entry.ApprovalStatus = "pending"
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	return s.costEntryRepo.Create(ctx, entry)
}

// UpdateCostEntry 更新成本条目
func (s *CostEntryService) UpdateCostEntry(ctx context.Context, entry *entity.CostEntry) error {
	// 验证成本条目是否存在
	existing, err := s.costEntryRepo.GetByID(ctx, entry.ID)
	if err != nil {
		return errors.New("成本条目不存在")
	}

	// 验证成本类别是否存在
	if entry.CostCategoryID != existing.CostCategoryID {
		_, err := s.costCategoryRepo.GetByID(ctx, entry.CostCategoryID)
		if err != nil {
			return errors.New("成本类别不存在")
		}
	}

	// 已审批的成本条目不能修改
	if existing.ApprovalStatus == "approved" {
		return errors.New("已审批的成本条目不能修改")
	}

	entry.UpdatedAt = time.Now()

	return s.costEntryRepo.Update(ctx, entry)
}

// DeleteCostEntry 删除成本条目
func (s *CostEntryService) DeleteCostEntry(ctx context.Context, id string) error {
	// 验证成本条目是否存在
	existing, err := s.costEntryRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("成本条目不存在")
	}

	// 已审批的成本条目不能删除
	if existing.ApprovalStatus == "approved" {
		return errors.New("已审批的成本条目不能删除")
	}

	return s.costEntryRepo.Delete(ctx, id)
}

// GetCostEntryByID 根据ID获取成本条目
func (s *CostEntryService) GetCostEntryByID(ctx context.Context, id string) (*entity.CostEntry, error) {
	return s.costEntryRepo.GetByID(ctx, id)
}

// GetCostEntryByCode 根据编码获取成本条目
func (s *CostEntryService) GetCostEntryByCode(ctx context.Context, code string) (*entity.CostEntry, error) {
	return s.costEntryRepo.GetByCode(ctx, code)
}

// ListCostEntries 列出成本条目
func (s *CostEntryService) ListCostEntries(ctx context.Context, categoryID *string, startDate, endDate *time.Time, status *string, page, pageSize int) ([]*entity.CostEntry, int64, error) {
	offset := (page - 1) * pageSize
	return s.costEntryRepo.List(ctx, categoryID, startDate, endDate, status, offset, pageSize)
}

// GetTotalByCategory 根据成本类别获取总金额
func (s *CostEntryService) GetTotalByCategory(ctx context.Context, categoryID string, startDate, endDate *time.Time) (float64, error) {
	// 验证成本类别是否存在
	_, err := s.costCategoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return 0, errors.New("成本类别不存在")
	}

	return s.costEntryRepo.GetTotalByCategory(ctx, categoryID, startDate, endDate)
}

// GetTotalByPeriod 根据时间段获取总金额
func (s *CostEntryService) GetTotalByPeriod(ctx context.Context, startDate, endDate *time.Time) (float64, error) {
	return s.costEntryRepo.GetTotalByPeriod(ctx, startDate, endDate)
}

// ApproveCostEntry 审批成本条目
func (s *CostEntryService) ApproveCostEntry(ctx context.Context, id, approvedBy string) error {
	// 验证成本条目是否存在
	entry, err := s.costEntryRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("成本条目不存在")
	}

	// 验证成本条目状态
	if entry.ApprovalStatus != "pending" {
		return errors.New("成本条目状态不是待审批")
	}

	// 更新审批状态
	entry.ApprovalStatus = "approved"
	entry.ApprovedBy = &approvedBy
	now := time.Now()
	entry.ApprovedAt = &now
	entry.UpdatedAt = now

	return s.costEntryRepo.Update(ctx, entry)
}

// RejectCostEntry 拒绝成本条目
func (s *CostEntryService) RejectCostEntry(ctx context.Context, id, approvedBy string) error {
	// 验证成本条目是否存在
	entry, err := s.costEntryRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("成本条目不存在")
	}

	// 验证成本条目状态
	if entry.ApprovalStatus != "pending" {
		return errors.New("成本条目状态不是待审批")
	}

	// 更新审批状态
	entry.ApprovalStatus = "rejected"
	entry.ApprovedBy = &approvedBy
	now := time.Now()
	entry.ApprovedAt = &now
	entry.UpdatedAt = now

	return s.costEntryRepo.Update(ctx, entry)
}
