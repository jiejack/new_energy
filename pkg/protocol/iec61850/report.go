package iec61850

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

// Report represents an IEC 61850 report
type Report struct {
	RptID    string
	SqNum    uint16
	Time     Timestamp
	ConfRev  uint32
	Values   []ReportValue
}

// ReportValue represents a value in a report
type ReportValue struct {
	Ref     string
	Value   interface{}
	Quality Quality
}

// DataChange represents a data change event
type DataChange struct {
	Ref     string
	Value   interface{}
	Quality Quality
}

// BRCB represents a buffered report control block
type BRCB struct {
	RptID      string
	Domain     string
	DatSet     string
	ConfRev    uint32
	TrgOps     TriggerOptions
	IntgPd     uint32
	RptEna     bool
	Buffered   bool
	SqNum      uint16
	Buffer     *ReportBuffer
	mutex      sync.RWMutex
}

// NewBRCB creates a new buffered report control block
func NewBRCB(rptID, domain string) *BRCB {
	return &BRCB{
		RptID:    rptID,
		Domain:   domain,
		Buffered: true,
		Buffer:   NewReportBuffer(256),
	}
}

// SetDataset sets the dataset for the BRCB
func (b *BRCB) SetDataset(dataset string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	b.DatSet = dataset
	return nil
}

// SetTriggerOptions sets the trigger options
func (b *BRCB) SetTriggerOptions(options TriggerOptions) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	b.TrgOps = options
	return nil
}

// SetIntegrityPeriod sets the integrity period
func (b *BRCB) SetIntegrityPeriod(period uint32) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	b.IntgPd = period
	return nil
}

// Enable enables the report
func (b *BRCB) Enable() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	b.RptEna = true
	return nil
}

// Disable disables the report
func (b *BRCB) Disable() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	b.RptEna = false
	return nil
}

// IsEnabled returns whether the report is enabled
func (b *BRCB) IsEnabled() bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	
	return b.RptEna
}

// AddReport adds a report to the buffer
func (b *BRCB) AddReport(report *Report) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	if !b.RptEna {
		return fmt.Errorf("report not enabled")
	}
	
	return b.Buffer.Add(report)
}

// URCB represents an unbuffered report control block
type URCB struct {
	RptID    string
	Domain   string
	DatSet   string
	ConfRev  uint32
	TrgOps   TriggerOptions
	RptEna   bool
	Buffered bool
	SqNum    uint16
	mutex    sync.RWMutex
}

// NewURCB creates a new unbuffered report control block
func NewURCB(rptID, domain string) *URCB {
	return &URCB{
		RptID:    rptID,
		Domain:   domain,
		Buffered: false,
	}
}

// AddReport adds a report (URCB does not buffer)
func (u *URCB) AddReport(report *Report) error {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	
	if !u.RptEna {
		return fmt.Errorf("report not enabled")
	}
	
	// URCB does not buffer, just return success
	return nil
}

// ReportBuffer represents a circular buffer for reports
type ReportBuffer struct {
	reports  []*Report
	capacity int
	size     int
	head     int
	tail     int
	mutex    sync.RWMutex
}

// NewReportBuffer creates a new report buffer
func NewReportBuffer(capacity int) *ReportBuffer {
	return &ReportBuffer{
		reports:  make([]*Report, capacity),
		capacity: capacity,
		size:     0,
		head:     0,
		tail:     0,
	}
}

// Capacity returns the buffer capacity
func (b *ReportBuffer) Capacity() int {
	return b.capacity
}

// Size returns the current buffer size
func (b *ReportBuffer) Size() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.size
}

// Add adds a report to the buffer
func (b *ReportBuffer) Add(report *Report) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	b.reports[b.tail] = report
	b.tail = (b.tail + 1) % b.capacity
	
	if b.size < b.capacity {
		b.size++
	} else {
		// Buffer is full, move head
		b.head = (b.head + 1) % b.capacity
	}
	
	return nil
}

// Get retrieves a report at index
func (b *ReportBuffer) Get(index int) *Report {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	
	if index < 0 || index >= b.size {
		return nil
	}
	
	actualIndex := (b.head + index) % b.capacity
	return b.reports[actualIndex]
}

// ReportHandler handles report services
type ReportHandler struct {
	config        *Config
	brcbs         map[string]*BRCB
	urcbs         map[string]*URCB
	pending       map[string][]*Report
	mutex         sync.RWMutex
}

// NewReportHandler creates a new report handler
func NewReportHandler(config *Config) *ReportHandler {
	return &ReportHandler{
		config:  config,
		brcbs:   make(map[string]*BRCB),
		urcbs:   make(map[string]*URCB),
		pending: make(map[string][]*Report),
	}
}

// AddBRCB adds a buffered report control block
func (h *ReportHandler) AddBRCB(brcb *BRCB) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.brcbs[brcb.RptID] = brcb
	return nil
}

// AddURCB adds an unbuffered report control block
func (h *ReportHandler) AddURCB(urcb *URCB) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.urcbs[urcb.RptID] = urcb
	return nil
}

// EnableReport enables a report
func (h *ReportHandler) EnableReport(rptID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if brcb, exists := h.brcbs[rptID]; exists {
		return brcb.Enable()
	}
	
	if urcb, exists := h.urcbs[rptID]; exists {
		urcb.mutex.Lock()
		urcb.RptEna = true
		urcb.mutex.Unlock()
		return nil
	}
	
	return ErrReportNotFound
}

// DisableReport disables a report
func (h *ReportHandler) DisableReport(rptID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if brcb, exists := h.brcbs[rptID]; exists {
		return brcb.Disable()
	}
	
	if urcb, exists := h.urcbs[rptID]; exists {
		urcb.mutex.Lock()
		urcb.RptEna = false
		urcb.mutex.Unlock()
		return nil
	}
	
	return ErrReportNotFound
}

// IsReportEnabled checks if a report is enabled
func (h *ReportHandler) IsReportEnabled(rptID string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	if brcb, exists := h.brcbs[rptID]; exists {
		return brcb.IsEnabled()
	}
	
	if urcb, exists := h.urcbs[rptID]; exists {
		urcb.mutex.RLock()
		defer urcb.mutex.RUnlock()
		return urcb.RptEna
	}
	
	return false
}

// ProcessDataChange processes a data change event
func (h *ReportHandler) ProcessDataChange(change *DataChange) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	// Check BRCBs
	for rptID, brcb := range h.brcbs {
		if !brcb.IsEnabled() {
			continue
		}
		
		// Check trigger options
		if brcb.TrgOps&TriggerDataChange != 0 || brcb.TrgOps&TriggerQualityChange != 0 {
			// Generate report
			report := &Report{
				RptID:   rptID,
				SqNum:   brcb.SqNum,
				Time:    NewTimestamp(time.Now()),
				ConfRev: brcb.ConfRev,
				Values: []ReportValue{
					{
						Ref:     change.Ref,
						Value:   change.Value,
						Quality: change.Quality,
					},
				},
			}
			
			brcb.SqNum++
			
			// Add to buffer
			_ = brcb.AddReport(report)
			
			// Add to pending
			h.pending[rptID] = append(h.pending[rptID], report)
		}
	}
	
	// Check URCBs
	for rptID, urcb := range h.urcbs {
		urcb.mutex.RLock()
		if !urcb.RptEna {
			urcb.mutex.RUnlock()
			continue
		}
		
		// Check trigger options
		if urcb.TrgOps&TriggerDataChange != 0 || urcb.TrgOps&TriggerQualityChange != 0 {
			// Generate report
			report := &Report{
				RptID:   rptID,
				SqNum:   urcb.SqNum,
				Time:    NewTimestamp(time.Now()),
				ConfRev: urcb.ConfRev,
				Values: []ReportValue{
					{
						Ref:     change.Ref,
						Value:   change.Value,
						Quality: change.Quality,
					},
				},
			}
			
			urcb.SqNum++
			
			// Add to pending
			h.pending[rptID] = append(h.pending[rptID], report)
		}
		urcb.mutex.RUnlock()
	}
	
	return nil
}

// GetPendingReports returns pending reports for a report ID
func (h *ReportHandler) GetPendingReports(rptID string) []*Report {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	return h.pending[rptID]
}

// EncodeReport encodes a report
func EncodeReport(report *Report) ([]byte, error) {
	// Calculate total length
	rptIDLen := len(report.RptID)
	valuesLen := len(report.Values) * 50 // Simplified estimate
	totalLen := 2 + rptIDLen + 2 + 8 + 4 + 4 + valuesLen
	
	data := make([]byte, totalLen)
	offset := 0
	
	// RptID
	binary.BigEndian.PutUint16(data[offset:], uint16(rptIDLen))
	offset += 2
	copy(data[offset:], report.RptID)
	offset += rptIDLen
	
	// SqNum
	binary.BigEndian.PutUint16(data[offset:], report.SqNum)
	offset += 2
	
	// Time
	binary.BigEndian.PutUint32(data[offset:], report.Time.SecondsSinceEpoch)
	offset += 4
	binary.BigEndian.PutUint32(data[offset:], report.Time.FractionOfSecond)
	offset += 4
	
	// ConfRev
	binary.BigEndian.PutUint32(data[offset:], report.ConfRev)
	offset += 4
	
	// Values count
	binary.BigEndian.PutUint32(data[offset:], uint32(len(report.Values)))
	offset += 4
	
	// Values (simplified)
	for _, val := range report.Values {
		refLen := len(val.Ref)
		binary.BigEndian.PutUint16(data[offset:], uint16(refLen))
		offset += 2
		copy(data[offset:], val.Ref)
		offset += refLen
		
		// Value (as float64 for simplicity)
		if f, ok := val.Value.(float64); ok {
			binary.BigEndian.PutUint64(data[offset:], uint64(f))
		}
		offset += 8
		
		binary.BigEndian.PutUint16(data[offset:], uint16(val.Quality))
		offset += 2
	}
	
	return data, nil
}

// DecodeReport decodes a report
func DecodeReport(data []byte) (*Report, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("data too short for report")
	}
	
	report := &Report{}
	offset := 0
	
	// RptID
	rptIDLen := int(binary.BigEndian.Uint16(data[offset:]))
	offset += 2
	
	if len(data) < offset+rptIDLen {
		return nil, fmt.Errorf("data too short for RptID")
	}
	
	report.RptID = string(data[offset : offset+rptIDLen])
	offset += rptIDLen
	
	// SqNum
	report.SqNum = binary.BigEndian.Uint16(data[offset:])
	offset += 2
	
	// Time
	report.Time.SecondsSinceEpoch = binary.BigEndian.Uint32(data[offset:])
	offset += 4
	report.Time.FractionOfSecond = binary.BigEndian.Uint32(data[offset:])
	offset += 4
	
	// ConfRev
	report.ConfRev = binary.BigEndian.Uint32(data[offset:])
	offset += 4
	
	// Values count
	valuesCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	
	// Values
	report.Values = make([]ReportValue, valuesCount)
	for i := 0; i < valuesCount; i++ {
		if len(data) < offset+2 {
			break
		}
		
		refLen := int(binary.BigEndian.Uint16(data[offset:]))
		offset += 2
		
		if len(data) < offset+refLen+10 {
			break
		}
		
		report.Values[i].Ref = string(data[offset : offset+refLen])
		offset += refLen
		
		report.Values[i].Value = float64(binary.BigEndian.Uint64(data[offset:]))
		offset += 8
		
		report.Values[i].Quality = Quality(binary.BigEndian.Uint16(data[offset:]))
		offset += 2
	}
	
	return report, nil
}
