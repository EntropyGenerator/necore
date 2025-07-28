package database

import (
	"necore/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectSqlite() {
	var err error
	instance.Database, err = gorm.Open(sqlite.Open("database.sqlite3"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	instance.Database.AutoMigrate(&model.User{})
}
