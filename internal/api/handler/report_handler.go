package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/pkg/errors"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

func (h *ReportHandler) GenerateReport(c *gin.Context) {
	var req struct {
		Type      string `form:"type" binding:"required,oneof=daily weekly monthly yearly"`
		StartTime string `form:"start_time" binding:"required"`
		EndTime   string `form:"end_time" binding:"required"`
		StationID string `form:"station_id"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError(err.Error()))
		return
	}

	startTime, err := time.Parse("2006-01-02", req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("invalid start_time format"))
		return
	}

	endTime, err := time.Parse("2006-01-02", req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("invalid end_time format"))
		return
	}

	reportReq := &service.ReportRequest{
		Type:      service.ReportType(req.Type),
		StartTime: startTime,
		EndTime:   endTime,
		StationID: req.StationID,
	}

	report, err := h.reportService.GenerateStationReport(c.Request.Context(), reportReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewSystemError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    report,
	})
}

func (h *ReportHandler) ExportReport(c *gin.Context) {
	var req struct {
		Type      string `form:"type" binding:"required,oneof=daily weekly monthly yearly"`
		StartTime string `form:"start_time" binding:"required"`
		EndTime   string `form:"end_time" binding:"required"`
		StationID string `form:"station_id"`
		Format    string `form:"format" binding:"required,oneof=excel csv"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError(err.Error()))
		return
	}

	startTime, err := time.Parse("2006-01-02", req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("invalid start_time format"))
		return
	}

	endTime, err := time.Parse("2006-01-02", req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("invalid end_time format"))
		return
	}

	reportReq := &service.ReportRequest{
		Type:      service.ReportType(req.Type),
		StartTime: startTime,
		EndTime:   endTime,
		StationID: req.StationID,
	}

	data, filename, err := h.reportService.ExportReport(c.Request.Context(), reportReq, req.Format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewSystemError(err.Error()))
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", data)
}
