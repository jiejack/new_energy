package forecast

import (
	"context"
	"fmt"
	"math"
	"time"
)

func CalculateMetrics(actual, predicted []float64) *EvaluationMetrics {
	n := len(actual)
	if n == 0 || n != len(predicted) {
		return &EvaluationMetrics{}
	}

	mae, mape, rmse := 0.0, 0.0, 0.0
	actualMean := 0.0
	validCount := 0

	for i := 0; i < n; i++ {
		if math.IsNaN(actual[i]) || math.IsInf(actual[i], 0) || math.IsNaN(predicted[i]) || math.IsInf(predicted[i], 0) {
			continue
		}
		
		diff := math.Abs(predicted[i] - actual[i])
		mae += diff
		if actual[i] != 0 {
			mape += diff / math.Abs(actual[i])
		}
		rmse += diff * diff
		actualMean += actual[i]
		validCount++
	}

	if validCount == 0 {
		return &EvaluationMetrics{}
	}

	mae /= float64(validCount)
	mape /= float64(validCount) * 100
	rmse = math.Sqrt(rmse / float64(validCount))
	actualMean /= float64(validCount)

	ssRes, ssTot := 0.0, 0.0
	for i := 0; i < n; i++ {
		if math.IsNaN(actual[i]) || math.IsInf(actual[i], 0) || math.IsNaN(predicted[i]) || math.IsInf(predicted[i], 0) {
			continue
		}
		ssRes += (actual[i] - predicted[i]) * (actual[i] - predicted[i])
		ssTot += (actual[i] - actualMean) * (actual[i] - actualMean)
	}

	r2 := 1.0
	if ssTot != 0 {
		r2 = 1 - ssRes/ssTot
	}

	return &EvaluationMetrics{
		MAE:  mae,
		MAPE: mape,
		RMSE: rmse,
		R2:   r2,
	}
}

type NaiveForecaster struct {
	trainingData []*TimeSeriesData
	modelInfo    *ModelInfo
}

func NewNaiveForecaster(modelID string) *NaiveForecaster {
	return &NaiveForecaster{
		modelInfo: &ModelInfo{
			ModelID:   modelID,
			ModelType: "naive",
			Version:   "1.0.0",
			CreatedAt: time.Now(),
			Status:    "untrained",
			Parameters: map[string]interface{}{
				"method": "last_value",
			},
		},
	}
}

func (n *NaiveForecaster) Train(ctx context.Context, data []*TimeSeriesData) error {
	n.trainingData = make([]*TimeSeriesData, len(data))
	copy(n.trainingData, data)
	
	now := time.Now()
	n.modelInfo.TrainedAt = &now
	n.modelInfo.Status = "trained"
	
	return nil
}

func (n *NaiveForecaster) Predict(ctx context.Context, horizon int) ([]*Prediction, error) {
	if len(n.trainingData) == 0 {
		return nil, ErrModelNotTrained
	}
	
	lastValue := n.trainingData[len(n.trainingData)-1].Value
	lastTime := n.trainingData[len(n.trainingData)-1].Timestamp
	
	predictions := make([]*Prediction, horizon)
	
	for i := 0; i < horizon; i++ {
		predictionTime := lastTime.Add(time.Duration(i+1) * time.Hour)
		
		predictions[i] = &Prediction{
			Timestamp:          predictionTime,
			Value:              lastValue,
			Confidence:         0.5,
		}
	}
	
	return predictions, nil
}

func (n *NaiveForecaster) GetModelInfo() *ModelInfo {
	return n.modelInfo
}

func (n *NaiveForecaster) Save(ctx context.Context, path string) error {
	return nil
}

func (n *NaiveForecaster) Load(ctx context.Context, path string) error {
	return nil
}

type SimpleAverageForecaster struct {
	trainingData []*TimeSeriesData
	windowSize   int
	modelInfo    *ModelInfo
}

func NewSimpleAverageForecaster(modelID string, windowSize int) *SimpleAverageForecaster {
	if windowSize <= 0 {
		windowSize = 24
	}
	
	return &SimpleAverageForecaster{
		windowSize: windowSize,
		modelInfo: &ModelInfo{
			ModelID:   modelID,
			ModelType: "simple_average",
			Version:   "1.0.0",
			CreatedAt: time.Now(),
			Status:    "untrained",
			Parameters: map[string]interface{}{
				"window_size": windowSize,
			},
		},
	}
}

func (s *SimpleAverageForecaster) Train(ctx context.Context, data []*TimeSeriesData) error {
	s.trainingData = make([]*TimeSeriesData, len(data))
	copy(s.trainingData, data)
	
	now := time.Now()
	s.modelInfo.TrainedAt = &now
	s.modelInfo.Status = "trained"
	
	return nil
}

func (s *SimpleAverageForecaster) Predict(ctx context.Context, horizon int) ([]*Prediction, error) {
	if len(s.trainingData) == 0 {
		return nil, ErrModelNotTrained
	}
	
	n := len(s.trainingData)
	windowSize := s.windowSize
	if n < windowSize {
		windowSize = n
	}
	
	sum := 0.0
	for i := n - windowSize; i < n; i++ {
		sum += s.trainingData[i].Value
	}
	averageValue := sum / float64(windowSize)
	
	lastTime := s.trainingData[n-1].Timestamp
	
	predictions := make([]*Prediction, horizon)
	
	for i := 0; i < horizon; i++ {
		predictionTime := lastTime.Add(time.Duration(i+1) * time.Hour)
		
		predictions[i] = &Prediction{
			Timestamp:          predictionTime,
			Value:              averageValue,
			Confidence:         0.6,
		}
	}
	
	return predictions, nil
}

func (s *SimpleAverageForecaster) GetModelInfo() *ModelInfo {
	return s.modelInfo
}

func (s *SimpleAverageForecaster) Save(ctx context.Context, path string) error {
	return nil
}

func (s *SimpleAverageForecaster) Load(ctx context.Context, path string) error {
	return nil
}

var (
	ErrModelNotTrained = fmt.Errorf("model not trained")
)
