package timeseries

import (
	"context"
	"time"
)

// TimeSeriesDB 时序数据库接口
type TimeSeriesDB interface {
	// 写入接口
	Write(ctx context.Context, points []*DataPoint) error
	WriteBatch(ctx context.Context, points []*DataPoint) error
	WriteWithTable(ctx context.Context, database, table string, points []*DataPoint) error

	// 查询接口
	Query(ctx context.Context, query *Query) (*QueryResult, error)
	QueryRange(ctx context.Context, start, end time.Time, pointIds []int64) ([]*DataPoint, error)
	QueryLatest(ctx context.Context, pointIds []int64) (map[int64]*DataPoint, error)

	// 聚合接口
	Aggregate(ctx context.Context, query *AggregateQuery) (*AggregateResult, error)

	// 降采样接口
	Downsample(ctx context.Context, query *DownsampleQuery) error

	// 管理接口
	CreateDatabase(ctx context.Context, name string) error
	DropDatabase(ctx context.Context, name string) error
	CreateTable(ctx context.Context, database, table string, schema *TableSchema) error
	DropTable(ctx context.Context, database, table string) error

	// 保留策略
	CreateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error
	UpdateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error
	DeleteRetentionPolicy(ctx context.Context, database, name string) error

	// 健康检查
	Ping(ctx context.Context) error
	IsConnected() bool

	// 连接管理
	Close() error
}

// TableSchema 表结构定义
type TableSchema struct {
	Name       string          `json:"name"`       // 表名
	Columns    []ColumnSchema  `json:"columns"`    // 列定义
	PrimaryKey []string        `json:"primary_key"` // 主键
	SortKey    []string        `json:"sort_key"`   // 排序键
	PartitionBy string         `json:"partition_by"` // 分区字段
	Engine     string          `json:"engine"`     // 存储引擎
}

// ColumnSchema 列结构定义
type ColumnSchema struct {
	Name       string `json:"name"`        // 列名
	Type       string `json:"type"`        // 数据类型
	Nullable   bool   `json:"nullable"`    // 是否可空
	Default    string `json:"default"`     // 默认值
	Comment    string `json:"comment"`     // 注释
}

// TimeSeriesDBType 时序数据库类型
type TimeSeriesDBType string

const (
	TimeSeriesDBTypeDoris      TimeSeriesDBType = "doris"
	TimeSeriesDBTypeClickHouse TimeSeriesDBType = "clickhouse"
	TimeSeriesDBTypeInfluxDB   TimeSeriesDBType = "influxdb"
	TimeSeriesDBTypeTimescaleDB TimeSeriesDBType = "timescaledb"
)

// TimeSeriesFactory 时序数据库工厂
type TimeSeriesFactory struct{}

// NewTimeSeriesDB 创建时序数据库客户端
func (f *TimeSeriesFactory) NewTimeSeriesDB(dbType TimeSeriesDBType, config interface{}) (TimeSeriesDB, error) {
	switch dbType {
	case TimeSeriesDBTypeDoris:
		dorisConfig, ok := config.(*DorisConfig)
		if !ok {
			return nil, ErrInvalidConfig
		}
		return NewDorisClient(dorisConfig)
	case TimeSeriesDBTypeClickHouse:
		chConfig, ok := config.(*ClickHouseConfig)
		if !ok {
			return nil, ErrInvalidConfig
		}
		return NewClickHouseClient(chConfig)
	default:
		return nil, ErrUnsupportedDBType
	}
}

// BatchWriter 批量写入器接口
type BatchWriter interface {
	// Add 添加数据点
	Add(point *DataPoint) error
	// Flush 刷新缓冲区
	Flush() error
	// Close 关闭写入器
	Close() error
	// Stats 获取统计信息
	Stats() *WriterStats
}

// WriterStats 写入器统计信息
type WriterStats struct {
	TotalPoints   int64 // 总数据点数
	SuccessPoints int64 // 成功数据点数
	FailedPoints  int64 // 失败数据点数
	TotalBatches  int64 // 总批次数
	TotalBytes    int64 // 总字节数
}

// QueryOptimizer 查询优化器接口
type QueryOptimizer interface {
	// Optimize 优化查询
	Optimize(query *Query) (*OptimizedQuery, error)
	// EstimateCost 估算查询成本
	EstimateCost(query *Query) (int64, error)
	// SuggestIndexes 建议索引
	SuggestIndexes(query *Query) ([]IndexSuggestion, error)
}

// OptimizedQuery 优化后的查询
type OptimizedQuery struct {
	OriginalQuery  *Query
	OptimizedSQL   string
	EstimatedCost  int64
	UsedIndexes    []string
	ExecutionPlan  string
}

// IndexSuggestion 索引建议
type IndexSuggestion struct {
	TableName  string   // 表名
	ColumnName string   // 列名
	IndexType  string   // 索引类型
	Reason     string   // 建议原因
}

// ConnectionPool 连接池接口
type ConnectionPool interface {
	// Get 获取连接
	Get() (interface{}, error)
	// Put 归还连接
	Put(conn interface{}) error
	// Close 关闭连接池
	Close() error
	// Stats 获取连接池统计
	Stats() *PoolStats
}

// PoolStats 连接池统计
type PoolStats struct {
	TotalConnections int64 // 总连接数
	IdleConnections  int64 // 空闲连接数
	ActiveConnections int64 // 活跃连接数
	WaitCount        int64 // 等待次数
	WaitDuration     time.Duration // 等待时长
}
