package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/crazyfrankie/cloud/internal/file/dao"
	"github.com/crazyfrankie/cloud/internal/file/model"
	"github.com/crazyfrankie/cloud/internal/storage"
)

type FileService struct {
	dao     *dao.FileDao
	storage *storage.Service
}

func NewFileService(d *dao.FileDao, storage *storage.Service) *FileService {
	return &FileService{dao: d, storage: storage}
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

// FindByIDs 查找所有文件记录
func (s *FileService) FindByIDs(ctx context.Context, uid int64, fileIds []int64) ([]*dao.File, error) {
	return s.dao.FindByIDs(ctx, uid, fileIds)
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
		fileResp := &model.FileResp{
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

		contents = append(contents, fileResp)
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
			objectKey := s.storage.ExtractObjectKey(file.URL)
			if objectKey != "" {
				err = s.storage.DeleteObject(ctx, objectKey)
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
