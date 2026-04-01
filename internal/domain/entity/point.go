package entity

import (
	"time"
)

type PointType string

const (
	PointTypeYaoXin   PointType = "yaoxin"
	PointTypeYaoCe    PointType = "yaoc"
	PointTypeYaoKong  PointType = "yaokong"
	PointTypeSetPoint PointType = "setpoint"
	PointTypeDianDu   PointType = "diandu"
)

type Point struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code        string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"type:varchar(200);not null"`
	Type        PointType `json:"type" gorm:"type:varchar(20);not null"`
	DeviceID    string    `json:"device_id" gorm:"type:varchar(36);index"`
	StationID   string    `json:"station_id" gorm:"type:varchar(36);index"`
	
	Unit        string    `json:"unit" gorm:"type:varchar(20)"`
	Precision   int       `json:"precision"`
	MinValue    float64   `json:"min_value"`
	MaxValue    float64   `json:"max_value"`
	
	Protocol    string    `json:"protocol" gorm:"type:varchar(50)"`
	Address     int       `json:"address"`
	DataFormat  string    `json:"data_format" gorm:"type:varchar(100)"`
	
	ScanInterval int      `json:"scan_interval"`
	Deadband     float64  `json:"deadband"`
	
	IsAlarm      bool     `json:"is_alarm"`
	AlarmHigh    float64  `json:"alarm_high"`
	AlarmLow     float64  `json:"alarm_low"`
	
	Status       int      `json:"status" gorm:"default:1"`
	Description  string   `json:"description" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (p *Point) TableName() string {
	return "points"
}

func NewPoint(code, name string, pointType PointType) *Point {
	return &Point{
		Code:     code,
		Name:     name,
		Type:     pointType,
		Status:   1,
	}
}

func (p *Point) SetRange(min, max float64) {
	p.MinValue = min
	p.MaxValue = max
}

func (p *Point) SetAlarmThreshold(high, low float64) {
	p.IsAlarm = true
	p.AlarmHigh = high
	p.AlarmLow = low
}

func (p *Point) IsInRange(value float64) bool {
	if p.MinValue == 0 && p.MaxValue == 0 {
		return true
	}
	return value >= p.MinValue && value <= p.MaxValue
}

func (p *Point) CheckAlarm(value float64) (bool, string) {
	if !p.IsAlarm {
		return false, ""
	}
	
	if p.AlarmHigh > 0 && value > p.AlarmHigh {
		return true, "high"
	}
	
	if p.AlarmLow > 0 && value < p.AlarmLow {
		return true, "low"
	}
	
	return false, ""
}
