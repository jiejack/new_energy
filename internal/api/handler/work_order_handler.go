package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/api/dto"
)

// WorkOrderHandler 工单处理器
type WorkOrderHandler struct {
	workOrderService *service.WorkOrderService
}

// NewWorkOrderHandler 创建工单处理器
func NewWorkOrderHandler(workOrderService *service.WorkOrderService) *WorkOrderHandler {
	return &WorkOrderHandler{
		workOrderService: workOrderService,
	}
}

// CreateWorkOrder 创建工单
// @Summary 创建工单
// @Description 创建新工单
// @Tags 工单管理
// @Accept json
// @Produce json
// @Param workOrder body service.CreateWorkOrderRequest true "工单信息"
// @Success 201 {object} dto.Response
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /work-orders [post]
func (h *WorkOrderHandler) CreateWorkOrder(c *gin.Context) {
	var req service.CreateWorkOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters",
			Timestamp: 0,
		})
		return
	}

	workOrder, err := h.workOrderService.CreateWorkOrder(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      workOrder,
		Timestamp: 0,
	})
}

// GetWorkOrder 获取工单详情
// @Summary 获取工单详情
// @Description 根据ID获取工单详细信息
// @Tags 工单管理
// @Accept json
// @Produce json
// @Param id path string true "工单ID"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.ErrorResponse
// @Router /work-orders/{id} [get]
func (h *WorkOrderHandler) GetWorkOrder(c *gin.Context) {
	id := c.Param("id")

	workOrder, err := h.workOrderService.GetWorkOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:      404,
			Message:   "Work order not found",
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      workOrder,
		Timestamp: 0,
	})
}

// ListWorkOrders 获取工单列表
// @Summary 获取工单列表
// @Description 获取所有工单的列表
// @Tags 工单管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response
// @Failure 500 {object} dto.ErrorResponse
// @Router /work-orders [get]
func (h *WorkOrderHandler) ListWorkOrders(c *gin.Context) {
	var filter service.WorkOrderFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid filter parameters",
			Timestamp: 0,
		})
		return
	}

	workOrders, err := h.workOrderService.ListWorkOrders(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      workOrders,
		Timestamp: 0,
	})
}

// UpdateWorkOrder 更新工单
// @Summary 更新工单
// @Description 更新工单信息
// @Tags 工单管理
// @Accept json
// @Produce json
// @Param id path string true "工单ID"
// @Param workOrder body service.UpdateWorkOrderRequest true "工单信息"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /work-orders/{id} [put]
func (h *WorkOrderHandler) UpdateWorkOrder(c *gin.Context) {
	id := c.Param("id")

	var req service.UpdateWorkOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters",
			Timestamp: 0,
		})
		return
	}

	workOrder, err := h.workOrderService.UpdateWorkOrder(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      workOrder,
		Timestamp: 0,
	})
}

// DeleteWorkOrder 删除工单
// @Summary 删除工单
// @Description 删除指定工单
// @Tags 工单管理
// @Accept json
// @Produce json
// @Param id path string true "工单ID"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Router /work-orders/{id} [delete]
func (h *WorkOrderHandler) DeleteWorkOrder(c *gin.Context) {
	id := c.Param("id")

	if err := h.workOrderService.DeleteWorkOrder(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:      404,
			Message:   "Work order not found",
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetWorkOrderStats 获取工单统计
// @Summary 获取工单统计
// @Description 获取工单状态统计信息
// @Tags 工单管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response
// @Failure 500 {object} dto.ErrorResponse
// @Router /work-orders/stats [get]
func (h *WorkOrderHandler) GetWorkOrderStats(c *gin.Context) {
	stats, err := h.workOrderService.GetWorkOrderStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      stats,
		Timestamp: 0,
	})
}
