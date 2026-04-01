package aggregator

import (
	"context"
	"sync"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
)

// AggregationStrategy 聚合策略
type AggregationStrategy int

const (
	// StrategyNone 不聚合
	StrategyNone AggregationStrategy = iota
	// StrategyByDevice 按设备聚合
	StrategyByDevice
	// StrategyByStation 按站点聚合
	StrategyByStation
	// StrategyByType 按告警类型聚合
	StrategyByType
	// StrategyByLevel 按告警级别聚合
	StrategyByLevel
	// StrategyByTimeWindow 按时间窗口聚合
	StrategyByTimeWindow
	// StrategyByDeviceAndType 按设备和类型聚合
	StrategyByDeviceAndType
	// StrategyByStationAndLevel 按站点和级别聚合
	StrategyByStationAndLevel
)

func (s AggregationStrategy) String() string {
	switch s {
	case StrategyNone:
		return "none"
	case StrategyByDevice:
		return "by_device"
	case StrategyByStation:
		return "by_station"
	case StrategyByType:
		return "by_type"
	case StrategyByLevel:
		return "by_level"
	case StrategyByTimeWindow:
		return "by_time_window"
	case StrategyByDeviceAndType:
		return "by_device_and_type"
	case StrategyByStationAndLevel:
		return "by_station_and_level"
	default:
		return "unknown"
	}
}

// AggregationConfig 聚合配置
type AggregationConfig struct {
	// Strategy 聚合策略
	Strategy AggregationStrategy
	// WindowDuration 时间窗口时长
	WindowDuration time.Duration
	// MaxGroupSize 最大分组大小
	MaxGroupSize int
	// MinGroupSize 最小分组大小
	MinGroupSize int
	// FlushInterval 刷新间隔
	FlushInterval time.Duration
	// EnableAutoFlush 启用自动刷新
	EnableAutoFlush bool
}

// DefaultAggregationConfig 默认聚合配置
func DefaultAggregationConfig() AggregationConfig {
	return AggregationConfig{
		Strategy:        StrategyByDevice,
		WindowDuration:  5 * time.Minute,
		MaxGroupSize:    100,
		MinGroupSize:    1,
		FlushInterval:   30 * time.Second,
		EnableAutoFlush: true,
	}
}

// AggregatedAlarm 聚合告警
type AggregatedAlarm struct {
	// ID 聚合ID
	ID string `json:"id"`
	// Strategy 聚合策略
	Strategy AggregationStrategy `json:"strategy"`
	// GroupKey 分组键
	GroupKey string `json:"group_key"`
	// Alarms 原始告警列表
	Alarms []*entity.Alarm `json:"alarms"`
	// Count 告警数量
	Count int `json:"count"`
	// FirstTriggeredAt 首次触发时间
	FirstTriggeredAt time.Time `json:"first_triggered_at"`
	// LastTriggeredAt 最后触发时间
	LastTriggeredAt time.Time `json:"last_triggered_at"`
	// MaxLevel 最高级别
	MaxLevel entity.AlarmLevel `json:"max_level"`
	// Summary 摘要
	Summary string `json:"summary"`
	// Metadata 元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// NewAggregatedAlarm 创建聚合告警
func NewAggregatedAlarm(strategy AggregationStrategy, groupKey string) *AggregatedAlarm {
	now := time.Now()
	return &AggregatedAlarm{
		Strategy:         strategy,
		GroupKey:         groupKey,
		Alarms:           make([]*entity.Alarm, 0),
		Count:            0,
		FirstTriggeredAt: now,
		LastTriggeredAt:  now,
		MaxLevel:         entity.AlarmLevelInfo,
		Metadata:         make(map[string]interface{}),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// Add 添加告警
func (a *AggregatedAlarm) Add(alarm *entity.Alarm) {
	a.Alarms = append(a.Alarms, alarm)
	a.Count++
	a.UpdatedAt = time.Now()

	// 更新时间范围
	if alarm.TriggeredAt.Before(a.FirstTriggeredAt) {
		a.FirstTriggeredAt = alarm.TriggeredAt
	}
	if alarm.TriggeredAt.After(a.LastTriggeredAt) {
		a.LastTriggeredAt = alarm.TriggeredAt
	}

	// 更新最高级别
	if alarm.Level > a.MaxLevel {
		a.MaxLevel = alarm.Level
	}
}

// AggregationHandler 聚合处理器
type AggregationHandler func(ctx context.Context, aggregated *AggregatedAlarm) error

// Aggregator 聚合器
type Aggregator struct {
	mu       sync.RWMutex
	config   AggregationConfig
	groups   map[string]*AggregatedAlarm
	handlers []AggregationHandler

	// 时间窗口
	windowStart time.Time
	windowEnd   time.Time

	// 控制通道
	flushChan chan struct{}
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// NewAggregator 创建聚合器
func NewAggregator(config AggregationConfig) *Aggregator {
	if config.WindowDuration <= 0 {
		config.WindowDuration = 5 * time.Minute
	}
	if config.FlushInterval <= 0 {
		config.FlushInterval = 30 * time.Second
	}

	now := time.Now()
	return &Aggregator{
		config:      config,
		groups:      make(map[string]*AggregatedAlarm),
		windowStart: now,
		windowEnd:   now.Add(config.WindowDuration),
		flushChan:   make(chan struct{}, 1),
		stopChan:    make(chan struct{}),
	}
}

// AddHandler 添加聚合处理器
func (a *Aggregator) AddHandler(handler AggregationHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.handlers = append(a.handlers, handler)
}

// getGroupKey 获取分组键
func (a *Aggregator) getGroupKey(alarm *entity.Alarm) string {
	switch a.config.Strategy {
	case StrategyByDevice:
		return alarm.DeviceID
	case StrategyByStation:
		return alarm.StationID
	case StrategyByType:
		return string(alarm.Type)
	case StrategyByLevel:
		return string(rune(alarm.Level))
	case StrategyByTimeWindow:
		// 按时间窗口分组
		windowNum := alarm.TriggeredAt.Unix() / int64(a.config.WindowDuration.Seconds())
		return string(rune(windowNum))
	case StrategyByDeviceAndType:
		return alarm.DeviceID + ":" + string(alarm.Type)
	case StrategyByStationAndLevel:
		return alarm.StationID + ":" + string(rune(alarm.Level))
	default:
		return alarm.ID
	}
}

// Aggregate 聚合告警
func (a *Aggregator) Aggregate(ctx context.Context, alarm *entity.Alarm) (*AggregatedAlarm, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	groupKey := a.getGroupKey(alarm)

	// 检查是否需要刷新时间窗口
	if a.config.Strategy == StrategyByTimeWindow && time.Now().After(a.windowEnd) {
		a.flushLocked(ctx)
		a.resetWindow()
	}

	// 获取或创建分组
	group, exists := a.groups[groupKey]
	if !exists {
		group = NewAggregatedAlarm(a.config.Strategy, groupKey)
		a.groups[groupKey] = group
	}

	// 检查分组大小限制
	if a.config.MaxGroupSize > 0 && group.Count >= a.config.MaxGroupSize {
		// 达到最大大小，先刷新
		a.flushGroupLocked(ctx, groupKey)
		group = NewAggregatedAlarm(a.config.Strategy, groupKey)
		a.groups[groupKey] = group
	}

	// 添加告警
	group.Add(alarm)

	return group, !exists
}

// AggregateBatch 批量聚合
func (a *Aggregator) AggregateBatch(ctx context.Context, alarms []*entity.Alarm) []*AggregatedAlarm {
	results := make([]*AggregatedAlarm, 0, len(alarms))
	for _, alarm := range alarms {
		if agg, _ := a.Aggregate(ctx, alarm); agg != nil {
			results = append(results, agg)
		}
	}
	return results
}

// GetGroup 获取分组
func (a *Aggregator) GetGroup(groupKey string) *AggregatedAlarm {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.groups[groupKey]
}

// GetAllGroups 获取所有分组
func (a *Aggregator) GetAllGroups() []*AggregatedAlarm {
	a.mu.RLock()
	defer a.mu.RUnlock()

	groups := make([]*AggregatedAlarm, 0, len(a.groups))
	for _, group := range a.groups {
		groups = append(groups, group)
	}
	return groups
}

// GetGroupCount 获取分组数量
func (a *Aggregator) GetGroupCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.groups)
}

// GetTotalAlarmCount 获取总告警数
func (a *Aggregator) GetTotalAlarmCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	total := 0
	for _, group := range a.groups {
		total += group.Count
	}
	return total
}

// Flush 刷新所有分组
func (a *Aggregator) Flush(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.flushLocked(ctx)
}

// flushLocked 刷新所有分组（需要持有锁）
func (a *Aggregator) flushLocked(ctx context.Context) error {
	for groupKey := range a.groups {
		if err := a.flushGroupLocked(ctx, groupKey); err != nil {
			return err
		}
	}
	return nil
}

// flushGroupLocked 刷新单个分组（需要持有锁）
func (a *Aggregator) flushGroupLocked(ctx context.Context, groupKey string) error {
	group, exists := a.groups[groupKey]
	if !exists {
		return nil
	}

	// 检查最小分组大小
	if a.config.MinGroupSize > 0 && group.Count < a.config.MinGroupSize {
		return nil
	}

	// 生成摘要
	group.Summary = a.generateSummary(group)

	// 调用处理器
	for _, handler := range a.handlers {
		if err := handler(ctx, group); err != nil {
			// 记录错误但继续处理
			continue
		}
	}

	// 删除已刷新的分组
	delete(a.groups, groupKey)

	return nil
}

// generateSummary 生成摘要
func (a *Aggregator) generateSummary(group *AggregatedAlarm) string {
	switch a.config.Strategy {
	case StrategyByDevice:
		if len(group.Alarms) > 0 {
			return group.Alarms[0].Title
		}
	case StrategyByStation:
		return "Station aggregated alarm"
	case StrategyByType:
		return "Type aggregated alarm"
	case StrategyByLevel:
		return "Level aggregated alarm"
	case StrategyByTimeWindow:
		return "Time window aggregated alarm"
	}
	return "Aggregated alarm"
}

// resetWindow 重置时间窗口
func (a *Aggregator) resetWindow() {
	now := time.Now()
	a.windowStart = now
	a.windowEnd = now.Add(a.config.WindowDuration)
}

// Start 启动聚合器
func (a *Aggregator) Start(ctx context.Context) {
	if !a.config.EnableAutoFlush {
		return
	}

	a.wg.Add(1)
	go a.runFlushLoop(ctx)
}

// Stop 停止聚合器
func (a *Aggregator) Stop() {
	close(a.stopChan)
	a.wg.Wait()
}

// runFlushLoop 刷新循环
func (a *Aggregator) runFlushLoop(ctx context.Context) {
	defer a.wg.Done()

	ticker := time.NewTicker(a.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// 退出前刷新
			_ = a.Flush(context.Background())
			return
		case <-a.stopChan:
			// 退出前刷新
			_ = a.Flush(context.Background())
			return
		case <-ticker.C:
			_ = a.Flush(ctx)
		case <-a.flushChan:
			_ = a.Flush(ctx)
		}
	}
}

// TriggerFlush 触发刷新
func (a *Aggregator) TriggerFlush() {
	select {
	case a.flushChan <- struct{}{}:
	default:
	}
}

// SetStrategy 设置聚合策略
func (a *Aggregator) SetStrategy(strategy AggregationStrategy) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.config.Strategy = strategy
}

// SetWindowDuration 设置时间窗口时长
func (a *Aggregator) SetWindowDuration(duration time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.config.WindowDuration = duration
}

// GetConfig 获取配置
func (a *Aggregator) GetConfig() AggregationConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config
}

// AggregationStats 聚合统计
type AggregationStats struct {
	TotalGroups      int   `json:"total_groups"`
	TotalAlarms      int   `json:"total_alarms"`
	MaxGroupSize     int   `json:"max_group_size"`
	MinGroupSize     int   `json:"min_group_size"`
	AvgGroupSize     float64 `json:"avg_group_size"`
	WindowStartTime  time.Time `json:"window_start_time"`
	WindowEndTime    time.Time `json:"window_end_time"`
}

// GetStats 获取统计信息
func (a *Aggregator) GetStats() AggregationStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	stats := AggregationStats{
		TotalGroups:     len(a.groups),
		WindowStartTime: a.windowStart,
		WindowEndTime:   a.windowEnd,
	}

	if len(a.groups) == 0 {
		return stats
	}

	totalAlarms := 0
	maxSize := 0
	minSize := int(^uint(0) >> 1) // Max int

	for _, group := range a.groups {
		totalAlarms += group.Count
		if group.Count > maxSize {
			maxSize = group.Count
		}
		if group.Count < minSize {
			minSize = group.Count
		}
	}

	stats.TotalAlarms = totalAlarms
	stats.MaxGroupSize = maxSize
	stats.MinGroupSize = minSize
	if len(a.groups) > 0 {
		stats.AvgGroupSize = float64(totalAlarms) / float64(len(a.groups))
	}

	return stats
}

// Clear 清空所有分组
func (a *Aggregator) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.groups = make(map[string]*AggregatedAlarm)
	a.resetWindow()
}

// RemoveGroup 移除分组
func (a *Aggregator) RemoveGroup(groupKey string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.groups, groupKey)
}

// MultiStrategyAggregator 多策略聚合器
type MultiStrategyAggregator struct {
	mu         sync.RWMutex
	aggregators map[AggregationStrategy]*Aggregator
}

// NewMultiStrategyAggregator 创建多策略聚合器
func NewMultiStrategyAggregator(configs map[AggregationStrategy]AggregationConfig) *MultiStrategyAggregator {
	aggregators := make(map[AggregationStrategy]*Aggregator)
	for strategy, config := range configs {
		config.Strategy = strategy
		aggregators[strategy] = NewAggregator(config)
	}
	return &MultiStrategyAggregator{
		aggregators: aggregators,
	}
}

// Aggregate 按指定策略聚合
func (m *MultiStrategyAggregator) Aggregate(ctx context.Context, strategy AggregationStrategy, alarm *entity.Alarm) (*AggregatedAlarm, bool) {
	m.mu.RLock()
	agg, exists := m.aggregators[strategy]
	m.mu.RUnlock()

	if !exists {
		return nil, false
	}
	return agg.Aggregate(ctx, alarm)
}

// AggregateAll 按所有策略聚合
func (m *MultiStrategyAggregator) AggregateAll(ctx context.Context, alarm *entity.Alarm) map[AggregationStrategy]*AggregatedAlarm {
	results := make(map[AggregationStrategy]*AggregatedAlarm)

	m.mu.RLock()
	defer m.mu.RUnlock()

	for strategy, agg := range m.aggregators {
		if result, _ := agg.Aggregate(ctx, alarm); result != nil {
			results[strategy] = result
		}
	}

	return results
}

// GetAggregator 获取指定策略的聚合器
func (m *MultiStrategyAggregator) GetAggregator(strategy AggregationStrategy) *Aggregator {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.aggregators[strategy]
}

// Start 启动所有聚合器
func (m *MultiStrategyAggregator) Start(ctx context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, agg := range m.aggregators {
		agg.Start(ctx)
	}
}

// Stop 停止所有聚合器
func (m *MultiStrategyAggregator) Stop() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, agg := range m.aggregators {
		agg.Stop()
	}
}

// FlushAll 刷新所有聚合器
func (m *MultiStrategyAggregator) FlushAll(ctx context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, agg := range m.aggregators {
		_ = agg.Flush(ctx)
	}
}
