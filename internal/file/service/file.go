package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/crazyfrankie/cloud/internal/file/dao"
	"github.com/crazyfrankie/cloud/internal/file/model"
	storagem "github.com/crazyfrankie/cloud/internal/storage/model"
	"github.com/crazyfrankie/cloud/internal/storage/service"
)

type FileService struct {
	dao            *dao.FileDao
	storageService *service.StorageService
}

func NewFileService(dao *dao.FileDao, storageService *service.StorageService) *FileService {
	return &FileService{dao: dao, storageService: storageService}
}

// PreUploadCheck 预上传检查，检查文件是否已存在
func (s *FileService) PreUploadCheck(ctx context.Context, req model.PreUploadCheckReq, uid int64) (*model.PreUploadCheckResp, error) {
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
		resp.FilePath = existingFile.Path
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

// CreateFile 创建文件记录
func (s *FileService) CreateFile(ctx context.Context, req model.CreateFileReq, uid int64) (*model.FileResp, error) {
	// 验证父目录是否存在（如果不是根目录）
	if req.Path != "/" {
		parentPath := filepath.Dir(req.Path)
		if parentPath != "/" {
			exists, err := s.dao.PathExists(ctx, uid, parentPath, true)
			if err != nil {
				return nil, fmt.Errorf("check parent directory error: %w", err)
			}
			if !exists {
				return nil, fmt.Errorf("parent directory not found: %s", parentPath)
			}
		}
	}

	// 检查路径是否已被占用
	exists, err := s.dao.PathExists(ctx, uid, req.Path, false)
	if err != nil {
		return nil, fmt.Errorf("check path exists error: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("path already exists: %s", req.Path)
	}

	// 如果是文件且有哈希值，检查是否已存在相同文件
	if !req.IsDir && req.Hash != "" {
		fileExists, _, err := s.dao.CheckFileExists(ctx, uid, req.Hash)
		if err != nil {
			return nil, fmt.Errorf("check file exists error: %w", err)
		}
		if fileExists {
			return nil, fmt.Errorf("file with same hash already exists")
		}
	}

	file := &dao.File{
		Name:           req.Name,
		Path:           req.Path,
		IsDir:          req.IsDir,
		Size:           req.Size,
		URL:            req.URL,
		Hash:           req.Hash,
		UID:            uid,
		DeviceId:       req.DeviceId,
		LastModifiedBy: strconv.FormatInt(uid, 10),
	}

	result, err := s.dao.CreateFile(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("create file error: %w", err)
	}

	// 从返回的 map 中获取文件 ID
	var fileId int64
	if fileIdVal, ok := result["fileId"]; ok {
		switch v := fileIdVal.(type) {
		case int64:
			fileId = v
		case int:
			fileId = int64(v)
		case float64:
			fileId = int64(v)
		default:
			return nil, fmt.Errorf("unexpected type for fileId: %T", fileIdVal)
		}
	} else {
		return nil, fmt.Errorf("fileId not found in result")
	}

	return &model.FileResp{
		ID:     fileId,
		Name:   req.Name,
		Path:   req.Path,
		IsDir:  req.IsDir,
		Size:   req.Size,
		URL:    req.URL,
		Hash:   req.Hash,
		Ctime:  file.Ctime,
		Utime:  file.Utime,
		Status: file.Status,
	}, nil
}

// ConfirmUpload 确认上传完成
func (s *FileService) ConfirmUpload(ctx context.Context, req model.CreateFileReq, uid int64) (*model.FileResp, error) {
	return s.CreateFile(ctx, req, uid)
}

// ListPathContents 列出指定路径下的内容
func (s *FileService) ListPathContents(ctx context.Context, uid int64, path string) (*model.ListContentsResp, error) {
	// 验证目录是否存在
	if path != "/" {
		exists, err := s.dao.PathExists(ctx, uid, path, true)
		if err != nil {
			return nil, fmt.Errorf("check directory error: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("directory not found: %s", path)
		}
	}

	// 获取目录内容
	files, err := s.dao.ListPathContents(ctx, uid, path)
	if err != nil {
		return nil, fmt.Errorf("list directory contents error: %w", err)
	}

	// 转换为响应格式
	contents := make([]*model.FileResp, 0, len(files))
	for _, file := range files {
		contents = append(contents, &model.FileResp{
			ID:     file.ID,
			Name:   file.Name,
			Path:   file.Path,
			IsDir:  file.IsDir,
			Size:   file.Size,
			URL:    file.URL,
			Hash:   file.Hash,
			Ctime:  file.Ctime,
			Utime:  file.Utime,
			Status: file.Status,
		})
	}

	return &model.ListContentsResp{
		Path:     path,
		Contents: contents,
		Total:    len(contents),
	}, nil
}

// MovePath 移动文件/文件夹到新路径
func (s *FileService) MovePath(ctx context.Context, uid int64, oldPath, newPath string) error {
	// 检查源路径是否存在
	sourceFile, err := s.dao.GetFileByPath(ctx, uid, oldPath)
	if err != nil {
		return fmt.Errorf("source path not found: %w", err)
	}

	// 检查目标路径是否已存在
	exists, err := s.dao.PathExists(ctx, uid, newPath, false)
	if err != nil {
		return fmt.Errorf("check target path error: %w", err)
	}
	if exists {
		return fmt.Errorf("target path already exists: %s", newPath)
	}

	// 如果是移动到自身或其子目录，则拒绝
	if sourceFile.IsDir && strings.HasPrefix(newPath, oldPath+"/") {
		return fmt.Errorf("cannot move a directory to its subdirectory")
	}

	return s.dao.MovePath(ctx, uid, oldPath, newPath)
}

// CopyPath 复制文件/文件夹到新路径
func (s *FileService) CopyPath(ctx context.Context, uid int64, sourcePath, targetPath string) error {
	// 检查源路径是否存在
	sourceFile, err := s.dao.GetFileByPath(ctx, uid, sourcePath)
	if err != nil {
		return fmt.Errorf("source path not found: %w", err)
	}

	// 检查目标路径是否已存在
	exists, err := s.dao.PathExists(ctx, uid, targetPath, false)
	if err != nil {
		return fmt.Errorf("check target path error: %w", err)
	}
	if exists {
		return fmt.Errorf("target path already exists: %s", targetPath)
	}

	// 验证新路径的父目录是否存在
	if targetPath != "/" {
		parentPath := filepath.Dir(targetPath)
		if parentPath != "/" {
			exists, err := s.dao.PathExists(ctx, uid, parentPath, true)
			if err != nil {
				return fmt.Errorf("check new parent directory error: %w", err)
			}
			if !exists {
				return fmt.Errorf("new parent directory does not exist: %s", parentPath)
			}
		}
	}

	// 如果是文件夹，不允许复制到自身或其子目录
	if sourceFile.IsDir && strings.HasPrefix(targetPath, sourcePath+"/") {
		return fmt.Errorf("cannot copy a directory to its subdirectory")
	}

	// 复制文件
	if !sourceFile.IsDir {
		newFile := &dao.File{
			Name:           filepath.Base(targetPath),
			Path:           targetPath,
			IsDir:          false,
			Size:           sourceFile.Size,
			URL:            sourceFile.URL,
			Hash:           sourceFile.Hash,
			UID:            uid,
			DeviceId:       sourceFile.DeviceId,
			LastModifiedBy: strconv.FormatInt(uid, 10),
		}
		_, err = s.dao.CreateFile(ctx, newFile)
		return err
	}

	// 复制文件夹及其内容
	return s.copyFolderRecursive(ctx, uid, sourcePath, targetPath)
}

// copyFolderRecursive 递归复制文件夹及其内容
func (s *FileService) copyFolderRecursive(ctx context.Context, uid int64, sourcePath, targetPath string) error {
	// 创建目标文件夹
	newFolder := &dao.File{
		Name:           filepath.Base(targetPath),
		Path:           targetPath,
		IsDir:          true,
		Size:           0,
		UID:            uid,
		LastModifiedBy: strconv.FormatInt(uid, 10),
	}

	_, err := s.dao.CreateFile(ctx, newFolder)
	if err != nil {
		return fmt.Errorf("create target folder error: %w", err)
	}

	// 获取源文件夹下的所有内容
	contents, err := s.dao.ListPathContents(ctx, uid, sourcePath)
	if err != nil {
		return fmt.Errorf("list source folder contents error: %w", err)
	}

	// 递归复制所有内容
	for _, item := range contents {
		sourcePath := item.Path
		itemName := item.Name
		newPath := filepath.Join(targetPath, itemName)

		if item.IsDir {
			// 递归复制子文件夹
			err = s.copyFolderRecursive(ctx, uid, sourcePath, newPath)
			if err != nil {
				return err
			}
		} else {
			// 复制文件
			newFile := &dao.File{
				Name:           itemName,
				Path:           newPath,
				IsDir:          false,
				Size:           item.Size,
				URL:            item.URL,
				Hash:           item.Hash,
				UID:            uid,
				DeviceId:       item.DeviceId,
				LastModifiedBy: strconv.FormatInt(uid, 10),
			}
			_, err = s.dao.CreateFile(ctx, newFile)
			if err != nil {
				return fmt.Errorf("copy file error: %w", err)
			}
		}
	}

	return nil
}

// OptimizedInitUpload 初始化优化的分块上传
func (s *FileService) OptimizedInitUpload(ctx context.Context, uid int64, req model.InitUploadReq) (*model.OptimizedInitUploadResp, error) {
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
		return &model.OptimizedInitUploadResp{
			FileExists: true,
			FileID:     existingFile.ID,
			FileURL:    existingFile.URL,
			Message:    "文件已存在，秒传成功",
		}, nil
	}

	// 文件不存在，需要上传
	// 计算最优分块大小和并发数
	chunkSize := calculateOptimalChunkSize(req.Size)
	concurrency := calculateRecommendedConcurrency(req.Size)
	totalChunks := int((req.Size + chunkSize - 1) / chunkSize)

	// 生成上传ID
	uploadId := fmt.Sprintf("%d_%s_%d", uid, req.Hash, req.Size)

	// 为每个分块生成预签名URL
	chunkUrls := make([]model.ChunkUploadUrl, totalChunks)
	for i := 0; i < totalChunks; i++ {
		// 使用统一的分块键格式：chunks/{uploadId}/{partNumber}
		// partNumber 从1开始，与前端保持一致
		partNumber := i + 1
		chunkKey := fmt.Sprintf("chunks/%s/%d", uploadId, partNumber)
		presignedUrl, err := s.storageService.PresignForChunk(ctx, chunkKey)
		if err != nil {
			return nil, fmt.Errorf("generate chunk presigned URL error: %w", err)
		}

		chunkUrls[i] = model.ChunkUploadUrl{
			PartNumber:   partNumber,
			PresignedUrl: presignedUrl,
		}
	}

	return &model.OptimizedInitUploadResp{
		FileExists:             false,
		UploadId:               uploadId,
		OptimalChunkSize:       chunkSize,
		TotalChunks:            totalChunks,
		RecommendedConcurrency: concurrency,
		ChunkUrls:              chunkUrls,
		UploadMethod:           "direct-to-storage",
		ExpiresIn:              3600, // 1 hour
	}, nil
}

// OptimizedCompleteUpload 完成优化的分块上传
func (s *FileService) OptimizedCompleteUpload(ctx context.Context, uid int64, uploadId string, req model.CompleteUploadReq) (*model.FileResp, error) {
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
		// chunk.PartNumber 是从前端传来的分块编号（从1开始）
		chunkKeys[i] = fmt.Sprintf("chunks/%s/%d", uploadId, chunk.PartNumber)
		chunkInfos[i] = storagem.ChunkInfo{
			Key:  chunkKeys[i],
			ETag: chunk.ETag,
		}
	}

	// 生成最终的对象键（保留文件扩展名）
	fileExt := filepath.Ext(req.FileName)
	var finalObjectKey string
	if req.FileHash != "" {
		finalObjectKey = fmt.Sprintf("files/%d/%s%s", uid, req.FileHash, fileExt)
	} else {
		// 如果没有哈希值，使用时间戳作为标识，但保留扩展名
		finalObjectKey = fmt.Sprintf("files/%d/%d%s", uid, time.Now().UnixNano(), fileExt)
	}

	// 合并所有分块为最终文件（使用ETag验证）
	err = s.storageService.ComposeObjectsWithETag(ctx, chunkInfos, finalObjectKey)
	if err != nil {
		return nil, fmt.Errorf("compose objects with etag validation error: %w", err)
	}

	// 生成最终文件URL
	fileUrl := fmt.Sprintf("%s/%s/%s", "http://localhost:9000", "cloud-files", finalObjectKey)

	// 创建文件记录
	file := &dao.File{
		Name:           req.FileName,
		Path:           filePath,
		IsDir:          false,
		Size:           0, // 需要从存储服务获取实际大小
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
				chunkSize := calculateOptimalChunkSize(0) // 使用默认分块大小
				file.Size = int64(len(req.UploadedChunks)) * chunkSize
			}
		} else {
			// 如果uploadId格式不正确，估算大小
			chunkSize := calculateOptimalChunkSize(0)
			file.Size = int64(len(req.UploadedChunks)) * chunkSize
		}
	}

	result, err := s.dao.CreateFile(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("create file record error: %w", err)
	}

	// 清理临时分块文件
	go func() {
		for _, chunkKey := range chunkKeys {
			_ = s.storageService.DeleteObject(context.Background(), chunkKey)
		}
	}()

	// 从返回的 map 中获取文件 ID
	var fileId int64
	if fileIdVal, ok := result["fileId"]; ok {
		switch v := fileIdVal.(type) {
		case int64:
			fileId = v
		case int:
			fileId = int64(v)
		case float64:
			fileId = int64(v)
		default:
			return nil, fmt.Errorf("unexpected type for fileId: %T", fileIdVal)
		}
	} else {
		return nil, fmt.Errorf("fileId not found in result")
	}

	return &model.FileResp{
		ID:     fileId,
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

// GetFileById 根据ID获取文件
func (s *FileService) GetFileById(ctx context.Context, fileId int64, uid int64) (*model.FileResp, error) {
	file, err := s.dao.GetFileById(ctx, fileId, uid)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	return &model.FileResp{
		ID:     file.ID,
		Name:   file.Name,
		Path:   file.Path,
		IsDir:  file.IsDir,
		Size:   file.Size,
		URL:    file.URL,
		Hash:   file.Hash,
		Ctime:  file.Ctime,
		Utime:  file.Utime,
		Status: file.Status,
	}, nil
}

// UpdateFile 更新文件信息
func (s *FileService) UpdateFile(ctx context.Context, fileId int64, uid int64, req model.UpdateFileReq) (*model.FileResp, error) {
	// 构建更新字段
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.NewPath != nil {
		// 验证新路径
		if err := s.validatePath(*req.NewPath); err != nil {
			return nil, fmt.Errorf("invalid new path: %w", err)
		}
		updates["path"] = *req.NewPath
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

	err := s.dao.UpdateFile(ctx, fileId, uid, updates)
	if err != nil {
		return nil, fmt.Errorf("update file error: %w", err)
	}

	return s.GetFileById(ctx, fileId, uid)
}

// GetUserFileStats 获取用户文件统计信息
func (s *FileService) GetUserFileStats(ctx context.Context, uid int64) (*model.FileStatsResp, error) {
	stats, err := s.dao.GetUserFileStats(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get user file stats error: %w", err)
	}

	return &model.FileStatsResp{
		TotalFiles:   stats.TotalFiles,
		TotalSize:    stats.TotalSize,
		TotalFolders: stats.TotalFolders,
		TotalSpace:   10 * 1024 * 1024 * 1024, // 默认 10GB
		UsedSpace:    stats.TotalSize,
	}, nil
}

// GetFileVersionsByHash 获取文件版本
func (s *FileService) GetFileVersionsByHash(ctx context.Context, uid int64, hash string) ([]*model.FileResp, error) {
	files, err := s.dao.GetFileVersionsByHash(ctx, uid, hash)
	if err != nil {
		return nil, fmt.Errorf("get file versions error: %w", err)
	}

	versions := make([]*model.FileResp, len(files))
	for i, file := range files {
		versions[i] = &model.FileResp{
			ID:     file.ID,
			Name:   file.Name,
			Path:   file.Path,
			IsDir:  file.IsDir,
			Size:   file.Size,
			URL:    file.URL,
			Hash:   file.Hash,
			Ctime:  file.Ctime,
			Utime:  file.Utime,
			Status: file.Status,
		}
	}

	return versions, nil
}

// VerifyFile 验证文件完整性
func (s *FileService) VerifyFile(ctx context.Context, uid int64, fileId int64) (bool, error) {
	file, err := s.dao.GetFileById(ctx, fileId, uid)
	if err != nil {
		return false, fmt.Errorf("file not found: %w", err)
	}

	if file.IsDir {
		return true, nil // 文件夹不需要验证
	}

	if file.URL == "" {
		return false, errors.New("file has no storage URL")
	}

	objectKey := s.storageService.ExtractObjectKey(file.URL)
	if objectKey == "" {
		return false, errors.New("invalid file URL")
	}

	objectInfo, err := s.storageService.GetObjectInfo(ctx, objectKey)
	if err != nil {
		return false, fmt.Errorf("storage object not found: %w", err)
	}

	return objectInfo.Size == file.Size, nil
}

// buildFilePath 构建完整的文件/文件夹路径
func (s *FileService) buildFilePath(parentPath, name string) string {
	if parentPath == "/" {
		return "/" + name
	}
	return parentPath + "/" + name
}

// validatePath 验证路径格式
func (s *FileService) validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must start with /")
	}
	if path != "/" && strings.HasSuffix(path, "/") {
		return fmt.Errorf("path cannot end with / (except root)")
	}
	if strings.Contains(path, "//") {
		return fmt.Errorf("path cannot contain //")
	}
	return nil
}

// 计算最优分块大小
func calculateOptimalChunkSize(fileSize int64) int64 {
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

// 计算推荐的并发数
func calculateRecommendedConcurrency(fileSize int64) int {
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

// DeleteByPath 根据路径删除文件/文件夹
func (s *FileService) DeleteByPath(ctx context.Context, uid int64, path string) error {
	// 检查路径是否存在
	file, err := s.dao.GetFileByPath(ctx, uid, path)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	// 如果是文件夹，则需要删除所有子文件和子文件夹
	if file.IsDir {
		// 获取文件夹下所有内容
		children, err := s.dao.ListPathContents(ctx, uid, path)
		if err != nil {
			return fmt.Errorf("list folder contents error: %w", err)
		}

		// 递归删除子文件和子文件夹
		for _, child := range children {
			err = s.DeleteByPath(ctx, uid, child.Path)
			if err != nil {
				return fmt.Errorf("delete child error: %w", err)
			}
		}
	} else {
		// 如果是文件，检查是否有其他引用
		hasReferences, err := s.dao.CheckFileReferences(ctx, file.URL, file.Hash)
		if err != nil {
			return fmt.Errorf("check file references error: %w", err)
		}

		// 如果没有其他引用，则从存储中删除
		if !hasReferences {
			// 从存储中删除文件
			objectKey := s.storageService.ExtractObjectKey(file.URL)
			if objectKey != "" {
				err = s.storageService.DeleteObject(ctx, objectKey)
				if err != nil {
					// 仅记录错误，继续执行软删除
					fmt.Printf("delete storage object error: %v\n", err)
				}
			}
		}
	}

	// 软删除文件/文件夹记录
	return s.dao.DeleteByPath(ctx, uid, path)
}

// BatchDeleteByPaths 批量删除多个路径
func (s *FileService) BatchDeleteByPaths(ctx context.Context, uid int64, paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	// 逐个删除路径
	for _, path := range paths {
		err := s.DeleteByPath(ctx, uid, path)
		if err != nil {
			return fmt.Errorf("delete path %s error: %w", path, err)
		}
	}

	return nil
}
