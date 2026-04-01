package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDevice(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		deviceName string
		deviceType DeviceType
		stationID  string
		want       *Device
	}{
		{
			name:       "创建逆变器设备",
			code:       "INV001",
			deviceName: "1号逆变器",
			deviceType: DeviceTypeInverter,
			stationID:  "station001",
			want: &Device{
				Code:      "INV001",
				Name:      "1号逆变器",
				Type:      DeviceTypeInverter,
				StationID: "station001",
				Status:    DeviceStatusOffline,
			},
		},
		{
			name:       "创建电表设备",
			code:       "METER001",
			deviceName: "总电表",
			deviceType: DeviceTypeMeter,
			stationID:  "station002",
			want: &Device{
				Code:      "METER001",
				Name:      "总电表",
				Type:      DeviceTypeMeter,
				StationID: "station002",
				Status:    DeviceStatusOffline,
			},
		},
		{
			name:       "创建储能设备",
			code:       "ESS001",
			deviceName: "储能系统1",
			deviceType: DeviceTypeESS,
			stationID:  "station003",
			want: &Device{
				Code:      "ESS001",
				Name:      "储能系统1",
				Type:      DeviceTypeESS,
				StationID: "station003",
				Status:    DeviceStatusOffline,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDevice(tt.code, tt.deviceName, tt.deviceType, tt.stationID)
			assert.Equal(t, tt.want.Code, got.Code)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Type, got.Type)
			assert.Equal(t, tt.want.StationID, got.StationID)
			assert.Equal(t, tt.want.Status, got.Status)
		})
	}
}

func TestDevice_SetOnline(t *testing.T) {
	device := NewDevice("INV001", "逆变器", DeviceTypeInverter, "station001")
	assert.Equal(t, DeviceStatusOffline, device.Status)
	assert.Nil(t, device.LastOnline)

	device.SetOnline()
	assert.Equal(t, DeviceStatusOnline, device.Status)
	assert.NotNil(t, device.LastOnline)
	assert.True(t, time.Since(*device.LastOnline) < time.Second)
}

func TestDevice_SetOffline(t *testing.T) {
	device := NewDevice("INV001", "逆变器", DeviceTypeInverter, "station001")
	device.SetOnline()

	device.SetOffline()
	assert.Equal(t, DeviceStatusOffline, device.Status)
}

func TestDevice_SetFault(t *testing.T) {
	device := NewDevice("INV001", "逆变器", DeviceTypeInverter, "station001")

	device.SetFault()
	assert.Equal(t, DeviceStatusFault, device.Status)
}

func TestDevice_SetMaintain(t *testing.T) {
	device := NewDevice("INV001", "逆变器", DeviceTypeInverter, "station001")

	device.SetMaintain()
	assert.Equal(t, DeviceStatusMaintain, device.Status)
}

func TestDevice_IsOnline(t *testing.T) {
	device := NewDevice("INV001", "逆变器", DeviceTypeInverter, "station001")

	assert.False(t, device.IsOnline())

	device.SetOnline()
	assert.True(t, device.IsOnline())

	device.SetOffline()
	assert.False(t, device.IsOnline())

	device.SetFault()
	assert.False(t, device.IsOnline())
}

func TestDevice_SetCommunication(t *testing.T) {
	device := NewDevice("INV001", "逆变器", DeviceTypeInverter, "station001")

	device.SetCommunication("modbus", "192.168.1.100", 502, 1)

	assert.Equal(t, "modbus", device.Protocol)
	assert.Equal(t, "192.168.1.100", device.IPAddress)
	assert.Equal(t, 502, device.Port)
	assert.Equal(t, 1, device.SlaveID)
}

func TestDevice_TableName(t *testing.T) {
	device := Device{}
	assert.Equal(t, "devices", device.TableName())
}

func TestDeviceType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		typeVal  DeviceType
		expected string
	}{
		{"逆变器", DeviceTypeInverter, "inverter"},
		{"电表", DeviceTypeMeter, "meter"},
		{"变压器", DeviceTypeTransformer, "transformer"},
		{"开关", DeviceTypeSwitch, "switch"},
		{"气象站", DeviceTypeWeather, "weather"},
		{"储能系统", DeviceTypeESS, "ess"},
		{"PCS", DeviceTypePCS, "pcs"},
		{"BMS", DeviceTypeBMS, "bms"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, DeviceType(tt.expected), tt.typeVal)
		})
	}
}

func TestDeviceStatus_Constants(t *testing.T) {
	assert.Equal(t, DeviceStatus(0), DeviceStatusOffline)
	assert.Equal(t, DeviceStatus(1), DeviceStatusOnline)
	assert.Equal(t, DeviceStatus(2), DeviceStatusFault)
	assert.Equal(t, DeviceStatus(3), DeviceStatusMaintain)
}
