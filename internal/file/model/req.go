package model

// CreateFileReq 创建文件请求 - 统一的文件/文件夹创建
type CreateFileReq struct {
	Name     string `json:"name" binding:"required"` // 文件/文件夹名称
	Path     string `json:"path" binding:"required"` // 完整路径，如 "/folder1/file.txt" 或 "/folder1"
	Size     int64  `json:"size"`                    // 文件大小，文件夹为0
	Hash     string `json:"hash"`                    // 文件哈希值，文件夹为空
	URL      string `json:"url"`                     // 文件存储URL，文件夹为空
	IsDir    bool   `json:"isDir"`                   // 是否为文件夹
	DeviceId string `json:"deviceId"`                // 设备ID
}

// PreUploadCheckReq 预上传检查请求
type PreUploadCheckReq struct {
	Name       string `json:"name" binding:"required"`       // 文件名
	Size       int64  `json:"size" binding:"required"`       // 文件大小
	Hash       string `json:"hash" binding:"required"`       // 文件哈希
	ParentPath string `json:"parentPath" binding:"required"` // 父目录路径，如 "/" 或 "/folder1"
}

// ListContentsReq 列出目录内容请求
type ListContentsReq struct {
	Path string `json:"path" binding:"required"` // 目录路径，如 "/" 或 "/folder1"
}

// BatchDeleteReq 批量删除请求
type BatchDeleteReq struct {
	Paths []string `json:"paths" binding:"required"` // 文件/文件夹路径列表
}

// UpdateFileReq 更新文件请求
type UpdateFileReq struct {
	Name     *string `json:"name,omitempty"`     // 新名称
	NewPath  *string `json:"newPath,omitempty"`  // 新路径（重命名或移动）
	Hash     *string `json:"hash,omitempty"`     // 新哈希值
	Size     *int64  `json:"size,omitempty"`     // 新大小
	URL      *string `json:"url,omitempty"`      // 新URL
	DeviceId *string `json:"deviceId,omitempty"` // 新设备ID
}

// InitUploadReq 分块上传初始化请求
type InitUploadReq struct {
	Name               string `json:"name" binding:"required"`       // 文件名
	Size               int64  `json:"size" binding:"required"`       // 文件大小
	Hash               string `json:"hash" binding:"required"`       // 文件哈希
	ParentPath         string `json:"parentPath" binding:"required"` // 父目录路径
	PreferredChunkSize int64  `json:"preferredChunkSize"`            // 客户端期望的分块大小
	DeviceInfo         string `json:"deviceInfo"`                    // 设备信息，用于记录和优化
}

// UploadedChunk 已上传的分块信息
type UploadedChunk struct {
	PartNumber int    `json:"partNumber" binding:"required"` // 分块编号（从1开始）
	ETag       string `json:"etag"`                          // 分块的ETag，用于验证完整性
}

// CompleteUploadReq 完成分块上传请求
type CompleteUploadReq struct {
	UploadedChunks    []UploadedChunk `json:"uploadedChunks" binding:"required"` // 已上传的分块列表
	FileHash          string          `json:"fileHash"`                          // 完整文件的哈希值，用于验证
	FileName          string          `json:"fileName" binding:"required"`       // 文件名
	ParentPath        string          `json:"parentPath" binding:"required"`     // 父目录路径
	ClientFingerprint string          `json:"clientFingerprint"`                 // 客户端唯一标识，用于统计
}
