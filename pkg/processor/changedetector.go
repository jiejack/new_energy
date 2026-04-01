package processor

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// ChangeType 变位类型
type ChangeType int

const (
	ChangeTypeNone ChangeType = iota // 无变化
	ChangeTypeRising                 // 上升变位
	ChangeTypeFalling                // 下降变位
	ChangeTypeToggle                 // 翻转变位
)

// ChangeEvent 变位事件
type ChangeEvent struct {
	PointID    string      `json:"point_id"`    // 测点ID
	Type       ChangeType  `json:"type"`        // 变位类型
	OldValue   float64     `json:"old_value"`   // 旧值
	NewValue   float64     `json:"new_value"`   // 新值
	Timestamp  time.Time   `json:"timestamp"`   // 时间戳
	Quality    QualityCode `json:"quality"`     // 质量码
	Debounced  bool        `json:"debounced"`   // 是否经过防抖
	Duration   time.Duration `json:"duration"`  // 持续时间
}

// ChangeResult 变位检测结果
type ChangeResult struct {
	Changed    bool          `json:"changed"`    // 是否发生变位
	Event      *ChangeEvent  `json:"event"`      // 变位事件
	Value      float64       `json:"value"`      // 当前值
	Timestamp  time.Time     `json:"timestamp"`  // 时间戳
	Quality    QualityCode   `json:"quality"`    // 质量码
}

// ChangeDetector 变位检测器接口
type ChangeDetector interface {
	// Detect 检测变位
	Detect(value float64) ChangeResult
	// Reset 重置检测器状态
	Reset()
	// Name 获取检测器名称
	Name() string
}

// BasicChangeDetector 基础变位检测器
type BasicChangeDetector struct {
	name      string
	lastValue float64
	lastTime  time.Time
	initialized bool
	mu        sync.RWMutex
	logger    *zap.Logger
}

// BasicChangeDetectorConfig 基础变位检测器配置
type BasicChangeDetectorConfig struct {
	Name   string
	Logger *zap.Logger
}

// NewBasicChangeDetector 创建基础变位检测器
func NewBasicChangeDetector(config BasicChangeDetectorConfig) *BasicChangeDetector {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	return &BasicChangeDetector{
		name:   config.Name,
		logger: config.Logger,
	}
}

// Detect 检测变位
func (d *BasicChangeDetector) Detect(value float64) ChangeResult {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	now := time.Now()
	result := ChangeResult{
		Value:     value,
		Timestamp: now,
		Quality:   QualityGood,
	}
	
	if !d.initialized {
		d.lastValue = value
		d.lastTime = now
		d.initialized = true
		return result
	}
	
	// 检测变位
	if value != d.lastValue {
		changeType := ChangeTypeNone
		if value > d.lastValue {
			changeType = ChangeTypeRising
		} else if value < d.lastValue {
			changeType = ChangeTypeFalling
		}
		
		result.Changed = true
		result.Event = &ChangeEvent{
			Type:      changeType,
			OldValue:  d.lastValue,
			NewValue:  value,
			Timestamp: now,
			Quality:   QualityGood,
			Duration:  now.Sub(d.lastTime),
		}
		
		d.logger.Debug("change detected",
			zap.String("type", changeTypeString(changeType)),
			zap.Float64("old", d.lastValue),
			zap.Float64("new", value),
		)
	}
	
	d.lastValue = value
	d.lastTime = now
	
	return result
}

// Reset 重置检测器状态
func (d *BasicChangeDetector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.lastValue = 0
	d.lastTime = time.Time{}
	d.initialized = false
}

// Name 获取检测器名称
func (d *BasicChangeDetector) Name() string {
	return d.name
}

// DeadbandChangeDetector 死区变位检测器
type DeadbandChangeDetector struct {
	name        string
	deadband    float64     // 死区范围
	lastValue   float64     // 上次有效值
	lastTime    time.Time   // 上次时间
	initialized bool
	mu          sync.RWMutex
	logger      *zap.Logger
}

// DeadbandChangeDetectorConfig 死区变位检测器配置
type DeadbandChangeDetectorConfig struct {
	Name     string
	Deadband float64
	Logger   *zap.Logger
}

// NewDeadbandChangeDetector 创建死区变位检测器
func NewDeadbandChangeDetector(config DeadbandChangeDetectorConfig) *DeadbandChangeDetector {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	if config.Deadband < 0 {
		config.Deadband = 0
	}
	return &DeadbandChangeDetector{
		name:     config.Name,
		deadband: config.Deadband,
		logger:   config.Logger,
	}
}

// Detect 检测变位
func (d *DeadbandChangeDetector) Detect(value float64) ChangeResult {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	now := time.Now()
	result := ChangeResult{
		Value:     value,
		Timestamp: now,
		Quality:   QualityGood,
	}
	
	if !d.initialized {
		d.lastValue = value
		d.lastTime = now
		d.initialized = true
		return result
	}
	
	// 计算变化量
	change := value - d.lastValue
	
	// 检查是否超过死区
	if abs(change) > d.deadband {
		changeType := ChangeTypeNone
		if change > 0 {
			changeType = ChangeTypeRising
		} else {
			changeType = ChangeTypeFalling
		}
		
		result.Changed = true
		result.Event = &ChangeEvent{
			Type:      changeType,
			OldValue:  d.lastValue,
			NewValue:  value,
			Timestamp: now,
			Quality:   QualityGood,
			Duration:  now.Sub(d.lastTime),
		}
		
		d.lastValue = value
		d.lastTime = now
		
		d.logger.Debug("deadband change detected",
			zap.Float64("deadband", d.deadband),
			zap.Float64("change", change),
			zap.Float64("old", d.lastValue),
			zap.Float64("new", value),
		)
	}
	
	return result
}

// Reset 重置检测器状态
func (d *DeadbandChangeDetector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.lastValue = 0
	d.lastTime = time.Time{}
	d.initialized = false
}

// Name 获取检测器名称
func (d *DeadbandChangeDetector) Name() string {
	return d.name
}

// SetDeadband 设置死区
func (d *DeadbandChangeDetector) SetDeadband(deadband float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.deadband = deadband
}

// Debouncer 防抖处理器
type Debouncer struct {
	name         string
	debounceTime time.Duration // 防抖时间
	pendingValue float64       // 待确认值
	pendingTime  time.Time     // 待确认时间
	stableValue  float64       // 已稳定值
	stableTime   time.Time     // 稳定时间
	initialized  bool
	mu           sync.RWMutex
	logger       *zap.Logger
}

// DebouncerConfig 防抖处理器配置
type DebouncerConfig struct {
	Name         string
	DebounceTime time.Duration
	Logger       *zap.Logger
}

// NewDebouncer 创建防抖处理器
func NewDebouncer(config DebouncerConfig) *Debouncer {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	if config.DebounceTime <= 0 {
		config.DebounceTime = 100 * time.Millisecond
	}
	return &Debouncer{
		name:         config.Name,
		debounceTime: config.DebounceTime,
		logger:       config.Logger,
	}
}

// Process 处理数据
func (d *Debouncer) Process(value float64) ChangeResult {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	now := time.Now()
	result := ChangeResult{
		Value:     value,
		Timestamp: now,
		Quality:   QualityGood,
	}
	
	if !d.initialized {
		d.stableValue = value
		d.stableTime = now
		d.pendingValue = value
		d.pendingTime = now
		d.initialized = true
		return result
	}
	
	// 检查值是否变化
	if value != d.pendingValue {
		// 值发生变化，更新待确认值
		d.pendingValue = value
		d.pendingTime = now
		
		d.logger.Debug("debounce: value changed",
			zap.Float64("stable", d.stableValue),
			zap.Float64("pending", value),
		)
	}
	
	// 检查是否达到稳定时间
	if now.Sub(d.pendingTime) >= d.debounceTime {
		// 检查是否与稳定值不同
		if d.pendingValue != d.stableValue {
			changeType := ChangeTypeNone
			if d.pendingValue > d.stableValue {
				changeType = ChangeTypeRising
			} else {
				changeType = ChangeTypeFalling
			}
			
			result.Changed = true
			result.Event = &ChangeEvent{
				Type:      changeType,
				OldValue:  d.stableValue,
				NewValue:  d.pendingValue,
				Timestamp: now,
				Quality:   QualityGood | QualityReasonDebounced,
				Debounced: true,
				Duration:  now.Sub(d.stableTime),
			}
			
			d.stableValue = d.pendingValue
			d.stableTime = now
			
			d.logger.Debug("debounce: value stabilized",
				zap.Float64("value", d.stableValue),
				zap.Duration("debounce_time", d.debounceTime),
			)
		}
	}
	
	// 返回稳定值
	result.Value = d.stableValue
	
	return result
}

// Reset 重置防抖器状态
func (d *Debouncer) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.pendingValue = 0
	d.pendingTime = time.Time{}
	d.stableValue = 0
	d.stableTime = time.Time{}
	d.initialized = false
}

// Name 获取防抖器名称
func (d *Debouncer) Name() string {
	return d.name
}

// SetDebounceTime 设置防抖时间
func (d *Debouncer) SetDebounceTime(debounceTime time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.debounceTime = debounceTime
}

// GetStableValue 获取稳定值
func (d *Debouncer) GetStableValue() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.stableValue
}

// SOEEvent SOE事件记录
type SOEEvent struct {
	ID         string      `json:"id"`          // 事件ID
	PointID    string      `json:"point_id"`    // 测点ID
	PointName  string      `json:"point_name"`  // 测点名称
	Type       ChangeType  `json:"type"`        // 变位类型
	OldValue   float64     `json:"old_value"`   // 旧值
	NewValue   float64     `json:"new_value"`   // 新值
	Timestamp  time.Time   `json:"timestamp"`   // 时间戳（毫秒精度）
	Quality    QualityCode `json:"quality"`     // 质量码
	Debounced  bool        `json:"debounced"`   // 是否经过防抖
	Priority   int         `json:"priority"`    // 优先级
	Acknowledged bool      `json:"acknowledged"` // 是否已确认
	AckTime    time.Time   `json:"ack_time"`    // 确认时间
	AckUser    string      `json:"ack_user"`    // 确认用户
}

// SOERecorder SOE事件记录器
type SOERecorder struct {
	name       string
	maxEvents  int                // 最大事件数
	events     []SOEEvent         // 事件列表
	eventChan  chan SOEEvent      // 事件通道
	mu         sync.RWMutex
	logger     *zap.Logger
	stopChan   chan struct{}
}

// SOERecorderConfig SOE记录器配置
type SOERecorderConfig struct {
	Name      string
	MaxEvents int
	Logger    *zap.Logger
}

// NewSOERecorder 创建SOE记录器
func NewSOERecorder(config SOERecorderConfig) *SOERecorder {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	if config.MaxEvents <= 0 {
		config.MaxEvents = 10000
	}
	
	recorder := &SOERecorder{
		name:      config.Name,
		maxEvents: config.MaxEvents,
		events:    make([]SOEEvent, 0, config.MaxEvents),
		eventChan: make(chan SOEEvent, 1000),
		stopChan:  make(chan struct{}),
		logger:    config.Logger,
	}
	
	// 启动事件处理协程
	go recorder.processEvents()
	
	return recorder
}

// Record 记录事件
func (r *SOERecorder) Record(event ChangeEvent, pointID, pointName string, priority int) {
	soeEvent := SOEEvent{
		ID:         generateEventID(),
		PointID:    pointID,
		PointName:  pointName,
		Type:       event.Type,
		OldValue:   event.OldValue,
		NewValue:   event.NewValue,
		Timestamp:  event.Timestamp,
		Quality:    event.Quality,
		Debounced:  event.Debounced,
		Priority:   priority,
	}
	
	select {
	case r.eventChan <- soeEvent:
	default:
		r.logger.Warn("SOE event channel full, event dropped",
			zap.String("point_id", pointID),
		)
	}
}

// processEvents 处理事件
func (r *SOERecorder) processEvents() {
	for {
		select {
		case event := <-r.eventChan:
			r.mu.Lock()
			r.events = append(r.events, event)
			
			// 限制事件数量
			if len(r.events) > r.maxEvents {
				r.events = r.events[len(r.events)-r.maxEvents:]
			}
			r.mu.Unlock()
			
			r.logger.Debug("SOE event recorded",
				zap.String("id", event.ID),
				zap.String("point_id", event.PointID),
				zap.String("type", changeTypeString(event.Type)),
			)
			
		case <-r.stopChan:
			return
		}
	}
}

// GetEvents 获取事件列表
func (r *SOERecorder) GetEvents(limit int) []SOEEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if limit <= 0 || limit > len(r.events) {
		limit = len(r.events)
	}
	
	result := make([]SOEEvent, limit)
	copy(result, r.events[len(r.events)-limit:])
	return result
}

// GetEventsByPoint 获取指定测点的事件
func (r *SOERecorder) GetEventsByPoint(pointID string, limit int) []SOEEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []SOEEvent
	for i := len(r.events) - 1; i >= 0 && len(result) < limit; i-- {
		if r.events[i].PointID == pointID {
			result = append(result, r.events[i])
		}
	}
	
	return result
}

// GetEventsByTimeRange 获取指定时间范围的事件
func (r *SOERecorder) GetEventsByTimeRange(start, end time.Time) []SOEEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []SOEEvent
	for _, event := range r.events {
		if event.Timestamp.After(start) && event.Timestamp.Before(end) {
			result = append(result, event)
		}
	}
	
	return result
}

// Acknowledge 确认事件
func (r *SOERecorder) Acknowledge(eventID, user string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for i := range r.events {
		if r.events[i].ID == eventID {
			r.events[i].Acknowledged = true
			r.events[i].AckTime = time.Now()
			r.events[i].AckUser = user
			return nil
		}
	}
	
	return nil
}

// Stop 停止记录器
func (r *SOERecorder) Stop() {
	close(r.stopChan)
}

// Name 获取记录器名称
func (r *SOERecorder) Name() string {
	return r.name
}

// ChangeDetectorWithDebounce 带防抖的变位检测器
type ChangeDetectorWithDebounce struct {
	name       string
	detector   ChangeDetector
	debouncer  *Debouncer
	mu         sync.RWMutex
	logger     *zap.Logger
}

// ChangeDetectorWithDebounceConfig 配置
type ChangeDetectorWithDebounceConfig struct {
	Name         string
	Detector     ChangeDetector
	DebounceTime time.Duration
	Logger       *zap.Logger
}

// NewChangeDetectorWithDebounce 创建带防抖的变位检测器
func NewChangeDetectorWithDebounce(config ChangeDetectorWithDebounceConfig) *ChangeDetectorWithDebounce {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	
	debouncer := NewDebouncer(DebouncerConfig{
		Name:         config.Name + "_debouncer",
		DebounceTime: config.DebounceTime,
		Logger:       config.Logger,
	})
	
	return &ChangeDetectorWithDebounce{
		name:      config.Name,
		detector:  config.Detector,
		debouncer: debouncer,
		logger:    config.Logger,
	}
}

// Detect 检测变位
func (d *ChangeDetectorWithDebounce) Detect(value float64) ChangeResult {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// 先进行防抖处理
	debounceResult := d.debouncer.Process(value)
	
	// 再进行变位检测
	detectResult := d.detector.Detect(debounceResult.Value)
	
	// 合并结果
	if detectResult.Changed {
		detectResult.Event.Debounced = true
		detectResult.Quality |= QualityReasonDebounced
	}
	
	return detectResult
}

// Reset 重置检测器状态
func (d *ChangeDetectorWithDebounce) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.detector.Reset()
	d.debouncer.Reset()
}

// Name 获取检测器名称
func (d *ChangeDetectorWithDebounce) Name() string {
	return d.name
}

// BinaryChangeDetector 双值变位检测器（用于开关量）
type BinaryChangeDetector struct {
	name        string
	lastState   bool // 上次状态
	lastTime    time.Time
	initialized bool
	mu          sync.RWMutex
	logger      *zap.Logger
}

// BinaryChangeDetectorConfig 配置
type BinaryChangeDetectorConfig struct {
	Name   string
	Logger *zap.Logger
}

// NewBinaryChangeDetector 创建双值变位检测器
func NewBinaryChangeDetector(config BinaryChangeDetectorConfig) *BinaryChangeDetector {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	return &BinaryChangeDetector{
		name:   config.Name,
		logger: config.Logger,
	}
}

// Detect 检测变位
func (d *BinaryChangeDetector) Detect(value bool) ChangeResult {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	now := time.Now()
	result := ChangeResult{
		Timestamp: now,
		Quality:   QualityGood,
	}
	
	// 转换为float64
	var floatValue float64
	if value {
		floatValue = 1.0
	} else {
		floatValue = 0.0
	}
	result.Value = floatValue
	
	if !d.initialized {
		d.lastState = value
		d.lastTime = now
		d.initialized = true
		return result
	}
	
	// 检测变位
	if value != d.lastState {
		changeType := ChangeTypeToggle
		if value {
			changeType = ChangeTypeRising
		} else {
			changeType = ChangeTypeFalling
		}
		
		var oldFloat float64
		if d.lastState {
			oldFloat = 1.0
		} else {
			oldFloat = 0.0
		}
		
		result.Changed = true
		result.Event = &ChangeEvent{
			Type:      changeType,
			OldValue:  oldFloat,
			NewValue:  floatValue,
			Timestamp: now,
			Quality:   QualityGood,
			Duration:  now.Sub(d.lastTime),
		}
		
		d.logger.Debug("binary change detected",
			zap.Bool("old", d.lastState),
			zap.Bool("new", value),
		)
	}
	
	d.lastState = value
	d.lastTime = now
	
	return result
}

// Reset 重置检测器状态
func (d *BinaryChangeDetector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.lastState = false
	d.lastTime = time.Time{}
	d.initialized = false
}

// Name 获取检测器名称
func (d *BinaryChangeDetector) Name() string {
	return d.name
}

// ChangeDetectorFactory 变位检测器工厂
type ChangeDetectorFactory struct {
	logger *zap.Logger
}

// NewChangeDetectorFactory 创建变位检测器工厂
func NewChangeDetectorFactory(logger *zap.Logger) *ChangeDetectorFactory {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ChangeDetectorFactory{logger: logger}
}

// CreateBasicChangeDetector 创建基础变位检测器
func (f *ChangeDetectorFactory) CreateBasicChangeDetector(name string) *BasicChangeDetector {
	return NewBasicChangeDetector(BasicChangeDetectorConfig{
		Name:   name,
		Logger: f.logger,
	})
}

// CreateDeadbandChangeDetector 创建死区变位检测器
func (f *ChangeDetectorFactory) CreateDeadbandChangeDetector(name string, deadband float64) *DeadbandChangeDetector {
	return NewDeadbandChangeDetector(DeadbandChangeDetectorConfig{
		Name:     name,
		Deadband: deadband,
		Logger:   f.logger,
	})
}

// CreateDebouncer 创建防抖处理器
func (f *ChangeDetectorFactory) CreateDebouncer(name string, debounceTime time.Duration) *Debouncer {
	return NewDebouncer(DebouncerConfig{
		Name:         name,
		DebounceTime: debounceTime,
		Logger:       f.logger,
	})
}

// CreateSOERecorder 创建SOE记录器
func (f *ChangeDetectorFactory) CreateSOERecorder(name string, maxEvents int) *SOERecorder {
	return NewSOERecorder(SOERecorderConfig{
		Name:      name,
		MaxEvents: maxEvents,
		Logger:    f.logger,
	})
}

// CreateBinaryChangeDetector 创建双值变位检测器
func (f *ChangeDetectorFactory) CreateBinaryChangeDetector(name string) *BinaryChangeDetector {
	return NewBinaryChangeDetector(BinaryChangeDetectorConfig{
		Name:   name,
		Logger: f.logger,
	})
}

// 辅助函数

// abs 绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// changeTypeString 变位类型字符串
func changeTypeString(t ChangeType) string {
	switch t {
	case ChangeTypeRising:
		return "rising"
	case ChangeTypeFalling:
		return "falling"
	case ChangeTypeToggle:
		return "toggle"
	default:
		return "none"
	}
}

// generateEventID 生成事件ID
func generateEventID() string {
	return time.Now().Format("20060102150405.000")
}
