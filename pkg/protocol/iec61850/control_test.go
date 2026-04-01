package iec61850

import (
	"testing"
	"time"
)

// TestControlHandler tests control handler
func TestControlHandler(t *testing.T) {
	t.Run("create handler", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewControlHandler(config)
		
		if handler == nil {
			t.Fatal("handler should not be nil")
		}
	})

	t.Run("select control object", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewControlHandler(config)
		
		ref := "LD0/XCBR1.Pos"
		err := handler.Select(ref)
		
		if err != nil {
			t.Errorf("failed to select control object: %v", err)
		}
		
		if handler.GetControlState(ref) != ControlStateSelected {
			t.Errorf("expected state Selected, got %s", handler.GetControlState(ref))
		}
	})

	t.Run("operate control object", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewControlHandler(config)
		
		ref := "LD0/XCBR1.Pos"
		_ = handler.Select(ref)
		
		err := handler.Operate(ref, true)
		if err != nil {
			t.Errorf("failed to operate control object: %v", err)
		}
		
		if handler.GetControlState(ref) != ControlStateCompleted {
			t.Errorf("expected state Completed, got %s", handler.GetControlState(ref))
		}
	})

	t.Run("cancel control operation", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewControlHandler(config)
		
		ref := "LD0/XCBR1.Pos"
		_ = handler.Select(ref)
		
		err := handler.Cancel(ref)
		if err != nil {
			t.Errorf("failed to cancel control operation: %v", err)
		}
		
		if handler.GetControlState(ref) != ControlStateCancelled {
			t.Errorf("expected state Cancelled, got %s", handler.GetControlState(ref))
		}
	})
}

// TestControlObject tests control object operations
func TestControlObject(t *testing.T) {
	t.Run("create control object", func(t *testing.T) {
		ref := "LD0/XCBR1.Pos"
		ctlObj := NewControlObject(ref, ControlModelSBONormal)
		
		if ctlObj.Reference != ref {
			t.Errorf("expected reference %s, got %s", ref, ctlObj.Reference)
		}
		if ctlObj.Model != ControlModelSBONormal {
			t.Errorf("expected model SBO Normal, got %s", ctlObj.Model)
		}
		if ctlObj.State != ControlStateIdle {
			t.Errorf("expected state Idle, got %s", ctlObj.State)
		}
	})

	t.Run("select control object", func(t *testing.T) {
		ctlObj := NewControlObject("LD0/XCBR1.Pos", ControlModelSBONormal)
		
		err := ctlObj.Select()
		if err != nil {
			t.Errorf("failed to select: %v", err)
		}
		
		if ctlObj.State != ControlStateSelected {
			t.Errorf("expected state Selected, got %s", ctlObj.State)
		}
	})

	t.Run("operate without select", func(t *testing.T) {
		ctlObj := NewControlObject("LD0/XCBR1.Pos", ControlModelSBONormal)
		
		err := ctlObj.Operate(true)
		if err == nil {
			t.Error("should return error when operating without select")
		}
	})

	t.Run("direct control model", func(t *testing.T) {
		ctlObj := NewControlObject("LD0/XCBR1.Pos", ControlModelDirectNormal)
		
		// Direct control doesn't need select
		err := ctlObj.Operate(true)
		if err != nil {
			t.Errorf("failed to operate with direct model: %v", err)
		}
		
		if ctlObj.State != ControlStateCompleted {
			t.Errorf("expected state Completed, got %s", ctlObj.State)
		}
	})
}

// TestControlStateMachine tests control state machine
func TestControlStateMachine(t *testing.T) {
	t.Run("state transitions", func(t *testing.T) {
		ctlObj := NewControlObject("LD0/XCBR1.Pos", ControlModelSBONormal)
		
		// Initial state
		if ctlObj.State != ControlStateIdle {
			t.Errorf("expected initial state Idle, got %s", ctlObj.State)
		}
		
		// Select
		_ = ctlObj.Select()
		if ctlObj.State != ControlStateSelected {
			t.Errorf("expected state Selected, got %s", ctlObj.State)
		}
		
		// Operate
		_ = ctlObj.Operate(true)
		if ctlObj.State != ControlStateCompleted {
			t.Errorf("expected state Completed, got %s", ctlObj.State)
		}
		
		// Reset
		ctlObj.Reset()
		if ctlObj.State != ControlStateIdle {
			t.Errorf("expected state Idle after reset, got %s", ctlObj.State)
		}
	})

	t.Run("invalid state transitions", func(t *testing.T) {
		ctlObj := NewControlObject("LD0/XCBR1.Pos", ControlModelSBONormal)
		
		// Try to operate without select
		err := ctlObj.Operate(true)
		if err == nil {
			t.Error("should return error for invalid transition")
		}
		
		// Try to select twice
		_ = ctlObj.Select()
		err = ctlObj.Select()
		if err == nil {
			t.Error("should return error for duplicate select")
		}
	})
}

// TestTerminalConfirmation tests terminal confirmation
func TestTerminalConfirmation(t *testing.T) {
	t.Run("set terminal confirmation", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewControlHandler(config)
		
		ref := "LD0/XCBR1.Pos"
		ctlObj := NewControlObject(ref, ControlModelSBOEnhanced)
		_ = handler.AddControlObject(ctlObj)
		
		// Select and operate
		_ = handler.Select(ref)
		_ = handler.Operate(ref, true)
		
		// Set terminal confirmation
		err := handler.SetTerminalConfirmation(ref, true)
		if err != nil {
			t.Errorf("failed to set terminal confirmation: %v", err)
		}
	})

	t.Run("get terminal confirmation", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewControlHandler(config)
		
		ref := "LD0/XCBR1.Pos"
		ctlObj := NewControlObject(ref, ControlModelSBOEnhanced)
		_ = handler.AddControlObject(ctlObj)
		
		// Get initial terminal confirmation
		confirmed := handler.GetTerminalConfirmation(ref)
		if confirmed {
			t.Error("terminal confirmation should be false initially")
		}
		
		// Set and get
		_ = handler.SetTerminalConfirmation(ref, true)
		confirmed = handler.GetTerminalConfirmation(ref)
		if !confirmed {
			t.Error("terminal confirmation should be true")
		}
	})
}

// TestControlTimeout tests control timeout
func TestControlTimeout(t *testing.T) {
	t.Run("select timeout", func(t *testing.T) {
		config := DefaultConfig()
		config.SelectTimeout = 100 * time.Millisecond
		
		handler := NewControlHandler(config)
		
		ref := "LD0/XCBR1.Pos"
		ctlObj := NewControlObject(ref, ControlModelSBONormal)
		_ = ctlObj.Select() // Set state to Selected first
		ctlObj.selectTime = time.Now().Add(-200 * time.Millisecond) // Simulate timeout
		_ = handler.AddControlObject(ctlObj)
		
		// Check if select has timed out
		if !handler.IsSelectTimeout(ref) {
			t.Error("select should have timed out")
		}
	})

	t.Run("control timeout", func(t *testing.T) {
		config := DefaultConfig()
		config.ControlTimeout = 100 * time.Millisecond
		
		handler := NewControlHandler(config)
		
		ref := "LD0/XCBR1.Pos"
		ctlObj := NewControlObject(ref, ControlModelSBONormal)
		ctlObj.State = ControlStateOperating // Set state to Operating first
		ctlObj.operateTime = time.Now().Add(-200 * time.Millisecond) // Simulate timeout
		_ = handler.AddControlObject(ctlObj)
		
		// Check if control has timed out
		if !handler.IsControlTimeout(ref) {
			t.Error("control should have timed out")
		}
	})
}

// TestControlEncoding tests control message encoding
func TestControlEncoding(t *testing.T) {
	t.Run("encode control request", func(t *testing.T) {
		req := &ControlRequest{
			Reference: "LD0/XCBR1.Pos",
			Value:     true,
			Test:      false,
			Check:     true,
		}
		
		data, err := EncodeControlRequest(req)
		if err != nil {
			t.Errorf("failed to encode control request: %v", err)
		}
		
		if len(data) == 0 {
			t.Error("encoded data should not be empty")
		}
	})

	t.Run("decode control response", func(t *testing.T) {
		req := &ControlRequest{
			Reference: "LD0/XCBR1.Pos",
			Value:     true,
			Test:      false,
			Check:     true,
		}
		
		data, _ := EncodeControlRequest(req)
		resp, err := DecodeControlResponse(data)
		
		if err != nil {
			t.Errorf("failed to decode control response: %v", err)
		}
		
		if resp.Reference != req.Reference {
			t.Errorf("expected reference %s, got %s", req.Reference, resp.Reference)
		}
	})
}
