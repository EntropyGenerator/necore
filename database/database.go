package database

import "gorm.io/gorm"

// DB gorm connector
var DB *gorm.DB

func GetInstance() *gorm.DB {
	return DB
}
