package dao

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type UploadDao struct {
	db *gorm.DB
}

func NewUploadDao(db *gorm.DB) *UploadDao {
	return &UploadDao{db: db}
}

func (d *UploadDao) CreateFolder(ctx context.Context, folder *Folder) (map[string]any, error) {
	now := time.Now().Unix()
	folder.Ctime = now
	folder.Utime = now

	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if folder.ParentId == 0 {
			folder.Path = "/" + folder.Name
		} else {
			var parent Folder
			err := tx.WithContext(ctx).Model(&Folder{}).Where("id = ? AND user_id = ?", folder.ParentId, folder.UserId).Find(&parent).Error
			if err != nil {
				return err
			}
			folder.Path = parent.Path + "/" + folder.Name
		}

		return tx.WithContext(ctx).Model(&Folder{}).Create(folder).Error
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"folderId": folder.ID,
		"path":     folder.Path,
	}, err
}

func (d *UploadDao) CreateFile(ctx context.Context, file *File) (map[string]any, error) {
	now := time.Now().Unix()
	file.Ctime = now
	file.Utime = now

	err := d.db.WithContext(ctx).Create(file).Error

	return map[string]any{
		"fileId": file.ID,
	}, err
}

// CreateFileWithTransaction 使用事务创建文件记录
func (d *UploadDao) CreateFileWithTransaction(ctx context.Context, file *File) (map[string]any, error) {
	now := time.Now().Unix()
	file.Ctime = now
	file.Utime = now

	var result map[string]any
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 再次检查文件是否已存在（防止并发）
		var existingFile File
		err := tx.WithContext(ctx).Where("uid = ? AND hash = ? AND folder_id = ?",
			file.UID, file.Hash, file.FolderID).First(&existingFile).Error

		if err == nil {
			// 文件已存在，返回现有文件信息
			result = map[string]any{
				"fileId":  existingFile.ID,
				"existed": true,
			}
			return nil
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		// 创建新文件记录
		err = tx.WithContext(ctx).Create(file).Error
		if err != nil {
			return err
		}

		result = map[string]any{
			"fileId":  file.ID,
			"existed": false,
		}
		return nil
	})

	return result, err
}

// CreateFileReference 创建文件引用（指向已存在的存储对象）
func (d *UploadDao) CreateFileReference(ctx context.Context, sourceFile *File, targetFolderID int64, newName string) (map[string]any, error) {
	newFile := &File{
		Name:           newName,
		Size:           sourceFile.Size,
		URL:            sourceFile.URL,
		Hash:           sourceFile.Hash,
		FolderID:       targetFolderID,
		UID:            sourceFile.UID,
		DeviceId:       sourceFile.DeviceId,
		LastModifiedBy: sourceFile.LastModifiedBy,
	}

	return d.CreateFile(ctx, newFile)
}

// BatchCreateFiles 批量创建文件记录
func (d *UploadDao) BatchCreateFiles(ctx context.Context, files []*File) error {
	if len(files) == 0 {
		return nil
	}

	now := time.Now().Unix()
	for _, file := range files {
		file.Ctime = now
		file.Utime = now
	}

	return d.db.WithContext(ctx).CreateInBatches(files, 100).Error
}

func (d *UploadDao) GetFolderByPath(ctx context.Context, userId int64, path string) (*Folder, error) {
	var folder Folder
	err := d.db.WithContext(ctx).Where("user_id = ? AND path = ? AND status = ?", userId, path, 1).First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (d *UploadDao) GetFolderById(ctx context.Context, userId int64, folderId int64) (*Folder, error) {
	var folder Folder
	err := d.db.WithContext(ctx).Model(&Folder{}).Where("id = ? AND user_id = ? AND status = ?", folderId, userId, 1).First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (d *UploadDao) ListFiles(ctx context.Context, userId int64, folderId int64) ([]*File, error) {
	var files []*File
	err := d.db.WithContext(ctx).Model(&File{}).Where("uid = ? AND folder_id = ?", userId, folderId).Find(&files).Error
	return files, err
}

func (d *UploadDao) ListFolders(ctx context.Context, userId int64, parentId int64) ([]*Folder, error) {
	var folders []*Folder
	err := d.db.WithContext(ctx).Where("user_id = ? AND parent_id = ? AND status = ?", userId, parentId, 1).Find(&folders).Error
	return folders, err
}

// GetFileByHash 根据哈希值查找文件
func (d *UploadDao) GetFileByHash(ctx context.Context, userId int64, hash string) (*File, error) {
	var file File
	err := d.db.WithContext(ctx).Where("uid = ? AND hash = ?", userId, hash).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// CheckFileExists 检查文件是否已存在
func (d *UploadDao) CheckFileExists(ctx context.Context, userId int64, hash string) (bool, *File, error) {
	var file File
	err := d.db.WithContext(ctx).Where("uid = ? AND hash = ?", userId, hash).First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &file, nil
}

// UpdateFileHash 更新文件哈希值（上传完成后）
func (d *UploadDao) UpdateFileHash(ctx context.Context, fileId int64, hash string) error {
	return d.db.WithContext(ctx).Model(&File{}).Where("id = ?", fileId).Update("hash", hash).Error
}

// GetFileById 根据ID获取文件
func (d *UploadDao) GetFileById(ctx context.Context, fileId int64, userId int64) (*File, error) {
	var file File
	err := d.db.WithContext(ctx).Where("id = ? AND uid = ?", fileId, userId).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetUserFileStats 获取用户文件统计信息
func (d *UploadDao) GetUserFileStats(ctx context.Context, userId int64) (*FileStats, error) {
	stats := &FileStats{}

	// 总文件数
	err := d.db.WithContext(ctx).Model(&File{}).Where("uid = ?", userId).Count(&stats.TotalFiles).Error
	if err != nil {
		return nil, err
	}

	// 总文件大小
	var totalSize sql.NullInt64
	err = d.db.WithContext(ctx).Model(&File{}).Where("uid = ?", userId).Select("SUM(size)").Scan(&totalSize).Error
	if err != nil {
		return nil, err
	}
	if totalSize.Valid {
		stats.TotalSize = totalSize.Int64
	}

	// 总文件夹数
	err = d.db.WithContext(ctx).Model(&Folder{}).Where("user_id = ? AND status = ?", userId, 1).Count(&stats.TotalFolders).Error
	if err != nil {
		return nil, err
	}

	// 重复文件数（基于哈希值）
	var duplicateQuery = `
		SELECT COUNT(*) - COUNT(DISTINCT hash) as duplicates 
		FROM file 
		WHERE uid = ? AND hash != '' AND hash IS NOT NULL
	`
	err = d.db.WithContext(ctx).Raw(duplicateQuery, userId).Scan(&stats.DuplicateFiles).Error
	if err != nil {
		return nil, err
	}

	// 节省的存储空间（重复文件的总大小）
	var savedQuery = `
		SELECT COALESCE(SUM(size), 0) as saved 
		FROM file f1 
		WHERE uid = ? AND hash != '' AND hash IS NOT NULL
		AND EXISTS (
			SELECT 1 FROM file f2 
			WHERE f2.uid = ? AND f2.hash = f1.hash AND f2.id < f1.id
		)
	`
	err = d.db.WithContext(ctx).Raw(savedQuery, userId, userId).Scan(&stats.StorageSaved).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// CountFilesByHash 统计相同哈希值的文件数量
func (d *UploadDao) CountFilesByHash(ctx context.Context, hash string) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&File{}).Where("hash = ?", hash).Count(&count).Error
	return count, err
}

// DeleteFile 删除文件
func (d *UploadDao) DeleteFile(ctx context.Context, fileId int64, userId int64) error {
	return d.db.WithContext(ctx).Where("id = ? AND uid = ?", fileId, userId).Delete(&File{}).Error
}

// DeleteFolder 删除文件夹
func (d *UploadDao) DeleteFolder(ctx context.Context, folderId int64, userId int64) error {
	return d.db.WithContext(ctx).Model(&Folder{}).Where("id = ? AND user_id = ?", folderId, userId).Update("status", 0).Error
}

// UpdateFile 更新文件信息
func (d *UploadDao) UpdateFile(ctx context.Context, fileId int64, userId int64, updates map[string]interface{}) error {
	return d.db.WithContext(ctx).Model(&File{}).Where("id = ? AND uid = ?", fileId, userId).Updates(updates).Error
}

// GetFileVersionsByHash 根据哈希值获取文件的所有版本
func (d *UploadDao) GetFileVersionsByHash(ctx context.Context, userId int64, hash string) ([]*File, error) {
	var files []*File
	err := d.db.WithContext(ctx).Where("uid = ? AND hash = ?", userId, hash).Order("ctime DESC").Find(&files).Error
	return files, err
}

// GetFilesByHashes 根据哈希值列表批量获取文件
func (d *UploadDao) GetFilesByHashes(ctx context.Context, userId int64, hashes []string) ([]*File, error) {
	var files []*File
	err := d.db.WithContext(ctx).Where("uid = ? AND hash IN ?", userId, hashes).Find(&files).Error
	return files, err
}

// CheckMultipleFilesExist 批量检查文件是否存在
func (d *UploadDao) CheckMultipleFilesExist(ctx context.Context, userId int64, hashes []string) (map[string]*File, error) {
	files, err := d.GetFilesByHashes(ctx, userId, hashes)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*File)
	for _, file := range files {
		result[file.Hash] = file
	}
	return result, nil
}

// GetFileStats 获取指定文件的统计信息
func (d *UploadDao) GetFileStats(ctx context.Context, fileId int64, userId int64) (*File, error) {
	var file File
	err := d.db.WithContext(ctx).Select("id, name, size, hash, folder_id, uid, ctime, utime").
		Where("id = ? AND uid = ?", fileId, userId).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetFileUploadHistory 获取用户的文件上传历史
func (d *UploadDao) GetFileUploadHistory(ctx context.Context, userId int64, limit int, offset int) ([]*File, int64, error) {
	var files []*File
	var total int64

	// 获取总数
	err := d.db.WithContext(ctx).Model(&File{}).Where("uid = ?", userId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = d.db.WithContext(ctx).Where("uid = ?", userId).
		Order("ctime DESC").Limit(limit).Offset(offset).Find(&files).Error
	if err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// GetLargeFiles 获取用户的大文件列表（大于指定大小）
func (d *UploadDao) GetLargeFiles(ctx context.Context, userId int64, minSize int64) ([]*File, error) {
	var files []*File
	err := d.db.WithContext(ctx).Where("uid = ? AND size >= ?", userId, minSize).
		Order("size DESC").Find(&files).Error
	return files, err
}

// GetRecentFiles 获取用户最近上传的文件
func (d *UploadDao) GetRecentFiles(ctx context.Context, userId int64, days int, limit int) ([]*File, error) {
	var files []*File
	cutoffTime := time.Now().AddDate(0, 0, -days).Unix()

	err := d.db.WithContext(ctx).Where("uid = ? AND ctime >= ?", userId, cutoffTime).
		Order("ctime DESC").Limit(limit).Find(&files).Error
	return files, err
}

// UpdateFileAccessTime 更新文件最后访问时间
func (d *UploadDao) UpdateFileAccessTime(ctx context.Context, fileId int64, userId int64) error {
	now := time.Now().Unix()
	return d.db.WithContext(ctx).Model(&File{}).
		Where("id = ? AND uid = ?", fileId, userId).
		Update("utime", now).Error
}

// GetDuplicateFiles 获取重复文件列表
func (d *UploadDao) GetDuplicateFiles(ctx context.Context, userId int64) ([]*File, error) {
	var files []*File
	subQuery := d.db.WithContext(ctx).Model(&File{}).
		Select("hash").Where("uid = ? AND hash != '' AND hash IS NOT NULL", userId).
		Group("hash").Having("COUNT(*) > 1")

	err := d.db.WithContext(ctx).Where("uid = ? AND hash IN (?)", userId, subQuery).
		Order("hash, ctime").Find(&files).Error
	return files, err
}
