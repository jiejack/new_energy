// Package iec61850 implements IEC 61850 protocol for new energy monitoring systems
package iec61850

import (
	"errors"
	"time"
)

// Common errors
var (
	ErrNotConnected       = errors.New("not connected to IED")
	ErrInvalidParameter   = errors.New("invalid parameter")
	ErrTimeout            = errors.New("operation timeout")
	ErrAssociationFailed  = errors.New("association failed")
	ErrServiceNotSupported = errors.New("service not supported")
	ErrControlFailed      = errors.New("control operation failed")
	ErrInvalidDataModel   = errors.New("invalid data model")
	ErrReportNotFound     = errors.New("report not found")
	ErrDatasetNotFound    = errors.New("dataset not found")
)

// ConnectionState represents the state of MMS connection
type ConnectionState int

const (
	ConnectionStateDisconnected ConnectionState = iota
	ConnectionStateConnecting
	ConnectionStateConnected
	ConnectionStateAssociating
	ConnectionStateAssociated
)

func (cs ConnectionState) String() string {
	switch cs {
	case ConnectionStateDisconnected:
		return "Disconnected"
	case ConnectionStateConnecting:
		return "Connecting"
	case ConnectionStateConnected:
		return "Connected"
	case ConnectionStateAssociating:
		return "Associating"
	case ConnectionStateAssociated:
		return "Associated"
	default:
		return "Unknown"
	}
}

// ControlModel represents IEC 61850 control models
type ControlModel int

const (
	ControlModelDirectNormal ControlModel = iota
	ControlModelDirectEnhanced
	ControlModelSBONormal
	ControlModelSBOEnhanced
)

func (cm ControlModel) String() string {
	switch cm {
	case ControlModelDirectNormal:
		return "Direct with normal security"
	case ControlModelDirectEnhanced:
		return "Direct with enhanced security"
	case ControlModelSBONormal:
		return "Select-before-operate with normal security"
	case ControlModelSBOEnhanced:
		return "Select-before-operate with enhanced security"
	default:
		return "Unknown"
	}
}

// ControlState represents the state of control operation
type ControlState int

const (
	ControlStateIdle ControlState = iota
	ControlStateSelected
	ControlStateOperating
	ControlStateCompleted
	ControlStateCancelled
	ControlStateFailed
)

func (cs ControlState) String() string {
	switch cs {
	case ControlStateIdle:
		return "Idle"
	case ControlStateSelected:
		return "Selected"
	case ControlStateOperating:
		return "Operating"
	case ControlStateCompleted:
		return "Completed"
	case ControlStateCancelled:
		return "Cancelled"
	case ControlStateFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// TriggerOptions represents report trigger options
type TriggerOptions int

const (
	TriggerNone        TriggerOptions = 0
	TriggerDataChange  TriggerOptions = 1 << iota
	TriggerQualityChange
	TriggerDataUpdate
	TriggerIntegrity
	TriggerGeneralInterrogation
)

// Quality represents IEC 61850 quality flags
type Quality uint16

const (
	QualityGood           Quality = 0x0000
	QualityInvalid        Quality = 0x0001
	QualityQuestionable   Quality = 0x0002
	QualityOld            Quality = 0x0004
	QualityOverflow       Quality = 0x0008
	QualityOutOfRange     Quality = 0x0010
	QualityBad            Quality = 0x0020
	QualityBlocked        Quality = 0x0040
	QualitySubstituted    Quality = 0x0080
	QualityTest           Quality = 0x0100
	QualityOperatorBlocked Quality = 0x0200
)

func (q Quality) IsGood() bool {
	return q == QualityGood
}

func (q Quality) IsInvalid() bool {
	return q&QualityInvalid != 0
}

func (q Quality) IsQuestionable() bool {
	return q&QualityQuestionable != 0
}

// Timestamp represents IEC 61850 timestamp
type Timestamp struct {
	SecondsSinceEpoch uint32
	FractionOfSecond  uint32
	Quality           Quality
}

// NewTimestamp creates a new timestamp from time.Time
func NewTimestamp(t time.Time) Timestamp {
	// 安全转换：int64到uint32，Unix时间戳通常在有效范围内
	seconds := uint32(min(max(t.Unix(), 0), math.MaxUint32))
	nanos := t.Nanosecond()
	// Fraction of second is in units of 2^-24 seconds
	fraction := uint32(float64(nanos) * 1.0 / float64(time.Second) * 16777216.0)
	
	return Timestamp{
		SecondsSinceEpoch: seconds,
		FractionOfSecond:  fraction,
		Quality:           QualityGood,
	}
}

// ToTime converts Timestamp to time.Time
func (ts Timestamp) ToTime() time.Time {
	nanos := int64(float64(ts.FractionOfSecond) / 16777216.0 * float64(time.Second))
	return time.Unix(int64(ts.SecondsSinceEpoch), nanos)
}

// MMSMessageType represents MMS message types
type MMSMessageType int

const (
	MMSMessageTypeConfirm MMSMessageType = iota
	MMSMessageTypeRequest
	MMSMessageTypeResponse
	MMSMessageTypeError
)

// MMSResult represents MMS operation result
type MMSResult int

const (
	MMSResultSuccess MMSResult = iota
	MMSResultAccessDenied
	MMSResultObjectNonExistent
	MMSResultParameterError
	MMSResultServiceError
)

// Config represents IEC 61850 client configuration
type Config struct {
	// Connection settings
	IEDName      string
	IPAddress    string
	Port         int
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	
	// Association settings
	AuthMechanism string
	AuthParameter string
	
	// Report settings
	ReportBufferSize int
	ReportInterval   time.Duration
	
	// Control settings
	ControlTimeout   time.Duration
	SelectTimeout    time.Duration
	
	// SV settings
	SVAppID        uint16
	SVBufferSize   int
	SVSamplingRate int
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Port:             102,
		ConnectTimeout:   10 * time.Second,
		ReadTimeout:      5 * time.Second,
		WriteTimeout:     5 * time.Second,
		ReportBufferSize: 256,
		ReportInterval:   1 * time.Second,
		ControlTimeout:   5 * time.Second,
		SelectTimeout:    10 * time.Second,
		SVBufferSize:     1024,
		SVSamplingRate:   4000,
	}
}
