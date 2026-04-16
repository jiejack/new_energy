package service

import (
	"context"
	"fmt"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// InventoryService 库存服务
type InventoryService struct {
	inventoryRepo        repository.InventoryRepository
	inventoryTransRepo   repository.InventoryTransactionRepository
	supplierRepo         repository.SupplierRepository
}

// NewInventoryService 创建库存服务
func NewInventoryService(
	inventoryRepo repository.InventoryRepository,
	inventoryTransRepo repository.InventoryTransactionRepository,
	supplierRepo repository.SupplierRepository,
) *InventoryService {
	return &InventoryService{
		inventoryRepo:        inventoryRepo,
		inventoryTransRepo:   inventoryTransRepo,
		supplierRepo:         supplierRepo,
	}
}

// InventoryFilter 库存过滤条件
type InventoryFilter struct {
	Type   string
	Status string
	Name   string
	Code   string
}

// CreateInventory 创建库存
func (s *InventoryService) CreateInventory(ctx context.Context, inventory *entity.Inventory) error {
	// 计算总价值
	inventory.TotalValue = inventory.Quantity * inventory.UnitPrice
	return s.inventoryRepo.Create(ctx, inventory)
}

// UpdateInventory 更新库存
func (s *InventoryService) UpdateInventory(ctx context.Context, inventory *entity.Inventory) error {
	// 计算总价值
	inventory.TotalValue = inventory.Quantity * inventory.UnitPrice
	return s.inventoryRepo.Update(ctx, inventory)
}

// DeleteInventory 删除库存
func (s *InventoryService) DeleteInventory(ctx context.Context, id string) error {
	return s.inventoryRepo.Delete(ctx, id)
}

// GetInventoryByID 根据ID获取库存
func (s *InventoryService) GetInventoryByID(ctx context.Context, id string) (*entity.Inventory, error) {
	return s.inventoryRepo.GetByID(ctx, id)
}

// GetInventoryByCode 根据编码获取库存
func (s *InventoryService) GetInventoryByCode(ctx context.Context, code string) (*entity.Inventory, error) {
	return s.inventoryRepo.GetByCode(ctx, code)
}

// ListInventories 获取库存列表
func (s *InventoryService) ListInventories(ctx context.Context, filter *InventoryFilter) ([]*entity.Inventory, error) {
	return s.inventoryRepo.List(ctx, filter)
}

// CountInventories 统计库存数量
func (s *InventoryService) CountInventories(ctx context.Context, filter *InventoryFilter) (int64, error) {
	return s.inventoryRepo.Count(ctx, filter)
}

// GetLowStockItems 获取低库存物品
func (s *InventoryService) GetLowStockItems(ctx context.Context) ([]*entity.Inventory, error) {
	return s.inventoryRepo.GetLowStockItems(ctx)
}

// InventoryTransactionRequest 库存交易请求
type InventoryTransactionRequest struct {
	InventoryID   string
	Type          string
	Quantity      float64
	UnitPrice     float64
	ReferenceID   string
	ReferenceType string
	OperatorID    string
	Notes         string
}

// ProcessInventoryTransaction 处理库存交易
func (s *InventoryService) ProcessInventoryTransaction(ctx context.Context, req *InventoryTransactionRequest) error {
	// 获取当前库存
	inventory, err := s.inventoryRepo.GetByID(ctx, req.InventoryID)
	if err != nil {
		return fmt.Errorf("inventory not found: %w", err)
	}

	// 计算交易后数量
	var afterQty float64
	switch req.Type {
	case "in":
		afterQty = inventory.Quantity + req.Quantity
	case "out":
		afterQty = inventory.Quantity - req.Quantity
		if afterQty < 0 {
			return fmt.Errorf("insufficient inventory: current=%f, requested=%f", inventory.Quantity, req.Quantity)
		}
	case "adjustment":
		afterQty = req.Quantity
	default:
		return fmt.Errorf("invalid transaction type: %s", req.Type)
	}

	// 计算总金额
	totalAmount := req.Quantity * req.UnitPrice

	// 创建交易记录
	transaction := &entity.InventoryTransaction{
		InventoryID:   req.InventoryID,
		Type:          req.Type,
		Quantity:      req.Quantity,
		UnitPrice:     req.UnitPrice,
		TotalAmount:   totalAmount,
		BeforeQty:     inventory.Quantity,
		AfterQty:      afterQty,
		ReferenceID:   req.ReferenceID,
		ReferenceType: req.ReferenceType,
		OperatorID:    req.OperatorID,
		Notes:         req.Notes,
	}

	// 保存交易记录
	if err := s.inventoryTransRepo.Create(ctx, transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// 更新库存数量
	if err := s.inventoryRepo.UpdateQuantity(ctx, req.InventoryID, afterQty); err != nil {
		return fmt.Errorf("failed to update inventory quantity: %w", err)
	}

	// 更新库存总价值
	inventory.Quantity = afterQty
	inventory.TotalValue = afterQty * inventory.UnitPrice
	if err := s.inventoryRepo.Update(ctx, inventory); err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

// GetInventoryTransactions 获取库存交易记录
func (s *InventoryService) GetInventoryTransactions(ctx context.Context, inventoryID string) ([]*entity.InventoryTransaction, error) {
	return s.inventoryTransRepo.ListByInventoryID(ctx, inventoryID)
}

// GetTransactionByID 根据ID获取交易记录
func (s *InventoryService) GetTransactionByID(ctx context.Context, id string) (*entity.InventoryTransaction, error) {
	return s.inventoryTransRepo.GetByID(ctx, id)
}

// GetTransactionsByReference 根据参考获取交易记录
func (s *InventoryService) GetTransactionsByReference(ctx context.Context, referenceID string, referenceType string) ([]*entity.InventoryTransaction, error) {
	return s.inventoryTransRepo.ListByReference(ctx, referenceID, referenceType)
}
