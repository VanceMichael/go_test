package service

import (
	"fmt"
	"path/filepath"
	"release-manager/internal/config"
	"release-manager/internal/model"
	"release-manager/internal/repository"
	"strings"
)

type VersionService struct {
	dirRepo     *repository.DirectoryRepository
	versionRepo *repository.VersionRepository
	opLogRepo   *repository.OperationLogRepository
	bosService  *BOSService
	logger      *config.Logger
}

func NewVersionService(
	dirRepo *repository.DirectoryRepository,
	versionRepo *repository.VersionRepository,
	opLogRepo *repository.OperationLogRepository,
	bosService *BOSService,
	logger *config.Logger,
) *VersionService {
	return &VersionService{
		dirRepo:     dirRepo,
		versionRepo: versionRepo,
		opLogRepo:   opLogRepo,
		bosService:  bosService,
		logger:      logger,
	}
}

type DirectoryTree struct {
	ID        uint             `json:"id"`
	Name      string           `json:"name"`
	Type      string           `json:"type"`
	ListType  string           `json:"listType"`
	SortOrder int              `json:"sortOrder"`
	Children  []*DirectoryTree `json:"children,omitempty"`
}

func (s *VersionService) GetDirectoryTree() ([]*DirectoryTree, error) {
	dirs, err := s.dirRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	dirMap := make(map[uint]*DirectoryTree)
	var roots []*DirectoryTree

	for _, d := range dirs {
		node := &DirectoryTree{
			ID:        d.ID,
			Name:      d.Name,
			Type:      string(d.Type),
			ListType:  string(d.ListType),
			SortOrder: d.SortOrder,
			Children:  []*DirectoryTree{},
		}
		dirMap[d.ID] = node
	}

	for _, d := range dirs {
		node := dirMap[d.ID]
		if d.ParentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := dirMap[*d.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return roots, nil
}

func (s *VersionService) GetVersions(dirID uint, page, pageSize int) ([]model.Version, int64, error) {
	dir, err := s.dirRepo.FindByID(dirID)
	if err != nil {
		return nil, 0, err
	}

	// 如果是基线目录，返回基线版本
	if dir.ListType == model.ListTypeBaseline {
		versions, err := s.versionRepo.GetBaselineVersions(dirID)
		return versions, int64(len(versions)), err
	}

	return s.versionRepo.FindByDirectoryID(dirID, page, pageSize)
}

type DownloadURLResponse struct {
	InternalURL string `json:"internalUrl"`
	ExternalURL string `json:"externalUrl"`
}

func (s *VersionService) GetDownloadURL(versionID uint) (*DownloadURLResponse, error) {
	version, err := s.versionRepo.FindByID(versionID)
	if err != nil {
		return nil, err
	}

	// 生成签名URL (有效期1小时)
	internalURL, err := s.bosService.GetSignedURL(version.BOSPath, 3600, true)
	if err != nil {
		return nil, err
	}

	externalURL, err := s.bosService.GetSignedURL(version.BOSPath, 3600, false)
	if err != nil {
		return nil, err
	}

	return &DownloadURLResponse{
		InternalURL: internalURL,
		ExternalURL: externalURL,
	}, nil
}

type VersionCompareResponse struct {
	Version1 *VersionInfo `json:"version1"`
	Version2 *VersionInfo `json:"version2"`
	Diff     *DiffResult  `json:"diff,omitempty"`
}

type VersionInfo struct {
	ID          uint   `json:"id"`
	VersionName string `json:"versionName"`
	Description string `json:"description"`
	FileSize    int64  `json:"fileSize"`
	Uploader    string `json:"uploader"`
	CreatedAt   string `json:"createdAt"`
}

type DiffResult struct {
	IsTextFile    bool   `json:"isTextFile"`
	Original      string `json:"original"`
	Modified      string `json:"modified"`
	OriginalSize  int64  `json:"originalSize"`
	ModifiedSize  int64  `json:"modifiedSize"`
	CanShowDiff   bool   `json:"canShowDiff"`
	Message       string `json:"message,omitempty"`
}

var textFileExtensions = map[string]bool{
	".txt":  true,
	".json": true,
	".xml":  true,
	".yaml": true,
	".yml":  true,
	".ini":  true,
	".conf": true,
	".cfg":  true,
	".properties": true,
	".env":  true,
	".md":   true,
	".html": true,
	".htm":  true,
	".css":  true,
	".js":   true,
	".ts":   true,
	".jsx":  true,
	".tsx":  true,
	".vue":  true,
	".py":   true,
	".go":   true,
	".java": true,
	".c":    true,
	".cpp":  true,
	".h":    true,
	".hpp":  true,
	".sh":   true,
	".bat":  true,
	".ps1":  true,
	".sql":  true,
	".log":  true,
	".csv":  true,
}

func isTextFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return textFileExtensions[ext]
}

func (s *VersionService) CompareVersions(versionID1, versionID2 uint) (*VersionCompareResponse, error) {
	v1, err := s.versionRepo.FindByID(versionID1)
	if err != nil {
		return nil, err
	}

	v2, err := s.versionRepo.FindByID(versionID2)
	if err != nil {
		return nil, err
	}

	if v1.DirectoryID != v2.DirectoryID {
		return nil, fmt.Errorf("两个版本必须属于同一目录")
	}

	resp := &VersionCompareResponse{
		Version1: &VersionInfo{
			ID:          v1.ID,
			VersionName: v1.VersionName,
			Description: v1.Description,
			FileSize:    v1.FileSize,
			Uploader:    getUploaderName(v1),
			CreatedAt:   v1.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		Version2: &VersionInfo{
			ID:          v2.ID,
			VersionName: v2.VersionName,
			Description: v2.Description,
			FileSize:    v2.FileSize,
			Uploader:    getUploaderName(v2),
			CreatedAt:   v2.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	isText1 := isTextFile(v1.BOSPath)
	isText2 := isTextFile(v2.BOSPath)

	diffResult := &DiffResult{
		IsTextFile:   isText1 && isText2,
		OriginalSize: v1.FileSize,
		ModifiedSize: v2.FileSize,
		CanShowDiff:  false,
	}

	if isText1 && isText2 {
		const maxSize = 10 * 1024 * 1024
		if v1.FileSize > maxSize || v2.FileSize > maxSize {
			diffResult.Message = "文件过大，无法显示内容差异（最大支持10MB）"
		} else {
			content1, err := s.bosService.GetObjectContent(v1.BOSPath)
			if err != nil {
				s.logger.Warnw("Failed to get content for version1", "error", err, "versionID", v1.ID)
				diffResult.Message = "获取版本1文件内容失败"
			} else {
				content2, err := s.bosService.GetObjectContent(v2.BOSPath)
				if err != nil {
					s.logger.Warnw("Failed to get content for version2", "error", err, "versionID", v2.ID)
					diffResult.Message = "获取版本2文件内容失败"
				} else {
					diffResult.Original = string(content1)
					diffResult.Modified = string(content2)
					diffResult.CanShowDiff = true
				}
			}
		}
	} else {
		diffResult.Message = "非文本文件，无法显示内容差异"
	}

	resp.Diff = diffResult
	return resp, nil
}

func getUploaderName(v *model.Version) string {
	if v.Uploader.ID != 0 {
		if v.Uploader.DisplayName != "" {
			return v.Uploader.DisplayName
		}
		return v.Uploader.Username
	}
	return "-"
}
