package fault

import (
	"context"
	"fmt"
	"math"
	"time"
)

type ThresholdDetector struct {
	deviceID      string
	metric        string
	lowerBound    float64
	upperBound    float64
	minDeviation  float64
	windowSize    int
	trainedData   []*TimeSeriesData
	createdAt     time.Time
	trainedAt     *time.Time
}

func NewThresholdDetector(deviceID, metric string, lowerBound, upperBound, minDeviation float64, windowSize int) *ThresholdDetector {
	return &ThresholdDetector{
		deviceID:     deviceID,
		metric:       metric,
		lowerBound:   lowerBound,
		upperBound:   upperBound,
		minDeviation: minDeviation,
		windowSize:   windowSize,
		createdAt:    time.Now(),
	}
}

func (d *ThresholdDetector) Detect(ctx context.Context, data []*TimeSeriesData) ([]*Anomaly, error) {
	var anomalies []*Anomaly
	
	for i, point := range data {
		if point.DeviceID != d.deviceID || point.Metric != d.metric {
			continue
		}
		
		// 检查是否超出阈值
		if point.Value < d.lowerBound || point.Value > d.upperBound {
			expectedValue := (d.lowerBound + d.upperBound) / 2
			deviation := math.Abs(point.Value - expectedValue)
			
			// 检查偏差是否足够大
			if deviation < d.minDeviation {
				continue
			}
			
			// 确定异常严重程度
			severity := d.calculateSeverity(deviation)
			
			anomaly := &Anomaly{
				ID:             fmt.Sprintf("anomaly-%s-%s-%d", d.deviceID, d.metric, i),
				DeviceID:       d.deviceID,
				Timestamp:      point.Timestamp,
				Metric:         d.metric,
				Value:          point.Value,
				ExpectedValue:  expectedValue,
				Deviation:      deviation,
				Severity:       severity,
				DetectorName:   "ThresholdDetector",
				Confidence:     d.calculateConfidence(deviation),
				AdditionalInfo: map[string]interface{}{"thresholds": map[string]float64{"lower": d.lowerBound, "upper": d.upperBound}},
			}
			
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies, nil
}

func (d *ThresholdDetector) Train(ctx context.Context, data []*TimeSeriesData) error {
	d.trainedData = data
	now := time.Now()
	d.trainedAt = &now
	
	// 可以根据训练数据调整阈值
	if len(data) > 0 {
		// 简单的统计分析来调整阈值
		var sum, sumSquared float64
		for _, point := range data {
			sum += point.Value
			sumSquared += point.Value * point.Value
		}
		
		mean := sum / float64(len(data))
		variance := (sumSquared / float64(len(data))) - (mean * mean)
		stdDev := math.Sqrt(variance)
		
		// 自动调整阈值为均值的±3倍标准差
		d.lowerBound = mean - 3*stdDev
		d.upperBound = mean + 3*stdDev
		d.minDeviation = stdDev * 0.5
	}
	
	return nil
}

func (d *ThresholdDetector) GetDetectorInfo() *DetectorInfo {
	return &DetectorInfo{
		DetectorID:   fmt.Sprintf("detector-%s-%s", d.deviceID, d.metric),
		DetectorType: "ThresholdDetector",
		Version:      "1.0.0",
		CreatedAt:    d.createdAt,
		TrainedAt:    d.trainedAt,
		Parameters: map[string]interface{}{
			"device_id":     d.deviceID,
			"metric":        d.metric,
			"lower_bound":   d.lowerBound,
			"upper_bound":   d.upperBound,
			"min_deviation": d.minDeviation,
			"window_size":   d.windowSize,
		},
		Metrics: map[string]float64{
			"trained_data_points": float64(len(d.trainedData)),
		},
		Status: "active",
	}
}

func (d *ThresholdDetector) calculateSeverity(deviation float64) AnomalySeverity {
	// 基于偏差程度计算严重程度
	normalRange := d.upperBound - d.lowerBound
	deviationRatio := deviation / (normalRange / 2)
	
	switch {
	case deviationRatio >= 3.0:
		return SeverityCritical
	case deviationRatio >= 2.0:
		return SeverityHigh
	case deviationRatio >= 1.0:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

func (d *ThresholdDetector) calculateConfidence(deviation float64) float64 {
	// 基于偏差程度计算置信度
	normalRange := d.upperBound - d.lowerBound
	deviationRatio := deviation / (normalRange / 2)
	
	// 置信度随偏差增大而增加，最大为0.99
	confidence := math.Min(0.99, 0.5+0.5*(deviationRatio/3.0))
	return confidence
}

type StatisticalDetector struct {
	deviceID    string
	metric      string
	windowSize  int
	mean        float64
	stdDev      float64
	threshold   float64
	trainedData []*TimeSeriesData
	createdAt   time.Time
	trainedAt   *time.Time
}

func NewStatisticalDetector(deviceID, metric string, windowSize int, threshold float64) *StatisticalDetector {
	return &StatisticalDetector{
		deviceID:   deviceID,
		metric:     metric,
		windowSize: windowSize,
		threshold:  threshold,
		createdAt:  time.Now(),
	}
}

func (d *StatisticalDetector) Detect(ctx context.Context, data []*TimeSeriesData) ([]*Anomaly, error) {
	var anomalies []*Anomaly
	
	// 确保有足够的数据点进行统计分析
	if len(data) < d.windowSize {
		return anomalies, nil
	}
	
	// 计算移动窗口的统计数据
	for i := d.windowSize - 1; i < len(data); i++ {
		point := data[i]
		if point.DeviceID != d.deviceID || point.Metric != d.metric {
			continue
		}
		
		// 计算窗口内的数据统计
		windowData := data[i-d.windowSize+1 : i+1]
		currentMean, currentStdDev := d.calculateStats(windowData)
		
		// 检查是否异常
		zScore := math.Abs(point.Value - currentMean) / currentStdDev
		if zScore > d.threshold {
			anomaly := &Anomaly{
				ID:             fmt.Sprintf("anomaly-%s-%s-%d", d.deviceID, d.metric, i),
				DeviceID:       d.deviceID,
				Timestamp:      point.Timestamp,
				Metric:         d.metric,
				Value:          point.Value,
				ExpectedValue:  currentMean,
				Deviation:      math.Abs(point.Value - currentMean),
				Severity:       d.calculateSeverity(zScore),
				DetectorName:   "StatisticalDetector",
				Confidence:     d.calculateConfidence(zScore),
				AdditionalInfo: map[string]interface{}{"z_score": zScore, "threshold": d.threshold},
			}
			
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies, nil
}

func (d *StatisticalDetector) Train(ctx context.Context, data []*TimeSeriesData) error {
	d.trainedData = data
	now := time.Now()
	d.trainedAt = &now
	
	// 计算整体统计数据
	if len(data) > 0 {
		d.mean, d.stdDev = d.calculateStats(data)
	}
	
	return nil
}

func (d *StatisticalDetector) GetDetectorInfo() *DetectorInfo {
	return &DetectorInfo{
		DetectorID:   fmt.Sprintf("detector-%s-%s", d.deviceID, d.metric),
		DetectorType: "StatisticalDetector",
		Version:      "1.0.0",
		CreatedAt:    d.createdAt,
		TrainedAt:    d.trainedAt,
		Parameters: map[string]interface{}{
			"device_id":   d.deviceID,
			"metric":      d.metric,
			"window_size": d.windowSize,
			"threshold":   d.threshold,
			"mean":        d.mean,
			"std_dev":     d.stdDev,
		},
		Metrics: map[string]float64{
			"trained_data_points": float64(len(d.trainedData)),
		},
		Status: "active",
	}
}

func (d *StatisticalDetector) calculateStats(data []*TimeSeriesData) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}
	
	var sum, sumSquared float64
	for _, point := range data {
		sum += point.Value
		sumSquared += point.Value * point.Value
	}
	
	mean := sum / float64(len(data))
	variance := (sumSquared / float64(len(data))) - (mean * mean)
	stdDev := math.Sqrt(variance)
	
	return mean, stdDev
}

func (d *StatisticalDetector) calculateSeverity(zScore float64) AnomalySeverity {
	switch {
	case zScore >= 4.0:
		return SeverityCritical
	case zScore >= 3.0:
		return SeverityHigh
	case zScore >= 2.0:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

func (d *StatisticalDetector) calculateConfidence(zScore float64) float64 {
	// 置信度随z-score增大而增加
	confidence := math.Min(0.99, 0.5+0.5*(zScore/4.0))
	return confidence
}