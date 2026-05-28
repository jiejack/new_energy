package storage

import (
	"testing"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDorisStorage(t *testing.T) {
	s := NewDorisStorage()
	assert.NotNil(t, s)
}

func TestDorisStorage_Init(t *testing.T) {
	s := NewDorisStorage()
	err := s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	require.NoError(t, err)
	assert.True(t, s.started)
	s.Close()
}

func TestDorisStorage_Init_CustomSettings(t *testing.T) {
	s := NewDorisStorage()
	err := s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table", BatchSize: 500, FlushInterval: 10})
	require.NoError(t, err)
	assert.Equal(t, 500, s.batchSize)
	s.Close()
}

func TestDorisStorage_Write(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
		},
	}
	err := s.Write(bd)
	assert.NoError(t, err)
}

func TestDorisStorage_Write_Empty(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	err := s.Write(&types.BatchData{DataPoints: []*types.DataPoint{}})
	assert.NoError(t, err)
}

func TestDorisStorage_WritePoint(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	dp := &types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()}
	err := s.WritePoint(dp)
	assert.NoError(t, err)
}

func TestDorisStorage_Flush(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	s.WritePoint(&types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()})
	err := s.Flush()
	assert.NoError(t, err)
}

func TestDorisStorage_Read(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	points, err := s.Read("SELECT * FROM test_table")
	assert.NoError(t, err)
	assert.NotNil(t, points)
}

func TestDorisStorage_ReadTimeRange(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	points, err := s.ReadTimeRange(start, end, "station-1", "dev1", "temp")
	assert.NoError(t, err)
	assert.NotNil(t, points)
}

func TestDorisStorage_Query(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	result, err := s.Query("SELECT * FROM test_table")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDorisStorage_Aggregate(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	result, err := s.Aggregate("avg", "temp", start, end, "1h")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDorisStorage_GetStats(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	stats, err := s.GetStats()
	assert.NoError(t, err)
	assert.Equal(t, "doris", stats["storage_type"])
}

func TestDorisStorage_Close(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	err := s.Close()
	assert.NoError(t, err)
	assert.False(t, s.started)
}

func TestDorisStorage_Close_NotStarted(t *testing.T) {
	s := NewDorisStorage()
	err := s.Close()
	assert.NoError(t, err)
}

func TestDorisStorage_MaterializedViews(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	err := s.CreateMaterializedView("test_mv", "target", "SELECT 1")
	assert.Error(t, err)

	views, err := s.ListMaterializedViews()
	assert.NoError(t, err)
	assert.NotNil(t, views)

	err = s.DropMaterializedView("test_mv")
	assert.Error(t, err)

	err = s.RefreshMaterializedView("test_mv")
	assert.Error(t, err)
}

func TestDorisStorage_ExplainQuery(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	result, err := s.ExplainQuery("SELECT * FROM test_table")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDorisStorage_PreAggregation(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	err := s.CreatePreAggregationTable("agg_table", "1h")
	assert.Error(t, err)

	err = s.CreatePreAggregationRule(nil)
	assert.Error(t, err)

	rules, err := s.ListPreAggregationRules()
	assert.NoError(t, err)
	assert.NotNil(t, rules)

	err = s.EnablePreAggregationRule("r1")
	assert.Error(t, err)

	err = s.DisablePreAggregationRule("r1")
	assert.Error(t, err)

	err = s.DeletePreAggregationRule("r1")
	assert.Error(t, err)

	err = s.RefreshPreAggregation("table")
	assert.Error(t, err)
}

func TestDorisStorage_Cache(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	stats, err := s.GetCacheStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	err = s.ClearCache()
	assert.NoError(t, err)
}

func TestDorisStorage_MultiDimension(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	result, err := s.MultiDimensionAggregation([]string{"temp"}, []string{"device_id"}, start, end, map[string]interface{}{"key": "val"})
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

func TestDorisStorage_PartitionManagement(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	now := time.Now()
	err := s.CreatePartition("p1", now, now.Add(24*time.Hour))
	assert.NoError(t, err)

	stats, err := s.GetPartitionStats()
	assert.NoError(t, err)
	assert.Len(t, stats, 1)

	err = s.MigratePartition("p1", TierWarm)
	assert.NoError(t, err)

	err = s.MigratePartition("nonexistent", TierCold)
	assert.Error(t, err)

	err = s.DropPartition("p1")
	assert.NoError(t, err)

	stats, _ = s.GetPartitionStats()
	assert.Len(t, stats, 0)
}

func TestDorisStorage_StorageOptimization(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	optConfig := DefaultDorisStorageOptimizationConfig()
	assert.Equal(t, "LZ4", optConfig.Compression)
	assert.Equal(t, 32, optConfig.Buckets)

	optConfig.EnableCompression = false
	s.SetStorageOptimizationConfig(optConfig)

	gotConfig := s.GetStorageOptimizationConfig()
	assert.False(t, gotConfig.EnableCompression)

	stats, err := s.GetStorageOptimizationStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestDorisStorage_AutoMigratePartitions(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	optConfig := DefaultDorisStorageOptimizationConfig()
	optConfig.HotPartitionDays = 3
	optConfig.WarmPartitionDays = 7
	s.SetStorageOptimizationConfig(optConfig)

	now := time.Now()
	s.CreatePartition("hot_p", now.Add(-1*24*time.Hour), now)
	s.CreatePartition("warm_p", now.Add(-5*24*time.Hour), now.Add(-4*24*time.Hour))
	s.CreatePartition("cold_p", now.Add(-10*24*time.Hour), now.Add(-9*24*time.Hour))

	err := s.AutoMigratePartitions()
	assert.NoError(t, err)
}

func TestDorisStorage_WriteTriggersFlush(t *testing.T) {
	s := NewDorisStorage()
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table", BatchSize: 2})
	defer s.Close()

	for i := 0; i < 3; i++ {
		s.WritePoint(&types.DataPoint{DeviceID: "dev1", Metric: "temp", Value: float64(i), Timestamp: time.Now()})
	}
}

func TestDorisStorage_CacheEviction(t *testing.T) {
	s := NewDorisStorage()
	s.cacheMaxSize = 5
	s.Init(types.StorageConfig{Type: "doris", Database: "test_db", Table: "test_table"})
	defer s.Close()

	for i := 0; i < 10; i++ {
		s.Query("SELECT * FROM test_table WHERE id = " + string(rune(i)))
	}

	stats, _ := s.GetCacheStats()
	assert.LessOrEqual(t, stats["total_items"], 10)
}
