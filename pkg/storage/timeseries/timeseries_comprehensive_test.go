package timeseries

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestClickHouseClient_IsClosed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: false,
	}
	assert.False(t, client.IsClosed())
	assert.True(t, client.IsConnected())

	client.closed = true
	assert.True(t, client.IsClosed())
	assert.False(t, client.IsConnected())
}

func TestClickHouseClient_Write_Closed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	points := []*DataPoint{{PointID: 1, Value: 10.0, Timestamp: time.Now()}}
	err := client.Write(context.Background(), points)
	assert.Equal(t, ErrClosed, err)
}

func TestClickHouseClient_WriteBatch_Closed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	points := []*DataPoint{{PointID: 1, Value: 10.0, Timestamp: time.Now()}}
	err := client.WriteBatch(context.Background(), points)
	assert.Equal(t, ErrClosed, err)
}

func TestClickHouseClient_WriteWithTable_Closed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	points := []*DataPoint{{PointID: 1, Value: 10.0, Timestamp: time.Now()}}
	err := client.WriteWithTable(context.Background(), "db", "table", points)
	assert.Equal(t, ErrClosed, err)
}

func TestClickHouseClient_WriteWithTable_EmptyPoints(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: false,
	}
	err := client.WriteWithTable(context.Background(), "db", "table", []*DataPoint{})
	assert.NoError(t, err)
}

func TestClickHouseClient_Query_Closed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	_, err := client.Query(context.Background(), &Query{})
	assert.Equal(t, ErrClosed, err)
}

func TestClickHouseClient_Query_NilQuery(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: false,
	}
	_, err := client.Query(context.Background(), nil)
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestClickHouseClient_QueryLatest_Closed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	_, err := client.QueryLatest(context.Background(), []int64{1})
	assert.Equal(t, ErrClosed, err)
}

func TestClickHouseClient_QueryLatest_EmptyPointIDs(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: false,
	}
	result, err := client.QueryLatest(context.Background(), []int64{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestClickHouseClient_Aggregate_Closed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	_, err := client.Aggregate(context.Background(), &AggregateQuery{})
	assert.Equal(t, ErrClosed, err)
}

func TestClickHouseClient_Aggregate_NilQuery(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: false,
	}
	_, err := client.Aggregate(context.Background(), nil)
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestClickHouseClient_Downsample_Closed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	err := client.Downsample(context.Background(), &DownsampleQuery{})
	assert.Equal(t, ErrClosed, err)
}

func TestClickHouseClient_CreateDatabase_Closed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	err := client.CreateDatabase(context.Background(), "test_db")
	assert.Equal(t, ErrClosed, err)
}

func TestClickHouseClient_DropDatabase_Closed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	err := client.DropDatabase(context.Background(), "test_db")
	assert.Equal(t, ErrClosed, err)
}

func TestClickHouseClient_CreateTable_Closed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	err := client.CreateTable(context.Background(), "db", "table", &TableSchema{})
	assert.Equal(t, ErrClosed, err)
}

func TestClickHouseClient_DropTable_Closed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	err := client.DropTable(context.Background(), "db", "table")
	assert.Equal(t, ErrClosed, err)
}

func TestClickHouseClient_Ping_Closed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	err := client.Ping(context.Background())
	assert.Equal(t, ErrClosed, err)
}

func TestClickHouseClient_Close_AlreadyClosed(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	err := client.Close()
	assert.NoError(t, err)
}

func TestClickHouseClient_Close_NilConn(t *testing.T) {
	client := &ClickHouseClient{
		config:  DefaultClickHouseConfig(),
		closed:  false,
		conn:    nil,
		logger:  zap.NewNop(),
	}
	err := client.Close()
	assert.NoError(t, err)
	assert.True(t, client.closed)
}

func TestClickHouseClient_CreateRetentionPolicy(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: false,
		logger: zap.NewNop(),
	}
	policy := &RetentionPolicy{
		Name:     "30d",
		Database: "test_db",
		Duration: 30 * 24 * time.Hour,
	}
	err := client.CreateRetentionPolicy(context.Background(), policy)
	assert.NoError(t, err)
}

func TestClickHouseClient_UpdateRetentionPolicy(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: false,
		logger: zap.NewNop(),
	}
	policy := &RetentionPolicy{
		Name:     "30d",
		Database: "test_db",
	}
	err := client.UpdateRetentionPolicy(context.Background(), policy)
	assert.NoError(t, err)
}

func TestClickHouseClient_DeleteRetentionPolicy(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: false,
		logger: zap.NewNop(),
	}
	err := client.DeleteRetentionPolicy(context.Background(), "test_db", "30d")
	assert.NoError(t, err)
}

func TestClickHouseClient_buildQuerySQL(t *testing.T) {
	client := &ClickHouseClient{
		config: &ClickHouseConfig{Database: "test_db"},
	}

	now := time.Now()
	query := &Query{
		Database:  "mydb",
		Table:     "mytable",
		PointIDs:  []int64{1, 2, 3},
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now,
		OrderBy:   "value",
		Order:     "DESC",
		Limit:     100,
		Offset:    10,
	}

	sql, args := client.buildQuerySQL(query)
	assert.Contains(t, sql, "mydb.mytable")
	assert.Contains(t, sql, "ORDER BY value DESC")
	assert.Contains(t, sql, "LIMIT 100 OFFSET 10")
	assert.NotEmpty(t, args)
}

func TestClickHouseClient_buildQuerySQL_DefaultDB(t *testing.T) {
	client := &ClickHouseClient{
		config: &ClickHouseConfig{Database: "default_db"},
	}

	query := &Query{}
	sql, args := client.buildQuerySQL(query)
	assert.Contains(t, sql, "default_db.data_points")
	assert.Contains(t, sql, "ORDER BY timestamp ASC")
	assert.Empty(t, args)
}

func TestClickHouseClient_buildQuerySQL_NoOffset(t *testing.T) {
	client := &ClickHouseClient{
		config: &ClickHouseConfig{Database: "test_db"},
	}

	query := &Query{Limit: 50}
	sql, _ := client.buildQuerySQL(query)
	assert.Contains(t, sql, "LIMIT 50")
	assert.NotContains(t, sql, "OFFSET")
}

func TestClickHouseClient_buildQuerySQL_ASCOrder(t *testing.T) {
	client := &ClickHouseClient{
		config: &ClickHouseConfig{Database: "test_db"},
	}

	query := &Query{OrderBy: "timestamp", Order: "asc"}
	sql, _ := client.buildQuerySQL(query)
	assert.Contains(t, sql, "ORDER BY timestamp ASC")
}

func TestClickHouseClient_buildCountSQL(t *testing.T) {
	client := &ClickHouseClient{
		config: &ClickHouseConfig{Database: "test_db"},
	}

	now := time.Now()
	query := &Query{
		Database:  "mydb",
		Table:     "mytable",
		PointIDs:  []int64{1, 2},
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now,
	}

	sql, args := client.buildCountSQL(query)
	assert.Contains(t, sql, "COUNT(*)")
	assert.Contains(t, sql, "mydb.mytable")
	assert.NotEmpty(t, args)
}

func TestClickHouseClient_buildCountSQL_DefaultDB(t *testing.T) {
	client := &ClickHouseClient{
		config: &ClickHouseConfig{Database: "default_db"},
	}

	query := &Query{}
	sql, args := client.buildCountSQL(query)
	assert.Contains(t, sql, "default_db.data_points")
	assert.Empty(t, args)
}

func TestClickHouseClient_buildAggregateSQL(t *testing.T) {
	client := &ClickHouseClient{
		config: &ClickHouseConfig{Database: "test_db"},
	}

	now := time.Now()
	query := &AggregateQuery{
		Database:  "mydb",
		Table:     "mytable",
		PointIDs:  []int64{1},
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now,
		Interval:  5 * time.Minute,
		AggFunc:   "avg",
	}

	sql, args := client.buildAggregateSQL(query)
	assert.Contains(t, sql, "mydb.mytable")
	assert.Contains(t, sql, "AVG")
	assert.NotEmpty(t, args)
}

func TestClickHouseClient_buildAggregateSQL_DefaultAggFunc(t *testing.T) {
	client := &ClickHouseClient{
		config: &ClickHouseConfig{Database: "test_db"},
	}

	query := &AggregateQuery{
		Interval: time.Minute,
	}
	sql, _ := client.buildAggregateSQL(query)
	assert.Contains(t, sql, "avg")
}

func TestClickHouseClient_intervalToSQL(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig()}

	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30 SECOND"},
		{5 * time.Minute, "5 MINUTE"},
		{2 * time.Hour, "2 HOUR"},
		{3 * 24 * time.Hour, "3 DAY"},
		{0 * time.Second, "0 SECOND"},
	}

	for _, tt := range tests {
		result := client.intervalToSQL(tt.duration)
		assert.Equal(t, tt.expected, result)
	}
}

func TestNewClickHouseClient_NilConfig(t *testing.T) {
	_, err := NewClickHouseClient(nil)
	assert.Error(t, err)
}

func TestNewClickHouseClient_NoAddr(t *testing.T) {
	config := &ClickHouseConfig{Addr: []string{}}
	_, err := NewClickHouseClient(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no addresses")
}

func TestNewClickHouseClient_ConnectionFailed(t *testing.T) {
	config := &ClickHouseConfig{
		Addr:        []string{"localhost:9999"},
		Database:    "test",
		ConnTimeout: 1 * time.Second,
	}
	_, err := NewClickHouseClient(config)
	assert.Error(t, err)
}

func TestClickHouseClient_QueryRange(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	_, err := client.QueryRange(context.Background(), time.Now().Add(-1*time.Hour), time.Now(), []int64{1})
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_IsClosed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	assert.False(t, client.IsClosed())
	assert.True(t, client.IsConnected())

	client.closed = true
	assert.True(t, client.IsClosed())
	assert.False(t, client.IsConnected())
}

func TestDorisClient_Write_Closed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	points := []*DataPoint{{PointID: 1, Value: 10.0, Timestamp: time.Now()}}
	err := client.Write(context.Background(), points)
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_WriteBatch_Closed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	points := []*DataPoint{{PointID: 1, Value: 10.0, Timestamp: time.Now()}}
	err := client.WriteBatch(context.Background(), points)
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_WriteWithTable_Closed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	points := []*DataPoint{{PointID: 1, Value: 10.0, Timestamp: time.Now()}}
	err := client.WriteWithTable(context.Background(), "db", "table", points)
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_WriteWithTable_EmptyPoints(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	err := client.WriteWithTable(context.Background(), "db", "table", []*DataPoint{})
	assert.NoError(t, err)
}

func TestDorisClient_WriteWithTable_InvalidDBName(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	points := []*DataPoint{{PointID: 1, Value: 10.0, Timestamp: time.Now()}}
	err := client.WriteWithTable(context.Background(), "123invalid", "table", points)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestDorisClient_WriteWithTable_InvalidTableName(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	points := []*DataPoint{{PointID: 1, Value: 10.0, Timestamp: time.Now()}}
	err := client.WriteWithTable(context.Background(), "valid_db", "123invalid", points)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")
}

func TestDorisClient_WriteWithTable_NilPoints(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	points := []*DataPoint{nil, nil, nil}
	err := client.WriteWithTable(context.Background(), "valid_db", "valid_table", points)
	assert.NoError(t, err)
}

func TestDorisClient_Query_Closed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	_, err := client.Query(context.Background(), &Query{})
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_Query_NilQuery(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	_, err := client.Query(context.Background(), nil)
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestDorisClient_QueryLatest_Closed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	_, err := client.QueryLatest(context.Background(), []int64{1})
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_QueryLatest_EmptyPointIDs(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	result, err := client.QueryLatest(context.Background(), []int64{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestDorisClient_Aggregate_Closed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	_, err := client.Aggregate(context.Background(), &AggregateQuery{})
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_Aggregate_NilQuery(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	_, err := client.Aggregate(context.Background(), nil)
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestDorisClient_Downsample_Closed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	err := client.Downsample(context.Background(), &DownsampleQuery{})
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_CreateDatabase_Closed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	err := client.CreateDatabase(context.Background(), "test_db")
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_CreateDatabase_InvalidName(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	err := client.CreateDatabase(context.Background(), "123invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestDorisClient_DropDatabase_Closed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	err := client.DropDatabase(context.Background(), "test_db")
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_DropDatabase_InvalidName(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	err := client.DropDatabase(context.Background(), "123invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestDorisClient_CreateTable_Closed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	err := client.CreateTable(context.Background(), "db", "table", &TableSchema{})
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_CreateTable_InvalidDBName(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	err := client.CreateTable(context.Background(), "123invalid", "table", &TableSchema{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestDorisClient_CreateTable_InvalidTableName(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	err := client.CreateTable(context.Background(), "valid_db", "123invalid", &TableSchema{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")
}

func TestDorisClient_CreateTable_InvalidColumnName(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	schema := &TableSchema{
		Columns: []ColumnSchema{
			{Name: "123invalid", Type: "Int64"},
		},
	}
	err := client.CreateTable(context.Background(), "valid_db", "valid_table", schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid column name")
}

func TestDorisClient_DropTable_Closed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	err := client.DropTable(context.Background(), "db", "table")
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_DropTable_InvalidDBName(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	err := client.DropTable(context.Background(), "123invalid", "table")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestDorisClient_Ping_Closed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	err := client.Ping(context.Background())
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_Close_AlreadyClosed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	err := client.Close()
	assert.NoError(t, err)
}

func TestDorisClient_Close_NilDB(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
		db:     nil,
		logger: zap.NewNop(),
	}
	err := client.Close()
	assert.NoError(t, err)
	assert.True(t, client.closed)
}

func TestDorisClient_CreateRetentionPolicy(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
		logger: zap.NewNop(),
	}
	policy := &RetentionPolicy{
		Name:     "30d",
		Database: "test_db",
		Duration: 30 * 24 * time.Hour,
	}
	err := client.CreateRetentionPolicy(context.Background(), policy)
	assert.NoError(t, err)
}

func TestDorisClient_UpdateRetentionPolicy(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
		logger: zap.NewNop(),
	}
	policy := &RetentionPolicy{
		Name:     "30d",
		Database: "test_db",
	}
	err := client.UpdateRetentionPolicy(context.Background(), policy)
	assert.NoError(t, err)
}

func TestDorisClient_DeleteRetentionPolicy(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
		logger: zap.NewNop(),
	}
	err := client.DeleteRetentionPolicy(context.Background(), "test_db", "30d")
	assert.NoError(t, err)
}

func TestDorisClient_buildDSN(t *testing.T) {
	config := &DorisConfig{
		User:        "root",
		Password:    "pass123",
		Port:        9030,
		Database:    "test_db",
		ConnTimeout: 10 * time.Second,
	}
	client := &DorisClient{config: config}

	dsn := client.buildDSN("192.168.1.1")
	assert.Contains(t, dsn, "root:pass123@tcp(192.168.1.1:9030)")
	assert.Contains(t, dsn, "test_db")
}

func TestDorisClient_buildQuerySQL(t *testing.T) {
	client := &DorisClient{
		config: &DorisConfig{Database: "test_db"},
	}

	now := time.Now()
	query := &Query{
		Database:  "mydb",
		Table:     "mytable",
		PointIDs:  []int64{1, 2, 3},
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now,
		OrderBy:   "value",
		Order:     "DESC",
		Limit:     100,
		Offset:    10,
	}

	sql, args := client.buildQuerySQL(query)
	assert.Contains(t, sql, "mydb.mytable")
	assert.Contains(t, sql, "ORDER BY value DESC")
	assert.Contains(t, sql, "LIMIT 100 OFFSET 10")
	assert.NotEmpty(t, args)
}

func TestDorisClient_buildQuerySQL_DefaultDB(t *testing.T) {
	client := &DorisClient{
		config: &DorisConfig{Database: "default_db"},
	}

	query := &Query{}
	sql, args := client.buildQuerySQL(query)
	assert.Contains(t, sql, "default_db.data_points")
	assert.Contains(t, sql, "ORDER BY timestamp ASC")
	assert.Empty(t, args)
}

func TestDorisClient_buildQuerySQL_NoOffset(t *testing.T) {
	client := &DorisClient{
		config: &DorisConfig{Database: "test_db"},
	}

	query := &Query{Limit: 50}
	sql, _ := client.buildQuerySQL(query)
	assert.Contains(t, sql, "LIMIT 50")
	assert.NotContains(t, sql, "OFFSET")
}

func TestDorisClient_buildCountSQL(t *testing.T) {
	client := &DorisClient{
		config: &DorisConfig{Database: "test_db"},
	}

	now := time.Now()
	query := &Query{
		Database:  "mydb",
		Table:     "mytable",
		PointIDs:  []int64{1, 2},
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now,
	}

	sql, args := client.buildCountSQL(query)
	assert.Contains(t, sql, "COUNT(*)")
	assert.Contains(t, sql, "mydb.mytable")
	assert.NotEmpty(t, args)
}

func TestDorisClient_buildCountSQL_DefaultDB(t *testing.T) {
	client := &DorisClient{
		config: &DorisConfig{Database: "default_db"},
	}

	query := &Query{}
	sql, args := client.buildCountSQL(query)
	assert.Contains(t, sql, "default_db.data_points")
	assert.Empty(t, args)
}

func TestDorisClient_buildAggregateSQL(t *testing.T) {
	client := &DorisClient{
		config: &DorisConfig{Database: "test_db"},
	}

	now := time.Now()
	query := &AggregateQuery{
		Database:  "mydb",
		Table:     "mytable",
		PointIDs:  []int64{1},
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now,
		Interval:  5 * time.Minute,
		AggFunc:   "sum",
	}

	sql, args := client.buildAggregateSQL(query)
	assert.Contains(t, sql, "mydb.mytable")
	assert.Contains(t, sql, "SUM")
	assert.NotEmpty(t, args)
}

func TestDorisClient_buildAggregateSQL_DefaultAggFunc(t *testing.T) {
	client := &DorisClient{
		config: &DorisConfig{Database: "test_db"},
	}

	query := &AggregateQuery{
		Interval:  time.Minute,
		StartTime: time.Now(),
	}
	sql, _ := client.buildAggregateSQL(query)
	assert.Contains(t, sql, "AVG")
}

func TestDorisClient_intervalToSQL(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig()}

	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30 SECOND"},
		{5 * time.Minute, "5 MINUTE"},
		{2 * time.Hour, "2 HOUR"},
		{3 * 24 * time.Hour, "3 DAY"},
		{0 * time.Second, "0 SECOND"},
	}

	for _, tt := range tests {
		result := client.intervalToSQL(tt.duration)
		assert.Equal(t, tt.expected, result)
	}
}

func TestDorisClient_serializeTags(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig()}

	result := client.serializeTags(map[string]string{"key1": "val1", "key2": "val2"})
	assert.Contains(t, result, "key1=val1")
	assert.Contains(t, result, "key2=val2")
}

func TestDorisClient_serializeTags_Nil(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig()}

	result := client.serializeTags(nil)
	assert.Equal(t, "", result)
}

func TestDorisClient_serializeTags_Empty(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig()}

	result := client.serializeTags(map[string]string{})
	assert.Equal(t, "", result)
}

func TestDorisClient_parseTags(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig()}

	result := client.parseTags("key1=val1,key2=val2")
	assert.Equal(t, "val1", result["key1"])
	assert.Equal(t, "val2", result["key2"])
}

func TestDorisClient_parseTags_Empty(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig()}

	result := client.parseTags("")
	assert.Empty(t, result)
}

func TestDorisClient_parseTags_NoValue(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig()}

	result := client.parseTags("keyonly")
	assert.Empty(t, result)
}

func TestNewDorisClient_NilConfig(t *testing.T) {
	_, err := NewDorisClient(nil)
	assert.Error(t, err)
}

func TestNewDorisClient_NoHosts(t *testing.T) {
	config := &DorisConfig{Hosts: []string{}}
	_, err := NewDorisClient(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no hosts")
}

func TestNewDorisClient_ConnectionFailed(t *testing.T) {
	config := &DorisConfig{
		Hosts:       []string{"localhost"},
		Port:        9999,
		ConnTimeout: 1 * time.Second,
	}
	_, err := NewDorisClient(config)
	assert.Error(t, err)
}

func TestDorisClient_QueryRange_Closed(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	_, err := client.QueryRange(context.Background(), time.Now().Add(-1*time.Hour), time.Now(), []int64{1})
	assert.Equal(t, ErrClosed, err)
}

func TestDorisClient_QueryLatest_InvalidDBName(t *testing.T) {
	client := &DorisClient{
		config: &DorisConfig{Database: "123invalid"},
		closed: false,
	}
	_, err := client.QueryLatest(context.Background(), []int64{1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

func TestValidateIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		idType     string
		expectErr  bool
	}{
		{"valid", "my_table", "table", false},
		{"valid_with_underscore", "my_table_1", "table", false},
		{"valid_start_underscore", "_table", "table", false},
		{"empty", "", "table", true},
		{"start_with_number", "1table", "table", true},
		{"special_chars", "my-table", "table", true},
		{"too_long", string(make([]byte, 65)), "table", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.identifier == string(make([]byte, 65)) {
				for i := range tt.identifier {
					tt.identifier = tt.identifier[:i] + "a" + tt.identifier[i+1:]
				}
				tt.identifier = ""
				for i := 0; i < 65; i++ {
					tt.identifier += "a"
				}
			}
			err := validateIdentifier(tt.identifier, tt.idType)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateIdentifier_TooLong(t *testing.T) {
	longName := ""
	for i := 0; i < 65; i++ {
		longName += "a"
	}
	err := validateIdentifier(longName, "table")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too long")
}

func TestTimeSeriesFactory_NewTimeSeriesDB_UnsupportedType(t *testing.T) {
	factory := &TimeSeriesFactory{}
	_, err := factory.NewTimeSeriesDB(TimeSeriesDBType("unsupported"), nil)
	assert.Equal(t, ErrUnsupportedDBType, err)
}

func TestTimeSeriesFactory_NewTimeSeriesDB_DorisInvalidConfig(t *testing.T) {
	factory := &TimeSeriesFactory{}
	_, err := factory.NewTimeSeriesDB(TimeSeriesDBTypeDoris, "invalid_config")
	assert.Equal(t, ErrInvalidConfig, err)
}

func TestTimeSeriesFactory_NewTimeSeriesDB_ClickHouseInvalidConfig(t *testing.T) {
	factory := &TimeSeriesFactory{}
	_, err := factory.NewTimeSeriesDB(TimeSeriesDBTypeClickHouse, "invalid_config")
	assert.Equal(t, ErrInvalidConfig, err)
}

func TestQueryError_WithCause(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	queryErr := NewQueryError("SELECT 1", "syntax error", cause)
	assert.Contains(t, queryErr.Error(), "syntax error")
	assert.Contains(t, queryErr.Error(), "SELECT 1")
	assert.Contains(t, queryErr.Error(), "underlying error")
	assert.Equal(t, cause, queryErr.Unwrap())
}

func TestQueryError_WithoutCause(t *testing.T) {
	queryErr := NewQueryError("SELECT 1", "syntax error", nil)
	assert.Contains(t, queryErr.Error(), "syntax error")
	assert.NotContains(t, queryErr.Error(), "cause")
	assert.Nil(t, queryErr.Unwrap())
}

func TestWriteError_WithCause(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	writeErr := NewWriteError(50, "timeout", cause)
	assert.Contains(t, writeErr.Error(), "timeout")
	assert.Contains(t, writeErr.Error(), "50")
	assert.Contains(t, writeErr.Error(), "connection refused")
	assert.Equal(t, cause, writeErr.Unwrap())
}

func TestWriteError_WithoutCause(t *testing.T) {
	writeErr := NewWriteError(50, "timeout", nil)
	assert.Contains(t, writeErr.Error(), "timeout")
	assert.NotContains(t, writeErr.Error(), "cause")
	assert.Nil(t, writeErr.Unwrap())
}

func TestConnectionError_WithCause(t *testing.T) {
	cause := fmt.Errorf("network unreachable")
	connErr := NewConnectionError("localhost:9000", "timeout", cause)
	assert.Contains(t, connErr.Error(), "localhost:9000")
	assert.Contains(t, connErr.Error(), "timeout")
	assert.Contains(t, connErr.Error(), "network unreachable")
	assert.Equal(t, cause, connErr.Unwrap())
}

func TestConnectionError_WithoutCause(t *testing.T) {
	connErr := NewConnectionError("localhost:9000", "timeout", nil)
	assert.Contains(t, connErr.Error(), "localhost:9000")
	assert.NotContains(t, connErr.Error(), "cause")
	assert.Nil(t, connErr.Unwrap())
}

func TestIsRetryableError_ConnectionError(t *testing.T) {
	connErr := NewConnectionError("localhost:9000", "timeout", nil)
	assert.True(t, IsRetryableError(connErr))
}

func TestIsRetryableError_WrappedConnectionError(t *testing.T) {
	connErr := NewConnectionError("localhost:9000", "timeout", nil)
	wrappedErr := fmt.Errorf("wrapped: %w", connErr)
	assert.True(t, IsRetryableError(wrappedErr))
}

func TestIsRetryableError_TimeoutError(t *testing.T) {
	assert.True(t, IsRetryableError(ErrTimeout))
}

func TestIsRetryableError_OtherErrors(t *testing.T) {
	assert.False(t, IsRetryableError(ErrInvalidConfig))
	assert.False(t, IsRetryableError(ErrWriteFailed))
	assert.False(t, IsRetryableError(ErrQueryFailed))
}

func TestIsNotFoundError_All(t *testing.T) {
	assert.True(t, IsNotFoundError(ErrDatabaseNotFound))
	assert.True(t, IsNotFoundError(ErrTableNotFound))
	assert.False(t, IsNotFoundError(ErrConnectionFailed))
	assert.False(t, IsNotFoundError(ErrWriteFailed))
	assert.False(t, IsNotFoundError(nil))
}

func TestBatchWriter_writeWithRetry_EmptyPoints(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := DefaultBatchWriterConfig()
	writer := NewBatchWriter(mock, config)
	defer writer.Close()

	bw := writer.(*timeseriesBatchWriter)
	err := bw.writeWithRetry([]*DataPoint{})
	assert.NoError(t, err)
}

func TestBatchWriter_writeWithRetry_Success(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		Database:     "test_db",
		Table:        "test_table",
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
		MaxRetries:   3,
		RetryDelay:   10 * time.Millisecond,
	}
	writer := NewBatchWriter(mock, config)
	defer writer.Close()

	bw := writer.(*timeseriesBatchWriter)
	points := []*DataPoint{{PointID: 1, Value: 10.0, Timestamp: time.Now()}}
	err := bw.writeWithRetry(points)
	assert.NoError(t, err)

	stats := writer.Stats()
	assert.Equal(t, int64(1), stats.SuccessPoints)
	assert.Equal(t, int64(1), stats.TotalBatches)
}

type failingTimeSeriesDB struct {
	MockTimeSeriesDB
	failCount int
	callCount int
}

func (m *failingTimeSeriesDB) WriteWithTable(ctx context.Context, database, table string, points []*DataPoint) error {
	m.callCount++
	if m.callCount <= m.failCount {
		return NewConnectionError("localhost:9000", "connection failed", nil)
	}
	return nil
}

func TestBatchWriter_writeWithRetry_RetryableError(t *testing.T) {
	mock := &failingTimeSeriesDB{failCount: 2}
	config := &BatchWriterConfig{
		Database:     "test_db",
		Table:        "test_table",
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
		MaxRetries:   3,
		RetryDelay:   10 * time.Millisecond,
	}
	writer := NewBatchWriter(mock, config)
	defer writer.Close()

	bw := writer.(*timeseriesBatchWriter)
	points := []*DataPoint{{PointID: 1, Value: 10.0, Timestamp: time.Now()}}
	err := bw.writeWithRetry(points)
	assert.NoError(t, err)
}

func TestBatchWriter_writeWithRetry_NonRetryableError(t *testing.T) {
	mock := &writeErrorMockDB{}
	config := &BatchWriterConfig{
		Database:     "test_db",
		Table:        "test_table",
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
		MaxRetries:   3,
		RetryDelay:   10 * time.Millisecond,
	}
	writer := NewBatchWriter(mock, config)
	defer writer.Close()

	bw := writer.(*timeseriesBatchWriter)
	points := []*DataPoint{{PointID: 1, Value: 10.0, Timestamp: time.Now()}}
	err := bw.writeWithRetry(points)
	assert.Error(t, err)
}

type writeErrorMockDB struct {
	MockTimeSeriesDB
}

func (m *writeErrorMockDB) WriteWithTable(ctx context.Context, database, table string, points []*DataPoint) error {
	return ErrInvalidConfig
}

func TestBatchWriter_writeWithRetry_MaxRetriesExceeded(t *testing.T) {
	mock := &alwaysFailDB{}
	config := &BatchWriterConfig{
		Database:     "test_db",
		Table:        "test_table",
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
		MaxRetries:   2,
		RetryDelay:   10 * time.Millisecond,
	}
	writer := NewBatchWriter(mock, config)
	defer writer.Close()

	bw := writer.(*timeseriesBatchWriter)
	points := []*DataPoint{{PointID: 1, Value: 10.0, Timestamp: time.Now()}}
	err := bw.writeWithRetry(points)
	assert.Error(t, err)

	stats := writer.Stats()
	assert.Equal(t, int64(1), stats.FailedPoints)
}

type alwaysFailDB struct {
	MockTimeSeriesDB
}

func (m *alwaysFailDB) WriteWithTable(ctx context.Context, database, table string, points []*DataPoint) error {
	return NewConnectionError("localhost:9000", "always fails", nil)
}

func TestAsyncBatchWriter_Add(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		Database:     "test_db",
		Table:        "test_table",
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
	}
	writer := NewAsyncBatchWriter(mock, config)

	point := &DataPoint{
		PointID:   1,
		Timestamp: time.Now(),
		Value:     100.0,
		Quality:   QualityGood,
	}

	resultCh := writer.Add(point)
	err := <-resultCh
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)
}

func TestAsyncBatchWriter_Errors(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		Database:     "test_db",
		Table:        "test_table",
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
	}
	writer := NewAsyncBatchWriter(mock, config)

	errorsCh := writer.Errors()
	assert.NotNil(t, errorsCh)

	writer.Close()
}

func TestAsyncBatchWriter_Stats(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		Database:     "test_db",
		Table:        "test_table",
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
	}
	writer := NewAsyncBatchWriter(mock, config)

	point := &DataPoint{
		PointID:   1,
		Timestamp: time.Now(),
		Value:     100.0,
		Quality:   QualityGood,
	}

	resultCh := writer.Add(point)
	<-resultCh

	stats := writer.Stats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.TotalPoints)

	writer.Close()
}

func TestNewTimeSeriesDBFromConfig_NilConfig(t *testing.T) {
	_, err := NewTimeSeriesDBFromConfig(nil)
	assert.Equal(t, ErrInvalidConfig, err)
}

func TestMustNewTimeSeriesDBFromConfig_Panic(t *testing.T) {
	assert.Panics(t, func() {
		MustNewTimeSeriesDBFromConfig(nil)
	})
}

func TestNewClickHouseClient_DefaultDatabase(t *testing.T) {
	config := &ClickHouseConfig{
		Addr:     []string{"localhost:99999"},
		Database: "",
	}
	_, err := NewClickHouseClient(config)
	assert.Error(t, err)
	if config.Database != "default" {
		t.Logf("expected database to be set to default, got %s", config.Database)
	}
}

func TestClickHouseClient_CreateTable_WithSchema(t *testing.T) {
	schema := &TableSchema{
		Name: "test_table",
		Columns: []ColumnSchema{
			{Name: "point_id", Type: "Int64", Nullable: false},
			{Name: "value", Type: "Float64", Nullable: true, Comment: "data value"},
		},
		SortKey:     []string{"point_id"},
		PartitionBy: "toYYYYMM(timestamp)",
		Engine:      "MergeTree",
	}
	assert.Equal(t, "test_table", schema.Name)
	assert.Len(t, schema.Columns, 2)
}

func TestDorisClient_CreateTable_WithSchema(t *testing.T) {
	schema := &TableSchema{
		Name: "test_table",
		Columns: []ColumnSchema{
			{Name: "point_id", Type: "BIGINT", Nullable: false},
			{Name: "value", Type: "DOUBLE", Nullable: true, Default: "0.0", Comment: "data value"},
		},
		SortKey:     []string{"point_id"},
		PartitionBy: "RANGE(timestamp)",
		Engine:      "DUP",
	}
	assert.Equal(t, "test_table", schema.Name)
	assert.Len(t, schema.Columns, 2)
}

func TestErrors_AllPredefined(t *testing.T) {
	predefinedErrors := []error{
		ErrInvalidConfig,
		ErrUnsupportedDBType,
		ErrConnectionFailed,
		ErrNotConnected,
		ErrDatabaseNotFound,
		ErrTableNotFound,
		ErrQueryFailed,
		ErrWriteFailed,
		ErrBatchTooLarge,
		ErrTimeout,
		ErrInvalidQuery,
		ErrInvalidDataPoint,
		ErrBufferFull,
		ErrClosed,
	}

	for _, err := range predefinedErrors {
		assert.NotNil(t, err)
		assert.NotEmpty(t, err.Error())
	}
}

func TestOptimizedQuery(t *testing.T) {
	oq := &OptimizedQuery{
		OriginalQuery: &Query{Database: "test"},
		OptimizedSQL:  "SELECT * FROM test",
		EstimatedCost: 100,
		UsedIndexes:   []string{"idx_point_id"},
		ExecutionPlan: "FULL SCAN",
	}
	assert.Equal(t, "SELECT * FROM test", oq.OptimizedSQL)
	assert.Equal(t, int64(100), oq.EstimatedCost)
	assert.Len(t, oq.UsedIndexes, 1)
}

func TestColumnSchema(t *testing.T) {
	col := ColumnSchema{
		Name:     "point_id",
		Type:     "Int64",
		Nullable: false,
		Default:  "0",
		Comment:  "point identifier",
	}
	assert.Equal(t, "point_id", col.Name)
	assert.Equal(t, "Int64", col.Type)
	assert.False(t, col.Nullable)
	assert.Equal(t, "0", col.Default)
	assert.Equal(t, "point identifier", col.Comment)
}

func TestWriteBatchRequest(t *testing.T) {
	req := &WriteBatchRequest{
		Database: "test_db",
		Table:    "test_table",
		Points: []*DataPoint{
			{PointID: 1, Value: 10.0, Timestamp: time.Now()},
			{PointID: 2, Value: 20.0, Timestamp: time.Now()},
		},
	}
	assert.Equal(t, "test_db", req.Database)
	assert.Len(t, req.Points, 2)
}

func TestWriteResult(t *testing.T) {
	result := &WriteResult{Success: 100, Failed: 5}
	assert.Equal(t, int64(100), result.Success)
	assert.Equal(t, int64(5), result.Failed)
}

func TestBatchWriterConfig_Defaults(t *testing.T) {
	config := DefaultBatchWriterConfig()
	assert.Equal(t, "nem_ts", config.Database)
	assert.Equal(t, "data_points", config.Table)
	assert.Equal(t, 10000, config.BatchSize)
	assert.Equal(t, 5*time.Second, config.FlushTimeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, config.RetryDelay)
}

func TestBatchWriterConfig_ZeroBatchSize(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{BatchSize: 0}
	writer := NewBatchWriter(mock, config)
	defer writer.Close()
}

func TestBatchWriterConfig_ZeroFlushTimeout(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{FlushTimeout: 0}
	writer := NewBatchWriter(mock, config)
	defer writer.Close()
}

func TestClickHouseClient_QueryRange_CallsQuery(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: true,
	}
	_, err := client.QueryRange(context.Background(), time.Now(), time.Now(), []int64{1, 2})
	assert.Error(t, err)
}

func TestDorisClient_QueryRange_CallsQuery(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: true,
	}
	_, err := client.QueryRange(context.Background(), time.Now(), time.Now(), []int64{1, 2})
	assert.Error(t, err)
}

func TestClickHouseClient_WriteWithTable_NilPoints(t *testing.T) {
	client := &ClickHouseClient{
		config: DefaultClickHouseConfig(),
		closed: false,
	}
	err := client.WriteWithTable(context.Background(), "db", "table", []*DataPoint{})
	assert.NoError(t, err)
}

func TestDorisClient_DropTable_InvalidTableName(t *testing.T) {
	client := &DorisClient{
		config: DefaultDorisConfig(),
		closed: false,
	}
	err := client.DropTable(context.Background(), "valid_db", "123invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")
}

func TestDorisClient_QueryLatest_InvalidDB(t *testing.T) {
	client := &DorisClient{
		config: &DorisConfig{Database: "123invalid"},
		closed: false,
	}
	_, err := client.QueryLatest(context.Background(), []int64{1})
	assert.Error(t, err)
}

func TestNewDorisClient_DefaultPort(t *testing.T) {
	config := &DorisConfig{
		Hosts:       []string{"nonexistent-host"},
		Port:        0,
		ConnTimeout: 1 * time.Second,
	}
	_, err := NewDorisClient(config)
	assert.Error(t, err)
	assert.Equal(t, 9030, config.Port)
}

func TestNewDorisClient_DefaultConns(t *testing.T) {
	config := &DorisConfig{
		Hosts:       []string{"nonexistent-host"},
		ConnTimeout: 1 * time.Second,
	}
	_, err := NewDorisClient(config)
	assert.Error(t, err)
	assert.Equal(t, 100, config.MaxOpenConns)
	assert.Equal(t, 20, config.MaxIdleConns)
	assert.Equal(t, 10000, config.BatchSize)
}

func TestNewClickHouseClient_DefaultConns(t *testing.T) {
	config := &ClickHouseConfig{
		Addr:     []string{"localhost:99999"},
		Database: "test",
	}
	_, err := NewClickHouseClient(config)
	assert.Error(t, err)
	assert.Equal(t, 100, config.MaxOpenConns)
	assert.Equal(t, 20, config.MaxIdleConns)
	assert.Equal(t, 10000, config.BatchSize)
}

func TestClickHouseClient_buildQuerySQL_PointIDsOnly(t *testing.T) {
	client := &ClickHouseClient{
		config: &ClickHouseConfig{Database: "test_db"},
	}
	query := &Query{
		PointIDs: []int64{1, 2},
	}
	sql, args := client.buildQuerySQL(query)
	assert.Contains(t, sql, "point_id IN")
	assert.Len(t, args, 1)
}

func TestClickHouseClient_buildQuerySQL_TimeRangeOnly(t *testing.T) {
	client := &ClickHouseClient{
		config: &ClickHouseConfig{Database: "test_db"},
	}
	now := time.Now()
	query := &Query{
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now,
	}
	sql, args := client.buildQuerySQL(query)
	assert.Contains(t, sql, "timestamp >=")
	assert.Contains(t, sql, "timestamp <=")
	assert.Len(t, args, 2)
}

func TestDorisClient_buildQuerySQL_PointIDsOnly(t *testing.T) {
	client := &DorisClient{
		config: &DorisConfig{Database: "test_db"},
	}
	query := &Query{
		PointIDs: []int64{1, 2},
	}
	sql, args := client.buildQuerySQL(query)
	assert.Contains(t, sql, "point_id IN")
	assert.Len(t, args, 2)
}

func TestDorisClient_buildQuerySQL_TimeRangeOnly(t *testing.T) {
	client := &DorisClient{
		config: &DorisConfig{Database: "test_db"},
	}
	now := time.Now()
	query := &Query{
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now,
	}
	sql, args := client.buildQuerySQL(query)
	assert.Contains(t, sql, "timestamp >=")
	assert.Contains(t, sql, "timestamp <=")
	assert.Len(t, args, 2)
}

func TestDorisClient_buildAggregateSQL_NoPointIDs(t *testing.T) {
	client := &DorisClient{
		config: &DorisConfig{Database: "test_db"},
	}
	query := &AggregateQuery{
		Interval: time.Minute,
	}
	sql, args := client.buildAggregateSQL(query)
	assert.Contains(t, sql, "AVG")
	assert.Empty(t, args)
}

func TestClickHouseClient_buildAggregateSQL_NoPointIDs(t *testing.T) {
	client := &ClickHouseClient{
		config: &ClickHouseConfig{Database: "test_db"},
	}
	query := &AggregateQuery{
		Interval: time.Minute,
	}
	sql, args := client.buildAggregateSQL(query)
	assert.Contains(t, sql, "avg")
	assert.Empty(t, args)
}

func TestBatchWriter_FlushWithMultiplePoints(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := &BatchWriterConfig{
		Database:     "test_db",
		Table:        "test_table",
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
		MaxRetries:   1,
		RetryDelay:   10 * time.Millisecond,
	}
	writer := NewBatchWriter(mock, config)

	for i := 0; i < 5; i++ {
		writer.Add(&DataPoint{
			PointID:   int64(i + 1),
			Timestamp: time.Now(),
			Value:     float64(i * 10),
			Quality:   QualityGood,
		})
	}

	err := writer.Flush()
	assert.NoError(t, err)

	stats := writer.Stats()
	assert.Equal(t, int64(5), stats.TotalPoints)
	assert.Equal(t, int64(5), stats.SuccessPoints)

	writer.Close()
}

func TestBatchWriter_AddNil(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	config := DefaultBatchWriterConfig()
	writer := NewBatchWriter(mock, config)
	defer writer.Close()

	err := writer.Add(nil)
	assert.NoError(t, err)
}

func TestClickHouseConfig_Defaults(t *testing.T) {
	config := DefaultClickHouseConfig()
	assert.Equal(t, []string{"localhost:9000"}, config.Addr)
	assert.Equal(t, "nem_ts", config.Database)
	assert.Equal(t, "default", config.User)
	assert.Equal(t, "", config.Password)
	assert.Equal(t, "zstd", config.Compression)
	assert.Equal(t, 100, config.MaxOpenConns)
	assert.Equal(t, 20, config.MaxIdleConns)
	assert.Equal(t, 10*time.Second, config.ConnTimeout)
	assert.Equal(t, 60*time.Second, config.QueryTimeout)
	assert.Equal(t, 30*time.Second, config.WriteTimeout)
	assert.Equal(t, 10000, config.BatchSize)
	assert.Equal(t, 65536, config.BlockSize)
	assert.False(t, config.Debug)
}

func TestDorisConfig_Defaults(t *testing.T) {
	config := DefaultDorisConfig()
	assert.Equal(t, []string{"localhost"}, config.Hosts)
	assert.Equal(t, 9030, config.Port)
	assert.Equal(t, "nem_ts", config.Database)
	assert.Equal(t, "root", config.User)
	assert.Equal(t, "", config.Password)
	assert.Equal(t, 100, config.MaxOpenConns)
	assert.Equal(t, 20, config.MaxIdleConns)
	assert.Equal(t, 10*time.Second, config.ConnTimeout)
	assert.Equal(t, 30*time.Second, config.WriteTimeout)
	assert.Equal(t, 60*time.Second, config.QueryTimeout)
	assert.Equal(t, 10000, config.BatchSize)
}

func TestValidateIdentifier_ValidNames(t *testing.T) {
	validNames := []string{"my_table", "_table", "Table1", "a"}
	for _, name := range validNames {
		err := validateIdentifier(name, "table")
		assert.NoError(t, err, "expected %s to be valid", name)
	}
}

func TestValidateIdentifier_InvalidNames(t *testing.T) {
	invalidNames := []string{"", "1table", "my-table", "my.table", "my table"}
	for _, name := range invalidNames {
		err := validateIdentifier(name, "table")
		assert.Error(t, err, "expected %s to be invalid", name)
	}
}

func TestIsRetryableError_WriteError(t *testing.T) {
	writeErr := NewWriteError(100, "failed", nil)
	assert.False(t, IsRetryableError(writeErr))
}

func TestIsRetryableError_QueryError(t *testing.T) {
	queryErr := NewQueryError("SELECT 1", "failed", nil)
	assert.False(t, IsRetryableError(queryErr))
}

func TestDorisClient_parseTags_SingleKV(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig()}
	result := client.parseTags("key1=val1")
	assert.Equal(t, "val1", result["key1"])
}

func TestDorisClient_parseTags_WithEqualsInValue(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig()}
	result := client.parseTags("key1=val1=val2")
	assert.Equal(t, "val1=val2", result["key1"])
}

func TestClickHouseClient_buildQuerySQL_WithLimitNoOffset(t *testing.T) {
	client := &ClickHouseClient{config: &ClickHouseConfig{Database: "test_db"}}
	query := &Query{Limit: 10}
	sql, _ := client.buildQuerySQL(query)
	assert.Contains(t, sql, "LIMIT 10")
	assert.NotContains(t, sql, "OFFSET")
}

func TestDorisClient_buildQuerySQL_WithLimitNoOffset(t *testing.T) {
	client := &DorisClient{config: &DorisConfig{Database: "test_db"}}
	query := &Query{Limit: 10}
	sql, _ := client.buildQuerySQL(query)
	assert.Contains(t, sql, "LIMIT 10")
	assert.NotContains(t, sql, "OFFSET")
}

func TestClickHouseClient_buildQuerySQL_WithLimitAndOffset(t *testing.T) {
	client := &ClickHouseClient{config: &ClickHouseConfig{Database: "test_db"}}
	query := &Query{Limit: 10, Offset: 5}
	sql, _ := client.buildQuerySQL(query)
	assert.Contains(t, sql, "LIMIT 10 OFFSET 5")
}

func TestDorisClient_buildQuerySQL_WithLimitAndOffset(t *testing.T) {
	client := &DorisClient{config: &DorisConfig{Database: "test_db"}}
	query := &Query{Limit: 10, Offset: 5}
	sql, _ := client.buildQuerySQL(query)
	assert.Contains(t, sql, "LIMIT 10 OFFSET 5")
}

func TestClickHouseClient_buildAggregateSQL_WithPointIDs(t *testing.T) {
	client := &ClickHouseClient{config: &ClickHouseConfig{Database: "test_db"}}
	query := &AggregateQuery{
		PointIDs: []int64{1, 2, 3},
		Interval: time.Minute,
	}
	sql, args := client.buildAggregateSQL(query)
	assert.Contains(t, sql, "point_id IN")
	assert.Len(t, args, 1)
}

func TestDorisClient_buildAggregateSQL_WithPointIDs(t *testing.T) {
	client := &DorisClient{config: &DorisConfig{Database: "test_db"}}
	query := &AggregateQuery{
		PointIDs: []int64{1, 2, 3},
		Interval: time.Minute,
	}
	sql, args := client.buildAggregateSQL(query)
	assert.Contains(t, sql, "point_id IN")
	assert.Len(t, args, 3)
}

func TestDorisClient_buildCountSQL_WithPointIDs(t *testing.T) {
	client := &DorisClient{config: &DorisConfig{Database: "test_db"}}
	query := &Query{PointIDs: []int64{1, 2}}
	sql, args := client.buildCountSQL(query)
	assert.Contains(t, sql, "point_id IN")
	assert.Len(t, args, 2)
}

func TestDorisClient_buildCountSQL_WithTimeRange(t *testing.T) {
	client := &DorisClient{config: &DorisConfig{Database: "test_db"}}
	now := time.Now()
	query := &Query{
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now,
	}
	sql, args := client.buildCountSQL(query)
	assert.Contains(t, sql, "timestamp >=")
	assert.Contains(t, sql, "timestamp <=")
	assert.Len(t, args, 2)
}

func TestClickHouseClient_buildCountSQL_WithPointIDs(t *testing.T) {
	client := &ClickHouseClient{config: &ClickHouseConfig{Database: "test_db"}}
	query := &Query{PointIDs: []int64{1, 2}}
	sql, args := client.buildCountSQL(query)
	assert.Contains(t, sql, "point_id IN")
	assert.Len(t, args, 1)
}

func TestClickHouseClient_buildCountSQL_WithTimeRange(t *testing.T) {
	client := &ClickHouseClient{config: &ClickHouseConfig{Database: "test_db"}}
	now := time.Now()
	query := &Query{
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now,
	}
	sql, args := client.buildCountSQL(query)
	assert.Contains(t, sql, "timestamp >=")
	assert.Contains(t, sql, "timestamp <=")
	assert.Len(t, args, 2)
}

func TestErrors_IsRetryable(t *testing.T) {
	require.True(t, IsRetryableError(ErrTimeout))
	require.True(t, IsRetryableError(NewConnectionError("host", "msg", nil)))
	require.False(t, IsRetryableError(ErrClosed))
	require.False(t, IsRetryableError(ErrInvalidConfig))
	require.False(t, IsRetryableError(errors.New("some error")))
}
