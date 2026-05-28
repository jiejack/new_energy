package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertService_New(t *testing.T) {
	config := AlertServiceConfig{
		MetricProvider: &MockMetricProvider{},
		AggregatorConfig: AggregatorConfig{Strategy: StrategyBySource, WindowDuration: 5 * time.Minute, FlushInterval: 30 * time.Second, MaxGroupSize: 100, EnableDeduplication: true, DedupWindow: 5 * time.Minute, EnableAutoFlush: true},
	}
	svc := NewAlertService(config)
	require.NotNil(t, svc)
	assert.NotNil(t, svc.GetAPI())
	assert.NotNil(t, svc.GetAggregator())
	assert.NotNil(t, svc.GetNotifier())
	assert.NotNil(t, svc.GetRuleManager())
}

func TestAlertService_StartStop(t *testing.T) {
	config := AlertServiceConfig{
		MetricProvider: &MockMetricProvider{},
		AggregatorConfig: AggregatorConfig{Strategy: StrategyBySource, WindowDuration: 5 * time.Minute, FlushInterval: 30 * time.Second, MaxGroupSize: 100, EnableDeduplication: true, DedupWindow: 5 * time.Minute, EnableAutoFlush: true},
	}
	svc := NewAlertService(config)

	err := svc.Start()
	require.NoError(t, err)

	svc.Stop()
}

func TestAlertService_AddRemoveRule(t *testing.T) {
	config := AlertServiceConfig{
		MetricProvider: &MockMetricProvider{},
		AggregatorConfig: AggregatorConfig{Strategy: StrategyBySource, WindowDuration: 5 * time.Minute, FlushInterval: 30 * time.Second, MaxGroupSize: 100, EnableDeduplication: true, DedupWindow: 5 * time.Minute, EnableAutoFlush: true},
	}
	svc := NewAlertService(config)

	rule := &AlertRule{
		ID:       "rule-1",
		Name:     "Test",
		Category: CategorySystemResource,
		Severity: SeverityWarning,
		Enabled:  true,
		Condition: AlertCondition{
			MetricName: "test_metric",
			Operator:   OpGT,
			Threshold:  100,
		},
	}

	err := svc.AddRule(rule)
	require.NoError(t, err)

	found, exists := svc.GetRule("rule-1")
	assert.True(t, exists)
	assert.Equal(t, "Test", found.Name)

	svc.RemoveRule("rule-1")
	_, exists = svc.GetRule("rule-1")
	assert.False(t, exists)
}

func TestAlertService_GetAllRules(t *testing.T) {
	config := AlertServiceConfig{
		MetricProvider: &MockMetricProvider{},
		AggregatorConfig: AggregatorConfig{Strategy: StrategyBySource, WindowDuration: 5 * time.Minute, FlushInterval: 30 * time.Second, MaxGroupSize: 100, EnableDeduplication: true, DedupWindow: 5 * time.Minute, EnableAutoFlush: true},
	}
	svc := NewAlertService(config)

	svc.AddRule(&AlertRule{ID: "r1", Name: "Rule 1", Enabled: true})
	svc.AddRule(&AlertRule{ID: "r2", Name: "Rule 2", Enabled: false})

	rules := svc.GetAllRules()
	assert.Equal(t, 2, len(rules))
}

func TestAlertService_CreateAlert(t *testing.T) {
	config := AlertServiceConfig{
		MetricProvider: &MockMetricProvider{},
		AggregatorConfig: AggregatorConfig{
			Strategy:    StrategyBySource,
			MaxGroupSize: 100,
		},
	}
	svc := NewAlertService(config)

	alert := &AlertInstance{
		ID:          "alert-1",
		RuleID:      "rule-1",
		RuleName:    "Test Rule",
		Category:    CategorySystemResource,
		Severity:    SeverityWarning,
		Title:       "Test Alert",
		Message:     "Test message",
		Value:       85.0,
		Threshold:   80.0,
		TriggeredAt: time.Now(),
		Source:      "server-1",
	}

	err := svc.CreateAlert(alert)
	require.NoError(t, err)
}

func TestAlertService_Silence(t *testing.T) {
	config := AlertServiceConfig{
		MetricProvider: &MockMetricProvider{},
		AggregatorConfig: AggregatorConfig{Strategy: StrategyBySource, WindowDuration: 5 * time.Minute, FlushInterval: 30 * time.Second, MaxGroupSize: 100, EnableDeduplication: true, DedupWindow: 5 * time.Minute, EnableAutoFlush: true},
	}
	svc := NewAlertService(config)

	silence := &Silence{
		ID:        "silence-1",
		Matchers:  map[string]string{"source": "server-1"},
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
		Reason:    "Maintenance",
		CreatedBy: "admin",
	}

	err := svc.AddSilence(silence)
	require.NoError(t, err)

	svc.RemoveSilence("silence-1")
}

func TestAlertService_SuppressionRule(t *testing.T) {
	config := AlertServiceConfig{
		MetricProvider: &MockMetricProvider{},
		AggregatorConfig: AggregatorConfig{Strategy: StrategyBySource, WindowDuration: 5 * time.Minute, FlushInterval: 30 * time.Second, MaxGroupSize: 100, EnableDeduplication: true, DedupWindow: 5 * time.Minute, EnableAutoFlush: true},
	}
	svc := NewAlertService(config)

	rule := &SuppressionRule{
		ID:             "supp-1",
		Name:           "Test Suppression",
		TargetMatchers: map[string]string{"severity": "warning"},
		Enabled:        true,
	}

	err := svc.AddSuppressionRule(rule)
	require.NoError(t, err)

	svc.RemoveSuppressionRule("supp-1")
}

func TestAlertService_NotificationChannel(t *testing.T) {
	config := AlertServiceConfig{
		MetricProvider: &MockMetricProvider{},
		AggregatorConfig: AggregatorConfig{Strategy: StrategyBySource, WindowDuration: 5 * time.Minute, FlushInterval: 30 * time.Second, MaxGroupSize: 100, EnableDeduplication: true, DedupWindow: 5 * time.Minute, EnableAutoFlush: true},
	}
	svc := NewAlertService(config)

	emailChannel := NewEmailChannel(&EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
	})
	svc.RegisterNotificationChannel(emailChannel)

	smsChannel := NewSMSChannel(&SMSConfig{
		Provider: "aliyun",
	})
	svc.RegisterNotificationChannel(smsChannel)

	dingtalkChannel := NewDingTalkChannel(&DingTalkConfig{
		WebhookURL: "https://oapi.dingtalk.com/robot/send",
	})
	svc.RegisterNotificationChannel(dingtalkChannel)

	wechatChannel := NewWeChatChannel(&WeChatConfig{
		CorpID:  "test-corp",
		AgentID: "test-agent",
	})
	svc.RegisterNotificationChannel(wechatChannel)

	svc.UnregisterNotificationChannel("email")
}

func TestAlertService_NotificationTemplate(t *testing.T) {
	config := AlertServiceConfig{
		MetricProvider: &MockMetricProvider{},
		AggregatorConfig: AggregatorConfig{Strategy: StrategyBySource, WindowDuration: 5 * time.Minute, FlushInterval: 30 * time.Second, MaxGroupSize: 100, EnableDeduplication: true, DedupWindow: 5 * time.Minute, EnableAutoFlush: true},
	}
	svc := NewAlertService(config)

	tmpl := &NotificationTemplate{
		ID:              "tmpl-1",
		Name:            "Test Template",
		Channel:         "email",
		SubjectTemplate: "Alert: {{.alert.Title}}",
		ContentTemplate: "Content: {{.alert.Message}}",
	}

	err := svc.AddNotificationTemplate(tmpl)
	require.NoError(t, err)

	svc.RemoveNotificationTemplate("tmpl-1")
}

func TestAlertService_EscalationRule(t *testing.T) {
	config := AlertServiceConfig{
		MetricProvider: &MockMetricProvider{},
		AggregatorConfig: AggregatorConfig{Strategy: StrategyBySource, WindowDuration: 5 * time.Minute, FlushInterval: 30 * time.Second, MaxGroupSize: 100, EnableDeduplication: true, DedupWindow: 5 * time.Minute, EnableAutoFlush: true},
	}
	svc := NewAlertService(config)

	rule := &EscalationRule{
		ID:           "esc-1",
		Name:         "Test Escalation",
		AlertMatcher: map[string]string{"severity": "critical"},
		Levels: []EscalationLevel{
			{Level: 1, After: 5 * time.Minute, Channels: []string{"email"}},
		},
		Enabled: true,
	}

	err := svc.AddEscalationRule(rule)
	require.NoError(t, err)

	svc.RemoveEscalationRule("esc-1")
}

func TestAlertService_GetStats(t *testing.T) {
	config := AlertServiceConfig{
		MetricProvider: &MockMetricProvider{},
		AggregatorConfig: AggregatorConfig{Strategy: StrategyBySource, WindowDuration: 5 * time.Minute, FlushInterval: 30 * time.Second, MaxGroupSize: 100, EnableDeduplication: true, DedupWindow: 5 * time.Minute, EnableAutoFlush: true},
	}
	svc := NewAlertService(config)

	svc.AddRule(&AlertRule{ID: "r1", Name: "Rule 1", Enabled: true})
	svc.AddRule(&AlertRule{ID: "r2", Name: "Rule 2", Enabled: false})

	stats := svc.GetStats()
	assert.Equal(t, 2, stats.TotalRules)
	assert.Equal(t, 1, stats.EnabledRules)
}

func TestRuleManager_AddRule_NoID(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	err := rm.AddRule(&AlertRule{Name: "No ID"})
	assert.Error(t, err)
}

func TestRuleManager_EvaluateRule(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	rm.AddRule(&AlertRule{
		ID:      "rule-1",
		Name:    "Test",
		Enabled: true,
		Condition: AlertCondition{
			MetricName: "test_metric",
			Operator:   OpGT,
			Threshold:  50,
		},
	})

	result, err := rm.EvaluateRule(context.Background(), "rule-1")
	require.NoError(t, err)
	assert.True(t, result.Triggered)
	assert.Equal(t, 100.0, result.Value)
}

func TestRuleManager_EvaluateRule_NotFound(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	_, err := rm.EvaluateRule(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestRuleManager_EvaluateRule_Disabled(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	rm.AddRule(&AlertRule{
		ID:      "rule-1",
		Name:    "Test",
		Enabled: false,
		Condition: AlertCondition{
			MetricName: "test_metric",
			Operator:   OpGT,
			Threshold:  50,
		},
	})

	result, err := rm.EvaluateRule(context.Background(), "rule-1")
	require.NoError(t, err)
	assert.False(t, result.Triggered)
}

func TestRuleManager_EvaluateAll(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	rm.AddRule(&AlertRule{ID: "r1", Name: "Rule 1", Enabled: true, Condition: AlertCondition{MetricName: "m1", Operator: OpGT, Threshold: 50}})
	rm.AddRule(&AlertRule{ID: "r2", Name: "Rule 2", Enabled: true, Condition: AlertCondition{MetricName: "m2", Operator: OpLT, Threshold: 50}})

	results, err := rm.EvaluateAll(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestRuleEvaluator_Operators(t *testing.T) {
	tests := []struct {
		operator  ComparisonOperator
		threshold float64
		expected  bool
	}{
		{OpEqual, 100, true},
		{OpEqual, 99, false},
		{OpNotEqual, 99, true},
		{OpNotEqual, 100, false},
		{OpGT, 99, true},
		{OpGT, 100, false},
		{OpGTE, 100, true},
		{OpGTE, 101, false},
		{OpLT, 101, true},
		{OpLT, 100, false},
		{OpLTE, 100, true},
		{OpLTE, 99, false},
	}

	for _, tt := range tests {
		rm := NewRuleManager(&MockMetricProvider{})
		rm.AddRule(&AlertRule{
			ID:      "test-rule",
			Enabled: true,
			Condition: AlertCondition{
				MetricName: "test_metric",
				Operator:   tt.operator,
				Threshold:  tt.threshold,
			},
		})
		result, err := rm.EvaluateRule(context.Background(), "test-rule")
		require.NoError(t, err)
		assert.Equal(t, tt.expected, result.Triggered, "operator %s, threshold=%.0f", tt.operator, tt.threshold)
	}
}

func TestRuleEvaluator_Aggregations(t *testing.T) {
	provider := &multiValueProvider{}
	tests := []struct {
		agg      AggregationFunc
		expected float64
	}{
		{AggAvg, 150.0},
		{AggMax, 200.0},
		{AggMin, 100.0},
		{AggSum, 300.0},
		{AggCount, 2.0},
	}

	for _, tt := range tests {
		rm := NewRuleManager(provider)
		rm.AddRule(&AlertRule{
			ID:      "test-rule",
			Enabled: true,
			Condition: AlertCondition{
				MetricName:         "test_metric",
				Operator:           OpGT,
				Threshold:          50,
				Aggregation:        tt.agg,
				AggregationWindow:  5 * time.Minute,
			},
		})
		result, err := rm.EvaluateRule(context.Background(), "test-rule")
		require.NoError(t, err)
		assert.Equal(t, tt.expected, result.Value, "aggregation %s", tt.agg)
	}
}

func TestRuleEvaluator_RateAggregation(t *testing.T) {
	rm := NewRuleManager(&rateProvider{})
	rm.AddRule(&AlertRule{
		ID:      "rate-rule",
		Enabled: true,
		Condition: AlertCondition{
			MetricName:        "test_metric",
			Operator:          OpGT,
			Threshold:         0,
			Aggregation:       AggRate,
			AggregationWindow: 5 * time.Minute,
		},
	})

	result, err := rm.EvaluateRule(context.Background(), "rate-rule")
	require.NoError(t, err)
	assert.True(t, result.Triggered)
}

func TestRuleEvaluator_DurationCheck(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	rm.AddRule(&AlertRule{
		ID:      "duration-rule",
		Enabled: true,
		Condition: AlertCondition{
			MetricName: "test_metric",
			Operator:   OpGT,
			Threshold:  50,
			Duration:   1 * time.Hour,
		},
	})

	result, err := rm.EvaluateRule(context.Background(), "duration-rule")
	require.NoError(t, err)
	assert.False(t, result.Triggered)
}

func TestRuleEvaluator_DefaultAggregation(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	rm.AddRule(&AlertRule{
		ID:      "default-agg-rule",
		Enabled: true,
		Condition: AlertCondition{
			MetricName: "test_metric",
			Operator:   OpGT,
			Threshold:  50,
		},
	})

	result, err := rm.EvaluateRule(context.Background(), "default-agg-rule")
	require.NoError(t, err)
	assert.True(t, result.Triggered)
	assert.Equal(t, 100.0, result.Value)
}

func TestAlertAggregator_Strategies(t *testing.T) {
	strategies := []AggregationStrategy{
		StrategyBySource,
		StrategyBySeverity,
		StrategyByCategory,
		StrategyByRule,
		StrategyComposite,
	}

	for _, strategy := range strategies {
		config := AggregatorConfig{Strategy: strategy, MaxGroupSize: 100}
		agg := NewAlertAggregator(config)

		alert := &AlertInstance{
			ID:          "alert-1",
			RuleID:      "rule-1",
			Category:    CategorySystemResource,
			Severity:    SeverityWarning,
			Title:       "Test",
			Message:     "Test msg",
			TriggeredAt: time.Now(),
			Source:      "server-1",
		}

		group, isNew, err := agg.Aggregate(context.Background(), alert)
		require.NoError(t, err, "strategy: %s", strategy)
		assert.True(t, isNew, "strategy: %s", strategy)
		assert.NotNil(t, group, "strategy: %s", strategy)
	}
}

func TestAlertAggregator_TimeWindowStrategy(t *testing.T) {
	config := AggregatorConfig{
		Strategy:       StrategyByTimeWindow,
		WindowDuration: 1 * time.Hour,
		MaxGroupSize:   100,
	}
	agg := NewAlertAggregator(config)

	alert := &AlertInstance{
		ID:          "alert-1",
		RuleID:      "rule-1",
		Category:    CategorySystemResource,
		Severity:    SeverityWarning,
		Title:       "Test",
		Message:     "Test msg",
		TriggeredAt: time.Now(),
		Source:      "server-1",
	}

	group, isNew, err := agg.Aggregate(context.Background(), alert)
	require.NoError(t, err)
	assert.True(t, isNew)
	assert.NotNil(t, group)
}

func TestAlertAggregator_LabelsStrategy(t *testing.T) {
	config := AggregatorConfig{
		Strategy:     StrategyByLabels,
		MaxGroupSize: 100,
	}
	agg := NewAlertAggregator(config)

	alert := &AlertInstance{
		ID:          "alert-1",
		RuleID:      "rule-1",
		Category:    CategorySystemResource,
		Severity:    SeverityWarning,
		Title:       "Test",
		Message:     "Test msg",
		TriggeredAt: time.Now(),
		Source:      "server-1",
		Labels:      map[string]string{"env": "prod"},
	}

	group, isNew, err := agg.Aggregate(context.Background(), alert)
	require.NoError(t, err)
	assert.True(t, isNew)
	assert.NotNil(t, group)
}

func TestAlertAggregator_Deduplication(t *testing.T) {
	config := AggregatorConfig{
		Strategy:            StrategyBySource,
		EnableDeduplication: true,
		DedupWindow:         5 * time.Minute,
		MaxGroupSize:        100,
	}
	agg := NewAlertAggregator(config)

	alert := &AlertInstance{
		ID:          "alert-1",
		RuleID:      "rule-1",
		Message:     "Test msg",
		TriggeredAt: time.Now(),
		Source:      "server-1",
	}

	_, _, err := agg.Aggregate(context.Background(), alert)
	require.NoError(t, err)

	_, _, err = agg.Aggregate(context.Background(), alert)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestAlertAggregator_Silenced(t *testing.T) {
	config := AggregatorConfig{
		Strategy:     StrategyBySource,
		MaxGroupSize: 100,
	}
	agg := NewAlertAggregator(config)

	agg.silenceManager.AddSilence(&Silence{
		ID:        "sil-1",
		Matchers:  map[string]string{"source": "server-1"},
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
	})

	alert := &AlertInstance{
		ID:          "alert-1",
		RuleID:      "rule-1",
		Message:     "Test msg",
		TriggeredAt: time.Now(),
		Source:      "server-1",
		Labels:      map[string]string{"source": "server-1"},
	}

	_, _, err := agg.Aggregate(context.Background(), alert)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "silenced")
}

func TestAlertAggregator_Suppressed(t *testing.T) {
	config := AggregatorConfig{
		Strategy:     StrategyBySource,
		MaxGroupSize: 100,
	}
	agg := NewAlertAggregator(config)

	agg.suppressionEngine.AddRule(&SuppressionRule{
		ID:             "supp-1",
		Name:           "Test",
		TargetMatchers: map[string]string{"source": "server-1"},
		Enabled:        true,
	})

	alert := &AlertInstance{
		ID:          "alert-1",
		RuleID:      "rule-1",
		Message:     "Test msg",
		TriggeredAt: time.Now(),
		Source:      "server-1",
		Labels:      map[string]string{"source": "server-1"},
	}

	_, _, err := agg.Aggregate(context.Background(), alert)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suppressed")
}

func TestAlertAggregator_MaxGroupSize(t *testing.T) {
	config := AggregatorConfig{
		Strategy:     StrategyBySource,
		MaxGroupSize: 2,
	}
	agg := NewAlertAggregator(config)

	for i := 0; i < 3; i++ {
		alert := &AlertInstance{
			ID:          fmt.Sprintf("alert-%d", i),
			RuleID:      "rule-1",
			Message:     fmt.Sprintf("Msg %d", i),
			TriggeredAt: time.Now(),
			Source:      "server-1",
		}
		_, _, err := agg.Aggregate(context.Background(), alert)
		require.NoError(t, err)
	}
}

func TestAlertAggregator_Flush(t *testing.T) {
	config := AggregatorConfig{
		Strategy:     StrategyBySource,
		MaxGroupSize: 100,
		MinGroupSize: 1,
	}
	agg := NewAlertAggregator(config)

	alert := &AlertInstance{
		ID:          "alert-1",
		RuleID:      "rule-1",
		Message:     "Test msg",
		TriggeredAt: time.Now(),
		Source:      "server-1",
	}
	agg.Aggregate(context.Background(), alert)

	err := agg.Flush(context.Background())
	require.NoError(t, err)
}

func TestAlertAggregator_TriggerFlush(t *testing.T) {
	config := AggregatorConfig{
		Strategy:        StrategyBySource,
		MaxGroupSize:    100,
		EnableAutoFlush: true,
		FlushInterval:   1 * time.Hour,
	}
	agg := NewAlertAggregator(config)
	ctx := context.Background()
	agg.Start(ctx)
	defer agg.Stop()

	agg.TriggerFlush()
	time.Sleep(100 * time.Millisecond)
}

func TestAlertAggregator_GetGroup(t *testing.T) {
	config := AggregatorConfig{Strategy: StrategyBySource, MaxGroupSize: 100}
	agg := NewAlertAggregator(config)

	alert := &AlertInstance{
		ID:          "alert-1",
		RuleID:      "rule-1",
		Message:     "Test msg",
		TriggeredAt: time.Now(),
		Source:      "server-1",
	}
	agg.Aggregate(context.Background(), alert)

	group := agg.GetGroup("server-1")
	assert.NotNil(t, group)
	assert.Equal(t, 1, group.Count)

	group = agg.GetGroup("nonexistent")
	assert.Nil(t, group)
}

func TestAlertAggregator_GetAllGroups(t *testing.T) {
	config := AggregatorConfig{Strategy: StrategyBySource, MaxGroupSize: 100}
	agg := NewAlertAggregator(config)

	alert1 := &AlertInstance{ID: "a1", RuleID: "r1", Message: "M1", TriggeredAt: time.Now(), Source: "s1"}
	alert2 := &AlertInstance{ID: "a2", RuleID: "r1", Message: "M2", TriggeredAt: time.Now(), Source: "s2"}
	agg.Aggregate(context.Background(), alert1)
	agg.Aggregate(context.Background(), alert2)

	groups := agg.GetAllGroups()
	assert.Equal(t, 2, len(groups))
}

func TestAlertAggregator_DefaultConfig(t *testing.T) {
	config := DefaultAggregatorConfig()
	assert.Equal(t, StrategyBySource, config.Strategy)
	assert.Equal(t, 5*time.Minute, config.WindowDuration)
	assert.True(t, config.EnableDeduplication)
}

func TestAlertAggregator_ZeroDefaults(t *testing.T) {
	config := AggregatorConfig{}
	agg := NewAlertAggregator(config)
	assert.Equal(t, 5*time.Minute, agg.config.WindowDuration)
	assert.Equal(t, 30*time.Second, agg.config.FlushInterval)
}

func TestAlertGroup_Add(t *testing.T) {
	group := NewAlertGroup(StrategyBySource, "server-1")

	alert1 := &AlertInstance{
		ID:          "a1",
		Severity:    SeverityWarning,
		TriggeredAt: time.Now().Add(-1 * time.Minute),
		Labels:      map[string]string{"env": "prod"},
	}
	alert2 := &AlertInstance{
		ID:          "a2",
		Severity:    SeverityCritical,
		TriggeredAt: time.Now(),
		Labels:      map[string]string{"zone": "us-east"},
	}

	group.Add(alert1)
	group.Add(alert2)

	assert.Equal(t, 2, group.Count)
	assert.Equal(t, SeverityCritical, group.MaxSeverity)
	assert.Equal(t, "prod", group.Labels["env"])
	assert.Equal(t, "us-east", group.Labels["zone"])
}

func TestAlertGroup_GenerateSummary(t *testing.T) {
	agg := NewAlertAggregator(AggregatorConfig{Strategy: StrategyBySource, MaxGroupSize: 100})

	group := NewAlertGroup(StrategyBySource, "server-1")
	summary := agg.generateSummary(group)
	assert.Equal(t, "", summary)

	group.Add(&AlertInstance{ID: "a1", Title: "Single Alert", TriggeredAt: time.Now()})
	summary = agg.generateSummary(group)
	assert.Equal(t, "Single Alert", summary)

	group.Add(&AlertInstance{ID: "a2", Title: "Second Alert", TriggeredAt: time.Now()})
	summary = agg.generateSummary(group)
	assert.Contains(t, summary, "alerts aggregated")
}

func TestSilenceManager_NoID(t *testing.T) {
	sm := NewSilenceManager()
	err := sm.AddSilence(&Silence{})
	assert.Error(t, err)
}

func TestSilenceManager_ExpiredSilence(t *testing.T) {
	sm := NewSilenceManager()
	sm.AddSilence(&Silence{
		ID:        "sil-1",
		Matchers:  map[string]string{"source": "server-1"},
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now().Add(-1 * time.Hour),
	})

	alert := &AlertInstance{Source: "server-1", Labels: map[string]string{"source": "server-1"}}
	assert.False(t, sm.IsSilenced(alert))
}

func TestSilenceManager_FutureSilence(t *testing.T) {
	sm := NewSilenceManager()
	sm.AddSilence(&Silence{
		ID:        "sil-1",
		Matchers:  map[string]string{"source": "server-1"},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(2 * time.Hour),
	})

	alert := &AlertInstance{Source: "server-1", Labels: map[string]string{"source": "server-1"}}
	assert.False(t, sm.IsSilenced(alert))
}

func TestSilenceManager_CleanupExpired(t *testing.T) {
	sm := NewSilenceManager()
	sm.AddSilence(&Silence{
		ID:        "sil-expired",
		Matchers:  map[string]string{},
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now().Add(-1 * time.Hour),
	})
	sm.AddSilence(&Silence{
		ID:        "sil-active",
		Matchers:  map[string]string{},
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
	})

	sm.CleanupExpired()
	assert.Equal(t, 1, len(sm.ListSilences()))
}

func TestSilenceManager_GetSilence(t *testing.T) {
	sm := NewSilenceManager()
	sm.AddSilence(&Silence{ID: "sil-1", Matchers: map[string]string{}})

	silence, exists := sm.GetSilence("sil-1")
	assert.True(t, exists)
	assert.Equal(t, "sil-1", silence.ID)

	_, exists = sm.GetSilence("nonexistent")
	assert.False(t, exists)
}

func TestSuppressionEngine_NoID(t *testing.T) {
	se := NewSuppressionEngine()
	err := se.AddRule(&SuppressionRule{})
	assert.Error(t, err)
}

func TestSuppressionEngine_DisabledRule(t *testing.T) {
	se := NewSuppressionEngine()
	se.AddRule(&SuppressionRule{
		ID:             "supp-1",
		Name:           "Disabled",
		TargetMatchers: map[string]string{"severity": "warning"},
		Enabled:        false,
	})

	alert := &AlertInstance{Severity: SeverityWarning, Labels: map[string]string{"severity": "warning"}}
	suppressed, _ := se.IsSuppressed(alert)
	assert.False(t, suppressed)
}

func TestAlertDeduplicator_Cleanup(t *testing.T) {
	dedup := NewAlertDeduplicator(1 * time.Millisecond)

	alert := &AlertInstance{ID: "a1", RuleID: "r1", Message: "M1", Source: "s1", TriggeredAt: time.Now()}
	dedup.Record(alert)

	time.Sleep(5 * time.Millisecond)
	dedup.Cleanup()

	assert.False(t, dedup.IsDuplicate(alert))
}

func TestAlertDeduplicator_Fingerprint(t *testing.T) {
	alert := &AlertInstance{
		ID:     "a1",
		RuleID: "r1",
		Source: "s1",
		Message: "Test",
		Value:  42.0,
		Labels: map[string]string{"env": "prod"},
	}
	fp := generateFingerprint(alert)
	assert.NotEmpty(t, fp)
}

func TestAlertDeduplicator_ExplicitFingerprint(t *testing.T) {
	dedup := NewAlertDeduplicator(5 * time.Minute)

	alert := &AlertInstance{
		ID:          "a1",
		Fingerprint: "custom-fp",
		TriggeredAt: time.Now(),
	}

	assert.False(t, dedup.IsDuplicate(alert))
	dedup.Record(alert)
	assert.True(t, dedup.IsDuplicate(alert))
}

func TestGenerateLabelsKey(t *testing.T) {
	key := generateLabelsKey(nil)
	assert.Equal(t, "no_labels", key)

	key = generateLabelsKey(map[string]string{"a": "b", "c": "d"})
	assert.NotEmpty(t, key)
}

func TestSilenceChecker(t *testing.T) {
	checker := NewSilenceChecker()

	assert.False(t, checker.IsSilent("alert-1"))

	checker.StartSilence("alert-1", 1*time.Hour, "maintenance")
	assert.True(t, checker.IsSilent("alert-1"))

	checker.EndSilence("alert-1")
	assert.False(t, checker.IsSilent("alert-1"))
}

func TestSilenceChecker_Expired(t *testing.T) {
	checker := NewSilenceChecker()
	checker.StartSilence("alert-1", 1*time.Millisecond, "test")
	time.Sleep(5 * time.Millisecond)
	assert.False(t, checker.IsSilent("alert-1"))
}

func TestSilenceChecker_CleanupExpired(t *testing.T) {
	checker := NewSilenceChecker()
	checker.StartSilence("alert-1", 1*time.Millisecond, "test")
	time.Sleep(5 * time.Millisecond)
	checker.CleanupExpired()
	assert.False(t, checker.IsSilent("alert-1"))
}

func TestNotificationRateLimiter_HourlyLimit(t *testing.T) {
	limiter := NewNotificationRateLimiter()
	limiter.SetLimit("email", 0, 2)

	assert.True(t, limiter.Allow("email"))
	assert.True(t, limiter.Allow("email"))
	assert.False(t, limiter.Allow("email"))
}

func TestAlertNotifier_SendNotification(t *testing.T) {
	notifier := NewAlertNotifier(nil)
	emailChannel := NewEmailChannel(&EmailConfig{SMTPHost: "localhost"})
	notifier.RegisterChannel(emailChannel)

	notification := &Notification{
		ID:       "notif-1",
		AlertID:  "alert-1",
		Channel:  "email",
		Priority: PriorityNormal,
		Status:   StatusPending,
		Subject:  "Test Alert",
		Content:  "This is a test",
	}

	result, err := notifier.SendNotification(context.Background(), notification)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestAlertNotifier_SendNotification_Silenced(t *testing.T) {
	notifier := NewAlertNotifier(nil)
	notifier.silenceChecker.StartSilence("alert-1", 1*time.Hour, "test")

	notification := &Notification{
		ID:      "notif-1",
		AlertID: "alert-1",
		Channel: "email",
	}

	result, err := notifier.SendNotification(context.Background(), notification)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, StatusCancelled, result.Status)
}

func TestAlertNotifier_SendNotification_RateLimited(t *testing.T) {
	notifier := NewAlertNotifier(nil)
	notifier.RegisterChannel(NewEmailChannel(&EmailConfig{}))
	notifier.rateLimiter.SetLimit("email", 1, 100)

	n1 := &Notification{ID: "n1", AlertID: "a1", Channel: "email"}
	n2 := &Notification{ID: "n2", AlertID: "a2", Channel: "email"}

	result1, _ := notifier.SendNotification(context.Background(), n1)
	assert.True(t, result1.Success)

	result2, _ := notifier.SendNotification(context.Background(), n2)
	assert.False(t, result2.Success)
	assert.Equal(t, StatusCancelled, result2.Status)
}

func TestAlertNotifier_SendNotification_ChannelNotFound(t *testing.T) {
	notifier := NewAlertNotifier(nil)

	notification := &Notification{
		ID:      "notif-1",
		AlertID: "alert-1",
		Channel: "nonexistent",
	}

	_, err := notifier.SendNotification(context.Background(), notification)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "channel not found")
}

func TestAlertNotifier_SendNotification_WithTemplate(t *testing.T) {
	notifier := NewAlertNotifier(nil)
	notifier.RegisterChannel(NewEmailChannel(&EmailConfig{}))
	notifier.templateManager.AddTemplate(&NotificationTemplate{
		ID:              "tmpl-1",
		SubjectTemplate: "Alert: {{.alert.Title}}",
		ContentTemplate: "Message: {{.alert.Message}}",
	})

	notification := &Notification{
		ID:           "notif-1",
		AlertID:      "alert-1",
		Channel:      "email",
		TemplateID:   "tmpl-1",
		TemplateData: map[string]interface{}{"alert": &AlertInstance{Title: "CPU Alert", Message: "CPU is high"}},
	}

	result, err := notifier.SendNotification(context.Background(), notification)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestAlertNotifier_SendNotification_TemplateNotFound(t *testing.T) {
	notifier := NewAlertNotifier(nil)
	notifier.RegisterChannel(NewEmailChannel(&EmailConfig{}))

	notification := &Notification{
		ID:         "notif-1",
		AlertID:    "alert-1",
		Channel:    "email",
		TemplateID: "nonexistent",
	}

	_, err := notifier.SendNotification(context.Background(), notification)
	assert.Error(t, err)
}

func TestAlertNotifier_NotifyAlert(t *testing.T) {
	notifier := NewAlertNotifier(nil)
	notifier.RegisterChannel(NewEmailChannel(&EmailConfig{}))
	notifier.RegisterChannel(NewSMSChannel(&SMSConfig{}))

	alert := &AlertInstance{
		ID:        "alert-1",
		Title:     "Test Alert",
		Message:   "Test message",
		Severity:  SeverityWarning,
	}

	recipients := []Recipient{{Email: "admin@example.com", Name: "Admin"}}
	err := notifier.NotifyAlert(context.Background(), alert, []string{"email", "sms"}, recipients)
	assert.NoError(t, err)
}

func TestAlertNotifier_NotifyAlert_MissingChannel(t *testing.T) {
	notifier := NewAlertNotifier(nil)

	alert := &AlertInstance{
		ID:        "alert-1",
		Title:     "Test Alert",
		Message:   "Test message",
		Severity:  SeverityWarning,
	}

	err := notifier.NotifyAlert(context.Background(), alert, []string{"nonexistent"}, nil)
	assert.NoError(t, err)
}

func TestAlertNotifier_SendBatch(t *testing.T) {
	notifier := NewAlertNotifier(nil)
	notifier.RegisterChannel(NewEmailChannel(&EmailConfig{}))

	notifications := []*Notification{
		{ID: "n1", AlertID: "a1", Channel: "email"},
		{ID: "n2", AlertID: "a2", Channel: "email"},
		{ID: "n3", AlertID: "a3", Channel: "nonexistent"},
	}

	results, err := notifier.SendBatch(context.Background(), notifications)
	require.NoError(t, err)
	assert.Equal(t, 3, len(results))
}

func TestAlertNotifier_WithStore(t *testing.T) {
	store := &mockNotificationStore{}
	notifier := NewAlertNotifier(store)
	notifier.RegisterChannel(NewEmailChannel(&EmailConfig{}))

	notification := &Notification{
		ID:      "notif-1",
		AlertID: "alert-1",
		Channel: "email",
	}

	result, err := notifier.SendNotification(context.Background(), notification)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestGetPriorityFromSeverity(t *testing.T) {
	assert.Equal(t, PriorityCritical, getPriorityFromSeverity(SeverityEmergency))
	assert.Equal(t, PriorityHigh, getPriorityFromSeverity(SeverityCritical))
	assert.Equal(t, PriorityNormal, getPriorityFromSeverity(SeverityWarning))
	assert.Equal(t, PriorityLow, getPriorityFromSeverity(SeverityInfo))
	assert.Equal(t, PriorityLow, getPriorityFromSeverity(AlertSeverity("unknown")))
}

func TestGetSeverityLevel(t *testing.T) {
	assert.Equal(t, 1, getSeverityLevel(SeverityInfo))
	assert.Equal(t, 2, getSeverityLevel(SeverityWarning))
	assert.Equal(t, 3, getSeverityLevel(SeverityCritical))
	assert.Equal(t, 4, getSeverityLevel(SeverityEmergency))
	assert.Equal(t, 0, getSeverityLevel(AlertSeverity("unknown")))
}

func TestEscalationEngine_NoID(t *testing.T) {
	ee := NewEscalationEngine()
	err := ee.AddRule(&EscalationRule{})
	assert.Error(t, err)
}

func TestEscalationEngine_DisabledRule(t *testing.T) {
	ee := NewEscalationEngine()
	ee.AddRule(&EscalationRule{
		ID:           "esc-1",
		AlertMatcher: map[string]string{"severity": "critical"},
		Enabled:      false,
	})

	alert := &AlertInstance{Severity: SeverityCritical, Labels: map[string]string{"severity": "critical"}}
	level := ee.CheckEscalation(alert, time.Now().Add(-10*time.Minute))
	assert.Nil(t, level)
}

func TestEscalationEngine_CategoryMatcher(t *testing.T) {
	ee := NewEscalationEngine()
	ee.AddRule(&EscalationRule{
		ID:           "esc-1",
		AlertMatcher: map[string]string{"category": "system_resource"},
		Levels: []EscalationLevel{{Level: 1, After: 5 * time.Minute}},
		Enabled:      true,
	})

	alert := &AlertInstance{Category: CategorySystemResource, Labels: map[string]string{}}
	level := ee.CheckEscalation(alert, time.Now().Add(-10*time.Minute))
	assert.NotNil(t, level)
	assert.Equal(t, 1, level.Level)
}

func TestEscalationEngine_LabelMatcher(t *testing.T) {
	ee := NewEscalationEngine()
	ee.AddRule(&EscalationRule{
		ID:           "esc-1",
		AlertMatcher: map[string]string{"env": "prod"},
		Levels: []EscalationLevel{{Level: 1, After: 5 * time.Minute}},
		Enabled:      true,
	})

	alert := &AlertInstance{Labels: map[string]string{"env": "prod"}}
	level := ee.CheckEscalation(alert, time.Now().Add(-10*time.Minute))
	assert.NotNil(t, level)

	alert2 := &AlertInstance{Labels: map[string]string{"env": "dev"}}
	level2 := ee.CheckEscalation(alert2, time.Now().Add(-10*time.Minute))
	assert.Nil(t, level2)
}

func TestNotificationTemplate_NoID(t *testing.T) {
	mgr := NewNotificationTemplateManager()
	err := mgr.AddTemplate(&NotificationTemplate{})
	assert.Error(t, err)
}

func TestNotificationTemplate_NotFound(t *testing.T) {
	mgr := NewNotificationTemplateManager()
	_, err := mgr.GetTemplate("nonexistent")
	assert.Error(t, err)
}

func TestNotificationTemplate_RenderHTML(t *testing.T) {
	tmpl := &NotificationTemplate{
		ID:           "tmpl-1",
		HTMLTemplate: "<h1>{{.alert.Title}}</h1>",
	}
	html, err := tmpl.RenderHTML(map[string]interface{}{"alert": &AlertInstance{Title: "Test"}})
	require.NoError(t, err)
	assert.Equal(t, "<h1>Test</h1>", html)
}

func TestNotificationTemplate_RenderEmpty(t *testing.T) {
	tmpl := &NotificationTemplate{ID: "tmpl-1"}
	subject, _ := tmpl.RenderSubject(nil)
	assert.Equal(t, "", subject)
	content, _ := tmpl.RenderContent(nil)
	assert.Equal(t, "", content)
	html, _ := tmpl.RenderHTML(nil)
	assert.Equal(t, "", html)
}

func TestRenderTemplate_Invalid(t *testing.T) {
	_, err := renderTemplate("{{.invalid", nil)
	assert.Error(t, err)
}

func TestAlertAPI_RegisterRoutes(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)
}

func TestAlertAPI_QueryAlerts(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	api.handleAlerts(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_CreateAlert(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	body, _ := json.Marshal(AlertInstance{
		ID:          "alert-1",
		RuleID:      "rule-1",
		Title:       "Test",
		Message:     "Test msg",
		TriggeredAt: time.Now(),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleAlerts(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_CreateAlert_InvalidBody(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()
	api.handleAlerts(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlertAPI_MethodNotAllowed(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/alerts", nil)
	w := httptest.NewRecorder()
	api.handleAlerts(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestAlertAPI_GetAlert(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/alert-1", nil)
	w := httptest.NewRecorder()
	api.handleAlertDetail(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_DeleteAlert(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/alerts/alert-1", nil)
	w := httptest.NewRecorder()
	api.handleAlertDetail(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_Acknowledge(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	body, _ := json.Marshal(map[string]string{"alert_id": "alert-1", "by": "admin"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/acknowledge", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleAcknowledge(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_Acknowledge_MissingID(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	body, _ := json.Marshal(map[string]string{"by": "admin"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/acknowledge", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleAcknowledge(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlertAPI_Clear(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	body, _ := json.Marshal(map[string]string{"alert_id": "alert-1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/clear", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleClear(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_Statistics(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/statistics", nil)
	w := httptest.NewRecorder()
	api.handleStatistics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_History(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/history?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	api.handleHistory(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_Rules(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/rules", nil)
	w := httptest.NewRecorder()
	api.handleRules(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_CreateRule(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	body, _ := json.Marshal(AlertRule{ID: "rule-1", Name: "Test"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/rules", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleRules(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_RuleDetail(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/rules/rule-1", nil)
	w := httptest.NewRecorder()
	api.handleRuleDetail(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAlertAPI_AggregatorStats(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/aggregator/stats", nil)
	w := httptest.NewRecorder()
	api.handleAggregatorStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_AggregatorFlush(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/aggregator/flush", nil)
	w := httptest.NewRecorder()
	api.handleAggregatorFlush(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_Silences(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/silences", nil)
	w := httptest.NewRecorder()
	api.handleSilences(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_CreateSilence(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	body, _ := json.Marshal(Silence{ID: "sil-1", Matchers: map[string]string{}, StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/silences", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleSilences(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_SilenceDetail(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/silences/sil-1", nil)
	w := httptest.NewRecorder()
	api.handleSilenceDetail(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAlertAPI_DeleteSilence(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/alerts/silences/sil-1", nil)
	w := httptest.NewRecorder()
	api.handleSilenceDetail(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_Suppressions(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/suppressions", nil)
	w := httptest.NewRecorder()
	api.handleSuppressions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_Templates(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/templates", nil)
	w := httptest.NewRecorder()
	api.handleTemplates(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_Escalations(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}

	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/escalations", nil)
	w := httptest.NewRecorder()
	api.handleEscalations(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSystemResourceAlertRules(t *testing.T) {
	rules := NewSystemResourceAlertRules()
	cpuRule := rules.CPUUsageRule(80.0, 5*time.Minute)
	assert.Equal(t, CategorySystemResource, cpuRule.Category)
	assert.Equal(t, SeverityWarning, cpuRule.Severity)

	memRule := rules.MemoryUsageRule(85.0, 5*time.Minute)
	assert.Equal(t, CategorySystemResource, memRule.Category)

	diskRule := rules.DiskUsageRule(90.0, 5*time.Minute)
	assert.Equal(t, CategorySystemResource, diskRule.Category)
}

func TestServiceHealthAlertRules(t *testing.T) {
	rules := NewServiceHealthAlertRules()

	downRule := rules.ServiceDownRule("api-server")
	assert.Equal(t, CategoryServiceHealth, downRule.Category)
	assert.Equal(t, SeverityCritical, downRule.Severity)

	respRule := rules.ServiceResponseTimeRule("api-server", 500)
	assert.Equal(t, CategoryPerformance, respRule.Category)

	errRule := rules.ServiceErrorRateRule("api-server", 5.0)
	assert.Equal(t, CategoryPerformance, errRule.Category)
}

func TestBusinessAlertRules(t *testing.T) {
	rules := NewBusinessAlertRules()

	offlineRule := rules.StationOfflineRule()
	assert.Equal(t, CategoryBusiness, offlineRule.Category)
	assert.Equal(t, SeverityCritical, offlineRule.Severity)

	anomalyRule := rules.PowerGenerationAnomalyRule(50.0)
	assert.Equal(t, CategoryBusiness, anomalyRule.Category)

	faultRule := rules.DeviceFaultRule()
	assert.Equal(t, CategoryBusiness, faultRule.Category)
}

func TestEmailChannel(t *testing.T) {
	ch := NewEmailChannel(&EmailConfig{SMTPHost: "localhost"})
	assert.Equal(t, "email", ch.Name())

	result, err := ch.Send(context.Background(), &Notification{ID: "n1"})
	require.NoError(t, err)
	assert.True(t, result.Success)

	results, err := ch.SendBatch(context.Background(), []*Notification{{ID: "n1"}})
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))

	err = ch.HealthCheck(context.Background())
	assert.NoError(t, err)

	err = ch.Close()
	assert.NoError(t, err)
}

func TestSMSChannel(t *testing.T) {
	ch := NewSMSChannel(&SMSConfig{Provider: "aliyun"})
	assert.Equal(t, "sms", ch.Name())

	result, err := ch.Send(context.Background(), &Notification{ID: "n1"})
	require.NoError(t, err)
	assert.True(t, result.Success)

	err = ch.HealthCheck(context.Background())
	assert.NoError(t, err)

	err = ch.Close()
	assert.NoError(t, err)
}

func TestDingTalkChannel(t *testing.T) {
	ch := NewDingTalkChannel(&DingTalkConfig{WebhookURL: "https://example.com"})
	assert.Equal(t, "dingtalk", ch.Name())

	result, err := ch.Send(context.Background(), &Notification{ID: "n1"})
	require.NoError(t, err)
	assert.True(t, result.Success)

	err = ch.HealthCheck(context.Background())
	assert.NoError(t, err)

	err = ch.Close()
	assert.NoError(t, err)
}

func TestWeChatChannel(t *testing.T) {
	ch := NewWeChatChannel(&WeChatConfig{CorpID: "test"})
	assert.Equal(t, "wechat", ch.Name())

	result, err := ch.Send(context.Background(), &Notification{ID: "n1"})
	require.NoError(t, err)
	assert.True(t, result.Success)

	err = ch.HealthCheck(context.Background())
	assert.NoError(t, err)

	err = ch.Close()
	assert.NoError(t, err)
}

func TestGenerateAlertID(t *testing.T) {
	id1 := generateAlertID()
	id2 := generateAlertID()
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "alert-")
}

type multiValueProvider struct{}

func (m *multiValueProvider) Query(ctx context.Context, metricName string, labels map[string]string, start, end time.Time) ([]MetricData, error) {
	return []MetricData{
		{Name: metricName, Value: 100.0, Timestamp: time.Now(), Labels: labels},
		{Name: metricName, Value: 200.0, Timestamp: time.Now(), Labels: labels},
	}, nil
}

func (m *multiValueProvider) QueryLatest(ctx context.Context, metricName string, labels map[string]string) (*MetricData, error) {
	return &MetricData{Name: metricName, Value: 100.0, Timestamp: time.Now(), Labels: labels}, nil
}

func (m *multiValueProvider) QueryRange(ctx context.Context, metricName string, labels map[string]string, duration time.Duration) ([]MetricData, error) {
	return []MetricData{
		{Name: metricName, Value: 100.0, Timestamp: time.Now().Add(-1 * time.Minute), Labels: labels},
		{Name: metricName, Value: 200.0, Timestamp: time.Now(), Labels: labels},
	}, nil
}

type rateProvider struct{}

func (r *rateProvider) Query(ctx context.Context, metricName string, labels map[string]string, start, end time.Time) ([]MetricData, error) {
	return []MetricData{
		{Name: metricName, Value: 100.0, Timestamp: time.Now().Add(-1 * time.Minute), Labels: labels},
		{Name: metricName, Value: 200.0, Timestamp: time.Now(), Labels: labels},
	}, nil
}

func (r *rateProvider) QueryLatest(ctx context.Context, metricName string, labels map[string]string) (*MetricData, error) {
	return &MetricData{Name: metricName, Value: 100.0, Timestamp: time.Now(), Labels: labels}, nil
}

func (r *rateProvider) QueryRange(ctx context.Context, metricName string, labels map[string]string, duration time.Duration) ([]MetricData, error) {
	return []MetricData{
		{Name: metricName, Value: 100.0, Timestamp: time.Now().Add(-60 * time.Second), Labels: labels},
		{Name: metricName, Value: 200.0, Timestamp: time.Now(), Labels: labels},
	}, nil
}

type mockAlertStore struct{}

func (m *mockAlertStore) Save(ctx context.Context, alert *AlertInstance) error                                          { return nil }
func (m *mockAlertStore) Update(ctx context.Context, alert *AlertInstance) error                                        { return nil }
func (m *mockAlertStore) Get(ctx context.Context, id string) (*AlertInstance, error)                                    { return &AlertInstance{ID: id}, nil }
func (m *mockAlertStore) Query(ctx context.Context, query *AlertQuery) ([]*AlertInstance, int64, error)                 { return []*AlertInstance{}, 0, nil }
func (m *mockAlertStore) Acknowledge(ctx context.Context, id string, by string) error                                   { return nil }
func (m *mockAlertStore) Clear(ctx context.Context, id string) error                                                    { return nil }
func (m *mockAlertStore) GetStatistics(ctx context.Context, query *StatisticsQuery) (*AlertStatistics, error)            { return &AlertStatistics{}, nil }
func (m *mockAlertStore) GetHistory(ctx context.Context, query *HistoryQuery) ([]*AlertHistory, int64, error)           { return []*AlertHistory{}, 0, nil }

type mockNotificationStore struct{}

func (m *mockNotificationStore) Save(ctx context.Context, notification *Notification) error              { return nil }
func (m *mockNotificationStore) Update(ctx context.Context, notification *Notification) error            { return nil }
func (m *mockNotificationStore) Get(ctx context.Context, id string) (*Notification, error)               { return nil, nil }
func (m *mockNotificationStore) GetPending(ctx context.Context, limit int) ([]*Notification, error)      { return nil, nil }
func (m *mockNotificationStore) GetByAlertID(ctx context.Context, alertID string) ([]*Notification, error) { return nil, nil }

func TestAlertAPI_UpdateRule(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	rm.AddRule(&AlertRule{ID: "rule-1", Name: "Original"})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	body, _ := json.Marshal(AlertRule{ID: "rule-1", Name: "Updated"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/alerts/rules/rule-1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleRuleDetail(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_DeleteRule(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	rm.AddRule(&AlertRule{ID: "rule-1", Name: "Test"})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/alerts/rules/rule-1", nil)
	w := httptest.NewRecorder()
	api.handleRuleDetail(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_SuppressionDetail(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/suppressions/supp-1", nil)
	w := httptest.NewRecorder()
	api.handleSuppressionDetail(w, req)
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestAlertAPI_DeleteSuppression(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/alerts/suppressions/supp-1", nil)
	w := httptest.NewRecorder()
	api.handleSuppressionDetail(w, req)
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestAlertAPI_TemplateDetail(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/templates/tmpl-1", nil)
	w := httptest.NewRecorder()
	api.handleTemplateDetail(w, req)
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestAlertAPI_DeleteTemplate(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/alerts/templates/tmpl-1", nil)
	w := httptest.NewRecorder()
	api.handleTemplateDetail(w, req)
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestAlertAPI_EscalationDetail(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/escalations/esc-1", nil)
	w := httptest.NewRecorder()
	api.handleEscalationDetail(w, req)
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestAlertAPI_DeleteEscalation(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/alerts/escalations/esc-1", nil)
	w := httptest.NewRecorder()
	api.handleEscalationDetail(w, req)
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestAlertService_EvaluateRules(t *testing.T) {
	config := AlertServiceConfig{
		MetricProvider: &MockMetricProvider{},
		AggregatorConfig: AggregatorConfig{Strategy: StrategyBySource, WindowDuration: 5 * time.Minute, FlushInterval: 30 * time.Second, MaxGroupSize: 100, EnableDeduplication: true, DedupWindow: 5 * time.Minute, EnableAutoFlush: true},
	}
	svc := NewAlertService(config)
	svc.AddRule(&AlertRule{
		ID:      "rule-1",
		Name:    "Test",
		Enabled: true,
		Condition: AlertCondition{
			MetricName: "test_metric",
			Operator:   OpGT,
			Threshold:  50,
		},
		NotifyChannels: []string{"email"},
	})
	svc.RegisterNotificationChannel(NewEmailChannel(&EmailConfig{}))

	svc.Start()
	time.Sleep(200 * time.Millisecond)
	svc.Stop()
}

func TestAlertAggregator_ResetWindow(t *testing.T) {
	config := AggregatorConfig{
		Strategy:        StrategyByTimeWindow,
		WindowDuration:  1 * time.Hour,
		MaxGroupSize:    100,
		EnableAutoFlush: true,
		FlushInterval:   30 * time.Second,
	}
	agg := NewAlertAggregator(config)
	ctx := context.Background()
	agg.Start(ctx)
	defer agg.Stop()

	alert := &AlertInstance{
		ID:          "alert-1",
		RuleID:      "rule-1",
		Message:     "Test msg",
		TriggeredAt: time.Now(),
		Source:      "server-1",
	}
	agg.Aggregate(ctx, alert)
}

func TestEmailChannel_SendBatch(t *testing.T) {
	ch := NewEmailChannel(&EmailConfig{SMTPHost: "localhost"})
	notifications := []*Notification{
		{ID: "n1", AlertID: "a1"},
		{ID: "n2", AlertID: "a2"},
	}
	results, err := ch.SendBatch(context.Background(), notifications)
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestSMSChannel_SendBatch(t *testing.T) {
	ch := NewSMSChannel(&SMSConfig{Provider: "aliyun"})
	notifications := []*Notification{{ID: "n1"}, {ID: "n2"}}
	results, err := ch.SendBatch(context.Background(), notifications)
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestDingTalkChannel_SendBatch(t *testing.T) {
	ch := NewDingTalkChannel(&DingTalkConfig{WebhookURL: "https://example.com"})
	notifications := []*Notification{{ID: "n1"}, {ID: "n2"}}
	results, err := ch.SendBatch(context.Background(), notifications)
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestWeChatChannel_SendBatch(t *testing.T) {
	ch := NewWeChatChannel(&WeChatConfig{CorpID: "test"})
	notifications := []*Notification{{ID: "n1"}, {ID: "n2"}}
	results, err := ch.SendBatch(context.Background(), notifications)
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestAlertAPI_CreateSuppression(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	body, _ := json.Marshal(SuppressionRule{ID: "supp-1", Name: "Test", TargetMatchers: map[string]string{}, Enabled: true})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/suppressions", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleSuppressions(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_CreateTemplate(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	body, _ := json.Marshal(NotificationTemplate{ID: "tmpl-1", Name: "Test"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/templates", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleTemplates(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_CreateEscalation(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	body, _ := json.Marshal(EscalationRule{ID: "esc-1", Name: "Test", AlertMatcher: map[string]string{}, Enabled: true})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/escalations", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleEscalations(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAlertAPI_Acknowledge_InvalidBody(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/acknowledge", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()
	api.handleAcknowledge(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlertAPI_Clear_InvalidBody(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/clear", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()
	api.handleClear(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlertAPI_Clear_MissingID(t *testing.T) {
	rm := NewRuleManager(&MockMetricProvider{})
	agg := NewAlertAggregator(DefaultAggregatorConfig())
	notifier := NewAlertNotifier(nil)
	store := &mockAlertStore{}
	api := NewAlertAPI(rm, agg, notifier, store)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/clear", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleClear(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
