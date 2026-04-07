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

type AlarmRuleHandler struct {
	ruleService *service.AlarmRuleService
}

func NewAlarmRuleHandler(ruleService *service.AlarmRuleService) *AlarmRuleHandler {
	return &AlarmRuleHandler{ruleService: ruleService}
}

func (h *AlarmRuleHandler) CreateAlarmRule(c *gin.Context) {
	var req service.CreateAlarmRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	userID := c.GetString("user_id")
	rule, err := h.ruleService.CreateRule(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: rule})
}

func (h *AlarmRuleHandler) GetAlarmRule(c *gin.Context) {
	id := c.Param("id")
	rule, err := h.ruleService.GetRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "Alarm rule not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: rule})
}

func (h *AlarmRuleHandler) ListAlarmRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	query := &repository.AlarmRuleQuery{Page: page, PageSize: pageSize}

	if ruleType := c.Query("type"); ruleType != "" {
		t := entity.AlarmRuleType(ruleType)
		query.Type = &t
	}
	if level := c.Query("level"); level != "" {
		l, _ := strconv.Atoi(level)
		al := entity.AlarmLevel(l)
		query.Level = &al
	}
	if status := c.Query("status"); status != "" {
		s, _ := strconv.Atoi(status)
		rs := entity.AlarmRuleStatus(s)
		query.Status = &rs
	}
	if stationID := c.Query("station_id"); stationID != "" {
		query.StationID = &stationID
	}

	rules, total, err := h.ruleService.ListRules(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"list":      rules,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (h *AlarmRuleHandler) UpdateAlarmRule(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateAlarmRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	userID := c.GetString("user_id")
	rule, err := h.ruleService.UpdateRule(c.Request.Context(), id, &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: rule})
}

func (h *AlarmRuleHandler) DeleteAlarmRule(c *gin.Context) {
	id := c.Param("id")
	if err := h.ruleService.DeleteRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "Alarm rule not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
}

func (h *AlarmRuleHandler) EnableAlarmRule(c *gin.Context) {
	id := c.Param("id")
	status := entity.AlarmRuleStatusEnabled
	req := &service.UpdateAlarmRuleRequest{Status: &status}
	
	userID := c.GetString("user_id")
	_, err := h.ruleService.UpdateRule(c.Request.Context(), id, req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
}

func (h *AlarmRuleHandler) DisableAlarmRule(c *gin.Context) {
	id := c.Param("id")
	status := entity.AlarmRuleStatusDisabled
	req := &service.UpdateAlarmRuleRequest{Status: &status}
	
	userID := c.GetString("user_id")
	_, err := h.ruleService.UpdateRule(c.Request.Context(), id, req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
}
