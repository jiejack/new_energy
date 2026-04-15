package forecast

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateMetrics(t *testing.T) {
	actual := []float64{10, 20, 30, 40, 50}
	predicted := []float64{12, 18, 32, 38, 52}
	
	metrics := CalculateMetrics(actual, predicted)
	
	assert.InDelta(t, 2.0, metrics.MAE, 0.001)
	assert.InDelta(t, 0.0, metrics.RMSE, 2.0)
	assert.InDelta(t, 0.9, metrics.R2, 0.1)
}

func TestCalculateMetricsWithNaN(t *testing.T) {
	actual := []float64{10, 20, 30, 40, 50}
	predicted := []float64{12, 18, 32, float64(math.NaN()), 52}
	
	metrics := CalculateMetrics(actual, predicted)
	
	assert.InDelta(t, 2.0, metrics.MAE, 0.001)
}

func TestNaiveForecaster(t *testing.T) {
	forecaster := NewNaiveForecaster("test-naive")
	
	assert.Equal(t, "test-naive", forecaster.GetModelInfo().ModelID)
	assert.Equal(t, "untrained", forecaster.GetModelInfo().Status)
	
	baseTime := time.Now()
	trainingData := []*TimeSeriesData{
		{Timestamp: baseTime, Value: 10},
		{Timestamp: baseTime.Add(time.Hour), Value: 20},
		{Timestamp: baseTime.Add(2 * time.Hour), Value: 30},
	}
	
	err := forecaster.Train(context.Background(), trainingData)
	assert.NoError(t, err)
	assert.Equal(t, "trained", forecaster.GetModelInfo().Status)
	
	predictions, err := forecaster.Predict(context.Background(), 5)
	assert.NoError(t, err)
	assert.Len(t, predictions, 5)
	assert.Equal(t, 30.0, predictions[0].Value)
}

func TestNaiveForecasterNotTrained(t *testing.T) {
	forecaster := NewNaiveForecaster("test-naive-untrained")
	
	predictions, err := forecaster.Predict(context.Background(), 5)
	assert.Error(t, err)
	assert.Nil(t, predictions)
}

func TestSimpleAverageForecaster(t *testing.T) {
	forecaster := NewSimpleAverageForecaster("test-average", 3)
	
	assert.Equal(t, "test-average", forecaster.GetModelInfo().ModelID)
	assert.Equal(t, "untrained", forecaster.GetModelInfo().Status)
	
	baseTime := time.Now()
	trainingData := []*TimeSeriesData{
		{Timestamp: baseTime, Value: 10},
		{Timestamp: baseTime.Add(time.Hour), Value: 20},
		{Timestamp: baseTime.Add(2 * time.Hour), Value: 30},
		{Timestamp: baseTime.Add(3 * time.Hour), Value: 40},
	}
	
	err := forecaster.Train(context.Background(), trainingData)
	assert.NoError(t, err)
	assert.Equal(t, "trained", forecaster.GetModelInfo().Status)
	
	predictions, err := forecaster.Predict(context.Background(), 3)
	assert.NoError(t, err)
	assert.Len(t, predictions, 3)
	assert.InDelta(t, 30.0, predictions[0].Value, 0.001)
}

func TestSimpleAverageForecasterSmallWindow(t *testing.T) {
	forecaster := NewSimpleAverageForecaster("test-average-small", 10)
	
	baseTime := time.Now()
	trainingData := []*TimeSeriesData{
		{Timestamp: baseTime, Value: 10},
		{Timestamp: baseTime.Add(time.Hour), Value: 20},
	}
	
	err := forecaster.Train(context.Background(), trainingData)
	assert.NoError(t, err)
	
	predictions, err := forecaster.Predict(context.Background(), 2)
	assert.NoError(t, err)
	assert.Len(t, predictions, 2)
	assert.InDelta(t, 15.0, predictions[0].Value, 0.001)
}

func TestEvaluationMetricsEdgeCases(t *testing.T) {
	metrics := CalculateMetrics([]float64{}, []float64{})
	assert.Equal(t, 0.0, metrics.MAE)
	assert.Equal(t, 0.0, metrics.RMSE)
	assert.Equal(t, 0.0, metrics.MAPE)
	assert.Equal(t, 0.0, metrics.R2)
	
	metrics = CalculateMetrics([]float64{10}, []float64{})
	assert.Equal(t, 0.0, metrics.MAE)
}

func TestModelInfo(t *testing.T) {
	forecaster := NewNaiveForecaster("test-model-info")
	modelInfo := forecaster.GetModelInfo()
	
	assert.Equal(t, "test-model-info", modelInfo.ModelID)
	assert.Equal(t, "naive", modelInfo.ModelType)
	assert.Equal(t, "1.0.0", modelInfo.Version)
	assert.False(t, modelInfo.CreatedAt.IsZero())
	assert.Nil(t, modelInfo.TrainedAt)
	assert.Equal(t, "untrained", modelInfo.Status)
}

func TestPredictionStructure(t *testing.T) {
	forecaster := NewNaiveForecaster("test-prediction")
	
	baseTime := time.Now()
	trainingData := []*TimeSeriesData{
		{Timestamp: baseTime, Value: 100},
	}
	
	forecaster.Train(context.Background(), trainingData)
	
	predictions, _ := forecaster.Predict(context.Background(), 1)
	prediction := predictions[0]
	
	assert.False(t, prediction.Timestamp.IsZero())
	assert.InDelta(t, 100.0, prediction.Value, 0.001)
	assert.InDelta(t, 0.5, prediction.Confidence, 0.001)
}
