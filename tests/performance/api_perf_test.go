package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/api/dto"
)

// BenchmarkAPIResponseTime API响应时间测试
func BenchmarkAPIResponseTime(b *testing.B) {
	router := setupTestRouter()
	
	testCases := []struct {
		name    string
		method  string
		path    string
		body    interface{}
	}{
		{
			name:   "GetRegions",
			method: "GET",
			path:   "/api/v1/regions",
		},
		{
			name:   "GetStations",
			method: "GET",
			path:   "/api/v1/stations",
		},
		{
			name:   "GetDevices",
			method: "GET",
			path:   "/api/v1/devices",
		},
		{
			name:   "GetPoints",
			method: "GET",
			path:   "/api/v1/points",
		},
		{
			name:   "GetAlarms",
			method: "GET",
			path:   "/api/v1/alarms",
		},
		{
			name:   "GetRealtimeData",
			method: "GET",
			path:   "/api/v1/data/realtime?point_ids=point-1,point-2,point-3",
		},
		{
			name:   "GetHistoryData",
			method: "GET",
			path:   "/api/v1/data/history?point_id=point-1&start_time=1234567890&end_time=1234568890",
		},
		{
			name:   "CreateStation",
			method: "POST",
			path:   "/api/v1/stations",
			body: dto.CreateStationRequest{
				Name:     "test-station",
				Type:     "solar",
				Capacity: 100.0,
			},
		},
		{
			name:   "CreateDevice",
			method: "POST",
			path:   "/api/v1/devices",
			body: dto.CreateDeviceRequest{
				Name:      "test-device",
				StationID: "station-1",
				Type:      "inverter",
			},
		},
		{
			name:   "CreatePoint",
			method: "POST",
			path:   "/api/v1/points",
			body: dto.CreatePointRequest{
				Code:     "POINT-001",
				Name:     "Test Point",
				Type:     "yaoc",
				DeviceID: "device-1",
			},
		},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			var bodyReader io.Reader
			if tc.body != nil {
				bodyBytes, _ := json.Marshal(tc.body)
				bodyReader = bytes.NewReader(bodyBytes)
			}
			
			req, _ := http.NewRequest(tc.method, tc.path, bodyReader)
			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			
			b.ResetTimer()
			
			var totalDuration time.Duration
			for i := 0; i < b.N; i++ {
				start := time.Now()
				
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				
				duration := time.Since(start)
				totalDuration += duration
				
				if w.Code < 200 || w.Code >= 300 {
					b.Errorf("Unexpected status code: %d", w.Code)
				}
			}
			
			avgDuration := totalDuration / time.Duration(b.N)
			b.ReportMetric(float64(avgDuration.Microseconds()), "avg_latency_us")
			b.ReportMetric(float64(avgDuration.Milliseconds()), "avg_latency_ms")
		})
	}
}

// BenchmarkAPIConcurrentHandling API并发处理能力测试
func BenchmarkAPIConcurrentHandling(b *testing.B) {
	router := setupTestRouter()
	
	concurrencyLevels := []int{10, 50, 100, 200, 500, 1000}
	
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
						
						req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/stations?page=%d", idx), nil)
						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)
						
						duration := time.Since(start)
						atomic.AddInt64(&totalDuration, int64(duration))
						
						if w.Code >= 200 && w.Code < 300 {
							atomic.AddInt64(&successCount, 1)
						} else {
							atomic.AddInt64(&failCount, 1)
						}
					}
				}(i)
			}
			
			wg.Wait()
			
			totalRequests := successCount + failCount
			b.ReportMetric(float64(successCount), "success_count")
			b.ReportMetric(float64(failCount), "fail_count")
			b.ReportMetric(float64(totalRequests)/float64(b.N*concurrency)*100, "completion_rate")
			
			if successCount > 0 {
				avgDuration := time.Duration(totalDuration / successCount)
				b.ReportMetric(float64(avgDuration.Milliseconds()), "avg_duration_ms")
			}
		})
	}
}

// BenchmarkAPIThroughput API吞吐量测试
func BenchmarkAPIThroughput(b *testing.B) {
	router := setupTestRouter()
	
	duration := 10 * time.Second
	b.ResetTimer()
	
	var requestCount int64
	var successCount int64
	var errorCount int64
	
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	
	var wg sync.WaitGroup
	workerCount := 100
	
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for {
				select {
				case <-ctx.Done():
					return
				default:
					atomic.AddInt64(&requestCount, 1)
					
					req, _ := http.NewRequest("GET", "/api/v1/stations", nil)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
					
					if w.Code >= 200 && w.Code < 300 {
						atomic.AddInt64(&successCount, 1)
					} else {
						atomic.AddInt64(&errorCount, 1)
					}
				}
			}
		}()
	}
	
	wg.Wait()
	
	b.ReportMetric(float64(successCount)/duration.Seconds(), "requests_per_second")
	b.ReportMetric(float64(successCount), "total_success")
	b.ReportMetric(float64(errorCount), "total_errors")
	b.ReportMetric(float64(successCount)/float64(requestCount)*100, "success_rate")
}

// BenchmarkAPIMixedLoad 混合负载测试
func BenchmarkAPIMixedLoad(b *testing.B) {
	router := setupTestRouter()
	
	// 模拟真实场景：70%读，20%写，10%删除
	readWeight := 70
	writeWeight := 20
	deleteWeight := 10
	
	b.ResetTimer()
	
	var readCount int64
	var writeCount int64
	var deleteCount int64
	var successCount int64
	var failCount int64
	
	for i := 0; i < b.N; i++ {
		weight := i % 100
		
		var req *http.Request
		var operation string
		
		if weight < readWeight {
			// 读操作
			operation = "read"
			req, _ = http.NewRequest("GET", "/api/v1/stations", nil)
			atomic.AddInt64(&readCount, 1)
		} else if weight < readWeight+writeWeight {
			// 写操作
			operation = "write"
			body := dto.CreateStationRequest{
				Name:     fmt.Sprintf("station-%d", i),
				Type:     "solar",
				Capacity: 100.0,
			}
			bodyBytes, _ := json.Marshal(body)
			req, _ = http.NewRequest("POST", "/api/v1/stations", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			atomic.AddInt64(&writeCount, 1)
		} else {
			// 删除操作
			operation = "delete"
			req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/v1/stations/%d", i%1000), nil)
			atomic.AddInt64(&deleteCount, 1)
		}
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code >= 200 && w.Code < 300 {
			atomic.AddInt64(&successCount, 1)
		} else {
			atomic.AddInt64(&failCount, 1)
		}
	}
	
	b.ReportMetric(float64(readCount), "read_operations")
	b.ReportMetric(float64(writeCount), "write_operations")
	b.ReportMetric(float64(deleteCount), "delete_operations")
	b.ReportMetric(float64(successCount), "success_count")
	b.ReportMetric(float64(failCount), "fail_count")
}

// BenchmarkAPIDatabaseQueryPerformance 数据库查询性能测试
func BenchmarkAPIDatabaseQueryPerformance(b *testing.B) {
	router := setupTestRouter()
	db := setupTestDatabase()
	defer db.Close()
	
	// 预填充测试数据
	seedTestData(db, 10000)
	
	queryTypes := []struct {
		name string
		path string
	}{
		{
			name: "SimpleQuery",
			path: "/api/v1/stations?limit=100",
		},
		{
			name: "FilterQuery",
			path: "/api/v1/stations?type=solar&status=1",
		},
		{
			name: "SortQuery",
			path: "/api/v1/stations?sort=created_at&order=desc",
		},
		{
			name: "PaginationQuery",
			path: "/api/v1/stations?page=10&page_size=50",
		},
		{
			name: "ComplexQuery",
			path: "/api/v1/stations?type=solar&status=1&sort=created_at&order=desc&page=5&page_size=100",
		},
	}
	
	for _, qt := range queryTypes {
		b.Run(qt.name, func(b *testing.B) {
			req, _ := http.NewRequest("GET", qt.path, nil)
			
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				
				if w.Code < 200 || w.Code >= 300 {
					b.Errorf("Query failed with status: %d", w.Code)
				}
			}
		})
	}
}

// BenchmarkAPIJSONEncoding JSON编码性能测试
func BenchmarkAPIJSONEncoding(b *testing.B) {
	dataSizes := []int{10, 100, 1000, 10000}
	
	for _, size := range dataSizes {
		b.Run(fmt.Sprintf("DataSize_%d", size), func(b *testing.B) {
			data := generateTestData(size)
			
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				_, err := json.Marshal(data)
				if err != nil {
					b.Errorf("JSON encoding failed: %v", err)
				}
			}
			
			b.ReportMetric(0, "ops")
		})
	}
}

// BenchmarkAPIJSONDecoding JSON解码性能测试
func BenchmarkAPIJSONDecoding(b *testing.B) {
	dataSizes := []int{10, 100, 1000, 10000}
	
	for _, size := range dataSizes {
		b.Run(fmt.Sprintf("DataSize_%d", size), func(b *testing.B) {
			data := generateTestData(size)
			jsonData, _ := json.Marshal(data)
			
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				var result []map[string]interface{}
				err := json.Unmarshal(jsonData, &result)
				if err != nil {
					b.Errorf("JSON decoding failed: %v", err)
				}
			}
			
			b.ReportMetric(0, "ops")
		})
	}
}

// BenchmarkAPIMiddlewarePerformance 中间件性能测试
func BenchmarkAPIMiddlewarePerformance(b *testing.B) {
	router := setupTestRouter()
	
	middlewares := []struct {
		name string
		path string
	}{
		{
			name: "AuthMiddleware",
			path: "/api/v1/protected/stations",
		},
		{
			name: "RateLimitMiddleware",
			path: "/api/v1/rate-limited/stations",
		},
		{
			name: "LoggingMiddleware",
			path: "/api/v1/logged/stations",
		},
		{
			name: "CORSMiddleware",
			path: "/api/v1/cors/stations",
		},
	}
	
	for _, mw := range middlewares {
		b.Run(mw.name, func(b *testing.B) {
			req, _ := http.NewRequest("GET", mw.path, nil)
			req.Header.Set("Authorization", "Bearer test-token")
			
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
			}
		})
	}
}

// BenchmarkAPICPUProfile CPU性能分析
func BenchmarkAPICPUProfile(b *testing.B) {
	cpuProfile, err := os.Create("cpu_api.prof")
	if err != nil {
		b.Fatalf("Failed to create CPU profile: %v", err)
	}
	defer cpuProfile.Close()
	
	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		b.Fatalf("Failed to start CPU profile: %v", err)
	}
	defer pprof.StopCPUProfile()
	
	router := setupTestRouter()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/stations", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkAPIMemoryProfile 内存性能分析
func BenchmarkAPIMemoryProfile(b *testing.B) {
	router := setupTestRouter()
	
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/stations", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
	
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	
	b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MB_allocated")
	b.ReportMetric(float64(m2.HeapAlloc)/1024/1024, "MB_heap")
	b.ReportMetric(float64(m2.NumGC), "gc_cycles")
}

// BenchmarkAPIWithCache 缓存性能对比测试
func BenchmarkAPIWithCache(b *testing.B) {
	router := setupTestRouter()
	cache := setupTestCache()
	defer cache.Close()
	
	b.Run("WithoutCache", func(b *testing.B) {
		req, _ := http.NewRequest("GET", "/api/v1/stations", nil)
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
	
	b.Run("WithCache", func(b *testing.B) {
		req, _ := http.NewRequest("GET", "/api/v1/stations", nil)
		req.Header.Set("X-Enable-Cache", "true")
		
		// 预热缓存
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// 辅助函数

func setupTestRouter() http.Handler {
	mux := http.NewServeMux()
	
	// 注册路由
	mux.HandleFunc("/api/v1/regions", mockAPIHandler)
	mux.HandleFunc("/api/v1/stations", mockAPIHandler)
	mux.HandleFunc("/api/v1/devices", mockAPIHandler)
	mux.HandleFunc("/api/v1/points", mockAPIHandler)
	mux.HandleFunc("/api/v1/alarms", mockAPIHandler)
	mux.HandleFunc("/api/v1/data/realtime", mockAPIHandler)
	mux.HandleFunc("/api/v1/data/history", mockAPIHandler)
	
	return mux
}

func mockAPIHandler(w http.ResponseWriter, r *http.Request) {
	// 模拟处理延迟
	time.Sleep(time.Microsecond * time.Duration(100+time.Now().UnixNano()%900))
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    generateTestData(10),
	}
	
	json.NewEncoder(w).Encode(response)
}

func setupTestDatabase() *MockDatabase {
	return NewMockDatabase(10000)
}

func setupTestCache() *MockCache {
	return NewMockCache()
}

func seedTestData(db *MockDatabase, count int) {
	// 模拟填充测试数据
	for i := 0; i < count; i++ {
		db.data = append(db.data, map[string]interface{}{
			"id":        i,
			"name":      fmt.Sprintf("item-%d", i),
			"status":    1,
			"created_at": time.Now().Add(-time.Duration(i) * time.Minute),
		})
	}
}

func generateTestData(size int) []map[string]interface{} {
	data := make([]map[string]interface{}, size)
	for i := 0; i < size; i++ {
		data[i] = map[string]interface{}{
			"id":        i,
			"name":      fmt.Sprintf("item-%d", i),
			"value":     float64(i) * 1.5,
			"timestamp": time.Now().Unix(),
		}
	}
	return data
}

// MockCache 模拟缓存
type MockCache struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string][]byte),
	}
}

func (c *MockCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.data[key]
	return data, ok
}

func (c *MockCache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *MockCache) Close() error {
	return nil
}
