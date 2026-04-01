package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/new-energy-monitoring/pkg/storage/query"
)

// BenchmarkQueryMillionRecords 500万记录查询响应测试
func BenchmarkQueryMillionRecords(b *testing.B) {
	recordCounts := []int{100000, 500000, 1000000, 2000000, 5000000}

	for _, records := range recordCounts {
		b.Run(fmt.Sprintf("Records_%d", records), func(b *testing.B) {
			// 创建模拟数据库
			mockDB := NewMockDatabase(records)

			executor := query.NewQueryExecutor(mockDB.GetDB(), query.ExecutorConfig{
				MaxParallelQueries:   10,
				MaxResultRows:        100000,
				DefaultTimeout:       30 * time.Second,
				SlowQueryThreshold:   1 * time.Second,
				EnableQueryPlan:      true,
				EnableParallel:       true,
				StreamBatchSize:      1000,
			})

			ctx := context.Background()

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				req := &query.QueryRequest{
					Type:      query.QueryTypeSelect,
					Table:     "point_data",
					Fields:    []string{"point_id", "value", "quality", "timestamp"},
					Limit:     1000,
					Offset:    (i % 100) * 1000,
					TimeRange: &query.TimeRange{
						Field: "timestamp",
						Start: time.Now().Add(-24 * time.Hour),
						End:   time.Now(),
					},
				}

				result, err := executor.Execute(ctx, req)
				if err != nil {
					b.Errorf("Query failed: %v", err)
					continue
				}

				if len(result.Data) == 0 {
					b.Errorf("Expected data, got empty result")
				}
			}

			stats := executor.GetStats()
			b.ReportMetric(float64(stats.AvgExecutionTime)/1e6, "avg_query_time_ms")
			b.ReportMetric(float64(stats.SuccessQueries), "success_queries")
		})
	}
}

// BenchmarkQueryComplexQuery 复杂查询性能测试
func BenchmarkQueryComplexQuery(b *testing.B) {
	mockDB := NewMockDatabase(1000000)

	executor := query.NewQueryExecutor(mockDB.GetDB(), query.ExecutorConfig{
		MaxParallelQueries:   10,
		MaxResultRows:        100000,
		DefaultTimeout:       30 * time.Second,
		SlowQueryThreshold:   1 * time.Second,
		EnableQueryPlan:      true,
		EnableParallel:       true,
		StreamBatchSize:      1000,
	})

	ctx := context.Background()

	testCases := []struct {
		name  string
		query *query.QueryRequest
	}{
		{
			name: "SimpleSelect",
			query: &query.QueryRequest{
				Type:  query.QueryTypeSelect,
				Table: "point_data",
				Fields: []string{"point_id", "value", "quality", "timestamp"},
				Conditions: []query.QueryCondition{
					{Field: "quality", Operator: "=", Value: 192},
				},
				Limit: 1000,
			},
		},
		{
			name: "TimeRangeQuery",
			query: &query.QueryRequest{
				Type:  query.QueryTypeTimeRange,
				Table: "point_data",
				Fields: []string{"point_id", "value", "timestamp"},
				TimeRange: &query.TimeRange{
					Field: "timestamp",
					Start: time.Now().Add(-1 * time.Hour),
					End:   time.Now(),
				},
				Limit: 1000,
			},
		},
		{
			name: "AggregateQuery",
			query: &query.QueryRequest{
				Type:  query.QueryTypeAggregate,
				Table: "point_data",
				Aggregates: []query.AggregateField{
					{Field: "value", Function: "AVG", Alias: "avg_value"},
					{Field: "value", Function: "MAX", Alias: "max_value"},
					{Field: "value", Function: "MIN", Alias: "min_value"},
					{Field: "point_id", Function: "COUNT", Alias: "count"},
				},
				GroupBy: []string{"point_id"},
				TimeRange: &query.TimeRange{
					Field: "timestamp",
					Start: time.Now().Add(-24 * time.Hour),
					End:   time.Now(),
				},
			},
		},
		{
			name: "JoinQuery",
			query: &query.QueryRequest{
				Type:  query.QueryTypeJoin,
				Table: "point_data",
				Fields: []string{"point_data.point_id", "point_data.value", "devices.name"},
				Joins: []query.JoinClause{
					{
						Type:  "INNER",
						Table: "devices",
						Alias: "devices",
						Conditions: []query.JoinCondition{
							{LeftField: "point_data.device_id", Operator: "=", RightField: "devices.id"},
						},
					},
				},
				Conditions: []query.QueryCondition{
					{Field: "point_data.quality", Operator: "=", Value: 192},
				},
				Limit: 1000,
			},
		},
		{
			name: "ComplexQuery",
			query: &query.QueryRequest{
				Type:  query.QueryTypeComplex,
				Table: "point_data",
				Fields: []string{"point_id", "value", "quality", "timestamp"},
				Conditions: []query.QueryCondition{
					{Field: "quality", Operator: ">=", Value: 128},
					{Field: "value", Operator: ">", Value: 0},
				},
				Aggregates: []query.AggregateField{
					{Field: "value", Function: "AVG", Alias: "avg_value"},
					{Field: "value", Function: "STDDEV", Alias: "stddev_value"},
				},
				GroupBy: []string{"point_id"},
				OrderBy: []query.OrderByField{
					{Field: "timestamp", Desc: true},
				},
				TimeRange: &query.TimeRange{
					Field: "timestamp",
					Start: time.Now().Add(-7 * 24 * time.Hour),
					End:   time.Now(),
				},
				Limit: 500,
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result, err := executor.Execute(ctx, tc.query)
				if err != nil {
					b.Errorf("Query %s failed: %v", tc.name, err)
					continue
				}

				if result.ExecutionTime > 1*time.Second {
					b.Logf("Slow query %s: %v", tc.name, result.ExecutionTime)
				}
			}
		})
	}
}

// BenchmarkQueryConcurrent 并发查询测试
func BenchmarkQueryConcurrent(b *testing.B) {
	concurrencyLevels := []int{10, 50, 100, 200, 500}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			mockDB := NewMockDatabase(1000000)

			executor := query.NewQueryExecutor(mockDB.GetDB(), query.ExecutorConfig{
				MaxParallelQueries:   concurrency,
				MaxResultRows:        100000,
				DefaultTimeout:       30 * time.Second,
				SlowQueryThreshold:   1 * time.Second,
				EnableQueryPlan:      true,
				EnableParallel:       true,
				StreamBatchSize:      1000,
			})

			ctx := context.Background()

			b.ResetTimer()

			var wg sync.WaitGroup
			wg.Add(concurrency)

			var successCount int64
			var failCount int64

			for i := 0; i < concurrency; i++ {
				go func(idx int) {
					defer wg.Done()

					for j := 0; j < b.N; j++ {
						req := &query.QueryRequest{
							Type:  query.QueryTypeSelect,
							Table: "point_data",
							Fields: []string{"point_id", "value", "quality", "timestamp"},
							Conditions: []query.QueryCondition{
								{Field: "point_id", Operator: "=", Value: fmt.Sprintf("point-%d", idx)},
							},
							Limit: 100,
						}

						_, err := executor.Execute(ctx, req)
						if err != nil {
							atomic.AddInt64(&failCount, 1)
						} else {
							atomic.AddInt64(&successCount, 1)
						}
					}
				}(i)
			}

			wg.Wait()

			b.ReportMetric(float64(successCount), "success_count")
			b.ReportMetric(float64(failCount), "fail_count")
		})
	}
}

// BenchmarkQueryCacheHitRate 缓存命中率测试
func BenchmarkQueryCacheHitRate(b *testing.B) {
	mockDB := NewMockDatabase(1000000)

	cacheConfig := query.DefaultCacheConfig()
	cacheConfig.MaxSize = 100 * 1024 * 1024 // 100MB
	cacheConfig.MaxEntries = 10000
	cacheConfig.DefaultTTL = 5 * time.Minute

	// 创建模拟Redis客户端
	mockRedis := NewMockRedisClient()
	queryCache := query.NewQueryCache(mockRedis, cacheConfig)

	executor := query.NewQueryExecutor(mockDB.GetDB(), query.ExecutorConfig{
		MaxParallelQueries:   10,
		MaxResultRows:        100000,
		DefaultTimeout:       30 * time.Second,
		SlowQueryThreshold:   1 * time.Second,
		EnableQueryPlan:      true,
		EnableParallel:       true,
		StreamBatchSize:      1000,
	})

	ctx := context.Background()

	// 预热查询
	warmupQueries := []*query.QueryRequest{
		{
			Type:   query.QueryTypeSelect,
			Table:  "point_data",
			Fields: []string{"point_id", "value", "quality", "timestamp"},
			Limit:  100,
		},
		{
			Type:   query.QueryTypeAggregate,
			Table:  "point_data",
			Fields: []string{"point_id"},
			Aggregates: []query.AggregateField{
				{Field: "value", Function: "AVG", Alias: "avg_value"},
			},
			GroupBy: []string{"point_id"},
		},
	}

	// 预热缓存
	for _, req := range warmupQueries {
		result, _ := executor.Execute(ctx, req)
		queryCache.Set(ctx, req, result)
	}

	b.ResetTimer()

	// 混合查询：80%重复查询（命中缓存），20%新查询（未命中）
	for i := 0; i < b.N; i++ {
		var req *query.QueryRequest

		if i%5 == 0 {
			// 新查询
			req = &query.QueryRequest{
				Type:   query.QueryTypeSelect,
				Table:  "point_data",
				Fields: []string{"point_id", "value", "quality", "timestamp"},
				Conditions: []query.QueryCondition{
					{Field: "point_id", Operator: "=", Value: fmt.Sprintf("point-%d", i)},
				},
				Limit: 100,
			}
		} else {
			// 重复查询
			req = warmupQueries[i%len(warmupQueries)]
		}

		// 先检查缓存
		result, status, err := queryCache.Get(ctx, req)
		if err != nil {
			b.Errorf("Cache get failed: %v", err)
			continue
		}

		if status == query.CacheStatusMiss {
			// 缓存未命中，执行查询
			result, err = executor.Execute(ctx, req)
			if err != nil {
				b.Errorf("Query failed: %v", err)
				continue
			}
			queryCache.Set(ctx, req, result)
		}
	}

	stats := queryCache.GetStats()
	b.ReportMetric(stats.HitRate*100, "hit_rate_percent")
	b.ReportMetric(float64(stats.Hits), "cache_hits")
	b.ReportMetric(float64(stats.Misses), "cache_misses")
}

// BenchmarkQueryStream 流式查询测试
func BenchmarkQueryStream(b *testing.B) {
	mockDB := NewMockDatabase(1000000)

	executor := query.NewQueryExecutor(mockDB.GetDB(), query.ExecutorConfig{
		MaxParallelQueries:   10,
		MaxResultRows:        100000,
		DefaultTimeout:       30 * time.Second,
		SlowQueryThreshold:   1 * time.Second,
		EnableQueryPlan:      true,
		EnableParallel:       true,
		StreamBatchSize:      1000,
	})

	ctx := context.Background()

	req := &query.QueryRequest{
		Type:  query.QueryTypeSelect,
		Table: "point_data",
		Fields: []string{"point_id", "value", "quality", "timestamp"},
		TimeRange: &query.TimeRange{
			Field: "timestamp",
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now(),
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stream, err := executor.StreamExecute(ctx, req)
		if err != nil {
			b.Errorf("Stream execute failed: %v", err)
			continue
		}

		var rowCount int
		for row := range stream {
			if row.Error != nil {
				b.Errorf("Stream row error: %v", row.Error)
				break
			}
			rowCount++
		}

		if rowCount == 0 {
			b.Errorf("Expected rows, got 0")
		}
	}
}

// BenchmarkQueryPlanOptimization 查询计划优化测试
func BenchmarkQueryPlanOptimization(b *testing.B) {
	optimizer := query.NewQueryPlanOptimizer()

	testQueries := []*query.QueryRequest{
		{
			Type:  query.QueryTypeSelect,
			Table: "point_data",
			Fields: []string{"point_id", "value", "quality"},
			Conditions: []query.QueryCondition{
				{Field: "quality", Operator: "=", Value: 192},
				{Field: "value", Operator: ">", Value: 100},
			},
			Limit: 1000,
		},
		{
			Type:  query.QueryTypeJoin,
			Table: "point_data",
			Joins: []query.JoinClause{
				{Type: "INNER", Table: "devices", Alias: "d"},
				{Type: "INNER", Table: "stations", Alias: "s"},
			},
			Conditions: []query.QueryCondition{
				{Field: "point_data.quality", Operator: "=", Value: 192},
			},
		},
		{
			Type:  query.QueryTypeAggregate,
			Table: "point_data",
			Aggregates: []query.AggregateField{
				{Field: "value", Function: "AVG", Alias: "avg"},
				{Field: "value", Function: "MAX", Alias: "max"},
				{Field: "value", Function: "MIN", Alias: "min"},
			},
			GroupBy: []string{"point_id", "device_id"},
			Conditions: []query.QueryCondition{
				{Field: "timestamp", Operator: ">=", Value: time.Now().Add(-24 * time.Hour)},
			},
		},
	}

	for i, req := range testQueries {
		b.Run(fmt.Sprintf("Query_%d", i), func(b *testing.B) {
			b.ResetTimer()

			for j := 0; j < b.N; j++ {
				plan, err := optimizer.Optimize(req)
				if err != nil {
					b.Errorf("Optimization failed: %v", err)
					continue
				}

				if plan.EstimatedCost <= 0 {
					b.Errorf("Invalid estimated cost: %f", plan.EstimatedCost)
				}
			}

			b.ReportMetric(0, "ops")
		})
	}
}

// BenchmarkQueryMemoryAllocation 查询内存分配测试
func BenchmarkQueryMemoryAllocation(b *testing.B) {
	mockDB := NewMockDatabase(1000000)

	executor := query.NewQueryExecutor(mockDB.GetDB(), query.ExecutorConfig{
		MaxParallelQueries:   10,
		MaxResultRows:        100000,
		DefaultTimeout:       30 * time.Second,
		SlowQueryThreshold:   1 * time.Second,
		EnableQueryPlan:      true,
		EnableParallel:       true,
		StreamBatchSize:      1000,
	})

	ctx := context.Background()

	req := &query.QueryRequest{
		Type:  query.QueryTypeSelect,
		Table: "point_data",
		Fields: []string{"point_id", "value", "quality", "timestamp"},
		Limit:  10000,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(ctx, req)
		if err != nil {
			b.Errorf("Query failed: %v", err)
		}
	}
}

// BenchmarkQueryCPUProfile CPU性能分析
func BenchmarkQueryCPUProfile(b *testing.B) {
	cpuProfile, err := os.Create("cpu_query.prof")
	if err != nil {
		b.Fatalf("Failed to create CPU profile: %v", err)
	}
	defer cpuProfile.Close()

	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		b.Fatalf("Failed to start CPU profile: %v", err)
	}
	defer pprof.StopCPUProfile()

	mockDB := NewMockDatabase(1000000)

	executor := query.NewQueryExecutor(mockDB.GetDB(), query.ExecutorConfig{
		MaxParallelQueries:   10,
		MaxResultRows:        100000,
		DefaultTimeout:       30 * time.Second,
		SlowQueryThreshold:   1 * time.Second,
		EnableQueryPlan:      true,
		EnableParallel:       true,
		StreamBatchSize:      1000,
	})

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := &query.QueryRequest{
			Type:  query.QueryTypeSelect,
			Table: "point_data",
			Fields: []string{"point_id", "value", "quality", "timestamp"},
			Conditions: []query.QueryCondition{
				{Field: "quality", Operator: "=", Value: 192},
			},
			Limit: 1000,
		}

		_, err := executor.Execute(ctx, req)
		if err != nil {
			b.Errorf("Query failed: %v", err)
		}
	}
}

// MockDatabase 模拟数据库
type MockDatabase struct {
	records int
	data    []map[string]interface{}
	mu      sync.RWMutex
}

// NewMockDatabase 创建模拟数据库
func NewMockDatabase(records int) *MockDatabase {
	db := &MockDatabase{
		records: records,
		data:    make([]map[string]interface{}, records),
	}

	// 生成模拟数据
	for i := 0; i < records; i++ {
		db.data[i] = map[string]interface{}{
			"id":        i,
			"point_id":  fmt.Sprintf("point-%d", i%10000),
			"device_id": fmt.Sprintf("device-%d", i%1000),
			"value":     float64(i % 1000),
			"quality":   192,
			"timestamp": time.Now().Add(-time.Duration(i) * time.Second),
		}
	}

	return db
}

// GetDB 获取数据库连接（模拟）
func (m *MockDatabase) GetDB() interface{} {
	return m
}

// Query 模拟查询
func (m *MockDatabase) Query(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 模拟查询延迟
	time.Sleep(time.Microsecond * time.Duration(m.records/10000))

	// 返回部分数据
	limit := 1000
	if len(m.data) < limit {
		return m.data, nil
	}
	return m.data[:limit], nil
}

// MockRedisClient 模拟Redis客户端
type MockRedisClient struct {
	data map[string][]byte
	mu   sync.RWMutex
}

// NewMockRedisClient 创建模拟Redis客户端
func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string][]byte),
	}
}

// Get 获取数据
func (m *MockRedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found")
	}
	return data, nil
}

// Set 设置数据
func (m *MockRedisClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
	return nil
}

// Del 删除数据
func (m *MockRedisClient) Del(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

// Exists 检查是否存在
func (m *MockRedisClient) Exists(ctx context.Context, key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.data[key]
	return exists
}

// Scan 扫描键
func (m *MockRedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}
