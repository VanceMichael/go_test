package service

import (
	"release-manager/internal/config"
	"release-manager/internal/repository"

	"github.com/baidubce/bce-sdk-go/services/bos"
)

type Services struct {
	Auth    *AuthService
	Version *VersionService
	Build   *BuildService
	Drive   *DriveService
	Admin   *AdminService
}

func NewServices(cfg *config.Config, repos *repository.Repositories, bosClient *bos.Client, logger *config.Logger) *Services {
	bosService := NewBOSService(cfg.BOS, bosClient, logger)

	return &Services{
		Auth:    NewAuthService(cfg, repos.User, logger),
		Version: NewVersionService(repos.Directory, repos.Version, repos.OpLog, bosService, logger),
		Build:   NewBuildService(cfg, repos.BuildTask, repos.OpLog, bosService, logger),
		Drive:   NewDriveService(repos.UserFile, repos.OpLog, bosService, logger),
		Admin:   NewAdminService(repos, bosService, logger),
	}
}
