package timeseries

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDataPoint(t *testing.T) {
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
}

func TestQualityDescription(t *testing.T) {
	tests := []struct {
		quality    int
		expected   string
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

func TestQuery(t *testing.T) {
	query := &Query{
		Database:  "test_db",
		Table:     "test_table",
		PointIDs:  []int64{1, 2, 3},
		StartTime: time.Now().Add(-24 * time.Hour),
		EndTime:   time.Now(),
		Limit:     100,
		Offset:    0,
		OrderBy:   "timestamp",
		Order:     "DESC",
	}

	assert.Equal(t, "test_db", query.Database)
	assert.Equal(t, "test_table", query.Table)
	assert.Len(t, query.PointIDs, 3)
	assert.Equal(t, 100, query.Limit)
}

func TestAggregateQuery(t *testing.T) {
	query := &AggregateQuery{
		Database:  "test_db",
		Table:     "test_table",
		PointIDs:  []int64{1, 2},
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now(),
		Interval:  5 * time.Minute,
		AggFunc:   "avg",
		GroupBy:   []string{"point_id"},
		Fill:      "null",
	}

	assert.Equal(t, "avg", query.AggFunc)
	assert.Equal(t, 5*time.Minute, query.Interval)
	assert.Equal(t, "null", query.Fill)
}

func TestQueryResult(t *testing.T) {
	result := &QueryResult{
		Points: []*DataPoint{
			{PointID: 1, Value: 10.0},
			{PointID: 2, Value: 20.0},
		},
		Total: 2,
	}

	assert.Len(t, result.Points, 2)
	assert.Equal(t, int64(2), result.Total)
}

func TestAggregateResult(t *testing.T) {
	result := &AggregateResult{
		Series: []TimeSeries{
			{
				PointID: 1,
				Tags:    map[string]string{"device": "d1"},
				Points: []AggPoint{
					{Timestamp: time.Now(), Value: 15.5, Count: 10},
				},
			},
		},
	}

	assert.Len(t, result.Series, 1)
	assert.Equal(t, int64(1), result.Series[0].PointID)
	assert.Len(t, result.Series[0].Points, 1)
}

func TestTableSchema(t *testing.T) {
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
}

func TestErrors(t *testing.T) {
	// Test QueryError
	queryErr := NewQueryError("SELECT * FROM test", "syntax error", nil)
	assert.Contains(t, queryErr.Error(), "syntax error")
	assert.Contains(t, queryErr.Error(), "SELECT * FROM test")

	// Test WriteError
	writeErr := NewWriteError(100, "connection refused", nil)
	assert.Contains(t, writeErr.Error(), "connection refused")
	assert.Contains(t, writeErr.Error(), "100")

	// Test ConnectionError
	connErr := NewConnectionError("localhost:9000", "timeout", nil)
	assert.Contains(t, connErr.Error(), "localhost:9000")
	assert.Contains(t, connErr.Error(), "timeout")
}

func TestIsRetryableError(t *testing.T) {
	// Connection error is retryable
	connErr := NewConnectionError("localhost:9000", "timeout", nil)
	assert.True(t, IsRetryableError(connErr))

	// Timeout error is retryable
	assert.True(t, IsRetryableError(ErrTimeout))

	// Other errors are not retryable
	assert.False(t, IsRetryableError(ErrInvalidConfig))
	assert.False(t, IsRetryableError(nil))
}

func TestIsNotFoundError(t *testing.T) {
	assert.True(t, IsNotFoundError(ErrDatabaseNotFound))
	assert.True(t, IsNotFoundError(ErrTableNotFound))
	assert.False(t, IsNotFoundError(ErrConnectionFailed))
	assert.False(t, IsNotFoundError(nil))
}

func TestDorisConfig(t *testing.T) {
	config := DefaultDorisConfig()
	assert.Equal(t, []string{"localhost"}, config.Hosts)
	assert.Equal(t, 9030, config.Port)
	assert.Equal(t, "nem_ts", config.Database)
	assert.Equal(t, "root", config.User)
	assert.Equal(t, 100, config.MaxOpenConns)
	assert.Equal(t, 20, config.MaxIdleConns)
	assert.Equal(t, 10000, config.BatchSize)
}

func TestClickHouseConfig(t *testing.T) {
	config := DefaultClickHouseConfig()
	assert.Equal(t, []string{"localhost:9000"}, config.Addr)
	assert.Equal(t, "nem_ts", config.Database)
	assert.Equal(t, "default", config.User)
	assert.Equal(t, "zstd", config.Compression)
	assert.Equal(t, 100, config.MaxOpenConns)
	assert.Equal(t, 20, config.MaxIdleConns)
	assert.Equal(t, 10000, config.BatchSize)
	assert.Equal(t, 65536, config.BlockSize)
}

func TestBatchWriterConfig(t *testing.T) {
	config := DefaultBatchWriterConfig()
	assert.Equal(t, "nem_ts", config.Database)
	assert.Equal(t, "data_points", config.Table)
	assert.Equal(t, 10000, config.BatchSize)
	assert.Equal(t, 5*time.Second, config.FlushTimeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, config.RetryDelay)
}

func TestWriterStats(t *testing.T) {
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
}

func TestRetentionPolicy(t *testing.T) {
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
	assert.True(t, policy.Default)
}

func TestDownsampleQuery(t *testing.T) {
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

func TestTimeSeriesDBType(t *testing.T) {
	assert.Equal(t, TimeSeriesDBType("doris"), TimeSeriesDBTypeDoris)
	assert.Equal(t, TimeSeriesDBType("clickhouse"), TimeSeriesDBTypeClickHouse)
	assert.Equal(t, TimeSeriesDBType("influxdb"), TimeSeriesDBTypeInfluxDB)
	assert.Equal(t, TimeSeriesDBType("timescaledb"), TimeSeriesDBTypeTimescaleDB)
}

func TestIndexSuggestion(t *testing.T) {
	suggestion := IndexSuggestion{
		TableName:  "data_points",
		ColumnName: "point_id",
		IndexType:  "btree",
		Reason:     "frequently used in WHERE clause",
	}

	assert.Equal(t, "data_points", suggestion.TableName)
	assert.Equal(t, "point_id", suggestion.ColumnName)
	assert.Equal(t, "btree", suggestion.IndexType)
}

func TestPoolStats(t *testing.T) {
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
}

// MockTimeSeriesDB 用于测试的模拟时序数据库
type MockTimeSeriesDB struct {
	points []*DataPoint
	closed bool
}

func NewMockTimeSeriesDB() *MockTimeSeriesDB {
	return &MockTimeSeriesDB{
		points: make([]*DataPoint, 0),
	}
}

func (m *MockTimeSeriesDB) Write(ctx context.Context, points []*DataPoint) error {
	m.points = append(m.points, points...)
	return nil
}

func (m *MockTimeSeriesDB) WriteBatch(ctx context.Context, points []*DataPoint) error {
	m.points = append(m.points, points...)
	return nil
}

func (m *MockTimeSeriesDB) WriteWithTable(ctx context.Context, database, table string, points []*DataPoint) error {
	m.points = append(m.points, points...)
	return nil
}

func (m *MockTimeSeriesDB) Query(ctx context.Context, query *Query) (*QueryResult, error) {
	return &QueryResult{Points: m.points, Total: int64(len(m.points))}, nil
}

func (m *MockTimeSeriesDB) QueryRange(ctx context.Context, start, end time.Time, pointIds []int64) ([]*DataPoint, error) {
	return m.points, nil
}

func (m *MockTimeSeriesDB) QueryLatest(ctx context.Context, pointIds []int64) (map[int64]*DataPoint, error) {
	result := make(map[int64]*DataPoint)
	for _, p := range m.points {
		result[p.PointID] = p
	}
	return result, nil
}

func (m *MockTimeSeriesDB) Aggregate(ctx context.Context, query *AggregateQuery) (*AggregateResult, error) {
	return &AggregateResult{}, nil
}

func (m *MockTimeSeriesDB) Downsample(ctx context.Context, query *DownsampleQuery) error {
	return nil
}

func (m *MockTimeSeriesDB) CreateDatabase(ctx context.Context, name string) error {
	return nil
}

func (m *MockTimeSeriesDB) DropDatabase(ctx context.Context, name string) error {
	return nil
}

func (m *MockTimeSeriesDB) CreateTable(ctx context.Context, database, table string, schema *TableSchema) error {
	return nil
}

func (m *MockTimeSeriesDB) DropTable(ctx context.Context, database, table string) error {
	return nil
}

func (m *MockTimeSeriesDB) CreateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error {
	return nil
}

func (m *MockTimeSeriesDB) UpdateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error {
	return nil
}

func (m *MockTimeSeriesDB) DeleteRetentionPolicy(ctx context.Context, database, name string) error {
	return nil
}

func (m *MockTimeSeriesDB) Ping(ctx context.Context) error {
	return nil
}

func (m *MockTimeSeriesDB) IsConnected() bool {
	return !m.closed
}

func (m *MockTimeSeriesDB) Close() error {
	m.closed = true
	return nil
}

func TestMockTimeSeriesDB(t *testing.T) {
	mock := NewMockTimeSeriesDB()

	// Test Write
	points := []*DataPoint{
		{PointID: 1, Value: 10.0, Timestamp: time.Now()},
		{PointID: 2, Value: 20.0, Timestamp: time.Now()},
	}
	err := mock.Write(context.Background(), points)
	assert.NoError(t, err)
	assert.Len(t, mock.points, 2)

	// Test Query
	result, err := mock.Query(context.Background(), &Query{})
	assert.NoError(t, err)
	assert.Len(t, result.Points, 2)

	// Test Close
	err = mock.Close()
	assert.NoError(t, err)
	assert.True(t, mock.closed)
}
