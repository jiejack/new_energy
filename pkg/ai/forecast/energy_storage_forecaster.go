package forecast

import (
	"context"
	"fmt"
	"math"
	"time"
)

const (
	DefaultBatteryCapacity  = 1000.0
	DefaultChargeEfficiency = 0.95
	DefaultDischargeEfficiency = 0.95
	DefaultMinSOC = 10.0
	DefaultMaxSOC = 90.0
)

type EnergyStorageForecaster struct {
	trainingData          []*TimeSeriesData
	batteryCapacity       float64
	chargeEfficiency      float64
	dischargeEfficiency   float64
	minSOC                float64
	maxSOC                float64
	initialSOC            float64
	currentSOC            float64
	modelInfo             *ModelInfo
	dailyChargePattern    map[int]float64
	dailyDischargePattern map[int]float64
	seasonalFactors       map[int]float64
	degradationRate       float64
	cycleCount            int
}

func NewEnergyStorageForecaster(modelID string, capacity float64) *EnergyStorageForecaster {
	now := time.Now()
	
	if capacity <= 0 {
		capacity = DefaultBatteryCapacity
	}
	
	esf := &EnergyStorageForecaster{
		batteryCapacity:       capacity,
		chargeEfficiency:      DefaultChargeEfficiency,
		dischargeEfficiency:   DefaultDischargeEfficiency,
		minSOC:                DefaultMinSOC,
		maxSOC:                DefaultMaxSOC,
		initialSOC:            50.0,
		currentSOC:            50.0,
		degradationRate:       0.0001,
		cycleCount:            0,
		modelInfo: &ModelInfo{
			ModelID:   modelID,
			ModelType: "energy_storage",
			Version:   "1.0.0",
			CreatedAt: now,
			Status:    "untrained",
			Parameters: map[string]interface{}{
				"battery_capacity":       capacity,
				"charge_efficiency":      DefaultChargeEfficiency,
				"discharge_efficiency":   DefaultDischargeEfficiency,
				"min_soc":                DefaultMinSOC,
				"max_soc":                DefaultMaxSOC,
				"initial_soc":            50.0,
				"degradation_rate":       0.0001,
			},
		},
		dailyChargePattern:    make(map[int]float64),
		dailyDischargePattern: make(map[int]float64),
		seasonalFactors:       make(map[int]float64),
	}
	
	esf.initializeDefaultPatterns()
	
	return esf
}

func (esf *EnergyStorageForecaster) SetEfficiency(charge, discharge float64) {
	esf.chargeEfficiency = charge
	esf.dischargeEfficiency = discharge
	esf.modelInfo.Parameters["charge_efficiency"] = charge
	esf.modelInfo.Parameters["discharge_efficiency"] = discharge
}

func (esf *EnergyStorageForecaster) SetSOCLimits(min, max float64) {
	esf.minSOC = min
	esf.maxSOC = max
	esf.modelInfo.Parameters["min_soc"] = min
	esf.modelInfo.Parameters["max_soc"] = max
}

func (esf *EnergyStorageForecaster) SetInitialSOC(soc float64) {
	esf.initialSOC = math.Max(0, math.Min(100, soc))
	esf.currentSOC = esf.initialSOC
	esf.modelInfo.Parameters["initial_soc"] = esf.initialSOC
}

func (esf *EnergyStorageForecaster) SetDegradationRate(rate float64) {
	esf.degradationRate = rate
	esf.modelInfo.Parameters["degradation_rate"] = rate
}

func (esf *EnergyStorageForecaster) initializeDefaultPatterns() {
	for hour := 0; hour < 24; hour++ {
		switch {
		case hour >= 0 && hour < 6:
			esf.dailyChargePattern[hour] = 0.3
			esf.dailyDischargePattern[hour] = 0.8
		case hour >= 6 && hour < 12:
			esf.dailyChargePattern[hour] = 1.0
			esf.dailyDischargePattern[hour] = 0.4
		case hour >= 12 && hour < 18:
			esf.dailyChargePattern[hour] = 0.8
			esf.dailyDischargePattern[hour] = 0.6
		default:
			esf.dailyChargePattern[hour] = 0.5
			esf.dailyDischargePattern[hour] = 1.0
		}
	}
	
	for month := 1; month <= 12; month++ {
		switch {
		case month >= 6 && month <= 8:
			esf.seasonalFactors[month] = 1.2
		case month >= 3 && month <= 5:
			esf.seasonalFactors[month] = 1.0
		case month >= 9 && month <= 11:
			esf.seasonalFactors[month] = 0.9
		default:
			esf.seasonalFactors[month] = 0.8
		}
	}
}

func (esf *EnergyStorageForecaster) calculateEffectiveCapacity() float64 {
	degradationFactor := 1.0 - esf.degradationRate*float64(esf.cycleCount)
	return esf.batteryCapacity * math.Max(0.5, degradationFactor)
}

func (esf *EnergyStorageForecaster) charge(currentSOC, power, durationHours float64) (float64, float64) {
	effectiveCapacity := esf.calculateEffectiveCapacity()
	maxEnergy := effectiveCapacity * (esf.maxSOC - currentSOC) / 100.0
	
	chargeEnergy := power * durationHours * esf.chargeEfficiency
	actualCharge := math.Min(chargeEnergy, maxEnergy)
	
	newSOC := currentSOC + (actualCharge / effectiveCapacity) * 100.0
	newSOC = math.Min(esf.maxSOC, newSOC)
	
	return newSOC, actualCharge
}

func (esf *EnergyStorageForecaster) discharge(currentSOC, power, durationHours float64) (float64, float64) {
	effectiveCapacity := esf.calculateEffectiveCapacity()
	minEnergy := effectiveCapacity * (currentSOC - esf.minSOC) / 100.0
	
	dischargeEnergy := power * durationHours / esf.dischargeEfficiency
	actualDischarge := math.Min(dischargeEnergy, minEnergy)
	
	newSOC := currentSOC - (actualDischarge / effectiveCapacity) * 100.0
	newSOC = math.Max(esf.minSOC, newSOC)
	
	return newSOC, actualDischarge
}

func (esf *EnergyStorageForecaster) estimatePower(timestamp time.Time, isCharge bool) float64 {
	hour := timestamp.Hour()
	month := int(timestamp.Month())
	
	basePower := esf.batteryCapacity * 0.2
	seasonalFactor := esf.seasonalFactors[month]
	
	var dailyFactor float64
	if isCharge {
		dailyFactor = esf.dailyChargePattern[hour]
	} else {
		dailyFactor = esf.dailyDischargePattern[hour]
	}
	
	return basePower * seasonalFactor * dailyFactor
}

func (esf *EnergyStorageForecaster) Train(ctx context.Context, data []*TimeSeriesData) error {
	if len(data) < 48 {
		return fmt.Errorf("insufficient training data: need at least 48 hours")
	}
	
	esf.trainingData = make([]*TimeSeriesData, len(data))
	copy(esf.trainingData, data)
	
	chargeSums := make(map[int]float64)
	chargeCounts := make(map[int]int)
	dischargeSums := make(map[int]float64)
	dischargeCounts := make(map[int]int)
	monthlySums := make(map[int]float64)
	monthlyCounts := make(map[int]int)
	
	for _, dp := range data {
		hour := dp.Timestamp.Hour()
		month := int(dp.Timestamp.Month())
		power := dp.Value
		
		if power > 0 {
			chargeSums[hour] += power
			chargeCounts[hour]++
		} else if power < 0 {
			dischargeSums[hour] += math.Abs(power)
			dischargeCounts[hour]++
		}
		
		monthlySums[month] += math.Abs(power)
		monthlyCounts[month]++
	}
	
	maxChargeValue := 0.0
	for hour := 0; hour < 24; hour++ {
		if chargeCounts[hour] > 0 {
			avg := chargeSums[hour] / float64(chargeCounts[hour])
			if avg > maxChargeValue {
				maxChargeValue = avg
			}
		}
	}
	
	if maxChargeValue > 0 {
		for hour := 0; hour < 24; hour++ {
			if chargeCounts[hour] > 0 {
				esf.dailyChargePattern[hour] = chargeSums[hour] / float64(chargeCounts[hour]) / maxChargeValue
			}
		}
	}
	
	maxDischargeValue := 0.0
	for hour := 0; hour < 24; hour++ {
		if dischargeCounts[hour] > 0 {
			avg := dischargeSums[hour] / float64(dischargeCounts[hour])
			if avg > maxDischargeValue {
				maxDischargeValue = avg
			}
		}
	}
	
	if maxDischargeValue > 0 {
		for hour := 0; hour < 24; hour++ {
			if dischargeCounts[hour] > 0 {
				esf.dailyDischargePattern[hour] = dischargeSums[hour] / float64(dischargeCounts[hour]) / maxDischargeValue
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
				esf.seasonalFactors[month] = monthlySums[month] / float64(monthlyCounts[month]) / maxMonthlyValue
			}
		}
	}
	
	now := time.Now()
	esf.modelInfo.TrainedAt = &now
	esf.modelInfo.Status = "trained"
	esf.currentSOC = esf.initialSOC
	
	return nil
}

func (esf *EnergyStorageForecaster) Predict(ctx context.Context, horizon int) ([]*Prediction, error) {
	if esf.modelInfo.Status != "trained" || len(esf.trainingData) == 0 {
		return nil, ErrModelNotTrained
	}
	
	lastTime := esf.trainingData[len(esf.trainingData)-1].Timestamp
	predictions := make([]*Prediction, horizon)
	
	currentSOC := esf.currentSOC
	
	for i := 0; i < horizon; i++ {
		predictionTime := lastTime.Add(time.Duration(i+1) * time.Hour)
		
		chargePower := esf.estimatePower(predictionTime, true)
		dischargePower := esf.estimatePower(predictionTime, false)
		
		var netPower float64
		var newSOC float64
		
		if chargePower > dischargePower {
			netPower = chargePower - dischargePower
			newSOC, _ = esf.charge(currentSOC, netPower, 1.0)
		} else {
			netPower = -(dischargePower - chargePower)
			newSOC, _ = esf.discharge(currentSOC, dischargePower-chargePower, 1.0)
		}
		
		confidence := esf.calculateConfidence(predictionTime)
		
		predictions[i] = &Prediction{
			Timestamp:          predictionTime,
			Value:              netPower,
			Confidence:         confidence,
			ConfidenceInterval: [2]float64{netPower * 0.85, netPower * 1.15},
		}
		
		if math.Abs(netPower) > esf.batteryCapacity*0.01 {
			esf.cycleCount++
		}
		
		currentSOC = newSOC
	}
	
	return predictions, nil
}

func (esf *EnergyStorageForecaster) PredictSOC(ctx context.Context, horizon int) ([]float64, error) {
	if esf.modelInfo.Status != "trained" || len(esf.trainingData) == 0 {
		return nil, ErrModelNotTrained
	}
	
	predictions, err := esf.Predict(ctx, horizon)
	if err != nil {
		return nil, err
	}
	
	socPredictions := make([]float64, horizon)
	currentSOC := esf.initialSOC
	
	for i, pred := range predictions {
		if pred.Value > 0 {
			currentSOC, _ = esf.charge(currentSOC, pred.Value, 1.0)
		} else if pred.Value < 0 {
			currentSOC, _ = esf.discharge(currentSOC, -pred.Value, 1.0)
		}
		socPredictions[i] = currentSOC
	}
	
	return socPredictions, nil
}

func (esf *EnergyStorageForecaster) calculateConfidence(timestamp time.Time) float64 {
	if esf.modelInfo.TrainedAt == nil {
		return 0.5
	}
	
	daysSinceTraining := time.Since(*esf.modelInfo.TrainedAt).Hours() / 24
	timeDecay := math.Max(0.5, 1.0-daysSinceTraining/30.0)
	
	degradationFactor := 1.0 - esf.degradationRate*float64(esf.cycleCount)
	degradationFactor = math.Max(0.5, degradationFactor)
	
	return 0.7 * timeDecay * degradationFactor
}

func (esf *EnergyStorageForecaster) GetModelInfo() *ModelInfo {
	return esf.modelInfo
}

func (esf *EnergyStorageForecaster) Save(ctx context.Context, path string) error {
	return nil
}

func (esf *EnergyStorageForecaster) Load(ctx context.Context, path string) error {
	return nil
}

type BatteryHealthForecaster struct {
	baseForecaster *EnergyStorageForecaster
	modelInfo      *ModelInfo
}

func NewBatteryHealthForecaster(modelID string, capacity float64) *BatteryHealthForecaster {
	return &BatteryHealthForecaster{
		baseForecaster: NewEnergyStorageForecaster(modelID, capacity),
		modelInfo: &ModelInfo{
			ModelID:   modelID,
			ModelType: "battery_health",
			Version:   "1.0.0",
			CreatedAt: time.Now(),
			Status:    "untrained",
			Parameters: map[string]interface{}{
				"battery_capacity": capacity,
			},
		},
	}
}

func (bhf *BatteryHealthForecaster) Train(ctx context.Context, data []*TimeSeriesData) error {
	err := bhf.baseForecaster.Train(ctx, data)
	if err != nil {
		return err
	}
	
	bhf.modelInfo.Status = "trained"
	bhf.modelInfo.TrainedAt = bhf.baseForecaster.modelInfo.TrainedAt
	
	return nil
}

func (bhf *BatteryHealthForecaster) Predict(ctx context.Context, horizon int) ([]*Prediction, error) {
	predictions := make([]*Prediction, horizon)
	lastTime := time.Now()
	
	baseHealth := 1.0 - bhf.baseForecaster.degradationRate*float64(bhf.baseForecaster.cycleCount)
	
	for i := 0; i < horizon; i++ {
		predictionTime := lastTime.Add(time.Duration(i+1) * 24 * time.Hour)
		
		health := baseHealth - bhf.baseForecaster.degradationRate*float64(i)*0.1
		health = math.Max(0.5, health)
		
		confidence := 0.8 - float64(i)*0.01
		confidence = math.Max(0.5, confidence)
		
		predictions[i] = &Prediction{
			Timestamp:          predictionTime,
			Value:              health * 100.0,
			Confidence:         confidence,
			ConfidenceInterval: [2]float64{health * 95.0, health * 105.0},
		}
	}
	
	return predictions, nil
}

func (bhf *BatteryHealthForecaster) GetModelInfo() *ModelInfo {
	return bhf.modelInfo
}

func (bhf *BatteryHealthForecaster) Save(ctx context.Context, path string) error {
	return bhf.baseForecaster.Save(ctx, path)
}

func (bhf *BatteryHealthForecaster) Load(ctx context.Context, path string) error {
	return bhf.baseForecaster.Load(ctx, path)
}
