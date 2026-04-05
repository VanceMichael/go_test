package handler

import (
	"release-manager/internal/config"
	"release-manager/internal/service"
)

type Handlers struct {
	Auth    *AuthHandler
	Version *VersionHandler
	Build   *BuildHandler
	Drive   *DriveHandler
	Admin   *AdminHandler
}

func NewHandlers(services *service.Services, logger *config.Logger) *Handlers {
	return &Handlers{
		Auth:    NewAuthHandler(services.Auth, logger),
		Version: NewVersionHandler(services.Version, logger),
		Build:   NewBuildHandler(services.Build, logger),
		Drive:   NewDriveHandler(services.Drive, logger),
		Admin:   NewAdminHandler(services.Admin, logger),
	}
}
