package visualization

import (
	"testing"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBasicVisualizer(t *testing.T) {
	v := NewBasicVisualizer()
	assert.NotNil(t, v)
}

func TestBasicVisualizer_Init_Valid(t *testing.T) {
	v := NewBasicVisualizer()
	err := v.Init(types.VisualizationConfig{Type: "basic"})
	require.NoError(t, err)
}

func TestBasicVisualizer_Init_Invalid(t *testing.T) {
	v := NewBasicVisualizer()
	err := v.Init(types.VisualizationConfig{Type: "invalid"})
	assert.Error(t, err)
}

func TestBasicVisualizer_CreateDashboard(t *testing.T) {
	v := NewBasicVisualizer()
	err := v.Init(types.VisualizationConfig{Type: "basic"})
	require.NoError(t, err)

	panels := []types.Panel{
		{ID: "panel-1", Title: "Test", Type: "graph"},
		{Title: "No ID", Type: "gauge"},
	}
	err = v.CreateDashboard("dash-1", panels)
	assert.NoError(t, err)
	assert.Equal(t, "panel-1", panels[0].ID)
	assert.NotEmpty(t, panels[1].ID)
}

func TestBasicVisualizer_CreateDashboard_EmptyName(t *testing.T) {
	v := NewBasicVisualizer()
	err := v.CreateDashboard("", []types.Panel{})
	assert.Error(t, err)
}

func TestBasicVisualizer_GetDashboard(t *testing.T) {
	v := NewBasicVisualizer()
	v.CreateDashboard("dash-1", []types.Panel{{ID: "p1", Title: "Test", Type: "graph"}})

	panels, err := v.GetDashboard("dash-1")
	assert.NoError(t, err)
	assert.Len(t, panels, 1)

	_, err = v.GetDashboard("nonexistent")
	assert.Error(t, err)
}

func TestBasicVisualizer_UpdatePanel(t *testing.T) {
	v := NewBasicVisualizer()
	v.CreateDashboard("dash-1", []types.Panel{{ID: "p1", Title: "Test", Type: "graph"}})

	err := v.UpdatePanel("dash-1", "p1", map[string]interface{}{"value": 42})
	assert.NoError(t, err)

	err = v.UpdatePanel("nonexistent", "p1", nil)
	assert.Error(t, err)

	err = v.UpdatePanel("dash-1", "nonexistent", nil)
	assert.Error(t, err)
}

func TestBasicVisualizer_ListDashboards(t *testing.T) {
	v := NewBasicVisualizer()
	v.CreateDashboard("dash-1", []types.Panel{})
	v.CreateDashboard("dash-2", []types.Panel{})

	dashboards := v.ListDashboards()
	assert.Len(t, dashboards, 2)
	assert.Contains(t, dashboards, "dash-1")
	assert.Contains(t, dashboards, "dash-2")
}

func TestBasicVisualizer_Close(t *testing.T) {
	v := NewBasicVisualizer()
	err := v.Close()
	assert.NoError(t, err)
}

func TestBasicVisualizer_GenerateTimeSeriesChart(t *testing.T) {
	v := NewBasicVisualizer()
	dataPoints := []*types.DataPoint{
		{Value: 10.0, Timestamp: time.Now()},
		{Value: 20.0, Timestamp: time.Now().Add(1 * time.Minute)},
	}
	chart := v.GenerateTimeSeriesChart(dataPoints)
	assert.NotNil(t, chart)
	assert.Equal(t, "time_series", chart["type"])
	assert.NotNil(t, chart["timestamps"])
	assert.NotNil(t, chart["values"])
}

func TestBasicVisualizer_GenerateGaugeChart(t *testing.T) {
	v := NewBasicVisualizer()
	chart := v.GenerateGaugeChart(75.0, 0, 100, "CPU Usage")
	assert.NotNil(t, chart)
	assert.Equal(t, "gauge", chart["type"])
	assert.Equal(t, 75.0, chart["value"])
	assert.Equal(t, 0.0, chart["min"])
	assert.Equal(t, 100.0, chart["max"])
	assert.Equal(t, "CPU Usage", chart["label"])
}

func TestBasicVisualizer_GenerateBarChart(t *testing.T) {
	v := NewBasicVisualizer()
	chart := v.GenerateBarChart([]string{"A", "B", "C"}, []float64{10, 20, 30})
	assert.NotNil(t, chart)
	assert.Equal(t, "bar", chart["type"])
}

func TestBasicVisualizer_GeneratePieChart(t *testing.T) {
	v := NewBasicVisualizer()
	chart := v.GeneratePieChart([]string{"A", "B"}, []float64{60, 40})
	assert.NotNil(t, chart)
	assert.Equal(t, "pie", chart["type"])
}
