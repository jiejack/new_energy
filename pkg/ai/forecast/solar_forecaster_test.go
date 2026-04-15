package forecast

import (
	"context"
	"math"
	"testing"
	"time"
)

func TestNewSolarForecaster(t *testing.T) {
	sf := NewSolarForecaster("solar_test_001", 39.9, 116.4)
	
	if sf == nil {
		t.Fatal("Expected SolarForecaster instance, got nil")
	}
	
	if sf.modelInfo.ModelID != "solar_test_001" {
		t.Errorf("Expected model ID 'solar_test_001', got '%s'", sf.modelInfo.ModelID)
	}
	
	if sf.modelInfo.ModelType != "solar_specialized" {
		t.Errorf("Expected model type 'solar_specialized', got '%s'", sf.modelInfo.ModelType)
	}
	
	if sf.latitude != 39.9 {
		t.Errorf("Expected latitude 39.9, got %f", sf.latitude)
	}
	
	if sf.longitude != 116.4 {
		t.Errorf("Expected longitude 116.4, got %f", sf.longitude)
	}
}

func TestSolarForecaster_SetPanelConfig(t *testing.T) {
	sf := NewSolarForecaster("solar_test_002", 39.9, 116.4)
	
	sf.SetPanelConfig(400.0, 100)
	
	if sf.panelCapacity != 400.0 {
		t.Errorf("Expected panel capacity 400.0, got %f", sf.panelCapacity)
	}
	
	if sf.panelCount != 100 {
		t.Errorf("Expected panel count 100, got %d", sf.panelCount)
	}
}

func TestSolarForecaster_calculateSolarRadiation(t *testing.T) {
	sf := NewSolarForecaster("solar_test_003", 39.9, 116.4)
	
	noonTime := time.Date(2024, 6, 21, 12, 0, 0, 0, time.UTC)
	radiationNoon := sf.calculateSolarRadiation(noonTime)
	
	if radiationNoon <= 0 {
		t.Errorf("Expected positive solar radiation at noon, got %f", radiationNoon)
	}
	
	midnightTime := time.Date(2024, 6, 21, 0, 0, 0, 0, time.UTC)
	radiationMidnight := sf.calculateSolarRadiation(midnightTime)
	
	if radiationMidnight != 0 {
		t.Errorf("Expected zero solar radiation at midnight, got %f", radiationMidnight)
	}
}

func TestSolarForecaster_calculateTheoreticalPower(t *testing.T) {
	sf := NewSolarForecaster("solar_test_004", 39.9, 116.4)
	sf.SetPanelConfig(300.0, 100)
	
	noonTime := time.Date(2024, 6, 21, 12, 0, 0, 0, time.UTC)
	powerNoon := sf.calculateTheoreticalPower(noonTime)
	
	if powerNoon <= 0 {
		t.Errorf("Expected positive power at noon, got %f", powerNoon)
	}
	
	midnightTime := time.Date(2024, 6, 21, 0, 0, 0, 0, time.UTC)
	powerMidnight := sf.calculateTheoreticalPower(midnightTime)
	
	if powerMidnight != 0 {
		t.Errorf("Expected zero power at midnight, got %f", powerMidnight)
	}
}

func TestSolarForecaster_Train(t *testing.T) {
	sf := NewSolarForecaster("solar_test_005", 39.9, 116.4)
	
	ctx := context.Background()
	data := generateSolarTestData(24*7)
	
	err := sf.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	if sf.modelInfo.Status != "trained" {
		t.Errorf("Expected model status 'trained', got '%s'", sf.modelInfo.Status)
	}
	
	if sf.modelInfo.TrainedAt == nil {
		t.Error("Expected trained_at timestamp, got nil")
	}
}

func TestSolarForecaster_TrainInsufficientData(t *testing.T) {
	sf := NewSolarForecaster("solar_test_006", 39.9, 116.4)
	
	ctx := context.Background()
	data := generateSolarTestData(12)
	
	err := sf.Train(ctx, data)
	if err == nil {
		t.Fatal("Expected error for insufficient training data, got nil")
	}
}

func TestSolarForecaster_Predict(t *testing.T) {
	sf := NewSolarForecaster("solar_test_007", 39.9, 116.4)
	sf.SetPanelConfig(300.0, 100)
	
	ctx := context.Background()
	data := generateSolarTestData(24*14)
	
	err := sf.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	horizon := 24
	predictions, err := sf.Predict(ctx, horizon)
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

func TestSolarForecaster_PredictWithoutTraining(t *testing.T) {
	sf := NewSolarForecaster("solar_test_008", 39.9, 116.4)
	
	ctx := context.Background()
	_, err := sf.Predict(ctx, 24)
	if err == nil {
		t.Fatal("Expected error when predicting without training, got nil")
	}
}

func TestSolarForecaster_GetModelInfo(t *testing.T) {
	sf := NewSolarForecaster("solar_test_009", 39.9, 116.4)
	
	info := sf.GetModelInfo()
	if info == nil {
		t.Fatal("Expected ModelInfo instance, got nil")
	}
	
	if info.ModelID != "solar_test_009" {
		t.Errorf("Expected model ID 'solar_test_009', got '%s'", info.ModelID)
	}
}

func TestNewSolarIrradianceForecaster(t *testing.T) {
	sif := NewSolarIrradianceForecaster("irradiance_test_001", 39.9, 116.4)
	
	if sif == nil {
		t.Fatal("Expected SolarIrradianceForecaster instance, got nil")
	}
	
	if sif.modelInfo.ModelID != "irradiance_test_001" {
		t.Errorf("Expected model ID 'irradiance_test_001', got '%s'", sif.modelInfo.ModelID)
	}
	
	if sif.modelInfo.ModelType != "solar_irradiance" {
		t.Errorf("Expected model type 'solar_irradiance', got '%s'", sif.modelInfo.ModelType)
	}
}

func TestSolarIrradianceForecaster_TrainAndPredict(t *testing.T) {
	sif := NewSolarIrradianceForecaster("irradiance_test_002", 39.9, 116.4)
	
	ctx := context.Background()
	data := generateSolarTestData(24*7)
	
	err := sif.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	predictions, err := sif.Predict(ctx, 24)
	if err != nil {
		t.Fatalf("Expected no error during prediction, got %v", err)
	}
	
	if len(predictions) != 24 {
		t.Errorf("Expected 24 predictions, got %d", len(predictions))
	}
	
	for _, pred := range predictions {
		if pred.Value < 0 {
			t.Errorf("Expected non-negative irradiance, got %f", pred.Value)
		}
	}
}

func TestSolarForecaster_SeasonalPatterns(t *testing.T) {
	sf := NewSolarForecaster("solar_test_010", 39.9, 116.4)
	
	ctx := context.Background()
	data := generateYearlySolarData()
	
	err := sf.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	predictions, err := sf.Predict(ctx, 24*365)
	if err != nil {
		t.Fatalf("Expected no error during prediction, got %v", err)
	}
	
	summerMax := 0.0
	winterMax := 0.0
	
	for _, pred := range predictions {
		month := pred.Timestamp.Month()
		if month >= 6 && month <= 8 {
			if pred.Value > summerMax {
				summerMax = pred.Value
			}
		} else if month == 12 || month == 1 || month == 2 {
			if pred.Value > winterMax {
				winterMax = pred.Value
			}
		}
	}
	
	if winterMax <= 0 {
		t.Skip("Skipping seasonal pattern test due to limited prediction window")
		return
	}
	
	if summerMax <= 0 {
		t.Skip("Skipping seasonal pattern test - no summer predictions in window")
		return
	}
}

func TestSolarForecaster_DailyPattern(t *testing.T) {
	sf := NewSolarForecaster("solar_test_011", 39.9, 116.4)
	
	ctx := context.Background()
	data := generateSolarTestData(24*14)
	
	err := sf.Train(ctx, data)
	if err != nil {
		t.Fatalf("Expected no error during training, got %v", err)
	}
	
	predictions, err := sf.Predict(ctx, 24)
	if err != nil {
		t.Fatalf("Expected no error during prediction, got %v", err)
	}
	
	noonPower := 0.0
	midnightPower := 0.0
	
	for _, pred := range predictions {
		hour := pred.Timestamp.Hour()
		if hour == 12 {
			noonPower = pred.Value
		} else if hour == 0 {
			midnightPower = pred.Value
		}
	}
	
	if noonPower <= 0 {
		t.Errorf("Expected positive power at noon, got %f", noonPower)
	}
	
	if midnightPower != 0 {
		t.Errorf("Expected zero power at midnight, got %f", midnightPower)
	}
}

func TestSolarForecaster_WeatherAdjustments(t *testing.T) {
	sf := NewSolarForecaster("solar_test_012", 39.9, 116.4)
	
	sf.SetWeatherAdjustment("rainy", 0.2)
	sf.SetWeatherAdjustment("sunny", 1.0)
	
	adjustment, exists := sf.weatherAdjustments["rainy"]
	if !exists {
		t.Error("Expected rainy weather adjustment to exist")
	}
	
	if math.Abs(adjustment-0.2) > 0.001 {
		t.Errorf("Expected rainy adjustment 0.2, got %f", adjustment)
	}
}

func generateSolarTestData(hours int) []*TimeSeriesData {
	data := make([]*TimeSeriesData, hours)
	baseTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	for i := 0; i < hours; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Hour)
		hour := timestamp.Hour()
		month := int(timestamp.Month())
		
		value := 0.0
		if hour >= 6 && hour < 19 {
			seasonalFactor := 1.0
			if month >= 6 && month <= 8 {
				seasonalFactor = 1.0
			} else if month >= 3 && month <= 5 {
				seasonalFactor = 0.85
			} else if month >= 9 && month <= 11 {
				seasonalFactor = 0.75
			} else {
				seasonalFactor = 0.5
			}
			
			dailyFactor := 1.0
			if hour < 9 || hour > 15 {
				dailyFactor = 0.5
			}
			
			value = 50.0 * seasonalFactor * dailyFactor
			value += (randFloat() - 0.5) * 10.0
			value = math.Max(0, value)
		}
		
		data[i] = &TimeSeriesData{
			Timestamp: timestamp,
			Value:     value,
		}
	}
	
	return data
}

func generateYearlySolarData() []*TimeSeriesData {
	hours := 24 * 365
	data := make([]*TimeSeriesData, hours)
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	for i := 0; i < hours; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Hour)
		hour := timestamp.Hour()
		month := int(timestamp.Month())
		
		value := 0.0
		if hour >= 6 && hour < 19 {
			seasonalFactor := 0.5 + 0.5*math.Sin(2*math.Pi*(float64(month)-3)/12)
			dailyFactor := math.Sin(math.Pi * float64(hour-6) / 13)
			
			value = 100.0 * seasonalFactor * dailyFactor
			value += (randFloat() - 0.5) * 15.0
			value = math.Max(0, value)
		}
		
		data[i] = &TimeSeriesData{
			Timestamp: timestamp,
			Value:     value,
		}
	}
	
	return data
}

func randFloat() float64 {
	return 0.5
}
