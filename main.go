package main

import (
	"necore/app"
	"necore/database"
	"necore/router"
)

func main() {
	database.ConnectSqlite()
	router.SetupRoutes()
	app.Start()
}
