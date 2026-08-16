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

var serverDatabase *gorm.DB

var documentDatabase *gorm.DB

var botTokenDatabase *gorm.DB

var departmentDatabase *gorm.DB

var wikiDatabase *gorm.DB

func ConnectSqlite() {
	var err error
	userDatabase, err = gorm.Open(sqlite.Open("data/user.sqlite3"), &gorm.Config{})
	if err != nil {
		panic("failed to connect user database")
	}
	userDatabase.AutoMigrate(&model.User{})

	articleDatabase, err = gorm.Open(sqlite.Open("data/article.sqlite3"), &gorm.Config{})
	if err != nil {
		panic("failed to connect information database")
	}
	articleDatabase.AutoMigrate(&model.Article{})

	serverDatabase, err = gorm.Open(sqlite.Open("data/server.sqlite3"), &gorm.Config{})
	if err != nil {
		panic("failed to connect server database")
	}
	serverDatabase.AutoMigrate(&model.Server{})

	documentDatabase, err = gorm.Open(sqlite.Open("data/document.sqlite3"), &gorm.Config{})
	if err != nil {
		panic("failed to connect document database")
	}
	documentDatabase.AutoMigrate(&model.DocumentNode{})

	botTokenDatabase, err = gorm.Open(sqlite.Open("data/bot_connection.sqlite3"), &gorm.Config{})
	if err != nil {
		panic("failed to connect bot connection database")
	}
	botTokenDatabase.AutoMigrate(&model.BotToken{})

	wikiDatabase, err = gorm.Open(sqlite.Open("data/wiki.sqlite3"), &gorm.Config{})
	if err != nil {
		panic("failed to connect wiki database")
	}
	wikiDatabase.AutoMigrate(&model.Glossary{}, &model.Item{}, &model.WikiTag{})

	var count int64
	wikiDatabase.Model(&model.WikiTag{}).Count(&count)

	departmentDatabase, err = gorm.Open(sqlite.Open("data/department.sqlite3"), &gorm.Config{})
	if err != nil {
		panic("failed to connect department database")
	}
	departmentDatabase.AutoMigrate(&model.Department{}, &model.DepartmentMember{})
}

func GetUserDatabase() *gorm.DB {
	return userDatabase
}

func GetArticleDatabase() *gorm.DB {
	return articleDatabase
}

func GetServerDatabase() *gorm.DB {
	return serverDatabase
}

func GetDocumentDatabase() *gorm.DB {
	return documentDatabase
}

func GetBotTokenDatabase() *gorm.DB {
	return botTokenDatabase
}

func GetDepartmentDatabase() *gorm.DB {
	return departmentDatabase
}

func GetWikiDatabase() *gorm.DB {
	return wikiDatabase
}
