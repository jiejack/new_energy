package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type OperationLogHandler struct {
	logService *service.OperationLogService
}

func NewOperationLogHandler(logService *service.OperationLogService) *OperationLogHandler {
	return &OperationLogHandler{logService: logService}
}

func (h *OperationLogHandler) CreateLog(c *gin.Context) {
	var req service.CreateOperationLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	req.IPAddress = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	log, err := h.logService.CreateLog(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: log})
}

func (h *OperationLogHandler) GetLog(c *gin.Context) {
	id := c.Param("id")
	log, err := h.logService.GetLog(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "Log not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: log})
}

func (h *OperationLogHandler) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	query := &repository.OperationLogQuery{
		Page:     page,
		PageSize: pageSize,
	}

	if userID := c.Query("user_id"); userID != "" {
		query.UserID = userID
	}
	if username := c.Query("username"); username != "" {
		query.Username = username
	}
	if action := c.Query("action"); action != "" {
		query.Action = action
	}
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := strconv.ParseInt(startTime, 10, 64); err == nil {
			query.StartTime = t
		}
	}
	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := strconv.ParseInt(endTime, 10, 64); err == nil {
			query.EndTime = t
		}
	}

	result, err := h.logService.ListLogs(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: result})
}

func (h *OperationLogHandler) DeleteOldLogs(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	count, err := h.logService.DeleteOldLogs(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"deleted_count": count,
		},
	})
}
