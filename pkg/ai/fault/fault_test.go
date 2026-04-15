package fault

import (
	"context"
	"testing"
	"time"
)

func TestThresholdDetector(t *testing.T) {
	detector := NewThresholdDetector("device-1", "temperature", 0, 50, 5, 10)
	
	// 测试检测功能
	data := []*TimeSeriesData{
		{DeviceID: "device-1", Metric: "temperature", Value: 60, Timestamp: time.Now()},
		{DeviceID: "device-1", Metric: "temperature", Value: 40, Timestamp: time.Now()},
		{DeviceID: "device-1", Metric: "temperature", Value: -10, Timestamp: time.Now()},
	}
	
	ctx := context.Background()
	anomalies, err := detector.Detect(ctx, data)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	
	if len(anomalies) != 2 {
		t.Errorf("Expected 2 anomalies, got %d", len(anomalies))
	}
	
	// 测试训练功能
	trainData := []*TimeSeriesData{
		{DeviceID: "device-1", Metric: "temperature", Value: 25, Timestamp: time.Now()},
		{DeviceID: "device-1", Metric: "temperature", Value: 30, Timestamp: time.Now()},
		{DeviceID: "device-1", Metric: "temperature", Value: 35, Timestamp: time.Now()},
	}
	
	err = detector.Train(ctx, trainData)
	if err != nil {
		t.Fatalf("Train failed: %v", err)
	}
	
	// 测试获取检测器信息
	info := detector.GetDetectorInfo()
	if info.DetectorType != "ThresholdDetector" {
		t.Errorf("Expected detector type ThresholdDetector, got %s", info.DetectorType)
	}
}

func TestStatisticalDetector(t *testing.T) {
	detector := NewStatisticalDetector("device-1", "voltage", 5, 2.0)
	
	// 测试检测功能（数据点不足）
	smallData := []*TimeSeriesData{
		{DeviceID: "device-1", Metric: "voltage", Value: 100, Timestamp: time.Now()},
		{DeviceID: "device-1", Metric: "voltage", Value: 110, Timestamp: time.Now()},
	}
	
	ctx := context.Background()
	anomalies, err := detector.Detect(ctx, smallData)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	
	if len(anomalies) != 0 {
		t.Errorf("Expected 0 anomalies for small data, got %d", len(anomalies))
	}
	
	// 测试训练功能
	trainData := []*TimeSeriesData{
		{DeviceID: "device-1", Metric: "voltage", Value: 220, Timestamp: time.Now()},
		{DeviceID: "device-1", Metric: "voltage", Value: 225, Timestamp: time.Now()},
		{DeviceID: "device-1", Metric: "voltage", Value: 230, Timestamp: time.Now()},
	}
	
	err = detector.Train(ctx, trainData)
	if err != nil {
		t.Fatalf("Train failed: %v", err)
	}
	
	// 测试获取检测器信息
	info := detector.GetDetectorInfo()
	if info.DetectorType != "StatisticalDetector" {
		t.Errorf("Expected detector type StatisticalDetector, got %s", info.DetectorType)
	}
}

func TestRuleBasedClassifier(t *testing.T) {
	classifier := NewRuleBasedClassifier("solar")
	
	// 测试分类功能
	anomaly := &Anomaly{
		DeviceID:  "device-1",
		Metric:    "temperature",
		Value:     65,
		Severity:  SeverityHigh,
		Timestamp: time.Now(),
	}
	
	ctx := context.Background()
	classification, err := classifier.Classify(ctx, anomaly)
	if err != nil {
		t.Fatalf("Classify failed: %v", err)
	}
	
	if classification.FaultType != "overheating" {
		t.Errorf("Expected fault type overheating, got %s", classification.FaultType)
	}
	
	// 测试训练功能
	trainData := []*FaultLabeledData{
		{
			Anomaly:   anomaly,
			FaultType: "overheating",
			FaultCode: "SOL-001",
			Label:     true,
			Timestamp: time.Now(),
		},
	}
	
	err = classifier.Train(ctx, trainData)
	if err != nil {
		t.Fatalf("Train failed: %v", err)
	}
	
	// 测试获取分类器信息
	info := classifier.GetClassifierInfo()
	if info.ClassifierType != "RuleBasedClassifier" {
		t.Errorf("Expected classifier type RuleBasedClassifier, got %s", info.ClassifierType)
	}
}

func TestSimpleHealthAssessor(t *testing.T) {
	assessor := NewSimpleHealthAssessor("device-1", "solar")
	
	// 测试健康评估功能
	data := []*TimeSeriesData{
		{DeviceID: "device-1", Metric: "temperature", Value: 30, Timestamp: time.Now()},
		{DeviceID: "device-1", Metric: "voltage", Value: 500, Timestamp: time.Now()},
	}
	
	ctx := context.Background()
	assessment, err := assessor.Assess(ctx, data)
	if err != nil {
		t.Fatalf("Assess failed: %v", err)
	}
	
	if assessment.HealthScore < 90 {
		t.Errorf("Expected health score >= 90, got %f", assessment.HealthScore)
	}
	
	// 测试RUL预测功能
	rul, err := assessor.PredictRUL(ctx, data)
	if err != nil {
		t.Fatalf("PredictRUL failed: %v", err)
	}
	
	if rul.PredictedRUL <= 0 {
		t.Errorf("Expected positive RUL, got %f", rul.PredictedRUL)
	}
	
	// 测试获取评估器信息
	info := assessor.GetAssessorInfo()
	if info.AssessorType != "SimpleHealthAssessor" {
		t.Errorf("Expected assessor type SimpleHealthAssessor, got %s", info.AssessorType)
	}
}

func TestFaultService(t *testing.T) {
	service := NewFaultService()
	
	// 注册检测器
	detector := NewThresholdDetector("device-1", "temperature", 0, 50, 5, 10)
	service.RegisterDetector("temp-detector", detector)
	
	// 注册分类器
	classifier := NewRuleBasedClassifier("solar")
	service.RegisterClassifier("solar-classifier", classifier)
	
	// 注册评估器
	assessor := NewSimpleHealthAssessor("device-1", "solar")
	service.RegisterAssessor("device-1", assessor)
	
	// 测试异常检测
	data := []*TimeSeriesData{
		{DeviceID: "device-1", Metric: "temperature", Value: 60, Timestamp: time.Now()},
	}
	
	ctx := context.Background()
	anomalies, err := service.DetectAnomalies(ctx, data)
	if err != nil {
		t.Fatalf("DetectAnomalies failed: %v", err)
	}
	
	if len(anomalies) == 0 {
		t.Error("Expected at least 1 anomaly")
	}
	
	// 测试故障分类
	classifications, err := service.ClassifyFaults(ctx, anomalies)
	if err != nil {
		t.Fatalf("ClassifyFaults failed: %v", err)
	}
	
	if len(classifications) == 0 {
		t.Error("Expected at least 1 classification")
	}
	
	// 测试健康评估
	assessment, err := service.AssessHealth(ctx, "device-1")
	if err != nil {
		t.Fatalf("AssessHealth failed: %v", err)
	}
	
	if assessment == nil {
		t.Error("Expected non-nil assessment")
	}
	
	// 测试RUL预测
	rul, err := service.PredictRUL(ctx, "device-1")
	if err != nil {
		t.Fatalf("PredictRUL failed: %v", err)
	}
	
	if rul == nil {
		t.Error("Expected non-nil RUL prediction")
	}
	
	// 测试获取故障事件
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now().Add(24 * time.Hour)
	events, err := service.GetFaultEvents(ctx, "device-1", startTime, endTime)
	if err != nil {
		t.Fatalf("GetFaultEvents failed: %v", err)
	}
	
	if len(events) == 0 {
		t.Error("Expected at least 1 fault event")
	}
}

func TestFaultService_Integration(t *testing.T) {
	// 集成测试：完整的故障检测、分类、评估流程
	service := NewFaultService()
	
	// 注册组件
	service.RegisterDetector("temp-detector", NewThresholdDetector("device-1", "temperature", 0, 50, 5, 10))
	service.RegisterClassifier("solar-classifier", NewRuleBasedClassifier("solar"))
	service.RegisterAssessor("device-1", NewSimpleHealthAssessor("device-1", "solar"))
	
	// 模拟异常数据
	data := []*TimeSeriesData{
		{DeviceID: "device-1", Metric: "temperature", Value: 65, Timestamp: time.Now()},
		{DeviceID: "device-1", Metric: "voltage", Value: 700, Timestamp: time.Now()},
	}
	
	ctx := context.Background()
	
	// 1. 检测异常
	anomalies, err := service.DetectAnomalies(ctx, data)
	if err != nil {
		t.Fatalf("DetectAnomalies failed: %v", err)
	}
	
	// 2. 分类故障
	classifications, err := service.ClassifyFaults(ctx, anomalies)
	if err != nil {
		t.Fatalf("ClassifyFaults failed: %v", err)
	}
	
	// 3. 评估健康
	assessment, err := service.AssessHealth(ctx, "device-1")
	if err != nil {
		t.Fatalf("AssessHealth failed: %v", err)
	}
	
	// 4. 预测RUL
	rul, err := service.PredictRUL(ctx, "device-1")
	if err != nil {
		t.Fatalf("PredictRUL failed: %v", err)
	}
	
	// 验证所有步骤都成功完成
	if len(anomalies) == 0 || len(classifications) == 0 || assessment == nil || rul == nil {
		t.Error("Integration test failed: one or more steps returned nil/empty results")
	}
	
	// 验证服务信息
	info := service.GetServiceInfo()
	if info["detectors_count"].(int) != 1 {
		t.Errorf("Expected 1 detector, got %d", info["detectors_count"])
	}
	if info["classifiers_count"].(int) != 1 {
		t.Errorf("Expected 1 classifier, got %d", info["classifiers_count"])
	}
	if info["assessors_count"].(int) != 1 {
		t.Errorf("Expected 1 assessor, got %d", info["assessors_count"])
	}
}