package inference

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewInferenceService(t *testing.T) {
	cache := NewSimpleCache()
	modelManager := NewSimpleModelManager()
	service := NewInferenceService(cache, modelManager)
	
	assert.NotNil(t, service)
}

func TestPredict_SolarForecast(t *testing.T) {
	cache := NewSimpleCache()
	modelManager := NewSimpleModelManager()
	service := NewInferenceService(cache, modelManager)
	
	ctx := context.Background()
	req := &PredictRequest{
		ModelID: "solar_forecast_v1.0",
		Version: "latest",
		Inputs: map[string]interface{}{
			"irradiance": 800.5,
			"temperature": 25.3,
		},
		Options: PredictOptions{
			IncludeConfidence: true,
			IncludeExplanation: false,
			CacheTTLSeconds: 3600,
		},
	}
	
	resp, err := service.Predict(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "solar_forecast_v1.0", resp.ModelID)
	assert.Equal(t, "kW", resp.Prediction.Unit)
	assert.False(t, resp.Metadata.Cached)
}

func TestPredict_WindForecast(t *testing.T) {
	cache := NewSimpleCache()
	modelManager := NewSimpleModelManager()
	service := NewInferenceService(cache, modelManager)
	
	ctx := context.Background()
	req := &PredictRequest{
		ModelID: "wind_forecast_v1.0",
		Version: "latest",
		Inputs: map[string]interface{}{
			"wind_speed": 10.5,
		},
		Options: PredictOptions{
			IncludeConfidence: true,
			CacheTTLSeconds: 3600,
		},
	}
	
	resp, err := service.Predict(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "wind_forecast_v1.0", resp.ModelID)
	assert.Equal(t, "kW", resp.Prediction.Unit)
}

func TestPredict_FaultDetector(t *testing.T) {
	cache := NewSimpleCache()
	modelManager := NewSimpleModelManager()
	service := NewInferenceService(cache, modelManager)
	
	ctx := context.Background()
	req := &PredictRequest{
		ModelID: "fault_detector_v1.0",
		Version: "latest",
		Inputs: map[string]interface{}{
			"temperature": 85.0,
			"vibration": 6.0,
			"current": 110.0,
		},
		Options: PredictOptions{
			IncludeConfidence: true,
			CacheTTLSeconds: 3600,
		},
	}
	
	resp, err := service.Predict(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "fault_detector_v1.0", resp.ModelID)
	assert.Equal(t, "score", resp.Prediction.Unit)
	assert.Greater(t, resp.Prediction.Value, 0.0)
}

func TestPredict_WithCache(t *testing.T) {
	cache := NewSimpleCache()
	modelManager := NewSimpleModelManager()
	service := NewInferenceService(cache, modelManager)
	
	ctx := context.Background()
	req := &PredictRequest{
		ModelID: "solar_forecast_v1.0",
		Version: "latest",
		Inputs: map[string]interface{}{
			"irradiance": 800.5,
			"temperature": 25.3,
		},
		Options: PredictOptions{
			IncludeConfidence: true,
			CacheTTLSeconds: 3600,
		},
	}
	
	resp1, err := service.Predict(ctx, req)
	assert.NoError(t, err)
	assert.False(t, resp1.Metadata.Cached)
	
	time.Sleep(10 * time.Millisecond)
	
	resp2, err := service.Predict(ctx, req)
	assert.NoError(t, err)
	assert.True(t, resp2.Metadata.Cached)
	assert.Equal(t, resp1.Prediction.Value, resp2.Prediction.Value)
}

func TestBatchPredict(t *testing.T) {
	cache := NewSimpleCache()
	modelManager := NewSimpleModelManager()
	service := NewInferenceService(cache, modelManager)
	
	ctx := context.Background()
	req := &BatchPredictRequest{
		ModelID: "solar_forecast_v1.0",
		Version: "latest",
		BatchID: "test_batch_001",
		Inputs: []map[string]interface{}{
			{
				"irradiance": 800.5,
				"temperature": 25.3,
			},
			{
				"irradiance": 750.0,
				"temperature": 26.0,
			},
		},
		Options: PredictOptions{
			IncludeConfidence: true,
			CacheTTLSeconds: 3600,
		},
	}
	
	resp, err := service.BatchPredict(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, JobStatusQueued, resp.Status)
	assert.NotEmpty(t, resp.JobID)
}

func TestGetBatchJobStatus(t *testing.T) {
	cache := NewSimpleCache()
	modelManager := NewSimpleModelManager()
	service := NewInferenceService(cache, modelManager)
	
	ctx := context.Background()
	batchReq := &BatchPredictRequest{
		ModelID: "solar_forecast_v1.0",
		Version: "latest",
		BatchID: "test_batch_001",
		Inputs: []map[string]interface{}{
			{
				"irradiance": 800.5,
				"temperature": 25.3,
			},
		},
		Options: PredictOptions{
			IncludeConfidence: true,
		},
	}
	
	batchResp, err := service.BatchPredict(ctx, batchReq)
	assert.NoError(t, err)
	
	time.Sleep(100 * time.Millisecond)
	
	jobStatus, err := service.GetBatchJobStatus(ctx, batchResp.JobID)
	assert.NoError(t, err)
	assert.NotNil(t, jobStatus)
	assert.Equal(t, batchResp.JobID, jobStatus.JobID)
}

func TestListModels(t *testing.T) {
	cache := NewSimpleCache()
	modelManager := NewSimpleModelManager()
	service := NewInferenceService(cache, modelManager)
	
	ctx := context.Background()
	models, err := service.ListModels(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, models)
	assert.GreaterOrEqual(t, len(models), 3)
}

func TestGetModel(t *testing.T) {
	cache := NewSimpleCache()
	modelManager := NewSimpleModelManager()
	service := NewInferenceService(cache, modelManager)
	
	ctx := context.Background()
	model, err := service.GetModel(ctx, "solar_forecast_v1.0")
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, "solar_forecast_v1.0", model.ModelID)
	assert.Equal(t, ModelTypeSolarForecast, model.Type)
}
