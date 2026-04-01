package rule

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// PointData 测点数据
type PointData struct {
	PointID    string
	Value      float64
	Timestamp  time.Time
	Quality    int // 数据质量
	Attributes map[string]interface{}
}

// TimeSeriesData 时间序列数据
type TimeSeriesData struct {
	PointID string
	Values  []PointData
}

// DataProvider 数据提供者接口
type DataProvider interface {
	// GetCurrentValue 获取当前值
	GetCurrentValue(ctx context.Context, pointID string) (*PointData, error)
	// GetTimeSeries 获取时间序列数据
	GetTimeSeries(ctx context.Context, pointID string, start, end time.Time) (*TimeSeriesData, error)
	// GetWindowData 获取窗口数据
	GetWindowData(ctx context.Context, pointID string, window time.Duration) (*TimeSeriesData, error)
}

// EvaluationContext 评估上下文
type EvaluationContext struct {
	CurrentTime time.Time
	PointValues map[string]float64 // 当前测点值缓存
	DataProvider DataProvider
}

// EvaluationResult 评估结果
type EvaluationResult struct {
	RuleID      string
	RuleName    string
	Triggered   bool
	Value       float64
	Threshold   float64
	Message     string
	Timestamp   time.Time
	Duration    time.Duration // 评估耗时
	Error       error
}

// Engine 规则执行引擎
type Engine struct {
	dataProvider DataProvider
	rules        map[string]*RuleDSL
	mu           sync.RWMutex
	cache        *evaluationCache
}

// evaluationCache 评估缓存
type evaluationCache struct {
	values     map[string]float64
	windows    map[string]*WindowResult
	mu         sync.RWMutex
	expiration time.Duration
}

// WindowResult 窗口计算结果
type WindowResult struct {
	Value     float64
	Timestamp time.Time
}

// NewEngine 创建规则执行引擎
func NewEngine(dataProvider DataProvider) *Engine {
	return &Engine{
		dataProvider: dataProvider,
		rules:        make(map[string]*RuleDSL),
		cache: &evaluationCache{
			values:     make(map[string]float64),
			windows:    make(map[string]*WindowResult),
			expiration: 5 * time.Second,
		},
	}
}

// AddRule 添加规则
func (e *Engine) AddRule(rule *RuleDSL) error {
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("invalid rule: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules[rule.ID] = rule
	return nil
}

// RemoveRule 移除规则
func (e *Engine) RemoveRule(ruleID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.rules, ruleID)
}

// GetRule 获取规则
func (e *Engine) GetRule(ruleID string) (*RuleDSL, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	rule, exists := e.rules[ruleID]
	return rule, exists
}

// GetAllRules 获取所有规则
func (e *Engine) GetAllRules() []*RuleDSL {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rules := make([]*RuleDSL, 0, len(e.rules))
	for _, rule := range e.rules {
		rules = append(rules, rule)
	}
	return rules
}

// Evaluate 评估单个规则
func (e *Engine) Evaluate(ctx context.Context, ruleID string) (*EvaluationResult, error) {
	startTime := time.Now()

	e.mu.RLock()
	rule, exists := e.rules[ruleID]
	e.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("rule not found: %s", ruleID)
	}

	// 检查规则是否启用
	if !rule.Enabled {
		return &EvaluationResult{
			RuleID:    ruleID,
			RuleName:  rule.Name,
			Triggered: false,
			Timestamp: time.Now(),
			Duration:  time.Since(startTime),
		}, nil
	}

	// 创建评估上下文
	evalCtx := &EvaluationContext{
		CurrentTime:  time.Now(),
		PointValues:  make(map[string]float64),
		DataProvider: e.dataProvider,
	}

	// 评估条件
	triggered, value, threshold, err := e.evaluateCondition(ctx, evalCtx, &rule.Condition)
	if err != nil {
		return &EvaluationResult{
			RuleID:    ruleID,
			RuleName:  rule.Name,
			Timestamp: time.Now(),
			Duration:  time.Since(startTime),
			Error:     err,
		}, nil
	}

	return &EvaluationResult{
		RuleID:    ruleID,
		RuleName:  rule.Name,
		Triggered: triggered,
		Value:     value,
		Threshold: threshold,
		Timestamp: time.Now(),
		Duration:  time.Since(startTime),
	}, nil
}

// EvaluateAll 评估所有启用的规则
func (e *Engine) EvaluateAll(ctx context.Context) ([]*EvaluationResult, error) {
	e.mu.RLock()
	rules := make([]*RuleDSL, 0, len(e.rules))
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

	results := make([]*EvaluationResult, 0, len(rules))
	for _, rule := range rules {
		result, err := e.Evaluate(ctx, rule.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate rule %s: %w", rule.ID, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// EvaluateBatch 批量评估指定规则
func (e *Engine) EvaluateBatch(ctx context.Context, ruleIDs []string) ([]*EvaluationResult, error) {
	results := make([]*EvaluationResult, 0, len(ruleIDs))

	for _, ruleID := range ruleIDs {
		result, err := e.Evaluate(ctx, ruleID)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate rule %s: %w", ruleID, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// EvaluateWithData 使用提供的数据评估规则
func (e *Engine) EvaluateWithData(ctx context.Context, rule *RuleDSL, data map[string]float64) (*EvaluationResult, error) {
	startTime := time.Now()

	if !rule.Enabled {
		return &EvaluationResult{
			RuleID:    rule.ID,
			RuleName:  rule.Name,
			Triggered: false,
			Timestamp: time.Now(),
			Duration:  time.Since(startTime),
		}, nil
	}

	evalCtx := &EvaluationContext{
		CurrentTime:  time.Now(),
		PointValues:  data,
		DataProvider: e.dataProvider,
	}

	triggered, value, threshold, err := e.evaluateCondition(ctx, evalCtx, &rule.Condition)
	if err != nil {
		return &EvaluationResult{
			RuleID:    rule.ID,
			RuleName:  rule.Name,
			Timestamp: time.Now(),
			Duration:  time.Since(startTime),
			Error:     err,
		}, nil
	}

	return &EvaluationResult{
		RuleID:    rule.ID,
		RuleName:  rule.Name,
		Triggered: triggered,
		Value:     value,
		Threshold: threshold,
		Timestamp: time.Now(),
		Duration:  time.Since(startTime),
	}, nil
}

// evaluateCondition 评估条件
func (e *Engine) evaluateCondition(ctx context.Context, evalCtx *EvaluationContext, condition *Condition) (bool, float64, float64, error) {
	if condition.Comparison != nil {
		return e.evaluateComparison(ctx, evalCtx, condition.Comparison)
	}
	if condition.Logical != nil {
		return e.evaluateLogical(ctx, evalCtx, condition.Logical)
	}
	return false, 0, 0, fmt.Errorf("invalid condition: no comparison or logical")
}

// evaluateComparison 评估比较条件
func (e *Engine) evaluateComparison(ctx context.Context, evalCtx *EvaluationContext, comp *ComparisonCondition) (bool, float64, float64, error) {
	// 获取左值
	leftValue, err := e.getValue(ctx, evalCtx, &comp.Left)
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to get left value: %w", err)
	}

	// 计算右值（阈值）
	rightValue, err := e.calculateThreshold(evalCtx, leftValue, &comp.Right)
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to calculate threshold: %w", err)
	}

	// 执行比较
	triggered := e.compare(leftValue, comp.Operator, rightValue)

	return triggered, leftValue, rightValue, nil
}

// evaluateLogical 评估逻辑条件
func (e *Engine) evaluateLogical(ctx context.Context, evalCtx *EvaluationContext, logical *LogicalCondition) (bool, float64, float64, error) {
	switch logical.Operator {
	case LogicalAND:
		for _, operand := range logical.Operands {
			cond := &Condition{}
			switch o := operand.(type) {
			case *ComparisonCondition:
				cond.Comparison = o
			case *LogicalCondition:
				cond.Logical = o
			}
			triggered, value, threshold, err := e.evaluateCondition(ctx, evalCtx, cond)
			if err != nil {
				return false, 0, 0, err
			}
			if !triggered {
				return false, value, threshold, nil
			}
		}
		return true, 0, 0, nil

	case LogicalOR:
		for _, operand := range logical.Operands {
			cond := &Condition{}
			switch o := operand.(type) {
			case *ComparisonCondition:
				cond.Comparison = o
			case *LogicalCondition:
				cond.Logical = o
			}
			triggered, value, threshold, err := e.evaluateCondition(ctx, evalCtx, cond)
			if err != nil {
				return false, 0, 0, err
			}
			if triggered {
				return true, value, threshold, nil
			}
		}
		return false, 0, 0, nil

	case LogicalNOT:
		if len(logical.Operands) == 0 {
			return false, 0, 0, fmt.Errorf("NOT operator requires 1 operand")
		}
		cond := &Condition{}
		switch o := logical.Operands[0].(type) {
		case *ComparisonCondition:
			cond.Comparison = o
		case *LogicalCondition:
			cond.Logical = o
		}
		triggered, value, threshold, err := e.evaluateCondition(ctx, evalCtx, cond)
		if err != nil {
			return false, 0, 0, err
		}
		return !triggered, value, threshold, nil

	default:
		return false, 0, 0, fmt.Errorf("unknown logical operator: %s", logical.Operator)
	}
}

// getValue 获取值
func (e *Engine) getValue(ctx context.Context, evalCtx *EvaluationContext, valueExpr *ValueExpression) (float64, error) {
	pointID := valueExpr.PointID

	// 如果有窗口函数，计算窗口值
	if valueExpr.Function != "" {
		return e.calculateWindowValue(ctx, evalCtx, valueExpr)
	}

	// 检查缓存
	if value, exists := evalCtx.PointValues[pointID]; exists {
		return value, nil
	}

	// 从数据提供者获取当前值
	if evalCtx.DataProvider != nil {
		data, err := evalCtx.DataProvider.GetCurrentValue(ctx, pointID)
		if err != nil {
			return 0, fmt.Errorf("failed to get current value for point %s: %w", pointID, err)
		}
		evalCtx.PointValues[pointID] = data.Value
		return data.Value, nil
	}

	return 0, fmt.Errorf("no data provider available")
}

// calculateWindowValue 计算窗口函数值
func (e *Engine) calculateWindowValue(ctx context.Context, evalCtx *EvaluationContext, valueExpr *ValueExpression) (float64, error) {
	if evalCtx.DataProvider == nil {
		return 0, fmt.Errorf("no data provider available")
	}

	// 获取窗口数据
	data, err := evalCtx.DataProvider.GetWindowData(ctx, valueExpr.PointID, valueExpr.WindowSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get window data: %w", err)
	}

	if len(data.Values) == 0 {
		return 0, fmt.Errorf("no data in window")
	}

	// 计算窗口函数
	var result float64
	switch valueExpr.Function {
	case WindowAvg:
		sum := 0.0
		for _, v := range data.Values {
			sum += v.Value
		}
		result = sum / float64(len(data.Values))

	case WindowMax:
		result = data.Values[0].Value
		for _, v := range data.Values {
			if v.Value > result {
				result = v.Value
			}
		}

	case WindowMin:
		result = data.Values[0].Value
		for _, v := range data.Values {
			if v.Value < result {
				result = v.Value
			}
		}

	case WindowSum:
		for _, v := range data.Values {
			result += v.Value
		}

	case WindowCount:
		result = float64(len(data.Values))

	default:
		return 0, fmt.Errorf("unknown window function: %s", valueExpr.Function)
	}

	return result, nil
}

// calculateThreshold 计算阈值
func (e *Engine) calculateThreshold(evalCtx *EvaluationContext, currentValue float64, threshold *Threshold) (float64, error) {
	switch threshold.Type {
	case ThresholdTypeAbsolute:
		return threshold.Value, nil

	case ThresholdTypePercentage:
		// 百分比阈值：当前值的百分比
		return currentValue * (threshold.Value / 100.0), nil

	case ThresholdTypeRate:
		// 变化率阈值：这里简化处理，实际需要历史数据
		return threshold.Value, nil

	default:
		return threshold.Value, nil
	}
}

// compare 执行比较操作
func (e *Engine) compare(left float64, op Operator, right float64) bool {
	switch op {
	case OpGT:
		return left > right
	case OpLT:
		return left < right
	case OpGTE:
		return left >= right
	case OpLTE:
		return left <= right
	case OpEQ:
		return left == right
	case OpNE:
		return left != right
	default:
		return false
	}
}

// EnableRule 启用规则
func (e *Engine) EnableRule(ruleID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	rule, exists := e.rules[ruleID]
	if !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	rule.Enabled = true
	return nil
}

// DisableRule 禁用规则
func (e *Engine) DisableRule(ruleID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	rule, exists := e.rules[ruleID]
	if !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	rule.Enabled = false
	return nil
}

// GetEnabledRules 获取所有启用的规则
func (e *Engine) GetEnabledRules() []*RuleDSL {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rules := make([]*RuleDSL, 0)
	for _, rule := range e.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}

	// 按优先级排序
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	return rules
}

// GetRulesByPriority 按优先级获取规则
func (e *Engine) GetRulesByPriority(minPriority, maxPriority int) []*RuleDSL {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rules := make([]*RuleDSL, 0)
	for _, rule := range e.rules {
		if rule.Priority >= minPriority && rule.Priority <= maxPriority {
			rules = append(rules, rule)
		}
	}

	// 按优先级排序
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	return rules
}

// ClearCache 清除缓存
func (e *Engine) ClearCache() {
	e.cache.mu.Lock()
	defer e.cache.mu.Unlock()
	e.cache.values = make(map[string]float64)
	e.cache.windows = make(map[string]*WindowResult)
}
