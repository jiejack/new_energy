// Package metrics 提供应用监控指标定义
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// 全局指标定义
var (
	// API 服务指标
	APIRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "API 请求总数",
		},
		[]string{"service", "method", "endpoint", "status"},
	)

	APIRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_request_duration_seconds",
			Help:    "API 请求持续时间",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method", "endpoint"},
	)

	// 数据采集服务指标
	CollectorDataPointsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "collector_data_points_total",
			Help: "采集的数据点总数",
		},
		[]string{"station_id", "device_id"},
	)

	CollectorErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "collector_errors_total",
			Help: "采集错误总数",
		},
		[]string{"station_id", "device_id", "error_type"},
	)

	CollectorQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "collector_queue_size",
			Help: "采集队列大小",
		},
	)

	CollectorBufferUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "collector_buffer_usage",
			Help: "采集缓冲区使用量",
		},
		[]string{"station_id"},
	)

	// 计算服务指标
	ComputeTasksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "compute_tasks_total",
			Help: "计算任务总数",
		},
		[]string{"task_type", "status"},
	)

	ComputeTaskDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "compute_task_duration_seconds",
			Help:    "计算任务持续时间",
			Buckets: []float64{.01, .05, .1, .5, 1, 2.5, 5, 10, 30, 60},
		},
		[]string{"task_type"},
	)

	ComputeActiveTasks = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "compute_active_tasks",
			Help: "正在执行的计算任务数",
		},
	)

	// 告警服务指标
	AlarmTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alarm_total",
			Help: "告警总数",
		},
		[]string{"station_id", "alarm_level", "alarm_type"},
	)

	AlarmNotificationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alarm_notification_total",
			Help: "告警通知总数",
		},
		[]string{"channel", "status"},
	)

	AlarmNotificationFailedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "alarm_notification_failed_total",
			Help: "告警通知失败总数",
		},
	)

	AlarmActiveCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "alarm_active_count",
			Help: "活跃告警数",
		},
		[]string{"station_id", "alarm_level"},
	)

	// AI 服务指标
	AIRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_requests_total",
			Help: "AI 服务请求总数",
		},
		[]string{"service_type", "model", "status"},
	)

	AIRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_request_duration_seconds",
			Help:    "AI 服务请求持续时间",
			Buckets: []float64{.1, .5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"service_type", "model"},
	)

	AITokensTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_tokens_total",
			Help: "AI 服务使用的 token 总数",
		},
		[]string{"service_type", "model", "token_type"},
	)

	// 数据库连接池指标
	DBConnectionsOpen = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_open",
			Help: "数据库打开的连接数",
		},
	)

	DBConnectionsInUse = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_in_use",
			Help: "数据库正在使用的连接数",
		},
	)

	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "数据库空闲连接数",
		},
	)

	DBWaitCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_wait_count_total",
			Help: "等待连接的总次数",
		},
	)

	DBWaitDuration = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_wait_duration_seconds_total",
			Help: "等待连接的总时间",
		},
	)

	// Redis 缓存指标
	CacheOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "缓存操作总数",
		},
		[]string{"operation", "status"},
	)

	CacheHitRate = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cache_hit_rate",
			Help: "缓存命中率",
		},
	)

	CacheMemoryUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cache_memory_usage_bytes",
			Help: "缓存内存使用量",
		},
	)

	// Kafka 消息队列指标
	KafkaMessagesProduced = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_produced_total",
			Help: "生产的消息总数",
		},
		[]string{"topic"},
	)

	KafkaMessagesConsumed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_consumed_total",
			Help: "消费的消息总数",
		},
		[]string{"topic", "consumer_group"},
	)

	KafkaConsumerLag = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_consumer_lag_messages",
			Help: "消费者延迟消息数",
		},
		[]string{"topic", "consumer_group", "partition"},
	)

	// 业务指标
	StationsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "stations_total",
			Help: "电站总数",
		},
	)

	DevicesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "devices_total",
			Help: "设备总数",
		},
		[]string{"station_id", "device_type"},
	)

	DevicesOnline = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "devices_online",
			Help: "在线设备数",
		},
		[]string{"station_id", "device_type"},
	)

	DataPointsStored = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "data_points_stored_total",
			Help: "存储的数据点总数",
		},
	)

	DataPointsQueried = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "data_points_queried_total",
			Help: "查询的数据点总数",
		},
	)
)
