package fault

import (
	"context"
	"fmt"
	"time"
)

type RuleBasedClassifier struct {
	deviceType    string
	faultRules    map[string][]FaultRule
	trainedData   []*FaultLabeledData
	createdAt     time.Time
	trainedAt     *time.Time
}

type FaultRule struct {
	FaultType      string
	FaultCode      string
	Description    string
	Recommendations []string
	Condition      func(anomaly *Anomaly) bool
}

func NewRuleBasedClassifier(deviceType string) *RuleBasedClassifier {
	classifier := &RuleBasedClassifier{
		deviceType:  deviceType,
		faultRules:  make(map[string][]FaultRule),
		createdAt:   time.Now(),
	}
	
	// 根据设备类型加载默认规则
	classifier.loadDefaultRules()
	
	return classifier
}

func (c *RuleBasedClassifier) loadDefaultRules() {
	switch c.deviceType {
	case "solar":
		c.loadSolarRules()
	case "wind":
		c.loadWindRules()
	case "battery":
		c.loadBatteryRules()
	}
}

func (c *RuleBasedClassifier) loadSolarRules() {
	c.faultRules["temperature"] = []FaultRule{
		{
			FaultType:      "overheating",
			FaultCode:      "SOL-001",
			Description:    "光伏面板温度过高",
			Recommendations: []string{"检查散热系统", "调整面板角度", "增加通风"},
			Condition: func(a *Anomaly) bool {
				return a.Value > 60 && a.Severity >= SeverityHigh
			},
		},
		{
			FaultType:      "low_temperature",
			FaultCode:      "SOL-002",
			Description:    "光伏面板温度过低",
			Recommendations: []string{"检查环境温度传感器", "确认保温措施"},
			Condition: func(a *Anomaly) bool {
				return a.Value < -10 && a.Severity >= SeverityMedium
			},
		},
	}
	
	c.faultRules["voltage"] = []FaultRule{
		{
			FaultType:      "over_voltage",
			FaultCode:      "SOL-003",
			Description:    "光伏系统电压过高",
			Recommendations: []string{"检查逆变器", "调整系统配置", "检查电网连接"},
			Condition: func(a *Anomaly) bool {
				return a.Value > 800 && a.Severity >= SeverityHigh
			},
		},
		{
			FaultType:      "under_voltage",
			FaultCode:      "SOL-004",
			Description:    "光伏系统电压过低",
			Recommendations: []string{"检查面板连接", "清理面板灰尘", "检查阴影遮挡"},
			Condition: func(a *Anomaly) bool {
				return a.Value < 200 && a.Severity >= SeverityMedium
			},
		},
	}
}

func (c *RuleBasedClassifier) loadWindRules() {
	c.faultRules["wind_speed"] = []FaultRule{
		{
			FaultType:      "high_wind",
			FaultCode:      "WIND-001",
			Description:    "风速过高",
			Recommendations: []string{"启动限速保护", "检查风机状态", "准备停机"},
			Condition: func(a *Anomaly) bool {
				return a.Value > 25 && a.Severity >= SeverityHigh
			},
		},
		{
			FaultType:      "low_wind",
			FaultCode:      "WIND-002",
			Description:    "风速过低",
			Recommendations: []string{"检查风速传感器", "优化风机角度"},
			Condition: func(a *Anomaly) bool {
				return a.Value < 2 && a.Severity >= SeverityMedium
			},
		},
	}
	
	c.faultRules["vibration"] = []FaultRule{
		{
			FaultType:      "excessive_vibration",
			FaultCode:      "WIND-003",
			Description:    "风机振动过大",
			Recommendations: []string{"检查轴承状态", "平衡叶轮", "检查基础固定"},
			Condition: func(a *Anomaly) bool {
				return a.Value > 5 && a.Severity >= SeverityHigh
			},
		},
	}
}

func (c *RuleBasedClassifier) loadBatteryRules() {
	c.faultRules["temperature"] = []FaultRule{
		{
			FaultType:      "battery_overheating",
			FaultCode:      "BAT-001",
			Description:    "电池温度过高",
			Recommendations: []string{"检查散热系统", "减少充电电流", "检查电池是否老化"},
			Condition: func(a *Anomaly) bool {
				return a.Value > 45 && a.Severity >= SeverityHigh
			},
		},
	}
	
	c.faultRules["soc"] = []FaultRule{
		{
			FaultType:      "over_discharge",
			FaultCode:      "BAT-002",
			Description:    "电池过度放电",
			Recommendations: []string{"立即充电", "检查负载", "评估电池健康状态"},
			Condition: func(a *Anomaly) bool {
				return a.Value < 20 && a.Severity >= SeverityHigh
			},
		},
		{
			FaultType:      "over_charge",
			FaultCode:      "BAT-003",
			Description:    "电池过度充电",
			Recommendations: []string{"停止充电", "检查充电系统", "评估电池健康状态"},
			Condition: func(a *Anomaly) bool {
				return a.Value > 95 && a.Severity >= SeverityHigh
			},
		},
	}
}

func (c *RuleBasedClassifier) Classify(ctx context.Context, anomaly *Anomaly) (*FaultClassification, error) {
	// 查找适合的规则
	rules, ok := c.faultRules[anomaly.Metric]
	if !ok {
		// 返回默认分类
		return &FaultClassification{
			ID:            fmt.Sprintf("fault-%s", anomaly.ID),
			AnomalyID:     anomaly.ID,
			FaultType:     "unknown",
			FaultCode:     "UNKNOWN-001",
			Description:   "未知故障类型",
			Confidence:    0.5,
			Recommendations: []string{"进一步检查设备状态"},
			AdditionalInfo: map[string]interface{}{"device_type": c.deviceType},
		}, nil
	}
	
	// 应用规则
	for _, rule := range rules {
		if rule.Condition(anomaly) {
			return &FaultClassification{
				ID:            fmt.Sprintf("fault-%s", anomaly.ID),
				AnomalyID:     anomaly.ID,
				FaultType:     rule.FaultType,
				FaultCode:     rule.FaultCode,
				Description:   rule.Description,
				Confidence:    c.calculateConfidence(anomaly, rule),
				Recommendations: rule.Recommendations,
				AdditionalInfo: map[string]interface{}{"device_type": c.deviceType, "rule_applied": rule.FaultCode},
			}, nil
		}
	}
	
	// 没有匹配的规则
	return &FaultClassification{
		ID:            fmt.Sprintf("fault-%s", anomaly.ID),
		AnomalyID:     anomaly.ID,
		FaultType:     "unclassified",
		FaultCode:     "UNCLASSIFIED-001",
		Description:   "未分类故障",
		Confidence:    0.3,
		Recommendations: []string{"手动检查设备状态"},
		AdditionalInfo: map[string]interface{}{"device_type": c.deviceType},
	}, nil
}

func (c *RuleBasedClassifier) Train(ctx context.Context, data []*FaultLabeledData) error {
	c.trainedData = data
	now := time.Now()
	c.trainedAt = &now
	
	// 可以根据训练数据调整规则
	// 这里简单实现，实际项目中可能需要更复杂的规则学习
	
	return nil
}

func (c *RuleBasedClassifier) GetClassifierInfo() *ClassifierInfo {
	return &ClassifierInfo{
		ClassifierID:   fmt.Sprintf("classifier-%s", c.deviceType),
		ClassifierType: "RuleBasedClassifier",
		Version:        "1.0.0",
		CreatedAt:      c.createdAt,
		TrainedAt:      c.trainedAt,
		Parameters: map[string]interface{}{
			"device_type": c.deviceType,
			"rule_count":  len(c.faultRules),
		},
		Metrics: map[string]float64{
			"trained_data_points": float64(len(c.trainedData)),
		},
		Status: "active",
	}
}

func (c *RuleBasedClassifier) calculateConfidence(anomaly *Anomaly, rule FaultRule) float64 {
	// 基于异常严重程度计算置信度
	switch anomaly.Severity {
	case SeverityCritical:
		return 0.95
	case SeverityHigh:
		return 0.85
	case SeverityMedium:
		return 0.75
	default:
		return 0.65
	}
}

type MLBasedClassifier struct {
	deviceType    string
	trainedData   []*FaultLabeledData
	createdAt     time.Time
	trainedAt     *time.Time
}

func NewMLBasedClassifier(deviceType string) *MLBasedClassifier {
	return &MLBasedClassifier{
		deviceType:  deviceType,
		createdAt:   time.Now(),
	}
}

func (c *MLBasedClassifier) Classify(ctx context.Context, anomaly *Anomaly) (*FaultClassification, error) {
	// 这里实现机器学习-based的分类
	// 简化实现，实际项目中可能需要更复杂的ML模型
	
	return &FaultClassification{
		ID:            fmt.Sprintf("fault-%s", anomaly.ID),
		AnomalyID:     anomaly.ID,
		FaultType:     "ml_classified",
		FaultCode:     "ML-001",
		Description:   "机器学习分类故障",
		Confidence:    0.8,
		Recommendations: []string{"根据模型建议处理"},
		AdditionalInfo: map[string]interface{}{"device_type": c.deviceType, "classifier": "MLBased"},
	}, nil
}

func (c *MLBasedClassifier) Train(ctx context.Context, data []*FaultLabeledData) error {
	c.trainedData = data
	now := time.Now()
	c.trainedAt = &now
	
	// 训练机器学习模型
	// 简化实现
	
	return nil
}

func (c *MLBasedClassifier) GetClassifierInfo() *ClassifierInfo {
	return &ClassifierInfo{
		ClassifierID:   fmt.Sprintf("classifier-ml-%s", c.deviceType),
		ClassifierType: "MLBasedClassifier",
		Version:        "1.0.0",
		CreatedAt:      c.createdAt,
		TrainedAt:      c.trainedAt,
		Parameters: map[string]interface{}{
			"device_type": c.deviceType,
		},
		Metrics: map[string]float64{
			"trained_data_points": float64(len(c.trainedData)),
		},
		Status: "active",
	}
}