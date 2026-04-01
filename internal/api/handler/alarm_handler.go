package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/api/dto"
)

// AlarmHandler 告警处理器
type AlarmHandler struct {
	alarmService *service.AlarmService
}

// NewAlarmHandler 创建告警处理器
func NewAlarmHandler(alarmService *service.AlarmService) *AlarmHandler {
	return &AlarmHandler{
		alarmService: alarmService,
	}
}

// GetAlarm 获取告警详情
// @Summary 获取告警详情
// @Description 根据ID获取告警详细信息
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param id path string true "告警ID"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.ErrorResponse
// @Router /alarms/{id} [get]
func (h *AlarmHandler) GetAlarm(c *gin.Context) {
	id := c.Param("id")

	alarm, err := h.alarmService.GetAlarm(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:      404,
			Message:   "Alarm not found",
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      alarm,
		Timestamp: 0,
	})
}

// ListAlarms 获取告警列表
// @Summary 获取告警列表
// @Description 获取所有告警的列表
// @Tags 告警管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response
// @Failure 500 {object} dto.ErrorResponse
// @Router /alarms [get]
func (h *AlarmHandler) ListAlarms(c *gin.Context) {
	alarms, err := h.alarmService.GetActiveAlarms(c.Request.Context(), nil, nil)
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
		Data:      alarms,
		Timestamp: 0,
	})
}

// AcknowledgeAlarm 确认告警
// @Summary 确认告警
// @Description 确认指定告警
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param id path string true "告警ID"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /alarms/{id}/ack [put]
func (h *AlarmHandler) AcknowledgeAlarm(c *gin.Context) {
	id := c.Param("id")

	// TODO: 从上下文中获取用户ID
	by := ""

	if err := h.alarmService.AcknowledgeAlarm(c.Request.Context(), id, by); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Timestamp: 0,
	})
}

// ClearAlarm 清除告警
// @Summary 清除告警
// @Description 清除指定告警
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param id path string true "告警ID"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /alarms/{id}/clear [put]
func (h *AlarmHandler) ClearAlarm(c *gin.Context) {
	id := c.Param("id")

	if err := h.alarmService.ClearAlarm(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Timestamp: 0,
	})
}
