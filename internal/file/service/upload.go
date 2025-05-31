package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/crazyfrankie/cloud/internal/file/dao"
	"github.com/crazyfrankie/cloud/internal/file/model"
	"github.com/crazyfrankie/cloud/internal/storage/service"
)

type UploadService struct {
	dao            *dao.UploadDao
	storageService *service.StorageService
}

func NewUploadService(dao *dao.UploadDao, storageService *service.StorageService) *UploadService {
	return &UploadService{dao: dao, storageService: storageService}
}

// PreUploadCheck 预上传检查，检查文件是否已存在
func (s *UploadService) PreUploadCheck(ctx context.Context, req model.PreUploadCheckReq, uid int64) (*model.PreUploadCheckResp, error) {
	// 验证文件夹是否存在
	if req.FolderID != 0 {
		_, err := s.dao.GetFolderById(ctx, uid, req.FolderID)
		if err != nil {
			return nil, fmt.Errorf("folder not found: %w", err)
		}
	}

	// 检查文件是否已存在（基于哈希值）
	exists, existingFile, err := s.dao.CheckFileExists(ctx, uid, req.Hash)
	if err != nil {
		return nil, fmt.Errorf("check file exists error: %w", err)
	}

	resp := &model.PreUploadCheckResp{
		FileExists: exists,
	}

	if exists {
		// 文件已存在，直接返回文件信息
		resp.FileID = existingFile.ID
	} else {
		// 文件不存在，生成预签名URL和对象键
		presignedUrl, objectKey, err := s.storageService.PresignWithPolicy(ctx, uid, req.Name, req.Size, req.Hash, "file")
		if err != nil {
			return nil, fmt.Errorf("generate presigned URL error: %w", err)
		}

		resp.PresignedUrl = presignedUrl
		resp.ObjectKey = objectKey
	}

	return resp, nil
}

// Upload 上传文件元数据
func (s *UploadService) Upload(ctx context.Context, req model.CreateFileReq, uid int64) (map[string]any, error) {
	file := &dao.File{
		Name:           req.Name,
		Size:           req.Size,
		URL:            req.URL,
		Hash:           req.Hash, // 添加哈希值
		FolderID:       req.FolderID,
		UID:            uid,
		DeviceId:       req.DeviceId,
		LastModifiedBy: strconv.FormatInt(uid, 10),
	}
	// 验证文件夹是否存在
	if file.FolderID != 0 {
		_, err := s.dao.GetFolderById(ctx, file.UID, file.FolderID)
		if err != nil {
			return nil, fmt.Errorf("folder not found: %w", err)
		}
	}

	return s.dao.CreateFile(ctx, file)
}

// CreateFolder 创建文件夹
func (s *UploadService) CreateFolder(ctx context.Context, req model.CreateFolderReq, uid int64) (map[string]any, error) {
	folder := &dao.Folder{
		Name:     req.Name,
		ParentId: req.ParentId,
		UserId:   uid,
	}
	// 检查同级目录下是否已存在同名文件夹
	if folder.ParentId != 0 {
		parent, err := s.dao.GetFolderById(ctx, folder.UserId, folder.ParentId)
		if err != nil {
			return nil, fmt.Errorf("parent folder not found: %w", err)
		}

		// 检查路径冲突
		expectedPath := parent.Path + "/" + folder.Name
		existing, err := s.dao.GetFolderByPath(ctx, folder.UserId, expectedPath)
		if err == nil && existing != nil {
			return nil, errors.New("folder already exists")
		}
	} else {
		// 检查根目录下是否已存在同名文件夹
		expectedPath := "/" + folder.Name
		existing, err := s.dao.GetFolderByPath(ctx, folder.UserId, expectedPath)
		if err == nil && existing != nil {
			return nil, errors.New("folder already exists")
		}
	}

	folder.Status = 1 // 正常状态
	return s.dao.CreateFolder(ctx, folder)
}

// ListFolderContents 列出文件夹内容
func (s *UploadService) ListFolderContents(ctx context.Context, userId int64, folderId int64) (map[string]interface{}, error) {
	files, err := s.dao.ListFiles(ctx, userId, folderId)
	if err != nil {
		return nil, err
	}
	dFiles := make([]model.FileResp, 0, len(files))
	for _, file := range files {
		dFiles = append(dFiles, model.FileResp{
			ID:    file.ID,
			Name:  file.Name,
			Size:  file.Size,
			URL:   file.URL,
			Hash:  file.Hash, // 添加Hash字段
			Utime: file.Utime,
		})
	}

	folders, err := s.dao.ListFolders(ctx, userId, folderId)
	if err != nil {
		return nil, err
	}
	dFolders := make([]model.FolderResp, 0, len(folders))
	for _, folder := range folders {
		dFolders = append(dFolders, model.FolderResp{
			ID:       folder.ID,
			Name:     folder.Name,
			Utime:    folder.Utime,
			Path:     folder.Path,
			ParentID: folder.ParentId,
		})
	}

	return map[string]interface{}{
		"files":   dFiles,
		"folders": dFolders,
	}, nil
}

// QuickUpload 快速上传（基于哈希的秒传功能）
func (s *UploadService) QuickUpload(ctx context.Context, req model.PreUploadCheckReq, uid int64) (*model.PreUploadCheckResp, error) {
	// 验证文件夹是否存在
	if req.FolderID != 0 {
		_, err := s.dao.GetFolderById(ctx, uid, req.FolderID)
		if err != nil {
			return nil, fmt.Errorf("folder not found: %w", err)
		}
	}

	// 检查是否有相同哈希的文件（支持跨用户的文件引用，但这里为了安全只查当前用户）
	exists, existingFile, err := s.dao.CheckFileExists(ctx, uid, req.Hash)
	// exists, existingFile, err := s.dao.CheckFileExists(ctx, req.Hash)
	if err != nil {
		return nil, fmt.Errorf("check file exists error: %w", err)
	}

	resp := &model.PreUploadCheckResp{
		FileExists: exists,
	}

	if exists {
		// 如果文件已存在且在不同文件夹，创建一个新的文件记录指向同一个存储对象
		if existingFile.FolderID != req.FolderID {
			newFile := &dao.File{
				Name:           req.Name,
				Size:           req.Size,
				URL:            existingFile.URL, // 使用已存在文件的URL
				Hash:           req.Hash,
				FolderID:       req.FolderID,
				UID:            uid,
				DeviceId:       "web-client",
				LastModifiedBy: strconv.FormatInt(uid, 10),
			}

			_, err := s.dao.CreateFile(ctx, newFile)
			if err != nil {
				return nil, fmt.Errorf("create file reference error: %w", err)
			}

			resp.FileID = newFile.ID
			resp.FileExists = true
			// 可以在响应中添加更多信息
		} else {
			resp.FileID = existingFile.ID
		}
	}

	return resp, nil
}

// ConfirmUpload 确认上传完成，保存文件元数据
func (s *UploadService) ConfirmUpload(ctx context.Context, req model.CreateFileReq, uid int64) (map[string]any, error) {
	// 验证文件夹是否存在
	if req.FolderID != 0 {
		_, err := s.dao.GetFolderById(ctx, uid, req.FolderID)
		if err != nil {
			return nil, fmt.Errorf("folder not found: %w", err)
		}
	}

	// 再次检查文件是否已存在（防止并发上传同一文件）
	exists, existingFile, err := s.dao.CheckFileExists(ctx, uid, req.Hash)
	if err != nil {
		return nil, fmt.Errorf("check file exists error: %w", err)
	}

	if exists {
		// 文件已存在，返回现有文件信息
		return map[string]any{
			"fileId":  existingFile.ID,
			"existed": true,
			"message": "文件已存在，已自动关联",
		}, nil
	}

	// 创建新文件记录
	file := &dao.File{
		Name:           req.Name,
		Size:           req.Size,
		URL:            req.URL,
		Hash:           req.Hash,
		FolderID:       req.FolderID,
		UID:            uid,
		DeviceId:       req.DeviceId,
		LastModifiedBy: strconv.FormatInt(uid, 10),
	}

	result, err := s.dao.CreateFile(ctx, file)
	if err != nil {
		return nil, err
	}

	// 添加成功标识
	result["existed"] = false
	result["message"] = "文件上传成功"

	return result, nil
}

// VerifyFileIntegrity 验证文件完整性
func (s *UploadService) VerifyFileIntegrity(ctx context.Context, fileId int64, uid int64) (bool, error) {
	// 获取文件信息
	file, err := s.dao.GetFileById(ctx, fileId, uid)
	if err != nil {
		return false, fmt.Errorf("file not found: %w", err)
	}

	// 这里可以添加更多的完整性验证逻辑
	// 比如验证文件大小、哈希值等
	if file.Hash == "" {
		return false, fmt.Errorf("file hash is empty")
	}

	return true, nil
}

// VerifyFile 验证文件完整性 (包装方法)
func (s *UploadService) VerifyFile(ctx context.Context, uid int64, fileId int64) (bool, error) {
	return s.VerifyFileIntegrity(ctx, fileId, uid)
}

// GetFileStatistics 获取文件统计信息
func (s *UploadService) GetFileStatistics(ctx context.Context, uid int64) (map[string]interface{}, error) {
	stats, err := s.dao.GetUserFileStats(ctx, uid)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"totalFiles":     stats.TotalFiles,
		"totalSize":      stats.TotalSize,
		"totalFolders":   stats.TotalFolders,
		"duplicateFiles": stats.DuplicateFiles,
		"storageSaved":   stats.StorageSaved,
	}, nil
}

// GetUserFileStats 获取用户文件统计 (包装方法)
func (s *UploadService) GetUserFileStats(ctx context.Context, uid int64) (*model.FileStatsResp, error) {
	stats, err := s.GetFileStatistics(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &model.FileStatsResp{
		TotalFiles: stats["totalFiles"].(int64),
		TotalSize:  stats["totalSize"].(int64),
		FileTypes:  make(map[string]int64), // TODO: 实现文件类型统计
	}, nil
}

// DeleteFile 删除文件
func (s *UploadService) DeleteFile(ctx context.Context, fileId int64, uid int64) error {
	// 获取文件信息
	file, err := s.dao.GetFileById(ctx, fileId, uid)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// 检查是否有其他文件引用相同的存储对象（基于哈希值）
	if file.Hash != "" {
		count, err := s.dao.CountFilesByHash(ctx, file.Hash)
		if err != nil {
			return fmt.Errorf("count files by hash error: %w", err)
		}

		// 如果只有一个文件引用这个存储对象，才删除存储对象
		if count == 1 {
			// TODO: 调用存储服务删除实际文件
			// s.storageService.DeleteObject(ctx, file.URL)
		}
	}

	// 删除数据库记录
	return s.dao.DeleteFile(ctx, fileId, uid)
}

// UpdateFile 更新文件信息
func (s *UploadService) UpdateFile(ctx context.Context, uid int64, fileId int64, req model.UpdateFileReq) (*model.FileResp, error) {
	// 获取现有文件信息验证权限
	_, err := s.dao.GetFileById(ctx, fileId, uid)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Hash != nil {
		updates["hash"] = *req.Hash
	}
	if req.Size != nil {
		updates["size"] = *req.Size
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.DeviceId != nil {
		updates["device_id"] = *req.DeviceId
	}

	// 执行更新
	err = s.dao.UpdateFile(ctx, fileId, uid, updates)
	if err != nil {
		return nil, fmt.Errorf("update file error: %w", err)
	}

	// 获取更新后的文件信息
	updatedFile, err := s.dao.GetFileById(ctx, fileId, uid)
	if err != nil {
		return nil, fmt.Errorf("get updated file error: %w", err)
	}

	return &model.FileResp{
		ID:    updatedFile.ID,
		Name:  updatedFile.Name,
		Size:  updatedFile.Size,
		URL:   updatedFile.URL,
		Hash:  updatedFile.Hash,
		Utime: updatedFile.Utime,
	}, nil
}

// GetFileVersions 获取文件版本信息
func (s *UploadService) GetFileVersions(ctx context.Context, uid int64, fileId int64) ([]*model.FileVersionResp, error) {
	// 获取文件信息
	file, err := s.dao.GetFileById(ctx, fileId, uid)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// 基于设备ID获取同一文件的不同版本
	versions, err := s.dao.GetFileVersionsByHash(ctx, uid, file.Hash)
	if err != nil {
		return nil, fmt.Errorf("get file versions error: %w", err)
	}

	result := make([]*model.FileVersionResp, 0, len(versions))
	for i, version := range versions {
		result = append(result, &model.FileVersionResp{
			ID:        version.ID,
			Version:   i + 1,
			Hash:      version.Hash,
			Size:      version.Size,
			URL:       version.URL,
			DeviceId:  version.DeviceId,
			CreatedAt: version.Ctime,
		})
	}

	return result, nil
}

// InitChunkedUpload 初始化分块上传
func (s *UploadService) InitChunkedUpload(ctx context.Context, uid int64, req model.InitChunkedUploadReq) (*model.InitChunkedUploadResp, error) {
	// 验证文件夹是否存在
	if req.FolderID != 0 {
		_, err := s.dao.GetFolderById(ctx, uid, req.FolderID)
		if err != nil {
			return nil, fmt.Errorf("folder not found: %w", err)
		}
	}

	// 检查文件是否已存在
	exists, existingFile, err := s.dao.CheckFileExists(ctx, uid, req.Hash)
	if err != nil {
		return nil, fmt.Errorf("check file exists error: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("file already exists with ID: %d", existingFile.ID)
	}

	// 初始化分块上传
	uploadId, chunkUrls, err := s.storageService.InitMultipartUpload(ctx, uid, req.Name, req.TotalChunks)
	if err != nil {
		return nil, fmt.Errorf("init multipart upload error: %w", err)
	}

	// 转换为响应格式
	chunkUrlsResp := make([]model.ChunkUploadUrl, 0, len(chunkUrls))
	for partNumber, url := range chunkUrls {
		chunkUrlsResp = append(chunkUrlsResp, model.ChunkUploadUrl{
			PartNumber:   partNumber,
			PresignedUrl: url,
		})
	}

	return &model.InitChunkedUploadResp{
		UploadId:  uploadId,
		ChunkUrls: chunkUrlsResp,
		ExpiresIn: 3600, // 1小时过期
	}, nil
}

// UploadChunk 上传单个分块
func (s *UploadService) UploadChunk(ctx context.Context, uid int64, uploadId string, req model.UploadChunkReq) (*model.UploadChunkResp, error) {
	// 这里主要是记录分块上传状态，实际上传由前端直接到存储服务
	// TODO: 可以在这里记录分块上传状态到数据库

	return &model.UploadChunkResp{
		PartNumber: req.ChunkNumber,
		ETag:       req.ChunkHash, // 使用分块哈希作为ETag
	}, nil
}

// CompleteChunkedUpload 完成分块上传
func (s *UploadService) CompleteChunkedUpload(ctx context.Context, uid int64, uploadId string, req model.CompleteChunkedUploadReq) (*model.CompleteChunkedUploadResp, error) {
	// 转换ChunkETag为interface{}切片
	chunkETags := make([]interface{}, len(req.ChunkETags))
	for i, tag := range req.ChunkETags {
		chunkETags[i] = tag
	}

	// 完成分块上传
	fileUrl, err := s.storageService.CompleteMultipartUpload(ctx, uploadId, chunkETags)
	if err != nil {
		return nil, fmt.Errorf("complete multipart upload error: %w", err)
	}

	// TODO: 创建文件记录到数据库
	// 需要从上传会话中获取文件信息

	return &model.CompleteChunkedUploadResp{
		FileID:  0, // TODO: 返回实际的文件ID
		FileUrl: fileUrl,
		Message: "Chunked upload completed successfully",
	}, nil
}

// AbortChunkedUpload 中止分块上传
func (s *UploadService) AbortChunkedUpload(ctx context.Context, uid int64, uploadId string) error {
	// 中止分块上传并清理
	return s.storageService.AbortMultipartUpload(ctx, uploadId)
}

// DeleteFolder 删除文件夹
func (s *UploadService) DeleteFolder(ctx context.Context, folderId int64, uid int64) error {
	// 检查文件夹是否存在
	_, err := s.dao.GetFolderById(ctx, uid, folderId)
	if err != nil {
		return fmt.Errorf("folder not found: %w", err)
	}

	// 检查文件夹是否为空
	files, err := s.dao.ListFiles(ctx, uid, folderId)
	if err != nil {
		return fmt.Errorf("list files error: %w", err)
	}
	if len(files) > 0 {
		return fmt.Errorf("folder is not empty, contains %d files", len(files))
	}

	subfolders, err := s.dao.ListFolders(ctx, uid, folderId)
	if err != nil {
		return fmt.Errorf("list subfolders error: %w", err)
	}
	if len(subfolders) > 0 {
		return fmt.Errorf("folder is not empty, contains %d subfolders", len(subfolders))
	}

	// 删除文件夹
	return s.dao.DeleteFolder(ctx, folderId, uid)
}

// BatchDelete 批量删除文件
func (s *UploadService) BatchDelete(ctx context.Context, fileIds []int64, uid int64) error {
	for _, fileId := range fileIds {
		if err := s.DeleteFile(ctx, fileId, uid); err != nil {
			return fmt.Errorf("delete file %d error: %w", fileId, err)
		}
	}
	return nil
}
