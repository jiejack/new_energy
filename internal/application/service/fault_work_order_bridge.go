package service

import (
	"context"
	"fmt"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type WorkOrderCreator interface {
	CreateFromFaultDetection(ctx context.Context, detection *entity.FaultDetectionResult) (string, error)
}

type FaultWorkOrderBridge struct {
	faultRepo        repository.FaultDetectionResultRepository
	workOrderCreator WorkOrderCreator
}

func NewFaultWorkOrderBridge(
	faultRepo repository.FaultDetectionResultRepository,
	workOrderCreator WorkOrderCreator,
) *FaultWorkOrderBridge {
	return &FaultWorkOrderBridge{
		faultRepo:        faultRepo,
		workOrderCreator: workOrderCreator,
	}
}

func (b *FaultWorkOrderBridge) CreateWorkOrderFromDetection(ctx context.Context, detectionID string) (string, error) {
	detection, err := b.faultRepo.GetByID(ctx, detectionID)
	if err != nil {
		return "", fmt.Errorf("detection not found: %w", err)
	}
	if detection.Severity == entity.FaultSeverityInfo {
		return "", nil
	}
	workOrderID, err := b.workOrderCreator.CreateFromFaultDetection(ctx, detection)
	if err != nil {
		return "", fmt.Errorf("create work order failed: %w", err)
	}
	if err := b.faultRepo.LinkWorkOrder(ctx, detectionID, workOrderID); err != nil {
		return "", fmt.Errorf("update detection work order id failed: %w", err)
	}
	return workOrderID, nil
}

func (b *FaultWorkOrderBridge) ShouldAutoCreateWorkOrder(severity entity.FaultSeverity) bool {
	return severity == entity.FaultSeverityFatal || severity == entity.FaultSeverityCritical
}

func (b *FaultWorkOrderBridge) GetWorkOrderSLA(severity entity.FaultSeverity) string {
	switch severity {
	case entity.FaultSeverityFatal:
		return "30min"
	case entity.FaultSeverityCritical:
		return "1h"
	case entity.FaultSeverityWarning:
		return "4h"
	default:
		return "24h"
	}
}
