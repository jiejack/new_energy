package forecast

import (
	"context"
	"fmt"
	"math"
	"time"
)

const (
	DefaultAirDensity        = 1.225
	DefaultRotorDiameter     = 100.0
	DefaultTurbineEfficiency = 0.45
	DefaultCutInSpeed        = 3.0
	DefaultRatedSpeed        = 12.0
	DefaultCutOutSpeed       = 25.0
)

type WindForecaster struct {
	trainingData      []*TimeSeriesData
	latitude          float64
	longitude         float64
	airDensity        float64
	rotorDiameter     float64
	turbineEfficiency float64
	cutInSpeed        float64
	ratedSpeed        float64
	cutOutSpeed       float64
	turbineCount      int
	modelInfo         *ModelInfo
	seasonalPatterns  map[int]float64
	dailyPatterns     map[int]float64
	windRose          map[int]map[float64]float64
}

func NewWindForecaster(modelID string, latitude, longitude float64) *WindForecaster {
	now := time.Now()
	
	wf := &WindForecaster{
		latitude:          latitude,
		longitude:         longitude,
		airDensity:        DefaultAirDensity,
		rotorDiameter:     DefaultRotorDiameter,
		turbineEfficiency: DefaultTurbineEfficiency,
		cutInSpeed:        DefaultCutInSpeed,
		ratedSpeed:        DefaultRatedSpeed,
		cutOutSpeed:       DefaultCutOutSpeed,
		turbineCount:      1,
		modelInfo: &ModelInfo{
			ModelID:   modelID,
			ModelType: "wind_specialized",
			Version:   "1.0.0",
			CreatedAt: now,
			Status:    "untrained",
			Parameters: map[string]interface{}{
				"latitude":           latitude,
				"longitude":          longitude,
				"air_density":        DefaultAirDensity,
				"rotor_diameter":     DefaultRotorDiameter,
				"turbine_efficiency": DefaultTurbineEfficiency,
				"cut_in_speed":       DefaultCutInSpeed,
				"rated_speed":        DefaultRatedSpeed,
				"cut_out_speed":      DefaultCutOutSpeed,
				"turbine_count":      1,
			},
		},
		seasonalPatterns: make(map[int]float64),
		dailyPatterns:    make(map[int]float64),
		windRose:         make(map[int]map[float64]float64),
	}
	
	wf.initializeDefaultPatterns()
	
	return wf
}

func (wf *WindForecaster) SetTurbineConfig(diameter, efficiency float64, count int) {
	wf.rotorDiameter = diameter
	wf.turbineEfficiency = efficiency
	wf.turbineCount = count
	wf.modelInfo.Parameters["rotor_diameter"] = diameter
	wf.modelInfo.Parameters["turbine_efficiency"] = efficiency
	wf.modelInfo.Parameters["turbine_count"] = count
}

func (wf *WindForecaster) SetSpeedThresholds(cutIn, rated, cutOut float64) {
	wf.cutInSpeed = cutIn
	wf.ratedSpeed = rated
	wf.cutOutSpeed = cutOut
	wf.modelInfo.Parameters["cut_in_speed"] = cutIn
	wf.modelInfo.Parameters["rated_speed"] = rated
	wf.modelInfo.Parameters["cut_out_speed"] = cutOut
}

func (wf *WindForecaster) SetAirDensity(density float64) {
	wf.airDensity = density
	wf.modelInfo.Parameters["air_density"] = density
}

func (wf *WindForecaster) initializeDefaultPatterns() {
	for month := 1; month <= 12; month++ {
		switch {
		case month >= 3 && month <= 5:
			wf.seasonalPatterns[month] = 0.9
		case month >= 6 && month <= 8:
			wf.seasonalPatterns[month] = 0.7
		case month >= 9 && month <= 11:
			wf.seasonalPatterns[month] = 0.85
		default:
			wf.seasonalPatterns[month] = 1.0
		}
	}
	
	for hour := 0; hour < 24; hour++ {
		switch {
		case hour >= 0 && hour < 6:
			wf.dailyPatterns[hour] = 0.8
		case hour >= 6 && hour < 12:
			wf.dailyPatterns[hour] = 0.9
		case hour >= 12 && hour < 18:
			wf.dailyPatterns[hour] = 1.0
		default:
			wf.dailyPatterns[hour] = 0.95
		}
	}
}

func (wf *WindForecaster) calculateTheoreticalPower(windSpeed float64) float64 {
	if windSpeed < wf.cutInSpeed || windSpeed > wf.cutOutSpeed {
		return 0.0
	}
	
	rotorArea := math.Pi * math.Pow(wf.rotorDiameter/2, 2)
	
	var power float64
	if windSpeed <= wf.ratedSpeed {
		power = 0.5 * wf.airDensity * rotorArea * 
			math.Pow(windSpeed, 3) * wf.turbineEfficiency / 1000.0
	} else {
		power = 0.5 * wf.airDensity * rotorArea * 
			math.Pow(wf.ratedSpeed, 3) * wf.turbineEfficiency / 1000.0
	}
	
	return power * float64(wf.turbineCount)
}

func (wf *WindForecaster) estimateWindSpeed(timestamp time.Time) float64 {
	hour := timestamp.Hour()
	
	seasonalFactor := wf.seasonalPatterns[int(timestamp.Month())]
	dailyFactor := wf.dailyPatterns[hour]
	
	baseWindSpeed := 8.0
	
	if len(wf.trainingData) > 0 {
		weekAgo := timestamp.Add(-168 * time.Hour)
		var matchingSpeeds []float64
		
		for _, dp := range wf.trainingData {
			if dp.Timestamp.Hour() == hour && 
			   dp.Timestamp.After(weekAgo) {
				if windSpeed, ok := dp.Features["wind_speed"]; ok {
					matchingSpeeds = append(matchingSpeeds, windSpeed)
				}
			}
		}
		
		if len(matchingSpeeds) > 0 {
			sum := 0.0
			for _, s := range matchingSpeeds {
				sum += s
			}
			baseWindSpeed = sum / float64(len(matchingSpeeds))
		}
	}
	
	return baseWindSpeed * seasonalFactor * dailyFactor
}

func (wf *WindForecaster) Train(ctx context.Context, data []*TimeSeriesData) error {
	if len(data) < 48 {
		return fmt.Errorf("insufficient training data: need at least 48 hours")
	}
	
	wf.trainingData = make([]*TimeSeriesData, len(data))
	copy(wf.trainingData, data)
	
	hourlySums := make(map[int]float64)
	hourlyCounts := make(map[int]int)
	monthlySums := make(map[int]float64)
	monthlyCounts := make(map[int]int)
	
	for _, dp := range data {
		hour := dp.Timestamp.Hour()
		month := int(dp.Timestamp.Month())
		windSpeed := dp.Value
		
		if windSpeed <= 0 {
			if ws, ok := dp.Features["wind_speed"]; ok {
				windSpeed = ws
			}
		}
		
		hourlySums[hour] += windSpeed
		hourlyCounts[hour]++
		monthlySums[month] += windSpeed
		monthlyCounts[month]++
	}
	
	maxHourlyValue := 0.0
	for hour := 0; hour < 24; hour++ {
		if hourlyCounts[hour] > 0 {
			avg := hourlySums[hour] / float64(hourlyCounts[hour])
			if avg > maxHourlyValue {
				maxHourlyValue = avg
			}
		}
	}
	
	if maxHourlyValue > 0 {
		for hour := 0; hour < 24; hour++ {
			if hourlyCounts[hour] > 0 {
				wf.dailyPatterns[hour] = hourlySums[hour] / float64(hourlyCounts[hour]) / maxHourlyValue
			}
		}
	}
	
	maxMonthlyValue := 0.0
	for month := 1; month <= 12; month++ {
		if monthlyCounts[month] > 0 {
			avg := monthlySums[month] / float64(monthlyCounts[month])
			if avg > maxMonthlyValue {
				maxMonthlyValue = avg
			}
		}
	}
	
	if maxMonthlyValue > 0 {
		for month := 1; month <= 12; month++ {
			if monthlyCounts[month] > 0 {
				wf.seasonalPatterns[month] = monthlySums[month] / float64(monthlyCounts[month]) / maxMonthlyValue
			}
		}
	}
	
	now := time.Now()
	wf.modelInfo.TrainedAt = &now
	wf.modelInfo.Status = "trained"
	
	return nil
}

func (wf *WindForecaster) Predict(ctx context.Context, horizon int) ([]*Prediction, error) {
	if wf.modelInfo.Status != "trained" || len(wf.trainingData) == 0 {
		return nil, ErrModelNotTrained
	}
	
	lastTime := wf.trainingData[len(wf.trainingData)-1].Timestamp
	predictions := make([]*Prediction, horizon)
	
	for i := 0; i < horizon; i++ {
		predictionTime := lastTime.Add(time.Duration(i+1) * time.Hour)
		
		estimatedWindSpeed := wf.estimateWindSpeed(predictionTime)
		predictedPower := wf.calculateTheoreticalPower(estimatedWindSpeed)
		confidence := wf.calculateConfidence(predictionTime)
		
		predictions[i] = &Prediction{
			Timestamp:          predictionTime,
			Value:              predictedPower,
			Confidence:         confidence,
			ConfidenceInterval: [2]float64{predictedPower * 0.8, predictedPower * 1.2},
		}
	}
	
	return predictions, nil
}

func (wf *WindForecaster) calculateConfidence(timestamp time.Time) float64 {
	if wf.modelInfo.TrainedAt == nil {
		return 0.5
	}
	
	daysSinceTraining := time.Since(*wf.modelInfo.TrainedAt).Hours() / 24
	timeDecay := math.Max(0.5, 1.0-daysSinceTraining/30.0)
	
	return 0.65 * timeDecay
}

func (wf *WindForecaster) GetModelInfo() *ModelInfo {
	return wf.modelInfo
}

func (wf *WindForecaster) Save(ctx context.Context, path string) error {
	return nil
}

func (wf *WindForecaster) Load(ctx context.Context, path string) error {
	return nil
}

type WindSpeedForecaster struct {
	baseForecaster *WindForecaster
	modelInfo      *ModelInfo
}

func NewWindSpeedForecaster(modelID string, latitude, longitude float64) *WindSpeedForecaster {
	return &WindSpeedForecaster{
		baseForecaster: NewWindForecaster(modelID, latitude, longitude),
		modelInfo: &ModelInfo{
			ModelID:   modelID,
			ModelType: "wind_speed",
			Version:   "1.0.0",
			CreatedAt: time.Now(),
			Status:    "untrained",
			Parameters: map[string]interface{}{
				"latitude":  latitude,
				"longitude": longitude,
			},
		},
	}
}

func (wsf *WindSpeedForecaster) Train(ctx context.Context, data []*TimeSeriesData) error {
	err := wsf.baseForecaster.Train(ctx, data)
	if err != nil {
		return err
	}
	
	wsf.modelInfo.Status = "trained"
	wsf.modelInfo.TrainedAt = wsf.baseForecaster.modelInfo.TrainedAt
	
	return nil
}

func (wsf *WindSpeedForecaster) Predict(ctx context.Context, horizon int) ([]*Prediction, error) {
	predictions, err := wsf.baseForecaster.Predict(ctx, horizon)
	if err != nil {
		return nil, err
	}
	
	lastTime := wsf.baseForecaster.trainingData[len(wsf.baseForecaster.trainingData)-1].Timestamp
	
	for i, pred := range predictions {
		predictionTime := lastTime.Add(time.Duration(i+1) * time.Hour)
		pred.Value = wsf.baseForecaster.estimateWindSpeed(predictionTime)
	}
	
	return predictions, nil
}

func (wsf *WindSpeedForecaster) GetModelInfo() *ModelInfo {
	return wsf.modelInfo
}

func (wsf *WindSpeedForecaster) Save(ctx context.Context, path string) error {
	return wsf.baseForecaster.Save(ctx, path)
}

func (wsf *WindSpeedForecaster) Load(ctx context.Context, path string) error {
	return wsf.baseForecaster.Load(ctx, path)
}
