package processing

import (
	"math"
	"testing"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBasicProcessor(t *testing.T) {
	p := NewBasicProcessor()
	assert.NotNil(t, p)
}

func TestBasicProcessor_Init_Valid(t *testing.T) {
	p := NewBasicProcessor()
	err := p.Init(types.ProcessingConfig{Type: "basic"})
	require.NoError(t, err)
}

func TestBasicProcessor_Init_Invalid(t *testing.T) {
	p := NewBasicProcessor()
	err := p.Init(types.ProcessingConfig{Type: "invalid"})
	assert.Error(t, err)
}

func TestBasicProcessor_Process_Nil(t *testing.T) {
	p := NewBasicProcessor()
	result, err := p.Process(nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestBasicProcessor_Process_Empty(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{DataPoints: []*types.DataPoint{}}
	result, err := p.Process(bd)
	assert.NoError(t, err)
	assert.Equal(t, bd, result)
}

func TestBasicProcessor_Process_NegativeValue(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "power", Value: -10, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	require.NoError(t, err)
	assert.Equal(t, 0.0, result.DataPoints[0].Value)
	assert.Equal(t, true, result.DataPoints[0].Attributes["cleaned"])
}

func TestBasicProcessor_Process_TemperatureOutOfRange(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "temperature", Value: -50, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	require.NoError(t, err)
	assert.Equal(t, true, result.DataPoints[0].Attributes["cleaned"])
}

func TestBasicProcessor_Process_VoltageOutOfRange(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "voltage", Value: -5, Timestamp: time.Now()},
			{DeviceID: "SOL-001", Metric: "voltage", Value: 15000, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 2},
	}
	result, err := p.Process(bd)
	require.NoError(t, err)
	assert.Equal(t, 0.0, result.DataPoints[0].Value)
	assert.Equal(t, 0.0, result.DataPoints[1].Value)
}

func TestBasicProcessor_Process_PowerTransform(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "power", Value: 5000, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	require.NoError(t, err)
	assert.Equal(t, 5.0, result.DataPoints[0].Value)
	assert.Equal(t, "power_kw", result.DataPoints[0].Metric)
}

func TestBasicProcessor_Process_EnergyTransform(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "energy", Value: 5000, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	require.NoError(t, err)
	assert.Equal(t, 5.0, result.DataPoints[0].Value)
	assert.Equal(t, "energy_kwh", result.DataPoints[0].Metric)
}

func TestBasicProcessor_Process_CurrentTooHigh(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "current", Value: 1500, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	require.NoError(t, err)
	assert.Equal(t, 0.0, result.DataPoints[0].Value)
	assert.Equal(t, true, result.DataPoints[0].Attributes["cleaned"])
}

func TestBasicProcessor_Process_DeviceTypeInference(t *testing.T) {
	p := NewBasicProcessor()

	tests := []struct {
		deviceID   string
		deviceType string
	}{
		{"SOL-001", "solar"},
		{"WND-001", "wind"},
		{"BAT-001", "battery"},
		{"UNK-001", "unknown"},
	}

	for _, tt := range tests {
		bd := &types.BatchData{
			DataPoints: []*types.DataPoint{
				{DeviceID: tt.deviceID, Metric: "power", Value: 100, Timestamp: time.Now()},
			},
			Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
		}
		result, err := p.Process(bd)
		require.NoError(t, err)
		assert.Equal(t, tt.deviceType, result.DataPoints[0].Tags["device_type"])
	}
}

func TestBasicProcessor_Process_ExistingDeviceType(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "power", Value: 100, Timestamp: time.Now(), Tags: map[string]string{"device_type": "custom"}},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	require.NoError(t, err)
	assert.Equal(t, "custom", result.DataPoints[0].Tags["device_type"])
}

func TestBasicProcessor_CalculateDataQuality(t *testing.T) {
	p := NewBasicProcessor()

	tests := []struct {
		metric  string
		value   float64
		quality float64
	}{
		{"temperature", 25, 1.0},
		{"temperature", -50, 0.5},
		{"voltage", 220, 1.0},
		{"voltage", -10, 0.5},
		{"current", 50, 1.0},
		{"current", 2000, 0.5},
		{"power", 100, 1.0},
		{"power", -10, 1.0},
		{"energy", 100, 1.0},
		{"energy", -10, 0.5},
	}

	for _, tt := range tests {
		dp := &types.DataPoint{Metric: tt.metric, Value: tt.value, Timestamp: time.Now()}
		quality := p.calculateDataQuality(dp)
		assert.Equal(t, tt.quality, quality, "metric=%s, value=%f", tt.metric, tt.value)
	}
}

func TestBasicProcessor_CalculateDataQuality_OldTimestamp(t *testing.T) {
	p := NewBasicProcessor()
	dp := &types.DataPoint{Metric: "temperature", Value: 25, Timestamp: time.Now().Add(-48 * time.Hour)}
	quality := p.calculateDataQuality(dp)
	assert.Equal(t, 0.8, quality)
}

func TestBasicProcessor_CalculateDataQuality_FutureTimestamp(t *testing.T) {
	p := NewBasicProcessor()
	dp := &types.DataPoint{Metric: "temperature", Value: 25, Timestamp: time.Now().Add(2 * time.Hour)}
	quality := p.calculateDataQuality(dp)
	assert.Equal(t, 0.8, quality)
}

func TestBasicProcessor_CalculateDataQuality_NaN(t *testing.T) {
	p := NewBasicProcessor()
	dp := &types.DataPoint{Metric: "temperature", Value: math.NaN(), Timestamp: time.Now()}
	quality := p.calculateDataQuality(dp)
	assert.Equal(t, 0.5, quality)
}

func TestBasicProcessor_Close(t *testing.T) {
	p := NewBasicProcessor()
	err := p.Close()
	assert.NoError(t, err)
}

func TestBasicProcessor_Process_WithTagsAndAttributes(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{
				DeviceID:   "SOL-001",
				Metric:     "power",
				Value:      100,
				Timestamp:  time.Now(),
				Tags:       map[string]string{"location": "roof"},
				Attributes: map[string]interface{}{"model": "panel-v1"},
			},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	require.NoError(t, err)
	assert.Equal(t, "roof", result.DataPoints[0].Tags["location"])
	assert.Equal(t, "panel-v1", result.DataPoints[0].Attributes["model"])
}

func TestBasicProcessor_Process_NilTagsAndAttributes(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "power", Value: 100, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	result, err := p.Process(bd)
	require.NoError(t, err)
	assert.NotNil(t, result.DataPoints[0].Tags)
	assert.NotNil(t, result.DataPoints[0].Attributes)
}

func TestBasicProcessor_Process_MultiplePoints(t *testing.T) {
	p := NewBasicProcessor()
	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "SOL-001", Metric: "power", Value: 100, Timestamp: time.Now()},
			{DeviceID: "WND-001", Metric: "power", Value: 200, Timestamp: time.Now()},
			{DeviceID: "BAT-001", Metric: "voltage", Value: 3.7, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 3},
	}
	result, err := p.Process(bd)
	require.NoError(t, err)
	assert.Len(t, result.DataPoints, 3)
	assert.Equal(t, "solar", result.DataPoints[0].Tags["device_type"])
	assert.Equal(t, "wind", result.DataPoints[1].Tags["device_type"])
	assert.Equal(t, "battery", result.DataPoints[2].Tags["device_type"])
}
