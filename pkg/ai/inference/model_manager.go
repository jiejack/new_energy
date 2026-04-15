package inference

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type SimpleModelManager struct {
	models      map[string]*ModelInfo
	modelLock   sync.RWMutex
	loadedModels map[string]interface{}
}

func NewSimpleModelManager() *SimpleModelManager {
	mm := &SimpleModelManager{
		models:       make(map[string]*ModelInfo),
		loadedModels: make(map[string]interface{}),
	}
	mm.initializeDefaultModels()
	return mm
}

func (mm *SimpleModelManager) initializeDefaultModels() {
	now := time.Now()
	mm.models["solar_forecast_v1.0"] = &ModelInfo{
		ModelID:     "solar_forecast_v1.0",
		Version:     "1.0.0",
		Type:        ModelTypeSolarForecast,
		Description: "光伏发电量预测模型 v1.0",
		CreatedAt:   now.AddDate(0, -1, 0),
		UpdatedAt:   now,
		Status:      "active",
		Metrics: map[string]float64{
			"mae":   150.5,
			"rmse":  200.3,
			"r2":    0.92,
		},
	}

	mm.models["wind_forecast_v1.0"] = &ModelInfo{
		ModelID:     "wind_forecast_v1.0",
		Version:     "1.0.0",
		Type:        ModelTypeWindForecast,
		Description: "风力发电量预测模型 v1.0",
		CreatedAt:   now.AddDate(0, -1, 0),
		UpdatedAt:   now,
		Status:      "active",
		Metrics: map[string]float64{
			"mae":   120.3,
			"rmse":  160.8,
			"r2":    0.89,
		},
	}

	mm.models["fault_detector_v1.0"] = &ModelInfo{
		ModelID:     "fault_detector_v1.0",
		Version:     "1.0.0",
		Type:        ModelTypeFaultDetector,
		Description: "设备故障检测模型 v1.0",
		CreatedAt:   now.AddDate(0, -1, 0),
		UpdatedAt:   now,
		Status:      "active",
		Metrics: map[string]float64{
			"precision": 0.95,
			"recall":    0.92,
			"f1":        0.93,
		},
	}
}

func (mm *SimpleModelManager) LoadModel(ctx context.Context, modelID, version string) error {
	mm.modelLock.Lock()
	defer mm.modelLock.Unlock()

	if _, ok := mm.models[modelID]; !ok {
		return fmt.Errorf("model not found: %s", modelID)
	}

	mm.loadedModels[modelID] = struct{}{}
	return nil
}

func (mm *SimpleModelManager) UnloadModel(ctx context.Context, modelID string) error {
	mm.modelLock.Lock()
	defer mm.modelLock.Unlock()

	delete(mm.loadedModels, modelID)
	return nil
}

func (mm *SimpleModelManager) GetModelInfo(ctx context.Context, modelID string) (*ModelInfo, error) {
	mm.modelLock.RLock()
	defer mm.modelLock.RUnlock()

	model, ok := mm.models[modelID]
	if !ok {
		return nil, fmt.Errorf("model not found: %s", modelID)
	}
	return model, nil
}

func (mm *SimpleModelManager) ListModels(ctx context.Context) ([]*ModelInfo, error) {
	mm.modelLock.RLock()
	defer mm.modelLock.RUnlock()

	models := make([]*ModelInfo, 0, len(mm.models))
	for _, model := range mm.models {
		models = append(models, model)
	}
	return models, nil
}

func (mm *SimpleModelManager) Predict(ctx context.Context, modelID string, inputs map[string]interface{}) (*Prediction, error) {
	mm.modelLock.RLock()
	model, ok := mm.models[modelID]
	mm.modelLock.RUnlock()

	if !ok {
		return nil, fmt.Errorf("model not found: %s", modelID)
	}

	switch model.Type {
	case ModelTypeSolarForecast:
		return mm.predictSolar(inputs)
	case ModelTypeWindForecast:
		return mm.predictWind(inputs)
	case ModelTypeFaultDetector:
		return mm.detectFault(inputs)
	case ModelTypeHealthScore:
		return mm.calculateHealthScore(inputs)
	default:
		return nil, fmt.Errorf("unsupported model type: %s", model.Type)
	}
}

func (mm *SimpleModelManager) predictSolar(inputs map[string]interface{}) (*Prediction, error) {
	irradiance, _ := inputs["irradiance"].(float64)
	temperature, _ := inputs["temperature"].(float64)

	if irradiance == 0 {
		return &Prediction{
			Value:      0,
			Unit:       "kW",
			Confidence: 0.99,
			ConfidenceInterval: []float64{0, 0},
		}, nil
	}

	basePower := irradiance * 5.0
	tempCorrection := 1.0 - (temperature-25.0)*0.004
	prediction := basePower * tempCorrection

	confidence := 0.85 + rand.Float64()*0.1
	margin := prediction * (1 - confidence) * 0.5

	return &Prediction{
		Value:               prediction,
		Unit:                "kW",
		Confidence:          confidence,
		ConfidenceInterval: []float64{prediction - margin, prediction + margin},
	}, nil
}

func (mm *SimpleModelManager) predictWind(inputs map[string]interface{}) (*Prediction, error) {
	windSpeed, _ := inputs["wind_speed"].(float64)

	if windSpeed < 3 {
		return &Prediction{
			Value:      0,
			Unit:       "kW",
			Confidence: 0.99,
			ConfidenceInterval: []float64{0, 0},
		}, nil
	}

	if windSpeed > 25 {
		return &Prediction{
			Value:      0,
			Unit:       "kW",
			Confidence: 0.99,
			ConfidenceInterval: []float64{0, 0},
		}, nil
	}

	var prediction float64
	if windSpeed < 12 {
		prediction = windSpeed * windSpeed * windSpeed * 0.3
	} else {
		prediction = 5000.0
	}

	confidence := 0.82 + rand.Float64()*0.12
	margin := prediction * (1 - confidence) * 0.5

	return &Prediction{
		Value:               prediction,
		Unit:                "kW",
		Confidence:          confidence,
		ConfidenceInterval: []float64{prediction - margin, prediction + margin},
	}, nil
}

func (mm *SimpleModelManager) detectFault(inputs map[string]interface{}) (*Prediction, error) {
	temperature, _ := inputs["temperature"].(float64)
	vibration, _ := inputs["vibration"].(float64)
	current, _ := inputs["current"].(float64)

	faultScore := 0.0

	if temperature > 80 {
		faultScore += 0.3
	}
	if vibration > 5.0 {
		faultScore += 0.4
	}
	if current > 100 {
		faultScore += 0.3
	}

	confidence := 0.75 + faultScore*0.2

	return &Prediction{
		Value:               faultScore,
		Unit:                "score",
		Confidence:          confidence,
		ConfidenceInterval: []float64{faultScore - 0.1, faultScore + 0.1},
	}, nil
}

func (mm *SimpleModelManager) calculateHealthScore(inputs map[string]interface{}) (*Prediction, error) {
	temperature, _ := inputs["temperature"].(float64)
	vibration, _ := inputs["vibration"].(float64)
	efficiency, _ := inputs["efficiency"].(float64)

	healthScore := 100.0

	if temperature > 70 {
		healthScore -= (temperature - 70) * 0.5
	}
	if vibration > 3.0 {
		healthScore -= (vibration - 3.0) * 5.0
	}
	if efficiency < 0.9 {
		healthScore -= (0.9 - efficiency) * 100.0
	}

	if healthScore < 0 {
		healthScore = 0
	}

	confidence := 0.80 + rand.Float64()*0.15

	return &Prediction{
		Value:               healthScore,
		Unit:                "points",
		Confidence:          confidence,
		ConfidenceInterval: []float64{healthScore - 5, healthScore + 5},
	}, nil
}
