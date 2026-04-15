package forecast

import (
	"context"
	"fmt"
	"math"
	"time"
)

const (
	SolarPanelEfficiency = 0.22
	DefaultPanelCapacity = 300.0
)

type SolarForecaster struct {
	trainingData       []*TimeSeriesData
	latitude           float64
	longitude          float64
	panelCapacity      float64
	panelCount         int
	modelInfo          *ModelInfo
	seasonalFactors    map[int]float64
	dailyPattern       map[int]float64
	weatherAdjustments map[string]float64
}

func NewSolarForecaster(modelID string, latitude, longitude float64) *SolarForecaster {
	now := time.Now()
	
	sf := &SolarForecaster{
		latitude:      latitude,
		longitude:     longitude,
		panelCapacity: DefaultPanelCapacity,
		panelCount:    1,
		modelInfo: &ModelInfo{
			ModelID:   modelID,
			ModelType: "solar_specialized",
			Version:   "1.0.0",
			CreatedAt: now,
			Status:    "untrained",
			Parameters: map[string]interface{}{
				"latitude":       latitude,
				"longitude":      longitude,
				"panel_capacity": DefaultPanelCapacity,
				"panel_count":    1,
			},
		},
		seasonalFactors:    make(map[int]float64),
		dailyPattern:       make(map[int]float64),
		weatherAdjustments: make(map[string]float64),
	}
	
	sf.initializeDefaultPatterns()
	
	return sf
}

func (sf *SolarForecaster) SetPanelConfig(capacity float64, count int) {
	sf.panelCapacity = capacity
	sf.panelCount = count
	sf.modelInfo.Parameters["panel_capacity"] = capacity
	sf.modelInfo.Parameters["panel_count"] = count
}

func (sf *SolarForecaster) SetWeatherAdjustment(weatherType string, factor float64) {
	sf.weatherAdjustments[weatherType] = factor
}

func (sf *SolarForecaster) initializeDefaultPatterns() {
	for month := 1; month <= 12; month++ {
		switch {
		case month >= 3 && month <= 5:
			sf.seasonalFactors[month] = 0.9
		case month >= 6 && month <= 8:
			sf.seasonalFactors[month] = 1.0
		case month >= 9 && month <= 11:
			sf.seasonalFactors[month] = 0.8
		default:
			sf.seasonalFactors[month] = 0.6
		}
	}
	
	for hour := 0; hour < 24; hour++ {
		switch {
		case hour >= 6 && hour < 9:
			sf.dailyPattern[hour] = 0.3 + float64(hour-6)*0.2
		case hour >= 9 && hour < 15:
			sf.dailyPattern[hour] = 1.0
		case hour >= 15 && hour < 19:
			sf.dailyPattern[hour] = 1.0 - float64(hour-15)*0.25
		default:
			sf.dailyPattern[hour] = 0.0
		}
	}
	
	sf.weatherAdjustments["clear"] = 1.0
	sf.weatherAdjustments["partly_cloudy"] = 0.7
	sf.weatherAdjustments["cloudy"] = 0.4
	sf.weatherAdjustments["rainy"] = 0.1
	sf.weatherAdjustments["snowy"] = 0.3
}

func (sf *SolarForecaster) calculateSolarRadiation(timestamp time.Time) float64 {
	dayOfYear := timestamp.YearDay()
	declination := 23.45 * math.Sin(2*math.Pi*(284+float64(dayOfYear))/365) * math.Pi / 180
	latitudeRad := sf.latitude * math.Pi / 180
	
	hour := timestamp.Hour()
	solarNoon := 12.0
	hourAngle := (float64(hour) - solarNoon) * 15 * math.Pi / 180
	
	cosZenith := math.Sin(latitudeRad)*math.Sin(declination) + 
		math.Cos(latitudeRad)*math.Cos(declination)*math.Cos(hourAngle)
	
	if cosZenith < 0 {
		return 0.0
	}
	
	return 1000.0 * cosZenith
}

func (sf *SolarForecaster) calculateTheoreticalPower(timestamp time.Time) float64 {
	solarRadiation := sf.calculateSolarRadiation(timestamp)
	if solarRadiation <= 0 {
		return 0.0
	}
	
	month := int(timestamp.Month())
	hour := timestamp.Hour()
	
	seasonalFactor := sf.seasonalFactors[month]
	dailyFactor := sf.dailyPattern[hour]
	
	power := solarRadiation * sf.panelCapacity * float64(sf.panelCount) * 
		SolarPanelEfficiency * seasonalFactor * dailyFactor / 1000.0
	
	return math.Max(0, power)
}

func (sf *SolarForecaster) Train(ctx context.Context, data []*TimeSeriesData) error {
	if len(data) < 24 {
		return fmt.Errorf("insufficient training data: need at least 24 hours")
	}
	
	sf.trainingData = make([]*TimeSeriesData, len(data))
	copy(sf.trainingData, data)
	
	hourlySums := make(map[int]float64)
	hourlyCounts := make(map[int]int)
	monthlySums := make(map[int]float64)
	monthlyCounts := make(map[int]int)
	
	for _, dp := range data {
		hour := dp.Timestamp.Hour()
		month := int(dp.Timestamp.Month())
		
		hourlySums[hour] += dp.Value
		hourlyCounts[hour]++
		monthlySums[month] += dp.Value
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
				sf.dailyPattern[hour] = hourlySums[hour] / float64(hourlyCounts[hour]) / maxHourlyValue
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
				sf.seasonalFactors[month] = monthlySums[month] / float64(monthlyCounts[month]) / maxMonthlyValue
			}
		}
	}
	
	now := time.Now()
	sf.modelInfo.TrainedAt = &now
	sf.modelInfo.Status = "trained"
	
	return nil
}

func (sf *SolarForecaster) Predict(ctx context.Context, horizon int) ([]*Prediction, error) {
	if sf.modelInfo.Status != "trained" || len(sf.trainingData) == 0 {
		return nil, ErrModelNotTrained
	}
	
	lastTime := sf.trainingData[len(sf.trainingData)-1].Timestamp
	predictions := make([]*Prediction, horizon)
	
	for i := 0; i < horizon; i++ {
		predictionTime := lastTime.Add(time.Duration(i+1) * time.Hour)
		
		theoreticalPower := sf.calculateTheoreticalPower(predictionTime)
		weatherFactor := sf.getWeatherFactor(predictionTime)
		
		predictedValue := theoreticalPower * weatherFactor
		
		confidence := sf.calculateConfidence(predictionTime)
		
		predictions[i] = &Prediction{
			Timestamp:          predictionTime,
			Value:              predictedValue,
			Confidence:         confidence,
			ConfidenceInterval: [2]float64{predictedValue * 0.85, predictedValue * 1.15},
		}
	}
	
	return predictions, nil
}

func (sf *SolarForecaster) getWeatherFactor(timestamp time.Time) float64 {
	if len(sf.trainingData) == 0 {
		return 1.0
	}
	
	weekAgo := timestamp.Add(-168 * time.Hour)
	var matchingData []*TimeSeriesData
	
	for _, dp := range sf.trainingData {
		if dp.Timestamp.Hour() == timestamp.Hour() && 
		   dp.Timestamp.After(weekAgo) {
			matchingData = append(matchingData, dp)
		}
	}
	
	if len(matchingData) == 0 {
		return 1.0
	}
	
	sum := 0.0
	for _, dp := range matchingData {
		theoretical := sf.calculateTheoreticalPower(dp.Timestamp)
		if theoretical > 0 {
			sum += dp.Value / theoretical
		}
	}
	
	return sum / float64(len(matchingData))
}

func (sf *SolarForecaster) calculateConfidence(timestamp time.Time) float64 {
	hour := timestamp.Hour()
	
	if sf.dailyPattern[hour] < 0.1 {
		return 0.95
	}
	
	if sf.modelInfo.TrainedAt == nil {
		return 0.5
	}
	
	daysSinceTraining := time.Since(*sf.modelInfo.TrainedAt).Hours() / 24
	timeDecay := math.Max(0.5, 1.0-daysSinceTraining/30.0)
	
	return 0.7 * timeDecay
}

func (sf *SolarForecaster) GetModelInfo() *ModelInfo {
	return sf.modelInfo
}

func (sf *SolarForecaster) Save(ctx context.Context, path string) error {
	return nil
}

func (sf *SolarForecaster) Load(ctx context.Context, path string) error {
	return nil
}

type SolarIrradianceForecaster struct {
	baseForecaster *SolarForecaster
	modelInfo      *ModelInfo
}

func NewSolarIrradianceForecaster(modelID string, latitude, longitude float64) *SolarIrradianceForecaster {
	return &SolarIrradianceForecaster{
		baseForecaster: NewSolarForecaster(modelID, latitude, longitude),
		modelInfo: &ModelInfo{
			ModelID:   modelID,
			ModelType: "solar_irradiance",
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

func (sif *SolarIrradianceForecaster) Train(ctx context.Context, data []*TimeSeriesData) error {
	err := sif.baseForecaster.Train(ctx, data)
	if err != nil {
		return err
	}
	
	sif.modelInfo.Status = "trained"
	sif.modelInfo.TrainedAt = sif.baseForecaster.modelInfo.TrainedAt
	
	return nil
}

func (sif *SolarIrradianceForecaster) Predict(ctx context.Context, horizon int) ([]*Prediction, error) {
	predictions, err := sif.baseForecaster.Predict(ctx, horizon)
	if err != nil {
		return nil, err
	}
	
	for _, pred := range predictions {
		pred.Value = sif.baseForecaster.calculateSolarRadiation(pred.Timestamp)
	}
	
	return predictions, nil
}

func (sif *SolarIrradianceForecaster) GetModelInfo() *ModelInfo {
	return sif.modelInfo
}

func (sif *SolarIrradianceForecaster) Save(ctx context.Context, path string) error {
	return sif.baseForecaster.Save(ctx, path)
}

func (sif *SolarIrradianceForecaster) Load(ctx context.Context, path string) error {
	return sif.baseForecaster.Load(ctx, path)
}
