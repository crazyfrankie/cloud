package dao

type File struct {
	ID             int64  `gorm:"primaryKey"`
	Name           string `gorm:"not null"`
	Size           int64  `gorm:"not null"`
	URL            string `gorm:"not null"`
	FolderID       int64  `gorm:"index:folder_uid"`
	UID            int64  `gorm:"index:folder_uid"`
	Version        int64  `gorm:"not null;default:1"`
	DeviceId       string `gorm:"type:varchar(64)"`
	LastModifiedBy string `gorm:"type:varchar(64)"`
	Ctime          int64  `gorm:"not null"`
	Utime          int64  `gorm:"not null"`
}

type Folder struct {
	ID       int64 `gorm:"primaryKey"`
	Name     string
	ParentId int64  `gorm:"index:uid_pid_status"`
	UserId   int64  `gorm:"index:uid_pid_status"`
	Path     string `gorm:"index"`
	Status   int    `gorm:"index:uid_pid_status"`
	Ctime    int64
	Utime    int64
}
