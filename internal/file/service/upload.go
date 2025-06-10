package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/crazyfrankie/cloud/pkg/utils"

	"github.com/crazyfrankie/cloud/internal/file/dao"
	"github.com/crazyfrankie/cloud/internal/file/model"
	"github.com/crazyfrankie/cloud/internal/storage"
	storagem "github.com/crazyfrankie/cloud/internal/storage/model"
)

type UploadService struct {
	dao            *dao.FileDao
	file           *FileService
	storageService *storage.Service
}

func NewUploadService(dao *dao.FileDao, file *FileService, storageService *storage.Service) *UploadService {
	return &UploadService{dao: dao, file: file, storageService: storageService}
}

// PreUploadCheck 预上传检查，检查文件是否已存在
func (s *UploadService) PreUploadCheck(ctx context.Context, req model.PreUploadCheckReq, uid int64) (*model.PreUploadCheckResp, error) {
	if req.ParentPath != "/" {
		exists, err := s.dao.PathExists(ctx, uid, req.ParentPath, true)
		if err != nil {
			return nil, fmt.Errorf("check parent directory error: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("parent directory not found: %s", req.ParentPath)
		}
	}

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
		resp.FilePath = existingFile.Path
	} else {
		presignedUrl, err := s.storageService.Presign(ctx, uid, req.Name, req.Size, req.Hash, "file")
		if err != nil {
			return nil, fmt.Errorf("generate presigned URL error: %w", err)
		}

		resp.PresignedUrl = presignedUrl
	}

	return resp, nil
}

// ConfirmUpload 确认上传完成
func (s *UploadService) ConfirmUpload(ctx context.Context, req model.CreateFileReq, uid int64) error {
	return s.file.CreateFile(ctx, req, uid)
}

// InitUpload 初始化优化的分块上传
func (s *UploadService) InitUpload(ctx context.Context, uid int64, req model.InitUploadReq) (*model.InitUploadResp, error) {
	// 验证父目录是否存在
	if req.ParentPath != "/" {
		exists, err := s.dao.PathExists(ctx, uid, req.ParentPath, true)
		if err != nil {
			return nil, fmt.Errorf("check parent directory error: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("parent directory not found: %s", req.ParentPath)
		}
	}

	// 检查文件是否已存在（基于哈希值实现秒传）
	exists, existingFile, err := s.dao.CheckFileExists(ctx, uid, req.Hash)
	if err != nil {
		return nil, fmt.Errorf("check file exists error: %w", err)
	}

	if exists {
		// 文件已存在，直接返回现有文件信息（秒传）
		return &model.InitUploadResp{
			FileExists: true,
			FileID:     existingFile.ID,
			FileURL:    existingFile.URL,
			Message:    "文件已存在，秒传成功",
		}, nil
	}

	// 文件不存在，需要上传
	// 计算最优分块大小和并发数
	chunkSize := utils.CalculateOptimalChunkSize(req.Size)
	concurrency := utils.CalculateRecommendedConcurrency(req.Size)
	totalChunks := int((req.Size + chunkSize - 1) / chunkSize)

	// 生成上传ID
	uploadId := fmt.Sprintf("%d_%s_%d", uid, req.Hash, req.Size)

	// 为每个分块生成预签名URL
	chunkUrls := make([]model.ChunkUploadUrl, totalChunks)
	for i := 0; i < totalChunks; i++ {
		// 使用统一的分块键格式：{uid}/chunks/{uploadId}/{partNumber}
		partNumber := i + 1
		chunkKey := fmt.Sprintf("%d/chunks/%s/%d", uid, uploadId, partNumber)
		presignedUrl, err := s.storageService.PresignForChunk(ctx, chunkKey)
		if err != nil {
			return nil, fmt.Errorf("generate chunk presigned URL error: %w", err)
		}

		chunkUrls[i] = model.ChunkUploadUrl{
			PartNumber:   partNumber,
			PresignedUrl: presignedUrl,
		}
	}

	// 检查是否有已上传的分块（断点续传）
	existingParts := s.GetUploadStatus(ctx, uid, uploadId)

	return &model.InitUploadResp{
		FileExists:             false,
		UploadId:               uploadId,
		OptimalChunkSize:       chunkSize,
		TotalChunks:            totalChunks,
		RecommendedConcurrency: concurrency,
		ChunkUrls:              chunkUrls,
		UploadMethod:           "direct-to-storage",
		ExpiresIn:              3600,          // 1 hour
		ExistingParts:          existingParts, // 包含已上传的分块信息
	}, nil
}

// CompleteUpload 完成分块上传
func (s *UploadService) CompleteUpload(ctx context.Context, uid int64, uploadId string, req model.CompleteUploadReq) (*model.FileResp, error) {
	// 验证所有分块
	if len(req.UploadedChunks) == 0 {
		return nil, errors.New("no uploaded chunks provided")
	}

	// 构建完整的文件路径
	filePath := s.buildFilePath(req.ParentPath, req.FileName)

	// 检查文件路径是否已存在
	exists, err := s.dao.PathExists(ctx, uid, filePath, false)
	if err != nil {
		return nil, fmt.Errorf("check path exists error: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("file path already exists: %s", filePath)
	}

	// 验证父目录是否存在
	if req.ParentPath != "/" {
		exists, err := s.dao.PathExists(ctx, uid, req.ParentPath, true)
		if err != nil {
			return nil, fmt.Errorf("check parent directory error: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("parent directory not found: %s", req.ParentPath)
		}
	}

	// 构建分块键列表和ETag信息
	chunkKeys := make([]string, len(req.UploadedChunks))
	chunkInfos := make([]storagem.ChunkInfo, len(req.UploadedChunks))
	for i, chunk := range req.UploadedChunks {
		// chunk.PartNumber 从 1 开始
		chunkKeys[i] = fmt.Sprintf("%d/chunks/%s/%d", uid, uploadId, chunk.PartNumber)
		chunkInfos[i] = storagem.ChunkInfo{
			Key:  chunkKeys[i],
			ETag: chunk.ETag,
		}
	}

	finalObjectKey := fmt.Sprintf("%d/%s", uid, req.FileName)
	// 使用ETag验证, 合并所有分块为最终文件
	err = s.storageService.ComposeObjectsWithETag(ctx, chunkInfos, finalObjectKey)
	if err != nil {
		return nil, fmt.Errorf("compose objects with etag validation error: %w", err)
	}

	// 生成最终文件URL
	fileUrl := fmt.Sprintf("%s/%s/%s", "http://localhost:9000", "cloud-file", finalObjectKey)

	// 创建文件记录
	file := &dao.File{
		Name:           req.FileName,
		Path:           filePath,
		IsDir:          false,
		URL:            fileUrl,
		Hash:           req.FileHash,
		UID:            uid,
		DeviceId:       req.ClientFingerprint,
		LastModifiedBy: strconv.FormatInt(uid, 10),
	}

	// 获取文件大小
	objectInfo, err := s.storageService.GetObjectInfo(ctx, finalObjectKey)
	if err == nil && objectInfo.Size > 0 {
		file.Size = objectInfo.Size
	} else {
		// 如果无法获取大小，使用预期的文件大小（从uploadId中解析）
		// uploadId 格式："{uid}_{hash}_{size}"
		parts := strings.Split(uploadId, "_")
		if len(parts) >= 3 {
			if expectedSize, parseErr := strconv.ParseInt(parts[2], 10, 64); parseErr == nil {
				file.Size = expectedSize
			} else {
				// 如果解析失败，估算大小
				chunkSize := utils.CalculateOptimalChunkSize(0) // 使用默认分块大小
				file.Size = int64(len(req.UploadedChunks)) * chunkSize
			}
		} else {
			// 如果uploadId格式不正确，估算大小
			chunkSize := utils.CalculateOptimalChunkSize(0)
			file.Size = int64(len(req.UploadedChunks)) * chunkSize
		}
	}

	err = s.dao.CreateFile(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("create file record error: %w", err)
	}

	// 清理临时分块文件
	go func() {
		for _, chunkKey := range chunkKeys {
			_ = s.storageService.DeleteObject(context.Background(), chunkKey)
		}
	}()

	return &model.FileResp{
		ID:     file.ID,
		Name:   file.Name,
		Path:   file.Path,
		IsDir:  file.IsDir,
		Size:   file.Size,
		URL:    fileUrl,
		Hash:   req.FileHash,
		Ctime:  file.Ctime,
		Utime:  file.Utime,
		Status: file.Status,
	}, nil
}

// GetUploadStatus 获取已上传的分块状态
func (s *UploadService) GetUploadStatus(ctx context.Context, uid int64, uploadId string) []*model.PartStatusResp {
	objectInfo := s.storageService.GetUploadChunkObjects(ctx, uid, uploadId)

	res := make([]*model.PartStatusResp, 0, len(objectInfo))
	for _, o := range objectInfo {
		res = append(res, &model.PartStatusResp{
			ObjectKey: o.Key,
			ETag:      o.ETag,
		})
	}

	return res
}

// buildFilePath 构建完整的文件/文件夹路径
func (s *UploadService) buildFilePath(parentPath, name string) string {
	if parentPath == "/" {
		return "/" + name
	}
	return parentPath + "/" + name
}
