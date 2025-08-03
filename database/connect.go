package database

import (
	"necore/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectSqlite() {
	db, err := gorm.Open(sqlite.Open("database.sqlite3"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&model.User{})

	instance = db
}
