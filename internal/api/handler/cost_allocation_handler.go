package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

// CostAllocationHandler 成本分配处理器
type CostAllocationHandler struct {
	costAllocationService *service.CostAllocationService
}

// NewCostAllocationHandler 创建成本分配处理器实例
func NewCostAllocationHandler(costAllocationService *service.CostAllocationService) *CostAllocationHandler {
	return &CostAllocationHandler{
		costAllocationService: costAllocationService,
	}
}

// CreateCostAllocation 创建成本分配
func (h *CostAllocationHandler) CreateCostAllocation(c *gin.Context) {
	var allocation entity.CostAllocation
	if err := c.ShouldBindJSON(&allocation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.costAllocationService.CreateCostAllocation(c.Request.Context(), &allocation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, allocation)
}

// UpdateCostAllocation 更新成本分配
func (h *CostAllocationHandler) UpdateCostAllocation(c *gin.Context) {
	id := c.Param("id")
	var allocation entity.CostAllocation
	if err := c.ShouldBindJSON(&allocation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allocation.ID = id

	if err := h.costAllocationService.UpdateCostAllocation(c.Request.Context(), &allocation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, allocation)
}

// DeleteCostAllocation 删除成本分配
func (h *CostAllocationHandler) DeleteCostAllocation(c *gin.Context) {
	id := c.Param("id")

	if err := h.costAllocationService.DeleteCostAllocation(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成本分配删除成功"})
}

// GetCostAllocationByID 根据ID获取成本分配
func (h *CostAllocationHandler) GetCostAllocationByID(c *gin.Context) {
	id := c.Param("id")

	allocation, err := h.costAllocationService.GetCostAllocationByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, allocation)
}

// ListCostAllocationsByCostEntryID 根据成本条目ID列出成本分配
func (h *CostAllocationHandler) ListCostAllocationsByCostEntryID(c *gin.Context) {
	costEntryID := c.Param("cost_entry_id")

	allocations, err := h.costAllocationService.ListCostAllocationsByCostEntryID(c.Request.Context(), costEntryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, allocations)
}

// ListCostAllocationsByAllocated 根据分配对象列出成本分配
func (h *CostAllocationHandler) ListCostAllocationsByAllocated(c *gin.Context) {
	allocatedTo := c.Query("allocated_to")
	allocatedID := c.Query("allocated_id")

	if allocatedTo == "" || allocatedID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "allocated_to 和 allocated_id 不能为空"})
		return
	}

	allocations, err := h.costAllocationService.ListCostAllocationsByAllocated(c.Request.Context(), allocatedTo, allocatedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, allocations)
}

// GetTotalByAllocated 根据分配对象获取总金额
func (h *CostAllocationHandler) GetTotalByAllocated(c *gin.Context) {
	allocatedTo := c.Query("allocated_to")
	allocatedID := c.Query("allocated_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if allocatedTo == "" || allocatedID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "allocated_to 和 allocated_id 不能为空"})
		return
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

	total, err := h.costAllocationService.GetTotalByAllocated(c.Request.Context(), allocatedTo, allocatedID, startDatePtr, endDatePtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total})
}
