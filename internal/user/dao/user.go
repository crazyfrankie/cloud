package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type UserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) *UserDao {
	db.AutoMigrate(&User{})
	return &UserDao{db: db}
}

func (d *UserDao) Insert(ctx context.Context, u *User) error {
	now := time.Now().Unix()
	u.Ctime = now
	u.Utime = now

	return d.db.Model(&User{}).WithContext(ctx).Create(u).Error
}

func (d *UserDao) FindByName(ctx context.Context, name string) (User, error) {
	var res User
	err := d.db.Model(&User{}).WithContext(ctx).Where("nickname = ?", name).Select("id", "uuid", "password").Find(&res).Error

	return res, err
}

func (d *UserDao) FindByID(ctx context.Context, uid int64) (User, error) {
	var res User
	err := d.db.Model(&User{}).WithContext(ctx).Where("id = ?", uid).Find(&res).Error

	return res, err
}

func (d *UserDao) UpdateUser(ctx context.Context, id int64, updated map[string]any) (User, error) {
	var res User
	updated["utime"] = time.Now().Unix()
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&User{}).Where("id = ?", id).Updates(updated)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return tx.Model(&User{}).Where("id = ?", id).First(&res).Error
	})
	if err != nil {
		return User{}, err
	}

	return res, nil
}

func (d *UserDao) UpdateAvatar(ctx context.Context, uid int64, objectKey string) error {
	now := time.Now().Unix()

	return d.db.WithContext(ctx).Model(&User{}).Where("id = ?", uid).Updates(map[string]any{
		"avatar": objectKey,
		"utime":  now,
	}).Error
}
