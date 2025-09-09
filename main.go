package main

import (
	"necore/app"
	"necore/controller/router"
	"necore/database"
)

func main() {
	// This will print a hash of "test". U can insert it into sqlite3 manually for an admin account (the group section should be `["admin"]`).
	// dao.DebugTestPassword()

	database.ConnectSqlite()
	router.SetupRoutes()
	app.Start()
}
