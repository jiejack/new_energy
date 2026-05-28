package timeseries

import (
	"context"
	"database/sql"
	sqldriver "database/sql/driver"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	internalconfig "github.com/new-energy-monitoring/internal/infrastructure/config"
)

type mockTimeSeriesDB struct {
	writeErr    error
	writeFn     func(ctx context.Context, database, table string, points []*DataPoint) error
	closed      bool
	writeCalled int
}

func (m *mockTimeSeriesDB) Write(ctx context.Context, points []*DataPoint) error {
	return m.WriteWithTable(ctx, "default", "data_points", points)
}

func (m *mockTimeSeriesDB) WriteBatch(ctx context.Context, points []*DataPoint) error {
	return m.WriteWithTable(ctx, "default", "data_points", points)
}

func (m *mockTimeSeriesDB) WriteWithTable(ctx context.Context, database, table string, points []*DataPoint) error {
	m.writeCalled++
	if m.writeFn != nil {
		return m.writeFn(ctx, database, table, points)
	}
	return m.writeErr
}

func (m *mockTimeSeriesDB) Query(ctx context.Context, query *Query) (*QueryResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockTimeSeriesDB) QueryRange(ctx context.Context, start, end time.Time, pointIds []int64) ([]*DataPoint, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockTimeSeriesDB) QueryLatest(ctx context.Context, pointIds []int64) (map[int64]*DataPoint, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockTimeSeriesDB) Aggregate(ctx context.Context, query *AggregateQuery) (*AggregateResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockTimeSeriesDB) Downsample(ctx context.Context, query *DownsampleQuery) error {
	return fmt.Errorf("not implemented")
}

func (m *mockTimeSeriesDB) CreateDatabase(ctx context.Context, name string) error { return nil }
func (m *mockTimeSeriesDB) DropDatabase(ctx context.Context, name string) error    { return nil }
func (m *mockTimeSeriesDB) CreateTable(ctx context.Context, database, table string, schema *TableSchema) error {
	return nil
}
func (m *mockTimeSeriesDB) DropTable(ctx context.Context, database, table string) error { return nil }
func (m *mockTimeSeriesDB) CreateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error {
	return nil
}
func (m *mockTimeSeriesDB) UpdateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error {
	return nil
}
func (m *mockTimeSeriesDB) DeleteRetentionPolicy(ctx context.Context, database, name string) error {
	return nil
}
func (m *mockTimeSeriesDB) Ping(ctx context.Context) error   { return nil }
func (m *mockTimeSeriesDB) IsConnected() bool                { return !m.closed }
func (m *mockTimeSeriesDB) Close() error                     { m.closed = true; return nil }

type mockCHBatch struct {
	appendErr error
	sendErr   error
	appended  int
}

func (m *mockCHBatch) Abort() error                                                      { return nil }
func (m *mockCHBatch) Append(v ...interface{}) error                                     { m.appended++; return m.appendErr }
func (m *mockCHBatch) AppendStruct(v interface{}) error                                  { return nil }
func (m *mockCHBatch) Column(i int) driver.BatchColumn                                   { return nil }
func (m *mockCHBatch) Flush() error                                                      { return nil }
func (m *mockCHBatch) Send() error                                                       { return m.sendErr }
func (m *mockCHBatch) IsSent() bool                                                      { return false }
func (m *mockCHBatch) Rows() int                                                         { return m.appended }
func (m *mockCHBatch) Columns() []column.Interface                                       { return nil }
func (m *mockCHBatch) Close() error                                                      { return nil }

type mockCHRows struct {
	nextFn  func() bool
	scanFn  func(dest ...interface{}) error
	closed  bool
}

func (m *mockCHRows) Next() bool                           { return m.nextFn() }
func (m *mockCHRows) Scan(dest ...interface{}) error       { return m.scanFn(dest...) }
func (m *mockCHRows) ScanStruct(dest interface{}) error    { return nil }
func (m *mockCHRows) ColumnTypes() []driver.ColumnType        { return nil }
func (m *mockCHRows) Totals(dest ...interface{}) error     { return nil }
func (m *mockCHRows) Columns() []string                    { return nil }
func (m *mockCHRows) Close() error                         { m.closed = true; return nil }
func (m *mockCHRows) Err() error                           { return nil }
func (m *mockCHRows) HasData() bool                        { return false }

type mockCHConn struct {
	prepareBatchFn func(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error)
	queryFn        func(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
	execFn         func(ctx context.Context, query string, args ...interface{}) error
	pingFn         func(ctx context.Context) error
	closeFn        func() error
}

func (m *mockCHConn) Contributors() []string                                                { return nil }
func (m *mockCHConn) ServerVersion() (*driver.ServerVersion, error)                         { return nil, nil }
func (m *mockCHConn) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}
func (m *mockCHConn) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	if m.queryFn != nil {
		return m.queryFn(ctx, query, args...)
	}
	return nil, fmt.Errorf("not implemented")
}
func (m *mockCHConn) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	return nil
}
func (m *mockCHConn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	if m.prepareBatchFn != nil {
		return m.prepareBatchFn(ctx, query, opts...)
	}
	return nil, fmt.Errorf("not implemented")
}
func (m *mockCHConn) Exec(ctx context.Context, query string, args ...interface{}) error {
	if m.execFn != nil {
		return m.execFn(ctx, query, args...)
	}
	return nil
}
func (m *mockCHConn) AsyncInsert(ctx context.Context, query string, wait bool, args ...interface{}) error {
	return nil
}
func (m *mockCHConn) Ping(ctx context.Context) error {
	if m.pingFn != nil {
		return m.pingFn(ctx)
	}
	return nil
}
func (m *mockCHConn) Stats() driver.Stats { return driver.Stats{} }
func (m *mockCHConn) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

var registeredDrivers sync.Map

func registerMockDriverN(name string, numCols int) {
	if _, loaded := registeredDrivers.LoadOrStore(name, true); loaded {
		return
	}
	sql.Register(name, &mockSQLDriverN{numCols: numCols})
}

type mockSQLDriverN struct {
	numCols int
}

func (d *mockSQLDriverN) Open(name string) (sqldriver.Conn, error) {
	return &mockSQLConnN{numCols: d.numCols}, nil
}

type mockSQLConnN struct {
	numCols int
}

func (c *mockSQLConnN) Prepare(query string) (sqldriver.Stmt, error) { return &mockSQLStmt{}, nil }
func (c *mockSQLConnN) Close() error                                { return nil }
func (c *mockSQLConnN) Begin() (sqldriver.Tx, error)                { return nil, nil }
func (c *mockSQLConnN) ExecContext(ctx context.Context, query string, args []sqldriver.NamedValue) (sqldriver.Result, error) {
	return sqldriver.RowsAffected(1), nil
}
func (c *mockSQLConnN) QueryContext(ctx context.Context, query string, args []sqldriver.NamedValue) (sqldriver.Rows, error) {
	return &mockSQLRowsN{numCols: c.numCols}, nil
}

type mockSQLStmt struct{}

func (s *mockSQLStmt) Close() error                                          { return nil }
func (s *mockSQLStmt) NumInput() int                                         { return -1 }
func (s *mockSQLStmt) Exec(args []sqldriver.Value) (sqldriver.Result, error) { return sqldriver.RowsAffected(1), nil }
func (s *mockSQLStmt) Query(args []sqldriver.Value) (sqldriver.Rows, error)  { return nil, nil }

type mockSQLRowsN struct {
	numCols int
	rowIdx  int
}

func (r *mockSQLRowsN) Columns() []string {
	cols := make([]string, r.numCols)
	for i := range cols {
		cols[i] = fmt.Sprintf("col_%d", i)
	}
	return cols
}
func (r *mockSQLRowsN) Close() error { return nil }
func (r *mockSQLRowsN) Next(dest []sqldriver.Value) error {
	if r.rowIdx > 0 {
		return io.EOF
	}
	r.rowIdx++
	for i := range dest {
		switch i {
		case 0:
			dest[i] = int64(1)
		case 1:
			dest[i] = time.Now()
		case 2:
			dest[i] = 80.0
		case 3:
			dest[i] = int64(QualityGood)
		case 4:
			dest[i] = ""
		case 5:
			dest[i] = int64(1)
		default:
			dest[i] = ""
		}
	}
	return nil
}

func newMockSQLDB() *sql.DB {
	return newMockSQLDBWithCols(6)
}

func newMockSQLDB5Cols() *sql.DB {
	return newMockSQLDBWithCols(5)
}

func newMockSQLDBWithCols(numCols int) *sql.DB {
	driverName := fmt.Sprintf("mocksql_%dcols", numCols)
	registerMockDriverN(driverName, numCols)
	db, _ := sql.Open(driverName, "mock")
	return db
}

func TestCov_CH_WriteWithTable_Success(t *testing.T) {
	mockConn := &mockCHConn{
		prepareBatchFn: func(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
			return &mockCHBatch{}, nil
		},
	}
	client := &ClickHouseClient{
		config:  DefaultClickHouseConfig(),
		logger:  zap.NewNop(),
		conn:    mockConn,
		closed:  false,
	}
	points := []*DataPoint{{PointID: 1, Value: 80.0, Timestamp: time.Now()}}
	err := client.WriteWithTable(context.Background(), "db", "metrics", points)
	assert.NoError(t, err)
}

func TestCov_CH_WriteWithTable_PrepareError(t *testing.T) {
	mockConn := &mockCHConn{
		prepareBatchFn: func(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
			return nil, fmt.Errorf("prepare failed")
		},
	}
	client := &ClickHouseClient{
		config:  DefaultClickHouseConfig(),
		logger:  zap.NewNop(),
		conn:    mockConn,
		closed:  false,
	}
	points := []*DataPoint{{PointID: 1, Value: 80.0, Timestamp: time.Now()}}
	err := client.WriteWithTable(context.Background(), "db", "metrics", points)
	assert.Error(t, err)
}

func TestCov_CH_WriteWithTable_AppendError(t *testing.T) {
	mockConn := &mockCHConn{
		prepareBatchFn: func(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
			return &mockCHBatch{appendErr: fmt.Errorf("append failed")}, nil
		},
	}
	client := &ClickHouseClient{
		config:  DefaultClickHouseConfig(),
		logger:  zap.NewNop(),
		conn:    mockConn,
		closed:  false,
	}
	points := []*DataPoint{{PointID: 1, Value: 80.0, Timestamp: time.Now()}}
	err := client.WriteWithTable(context.Background(), "db", "metrics", points)
	assert.Error(t, err)
}

func TestCov_CH_WriteWithTable_SendError(t *testing.T) {
	mockConn := &mockCHConn{
		prepareBatchFn: func(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
			return &mockCHBatch{sendErr: fmt.Errorf("send failed")}, nil
		},
	}
	client := &ClickHouseClient{
		config:  DefaultClickHouseConfig(),
		logger:  zap.NewNop(),
		conn:    mockConn,
		closed:  false,
	}
	points := []*DataPoint{{PointID: 1, Value: 80.0, Timestamp: time.Now()}}
	err := client.WriteWithTable(context.Background(), "db", "metrics", points)
	assert.Error(t, err)
}

func TestCov_CH_WriteWithTable_Closed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	err := client.WriteWithTable(context.Background(), "db", "metrics", []*DataPoint{{PointID: 1, Value: 80.0, Timestamp: time.Now()}})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_CH_WriteWithTable_EmptyPoints(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	err := client.WriteWithTable(context.Background(), "db", "metrics", []*DataPoint{})
	assert.NoError(t, err)
}

func TestCov_CH_Write_Closed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	err := client.Write(context.Background(), []*DataPoint{{PointID: 1, Value: 80.0, Timestamp: time.Now()}})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_CH_WriteBatch_Closed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	err := client.WriteBatch(context.Background(), []*DataPoint{{PointID: 1, Value: 80.0, Timestamp: time.Now()}})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_CH_Query_Success(t *testing.T) {
	mockConn := &mockCHConn{
		queryFn: func(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
			return &mockCHRows{
				nextFn: func() bool { return false },
				scanFn: func(dest ...interface{}) error { return nil },
			}, nil
		},
	}
	client := &ClickHouseClient{
		config:  DefaultClickHouseConfig(),
		logger:  zap.NewNop(),
		conn:    mockConn,
		closed:  false,
	}
	q := &Query{Database: "testdb", Table: "metrics", PointIDs: []int64{1}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now()}
	result, err := client.Query(context.Background(), q)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCov_CH_Query_QueryError(t *testing.T) {
	mockConn := &mockCHConn{
		queryFn: func(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
			return nil, fmt.Errorf("query failed")
		},
	}
	client := &ClickHouseClient{
		config:  DefaultClickHouseConfig(),
		logger:  zap.NewNop(),
		conn:    mockConn,
		closed:  false,
	}
	q := &Query{Database: "testdb", Table: "metrics", PointIDs: []int64{1}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now()}
	_, err := client.Query(context.Background(), q)
	assert.Error(t, err)
}

func TestCov_CH_Query_Closed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	_, err := client.Query(context.Background(), &Query{})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_CH_Query_NilQuery(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	_, err := client.Query(context.Background(), nil)
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestCov_CH_QueryLatest_Success(t *testing.T) {
	mockConn := &mockCHConn{
		queryFn: func(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
			return &mockCHRows{
				nextFn: func() bool { return false },
				scanFn: func(dest ...interface{}) error { return nil },
			}, nil
		},
	}
	client := &ClickHouseClient{
		config:  DefaultClickHouseConfig(),
		logger:  zap.NewNop(),
		conn:    mockConn,
		closed:  false,
	}
	result, err := client.QueryLatest(context.Background(), []int64{1, 2})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCov_CH_QueryLatest_Closed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	_, err := client.QueryLatest(context.Background(), []int64{1})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_CH_QueryLatest_EmptyPointIDs(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	result, err := client.QueryLatest(context.Background(), []int64{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestCov_CH_Aggregate_Success(t *testing.T) {
	mockConn := &mockCHConn{
		queryFn: func(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
			return &mockCHRows{
				nextFn: func() bool { return false },
				scanFn: func(dest ...interface{}) error { return nil },
			}, nil
		},
	}
	client := &ClickHouseClient{
		config:  DefaultClickHouseConfig(),
		logger:  zap.NewNop(),
		conn:    mockConn,
		closed:  false,
	}
	q := &AggregateQuery{Database: "testdb", Table: "metrics", PointIDs: []int64{1}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now(), Interval: 5 * time.Minute, AggFunc: "avg"}
	_, err := client.Aggregate(context.Background(), q)
	assert.NoError(t, err)
}

func TestCov_CH_Aggregate_Closed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	_, err := client.Aggregate(context.Background(), &AggregateQuery{})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_CH_Aggregate_NilQuery(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	_, err := client.Aggregate(context.Background(), nil)
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestCov_CH_Downsample_Success(t *testing.T) {
	mockConn := &mockCHConn{
		queryFn: func(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
			return &mockCHRows{nextFn: func() bool { return false }, scanFn: func(dest ...interface{}) error { return nil }}, nil
		},
		execFn: func(ctx context.Context, query string, args ...interface{}) error {
			return nil
		},
	}
	client := &ClickHouseClient{
		config:  DefaultClickHouseConfig(),
		logger:  zap.NewNop(),
		conn:    mockConn,
		closed:  false,
	}
	err := client.Downsample(context.Background(), &DownsampleQuery{
		SourceDatabase: "testdb", SourceTable: "metrics", TargetDatabase: "testdb", TargetTable: "metrics_down",
		PointIDs: []int64{1}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now(),
		Interval: 5 * time.Minute, AggFunc: "avg",
	})
	assert.NoError(t, err)
}

func TestCov_CH_Downsample_Error(t *testing.T) {
	mockConn := &mockCHConn{
		queryFn: func(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
			return nil, fmt.Errorf("aggregate query failed")
		},
	}
	client := &ClickHouseClient{
		config:  DefaultClickHouseConfig(),
		logger:  zap.NewNop(),
		conn:    mockConn,
		closed:  false,
	}
	err := client.Downsample(context.Background(), &DownsampleQuery{
		SourceDatabase: "testdb", SourceTable: "metrics", TargetDatabase: "testdb", TargetTable: "metrics_down",
		PointIDs: []int64{1}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now(),
		Interval: 5 * time.Minute, AggFunc: "avg",
	})
	assert.Error(t, err)
}

func TestCov_CH_Downsample_Closed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	err := client.Downsample(context.Background(), &DownsampleQuery{})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_CH_CreateDatabase_Success(t *testing.T) {
	mockConn := &mockCHConn{
		execFn: func(ctx context.Context, query string, args ...interface{}) error { return nil },
	}
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), conn: mockConn, closed: false}
	err := client.CreateDatabase(context.Background(), "newdb")
	assert.NoError(t, err)
}

func TestCov_CH_CreateDatabase_Error(t *testing.T) {
	mockConn := &mockCHConn{
		execFn: func(ctx context.Context, query string, args ...interface{}) error {
			return fmt.Errorf("create db failed")
		},
	}
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), conn: mockConn, closed: false}
	err := client.CreateDatabase(context.Background(), "newdb")
	assert.Error(t, err)
}

func TestCov_CH_CreateDatabase_Closed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	err := client.CreateDatabase(context.Background(), "newdb")
	assert.Equal(t, ErrClosed, err)
}

func TestCov_CH_DropDatabase_Success(t *testing.T) {
	mockConn := &mockCHConn{
		execFn: func(ctx context.Context, query string, args ...interface{}) error { return nil },
	}
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), conn: mockConn, closed: false}
	err := client.DropDatabase(context.Background(), "olddb")
	assert.NoError(t, err)
}

func TestCov_CH_DropDatabase_Error(t *testing.T) {
	mockConn := &mockCHConn{
		execFn: func(ctx context.Context, query string, args ...interface{}) error {
			return fmt.Errorf("drop db failed")
		},
	}
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), conn: mockConn, closed: false}
	err := client.DropDatabase(context.Background(), "olddb")
	assert.Error(t, err)
}

func TestCov_CH_DropDatabase_Closed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	err := client.DropDatabase(context.Background(), "olddb")
	assert.Equal(t, ErrClosed, err)
}

func TestCov_CH_CreateTable_Success(t *testing.T) {
	mockConn := &mockCHConn{
		execFn: func(ctx context.Context, query string, args ...interface{}) error { return nil },
	}
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), conn: mockConn, closed: false}
	schema := &TableSchema{Columns: []ColumnSchema{{Name: "value", Type: "DOUBLE"}}}
	err := client.CreateTable(context.Background(), "db", "metrics", schema)
	assert.NoError(t, err)
}

func TestCov_CH_CreateTable_Error(t *testing.T) {
	mockConn := &mockCHConn{
		execFn: func(ctx context.Context, query string, args ...interface{}) error {
			return fmt.Errorf("create table failed")
		},
	}
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), conn: mockConn, closed: false}
	schema := &TableSchema{Columns: []ColumnSchema{{Name: "value", Type: "DOUBLE"}}}
	err := client.CreateTable(context.Background(), "db", "metrics", schema)
	assert.Error(t, err)
}

func TestCov_CH_CreateTable_Closed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	err := client.CreateTable(context.Background(), "db", "metrics", &TableSchema{})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_CH_DropTable_Success(t *testing.T) {
	mockConn := &mockCHConn{
		execFn: func(ctx context.Context, query string, args ...interface{}) error { return nil },
	}
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), conn: mockConn, closed: false}
	err := client.DropTable(context.Background(), "db", "metrics")
	assert.NoError(t, err)
}

func TestCov_CH_DropTable_Error(t *testing.T) {
	mockConn := &mockCHConn{
		execFn: func(ctx context.Context, query string, args ...interface{}) error {
			return fmt.Errorf("drop table failed")
		},
	}
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), conn: mockConn, closed: false}
	err := client.DropTable(context.Background(), "db", "metrics")
	assert.Error(t, err)
}

func TestCov_CH_DropTable_Closed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	err := client.DropTable(context.Background(), "db", "metrics")
	assert.Equal(t, ErrClosed, err)
}

func TestCov_CH_Ping_Success(t *testing.T) {
	mockConn := &mockCHConn{
		pingFn: func(ctx context.Context) error { return nil },
	}
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), conn: mockConn, closed: false}
	err := client.Ping(context.Background())
	assert.NoError(t, err)
}

func TestCov_CH_Ping_Error(t *testing.T) {
	mockConn := &mockCHConn{
		pingFn: func(ctx context.Context) error { return fmt.Errorf("ping failed") },
	}
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), conn: mockConn, closed: false}
	err := client.Ping(context.Background())
	assert.Error(t, err)
}

func TestCov_CH_Ping_Closed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	err := client.Ping(context.Background())
	assert.Equal(t, ErrClosed, err)
}

func TestCov_CH_Close_AlreadyClosed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	err := client.Close()
	assert.NoError(t, err)
}

func TestCov_CH_Close_NilConn(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), conn: nil, closed: false}
	err := client.Close()
	assert.NoError(t, err)
	assert.True(t, client.IsClosed())
}

func TestCov_CH_Close_WithConn(t *testing.T) {
	mockConn := &mockCHConn{closeFn: func() error { return nil }}
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), conn: mockConn, closed: false}
	err := client.Close()
	assert.NoError(t, err)
	assert.True(t, client.IsClosed())
}

func TestCov_CH_QueryRange_Closed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: true}
	_, err := client.QueryRange(context.Background(), time.Now().Add(-1*time.Hour), time.Now(), []int64{1})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_CH_QueryRange_Success(t *testing.T) {
	mockConn := &mockCHConn{
		queryFn: func(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
			return &mockCHRows{nextFn: func() bool { return false }, scanFn: func(dest ...interface{}) error { return nil }}, nil
		},
	}
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), conn: mockConn, closed: false}
	_, err := client.QueryRange(context.Background(), time.Now().Add(-1*time.Hour), time.Now(), []int64{1})
	assert.NoError(t, err)
}

func TestCov_CH_CreateRetentionPolicy(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	err := client.CreateRetentionPolicy(context.Background(), &RetentionPolicy{Name: "rp1", Database: "db"})
	assert.NoError(t, err)
}

func TestCov_CH_UpdateRetentionPolicy(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	err := client.UpdateRetentionPolicy(context.Background(), &RetentionPolicy{Name: "rp1", Database: "db"})
	assert.NoError(t, err)
}

func TestCov_CH_DeleteRetentionPolicy(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	err := client.DeleteRetentionPolicy(context.Background(), "db", "rp1")
	assert.NoError(t, err)
}

func TestCov_CH_IsClosed(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	assert.False(t, client.IsClosed())
	assert.True(t, client.IsConnected())
	client.closed = true
	assert.True(t, client.IsClosed())
	assert.False(t, client.IsConnected())
}

func TestCov_CH_buildQuerySQL_WithPointIDs(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	q := &Query{Database: "testdb", Table: "metrics", PointIDs: []int64{1, 2, 3}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now()}
	sql, args := client.buildQuerySQL(q)
	assert.Contains(t, sql, "point_id IN (?)")
	assert.Len(t, args, 3)
}

func TestCov_CH_buildQuerySQL_DefaultDB(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	q := &Query{PointIDs: []int64{1}, StartTime: time.Now().Add(-1 * time.Hour)}
	sql, _ := client.buildQuerySQL(q)
	assert.Contains(t, sql, "nem_ts")
}

func TestCov_CH_buildQuerySQL_WithOffset(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	q := &Query{PointIDs: []int64{1}, StartTime: time.Now().Add(-1 * time.Hour), Limit: 100, Offset: 50}
	sql, _ := client.buildQuerySQL(q)
	assert.Contains(t, sql, "LIMIT 100 OFFSET 50")
}

func TestCov_CH_buildQuerySQL_DESCOrder(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	q := &Query{PointIDs: []int64{1}, StartTime: time.Now().Add(-1 * time.Hour), OrderBy: "value", Order: "DESC"}
	sql, _ := client.buildQuerySQL(q)
	assert.Contains(t, sql, "ORDER BY value DESC")
}

func TestCov_CH_buildCountSQL(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	q := &Query{Database: "testdb", PointIDs: []int64{1, 2}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now()}
	sql, args := client.buildCountSQL(q)
	assert.Contains(t, sql, "COUNT(*)")
	assert.Len(t, args, 3)
}

func TestCov_CH_buildAggregateSQL(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	q := &AggregateQuery{Database: "testdb", Table: "metrics", PointIDs: []int64{1, 2}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now(), Interval: 5 * time.Minute, AggFunc: "avg"}
	sql, args := client.buildAggregateSQL(q)
	assert.Contains(t, sql, "AVG(value)")
	assert.Len(t, args, 3)
}

func TestCov_CH_buildAggregateSQL_DefaultAggFunc(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	q := &AggregateQuery{Database: "testdb", StartTime: time.Now().Add(-1 * time.Hour), Interval: 5 * time.Minute}
	sql, _ := client.buildAggregateSQL(q)
	assert.Contains(t, sql, "avg(value)")
}

func TestCov_CH_intervalToSQL(t *testing.T) {
	client := &ClickHouseClient{config: DefaultClickHouseConfig(), logger: zap.NewNop(), closed: false}
	assert.Equal(t, "30 SECOND", client.intervalToSQL(30*time.Second))
	assert.Equal(t, "5 MINUTE", client.intervalToSQL(5*time.Minute))
	assert.Equal(t, "2 HOUR", client.intervalToSQL(2*time.Hour))
	assert.Equal(t, "2 DAY", client.intervalToSQL(48*time.Hour))
}

func TestCov_CH_NewClickHouseClient_NoAddr(t *testing.T) {
	_, err := NewClickHouseClient(&ClickHouseConfig{})
	assert.Error(t, err)
}

func TestCov_CH_NewClickHouseClient_NilConfig(t *testing.T) {
	_, err := NewClickHouseClient(nil)
	assert.Error(t, err)
}

// Doris tests with mock sql.DB

func TestCov_DO_WriteWithTable_Success(t *testing.T) {
	db := newMockSQLDB()
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), db: db, closed: false}
	points := []*DataPoint{{PointID: 1, Value: 80.0, Timestamp: time.Now()}}
	err := client.WriteWithTable(context.Background(), "db", "metrics", points)
	assert.NoError(t, err)
}

func TestCov_DO_Query_Success(t *testing.T) {
	db := newMockSQLDB5Cols()
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), db: db, closed: false}
	q := &Query{Database: "testdb", Table: "metrics", PointIDs: []int64{1}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now()}
	_, err := client.Query(context.Background(), q)
	assert.NoError(t, err)
}

func TestCov_DO_QueryLatest_Success(t *testing.T) {
	db := newMockSQLDB()
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), db: db, closed: false}
	_, err := client.QueryLatest(context.Background(), []int64{1, 2})
	assert.NoError(t, err)
}

func TestCov_DO_Aggregate_Success(t *testing.T) {
	db := newMockSQLDB5Cols()
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), db: db, closed: false}
	q := &AggregateQuery{Database: "testdb", Table: "metrics", PointIDs: []int64{1}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now(), Interval: 5 * time.Minute, AggFunc: "avg"}
	_, err := client.Aggregate(context.Background(), q)
	assert.NoError(t, err)
}

func TestCov_DO_Downsample_Success(t *testing.T) {
	db := newMockSQLDB5Cols()
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), db: db, closed: false}
	err := client.Downsample(context.Background(), &DownsampleQuery{
		SourceDatabase: "testdb", SourceTable: "metrics", TargetDatabase: "testdb", TargetTable: "metrics_down",
		PointIDs: []int64{1}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now(),
		Interval: 5 * time.Minute, AggFunc: "avg",
	})
	assert.NoError(t, err)
}

func TestCov_DO_CreateDatabase_Success(t *testing.T) {
	db := newMockSQLDB()
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), db: db, closed: false}
	err := client.CreateDatabase(context.Background(), "newdb")
	assert.NoError(t, err)
}

func TestCov_DO_DropDatabase_Success(t *testing.T) {
	db := newMockSQLDB()
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), db: db, closed: false}
	err := client.DropDatabase(context.Background(), "olddb")
	assert.NoError(t, err)
}

func TestCov_DO_CreateTable_Success(t *testing.T) {
	db := newMockSQLDB()
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), db: db, closed: false}
	schema := &TableSchema{Columns: []ColumnSchema{{Name: "value", Type: "DOUBLE"}}}
	err := client.CreateTable(context.Background(), "db", "metrics", schema)
	assert.NoError(t, err)
}

func TestCov_DO_DropTable_Success(t *testing.T) {
	db := newMockSQLDB()
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), db: db, closed: false}
	err := client.DropTable(context.Background(), "db", "metrics")
	assert.NoError(t, err)
}

func TestCov_DO_Ping_Success(t *testing.T) {
	db := newMockSQLDB()
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), db: db, closed: false}
	err := client.Ping(context.Background())
	assert.NoError(t, err)
}

func TestCov_DO_Close_WithDB(t *testing.T) {
	db := newMockSQLDB()
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), db: db, closed: false}
	err := client.Close()
	assert.NoError(t, err)
	assert.True(t, client.IsClosed())
}

func TestCov_DO_QueryRange_Success(t *testing.T) {
	db := newMockSQLDB5Cols()
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), db: db, closed: false}
	_, err := client.QueryRange(context.Background(), time.Now().Add(-1*time.Hour), time.Now(), []int64{1})
	assert.NoError(t, err)
}

func TestCov_DO_WriteWithTable_Closed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	err := client.WriteWithTable(context.Background(), "db", "metrics", []*DataPoint{{PointID: 1, Value: 80.0, Timestamp: time.Now()}})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_DO_WriteWithTable_EmptyPoints(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	err := client.WriteWithTable(context.Background(), "db", "metrics", []*DataPoint{})
	assert.NoError(t, err)
}

func TestCov_DO_WriteWithTable_NilPointsInList(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	err := client.WriteWithTable(context.Background(), "db", "metrics", []*DataPoint{nil, nil})
	assert.NoError(t, err)
}

func TestCov_DO_WriteWithTable_InvalidDBName(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	err := client.WriteWithTable(context.Background(), "invalid-db", "metrics", []*DataPoint{{PointID: 1, Value: 80.0, Timestamp: time.Now()}})
	assert.Error(t, err)
}

func TestCov_DO_WriteWithTable_InvalidTableName(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	err := client.WriteWithTable(context.Background(), "db", "invalid-table", []*DataPoint{{PointID: 1, Value: 80.0, Timestamp: time.Now()}})
	assert.Error(t, err)
}

func TestCov_DO_Write_Closed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	err := client.Write(context.Background(), []*DataPoint{{PointID: 1, Value: 80.0, Timestamp: time.Now()}})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_DO_WriteBatch_Closed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	err := client.WriteBatch(context.Background(), []*DataPoint{{PointID: 1, Value: 80.0, Timestamp: time.Now()}})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_DO_Query_Closed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	_, err := client.Query(context.Background(), &Query{})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_DO_Query_NilQuery(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	_, err := client.Query(context.Background(), nil)
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestCov_DO_QueryLatest_Closed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	_, err := client.QueryLatest(context.Background(), []int64{1})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_DO_QueryLatest_EmptyPointIDs(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	result, err := client.QueryLatest(context.Background(), []int64{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestCov_DO_Aggregate_Closed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	_, err := client.Aggregate(context.Background(), &AggregateQuery{})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_DO_Aggregate_NilQuery(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	_, err := client.Aggregate(context.Background(), nil)
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestCov_DO_Downsample_Closed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	err := client.Downsample(context.Background(), &DownsampleQuery{})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_DO_CreateDatabase_Closed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	err := client.CreateDatabase(context.Background(), "newdb")
	assert.Equal(t, ErrClosed, err)
}

func TestCov_DO_CreateDatabase_InvalidName(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	err := client.CreateDatabase(context.Background(), "invalid-db")
	assert.Error(t, err)
}

func TestCov_DO_DropDatabase_Closed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	err := client.DropDatabase(context.Background(), "olddb")
	assert.Equal(t, ErrClosed, err)
}

func TestCov_DO_DropDatabase_InvalidName(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	err := client.DropDatabase(context.Background(), "invalid-db")
	assert.Error(t, err)
}

func TestCov_DO_CreateTable_Closed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	err := client.CreateTable(context.Background(), "db", "metrics", &TableSchema{})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_DO_CreateTable_InvalidDBName(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	err := client.CreateTable(context.Background(), "invalid-db", "metrics", &TableSchema{})
	assert.Error(t, err)
}

func TestCov_DO_CreateTable_InvalidTableName(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	err := client.CreateTable(context.Background(), "db", "invalid-table", &TableSchema{})
	assert.Error(t, err)
}

func TestCov_DO_CreateTable_InvalidColumnName(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	schema := &TableSchema{Columns: []ColumnSchema{{Name: "invalid-col", Type: "DOUBLE"}}}
	err := client.CreateTable(context.Background(), "db", "metrics", schema)
	assert.Error(t, err)
}

func TestCov_DO_DropTable_Closed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	err := client.DropTable(context.Background(), "db", "metrics")
	assert.Equal(t, ErrClosed, err)
}

func TestCov_DO_DropTable_InvalidDBName(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	err := client.DropTable(context.Background(), "invalid-db", "metrics")
	assert.Error(t, err)
}

func TestCov_DO_Ping_Closed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	err := client.Ping(context.Background())
	assert.Equal(t, ErrClosed, err)
}

func TestCov_DO_Close_AlreadyClosed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	err := client.Close()
	assert.NoError(t, err)
}

func TestCov_DO_Close_NilDB(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	err := client.Close()
	assert.NoError(t, err)
}

func TestCov_DO_QueryRange_Closed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: true}
	_, err := client.QueryRange(context.Background(), time.Now().Add(-1*time.Hour), time.Now(), []int64{1})
	assert.Equal(t, ErrClosed, err)
}

func TestCov_DO_CreateRetentionPolicy(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	err := client.CreateRetentionPolicy(context.Background(), &RetentionPolicy{Name: "rp1", Database: "db"})
	assert.NoError(t, err)
}

func TestCov_DO_UpdateRetentionPolicy(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	err := client.UpdateRetentionPolicy(context.Background(), &RetentionPolicy{Name: "rp1", Database: "db"})
	assert.NoError(t, err)
}

func TestCov_DO_DeleteRetentionPolicy(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	err := client.DeleteRetentionPolicy(context.Background(), "db", "rp1")
	assert.NoError(t, err)
}

func TestCov_DO_IsClosed(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	assert.False(t, client.IsClosed())
	assert.True(t, client.IsConnected())
	client.closed = true
	assert.True(t, client.IsClosed())
	assert.False(t, client.IsConnected())
}

func TestCov_DO_NewDorisClient_NoHosts(t *testing.T) {
	_, err := NewDorisClient(&DorisConfig{})
	assert.Error(t, err)
}

func TestCov_DO_NewDorisClient_NilConfig(t *testing.T) {
	_, err := NewDorisClient(nil)
	assert.Error(t, err)
}

func TestCov_DO_buildDSN(t *testing.T) {
	client := &DorisClient{config: &DorisConfig{Hosts: []string{"localhost"}, Port: 9030, Database: "testdb", User: "root", Password: "secret"}}
	dsn := client.buildDSN("localhost")
	assert.Contains(t, dsn, "root:secret@tcp(localhost:9030)/testdb")
}

func TestCov_DO_buildDSN_NoPassword(t *testing.T) {
	client := &DorisClient{config: &DorisConfig{Hosts: []string{"localhost"}, Port: 9030, Database: "testdb", User: "root"}}
	dsn := client.buildDSN("localhost")
	assert.Contains(t, dsn, "root:@tcp(localhost:9030)/testdb")
}

func TestCov_DO_serializeTags(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	result := client.serializeTags(map[string]string{"host": "s1", "region": "us"})
	assert.Contains(t, result, "host=s1")
	assert.Contains(t, result, "region=us")
}

func TestCov_DO_serializeTags_Nil(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	assert.Equal(t, "", client.serializeTags(nil))
}

func TestCov_DO_serializeTags_Empty(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	assert.Equal(t, "", client.serializeTags(map[string]string{}))
}

func TestCov_DO_parseTags(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	result := client.parseTags("host=s1,region=us")
	assert.Equal(t, "s1", result["host"])
	assert.Equal(t, "us", result["region"])
}

func TestCov_DO_parseTags_Empty(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	assert.Empty(t, client.parseTags(""))
}

func TestCov_DO_parseTags_NoValue(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	assert.Empty(t, client.parseTags("host"))
}

func TestCov_DO_buildQuerySQL(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	q := &Query{Database: "testdb", Table: "metrics", PointIDs: []int64{1, 2}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now()}
	sql, args := client.buildQuerySQL(q)
	assert.Contains(t, sql, "point_id IN (")
	assert.Len(t, args, 4)
}

func TestCov_DO_buildQuerySQL_DefaultDB(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	q := &Query{PointIDs: []int64{1}, StartTime: time.Now().Add(-1 * time.Hour)}
	sql, _ := client.buildQuerySQL(q)
	assert.Contains(t, sql, "nem_ts")
}

func TestCov_DO_buildCountSQL(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	q := &Query{Database: "testdb", PointIDs: []int64{1, 2}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now()}
	sql, args := client.buildCountSQL(q)
	assert.Contains(t, sql, "COUNT(*)")
	assert.Len(t, args, 4)
}

func TestCov_DO_buildAggregateSQL(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	q := &AggregateQuery{Database: "testdb", Table: "metrics", PointIDs: []int64{1, 2}, StartTime: time.Now().Add(-1 * time.Hour), EndTime: time.Now(), Interval: 5 * time.Minute, AggFunc: "avg"}
	sql, args := client.buildAggregateSQL(q)
	assert.Contains(t, sql, "AVG(value)")
	assert.Len(t, args, 4)
}

func TestCov_DO_buildAggregateSQL_DefaultAggFunc(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	q := &AggregateQuery{Database: "testdb", StartTime: time.Now().Add(-1 * time.Hour), Interval: 5 * time.Minute}
	sql, _ := client.buildAggregateSQL(q)
	assert.Contains(t, sql, "AVG(value)")
}

func TestCov_DO_intervalToSQL(t *testing.T) {
	client := &DorisClient{config: DefaultDorisConfig(), logger: zap.NewNop(), closed: false}
	assert.Equal(t, "30 SECOND", client.intervalToSQL(30*time.Second))
	assert.Equal(t, "5 MINUTE", client.intervalToSQL(5*time.Minute))
	assert.Equal(t, "2 HOUR", client.intervalToSQL(2*time.Hour))
	assert.Equal(t, "2 DAY", client.intervalToSQL(48*time.Hour))
}

func TestCov_ValidateIdentifier(t *testing.T) {
	assert.NoError(t, validateIdentifier("valid_name", "table"))
	assert.NoError(t, validateIdentifier("table123", "table"))
	assert.NoError(t, validateIdentifier("_underscore", "table"))
	assert.Error(t, validateIdentifier("", "table"))
	assert.Error(t, validateIdentifier("123abc", "table"))
	assert.Error(t, validateIdentifier("invalid-name", "table"))
	assert.Error(t, validateIdentifier("has space", "table"))
}

func TestCov_ValidateIdentifier_TooLong(t *testing.T) {
	longName := ""
	for i := 0; i < 65; i++ {
		longName += "a"
	}
	assert.Error(t, validateIdentifier(longName, "table"))
}

func TestCov_QualityDescription_All(t *testing.T) {
	assert.Equal(t, "good", QualityDescription(QualityGood))
	assert.Equal(t, "bad", QualityDescription(QualityBad))
	assert.Equal(t, "uncertain", QualityDescription(QualityUncertain))
	assert.Equal(t, "missing", QualityDescription(QualityMissing))
	assert.Equal(t, "unknown", QualityDescription(999))
}

func TestCov_IsRetryableError(t *testing.T) {
	assert.True(t, IsRetryableError(NewConnectionError("localhost", "failed", nil)))
	assert.False(t, IsRetryableError(NewQueryError("SELECT 1", "failed", nil)))
	assert.True(t, IsRetryableError(ErrTimeout))
	assert.False(t, IsRetryableError(nil))
}

func TestCov_IsNotFoundError(t *testing.T) {
	assert.True(t, IsNotFoundError(ErrDatabaseNotFound))
	assert.True(t, IsNotFoundError(ErrTableNotFound))
	assert.False(t, IsNotFoundError(ErrClosed))
}

func TestCov_QueryError_WithCause(t *testing.T) {
	cause := fmt.Errorf("syntax error")
	err := NewQueryError("SELECT 1", "query failed", cause)
	assert.Contains(t, err.Error(), "query failed")
	assert.Contains(t, err.Error(), "SELECT 1")
	assert.Contains(t, err.Error(), "syntax error")
	assert.Equal(t, cause, err.Unwrap())
}

func TestCov_QueryError_WithoutCause(t *testing.T) {
	err := NewQueryError("SELECT 1", "query failed", nil)
	assert.Contains(t, err.Error(), "query failed")
	assert.Contains(t, err.Error(), "SELECT 1")
	assert.Nil(t, err.Unwrap())
}

func TestCov_WriteError_WithCause(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := NewWriteError(10, "write failed", cause)
	assert.Contains(t, err.Error(), "write failed")
	assert.Contains(t, err.Error(), "10")
	assert.Contains(t, err.Error(), "connection refused")
	assert.Equal(t, cause, err.Unwrap())
}

func TestCov_WriteError_WithoutCause(t *testing.T) {
	err := NewWriteError(5, "write failed", nil)
	assert.Contains(t, err.Error(), "write failed")
	assert.Contains(t, err.Error(), "5")
	assert.Nil(t, err.Unwrap())
}

func TestCov_ConnectionError_WithCause(t *testing.T) {
	cause := fmt.Errorf("timeout")
	err := NewConnectionError("localhost:9000", "connection failed", cause)
	assert.Contains(t, err.Error(), "connection failed")
	assert.Contains(t, err.Error(), "localhost:9000")
	assert.Contains(t, err.Error(), "timeout")
	assert.Equal(t, cause, err.Unwrap())
}

func TestCov_ConnectionError_WithoutCause(t *testing.T) {
	err := NewConnectionError("localhost:9000", "connection failed", nil)
	assert.Contains(t, err.Error(), "connection failed")
	assert.Contains(t, err.Error(), "localhost:9000")
	assert.Nil(t, err.Unwrap())
}

func TestCov_Factory_UnsupportedType(t *testing.T) {
	factory := &TimeSeriesFactory{}
	_, err := factory.NewTimeSeriesDB("invalid", nil)
	assert.Error(t, err)
}

func TestCov_Factory_DorisInvalidConfig(t *testing.T) {
	factory := &TimeSeriesFactory{}
	_, err := factory.NewTimeSeriesDB(TimeSeriesDBTypeDoris, "invalid")
	assert.Error(t, err)
}

func TestCov_Factory_ClickHouseInvalidConfig(t *testing.T) {
	factory := &TimeSeriesFactory{}
	_, err := factory.NewTimeSeriesDB(TimeSeriesDBTypeClickHouse, "invalid")
	assert.Error(t, err)
}

func TestCov_Factory_NewTimeSeriesDBFromConfig_Nil(t *testing.T) {
	_, err := NewTimeSeriesDBFromConfig(nil)
	assert.Error(t, err)
}

func TestCov_Factory_NewTimeSeriesDBFromConfig_Unsupported(t *testing.T) {
	_, err := NewTimeSeriesDBFromConfig(&internalconfig.TimeSeriesConfig{Type: "influxdb"})
	assert.Error(t, err)
}

func TestCov_Factory_NewTimeSeriesDBFromConfig_Doris(t *testing.T) {
	_, err := NewTimeSeriesDBFromConfig(&internalconfig.TimeSeriesConfig{
		Type:  "doris",
		Doris: internalconfig.DorisConfig{Hosts: []string{"localhost:9030"}, Database: "testdb", User: "root"},
	})
	assert.Error(t, err)
}

func TestCov_Factory_NewTimeSeriesDBFromConfig_ClickHouse(t *testing.T) {
	_, err := NewTimeSeriesDBFromConfig(&internalconfig.TimeSeriesConfig{
		Type:      "clickhouse",
		ClickHouse: internalconfig.ClickHouseConfig{Addr: []string{"localhost:9000"}, Database: "testdb"},
	})
	assert.Error(t, err)
}

func TestCov_Factory_MustNewTimeSeriesDBFromConfig_Panic(t *testing.T) {
	assert.Panics(t, func() {
		MustNewTimeSeriesDBFromConfig(nil)
	})
}

func TestCov_Factory_MustNewTimeSeriesDBFromConfig_Unsupported(t *testing.T) {
	assert.Panics(t, func() {
		MustNewTimeSeriesDBFromConfig(&internalconfig.TimeSeriesConfig{Type: "influxdb"})
	})
}

func TestCov_BatchWriter_DefaultConfig(t *testing.T) {
	cfg := DefaultBatchWriterConfig()
	assert.Equal(t, 10000, cfg.BatchSize)
	assert.Equal(t, 5*time.Second, cfg.FlushTimeout)
	assert.Equal(t, 3, cfg.MaxRetries)
}

func TestCov_BatchWriter_WriteWithRetry_Success(t *testing.T) {
	mockDB := &mockTimeSeriesDB{}
	cfg := &BatchWriterConfig{Database: "testdb", Table: "metrics", BatchSize: 100, FlushTimeout: 5 * time.Second, MaxRetries: 3, RetryDelay: 1 * time.Millisecond}
	writer := NewBatchWriter(mockDB, cfg)
	defer writer.Close()
	err := writer.Add(&DataPoint{PointID: 1, Value: 80.0, Timestamp: time.Now()})
	assert.NoError(t, err)
	err = writer.Flush()
	assert.NoError(t, err)
	stats := writer.Stats()
	assert.Equal(t, int64(1), stats.TotalPoints)
	assert.Equal(t, int64(1), stats.SuccessPoints)
}

func TestCov_BatchWriter_WriteWithRetry_RetryableError(t *testing.T) {
	callCount := 0
	mockDB := &mockTimeSeriesDB{
		writeFn: func(ctx context.Context, database, table string, points []*DataPoint) error {
			callCount++
			if callCount < 3 {
				return NewConnectionError("localhost", "connection failed", nil)
			}
			return nil
		},
	}
	cfg := &BatchWriterConfig{Database: "testdb", Table: "metrics", BatchSize: 100, FlushTimeout: 5 * time.Second, MaxRetries: 3, RetryDelay: 1 * time.Millisecond}
	writer := NewBatchWriter(mockDB, cfg)
	defer writer.Close()
	err := writer.Add(&DataPoint{PointID: 1, Value: 80.0, Timestamp: time.Now()})
	assert.NoError(t, err)
	err = writer.Flush()
	assert.NoError(t, err)
	assert.True(t, callCount >= 3)
}

func TestCov_BatchWriter_WriteWithRetry_NonRetryableError(t *testing.T) {
	mockDB := &mockTimeSeriesDB{writeErr: NewQueryError("INSERT", "query failed", nil)}
	cfg := &BatchWriterConfig{Database: "testdb", Table: "metrics", BatchSize: 100, FlushTimeout: 5 * time.Second, MaxRetries: 3, RetryDelay: 1 * time.Millisecond}
	writer := NewBatchWriter(mockDB, cfg)
	defer writer.Close()
	err := writer.Add(&DataPoint{PointID: 1, Value: 80.0, Timestamp: time.Now()})
	assert.NoError(t, err)
	err = writer.Flush()
	assert.Error(t, err)
}

func TestCov_BatchWriter_WriteWithRetry_AllRetriesFail(t *testing.T) {
	mockDB := &mockTimeSeriesDB{writeErr: NewConnectionError("localhost", "connection failed", nil)}
	cfg := &BatchWriterConfig{Database: "testdb", Table: "metrics", BatchSize: 100, FlushTimeout: 5 * time.Second, MaxRetries: 2, RetryDelay: 1 * time.Millisecond}
	writer := NewBatchWriter(mockDB, cfg)
	defer writer.Close()
	err := writer.Add(&DataPoint{PointID: 1, Value: 80.0, Timestamp: time.Now()})
	assert.NoError(t, err)
	err = writer.Flush()
	assert.Error(t, err)
}

func TestCov_BatchWriter_NilConfig(t *testing.T) {
	mockDB := &mockTimeSeriesDB{}
	writer := NewBatchWriter(mockDB, nil)
	assert.NotNil(t, writer)
	writer.Close()
}

func TestCov_BatchWriter_AddNilPoint(t *testing.T) {
	mockDB := &mockTimeSeriesDB{}
	writer := NewBatchWriter(mockDB, DefaultBatchWriterConfig())
	defer writer.Close()
	err := writer.Add(nil)
	assert.NoError(t, err)
}

func TestCov_BatchWriter_FlushEmpty(t *testing.T) {
	mockDB := &mockTimeSeriesDB{}
	writer := NewBatchWriter(mockDB, DefaultBatchWriterConfig())
	defer writer.Close()
	err := writer.Flush()
	assert.NoError(t, err)
}

func TestCov_AsyncBatchWriter_AddAndClose(t *testing.T) {
	mockDB := &mockTimeSeriesDB{}
	writer := NewAsyncBatchWriter(mockDB, DefaultBatchWriterConfig())
	assert.NotNil(t, writer)
	errCh := writer.Add(&DataPoint{PointID: 1, Value: 80.0, Timestamp: time.Now()})
	assert.NotNil(t, errCh)
	stats := writer.Stats()
	assert.NotNil(t, stats)
	errCh2 := writer.Errors()
	assert.NotNil(t, errCh2)
	err := writer.Close()
	assert.NoError(t, err)
}
