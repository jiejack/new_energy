package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAPIHandler() *APIHandler {
	cache := NewSimpleCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	return NewAPIHandler(svc)
}

func TestAPIHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupAPIHandler()
	h.RegisterRoutes(r)
	routes := r.Routes()
	assert.GreaterOrEqual(t, len(routes), 7)
}

func TestAPIHandler_Predict_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createInferenceJSONRequest(map[string]interface{}{
		"model_id": "solar_forecast_v1.0",
		"version":  "1.0",
		"inputs": map[string]interface{}{
			"irradiance":  800.0,
			"temperature": 25.0,
		},
		"options": map[string]interface{}{
			"include_confidence":  true,
			"include_explanation": true,
			"cache_ttl_seconds":   300,
		},
	})

	h.Predict(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIHandler_Predict_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createInferenceJSONRequest("invalid")

	h.Predict(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIHandler_Predict_ModelError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createInferenceJSONRequest(map[string]interface{}{
		"model_id": "nonexistent_model",
		"version":  "1.0",
		"inputs":   map[string]interface{}{},
	})

	h.Predict(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAPIHandler_Predict_DefaultVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createInferenceJSONRequest(map[string]interface{}{
		"model_id": "solar_forecast_v1.0",
		"inputs": map[string]interface{}{
			"irradiance":  800.0,
			"temperature": 25.0,
		},
	})

	h.Predict(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIHandler_RealtimePredict_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createInferenceJSONRequest(map[string]interface{}{
		"model_id":   "solar_forecast_v1.0",
		"station_id": "station-1",
		"inputs": map[string]interface{}{
			"irradiance":  800.0,
			"temperature": 25.0,
		},
	})

	h.RealtimePredict(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIHandler_RealtimePredict_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createInferenceJSONRequest("invalid")

	h.RealtimePredict(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIHandler_RealtimePredict_ModelError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createInferenceJSONRequest(map[string]interface{}{
		"model_id":   "nonexistent",
		"station_id": "station-1",
		"inputs":     map[string]interface{}{},
	})

	h.RealtimePredict(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAPIHandler_ListPredictions_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/?station_id=s1&start_time=2024-01-01&end_time=2024-12-31&page=1&page_size=10", nil)
	c.Request = req

	h.ListPredictions(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIHandler_ListPredictions_MissingParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/?station_id=s1", nil)
	c.Request = req

	h.ListPredictions(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIHandler_GetPrediction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "pred-123"}}
	c.Request, _ = http.NewRequest("GET", "/", nil)

	h.GetPrediction(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIHandler_BatchPredict_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createInferenceJSONRequest(map[string]interface{}{
		"model_id": "solar_forecast_v1.0",
		"version":  "1.0",
		"inputs": []map[string]interface{}{
			{"irradiance": 800.0, "temperature": 25.0},
			{"irradiance": 600.0, "temperature": 20.0},
		},
	})

	h.BatchPredict(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIHandler_BatchPredict_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createInferenceJSONRequest("invalid")

	h.BatchPredict(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIHandler_BatchPredict_DefaultVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = createInferenceJSONRequest(map[string]interface{}{
		"model_id": "solar_forecast_v1.0",
		"inputs": []map[string]interface{}{
			{"irradiance": 800.0},
		},
	})

	h.BatchPredict(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIHandler_GetBatchJobStatus_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "job_id", Value: "nonexistent"}}
	c.Request, _ = http.NewRequest("GET", "/", nil)

	h.GetBatchJobStatus(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPIHandler_GetBatchJobStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := NewSimpleCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	h := NewAPIHandler(svc)

	batchReq := &BatchPredictRequest{
		ModelID: "solar_forecast_v1.0",
		Version: "1.0",
		Inputs: []map[string]interface{}{
			{"irradiance": 800.0},
		},
	}
	batchResp, err := svc.BatchPredict(context.Background(), batchReq)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "job_id", Value: batchResp.JobID}}
	c.Request, _ = http.NewRequest("GET", "/", nil)

	h.GetBatchJobStatus(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIHandler_ListModels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)

	h.ListModels(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIHandler_ListModels_FilterByType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/?type=solar_forecast", nil)
	c.Request = req

	h.ListModels(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp ListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.GreaterOrEqual(t, resp.Total, 1)
}

func TestAPIHandler_GetModel_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "model_id", Value: "solar_forecast_v1.0"}}
	c.Request, _ = http.NewRequest("GET", "/", nil)

	h.GetModel(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIHandler_GetModel_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupAPIHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "model_id", Value: "nonexistent"}}
	c.Request, _ = http.NewRequest("GET", "/", nil)

	h.GetModel(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPredictRequestHTTP_Struct(t *testing.T) {
	req := PredictRequestHTTP{
		ModelID: "solar_forecast_v1.0",
		Version: "1.0",
		Inputs:  map[string]interface{}{"irradiance": 800.0},
		Options: PredictOptionsHTTP{IncludeConfidence: true, CacheTTLSeconds: 300},
	}
	assert.Equal(t, "solar_forecast_v1.0", req.ModelID)
	assert.True(t, req.Options.IncludeConfidence)
}

func TestPredictResponseHTTP_Struct(t *testing.T) {
	resp := PredictResponseHTTP{
		RequestID: "req-1",
		ModelID:   "solar_forecast_v1.0",
		Version:   "1.0",
		Prediction: Prediction{Value: 4000.0, Unit: "kW"},
		Metadata:  InferenceMetadata{InferenceTimeMs: 50, Cached: false},
	}
	assert.Equal(t, "req-1", resp.RequestID)
}

func TestBatchPredictRequestHTTP_Struct(t *testing.T) {
	req := BatchPredictRequestHTTP{
		ModelID:     "solar_forecast_v1.0",
		Version:     "1.0",
		BatchID:     "batch-1",
		Inputs:      []map[string]interface{}{{"irradiance": 800.0}},
		CallbackURL: "http://callback.example.com",
	}
	assert.Equal(t, "batch-1", req.BatchID)
	assert.Equal(t, "http://callback.example.com", req.CallbackURL)
}

func TestBatchPredictResponseHTTP_Struct(t *testing.T) {
	resp := BatchPredictResponseHTTP{
		JobID:               "job-1",
		Status:              "queued",
		EstimatedTimeSeconds: 10,
		QueuePosition:        1,
	}
	assert.Equal(t, 10, resp.EstimatedTimeSeconds)
}

func TestBatchJobStatusHTTP_Struct(t *testing.T) {
	now := time.Now()
	resp := BatchJobStatusHTTP{
		JobID:       "job-1",
		Status:      "completed",
		Progress:    100,
		ResultURL:   "/api/results/job-1",
		StartedAt:   &now,
		CompletedAt: &now,
		ItemCount:   5,
	}
	assert.Equal(t, 100, resp.Progress)
	assert.Equal(t, "/api/results/job-1", resp.ResultURL)
}

func TestModelInfoHTTP_Struct(t *testing.T) {
	info := ModelInfoHTTP{
		ModelID:     "solar_forecast_v1.0",
		Version:     "1.0",
		Type:        "solar_forecast",
		Name:        "Solar Forecast",
		Description: "Solar power forecast model",
		Status:      "active",
		IsDefault:   true,
		Metrics:     map[string]float64{"r2": 0.92},
	}
	assert.True(t, info.IsDefault)
}

func TestListResponse_Struct(t *testing.T) {
	resp := ListResponse{
		Total:    10,
		Page:     1,
		PageSize: 5,
		Data:     []string{"item1", "item2"},
	}
	assert.Equal(t, 10, resp.Total)
	assert.Equal(t, 5, resp.PageSize)
}

func TestExplanation_Struct(t *testing.T) {
	exp := &Explanation{
		FeatureImportance: map[string]float64{"irradiance": 0.6, "temperature": 0.4},
		SHAPValues:        map[string]float64{"irradiance": 0.5},
		LIMEExplanation:   "irradiance is the most important feature",
	}
	assert.Equal(t, 0.6, exp.FeatureImportance["irradiance"])
	assert.Equal(t, "irradiance is the most important feature", exp.LIMEExplanation)
}

func TestBatchPredictRequest_Struct(t *testing.T) {
	req := &BatchPredictRequest{
		ModelID:     "solar_forecast_v1.0",
		Version:     "1.0",
		BatchID:     "batch-1",
		Inputs:      []map[string]interface{}{{"irradiance": 800.0}},
		CallbackURL: "http://callback.example.com",
	}
	assert.Equal(t, "http://callback.example.com", req.CallbackURL)
}

func TestBatchPredictResponse_Struct(t *testing.T) {
	resp := &BatchPredictResponse{
		JobID:               "job-1",
		Status:              JobStatusQueued,
		EstimatedTimeSeconds: 10,
		QueuePosition:        1,
	}
	assert.Equal(t, 1, resp.QueuePosition)
}

func TestInferenceServiceInterface_Compliance(t *testing.T) {
	var _ InferenceService = NewInferenceService(NewSimpleCache(), NewSimpleModelManager())
}

func TestCacheStrategyInterface_Compliance(t *testing.T) {
	var _ CacheStrategy = NewSimpleCache()
}

func TestModelManagerInterface_Compliance(t *testing.T) {
	var _ ModelManager = NewSimpleModelManager()
}

func createInferenceJSONRequest(data interface{}) *http.Request {
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
