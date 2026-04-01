package entity

import (
	"time"
)

type DeviceType string
type DeviceStatus int

const (
	DeviceTypeInverter    DeviceType = "inverter"
	DeviceTypeMeter       DeviceType = "meter"
	DeviceTypeTransformer DeviceType = "transformer"
	DeviceTypeSwitch      DeviceType = "switch"
	DeviceTypeWeather     DeviceType = "weather"
	DeviceTypeESS         DeviceType = "ess"
	DeviceTypePCS         DeviceType = "pcs"
	DeviceTypeBMS         DeviceType = "bms"
)

const (
	DeviceStatusOffline    DeviceStatus = 0
	DeviceStatusOnline     DeviceStatus = 1
	DeviceStatusFault      DeviceStatus = 2
	DeviceStatusMaintain   DeviceStatus = 3
)

type Device struct {
	ID           string       `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code         string       `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name         string       `json:"name" gorm:"type:varchar(200);not null"`
	Type         DeviceType   `json:"type" gorm:"type:varchar(50);not null"`
	StationID    string       `json:"station_id" gorm:"type:varchar(36);index;not null"`
	
	Manufacturer string       `json:"manufacturer" gorm:"type:varchar(100)"`
	Model        string       `json:"model" gorm:"type:varchar(100)"`
	SerialNumber string       `json:"serial_number" gorm:"type:varchar(100)"`
	
	RatedPower   float64      `json:"rated_power"`
	RatedVoltage float64      `json:"rated_voltage"`
	RatedCurrent float64      `json:"rated_current"`
	
	Protocol     string       `json:"protocol" gorm:"type:varchar(50)"`
	IPAddress    string       `json:"ip_address" gorm:"type:varchar(50)"`
	Port         int          `json:"port"`
	SlaveID      int          `json:"slave_id"`
	
	Status       DeviceStatus `json:"status" gorm:"default:0"`
	LastOnline   *time.Time   `json:"last_online"`
	
	InstallDate  *time.Time   `json:"install_date"`
	WarrantyDate *time.Time   `json:"warranty_date"`
	
	Description  string       `json:"description" gorm:"type:text"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	
	Points       []*Point     `json:"points" gorm:"foreignKey:DeviceID"`
}

func (d *Device) TableName() string {
	return "devices"
}

func NewDevice(code, name string, deviceType DeviceType, stationID string) *Device {
	return &Device{
		Code:      code,
		Name:      name,
		Type:      deviceType,
		StationID: stationID,
		Status:    DeviceStatusOffline,
	}
}

func (d *Device) SetOnline() {
	now := time.Now()
	d.Status = DeviceStatusOnline
	d.LastOnline = &now
}

func (d *Device) SetOffline() {
	d.Status = DeviceStatusOffline
}

func (d *Device) SetFault() {
	d.Status = DeviceStatusFault
}

func (d *Device) SetMaintain() {
	d.Status = DeviceStatusMaintain
}

func (d *Device) IsOnline() bool {
	return d.Status == DeviceStatusOnline
}

func (d *Device) SetCommunication(protocol, ip string, port, slaveID int) {
	d.Protocol = protocol
	d.IPAddress = ip
	d.Port = port
	d.SlaveID = slaveID
}
