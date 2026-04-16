package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

// CostCategoryHandler 成本类别处理器
type CostCategoryHandler struct {
	costCategoryService *service.CostCategoryService
}

// NewCostCategoryHandler 创建成本类别处理器实例
func NewCostCategoryHandler(costCategoryService *service.CostCategoryService) *CostCategoryHandler {
	return &CostCategoryHandler{
		costCategoryService: costCategoryService,
	}
}

// CreateCostCategory 创建成本类别
func (h *CostCategoryHandler) CreateCostCategory(c *gin.Context) {
	var category entity.CostCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.costCategoryService.CreateCostCategory(c.Request.Context(), &category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// UpdateCostCategory 更新成本类别
func (h *CostCategoryHandler) UpdateCostCategory(c *gin.Context) {
	id := c.Param("id")
	var category entity.CostCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.ID = id

	if err := h.costCategoryService.UpdateCostCategory(c.Request.Context(), &category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCostCategory 删除成本类别
func (h *CostCategoryHandler) DeleteCostCategory(c *gin.Context) {
	id := c.Param("id")

	if err := h.costCategoryService.DeleteCostCategory(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成本类别删除成功"})
}

// GetCostCategoryByID 根据ID获取成本类别
func (h *CostCategoryHandler) GetCostCategoryByID(c *gin.Context) {
	id := c.Param("id")

	category, err := h.costCategoryService.GetCostCategoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// GetCostCategoryByCode 根据编码获取成本类别
func (h *CostCategoryHandler) GetCostCategoryByCode(c *gin.Context) {
	code := c.Param("code")

	category, err := h.costCategoryService.GetCostCategoryByCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// ListCostCategories 列出成本类别
func (h *CostCategoryHandler) ListCostCategories(c *gin.Context) {
	parentID := c.Query("parent_id")
	status := c.Query("status")

	var parentIDPtr *string
	if parentID != "" {
		parentIDPtr = &parentID
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	categories, err := h.costCategoryService.ListCostCategories(c.Request.Context(), parentIDPtr, statusPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCostCategoryTree 获取成本类别树
func (h *CostCategoryHandler) GetCostCategoryTree(c *gin.Context) {
	categories, err := h.costCategoryService.GetCostCategoryTree(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}
