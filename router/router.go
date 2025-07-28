package router

import (
	"necore/app"
	"necore/handler"

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
	(*instance.Router).Get("/slogan", handler.SloganHandler)

}
