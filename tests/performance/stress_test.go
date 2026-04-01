package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// BenchmarkAPIPressure API接口压力测试
func BenchmarkAPIPressure(b *testing.B) {
	// 启动测试服务器
	server := NewMockAPIServer(":18080")
	go server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond) // 等待服务器启动

	endpoints := []struct {
		name   string
		method string
		path   string
		body   interface{}
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
			body: map[string]interface{}{
				"name":     "test-station",
				"type":     "solar",
				"capacity": 100.0,
			},
		},
	}

	for _, endpoint := range endpoints {
		b.Run(endpoint.name, func(b *testing.B) {
			client := &http.Client{
				Timeout: 10 * time.Second,
				Transport: &http.Transport{
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 100,
					IdleConnTimeout:     90 * time.Second,
				},
			}

			b.ResetTimer()

			var successCount int64
			var failCount int64
			var totalDuration int64

			for i := 0; i < b.N; i++ {
				start := time.Now()

				var req *http.Request
				var err error

				if endpoint.body != nil {
					bodyBytes, _ := json.Marshal(endpoint.body)
					req, err = http.NewRequest(endpoint.method, "http://localhost:18080"+endpoint.path, bytes.NewReader(bodyBytes))
					req.Header.Set("Content-Type", "application/json")
				} else {
					req, err = http.NewRequest(endpoint.method, "http://localhost:18080"+endpoint.path, nil)
				}

				if err != nil {
					b.Errorf("Failed to create request: %v", err)
					continue
				}

				resp, err := client.Do(req)
				if err != nil {
					atomic.AddInt64(&failCount, 1)
					continue
				}

				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				duration := time.Since(start)
				atomic.AddInt64(&totalDuration, int64(duration))

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&failCount, 1)
				}
			}

			b.ReportMetric(float64(successCount), "success_count")
			b.ReportMetric(float64(failCount), "fail_count")
			if successCount > 0 {
				avgDuration := time.Duration(totalDuration / successCount)
				b.ReportMetric(float64(avgDuration.Milliseconds()), "avg_duration_ms")
			}
		})
	}
}

// BenchmarkAPIConcurrentPressure API并发压力测试
func BenchmarkAPIConcurrentPressure(b *testing.B) {
	// 启动测试服务器
	server := NewMockAPIServer(":18081")
	go server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	concurrencyLevels := []int{10, 50, 100, 200, 500}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrent_%d", concurrency), func(b *testing.B) {
			client := &http.Client{
				Timeout: 10 * time.Second,
				Transport: &http.Transport{
					MaxIdleConns:        concurrency * 2,
					MaxIdleConnsPerHost: concurrency * 2,
					IdleConnTimeout:     90 * time.Second,
				},
			}

			b.ResetTimer()

			var wg sync.WaitGroup
			wg.Add(concurrency)

			var successCount int64
			var failCount int64

			for i := 0; i < concurrency; i++ {
				go func(idx int) {
					defer wg.Done()

					for j := 0; j < b.N; j++ {
						req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:18081/api/v1/stations?page=%d", idx), nil)
						resp, err := client.Do(req)
						if err != nil {
							atomic.AddInt64(&failCount, 1)
							continue
						}

						io.Copy(io.Discard, resp.Body)
						resp.Body.Close()

						if resp.StatusCode >= 200 && resp.StatusCode < 300 {
							atomic.AddInt64(&successCount, 1)
						} else {
							atomic.AddInt64(&failCount, 1)
						}
					}
				}(i)
			}

			wg.Wait()

			b.ReportMetric(float64(successCount), "success_count")
			b.ReportMetric(float64(failCount), "fail_count")
			b.ReportMetric(float64(successCount+failCount)/float64(b.N*concurrency)*100, "completion_rate")
		})
	}
}

// BenchmarkWebSocketPressure WebSocket连接压力测试
func BenchmarkWebSocketPressure(b *testing.B) {
	// 启动WebSocket服务器
	wsServer := NewMockWSServer(":18082")
	go wsServer.Start()
	defer wsServer.Stop()

	time.Sleep(100 * time.Millisecond)

	connectionCounts := []int{10, 50, 100, 200, 500}

	for _, count := range connectionCounts {
		b.Run(fmt.Sprintf("Connections_%d", count), func(b *testing.B) {
			connections := make([]*websocket.Conn, count)
			var connectErrors int64

			// 建立连接
			for i := 0; i < count; i++ {
				conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:18082/ws", nil)
				if err != nil {
					atomic.AddInt64(&connectErrors, 1)
					continue
				}
				connections[i] = conn
			}

			defer func() {
				for _, conn := range connections {
					if conn != nil {
						conn.Close()
					}
				}
			}()

			b.ResetTimer()

			var wg sync.WaitGroup
			wg.Add(count)

			var msgSent int64
			var msgReceived int64
			var msgErrors int64

			for i := 0; i < count; i++ {
				if connections[i] == nil {
					wg.Done()
					continue
				}

				go func(idx int, conn *websocket.Conn) {
					defer wg.Done()

					// 启动消息接收协程
					go func() {
						for {
							_, _, err := conn.ReadMessage()
							if err != nil {
								break
							}
							atomic.AddInt64(&msgReceived, 1)
						}
					}()

					// 发送消息
					for j := 0; j < b.N; j++ {
						msg := map[string]interface{}{
							"type":      "subscribe",
							"point_ids": []string{fmt.Sprintf("point-%d", idx)},
						}

						err := conn.WriteJSON(msg)
						if err != nil {
							atomic.AddInt64(&msgErrors, 1)
							continue
						}
						atomic.AddInt64(&msgSent, 1)

						time.Sleep(time.Millisecond * 10) // 控制发送速率
					}
				}(i, connections[i])
			}

			wg.Wait()

			b.ReportMetric(float64(msgSent), "messages_sent")
			b.ReportMetric(float64(msgReceived), "messages_received")
			b.ReportMetric(float64(msgErrors), "message_errors")
			b.ReportMetric(float64(connectErrors), "connect_errors")
		})
	}
}

// BenchmarkDatabasePoolPressure 数据库连接池压力测试
func BenchmarkDatabasePoolPressure(b *testing.B) {
	poolSizes := []int{10, 20, 50, 100, 200}

	for _, poolSize := range poolSizes {
		b.Run(fmt.Sprintf("PoolSize_%d", poolSize), func(b *testing.B) {
			pool := NewMockDBPool(poolSize)

			b.ResetTimer()

			var wg sync.WaitGroup
			wg.Add(poolSize)

			var successCount int64
			var failCount int64
			var totalWaitTime int64

			for i := 0; i < poolSize; i++ {
				go func(idx int) {
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
						time.Sleep(time.Millisecond * time.Duration(1+idx%10))

						pool.Put(conn)
						atomic.AddInt64(&successCount, 1)
					}
				}(i)
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

// BenchmarkMessageQueuePressure 消息队列压力测试
func BenchmarkMessageQueuePressure(b *testing.B) {
	queueSizes := []int{1000, 5000, 10000, 50000}

	for _, queueSize := range queueSizes {
		b.Run(fmt.Sprintf("QueueSize_%d", queueSize), func(b *testing.B) {
			mq := NewMockMessageQueue(queueSize)

			producerCount := 10
			consumerCount := 10

			b.ResetTimer()

			var wg sync.WaitGroup
			wg.Add(producerCount + consumerCount)

			var producedCount int64
			var consumedCount int64
			var produceErrors int64
			var consumeErrors int64

			// 启动生产者
			for i := 0; i < producerCount; i++ {
				go func(idx int) {
					defer wg.Done()

					for j := 0; j < b.N; j++ {
						msg := fmt.Sprintf("message-%d-%d", idx, j)
						err := mq.Publish("test-topic", []byte(msg))
						if err != nil {
							atomic.AddInt64(&produceErrors, 1)
						} else {
							atomic.AddInt64(&producedCount, 1)
						}
					}
				}(i)
			}

			// 启动消费者
			for i := 0; i < consumerCount; i++ {
				go func(idx int) {
					defer wg.Done()

					for j := 0; j < b.N; j++ {
						msg, err := mq.Consume("test-topic")
						if err != nil {
							atomic.AddInt64(&consumeErrors, 1)
						} else if msg != nil {
							atomic.AddInt64(&consumedCount, 1)
						}
					}
				}(i)
			}

			wg.Wait()

			b.ReportMetric(float64(producedCount), "messages_produced")
			b.ReportMetric(float64(consumedCount), "messages_consumed")
			b.ReportMetric(float64(produceErrors), "produce_errors")
			b.ReportMetric(float64(consumeErrors), "consume_errors")
		})
	}
}

// BenchmarkMemoryPressure 内存压力测试
func BenchmarkMemoryPressure(b *testing.B) {
	dataSizes := []int{1024, 10240, 102400, 1048576} // 1KB, 10KB, 100KB, 1MB

	for _, size := range dataSizes {
		b.Run(fmt.Sprintf("DataSize_%dKB", size/1024), func(b *testing.B) {
			runtime.GC()
			var m1 runtime.MemStats
			runtime.ReadMemStats(&m1)

			b.ResetTimer()

			data := make([][]byte, b.N)
			for i := 0; i < b.N; i++ {
				data[i] = make([]byte, size)
				for j := 0; j < size; j++ {
					data[i][j] = byte(j % 256)
				}
			}

			runtime.GC()
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)

			b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MB_allocated")
			b.ReportMetric(float64(m2.HeapInuse)/1024/1024, "MB_heap_inuse")
			b.ReportMetric(float64(m2.NumGC), "gc_cycles")
		})
	}
}

// BenchmarkCPUIntensiveTask CPU密集型任务压力测试
func BenchmarkCPUIntensiveTask(b *testing.B) {
	cpuProfile, err := os.Create("cpu_stress.prof")
	if err != nil {
		b.Fatalf("Failed to create CPU profile: %v", err)
	}
	defer cpuProfile.Close()

	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		b.Fatalf("Failed to start CPU profile: %v", err)
	}
	defer pprof.StopCPUProfile()

	taskCounts := []int{10, 50, 100, 200}

	for _, tasks := range taskCounts {
		b.Run(fmt.Sprintf("Tasks_%d", tasks), func(b *testing.B) {
			b.ResetTimer()

			var wg sync.WaitGroup
			wg.Add(tasks)

			for i := 0; i < tasks; i++ {
				go func() {
					defer wg.Done()

					for j := 0; j < b.N; j++ {
						// CPU密集型计算
						result := 0
						for k := 0; k < 10000; k++ {
							result += k * k
						}
						_ = result
					}
				}()
			}

			wg.Wait()
		})
	}
}

// MockAPIServer 模拟API服务器
type MockAPIServer struct {
	addr   string
	server *http.Server
}

// NewMockAPIServer 创建模拟API服务器
func NewMockAPIServer(addr string) *MockAPIServer {
	mux := http.NewServeMux()

	// 注册路由
	mux.HandleFunc("/api/v1/regions", mockHandler)
	mux.HandleFunc("/api/v1/stations", mockHandler)
	mux.HandleFunc("/api/v1/devices", mockHandler)
	mux.HandleFunc("/api/v1/points", mockHandler)
	mux.HandleFunc("/api/v1/alarms", mockHandler)
	mux.HandleFunc("/api/v1/data/realtime", mockHandler)
	mux.HandleFunc("/api/v1/data/history", mockHandler)

	return &MockAPIServer{
		addr: addr,
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

// Start 启动服务器
func (s *MockAPIServer) Start() {
	s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *MockAPIServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.server.Shutdown(ctx)
}

// mockHandler 模拟处理器
func mockHandler(w http.ResponseWriter, r *http.Request) {
	// 模拟处理延迟
	time.Sleep(time.Millisecond * time.Duration(1+time.Now().UnixNano()%10))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    []interface{}{},
	})
}

// MockWSServer 模拟WebSocket服务器
type MockWSServer struct {
	addr   string
	server *http.Server
	upgrader websocket.Upgrader
}

// NewMockWSServer 创建模拟WebSocket服务器
func NewMockWSServer(addr string) *MockWSServer {
	s := &MockWSServer{
		addr: addr,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWS)

	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return s
}

// Start 启动服务器
func (s *MockWSServer) Start() {
	s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *MockWSServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.server.Shutdown(ctx)
}

// handleWS 处理WebSocket连接
func (s *MockWSServer) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 模拟数据推送
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			msg := map[string]interface{}{
				"type":      "data",
				"point_id":  "point-1",
				"value":     123.45,
				"timestamp": time.Now().Unix(),
			}
			if err := conn.WriteJSON(msg); err != nil {
				break
			}
		}
	}()

	// 读取客户端消息
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// MockDBPool 模拟数据库连接池
type MockDBPool struct {
	connections chan *MockDBConnection
	size        int
}

// MockDBConnection 模拟数据库连接
type MockDBConnection struct {
	ID        int
	CreatedAt time.Time
}

// NewMockDBPool 创建模拟数据库连接池
func NewMockDBPool(size int) *MockDBPool {
	pool := &MockDBPool{
		connections: make(chan *MockDBConnection, size),
		size:        size,
	}

	for i := 0; i < size; i++ {
		pool.connections <- &MockDBConnection{
			ID:        i,
			CreatedAt: time.Now(),
		}
	}

	return pool
}

// Get 获取连接
func (p *MockDBPool) Get() (*MockDBConnection, error) {
	select {
	case conn := <-p.connections:
		return conn, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("connection pool timeout")
	}
}

// Put 归还连接
func (p *MockDBPool) Put(conn *MockDBConnection) {
	select {
	case p.connections <- conn:
	default:
		// 连接池已满，丢弃连接
	}
}

// MockMessageQueue 模拟消息队列
type MockMessageQueue struct {
	topic   string
	queue   chan []byte
	maxSize int
	mu      sync.RWMutex
}

// NewMockMessageQueue 创建模拟消息队列
func NewMockMessageQueue(maxSize int) *MockMessageQueue {
	return &MockMessageQueue{
		topic:   "test-topic",
		queue:   make(chan []byte, maxSize),
		maxSize: maxSize,
	}
}

// Publish 发布消息
func (mq *MockMessageQueue) Publish(topic string, message []byte) error {
	select {
	case mq.queue <- message:
		return nil
	default:
		return fmt.Errorf("queue full")
	}
}

// Consume 消费消息
func (mq *MockMessageQueue) Consume(topic string) ([]byte, error) {
	select {
	case msg := <-mq.queue:
		return msg, nil
	case <-time.After(100 * time.Millisecond):
		return nil, nil
	}
}
