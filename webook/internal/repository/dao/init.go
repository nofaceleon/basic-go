package dao

import "gorm.io/gorm"

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{}) //数据库表的自动迁移
}
