package fault

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateMetricHealth_Temperature(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "solar")

	tests := []struct {
		name     string
		value    float64
		minScore float64
		maxScore float64
	}{
		{"optimal 30", 30, 99, 101},
		{"good 20", 20, 89, 91},
		{"good 40", 40, 69, 71},
		{"poor 0", 0, 39, 41},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			point := &TimeSeriesData{DeviceID: "dev-1", Metric: "temperature", Value: tt.value, Timestamp: time.Now()}
			score := a.calculateMetricHealth(point)
			assert.True(t, score >= tt.minScore && score <= tt.maxScore, "score=%f, expected range [%f,%f]", score, tt.minScore, tt.maxScore)
		})
	}
}

func TestCalculateMetricHealth_Voltage_Solar(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "solar")

	tests := []struct {
		name     string
		value    float64
		minScore float64
		maxScore float64
	}{
		{"optimal 500", 500, 99, 101},
		{"good 350", 350, 84, 86},
		{"poor 200", 200, 39, 41},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			point := &TimeSeriesData{DeviceID: "dev-1", Metric: "voltage", Value: tt.value, Timestamp: time.Now()}
			score := a.calculateMetricHealth(point)
			assert.True(t, score >= tt.minScore && score <= tt.maxScore, "score=%f", score)
		})
	}
}

func TestCalculateMetricHealth_Voltage_Wind(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "wind")

	point := &TimeSeriesData{DeviceID: "dev-1", Metric: "voltage", Value: 700, Timestamp: time.Now()}
	score := a.calculateMetricHealth(point)
	assert.Equal(t, 100.0, score)

	point = &TimeSeriesData{DeviceID: "dev-1", Metric: "voltage", Value: 200, Timestamp: time.Now()}
	score = a.calculateMetricHealth(point)
	assert.Equal(t, 40.0, score)
}

func TestCalculateMetricHealth_Voltage_Battery(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "battery")

	point := &TimeSeriesData{DeviceID: "dev-1", Metric: "voltage", Value: 3.5, Timestamp: time.Now()}
	score := a.calculateMetricHealth(point)
	assert.Equal(t, 100.0, score)

	point = &TimeSeriesData{DeviceID: "dev-1", Metric: "voltage", Value: 2.0, Timestamp: time.Now()}
	score = a.calculateMetricHealth(point)
	assert.Equal(t, 40.0, score)
}

func TestCalculateMetricHealth_Voltage_Unknown(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "unknown")
	point := &TimeSeriesData{DeviceID: "dev-1", Metric: "voltage", Value: 500, Timestamp: time.Now()}
	score := a.calculateMetricHealth(point)
	assert.Equal(t, 70.0, score)
}

func TestCalculateMetricHealth_SOC(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "battery")

	point := &TimeSeriesData{DeviceID: "dev-1", Metric: "soc", Value: 50, Timestamp: time.Now()}
	score := a.calculateMetricHealth(point)
	assert.Equal(t, 100.0, score)

	point = &TimeSeriesData{DeviceID: "dev-1", Metric: "soc", Value: 15, Timestamp: time.Now()}
	score = a.calculateMetricHealth(point)
	assert.True(t, score >= 60 && score <= 100, "score=%f", score)

	point = &TimeSeriesData{DeviceID: "dev-1", Metric: "soc", Value: 85, Timestamp: time.Now()}
	score = a.calculateMetricHealth(point)
	assert.True(t, score >= 0 && score <= 60, "score=%f", score)

	point = &TimeSeriesData{DeviceID: "dev-1", Metric: "soc", Value: 5, Timestamp: time.Now()}
	score = a.calculateMetricHealth(point)
	assert.Equal(t, 20.0, score)
}

func TestCalculateMetricHealth_Vibration(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "wind")

	point := &TimeSeriesData{DeviceID: "dev-1", Metric: "vibration", Value: 0.5, Timestamp: time.Now()}
	score := a.calculateMetricHealth(point)
	assert.Equal(t, 100.0, score)

	point = &TimeSeriesData{DeviceID: "dev-1", Metric: "vibration", Value: 6, Timestamp: time.Now()}
	score = a.calculateMetricHealth(point)
	assert.Equal(t, 20.0, score)
}

func TestCalculateMetricHealth_UnknownMetric(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "solar")
	point := &TimeSeriesData{DeviceID: "dev-1", Metric: "unknown_metric", Value: 50, Timestamp: time.Now()}
	score := a.calculateMetricHealth(point)
	assert.Equal(t, 70.0, score)
}

func TestDetermineHealthStatus_All(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "solar")

	tests := []struct {
		score  float64
		status HealthStatus
	}{
		{95, HealthStatusExcellent},
		{90, HealthStatusExcellent},
		{80, HealthStatusGood},
		{75, HealthStatusGood},
		{65, HealthStatusFair},
		{60, HealthStatusFair},
		{50, HealthStatusPoor},
		{40, HealthStatusPoor},
		{30, HealthStatusCritical},
		{0, HealthStatusCritical},
	}

	for _, tt := range tests {
		result := a.determineHealthStatus(tt.score)
		assert.Equal(t, tt.status, result, "for score %f", tt.score)
	}
}

func TestIdentifyIssues(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "solar")

	tests := []struct {
		name     string
		metrics  map[string]float64
		expected int
	}{
		{"no issues", map[string]float64{"temperature": 90, "voltage": 85}, 0},
		{"temperature issue", map[string]float64{"temperature": 50}, 1},
		{"voltage issue", map[string]float64{"voltage": 30}, 1},
		{"soc issue", map[string]float64{"soc": 40}, 1},
		{"vibration issue", map[string]float64{"vibration": 30}, 1},
		{"unknown metric issue", map[string]float64{"custom_metric": 30}, 1},
		{"multiple issues", map[string]float64{"temperature": 30, "voltage": 20}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := a.identifyIssues(tt.metrics)
			assert.Len(t, issues, tt.expected)
		})
	}
}

func TestGetBaseRUL_All(t *testing.T) {
	tests := []struct {
		deviceType string
		expected   float64
	}{
		{"solar", 87600},
		{"wind", 70080},
		{"battery", 43800},
		{"unknown", 58400},
	}

	for _, tt := range tests {
		t.Run(tt.deviceType, func(t *testing.T) {
			a := NewSimpleHealthAssessor("dev-1", tt.deviceType)
			rul := a.getBaseRUL()
			assert.Equal(t, tt.expected, rul)
		})
	}
}

func TestDetermineHealthTrend(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "solar")

	assert.Equal(t, "stable", a.determineHealthTrend())

	a.healthHistory = append(a.healthHistory,
		&HealthAssessment{HealthScore: 80},
		&HealthAssessment{HealthScore: 85},
		&HealthAssessment{HealthScore: 90},
	)
	assert.Equal(t, "improving", a.determineHealthTrend())

	a.healthHistory = append(a.healthHistory,
		&HealthAssessment{HealthScore: 95},
		&HealthAssessment{HealthScore: 90},
		&HealthAssessment{HealthScore: 80},
	)
	assert.Equal(t, "declining", a.determineHealthTrend())

	a.healthHistory = append(a.healthHistory,
		&HealthAssessment{HealthScore: 80},
		&HealthAssessment{HealthScore: 81},
		&HealthAssessment{HealthScore: 82},
	)
	assert.Equal(t, "stable", a.determineHealthTrend())
}

func TestWindClassifier(t *testing.T) {
	classifier := NewRuleBasedClassifier("wind")

	anomaly := &Anomaly{
		DeviceID:  "dev-1",
		Metric:    "wind_speed",
		Value:     30,
		Severity:  SeverityHigh,
		Timestamp: time.Now(),
	}
	classification, err := classifier.Classify(context.Background(), anomaly)
	require.NoError(t, err)
	assert.Equal(t, "high_wind", classification.FaultType)
	assert.Equal(t, "WIND-001", classification.FaultCode)

	lowWindAnomaly := &Anomaly{
		DeviceID:  "dev-1",
		Metric:    "wind_speed",
		Value:     1,
		Severity:  SeverityMedium,
		Timestamp: time.Now(),
	}
	classification, err = classifier.Classify(context.Background(), lowWindAnomaly)
	require.NoError(t, err)
	assert.Equal(t, "low_wind", classification.FaultType)

	vibrationAnomaly := &Anomaly{
		DeviceID:  "dev-1",
		Metric:    "vibration",
		Value:     6,
		Severity:  SeverityHigh,
		Timestamp: time.Now(),
	}
	classification, err = classifier.Classify(context.Background(), vibrationAnomaly)
	require.NoError(t, err)
	assert.Equal(t, "excessive_vibration", classification.FaultType)
}

func TestBatteryClassifier(t *testing.T) {
	classifier := NewRuleBasedClassifier("battery")

	overheatingAnomaly := &Anomaly{
		DeviceID:  "dev-1",
		Metric:    "temperature",
		Value:     50,
		Severity:  SeverityHigh,
		Timestamp: time.Now(),
	}
	classification, err := classifier.Classify(context.Background(), overheatingAnomaly)
	require.NoError(t, err)
	assert.Equal(t, "battery_overheating", classification.FaultType)
	assert.Equal(t, "BAT-001", classification.FaultCode)

	overDischargeAnomaly := &Anomaly{
		DeviceID:  "dev-1",
		Metric:    "soc",
		Value:     15,
		Severity:  SeverityHigh,
		Timestamp: time.Now(),
	}
	classification, err = classifier.Classify(context.Background(), overDischargeAnomaly)
	require.NoError(t, err)
	assert.Equal(t, "over_discharge", classification.FaultType)

	overChargeAnomaly := &Anomaly{
		DeviceID:  "dev-1",
		Metric:    "soc",
		Value:     98,
		Severity:  SeverityHigh,
		Timestamp: time.Now(),
	}
	classification, err = classifier.Classify(context.Background(), overChargeAnomaly)
	require.NoError(t, err)
	assert.Equal(t, "over_charge", classification.FaultType)
}

func TestMLBasedClassifier(t *testing.T) {
	classifier := NewMLBasedClassifier("solar")
	require.NotNil(t, classifier)

	anomaly := &Anomaly{
		ID:        "anom-1",
		DeviceID:  "dev-1",
		Metric:    "temperature",
		Value:     65,
		Severity:  SeverityHigh,
		Timestamp: time.Now(),
	}

	classification, err := classifier.Classify(context.Background(), anomaly)
	require.NoError(t, err)
	assert.Equal(t, "ml_classified", classification.FaultType)
	assert.Equal(t, "ML-001", classification.FaultCode)
	assert.Equal(t, 0.8, classification.Confidence)

	err = classifier.Train(context.Background(), []*FaultLabeledData{
		{Anomaly: anomaly, FaultType: "overheating", FaultCode: "SOL-001", Label: true, Timestamp: time.Now()},
	})
	assert.NoError(t, err)

	info := classifier.GetClassifierInfo()
	assert.Equal(t, "MLBasedClassifier", info.ClassifierType)
	assert.NotNil(t, info.TrainedAt)
}

func TestCalculateConfidence_Classifier(t *testing.T) {
	classifier := NewRuleBasedClassifier("solar")

	tests := []struct {
		severity   AnomalySeverity
		confidence float64
	}{
		{SeverityCritical, 0.95},
		{SeverityHigh, 0.85},
		{SeverityMedium, 0.75},
		{SeverityLow, 0.65},
	}

	for _, tt := range tests {
		anomaly := &Anomaly{Severity: tt.severity}
		rule := FaultRule{FaultType: "test", FaultCode: "T-001"}
		conf := classifier.calculateConfidence(anomaly, rule)
		assert.Equal(t, tt.confidence, conf)
	}
}

func TestStatisticalDetector_CalculateSeverity(t *testing.T) {
	detector := NewStatisticalDetector("dev-1", "temp", 3, 2.0)

	tests := []struct {
		zScore   float64
		severity AnomalySeverity
	}{
		{4.5, SeverityCritical},
		{3.5, SeverityHigh},
		{2.5, SeverityMedium},
		{1.5, SeverityLow},
	}

	for _, tt := range tests {
		result := detector.calculateSeverity(tt.zScore)
		assert.Equal(t, tt.severity, result)
	}
}

func TestStatisticalDetector_CalculateConfidence(t *testing.T) {
	detector := NewStatisticalDetector("dev-1", "temp", 3, 2.0)

	conf := detector.calculateConfidence(2.0)
	assert.True(t, conf >= 0.5 && conf <= 0.99)

	conf = detector.calculateConfidence(4.0)
	assert.True(t, conf >= 0.5 && conf <= 0.99)
}

func TestRULPredictor(t *testing.T) {
	predictor := NewRULPredictor()
	require.NotNil(t, predictor)

	prediction, err := predictor.Predict(context.Background(), "dev-1")
	require.NoError(t, err)
	assert.Equal(t, "dev-1", prediction.DeviceID)
	assert.Equal(t, 2160.0, prediction.PredictedRUL)
	assert.Equal(t, 0.82, prediction.Confidence)
	assert.Equal(t, "declining", prediction.HealthTrend)
}

func TestCalculateWienerRUL(t *testing.T) {
	rul := CalculateWienerRUL(100, 20, 10)
	assert.Equal(t, 8.0, rul)

	rul = CalculateWienerRUL(20, 20, 10)
	assert.Equal(t, 0.0, rul)

	rul = CalculateWienerRUL(100, 20, 0)
	assert.True(t, rul > 0)

	rul = CalculateWienerRUL(100, 20, -1)
	assert.True(t, rul > 0)
}

func TestCalculateRULConfidence(t *testing.T) {
	conf := CalculateRULConfidence(100, 0.1)
	assert.True(t, conf > 0 && conf <= 1.0)

	conf = CalculateRULConfidence(0, 0.1)
	assert.Equal(t, 0.0, conf)

	conf = CalculateRULConfidence(-10, 0.1)
	assert.Equal(t, 0.0, conf)
}

func TestDetermineMaintenanceWindow(t *testing.T) {
	tests := []struct {
		rul      float64
		contains string
	}{
		{500, "紧急维护"},
		{1500, "预防性维护"},
		{3000, "计划维护"},
		{5000, "状态良好"},
	}

	for _, tt := range tests {
		result := DetermineMaintenanceWindow(tt.rul)
		assert.Contains(t, result, tt.contains)
	}
}

func TestThresholdDetector_CalculateSeverity(t *testing.T) {
	detector := NewThresholdDetector("dev-1", "temperature", 0, 100, 1, 5)

	tests := []struct {
		deviation float64
		severity  AnomalySeverity
	}{
		{150, SeverityCritical},
		{100, SeverityHigh},
		{60, SeverityMedium},
		{30, SeverityLow},
	}

	for _, tt := range tests {
		result := detector.calculateSeverity(tt.deviation)
		assert.Equal(t, tt.severity, result)
	}
}

func TestUpdateFaultEvent_WithAssessmentAndRUL(t *testing.T) {
	svc := NewFaultService()
	anomaly := &Anomaly{ID: "anom-1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}
	svc.createFaultEvent(anomaly, nil, nil, nil)

	assessment := &HealthAssessment{DeviceID: "dev-1", HealthScore: 80.0}
	svc.updateFaultEvent("anom-1", nil, assessment, nil)
	assert.Equal(t, "assessed", svc.faultEvents[0].Status)

	rul := &RULPrediction{DeviceID: "dev-1", PredictedRUL: 500.0}
	svc.updateFaultEvent("anom-1", nil, nil, rul)
	assert.Equal(t, "rul_predicted", svc.faultEvents[0].Status)
}

func TestGetMetricWeight(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "solar")

	tests := []struct {
		metric string
		weight float64
	}{
		{"temperature", 0.3},
		{"voltage", 0.3},
		{"soc", 0.25},
		{"vibration", 0.15},
		{"unknown", 0.1},
	}

	for _, tt := range tests {
		weight := a.getMetricWeight(tt.metric)
		assert.Equal(t, tt.weight, weight)
	}
}

func TestCalculateOverallHealth_Empty(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "solar")
	score := a.calculateOverallHealth(map[string]float64{})
	assert.Equal(t, 0.0, score)
}

func TestAssess_NoMatchingDevice(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "solar")
	data := []*TimeSeriesData{
		{DeviceID: "dev-2", Metric: "temperature", Value: 30, Timestamp: time.Now()},
	}
	assessment, err := a.Assess(context.Background(), data)
	require.NoError(t, err)
	assert.Equal(t, 0.0, assessment.HealthScore)
}

func TestPredictRUL_NoMatchingDevice(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "solar")
	data := []*TimeSeriesData{
		{DeviceID: "dev-2", Metric: "temperature", Value: 30, Timestamp: time.Now()},
	}
	rul, err := a.PredictRUL(context.Background(), data)
	require.NoError(t, err)
	assert.Equal(t, 0.0, rul.PredictedRUL)
}

func TestClassify_NoMatchingMetric(t *testing.T) {
	classifier := NewRuleBasedClassifier("solar")
	anomaly := &Anomaly{
		ID: "anom-1", DeviceID: "dev-1", Metric: "unknown_metric",
		Value: 50, Severity: SeverityHigh, Timestamp: time.Now(),
	}
	classification, err := classifier.Classify(context.Background(), anomaly)
	require.NoError(t, err)
	assert.Equal(t, "unknown", classification.FaultType)
	assert.Equal(t, "UNKNOWN-001", classification.FaultCode)
}

func TestClassify_NoMatchingRule(t *testing.T) {
	classifier := NewRuleBasedClassifier("solar")
	anomaly := &Anomaly{
		ID: "anom-1", DeviceID: "dev-1", Metric: "temperature",
		Value: 40, Severity: SeverityLow, Timestamp: time.Now(),
	}
	classification, err := classifier.Classify(context.Background(), anomaly)
	require.NoError(t, err)
	assert.Equal(t, "unclassified", classification.FaultType)
	assert.Equal(t, "UNCLASSIFIED-001", classification.FaultCode)
}

func TestHealthHistoryLimit(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "solar")
	data := []*TimeSeriesData{
		{DeviceID: "dev-1", Metric: "temperature", Value: 30, Timestamp: time.Now()},
	}

	for i := 0; i < 110; i++ {
		_, err := a.Assess(context.Background(), data)
		require.NoError(t, err)
	}
	assert.LessOrEqual(t, len(a.healthHistory), 100)
}

func TestRULHistoryLimit(t *testing.T) {
	a := NewSimpleHealthAssessor("dev-1", "solar")
	data := []*TimeSeriesData{
		{DeviceID: "dev-1", Metric: "temperature", Value: 30, Timestamp: time.Now()},
	}

	for i := 0; i < 110; i++ {
		_, err := a.PredictRUL(context.Background(), data)
		require.NoError(t, err)
	}
	assert.LessOrEqual(t, len(a.rulHistory), 100)
}

func TestFaultServiceImpl_PredictRUL_Error(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterAssessor("dev-1", &mockHealthAssessor{err: assert.AnError})
	_, err := svc.PredictRUL(context.Background(), "dev-1")
	assert.Error(t, err)
}

func TestStatisticalDetector_Train_WithData(t *testing.T) {
	detector := NewStatisticalDetector("dev-1", "temp", 3, 2.0)
	data := []*TimeSeriesData{
		{DeviceID: "dev-1", Metric: "temp", Value: 25, Timestamp: time.Now()},
		{DeviceID: "dev-1", Metric: "temp", Value: 30, Timestamp: time.Now()},
	}
	err := detector.Train(context.Background(), data)
	require.NoError(t, err)
	info := detector.GetDetectorInfo()
	assert.NotNil(t, info.TrainedAt)
}
