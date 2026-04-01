package query

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QueryType 查询类型
type QueryType string

const (
	QueryTypeSelect    QueryType = "select"
	QueryTypeAggregate QueryType = "aggregate"
	QueryTypeTimeRange QueryType = "time_range"
	QueryTypeJoin      QueryType = "join"
	QueryTypeComplex   QueryType = "complex"
)

// QueryPriority 查询优先级
type QueryPriority int

const (
	PriorityLow      QueryPriority = 1
	PriorityNormal   QueryPriority = 5
	PriorityHigh     QueryPriority = 10
	PriorityCritical QueryPriority = 20
)

// QueryStatus 查询状态
type QueryStatus string

const (
	QueryStatusPending   QueryStatus = "pending"
	QueryStatusRunning   QueryStatus = "running"
	QueryStatusCompleted QueryStatus = "completed"
	QueryStatusFailed    QueryStatus = "failed"
	QueryStatusCanceled  QueryStatus = "canceled"
)

// QueryRequest 查询请求
type QueryRequest struct {
	ID          string                 `json:"id"`
	Type        QueryType              `json:"type"`
	Priority    QueryPriority          `json:"priority"`
	Database    string                 `json:"database"`
	Table       string                 `json:"table"`
	Fields      []string               `json:"fields"`
	Conditions  []QueryCondition       `json:"conditions"`
	OrderBy     []OrderByField         `json:"order_by"`
	GroupBy     []string               `json:"group_by"`
	Limit       int                    `json:"limit"`
	Offset      int                    `json:"offset"`
	TimeRange   *TimeRange             `json:"time_range"`
	Joins       []JoinClause           `json:"joins"`
	Aggregates  []AggregateField       `json:"aggregates"`
	Options     map[string]interface{} `json:"options"`
	CreatedAt   time.Time              `json:"created_at"`
	Timeout     time.Duration          `json:"timeout"`
}

// QueryCondition 查询条件
type QueryCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	AndOr    string      `json:"and_or"` // AND, OR
}

// OrderByField 排序字段
type OrderByField struct {
	Field string `json:"field"`
	Desc  bool   `json:"desc"`
}

// TimeRange 时间范围
type TimeRange struct {
	Field    string    `json:"field"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Interval string    `json:"interval"` // 1m, 5m, 1h, 1d
}

// JoinClause JOIN子句
type JoinClause struct {
	Type       string           `json:"type"` // INNER, LEFT, RIGHT
	Table      string           `json:"table"`
	Alias      string           `json:"alias"`
	Conditions []JoinCondition  `json:"conditions"`
}

// JoinCondition JOIN条件
type JoinCondition struct {
	LeftField  string `json:"left_field"`
	Operator   string `json:"operator"`
	RightField string `json:"right_field"`
}

// AggregateField 聚合字段
type AggregateField struct {
	Field      string `json:"field"`
	Function   string `json:"function"` // SUM, AVG, COUNT, MAX, MIN
	Alias      string `json:"alias"`
	Distinct   bool   `json:"distinct"`
}

// QueryResult 查询结果
type QueryResult struct {
	QueryID     string                   `json:"query_id"`
	Status      QueryStatus              `json:"status"`
	Data        []map[string]interface{} `json:"data"`
	Total       int64                    `json:"total"`
	Fields      []string                 `json:"fields"`
	ExecutionTime time.Duration          `json:"execution_time"`
	Cached      bool                     `json:"cached"`
	Error       string                   `json:"error,omitempty"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
}

// QueryPlan 查询计划
type QueryPlan struct {
	ID           string        `json:"id"`
	QueryID      string        `json:"query_id"`
	Steps        []QueryStep   `json:"steps"`
	EstimatedCost float64      `json:"estimated_cost"`
	EstimatedRows int64        `json:"estimated_rows"`
	Parallel      bool         `json:"parallel"`
	CreatedAt     time.Time    `json:"created_at"`
}

// QueryStep 查询步骤
type QueryStep struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"` // scan, filter, join, aggregate, sort, limit
	Table       string        `json:"table"`
	Conditions  []string      `json:"conditions"`
	Cost        float64       `json:"cost"`
	Rows        int64         `json:"rows"`
	Parallel    bool          `json:"parallel"`
	DependsOn   []string      `json:"depends_on"`
}

// QueryExecutor 查询执行器
type QueryExecutor struct {
	db            *gorm.DB
	planOptimizer *QueryPlanOptimizer
	parallel      *ParallelExecutor
	streamer      *ResultStreamer
	config        ExecutorConfig
	stats         *ExecutorStats
	mu            sync.RWMutex
}

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
	MaxParallelQueries   int           `json:"max_parallel_queries"`
	MaxResultRows        int           `json:"max_result_rows"`
	DefaultTimeout       time.Duration `json:"default_timeout"`
	SlowQueryThreshold   time.Duration `json:"slow_query_threshold"`
	EnableQueryPlan      bool          `json:"enable_query_plan"`
	EnableParallel       bool          `json:"enable_parallel"`
	StreamBatchSize      int           `json:"stream_batch_size"`
}

// ExecutorStats 执行器统计
type ExecutorStats struct {
	TotalQueries     int64         `json:"total_queries"`
	SuccessQueries   int64         `json:"success_queries"`
	FailedQueries    int64         `json:"failed_queries"`
	AvgExecutionTime time.Duration `json:"avg_execution_time"`
	TotalRows        int64         `json:"total_rows"`
	ParallelQueries  int64         `json:"parallel_queries"`
	CachedQueries    int64         `json:"cached_queries"`
}

// NewQueryExecutor 创建查询执行器
func NewQueryExecutor(db *gorm.DB, config ExecutorConfig) *QueryExecutor {
	if config.MaxParallelQueries <= 0 {
		config.MaxParallelQueries = 10
	}
	if config.MaxResultRows <= 0 {
		config.MaxResultRows = 100000
	}
	if config.DefaultTimeout <= 0 {
		config.DefaultTimeout = 30 * time.Second
	}
	if config.SlowQueryThreshold <= 0 {
		config.SlowQueryThreshold = 1 * time.Second
	}
	if config.StreamBatchSize <= 0 {
		config.StreamBatchSize = 1000
	}

	return &QueryExecutor{
		db:            db,
		planOptimizer: NewQueryPlanOptimizer(),
		parallel:      NewParallelExecutor(config.MaxParallelQueries),
		streamer:      NewResultStreamer(config.StreamBatchSize),
		config:        config,
		stats:         &ExecutorStats{},
	}
}

// Execute 执行查询
func (e *QueryExecutor) Execute(ctx context.Context, req *QueryRequest) (*QueryResult, error) {
	// 生成查询ID
	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	req.CreatedAt = time.Now()

	// 设置超时
	if req.Timeout <= 0 {
		req.Timeout = e.config.DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	// 生成查询计划
	var plan *QueryPlan
	var err error
	if e.config.EnableQueryPlan {
		plan, err = e.planOptimizer.Optimize(req)
		if err != nil {
			return nil, fmt.Errorf("failed to optimize query: %w", err)
		}
	}

	// 执行查询
	startTime := time.Now()
	result := &QueryResult{
		QueryID: req.ID,
		Status:  QueryStatusRunning,
	}

	// 根据查询类型选择执行方式
	if e.config.EnableParallel && plan != nil && plan.Parallel {
		result, err = e.executeParallel(ctx, req, plan)
	} else {
		result, err = e.executeSingle(ctx, req)
	}

	// 更新统计
	e.updateStats(result, time.Since(startTime), err)

	if err != nil {
		result.Status = QueryStatusFailed
		result.Error = err.Error()
		return result, err
	}

	result.Status = QueryStatusCompleted
	result.ExecutionTime = time.Since(startTime)

	return result, nil
}

// executeSingle 单线程执行查询
func (e *QueryExecutor) executeSingle(ctx context.Context, req *QueryRequest) (*QueryResult, error) {
	query := e.db.WithContext(ctx).Table(req.Table)

	// 选择字段
	if len(req.Fields) > 0 {
		query = query.Select(req.Fields)
	}

	// 应用条件
	for _, cond := range req.Conditions {
		query = e.applyCondition(query, cond)
	}

	// 时间范围
	if req.TimeRange != nil {
		query = query.Where(fmt.Sprintf("%s >= ? AND %s <= ?",
			req.TimeRange.Field, req.TimeRange.Field),
			req.TimeRange.Start, req.TimeRange.End)
	}

	// JOIN
	for _, join := range req.Joins {
		query = e.applyJoin(query, join)
	}

	// GROUP BY
	if len(req.GroupBy) > 0 {
		query = query.Group(fmt.Sprintf("%s", req.GroupBy))
	}

	// 聚合
	if len(req.Aggregates) > 0 {
		selectFields := make([]string, 0, len(req.Aggregates))
		for _, agg := range req.Aggregates {
			field := e.buildAggregateField(agg)
			selectFields = append(selectFields, field)
		}
		query = query.Select(selectFields)
	}

	// 排序
	for _, order := range req.OrderBy {
		orderStr := order.Field
		if order.Desc {
			orderStr += " DESC"
		}
		query = query.Order(orderStr)
	}

	// 分页
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}
	if req.Limit > 0 {
		if req.Limit > e.config.MaxResultRows {
			req.Limit = e.config.MaxResultRows
		}
		query = query.Limit(req.Limit)
	}

	// 执行查询
	result := &QueryResult{
		QueryID: req.ID,
		Data:    make([]map[string]interface{}, 0),
	}

	if err := query.Find(&result.Data).Error; err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// 获取总数
	countQuery := e.db.WithContext(ctx).Table(req.Table)
	for _, cond := range req.Conditions {
		countQuery = e.applyCondition(countQuery, cond)
	}
	if req.TimeRange != nil {
		countQuery = countQuery.Where(fmt.Sprintf("%s >= ? AND %s <= ?",
			req.TimeRange.Field, req.TimeRange.Field),
			req.TimeRange.Start, req.TimeRange.End)
	}
	countQuery.Count(&result.Total)

	result.Fields = e.extractFields(result.Data)

	return result, nil
}

// executeParallel 并行执行查询
func (e *QueryExecutor) executeParallel(ctx context.Context, req *QueryRequest, plan *QueryPlan) (*QueryResult, error) {
	atomic.AddInt64(&e.stats.ParallelQueries, 1)

	// 分解查询为多个子查询
	subQueries, err := e.decomposeQuery(req, plan)
	if err != nil {
		return nil, err
	}

	// 并行执行子查询
	results := make(chan *QueryResult, len(subQueries))
	errors := make(chan error, len(subQueries))

	var wg sync.WaitGroup
	for _, subReq := range subQueries {
		wg.Add(1)
		go func(r *QueryRequest) {
			defer wg.Done()
			result, err := e.executeSingle(ctx, r)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}(subReq)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// 收集结果
	finalResult := &QueryResult{
		QueryID: req.ID,
		Data:    make([]map[string]interface{}, 0),
	}

	for result := range results {
		finalResult.Data = append(finalResult.Data, result.Data...)
		finalResult.Total += result.Total
	}

	// 检查错误
	select {
	case err := <-errors:
		if err != nil {
			return nil, err
		}
	default:
	}

	// 合并后排序和分页
	finalResult.Data = e.sortAndPaginate(finalResult.Data, req)
	finalResult.Fields = e.extractFields(finalResult.Data)

	return finalResult, nil
}

// StreamExecute 流式执行查询
func (e *QueryExecutor) StreamExecute(ctx context.Context, req *QueryRequest) (<-chan *StreamRow, error) {
	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	stream := make(chan *StreamRow, e.config.StreamBatchSize)

	go func() {
		defer close(stream)

		// 分批查询
		batchSize := e.config.StreamBatchSize
		offset := 0

		for {
			batchReq := *req
			batchReq.Limit = batchSize
			batchReq.Offset = offset

			result, err := e.executeSingle(ctx, &batchReq)
			if err != nil {
				stream <- &StreamRow{Error: err}
				return
			}

			if len(result.Data) == 0 {
				return
			}

			for _, row := range result.Data {
				select {
				case <-ctx.Done():
					return
				case stream <- &StreamRow{Data: row}:
				}
			}

			if len(result.Data) < batchSize {
				return
			}

			offset += batchSize
		}
	}()

	return stream, nil
}

// StreamRow 流式行数据
type StreamRow struct {
	Data  map[string]interface{}
	Error error
}

// applyCondition 应用查询条件
func (e *QueryExecutor) applyCondition(query *gorm.DB, cond QueryCondition) *gorm.DB {
	switch cond.Operator {
	case "=":
		return query.Where(fmt.Sprintf("%s = ?", cond.Field), cond.Value)
	case "!=":
		return query.Where(fmt.Sprintf("%s != ?", cond.Field), cond.Value)
	case ">":
		return query.Where(fmt.Sprintf("%s > ?", cond.Field), cond.Value)
	case ">=":
		return query.Where(fmt.Sprintf("%s >= ?", cond.Field), cond.Value)
	case "<":
		return query.Where(fmt.Sprintf("%s < ?", cond.Field), cond.Value)
	case "<=":
		return query.Where(fmt.Sprintf("%s <= ?", cond.Field), cond.Value)
	case "IN":
		return query.Where(fmt.Sprintf("%s IN ?", cond.Field), cond.Value)
	case "NOT IN":
		return query.Where(fmt.Sprintf("%s NOT IN ?", cond.Field), cond.Value)
	case "LIKE":
		return query.Where(fmt.Sprintf("%s LIKE ?", cond.Field), cond.Value)
	case "BETWEEN":
		if vals, ok := cond.Value.([]interface{}); ok && len(vals) == 2 {
			return query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", cond.Field), vals[0], vals[1])
		}
	case "IS NULL":
		return query.Where(fmt.Sprintf("%s IS NULL", cond.Field))
	case "IS NOT NULL":
		return query.Where(fmt.Sprintf("%s IS NOT NULL", cond.Field))
	}
	return query
}

// applyJoin 应用JOIN
func (e *QueryExecutor) applyJoin(query *gorm.DB, join JoinClause) *gorm.DB {
	joinConditions := make([]string, 0)
	joinArgs := make([]interface{}, 0)

	for _, cond := range join.Conditions {
		joinConditions = append(joinConditions,
			fmt.Sprintf("%s %s %s", cond.LeftField, cond.Operator, cond.RightField))
	}

	joinStr := ""
	if len(joinConditions) > 0 {
		for i, c := range joinConditions {
			if i > 0 {
				joinStr += " AND "
			}
			joinStr += c
		}
	}

	tableAlias := join.Table
	if join.Alias != "" {
		tableAlias = join.Table + " AS " + join.Alias
	}

	switch join.Type {
	case "INNER":
		return query.Joins(fmt.Sprintf("INNER JOIN %s ON %s", tableAlias, joinStr), joinArgs...)
	case "LEFT":
		return query.Joins(fmt.Sprintf("LEFT JOIN %s ON %s", tableAlias, joinStr), joinArgs...)
	case "RIGHT":
		return query.Joins(fmt.Sprintf("RIGHT JOIN %s ON %s", tableAlias, joinStr), joinArgs...)
	}

	return query
}

// buildAggregateField 构建聚合字段
func (e *QueryExecutor) buildAggregateField(agg AggregateField) string {
	field := agg.Field
	if agg.Distinct {
		field = "DISTINCT " + field
	}

	expr := fmt.Sprintf("%s(%s)", agg.Function, field)
	if agg.Alias != "" {
		expr += " AS " + agg.Alias
	}

	return expr
}

// decomposeQuery 分解查询
func (e *QueryExecutor) decomposeQuery(req *QueryRequest, plan *QueryPlan) ([]*QueryRequest, error) {
	// 根据时间范围分解查询
	if req.TimeRange == nil || req.TimeRange.Interval == "" {
		return []*QueryRequest{req}, nil
	}

	interval, err := time.ParseDuration(req.TimeRange.Interval)
	if err != nil {
		return nil, fmt.Errorf("invalid interval: %w", err)
	}

	var subQueries []*QueryRequest
	start := req.TimeRange.Start
	end := req.TimeRange.End

	for start.Before(end) {
		subEnd := start.Add(interval)
		if subEnd.After(end) {
			subEnd = end
		}

		subReq := *req
		subReq.TimeRange = &TimeRange{
			Field:    req.TimeRange.Field,
			Start:    start,
			End:      subEnd,
			Interval: req.TimeRange.Interval,
		}
		subQueries = append(subQueries, &subReq)

		start = subEnd
	}

	return subQueries, nil
}

// sortAndPaginate 排序和分页
func (e *QueryExecutor) sortAndPaginate(data []map[string]interface{}, req *QueryRequest) []map[string]interface{} {
	// 简单排序实现
	if len(req.OrderBy) > 0 {
		// 使用稳定排序
		for i := 0; i < len(data)-1; i++ {
			for j := i + 1; j < len(data); j++ {
				if e.compareRows(data[i], data[j], req.OrderBy) > 0 {
					data[i], data[j] = data[j], data[i]
				}
			}
		}
	}

	// 分页
	start := req.Offset
	if start >= len(data) {
		return []map[string]interface{}{}
	}

	end := len(data)
	if req.Limit > 0 && start+req.Limit < end {
		end = start + req.Limit
	}

	return data[start:end]
}

// compareRows 比较两行数据
func (e *QueryExecutor) compareRows(a, b map[string]interface{}, orderBy []OrderByField) int {
	for _, order := range orderBy {
		aVal, aOk := a[order.Field]
		bVal, bOk := b[order.Field]

		if !aOk && !bOk {
			continue
		}
		if !aOk {
			return 1
		}
		if !bOk {
			return -1
		}

		cmp := e.compareValues(aVal, bVal)
		if cmp != 0 {
			if order.Desc {
				return -cmp
			}
			return cmp
		}
	}
	return 0
}

// compareValues 比较两个值
func (e *QueryExecutor) compareValues(a, b interface{}) int {
	switch aVal := a.(type) {
	case int:
		if bVal, ok := b.(int); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
		}
	case int64:
		if bVal, ok := b.(int64); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
		}
	case float64:
		if bVal, ok := b.(float64); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
		}
	case string:
		if bVal, ok := b.(string); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
		}
	case time.Time:
		if bVal, ok := b.(time.Time); ok {
			if aVal.Before(bVal) {
				return -1
			} else if aVal.After(bVal) {
				return 1
			}
		}
	}
	return 0
}

// extractFields 提取字段名
func (e *QueryExecutor) extractFields(data []map[string]interface{}) []string {
	if len(data) == 0 {
		return []string{}
	}

	fields := make([]string, 0, len(data[0]))
	for field := range data[0] {
		fields = append(fields, field)
	}
	return fields
}

// updateStats 更新统计
func (e *QueryExecutor) updateStats(result *QueryResult, duration time.Duration, err error) {
	atomic.AddInt64(&e.stats.TotalQueries, 1)

	if err != nil {
		atomic.AddInt64(&e.stats.FailedQueries, 1)
	} else {
		atomic.AddInt64(&e.stats.SuccessQueries, 1)
		atomic.AddInt64(&e.stats.TotalRows, int64(len(result.Data)))
	}

	// 更新平均执行时间
	for {
		old := atomic.LoadInt64((*int64)(&e.stats.AvgExecutionTime))
		newAvg := (old + int64(duration)) / 2
		if atomic.CompareAndSwapInt64((*int64)(&e.stats.AvgExecutionTime), old, newAvg) {
			break
		}
	}
}

// GetStats 获取统计信息
func (e *QueryExecutor) GetStats() *ExecutorStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stats := &ExecutorStats{
		TotalQueries:     atomic.LoadInt64(&e.stats.TotalQueries),
		SuccessQueries:   atomic.LoadInt64(&e.stats.SuccessQueries),
		FailedQueries:    atomic.LoadInt64(&e.stats.FailedQueries),
		AvgExecutionTime: e.stats.AvgExecutionTime,
		TotalRows:        atomic.LoadInt64(&e.stats.TotalRows),
		ParallelQueries:  atomic.LoadInt64(&e.stats.ParallelQueries),
		CachedQueries:    atomic.LoadInt64(&e.stats.CachedQueries),
	}

	return stats
}

// QueryPlanOptimizer 查询计划优化器
type QueryPlanOptimizer struct {
	rules []OptimizationRule
}

// OptimizationRule 优化规则
type OptimizationRule interface {
	Name() string
	Apply(plan *QueryPlan) (*QueryPlan, error)
}

// NewQueryPlanOptimizer 创建查询计划优化器
func NewQueryPlanOptimizer() *QueryPlanOptimizer {
	return &QueryPlanOptimizer{
		rules: []OptimizationRule{
			&IndexScanRule{},
			&PushDownFilterRule{},
			&JoinOrderRule{},
			&ParallelExecutionRule{},
		},
	}
}

// Optimize 优化查询计划
func (o *QueryPlanOptimizer) Optimize(req *QueryRequest) (*QueryPlan, error) {
	// 生成初始计划
	plan := o.generateInitialPlan(req)

	// 应用优化规则
	for _, rule := range o.rules {
		var err error
		plan, err = rule.Apply(plan)
		if err != nil {
			// 优化失败不影响执行
			continue
		}
	}

	return plan, nil
}

// generateInitialPlan 生成初始查询计划
func (o *QueryPlanOptimizer) generateInitialPlan(req *QueryRequest) *QueryPlan {
	plan := &QueryPlan{
		ID:        uuid.New().String(),
		QueryID:   req.ID,
		Steps:     make([]QueryStep, 0),
		CreatedAt: time.Now(),
	}

	// 扫描步骤
	scanStep := QueryStep{
		ID:     uuid.New().String(),
		Type:   "scan",
		Table:  req.Table,
		Cost:   1.0,
		Rows:    1000,
	}
	plan.Steps = append(plan.Steps, scanStep)

	// 过滤步骤
	if len(req.Conditions) > 0 || req.TimeRange != nil {
		filterStep := QueryStep{
			ID:        uuid.New().String(),
			Type:      "filter",
			Conditions: o.buildConditions(req),
			Cost:      0.5,
			Rows:       100,
			DependsOn:  []string{scanStep.ID},
		}
		plan.Steps = append(plan.Steps, filterStep)
	}

	// JOIN步骤
	for _, join := range req.Joins {
		joinStep := QueryStep{
			ID:       uuid.New().String(),
			Type:     "join",
			Table:    join.Table,
			Cost:     2.0,
			Rows:      500,
			DependsOn: []string{plan.Steps[len(plan.Steps)-1].ID},
		}
		plan.Steps = append(plan.Steps, joinStep)
	}

	// 聚合步骤
	if len(req.Aggregates) > 0 || len(req.GroupBy) > 0 {
		aggStep := QueryStep{
			ID:       uuid.New().String(),
			Type:     "aggregate",
			Cost:     1.0,
			Rows:      10,
			DependsOn: []string{plan.Steps[len(plan.Steps)-1].ID},
		}
		plan.Steps = append(plan.Steps, aggStep)
	}

	// 排序步骤
	if len(req.OrderBy) > 0 {
		sortStep := QueryStep{
			ID:       uuid.New().String(),
			Type:     "sort",
			Cost:     0.5,
			DependsOn: []string{plan.Steps[len(plan.Steps)-1].ID},
		}
		plan.Steps = append(plan.Steps, sortStep)
	}

	// 计算总成本
	for _, step := range plan.Steps {
		plan.EstimatedCost += step.Cost
		if step.Rows > plan.EstimatedRows {
			plan.EstimatedRows = step.Rows
		}
	}

	return plan
}

// buildConditions 构建条件字符串
func (o *QueryPlanOptimizer) buildConditions(req *QueryRequest) []string {
	conditions := make([]string, 0)

	for _, cond := range req.Conditions {
		conditions = append(conditions,
			fmt.Sprintf("%s %s %v", cond.Field, cond.Operator, cond.Value))
	}

	if req.TimeRange != nil {
		conditions = append(conditions,
			fmt.Sprintf("%s BETWEEN %s AND %s",
				req.TimeRange.Field, req.TimeRange.Start, req.TimeRange.End))
	}

	return conditions
}

// IndexScanRule 索引扫描规则
type IndexScanRule struct{}

func (r *IndexScanRule) Name() string { return "index_scan" }

func (r *IndexScanRule) Apply(plan *QueryPlan) (*QueryPlan, error) {
	// 检查是否可以使用索引扫描
	for i, step := range plan.Steps {
		if step.Type == "scan" && len(step.Conditions) > 0 {
			plan.Steps[i].Type = "index_scan"
			plan.Steps[i].Cost *= 0.1 // 索引扫描成本更低
		}
	}
	return plan, nil
}

// PushDownFilterRule 过滤下推规则
type PushDownFilterRule struct{}

func (r *PushDownFilterRule) Name() string { return "push_down_filter" }

func (r *PushDownFilterRule) Apply(plan *QueryPlan) (*QueryPlan, error) {
	// 将过滤条件尽可能下推到扫描阶段
	for i, step := range plan.Steps {
		if step.Type == "filter" && i > 0 {
			prevStep := &plan.Steps[i-1]
			if prevStep.Type == "scan" || prevStep.Type == "index_scan" {
				// 合并过滤条件到扫描步骤
				prevStep.Conditions = append(prevStep.Conditions, step.Conditions...)
				prevStep.Rows = step.Rows
				// 移除过滤步骤
				plan.Steps = append(plan.Steps[:i], plan.Steps[i+1:]...)
				break
			}
		}
	}
	return plan, nil
}

// JoinOrderRule JOIN顺序优化规则
type JoinOrderRule struct{}

func (r *JoinOrderRule) Name() string { return "join_order" }

func (r *JoinOrderRule) Apply(plan *QueryPlan) (*QueryPlan, error) {
	// 根据表大小调整JOIN顺序
	// 这里简化实现，实际需要统计信息
	return plan, nil
}

// ParallelExecutionRule 并行执行规则
type ParallelExecutionRule struct{}

func (r *ParallelExecutionRule) Name() string { return "parallel_execution" }

func (r *ParallelExecutionRule) Apply(plan *QueryPlan) (*QueryPlan, error) {
	// 检查是否适合并行执行
	if plan.EstimatedRows > 10000 || plan.EstimatedCost > 5.0 {
		plan.Parallel = true
		for i := range plan.Steps {
			if plan.Steps[i].Type == "scan" {
				plan.Steps[i].Parallel = true
			}
		}
	}
	return plan, nil
}

// ParallelExecutor 并行执行器
type ParallelExecutor struct {
	maxWorkers int
	workerPool chan struct{}
}

// NewParallelExecutor 创建并行执行器
func NewParallelExecutor(maxWorkers int) *ParallelExecutor {
	return &ParallelExecutor{
		maxWorkers: maxWorkers,
		workerPool: make(chan struct{}, maxWorkers),
	}
}

// Execute 并行执行任务
func (p *ParallelExecutor) Execute(ctx context.Context, tasks []func() error) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go func(t func() error) {
			defer wg.Done()

			// 获取worker
			p.workerPool <- struct{}{}
			defer func() { <-p.workerPool }()

			if err := t(); err != nil {
				select {
				case errChan <- err:
				default:
				}
			}
		}(task)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// ResultStreamer 结果流式处理器
type ResultStreamer struct {
	batchSize int
}

// NewResultStreamer 创建结果流式处理器
func NewResultStreamer(batchSize int) *ResultStreamer {
	return &ResultStreamer{batchSize: batchSize}
}

// Stream 流式返回结果
func (s *ResultStreamer) Stream(ctx context.Context, query *gorm.DB, handler func([]map[string]interface{}) error) error {
	offset := 0

	for {
		var batch []map[string]interface{}
		if err := query.Offset(offset).Limit(s.batchSize).Find(&batch).Error; err != nil {
			return err
		}

		if len(batch) == 0 {
			return nil
		}

		if err := handler(batch); err != nil {
			return err
		}

		if len(batch) < s.batchSize {
			return nil
		}

		offset += s.batchSize

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

// QueryBuilder 查询构建器
type QueryBuilder struct {
	request *QueryRequest
}

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		request: &QueryRequest{
			Fields:     make([]string, 0),
			Conditions: make([]QueryCondition, 0),
			OrderBy:    make([]OrderByField, 0),
			GroupBy:    make([]string, 0),
			Joins:      make([]JoinClause, 0),
			Aggregates: make([]AggregateField, 0),
			Options:    make(map[string]interface{}),
		},
	}
}

// Table 设置表名
func (b *QueryBuilder) Table(table string) *QueryBuilder {
	b.request.Table = table
	return b
}

// Select 设置查询字段
func (b *QueryBuilder) Select(fields ...string) *QueryBuilder {
	b.request.Fields = append(b.request.Fields, fields...)
	return b
}

// Where 添加查询条件
func (b *QueryBuilder) Where(field, operator string, value interface{}) *QueryBuilder {
	b.request.Conditions = append(b.request.Conditions, QueryCondition{
		Field:    field,
		Operator: operator,
		Value:    value,
		AndOr:    "AND",
	})
	return b
}

// OrWhere 添加OR条件
func (b *QueryBuilder) OrWhere(field, operator string, value interface{}) *QueryBuilder {
	b.request.Conditions = append(b.request.Conditions, QueryCondition{
		Field:    field,
		Operator: operator,
		Value:    value,
		AndOr:    "OR",
	})
	return b
}

// TimeRange 设置时间范围
func (b *QueryBuilder) TimeRange(field string, start, end time.Time) *QueryBuilder {
	b.request.TimeRange = &TimeRange{
		Field: field,
		Start: start,
		End:   end,
	}
	return b
}

// Join 添加JOIN
func (b *QueryBuilder) Join(joinType, table, alias string, conditions ...JoinCondition) *QueryBuilder {
	b.request.Joins = append(b.request.Joins, JoinClause{
		Type:       joinType,
		Table:      table,
		Alias:      alias,
		Conditions: conditions,
	})
	return b
}

// GroupBy 设置分组
func (b *QueryBuilder) GroupBy(fields ...string) *QueryBuilder {
	b.request.GroupBy = append(b.request.GroupBy, fields...)
	return b
}

// OrderBy 设置排序
func (b *QueryBuilder) OrderBy(field string, desc bool) *QueryBuilder {
	b.request.OrderBy = append(b.request.OrderBy, OrderByField{
		Field: field,
		Desc:  desc,
	})
	return b
}

// Limit 设置限制
func (b *QueryBuilder) Limit(limit int) *QueryBuilder {
	b.request.Limit = limit
	return b
}

// Offset 设置偏移
func (b *QueryBuilder) Offset(offset int) *QueryBuilder {
	b.request.Offset = offset
	return b
}

// Aggregate 添加聚合
func (b *QueryBuilder) Aggregate(field, function, alias string) *QueryBuilder {
	b.request.Aggregates = append(b.request.Aggregates, AggregateField{
		Field:    field,
		Function: function,
		Alias:    alias,
	})
	return b
}

// Priority 设置优先级
func (b *QueryBuilder) Priority(priority QueryPriority) *QueryBuilder {
	b.request.Priority = priority
	return b
}

// Timeout 设置超时
func (b *QueryBuilder) Timeout(timeout time.Duration) *QueryBuilder {
	b.request.Timeout = timeout
	return b
}

// Build 构建查询请求
func (b *QueryBuilder) Build() *QueryRequest {
	return b.request
}

// QueryError 查询错误
type QueryError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	QueryID string `json:"query_id"`
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("query error [%s]: %s", e.Code, e.Message)
}

// 常见查询错误
var (
	ErrQueryTimeout     = &QueryError{Code: "TIMEOUT", Message: "query timeout"}
	ErrQueryCanceled    = &QueryError{Code: "CANCELED", Message: "query canceled"}
	ErrQueryTooComplex  = &QueryError{Code: "TOO_COMPLEX", Message: "query too complex"}
	ErrQueryInvalid     = &QueryError{Code: "INVALID", Message: "invalid query"}
	ErrQueryRateLimited = &QueryError{Code: "RATE_LIMITED", Message: "query rate limited"}
)

// IsQueryError 判断是否为查询错误
func IsQueryError(err error) bool {
	var queryErr *QueryError
	return errors.As(err, &queryErr)
}

// GetQueryErrorCode 获取查询错误码
func GetQueryErrorCode(err error) string {
	var queryErr *QueryError
	if errors.As(err, &queryErr) {
		return queryErr.Code
	}
	return "UNKNOWN"
}

// QueryRequestJSON JSON序列化查询请求
func QueryRequestJSON(req *QueryRequest) (string, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParseQueryRequestJSON 从JSON解析查询请求
func ParseQueryRequestJSON(jsonStr string) (*QueryRequest, error) {
	var req QueryRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		return nil, err
	}
	return &req, nil
}
