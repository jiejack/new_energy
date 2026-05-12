package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

type FaultHandler struct {
	faultService service.FaultService
}

func NewFaultHandler(faultService service.FaultService) *FaultHandler {
	return &FaultHandler{faultService: faultService}
}

func (h *FaultHandler) DetectFaults(c *gin.Context) {
	var req struct {
		DeviceID string `json:"device_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	results, err := h.faultService.DetectFaults(c.Request.Context(), req.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: results})
}

func (h *FaultHandler) GetDetections(c *gin.Context) {
	deviceID := c.Query("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "device_id is required"})
		return
	}
	var severity *entity.FaultSeverity
	if s := c.Query("severity"); s != "" {
		sv := entity.FaultSeverity(s)
		severity = &sv
	}
	var status *entity.FaultDetectionStatus
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			sv := entity.FaultDetectionStatus(v)
			status = &sv
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	results, total, err := h.faultService.GetDetections(c.Request.Context(), deviceID, severity, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PagedResponse{Code: 0, Message: "success", Data: results, Total: total, Page: page, PageSize: pageSize})
}

func (h *FaultHandler) GetDetectionByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.faultService.GetDetectionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "fault detection result not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: result})
}

func (h *FaultHandler) UpdateDetectionStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	if err := h.faultService.UpdateDetectionStatus(c.Request.Context(), id, entity.FaultDetectionStatus(req.Status)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
}

func (h *FaultHandler) GetDeviceHealth(c *gin.Context) {
	deviceID := c.Param("device_id")
	score, err := h.faultService.GetDeviceHealth(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: map[string]interface{}{"device_id": deviceID, "health_score": score}})
}

func (h *FaultHandler) AnalyzeRootCause(c *gin.Context) {
	var req struct {
		DetectionID string `json:"detection_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	result, err := h.faultService.AnalyzeRootCause(c.Request.Context(), req.DetectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: result})
}

func (h *FaultHandler) CreateWorkOrderFromDetection(c *gin.Context) {
	var req struct {
		DetectionID string `json:"detection_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	workOrderID, err := h.faultService.CreateWorkOrderFromDetection(c.Request.Context(), req.DetectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: map[string]interface{}{"work_order_id": workOrderID}})
}
