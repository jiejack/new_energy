package collector

import (
	"context"
	"time"
)

// CollectorStatus 采集器状态枚举
type CollectorStatus int

const (
	StatusUninitialized CollectorStatus = iota // 未初始化
	StatusInitialized                           // 已初始化
	StatusRunning                               // 运行中
	StatusStopped                               // 已停止
	StatusError                                 // 错误状态
)

func (s CollectorStatus) String() string {
	switch s {
	case StatusUninitialized:
		return "Uninitialized"
	case StatusInitialized:
		return "Initialized"
	case StatusRunning:
		return "Running"
	case StatusStopped:
		return "Stopped"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// CollectorConfig 采集器配置结构
type CollectorConfig struct {
	// 基础配置
	ID          string        `json:"id" yaml:"id"`                   // 采集器唯一标识
	Name        string        `json:"name" yaml:"name"`               // 采集器名称
	Description string        `json:"description" yaml:"description"` // 描述信息
	Protocol    string        `json:"protocol" yaml:"protocol"`       // 协议类型 (modbus, iec104, opcua等)
	
	// 连接配置
	Endpoint    string        `json:"endpoint" yaml:"endpoint"`       // 连接端点
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`         // 超时时间
	RetryCount  int           `json:"retryCount" yaml:"retryCount"`   // 重试次数
	RetryDelay  time.Duration `json:"retryDelay" yaml:"retryDelay"`   // 重试延迟
	
	// 采集配置
	Interval    time.Duration `json:"interval" yaml:"interval"`       // 采集间隔
	BufferSize  int           `json:"bufferSize" yaml:"bufferSize"`   // 缓冲区大小
	BatchSize   int           `json:"batchSize" yaml:"batchSize"`     // 批量大小
	
	// 并发配置
	MaxWorkers  int           `json:"maxWorkers" yaml:"maxWorkers"`   // 最大工作协程数
	MaxPoints   int           `json:"maxPoints" yaml:"maxPoints"`     // 最大采集点数
	
	// 优先级
	Priority    int           `json:"priority" yaml:"priority"`       // 优先级 (1-10, 10最高)
	
	// 标签
	Labels      map[string]string `json:"labels" yaml:"labels"`       // 标签信息
}

// PointData 采集点数据
type PointData struct {
	PointID    string      `json:"pointId"`    // 采集点ID
	Value      interface{} `json:"value"`      // 采集值
	Quality    int         `json:"quality"`    // 数据质量
	Timestamp  time.Time   `json:"timestamp"`  // 时间戳
	Attributes map[string]interface{} `json:"attributes,omitempty"` // 扩展属性
}

// CollectResult 采集结果
type CollectResult struct {
	CollectorID string      `json:"collectorId"` // 采集器ID
	Success     bool        `json:"success"`     // 是否成功
	Error       error       `json:"error"`       // 错误信息
	Data        []PointData `json:"data"`        // 采集数据
	Count       int         `json:"count"`       // 数据点数量
	Duration    time.Duration `json:"duration"`  // 采集耗时
	Timestamp   time.Time   `json:"timestamp"`   // 时间戳
}

// CollectorMetrics 采集器指标
type CollectorMetrics struct {
	TotalCollects     int64         // 总采集次数
	SuccessCollects   int64         // 成功采集次数
	FailedCollects    int64         // 失败采集次数
	TotalPoints       int64         // 总数据点数
	TotalDuration     time.Duration // 总耗时
	AverageDuration   time.Duration // 平均耗时
	LastCollectTime   time.Time     // 最后采集时间
	LastErrorTime     time.Time     // 最后错误时间
	LastError         error         // 最后错误信息
}

// Collector 采集器接口定义
type Collector interface {
	// Initialize 初始化采集器
	// ctx: 上下文，用于控制初始化超时和取消
	// config: 采集器配置
	Initialize(ctx context.Context, config *CollectorConfig) error
	
	// Start 启动采集器
	// ctx: 上下文，用于控制启动超时和取消
	Start(ctx context.Context) error
	
	// Stop 停止采集器
	// ctx: 上下文，用于控制停止超时和取消
	Stop(ctx context.Context) error
	
	// Collect 执行一次数据采集
	// ctx: 上下文，用于控制采集超时和取消
	// 返回采集结果
	Collect(ctx context.Context) (*CollectResult, error)
	
	// GetStatus 获取采集器状态
	GetStatus() CollectorStatus
	
	// GetConfig 获取采集器配置
	GetConfig() *CollectorConfig
	
	// GetMetrics 获取采集器指标
	GetMetrics() *CollectorMetrics
	
	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error
}

// CollectorFactory 采集器工厂函数类型
type CollectorFactory func() Collector

// TaskType 任务类型
type TaskType int

const (
	TaskTypePeriodic TaskType = iota // 周期性任务
	TaskTypeEvent                    // 事件触发任务
	TaskTypeOnce                     // 一次性任务
)

func (t TaskType) String() string {
	switch t {
	case TaskTypePeriodic:
		return "Periodic"
	case TaskTypeEvent:
		return "Event"
	case TaskTypeOnce:
		return "Once"
	default:
		return "Unknown"
	}
}

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota // 待执行
	TaskStatusRunning                   // 执行中
	TaskStatusCompleted                 // 已完成
	TaskStatusFailed                    // 失败
	TaskStatusCancelled                 // 已取消
)

func (s TaskStatus) String() string {
	switch s {
	case TaskStatusPending:
		return "Pending"
	case TaskStatusRunning:
		return "Running"
	case TaskStatusCompleted:
		return "Completed"
	case TaskStatusFailed:
		return "Failed"
	case TaskStatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// Task 采集任务定义
type Task struct {
	ID          string           // 任务ID
	Name        string           // 任务名称
	Type        TaskType         // 任务类型
	Priority    int              // 优先级 (1-10, 10最高)
	CollectorID string           // 采集器ID
	Interval    time.Duration    // 执行间隔 (周期性任务)
	Timeout     time.Duration    // 超时时间
	MaxRetry    int              // 最大重试次数
	Status      TaskStatus       // 任务状态
	CreateTime  time.Time        // 创建时间
	LastRunTime time.Time        // 最后执行时间
	NextRunTime time.Time        // 下次执行时间
	Labels      map[string]string // 标签
}

// TaskResult 任务执行结果
type TaskResult struct {
	TaskID    string         // 任务ID
	Success   bool           // 是否成功
	Error     error          // 错误信息
	Result    *CollectResult // 采集结果
	StartTime time.Time      // 开始时间
	EndTime   time.Time      // 结束时间
	Duration  time.Duration  // 执行耗时
}

// EventHandler 事件处理器接口
type EventHandler interface {
	// OnTaskComplete 任务完成事件
	OnTaskComplete(result *TaskResult)
	
	// OnTaskFailed 任务失败事件
	OnTaskFailed(result *TaskResult)
	
	// OnCollectorError 采集器错误事件
	OnCollectorError(collectorID string, err error)
}

// DataWriter 数据写入器接口
type DataWriter interface {
	// Write 写入数据
	Write(ctx context.Context, data []PointData) error
	
	// WriteBatch 批量写入数据
	WriteBatch(ctx context.Context, batch [][]PointData) error
	
	// Close 关闭写入器
	Close() error
}

// DataReader 数据读取器接口
type DataReader interface {
	// Read 读取数据
	Read(ctx context.Context, pointIDs []string, start, end time.Time) ([]PointData, error)
	
	// ReadLatest 读取最新数据
	ReadLatest(ctx context.Context, pointIDs []string) ([]PointData, error)
	
	// Close 关闭读取器
	Close() error
}
