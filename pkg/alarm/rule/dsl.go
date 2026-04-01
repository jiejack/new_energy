package rule

import (
	"encoding/json"
	"fmt"
	"time"
)

// ThresholdType 阈值类型
type ThresholdType string

const (
	ThresholdTypeAbsolute   ThresholdType = "absolute"   // 绝对值
	ThresholdTypePercentage ThresholdType = "percentage" // 百分比
	ThresholdTypeRate       ThresholdType = "rate"       // 变化率
)

// Operator 比较运算符
type Operator string

const (
	OpGT  Operator = ">"  // 大于
	OpLT  Operator = "<"  // 小于
	OpGTE Operator = ">=" // 大于等于
	OpLTE Operator = "<=" // 小于等于
	OpEQ  Operator = "==" // 等于
	OpNE  Operator = "!=" // 不等于
)

// LogicalOperator 逻辑运算符
type LogicalOperator string

const (
	LogicalAND LogicalOperator = "AND"
	LogicalOR  LogicalOperator = "OR"
	LogicalNOT LogicalOperator = "NOT"
)

// WindowFunction 时间窗口函数
type WindowFunction string

const (
	WindowAvg   WindowFunction = "avg"   // 平均值
	WindowMax   WindowFunction = "max"   // 最大值
	WindowMin   WindowFunction = "min"   // 最小值
	WindowSum   WindowFunction = "sum"   // 求和
	WindowCount WindowFunction = "count" // 计数
)

// Expression 表达式接口
type Expression interface {
	String() string
	Validate() error
}

// ValueExpression 值表达式
type ValueExpression struct {
	PointID    string         `json:"point_id"`              // 测点ID
	Function   WindowFunction `json:"function,omitempty"`    // 窗口函数
	WindowSize time.Duration  `json:"window_size,omitempty"` // 时间窗口大小
}

func (v *ValueExpression) String() string {
	if v.Function != "" {
		return fmt.Sprintf("%s(%s, %v)", v.Function, v.PointID, v.WindowSize)
	}
	return v.PointID
}

func (v *ValueExpression) Validate() error {
	if v.PointID == "" {
		return fmt.Errorf("point_id is required")
	}
	if v.Function != "" && v.WindowSize <= 0 {
		return fmt.Errorf("window_size is required when function is specified")
	}
	return nil
}

// Threshold 阈值
type Threshold struct {
	Type  ThresholdType `json:"type"`  // 阈值类型
	Value float64       `json:"value"` // 阈值
}

func (t *Threshold) String() string {
	return fmt.Sprintf("%s(%v)", t.Type, t.Value)
}

func (t *Threshold) Validate() error {
	if t.Type == ThresholdTypePercentage && (t.Value < 0 || t.Value > 100) {
		return fmt.Errorf("percentage threshold must be between 0 and 100")
	}
	return nil
}

// ComparisonCondition 比较条件
type ComparisonCondition struct {
	Left     ValueExpression `json:"left"`     // 左值
	Operator Operator        `json:"operator"`  // 比较运算符
	Right    Threshold       `json:"right"`     // 右值（阈值）
}

func (c *ComparisonCondition) String() string {
	return fmt.Sprintf("%s %s %s", c.Left.String(), c.Operator, c.Right.String())
}

func (c *ComparisonCondition) Validate() error {
	if err := c.Left.Validate(); err != nil {
		return fmt.Errorf("invalid left value: %w", err)
	}
	if err := c.Right.Validate(); err != nil {
		return fmt.Errorf("invalid right value: %w", err)
	}
	if !isValidOperator(c.Operator) {
		return fmt.Errorf("invalid operator: %s", c.Operator)
	}
	return nil
}

// LogicalCondition 逻辑条件
type LogicalCondition struct {
	Operator LogicalOperator `json:"operator"` // 逻辑运算符
	Operands []Expression    `json:"operands"` // 操作数
}

func (l *LogicalCondition) String() string {
	if l.Operator == LogicalNOT {
		if len(l.Operands) > 0 {
			return fmt.Sprintf("NOT %s", l.Operands[0].String())
		}
		return "NOT"
	}

	result := "("
	for i, operand := range l.Operands {
		if i > 0 {
			result += fmt.Sprintf(" %s ", l.Operator)
		}
		result += operand.String()
	}
	result += ")"
	return result
}

func (l *LogicalCondition) Validate() error {
	if !isValidLogicalOperator(l.Operator) {
		return fmt.Errorf("invalid logical operator: %s", l.Operator)
	}

	if l.Operator == LogicalNOT {
		if len(l.Operands) != 1 {
			return fmt.Errorf("NOT operator requires exactly 1 operand")
		}
	} else {
		if len(l.Operands) < 2 {
			return fmt.Errorf("%s operator requires at least 2 operands", l.Operator)
		}
	}

	for i, operand := range l.Operands {
		if err := operand.Validate(); err != nil {
			return fmt.Errorf("invalid operand %d: %w", i, err)
		}
	}
	return nil
}

// Condition 条件（可以是比较条件或逻辑条件）
type Condition struct {
	Comparison *ComparisonCondition `json:"comparison,omitempty"`
	Logical    *LogicalCondition    `json:"logical,omitempty"`
}

func (c *Condition) String() string {
	if c.Comparison != nil {
		return c.Comparison.String()
	}
	if c.Logical != nil {
		return c.Logical.String()
	}
	return ""
}

func (c *Condition) Validate() error {
	if c.Comparison == nil && c.Logical == nil {
		return fmt.Errorf("condition must have either comparison or logical")
	}
	if c.Comparison != nil && c.Logical != nil {
		return fmt.Errorf("condition cannot have both comparison and logical")
	}

	if c.Comparison != nil {
		return c.Comparison.Validate()
	}
	return c.Logical.Validate()
}

// Action 动作
type Action struct {
	Type       string                 `json:"type"`       // 动作类型
	Parameters map[string]interface{} `json:"parameters"` // 动作参数
}

func (a *Action) Validate() error {
	if a.Type == "" {
		return fmt.Errorf("action type is required")
	}
	return nil
}

// RuleDSL 规则DSL定义
type RuleDSL struct {
	ID          string                 `json:"id"`                    // 规则ID
	Name        string                 `json:"name"`                  // 规则名称
	Description string                 `json:"description,omitempty"` // 规则描述
	Version     string                 `json:"version"`               // 版本号
	Priority    int                    `json:"priority"`              // 优先级（1-100，数字越大优先级越高）
	Enabled     bool                   `json:"enabled"`               // 是否启用
	Tags        []string               `json:"tags,omitempty"`        // 标签
	Condition   Condition              `json:"condition"`             // 条件
	Actions     []Action               `json:"actions"`               // 动作列表
	Metadata    map[string]interface{} `json:"metadata,omitempty"`    // 元数据
	CreatedAt   time.Time              `json:"created_at"`            // 创建时间
	UpdatedAt   time.Time              `json:"updated_at"`            // 更新时间
	CreatedBy   string                 `json:"created_by"`            // 创建人
	UpdatedBy   string                 `json:"updated_by"`            // 更新人
}

func (r *RuleDSL) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

func (r *RuleDSL) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("rule id is required")
	}
	if r.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if r.Version == "" {
		return fmt.Errorf("rule version is required")
	}
	if r.Priority < 1 || r.Priority > 100 {
		return fmt.Errorf("priority must be between 1 and 100")
	}

	if err := r.Condition.Validate(); err != nil {
		return fmt.Errorf("invalid condition: %w", err)
	}

	for i, action := range r.Actions {
		if err := action.Validate(); err != nil {
			return fmt.Errorf("invalid action %d: %w", i, err)
		}
	}

	return nil
}

// ToJSON 转换为JSON字符串
func (r *RuleDSL) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("failed to marshal rule: %w", err)
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析
func FromJSON(jsonStr string) (*RuleDSL, error) {
	var rule RuleDSL
	if err := json.Unmarshal([]byte(jsonStr), &rule); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
	}
	if err := rule.Validate(); err != nil {
		return nil, fmt.Errorf("invalid rule: %w", err)
	}
	return &rule, nil
}

// NewRuleDSL 创建新的规则DSL
func NewRuleDSL(id, name, version string) *RuleDSL {
	now := time.Now()
	return &RuleDSL{
		ID:        id,
		Name:      name,
		Version:   version,
		Priority:  50,
		Enabled:   true,
		Tags:      []string{},
		Actions:   []Action{},
		Metadata:  make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddTag 添加标签
func (r *RuleDSL) AddTag(tag string) {
	r.Tags = append(r.Tags, tag)
}

// RemoveTag 移除标签
func (r *RuleDSL) RemoveTag(tag string) {
	for i, t := range r.Tags {
		if t == tag {
			r.Tags = append(r.Tags[:i], r.Tags[i+1:]...)
			break
		}
	}
}

// AddAction 添加动作
func (r *RuleDSL) AddAction(actionType string, params map[string]interface{}) {
	r.Actions = append(r.Actions, Action{
		Type:       actionType,
		Parameters: params,
	})
}

// SetCondition 设置条件
func (r *RuleDSL) SetCondition(condition Condition) {
	r.Condition = condition
}

// SetComparisonCondition 设置比较条件
func (r *RuleDSL) SetComparisonCondition(pointID string, op Operator, thresholdType ThresholdType, thresholdValue float64) {
	r.Condition = Condition{
		Comparison: &ComparisonCondition{
			Left: ValueExpression{
				PointID: pointID,
			},
			Operator: op,
			Right: Threshold{
				Type:  thresholdType,
				Value: thresholdValue,
			},
		},
	}
}

// SetWindowCondition 设置窗口函数条件
func (r *RuleDSL) SetWindowCondition(pointID string, function WindowFunction, windowSize time.Duration, op Operator, thresholdType ThresholdType, thresholdValue float64) {
	r.Condition = Condition{
		Comparison: &ComparisonCondition{
			Left: ValueExpression{
				PointID:    pointID,
				Function:   function,
				WindowSize: windowSize,
			},
			Operator: op,
			Right: Threshold{
				Type:  thresholdType,
				Value: thresholdValue,
			},
		},
	}
}

// 辅助函数
func isValidOperator(op Operator) bool {
	switch op {
	case OpGT, OpLT, OpGTE, OpLTE, OpEQ, OpNE:
		return true
	default:
		return false
	}
}

func isValidLogicalOperator(op LogicalOperator) bool {
	switch op {
	case LogicalAND, LogicalOR, LogicalNOT:
		return true
	default:
		return false
	}
}
