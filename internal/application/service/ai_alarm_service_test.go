package service

import (
	"context"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type mockAlarmRuleRepository struct {
	createFn              func(ctx context.Context, rule *entity.AlarmRule) error
	updateFn              func(ctx context.Context, rule *entity.AlarmRule) error
	deleteFn              func(ctx context.Context, id string) error
	getByIDFn             func(ctx context.Context, id string) (*entity.AlarmRule, error)
	getByNameFn           func(ctx context.Context, name string) (*entity.AlarmRule, error)
	listFn                func(ctx context.Context, query *repository.AlarmRuleQuery) ([]*entity.AlarmRule, int64, error)
	getEnabledFn          func(ctx context.Context) ([]*entity.AlarmRule, error)
	getRulesByPointIDFn   func(ctx context.Context, pointID string) ([]*entity.AlarmRule, error)
	getRulesByDeviceIDFn  func(ctx context.Context, deviceID string) ([]*entity.AlarmRule, error)
	getRulesByStationIDFn func(ctx context.Context, stationID string) ([]*entity.AlarmRule, error)
}

func (m *mockAlarmRuleRepository) Create(ctx context.Context, rule *entity.AlarmRule) error {
	if m.createFn != nil {
		return m.createFn(ctx, rule)
	}
	return nil
}

func (m *mockAlarmRuleRepository) Update(ctx context.Context, rule *entity.AlarmRule) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, rule)
	}
	return nil
}

func (m *mockAlarmRuleRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockAlarmRuleRepository) GetByID(ctx context.Context, id string) (*entity.AlarmRule, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockAlarmRuleRepository) GetByName(ctx context.Context, name string) (*entity.AlarmRule, error) {
	if m.getByNameFn != nil {
		return m.getByNameFn(ctx, name)
	}
	return nil, nil
}

func (m *mockAlarmRuleRepository) List(ctx context.Context, query *repository.AlarmRuleQuery) ([]*entity.AlarmRule, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, query)
	}
	return nil, 0, nil
}

func (m *mockAlarmRuleRepository) GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error) {
	if m.getEnabledFn != nil {
		return m.getEnabledFn(ctx)
	}
	return nil, nil
}

func (m *mockAlarmRuleRepository) GetRulesByPointID(ctx context.Context, pointID string) ([]*entity.AlarmRule, error) {
	if m.getRulesByPointIDFn != nil {
		return m.getRulesByPointIDFn(ctx, pointID)
	}
	return nil, nil
}

func (m *mockAlarmRuleRepository) GetRulesByDeviceID(ctx context.Context, deviceID string) ([]*entity.AlarmRule, error) {
	if m.getRulesByDeviceIDFn != nil {
		return m.getRulesByDeviceIDFn(ctx, deviceID)
	}
	return nil, nil
}

func (m *mockAlarmRuleRepository) GetRulesByStationID(ctx context.Context, stationID string) ([]*entity.AlarmRule, error) {
	if m.getRulesByStationIDFn != nil {
		return m.getRulesByStationIDFn(ctx, stationID)
	}
	return nil, nil
}

func newTestAIAlarmService() AIAlarmService {
	return NewAIAlarmService(&mockAlarmRuleRepository{})
}

func TestAIAlarmService_SmartAlarm_CriticalThreshold(t *testing.T) {
	svc := newTestAIAlarmService()
	result, err := svc.SmartAlarm(context.Background(), "device-001", "temperature", 85.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.ShouldAlarm {
		t.Error("expected ShouldAlarm to be true")
	}
	if result.Severity != "critical" {
		t.Errorf("expected severity critical, got %s", result.Severity)
	}
	if result.Confidence != 0.95 {
		t.Errorf("expected confidence 0.95, got %f", result.Confidence)
	}
	if result.DeviceID != "device-001" {
		t.Errorf("expected deviceID device-001, got %s", result.DeviceID)
	}
	if result.Value != 85.0 {
		t.Errorf("expected value 85.0, got %f", result.Value)
	}
}

func TestAIAlarmService_SmartAlarm_WarningThreshold(t *testing.T) {
	svc := newTestAIAlarmService()
	result, err := svc.SmartAlarm(context.Background(), "device-001", "temperature", 65.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.ShouldAlarm {
		t.Error("expected ShouldAlarm to be true")
	}
	if result.Severity != "warning" {
		t.Errorf("expected severity warning, got %s", result.Severity)
	}
	if result.Confidence != 0.85 {
		t.Errorf("expected confidence 0.85, got %f", result.Confidence)
	}
}

func TestAIAlarmService_SmartAlarm_NormalValue(t *testing.T) {
	svc := newTestAIAlarmService()
	result, err := svc.SmartAlarm(context.Background(), "device-001", "temperature", 40.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShouldAlarm {
		t.Error("expected ShouldAlarm to be false")
	}
	if result.Severity != "" {
		t.Errorf("expected empty severity, got %s", result.Severity)
	}
	if result.Confidence != 0.9 {
		t.Errorf("expected confidence 0.9, got %f", result.Confidence)
	}
	if result.Reason != "指标正常" {
		t.Errorf("expected reason '指标正常', got %s", result.Reason)
	}
}

func TestAIAlarmService_SmartAlarm_UnknownMetric(t *testing.T) {
	svc := newTestAIAlarmService()
	result, err := svc.SmartAlarm(context.Background(), "device-001", "unknown_metric", 100.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShouldAlarm {
		t.Error("expected ShouldAlarm to be false for unknown metric")
	}
	if result.Confidence != 0.5 {
		t.Errorf("expected confidence 0.5, got %f", result.Confidence)
	}
	if result.Reason != "无预设阈值规则" {
		t.Errorf("expected reason '无预设阈值规则', got %s", result.Reason)
	}
}

func TestAIAlarmService_SmartAlarm_VoltageCritical(t *testing.T) {
	svc := newTestAIAlarmService()
	result, err := svc.SmartAlarm(context.Background(), "device-002", "voltage", 285.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.ShouldAlarm {
		t.Error("expected ShouldAlarm to be true")
	}
	if result.Severity != "critical" {
		t.Errorf("expected severity critical, got %s", result.Severity)
	}
}

func TestAIAlarmService_PredictAlarm(t *testing.T) {
	svc := newTestAIAlarmService()
	horizon := 15 * time.Minute
	result, err := svc.PredictAlarm(context.Background(), "device-001", "temperature", horizon)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DeviceID != "device-001" {
		t.Errorf("expected deviceID device-001, got %s", result.DeviceID)
	}
	if result.MetricName != "temperature" {
		t.Errorf("expected metricName temperature, got %s", result.MetricName)
	}
	if result.Probability <= 0 {
		t.Errorf("expected positive probability, got %f", result.Probability)
	}
	if result.Severity == "" {
		t.Error("expected non-empty severity")
	}
}

func TestAIAlarmService_PredictAlarm_UnknownMetric(t *testing.T) {
	svc := newTestAIAlarmService()
	horizon := 1 * time.Hour
	result, err := svc.PredictAlarm(context.Background(), "device-001", "unknown_metric", horizon)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Probability != 0.2 {
		t.Errorf("expected probability 0.2 for unknown metric, got %f", result.Probability)
	}
}

func TestAIAlarmService_CorrelateAlarms(t *testing.T) {
	svc := newTestAIAlarmService()
	correlations, err := svc.CorrelateAlarms(context.Background(), "station-001", 1*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(correlations) == 0 {
		t.Fatal("expected at least one correlation")
	}
	for _, c := range correlations {
		if len(c.AlarmIDs) == 0 {
			t.Error("expected non-empty AlarmIDs")
		}
		if len(c.DeviceIDs) == 0 {
			t.Error("expected non-empty DeviceIDs")
		}
		if c.Correlation <= 0 || c.Correlation > 1 {
			t.Errorf("expected correlation in (0,1], got %f", c.Correlation)
		}
		if c.Pattern == "" {
			t.Error("expected non-empty Pattern")
		}
	}
}

func TestCalculateAlarmCorrelation_ValidData(t *testing.T) {
	values1 := []float64{10, 20, 30, 40, 50}
	values2 := []float64{12, 22, 32, 42, 52}
	corr := CalculateAlarmCorrelation(values1, values2)
	if corr < 0.99 {
		t.Errorf("expected near-perfect correlation for linearly related data, got %f", corr)
	}
}

func TestCalculateAlarmCorrelation_InsufficientData(t *testing.T) {
	values1 := []float64{10}
	values2 := []float64{12}
	corr := CalculateAlarmCorrelation(values1, values2)
	if corr != 0 {
		t.Errorf("expected 0 correlation for insufficient data, got %f", corr)
	}
}

func TestCalculateAlarmCorrelation_LengthMismatch(t *testing.T) {
	values1 := []float64{10, 20, 30}
	values2 := []float64{12, 22}
	corr := CalculateAlarmCorrelation(values1, values2)
	if corr != 0 {
		t.Errorf("expected 0 correlation for length mismatch, got %f", corr)
	}
}

func TestCalculateAlarmCorrelation_ZeroVariance(t *testing.T) {
	values1 := []float64{10, 10, 10}
	values2 := []float64{12, 22, 32}
	corr := CalculateAlarmCorrelation(values1, values2)
	if corr != 0 {
		t.Errorf("expected 0 correlation for zero variance, got %f", corr)
	}
}

func TestCalculateAlarmCorrelation_NegativeCorrelation(t *testing.T) {
	values1 := []float64{10, 20, 30, 40, 50}
	values2 := []float64{50, 40, 30, 20, 10}
	corr := CalculateAlarmCorrelation(values1, values2)
	if corr >= 0 {
		t.Errorf("expected negative correlation, got %f", corr)
	}
}
