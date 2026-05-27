package state

import (
	"context"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertState_String(t *testing.T) {
	assert.Equal(t, "Active", StateActive.String())
	assert.Equal(t, "Acknowledged", StateAcknowledged.String())
	assert.Equal(t, "Cleared", StateCleared.String())
	assert.Equal(t, "Suppressed", StateSuppressed.String())
	assert.Equal(t, "Unknown", AlertState(99).String())
}

func TestAlertState_ToEntityStatus(t *testing.T) {
	assert.Equal(t, entity.AlarmStatusActive, StateActive.ToEntityStatus())
	assert.Equal(t, entity.AlarmStatusAcknowledged, StateAcknowledged.ToEntityStatus())
	assert.Equal(t, entity.AlarmStatusCleared, StateCleared.ToEntityStatus())
	assert.Equal(t, entity.AlarmStatusSuppressed, StateSuppressed.ToEntityStatus())
	assert.Equal(t, entity.AlarmStatusActive, AlertState(99).ToEntityStatus())
}

func TestStateFromEntityStatus(t *testing.T) {
	assert.Equal(t, StateActive, StateFromEntityStatus(entity.AlarmStatusActive))
	assert.Equal(t, StateAcknowledged, StateFromEntityStatus(entity.AlarmStatusAcknowledged))
	assert.Equal(t, StateCleared, StateFromEntityStatus(entity.AlarmStatusCleared))
	assert.Equal(t, StateSuppressed, StateFromEntityStatus(entity.AlarmStatusSuppressed))
	assert.Equal(t, StateActive, StateFromEntityStatus(entity.AlarmStatus(99)))
}

func TestStateTransition_String(t *testing.T) {
	assert.Equal(t, "None", TransitionNone.String())
	assert.Equal(t, "Trigger", TransitionTrigger.String())
	assert.Equal(t, "Acknowledge", TransitionAcknowledge.String())
	assert.Equal(t, "Clear", TransitionClear.String())
	assert.Equal(t, "Suppress", TransitionSuppress.String())
	assert.Equal(t, "Unsuppress", TransitionUnsuppress.String())
	assert.Equal(t, "Reactivate", TransitionReactivate.String())
	assert.Equal(t, "Unknown", StateTransition(99).String())
}

func TestNewStateMachine(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})
	require.NotNil(t, sm)
}

func TestNewStateMachine_CustomConfig(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{
		WorkerCount: 2,
		BufferSize:  500,
	})
	require.NotNil(t, sm)
}

func TestStateMachine_CanTransition(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})

	assert.True(t, sm.CanTransition(StateActive, TransitionAcknowledge))
	assert.True(t, sm.CanTransition(StateActive, TransitionClear))
	assert.True(t, sm.CanTransition(StateActive, TransitionSuppress))
	assert.True(t, sm.CanTransition(StateAcknowledged, TransitionClear))
	assert.True(t, sm.CanTransition(StateAcknowledged, TransitionReactivate))
	assert.True(t, sm.CanTransition(StateSuppressed, TransitionUnsuppress))
	assert.True(t, sm.CanTransition(StateCleared, TransitionTrigger))

	assert.False(t, sm.CanTransition(StateActive, TransitionUnsuppress))
	assert.False(t, sm.CanTransition(StateCleared, TransitionAcknowledge))
	assert.False(t, sm.CanTransition(StateAcknowledged, TransitionTrigger))
}

func TestStateMachine_GetNextState(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})

	next, err := sm.GetNextState(StateActive, TransitionAcknowledge)
	require.NoError(t, err)
	assert.Equal(t, StateAcknowledged, next)

	next, err = sm.GetNextState(StateActive, TransitionClear)
	require.NoError(t, err)
	assert.Equal(t, StateCleared, next)

	next, err = sm.GetNextState(StateActive, TransitionSuppress)
	require.NoError(t, err)
	assert.Equal(t, StateSuppressed, next)

	_, err = sm.GetNextState(StateActive, TransitionUnsuppress)
	assert.Error(t, err)
}

func TestStateMachine_Transition(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")

	err := sm.Transition(ctx, alarm, TransitionAcknowledge, "admin", "acknowledged")
	require.NoError(t, err)
	assert.Equal(t, entity.AlarmStatusAcknowledged, alarm.Status)
	assert.Equal(t, "admin", alarm.AcknowledgedBy)
}

func TestStateMachine_Transition_Clear(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	sm.Transition(ctx, alarm, TransitionClear, "system", "auto clear")
	assert.Equal(t, entity.AlarmStatusCleared, alarm.Status)
	assert.NotNil(t, alarm.ClearedAt)
}

func TestStateMachine_Transition_Suppress(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	sm.Transition(ctx, alarm, TransitionSuppress, "admin", "suppress")
	assert.Equal(t, entity.AlarmStatusSuppressed, alarm.Status)
}

func TestStateMachine_Transition_Unsuppress(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	sm.Transition(ctx, alarm, TransitionSuppress, "admin", "")
	sm.Transition(ctx, alarm, TransitionUnsuppress, "admin", "")
	assert.Equal(t, entity.AlarmStatusActive, alarm.Status)
}

func TestStateMachine_Transition_Reactivate(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	sm.Transition(ctx, alarm, TransitionAcknowledge, "admin", "")
	sm.Transition(ctx, alarm, TransitionReactivate, "admin", "")
	assert.Equal(t, entity.AlarmStatusActive, alarm.Status)
	assert.Nil(t, alarm.AcknowledgedAt)
	assert.Equal(t, "", alarm.AcknowledgedBy)
}

func TestStateMachine_Transition_Trigger(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	sm.Transition(ctx, alarm, TransitionClear, "system", "")
	sm.Transition(ctx, alarm, TransitionTrigger, "system", "re-triggered")
	assert.Equal(t, entity.AlarmStatusActive, alarm.Status)
}

func TestStateMachine_Transition_Invalid(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})
	ctx := context.Background()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	err := sm.Transition(ctx, alarm, TransitionUnsuppress, "admin", "")
	assert.Error(t, err)
}

func TestStateMachine_AddRule(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})
	sm.AddRule(TransitionRule{
		FromState:  StateCleared,
		ToState:    StateSuppressed,
		Transition: StateTransition(100),
		Allowed:    true,
	})
	assert.True(t, sm.CanTransition(StateCleared, StateTransition(100)))
}

func TestStateMachine_AddHandler(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})
	handlerCalled := false
	sm.AddHandler(func(ctx context.Context, event StateChangeEvent) error {
		handlerCalled = true
		return nil
	})

	ctx := context.Background()
	sm.Start(ctx)
	defer sm.Stop()

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	sm.Transition(ctx, alarm, TransitionClear, "admin", "")

	time.Sleep(100 * time.Millisecond)
	assert.True(t, handlerCalled)
}

func TestStateMachine_StartStop(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{WorkerCount: 2})
	ctx := context.Background()

	sm.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	sm.Stop()
}

func TestStateMachine_GetValidTransitions(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})

	transitions := sm.GetValidTransitions(StateActive)
	assert.Equal(t, 3, len(transitions))

	transitions = sm.GetValidTransitions(StateCleared)
	assert.Equal(t, 1, len(transitions))

	transitions = sm.GetValidTransitions(AlertState(99))
	assert.Equal(t, 0, len(transitions))
}

func TestStateMachine_ValidateTransition(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	err := sm.ValidateTransition(alarm, TransitionAcknowledge)
	require.NoError(t, err)

	alarm.Status = entity.AlarmStatusCleared
	err = sm.ValidateTransition(alarm, TransitionAcknowledge)
	assert.Error(t, err)
}

func TestStateMachine_GetStateInfo(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	info := sm.GetStateInfo(alarm)
	assert.Equal(t, StateActive, info.CurrentState)
	assert.Equal(t, 3, len(info.ValidTransitions))
}

func TestStateMachine_BatchTransition(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})
	ctx := context.Background()

	alarms := []*entity.Alarm{
		entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A1", ""),
		entity.NewAlarm("pt2", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "A2", ""),
	}

	errs := sm.BatchTransition(ctx, alarms, TransitionClear, "admin", "batch clear")
	assert.Equal(t, 2, len(errs))
	assert.Nil(t, errs[0])
	assert.Nil(t, errs[1])
	assert.Equal(t, entity.AlarmStatusCleared, alarms[0].Status)
	assert.Equal(t, entity.AlarmStatusCleared, alarms[1].Status)
}

func TestStateMachine_CreateSnapshot(t *testing.T) {
	sm := NewStateMachine(StateMachineConfig{})

	alarm := entity.NewAlarm("pt1", "dev1", "st1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
	snapshot := sm.CreateSnapshot(alarm)
	assert.Equal(t, StateActive, snapshot.State)
	assert.Equal(t, alarm.ID, snapshot.AlarmID)
	assert.False(t, snapshot.UpdatedAt.IsZero())
}

func TestStateMachineConfig_Defaults(t *testing.T) {
	cfg := StateMachineConfig{}
	sm := NewStateMachine(cfg)
	require.NotNil(t, sm)
}

func TestStateChangeEvent_Struct(t *testing.T) {
	event := StateChangeEvent{
		AlarmID:    "alarm1",
		FromState:  StateActive,
		ToState:    StateCleared,
		Transition: TransitionClear,
		Timestamp:  time.Now(),
		Operator:   "admin",
		Reason:     "resolved",
		Metadata:   map[string]any{"key": "value"},
	}
	assert.Equal(t, "alarm1", event.AlarmID)
	assert.Equal(t, StateActive, event.FromState)
}

func TestTransitionRule_Struct(t *testing.T) {
	rule := TransitionRule{
		FromState:     StateActive,
		ToState:       StateCleared,
		Transition:    TransitionClear,
		Allowed:       true,
		RequireReason: true,
	}
	assert.True(t, rule.Allowed)
	assert.True(t, rule.RequireReason)
}

func TestStateInfo_Struct(t *testing.T) {
	info := StateInfo{
		CurrentState:     StateActive,
		ValidTransitions: []StateTransition{TransitionAcknowledge, TransitionClear},
		TransitionCount:  5,
	}
	assert.Equal(t, StateActive, info.CurrentState)
	assert.Equal(t, 2, len(info.ValidTransitions))
}

func TestStateSnapshot_Struct(t *testing.T) {
	snap := StateSnapshot{
		AlarmID:   "alarm1",
		State:     StateActive,
		UpdatedAt: time.Now(),
		UpdatedBy: "admin",
	}
	assert.Equal(t, "alarm1", snap.AlarmID)
}
