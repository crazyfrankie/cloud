package dao

// File 统一的文件/文件夹模型
type File struct {
	ID             int64  `gorm:"primaryKey"`
	Name           string `gorm:"not null"`                    // 文件/文件夹名称
	Path           string `gorm:"not null;index"`              // 完整路径，如 "/", "/folder1", "/folder1/file.txt"
	IsDir          bool   `gorm:"not null;default:false"`      // 是否为文件夹
	Size           int64  `gorm:"not null;default:0"`          // 文件大小，文件夹为0
	URL            string `gorm:"type:text"`                   // 文件存储URL，文件夹为空
	Hash           string `gorm:"type:varchar(128);index"`     // 文件哈希值，文件夹为空
	UID            int64  `gorm:"not null;index:idx_uid_path"` // 用户ID
	Version        int64  `gorm:"not null;default:1"`          // 文件版本
	DeviceId       string `gorm:"type:varchar(64)"`            // 设备ID
	LastModifiedBy string `gorm:"type:varchar(64)"`            // 最后修改者
	Status         int    `gorm:"not null;default:1"`          // 状态：1-正常，0-已删除
	Ctime          int64  `gorm:"not null"`                    // 创建时间
	Utime          int64  `gorm:"not null"`                    // 更新时间
}

// FileStats 文件统计信息结构
type FileStats struct {
	TotalFiles     int64 `json:"totalFiles"`
	TotalSize      int64 `json:"totalSize"`
	TotalFolders   int64 `json:"totalFolders"`
	DuplicateFiles int64 `json:"duplicateFiles"`
	StorageSaved   int64 `json:"storageSaved"`
}
