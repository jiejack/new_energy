package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/new-energy-monitoring/internal/infrastructure/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	ErrMonitorNotRunning   = errors.New("monitor is not running")
	ErrMonitorRunning      = errors.New("monitor is already running")
	ErrAlertNotFound       = errors.New("alert not found")
	ErrInvalidAlertConfig  = errors.New("invalid alert config")
)

// Prometheus指标
var (
	monitorTasksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "statistics_monitor_tasks_total",
		Help: "Total number of monitored tasks",
	}, []string{"status"})

	monitorTaskDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "statistics_monitor_task_duration_seconds",
		Help:    "Task duration histogram",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120, 300},
	}, []string{"task_type"})

	monitorAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "statistics_monitor_alerts_total",
		Help: "Total number of alerts",
	}, []string{"type", "severity"})

	monitorAlertsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "statistics_monitor_alerts_active",
		Help: "Number of active alerts",
	})

	monitorTaskFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "statistics_monitor_task_failures_total",
		Help: "Total number of task failures",
	}, []string{"task_id", "error_type"})

	monitorTaskLatency = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "statistics_monitor_task_latency_seconds",
		Help:       "Task latency summary",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.01, 0.99: 0.001},
	}, []string{"task_type"})
)

// AlertSeverity 告警严重级别
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus 告警状态
type AlertStatus string

const (
	AlertStatusActive   AlertStatus = "active"
	AlertStatusResolved AlertStatus = "resolved"
	AlertStatusSilenced AlertStatus = "silenced"
)

// TaskExecutionStatus 任务执行状态
type TaskExecutionStatus string

const (
	ExecutionStatusSuccess TaskExecutionStatus = "success"
	ExecutionStatusFailed  TaskExecutionStatus = "failed"
	ExecutionStatusTimeout TaskExecutionStatus = "timeout"
	ExecutionStatusSkipped TaskExecutionStatus = "skipped"
)

// TaskExecutionRecord 任务执行记录
type TaskExecutionRecord struct {
	TaskID       string              `json:"taskId"`
	TaskName     string              `json:"taskName"`
	TaskType     string              `json:"taskType"`
	StartTime    time.Time           `json:"startTime"`
	EndTime      time.Time           `json:"endTime"`
	Duration     time.Duration       `json:"duration"`
	Status       TaskExecutionStatus `json:"status"`
	Error        string              `json:"error"`
	RetryCount   int                 `json:"retryCount"`
	ShardIndex   int                 `json:"shardIndex"`
	TotalShards  int                 `json:"totalShards"`
	NodeID       string              `json:"nodeId"`
	Metrics      map[string]interface{} `json:"metrics"`
}

// Alert 告警
type Alert struct {
	ID          string            `json:"id"`
	TaskID      string            `json:"taskId"`
	TaskName    string            `json:"taskName"`
	Type        string            `json:"type"`
	Severity    AlertSeverity     `json:"severity"`
	Status      AlertStatus       `json:"status"`
	Message     string            `json:"message"`
	Details     map[string]interface{} `json:"details"`
	StartTime   time.Time         `json:"startTime"`
	EndTime     *time.Time        `json:"endTime"`
	Count       int               `json:"count"`
	LastSeen    time.Time         `json:"lastSeen"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	TaskPattern     string        `json:"taskPattern"`     // 任务ID匹配模式
	Type            string        `json:"type"`            // failure, timeout, latency, custom
	Severity        AlertSeverity `json:"severity"`
	Threshold       float64       `json:"threshold"`       // 阈值
	Duration        time.Duration `json:"duration"`        // 持续时间
	Count           int           `json:"count"`           // 触发次数
	Enabled         bool          `json:"enabled"`
	CoolDown        time.Duration `json:"coolDown"`        // 冷却时间
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	LastTriggered   time.Time     `json:"lastTriggered"`
}

// TaskMetrics 任务指标
type TaskMetrics struct {
	TaskID            string        `json:"taskId"`
	TaskName          string        `json:"taskName"`
	TaskType          string        `json:"taskType"`
	TotalExecutions   int64         `json:"totalExecutions"`
	SuccessCount      int64         `json:"successCount"`
	FailureCount      int64         `json:"failureCount"`
	TimeoutCount      int64         `json:"timeoutCount"`
	SuccessRate       float64       `json:"successRate"`
	AverageDuration   time.Duration `json:"averageDuration"`
	MinDuration       time.Duration `json:"minDuration"`
	MaxDuration       time.Duration `json:"maxDuration"`
	P50Duration       time.Duration `json:"p50Duration"`
	P95Duration       time.Duration `json:"p95Duration"`
	P99Duration       time.Duration `json:"p99Duration"`
	LastExecutionTime time.Time     `json:"lastExecutionTime"`
	LastStatus        TaskExecutionStatus `json:"lastStatus"`
	LastError        string        `json:"lastError"`
	ConsecutiveFailures int         `json:"consecutiveFailures"`
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	EnableMetrics       bool          `json:"enableMetrics"`       // 启用指标
	EnableAlerts        bool          `json:"enableAlerts"`        // 启用告警
	RecordRetention     time.Duration `json:"recordRetention"`     // 记录保留时间
	MaxRecords          int           `json:"maxRecords"`          // 最大记录数
	MetricsInterval     time.Duration `json:"metricsInterval"`     // 指标采集间隔
	AlertCheckInterval  time.Duration `json:"alertCheckInterval"`  // 告警检查间隔
	AlertCooldown       time.Duration `json:"alertCooldown"`       // 告警冷却时间
	EnableNotifications bool          `json:"enableNotifications"` // 启用通知
}

// DefaultMonitorConfig 默认监控配置
func DefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		EnableMetrics:       true,
		EnableAlerts:        true,
		RecordRetention:     24 * time.Hour,
		MaxRecords:          10000,
		MetricsInterval:     10 * time.Second,
		AlertCheckInterval:  30 * time.Second,
		AlertCooldown:       5 * time.Minute,
		EnableNotifications: true,
	}
}

// TaskMonitor 任务监控器
type TaskMonitor struct {
	config     *MonitorConfig
	running    int32
	closed     int32

	// 执行记录
	records    []*TaskExecutionRecord
	recordMu   sync.RWMutex

	// 任务指标
	metrics    map[string]*TaskMetrics
	metricsMu  sync.RWMutex

	// 告警
	alerts     map[string]*Alert
	alertMu    sync.RWMutex

	// 告警规则
	rules      map[string]*AlertRule
	ruleMu     sync.RWMutex

	// 告警处理器
	alertHandlers []AlertHandler

	// 指标汇总
	summary    *MonitorSummary
	summaryMu  sync.RWMutex

	// 控制
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	// 日志
	logger     *zap.Logger
}

// MonitorSummary 监控汇总
type MonitorSummary struct {
	TotalTasks         int64         `json:"totalTasks"`
	ActiveTasks        int64         `json:"activeTasks"`
	TotalExecutions    int64         `json:"totalExecutions"`
	SuccessExecutions  int64         `json:"successExecutions"`
	FailedExecutions   int64         `json:"failedExecutions"`
	TimeoutExecutions  int64         `json:"timeoutExecutions"`
	OverallSuccessRate float64       `json:"overallSuccessRate"`
	AverageLatency     time.Duration `json:"averageLatency"`
	ActiveAlerts       int           `json:"activeAlerts"`
	LastUpdateTime     time.Time     `json:"lastUpdateTime"`
}

// AlertHandler 告警处理器
type AlertHandler func(alert *Alert) error

// NewTaskMonitor 创建任务监控器
func NewTaskMonitor(config *MonitorConfig) *TaskMonitor {
	if config == nil {
		config = DefaultMonitorConfig()
	}

	m := &TaskMonitor{
		config:        config,
		records:       make([]*TaskExecutionRecord, 0),
		metrics:       make(map[string]*TaskMetrics),
		alerts:        make(map[string]*Alert),
		rules:         make(map[string]*AlertRule),
		alertHandlers: make([]AlertHandler, 0),
		summary:       &MonitorSummary{},
		logger:        logger.Named("task-monitor"),
	}

	m.ctx, m.cancel = context.WithCancel(context.Background())

	// 添加默认告警规则
	m.addDefaultAlertRules()

	return m
}

// addDefaultAlertRules 添加默认告警规则
func (m *TaskMonitor) addDefaultAlertRules() {
	// 任务失败告警
	m.AddAlertRule(&AlertRule{
		ID:          "task-failure",
		Name:        "任务失败告警",
		TaskPattern: "*",
		Type:        "failure",
		Severity:    AlertSeverityError,
		Threshold:   3,
		Count:       3,
		Enabled:     true,
		CoolDown:    5 * time.Minute,
		Labels:      map[string]string{"category": "task"},
	})

	// 任务超时告警
	m.AddAlertRule(&AlertRule{
		ID:          "task-timeout",
		Name:        "任务超时告警",
		TaskPattern: "*",
		Type:        "timeout",
		Severity:    AlertSeverityWarning,
		Threshold:   2,
		Count:       2,
		Enabled:     true,
		CoolDown:    10 * time.Minute,
		Labels:      map[string]string{"category": "task"},
	})

	// 任务延迟告警
	m.AddAlertRule(&AlertRule{
		ID:          "task-latency",
		Name:        "任务延迟告警",
		TaskPattern: "*",
		Type:        "latency",
		Severity:    AlertSeverityWarning,
		Threshold:   30, // 30秒
		Count:       3,
		Enabled:     true,
		CoolDown:    5 * time.Minute,
		Labels:      map[string]string{"category": "performance"},
	})

	// 成功率下降告警
	m.AddAlertRule(&AlertRule{
		ID:          "success-rate-drop",
		Name:        "成功率下降告警",
		TaskPattern: "*",
		Type:        "success_rate",
		Severity:    AlertSeverityWarning,
		Threshold:   0.8, // 80%
		Count:       5,
		Enabled:     true,
		CoolDown:    10 * time.Minute,
		Labels:      map[string]string{"category": "reliability"},
	})
}

// Start 启动监控器
func (m *TaskMonitor) Start() error {
	if atomic.LoadInt32(&m.running) == 1 {
		return ErrMonitorRunning
	}

	atomic.StoreInt32(&m.running, 1)
	atomic.StoreInt32(&m.closed, 0)

	m.logger.Info("Starting task monitor")

	// 启动指标采集
	if m.config.EnableMetrics {
		m.wg.Add(1)
		go m.collectMetrics()
	}

	// 启动告警检查
	if m.config.EnableAlerts {
		m.wg.Add(1)
		go m.checkAlerts()
	}

	// 启动记录清理
	m.wg.Add(1)
	go m.cleanupRecords()

	return nil
}

// Stop 停止监控器
func (m *TaskMonitor) Stop() error {
	if atomic.LoadInt32(&m.running) == 0 {
		return ErrMonitorNotRunning
	}

	m.logger.Info("Stopping task monitor")

	atomic.StoreInt32(&m.running, 0)
	atomic.StoreInt32(&m.closed, 1)

	m.cancel()

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("Task monitor stopped successfully")
		return nil
	case <-time.After(10 * time.Second):
		m.logger.Warn("Task monitor stop timeout")
		return errors.New("stop timeout")
	}
}

// RecordExecution 记录执行
func (m *TaskMonitor) RecordExecution(taskID string, startTime time.Time, duration time.Duration, success bool, result *ExecutionResult) {
	status := ExecutionStatusSuccess
	if !success {
		status = ExecutionStatusFailed
	}

	record := &TaskExecutionRecord{
		TaskID:    taskID,
		StartTime: startTime,
		EndTime:   startTime.Add(duration),
		Duration:  duration,
		Status:    status,
	}

	if result != nil {
		record.TaskName = result.TaskID
		record.Metrics = result.Metrics
		if result.Error != "" {
			record.Error = result.Error
		}
	}

	m.addRecord(record)
	m.updateMetrics(record)

	// 更新Prometheus指标
	if m.config.EnableMetrics {
		statusLabel := "success"
		if !success {
			statusLabel = "failed"
		}
		monitorTasksTotal.WithLabelValues(statusLabel).Inc()
		monitorTaskDuration.WithLabelValues(record.TaskType).Observe(duration.Seconds())
		monitorTaskLatency.WithLabelValues(record.TaskType).Observe(duration.Seconds())
	}
}

// addRecord 添加执行记录
func (m *TaskMonitor) addRecord(record *TaskExecutionRecord) {
	m.recordMu.Lock()
	defer m.recordMu.Unlock()

	m.records = append(m.records, record)

	// 限制记录数量
	if len(m.records) > m.config.MaxRecords {
		m.records = m.records[1:]
	}
}

// updateMetrics 更新指标
func (m *TaskMonitor) updateMetrics(record *TaskExecutionRecord) {
	m.metricsMu.Lock()
	defer m.metricsMu.Unlock()

	metrics, exists := m.metrics[record.TaskID]
	if !exists {
		metrics = &TaskMetrics{
			TaskID:      record.TaskID,
			TaskName:    record.TaskName,
			TaskType:    record.TaskType,
			MinDuration: time.Duration(1<<63 - 1),
		}
		m.metrics[record.TaskID] = metrics
	}

	metrics.TotalExecutions++
	metrics.LastExecutionTime = record.StartTime
	metrics.LastStatus = record.Status

	if record.Status == ExecutionStatusSuccess {
		metrics.SuccessCount++
		metrics.ConsecutiveFailures = 0
	} else if record.Status == ExecutionStatusFailed {
		metrics.FailureCount++
		metrics.ConsecutiveFailures++
		metrics.LastError = record.Error
	} else if record.Status == ExecutionStatusTimeout {
		metrics.TimeoutCount++
		metrics.ConsecutiveFailures++
	}

	// 计算成功率
	if metrics.TotalExecutions > 0 {
		metrics.SuccessRate = float64(metrics.SuccessCount) / float64(metrics.TotalExecutions)
	}

	// 更新持续时间统计
	if record.Duration < metrics.MinDuration {
		metrics.MinDuration = record.Duration
	}
	if record.Duration > metrics.MaxDuration {
		metrics.MaxDuration = record.Duration
	}

	// 计算平均持续时间
	totalDuration := metrics.AverageDuration * time.Duration(metrics.TotalExecutions-1)
	metrics.AverageDuration = (totalDuration + record.Duration) / time.Duration(metrics.TotalExecutions)
}

// collectMetrics 采集指标
func (m *TaskMonitor) collectMetrics() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.doCollectMetrics()

		case <-m.ctx.Done():
			return
		}
	}
}

// doCollectMetrics 执行指标采集
func (m *TaskMonitor) doCollectMetrics() {
	m.metricsMu.RLock()
	defer m.metricsMu.RUnlock()

	var totalExecutions, successExecutions, failedExecutions, timeoutExecutions int64
	var totalLatency time.Duration
	var activeTasks int

	for _, metrics := range m.metrics {
		totalExecutions += metrics.TotalExecutions
		successExecutions += metrics.SuccessCount
		failedExecutions += metrics.FailureCount
		timeoutExecutions += metrics.TimeoutCount
		totalLatency += metrics.AverageDuration
		activeTasks++
	}

	var overallSuccessRate float64
	if totalExecutions > 0 {
		overallSuccessRate = float64(successExecutions) / float64(totalExecutions)
	}

	var averageLatency time.Duration
	if activeTasks > 0 {
		averageLatency = totalLatency / time.Duration(activeTasks)
	}

	m.summaryMu.Lock()
	m.summary.TotalTasks = int64(len(m.metrics))
	m.summary.ActiveTasks = int64(activeTasks)
	m.summary.TotalExecutions = totalExecutions
	m.summary.SuccessExecutions = successExecutions
	m.summary.FailedExecutions = failedExecutions
	m.summary.TimeoutExecutions = timeoutExecutions
	m.summary.OverallSuccessRate = overallSuccessRate
	m.summary.AverageLatency = averageLatency
	m.summary.LastUpdateTime = time.Now()
	m.summaryMu.Unlock()
}

// checkAlerts 检查告警
func (m *TaskMonitor) checkAlerts() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.AlertCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.doCheckAlerts()

		case <-m.ctx.Done():
			return
		}
	}
}

// doCheckAlerts 执行告警检查
func (m *TaskMonitor) doCheckAlerts() {
	m.ruleMu.RLock()
	rules := make([]*AlertRule, 0, len(m.rules))
	for _, rule := range m.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	m.ruleMu.RUnlock()

	m.metricsMu.RLock()
	taskMetrics := make([]*TaskMetrics, 0, len(m.metrics))
	for _, metrics := range m.metrics {
		taskMetrics = append(taskMetrics, metrics)
	}
	m.metricsMu.RUnlock()

	for _, rule := range rules {
		for _, metrics := range taskMetrics {
			m.checkRuleAgainstMetrics(rule, metrics)
		}
	}
}

// checkRuleAgainstMetrics 检查规则与指标
func (m *TaskMonitor) checkRuleAgainstMetrics(rule *AlertRule, metrics *TaskMetrics) {
	// 检查冷却时间
	if time.Since(rule.LastTriggered) < rule.CoolDown {
		return
	}

	var shouldAlert bool
	var message string

	switch rule.Type {
	case "failure":
		if metrics.ConsecutiveFailures >= int(rule.Threshold) {
			shouldAlert = true
			message = fmt.Sprintf("任务 %s 连续失败 %d 次", metrics.TaskID, metrics.ConsecutiveFailures)
		}

	case "timeout":
		if metrics.TimeoutCount >= int64(rule.Threshold) {
			shouldAlert = true
			message = fmt.Sprintf("任务 %s 超时次数达到 %d 次", metrics.TaskID, metrics.TimeoutCount)
		}

	case "latency":
		if metrics.AverageDuration.Seconds() > rule.Threshold {
			shouldAlert = true
			message = fmt.Sprintf("任务 %s 平均延迟 %.2f 秒，超过阈值 %.2f 秒",
				metrics.TaskID, metrics.AverageDuration.Seconds(), rule.Threshold)
		}

	case "success_rate":
		if metrics.TotalExecutions >= int64(rule.Count) && metrics.SuccessRate < rule.Threshold {
			shouldAlert = true
			message = fmt.Sprintf("任务 %s 成功率 %.2f%%，低于阈值 %.2f%%",
				metrics.TaskID, metrics.SuccessRate*100, rule.Threshold*100)
		}
	}

	if shouldAlert {
		m.triggerAlert(rule, metrics, message)
	}
}

// triggerAlert 触发告警
func (m *TaskMonitor) triggerAlert(rule *AlertRule, metrics *TaskMetrics, message string) {
	alertID := fmt.Sprintf("%s-%s", rule.ID, metrics.TaskID)

	m.alertMu.Lock()
	alert, exists := m.alerts[alertID]
	if !exists {
		alert = &Alert{
			ID:        alertID,
			TaskID:    metrics.TaskID,
			TaskName:  metrics.TaskName,
			Type:      rule.Type,
			Severity:  rule.Severity,
			Status:    AlertStatusActive,
			Message:   message,
			StartTime: time.Now(),
			Count:     0,
			LastSeen:  time.Now(),
			Labels:    rule.Labels,
			Details: map[string]interface{}{
				"ruleId":   rule.ID,
				"ruleName": rule.Name,
			},
		}
		m.alerts[alertID] = alert
	}

	alert.Count++
	alert.LastSeen = time.Now()
	alert.Message = message
	m.alertMu.Unlock()

	// 更新规则触发时间
	m.ruleMu.Lock()
	rule.LastTriggered = time.Now()
	m.ruleMu.Unlock()

	// 更新指标
	monitorAlertsTotal.WithLabelValues(rule.Type, string(rule.Severity)).Inc()
	monitorAlertsActive.Set(float64(len(m.alerts)))

	// 更新汇总
	m.summaryMu.Lock()
	m.summary.ActiveAlerts = len(m.alerts)
	m.summaryMu.Unlock()

	// 调用告警处理器
	for _, handler := range m.alertHandlers {
		if err := handler(alert); err != nil {
			m.logger.Error("Alert handler failed",
				zap.String("alertID", alertID),
				zap.Error(err))
		}
	}

	m.logger.Warn("Alert triggered",
		zap.String("alertID", alertID),
		zap.String("taskID", metrics.TaskID),
		zap.String("type", rule.Type),
		zap.String("severity", string(rule.Severity)),
		zap.String("message", message))
}

// AddAlertRule 添加告警规则
func (m *TaskMonitor) AddAlertRule(rule *AlertRule) error {
	if rule == nil || rule.ID == "" {
		return ErrInvalidAlertConfig
	}

	m.ruleMu.Lock()
	defer m.ruleMu.Unlock()

	m.rules[rule.ID] = rule
	m.logger.Info("Alert rule added", zap.String("ruleID", rule.ID))
	return nil
}

// RemoveAlertRule 移除告警规则
func (m *TaskMonitor) RemoveAlertRule(ruleID string) {
	m.ruleMu.Lock()
	defer m.ruleMu.Unlock()

	delete(m.rules, ruleID)
	m.logger.Info("Alert rule removed", zap.String("ruleID", ruleID))
}

// AddAlertHandler 添加告警处理器
func (m *TaskMonitor) AddAlertHandler(handler AlertHandler) {
	m.alertHandlers = append(m.alertHandlers, handler)
}

// ResolveAlert 解决告警
func (m *TaskMonitor) ResolveAlert(alertID string) error {
	m.alertMu.Lock()
	defer m.alertMu.Unlock()

	alert, exists := m.alerts[alertID]
	if !exists {
		return ErrAlertNotFound
	}

	now := time.Now()
	alert.Status = AlertStatusResolved
	alert.EndTime = &now

	m.logger.Info("Alert resolved", zap.String("alertID", alertID))
	return nil
}

// GetAlerts 获取告警列表
func (m *TaskMonitor) GetAlerts(status AlertStatus) []*Alert {
	m.alertMu.RLock()
	defer m.alertMu.RUnlock()

	alerts := make([]*Alert, 0)
	for _, alert := range m.alerts {
		if status == "" || alert.Status == status {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// GetTaskMetrics 获取任务指标
func (m *TaskMonitor) GetTaskMetrics(taskID string) (*TaskMetrics, error) {
	m.metricsMu.RLock()
	defer m.metricsMu.RUnlock()

	metrics, exists := m.metrics[taskID]
	if !exists {
		return nil, errors.New("task metrics not found")
	}
	return metrics, nil
}

// GetAllMetrics 获取所有指标
func (m *TaskMonitor) GetAllMetrics() []*TaskMetrics {
	m.metricsMu.RLock()
	defer m.metricsMu.RUnlock()

	metrics := make([]*TaskMetrics, 0, len(m.metrics))
	for _, m := range m.metrics {
		metrics = append(metrics, m)
	}
	return metrics
}

// GetSummary 获取汇总
func (m *TaskMonitor) GetSummary() *MonitorSummary {
	m.summaryMu.RLock()
	defer m.summaryMu.RUnlock()

	return &MonitorSummary{
		TotalTasks:         m.summary.TotalTasks,
		ActiveTasks:        m.summary.ActiveTasks,
		TotalExecutions:    m.summary.TotalExecutions,
		SuccessExecutions:  m.summary.SuccessExecutions,
		FailedExecutions:   m.summary.FailedExecutions,
		TimeoutExecutions:  m.summary.TimeoutExecutions,
		OverallSuccessRate: m.summary.OverallSuccessRate,
		AverageLatency:     m.summary.AverageLatency,
		ActiveAlerts:       m.summary.ActiveAlerts,
		LastUpdateTime:     m.summary.LastUpdateTime,
	}
}

// GetRecords 获取执行记录
func (m *TaskMonitor) GetRecords(taskID string, limit int) []*TaskExecutionRecord {
	m.recordMu.RLock()
	defer m.recordMu.RUnlock()

	records := make([]*TaskExecutionRecord, 0)
	count := 0

	for i := len(m.records) - 1; i >= 0 && count < limit; i-- {
		if taskID == "" || m.records[i].TaskID == taskID {
			records = append(records, m.records[i])
			count++
		}
	}

	return records
}

// cleanupRecords 清理过期记录
func (m *TaskMonitor) cleanupRecords() {
	defer m.wg.Done()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.doCleanupRecords()

		case <-m.ctx.Done():
			return
		}
	}
}

// doCleanupRecords 执行记录清理
func (m *TaskMonitor) doCleanupRecords() {
	m.recordMu.Lock()
	defer m.recordMu.Unlock()

	now := time.Now()
	validRecords := make([]*TaskExecutionRecord, 0)

	for _, record := range m.records {
		if now.Sub(record.StartTime) < m.config.RecordRetention {
			validRecords = append(validRecords, record)
		}
	}

	m.records = validRecords
}

// IsRunning 检查是否运行中
func (m *TaskMonitor) IsRunning() bool {
	return atomic.LoadInt32(&m.running) == 1
}

// ExportMetrics 导出指标
func (m *TaskMonitor) ExportMetrics() ([]byte, error) {
	summary := m.GetSummary()
	metrics := m.GetAllMetrics()

	export := struct {
		Summary *MonitorSummary `json:"summary"`
		Metrics []*TaskMetrics  `json:"metrics"`
	}{
		Summary: summary,
		Metrics: metrics,
	}

	return json.Marshal(export)
}

// MonitorBuilder 监控器构建器
type MonitorBuilder struct {
	config *MonitorConfig
}

// NewMonitorBuilder 创建监控器构建器
func NewMonitorBuilder() *MonitorBuilder {
	return &MonitorBuilder{
		config: DefaultMonitorConfig(),
	}
}

// WithMetrics 设置指标
func (b *MonitorBuilder) WithMetrics(enable bool) *MonitorBuilder {
	b.config.EnableMetrics = enable
	return b
}

// WithAlerts 设置告警
func (b *MonitorBuilder) WithAlerts(enable bool) *MonitorBuilder {
	b.config.EnableAlerts = enable
	return b
}

// WithRecordRetention 设置记录保留时间
func (b *MonitorBuilder) WithRecordRetention(d time.Duration) *MonitorBuilder {
	b.config.RecordRetention = d
	return b
}

// WithMaxRecords 设置最大记录数
func (b *MonitorBuilder) WithMaxRecords(n int) *MonitorBuilder {
	b.config.MaxRecords = n
	return b
}

// WithMetricsInterval 设置指标采集间隔
func (b *MonitorBuilder) WithMetricsInterval(d time.Duration) *MonitorBuilder {
	b.config.MetricsInterval = d
	return b
}

// WithAlertCheckInterval 设置告警检查间隔
func (b *MonitorBuilder) WithAlertCheckInterval(d time.Duration) *MonitorBuilder {
	b.config.AlertCheckInterval = d
	return b
}

// WithNotifications 设置通知
func (b *MonitorBuilder) WithNotifications(enable bool) *MonitorBuilder {
	b.config.EnableNotifications = enable
	return b
}

// Build 构建监控器
func (b *MonitorBuilder) Build() *TaskMonitor {
	return NewTaskMonitor(b.config)
}
