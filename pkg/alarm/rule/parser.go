package rule

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TokenType 词法单元类型
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenIdentifier  // 标识符
	TokenNumber      // 数字
	TokenString      // 字符串
	TokenOperator    // 比较运算符
	TokenLogicalOp   // 逻辑运算符
	TokenLParen      // (
	TokenRParen      // )
	TokenComma       // ,
	TokenDuration    // 时间间隔
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
	input   string
	pos     int
	start   int
	tokens  []Token
	current Token
}

// NewLexer 创建词法分析器
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		tokens: make([]Token, 0),
	}
}

// Lex 执行词法分析
func (l *Lexer) Lex() ([]Token, error) {
	for l.pos < len(l.input) {
		l.skipWhitespace()
		if l.pos >= len(l.input) {
			break
		}

		ch := l.input[l.pos]

		switch {
		case isLetter(ch):
			l.readIdentifier()
		case isDigit(ch) || (ch == '-' && l.pos+1 < len(l.input) && isDigit(l.input[l.pos+1])):
			l.readNumber()
		case ch == '"':
			if err := l.readString(); err != nil {
				return nil, err
			}
		case ch == '(':
			l.tokens = append(l.tokens, Token{TokenLParen, "(", l.pos})
			l.pos++
		case ch == ')':
			l.tokens = append(l.tokens, Token{TokenRParen, ")", l.pos})
			l.pos++
		case ch == ',':
			l.tokens = append(l.tokens, Token{TokenComma, ",", l.pos})
			l.pos++
		case ch == '>' || ch == '<' || ch == '=' || ch == '!':
			l.readOperator()
		default:
			return nil, fmt.Errorf("unexpected character '%c' at position %d", ch, l.pos)
		}
	}

	l.tokens = append(l.tokens, Token{TokenEOF, "", l.pos})
	return l.tokens, nil
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && isWhitespace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *Lexer) readIdentifier() {
	l.start = l.pos
	for l.pos < len(l.input) && (isLetter(l.input[l.pos]) || isDigit(l.input[l.pos]) || l.input[l.pos] == '_' || l.input[l.pos] == '-') {
		l.pos++
	}

	value := l.input[l.start:l.pos]

	// 检查是否为逻辑运算符
	if value == "AND" || value == "OR" || value == "NOT" {
		l.tokens = append(l.tokens, Token{TokenLogicalOp, value, l.start})
		return
	}

	// 检查是否为时间间隔（如 5m, 1h, 30s）
	if l.pos < len(l.input) && (l.input[l.pos] == 's' || l.input[l.pos] == 'm' || l.input[l.pos] == 'h') {
		duration := value + string(l.input[l.pos])
		l.pos++
		l.tokens = append(l.tokens, Token{TokenDuration, duration, l.start})
		return
	}

	l.tokens = append(l.tokens, Token{TokenIdentifier, value, l.start})
}

func (l *Lexer) readNumber() {
	l.start = l.pos
	if l.input[l.pos] == '-' {
		l.pos++
	}

	for l.pos < len(l.input) && (isDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
		l.pos++
	}

	// 检查是否为时间间隔
	if l.pos < len(l.input) && (l.input[l.pos] == 's' || l.input[l.pos] == 'm' || l.input[l.pos] == 'h') {
		duration := l.input[l.start:l.pos] + string(l.input[l.pos])
		l.pos++
		l.tokens = append(l.tokens, Token{TokenDuration, duration, l.start})
		return
	}

	l.tokens = append(l.tokens, Token{TokenNumber, l.input[l.start:l.pos], l.start})
}

func (l *Lexer) readString() error {
	l.pos++ // 跳过开始的引号
	l.start = l.pos

	for l.pos < len(l.input) && l.input[l.pos] != '"' {
		if l.input[l.pos] == '\\' {
			l.pos++ // 跳过转义字符
		}
		l.pos++
	}

	if l.pos >= len(l.input) {
		return fmt.Errorf("unterminated string starting at position %d", l.start-1)
	}

	value := l.input[l.start:l.pos]
	l.tokens = append(l.tokens, Token{TokenString, value, l.start})
	l.pos++ // 跳过结束的引号
	return nil
}

func (l *Lexer) readOperator() {
	l.start = l.pos
	ch := l.input[l.pos]

	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '=' {
		l.pos += 2
		l.tokens = append(l.tokens, Token{TokenOperator, l.input[l.start:l.pos], l.start})
		return
	}

	l.pos++
	l.tokens = append(l.tokens, Token{TokenOperator, string(ch), l.start})
}

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

// Parse 解析DSL字符串
func Parse(input string) (*Condition, error) {
	lexer := NewLexer(input)
	tokens, err := lexer.Lex()
	if err != nil {
		return nil, fmt.Errorf("lexer error: %w", err)
	}

	parser := NewParser(tokens)
	return parser.ParseCondition()
}

// ParseCondition 解析条件
func (p *Parser) ParseCondition() (*Condition, error) {
	// 检查是否为逻辑运算符开头
	if p.current().Type == TokenLogicalOp {
		return p.parseLogicalCondition()
	}

	// 否则解析比较条件
	comparison, err := p.parseComparisonCondition()
	if err != nil {
		return nil, err
	}

	// 检查后面是否跟随逻辑运算符
	if p.current().Type == TokenLogicalOp {
		// 创建逻辑条件
		logical := &LogicalCondition{
			Operator: LogicalOperator(p.current().Value),
			Operands: []Expression{comparison},
		}
		p.advance()

		// 解析右侧条件
		rightCond, err := p.ParseCondition()
		if err != nil {
			return nil, err
		}

		if rightCond.Comparison != nil {
			logical.Operands = append(logical.Operands, rightCond.Comparison)
		} else if rightCond.Logical != nil {
			logical.Operands = append(logical.Operands, rightCond.Logical)
		}

		return &Condition{Logical: logical}, nil
	}

	return &Condition{Comparison: comparison}, nil
}

// parseLogicalCondition 解析逻辑条件
func (p *Parser) parseLogicalCondition() (*Condition, error) {
	logicalOp := LogicalOperator(p.current().Value)
	p.advance()

	var operands []Expression

	if logicalOp == LogicalNOT {
		// NOT 只有一个操作数
		cond, err := p.ParseCondition()
		if err != nil {
			return nil, err
		}
		if cond.Comparison != nil {
			operands = []Expression{cond.Comparison}
		} else if cond.Logical != nil {
			operands = []Expression{cond.Logical}
		}
	} else {
		// AND/OR 需要至少两个操作数
		// 期望左括号
		if p.current().Type != TokenLParen {
			return nil, fmt.Errorf("expected '(' after %s, got %s", logicalOp, p.current().Value)
		}
		p.advance()

		// 解析第一个操作数
		cond, err := p.ParseCondition()
		if err != nil {
			return nil, err
		}
		if cond.Comparison != nil {
			operands = append(operands, cond.Comparison)
		} else if cond.Logical != nil {
			operands = append(operands, cond.Logical)
		}

		// 解析后续操作数
		for p.current().Type == TokenComma {
			p.advance()
			cond, err := p.ParseCondition()
			if err != nil {
				return nil, err
			}
			if cond.Comparison != nil {
				operands = append(operands, cond.Comparison)
			} else if cond.Logical != nil {
				operands = append(operands, cond.Logical)
			}
		}

		// 期望右括号
		if p.current().Type != TokenRParen {
			return nil, fmt.Errorf("expected ')' after operands, got %s", p.current().Value)
		}
		p.advance()
	}

	return &Condition{
		Logical: &LogicalCondition{
			Operator: logicalOp,
			Operands: operands,
		},
	}, nil
}

// parseComparisonCondition 解析比较条件
func (p *Parser) parseComparisonCondition() (*ComparisonCondition, error) {
	// 解析左值
	left, err := p.parseValueExpression()
	if err != nil {
		return nil, fmt.Errorf("failed to parse left value: %w", err)
	}

	// 解析运算符
	if p.current().Type != TokenOperator {
		return nil, fmt.Errorf("expected operator, got %s", p.current().Value)
	}
	op := Operator(p.current().Value)
	p.advance()

	// 解析右值（阈值）
	right, err := p.parseThreshold()
	if err != nil {
		return nil, fmt.Errorf("failed to parse threshold: %w", err)
	}

	return &ComparisonCondition{
		Left:     *left,
		Operator: op,
		Right:    *right,
	}, nil
}

// parseValueExpression 解析值表达式
func (p *Parser) parseValueExpression() (*ValueExpression, error) {
	if p.current().Type != TokenIdentifier {
		return nil, fmt.Errorf("expected identifier, got %s", p.current().Value)
	}

	identifier := p.current().Value
	p.advance()

	// 检查是否为窗口函数
	if p.current().Type == TokenLParen {
		// 可能是窗口函数
		if isWindowFunction(identifier) {
			p.advance() // 跳过 (

			// 解析测点ID
			if p.current().Type != TokenIdentifier {
				return nil, fmt.Errorf("expected point_id in window function, got %s", p.current().Value)
			}
			pointID := p.current().Value
			p.advance()

			// 期望逗号
			if p.current().Type != TokenComma {
				return nil, fmt.Errorf("expected ',' after point_id, got %s", p.current().Value)
			}
			p.advance()

			// 解析时间窗口
			if p.current().Type != TokenDuration && p.current().Type != TokenNumber {
				return nil, fmt.Errorf("expected duration, got %s", p.current().Value)
			}

			var windowSize time.Duration
			var err error

			if p.current().Type == TokenDuration {
				windowSize, err = parseDuration(p.current().Value)
				if err != nil {
					return nil, fmt.Errorf("invalid duration: %w", err)
				}
			} else {
				// 数字后面可能跟时间单位
				numStr := p.current().Value
				p.advance()
				if p.current().Type != TokenIdentifier {
					return nil, fmt.Errorf("expected time unit after number, got %s", p.current().Value)
				}
				windowSize, err = parseDuration(numStr + p.current().Value)
				if err != nil {
					return nil, fmt.Errorf("invalid duration: %w", err)
				}
			}
			p.advance()

			// 期望右括号
			if p.current().Type != TokenRParen {
				return nil, fmt.Errorf("expected ')' after window function arguments, got %s", p.current().Value)
			}
			p.advance()

			return &ValueExpression{
				PointID:    pointID,
				Function:   WindowFunction(identifier),
				WindowSize: windowSize,
			}, nil
		}
	}

	// 普通测点ID
	return &ValueExpression{
		PointID: identifier,
	}, nil
}

// parseThreshold 解析阈值
func (p *Parser) parseThreshold() (*Threshold, error) {
	// 检查是否为阈值类型前缀
	thresholdType := ThresholdTypeAbsolute

	if p.current().Type == TokenIdentifier {
		if isThresholdType(p.current().Value) {
			thresholdType = ThresholdType(p.current().Value)
			p.advance()

			// 期望左括号
			if p.current().Type != TokenLParen {
				return nil, fmt.Errorf("expected '(' after threshold type, got %s", p.current().Value)
			}
			p.advance()
		}
	}

	// 解析数值
	if p.current().Type != TokenNumber {
		return nil, fmt.Errorf("expected number for threshold value, got %s", p.current().Value)
	}

	value, err := strconv.ParseFloat(p.current().Value, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid threshold value: %w", err)
	}
	p.advance()

	// 如果有阈值类型前缀，期望右括号
	if thresholdType != ThresholdTypeAbsolute {
		if p.current().Type != TokenRParen {
			return nil, fmt.Errorf("expected ')' after threshold value, got %s", p.current().Value)
		}
		p.advance()
	}

	return &Threshold{
		Type:  thresholdType,
		Value: value,
	}, nil
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

// 辅助函数
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isWindowFunction(name string) bool {
	switch name {
	case "avg", "max", "min", "sum", "count":
		return true
	default:
		return false
	}
}

func isThresholdType(name string) bool {
	switch name {
	case "absolute", "percentage", "rate":
		return true
	default:
		return false
	}
}

func parseDuration(s string) (time.Duration, error) {
	// 支持格式: 5s, 10m, 1h, 2h30m
	return time.ParseDuration(strings.ToLower(s))
}

// ParseString 解析字符串形式的DSL
func ParseString(dsl string) (*Condition, error) {
	return Parse(dsl)
}

// ParseRuleDSL 解析完整的规则DSL
func ParseRuleDSL(dsl string) (*RuleDSL, error) {
	// 尝试作为JSON解析
	if strings.HasPrefix(strings.TrimSpace(dsl), "{") {
		return FromJSON(dsl)
	}

	// 否则作为条件表达式解析
	condition, err := Parse(dsl)
	if err != nil {
		return nil, err
	}

	// 创建基础规则
	rule := NewRuleDSL("", "", "1.0")
	rule.Condition = *condition
	return rule, nil
}
