package alerting

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCov_AlertService_StartStop(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	require.NotNil(t, service)
	err := service.Start()
	require.NoError(t, err)
	service.Stop()
}

func TestCov_AlertService_AddRemoveRule(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	rule := &AlertRule{
		ID:       "svc-rule-1",
		Name:     "svc test rule",
		Category: CategorySystemResource,
		Severity: SeverityWarning,
		Enabled:  true,
		Condition: AlertCondition{
			MetricName: "test_metric",
			Operator:   OpGT,
			Threshold:  50,
		},
	}
	err := service.AddRule(rule)
	require.NoError(t, err)
	service.RemoveRule("svc-rule-1")
}

func TestCov_AlertService_CreateAlert(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	alert := &AlertInstance{
		ID:          "svc-alert-1",
		RuleID:      "rule-1",
		RuleName:    "test",
		Category:    CategorySystemResource,
		Severity:    SeverityWarning,
		Title:       "test alert",
		Message:     "test message",
		Value:       100.0,
		Threshold:   80.0,
		TriggeredAt: time.Now(),
		Source:      "server-1",
	}
	err := service.CreateAlert(alert)
	assert.NoError(t, err)
}

func TestCov_AlertService_SilenceOperations(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	silence := &Silence{
		ID:        "svc-silence-1",
		Matchers:  map[string]string{"source": "server-1"},
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
		Reason:    "maintenance",
		CreatedBy: "admin",
		CreatedAt: time.Now(),
	}
	err := service.AddSilence(silence)
	require.NoError(t, err)
	service.RemoveSilence("svc-silence-1")
}

func TestCov_AlertService_SuppressionOperations(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	rule := &SuppressionRule{
		ID:             "svc-sup-1",
		Name:           "test suppression",
		SourceMatchers: map[string]string{"severity": "critical"},
		TargetMatchers: map[string]string{"severity": "warning"},
		Enabled:        true,
		CreatedAt:      time.Now(),
	}
	err := service.AddSuppressionRule(rule)
	require.NoError(t, err)
	service.RemoveSuppressionRule("svc-sup-1")
}

func TestCov_AlertService_NotificationOperations(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	channel := &EmailChannel{}
	service.RegisterNotificationChannel(channel)
	service.UnregisterNotificationChannel("email")
}

func TestCov_AlertService_TemplateOperations(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	tmpl := &NotificationTemplate{
		ID:              "svc-tmpl-1",
		Name:            "svc template",
		Channel:         "email",
		SubjectTemplate: "Alert: {{.alert.Title}}",
		ContentTemplate: "Content: {{.alert.Message}}",
	}
	err := service.AddNotificationTemplate(tmpl)
	require.NoError(t, err)
	service.RemoveNotificationTemplate("svc-tmpl-1")
}

func TestCov_AlertService_EscalationOperations(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	rule := &EscalationRule{
		ID:           "svc-esc-1",
		Name:         "svc escalation",
		AlertMatcher: map[string]string{"severity": "critical"},
		Levels: []EscalationLevel{
			{Level: 1, After: 5 * time.Minute, Channels: []string{"email"}},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
	}
	err := service.AddEscalationRule(rule)
	require.NoError(t, err)
	service.RemoveEscalationRule("svc-esc-1")
}

func TestCov_AlertService_GetAPI(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	api := service.GetAPI()
	assert.NotNil(t, api)
}

func TestCov_AlertService_GetAggregator(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	agg := service.GetAggregator()
	assert.NotNil(t, agg)
}

func TestCov_AlertService_GetNotifier(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	notifier := service.GetNotifier()
	assert.NotNil(t, notifier)
}

func TestCov_AlertService_GetRuleManager(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	rm := service.GetRuleManager()
	assert.NotNil(t, rm)
}

func TestCov_AlertService_GetStats(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	stats := service.GetStats()
	assert.NotNil(t, stats)
}

func TestCov_SilenceChecker(t *testing.T) {
	checker := NewSilenceChecker()
	assert.False(t, checker.IsSilent("alert-1"))
	checker.StartSilence("alert-1", 1*time.Hour, "maintenance")
	assert.True(t, checker.IsSilent("alert-1"))
	checker.EndSilence("alert-1")
	assert.False(t, checker.IsSilent("alert-1"))
}

func TestCov_SilenceChecker_CleanupExpired(t *testing.T) {
	checker := NewSilenceChecker()
	checker.StartSilence("ch1", 1*time.Nanosecond, "test")
	checker.StartSilence("ch2", 1*time.Hour, "test")
	time.Sleep(10 * time.Millisecond)
	checker.CleanupExpired()
	assert.False(t, checker.IsSilent("ch1"))
	assert.True(t, checker.IsSilent("ch2"))
}

func TestCov_AlertGroup_Add(t *testing.T) {
	group := NewAlertGroup(StrategyByRule, "rule-1")
	assert.Equal(t, 0, group.Count)
	alert := &AlertInstance{ID: "grp-1", Title: "Group Alert", TriggeredAt: time.Now()}
	group.Add(alert)
	assert.Equal(t, 1, group.Count)
}

func TestCov_AlertAggregator_GetGroup(t *testing.T) {
	config := DefaultAggregatorConfig()
	config.EnableDeduplication = false
	aggregator := NewAlertAggregator(config)
	alert := &AlertInstance{
		ID:          "agg-1",
		RuleID:      "rule-1",
		RuleName:    "test rule",
		Category:    CategorySystemResource,
		Severity:    SeverityWarning,
		Title:       "test",
		Message:     "msg",
		Value:       100.0,
		Threshold:   80.0,
		TriggeredAt: time.Now(),
		Source:      "server-1",
	}
	_, _, err := aggregator.Aggregate(context.Background(), alert)
	require.NoError(t, err)
	groups := aggregator.GetAllGroups()
	assert.NotEmpty(t, groups)
}

func TestCov_AlertAggregator_GetAllGroups(t *testing.T) {
	config := DefaultAggregatorConfig()
	config.EnableDeduplication = false
	aggregator := NewAlertAggregator(config)
	alert1 := &AlertInstance{ID: "a1", RuleID: "r1", RuleName: "rn1", Category: CategorySystemResource, Severity: SeverityWarning, Title: "t1", TriggeredAt: time.Now(), Source: "s1"}
	alert2 := &AlertInstance{ID: "a2", RuleID: "r2", RuleName: "rn2", Category: CategorySystemResource, Severity: SeverityCritical, Title: "t2", TriggeredAt: time.Now(), Source: "s2"}
	_, _, err := aggregator.Aggregate(context.Background(), alert1)
	require.NoError(t, err)
	_, _, err = aggregator.Aggregate(context.Background(), alert2)
	require.NoError(t, err)
	groups := aggregator.GetAllGroups()
	assert.GreaterOrEqual(t, len(groups), 1)
}

func TestCov_AlertAggregator_Flush(t *testing.T) {
	config := DefaultAggregatorConfig()
	config.EnableDeduplication = false
	aggregator := NewAlertAggregator(config)
	alert := &AlertInstance{ID: "flush-1", RuleID: "r1", RuleName: "rn1", Category: CategorySystemResource, Severity: SeverityWarning, Title: "t1", TriggeredAt: time.Now(), Source: "s1"}
	_, _, err := aggregator.Aggregate(context.Background(), alert)
	require.NoError(t, err)
	_ = aggregator.Flush(context.Background())
	groups := aggregator.GetAllGroups()
	assert.Empty(t, groups)
}

func TestCov_AlertAggregator_TriggerFlush(t *testing.T) {
	config := DefaultAggregatorConfig()
	aggregator := NewAlertAggregator(config)
	aggregator.TriggerFlush()
}

func TestCov_AlertAggregator_StartStop(t *testing.T) {
	config := DefaultAggregatorConfig()
	config.FlushInterval = 1 * time.Second
	aggregator := NewAlertAggregator(config)
	aggregator.Start(context.Background())
	time.Sleep(100 * time.Millisecond)
	aggregator.Stop()
}

func TestCov_SilenceManager_GetSilence(t *testing.T) {
	manager := NewSilenceManager()
	silence := &Silence{
		ID:        "sil-get-1",
		Matchers:  map[string]string{"source": "server-1"},
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
		Reason:    "test",
		CreatedBy: "admin",
		CreatedAt: time.Now(),
	}
	_ = manager.AddSilence(silence)
	retrieved, ok := manager.GetSilence("sil-get-1")
	assert.True(t, ok)
	assert.Equal(t, "test", retrieved.Reason)
	_, ok = manager.GetSilence("nonexistent")
	assert.False(t, ok)
}

func TestCov_SilenceManager_ListSilences(t *testing.T) {
	manager := NewSilenceManager()
	_ = manager.AddSilence(&Silence{ID: "sil-l1", Matchers: map[string]string{"a": "b"}, StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour), CreatedAt: time.Now()})
	_ = manager.AddSilence(&Silence{ID: "sil-l2", Matchers: map[string]string{"c": "d"}, StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour), CreatedAt: time.Now()})
	list := manager.ListSilences()
	assert.Len(t, list, 2)
}

func TestCov_SilenceManager_CleanupExpired(t *testing.T) {
	manager := NewSilenceManager()
	_ = manager.AddSilence(&Silence{ID: "sil-exp", Matchers: map[string]string{"a": "b"}, StartTime: time.Now().Add(-2 * time.Hour), EndTime: time.Now().Add(-1 * time.Hour), CreatedAt: time.Now()})
	_ = manager.AddSilence(&Silence{ID: "sil-act", Matchers: map[string]string{"c": "d"}, StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour), CreatedAt: time.Now()})
	manager.CleanupExpired()
	_, ok := manager.GetSilence("sil-exp")
	assert.False(t, ok)
	_, ok = manager.GetSilence("sil-act")
	assert.True(t, ok)
}

func TestCov_AlertDeduplicator_Cleanup(t *testing.T) {
	dedup := NewAlertDeduplicator(5 * time.Minute)
	alert := &AlertInstance{ID: "dedup-1", RuleID: "r1", RuleName: "rn", Message: "msg", Value: 100, TriggeredAt: time.Now(), Source: "s1"}
	dedup.Record(alert)
	dedup.Cleanup()
}

func TestCov_RuleEvaluator_Evaluate(t *testing.T) {
	provider := &MockMetricProvider{}
	rule := &AlertRule{
		ID:       "eval-1",
		Name:     "eval rule",
		Category: CategorySystemResource,
		Severity: SeverityWarning,
		Enabled:  true,
		Condition: AlertCondition{
			MetricName: "test_metric",
			Operator:   OpGT,
			Threshold:  50,
		},
	}
	evaluator := NewRuleEvaluator(rule, provider)
	result, err := evaluator.Evaluate(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCov_RuleEvaluator_EvaluateAllOperators(t *testing.T) {
	provider := &MockMetricProvider{}
	operators := []ComparisonOperator{OpGT, OpGTE, OpLT, OpLTE, OpEqual, OpNotEqual}
	for _, op := range operators {
		rule := &AlertRule{
			ID:       "eval-op-" + string(op),
			Name:     "eval op rule",
			Category: CategorySystemResource,
			Severity: SeverityWarning,
			Enabled:  true,
			Condition: AlertCondition{
				MetricName: "test_metric",
				Operator:   op,
				Threshold:  50,
			},
		}
		evaluator := NewRuleEvaluator(rule, provider)
		result, err := evaluator.Evaluate(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, result)
	}
}

func TestCov_RuleEvaluator_EvaluateWithAggregation(t *testing.T) {
	provider := &MockMetricProvider{}
	rule := &AlertRule{
		ID:       "eval-agg-1",
		Name:     "eval agg rule",
		Category: CategorySystemResource,
		Severity: SeverityWarning,
		Enabled:  true,
		Condition: AlertCondition{
			MetricName:  "test_metric",
			Operator:    OpGT,
			Threshold:   50,
			Aggregation: AggAvg,
		},
	}
	evaluator := NewRuleEvaluator(rule, provider)
	result, err := evaluator.Evaluate(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCov_RuleManager_EvaluateRule(t *testing.T) {
	provider := &MockMetricProvider{}
	manager := NewRuleManager(provider)
	rule := &AlertRule{
		ID:       "rm-eval-1",
		Name:     "rm eval rule",
		Category: CategorySystemResource,
		Severity: SeverityWarning,
		Enabled:  true,
		Condition: AlertCondition{
			MetricName: "test_metric",
			Operator:   OpGT,
			Threshold:  50,
		},
	}
	manager.AddRule(rule)
	result, err := manager.EvaluateRule(context.Background(), "rm-eval-1")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCov_RuleManager_EvaluateRule_NotFound(t *testing.T) {
	provider := &MockMetricProvider{}
	manager := NewRuleManager(provider)
	_, err := manager.EvaluateRule(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestCov_RuleManager_EvaluateAll(t *testing.T) {
	provider := &MockMetricProvider{}
	manager := NewRuleManager(provider)
	rule1 := &AlertRule{ID: "ea-1", Name: "rule1", Category: CategorySystemResource, Severity: SeverityWarning, Enabled: true, Condition: AlertCondition{MetricName: "m1", Operator: OpGT, Threshold: 50}}
	rule2 := &AlertRule{ID: "ea-2", Name: "rule2", Category: CategorySystemResource, Severity: SeverityCritical, Enabled: true, Condition: AlertCondition{MetricName: "m2", Operator: OpLT, Threshold: 10}}
	manager.AddRule(rule1)
	manager.AddRule(rule2)
	results, err := manager.EvaluateAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestCov_RuleManager_GetEnabledRules(t *testing.T) {
	provider := &MockMetricProvider{}
	manager := NewRuleManager(provider)
	manager.AddRule(&AlertRule{ID: "en-1", Name: "enabled", Enabled: true, Condition: AlertCondition{MetricName: "m1", Operator: OpGT, Threshold: 50}})
	manager.AddRule(&AlertRule{ID: "en-2", Name: "disabled", Enabled: false, Condition: AlertCondition{MetricName: "m2", Operator: OpGT, Threshold: 50}})
	enabled := manager.GetEnabledRules()
	assert.Len(t, enabled, 1)
}

func TestCov_SystemResourceAlertRules(t *testing.T) {
	rules := NewSystemResourceAlertRules()
	cpuRule := rules.CPUUsageRule(80.0, 5*time.Minute)
	assert.Equal(t, "system_cpu_usage", cpuRule.ID)
	memRule := rules.MemoryUsageRule(85.0, 5*time.Minute)
	assert.Equal(t, "system_memory_usage", memRule.ID)
	diskRule := rules.DiskUsageRule(90.0, 5*time.Minute)
	assert.Equal(t, "system_disk_usage", diskRule.ID)
}

func TestCov_ServiceHealthAlertRules(t *testing.T) {
	rules := NewServiceHealthAlertRules()
	require.NotNil(t, rules)
	downRule := rules.ServiceDownRule("my-service")
	assert.NotNil(t, downRule)
	errRateRule := rules.ServiceErrorRateRule("my-service", 5.0)
	assert.NotNil(t, errRateRule)
	respTimeRule := rules.ServiceResponseTimeRule("my-service", 500.0)
	assert.NotNil(t, respTimeRule)
}

func TestCov_BusinessAlertRules(t *testing.T) {
	rules := NewBusinessAlertRules()
	require.NotNil(t, rules)
	offlineRule := rules.StationOfflineRule()
	assert.NotNil(t, offlineRule)
	anomalyRule := rules.PowerGenerationAnomalyRule(20.0)
	assert.NotNil(t, anomalyRule)
	faultRule := rules.DeviceFaultRule()
	assert.NotNil(t, faultRule)
}

func TestCov_NotificationTemplateManager_RemoveTemplate(t *testing.T) {
	manager := NewNotificationTemplateManager()
	tmpl := &NotificationTemplate{ID: "rm-tmpl-1", Name: "test", Channel: "email", SubjectTemplate: "sub", ContentTemplate: "content"}
	manager.AddTemplate(tmpl)
	manager.RemoveTemplate("rm-tmpl-1")
	_, err := manager.GetTemplate("rm-tmpl-1")
	assert.Error(t, err)
}

func TestCov_NotificationTemplateManager_GetTemplate_NotFound(t *testing.T) {
	manager := NewNotificationTemplateManager()
	_, err := manager.GetTemplate("nonexistent")
	assert.Error(t, err)
}

func TestCov_NotificationTemplateManager_RemoveTemplate_NotFound(t *testing.T) {
	manager := NewNotificationTemplateManager()
	manager.RemoveTemplate("nonexistent")
}

func TestCov_NotificationTemplate_RenderHTML(t *testing.T) {
	tmpl := &NotificationTemplate{
		ID:              "html-tmpl",
		Name:            "html template",
		Channel:         "email",
		SubjectTemplate: "Alert: {{.alert.Title}}",
		ContentTemplate: "Content: {{.alert.Message}}",
		HTMLTemplate:    "<h1>{{.alert.Title}}</h1><p>{{.alert.Message}}</p>",
	}
	data := map[string]interface{}{
		"alert": &AlertInstance{Title: "CPU Alert", Message: "CPU is high"},
	}
	html, err := tmpl.RenderHTML(data)
	require.NoError(t, err)
	assert.Contains(t, html, "CPU Alert")
}

func TestCov_EmailChannel(t *testing.T) {
	ch := &EmailChannel{}
	assert.Equal(t, "email", ch.Name())
}

func TestCov_SMSChannel(t *testing.T) {
	ch := &SMSChannel{}
	assert.Equal(t, "sms", ch.Name())
}

func TestCov_DingTalkChannel(t *testing.T) {
	ch := &DingTalkChannel{}
	assert.Equal(t, "dingtalk", ch.Name())
}

func TestCov_WeChatChannel(t *testing.T) {
	ch := &WeChatChannel{}
	assert.Equal(t, "wechat", ch.Name())
}

func TestCov_AlertAPI_RegisterRoutes(t *testing.T) {
	cfg := AlertServiceConfig{MetricProvider: &MockMetricProvider{}}
	service := NewAlertService(cfg)
	api := service.GetAPI()
	require.NotNil(t, api)
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)
}
