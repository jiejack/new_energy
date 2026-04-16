package performance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata"
	"github.com/new-energy-monitoring/pkg/bigdata/storage"
	"github.com/new-energy-monitoring/pkg/bigdata/types"
)

func BenchmarkBigDataWrite(b *testing.B) {
	config := types.StorageConfig{
		Type:     "clickhouse",
		Host:     "localhost",
		Port:     9000,
		Database: "test_perf",
		Table:    "test_data_perf",
		Username: "default",
		Password: "",
		BatchSize: 1000,
	}

	store := storage.NewClickHouseStorage()
	if err := store.Init(config); err != nil {
		b.Fatalf("Failed to init storage: %v", err)
	}
	defer store.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testData := createTestBatchData(1000)
		if err := store.Write(testData); err != nil {
			b.Errorf("Write failed: %v", err)
		}
	}
}

func BenchmarkBigDataRead(b *testing.B) {
	config := types.StorageConfig{
		Type:     "clickhouse",
		Host:     "localhost",
		Port:     9000,
		Database: "test_perf",
		Table:    "test_data_perf",
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
		_, err := store.Read("SELECT * FROM test_data_perf LIMIT 1000")
		if err != nil {
			b.Errorf("Read failed: %v", err)
		}
	}
}

func BenchmarkBigDataAggregate(b *testing.B) {
	config := types.StorageConfig{
		Type:     "clickhouse",
		Host:     "localhost",
		Port:     9000,
		Database: "test_perf",
		Table:    "test_data_perf",
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
		_, err := store.Aggregate("sum", "power", time.Now().Add(-24*time.Hour), time.Now(), "station_id")
		if err != nil {
			b.Errorf("Aggregate failed: %v", err)
		}
	}
}

func BenchmarkBigDataMultiDimension(b *testing.B) {
	config := types.StorageConfig{
		Type:     "clickhouse",
		Host:     "localhost",
		Port:     9000,
		Database: "test_perf",
		Table:    "test_data_perf",
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

func BenchmarkFlinkProcess(b *testing.B) {
	service := bigdata.NewBigDataService()

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

	if err := service.Init(storageConfig, analysisConfig, visualizationConfig, processingConfig, ingestionConfig); err != nil {
		b.Fatalf("Failed to init service: %v", err)
	}
	defer service.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testData := createTestBatchData(100)
		_, err := service.Process(testData)
		if err != nil {
			b.Errorf("Process failed: %v", err)
		}
	}
}

func BenchmarkDorisWrite(b *testing.B) {
	config := types.StorageConfig{
		Type:     "doris",
		Host:     "localhost",
		Port:     9030,
		Database: "test_perf",
		Table:    "test_data_perf",
		Username: "root",
		Password: "",
		BatchSize: 1000,
	}

	store := storage.NewDorisStorage()
	if err := store.Init(config); err != nil {
		b.Fatalf("Failed to init storage: %v", err)
	}
	defer store.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testData := createTestBatchData(1000)
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
		Database: "test_perf",
		Table:    "test_data_perf",
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
		_, err := store.Query("SELECT * FROM test_data_perf LIMIT 1000")
		if err != nil {
			b.Errorf("Query failed: %v", err)
		}
	}
}

func BenchmarkDorisCacheHit(b *testing.B) {
	config := types.StorageConfig{
		Type:     "doris",
		Host:     "localhost",
		Port:     9030,
		Database: "test_perf",
		Table:    "test_data_perf",
		Username: "root",
		Password: "",
	}

	store := storage.NewDorisStorage()
	if err := store.Init(config); err != nil {
		b.Fatalf("Failed to init storage: %v", err)
	}
	defer store.Close()

	query := "SELECT * FROM test_data_perf LIMIT 1000"
	for i := 0; i < 10; i++ {
		_, _ = store.Query(query)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.Query(query)
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
		Database: "test_perf",
		Table:    "test_data_perf",
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
