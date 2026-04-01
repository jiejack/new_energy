package rule

import (
	"context"
	"testing"
	"time"
)

// TestPointManager 测试计算点管理器
func TestPointManager(t *testing.T) {
	pm := NewPointManager()
	ctx := context.Background()

	// 测试创建计算点
	point := &ComputePoint{
		ID:           "test-point-1",
		Name:         "Test Point",
		Type:         PointTypeVirtual,
		Formula:      "A + B",
		Dependencies: []string{"point-a", "point-b"},
	}

	err := pm.CreatePoint(ctx, point)
	if err != nil {
		t.Fatalf("Failed to create point: %v", err)
	}

	// 测试获取计算点
	retrieved, err := pm.GetPoint(ctx, "test-point-1")
	if err != nil {
		t.Fatalf("Failed to get point: %v", err)
	}

	if retrieved.ID != point.ID {
		t.Errorf("Expected ID %s, got %s", point.ID, retrieved.ID)
	}

	// 测试统计
	stats := pm.GetStats(ctx)
	if stats.TotalPoints != 1 {
		t.Errorf("Expected 1 total point, got %d", stats.TotalPoints)
	}
}

// TestComputeCache 测试计算结果缓存
func TestComputeCache(t *testing.T) {
	config := &CacheConfig{
		EnableLocalCache: true,
		EnableRedisCache: false,
		LocalCacheSize:   100,
		DefaultTTL:       5 * time.Minute,
		Policy:           CachePolicyLRU,
	}

	cache := NewComputeCache(config, nil)
	ctx := context.Background()

	// 测试设置缓存
	result := &ComputeResult{
		PointID:   "test-point",
		Value:     100.5,
		Quality:   100,
		Timestamp: time.Now(),
	}

	err := cache.Set(ctx, "test-key", result)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// 测试获取缓存
	retrieved, err := cache.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}

	if retrieved.Value != result.Value {
		t.Errorf("Expected value %f, got %f", result.Value, retrieved.Value)
	}

	// 测试命中率
	hitRate := cache.GetHitRate()
	if hitRate != 1.0 {
		t.Errorf("Expected hit rate 1.0, got %f", hitRate)
	}
}

// TestLocalCache 测试本地缓存
func TestLocalCache(t *testing.T) {
	cache := NewLocalCache(100, 5*time.Minute, CachePolicyLRU)

	// 测试设置
	result := &ComputeResult{
		PointID:   "test",
		Value:     50.0,
		Timestamp: time.Now(),
	}

	cache.Set("key1", result)

	// 测试获取
	retrieved, exists := cache.Get("key1")
	if !exists {
		t.Fatal("Cache key should exist")
	}

	if retrieved.Value != result.Value {
		t.Errorf("Expected value %f, got %f", result.Value, retrieved.Value)
	}

	// 测试删除
	cache.Delete("key1")
	_, exists = cache.Get("key1")
	if exists {
		t.Error("Cache key should not exist after delete")
	}
}

// TestLocalLock 测试本地锁
func TestLocalLock(t *testing.T) {
	lock := NewLocalLock()
	ctx := context.Background()

	// 测试获取锁
	acquired, err := lock.Acquire(ctx, "test-lock", 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	if !acquired {
		t.Error("Lock should be acquired")
	}

	// 测试锁是否被持有
	held, err := lock.IsHeld(ctx, "test-lock")
	if err != nil {
		t.Fatalf("Failed to check lock: %v", err)
	}

	if !held {
		t.Error("Lock should be held")
	}

	// 测试释放锁
	err = lock.Release(ctx, "test-lock")
	if err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	// 再次检查
	held, _ = lock.IsHeld(ctx, "test-lock")
	if held {
		t.Error("Lock should not be held after release")
	}
}

// TestPriorityQueue 测试优先级队列
func TestPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue()

	task1 := &ComputeTask{
		ID:       "task-1",
		Priority: 5,
	}

	task2 := &ComputeTask{
		ID:       "task-2",
		Priority: 10,
	}

	task3 := &ComputeTask{
		ID:       "task-3",
		Priority: 3,
	}

	pq.Push(task1)
	pq.Push(task2)
	pq.Push(task3)

	// 应该按优先级顺序取出
	first := pq.Pop()
	if first.ID != "task-2" {
		t.Errorf("Expected task-2 (priority 10), got %s", first.ID)
	}

	second := pq.Pop()
	if second.ID != "task-1" {
		t.Errorf("Expected task-1 (priority 5), got %s", second.ID)
	}

	third := pq.Pop()
	if third.ID != "task-3" {
		t.Errorf("Expected task-3 (priority 3), got %s", third.ID)
	}
}

// TestDependencyGraph 测试依赖图
func TestDependencyGraph(t *testing.T) {
	dg := NewDependencyGraph()

	// 添加节点
	dg.AddNode("A", []string{})
	dg.AddNode("B", []string{"A"})
	dg.AddNode("C", []string{"A", "B"})

	// 测试拓扑排序
	order := dg.GetTopologicalOrder()
	if len(order) == 0 {
		t.Error("Topological order should not be empty")
	}

	// A应该在第一层
	found := false
	for _, id := range order[0] {
		if id == "A" {
			found = true
			break
		}
	}
	if !found {
		t.Error("A should be in the first level")
	}
}

// TestTriggerManager 测试触发器管理器
func TestTriggerManager(t *testing.T) {
	// 创建模拟执行器
	executor := &mockExecutor{}
	tm := NewTriggerManager(executor)

	// 测试创建触发器
	trigger := &Trigger{
		ID:       "trigger-1",
		Name:     "Test Trigger",
		Type:     TriggerTypeDataChange,
		Enabled:  true,
		PointIDs: []string{"point-1"},
		Condition: &TriggerCondition{
			ChangeThreshold: floatPtr(10.0),
		},
	}

	err := tm.CreateTrigger(trigger)
	if err != nil {
		t.Fatalf("Failed to create trigger: %v", err)
	}

	// 测试获取触发器
	retrieved, err := tm.GetTrigger("trigger-1")
	if err != nil {
		t.Fatalf("Failed to get trigger: %v", err)
	}

	if retrieved.ID != trigger.ID {
		t.Errorf("Expected ID %s, got %s", trigger.ID, retrieved.ID)
	}

	// 测试按计算点获取
	triggers := tm.GetTriggersByPoint("point-1")
	if len(triggers) != 1 {
		t.Errorf("Expected 1 trigger, got %d", len(triggers))
	}
}

// TestRuleEngine 测试规则引擎
func TestRuleEngine(t *testing.T) {
	// 创建模拟数据提供者
	provider := &mockDataProvider{}
	cache := NewComputeCache(&CacheConfig{
		EnableLocalCache: true,
		LocalCacheSize:   100,
		DefaultTTL:       5 * time.Minute,
	}, nil)

	engine := NewRuleEngine(cache, provider)

	// 测试加载规则
	rule := &Rule{
		ID:        "rule-1",
		Name:      "Test Rule",
		Type:      RuleTypeFormula,
		Enabled:   true,
		PointID:   "point-1",
		Formula:   "input1 + input2",
		Inputs: []RuleInput{
			{Name: "input1", PointID: "point-a", Required: true},
			{Name: "input2", PointID: "point-b", Required: true},
		},
		Timeout: 5 * time.Second,
	}

	err := engine.LoadRule(rule)
	if err != nil {
		t.Fatalf("Failed to load rule: %v", err)
	}

	// 测试获取规则
	retrieved, err := engine.GetRule("rule-1")
	if err != nil {
		t.Fatalf("Failed to get rule: %v", err)
	}

	if retrieved.ID != rule.ID {
		t.Errorf("Expected ID %s, got %s", rule.ID, retrieved.ID)
	}

	// 测试统计
	stats := engine.GetStats()
	if stats.TotalRules != 1 {
		t.Errorf("Expected 1 total rule, got %d", stats.TotalRules)
	}
}

// mockExecutor 模拟执行器
type mockExecutor struct{}

func (m *mockExecutor) Execute(ctx context.Context, pointIDs []string) (map[string]*ComputeResult, error) {
	results := make(map[string]*ComputeResult)
	for _, id := range pointIDs {
		results[id] = &ComputeResult{
			PointID:   id,
			Value:     100.0,
			Timestamp: time.Now(),
		}
	}
	return results, nil
}

// mockDataProvider 模拟数据提供者
type mockDataProvider struct{}

func (m *mockDataProvider) GetCurrentValue(ctx context.Context, pointID string) (float64, error) {
	return 50.0, nil
}

func (m *mockDataProvider) GetTimeSeries(ctx context.Context, pointID string, start, end time.Time) ([]float64, error) {
	return []float64{10.0, 20.0, 30.0}, nil
}

func (m *mockDataProvider) GetAggregatedValue(ctx context.Context, pointID string, aggFunc string, window time.Duration) (float64, error) {
	return 25.0, nil
}

// floatPtr 辅助函数
func floatPtr(v float64) *float64 {
	return &v
}
