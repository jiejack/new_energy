package storage

import (
	"crypto/md5"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
)

const (
	defaultBatchSize    = 1000
	defaultFlushInterval = 5 * time.Second
)

type ClickHouseStorage struct {
	config         types.StorageConfig
	batchBuffer    []*types.DataPoint
	batchSize      int
	flushInterval  time.Duration
	mu             sync.Mutex
	stopChan       chan struct{}
	started        bool
	materializedViews []string
	preAggregationTables []string
	preAggregationRules []PreAggregationRule
	queryCache     map[string]*CacheItem
	cacheTTL       time.Duration
	cacheMaxSize   int
}

type ClickHouseTableSchema struct {
	Name        string
	Columns     []string
	Engine      string
	OrderBy     []string
	PartitionBy string
	TTL         string
}

type PreAggregationRule struct {
	ID           string
	SourceTable  string
	TargetTable  string
	Aggregation  string
	GroupBy      []string
	TimeInterval string
	Enabled      bool
	CreatedAt    time.Time
}

type CacheItem struct {
	Key        string
	Value      interface{}
	CreatedAt  time.Time
	ExpiresAt  time.Time
	AccessCount int
}

func NewClickHouseStorage() *ClickHouseStorage {
	return &ClickHouseStorage{
		batchSize:     defaultBatchSize,
		flushInterval: defaultFlushInterval,
		batchBuffer:   make([]*types.DataPoint, 0, defaultBatchSize),
		stopChan:      make(chan struct{}),
		materializedViews: make([]string, 0),
		preAggregationTables: make([]string, 0),
		preAggregationRules: make([]PreAggregationRule, 0),
		queryCache:     make(map[string]*CacheItem),
		cacheTTL:       5 * time.Minute,
		cacheMaxSize:   1000,
	}
}

func (s *ClickHouseStorage) Init(config types.StorageConfig) error {
	if config.Type != "clickhouse" {
		return &types.Error{
			Code:    types.ErrCodeInvalidConfig,
			Message: fmt.Sprintf("invalid storage type: %s, expected clickhouse", config.Type),
		}
	}

	fmt.Printf("Initializing ClickHouse storage with config: %+v\n", config)

	if config.BatchSize > 0 {
		s.batchSize = config.BatchSize
	}
	if config.FlushInterval > 0 {
		s.flushInterval = time.Duration(config.FlushInterval) * time.Second
	}

	schema := s.getDefaultTableSchema(config.Table)
	fmt.Printf("Creating table if not exists: %s with schema: %+v\n", config.Table, schema)

	s.config = config
	s.started = true

	go s.flushLoop()

	return nil
}

func (s *ClickHouseStorage) getDefaultTableSchema(tableName string) ClickHouseTableSchema {
	return ClickHouseTableSchema{
		Name:    tableName,
		Columns: []string{
			"timestamp DateTime",
			"device_id String",
			"station_id String",
			"metric_name String",
			"metric_value Float64",
			"quality Int32",
			"tags Map(String, String)",
		},
		Engine:      "MergeTree()",
		OrderBy:     []string{"(station_id, device_id, metric_name, timestamp)"},
		PartitionBy: "toYYYYMM(timestamp)",
		TTL:         "timestamp + INTERVAL 1 YEAR",
	}
}

func (s *ClickHouseStorage) Write(data *types.BatchData) error {
	if len(data.DataPoints) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.batchBuffer = append(s.batchBuffer, data.DataPoints...)

	if len(s.batchBuffer) >= s.batchSize {
		return s.flushLocked()
	}

	return nil
}

func (s *ClickHouseStorage) WritePoint(point *types.DataPoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.batchBuffer = append(s.batchBuffer, point)

	if len(s.batchBuffer) >= s.batchSize {
		return s.flushLocked()
	}

	return nil
}

func (s *ClickHouseStorage) flushLoop() {
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			if len(s.batchBuffer) > 0 {
				_ = s.flushLocked()
			}
			s.mu.Unlock()
		case <-s.stopChan:
			s.mu.Lock()
			if len(s.batchBuffer) > 0 {
				_ = s.flushLocked()
			}
			s.mu.Unlock()
			return
		}
	}
}

func (s *ClickHouseStorage) flushLocked() error {
	if len(s.batchBuffer) == 0 {
		return nil
	}

	fmt.Printf("Flushing %d data points to ClickHouse table %s\n", len(s.batchBuffer), s.config.Table)

	s.batchBuffer = s.batchBuffer[:0]
	return nil
}

func (s *ClickHouseStorage) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flushLocked()
}

func (s *ClickHouseStorage) Read(query string) ([]*types.DataPoint, error) {
	fmt.Printf("Reading data from ClickHouse with query: %s\n", query)

	// 查询优化：添加LIMIT子句限制返回数据量
	if !strings.Contains(strings.ToUpper(query), "LIMIT") {
		query += " LIMIT 1000"
	}

	// 执行查询（模拟）
	fmt.Printf("Executing optimized query: %s\n", query)

	return []*types.DataPoint{}, nil
}

func (s *ClickHouseStorage) ReadTimeRange(
	startTime, endTime time.Time,
	stationID, deviceID, metricName string,
) ([]*types.DataPoint, error) {
	// 优化查询：使用分区键和排序键，确保查询高效
	query := fmt.Sprintf(
		"SELECT timestamp, device_id, station_id, metric_name, metric_value, quality, tags "+
		"FROM %s "+
		"WHERE timestamp >= '%s' AND timestamp <= '%s'",
		s.config.Table,
		startTime.Format("2006-01-02 15:04:05"), // 使用更高效的时间格式
		endTime.Format("2006-01-02 15:04:05"),
	)

	// 构建WHERE条件，按照索引顺序添加
	conditions := []string{}
	if stationID != "" {
		conditions = append(conditions, fmt.Sprintf("station_id = '%s'", stationID))
	}
	if deviceID != "" {
		conditions = append(conditions, fmt.Sprintf("device_id = '%s'", deviceID))
	}
	if metricName != "" {
		conditions = append(conditions, fmt.Sprintf("metric_name = '%s'", metricName))
	}

	// 添加条件到查询
	for _, cond := range conditions {
		query += " AND " + cond
	}

	// 优化排序：使用索引顺序
	query += " ORDER BY station_id, device_id, metric_name, timestamp"

	// 添加LIMIT限制
	query += " LIMIT 10000"

	return s.Read(query)
}

func (s *ClickHouseStorage) Query(query string) (interface{}, error) {
	// 生成缓存键
	cacheKey := s.generateCacheKey(query)

	// 尝试从缓存获取
	if item, found := s.getFromCache(cacheKey); found {
		fmt.Printf("Cache hit for query: %s\n", query)
		return item.Value, nil
	}

	// 缓存未命中，执行查询
	fmt.Printf("Executing query on ClickHouse: %s\n", query)

	result := []map[string]interface{}{
		{
			"query":      query,
			"executed_at": time.Now().Format(time.RFC3339),
			"rows":       0,
		},
	}

	// 将结果存入缓存
	s.setToCache(cacheKey, result)
	fmt.Printf("Cache miss, stored result in cache: %s\n", cacheKey)

	// 定期清理过期缓存
	go func() {
		s.clearExpiredCache()
	}()

	return result, nil
}

func (s *ClickHouseStorage) Aggregate(
	aggregation string,
	metricName string,
	startTime, endTime time.Time,
	groupBy string,
) (interface{}, error) {
	// 优化聚合查询：使用正确的聚合函数和索引
	query := fmt.Sprintf(
		"SELECT %s(metric_value) as value, %s FROM %s "+
		"WHERE metric_name = '%s' AND timestamp >= '%s' AND timestamp <= '%s' "+
		"GROUP BY %s "+
		"ORDER BY %s",
		aggregation,
		groupBy,
		s.config.Table,
		metricName,
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		groupBy,
		groupBy,
	)

	// 执行优化后的查询
	fmt.Printf("Executing optimized aggregate query: %s\n", query)

	return s.Query(query)
}

func (s *ClickHouseStorage) GetStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"storage_type":   "clickhouse",
		"table":          s.config.Table,
		"batch_size":     s.batchSize,
		"flush_interval": s.flushInterval.String(),
		"buffer_size":    len(s.batchBuffer),
		"started":        s.started,
	}
	return stats, nil
}

func (s *ClickHouseStorage) CreateMaterializedView(name string, targetTable string, query string) error {
	if !s.started {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 构建物化视图创建语句
	createQuery := fmt.Sprintf(
		"CREATE MATERIALIZED VIEW IF NOT EXISTS %s TO %s AS %s",
		name, targetTable, query,
	)

	fmt.Printf("Creating materialized view: %s\n", name)
	fmt.Printf("Query: %s\n", createQuery)

	// 执行查询（模拟）
	_, err := s.Query(createQuery)
	if err != nil {
		return err
	}

	// 添加到物化视图列表
	s.materializedViews = append(s.materializedViews, name)
	return nil
}

func (s *ClickHouseStorage) ListMaterializedViews() ([]string, error) {
	if !s.started {
		return nil, &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 模拟查询物化视图列表
	fmt.Println("Listing materialized views")
	return s.materializedViews, nil
}

func (s *ClickHouseStorage) DropMaterializedView(name string) error {
	if !s.started {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 构建删除语句
	dropQuery := fmt.Sprintf("DROP VIEW IF EXISTS %s", name)
	fmt.Printf("Dropping materialized view: %s\n", name)

	// 执行查询（模拟）
	_, err := s.Query(dropQuery)
	if err != nil {
		return err
	}

	// 从列表中移除
	for i, view := range s.materializedViews {
		if view == name {
			s.materializedViews = append(s.materializedViews[:i], s.materializedViews[i+1:]...)
			break
		}
	}

	return nil
}

func (s *ClickHouseStorage) RefreshMaterializedView(name string) error {
	if !s.started {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 构建刷新语句（ClickHouse会自动刷新物化视图，这里只是模拟）
	refreshQuery := fmt.Sprintf("SYSTEM REFRESH MATERIALIZED VIEW %s", name)
	fmt.Printf("Refreshing materialized view: %s\n", name)

	// 执行查询（模拟）
	_, err := s.Query(refreshQuery)
	return err
}

func (s *ClickHouseStorage) ExplainQuery(query string) (interface{}, error) {
	if !s.started {
		return nil, &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 构建EXPLAIN查询
	explainQuery := fmt.Sprintf("EXPLAIN SYNTAX %s", query)
	fmt.Printf("Explaining query: %s\n", explainQuery)

	// 执行EXPLAIN查询（模拟）
	result := map[string]interface{}{
		"original_query": query,
		"explain_query":  explainQuery,
		"execution_plan": []string{
			"1. Read data from table",
			"2. Apply WHERE conditions",
			"3. Group by specified columns",
			"4. Apply aggregation functions",
			"5. Sort results",
			"6. Limit output",
		},
		"optimization_suggestions": []string{
			"Use partition pruning by timestamp",
			"Ensure WHERE conditions match index order",
			"Use appropriate data types for columns",
			"Consider using materialized views for frequent queries",
		},
		"estimated_rows": 1000,
		"estimated_time": "~100ms",
	}

	return result, nil
}

func (s *ClickHouseStorage) CreatePreAggregationTable(tableName string, timeInterval string) error {
	if !s.started {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 构建预聚合表创建语句
	createQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			timestamp DateTime,
			station_id String,
			device_id String,
			metric_name String,
			metric_value_sum Float64,
			metric_value_avg Float64,
			metric_value_min Float64,
			metric_value_max Float64,
			metric_value_count UInt64
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(timestamp)
		ORDER BY (station_id, device_id, metric_name, timestamp)
		TTL timestamp + INTERVAL 1 YEAR
	`, tableName)

	fmt.Printf("Creating pre-aggregation table: %s\n", tableName)
	fmt.Printf("Time interval: %s\n", timeInterval)

	// 执行查询（模拟）
	_, err := s.Query(createQuery)
	if err != nil {
		return err
	}

	// 添加到预聚合表列表
	s.preAggregationTables = append(s.preAggregationTables, tableName)
	return nil
}

func (s *ClickHouseStorage) CreatePreAggregationRule(rule PreAggregationRule) error {
	if !s.started {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 生成规则ID
	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule_%d", time.Now().UnixNano())
	}
	rule.CreatedAt = time.Now()

	// 构建预聚合查询
	groupByClause := strings.Join(rule.GroupBy, ", ")
	query := fmt.Sprintf(`
		SELECT
			toStartOf%s(timestamp) as timestamp,
			%s,
			sum(metric_value) as metric_value_sum,
			avg(metric_value) as metric_value_avg,
			min(metric_value) as metric_value_min,
			max(metric_value) as metric_value_max,
			count() as metric_value_count
		FROM %s
		GROUP BY timestamp, %s
	`, rule.TimeInterval, groupByClause, rule.SourceTable, groupByClause)

	// 创建物化视图用于预聚合
	err := s.CreateMaterializedView(
		fmt.Sprintf("mv_%s", rule.ID),
		rule.TargetTable,
		query,
	)
	if err != nil {
		return err
	}

	// 添加到预聚合规则列表
	s.preAggregationRules = append(s.preAggregationRules, rule)
	fmt.Printf("Created pre-aggregation rule: %s\n", rule.ID)

	return nil
}

func (s *ClickHouseStorage) ListPreAggregationRules() ([]PreAggregationRule, error) {
	if !s.started {
		return nil, &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	fmt.Println("Listing pre-aggregation rules")
	return s.preAggregationRules, nil
}

func (s *ClickHouseStorage) EnablePreAggregationRule(ruleID string) error {
	if !s.started {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 查找并启用规则
	for i, rule := range s.preAggregationRules {
		if rule.ID == ruleID {
			s.preAggregationRules[i].Enabled = true
			fmt.Printf("Enabled pre-aggregation rule: %s\n", ruleID)
			return nil
		}
	}

	return &types.Error{
		Code:    types.ErrCodeStorageError,
		Message: fmt.Sprintf("pre-aggregation rule not found: %s", ruleID),
	}
}

func (s *ClickHouseStorage) DisablePreAggregationRule(ruleID string) error {
	if !s.started {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 查找并禁用规则
	for i, rule := range s.preAggregationRules {
		if rule.ID == ruleID {
			s.preAggregationRules[i].Enabled = false
			fmt.Printf("Disabled pre-aggregation rule: %s\n", ruleID)
			return nil
		}
	}

	return &types.Error{
		Code:    types.ErrCodeStorageError,
		Message: fmt.Sprintf("pre-aggregation rule not found: %s", ruleID),
	}
}

func (s *ClickHouseStorage) DeletePreAggregationRule(ruleID string) error {
	if !s.started {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 查找规则
	var ruleToDelete PreAggregationRule
	var ruleIndex int
	found := false

	for i, rule := range s.preAggregationRules {
		if rule.ID == ruleID {
			ruleToDelete = rule
			ruleIndex = i
			found = true
			break
		}
	}

	if !found {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: fmt.Sprintf("pre-aggregation rule not found: %s", ruleID),
		}
	}

	// 删除对应的物化视图
	err := s.DropMaterializedView(fmt.Sprintf("mv_%s", ruleID))
	if err != nil {
		return err
	}

	// 从规则列表中移除
	s.preAggregationRules = append(s.preAggregationRules[:ruleIndex], s.preAggregationRules[ruleIndex+1:]...)
	fmt.Printf("Deleted pre-aggregation rule: %s\n", ruleID)

	return nil
}

func (s *ClickHouseStorage) RefreshPreAggregation(tableName string) error {
	if !s.started {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 构建刷新查询（模拟）
	refreshQuery := fmt.Sprintf("SYSTEM REFRESH MATERIALIZED VIEW mv_%s", tableName)
	fmt.Printf("Refreshing pre-aggregation for table: %s\n", tableName)

	// 执行查询（模拟）
	_, err := s.Query(refreshQuery)
	return err
}

func (s *ClickHouseStorage) generateCacheKey(query string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(query)))
}

func (s *ClickHouseStorage) getFromCache(key string) (*CacheItem, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.queryCache[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		delete(s.queryCache, key)
		return nil, false
	}

	// 更新访问计数
	item.AccessCount++
	s.queryCache[key] = item

	return item, true
}

func (s *ClickHouseStorage) setToCache(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查缓存大小
	if len(s.queryCache) >= s.cacheMaxSize {
		s.evictOldestCacheItems()
	}

	// 创建缓存项
	now := time.Now()
	item := &CacheItem{
		Key:        key,
		Value:      value,
		CreatedAt:  now,
		ExpiresAt:  now.Add(s.cacheTTL),
		AccessCount: 1,
	}

	s.queryCache[key] = item
}

func (s *ClickHouseStorage) evictOldestCacheItems() {
	// 按访问次数和创建时间排序，删除最旧的10%缓存项
	evictCount := len(s.queryCache) / 10
	if evictCount < 1 {
		evictCount = 1
	}

	// 收集所有缓存项
	items := make([]*CacheItem, 0, len(s.queryCache))
	for _, item := range s.queryCache {
		items = append(items, item)
	}

	// 简单的淘汰策略：删除最早创建的项
	for i := 0; i < evictCount && len(s.queryCache) > 0; i++ {
		oldestKey := ""
		oldestTime := time.Now()

		for key, item := range s.queryCache {
			if item.CreatedAt.Before(oldestTime) {
				oldestTime = item.CreatedAt
				oldestKey = key
			}
		}

		if oldestKey != "" {
			delete(s.queryCache, oldestKey)
		}
	}
}

func (s *ClickHouseStorage) clearExpiredCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, item := range s.queryCache {
		if now.After(item.ExpiresAt) {
			delete(s.queryCache, key)
		}
	}
}

func (s *ClickHouseStorage) GetCacheStats() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	totalItems := len(s.queryCache)
	totalAccesses := 0
	expiringSoon := 0
	now := time.Now()

	for _, item := range s.queryCache {
		totalAccesses += item.AccessCount
		if now.Add(time.Minute).After(item.ExpiresAt) {
			expiringSoon++
		}
	}

	return map[string]interface{}{
		"total_items":    totalItems,
		"total_accesses": totalAccesses,
		"expiring_soon":  expiringSoon,
		"max_size":       s.cacheMaxSize,
		"ttl_seconds":    s.cacheTTL.Seconds(),
	}
}

func (s *ClickHouseStorage) ClearCache() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.queryCache = make(map[string]*CacheItem)
	fmt.Println("Cache cleared")
	return nil
}

func (s *ClickHouseStorage) MultiDimensionAggregation(
	metrics []string,
	dimensions []string,
	startTime, endTime time.Time,
	filters map[string]interface{},
) (interface{}, error) {
	if !s.started {
		return nil, &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 构建查询
	metricClauses := []string{}
	for _, metric := range metrics {
		metricClauses = append(metricClauses, fmt.Sprintf("sum(%s) as sum_%s, avg(%s) as avg_%s, min(%s) as min_%s, max(%s) as max_%s",
			metric, metric, metric, metric, metric, metric, metric, metric))
	}

	dimensionClauses := strings.Join(dimensions, ", ")
	metricClause := strings.Join(metricClauses, ", ")

	query := fmt.Sprintf(`
		SELECT
			%s,
			%s
		FROM %s
		WHERE timestamp >= '%s' AND timestamp <= '%s'
	`, dimensionClauses, metricClause, s.config.Table, startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))

	// 添加过滤条件
	for key, value := range filters {
		query += fmt.Sprintf(" AND %s = '%v'", key, value)
	}

	// 添加分组
	query += fmt.Sprintf(" GROUP BY %s", dimensionClauses)

	// 执行查询
	fmt.Printf("Executing multi-dimension aggregation query: %s\n", query)
	result, err := s.Query(query)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ClickHouseStorage) DimensionDrillDown(
	baseDimensions []string,
	drillDownDimension string,
	metrics []string,
	startTime, endTime time.Time,
	filters map[string]interface{},
) (interface{}, error) {
	if !s.started {
		return nil, &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 合并维度
	allDimensions := append(baseDimensions, drillDownDimension)

	// 构建查询
	metricClauses := []string{}
	for _, metric := range metrics {
		metricClauses = append(metricClauses, fmt.Sprintf("sum(%s) as sum_%s", metric, metric))
	}

	dimensionClauses := strings.Join(allDimensions, ", ")
	metricClause := strings.Join(metricClauses, ", ")

	query := fmt.Sprintf(`
		SELECT
			%s,
			%s
		FROM %s
		WHERE timestamp >= '%s' AND timestamp <= '%s'
	`, dimensionClauses, metricClause, s.config.Table, startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))

	// 添加过滤条件
	for key, value := range filters {
		query += fmt.Sprintf(" AND %s = '%v'", key, value)
	}

	// 添加分组
	query += fmt.Sprintf(" GROUP BY %s", dimensionClauses)

	// 执行查询
	fmt.Printf("Executing dimension drill-down query: %s\n", query)
	result, err := s.Query(query)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ClickHouseStorage) DimensionCrossAnalysis(
	dimensions1 []string,
	dimensions2 []string,
	metric string,
	startTime, endTime time.Time,
	filters map[string]interface{},
) (interface{}, error) {
	if !s.started {
		return nil, &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 合并维度
	allDimensions := append(dimensions1, dimensions2...)

	// 构建查询
	dimensionClauses := strings.Join(allDimensions, ", ")

	query := fmt.Sprintf(`
		SELECT
			%s,
			sum(%s) as sum_%s,
			avg(%s) as avg_%s
		FROM %s
		WHERE timestamp >= '%s' AND timestamp <= '%s'
	`, dimensionClauses, metric, metric, metric, metric, s.config.Table, startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))

	// 添加过滤条件
	for key, value := range filters {
		query += fmt.Sprintf(" AND %s = '%v'", key, value)
	}

	// 添加分组
	query += fmt.Sprintf(" GROUP BY %s", dimensionClauses)

	// 执行查询
	fmt.Printf("Executing dimension cross-analysis query: %s\n", query)
	result, err := s.Query(query)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ClickHouseStorage) GetDimensionValues(
	dimension string,
	startTime, endTime time.Time,
	filters map[string]interface{},
) (interface{}, error) {
	if !s.started {
		return nil, &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	// 构建查询
	query := fmt.Sprintf(`
		SELECT DISTINCT %s
		FROM %s
		WHERE timestamp >= '%s' AND timestamp <= '%s'
	`, dimension, s.config.Table, startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))

	// 添加过滤条件
	for key, value := range filters {
		query += fmt.Sprintf(" AND %s = '%v'", key, value)
	}

	// 执行查询
	fmt.Printf("Executing dimension values query: %s\n", query)
	result, err := s.Query(query)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ClickHouseStorage) Close() error {
	if !s.started {
		return nil
	}

	close(s.stopChan)

	s.mu.Lock()
	if len(s.batchBuffer) > 0 {
		_ = s.flushLocked()
	}
	s.mu.Unlock()

	fmt.Println("Closing ClickHouse connection")
	s.started = false

	return nil
}