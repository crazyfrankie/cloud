package dao

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type FileDao struct {
	db *gorm.DB
}

func NewFileDao(db *gorm.DB) *FileDao {
	return &FileDao{db: db}
}

// CreateFile 创建文件/文件夹记录
func (d *FileDao) CreateFile(ctx context.Context, file *File) error {
	now := time.Now().Unix()
	file.Ctime = now
	file.Utime = now
	file.Status = 1

	// 验证路径格式
	if err := d.validatePath(file.Path); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// 确保父目录存在（如果不是根目录）
	if file.Path != "/" {
		parentPath := filepath.Dir(file.Path)
		if parentPath != "/" {
			exists, err := d.PathExists(ctx, file.UID, parentPath, true)
			if err != nil {
				return fmt.Errorf("check parent directory error: %w", err)
			}
			if !exists {
				return fmt.Errorf("parent directory does not exist: %s", parentPath)
			}
		}
	}

	err := d.db.WithContext(ctx).Create(file).Error
	if err != nil {
		return err
	}

	return nil
}

// GetFileByPath 根据路径获取文件/文件夹
func (d *FileDao) GetFileByPath(ctx context.Context, uid int64, path string) (*File, error) {
	var file File
	err := d.db.WithContext(ctx).Where("uid = ? AND path = ? AND status = ?", uid, path, 1).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetFileById 根据ID获取文件
func (d *FileDao) GetFileById(ctx context.Context, fileId int64, uid int64) (*File, error) {
	var file File
	err := d.db.WithContext(ctx).Where("id = ? AND uid = ? AND status = ?", fileId, uid, 1).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// FindByID 根据ID查找单个文件
func (d *FileDao) FindByID(ctx context.Context, uid int64, fileId int64) (*File, error) {
	var file File
	err := d.db.WithContext(ctx).Where("id = ? AND uid = ? AND status = ?", fileId, uid, 1).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// ListPathContents 列出指定路径下的内容
func (d *FileDao) ListPathContents(ctx context.Context, uid int64, path string) ([]*File, error) {
	var files []*File

	// 确保路径以 / 结尾（除了根目录）
	searchPath := path
	if path != "/" && !strings.HasSuffix(path, "/") {
		searchPath = path + "/"
	}

	// 查询直接子项：路径以 searchPath 开头，但不包含更深层级的子目录
	var query *gorm.DB
	if path == "/" {
		// 根目录：查询路径不包含 / 的项（除了开头的 /）
		query = d.db.WithContext(ctx).Where("uid = ? AND status = ? AND path LIKE ? AND path NOT LIKE ?",
			uid, 1, "/%", "/%/%")
	} else {
		// 子目录：查询路径以 searchPath 开头但不包含更深层级的项
		likePattern := searchPath + "%"
		notLikePattern := searchPath + "%/%"
		query = d.db.WithContext(ctx).Where("uid = ? AND status = ? AND path LIKE ? AND path NOT LIKE ?",
			uid, 1, likePattern, notLikePattern)
	}

	err := query.Order("is_dir DESC, name ASC").Find(&files).Error
	return files, err
}

// PathExists 检查路径是否存在
func (d *FileDao) PathExists(ctx context.Context, uid int64, path string, mustBeDir bool) (bool, error) {
	query := d.db.WithContext(ctx).Model(&File{}).Where("uid = ? AND path = ? AND status = ?", uid, path, 1)

	if mustBeDir {
		query = query.Where("is_dir = ?", true)
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// CheckFileExists 检查文件是否已存在（基于哈希值）
func (d *FileDao) CheckFileExists(ctx context.Context, uid int64, hash string) (bool, *File, error) {
	var file File
	err := d.db.WithContext(ctx).Where("uid = ? AND hash = ? AND status = ? AND is_dir = ?", uid, hash, 1, false).First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &file, nil
}

// DeleteByPath 根据路径删除文件/文件夹
func (d *FileDao) DeleteByPath(ctx context.Context, uid int64, path string) error {
	// 软删除：设置 status = 0
	return d.db.WithContext(ctx).Model(&File{}).
		Where("uid = ? AND path = ? AND status = ?", uid, path, 1).
		Updates(map[string]interface{}{
			"status": 0,
			"utime":  time.Now().Unix(),
		}).Error
}

// BatchDeleteByPaths 批量删除文件/文件夹
func (d *FileDao) BatchDeleteByPaths(ctx context.Context, uid int64, paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	// 软删除：设置 status = 0
	return d.db.WithContext(ctx).Model(&File{}).
		Where("uid = ? AND path IN ? AND status = ?", uid, paths, 1).
		Updates(map[string]interface{}{
			"status": 0,
			"utime":  time.Now().Unix(),
		}).Error
}

// UpdateFile 更新文件信息
func (d *FileDao) UpdateFile(ctx context.Context, fileId int64, uid int64, updates map[string]interface{}) error {
	updates["utime"] = time.Now().Unix()
	return d.db.WithContext(ctx).Model(&File{}).
		Where("id = ? AND uid = ? AND status = ?", fileId, uid, 1).
		Updates(updates).Error
}

func (d *FileDao) FindByIds(ctx context.Context, uid int64, fileIds []int64) ([]File, error) {
	var files []File
	if len(fileIds) == 0 {
		return files, nil
	}

	err := d.db.WithContext(ctx).Where("id IN ? AND uid = ? AND status = ?", fileIds, uid, 1).Find(&files).Error
	return files, err
}

// FindByIDs 根据ID列表查找文件（返回指针切片）
func (d *FileDao) FindByIDs(ctx context.Context, uid int64, fileIds []int64) ([]*File, error) {
	var files []*File
	if len(fileIds) == 0 {
		return files, nil
	}

	err := d.db.WithContext(ctx).Where("id IN ? AND uid = ? AND status = ?", fileIds, uid, 1).Find(&files).Error
	return files, err
}

// MovePath 移动文件/文件夹到新路径
func (d *FileDao) MovePath(ctx context.Context, uid int64, oldPath, newPath string) error {
	// 验证新路径格式
	if err := d.validatePath(newPath); err != nil {
		return fmt.Errorf("invalid new path: %w", err)
	}

	// 确保新路径的父目录存在
	if newPath != "/" {
		parentPath := filepath.Dir(newPath)
		if parentPath != "/" {
			exists, err := d.PathExists(ctx, uid, parentPath, true)
			if err != nil {
				return fmt.Errorf("check new parent directory error: %w", err)
			}
			if !exists {
				return fmt.Errorf("new parent directory does not exist: %s", parentPath)
			}
		}
	}

	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新主文件/文件夹路径
		err := tx.Model(&File{}).
			Where("uid = ? AND path = ? AND status = ?", uid, oldPath, 1).
			Updates(map[string]interface{}{
				"path":  newPath,
				"name":  filepath.Base(newPath),
				"utime": time.Now().Unix(),
			}).Error
		if err != nil {
			return err
		}

		// 如果是文件夹，还需要更新所有子项的路径
		var file File
		err = tx.Where("uid = ? AND path = ? AND status = ?", uid, newPath, 1).First(&file).Error
		if err != nil {
			return err
		}

		if file.IsDir {
			// 更新所有子项路径
			oldPrefix := oldPath
			if !strings.HasSuffix(oldPrefix, "/") {
				oldPrefix += "/"
			}

			newPrefix := newPath
			if !strings.HasSuffix(newPrefix, "/") {
				newPrefix += "/"
			}

			// 批量更新子项路径
			err = tx.Model(&File{}).
				Where("uid = ? AND path LIKE ? AND status = ?", uid, oldPrefix+"%", 1).
				Update("path", gorm.Expr("REPLACE(path, ?, ?)", oldPrefix, newPrefix)).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// GetUserFileStats 获取用户文件统计信息
func (d *FileDao) GetUserFileStats(ctx context.Context, uid int64) (*FileStats, error) {
	stats := &FileStats{}

	// 总文件数
	err := d.db.WithContext(ctx).Model(&File{}).
		Where("uid = ? AND status = ? AND is_dir = ?", uid, 1, false).
		Count(&stats.TotalFiles).Error
	if err != nil {
		return nil, err
	}

	// 总文件夹数
	err = d.db.WithContext(ctx).Model(&File{}).
		Where("uid = ? AND status = ? AND is_dir = ?", uid, 1, true).
		Count(&stats.TotalFolders).Error
	if err != nil {
		return nil, err
	}

	// 总文件大小
	var totalSize int64
	err = d.db.WithContext(ctx).Model(&File{}).
		Where("uid = ? AND status = ? AND is_dir = ?", uid, 1, false).
		Select("COALESCE(SUM(size), 0)").Scan(&totalSize).Error
	if err != nil {
		return nil, err
	}
	stats.TotalSize = totalSize

	// 重复文件统计
	var duplicateCount int64
	err = d.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) - COUNT(DISTINCT hash) as duplicates 
		FROM file 
		WHERE uid = ? AND status = ? AND is_dir = ? AND hash != '' AND hash IS NOT NULL
	`, uid, 1, false).Scan(&duplicateCount).Error
	if err != nil {
		return nil, err
	}
	stats.DuplicateFiles = duplicateCount

	return stats, nil
}

// GetFileVersionsByHash 根据哈希值获取文件的所有版本
func (d *FileDao) GetFileVersionsByHash(ctx context.Context, uid int64, hash string) ([]*File, error) {
	var files []*File
	err := d.db.WithContext(ctx).Where("uid = ? AND hash = ? AND status = ? AND is_dir = ?",
		uid, hash, 1, false).Order("ctime DESC").Find(&files).Error
	return files, err
}

// CheckFileReferences 检查文件是否还有其他引用
func (d *FileDao) CheckFileReferences(ctx context.Context, url string, hash string) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&File{}).
		Where("(url = ? OR hash = ?) AND status = ? AND is_dir = ?", url, hash, 1, false).
		Count(&count).Error
	return count > 1, err
}

// validatePath 验证路径格式
func (d *FileDao) validatePath(path string) error {
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
