package router

import (
	"necore/app"
	"necore/controller/middleware"
	"necore/service/handler"

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
	authGroup.Get("/status", middleware.AuthNeeded(), handler.GetStatus)
	authGroup.Post("/login", handler.Login)
	authGroup.Post("/register", middleware.AuthNeeded(), handler.AddUser)
	authGroup.Get("/user/:id", handler.GetUserInfo)
	authGroup.Get("/avatar/:id", handler.GetUserAvatar)
	authGroup.Get("/userlist", handler.GetUserList)
	authGroup.Delete("/user/:id", middleware.AuthNeeded(), handler.DeleteUser)
	authGroup.Post("/password", middleware.AuthNeeded(), handler.UpdateUserPassword)
	authGroup.Post("/avatar", middleware.AuthNeeded(), handler.UpdateUserAvatar)
	authGroup.Patch("/user", middleware.AuthNeeded(), handler.UpdateUserInfo)
	authGroup.Post("/logout", middleware.AuthNeeded(), handler.Logout)

	articleGroup := (*router).Group("/news")
	articleGroup.Get("/total/:target", handler.GetArticleCountByCategory)
	articleGroup.Post("/list", handler.GetArticleList)
	articleGroup.Get("/detail/:id", handler.GetArticleById)
	articleGroup.Patch("/:id", middleware.AuthNeeded(), handler.UpdateArticle)
	articleGroup.Post("/upload/:id", middleware.AuthNeeded(), handler.UploadArticleFile)
	articleGroup.Delete("/upload/:id", middleware.AuthNeeded(), handler.DeleteArticleFile)
	articleGroup.Post("/create", middleware.AuthNeeded(), handler.CreateArticle)
	articleGroup.Delete("/:id", middleware.AuthNeeded(), handler.DeleteArticle)
}
