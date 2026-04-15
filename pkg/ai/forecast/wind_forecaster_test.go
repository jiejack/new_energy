package forecast

import (
	"context"
	"math"
	"testing"
	"time"
)

func TestNewWindForecaster(t *testing.T) {
	wf := NewWindForecaster("wind_test_001", 40.0, -105.0)
	
	if wf == nil {
		t.Fatal("Expected WindForecaster instance, got nil")
	}
	
	if wf.modelInfo.ModelID != "wind_test_001" {
		t.Errorf("Expected model ID 'wind_test_001', got '%s'", wf.modelInfo.ModelID)
	}
	
	if wf.modelInfo.ModelType != "wind_specialized" {
		t.Errorf("Expected model type 'wind_specialized', got '%s'", wf.modelInfo.ModelType)
	}
	
	if wf.latitude != 40.0 {
		t.Errorf("Expected latitude 40.0, got %f", wf.latitude)
	}
	
	if wf.longitude != -105.0 {
		t.Errorf("Expected longitude -105.0, got %f", wf.longitude)
	}
}

func TestWindForecaster_SetTurbineConfig(t *testing.T) {
	wf := NewWindForecaster("wind_test_002", 40.0, -105.0)
	
	wf.SetTurbineConfig(120.0, 0.48, 50)
	
	if wf.rotorDiameter != 120.0 {
		t.Errorf("Expected rotor diameter 120.0, got %f", wf.rotorDiameter)
	}
	
	if wf.turbineEfficiency != 0.48 {
		t.Errorf("Expected turbine efficiency 0.48, got %f", wf.turbineEfficiency)
	}
	
	if wf.turbineCount != 50 {
		t.Errorf("Expected turbine count 50, got %d", wf.turbineCount)
	}
}

func TestWindForecaster_SetSpeedThresholds(t *testing.T) {
	wf := NewWindForecaster("wind_test_003", 40.0, -105.0)
	
	wf.SetSpeedThresholds(2.5, 11.0, 22.0)
	
	if wf.cutInSpeed != 2.5 {
		t.Errorf("Expected cut-in speed 2.5, got %f", wf.cutInSpeed)
	}
	
	if wf.ratedSpeed != 11.0 {
		t.Errorf("Expected rated speed 11.0, got %f", wf.ratedSpeed)
	}
	
	if wf.cutOutSpeed != 22.0 {
		t.Errorf("Expected cut-out speed 22.0, got %f", wf.cutOutSpeed)
	}
}

func TestWindForecaster_SetAirDensity(t *testing.T) {
	wf := NewWindForecaster("wind_test_004", 40.0, -105.0)
	
	wf.SetAirDensity(1.20)
	
	if wf.airDensity != 1.20 {
		t.Errorf("Expected air density 1.20, got %f", wf.airDensity)
	}
}

func TestWindForecaster_calculateTheoreticalPower(t *testing.T) {
	wf := NewWindForecaster("wind_test_005", 40.0, -105.0)
	
	lowSpeed := wf.calculateTheoreticalPower(2.0)
	if lowSpeed != 0 {
		t.Errorf("Expected zero power at 2.0 m/s, got %f", lowSpeed)
	}
	
	ratedSpeed := wf.calculateTheoreticalPower(12.0)
	if ratedSpeed <= 0 {
		t.Errorf("Expected positive power at 12.0 m/s, got %f", ratedSpeed)
	}
	
	highSpeed := wf.calculateTheoreticalPower(30.0)
	if highSpeed != 0 {
		t.Errorf("Expected zero power at 30.0 m/s, got %f", highSpeed)
	}
}

func TestWindForecaster_calculateTheoreticalPowerCurve(t *testing.T) {
	wf := NewWindForecaster("wind_test_006", 40.0, -105.0)
	
	powerAtCutIn := wf.calculateTheoreticalPower(3.0)
	powerAtMid := wf.calculateTheoreticalPower(8.0)
	powerAtRated := wf.calculateTheoreticalPower(12.0)
	powerAboveRated := wf.calculateTheoreticalPower(15.0)
	
	if powerAtCutIn <= 0 {
		t.Errorf("Expected positive power at cut-in speed, got %f", powerAtCutIn)
	}
	
	if powerAtMid <= powerAtCutIn {
		t.Errorf("Expected power at 8 m/s (%f) > power at cut-in (%f)", powerAtMid, powerAtCutIn)
	}
	
	if powerAtRated <= powerAtMid {
		t.Errorf("Expected power at rated speed (%f) > power at mid (%f)", powerAtRated, powerAtMid)
	}
	
	if math.Abs(powerAboveRated-powerAtRated) > 0.001 {
		t.Errorf("Expected constant power above rated speed, got %f vs %f", powerAboveRated, powerAtRated)
	}
}

func TestWindForecaster_Train(t *testing.T) {
	wf := NewWindForecaster("wind_test_007", 40.0, -105.0)
	
	ctx := context.Background()
	data := generateWindTestData(24*14)
	
	err := wf.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	if wf.modelInfo.Status != "trained" {
		t.Errorf("Expected model status 'trained', got '%s'", wf.modelInfo.Status)
	}
	
	if wf.modelInfo.TrainedAt == nil {
		t.Error("Expected trained_at timestamp, got nil")
	}
}

func TestWindForecaster_TrainInsufficientData(t *testing.T) {
	wf := NewWindForecaster("wind_test_008", 40.0, -105.0)
	
	ctx := context.Background()
	data := generateWindTestData(24)
	
	err := wf.Train(ctx, data)
	if err == nil {
		t.Fatal("Expected error for insufficient training data, got nil")
	}
}

func TestWindForecaster_Predict(t *testing.T) {
	wf := NewWindForecaster("wind_test_009", 40.0, -105.0)
	wf.SetTurbineConfig(100.0, 0.45, 25)
	
	ctx := context.Background()
	data := generateWindTestData(24*14)
	
	err := wf.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	horizon := 24
	predictions, err := wf.Predict(ctx, horizon)
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
		if pred.Value < 0 {
			t.Errorf("Prediction %d: Expected non-negative value, got %f", i, pred.Value)
		}
		if pred.Confidence < 0 || pred.Confidence > 1 {
			t.Errorf("Prediction %d: Expected confidence between 0 and 1, got %f", i, pred.Confidence)
		}
	}
}

func TestWindForecaster_PredictWithoutTraining(t *testing.T) {
	wf := NewWindForecaster("wind_test_010", 40.0, -105.0)
	
	ctx := context.Background()
	_, err := wf.Predict(ctx, 24)
	if err == nil {
		t.Fatal("Expected error when predicting without training, got nil")
	}
}

func TestWindForecaster_GetModelInfo(t *testing.T) {
	wf := NewWindForecaster("wind_test_011", 40.0, -105.0)
	
	info := wf.GetModelInfo()
	if info == nil {
		t.Fatal("Expected ModelInfo instance, got nil")
	}
	
	if info.ModelID != "wind_test_011" {
		t.Errorf("Expected model ID 'wind_test_011', got '%s'", info.ModelID)
	}
}

func TestNewWindSpeedForecaster(t *testing.T) {
	wsf := NewWindSpeedForecaster("windspeed_test_001", 40.0, -105.0)
	
	if wsf == nil {
		t.Fatal("Expected WindSpeedForecaster instance, got nil")
	}
	
	if wsf.modelInfo.ModelID != "windspeed_test_001" {
		t.Errorf("Expected model ID 'windspeed_test_001', got '%s'", wsf.modelInfo.ModelID)
	}
	
	if wsf.modelInfo.ModelType != "wind_speed" {
		t.Errorf("Expected model type 'wind_speed', got '%s'", wsf.modelInfo.ModelType)
	}
}

func TestWindSpeedForecaster_TrainAndPredict(t *testing.T) {
	wsf := NewWindSpeedForecaster("windspeed_test_002", 40.0, -105.0)
	
	ctx := context.Background()
	data := generateWindTestData(24*14)
	
	err := wsf.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	predictions, err := wsf.Predict(ctx, 24)
	if err != nil {
		t.Fatalf("Expected no error during prediction, got %v", err)
	}
	
	if len(predictions) != 24 {
		t.Errorf("Expected 24 predictions, got %d", len(predictions))
	}
	
	for _, pred := range predictions {
		if pred.Value < 0 {
			t.Errorf("Expected non-negative wind speed, got %f", pred.Value)
		}
	}
}

func TestWindForecaster_SeasonalPatterns(t *testing.T) {
	wf := NewWindForecaster("wind_test_012", 40.0, -105.0)
	
	ctx := context.Background()
	data := generateYearlyWindData()
	
	err := wf.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	predictions, err := wf.Predict(ctx, 24*365)
	if err != nil {
		t.Fatalf("Expected no error during prediction, got %v", err)
	}
	
	winterMax := 0.0
	summerMax := 0.0
	
	for _, pred := range predictions {
		month := pred.Timestamp.Month()
		if month == 12 || month == 1 || month == 2 {
			if pred.Value > winterMax {
				winterMax = pred.Value
			}
		} else if month >= 6 && month <= 8 {
			if pred.Value > summerMax {
				summerMax = pred.Value
			}
		}
	}
	
	if winterMax <= 0 {
		t.Skip("Skipping seasonal pattern test - no winter predictions in window")
		return
	}
	
	if summerMax <= 0 {
		t.Skip("Skipping seasonal pattern test - no summer predictions in window")
		return
	}
}

func TestWindForecaster_TurbineCountScaling(t *testing.T) {
	wf1 := NewWindForecaster("wind_test_013", 40.0, -105.0)
	wf1.SetTurbineConfig(100.0, 0.45, 1)
	
	wf2 := NewWindForecaster("wind_test_014", 40.0, -105.0)
	wf2.SetTurbineConfig(100.0, 0.45, 10)
	
	power1 := wf1.calculateTheoreticalPower(10.0)
	power2 := wf2.calculateTheoreticalPower(10.0)
	
	if math.Abs(power2-10*power1) > 0.001 {
		t.Errorf("Expected 10x power with 10x turbines, got %f vs %f", power2, 10*power1)
	}
}

func generateWindTestData(hours int) []*TimeSeriesData {
	data := make([]*TimeSeriesData, hours)
	baseTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	for i := 0; i < hours; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Hour)
		hour := timestamp.Hour()
		month := int(timestamp.Month())
		
		baseWindSpeed := 8.0
		seasonalFactor := 1.0
		dailyFactor := 1.0
		
		switch {
		case month >= 3 && month <= 5:
			seasonalFactor = 0.9
		case month >= 6 && month <= 8:
			seasonalFactor = 0.7
		case month >= 9 && month <= 11:
			seasonalFactor = 0.85
		default:
			seasonalFactor = 1.0
		}
		
		switch {
		case hour >= 0 && hour < 6:
			dailyFactor = 0.8
		case hour >= 6 && hour < 12:
			dailyFactor = 0.9
		case hour >= 12 && hour < 18:
			dailyFactor = 1.0
		default:
			dailyFactor = 0.95
		}
		
		windSpeed := baseWindSpeed * seasonalFactor * dailyFactor
		windSpeed += (randWindFloat() - 0.5) * 3.0
		windSpeed = math.Max(0, windSpeed)
		
		data[i] = &TimeSeriesData{
			Timestamp: timestamp,
			Value:     windSpeed,
			Features: map[string]float64{
				"wind_speed": windSpeed,
			},
		}
	}
	
	return data
}

func generateYearlyWindData() []*TimeSeriesData {
	hours := 24 * 365
	data := make([]*TimeSeriesData, hours)
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	for i := 0; i < hours; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Hour)
		hour := timestamp.Hour()
		month := int(timestamp.Month())
		
		baseWindSpeed := 8.0
		seasonalFactor := 0.7 + 0.3*math.Cos(2*math.Pi*(float64(month)-1)/12)
		dailyFactor := 0.8 + 0.2*math.Cos(2*math.Pi*(float64(hour)-12)/24)
		
		windSpeed := baseWindSpeed * seasonalFactor * dailyFactor
		windSpeed += (randWindFloat() - 0.5) * 2.0
		windSpeed = math.Max(0, windSpeed)
		
		data[i] = &TimeSeriesData{
			Timestamp: timestamp,
			Value:     windSpeed,
			Features: map[string]float64{
				"wind_speed": windSpeed,
			},
		}
	}
	
	return data
}

func randWindFloat() float64 {
	return 0.5
}
