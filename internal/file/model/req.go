package model

type CreateFileReq struct {
	Name     string `json:"name" binding:"required"`
	Size     int64  `json:"size" binding:"required"`
	Hash     string `json:"hash" binding:"required"`
	FolderID int64  `json:"folderId"`
	URL      string `json:"url" binding:"required"`
	DeviceId string `json:"deviceId"`
}

// PreUploadCheckReq 预上传检查请求
type PreUploadCheckReq struct {
	Name     string `json:"name" binding:"required"`
	Size     int64  `json:"size" binding:"required"`
	Hash     string `json:"hash" binding:"required"`
	FolderID int64  `json:"folderId"`
}

type CreateFolderReq struct {
	Name     string `json:"name" binding:"required"`
	ParentId int64  `json:"parentId"`
}

// BatchDeleteReq 批量删除请求
type BatchDeleteReq struct {
	FileIds []int64 `json:"fileIds" binding:"required"`
}

// UpdateFileReq 更新文件请求
type UpdateFileReq struct {
	Name     *string `json:"name,omitempty"`
	Hash     *string `json:"hash,omitempty"`
	Size     *int64  `json:"size,omitempty"`
	URL      *string `json:"url,omitempty"`
	DeviceId *string `json:"deviceId,omitempty"`
}

// InitChunkedUploadReq 分块上传初始化请求
type InitChunkedUploadReq struct {
	Name        string `json:"name" binding:"required"`
	Size        int64  `json:"size" binding:"required"`
	Hash        string `json:"hash" binding:"required"`
	FolderID    int64  `json:"folderId"`
	ChunkSize   int64  `json:"chunkSize" binding:"required"`
	TotalChunks int    `json:"totalChunks" binding:"required"`
}

// UploadChunkReq 分块上传请求
type UploadChunkReq struct {
	ChunkNumber int    `json:"chunkNumber" binding:"required"`
	ChunkHash   string `json:"chunkHash" binding:"required"`
	ChunkSize   int64  `json:"chunkSize" binding:"required"`
}

// CompleteChunkedUploadReq 完成分块上传请求
type CompleteChunkedUploadReq struct {
	ChunkETags []ChunkETag `json:"chunkETags" binding:"required"`
}

// ChunkETag 分块ETag信息
type ChunkETag struct {
	PartNumber int    `json:"partNumber" binding:"required"`
	ETag       string `json:"etag" binding:"required"`
}
