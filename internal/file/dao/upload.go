package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type UploadDao struct {
	db *gorm.DB
}

func NewUploadDao(db *gorm.DB) *UploadDao {
	return &UploadDao{db: db}
}

func (d *UploadDao) CreateFolder(ctx context.Context, folder *Folder) error {
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

	return err
}

func (d *UploadDao) CreateFile(ctx context.Context, file *File) error {
	now := time.Now().Unix()
	file.Ctime = now
	file.Utime = now

	return d.db.WithContext(ctx).Create(file).Error
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
	err := d.db.WithContext(ctx).Where("id = ? AND user_id = ? AND status = ?", folderId, userId, 1).First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (d *UploadDao) ListFiles(ctx context.Context, userId int64, folderId int64) ([]*File, error) {
	var files []*File
	err := d.db.WithContext(ctx).Where("uid = ? AND folder_id = ?", userId, folderId).Find(&files).Error
	return files, err
}

func (d *UploadDao) ListFolders(ctx context.Context, userId int64, parentId int64) ([]*Folder, error) {
	var folders []*Folder
	err := d.db.WithContext(ctx).Where("user_id = ? AND parent_id = ? AND status = ?", userId, parentId, 1).Find(&folders).Error
	return folders, err
}
