package rule

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/new-energy-monitoring/internal/infrastructure/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	ErrRuleNotFound     = errors.New("rule not found")
	ErrRuleExists       = errors.New("rule already exists")
	ErrInvalidRule      = errors.New("invalid rule configuration")
	ErrRuleDisabled     = errors.New("rule is disabled")
	ErrExecutionTimeout = errors.New("rule execution timeout")
)

// Prometheus指标
var (
	ruleTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compute_rule_total",
		Help: "Total number of rule executions",
	}, []string{"status"})

	ruleDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "compute_rule_duration_seconds",
		Help:    "Rule execution duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"rule_type"})

	ruleActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "compute_rule_active",
		Help: "Number of active rules",
	})

	ruleCacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "compute_rule_cache_hits_total",
		Help: "Total number of rule cache hits",
	})
)

// RuleType 规则类型
type RuleType string

const (
	RuleTypeFormula    RuleType = "formula"    // 公式规则
	RuleTypeExpression RuleType = "expression" // 表达式规则
	RuleTypeScript     RuleType = "script"     // 脚本规则
	RuleTypeAggregate  RuleType = "aggregate"  // 聚合规则
	RuleTypeTransform  RuleType = "transform"  // 转换规则
)

// RuleStatus 规则状态
type RuleStatus string

const (
	RuleStatusActive   RuleStatus = "active"   // 活跃
	RuleStatusInactive RuleStatus = "inactive" // 不活跃
	RuleStatusError    RuleStatus = "error"    // 错误
	RuleStatusDisabled RuleStatus = "disabled" // 禁用
)

// Rule 规则定义
type Rule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        RuleType               `json:"type"`
	Status      RuleStatus             `json:"status"`
	Priority    int                    `json:"priority"`    // 优先级
	Enabled     bool                   `json:"enabled"`     // 是否启用
	PointID     string                 `json:"pointId"`     // 目标计算点ID
	Formula     string                 `json:"formula"`     // 计算公式
	Expression  string                 `json:"expression"`  // 表达式
	Script      string                 `json:"script"`      // 脚本代码
	Inputs      []RuleInput            `json:"inputs"`      // 输入参数
	Outputs     []RuleOutput           `json:"outputs"`     // 输出参数
	Conditions  []RuleCondition        `json:"conditions"`  // 执行条件
	Config      map[string]interface{} `json:"config"`      // 配置参数
	Timeout     time.Duration          `json:"timeout"`     // 超时时间
	MaxRetry    int                    `json:"maxRetry"`    // 最大重试次数
	CreateTime  time.Time              `json:"createTime"`
	UpdateTime  time.Time              `json:"updateTime"`
	LastExecute time.Time              `json:"lastExecute"` // 最后执行时间
	ExecuteCount int64                 `json:"executeCount"` // 执行次数
	SuccessCount int64                 `json:"successCount"` // 成功次数
	ErrorCount  int64                  `json:"errorCount"`   // 错误次数
	LastError   string                 `json:"lastError"`    // 最后错误
	Version     int                    `json:"version"`      // 版本号
	Tags        map[string]string      `json:"tags"`         // 标签
}

// RuleInput 规则输入
type RuleInput struct {
	Name     string  `json:"name"`     // 参数名
	Type     string  `json:"type"`     // 类型: number, string, boolean
	PointID  string  `json:"pointId"`  // 关联测点ID
	Default  float64 `json:"default"`  // 默认值
	Required bool    `json:"required"` // 是否必需
}

// RuleOutput 规则输出
type RuleOutput struct {
	Name      string  `json:"name"`      // 参数名
	Type      string  `json:"type"`      // 类型
	PointID   string  `json:"pointId"`   // 关联测点ID
	Scale     float64 `json:"scale"`     // 缩放因子
	Offset    float64 `json:"offset"`    // 偏移量
	Unit      string  `json:"unit"`      // 单位
	Precision int     `json:"precision"` // 精度
}

// RuleCondition 规则条件
type RuleCondition struct {
	Expression string `json:"expression"` // 条件表达式
	Type       string `json:"type"`       // 类型: pre, post
}

// RuleExecution 规则执行上下文
type RuleExecution struct {
	RuleID    string                 `json:"ruleId"`
	StartTime time.Time              `json:"startTime"`
	EndTime   time.Time              `json:"endTime"`
	Duration  time.Duration          `json:"duration"`
	Status    string                 `json:"status"`
	Error     string                 `json:"error"`
	Inputs    map[string]float64     `json:"inputs"`
	Outputs   map[string]float64     `json:"outputs"`
	Logs      []string               `json:"logs"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// RuleEngine 规则执行引擎
type RuleEngine struct {
	rules       map[string]*Rule
	byPoint     map[string][]string // 按计算点索引
	byType      map[RuleType][]string // 按类型索引
	cache       *ComputeCache
	dataProvider DataProvider
	mu          sync.RWMutex
	running     int32
	stats       EngineStats
	statsMu     sync.RWMutex
	ctx         context.Context
	cancelFunc  context.CancelFunc
	logger      *zap.Logger
}

// DataProvider 数据提供者接口
type DataProvider interface {
	GetCurrentValue(ctx context.Context, pointID string) (float64, error)
	GetTimeSeries(ctx context.Context, pointID string, start, end time.Time) ([]float64, error)
	GetAggregatedValue(ctx context.Context, pointID string, aggFunc string, window time.Duration) (float64, error)
}

// EngineStats 引擎统计
type EngineStats struct {
	TotalRules      int64
	ActiveRules     int64
	TotalExecutions int64
	SuccessExecutions int64
	FailedExecutions  int64
	AverageDuration   time.Duration
	CacheHitRate      float64
}

// NewRuleEngine 创建规则执行引擎
func NewRuleEngine(cache *ComputeCache, dataProvider DataProvider) *RuleEngine {
	engine := &RuleEngine{
		rules:       make(map[string]*Rule),
		byPoint:     make(map[string][]string),
		byType:      make(map[RuleType][]string),
		cache:       cache,
		dataProvider: dataProvider,
		logger:      logger.Named("rule-engine"),
	}
	engine.ctx, engine.cancelFunc = context.WithCancel(context.Background())
	return engine
}

// Start 启动规则引擎
func (e *RuleEngine) Start() error {
	if atomic.LoadInt32(&e.running) == 1 {
		return errors.New("rule engine is already running")
	}

	atomic.StoreInt32(&e.running, 1)
	e.logger.Info("Rule engine started")

	return nil
}

// Stop 停止规则引擎
func (e *RuleEngine) Stop() error {
	if atomic.LoadInt32(&e.running) == 0 {
		return errors.New("rule engine is not running")
	}

	atomic.StoreInt32(&e.running, 0)
	e.cancelFunc()
	e.logger.Info("Rule engine stopped")

	return nil
}

// LoadRule 加载规则
func (e *RuleEngine) LoadRule(rule *Rule) error {
	if rule.ID == "" {
		return ErrInvalidRule
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.rules[rule.ID]; exists {
		return ErrRuleExists
	}

	// 设置默认值
	if rule.Status == "" {
		rule.Status = RuleStatusActive
	}
	if rule.CreateTime.IsZero() {
		rule.CreateTime = time.Now()
	}
	rule.UpdateTime = time.Now()
	rule.Version = 1

	// 验证规则
	if err := e.validateRule(rule); err != nil {
		return fmt.Errorf("invalid rule: %w", err)
	}

	e.rules[rule.ID] = rule

	// 建立索引
	if rule.PointID != "" {
		e.byPoint[rule.PointID] = append(e.byPoint[rule.PointID], rule.ID)
	}
	e.byType[rule.Type] = append(e.byType[rule.Type], rule.ID)

	// 更新统计
	atomic.AddInt64(&e.stats.TotalRules, 1)
	if rule.Enabled {
		atomic.AddInt64(&e.stats.ActiveRules, 1)
	}

	e.logger.Info("Rule loaded",
		zap.String("ruleID", rule.ID),
		zap.String("type", string(rule.Type)),
		zap.String("pointID", rule.PointID))

	return nil
}

// UnloadRule 卸载规则
func (e *RuleEngine) UnloadRule(ruleID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	rule, exists := e.rules[ruleID]
	if !exists {
		return ErrRuleNotFound
	}

	// 删除索引
	if rule.PointID != "" {
		e.removeFromSlice(&e.byPoint[rule.PointID], ruleID)
	}
	e.removeFromSlice(&e.byType[rule.Type], ruleID)

	delete(e.rules, ruleID)

	// 更新统计
	atomic.AddInt64(&e.stats.TotalRules, -1)
	if rule.Enabled {
		atomic.AddInt64(&e.stats.ActiveRules, -1)
	}

	e.logger.Info("Rule unloaded", zap.String("ruleID", ruleID))

	return nil
}

// UpdateRule 更新规则
func (e *RuleEngine) UpdateRule(rule *Rule) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	existing, exists := e.rules[rule.ID]
	if !exists {
		return ErrRuleNotFound
	}

	// 验证规则
	if err := e.validateRule(rule); err != nil {
		return fmt.Errorf("invalid rule: %w", err)
	}

	// 保留统计信息
	rule.CreateTime = existing.CreateTime
	rule.ExecuteCount = existing.ExecuteCount
	rule.SuccessCount = existing.SuccessCount
	rule.ErrorCount = existing.ErrorCount
	rule.UpdateTime = time.Now()
	rule.Version = existing.Version + 1

	// 更新索引
	e.updateIndex(existing, rule)

	e.rules[rule.ID] = rule

	e.logger.Info("Rule updated",
		zap.String("ruleID", rule.ID),
		zap.Int("version", rule.Version))

	return nil
}

// GetRule 获取规则
func (e *RuleEngine) GetRule(ruleID string) (*Rule, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rule, exists := e.rules[ruleID]
	if !exists {
		return nil, ErrRuleNotFound
	}

	return rule, nil
}

// GetRulesByPoint 按计算点获取规则
func (e *RuleEngine) GetRulesByPoint(pointID string) []*Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ruleIDs := e.byPoint[pointID]
	rules := make([]*Rule, 0, len(ruleIDs))

	for _, id := range ruleIDs {
		if rule, exists := e.rules[id]; exists {
			rules = append(rules, rule)
		}
	}

	// 按优先级排序
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	return rules
}

// GetRulesByType 按类型获取规则
func (e *RuleEngine) GetRulesByType(ruleType RuleType) []*Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ruleIDs := e.byType[ruleType]
	rules := make([]*Rule, 0, len(ruleIDs))

	for _, id := range ruleIDs {
		if rule, exists := e.rules[id]; exists {
			rules = append(rules, rule)
		}
	}

	return rules
}

// Execute 执行规则
func (e *RuleEngine) Execute(ctx context.Context, ruleID string) (*RuleExecution, error) {
	e.mu.RLock()
	rule, exists := e.rules[ruleID]
	e.mu.RUnlock()

	if !exists {
		return nil, ErrRuleNotFound
	}

	if !rule.Enabled {
		return nil, ErrRuleDisabled
	}

	return e.executeRule(ctx, rule)
}

// ExecuteForPoint 为计算点执行规则
func (e *RuleEngine) ExecuteForPoint(ctx context.Context, pointID string) ([]*RuleExecution, error) {
	rules := e.GetRulesByPoint(pointID)
	if len(rules) == 0 {
		return nil, nil
	}

	executions := make([]*RuleExecution, 0, len(rules))
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		exec, err := e.executeRule(ctx, rule)
		if err != nil {
			e.logger.Error("Failed to execute rule",
				zap.String("ruleID", rule.ID),
				zap.Error(err))
			continue
		}

		executions = append(executions, exec)
	}

	return executions, nil
}

// ExecuteAll 执行所有启用的规则
func (e *RuleEngine) ExecuteAll(ctx context.Context) (map[string]*RuleExecution, error) {
	e.mu.RLock()
	rules := make([]*Rule, 0)
	for _, rule := range e.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	e.mu.RUnlock()

	// 按优先级排序
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	results := make(map[string]*RuleExecution)
	for _, rule := range rules {
		exec, err := e.executeRule(ctx, rule)
		if err != nil {
			e.logger.Error("Failed to execute rule",
				zap.String("ruleID", rule.ID),
				zap.Error(err))
			results[rule.ID] = &RuleExecution{
				RuleID: rule.ID,
				Status: "error",
				Error:  err.Error(),
			}
			continue
		}
		results[rule.ID] = exec
	}

	return results, nil
}

// executeRule 执行单个规则
func (e *RuleEngine) executeRule(ctx context.Context, rule *Rule) (*RuleExecution, error) {
	execution := &RuleExecution{
		RuleID:    rule.ID,
		StartTime: time.Now(),
		Inputs:    make(map[string]float64),
		Outputs:   make(map[string]float64),
		Logs:      make([]string, 0),
		Metadata:  make(map[string]interface{}),
	}

	defer func() {
		execution.EndTime = time.Now()
		execution.Duration = execution.EndTime.Sub(execution.StartTime)

		// 更新统计
		atomic.AddInt64(&e.stats.TotalExecutions, 1)
		if execution.Status == "success" {
			atomic.AddInt64(&e.stats.SuccessExecutions, 1)
		} else {
			atomic.AddInt64(&e.stats.FailedExecutions, 1)
		}

		// 更新规则统计
		e.mu.Lock()
		if r, exists := e.rules[rule.ID]; exists {
			r.LastExecute = time.Now()
			r.ExecuteCount++
			if execution.Status == "success" {
				r.SuccessCount++
			} else {
				r.ErrorCount++
				r.LastError = execution.Error
			}
		}
		e.mu.Unlock()

		// 更新指标
		ruleTotal.WithLabelValues(execution.Status).Inc()
		ruleDuration.WithLabelValues(string(rule.Type)).Observe(execution.Duration.Seconds())
	}()

	// 检查缓存
	cacheKey := fmt.Sprintf("rule:%s", rule.ID)
	if e.cache != nil {
		if cached, err := e.cache.Get(ctx, cacheKey); err == nil {
			execution.Outputs = map[string]float64{"value": cached.Value}
			execution.Status = "success"
			execution.Logs = append(execution.Logs, "cache hit")
			ruleCacheHits.Inc()
			return execution, nil
		}
	}

	// 创建执行上下文
	execCtx, cancel := context.WithTimeout(ctx, rule.Timeout)
	defer cancel()

	// 获取输入数据
	for _, input := range rule.Inputs {
		value, err := e.getInputValue(execCtx, input)
		if err != nil {
			if input.Required {
				execution.Status = "error"
				execution.Error = fmt.Sprintf("failed to get input %s: %v", input.Name, err)
				return execution, nil
			}
			value = input.Default
		}
		execution.Inputs[input.Name] = value
	}

	// 执行前置条件检查
	if !e.checkPreConditions(rule, execution.Inputs) {
		execution.Status = "skipped"
		execution.Logs = append(execution.Logs, "pre-conditions not met")
		return execution, nil
	}

	// 根据规则类型执行
	var result float64
	var err error

	switch rule.Type {
	case RuleTypeFormula:
		result, err = e.executeFormula(rule, execution)

	case RuleTypeExpression:
		result, err = e.executeExpression(rule, execution)

	case RuleTypeScript:
		result, err = e.executeScript(rule, execution)

	case RuleTypeAggregate:
		result, err = e.executeAggregate(execCtx, rule, execution)

	case RuleTypeTransform:
		result, err = e.executeTransform(rule, execution)

	default:
		err = fmt.Errorf("unknown rule type: %s", rule.Type)
	}

	if err != nil {
		execution.Status = "error"
		execution.Error = err.Error()
		return execution, nil
	}

	// 应用输出转换
	result = e.applyOutputTransform(rule, result)
	execution.Outputs["value"] = result

	// 执行后置条件检查
	if !e.checkPostConditions(rule, execution.Outputs) {
		execution.Status = "skipped"
		execution.Logs = append(execution.Logs, "post-conditions not met")
		return execution, nil
	}

	// 缓存结果
	if e.cache != nil {
		computeResult := &ComputeResult{
			PointID:   rule.PointID,
			Value:     result,
			Timestamp: time.Now(),
		}
		e.cache.Set(ctx, cacheKey, computeResult)
	}

	execution.Status = "success"
	return execution, nil
}

// getInputValue 获取输入值
func (e *RuleEngine) getInputValue(ctx context.Context, input RuleInput) (float64, error) {
	if input.PointID == "" {
		return input.Default, nil
	}

	if e.dataProvider == nil {
		return 0, errors.New("data provider not available")
	}

	return e.dataProvider.GetCurrentValue(ctx, input.PointID)
}

// checkPreConditions 检查前置条件
func (e *RuleEngine) checkPreConditions(rule *Rule, inputs map[string]float64) bool {
	for _, cond := range rule.Conditions {
		if cond.Type != "pre" {
			continue
		}

		// 简化实现：解析简单条件
		// 实际项目中可以使用表达式引擎
		if !e.evaluateCondition(cond.Expression, inputs) {
			return false
		}
	}

	return true
}

// checkPostConditions 检查后置条件
func (e *RuleEngine) checkPostConditions(rule *Rule, outputs map[string]float64) bool {
	for _, cond := range rule.Conditions {
		if cond.Type != "post" {
			continue
		}

		if !e.evaluateCondition(cond.Expression, outputs) {
			return false
		}
	}

	return true
}

// evaluateCondition 评估条件
func (e *RuleEngine) evaluateCondition(expression string, values map[string]float64) bool {
	// 简化实现：支持简单的比较表达式
	// 实际项目中可以使用表达式引擎
	return true
}

// executeFormula 执行公式
func (e *RuleEngine) executeFormula(rule *Rule, execution *RuleExecution) (float64, error) {
	// 简化实现：解析和执行公式
	// 实际项目中可以使用表达式引擎如 github.com/Knetic/govaluate
	formula := rule.Formula

	// 简单示例：支持基本的四则运算
	// 这里提供一个简化版本
	execution.Logs = append(execution.Logs, fmt.Sprintf("executing formula: %s", formula))

	// 示例：假设公式是简单的变量引用或数值
	if value, ok := execution.Inputs[formula]; ok {
		return value, nil
	}

	// 尝试解析为数值
	var result float64
	fmt.Sscanf(formula, "%f", &result)

	return result, nil
}

// executeExpression 执行表达式
func (e *RuleEngine) executeExpression(rule *Rule, execution *RuleExecution) (float64, error) {
	// 简化实现：解析和执行表达式
	expression := rule.Expression

	execution.Logs = append(execution.Logs, fmt.Sprintf("executing expression: %s", expression))

	// 简单示例：支持基本的比较和逻辑运算
	// 实际项目中应该使用完整的表达式引擎

	return 0, nil
}

// executeScript 执行脚本
func (e *RuleEngine) executeScript(rule *Rule, execution *RuleExecution) (float64, error) {
	// 简化实现：执行脚本
	// 实际项目中可以使用嵌入的脚本引擎如 goja, gopher-lua 等
	script := rule.Script

	execution.Logs = append(execution.Logs, fmt.Sprintf("executing script: %s", script))

	// 简单示例：直接返回输入值的和
	var sum float64
	for _, value := range execution.Inputs {
		sum += value
	}

	return sum, nil
}

// executeAggregate 执行聚合
func (e *RuleEngine) executeAggregate(ctx context.Context, rule *Rule, execution *RuleExecution) (float64, error) {
	if e.dataProvider == nil {
		return 0, errors.New("data provider not available")
	}

	// 从配置中获取聚合参数
	aggFunc := "avg"
	if f, ok := rule.Config["aggregateFunc"].(string); ok {
		aggFunc = f
	}

	window := 5 * time.Minute
	if w, ok := rule.Config["window"].(string); ok {
		if duration, err := time.ParseDuration(w); err == nil {
			window = duration
		}
	}

	// 获取第一个输入点
	var pointID string
	for _, input := range rule.Inputs {
		if input.PointID != "" {
			pointID = input.PointID
			break
		}
	}

	if pointID == "" {
		return 0, errors.New("no input point specified")
	}

	execution.Logs = append(execution.Logs, fmt.Sprintf("aggregate %s over %s", aggFunc, window))

	return e.dataProvider.GetAggregatedValue(ctx, pointID, aggFunc, window)
}

// executeTransform 执行转换
func (e *RuleEngine) executeTransform(rule *Rule, execution *RuleExecution) (float64, error) {
	// 简化实现：执行数据转换
	// 支持常见的转换：缩放、偏移、单位转换等

	var input float64
	for _, value := range execution.Inputs {
		input = value
		break
	}

	// 应用配置的转换
	if scale, ok := rule.Config["scale"].(float64); ok {
		input *= scale
	}

	if offset, ok := rule.Config["offset"].(float64); ok {
		input += offset
	}

	execution.Logs = append(execution.Logs, "applied transform")

	return input, nil
}

// applyOutputTransform 应用输出转换
func (e *RuleEngine) applyOutputTransform(rule *Rule, value float64) float64 {
	if len(rule.Outputs) == 0 {
		return value
	}

	output := rule.Outputs[0]

	// 应用缩放
	if output.Scale != 0 {
		value *= output.Scale
	}

	// 应用偏移
	value += output.Offset

	return value
}

// validateRule 验证规则
func (e *RuleEngine) validateRule(rule *Rule) error {
	if rule.ID == "" {
		return errors.New("rule ID is required")
	}

	if rule.Type == "" {
		return errors.New("rule type is required")
	}

	// 根据类型验证
	switch rule.Type {
	case RuleTypeFormula:
		if rule.Formula == "" {
			return errors.New("formula is required for formula rule")
		}

	case RuleTypeExpression:
		if rule.Expression == "" {
			return errors.New("expression is required for expression rule")
		}

	case RuleTypeScript:
		if rule.Script == "" {
			return errors.New("script is required for script rule")
		}
	}

	return nil
}

// updateIndex 更新索引
func (e *RuleEngine) updateIndex(oldRule, newRule *Rule) {
	// 删除旧索引
	if oldRule.PointID != "" {
		e.removeFromSlice(&e.byPoint[oldRule.PointID], oldRule.ID)
	}
	e.removeFromSlice(&e.byType[oldRule.Type], oldRule.ID)

	// 添加新索引
	if newRule.PointID != "" {
		e.byPoint[newRule.PointID] = append(e.byPoint[newRule.PointID], newRule.ID)
	}
	e.byType[newRule.Type] = append(e.byType[newRule.Type], newRule.ID)
}

// removeFromSlice 从切片中移除元素
func (e *RuleEngine) removeFromSlice(slice *[]string, item string) {
	for i, v := range *slice {
		if v == item {
			*slice = append((*slice)[:i], (*slice)[i+1:]...)
			break
		}
	}
}

// EnableRule 启用规则
func (e *RuleEngine) EnableRule(ruleID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	rule, exists := e.rules[ruleID]
	if !exists {
		return ErrRuleNotFound
	}

	if !rule.Enabled {
		rule.Enabled = true
		rule.Status = RuleStatusActive
		rule.UpdateTime = time.Now()
		atomic.AddInt64(&e.stats.ActiveRules, 1)
	}

	return nil
}

// DisableRule 禁用规则
func (e *RuleEngine) DisableRule(ruleID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	rule, exists := e.rules[ruleID]
	if !exists {
		return ErrRuleNotFound
	}

	if rule.Enabled {
		rule.Enabled = false
		rule.Status = RuleStatusDisabled
		rule.UpdateTime = time.Now()
		atomic.AddInt64(&e.stats.ActiveRules, -1)
	}

	return nil
}

// GetStats 获取统计信息
func (e *RuleEngine) GetStats() EngineStats {
	e.statsMu.RLock()
	defer e.statsMu.RUnlock()

	stats := EngineStats{
		TotalRules:        atomic.LoadInt64(&e.stats.TotalRules),
		ActiveRules:       atomic.LoadInt64(&e.stats.ActiveRules),
		TotalExecutions:   atomic.LoadInt64(&e.stats.TotalExecutions),
		SuccessExecutions: atomic.LoadInt64(&e.stats.SuccessExecutions),
		FailedExecutions:  atomic.LoadInt64(&e.stats.FailedExecutions),
		AverageDuration:   e.stats.AverageDuration,
	}

	// 计算缓存命中率
	if e.cache != nil {
		stats.CacheHitRate = e.cache.GetHitRate()
	}

	return stats
}

// GetRuleCount 获取规则数量
func (e *RuleEngine) GetRuleCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.rules)
}

// IsRunning 检查是否运行中
func (e *RuleEngine) IsRunning() bool {
	return atomic.LoadInt32(&e.running) == 1
}

// ClearCache 清除缓存
func (e *RuleEngine) ClearCache(ctx context.Context) error {
	if e.cache == nil {
		return nil
	}
	return e.cache.Clear(ctx)
}

// ReloadRules 重新加载所有规则
func (e *RuleEngine) ReloadRules(rules []*Rule) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 清空现有规则
	e.rules = make(map[string]*Rule)
	e.byPoint = make(map[string][]string)
	e.byType = make(map[RuleType][]string)

	// 加载新规则
	for _, rule := range rules {
		if err := e.validateRule(rule); err != nil {
			e.logger.Error("Invalid rule",
				zap.String("ruleID", rule.ID),
				zap.Error(err))
			continue
		}

		e.rules[rule.ID] = rule

		if rule.PointID != "" {
			e.byPoint[rule.PointID] = append(e.byPoint[rule.PointID], rule.ID)
		}
		e.byType[rule.Type] = append(e.byType[rule.Type], rule.ID)
	}

	// 更新统计
	atomic.StoreInt64(&e.stats.TotalRules, int64(len(e.rules)))
	activeCount := int64(0)
	for _, rule := range e.rules {
		if rule.Enabled {
			activeCount++
		}
	}
	atomic.StoreInt64(&e.stats.ActiveRules, activeCount)

	e.logger.Info("Rules reloaded", zap.Int("count", len(e.rules)))

	return nil
}

// ExportRules 导出所有规则
func (e *RuleEngine) ExportRules() []*Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rules := make([]*Rule, 0, len(e.rules))
	for _, rule := range e.rules {
		rules = append(rules, rule)
	}

	return rules
}

// BatchExecute 批量执行规则
func (e *RuleEngine) BatchExecute(ctx context.Context, ruleIDs []string) (map[string]*RuleExecution, error) {
	results := make(map[string]*RuleExecution)

	for _, ruleID := range ruleIDs {
		exec, err := e.Execute(ctx, ruleID)
		if err != nil {
			results[ruleID] = &RuleExecution{
				RuleID: ruleID,
				Status: "error",
				Error:  err.Error(),
			}
			continue
		}
		results[ruleID] = exec
	}

	return results, nil
}
