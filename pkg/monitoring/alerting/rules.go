package alerting

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AlertRule 告警规则
type AlertRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    AlertCategory     `json:"category"`
	Severity    AlertSeverity     `json:"severity"`
	Enabled     bool              `json:"enabled"`

	// 触发条件
	Condition   AlertCondition    `json:"condition"`

	// 通知配置
	NotifyChannels []string       `json:"notify_channels"`
	NotifyTemplate string         `json:"notify_template"`

	// 抑制与静默
	SuppressionRules []string     `json:"suppression_rules"`
	SilenceDuration  time.Duration `json:"silence_duration"`

	// 元数据
	Tags        map[string]string `json:"tags"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CreatedBy   string            `json:"created_by"`
}

// AlertCategory 告警类别
type AlertCategory string

const (
	CategorySystemResource AlertCategory = "system_resource" // 系统资源告警
	CategoryServiceHealth  AlertCategory = "service_health"  // 服务可用性告警
	CategoryPerformance    AlertCategory = "performance"     // 性能告警
	CategoryBusiness       AlertCategory = "business"        // 业务告警
	CategorySecurity       AlertCategory = "security"        // 安全告警
)

// AlertSeverity 告警严重程度
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
	SeverityEmergency AlertSeverity = "emergency"
)

// AlertCondition 告警条件
type AlertCondition struct {
	// 指标名称
	MetricName string `json:"metric_name"`

	// 比较操作符
	Operator ComparisonOperator `json:"operator"`

	// 阈值
	Threshold float64 `json:"threshold"`

	// 持续时间
	Duration time.Duration `json:"duration"`

	// 聚合函数
	Aggregation AggregationFunc `json:"aggregation"`

	// 聚合窗口
	AggregationWindow time.Duration `json:"aggregation_window"`

	// 标签过滤
	LabelFilters map[string]string `json:"label_filters"`

	// 复合条件
	Conditions []AlertCondition `json:"conditions,omitempty"`
	LogicalOperator LogicalOperator `json:"logical_operator,omitempty"`
}

// ComparisonOperator 比较操作符
type ComparisonOperator string

const (
	OpEqual    ComparisonOperator = "=="
	OpNotEqual ComparisonOperator = "!="
	OpGT       ComparisonOperator = ">"
	OpGTE      ComparisonOperator = ">="
	OpLT       ComparisonOperator = "<"
	OpLTE      ComparisonOperator = "<="
)

// AggregationFunc 聚合函数
type AggregationFunc string

const (
	AggAvg   AggregationFunc = "avg"
	AggMax   AggregationFunc = "max"
	AggMin   AggregationFunc = "min"
	AggSum   AggregationFunc = "sum"
	AggCount AggregationFunc = "count"
	AggRate  AggregationFunc = "rate"
)

// LogicalOperator 逻辑操作符
type LogicalOperator string

const (
	LogicalAnd LogicalOperator = "and"
	LogicalOr  LogicalOperator = "or"
)

// MetricData 指标数据
type MetricData struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	Labels    map[string]string `json:"labels"`
}

// RuleEvaluationResult 规则评估结果
type RuleEvaluationResult struct {
	RuleID      string        `json:"rule_id"`
	RuleName    string        `json:"rule_name"`
	Triggered   bool          `json:"triggered"`
	Value       float64       `json:"value"`
	Threshold   float64       `json:"threshold"`
	Duration    time.Duration `json:"duration"`
	Timestamp   time.Time     `json:"timestamp"`
	Labels      map[string]string `json:"labels"`
	Error       error         `json:"error,omitempty"`
}

// MetricProvider 指标提供者接口
type MetricProvider interface {
	// Query 查询指标
	Query(ctx context.Context, metricName string, labels map[string]string, start, end time.Time) ([]MetricData, error)

	// QueryLatest 查询最新值
	QueryLatest(ctx context.Context, metricName string, labels map[string]string) (*MetricData, error)

	// QueryRange 查询时间范围数据
	QueryRange(ctx context.Context, metricName string, labels map[string]string, duration time.Duration) ([]MetricData, error)
}

// RuleManager 规则管理器
type RuleManager struct {
	mu            sync.RWMutex
	rules         map[string]*AlertRule
	metricProvider MetricProvider
	evaluators    map[string]*RuleEvaluator
}

// NewRuleManager 创建规则管理器
func NewRuleManager(metricProvider MetricProvider) *RuleManager {
	return &RuleManager{
		rules:          make(map[string]*AlertRule),
		metricProvider: metricProvider,
		evaluators:     make(map[string]*RuleEvaluator),
	}
}

// AddRule 添加规则
func (rm *RuleManager) AddRule(rule *AlertRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule id is required")
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	rule.UpdatedAt = time.Now()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}

	rm.rules[rule.ID] = rule
	rm.evaluators[rule.ID] = NewRuleEvaluator(rule, rm.metricProvider)

	return nil
}

// RemoveRule 移除规则
func (rm *RuleManager) RemoveRule(ruleID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.rules, ruleID)
	delete(rm.evaluators, ruleID)
}

// GetRule 获取规则
func (rm *RuleManager) GetRule(ruleID string) (*AlertRule, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rule, exists := rm.rules[ruleID]
	return rule, exists
}

// GetAllRules 获取所有规则
func (rm *RuleManager) GetAllRules() []*AlertRule {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rules := make([]*AlertRule, 0, len(rm.rules))
	for _, rule := range rm.rules {
		rules = append(rules, rule)
	}
	return rules
}

// GetEnabledRules 获取启用的规则
func (rm *RuleManager) GetEnabledRules() []*AlertRule {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rules := make([]*AlertRule, 0)
	for _, rule := range rm.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	return rules
}

// EvaluateRule 评估规则
func (rm *RuleManager) EvaluateRule(ctx context.Context, ruleID string) (*RuleEvaluationResult, error) {
	rm.mu.RLock()
	evaluator, exists := rm.evaluators[ruleID]
	rm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("rule not found: %s", ruleID)
	}

	return evaluator.Evaluate(ctx)
}

// EvaluateAll 评估所有启用的规则
func (rm *RuleManager) EvaluateAll(ctx context.Context) ([]*RuleEvaluationResult, error) {
	rules := rm.GetEnabledRules()
	results := make([]*RuleEvaluationResult, 0, len(rules))

	for _, rule := range rules {
		result, err := rm.EvaluateRule(ctx, rule.ID)
		if err != nil {
			results = append(results, &RuleEvaluationResult{
				RuleID: rule.ID,
				RuleName: rule.Name,
				Error: err,
				Timestamp: time.Now(),
			})
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// RuleEvaluator 规则评估器
type RuleEvaluator struct {
	rule           *AlertRule
	metricProvider MetricProvider
	state          *EvaluationState
}

// EvaluationState 评估状态
type EvaluationState struct {
	mu               sync.RWMutex
	lastEvaluation   time.Time
	triggeredAt      *time.Time
	lastValue        float64
	consecutiveCount int
}

// NewRuleEvaluator 创建规则评估器
func NewRuleEvaluator(rule *AlertRule, metricProvider MetricProvider) *RuleEvaluator {
	return &RuleEvaluator{
		rule:           rule,
		metricProvider: metricProvider,
		state: &EvaluationState{},
	}
}

// Evaluate 评估规则
func (re *RuleEvaluator) Evaluate(ctx context.Context) (*RuleEvaluationResult, error) {
	startTime := time.Now()

	// 检查规则是否启用
	if !re.rule.Enabled {
		return &RuleEvaluationResult{
			RuleID:    re.rule.ID,
			RuleName:  re.rule.Name,
			Triggered: false,
			Timestamp: time.Now(),
			Duration:  time.Since(startTime),
		}, nil
	}

	// 获取指标数据
	metricData, err := re.getMetricData(ctx)
	if err != nil {
		return &RuleEvaluationResult{
			RuleID:    re.rule.ID,
			RuleName:  re.rule.Name,
			Error:     fmt.Errorf("failed to get metric data: %w", err),
			Timestamp: time.Now(),
			Duration:  time.Since(startTime),
		}, nil
	}

	// 计算聚合值
	value, err := re.calculateAggregation(metricData)
	if err != nil {
		return &RuleEvaluationResult{
			RuleID:    re.rule.ID,
			RuleName:  re.rule.Name,
			Error:     fmt.Errorf("failed to calculate aggregation: %w", err),
			Timestamp: time.Now(),
			Duration:  time.Since(startTime),
		}, nil
	}

	// 评估条件
	triggered := re.evaluateCondition(value)

	// 更新状态
	re.updateState(value, triggered)

	// 检查持续时间要求
	if triggered && re.rule.Condition.Duration > 0 {
		triggered = re.checkDuration()
	}

	return &RuleEvaluationResult{
		RuleID:    re.rule.ID,
		RuleName:  re.rule.Name,
		Triggered: triggered,
		Value:     value,
		Threshold: re.rule.Condition.Threshold,
		Duration:  time.Since(startTime),
		Timestamp: time.Now(),
		Labels:    re.rule.Condition.LabelFilters,
	}, nil
}

// getMetricData 获取指标数据
func (re *RuleEvaluator) getMetricData(ctx context.Context) ([]MetricData, error) {
	condition := &re.rule.Condition

	// 如果有聚合窗口，查询时间范围数据
	if condition.AggregationWindow > 0 {
		return re.metricProvider.QueryRange(ctx, condition.MetricName, condition.LabelFilters, condition.AggregationWindow)
	}

	// 否则查询最新值
	data, err := re.metricProvider.QueryLatest(ctx, condition.MetricName, condition.LabelFilters)
	if err != nil {
		return nil, err
	}

	return []MetricData{*data}, nil
}

// calculateAggregation 计算聚合值
func (re *RuleEvaluator) calculateAggregation(data []MetricData) (float64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("no data available")
	}

	switch re.rule.Condition.Aggregation {
	case AggAvg:
		sum := 0.0
		for _, d := range data {
			sum += d.Value
		}
		return sum / float64(len(data)), nil

	case AggMax:
		max := data[0].Value
		for _, d := range data {
			if d.Value > max {
				max = d.Value
			}
		}
		return max, nil

	case AggMin:
		min := data[0].Value
		for _, d := range data {
			if d.Value < min {
				min = d.Value
			}
		}
		return min, nil

	case AggSum:
		sum := 0.0
		for _, d := range data {
			sum += d.Value
		}
		return sum, nil

	case AggCount:
		return float64(len(data)), nil

	case AggRate:
		if len(data) < 2 {
			return 0, fmt.Errorf("insufficient data for rate calculation")
		}
		// 计算变化率
		first := data[0]
		last := data[len(data)-1]
		timeDiff := last.Timestamp.Sub(first.Timestamp).Seconds()
		if timeDiff == 0 {
			return 0, fmt.Errorf("time difference is zero")
		}
		return (last.Value - first.Value) / timeDiff, nil

	default:
		// 默认返回最新值
		return data[len(data)-1].Value, nil
	}
}

// evaluateCondition 评估条件
func (re *RuleEvaluator) evaluateCondition(value float64) bool {
	threshold := re.rule.Condition.Threshold

	switch re.rule.Condition.Operator {
	case OpEqual:
		return value == threshold
	case OpNotEqual:
		return value != threshold
	case OpGT:
		return value > threshold
	case OpGTE:
		return value >= threshold
	case OpLT:
		return value < threshold
	case OpLTE:
		return value <= threshold
	default:
		return false
	}
}

// updateState 更新状态
func (re *RuleEvaluator) updateState(value float64, triggered bool) {
	re.state.mu.Lock()
	defer re.state.mu.Unlock()

	re.state.lastEvaluation = time.Now()
	re.state.lastValue = value

	if triggered {
		if re.state.triggeredAt == nil {
			now := time.Now()
			re.state.triggeredAt = &now
			re.state.consecutiveCount = 1
		} else {
			re.state.consecutiveCount++
		}
	} else {
		re.state.triggeredAt = nil
		re.state.consecutiveCount = 0
	}
}

// checkDuration 检查持续时间
func (re *RuleEvaluator) checkDuration() bool {
	re.state.mu.RLock()
	defer re.state.mu.RUnlock()

	if re.state.triggeredAt == nil {
		return false
	}

	return time.Since(*re.state.triggeredAt) >= re.rule.Condition.Duration
}

// SystemResourceAlertRules 系统资源告警规则预设
type SystemResourceAlertRules struct{}

// NewSystemResourceAlertRules 创建系统资源告警规则预设
func NewSystemResourceAlertRules() *SystemResourceAlertRules {
	return &SystemResourceAlertRules{}
}

// CPUUsageRule CPU使用率告警规则
func (s *SystemResourceAlertRules) CPUUsageRule(threshold float64, duration time.Duration) *AlertRule {
	return &AlertRule{
		ID:          "system_cpu_usage",
		Name:        "CPU使用率告警",
		Description: fmt.Sprintf("CPU使用率超过%.1f%%持续%s", threshold, duration),
		Category:    CategorySystemResource,
		Severity:    SeverityWarning,
		Enabled:     true,
		Condition: AlertCondition{
			MetricName:       "system_cpu_usage_percent",
			Operator:         OpGT,
			Threshold:        threshold,
			Duration:         duration,
			Aggregation:      AggAvg,
			AggregationWindow: 1 * time.Minute,
		},
		NotifyChannels: []string{"email", "sms"},
		SilenceDuration: 5 * time.Minute,
		Tags: map[string]string{
			"resource": "cpu",
			"type":     "usage",
		},
	}
}

// MemoryUsageRule 内存使用率告警规则
func (s *SystemResourceAlertRules) MemoryUsageRule(threshold float64, duration time.Duration) *AlertRule {
	return &AlertRule{
		ID:          "system_memory_usage",
		Name:        "内存使用率告警",
		Description: fmt.Sprintf("内存使用率超过%.1f%%持续%s", threshold, duration),
		Category:    CategorySystemResource,
		Severity:    SeverityWarning,
		Enabled:     true,
		Condition: AlertCondition{
			MetricName:       "system_memory_usage_percent",
			Operator:         OpGT,
			Threshold:        threshold,
			Duration:         duration,
			Aggregation:      AggAvg,
			AggregationWindow: 1 * time.Minute,
		},
		NotifyChannels: []string{"email", "sms"},
		SilenceDuration: 5 * time.Minute,
		Tags: map[string]string{
			"resource": "memory",
			"type":     "usage",
		},
	}
}

// DiskUsageRule 磁盘使用率告警规则
func (s *SystemResourceAlertRules) DiskUsageRule(threshold float64, duration time.Duration) *AlertRule {
	return &AlertRule{
		ID:          "system_disk_usage",
		Name:        "磁盘使用率告警",
		Description: fmt.Sprintf("磁盘使用率超过%.1f%%持续%s", threshold, duration),
		Category:    CategorySystemResource,
		Severity:    SeverityWarning,
		Enabled:     true,
		Condition: AlertCondition{
			MetricName:       "system_disk_usage_percent",
			Operator:         OpGT,
			Threshold:        threshold,
			Duration:         duration,
			Aggregation:      AggAvg,
			AggregationWindow: 1 * time.Minute,
		},
		NotifyChannels: []string{"email", "sms"},
		SilenceDuration: 10 * time.Minute,
		Tags: map[string]string{
			"resource": "disk",
			"type":     "usage",
		},
	}
}

// ServiceHealthAlertRules 服务可用性告警规则预设
type ServiceHealthAlertRules struct{}

// NewServiceHealthAlertRules 创建服务可用性告警规则预设
func NewServiceHealthAlertRules() *ServiceHealthAlertRules {
	return &ServiceHealthAlertRules{}
}

// ServiceDownRule 服务宕机告警规则
func (s *ServiceHealthAlertRules) ServiceDownRule(serviceName string) *AlertRule {
	return &AlertRule{
		ID:          fmt.Sprintf("service_%s_down", serviceName),
		Name:        fmt.Sprintf("%s服务宕机告警", serviceName),
		Description: fmt.Sprintf("%s服务不可用", serviceName),
		Category:    CategoryServiceHealth,
		Severity:    SeverityCritical,
		Enabled:     true,
		Condition: AlertCondition{
			MetricName:  "service_up",
			Operator:    OpEqual,
			Threshold:   0,
			Duration:    30 * time.Second,
			LabelFilters: map[string]string{
				"service": serviceName,
			},
		},
		NotifyChannels: []string{"email", "sms", "dingtalk"},
		SilenceDuration: 1 * time.Minute,
		Tags: map[string]string{
			"service": serviceName,
			"type":    "availability",
		},
	}
}

// ServiceResponseTimeRule 服务响应时间告警规则
func (s *ServiceHealthAlertRules) ServiceResponseTimeRule(serviceName string, threshold float64) *AlertRule {
	return &AlertRule{
		ID:          fmt.Sprintf("service_%s_response_time", serviceName),
		Name:        fmt.Sprintf("%s服务响应时间告警", serviceName),
		Description: fmt.Sprintf("%s服务响应时间超过%.0fms", serviceName, threshold),
		Category:    CategoryPerformance,
		Severity:    SeverityWarning,
		Enabled:     true,
		Condition: AlertCondition{
			MetricName:       "service_response_time_ms",
			Operator:         OpGT,
			Threshold:        threshold,
			Duration:         1 * time.Minute,
			Aggregation:      AggAvg,
			AggregationWindow: 30 * time.Second,
			LabelFilters: map[string]string{
				"service": serviceName,
			},
		},
		NotifyChannels: []string{"email"},
		SilenceDuration: 5 * time.Minute,
		Tags: map[string]string{
			"service": serviceName,
			"type":    "performance",
		},
	}
}

// ServiceErrorRateRule 服务错误率告警规则
func (s *ServiceHealthAlertRules) ServiceErrorRateRule(serviceName string, threshold float64) *AlertRule {
	return &AlertRule{
		ID:          fmt.Sprintf("service_%s_error_rate", serviceName),
		Name:        fmt.Sprintf("%s服务错误率告警", serviceName),
		Description: fmt.Sprintf("%s服务错误率超过%.1f%%", serviceName, threshold),
		Category:    CategoryPerformance,
		Severity:    SeverityCritical,
		Enabled:     true,
		Condition: AlertCondition{
			MetricName:       "service_error_rate_percent",
			Operator:         OpGT,
			Threshold:        threshold,
			Duration:         1 * time.Minute,
			Aggregation:      AggAvg,
			AggregationWindow: 30 * time.Second,
			LabelFilters: map[string]string{
				"service": serviceName,
			},
		},
		NotifyChannels: []string{"email", "sms", "dingtalk"},
		SilenceDuration: 3 * time.Minute,
		Tags: map[string]string{
			"service": serviceName,
			"type":    "error",
		},
	}
}

// BusinessAlertRules 业务告警规则预设
type BusinessAlertRules struct{}

// NewBusinessAlertRules 创建业务告警规则预设
func NewBusinessAlertRules() *BusinessAlertRules {
	return &BusinessAlertRules{}
}

// StationOfflineRule 电站离线告警规则
func (b *BusinessAlertRules) StationOfflineRule() *AlertRule {
	return &AlertRule{
		ID:          "business_station_offline",
		Name:        "电站离线告警",
		Description: "电站设备离线",
		Category:    CategoryBusiness,
		Severity:    SeverityCritical,
		Enabled:     true,
		Condition: AlertCondition{
			MetricName:  "station_online_count",
			Operator:    OpLT,
			Threshold:   1,
			Duration:    5 * time.Minute,
		},
		NotifyChannels: []string{"email", "sms", "wechat"},
		SilenceDuration: 10 * time.Minute,
		Tags: map[string]string{
			"type": "station",
		},
	}
}

// PowerGenerationAnomalyRule 发电量异常告警规则
func (b *BusinessAlertRules) PowerGenerationAnomalyRule(threshold float64) *AlertRule {
	return &AlertRule{
		ID:          "business_power_generation_anomaly",
		Name:        "发电量异常告警",
		Description: fmt.Sprintf("发电量低于预期%.1f%%", threshold),
		Category:    CategoryBusiness,
		Severity:    SeverityWarning,
		Enabled:     true,
		Condition: AlertCondition{
			MetricName:       "power_generation_percent",
			Operator:         OpLT,
			Threshold:        threshold,
			Duration:         30 * time.Minute,
			Aggregation:      AggAvg,
			AggregationWindow: 5 * time.Minute,
		},
		NotifyChannels: []string{"email"},
		SilenceDuration: 30 * time.Minute,
		Tags: map[string]string{
			"type": "power",
		},
	}
}

// DeviceFaultRule 设备故障告警规则
func (b *BusinessAlertRules) DeviceFaultRule() *AlertRule {
	return &AlertRule{
		ID:          "business_device_fault",
		Name:        "设备故障告警",
		Description: "设备发生故障",
		Category:    CategoryBusiness,
		Severity:    SeverityCritical,
		Enabled:     true,
		Condition: AlertCondition{
			MetricName:  "device_fault_count",
			Operator:    OpGT,
			Threshold:   0,
			Duration:    1 * time.Minute,
		},
		NotifyChannels: []string{"email", "sms", "dingtalk"},
		SilenceDuration: 5 * time.Minute,
		Tags: map[string]string{
			"type": "device",
		},
	}
}
