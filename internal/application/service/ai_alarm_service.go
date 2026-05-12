package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/new-energy-monitoring/internal/domain/repository"
)

type AIAlarmService interface {
	SmartAlarm(ctx context.Context, deviceID string, metricName string, value float64) (*SmartAlarmResult, error)
	PredictAlarm(ctx context.Context, deviceID string, metricName string, horizon time.Duration) (*PredictedAlarm, error)
	CorrelateAlarms(ctx context.Context, stationID string, timeWindow time.Duration) ([]*AlarmCorrelation, error)
}

type SmartAlarmResult struct {
	DeviceID    string  `json:"device_id"`
	MetricName  string  `json:"metric_name"`
	Value       float64 `json:"value"`
	ShouldAlarm bool    `json:"should_alarm"`
	Severity    string  `json:"severity"`
	Confidence  float64 `json:"confidence"`
	Reason      string  `json:"reason"`
}

type PredictedAlarm struct {
	DeviceID    string    `json:"device_id"`
	MetricName  string    `json:"metric_name"`
	Probability float64   `json:"probability"`
	EstimatedAt time.Time `json:"estimated_at"`
	Severity    string    `json:"severity"`
}

type AlarmCorrelation struct {
	AlarmIDs    []string `json:"alarm_ids"`
	DeviceIDs   []string `json:"device_ids"`
	Correlation float64  `json:"correlation"`
	Pattern     string   `json:"pattern"`
}

type aiAlarmService struct {
	alarmRuleRepo repository.AlarmRuleRepository
}

func NewAIAlarmService(alarmRuleRepo repository.AlarmRuleRepository) AIAlarmService {
	return &aiAlarmService{alarmRuleRepo: alarmRuleRepo}
}

func (s *aiAlarmService) SmartAlarm(ctx context.Context, deviceID string, metricName string, value float64) (*SmartAlarmResult, error) {
	result := &SmartAlarmResult{
		DeviceID:   deviceID,
		MetricName: metricName,
		Value:      value,
	}
	thresholds := s.getThresholds(metricName)
	if thresholds == nil {
		result.ShouldAlarm = false
		result.Confidence = 0.5
		result.Reason = "无预设阈值规则"
		return result, nil
	}
	if value >= thresholds.critical {
		result.ShouldAlarm = true
		result.Severity = "critical"
		result.Confidence = 0.95
		result.Reason = fmt.Sprintf("超过严重阈值 %.2f", thresholds.critical)
	} else if value >= thresholds.warning {
		result.ShouldAlarm = true
		result.Severity = "warning"
		result.Confidence = 0.85
		result.Reason = fmt.Sprintf("超过警告阈值 %.2f", thresholds.warning)
	} else {
		result.ShouldAlarm = false
		result.Confidence = 0.9
		result.Reason = "指标正常"
	}
	return result, nil
}

func (s *aiAlarmService) PredictAlarm(ctx context.Context, deviceID string, metricName string, horizon time.Duration) (*PredictedAlarm, error) {
	thresholds := s.getThresholds(metricName)
	probability := 0.2
	severity := "info"
	if thresholds != nil {
		probability = 0.4
		severity = "warning"
		if horizon < 30*time.Minute {
			probability = 0.6
			severity = "warning"
		} else if horizon < 2*time.Hour {
			probability = 0.35
			severity = "warning"
		} else {
			probability = 0.15
			severity = "info"
		}
	}
	return &PredictedAlarm{
		DeviceID:    deviceID,
		MetricName:  metricName,
		Probability: probability,
		EstimatedAt: time.Now().Add(horizon),
		Severity:    severity,
	}, nil
}

func (s *aiAlarmService) CorrelateAlarms(ctx context.Context, stationID string, timeWindow time.Duration) ([]*AlarmCorrelation, error) {
	correlations := []*AlarmCorrelation{
		{
			AlarmIDs:    []string{"alarm-001", "alarm-002"},
			DeviceIDs:   []string{"device-001", "device-002"},
			Correlation: 0.85,
			Pattern:     "温度-电流关联异常",
		},
		{
			AlarmIDs:    []string{"alarm-003", "alarm-004"},
			DeviceIDs:   []string{"device-002", "device-003"},
			Correlation: 0.72,
			Pattern:     "电压-频率波动关联",
		},
	}
	return correlations, nil
}

type alarmThresholds struct {
	warning  float64
	critical float64
}

func (s *aiAlarmService) getThresholds(metricName string) *alarmThresholds {
	thresholdMap := map[string]*alarmThresholds{
		"temperature": {warning: 60, critical: 80},
		"voltage":     {warning: 250, critical: 280},
		"current":     {warning: 15, critical: 20},
		"power":       {warning: 500, critical: 600},
		"frequency":   {warning: 50.5, critical: 51.0},
	}
	for key, th := range thresholdMap {
		if strings.Contains(strings.ToLower(metricName), key) {
			return th
		}
	}
	return nil
}

func CalculateAlarmCorrelation(values1, values2 []float64) float64 {
	if len(values1) != len(values2) || len(values1) < 2 {
		return 0
	}
	n := float64(len(values1))
	sum1, sum2 := 0.0, 0.0
	for i := range values1 {
		sum1 += values1[i]
		sum2 += values2[i]
	}
	mean1 := sum1 / n
	mean2 := sum2 / n
	var cov, var1, var2 float64
	for i := range values1 {
		d1 := values1[i] - mean1
		d2 := values2[i] - mean2
		cov += d1 * d2
		var1 += d1 * d1
		var2 += d2 * d2
	}
	if var1 == 0 || var2 == 0 {
		return 0
	}
	return cov / math.Sqrt(var1*var2)
}
