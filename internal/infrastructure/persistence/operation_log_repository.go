package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type operationLogRepository struct {
	db *Database
}

func NewOperationLogRepository(db *Database) repository.OperationLogRepository {
	return &operationLogRepository{db: db}
}

func (r *operationLogRepository) Create(ctx context.Context, log *entity.OperationLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *operationLogRepository) GetByID(ctx context.Context, id string) (*entity.OperationLog, error) {
	var log entity.OperationLog
	err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *operationLogRepository) List(ctx context.Context, query *repository.OperationLogQuery) ([]*entity.OperationLog, int64, error) {
	var logs []*entity.OperationLog
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.OperationLog{})

	if query.UserID != "" {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.Username != "" {
		db = db.Where("username LIKE ?", "%"+query.Username+"%")
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.StartTime > 0 {
		db = db.Where("created_at >= ?", time.Unix(query.StartTime, 0))
	}
	if query.EndTime > 0 {
		db = db.Where("created_at <= ?", time.Unix(query.EndTime, 0))
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *operationLogRepository) DeleteBefore(ctx context.Context, before int64) (int64, error) {
	result := r.db.WithContext(ctx).Where("created_at < ?", time.Unix(before, 0)).Delete(&entity.OperationLog{})
	return result.RowsAffected, result.Error
}
