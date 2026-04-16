package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

// InventoryHandler 库存处理器
type InventoryHandler struct {
	inventoryService *service.InventoryService
}

// NewInventoryHandler 创建库存处理器
func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

// CreateInventory 创建库存
// @Summary 创建库存
// @Description 创建新的库存记录
// @Tags 库存管理
// @Accept json
// @Produce json
// @Param inventory body entity.Inventory true "库存信息"
// @Success 200 {object} entity.Inventory
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/inventory [post]
func (h *InventoryHandler) CreateInventory(c *gin.Context) {
	var inventory entity.Inventory
	if err := c.ShouldBindJSON(&inventory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.inventoryService.CreateInventory(c.Request.Context(), &inventory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inventory)
}

// UpdateInventory 更新库存
// @Summary 更新库存
// @Description 更新库存记录
// @Tags 库存管理
// @Accept json
// @Produce json
// @Param id path string true "库存ID"
// @Param inventory body entity.Inventory true "库存信息"
// @Success 200 {object} entity.Inventory
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/inventory/{id} [put]
func (h *InventoryHandler) UpdateInventory(c *gin.Context) {
	id := c.Param("id")
	var inventory entity.Inventory
	if err := c.ShouldBindJSON(&inventory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inventory.ID = id
	if err := h.inventoryService.UpdateInventory(c.Request.Context(), &inventory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inventory)
}

// DeleteInventory 删除库存
// @Summary 删除库存
// @Description 删除库存记录
// @Tags 库存管理
// @Produce json
// @Param id path string true "库存ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/inventory/{id} [delete]
func (h *InventoryHandler) DeleteInventory(c *gin.Context) {
	id := c.Param("id")

	if err := h.inventoryService.DeleteInventory(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inventory deleted successfully"})
}

// GetInventory 获取库存
// @Summary 获取库存
// @Description 根据ID获取库存记录
// @Tags 库存管理
// @Produce json
// @Param id path string true "库存ID"
// @Success 200 {object} entity.Inventory
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/inventory/{id} [get]
func (h *InventoryHandler) GetInventory(c *gin.Context) {
	id := c.Param("id")

	inventory, err := h.inventoryService.GetInventoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inventory)
}

// ListInventories 获取库存列表
// @Summary 获取库存列表
// @Description 获取库存记录列表
// @Tags 库存管理
// @Produce json
// @Param type query string false "库存类型"
// @Param status query string false "库存状态"
// @Param name query string false "库存名称"
// @Param code query string false "库存编码"
// @Success 200 {array} entity.Inventory
// @Failure 500 {object} map[string]string
// @Router /api/v1/inventory [get]
func (h *InventoryHandler) ListInventories(c *gin.Context) {
	filter := &service.InventoryFilter{
		Type:   c.Query("type"),
		Status: c.Query("status"),
		Name:   c.Query("name"),
		Code:   c.Query("code"),
	}

	inventories, err := h.inventoryService.ListInventories(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inventories)
}

// GetLowStockItems 获取低库存物品
// @Summary 获取低库存物品
// @Description 获取低于最小库存的物品
// @Tags 库存管理
// @Produce json
// @Success 200 {array} entity.Inventory
// @Failure 500 {object} map[string]string
// @Router /api/v1/inventory/low-stock [get]
func (h *InventoryHandler) GetLowStockItems(c *gin.Context) {
	items, err := h.inventoryService.GetLowStockItems(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

// ProcessTransaction 处理库存交易
// @Summary 处理库存交易
// @Description 处理库存入库、出库或调整
// @Tags 库存管理
// @Accept json
// @Produce json
// @Param transaction body service.InventoryTransactionRequest true "交易信息"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/inventory/transactions [post]
func (h *InventoryHandler) ProcessTransaction(c *gin.Context) {
	var req service.InventoryTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.inventoryService.ProcessInventoryTransaction(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction processed successfully"})
}

// GetTransactions 获取库存交易记录
// @Summary 获取库存交易记录
// @Description 根据库存ID获取交易记录
// @Tags 库存管理
// @Produce json
// @Param inventory_id path string true "库存ID"
// @Success 200 {array} entity.InventoryTransaction
// @Failure 500 {object} map[string]string
// @Router /api/v1/inventory/{inventory_id}/transactions [get]
func (h *InventoryHandler) GetTransactions(c *gin.Context) {
	inventoryID := c.Param("inventory_id")

	transactions, err := h.inventoryService.GetInventoryTransactions(c.Request.Context(), inventoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
}
