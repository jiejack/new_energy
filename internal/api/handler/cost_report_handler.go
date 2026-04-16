package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

// CostReportHandler 成本报表处理器
type CostReportHandler struct {
	costReportService *service.CostReportService
}

// NewCostReportHandler 创建成本报表处理器实例
func NewCostReportHandler(costReportService *service.CostReportService) *CostReportHandler {
	return &CostReportHandler{
		costReportService: costReportService,
	}
}

// CreateCostReport 创建成本报表
func (h *CostReportHandler) CreateCostReport(c *gin.Context) {
	var report entity.CostReport
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.costReportService.CreateCostReport(c.Request.Context(), &report); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, report)
}

// UpdateCostReport 更新成本报表
func (h *CostReportHandler) UpdateCostReport(c *gin.Context) {
	id := c.Param("id")
	var report entity.CostReport
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report.ID = id

	if err := h.costReportService.UpdateCostReport(c.Request.Context(), &report); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// DeleteCostReport 删除成本报表
func (h *CostReportHandler) DeleteCostReport(c *gin.Context) {
	id := c.Param("id")

	if err := h.costReportService.DeleteCostReport(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成本报表删除成功"})
}

// GetCostReportByID 根据ID获取成本报表
func (h *CostReportHandler) GetCostReportByID(c *gin.Context) {
	id := c.Param("id")

	report, err := h.costReportService.GetCostReportByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetCostReportByCode 根据编码获取成本报表
func (h *CostReportHandler) GetCostReportByCode(c *gin.Context) {
	code := c.Param("code")

	report, err := h.costReportService.GetCostReportByCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// ListCostReports 列出成本报表
func (h *CostReportHandler) ListCostReports(c *gin.Context) {
	reportType := c.Query("report_type")
	status := c.Query("status")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")

	var reportTypePtr *string
	if reportType != "" {
		reportTypePtr = &reportType
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
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

	// 转换分页参数
	pageInt := 1
	pageSizeInt := 10
	if _, err := fmt.Sscanf(page, "%d", &pageInt); err != nil {
		pageInt = 1
	}
	if _, err := fmt.Sscanf(pageSize, "%d", &pageSizeInt); err != nil {
		pageSizeInt = 10
	}

	reports, total, err := h.costReportService.ListCostReports(c.Request.Context(), reportTypePtr, statusPtr, startDatePtr, endDatePtr, pageInt, pageSizeInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": reports,
		"total": total,
		"page":  pageInt,
		"size":  pageSizeInt,
	})
}

// GenerateCostReport 生成成本报表
func (h *CostReportHandler) GenerateCostReport(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		GeneratedBy string `json:"generated_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.costReportService.GenerateCostReport(c.Request.Context(), id, req.GeneratedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成本报表生成成功"})
}

// ApproveCostReport 审批成本报表
func (h *CostReportHandler) ApproveCostReport(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ApprovedBy string `json:"approved_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.costReportService.ApproveCostReport(c.Request.Context(), id, req.ApprovedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成本报表审批成功"})
}

// RejectCostReport 拒绝成本报表
func (h *CostReportHandler) RejectCostReport(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ApprovedBy string `json:"approved_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.costReportService.RejectCostReport(c.Request.Context(), id, req.ApprovedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成本报表拒绝成功"})
}
