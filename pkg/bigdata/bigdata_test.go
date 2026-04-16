package bigdata

import (
	"testing"
	"time"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/new-energy-monitoring/pkg/bigdata/storage"
	"github.com/new-energy-monitoring/pkg/bigdata/types"
	"github.com/new-energy-monitoring/pkg/bigdata/visualization"
)

func TestBigDataService(t *testing.T) {
	// 创建大数据服务
	service := NewBigDataService()

	// 配置
	storageConfig := types.StorageConfig{
		Type:     "clickhouse",
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Table:    "test_data",
		Username: "default",
		Password: "",
		Options:  map[string]interface{}{"secure": false},
	}

	analysisConfig := types.AnalysisConfig{
		Type:    "basic",
		Master:  "local",
		AppName: "test-analysis",
		Executor: 1,
		Memory:  "1g",
	}

	visualizationConfig := types.VisualizationConfig{
		Type:   "basic",
		Host:   "localhost",
		Port:   3000,
		APIKey: "test-key",
	}

	processingConfig := types.ProcessingConfig{
		Type:        "basic",
		WindowSize:  "10s",
		SlideSize:   "5s",
		Parallelism: 1,
	}

	ingestionConfig := types.IngestionConfig{
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
	// 先创建仪表板
	if vis, ok := service.visualization.(*visualization.BasicVisualizer); ok {
		err = vis.CreateDashboard("test-dashboard", []types.Panel{
			{
				ID:    "test-panel",
				Title: "Test Panel",
				Type:  "graph",
			},
		})
		if err != nil {
			t.Fatalf("Failed to create dashboard: %v", err)
		}
	}

	// 然后更新面板
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

func BenchmarkBigDataServiceProcess(b *testing.B) {
	service := NewBigDataService()
	
	storageConfig := types.StorageConfig{
		Type:     "clickhouse",
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Table:    "test_data",
		Username: "default",
		Password: "",
	}
	
	analysisConfig := types.AnalysisConfig{
		Type:    "basic",
		Master:  "local",
		AppName: "benchmark-analysis",
		Executor: 1,
		Memory:  "1g",
	}
	
	visualizationConfig := types.VisualizationConfig{
		Type:   "basic",
		Host:   "localhost",
		Port:   3000,
	}
	
	processingConfig := types.ProcessingConfig{
		Type:        "basic",
		WindowSize:  "10s",
		SlideSize:   "5s",
		Parallelism: 1,
	}
	
	ingestionConfig := types.IngestionConfig{
		Type:       "basic",
		Topic:      "test-topic",
		Broker:     "localhost:9092",
		ConsumerID: "test-consumer",
		BatchSize:  100,
	}
	
	if err := service.Init(storageConfig, analysisConfig, visualizationConfig, processingConfig, ingestionConfig); err != nil {
		b.Fatalf("Failed to init service: %v", err)
	}
	defer service.Close()
	
	testData := createTestBatchData(100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.Process(testData)
		if err != nil {
			b.Errorf("Process failed: %v", err)
		}
	}
}

func BenchmarkClickHouseWrite(b *testing.B) {
	config := types.StorageConfig{
		Type:     "clickhouse",
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Table:    "test_data",
		Username: "default",
		Password: "",
		BatchSize: 1000,
	}
	
	store := storage.NewClickHouseStorage()
	if err := store.Init(config); err != nil {
		b.Fatalf("Failed to init storage: %v", err)
	}
	defer store.Close()
	
	testData := createTestBatchData(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := store.Write(testData); err != nil {
			b.Errorf("Write failed: %v", err)
		}
	}
}

func BenchmarkClickHouseQuery(b *testing.B) {
	config := types.StorageConfig{
		Type:     "clickhouse",
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Table:    "test_data",
		Username: "default",
		Password: "",
	}
	
	store := storage.NewClickHouseStorage()
	if err := store.Init(config); err != nil {
		b.Fatalf("Failed to init storage: %v", err)
	}
	defer store.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.Query("SELECT * FROM test_data LIMIT 1000")
		if err != nil {
			b.Errorf("Query failed: %v", err)
		}
	}
}

func BenchmarkDorisWrite(b *testing.B) {
	config := types.StorageConfig{
		Type:     "doris",
		Host:     "localhost",
		Port:     9030,
		Database: "test",
		Table:    "test_data",
		Username: "root",
		Password: "",
		BatchSize: 1000,
	}
	
	store := storage.NewDorisStorage()
	if err := store.Init(config); err != nil {
		b.Fatalf("Failed to init storage: %v", err)
	}
	defer store.Close()
	
	testData := createTestBatchData(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := store.Write(testData); err != nil {
			b.Errorf("Write failed: %v", err)
		}
	}
}

func BenchmarkDorisQuery(b *testing.B) {
	config := types.StorageConfig{
		Type:     "doris",
		Host:     "localhost",
		Port:     9030,
		Database: "test",
		Table:    "test_data",
		Username: "root",
		Password: "",
	}
	
	store := storage.NewDorisStorage()
	if err := store.Init(config); err != nil {
		b.Fatalf("Failed to init storage: %v", err)
	}
	defer store.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.Query("SELECT * FROM test_data LIMIT 1000")
		if err != nil {
			b.Errorf("Query failed: %v", err)
		}
	}
}

func BenchmarkDorisMultiDimension(b *testing.B) {
	config := types.StorageConfig{
		Type:     "doris",
		Host:     "localhost",
		Port:     9030,
		Database: "test",
		Table:    "test_data",
		Username: "root",
		Password: "",
	}
	
	store := storage.NewDorisStorage()
	if err := store.Init(config); err != nil {
		b.Fatalf("Failed to init storage: %v", err)
	}
	defer store.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.MultiDimensionAggregation(
			[]string{"metric_value"},
			[]string{"station_id", "device_id"},
			time.Now().Add(-24*time.Hour),
			time.Now(),
			map[string]interface{}{},
		)
		if err != nil {
			b.Errorf("MultiDimensionAggregation failed: %v", err)
		}
	}
}

func createTestData() *types.BatchData {
	dataPoints := []*types.DataPoint{
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

	return &types.BatchData{
		DataPoints: dataPoints,
		Metadata: types.Metadata{
			Source:      "test",
			BatchID:     "test-batch-1",
			Timestamp:   time.Now(),
			RecordCount: len(dataPoints),
			Properties:  map[string]interface{}{"test": true},
		},
	}
}

func createTestBatchData(count int) *types.BatchData {
	dataPoints := make([]*types.DataPoint, 0, count)
	for i := 0; i < count; i++ {
		dataPoints = append(dataPoints, &types.DataPoint{
			Timestamp:  time.Now(),
			DeviceID:   fmt.Sprintf("DEV-%03d", i%100),
			Metric:     "power",
			Value:      float64(i * 100),
			Tags:       map[string]string{"location": fmt.Sprintf("site-%d", i%5), "station": fmt.Sprintf("STN-%03d", i%10)},
			Attributes: map[string]interface{}{"source": "benchmark"},
		})
	}

	return &types.BatchData{
		DataPoints: dataPoints,
		Metadata: types.Metadata{
			Source:      "benchmark",
			BatchID:     fmt.Sprintf("batch-%d", time.Now().UnixNano()),
			Timestamp:   time.Now(),
			RecordCount: len(dataPoints),
		},
	}
}

func BenchmarkBigDataStressConcurrentWrite(b *testing.B) {
	config := types.StorageConfig{
		Type:     "doris",
		Host:     "localhost",
		Port:     9030,
		Database: "test",
		Table:    "test_data",
		Username: "root",
		Password: "",
		BatchSize: 1000,
	}
	
	store := storage.NewDorisStorage()
	if err := store.Init(config); err != nil {
		b.Fatalf("Failed to init storage: %v", err)
	}
	defer store.Close()
	
	concurrencyLevels := []int{10, 20, 50, 100}
	
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrent_%d", concurrency), func(b *testing.B) {
			var successCount int64
			var failCount int64
			var wg sync.WaitGroup
			
			b.ResetTimer()
			
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					testData := createTestBatchData(100)
					
					for j := 0; j < b.N; j++ {
						err := store.Write(testData)
						if err != nil {
							atomic.AddInt64(&failCount, 1)
						} else {
							atomic.AddInt64(&successCount, 1)
						}
					}
				}(i)
			}
			
			wg.Wait()
			
			b.ReportMetric(float64(successCount), "success_count")
			b.ReportMetric(float64(failCount), "fail_count")
			if successCount > 0 {
				b.ReportMetric(float64(successCount+failCount)/float64(b.N*concurrency)*100, "completion_rate")
			}
		})
	}
}

func BenchmarkBigDataStressMixedLoad(b *testing.B) {
	config := types.StorageConfig{
		Type:     "doris",
		Host:     "localhost",
		Port:     9030,
		Database: "test",
		Table:    "test_data",
		Username: "root",
		Password: "",
	}
	
	store := storage.NewDorisStorage()
	if err := store.Init(config); err != nil {
		b.Fatalf("Failed to init storage: %v", err)
	}
	defer store.Close()
	
	testData := createTestBatchData(100)
	
	concurrencyLevels := []int{10, 20, 50}
	
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Mixed_%d", concurrency), func(b *testing.B) {
			var successCount int64
			var failCount int64
			var wg sync.WaitGroup
			
			b.ResetTimer()
			
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					
					for j := 0; j < b.N; j++ {
						var err error
						switch idx % 3 {
						case 0:
							err = store.Write(testData)
						case 1:
							_, err = store.Query("SELECT * FROM test_data LIMIT 100")
						case 2:
							_, err = store.MultiDimensionAggregation(
								[]string{"metric_value"},
								[]string{"station_id", "device_id"},
								time.Now().Add(-24*time.Hour),
								time.Now(),
								map[string]interface{}{},
							)
						}
						
						if err != nil {
							atomic.AddInt64(&failCount, 1)
						} else {
							atomic.AddInt64(&successCount, 1)
						}
					}
				}(i)
			}
			
			wg.Wait()
			
			b.ReportMetric(float64(successCount), "success_count")
			b.ReportMetric(float64(failCount), "fail_count")
			if successCount > 0 {
				b.ReportMetric(float64(successCount+failCount)/float64(b.N*concurrency)*100, "completion_rate")
			}
		})
	}
}

func BenchmarkBigDataStressSustainedLoad(b *testing.B) {
	config := types.StorageConfig{
		Type:     "doris",
		Host:     "localhost",
		Port:     9030,
		Database: "test",
		Table:    "test_data",
		Username: "root",
		Password: "",
		BatchSize: 1000,
	}
	
	store := storage.NewDorisStorage()
	if err := store.Init(config); err != nil {
		b.Fatalf("Failed to init storage: %v", err)
	}
	defer store.Close()
	
	concurrency := 20
	durationSeconds := 10
	
	b.Run(fmt.Sprintf("Sustained_%ds", durationSeconds), func(b *testing.B) {
		var successCount int64
		var failCount int64
		var wg sync.WaitGroup
		var stopFlag int32
		
		startTime := time.Now()
		
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				testData := createTestBatchData(100)
				
				for atomic.LoadInt32(&stopFlag) == 0 {
					err := store.Write(testData)
					if err != nil {
						atomic.AddInt64(&failCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}
					
					time.Sleep(10 * time.Millisecond)
				}
			}(i)
		}
		
		time.Sleep(time.Duration(durationSeconds) * time.Second)
		atomic.StoreInt32(&stopFlag, 1)
		
		wg.Wait()
		elapsedTime := time.Since(startTime)
		
		b.ReportMetric(float64(successCount), "success_count")
		b.ReportMetric(float64(failCount), "fail_count")
		b.ReportMetric(elapsedTime.Seconds(), "duration_seconds")
		if successCount > 0 {
			b.ReportMetric(float64(successCount)/elapsedTime.Seconds(), "ops_per_second")
		}
	})
}
