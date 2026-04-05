package service

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"release-manager/internal/config"

	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/google/uuid"
)

type BOSService struct {
	cfg    config.BOSConfig
	client *bos.Client
	logger *config.Logger
}

func NewBOSService(cfg config.BOSConfig, client *bos.Client, logger *config.Logger) *BOSService {
	return &BOSService{
		cfg:    cfg,
		client: client,
		logger: logger,
	}
}

type UploadResult struct {
	BOSPath     string
	InternalURL string
	ExternalURL string
}

func (s *BOSService) Upload(reader io.Reader, filename string, size int64, prefix string) (*UploadResult, error) {
	if s.client == nil {
		// 开发环境模拟
		s.logger.Warnw("BOS client not configured, using mock upload")
		mockPath := fmt.Sprintf("%s/%s/%s", prefix, time.Now().Format("2006/01/02"), filename)
		return &UploadResult{
			BOSPath:     mockPath,
			InternalURL: fmt.Sprintf("http://internal.example.com/%s", mockPath),
			ExternalURL: fmt.Sprintf("http://external.example.com/%s", mockPath),
		}, nil
	}

	// 生成唯一路径
	ext := filepath.Ext(filename)
	objectKey := fmt.Sprintf("%s/%s/%s%s", prefix, time.Now().Format("2006/01/02"), uuid.New().String(), ext)

	// 读取内容到字节数组
	data, err := io.ReadAll(reader)
	if err != nil {
		s.logger.Errorw("Failed to read file content", "error", err)
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 上传到BOS
	_, err = s.client.PutObjectFromBytes(s.cfg.Bucket, objectKey, data, nil)
	if err != nil {
		s.logger.Errorw("Failed to upload to BOS", "error", err, "key", objectKey)
		return nil, fmt.Errorf("上传失败: %w", err)
	}

	return &UploadResult{
		BOSPath:     objectKey,
		InternalURL: s.getInternalURL(objectKey),
		ExternalURL: s.getExternalURL(objectKey),
	}, nil
}

func (s *BOSService) Delete(objectKey string) error {
	if s.client == nil {
		s.logger.Warnw("BOS client not configured, skipping delete")
		return nil
	}

	err := s.client.DeleteObject(s.cfg.Bucket, objectKey)
	if err != nil {
		s.logger.Errorw("Failed to delete from BOS", "error", err, "key", objectKey)
		return fmt.Errorf("删除失败: %w", err)
	}
	return nil
}

func (s *BOSService) GetSignedURL(objectKey string, expireSeconds int, internal bool) (string, error) {
	if s.client == nil {
		// 开发环境模拟
		if internal {
			return fmt.Sprintf("http://internal.example.com/%s?token=mock", objectKey), nil
		}
		return fmt.Sprintf("http://external.example.com/%s?token=mock", objectKey), nil
	}

	url := s.client.BasicGeneratePresignedUrl(s.cfg.Bucket, objectKey, expireSeconds)

	// 替换域名
	if internal && s.cfg.InternalDomain != "" {
		url = s.replaceHost(url, s.cfg.InternalDomain)
	} else if !internal && s.cfg.ExternalDomain != "" {
		url = s.replaceHost(url, s.cfg.ExternalDomain)
	}

	return url, nil
}

func (s *BOSService) getInternalURL(objectKey string) string {
	if s.cfg.InternalDomain != "" {
		return fmt.Sprintf("https://%s/%s", s.cfg.InternalDomain, objectKey)
	}
	return fmt.Sprintf("https://%s.%s/%s", s.cfg.Bucket, s.cfg.Endpoint, objectKey)
}

func (s *BOSService) getExternalURL(objectKey string) string {
	if s.cfg.ExternalDomain != "" {
		return fmt.Sprintf("https://%s/%s", s.cfg.ExternalDomain, objectKey)
	}
	return fmt.Sprintf("https://%s.%s/%s", s.cfg.Bucket, s.cfg.Endpoint, objectKey)
}

func (s *BOSService) replaceHost(url, newHost string) string {
	// 简单替换host部分
	return fmt.Sprintf("https://%s%s", newHost, url[len("https://"):])
}

func (s *BOSService) GetObjectContent(objectKey string) ([]byte, error) {
	if s.client == nil {
		// 开发环境模拟
		s.logger.Warnw("BOS client not configured, returning mock content")
		return []byte(fmt.Sprintf("Mock content for: %s", objectKey)), nil
	}

	// 从BOS获取对象内容
	result, err := s.client.GetObject(s.cfg.Bucket, objectKey, nil)
	if err != nil {
		s.logger.Errorw("Failed to get object from BOS", "error", err, "key", objectKey)
		return nil, fmt.Errorf("获取文件内容失败: %w", err)
	}

	// 检查结果是否有Body字段
	if result.Body != nil {
		defer result.Body.Close()
		content, err := io.ReadAll(result.Body)
		if err != nil {
			s.logger.Errorw("Failed to read object content", "error", err, "key", objectKey)
			return nil, fmt.Errorf("读取文件内容失败: %w", err)
		}
		return content, nil
	}

	// 如果没有Body，尝试其他方式
	s.logger.Warnw("GetObject result has no Body field", "key", objectKey)
	return []byte{}, nil
}
