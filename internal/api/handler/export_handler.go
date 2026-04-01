package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/application/service"
)

// ExportHandler 导出处理器
type ExportHandler struct {
	exportService service.ExportServiceInterface
}

// NewExportHandler 创建导出处理器
func NewExportHandler(exportService service.ExportServiceInterface) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
	}
}

// Export 导出数据
// @Summary 导出数据
// @Description 导出告警、设备、厂站等数据，支持Excel和CSV格式。通过type参数指定导出数据类型，format参数指定导出格式
// @Tags 数据导出
// @Accept json
// @Produce application/octet-stream
// @Param request body service.ExportRequest true "导出请求参数"
// @Example request {"type":"alarm","format":"excel","start_time":1709500800000,"end_time":1709587200000,"filters":{"station_id":"station-001","level":3}}
// @Success 200 {file} file "导出文件"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误或导出类型/格式无效"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /export [post]
func (h *ExportHandler) Export(c *gin.Context) {
	var req service.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters: " + err.Error(),
			Timestamp: time.Now().UnixMilli(),
		})
		return
	}

	// 验证导出类型
	if req.Type != service.ExportTypeAlarm && req.Type != service.ExportTypeDevice && req.Type != service.ExportTypeStation {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid export type, must be one of: alarm, device, station",
			Timestamp: time.Now().UnixMilli(),
		})
		return
	}

	// 验证导出格式
	if req.Format != service.ExportFormatExcel && req.Format != service.ExportFormatCSV {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid export format, must be one of: excel, csv",
			Timestamp: time.Now().UnixMilli(),
		})
		return
	}

	// 执行导出
	result, err := h.exportService.Export(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   "Failed to export data: " + err.Error(),
			Timestamp: time.Now().UnixMilli(),
		})
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Disposition", "attachment; filename="+result.Filename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	// 返回文件
	c.Data(http.StatusOK, result.ContentType, result.Buffer.Bytes())
}

// ExportAlarms 导出告警数据
// @Summary 导出告警数据
// @Description 导出告警数据，支持Excel和CSV格式。可以按时间范围、厂站、告警级别等条件过滤
// @Tags 数据导出
// @Accept json
// @Produce application/octet-stream
// @Param format query string true "导出格式 (excel/csv)" Enums(excel, csv) example(excel)
// @Param start_time query int64 false "开始时间（毫秒时间戳）" example(1709500800000)
// @Param end_time query int64 false "结束时间（毫秒时间戳）" example(1709587200000)
// @Param station_id query string false "厂站ID" example(station-001)
// @Param level query int false "告警级别" minimum(1) maximum(4) example(3)
// @Success 200 {file} file "导出文件"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /export/alarms [get]
func (h *ExportHandler) ExportAlarms(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "excel"
	}

	req := &service.ExportRequest{
		Type:    service.ExportTypeAlarm,
		Format:  service.ExportFormat(format),
		Filters: make(map[string]interface{}),
	}

	if startTime := c.Query("start_time"); startTime != "" {
		var st int64
		if _, err := time.Parse(time.RFC3339, startTime); err == nil {
			req.StartTime = st
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		var et int64
		if _, err := time.Parse(time.RFC3339, endTime); err == nil {
			req.EndTime = et
		}
	}

	if stationID := c.Query("station_id"); stationID != "" {
		req.Filters["station_id"] = stationID
	}

	if level := c.Query("level"); level != "" {
		var l int
		if _, err := time.Parse(time.RFC3339, level); err == nil {
			req.Filters["level"] = l
		}
	}

	result, err := h.exportService.Export(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   "Failed to export alarms: " + err.Error(),
			Timestamp: time.Now().UnixMilli(),
		})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Disposition", "attachment; filename="+result.Filename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	c.Data(http.StatusOK, result.ContentType, result.Buffer.Bytes())
}

// ExportDevices 导出设备数据
// @Summary 导出设备数据
// @Description 导出设备数据，支持Excel和CSV格式。可以按厂站、设备类型等条件过滤
// @Tags 数据导出
// @Accept json
// @Produce application/octet-stream
// @Param format query string true "导出格式 (excel/csv)" Enums(excel, csv) example(excel)
// @Param station_id query string false "厂站ID" example(station-001)
// @Param type query string false "设备类型" example(inverter)
// @Success 200 {file} file "导出文件"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /export/devices [get]
func (h *ExportHandler) ExportDevices(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "excel"
	}

	req := &service.ExportRequest{
		Type:    service.ExportTypeDevice,
		Format:  service.ExportFormat(format),
		Filters: make(map[string]interface{}),
	}

	if stationID := c.Query("station_id"); stationID != "" {
		req.Filters["station_id"] = stationID
	}

	if deviceType := c.Query("type"); deviceType != "" {
		req.Filters["type"] = deviceType
	}

	result, err := h.exportService.Export(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   "Failed to export devices: " + err.Error(),
			Timestamp: time.Now().UnixMilli(),
		})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Disposition", "attachment; filename="+result.Filename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	c.Data(http.StatusOK, result.ContentType, result.Buffer.Bytes())
}

// ExportStations 导出厂站数据
// @Summary 导出厂站数据
// @Description 导出厂站数据，支持Excel和CSV格式。可以按子区域、厂站类型等条件过滤
// @Tags 数据导出
// @Accept json
// @Produce application/octet-stream
// @Param format query string true "导出格式 (excel/csv)" Enums(excel, csv) example(excel)
// @Param sub_region_id query string false "子区域ID" example(region-002)
// @Param type query string false "厂站类型" example(pv)
// @Success 200 {file} file "导出文件"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /export/stations [get]
func (h *ExportHandler) ExportStations(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "excel"
	}

	req := &service.ExportRequest{
		Type:    service.ExportTypeStation,
		Format:  service.ExportFormat(format),
		Filters: make(map[string]interface{}),
	}

	if subRegionID := c.Query("sub_region_id"); subRegionID != "" {
		req.Filters["sub_region_id"] = subRegionID
	}

	if stationType := c.Query("type"); stationType != "" {
		req.Filters["type"] = stationType
	}

	result, err := h.exportService.Export(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   "Failed to export stations: " + err.Error(),
			Timestamp: time.Now().UnixMilli(),
		})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Disposition", "attachment; filename="+result.Filename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	c.Data(http.StatusOK, result.ContentType, result.Buffer.Bytes())
}
