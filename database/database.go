package database

import "github.com/gofiber/storage/sqlite3/v2"

type databaseInstance struct {
	db *sqlite3.Storage
}

var instance *databaseInstance

func init() {
	instance = &databaseInstance{
		db: sqlite3.New(sqlite3.Config{
			Database: "./database/database.db",
		})}
}

func GetInstance() *databaseInstance {
	return instance
}
