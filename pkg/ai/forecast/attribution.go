package forecast

type AttributionFactor struct {
	Name        string  `json:"name"`
	Contribution float64 `json:"contribution"`
	Description string  `json:"description"`
}

type AttributionResult struct {
	StationID     string              `json:"station_id"`
	PrimaryCause  string              `json:"primary_cause"`
	Confidence    float64             `json:"confidence"`
	Factors       []AttributionFactor `json:"factors"`
	Suggestion    string              `json:"suggestion"`
	Deviation     float64             `json:"deviation"`
}

type AttributionAnalyzer interface {
	Analyze(stationID string, predictedPower, actualPower float64) (*AttributionResult, error)
}

type attributionAnalyzer struct{}

func NewAttributionAnalyzer() AttributionAnalyzer {
	return &attributionAnalyzer{}
}

func (a *attributionAnalyzer) Analyze(stationID string, predictedPower, actualPower float64) (*AttributionResult, error) {
	deviation := 0.0
	if actualPower != 0 {
		deviation = (predictedPower - actualPower) / actualPower * 100
	}

	var primaryCause string
	var factors []AttributionFactor
	var suggestion string

	absDev := abs(deviation)
	if absDev > 30 {
		primaryCause = "nwp"
		factors = []AttributionFactor{
			{Name: "nwp_irradiance", Contribution: 0.5, Description: "NWP辐照度预报偏差"},
			{Name: "nwp_temperature", Contribution: 0.3, Description: "NWP温度预报偏差"},
			{Name: "model_error", Contribution: 0.2, Description: "模型预测误差"},
		}
		suggestion = "建议检查NWP数据源质量，考虑引入多源气象数据融合"
	} else if absDev > 15 {
		primaryCause = "device"
		factors = []AttributionFactor{
			{Name: "device_degradation", Contribution: 0.4, Description: "设备性能衰减"},
			{Name: "partial_shading", Contribution: 0.3, Description: "局部遮挡影响"},
			{Name: "weather", Contribution: 0.3, Description: "天气突变"},
		}
		suggestion = "建议检查设备运行状态，排查遮挡问题"
	} else {
		primaryCause = "weather"
		factors = []AttributionFactor{
			{Name: "cloud_cover", Contribution: 0.5, Description: "云量变化"},
			{Name: "aerosol", Contribution: 0.3, Description: "气溶胶影响"},
			{Name: "measurement", Contribution: 0.2, Description: "测量误差"},
		}
		suggestion = "偏差在正常范围内，持续监控"
	}

	confidence := 0.8
	if absDev > 20 {
		confidence = 0.9
	} else if absDev < 5 {
		confidence = 0.7
	}

	return &AttributionResult{
		StationID:    stationID,
		PrimaryCause: primaryCause,
		Confidence:   confidence,
		Factors:      factors,
		Suggestion:   suggestion,
		Deviation:    deviation,
	}, nil
}
