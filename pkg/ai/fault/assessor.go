package fault

import (
	"context"
	"fmt"
	"math"
	"time"
)

type SimpleHealthAssessor struct {
	deviceID      string
	deviceType    string
	healthHistory []*HealthAssessment
	rulHistory    []*RULPrediction
	createdAt     time.Time
	trainedAt     *time.Time
}

func NewSimpleHealthAssessor(deviceID, deviceType string) *SimpleHealthAssessor {
	return &SimpleHealthAssessor{
		deviceID:   deviceID,
		deviceType: deviceType,
		createdAt:  time.Now(),
	}
}

func (a *SimpleHealthAssessor) Assess(ctx context.Context, data []*TimeSeriesData) (*HealthAssessment, error) {
	if len(data) == 0 {
		return &HealthAssessment{
			DeviceID:       a.deviceID,
			Timestamp:      time.Now(),
			HealthScore:    0,
			HealthStatus:   HealthStatusCritical,
			Confidence:     0.5,
			AdditionalInfo: map[string]interface{}{"error": "no data provided"},
		}, nil
	}
	
	// 计算各项指标的健康分数
	metrics := make(map[string]float64)
	for _, point := range data {
		if point.DeviceID == a.deviceID {
			metrics[point.Metric] = a.calculateMetricHealth(point)
		}
	}
	
	// 计算整体健康分数
	overallScore := a.calculateOverallHealth(metrics)
	status := a.determineHealthStatus(overallScore)
	
	// 识别问题
	issues := a.identifyIssues(metrics)
	
	assessment := &HealthAssessment{
		DeviceID:       a.deviceID,
		Timestamp:      time.Now(),
		HealthScore:    overallScore,
		HealthStatus:   status,
		ComponentHealth: metrics,
		Issues:         issues,
		Confidence:     a.calculateConfidence(metrics),
		AdditionalInfo: map[string]interface{}{"device_type": a.deviceType, "data_points": len(data)},
	}
	
	// 保存历史记录
	a.healthHistory = append(a.healthHistory, assessment)
	if len(a.healthHistory) > 100 {
		a.healthHistory = a.healthHistory[len(a.healthHistory)-100:]
	}
	
	return assessment, nil
}

func (a *SimpleHealthAssessor) PredictRUL(ctx context.Context, data []*TimeSeriesData) (*RULPrediction, error) {
	if len(data) == 0 {
		return &RULPrediction{
			DeviceID:       a.deviceID,
			Timestamp:      time.Now(),
			PredictedRUL:   0,
			Confidence:     0.5,
			RULInterval:    [2]float64{0, 0},
			HealthTrend:    "unknown",
			AdditionalInfo: map[string]interface{}{"error": "no data provided"},
		}, nil
	}
	
	// 基于健康历史和当前数据预测RUL
	baseRUL := a.getBaseRUL()
	healthScore := a.calculateOverallHealthForRUL(data)
	
	// 根据健康分数调整RUL
	adjustedRUL := baseRUL * (healthScore / 100.0)
	
	// 计算置信区间
	confidenceInterval := a.calculateRULInterval(adjustedRUL, healthScore)
	
	// 确定健康趋势
	trend := a.determineHealthTrend()
	
	rulPrediction := &RULPrediction{
		DeviceID:       a.deviceID,
		Timestamp:      time.Now(),
		PredictedRUL:   adjustedRUL,
		Confidence:     a.calculateRULConfidence(healthScore),
		RULInterval:    confidenceInterval,
		HealthTrend:    trend,
		AdditionalInfo: map[string]interface{}{"device_type": a.deviceType, "base_rul": baseRUL},
	}
	
	// 保存历史记录
	a.rulHistory = append(a.rulHistory, rulPrediction)
	if len(a.rulHistory) > 100 {
		a.rulHistory = a.rulHistory[len(a.rulHistory)-100:]
	}
	
	return rulPrediction, nil
}

func (a *SimpleHealthAssessor) GetAssessorInfo() *AssessorInfo {
	return &AssessorInfo{
		AssessorID:   fmt.Sprintf("assessor-%s", a.deviceID),
		AssessorType: "SimpleHealthAssessor",
		Version:      "1.0.0",
		CreatedAt:    a.createdAt,
		TrainedAt:    a.trainedAt,
		Parameters: map[string]interface{}{
			"device_id":   a.deviceID,
			"device_type": a.deviceType,
		},
		Metrics: map[string]float64{
			"health_history_count": float64(len(a.healthHistory)),
			"rul_history_count":    float64(len(a.rulHistory)),
		},
		Status: "active",
	}
}

func (a *SimpleHealthAssessor) calculateMetricHealth(point *TimeSeriesData) float64 {
	// 根据不同指标计算健康分数
	switch point.Metric {
	case "temperature":
		// 温度健康分数：25-35度为最佳
		if point.Value >= 25 && point.Value <= 35 {
			return 100
		} else if point.Value >= 15 && point.Value < 25 {
			return 80 + (point.Value-15)*2
		} else if point.Value > 35 && point.Value <= 45 {
			return 80 - (point.Value-35)*2
		} else if point.Value >= 5 && point.Value < 15 {
			return 60 + (point.Value-5)*2
		} else if point.Value > 45 && point.Value <= 55 {
			return 60 - (point.Value-45)*2
		} else {
			return 40
		}
	case "voltage":
		// 电压健康分数：根据设备类型不同有不同标准
		switch a.deviceType {
		case "solar":
			if point.Value >= 400 && point.Value <= 600 {
				return 100
			} else if point.Value >= 300 && point.Value < 400 {
				return 70 + (point.Value-300)*0.3
			} else if point.Value > 600 && point.Value <= 700 {
				return 70 - (point.Value-600)*0.3
			} else {
				return 40
			}
		case "wind":
			if point.Value >= 600 && point.Value <= 800 {
				return 100
			} else if point.Value >= 400 && point.Value < 600 {
				return 70 + (point.Value-400)*0.15
			} else if point.Value > 800 && point.Value <= 1000 {
				return 70 - (point.Value-800)*0.15
			} else {
				return 40
			}
		case "battery":
			if point.Value >= 3.2 && point.Value <= 3.8 {
				return 100
			} else if point.Value >= 3.0 && point.Value < 3.2 {
				return 80 + (point.Value-3.0)*100
			} else if point.Value > 3.8 && point.Value <= 4.0 {
				return 80 - (point.Value-3.8)*100
			} else {
				return 40
			}
		default:
			return 70
		}
	case "soc":
		// SOC健康分数：20-80%为最佳
		if point.Value >= 20 && point.Value <= 80 {
			return 100
		} else if point.Value >= 10 && point.Value < 20 {
			return 60 + (point.Value-10)*4
		} else if point.Value > 80 && point.Value <= 90 {
			return 60 - (point.Value-80)*4
		} else {
			return 20
		}
	case "vibration":
		// 振动健康分数：越小越好
		if point.Value < 1 {
			return 100
		} else if point.Value < 2 {
			return 90 - (point.Value-1)*10
		} else if point.Value < 3 {
			return 80 - (point.Value-2)*20
		} else if point.Value < 5 {
			return 40 - (point.Value-3)*10
		} else {
			return 20
		}
	default:
		return 70
	}
}

func (a *SimpleHealthAssessor) calculateOverallHealth(metrics map[string]float64) float64 {
	if len(metrics) == 0 {
		return 0
	}
	
	var sum, weightSum float64
	for metric, score := range metrics {
		weight := a.getMetricWeight(metric)
		sum += score * weight
		weightSum += weight
	}
	
	if weightSum == 0 {
		return 0
	}
	
	return sum / weightSum
}

func (a *SimpleHealthAssessor) calculateOverallHealthForRUL(data []*TimeSeriesData) float64 {
	metrics := make(map[string]float64)
	for _, point := range data {
		if point.DeviceID == a.deviceID {
			metrics[point.Metric] = a.calculateMetricHealth(point)
		}
	}
	return a.calculateOverallHealth(metrics)
}

func (a *SimpleHealthAssessor) getMetricWeight(metric string) float64 {
	switch metric {
	case "temperature":
		return 0.3
	case "voltage":
		return 0.3
	case "soc":
		return 0.25
	case "vibration":
		return 0.15
	default:
		return 0.1
	}
}

func (a *SimpleHealthAssessor) determineHealthStatus(score float64) HealthStatus {
	switch {
	case score >= 90:
		return HealthStatusExcellent
	case score >= 75:
		return HealthStatusGood
	case score >= 60:
		return HealthStatusFair
	case score >= 40:
		return HealthStatusPoor
	default:
		return HealthStatusCritical
	}
}

func (a *SimpleHealthAssessor) identifyIssues(metrics map[string]float64) []string {
	var issues []string
	
	for metric, score := range metrics {
		if score < 60 {
			switch metric {
			case "temperature":
				issues = append(issues, "温度异常")
			case "voltage":
				issues = append(issues, "电压异常")
			case "soc":
				issues = append(issues, "电池SOC异常")
			case "vibration":
				issues = append(issues, "振动异常")
			default:
				issues = append(issues, fmt.Sprintf("%s异常", metric))
			}
		}
	}
	
	return issues
}

func (a *SimpleHealthAssessor) calculateConfidence(metrics map[string]float64) float64 {
	if len(metrics) == 0 {
		return 0.5
	}
	
	// 基于指标数量和一致性计算置信度
	baseConfidence := 0.5 + 0.3*(float64(len(metrics))/5.0)
	
	// 计算指标一致性
	if len(metrics) > 1 {
		var sum, sumSquared float64
		for _, score := range metrics {
			sum += score
			sumSquared += score * score
		}
		mean := sum / float64(len(metrics))
		variance := (sumSquared / float64(len(metrics))) - (mean * mean)
		stdDev := math.Sqrt(variance)
		
		// 一致性越高，置信度越高
		consistencyFactor := math.Max(0, 1.0-stdDev/30.0)
		baseConfidence *= consistencyFactor
	}
	
	return math.Min(0.99, baseConfidence)
}

func (a *SimpleHealthAssessor) getBaseRUL() float64 {
	// 根据设备类型返回基础RUL（小时）
	switch a.deviceType {
	case "solar":
		return 87600 // 10年
	case "wind":
		return 70080 // 8年
	case "battery":
		return 43800 // 5年
	default:
		return 58400 // 6.5年
	}
}

func (a *SimpleHealthAssessor) calculateRULInterval(rul, healthScore float64) [2]float64 {
	// 健康分数越高，置信区间越窄
	uncertainty := 0.5 - (healthScore/200.0)
	margin := rul * uncertainty
	
	return [2]float64{
		math.Max(0, rul-margin),
		rul + margin,
	}
}

func (a *SimpleHealthAssessor) determineHealthTrend() string {
	if len(a.healthHistory) < 3 {
		return "stable"
	}
	
	// 分析最近3次健康评估的趋势
	recent := a.healthHistory[len(a.healthHistory)-3:]
	trend := recent[2].HealthScore - recent[0].HealthScore
	
	if trend > 5 {
		return "improving"
	} else if trend < -5 {
		return "declining"
	} else {
		return "stable"
	}
}

func (a *SimpleHealthAssessor) calculateRULConfidence(healthScore float64) float64 {
	// 健康分数越高，RUL预测置信度越高
	return math.Min(0.99, 0.5+0.5*(healthScore/100.0))
}