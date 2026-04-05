package repository

import (
	"release-manager/internal/model"

	"gorm.io/gorm"
)

type OperationLogRepository struct {
	db *gorm.DB
}

func NewOperationLogRepository(db *gorm.DB) *OperationLogRepository {
	return &OperationLogRepository{db: db}
}

func (r *OperationLogRepository) Create(log *model.OperationLog) error {
	return r.db.Create(log).Error
}

func (r *OperationLogRepository) FindAll(page, pageSize int, userID *uint, action string) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64

	query := r.db.Model(&model.OperationLog{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, total, err
}
