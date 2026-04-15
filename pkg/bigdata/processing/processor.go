package processing

import (
	"fmt"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata"
)

// BasicProcessor 实现了Processing接口，提供基本的数据处理功能
type BasicProcessor struct {
	config bigdata.ProcessingConfig
}

// NewBasicProcessor 创建一个新的基本处理器实例
func NewBasicProcessor() *BasicProcessor {
	return &BasicProcessor{}
}

// Init 初始化处理器
func (p *BasicProcessor) Init(config bigdata.ProcessingConfig) error {
	if config.Type != "basic" {
		return &bigdata.Error{
			Code:    bigdata.ErrCodeInvalidConfig,
			Message: fmt.Sprintf("invalid processing type: %s, expected basic", config.Type),
		}
	}

	p.config = config
	return nil
}

// Process 处理批量数据
func (p *BasicProcessor) Process(data *bigdata.BatchData) (*bigdata.BatchData, error) {
	if data == nil || len(data.DataPoints) == 0 {
		return data, nil
	}

	// 执行数据处理
	processedPoints := p.processDataPoints(data.DataPoints)

	// 创建新的批量数据
	processedData := &bigdata.BatchData{
		DataPoints: processedPoints,
		Metadata: bigdata.Metadata{
			Source:      data.Metadata.Source,
			BatchID:     fmt.Sprintf("%s-processed", data.Metadata.BatchID),
			Timestamp:   time.Now(),
			RecordCount: len(processedPoints),
			Properties: map[string]interface{}{
				"processed":       true,
				"original_count":  data.Metadata.RecordCount,
				"processing_time": time.Since(data.Metadata.Timestamp).String(),
			},
		},
	}

	return processedData, nil
}

// processDataPoints 处理数据点
func (p *BasicProcessor) processDataPoints(dataPoints []*bigdata.DataPoint) []*bigdata.DataPoint {
	var processedPoints []*bigdata.DataPoint

	for _, point := range dataPoints {
		// 复制数据点
		processedPoint := &bigdata.DataPoint{
			Timestamp:  point.Timestamp,
			DeviceID:   point.DeviceID,
			Metric:     point.Metric,
			Value:      point.Value,
			Tags:       make(map[string]string),
			Attributes: make(map[string]interface{}),
		}

		// 复制标签
		if point.Tags != nil {
			for k, v := range point.Tags {
				processedPoint.Tags[k] = v
			}
		}

		// 复制属性
		if point.Attributes != nil {
			for k, v := range point.Attributes {
				processedPoint.Attributes[k] = v
			}
		}

		// 执行数据处理
		p.processSinglePoint(processedPoint)

		processedPoints = append(processedPoints, processedPoint)
	}

	return processedPoints
}

// processSinglePoint 处理单个数据点
func (p *BasicProcessor) processSinglePoint(point *bigdata.DataPoint) {
	// 1. 数据清洗
	p.cleanData(point)

	// 2. 数据转换
	p.transformData(point)

	// 3. 数据增强
	p.enhanceData(point)
}

// cleanData 清洗数据
func (p *BasicProcessor) cleanData(point *bigdata.DataPoint) {
	// 处理异常值
	if point.Value < 0 {
		point.Value = 0
		point.Attributes["cleaned"] = true
		point.Attributes["cleaned_reason"] = "negative value"
	}

	// 处理极端值
	if point.Metric == "temperature" && (point.Value < -40 || point.Value > 125) {
		point.Value = 25 // 设置为默认值
		point.Attributes["cleaned"] = true
		point.Attributes["cleaned_reason"] = "temperature out of range"
	}

	if point.Metric == "voltage" && (point.Value < 0 || point.Value > 10000) {
		point.Value = 0
		point.Attributes["cleaned"] = true
		point.Attributes["cleaned_reason"] = "voltage out of range"
	}
}

// transformData 转换数据
func (p *BasicProcessor) transformData(point *bigdata.DataPoint) {
	// 根据指标类型进行转换
	switch point.Metric {
	case "power":
		// 转换为kW
		if point.Value > 1000 {
			point.Value = point.Value / 1000
			point.Metric = "power_kw"
			point.Attributes["transformed"] = true
			point.Attributes["transformation"] = "W to kW"
		}

	case "energy":
		// 转换为kWh
		if point.Value > 1000 {
			point.Value = point.Value / 1000
			point.Metric = "energy_kwh"
			point.Attributes["transformed"] = true
			point.Attributes["transformation"] = "Wh to kWh"
		}

	case "current":
		// 确保电流值合理
		if point.Value > 1000 {
			point.Value = 0
			point.Attributes["cleaned"] = true
			point.Attributes["cleaned_reason"] = "current too high"
		}
	}
}

// enhanceData 增强数据
func (p *BasicProcessor) enhanceData(point *bigdata.DataPoint) {
	// 添加时间特征
	point.Attributes["hour"] = point.Timestamp.Hour()
	point.Attributes["day_of_week"] = point.Timestamp.Weekday()
	point.Attributes["month"] = point.Timestamp.Month()
	point.Attributes["is_weekend"] = point.Timestamp.Weekday() == time.Saturday || point.Timestamp.Weekday() == time.Sunday

	// 添加设备类型标签
	if _, ok := point.Tags["device_type"]; !ok {
		// 根据设备ID推断设备类型
		if point.DeviceID[:3] == "SOL" {
			point.Tags["device_type"] = "solar"
		} else if point.DeviceID[:3] == "WND" {
			point.Tags["device_type"] = "wind"
		} else if point.DeviceID[:3] == "BAT" {
			point.Tags["device_type"] = "battery"
		} else {
			point.Tags["device_type"] = "unknown"
		}
	}

	// 添加数据质量指标
	point.Attributes["data_quality"] = p.calculateDataQuality(point)
}

// calculateDataQuality 计算数据质量
func (p *BasicProcessor) calculateDataQuality(point *bigdata.DataPoint) float64 {
	quality := 1.0

	// 检查值是否合理
	switch point.Metric {
	case "temperature":
		if point.Value >= -40 && point.Value <= 125 {
			quality *= 1.0
		} else {
			quality *= 0.5
		}

	case "voltage":
		if point.Value >= 0 && point.Value <= 10000 {
			quality *= 1.0
		} else {
			quality *= 0.5
		}

	case "current":
		if point.Value >= 0 && point.Value <= 1000 {
			quality *= 1.0
		} else {
			quality *= 0.5
		}

	case "power":
	case "energy":
		if point.Value >= 0 {
			quality *= 1.0
		} else {
			quality *= 0.5
		}
	}

	// 检查时间戳是否合理
	now := time.Now()
	if point.Timestamp.Before(now.Add(-24 * time.Hour)) || point.Timestamp.After(now.Add(1 * time.Hour)) {
		quality *= 0.8
	}

	return quality
}

// Close 关闭处理器
func (p *BasicProcessor) Close() error {
	// 基本处理器不需要特殊清理
	return nil
}
