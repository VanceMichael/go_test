package repository

import (
	"release-manager/internal/model"

	"gorm.io/gorm"
)

type Repositories struct {
	User      *UserRepository
	Directory *DirectoryRepository
	Version   *VersionRepository
	BuildTask *BuildTaskRepository
	UserFile  *UserFileRepository
	OpLog     *OperationLogRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:      NewUserRepository(db),
		Directory: NewDirectoryRepository(db),
		Version:   NewVersionRepository(db),
		BuildTask: NewBuildTaskRepository(db),
		UserFile:  NewUserFileRepository(db),
		OpLog:     NewOperationLogRepository(db),
	}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Directory{},
		&model.Version{},
		&model.BaselineConfig{},
		&model.BuildTask{},
		&model.UserFile{},
		&model.OperationLog{},
	)
}
