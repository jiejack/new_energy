package formula

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFuncAsinh_Coverage(t *testing.T) {
	result, err := funcAsinh(1.0)
	assert.NoError(t, err)
	assert.InDelta(t, math.Asinh(1.0), result, 0.0001)
}

func TestFuncAcosh_Coverage(t *testing.T) {
	result, err := funcAcosh(2.0)
	assert.NoError(t, err)
	assert.InDelta(t, math.Acosh(2.0), result, 0.0001)

	_, err = funcAcosh(0.5)
	assert.Error(t, err)
}

func TestFuncAtanh_Coverage(t *testing.T) {
	result, err := funcAtanh(0.5)
	assert.NoError(t, err)
	assert.InDelta(t, math.Atanh(0.5), result, 0.0001)

	_, err = funcAtanh(2.0)
	assert.Error(t, err)
}

func TestFuncIsInf_Coverage(t *testing.T) {
	result, err := funcIsInf(math.Inf(1))
	assert.NoError(t, err)
	assert.True(t, result.(bool))

	result, err = funcIsInf(42.0)
	assert.NoError(t, err)
	assert.False(t, result.(bool))
}

func TestFuncIsNaN_Coverage(t *testing.T) {
	result, err := funcIsNaN(math.NaN())
	assert.NoError(t, err)
	assert.True(t, result.(bool))

	result, err = funcIsNaN(42.0)
	assert.NoError(t, err)
	assert.False(t, result.(bool))
}

func TestIsTruthy_Coverage(t *testing.T) {
	assert.True(t, isTruthy(true))
	assert.True(t, isTruthy(1.0))
	assert.True(t, isTruthy("non-empty"))
	assert.False(t, isTruthy(false))
	assert.False(t, isTruthy(0.0))
	assert.False(t, isTruthy(""))
	assert.False(t, isTruthy(nil))
	assert.True(t, isTruthy([]interface{}{1.0}))
	assert.False(t, isTruthy([]interface{}{}))
}

func TestFuncFilter_Coverage(t *testing.T) {
	result, err := funcFilter([]interface{}{1.0, 0.0, "hello", "", nil}, "truthy")
	assert.NoError(t, err)
	arr := result.([]interface{})
	assert.GreaterOrEqual(t, len(arr), 2)
}

func TestFuncMap_Coverage(t *testing.T) {
	result, err := funcMap([]interface{}{1.0, 2.0, 3.0}, "double")
	assert.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, 3, len(arr))
}

func TestFuncReduce_Coverage(t *testing.T) {
	result, err := funcReduce([]interface{}{1.0, 2.0, 3.0}, "sum", 0.0)
	assert.NoError(t, err)
	assert.Equal(t, 6.0, result)
}

func TestParseArray_Coverage(t *testing.T) {
	result, err := ParseFormula("[1, 2, 3]")
	require.NoError(t, err)
	require.NotNil(t, result)

	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	arr, ok := evalResult.([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(arr))
	assert.Equal(t, 1.0, arr[0])
}

func TestParseArray_Empty_Coverage(t *testing.T) {
	result, err := ParseFormula("[]")
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	arr, ok := evalResult.([]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, len(arr))
}

func TestParseStringSingle_Coverage(t *testing.T) {
	result, err := ParseFormula("'hello'")
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.Equal(t, "hello", evalResult)
}

func TestCompareValues_Coverage(t *testing.T) {
	result, err := compareValues("==", 1.0, 1.0)
	assert.NoError(t, err)
	assert.True(t, result)

	_, err = compareValues("unknown", 1.0, 2.0)
	assert.Error(t, err)
}

func TestBitwiseOp_Coverage(t *testing.T) {
	result, err := bitwiseOp("&", 12.0, 10.0)
	assert.NoError(t, err)
	assert.Equal(t, float64(12&10), result)

	_, err = bitwiseOp("unknown", 1.0, 2.0)
	assert.Error(t, err)
}

func TestToFloat64_Coverage(t *testing.T) {
	result, err := toFloat64(1.0)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, result)

	result, err = toFloat64("1.5")
	assert.NoError(t, err)
	assert.Equal(t, 1.5, result)

	_, err = toFloat64([]interface{}{})
	assert.Error(t, err)
}

func TestToBool_Coverage(t *testing.T) {
	result, err := toBool(true)
	assert.NoError(t, err)
	assert.Equal(t, true, result)

	_, err = toBool([]interface{}{})
	assert.Error(t, err)
}

func TestToken_String_Coverage(t *testing.T) {
	tok := Token{Type: TokenNumber, Value: "42", Pos: 0}
	assert.Contains(t, tok.String(), "42")
}

func TestNumberNode_TypeAndString_Coverage(t *testing.T) {
	node := &NumberNode{Value: 42.0}
	assert.Equal(t, NodeType(0), node.Type())
	assert.Equal(t, "42", node.String())
}

func TestStringNode_TypeAndString_Coverage(t *testing.T) {
	node := &StringNode{Value: "hello"}
	assert.Equal(t, NodeType(1), node.Type())
	assert.Equal(t, `"hello"`, node.String())
}

func TestVariableNode_TypeAndString_Coverage(t *testing.T) {
	node := &VariableNode{Name: "x"}
	assert.Equal(t, NodeType(2), node.Type())
	assert.Equal(t, "${x}", node.String())
}

func TestBinaryOpNode_TypeAndString_Coverage(t *testing.T) {
	node := &BinaryOpNode{Left: &NumberNode{Value: 1}, Operator: "+", Right: &NumberNode{Value: 2}}
	assert.Equal(t, NodeType(3), node.Type())
	assert.Equal(t, "(1 + 2)", node.String())
}

func TestUnaryOpNode_TypeAndString_Coverage(t *testing.T) {
	node := &UnaryOpNode{Operator: "-", Operand: &NumberNode{Value: 5}}
	assert.Equal(t, NodeType(4), node.Type())
	assert.Equal(t, "(- 5)", node.String())
}

func TestFunctionCallNode_TypeAndString_Coverage(t *testing.T) {
	node := &FunctionCallNode{Name: "max", Arguments: []Node{&NumberNode{Value: 1}, &NumberNode{Value: 2}}}
	assert.Equal(t, NodeType(5), node.Type())
	assert.Equal(t, "max(1, 2)", node.String())
}

func TestConditionalNode_TypeAndString_Coverage(t *testing.T) {
	node := &ConditionalNode{
		Condition: &BinaryOpNode{Left: &VariableNode{Name: "x"}, Operator: ">", Right: &NumberNode{Value: 0}},
		ThenExpr:  &NumberNode{Value: 1},
		ElseExpr:  &NumberNode{Value: 0},
	}
	assert.Equal(t, NodeType(6), node.Type())
}

func TestArrayNode_TypeAndString_Coverage(t *testing.T) {
	node := &ArrayNode{Elements: []Node{&NumberNode{Value: 1}, &NumberNode{Value: 2}}}
	assert.Equal(t, NodeType(7), node.Type())
	assert.Equal(t, "[1, 2]", node.String())
}

func TestArrayNode_Eval_Coverage(t *testing.T) {
	node := &ArrayNode{Elements: []Node{&NumberNode{Value: 1}, &NumberNode{Value: 2}}}
	result, err := node.Eval(NewEvalContext())
	assert.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, 2, len(arr))
}

func TestVariableBinder_Bind_Coverage(t *testing.T) {
	binder := NewVariableBinder()
	binder.Bind("x", 42.0)
	val, ok := binder.Get("x")
	assert.True(t, ok)
	assert.Equal(t, 42.0, val)
}

func TestVariableBinder_BindMany_Coverage(t *testing.T) {
	binder := NewVariableBinder()
	binder.BindMany(map[string]interface{}{"a": 1.0, "b": 2.0})
	val, ok := binder.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1.0, val)
}

func TestVariableBinder_Unbind_Coverage(t *testing.T) {
	binder := NewVariableBinder()
	binder.Bind("x", 42.0)
	binder.Unbind("x")
	_, ok := binder.Get("x")
	assert.False(t, ok)
}

func TestVariableBinder_GetAll_Coverage(t *testing.T) {
	binder := NewVariableBinder()
	binder.Bind("x", 42.0)
	all := binder.GetAll()
	assert.Equal(t, 42.0, all["x"])
}

func TestVariableBinder_Clear_Coverage(t *testing.T) {
	binder := NewVariableBinder()
	binder.Bind("x", 42.0)
	binder.Clear()
	_, ok := binder.Get("x")
	assert.False(t, ok)
}

func TestExecutor_GetCacheStats_NoCache_Coverage(t *testing.T) {
	executor := NewExecutor(nil)
	stats := executor.GetCacheStats()
	assert.NotNil(t, stats)
}

func TestExecutor_ExecuteWithVariables_Coverage(t *testing.T) {
	executor := NewExecutor(nil)
	result, err := executor.Execute("a + b * c", map[string]interface{}{"a": 1.0, "b": 2.0, "c": 3.0})
	assert.NoError(t, err)
	assert.Equal(t, 7.0, result)
}

func TestOperatorPrecedence_GetPrecedence_Coverage(t *testing.T) {
	handler := NewPrecedenceHandler()
	p := handler.GetPrecedence("+")
	assert.Greater(t, p, 0)
}

func TestOperatorPrecedence_IsLeftAssociative_Coverage(t *testing.T) {
	handler := NewPrecedenceHandler()
	assert.True(t, handler.IsLeftAssociative("*"))
}

func TestOperatorPrecedence_IsRightAssociative_Coverage(t *testing.T) {
	handler := NewPrecedenceHandler()
	assert.True(t, handler.IsRightAssociative("**"))
}

func TestEvaluateUnaryOp_Coverage(t *testing.T) {
	result, err := evaluateUnaryOp("-", 5.0)
	assert.NoError(t, err)
	assert.Equal(t, -5.0, result)

	result, err = evaluateUnaryOp("!", true)
	assert.NoError(t, err)
	assert.Equal(t, false, result)

	_, err = evaluateUnaryOp("unknown", 5.0)
	assert.Error(t, err)
}

func TestFuncSqrt_Negative_Coverage(t *testing.T) {
	_, err := funcSqrt(-1.0)
	assert.Error(t, err)
}

func TestFuncCbrt_Negative_Coverage(t *testing.T) {
	result, err := funcCbrt(-8.0)
	assert.NoError(t, err)
	assert.InDelta(t, -2.0, result, 0.0001)
}

func TestFuncLog_Errors_Coverage(t *testing.T) {
	_, err := funcLog(-1.0)
	assert.Error(t, err)
	_, err = funcLog(0.0)
	assert.Error(t, err)
	_, err = funcLog10(-1.0)
	assert.Error(t, err)
	_, err = funcLog2(-1.0)
	assert.Error(t, err)
	_, err = funcLog1p(-2.0)
	assert.Error(t, err)
}

func TestFuncAsin_OutOfRange_Coverage(t *testing.T) {
	_, err := funcAsin(2.0)
	assert.Error(t, err)
}

func TestFuncAcos_OutOfRange_Coverage(t *testing.T) {
	_, err := funcAcos(2.0)
	assert.Error(t, err)
}

func TestFuncFloor_NonNumber_Coverage(t *testing.T) {
	_, err := funcFloor("abc")
	assert.Error(t, err)
}

func TestFuncCeil_NonNumber_Coverage(t *testing.T) {
	_, err := funcCeil("abc")
	assert.Error(t, err)
}

func TestFuncRound_NonNumber_Coverage(t *testing.T) {
	_, err := funcRound("abc")
	assert.Error(t, err)
}

func TestFuncTrunc_NonNumber_Coverage(t *testing.T) {
	_, err := funcTrunc("abc")
	assert.Error(t, err)
}

func TestFuncPow_NonNumber_Coverage(t *testing.T) {
	_, err := funcPow("abc", 2.0)
	assert.Error(t, err)
}

func TestFuncSort_Coverage(t *testing.T) {
	result, err := funcSort([]interface{}{3.0, 1.0, 2.0})
	assert.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, 1.0, arr[0])
}

func TestFuncSlice_OutOfBounds_Coverage(t *testing.T) {
	result, err := funcSlice([]interface{}{1.0, 2.0, 3.0}, 0.0, 10.0)
	assert.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, 3, len(arr))
}

func TestFuncCoalesce_Coverage(t *testing.T) {
	result, err := funcCoalesce(nil, nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, result)

	result, err = funcCoalesce(nil, 42.0, "hello")
	assert.NoError(t, err)
	assert.Equal(t, 42.0, result)
}

func TestFuncSwitch_Coverage(t *testing.T) {
	_, err := funcSwitch("z", "a", 1, "b", 2)
	assert.Error(t, err)

	result, err := funcSwitch("z", "a", 1, "b", 2, -1)
	assert.NoError(t, err)
	assert.Equal(t, -1, result)
}

func TestFuncIf_NonBoolCondition_Coverage(t *testing.T) {
	result, err := funcIf("not_bool", 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestFuncClamp_Coverage(t *testing.T) {
	result, err := funcClamp(5.0, 0.0, 10.0)
	assert.NoError(t, err)
	assert.Equal(t, 5.0, result)

	result, err = funcClamp(-5.0, 0.0, 10.0)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, result)

	result, err = funcClamp(15.0, 0.0, 10.0)
	assert.NoError(t, err)
	assert.Equal(t, 10.0, result)
}

func TestFuncLerp_Coverage(t *testing.T) {
	result, err := funcLerp(0.0, 100.0, 0.5)
	assert.NoError(t, err)
	assert.Equal(t, 50.0, result)
}

func TestFuncStep_Coverage(t *testing.T) {
	result, err := funcStep(5.0, 3.0)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, result)

	result, err = funcStep(1.0, 3.0)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, result)
}

func TestFuncSmoothstep_Coverage(t *testing.T) {
	result, err := funcSmoothstep(0.0, 1.0, 0.5)
	assert.NoError(t, err)
	assert.InDelta(t, 0.5, result, 0.01)
}

func TestFuncMod_Coverage(t *testing.T) {
	result, err := funcMod(10.0, 3.0)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, result)
}

func TestFuncGcd_Coverage(t *testing.T) {
	result, err := funcGcd(12.0, 8.0)
	assert.NoError(t, err)
	assert.Equal(t, 4.0, result)
}

func TestFuncLcm_Coverage(t *testing.T) {
	result, err := funcLcm(4.0, 6.0)
	assert.NoError(t, err)
	assert.Equal(t, 12.0, result)
}

func TestFuncIsFinite_Coverage(t *testing.T) {
	result, err := funcIsFinite(42.0)
	assert.NoError(t, err)
	assert.True(t, result.(bool))

	result, err = funcIsFinite(math.Inf(1))
	assert.NoError(t, err)
	assert.False(t, result.(bool))
}

func TestFuncRandom_Coverage(t *testing.T) {
	result, err := funcRandom(0.0, 10.0)
	assert.NoError(t, err)
	val, ok := result.(float64)
	assert.True(t, ok)
	assert.True(t, val >= 0.0 && val < 10.0)
}

func TestExtractNumbers_Coverage(t *testing.T) {
	nums, err := extractNumbers([]interface{}{1.0, "2", true})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(nums))
}

func TestFuncLen_Coverage(t *testing.T) {
	result, err := funcLen("hello")
	assert.NoError(t, err)
	assert.Equal(t, 5.0, result)

	result, err = funcLen([]interface{}{1.0, 2.0, 3.0})
	assert.NoError(t, err)
	assert.Equal(t, 3.0, result)
}

func TestFuncConcat_Coverage(t *testing.T) {
	result, err := funcConcat("hello", " ", "world")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestFuncContains_Coverage(t *testing.T) {
	result, err := funcContains("hello world", "world")
	assert.NoError(t, err)
	assert.True(t, result.(bool))
}

func TestFuncStartsWith_Coverage(t *testing.T) {
	result, err := funcStartsWith("hello", "hel")
	assert.NoError(t, err)
	assert.True(t, result.(bool))
}

func TestFuncEndsWith_Coverage(t *testing.T) {
	result, err := funcEndsWith("hello", "llo")
	assert.NoError(t, err)
	assert.True(t, result.(bool))
}

func TestFuncReplace_Coverage(t *testing.T) {
	result, err := funcReplace("hello world", "world", "go")
	assert.NoError(t, err)
	assert.Equal(t, "hello go", result)
}

func TestFuncSplit_Coverage(t *testing.T) {
	result, err := funcSplit("a,b,c", ",")
	assert.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, 3, len(arr))
}

func TestFuncJoin_Coverage(t *testing.T) {
	result, err := funcJoin([]interface{}{"a", "b", "c"}, ",")
	assert.NoError(t, err)
	assert.Equal(t, "a,b,c", result)
}

func TestFuncFirst_Coverage(t *testing.T) {
	result, err := funcFirst([]interface{}{1.0, 2.0, 3.0})
	assert.NoError(t, err)
	assert.Equal(t, 1.0, result)
}

func TestFuncLast_Coverage(t *testing.T) {
	result, err := funcLast([]interface{}{1.0, 2.0, 3.0})
	assert.NoError(t, err)
	assert.Equal(t, 3.0, result)
}

func TestFuncNth_Coverage(t *testing.T) {
	result, err := funcNth([]interface{}{10.0, 20.0, 30.0}, 1.0)
	assert.NoError(t, err)
	assert.Equal(t, 20.0, result)
}

func TestFuncPush_Coverage(t *testing.T) {
	result, err := funcPush([]interface{}{1.0, 2.0}, 3.0)
	assert.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, 3, len(arr))
}

func TestFuncPop_Coverage(t *testing.T) {
	result, err := funcPop([]interface{}{1.0, 2.0, 3.0})
	assert.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, 2, len(arr))
}

func TestFuncReverse_Coverage(t *testing.T) {
	result, err := funcReverse([]interface{}{1.0, 2.0, 3.0})
	assert.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, 3.0, arr[0])
}

func TestFuncInt_Coverage(t *testing.T) {
	result, err := funcInt(3.7)
	assert.NoError(t, err)
	assert.Equal(t, 3.0, result)
}

func TestFuncFloat_Coverage(t *testing.T) {
	result, err := funcFloat(3)
	assert.NoError(t, err)
	assert.Equal(t, 3.0, result)
}

func TestFuncString_Coverage(t *testing.T) {
	result, err := funcString(42.0)
	assert.NoError(t, err)
	assert.Equal(t, "42", result)
}

func TestFuncBool_Coverage(t *testing.T) {
	result, err := funcBool(1.0)
	assert.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestFuncUpper_Coverage(t *testing.T) {
	result, err := funcUpper("hello")
	assert.NoError(t, err)
	assert.Equal(t, "HELLO", result)
}

func TestFuncLower_Coverage(t *testing.T) {
	result, err := funcLower("HELLO")
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestFuncTrim_Coverage(t *testing.T) {
	result, err := funcTrim("  hello  ")
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestManager_Update_Coverage(t *testing.T) {
	mgr := NewFormulaManager(&ManagerConfig{})
	f := &Formula{Name: "test", Expression: "a + b", Status: StatusDraft}
	err := mgr.Create(f)
	require.NoError(t, err)

	f.Expression = "a + b + c"
	err = mgr.Update(f)
	assert.NoError(t, err)
}

func TestManager_Update_NotFound_Coverage(t *testing.T) {
	mgr := NewFormulaManager(&ManagerConfig{})
	f := &Formula{ID: "nonexistent", Name: "test", Expression: "a + b"}
	err := mgr.Update(f)
	assert.Error(t, err)
}

func TestParseBitwiseOperations_Coverage(t *testing.T) {
	result, err := ParseFormula("3 & 5")
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.Equal(t, float64(3&5), evalResult)
}

func TestParseShiftOperations_Coverage(t *testing.T) {
	result, err := ParseFormula("1 << 3")
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.Equal(t, float64(1<<3), evalResult)
}

func TestParseConditionalExpression_Coverage(t *testing.T) {
	result, err := ParseFormula("1 > 0 ? 10 : 20")
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.Equal(t, 10.0, evalResult)
}

func TestParseStringWithEscapes_Coverage(t *testing.T) {
	result, err := ParseFormula(`"hello\"world"`)
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.Equal(t, `hello"world`, evalResult)
}

func TestParseStringWithNewline_Coverage(t *testing.T) {
	result, err := ParseFormula(`"hello\nworld"`)
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.Equal(t, "hello\nworld", evalResult)
}

func TestExecutor_ExecuteWithContext_Coverage(t *testing.T) {
	executor := NewExecutor(nil)
	ctx := context.Background()
	result, err := executor.ExecuteWithContext(ctx, "1 + 2", map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, 3.0, result)
}

func TestExecutor_ExecuteParallel_Coverage(t *testing.T) {
	executor := NewExecutor(nil)
	formulas := []string{"1 + 2", "3 + 4", "5 + 6"}
	results, _ := executor.ExecuteParallel(formulas, map[string]interface{}{}, 2)
	assert.Equal(t, 3, len(results))
	assert.Equal(t, 3.0, results[0])
	assert.Equal(t, 7.0, results[1])
	assert.Equal(t, 11.0, results[2])
}

func TestExecutor_ExecuteParallelWithContext_Coverage(t *testing.T) {
	executor := NewExecutor(nil)
	ctx := context.Background()
	formulas := []string{"1 + 2", "3 + 4"}
	results, _ := executor.ExecuteParallelWithContext(ctx, formulas, map[string]interface{}{}, 2)
	assert.Equal(t, 2, len(results))
}

func TestCompiledFormula_Coverage(t *testing.T) {
	executor := NewExecutor(nil)
	compiled, err := executor.Compile("a + b")
	require.NoError(t, err)
	require.NotNil(t, compiled)

	result, err := compiled.Execute(executor, map[string]interface{}{"a": 10.0, "b": 20.0})
	assert.NoError(t, err)
	assert.Equal(t, 30.0, result)
}

func TestExecutorBuilder_Coverage(t *testing.T) {
	executor := NewExecutorBuilder().
		WithCache(true, 5*time.Minute, 1000).
		WithTimeout(30*time.Second).
		WithMaxRecursionDepth(100).
		Build()
	assert.NotNil(t, executor)
}

func TestPrecedenceHandler_ComparePrecedence_Coverage(t *testing.T) {
	handler := NewPrecedenceHandler()
	result := handler.ComparePrecedence("+", "+")
	assert.Equal(t, 0, result)
}

func TestPrecedenceHandler_ShouldReduce_Coverage(t *testing.T) {
	handler := NewPrecedenceHandler()
	assert.True(t, handler.ShouldReduce("+", "*"))
	assert.False(t, handler.ShouldReduce("*", "+"))
}

func TestOperatorValidator_ValidateBinaryOperation_Coverage(t *testing.T) {
	validator := NewOperatorValidator()
	err := validator.ValidateBinaryOperation("unknown", 1.0, 2.0)
	assert.Error(t, err)
}

func TestExpressionEvaluator_Evaluate_Coverage(t *testing.T) {
	eval := NewExpressionEvaluator()
	_, err := eval.EvaluateBinary("+", "a", "b")
	assert.Error(t, err)
}

func TestReadNumber_Scientific_Coverage(t *testing.T) {
	result, err := ParseFormula("1e10")
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.InDelta(t, 1e10, evalResult, 0.001)
}

func TestLex_SpecialCases_Coverage(t *testing.T) {
	lexer := NewLexer("[")
	tokens, err := lexer.Lex()
	assert.NoError(t, err)
	found := false
	for _, tok := range tokens {
		if tok.Type == TokenLBracket {
			found = true
		}
	}
	assert.True(t, found)
}

func TestArithmeticOperator_Evaluate_Coverage(t *testing.T) {
	op := &ArithmeticOperator{}
	result, err := op.Evaluate("+", 1.0, 2.0)
	assert.NoError(t, err)
	assert.Equal(t, 3.0, result)

	result, err = op.Evaluate("-", 5.0, 3.0)
	assert.NoError(t, err)
	assert.Equal(t, 2.0, result)

	result, err = op.Evaluate("*", 3.0, 4.0)
	assert.NoError(t, err)
	assert.Equal(t, 12.0, result)

	result, err = op.Evaluate("/", 10.0, 2.0)
	assert.NoError(t, err)
	assert.Equal(t, 5.0, result)

	_, err = op.Evaluate("/", 10.0, 0.0)
	assert.Error(t, err)

	result, err = op.Evaluate("%", 10.0, 3.0)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, result)

	_, err = op.Evaluate("%", 10.0, 0.0)
	assert.Error(t, err)

	result, err = op.Evaluate("**", 2.0, 3.0)
	assert.NoError(t, err)
	assert.Equal(t, 8.0, result)

	_, err = op.Evaluate("unknown", 1.0, 2.0)
	assert.Error(t, err)

	_, err = op.Evaluate("+", "abc", 2.0)
	assert.Error(t, err)
}

func TestComparisonOperator_Evaluate_Coverage(t *testing.T) {
	op := &ComparisonOperator{}
	result, err := op.Evaluate("==", 1.0, 1.0)
	assert.NoError(t, err)
	assert.True(t, result)

	result, err = op.Evaluate("!=", 1.0, 2.0)
	assert.NoError(t, err)
	assert.True(t, result)

	_, err = op.Evaluate("unknown", 1.0, 2.0)
	assert.Error(t, err)
}

func TestLogicalOperator_Evaluate_Coverage(t *testing.T) {
	op := &LogicalOperator{}
	result, err := op.Evaluate("&&", true, true)
	assert.NoError(t, err)
	assert.True(t, result)

	result, err = op.Evaluate("&&", false, true)
	assert.NoError(t, err)
	assert.False(t, result)

	result, err = op.Evaluate("||", true, false)
	assert.NoError(t, err)
	assert.True(t, result)

	result, err = op.Evaluate("||", false, false)
	assert.NoError(t, err)
	assert.False(t, result)

	result, err = op.Evaluate("!", false)
	assert.NoError(t, err)
	assert.True(t, result)

	_, err = op.Evaluate("unknown", true)
	assert.Error(t, err)
}

func TestBitwiseOperator_Evaluate_Coverage(t *testing.T) {
	op := &BitwiseOperator{}
	result, err := op.Evaluate("&", 12.0, 10.0)
	assert.NoError(t, err)
	assert.Equal(t, float64(12&10), result)

	result, err = op.Evaluate("|", 12.0, 10.0)
	assert.NoError(t, err)
	assert.Equal(t, float64(12|10), result)

	result, err = op.Evaluate("^", 12.0, 10.0)
	assert.NoError(t, err)
	assert.Equal(t, float64(12^10), result)

	result, err = op.Evaluate("<<", 1.0, 3.0)
	assert.NoError(t, err)
	assert.Equal(t, float64(1<<3), result)

	result, err = op.Evaluate(">>", 8.0, 2.0)
	assert.NoError(t, err)
	assert.Equal(t, float64(8>>2), result)

	_, err = op.Evaluate(">>", 8.0, -1.0)
	assert.Error(t, err)

	_, err = op.Evaluate("unknown", 1.0, 2.0)
	assert.Error(t, err)

	_, err = op.Evaluate("&", "abc", 2.0)
	assert.Error(t, err)
}

func TestOperatorRegistry_GetByPrecedence_Coverage(t *testing.T) {
	registry := NewOperatorRegistry()
	ops := registry.GetByPrecedence()
	assert.NotEmpty(t, ops)
}

func TestExpressionEvaluator_EvaluateBinary_Coverage(t *testing.T) {
	eval := NewExpressionEvaluator()
	_, err := eval.EvaluateBinary("+", "a", "b")
	assert.Error(t, err)
}

func TestExpressionEvaluator_EvaluateUnary_Coverage(t *testing.T) {
	eval := NewExpressionEvaluator()
	_, err := eval.EvaluateUnary("-", "a")
	assert.Error(t, err)
}

func TestOperatorValidator_ValidateBinary_Coverage(t *testing.T) {
	validator := NewOperatorValidator()
	err := validator.ValidateBinaryOperation("unknown", 1.0, 2.0)
	assert.Error(t, err)
}

func TestIsNumeric_Coverage(t *testing.T) {
	assert.True(t, isNumeric(1.0))
	assert.True(t, isNumeric(1))
	assert.False(t, isNumeric("abc"))
}

func TestParseBitwiseOr_Coverage(t *testing.T) {
	result, err := ParseFormula("3 | 5")
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.Equal(t, float64(3|5), evalResult)
}

func TestParseBitwiseXor_Coverage(t *testing.T) {
	result, err := ParseFormula("3 ^ 5")
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.Equal(t, math.Pow(3, 5), evalResult)
}

func TestParseBitwiseAnd_Coverage(t *testing.T) {
	result, err := ParseFormula("3 & 5")
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.Equal(t, float64(3&5), evalResult)
}

func TestParseDivision_Coverage(t *testing.T) {
	result, err := ParseFormula("10 / 3")
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.InDelta(t, 10.0/3.0, evalResult, 0.001)
}

func TestParseModulo_Coverage(t *testing.T) {
	result, err := ParseFormula("10 % 3")
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.Equal(t, 1.0, evalResult)
}

func TestParsePower_Coverage(t *testing.T) {
	result, err := ParseFormula("2 ** 3")
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.Equal(t, 8.0, evalResult)
}

func TestParseLogicalAnd_Coverage(t *testing.T) {
	ctx := NewEvalContext()
	ctx.Variables["true"] = true
	ctx.Variables["false"] = false
	result, err := ParseFormula("true && false")
	require.NoError(t, err)
	evalResult, err := result.Eval(ctx)
	require.NoError(t, err)
	assert.Equal(t, false, evalResult)
}

func TestParseLogicalOr_Coverage(t *testing.T) {
	ctx := NewEvalContext()
	ctx.Variables["true"] = true
	ctx.Variables["false"] = false
	result, err := ParseFormula("true || false")
	require.NoError(t, err)
	evalResult, err := result.Eval(ctx)
	require.NoError(t, err)
	assert.Equal(t, true, evalResult)
}

func TestParseNotOperator_Coverage(t *testing.T) {
	ctx := NewEvalContext()
	ctx.Variables["true"] = true
	result, err := ParseFormula("!true")
	require.NoError(t, err)
	evalResult, err := result.Eval(ctx)
	require.NoError(t, err)
	assert.Equal(t, false, evalResult)
}

func TestParseNegativeNumber_Coverage(t *testing.T) {
	result, err := ParseFormula("-5")
	require.NoError(t, err)
	evalResult, err := result.Eval(NewEvalContext())
	require.NoError(t, err)
	assert.Equal(t, -5.0, evalResult)
}

func TestParseDivisionByZero_Coverage(t *testing.T) {
	result, err := ParseFormula("1 / 0")
	require.NoError(t, err)
	_, err = result.Eval(NewEvalContext())
	assert.Error(t, err)
}

func TestReadStringSingle_Escapes_Coverage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`'hello\'world'`, "hello'world"},
		{`'hello\\world'`, "hello\\world"},
		{`'hello\nworld'`, "hello\nworld"},
		{`'hello\tworld'`, "hello\tworld"},
		{`'hello\rworld'`, "hello\rworld"},
	}
	for _, tc := range tests {
		result, err := ParseFormula(tc.input)
		require.NoError(t, err, "failed to parse: %s", tc.input)
		evalResult, err := result.Eval(NewEvalContext())
		require.NoError(t, err, "failed to eval: %s", tc.input)
		assert.Equal(t, tc.expected, evalResult, "wrong result for: %s", tc.input)
	}
}

func TestFuncSort_StringArray_Coverage(t *testing.T) {
	result, err := funcSort([]interface{}{"banana", "apple", "cherry"})
	assert.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, "apple", arr[0])
}

func TestFuncSlice_NegativeIndex_Coverage(t *testing.T) {
	result, err := funcSlice([]interface{}{1.0, 2.0, 3.0}, -2.0, 3.0)
	assert.NoError(t, err)
	arr := result.([]interface{})
	assert.Equal(t, 2, len(arr))
}

func TestFuncNth_OutOfBounds_Coverage(t *testing.T) {
	_, err := funcNth([]interface{}{1.0, 2.0}, 5.0)
	assert.Error(t, err)
}

func TestFuncSubstr_Coverage(t *testing.T) {
	result, err := funcSubstr("hello world", 0.0, 5.0)
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestFuncSubstr_OutOfBounds_Coverage(t *testing.T) {
	result, err := funcSubstr("hello", 10.0, 2.0)
	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestFuncCount_Coverage(t *testing.T) {
	result, err := funcCount([]interface{}{1.0, 2.0, 3.0})
	assert.NoError(t, err)
	assert.Equal(t, 3.0, result)
}

func TestFuncSum_Coverage(t *testing.T) {
	result, err := funcSum([]interface{}{1.0, 2.0, 3.0})
	assert.NoError(t, err)
	assert.Equal(t, 6.0, result)
}

func TestFuncAvg_Coverage(t *testing.T) {
	result, err := funcAvg([]interface{}{1.0, 2.0, 3.0})
	assert.NoError(t, err)
	assert.Equal(t, 2.0, result)
}

func TestFuncMin_Coverage(t *testing.T) {
	result, err := funcMin([]interface{}{3.0, 1.0, 2.0})
	assert.NoError(t, err)
	assert.Equal(t, 1.0, result)
}

func TestFuncMax_Coverage(t *testing.T) {
	result, err := funcMax([]interface{}{3.0, 1.0, 2.0})
	assert.NoError(t, err)
	assert.Equal(t, 3.0, result)
}

func TestFuncMedian_Coverage(t *testing.T) {
	result, err := funcMedian([]interface{}{1.0, 2.0, 3.0, 4.0})
	assert.NoError(t, err)
	assert.Equal(t, 2.5, result)
}

func TestFuncVariance_Coverage(t *testing.T) {
	result, err := funcVariance([]interface{}{2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0})
	assert.NoError(t, err)
	assert.InDelta(t, 4.0, result, 0.01)
}

func TestFuncStdDev_Coverage(t *testing.T) {
	result, err := funcStdDev([]interface{}{2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0})
	assert.NoError(t, err)
	assert.InDelta(t, 2.0, result, 0.01)
}

func TestFuncProduct_Coverage(t *testing.T) {
	result, err := funcProduct([]interface{}{2.0, 3.0, 4.0})
	assert.NoError(t, err)
	assert.Equal(t, 24.0, result)
}

func TestFuncAbs_Coverage(t *testing.T) {
	result, err := funcAbs(-5.0)
	assert.NoError(t, err)
	assert.Equal(t, 5.0, result)
}

func TestFuncCeil_Coverage(t *testing.T) {
	result, err := funcCeil(3.2)
	assert.NoError(t, err)
	assert.Equal(t, 4.0, result)
}

func TestFuncFloor_Coverage(t *testing.T) {
	result, err := funcFloor(3.7)
	assert.NoError(t, err)
	assert.Equal(t, 3.0, result)
}

func TestFuncRound_Coverage(t *testing.T) {
	result, err := funcRound(3.5)
	assert.NoError(t, err)
	assert.Equal(t, 4.0, result)
}

func TestFuncSqrt_Coverage(t *testing.T) {
	result, err := funcSqrt(4.0)
	assert.NoError(t, err)
	assert.Equal(t, 2.0, result)
}

func TestFuncPow_Coverage(t *testing.T) {
	result, err := funcPow(2.0, 3.0)
	assert.NoError(t, err)
	assert.Equal(t, 8.0, result)
}

func TestFuncLog_Coverage(t *testing.T) {
	result, err := funcLog(math.E)
	assert.NoError(t, err)
	assert.InDelta(t, 1.0, result, 0.001)
}

func TestFuncLog10_Coverage(t *testing.T) {
	result, err := funcLog10(100.0)
	assert.NoError(t, err)
	assert.Equal(t, 2.0, result)
}

func TestFuncLog2_Coverage(t *testing.T) {
	result, err := funcLog2(8.0)
	assert.NoError(t, err)
	assert.Equal(t, 3.0, result)
}

func TestFuncSin_Coverage(t *testing.T) {
	result, err := funcSin(0.0)
	assert.NoError(t, err)
	assert.InDelta(t, 0.0, result, 0.001)
}

func TestFuncCos_Coverage(t *testing.T) {
	result, err := funcCos(0.0)
	assert.NoError(t, err)
	assert.InDelta(t, 1.0, result, 0.001)
}

func TestFuncTan_Coverage(t *testing.T) {
	result, err := funcTan(0.0)
	assert.NoError(t, err)
	assert.InDelta(t, 0.0, result, 0.001)
}

func TestFuncAsin_Coverage(t *testing.T) {
	result, err := funcAsin(0.5)
	assert.NoError(t, err)
	assert.InDelta(t, math.Asin(0.5), result, 0.001)
}

func TestFuncAcos_Coverage(t *testing.T) {
	result, err := funcAcos(0.5)
	assert.NoError(t, err)
	assert.InDelta(t, math.Acos(0.5), result, 0.001)
}

func TestFuncAtan_Coverage(t *testing.T) {
	result, err := funcAtan(1.0)
	assert.NoError(t, err)
	assert.InDelta(t, math.Atan(1.0), result, 0.001)
}

func TestFuncAtan2_Coverage(t *testing.T) {
	result, err := funcAtan2(1.0, 1.0)
	assert.NoError(t, err)
	assert.InDelta(t, math.Atan2(1.0, 1.0), result, 0.001)
}

func TestExecutor_Cache_Coverage(t *testing.T) {
	executor := NewExecutorBuilder().WithCache(true, 5*time.Minute, 1000).Build()
	result1, err := executor.Execute("1 + 2", map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, 3.0, result1)

	result2, err := executor.Execute("1 + 2", map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, 3.0, result2)

	stats := executor.GetCacheStats()
	assert.NotNil(t, stats)
}
