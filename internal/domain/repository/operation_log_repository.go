package repository

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
)

type OperationLogRepository interface {
	Create(ctx context.Context, log *entity.OperationLog) error
	GetByID(ctx context.Context, id string) (*entity.OperationLog, error)
	List(ctx context.Context, query *OperationLogQuery) ([]*entity.OperationLog, int64, error)
	DeleteBefore(ctx context.Context, before int64) (int64, error)
}

type OperationLogQuery struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	UserID    string `form:"user_id"`
	Username  string `form:"username"`
	Action    string `form:"action"`
	StartTime int64  `form:"start_time"`
	EndTime   int64  `form:"end_time"`
}
