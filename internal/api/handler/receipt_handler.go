package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

// ReceiptHandler 收货单处理器
type ReceiptHandler struct {
	receiptService *service.ReceiptService
}

// NewReceiptHandler 创建收货单处理器实例
func NewReceiptHandler(receiptService *service.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{
		receiptService: receiptService,
	}
}

// CreateReceipt 创建收货单
func (h *ReceiptHandler) CreateReceipt(c *gin.Context) {
	var req struct {
		PurchaseOrderID string              `json:"purchase_order_id" binding:"required"`
		Items           []entity.ReceiptItem `json:"items" binding:"required"`
		ReceiptDate     time.Time           `json:"receipt_date"`
		Notes           string              `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 转换Items为指针类型
	items := make([]*entity.ReceiptItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = &item
	}

	receipt := &entity.Receipt{
		PurchaseOrderID: req.PurchaseOrderID,
		Items:           items,
		ReceiptDate:     req.ReceiptDate,
		Notes:           req.Notes,
	}

	err := h.receiptService.CreateReceipt(c.Request.Context(), receipt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// UpdateReceipt 更新收货单
func (h *ReceiptHandler) UpdateReceipt(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的收货单ID"})
		return
	}

	var req struct {
		PurchaseOrderID string              `json:"purchase_order_id" binding:"required"`
		Items           []entity.ReceiptItem `json:"items" binding:"required"`
		ReceiptDate     time.Time           `json:"receipt_date"`
		Notes           string              `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 转换Items为指针类型
	items := make([]*entity.ReceiptItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = &item
	}

	receipt := &entity.Receipt{
		ID:              id,
		PurchaseOrderID: req.PurchaseOrderID,
		Items:           items,
		ReceiptDate:     req.ReceiptDate,
		Notes:           req.Notes,
	}

	err := h.receiptService.UpdateReceipt(c.Request.Context(), receipt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// DeleteReceipt 删除收货单
func (h *ReceiptHandler) DeleteReceipt(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的收货单ID"})
		return
	}

	err := h.receiptService.DeleteReceipt(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "收货单删除成功"})
}

// GetReceipt 获取收货单详情
func (h *ReceiptHandler) GetReceipt(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的收货单ID"})
		return
	}

	receipt, err := h.receiptService.GetReceiptByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// ListReceipts 列出收货单
func (h *ReceiptHandler) ListReceipts(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	purchaseOrderIDStr := c.Query("purchase_order_id")
	var purchaseOrderID *string
	if purchaseOrderIDStr != "" {
		purchaseOrderID = &purchaseOrderIDStr
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

	receipts, total, err := h.receiptService.ListReceipts(c.Request.Context(), purchaseOrderID, statusPtr, startDate, endDate, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  receipts,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// UpdateReceiptStatus 更新收货单状态
func (h *ReceiptHandler) UpdateReceiptStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的收货单ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	err := h.receiptService.UpdateReceiptStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "收货单状态更新成功"})
}
