package model

type FileResp struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	URL   string `json:"url"`
	Hash  string `json:"hash"`
	Utime int64  `json:"utime"`
}

// PreUploadCheckResp 预上传检查响应
type PreUploadCheckResp struct {
	FileExists   bool   `json:"fileExists"`   // 文件是否已存在
	FileID       int64  `json:"fileId"`       // 如果存在，返回文件ID
	PresignedUrl string `json:"presignedUrl"` // 如果不存在，返回预签名URL
	ObjectKey    string `json:"objectKey"`    // 对象键
}

type FolderResp struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Utime    int64  `json:"utime"`
	Path     string `json:"path"`
	ParentID int64  `json:"parentID"`
}

// FileStatsResp 文件统计响应
type FileStatsResp struct {
	TotalFiles int64            `json:"totalFiles"`
	TotalSize  int64            `json:"totalSize"`
	FileTypes  map[string]int64 `json:"fileTypes"`
}

// FileVersionResp 文件版本响应
type FileVersionResp struct {
	ID        int64  `json:"id"`
	Version   int    `json:"version"`
	Hash      string `json:"hash"`
	Size      int64  `json:"size"`
	URL       string `json:"url"`
	DeviceId  string `json:"deviceId"`
	CreatedAt int64  `json:"createdAt"`
}

// InitChunkedUploadResp 分块上传初始化响应
type InitChunkedUploadResp struct {
	UploadId  string           `json:"uploadId"`
	ChunkUrls []ChunkUploadUrl `json:"chunkUrls"`
	ExpiresIn int64            `json:"expiresIn"`
}

// ChunkUploadUrl 分块上传URL
type ChunkUploadUrl struct {
	PartNumber   int    `json:"partNumber"`
	PresignedUrl string `json:"presignedUrl"`
}

// UploadChunkResp 分块上传响应
type UploadChunkResp struct {
	PartNumber int    `json:"partNumber"`
	ETag       string `json:"etag"`
}

// CompleteChunkedUploadResp 完成分块上传响应
type CompleteChunkedUploadResp struct {
	FileID  int64  `json:"fileId"`
	FileUrl string `json:"fileUrl"`
	Message string `json:"message"`
}
