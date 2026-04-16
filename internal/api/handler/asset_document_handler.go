package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/api/dto"
)

type AssetDocumentHandler struct {
	service *service.AssetDocumentService
}

func NewAssetDocumentHandler(service *service.AssetDocumentService) *AssetDocumentHandler {
	return &AssetDocumentHandler{service: service}
}

// CreateDocument godoc
// @Summary Create a new asset document
// @Description Create a new document for an asset
// @Tags asset-documents
// @Accept json
// @Produce json
// @Param document body dto.AssetDocumentRequest true "Document details"
// @Success 201 {object} dto.AssetDocumentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/documents [post]
func (h *AssetDocumentHandler) CreateDocument(c *gin.Context) {
	var req dto.AssetDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	serviceReq := &service.CreateDocumentRequest{
		AssetID:      req.AssetID,
		DocumentType: req.DocumentType,
		Title:        req.Title,
		FilePath:     req.FilePath,
		Description:  req.Description,
		UploadDate:   req.UploadDate,
	}

	document, err := h.service.CreateDocument(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, document)
}

// GetDocument godoc
// @Summary Get a document by ID
// @Description Get document details by its ID
// @Tags asset-documents
// @Accept json
// @Produce json
// @Param id path string true "Document ID"
// @Success 200 {object} dto.AssetDocumentResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/documents/{id} [get]
func (h *AssetDocumentHandler) GetDocument(c *gin.Context) {
	id := c.Param("id")

	document, err := h.service.GetDocument(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, document)
}

// UpdateDocument godoc
// @Summary Update a document
// @Description Update an existing document with the provided details
// @Tags asset-documents
// @Accept json
// @Produce json
// @Param id path string true "Document ID"
// @Param document body dto.AssetDocumentRequest true "Document details"
// @Success 200 {object} dto.AssetDocumentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/documents/{id} [put]
func (h *AssetDocumentHandler) UpdateDocument(c *gin.Context) {
	id := c.Param("id")
	var req dto.AssetDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	serviceReq := &service.UpdateDocumentRequest{
		AssetID:      req.AssetID,
		DocumentType: req.DocumentType,
		Title:        req.Title,
		FilePath:     req.FilePath,
		Description:  req.Description,
		UploadDate:   req.UploadDate,
	}

	document, err := h.service.UpdateDocument(c.Request.Context(), id, serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, document)
}

// DeleteDocument godoc
// @Summary Delete a document
// @Description Delete an existing document by its ID
// @Tags asset-documents
// @Accept json
// @Produce json
// @Param id path string true "Document ID"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/documents/{id} [delete]
func (h *AssetDocumentHandler) DeleteDocument(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteDocument(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListDocuments godoc
// @Summary List all documents
// @Description Get a list of all documents with optional filters
// @Tags asset-documents
// @Accept json
// @Produce json
// @Param asset_id query string false "Asset ID"
// @Param type query string false "Document type"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.AssetDocumentListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/documents [get]
func (h *AssetDocumentHandler) ListDocuments(c *gin.Context) {
	assetID := c.Query("asset_id")
	documentType := c.Query("type")
	page := dto.GetIntQuery(c, "page", 1)
	pageSize := dto.GetIntQuery(c, "page_size", 10)

	documents, total, err := h.service.ListDocuments(c.Request.Context(), assetID, documentType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.AssetDocumentListResponse{
		Items: documents,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}