package database

import (
	"gorm.io/gorm"
)

// DB gorm connector
var instance *gorm.DB

func GetInstance() *gorm.DB {
	return instance
}
