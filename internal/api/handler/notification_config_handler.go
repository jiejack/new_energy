package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

type NotificationConfigHandler struct {
	configService *service.NotificationConfigService
}

func NewNotificationConfigHandler(configService *service.NotificationConfigService) *NotificationConfigHandler {
	return &NotificationConfigHandler{configService: configService}
}

func (h *NotificationConfigHandler) GetAllConfigs(c *gin.Context) {
	configs, err := h.configService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: configs})
}

func (h *NotificationConfigHandler) GetConfigByType(c *gin.Context) {
	notifType := entity.NotificationType(c.Param("type"))
	config, err := h.configService.GetByType(c.Request.Context(), notifType)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "Config not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: config})
}

func (h *NotificationConfigHandler) UpdateConfig(c *gin.Context) {
	notifType := entity.NotificationType(c.Param("type"))
	var req service.UpdateNotificationConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	config, err := h.configService.UpdateConfig(c.Request.Context(), notifType, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: config})
}

func (h *NotificationConfigHandler) EnableConfig(c *gin.Context) {
	notifType := entity.NotificationType(c.Param("type"))
	if err := h.configService.EnableConfig(c.Request.Context(), notifType); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
}

func (h *NotificationConfigHandler) DisableConfig(c *gin.Context) {
	notifType := entity.NotificationType(c.Param("type"))
	if err := h.configService.DisableConfig(c.Request.Context(), notifType); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
}

func (h *NotificationConfigHandler) TestConfig(c *gin.Context) {
	notifType := entity.NotificationType(c.Param("type"))
	if err := h.configService.TestConfig(c.Request.Context(), notifType); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "Test successful"})
}
