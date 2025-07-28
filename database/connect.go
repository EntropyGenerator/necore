package database

import (
	"necore/database/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectSqlite() {
	var err error
	DB, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	DB.AutoMigrate(&model.User{})
}
