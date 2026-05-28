package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComp_APIHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ai/models", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestComp_APIHandler_Predict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	body := PredictRequestHTTP{
		ModelID: "solar_forecast_v1.0",
		Version: "1.0",
		Inputs: map[string]interface{}{
			"irradiance":  800.0,
			"temperature": 25.0,
		},
		Options: PredictOptionsHTTP{IncludeConfidence: true},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ai/predictions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp, "request_id")
}

func TestComp_APIHandler_Predict_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ai/predictions", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestComp_APIHandler_RealtimePredict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	body := map[string]interface{}{
		"model_id":   "solar_forecast_v1.0",
		"station_id": "st1",
		"inputs": map[string]interface{}{
			"irradiance":  800.0,
			"temperature": 25.0,
		},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ai/predictions/realtime", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestComp_APIHandler_RealtimePredict_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ai/predictions/realtime", bytes.NewBufferString("bad"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestComp_APIHandler_ListPredictions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ai/predictions?station_id=st1&start_time=2024-01-01&end_time=2024-01-31", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestComp_APIHandler_GetPrediction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ai/predictions/test-pred-123", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestComp_APIHandler_BatchPredict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	body := &BatchPredictRequest{
		ModelID: "solar_forecast_v1.0",
		Version: "1.0",
		Inputs: []map[string]interface{}{
			{"irradiance": 800.0, "temperature": 25.0},
		},
		Options: PredictOptions{},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ai/batch-predict", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestComp_APIHandler_BatchPredict_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ai/batch-predict", bytes.NewBufferString("bad"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestComp_APIHandler_GetBatchJobStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	batchReq := &BatchPredictRequest{
		ModelID: "solar_forecast_v1.0",
		Version: "1.0",
		Inputs: []map[string]interface{}{
			{"irradiance": 800.0},
		},
		Options: PredictOptions{},
	}
	batchResp, err := svc.BatchPredict(context.Background(), batchReq)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ai/batch-predict/"+batchResp.JobID, nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestComp_APIHandler_GetBatchJobStatus_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ai/batch-predict/nonexistent", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestComp_APIHandler_ListModels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ai/models", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp, "data")
	assert.Contains(t, resp, "total")
}

func TestComp_APIHandler_GetModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ai/models/solar_forecast_v1.0", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestComp_APIHandler_GetModel_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)

	r := gin.New()
	handler.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ai/models/nonexistent", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestComp_TwoLevelCache_LocalOnly(t *testing.T) {
	cache, err := NewTwoLevelCache(
		getDefaultBigCacheConfig(),
		nil,
	)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	resp := &PredictResponse{
		RequestID: "tlc-test-1",
		ModelID:   "solar_forecast_v1.0",
		Prediction: Prediction{
			Value: 100.0,
			Unit:  "kW",
		},
	}

	err = cache.Set(ctx, "test-key-1", resp, 5*time.Minute)
	require.NoError(t, err)

	got, err := cache.Get(ctx, "test-key-1")
	require.NoError(t, err)
	assert.Equal(t, "tlc-test-1", got.RequestID)
	assert.Equal(t, 100.0, got.Prediction.Value)
}

func TestComp_TwoLevelCache_Miss(t *testing.T) {
	cache, err := NewTwoLevelCache(
		getDefaultBigCacheConfig(),
		nil,
	)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()
	_, err = cache.Get(ctx, "nonexistent-key")
	assert.Error(t, err)
}

func TestComp_TwoLevelCache_Delete(t *testing.T) {
	cache, err := NewTwoLevelCache(
		getDefaultBigCacheConfig(),
		nil,
	)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	resp := &PredictResponse{RequestID: "del-test"}
	cache.Set(ctx, "del-key", resp, 5*time.Minute)

	err = cache.Delete(ctx, "del-key")
	require.NoError(t, err)

	_, err = cache.Get(ctx, "del-key")
	assert.Error(t, err)
}

func TestComp_TwoLevelCache_DeleteByPattern(t *testing.T) {
	cache, err := NewTwoLevelCache(
		getDefaultBigCacheConfig(),
		nil,
	)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	resp := &PredictResponse{RequestID: "pattern-test"}
	cache.Set(ctx, "ai:predict:1", resp, 5*time.Minute)

	err = cache.DeleteByPattern(ctx, "ai:predict:")
	require.NoError(t, err)
}

func TestComp_TwoLevelCache_GenerateCacheKey(t *testing.T) {
	cache, err := NewTwoLevelCache(
		getDefaultBigCacheConfig(),
		nil,
	)
	require.NoError(t, err)
	defer cache.Close()

	key1 := cache.GenerateCacheKey("model1", "v1", map[string]interface{}{"a": 1, "b": 2})
	key2 := cache.GenerateCacheKey("model1", "v1", map[string]interface{}{"b": 2, "a": 1})
	key3 := cache.GenerateCacheKey("model1", "v2", map[string]interface{}{"a": 1, "b": 2})

	assert.Equal(t, key1, key2)
	assert.NotEqual(t, key1, key3)
	assert.Contains(t, key1, "ai:predict:model1:v1:")
}

func getDefaultBigCacheConfig() *bigcache.Config {
	cfg := bigcache.DefaultConfig(10 * time.Minute)
	return &cfg
}
