package router

import (
	"necore/app"
	"necore/controller/middleware"
	"necore/service"

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
	(*router).Get("/slogan", service.SloganHandler)

	authGroup := (*router).Group("/auth")
	authGroup.Get("/status", middleware.AuthNeeded(), service.GetStatus)
	authGroup.Post("/login", service.Login)
	authGroup.Post("/register", middleware.AuthNeeded(), service.AddUser)
	authGroup.Get("/user/:id", service.GetUserInfo)
	authGroup.Get("/avatar/:id", service.GetUserAvatar)
	authGroup.Get("/userlist", service.GetUserList)
	authGroup.Delete("/user/:id", middleware.AuthNeeded(), service.DeleteUser)
	authGroup.Post("/password", middleware.AuthNeeded(), service.UpdateUserPassword)
	authGroup.Post("/avatar", middleware.AuthNeeded(), service.UpdateUserAvatar)
	authGroup.Patch("/user", middleware.AuthNeeded(), service.UpdateUserInfo)
	authGroup.Post("/logout", middleware.AuthNeeded(), service.Logout)

	articleGroup := (*router).Group("/news")
	articleGroup.Get("/total/:target", service.GetArticleCountByCategory)
	articleGroup.Post("/list", service.GetArticleList)
	articleGroup.Get("/detail/:id", service.GetArticleById)
	articleGroup.Patch("/:id", middleware.AuthNeeded(), service.UpdateArticle)
	articleGroup.Post("/upload/:id", middleware.AuthNeeded(), service.UploadArticleFile)
	articleGroup.Delete("/upload/:id", middleware.AuthNeeded(), service.DeleteArticleFile)
	articleGroup.Post("/create", middleware.AuthNeeded(), service.CreateArticle)
	articleGroup.Delete("/:id", middleware.AuthNeeded(), service.DeleteArticle)

	serverGroup := (*router).Group("/server")
	serverGroup.Get("/", service.GetServerList)
	serverGroup.Post("/status", service.GetServerStatus)
	serverGroup.Post("/", middleware.AuthNeeded(), service.AddServer)
	serverGroup.Delete("/:id", middleware.AuthNeeded(), service.DeleteServer)
	serverGroup.Patch("/", middleware.AuthNeeded(), service.AddServer)

	documentGroup := (*router).Group("/documents")
	documentGroup.Delete("/node/:id", middleware.AuthNeeded(), service.DeleteDocumentNode)
	documentGroup.Post("/node/:id", middleware.AuthNeeded(), service.UpdateDocumentNodeParentId)
	documentGroup.Put("/node/:id", middleware.AuthNeeded(), service.UpdateDocumentNodeContent)
	documentGroup.Patch("/node/:id", middleware.AuthNeeded(), service.UpdateDocumentNodeName)
	documentGroup.Post("/node", middleware.AuthNeeded(), service.CreateDocumentNode)
	documentGroup.Get("/layer/private/:parentId", middleware.AuthNeeded(), service.GetDocumentNodeChildrenPrivate)
	documentGroup.Get("/layer/:parentId", service.GetDocumentNodeChildren)
	documentGroup.Get("/private/:id", middleware.AuthNeeded(), service.GetDocumentNodeContentPrivate)
	documentGroup.Get("/:id", service.GetDocumentNodeContent)
	documentGroup.Post("/upload/:id", middleware.AuthNeeded(), service.UploadDocumentFile)
	documentGroup.Delete("/upload/:id", middleware.AuthNeeded(), service.DeleteDocumentFile)
	// documentGroup.Post("/category", middleware.AuthNeeded(), service.CreateDocumentCategory)
	// documentGroup.Patch("/category", middleware.AuthNeeded(), service.DeleteDocumentCategory)
	// documentGroup.Get("/categories", service.GetDocumentCategories)
	// documentGroup.Post("/tab", middleware.AuthNeeded(), service.CreateDocumentTab)
	// documentGroup.Patch("/tab", middleware.AuthNeeded(), service.DeleteDocumentTab)
	// documentGroup.Post("/tabs", service.GetDocumentTabs)
	// documentGroup.Post("/list", service.GetDocumentList)
	// documentGroup.Post("/private/list", middleware.AuthNeeded(), service.GetDocumentPrivateList)
	// documentGroup.Get("/:id", service.GetDocumentById)
	// documentGroup.Get("/private/:id", middleware.AuthNeeded(), service.GetDocumentPrivateById)
	// documentGroup.Post("/id", middleware.AuthNeeded(), service.CreateDocument)
	// documentGroup.Post("/upload/:id", middleware.AuthNeeded(), service.UploadDocumentFile)
	// documentGroup.Post("/delete/:id", middleware.AuthNeeded(), service.DeleteDocumentFile)
	// documentGroup.Patch("/:id", middleware.AuthNeeded(), service.UpdateDocument)
	// documentGroup.Delete("/:id", middleware.AuthNeeded(), service.DeleteDocument)
	// documentGroup.Post("/latest", middleware.AuthNeeded(), service.GetDocumentByNum)
	// documentGroup.Post("/search", service.SearchDocument)
	// documentGroup.Post("/private/search", middleware.AuthNeeded(), service.SearchDocument)

	(*router).Static("/contents", "./contents")
}
