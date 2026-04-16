package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// PurchaseOrderService 采购订单服务
type PurchaseOrderService struct {
	purchaseOrderRepo repository.PurchaseOrderRepository
	supplierRepo      repository.SupplierRepository
}

// NewPurchaseOrderService 创建采购订单服务实例
func NewPurchaseOrderService(
	purchaseOrderRepo repository.PurchaseOrderRepository,
	supplierRepo repository.SupplierRepository,
) *PurchaseOrderService {
	return &PurchaseOrderService{
		purchaseOrderRepo: purchaseOrderRepo,
		supplierRepo:      supplierRepo,
	}
}

// CreatePurchaseOrder 创建采购订单
func (s *PurchaseOrderService) CreatePurchaseOrder(ctx context.Context, order *entity.PurchaseOrder) error {
	// 验证供应商是否存在
	_, err := s.supplierRepo.GetByID(ctx, order.SupplierID)
	if err != nil {
		return errors.New("供应商不存在")
	}

	// 生成订单编号
	order.Code = fmt.Sprintf("PO%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
	order.Status = "pending"
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	// 计算总金额
	var totalAmount float64
	for _, item := range order.Items {
		item.PurchaseOrderID = order.ID
		item.TotalAmount = item.Quantity * item.UnitPrice
		totalAmount += item.TotalAmount
	}
	order.TotalAmount = totalAmount

	return s.purchaseOrderRepo.Create(ctx, order)
}

// UpdatePurchaseOrder 更新采购订单
func (s *PurchaseOrderService) UpdatePurchaseOrder(ctx context.Context, order *entity.PurchaseOrder) error {
	// 验证订单是否存在
	existingOrder, err := s.purchaseOrderRepo.GetByID(ctx, order.ID)
	if err != nil {
		return errors.New("采购订单不存在")
	}

	// 验证供应商是否存在
	if order.SupplierID != existingOrder.SupplierID {
		_, err := s.supplierRepo.GetByID(ctx, order.SupplierID)
		if err != nil {
			return errors.New("供应商不存在")
		}
	}

	// 计算总金额
	var totalAmount float64
	for _, item := range order.Items {
		item.PurchaseOrderID = order.ID
		item.TotalAmount = item.Quantity * item.UnitPrice
		totalAmount += item.TotalAmount
	}
	order.TotalAmount = totalAmount

	order.UpdatedAt = time.Now()

	return s.purchaseOrderRepo.Update(ctx, order)
}

// DeletePurchaseOrder 删除采购订单
func (s *PurchaseOrderService) DeletePurchaseOrder(ctx context.Context, id string) error {
	// 验证订单是否存在
	existingOrder, err := s.purchaseOrderRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("采购订单不存在")
	}

	// 只有待处理状态的订单可以删除
	if existingOrder.Status != "pending" {
		return errors.New("只有待处理状态的订单可以删除")
	}

	return s.purchaseOrderRepo.Delete(ctx, id)
}

// GetPurchaseOrderByID 根据ID获取采购订单
func (s *PurchaseOrderService) GetPurchaseOrderByID(ctx context.Context, id string) (*entity.PurchaseOrder, error) {
	return s.purchaseOrderRepo.GetByID(ctx, id)
}

// ListPurchaseOrders 列出采购订单
func (s *PurchaseOrderService) ListPurchaseOrders(ctx context.Context, supplierID *string, status *string, startDate, endDate *time.Time, page, pageSize int) ([]*entity.PurchaseOrder, int64, error) {
	offset := (page - 1) * pageSize
	return s.purchaseOrderRepo.List(ctx, supplierID, status, startDate, endDate, offset, pageSize)
}

// GetPurchaseOrderByCode 根据订单编号获取采购订单
func (s *PurchaseOrderService) GetPurchaseOrderByCode(ctx context.Context, code string) (*entity.PurchaseOrder, error) {
	return s.purchaseOrderRepo.GetByCode(ctx, code)
}

// UpdatePurchaseOrderStatus 更新采购订单状态
func (s *PurchaseOrderService) UpdatePurchaseOrderStatus(ctx context.Context, id string, status string) error {
	// 验证订单是否存在
	order, err := s.purchaseOrderRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("采购订单不存在")
	}

	// 验证状态是否有效
	validStatuses := []string{"pending", "approved", "rejected", "sent", "received", "cancelled"}
	valid := false
	for _, s := range validStatuses {
		if s == status {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New("无效的订单状态")
	}

	order.Status = status
	order.UpdatedAt = time.Now()

	return s.purchaseOrderRepo.Update(ctx, order)
}
