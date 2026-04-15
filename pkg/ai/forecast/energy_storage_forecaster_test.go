package forecast

import (
	"context"
	"testing"
	"time"
)

func TestNewEnergyStorageForecaster(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_001", 1000.0)
	
	if esf == nil {
		t.Fatal("Expected EnergyStorageForecaster instance, got nil")
	}
	
	if esf.modelInfo.ModelID != "storage_test_001" {
		t.Errorf("Expected model ID 'storage_test_001', got '%s'", esf.modelInfo.ModelID)
	}
	
	if esf.modelInfo.ModelType != "energy_storage" {
		t.Errorf("Expected model type 'energy_storage', got '%s'", esf.modelInfo.ModelType)
	}
	
	if esf.batteryCapacity != 1000.0 {
		t.Errorf("Expected battery capacity 1000.0, got %f", esf.batteryCapacity)
	}
}

func TestEnergyStorageForecaster_SetEfficiency(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_002", 1000.0)
	
	esf.SetEfficiency(0.92, 0.93)
	
	if esf.chargeEfficiency != 0.92 {
		t.Errorf("Expected charge efficiency 0.92, got %f", esf.chargeEfficiency)
	}
	
	if esf.dischargeEfficiency != 0.93 {
		t.Errorf("Expected discharge efficiency 0.93, got %f", esf.dischargeEfficiency)
	}
}

func TestEnergyStorageForecaster_SetSOCLimits(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_003", 1000.0)
	
	esf.SetSOCLimits(15.0, 85.0)
	
	if esf.minSOC != 15.0 {
		t.Errorf("Expected min SOC 15.0, got %f", esf.minSOC)
	}
	
	if esf.maxSOC != 85.0 {
		t.Errorf("Expected max SOC 85.0, got %f", esf.maxSOC)
	}
}

func TestEnergyStorageForecaster_SetInitialSOC(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_004", 1000.0)
	
	esf.SetInitialSOC(60.0)
	
	if esf.initialSOC != 60.0 {
		t.Errorf("Expected initial SOC 60.0, got %f", esf.initialSOC)
	}
	
	if esf.currentSOC != 60.0 {
		t.Errorf("Expected current SOC 60.0, got %f", esf.currentSOC)
	}
}

func TestEnergyStorageForecaster_SetInitialSOCBounds(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_005", 1000.0)
	
	esf.SetInitialSOC(150.0)
	if esf.initialSOC != 100.0 {
		t.Errorf("Expected initial SOC capped at 100.0, got %f", esf.initialSOC)
	}
	
	esf.SetInitialSOC(-20.0)
	if esf.initialSOC != 0.0 {
		t.Errorf("Expected initial SOC floored at 0.0, got %f", esf.initialSOC)
	}
}

func TestEnergyStorageForecaster_calculateEffectiveCapacity(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_006", 1000.0)
	
	initialCapacity := esf.calculateEffectiveCapacity()
	if initialCapacity != 1000.0 {
		t.Errorf("Expected initial capacity 1000.0, got %f", initialCapacity)
	}
	
	esf.cycleCount = 1000
	degradedCapacity := esf.calculateEffectiveCapacity()
	if degradedCapacity >= initialCapacity {
		t.Errorf("Expected degraded capacity to be less than initial, got %f vs %f", degradedCapacity, initialCapacity)
	}
}

func TestEnergyStorageForecaster_charge(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_007", 1000.0)
	esf.SetInitialSOC(50.0)
	
	newSOC, charged := esf.charge(50.0, 100.0, 1.0)
	
	if newSOC <= 50.0 {
		t.Errorf("Expected SOC to increase after charging, got %f", newSOC)
	}
	
	if charged <= 0 {
		t.Errorf("Expected positive charged energy, got %f", charged)
	}
}

func TestEnergyStorageForecaster_discharge(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_008", 1000.0)
	esf.SetInitialSOC(50.0)
	
	newSOC, discharged := esf.discharge(50.0, 100.0, 1.0)
	
	if newSOC >= 50.0 {
		t.Errorf("Expected SOC to decrease after discharging, got %f", newSOC)
	}
	
	if discharged <= 0 {
		t.Errorf("Expected positive discharged energy, got %f", discharged)
	}
}

func TestEnergyStorageForecaster_chargeMaxSOC(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_009", 1000.0)
	esf.SetInitialSOC(85.0)
	
	newSOC, _ := esf.charge(85.0, 500.0, 1.0)
	
	if newSOC > 90.0 {
		t.Errorf("Expected SOC not to exceed max SOC, got %f", newSOC)
	}
}

func TestEnergyStorageForecaster_dischargeMinSOC(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_010", 1000.0)
	esf.SetInitialSOC(15.0)
	
	newSOC, _ := esf.discharge(15.0, 500.0, 1.0)
	
	if newSOC < 10.0 {
		t.Errorf("Expected SOC not to go below min SOC, got %f", newSOC)
	}
}

func TestEnergyStorageForecaster_Train(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_011", 1000.0)
	
	ctx := context.Background()
	data := generateStorageTestData(24*14)
	
	err := esf.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	if esf.modelInfo.Status != "trained" {
		t.Errorf("Expected model status 'trained', got '%s'", esf.modelInfo.Status)
	}
	
	if esf.modelInfo.TrainedAt == nil {
		t.Error("Expected trained_at timestamp, got nil")
	}
}

func TestEnergyStorageForecaster_TrainInsufficientData(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_012", 1000.0)
	
	ctx := context.Background()
	data := generateStorageTestData(24)
	
	err := esf.Train(ctx, data)
	if err == nil {
		t.Fatal("Expected error for insufficient training data, got nil")
	}
}

func TestEnergyStorageForecaster_Predict(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_013", 1000.0)
	
	ctx := context.Background()
	data := generateStorageTestData(24*14)
	
	err := esf.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	horizon := 24
	predictions, err := esf.Predict(ctx, horizon)
	if err != nil {
		t.Fatalf("Expected no error during prediction, got %v", err)
	}
	
	if len(predictions) != horizon {
		t.Errorf("Expected %d predictions, got %d", horizon, len(predictions))
	}
	
	for i, pred := range predictions {
		if pred.Timestamp.IsZero() {
			t.Errorf("Prediction %d: Expected valid timestamp, got zero", i)
		}
		if pred.Confidence < 0 || pred.Confidence > 1 {
			t.Errorf("Prediction %d: Expected confidence between 0 and 1, got %f", i, pred.Confidence)
		}
	}
}

func TestEnergyStorageForecaster_PredictWithoutTraining(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_014", 1000.0)
	
	ctx := context.Background()
	_, err := esf.Predict(ctx, 24)
	if err == nil {
		t.Fatal("Expected error when predicting without training, got nil")
	}
}

func TestEnergyStorageForecaster_PredictSOC(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_015", 1000.0)
	
	ctx := context.Background()
	data := generateStorageTestData(24*14)
	
	err := esf.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	horizon := 24
	socPredictions, err := esf.PredictSOC(ctx, horizon)
	if err != nil {
		t.Fatalf("Expected no error during SOC prediction, got %v", err)
	}
	
	if len(socPredictions) != horizon {
		t.Errorf("Expected %d SOC predictions, got %d", horizon, len(socPredictions))
	}
	
	for i, soc := range socPredictions {
		if soc < 0 || soc > 100 {
			t.Errorf("SOC Prediction %d: Expected SOC between 0 and 100, got %f", i, soc)
		}
	}
}

func TestEnergyStorageForecaster_GetModelInfo(t *testing.T) {
	esf := NewEnergyStorageForecaster("storage_test_016", 1000.0)
	
	info := esf.GetModelInfo()
	if info == nil {
		t.Fatal("Expected ModelInfo instance, got nil")
	}
	
	if info.ModelID != "storage_test_016" {
		t.Errorf("Expected model ID 'storage_test_016', got '%s'", info.ModelID)
	}
}

func TestNewBatteryHealthForecaster(t *testing.T) {
	bhf := NewBatteryHealthForecaster("health_test_001", 1000.0)
	
	if bhf == nil {
		t.Fatal("Expected BatteryHealthForecaster instance, got nil")
	}
	
	if bhf.modelInfo.ModelID != "health_test_001" {
		t.Errorf("Expected model ID 'health_test_001', got '%s'", bhf.modelInfo.ModelID)
	}
	
	if bhf.modelInfo.ModelType != "battery_health" {
		t.Errorf("Expected model type 'battery_health', got '%s'", bhf.modelInfo.ModelType)
	}
}

func TestBatteryHealthForecaster_TrainAndPredict(t *testing.T) {
	bhf := NewBatteryHealthForecaster("health_test_002", 1000.0)
	
	ctx := context.Background()
	data := generateStorageTestData(24*14)
	
	err := bhf.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	predictions, err := bhf.Predict(ctx, 30)
	if err != nil {
		t.Fatalf("Expected no error during prediction, got %v", err)
	}
	
	if len(predictions) != 30 {
		t.Errorf("Expected 30 health predictions, got %d", len(predictions))
	}
	
	for i, pred := range predictions {
		if pred.Value < 50.0 || pred.Value > 100.0 {
			t.Errorf("Health Prediction %d: Expected health between 50 and 100, got %f", i, pred.Value)
		}
	}
}

func TestBatteryHealthForecaster_DecliningHealth(t *testing.T) {
	bhf := NewBatteryHealthForecaster("health_test_003", 1000.0)
	
	ctx := context.Background()
	data := generateStorageTestData(24*14)
	
	err := bhf.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	predictions, err := bhf.Predict(ctx, 30)
	if err != nil {
		t.Fatalf("Expected no error during prediction, got %v", err)
	}
	
	if len(predictions) < 2 {
		t.Skip("Need at least 2 predictions for this test")
		return
	}
	
	firstHealth := predictions[0].Value
	lastHealth := predictions[len(predictions)-1].Value
	
	if lastHealth > firstHealth {
		t.Errorf("Expected battery health to decline over time, got %f then %f", firstHealth, lastHealth)
	}
}

func generateStorageTestData(hours int) []*TimeSeriesData {
	data := make([]*TimeSeriesData, hours)
	baseTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	for i := 0; i < hours; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Hour)
		hour := timestamp.Hour()
		month := int(timestamp.Month())
		
		var power float64
		switch {
		case hour >= 6 && hour < 12:
			power = 200.0
		case hour >= 18 && hour < 24:
			power = -200.0
		default:
			power = 0.0
		}
		
		seasonalFactor := 1.0
		switch {
		case month >= 6 && month <= 8:
			seasonalFactor = 1.2
		case month >= 3 && month <= 5:
			seasonalFactor = 1.0
		case month >= 9 && month <= 11:
			seasonalFactor = 0.9
		default:
			seasonalFactor = 0.8
		}
		
		power *= seasonalFactor
		power += (randStorageFloat() - 0.5) * 40.0
		
		data[i] = &TimeSeriesData{
			Timestamp: timestamp,
			Value:     power,
		}
	}
	
	return data
}

func randStorageFloat() float64 {
	return 0.5
}
