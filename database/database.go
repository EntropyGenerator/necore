package database

import (
	"necore/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB gorm connector
var userDatabase *gorm.DB

// "information" | "magazine" | "notice" | "activity" | "document"
var articleDatabase *gorm.DB

func ConnectSqlite() {
	var err error
	userDatabase, err = gorm.Open(sqlite.Open("data/user.sqlite3"), &gorm.Config{})
	if err != nil {
		panic("failed to connect user database")
	}
	articleDatabase, err = gorm.Open(sqlite.Open("data/article.sqlite3"), &gorm.Config{})
	if err != nil {
		panic("failed to connect information database")
	}
	// Migrate the schema
	userDatabase.AutoMigrate(&model.User{})
	articleDatabase.AutoMigrate(&model.Article{})
}

func GetUserDatabase() *gorm.DB {
	return userDatabase
}

func GetArticleDatabase() *gorm.DB {
	return articleDatabase
}
