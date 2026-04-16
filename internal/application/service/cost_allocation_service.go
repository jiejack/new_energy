package service

import (
	"context"
	"errors"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// CostAllocationService 成本分配服务
type CostAllocationService struct {
	costAllocationRepo repository.CostAllocationRepository
	costEntryRepo     repository.CostEntryRepository
}

// NewCostAllocationService 创建成本分配服务实例
func NewCostAllocationService(
	costAllocationRepo repository.CostAllocationRepository,
	costEntryRepo repository.CostEntryRepository,
) *CostAllocationService {
	return &CostAllocationService{
		costAllocationRepo: costAllocationRepo,
		costEntryRepo:     costEntryRepo,
	}
}

// CreateCostAllocation 创建成本分配
func (s *CostAllocationService) CreateCostAllocation(ctx context.Context, allocation *entity.CostAllocation) error {
	// 验证成本条目是否存在
	costEntry, err := s.costEntryRepo.GetByID(ctx, allocation.CostEntryID)
	if err != nil {
		return errors.New("成本条目不存在")
	}

	// 验证成本条目是否已审批
	if costEntry.ApprovalStatus != "approved" {
		return errors.New("成本条目未审批，不能进行分配")
	}

	// 验证分配金额是否合法
	if allocation.Amount <= 0 {
		return errors.New("分配金额必须大于0")
	}

	// 验证分配百分比是否合法
	if allocation.Percentage <= 0 || allocation.Percentage > 100 {
		return errors.New("分配百分比必须在0-100之间")
	}

	// 检查分配金额是否超过成本条目金额
	if allocation.Amount > costEntry.Amount {
		return errors.New("分配金额不能超过成本条目金额")
	}

	// 检查已分配金额是否超过成本条目金额
	existingAllocations, err := s.costAllocationRepo.ListByCostEntryID(ctx, allocation.CostEntryID)
	if err != nil {
		return err
	}

	var totalAllocated float64
	for _, existing := range existingAllocations {
		totalAllocated += existing.Amount
	}

	if totalAllocated+allocation.Amount > costEntry.Amount {
		return errors.New("累计分配金额不能超过成本条目金额")
	}

	allocation.CreatedAt = time.Now()
	allocation.UpdatedAt = time.Now()

	return s.costAllocationRepo.Create(ctx, allocation)
}

// UpdateCostAllocation 更新成本分配
func (s *CostAllocationService) UpdateCostAllocation(ctx context.Context, allocation *entity.CostAllocation) error {
	// 验证成本分配是否存在
	_, err := s.costAllocationRepo.GetByID(ctx, allocation.ID)
	if err != nil {
		return errors.New("成本分配不存在")
	}

	// 验证成本条目是否存在
	costEntry, err := s.costEntryRepo.GetByID(ctx, allocation.CostEntryID)
	if err != nil {
		return errors.New("成本条目不存在")
	}

	// 验证成本条目是否已审批
	if costEntry.ApprovalStatus != "approved" {
		return errors.New("成本条目未审批，不能修改分配")
	}

	// 验证分配金额是否合法
	if allocation.Amount <= 0 {
		return errors.New("分配金额必须大于0")
	}

	// 验证分配百分比是否合法
	if allocation.Percentage <= 0 || allocation.Percentage > 100 {
		return errors.New("分配百分比必须在0-100之间")
	}

	// 检查分配金额是否超过成本条目金额
	if allocation.Amount > costEntry.Amount {
		return errors.New("分配金额不能超过成本条目金额")
	}

	// 检查已分配金额是否超过成本条目金额
	existingAllocations, err := s.costAllocationRepo.ListByCostEntryID(ctx, allocation.CostEntryID)
	if err != nil {
		return err
	}

	var totalAllocated float64
	for _, alloc := range existingAllocations {
		if alloc.ID != allocation.ID {
			totalAllocated += alloc.Amount
		}
	}

	if totalAllocated+allocation.Amount > costEntry.Amount {
		return errors.New("累计分配金额不能超过成本条目金额")
	}

	allocation.UpdatedAt = time.Now()

	return s.costAllocationRepo.Update(ctx, allocation)
}

// DeleteCostAllocation 删除成本分配
func (s *CostAllocationService) DeleteCostAllocation(ctx context.Context, id string) error {
	// 验证成本分配是否存在
	_, err := s.costAllocationRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("成本分配不存在")
	}

	return s.costAllocationRepo.Delete(ctx, id)
}

// GetCostAllocationByID 根据ID获取成本分配
func (s *CostAllocationService) GetCostAllocationByID(ctx context.Context, id string) (*entity.CostAllocation, error) {
	return s.costAllocationRepo.GetByID(ctx, id)
}

// ListCostAllocationsByCostEntryID 根据成本条目ID列出成本分配
func (s *CostAllocationService) ListCostAllocationsByCostEntryID(ctx context.Context, costEntryID string) ([]*entity.CostAllocation, error) {
	// 验证成本条目是否存在
	_, err := s.costEntryRepo.GetByID(ctx, costEntryID)
	if err != nil {
		return nil, errors.New("成本条目不存在")
	}

	return s.costAllocationRepo.ListByCostEntryID(ctx, costEntryID)
}

// ListCostAllocationsByAllocated 根据分配对象列出成本分配
func (s *CostAllocationService) ListCostAllocationsByAllocated(ctx context.Context, allocatedTo, allocatedID string) ([]*entity.CostAllocation, error) {
	return s.costAllocationRepo.ListByAllocated(ctx, allocatedTo, allocatedID)
}

// GetTotalByAllocated 根据分配对象获取总金额
func (s *CostAllocationService) GetTotalByAllocated(ctx context.Context, allocatedTo, allocatedID string, startDate, endDate *time.Time) (float64, error) {
	return s.costAllocationRepo.GetTotalByAllocated(ctx, allocatedTo, allocatedID, startDate, endDate)
}
