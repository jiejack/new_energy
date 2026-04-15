package bigdata

import (
	"testing"
	"time"
)

func TestBigDataService(t *testing.T) {
	// 创建大数据服务
	service := NewBigDataService()

	// 配置
	storageConfig := StorageConfig{
		Type:     "clickhouse",
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Table:    "test_data",
		Username: "default",
		Password: "",
		Options:  map[string]interface{}{"secure": false},
	}

	analysisConfig := AnalysisConfig{
		Type:    "basic",
		Master:  "local",
		AppName: "test-analysis",
		Executor: 1,
		Memory:  "1g",
	}

	visualizationConfig := VisualizationConfig{
		Type:   "basic",
		Host:   "localhost",
		Port:   3000,
		APIKey: "test-key",
	}

	processingConfig := ProcessingConfig{
		Type:        "basic",
		WindowSize:  "10s",
		SlideSize:   "5s",
		Parallelism: 1,
	}

	ingestionConfig := IngestionConfig{
		Type:       "basic",
		Topic:      "test-topic",
		Broker:     "localhost:9092",
		ConsumerID: "test-consumer",
		BatchSize:  100,
	}

	// 初始化服务
	err := service.Init(storageConfig, analysisConfig, visualizationConfig, processingConfig, ingestionConfig)
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// 启动摄取
	err = service.StartIngestion()
	if err != nil {
		t.Fatalf("Failed to start ingestion: %v", err)
	}

	// 创建测试数据
	testData := createTestData()

	// 测试数据摄取
	err = service.Ingest(testData)
	if err != nil {
		t.Fatalf("Failed to ingest data: %v", err)
	}

	// 测试数据处理
	processedData, err := service.Process(testData)
	if err != nil {
		t.Fatalf("Failed to process data: %v", err)
	}

	if len(processedData.DataPoints) != len(testData.DataPoints) {
		t.Errorf("Expected %d processed points, got %d", len(testData.DataPoints), len(processedData.DataPoints))
	}

	// 测试数据分析
	analysisResult, err := service.Analyze("test query")
	if err != nil {
		t.Fatalf("Failed to analyze data: %v", err)
	}

	if analysisResult == nil {
		t.Error("Expected non-nil analysis result")
	}

	// 测试数据可视化
	err = service.Visualize("test-dashboard", "test-panel", map[string]interface{}{"test": "data"})
	if err != nil {
		t.Fatalf("Failed to visualize data: %v", err)
	}

	// 停止摄取
	err = service.StopIngestion()
	if err != nil {
		t.Fatalf("Failed to stop ingestion: %v", err)
	}

	// 关闭服务
	err = service.Close()
	if err != nil {
		t.Fatalf("Failed to close service: %v", err)
	}
}

func createTestData() *BatchData {
	dataPoints := []*DataPoint{
		{
			Timestamp: time.Now(),
			DeviceID:  "SOL-001",
			Metric:    "power",
			Value:     5000,
			Tags:      map[string]string{"location": "roof"},
			Attributes: map[string]interface{}{"model": "panel-1"},
		},
		{
			Timestamp: time.Now().Add(1 * time.Minute),
			DeviceID:  "SOL-001",
			Metric:    "temperature",
			Value:     45,
			Tags:      map[string]string{"location": "roof"},
			Attributes: map[string]interface{}{"model": "panel-1"},
		},
		{
			Timestamp: time.Now().Add(2 * time.Minute),
			DeviceID:  "WND-001",
			Metric:    "power",
			Value:     10000,
			Tags:      map[string]string{"location": "field"},
			Attributes: map[string]interface{}{"model": "turbine-1"},
		},
	}

	return &BatchData{
		DataPoints: dataPoints,
		Metadata: Metadata{
			Source:      "test",
			BatchID:     "test-batch-1",
			Timestamp:   time.Now(),
			RecordCount: len(dataPoints),
			Properties:  map[string]interface{}{"test": true},
		},
	}
}
