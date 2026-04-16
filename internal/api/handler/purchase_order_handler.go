package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

// PurchaseOrderHandler 采购订单处理器
type PurchaseOrderHandler struct {
	purchaseOrderService *service.PurchaseOrderService
}

// NewPurchaseOrderHandler 创建采购订单处理器实例
func NewPurchaseOrderHandler(purchaseOrderService *service.PurchaseOrderService) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{
		purchaseOrderService: purchaseOrderService,
	}
}

// CreatePurchaseOrder 创建采购订单
func (h *PurchaseOrderHandler) CreatePurchaseOrder(c *gin.Context) {
	var req struct {
		SupplierID  string                    `json:"supplier_id" binding:"required"`
		Items       []entity.PurchaseOrderItem `json:"items" binding:"required"`
		ExpectedDate *time.Time                `json:"expected_date"`
		Notes       string                    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 转换Items为指针类型
	items := make([]*entity.PurchaseOrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = &item
	}

	order := &entity.PurchaseOrder{
		SupplierID:  req.SupplierID,
		Items:       items,
		ExpectedDate: req.ExpectedDate,
		Notes:       req.Notes,
	}

	err := h.purchaseOrderService.CreatePurchaseOrder(c.Request.Context(), order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// UpdatePurchaseOrder 更新采购订单
func (h *PurchaseOrderHandler) UpdatePurchaseOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订单ID"})
		return
	}

	var req struct {
		SupplierID  string                    `json:"supplier_id" binding:"required"`
		Items       []entity.PurchaseOrderItem `json:"items" binding:"required"`
		ExpectedDate *time.Time                `json:"expected_date"`
		Notes       string                    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 转换Items为指针类型
	items := make([]*entity.PurchaseOrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = &item
	}

	order := &entity.PurchaseOrder{
		ID:          id,
		SupplierID:  req.SupplierID,
		Items:       items,
		ExpectedDate: req.ExpectedDate,
		Notes:       req.Notes,
	}

	err := h.purchaseOrderService.UpdatePurchaseOrder(c.Request.Context(), order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// DeletePurchaseOrder 删除采购订单
func (h *PurchaseOrderHandler) DeletePurchaseOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订单ID"})
		return
	}

	err := h.purchaseOrderService.DeletePurchaseOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "采购订单删除成功"})
}

// GetPurchaseOrder 获取采购订单详情
func (h *PurchaseOrderHandler) GetPurchaseOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订单ID"})
		return
	}

	order, err := h.purchaseOrderService.GetPurchaseOrderByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// ListPurchaseOrders 列出采购订单
func (h *PurchaseOrderHandler) ListPurchaseOrders(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	supplierIDStr := c.Query("supplier_id")
	var supplierID *string
	if supplierIDStr != "" {
		supplierID = &supplierIDStr
	}

	status := c.Query("status")
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	startDateStr := c.Query("start_date")
	var startDate *time.Time
	if startDateStr != "" {
		date, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = &date
		}
	}

	endDateStr := c.Query("end_date")
	var endDate *time.Time
	if endDateStr != "" {
		date, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = &date
		}
	}

	orders, total, err := h.purchaseOrderService.ListPurchaseOrders(c.Request.Context(), supplierID, statusPtr, startDate, endDate, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  orders,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// UpdatePurchaseOrderStatus 更新采购订单状态
func (h *PurchaseOrderHandler) UpdatePurchaseOrderStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订单ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	err := h.purchaseOrderService.UpdatePurchaseOrderStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "采购订单状态更新成功"})
}
