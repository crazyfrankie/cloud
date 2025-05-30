package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/crazyfrankie/cloud/internal/file/domain"

	"github.com/crazyfrankie/cloud/internal/file/dao"
)

type UploadService struct {
	dao *dao.UploadDao
}

func NewUploadService(dao *dao.UploadDao) *UploadService {
	return &UploadService{dao: dao}
}

// Upload 上传文件元数据
func (s *UploadService) Upload(ctx context.Context, f *dao.File) error {
	// 验证文件夹是否存在
	if f.FolderID != 0 {
		_, err := s.dao.GetFolderById(ctx, f.UID, f.FolderID)
		if err != nil {
			return fmt.Errorf("folder not found: %w", err)
		}
	}

	return s.dao.CreateFile(ctx, f)
}

// CreateFolder 创建文件夹
func (s *UploadService) CreateFolder(ctx context.Context, folder *dao.Folder) error {
	// 检查同级目录下是否已存在同名文件夹
	if folder.ParentId != 0 {
		parent, err := s.dao.GetFolderById(ctx, folder.UserId, folder.ParentId)
		if err != nil {
			return fmt.Errorf("parent folder not found: %w", err)
		}

		// 检查路径冲突
		expectedPath := parent.Path + "/" + folder.Name
		existing, err := s.dao.GetFolderByPath(ctx, folder.UserId, expectedPath)
		if err == nil && existing != nil {
			return errors.New("folder already exists")
		}
	} else {
		// 检查根目录下是否已存在同名文件夹
		expectedPath := "/" + folder.Name
		existing, err := s.dao.GetFolderByPath(ctx, folder.UserId, expectedPath)
		if err == nil && existing != nil {
			return errors.New("folder already exists")
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
	dFiles := make([]domain.File, 0, len(files))
	for _, file := range files {
		dFiles = append(dFiles, domain.File{
			ID:    file.ID,
			Name:  file.Name,
			Size:  file.Size,
			URL:   file.URL,
			Utime: file.Utime,
		})
	}

	folders, err := s.dao.ListFolders(ctx, userId, folderId)
	if err != nil {
		return nil, err
	}
	dFolders := make([]domain.Folder, 0, len(folders))
	for _, folder := range folders {
		dFolders = append(dFolders, domain.Folder{
			ID:    folder.ID,
			Name:  folder.Name,
			Utime: folder.Utime,
			Path:  folder.Path,
		})
	}

	return map[string]interface{}{
		"files":   dFiles,
		"folders": dFolders,
	}, nil
}
