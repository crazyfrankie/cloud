package model

import "github.com/minio/minio-go/v7"

// FileResp 统一的文件/文件夹响应
type FileResp struct {
	ID     int64  `json:"id"`     // 文件/文件夹ID
	Name   string `json:"name"`   // 名称
	Path   string `json:"path"`   // 完整路径
	IsDir  bool   `json:"isDir"`  // 是否为文件夹
	Size   int64  `json:"size"`   // 大小（文件夹为0）
	URL    string `json:"url"`    // 文件URL（文件夹为空）
	Hash   string `json:"hash"`   // 文件哈希（文件夹为空）
	Ctime  int64  `json:"ctime"`  // 创建时间
	Utime  int64  `json:"utime"`  // 更新时间
	Status int    `json:"status"` // 状态
}

// PreUploadCheckResp 预上传检查响应
type PreUploadCheckResp struct {
	FileExists   bool   `json:"fileExists"`   // 文件是否已存在
	FileID       int64  `json:"fileId"`       // 如果存在，返回文件ID
	FilePath     string `json:"filePath"`     // 如果存在，返回文件路径
	PresignedUrl string `json:"presignedUrl"` // 如果不存在，返回预签名URL
}

// ListContentsResp 目录内容响应
type ListContentsResp struct {
	Path     string      `json:"path"`     // 当前目录路径
	Contents []*FileResp `json:"contents"` // 目录内容（文件+文件夹）
	Total    int         `json:"total"`    // 总数量
}

// FileStatsResp 文件统计响应
type FileStatsResp struct {
	TotalFiles   int64 `json:"totalFiles"`   // 总文件数
	TotalFolders int64 `json:"totalFolders"` // 总文件夹数
	TotalSize    int64 `json:"totalSize"`    // 总大小
	TotalSpace   int64 `json:"totalSpace"`   // 总空间
	UsedSpace    int64 `json:"usedSpace"`    // 已用空间
}

// FileVersionResp 文件版本响应
type FileVersionResp struct {
	ID        int64  `json:"id"`
	Version   int    `json:"version"`
	Hash      string `json:"hash"`
	Size      int64  `json:"size"`
	URL       string `json:"url"`
	Path      string `json:"path"`
	DeviceId  string `json:"deviceId"`
	CreatedAt int64  `json:"createdAt"`
}

// CreateItemResp 创建文件/文件夹响应
type CreateItemResp struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Ctime int64  `json:"ctime"`
}

// ChunkUploadUrl 分块上传URL
type ChunkUploadUrl struct {
	PartNumber   int    `json:"partNumber"`
	PresignedUrl string `json:"presignedUrl"`
}

// InitUploadResp 分块上传初始化响应
type InitUploadResp struct {
	UploadId               string            `json:"uploadId"`
	ChunkUrls              []ChunkUploadUrl  `json:"chunkUrls"`
	ExpiresIn              int64             `json:"expiresIn"`
	RecommendedConcurrency int               `json:"recommendedConcurrency"`
	OptimalChunkSize       int64             `json:"optimalChunkSize"`
	TotalChunks            int               `json:"totalChunks"`
	UploadMethod           string            `json:"uploadMethod"`              // "direct-to-storage" 或 "fast"
	FileExists             bool              `json:"fileExists"`                // 文件是否已存在（秒传）
	FileID                 int64             `json:"fileId,omitempty"`          // 如果文件已存在，返回文件ID
	FileURL                string            `json:"fileUrl,omitempty"`         // 如果文件已存在，返回文件URL
	Message                string            `json:"message,omitempty"`         // 附加信息
	ServerSignature        string            `json:"serverSignature,omitempty"` // 服务器签名，用于验证
	ExistingParts          []*PartStatusResp `json:"existingParts,omitempty"`   // 已上传的分块信息（断点续传）
}

type PartStatusResp struct {
	ObjectKey string `json:"objectKey"`
	ETag      string `json:"etag"`
}

// DownloadFileInfo 下载文件信息
type DownloadFileInfo struct {
	Object   *minio.Object // MinIO object
	FileName string
	Size     int64
}

// DownloadFileResp 下载文件响应
type DownloadFileResp struct {
	Type      string `json:"type"`              // single/zip
	TotalSize int64  `json:"totalSize"`         // 总大小
	ZipName   string `json:"zipName,omitempty"` // ZIP文件名（多文件时）
	DLink     string `json:"dlink,omitempty"`   // 直接下载链接（单文件时）
	ZipData   []byte `json:"-"`                 // ZIP数据（不序列化到JSON）
}

// DownloadProgressInfo 下载进度信息
type DownloadProgressInfo struct {
	FileID       int64  `json:"fileId"`
	FileName     string `json:"fileName"`
	ContentType  string `json:"contentType"`
	TotalSize    int64  `json:"totalSize"`
	Range        string `json:"range,omitempty"`        // HTTP Range请求范围
	AcceptRanges bool   `json:"acceptRanges"`           // 是否支持断点续传
	LastModified string `json:"lastModified,omitempty"` // 最后修改时间
	ETag         string `json:"etag,omitempty"`         // 文件ETag
}

// PreviewFileResp 文件预览响应
type PreviewFileResp struct {
	FileID         int64  `json:"fileId"`
	FileName       string `json:"fileName"`
	FileType       string `json:"fileType"`       // image/text/pdf/video/audio
	PreviewType    string `json:"previewType"`    // direct/proxy/text/image
	PreviewURL     string `json:"previewUrl"`     // 预览URL
	ThumbnailURL   string `json:"thumbnailUrl"`   // 缩略图URL（如果有）
	ContentType    string `json:"contentType"`    // MIME类型
	Size           int64  `json:"size"`           // 文件大小
	IsEditable     bool   `json:"isEditable"`     // 是否可编辑
	TextContent    string `json:"textContent"`    // 文本内容（仅文本文件）
	Duration       int64  `json:"duration"`       // 视频/音频时长（秒）
	Width          int    `json:"width"`          // 图片/视频宽度
	Height         int    `json:"height"`         // 图片/视频高度
	MaxPreviewSize int64  `json:"maxPreviewSize"` // 最大预览大小限制
}

// UpdateContentResp 更新内容响应
type UpdateContentResp struct {
	Success      bool   `json:"success"`
	NewHash      string `json:"newHash"`      // 新的文件哈希
	NewSize      int64  `json:"newSize"`      // 新的文件大小
	UpdateTime   int64  `json:"updateTime"`   // 更新时间
	PresignedURL string `json:"presignedUrl"` // 用于上传更新内容的预签名URL
}
