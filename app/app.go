package app

import (
	"fmt"
	"log"
	"necore/config"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type fiberAppInstance struct {
	App *fiber.App
}

var instance *fiberAppInstance

func init() {
	app := fiber.New(fiber.Config{
		Prefork:   true,
		AppName:   "NMO Ecosystem Core",
		BodyLimit: 512 * 1024 * 1024,
	})

	instance = &fiberAppInstance{
		App: app,
	}
}

func GetInstance() *fiberAppInstance {
	return instance
}

func Start() {
	port, err := strconv.Atoi(config.Config("PORT"))
	if err != nil {
		port = 3000
	}
	log.Fatal(instance.App.Listen(fmt.Sprintf(":%d", port)))
}
