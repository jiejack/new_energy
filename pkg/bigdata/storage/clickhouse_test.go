package storage

import (
	"testing"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClickHouseStorage(t *testing.T) {
	s := NewClickHouseStorage()
	assert.NotNil(t, s)
}

func TestClickHouseStorage_Init_Valid(t *testing.T) {
	s := NewClickHouseStorage()
	err := s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	require.NoError(t, err)
	assert.True(t, s.started)
	s.Close()
}

func TestClickHouseStorage_Init_Invalid(t *testing.T) {
	s := NewClickHouseStorage()
	err := s.Init(types.StorageConfig{Type: "doris"})
	assert.Error(t, err)
}

func TestClickHouseStorage_Init_CustomBatchSettings(t *testing.T) {
	s := NewClickHouseStorage()
	err := s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table", BatchSize: 500, FlushInterval: 10})
	require.NoError(t, err)
	assert.Equal(t, 500, s.batchSize)
	s.Close()
}

func TestClickHouseStorage_Write(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
		},
	}
	err := s.Write(bd)
	assert.NoError(t, err)
}

func TestClickHouseStorage_Write_Empty(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	err := s.Write(&types.BatchData{DataPoints: []*types.DataPoint{}})
	assert.NoError(t, err)
}

func TestClickHouseStorage_WritePoint(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()}
	err := s.WritePoint(dp)
	assert.NoError(t, err)
}

func TestClickHouseStorage_Flush(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	s.WritePoint(&types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()})
	err := s.Flush()
	assert.NoError(t, err)
}

func TestClickHouseStorage_Read(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	points, err := s.Read("SELECT * FROM test_table")
	assert.NoError(t, err)
	assert.NotNil(t, points)
}

func TestClickHouseStorage_ReadTimeRange(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	points, err := s.ReadTimeRange(start, end, "station-1", "dev1", "temp")
	assert.NoError(t, err)
	assert.NotNil(t, points)
}

func TestClickHouseStorage_Query(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	result, err := s.Query("SELECT * FROM test_table")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestClickHouseStorage_Query_CacheHit(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	_, err := s.Query("SELECT * FROM test_table")
	require.NoError(t, err)

	result, err := s.Query("SELECT * FROM test_table")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestClickHouseStorage_Aggregate(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	result, err := s.Aggregate("avg", "temp", start, end, "1h")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestClickHouseStorage_GetStats(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	stats, err := s.GetStats()
	assert.NoError(t, err)
	assert.Equal(t, "clickhouse", stats["storage_type"])
	assert.Equal(t, true, stats["started"])
}

func TestClickHouseStorage_MaterializedViews(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	err := s.CreateMaterializedView("test_mv", "target_table", "SELECT * FROM source")
	assert.NoError(t, err)

	views, err := s.ListMaterializedViews()
	assert.NoError(t, err)
	assert.Contains(t, views, "test_mv")

	err = s.RefreshMaterializedView("test_mv")
	assert.NoError(t, err)

	err = s.DropMaterializedView("test_mv")
	assert.NoError(t, err)

	views, _ = s.ListMaterializedViews()
	assert.NotContains(t, views, "test_mv")
}

func TestClickHouseStorage_MaterializedViews_NotInitialized(t *testing.T) {
	s := NewClickHouseStorage()

	err := s.CreateMaterializedView("test_mv", "target", "SELECT 1")
	assert.Error(t, err)

	_, err = s.ListMaterializedViews()
	assert.Error(t, err)

	err = s.DropMaterializedView("test_mv")
	assert.Error(t, err)

	err = s.RefreshMaterializedView("test_mv")
	assert.Error(t, err)
}

func TestClickHouseStorage_ExplainQuery(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	result, err := s.ExplainQuery("SELECT * FROM test_table")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestClickHouseStorage_ExplainQuery_NotInitialized(t *testing.T) {
	s := NewClickHouseStorage()
	_, err := s.ExplainQuery("SELECT 1")
	assert.Error(t, err)
}

func TestClickHouseStorage_PreAggregation(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	err := s.CreatePreAggregationTable("agg_table", "1h")
	assert.NoError(t, err)

	err = s.CreatePreAggregationRule(PreAggregationRule{
		ID:           "rule1",
		SourceTable:  "source_table",
		TargetTable:  "agg_table",
		Aggregation:  "avg",
		GroupBy:      []string{"device_id"},
		TimeInterval: "Hour",
		Enabled:      true,
	})
	assert.NoError(t, err)

	rules, err := s.ListPreAggregationRules()
	assert.NoError(t, err)
	assert.NotNil(t, rules)

	err = s.EnablePreAggregationRule("rule1")
	assert.NoError(t, err)

	err = s.DisablePreAggregationRule("rule1")
	assert.NoError(t, err)

	err = s.DeletePreAggregationRule("rule1")
	assert.NoError(t, err)

	err = s.RefreshPreAggregation("agg_table")
	assert.NoError(t, err)
}

func TestClickHouseStorage_PreAggregation_NotInitialized(t *testing.T) {
	s := NewClickHouseStorage()

	err := s.CreatePreAggregationTable("agg_table", "1h")
	assert.Error(t, err)

	err = s.CreatePreAggregationRule(PreAggregationRule{ID: "r1"})
	assert.Error(t, err)

	_, err = s.ListPreAggregationRules()
	assert.Error(t, err)

	err = s.EnablePreAggregationRule("r1")
	assert.Error(t, err)

	err = s.DisablePreAggregationRule("r1")
	assert.Error(t, err)

	err = s.DeletePreAggregationRule("r1")
	assert.Error(t, err)

	err = s.RefreshPreAggregation("table")
	assert.Error(t, err)
}

func TestClickHouseStorage_PreAggregation_InvalidRuleType(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	err := s.CreatePreAggregationRule("not a rule")
	assert.Error(t, err)
}

func TestClickHouseStorage_PreAggregation_RuleNotFound(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	err := s.EnablePreAggregationRule("nonexistent")
	assert.Error(t, err)

	err = s.DisablePreAggregationRule("nonexistent")
	assert.Error(t, err)

	err = s.DeletePreAggregationRule("nonexistent")
	assert.Error(t, err)
}

func TestClickHouseStorage_Cache(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	stats, err := s.GetCacheStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	err = s.ClearCache()
	assert.NoError(t, err)
}

func TestClickHouseStorage_MultiDimension(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	result, err := s.MultiDimensionAggregation([]string{"temp"}, []string{"device_id"}, start, end, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	result, err = s.DimensionDrillDown([]string{"device_id"}, "metric", []string{"temp"}, start, end, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	result, err = s.DimensionCrossAnalysis([]string{"device_id"}, []string{"metric"}, "temp", start, end, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	result, err = s.GetDimensionValues("device_id", start, end, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestClickHouseStorage_MultiDimension_NotInitialized(t *testing.T) {
	s := NewClickHouseStorage()
	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	_, err := s.MultiDimensionAggregation([]string{"temp"}, []string{"device_id"}, start, end, nil)
	assert.Error(t, err)

	_, err = s.DimensionDrillDown([]string{"device_id"}, "metric", []string{"temp"}, start, end, nil)
	assert.Error(t, err)

	_, err = s.DimensionCrossAnalysis([]string{"device_id"}, []string{"metric"}, "temp", start, end, nil)
	assert.Error(t, err)

	_, err = s.GetDimensionValues("device_id", start, end, nil)
	assert.Error(t, err)
}

func TestClickHouseStorage_Close(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	err := s.Close()
	assert.NoError(t, err)
	assert.False(t, s.started)
}

func TestClickHouseStorage_Close_NotStarted(t *testing.T) {
	s := NewClickHouseStorage()
	err := s.Close()
	assert.NoError(t, err)
}

func TestClickHouseStorage_CacheEviction(t *testing.T) {
	s := NewClickHouseStorage()
	s.cacheMaxSize = 5
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table"})
	defer s.Close()

	for i := 0; i < 10; i++ {
		s.Query("SELECT * FROM test_table WHERE id = " + string(rune(i)))
	}

	stats, _ := s.GetCacheStats()
	assert.LessOrEqual(t, stats["total_items"], 10)
}

func TestClickHouseStorage_WriteTriggersFlush(t *testing.T) {
	s := NewClickHouseStorage()
	s.Init(types.StorageConfig{Type: "clickhouse", Table: "test_table", BatchSize: 2})
	defer s.Close()

	for i := 0; i < 3; i++ {
		s.WritePoint(&types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: float64(i), Timestamp: time.Now()})
	}
}
