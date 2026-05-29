package fault

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRULPredictor(t *testing.T) {
	predictor := NewRULPredictor()
	assert.NotNil(t, predictor)
}

func TestRULPredictor_Predict(t *testing.T) {
	predictor := NewRULPredictor()
	prediction, err := predictor.Predict(context.Background(), "device-1")
	require.NoError(t, err)
	assert.Equal(t, "device-1", prediction.DeviceID)
	assert.Equal(t, float64(2160), prediction.PredictedRUL)
	assert.Equal(t, 0.82, prediction.Confidence)
	assert.Equal(t, "declining", prediction.HealthTrend)
	assert.False(t, prediction.Timestamp.IsZero())
}

func TestCalculateWienerRUL_Normal(t *testing.T) {
	rul := CalculateWienerRUL(100.0, 20.0, 5.0)
	assert.InDelta(t, 16.0, rul, 0.001)
}

func TestCalculateWienerRUL_ZeroDriftRate(t *testing.T) {
	rul := CalculateWienerRUL(100.0, 20.0, 0.0)
	assert.True(t, rul > 0)
}

func TestCalculateWienerRUL_NegativeDriftRate(t *testing.T) {
	rul := CalculateWienerRUL(100.0, 20.0, -1.0)
	assert.True(t, rul > 0)
}

func TestCalculateWienerRUL_AlreadyFailed(t *testing.T) {
	rul := CalculateWienerRUL(10.0, 20.0, 5.0)
	assert.Equal(t, 0.0, rul)
}

func TestCalculateWienerRUL_AtThreshold(t *testing.T) {
	rul := CalculateWienerRUL(20.0, 20.0, 5.0)
	assert.Equal(t, 0.0, rul)
}

func TestCalculateRULConfidence_Normal(t *testing.T) {
	confidence := CalculateRULConfidence(1000.0, 0.01)
	assert.Greater(t, confidence, 0.0)
	assert.LessOrEqual(t, confidence, 1.0)
}

func TestCalculateRULConfidence_ZeroRUL(t *testing.T) {
	confidence := CalculateRULConfidence(0.0, 0.01)
	assert.Equal(t, 0.0, confidence)
}

func TestCalculateRULConfidence_NegativeRUL(t *testing.T) {
	confidence := CalculateRULConfidence(-100.0, 0.01)
	assert.Equal(t, 0.0, confidence)
}

func TestCalculateRULConfidence_HighDiffusion(t *testing.T) {
	confidence := CalculateRULConfidence(100.0, 10.0)
	assert.GreaterOrEqual(t, confidence, 0.0)
}

func TestDetermineMaintenanceWindow_Urgent(t *testing.T) {
	result := DetermineMaintenanceWindow(500.0)
	assert.Contains(t, result, "7")
}

func TestDetermineMaintenanceWindow_Preventive(t *testing.T) {
	result := DetermineMaintenanceWindow(1500.0)
	assert.Contains(t, result, "30")
}

func TestDetermineMaintenanceWindow_Planned(t *testing.T) {
	result := DetermineMaintenanceWindow(3000.0)
	assert.Contains(t, result, "90")
}

func TestDetermineMaintenanceWindow_Good(t *testing.T) {
	result := DetermineMaintenanceWindow(5000.0)
	assert.Contains(t, result, "良好")
}

func TestDetermineMaintenanceWindow_Boundary7(t *testing.T) {
	result := DetermineMaintenanceWindow(720.0)
	assert.Contains(t, result, "7")
}

func TestDetermineMaintenanceWindow_Boundary30(t *testing.T) {
	result := DetermineMaintenanceWindow(2160.0)
	assert.Contains(t, result, "30")
}

func TestDetermineMaintenanceWindow_Boundary90(t *testing.T) {
	result := DetermineMaintenanceWindow(4320.0)
	assert.Contains(t, result, "90")
}

func TestFaultServiceImpl_ClassifyFaults_MultipleAnomalies(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterClassifier("rule", &mockFaultClassifier{})

	anomalies := []*Anomaly{
		{ID: "a1", DeviceID: "dev-1", Metric: "temp"},
		{ID: "a2", DeviceID: "dev-2", Metric: "vibration"},
	}
	classifications, err := svc.ClassifyFaults(context.Background(), anomalies)
	require.NoError(t, err)
	assert.Len(t, classifications, 2)
}

func TestFaultServiceImpl_GetServiceInfo_WithEvents(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterDetector("d1", &mockFaultDetector{
		anomalies: []*Anomaly{{ID: "a1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}},
	})
	svc.DetectAnomalies(context.Background(), []*TimeSeriesData{})

	info := svc.GetServiceInfo()
	assert.Equal(t, 1, info["events_count"])
}

func TestFaultServiceImpl_CreateFaultEvent_WithClassification(t *testing.T) {
	svc := NewFaultService()
	anomaly := &Anomaly{ID: "a1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}
	classification := &FaultClassification{ID: "cls-1", FaultType: "electrical"}
	assessment := &HealthAssessment{DeviceID: "dev-1", HealthScore: 80.0}
	rul := &RULPrediction{DeviceID: "dev-1", PredictedRUL: 500.0}

	svc.createFaultEvent(anomaly, classification, assessment, rul)
	assert.Len(t, svc.faultEvents, 1)
	assert.Equal(t, "active", svc.faultEvents[0].Status)
	assert.NotNil(t, svc.faultEvents[0].Classification)
	assert.NotNil(t, svc.faultEvents[0].Assessment)
	assert.NotNil(t, svc.faultEvents[0].RUL)
}

func TestFaultServiceImpl_UpdateFaultEvent_WithAssessment(t *testing.T) {
	svc := NewFaultService()
	anomaly := &Anomaly{ID: "a1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}
	svc.createFaultEvent(anomaly, nil, nil, nil)

	assessment := &HealthAssessment{DeviceID: "dev-1", HealthScore: 80.0}
	svc.updateFaultEvent("a1", nil, assessment, nil)
	assert.Equal(t, "assessed", svc.faultEvents[0].Status)
	assert.NotNil(t, svc.faultEvents[0].Assessment)
}

func TestFaultServiceImpl_UpdateFaultEvent_WithRUL(t *testing.T) {
	svc := NewFaultService()
	anomaly := &Anomaly{ID: "a1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}
	svc.createFaultEvent(anomaly, nil, nil, nil)

	rul := &RULPrediction{DeviceID: "dev-1", PredictedRUL: 500.0}
	svc.updateFaultEvent("a1", nil, nil, rul)
	assert.Equal(t, "rul_predicted", svc.faultEvents[0].Status)
	assert.NotNil(t, svc.faultEvents[0].RUL)
}

func TestFaultServiceImpl_UpdateFaultEvent_AllFields(t *testing.T) {
	svc := NewFaultService()
	anomaly := &Anomaly{ID: "a1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}
	svc.createFaultEvent(anomaly, nil, nil, nil)

	classification := &FaultClassification{ID: "cls-1", FaultType: "electrical"}
	assessment := &HealthAssessment{DeviceID: "dev-1", HealthScore: 80.0}
	rul := &RULPrediction{DeviceID: "dev-1", PredictedRUL: 500.0}
	svc.updateFaultEvent("a1", classification, assessment, rul)
	assert.Equal(t, "rul_predicted", svc.faultEvents[0].Status)
}

func TestFaultServiceImpl_UpdateFaultEvent_NotFound(t *testing.T) {
	svc := NewFaultService()
	anomaly := &Anomaly{ID: "a1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}
	svc.createFaultEvent(anomaly, nil, nil, nil)

	svc.updateFaultEvent("nonexistent", &FaultClassification{ID: "cls-1"}, nil, nil)
	assert.Len(t, svc.faultEvents, 1)
	assert.Equal(t, "active", svc.faultEvents[0].Status)
}

func TestFaultServiceImpl_PredictRUL_Error(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterAssessor("dev-1", &mockHealthAssessor{err: assert.AnError})
	_, err := svc.PredictRUL(context.Background(), "dev-1")
	assert.Error(t, err)
}

func TestFaultServiceImpl_AssessHealth_ErrorPath(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterAssessor("dev-1", &mockHealthAssessor{err: assert.AnError})
	_, err := svc.AssessHealth(context.Background(), "dev-1")
	assert.Error(t, err)
}

func TestFaultServiceImpl_GetFaultEvents_EmptyTimeRange(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterDetector("d1", &mockFaultDetector{
		anomalies: []*Anomaly{{ID: "a1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}},
	})
	svc.DetectAnomalies(context.Background(), []*TimeSeriesData{})

	events, err := svc.GetFaultEvents(context.Background(), "dev-1", time.Now().Add(100*time.Hour), time.Now().Add(200*time.Hour))
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestFaultServiceImpl_DetectAnomalies_SingleDetector(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterDetector("d1", &mockFaultDetector{
		anomalies: []*Anomaly{{ID: "a1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}},
	})
	anomalies, err := svc.DetectAnomalies(context.Background(), []*TimeSeriesData{})
	require.NoError(t, err)
	assert.Len(t, anomalies, 1)
}

func TestFaultServiceImpl_DeduplicateAnomalies_DifferentMetrics(t *testing.T) {
	svc := NewFaultService()
	ts := time.Now()
	anomalies := []*Anomaly{
		{ID: "a1", DeviceID: "dev-1", Metric: "temp", Timestamp: ts},
		{ID: "a2", DeviceID: "dev-1", Metric: "vibration", Timestamp: ts},
	}
	unique := svc.deduplicateAnomalies(anomalies)
	assert.Len(t, unique, 2)
}

func TestFaultServiceImpl_DeduplicateAnomalies_AllUnique(t *testing.T) {
	svc := NewFaultService()
	anomalies := []*Anomaly{
		{ID: "a1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()},
		{ID: "a2", DeviceID: "dev-2", Metric: "temp", Timestamp: time.Now()},
	}
	unique := svc.deduplicateAnomalies(anomalies)
	assert.Len(t, unique, 2)
}

func TestFaultServiceImpl_DeduplicateAnomalies_Empty(t *testing.T) {
	svc := NewFaultService()
	unique := svc.deduplicateAnomalies([]*Anomaly{})
	assert.Empty(t, unique)
}

func TestFaultEvent_AdditionalInfo(t *testing.T) {
	event := &FaultEvent{
		ID:             "event-1",
		DeviceID:       "dev-1",
		Anomaly:        &Anomaly{ID: "a1"},
		Classification: &FaultClassification{ID: "cls-1"},
		Assessment:     &HealthAssessment{DeviceID: "dev-1"},
		RUL:            &RULPrediction{DeviceID: "dev-1"},
		Status:         "active",
		Actions:        []string{"待处理"},
		AdditionalInfo: map[string]interface{}{"key": "value"},
	}
	assert.Equal(t, "event-1", event.ID)
	assert.NotNil(t, event.Classification)
	assert.NotNil(t, event.Assessment)
	assert.NotNil(t, event.RUL)
	assert.Equal(t, []string{"待处理"}, event.Actions)
	assert.Equal(t, "value", event.AdditionalInfo["key"])
}

func TestDetectorInfo_Struct_Full(t *testing.T) {
	info := &DetectorInfo{
		DetectorID:   "d1",
		DetectorType: "threshold",
		Version:      "2.0",
		Status:       "inactive",
		Parameters:   map[string]interface{}{"threshold": 100.0, "window": 60},
	}
	assert.Equal(t, "2.0", info.Version)
	assert.Equal(t, "inactive", info.Status)
}

func TestClassifierInfo_Struct_Full(t *testing.T) {
	info := &ClassifierInfo{
		ClassifierID:   "c1",
		ClassifierType: "ml",
		Version:        "2.0",
		Status:         "inactive",
	}
	assert.Equal(t, "ml", info.ClassifierType)
}

func TestAssessorInfo_Struct_Full(t *testing.T) {
	info := &AssessorInfo{
		AssessorID:   "a1",
		AssessorType: "ml",
		Version:      "2.0",
		Status:       "inactive",
	}
	assert.Equal(t, "ml", info.AssessorType)
}

func TestFaultClassification_Struct_Full(t *testing.T) {
	cls := &FaultClassification{
		ID:              "cls-1",
		AnomalyID:       "a1",
		FaultType:       "electrical",
		FaultCode:       "E001",
		Description:     "overvoltage fault",
		Confidence:      0.95,
		Recommendations: []string{"check voltage", "replace fuse"},
	}
	assert.Equal(t, "E001", cls.FaultCode)
	assert.Len(t, cls.Recommendations, 2)
}

func TestHealthAssessment_Struct_Full(t *testing.T) {
	ha := &HealthAssessment{
		DeviceID:       "dev-1",
		HealthScore:    75.0,
		HealthStatus:   HealthStatusFair,
		ComponentHealth: map[string]float64{"inverter": 80.0, "panel": 70.0},
		Issues:         []string{"high temperature", "low efficiency"},
		Confidence:     0.9,
	}
	assert.Len(t, ha.ComponentHealth, 2)
	assert.Len(t, ha.Issues, 2)
}

func TestRULPrediction_Struct_Full(t *testing.T) {
	pred := &RULPrediction{
		DeviceID:     "dev-1",
		PredictedRUL: 720.0,
		Confidence:   0.85,
		RULInterval:  [2]float64{600, 840},
		HealthTrend:  "declining",
		Timestamp:    time.Now(),
	}
	assert.Equal(t, [2]float64{600, 840}, pred.RULInterval)
	assert.False(t, pred.Timestamp.IsZero())
}

func TestTimeSeriesData_Struct_Full(t *testing.T) {
	data := &TimeSeriesData{
		DeviceID:  "dev-1",
		Metric:    "power",
		Value:     100.0,
		Timestamp: time.Now(),
		Features:  map[string]float64{"irradiance": 800.0, "temp": 25.0},
	}
	assert.Len(t, data.Features, 2)
}

func TestFaultLabeledData_Struct_Full(t *testing.T) {
	data := &FaultLabeledData{
		Anomaly:   &Anomaly{ID: "a1"},
		FaultType: "electrical",
		FaultCode: "E001",
		Label:     true,
		Features:  map[string]float64{"temperature": 90.0, "vibration": 5.0},
		Timestamp: time.Now(),
	}
	assert.False(t, data.Timestamp.IsZero())
	assert.Len(t, data.Features, 2)
}
