package alerting

import (
	"context"
	"testing"
	"time"
)

// TestAlertRule 测试告警规则
func TestAlertRule(t *testing.T) {
	// 创建系统资源告警规则
	rules := NewSystemResourceAlertRules()

	cpuRule := rules.CPUUsageRule(80.0, 5*time.Minute)
	if cpuRule.ID != "system_cpu_usage" {
		t.Errorf("expected rule ID 'system_cpu_usage', got '%s'", cpuRule.ID)
	}

	if cpuRule.Condition.Threshold != 80.0 {
		t.Errorf("expected threshold 80.0, got %f", cpuRule.Condition.Threshold)
	}

	memoryRule := rules.MemoryUsageRule(85.0, 5*time.Minute)
	if memoryRule.ID != "system_memory_usage" {
		t.Errorf("expected rule ID 'system_memory_usage', got '%s'", memoryRule.ID)
	}

	diskRule := rules.DiskUsageRule(90.0, 5*time.Minute)
	if diskRule.ID != "system_disk_usage" {
		t.Errorf("expected rule ID 'system_disk_usage', got '%s'", diskRule.ID)
	}
}

// TestRuleManager 测试规则管理器
func TestRuleManager(t *testing.T) {
	// 创建模拟的指标提供者
	provider := &MockMetricProvider{}
	manager := NewRuleManager(provider)

	// 创建测试规则
	rule := &AlertRule{
		ID:          "test_rule_1",
		Name:        "测试规则",
		Description: "这是一个测试规则",
		Category:    CategorySystemResource,
		Severity:    SeverityWarning,
		Enabled:     true,
		Condition: AlertCondition{
			MetricName:  "test_metric",
			Operator:    OpGT,
			Threshold:   100,
			Duration:    1 * time.Minute,
		},
	}

	// 添加规则
	err := manager.AddRule(rule)
	if err != nil {
		t.Fatalf("failed to add rule: %v", err)
	}

	// 获取规则
	retrievedRule, exists := manager.GetRule("test_rule_1")
	if !exists {
		t.Fatal("rule not found")
	}

	if retrievedRule.Name != "测试规则" {
		t.Errorf("expected rule name '测试规则', got '%s'", retrievedRule.Name)
	}

	// 获取所有规则
	allRules := manager.GetAllRules()
	if len(allRules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(allRules))
	}

	// 移除规则
	manager.RemoveRule("test_rule_1")
	_, exists = manager.GetRule("test_rule_1")
	if exists {
		t.Fatal("rule should be removed")
	}
}

// TestAlertAggregator 测试告警聚合器
func TestAlertAggregator(t *testing.T) {
	config := DefaultAggregatorConfig()
	aggregator := NewAlertAggregator(config)

	// 创建测试告警
	alert1 := &AlertInstance{
		ID:          "alert_1",
		RuleID:      "rule_1",
		RuleName:    "测试规则1",
		Category:    CategorySystemResource,
		Severity:    SeverityWarning,
		Title:       "CPU使用率告警",
		Message:     "CPU使用率超过80%",
		Value:       85.5,
		Threshold:   80.0,
		TriggeredAt: time.Now(),
		Source:      "server-1",
	}

	alert2 := &AlertInstance{
		ID:          "alert_2",
		RuleID:      "rule_1",
		RuleName:    "测试规则1",
		Category:    CategorySystemResource,
		Severity:    SeverityCritical,
		Title:       "CPU使用率告警",
		Message:     "CPU使用率超过90%",
		Value:       92.0,
		Threshold:   90.0,
		TriggeredAt: time.Now(),
		Source:      "server-1",
	}

	// 聚合告警
	ctx := context.Background()
	group1, isNew1, err := aggregator.Aggregate(ctx, alert1)
	if err != nil {
		t.Fatalf("failed to aggregate alert1: %v", err)
	}

	if !isNew1 {
		t.Error("expected new group for alert1")
	}

	if group1.Count != 1 {
		t.Errorf("expected group count 1, got %d", group1.Count)
	}

	group2, isNew2, err := aggregator.Aggregate(ctx, alert2)
	if err != nil {
		t.Fatalf("failed to aggregate alert2: %v", err)
	}

	if isNew2 {
		t.Error("expected existing group for alert2")
	}

	if group2.Count != 2 {
		t.Errorf("expected group count 2, got %d", group2.Count)
	}

	// 检查统计信息
	stats := aggregator.GetStats()
	if stats.TotalGroups != 1 {
		t.Errorf("expected 1 total group, got %d", stats.TotalGroups)
	}

	if stats.TotalAlerts != 2 {
		t.Errorf("expected 2 total alerts, got %d", stats.TotalAlerts)
	}
}

// TestSilenceManager 测试静默管理器
func TestSilenceManager(t *testing.T) {
	manager := NewSilenceManager()

	// 创建静默规则
	silence := &Silence{
		ID:        "silence_1",
		Matchers:  map[string]string{"source": "server-1"},
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
		Reason:    "维护中",
		CreatedBy: "admin",
		CreatedAt: time.Now(),
	}

	// 添加静默规则
	err := manager.AddSilence(silence)
	if err != nil {
		t.Fatalf("failed to add silence: %v", err)
	}

	// 测试告警
	alert := &AlertInstance{
		ID:     "alert_1",
		Source: "server-1",
		Labels: map[string]string{"source": "server-1"},
	}

	// 检查是否被静默
	isSilenced := manager.IsSilenced(alert)
	if !isSilenced {
		t.Error("expected alert to be silenced")
	}

	// 移除静默规则
	manager.RemoveSilence("silence_1")

	// 再次检查
	isSilenced = manager.IsSilenced(alert)
	if isSilenced {
		t.Error("expected alert not to be silenced after removal")
	}
}

// TestSuppressionEngine 测试抑制引擎
func TestSuppressionEngine(t *testing.T) {
	engine := NewSuppressionEngine()

	// 创建抑制规则
	rule := &SuppressionRule{
		ID:             "suppression_1",
		Name:           "测试抑制规则",
		SourceMatchers: map[string]string{"severity": "critical"},
		TargetMatchers: map[string]string{"severity": "warning"},
		Enabled:        true,
		CreatedAt:      time.Now(),
	}

	// 添加抑制规则
	err := engine.AddRule(rule)
	if err != nil {
		t.Fatalf("failed to add suppression rule: %v", err)
	}

	// 测试告警
	alert := &AlertInstance{
		ID:       "alert_1",
		Severity: SeverityWarning,
		Labels:   map[string]string{"severity": "warning"},
	}

	// 检查是否被抑制
	isSuppressed, reason := engine.IsSuppressed(alert)
	if !isSuppressed {
		t.Error("expected alert to be suppressed")
	}

	if reason == "" {
		t.Error("expected suppression reason")
	}

	// 移除抑制规则
	engine.RemoveRule("suppression_1")

	// 再次检查
	isSuppressed, _ = engine.IsSuppressed(alert)
	if isSuppressed {
		t.Error("expected alert not to be suppressed after removal")
	}
}

// TestAlertDeduplicator 测试告警去重器
func TestAlertDeduplicator(t *testing.T) {
	deduplicator := NewAlertDeduplicator(5 * time.Minute)

	// 创建测试告警
	alert := &AlertInstance{
		ID:          "alert_1",
		RuleID:      "rule_1",
		RuleName:    "测试规则",
		Message:     "测试消息",
		Value:       100.0,
		TriggeredAt: time.Now(),
		Source:      "server-1",
	}

	// 第一次检查
	isDuplicate := deduplicator.IsDuplicate(alert)
	if isDuplicate {
		t.Error("expected alert not to be duplicate on first check")
	}

	// 记录告警
	deduplicator.Record(alert)

	// 第二次检查
	isDuplicate = deduplicator.IsDuplicate(alert)
	if !isDuplicate {
		t.Error("expected alert to be duplicate on second check")
	}
}

// TestNotificationTemplateManager 测试通知模板管理器
func TestNotificationTemplateManager(t *testing.T) {
	manager := NewNotificationTemplateManager()

	// 创建模板
	template := &NotificationTemplate{
		ID:              "template_1",
		Name:            "告警通知模板",
		Channel:         "email",
		SubjectTemplate: "告警: {{.alert.Title}}",
		ContentTemplate: "告警内容: {{.alert.Message}}",
	}

	// 添加模板
	err := manager.AddTemplate(template)
	if err != nil {
		t.Fatalf("failed to add template: %v", err)
	}

	// 获取模板
	retrievedTemplate, err := manager.GetTemplate("template_1")
	if err != nil {
		t.Fatalf("failed to get template: %v", err)
	}

	if retrievedTemplate.Name != "告警通知模板" {
		t.Errorf("expected template name '告警通知模板', got '%s'", retrievedTemplate.Name)
	}

	// 渲染模板
	alert := &AlertInstance{
		Title:   "CPU告警",
		Message: "CPU使用率超过80%",
	}

	data := map[string]interface{}{
		"alert": alert,
	}

	subject, err := retrievedTemplate.RenderSubject(data)
	if err != nil {
		t.Fatalf("failed to render subject: %v", err)
	}

	if subject != "告警: CPU告警" {
		t.Errorf("expected subject '告警: CPU告警', got '%s'", subject)
	}

	content, err := retrievedTemplate.RenderContent(data)
	if err != nil {
		t.Fatalf("failed to render content: %v", err)
	}

	if content != "告警内容: CPU使用率超过80%" {
		t.Errorf("expected content '告警内容: CPU使用率超过80%%', got '%s'", content)
	}
}

// TestEscalationEngine 测试升级引擎
func TestEscalationEngine(t *testing.T) {
	engine := NewEscalationEngine()

	// 创建升级规则
	rule := &EscalationRule{
		ID:           "escalation_1",
		Name:         "测试升级规则",
		AlertMatcher: map[string]string{"severity": "critical"},
		Levels: []EscalationLevel{
			{
				Level:      1,
				After:      5 * time.Minute,
				Channels:   []string{"email"},
				Recipients: []Recipient{{Email: "admin@example.com"}},
			},
			{
				Level:      2,
				After:      15 * time.Minute,
				Channels:   []string{"email", "sms"},
				Recipients: []Recipient{{Email: "manager@example.com"}},
			},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	// 添加升级规则
	err := engine.AddRule(rule)
	if err != nil {
		t.Fatalf("failed to add escalation rule: %v", err)
	}

	// 测试告警
	alert := &AlertInstance{
		ID:       "alert_1",
		Severity: SeverityCritical,
		Labels:   map[string]string{"severity": "critical"},
	}

	// 检查升级（刚触发）
	triggeredAt := time.Now()
	level := engine.CheckEscalation(alert, triggeredAt)
	if level != nil {
		t.Error("expected no escalation level immediately after trigger")
	}

	// 检查升级（5分钟后）
	triggeredAt = time.Now().Add(-6 * time.Minute)
	level = engine.CheckEscalation(alert, triggeredAt)
	if level == nil {
		t.Fatal("expected escalation level 1")
	}
	if level.Level != 1 {
		t.Errorf("expected level 1, got %d", level.Level)
	}

	// 检查升级（15分钟后）
	triggeredAt = time.Now().Add(-16 * time.Minute)
	level = engine.CheckEscalation(alert, triggeredAt)
	if level == nil {
		t.Fatal("expected escalation level 2")
	}
	if level.Level != 2 {
		t.Errorf("expected level 2, got %d", level.Level)
	}
}

// TestNotificationRateLimiter 测试通知限流器
func TestNotificationRateLimiter(t *testing.T) {
	limiter := NewNotificationRateLimiter()

	// 设置限制
	limiter.SetLimit("email", 5, 100)

	// 测试允许发送
	for i := 0; i < 5; i++ {
		if !limiter.Allow("email") {
			t.Errorf("expected allow on iteration %d", i)
		}
	}

	// 测试达到限制
	if limiter.Allow("email") {
		t.Error("expected not allow after reaching limit")
	}

	// 测试未设置限制的渠道
	if !limiter.Allow("sms") {
		t.Error("expected allow for channel without limit")
	}
}

// MockMetricProvider 模拟指标提供者
type MockMetricProvider struct{}

func (m *MockMetricProvider) Query(ctx context.Context, metricName string, labels map[string]string, start, end time.Time) ([]MetricData, error) {
	return []MetricData{
		{
			Name:      metricName,
			Value:     100.0,
			Timestamp: time.Now(),
			Labels:    labels,
		},
	}, nil
}

func (m *MockMetricProvider) QueryLatest(ctx context.Context, metricName string, labels map[string]string) (*MetricData, error) {
	return &MetricData{
		Name:      metricName,
		Value:     100.0,
		Timestamp: time.Now(),
		Labels:    labels,
	}, nil
}

func (m *MockMetricProvider) QueryRange(ctx context.Context, metricName string, labels map[string]string, duration time.Duration) ([]MetricData, error) {
	return []MetricData{
		{
			Name:      metricName,
			Value:     100.0,
			Timestamp: time.Now(),
			Labels:    labels,
		},
	}, nil
}
