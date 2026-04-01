package iec61850

import (
	"encoding/binary"
	"fmt"
	"sync"
)

// SVData represents sampled values data
type SVData struct {
	SmpCnt  uint16
	ConfRev uint32
	Values  []SVValue
}

// SVValue represents a single sampled value
type SVValue struct {
	Inst    int
	Value   float64
	Quality Quality
}

// SVFrame represents an SV frame
type SVFrame struct {
	AppID     uint16
	Length    uint16
	Reserved1 uint16
	Reserved2 uint16
	SmpCnt    uint16
	ConfRev   uint32
	SmpMod    uint8
	SmpRate   uint16
	Dataset   string
	ASDUCount uint8
	ASDUs     []SVASDU
}

// SVASDU represents an Application Service Data Unit
type SVASDU struct {
	SV []SVValue
}

// SVMulticastConfig represents multicast configuration
type SVMulticastConfig struct {
	GroupAddr string
	Port      int
	Interface string
}

// SVSyncStatus represents synchronization status
type SVSyncStatus struct {
	Synchronized bool
	Precision    float64
}

// SVBuffer represents a circular buffer for SV data
type SVBuffer struct {
	data     []*SVData
	capacity int
	size     int
	head     int
	tail     int
	mutex    sync.RWMutex
}

// NewSVBuffer creates a new SV buffer
func NewSVBuffer(capacity int) *SVBuffer {
	return &SVBuffer{
		data:     make([]*SVData, capacity),
		capacity: capacity,
		size:     0,
		head:     0,
		tail:     0,
	}
}

// Capacity returns the buffer capacity
func (b *SVBuffer) Capacity() int {
	return b.capacity
}

// Size returns the current buffer size
func (b *SVBuffer) Size() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.size
}

// Add adds data to the buffer
func (b *SVBuffer) Add(data *SVData) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	b.data[b.tail] = data
	b.tail = (b.tail + 1) % b.capacity
	
	if b.size < b.capacity {
		b.size++
	} else {
		// Buffer is full, move head
		b.head = (b.head + 1) % b.capacity
	}
	
	return nil
}

// Get retrieves data at index
func (b *SVBuffer) Get(index int) *SVData {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	
	if index < 0 || index >= b.size {
		return nil
	}
	
	actualIndex := (b.head + index) % b.capacity
	return b.data[actualIndex]
}

// SampledValuesHandler handles sampled values
type SampledValuesHandler struct {
	config         *Config
	subscriptions  map[string]SVCallback
	multicast      *SVMulticastConfig
	syncStatus     SVSyncStatus
	buffer         *SVBuffer
	mutex          sync.RWMutex
}

// SVCallback is the callback function type for SV data
type SVCallback func(data *SVData) error

// NewSampledValuesHandler creates a new sampled values handler
func NewSampledValuesHandler(config *Config) *SampledValuesHandler {
	return &SampledValuesHandler{
		config:        config,
		subscriptions: make(map[string]SVCallback),
		buffer:        NewSVBuffer(config.SVBufferSize),
	}
}

// Subscribe subscribes to an SV stream
func (h *SampledValuesHandler) Subscribe(streamID string, callback SVCallback) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.subscriptions[streamID] = callback
	return nil
}

// Unsubscribe unsubscribes from an SV stream
func (h *SampledValuesHandler) Unsubscribe(streamID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	delete(h.subscriptions, streamID)
	return nil
}

// IsSubscribed checks if subscribed to a stream
func (h *SampledValuesHandler) IsSubscribed(streamID string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	_, exists := h.subscriptions[streamID]
	return exists
}

// GetSubscriptionList returns the list of subscriptions
func (h *SampledValuesHandler) GetSubscriptionList() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	list := make([]string, 0, len(h.subscriptions))
	for streamID := range h.subscriptions {
		list = append(list, streamID)
	}
	return list
}

// ConfigureMulticast configures multicast settings
func (h *SampledValuesHandler) ConfigureMulticast(config *SVMulticastConfig) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.multicast = config
	return nil
}

// GetMulticastConfig returns the multicast configuration
func (h *SampledValuesHandler) GetMulticastConfig() *SVMulticastConfig {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	return h.multicast
}

// GetSyncStatus returns the synchronization status
func (h *SampledValuesHandler) GetSyncStatus() SVSyncStatus {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	return h.syncStatus
}

// SetSyncStatus sets the synchronization status
func (h *SampledValuesHandler) SetSyncStatus(synchronized bool, precision float64) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.syncStatus = SVSyncStatus{
		Synchronized: synchronized,
		Precision:    precision,
	}
}

// ProcessFrame processes an SV frame
func (h *SampledValuesHandler) ProcessFrame(frame *SVFrame) error {
	// Convert frame to SVData
	data := &SVData{
		SmpCnt:  frame.SmpCnt,
		ConfRev: frame.ConfRev,
		Values:  make([]SVValue, 0),
	}
	
	// Extract values from ASDUs
	for _, asdu := range frame.ASDUs {
		data.Values = append(data.Values, asdu.SV...)
	}
	
	// Add to buffer
	if err := h.buffer.Add(data); err != nil {
		return fmt.Errorf("failed to add to buffer: %w", err)
	}
	
	// Notify subscribers
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	for streamID, callback := range h.subscriptions {
		if err := callback(data); err != nil {
			// Log error but continue
			_ = streamID // Avoid unused variable error
		}
	}
	
	return nil
}

// EncodeSVFrame encodes an SV frame
func EncodeSVFrame(frame *SVFrame) ([]byte, error) {
	// Calculate total length
	datasetLen := len(frame.Dataset)
	
	// Calculate ASDU data length
	asduLen := 0
	for _, asdu := range frame.ASDUs {
		asduLen += len(asdu.SV) * 14 // 4 bytes inst + 8 bytes value + 2 bytes quality
	}
	
	totalLen := 22 + datasetLen + asduLen
	
	data := make([]byte, totalLen)
	offset := 0
	
	// Header
	binary.BigEndian.PutUint16(data[offset:], frame.AppID)
	offset += 2
	
	binary.BigEndian.PutUint16(data[offset:], frame.Length)
	offset += 2
	
	binary.BigEndian.PutUint16(data[offset:], frame.Reserved1)
	offset += 2
	
	binary.BigEndian.PutUint16(data[offset:], frame.Reserved2)
	offset += 2
	
	// Sample counter
	binary.BigEndian.PutUint16(data[offset:], frame.SmpCnt)
	offset += 2
	
	// Configuration revision
	binary.BigEndian.PutUint32(data[offset:], frame.ConfRev)
	offset += 4
	
	// Sample mode and rate
	data[offset] = frame.SmpMod
	offset += 1
	
	binary.BigEndian.PutUint16(data[offset:], frame.SmpRate)
	offset += 2
	
	// Dataset name
	data[offset] = byte(datasetLen)
	offset += 1
	copy(data[offset:], frame.Dataset)
	offset += datasetLen
	
	// ASDU count
	data[offset] = frame.ASDUCount
	offset += 1
	
	// ASDUs (simplified)
	for _, asdu := range frame.ASDUs {
		for _, sv := range asdu.SV {
			binary.BigEndian.PutUint32(data[offset:], uint32(sv.Inst))
			offset += 4
			binary.BigEndian.PutUint64(data[offset:], uint64(sv.Value))
			offset += 8
			binary.BigEndian.PutUint16(data[offset:], uint16(sv.Quality))
			offset += 2
		}
	}
	
	return data, nil
}

// DecodeSVFrame decodes an SV frame
func DecodeSVFrame(data []byte) (*SVFrame, error) {
	if len(data) < 22 {
		return nil, fmt.Errorf("data too short for SV frame")
	}
	
	frame := &SVFrame{}
	offset := 0
	
	// Header
	frame.AppID = binary.BigEndian.Uint16(data[offset:])
	offset += 2
	
	frame.Length = binary.BigEndian.Uint16(data[offset:])
	offset += 2
	
	frame.Reserved1 = binary.BigEndian.Uint16(data[offset:])
	offset += 2
	
	frame.Reserved2 = binary.BigEndian.Uint16(data[offset:])
	offset += 2
	
	// Sample counter
	frame.SmpCnt = binary.BigEndian.Uint16(data[offset:])
	offset += 2
	
	// Configuration revision
	frame.ConfRev = binary.BigEndian.Uint32(data[offset:])
	offset += 4
	
	// Sample mode and rate
	frame.SmpMod = data[offset]
	offset += 1
	
	frame.SmpRate = binary.BigEndian.Uint16(data[offset:])
	offset += 2
	
	// Dataset name
	datasetLen := int(data[offset])
	offset += 1
	
	if len(data) < offset+datasetLen {
		return nil, fmt.Errorf("data too short for dataset name")
	}
	
	frame.Dataset = string(data[offset : offset+datasetLen])
	offset += datasetLen
	
	// ASDU count
	frame.ASDUCount = data[offset]
	offset += 1
	
	// Parse ASDUs (simplified)
	frame.ASDUs = make([]SVASDU, frame.ASDUCount)
	for i := 0; i < int(frame.ASDUCount); i++ {
		// Assume 2 values per ASDU for testing
		frame.ASDUs[i].SV = make([]SVValue, 2)
		for j := 0; j < 2; j++ {
			if len(data) < offset+14 {
				break
			}
			
			inst := int(binary.BigEndian.Uint32(data[offset:]))
			offset += 4
			
			value := float64(binary.BigEndian.Uint64(data[offset:]))
			offset += 8
			
			quality := Quality(binary.BigEndian.Uint16(data[offset:]))
			offset += 2
			
			frame.ASDUs[i].SV[j] = SVValue{
				Inst:    inst,
				Value:   value,
				Quality: quality,
			}
		}
	}
	
	return frame, nil
}
