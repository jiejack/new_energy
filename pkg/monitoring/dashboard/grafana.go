package dashboard

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Dashboard Grafana仪表盘配置
type Dashboard struct {
	ID            int          `json:"id,omitempty"`
	UID           string       `json:"uid,omitempty"`
	Title         string       `json:"title"`
	Tags          []string     `json:"tags,omitempty"`
	Timezone      string       `json:"timezone"`
	SchemaVersion int          `json:"schemaVersion"`
	Version       int          `json:"version,omitempty"`
	Refresh       string       `json:"refresh,omitempty"`
	Time          TimeRange    `json:"time"`
	Panels        []Panel      `json:"panels"`
	Templating    Templating   `json:"templating,omitempty"`
	Annotations   Annotations  `json:"annotations,omitempty"`
	Editable      bool         `json:"editable"`
	GraphTooltip  int          `json:"graphTooltip"`
	Description   string       `json:"description,omitempty"`
}

// TimeRange 时间范围
type TimeRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Templating 模板变量
type Templating struct {
	List []TemplateVariable `json:"list"`
}

// TemplateVariable 模板变量
type TemplateVariable struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	DataSource  string   `json:"datasource,omitempty"`
	Query       string   `json:"query,omitempty"`
	Refresh     int      `json:"refresh"`
	IncludeAll  bool     `json:"includeAll"`
	AllValue    string   `json:"allValue,omitempty"`
	Multi       bool     `json:"multi"`
	Sort        int      `json:"sort"`
	Label       string   `json:"label,omitempty"`
	Hide        int      `json:"hide"`
	Options     []Option `json:"options,omitempty"`
	Definition  string   `json:"definition,omitempty"`
}

// Option 选项
type Option struct {
	Text     string `json:"text"`
	Value    string `json:"value"`
	Selected bool   `json:"selected"`
}

// Annotations 注解配置
type Annotations struct {
	List []Annotation `json:"list"`
}

// Annotation 注解
type Annotation struct {
	Name        string `json:"name"`
	DataSource  string `json:"datasource"`
	Enable      bool   `json:"enable"`
	IconColor   string `json:"iconColor"`
	Query       string `json:"query,omitempty"`
	Step        string `json:"step,omitempty"`
	TitleFormat string `json:"titleFormat,omitempty"`
}

// Panel 面板配置
type Panel struct {
	ID               int            `json:"id"`
	Title            string         `json:"title"`
	Type             string         `json:"type"`
	GridPos          GridPos        `json:"gridPos"`
	Targets          []Target       `json:"targets"`
	Options          interface{}    `json:"options,omitempty"`
	FieldConfig      *FieldConfig   `json:"fieldConfig,omitempty"`
	Transformations  []Transformation `json:"transformations,omitempty"`
	Description      string         `json:"description,omitempty"`
	Datasource       interface{}    `json:"datasource,omitempty"`
	Links            []Link         `json:"links,omitempty"`
	Repeat           string         `json:"repeat,omitempty"`
	RepeatDirection  string         `json:"repeatDirection,omitempty"`
	Transparent      bool           `json:"transparent"`
}

// GridPos 网格位置
type GridPos struct {
	H int `json:"h"`
	W int `json:"w"`
	X int `json:"x"`
	Y int `json:"y"`
}

// Target 查询目标
type Target struct {
	RefID      string `json:"refId"`
	Expr       string `json:"expr,omitempty"`
	LegendFormat string `json:"legendFormat,omitempty"`
	Instant    bool   `json:"instant,omitempty"`
	Interval   string `json:"interval,omitempty"`
	IntervalFactor int `json:"intervalFactor,omitempty"`
	Step       string `json:"step,omitempty"`
	Target     string `json:"target,omitempty"`
	Query      string `json:"query,omitempty"`
	Datasource interface{} `json:"datasource,omitempty"`
}

// FieldConfig 字段配置
type FieldConfig struct {
	Defaults  Defaults `json:"defaults"`
	Overrides []Override `json:"overrides,omitempty"`
}

// Defaults 默认配置
type Defaults struct {
	Unit       string     `json:"unit,omitempty"`
	Decimals   int        `json:"decimals,omitempty"`
	Min        *float64   `json:"min,omitempty"`
	Max        *float64   `json:"max,omitempty"`
	Color      *Color     `json:"color,omitempty"`
	Thresholds Thresholds `json:"thresholds,omitempty"`
	Custom     interface{} `json:"custom,omitempty"`
}

// Color 颜色配置
type Color struct {
	Mode       string `json:"mode"`
	FixedColor string `json:"fixedColor,omitempty"`
}

// Thresholds 阈值配置
type Thresholds struct {
	Mode  string       `json:"mode"`
	Steps []ThresholdStep `json:"steps"`
}

// ThresholdStep 阈值步骤
type ThresholdStep struct {
	Color string  `json:"color"`
	Value float64 `json:"value"`
}

// Override 覆盖配置
type Override struct {
	Matcher    Matcher    `json:"matcher"`
	Properties []Property `json:"properties"`
}

// Matcher 匹配器
type Matcher struct {
	ID      string `json:"id"`
	Options string `json:"options"`
}

// Property 属性
type Property struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

// Transformation 转换配置
type Transformation struct {
	ID      string                 `json:"id"`
	Options map[string]interface{} `json:"options"`
}

// Link 链接
type Link struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	TargetBlank bool   `json:"targetBlank"`
}

// DataSource 数据源配置
type DataSource struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	Access    string `json:"access"`
	IsDefault bool   `json:"isDefault"`
	Database  string `json:"database,omitempty"`
	User      string `json:"user,omitempty"`
	JSONData  JSONData `json:"jsonData,omitempty"`
	SecureJSONData map[string]string `json:"secureJsonData,omitempty"`
}

// JSONData JSON数据
type JSONData struct {
	HTTPMethod    string `json:"httpMethod,omitempty"`
	ManageAlerts  bool   `json:"manageAlerts,omitempty"`
	MaxDataPoints int    `json:"maxDataPoints,omitempty"`
	TimeInterval  string `json:"timeInterval,omitempty"`
}

// AlertRule 告警规则
type AlertRule struct {
	Name        string        `json:"name"`
	Message     string        `json:"message"`
	Conditions  []Condition   `json:"conditions"`
	ExecutionErrorState string `json:"executionErrorState"`
	Frequency   string        `json:"frequency"`
	Handler     int           `json:"handler"`
	NoDataState string        `json:"noDataState"`
	Notifications []Notification `json:"notifications,omitempty"`
}

// Condition 告警条件
type Condition struct {
	Evaluator Evaluator `json:"evaluator"`
	Operator  Operator  `json:"operator"`
	Query     QueryCond `json:"query"`
	Reducer   Reducer   `json:"reducer"`
	Type      string    `json:"type"`
}

// Evaluator 评估器
type Evaluator struct {
	Params []float64 `json:"params"`
	Type   string    `json:"type"`
}

// Operator 操作符
type Operator struct {
	Type string `json:"type"`
}

// QueryCond 查询条件
type QueryCond struct {
	Params []string `json:"params"`
}

// Reducer 归约器
type Reducer struct {
	Params []string `json:"params"`
	Type   string   `json:"type"`
}

// Notification 通知配置
type Notification struct {
	UID string `json:"uid"`
}

// DashboardConfig 仪表盘配置器
type DashboardConfig struct {
	dashboard   *Dashboard
	dataSources []DataSource
	alertRules  []AlertRule
}

// NewDashboardConfig 创建仪表盘配置器
func NewDashboardConfig(title string) *DashboardConfig {
	return &DashboardConfig{
		dashboard: &Dashboard{
			Title:         title,
			Tags:          []string{"new-energy-monitoring"},
			Timezone:      "browser",
			SchemaVersion: 38,
			Refresh:       "30s",
			Time: TimeRange{
				From: "now-1h",
				To:   "now",
			},
			Panels:       make([]Panel, 0),
			Editable:     true,
			GraphTooltip: 1,
		},
		dataSources: make([]DataSource, 0),
		alertRules:  make([]AlertRule, 0),
	}
}

// SetUID 设置UID
func (dc *DashboardConfig) SetUID(uid string) *DashboardConfig {
	dc.dashboard.UID = uid
	return dc
}

// SetDescription 设置描述
func (dc *DashboardConfig) SetDescription(desc string) *DashboardConfig {
	dc.dashboard.Description = desc
	return dc
}

// SetRefresh 设置刷新间隔
func (dc *DashboardConfig) SetRefresh(refresh string) *DashboardConfig {
	dc.dashboard.Refresh = refresh
	return dc
}

// SetTimeRange 设置时间范围
func (dc *DashboardConfig) SetTimeRange(from, to string) *DashboardConfig {
	dc.dashboard.Time = TimeRange{
		From: from,
		To:   to,
	}
	return dc
}

// AddTag 添加标签
func (dc *DashboardConfig) AddTag(tag string) *DashboardConfig {
	dc.dashboard.Tags = append(dc.dashboard.Tags, tag)
	return dc
}

// AddDataSource 添加数据源
func (dc *DashboardConfig) AddDataSource(ds DataSource) *DashboardConfig {
	dc.dataSources = append(dc.dataSources, ds)
	return dc
}

// AddTemplateVariable 添加模板变量
func (dc *DashboardConfig) AddTemplateVariable(tv TemplateVariable) *DashboardConfig {
	dc.dashboard.Templating.List = append(dc.dashboard.Templating.List, tv)
	return dc
}

// AddAnnotation 添加注解
func (dc *DashboardConfig) AddAnnotation(ann Annotation) *DashboardConfig {
	dc.dashboard.Annotations.List = append(dc.dashboard.Annotations.List, ann)
	return dc
}

// AddPanel 添加面板
func (dc *DashboardConfig) AddPanel(panel Panel) *DashboardConfig {
	panel.ID = len(dc.dashboard.Panels) + 1
	dc.dashboard.Panels = append(dc.dashboard.Panels, panel)
	return dc
}

// AddAlertRule 添加告警规则
func (dc *DashboardConfig) AddAlertRule(rule AlertRule) *DashboardConfig {
	dc.alertRules = append(dc.alertRules, rule)
	return dc
}

// Build 构建仪表盘
func (dc *DashboardConfig) Build() *Dashboard {
	return dc.dashboard
}

// GetDataSources 获取数据源
func (dc *DashboardConfig) GetDataSources() []DataSource {
	return dc.dataSources
}

// GetAlertRules 获取告警规则
func (dc *DashboardConfig) GetAlertRules() []AlertRule {
	return dc.alertRules
}

// ToJSON 转换为JSON
func (d *Dashboard) ToJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}

// ToJSONString 转换为JSON字符串
func (d *Dashboard) ToJSONString() (string, error) {
	data, err := d.ToJSON()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveToFile 保存到文件
func (d *Dashboard) SaveToFile(filename string) error {
	data, err := d.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal dashboard: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadFromFile 从文件加载
func LoadFromFile(filename string) (*Dashboard, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var dashboard Dashboard
	if err := json.Unmarshal(data, &dashboard); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dashboard: %w", err)
	}

	return &dashboard, nil
}

// NewTimeSeriesPanel 创建时间序列面板
func NewTimeSeriesPanel(title string, x, y, w, h int) Panel {
	return Panel{
		Title: title,
		Type:  "timeseries",
		GridPos: GridPos{
			X: x, Y: y, W: w, H: h,
		},
		Targets: make([]Target, 0),
		Options: map[string]interface{}{
			"legend": map[string]interface{}{
				"displayMode": "list",
				"placement":   "bottom",
				"showLegend":  true,
			},
			"tooltip": map[string]interface{}{
				"mode": "multi",
				"sort": "none",
			},
		},
		FieldConfig: &FieldConfig{
			Defaults: Defaults{
				Custom: map[string]interface{}{
					"lineWidth":     1,
					"fillOpacity":   10,
					"gradientMode":  "none",
					"spanNulls":     false,
					"showPoints":    "never",
					"pointSize":     5,
					"stacking":      map[string]interface{}{"mode": "none", "group": "A"},
					"axisPlacement": "auto",
					"axisLabel":     "",
					"scaleDistribution": map[string]interface{}{
						"type": "linear",
					},
					"hideFrom": map[string]interface{}{
						"legend":  false,
						"tooltip": false,
						"viz":     false,
					},
					"thresholdsStyle": map[string]interface{}{
						"mode": "off",
					},
				},
			},
		},
	}
}

// NewStatPanel 创建统计面板
func NewStatPanel(title string, x, y, w, h int) Panel {
	return Panel{
		Title: title,
		Type:  "stat",
		GridPos: GridPos{
			X: x, Y: y, W: w, H: h,
		},
		Targets: make([]Target, 0),
		Options: map[string]interface{}{
			"colorMode":    "value",
			"graphMode":    "area",
			"justifyMode":  "auto",
			"orientation":  "auto",
			"reduceOptions": map[string]interface{}{
				"calcs":    []string{"lastNotNull"},
				"fields":   "",
				"values":   false,
			},
			"textMode":     "auto",
		},
		FieldConfig: &FieldConfig{
			Defaults: Defaults{},
		},
	}
}

// NewGaugePanel 创建仪表盘面板
func NewGaugePanel(title string, x, y, w, h int) Panel {
	return Panel{
		Title: title,
		Type:  "gauge",
		GridPos: GridPos{
			X: x, Y: y, W: w, H: h,
		},
		Targets: make([]Target, 0),
		Options: map[string]interface{}{
			"orientation": "auto",
			"reduceOptions": map[string]interface{}{
				"calcs":    []string{"lastNotNull"},
				"fields":   "",
				"values":   false,
			},
			"showThresholdLabels": false,
			"showThresholdMarkers": true,
		},
		FieldConfig: &FieldConfig{
			Defaults: Defaults{},
		},
	}
}

// NewTablePanel 创建表格面板
func NewTablePanel(title string, x, y, w, h int) Panel {
	return Panel{
		Title: title,
		Type:  "table",
		GridPos: GridPos{
			X: x, Y: y, W: w, H: h,
		},
		Targets: make([]Target, 0),
		Options: map[string]interface{}{
			"showHeader": true,
		},
		FieldConfig: &FieldConfig{
			Defaults: Defaults{
				Custom: map[string]interface{}{
					"align":       "auto",
					"filterable":  false,
				},
			},
		},
	}
}

// NewHeatmapPanel 创建热力图面板
func NewHeatmapPanel(title string, x, y, w, h int) Panel {
	return Panel{
		Title: title,
		Type:  "heatmap",
		GridPos: GridPos{
			X: x, Y: y, W: w, H: h,
		},
		Targets: make([]Target, 0),
		Options: map[string]interface{}{
			"calculate": false,
			"color": map[string]interface{}{
				"exponent":    0.5,
				"fill":        "dark-orange",
				"mode":        "spectrum",
				"reverse":     false,
				"scale":       "exponential",
				"scheme":      "Oranges",
				"steps":       128,
			},
			"dataFormat": "timeseries",
			"yBucketBound": "auto",
		},
	}
}

// AddTarget 添加查询目标
func (p *Panel) AddTarget(expr, legendFormat string) *Panel {
	p.Targets = append(p.Targets, Target{
		RefID:        string(rune('A' + len(p.Targets))),
		Expr:         expr,
		LegendFormat: legendFormat,
	})
	return p
}

// SetUnit 设置单位
func (p *Panel) SetUnit(unit string) *Panel {
	if p.FieldConfig == nil {
		p.FieldConfig = &FieldConfig{}
	}
	p.FieldConfig.Defaults.Unit = unit
	return p
}

// SetDecimals 设置小数位数
func (p *Panel) SetDecimals(decimals int) *Panel {
	if p.FieldConfig == nil {
		p.FieldConfig = &FieldConfig{}
	}
	p.FieldConfig.Defaults.Decimals = decimals
	return p
}

// SetThresholds 设置阈值
func (p *Panel) SetThresholds(steps []ThresholdStep) *Panel {
	if p.FieldConfig == nil {
		p.FieldConfig = &FieldConfig{}
	}
	p.FieldConfig.Defaults.Thresholds = Thresholds{
		Mode:  "absolute",
		Steps: steps,
	}
	return p
}

// SetMin 设置最小值
func (p *Panel) SetMin(min float64) *Panel {
	if p.FieldConfig == nil {
		p.FieldConfig = &FieldConfig{}
	}
	p.FieldConfig.Defaults.Min = &min
	return p
}

// SetMax 设置最大值
func (p *Panel) SetMax(max float64) *Panel {
	if p.FieldConfig == nil {
		p.FieldConfig = &FieldConfig{}
	}
	p.FieldConfig.Defaults.Max = &max
	return p
}

// NewPrometheusDataSource 创建Prometheus数据源
func NewPrometheusDataSource(name, url string, isDefault bool) DataSource {
	return DataSource{
		Name:      name,
		Type:      "prometheus",
		URL:       url,
		Access:    "proxy",
		IsDefault: isDefault,
		JSONData: JSONData{
			HTTPMethod:   "POST",
			ManageAlerts: true,
		},
	}
}

// NewJaegerDataSource 创建Jaeger数据源
func NewJaegerDataSource(name, url string) DataSource {
	return DataSource{
		Name:   name,
		Type:   "jaeger",
		URL:    url,
		Access: "proxy",
		JSONData: JSONData{
			HTTPMethod: "GET",
		},
	}
}

// NewLokiDataSource 创建Loki数据源
func NewLokiDataSource(name, url string) DataSource {
	return DataSource{
		Name:   name,
		Type:   "loki",
		URL:    url,
		Access: "proxy",
		JSONData: JSONData{
			MaxDataPoints: 1000,
		},
	}
}

// NewAlertRule 创建告警规则
func NewAlertRule(name, message string, frequency time.Duration) AlertRule {
	return AlertRule{
		Name:               name,
		Message:            message,
		Conditions:         make([]Condition, 0),
		ExecutionErrorState: "alerting",
		Frequency:          frequency.String(),
		Handler:            1,
		NoDataState:        "no_data",
		Notifications:      make([]Notification, 0),
	}
}

// AddCondition 添加条件
func (ar *AlertRule) AddCondition(queryRefID, evaluatorType string, evaluatorParams []float64) *AlertRule {
	ar.Conditions = append(ar.Conditions, Condition{
		Evaluator: Evaluator{
			Params: evaluatorParams,
			Type:   evaluatorType,
		},
		Operator: Operator{
			Type: "and",
		},
		Query: QueryCond{
			Params: []string{queryRefID, "5m", "now"},
		},
		Reducer: Reducer{
			Params: []string{},
			Type:   "avg",
		},
		Type: "query",
	})
	return ar
}

// AddNotification 添加通知
func (ar *AlertRule) AddNotification(uid string) *AlertRule {
	ar.Notifications = append(ar.Notifications, Notification{
		UID: uid,
	})
	return ar
}

// NewEnergyMonitoringDashboard 创建新能源监控仪表盘
func NewEnergyMonitoringDashboard() *Dashboard {
	config := NewDashboardConfig("New Energy Monitoring Dashboard").
		SetUID("new-energy-monitoring").
		SetDescription("New Energy Monitoring System Dashboard").
		SetRefresh("30s").
		SetTimeRange("now-6h", "now").
		AddTag("energy").
		AddTag("monitoring")

	// 添加模板变量
	config.AddTemplateVariable(TemplateVariable{
		Name:       "station",
		Type:       "query",
		DataSource: "Prometheus",
		Query:      "label_values(nem_station_power_watts, station)",
		Refresh:    1,
		IncludeAll: true,
		AllValue:   ".*",
		Multi:      true,
		Sort:       1,
		Label:      "Station",
	})

	// 添加面板
	// 第一行：关键指标
	config.AddPanel(NewStatPanel("Total Power", 0, 0, 6, 4).
		AddTarget(`sum(nem_station_power_watts{station=~"$station"})`, "Total Power").
		SetUnit("watt").
		SetDecimals(2))

	config.AddPanel(NewStatPanel("Active Stations", 6, 0, 6, 4).
		AddTarget(`count(nem_station_status{station=~"$station"} == 1)`, "Active").
		SetUnit("none"))

	config.AddPanel(NewStatPanel("Total Devices", 12, 0, 6, 4).
		AddTarget(`count(nem_device_status{station=~"$station"} == 1)`, "Devices").
		SetUnit("none"))

	config.AddPanel(NewStatPanel("Active Alarms", 18, 0, 6, 4).
		AddTarget(`sum(nem_alarms_active_total{station=~"$station"})`, "Alarms").
		SetUnit("none").
		SetThresholds([]ThresholdStep{
			{Color: "green", Value: 0},
			{Color: "yellow", Value: 5},
			{Color: "red", Value: 10},
		}))

	// 第二行：功率趋势
	config.AddPanel(NewTimeSeriesPanel("Station Power Trend", 0, 4, 12, 8).
		AddTarget(`nem_station_power_watts{station=~"$station"}`, "{{station}}").
		SetUnit("watt").
		SetDecimals(2))

	config.AddPanel(NewTimeSeriesPanel("Request Rate", 12, 4, 12, 8).
		AddTarget(`rate(nem_requests_total{service="api-server"}[5m])`, "Requests/s").
		SetUnit("reqps"))

	// 第三行：设备状态
	config.AddPanel(NewGaugePanel("System Health", 0, 12, 6, 6).
		AddTarget(`avg(nem_system_health_score)`, "Health").
		SetUnit("percent").
		SetMin(0).
		SetMax(100).
		SetThresholds([]ThresholdStep{
			{Color: "red", Value: 0},
			{Color: "yellow", Value: 60},
			{Color: "green", Value: 80},
		}))

	config.AddPanel(NewTimeSeriesPanel("Response Time", 6, 12, 12, 6).
		AddTarget(`histogram_quantile(0.50, rate(nem_request_duration_seconds_bucket[5m]))`, "P50").
		AddTarget(`histogram_quantile(0.95, rate(nem_request_duration_seconds_bucket[5m]))`, "P95").
		AddTarget(`histogram_quantile(0.99, rate(nem_request_duration_seconds_bucket[5m]))`, "P99").
		SetUnit("s").
		SetDecimals(3))

	config.AddPanel(NewTimeSeriesPanel("Error Rate", 18, 12, 6, 6).
		AddTarget(`rate(nem_errors_total[5m])`, "Errors/s").
		SetUnit("errorsps"))

	// 第四行：数据采集
	config.AddPanel(NewTimeSeriesPanel("Data Collection Rate", 0, 18, 12, 6).
		AddTarget(`rate(nem_data_points_collected_total[1m])`, "Points/s").
		SetUnit("pointsps"))

	config.AddPanel(NewStatPanel("Data Processed", 12, 18, 12, 6).
		AddTarget(`sum(nem_data_processed_bytes_total)`, "Bytes").
		SetUnit("decbytes"))

	return config.Build()
}

// NewSystemOverviewDashboard 创建系统概览仪表盘
func NewSystemOverviewDashboard() *Dashboard {
	config := NewDashboardConfig("System Overview").
		SetUID("system-overview").
		SetDescription("System Overview Dashboard").
		SetRefresh("10s").
		SetTimeRange("now-1h", "now")

	// CPU使用率
	config.AddPanel(NewTimeSeriesPanel("CPU Usage", 0, 0, 12, 6).
		AddTarget(`rate(process_cpu_seconds_total[1m])`, "CPU").
		SetUnit("percent").
		SetDecimals(1))

	// 内存使用
	config.AddPanel(NewTimeSeriesPanel("Memory Usage", 12, 0, 12, 6).
		AddTarget(`process_resident_memory_bytes`, "Memory").
		SetUnit("bytes"))

	// Goroutines
	config.AddPanel(NewTimeSeriesPanel("Goroutines", 0, 6, 12, 6).
		AddTarget(`go_goroutines`, "Goroutines").
		SetUnit("none"))

	// GC暂停
	config.AddPanel(NewTimeSeriesPanel("GC Pause", 12, 6, 12, 6).
		AddTarget(`rate(go_gc_duration_seconds_sum[5m])`, "GC Pause").
		SetUnit("s"))

	return config.Build()
}

// NewAlertDashboard 创建告警仪表盘
func NewAlertDashboard() *Dashboard {
	config := NewDashboardConfig("Alert Dashboard").
		SetUID("alert-dashboard").
		SetDescription("Alert Monitoring Dashboard").
		SetRefresh("10s").
		SetTimeRange("now-24h", "now")

	// 活跃告警
	config.AddPanel(NewStatPanel("Active Alarms", 0, 0, 6, 4).
		AddTarget(`sum(nem_alarms_active_total)`, "Active").
		SetUnit("none").
		SetThresholds([]ThresholdStep{
			{Color: "green", Value: 0},
			{Color: "yellow", Value: 5},
			{Color: "red", Value: 10},
		}))

	// 告警趋势
	config.AddPanel(NewTimeSeriesPanel("Alarm Trend", 6, 0, 18, 4).
		AddTarget(`rate(nem_alarms_total[1h])`, "Alarms/h").
		SetUnit("alarmsh"))

	// 告警分布
	config.AddPanel(NewTablePanel("Recent Alarms", 0, 4, 24, 8).
		AddTarget(`nem_alarms_recent`, "Alarms"))

	return config.Build()
}

// DashboardManager 仪表盘管理器
type DashboardManager struct {
	dashboards map[string]*Dashboard
	dataSources map[string]DataSource
}

// NewDashboardManager 创建仪表盘管理器
func NewDashboardManager() *DashboardManager {
	return &DashboardManager{
		dashboards:  make(map[string]*Dashboard),
		dataSources: make(map[string]DataSource),
	}
}

// AddDashboard 添加仪表盘
func (dm *DashboardManager) AddDashboard(dashboard *Dashboard) {
	dm.dashboards[dashboard.UID] = dashboard
}

// GetDashboard 获取仪表盘
func (dm *DashboardManager) GetDashboard(uid string) (*Dashboard, bool) {
	d, ok := dm.dashboards[uid]
	return d, ok
}

// RemoveDashboard 移除仪表盘
func (dm *DashboardManager) RemoveDashboard(uid string) {
	delete(dm.dashboards, uid)
}

// ListDashboards 列出所有仪表盘
func (dm *DashboardManager) ListDashboards() []*Dashboard {
	result := make([]*Dashboard, 0, len(dm.dashboards))
	for _, d := range dm.dashboards {
		result = append(result, d)
	}
	return result
}

// AddDataSource 添加数据源
func (dm *DashboardManager) AddDataSource(ds DataSource) {
	dm.dataSources[ds.Name] = ds
}

// GetDataSources 获取所有数据源
func (dm *DashboardManager) GetDataSources() []DataSource {
	result := make([]DataSource, 0, len(dm.dataSources))
	for _, ds := range dm.dataSources {
		result = append(result, ds)
	}
	return result
}

// ExportAll 导出所有仪表盘
func (dm *DashboardManager) ExportAll(dir string) error {
	for uid, dashboard := range dm.dashboards {
		filename := fmt.Sprintf("%s/%s.json", dir, uid)
		if err := dashboard.SaveToFile(filename); err != nil {
			return fmt.Errorf("failed to export dashboard %s: %w", uid, err)
		}
	}
	return nil
}

// ExportDataSources 导出数据源配置
func (dm *DashboardManager) ExportDataSources(filename string) error {
	data, err := json.MarshalIndent(dm.dataSources, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data sources: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
