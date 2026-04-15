package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

type EnergyEfficiencyHandler struct {
	eeService *service.EnergyEfficiencyService
}

func NewEnergyEfficiencyHandler(eeService *service.EnergyEfficiencyService) *EnergyEfficiencyHandler {
	return &EnergyEfficiencyHandler{eeService: eeService}
}

func (h *EnergyEfficiencyHandler) CreateEnergyEfficiencyRecord(c *gin.Context) {
	var req service.CreateEnergyEfficiencyRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	record, err := h.eeService.CreateRecord(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: record})
}

func (h *EnergyEfficiencyHandler) BatchCreateEnergyEfficiencyRecords(c *gin.Context) {
	var reqs []*service.CreateEnergyEfficiencyRecordRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	records, err := h.eeService.BatchCreateRecords(c.Request.Context(), reqs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: records})
}

func (h *EnergyEfficiencyHandler) GetEnergyEfficiencyRecord(c *gin.Context) {
	id := c.Param("id")
	record, err := h.eeService.GetRecord(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "Energy efficiency record not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: record})
}

func (h *EnergyEfficiencyHandler) ListEnergyEfficiencyRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	req := &service.QueryEnergyEfficiencyRecordsRequest{
		Page:     page,
		PageSize: pageSize,
	}

	if eeType := c.Query("type"); eeType != "" {
		t := entity.EnergyEfficiencyType(eeType)
		req.Type = &t
	}
	if level := c.Query("level"); level != "" {
		l := entity.EnergyEfficiencyLevel(level)
		req.Level = &l
	}
	if targetID := c.Query("target_id"); targetID != "" {
		req.TargetID = &targetID
	}
	if period := c.Query("period"); period != "" {
		req.Period = &period
	}
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = &endTime
		}
	}

	records, total, err := h.eeService.ListRecords(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"list":       records,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
		},
	})
}

func (h *EnergyEfficiencyHandler) GetEnergyEfficiencyTrend(c *gin.Context) {
	targetID := c.Query("target_id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "target_id is required"})
		return
	}

	eeType := entity.EnergyEfficiencyType(c.DefaultQuery("type", string(entity.EnergyEfficiencyTypeDevice)))
	
	startTimeStr := c.DefaultQuery("start_time", time.Now().AddDate(0, -1, 0).Format(time.RFC3339))
	endTimeStr := c.DefaultQuery("end_time", time.Now().Format(time.RFC3339))
	
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "invalid start_time format"})
		return
	}
	
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "invalid end_time format"})
		return
	}

	trendData, err := h.eeService.GetTrendData(c.Request.Context(), targetID, eeType, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: trendData})
}

func (h *EnergyEfficiencyHandler) GetEnergyEfficiencyStatistics(c *gin.Context) {
	targetID := c.Query("target_id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "target_id is required"})
		return
	}

	eeType := entity.EnergyEfficiencyType(c.DefaultQuery("type", string(entity.EnergyEfficiencyTypeDevice)))
	period := c.DefaultQuery("period", "month")
	
	startTimeStr := c.DefaultQuery("start_time", time.Now().AddDate(0, -1, 0).Format(time.RFC3339))
	endTimeStr := c.DefaultQuery("end_time", time.Now().Format(time.RFC3339))
	
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "invalid start_time format"})
		return
	}
	
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "invalid end_time format"})
		return
	}

	stats, err := h.eeService.GetStatistics(c.Request.Context(), targetID, eeType, period, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: stats})
}

func (h *EnergyEfficiencyHandler) GetEnergyEfficiencyComparison(c *gin.Context) {
	targetID := c.Query("target_id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "target_id is required"})
		return
	}

	eeType := entity.EnergyEfficiencyType(c.DefaultQuery("type", string(entity.EnergyEfficiencyTypeDevice)))
	period := c.DefaultQuery("period", "month")
	
	currentStartStr := c.DefaultQuery("current_start", time.Now().AddDate(0, -1, 0).Format(time.RFC3339))
	currentEndStr := c.DefaultQuery("current_end", time.Now().Format(time.RFC3339))
	
	currentStart, err := time.Parse(time.RFC3339, currentStartStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "invalid current_start format"})
		return
	}
	
	currentEnd, err := time.Parse(time.RFC3339, currentEndStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "invalid current_end format"})
		return
	}

	comparison, err := h.eeService.GetComparisonData(c.Request.Context(), targetID, eeType, period, currentStart, currentEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: comparison})
}

func (h *EnergyEfficiencyHandler) CreateEnergyEfficiencyAnalysis(c *gin.Context) {
	var req struct {
		TargetID       string    `json:"target_id" binding:"required"`
		TargetName     string    `json:"target_name" binding:"required"`
		Type           string    `json:"type" binding:"required"`
		TimeRangeStart time.Time `json:"time_range_start" binding:"required"`
		TimeRangeEnd   time.Time `json:"time_range_end" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	eeType := entity.EnergyEfficiencyType(req.Type)
	analysis, err := h.eeService.CreateAnalysis(c.Request.Context(), req.TargetID, eeType, req.TargetName, req.TimeRangeStart, req.TimeRangeEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: analysis})
}

func (h *EnergyEfficiencyHandler) GetEnergyEfficiencyAnalysis(c *gin.Context) {
	id := c.Param("id")
	analysis, err := h.eeService.GetAnalysis(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "Energy efficiency analysis not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: analysis})
}

func (h *EnergyEfficiencyHandler) ListEnergyEfficiencyAnalyses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	req := &service.QueryEnergyEfficiencyAnalysesRequest{
		Page:     page,
		PageSize: pageSize,
	}

	if eeType := c.Query("type"); eeType != "" {
		t := entity.EnergyEfficiencyType(eeType)
		req.Type = &t
	}
	if targetID := c.Query("target_id"); targetID != "" {
		req.TargetID = &targetID
	}
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = &endTime
		}
	}

	analyses, total, err := h.eeService.ListAnalyses(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"list":       analyses,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
		},
	})
}

func (h *EnergyEfficiencyHandler) GetLatestEnergyEfficiencyAnalysis(c *gin.Context) {
	targetID := c.Query("target_id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "target_id is required"})
		return
	}

	eeType := entity.EnergyEfficiencyType(c.DefaultQuery("type", string(entity.EnergyEfficiencyTypeDevice)))
	
	analysis, err := h.eeService.GetLatestAnalysis(c.Request.Context(), targetID, eeType)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "No analysis found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: analysis})
}
