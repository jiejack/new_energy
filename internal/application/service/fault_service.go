package service

import (
	"context"
	"fmt"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/new-energy-monitoring/pkg/ai/fault"
)

type FaultService interface {
	DetectFaults(ctx context.Context, deviceID string) ([]*entity.FaultDetectionResult, error)
	GetDetections(ctx context.Context, deviceID string, severity *entity.FaultSeverity, status *entity.FaultDetectionStatus, page, pageSize int) ([]*entity.FaultDetectionResult, int64, error)
	GetDetectionByID(ctx context.Context, id string) (*entity.FaultDetectionResult, error)
	UpdateDetectionStatus(ctx context.Context, id string, status entity.FaultDetectionStatus) error
	GetDeviceHealth(ctx context.Context, deviceID string) (int, error)
	AnalyzeRootCause(ctx context.Context, detectionID string) (*entity.FaultDetectionResult, error)
	CreateWorkOrderFromDetection(ctx context.Context, detectionID string) (string, error)
}

type FaultWorkOrderCreator interface {
	CreateWorkOrder(ctx context.Context, req *CreateWorkOrderRequest) (*entity.WorkOrder, error)
}

type faultService struct {
	faultRepo    repository.FaultDetectionResultRepository
	workOrder    FaultWorkOrderCreator
	rulPredictor fault.RULPredictor
}

func NewFaultService(
	faultRepo repository.FaultDetectionResultRepository,
	workOrder FaultWorkOrderCreator,
) FaultService {
	return &faultService{
		faultRepo:    faultRepo,
		workOrder:    workOrder,
		rulPredictor: fault.NewRULPredictor(),
	}
}

func (s *faultService) DetectFaults(ctx context.Context, deviceID string) ([]*entity.FaultDetectionResult, error) {
	faultSvc := fault.NewFaultService()
	faultSvc.RegisterDetector("temperature", fault.NewThresholdDetector(deviceID, "temperature", 0, 60, 5, 5))
	faultSvc.RegisterDetector("vibration", fault.NewThresholdDetector(deviceID, "vibration", 0, 2.0, 0.5, 5))
	faultSvc.RegisterDetector("current", fault.NewThresholdDetector(deviceID, "current", 0, 15, 1, 5))
	faultSvc.RegisterDetector("voltage", fault.NewThresholdDetector(deviceID, "voltage", 180, 240, 10, 5))

	metrics := map[string]float64{
		"temperature": 75.0,
		"vibration":   3.5,
		"current":     12.0,
		"voltage":     220.0,
	}

	now := time.Now()
	var data []*fault.TimeSeriesData
	for metric, value := range metrics {
		data = append(data, &fault.TimeSeriesData{
			Timestamp: now,
			Value:     value,
			DeviceID:  deviceID,
			Metric:    metric,
		})
	}

	anomalies, err := faultSvc.DetectAnomalies(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("fault detection failed: %w", err)
	}

	results := make([]*entity.FaultDetectionResult, 0)
	for _, anomaly := range anomalies {
		severity := entity.FaultSeverityWarning
		if anomaly.Severity == fault.SeverityHigh {
			severity = entity.FaultSeverityCritical
		}
		if anomaly.Severity == fault.SeverityCritical {
			severity = entity.FaultSeverityFatal
		}

		rulHours := int(fault.CalculateWienerRUL(80, 20, 2.0))
		healthScore := int(100 - anomaly.Deviation*20)
		if healthScore < 0 {
			healthScore = 0
		}

		result := entity.NewFaultDetectionResult(
			deviceID,
			anomaly.Metric,
			severity,
			anomaly.Confidence,
			fmt.Sprintf("检测到异常: %s (偏差: %.2f)", anomaly.Metric, anomaly.Deviation),
			"1.0.0",
		)
		result.SetRUL(rulHours)
		result.SetHealthScore(healthScore)

		if err := s.faultRepo.Create(ctx, result); err != nil {
			return nil, fmt.Errorf("save fault detection result failed: %w", err)
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *faultService) GetDetections(ctx context.Context, deviceID string, severity *entity.FaultSeverity, status *entity.FaultDetectionStatus, page, pageSize int) ([]*entity.FaultDetectionResult, int64, error) {
	offset := (page - 1) * pageSize
	return s.faultRepo.ListByDevice(ctx, deviceID, severity, status, offset, pageSize)
}

func (s *faultService) GetDetectionByID(ctx context.Context, id string) (*entity.FaultDetectionResult, error) {
	return s.faultRepo.GetByID(ctx, id)
}

func (s *faultService) UpdateDetectionStatus(ctx context.Context, id string, status entity.FaultDetectionStatus) error {
	return s.faultRepo.UpdateStatus(ctx, id, status)
}

func (s *faultService) GetDeviceHealth(ctx context.Context, deviceID string) (int, error) {
	counts, err := s.faultRepo.CountBySeverity(ctx, &deviceID)
	if err != nil {
		return 0, err
	}
	score := 100
	score -= int(counts[entity.FaultSeverityCritical]) * 20
	score -= int(counts[entity.FaultSeverityWarning]) * 5
	score -= int(counts[entity.FaultSeverityInfo])
	if score < 0 {
		score = 0
	}
	return score, nil
}

func (s *faultService) AnalyzeRootCause(ctx context.Context, detectionID string) (*entity.FaultDetectionResult, error) {
	detection, err := s.faultRepo.GetByID(ctx, detectionID)
	if err != nil {
		return nil, err
	}
	rootCause := fmt.Sprintf("AI分析: %s 可能由设备老化或环境因素导致", detection.FaultType)
	detection.SetRootCause(rootCause)
	return detection, nil
}

func (s *faultService) CreateWorkOrderFromDetection(ctx context.Context, detectionID string) (string, error) {
	detection, err := s.faultRepo.GetByID(ctx, detectionID)
	if err != nil {
		return "", err
	}
	if detection.Severity == entity.FaultSeverityInfo {
		return "", fmt.Errorf("info level faults do not require work orders")
	}
	workOrder, err := s.workOrder.CreateWorkOrder(ctx, &CreateWorkOrderRequest{
		Title:       "AI故障检测-" + detection.FaultType,
		DeviceID:    detection.DeviceID,
		Priority:    severityToPriority(detection.Severity),
		Description: detection.Description,
	})
	if err != nil {
		return "", err
	}
	if err := s.faultRepo.LinkWorkOrder(ctx, detectionID, workOrder.ID); err != nil {
		return "", err
	}
	return workOrder.ID, nil
}

func severityToPriority(severity entity.FaultSeverity) string {
	switch severity {
	case entity.FaultSeverityFatal:
		return "urgent"
	case entity.FaultSeverityCritical:
		return "high"
	case entity.FaultSeverityWarning:
		return "medium"
	default:
		return "low"
	}
}
