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

func TestBigDataService_Init_ClickHouse(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	err = svc.Close()
	assert.NoError(t, err)
}

func TestBigDataService_Init_Doris(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "doris"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	err = svc.Close()
	assert.NoError(t, err)
}

func TestBigDataService_Init_Unsupported(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "unsupported"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	assert.Error(t, err)
}

func TestBigDataService_Ingest(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	svc.StartIngestion()

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
		},
	}
	err = svc.Ingest(bd)
	assert.NoError(t, err)

	svc.StopIngestion()
}

func TestBigDataService_Store(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
		},
	}
	err = svc.Store(bd)
	assert.NoError(t, err)
}

func TestBigDataService_Analyze(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	result, err := svc.Analyze("test query")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestBigDataService_Visualize(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	err = svc.Visualize("nonexistent-dashboard", "panel-1", map[string]interface{}{"key": "value"})
	assert.Error(t, err)
}

func TestBigDataService_Process(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
		},
	}
	result, err := svc.Process(bd)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestBigDataService_StartStopIngestion(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	svc.StartIngestion()
	svc.StopIngestion()
}

func TestBigDataService_WritePoint(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()}
	err = svc.WritePoint(dp)
	assert.NoError(t, err)
}

func TestBigDataService_ReadTimeRange(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	result, err := svc.ReadTimeRange(start, end, "station-1", "dev1", "temp")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestBigDataService_Aggregate(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	result, err := svc.Aggregate("avg", "temp", start, end, "1h")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestBigDataService_Flush(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	err = svc.Flush()
	assert.NoError(t, err)
}

func TestBigDataService_GetStorageStats(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	stats, err := svc.GetStorageStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestBigDataService_GetProcessingStats(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	stats, err := svc.GetProcessingStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestBigDataService_MaterializedViews(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	err = svc.CreateMaterializedView("test_mv", "target_table", "SELECT * FROM source")
	assert.NoError(t, err)

	views, err := svc.ListMaterializedViews()
	assert.NoError(t, err)
	assert.NotNil(t, views)

	svc.RefreshMaterializedView("test_mv")
	svc.DropMaterializedView("test_mv")
}

func TestBigDataService_PreAggregation(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	err = svc.CreatePreAggregationTable("agg_table", "1h")
	assert.NoError(t, err)

	err = svc.CreatePreAggregationRule(storage.PreAggregationRule{
		ID:           "rule1",
		SourceTable:  "source_table",
		TargetTable:  "agg_table",
		Aggregation:  "avg",
		GroupBy:      []string{"device_id"},
		TimeInterval: "1h",
		Enabled:      true,
	})
	assert.NoError(t, err)

	rules, err := svc.ListPreAggregationRules()
	assert.NoError(t, err)
	assert.NotNil(t, rules)

	svc.EnablePreAggregationRule("rule1")
	svc.DisablePreAggregationRule("rule1")
	svc.DeletePreAggregationRule("rule1")
	svc.RefreshPreAggregation("agg_table")
}

func TestBigDataService_Cache(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	stats, err := svc.GetCacheStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	err = svc.ClearCache()
	assert.NoError(t, err)
}

func TestBigDataService_MultiDimension(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	result, err := svc.MultiDimensionAggregation([]string{"temp", "humidity"}, []string{"device_id"}, start, end, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	result, err = svc.DimensionDrillDown([]string{"device_id"}, "metric", []string{"temp"}, start, end, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	result, err = svc.DimensionCrossAnalysis([]string{"device_id"}, []string{"metric"}, "temp", start, end, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	values, err := svc.GetDimensionValues("device_id", start, end, nil)
	assert.NoError(t, err)
	assert.NotNil(t, values)
}

func TestBigDataService_ExplainQuery(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(types.StorageConfig{Type: "clickhouse"}, types.AnalysisConfig{Type: "basic"}, types.VisualizationConfig{Type: "basic"}, types.ProcessingConfig{Type: "basic"}, types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
	defer svc.Close()

	result, err := svc.ExplainQuery("SELECT * FROM table")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestClickHouseStorage_Comprehensive(t *testing.T) {
	s := storage.NewClickHouseStorage()
	err := s.Init(types.StorageConfig{Type: "clickhouse"})
	require.NoError(t, err)

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
			{DeviceID: "dev2", Metric: "humidity", Value: 60.0, Timestamp: time.Now()},
		},
	}
	err = s.Write(bd)
	assert.NoError(t, err)

	points, err := s.Read("dev1")
	assert.NoError(t, err)
	assert.NotNil(t, points)

	result, err := s.Query("SELECT * FROM table")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	aggResult, err := s.Aggregate("avg", "temp", start, end, "1h")
	assert.NoError(t, err)
	assert.NotNil(t, aggResult)

	stats, err := s.GetStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	err = s.Flush()
	assert.NoError(t, err)

	err = s.Close()
	assert.NoError(t, err)
}

func TestDorisStorage_Comprehensive(t *testing.T) {
	s := storage.NewDorisStorage()
	err := s.Init(types.StorageConfig{Type: "doris"})
	require.NoError(t, err)

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
		},
	}
	err = s.Write(bd)
	assert.NoError(t, err)

	points, err := s.Read("dev1")
	assert.NoError(t, err)
	assert.NotNil(t, points)

	result, err := s.Query("SELECT * FROM table")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	aggResult, err := s.Aggregate("avg", "temp", start, end, "1h")
	assert.NoError(t, err)
	assert.NotNil(t, aggResult)

	stats, err := s.GetStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	err = s.Flush()
	assert.NoError(t, err)

	err = s.Close()
	assert.NoError(t, err)
}

func TestBasicProcessor_Comprehensive(t *testing.T) {
	p := processing.NewBasicProcessor()
	err := p.Init(types.ProcessingConfig{Type: "basic"})
	require.NoError(t, err)

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
			{DeviceID: "dev1", Metric: "temp", Value: 26.0, Timestamp: time.Now()},
		},
	}
	result, err := p.Process(bd)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	err = p.Close()
	assert.NoError(t, err)
}

func TestBasicIngester_Comprehensive(t *testing.T) {
	ing := ingestion.NewBasicIngester()
	ing.Start()
	stats := ing.GetStats()
	assert.NotNil(t, stats)
	ing.Stop()
}

func TestBasicAnalyzer_Comprehensive(t *testing.T) {
	a := analysis.NewBasicAnalyzer()
	err := a.Init(types.AnalysisConfig{Type: "basic"})
	require.NoError(t, err)

	result, err := a.Execute("test query")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
			{DeviceID: "dev1", Metric: "temp", Value: 26.0, Timestamp: time.Now()},
			{DeviceID: "dev1", Metric: "temp", Value: 27.0, Timestamp: time.Now()},
		},
	}
	result, err = a.Process(bd)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	_, err = a.Process(nil)
	assert.Error(t, err)

	_, err = a.Process(&types.BatchData{DataPoints: []*types.DataPoint{}})
	assert.Error(t, err)

	err = a.Close()
	assert.NoError(t, err)
}

func TestBasicAnalyzer_InvalidType(t *testing.T) {
	a := analysis.NewBasicAnalyzer()
	err := a.Init(types.AnalysisConfig{Type: "invalid"})
	assert.Error(t, err)
}

func TestBasicVisualizer_Comprehensive(t *testing.T) {
	v := visualization.NewBasicVisualizer()
	err := v.Init(types.VisualizationConfig{Type: "basic"})
	require.NoError(t, err)

	panels := []types.Panel{
		{ID: "panel-1", Title: "Test Panel", Type: "line"},
	}
	err = v.CreateDashboard("dashboard-1", panels)
	assert.NoError(t, err)

	err = v.CreateDashboard("", panels)
	assert.Error(t, err)

	retrieved, err := v.GetDashboard("dashboard-1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)

	_, err = v.GetDashboard("nonexistent")
	assert.Error(t, err)

	err = v.UpdatePanel("dashboard-1", "panel-1", map[string]interface{}{"data": "updated"})
	assert.NoError(t, err)

	err = v.UpdatePanel("nonexistent", "panel-1", nil)
	assert.Error(t, err)

	err = v.UpdatePanel("dashboard-1", "nonexistent", nil)
	assert.Error(t, err)

	dashboards := v.ListDashboards()
	assert.Contains(t, dashboards, "dashboard-1")

	err = v.Close()
	assert.NoError(t, err)
}

func TestBasicVisualizer_Charts(t *testing.T) {
	v := visualization.NewBasicVisualizer()

	dataPoints := []*types.DataPoint{
		{Value: 10.0, Timestamp: time.Now()},
		{Value: 20.0, Timestamp: time.Now().Add(1 * time.Minute)},
	}
	chart := v.GenerateTimeSeriesChart(dataPoints)
	assert.NotNil(t, chart)
	assert.Equal(t, "time_series", chart["type"])

	gauge := v.GenerateGaugeChart(75.0, 0, 100, "CPU Usage")
	assert.NotNil(t, gauge)
	assert.Equal(t, "gauge", gauge["type"])

	bar := v.GenerateBarChart([]string{"A", "B", "C"}, []float64{10, 20, 30})
	assert.NotNil(t, bar)
	assert.Equal(t, "bar", bar["type"])

	pie := v.GeneratePieChart([]string{"A", "B"}, []float64{60, 40})
	assert.NotNil(t, pie)
	assert.Equal(t, "pie", pie["type"])
}

func TestBasicVisualizer_InvalidType(t *testing.T) {
	v := visualization.NewBasicVisualizer()
	err := v.Init(types.VisualizationConfig{Type: "invalid"})
	assert.Error(t, err)
}

func TestTypes_Error(t *testing.T) {
	e := &types.Error{Code: types.ErrCodeInvalidConfig, Message: "test error"}
	assert.Equal(t, types.ErrCodeInvalidConfig, e.Code)
	assert.Equal(t, "test error", e.Message)
	assert.Contains(t, e.Error(), "test error")
}

func TestTypes_DataPoint(t *testing.T) {
	dp := &types.DataPoint{
		DeviceID:  "dev1",
		Metric:    "temp",
		Value:     25.5,
		Timestamp: time.Now(),
		Tags:      map[string]string{"location": "room1"},
	}
	assert.Equal(t, "dev1", dp.DeviceID)
	assert.Equal(t, "temp", dp.Metric)
	assert.InDelta(t, 25.5, dp.Value, 0.001)
}

func TestTypes_BatchData(t *testing.T) {
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", RecordCount: 1},
	}
	assert.Len(t, bd.DataPoints, 1)
	assert.Equal(t, "test", bd.Metadata.Source)
}

func TestTypes_StorageConfig(t *testing.T) {
	cfg := types.StorageConfig{
		Type:     "clickhouse",
		Host:     "localhost",
		Port:     9000,
		Database: "test_db",
		Username: "default",
		Password: "",
	}
	assert.Equal(t, "clickhouse", cfg.Type)
	assert.Equal(t, "localhost", cfg.Host)
}

func TestTypes_Pool(t *testing.T) {
	dp := types.DataPointPool.Get()
	assert.NotNil(t, dp)

	bp := types.BatchDataPool.Get()
	assert.NotNil(t, bp)
}
