package main

import (
	"necore/app"
	"necore/router"
)

func main() {
	router.SetupRoutes()
	app.Start()
}
