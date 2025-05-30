package dao

import (
	"database/sql"
)

type User struct {
	ID       int64        `gorm:"primaryKey"`
	Nickname string       `gorm:"type:varchar(64);index:nick_name"`
	Password []byte       `gorm:"type:varchar(128)"`
	Birthday sql.NullTime `gorm:"type:date"`
	Avatar   string
	Ctime    int64
	Utime    int64
}
