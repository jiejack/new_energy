package rule

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDataProvider struct {
	values map[string]float64
}

func (m *mockDataProvider) GetCurrentValue(ctx context.Context, pointID string) (*PointData, error) {
	if value, exists := m.values[pointID]; exists {
		return &PointData{PointID: pointID, Value: value, Timestamp: time.Now()}, nil
	}
	return nil, nil
}

func (m *mockDataProvider) GetTimeSeries(ctx context.Context, pointID string, start, end time.Time) (*TimeSeriesData, error) {
	return &TimeSeriesData{PointID: pointID, Values: []PointData{}}, nil
}

func (m *mockDataProvider) GetWindowData(ctx context.Context, pointID string, window time.Duration) (*TimeSeriesData, error) {
	if value, exists := m.values[pointID]; exists {
		return &TimeSeriesData{
			PointID: pointID,
			Values: []PointData{
				{PointID: pointID, Value: value, Timestamp: time.Now()},
				{PointID: pointID, Value: value * 0.9, Timestamp: time.Now().Add(-1 * time.Minute)},
			},
		}, nil
	}
	return &TimeSeriesData{PointID: pointID, Values: []PointData{}}, nil
}

func TestValueExpression_String_Plain(t *testing.T) {
	v := &ValueExpression{PointID: "point_001"}
	assert.Equal(t, "point_001", v.String())
}

func TestValueExpression_String_WithFunction(t *testing.T) {
	v := &ValueExpression{PointID: "point_001", Function: WindowAvg, WindowSize: 5 * time.Minute}
	assert.Contains(t, v.String(), "avg")
	assert.Contains(t, v.String(), "point_001")
}

func TestValueExpression_Validate_NoPointID(t *testing.T) {
	v := &ValueExpression{PointID: ""}
	assert.Error(t, v.Validate())
}

func TestValueExpression_Validate_FunctionWithoutWindow(t *testing.T) {
	v := &ValueExpression{PointID: "point_001", Function: WindowAvg, WindowSize: 0}
	assert.Error(t, v.Validate())
}

func TestValueExpression_Validate_Valid(t *testing.T) {
	v := &ValueExpression{PointID: "point_001"}
	assert.NoError(t, v.Validate())
}

func TestThreshold_String(t *testing.T) {
	th := &Threshold{Type: ThresholdTypeAbsolute, Value: 100}
	assert.Equal(t, "absolute(100)", th.String())
}

func TestThreshold_Validate_PercentageOutOfRange(t *testing.T) {
	th := &Threshold{Type: ThresholdTypePercentage, Value: 150}
	assert.Error(t, th.Validate())
	th.Value = -10
	assert.Error(t, th.Validate())
}

func TestThreshold_Validate_PercentageValid(t *testing.T) {
	th := &Threshold{Type: ThresholdTypePercentage, Value: 80}
	assert.NoError(t, th.Validate())
}

func TestThreshold_Validate_Absolute(t *testing.T) {
	th := &Threshold{Type: ThresholdTypeAbsolute, Value: 100}
	assert.NoError(t, th.Validate())
}

func TestComparisonCondition_Validate(t *testing.T) {
	cc := &ComparisonCondition{
		Left:     ValueExpression{PointID: "point_001"},
		Operator: OpGT,
		Right:    Threshold{Type: ThresholdTypeAbsolute, Value: 100},
	}
	assert.NoError(t, cc.Validate())
}

func TestComparisonCondition_Validate_InvalidOperator(t *testing.T) {
	cc := &ComparisonCondition{
		Left:     ValueExpression{PointID: "point_001"},
		Operator: Operator("??"),
		Right:    Threshold{Type: ThresholdTypeAbsolute, Value: 100},
	}
	assert.Error(t, cc.Validate())
}

func TestComparisonCondition_Validate_InvalidLeft(t *testing.T) {
	cc := &ComparisonCondition{
		Left:     ValueExpression{PointID: ""},
		Operator: OpGT,
		Right:    Threshold{Type: ThresholdTypeAbsolute, Value: 100},
	}
	assert.Error(t, cc.Validate())
}

func TestComparisonCondition_Validate_InvalidRight(t *testing.T) {
	cc := &ComparisonCondition{
		Left:     ValueExpression{PointID: "point_001"},
		Operator: OpGT,
		Right:    Threshold{Type: ThresholdTypePercentage, Value: 150},
	}
	assert.Error(t, cc.Validate())
}

func TestComparisonCondition_String(t *testing.T) {
	cc := &ComparisonCondition{
		Left:     ValueExpression{PointID: "point_001"},
		Operator: OpGT,
		Right:    Threshold{Type: ThresholdTypeAbsolute, Value: 100},
	}
	s := cc.String()
	assert.Contains(t, s, "point_001")
	assert.Contains(t, s, ">")
}

func TestLogicalCondition_String_AND(t *testing.T) {
	lc := &LogicalCondition{
		Operator: LogicalAND,
		Operands: []Expression{
			&ComparisonCondition{
				Left:     ValueExpression{PointID: "p1"},
				Operator: OpGT,
				Right:    Threshold{Type: ThresholdTypeAbsolute, Value: 100},
			},
			&ComparisonCondition{
				Left:     ValueExpression{PointID: "p2"},
				Operator: OpLT,
				Right:    Threshold{Type: ThresholdTypeAbsolute, Value: 50},
			},
		},
	}
	s := lc.String()
	assert.Contains(t, s, "AND")
}

func TestLogicalCondition_String_NOT(t *testing.T) {
	lc := &LogicalCondition{
		Operator: LogicalNOT,
		Operands: []Expression{
			&ComparisonCondition{
				Left:     ValueExpression{PointID: "p1"},
				Operator: OpGT,
				Right:    Threshold{Type: ThresholdTypeAbsolute, Value: 100},
			},
		},
	}
	s := lc.String()
	assert.Contains(t, s, "NOT")
}

func TestLogicalCondition_String_NOT_NoOperands(t *testing.T) {
	lc := &LogicalCondition{Operator: LogicalNOT, Operands: []Expression{}}
	assert.Equal(t, "NOT", lc.String())
}

func TestLogicalCondition_Validate_InvalidOperator(t *testing.T) {
	lc := &LogicalCondition{
		Operator: LogicalOperator("XOR"),
		Operands: []Expression{
			&ComparisonCondition{Left: ValueExpression{PointID: "p1"}, Operator: OpGT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 100}},
			&ComparisonCondition{Left: ValueExpression{PointID: "p2"}, Operator: OpLT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 50}},
		},
	}
	assert.Error(t, lc.Validate())
}

func TestLogicalCondition_Validate_NOT_WrongOperands(t *testing.T) {
	lc := &LogicalCondition{
		Operator: LogicalNOT,
		Operands: []Expression{
			&ComparisonCondition{Left: ValueExpression{PointID: "p1"}, Operator: OpGT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 100}},
			&ComparisonCondition{Left: ValueExpression{PointID: "p2"}, Operator: OpLT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 50}},
		},
	}
	assert.Error(t, lc.Validate())
}

func TestLogicalCondition_Validate_AND_InsufficientOperands(t *testing.T) {
	lc := &LogicalCondition{
		Operator: LogicalAND,
		Operands: []Expression{
			&ComparisonCondition{Left: ValueExpression{PointID: "p1"}, Operator: OpGT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 100}},
		},
	}
	assert.Error(t, lc.Validate())
}

func TestLogicalCondition_Validate_InvalidOperand(t *testing.T) {
	lc := &LogicalCondition{
		Operator: LogicalAND,
		Operands: []Expression{
			&ComparisonCondition{Left: ValueExpression{PointID: ""}, Operator: OpGT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 100}},
			&ComparisonCondition{Left: ValueExpression{PointID: "p2"}, Operator: OpLT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 50}},
		},
	}
	assert.Error(t, lc.Validate())
}

func TestCondition_String_Comparison(t *testing.T) {
	c := &Condition{
		Comparison: &ComparisonCondition{
			Left:     ValueExpression{PointID: "p1"},
			Operator: OpGT,
			Right:    Threshold{Type: ThresholdTypeAbsolute, Value: 100},
		},
	}
	assert.NotEmpty(t, c.String())
}

func TestCondition_String_Logical(t *testing.T) {
	c := &Condition{
		Logical: &LogicalCondition{
			Operator: LogicalAND,
			Operands: []Expression{
				&ComparisonCondition{Left: ValueExpression{PointID: "p1"}, Operator: OpGT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 100}},
				&ComparisonCondition{Left: ValueExpression{PointID: "p2"}, Operator: OpLT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 50}},
			},
		},
	}
	assert.NotEmpty(t, c.String())
}

func TestCondition_String_Empty(t *testing.T) {
	c := &Condition{}
	assert.Equal(t, "", c.String())
}

func TestCondition_Validate_NilBoth(t *testing.T) {
	c := &Condition{}
	assert.Error(t, c.Validate())
}

func TestCondition_Validate_BothSet(t *testing.T) {
	c := &Condition{
		Comparison: &ComparisonCondition{Left: ValueExpression{PointID: "p1"}, Operator: OpGT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 100}},
		Logical:    &LogicalCondition{Operator: LogicalAND, Operands: []Expression{}},
	}
	assert.Error(t, c.Validate())
}

func TestAction_Validate_EmptyType(t *testing.T) {
	a := &Action{Type: ""}
	assert.Error(t, a.Validate())
}

func TestAction_Validate_Valid(t *testing.T) {
	a := &Action{Type: "notify", Parameters: map[string]interface{}{"channel": "email"}}
	assert.NoError(t, a.Validate())
}

func TestRuleDSL_Validate_NoID(t *testing.T) {
	r := NewRuleDSL("", "test", "1.0")
	r.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	assert.Error(t, r.Validate())
}

func TestRuleDSL_Validate_NoName(t *testing.T) {
	r := NewRuleDSL("r1", "", "1.0")
	r.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	assert.Error(t, r.Validate())
}

func TestRuleDSL_Validate_NoVersion(t *testing.T) {
	r := NewRuleDSL("r1", "test", "")
	r.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	assert.Error(t, r.Validate())
}

func TestRuleDSL_Validate_InvalidPriority(t *testing.T) {
	r := NewRuleDSL("r1", "test", "1.0")
	r.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r.Priority = 0
	assert.Error(t, r.Validate())
	r.Priority = 101
	assert.Error(t, r.Validate())
}

func TestRuleDSL_Validate_InvalidAction(t *testing.T) {
	r := NewRuleDSL("r1", "test", "1.0")
	r.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r.AddAction("", nil)
	assert.Error(t, r.Validate())
}

func TestRuleDSL_ToJSON_FromJSON(t *testing.T) {
	r := NewRuleDSL("r1", "test", "1.0")
	r.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r.AddTag("critical")
	r.AddAction("notify", map[string]interface{}{"channel": "email"})

	jsonStr, err := r.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	parsed, err := FromJSON(jsonStr)
	require.NoError(t, err)
	assert.Equal(t, r.ID, parsed.ID)
	assert.Equal(t, r.Name, parsed.Name)
}

func TestRuleDSL_FromJSON_Invalid(t *testing.T) {
	_, err := FromJSON("invalid json")
	assert.Error(t, err)
}

func TestRuleDSL_AddTag_RemoveTag(t *testing.T) {
	r := NewRuleDSL("r1", "test", "1.0")
	r.AddTag("tag1")
	r.AddTag("tag2")
	assert.Len(t, r.Tags, 2)
	r.RemoveTag("tag1")
	assert.Len(t, r.Tags, 1)
	assert.Equal(t, "tag2", r.Tags[0])
}

func TestRuleDSL_RemoveTag_NonExistent(t *testing.T) {
	r := NewRuleDSL("r1", "test", "1.0")
	r.AddTag("tag1")
	r.RemoveTag("nonexistent")
	assert.Len(t, r.Tags, 1)
}

func TestRuleDSL_SetCondition(t *testing.T) {
	r := NewRuleDSL("r1", "test", "1.0")
	cond := Condition{
		Comparison: &ComparisonCondition{
			Left:     ValueExpression{PointID: "p1"},
			Operator: OpEQ,
			Right:    Threshold{Type: ThresholdTypeAbsolute, Value: 42},
		},
	}
	r.SetCondition(cond)
	assert.NotNil(t, r.Condition.Comparison)
	assert.Equal(t, OpEQ, r.Condition.Comparison.Operator)
}

func TestRuleDSL_SetWindowCondition(t *testing.T) {
	r := NewRuleDSL("r1", "test", "1.0")
	r.SetWindowCondition("p1", WindowAvg, 5*time.Minute, OpGT, ThresholdTypeAbsolute, 100)
	assert.NotNil(t, r.Condition.Comparison)
	assert.Equal(t, WindowAvg, r.Condition.Comparison.Left.Function)
	assert.Equal(t, 5*time.Minute, r.Condition.Comparison.Left.WindowSize)
}

func TestRuleDSL_String(t *testing.T) {
	r := NewRuleDSL("r1", "test", "1.0")
	r.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	s := r.String()
	assert.NotEmpty(t, s)
	assert.Contains(t, s, "r1")
}

func TestParser_SimpleComparison(t *testing.T) {
	cond, err := Parse("point_001 > 100")
	require.NoError(t, err)
	assert.NotNil(t, cond.Comparison)
	assert.Equal(t, OpGT, cond.Comparison.Operator)
	assert.Equal(t, "point_001", cond.Comparison.Left.PointID)
	assert.Equal(t, 100.0, cond.Comparison.Right.Value)
}

func TestParser_WindowFunction(t *testing.T) {
	cond, err := Parse("avg(point_001, 5m) > 50")
	require.NoError(t, err)
	assert.NotNil(t, cond.Comparison)
	assert.Equal(t, WindowAvg, cond.Comparison.Left.Function)
	assert.Equal(t, "point_001", cond.Comparison.Left.PointID)
	assert.Equal(t, 5*time.Minute, cond.Comparison.Left.WindowSize)
}

func TestParser_PercentageThreshold(t *testing.T) {
	cond, err := Parse("point_001 > percentage(80)")
	require.NoError(t, err)
	assert.NotNil(t, cond.Comparison)
	assert.Equal(t, ThresholdTypePercentage, cond.Comparison.Right.Type)
	assert.Equal(t, 80.0, cond.Comparison.Right.Value)
}

func TestParser_AND(t *testing.T) {
	cond, err := Parse("AND(point_001 > 100, point_002 < 50)")
	require.NoError(t, err)
	assert.NotNil(t, cond.Logical)
	assert.Equal(t, LogicalAND, cond.Logical.Operator)
	assert.Len(t, cond.Logical.Operands, 2)
}

func TestParser_OR(t *testing.T) {
	cond, err := Parse("OR(point_001 > 100, point_002 < 50)")
	require.NoError(t, err)
	assert.NotNil(t, cond.Logical)
	assert.Equal(t, LogicalOR, cond.Logical.Operator)
	assert.Len(t, cond.Logical.Operands, 2)
}

func TestParser_NOT(t *testing.T) {
	cond, err := Parse("NOT point_001 > 100")
	require.NoError(t, err)
	assert.NotNil(t, cond.Logical)
	assert.Equal(t, LogicalNOT, cond.Logical.Operator)
	assert.Len(t, cond.Logical.Operands, 1)
}

func TestParser_InvalidInput(t *testing.T) {
	_, err := Parse("@#$%")
	assert.Error(t, err)
}

func TestParser_UnexpectedChar(t *testing.T) {
	_, err := Parse("point_001 @ 100")
	assert.Error(t, err)
}

func TestParseString(t *testing.T) {
	cond, err := ParseString("point_001 > 100")
	require.NoError(t, err)
	assert.NotNil(t, cond)
}

func TestParseRuleDSL_JSON(t *testing.T) {
	r := NewRuleDSL("r1", "test", "1.0")
	r.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	jsonStr, err := r.ToJSON()
	require.NoError(t, err)

	parsed, err := ParseRuleDSL(jsonStr)
	require.NoError(t, err)
	assert.Equal(t, "r1", parsed.ID)
}

func TestParseRuleDSL_Expression(t *testing.T) {
	parsed, err := ParseRuleDSL("point_001 > 100")
	require.NoError(t, err)
	assert.NotNil(t, parsed)
}

func TestLexer_String(t *testing.T) {
	tokens, err := NewLexer(`point_001 > "hello"`).Lex()
	require.NoError(t, err)
	hasString := false
	for _, tok := range tokens {
		if tok.Type == TokenString {
			hasString = true
			assert.Equal(t, "hello", tok.Value)
		}
	}
	assert.True(t, hasString)
}

func TestLexer_UnterminatedString(t *testing.T) {
	_, err := NewLexer(`"unterminated`).Lex()
	assert.Error(t, err)
}

func TestLexer_Duration(t *testing.T) {
	tokens, err := NewLexer("5m").Lex()
	require.NoError(t, err)
	assert.Equal(t, TokenDuration, tokens[0].Type)
	assert.Equal(t, "5m", tokens[0].Value)
}

func TestLexer_NegativeNumber(t *testing.T) {
	tokens, err := NewLexer("-10.5").Lex()
	require.NoError(t, err)
	assert.Equal(t, TokenNumber, tokens[0].Type)
	assert.Equal(t, "-10.5", tokens[0].Value)
}

func TestLexer_Operators(t *testing.T) {
	ops := []string{">", "<", ">=", "<=", "==", "!="}
	for _, op := range ops {
		tokens, err := NewLexer(op).Lex()
		require.NoError(t, err)
		found := false
		for _, tok := range tokens {
			if tok.Type == TokenOperator && tok.Value == op {
				found = true
				break
			}
		}
		assert.True(t, found, "expected operator %s", op)
	}
}

func TestLexer_LogicalOps(t *testing.T) {
	tokens, err := NewLexer("AND OR NOT").Lex()
	require.NoError(t, err)
	logicalCount := 0
	for _, tok := range tokens {
		if tok.Type == TokenLogicalOp {
			logicalCount++
		}
	}
	assert.Equal(t, 3, logicalCount)
}

func TestLexer_ParensAndComma(t *testing.T) {
	tokens, err := NewLexer("( , )").Lex()
	require.NoError(t, err)
	types := []TokenType{}
	for _, tok := range tokens {
		if tok.Type != TokenEOF {
			types = append(types, tok.Type)
		}
	}
	assert.Contains(t, types, TokenLParen)
	assert.Contains(t, types, TokenRParen)
	assert.Contains(t, types, TokenComma)
}

func TestToken_String(t *testing.T) {
	tok := Token{Type: TokenIdentifier, Value: "test", Pos: 0}
	s := tok.String()
	assert.Contains(t, s, "test")
}

func TestEngine_AddRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 120.0}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	require.NoError(t, engine.AddRule(rule))
}

func TestEngine_AddRule_Invalid(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rule := &RuleDSL{}
	assert.Error(t, engine.AddRule(rule))
}

func TestEngine_RemoveRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 120.0}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	engine.AddRule(rule)
	engine.RemoveRule("r1")
	_, exists := engine.GetRule("r1")
	assert.False(t, exists)
}

func TestEngine_GetRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	engine.AddRule(rule)
	found, exists := engine.GetRule("r1")
	assert.True(t, exists)
	assert.Equal(t, "r1", found.ID)
	_, exists = engine.GetRule("nonexistent")
	assert.False(t, exists)
}

func TestEngine_GetAllRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	r1 := NewRuleDSL("r1", "test1", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r2 := NewRuleDSL("r2", "test2", "1.0")
	r2.SetComparisonCondition("p2", OpLT, ThresholdTypeAbsolute, 50)
	engine.AddRule(r1)
	engine.AddRule(r2)
	rules := engine.GetAllRules()
	assert.Len(t, rules, 2)
}

func TestEngine_Evaluate_Triggered(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 120.0}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	engine.AddRule(rule)
	result, err := engine.Evaluate(context.Background(), "r1")
	require.NoError(t, err)
	assert.True(t, result.Triggered)
	assert.Equal(t, 120.0, result.Value)
}

func TestEngine_Evaluate_NotTriggered(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 80.0}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	engine.AddRule(rule)
	result, err := engine.Evaluate(context.Background(), "r1")
	require.NoError(t, err)
	assert.False(t, result.Triggered)
}

func TestEngine_Evaluate_DisabledRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 120.0}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.Enabled = false
	engine.AddRule(rule)
	result, err := engine.Evaluate(context.Background(), "r1")
	require.NoError(t, err)
	assert.False(t, result.Triggered)
}

func TestEngine_Evaluate_NotFound(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	_, err := engine.Evaluate(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestEngine_EvaluateAll(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 120.0, "p2": 40.0}}
	engine := NewEngine(provider)
	r1 := NewRuleDSL("r1", "test1", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r2 := NewRuleDSL("r2", "test2", "1.0")
	r2.SetComparisonCondition("p2", OpLT, ThresholdTypeAbsolute, 50)
	engine.AddRule(r1)
	engine.AddRule(r2)
	results, err := engine.EvaluateAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestEngine_EvaluateBatch(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 120.0}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	engine.AddRule(rule)
	results, err := engine.EvaluateBatch(context.Background(), []string{"r1"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestEngine_EvaluateBatch_Error(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	_, err := engine.EvaluateBatch(context.Background(), []string{"nonexistent"})
	assert.Error(t, err)
}

func TestEngine_EvaluateWithData(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	result, err := engine.EvaluateWithData(context.Background(), rule, map[string]float64{"p1": 150.0})
	require.NoError(t, err)
	assert.True(t, result.Triggered)
	assert.Equal(t, 150.0, result.Value)
}

func TestEngine_EvaluateWithData_Disabled(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.Enabled = false
	result, err := engine.EvaluateWithData(context.Background(), rule, map[string]float64{"p1": 150.0})
	require.NoError(t, err)
	assert.False(t, result.Triggered)
}

func TestEngine_EvaluateWithData_NilProvider(t *testing.T) {
	engine := NewEngine(nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	result, err := engine.EvaluateWithData(context.Background(), rule, map[string]float64{})
	require.NoError(t, err)
	assert.Error(t, result.Error)
}

func TestEngine_Evaluate_PercentageThreshold(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 200.0}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypePercentage, 80)
	engine.AddRule(rule)
	result, err := engine.Evaluate(context.Background(), "r1")
	require.NoError(t, err)
	assert.Equal(t, 160.0, result.Threshold)
}

func TestEngine_Evaluate_RateThreshold(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 200.0}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeRate, 10)
	engine.AddRule(rule)
	result, err := engine.Evaluate(context.Background(), "r1")
	require.NoError(t, err)
	assert.Equal(t, 10.0, result.Threshold)
}

func TestEngine_Evaluate_WindowFunction(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 120.0}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetWindowCondition("p1", WindowAvg, 5*time.Minute, OpGT, ThresholdTypeAbsolute, 100)
	engine.AddRule(rule)
	result, err := engine.Evaluate(context.Background(), "r1")
	require.NoError(t, err)
	assert.True(t, result.Triggered)
}

func TestEngine_Evaluate_LogicalAND(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 120.0, "p2": 40.0}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.Condition = Condition{
		Logical: &LogicalCondition{
			Operator: LogicalAND,
			Operands: []Expression{
				&ComparisonCondition{Left: ValueExpression{PointID: "p1"}, Operator: OpGT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 100}},
				&ComparisonCondition{Left: ValueExpression{PointID: "p2"}, Operator: OpLT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 50}},
			},
		},
	}
	engine.AddRule(rule)
	result, err := engine.Evaluate(context.Background(), "r1")
	require.NoError(t, err)
	assert.True(t, result.Triggered)
}

func TestEngine_Evaluate_LogicalAND_Fail(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 120.0, "p2": 60.0}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.Condition = Condition{
		Logical: &LogicalCondition{
			Operator: LogicalAND,
			Operands: []Expression{
				&ComparisonCondition{Left: ValueExpression{PointID: "p1"}, Operator: OpGT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 100}},
				&ComparisonCondition{Left: ValueExpression{PointID: "p2"}, Operator: OpLT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 50}},
			},
		},
	}
	engine.AddRule(rule)
	result, err := engine.Evaluate(context.Background(), "r1")
	require.NoError(t, err)
	assert.False(t, result.Triggered)
}

func TestEngine_Evaluate_LogicalOR(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 120.0, "p2": 60.0}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.Condition = Condition{
		Logical: &LogicalCondition{
			Operator: LogicalOR,
			Operands: []Expression{
				&ComparisonCondition{Left: ValueExpression{PointID: "p1"}, Operator: OpGT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 100}},
				&ComparisonCondition{Left: ValueExpression{PointID: "p2"}, Operator: OpLT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 50}},
			},
		},
	}
	engine.AddRule(rule)
	result, err := engine.Evaluate(context.Background(), "r1")
	require.NoError(t, err)
	assert.True(t, result.Triggered)
}

func TestEngine_Evaluate_LogicalNOT(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 80.0}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.Condition = Condition{
		Logical: &LogicalCondition{
			Operator: LogicalNOT,
			Operands: []Expression{
				&ComparisonCondition{Left: ValueExpression{PointID: "p1"}, Operator: OpGT, Right: Threshold{Type: ThresholdTypeAbsolute, Value: 100}},
			},
		},
	}
	engine.AddRule(rule)
	result, err := engine.Evaluate(context.Background(), "r1")
	require.NoError(t, err)
	assert.True(t, result.Triggered)
}

func TestEngine_EnableRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.Enabled = false
	engine.AddRule(rule)
	require.NoError(t, engine.EnableRule("r1"))
	found, _ := engine.GetRule("r1")
	assert.True(t, found.Enabled)
}

func TestEngine_EnableRule_NotFound(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	assert.Error(t, engine.EnableRule("nonexistent"))
}

func TestEngine_DisableRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	engine.AddRule(rule)
	require.NoError(t, engine.DisableRule("r1"))
	found, _ := engine.GetRule("r1")
	assert.False(t, found.Enabled)
}

func TestEngine_DisableRule_NotFound(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	assert.Error(t, engine.DisableRule("nonexistent"))
}

func TestEngine_GetEnabledRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	r1 := NewRuleDSL("r1", "test1", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r2 := NewRuleDSL("r2", "test2", "1.0")
	r2.SetComparisonCondition("p2", OpLT, ThresholdTypeAbsolute, 50)
	r2.Enabled = false
	engine.AddRule(r1)
	engine.AddRule(r2)
	enabled := engine.GetEnabledRules()
	assert.Len(t, enabled, 1)
}

func TestEngine_GetRulesByPriority(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	r1 := NewRuleDSL("r1", "test1", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r1.Priority = 30
	r2 := NewRuleDSL("r2", "test2", "1.0")
	r2.SetComparisonCondition("p2", OpLT, ThresholdTypeAbsolute, 50)
	r2.Priority = 70
	engine.AddRule(r1)
	engine.AddRule(r2)
	rules := engine.GetRulesByPriority(50, 100)
	assert.Len(t, rules, 1)
	assert.Equal(t, "r2", rules[0].ID)
}

func TestEngine_ClearCache(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	engine.ClearCache()
	assert.Empty(t, engine.cache.values)
}

func TestEngine_Compare_AllOps(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	assert.True(t, engine.compare(10, OpGT, 5))
	assert.False(t, engine.compare(5, OpGT, 10))
	assert.True(t, engine.compare(5, OpLT, 10))
	assert.False(t, engine.compare(10, OpLT, 5))
	assert.True(t, engine.compare(10, OpGTE, 10))
	assert.True(t, engine.compare(11, OpGTE, 10))
	assert.False(t, engine.compare(9, OpGTE, 10))
	assert.True(t, engine.compare(10, OpLTE, 10))
	assert.True(t, engine.compare(9, OpLTE, 10))
	assert.False(t, engine.compare(11, OpLTE, 10))
	assert.True(t, engine.compare(10, OpEQ, 10))
	assert.False(t, engine.compare(10, OpEQ, 11))
	assert.True(t, engine.compare(10, OpNE, 11))
	assert.False(t, engine.compare(10, OpNE, 10))
	assert.False(t, engine.compare(10, Operator("??"), 10))
}

func TestEngine_Evaluate_NoProvider(t *testing.T) {
	engine := NewEngine(nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	engine.AddRule(rule)
	result, err := engine.Evaluate(context.Background(), "r1")
	require.NoError(t, err)
	assert.Error(t, result.Error)
}

func TestVersionManager_CreateVersion(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	v, err := vm.CreateVersion(rule, "initial", "created", "admin")
	require.NoError(t, err)
	assert.Equal(t, "1.0", v.Version)
	assert.Equal(t, VersionStatusActive, v.Status)
}

func TestVersionManager_CreateVersion_InvalidRule(t *testing.T) {
	vm := NewVersionManager()
	rule := &RuleDSL{}
	_, err := vm.CreateVersion(rule, "", "", "")
	assert.Error(t, err)
}

func TestVersionManager_CreateVersion_Duplicate(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	_, err := vm.CreateVersion(rule, "v1 again", "duplicate", "admin")
	assert.Error(t, err)
}

func TestVersionManager_GetVersion(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	v, err := vm.GetVersion("r1", "1.0")
	require.NoError(t, err)
	assert.Equal(t, "1.0", v.Version)
}

func TestVersionManager_GetVersion_NotFound(t *testing.T) {
	vm := NewVersionManager()
	_, err := vm.GetVersion("r1", "1.0")
	assert.Error(t, err)
}

func TestVersionManager_GetVersion_VersionNotFound(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	_, err := vm.GetVersion("r1", "2.0")
	assert.Error(t, err)
}

func TestVersionManager_GetActiveVersion(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	v, err := vm.GetActiveVersion("r1")
	require.NoError(t, err)
	assert.Equal(t, VersionStatusActive, v.Status)
}

func TestVersionManager_GetActiveVersion_NoRule(t *testing.T) {
	vm := NewVersionManager()
	_, err := vm.GetActiveVersion("nonexistent")
	assert.Error(t, err)
}

func TestVersionManager_GetHistory(t *testing.T) {
	vm := NewVersionManager()
	r1 := NewRuleDSL("r1", "test", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(r1, "v1", "created", "admin")
	r2 := NewRuleDSL("r1", "test", "2.0")
	r2.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 200)
	vm.CreateVersion(r2, "v2", "updated", "admin")
	history, err := vm.GetHistory("r1")
	require.NoError(t, err)
	assert.Len(t, history, 2)
}

func TestVersionManager_GetHistory_NoRule(t *testing.T) {
	vm := NewVersionManager()
	_, err := vm.GetHistory("nonexistent")
	assert.Error(t, err)
}

func TestVersionManager_Rollback(t *testing.T) {
	vm := NewVersionManager()
	r1 := NewRuleDSL("r1", "test", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(r1, "v1", "created", "admin")
	r2 := NewRuleDSL("r1", "test", "2.0")
	r2.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 200)
	vm.CreateVersion(r2, "v2", "updated", "admin")
	rollback, err := vm.Rollback("r1", "1.0")
	require.NoError(t, err)
	assert.Contains(t, rollback.Version, "rollback")
	assert.Equal(t, VersionStatusActive, rollback.Status)
}

func TestVersionManager_Rollback_NoRule(t *testing.T) {
	vm := NewVersionManager()
	_, err := vm.Rollback("nonexistent", "1.0")
	assert.Error(t, err)
}

func TestVersionManager_Rollback_VersionNotFound(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	_, err := vm.Rollback("r1", "99.0")
	assert.Error(t, err)
}

func TestVersionManager_CompareVersions(t *testing.T) {
	vm := NewVersionManager()
	r1 := NewRuleDSL("r1", "test", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(r1, "v1", "created", "admin")
	r2 := NewRuleDSL("r1", "test v2", "2.0")
	r2.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 200)
	vm.CreateVersion(r2, "v2", "updated", "admin")
	diff, err := vm.CompareVersions("r1", "1.0", "2.0")
	require.NoError(t, err)
	assert.True(t, diff.HasChanges)
}

func TestVersionManager_CompareVersions_Error(t *testing.T) {
	vm := NewVersionManager()
	_, err := vm.CompareVersions("nonexistent", "1.0", "2.0")
	assert.Error(t, err)
}

func TestVersionManager_DeleteVersion(t *testing.T) {
	vm := NewVersionManager()
	r1 := NewRuleDSL("r1", "test", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(r1, "v1", "created", "admin")
	r2 := NewRuleDSL("r1", "test", "2.0")
	r2.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 200)
	vm.CreateVersion(r2, "v2", "updated", "admin")
	err := vm.DeleteVersion("r1", "1.0")
	require.NoError(t, err)
	assert.Equal(t, 1, vm.GetVersionCount("r1"))
}

func TestVersionManager_DeleteVersion_Active(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	err := vm.DeleteVersion("r1", "1.0")
	assert.Error(t, err)
}

func TestVersionManager_DeleteVersion_NoRule(t *testing.T) {
	vm := NewVersionManager()
	err := vm.DeleteVersion("nonexistent", "1.0")
	assert.Error(t, err)
}

func TestVersionManager_DeleteVersion_NotFound(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	err := vm.DeleteVersion("r1", "99.0")
	assert.Error(t, err)
}

func TestVersionManager_ListAllVersions(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	all := vm.ListAllVersions()
	assert.Len(t, all, 1)
}

func TestVersionManager_ArchiveVersion(t *testing.T) {
	vm := NewVersionManager()
	r1 := NewRuleDSL("r1", "test", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(r1, "v1", "created", "admin")
	r2 := NewRuleDSL("r1", "test", "2.0")
	r2.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 200)
	vm.CreateVersion(r2, "v2", "updated", "admin")
	err := vm.ArchiveVersion("r1", "1.0")
	require.NoError(t, err)
}

func TestVersionManager_ArchiveVersion_Active(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	err := vm.ArchiveVersion("r1", "1.0")
	assert.Error(t, err)
}

func TestVersionManager_ArchiveVersion_NoRule(t *testing.T) {
	vm := NewVersionManager()
	err := vm.ArchiveVersion("nonexistent", "1.0")
	assert.Error(t, err)
}

func TestVersionManager_ArchiveVersion_NotFound(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	err := vm.ArchiveVersion("r1", "99.0")
	assert.Error(t, err)
}

func TestVersionManager_GetVersionsByTag(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	v, _ := vm.CreateVersion(rule, "v1", "created", "admin")
	vm.AddTagToVersion("r1", "1.0", "release")
	_ = v
	versions, err := vm.GetVersionsByTag("r1", "release")
	require.NoError(t, err)
	assert.Len(t, versions, 1)
}

func TestVersionManager_GetVersionsByTag_NoRule(t *testing.T) {
	vm := NewVersionManager()
	_, err := vm.GetVersionsByTag("nonexistent", "tag")
	assert.Error(t, err)
}

func TestVersionManager_AddTagToVersion(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	err := vm.AddTagToVersion("r1", "1.0", "release")
	require.NoError(t, err)
}

func TestVersionManager_AddTagToVersion_Duplicate(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	vm.AddTagToVersion("r1", "1.0", "release")
	err := vm.AddTagToVersion("r1", "1.0", "release")
	assert.NoError(t, err)
}

func TestVersionManager_AddTagToVersion_NoRule(t *testing.T) {
	vm := NewVersionManager()
	err := vm.AddTagToVersion("nonexistent", "1.0", "tag")
	assert.Error(t, err)
}

func TestVersionManager_AddTagToVersion_NotFound(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	err := vm.AddTagToVersion("r1", "99.0", "tag")
	assert.Error(t, err)
}

func TestVersionManager_RemoveTagFromVersion(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	vm.AddTagToVersion("r1", "1.0", "release")
	err := vm.RemoveTagFromVersion("r1", "1.0", "release")
	require.NoError(t, err)
}

func TestVersionManager_RemoveTagFromVersion_TagNotFound(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	err := vm.RemoveTagFromVersion("r1", "1.0", "nonexistent")
	assert.Error(t, err)
}

func TestVersionManager_RemoveTagFromVersion_NoRule(t *testing.T) {
	vm := NewVersionManager()
	err := vm.RemoveTagFromVersion("nonexistent", "1.0", "tag")
	assert.Error(t, err)
}

func TestVersionManager_RemoveTagFromVersion_VersionNotFound(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	err := vm.RemoveTagFromVersion("r1", "99.0", "tag")
	assert.Error(t, err)
}

func TestVersionManager_GetVersionCount(t *testing.T) {
	vm := NewVersionManager()
	assert.Equal(t, 0, vm.GetVersionCount("r1"))
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	assert.Equal(t, 1, vm.GetVersionCount("r1"))
}

func TestVersionManager_SearchVersions(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "initial release", "created", "admin")
	results, err := vm.SearchVersions("r1", "initial")
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestVersionManager_SearchVersions_NoRule(t *testing.T) {
	vm := NewVersionManager()
	_, err := vm.SearchVersions("nonexistent", "query")
	assert.Error(t, err)
}

func TestVersionManager_SearchVersions_ByTag(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	vm.CreateVersion(rule, "v1", "created", "admin")
	vm.AddTagToVersion("r1", "1.0", "production")
	results, err := vm.SearchVersions("r1", "production")
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestExportVersion_ImportVersion(t *testing.T) {
	vm := NewVersionManager()
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	v, _ := vm.CreateVersion(rule, "v1", "created", "admin")
	jsonStr, err := v.ExportVersion()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonStr)
	imported, err := ImportVersion(jsonStr)
	require.NoError(t, err)
	assert.Equal(t, v.Version, imported.Version)
}

func TestImportVersion_Invalid(t *testing.T) {
	_, err := ImportVersion("invalid json")
	assert.Error(t, err)
}

func TestRuleManager_CreateRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 120.0}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	require.NoError(t, rm.CreateRule(context.Background(), rule))
}

func TestRuleManager_CreateRule_Duplicate(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{"p1": 120.0}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	err := rm.CreateRule(context.Background(), rule)
	assert.Error(t, err)
}

func TestRuleManager_CreateRule_Invalid(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := &RuleDSL{}
	assert.Error(t, rm.CreateRule(context.Background(), rule))
}

func TestRuleManager_GetRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	found, err := rm.GetRule(context.Background(), "r1")
	require.NoError(t, err)
	assert.Equal(t, "r1", found.ID)
}

func TestRuleManager_GetRule_NotFound(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	_, err := rm.GetRule(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestRuleManager_UpdateRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	rule.Name = "updated"
	rule.Version = "2.0"
	require.NoError(t, rm.UpdateRule(context.Background(), rule))
}

func TestRuleManager_UpdateRule_NotFound(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	assert.Error(t, rm.UpdateRule(context.Background(), rule))
}

func TestRuleManager_DeleteRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	require.NoError(t, rm.DeleteRule(context.Background(), "r1"))
	_, err := rm.GetRule(context.Background(), "r1")
	assert.Error(t, err)
}

func TestRuleManager_DeleteRule_NotFound(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	assert.Error(t, rm.DeleteRule(context.Background(), "nonexistent"))
}

func TestRuleManager_ListRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	r1 := NewRuleDSL("r1", "test1", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r1.CreatedBy = "admin"
	r2 := NewRuleDSL("r2", "test2", "1.0")
	r2.SetComparisonCondition("p2", OpLT, ThresholdTypeAbsolute, 50)
	r2.CreatedBy = "admin"
	rm.CreateRule(context.Background(), r1)
	rm.CreateRule(context.Background(), r2)
	rules, err := rm.ListRules(context.Background())
	require.NoError(t, err)
	assert.Len(t, rules, 2)
}

func TestRuleManager_ListRulesByTag(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rule.AddTag("critical")
	rm.CreateRule(context.Background(), rule)
	rules, err := rm.ListRulesByTag(context.Background(), "critical")
	require.NoError(t, err)
	assert.Len(t, rules, 1)
}

func TestRuleManager_SearchRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "voltage alarm", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	rules, err := rm.SearchRules(context.Background(), "voltage")
	require.NoError(t, err)
	assert.Len(t, rules, 1)
}

func TestRuleManager_EnableRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rule.Enabled = false
	rm.CreateRule(context.Background(), rule)
	require.NoError(t, rm.EnableRule(context.Background(), "r1"))
	found, _ := rm.GetRule(context.Background(), "r1")
	assert.True(t, found.Enabled)
}

func TestRuleManager_EnableRule_NotFound(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	assert.Error(t, rm.EnableRule(context.Background(), "nonexistent"))
}

func TestRuleManager_DisableRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	require.NoError(t, rm.DisableRule(context.Background(), "r1"))
	found, _ := rm.GetRule(context.Background(), "r1")
	assert.False(t, found.Enabled)
}

func TestRuleManager_DisableRule_NotFound(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	assert.Error(t, rm.DisableRule(context.Background(), "nonexistent"))
}

func TestRuleManager_ValidateRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	assert.NoError(t, rm.ValidateRule(rule))
}

func TestRuleManager_ImportExportRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	r1 := NewRuleDSL("r1", "test1", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r1.CreatedBy = "admin"
	rm.CreateRule(context.Background(), r1)
	data, err := rm.ExportRules(context.Background(), []string{"r1"})
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	rm2 := NewRuleManager(engine, nil)
	imported, err := rm2.ImportRules(context.Background(), data)
	require.NoError(t, err)
	assert.Len(t, imported, 1)
}

func TestRuleManager_ExportRules_NotFound(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	_, err := rm.ExportRules(context.Background(), []string{"nonexistent"})
	assert.Error(t, err)
}

func TestRuleManager_ExportAllRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	data, err := rm.ExportAllRules(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRuleManager_CloneRule(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	cloned, err := rm.CloneRule(context.Background(), "r1", "r2", "cloned rule")
	require.NoError(t, err)
	assert.Equal(t, "r2", cloned.ID)
	assert.Equal(t, "cloned rule", cloned.Name)
}

func TestRuleManager_CloneRule_NotFound(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	_, err := rm.CloneRule(context.Background(), "nonexistent", "r2", "cloned")
	assert.Error(t, err)
}

func TestRuleManager_CloneRule_DuplicateID(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	_, err := rm.CloneRule(context.Background(), "r1", "r1", "cloned")
	assert.Error(t, err)
}

func TestRuleManager_GetRuleStatistics(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	r1 := NewRuleDSL("r1", "test1", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r1.CreatedBy = "admin"
	r1.AddTag("critical")
	r2 := NewRuleDSL("r2", "test2", "1.0")
	r2.SetComparisonCondition("p2", OpLT, ThresholdTypeAbsolute, 50)
	r2.CreatedBy = "admin"
	r2.Enabled = false
	rm.CreateRule(context.Background(), r1)
	rm.CreateRule(context.Background(), r2)
	stats, err := rm.GetRuleStatistics(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 1, stats.Enabled)
	assert.Equal(t, 1, stats.Disabled)
}

func TestRuleManager_GetVersionManager(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	assert.NotNil(t, rm.GetVersionManager())
}

func TestRuleManager_GetEngine(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	assert.Equal(t, engine, rm.GetEngine())
}

func TestRuleManager_LoadRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	require.NoError(t, rm.LoadRules(context.Background()))
}

func TestRuleManager_BatchCreateRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	r1 := NewRuleDSL("r1", "test1", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r1.CreatedBy = "admin"
	r2 := NewRuleDSL("r2", "test2", "1.0")
	r2.SetComparisonCondition("p2", OpLT, ThresholdTypeAbsolute, 50)
	r2.CreatedBy = "admin"
	created, errors := rm.BatchCreateRules(context.Background(), []*RuleDSL{r1, r2})
	assert.Len(t, created, 2)
	assert.Empty(t, errors)
}

func TestRuleManager_BatchDeleteRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	errors := rm.BatchDeleteRules(context.Background(), []string{"r1"})
	assert.Empty(t, errors)
}

func TestRuleManager_BatchEnableRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rule.Enabled = false
	rm.CreateRule(context.Background(), rule)
	errors := rm.BatchEnableRules(context.Background(), []string{"r1"})
	assert.Empty(t, errors)
}

func TestRuleManager_BatchDisableRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	errors := rm.BatchDisableRules(context.Background(), []string{"r1"})
	assert.Empty(t, errors)
}

func TestRuleManager_GetRulesByPriority(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	r1 := NewRuleDSL("r1", "test1", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r1.CreatedBy = "admin"
	r1.Priority = 30
	r2 := NewRuleDSL("r2", "test2", "1.0")
	r2.SetComparisonCondition("p2", OpLT, ThresholdTypeAbsolute, 50)
	r2.CreatedBy = "admin"
	r2.Priority = 70
	rm.CreateRule(context.Background(), r1)
	rm.CreateRule(context.Background(), r2)
	rules, err := rm.GetRulesByPriority(context.Background(), 50, 100)
	require.NoError(t, err)
	assert.Len(t, rules, 1)
}

func TestRuleManager_GetEnabledRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	r1 := NewRuleDSL("r1", "test1", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r1.CreatedBy = "admin"
	r2 := NewRuleDSL("r2", "test2", "1.0")
	r2.SetComparisonCondition("p2", OpLT, ThresholdTypeAbsolute, 50)
	r2.CreatedBy = "admin"
	r2.Enabled = false
	rm.CreateRule(context.Background(), r1)
	rm.CreateRule(context.Background(), r2)
	rules, err := rm.GetEnabledRules(context.Background())
	require.NoError(t, err)
	assert.Len(t, rules, 1)
}

func TestRuleManager_GetDisabledRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	r1 := NewRuleDSL("r1", "test1", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r1.CreatedBy = "admin"
	r2 := NewRuleDSL("r2", "test2", "1.0")
	r2.SetComparisonCondition("p2", OpLT, ThresholdTypeAbsolute, 50)
	r2.CreatedBy = "admin"
	r2.Enabled = false
	rm.CreateRule(context.Background(), r1)
	rm.CreateRule(context.Background(), r2)
	rules, err := rm.GetDisabledRules(context.Background())
	require.NoError(t, err)
	assert.Len(t, rules, 1)
}

func TestRuleManager_CountRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	assert.Equal(t, 1, rm.CountRules(context.Background()))
}

func TestRuleManager_CountEnabledRules(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	assert.Equal(t, 1, rm.CountEnabledRules(context.Background()))
}

func TestRuleManager_ImportRules_InvalidJSON(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	_, err := rm.ImportRules(context.Background(), []byte("invalid json"))
	assert.Error(t, err)
}

func TestRuleManager_ImportFromFile_NotFound(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	_, err := rm.ImportFromFile(context.Background(), "/nonexistent/file.json")
	assert.Error(t, err)
}

func TestRuleManager_ExportToFile(t *testing.T) {
	provider := &mockDataProvider{values: map[string]float64{}}
	engine := NewEngine(provider)
	rm := NewRuleManager(engine, nil)
	rule := NewRuleDSL("r1", "test", "1.0")
	rule.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	rule.CreatedBy = "admin"
	rm.CreateRule(context.Background(), rule)
	tmpFile := t.TempDir() + "/rules.json"
	err := rm.ExportToFile(context.Background(), tmpFile, []string{"r1"})
	require.NoError(t, err)
}

func TestEqualStringSlices(t *testing.T) {
	assert.True(t, equalStringSlices([]string{"a", "b"}, []string{"a", "b"}))
	assert.False(t, equalStringSlices([]string{"a", "b"}, []string{"a", "c"}))
	assert.False(t, equalStringSlices([]string{"a"}, []string{"a", "b"}))
	assert.True(t, equalStringSlices([]string{}, []string{}))
}

func TestCompareRules_NoChanges(t *testing.T) {
	r1 := NewRuleDSL("r1", "test", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r2 := NewRuleDSL("r1", "test", "1.0")
	r2.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	diff := compareRules(r1, r2)
	assert.False(t, diff.HasChanges)
}

func TestCompareRules_WithChanges(t *testing.T) {
	r1 := NewRuleDSL("r1", "test", "1.0")
	r1.SetComparisonCondition("p1", OpGT, ThresholdTypeAbsolute, 100)
	r2 := NewRuleDSL("r1", "updated", "2.0")
	r2.SetComparisonCondition("p1", OpLT, ThresholdTypeAbsolute, 50)
	r2.Enabled = false
	r2.AddTag("new")
	diff := compareRules(r1, r2)
	assert.True(t, diff.HasChanges)
}

func TestIsValidOperator(t *testing.T) {
	assert.True(t, isValidOperator(OpGT))
	assert.True(t, isValidOperator(OpLT))
	assert.True(t, isValidOperator(OpGTE))
	assert.True(t, isValidOperator(OpLTE))
	assert.True(t, isValidOperator(OpEQ))
	assert.True(t, isValidOperator(OpNE))
	assert.False(t, isValidOperator(Operator("??")))
}

func TestIsValidLogicalOperator(t *testing.T) {
	assert.True(t, isValidLogicalOperator(LogicalAND))
	assert.True(t, isValidLogicalOperator(LogicalOR))
	assert.True(t, isValidLogicalOperator(LogicalNOT))
	assert.False(t, isValidLogicalOperator(LogicalOperator("XOR")))
}
