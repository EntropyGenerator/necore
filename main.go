package main

import (
	"necore/app"
	"necore/routes/slogan"
)

func main() {
	slogan.Load()
	app.Start()
}
