package app

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

type fiberAppInstance struct {
	App *fiber.App
}

var instance *fiberAppInstance

func init() {
	app := fiber.New()

	instance = &fiberAppInstance{
		App: app,
	}
}

func GetInstance() *fiberAppInstance {
	return instance
}

func Start() {
	log.Fatal(instance.App.Listen(":3000"))
}
