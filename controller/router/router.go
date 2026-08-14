package router

import (
	"necore/app"
	"necore/config"
	"necore/controller/middleware"
	"necore/service"
	"strconv"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
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
	loginLimiter := limiter.New(limiter.Config{
		Max:        8,
		Expiration: time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).
				JSON(fiber.Map{"error": "Too many login attempts"})
		},
	})

	// 全局限流（DDOS 防护）：按 IP 限制 API 请求频率，防止单来源打满后端。
	// 静态文件 /contents 与 /slogan 不限流（浏览器加载图片会大量占用额度）。
	rateLimitMax := 600
	if v, err := strconv.Atoi(config.Config("RATE_LIMIT_MAX")); err == nil && v > 0 {
		rateLimitMax = v
	}
	rateLimitExpiration := 60
	if v, err := strconv.Atoi(config.Config("RATE_LIMIT_EXPIRATION")); err == nil && v > 0 {
		rateLimitExpiration = v
	}
	globalLimiter := limiter.New(limiter.Config{
		Max:        rateLimitMax,
		Expiration: time.Duration(rateLimitExpiration) * time.Second,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).
				JSON(fiber.Map{"error": "Too many requests"})
		},
	})

	router := instance.Router
	(*router).Get("/slogan", service.SloganHandler)

	authGroup := (*router).Group("/auth")
	authGroup.Use(globalLimiter)
	authGroup.Get("/status", middleware.AuthNeeded(), service.GetStatus)
	authGroup.Post("/login", loginLimiter, service.Login)
	authGroup.Post("/register", middleware.AuthNeeded(), service.AddUser)
	authGroup.Get("/user/:id", service.GetUserInfo)
	authGroup.Get("/avatar/:id", service.GetUserAvatar)
	authGroup.Get("/userlist", middleware.AuthNeeded(), service.GetUserList)
	authGroup.Delete("/user/:id", middleware.AuthNeeded(), service.DeleteUser)
	authGroup.Post("/password", middleware.AuthNeeded(), service.UpdateUserPassword)
	authGroup.Post("/avatar", middleware.AuthNeeded(), service.UpdateUserAvatar)
	authGroup.Patch("/user", middleware.AuthNeeded(), service.UpdateUserInfo)

	articleGroup := (*router).Group("/news")
	articleGroup.Use(globalLimiter)
	articleGroup.Get("/total/:target", service.GetArticleCountByCategory)
	articleGroup.Post("/list", service.GetArticleList)
	articleGroup.Get("/detail/:id", service.GetArticleById)
	articleGroup.Patch("/:id", middleware.AuthNeeded(), service.UpdateArticle)
	articleGroup.Post("/upload/:id", middleware.AuthNeeded(), service.UploadArticleFile)
	articleGroup.Delete("/upload/:id", middleware.AuthNeeded(), service.DeleteArticleFile)
	articleGroup.Post("/create", middleware.AuthNeeded(), service.CreateArticle)
	articleGroup.Delete("/:id", middleware.AuthNeeded(), service.DeleteArticle)

	serverGroup := (*router).Group("/server")
	serverGroup.Use(globalLimiter)
	serverGroup.Get("/", service.GetServerList)
	serverGroup.Post("/status", service.GetServerStatus)
	serverGroup.Post("/create", middleware.AuthNeeded(), service.AddServer)
	serverGroup.Delete("/:id", middleware.AuthNeeded(), service.DeleteServer)
	serverGroup.Patch("/", middleware.AuthNeeded(), service.UpdateServer)

	documentGroup := (*router).Group("/documents")
	documentGroup.Use(globalLimiter)
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
	(*router).Get("/contents/:id/*", service.ContentFileHandler)

	botGroup := (*router).Group("/bots")
	botGroup.Use(globalLimiter)

	botGroup.Use("/ws/updates/:identifier", service.BotConectionChecker)
	botGroup.Get("/ws/updates/:identifier", websocket.New(service.HandleWSConnection))

	botGroup.Post("/token", middleware.AuthNeeded(), service.CreateBotToken)
	botGroup.Get("/token", middleware.AuthNeeded(), service.GetBotTokenList)
	botGroup.Get("/token/:id", middleware.AuthNeeded(), service.GetBotToken)
	botGroup.Delete("/token/:id", middleware.AuthNeeded(), service.DeleteBotToken)
	botGroup.Get("/status", middleware.AuthNeeded(), service.GetWSStatus)
	botGroup.Delete("/ws/kick/:session_id", middleware.AuthNeeded(), service.KickConnection)

	departmentGroup := (*router).Group("/department")
	departmentGroup.Use(globalLimiter)
	departmentGroup.Get("/", service.GetDepartmentList)
	departmentGroup.Post("/create", middleware.AuthNeeded(), service.CreateDepartment)
	departmentGroup.Patch("/", middleware.AuthNeeded(), service.UpdateDepartment)
	departmentGroup.Patch("/order", middleware.AuthNeeded(), service.UpdateDepartmentOrder)
	departmentGroup.Delete("/:id", middleware.AuthNeeded(), service.DeleteDepartment)
	departmentGroup.Post("/:id/member", middleware.AuthNeeded(), service.AddDepartmentMember)
	departmentGroup.Delete("/:id/member/:username", middleware.AuthNeeded(), service.RemoveDepartmentMember)
	departmentGroup.Patch("/:id/member/:username/leader", middleware.AuthNeeded(), service.UpdateDepartmentMemberLeaderStatus)
	departmentGroup.Patch("/:id/member/order", middleware.AuthNeeded(), service.UpdateDepartmentMemberOrder)

	wikiGroup := (*router).Group("/wiki")
	wikiGroup.Use(globalLimiter)
	wikiGroup.Get("/types", service.GetWikiTypes)
	wikiGroup.Get("/tags", service.GetWikiTags)
	wikiGroup.Post("/tags", middleware.AuthNeeded(), service.CreateWikiTag)
	wikiGroup.Delete("/tags/:id", middleware.AuthNeeded(), service.DeleteWikiTag)
	wikiGroup.Get("/glossary", service.GetGlossaryList)
	wikiGroup.Get("/glossary/:id", service.GetGlossaryById)
	wikiGroup.Get("/item", service.GetItemList)
	wikiGroup.Get("/item/:id", service.GetItemById)

	wikiGroup.Post("/glossary", middleware.AuthNeeded(), service.CreateGlossary)
	wikiGroup.Patch("/glossary/:id", middleware.AuthNeeded(), service.UpdateGlossary)
	wikiGroup.Delete("/glossary/:id", middleware.AuthNeeded(), service.DeleteGlossary)
	wikiGroup.Post("/item", middleware.AuthNeeded(), service.CreateItem)
	wikiGroup.Patch("/item/:id", middleware.AuthNeeded(), service.UpdateItem)
	wikiGroup.Delete("/item/:id", middleware.AuthNeeded(), service.DeleteItem)

	wikiGroup.Post("/upload/:id", middleware.AuthNeeded(), service.UploadWikiFile)
	wikiGroup.Delete("/upload/:id", middleware.AuthNeeded(), service.DeleteWikiFile)
}
