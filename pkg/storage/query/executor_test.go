package query

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryBuilder(t *testing.T) {
	now := time.Now()
	req := NewQueryBuilder().
		Table("devices").
		Select("id", "name", "status").
		Where("status", "=", "active").
		Where("region", "IN", []string{"east", "west"}).
		OrWhere("priority", ">", 5).
		TimeRange("created_at", now.Add(-24*time.Hour), now).
		Join("INNER", "stations", "s", JoinCondition{
			LeftField:  "devices.station_id",
			Operator:   "=",
			RightField: "s.id",
		}).
		GroupBy("status").
		OrderBy("created_at", true).
		Limit(100).
		Offset(0).
		Aggregate("power", "SUM", "total_power").
		Priority(PriorityHigh).
		Timeout(30 * time.Second).
		Build()

	require.NotNil(t, req)
	assert.Equal(t, "devices", req.Table)
	assert.Equal(t, 3, len(req.Fields))
	assert.Equal(t, 3, len(req.Conditions))
	assert.NotNil(t, req.TimeRange)
	assert.Equal(t, 1, len(req.Joins))
	assert.Equal(t, 1, len(req.GroupBy))
	assert.Equal(t, 1, len(req.OrderBy))
	assert.Equal(t, 100, req.Limit)
	assert.Equal(t, 1, len(req.Aggregates))
	assert.Equal(t, PriorityHigh, req.Priority)
	assert.Equal(t, 30*time.Second, req.Timeout)
}

func TestQueryBuilder_Defaults(t *testing.T) {
	req := NewQueryBuilder().Table("test").Build()
	require.NotNil(t, req)
	assert.Equal(t, "test", req.Table)
	assert.Equal(t, 0, len(req.Fields))
	assert.Equal(t, 0, len(req.Conditions))
}

func TestQueryType_Constants(t *testing.T) {
	assert.Equal(t, QueryType("select"), QueryTypeSelect)
	assert.Equal(t, QueryType("aggregate"), QueryTypeAggregate)
	assert.Equal(t, QueryType("time_range"), QueryTypeTimeRange)
	assert.Equal(t, QueryType("join"), QueryTypeJoin)
	assert.Equal(t, QueryType("complex"), QueryTypeComplex)
}

func TestQueryPriority_Constants(t *testing.T) {
	assert.Equal(t, QueryPriority(1), PriorityLow)
	assert.Equal(t, QueryPriority(5), PriorityNormal)
	assert.Equal(t, QueryPriority(10), PriorityHigh)
	assert.Equal(t, QueryPriority(20), PriorityCritical)
}

func TestQueryStatus_Constants(t *testing.T) {
	assert.Equal(t, QueryStatus("pending"), QueryStatusPending)
	assert.Equal(t, QueryStatus("running"), QueryStatusRunning)
	assert.Equal(t, QueryStatus("completed"), QueryStatusCompleted)
	assert.Equal(t, QueryStatus("failed"), QueryStatusFailed)
	assert.Equal(t, QueryStatus("canceled"), QueryStatusCanceled)
}

func TestQueryError(t *testing.T) {
	err := &QueryError{Code: "TEST", Message: "test error", QueryID: "q1"}
	assert.Equal(t, "query error [TEST]: test error", err.Error())
}

func TestIsQueryError(t *testing.T) {
	assert.True(t, IsQueryError(ErrQueryTimeout))
	assert.True(t, IsQueryError(ErrQueryCanceled))
	assert.True(t, IsQueryError(ErrQueryTooComplex))
	assert.True(t, IsQueryError(ErrQueryInvalid))
	assert.True(t, IsQueryError(ErrQueryRateLimited))
	assert.False(t, IsQueryError(errors.New("regular error")))
}

func TestGetQueryErrorCode(t *testing.T) {
	assert.Equal(t, "TIMEOUT", GetQueryErrorCode(ErrQueryTimeout))
	assert.Equal(t, "CANCELED", GetQueryErrorCode(ErrQueryCanceled))
	assert.Equal(t, "UNKNOWN", GetQueryErrorCode(errors.New("regular error")))
}

func TestQueryRequestJSON(t *testing.T) {
	req := &QueryRequest{
		ID:       "test-1",
		Type:     QueryTypeSelect,
		Table:    "devices",
		Fields:   []string{"id", "name"},
		Limit:    10,
		Priority: PriorityNormal,
	}

	jsonStr, err := QueryRequestJSON(req)
	require.NoError(t, err)
	assert.Contains(t, jsonStr, "devices")

	parsed, err := ParseQueryRequestJSON(jsonStr)
	require.NoError(t, err)
	assert.Equal(t, "test-1", parsed.ID)
	assert.Equal(t, "devices", parsed.Table)
	assert.Equal(t, 10, parsed.Limit)
}

func TestParseQueryRequestJSON_Invalid(t *testing.T) {
	_, err := ParseQueryRequestJSON("not json")
	assert.Error(t, err)
}

func TestQueryPlanOptimizer(t *testing.T) {
	optimizer := NewQueryPlanOptimizer()
	require.NotNil(t, optimizer)

	req := NewQueryBuilder().
		Table("devices").
		Where("status", "=", "active").
		OrderBy("id", false).
		Build()

	plan, err := optimizer.Optimize(req)
	require.NoError(t, err)
	assert.NotNil(t, plan)
	assert.True(t, len(plan.Steps) > 0)
	assert.True(t, plan.EstimatedCost > 0)
}

func TestQueryPlanOptimizer_WithJoins(t *testing.T) {
	optimizer := NewQueryPlanOptimizer()
	req := NewQueryBuilder().
		Table("devices").
		Join("INNER", "stations", "s", JoinCondition{
			LeftField:  "devices.station_id",
			Operator:   "=",
			RightField: "s.id",
		}).
		Build()

	plan, err := optimizer.Optimize(req)
	require.NoError(t, err)
	assert.NotNil(t, plan)
}

func TestQueryPlanOptimizer_WithAggregates(t *testing.T) {
	optimizer := NewQueryPlanOptimizer()
	req := NewQueryBuilder().
		Table("devices").
		Aggregate("power", "SUM", "total_power").
		GroupBy("station_id").
		Build()

	plan, err := optimizer.Optimize(req)
	require.NoError(t, err)
	assert.NotNil(t, plan)
}

func TestQueryPlanOptimizer_WithTimeRange(t *testing.T) {
	optimizer := NewQueryPlanOptimizer()
	req := NewQueryBuilder().
		Table("devices").
		TimeRange("created_at", time.Now().Add(-24*time.Hour), time.Now()).
		Build()

	plan, err := optimizer.Optimize(req)
	require.NoError(t, err)
	assert.NotNil(t, plan)
}

func TestParallelExecutor(t *testing.T) {
	pe := NewParallelExecutor(4)
	require.NotNil(t, pe)

	tasks := []func() error{
		func() error { return nil },
		func() error { return nil },
		func() error { return nil },
	}
	err := pe.Execute(context.Background(), tasks)
	require.NoError(t, err)
}

func TestNewResultStreamer(t *testing.T) {
	rs := NewResultStreamer(100)
	require.NotNil(t, rs)
	assert.Equal(t, 100, rs.batchSize)
}

func TestQueryResult_Struct(t *testing.T) {
	result := &QueryResult{
		QueryID:       "q1",
		Status:        QueryStatusCompleted,
		Data:          []map[string]interface{}{{"id": 1}},
		Total:         1,
		Fields:        []string{"id"},
		ExecutionTime: 100 * time.Millisecond,
		Cached:        false,
	}
	assert.Equal(t, "q1", result.QueryID)
	assert.Equal(t, QueryStatusCompleted, result.Status)
	assert.Equal(t, int64(1), result.Total)
}

func TestQueryPlan_Struct(t *testing.T) {
	plan := &QueryPlan{
		ID:            "p1",
		QueryID:       "q1",
		Steps:         []QueryStep{{ID: "s1", Type: "scan"}},
		EstimatedCost: 1.5,
		EstimatedRows: 1000,
		Parallel:      false,
	}
	assert.Equal(t, "p1", plan.ID)
	assert.Equal(t, 1, len(plan.Steps))
}

func TestQueryStep_Struct(t *testing.T) {
	step := &QueryStep{
		ID:       "s1",
		Type:     "scan",
		Table:    "devices",
		Cost:     1.0,
		Rows:      1000,
		Parallel: false,
	}
	assert.Equal(t, "s1", step.ID)
	assert.Equal(t, "scan", step.Type)
}

func TestExecutorConfig_Defaults(t *testing.T) {
	cfg := ExecutorConfig{}
	assert.Equal(t, 0, cfg.MaxParallelQueries)
	assert.Equal(t, 0, cfg.MaxResultRows)
}

func TestExecutorStats_Struct(t *testing.T) {
	stats := &ExecutorStats{
		TotalQueries:     100,
		SuccessQueries:   95,
		FailedQueries:    5,
		TotalRows:        10000,
		ParallelQueries:  10,
		CachedQueries:    20,
	}
	assert.Equal(t, int64(100), stats.TotalQueries)
	assert.Equal(t, int64(95), stats.SuccessQueries)
}

func TestStreamRow_Struct(t *testing.T) {
	row := &StreamRow{
		Data:  map[string]interface{}{"id": 1},
		Error: nil,
	}
	assert.NotNil(t, row.Data)
}

func TestQueryCondition_Struct(t *testing.T) {
	cond := QueryCondition{
		Field:    "status",
		Operator: "=",
		Value:    "active",
		AndOr:    "AND",
	}
	assert.Equal(t, "status", cond.Field)
	assert.Equal(t, "=", cond.Operator)
}

func TestOrderByField_Struct(t *testing.T) {
	field := OrderByField{Field: "created_at", Desc: true}
	assert.Equal(t, "created_at", field.Field)
	assert.True(t, field.Desc)
}

func TestTimeRange_Struct(t *testing.T) {
	tr := &TimeRange{
		Field:    "created_at",
		Start:    time.Now().Add(-24 * time.Hour),
		End:      time.Now(),
		Interval: "1h",
	}
	assert.Equal(t, "created_at", tr.Field)
	assert.Equal(t, "1h", tr.Interval)
}

func TestJoinClause_Struct(t *testing.T) {
	jc := JoinClause{
		Type:  "INNER",
		Table: "stations",
		Alias: "s",
		Conditions: []JoinCondition{
			{LeftField: "d.station_id", Operator: "=", RightField: "s.id"},
		},
	}
	assert.Equal(t, "INNER", jc.Type)
	assert.Equal(t, 1, len(jc.Conditions))
}

func TestAggregateField_Struct(t *testing.T) {
	af := AggregateField{
		Field:    "power",
		Function: "SUM",
		Alias:    "total_power",
		Distinct: false,
	}
	assert.Equal(t, "SUM", af.Function)
}

func TestQueryRequest_JSON_Roundtrip(t *testing.T) {
	now := time.Now()
	req := &QueryRequest{
		ID:       "test-json",
		Type:     QueryTypeSelect,
		Priority: PriorityHigh,
		Database: "testdb",
		Table:    "devices",
		Fields:   []string{"id", "name"},
		Conditions: []QueryCondition{
			{Field: "status", Operator: "=", Value: "active", AndOr: "AND"},
		},
		OrderBy:  []OrderByField{{Field: "id", Desc: false}},
		GroupBy:  []string{"status"},
		Limit:    100,
		Offset:   0,
		TimeRange: &TimeRange{Field: "ts", Start: now, End: now, Interval: "1h"},
		Joins: []JoinClause{{
			Type:       "INNER",
			Table:      "stations",
			Alias:      "s",
			Conditions: []JoinCondition{{LeftField: "d.sid", Operator: "=", RightField: "s.id"}},
		}},
		Aggregates: []AggregateField{{Field: "power", Function: "SUM", Alias: "total"}},
		Options:    map[string]interface{}{"cache": true},
		CreatedAt:  now,
		Timeout:    30 * time.Second,
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed QueryRequest
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "test-json", parsed.ID)
	assert.Equal(t, "devices", parsed.Table)
}
