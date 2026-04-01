package iec61850

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

// ControlObject represents a controllable object
type ControlObject struct {
	Reference       string
	Model           ControlModel
	State           ControlState
	Value           interface{}
	selectTime      time.Time
	operateTime     time.Time
	terminalConfirm bool
	mutex           sync.RWMutex
}

// NewControlObject creates a new control object
func NewControlObject(ref string, model ControlModel) *ControlObject {
	return &ControlObject{
		Reference: ref,
		Model:     model,
		State:     ControlStateIdle,
	}
}

// Select selects the control object for operation
func (c *ControlObject) Select() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Check if already selected
	if c.State == ControlStateSelected {
		return fmt.Errorf("already selected")
	}
	
	// Check if in valid state
	if c.State != ControlStateIdle {
		return fmt.Errorf("invalid state for select: %s", c.State)
	}
	
	c.State = ControlStateSelected
	c.selectTime = time.Now()
	
	return nil
}

// Operate operates the control object
func (c *ControlObject) Operate(value interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Direct control doesn't need select
	if c.Model == ControlModelDirectNormal || c.Model == ControlModelDirectEnhanced {
		c.State = ControlStateOperating
		c.operateTime = time.Now()
		c.Value = value
		
		// Simulate immediate completion
		c.State = ControlStateCompleted
		return nil
	}
	
	// SBO control needs select first
	if c.State != ControlStateSelected {
		return fmt.Errorf("must select before operate")
	}
	
	c.State = ControlStateOperating
	c.operateTime = time.Now()
	c.Value = value
	
	// Simulate immediate completion
	c.State = ControlStateCompleted
	
	return nil
}

// Cancel cancels the control operation
func (c *ControlObject) Cancel() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if c.State == ControlStateIdle || c.State == ControlStateCompleted {
		return fmt.Errorf("cannot cancel in state %s", c.State)
	}
	
	c.State = ControlStateCancelled
	return nil
}

// Reset resets the control object to idle state
func (c *ControlObject) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.State = ControlStateIdle
	c.Value = nil
	c.terminalConfirm = false
}

// SetTerminalConfirmation sets the terminal confirmation
func (c *ControlObject) SetTerminalConfirmation(confirmed bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.terminalConfirm = confirmed
}

// GetTerminalConfirmation returns the terminal confirmation
func (c *ControlObject) GetTerminalConfirmation() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return c.terminalConfirm
}

// IsSelectTimeout checks if select has timed out
func (c *ControlObject) IsSelectTimeout(timeout time.Duration) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	if c.State != ControlStateSelected {
		return false
	}
	
	return time.Since(c.selectTime) > timeout
}

// IsControlTimeout checks if control has timed out
func (c *ControlObject) IsControlTimeout(timeout time.Duration) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	if c.State != ControlStateOperating {
		return false
	}
	
	return time.Since(c.operateTime) > timeout
}

// ControlRequest represents a control request
type ControlRequest struct {
	Reference string
	Value     interface{}
	Test      bool
	Check     bool
}

// ControlResponse represents a control response
type ControlResponse struct {
	Reference string
	Result    MMSResult
	AddCause  string
}

// ControlHandler handles control services
type ControlHandler struct {
	config    *Config
	objects   map[string]*ControlObject
	mutex     sync.RWMutex
}

// NewControlHandler creates a new control handler
func NewControlHandler(config *Config) *ControlHandler {
	return &ControlHandler{
		config:  config,
		objects: make(map[string]*ControlObject),
	}
}

// AddControlObject adds a control object
func (h *ControlHandler) AddControlObject(obj *ControlObject) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.objects[obj.Reference] = obj
	return nil
}

// GetControlObject retrieves a control object
func (h *ControlHandler) GetControlObject(ref string) (*ControlObject, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	obj, exists := h.objects[ref]
	return obj, exists
}

// Select selects a control object
func (h *ControlHandler) Select(ref string) error {
	h.mutex.RLock()
	obj, exists := h.objects[ref]
	h.mutex.RUnlock()
	
	if !exists {
		// Create new control object with default model
		obj = NewControlObject(ref, ControlModelSBONormal)
		h.mutex.Lock()
		h.objects[ref] = obj
		h.mutex.Unlock()
	}
	
	return obj.Select()
}

// Operate operates a control object
func (h *ControlHandler) Operate(ref string, value interface{}) error {
	h.mutex.RLock()
	obj, exists := h.objects[ref]
	h.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("control object not found: %s", ref)
	}
	
	return obj.Operate(value)
}

// Cancel cancels a control operation
func (h *ControlHandler) Cancel(ref string) error {
	h.mutex.RLock()
	obj, exists := h.objects[ref]
	h.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("control object not found: %s", ref)
	}
	
	return obj.Cancel()
}

// GetControlState returns the control state
func (h *ControlHandler) GetControlState(ref string) ControlState {
	h.mutex.RLock()
	obj, exists := h.objects[ref]
	h.mutex.RUnlock()
	
	if !exists {
		return ControlStateIdle
	}
	
	obj.mutex.RLock()
	defer obj.mutex.RUnlock()
	return obj.State
}

// SetTerminalConfirmation sets the terminal confirmation
func (h *ControlHandler) SetTerminalConfirmation(ref string, confirmed bool) error {
	h.mutex.RLock()
	obj, exists := h.objects[ref]
	h.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("control object not found: %s", ref)
	}
	
	obj.SetTerminalConfirmation(confirmed)
	return nil
}

// GetTerminalConfirmation returns the terminal confirmation
func (h *ControlHandler) GetTerminalConfirmation(ref string) bool {
	h.mutex.RLock()
	obj, exists := h.objects[ref]
	h.mutex.RUnlock()
	
	if !exists {
		return false
	}
	
	return obj.GetTerminalConfirmation()
}

// IsSelectTimeout checks if select has timed out
func (h *ControlHandler) IsSelectTimeout(ref string) bool {
	h.mutex.RLock()
	obj, exists := h.objects[ref]
	h.mutex.RUnlock()
	
	if !exists {
		return false
	}
	
	return obj.IsSelectTimeout(h.config.SelectTimeout)
}

// IsControlTimeout checks if control has timed out
func (h *ControlHandler) IsControlTimeout(ref string) bool {
	h.mutex.RLock()
	obj, exists := h.objects[ref]
	h.mutex.RUnlock()
	
	if !exists {
		return false
	}
	
	return obj.IsControlTimeout(h.config.ControlTimeout)
}

// EncodeControlRequest encodes a control request
func EncodeControlRequest(req *ControlRequest) ([]byte, error) {
	// Calculate length
	refLen := len(req.Reference)
	totalLen := 2 + refLen + 1 + 1 + 1
	
	data := make([]byte, totalLen)
	offset := 0
	
	// Reference
	binary.BigEndian.PutUint16(data[offset:], uint16(refLen))
	offset += 2
	copy(data[offset:], req.Reference)
	offset += refLen
	
	// Value (as bool for simplicity)
	if b, ok := req.Value.(bool); ok {
		if b {
			data[offset] = 1
		} else {
			data[offset] = 0
		}
	}
	offset += 1
	
	// Test flag
	if req.Test {
		data[offset] = 1
	} else {
		data[offset] = 0
	}
	offset += 1
	
	// Check flag
	if req.Check {
		data[offset] = 1
	} else {
		data[offset] = 0
	}
	
	return data, nil
}

// DecodeControlResponse decodes a control response
func DecodeControlResponse(data []byte) (*ControlResponse, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short for control response")
	}
	
	resp := &ControlResponse{}
	offset := 0
	
	// Reference
	refLen := int(binary.BigEndian.Uint16(data[offset:]))
	offset += 2
	
	if len(data) < offset+refLen+2 {
		return nil, fmt.Errorf("data too short for reference")
	}
	
	resp.Reference = string(data[offset : offset+refLen])
	offset += refLen
	
	// Result
	resp.Result = MMSResult(data[offset])
	offset += 1
	
	// AddCause (simplified)
	if len(data) > offset {
		addCauseLen := int(data[offset])
		offset += 1
		if len(data) >= offset+addCauseLen {
			resp.AddCause = string(data[offset : offset+addCauseLen])
		}
	}
	
	return resp, nil
}
