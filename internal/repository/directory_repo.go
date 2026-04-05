package repository

import (
	"release-manager/internal/model"

	"gorm.io/gorm"
)

type DirectoryRepository struct {
	db *gorm.DB
}

func NewDirectoryRepository(db *gorm.DB) *DirectoryRepository {
	return &DirectoryRepository{db: db}
}

func (r *DirectoryRepository) FindAll() ([]model.Directory, error) {
	var dirs []model.Directory
	err := r.db.Order("sort_order ASC, id ASC").Find(&dirs).Error
	return dirs, err
}

func (r *DirectoryRepository) FindByID(id uint) (*model.Directory, error) {
	var dir model.Directory
	err := r.db.First(&dir, id).Error
	if err != nil {
		return nil, err
	}
	return &dir, nil
}

func (r *DirectoryRepository) FindByParentID(parentID *uint) ([]model.Directory, error) {
	var dirs []model.Directory
	query := r.db.Order("sort_order ASC, id ASC")
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	err := query.Find(&dirs).Error
	return dirs, err
}

func (r *DirectoryRepository) Create(dir *model.Directory) error {
	return r.db.Create(dir).Error
}

func (r *DirectoryRepository) Update(dir *model.Directory) error {
	return r.db.Save(dir).Error
}

func (r *DirectoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Directory{}, id).Error
}

func (r *DirectoryRepository) HasChildren(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Directory{}).Where("parent_id = ?", id).Count(&count).Error
	return count > 0, err
}
