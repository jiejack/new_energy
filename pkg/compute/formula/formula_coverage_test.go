package formula

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexer_UnterminatedVariable(t *testing.T) {
	lexer := NewLexer("${unterminated")
	_, err := lexer.Lex()
	assert.Error(t, err)
}

func TestLexer_UnterminatedString(t *testing.T) {
	lexer := NewLexer(`"unterminated`)
	_, err := lexer.Lex()
	assert.Error(t, err)
}

func TestLexer_SingleQuotedString(t *testing.T) {
	lexer := NewLexer(`'hello'`)
	tokens, err := lexer.Lex()
	require.NoError(t, err)
	assert.Equal(t, TokenString, tokens[0].Type)
	assert.Equal(t, "hello", tokens[0].Value)
}

func TestLexer_SingleQuotedStringEscapes(t *testing.T) {
	lexer := NewLexer(`'hello\nworld'`)
	tokens, err := lexer.Lex()
	require.NoError(t, err)
	assert.Equal(t, TokenString, tokens[0].Type)
	assert.Contains(t, tokens[0].Value, "\n")
}

func TestLexer_SingleQuotedUnterminated(t *testing.T) {
	lexer := NewLexer(`'unterminated`)
	_, err := lexer.Lex()
	assert.Error(t, err)
}

func TestLexer_UnexpectedCharacter(t *testing.T) {
	lexer := NewLexer("a @ b")
	_, err := lexer.Lex()
	assert.Error(t, err)
}

func TestLexer_StringEscapes(t *testing.T) {
	lexer := NewLexer(`"hello\tworld\n\r\\"`)
	tokens, err := lexer.Lex()
	require.NoError(t, err)
	assert.Equal(t, TokenString, tokens[0].Type)
	assert.Contains(t, tokens[0].Value, "\t")
	assert.Contains(t, tokens[0].Value, "\n")
	assert.Contains(t, tokens[0].Value, "\r")
	assert.Contains(t, tokens[0].Value, "\\")
}

func TestLexer_StringEscapeQuote(t *testing.T) {
	lexer := NewLexer(`"say \"hello\""`)
	tokens, err := lexer.Lex()
	require.NoError(t, err)
	assert.Contains(t, tokens[0].Value, `"hello"`)
}

func TestLexer_SingleQuotedEscapeQuote(t *testing.T) {
	lexer := NewLexer(`'it\'s'`)
	tokens, err := lexer.Lex()
	require.NoError(t, err)
	assert.Contains(t, tokens[0].Value, "'")
}

func TestLexer_SingleQuotedUnknownEscape(t *testing.T) {
	lexer := NewLexer(`'\x'`)
	tokens, err := lexer.Lex()
	require.NoError(t, err)
	assert.Contains(t, tokens[0].Value, "x")
}

func TestLexer_ScientificNotation(t *testing.T) {
	lexer := NewLexer("1.5e10")
	tokens, err := lexer.Lex()
	require.NoError(t, err)
	assert.Equal(t, TokenNumber, tokens[0].Type)
	assert.Equal(t, "1.5e10", tokens[0].Value)
}

func TestLexer_ScientificNotationNegative(t *testing.T) {
	lexer := NewLexer("1.5E-3")
	tokens, err := lexer.Lex()
	require.NoError(t, err)
	assert.Equal(t, TokenNumber, tokens[0].Type)
}

func TestLexer_DotNumber(t *testing.T) {
	lexer := NewLexer(".5")
	tokens, err := lexer.Lex()
	require.NoError(t, err)
	assert.Equal(t, TokenNumber, tokens[0].Type)
}

func TestLexer_AllTokenTypes(t *testing.T) {
	lexer := NewLexer(`a + b * (c - d) / e % f ^ g, h: i ? j ; k[l]`)
	tokens, err := lexer.Lex()
	require.NoError(t, err)
	assert.Greater(t, len(tokens), 15)
}

func TestLexer_TwoCharOperators(t *testing.T) {
	ops := []string{"==", "!=", ">=", "<=", "&&", "||", "<<", ">>", "**"}
	for _, op := range ops {
		lexer := NewLexer(op)
		tokens, err := lexer.Lex()
		require.NoError(t, err, "operator: %s", op)
		assert.Equal(t, TokenOperator, tokens[0].Type, "operator: %s", op)
		assert.Equal(t, op, tokens[0].Value, "operator: %s", op)
	}
}

func TestToken_String(t *testing.T) {
	tok := Token{Type: TokenNumber, Value: "42", Pos: 0}
	s := tok.String()
	assert.Contains(t, s, "42")
}

func TestParser_UnexpectedToken(t *testing.T) {
	_, err := ParseFormula(",")
	assert.Error(t, err)
}

func TestParser_MissingColonInConditional(t *testing.T) {
	_, err := ParseFormula("a > b ? c d")
	assert.Error(t, err)
}

func TestParser_MissingClosingParen(t *testing.T) {
	_, err := ParseFormula("(a + b")
	assert.Error(t, err)
}

func TestParser_MissingCommaInFunction(t *testing.T) {
	_, err := ParseFormula("max(a b)")
	assert.Error(t, err)
}

func TestParser_MissingCommaInArray(t *testing.T) {
	_, err := ParseFormula("[a b]")
	assert.Error(t, err)
}

func TestParser_ArrayExpression(t *testing.T) {
	node, err := ParseFormula("[1, 2, 3]")
	require.NoError(t, err)
	assert.Equal(t, NodeTypeArray, node.Type())
	arr, ok := node.(*ArrayNode)
	require.True(t, ok)
	assert.Len(t, arr.Elements, 3)
}

func TestParser_BitwiseOperations(t *testing.T) {
	node, err := ParseFormula("a & b | c")
	require.NoError(t, err)
	assert.Equal(t, NodeTypeBinaryOp, node.Type())
}

func TestParser_ShiftOperations(t *testing.T) {
	node, err := ParseFormula("a << b >> c")
	require.NoError(t, err)
	assert.Equal(t, NodeTypeBinaryOp, node.Type())
}

func TestParser_PowerRightAssociative(t *testing.T) {
	node, err := ParseFormula("2 ^ 3 ^ 2")
	require.NoError(t, err)
	assert.Equal(t, NodeTypeBinaryOp, node.Type())
}

func TestParser_UnaryPlus(t *testing.T) {
	node, err := ParseFormula("+a")
	require.NoError(t, err)
	assert.Equal(t, NodeTypeUnaryOp, node.Type())
}

func TestASTNode_StringMethods(t *testing.T) {
	numNode := &NumberNode{Value: 42.5}
	assert.Contains(t, numNode.String(), "42.5")

	strNode := &StringNode{Value: "hello"}
	assert.Contains(t, strNode.String(), "hello")

	varNode := &VariableNode{Name: "x"}
	assert.Contains(t, varNode.String(), "x")

	binNode := &BinaryOpNode{Operator: "+", Left: numNode, Right: varNode}
	assert.Contains(t, binNode.String(), "+")

	unaryNode := &UnaryOpNode{Operator: "-", Operand: numNode}
	assert.Contains(t, unaryNode.String(), "-")

	funcNode := &FunctionCallNode{Name: "max", Arguments: []Node{numNode, varNode}}
	assert.Contains(t, funcNode.String(), "max")

	condNode := &ConditionalNode{Condition: varNode, ThenExpr: numNode, ElseExpr: strNode}
	assert.Contains(t, condNode.String(), "?")

	arrNode := &ArrayNode{Elements: []Node{numNode, strNode}}
	assert.Contains(t, arrNode.String(), "[")
}

func TestVariableNode_Eval_NilContext(t *testing.T) {
	node := &VariableNode{Name: "x"}
	_, err := node.Eval(nil)
	assert.Error(t, err)
}

func TestVariableNode_Eval_NilVariables(t *testing.T) {
	node := &VariableNode{Name: "x"}
	ctx := &EvalContext{}
	_, err := node.Eval(ctx)
	assert.Error(t, err)
}

func TestFunctionCallNode_Eval_NilContext(t *testing.T) {
	node := &FunctionCallNode{Name: "max"}
	_, err := node.Eval(nil)
	assert.Error(t, err)
}

func TestFunctionCallNode_Eval_NilFunctions(t *testing.T) {
	node := &FunctionCallNode{Name: "max"}
	ctx := &EvalContext{}
	_, err := node.Eval(ctx)
	assert.Error(t, err)
}

func TestFunctionCallNode_Eval_UndefinedFunction(t *testing.T) {
	node := &FunctionCallNode{Name: "nonexistent"}
	ctx := NewEvalContext()
	_, err := node.Eval(ctx)
	assert.Error(t, err)
}

func TestConditionalNode_Eval_NonBoolCondition(t *testing.T) {
	cond := &NumberNode{Value: 5.0}
	then := &NumberNode{Value: 1.0}
	else_ := &NumberNode{Value: 0.0}
	node := &ConditionalNode{Condition: cond, ThenExpr: then, ElseExpr: else_}
	ctx := NewEvalContext()
	result, err := node.Eval(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1.0, result)
}

func TestConditionalNode_Eval_InvalidConditionType(t *testing.T) {
	cond := &ArrayNode{Elements: []Node{}}
	then := &NumberNode{Value: 1.0}
	else_ := &NumberNode{Value: 0.0}
	node := &ConditionalNode{Condition: cond, ThenExpr: then, ElseExpr: else_}
	ctx := NewEvalContext()
	_, err := node.Eval(ctx)
	assert.Error(t, err)
}

func TestEvaluateBinaryOp_StringConcatenation(t *testing.T) {
	result, err := evaluateBinaryOp("+", "hello", "world")
	require.NoError(t, err)
	assert.Equal(t, "helloworld", result)
}

func TestEvaluateBinaryOp_StringNumberConcat(t *testing.T) {
	result, err := evaluateBinaryOp("+", "val", 42.0)
	require.NoError(t, err)
	assert.Equal(t, "val42", result)
}

func TestEvaluateBinaryOp_NumberStringConcat(t *testing.T) {
	result, err := evaluateBinaryOp("+", 42.0, "val")
	require.NoError(t, err)
	assert.Equal(t, "42val", result)
}

func TestEvaluateBinaryOp_UnknownOperator(t *testing.T) {
	_, err := evaluateBinaryOp("@@", 1.0, 2.0)
	assert.Error(t, err)
}

func TestEvaluateBinaryOp_LeftOperandError(t *testing.T) {
	_, err := evaluateBinaryOp("+", struct{}{}, 2.0)
	assert.Error(t, err)
}

func TestEvaluateBinaryOp_RightOperandError(t *testing.T) {
	_, err := evaluateBinaryOp("+", 1.0, struct{}{})
	assert.Error(t, err)
}

func TestEvaluateUnaryOp_Plus(t *testing.T) {
	result, err := evaluateUnaryOp("+", 5.0)
	require.NoError(t, err)
	assert.Equal(t, 5.0, result)
}

func TestEvaluateUnaryOp_Unknown(t *testing.T) {
	_, err := evaluateUnaryOp("@", 5.0)
	assert.Error(t, err)
}

func TestCompareValues_BooleanEqual(t *testing.T) {
	result, err := compareValues("==", true, true)
	require.NoError(t, err)
	assert.True(t, result)
}

func TestCompareValues_BooleanNotEqual(t *testing.T) {
	result, err := compareValues("!=", true, false)
	require.NoError(t, err)
	assert.True(t, result)
}

func TestCompareValues_MixedTypes(t *testing.T) {
	result, err := compareValues("==", 1.0, "1")
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCompareValues_MixedTypesNotEqual(t *testing.T) {
	result, err := compareValues("!=", 1.0, "1")
	require.NoError(t, err)
	assert.True(t, result)
}

func TestCompareValues_IncompatibleTypes(t *testing.T) {
	_, err := compareValues(">", 1.0, "abc")
	assert.Error(t, err)
}

func TestCompareValues_StringComparison(t *testing.T) {
	result, err := compareValues(">", "b", "a")
	require.NoError(t, err)
	assert.True(t, result)
}

func TestBitwiseOp_ShiftLeft(t *testing.T) {
	result, err := bitwiseOp("<<", 1.0, 4.0)
	require.NoError(t, err)
	assert.Equal(t, float64(16), result)
}

func TestBitwiseOp_ShiftRight(t *testing.T) {
	result, err := bitwiseOp(">>", 16.0, 4.0)
	require.NoError(t, err)
	assert.Equal(t, float64(1), result)
}

func TestBitwiseOp_UnknownOperator(t *testing.T) {
	_, err := bitwiseOp("@@", 1.0, 2.0)
	assert.Error(t, err)
}

func TestBitwiseOp_LeftOperandError(t *testing.T) {
	_, err := bitwiseOp("&", "abc", 2.0)
	assert.Error(t, err)
}

func TestBitwiseOp_RightOperandError(t *testing.T) {
	_, err := bitwiseOp("&", 1.0, "abc")
	assert.Error(t, err)
}

func TestToFloat64_VariousTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
		hasError bool
	}{
		{"float32", float32(3.14), float64(float32(3.14)), false},
		{"int", 42, 42.0, false},
		{"int64", int64(100), 100.0, false},
		{"int32", int32(50), 50.0, false},
		{"uint", uint(10), 10.0, false},
		{"uint64", uint64(200), 200.0, false},
		{"uint32", uint32(30), 30.0, false},
		{"bool_true", true, 1.0, false},
		{"bool_false", false, 0.0, false},
		{"string_number", "42.5", 42.5, false},
		{"string_invalid", "abc", 0, true},
		{"unsupported", struct{}{}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toFloat64(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestToBool_VariousTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
		hasError bool
	}{
		{"float32_nonzero", float32(1.5), true, false},
		{"int_nonzero", 42, true, false},
		{"int64_nonzero", int64(1), true, false},
		{"string_nonempty", "hello", true, false},
		{"string_empty", "", false, false},
		{"unsupported", struct{}{}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toBool(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestResultCache_ExpiredItem(t *testing.T) {
	cache := NewResultCache(100, 1*time.Nanosecond)
	cache.Set("key1", "value1")
	time.Sleep(10 * time.Millisecond)
	_, exists := cache.Get("key1")
	assert.False(t, exists)
}

func TestResultCache_DeleteNonExistent(t *testing.T) {
	cache := NewResultCache(100, 5*time.Minute)
	cache.Delete("nonexistent")
	stats := cache.Stats()
	assert.Equal(t, 0, stats.Size)
}

func TestCompiledFormula_GetFunctions(t *testing.T) {
	executor := NewExecutor(nil)
	compiled, err := executor.Compile("max(a, b) + min(c, d)")
	require.NoError(t, err)
	functions := compiled.GetFunctions()
	assert.Contains(t, functions, "max")
	assert.Contains(t, functions, "min")
}

func TestExecutor_NoCache(t *testing.T) {
	config := &ExecutorConfig{EnableCache: false}
	executor := NewExecutor(config)
	assert.Nil(t, executor.GetCacheStats())
	executor.ClearCache()
}

func TestExecutionContext_ParentLookup(t *testing.T) {
	parent := NewExecutionContext()
	parent.SetVariable("a", 10.0)
	parent.RegisterFunction("fn1", func(args ...interface{}) (interface{}, error) { return nil, nil })
	parent.SetMetadata("key1", "val1")

	child := parent.CreateChild()
	val, exists := child.GetVariable("a")
	assert.True(t, exists)
	assert.Equal(t, 10.0, val)

	fn, exists := child.GetFunction("fn1")
	assert.True(t, exists)
	assert.NotNil(t, fn)

	meta, exists := child.GetMetadata("key1")
	assert.True(t, exists)
	assert.Equal(t, "val1", meta)
}

func TestExecutionContext_ParentVariableOverride(t *testing.T) {
	parent := NewExecutionContext()
	parent.SetVariable("a", 10.0)
	child := parent.CreateChild()
	child.SetVariable("a", 20.0)
	val, exists := child.GetVariable("a")
	assert.True(t, exists)
	assert.Equal(t, 20.0, val)
}

func TestExecutionContext_NoParent(t *testing.T) {
	ctx := NewExecutionContext()
	_, exists := ctx.GetVariable("nonexistent")
	assert.False(t, exists)
	_, exists = ctx.GetFunction("nonexistent")
	assert.False(t, exists)
	_, exists = ctx.GetMetadata("nonexistent")
	assert.False(t, exists)
}

func TestExecutionContext_ToEvalContext(t *testing.T) {
	parent := NewExecutionContext()
	parent.SetVariable("a", 10.0)
	parent.RegisterFunction("fn1", func(args ...interface{}) (interface{}, error) { return nil, nil })

	child := parent.CreateChild()
	child.SetVariable("b", 20.0)

	evalCtx := child.ToEvalContext()
	assert.Equal(t, 10.0, evalCtx.Variables["a"])
	assert.Equal(t, 20.0, evalCtx.Variables["b"])
	assert.NotNil(t, evalCtx.Functions["fn1"])
}

func TestOperatorRegistry_GetByPrecedence(t *testing.T) {
	registry := NewOperatorRegistry()
	ops := registry.GetByPrecedence()
	assert.NotEmpty(t, ops)
	for i := 1; i < len(ops); i++ {
		assert.GreaterOrEqual(t, ops[i-1].Precedence, ops[i].Precedence)
	}
}

func TestOperatorRegistry_GetNonExistent(t *testing.T) {
	registry := NewOperatorRegistry()
	_, exists := registry.Get("@@")
	assert.False(t, exists)
}

func TestPrecedenceHandler_AllMethods(t *testing.T) {
	handler := NewPrecedenceHandler()
	assert.Greater(t, handler.GetPrecedence("*"), handler.GetPrecedence("=="))
	assert.True(t, handler.IsLeftAssociative("*"))
	assert.False(t, handler.IsRightAssociative("*"))
	assert.True(t, handler.IsRightAssociative("^"))
	assert.False(t, handler.IsRightAssociative("*"))
	assert.Equal(t, 1, handler.ComparePrecedence("*", "=="))
	assert.Equal(t, -1, handler.ComparePrecedence("==", "*"))
	assert.Equal(t, 0, handler.ComparePrecedence("==", "!="))
	assert.True(t, handler.ShouldReduce("*", "=="))
	assert.False(t, handler.ShouldReduce("==", "*"))
	assert.Equal(t, -1, handler.GetPrecedence("@@"))
	assert.True(t, handler.IsLeftAssociative("@@"))
	assert.False(t, handler.IsRightAssociative("@@"))
}

func TestExpressionEvaluator_AllTypes(t *testing.T) {
	eval := NewExpressionEvaluator()
	result, err := eval.EvaluateBinary("-", 10.0, 3.0)
	require.NoError(t, err)
	assert.Equal(t, 7.0, result)

	result, err = eval.EvaluateBinary("*", 4.0, 5.0)
	require.NoError(t, err)
	assert.Equal(t, 20.0, result)

	result, err = eval.EvaluateBinary("/", 20.0, 4.0)
	require.NoError(t, err)
	assert.Equal(t, 5.0, result)

	result, err = eval.EvaluateBinary("%", 10.0, 3.0)
	require.NoError(t, err)
	assert.InDelta(t, 1.0, result, 0.001)

	result, err = eval.EvaluateBinary("^", 2.0, 3.0)
	require.NoError(t, err)
	assert.Equal(t, 8.0, result)

	result, err = eval.EvaluateBinary("**", 2.0, 3.0)
	require.NoError(t, err)
	assert.Equal(t, 8.0, result)

	boolResult, err := eval.EvaluateBinary("==", 10.0, 10.0)
	require.NoError(t, err)
	assert.Equal(t, true, boolResult)

	boolResult, err = eval.EvaluateBinary("!=", 10.0, 5.0)
	require.NoError(t, err)
	assert.Equal(t, true, boolResult)

	boolResult, err = eval.EvaluateBinary(">", 10.0, 5.0)
	require.NoError(t, err)
	assert.Equal(t, true, boolResult)

	boolResult, err = eval.EvaluateBinary("<", 5.0, 10.0)
	require.NoError(t, err)
	assert.Equal(t, true, boolResult)

	boolResult, err = eval.EvaluateBinary(">=", 10.0, 10.0)
	require.NoError(t, err)
	assert.Equal(t, true, boolResult)

	boolResult, err = eval.EvaluateBinary("<=", 5.0, 10.0)
	require.NoError(t, err)
	assert.Equal(t, true, boolResult)

	boolResult, err = eval.EvaluateBinary("&&", true, true)
	require.NoError(t, err)
	assert.Equal(t, true, boolResult)

	boolResult, err = eval.EvaluateBinary("||", false, true)
	require.NoError(t, err)
	assert.Equal(t, true, boolResult)

	result, err = eval.EvaluateBinary("&", float64(12), float64(10))
	require.NoError(t, err)
	assert.Equal(t, float64(8), result)

	result, err = eval.EvaluateBinary("|", float64(12), float64(10))
	require.NoError(t, err)
	assert.Equal(t, float64(14), result)

	result, err = eval.EvaluateBinary("<<", float64(1), float64(4))
	require.NoError(t, err)
	assert.Equal(t, float64(16), result)

	result, err = eval.EvaluateBinary(">>", float64(16), float64(4))
	require.NoError(t, err)
	assert.Equal(t, float64(1), result)
}

func TestExpressionEvaluator_UnknownOperator(t *testing.T) {
	eval := NewExpressionEvaluator()
	_, err := eval.EvaluateBinary("@@", 1.0, 2.0)
	assert.Error(t, err)
}

func TestExpressionEvaluator_UnaryOperations(t *testing.T) {
	eval := NewExpressionEvaluator()
	result, err := eval.EvaluateUnary("-", 10.0)
	require.NoError(t, err)
	assert.Equal(t, -10.0, result)

	result, err = eval.EvaluateUnary("+", 10.0)
	require.NoError(t, err)
	assert.Equal(t, 10.0, result)

	result, err = eval.EvaluateUnary("!", true)
	require.NoError(t, err)
	assert.Equal(t, false, result)

	_, err = eval.EvaluateUnary("@", 10.0)
	assert.Error(t, err)
}

func TestExpressionEvaluator_AssignmentType(t *testing.T) {
	eval := NewExpressionEvaluator()
	_, err := eval.EvaluateBinary("=", 1.0, 2.0)
	assert.Error(t, err)
}

func TestOperatorValidator_AllMethods(t *testing.T) {
	validator := NewOperatorValidator()
	err := validator.ValidateBinaryOperation("*", 10.0, 20.0)
	assert.NoError(t, err)

	err = validator.ValidateBinaryOperation("/", 10.0, 0.0)
	assert.Error(t, err)

	err = validator.ValidateBinaryOperation("%", 10.0, 0.0)
	assert.Error(t, err)

	err = validator.ValidateBinaryOperation("&", "abc", 2.0)
	assert.Error(t, err)

	err = validator.ValidateBinaryOperation("!", 1.0, 2.0)
	assert.Error(t, err)

	err = validator.ValidateBinaryOperation("@@", 1.0, 2.0)
	assert.Error(t, err)

	err = validator.ValidateUnaryOperation("-", 10.0)
	assert.NoError(t, err)

	err = validator.ValidateUnaryOperation("!", true)
	assert.NoError(t, err)

	err = validator.ValidateUnaryOperation("+", "abc")
	assert.Error(t, err)

	err = validator.ValidateUnaryOperation("*", 10.0)
	assert.Error(t, err)

	err = validator.ValidateUnaryOperation("@@", 10.0)
	assert.Error(t, err)
}

func TestBitwiseOperator_NegativeShift(t *testing.T) {
	op := &BitwiseOperator{}
	_, err := op.Evaluate("<<", 1.0, -1.0)
	assert.Error(t, err)
	_, err = op.Evaluate(">>", 1.0, -1.0)
	assert.Error(t, err)
	_, err = op.Evaluate("@@", 1.0, 2.0)
	assert.Error(t, err)
}

func TestBitwiseOperator_Xor(t *testing.T) {
	op := &BitwiseOperator{}
	result, err := op.Evaluate("^", float64(12), float64(10))
	require.NoError(t, err)
	assert.Equal(t, float64(6), result)
}

func TestLogicalOperator_All(t *testing.T) {
	op := &LogicalOperator{}
	result, err := op.Evaluate("&&", true, false)
	require.NoError(t, err)
	assert.Equal(t, false, result)

	result, err = op.Evaluate("||", true, false)
	require.NoError(t, err)
	assert.True(t, result)

	result, err = op.Evaluate("!", true)
	require.NoError(t, err)
	assert.Equal(t, false, result)

	_, err = op.Evaluate("&&", true)
	assert.Error(t, err)

	_, err = op.Evaluate("||", true)
	assert.Error(t, err)

	_, err = op.Evaluate("!", true, false)
	assert.Error(t, err)

	_, err = op.Evaluate("@@", true)
	assert.Error(t, err)
}

func TestArithmeticOperator_DivisionByZero(t *testing.T) {
	op := &ArithmeticOperator{}
	_, err := op.Evaluate("/", 10.0, 0.0)
	assert.Error(t, err)
	_, err = op.Evaluate("%", 10.0, 0.0)
	assert.Error(t, err)
	_, err = op.Evaluate("@@", 10.0, 5.0)
	assert.Error(t, err)
}

func TestComparisonOperator_All(t *testing.T) {
	op := &ComparisonOperator{}
	result, err := op.Evaluate("==", 10.0, 10.0)
	require.NoError(t, err)
	assert.True(t, result)

	result, err = op.Evaluate("!=", 10.0, 5.0)
	require.NoError(t, err)
	assert.True(t, result)

	result, err = op.Evaluate(">", 10.0, 5.0)
	require.NoError(t, err)
	assert.True(t, result)

	result, err = op.Evaluate("<", 5.0, 10.0)
	require.NoError(t, err)
	assert.True(t, result)

	result, err = op.Evaluate(">=", 10.0, 10.0)
	require.NoError(t, err)
	assert.True(t, result)

	result, err = op.Evaluate("<=", 5.0, 10.0)
	require.NoError(t, err)
	assert.True(t, result)
}

func TestIsNumeric(t *testing.T) {
	assert.True(t, isNumeric(42))
	assert.True(t, isNumeric(3.14))
	assert.True(t, isNumeric(float32(1.5)))
	assert.True(t, isNumeric(int64(100)))
	assert.True(t, isNumeric("42"))
	assert.False(t, isNumeric("abc"))
}

func TestIsOperator_Function(t *testing.T) {
	for _, op := range []string{"+", "-", "*", "/", "%", "^", "**", "==", "!=", ">", "<", ">=", "<=", "&&", "||", "!", "&", "|", "<<", ">>"} {
		assert.True(t, IsOperator(op), "expected %s to be operator", op)
	}
	assert.False(t, IsOperator("@"))
	assert.False(t, IsOperator("hello"))
}

func TestGetOperatorType_Default(t *testing.T) {
	assert.Equal(t, OpTypeArithmetic, GetOperatorType("@"))
}

func TestGetPrecedenceFromTable_Unknown(t *testing.T) {
	assert.Equal(t, -1, GetPrecedenceFromTable("@@"))
}

func TestGetAssociativityFromTable_Unknown(t *testing.T) {
	assert.Equal(t, "left", GetAssociativityFromTable("@@"))
}

func TestExecutor_ExecuteBatchWithContext(t *testing.T) {
	executor := NewExecutor(nil)
	ctx := context.Background()
	formulas := []string{"a + b", "a * b"}
	variables := map[string]interface{}{"a": 10.0, "b": 5.0}
	results, errors := executor.ExecuteBatchWithContext(ctx, formulas, variables)
	assert.Len(t, results, 2)
	assert.Len(t, errors, 2)
	assert.Equal(t, 15.0, results[0])
	assert.Equal(t, 50.0, results[1])
}

func TestExecutor_ExecuteParallelWithContext(t *testing.T) {
	executor := NewExecutor(nil)
	ctx := context.Background()
	formulas := []string{"a + b", "a * b", "a - b"}
	variables := map[string]interface{}{"a": 10.0, "b": 5.0}
	results, errors := executor.ExecuteParallelWithContext(ctx, formulas, variables, 2)
	assert.Len(t, results, 3)
	assert.Len(t, errors, 3)
	assert.Equal(t, 15.0, results[0])
	assert.Equal(t, 50.0, results[1])
	assert.Equal(t, 5.0, results[2])
}

func TestExecutor_ExecuteParallel_ZeroWorkers(t *testing.T) {
	executor := NewExecutor(nil)
	formulas := []string{"a + b"}
	variables := map[string]interface{}{"a": 10.0, "b": 5.0}
	results, errors := executor.ExecuteParallel(formulas, variables, 0)
	assert.Len(t, results, 1)
	assert.Len(t, errors, 1)
}

func TestExecutor_Execute_MaxRecursionDepth(t *testing.T) {
	config := &ExecutorConfig{MaxRecursionDepth: 1}
	executor := NewExecutor(config)
	_, err := executor.Execute("a + b", map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.NoError(t, err)
}

func TestExecutor_ExecuteWithTimeout_Expired(t *testing.T) {
	executor := NewExecutor(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := executor.ExecuteWithContext(ctx, "a + b", map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.Error(t, err)
}

func TestFunctionRegistry_GetNonExistent(t *testing.T) {
	registry := NewFunctionRegistry()
	_, exists := registry.Get("nonexistent_function")
	assert.False(t, exists)
}

func TestFuncSqrt_Negative(t *testing.T) {
	_, err := funcSqrt(-1.0)
	assert.Error(t, err)
}

func TestFuncLog_NonPositive(t *testing.T) {
	_, err := funcLog(0.0)
	assert.Error(t, err)
	_, err = funcLog(-1.0)
	assert.Error(t, err)
}

func TestFuncLog10_NonPositive(t *testing.T) {
	_, err := funcLog10(0.0)
	assert.Error(t, err)
}

func TestFuncLog2_NonPositive(t *testing.T) {
	_, err := funcLog2(0.0)
	assert.Error(t, err)
}

func TestFuncLog1p_Invalid(t *testing.T) {
	_, err := funcLog1p(-1.0)
	assert.Error(t, err)
	_, err = funcLog1p(-2.0)
	assert.Error(t, err)
}

func TestFuncAsin_OutOfRange(t *testing.T) {
	_, err := funcAsin(2.0)
	assert.Error(t, err)
	_, err = funcAsin(-2.0)
	assert.Error(t, err)
}

func TestFuncAcos_OutOfRange(t *testing.T) {
	_, err := funcAcos(2.0)
	assert.Error(t, err)
}

func TestFuncAcosh_LessThanOne(t *testing.T) {
	_, err := funcAcosh(0.5)
	assert.Error(t, err)
}

func TestFuncAtanh_OutOfRange(t *testing.T) {
	_, err := funcAtanh(1.0)
	assert.Error(t, err)
	_, err = funcAtanh(-1.0)
	assert.Error(t, err)
}

func TestFuncAvg_EmptyArray(t *testing.T) {
	result, err := funcAvg([]interface{}{[]interface{}{}})
	require.NoError(t, err)
	assert.Equal(t, 0.0, result)
}

func TestFuncSum_NoArgs(t *testing.T) {
	_, err := funcSum()
	assert.Error(t, err)
}

func TestFuncMax_NoArgs(t *testing.T) {
	_, err := funcMax()
	assert.Error(t, err)
}

func TestFuncMin_NoArgs(t *testing.T) {
	_, err := funcMin()
	assert.Error(t, err)
}

func TestFuncProduct_NoArgs(t *testing.T) {
	_, err := funcProduct()
	assert.Error(t, err)
}

func TestFuncVariance_NoArgs(t *testing.T) {
	_, err := funcVariance()
	assert.Error(t, err)
}

func TestFuncVariance_SingleValue(t *testing.T) {
	result, err := funcVariance(5.0)
	require.NoError(t, err)
	assert.Equal(t, 0.0, result)
}

func TestFuncMedian_NoArgs(t *testing.T) {
	_, err := funcMedian()
	assert.Error(t, err)
}

func TestFuncMedian_EvenCount(t *testing.T) {
	result, err := funcMedian(1.0, 2.0, 3.0, 4.0)
	require.NoError(t, err)
	assert.Equal(t, 2.5, result)
}

func TestFuncIf_WrongArgs(t *testing.T) {
	_, err := funcIf(true)
	assert.Error(t, err)
	_, err = funcIf(true, 1, 2, 3, 4)
	assert.Error(t, err)
}

func TestFuncIf_InvalidCondition(t *testing.T) {
	_, err := funcIf(struct{}{}, 1, 2)
	assert.Error(t, err)
}

func TestFuncSwitch_NoMatch(t *testing.T) {
	_, err := funcSwitch("x", "a", 1, "b", 2)
	assert.Error(t, err)
}

func TestFuncSwitch_Default(t *testing.T) {
	result, err := funcSwitch("x", "a", 1, 99)
	require.NoError(t, err)
	assert.Equal(t, 99, result)
}

func TestFuncSwitch_TooFewArgs(t *testing.T) {
	_, err := funcSwitch("x", "a")
	assert.Error(t, err)
}

func TestFuncCoalesce_AllNil(t *testing.T) {
	result, err := funcCoalesce(nil, nil)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestFuncCoalesce_NaN(t *testing.T) {
	result, err := funcCoalesce(math.NaN(), 42.0)
	require.NoError(t, err)
	assert.Equal(t, 42.0, result)
}

func TestFuncCoalesce_EmptyString(t *testing.T) {
	result, err := funcCoalesce("", "hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestFuncCoalesce_NoArgs(t *testing.T) {
	_, err := funcCoalesce()
	assert.Error(t, err)
}

func TestFuncLen_UnsupportedType(t *testing.T) {
	_, err := funcLen(42)
	assert.Error(t, err)
}

func TestFuncUpper_NonString(t *testing.T) {
	_, err := funcUpper(42)
	assert.Error(t, err)
}

func TestFuncLower_NonString(t *testing.T) {
	_, err := funcLower(42)
	assert.Error(t, err)
}

func TestFuncTrim_NonString(t *testing.T) {
	_, err := funcTrim(42)
	assert.Error(t, err)
}

func TestFuncSubstr_NonString(t *testing.T) {
	_, err := funcSubstr(42, 0)
	assert.Error(t, err)
}

func TestFuncSubstr_NegativeStart(t *testing.T) {
	result, err := funcSubstr("hello", -3.0)
	require.NoError(t, err)
	assert.Equal(t, "llo", result)
}

func TestFuncSubstr_StartBeyondLength(t *testing.T) {
	result, err := funcSubstr("hello", 100.0)
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestFuncSubstr_WithLength(t *testing.T) {
	result, err := funcSubstr("hello world", 0.0, 5.0)
	require.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestFuncSubstr_WrongArgCount(t *testing.T) {
	_, err := funcSubstr("hello")
	assert.Error(t, err)
	_, err = funcSubstr("hello", 0, 5, 3)
	assert.Error(t, err)
}

func TestFuncContains_NonStringArgs(t *testing.T) {
	_, err := funcContains(42, "hello")
	assert.Error(t, err)
	_, err = funcContains("hello", 42)
	assert.Error(t, err)
}

func TestFuncStartsWith_NonStringArgs(t *testing.T) {
	_, err := funcStartsWith(42, "he")
	assert.Error(t, err)
	_, err = funcStartsWith("hello", 42)
	assert.Error(t, err)
}

func TestFuncStartsWith_PrefixLonger(t *testing.T) {
	result, err := funcStartsWith("hi", "hello")
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestFuncEndsWith_NonStringArgs(t *testing.T) {
	_, err := funcEndsWith(42, "lo")
	assert.Error(t, err)
}

func TestFuncEndsWith_SuffixLonger(t *testing.T) {
	result, err := funcEndsWith("hi", "hello")
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestFuncReplace_NonStringArgs(t *testing.T) {
	_, err := funcReplace(42, "a", "b")
	assert.Error(t, err)
	_, err = funcReplace("hello", 42, "b")
	assert.Error(t, err)
	_, err = funcReplace("hello", "a", 42)
	assert.Error(t, err)
}

func TestFuncReplace_EmptyOld(t *testing.T) {
	result, err := funcReplace("hello", "", "x")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestFuncSplit_NonStringArgs(t *testing.T) {
	_, err := funcSplit(42, ",")
	assert.Error(t, err)
	_, err = funcSplit("a,b", 42)
	assert.Error(t, err)
}

func TestFuncSplit_EmptySeparator(t *testing.T) {
	result, err := funcSplit("abc", "")
	require.NoError(t, err)
	arr, ok := result.([]interface{})
	require.True(t, ok)
	assert.Len(t, arr, 3)
}

func TestFuncJoin_NonArrayFirst(t *testing.T) {
	_, err := funcJoin("not_array", ",")
	assert.Error(t, err)
	_, err = funcJoin([]interface{}{1, 2}, 42)
	assert.Error(t, err)
}

func TestFuncFirst_EmptyArray(t *testing.T) {
	_, err := funcFirst([]interface{}{})
	assert.Error(t, err)
}

func TestFuncFirst_NonArray(t *testing.T) {
	_, err := funcFirst(42)
	assert.Error(t, err)
}

func TestFuncLast_EmptyArray(t *testing.T) {
	_, err := funcLast([]interface{}{})
	assert.Error(t, err)
}

func TestFuncLast_NonArray(t *testing.T) {
	_, err := funcLast(42)
	assert.Error(t, err)
}

func TestFuncNth_OutOfRange(t *testing.T) {
	_, err := funcNth([]interface{}{1.0, 2.0}, 5.0)
	assert.Error(t, err)
}

func TestFuncNth_NegativeIndex(t *testing.T) {
	result, err := funcNth([]interface{}{1.0, 2.0, 3.0}, -1.0)
	require.NoError(t, err)
	assert.Equal(t, 3.0, result)
}

func TestFuncNth_NonArray(t *testing.T) {
	_, err := funcNth(42, 0)
	assert.Error(t, err)
}

func TestFuncSlice_WrongArgCount(t *testing.T) {
	_, err := funcSlice([]interface{}{1.0})
	assert.Error(t, err)
	_, err = funcSlice([]interface{}{1.0}, 0, 1, 2)
	assert.Error(t, err)
}

func TestFuncSlice_NonArray(t *testing.T) {
	_, err := funcSlice(42, 0)
	assert.Error(t, err)
}

func TestFuncSlice_NegativeStart(t *testing.T) {
	result, err := funcSlice([]interface{}{1.0, 2.0, 3.0}, -2.0)
	require.NoError(t, err)
	arr := result.([]interface{})
	assert.Len(t, arr, 2)
}

func TestFuncSlice_StartBeyondLength(t *testing.T) {
	result, err := funcSlice([]interface{}{1.0, 2.0}, 100.0)
	require.NoError(t, err)
	arr := result.([]interface{})
	assert.Empty(t, arr)
}

func TestFuncSlice_NegativeEnd(t *testing.T) {
	result, err := funcSlice([]interface{}{1.0, 2.0, 3.0, 4.0}, 1.0, -1.0)
	require.NoError(t, err)
	arr := result.([]interface{})
	assert.Len(t, arr, 2)
}

func TestFuncSlice_EndLessThanStart(t *testing.T) {
	result, err := funcSlice([]interface{}{1.0, 2.0, 3.0}, 2.0, 1.0)
	require.NoError(t, err)
	arr := result.([]interface{})
	assert.Empty(t, arr)
}

func TestFuncPush_NonArray(t *testing.T) {
	_, err := funcPush(42, 1)
	assert.Error(t, err)
}

func TestFuncPush_TooFewArgs(t *testing.T) {
	_, err := funcPush([]interface{}{1.0})
	assert.Error(t, err)
}

func TestFuncPop_EmptyArray(t *testing.T) {
	_, err := funcPop([]interface{}{})
	assert.Error(t, err)
}

func TestFuncPop_NonArray(t *testing.T) {
	_, err := funcPop(42)
	assert.Error(t, err)
}

func TestFuncReverse_NonArray(t *testing.T) {
	_, err := funcReverse(42)
	assert.Error(t, err)
}

func TestFuncSort_NonArray(t *testing.T) {
	_, err := funcSort(42)
	assert.Error(t, err)
}

func TestFuncSort_MixedTypes(t *testing.T) {
	result, err := funcSort([]interface{}{"b", "a", "c"})
	require.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, "a", arr[0])
}

func TestFuncFilter_NonArray(t *testing.T) {
	_, err := funcFilter(42, "truthy")
	assert.Error(t, err)
}

func TestFuncFilter_NonStringPredicate(t *testing.T) {
	_, err := funcFilter([]interface{}{1.0}, 42)
	assert.Error(t, err)
}

func TestFuncFilter_Truthy(t *testing.T) {
	result, err := funcFilter([]interface{}{1.0, 0.0, "hello", "", true, false, nil}, "truthy")
	require.NoError(t, err)
	arr := result.([]interface{})
	assert.Len(t, arr, 3)
}

func TestFuncMap_NonArray(t *testing.T) {
	_, err := funcMap(42, "double")
	assert.Error(t, err)
}

func TestFuncMap_NonStringFunc(t *testing.T) {
	_, err := funcMap([]interface{}{1.0}, 42)
	assert.Error(t, err)
}

func TestFuncMap_Double(t *testing.T) {
	result, err := funcMap([]interface{}{1.0, 2.0, 3.0}, "double")
	require.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, 2.0, arr[0])
	assert.Equal(t, 4.0, arr[1])
}

func TestFuncMap_String(t *testing.T) {
	result, err := funcMap([]interface{}{1.0, 2.0}, "string")
	require.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, "1", arr[0])
}

func TestFuncMap_UnknownFunc(t *testing.T) {
	result, err := funcMap([]interface{}{1.0}, "unknown")
	require.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, 1.0, arr[0])
}

func TestFuncReduce_NonArray(t *testing.T) {
	_, err := funcReduce(42, "sum", 0)
	assert.Error(t, err)
}

func TestFuncReduce_NonStringFunc(t *testing.T) {
	_, err := funcReduce([]interface{}{1.0}, 42, 0)
	assert.Error(t, err)
}

func TestFuncReduce_TooFewArgs(t *testing.T) {
	_, err := funcReduce([]interface{}{1.0}, "sum")
	assert.Error(t, err)
}

func TestFuncReduce_Sum(t *testing.T) {
	result, err := funcReduce([]interface{}{1.0, 2.0, 3.0}, "sum", 0.0)
	require.NoError(t, err)
	assert.Equal(t, 6.0, result)
}

func TestFuncReduce_Product(t *testing.T) {
	result, err := funcReduce([]interface{}{2.0, 3.0, 4.0}, "product", 1.0)
	require.NoError(t, err)
	assert.Equal(t, 24.0, result)
}

func TestFuncMod_DivisionByZero(t *testing.T) {
	_, err := funcMod(10.0, 0.0)
	assert.Error(t, err)
}

func TestFuncGcd_NegativeArgs(t *testing.T) {
	result, err := funcGcd(-12.0, -8.0)
	require.NoError(t, err)
	assert.Equal(t, float64(4), result)
}

func TestFuncLcm_ZeroArgs(t *testing.T) {
	result, err := funcLcm(0.0, 5.0)
	require.NoError(t, err)
	assert.Equal(t, 0.0, result)
}

func TestFuncLcm_NegativeArgs(t *testing.T) {
	result, err := funcLcm(-4.0, -6.0)
	require.NoError(t, err)
	assert.Equal(t, float64(12), result)
}

func TestFuncIsFinite_InvalidInput(t *testing.T) {
	result, err := funcIsFinite("abc")
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestFuncIsInf_InvalidInput(t *testing.T) {
	result, err := funcIsInf("abc")
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestFuncIsNaN_InvalidInput(t *testing.T) {
	result, err := funcIsNaN("abc")
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestFuncRandom_WithSeed(t *testing.T) {
	result, err := funcRandom(42.0)
	require.NoError(t, err)
	val, ok := result.(float64)
	assert.True(t, ok)
	assert.True(t, val >= 0.0 && val < 1.0)
}

func TestExtractNumbers_ArrayArg(t *testing.T) {
	numbers, err := extractNumbers([]interface{}{[]interface{}{1.0, 2.0, 3.0}})
	require.NoError(t, err)
	assert.Len(t, numbers, 3)
}

func TestExtractNumbers_InvalidType(t *testing.T) {
	_, err := extractNumbers([]interface{}{struct{}{}})
	assert.Error(t, err)
}

func TestIsTruthy(t *testing.T) {
	assert.False(t, isTruthy(nil))
	assert.False(t, isTruthy(false))
	assert.False(t, isTruthy(0.0))
	assert.False(t, isTruthy(""))
	assert.False(t, isTruthy([]interface{}{}))
	assert.True(t, isTruthy(true))
	assert.True(t, isTruthy(1.0))
	assert.True(t, isTruthy("hello"))
	assert.True(t, isTruthy([]interface{}{1}))
	assert.True(t, isTruthy(struct{}{}))
}

func TestCheckArgCount(t *testing.T) {
	err := checkArgCount("test", []interface{}{1, 2}, 2)
	assert.NoError(t, err)
	err = checkArgCount("test", []interface{}{1}, 2)
	assert.Error(t, err)
}

func TestExecutor_Execute_StringConcat(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute(`"hello" + " " + "world"`, nil)
	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestExecutor_Execute_BitwiseOr(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("a | b", map[string]interface{}{"a": float64(12), "b": float64(10)})
	require.NoError(t, err)
	assert.Equal(t, float64(14), result)
}

func TestExecutor_Execute_BitwiseAnd(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("a & b", map[string]interface{}{"a": float64(12), "b": float64(10)})
	require.NoError(t, err)
	assert.Equal(t, float64(8), result)
}

func TestExecutor_Execute_ShiftLeft(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("a << b", map[string]interface{}{"a": float64(1), "b": float64(4)})
	require.NoError(t, err)
	assert.Equal(t, float64(16), result)
}

func TestExecutor_Execute_ShiftRight(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("a >> b", map[string]interface{}{"a": float64(16), "b": float64(4)})
	require.NoError(t, err)
	assert.Equal(t, float64(1), result)
}

func TestFormulaManager_Import_InvalidExpression(t *testing.T) {
	mgr := NewFormulaManager(nil)
	formulas := []*Formula{{ID: "f1", Name: "test", Expression: "@#$%"}}
	err := mgr.Import(formulas, false)
	assert.Error(t, err)
}

func TestFormulaManager_Update_SameExpression(t *testing.T) {
	mgr := NewFormulaManager(nil)
	mgr.Create(&Formula{ID: "f1", Name: "test", Expression: "a + b"})
	f, _ := mgr.Get("f1")
	f.Name = "updated"
	err := mgr.Update(f)
	require.NoError(t, err)
	updated, _ := mgr.Get("f1")
	assert.Equal(t, "updated", updated.Name)
}

func TestFormulaManager_Rollback_NoVersionHistory(t *testing.T) {
	mgr := NewFormulaManager(nil)
	f := &Formula{ID: "f1", Name: "test", Expression: "a + b"}
	mgr.formulas["f1"] = f
	err := mgr.Rollback("f1", "1.0.0")
	assert.Error(t, err)
}
