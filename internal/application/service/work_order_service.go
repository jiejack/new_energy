package service

import (
	"context"
	"fmt"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type WorkOrderService struct {
	workOrderRepo repository.WorkOrderRepository
	deviceRepo    repository.DeviceRepository
}

func NewWorkOrderService(workOrderRepo repository.WorkOrderRepository, deviceRepo repository.DeviceRepository) *WorkOrderService {
	return &WorkOrderService{
		workOrderRepo: workOrderRepo,
		deviceRepo:    deviceRepo,
	}
}

func (s *WorkOrderService) CreateWorkOrder(ctx context.Context, req *CreateWorkOrderRequest) (*entity.WorkOrder, error) {
	// 验证设备是否存在
	if req.DeviceID != "" {
		_, err := s.deviceRepo.GetByID(ctx, req.DeviceID)
		if err != nil {
			return nil, fmt.Errorf("device not found: %w", err)
		}
	}

	workOrder := &entity.WorkOrder{
		DeviceID:    req.DeviceID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      "open",
		Assignee:    req.Assignee,
		DueDate:     req.DueDate,
		CreatedBy:   req.CreatedBy,
		Notes:       req.Notes,
	}

	if err := s.workOrderRepo.Create(ctx, workOrder); err != nil {
		return nil, fmt.Errorf("failed to create work order: %w", err)
	}

	return workOrder, nil
}

func (s *WorkOrderService) UpdateWorkOrder(ctx context.Context, id string, req *UpdateWorkOrderRequest) (*entity.WorkOrder, error) {
	workOrder, err := s.workOrderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("work order not found: %w", err)
	}

	if req.Title != "" {
		workOrder.Title = req.Title
	}
	if req.Description != "" {
		workOrder.Description = req.Description
	}
	if req.Priority != "" {
		workOrder.Priority = req.Priority
	}
	if req.Status != "" {
		workOrder.Status = req.Status
		
		// 更新开始和结束时间
		if req.Status == "in_progress" && workOrder.StartDate == nil {
			now := time.Now()
			workOrder.StartDate = &now
		} else if req.Status == "completed" && workOrder.EndDate == nil {
			now := time.Now()
			workOrder.EndDate = &now
		}
	}
	if req.Assignee != "" {
		workOrder.Assignee = req.Assignee
	}
	if req.DueDate != nil {
		workOrder.DueDate = req.DueDate
	}
	if req.Notes != "" {
		workOrder.Notes = req.Notes
	}

	if err := s.workOrderRepo.Update(ctx, workOrder); err != nil {
		return nil, fmt.Errorf("failed to update work order: %w", err)
	}

	return workOrder, nil
}

func (s *WorkOrderService) DeleteWorkOrder(ctx context.Context, id string) error {
	workOrder, err := s.workOrderRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("work order not found: %w", err)
	}

	return s.workOrderRepo.Delete(ctx, workOrder.ID)
}

func (s *WorkOrderService) GetWorkOrder(ctx context.Context, id string) (*entity.WorkOrder, error) {
	return s.workOrderRepo.GetByID(ctx, id)
}

func (s *WorkOrderService) ListWorkOrders(ctx context.Context, filter *WorkOrderFilter) ([]*entity.WorkOrder, error) {
	return s.workOrderRepo.List(ctx, filter)
}

func (s *WorkOrderService) GetWorkOrdersByDevice(ctx context.Context, deviceID string) ([]*entity.WorkOrder, error) {
	filter := &WorkOrderFilter{
		DeviceID: &deviceID,
	}
	return s.workOrderRepo.List(ctx, filter)
}

func (s *WorkOrderService) GetWorkOrderStats(ctx context.Context) (map[string]interface{}, error) {
	total, err := s.workOrderRepo.Count(ctx, nil)
	if err != nil {
		return nil, err
	}

	open, err := s.workOrderRepo.Count(ctx, &WorkOrderFilter{Status: "open"})
	if err != nil {
		return nil, err
	}

	inProgress, err := s.workOrderRepo.Count(ctx, &WorkOrderFilter{Status: "in_progress"})
	if err != nil {
		return nil, err
	}

	completed, err := s.workOrderRepo.Count(ctx, &WorkOrderFilter{Status: "completed"})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":       total,
		"open":        open,
		"in_progress": inProgress,
		"completed":   completed,
	}, nil
}

type CreateWorkOrderRequest struct {
	DeviceID    string     `json:"device_id"`
	Type        string     `json:"type" binding:"required"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Assignee    string     `json:"assignee"`
	DueDate     *time.Time `json:"due_date"`
	CreatedBy   string     `json:"created_by"`
	Notes       string     `json:"notes"`
}

type UpdateWorkOrderRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"`
	Assignee    string     `json:"assignee"`
	DueDate     *time.Time `json:"due_date"`
	Notes       string     `json:"notes"`
}

type WorkOrderFilter struct {
	DeviceID  *string `json:"device_id"`
	Type      *string `json:"type"`
	Status    string  `json:"status"`
	Priority  *string `json:"priority"`
	Assignee  *string `json:"assignee"`
	CreatedBy *string `json:"created_by"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}
