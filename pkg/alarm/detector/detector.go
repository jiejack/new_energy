package detector

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/pkg/alarm/aggregator"
	"github.com/new-energy-monitoring/pkg/alarm/dedup"
	"github.com/new-energy-monitoring/pkg/alarm/state"
)

// RuleType 规则类型
type RuleType int

const (
	RuleTypeThreshold RuleType = iota
	RuleTypeRange
	RuleTypeRate
	RuleTypeDeviation
	RuleTypeDuration
	RuleTypeExpression
)

func (t RuleType) String() string {
	switch t {
	case RuleTypeThreshold:
		return "threshold"
	case RuleTypeRange:
		return "range"
	case RuleTypeRate:
		return "rate"
	case RuleTypeDeviation:
		return "deviation"
	case RuleTypeDuration:
		return "duration"
	case RuleTypeExpression:
		return "expression"
	default:
		return "unknown"
	}
}

// Operator 比较操作符
type Operator int

const (
	OpEqual Operator = iota
	OpNotEqual
	OpGreaterThan
	OpGreaterEqual
	OpLessThan
	OpLessEqual
)

func (o Operator) String() string {
	switch o {
	case OpEqual:
		return "=="
	case OpNotEqual:
		return "!="
	case OpGreaterThan:
		return ">"
	case OpGreaterEqual:
		return ">="
	case OpLessThan:
		return "<"
	case OpLessEqual:
		return "<="
	default:
		return "?"
	}
}

// Rule 检测规则
type Rule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        RuleType          `json:"type"`
	Enabled     bool              `json:"enabled"`
	PointIDs    []string          `json:"point_ids"`
	DeviceIDs   []string          `json:"device_ids"`
	StationIDs  []string          `json:"station_ids"`
	Level       entity.AlarmLevel `json:"level"`
	AlarmType   entity.AlarmType  `json:"alarm_type"`
	Title       string            `json:"title"`
	Message     string            `json:"message"`
	
	// 阈值规则参数
	Operator    Operator          `json:"operator"`
	Threshold   float64           `json:"threshold"`
	
	// 范围规则参数
	MinValue    float64           `json:"min_value"`
	MaxValue    float64           `json:"max_value"`
	
	// 变化率规则参数
	RateThreshold float64         `json:"rate_threshold"`
	RateWindow    time.Duration   `json:"rate_window"`
	
	// 偏差规则参数
	DeviationThreshold float64    `json:"deviation_threshold"`
	BaselineWindow     time.Duration `json:"baseline_window"`
	
	// 持续时间规则参数
	DurationThreshold time.Duration `json:"duration_threshold"`
	
	// 表达式规则参数
	Expression  string            `json:"expression"`
	
	// 滑动窗口参数
	WindowDuration time.Duration  `json:"window_duration"`
	WindowCount    int            `json:"window_count"`
	
	// 抑制参数
	SuppressDuration time.Duration `json:"suppress_duration"`
	
	// 元数据
	Metadata    map[string]any    `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// NewRule 创建规则
func NewRule(id, name string, ruleType RuleType) *Rule {
	return &Rule{
		ID:        id,
		Name:      name,
		Type:      ruleType,
		Enabled:   true,
		Metadata:  make(map[string]any),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// DataPoint 数据点
type DataPoint struct {
	PointID    string    `json:"point_id"`
	DeviceID   string    `json:"device_id"`
	StationID  string    `json:"station_id"`
	Value      float64   `json:"value"`
	Quality    int       `json:"quality"`
	Timestamp  time.Time `json:"timestamp"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// DetectionResult 检测结果
type DetectionResult struct {
	Triggered    bool             `json:"triggered"`
	Rule         *Rule            `json:"rule"`
	DataPoint    *DataPoint       `json:"data_point"`
	Value        float64          `json:"value"`
	Threshold    float64          `json:"threshold"`
	Message      string           `json:"message"`
	TriggeredAt  time.Time        `json:"triggered_at"`
	Alarm        *entity.Alarm    `json:"alarm,omitempty"`
}

// DetectionHandler 检测处理器
type DetectionHandler func(ctx context.Context, result *DetectionResult) error

// SlidingWindow 滑动窗口
type SlidingWindow struct {
	mu        sync.RWMutex
	pointID   string
	duration  time.Duration
	maxCount  int
	data      []*windowEntry
	startTime time.Time
}

type windowEntry struct {
	value     float64
	timestamp time.Time
}

// NewSlidingWindow 创建滑动窗口
func NewSlidingWindow(pointID string, duration time.Duration, maxCount int) *SlidingWindow {
	return &SlidingWindow{
		pointID:   pointID,
		duration:  duration,
		maxCount:  maxCount,
		data:      make([]*windowEntry, 0, maxCount),
		startTime: time.Now(),
	}
}

// Add 添加数据点
func (w *SlidingWindow) Add(value float64, timestamp time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 添加新数据
	w.data = append(w.data, &windowEntry{
		value:     value,
		timestamp: timestamp,
	})

	// 清理过期数据
	cutoff := timestamp.Add(-w.duration)
	newData := make([]*windowEntry, 0, len(w.data))
	for _, entry := range w.data {
		if entry.timestamp.After(cutoff) {
			newData = append(newData, entry)
		}
	}
	w.data = newData

	// 限制数量
	if len(w.data) > w.maxCount {
		w.data = w.data[len(w.data)-w.maxCount:]
	}
}

// GetValues 获取窗口内的值
func (w *SlidingWindow) GetValues() []float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	values := make([]float64, len(w.data))
	for i, entry := range w.data {
		values[i] = entry.value
	}
	return values
}

// GetStats 获取窗口统计信息
func (w *SlidingWindow) GetStats() WindowStats {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if len(w.data) == 0 {
		return WindowStats{}
	}

	var sum, min, max float64
	min = w.data[0].value
	max = w.data[0].value

	for _, entry := range w.data {
		sum += entry.value
		if entry.value < min {
			min = entry.value
		}
		if entry.value > max {
			max = entry.value
		}
	}

	avg := sum / float64(len(w.data))
	
	// 计算标准差
	var variance float64
	for _, entry := range w.data {
		variance += math.Pow(entry.value-avg, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(w.data)))

	return WindowStats{
		Count:   len(w.data),
		Sum:     sum,
		Avg:     avg,
		Min:     min,
		Max:     max,
		StdDev:  stdDev,
		FirstAt: w.data[0].timestamp,
		LastAt:  w.data[len(w.data)-1].timestamp,
	}
}

// WindowStats 窗口统计信息
type WindowStats struct {
	Count   int       `json:"count"`
	Sum     float64   `json:"sum"`
	Avg     float64   `json:"avg"`
	Min     float64   `json:"min"`
	Max     float64   `json:"max"`
	StdDev  float64   `json:"std_dev"`
	FirstAt time.Time `json:"first_at"`
	LastAt  time.Time `json:"last_at"`
}

// DetectorConfig 检测器配置
type DetectorConfig struct {
	WorkerCount      int
	BufferSize       int
	WindowDuration   time.Duration
	MaxWindowsPerPoint int
	EnableDedup      bool
	EnableAggregator bool
}

// DefaultDetectorConfig 默认检测器配置
func DefaultDetectorConfig() DetectorConfig {
	return DetectorConfig{
		WorkerCount:        8,
		BufferSize:         10000,
		WindowDuration:     5 * time.Minute,
		MaxWindowsPerPoint: 100,
		EnableDedup:        true,
		EnableAggregator:   true,
	}
}

// Detector 检测器
type Detector struct {
	mu          sync.RWMutex
	config      DetectorConfig
	rules       map[string]*Rule
	windows     map[string]*SlidingWindow
	handlers    []DetectionHandler
	
	// 组件
	stateMachine *state.StateMachine
	deduplicator *dedup.Deduplicator
	aggregator   *aggregator.Aggregator
	
	// 通道
	dataChan    chan *DataPoint
	resultChan  chan *DetectionResult
	stopChan    chan struct{}
	wg          sync.WaitGroup
	
	// 统计
	stats       DetectorStats
}

// DetectorStats 检测器统计
type DetectorStats struct {
	mu              sync.RWMutex
	TotalProcessed  int64 `json:"total_processed"`
	TotalTriggered  int64 `json:"total_triggered"`
	TotalSuppressed int64 `json:"total_suppressed"`
	RuleMatches     map[string]int64 `json:"rule_matches"`
}

// NewDetector 创建检测器
func NewDetector(config DetectorConfig) *Detector {
	if config.WorkerCount <= 0 {
		config.WorkerCount = 8
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 10000
	}

	return &Detector{
		config:    config,
		rules:     make(map[string]*Rule),
		windows:   make(map[string]*SlidingWindow),
		dataChan:  make(chan *DataPoint, config.BufferSize),
		resultChan: make(chan *DetectionResult, config.BufferSize),
		stopChan:  make(chan struct{}),
		stats: DetectorStats{
			RuleMatches: make(map[string]int64),
		},
	}
}

// SetStateMachine 设置状态机
func (d *Detector) SetStateMachine(sm *state.StateMachine) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stateMachine = sm
}

// SetDeduplicator 设置去重器
func (d *Detector) SetDeduplicator(dedup *dedup.Deduplicator) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.deduplicator = dedup
}

// SetAggregator 设置聚合器
func (d *Detector) SetAggregator(agg *aggregator.Aggregator) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.aggregator = agg
}

// AddRule 添加规则
func (d *Detector) AddRule(rule *Rule) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.rules[rule.ID] = rule
}

// RemoveRule 移除规则
func (d *Detector) RemoveRule(ruleID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.rules, ruleID)
}

// GetRule 获取规则
func (d *Detector) GetRule(ruleID string) *Rule {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.rules[ruleID]
}

// GetAllRules 获取所有规则
func (d *Detector) GetAllRules() []*Rule {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rules := make([]*Rule, 0, len(d.rules))
	for _, rule := range d.rules {
		rules = append(rules, rule)
	}
	return rules
}

// AddHandler 添加检测处理器
func (d *Detector) AddHandler(handler DetectionHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers = append(d.handlers, handler)
}

// Detect 检测数据点
func (d *Detector) Detect(ctx context.Context, point *DataPoint) error {
	select {
	case d.dataChan <- point:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return errors.New("detector buffer full")
	}
}

// DetectBatch 批量检测
func (d *Detector) DetectBatch(ctx context.Context, points []*DataPoint) error {
	for _, point := range points {
		if err := d.Detect(ctx, point); err != nil {
			return err
		}
	}
	return nil
}

// Start 启动检测器
func (d *Detector) Start(ctx context.Context) {
	// 启动数据处理器
	for i := 0; i < d.config.WorkerCount; i++ {
		d.wg.Add(1)
		go d.dataWorker(ctx, i)
	}

	// 启动结果处理器
	d.wg.Add(1)
	go d.resultWorker(ctx)
}

// Stop 停止检测器
func (d *Detector) Stop() {
	close(d.stopChan)
	d.wg.Wait()
	close(d.dataChan)
	close(d.resultChan)
}

// dataWorker 数据处理工作协程
func (d *Detector) dataWorker(ctx context.Context, workerID int) {
	defer d.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopChan:
			return
		case point, ok := <-d.dataChan:
			if !ok {
				return
			}
			d.processDataPoint(ctx, point)
		}
	}
}

// resultWorker 结果处理工作协程
func (d *Detector) resultWorker(ctx context.Context) {
	defer d.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopChan:
			return
		case result, ok := <-d.resultChan:
			if !ok {
				return
			}
			d.handleResult(ctx, result)
		}
	}
}

// processDataPoint 处理数据点
func (d *Detector) processDataPoint(ctx context.Context, point *DataPoint) {
	d.stats.mu.Lock()
	d.stats.TotalProcessed++
	d.stats.mu.Unlock()

	// 更新滑动窗口
	d.updateWindow(point)

	// 获取匹配的规则
	rules := d.getMatchingRules(point)

	// 对每个规则进行检测
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		result := d.detectRule(ctx, rule, point)
		if result != nil && result.Triggered {
			select {
			case d.resultChan <- result:
			default:
				// 缓冲区满，丢弃结果
			}
		}
	}
}

// updateWindow 更新滑动窗口
func (d *Detector) updateWindow(point *DataPoint) {
	d.mu.Lock()
	defer d.mu.Unlock()

	window, exists := d.windows[point.PointID]
	if !exists {
		window = NewSlidingWindow(
			point.PointID,
			d.config.WindowDuration,
			d.config.MaxWindowsPerPoint,
		)
		d.windows[point.PointID] = window
	}
	window.Add(point.Value, point.Timestamp)
}

// getMatchingRules 获取匹配的规则
func (d *Detector) getMatchingRules(point *DataPoint) []*Rule {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var matching []*Rule
	for _, rule := range d.rules {
		if !rule.Enabled {
			continue
		}

		// 检查测点匹配
		if len(rule.PointIDs) > 0 {
			matched := false
			for _, pid := range rule.PointIDs {
				if pid == point.PointID {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// 检查设备匹配
		if len(rule.DeviceIDs) > 0 {
			matched := false
			for _, did := range rule.DeviceIDs {
				if did == point.DeviceID {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// 检查站点匹配
		if len(rule.StationIDs) > 0 {
			matched := false
			for _, sid := range rule.StationIDs {
				if sid == point.StationID {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		matching = append(matching, rule)
	}

	return matching
}

// detectRule 检测规则
func (d *Detector) detectRule(ctx context.Context, rule *Rule, point *DataPoint) *DetectionResult {
	result := &DetectionResult{
		Rule:        rule,
		DataPoint:   point,
		Value:       point.Value,
		TriggeredAt: time.Now(),
	}

	var triggered bool
	var threshold float64

	switch rule.Type {
	case RuleTypeThreshold:
		triggered, threshold = d.detectThreshold(rule, point)
	case RuleTypeRange:
		triggered, threshold = d.detectRange(rule, point)
	case RuleTypeRate:
		triggered, threshold = d.detectRate(rule, point)
	case RuleTypeDeviation:
		triggered, threshold = d.detectDeviation(rule, point)
	case RuleTypeDuration:
		triggered, threshold = d.detectDuration(rule, point)
	default:
		return nil
	}

	result.Triggered = triggered
	result.Threshold = threshold

	if triggered {
		result.Message = d.generateMessage(rule, point, threshold)
		result.Alarm = d.createAlarm(rule, point, threshold)
	}

	return result
}

// detectThreshold 阈值检测
func (d *Detector) detectThreshold(rule *Rule, point *DataPoint) (bool, float64) {
	switch rule.Operator {
	case OpEqual:
		return point.Value == rule.Threshold, rule.Threshold
	case OpNotEqual:
		return point.Value != rule.Threshold, rule.Threshold
	case OpGreaterThan:
		return point.Value > rule.Threshold, rule.Threshold
	case OpGreaterEqual:
		return point.Value >= rule.Threshold, rule.Threshold
	case OpLessThan:
		return point.Value < rule.Threshold, rule.Threshold
	case OpLessEqual:
		return point.Value <= rule.Threshold, rule.Threshold
	default:
		return false, rule.Threshold
	}
}

// detectRange 范围检测
func (d *Detector) detectRange(rule *Rule, point *DataPoint) (bool, float64) {
	// 超出范围触发告警
	if point.Value < rule.MinValue || point.Value > rule.MaxValue {
		if point.Value < rule.MinValue {
			return true, rule.MinValue
		}
		return true, rule.MaxValue
	}
	return false, 0
}

// detectRate 变化率检测
func (d *Detector) detectRate(rule *Rule, point *DataPoint) (bool, float64) {
	d.mu.RLock()
	window, exists := d.windows[point.PointID]
	d.mu.RUnlock()

	if !exists {
		return false, 0
	}

	stats := window.GetStats()
	if stats.Count < 2 {
		return false, 0
	}

	// 计算变化率
	duration := stats.LastAt.Sub(stats.FirstAt).Seconds()
	if duration <= 0 {
		return false, 0
	}

	rate := math.Abs(stats.Max-stats.Min) / duration
	return rate > rule.RateThreshold, rule.RateThreshold
}

// detectDeviation 偏差检测
func (d *Detector) detectDeviation(rule *Rule, point *DataPoint) (bool, float64) {
	d.mu.RLock()
	window, exists := d.windows[point.PointID]
	d.mu.RUnlock()

	if !exists {
		return false, 0
	}

	stats := window.GetStats()
	if stats.Count < 2 {
		return false, 0
	}

	// 计算偏差
	deviation := math.Abs(point.Value - stats.Avg)
	return deviation > rule.DeviationThreshold, rule.DeviationThreshold
}

// detectDuration 持续时间检测
func (d *Detector) detectDuration(rule *Rule, point *DataPoint) (bool, float64) {
	d.mu.RLock()
	window, exists := d.windows[point.PointID]
	d.mu.RUnlock()

	if !exists {
		return false, 0
	}

	stats := window.GetStats()
	if stats.Count < 2 {
		return false, 0
	}

	// 检查持续时间
	duration := stats.LastAt.Sub(stats.FirstAt)
	return duration >= rule.DurationThreshold, float64(rule.DurationThreshold.Seconds())
}

// generateMessage 生成消息
func (d *Detector) generateMessage(rule *Rule, point *DataPoint, threshold float64) string {
	if rule.Message != "" {
		return rule.Message
	}

	return fmt.Sprintf("告警规则[%s]触发: 测点%s当前值%.2f, 阈值%.2f",
		rule.Name, point.PointID, point.Value, threshold)
}

// createAlarm 创建告警
func (d *Detector) createAlarm(rule *Rule, point *DataPoint, threshold float64) *entity.Alarm {
	alarm := entity.NewAlarm(
		point.PointID,
		point.DeviceID,
		point.StationID,
		rule.AlarmType,
		rule.Level,
		rule.Title,
		d.generateMessage(rule, point, threshold),
	)
	alarm.Value = point.Value
	alarm.Threshold = threshold
	return alarm
}

// handleResult 处理结果
func (d *Detector) handleResult(ctx context.Context, result *DetectionResult) {
	if !result.Triggered {
		return
	}

	// 更新统计
	d.stats.mu.Lock()
	d.stats.TotalTriggered++
	d.stats.RuleMatches[result.Rule.ID]++
	d.stats.mu.Unlock()

	// 去重检查
	if d.config.EnableDedup && d.deduplicator != nil {
		if d.deduplicator.IsDuplicate(ctx, result.Alarm) {
			d.stats.mu.Lock()
			d.stats.TotalSuppressed++
			d.stats.mu.Unlock()
			return
		}
	}

	// 聚合处理
	if d.config.EnableAggregator && d.aggregator != nil {
		d.aggregator.Aggregate(ctx, result.Alarm)
	}

	// 调用处理器
	d.mu.RLock()
	handlers := make([]DetectionHandler, len(d.handlers))
	copy(handlers, d.handlers)
	d.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx, result); err != nil {
			// 记录错误但继续处理
			fmt.Printf("detection handler error: %v\n", err)
		}
	}
}

// GetStats 获取统计信息
func (d *Detector) GetStats() DetectorStats {
	d.stats.mu.RLock()
	defer d.stats.mu.RUnlock()

	stats := DetectorStats{
		TotalProcessed:  d.stats.TotalProcessed,
		TotalTriggered:  d.stats.TotalTriggered,
		TotalSuppressed: d.stats.TotalSuppressed,
		RuleMatches:     make(map[string]int64),
	}

	for k, v := range d.stats.RuleMatches {
		stats.RuleMatches[k] = v
	}

	return stats
}

// GetWindowStats 获取窗口统计
func (d *Detector) GetWindowStats(pointID string) *WindowStats {
	d.mu.RLock()
	defer d.mu.RUnlock()

	window, exists := d.windows[pointID]
	if !exists {
		return nil
	}

	stats := window.GetStats()
	return &stats
}

// ClearWindows 清空窗口
func (d *Detector) ClearWindows() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.windows = make(map[string]*SlidingWindow)
}

// GetRuleCount 获取规则数量
func (d *Detector) GetRuleCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.rules)
}

// GetWindowCount 获取窗口数量
func (d *Detector) GetWindowCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.windows)
}

// RealtimeDetector 实时检测器（简化版）
type RealtimeDetector struct {
	detector *Detector
}

// NewRealtimeDetector 创建实时检测器
func NewRealtimeDetector(config DetectorConfig) *RealtimeDetector {
	return &RealtimeDetector{
		detector: NewDetector(config),
	}
}

// Start 启动
func (r *RealtimeDetector) Start(ctx context.Context) {
	r.detector.Start(ctx)
}

// Stop 停止
func (r *RealtimeDetector) Stop() {
	r.detector.Stop()
}

// AddRule 添加规则
func (r *RealtimeDetector) AddRule(rule *Rule) {
	r.detector.AddRule(rule)
}

// RemoveRule 移除规则
func (r *RealtimeDetector) RemoveRule(ruleID string) {
	r.detector.RemoveRule(ruleID)
}

// Detect 检测
func (r *RealtimeDetector) Detect(ctx context.Context, point *DataPoint) error {
	return r.detector.Detect(ctx, point)
}

// DetectBatch 批量检测
func (r *RealtimeDetector) DetectBatch(ctx context.Context, points []*DataPoint) error {
	return r.detector.DetectBatch(ctx, points)
}

// AddHandler 添加处理器
func (r *RealtimeDetector) AddHandler(handler DetectionHandler) {
	r.detector.AddHandler(handler)
}

// GetStats 获取统计
func (r *RealtimeDetector) GetStats() DetectorStats {
	return r.detector.GetStats()
}

// GetDetector 获取底层检测器
func (r *RealtimeDetector) GetDetector() *Detector {
	return r.detector
}
