package main

import (
	"log"
	"necore/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	routes.LoadRoutes(app)
	log.Fatal(app.Listen(":3000"))
}
