package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/api/dto"
)

type AssetHandler struct {
	service *service.AssetService
}

func NewAssetHandler(service *service.AssetService) *AssetHandler {
	return &AssetHandler{service: service}
}

// CreateAsset godoc
// @Summary Create a new asset
// @Description Create a new asset with the provided details
// @Tags assets
// @Accept json
// @Produce json
// @Param asset body dto.AssetRequest true "Asset details"
// @Success 201 {object} dto.AssetResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets [post]
func (h *AssetHandler) CreateAsset(c *gin.Context) {
	var req dto.AssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	serviceReq := &service.CreateAssetRequest{
		Code:          req.Code,
		Name:          req.Name,
		Category:      req.Category,
		AssetType:     req.AssetType,
		Manufacturer:  req.Manufacturer,
		Model:         req.Model,
		SerialNumber:  req.SerialNumber,
		PurchasePrice: req.PurchasePrice,
		PurchaseDate:  req.PurchaseDate,
		ExpectedLife:  req.ExpectedLife,
		ResidualValue: req.ResidualValue,
		Location:      req.Location,
		Status:        req.Status,
		Description:   req.Description,
	}

	asset, err := h.service.CreateAsset(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, asset)
}

// GetAsset godoc
// @Summary Get an asset by ID
// @Description Get asset details by its ID
// @Tags assets
// @Accept json
// @Produce json
// @Param id path string true "Asset ID"
// @Success 200 {object} dto.AssetResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/{id} [get]
func (h *AssetHandler) GetAsset(c *gin.Context) {
	id := c.Param("id")

	asset, err := h.service.GetAsset(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, asset)
}

// UpdateAsset godoc
// @Summary Update an asset
// @Description Update an existing asset with the provided details
// @Tags assets
// @Accept json
// @Produce json
// @Param id path string true "Asset ID"
// @Param asset body dto.AssetRequest true "Asset details"
// @Success 200 {object} dto.AssetResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/{id} [put]
func (h *AssetHandler) UpdateAsset(c *gin.Context) {
	id := c.Param("id")
	var req dto.AssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	serviceReq := &service.UpdateAssetRequest{
		Code:          req.Code,
		Name:          req.Name,
		Category:      req.Category,
		AssetType:     req.AssetType,
		Manufacturer:  req.Manufacturer,
		Model:         req.Model,
		SerialNumber:  req.SerialNumber,
		PurchasePrice: req.PurchasePrice,
		PurchaseDate:  req.PurchaseDate,
		ExpectedLife:  req.ExpectedLife,
		ResidualValue: req.ResidualValue,
		Location:      req.Location,
		Status:        req.Status,
		Description:   req.Description,
	}

	asset, err := h.service.UpdateAsset(c.Request.Context(), id, serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, asset)
}

// DeleteAsset godoc
// @Summary Delete an asset
// @Description Delete an existing asset by its ID
// @Tags assets
// @Accept json
// @Produce json
// @Param id path string true "Asset ID"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/{id} [delete]
func (h *AssetHandler) DeleteAsset(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteAsset(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListAssets godoc
// @Summary List all assets
// @Description Get a list of all assets with optional filters
// @Tags assets
// @Accept json
// @Produce json
// @Param name query string false "Asset name"
// @Param category query string false "Asset category"
// @Param status query string false "Asset status"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.AssetListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets [get]
func (h *AssetHandler) ListAssets(c *gin.Context) {
	name := c.Query("name")
	category := c.Query("category")
	status := c.Query("status")
	page := dto.GetIntQuery(c, "page", 1)
	pageSize := dto.GetIntQuery(c, "page_size", 10)

	assets, total, err := h.service.ListAssets(c.Request.Context(), name, category, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.AssetListResponse{
		Items: assets,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}

// CalculateDepreciation godoc
// @Summary Calculate asset depreciation
// @Description Calculate depreciation for an asset using the specified method
// @Tags assets
// @Accept json
// @Produce json
// @Param id path string true "Asset ID"
// @Param method query string true "Depreciation method (straight-line or declining-balance)"
// @Success 200 {object} dto.DepreciationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/{id}/depreciation [get]
func (h *AssetHandler) CalculateDepreciation(c *gin.Context) {
	id := c.Param("id")
	method := c.Query("method")

	if method == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "depreciation method is required"})
		return
	}

	depreciation, err := h.service.CalculateDepreciation(c.Request.Context(), id, method)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, depreciation)
}