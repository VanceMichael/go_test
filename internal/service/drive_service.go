package service

import (
	"fmt"
	"io"

	"release-manager/internal/config"
	"release-manager/internal/model"
	"release-manager/internal/repository"
)

type DriveService struct {
	fileRepo   *repository.UserFileRepository
	opLogRepo  *repository.OperationLogRepository
	bosService *BOSService
	logger     *config.Logger
}

func NewDriveService(
	fileRepo *repository.UserFileRepository,
	opLogRepo *repository.OperationLogRepository,
	bosService *BOSService,
	logger *config.Logger,
) *DriveService {
	return &DriveService{
		fileRepo:   fileRepo,
		opLogRepo:  opLogRepo,
		bosService: bosService,
		logger:     logger,
	}
}

func (s *DriveService) GetPersonalFiles(userID uint, page, pageSize int) ([]model.UserFile, int64, error) {
	return s.fileRepo.FindByUserID(userID, page, pageSize)
}

func (s *DriveService) GetPublicFiles(page, pageSize int) ([]model.UserFile, int64, error) {
	return s.fileRepo.FindPublic(page, pageSize)
}

type UploadFileRequest struct {
	FileName string
	FileSize int64
	IsPublic bool
	Reader   io.Reader
}

func (s *DriveService) Upload(userID uint, req *UploadFileRequest, ip string) (*model.UserFile, error) {
	// 确定上传路径前缀
	prefix := fmt.Sprintf("drive/personal/%d", userID)
	if req.IsPublic {
		prefix = "drive/public"
	}

	// 上传到BOS
	result, err := s.bosService.Upload(req.Reader, req.FileName, req.FileSize, prefix)
	if err != nil {
		return nil, err
	}

	// 保存文件记录
	file := &model.UserFile{
		UserID:   userID,
		FileName: req.FileName,
		BOSPath:  result.BOSPath,
		FileSize: req.FileSize,
		IsPublic: req.IsPublic,
	}

	if err := s.fileRepo.Create(file); err != nil {
		// 回滚BOS上传
		s.bosService.Delete(result.BOSPath)
		return nil, err
	}

	// 记录操作日志
	s.opLogRepo.Create(&model.OperationLog{
		UserID:     userID,
		Action:     "FILE_UPLOAD",
		TargetType: "USER_FILE",
		TargetID:   file.ID,
		Detail:     fmt.Sprintf("上传文件: %s", req.FileName),
		IPAddress:  ip,
	})

	s.logger.Infow("File uploaded", "fileId", file.ID, "userId", userID, "fileName", req.FileName)

	return file, nil
}

func (s *DriveService) Delete(userID uint, fileID uint, ip string) error {
	file, err := s.fileRepo.FindByID(fileID)
	if err != nil {
		return err
	}

	// 检查权限
	if file.UserID != userID {
		return fmt.Errorf("无权删除此文件")
	}

	// 删除BOS文件
	if err := s.bosService.Delete(file.BOSPath); err != nil {
		s.logger.Warnw("Failed to delete BOS file", "error", err, "path", file.BOSPath)
	}

	// 删除数据库记录
	if err := s.fileRepo.Delete(fileID); err != nil {
		return err
	}

	// 记录操作日志
	s.opLogRepo.Create(&model.OperationLog{
		UserID:     userID,
		Action:     "FILE_DELETE",
		TargetType: "USER_FILE",
		TargetID:   fileID,
		Detail:     fmt.Sprintf("删除文件: %s", file.FileName),
		IPAddress:  ip,
	})

	s.logger.Infow("File deleted", "fileId", fileID, "userId", userID)

	return nil
}

func (s *DriveService) GetFileURL(userID uint, fileID uint) (*DownloadURLResponse, error) {
	file, err := s.fileRepo.FindByID(fileID)
	if err != nil {
		return nil, err
	}

	// 检查权限：公开文件或自己的文件
	if !file.IsPublic && file.UserID != userID {
		return nil, fmt.Errorf("无权访问此文件")
	}

	internalURL, err := s.bosService.GetSignedURL(file.BOSPath, 3600, true)
	if err != nil {
		return nil, err
	}

	externalURL, err := s.bosService.GetSignedURL(file.BOSPath, 3600, false)
	if err != nil {
		return nil, err
	}

	return &DownloadURLResponse{
		InternalURL: internalURL,
		ExternalURL: externalURL,
	}, nil
}
