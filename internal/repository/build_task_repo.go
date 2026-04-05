package repository

import (
	"release-manager/internal/model"

	"gorm.io/gorm"
)

type BuildTaskRepository struct {
	db *gorm.DB
}

func NewBuildTaskRepository(db *gorm.DB) *BuildTaskRepository {
	return &BuildTaskRepository{db: db}
}

func (r *BuildTaskRepository) Create(task *model.BuildTask) error {
	return r.db.Create(task).Error
}

func (r *BuildTaskRepository) FindByUserID(userID uint, page, pageSize int) ([]model.BuildTask, int64, error) {
	var tasks []model.BuildTask
	var total int64

	r.db.Model(&model.BuildTask{}).Where("user_id = ?", userID).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.Where("user_id = ?", userID).
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, total, err
}

func (r *BuildTaskRepository) FindByID(id uint) (*model.BuildTask, error) {
	var task model.BuildTask
	err := r.db.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *BuildTaskRepository) Update(task *model.BuildTask) error {
	return r.db.Save(task).Error
}

func (r *BuildTaskRepository) FindPending() ([]model.BuildTask, error) {
	var tasks []model.BuildTask
	err := r.db.Where("status = ?", model.BuildStatusPending).
		Order("created_at ASC").
		Find(&tasks).Error
	return tasks, err
}
