package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/application/service"
)

type ModelHandler struct {
	modelSvc service.ModelService
}

func NewModelHandler(modelSvc service.ModelService) *ModelHandler {
	return &ModelHandler{modelSvc: modelSvc}
}

func (h *ModelHandler) RegisterModel(c *gin.Context) {
	var req struct {
		ModelName    string  `json:"model_name" binding:"required"`
		Version      string  `json:"version" binding:"required"`
		ArtifactPath string  `json:"artifact_path" binding:"required"`
		Accuracy     float64 `json:"accuracy"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	model, err := h.modelSvc.RegisterModel(c.Request.Context(), req.ModelName, req.Version, req.ArtifactPath, req.Accuracy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.Response{Code: 0, Message: "success", Data: model})
}

func (h *ModelHandler) GetModel(c *gin.Context) {
	id := c.Param("id")
	model, err := h.modelSvc.GetModel(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "model not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: model})
}

func (h *ModelHandler) ListModels(c *gin.Context) {
	modelName := c.Query("model_name")
	models, err := h.modelSvc.ListModels(c.Request.Context(), modelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: models})
}

func (h *ModelHandler) GetProductionModel(c *gin.Context) {
	modelName := c.Param("model_name")
	model, err := h.modelSvc.GetProductionModel(c.Request.Context(), modelName)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "no production model found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: model})
}

func (h *ModelHandler) PromoteToProduction(c *gin.Context) {
	id := c.Param("id")
	err := h.modelSvc.PromoteToProduction(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
}

func (h *ModelHandler) RetireModel(c *gin.Context) {
	id := c.Param("id")
	err := h.modelSvc.RetireModel(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
}
