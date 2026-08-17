package service

import (
	"github.com/gofiber/fiber/v2"
)

const MaxUploadSize = 20 << 20

func rejectOversizedUpload(c *fiber.Ctx) error {
	return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
		"error": "File too large, maximum size is 20MB",
	})
}
