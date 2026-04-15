package datacollector

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type APIHandler struct {
	importer *CSVImporter
	validator *SimpleValidator
	cleaner *AdvancedCleaner
}

func NewAPIHandler() *APIHandler {
	return &APIHandler{
		importer: NewCSVImporter(),
		validator: NewSimpleValidator(ValidationConfig{
			CheckMissing:  true,
			CheckOutliers: true,
			CheckRange:    true,
		}),
		cleaner: NewAdvancedCleaner(),
	}
}

func (h *APIHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1/data")
	{
		api.POST("/import", h.ImportData)
		api.GET("/import/:job_id", h.GetImportJobStatus)
		api.POST("/collect", h.CollectData)
		api.POST("/quality/check", h.CheckDataQuality)
		api.POST("/clean", h.CleanData)
		api.GET("/query", h.QueryData)
	}
}

type ImportDataRequest struct {
	FilePath   string            `json:"file_path" binding:"required"`
	Format     DataFormat        `json:"format" binding:"required"`
	Mapping    map[string]string `json:"mapping" binding:"required"`
	Validation ValidationConfig  `json:"validation"`
	Options    ImportOptions     `json:"options"`
}

type ImportDataResponse struct {
	JobID   string    `json:"job_id"`
	Status  JobStatus `json:"status"`
	Message string    `json:"message"`
}

func (h *APIHandler) ImportData(c *gin.Context) {
	var req ImportDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	batchReq := &BatchImportRequest{
		FilePath:   req.FilePath,
		Format:     req.Format,
		Mapping:    req.Mapping,
		Validation: req.Validation,
		Options:    req.Options,
	}

	job, err := h.importer.Import(c.Request.Context(), batchReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, ImportDataResponse{
		JobID:   job.ID,
		Status:  job.Status,
		Message: "Import job submitted successfully",
	})
}

func (h *APIHandler) GetImportJobStatus(c *gin.Context) {
	jobID := c.Param("job_id")

	job, err := h.importer.GetJobStatus(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, job)
}

type CollectDataRequest struct {
	Points []*DataPoint `json:"points" binding:"required"`
}

type CollectDataResponse struct {
	ReceivedCount int `json:"received_count"`
	SuccessCount  int `json:"success_count"`
	FailedCount   int `json:"failed_count"`
}

func (h *APIHandler) CollectData(c *gin.Context) {
	var req CollectDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	validated, err := h.validator.Validate(c.Request.Context(), req.Points)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	successCount := 0
	failedCount := 0

	for _, p := range validated {
		if p.Quality == QualityGood || p.Quality == QualitySuspect {
			successCount++
		} else {
			failedCount++
		}
	}

	c.JSON(http.StatusOK, CollectDataResponse{
		ReceivedCount: len(req.Points),
		SuccessCount:  successCount,
		FailedCount:   failedCount,
	})
}

type CheckDataQualityRequest struct {
	Points    []*DataPoint     `json:"points" binding:"required"`
	Config    ValidationConfig `json:"config"`
	StationID string           `json:"station_id"`
	DeviceID  string           `json:"device_id"`
	Metric    string           `json:"metric"`
}

type CheckDataQualityResponse struct {
	TotalPoints   int     `json:"total_points"`
	GoodPoints    int     `json:"good_points"`
	SuspectPoints int     `json:"suspect_points"`
	BadPoints     int     `json:"bad_points"`
	QualityScore  float64 `json:"quality_score"`
}

func (h *APIHandler) CheckDataQuality(c *gin.Context) {
	var req CheckDataQualityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	validator := NewSimpleValidator(req.Config)
	validated, err := validator.Validate(c.Request.Context(), req.Points)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	totalPoints := len(validated)
	goodPoints := 0
	suspectPoints := 0
	badPoints := 0

	for _, p := range validated {
		switch p.Quality {
		case QualityGood:
			goodPoints++
		case QualitySuspect:
			suspectPoints++
		case QualityBad, QualityMissing:
			badPoints++
		}
	}

	qualityScore := 0.0
	if totalPoints > 0 {
		qualityScore = float64(goodPoints + suspectPoints/2) / float64(totalPoints)
	}

	c.JSON(http.StatusOK, CheckDataQualityResponse{
		TotalPoints:   totalPoints,
		GoodPoints:    goodPoints,
		SuspectPoints: suspectPoints,
		BadPoints:     badPoints,
		QualityScore:  qualityScore,
	})
}

type CleanDataRequest struct {
	Points  []*DataPoint `json:"points" binding:"required"`
	Options CleanOptions `json:"options"`
}

type CleanDataResponse struct {
	OriginalCount int           `json:"original_count"`
	CleanedCount  int           `json:"cleaned_count"`
	CleanedPoints []*DataPoint  `json:"cleaned_points"`
}

func (h *APIHandler) CleanData(c *gin.Context) {
	var req CleanDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	cleaned, err := h.cleaner.CleanPipeline(c.Request.Context(), req.Points, req.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CleanDataResponse{
		OriginalCount: len(req.Points),
		CleanedCount:  len(cleaned),
		CleanedPoints: cleaned,
	})
}

type QueryDataRequest struct {
	StationID string    `json:"station_id" form:"station_id"`
	DeviceID  string    `json:"device_id" form:"device_id"`
	Metric    string    `json:"metric" form:"metric"`
	StartTime time.Time `json:"start_time" form:"start_time"`
	EndTime   time.Time `json:"end_time" form:"end_time"`
	Limit     int       `json:"limit" form:"limit"`
}

func (h *APIHandler) QueryData(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Query data endpoint not implemented yet",
	})
}
