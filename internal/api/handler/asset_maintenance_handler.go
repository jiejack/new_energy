package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/api/dto"
)

type AssetMaintenanceHandler struct {
	service *service.AssetMaintenanceService
}

func NewAssetMaintenanceHandler(service *service.AssetMaintenanceService) *AssetMaintenanceHandler {
	return &AssetMaintenanceHandler{service: service}
}

// CreateMaintenanceRecord godoc
// @Summary Create a new maintenance record
// @Description Create a new maintenance record for an asset
// @Tags asset-maintenance
// @Accept json
// @Produce json
// @Param record body service.CreateMaintenanceRequest true "Maintenance record details"
// @Success 201 {object} dto.AssetMaintenanceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/maintenance [post]
func (h *AssetMaintenanceHandler) CreateMaintenanceRecord(c *gin.Context) {
	var req service.CreateMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	record, err := h.service.CreateMaintenanceRecord(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

// GetMaintenanceRecord godoc
// @Summary Get a maintenance record by ID
// @Description Get maintenance record details by its ID
// @Tags asset-maintenance
// @Accept json
// @Produce json
// @Param id path string true "Maintenance record ID"
// @Success 200 {object} dto.AssetMaintenanceResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/maintenance/{id} [get]
func (h *AssetMaintenanceHandler) GetMaintenanceRecord(c *gin.Context) {
	id := c.Param("id")

	record, err := h.service.GetMaintenanceRecord(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

// UpdateMaintenanceRecord godoc
// @Summary Update a maintenance record
// @Description Update an existing maintenance record with the provided details
// @Tags asset-maintenance
// @Accept json
// @Produce json
// @Param id path string true "Maintenance record ID"
// @Param record body service.UpdateMaintenanceRequest true "Maintenance record details"
// @Success 200 {object} dto.AssetMaintenanceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/maintenance/{id} [put]
func (h *AssetMaintenanceHandler) UpdateMaintenanceRecord(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	record, err := h.service.UpdateMaintenanceRecord(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

// DeleteMaintenanceRecord godoc
// @Summary Delete a maintenance record
// @Description Delete an existing maintenance record by its ID
// @Tags asset-maintenance
// @Accept json
// @Produce json
// @Param id path string true "Maintenance record ID"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/maintenance/{id} [delete]
func (h *AssetMaintenanceHandler) DeleteMaintenanceRecord(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteMaintenanceRecord(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListMaintenanceRecords godoc
// @Summary List all maintenance records
// @Description Get a list of all maintenance records with optional filters
// @Tags asset-maintenance
// @Accept json
// @Produce json
// @Param asset_id query string false "Asset ID"
// @Param type query string false "Maintenance type"
// @Param status query string false "Maintenance status"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.AssetMaintenanceListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/maintenance [get]
func (h *AssetMaintenanceHandler) ListMaintenanceRecords(c *gin.Context) {
	assetID := c.Query("asset_id")
	maintenanceType := c.Query("type")
	status := c.Query("status")
	page := dto.GetIntQuery(c, "page", 1)
	pageSize := dto.GetIntQuery(c, "page_size", 10)

	records, total, err := h.service.ListMaintenanceRecords(c.Request.Context(), assetID, maintenanceType, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.AssetMaintenanceListResponse{
		Items: records,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}

// GetMaintenanceCosts godoc
// @Summary Get maintenance costs for an asset
// @Description Get total maintenance costs for a specific asset within a date range
// @Tags asset-maintenance
// @Accept json
// @Produce json
// @Param asset_id path string true "Asset ID"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} dto.MaintenanceCostResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/{asset_id}/maintenance/costs [get]
func (h *AssetMaintenanceHandler) GetMaintenanceCosts(c *gin.Context) {
	assetID := c.Param("asset_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	cost, err := h.service.GetMaintenanceCosts(c.Request.Context(), assetID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MaintenanceCostResponse{
		AssetID: assetID,
		Cost:    cost,
	})
}