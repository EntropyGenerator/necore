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

func Initialize() *fiberAppInstance {
	app := fiber.New(fiber.Config{
		Prefork:                 false,
		AppName:                 "NMO Ecosystem Core",
		BodyLimit:               25 * 1024 * 1024,
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
