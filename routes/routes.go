package routes

import "github.com/gofiber/fiber/v2"

func LoadRoutes(app *fiber.App) {
	api := app.Group("/necore")

	api.Get("/slogan", sloganHandler)
}
