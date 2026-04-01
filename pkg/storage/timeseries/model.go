package timeseries

import (
	"time"
)

// DataPoint 时序数据点
type DataPoint struct {
	PointID   int64             `json:"point_id"`   // 测点ID
	Timestamp time.Time         `json:"timestamp"`  // 时间戳
	Value     float64           `json:"value"`      // 数值
	Quality   int               `json:"quality"`    // 数据质量标识
	Tags      map[string]string `json:"tags"`       // 标签信息
}

// Query 查询条件
type Query struct {
	Database  string    `json:"database"`   // 数据库名
	Table     string    `json:"table"`      // 表名
	PointIDs  []int64   `json:"point_ids"`  // 测点ID列表
	StartTime time.Time `json:"start_time"` // 开始时间
	EndTime   time.Time `json:"end_time"`   // 结束时间
	Limit     int       `json:"limit"`      // 返回条数限制
	Offset    int       `json:"offset"`     // 偏移量
	OrderBy   string    `json:"order_by"`   // 排序字段
	Order     string    `json:"order"`      // 排序方式 ASC/DESC
}

// QueryResult 查询结果
type QueryResult struct {
	Points []*DataPoint `json:"points"` // 数据点列表
	Total  int64        `json:"total"`  // 总数
}

// AggregateQuery 聚合查询条件
type AggregateQuery struct {
	Database    string        `json:"database"`     // 数据库名
	Table       string        `json:"table"`        // 表名
	PointIDs    []int64       `json:"point_ids"`    // 测点ID列表
	StartTime   time.Time     `json:"start_time"`   // 开始时间
	EndTime     time.Time     `json:"end_time"`     // 结束时间
	Interval    time.Duration `json:"interval"`     // 聚合间隔
	AggFunc     string        `json:"agg_func"`     // 聚合函数: avg, sum, max, min, count
	GroupBy     []string      `json:"group_by"`     // 分组字段
	Fill        string        `json:"fill"`         // 空值填充策略: null, previous, linear, 0
}

// AggregateResult 聚合结果
type AggregateResult struct {
	Series []TimeSeries `json:"series"` // 时间序列列表
}

// TimeSeries 时间序列
type TimeSeries struct {
	PointID   int64           `json:"point_id"`   // 测点ID
	Tags      map[string]string `json:"tags"`     // 标签
	Points    []AggPoint      `json:"points"`     // 聚合数据点
}

// AggPoint 聚合数据点
type AggPoint struct {
	Timestamp time.Time `json:"timestamp"` // 时间戳
	Value     float64   `json:"value"`     // 聚合值
	Count     int64     `json:"count"`     // 数据点数量
}

// WriteBatchRequest 批量写入请求
type WriteBatchRequest struct {
	Database string       `json:"database"` // 数据库名
	Table    string       `json:"table"`    // 表名
	Points   []*DataPoint `json:"points"`   // 数据点列表
}

// WriteResult 写入结果
type WriteResult struct {
	Success int64 `json:"success"` // 成功数量
	Failed  int64 `json:"failed"`  // 失败数量
}

// DownsampleQuery 降采样查询
type DownsampleQuery struct {
	SourceDatabase string        `json:"source_database"` // 源数据库
	SourceTable    string        `json:"source_table"`    // 源表
	TargetDatabase string        `json:"target_database"` // 目标数据库
	TargetTable    string        `json:"target_table"`    // 目标表
	PointIDs       []int64       `json:"point_ids"`       // 测点ID列表
	StartTime      time.Time     `json:"start_time"`      // 开始时间
	EndTime        time.Time     `json:"end_time"`        // 结束时间
	Interval       time.Duration `json:"interval"`        // 降采样间隔
	AggFunc        string        `json:"agg_func"`        // 聚合函数
}

// RetentionPolicy 数据保留策略
type RetentionPolicy struct {
	Name               string        `json:"name"`                // 策略名称
	Database           string        `json:"database"`            // 数据库名
	Duration           time.Duration `json:"duration"`            // 保留时长
	ShardGroupDuration time.Duration `json:"shard_group_duration"` // 分片组时长
	ReplicaN           int           `json:"replica_n"`           // 副本数
	Default            bool          `json:"default"`             // 是否默认策略
}

// DataQuality 数据质量常量
const (
	QualityGood      = 0  // 好数据
	QualityBad       = 1  // 坏数据
	QualityUncertain = 2  // 不确定
	QualityMissing   = 3  // 缺失
)

// QualityDescription 质量描述
func QualityDescription(quality int) string {
	switch quality {
	case QualityGood:
		return "good"
	case QualityBad:
		return "bad"
	case QualityUncertain:
		return "uncertain"
	case QualityMissing:
		return "missing"
	default:
		return "unknown"
	}
}
