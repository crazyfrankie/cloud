package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"

	"github.com/crazyfrankie/cloud/internal/file/model"
	"github.com/crazyfrankie/cloud/internal/storage/service"
	"github.com/crazyfrankie/cloud/pkg/conf"
	"github.com/crazyfrankie/cloud/pkg/utils"
)

type PreviewService struct {
	fileService    *FileService
	storageService *service.StorageService
	config         *conf.Config
}

func NewPreviewService(fileService *FileService, storageService *service.StorageService) *PreviewService {
	return &PreviewService{
		fileService:    fileService,
		storageService: storageService,
		config:         conf.GetConf(),
	}
}

// GetFilePreview 获取文件预览信息
func (s *PreviewService) GetFilePreview(ctx context.Context, fileID int64, userID int64) (*model.PreviewFileResp, error) {
	// 获取文件信息
	fileInfo, err := s.fileService.GetFileById(ctx, fileID, userID)
	if err != nil {
		return nil, fmt.Errorf("get file info error: %w", err)
	}

	if fileInfo.IsDir {
		return nil, fmt.Errorf("cannot preview directory")
	}

	// 确定文件类型和预览策略
	ext := strings.ToLower(filepath.Ext(fileInfo.Name))
	previewResp := &model.PreviewFileResp{
		FileID:      fileInfo.ID,
		FileName:    fileInfo.Name,
		ContentType: utils.GetContentType(ext),
		Size:        fileInfo.Size,
	}

	// 设置预览大小限制 (100MB)
	maxPreviewSize := int64(100 * 1024 * 1024)
	previewResp.MaxPreviewSize = maxPreviewSize

	// 根据文件类型确定预览方式
	if s.isImageFile(ext) {
		previewResp.FileType = "image"
		previewResp.PreviewType = "direct"
		previewResp.IsEditable = false

		// 图片使用预签名URL直接预览
		objectKey := s.storageService.ExtractObjectKey(fileInfo.URL)
		previewURL, err := s.storageService.PresignDownload(ctx, objectKey, fileInfo.Name, time.Hour*24)
		if err != nil {
			return nil, fmt.Errorf("generate preview URL error: %w", err)
		}
		previewResp.PreviewURL = previewURL

	} else if s.isPDFFile(ext) {
		previewResp.FileType = "pdf"
		previewResp.PreviewType = "proxy" // 改为流式传输
		previewResp.IsEditable = false

		// PDF使用流式传输URL，而不是预签名URL
		previewResp.FileID = fileInfo.ID

	} else if s.isVideoFile(ext) {
		previewResp.FileType = "video"
		previewResp.PreviewType = "direct"
		previewResp.IsEditable = false

		// 视频文件大小检查
		if fileInfo.Size > maxPreviewSize {
			return nil, fmt.Errorf("video file too large for preview (max %d bytes)", maxPreviewSize)
		}

		// 视频使用预签名URL直接预览
		objectKey := s.storageService.ExtractObjectKey(fileInfo.URL)
		previewURL, err := s.storageService.PresignDownload(ctx, objectKey, fileInfo.Name, time.Hour*24)
		if err != nil {
			return nil, fmt.Errorf("generate preview URL error: %w", err)
		}
		previewResp.PreviewURL = previewURL

	} else if s.isAudioFile(ext) {
		previewResp.FileType = "audio"
		previewResp.PreviewType = "direct"
		previewResp.IsEditable = false

		// 音频使用预签名URL直接预览
		objectKey := s.storageService.ExtractObjectKey(fileInfo.URL)
		previewURL, err := s.storageService.PresignDownload(ctx, objectKey, fileInfo.Name, time.Hour*24)
		if err != nil {
			return nil, fmt.Errorf("generate preview URL error: %w", err)
		}
		previewResp.PreviewURL = previewURL

	} else {
		// 其他文件类型不支持预览
		previewResp.FileType = "unknown"
		previewResp.PreviewType = "none"
		previewResp.IsEditable = false
		return previewResp, nil // 不返回错误，只是标记为不支持
	}

	return previewResp, nil
}

// StreamFile 提供文件流式传输（主要用于PDF预览）
func (s *PreviewService) StreamFile(ctx context.Context, fileID int64, userID int64) (*minio.Object, *model.FileResp, error) {
	// 获取文件信息
	fileInfo, err := s.fileService.GetFileById(ctx, fileID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("get file info error: %w", err)
	}

	if fileInfo.IsDir {
		return nil, nil, fmt.Errorf("cannot stream directory")
	}

	// 从存储服务获取对象
	object, err := s.storageService.GetObject(ctx, fileInfo.URL, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("get object error: %w", err)
	}

	return object, fileInfo, nil
}

// 文件类型判断方法
func (s *PreviewService) isImageFile(ext string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

func (s *PreviewService) isPDFFile(ext string) bool {
	return ext == ".pdf"
}

func (s *PreviewService) isVideoFile(ext string) bool {
	videoExts := []string{".mp4", ".avi", ".mkv", ".mov", ".webm"}
	for _, videoExt := range videoExts {
		if ext == videoExt {
			return true
		}
	}
	return false
}

func (s *PreviewService) isAudioFile(ext string) bool {
	audioExts := []string{".mp3", ".wav", ".ogg", ".aac", ".flac", ".m4a"}
	for _, audioExt := range audioExts {
		if ext == audioExt {
			return true
		}
	}
	return false
}
