package entity

import (
	"time"
)

type StationType string
type StationStatus int

const (
	StationTypePV      StationType = "pv"
	StationTypeWind    StationType = "wind"
	StationTypeESS     StationType = "ess"
	StationTypeHybrid  StationType = "hybrid"
	StationTypeSubstation StationType = "substation"
)

const (
	StationStatusInactive StationStatus = 0
	StationStatusActive   StationStatus = 1
	StationStatusFault    StationStatus = 2
)

type Station struct {
	ID           string        `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code         string        `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name         string        `json:"name" gorm:"type:varchar(200);not null"`
	Type         StationType   `json:"type" gorm:"type:varchar(50);not null"`
	SubRegionID  string        `json:"sub_region_id" gorm:"type:varchar(36);index;not null"`
	
	Capacity     float64       `json:"capacity"`
	VoltageLevel string        `json:"voltage_level" gorm:"type:varchar(50)"`
	
	Longitude    float64       `json:"longitude"`
	Latitude     float64       `json:"latitude"`
	Address      string        `json:"address" gorm:"type:varchar(500)"`
	
	Status       StationStatus `json:"status" gorm:"default:1"`
	CommissionDate *time.Time  `json:"commission_date"`
	
	Description  string        `json:"description" gorm:"type:text"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	
	Devices      []*Device     `json:"devices" gorm:"foreignKey:StationID"`
}

func (s *Station) TableName() string {
	return "stations"
}

func NewStation(code, name string, stationType StationType, subRegionID string) *Station {
	return &Station{
		Code:        code,
		Name:        name,
		Type:        stationType,
		SubRegionID: subRegionID,
		Status:      StationStatusActive,
	}
}

func (s *Station) SetLocation(longitude, latitude float64, address string) {
	s.Longitude = longitude
	s.Latitude = latitude
	s.Address = address
}

func (s *Station) SetCapacity(capacity float64, voltageLevel string) {
	s.Capacity = capacity
	s.VoltageLevel = voltageLevel
}

func (s *Station) Activate() {
	s.Status = StationStatusActive
}

func (s *Station) Deactivate() {
	s.Status = StationStatusInactive
}

func (s *Station) IsActive() bool {
	return s.Status == StationStatusActive
}
