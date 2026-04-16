package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/api/dto"
)

type AssetDepreciationHandler struct {
	service *service.AssetDepreciationService
}

func NewAssetDepreciationHandler(service *service.AssetDepreciationService) *AssetDepreciationHandler {
	return &AssetDepreciationHandler{service: service}
}

// CreateDepreciationRecord godoc
// @Summary Create a new depreciation record
// @Description Create a new depreciation record for an asset
// @Tags asset-depreciation
// @Accept json
// @Produce json
// @Param record body dto.AssetDepreciationRequest true "Depreciation record details"
// @Success 201 {object} dto.AssetDepreciationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/depreciation [post]
func (h *AssetDepreciationHandler) CreateDepreciationRecord(c *gin.Context) {
	var req dto.AssetDepreciationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	record, err := h.service.CreateDepreciationRecord(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

// GetDepreciationRecord godoc
// @Summary Get a depreciation record by ID
// @Description Get depreciation record details by its ID
// @Tags asset-depreciation
// @Accept json
// @Produce json
// @Param id path string true "Depreciation record ID"
// @Success 200 {object} dto.AssetDepreciationResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/depreciation/{id} [get]
func (h *AssetDepreciationHandler) GetDepreciationRecord(c *gin.Context) {
	id := c.Param("id")

	record, err := h.service.GetDepreciationRecord(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

// UpdateDepreciationRecord godoc
// @Summary Update a depreciation record
// @Description Update an existing depreciation record with the provided details
// @Tags asset-depreciation
// @Accept json
// @Produce json
// @Param id path string true "Depreciation record ID"
// @Param record body dto.AssetDepreciationRequest true "Depreciation record details"
// @Success 200 {object} dto.AssetDepreciationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/depreciation/{id} [put]
func (h *AssetDepreciationHandler) UpdateDepreciationRecord(c *gin.Context) {
	id := c.Param("id")
	var req dto.AssetDepreciationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	record, err := h.service.UpdateDepreciationRecord(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

// DeleteDepreciationRecord godoc
// @Summary Delete a depreciation record
// @Description Delete an existing depreciation record by its ID
// @Tags asset-depreciation
// @Accept json
// @Produce json
// @Param id path string true "Depreciation record ID"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/depreciation/{id} [delete]
func (h *AssetDepreciationHandler) DeleteDepreciationRecord(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteDepreciationRecord(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListDepreciationRecords godoc
// @Summary List all depreciation records
// @Description Get a list of all depreciation records with optional filters
// @Tags asset-depreciation
// @Accept json
// @Produce json
// @Param asset_id query string false "Asset ID"
// @Param method query string false "Depreciation method"
// @Param year query int false "Depreciation year"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.AssetDepreciationListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/depreciation [get]
func (h *AssetDepreciationHandler) ListDepreciationRecords(c *gin.Context) {
	assetID := c.Query("asset_id")
	method := c.Query("method")
	year := dto.GetIntQuery(c, "year", 0)
	page := dto.GetIntQuery(c, "page", 1)
	pageSize := dto.GetIntQuery(c, "page_size", 10)

	records, total, err := h.service.ListDepreciationRecords(c.Request.Context(), assetID, method, year, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.AssetDepreciationListResponse{
		Items: records,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}

// GetDepreciationSummary godoc
// @Summary Get depreciation summary for an asset
// @Description Get total depreciation for a specific asset up to a certain date
// @Tags asset-depreciation
// @Accept json
// @Produce json
// @Param asset_id path string true "Asset ID"
// @Param up_to_date query string false "Up to date (YYYY-MM-DD)"
// @Success 200 {object} dto.DepreciationSummaryResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/{asset_id}/depreciation/summary [get]
func (h *AssetDepreciationHandler) GetDepreciationSummary(c *gin.Context) {
	assetID := c.Param("asset_id")
	upToDate := c.Query("up_to_date")

	summary, err := h.service.GetDepreciationSummary(c.Request.Context(), assetID, upToDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.DepreciationSummaryResponse{
		AssetID:          assetID,
		TotalDepreciation: summary,
	})
}