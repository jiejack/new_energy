package inference

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type APIHandler struct {
	service InferenceService
}

func NewAPIHandler(service InferenceService) *APIHandler {
	return &APIHandler{
		service: service,
	}
}

func (h *APIHandler) RegisterRoutes(router *gin.Engine) {
	aiGroup := router.Group("/api/v1/ai")
	{
		predictions := aiGroup.Group("/predictions")
		{
			predictions.POST("", h.Predict)
			predictions.POST("/realtime", h.RealtimePredict)
			predictions.GET("", h.ListPredictions)
			predictions.GET("/:id", h.GetPrediction)
		}

		batch := aiGroup.Group("/batch-predict")
		{
			batch.POST("", h.BatchPredict)
			batch.GET("/:job_id", h.GetBatchJobStatus)
		}

		models := aiGroup.Group("/models")
		{
			models.GET("", h.ListModels)
			models.GET("/:model_id", h.GetModel)
		}
	}
}

type PredictRequestHTTP struct {
	ModelID string                 `json:"model_id" binding:"required"`
	Version string                 `json:"version"`
	Inputs  map[string]interface{} `json:"inputs" binding:"required"`
	Options PredictOptionsHTTP     `json:"options"`
}

type PredictOptionsHTTP struct {
	IncludeConfidence bool `json:"include_confidence"`
	IncludeExplanation bool `json:"include_explanation"`
	CacheTTLSeconds   int  `json:"cache_ttl_seconds"`
}

type PredictResponseHTTP struct {
	RequestID   string          `json:"request_id"`
	ModelID     string          `json:"model_id"`
	Version     string          `json:"version"`
	Prediction  Prediction      `json:"prediction"`
	Explanation *Explanation    `json:"explanation,omitempty"`
	Metadata    InferenceMetadata `json:"metadata"`
}

type BatchPredictRequestHTTP struct {
	ModelID     string                   `json:"model_id" binding:"required"`
	Version     string                   `json:"version"`
	BatchID     string                   `json:"batch_id"`
	Inputs      []map[string]interface{} `json:"inputs" binding:"required"`
	CallbackURL string                   `json:"callback_url"`
	Options     PredictOptionsHTTP       `json:"options"`
}

type BatchPredictResponseHTTP struct {
	JobID               string `json:"job_id"`
	Status              string `json:"status"`
	EstimatedTimeSeconds int    `json:"estimated_time_seconds"`
	QueuePosition        int    `json:"queue_position"`
}

type BatchJobStatusHTTP struct {
	JobID        string     `json:"job_id"`
	Status       string     `json:"status"`
	Progress     int        `json:"progress"`
	ResultURL    string     `json:"result_url,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ItemCount    int        `json:"item_count"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

type ModelInfoHTTP struct {
	ModelID      string    `json:"model_id"`
	Version      string    `json:"version"`
	Type         string    `json:"type"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	IsDefault    bool      `json:"is_default"`
	Metrics      map[string]float64 `json:"metrics"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ListResponse struct {
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Data     interface{} `json:"data"`
}

func (h *APIHandler) Predict(c *gin.Context) {
	var req PredictRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	predReq := &PredictRequest{
		ModelID: req.ModelID,
		Version: req.Version,
		Inputs:  req.Inputs,
		Options: PredictOptions{
			IncludeConfidence: req.Options.IncludeConfidence,
			IncludeExplanation: req.Options.IncludeExplanation,
			CacheTTLSeconds:   req.Options.CacheTTLSeconds,
		},
	}

	if predReq.Version == "" {
		predReq.Version = "latest"
	}
	if predReq.Options.CacheTTLSeconds == 0 {
		predReq.Options.CacheTTLSeconds = 3600
	}

	resp, err := h.service.Predict(ctx, predReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) RealtimePredict(c *gin.Context) {
	var req struct {
		ModelID   string                 `json:"model_id" binding:"required"`
		StationID string                 `json:"station_id" binding:"required"`
		Inputs    map[string]interface{} `json:"inputs" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	predReq := &PredictRequest{
		ModelID: req.ModelID,
		Version: "latest",
		Inputs:  req.Inputs,
		Options: PredictOptions{
			IncludeConfidence: true,
			CacheTTLSeconds:   300,
		},
	}

	resp, err := h.service.Predict(ctx, predReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prediction_id":       resp.RequestID,
		"prediction_value":    resp.Prediction.Value,
		"prediction_unit":     resp.Prediction.Unit,
		"confidence":          resp.Prediction.Confidence,
		"confidence_interval": resp.Prediction.ConfidenceInterval,
		"inference_time_ms":   resp.Metadata.InferenceTimeMs,
	})
}

func (h *APIHandler) ListPredictions(c *gin.Context) {
	stationID := c.Query("station_id")
	modelID := c.Query("model_id")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	if stationID == "" || startTime == "" || endTime == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "station_id, start_time, and end_time are required"})
		return
	}

	mockPredictions := []map[string]interface{}{
		{
			"prediction_id":   "pred_" + uuid.New().String(),
			"model_id":        modelID,
			"station_id":      stationID,
			"target_time":     startTime,
			"prediction_value": 4500.5,
			"prediction_unit":  "kW",
			"confidence":      0.92,
			"actual_value":    4480.3,
			"error":           20.2,
			"created_at":      time.Now().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, ListResponse{
		Total:    1,
		Page:     page,
		PageSize: pageSize,
		Data:     mockPredictions,
	})
}

func (h *APIHandler) GetPrediction(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"prediction_id": id,
		"message":       "Get prediction endpoint - to be implemented",
	})
}

func (h *APIHandler) BatchPredict(c *gin.Context) {
	var req BatchPredictRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	batchReq := &BatchPredictRequest{
		ModelID:     req.ModelID,
		Version:     req.Version,
		BatchID:     req.BatchID,
		Inputs:      req.Inputs,
		CallbackURL: req.CallbackURL,
		Options: PredictOptions{
			IncludeConfidence: req.Options.IncludeConfidence,
			CacheTTLSeconds:   req.Options.CacheTTLSeconds,
		},
	}

	if batchReq.Version == "" {
		batchReq.Version = "latest"
	}

	resp, err := h.service.BatchPredict(ctx, batchReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, BatchPredictResponseHTTP{
		JobID:               resp.JobID,
		Status:              string(resp.Status),
		EstimatedTimeSeconds: resp.EstimatedTimeSeconds,
		QueuePosition:        resp.QueuePosition,
	})
}

func (h *APIHandler) GetBatchJobStatus(c *gin.Context) {
	jobID := c.Param("job_id")

	ctx := context.Background()
	status, err := h.service.GetBatchJobStatus(ctx, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, BatchJobStatusHTTP{
		JobID:        status.JobID,
		Status:       string(status.Status),
		Progress:     status.Progress,
		ResultURL:    status.ResultURL,
		StartedAt:    status.StartedAt,
		CompletedAt:  status.CompletedAt,
		ItemCount:    status.ItemCount,
		ErrorMessage: status.ErrorMessage,
	})
}

func (h *APIHandler) ListModels(c *gin.Context) {
	modelType := c.Query("type")

	ctx := context.Background()
	models, err := h.service.ListModels(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filteredModels := make([]*ModelInfoHTTP, 0)
	for _, model := range models {
		if modelType == "" || string(model.Type) == modelType {
			filteredModels = append(filteredModels, &ModelInfoHTTP{
				ModelID:     model.ModelID,
				Version:     model.Version,
				Type:        string(model.Type),
				Name:        model.Description,
				Description: model.Description,
				Status:      model.Status,
				IsDefault:   false,
				Metrics:     model.Metrics,
				CreatedAt:   model.CreatedAt,
				UpdatedAt:   model.UpdatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, ListResponse{
		Total:    len(filteredModels),
		Page:     1,
		PageSize: len(filteredModels),
		Data:     filteredModels,
	})
}

func (h *APIHandler) GetModel(c *gin.Context) {
	modelID := c.Param("model_id")

	ctx := context.Background()
	model, err := h.service.GetModel(ctx, modelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ModelInfoHTTP{
		ModelID:     model.ModelID,
		Version:     model.Version,
		Type:        string(model.Type),
		Name:        model.Description,
		Description: model.Description,
		Status:      model.Status,
		IsDefault:   false,
		Metrics:     model.Metrics,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	})
}
