package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/api/dto"
)

// DeviceHandler 设备处理器
type DeviceHandler struct {
	deviceService *service.DeviceService
}

// NewDeviceHandler 创建设备处理器
func NewDeviceHandler(deviceService *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

// CreateDevice 创建设备
// @Summary 创建设备
// @Description 创建新设备
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param device body service.CreateDeviceRequest true "设备信息"
// @Success 201 {object} dto.Response
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /devices [post]
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var req service.CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters",
			Timestamp: 0,
		})
		return
	}

	device, err := h.deviceService.CreateDevice(c.Request.Context(), &req)
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
		Data:      device,
		Timestamp: 0,
	})
}

// GetDevice 获取设备详情
// @Summary 获取设备详情
// @Description 根据ID获取设备详细信息
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param id path string true "设备ID"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.ErrorResponse
// @Router /devices/{id} [get]
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	id := c.Param("id")

	device, err := h.deviceService.GetDevice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:      404,
			Message:   "Device not found",
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      device,
		Timestamp: 0,
	})
}

// ListDevices 获取设备列表
// @Summary 获取设备列表
// @Description 获取所有设备的列表
// @Tags 设备管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response
// @Failure 500 {object} dto.ErrorResponse
// @Router /devices [get]
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	devices, err := h.deviceService.ListDevices(c.Request.Context(), nil, nil)
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
		Data:      devices,
		Timestamp: 0,
	})
}

// UpdateDevice 更新设备
// @Summary 更新设备
// @Description 更新设备信息
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param id path string true "设备ID"
// @Param device body service.UpdateDeviceRequest true "设备信息"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /devices/{id} [put]
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	id := c.Param("id")

	var req service.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters",
			Timestamp: 0,
		})
		return
	}

	device, err := h.deviceService.UpdateDevice(c.Request.Context(), id, &req)
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
		Data:      device,
		Timestamp: 0,
	})
}

// DeleteDevice 删除设备
// @Summary 删除设备
// @Description 删除指定设备
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param id path string true "设备ID"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Router /devices/{id} [delete]
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	id := c.Param("id")

	if err := h.deviceService.DeleteDevice(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:      404,
			Message:   "Device not found",
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
