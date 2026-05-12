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

type ForecastHandler struct {
	forecastService service.ForecastService
}

func NewForecastHandler(forecastService service.ForecastService) *ForecastHandler {
	return &ForecastHandler{forecastService: forecastService}
}

func (h *ForecastHandler) PowerForecast(c *gin.Context) {
	var req struct {
		StationID    string `json:"station_id" binding:"required"`
		ForecastType string `json:"forecast_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	results, err := h.forecastService.PowerForecast(c.Request.Context(), req.StationID, entity.ForecastType(req.ForecastType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: results})
}

func (h *ForecastHandler) GetResults(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "station_id is required"})
		return
	}
	var forecastType *entity.ForecastType
	if ft := c.Query("forecast_type"); ft != "" {
		ftVal := entity.ForecastType(ft)
		forecastType = &ftVal
	}
	var startTime, endTime *time.Time
	if st := c.Query("start_time"); st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			startTime = &t
		}
	}
	if et := c.Query("end_time"); et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			endTime = &t
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	results, total, err := h.forecastService.GetResults(c.Request.Context(), stationID, forecastType, startTime, endTime, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PagedResponse{Code: 0, Message: "success", Data: results, Total: total, Page: page, PageSize: pageSize})
}

func (h *ForecastHandler) GetAccuracy(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "station_id is required"})
		return
	}
	var forecastType *entity.ForecastType
	if ft := c.Query("forecast_type"); ft != "" {
		ftVal := entity.ForecastType(ft)
		forecastType = &ftVal
	}
	startStr := c.Query("start_time")
	endStr := c.Query("end_time")
	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "start_time and end_time are required"})
		return
	}
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "invalid start_time format"})
		return
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "invalid end_time format"})
		return
	}
	stats, err := h.forecastService.GetAccuracy(c.Request.Context(), stationID, forecastType, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: stats})
}

func (h *ForecastHandler) EvaluateModel(c *gin.Context) {
	var req struct {
		StationID         string  `json:"station_id" binding:"required"`
		ForecastType      string  `json:"forecast_type"`
		StartTime         string  `json:"start_time" binding:"required"`
		EndTime           string  `json:"end_time" binding:"required"`
		InstalledCapacity float64 `json:"installed_capacity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "invalid start_time format"})
		return
	}
	end, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "invalid end_time format"})
		return
	}
	var forecastType *entity.ForecastType
	if req.ForecastType != "" {
		ftVal := entity.ForecastType(req.ForecastType)
		forecastType = &ftVal
	}
	report, err := h.forecastService.EvaluateModel(c.Request.Context(), req.StationID, forecastType, start, end, req.InstalledCapacity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: report})
}

func (h *ForecastHandler) AttributionAnalysis(c *gin.Context) {
	var req struct {
		StationID      string  `json:"station_id" binding:"required"`
		TargetTime     string  `json:"target_time" binding:"required"`
		PredictedPower float64 `json:"predicted_power" binding:"required"`
		ActualPower    float64 `json:"actual_power" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	result, err := h.forecastService.AttributionAnalysis(c.Request.Context(), req.StationID, req.TargetTime, req.PredictedPower, req.ActualPower)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: result})
}
