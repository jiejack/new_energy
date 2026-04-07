package timeseries

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBatchWriter(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		Database:     "test_db",
		Table:        "test_table",
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
		MaxRetries:   3,
		RetryDelay:   100 * time.Millisecond,
	}

	writer := NewBatchWriter(mock, config)

	assert.NotNil(t, writer)

	// 关闭写入器
	err := writer.Close()
	assert.NoError(t, err)
}

func TestNewBatchWriter_DefaultConfig(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	writer := NewBatchWriter(mock, nil)

	assert.NotNil(t, writer)

	err := writer.Close()
	assert.NoError(t, err)
}

func TestNewBatchWriter_InvalidBatchSize(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		BatchSize: -1,
	}

	writer := NewBatchWriter(mock, config)

	assert.NotNil(t, writer)

	err := writer.Close()
	assert.NoError(t, err)
}

func TestNewBatchWriter_InvalidFlushTimeout(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		FlushTimeout: -1,
	}

	writer := NewBatchWriter(mock, config)

	assert.NotNil(t, writer)

	err := writer.Close()
	assert.NoError(t, err)
}

func TestBatchWriter_Add(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		BatchSize:    10,
		FlushTimeout: 5 * time.Second,
	}

	writer := NewBatchWriter(mock, config)
	defer writer.Close()

	point := &DataPoint{
		PointID:   1,
		Timestamp: time.Now(),
		Value:     100.5,
		Quality:   QualityGood,
	}

	err := writer.Add(point)
	assert.NoError(t, err)

	// 添加 nil 点
	err = writer.Add(nil)
	assert.NoError(t, err)
}

func TestBatchWriter_AddMultiple(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		BatchSize:    5,
		FlushTimeout: 5 * time.Second,
	}

	writer := NewBatchWriter(mock, config)
	defer writer.Close()

	// 添加多个数据点
	for i := 0; i < 10; i++ {
		point := &DataPoint{
			PointID:   int64(i + 1),
			Timestamp: time.Now(),
			Value:     float64(i * 10),
			Quality:   QualityGood,
		}
		err := writer.Add(point)
		assert.NoError(t, err)
	}

	// 等待刷新完成
	time.Sleep(200 * time.Millisecond)
}

func TestBatchWriter_Flush(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
	}

	writer := NewBatchWriter(mock, config)
	defer writer.Close()

	// 添加数据点
	for i := 0; i < 5; i++ {
		point := &DataPoint{
			PointID:   int64(i + 1),
			Timestamp: time.Now(),
			Value:     float64(i * 10),
			Quality:   QualityGood,
		}
		writer.Add(point)
	}

	// 手动刷新
	err := writer.Flush()
	assert.NoError(t, err)
}

func TestBatchWriter_Flush_Empty(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
	}

	writer := NewBatchWriter(mock, config)
	defer writer.Close()

	// 刷新空缓冲区
	err := writer.Flush()
	assert.NoError(t, err)
}

func TestBatchWriter_Stats(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
	}

	writer := NewBatchWriter(mock, config)
	defer writer.Close()

	// 添加数据点
	for i := 0; i < 5; i++ {
		point := &DataPoint{
			PointID:   int64(i + 1),
			Timestamp: time.Now(),
			Value:     float64(i * 10),
			Quality:   QualityGood,
		}
		writer.Add(point)
	}

	stats := writer.Stats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(5), stats.TotalPoints)
}

func TestBatchWriter_Close(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
	}

	writer := NewBatchWriter(mock, config)

	// 添加数据点
	point := &DataPoint{
		PointID:   1,
		Timestamp: time.Now(),
		Value:     100.0,
		Quality:   QualityGood,
	}
	writer.Add(point)

	// 关闭写入器
	err := writer.Close()
	assert.NoError(t, err)
}

func TestMockTimeSeriesDB_Write(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	points := []*DataPoint{
		{PointID: 1, Value: 10.0, Timestamp: time.Now()},
		{PointID: 2, Value: 20.0, Timestamp: time.Now()},
	}

	err := mock.Write(context.Background(), points)
	assert.NoError(t, err)
	assert.Len(t, mock.points, 2)
}

func TestMockTimeSeriesDB_WriteBatch(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	points := []*DataPoint{
		{PointID: 1, Value: 10.0, Timestamp: time.Now()},
		{PointID: 2, Value: 20.0, Timestamp: time.Now()},
	}

	err := mock.WriteBatch(context.Background(), points)
	assert.NoError(t, err)
	assert.Len(t, mock.points, 2)
}

func TestMockTimeSeriesDB_WriteWithTable(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	points := []*DataPoint{
		{PointID: 1, Value: 10.0, Timestamp: time.Now()},
	}

	err := mock.WriteWithTable(context.Background(), "test_db", "test_table", points)
	assert.NoError(t, err)
	assert.Len(t, mock.points, 1)
}

func TestMockTimeSeriesDB_Query(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	// 添加数据
	mock.points = []*DataPoint{
		{PointID: 1, Value: 10.0, Timestamp: time.Now()},
		{PointID: 2, Value: 20.0, Timestamp: time.Now()},
	}

	result, err := mock.Query(context.Background(), &Query{})
	assert.NoError(t, err)
	assert.Len(t, result.Points, 2)
	assert.Equal(t, int64(2), result.Total)
}

func TestMockTimeSeriesDB_QueryRange(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	mock.points = []*DataPoint{
		{PointID: 1, Value: 10.0, Timestamp: time.Now()},
	}

	points, err := mock.QueryRange(context.Background(), time.Now().Add(-1*time.Hour), time.Now(), []int64{1})
	assert.NoError(t, err)
	assert.Len(t, points, 1)
}

func TestMockTimeSeriesDB_QueryLatest(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	mock.points = []*DataPoint{
		{PointID: 1, Value: 10.0, Timestamp: time.Now()},
		{PointID: 2, Value: 20.0, Timestamp: time.Now()},
	}

	result, err := mock.QueryLatest(context.Background(), []int64{1, 2})
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 10.0, result[1].Value)
	assert.Equal(t, 20.0, result[2].Value)
}

func TestMockTimeSeriesDB_Aggregate(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	result, err := mock.Aggregate(context.Background(), &AggregateQuery{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestMockTimeSeriesDB_Downsample(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	err := mock.Downsample(context.Background(), &DownsampleQuery{})
	assert.NoError(t, err)
}

func TestMockTimeSeriesDB_CreateDatabase(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	err := mock.CreateDatabase(context.Background(), "test_db")
	assert.NoError(t, err)
}

func TestMockTimeSeriesDB_DropDatabase(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	err := mock.DropDatabase(context.Background(), "test_db")
	assert.NoError(t, err)
}

func TestMockTimeSeriesDB_CreateTable(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	schema := &TableSchema{
		Name: "test_table",
		Columns: []ColumnSchema{
			{Name: "point_id", Type: "Int64"},
		},
	}

	err := mock.CreateTable(context.Background(), "test_db", "test_table", schema)
	assert.NoError(t, err)
}

func TestMockTimeSeriesDB_DropTable(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	err := mock.DropTable(context.Background(), "test_db", "test_table")
	assert.NoError(t, err)
}

func TestMockTimeSeriesDB_CreateRetentionPolicy(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	policy := &RetentionPolicy{
		Name:     "30d",
		Database: "test_db",
	}

	err := mock.CreateRetentionPolicy(context.Background(), policy)
	assert.NoError(t, err)
}

func TestMockTimeSeriesDB_UpdateRetentionPolicy(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	policy := &RetentionPolicy{
		Name:     "30d",
		Database: "test_db",
	}

	err := mock.UpdateRetentionPolicy(context.Background(), policy)
	assert.NoError(t, err)
}

func TestMockTimeSeriesDB_DeleteRetentionPolicy(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	err := mock.DeleteRetentionPolicy(context.Background(), "test_db", "30d")
	assert.NoError(t, err)
}

func TestMockTimeSeriesDB_Ping(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	err := mock.Ping(context.Background())
	assert.NoError(t, err)
}

func TestMockTimeSeriesDB_IsConnected(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	assert.True(t, mock.IsConnected())

	mock.Close()
	assert.False(t, mock.IsConnected())
}

func TestMockTimeSeriesDB_Close(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	err := mock.Close()
	assert.NoError(t, err)
	assert.True(t, mock.closed)
}

func TestDataPoint_WithTags(t *testing.T) {
	point := &DataPoint{
		PointID:   1,
		Timestamp: time.Now(),
		Value:     100.5,
		Quality:   QualityGood,
		Tags: map[string]string{
			"station": "station1",
			"device":  "device1",
		},
	}

	assert.Equal(t, int64(1), point.PointID)
	assert.Equal(t, 100.5, point.Value)
	assert.Equal(t, QualityGood, point.Quality)
	assert.Equal(t, "station1", point.Tags["station"])
	assert.Equal(t, "device1", point.Tags["device"])
}

func TestQuery_WithFilters(t *testing.T) {
	query := &Query{
		Database:  "test_db",
		Table:     "test_table",
		PointIDs:  []int64{1, 2, 3},
		StartTime: time.Now().Add(-24 * time.Hour),
		EndTime:   time.Now(),
		Limit:     100,
		Offset:    10,
		OrderBy:   "timestamp",
		Order:     "DESC",
	}

	assert.Equal(t, "test_db", query.Database)
	assert.Equal(t, "test_table", query.Table)
	assert.Len(t, query.PointIDs, 3)
	assert.Equal(t, 100, query.Limit)
	assert.Equal(t, 10, query.Offset)
	assert.Equal(t, "timestamp", query.OrderBy)
	assert.Equal(t, "DESC", query.Order)
}

func TestAggregateQuery_WithGroupBy(t *testing.T) {
	query := &AggregateQuery{
		Database:  "test_db",
		Table:     "test_table",
		PointIDs:  []int64{1, 2},
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now(),
		Interval:  5 * time.Minute,
		AggFunc:   "avg",
		GroupBy:   []string{"point_id", "device"},
		Fill:      "null",
	}

	assert.Equal(t, "avg", query.AggFunc)
	assert.Equal(t, 5*time.Minute, query.Interval)
	assert.Len(t, query.GroupBy, 2)
	assert.Equal(t, "null", query.Fill)
}

func TestDownsampleQuery_Fields(t *testing.T) {
	query := &DownsampleQuery{
		SourceDatabase: "nem_ts_raw",
		SourceTable:    "data_points",
		TargetDatabase: "nem_ts_downsampled",
		TargetTable:    "data_points_5m",
		PointIDs:       []int64{1, 2, 3},
		StartTime:      time.Now().Add(-24 * time.Hour),
		EndTime:        time.Now(),
		Interval:       5 * time.Minute,
		AggFunc:        "avg",
	}

	assert.Equal(t, "nem_ts_raw", query.SourceDatabase)
	assert.Equal(t, "nem_ts_downsampled", query.TargetDatabase)
	assert.Equal(t, 5*time.Minute, query.Interval)
	assert.Equal(t, "avg", query.AggFunc)
}

func TestRetentionPolicy_Fields(t *testing.T) {
	policy := &RetentionPolicy{
		Name:               "30d",
		Database:           "nem_ts",
		Duration:           30 * 24 * time.Hour,
		ShardGroupDuration: 24 * time.Hour,
		ReplicaN:           3,
		Default:            true,
	}

	assert.Equal(t, "30d", policy.Name)
	assert.Equal(t, 30*24*time.Hour, policy.Duration)
	assert.Equal(t, 24*time.Hour, policy.ShardGroupDuration)
	assert.Equal(t, 3, policy.ReplicaN)
	assert.True(t, policy.Default)
}

func TestTableSchema_Complete(t *testing.T) {
	schema := &TableSchema{
		Name: "data_points",
		Columns: []ColumnSchema{
			{Name: "point_id", Type: "Int64", Nullable: false},
			{Name: "timestamp", Type: "DateTime", Nullable: false},
			{Name: "value", Type: "Float64", Nullable: false},
			{Name: "quality", Type: "Int32", Nullable: true, Default: "0"},
		},
		PrimaryKey:  []string{"point_id", "timestamp"},
		SortKey:     []string{"timestamp"},
		PartitionBy: "toYYYYMM(timestamp)",
		Engine:      "MergeTree",
	}

	assert.Equal(t, "data_points", schema.Name)
	assert.Len(t, schema.Columns, 4)
	assert.Equal(t, "Int64", schema.Columns[0].Type)
	assert.False(t, schema.Columns[0].Nullable)
	assert.True(t, schema.Columns[3].Nullable)
	assert.Equal(t, "0", schema.Columns[3].Default)
	assert.Equal(t, []string{"point_id", "timestamp"}, schema.PrimaryKey)
	assert.Equal(t, []string{"timestamp"}, schema.SortKey)
	assert.Equal(t, "toYYYYMM(timestamp)", schema.PartitionBy)
	assert.Equal(t, "MergeTree", schema.Engine)
}

func TestWriterStats_Fields(t *testing.T) {
	stats := &WriterStats{
		TotalPoints:   1000,
		SuccessPoints: 950,
		FailedPoints:  50,
		TotalBatches:  10,
		TotalBytes:    1024000,
	}

	assert.Equal(t, int64(1000), stats.TotalPoints)
	assert.Equal(t, int64(950), stats.SuccessPoints)
	assert.Equal(t, int64(50), stats.FailedPoints)
	assert.Equal(t, int64(10), stats.TotalBatches)
	assert.Equal(t, int64(1024000), stats.TotalBytes)
}

func TestPoolStats_Fields(t *testing.T) {
	stats := &PoolStats{
		TotalConnections:  100,
		IdleConnections:   20,
		ActiveConnections: 80,
		WaitCount:         10,
		WaitDuration:      5 * time.Second,
	}

	assert.Equal(t, int64(100), stats.TotalConnections)
	assert.Equal(t, int64(20), stats.IdleConnections)
	assert.Equal(t, int64(80), stats.ActiveConnections)
	assert.Equal(t, int64(10), stats.WaitCount)
	assert.Equal(t, 5*time.Second, stats.WaitDuration)
}

func TestAggregateResult_WithSeries(t *testing.T) {
	result := &AggregateResult{
		Series: []TimeSeries{
			{
				PointID: 1,
				Tags:    map[string]string{"device": "d1"},
				Points: []AggPoint{
					{Timestamp: time.Now(), Value: 15.5, Count: 10},
					{Timestamp: time.Now().Add(5 * time.Minute), Value: 16.0, Count: 10},
				},
			},
		},
	}

	assert.Len(t, result.Series, 1)
	assert.Equal(t, int64(1), result.Series[0].PointID)
	assert.Len(t, result.Series[0].Points, 2)
	assert.Equal(t, 15.5, result.Series[0].Points[0].Value)
	assert.Equal(t, int64(10), result.Series[0].Points[0].Count)
}

func TestQueryResult_WithPoints(t *testing.T) {
	result := &QueryResult{
		Points: []*DataPoint{
			{PointID: 1, Value: 10.0, Timestamp: time.Now()},
			{PointID: 2, Value: 20.0, Timestamp: time.Now()},
			{PointID: 3, Value: 30.0, Timestamp: time.Now()},
		},
		Total: 3,
	}

	assert.Len(t, result.Points, 3)
	assert.Equal(t, int64(3), result.Total)
}

func TestIndexSuggestion_Fields(t *testing.T) {
	suggestion := IndexSuggestion{
		TableName:  "data_points",
		ColumnName: "point_id",
		IndexType:  "btree",
		Reason:     "frequently used in WHERE clause",
	}

	assert.Equal(t, "data_points", suggestion.TableName)
	assert.Equal(t, "point_id", suggestion.ColumnName)
	assert.Equal(t, "btree", suggestion.IndexType)
	assert.Equal(t, "frequently used in WHERE clause", suggestion.Reason)
}

func TestWriteBatchRequest_Fields(t *testing.T) {
	req := &WriteBatchRequest{
		Database: "test_db",
		Table:    "test_table",
		Points: []*DataPoint{
			{PointID: 1, Value: 10.0, Timestamp: time.Now()},
		},
	}

	assert.Equal(t, "test_db", req.Database)
	assert.Equal(t, "test_table", req.Table)
	assert.Len(t, req.Points, 1)
}

func TestWriteResult_Fields(t *testing.T) {
	result := &WriteResult{
		Success: 95,
		Failed:  5,
	}

	assert.Equal(t, int64(95), result.Success)
	assert.Equal(t, int64(5), result.Failed)
}

func TestTimeSeriesDBType_Constants(t *testing.T) {
	assert.Equal(t, TimeSeriesDBType("doris"), TimeSeriesDBTypeDoris)
	assert.Equal(t, TimeSeriesDBType("clickhouse"), TimeSeriesDBTypeClickHouse)
	assert.Equal(t, TimeSeriesDBType("influxdb"), TimeSeriesDBTypeInfluxDB)
	assert.Equal(t, TimeSeriesDBType("timescaledb"), TimeSeriesDBTypeTimescaleDB)
}

func TestQualityDescription_All(t *testing.T) {
	tests := []struct {
		quality  int
		expected string
	}{
		{QualityGood, "good"},
		{QualityBad, "bad"},
		{QualityUncertain, "uncertain"},
		{QualityMissing, "missing"},
		{999, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := QualityDescription(tt.quality)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQuality_Constants(t *testing.T) {
	assert.Equal(t, 0, QualityGood)
	assert.Equal(t, 1, QualityBad)
	assert.Equal(t, 2, QualityUncertain)
	assert.Equal(t, 3, QualityMissing)
}
