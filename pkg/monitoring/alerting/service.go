package alerting

import (
	"context"
	"fmt"
	"time"
)

// AlertService 告警服务
type AlertService struct {
	ruleManager *RuleManager
	aggregator  *AlertAggregator
	notifier    *AlertNotifier
	api         *AlertAPI

	ctx    context.Context
	cancel context.CancelFunc
}

// AlertServiceConfig 告警服务配置
type AlertServiceConfig struct {
	// 规则配置
	MetricProvider MetricProvider `json:"-"`

	// 聚合器配置
	AggregatorConfig AggregatorConfig `json:"aggregator_config"`

	// 通知配置
	NotificationStore NotificationStore `json:"-"`

	// 告警存储
	AlertStore AlertStore `json:"-"`
}

// NewAlertService 创建告警服务
func NewAlertService(config AlertServiceConfig) *AlertService {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建规则管理器
	ruleManager := NewRuleManager(config.MetricProvider)

	// 创建聚合器
	aggregator := NewAlertAggregator(config.AggregatorConfig)

	// 创建通知器
	notifier := NewAlertNotifier(config.NotificationStore)

	// 创建API
	api := NewAlertAPI(ruleManager, aggregator, notifier, config.AlertStore)

	return &AlertService{
		ruleManager: ruleManager,
		aggregator:  aggregator,
		notifier:    notifier,
		api:         api,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动告警服务
func (as *AlertService) Start() error {
	// 启动聚合器
	as.aggregator.Start(as.ctx)

	// 启动规则评估循环
	go as.runEvaluationLoop()

	// 启动静默清理循环
	go as.runSilenceCleanupLoop()

	// 启动去重清理循环
	go as.runDedupCleanupLoop()

	return nil
}

// Stop 停止告警服务
func (as *AlertService) Stop() {
	as.cancel()
	as.aggregator.Stop()
}

// runEvaluationLoop 运行规则评估循环
func (as *AlertService) runEvaluationLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-as.ctx.Done():
			return
		case <-ticker.C:
			as.evaluateRules()
		}
	}
}

// evaluateRules 评估规则
func (as *AlertService) evaluateRules() {
	results, err := as.ruleManager.EvaluateAll(as.ctx)
	if err != nil {
		fmt.Printf("评估规则失败: %v\n", err)
		return
	}

	for _, result := range results {
		if result.Triggered {
			as.handleTriggeredRule(result)
		}
	}
}

// handleTriggeredRule 处理触发的规则
func (as *AlertService) handleTriggeredRule(result *RuleEvaluationResult) {
	// 获取规则
	rule, exists := as.ruleManager.GetRule(result.RuleID)
	if !exists {
		return
	}

	// 创建告警实例
	alert := &AlertInstance{
		ID:          generateAlertID(),
		RuleID:      result.RuleID,
		RuleName:    result.RuleName,
		Category:    rule.Category,
		Severity:    rule.Severity,
		Title:       rule.Name,
		Message:     fmt.Sprintf("%s: 当前值=%.2f, 阈值=%.2f", rule.Description, result.Value, result.Threshold),
		Value:       result.Value,
		Threshold:   result.Threshold,
		TriggeredAt: time.Now(),
		Labels:      result.Labels,
	}

	// 聚合告警
	group, isNew, err := as.aggregator.Aggregate(as.ctx, alert)
	if err != nil {
		fmt.Printf("聚合告警失败: %v\n", err)
		return
	}

	// 如果是新分组，发送通知
	if isNew && len(rule.NotifyChannels) > 0 {
		// 获取默认接收者
		recipients := as.getDefaultRecipients(rule)
		as.notifier.NotifyAlert(as.ctx, alert, rule.NotifyChannels, recipients)
	}

	_ = group
}

// getDefaultRecipients 获取默认接收者
func (as *AlertService) getDefaultRecipients(rule *AlertRule) []Recipient {
	// 这里应该从配置或数据库中获取接收者
	// 为了示例，返回默认接收者
	return []Recipient{
		{Email: "admin@example.com", Name: "管理员"},
	}
}

// runSilenceCleanupLoop 运行静默清理循环
func (as *AlertService) runSilenceCleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-as.ctx.Done():
			return
		case <-ticker.C:
			as.aggregator.silenceManager.CleanupExpired()
		}
	}
}

// runDedupCleanupLoop 运行去重清理循环
func (as *AlertService) runDedupCleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-as.ctx.Done():
			return
		case <-ticker.C:
			as.aggregator.deduplicator.Cleanup()
		}
	}
}

// AddRule 添加规则
func (as *AlertService) AddRule(rule *AlertRule) error {
	return as.ruleManager.AddRule(rule)
}

// RemoveRule 移除规则
func (as *AlertService) RemoveRule(ruleID string) {
	as.ruleManager.RemoveRule(ruleID)
}

// GetRule 获取规则
func (as *AlertService) GetRule(ruleID string) (*AlertRule, bool) {
	return as.ruleManager.GetRule(ruleID)
}

// GetAllRules 获取所有规则
func (as *AlertService) GetAllRules() []*AlertRule {
	return as.ruleManager.GetAllRules()
}

// CreateAlert 创建告警
func (as *AlertService) CreateAlert(alert *AlertInstance) error {
	// 聚合告警
	_, _, err := as.aggregator.Aggregate(as.ctx, alert)
	return err
}

// AddSilence 添加静默规则
func (as *AlertService) AddSilence(silence *Silence) error {
	return as.aggregator.silenceManager.AddSilence(silence)
}

// RemoveSilence 移除静默规则
func (as *AlertService) RemoveSilence(silenceID string) {
	as.aggregator.silenceManager.RemoveSilence(silenceID)
}

// AddSuppressionRule 添加抑制规则
func (as *AlertService) AddSuppressionRule(rule *SuppressionRule) error {
	return as.aggregator.suppressionEngine.AddRule(rule)
}

// RemoveSuppressionRule 移除抑制规则
func (as *AlertService) RemoveSuppressionRule(ruleID string) {
	as.aggregator.suppressionEngine.RemoveRule(ruleID)
}

// RegisterNotificationChannel 注册通知渠道
func (as *AlertService) RegisterNotificationChannel(channel NotificationChannel) {
	as.notifier.RegisterChannel(channel)
}

// UnregisterNotificationChannel 注销通知渠道
func (as *AlertService) UnregisterNotificationChannel(name string) {
	as.notifier.UnregisterChannel(name)
}

// AddNotificationTemplate 添加通知模板
func (as *AlertService) AddNotificationTemplate(template *NotificationTemplate) error {
	return as.notifier.templateManager.AddTemplate(template)
}

// RemoveNotificationTemplate 移除通知模板
func (as *AlertService) RemoveNotificationTemplate(templateID string) {
	as.notifier.templateManager.RemoveTemplate(templateID)
}

// AddEscalationRule 添加升级规则
func (as *AlertService) AddEscalationRule(rule *EscalationRule) error {
	return as.notifier.escalationEngine.AddRule(rule)
}

// RemoveEscalationRule 移除升级规则
func (as *AlertService) RemoveEscalationRule(ruleID string) {
	as.notifier.escalationEngine.RemoveRule(ruleID)
}

// GetAPI 获取API
func (as *AlertService) GetAPI() *AlertAPI {
	return as.api
}

// GetAggregator 获取聚合器
func (as *AlertService) GetAggregator() *AlertAggregator {
	return as.aggregator
}

// GetNotifier 获取通知器
func (as *AlertService) GetNotifier() *AlertNotifier {
	return as.notifier
}

// GetRuleManager 获取规则管理器
func (as *AlertService) GetRuleManager() *RuleManager {
	return as.ruleManager
}

// GetStats 获取统计信息
func (as *AlertService) GetStats() *ServiceStats {
	aggStats := as.aggregator.GetStats()

	return &ServiceStats{
		TotalRules:      len(as.ruleManager.GetAllRules()),
		EnabledRules:    len(as.ruleManager.GetEnabledRules()),
		TotalGroups:     aggStats.TotalGroups,
		TotalAlerts:     aggStats.TotalAlerts,
		MaxGroupSize:    aggStats.MaxGroupSize,
		AvgGroupSize:    aggStats.AvgGroupSize,
		TotalSilences:   len(as.aggregator.silenceManager.ListSilences()),
		TotalSuppressions: len(as.aggregator.suppressionEngine.rules),
	}
}

// ServiceStats 服务统计
type ServiceStats struct {
	TotalRules        int     `json:"total_rules"`
	EnabledRules      int     `json:"enabled_rules"`
	TotalGroups       int     `json:"total_groups"`
	TotalAlerts       int     `json:"total_alerts"`
	MaxGroupSize      int     `json:"max_group_size"`
	AvgGroupSize      float64 `json:"avg_group_size"`
	TotalSilences     int     `json:"total_silences"`
	TotalSuppressions int     `json:"total_suppressions"`
}
