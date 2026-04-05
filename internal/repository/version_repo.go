package repository

import (
	"release-manager/internal/model"

	"gorm.io/gorm"
)

type VersionRepository struct {
	db *gorm.DB
}

func NewVersionRepository(db *gorm.DB) *VersionRepository {
	return &VersionRepository{db: db}
}

func (r *VersionRepository) FindByDirectoryID(dirID uint, page, pageSize int) ([]model.Version, int64, error) {
	var versions []model.Version
	var total int64

	r.db.Model(&model.Version{}).Where("directory_id = ?", dirID).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.Where("directory_id = ?", dirID).
		Preload("Uploader").
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&versions).Error
	return versions, total, err
}

func (r *VersionRepository) FindByID(id uint) (*model.Version, error) {
	var version model.Version
	err := r.db.Preload("Uploader").First(&version, id).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *VersionRepository) Create(version *model.Version) error {
	return r.db.Create(version).Error
}

func (r *VersionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Version{}, id).Error
}

func (r *VersionRepository) SetBaseline(baselineDirID, versionID, userID uint) error {
	// 先删除旧的基线配置
	r.db.Where("baseline_dir_id = ?", baselineDirID).Delete(&model.BaselineConfig{})

	// 创建新的基线配置
	config := &model.BaselineConfig{
		BaselineDirID: baselineDirID,
		VersionID:     versionID,
		CreatedBy:     userID,
	}
	return r.db.Create(config).Error
}

func (r *VersionRepository) GetBaseline(baselineDirID uint) (*model.BaselineConfig, error) {
	var config model.BaselineConfig
	err := r.db.Where("baseline_dir_id = ?", baselineDirID).
		Preload("Version").
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *VersionRepository) GetBaselineVersions(baselineDirID uint) ([]model.Version, error) {
	var configs []model.BaselineConfig
	err := r.db.Where("baseline_dir_id = ?", baselineDirID).
		Preload("Version").
		Preload("Version.Uploader").
		Find(&configs).Error
	if err != nil {
		return nil, err
	}

	versions := make([]model.Version, len(configs))
	for i, c := range configs {
		versions[i] = c.Version
	}
	return versions, nil
}
