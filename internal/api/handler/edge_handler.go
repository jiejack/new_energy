package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

type EdgeHandler struct {
	edgeSvc service.EdgeService
}

func NewEdgeHandler(edgeSvc service.EdgeService) *EdgeHandler {
	return &EdgeHandler{edgeSvc: edgeSvc}
}

func (h *EdgeHandler) ListNodes(c *gin.Context) {
	var stationID *string
	if s := c.Query("station_id"); s != "" {
		stationID = &s
	}
	var status *entity.EdgeNodeStatus
	if st := c.Query("status"); st != "" {
		sv := entity.EdgeNodeStatus(st)
		status = &sv
	}
	nodes, err := h.edgeSvc.ListNodes(c.Request.Context(), stationID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: nodes})
}

func (h *EdgeHandler) GetNode(c *gin.Context) {
	id := c.Param("id")
	node, err := h.edgeSvc.GetNode(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "node not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: node})
}

func (h *EdgeHandler) RegisterNode(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		StationID string `json:"station_id" binding:"required"`
		IPAddress string `json:"ip_address" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	node, err := h.edgeSvc.RegisterNode(c.Request.Context(), req.Name, req.StationID, req.IPAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.Response{Code: 0, Message: "success", Data: node})
}

func (h *EdgeHandler) UpdateNodeConfig(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ConfigVersion string `json:"config_version" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	err := h.edgeSvc.UpdateNodeConfig(c.Request.Context(), id, req.ConfigVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
}

func (h *EdgeHandler) DeployModel(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ModelName string `json:"model_name" binding:"required"`
		Version   string `json:"version" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	err := h.edgeSvc.DeployModel(c.Request.Context(), id, req.ModelName, req.Version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
}

func (h *EdgeHandler) GetNodeStatus(c *gin.Context) {
	id := c.Param("id")
	node, err := h.edgeSvc.GetNodeStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "node not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: node})
}

func (h *EdgeHandler) TriggerSync(c *gin.Context) {
	id := c.Param("id")
	err := h.edgeSvc.TriggerSync(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
}
