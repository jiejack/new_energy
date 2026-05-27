package inference

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCache struct {
	data map[string]*PredictResponse
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]*PredictResponse)}
}

func (m *mockCache) Get(ctx context.Context, key string) (*PredictResponse, error) {
	resp, ok := m.data[key]
	if !ok {
		return nil, assert.AnError
	}
	return resp, nil
}

func (m *mockCache) Set(ctx context.Context, key string, resp *PredictResponse, ttl time.Duration) error {
	m.data[key] = resp
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockCache) DeleteByPattern(ctx context.Context, pattern string) error {
	m.data = make(map[string]*PredictResponse)
	return nil
}

func TestInferenceService_Create(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	assert.NotNil(t, svc)
}

func TestInferenceService_Predict(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	req := &PredictRequest{
		ModelID: "solar_forecast_v1.0",
		Version: "1.0",
		Inputs: map[string]interface{}{
			"irradiance":  800.0,
			"temperature": 25.0,
		},
		Options: PredictOptions{},
	}

	resp, err := svc.Predict(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.RequestID)
	assert.Equal(t, "solar_forecast_v1.0", resp.ModelID)
	assert.False(t, resp.Metadata.Cached)
	assert.Greater(t, resp.Prediction.Value, 0.0)
	assert.Equal(t, "kW", resp.Prediction.Unit)
}

func TestInferenceService_Predict_WithCache(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	req := &PredictRequest{
		ModelID: "solar_forecast_v1.0",
		Version: "1.0",
		Inputs: map[string]interface{}{
			"irradiance":  800.0,
			"temperature": 25.0,
		},
		Options: PredictOptions{CacheTTLSeconds: 300},
	}

	resp1, err := svc.Predict(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, resp1.Metadata.Cached)

	resp2, err := svc.Predict(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, resp2.Metadata.Cached)
}

func TestInferenceService_Predict_WithExplanation(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	req := &PredictRequest{
		ModelID: "solar_forecast_v1.0",
		Version: "1.0",
		Inputs: map[string]interface{}{
			"irradiance":  800.0,
			"temperature": 25.0,
		},
		Options: PredictOptions{IncludeExplanation: true},
	}

	resp, err := svc.Predict(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp.Explanation)
	assert.NotNil(t, resp.Explanation.FeatureImportance)
}

func TestInferenceService_Predict_ModelNotFound(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	req := &PredictRequest{
		ModelID: "nonexistent_model",
		Version: "1.0",
		Inputs:  map[string]interface{}{},
		Options: PredictOptions{},
	}

	_, err := svc.Predict(context.Background(), req)
	assert.Error(t, err)
}

func TestInferenceService_Predict_SolarZeroIrradiance(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	req := &PredictRequest{
		ModelID: "solar_forecast_v1.0",
		Version: "1.0",
		Inputs: map[string]interface{}{
			"irradiance":  0.0,
			"temperature": 25.0,
		},
		Options: PredictOptions{},
	}

	resp, err := svc.Predict(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 0.0, resp.Prediction.Value)
}

func TestInferenceService_Predict_WindLowSpeed(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	req := &PredictRequest{
		ModelID: "wind_forecast_v1.0",
		Version: "1.0",
		Inputs: map[string]interface{}{
			"wind_speed": 2.0,
		},
		Options: PredictOptions{},
	}

	resp, err := svc.Predict(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 0.0, resp.Prediction.Value)
}

func TestInferenceService_Predict_WindHighSpeed(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	req := &PredictRequest{
		ModelID: "wind_forecast_v1.0",
		Version: "1.0",
		Inputs: map[string]interface{}{
			"wind_speed": 30.0,
		},
		Options: PredictOptions{},
	}

	resp, err := svc.Predict(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 0.0, resp.Prediction.Value)
}

func TestInferenceService_Predict_WindNormalSpeed(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	req := &PredictRequest{
		ModelID: "wind_forecast_v1.0",
		Version: "1.0",
		Inputs: map[string]interface{}{
			"wind_speed": 10.0,
		},
		Options: PredictOptions{},
	}

	resp, err := svc.Predict(context.Background(), req)
	require.NoError(t, err)
	assert.Greater(t, resp.Prediction.Value, 0.0)
}

func TestInferenceService_Predict_FaultDetector(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	req := &PredictRequest{
		ModelID: "fault_detector_v1.0",
		Version: "1.0",
		Inputs: map[string]interface{}{
			"temperature": 90.0,
			"vibration":   6.0,
			"current":     120.0,
		},
		Options: PredictOptions{},
	}

	resp, err := svc.Predict(context.Background(), req)
	require.NoError(t, err)
	assert.Greater(t, resp.Prediction.Value, 0.0)
	assert.Equal(t, "score", resp.Prediction.Unit)
}

func TestInferenceService_BatchPredict(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	req := &BatchPredictRequest{
		ModelID: "solar_forecast_v1.0",
		Version: "1.0",
		Inputs: []map[string]interface{}{
			{"irradiance": 800.0, "temperature": 25.0},
			{"irradiance": 600.0, "temperature": 20.0},
		},
		Options: PredictOptions{},
	}

	resp, err := svc.BatchPredict(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.JobID)
	assert.Equal(t, JobStatusQueued, resp.Status)
	assert.Equal(t, 4, resp.EstimatedTimeSeconds)

	time.Sleep(100 * time.Millisecond)

	jobStatus, err := svc.GetBatchJobStatus(context.Background(), resp.JobID)
	require.NoError(t, err)
	assert.Equal(t, JobStatusCompleted, jobStatus.Status)
	assert.Equal(t, 100, jobStatus.Progress)
}

func TestInferenceService_GetBatchJobStatus_NotFound(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	_, err := svc.GetBatchJobStatus(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job not found")
}

func TestInferenceService_ListModels(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	models, err := svc.ListModels(context.Background())
	require.NoError(t, err)
	assert.Len(t, models, 3)
}

func TestInferenceService_GetModel(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	model, err := svc.GetModel(context.Background(), "solar_forecast_v1.0")
	require.NoError(t, err)
	assert.Equal(t, "solar_forecast_v1.0", model.ModelID)
	assert.Equal(t, ModelTypeSolarForecast, model.Type)
}

func TestInferenceService_GetModel_NotFound(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	_, err := svc.GetModel(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestSimpleModelManager_LoadModel(t *testing.T) {
	mm := NewSimpleModelManager()

	err := mm.LoadModel(context.Background(), "solar_forecast_v1.0", "1.0")
	assert.NoError(t, err)
}

func TestSimpleModelManager_LoadModel_NotFound(t *testing.T) {
	mm := NewSimpleModelManager()

	err := mm.LoadModel(context.Background(), "nonexistent", "1.0")
	assert.Error(t, err)
}

func TestSimpleModelManager_UnloadModel(t *testing.T) {
	mm := NewSimpleModelManager()

	mm.LoadModel(context.Background(), "solar_forecast_v1.0", "1.0")
	err := mm.UnloadModel(context.Background(), "solar_forecast_v1.0")
	assert.NoError(t, err)
}

func TestSimpleModelManager_Predict_HealthScore(t *testing.T) {
	mm := NewSimpleModelManager()
	mm.models["health_v1.0"] = &ModelInfo{
		ModelID: "health_v1.0",
		Version: "1.0",
		Type:    ModelTypeHealthScore,
		Status:  "active",
	}

	pred, err := mm.Predict(context.Background(), "health_v1.0", map[string]interface{}{
		"temperature": 60.0,
		"vibration":   2.0,
		"efficiency":  0.95,
	})
	require.NoError(t, err)
	assert.Greater(t, pred.Value, 0.0)
	assert.Equal(t, "points", pred.Unit)
}

func TestSimpleModelManager_Predict_UnsupportedType(t *testing.T) {
	mm := NewSimpleModelManager()
	mm.models["unsupported_v1.0"] = &ModelInfo{
		ModelID: "unsupported_v1.0",
		Version: "1.0",
		Type:    ModelType("unsupported"),
		Status:  "active",
	}

	_, err := mm.Predict(context.Background(), "unsupported_v1.0", map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported model type")
}

func TestSimpleModelManager_Predict_HealthScoreZero(t *testing.T) {
	mm := NewSimpleModelManager()
	mm.models["health_v1.0"] = &ModelInfo{
		ModelID: "health_v1.0",
		Version: "1.0",
		Type:    ModelTypeHealthScore,
		Status:  "active",
	}

	pred, err := mm.Predict(context.Background(), "health_v1.0", map[string]interface{}{
		"temperature": 200.0,
		"vibration":   50.0,
		"efficiency":  0.1,
	})
	require.NoError(t, err)
	assert.Equal(t, 0.0, pred.Value)
}

func TestSimpleCache(t *testing.T) {
	cache := NewSimpleCache()

	resp := &PredictResponse{
		RequestID: "test-123",
		ModelID:   "solar_forecast_v1.0",
	}

	err := cache.Set(context.Background(), "key1", resp, 5*time.Minute)
	assert.NoError(t, err)

	result, err := cache.Get(context.Background(), "key1")
	assert.NoError(t, err)
	assert.Equal(t, "test-123", result.RequestID)
}

func TestSimpleCache_Miss(t *testing.T) {
	cache := NewSimpleCache()

	_, err := cache.Get(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestSimpleCache_Expired(t *testing.T) {
	cache := NewSimpleCache()

	resp := &PredictResponse{RequestID: "test-123"}
	err := cache.Set(context.Background(), "key1", resp, 1*time.Nanosecond)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	_, err = cache.Get(context.Background(), "key1")
	assert.Error(t, err)
}

func TestSimpleCache_Delete(t *testing.T) {
	cache := NewSimpleCache()

	resp := &PredictResponse{RequestID: "test-123"}
	cache.Set(context.Background(), "key1", resp, 5*time.Minute)

	err := cache.Delete(context.Background(), "key1")
	assert.NoError(t, err)

	_, err = cache.Get(context.Background(), "key1")
	assert.Error(t, err)
}

func TestSimpleCache_DeleteByPattern(t *testing.T) {
	cache := NewSimpleCache()

	resp1 := &PredictResponse{RequestID: "test-1"}
	resp2 := &PredictResponse{RequestID: "test-2"}
	cache.Set(context.Background(), "ai:predict:1", resp1, 5*time.Minute)
	cache.Set(context.Background(), "ai:predict:2", resp2, 5*time.Minute)
	cache.Set(context.Background(), "other:key", resp1, 5*time.Minute)

	err := cache.DeleteByPattern(context.Background(), "ai:predict:")
	assert.NoError(t, err)

	_, err = cache.Get(context.Background(), "ai:predict:1")
	assert.Error(t, err)

	_, err = cache.Get(context.Background(), "other:key")
	assert.NoError(t, err)
}

func TestSimpleCache_DeleteByPattern_Wildcard(t *testing.T) {
	cache := NewSimpleCache()

	resp := &PredictResponse{RequestID: "test-1"}
	cache.Set(context.Background(), "key1", resp, 5*time.Minute)
	cache.Set(context.Background(), "key2", resp, 5*time.Minute)

	err := cache.DeleteByPattern(context.Background(), "*")
	assert.NoError(t, err)

	_, err = cache.Get(context.Background(), "key1")
	assert.Error(t, err)
}

func TestMatchPattern(t *testing.T) {
	assert.True(t, matchPattern("ai:predict:1", "*"))
	assert.True(t, matchPattern("ai:predict:1", "ai:"))
	assert.False(t, matchPattern("bi:predict:1", "ai:"))
}

func TestHashInputs(t *testing.T) {
	inputs1 := map[string]interface{}{"a": 1, "b": 2}
	inputs2 := map[string]interface{}{"b": 2, "a": 1}

	hash1 := hashInputs(inputs1)
	hash2 := hashInputs(inputs2)

	assert.Equal(t, hash1, hash2)
	assert.NotEmpty(t, hash1)
}

func TestGenerateFeatureImportance(t *testing.T) {
	inputs := map[string]interface{}{
		"irradiance":  800.0,
		"temperature": 25.0,
	}

	importance := generateFeatureImportance(inputs)
	assert.Len(t, importance, 2)

	total := 0.0
	for _, v := range importance {
		total += v
	}
	assert.InDelta(t, 1.0, total, 0.001)
}

func TestNewAPIHandler(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)
	handler := NewAPIHandler(svc)
	assert.NotNil(t, handler)
}

func TestPredictRequest_Struct(t *testing.T) {
	req := &PredictRequest{
		ModelID: "test_model",
		Version: "1.0",
		Inputs:  map[string]interface{}{"key": "value"},
		Options: PredictOptions{IncludeConfidence: true, IncludeExplanation: true, CacheTTLSeconds: 300},
	}
	assert.Equal(t, "test_model", req.ModelID)
	assert.True(t, req.Options.IncludeConfidence)
	assert.True(t, req.Options.IncludeExplanation)
	assert.Equal(t, 300, req.Options.CacheTTLSeconds)
}

func TestBatchJobStatus_Struct(t *testing.T) {
	now := time.Now()
	status := &BatchJobStatus{
		JobID:       "job-123",
		Status:      JobStatusRunning,
		Progress:    50,
		StartedAt:   &now,
		ItemCount:   10,
		ErrorMessage: "",
	}
	assert.Equal(t, "job-123", status.JobID)
	assert.Equal(t, JobStatusRunning, status.Status)
	assert.Equal(t, 50, status.Progress)
}

func TestModelInfo_Struct(t *testing.T) {
	info := &ModelInfo{
		ModelID:     "test",
		Version:     "1.0",
		Type:        ModelTypeSolarForecast,
		Description: "Test model",
		Status:      "active",
		Metrics:     map[string]float64{"accuracy": 0.95},
	}
	assert.Equal(t, "test", info.ModelID)
	assert.Equal(t, ModelTypeSolarForecast, info.Type)
}

func TestPrediction_Struct(t *testing.T) {
	pred := &Prediction{
		Value:              100.0,
		Unit:               "kW",
		Confidence:         0.95,
		ConfidenceInterval: []float64{90.0, 110.0},
	}
	assert.Equal(t, 100.0, pred.Value)
	assert.Equal(t, "kW", pred.Unit)
	assert.Equal(t, 0.95, pred.Confidence)
}

func TestInferenceMetadata_Struct(t *testing.T) {
	meta := InferenceMetadata{
		InferenceTimeMs: 42,
		Cached:          false,
		Timestamp:       time.Now(),
	}
	assert.Equal(t, int64(42), meta.InferenceTimeMs)
	assert.False(t, meta.Cached)
}

func TestJobStatus_Constants(t *testing.T) {
	assert.Equal(t, JobStatus("queued"), JobStatusQueued)
	assert.Equal(t, JobStatus("running"), JobStatusRunning)
	assert.Equal(t, JobStatus("completed"), JobStatusCompleted)
	assert.Equal(t, JobStatus("failed"), JobStatusFailed)
	assert.Equal(t, JobStatus("cancelled"), JobStatusCancelled)
}

func TestModelType_Constants(t *testing.T) {
	assert.Equal(t, ModelType("solar_forecast"), ModelTypeSolarForecast)
	assert.Equal(t, ModelType("wind_forecast"), ModelTypeWindForecast)
	assert.Equal(t, ModelType("fault_detector"), ModelTypeFaultDetector)
	assert.Equal(t, ModelType("health_score"), ModelTypeHealthScore)
}

func TestInferenceType_Constants(t *testing.T) {
	assert.Equal(t, InferenceType("realtime"), InferenceTypeRealtime)
	assert.Equal(t, InferenceType("batch"), InferenceTypeBatch)
}

func TestInferenceService_BatchPredict_Cancelled(t *testing.T) {
	cache := newMockCache()
	mm := NewSimpleModelManager()
	svc := NewInferenceService(cache, mm)

	ctx, cancel := context.WithCancel(context.Background())

	req := &BatchPredictRequest{
		ModelID: "solar_forecast_v1.0",
		Version: "1.0",
		Inputs:  make([]map[string]interface{}, 100),
		Options: PredictOptions{},
	}
	for i := range req.Inputs {
		req.Inputs[i] = map[string]interface{}{"irradiance": 800.0}
	}

	resp, err := svc.BatchPredict(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.JobID)

	cancel()

	time.Sleep(200 * time.Millisecond)

	jobStatus, err := svc.GetBatchJobStatus(context.Background(), resp.JobID)
	if err == nil {
		assert.True(t, jobStatus.Status == JobStatusCancelled || jobStatus.Status == JobStatusCompleted)
	}
}
