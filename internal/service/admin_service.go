package service

import (
	"fmt"
	"io"

	"release-manager/internal/config"
	"release-manager/internal/model"
	"release-manager/internal/repository"
)

type AdminService struct {
	repos      *repository.Repositories
	bosService *BOSService
	logger     *config.Logger
}

func NewAdminService(repos *repository.Repositories, bosService *BOSService, logger *config.Logger) *AdminService {
	return &AdminService{
		repos:      repos,
		bosService: bosService,
		logger:     logger,
	}
}

func (s *AdminService) GetUsers(page, pageSize int) ([]model.User, int64, error) {
	return s.repos.User.FindAll(page, pageSize)
}

func (s *AdminService) SetAdmin(userID uint, isAdmin bool, operatorID uint, ip string) error {
	if err := s.repos.User.SetAdmin(userID, isAdmin); err != nil {
		return err
	}

	action := "GRANT_ADMIN"
	detail := "授予管理员权限"
	if !isAdmin {
		action = "REVOKE_ADMIN"
		detail = "撤销管理员权限"
	}

	s.repos.OpLog.Create(&model.OperationLog{
		UserID:     operatorID,
		Action:     action,
		TargetType: "USER",
		TargetID:   userID,
		Detail:     detail,
		IPAddress:  ip,
	})

	s.logger.Infow("Admin status changed", "targetUserId", userID, "isAdmin", isAdmin, "operatorId", operatorID)

	return nil
}

type CreateDirectoryRequest struct {
	Name      string `json:"name" binding:"required"`
	ParentID  *uint  `json:"parentId"`
	Type      string `json:"type" binding:"required"`
	ListType  string `json:"listType" binding:"required"`
	SortOrder int    `json:"sortOrder"`
}

func (s *AdminService) CreateDirectory(req *CreateDirectoryRequest, operatorID uint, ip string) (*model.Directory, error) {
	dir := &model.Directory{
		Name:      req.Name,
		ParentID:  req.ParentID,
		Type:      model.DirectoryType(req.Type),
		ListType:  model.ListType(req.ListType),
		SortOrder: req.SortOrder,
	}

	if err := s.repos.Directory.Create(dir); err != nil {
		return nil, err
	}

	s.repos.OpLog.Create(&model.OperationLog{
		UserID:     operatorID,
		Action:     "CREATE_DIRECTORY",
		TargetType: "DIRECTORY",
		TargetID:   dir.ID,
		Detail:     fmt.Sprintf("创建目录: %s", req.Name),
		IPAddress:  ip,
	})

	s.logger.Infow("Directory created", "dirId", dir.ID, "name", req.Name, "operatorId", operatorID)

	return dir, nil
}

type UpdateDirectoryRequest struct {
	Name      string `json:"name" binding:"required"`
	Type      string `json:"type" binding:"required"`
	ListType  string `json:"listType" binding:"required"`
	SortOrder int    `json:"sortOrder"`
}

func (s *AdminService) UpdateDirectory(dirID uint, req *UpdateDirectoryRequest, operatorID uint, ip string) error {
	dir, err := s.repos.Directory.FindByID(dirID)
	if err != nil {
		return err
	}

	dir.Name = req.Name
	dir.Type = model.DirectoryType(req.Type)
	dir.ListType = model.ListType(req.ListType)
	dir.SortOrder = req.SortOrder

	if err := s.repos.Directory.Update(dir); err != nil {
		return err
	}

	s.repos.OpLog.Create(&model.OperationLog{
		UserID:     operatorID,
		Action:     "UPDATE_DIRECTORY",
		TargetType: "DIRECTORY",
		TargetID:   dirID,
		Detail:     fmt.Sprintf("更新目录: %s", req.Name),
		IPAddress:  ip,
	})

	return nil
}

func (s *AdminService) DeleteDirectory(dirID uint, operatorID uint, ip string) error {
	// 检查是否有子目录
	hasChildren, err := s.repos.Directory.HasChildren(dirID)
	if err != nil {
		return err
	}
	if hasChildren {
		return fmt.Errorf("目录下存在子目录，无法删除")
	}

	dir, err := s.repos.Directory.FindByID(dirID)
	if err != nil {
		return err
	}

	if err := s.repos.Directory.Delete(dirID); err != nil {
		return err
	}

	s.repos.OpLog.Create(&model.OperationLog{
		UserID:     operatorID,
		Action:     "DELETE_DIRECTORY",
		TargetType: "DIRECTORY",
		TargetID:   dirID,
		Detail:     fmt.Sprintf("删除目录: %s", dir.Name),
		IPAddress:  ip,
	})

	s.logger.Infow("Directory deleted", "dirId", dirID, "operatorId", operatorID)

	return nil
}

type UploadVersionRequest struct {
	DirectoryID uint
	VersionName string
	Description string
	FileName    string
	FileSize    int64
	Reader      io.Reader
}

func (s *AdminService) UploadVersion(req *UploadVersionRequest, operatorID uint, ip string) (*model.Version, error) {
	// 上传到BOS
	result, err := s.bosService.Upload(req.Reader, req.FileName, req.FileSize, fmt.Sprintf("versions/%d", req.DirectoryID))
	if err != nil {
		return nil, err
	}

	version := &model.Version{
		DirectoryID: req.DirectoryID,
		VersionName: req.VersionName,
		Description: req.Description,
		BOSPath:     result.BOSPath,
		InternalURL: result.InternalURL,
		ExternalURL: result.ExternalURL,
		FileSize:    req.FileSize,
		UploaderID:  operatorID,
	}

	if err := s.repos.Version.Create(version); err != nil {
		s.bosService.Delete(result.BOSPath)
		return nil, err
	}

	s.repos.OpLog.Create(&model.OperationLog{
		UserID:     operatorID,
		Action:     "UPLOAD_VERSION",
		TargetType: "VERSION",
		TargetID:   version.ID,
		Detail:     fmt.Sprintf("上传版本: %s", req.VersionName),
		IPAddress:  ip,
	})

	s.logger.Infow("Version uploaded", "versionId", version.ID, "name", req.VersionName, "operatorId", operatorID)

	return version, nil
}

func (s *AdminService) DeleteVersion(versionID uint, operatorID uint, ip string) error {
	version, err := s.repos.Version.FindByID(versionID)
	if err != nil {
		return err
	}

	// 删除BOS文件
	if err := s.bosService.Delete(version.BOSPath); err != nil {
		s.logger.Warnw("Failed to delete BOS file", "error", err, "path", version.BOSPath)
	}

	if err := s.repos.Version.Delete(versionID); err != nil {
		return err
	}

	s.repos.OpLog.Create(&model.OperationLog{
		UserID:     operatorID,
		Action:     "DELETE_VERSION",
		TargetType: "VERSION",
		TargetID:   versionID,
		Detail:     fmt.Sprintf("删除版本: %s", version.VersionName),
		IPAddress:  ip,
	})

	s.logger.Infow("Version deleted", "versionId", versionID, "operatorId", operatorID)

	return nil
}

type SetBaselineRequest struct {
	BaselineDirID uint `json:"baselineDirId" binding:"required"`
	VersionID     uint `json:"versionId" binding:"required"`
}

func (s *AdminService) SetBaseline(req *SetBaselineRequest, operatorID uint, ip string) error {
	// 验证基线目录
	dir, err := s.repos.Directory.FindByID(req.BaselineDirID)
	if err != nil {
		return err
	}
	if dir.ListType != model.ListTypeBaseline {
		return fmt.Errorf("目标目录不是基线目录")
	}

	// 验证版本存在
	version, err := s.repos.Version.FindByID(req.VersionID)
	if err != nil {
		return err
	}

	if err := s.repos.Version.SetBaseline(req.BaselineDirID, req.VersionID, operatorID); err != nil {
		return err
	}

	s.repos.OpLog.Create(&model.OperationLog{
		UserID:     operatorID,
		Action:     "SET_BASELINE",
		TargetType: "BASELINE_CONFIG",
		TargetID:   req.BaselineDirID,
		Detail:     fmt.Sprintf("设置基线版本: %s -> %s", dir.Name, version.VersionName),
		IPAddress:  ip,
	})

	s.logger.Infow("Baseline set", "baselineDirId", req.BaselineDirID, "versionId", req.VersionID, "operatorId", operatorID)

	return nil
}

func (s *AdminService) GetLogs(page, pageSize int, userID *uint, action string) ([]model.OperationLog, int64, error) {
	return s.repos.OpLog.FindAll(page, pageSize, userID, action)
}
