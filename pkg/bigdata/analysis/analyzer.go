package analysis

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata"
)

// BasicAnalyzer 实现了Analysis接口，提供基本的数据分析功能
type BasicAnalyzer struct {
	config bigdata.AnalysisConfig
}

// NewBasicAnalyzer 创建一个新的基本分析器实例
func NewBasicAnalyzer() *BasicAnalyzer {
	return &BasicAnalyzer{}
}

// Init 初始化分析器
func (a *BasicAnalyzer) Init(config bigdata.AnalysisConfig) error {
	if config.Type != "basic" {
		return &bigdata.Error{
			Code:    bigdata.ErrCodeInvalidConfig,
			Message: fmt.Sprintf("invalid analysis type: %s, expected basic", config.Type),
		}
	}

	a.config = config
	return nil
}

// Execute 执行分析查询
func (a *BasicAnalyzer) Execute(query string) (interface{}, error) {
	// 这里实现基本的分析查询
	// 简化实现，实际项目中可能需要更复杂的查询解析和执行
	
	// 示例：支持简单的聚合查询
	// 格式: "aggregate by device_id,metric:sum,avg,max,min"
	
	return map[string]interface{}{
		"query":     query,
		"timestamp": time.Now(),
		"message":   "Basic analysis executed",
	}, nil
}

// Process 处理批量数据
func (a *BasicAnalyzer) Process(data *bigdata.BatchData) (interface{}, error) {
	if data == nil || len(data.DataPoints) == 0 {
		return nil, &bigdata.Error{
			Code:    bigdata.ErrCodeAnalysisError,
			Message: "no data to process",
		}
	}

	// 计算基本统计信息
	stats := a.calculateStats(data.DataPoints)

	// 按设备和指标分组分析
	groupedStats := a.groupByDeviceAndMetric(data.DataPoints)

	// 时间序列分析
	timeSeriesAnalysis := a.analyzeTimeSeries(data.DataPoints)

	result := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_points":    len(data.DataPoints),
			"time_range":      fmt.Sprintf("%s to %s", data.DataPoints[0].Timestamp, data.DataPoints[len(data.DataPoints)-1].Timestamp),
			"basic_statistics": stats,
		},
		"grouped_statistics": groupedStats,
		"time_series_analysis": timeSeriesAnalysis,
		"metadata":          data.Metadata,
	}

	return result, nil
}

// calculateStats 计算基本统计信息
func (a *BasicAnalyzer) calculateStats(dataPoints []*bigdata.DataPoint) map[string]float64 {
	if len(dataPoints) == 0 {
		return nil
	}

	var sum, min, max float64
	values := make([]float64, len(dataPoints))

	for i, point := range dataPoints {
		value := point.Value
		values[i] = value
		sum += value

		if i == 0 {
			min = value
			max = value
		} else {
			if value < min {
				min = value
			}
			if value > max {
				max = value
			}
		}
	}

	mean := sum / float64(len(dataPoints))

	// 计算标准差
	var variance float64
	for _, value := range values {
		variance += math.Pow(value-mean, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(dataPoints)))

	// 计算中位数
	sort.Float64s(values)
	var median float64
	n := len(values)
	if n%2 == 0 {
		median = (values[n/2-1] + values[n/2]) / 2
	} else {
		median = values[n/2]
	}

	return map[string]float64{
		"sum":     sum,
		"mean":    mean,
		"min":     min,
		"max":     max,
		"std_dev": stdDev,
		"median":  median,
	}
}

// groupByDeviceAndMetric 按设备和指标分组分析
func (a *BasicAnalyzer) groupByDeviceAndMetric(dataPoints []*bigdata.DataPoint) map[string]map[string]map[string]float64 {
	groups := make(map[string]map[string]map[string]float64)

	for _, point := range dataPoints {
		if _, ok := groups[point.DeviceID]; !ok {
			groups[point.DeviceID] = make(map[string]map[string]float64)
		}

		if _, ok := groups[point.DeviceID][point.Metric]; !ok {
			groups[point.DeviceID][point.Metric] = make(map[string]float64)
			groups[point.DeviceID][point.Metric]["count"] = 0
			groups[point.DeviceID][point.Metric]["sum"] = 0
			groups[point.DeviceID][point.Metric]["min"] = point.Value
			groups[point.DeviceID][point.Metric]["max"] = point.Value
		}

		group := groups[point.DeviceID][point.Metric]
		group["count"]++
		group["sum"] += point.Value

		if point.Value < group["min"] {
			group["min"] = point.Value
		}
		if point.Value > group["max"] {
			group["max"] = point.Value
		}
	}

	// 计算每个组的平均值
	for deviceID, metrics := range groups {
		for metric, stats := range metrics {
			stats["mean"] = stats["sum"] / stats["count"]
			groups[deviceID][metric] = stats
		}
	}

	return groups
}

// analyzeTimeSeries 分析时间序列数据
func (a *BasicAnalyzer) analyzeTimeSeries(dataPoints []*bigdata.DataPoint) map[string]interface{} {
	// 按设备和指标分组
	groups := make(map[string]map[string][]*bigdata.DataPoint)

	for _, point := range dataPoints {
		if _, ok := groups[point.DeviceID]; !ok {
			groups[point.DeviceID] = make(map[string][]*bigdata.DataPoint)
		}
		groups[point.DeviceID][point.Metric] = append(groups[point.DeviceID][point.Metric], point)
	}

	// 分析每个时间序列
	analysis := make(map[string]interface{})

	for deviceID, metrics := range groups {
		deviceAnalysis := make(map[string]interface{})

		for metric, points := range metrics {
			// 按时间排序
			sort.Slice(points, func(i, j int) bool {
				return points[i].Timestamp.Before(points[j].Timestamp)
			})

			// 计算趋势
			trend := a.calculateTrend(points)

			// 计算波动率
			volatility := a.calculateVolatility(points)

			// 检测异常
			anomalies := a.detectAnomalies(points)

			metricAnalysis := map[string]interface{}{
				"data_points": len(points),
				"time_range":  fmt.Sprintf("%s to %s", points[0].Timestamp, points[len(points)-1].Timestamp),
				"trend":       trend,
				"volatility":  volatility,
				"anomalies":   anomalies,
			}

			deviceAnalysis[metric] = metricAnalysis
		}

		analysis[deviceID] = deviceAnalysis
	}

	return analysis
}

// calculateTrend 计算时间序列趋势
func (a *BasicAnalyzer) calculateTrend(points []*bigdata.DataPoint) float64 {
	if len(points) < 2 {
		return 0
	}

	// 使用简单线性回归计算趋势
	n := float64(len(points))
	var sumX, sumY, sumXY, sumX2 float64

	for i, point := range points {
		x := float64(i)
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	return slope
}

// calculateVolatility 计算时间序列波动率
func (a *BasicAnalyzer) calculateVolatility(points []*bigdata.DataPoint) float64 {
	if len(points) < 2 {
		return 0
	}

	// 计算收益率
	returns := make([]float64, len(points)-1)
	for i := 1; i < len(points); i++ {
		if points[i-1].Value != 0 {
			returns[i-1] = (points[i].Value - points[i-1].Value) / points[i-1].Value
		}
	}

	// 计算收益率的标准差
	if len(returns) == 0 {
		return 0
	}

	var sum, mean float64
	for _, r := range returns {
		sum += r
	}
	mean = sum / float64(len(returns))

	var variance float64
	for _, r := range returns {
		variance += math.Pow(r-mean, 2)
	}

	return math.Sqrt(variance / float64(len(returns)))
}

// detectAnomalies 检测时间序列异常
func (a *BasicAnalyzer) detectAnomalies(points []*bigdata.DataPoint) []map[string]interface{} {
	if len(points) < 3 {
		return []map[string]interface{}{}
	}

	// 使用移动平均和标准差检测异常
	windowSize := 3
	var anomalies []map[string]interface{}

	for i := windowSize - 1; i < len(points); i++ {
		// 计算移动平均
		var sum float64
		for j := i - windowSize + 1; j <= i; j++ {
			sum += points[j].Value
		}
		mean := sum / float64(windowSize)

		// 计算移动标准差
		var variance float64
		for j := i - windowSize + 1; j <= i; j++ {
			variance += math.Pow(points[j].Value-mean, 2)
		}
		stdDev := math.Sqrt(variance / float64(windowSize))

		// 检测异常（超过2个标准差）
		if math.Abs(points[i].Value-mean) > 2*stdDev {
			anomaly := map[string]interface{}{
				"timestamp":   points[i].Timestamp,
				"value":       points[i].Value,
				"expected":    mean,
				"deviation":   math.Abs(points[i].Value - mean),
				"std_dev":     stdDev,
				"severity":    "high",
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}

// Close 关闭分析器
func (a *BasicAnalyzer) Close() error {
	// 基本分析器不需要特殊清理
	return nil
}
