package performance

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockDatabase 扩展方法

// Insert 插入数据
func (m *MockDatabase) Insert(table string, record map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 模拟插入延迟
	time.Sleep(time.Microsecond * 10)

	m.data = append(m.data, record)
	m.records++
	return nil
}

// BatchInsert 批量插入
func (m *MockDatabase) BatchInsert(table string, records []map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 模拟批量插入延迟
	time.Sleep(time.Microsecond * time.Duration(len(records)/100))

	m.data = append(m.data, records...)
	m.records += len(records)
	return nil
}

// Update 更新数据
func (m *MockDatabase) Update(table string, conditions map[string]interface{}, updates map[string]interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var affected int64
	for i, record := range m.data {
		if matchConditions(record, conditions) {
			for k, v := range updates {
				m.data[i][k] = v
			}
			affected++
		}
	}

	return affected, nil
}

// Delete 删除数据
func (m *MockDatabase) Delete(table string, conditions map[string]interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var newRecords []map[string]interface{}
	var deleted int64

	for _, record := range m.data {
		if !matchConditions(record, conditions) {
			newRecords = append(newRecords, record)
		} else {
			deleted++
		}
	}

	m.data = newRecords
	m.records = len(m.data)
	return deleted, nil
}

// ExecuteQuery 执行查询
func (m *MockDatabase) ExecuteQuery(ctx context.Context, query *MockQuery) ([]map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 模拟查询延迟
	time.Sleep(time.Microsecond * time.Duration(m.records/10000))

	var result []map[string]interface{}

	// 应用条件过滤
	for _, record := range m.data {
		if matchQueryConditions(record, query) {
			result = append(result, record)
			if query.Limit > 0 && len(result) >= query.Limit {
				break
			}
		}
	}

	return result, nil
}

// CreateIndex 创建索引
func (m *MockDatabase) CreateIndex(table, field string) error {
	// 模拟创建索引
	return nil
}

// BeginTx 开始事务
func (m *MockDatabase) BeginTx() (*MockTransaction, error) {
	return &MockTransaction{db: m}, nil
}

// Close 关闭数据库
func (m *MockDatabase) Close() error {
	return nil
}

// matchConditions 检查记录是否匹配条件
func matchConditions(record map[string]interface{}, conditions map[string]interface{}) bool {
	for k, v := range conditions {
		if record[k] != v {
			return false
		}
	}
	return true
}

// matchQueryConditions 检查记录是否匹配查询条件
func matchQueryConditions(record map[string]interface{}, query *MockQuery) bool {
	// 检查普通条件
	for _, cond := range query.Conditions {
		value, exists := record[cond.Field]
		if !exists {
			return false
		}

		switch cond.Operator {
		case "=":
			if value != cond.Value {
				return false
			}
		case ">":
			if compareValues(value, cond.Value) <= 0 {
				return false
			}
		case "<":
			if compareValues(value, cond.Value) >= 0 {
				return false
			}
		case ">=":
			if compareValues(value, cond.Value) < 0 {
				return false
			}
		case "<=":
			if compareValues(value, cond.Value) > 0 {
				return false
			}
		}
	}

	// 检查时间范围
	if query.TimeRange != nil {
		timestamp, ok := record[query.TimeRange.Field].(time.Time)
		if !ok {
			return false
		}
		if timestamp.Before(query.TimeRange.Start) || timestamp.After(query.TimeRange.End) {
			return false
		}
	}

	return true
}

// compareValues 比较两个值
func compareValues(a, b interface{}) int {
	switch a.(type) {
	case int:
		ai := a.(int)
		bi := b.(int)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case float64:
		af := a.(float64)
		bf := b.(float64)
		if af < bf {
			return -1
		} else if af > bf {
			return 1
		}
		return 0
	case string:
		as := a.(string)
		bs := b.(string)
		if as < bs {
			return -1
		} else if as > bs {
			return 1
		}
		return 0
	default:
		return 0
	}
}

// MockConnectionPool 模拟连接池扩展

// GetStats 获取连接池统计
func (p *MockConnectionPool) GetStats() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	return map[string]interface{}{
		"total_connections": p.maxOpen,
		"idle_connections":  len(p.connections),
		"max_lifetime":      p.maxLifetime,
	}
}

// MockMessageQueue 扩展方法

// GetStats 获取消息队列统计
func (mq *MockMessageQueue) GetStats() map[string]interface{} {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	return map[string]interface{}{
		"queue_size":   len(mq.queue),
		"max_size":     mq.maxSize,
		"usage_percent": float64(len(mq.queue)) / float64(mq.maxSize) * 100,
	}
}

// MockCache 扩展方法

// GetStats 获取缓存统计
func (c *MockCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"total_keys":   len(c.data),
		"memory_usage": len(c.data) * 100, // 简化估算
	}
}

// Delete 删除缓存
func (c *MockCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}

// Clear 清空缓存
func (c *MockCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string][]byte)
	return nil
}

// MockCollector 扩展方法

// CollectorConfig 采集器配置
type CollectorConfig struct {
	ID         string
	Name       string
	Protocol   string
	Interval   time.Duration
	BufferSize int
	BatchSize  int
	MaxWorkers int
	MaxPoints  int
}

// Initialize 初始化
func (m *MockCollector) Initialize(ctx context.Context, config *CollectorConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = &collector.CollectorConfig{
		ID:         config.ID,
		Name:       config.Name,
		Protocol:   config.Protocol,
		Interval:   config.Interval,
		BufferSize: config.BufferSize,
		BatchSize:  config.BatchSize,
		MaxWorkers: config.MaxWorkers,
		MaxPoints:  config.MaxPoints,
	}
	m.status = collector.StatusInitialized
	return nil
}

// MockWorkerPool 模拟工作池
type MockWorkerPool struct {
	workers    int
	tasks      chan MockTask
	results    chan MockResult
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

type MockTask struct {
	ID   string
	Data interface{}
}

type MockResult struct {
	ID     string
	Error  error
	Result interface{}
}

func NewMockWorkerPool(workers int) *MockWorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &MockWorkerPool{
		workers: workers,
		tasks:   make(chan MockTask, workers*10),
		results: make(chan MockResult, workers*10),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (p *MockWorkerPool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *MockWorkerPool) worker(id int) {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		case task := <-p.tasks:
			// 模拟处理任务
			time.Sleep(time.Microsecond * 100)
			p.results <- MockResult{
				ID:     task.ID,
				Result: fmt.Sprintf("processed by worker %d", id),
			}
		}
	}
}

func (p *MockWorkerPool) Submit(task MockTask) error {
	select {
	case p.tasks <- task:
		return nil
	default:
		return fmt.Errorf("task queue full")
	}
}

func (p *MockWorkerPool) Stop() {
	p.cancel()
	p.wg.Wait()
}

// MockDataBuffer 模拟数据缓冲区
type MockDataBuffer struct {
	data     [][]byte
	size     int
	capacity int
	mu       sync.RWMutex
}

func NewMockDataBuffer(capacity int) *MockDataBuffer {
	return &MockDataBuffer{
		data:     make([][]byte, 0, capacity),
		capacity: capacity,
	}
}

func (b *MockDataBuffer) Write(data []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.size+len(data) > b.capacity {
		// 淘汰旧数据
		overflow := b.size + len(data) - b.capacity
		for b.size > 0 && overflow > 0 {
			b.size -= len(b.data[0])
			b.data = b.data[1:]
			overflow -= b.size
		}
	}

	b.data = append(b.data, data)
	b.size += len(data)
	return nil
}

func (b *MockDataBuffer) Read() ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.data) == 0 {
		return nil, fmt.Errorf("buffer empty")
	}

	data := b.data[0]
	b.data = b.data[1:]
	b.size -= len(data)
	return data, nil
}

func (b *MockDataBuffer) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

func (b *MockDataBuffer) Capacity() int {
	return b.capacity
}
