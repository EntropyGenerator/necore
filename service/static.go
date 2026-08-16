package service

import (
	"necore/controller/middleware"
	"necore/dao"
	"necore/util"

	"github.com/gofiber/fiber/v2"
)

// ContentFileHandler serves uploaded files from ./contents/{objectId}/{filename}.
// Files belonging to a private document node (or a node under a private folder)
// require authentication and document_admin permission; everything else
// (articles, wiki, servers) stays public.
func ContentFileHandler(c *fiber.Ctx) error {
	objectID := c.Params("id")
	filename := c.Params("*")

	if objectID == "" || filename == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	private, err := dao.IsDocumentNodeEffectivelyPrivate(objectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}
	if private {
		if !middleware.Authenticate(c) {
			return nil
		}
		if !checkDocumentPermission(c) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden",
			})
		}
	}

	target, err := util.SafeContentPath("./contents", objectID, filename)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.SendFile(target)
}
