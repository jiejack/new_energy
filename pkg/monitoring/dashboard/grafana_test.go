package dashboard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDashboardConfig(t *testing.T) {
	dc := NewDashboardConfig("Test Dashboard")
	require.NotNil(t, dc)

	d := dc.Build()
	assert.Equal(t, "Test Dashboard", d.Title)
	assert.Equal(t, "browser", d.Timezone)
	assert.Equal(t, 38, d.SchemaVersion)
	assert.Equal(t, "30s", d.Refresh)
	assert.True(t, d.Editable)
	assert.Equal(t, 1, d.GraphTooltip)
	assert.Contains(t, d.Tags, "new-energy-monitoring")
}

func TestDashboardConfig_SetUID(t *testing.T) {
	dc := NewDashboardConfig("Test").SetUID("test-uid")
	d := dc.Build()
	assert.Equal(t, "test-uid", d.UID)
}

func TestDashboardConfig_SetDescription(t *testing.T) {
	dc := NewDashboardConfig("Test").SetDescription("A test dashboard")
	d := dc.Build()
	assert.Equal(t, "A test dashboard", d.Description)
}

func TestDashboardConfig_SetRefresh(t *testing.T) {
	dc := NewDashboardConfig("Test").SetRefresh("10s")
	d := dc.Build()
	assert.Equal(t, "10s", d.Refresh)
}

func TestDashboardConfig_SetTimeRange(t *testing.T) {
	dc := NewDashboardConfig("Test").SetTimeRange("now-6h", "now")
	d := dc.Build()
	assert.Equal(t, "now-6h", d.Time.From)
	assert.Equal(t, "now", d.Time.To)
}

func TestDashboardConfig_AddTag(t *testing.T) {
	dc := NewDashboardConfig("Test").AddTag("custom-tag")
	d := dc.Build()
	assert.Contains(t, d.Tags, "custom-tag")
}

func TestDashboardConfig_AddDataSource(t *testing.T) {
	ds := DataSource{Name: "Prometheus", Type: "prometheus", URL: "http://localhost:9090"}
	dc := NewDashboardConfig("Test").AddDataSource(ds)
	assert.Equal(t, 1, len(dc.GetDataSources()))
	assert.Equal(t, "Prometheus", dc.GetDataSources()[0].Name)
}

func TestDashboardConfig_AddTemplateVariable(t *testing.T) {
	tv := TemplateVariable{
		Name:       "station",
		Type:       "query",
		DataSource: "Prometheus",
		Query:      "label_values(station)",
		Refresh:    1,
		IncludeAll: true,
		Multi:      true,
	}
	dc := NewDashboardConfig("Test").AddTemplateVariable(tv)
	d := dc.Build()
	assert.Equal(t, 1, len(d.Templating.List))
	assert.Equal(t, "station", d.Templating.List[0].Name)
}

func TestDashboardConfig_AddAnnotation(t *testing.T) {
	ann := Annotation{
		Name:       "Deployments",
		DataSource: "Prometheus",
		Enable:     true,
		IconColor:  "green",
	}
	dc := NewDashboardConfig("Test").AddAnnotation(ann)
	d := dc.Build()
	assert.Equal(t, 1, len(d.Annotations.List))
	assert.Equal(t, "Deployments", d.Annotations.List[0].Name)
}

func TestDashboardConfig_AddPanel(t *testing.T) {
	panel := NewTimeSeriesPanel("CPU Usage", 0, 0, 12, 6)
	dc := NewDashboardConfig("Test").AddPanel(panel)
	d := dc.Build()
	assert.Equal(t, 1, len(d.Panels))
	assert.Equal(t, "CPU Usage", d.Panels[0].Title)
	assert.Equal(t, 1, d.Panels[0].ID)
}

func TestDashboardConfig_AddAlertRule(t *testing.T) {
	rule := NewAlertRule("High CPU", "CPU is too high", 5*time.Minute)
	dc := NewDashboardConfig("Test").AddAlertRule(rule)
	assert.Equal(t, 1, len(dc.GetAlertRules()))
	assert.Equal(t, "High CPU", dc.GetAlertRules()[0].Name)
}

func TestDashboardConfig_Chaining(t *testing.T) {
	dc := NewDashboardConfig("Chained").
		SetUID("chained-uid").
		SetDescription("Chained config").
		SetRefresh("5s").
		SetTimeRange("now-3h", "now").
		AddTag("chain1").
		AddTag("chain2")

	d := dc.Build()
	assert.Equal(t, "chained-uid", d.UID)
	assert.Equal(t, "Chained config", d.Description)
	assert.Equal(t, "5s", d.Refresh)
	assert.Equal(t, 3, len(d.Tags))
}

func TestNewTimeSeriesPanel(t *testing.T) {
	p := NewTimeSeriesPanel("Test TS", 0, 0, 12, 8)
	assert.Equal(t, "Test TS", p.Title)
	assert.Equal(t, "timeseries", p.Type)
	assert.Equal(t, 0, p.GridPos.X)
	assert.Equal(t, 0, p.GridPos.Y)
	assert.Equal(t, 12, p.GridPos.W)
	assert.Equal(t, 8, p.GridPos.H)
	assert.NotNil(t, p.Options)
	assert.NotNil(t, p.FieldConfig)
}

func TestNewStatPanel(t *testing.T) {
	p := NewStatPanel("Test Stat", 6, 0, 6, 4)
	assert.Equal(t, "Test Stat", p.Title)
	assert.Equal(t, "stat", p.Type)
	assert.Equal(t, 6, p.GridPos.X)
}

func TestNewGaugePanel(t *testing.T) {
	p := NewGaugePanel("Test Gauge", 0, 0, 6, 6)
	assert.Equal(t, "Test Gauge", p.Title)
	assert.Equal(t, "gauge", p.Type)
}

func TestNewTablePanel(t *testing.T) {
	p := NewTablePanel("Test Table", 0, 0, 24, 8)
	assert.Equal(t, "Test Table", p.Title)
	assert.Equal(t, "table", p.Type)
}

func TestNewHeatmapPanel(t *testing.T) {
	p := NewHeatmapPanel("Test Heatmap", 0, 0, 12, 8)
	assert.Equal(t, "Test Heatmap", p.Title)
	assert.Equal(t, "heatmap", p.Type)
}

func TestPanel_AddTarget(t *testing.T) {
	p := NewTimeSeriesPanel("Test", 0, 0, 12, 8)
	p.AddTarget(`rate(http_requests_total[5m])`, "Requests/s")
	assert.Equal(t, 1, len(p.Targets))
	assert.Equal(t, "A", p.Targets[0].RefID)
	assert.Equal(t, `rate(http_requests_total[5m])`, p.Targets[0].Expr)
	assert.Equal(t, "Requests/s", p.Targets[0].LegendFormat)

	p.AddTarget(`rate(http_errors_total[5m])`, "Errors/s")
	assert.Equal(t, 2, len(p.Targets))
	assert.Equal(t, "B", p.Targets[1].RefID)
}

func TestPanel_SetUnit(t *testing.T) {
	p := NewStatPanel("Test", 0, 0, 6, 4)
	p.SetUnit("watt")
	assert.Equal(t, "watt", p.FieldConfig.Defaults.Unit)
}

func TestPanel_SetDecimals(t *testing.T) {
	p := NewStatPanel("Test", 0, 0, 6, 4)
	p.SetDecimals(2)
	assert.Equal(t, 2, p.FieldConfig.Defaults.Decimals)
}

func TestPanel_SetThresholds(t *testing.T) {
	p := NewStatPanel("Test", 0, 0, 6, 4)
	steps := []ThresholdStep{
		{Color: "green", Value: 0},
		{Color: "red", Value: 10},
	}
	p.SetThresholds(steps)
	assert.Equal(t, "absolute", p.FieldConfig.Defaults.Thresholds.Mode)
	assert.Equal(t, 2, len(p.FieldConfig.Defaults.Thresholds.Steps))
}

func TestPanel_SetMin(t *testing.T) {
	p := NewGaugePanel("Test", 0, 0, 6, 6)
	p.SetMin(0)
	assert.NotNil(t, p.FieldConfig.Defaults.Min)
	assert.Equal(t, float64(0), *p.FieldConfig.Defaults.Min)
}

func TestPanel_SetMax(t *testing.T) {
	p := NewGaugePanel("Test", 0, 0, 6, 6)
	p.SetMax(100)
	assert.NotNil(t, p.FieldConfig.Defaults.Max)
	assert.Equal(t, float64(100), *p.FieldConfig.Defaults.Max)
}

func TestPanel_SetUnit_NoFieldConfig(t *testing.T) {
	p := &Panel{Title: "Test"}
	p.SetUnit("bytes")
	assert.NotNil(t, p.FieldConfig)
	assert.Equal(t, "bytes", p.FieldConfig.Defaults.Unit)
}

func TestNewPrometheusDataSource(t *testing.T) {
	ds := NewPrometheusDataSource("Prom", "http://localhost:9090", true)
	assert.Equal(t, "Prom", ds.Name)
	assert.Equal(t, "prometheus", ds.Type)
	assert.Equal(t, "http://localhost:9090", ds.URL)
	assert.Equal(t, "proxy", ds.Access)
	assert.True(t, ds.IsDefault)
	assert.Equal(t, "POST", ds.JSONData.HTTPMethod)
}

func TestNewJaegerDataSource(t *testing.T) {
	ds := NewJaegerDataSource("Jaeger", "http://localhost:16686")
	assert.Equal(t, "Jaeger", ds.Name)
	assert.Equal(t, "jaeger", ds.Type)
	assert.Equal(t, "GET", ds.JSONData.HTTPMethod)
}

func TestNewLokiDataSource(t *testing.T) {
	ds := NewLokiDataSource("Loki", "http://localhost:3100")
	assert.Equal(t, "Loki", ds.Name)
	assert.Equal(t, "loki", ds.Type)
	assert.Equal(t, 1000, ds.JSONData.MaxDataPoints)
}

func TestNewAlertRule(t *testing.T) {
	rule := NewAlertRule("Test Alert", "Alert message", 5*time.Minute)
	assert.Equal(t, "Test Alert", rule.Name)
	assert.Equal(t, "Alert message", rule.Message)
	assert.Equal(t, "5m0s", rule.Frequency)
	assert.Equal(t, "alerting", rule.ExecutionErrorState)
	assert.Equal(t, "no_data", rule.NoDataState)
	assert.Equal(t, 1, rule.Handler)
}

func TestAlertRule_AddCondition(t *testing.T) {
	rule := NewAlertRule("Test", "msg", time.Minute)
	rule.AddCondition("A", "gt", []float64{90})
	assert.Equal(t, 1, len(rule.Conditions))
	assert.Equal(t, "gt", rule.Conditions[0].Evaluator.Type)
	assert.Equal(t, []float64{90}, rule.Conditions[0].Evaluator.Params)
	assert.Equal(t, "and", rule.Conditions[0].Operator.Type)
	assert.Equal(t, "query", rule.Conditions[0].Type)
}

func TestAlertRule_AddNotification(t *testing.T) {
	rule := NewAlertRule("Test", "msg", time.Minute)
	rule.AddNotification("slack-channel")
	assert.Equal(t, 1, len(rule.Notifications))
	assert.Equal(t, "slack-channel", rule.Notifications[0].UID)
}

func TestDashboard_ToJSON(t *testing.T) {
	d := NewDashboardConfig("Test").Build()
	data, err := d.ToJSON()
	require.NoError(t, err)
	assert.True(t, len(data) > 0)

	var parsed Dashboard
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "Test", parsed.Title)
}

func TestDashboard_ToJSONString(t *testing.T) {
	d := NewDashboardConfig("Test").Build()
	str, err := d.ToJSONString()
	require.NoError(t, err)
	assert.Contains(t, str, "Test")
}

func TestDashboard_SaveToFile_LoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_dashboard.json")

	d := NewDashboardConfig("Test Save").SetUID("test-save").Build()
	err := d.SaveToFile(filename)
	require.NoError(t, err)

	_, err = os.Stat(filename)
	require.NoError(t, err)

	loaded, err := LoadFromFile(filename)
	require.NoError(t, err)
	assert.Equal(t, "Test Save", loaded.Title)
	assert.Equal(t, "test-save", loaded.UID)
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/file.json")
	assert.Error(t, err)
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(filename, []byte("not json"), 0644)
	_, err := LoadFromFile(filename)
	assert.Error(t, err)
}

func TestDashboardManager(t *testing.T) {
	dm := NewDashboardManager()
	require.NotNil(t, dm)

	d1 := NewDashboardConfig("Dashboard 1").SetUID("dash-1").Build()
	d2 := NewDashboardConfig("Dashboard 2").SetUID("dash-2").Build()

	dm.AddDashboard(d1)
	dm.AddDashboard(d2)

	got, ok := dm.GetDashboard("dash-1")
	assert.True(t, ok)
	assert.Equal(t, "Dashboard 1", got.Title)

	_, ok = dm.GetDashboard("nonexistent")
	assert.False(t, ok)

	list := dm.ListDashboards()
	assert.Equal(t, 2, len(list))

	dm.RemoveDashboard("dash-1")
	_, ok = dm.GetDashboard("dash-1")
	assert.False(t, ok)
}

func TestDashboardManager_DataSources(t *testing.T) {
	dm := NewDashboardManager()
	ds := NewPrometheusDataSource("Prom", "http://localhost:9090", true)
	dm.AddDataSource(ds)

	sources := dm.GetDataSources()
	assert.Equal(t, 1, len(sources))
}

func TestDashboardManager_ExportAll(t *testing.T) {
	tmpDir := t.TempDir()
	dm := NewDashboardManager()
	d := NewDashboardConfig("Export Test").SetUID("export-test").Build()
	dm.AddDashboard(d)

	err := dm.ExportAll(tmpDir)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(tmpDir, "export-test.json"))
	require.NoError(t, err)
}

func TestDashboardManager_ExportAll_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	dm := NewDashboardManager()
	err := dm.ExportAll(tmpDir)
	require.NoError(t, err)
}

func TestDashboardManager_ExportDataSources(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "datasources.json")
	dm := NewDashboardManager()
	dm.AddDataSource(NewPrometheusDataSource("Prom", "http://localhost:9090", true))

	err := dm.ExportDataSources(filename)
	require.NoError(t, err)

	_, err = os.Stat(filename)
	require.NoError(t, err)
}

func TestNewEnergyMonitoringDashboard(t *testing.T) {
	d := NewEnergyMonitoringDashboard()
	assert.Equal(t, "New Energy Monitoring Dashboard", d.Title)
	assert.Equal(t, "new-energy-monitoring", d.UID)
	assert.True(t, len(d.Panels) > 0)
	assert.Equal(t, 1, len(d.Templating.List))
}

func TestNewSystemOverviewDashboard(t *testing.T) {
	d := NewSystemOverviewDashboard()
	assert.Equal(t, "System Overview", d.Title)
	assert.Equal(t, "system-overview", d.UID)
	assert.True(t, len(d.Panels) > 0)
}

func TestNewAlertDashboard(t *testing.T) {
	d := NewAlertDashboard()
	assert.Equal(t, "Alert Dashboard", d.Title)
	assert.Equal(t, "alert-dashboard", d.UID)
	assert.True(t, len(d.Panels) > 0)
}
