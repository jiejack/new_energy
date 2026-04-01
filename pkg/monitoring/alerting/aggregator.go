package alerting

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// AlertAggregator 告警聚合器
type AlertAggregator struct {
	mu              sync.RWMutex
	config          AggregatorConfig
	groups          map[string]*AlertGroup
	silenceManager  *SilenceManager
	suppressionEngine *SuppressionEngine
	deduplicator    *AlertDeduplicator

	// 时间窗口
	windowStart time.Time
	windowEnd   time.Time

	// 控制通道
	flushChan chan struct{}
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// AggregatorConfig 聚合器配置
type AggregatorConfig struct {
	// 聚合策略
	Strategy AggregationStrategy `json:"strategy"`

	// 时间窗口配置
	WindowDuration time.Duration `json:"window_duration"`
	FlushInterval  time.Duration `json:"flush_interval"`

	// 分组配置
	MaxGroupSize int `json:"max_group_size"`
	MinGroupSize int `json:"min_group_size"`

	// 降噪配置
	EnableDeduplication bool `json:"enable_deduplication"`
	DedupWindow         time.Duration `json:"dedup_window"`

	// 自动刷新
	EnableAutoFlush bool `json:"enable_auto_flush"`
}

// AggregationStrategy 聚合策略
type AggregationStrategy string

const (
	StrategyNone          AggregationStrategy = "none"
	StrategyBySource      AggregationStrategy = "by_source"
	StrategyBySeverity    AggregationStrategy = "by_severity"
	StrategyByCategory    AggregationStrategy = "by_category"
	StrategyByTimeWindow  AggregationStrategy = "by_time_window"
	StrategyByLabels      AggregationStrategy = "by_labels"
	StrategyByRule        AggregationStrategy = "by_rule"
	StrategyComposite     AggregationStrategy = "composite"
)

// DefaultAggregatorConfig 默认聚合器配置
func DefaultAggregatorConfig() AggregatorConfig {
	return AggregatorConfig{
		Strategy:            StrategyBySource,
		WindowDuration:      5 * time.Minute,
		FlushInterval:       30 * time.Second,
		MaxGroupSize:        100,
		MinGroupSize:        1,
		EnableDeduplication: true,
		DedupWindow:         5 * time.Minute,
		EnableAutoFlush:     true,
	}
}

// AlertGroup 告警分组
type AlertGroup struct {
	ID             string            `json:"id"`
	Strategy       AggregationStrategy `json:"strategy"`
	GroupKey       string            `json:"group_key"`
	Alerts         []*AlertInstance  `json:"alerts"`
	Count          int               `json:"count"`
	FirstTriggered time.Time         `json:"first_triggered"`
	LastTriggered  time.Time         `json:"last_triggered"`
	MaxSeverity    AlertSeverity     `json:"max_severity"`
	Summary        string            `json:"summary"`
	Labels         map[string]string `json:"labels"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// AlertInstance 告警实例
type AlertInstance struct {
	ID           string            `json:"id"`
	RuleID       string            `json:"rule_id"`
	RuleName     string            `json:"rule_name"`
	Category     AlertCategory     `json:"category"`
	Severity     AlertSeverity     `json:"severity"`
	Title        string            `json:"title"`
	Message      string            `json:"message"`
	Value        float64           `json:"value"`
	Threshold    float64           `json:"threshold"`
	Labels       map[string]string `json:"labels"`
	TriggeredAt  time.Time         `json:"triggered_at"`
	Source       string            `json:"source"`
	Fingerprint  string            `json:"fingerprint"`
}

// NewAlertAggregator 创建告警聚合器
func NewAlertAggregator(config AggregatorConfig) *AlertAggregator {
	if config.WindowDuration <= 0 {
		config.WindowDuration = 5 * time.Minute
	}
	if config.FlushInterval <= 0 {
		config.FlushInterval = 30 * time.Second
	}

	now := time.Now()
	return &AlertAggregator{
		config:         config,
		groups:         make(map[string]*AlertGroup),
		silenceManager: NewSilenceManager(),
		suppressionEngine: NewSuppressionEngine(),
		deduplicator:   NewAlertDeduplicator(config.DedupWindow),
		windowStart:    now,
		windowEnd:      now.Add(config.WindowDuration),
		flushChan:      make(chan struct{}, 1),
		stopChan:       make(chan struct{}),
	}
}

// Aggregate 聚合告警
func (aa *AlertAggregator) Aggregate(ctx context.Context, alert *AlertInstance) (*AlertGroup, bool, error) {
	aa.mu.Lock()
	defer aa.mu.Unlock()

	// 检查是否需要刷新时间窗口
	if aa.config.Strategy == StrategyByTimeWindow && time.Now().After(aa.windowEnd) {
		if err := aa.flushLocked(ctx); err != nil {
			return nil, false, err
		}
		aa.resetWindow()
	}

	// 去重检查
	if aa.config.EnableDeduplication {
		if aa.deduplicator.IsDuplicate(alert) {
			return nil, false, fmt.Errorf("duplicate alert")
		}
		aa.deduplicator.Record(alert)
	}

	// 检查静默
	if aa.silenceManager.IsSilenced(alert) {
		return nil, false, fmt.Errorf("alert is silenced")
	}

	// 检查抑制
	if suppressed, reason := aa.suppressionEngine.IsSuppressed(alert); suppressed {
		return nil, false, fmt.Errorf("alert is suppressed: %s", reason)
	}

	// 计算分组键
	groupKey := aa.getGroupKey(alert)

	// 获取或创建分组
	group, exists := aa.groups[groupKey]
	if !exists {
		group = NewAlertGroup(aa.config.Strategy, groupKey)
		aa.groups[groupKey] = group
	}

	// 检查分组大小限制
	if aa.config.MaxGroupSize > 0 && group.Count >= aa.config.MaxGroupSize {
		// 达到最大大小，先刷新
		if err := aa.flushGroupLocked(ctx, groupKey); err != nil {
			return nil, false, err
		}
		group = NewAlertGroup(aa.config.Strategy, groupKey)
		aa.groups[groupKey] = group
	}

	// 添加告警
	group.Add(alert)

	return group, !exists, nil
}

// getGroupKey 获取分组键
func (aa *AlertAggregator) getGroupKey(alert *AlertInstance) string {
	switch aa.config.Strategy {
	case StrategyBySource:
		return alert.Source

	case StrategyBySeverity:
		return string(alert.Severity)

	case StrategyByCategory:
		return string(alert.Category)

	case StrategyByTimeWindow:
		windowNum := alert.TriggeredAt.Unix() / int64(aa.config.WindowDuration.Seconds())
		return fmt.Sprintf("window_%d", windowNum)

	case StrategyByLabels:
		// 使用标签生成键
		return generateLabelsKey(alert.Labels)

	case StrategyByRule:
		return alert.RuleID

	case StrategyComposite:
		// 组合多个维度
		return fmt.Sprintf("%s:%s:%s", alert.Source, alert.Category, alert.Severity)

	default:
		return alert.ID
	}
}

// Flush 刷新所有分组
func (aa *AlertAggregator) Flush(ctx context.Context) error {
	aa.mu.Lock()
	defer aa.mu.Unlock()
	return aa.flushLocked(ctx)
}

// flushLocked 刷新所有分组（需要持有锁）
func (aa *AlertAggregator) flushLocked(ctx context.Context) error {
	for groupKey := range aa.groups {
		if err := aa.flushGroupLocked(ctx, groupKey); err != nil {
			return err
		}
	}
	return nil
}

// flushGroupLocked 刷新单个分组（需要持有锁）
func (aa *AlertAggregator) flushGroupLocked(ctx context.Context, groupKey string) error {
	group, exists := aa.groups[groupKey]
	if !exists {
		return nil
	}

	// 检查最小分组大小
	if aa.config.MinGroupSize > 0 && group.Count < aa.config.MinGroupSize {
		return nil
	}

	// 生成摘要
	group.Summary = aa.generateSummary(group)

	// 这里可以添加通知逻辑
	// 例如：调用通知管理器发送聚合后的告警

	// 删除已刷新的分组
	delete(aa.groups, groupKey)

	return nil
}

// generateSummary 生成摘要
func (aa *AlertAggregator) generateSummary(group *AlertGroup) string {
	if len(group.Alerts) == 0 {
		return ""
	}

	if len(group.Alerts) == 1 {
		return group.Alerts[0].Title
	}

	return fmt.Sprintf("%d alerts aggregated from %s to %s",
		group.Count,
		group.FirstTriggered.Format("15:04:05"),
		group.LastTriggered.Format("15:04:05"))
}

// resetWindow 重置时间窗口
func (aa *AlertAggregator) resetWindow() {
	now := time.Now()
	aa.windowStart = now
	aa.windowEnd = now.Add(aa.config.WindowDuration)
}

// Start 启动聚合器
func (aa *AlertAggregator) Start(ctx context.Context) {
	if !aa.config.EnableAutoFlush {
		return
	}

	aa.wg.Add(1)
	go aa.runFlushLoop(ctx)
}

// Stop 停止聚合器
func (aa *AlertAggregator) Stop() {
	close(aa.stopChan)
	aa.wg.Wait()
}

// runFlushLoop 刷新循环
func (aa *AlertAggregator) runFlushLoop(ctx context.Context) {
	defer aa.wg.Done()

	ticker := time.NewTicker(aa.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			_ = aa.Flush(context.Background())
			return
		case <-aa.stopChan:
			_ = aa.Flush(context.Background())
			return
		case <-ticker.C:
			_ = aa.Flush(ctx)
		case <-aa.flushChan:
			_ = aa.Flush(ctx)
		}
	}
}

// TriggerFlush 触发刷新
func (aa *AlertAggregator) TriggerFlush() {
	select {
	case aa.flushChan <- struct{}{}:
	default:
	}
}

// GetGroup 获取分组
func (aa *AlertAggregator) GetGroup(groupKey string) *AlertGroup {
	aa.mu.RLock()
	defer aa.mu.RUnlock()
	return aa.groups[groupKey]
}

// GetAllGroups 获取所有分组
func (aa *AlertAggregator) GetAllGroups() []*AlertGroup {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	groups := make([]*AlertGroup, 0, len(aa.groups))
	for _, group := range aa.groups {
		groups = append(groups, group)
	}
	return groups
}

// GetStats 获取统计信息
func (aa *AlertAggregator) GetStats() AggregatorStats {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	stats := AggregatorStats{
		TotalGroups:     len(aa.groups),
		WindowStartTime: aa.windowStart,
		WindowEndTime:   aa.windowEnd,
	}

	if len(aa.groups) == 0 {
		return stats
	}

	totalAlerts := 0
	maxSize := 0
	minSize := int(^uint(0) >> 1)

	for _, group := range aa.groups {
		totalAlerts += group.Count
		if group.Count > maxSize {
			maxSize = group.Count
		}
		if group.Count < minSize {
			minSize = group.Count
		}
	}

	stats.TotalAlerts = totalAlerts
	stats.MaxGroupSize = maxSize
	stats.MinGroupSize = minSize
	if len(aa.groups) > 0 {
		stats.AvgGroupSize = float64(totalAlerts) / float64(len(aa.groups))
	}

	return stats
}

// AggregatorStats 聚合器统计
type AggregatorStats struct {
	TotalGroups     int       `json:"total_groups"`
	TotalAlerts     int       `json:"total_alerts"`
	MaxGroupSize    int       `json:"max_group_size"`
	MinGroupSize    int       `json:"min_group_size"`
	AvgGroupSize    float64   `json:"avg_group_size"`
	WindowStartTime time.Time `json:"window_start_time"`
	WindowEndTime   time.Time `json:"window_end_time"`
}

// NewAlertGroup 创建告警分组
func NewAlertGroup(strategy AggregationStrategy, groupKey string) *AlertGroup {
	now := time.Now()
	return &AlertGroup{
		Strategy:       strategy,
		GroupKey:       groupKey,
		Alerts:         make([]*AlertInstance, 0),
		Count:          0,
		FirstTriggered: now,
		LastTriggered:  now,
		MaxSeverity:    SeverityInfo,
		Labels:         make(map[string]string),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Add 添加告警
func (ag *AlertGroup) Add(alert *AlertInstance) {
	ag.Alerts = append(ag.Alerts, alert)
	ag.Count++
	ag.UpdatedAt = time.Now()

	// 更新时间范围
	if alert.TriggeredAt.Before(ag.FirstTriggered) {
		ag.FirstTriggered = alert.TriggeredAt
	}
	if alert.TriggeredAt.After(ag.LastTriggered) {
		ag.LastTriggered = alert.TriggeredAt
	}

	// 更新最高严重程度
	if getSeverityLevel(alert.Severity) > getSeverityLevel(ag.MaxSeverity) {
		ag.MaxSeverity = alert.Severity
	}

	// 合并标签
	for k, v := range alert.Labels {
		ag.Labels[k] = v
	}
}

// getSeverityLevel 获取严重程度级别
func getSeverityLevel(severity AlertSeverity) int {
	switch severity {
	case SeverityInfo:
		return 1
	case SeverityWarning:
		return 2
	case SeverityCritical:
		return 3
	case SeverityEmergency:
		return 4
	default:
		return 0
	}
}

// SilenceManager 静默管理器
type SilenceManager struct {
	mu       sync.RWMutex
	silences map[string]*Silence
}

// Silence 静默规则
type Silence struct {
	ID        string            `json:"id"`
	Matchers  map[string]string `json:"matchers"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
	Reason    string            `json:"reason"`
	CreatedBy string            `json:"created_by"`
	CreatedAt time.Time         `json:"created_at"`
}

// NewSilenceManager 创建静默管理器
func NewSilenceManager() *SilenceManager {
	return &SilenceManager{
		silences: make(map[string]*Silence),
	}
}

// IsSilenced 检查告警是否被静默
func (sm *SilenceManager) IsSilenced(alert *AlertInstance) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	now := time.Now()
	for _, silence := range sm.silences {
		// 检查时间范围
		if now.Before(silence.StartTime) || now.After(silence.EndTime) {
			continue
		}

		// 检查匹配器
		if sm.matchAlert(alert, silence.Matchers) {
			return true
		}
	}

	return false
}

// matchAlert 匹配告警
func (sm *SilenceManager) matchAlert(alert *AlertInstance, matchers map[string]string) bool {
	for key, value := range matchers {
		alertValue, exists := alert.Labels[key]
		if !exists || alertValue != value {
			return false
		}
	}
	return true
}

// AddSilence 添加静默规则
func (sm *SilenceManager) AddSilence(silence *Silence) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if silence.ID == "" {
		return fmt.Errorf("silence id is required")
	}

	sm.silences[silence.ID] = silence
	return nil
}

// RemoveSilence 移除静默规则
func (sm *SilenceManager) RemoveSilence(silenceID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.silences, silenceID)
}

// GetSilence 获取静默规则
func (sm *SilenceManager) GetSilence(silenceID string) (*Silence, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	silence, exists := sm.silences[silenceID]
	return silence, exists
}

// ListSilences 列出静默规则
func (sm *SilenceManager) ListSilences() []*Silence {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	silences := make([]*Silence, 0, len(sm.silences))
	for _, silence := range sm.silences {
		silences = append(silences, silence)
	}
	return silences
}

// CleanupExpired 清理过期的静默规则
func (sm *SilenceManager) CleanupExpired() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for id, silence := range sm.silences {
		if now.After(silence.EndTime) {
			delete(sm.silences, id)
		}
	}
}

// SuppressionEngine 抑制引擎
type SuppressionEngine struct {
	mu      sync.RWMutex
	rules   map[string]*SuppressionRule
}

// SuppressionRule 抑制规则
type SuppressionRule struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	SourceMatchers  map[string]string `json:"source_matchers"`
	TargetMatchers  map[string]string `json:"target_matchers"`
	Enabled         bool              `json:"enabled"`
	CreatedAt       time.Time         `json:"created_at"`
}

// NewSuppressionEngine 创建抑制引擎
func NewSuppressionEngine() *SuppressionEngine {
	return &SuppressionEngine{
		rules: make(map[string]*SuppressionRule),
	}
}

// IsSuppressed 检查告警是否被抑制
func (se *SuppressionEngine) IsSuppressed(alert *AlertInstance) (bool, string) {
	se.mu.RLock()
	defer se.mu.RUnlock()

	for _, rule := range se.rules {
		if !rule.Enabled {
			continue
		}

		// 检查是否匹配目标
		if se.matchAlert(alert, rule.TargetMatchers) {
			return true, fmt.Sprintf("suppressed by rule: %s", rule.Name)
		}
	}

	return false, ""
}

// matchAlert 匹配告警
func (se *SuppressionEngine) matchAlert(alert *AlertInstance, matchers map[string]string) bool {
	for key, value := range matchers {
		alertValue, exists := alert.Labels[key]
		if !exists || alertValue != value {
			return false
		}
	}
	return true
}

// AddRule 添加抑制规则
func (se *SuppressionEngine) AddRule(rule *SuppressionRule) error {
	se.mu.Lock()
	defer se.mu.Unlock()

	if rule.ID == "" {
		return fmt.Errorf("rule id is required")
	}

	se.rules[rule.ID] = rule
	return nil
}

// RemoveRule 移除抑制规则
func (se *SuppressionEngine) RemoveRule(ruleID string) {
	se.mu.Lock()
	defer se.mu.Unlock()
	delete(se.rules, ruleID)
}

// AlertDeduplicator 告警去重器
type AlertDeduplicator struct {
	mu      sync.RWMutex
	window  time.Duration
	seen    map[string]*DedupEntry
}

// DedupEntry 去重条目
type DedupEntry struct {
	Fingerprint string
	AlertID     string
	Timestamp   time.Time
}

// NewAlertDeduplicator 创建告警去重器
func NewAlertDeduplicator(window time.Duration) *AlertDeduplicator {
	return &AlertDeduplicator{
		window: window,
		seen:   make(map[string]*DedupEntry),
	}
}

// IsDuplicate 检查是否重复
func (ad *AlertDeduplicator) IsDuplicate(alert *AlertInstance) bool {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	fingerprint := alert.Fingerprint
	if fingerprint == "" {
		fingerprint = generateFingerprint(alert)
	}

	entry, exists := ad.seen[fingerprint]
	if !exists {
		return false
	}

	// 检查是否在时间窗口内
	return time.Since(entry.Timestamp) < ad.window
}

// Record 记录告警
func (ad *AlertDeduplicator) Record(alert *AlertInstance) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	fingerprint := alert.Fingerprint
	if fingerprint == "" {
		fingerprint = generateFingerprint(alert)
	}

	ad.seen[fingerprint] = &DedupEntry{
		Fingerprint: fingerprint,
		AlertID:     alert.ID,
		Timestamp:   time.Now(),
	}
}

// Cleanup 清理过期条目
func (ad *AlertDeduplicator) Cleanup() {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	now := time.Now()
	for key, entry := range ad.seen {
		if now.Sub(entry.Timestamp) > ad.window {
			delete(ad.seen, key)
		}
	}
}

// generateFingerprint 生成指纹
func generateFingerprint(alert *AlertInstance) string {
	data := fmt.Sprintf("%s:%s:%s:%v:%v",
		alert.RuleID,
		alert.Source,
		alert.Message,
		alert.Value,
		alert.Labels)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// generateLabelsKey 生成标签键
func generateLabelsKey(labels map[string]string) string {
	if len(labels) == 0 {
		return "no_labels"
	}

	data := ""
	for k, v := range labels {
		data += fmt.Sprintf("%s=%s;", k, v)
	}

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}
