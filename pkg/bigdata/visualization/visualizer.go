package visualization

import (
	"fmt"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
)

// BasicVisualizer 实现了types.Visualization接口，提供基本的可视化功能
type BasicVisualizer struct {
	config types.VisualizationConfig
	dashboards map[string][]types.Panel
}

// NewBasicVisualizer 创建一个新的基本可视化器实例
func NewBasicVisualizer() *BasicVisualizer {
	return &BasicVisualizer{
		dashboards: make(map[string][]types.Panel),
	}
}

// Init 初始化可视化器
func (v *BasicVisualizer) Init(config types.VisualizationConfig) error {
	if config.Type != "basic" {
		return &types.Error{
			Code:    types.ErrCodeInvalidConfig,
			Message: fmt.Sprintf("invalid visualization type: %s, expected basic", config.Type),
		}
	}

	v.config = config
	return nil
}

// CreateDashboard 创建仪表板
func (v *BasicVisualizer) CreateDashboard(name string, panels []types.Panel) error {
	if name == "" {
		return &types.Error{
			Code:    types.ErrCodeVisualizationError,
			Message: "dashboard name cannot be empty",
		}
	}

	// 为每个面板生成ID（如果没有）
	for i := range panels {
		if panels[i].ID == "" {
			panels[i].ID = fmt.Sprintf("panel-%d-%d", len(v.dashboards), i)
		}
	}

	v.dashboards[name] = panels
	return nil
}

// UpdatePanel 更新面板数据
func (v *BasicVisualizer) UpdatePanel(dashboardID, panelID string, data interface{}) error {
	if _, ok := v.dashboards[dashboardID]; !ok {
		return &types.Error{
			Code:    types.ErrCodeVisualizationError,
			Message: fmt.Sprintf("dashboard %s not found", dashboardID),
		}
	}

	panels := v.dashboards[dashboardID]
	found := false

	for i, panel := range panels {
		if panel.ID == panelID {
			panels[i].Data = data
			found = true
			break
		}
	}

	if !found {
		return &types.Error{
			Code:    types.ErrCodeVisualizationError,
			Message: fmt.Sprintf("panel %s not found in dashboard %s", panelID, dashboardID),
		}
	}

	v.dashboards[dashboardID] = panels
	return nil
}

// GetDashboard 获取仪表板
func (v *BasicVisualizer) GetDashboard(dashboardID string) ([]types.Panel, error) {
	if panels, ok := v.dashboards[dashboardID]; ok {
		return panels, nil
	}
	return nil, &types.Error{
		Code:    types.ErrCodeVisualizationError,
		Message: fmt.Sprintf("dashboard %s not found", dashboardID),
	}
}

// ListDashboards 列出所有仪表板
func (v *BasicVisualizer) ListDashboards() []string {
	var dashboards []string
	for name := range v.dashboards {
		dashboards = append(dashboards, name)
	}
	return dashboards
}

// Close 关闭可视化器
func (v *BasicVisualizer) Close() error {
	// 基本可视化器不需要特殊清理
	return nil
}

// GenerateTimeSeriesChart 生成时间序列图表数据
func (v *BasicVisualizer) GenerateTimeSeriesChart(dataPoints []*types.DataPoint) map[string]interface{} {
	timestamps := make([]string, len(dataPoints))
	values := make([]float64, len(dataPoints))

	for i, point := range dataPoints {
		timestamps[i] = point.Timestamp.Format(time.RFC3339)
		values[i] = point.Value
	}

	return map[string]interface{}{
		"type":        "time_series",
		"timestamps":  timestamps,
		"values":      values,
		"timestamp":   time.Now(),
	}
}

// GenerateGaugeChart 生成仪表盘图表数据
func (v *BasicVisualizer) GenerateGaugeChart(value, min, max float64, label string) map[string]interface{} {
	return map[string]interface{}{
		"type":      "gauge",
		"value":     value,
		"min":       min,
		"max":       max,
		"label":     label,
		"timestamp": time.Now(),
	}
}

// GenerateBarChart 生成柱状图数据
func (v *BasicVisualizer) GenerateBarChart(labels []string, values []float64) map[string]interface{} {
	return map[string]interface{}{
		"type":      "bar",
		"labels":    labels,
		"values":    values,
		"timestamp": time.Now(),
	}
}

// GeneratePieChart 生成饼图数据
func (v *BasicVisualizer) GeneratePieChart(labels []string, values []float64) map[string]interface{} {
	return map[string]interface{}{
		"type":      "pie",
		"labels":    labels,
		"values":    values,
		"timestamp": time.Now(),
	}
}
