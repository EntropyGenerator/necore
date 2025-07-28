package database

import (
	"context"

	"gorm.io/gorm"
)

type DB struct {
	Database *gorm.DB
	Context  *context.Context
}

// DB gorm connector
var instance *DB

func GetInstance() *DB {
	return instance
}
