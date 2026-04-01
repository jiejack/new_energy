package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// AlarmRuleHandler 告警规则处理器
type AlarmRuleHandler struct {
	ruleService *service.AlarmRuleService
}

// NewAlarmRuleHandler 创建告警规则处理器
func NewAlarmRuleHandler(ruleService *service.AlarmRuleService) *AlarmRuleHandler {
	return &AlarmRuleHandler{
		ruleService: ruleService,
	}
}

// CreateAlarmRule 创建告警规则
// @Summary 创建告警规则
// @Description 创建新的告警规则，支持限值告警、趋势告警和自定义告警三种类型
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param rule body service.CreateAlarmRuleRequest true "告警规则信息"
// @Example request {"name":"逆变器温度过高告警","description":"监测逆变器温度，超过阈值触发告警","type":"limit","level":3,"condition":">","threshold":85.0,"duration":60,"notify_channels":["email","sms"],"notify_users":["admin","operator"],"status":1}
// @Success 200 {object} dto.Response{data=entity.AlarmRule} "创建成功"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /alarm-rules [post]
func (h *AlarmRuleHandler) CreateAlarmRule(c *gin.Context) {
	var req service.CreateAlarmRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters: " + err.Error(),
			Timestamp: 0,
		})
		return
	}

	// TODO: 从上下文中获取用户信息
	if req.CreatedBy == "" {
		req.CreatedBy = "system"
	}

	rule, err := h.ruleService.CreateAlarmRule(c.Request.Context(), &req)
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
		Data:      rule,
		Timestamp: 0,
	})
}

// UpdateAlarmRule 更新告警规则
// @Summary 更新告警规则
// @Description 更新指定的告警规则，包括规则名称、描述、级别、条件、阈值等
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param id path string true "告警规则ID" example(alarm-rule-001)
// @Param rule body service.UpdateAlarmRuleRequest true "告警规则信息"
// @Example request {"name":"逆变器温度过高告警-更新","description":"监测逆变器温度，超过阈值触发告警","level":4,"condition":">=","threshold":90.0,"duration":30,"notify_channels":["email"],"notify_users":["admin"]}
// @Success 200 {object} dto.Response{data=entity.AlarmRule} "更新成功"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 404 {object} dto.ErrorResponse "告警规则不存在"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /alarm-rules/{id} [put]
func (h *AlarmRuleHandler) UpdateAlarmRule(c *gin.Context) {
	id := c.Param("id")

	var req service.UpdateAlarmRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters: " + err.Error(),
			Timestamp: 0,
		})
		return
	}

	// TODO: 从上下文中获取用户信息
	if req.UpdatedBy == "" {
		req.UpdatedBy = "system"
	}

	rule, err := h.ruleService.UpdateAlarmRule(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:      404,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      rule,
		Timestamp: 0,
	})
}

// DeleteAlarmRule 删除告警规则
// @Summary 删除告警规则
// @Description 删除指定的告警规则，删除后不可恢复
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param id path string true "告警规则ID" example(alarm-rule-001)
// @Success 200 {object} dto.Response "删除成功"
// @Failure 404 {object} dto.ErrorResponse "告警规则不存在"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /alarm-rules/{id} [delete]
func (h *AlarmRuleHandler) DeleteAlarmRule(c *gin.Context) {
	id := c.Param("id")

	if err := h.ruleService.DeleteAlarmRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:      404,
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

// GetAlarmRule 获取告警规则详情
// @Summary 获取告警规则详情
// @Description 根据ID获取告警规则详细信息，包括规则配置、关联对象、通知配置等
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param id path string true "告警规则ID" example(alarm-rule-001)
// @Success 200 {object} dto.Response{data=entity.AlarmRule} "获取成功"
// @Failure 404 {object} dto.ErrorResponse "告警规则不存在"
// @Router /alarm-rules/{id} [get]
func (h *AlarmRuleHandler) GetAlarmRule(c *gin.Context) {
	id := c.Param("id")

	rule, err := h.ruleService.GetAlarmRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:      404,
			Message:   "Alarm rule not found",
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      rule,
		Timestamp: 0,
	})
}

// ListAlarmRules 获取告警规则列表
// @Summary 获取告警规则列表
// @Description 获取告警规则列表，支持分页、过滤和排序。可以按规则名称、类型、级别、状态、采集点、设备、厂站等条件过滤
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param name query string false "规则名称" example(温度告警)
// @Param type query string false "规则类型(limit/trend/custom)" Enums(limit, trend, custom) example(limit)
// @Param level query int false "告警级别(1-4)" minimum(1) maximum(4) example(3)
// @Param status query int false "状态(0-禁用,1-启用)" Enums(0, 1) example(1)
// @Param point_id query string false "采集点ID" example(point-001)
// @Param device_id query string false "设备ID" example(device-001)
// @Param station_id query string false "厂站ID" example(station-001)
// @Param page query int false "页码" minimum(1) default(1) example(1)
// @Param page_size query int false "每页数量" minimum(1) maximum(100) default(20) example(20)
// @Param order_by query string false "排序字段" default(created_at) example(created_at)
// @Param order query string false "排序方式(asc/desc)" Enums(asc, desc) default(desc) example(desc)
// @Success 200 {object} dto.PagedResponse{data=[]entity.AlarmRule} "获取成功"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /alarm-rules [get]
func (h *AlarmRuleHandler) ListAlarmRules(c *gin.Context) {
	query := &repository.AlarmRuleQuery{
		Name:      c.Query("name"),
		PointID:   c.Query("point_id"),
		DeviceID:  c.Query("device_id"),
		StationID: c.Query("station_id"),
		OrderBy:   c.Query("order_by"),
		Order:     c.Query("order"),
	}

	// 解析类型
	if ruleType := c.Query("type"); ruleType != "" {
		t := entity.AlarmRuleType(ruleType)
		query.Type = &t
	}

	// 解析级别
	if level := c.Query("level"); level != "" {
		if l, err := strconv.Atoi(level); err == nil {
			alarmLevel := entity.AlarmLevel(l)
			query.Level = &alarmLevel
		}
	}

	// 解析状态
	if status := c.Query("status"); status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			alarmStatus := entity.AlarmRuleStatus(s)
			query.Status = &alarmStatus
		}
	}

	// 解析分页参数
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			query.Page = p
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			query.PageSize = ps
		}
	}

	rules, total, err := h.ruleService.ListAlarmRules(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.PagedResponse{
		Code:      0,
		Message:   "success",
		Data:      rules,
		Total:     total,
		Page:      query.Page,
		PageSize:  query.PageSize,
		Timestamp: 0,
	})
}

// EnableAlarmRule 启用告警规则
// @Summary 启用告警规则
// @Description 启用指定的告警规则，启用后规则将开始监测并触发告警
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param id path string true "告警规则ID" example(alarm-rule-001)
// @Success 200 {object} dto.Response "启用成功"
// @Failure 404 {object} dto.ErrorResponse "告警规则不存在"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /alarm-rules/{id}/enable [put]
func (h *AlarmRuleHandler) EnableAlarmRule(c *gin.Context) {
	id := c.Param("id")

	// TODO: 从上下文中获取用户信息
	updatedBy := "system"

	if err := h.ruleService.EnableAlarmRule(c.Request.Context(), id, updatedBy); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:      404,
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

// DisableAlarmRule 禁用告警规则
// @Summary 禁用告警规则
// @Description 禁用指定的告警规则，禁用后规则将停止监测，不再触发告警
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param id path string true "告警规则ID" example(alarm-rule-001)
// @Success 200 {object} dto.Response "禁用成功"
// @Failure 404 {object} dto.ErrorResponse "告警规则不存在"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /alarm-rules/{id}/disable [put]
func (h *AlarmRuleHandler) DisableAlarmRule(c *gin.Context) {
	id := c.Param("id")

	// TODO: 从上下文中获取用户信息
	updatedBy := "system"

	if err := h.ruleService.DisableAlarmRule(c.Request.Context(), id, updatedBy); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:      404,
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

// GetRulesByPointID 根据采集点ID获取告警规则
// @Summary 根据采集点ID获取告警规则
// @Description 获取指定采集点的所有告警规则，包括启用和禁用的规则
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param point_id path string true "采集点ID" example(point-001)
// @Success 200 {object} dto.Response{data=[]entity.AlarmRule} "获取成功"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /points/{point_id}/alarm-rules [get]
func (h *AlarmRuleHandler) GetRulesByPointID(c *gin.Context) {
	pointID := c.Param("point_id")

	rules, err := h.ruleService.GetRulesByPointID(c.Request.Context(), pointID)
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
		Data:      rules,
		Timestamp: 0,
	})
}

// GetRulesByDeviceID 根据设备ID获取告警规则
// @Summary 根据设备ID获取告警规则
// @Description 获取指定设备的所有告警规则，包括启用和禁用的规则
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param device_id path string true "设备ID" example(device-001)
// @Success 200 {object} dto.Response{data=[]entity.AlarmRule} "获取成功"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /devices/{device_id}/alarm-rules [get]
func (h *AlarmRuleHandler) GetRulesByDeviceID(c *gin.Context) {
	deviceID := c.Param("device_id")

	rules, err := h.ruleService.GetRulesByDeviceID(c.Request.Context(), deviceID)
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
		Data:      rules,
		Timestamp: 0,
	})
}

// GetRulesByStationID 根据厂站ID获取告警规则
// @Summary 根据厂站ID获取告警规则
// @Description 获取指定厂站的所有告警规则，包括启用和禁用的规则
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param station_id path string true "厂站ID" example(station-001)
// @Success 200 {object} dto.Response{data=[]entity.AlarmRule} "获取成功"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /stations/{station_id}/alarm-rules [get]
func (h *AlarmRuleHandler) GetRulesByStationID(c *gin.Context) {
	stationID := c.Param("station_id")

	rules, err := h.ruleService.GetRulesByStationID(c.Request.Context(), stationID)
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
		Data:      rules,
		Timestamp: 0,
	})
}
