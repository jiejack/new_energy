package fault

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockFaultDetector struct {
	anomalies []*Anomaly
	err       error
}

func (m *mockFaultDetector) Detect(ctx context.Context, data []*TimeSeriesData) ([]*Anomaly, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.anomalies, nil
}

func (m *mockFaultDetector) Train(ctx context.Context, data []*TimeSeriesData) error {
	return nil
}

func (m *mockFaultDetector) GetDetectorInfo() *DetectorInfo {
	return &DetectorInfo{DetectorID: "mock", DetectorType: "threshold", Status: "active"}
}

type mockFaultClassifier struct {
	err error
}

func (m *mockFaultClassifier) Classify(ctx context.Context, anomaly *Anomaly) (*FaultClassification, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &FaultClassification{
		ID:        "cls-1",
		AnomalyID: anomaly.ID,
		FaultType: "electrical",
		FaultCode: "E001",
		Confidence: 0.9,
	}, nil
}

func (m *mockFaultClassifier) Train(ctx context.Context, data []*FaultLabeledData) error {
	return nil
}

func (m *mockFaultClassifier) GetClassifierInfo() *ClassifierInfo {
	return &ClassifierInfo{ClassifierID: "mock", ClassifierType: "rule_based", Status: "active"}
}

type mockHealthAssessor struct {
	err error
}

func (m *mockHealthAssessor) Assess(ctx context.Context, data []*TimeSeriesData) (*HealthAssessment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &HealthAssessment{
		DeviceID:     "dev-1",
		HealthScore:  95.0,
		HealthStatus: HealthStatusGood,
		Confidence:   0.9,
	}, nil
}

func (m *mockHealthAssessor) PredictRUL(ctx context.Context, data []*TimeSeriesData) (*RULPrediction, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &RULPrediction{
		DeviceID:     "dev-1",
		PredictedRUL: 720.0,
		Confidence:   0.85,
		HealthTrend:  "stable",
	}, nil
}

func (m *mockHealthAssessor) GetAssessorInfo() *AssessorInfo {
	return &AssessorInfo{AssessorID: "mock", AssessorType: "statistical", Status: "active"}
}

func TestFaultServiceImpl_New(t *testing.T) {
	svc := NewFaultService()
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.detectors)
	assert.NotNil(t, svc.classifiers)
	assert.NotNil(t, svc.assessors)
}

func TestFaultServiceImpl_RegisterDetector(t *testing.T) {
	svc := NewFaultService()
	detector := &mockFaultDetector{}
	svc.RegisterDetector("threshold", detector)
	assert.Contains(t, svc.detectors, "threshold")
}

func TestFaultServiceImpl_RegisterClassifier(t *testing.T) {
	svc := NewFaultService()
	classifier := &mockFaultClassifier{}
	svc.RegisterClassifier("rule_based", classifier)
	assert.Contains(t, svc.classifiers, "rule_based")
}

func TestFaultServiceImpl_RegisterAssessor(t *testing.T) {
	svc := NewFaultService()
	assessor := &mockHealthAssessor{}
	svc.RegisterAssessor("dev-1", assessor)
	assert.Contains(t, svc.assessors, "dev-1")
}

func TestFaultServiceImpl_DetectAnomalies(t *testing.T) {
	svc := NewFaultService()
	detector := &mockFaultDetector{
		anomalies: []*Anomaly{
			{ID: "anom-1", DeviceID: "dev-1", Metric: "temperature", Severity: SeverityHigh, Value: 90.0, ExpectedValue: 80.0},
		},
	}
	svc.RegisterDetector("threshold", detector)

	data := []*TimeSeriesData{
		{DeviceID: "dev-1", Metric: "temperature", Value: 90.0, Timestamp: time.Now()},
	}

	anomalies, err := svc.DetectAnomalies(context.Background(), data)
	require.NoError(t, err)
	assert.Len(t, anomalies, 1)
	assert.Equal(t, "anom-1", anomalies[0].ID)
}

func TestFaultServiceImpl_DetectAnomalies_MultipleDetectors(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterDetector("threshold", &mockFaultDetector{
		anomalies: []*Anomaly{{ID: "anom-1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}},
	})
	svc.RegisterDetector("statistical", &mockFaultDetector{
		anomalies: []*Anomaly{{ID: "anom-2", DeviceID: "dev-1", Metric: "vibration", Timestamp: time.Now()}},
	})

	anomalies, err := svc.DetectAnomalies(context.Background(), []*TimeSeriesData{})
	require.NoError(t, err)
	assert.Len(t, anomalies, 2)
}

func TestFaultServiceImpl_DetectAnomalies_Deduplication(t *testing.T) {
	svc := NewFaultService()
	ts := time.Now()
	svc.RegisterDetector("d1", &mockFaultDetector{
		anomalies: []*Anomaly{{ID: "anom-1", DeviceID: "dev-1", Metric: "temp", Timestamp: ts}},
	})
	svc.RegisterDetector("d2", &mockFaultDetector{
		anomalies: []*Anomaly{{ID: "anom-2", DeviceID: "dev-1", Metric: "temp", Timestamp: ts}},
	})

	anomalies, err := svc.DetectAnomalies(context.Background(), []*TimeSeriesData{})
	require.NoError(t, err)
	assert.Len(t, anomalies, 1)
}

func TestFaultServiceImpl_DetectAnomalies_Error(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterDetector("bad", &mockFaultDetector{err: assert.AnError})

	_, err := svc.DetectAnomalies(context.Background(), []*TimeSeriesData{})
	assert.Error(t, err)
}

func TestFaultServiceImpl_ClassifyFaults(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterClassifier("rule_based", &mockFaultClassifier{})

	anomalies := []*Anomaly{{ID: "anom-1", DeviceID: "dev-1"}}
	classifications, err := svc.ClassifyFaults(context.Background(), anomalies)
	require.NoError(t, err)
	assert.Len(t, classifications, 1)
	assert.Equal(t, "electrical", classifications[0].FaultType)
}

func TestFaultServiceImpl_ClassifyFaults_Error(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterClassifier("bad", &mockFaultClassifier{err: assert.AnError})

	anomalies := []*Anomaly{{ID: "anom-1", DeviceID: "dev-1"}}
	_, err := svc.ClassifyFaults(context.Background(), anomalies)
	assert.Error(t, err)
}

func TestFaultServiceImpl_AssessHealth(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterAssessor("dev-1", &mockHealthAssessor{})

	assessment, err := svc.AssessHealth(context.Background(), "dev-1")
	require.NoError(t, err)
	assert.Equal(t, HealthStatusGood, assessment.HealthStatus)
	assert.Equal(t, 95.0, assessment.HealthScore)
}

func TestFaultServiceImpl_AssessHealth_NoAssessor(t *testing.T) {
	svc := NewFaultService()
	_, err := svc.AssessHealth(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestFaultServiceImpl_AssessHealth_Error(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterAssessor("dev-1", &mockHealthAssessor{err: assert.AnError})
	_, err := svc.AssessHealth(context.Background(), "dev-1")
	assert.Error(t, err)
}

func TestFaultServiceImpl_PredictRUL(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterAssessor("dev-1", &mockHealthAssessor{})

	prediction, err := svc.PredictRUL(context.Background(), "dev-1")
	require.NoError(t, err)
	assert.Equal(t, 720.0, prediction.PredictedRUL)
	assert.Equal(t, 0.85, prediction.Confidence)
	assert.Equal(t, "stable", prediction.HealthTrend)
}

func TestFaultServiceImpl_PredictRUL_NoAssessor(t *testing.T) {
	svc := NewFaultService()
	_, err := svc.PredictRUL(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestFaultServiceImpl_GetFaultEvents(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterDetector("threshold", &mockFaultDetector{
		anomalies: []*Anomaly{{ID: "anom-1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}},
	})
	svc.DetectAnomalies(context.Background(), []*TimeSeriesData{})

	events, err := svc.GetFaultEvents(context.Background(), "dev-1", time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
	require.NoError(t, err)
	assert.NotEmpty(t, events)
}

func TestFaultServiceImpl_GetFaultEvents_NoMatch(t *testing.T) {
	svc := NewFaultService()
	events, err := svc.GetFaultEvents(context.Background(), "dev-1", time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestFaultServiceImpl_GetServiceInfo(t *testing.T) {
	svc := NewFaultService()
	svc.RegisterDetector("threshold", &mockFaultDetector{})
	svc.RegisterClassifier("rule_based", &mockFaultClassifier{})
	svc.RegisterAssessor("dev-1", &mockHealthAssessor{})

	info := svc.GetServiceInfo()
	assert.Equal(t, "FaultService", info["service_name"])
	assert.Equal(t, 1, info["detectors_count"])
	assert.Equal(t, 1, info["classifiers_count"])
	assert.Equal(t, 1, info["assessors_count"])
}

func TestThresholdDetector_Detect(t *testing.T) {
	detector := NewThresholdDetector("dev-1", "temperature", 0.0, 80.0, 1.0, 5)

	data := []*TimeSeriesData{
		{DeviceID: "dev-1", Metric: "temperature", Value: 85.0, Timestamp: time.Now()},
		{DeviceID: "dev-1", Metric: "vibration", Value: 3.0, Timestamp: time.Now()},
	}

	anomalies, err := detector.Detect(context.Background(), data)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(anomalies), 1)
}

func TestThresholdDetector_Detect_NoThreshold(t *testing.T) {
	detector := NewThresholdDetector("dev-1", "temperature", 0.0, 100.0, 1.0, 5)

	data := []*TimeSeriesData{
		{DeviceID: "dev-1", Metric: "temperature", Value: 50.0, Timestamp: time.Now()},
	}

	anomalies, err := detector.Detect(context.Background(), data)
	require.NoError(t, err)
	assert.Empty(t, anomalies)
}

func TestThresholdDetector_Train(t *testing.T) {
	detector := NewThresholdDetector("dev-1", "temperature", 0.0, 100.0, 1.0, 5)
	err := detector.Train(context.Background(), []*TimeSeriesData{})
	assert.NoError(t, err)
}

func TestThresholdDetector_GetDetectorInfo(t *testing.T) {
	detector := NewThresholdDetector("dev-1", "temperature", 0.0, 100.0, 1.0, 5)
	info := detector.GetDetectorInfo()
	assert.Contains(t, info.DetectorID, "dev-1")
	assert.Equal(t, "ThresholdDetector", info.DetectorType)
}

func TestStatisticalDetector_Detect(t *testing.T) {
	detector := NewStatisticalDetector("dev-1", "temp", 3, 2.0)

	data := []*TimeSeriesData{
		{DeviceID: "dev-1", Metric: "temp", Value: 10, Timestamp: time.Now()},
		{DeviceID: "dev-1", Metric: "temp", Value: 10, Timestamp: time.Now().Add(1 * time.Second)},
		{DeviceID: "dev-1", Metric: "temp", Value: 10, Timestamp: time.Now().Add(2 * time.Second)},
		{DeviceID: "dev-1", Metric: "temp", Value: 100, Timestamp: time.Now().Add(3 * time.Second)},
	}

	_, err := detector.Detect(context.Background(), data)
	require.NoError(t, err)
}

func TestStatisticalDetector_Detect_NoAnomalies(t *testing.T) {
	detector := NewStatisticalDetector("dev-1", "temp", 3, 5.0)

	data := []*TimeSeriesData{
		{DeviceID: "dev-1", Metric: "temp", Value: 10, Timestamp: time.Now()},
		{DeviceID: "dev-1", Metric: "temp", Value: 11, Timestamp: time.Now().Add(1 * time.Second)},
		{DeviceID: "dev-1", Metric: "temp", Value: 9, Timestamp: time.Now().Add(2 * time.Second)},
	}

	anomalies, err := detector.Detect(context.Background(), data)
	require.NoError(t, err)
	assert.Empty(t, anomalies)
}

func TestStatisticalDetector_Train(t *testing.T) {
	detector := NewStatisticalDetector("dev-1", "temp", 3, 2.0)
	err := detector.Train(context.Background(), []*TimeSeriesData{})
	assert.NoError(t, err)
}

func TestStatisticalDetector_GetDetectorInfo(t *testing.T) {
	detector := NewStatisticalDetector("dev-1", "temp", 3, 2.0)
	info := detector.GetDetectorInfo()
	assert.Contains(t, info.DetectorID, "dev-1")
	assert.Equal(t, "StatisticalDetector", info.DetectorType)
}

func TestAnomalySeverity_Constants(t *testing.T) {
	assert.Equal(t, AnomalySeverity("low"), SeverityLow)
	assert.Equal(t, AnomalySeverity("medium"), SeverityMedium)
	assert.Equal(t, AnomalySeverity("high"), SeverityHigh)
	assert.Equal(t, AnomalySeverity("critical"), SeverityCritical)
}

func TestHealthStatus_Constants(t *testing.T) {
	assert.Equal(t, HealthStatus("excellent"), HealthStatusExcellent)
	assert.Equal(t, HealthStatus("good"), HealthStatusGood)
	assert.Equal(t, HealthStatus("fair"), HealthStatusFair)
	assert.Equal(t, HealthStatus("poor"), HealthStatusPoor)
	assert.Equal(t, HealthStatus("critical"), HealthStatusCritical)
}

func TestAnomaly_Struct(t *testing.T) {
	anomaly := &Anomaly{
		ID:            "anom-1",
		DeviceID:      "dev-1",
		Metric:        "temperature",
		Value:         90.0,
		ExpectedValue: 80.0,
		Deviation:     10.0,
		Severity:      SeverityHigh,
		Confidence:    0.95,
	}
	assert.Equal(t, "anom-1", anomaly.ID)
	assert.Equal(t, SeverityHigh, anomaly.Severity)
}

func TestFaultClassification_Struct(t *testing.T) {
	classification := &FaultClassification{
		ID:            "cls-1",
		AnomalyID:     "anom-1",
		FaultType:     "electrical",
		FaultCode:     "E001",
		Description:   "overvoltage",
		Confidence:    0.95,
		Recommendations: []string{"check voltage"},
	}
	assert.Equal(t, "cls-1", classification.ID)
	assert.Equal(t, 0.95, classification.Confidence)
}

func TestHealthAssessment_Struct(t *testing.T) {
	assessment := &HealthAssessment{
		DeviceID:       "dev-1",
		HealthScore:    75.0,
		HealthStatus:   HealthStatusFair,
		ComponentHealth: map[string]float64{"inverter": 80.0},
		Issues:         []string{"high temperature"},
		Confidence:     0.9,
	}
	assert.Equal(t, "dev-1", assessment.DeviceID)
	assert.Equal(t, HealthStatusFair, assessment.HealthStatus)
}

func TestRULPrediction_Struct(t *testing.T) {
	prediction := &RULPrediction{
		DeviceID:     "dev-1",
		PredictedRUL: 720.0,
		Confidence:   0.85,
		RULInterval:  [2]float64{600, 840},
		HealthTrend:  "declining",
	}
	assert.Equal(t, 720.0, prediction.PredictedRUL)
	assert.Equal(t, "declining", prediction.HealthTrend)
}

func TestFaultEvent_Struct(t *testing.T) {
	event := &FaultEvent{
		ID:       "event-1",
		DeviceID: "dev-1",
		Anomaly:  &Anomaly{ID: "anom-1"},
		Status:   "active",
	}
	assert.Equal(t, "event-1", event.ID)
	assert.Equal(t, "active", event.Status)
}

func TestDetectorInfo_Struct(t *testing.T) {
	info := &DetectorInfo{
		DetectorID:   "threshold-1",
		DetectorType: "threshold",
		Version:      "1.0",
		Status:       "active",
		Parameters:   map[string]interface{}{"threshold": 80.0},
	}
	assert.Equal(t, "threshold-1", info.DetectorID)
}

func TestClassifierInfo_Struct(t *testing.T) {
	info := &ClassifierInfo{
		ClassifierID:   "rule-1",
		ClassifierType: "rule_based",
		Version:        "1.0",
		Status:         "active",
	}
	assert.Equal(t, "rule-1", info.ClassifierID)
}

func TestAssessorInfo_Struct(t *testing.T) {
	info := &AssessorInfo{
		AssessorID:   "assessor-1",
		AssessorType: "statistical",
		Version:      "1.0",
		Status:       "active",
	}
	assert.Equal(t, "assessor-1", info.AssessorID)
}

func TestTimeSeriesData_Struct(t *testing.T) {
	data := &TimeSeriesData{
		Timestamp: time.Now(),
		Value:     100.0,
		DeviceID:  "dev-1",
		Metric:    "power",
		Features:  map[string]float64{"irradiance": 800.0},
	}
	assert.Equal(t, "dev-1", data.DeviceID)
	assert.Equal(t, 100.0, data.Value)
}

func TestFaultLabeledData_Struct(t *testing.T) {
	data := &FaultLabeledData{
		Anomaly:   &Anomaly{ID: "anom-1"},
		FaultType: "electrical",
		FaultCode: "E001",
		Label:     true,
		Features:  map[string]float64{"temperature": 90.0},
	}
	assert.True(t, data.Label)
}

func TestFaultServiceImpl_DetectAnomalies_NoDetectors(t *testing.T) {
	svc := NewFaultService()
	anomalies, err := svc.DetectAnomalies(context.Background(), []*TimeSeriesData{})
	require.NoError(t, err)
	assert.Empty(t, anomalies)
}

func TestFaultServiceImpl_ClassifyFaults_NoClassifiers(t *testing.T) {
	svc := NewFaultService()
	classifications, err := svc.ClassifyFaults(context.Background(), []*Anomaly{{ID: "anom-1"}})
	require.NoError(t, err)
	assert.Empty(t, classifications)
}

func TestFaultServiceImpl_deduplicateAnomalies(t *testing.T) {
	svc := NewFaultService()
	ts := time.Now()
	anomalies := []*Anomaly{
		{ID: "anom-1", DeviceID: "dev-1", Metric: "temp", Timestamp: ts},
		{ID: "anom-2", DeviceID: "dev-1", Metric: "temp", Timestamp: ts},
		{ID: "anom-3", DeviceID: "dev-1", Metric: "vibration", Timestamp: ts},
	}
	unique := svc.deduplicateAnomalies(anomalies)
	assert.Len(t, unique, 2)
}

func TestFaultServiceImpl_createFaultEvent(t *testing.T) {
	svc := NewFaultService()
	anomaly := &Anomaly{ID: "anom-1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}
	svc.createFaultEvent(anomaly, nil, nil, nil)
	assert.Len(t, svc.faultEvents, 1)
	assert.Equal(t, "active", svc.faultEvents[0].Status)
}

func TestFaultServiceImpl_updateFaultEvent(t *testing.T) {
	svc := NewFaultService()
	anomaly := &Anomaly{ID: "anom-1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}
	svc.createFaultEvent(anomaly, nil, nil, nil)

	classification := &FaultClassification{ID: "cls-1", AnomalyID: "anom-1"}
	svc.updateFaultEvent("anom-1", classification, nil, nil)
	assert.Equal(t, "classified", svc.faultEvents[0].Status)
}

func TestFaultServiceImpl_updateFaultEventsWithAssessment(t *testing.T) {
	svc := NewFaultService()
	anomaly := &Anomaly{ID: "anom-1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}
	svc.createFaultEvent(anomaly, nil, nil, nil)

	assessment := &HealthAssessment{DeviceID: "dev-1", HealthScore: 80.0}
	svc.updateFaultEventsWithAssessment("dev-1", assessment)
	assert.Equal(t, "assessed", svc.faultEvents[0].Status)
}

func TestFaultServiceImpl_updateFaultEventsWithRUL(t *testing.T) {
	svc := NewFaultService()
	anomaly := &Anomaly{ID: "anom-1", DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}
	svc.createFaultEvent(anomaly, nil, nil, nil)

	assessment := &HealthAssessment{DeviceID: "dev-1", HealthScore: 80.0}
	svc.updateFaultEventsWithAssessment("dev-1", assessment)

	rul := &RULPrediction{DeviceID: "dev-1", PredictedRUL: 500.0}
	svc.updateFaultEventsWithRUL("dev-1", rul)
	assert.Equal(t, "rul_predicted", svc.faultEvents[0].Status)
}

func TestFaultServiceImpl_MaxEvents(t *testing.T) {
	svc := NewFaultService()
	for i := 0; i < 1100; i++ {
		anomaly := &Anomaly{ID: fmt.Sprintf("anom-%d", i), DeviceID: "dev-1", Metric: "temp", Timestamp: time.Now()}
		svc.createFaultEvent(anomaly, nil, nil, nil)
	}
	assert.LessOrEqual(t, len(svc.faultEvents), 1000)
}
