package service

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// calculateHash 计算内容的MD5哈希值
func calculateHash(content []byte) string {
	hash := md5.Sum(content)
	return fmt.Sprintf("%x", hash)
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

	// 设置预览大小限制 (50MB)
	maxPreviewSize := int64(50 * 1024 * 1024)
	previewResp.MaxPreviewSize = maxPreviewSize

	// 统一使用KKFileView进行预览（除了文本文件需要特殊处理）
	if s.isTextFile(ext) {
		previewResp.FileType = "text"
		previewResp.PreviewType = "text"
		previewResp.IsEditable = true

		// 文本文件需要读取内容进行编辑
		if fileInfo.Size > maxPreviewSize {
			return nil, fmt.Errorf("file too large for preview (max %d bytes)", maxPreviewSize)
		}

		content, err := s.getTextFileContent(ctx, fileInfo.URL)
		if err != nil {
			return nil, fmt.Errorf("read text content error: %w", err)
		}
		previewResp.TextContent = content

		// 文本文件也可以通过KKFileView预览（只读模式）
		previewURL, err := s.generateKKFileViewURL(ctx, fileInfo.URL, fileInfo.Name)
		if err != nil {
			// 如果KKFileView不可用，文本文件仍可以显示内容
			return previewResp, nil
		}
		previewResp.PreviewURL = previewURL

	} else if s.isImageFile(ext) {
		previewResp.FileType = "image"
		previewResp.PreviewType = "kkfileview"
		previewResp.IsEditable = false

		// 图片使用KKFileView预览
		previewURL, err := s.generateKKFileViewURL(ctx, fileInfo.URL, fileInfo.Name)
		if err != nil {
			return nil, fmt.Errorf("generate KKFileView URL error: %w", err)
		}
		previewResp.PreviewURL = previewURL

	} else if s.isPDFFile(ext) {
		previewResp.FileType = "pdf"
		previewResp.PreviewType = "kkfileview"
		previewResp.IsEditable = false

		// PDF使用KKFileView预览
		previewURL, err := s.generateKKFileViewURL(ctx, fileInfo.URL, fileInfo.Name)
		if err != nil {
			return nil, fmt.Errorf("generate KKFileView URL error: %w", err)
		}
		previewResp.PreviewURL = previewURL

	} else if s.isVideoFile(ext) {
		previewResp.FileType = "video"
		previewResp.PreviewType = "kkfileview"
		previewResp.IsEditable = false

		// 视频文件大小检查
		if fileInfo.Size > maxPreviewSize {
			return nil, fmt.Errorf("video file too large for preview (max %d bytes)", maxPreviewSize)
		}

		// 视频使用KKFileView预览
		previewURL, err := s.generateKKFileViewURL(ctx, fileInfo.URL, fileInfo.Name)
		if err != nil {
			return nil, fmt.Errorf("generate KKFileView URL error: %w", err)
		}
		previewResp.PreviewURL = previewURL

	} else if s.isAudioFile(ext) {
		previewResp.FileType = "audio"
		previewResp.PreviewType = "kkfileview"
		previewResp.IsEditable = false

		// 音频使用KKFileView预览
		previewURL, err := s.generateKKFileViewURL(ctx, fileInfo.URL, fileInfo.Name)
		if err != nil {
			return nil, fmt.Errorf("generate KKFileView URL error: %w", err)
		}
		previewResp.PreviewURL = previewURL

	} else if s.isOfficeFile(ext) {
		previewResp.FileType = "office"
		previewResp.PreviewType = "kkfileview"
		previewResp.IsEditable = false

		// Office文件使用KKFileView预览
		previewURL, err := s.generateKKFileViewURL(ctx, fileInfo.URL, fileInfo.Name)
		if err != nil {
			return nil, fmt.Errorf("generate KKFileView URL error: %w", err)
		}
		previewResp.PreviewURL = previewURL

	} else {
		// 其他文件类型暂不支持预览
		previewResp.FileType = "unknown"
		previewResp.PreviewType = "none"
		previewResp.IsEditable = false
		return previewResp, fmt.Errorf("file type not supported for preview")
	}

	return previewResp, nil
}

// SaveTextFile 保存文本文件内容
func (s *PreviewService) SaveTextFile(ctx context.Context, fileID int64, userID int64, content string) error {
	// 获取文件信息
	fileInfo, err := s.fileService.GetFileById(ctx, fileID, userID)
	if err != nil {
		return fmt.Errorf("get file info error: %w", err)
	}

	// 验证是否为文本文件
	ext := strings.ToLower(filepath.Ext(fileInfo.Name))
	if !s.isTextFile(ext) {
		return fmt.Errorf("file is not editable")
	}

	// 计算新内容的哈希和大小
	newSize := int64(len(content))
	newHash := calculateHash([]byte(content))

	// 构建新的对象URL
	newURL := fmt.Sprintf("%d/%s", userID, newHash)

	// 更新文件记录 - 修正方法调用签名
	updateReq := model.UpdateFileReq{
		Hash: &newHash,
		Size: &newSize,
		URL:  &newURL,
	}

	err = s.fileService.UpdateFile(ctx, fileID, userID, updateReq)
	if err != nil {
		return fmt.Errorf("update file info error: %w", err)
	}

	return nil
}

// PrepareContentUpdate 准备文件内容更新
func (s *PreviewService) PrepareContentUpdate(ctx context.Context, fileID int64, userID int64, content string) (*model.UpdateContentResp, error) {
	// 获取文件信息
	fileInfo, err := s.fileService.GetFileById(ctx, fileID, userID)
	if err != nil {
		return nil, fmt.Errorf("get file info error: %w", err)
	}

	// 验证是否为可编辑的文本文件
	ext := strings.ToLower(filepath.Ext(fileInfo.Name))
	if !s.isTextFile(ext) {
		return nil, fmt.Errorf("file is not editable")
	}

	// 计算新内容的哈希和大小
	newSize := int64(len(content))
	newHash := calculateHash([]byte(content))

	// 生成预签名上传URL - 使用现有的Presign方法
	presignedURL, err := s.storageService.Presign(ctx, userID, fileInfo.Name, newSize, "file")
	if err != nil {
		return nil, fmt.Errorf("generate presigned URL error: %w", err)
	}

	return &model.UpdateContentResp{
		Success:      true,
		NewHash:      newHash,
		NewSize:      newSize,
		UpdateTime:   time.Now().Unix(),
		PresignedURL: presignedURL,
	}, nil
}

// ConfirmContentUpdate 确认文件内容更新
func (s *PreviewService) ConfirmContentUpdate(ctx context.Context, fileID int64, userID int64, newHash string, newSize int64) error {
	// 获取文件信息
	fileInfo, err := s.fileService.GetFileById(ctx, fileID, userID)
	if err != nil {
		return fmt.Errorf("get file info error: %w", err)
	}

	// 验证是否为可编辑的文本文件
	ext := strings.ToLower(filepath.Ext(fileInfo.Name))
	if !s.isTextFile(ext) {
		return fmt.Errorf("file is not editable")
	}

	// 构建新的对象URL
	newURL := fmt.Sprintf("users/%d/files/%s", userID, newHash)

	// 更新文件记录 - 修正方法调用签名
	updateReq := model.UpdateFileReq{
		Hash: &newHash,
		Size: &newSize,
		URL:  &newURL,
	}

	err = s.fileService.UpdateFile(ctx, fileID, userID, updateReq)
	if err != nil {
		return fmt.Errorf("update file info error: %w", err)
	}

	return nil
}

// generateKKFileViewURL 生成KKFileView预览URL
func (s *PreviewService) generateKKFileViewURL(ctx context.Context, objectURL, fileName string) (string, error) {
	objectKey := s.storageService.ExtractObjectKey(objectURL)
	downloadURL, err := s.storageService.PresignDownload(ctx, objectKey, fileName, time.Hour)
	if err != nil {
		return "", fmt.Errorf("generate download URL error: %w", err)
	}

	// 构建KKFileView预览URL
	// KKFileView API: /onlinePreview?url=文件下载地址
	encodedURL := url.QueryEscape(downloadURL)
	kkViewURL := fmt.Sprintf("%s/onlinePreview?url=%s", s.config.KKFileView.BaseURL, encodedURL)

	return kkViewURL, nil
}

// getTextFileContent 获取文本文件内容
func (s *PreviewService) getTextFileContent(ctx context.Context, objectURL string) (string, error) {
	// 从存储服务获取对象，添加必需的GetObjectOptions参数
	object, err := s.storageService.GetObject(ctx, objectURL, minio.GetObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("get object error: %w", err)
	}
	defer object.Close()

	// 读取文件内容
	content, err := io.ReadAll(object)
	if err != nil {
		return "", fmt.Errorf("read content error: %w", err)
	}

	return string(content), nil
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

func (s *PreviewService) isTextFile(ext string) bool {
	textExts := []string{".txt", ".md", ".json", ".xml", ".csv", ".html", ".htm", ".css", ".js", ".ts", ".go", ".py", ".java", ".cpp", ".c", ".log", ".yaml", ".yml", ".ini", ".conf"}
	for _, textExt := range textExts {
		if ext == textExt {
			return true
		}
	}
	return false
}

func (s *PreviewService) isPDFFile(ext string) bool {
	return ext == ".pdf"
}

func (s *PreviewService) isVideoFile(ext string) bool {
	videoExts := []string{".mp4", ".avi", ".mkv", ".mov", ".webm", ".flv", ".wmv"}
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

func (s *PreviewService) isOfficeFile(ext string) bool {
	officeExts := []string{".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".odt", ".ods", ".odp"}
	for _, officeExt := range officeExts {
		if ext == officeExt {
			return true
		}
	}
	return false
}

// CheckKKFileViewHealth 检查KKFileView服务健康状态
func (s *PreviewService) CheckKKFileViewHealth(ctx context.Context) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(s.config.KKFileView.BaseURL + "/index")
	if err != nil {
		return fmt.Errorf("KKFileView service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("KKFileView service returned status: %d", resp.StatusCode)
	}

	return nil
}
