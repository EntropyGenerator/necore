package service

import (
	"github.com/gofiber/fiber/v2"
)

// MaxUploadSize 限制单个上传文件的大小（20MB），防止大文件打满磁盘。
const MaxUploadSize = 20 << 20

// rejectOversizedUpload 返回 413 错误响应（文件超限）。
func rejectOversizedUpload(c *fiber.Ctx) error {
	return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
		"error": "File too large, maximum size is 20MB",
	})
}
