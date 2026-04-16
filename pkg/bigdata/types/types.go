package types

import (
	"sync"
	"time"
)

// DataPoint 表示一个数据点
type DataPoint struct {
	Timestamp  time.Time              `json:"timestamp"`
	DeviceID   string                 `json:"device_id"`
	Metric     string                 `json:"metric"`
	Value      float64                `json:"value"`
	Tags       map[string]string      `json:"tags,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// BatchData 表示批量数据
type BatchData struct {
	DataPoints []*DataPoint `json:"data_points"`
	Metadata   Metadata     `json:"metadata"`
}

// Metadata 表示数据的元数据
type Metadata struct {
	Source      string                 `json:"source"`
	BatchID     string                 `json:"batch_id"`
	Timestamp   time.Time              `json:"timestamp"`
	RecordCount int                    `json:"record_count"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type           string                 `json:"type"` // clickhouse, doris, influxdb, etc.
	Host           string                 `json:"host"`
	Port           int                    `json:"port"`
	Database       string                 `json:"database"`
	Table          string                 `json:"table"`
	Username       string                 `json:"username"`
	Password       string                 `json:"password"`
	BatchSize      int                    `json:"batch_size"`
	FlushInterval  int                    `json:"flush_interval"`
	Options        map[string]interface{} `json:"options,omitempty"`
}

// AnalysisConfig 分析配置
type AnalysisConfig struct {
	Type       string                 `json:"type"` // spark, flink, etc.
	Master     string                 `json:"master"`
	AppName    string                 `json:"app_name"`
	Executor   int                    `json:"executor"`
	Memory     string                 `json:"memory"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// VisualizationConfig 可视化配置
type VisualizationConfig struct {
	Type     string                 `json:"type"` // grafana, echarts, etc.
	Host     string                 `json:"host"`
	Port     int                    `json:"port"`
	APIKey   string                 `json:"api_key"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// ProcessingConfig 处理配置
type ProcessingConfig struct {
	Type        string                 `json:"type"` // stream, batch
	WindowSize  string                 `json:"window_size"`
	SlideSize   string                 `json:"slide_size"`
	Parallelism int                    `json:"parallelism"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// IngestionConfig 摄取配置
type IngestionConfig struct {
	Type        string                 `json:"type"` // kafka, mqtt, http
	Topic       string                 `json:"topic"`
	Broker      string                 `json:"broker"`
	ConsumerID  string                 `json:"consumer_id"`
	BatchSize   int                    `json:"batch_size"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// Storage interface 存储接口
type Storage interface {
	Init(config StorageConfig) error
	Write(data *BatchData) error
	WritePoint(point *DataPoint) error
	Read(query string) ([]*DataPoint, error)
	ReadTimeRange(startTime, endTime time.Time, stationID, deviceID, metricName string) ([]*DataPoint, error)
	Query(query string) (interface{}, error)
	Aggregate(aggregation, metricName string, startTime, endTime time.Time, groupBy string) (interface{}, error)
	Flush() error
	GetStats() (map[string]interface{}, error)
	Close() error
	// 物化视图相关方法
	CreateMaterializedView(name string, targetTable string, query string) error
	ListMaterializedViews() ([]string, error)
	DropMaterializedView(name string) error
	RefreshMaterializedView(name string) error
	// 查询优化相关方法
	ExplainQuery(query string) (interface{}, error)
	// 预聚合相关方法
	CreatePreAggregationTable(tableName string, timeInterval string) error
	CreatePreAggregationRule(rule interface{}) error
	ListPreAggregationRules() (interface{}, error)
	EnablePreAggregationRule(ruleID string) error
	DisablePreAggregationRule(ruleID string) error
	DeletePreAggregationRule(ruleID string) error
	RefreshPreAggregation(tableName string) error
	// 缓存相关方法
	GetCacheStats() (map[string]interface{}, error)
	ClearCache() error
	// 多维度分析相关方法
	MultiDimensionAggregation(metrics []string, dimensions []string, startTime, endTime time.Time, filters map[string]interface{}) (interface{}, error)
	DimensionDrillDown(baseDimensions []string, drillDownDimension string, metrics []string, startTime, endTime time.Time, filters map[string]interface{}) (interface{}, error)
	DimensionCrossAnalysis(dimensions1 []string, dimensions2 []string, metric string, startTime, endTime time.Time, filters map[string]interface{}) (interface{}, error)
	GetDimensionValues(dimension string, startTime, endTime time.Time, filters map[string]interface{}) (interface{}, error)
}

// Analysis interface 分析接口
type Analysis interface {
	Init(config AnalysisConfig) error
	Execute(query string) (interface{}, error)
	Process(data *BatchData) (interface{}, error)
	Close() error
}

// Visualization interface 可视化接口
type Visualization interface {
	Init(config VisualizationConfig) error
	CreateDashboard(name string, panels []Panel) error
	UpdatePanel(dashboardID, panelID string, data interface{}) error
	Close() error
}

// Panel 表示可视化面板
type Panel struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"` // graph, gauge, table, etc.
	Data        interface{}            `json:"data"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// Processing interface 处理接口
type Processing interface {
	Init(config ProcessingConfig) error
	Process(data *BatchData) (*BatchData, error)
	Close() error
}

// Ingestion interface 摄取接口
type Ingestion interface {
	Init(config IngestionConfig) error
	Start() error
	Stop() error
	Close() error
}

// BigDataService 大数据服务接口
type BigDataService interface {
	Ingest(data *BatchData) error
	Store(data *BatchData) error
	Analyze(query string) (interface{}, error)
	Visualize(dashboardID, panelID string, data interface{}) error
	Process(data *BatchData) (*BatchData, error)
	// 物化视图相关方法
	CreateMaterializedView(name string, targetTable string, query string) error
	ListMaterializedViews() ([]string, error)
	DropMaterializedView(name string) error
	RefreshMaterializedView(name string) error
	// 查询优化相关方法
	ExplainQuery(query string) (interface{}, error)
	// 预聚合相关方法
	CreatePreAggregationTable(tableName string, timeInterval string) error
	CreatePreAggregationRule(rule interface{}) error
	ListPreAggregationRules() (interface{}, error)
	EnablePreAggregationRule(ruleID string) error
	DisablePreAggregationRule(ruleID string) error
	DeletePreAggregationRule(ruleID string) error
	RefreshPreAggregation(tableName string) error
	// 缓存相关方法
	GetCacheStats() (map[string]interface{}, error)
	ClearCache() error
	// 多维度分析相关方法
	MultiDimensionAggregation(metrics []string, dimensions []string, startTime, endTime time.Time, filters map[string]interface{}) (interface{}, error)
	DimensionDrillDown(baseDimensions []string, drillDownDimension string, metrics []string, startTime, endTime time.Time, filters map[string]interface{}) (interface{}, error)
	DimensionCrossAnalysis(dimensions1 []string, dimensions2 []string, metric string, startTime, endTime time.Time, filters map[string]interface{}) (interface{}, error)
	GetDimensionValues(dimension string, startTime, endTime time.Time, filters map[string]interface{}) (interface{}, error)
}

// Error 大数据模块错误
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

// Constants for error codes
const (
	ErrCodeInvalidConfig    = "INVALID_CONFIG"
	ErrCodeStorageError     = "STORAGE_ERROR"
	ErrCodeAnalysisError    = "ANALYSIS_ERROR"
	ErrCodeVisualizationError = "VISUALIZATION_ERROR"
	ErrCodeProcessingError  = "PROCESSING_ERROR"
	ErrCodeIngestionError   = "INGESTION_ERROR"
)

// DataPointPool 数据点对象池
var DataPointPool = &sync.Pool{
	New: func() interface{} {
		return &DataPoint{
			Tags:       make(map[string]string),
			Attributes: make(map[string]interface{}),
		}
	},
}

// BatchDataPool 批量数据对象池
var BatchDataPool = &sync.Pool{
	New: func() interface{} {
		return &BatchData{
			DataPoints: make([]*DataPoint, 0, 100),
			Metadata: Metadata{
				Properties: make(map[string]interface{}),
			},
		}
	},
}

// AcquireDataPoint 从池中获取一个DataPoint对象
func AcquireDataPoint() *DataPoint {
	return DataPointPool.Get().(*DataPoint)
}

// ReleaseDataPoint 将DataPoint对象归还到池中
func ReleaseDataPoint(dp *DataPoint) {
	if dp == nil {
		return
	}
	
	dp.Timestamp = time.Time{}
	dp.DeviceID = ""
	dp.Metric = ""
	dp.Value = 0
	
	for k := range dp.Tags {
		delete(dp.Tags, k)
	}
	
	for k := range dp.Attributes {
		delete(dp.Attributes, k)
	}
	
	DataPointPool.Put(dp)
}

// AcquireBatchData 从池中获取一个BatchData对象
func AcquireBatchData() *BatchData {
	return BatchDataPool.Get().(*BatchData)
}

// ReleaseBatchData 将BatchData对象归还到池中
func ReleaseBatchData(bd *BatchData) {
	if bd == nil {
		return
	}
	
	for _, dp := range bd.DataPoints {
		ReleaseDataPoint(dp)
	}
	
	bd.DataPoints = bd.DataPoints[:0]
	bd.Metadata.Source = ""
	bd.Metadata.BatchID = ""
	bd.Metadata.Timestamp = time.Time{}
	bd.Metadata.RecordCount = 0
	
	for k := range bd.Metadata.Properties {
		delete(bd.Metadata.Properties, k)
	}
	
	BatchDataPool.Put(bd)
}
