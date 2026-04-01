package query

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// MonitorConfig 监控配置
type MonitorConfig struct {
	Enabled             bool          `json:"enabled"`
	SlowQueryThreshold  time.Duration `json:"slow_query_threshold"`
	MaxSlowQueries      int           `json:"max_slow_queries"`
	MaxQueryHistory     int           `json:"max_query_history"`
	StatsInterval       time.Duration `json:"stats_interval"`
	ReportInterval      time.Duration `json:"report_interval"`
	LogSlowQueries      bool          `json:"log_slow_queries"`
	LogFilePath         string        `json:"log_file_path"`
	EnableProfiling     bool          `json:"enable_profiling"`
	CollectQueryPlan    bool          `json:"collect_query_plan"`
	AlertThreshold      time.Duration `json:"alert_threshold"`
}

// DefaultMonitorConfig 默认监控配置
func DefaultMonitorConfig() MonitorConfig {
	return MonitorConfig{
		Enabled:            true,
		SlowQueryThreshold: 1 * time.Second,
		MaxSlowQueries:     1000,
		MaxQueryHistory:    10000,
		StatsInterval:      10 * time.Second,
		ReportInterval:     1 * time.Minute,
		LogSlowQueries:     true,
		LogFilePath:        "./logs/slow_queries.log",
		EnableProfiling:    true,
		CollectQueryPlan:   true,
		AlertThreshold:     5 * time.Second,
	}
}

// QueryMonitor 查询监控器
type QueryMonitor struct {
	config       MonitorConfig
	stats        *QueryStatistics
	slowQueries  *SlowQueryLog
	history      *QueryHistory
	analyzer     *QueryAnalyzer
	reporter     *PerformanceReporter
	alertManager *AlertManager
	mu           sync.RWMutex
	running      bool
}

// QueryStatistics 查询统计
type QueryStatistics struct {
	TotalQueries       int64         `json:"total_queries"`
	SuccessQueries     int64         `json:"success_queries"`
	FailedQueries      int64         `json:"failed_queries"`
	SlowQueries        int64         `json:"slow_queries"`
	CachedQueries      int64         `json:"cached_queries"`
	TotalRows          int64         `json:"total_rows"`
	TotalExecutionTime time.Duration `json:"total_execution_time"`
	AvgExecutionTime   time.Duration `json:"avg_execution_time"`
	MaxExecutionTime   time.Duration `json:"max_execution_time"`
	MinExecutionTime   time.Duration `json:"min_execution_time"`
	P95ExecutionTime   time.Duration `json:"p95_execution_time"`
	P99ExecutionTime   time.Duration `json:"p99_execution_time"`
	QueriesPerSecond   float64       `json:"queries_per_second"`
	StartTime          time.Time     `json:"start_time"`
	LastUpdate         time.Time     `json:"last_update"`

	// 按类型统计
	ByType     map[QueryType]*TypeStatistics `json:"by_type"`
	ByTable    map[string]*TypeStatistics    `json:"by_table"`
	ByPriority map[QueryPriority]*TypeStatistics `json:"by_priority"`

	// 错误统计
	Errors      map[string]int64 `json:"errors"`
	Timeouts    int64            `json:"timeouts"`
	Cancellations int64          `json:"cancellations"`
}

// TypeStatistics 类型统计
type TypeStatistics struct {
	Count        int64         `json:"count"`
	TotalTime    time.Duration `json:"total_time"`
	AvgTime      time.Duration `json:"avg_time"`
	MaxTime      time.Duration `json:"max_time"`
	TotalRows    int64         `json:"total_rows"`
	ErrorCount   int64         `json:"error_count"`
	SlowCount    int64         `json:"slow_count"`
}

// SlowQueryLog 慢查询日志
type SlowQueryLog struct {
	queries  []*SlowQueryEntry
	maxSize  int
	mu       sync.RWMutex
	file     *os.File
	filePath string
}

// SlowQueryEntry 慢查询条目
type SlowQueryEntry struct {
	QueryID       string        `json:"query_id"`
	QueryType     QueryType     `json:"query_type"`
	QueryText     string        `json:"query_text"`
	Database      string        `json:"database"`
	Table         string        `json:"table"`
	ExecutionTime time.Duration `json:"execution_time"`
	RowsReturned  int           `json:"rows_returned"`
	RowsScanned   int64         `json:"rows_scanned"`
	Timestamp     time.Time     `json:"timestamp"`
	User          string        `json:"user"`
	ClientIP      string        `json:"client_ip"`
	Plan          *QueryPlan    `json:"plan,omitempty"`
	Conditions    []QueryCondition `json:"conditions"`
	Error         string        `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// QueryHistory 查询历史
type QueryHistory struct {
	queries  []*QueryHistoryEntry
	maxSize  int
	mu       sync.RWMutex
}

// QueryHistoryEntry 查询历史条目
type QueryHistoryEntry struct {
	QueryID       string        `json:"query_id"`
	QueryType     QueryType     `json:"query_type"`
	QueryText     string        `json:"query_text"`
	Database      string        `json:"database"`
	Table         string        `json:"table"`
	ExecutionTime time.Duration `json:"execution_time"`
	Status        QueryStatus   `json:"status"`
	Timestamp     time.Time     `json:"timestamp"`
	Cached        bool          `json:"cached"`
	RowsReturned  int           `json:"rows_returned"`
}

// QueryAnalyzer 查询分析器
type QueryAnalyzer struct {
	patterns   map[string]*QueryPattern
	suggestions map[string][]string
	mu         sync.RWMutex
}

// QueryPattern 查询模式
type QueryPattern struct {
	Pattern       string        `json:"pattern"`
	Count         int64         `json:"count"`
	AvgTime       time.Duration `json:"avg_time"`
	MaxTime       time.Duration `json:"max_time"`
	TotalTime     time.Duration `json:"total_time"`
	LastSeen      time.Time     `json:"last_seen"`
	CommonTables  []string      `json:"common_tables"`
	Optimization  []string      `json:"optimization"`
}

// PerformanceReporter 性能报告器
type PerformanceReporter struct {
	reports    []*PerformanceReport
	interval   time.Duration
	subscribers []chan *PerformanceReport
	mu         sync.RWMutex
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	GeneratedAt     time.Time     `json:"generated_at"`
	Period          time.Duration `json:"period"`
	TotalQueries    int64         `json:"total_queries"`
	SuccessRate     float64       `json:"success_rate"`
	AvgLatency      time.Duration `json:"avg_latency"`
	P95Latency      time.Duration `json:"p95_latency"`
	P99Latency      time.Duration `json:"p99_latency"`
	QPS             float64       `json:"qps"`
	SlowQueryCount  int64         `json:"slow_query_count"`
	CacheHitRate    float64       `json:"cache_hit_rate"`
	TopSlowQueries  []*SlowQueryEntry `json:"top_slow_queries"`
	TopTables       []TableUsage  `json:"top_tables"`
	TopErrors       []ErrorInfo   `json:"top_errors"`
	Recommendations []string      `json:"recommendations"`
	HealthScore     float64       `json:"health_score"`
}

// TableUsage 表使用情况
type TableUsage struct {
	Table     string  `json:"table"`
	Count     int64   `json:"count"`
	AvgTime   time.Duration `json:"avg_time"`
	SlowCount int64   `json:"slow_count"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Error string `json:"error"`
	Count int64  `json:"count"`
}

// AlertManager 告警管理器
type AlertManager struct {
	threshold  time.Duration
	alerts     []*Alert
	subscribers []chan *Alert
	mu         sync.RWMutex
}

// Alert 告警
type Alert struct {
	ID        string        `json:"id"`
	Type      string        `json:"type"`
	Level     string        `json:"level"` // info, warning, critical
	Message   string        `json:"message"`
	QueryID   string        `json:"query_id"`
	Timestamp time.Time     `json:"timestamp"`
	Value     time.Duration `json:"value"`
	Threshold time.Duration `json:"threshold"`
	Resolved  bool          `json:"resolved"`
}

// NewQueryMonitor 创建查询监控器
func NewQueryMonitor(config MonitorConfig) *QueryMonitor {
	monitor := &QueryMonitor{
		config:   config,
		stats:    NewQueryStatistics(),
		analyzer: NewQueryAnalyzer(),
		reporter: NewPerformanceReporter(config.ReportInterval),
	}

	monitor.slowQueries = NewSlowQueryLog(config.MaxSlowQueries, config.LogFilePath)
	monitor.history = NewQueryHistory(config.MaxQueryHistory)
	monitor.alertManager = NewAlertManager(config.AlertThreshold)

	return monitor
}

// NewQueryStatistics 创建查询统计
func NewQueryStatistics() *QueryStatistics {
	return &QueryStatistics{
		StartTime:      time.Now(),
		LastUpdate:     time.Now(),
		MinExecutionTime: time.Hour, // 初始化为很大的值
		ByType:         make(map[QueryType]*TypeStatistics),
		ByTable:        make(map[string]*TypeStatistics),
		ByPriority:     make(map[QueryPriority]*TypeStatistics),
		Errors:         make(map[string]int64),
	}
}

// NewSlowQueryLog 创建慢查询日志
func NewSlowQueryLog(maxSize int, filePath string) *SlowQueryLog {
	log := &SlowQueryLog{
		queries:  make([]*SlowQueryEntry, 0),
		maxSize:  maxSize,
		filePath: filePath,
	}

	// 创建日志文件
	if filePath != "" {
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err == nil {
			file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				log.file = file
			}
		}
	}

	return log
}

// NewQueryHistory 创建查询历史
func NewQueryHistory(maxSize int) *QueryHistory {
	return &QueryHistory{
		queries: make([]*QueryHistoryEntry, 0),
		maxSize: maxSize,
	}
}

// NewQueryAnalyzer 创建查询分析器
func NewQueryAnalyzer() *QueryAnalyzer {
	return &QueryAnalyzer{
		patterns:    make(map[string]*QueryPattern),
		suggestions: make(map[string][]string),
	}
}

// NewPerformanceReporter 创建性能报告器
func NewPerformanceReporter(interval time.Duration) *PerformanceReporter {
	return &PerformanceReporter{
		reports:    make([]*PerformanceReport, 0),
		interval:   interval,
		subscribers: make([]chan *PerformanceReport, 0),
	}
}

// NewAlertManager 创建告警管理器
func NewAlertManager(threshold time.Duration) *AlertManager {
	return &AlertManager{
		threshold:   threshold,
		alerts:      make([]*Alert, 0),
		subscribers: make([]chan *Alert, 0),
	}
}

// Start 启动监控
func (m *QueryMonitor) Start(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	// 启动定时任务
	go m.statsCollector(ctx)
	go m.reportGenerator(ctx)
}

// Stop 停止监控
func (m *QueryMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false

	if m.slowQueries.file != nil {
		m.slowQueries.file.Close()
	}
}

// RecordQuery 记录查询
func (m *QueryMonitor) RecordQuery(req *QueryRequest, result *QueryResult, err error) {
	if !m.config.Enabled {
		return
	}

	// 更新统计
	m.updateStats(req, result, err)

	// 检查慢查询
	if result != nil && result.ExecutionTime > m.config.SlowQueryThreshold {
		m.recordSlowQuery(req, result)
	}

	// 记录历史
	m.recordHistory(req, result)

	// 分析查询
	m.analyzer.Analyze(req, result)

	// 检查告警
	if result != nil && result.ExecutionTime > m.config.AlertThreshold {
		m.alertManager.CheckAndAlert(req, result)
	}
}

// updateStats 更新统计
func (m *QueryMonitor) updateStats(req *QueryRequest, result *QueryResult, err error) {
	stats := m.stats

	// 更新总数
	atomic.AddInt64(&stats.TotalQueries, 1)

	if err != nil {
		atomic.AddInt64(&stats.FailedQueries, 1)
		// 记录错误
		errorKey := err.Error()
		if len(errorKey) > 100 {
			errorKey = errorKey[:100]
		}
		m.mu.Lock()
		stats.Errors[errorKey]++
		m.mu.Unlock()
	} else {
		atomic.AddInt64(&stats.SuccessQueries, 1)
	}

	// 更新执行时间
	if result != nil {
		execTime := result.ExecutionTime

		// 总执行时间
		for {
			old := atomic.LoadInt64((*int64)(&stats.TotalExecutionTime))
			newVal := old + int64(execTime)
			if atomic.CompareAndSwapInt64((*int64)(&stats.TotalExecutionTime), old, newVal) {
				break
			}
		}

		// 最大执行时间
		for {
			old := atomic.LoadInt64((*int64)(&stats.MaxExecutionTime))
			if int64(execTime) <= old {
				break
			}
			if atomic.CompareAndSwapInt64((*int64)(&stats.MaxExecutionTime), old, int64(execTime)) {
				break
			}
		}

		// 最小执行时间
		m.mu.Lock()
		if execTime < stats.MinExecutionTime {
			stats.MinExecutionTime = execTime
		}
		m.mu.Unlock()

		// 更新行数
		atomic.AddInt64(&stats.TotalRows, int64(len(result.Data)))

		// 缓存命中
		if result.Cached {
			atomic.AddInt64(&stats.CachedQueries, 1)
		}

		// 更新类型统计
		m.updateTypeStats(req, result, err)
	}

	// 更新平均执行时间
	total := atomic.LoadInt64(&stats.TotalQueries)
	if total > 0 {
		totalTime := atomic.LoadInt64((*int64)(&stats.TotalExecutionTime))
		stats.AvgExecutionTime = time.Duration(totalTime / total)
	}

	stats.LastUpdate = time.Now()
}

// updateTypeStats 更新类型统计
func (m *QueryMonitor) updateTypeStats(req *QueryRequest, result *QueryResult, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := m.stats

	// 按类型统计
	if _, ok := stats.ByType[req.Type]; !ok {
		stats.ByType[req.Type] = &TypeStatistics{}
	}
	typeStats := stats.ByType[req.Type]
	typeStats.Count++
	if result != nil {
		typeStats.TotalTime += result.ExecutionTime
		typeStats.AvgTime = time.Duration(int64(typeStats.TotalTime) / typeStats.Count)
		if result.ExecutionTime > typeStats.MaxTime {
			typeStats.MaxTime = result.ExecutionTime
		}
		typeStats.TotalRows += int64(len(result.Data))
		if result.ExecutionTime > m.config.SlowQueryThreshold {
			typeStats.SlowCount++
		}
	}
	if err != nil {
		typeStats.ErrorCount++
	}

	// 按表统计
	if req.Table != "" {
		if _, ok := stats.ByTable[req.Table]; !ok {
			stats.ByTable[req.Table] = &TypeStatistics{}
		}
		tableStats := stats.ByTable[req.Table]
		tableStats.Count++
		if result != nil {
			tableStats.TotalTime += result.ExecutionTime
			tableStats.AvgTime = time.Duration(int64(tableStats.TotalTime) / tableStats.Count)
			if result.ExecutionTime > tableStats.MaxTime {
				tableStats.MaxTime = result.ExecutionTime
			}
			if result.ExecutionTime > m.config.SlowQueryThreshold {
				tableStats.SlowCount++
			}
		}
	}

	// 按优先级统计
	if _, ok := stats.ByPriority[req.Priority]; !ok {
		stats.ByPriority[req.Priority] = &TypeStatistics{}
	}
	priorityStats := stats.ByPriority[req.Priority]
	priorityStats.Count++
	if result != nil {
		priorityStats.TotalTime += result.ExecutionTime
		priorityStats.AvgTime = time.Duration(int64(priorityStats.TotalTime) / priorityStats.Count)
	}
}

// recordSlowQuery 记录慢查询
func (m *QueryMonitor) recordSlowQuery(req *QueryRequest, result *QueryResult) {
	atomic.AddInt64(&m.stats.SlowQueries, 1)

	entry := &SlowQueryEntry{
		QueryID:       req.ID,
		QueryType:     req.Type,
		QueryText:     m.buildQueryText(req),
		Database:      req.Database,
		Table:         req.Table,
		ExecutionTime: result.ExecutionTime,
		RowsReturned:  len(result.Data),
		Timestamp:     time.Now(),
		Conditions:    req.Conditions,
	}

	m.slowQueries.Add(entry)
}

// recordHistory 记录历史
func (m *QueryMonitor) recordHistory(req *QueryRequest, result *QueryResult) {
	if result == nil {
		return
	}

	entry := &QueryHistoryEntry{
		QueryID:       req.ID,
		QueryType:     req.Type,
		QueryText:     m.buildQueryText(req),
		Database:      req.Database,
		Table:         req.Table,
		ExecutionTime: result.ExecutionTime,
		Status:        result.Status,
		Timestamp:     time.Now(),
		Cached:        result.Cached,
		RowsReturned:  len(result.Data),
	}

	m.history.Add(entry)
}

// buildQueryText 构建查询文本
func (m *QueryMonitor) buildQueryText(req *QueryRequest) string {
	text := fmt.Sprintf("SELECT %v FROM %s", req.Fields, req.Table)

	if len(req.Conditions) > 0 {
		text += " WHERE "
		for i, cond := range req.Conditions {
			if i > 0 {
				text += " AND "
			}
			text += fmt.Sprintf("%s %s %v", cond.Field, cond.Operator, cond.Value)
		}
	}

	if len(req.GroupBy) > 0 {
		text += fmt.Sprintf(" GROUP BY %v", req.GroupBy)
	}

	if len(req.OrderBy) > 0 {
		text += " ORDER BY "
		for i, order := range req.OrderBy {
			if i > 0 {
				text += ", "
			}
			text += order.Field
			if order.Desc {
				text += " DESC"
			}
		}
	}

	if req.Limit > 0 {
		text += fmt.Sprintf(" LIMIT %d", req.Limit)
	}

	return text
}

// GetStats 获取统计信息
func (m *QueryMonitor) GetStats() *QueryStatistics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 计算QPS
	elapsed := time.Since(m.stats.StartTime).Seconds()
	if elapsed > 0 {
		m.stats.QueriesPerSecond = float64(m.stats.TotalQueries) / elapsed
	}

	// 计算百分位
	m.calculatePercentiles()

	return m.stats
}

// calculatePercentiles 计算百分位
func (m *QueryMonitor) calculatePercentiles() {
	// 从历史数据计算P95和P99
	times := m.history.GetExecutionTimes()
	if len(times) == 0 {
		return
	}

	// 排序
	for i := 0; i < len(times)-1; i++ {
		for j := i + 1; j < len(times); j++ {
			if times[i] > times[j] {
				times[i], times[j] = times[j], times[i]
			}
		}
	}

	// P95
	p95Idx := int(float64(len(times)) * 0.95)
	if p95Idx >= len(times) {
		p95Idx = len(times) - 1
	}
	m.stats.P95ExecutionTime = times[p95Idx]

	// P99
	p99Idx := int(float64(len(times)) * 0.99)
	if p99Idx >= len(times) {
		p99Idx = len(times) - 1
	}
	m.stats.P99ExecutionTime = times[p99Idx]
}

// GetSlowQueries 获取慢查询
func (m *QueryMonitor) GetSlowQueries(limit int) []*SlowQueryEntry {
	return m.slowQueries.Get(limit)
}

// GetHistory 获取查询历史
func (m *QueryMonitor) GetHistory(limit int) []*QueryHistoryEntry {
	return m.history.Get(limit)
}

// GenerateReport 生成报告
func (m *QueryMonitor) GenerateReport() *PerformanceReport {
	stats := m.GetStats()

	report := &PerformanceReport{
		GeneratedAt:    time.Now(),
		Period:         time.Since(stats.StartTime),
		TotalQueries:   stats.TotalQueries,
		AvgLatency:     stats.AvgExecutionTime,
		P95Latency:     stats.P95ExecutionTime,
		P99Latency:     stats.P99ExecutionTime,
		QPS:            stats.QueriesPerSecond,
		SlowQueryCount: stats.SlowQueries,
	}

	// 计算成功率
	if stats.TotalQueries > 0 {
		report.SuccessRate = float64(stats.SuccessQueries) / float64(stats.TotalQueries)
	}

	// 计算缓存命中率
	if stats.TotalQueries > 0 {
		report.CacheHitRate = float64(stats.CachedQueries) / float64(stats.TotalQueries)
	}

	// 获取Top慢查询
	report.TopSlowQueries = m.slowQueries.Get(10)

	// 获取Top表
	report.TopTables = m.getTopTables(10)

	// 获取Top错误
	report.TopErrors = m.getTopErrors(10)

	// 生成建议
	report.Recommendations = m.generateRecommendations(stats)

	// 计算健康分数
	report.HealthScore = m.calculateHealthScore(stats, report)

	m.reporter.AddReport(report)

	return report
}

// getTopTables 获取Top表
func (m *QueryMonitor) getTopTables(limit int) []TableUsage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var tables []TableUsage
	for table, stats := range m.stats.ByTable {
		tables = append(tables, TableUsage{
			Table:     table,
			Count:     stats.Count,
			AvgTime:   stats.AvgTime,
			SlowCount: stats.SlowCount,
		})
	}

	// 按计数排序
	for i := 0; i < len(tables)-1; i++ {
		for j := i + 1; j < len(tables); j++ {
			if tables[i].Count < tables[j].Count {
				tables[i], tables[j] = tables[j], tables[i]
			}
		}
	}

	if len(tables) > limit {
		tables = tables[:limit]
	}

	return tables
}

// getTopErrors 获取Top错误
func (m *QueryMonitor) getTopErrors(limit int) []ErrorInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errors []ErrorInfo
	for err, count := range m.stats.Errors {
		errors = append(errors, ErrorInfo{
			Error: err,
			Count: count,
		})
	}

	// 按计数排序
	for i := 0; i < len(errors)-1; i++ {
		for j := i + 1; j < len(errors); j++ {
			if errors[i].Count < errors[j].Count {
				errors[i], errors[j] = errors[j], errors[i]
			}
		}
	}

	if len(errors) > limit {
		errors = errors[:limit]
	}

	return errors
}

// generateRecommendations 生成建议
func (m *QueryMonitor) generateRecommendations(stats *QueryStatistics) []string {
	var recommendations []string

	// 检查慢查询比例
	if stats.TotalQueries > 0 {
		slowRate := float64(stats.SlowQueries) / float64(stats.TotalQueries)
		if slowRate > 0.1 {
			recommendations = append(recommendations,
				fmt.Sprintf("慢查询比例过高(%.2f%%)，建议优化查询或添加索引", slowRate*100))
		}
	}

	// 检查平均延迟
	if stats.AvgExecutionTime > 500*time.Millisecond {
		recommendations = append(recommendations,
			"平均查询延迟较高，建议检查数据库性能或优化查询")
	}

	// 检查错误率
	if stats.TotalQueries > 0 {
		errorRate := float64(stats.FailedQueries) / float64(stats.TotalQueries)
		if errorRate > 0.05 {
			recommendations = append(recommendations,
				fmt.Sprintf("错误率过高(%.2f%%)，建议检查错误原因", errorRate*100))
		}
	}

	// 检查缓存命中率
	if stats.TotalQueries > 0 {
		cacheRate := float64(stats.CachedQueries) / float64(stats.TotalQueries)
		if cacheRate < 0.3 {
			recommendations = append(recommendations,
				"缓存命中率较低，建议增加缓存或调整缓存策略")
		}
	}

	// 检查特定表的慢查询
	for table, tableStats := range stats.ByTable {
		if tableStats.Count > 100 && tableStats.SlowCount > 10 {
			recommendations = append(recommendations,
				fmt.Sprintf("表 %s 存在较多慢查询，建议检查索引或优化查询", table))
		}
	}

	return recommendations
}

// calculateHealthScore 计算健康分数
func (m *QueryMonitor) calculateHealthScore(stats *QueryStatistics, report *PerformanceReport) float64 {
	score := 100.0

	// 成功率影响
	score -= (1 - report.SuccessRate) * 30

	// 慢查询影响
	if stats.TotalQueries > 0 {
		slowRate := float64(stats.SlowQueries) / float64(stats.TotalQueries)
		score -= slowRate * 20
	}

	// 平均延迟影响
	if stats.AvgExecutionTime > 100*time.Millisecond {
		penalty := float64(stats.AvgExecutionTime.Milliseconds()) / 1000 * 10
		score -= penalty
	}

	// 错误率影响
	if stats.TotalQueries > 0 {
		errorRate := float64(stats.FailedQueries) / float64(stats.TotalQueries)
		score -= errorRate * 40
	}

	// 确保分数在0-100之间
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// statsCollector 统计收集器
func (m *QueryMonitor) statsCollector(ctx context.Context) {
	ticker := time.NewTicker(m.config.StatsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mu.Lock()
			if !m.running {
				m.mu.Unlock()
				return
			}
			m.mu.Unlock()
		}
	}
}

// reportGenerator 报告生成器
func (m *QueryMonitor) reportGenerator(ctx context.Context) {
	ticker := time.NewTicker(m.config.ReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mu.Lock()
			if !m.running {
				m.mu.Unlock()
				return
			}
			m.mu.Unlock()

			report := m.GenerateReport()
			m.reporter.Notify(report)
		}
	}
}

// Add 添加慢查询
func (l *SlowQueryLog) Add(entry *SlowQueryEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.queries = append(l.queries, entry)

	// 超过最大数量，移除最旧的
	if len(l.queries) > l.maxSize {
		l.queries = l.queries[1:]
	}

	// 写入文件
	if l.file != nil {
		data, _ := json.Marshal(entry)
		l.file.WriteString(string(data) + "\n")
	}
}

// Get 获取慢查询
func (l *SlowQueryLog) Get(limit int) []*SlowQueryEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if limit <= 0 || limit > len(l.queries) {
		limit = len(l.queries)
	}

	// 返回最新的
	start := len(l.queries) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*SlowQueryEntry, limit)
	copy(result, l.queries[start:])

	// 反转顺序，最新的在前
	for i := 0; i < len(result)/2; i++ {
		j := len(result) - 1 - i
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// Add 添加历史记录
func (h *QueryHistory) Add(entry *QueryHistoryEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.queries = append(h.queries, entry)

	if len(h.queries) > h.maxSize {
		h.queries = h.queries[1:]
	}
}

// Get 获取历史记录
func (h *QueryHistory) Get(limit int) []*QueryHistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if limit <= 0 || limit > len(h.queries) {
		limit = len(h.queries)
	}

	start := len(h.queries) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*QueryHistoryEntry, limit)
	copy(result, h.queries[start:])

	return result
}

// GetExecutionTimes 获取执行时间列表
func (h *QueryHistory) GetExecutionTimes() []time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()

	times := make([]time.Duration, len(h.queries))
	for i, entry := range h.queries {
		times[i] = entry.ExecutionTime
	}
	return times
}

// Analyze 分析查询
func (a *QueryAnalyzer) Analyze(req *QueryRequest, result *QueryResult) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 生成查询模式
	pattern := a.generatePattern(req)

	if _, ok := a.patterns[pattern]; !ok {
		a.patterns[pattern] = &QueryPattern{
			Pattern: pattern,
		}
	}

	p := a.patterns[pattern]
	p.Count++
	p.LastSeen = time.Now()

	if result != nil {
		p.TotalTime += result.ExecutionTime
		p.AvgTime = time.Duration(int64(p.TotalTime) / p.Count)
		if result.ExecutionTime > p.MaxTime {
			p.MaxTime = result.ExecutionTime
		}
	}

	// 更新常用表
	if req.Table != "" {
		found := false
		for _, t := range p.CommonTables {
			if t == req.Table {
				found = true
				break
			}
		}
		if !found {
			p.CommonTables = append(p.CommonTables, req.Table)
		}
	}

	// 生成优化建议
	a.generateOptimization(pattern, p)
}

// generatePattern 生成查询模式
func (a *QueryAnalyzer) generatePattern(req *QueryRequest) string {
	return fmt.Sprintf("%s:%s:%d", req.Type, req.Table, len(req.Conditions))
}

// generateOptimization 生成优化建议
func (a *QueryAnalyzer) generateOptimization(pattern string, p *QueryPattern) {
	var suggestions []string

	// 检查平均执行时间
	if p.AvgTime > 500*time.Millisecond {
		suggestions = append(suggestions, "考虑添加索引或优化查询条件")
	}

	// 检查是否有大量条件
	if len(p.CommonTables) > 3 {
		suggestions = append(suggestions, "查询涉及多个表，考虑优化JOIN策略")
	}

	p.Optimization = suggestions
}

// GetPattern 获取查询模式
func (a *QueryAnalyzer) GetPattern(pattern string) (*QueryPattern, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	p, ok := a.patterns[pattern]
	return p, ok
}

// GetAllPatterns 获取所有模式
func (a *QueryAnalyzer) GetAllPatterns() map[string]*QueryPattern {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]*QueryPattern)
	for k, v := range a.patterns {
		result[k] = v
	}
	return result
}

// AddReport 添加报告
func (r *PerformanceReporter) AddReport(report *PerformanceReport) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.reports = append(r.reports, report)

	// 只保留最近10个报告
	if len(r.reports) > 10 {
		r.reports = r.reports[1:]
	}
}

// Subscribe 订阅报告
func (r *PerformanceReporter) Subscribe() <-chan *PerformanceReport {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch := make(chan *PerformanceReport, 10)
	r.subscribers = append(r.subscribers, ch)
	return ch
}

// Notify 通知订阅者
func (r *PerformanceReporter) Notify(report *PerformanceReport) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, ch := range r.subscribers {
		select {
		case ch <- report:
		default:
			// 通道满，跳过
		}
	}
}

// GetLatestReport 获取最新报告
func (r *PerformanceReporter) GetLatestReport() *PerformanceReport {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.reports) == 0 {
		return nil
	}
	return r.reports[len(r.reports)-1]
}

// CheckAndAlert 检查并发送告警
func (a *AlertManager) CheckAndAlert(req *QueryRequest, result *QueryResult) {
	if result.ExecutionTime <= a.threshold {
		return
	}

	alert := &Alert{
		ID:        fmt.Sprintf("alert-%d", time.Now().UnixNano()),
		Type:      "slow_query",
		Level:     "warning",
		Message:   fmt.Sprintf("查询执行时间 %.2fs 超过阈值 %.2fs", result.ExecutionTime.Seconds(), a.threshold.Seconds()),
		QueryID:   req.ID,
		Timestamp: time.Now(),
		Value:     result.ExecutionTime,
		Threshold: a.threshold,
	}

	a.mu.Lock()
	a.alerts = append(a.alerts, alert)
	// 只保留最近100个告警
	if len(a.alerts) > 100 {
		a.alerts = a.alerts[1:]
	}
	subscribers := make([]chan *Alert, len(a.subscribers))
	copy(subscribers, a.subscribers)
	a.mu.Unlock()

	// 通知订阅者
	for _, ch := range subscribers {
		select {
		case ch <- alert:
		default:
		}
	}
}

// Subscribe 订阅告警
func (a *AlertManager) Subscribe() <-chan *Alert {
	a.mu.Lock()
	defer a.mu.Unlock()

	ch := make(chan *Alert, 10)
	a.subscribers = append(a.subscribers, ch)
	return ch
}

// GetAlerts 获取告警列表
func (a *AlertManager) GetAlerts(limit int) []*Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if limit <= 0 || limit > len(a.alerts) {
		limit = len(a.alerts)
	}

	start := len(a.alerts) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*Alert, limit)
	copy(result, a.alerts[start:])

	return result
}

// QueryMonitorMiddleware 查询监控中间件
type QueryMonitorMiddleware struct {
	monitor *QueryMonitor
}

// NewQueryMonitorMiddleware 创建监控中间件
func NewQueryMonitorMiddleware(monitor *QueryMonitor) *QueryMonitorMiddleware {
	return &QueryMonitorMiddleware{
		monitor: monitor,
	}
}

// Before 查询前处理
func (m *QueryMonitorMiddleware) Before(ctx context.Context, req *QueryRequest) context.Context {
	// 记录开始时间
	return context.WithValue(ctx, "query_start_time", time.Now())
}

// After 查询后处理
func (m *QueryMonitorMiddleware) After(ctx context.Context, req *QueryRequest, result *QueryResult, err error) {
	m.monitor.RecordQuery(req, result, err)
}

// MetricsExporter 指标导出器
type MetricsExporter struct {
	monitor *QueryMonitor
}

// NewMetricsExporter 创建指标导出器
func NewMetricsExporter(monitor *QueryMonitor) *MetricsExporter {
	return &MetricsExporter{
		monitor: monitor,
	}
}

// ExportPrometheus 导出Prometheus格式指标
func (e *MetricsExporter) ExportPrometheus() string {
	stats := e.monitor.GetStats()

	var metrics string

	// 总查询数
	metrics += fmt.Sprintf("# HELP query_total Total number of queries\n")
	metrics += fmt.Sprintf("# TYPE query_total counter\n")
	metrics += fmt.Sprintf("query_total %d\n", stats.TotalQueries)

	// 成功查询数
	metrics += fmt.Sprintf("# HELP query_success_total Total number of successful queries\n")
	metrics += fmt.Sprintf("# TYPE query_success_total counter\n")
	metrics += fmt.Sprintf("query_success_total %d\n", stats.SuccessQueries)

	// 失败查询数
	metrics += fmt.Sprintf("# HELP query_failed_total Total number of failed queries\n")
	metrics += fmt.Sprintf("# TYPE query_failed_total counter\n")
	metrics += fmt.Sprintf("query_failed_total %d\n", stats.FailedQueries)

	// 慢查询数
	metrics += fmt.Sprintf("# HELP query_slow_total Total number of slow queries\n")
	metrics += fmt.Sprintf("# TYPE query_slow_total counter\n")
	metrics += fmt.Sprintf("query_slow_total %d\n", stats.SlowQueries)

	// 平均执行时间
	metrics += fmt.Sprintf("# HELP query_duration_avg Average query duration in seconds\n")
	metrics += fmt.Sprintf("# TYPE query_duration_avg gauge\n")
	metrics += fmt.Sprintf("query_duration_avg %f\n", stats.AvgExecutionTime.Seconds())

	// QPS
	metrics += fmt.Sprintf("# HELP query_qps Queries per second\n")
	metrics += fmt.Sprintf("# TYPE query_qps gauge\n")
	metrics += fmt.Sprintf("query_qps %f\n", stats.QueriesPerSecond)

	return metrics
}
