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

type CarbonEmissionHandler struct {
	ceService *service.CarbonEmissionService
}

func NewCarbonEmissionHandler(ceService *service.CarbonEmissionService) *CarbonEmissionHandler {
	return &CarbonEmissionHandler{ceService: ceService}
}

func (h *CarbonEmissionHandler) CreateCarbonEmissionFactor(c *gin.Context) {
	var req service.CreateCarbonEmissionFactorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	userID := c.GetString("user_id")
	factor, err := h.ceService.CreateFactor(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: factor})
}

func (h *CarbonEmissionHandler) UpdateCarbonEmissionFactor(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateCarbonEmissionFactorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	userID := c.GetString("user_id")
	factor, err := h.ceService.UpdateFactor(c.Request.Context(), id, &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: factor})
}

func (h *CarbonEmissionHandler) GetCarbonEmissionFactor(c *gin.Context) {
	id := c.Param("id")
	factor, err := h.ceService.GetFactor(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "Carbon emission factor not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: factor})
}

func (h *CarbonEmissionHandler) ListCarbonEmissionFactors(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	req := &service.QueryCarbonEmissionFactorsRequest{
		Page:     page,
		PageSize: pageSize,
	}

	if scope := c.Query("scope"); scope != "" {
		s := entity.CarbonEmissionScope(scope)
		req.Scope = &s
	}
	if source := c.Query("source"); source != "" {
		req.Source = &source
	}
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		req.IsActive = &isActive
	}

	factors, total, err := h.ceService.ListFactors(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"list":       factors,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
		},
	})
}

func (h *CarbonEmissionHandler) GetActiveCarbonEmissionFactors(c *gin.Context) {
	var scope *entity.CarbonEmissionScope
	if scopeStr := c.Query("scope"); scopeStr != "" {
		s := entity.CarbonEmissionScope(scopeStr)
		scope = &s
	}

	factors, err := h.ceService.GetActiveFactors(c.Request.Context(), scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: factors})
}

func (h *CarbonEmissionHandler) CreateCarbonEmissionRecord(c *gin.Context) {
	var req service.CreateCarbonEmissionRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	userID := c.GetString("user_id")
	record, err := h.ceService.CreateRecord(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: record})
}

func (h *CarbonEmissionHandler) BatchCreateCarbonEmissionRecords(c *gin.Context) {
	var reqs []*service.CreateCarbonEmissionRecordRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	userID := c.GetString("user_id")
	records, err := h.ceService.BatchCreateRecords(c.Request.Context(), reqs, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: records})
}

func (h *CarbonEmissionHandler) GetCarbonEmissionRecord(c *gin.Context) {
	id := c.Param("id")
	record, err := h.ceService.GetRecord(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "Carbon emission record not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: record})
}

func (h *CarbonEmissionHandler) ListCarbonEmissionRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	req := &service.QueryCarbonEmissionRecordsRequest{
		Page:      page,
		PageSize:  pageSize,
	}

	if scope := c.Query("scope"); scope != "" {
		s := entity.CarbonEmissionScope(scope)
		req.Scope = &s
	}
	if targetID := c.Query("target_id"); targetID != "" {
		req.TargetID = &targetID
	}
	if period := c.Query("period"); period != "" {
		req.Period = &period
	}
	if status := c.Query("status"); status != "" {
		s := entity.CarbonEmissionStatus(status)
		req.Status = &s
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

	records, total, err := h.ceService.ListRecords(c.Request.Context(), req)
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

func (h *CarbonEmissionHandler) GetCarbonEmissionTrend(c *gin.Context) {
	targetID := c.Query("target_id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "target_id is required"})
		return
	}

	period := c.DefaultQuery("period", "month")
	startTimeStr := c.DefaultQuery("start_time", time.Now().AddDate(0, -6, 0).Format(time.RFC3339))
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

	trendData, err := h.ceService.GetTrendData(c.Request.Context(), targetID, period, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: trendData})
}

func (h *CarbonEmissionHandler) CreateCarbonEmissionSummary(c *gin.Context) {
	var req struct {
		TargetID     string    `json:"target_id" binding:"required"`
		TargetName   string    `json:"target_name" binding:"required"`
		Period       string    `json:"period" binding:"required"`
		PeriodStart  time.Time `json:"period_start" binding:"required"`
		PeriodEnd    time.Time `json:"period_end" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	userID := c.GetString("user_id")
	summary, err := h.ceService.CreateSummary(c.Request.Context(), req.TargetID, req.TargetName, req.Period, req.PeriodStart, req.PeriodEnd, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: summary})
}

func (h *CarbonEmissionHandler) GetCarbonEmissionSummary(c *gin.Context) {
	id := c.Param("id")
	summary, err := h.ceService.GetSummary(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "Carbon emission summary not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: summary})
}

func (h *CarbonEmissionHandler) ListCarbonEmissionSummaries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	req := &service.QueryCarbonEmissionSummariesRequest{
		Page:      page,
		PageSize:  pageSize,
	}

	if targetID := c.Query("target_id"); targetID != "" {
		req.TargetID = &targetID
	}
	if period := c.Query("period"); period != "" {
		req.Period = &period
	}
	if status := c.Query("status"); status != "" {
		s := entity.CarbonEmissionStatus(status)
		req.Status = &s
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

	summaries, total, err := h.ceService.ListSummaries(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"list":       summaries,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
		},
	})
}

func (h *CarbonEmissionHandler) CreateCarbonReductionTarget(c *gin.Context) {
	var req service.CreateCarbonReductionTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	userID := c.GetString("user_id")
	target, err := h.ceService.CreateTarget(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: target})
}

func (h *CarbonEmissionHandler) UpdateCarbonReductionTarget(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateCarbonReductionTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	userID := c.GetString("user_id")
	target, err := h.ceService.UpdateTarget(c.Request.Context(), id, &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: target})
}

func (h *CarbonEmissionHandler) GetCarbonReductionTarget(c *gin.Context) {
	id := c.Param("id")
	target, err := h.ceService.GetTarget(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "Carbon reduction target not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: target})
}

func (h *CarbonEmissionHandler) ListCarbonReductionTargets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	req := &service.QueryCarbonReductionTargetsRequest{
		Page:     page,
		PageSize: pageSize,
	}

	if targetID := c.Query("target_id"); targetID != "" {
		req.TargetID = &targetID
	}
	if status := c.Query("status"); status != "" {
		req.Status = &status
	}
	if startYearStr := c.Query("start_year"); startYearStr != "" {
		startYear, _ := strconv.Atoi(startYearStr)
		req.StartYear = &startYear
	}
	if endYearStr := c.Query("end_year"); endYearStr != "" {
		endYear, _ := strconv.Atoi(endYearStr)
		req.EndYear = &endYear
	}

	targets, total, err := h.ceService.ListTargets(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"list":       targets,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
		},
	})
}

func (h *CarbonEmissionHandler) GetActiveCarbonReductionTargets(c *gin.Context) {
	targetID := c.Query("target_id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "target_id is required"})
		return
	}

	targets, err := h.ceService.GetActiveTargets(c.Request.Context(), targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: targets})
}
