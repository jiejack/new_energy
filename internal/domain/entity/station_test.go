package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStation(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		stationName string
		stationType StationType
		subRegionID string
		want        *Station
	}{
		{
			name:        "创建光伏电站",
			code:        "PV001",
			stationName: "光伏电站1",
			stationType: StationTypePV,
			subRegionID: "region001",
			want: &Station{
				Code:        "PV001",
				Name:        "光伏电站1",
				Type:        StationTypePV,
				SubRegionID: "region001",
				Status:      StationStatusActive,
			},
		},
		{
			name:        "创建风电场",
			code:        "WIND001",
			stationName: "风电场1",
			stationType: StationTypeWind,
			subRegionID: "region002",
			want: &Station{
				Code:        "WIND001",
				Name:        "风电场1",
				Type:        StationTypeWind,
				SubRegionID: "region002",
				Status:      StationStatusActive,
			},
		},
		{
			name:        "创建储能电站",
			code:        "ESS001",
			stationName: "储能电站1",
			stationType: StationTypeESS,
			subRegionID: "region003",
			want: &Station{
				Code:        "ESS001",
				Name:        "储能电站1",
				Type:        StationTypeESS,
				SubRegionID: "region003",
				Status:      StationStatusActive,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStation(tt.code, tt.stationName, tt.stationType, tt.subRegionID)
			assert.Equal(t, tt.want.Code, got.Code)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Type, got.Type)
			assert.Equal(t, tt.want.SubRegionID, got.SubRegionID)
			assert.Equal(t, tt.want.Status, got.Status)
		})
	}
}

func TestStation_SetLocation(t *testing.T) {
	station := NewStation("PV001", "光伏电站", StationTypePV, "region001")

	station.SetLocation(116.4074, 39.9042, "北京市东城区")

	assert.Equal(t, 116.4074, station.Longitude)
	assert.Equal(t, 39.9042, station.Latitude)
	assert.Equal(t, "北京市东城区", station.Address)
}

func TestStation_SetCapacity(t *testing.T) {
	station := NewStation("PV001", "光伏电站", StationTypePV, "region001")

	station.SetCapacity(100.5, "10kV")

	assert.Equal(t, 100.5, station.Capacity)
	assert.Equal(t, "10kV", station.VoltageLevel)
}

func TestStation_Activate(t *testing.T) {
	station := NewStation("PV001", "光伏电站", StationTypePV, "region001")
	station.Status = StationStatusInactive

	station.Activate()
	assert.Equal(t, StationStatusActive, station.Status)
}

func TestStation_Deactivate(t *testing.T) {
	station := NewStation("PV001", "光伏电站", StationTypePV, "region001")
	assert.Equal(t, StationStatusActive, station.Status)

	station.Deactivate()
	assert.Equal(t, StationStatusInactive, station.Status)
}

func TestStation_IsActive(t *testing.T) {
	station := NewStation("PV001", "光伏电站", StationTypePV, "region001")
	assert.True(t, station.IsActive())

	station.Deactivate()
	assert.False(t, station.IsActive())

	station.Activate()
	assert.True(t, station.IsActive())

	station.Status = StationStatusFault
	assert.False(t, station.IsActive())
}

func TestStation_TableName(t *testing.T) {
	station := Station{}
	assert.Equal(t, "stations", station.TableName())
}

func TestStationType_Constants(t *testing.T) {
	assert.Equal(t, StationType("pv"), StationTypePV)
	assert.Equal(t, StationType("wind"), StationTypeWind)
	assert.Equal(t, StationType("ess"), StationTypeESS)
	assert.Equal(t, StationType("hybrid"), StationTypeHybrid)
	assert.Equal(t, StationType("substation"), StationTypeSubstation)
}

func TestStationStatus_Constants(t *testing.T) {
	assert.Equal(t, StationStatus(0), StationStatusInactive)
	assert.Equal(t, StationStatus(1), StationStatusActive)
	assert.Equal(t, StationStatus(2), StationStatusFault)
}
