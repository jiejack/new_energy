package bigdata

import (
	"testing"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBigDataService_UninitializedMethods(t *testing.T) {
	svc := NewBigDataService()

	_, err := svc.Analyze("test")
	assert.Error(t, err)

	err = svc.Visualize("d1", "p1", nil)
	assert.Error(t, err)

	_, err = svc.Process(nil)
	assert.Error(t, err)

	err = svc.StartIngestion()
	assert.Error(t, err)

	err = svc.StopIngestion()
	assert.Error(t, err)

	err = svc.WritePoint(nil)
	assert.Error(t, err)

	_, err = svc.ReadTimeRange(time.Now(), time.Now(), "", "", "")
	assert.Error(t, err)

	_, err = svc.Aggregate("avg", "temp", time.Now(), time.Now(), "1h")
	assert.Error(t, err)

	err = svc.Flush()
	assert.Error(t, err)

	_, err = svc.GetStorageStats()
	assert.Error(t, err)

	err = svc.CreateMaterializedView("mv", "table", "SELECT 1")
	assert.Error(t, err)

	_, err = svc.ListMaterializedViews()
	assert.Error(t, err)

	err = svc.DropMaterializedView("mv")
	assert.Error(t, err)

	err = svc.RefreshMaterializedView("mv")
	assert.Error(t, err)

	_, err = svc.ExplainQuery("SELECT 1")
	assert.Error(t, err)

	err = svc.CreatePreAggregationTable("table", "1h")
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

	err = svc.RefreshPreAggregation("table")
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
}

func TestBigDataService_UninitializedIngest(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Ingest(nil)
	assert.Error(t, err)
}

func TestBigDataService_UninitializedStore(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Store(nil)
	assert.Error(t, err)
}

func TestBigDataService_Init_UnsupportedAnalysis(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(
		types.StorageConfig{Type: "clickhouse"},
		types.AnalysisConfig{Type: "unsupported"},
		types.VisualizationConfig{Type: "basic"},
		types.ProcessingConfig{Type: "basic"},
		types.IngestionConfig{Type: "basic"},
	)
	assert.Error(t, err)
}

func TestBigDataService_Init_UnsupportedVisualization(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(
		types.StorageConfig{Type: "clickhouse"},
		types.AnalysisConfig{Type: "basic"},
		types.VisualizationConfig{Type: "unsupported"},
		types.ProcessingConfig{Type: "basic"},
		types.IngestionConfig{Type: "basic"},
	)
	assert.Error(t, err)
}

func TestBigDataService_Init_UnsupportedProcessing(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(
		types.StorageConfig{Type: "clickhouse"},
		types.AnalysisConfig{Type: "basic"},
		types.VisualizationConfig{Type: "basic"},
		types.ProcessingConfig{Type: "unsupported"},
		types.IngestionConfig{Type: "basic"},
	)
	assert.Error(t, err)
}

func TestBigDataService_Init_UnsupportedIngestion(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(
		types.StorageConfig{Type: "clickhouse"},
		types.AnalysisConfig{Type: "basic"},
		types.VisualizationConfig{Type: "basic"},
		types.ProcessingConfig{Type: "basic"},
		types.IngestionConfig{Type: "unsupported"},
	)
	assert.Error(t, err)
}

func TestBigDataService_Close_NilComponents(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Close()
	assert.NoError(t, err)
}

func TestBigDataService_Init_Flink(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(
		types.StorageConfig{Type: "clickhouse"},
		types.AnalysisConfig{Type: "basic"},
		types.VisualizationConfig{Type: "basic"},
		types.ProcessingConfig{Type: "flink"},
		types.IngestionConfig{Type: "basic"},
	)
	require.NoError(t, err)
	svc.Close()
}

func TestBigDataService_Visualize_WithDashboard(t *testing.T) {
	svc := NewBigDataService()
	err := svc.Init(
		types.StorageConfig{Type: "clickhouse"},
		types.AnalysisConfig{Type: "basic"},
		types.VisualizationConfig{Type: "basic"},
		types.ProcessingConfig{Type: "basic"},
		types.IngestionConfig{Type: "basic"},
	)
	require.NoError(t, err)
	defer svc.Close()

	svc.Visualize("nonexistent", "p1", nil)
}
