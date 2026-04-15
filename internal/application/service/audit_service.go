package service

import (
	"context"
	"fmt"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type AuditService struct {
	logRepo repository.OperationLogRepository
}

func NewAuditService(logRepo repository.OperationLogRepository) *AuditService {
	return &AuditService{logRepo: logRepo}
}

type LogOperationRequest struct {
	UserID       string
	Username     string
	Action       string
	ResourceType string
	ResourceID   string
	Details      entity.Details
	IPAddress    string
	UserAgent    string
}

func (s *AuditService) LogOperation(ctx context.Context, req *LogOperationRequest) error {
	log := entity.NewOperationLog(req.UserID, req.Username, req.Action)
	
	if req.ResourceType != "" {
		log.SetResource(req.ResourceType, req.ResourceID)
	}
	
	if req.Details != nil {
		log.SetDetails(req.Details)
	}
	
	log.SetRequestInfo(req.IPAddress, req.UserAgent)

	return s.logRepo.Create(ctx, log)
}

func (s *AuditService) GetLog(ctx context.Context, id string) (*entity.OperationLog, error) {
	return s.logRepo.GetByID(ctx, id)
}

func (s *AuditService) ListLogs(ctx context.Context, userID *string, action *string, startTime, endTime int64, page, pageSize int) (*OperationLogListResponse, error) {
	query := &repository.OperationLogQuery{
		Page:      page,
		PageSize:  pageSize,
		StartTime: startTime,
		EndTime:   endTime,
	}
	
	if userID != nil {
		query.UserID = *userID
	}
	
	if action != nil {
		query.Action = *action
	}
	
	logs, total, err := s.logRepo.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list logs: %w", err)
	}

	list := make([]*OperationLogResponse, len(logs))
	for i, log := range logs {
		list[i] = &OperationLogResponse{
			ID:           log.ID,
			UserID:       log.UserID,
			Username:     log.Username,
			Action:       log.Action,
			ResourceType: log.ResourceType,
			ResourceID:   log.ResourceID,
			Details:      log.Details,
			IPAddress:    log.IPAddress,
			UserAgent:    log.UserAgent,
			CreatedAt:    log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &OperationLogListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AuditService) GetUserLogs(ctx context.Context, userID string, page, pageSize int) (*OperationLogListResponse, error) {
	query := &repository.OperationLogQuery{
		Page:      page,
		PageSize:  pageSize,
		UserID:    userID,
	}
	
	logs, total, err := s.logRepo.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get user logs: %w", err)
	}

	list := make([]*OperationLogResponse, len(logs))
	for i, log := range logs {
		list[i] = &OperationLogResponse{
			ID:           log.ID,
			UserID:       log.UserID,
			Username:     log.Username,
			Action:       log.Action,
			ResourceType: log.ResourceType,
			ResourceID:   log.ResourceID,
			Details:      log.Details,
			IPAddress:    log.IPAddress,
			UserAgent:    log.UserAgent,
			CreatedAt:    log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &OperationLogListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AuditService) LogLogin(ctx context.Context, userID, username string, success bool, ipAddress, userAgent string) error {
	details := entity.Details{
		"success": success,
	}
	
	return s.LogOperation(ctx, &LogOperationRequest{
		UserID:    userID,
		Username:  username,
		Action:    entity.ActionLogin,
		Details:   details,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

func (s *AuditService) LogLogout(ctx context.Context, userID, username, ipAddress, userAgent string) error {
	return s.LogOperation(ctx, &LogOperationRequest{
		UserID:    userID,
		Username:  username,
		Action:    entity.ActionLogout,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

func (s *AuditService) LogCreateUser(ctx context.Context, operatorID, username string, newUser *entity.User, ipAddress, userAgent string) error {
	details := entity.Details{
		"new_user_id":       newUser.ID,
		"new_username":      newUser.Username,
		"new_user_email":    newUser.Email,
		"new_user_realname": newUser.RealName,
	}
	
	return s.LogOperation(ctx, &LogOperationRequest{
		UserID:       operatorID,
		Username:     username,
		Action:       entity.ActionCreateUser,
		ResourceType: entity.ResourceUser,
		ResourceID:   newUser.ID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	})
}

func (s *AuditService) LogUpdateUser(ctx context.Context, operatorID, username string, updatedUser *entity.User, changes map[string]interface{}, ipAddress, userAgent string) error {
	details := entity.Details{
		"updated_user_id": updatedUser.ID,
		"changes":         changes,
	}
	
	return s.LogOperation(ctx, &LogOperationRequest{
		UserID:       operatorID,
		Username:     username,
		Action:       entity.ActionUpdateUser,
		ResourceType: entity.ResourceUser,
		ResourceID:   updatedUser.ID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	})
}

func (s *AuditService) LogDeleteUser(ctx context.Context, operatorID, username string, deletedUser *entity.User, ipAddress, userAgent string) error {
	details := entity.Details{
		"deleted_user_id":       deletedUser.ID,
		"deleted_username":      deletedUser.Username,
		"deleted_user_email":    deletedUser.Email,
		"deleted_user_realname": deletedUser.RealName,
	}
	
	return s.LogOperation(ctx, &LogOperationRequest{
		UserID:       operatorID,
		Username:     username,
		Action:       entity.ActionDeleteUser,
		ResourceType: entity.ResourceUser,
		ResourceID:   deletedUser.ID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	})
}

func (s *AuditService) LogChangePassword(ctx context.Context, operatorID, username, targetUserID string, ipAddress, userAgent string) error {
	details := entity.Details{
		"target_user_id": targetUserID,
	}
	
	return s.LogOperation(ctx, &LogOperationRequest{
		UserID:       operatorID,
		Username:     username,
		Action:       entity.ActionChangePassword,
		ResourceType: entity.ResourceUser,
		ResourceID:   targetUserID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	})
}

func (s *AuditService) LogAssignRole(ctx context.Context, operatorID, username, targetUserID string, roleIDs []string, ipAddress, userAgent string) error {
	details := entity.Details{
		"target_user_id": targetUserID,
		"role_ids":       roleIDs,
	}
	
	return s.LogOperation(ctx, &LogOperationRequest{
		UserID:       operatorID,
		Username:     username,
		Action:       entity.ActionAssignRole,
		ResourceType: entity.ResourceUser,
		ResourceID:   targetUserID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	})
}

func (s *AuditService) LogCreateRole(ctx context.Context, operatorID, username string, newRole *entity.Role, ipAddress, userAgent string) error {
	details := entity.Details{
		"new_role_id":          newRole.ID,
		"new_role_code":        newRole.Code,
		"new_role_name":        newRole.Name,
		"new_role_description": newRole.Description,
	}
	
	return s.LogOperation(ctx, &LogOperationRequest{
		UserID:       operatorID,
		Username:     username,
		Action:       entity.ActionCreateRole,
		ResourceType: entity.ResourceRole,
		ResourceID:   newRole.ID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	})
}

func (s *AuditService) LogAssignPermission(ctx context.Context, operatorID, username, roleID string, permissionIDs []string, ipAddress, userAgent string) error {
	details := entity.Details{
		"role_id":        roleID,
		"permission_ids": permissionIDs,
	}
	
	return s.LogOperation(ctx, &LogOperationRequest{
		UserID:       operatorID,
		Username:     username,
		Action:       entity.ActionAssignPermission,
		ResourceType: entity.ResourceRole,
		ResourceID:   roleID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	})
}
