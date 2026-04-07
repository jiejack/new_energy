package service

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type OperationLogService struct {
	logRepo repository.OperationLogRepository
}

func NewOperationLogService(logRepo repository.OperationLogRepository) *OperationLogService {
	return &OperationLogService{logRepo: logRepo}
}

type CreateOperationLogRequest struct {
	UserID       string                 `json:"user_id" binding:"required"`
	Username     string                 `json:"username" binding:"required"`
	Action       string                 `json:"action" binding:"required"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Details      map[string]interface{} `json:"details"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
}

type OperationLogResponse struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Username     string                 `json:"username"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Details      map[string]interface{} `json:"details"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	CreatedAt    string                 `json:"created_at"`
}

type OperationLogListResponse struct {
	List     []*OperationLogResponse `json:"list"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}

func (s *OperationLogService) CreateLog(ctx context.Context, req *CreateOperationLogRequest) (*OperationLogResponse, error) {
	log := entity.NewOperationLog(req.UserID, req.Username, req.Action)
	log.SetResource(req.ResourceType, req.ResourceID)
	if req.Details != nil {
		log.SetDetails(req.Details)
	}
	log.SetRequestInfo(req.IPAddress, req.UserAgent)

	if err := s.logRepo.Create(ctx, log); err != nil {
		return nil, err
	}

	return s.toResponse(log), nil
}

func (s *OperationLogService) GetLog(ctx context.Context, id string) (*OperationLogResponse, error) {
	log, err := s.logRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toResponse(log), nil
}

func (s *OperationLogService) ListLogs(ctx context.Context, query *repository.OperationLogQuery) (*OperationLogListResponse, error) {
	logs, total, err := s.logRepo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	list := make([]*OperationLogResponse, len(logs))
	for i, log := range logs {
		list[i] = s.toResponse(log)
	}

	return &OperationLogListResponse{
		List:     list,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (s *OperationLogService) DeleteOldLogs(ctx context.Context, days int) (int64, error) {
	before := time.Now().AddDate(0, 0, -days).Unix()
	return s.logRepo.DeleteBefore(ctx, before)
}

func (s *OperationLogService) toResponse(log *entity.OperationLog) *OperationLogResponse {
	return &OperationLogResponse{
		ID:           log.ID,
		UserID:       log.UserID,
		Username:     log.Username,
		Action:       log.Action,
		ResourceType: log.ResourceType,
		ResourceID:   log.ResourceID,
		Details:      log.Details,
		IPAddress:    log.IPAddress,
		UserAgent:    log.UserAgent,
		CreatedAt:    log.CreatedAt.Format(time.RFC3339),
	}
}
