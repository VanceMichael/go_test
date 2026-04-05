package repository

import (
	"release-manager/internal/model"

	"gorm.io/gorm"
)

type UserFileRepository struct {
	db *gorm.DB
}

func NewUserFileRepository(db *gorm.DB) *UserFileRepository {
	return &UserFileRepository{db: db}
}

func (r *UserFileRepository) Create(file *model.UserFile) error {
	return r.db.Create(file).Error
}

func (r *UserFileRepository) FindByUserID(userID uint, page, pageSize int) ([]model.UserFile, int64, error) {
	var files []model.UserFile
	var total int64

	r.db.Model(&model.UserFile{}).Where("user_id = ?", userID).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.Where("user_id = ?", userID).
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&files).Error
	return files, total, err
}

func (r *UserFileRepository) FindPublic(page, pageSize int) ([]model.UserFile, int64, error) {
	var files []model.UserFile
	var total int64

	r.db.Model(&model.UserFile{}).Where("is_public = ?", true).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.Where("is_public = ?", true).
		Preload("User").
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&files).Error
	return files, total, err
}

func (r *UserFileRepository) FindByID(id uint) (*model.UserFile, error) {
	var file model.UserFile
	err := r.db.First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *UserFileRepository) Delete(id uint) error {
	return r.db.Delete(&model.UserFile{}, id).Error
}
