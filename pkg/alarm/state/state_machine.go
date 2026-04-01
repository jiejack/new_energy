package state

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
)

// AlertState 告警状态枚举
type AlertState int

const (
	StateActive AlertState = iota
	StateAcknowledged
	StateCleared
	StateSuppressed
)

func (s AlertState) String() string {
	switch s {
	case StateActive:
		return "Active"
	case StateAcknowledged:
		return "Acknowledged"
	case StateCleared:
		return "Cleared"
	case StateSuppressed:
		return "Suppressed"
	default:
		return "Unknown"
	}
}

// ToEntityStatus 转换为实体状态
func (s AlertState) ToEntityStatus() entity.AlarmStatus {
	switch s {
	case StateActive:
		return entity.AlarmStatusActive
	case StateAcknowledged:
		return entity.AlarmStatusAcknowledged
	case StateCleared:
		return entity.AlarmStatusCleared
	case StateSuppressed:
		return entity.AlarmStatusSuppressed
	default:
		return entity.AlarmStatusActive
	}
}

// StateFromEntityStatus 从实体状态转换
func StateFromEntityStatus(status entity.AlarmStatus) AlertState {
	switch status {
	case entity.AlarmStatusActive:
		return StateActive
	case entity.AlarmStatusAcknowledged:
		return StateAcknowledged
	case entity.AlarmStatusCleared:
		return StateCleared
	case entity.AlarmStatusSuppressed:
		return StateSuppressed
	default:
		return StateActive
	}
}

// StateTransition 状态转换
type StateTransition int

const (
	TransitionNone StateTransition = iota
	TransitionTrigger
	TransitionAcknowledge
	TransitionClear
	TransitionSuppress
	TransitionUnsuppress
	TransitionReactivate
)

func (t StateTransition) String() string {
	switch t {
	case TransitionNone:
		return "None"
	case TransitionTrigger:
		return "Trigger"
	case TransitionAcknowledge:
		return "Acknowledge"
	case TransitionClear:
		return "Clear"
	case TransitionSuppress:
		return "Suppress"
	case TransitionUnsuppress:
		return "Unsuppress"
	case TransitionReactivate:
		return "Reactivate"
	default:
		return "Unknown"
	}
}

// StateChangeEvent 状态变更事件
type StateChangeEvent struct {
	AlarmID      string          `json:"alarm_id"`
	FromState    AlertState      `json:"from_state"`
	ToState      AlertState      `json:"to_state"`
	Transition   StateTransition `json:"transition"`
	Timestamp    time.Time       `json:"timestamp"`
	Operator     string          `json:"operator,omitempty"`
	Reason       string          `json:"reason,omitempty"`
	Metadata     map[string]any  `json:"metadata,omitempty"`
}

// StateChangeHandler 状态变更处理器
type StateChangeHandler func(ctx context.Context, event StateChangeEvent) error

// TransitionRule 状态转换规则
type TransitionRule struct {
	FromState     AlertState
	ToState       AlertState
	Transition    StateTransition
	Allowed       bool
	RequireReason bool
}

// StateMachine 状态机
type StateMachine struct {
	mu              sync.RWMutex
	rules           map[AlertState]map[StateTransition]TransitionRule
	handlers        []StateChangeHandler
	eventChan       chan StateChangeEvent
	workerCount     int
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// StateMachineConfig 状态机配置
type StateMachineConfig struct {
	WorkerCount   int
	BufferSize    int
	EventHandlers []StateChangeHandler
}

// NewStateMachine 创建状态机
func NewStateMachine(cfg StateMachineConfig) *StateMachine {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 4
	}
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 1000
	}

	sm := &StateMachine{
		rules:       make(map[AlertState]map[StateTransition]TransitionRule),
		handlers:    cfg.EventHandlers,
		eventChan:   make(chan StateChangeEvent, cfg.BufferSize),
		workerCount: cfg.WorkerCount,
		stopChan:    make(chan struct{}),
	}

	sm.initDefaultRules()
	return sm
}

// initDefaultRules 初始化默认转换规则
func (sm *StateMachine) initDefaultRules() {
	// Active 状态的转换规则
	sm.addRule(TransitionRule{
		FromState:  StateActive,
		ToState:    StateAcknowledged,
		Transition: TransitionAcknowledge,
		Allowed:    true,
	})
	sm.addRule(TransitionRule{
		FromState:  StateActive,
		ToState:    StateCleared,
		Transition: TransitionClear,
		Allowed:    true,
	})
	sm.addRule(TransitionRule{
		FromState:  StateActive,
		ToState:    StateSuppressed,
		Transition: TransitionSuppress,
		Allowed:    true,
	})

	// Acknowledged 状态的转换规则
	sm.addRule(TransitionRule{
		FromState:  StateAcknowledged,
		ToState:    StateCleared,
		Transition: TransitionClear,
		Allowed:    true,
	})
	sm.addRule(TransitionRule{
		FromState:  StateAcknowledged,
		ToState:    StateSuppressed,
		Transition: TransitionSuppress,
		Allowed:    true,
	})
	sm.addRule(TransitionRule{
		FromState:  StateAcknowledged,
		ToState:    StateActive,
		Transition: TransitionReactivate,
		Allowed:    true,
	})

	// Suppressed 状态的转换规则
	sm.addRule(TransitionRule{
		FromState:  StateSuppressed,
		ToState:    StateActive,
		Transition: TransitionUnsuppress,
		Allowed:    true,
	})
	sm.addRule(TransitionRule{
		FromState:  StateSuppressed,
		ToState:    StateCleared,
		Transition: TransitionClear,
		Allowed:    true,
	})

	// Cleared 状态的转换规则 - 已清除的告警可以重新激活
	sm.addRule(TransitionRule{
		FromState:  StateCleared,
		ToState:    StateActive,
		Transition: TransitionTrigger,
		Allowed:    true,
	})
}

// addRule 添加转换规则
func (sm *StateMachine) addRule(rule TransitionRule) {
	if _, ok := sm.rules[rule.FromState]; !ok {
		sm.rules[rule.FromState] = make(map[StateTransition]TransitionRule)
	}
	sm.rules[rule.FromState][rule.Transition] = rule
}

// AddRule 添加自定义转换规则
func (sm *StateMachine) AddRule(rule TransitionRule) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.addRule(rule)
}

// CanTransition 检查是否可以进行状态转换
func (sm *StateMachine) CanTransition(currentState AlertState, transition StateTransition) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stateRules, ok := sm.rules[currentState]
	if !ok {
		return false
	}
	rule, ok := stateRules[transition]
	if !ok {
		return false
	}
	return rule.Allowed
}

// GetNextState 获取转换后的状态
func (sm *StateMachine) GetNextState(currentState AlertState, transition StateTransition) (AlertState, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stateRules, ok := sm.rules[currentState]
	if !ok {
		return currentState, fmt.Errorf("no rules defined for state: %s", currentState)
	}
	rule, ok := stateRules[transition]
	if !ok {
		return currentState, fmt.Errorf("transition %s not allowed from state %s", transition, currentState)
	}
	if !rule.Allowed {
		return currentState, fmt.Errorf("transition %s is not allowed from state %s", transition, currentState)
	}
	return rule.ToState, nil
}

// Transition 执行状态转换
func (sm *StateMachine) Transition(ctx context.Context, alarm *entity.Alarm, transition StateTransition, operator, reason string) error {
	currentState := StateFromEntityStatus(alarm.Status)
	nextState, err := sm.GetNextState(currentState, transition)
	if err != nil {
		return err
	}

	// 更新告警状态
	now := time.Now()
	switch transition {
	case TransitionTrigger:
		alarm.Status = entity.AlarmStatusActive
		alarm.TriggeredAt = now
	case TransitionAcknowledge:
		alarm.Status = entity.AlarmStatusAcknowledged
		alarm.AcknowledgedAt = &now
		alarm.AcknowledgedBy = operator
	case TransitionClear:
		alarm.Status = entity.AlarmStatusCleared
		alarm.ClearedAt = &now
	case TransitionSuppress:
		alarm.Status = entity.AlarmStatusSuppressed
	case TransitionUnsuppress:
		alarm.Status = entity.AlarmStatusActive
	case TransitionReactivate:
		alarm.Status = entity.AlarmStatusActive
		alarm.AcknowledgedAt = nil
		alarm.AcknowledgedBy = ""
	}

	// 发送状态变更事件
	event := StateChangeEvent{
		AlarmID:    alarm.ID,
		FromState:  currentState,
		ToState:    nextState,
		Transition: transition,
		Timestamp:  now,
		Operator:   operator,
		Reason:     reason,
	}

	select {
	case sm.eventChan <- event:
	default:
		// 缓冲区满，异步处理
		go func() {
			sm.eventChan <- event
		}()
	}

	return nil
}

// AddHandler 添加状态变更处理器
func (sm *StateMachine) AddHandler(handler StateChangeHandler) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.handlers = append(sm.handlers, handler)
}

// Start 启动状态机
func (sm *StateMachine) Start(ctx context.Context) {
	for i := 0; i < sm.workerCount; i++ {
		sm.wg.Add(1)
		go sm.eventWorker(ctx, i)
	}
}

// Stop 停止状态机
func (sm *StateMachine) Stop() {
	close(sm.stopChan)
	sm.wg.Wait()
	close(sm.eventChan)
}

// eventWorker 事件处理工作协程
func (sm *StateMachine) eventWorker(ctx context.Context, workerID int) {
	defer sm.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChan:
			return
		case event, ok := <-sm.eventChan:
			if !ok {
				return
			}
			sm.handleEvent(ctx, event)
		}
	}
}

// handleEvent 处理状态变更事件
func (sm *StateMachine) handleEvent(ctx context.Context, event StateChangeEvent) {
	sm.mu.RLock()
	handlers := make([]StateChangeHandler, len(sm.handlers))
	copy(handlers, sm.handlers)
	sm.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			// 记录错误但继续处理其他处理器
			fmt.Printf("state change handler error: %v, event: %+v\n", err, event)
		}
	}
}

// GetValidTransitions 获取当前状态可用的转换
func (sm *StateMachine) GetValidTransitions(currentState AlertState) []StateTransition {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var transitions []StateTransition
	stateRules, ok := sm.rules[currentState]
	if !ok {
		return transitions
	}

	for transition, rule := range stateRules {
		if rule.Allowed {
			transitions = append(transitions, transition)
		}
	}
	return transitions
}

// ValidateTransition 验证状态转换是否有效
func (sm *StateMachine) ValidateTransition(alarm *entity.Alarm, transition StateTransition) error {
	currentState := StateFromEntityStatus(alarm.Status)
	
	if !sm.CanTransition(currentState, transition) {
		return errors.New("invalid state transition")
	}

	// 检查特定转换的业务规则
	switch transition {
	case TransitionAcknowledge:
		if alarm.Status != entity.AlarmStatusActive {
			return errors.New("can only acknowledge active alarms")
		}
	case TransitionClear:
		if alarm.Status != entity.AlarmStatusActive && alarm.Status != entity.AlarmStatusAcknowledged && alarm.Status != entity.AlarmStatusSuppressed {
			return errors.New("can only clear active, acknowledged or suppressed alarms")
		}
	case TransitionSuppress:
		if alarm.Status == entity.AlarmStatusCleared {
			return errors.New("cannot suppress cleared alarms")
		}
	}

	return nil
}

// StateInfo 状态信息
type StateInfo struct {
	CurrentState       AlertState       `json:"current_state"`
	ValidTransitions   []StateTransition `json:"valid_transitions"`
	LastTransitionTime time.Time        `json:"last_transition_time,omitempty"`
	TransitionCount    int              `json:"transition_count"`
}

// GetStateInfo 获取告警状态信息
func (sm *StateMachine) GetStateInfo(alarm *entity.Alarm) StateInfo {
	currentState := StateFromEntityStatus(alarm.Status)
	return StateInfo{
		CurrentState:     currentState,
		ValidTransitions: sm.GetValidTransitions(currentState),
	}
}

// BatchTransition 批量状态转换
func (sm *StateMachine) BatchTransition(ctx context.Context, alarms []*entity.Alarm, transition StateTransition, operator, reason string) []error {
	errs := make([]error, len(alarms))
	for i, alarm := range alarms {
		errs[i] = sm.Transition(ctx, alarm, transition, operator, reason)
	}
	return errs
}

// StateSnapshot 状态快照
type StateSnapshot struct {
	AlarmID    string     `json:"alarm_id"`
	State      AlertState `json:"state"`
	UpdatedAt  time.Time  `json:"updated_at"`
	UpdatedBy  string     `json:"updated_by,omitempty"`
}

// CreateSnapshot 创建状态快照
func (sm *StateMachine) CreateSnapshot(alarm *entity.Alarm) StateSnapshot {
	return StateSnapshot{
		AlarmID:   alarm.ID,
		State:     StateFromEntityStatus(alarm.Status),
		UpdatedAt: time.Now(),
	}
}
