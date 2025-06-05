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

	// 新增字段：文件操作信息
	Action       string `json:"action,omitempty"`       // 操作类型：preview/download/text
	PreviewURL   string `json:"previewUrl,omitempty"`   // 预览URL
	DownloadURL  string `json:"downloadUrl,omitempty"`  // 下载URL
	Previewable  bool   `json:"previewable,omitempty"`  // 是否可预览
	HasThumbnail bool   `json:"hasThumbnail,omitempty"` // 是否有缩略图
	ContentType  string `json:"contentType,omitempty"`  // MIME类型
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
