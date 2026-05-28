package timeseries

import (
	"context"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestComp_DorisClient_ClosedAll(t *testing.T) {
	client := &DorisClient{closed: true, logger: zap.L().Named("doris"), config: &DorisConfig{Database: "test"}}

	err := client.Write(context.Background(), []*DataPoint{{PointID: 1}})
	assert.Equal(t, ErrClosed, err)

	err = client.WriteBatch(context.Background(), []*DataPoint{{PointID: 1}})
	assert.Equal(t, ErrClosed, err)

	_, err = client.Query(context.Background(), &Query{Database: "db", Table: "tbl"})
	assert.Equal(t, ErrClosed, err)

	_, err = client.QueryRange(context.Background(), time.Now(), time.Now(), []int64{1})
	assert.Equal(t, ErrClosed, err)

	_, err = client.Aggregate(context.Background(), &AggregateQuery{Database: "db", Table: "tbl"})
	assert.Equal(t, ErrClosed, err)

	err = client.CreateDatabase(context.Background(), "db")
	assert.Equal(t, ErrClosed, err)

	err = client.DropDatabase(context.Background(), "db")
	assert.Equal(t, ErrClosed, err)

	err = client.CreateTable(context.Background(), "db", "tbl", &TableSchema{})
	assert.Equal(t, ErrClosed, err)

	err = client.DropTable(context.Background(), "db", "tbl")
	assert.Equal(t, ErrClosed, err)

	err = client.Ping(context.Background())
	assert.Equal(t, ErrClosed, err)

	_, err = client.QueryLatest(context.Background(), []int64{1})
	assert.Equal(t, ErrClosed, err)

	err = client.Downsample(context.Background(), nil)
	assert.Equal(t, ErrClosed, err)
}

func TestComp_DorisClient_QueryNil(t *testing.T) {
	client := &DorisClient{logger: zap.L().Named("doris"), config: &DorisConfig{Database: "test"}}
	_, err := client.Query(context.Background(), nil)
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestComp_DorisClient_AggregateNil(t *testing.T) {
	client := &DorisClient{logger: zap.L().Named("doris"), config: &DorisConfig{Database: "test"}}
	_, err := client.Aggregate(context.Background(), nil)
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestComp_DorisClient_WriteWithTable_EmptyPoints(t *testing.T) {
	client := &DorisClient{logger: zap.L().Named("doris"), config: &DorisConfig{Database: "test"}}
	err := client.WriteWithTable(context.Background(), "db", "tbl", []*DataPoint{})
	assert.NoError(t, err)
}

func TestComp_DorisClient_WriteWithTable_InvalidDB(t *testing.T) {
	client := &DorisClient{logger: zap.L().Named("doris"), config: &DorisConfig{Database: "test"}}
	err := client.WriteWithTable(context.Background(), "123invalid", "tbl", []*DataPoint{{PointID: 1}})
	assert.Error(t, err)
}

func TestComp_DorisClient_NewDorisClient_EmptyHosts(t *testing.T) {
	cfg := &DorisConfig{
		Hosts:    []string{},
		Database: "test_db",
		User:     "root",
	}
	_, err := NewDorisClient(cfg)
	assert.Error(t, err)
}

func TestComp_DorisClient_SerializeTags(t *testing.T) {
	client := &DorisClient{logger: zap.L().Named("doris"), config: &DorisConfig{Database: "test"}}
	assert.Equal(t, "", client.serializeTags(nil))
	assert.Equal(t, "", client.serializeTags(map[string]string{}))
	result := client.serializeTags(map[string]string{"key1": "val1", "key2": "val2"})
	assert.Contains(t, result, "key1=val1")
	assert.Contains(t, result, "key2=val2")
}

func TestComp_DorisClient_QueryLatest_EmptyIDs(t *testing.T) {
	client := &DorisClient{logger: zap.L().Named("doris"), config: &DorisConfig{Database: "test"}}
	result, err := client.QueryLatest(context.Background(), []int64{})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestComp_DorisClient_Close(t *testing.T) {
	client := &DorisClient{logger: zap.L().Named("doris"), config: &DorisConfig{Database: "test"}}
	err := client.Close()
	assert.NoError(t, err)
	assert.True(t, client.closed)
}

func TestComp_ClickHouseClient_ClosedAll(t *testing.T) {
	client := &ClickHouseClient{closed: true, logger: zap.L().Named("ch"), config: &ClickHouseConfig{Database: "test"}}

	err := client.Write(context.Background(), []*DataPoint{{PointID: 1}})
	assert.Equal(t, ErrClosed, err)

	err = client.WriteBatch(context.Background(), []*DataPoint{{PointID: 1}})
	assert.Equal(t, ErrClosed, err)

	_, err = client.Query(context.Background(), &Query{Database: "db", Table: "tbl"})
	assert.Equal(t, ErrClosed, err)

	_, err = client.QueryRange(context.Background(), time.Now(), time.Now(), []int64{1})
	assert.Equal(t, ErrClosed, err)

	_, err = client.Aggregate(context.Background(), &AggregateQuery{Database: "db", Table: "tbl"})
	assert.Equal(t, ErrClosed, err)

	err = client.CreateDatabase(context.Background(), "db")
	assert.Equal(t, ErrClosed, err)

	err = client.DropDatabase(context.Background(), "db")
	assert.Equal(t, ErrClosed, err)

	err = client.CreateTable(context.Background(), "db", "tbl", &TableSchema{})
	assert.Equal(t, ErrClosed, err)

	err = client.DropTable(context.Background(), "db", "tbl")
	assert.Equal(t, ErrClosed, err)

	err = client.Ping(context.Background())
	assert.Equal(t, ErrClosed, err)

	_, err = client.QueryLatest(context.Background(), []int64{1})
	assert.Equal(t, ErrClosed, err)

	err = client.Downsample(context.Background(), nil)
	assert.Equal(t, ErrClosed, err)
}

func TestComp_ClickHouseClient_QueryNil(t *testing.T) {
	client := &ClickHouseClient{logger: zap.L().Named("ch"), config: &ClickHouseConfig{Database: "test"}}
	_, err := client.Query(context.Background(), nil)
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestComp_ClickHouseClient_AggregateNil(t *testing.T) {
	client := &ClickHouseClient{logger: zap.L().Named("ch"), config: &ClickHouseConfig{Database: "test"}}
	_, err := client.Aggregate(context.Background(), nil)
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestComp_ClickHouseClient_WriteWithTable_EmptyPoints(t *testing.T) {
	client := &ClickHouseClient{logger: zap.L().Named("ch"), config: &ClickHouseConfig{Database: "test"}}
	err := client.WriteWithTable(context.Background(), "db", "tbl", []*DataPoint{})
	assert.NoError(t, err)
}

func TestComp_ClickHouseClient_Close(t *testing.T) {
	client := &ClickHouseClient{logger: zap.L().Named("ch"), config: &ClickHouseConfig{Database: "test"}}
	err := client.Close()
	assert.NoError(t, err)
	assert.True(t, client.closed)
}

func TestComp_BatchWriter_FlushWithRetry(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	cfg := &BatchWriterConfig{
		Database:     "test_db",
		Table:        "test_table",
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
		MaxRetries:   3,
		RetryDelay:   10 * time.Millisecond,
	}

	writer := NewBatchWriter(mock, cfg)
	defer writer.Close()

	point := &DataPoint{
		PointID:   1,
		Timestamp: time.Now(),
		Value:     100.5,
		Quality:   QualityGood,
	}

	err := writer.Add(point)
	assert.NoError(t, err)

	err = writer.Flush()
	assert.NoError(t, err)

	stats := writer.Stats()
	assert.Equal(t, int64(1), stats.TotalPoints)
	assert.Equal(t, int64(1), stats.SuccessPoints)
}

func TestComp_AsyncBatchWriter(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	cfg := &BatchWriterConfig{
		Database:     "test_db",
		Table:        "test_table",
		BatchSize:    5,
		FlushTimeout: 200 * time.Millisecond,
		MaxRetries:   3,
		RetryDelay:   10 * time.Millisecond,
	}

	writer := NewAsyncBatchWriter(mock, cfg)

	for i := 0; i < 10; i++ {
		point := &DataPoint{
			PointID:   int64(i + 1),
			Timestamp: time.Now(),
			Value:     float64(i * 10),
			Quality:   QualityGood,
		}
		errCh := writer.Add(point)
		assert.NotNil(t, errCh)
	}

	time.Sleep(500 * time.Millisecond)

	errCh := writer.Errors()
	assert.NotNil(t, errCh)

	stats := writer.Stats()
	assert.NotNil(t, stats)

	err := writer.Close()
	assert.NoError(t, err)
}

func TestComp_AsyncBatchWriter_AddAfterClose(t *testing.T) {
	mock := NewMockTimeSeriesDB()
	cfg := &BatchWriterConfig{
		Database:     "test_db",
		Table:        "test_table",
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
	}

	writer := NewAsyncBatchWriter(mock, cfg)
	writer.Close()

	point := &DataPoint{
		PointID:   1,
		Timestamp: time.Now(),
		Value:     100.0,
		Quality:   QualityGood,
	}
	errCh := writer.Add(point)
	select {
	case err := <-errCh:
		assert.Error(t, err)
	default:
	}
}

func TestComp_QueryError_Unwrap(t *testing.T) {
	err := NewQueryError("SELECT *", "test query error", ErrInvalidQuery)
	assert.Contains(t, err.Error(), "test query error")
	assert.Equal(t, ErrInvalidQuery, err.Unwrap())
}

func TestComp_WriteError_Unwrap(t *testing.T) {
	err := NewWriteError(1, "test write error", ErrClosed)
	assert.Contains(t, err.Error(), "test write error")
	assert.Equal(t, ErrClosed, err.Unwrap())
}

func TestComp_ConnectionError_Unwrap(t *testing.T) {
	err := NewConnectionError("localhost", "test connection error", ErrClosed)
	assert.Contains(t, err.Error(), "test connection error")
	assert.Equal(t, ErrClosed, err.Unwrap())
}

func TestComp_IsRetryableError_Nil(t *testing.T) {
	assert.False(t, IsRetryableError(nil))
}

func TestComp_IsNotFoundError_Nil(t *testing.T) {
	assert.False(t, IsNotFoundError(nil))
}

func TestComp_ValidateIdentifier(t *testing.T) {
	assert.NoError(t, validateIdentifier("valid_name", "table"))
	assert.NoError(t, validateIdentifier("valid_name_123", "table"))
	assert.Error(t, validateIdentifier("123invalid", "table"))
	assert.Error(t, validateIdentifier("", "table"))
	assert.Error(t, validateIdentifier("invalid-name", "table"))
	assert.Error(t, validateIdentifier("invalid name", "table"))
}

func TestComp_NewTimeSeriesDBFromConfig_Nil(t *testing.T) {
	_, err := NewTimeSeriesDBFromConfig(nil)
	assert.Error(t, err)
}

func TestComp_NewTimeSeriesDBFromConfig_Unsupported(t *testing.T) {
	cfg := &config.TimeSeriesConfig{Type: "unsupported"}
	_, err := NewTimeSeriesDBFromConfig(cfg)
	assert.Error(t, err)
}
