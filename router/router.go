package router

import (
	"necore/app"
	"necore/handler"
	"necore/middleware"

	"github.com/gofiber/fiber/v2"
)

type routerInstance struct {
	Router *fiber.Router
}

var instance *routerInstance

func init() {
	app := app.GetInstance()
	api := app.App.Group("/necore")

	instance = &routerInstance{
		Router: &api,
	}
}

func GetInstance() *routerInstance {
	return instance
}

func SetupRoutes() {
	router := instance.Router
	(*router).Get("/slogan", handler.SloganHandler)

	authGroup := (*router).Group("/auth")
	authGroup.Post("/login", handler.Login)
	authGroup.Post("/register", middleware.AuthNeeded(), handler.AddUser)
}
