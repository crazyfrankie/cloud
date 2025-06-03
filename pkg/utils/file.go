package utils

// CalculateOptimalChunkSize 计算最优分块大小
func CalculateOptimalChunkSize(fileSize int64) int64 {
	if fileSize <= 0 {
		return 5 * 1024 * 1024 // 默认 5MB
	}

	// 根据文件大小确定合适的分块大小
	switch {
	case fileSize < 10*1024*1024: // 小于 10MB
		return 1 * 1024 * 1024 // 1MB
	case fileSize < 100*1024*1024: // 小于 100MB
		return 5 * 1024 * 1024 // 5MB
	case fileSize < 1024*1024*1024: // 小于 1GB
		return 10 * 1024 * 1024 // 10MB
	default: // 大于 1GB
		return 20 * 1024 * 1024 // 20MB
	}
}

// CalculateRecommendedConcurrency 计算推荐的并发数
func CalculateRecommendedConcurrency(fileSize int64) int {
	if fileSize <= 0 {
		return 3 // 默认值
	}

	// 根据文件大小确定合适的并发数
	switch {
	case fileSize < 10*1024*1024: // 小于 10MB
		return 2
	case fileSize < 100*1024*1024: // 小于 100MB
		return 4
	case fileSize < 1024*1024*1024: // 小于 1GB
		return 6
	default: // 大于 1GB
		return 8
	}
}

// FilePreviewConfig 文件预览配置
type FilePreviewConfig struct {
	// 支持在线预览的文件类型
	PreviewableTypes map[string]bool
	// 支持缩略图的文件类型
	ThumbnailTypes map[string]bool
	// 文本文件类型
	TextTypes map[string]bool
}

// GetFilePreviewConfig 获取文件预览配置
func GetFilePreviewConfig() *FilePreviewConfig {
	return &FilePreviewConfig{
		PreviewableTypes: map[string]bool{
			// 图片类型
			".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".webp": true, ".svg": true,
			// PDF
			".pdf": true,
			// 视频类型（部分浏览器支持）
			".mp4": true, ".webm": true,
			// 音频类型
			".mp3": true, ".wav": true, ".ogg": true, ".aac": true,
			// 文本类型
			".txt": true, ".md": true, ".json": true, ".xml": true, ".csv": true,
			".html": true, ".htm": true, ".css": true, ".js": true, ".ts": true,
			".go": true, ".py": true, ".java": true, ".cpp": true, ".c": true,
		},
		ThumbnailTypes: map[string]bool{
			// 图片类型支持缩略图
			".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".webp": true,
			// PDF支持缩略图
			".pdf": true,
			// 视频支持缩略图
			".mp4": true, ".avi": true, ".mkv": true, ".mov": true,
		},
		TextTypes: map[string]bool{
			".txt": true, ".md": true, ".json": true, ".xml": true, ".csv": true,
			".html": true, ".htm": true, ".css": true, ".js": true, ".ts": true,
			".go": true, ".py": true, ".java": true, ".cpp": true, ".c": true,
			".log": true, ".yaml": true, ".yml": true, ".ini": true, ".conf": true,
		},
	}
}

// FileAction 文件操作类型
type FileAction string

const (
	ActionPreview  FileAction = "preview"  // 在线预览
	ActionDownload FileAction = "download" // 直接下载
	ActionText     FileAction = "text"     // 文本预览
)

// FileActionInfo 文件操作信息
type FileActionInfo struct {
	Action       FileAction `json:"action"`       // 操作类型
	URL          string     `json:"url"`          // 对应的URL
	Previewable  bool       `json:"previewable"`  // 是否可预览
	Downloadable bool       `json:"downloadable"` // 是否可下载
	HasThumbnail bool       `json:"hasThumbnail"` // 是否有缩略图
	ContentType  string     `json:"contentType"`  // MIME类型
}

// GetContentType 根据文件扩展名获取MIME类型
func GetContentType(ext string) string {
	mimeTypes := map[string]string{
		// 图片
		".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png",
		".gif": "image/gif", ".bmp": "image/bmp", ".webp": "image/webp", ".svg": "image/svg+xml",
		// 文档
		".pdf": "application/pdf",
		".doc": "application/msword", ".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls": "application/vnd.ms-excel", ".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt": "application/vnd.ms-powerpoint", ".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		// 文本
		".txt": "text/plain", ".md": "text/markdown", ".json": "application/json",
		".xml": "application/xml", ".csv": "text/csv", ".html": "text/html", ".css": "text/css",
		// 代码
		".js": "application/javascript", ".ts": "application/typescript",
		".go": "text/plain", ".py": "text/plain", ".java": "text/plain",
		// 压缩文件
		".zip": "application/zip", ".rar": "application/x-rar-compressed", ".7z": "application/x-7z-compressed",
		// 视频
		".mp4": "video/mp4", ".avi": "video/x-msvideo", ".mkv": "video/x-matroska",
		// 音频
		".mp3": "audio/mpeg", ".wav": "audio/wav", ".ogg": "audio/ogg",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	return "application/octet-stream"
}
