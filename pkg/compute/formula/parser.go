package formula

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// TokenType 词法单元类型
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenNumber      // 数字
	TokenString      // 字符串
	TokenIdentifier  // 标识符（变量名、函数名）
	TokenVariable    // 变量引用 ${point-001}
	TokenOperator    // 运算符 + - * / ^ 等
	TokenLParen      // (
	TokenRParen      // )
	TokenLBracket    // [
	TokenRBracket    // ]
	TokenComma       // ,
	TokenColon       // :
	TokenQuestion    // ?
	TokenSemicolon   // ;
)

// Token 词法单元
type Token struct {
	Type  TokenType
	Value string
	Pos   int // 位置
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%d, %q, %d)", t.Type, t.Value, t.Pos)
}

// Lexer 词法分析器
type Lexer struct {
	input string
	pos   int
	start int
}

// NewLexer 创建词法分析器
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
	}
}

// Lex 执行词法分析，返回所有token
func (l *Lexer) Lex() ([]Token, error) {
	tokens := make([]Token, 0)

	for l.pos < len(l.input) {
		l.skipWhitespace()
		if l.pos >= len(l.input) {
			break
		}

		ch := l.input[l.pos]

		switch {
		case isDigit(ch) || (ch == '.' && l.pos+1 < len(l.input) && isDigit(l.input[l.pos+1])):
			token, err := l.readNumber()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)

		case isLetter(ch) || ch == '_':
			token := l.readIdentifier()
			tokens = append(tokens, token)

		case ch == '$' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '{':
			token, err := l.readVariable()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)

		case ch == '"':
			token, err := l.readString()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)

		case ch == '\'':
			token, err := l.readStringSingle()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)

		case ch == '(':
			tokens = append(tokens, Token{TokenLParen, "(", l.pos})
			l.pos++

		case ch == ')':
			tokens = append(tokens, Token{TokenRParen, ")", l.pos})
			l.pos++

		case ch == '[':
			tokens = append(tokens, Token{TokenLBracket, "[", l.pos})
			l.pos++

		case ch == ']':
			tokens = append(tokens, Token{TokenRBracket, "]", l.pos})
			l.pos++

		case ch == ',':
			tokens = append(tokens, Token{TokenComma, ",", l.pos})
			l.pos++

		case ch == ':':
			tokens = append(tokens, Token{TokenColon, ":", l.pos})
			l.pos++

		case ch == '?':
			tokens = append(tokens, Token{TokenQuestion, "?", l.pos})
			l.pos++

		case ch == ';':
			tokens = append(tokens, Token{TokenSemicolon, ";", l.pos})
			l.pos++

		case isOperator(ch):
			token := l.readOperator()
			tokens = append(tokens, token)

		default:
			return nil, fmt.Errorf("unexpected character '%c' at position %d", ch, l.pos)
		}
	}

	tokens = append(tokens, Token{TokenEOF, "", l.pos})
	return tokens, nil
}

// skipWhitespace 跳过空白字符
func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && isWhitespace(l.input[l.pos]) {
		l.pos++
	}
}

// readNumber 读取数字
func (l *Lexer) readNumber() (Token, error) {
	l.start = l.pos

	// 处理负号
	if l.pos < len(l.input) && l.input[l.pos] == '-' {
		l.pos++
	}

	// 整数部分
	for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
		l.pos++
	}

	// 小数部分
	if l.pos < len(l.input) && l.input[l.pos] == '.' {
		l.pos++
		for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
			l.pos++
		}
	}

	// 科学计数法
	if l.pos < len(l.input) && (l.input[l.pos] == 'e' || l.input[l.pos] == 'E') {
		l.pos++
		if l.pos < len(l.input) && (l.input[l.pos] == '+' || l.input[l.pos] == '-') {
			l.pos++
		}
		for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
			l.pos++
		}
	}

	value := l.input[l.start:l.pos]
	return Token{TokenNumber, value, l.start}, nil
}

// readIdentifier 读取标识符
func (l *Lexer) readIdentifier() Token {
	l.start = l.pos
	for l.pos < len(l.input) && (isLetter(l.input[l.pos]) || isDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
		l.pos++
	}
	value := l.input[l.start:l.pos]
	return Token{TokenIdentifier, value, l.start}
}

// readVariable 读取变量引用 ${...}
func (l *Lexer) readVariable() (Token, error) {
	l.start = l.pos
	l.pos += 2 // 跳过 ${

	// 读取到 }
	for l.pos < len(l.input) && l.input[l.pos] != '}' {
		l.pos++
	}

	if l.pos >= len(l.input) {
		return Token{}, fmt.Errorf("unterminated variable reference at position %d", l.start)
	}

	value := l.input[l.start+2 : l.pos]
	l.pos++ // 跳过 }
	return Token{TokenVariable, value, l.start}, nil
}

// readString 读取双引号字符串
func (l *Lexer) readString() (Token, error) {
	l.pos++ // 跳过开始的引号
	l.start = l.pos

	var sb strings.Builder
	for l.pos < len(l.input) && l.input[l.pos] != '"' {
		if l.input[l.pos] == '\\' && l.pos+1 < len(l.input) {
			l.pos++
			switch l.input[l.pos] {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			default:
				sb.WriteByte(l.input[l.pos])
			}
		} else {
			sb.WriteByte(l.input[l.pos])
		}
		l.pos++
	}

	if l.pos >= len(l.input) {
		return Token{}, fmt.Errorf("unterminated string starting at position %d", l.start-1)
	}

	l.pos++ // 跳过结束的引号
	return Token{TokenString, sb.String(), l.start}, nil
}

// readStringSingle 读取单引号字符串
func (l *Lexer) readStringSingle() (Token, error) {
	l.pos++ // 跳过开始的引号
	l.start = l.pos

	var sb strings.Builder
	for l.pos < len(l.input) && l.input[l.pos] != '\'' {
		if l.input[l.pos] == '\\' && l.pos+1 < len(l.input) {
			l.pos++
			switch l.input[l.pos] {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\':
				sb.WriteByte('\\')
			case '\'':
				sb.WriteByte('\'')
			default:
				sb.WriteByte(l.input[l.pos])
			}
		} else {
			sb.WriteByte(l.input[l.pos])
		}
		l.pos++
	}

	if l.pos >= len(l.input) {
		return Token{}, fmt.Errorf("unterminated string starting at position %d", l.start-1)
	}

	l.pos++ // 跳过结束的引号
	return Token{TokenString, sb.String(), l.start}, nil
}

// readOperator 读取运算符
func (l *Lexer) readOperator() Token {
	l.start = l.pos

	// 处理多字符运算符
	if l.pos+1 < len(l.input) {
		twoChar := l.input[l.pos : l.pos+2]
		switch twoChar {
		case "==", "!=", ">=", "<=", "&&", "||", "<<", ">>", "**":
			l.pos += 2
			return Token{TokenOperator, twoChar, l.start}
		}
	}

	ch := l.input[l.pos]
	l.pos++
	return Token{TokenOperator, string(ch), l.start}
}

// 辅助函数
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isOperator(ch byte) bool {
	return ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '%' ||
		ch == '^' || ch == '!' || ch == '>' || ch == '<' || ch == '=' ||
		ch == '&' || ch == '|'
}

// ==================== AST 节点定义 ====================

// NodeType AST节点类型
type NodeType int

const (
	NodeTypeNumber NodeType = iota
	NodeTypeString
	NodeTypeVariable
	NodeTypeBinaryOp
	NodeTypeUnaryOp
	NodeTypeFunctionCall
	NodeTypeConditional
	NodeTypeArray
)

// Node AST节点接口
type Node interface {
	Type() NodeType
	String() string
	Eval(ctx *EvalContext) (interface{}, error)
}

// NumberNode 数字节点
type NumberNode struct {
	Value float64
}

func (n *NumberNode) Type() NodeType { return NodeTypeNumber }
func (n *NumberNode) String() string     { return fmt.Sprintf("%g", n.Value) }
func (n *NumberNode) Eval(ctx *EvalContext) (interface{}, error) {
	return n.Value, nil
}

// StringNode 字符串节点
type StringNode struct {
	Value string
}

func (n *StringNode) Type() NodeType { return NodeTypeString }
func (n *StringNode) String() string     { return fmt.Sprintf("%q", n.Value) }
func (n *StringNode) Eval(ctx *EvalContext) (interface{}, error) {
	return n.Value, nil
}

// VariableNode 变量节点
type VariableNode struct {
	Name string
}

func (n *VariableNode) Type() NodeType { return NodeTypeVariable }
func (n *VariableNode) String() string     { return fmt.Sprintf("${%s}", n.Name) }
func (n *VariableNode) Eval(ctx *EvalContext) (interface{}, error) {
	if ctx == nil || ctx.Variables == nil {
		return nil, fmt.Errorf("variable context is nil")
	}
	val, exists := ctx.Variables[n.Name]
	if !exists {
		return nil, fmt.Errorf("undefined variable: %s", n.Name)
	}
	return val, nil
}

// BinaryOpNode 二元运算节点
type BinaryOpNode struct {
	Operator string
	Left     Node
	Right    Node
}

func (n *BinaryOpNode) Type() NodeType { return NodeTypeBinaryOp }
func (n *BinaryOpNode) String() string {
	return fmt.Sprintf("(%s %s %s)", n.Left.String(), n.Operator, n.Right.String())
}
func (n *BinaryOpNode) Eval(ctx *EvalContext) (interface{}, error) {
	left, err := n.Left.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("left operand error: %w", err)
	}

	right, err := n.Right.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("right operand error: %w", err)
	}

	return evaluateBinaryOp(n.Operator, left, right)
}

// UnaryOpNode 一元运算节点
type UnaryOpNode struct {
	Operator string
	Operand  Node
}

func (n *UnaryOpNode) Type() NodeType { return NodeTypeUnaryOp }
func (n *UnaryOpNode) String() string {
	return fmt.Sprintf("(%s %s)", n.Operator, n.Operand.String())
}
func (n *UnaryOpNode) Eval(ctx *EvalContext) (interface{}, error) {
	operand, err := n.Operand.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("operand error: %w", err)
	}

	return evaluateUnaryOp(n.Operator, operand)
}

// FunctionCallNode 函数调用节点
type FunctionCallNode struct {
	Name      string
	Arguments []Node
}

func (n *FunctionCallNode) Type() NodeType { return NodeTypeFunctionCall }
func (n *FunctionCallNode) String() string {
	args := make([]string, len(n.Arguments))
	for i, arg := range n.Arguments {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", n.Name, strings.Join(args, ", "))
}
func (n *FunctionCallNode) Eval(ctx *EvalContext) (interface{}, error) {
	if ctx == nil || ctx.Functions == nil {
		return nil, fmt.Errorf("function context is nil")
	}

	fn, exists := ctx.Functions[n.Name]
	if !exists {
		return nil, fmt.Errorf("undefined function: %s", n.Name)
	}

	args := make([]interface{}, len(n.Arguments))
	for i, arg := range n.Arguments {
		val, err := arg.Eval(ctx)
		if err != nil {
			return nil, fmt.Errorf("argument %d error: %w", i, err)
		}
		args[i] = val
	}

	return fn(args...)
}

// ConditionalNode 条件运算节点（三元运算符）
type ConditionalNode struct {
	Condition Node
	ThenExpr  Node
	ElseExpr  Node
}

func (n *ConditionalNode) Type() NodeType { return NodeTypeConditional }
func (n *ConditionalNode) String() string {
	return fmt.Sprintf("(%s ? %s : %s)", n.Condition.String(), n.ThenExpr.String(), n.ElseExpr.String())
}
func (n *ConditionalNode) Eval(ctx *EvalContext) (interface{}, error) {
	cond, err := n.Condition.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("condition error: %w", err)
	}

	condBool, ok := cond.(bool)
	if !ok {
		// 尝试将数值转换为布尔值
		if condNum, ok := cond.(float64); ok {
			condBool = condNum != 0
		} else {
			return nil, fmt.Errorf("condition must be boolean, got %T", cond)
		}
	}

	if condBool {
		return n.ThenExpr.Eval(ctx)
	}
	return n.ElseExpr.Eval(ctx)
}

// ArrayNode 数组节点
type ArrayNode struct {
	Elements []Node
}

func (n *ArrayNode) Type() NodeType { return NodeTypeArray }
func (n *ArrayNode) String() string {
	elements := make([]string, len(n.Elements))
	for i, elem := range n.Elements {
		elements[i] = elem.String()
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}
func (n *ArrayNode) Eval(ctx *EvalContext) (interface{}, error) {
	elements := make([]interface{}, len(n.Elements))
	for i, elem := range n.Elements {
		val, err := elem.Eval(ctx)
		if err != nil {
			return nil, fmt.Errorf("element %d error: %w", i, err)
		}
		elements[i] = val
	}
	return elements, nil
}

// ==================== 运算符求值 ====================

// evaluateBinaryOp 执行二元运算
func evaluateBinaryOp(op string, left, right interface{}) (interface{}, error) {
	// 处理逻辑运算符
	if op == "&&" || op == "||" {
		leftBool, err := toBool(left)
		if err != nil {
			return nil, err
		}
		rightBool, err := toBool(right)
		if err != nil {
			return nil, err
		}
		switch op {
		case "&&":
			return leftBool && rightBool, nil
		case "||":
			return leftBool || rightBool, nil
		}
	}

	// 处理比较运算符
	if op == "==" || op == "!=" || op == ">" || op == "<" || op == ">=" || op == "<=" {
		return compareValues(op, left, right)
	}

	// 处理位运算符（注意：^ 是幂运算符，不是位异或）
	if op == "&" || op == "|" || op == "<<" || op == ">>" {
		return bitwiseOp(op, left, right)
	}

	// 处理字符串拼接
	if op == "+" {
		if isString(left) || isString(right) {
			return fmt.Sprintf("%v%v", left, right), nil
		}
	}

	// 数值运算
	leftNum, err := toFloat64(left)
	if err != nil {
		return nil, fmt.Errorf("left operand: %w", err)
	}

	rightNum, err := toFloat64(right)
	if err != nil {
		return nil, fmt.Errorf("right operand: %w", err)
	}

	switch op {
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
		return nil, fmt.Errorf("unknown binary operator: %s", op)
	}
}

// evaluateUnaryOp 执行一元运算
func evaluateUnaryOp(op string, operand interface{}) (interface{}, error) {
	switch op {
	case "-":
		num, err := toFloat64(operand)
		if err != nil {
			return nil, err
		}
		return -num, nil
	case "!":
		b, err := toBool(operand)
		if err != nil {
			return nil, err
		}
		return !b, nil
	case "+":
		return toFloat64(operand)
	default:
		return nil, fmt.Errorf("unknown unary operator: %s", op)
	}
}

// compareValues 比较两个值
func compareValues(op string, left, right interface{}) (bool, error) {
	// 尝试数值比较
	leftNum, leftIsNum := left.(float64)
	rightNum, rightIsNum := right.(float64)

	if leftIsNum && rightIsNum {
		switch op {
		case "==":
			return leftNum == rightNum, nil
		case "!=":
			return leftNum != rightNum, nil
		case ">":
			return leftNum > rightNum, nil
		case "<":
			return leftNum < rightNum, nil
		case ">=":
			return leftNum >= rightNum, nil
		case "<=":
			return leftNum <= rightNum, nil
		}
	}

	// 字符串比较
	leftStr, leftIsStr := left.(string)
	rightStr, rightIsStr := right.(string)

	if leftIsStr && rightIsStr {
		switch op {
		case "==":
			return leftStr == rightStr, nil
		case "!=":
			return leftStr != rightStr, nil
		case ">":
			return leftStr > rightStr, nil
		case "<":
			return leftStr < rightStr, nil
		case ">=":
			return leftStr >= rightStr, nil
		case "<=":
			return leftStr <= rightStr, nil
		}
	}

	// 布尔比较
	leftBool, leftIsBool := left.(bool)
	rightBool, rightIsBool := right.(bool)

	if leftIsBool && rightIsBool {
		switch op {
		case "==":
			return leftBool == rightBool, nil
		case "!=":
			return leftBool != rightBool, nil
		}
	}

	// 通用相等比较
	switch op {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	default:
		return false, fmt.Errorf("cannot compare %T with %T using %s", left, right, op)
	}
}

// bitwiseOp 执行位运算
func bitwiseOp(op string, left, right interface{}) (interface{}, error) {
	leftInt, err := toInt64(left)
	if err != nil {
		return nil, fmt.Errorf("left operand: %w", err)
	}

	rightInt, err := toInt64(right)
	if err != nil {
		return nil, fmt.Errorf("right operand: %w", err)
	}

	switch op {
	case "&":
		return float64(leftInt & rightInt), nil
	case "|":
		return float64(leftInt | rightInt), nil
	case "^":
		return float64(leftInt ^ rightInt), nil
	case "<<":
		// 安全转换：int64到uint，确保右移值在合理范围内
		/* #nosec G115 -- shift值已限制在0-63范围内，转换为uint是安全的 */
		shift := uint(min(max(rightInt, 0), 63))
		return float64(leftInt << shift), nil
	case ">>":
		// 安全转换：int64到uint，确保右移值在合理范围内
		/* #nosec G115 -- shift值已限制在0-63范围内，转换为uint是安全的 */
		shift := uint(min(max(rightInt, 0), 63))
		return float64(leftInt >> shift), nil
	default:
		return nil, fmt.Errorf("unknown bitwise operator: %s", op)
	}
}

// 类型转换辅助函数
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string %q to number", val)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to number", v)
	}
}

func toInt64(v interface{}) (int64, error) {
	f, err := toFloat64(v)
	if err != nil {
		return 0, err
	}
	return int64(f), nil
}

func toBool(v interface{}) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case float64:
		return val != 0, nil
	case float32:
		return val != 0, nil
	case int:
		return val != 0, nil
	case int64:
		return val != 0, nil
	case string:
		return val != "", nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

func isString(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

// ==================== 语法分析器 ====================

// Parser 语法分析器
type Parser struct {
	tokens []Token
	pos    int
}

// NewParser 创建语法分析器
func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

// Parse 解析表达式
func (p *Parser) Parse() (Node, error) {
	return p.parseConditional()
}

// parseConditional 解析条件表达式（三元运算符）
func (p *Parser) parseConditional() (Node, error) {
	cond, err := p.parseLogicalOr()
	if err != nil {
		return nil, err
	}

	if p.current().Type == TokenQuestion {
		p.advance()
		thenExpr, err := p.parseConditional()
		if err != nil {
			return nil, err
		}

		if p.current().Type != TokenColon {
			return nil, fmt.Errorf("expected ':' in conditional expression, got %s", p.current().Value)
		}
		p.advance()

		elseExpr, err := p.parseConditional()
		if err != nil {
			return nil, err
		}

		return &ConditionalNode{
			Condition: cond,
			ThenExpr:  thenExpr,
			ElseExpr:  elseExpr,
		}, nil
	}

	return cond, nil
}

// parseLogicalOr 解析逻辑或运算
func (p *Parser) parseLogicalOr() (Node, error) {
	left, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenOperator && p.current().Value == "||" {
		op := p.current().Value
		p.advance()
		right, err := p.parseLogicalAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Operator: op, Left: left, Right: right}
	}

	return left, nil
}

// parseLogicalAnd 解析逻辑与运算
func (p *Parser) parseLogicalAnd() (Node, error) {
	left, err := p.parseBitwiseOr()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenOperator && p.current().Value == "&&" {
		op := p.current().Value
		p.advance()
		right, err := p.parseBitwiseOr()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Operator: op, Left: left, Right: right}
	}

	return left, nil
}

// parseBitwiseOr 解析位或运算
func (p *Parser) parseBitwiseOr() (Node, error) {
	left, err := p.parseBitwiseXor()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenOperator && p.current().Value == "|" {
		op := p.current().Value
		p.advance()
		right, err := p.parseBitwiseXor()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Operator: op, Left: left, Right: right}
	}

	return left, nil
}

// parseBitwiseXor 解析位异或运算
func (p *Parser) parseBitwiseXor() (Node, error) {
	left, err := p.parseBitwiseAnd()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenOperator && p.current().Value == "^" {
		op := p.current().Value
		p.advance()
		right, err := p.parseBitwiseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Operator: op, Left: left, Right: right}
	}

	return left, nil
}

// parseBitwiseAnd 解析位与运算
func (p *Parser) parseBitwiseAnd() (Node, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenOperator && p.current().Value == "&" {
		op := p.current().Value
		p.advance()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Operator: op, Left: left, Right: right}
	}

	return left, nil
}

// parseEquality 解析相等性运算
func (p *Parser) parseEquality() (Node, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenOperator && (p.current().Value == "==" || p.current().Value == "!=") {
		op := p.current().Value
		p.advance()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Operator: op, Left: left, Right: right}
	}

	return left, nil
}

// parseComparison 解析比较运算
func (p *Parser) parseComparison() (Node, error) {
	left, err := p.parseShift()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenOperator && isComparisonOp(p.current().Value) {
		op := p.current().Value
		p.advance()
		right, err := p.parseShift()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Operator: op, Left: left, Right: right}
	}

	return left, nil
}

// parseShift 解析位移运算
func (p *Parser) parseShift() (Node, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenOperator && (p.current().Value == "<<" || p.current().Value == ">>") {
		op := p.current().Value
		p.advance()
		right, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Operator: op, Left: left, Right: right}
	}

	return left, nil
}

// parseAdditive 解析加减运算
func (p *Parser) parseAdditive() (Node, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenOperator && (p.current().Value == "+" || p.current().Value == "-") {
		op := p.current().Value
		p.advance()
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Operator: op, Left: left, Right: right}
	}

	return left, nil
}

// parseMultiplicative 解析乘除运算
func (p *Parser) parseMultiplicative() (Node, error) {
	left, err := p.parsePower()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenOperator && (p.current().Value == "*" || p.current().Value == "/" || p.current().Value == "%") {
		op := p.current().Value
		p.advance()
		right, err := p.parsePower()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Operator: op, Left: left, Right: right}
	}

	return left, nil
}

// parsePower 解析幂运算
func (p *Parser) parsePower() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenOperator && (p.current().Value == "^" || p.current().Value == "**") {
		op := p.current().Value
		p.advance()
		right, err := p.parsePower() // 右结合
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Operator: op, Left: left, Right: right}
	}

	return left, nil
}

// parseUnary 解析一元运算
func (p *Parser) parseUnary() (Node, error) {
	if p.current().Type == TokenOperator && (p.current().Value == "-" || p.current().Value == "+" || p.current().Value == "!") {
		op := p.current().Value
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryOpNode{Operator: op, Operand: operand}, nil
	}

	return p.parsePrimary()
}

// parsePrimary 解析基本表达式
func (p *Parser) parsePrimary() (Node, error) {
	token := p.current()

	switch token.Type {
	case TokenNumber:
		p.advance()
		value, err := strconv.ParseFloat(token.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", token.Value)
		}
		return &NumberNode{Value: value}, nil

	case TokenString:
		p.advance()
		return &StringNode{Value: token.Value}, nil

	case TokenVariable:
		p.advance()
		return &VariableNode{Name: token.Value}, nil

	case TokenIdentifier:
		p.advance()
		// 检查是否为函数调用
		if p.current().Type == TokenLParen {
			return p.parseFunctionCall(token.Value)
		}
		// 否则作为变量处理
		return &VariableNode{Name: token.Value}, nil

	case TokenLParen:
		p.advance()
		expr, err := p.parseConditional()
		if err != nil {
			return nil, err
		}
		if p.current().Type != TokenRParen {
			return nil, fmt.Errorf("expected ')', got %s", p.current().Value)
		}
		p.advance()
		return expr, nil

	case TokenLBracket:
		return p.parseArray()

	default:
		return nil, fmt.Errorf("unexpected token: %s", token.Value)
	}
}

// parseFunctionCall 解析函数调用
func (p *Parser) parseFunctionCall(name string) (Node, error) {
	p.advance() // 跳过 (

	args := make([]Node, 0)
	for p.current().Type != TokenRParen {
		arg, err := p.parseConditional()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		if p.current().Type == TokenComma {
			p.advance()
		} else if p.current().Type != TokenRParen {
			return nil, fmt.Errorf("expected ',' or ')', got %s", p.current().Value)
		}
	}

	p.advance() // 跳过 )

	return &FunctionCallNode{Name: name, Arguments: args}, nil
}

// parseArray 解析数组
func (p *Parser) parseArray() (Node, error) {
	p.advance() // 跳过 [

	elements := make([]Node, 0)
	for p.current().Type != TokenRBracket {
		elem, err := p.parseConditional()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)

		if p.current().Type == TokenComma {
			p.advance()
		} else if p.current().Type != TokenRBracket {
			return nil, fmt.Errorf("expected ',' or ']', got %s", p.current().Value)
		}
	}

	p.advance() // 跳过 ]

	return &ArrayNode{Elements: elements}, nil
}

// current 获取当前token
func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{TokenEOF, "", p.pos}
	}
	return p.tokens[p.pos]
}

// advance 前进到下一个token
func (p *Parser) advance() {
	p.pos++
}

// isComparisonOp 判断是否为比较运算符
func isComparisonOp(op string) bool {
	return op == ">" || op == "<" || op == ">=" || op == "<="
}

// ==================== 公开接口 ====================

// ParseFormula 解析公式字符串，返回AST
func ParseFormula(input string) (Node, error) {
	lexer := NewLexer(input)
	tokens, err := lexer.Lex()
	if err != nil {
		return nil, fmt.Errorf("lexer error: %w", err)
	}

	parser := NewParser(tokens)
	return parser.Parse()
}

// EvalContext 求值上下文
type EvalContext struct {
	Variables map[string]interface{}
	Functions map[string]Function
}

// Function 函数类型
type Function func(args ...interface{}) (interface{}, error)

// NewEvalContext 创建求值上下文
func NewEvalContext() *EvalContext {
	return &EvalContext{
		Variables: make(map[string]interface{}),
		Functions: make(map[string]Function),
	}
}

// SetVariable 设置变量
func (ctx *EvalContext) SetVariable(name string, value interface{}) {
	ctx.Variables[name] = value
}

// GetVariable 获取变量
func (ctx *EvalContext) GetVariable(name string) (interface{}, bool) {
	val, exists := ctx.Variables[name]
	return val, exists
}

// RegisterFunction 注册函数
func (ctx *EvalContext) RegisterFunction(name string, fn Function) {
	ctx.Functions[name] = fn
}

// Evaluate 直接求值公式
func Evaluate(input string, ctx *EvalContext) (interface{}, error) {
	node, err := ParseFormula(input)
	if err != nil {
		return nil, err
	}
	return node.Eval(ctx)
}
