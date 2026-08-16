package app

import (
	"fmt"
	"log"
	"necore/config"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type fiberAppInstance struct {
	App *fiber.App
}

var instance *fiberAppInstance

// Initialize 在 config.Init 之后创建 fiber 应用（依赖 .env 配置）。
func Initialize() *fiberAppInstance {
	app := fiber.New(fiber.Config{
		Prefork:   false,
		AppName:   "NMO Ecosystem Core",
		BodyLimit: 25 * 1024 * 1024, // 单文件上限 20MB，25MB 在解析前拦截超大请求体

		// 反向代理场景：只有来自可信代理的 X-Forwarded-For 才参与限流 IP 判定，
		// 否则 nginx 后的所有用户共享同一 socket IP 的限流桶，任一人可打满全局限流。
		EnableTrustedProxyCheck: true,
		TrustedProxies:          parseTrustedProxies(config.Config("TRUSTED_PROXIES")),
		ProxyHeader:             fiber.HeaderXForwardedFor,
	})

	instance = &fiberAppInstance{
		App: app,
	}
	return instance
}

func parseTrustedProxies(raw string) []string {
	if raw = strings.TrimSpace(raw); raw == "" {
		return []string{}
	}
	var proxies []string
	for _, part := range strings.Split(raw, ",") {
		if part = strings.TrimSpace(part); part != "" {
			proxies = append(proxies, part)
		}
	}
	return proxies
}

func GetInstance() *fiberAppInstance {
	if instance == nil {
		log.Fatal("app not initialized: call app.Initialize() after config.Init()")
	}
	return instance
}

func Start() {
	port, err := strconv.Atoi(config.Config("PORT"))
	if err != nil {
		port = 3000
	}
	log.Fatal(instance.App.Listen(fmt.Sprintf(":%d", port)))
}
