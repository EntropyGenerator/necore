package main

import (
	"necore/app"
	"necore/controller/router"
	"necore/database"
)

func main() {
	database.ConnectSqlite()
	router.SetupRoutes()
	app.Start()
}
