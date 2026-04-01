package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkDatabaseBatchInsert 批量插入性能测试
func BenchmarkDatabaseBatchInsert(b *testing.B) {
	batchSizes := []int{100, 500, 1000, 5000, 10000}
	
	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize_%d", batchSize), func(b *testing.B) {
			db := NewMockDatabase(0)
			
			b.ResetTimer()
			
			var totalInserted int64
			var totalDuration time.Duration
			
			for i := 0; i < b.N; i++ {
				records := generatePointDataRecords(batchSize)
				
				start := time.Now()
				err := db.BatchInsert("point_data", records)
				duration := time.Since(start)
				
				if err != nil {
					b.Errorf("Batch insert failed: %v", err)
					continue
				}
				
				atomic.AddInt64(&totalInserted, int64(batchSize))
				totalDuration += duration
			}
			
			b.ReportMetric(float64(totalInserted)/totalDuration.Seconds(), "records_per_second")
			b.ReportMetric(float64(totalDuration.Milliseconds())/float64(b.N), "avg_batch_time_ms")
		})
	}
}

// BenchmarkDatabaseComplexQuery 复杂查询性能测试
func BenchmarkDatabaseComplexQuery(b *testing.B) {
	db := NewMockDatabase(1000000) // 100万条记录
	
	queryTypes := []struct {
		name  string
		query MockQuery
	}{
		{
			name: "SimpleSelect",
			query: MockQuery{
				Table: "point_data",
				Conditions: []MockCondition{
					{Field: "quality", Operator: "=", Value: 192},
				},
				Limit: 1000,
			},
		},
		{
			name: "RangeQuery",
			query: MockQuery{
				Table: "point_data",
				Conditions: []MockCondition{
					{Field: "value", Operator: ">", Value: 500},
					{Field: "value", Operator: "<", Value: 1000},
				},
				Limit: 1000,
			},
		},
		{
			name: "TimeRangeQuery",
			query: MockQuery{
				Table: "point_data",
				TimeRange: &MockTimeRange{
					Field: "timestamp",
					Start: time.Now().Add(-24 * time.Hour),
					End:   time.Now(),
				},
				Limit: 1000,
			},
		},
		{
			name: "MultiConditionQuery",
			query: MockQuery{
				Table: "point_data",
				Conditions: []MockCondition{
					{Field: "quality", Operator: ">=", Value: 128},
					{Field: "station_id", Operator: "=", Value: "station-001"},
					{Field: "device_id", Operator: "=", Value: "device-001"},
				},
				TimeRange: &MockTimeRange{
					Field: "timestamp",
					Start: time.Now().Add(-1 * time.Hour),
					End:   time.Now(),
				},
				Limit: 1000,
			},
		},
		{
			name: "AggregateQuery",
			query: MockQuery{
				Table: "point_data",
				Aggregates: []MockAggregate{
					{Field: "value", Function: "AVG", Alias: "avg_value"},
					{Field: "value", Function: "MAX", Alias: "max_value"},
					{Field: "value", Function: "MIN", Alias: "min_value"},
					{Field: "value", Function: "COUNT", Alias: "count"},
				},
				GroupBy: []string{"point_id"},
				Limit:   100,
			},
		},
		{
			name: "JoinQuery",
			query: MockQuery{
				Table: "point_data",
				Joins: []MockJoin{
					{Type: "INNER", Table: "devices", Alias: "d", OnField: "device_id"},
					{Type: "INNER", Table: "stations", Alias: "s", OnField: "station_id"},
				},
				Conditions: []MockCondition{
					{Field: "quality", Operator: "=", Value: 192},
				},
				Limit: 1000,
			},
		},
	}
	
	for _, qt := range queryTypes {
		b.Run(qt.name, func(b *testing.B) {
			ctx := context.Background()
			
			b.ResetTimer()
			
			var totalDuration time.Duration
			
			for i := 0; i < b.N; i++ {
				start := time.Now()
				_, err := db.ExecuteQuery(ctx, &qt.query)
				duration := time.Since(start)
				
				if err != nil {
					b.Errorf("Query failed: %v", err)
					continue
				}
				
				totalDuration += duration
			}
			
			avgDuration := totalDuration / time.Duration(b.N)
			b.ReportMetric(float64(avgDuration.Milliseconds()), "avg_query_time_ms")
		})
	}
}

// BenchmarkDatabaseIndexEfficiency 索引效率测试
func BenchmarkDatabaseIndexEfficiency(b *testing.B) {
	db := NewMockDatabase(1000000)
	
	// 无索引查询
	b.Run("WithoutIndex", func(b *testing.B) {
		query := &MockQuery{
			Table: "point_data",
			Conditions: []MockCondition{
				{Field: "non_indexed_field", Operator: "=", Value: "value-12345"},
			},
			Limit: 100,
		}
		
		ctx := context.Background()
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			_, err := db.ExecuteQuery(ctx, query)
			if err != nil {
				b.Errorf("Query failed: %v", err)
			}
		}
	})
	
	// 有索引查询
	b.Run("WithIndex", func(b *testing.B) {
		db.CreateIndex("point_data", "point_id")
		
		query := &MockQuery{
			Table: "point_data",
			Conditions: []MockCondition{
				{Field: "point_id", Operator: "=", Value: "point-12345"},
			},
			Limit: 100,
		}
		
		ctx := context.Background()
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			_, err := db.ExecuteQuery(ctx, query)
			if err != nil {
				b.Errorf("Query failed: %v", err)
			}
		}
	})
}

// BenchmarkDatabaseConnectionPool 连接池性能测试
func BenchmarkDatabaseConnectionPool(b *testing.B) {
	poolConfigs := []struct {
		name        string
		maxOpen     int
		maxIdle     int
		maxLifetime time.Duration
	}{
		{"SmallPool", 10, 5, 5 * time.Minute},
		{"MediumPool", 50, 25, 10 * time.Minute},
		{"LargePool", 100, 50, 15 * time.Minute},
		{"XLargePool", 200, 100, 30 * time.Minute},
	}
	
	for _, config := range poolConfigs {
		b.Run(config.name, func(b *testing.B) {
			pool := NewMockConnectionPool(config.maxOpen, config.maxIdle, config.maxLifetime)
			defer pool.Close()
			
			concurrency := 100
			
			b.ResetTimer()
			
			var wg sync.WaitGroup
			wg.Add(concurrency)
			
			var successCount int64
			var failCount int64
			var totalWaitTime int64
			
			for i := 0; i < concurrency; i++ {
				go func() {
					defer wg.Done()
					
					for j := 0; j < b.N; j++ {
						start := time.Now()
						conn, err := pool.Get()
						waitTime := time.Since(start)
						atomic.AddInt64(&totalWaitTime, int64(waitTime))
						
						if err != nil {
							atomic.AddInt64(&failCount, 1)
							continue
						}
						
						// 模拟数据库操作
						time.Sleep(time.Millisecond * 5)
						
						pool.Put(conn)
						atomic.AddInt64(&successCount, 1)
					}
				}()
			}
			
			wg.Wait()
			
			b.ReportMetric(float64(successCount), "success_count")
			b.ReportMetric(float64(failCount), "fail_count")
			
			if successCount > 0 {
				avgWait := time.Duration(totalWaitTime / successCount)
				b.ReportMetric(float64(avgWait.Microseconds()), "avg_wait_us")
			}
		})
	}
}

// BenchmarkDatabaseTransaction 事务性能测试
func BenchmarkDatabaseTransaction(b *testing.B) {
	db := NewMockDatabase(10000)
	
	b.Run("SingleTransaction", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			tx, err := db.BeginTx()
			if err != nil {
				b.Errorf("Begin transaction failed: %v", err)
				continue
			}
			
			// 执行多个操作
			for j := 0; j < 10; j++ {
				record := generatePointDataRecord()
				if err := tx.Insert("point_data", record); err != nil {
					tx.Rollback()
					b.Errorf("Insert failed: %v", err)
					continue
				}
			}
			
			if err := tx.Commit(); err != nil {
				b.Errorf("Commit failed: %v", err)
			}
		}
	})
	
	b.Run("BatchTransaction", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			tx, err := db.BeginTx()
			if err != nil {
				b.Errorf("Begin transaction failed: %v", err)
				continue
			}
			
			records := generatePointDataRecords(100)
			if err := tx.BatchInsert("point_data", records); err != nil {
				tx.Rollback()
				b.Errorf("Batch insert failed: %v", err)
				continue
			}
			
			if err := tx.Commit(); err != nil {
				b.Errorf("Commit failed: %v", err)
			}
		}
	})
}

// BenchmarkDatabaseReadWriteMix 读写混合测试
func BenchmarkDatabaseReadWriteMix(b *testing.B) {
	db := NewMockDatabase(100000)
	
	// 不同读写比例
	ratios := []struct {
		name       string
		readRatio  int
		writeRatio int
	}{
		{"ReadHeavy_90_10", 90, 10},
		{"ReadHeavy_80_20", 80, 20},
		{"Balanced_50_50", 50, 50},
		{"WriteHeavy_20_80", 20, 80},
		{"WriteHeavy_10_90", 10, 90},
	}
	
	for _, ratio := range ratios {
		b.Run(ratio.name, func(b *testing.B) {
			var readCount int64
			var writeCount int64
			var successCount int64
			var failCount int64
			
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				weight := i % 100
				
				if weight < ratio.readRatio {
					// 读操作
					atomic.AddInt64(&readCount, 1)
					query := &MockQuery{
						Table: "point_data",
						Limit: 100,
					}
					_, err := db.ExecuteQuery(context.Background(), query)
					if err != nil {
						atomic.AddInt64(&failCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}
				} else {
					// 写操作
					atomic.AddInt64(&writeCount, 1)
					record := generatePointDataRecord()
					err := db.Insert("point_data", record)
					if err != nil {
						atomic.AddInt64(&failCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}
				}
			}
			
			b.ReportMetric(float64(readCount), "read_count")
			b.ReportMetric(float64(writeCount), "write_count")
			b.ReportMetric(float64(successCount), "success_count")
			b.ReportMetric(float64(failCount), "fail_count")
		})
	}
}

// BenchmarkDatabaseConcurrentQuery 并发查询测试
func BenchmarkDatabaseConcurrentQuery(b *testing.B) {
	db := NewMockDatabase(1000000)
	
	concurrencyLevels := []int{10, 50, 100, 200, 500}
	
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			var wg sync.WaitGroup
			wg.Add(concurrency)
			
			var successCount int64
			var failCount int64
			var totalDuration int64
			
			b.ResetTimer()
			
			for i := 0; i < concurrency; i++ {
				go func(idx int) {
					defer wg.Done()
					
					for j := 0; j < b.N; j++ {
						start := time.Now()
						
						query := &MockQuery{
							Table: "point_data",
							Conditions: []MockCondition{
								{Field: "point_id", Operator: "=", Value: fmt.Sprintf("point-%d", idx)},
							},
							Limit: 100,
						}
						
						_, err := db.ExecuteQuery(context.Background(), query)
						duration := time.Since(start)
						
						atomic.AddInt64(&totalDuration, int64(duration))
						
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
			
			if successCount > 0 {
				avgDuration := time.Duration(totalDuration / successCount)
				b.ReportMetric(float64(avgDuration.Milliseconds()), "avg_duration_ms")
			}
		})
	}
}

// BenchmarkDatabaseMemoryUsage 内存使用测试
func BenchmarkDatabaseMemoryUsage(b *testing.B) {
	recordCounts := []int{10000, 50000, 100000, 500000, 1000000}
	
	for _, count := range recordCounts {
		b.Run(fmt.Sprintf("Records_%d", count), func(b *testing.B) {
			runtime.GC()
			var m1 runtime.MemStats
			runtime.ReadMemStats(&m1)
			
			db := NewMockDatabase(count)
			
			runtime.GC()
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)
			
			b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MB_allocated")
			b.ReportMetric(float64(m2.HeapAlloc)/1024/1024, "MB_heap")
			b.ReportMetric(float64(count), "record_count")
		})
	}
}

// 辅助类型和函数

type MockQuery struct {
	Table      string
	Conditions []MockCondition
	TimeRange  *MockTimeRange
	Aggregates []MockAggregate
	GroupBy    []string
	Joins      []MockJoin
	Limit      int
}

type MockCondition struct {
	Field    string
	Operator string
	Value    interface{}
}

type MockTimeRange struct {
	Field string
	Start time.Time
	End   time.Time
}

type MockAggregate struct {
	Field    string
	Function string
	Alias    string
}

type MockJoin struct {
	Type    string
	Table   string
	Alias   string
	OnField string
}

func generatePointDataRecords(count int) []map[string]interface{} {
	records := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		records[i] = generatePointDataRecord()
	}
	return records
}

func generatePointDataRecord() map[string]interface{} {
	return map[string]interface{}{
		"point_id":   fmt.Sprintf("point-%d", time.Now().UnixNano()%10000),
		"device_id":  fmt.Sprintf("device-%d", time.Now().UnixNano()%1000),
		"station_id": fmt.Sprintf("station-%d", time.Now().UnixNano()%100),
		"value":      float64(time.Now().UnixNano() % 10000) / 100,
		"quality":    192,
		"timestamp":  time.Now(),
	}
}

// MockConnectionPool 模拟连接池
type MockConnectionPool struct {
	connections chan *MockConnection
	maxOpen     int
	maxIdle     int
	maxLifetime time.Duration
	mu          sync.Mutex
}

type MockConnection struct {
	ID        int
	CreatedAt time.Time
}

func NewMockConnectionPool(maxOpen, maxIdle int, maxLifetime time.Duration) *MockConnectionPool {
	pool := &MockConnectionPool{
		connections: make(chan *MockConnection, maxOpen),
		maxOpen:     maxOpen,
		maxIdle:     maxIdle,
		maxLifetime: maxLifetime,
	}
	
	for i := 0; i < maxIdle; i++ {
		pool.connections <- &MockConnection{
			ID:        i,
			CreatedAt: time.Now(),
		}
	}
	
	return pool
}

func (p *MockConnectionPool) Get() (*MockConnection, error) {
	select {
	case conn := <-p.connections:
		return conn, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("connection pool timeout")
	}
}

func (p *MockConnectionPool) Put(conn *MockConnection) {
	select {
	case p.connections <- conn:
	default:
	}
}

func (p *MockConnectionPool) Close() {
	close(p.connections)
}

// MockTransaction 模拟事务
type MockTransaction struct {
	db *MockDatabase
}

func (tx *MockTransaction) Insert(table string, record map[string]interface{}) error {
	return tx.db.Insert(table, record)
}

func (tx *MockTransaction) BatchInsert(table string, records []map[string]interface{}) error {
	return tx.db.BatchInsert(table, records)
}

func (tx *MockTransaction) Commit() error {
	return nil
}

func (tx *MockTransaction) Rollback() error {
	return nil
}
