package migrate

import (
	"log"

	"gorm.io/gorm"

	"github.com/crazyfrankie/cloud/internal/file/dao"
	userDao "github.com/crazyfrankie/cloud/internal/user/dao"
)

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) error {
	log.Println("Starting database migration...")

	err := db.AutoMigrate(
		&userDao.User{},
		&dao.File{},
		&dao.Folder{},
	)

	if err != nil {
		log.Printf("Database migration failed: %v", err)
		return err
	}

	log.Println("Database migration completed successfully")
	return nil
}
