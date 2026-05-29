package bigdata

import (
	"testing"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/analysis"
	"github.com/new-energy-monitoring/pkg/bigdata/ingestion"
	"github.com/new-energy-monitoring/pkg/bigdata/processing"
	"github.com/new-energy-monitoring/pkg/bigdata/storage"
	"github.com/new-energy-monitoring/pkg/bigdata/types"
	"github.com/new-energy-monitoring/pkg/bigdata/visualization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBigDataService_NilComponents(t *testing.T) {
	svc := NewBigDataService()

	err := svc.Ingest(&types.BatchData{})
	assert.Error(t, err)

	err = svc.Store(&types.BatchData{})
	assert.Error(t, err)

	_, err = svc.Analyze("test")
	assert.Error(t, err)

	err = svc.Visualize("d1", "p1", nil)
	assert.Error(t, err)

	_, err = svc.Process(&types.BatchData{})
	assert.Error(t, err)

	err = svc.StartIngestion()
	assert.Error(t, err)

	err = svc.StopIngestion()
	assert.Error(t, err)

	err = svc.WritePoint(&types.DataPoint{})
	assert.Error(t, err)

	_, err = svc.ReadTimeRange(time.Now(), time.Now(), "", "", "")
	assert.Error(t, err)

	_, err = svc.Aggregate("avg", "temp", time.Now(), time.Now(), "1h")
	assert.Error(t, err)

	err = svc.Flush()
	assert.Error(t, err)

	_, err = svc.GetStorageStats()
	assert.Error(t, err)

	err = svc.CreateMaterializedView("mv", "tbl", "SELECT 1")
	assert.Error(t, err)

	_, err = svc.ListMaterializedViews()
	assert.Error(t, err)

	err = svc.DropMaterializedView("mv")
	assert.Error(t, err)

	err = svc.RefreshMaterializedView("mv")
	assert.Error(t, err)

	_, err = svc.ExplainQuery("SELECT 1")
	assert.Error(t, err)

	err = svc.CreatePreAggregationTable("tbl", "1h")
	assert.Error(t, err)

	err = svc.CreatePreAggregationRule(nil)
	assert.Error(t, err)

	_, err = svc.ListPreAggregationRules()
	assert.Error(t, err)

	err = svc.EnablePreAggregationRule("r1")
	assert.Error(t, err)

	err = svc.DisablePreAggregationRule("r1")
	assert.Error(t, err)

	err = svc.DeletePreAggregationRule("r1")
	assert.Error(t, err)

	err = svc.RefreshPreAggregation("tbl")
	assert.Error(t, err)

	_, err = svc.GetCacheStats()
	assert.Error(t, err)

	err = svc.ClearCache()
	assert.Error(t, err)

	_, err = svc.MultiDimensionAggregation(nil, nil, time.Now(), time.Now(), nil)
	assert.Error(t, err)

	_, err = svc.DimensionDrillDown(nil, "", nil, time.Now(), time.Now(), nil)
	assert.Error(t, err)

	_, err = svc.DimensionCrossAnalysis(nil, nil, "", time.Now(), time.Now(), nil)
	assert.Error(t, err)

	_, err = svc.GetDimensionValues("", time.Now(), time.Now(), nil)
	assert.Error(t, err)

	err = svc.Close()
	assert.NoError(t, err)
}

func TestBigDataService_Init_AnalysisInvalid(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(
		types.StorageConfig{Type: "clickhouse"},
		types.AnalysisConfig{Type: "invalid"},
		types.VisualizationConfig{Type: "basic"},
		types.ProcessingConfig{Type: "basic"},
		types.IngestionConfig{Type: "basic"},
	)
	assert.Error(t, err)
}

func TestBigDataService_Init_VisualizationInvalid(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(
		types.StorageConfig{Type: "clickhouse"},
		types.AnalysisConfig{Type: "basic"},
		types.VisualizationConfig{Type: "invalid"},
		types.ProcessingConfig{Type: "basic"},
		types.IngestionConfig{Type: "basic"},
	)
	assert.Error(t, err)
}

func TestBigDataService_Init_ProcessingInvalid(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(
		types.StorageConfig{Type: "clickhouse"},
		types.AnalysisConfig{Type: "basic"},
		types.VisualizationConfig{Type: "basic"},
		types.ProcessingConfig{Type: "invalid"},
		types.IngestionConfig{Type: "basic"},
	)
	assert.Error(t, err)
}

func TestBigDataService_Init_IngestionInvalid(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(
		types.StorageConfig{Type: "clickhouse"},
		types.AnalysisConfig{Type: "basic"},
		types.VisualizationConfig{Type: "basic"},
		types.ProcessingConfig{Type: "basic"},
		types.IngestionConfig{Type: "invalid"},
	)
	assert.Error(t, err)
}

func TestBigDataService_Init_FlinkProcessing(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(
		types.StorageConfig{Type: "clickhouse"},
		types.AnalysisConfig{Type: "basic"},
		types.VisualizationConfig{Type: "basic"},
		types.ProcessingConfig{Type: "flink"},
		types.IngestionConfig{Type: "basic"},
	)
	require.NoError(t, err)
	defer svc.Close()

	stats, err := svc.GetProcessingStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestBigDataService_Init_DorisStorage(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(
		types.StorageConfig{Type: "doris"},
		types.AnalysisConfig{Type: "basic"},
		types.VisualizationConfig{Type: "basic"},
		types.ProcessingConfig{Type: "basic"},
		types.IngestionConfig{Type: "basic"},
	)
	require.NoError(t, err)
	defer svc.Close()

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
		},
	}
	err = svc.Store(bd)
	assert.NoError(t, err)

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 26.0, Timestamp: time.Now()}
	err = svc.WritePoint(dp)
	assert.NoError(t, err)

	err = svc.Flush()
	assert.NoError(t, err)

	_, err = svc.GetStorageStats()
	assert.NoError(t, err)

	err = svc.CreateMaterializedView("mv1", "tbl", "SELECT 1")
	assert.Error(t, err)

	_, err = svc.ListMaterializedViews()
	assert.NoError(t, err)

	err = svc.DropMaterializedView("mv1")
	assert.Error(t, err)

	err = svc.RefreshMaterializedView("mv1")
	assert.Error(t, err)

	_, err = svc.ExplainQuery("SELECT 1")
	assert.NoError(t, err)

	err = svc.CreatePreAggregationTable("tbl", "1h")
	assert.Error(t, err)

	err = svc.CreatePreAggregationRule(nil)
	assert.Error(t, err)

	_, err = svc.ListPreAggregationRules()
	assert.NoError(t, err)

	err = svc.EnablePreAggregationRule("r1")
	assert.Error(t, err)

	err = svc.DisablePreAggregationRule("r1")
	assert.Error(t, err)

	err = svc.DeletePreAggregationRule("r1")
	assert.Error(t, err)

	err = svc.RefreshPreAggregation("tbl")
	assert.Error(t, err)

	_, err = svc.GetCacheStats()
	assert.NoError(t, err)

	err = svc.ClearCache()
	assert.NoError(t, err)

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	_, err = svc.MultiDimensionAggregation([]string{"temp"}, []string{"device_id"}, start, end, nil)
	assert.NoError(t, err)

	_, err = svc.DimensionDrillDown([]string{"device_id"}, "metric", []string{"temp"}, start, end, nil)
	assert.NoError(t, err)

	_, err = svc.DimensionCrossAnalysis([]string{"device_id"}, []string{"metric"}, "temp", start, end, nil)
	assert.NoError(t, err)

	_, err = svc.GetDimensionValues("device_id", start, end, nil)
	assert.NoError(t, err)
}

func TestTypes_Pool_ReleaseNil(t *testing.T) {
	types.ReleaseDataPoint(nil)
	types.ReleaseBatchData(nil)
}

func TestTypes_Pool_AcquireRelease(t *testing.T) {
	dp := types.AcquireDataPoint()
	dp.DeviceID = "dev1"
	dp.Metric = "temp"
	dp.Value = 25.5
	dp.Tags["key"] = "value"
	dp.Attributes["attr"] = "val"
	types.ReleaseDataPoint(dp)

	dp2 := types.AcquireDataPoint()
	assert.Empty(t, dp2.DeviceID)
	assert.Empty(t, dp2.Metric)
	types.ReleaseDataPoint(dp2)

	bd := types.AcquireBatchData()
	bd.Metadata.Source = "test"
	bd.Metadata.BatchID = "b1"
	bd.DataPoints = append(bd.DataPoints, &types.DataPoint{DeviceID: "d1"})
	types.ReleaseBatchData(bd)

	bd2 := types.AcquireBatchData()
	assert.Empty(t, bd2.Metadata.Source)
	assert.Empty(t, bd2.DataPoints)
	types.ReleaseBatchData(bd2)
}

func TestTypes_ErrorCodes(t *testing.T) {
	codes := []string{
		types.ErrCodeInvalidConfig,
		types.ErrCodeStorageError,
		types.ErrCodeAnalysisError,
		types.ErrCodeVisualizationError,
		types.ErrCodeProcessingError,
		types.ErrCodeIngestionError,
	}
	for _, code := range codes {
		assert.NotEmpty(t, code)
	}
}

func TestBasicIngester_InitInvalid(t *testing.T) {
	ing := ingestion.NewBasicIngester()
	err := ing.Init(types.IngestionConfig{Type: "invalid"})
	assert.Error(t, err)
}

func TestBasicIngester_IngestNotRunning(t *testing.T) {
	ing := ingestion.NewBasicIngester()
	err := ing.Init(types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)

	err = ing.Ingest(&types.BatchData{})
	assert.Error(t, err)
}

func TestBasicIngester_StartTwice(t *testing.T) {
	ing := ingestion.NewBasicIngester()
	err := ing.Init(types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)

	err = ing.Start()
	require.NoError(t, err)

	err = ing.Start()
	assert.Error(t, err)

	err = ing.Stop()
	require.NoError(t, err)
}

func TestBasicIngester_StopNotRunning(t *testing.T) {
	ing := ingestion.NewBasicIngester()
	err := ing.Stop()
	assert.Error(t, err)
}

func TestBasicIngester_CloseWhenRunning(t *testing.T) {
	ing := ingestion.NewBasicIngester()
	err := ing.Init(types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)

	err = ing.Start()
	require.NoError(t, err)

	err = ing.Close()
	assert.NoError(t, err)
}

func TestBasicIngester_GetStats(t *testing.T) {
	ing := ingestion.NewBasicIngester()
	err := ing.Init(types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)

	stats := ing.GetStats()
	assert.NotNil(t, stats)
	assert.False(t, stats["running"].(bool))
}

func TestBasicProcessor_NilData(t *testing.T) {
	p := processing.NewBasicProcessor()
	err := p.Init(types.ProcessingConfig{Type: "basic"})
	require.NoError(t, err)

	result, err := p.Process(nil)
	assert.NoError(t, err)
	assert.Nil(t, result)

	result, err = p.Process(&types.BatchData{DataPoints: []*types.DataPoint{}})
	assert.NoError(t, err)
	assert.Empty(t, result.DataPoints)
}

func TestBasicProcessor_InitInvalid(t *testing.T) {
	p := processing.NewBasicProcessor()
	err := p.Init(types.ProcessingConfig{Type: "invalid"})
	assert.Error(t, err)
}

func TestBasicProcessor_VariousMetrics(t *testing.T) {
	p := processing.NewBasicProcessor()
	err := p.Init(types.ProcessingConfig{Type: "basic"})
	require.NoError(t, err)

	now := time.Now()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "power", Value: 5000, Timestamp: now},
			{DeviceID: "WND-001", Metric: "energy", Value: 10000, Timestamp: now},
			{DeviceID: "BAT-001", Metric: "current", Value: 1500, Timestamp: now},
			{DeviceID: "DEV-001", Metric: "voltage", Value: 50000, Timestamp: now},
			{DeviceID: "SOL-002", Metric: "temperature", Value: -10, Timestamp: now},
			{DeviceID: "SOL-003", Metric: "temperature", Value: 200, Timestamp: now},
			{DeviceID: "DEV-002", Metric: "power", Value: -100, Timestamp: now},
			{DeviceID: "DEV-003", Metric: "energy", Value: -50, Timestamp: now},
		},
	}
	result, err := p.Process(bd)
	assert.NoError(t, err)
	assert.Len(t, result.DataPoints, 8)
}

func TestBasicAnalyzer_Process(t *testing.T) {
	a := analysis.NewBasicAnalyzer()
	err := a.Init(types.AnalysisConfig{Type: "basic"})
	require.NoError(t, err)

	now := time.Now()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.0, Timestamp: now},
			{DeviceID: "dev1", Metric: "temp", Value: 26.0, Timestamp: now.Add(1 * time.Minute)},
			{DeviceID: "dev1", Metric: "temp", Value: 27.0, Timestamp: now.Add(2 * time.Minute)},
			{DeviceID: "dev2", Metric: "humidity", Value: 60.0, Timestamp: now},
			{DeviceID: "dev2", Metric: "humidity", Value: 65.0, Timestamp: now.Add(1 * time.Minute)},
		},
	}
	result, err := a.Process(bd)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDorisStorage_PartitionManagement(t *testing.T) {
	s := storage.NewDorisStorage()
	err := s.Init(types.StorageConfig{Type: "doris", Database: "test", Table: "data"})
	require.NoError(t, err)
	defer s.Close()

	now := time.Now()
	err = s.CreatePartition("p1", now, now.Add(24*time.Hour))
	assert.NoError(t, err)

	err = s.CreatePartition("p2", now.Add(24*time.Hour), now.Add(48*time.Hour))
	assert.NoError(t, err)

	stats, err := s.GetPartitionStats()
	assert.NoError(t, err)
	assert.Len(t, stats, 2)

	err = s.MigratePartition("p1", storage.TierWarm)
	assert.NoError(t, err)

	err = s.MigratePartition("p1", storage.TierCold)
	assert.NoError(t, err)

	err = s.MigratePartition("nonexistent", storage.TierHot)
	assert.Error(t, err)

	err = s.AutoMigratePartitions()
	assert.NoError(t, err)

	optStats, err := s.GetStorageOptimizationStats()
	assert.NoError(t, err)
	assert.NotNil(t, optStats)

	err = s.DropPartition("p1")
	assert.NoError(t, err)

	err = s.DropPartition("p2")
	assert.NoError(t, err)
}

func TestDorisStorage_StorageOptimizationConfig(t *testing.T) {
	s := storage.NewDorisStorage()
	err := s.Init(types.StorageConfig{Type: "doris"})
	require.NoError(t, err)
	defer s.Close()

	defaultConfig := s.GetStorageOptimizationConfig()
	assert.True(t, defaultConfig.EnableCompression)
	assert.True(t, defaultConfig.EnableTieredStorage)

	newConfig := storage.DefaultDorisStorageOptimizationConfig()
	newConfig.EnableCompression = false
	newConfig.EnableTieredStorage = false
	newConfig.CompressionType = "ZSTD"
	s.SetStorageOptimizationConfig(newConfig)

	updatedConfig := s.GetStorageOptimizationConfig()
	assert.False(t, updatedConfig.EnableCompression)
	assert.False(t, updatedConfig.EnableTieredStorage)
	assert.Equal(t, "ZSTD", updatedConfig.CompressionType)
}

func TestDorisStorage_WritePointAndFlush(t *testing.T) {
	s := storage.NewDorisStorage()
	err := s.Init(types.StorageConfig{Type: "doris", BatchSize: 2})
	require.NoError(t, err)
	defer s.Close()

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()}
	err = s.WritePoint(dp)
	assert.NoError(t, err)

	err = s.Flush()
	assert.NoError(t, err)

	_, err = s.ReadTimeRange(time.Now().Add(-1*time.Hour), time.Now(), "st1", "dev1", "temp")
	assert.NoError(t, err)
}

func TestDorisStorage_CloseNotStarted(t *testing.T) {
	s := storage.NewDorisStorage()
	err := s.Close()
	assert.NoError(t, err)
}

func TestClickHouseStorage_CacheOperations(t *testing.T) {
	s := storage.NewClickHouseStorage()
	err := s.Init(types.StorageConfig{Type: "clickhouse", Table: "test"})
	require.NoError(t, err)
	defer s.Close()

	_, err = s.Query("SELECT 1")
	assert.NoError(t, err)

	cacheStats, err := s.GetCacheStats()
	assert.NoError(t, err)
	assert.NotNil(t, cacheStats)

	err = s.ClearCache()
	assert.NoError(t, err)

	cacheStats2, err := s.GetCacheStats()
	assert.NoError(t, err)
	assert.Equal(t, 0, cacheStats2["total_items"])
}

func TestClickHouseStorage_WritePoint(t *testing.T) {
	s := storage.NewClickHouseStorage()
	err := s.Init(types.StorageConfig{Type: "clickhouse", Table: "test"})
	require.NoError(t, err)
	defer s.Close()

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()}
	err = s.WritePoint(dp)
	assert.NoError(t, err)
}

func TestClickHouseStorage_WriteEmpty(t *testing.T) {
	s := storage.NewClickHouseStorage()
	err := s.Init(types.StorageConfig{Type: "clickhouse", Table: "test"})
	require.NoError(t, err)
	defer s.Close()

	err = s.Write(&types.BatchData{DataPoints: []*types.DataPoint{}})
	assert.NoError(t, err)
}

func TestClickHouseStorage_CloseNotStarted(t *testing.T) {
	s := storage.NewClickHouseStorage()
	err := s.Close()
	assert.NoError(t, err)
}

func TestClickHouseStorage_InitInvalid(t *testing.T) {
	s := storage.NewClickHouseStorage()
	err := s.Init(types.StorageConfig{Type: "invalid"})
	assert.Error(t, err)
}

func TestClickHouseStorage_MaterializedViewNotStarted(t *testing.T) {
	s := storage.NewClickHouseStorage()
	err := s.CreateMaterializedView("mv", "tbl", "SELECT 1")
	assert.Error(t, err)

	_, err = s.ListMaterializedViews()
	assert.Error(t, err)

	err = s.DropMaterializedView("mv")
	assert.Error(t, err)

	err = s.RefreshMaterializedView("mv")
	assert.Error(t, err)

	_, err = s.ExplainQuery("SELECT 1")
	assert.Error(t, err)

	err = s.CreatePreAggregationTable("tbl", "1h")
	assert.Error(t, err)

	_, err = s.ListPreAggregationRules()
	assert.Error(t, err)

	err = s.EnablePreAggregationRule("r1")
	assert.Error(t, err)

	err = s.DisablePreAggregationRule("r1")
	assert.Error(t, err)

	err = s.DeletePreAggregationRule("r1")
	assert.Error(t, err)

	err = s.RefreshPreAggregation("tbl")
	assert.Error(t, err)
}

func TestClickHouseStorage_PreAggregationRuleNotFound(t *testing.T) {
	s := storage.NewClickHouseStorage()
	err := s.Init(types.StorageConfig{Type: "clickhouse", Table: "test"})
	require.NoError(t, err)
	defer s.Close()

	err = s.EnablePreAggregationRule("nonexistent")
	assert.Error(t, err)

	err = s.DisablePreAggregationRule("nonexistent")
	assert.Error(t, err)

	err = s.DeletePreAggregationRule("nonexistent")
	assert.Error(t, err)
}

func TestClickHouseStorage_CreatePreAggregationRuleInvalid(t *testing.T) {
	s := storage.NewClickHouseStorage()
	err := s.Init(types.StorageConfig{Type: "clickhouse", Table: "test"})
	require.NoError(t, err)
	defer s.Close()

	err = s.CreatePreAggregationRule("not a rule")
	assert.Error(t, err)
}

func TestClickHouseStorage_MultiDimensionNotStarted(t *testing.T) {
	s := storage.NewClickHouseStorage()
	_, err := s.MultiDimensionAggregation(nil, nil, time.Now(), time.Now(), nil)
	assert.Error(t, err)

	_, err = s.DimensionDrillDown(nil, "", nil, time.Now(), time.Now(), nil)
	assert.Error(t, err)

	_, err = s.DimensionCrossAnalysis(nil, nil, "", time.Now(), time.Now(), nil)
	assert.Error(t, err)

	_, err = s.GetDimensionValues("", time.Now(), time.Now(), nil)
	assert.Error(t, err)
}

func TestBasicVisualizer_CreateDashboardAutoIDCov(t *testing.T) {
	v := visualization.NewBasicVisualizer()
	err := v.Init(types.VisualizationConfig{Type: "basic"})
	require.NoError(t, err)

	panels := []types.Panel{
		{Title: "Panel 1", Type: "graph"},
		{Title: "Panel 2", Type: "gauge"},
	}
	err = v.CreateDashboard("auto-id-dashboard", panels)
	assert.NoError(t, err)

	retrieved, err := v.GetDashboard("auto-id-dashboard")
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
	assert.NotEmpty(t, retrieved[0].ID)
	assert.NotEmpty(t, retrieved[1].ID)
}

func TestBasicVisualizer_InvalidTypeCov(t *testing.T) {
	v := visualization.NewBasicVisualizer()
	err := v.Init(types.VisualizationConfig{Type: "grafana"})
	assert.Error(t, err)
}
