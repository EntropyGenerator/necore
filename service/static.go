package service

import (
	"necore/controller/middleware"
	"necore/dao"
	"necore/util"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ContentFileHandler serves uploaded files from ./contents/{objectId}/{filename}.
// Files belonging to a private document node (or a node under a private folder)
// require authentication and document_admin permission; everything else
// (articles, wiki, servers) stays public.
func ContentFileHandler(c *fiber.Ctx) error {
	first := c.Params("id")
	rest := c.Params("*")

	if first == "" || rest == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	baseDir := "./contents"
	objectID := first
	filename := rest

	if first == "wiki" {
		slash := strings.IndexByte(rest, '/')
		if slash <= 0 || slash == len(rest)-1 {
			return c.SendStatus(fiber.StatusNotFound)
		}
		baseDir = "./contents/wiki"
		objectID = rest[:slash]
		filename = rest[slash+1:]
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

	target, err := util.SafeContentPath(baseDir, objectID, filename)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.SendFile(target)
}
