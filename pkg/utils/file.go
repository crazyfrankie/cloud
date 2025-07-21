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
