package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// ReceiptService 收货单服务
type ReceiptService struct {
	receiptRepo       repository.ReceiptRepository
	purchaseOrderRepo repository.PurchaseOrderRepository
	inventoryRepo     repository.InventoryRepository
	inventoryTransactionRepo repository.InventoryTransactionRepository
}

// NewReceiptService 创建收货单服务实例
func NewReceiptService(
	receiptRepo repository.ReceiptRepository,
	purchaseOrderRepo repository.PurchaseOrderRepository,
	inventoryRepo repository.InventoryRepository,
	inventoryTransactionRepo repository.InventoryTransactionRepository,
) *ReceiptService {
	return &ReceiptService{
		receiptRepo:       receiptRepo,
		purchaseOrderRepo: purchaseOrderRepo,
		inventoryRepo:     inventoryRepo,
		inventoryTransactionRepo: inventoryTransactionRepo,
	}
}

// CreateReceipt 创建收货单
func (s *ReceiptService) CreateReceipt(ctx context.Context, receipt *entity.Receipt) error {
	// 验证采购订单是否存在
	_, err := s.purchaseOrderRepo.GetByID(ctx, receipt.PurchaseOrderID)
	if err != nil {
		return errors.New("采购订单不存在")
	}

	// 生成收货单编号
	receipt.Code = fmt.Sprintf("RC%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
	receipt.Status = "pending"
	receipt.CreatedAt = time.Now()
	receipt.UpdatedAt = time.Now()

	// 计算总数量
	var totalQuantity int
	for _, item := range receipt.Items {
		item.ReceiptID = receipt.ID
		totalQuantity += int(item.Quantity)
	}
	receipt.TotalItems = totalQuantity

	return s.receiptRepo.Create(ctx, receipt)
}

// UpdateReceipt 更新收货单
func (s *ReceiptService) UpdateReceipt(ctx context.Context, receipt *entity.Receipt) error {
	// 验证收货单是否存在
	existingReceipt, err := s.receiptRepo.GetByID(ctx, receipt.ID)
	if err != nil {
		return errors.New("收货单不存在")
	}

	// 验证采购订单是否存在
	if receipt.PurchaseOrderID != existingReceipt.PurchaseOrderID {
		_, err := s.purchaseOrderRepo.GetByID(ctx, receipt.PurchaseOrderID)
		if err != nil {
			return errors.New("采购订单不存在")
		}
	}

	// 计算总数量
	var totalQuantity int
	for _, item := range receipt.Items {
		item.ReceiptID = receipt.ID
		totalQuantity += int(item.Quantity)
	}
	receipt.TotalItems = totalQuantity

	receipt.UpdatedAt = time.Now()

	return s.receiptRepo.Update(ctx, receipt)
}

// DeleteReceipt 删除收货单
func (s *ReceiptService) DeleteReceipt(ctx context.Context, id string) error {
	// 验证收货单是否存在
	existingReceipt, err := s.receiptRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("收货单不存在")
	}

	// 只有待处理状态的收货单可以删除
	if existingReceipt.Status != "pending" {
		return errors.New("只有待处理状态的收货单可以删除")
	}

	return s.receiptRepo.Delete(ctx, id)
}

// GetReceiptByID 根据ID获取收货单
func (s *ReceiptService) GetReceiptByID(ctx context.Context, id string) (*entity.Receipt, error) {
	return s.receiptRepo.GetByID(ctx, id)
}

// ListReceipts 列出收货单
func (s *ReceiptService) ListReceipts(ctx context.Context, purchaseOrderID *string, status *string, startDate, endDate *time.Time, page, pageSize int) ([]*entity.Receipt, int64, error) {
	offset := (page - 1) * pageSize
	return s.receiptRepo.List(ctx, purchaseOrderID, status, startDate, endDate, offset, pageSize)
}

// GetReceiptByCode 根据收货单编号获取收货单
func (s *ReceiptService) GetReceiptByCode(ctx context.Context, code string) (*entity.Receipt, error) {
	return s.receiptRepo.GetByCode(ctx, code)
}

// GetReceiptsByPurchaseOrderID 根据采购订单ID获取收货单
func (s *ReceiptService) GetReceiptsByPurchaseOrderID(ctx context.Context, purchaseOrderID string) ([]*entity.Receipt, error) {
	return s.receiptRepo.GetByPurchaseOrderID(ctx, purchaseOrderID)
}

// UpdateReceiptStatus 更新收货单状态
func (s *ReceiptService) UpdateReceiptStatus(ctx context.Context, id string, status string) error {
	// 验证收货单是否存在
	receipt, err := s.receiptRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("收货单不存在")
	}

	// 验证状态是否有效
	validStatuses := []string{"pending", "completed", "cancelled"}
	valid := false
	for _, s := range validStatuses {
		if s == status {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New("无效的收货单状态")
	}

	// 如果状态改为已完成，更新库存
	if status == "completed" {
		for _, item := range receipt.Items {
			// 查找库存记录
			inventory, err := s.inventoryRepo.GetByID(ctx, item.InventoryID)
			itemAmount := item.Quantity * item.UnitPrice
			if err != nil {
				return errors.New("库存记录不存在")
			}

			// 更新现有库存
			inventory.Quantity += item.Quantity
			inventory.TotalValue += itemAmount
			inventory.UpdatedAt = time.Now()
			err = s.inventoryRepo.Update(ctx, inventory)
			if err != nil {
				return errors.New("更新库存失败")
			}

			// 创建库存交易记录
			transaction := &entity.InventoryTransaction{
				InventoryID:    inventory.ID,
				Type:          "in",
				Quantity:      item.Quantity,
				UnitPrice:     item.UnitPrice,
				TotalAmount:   itemAmount,
				ReferenceID:   receipt.ID,
				ReferenceType: "receipt",
				OperatorID:    "system", // 这里应该从上下文获取操作人ID
				Notes:         fmt.Sprintf("采购入库: %s", receipt.Code),
				CreatedAt:     time.Now(),
			}
			err = s.inventoryTransactionRepo.Create(ctx, transaction)
			if err != nil {
				return errors.New("创建库存交易记录失败")
			}
		}

		// 更新采购订单状态为已收货
		purchaseOrder, err := s.purchaseOrderRepo.GetByID(ctx, receipt.PurchaseOrderID)
		if err == nil {
			purchaseOrder.Status = "received"
			purchaseOrder.UpdatedAt = time.Now()
			s.purchaseOrderRepo.Update(ctx, purchaseOrder)
		}
	}

	receipt.Status = status
	receipt.UpdatedAt = time.Now()

	return s.receiptRepo.Update(ctx, receipt)
}
