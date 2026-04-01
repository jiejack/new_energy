package formula

import (
	"fmt"
	"math"
)

// OperatorType 运算符类型
type OperatorType int

const (
	OpTypeArithmetic OperatorType = iota // 算术运算符
	OpTypeComparison                     // 比较运算符
	OpTypeLogical                        // 逻辑运算符
	OpTypeBitwise                        // 位运算符
	OpTypeAssignment                     // 赋值运算符
)

// OperatorInfo 运算符信息
type OperatorInfo struct {
	Symbol      string       // 运算符符号
	Name        string       // 运算符名称
	Type        OperatorType // 运算符类型
	Precedence  int          // 优先级（数值越大优先级越高）
	Associativity string     // 结合性：left, right
	IsUnary     bool         // 是否为一元运算符
	IsBinary    bool         // 是否为二元运算符
}

// OperatorRegistry 运算符注册表
type OperatorRegistry struct {
	operators map[string]*OperatorInfo
}

// NewOperatorRegistry 创建运算符注册表
func NewOperatorRegistry() *OperatorRegistry {
	registry := &OperatorRegistry{
		operators: make(map[string]*OperatorInfo),
	}

	// 注册默认运算符
	registry.registerDefaultOperators()

	return registry
}

// Register 注册运算符
func (r *OperatorRegistry) Register(info *OperatorInfo) {
	r.operators[info.Symbol] = info
}

// Get 获取运算符信息
func (r *OperatorRegistry) Get(symbol string) (*OperatorInfo, bool) {
	info, exists := r.operators[symbol]
	return info, exists
}

// GetByPrecedence 按优先级获取运算符
func (r *OperatorRegistry) GetByPrecedence() []*OperatorInfo {
	result := make([]*OperatorInfo, 0, len(r.operators))
	for _, info := range r.operators {
		result = append(result, info)
	}

	// 按优先级排序（从高到低）
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Precedence > result[i].Precedence {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// registerDefaultOperators 注册默认运算符
func (r *OperatorRegistry) registerDefaultOperators() {
	// ==================== 算术运算符 ====================
	r.Register(&OperatorInfo{
		Symbol:       "+",
		Name:         "addition",
		Type:         OpTypeArithmetic,
		Precedence:   4,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       "-",
		Name:         "subtraction",
		Type:         OpTypeArithmetic,
		Precedence:   4,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       "*",
		Name:         "multiplication",
		Type:         OpTypeArithmetic,
		Precedence:   5,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       "/",
		Name:         "division",
		Type:         OpTypeArithmetic,
		Precedence:   5,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       "%",
		Name:         "modulo",
		Type:         OpTypeArithmetic,
		Precedence:   5,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       "^",
		Name:         "power",
		Type:         OpTypeArithmetic,
		Precedence:   6,
		Associativity: "right",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       "**",
		Name:         "power_alt",
		Type:         OpTypeArithmetic,
		Precedence:   6,
		Associativity: "right",
		IsBinary:     true,
	})

	// 一元运算符
	r.Register(&OperatorInfo{
		Symbol:       "+",
		Name:         "unary_plus",
		Type:         OpTypeArithmetic,
		Precedence:   7,
		Associativity: "right",
		IsUnary:      true,
	})

	r.Register(&OperatorInfo{
		Symbol:       "-",
		Name:         "unary_minus",
		Type:         OpTypeArithmetic,
		Precedence:   7,
		Associativity: "right",
		IsUnary:      true,
	})

	// ==================== 比较运算符 ====================
	r.Register(&OperatorInfo{
		Symbol:       "==",
		Name:         "equal",
		Type:         OpTypeComparison,
		Precedence:   2,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       "!=",
		Name:         "not_equal",
		Type:         OpTypeComparison,
		Precedence:   2,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       ">",
		Name:         "greater_than",
		Type:         OpTypeComparison,
		Precedence:   3,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       "<",
		Name:         "less_than",
		Type:         OpTypeComparison,
		Precedence:   3,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       ">=",
		Name:         "greater_than_or_equal",
		Type:         OpTypeComparison,
		Precedence:   3,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       "<=",
		Name:         "less_than_or_equal",
		Type:         OpTypeComparison,
		Precedence:   3,
		Associativity: "left",
		IsBinary:     true,
	})

	// ==================== 逻辑运算符 ====================
	r.Register(&OperatorInfo{
		Symbol:       "&&",
		Name:         "logical_and",
		Type:         OpTypeLogical,
		Precedence:   1,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       "||",
		Name:         "logical_or",
		Type:         OpTypeLogical,
		Precedence:   0,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       "!",
		Name:         "logical_not",
		Type:         OpTypeLogical,
		Precedence:   7,
		Associativity: "right",
		IsUnary:      true,
	})

	// ==================== 位运算符 ====================
	r.Register(&OperatorInfo{
		Symbol:       "&",
		Name:         "bitwise_and",
		Type:         OpTypeBitwise,
		Precedence:   2,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       "|",
		Name:         "bitwise_or",
		Type:         OpTypeBitwise,
		Precedence:   1,
		Associativity: "left",
		IsBinary:     true,
	})

	// 注意：^ 已作为幂运算符注册，位异或请使用 xor() 函数

	r.Register(&OperatorInfo{
		Symbol:       "<<",
		Name:         "left_shift",
		Type:         OpTypeBitwise,
		Precedence:   3,
		Associativity: "left",
		IsBinary:     true,
	})

	r.Register(&OperatorInfo{
		Symbol:       ">>",
		Name:         "right_shift",
		Type:         OpTypeBitwise,
		Precedence:   3,
		Associativity: "left",
		IsBinary:     true,
	})
}

// ==================== 运算符优先级常量 ====================

const (
	PrecedenceLogicalOr  = 0  // ||
	PrecedenceLogicalAnd = 1  // &&
	PrecedenceBitwiseOr  = 1  // |
	PrecedenceBitwiseXor = 2  // ^
	PrecedenceBitwiseAnd = 2  // &
	PrecedenceEquality   = 2  // ==, !=
	PrecedenceComparison = 3  // >, <, >=, <=
	PrecedenceShift      = 3  // <<, >>
	PrecedenceAddSub     = 4  // +, -
	PrecedenceMulDiv     = 5  // *, /, %
	PrecedencePower      = 6  // ^, **
	PrecedenceUnary      = 7  // 一元运算符
	PrecedenceCall       = 8  // 函数调用
	PrecedenceMember     = 9  // 成员访问
)

// ==================== 运算符求值函数 ====================

// ArithmeticOperator 算术运算符求值
type ArithmeticOperator struct{}

// Evaluate 执行算术运算
func (op *ArithmeticOperator) Evaluate(operator string, left, right interface{}) (interface{}, error) {
	leftNum, err := toFloat64(left)
	if err != nil {
		return nil, fmt.Errorf("left operand: %w", err)
	}

	rightNum, err := toFloat64(right)
	if err != nil {
		return nil, fmt.Errorf("right operand: %w", err)
	}

	switch operator {
	case "+":
		return leftNum + rightNum, nil
	case "-":
		return leftNum - rightNum, nil
	case "*":
		return leftNum * rightNum, nil
	case "/":
		if rightNum == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return leftNum / rightNum, nil
	case "%":
		if rightNum == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return math.Mod(leftNum, rightNum), nil
	case "^", "**":
		return math.Pow(leftNum, rightNum), nil
	default:
		return nil, fmt.Errorf("unknown arithmetic operator: %s", operator)
	}
}

// ComparisonOperator 比较运算符求值
type ComparisonOperator struct{}

// Evaluate 执行比较运算
func (op *ComparisonOperator) Evaluate(operator string, left, right interface{}) (bool, error) {
	return compareValues(operator, left, right)
}

// LogicalOperator 逻辑运算符求值
type LogicalOperator struct{}

// Evaluate 执行逻辑运算
func (op *LogicalOperator) Evaluate(operator string, operands ...interface{}) (bool, error) {
	switch operator {
	case "&&":
		if len(operands) != 2 {
			return false, fmt.Errorf("&& requires 2 operands")
		}
		left, err := toBool(operands[0])
		if err != nil {
			return false, err
		}
		if !left {
			return false, nil
		}
		return toBool(operands[1])

	case "||":
		if len(operands) != 2 {
			return false, fmt.Errorf("|| requires 2 operands")
		}
		left, err := toBool(operands[0])
		if err != nil {
			return false, err
		}
		if left {
			return true, nil
		}
		return toBool(operands[1])

	case "!":
		if len(operands) != 1 {
			return false, fmt.Errorf("! requires 1 operand")
		}
		val, err := toBool(operands[0])
		if err != nil {
			return false, err
		}
		return !val, nil

	default:
		return false, fmt.Errorf("unknown logical operator: %s", operator)
	}
}

// BitwiseOperator 位运算符求值
type BitwiseOperator struct{}

// Evaluate 执行位运算
func (op *BitwiseOperator) Evaluate(operator string, left, right interface{}) (interface{}, error) {
	leftInt, err := toInt64(left)
	if err != nil {
		return nil, fmt.Errorf("left operand: %w", err)
	}

	rightInt, err := toInt64(right)
	if err != nil {
		return nil, fmt.Errorf("right operand: %w", err)
	}

	switch operator {
	case "&":
		return float64(leftInt & rightInt), nil
	case "|":
		return float64(leftInt | rightInt), nil
	case "^":
		return float64(leftInt ^ rightInt), nil
	case "<<":
		if rightInt < 0 {
			return nil, fmt.Errorf("negative shift count")
		}
		return float64(leftInt << uint(rightInt)), nil
	case ">>":
		if rightInt < 0 {
			return nil, fmt.Errorf("negative shift count")
		}
		return float64(leftInt >> uint(rightInt)), nil
	default:
		return nil, fmt.Errorf("unknown bitwise operator: %s", operator)
	}
}

// ==================== 运算符优先级处理器 ====================

// PrecedenceHandler 优先级处理器
type PrecedenceHandler struct {
	registry *OperatorRegistry
}

// NewPrecedenceHandler 创建优先级处理器
func NewPrecedenceHandler() *PrecedenceHandler {
	return &PrecedenceHandler{
		registry: NewOperatorRegistry(),
	}
}

// GetPrecedence 获取运算符优先级
func (h *PrecedenceHandler) GetPrecedence(operator string) int {
	info, exists := h.registry.Get(operator)
	if !exists {
		return -1
	}
	return info.Precedence
}

// IsLeftAssociative 判断是否左结合
func (h *PrecedenceHandler) IsLeftAssociative(operator string) bool {
	info, exists := h.registry.Get(operator)
	if !exists {
		return true
	}
	return info.Associativity == "left"
}

// IsRightAssociative 判断是否右结合
func (h *PrecedenceHandler) IsRightAssociative(operator string) bool {
	info, exists := h.registry.Get(operator)
	if !exists {
		return false
	}
	return info.Associativity == "right"
}

// ComparePrecedence 比较两个运算符的优先级
// 返回值: 1 表示 op1 优先级更高, -1 表示 op2 优先级更高, 0 表示相等
func (h *PrecedenceHandler) ComparePrecedence(op1, op2 string) int {
	p1 := h.GetPrecedence(op1)
	p2 := h.GetPrecedence(op2)

	if p1 > p2 {
		return 1
	} else if p1 < p2 {
		return -1
	}
	return 0
}

// ShouldReduce 判断是否应该归约
// 在解析表达式时，判断是否应该将栈顶运算符与当前运算符进行归约
func (h *PrecedenceHandler) ShouldReduce(stackOp, currentOp string) bool {
	cmp := h.ComparePrecedence(stackOp, currentOp)

	if cmp > 0 {
		return true
	}

	if cmp == 0 && h.IsLeftAssociative(stackOp) {
		return true
	}

	return false
}

// ==================== 表达式求值器 ====================

// ExpressionEvaluator 表达式求值器
type ExpressionEvaluator struct {
	arithmetic  *ArithmeticOperator
	comparison  *ComparisonOperator
	logical     *LogicalOperator
	bitwise     *BitwiseOperator
	precedence  *PrecedenceHandler
}

// NewExpressionEvaluator 创建表达式求值器
func NewExpressionEvaluator() *ExpressionEvaluator {
	return &ExpressionEvaluator{
		arithmetic:  &ArithmeticOperator{},
		comparison:  &ComparisonOperator{},
		logical:     &LogicalOperator{},
		bitwise:     &BitwiseOperator{},
		precedence:  NewPrecedenceHandler(),
	}
}

// EvaluateBinary 执行二元运算
func (e *ExpressionEvaluator) EvaluateBinary(operator string, left, right interface{}) (interface{}, error) {
	info, exists := e.precedence.registry.Get(operator)
	if !exists {
		return nil, fmt.Errorf("unknown operator: %s", operator)
	}

	switch info.Type {
	case OpTypeArithmetic:
		return e.arithmetic.Evaluate(operator, left, right)
	case OpTypeComparison:
		return e.comparison.Evaluate(operator, left, right)
	case OpTypeLogical:
		return e.logical.Evaluate(operator, left, right)
	case OpTypeBitwise:
		return e.bitwise.Evaluate(operator, left, right)
	default:
		return nil, fmt.Errorf("unsupported operator type: %v", info.Type)
	}
}

// EvaluateUnary 执行一元运算
func (e *ExpressionEvaluator) EvaluateUnary(operator string, operand interface{}) (interface{}, error) {
	switch operator {
	case "-":
		num, err := toFloat64(operand)
		if err != nil {
			return nil, err
		}
		return -num, nil
	case "+":
		return toFloat64(operand)
	case "!":
		b, err := toBool(operand)
		if err != nil {
			return nil, err
		}
		return !b, nil
	default:
		return nil, fmt.Errorf("unknown unary operator: %s", operator)
	}
}

// ==================== 运算符工具函数 ====================

// IsOperator 判断是否为运算符
func IsOperator(token string) bool {
	operators := []string{
		"+", "-", "*", "/", "%", "^", "**",
		"==", "!=", ">", "<", ">=", "<=",
		"&&", "||", "!",
		"&", "|", "<<", ">>",
	}

	for _, op := range operators {
		if token == op {
			return true
		}
	}
	return false
}

// IsUnaryOperator 判断是否可以作为一元运算符
func IsUnaryOperator(token string) bool {
	return token == "+" || token == "-" || token == "!"
}

// IsBinaryOperator 判断是否可以作为二元运算符
func IsBinaryOperator(token string) bool {
	return IsOperator(token) && !IsUnaryOnly(token)
}

// IsUnaryOnly 判断是否只能作为一元运算符
func IsUnaryOnly(token string) bool {
	return false // 所有一元运算符都可以作为二元运算符（除了 !）
}

// GetOperatorType 获取运算符类型
func GetOperatorType(token string) OperatorType {
	switch token {
	case "+", "-", "*", "/", "%", "^", "**":
		return OpTypeArithmetic
	case "==", "!=", ">", "<", ">=", "<=":
		return OpTypeComparison
	case "&&", "||", "!":
		return OpTypeLogical
	case "&", "|", "<<", ">>":
		return OpTypeBitwise
	default:
		return OpTypeArithmetic
	}
}

// ==================== 运算符验证 ====================

// OperatorValidator 运算符验证器
type OperatorValidator struct {
	registry *OperatorRegistry
}

// NewOperatorValidator 创建运算符验证器
func NewOperatorValidator() *OperatorValidator {
	return &OperatorValidator{
		registry: NewOperatorRegistry(),
	}
}

// ValidateBinaryOperation 验证二元运算
func (v *OperatorValidator) ValidateBinaryOperation(operator string, left, right interface{}) error {
	info, exists := v.registry.Get(operator)
	if !exists {
		return fmt.Errorf("unknown operator: %s", operator)
	}

	if !info.IsBinary {
		return fmt.Errorf("operator %s is not a binary operator", operator)
	}

	// 检查操作数类型
	switch info.Type {
	case OpTypeArithmetic:
		if !isNumeric(left) || !isNumeric(right) {
			return fmt.Errorf("arithmetic operators require numeric operands")
		}
		if operator == "/" || operator == "%" {
			rightNum, _ := toFloat64(right)
			if rightNum == 0 {
				return fmt.Errorf("division by zero")
			}
		}

	case OpTypeBitwise:
		if !isNumeric(left) || !isNumeric(right) {
			return fmt.Errorf("bitwise operators require numeric operands")
		}

	case OpTypeComparison:
		// 比较运算符可以比较不同类型

	case OpTypeLogical:
		// 逻辑运算符可以处理任何类型
	}

	return nil
}

// ValidateUnaryOperation 验证一元运算
func (v *OperatorValidator) ValidateUnaryOperation(operator string, operand interface{}) error {
	info, exists := v.registry.Get(operator)
	if !exists {
		return fmt.Errorf("unknown operator: %s", operator)
	}

	if !info.IsUnary {
		return fmt.Errorf("operator %s is not a unary operator", operator)
	}

	switch operator {
	case "-", "+":
		if !isNumeric(operand) {
			return fmt.Errorf("unary %s requires numeric operand", operator)
		}
	case "!":
		// 逻辑非可以处理任何类型
	}

	return nil
}

// isNumeric 判断是否为数值类型
func isNumeric(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	default:
		_, err := toFloat64(v)
		return err == nil
	}
}

// ==================== 运算符优先级表 ====================

// OperatorPrecedenceTable 运算符优先级表
var OperatorPrecedenceTable = map[string]int{
	// 逻辑运算符（最低优先级）
	"||": 0,
	"&&": 1,

	// 位运算符
	"|": 1,
	"&": 2,

	// 比较运算符
	"==": 2,
	"!=": 2,
	">":  3,
	"<":  3,
	">=": 3,
	"<=": 3,

	// 位移运算符
	"<<": 3,
	">>": 3,

	// 算术运算符
	"+": 4,
	"-": 4,
	"*": 5,
	"/": 5,
	"%": 5,

	// 幂运算（右结合）
	"^":  6,
	"**": 6,

	// 一元运算符（最高优先级）
	"unary+": 7,
	"unary-": 7,
	"!":      7,
}

// GetPrecedenceFromTable 从优先级表获取优先级
func GetPrecedenceFromTable(operator string) int {
	if precedence, exists := OperatorPrecedenceTable[operator]; exists {
		return precedence
	}
	return -1
}

// ==================== 运算符结合性表 ====================

// OperatorAssociativityTable 运算符结合性表
var OperatorAssociativityTable = map[string]string{
	// 左结合运算符
	"||": "left",
	"&&": "left",
	"|":  "left",
	"^":  "left",
	"&":  "left",
	"==": "left",
	"!=": "left",
	">":  "left",
	"<":  "left",
	">=": "left",
	"<=": "left",
	"<<": "left",
	">>": "left",
	"+":  "left",
	"-":  "left",
	"*":  "left",
	"/":  "left",
	"%":  "left",

	// 右结合运算符
	"**":     "right",
	"unary+": "right",
	"unary-": "right",
	"!":      "right",
}

// GetAssociativityFromTable 从结合性表获取结合性
func GetAssociativityFromTable(operator string) string {
	if associativity, exists := OperatorAssociativityTable[operator]; exists {
		return associativity
	}
	return "left" // 默认左结合
}
