package rule

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/new-energy-monitoring/internal/infrastructure/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	ErrTriggerNotFound   = errors.New("trigger not found")
	ErrTriggerExists     = errors.New("trigger already exists")
	ErrInvalidTrigger    = errors.New("invalid trigger configuration")
	ErrTriggerDisabled   = errors.New("trigger is disabled")
	ErrConditionNotMet   = errors.New("trigger condition not met")
)

// Prometheus指标
var (
	triggerTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compute_trigger_total",
		Help: "Total number of trigger activations",
	}, []string{"type", "status"})

	triggerLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "compute_trigger_latency_seconds",
		Help:    "Trigger processing latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"type"})

	triggerActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "compute_trigger_active",
		Help: "Number of active triggers",
	})
)

// TriggerType 触发器类型
type TriggerType string

const (
	TriggerTypeDataChange TriggerType = "data_change" // 数据变化触发
	TriggerTypeEvent      TriggerType = "event"       // 事件触发
	TriggerTypeCondition  TriggerType = "condition"   // 条件触发
	TriggerTypeComposite  TriggerType = "composite"   // 复合触发
)

// TriggerStatus 触发器状态
type TriggerStatus string

const (
	TriggerStatusActive   TriggerStatus = "active"   // 活跃
	TriggerStatusInactive TriggerStatus = "inactive" // 不活跃
	TriggerStatusError    TriggerStatus = "error"    // 错误
	TriggerStatusDisabled TriggerStatus = "disabled" // 禁用
)

// Trigger 触发器结构
type Trigger struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        TriggerType            `json:"type"`
	Status      TriggerStatus          `json:"status"`
	Priority    int                    `json:"priority"`    // 优先级
	Enabled     bool                   `json:"enabled"`     // 是否启用
	PointIDs    []string               `json:"pointIds"`    // 关联的计算点ID
	Condition   *TriggerCondition      `json:"condition"`   // 触发条件
	Actions     []TriggerAction        `json:"actions"`     // 触发动作
	ChainConfig *TriggerChainConfig    `json:"chainConfig"` // 触发器链配置
	Config      map[string]interface{} `json:"config"`      // 配置参数
	CreateTime  time.Time              `json:"createTime"`
	UpdateTime  time.Time              `json:"updateTime"`
	LastTrigger time.Time              `json:"lastTrigger"` // 最后触发时间
	TriggerCount int64                 `json:"triggerCount"` // 触发次数
	ErrorCount  int64                  `json:"errorCount"`   // 错误次数
	LastError   string                 `json:"lastError"`    // 最后错误
}

// TriggerCondition 触发条件
type TriggerCondition struct {
	// 数据变化触发
	ChangeThreshold *float64 `json:"changeThreshold"` // 变化阈值
	ChangePercent   *float64 `json:"changePercent"`   // 变化百分比
	Direction       string   `json:"direction"`       // up/down/both

	// 事件触发
	EventType string   `json:"eventType"`   // 事件类型
	EventData string   `json:"eventData"`   // 事件数据匹配

	// 条件触发
	Expression string   `json:"expression"` // 条件表达式
	Variables  []string `json:"variables"`  // 变量列表

	// 复合触发
	LogicOperator string                `json:"logicOperator"` // and/or
	SubConditions []*TriggerCondition   `json:"subConditions"` // 子条件

	// 通用
	CooldownTime time.Duration `json:"cooldownTime"` // 冷却时间
	MinInterval  time.Duration `json:"minInterval"`  // 最小触发间隔
}

// TriggerAction 触发动作
type TriggerAction struct {
	Type       string                 `json:"type"`       // 动作类型: compute, notify, script
	Target     string                 `json:"target"`     // 目标
	Params     map[string]interface{} `json:"params"`     // 参数
	Async      bool                   `json:"async"`      // 是否异步执行
	Timeout    time.Duration          `json:"timeout"`    // 超时时间
	RetryCount int                    `json:"retryCount"` // 重试次数
}

// TriggerChainConfig 触发器链配置
type TriggerChainConfig struct {
	ChainID      string   `json:"chainId"`      // 链ID
	Position     int      `json:"position"`     // 链中位置
	NextTriggers []string `json:"nextTriggers"` // 下一个触发器ID列表
	StopOnSuccess bool    `json:"stopOnSuccess"` // 成功后是否停止
	StopOnFailure bool    `json:"stopOnFailure"` // 失败后是否停止
}

// TriggerEvent 触发事件
type TriggerEvent struct {
	ID         string                 `json:"id"`
	Type       TriggerType            `json:"type"`
	Source     string                 `json:"source"`     // 事件源
	PointID    string                 `json:"pointId"`    // 相关计算点ID
	OldValue   float64                `json:"oldValue"`   // 旧值
	NewValue   float64                `json:"newValue"`   // 新值
	Timestamp  time.Time              `json:"timestamp"`  // 时间戳
	Payload    map[string]interface{} `json:"payload"`    // 载荷
	Metadata   map[string]string      `json:"metadata"`   // 元数据
}

// TriggerResult 触发结果
type TriggerResult struct {
	TriggerID  string                 `json:"triggerId"`
	Triggered  bool                   `json:"triggered"`
	Executed   bool                   `json:"executed"`
	StartTime  time.Time              `json:"startTime"`
	EndTime    time.Time              `json:"endTime"`
	Duration   time.Duration          `json:"duration"`
	Error      string                 `json:"error"`
	Actions    []ActionResult         `json:"actions"`
	Context    map[string]interface{} `json:"context"`
}

// ActionResult 动作执行结果
type ActionResult struct {
	Type     string        `json:"type"`
	Target   string        `json:"target"`
	Success  bool          `json:"success"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error"`
	Output   interface{}   `json:"output"`
}

// TriggerManager 触发器管理器
type TriggerManager struct {
	triggers   map[string]*Trigger
	byPoint    map[string][]string // 按计算点索引
	byType     map[TriggerType][]string // 按类型索引
	chains     map[string]*TriggerChain
	executor   ComputeExecutor
	eventQueue chan *TriggerEvent
	mu         sync.RWMutex
	running    int32
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
	logger     *zap.Logger
}

// TriggerChain 触发器链
type TriggerChain struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Triggers    []string   `json:"triggers"`    // 触发器ID列表（按顺序）
	Parallel    bool       `json:"parallel"`    // 是否并行执行
	StopOnFirst bool       `json:"stopOnFirst"` // 第一个成功后停止
	Status      string     `json:"status"`
}

// NewTriggerManager 创建触发器管理器
func NewTriggerManager(executor ComputeExecutor) *TriggerManager {
	var log *zap.Logger
	if logger.Log != nil {
		log = logger.Named("trigger-manager")
	} else {
		log = zap.NewNop()
	}
	
	tm := &TriggerManager{
		triggers:   make(map[string]*Trigger),
		byPoint:    make(map[string][]string),
		byType:     make(map[TriggerType][]string),
		chains:     make(map[string]*TriggerChain),
		executor:   executor,
		eventQueue: make(chan *TriggerEvent, 10000),
		logger:     log,
	}
	tm.ctx, tm.cancelFunc = context.WithCancel(context.Background())
	return tm
}

// Start 启动触发器管理器
func (tm *TriggerManager) Start() error {
	if atomic.LoadInt32(&tm.running) == 1 {
		return errors.New("trigger manager is already running")
	}

	atomic.StoreInt32(&tm.running, 1)

	tm.logger.Info("Starting trigger manager")

	// 启动事件处理协程
	tm.wg.Add(1)
	go tm.processEvents()

	// 启动指标收集
	tm.wg.Add(1)
	go tm.collectMetrics()

	return nil
}

// Stop 停止触发器管理器
func (tm *TriggerManager) Stop() error {
	if atomic.LoadInt32(&tm.running) == 0 {
		return errors.New("trigger manager is not running")
	}

	tm.logger.Info("Stopping trigger manager")

	atomic.StoreInt32(&tm.running, 0)
	tm.cancelFunc()

	done := make(chan struct{})
	go func() {
		tm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(10 * time.Second):
		return errors.New("stop timeout")
	}
}

// CreateTrigger 创建触发器
func (tm *TriggerManager) CreateTrigger(trigger *Trigger) error {
	if trigger.ID == "" {
		return ErrInvalidTrigger
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.triggers[trigger.ID]; exists {
		return ErrTriggerExists
	}

	// 设置默认值
	if trigger.Status == "" {
		trigger.Status = TriggerStatusActive
	}
	if trigger.CreateTime.IsZero() {
		trigger.CreateTime = time.Now()
	}
	trigger.UpdateTime = time.Now()

	tm.triggers[trigger.ID] = trigger

	// 建立索引
	for _, pointID := range trigger.PointIDs {
		tm.byPoint[pointID] = append(tm.byPoint[pointID], trigger.ID)
	}
	tm.byType[trigger.Type] = append(tm.byType[trigger.Type], trigger.ID)

	tm.logger.Info("Trigger created",
		zap.String("triggerID", trigger.ID),
		zap.String("type", string(trigger.Type)),
		zap.Int("priority", trigger.Priority))

	return nil
}

// UpdateTrigger 更新触发器
func (tm *TriggerManager) UpdateTrigger(trigger *Trigger) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	existing, exists := tm.triggers[trigger.ID]
	if !exists {
		return ErrTriggerNotFound
	}

	// 保留统计信息
	trigger.CreateTime = existing.CreateTime
	trigger.TriggerCount = existing.TriggerCount
	trigger.ErrorCount = existing.ErrorCount
	trigger.UpdateTime = time.Now()

	// 更新索引
	tm.updateIndex(existing, trigger)

	tm.triggers[trigger.ID] = trigger

	return nil
}

// DeleteTrigger 删除触发器
func (tm *TriggerManager) DeleteTrigger(triggerID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	trigger, exists := tm.triggers[triggerID]
	if !exists {
		return ErrTriggerNotFound
	}

	// 删除索引
	for _, pointID := range trigger.PointIDs {
		slice := tm.byPoint[pointID]
		tm.removeFromSlice(&slice, triggerID)
		tm.byPoint[pointID] = slice
	}
	slice := tm.byType[trigger.Type]
	tm.removeFromSlice(&slice, triggerID)
	tm.byType[trigger.Type] = slice

	delete(tm.triggers, triggerID)

	tm.logger.Info("Trigger deleted", zap.String("triggerID", triggerID))

	return nil
}

// GetTrigger 获取触发器
func (tm *TriggerManager) GetTrigger(triggerID string) (*Trigger, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	trigger, exists := tm.triggers[triggerID]
	if !exists {
		return nil, ErrTriggerNotFound
	}

	return trigger, nil
}

// GetTriggersByPoint 按计算点获取触发器
func (tm *TriggerManager) GetTriggersByPoint(pointID string) []*Trigger {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	triggerIDs := tm.byPoint[pointID]
	triggers := make([]*Trigger, 0, len(triggerIDs))

	for _, id := range triggerIDs {
		if trigger, exists := tm.triggers[id]; exists {
			triggers = append(triggers, trigger)
		}
	}

	// 按优先级排序
	sort.Slice(triggers, func(i, j int) bool {
		return triggers[i].Priority > triggers[j].Priority
	})

	return triggers
}

// GetTriggersByType 按类型获取触发器
func (tm *TriggerManager) GetTriggersByType(triggerType TriggerType) []*Trigger {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	triggerIDs := tm.byType[triggerType]
	triggers := make([]*Trigger, 0, len(triggerIDs))

	for _, id := range triggerIDs {
		if trigger, exists := tm.triggers[id]; exists {
			triggers = append(triggers, trigger)
		}
	}

	return triggers
}

// OnDataChange 数据变化触发
func (tm *TriggerManager) OnDataChange(ctx context.Context, pointID string, oldValue, newValue float64) error {
	event := &TriggerEvent{
		ID:        fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		Type:      TriggerTypeDataChange,
		Source:    "data-change",
		PointID:   pointID,
		OldValue:  oldValue,
		NewValue:  newValue,
		Timestamp: time.Now(),
	}

	select {
	case tm.eventQueue <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return errors.New("event queue is full")
	}
}

// OnEvent 事件触发
func (tm *TriggerManager) OnEvent(ctx context.Context, eventType string, payload map[string]interface{}) error {
	event := &TriggerEvent{
		ID:        fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		Type:      TriggerTypeEvent,
		Source:    "event",
		Timestamp: time.Now(),
		Payload:   payload,
		Metadata:  map[string]string{"eventType": eventType},
	}

	select {
	case tm.eventQueue <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return errors.New("event queue is full")
	}
}

// OnCondition 条件触发
func (tm *TriggerManager) OnCondition(ctx context.Context, pointID string, condition string, value float64) error {
	event := &TriggerEvent{
		ID:        fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		Type:      TriggerTypeCondition,
		Source:    "condition",
		PointID:   pointID,
		NewValue:  value,
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"condition": condition},
	}

	select {
	case tm.eventQueue <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return errors.New("event queue is full")
	}
}

// processEvents 处理事件
func (tm *TriggerManager) processEvents() {
	defer tm.wg.Done()

	for {
		select {
		case event := <-tm.eventQueue:
			tm.handleEvent(event)

		case <-tm.ctx.Done():
			return
		}
	}
}

// handleEvent 处理事件
func (tm *TriggerManager) handleEvent(event *TriggerEvent) {
	startTime := time.Now()

	// 获取相关触发器
	var triggers []*Trigger
	if event.PointID != "" {
		triggers = tm.GetTriggersByPoint(event.PointID)
	} else {
		triggers = tm.GetTriggersByType(event.Type)
	}

	// 执行触发器
	for _, trigger := range triggers {
		if !trigger.Enabled {
			continue
		}

		// 检查冷却时间
		if !trigger.LastTrigger.IsZero() && trigger.Condition != nil {
			if time.Since(trigger.LastTrigger) < trigger.Condition.CooldownTime {
				continue
			}
		}

		// 评估触发条件
		result := tm.evaluateTrigger(trigger, event)

		// 更新指标
		triggerTotal.WithLabelValues(string(trigger.Type), fmt.Sprintf("%v", result.Triggered)).Inc()

		if result.Triggered {
			// 执行动作
			tm.executeActions(trigger, result)

			// 更新触发器状态
			tm.mu.Lock()
			if t, exists := tm.triggers[trigger.ID]; exists {
				t.LastTrigger = time.Now()
				t.TriggerCount++
				if !result.Executed {
					t.ErrorCount++
					t.LastError = result.Error
				}
			}
			tm.mu.Unlock()

			// 执行触发器链
			if trigger.ChainConfig != nil {
				tm.executeChain(trigger, event, result)
			}
		}
	}

	triggerLatency.WithLabelValues(string(event.Type)).Observe(time.Since(startTime).Seconds())
}

// evaluateTrigger 评估触发器
func (tm *TriggerManager) evaluateTrigger(trigger *Trigger, event *TriggerEvent) *TriggerResult {
	result := &TriggerResult{
		TriggerID: trigger.ID,
		StartTime: time.Now(),
		Context:   make(map[string]interface{}),
	}

	defer func() {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
	}()

	// 根据触发器类型评估
	switch trigger.Type {
	case TriggerTypeDataChange:
		result.Triggered = tm.evaluateDataChange(trigger, event)

	case TriggerTypeEvent:
		result.Triggered = tm.evaluateEvent(trigger, event)

	case TriggerTypeCondition:
		result.Triggered = tm.evaluateCondition(trigger, event)

	case TriggerTypeComposite:
		result.Triggered = tm.evaluateComposite(trigger, event)

	default:
		result.Error = fmt.Sprintf("unknown trigger type: %s", trigger.Type)
	}

	return result
}

// evaluateDataChange 评估数据变化触发
func (tm *TriggerManager) evaluateDataChange(trigger *Trigger, event *TriggerEvent) bool {
	if trigger.Condition == nil {
		return false
	}

	cond := trigger.Condition
	change := event.NewValue - event.OldValue

	// 检查变化方向
	if cond.Direction == "up" && change <= 0 {
		return false
	}
	if cond.Direction == "down" && change >= 0 {
		return false
	}

	// 检查变化阈值
	if cond.ChangeThreshold != nil {
		absChange := change
		if absChange < 0 {
			absChange = -absChange
		}
		if absChange < *cond.ChangeThreshold {
			return false
		}
	}

	// 检查变化百分比
	if cond.ChangePercent != nil && event.OldValue != 0 {
		percentChange := (change / event.OldValue) * 100
		if percentChange < 0 {
			percentChange = -percentChange
		}
		if percentChange < *cond.ChangePercent {
			return false
		}
	}

	return true
}

// evaluateEvent 评估事件触发
func (tm *TriggerManager) evaluateEvent(trigger *Trigger, event *TriggerEvent) bool {
	if trigger.Condition == nil {
		return true
	}

	cond := trigger.Condition

	// 检查事件类型
	if cond.EventType != "" && event.Metadata["eventType"] != cond.EventType {
		return false
	}

	// 检查事件数据
	if cond.EventData != "" {
		// 简单匹配，实际可以使用正则表达式
		if data, ok := event.Payload["data"].(string); !ok || data != cond.EventData {
			return false
		}
	}

	return true
}

// evaluateCondition 评估条件触发
func (tm *TriggerManager) evaluateCondition(trigger *Trigger, event *TriggerEvent) bool {
	if trigger.Condition == nil || trigger.Condition.Expression == "" {
		return false
	}

	// 简化实现：解析简单的条件表达式
	// 实际项目中可以使用表达式引擎
	_ = trigger.Condition.Expression

	// 支持简单的比较表达式: value > 100, value < 50 等
	// 这里简化处理
	return event.NewValue > 0 // 示例
}

// evaluateComposite 评估复合触发
func (tm *TriggerManager) evaluateComposite(trigger *Trigger, event *TriggerEvent) bool {
	if trigger.Condition == nil || len(trigger.Condition.SubConditions) == 0 {
		return false
	}

	cond := trigger.Condition

	switch cond.LogicOperator {
	case "and":
		for _, subCond := range cond.SubConditions {
			subTrigger := &Trigger{
				Type:      trigger.Type,
				Condition: subCond,
			}
			if !tm.evaluateDataChange(subTrigger, event) {
				return false
			}
		}
		return true

	case "or":
		for _, subCond := range cond.SubConditions {
			subTrigger := &Trigger{
				Type:      trigger.Type,
				Condition: subCond,
			}
			if tm.evaluateDataChange(subTrigger, event) {
				return true
			}
		}
		return false

	default:
		return false
	}
}

// executeActions 执行动作
func (tm *TriggerManager) executeActions(trigger *Trigger, result *TriggerResult) {
	result.Actions = make([]ActionResult, 0, len(trigger.Actions))

	for _, action := range trigger.Actions {
		actionResult := tm.executeAction(trigger, action, result)
		result.Actions = append(result.Actions, actionResult)

		if !actionResult.Success && !action.Async {
			result.Error = actionResult.Error
			result.Executed = false
			return
		}
	}

	result.Executed = true
}

// executeAction 执行单个动作
func (tm *TriggerManager) executeAction(trigger *Trigger, action TriggerAction, result *TriggerResult) ActionResult {
	actionResult := ActionResult{
		Type:   action.Type,
		Target: action.Target,
	}

	startTime := time.Now()
	defer func() {
		actionResult.Duration = time.Since(startTime)
	}()

	ctx, cancel := context.WithTimeout(tm.ctx, action.Timeout)
	defer cancel()

	switch action.Type {
	case "compute":
		// 执行计算
		pointIDs := []string{action.Target}
		if ids, ok := action.Params["pointIds"].([]string); ok {
			pointIDs = ids
		}

		_, err := tm.executor.Execute(ctx, pointIDs)
		actionResult.Success = err == nil
		if err != nil {
			actionResult.Error = err.Error()
		}

	case "notify":
		// 发送通知（简化实现）
		actionResult.Success = true

	case "script":
		// 执行脚本（简化实现）
		actionResult.Success = true

	default:
		actionResult.Error = fmt.Sprintf("unknown action type: %s", action.Type)
	}

	return actionResult
}

// executeChain 执行触发器链
func (tm *TriggerManager) executeChain(trigger *Trigger, event *TriggerEvent, result *TriggerResult) {
	if trigger.ChainConfig == nil {
		return
	}

	chainConfig := trigger.ChainConfig

	// 检查是否需要停止
	if chainConfig.StopOnSuccess && result.Executed {
		return
	}
	if chainConfig.StopOnFailure && !result.Executed {
		return
	}

	// 执行下一个触发器
	for _, nextTriggerID := range chainConfig.NextTriggers {
		nextTrigger, err := tm.GetTrigger(nextTriggerID)
		if err != nil {
			continue
		}

		// 递归执行
		nextResult := tm.evaluateTrigger(nextTrigger, event)
		if nextResult.Triggered {
			tm.executeActions(nextTrigger, nextResult)
		}
	}
}

// CreateChain 创建触发器链
func (tm *TriggerManager) CreateChain(chain *TriggerChain) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if chain.ID == "" {
		return errors.New("chain ID is required")
	}

	if _, exists := tm.chains[chain.ID]; exists {
		return errors.New("chain already exists")
	}

	chain.Status = "active"
	tm.chains[chain.ID] = chain

	return nil
}

// GetChain 获取触发器链
func (tm *TriggerManager) GetChain(chainID string) (*TriggerChain, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	chain, exists := tm.chains[chainID]
	if !exists {
		return nil, errors.New("chain not found")
	}

	return chain, nil
}

// updateIndex 更新索引
func (tm *TriggerManager) updateIndex(oldTrigger, newTrigger *Trigger) {
	// 删除旧索引
	for _, pointID := range oldTrigger.PointIDs {
		slice := tm.byPoint[pointID]
		tm.removeFromSlice(&slice, oldTrigger.ID)
		tm.byPoint[pointID] = slice
	}
	slice := tm.byType[oldTrigger.Type]
	tm.removeFromSlice(&slice, oldTrigger.ID)
	tm.byType[oldTrigger.Type] = slice

	// 添加新索引
	for _, pointID := range newTrigger.PointIDs {
		tm.byPoint[pointID] = append(tm.byPoint[pointID], newTrigger.ID)
	}
	tm.byType[newTrigger.Type] = append(tm.byType[newTrigger.Type], newTrigger.ID)
}

// removeFromSlice 从切片中移除元素
func (tm *TriggerManager) removeFromSlice(slice *[]string, item string) {
	for i, v := range *slice {
		if v == item {
			*slice = append((*slice)[:i], (*slice)[i+1:]...)
			break
		}
	}
}

// collectMetrics 收集指标
func (tm *TriggerManager) collectMetrics() {
	defer tm.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tm.mu.RLock()
			count := int64(len(tm.triggers))
			tm.mu.RUnlock()

			triggerActive.Set(float64(count))

		case <-tm.ctx.Done():
			return
		}
	}
}

// EnableTrigger 启用触发器
func (tm *TriggerManager) EnableTrigger(triggerID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	trigger, exists := tm.triggers[triggerID]
	if !exists {
		return ErrTriggerNotFound
	}

	trigger.Enabled = true
	trigger.Status = TriggerStatusActive
	trigger.UpdateTime = time.Now()

	return nil
}

// DisableTrigger 禁用触发器
func (tm *TriggerManager) DisableTrigger(triggerID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	trigger, exists := tm.triggers[triggerID]
	if !exists {
		return ErrTriggerNotFound
	}

	trigger.Enabled = false
	trigger.Status = TriggerStatusDisabled
	trigger.UpdateTime = time.Now()

	return nil
}

// GetStats 获取统计信息
func (tm *TriggerManager) GetStats() map[string]interface{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	stats := map[string]interface{}{
		"totalTriggers": len(tm.triggers),
		"totalChains":   len(tm.chains),
		"byType":        make(map[TriggerType]int),
		"byStatus":      make(map[TriggerStatus]int),
	}

	for _, trigger := range tm.triggers {
		stats["byType"].(map[TriggerType]int)[trigger.Type]++
		stats["byStatus"].(map[TriggerStatus]int)[trigger.Status]++
	}

	return stats
}
