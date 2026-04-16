package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

// CostEntryHandler 成本条目处理器
type CostEntryHandler struct {
	costEntryService *service.CostEntryService
}

// NewCostEntryHandler 创建成本条目处理器实例
func NewCostEntryHandler(costEntryService *service.CostEntryService) *CostEntryHandler {
	return &CostEntryHandler{
		costEntryService: costEntryService,
	}
}

// CreateCostEntry 创建成本条目
func (h *CostEntryHandler) CreateCostEntry(c *gin.Context) {
	var entry entity.CostEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.costEntryService.CreateCostEntry(c.Request.Context(), &entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// UpdateCostEntry 更新成本条目
func (h *CostEntryHandler) UpdateCostEntry(c *gin.Context) {
	id := c.Param("id")
	var entry entity.CostEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry.ID = id

	if err := h.costEntryService.UpdateCostEntry(c.Request.Context(), &entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// DeleteCostEntry 删除成本条目
func (h *CostEntryHandler) DeleteCostEntry(c *gin.Context) {
	id := c.Param("id")

	if err := h.costEntryService.DeleteCostEntry(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成本条目删除成功"})
}

// GetCostEntryByID 根据ID获取成本条目
func (h *CostEntryHandler) GetCostEntryByID(c *gin.Context) {
	id := c.Param("id")

	entry, err := h.costEntryService.GetCostEntryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// GetCostEntryByCode 根据编码获取成本条目
func (h *CostEntryHandler) GetCostEntryByCode(c *gin.Context) {
	code := c.Param("code")

	entry, err := h.costEntryService.GetCostEntryByCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// ListCostEntries 列出成本条目
func (h *CostEntryHandler) ListCostEntries(c *gin.Context) {
	categoryID := c.Query("category_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	status := c.Query("status")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")

	var categoryIDPtr *string
	if categoryID != "" {
		categoryIDPtr = &categoryID
	}

	var startDatePtr, endDatePtr *time.Time
	if startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDatePtr = &startDate
		}
	}

	if endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDatePtr = &endDate
		}
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	// 转换分页参数
	pageInt := 1
	pageSizeInt := 10
	if _, err := fmt.Sscanf(page, "%d", &pageInt); err != nil {
		pageInt = 1
	}
	if _, err := fmt.Sscanf(pageSize, "%d", &pageSizeInt); err != nil {
		pageSizeInt = 10
	}

	entries, total, err := h.costEntryService.ListCostEntries(c.Request.Context(), categoryIDPtr, startDatePtr, endDatePtr, statusPtr, pageInt, pageSizeInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": entries,
		"total": total,
		"page":  pageInt,
		"size":  pageSizeInt,
	})
}

// ApproveCostEntry 审批成本条目
func (h *CostEntryHandler) ApproveCostEntry(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ApprovedBy string `json:"approved_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.costEntryService.ApproveCostEntry(c.Request.Context(), id, req.ApprovedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成本条目审批成功"})
}

// RejectCostEntry 拒绝成本条目
func (h *CostEntryHandler) RejectCostEntry(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ApprovedBy string `json:"approved_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.costEntryService.RejectCostEntry(c.Request.Context(), id, req.ApprovedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成本条目拒绝成功"})
}

// GetTotalByCategory 根据成本类别获取总金额
func (h *CostEntryHandler) GetTotalByCategory(c *gin.Context) {
	categoryID := c.Param("category_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDatePtr, endDatePtr *time.Time
	if startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDatePtr = &startDate
		}
	}

	if endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDatePtr = &endDate
		}
	}

	total, err := h.costEntryService.GetTotalByCategory(c.Request.Context(), categoryID, startDatePtr, endDatePtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total})
}

// GetTotalByPeriod 根据时间段获取总金额
func (h *CostEntryHandler) GetTotalByPeriod(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDatePtr, endDatePtr *time.Time
	if startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDatePtr = &startDate
		}
	}

	if endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDatePtr = &endDate
		}
	}

	total, err := h.costEntryService.GetTotalByPeriod(c.Request.Context(), startDatePtr, endDatePtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total})
}
