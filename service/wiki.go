package service

import (
	"errors"
	"fmt"
	"necore/dao"
	"necore/model"
	"necore/util"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetWikiTypes(c *fiber.Ctx) error {
	glossaryTags, err := dao.GetWikiTagsByCategory("glossary")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	itemTags, err := dao.GetWikiTagsByCategory("item")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	glossaryNames := make([]string, len(glossaryTags))
	for i, t := range glossaryTags {
		glossaryNames[i] = t.Name
	}
	itemNames := make([]string, len(itemTags))
	for i, t := range itemTags {
		itemNames[i] = t.Name
	}

	if glossaryNames == nil {
		glossaryNames = []string{}
	}
	if itemNames == nil {
		itemNames = []string{}
	}

	return c.JSON(fiber.Map{
		"glossaryTypes": glossaryNames,
		"itemTypes":     itemNames,
	})
}

func checkWikiPermission(c *fiber.Ctx) bool {
	user := c.Locals("currentUser").(model.User)
	isAdmin := dao.ContainsGroup(user.Group, "admin")
	isDocsAdmin := dao.ContainsGroup(user.Group, "document_admin")
	return isAdmin || isDocsAdmin
}

// ---- Glossary handlers ----

func GetGlossaryList(c *fiber.Ctx) error {
	list, err := dao.GetGlossaryList()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if list == nil {
		list = make([]model.Glossary, 0)
	}
	return c.JSON(fiber.Map{
		"glossaries": list,
	})
}

func GetGlossaryById(c *fiber.Ctx) error {
	entry, err := dao.GetGlossaryById(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "glossary not found",
		})
	}
	return c.JSON(entry)
}

func CreateGlossary(c *fiber.Ctx) error {
	if !checkWikiPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to create glossary",
		})
	}
	id := uuid.New().String()
	entry := model.Glossary{Id: id}
	if err := c.BodyParser(&entry); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}
	entry.Id = id
	if err := dao.CreateGlossary(entry); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"id": id,
	})
}

func UpdateGlossary(c *fiber.Ctx) error {
	if !checkWikiPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to update glossary",
		})
	}
	var entry model.Glossary
	if err := c.BodyParser(&entry); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}
	entry.Id = c.Params("id")
	if err := dao.UpdateGlossary(entry); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func DeleteGlossary(c *fiber.Ctx) error {
	if !checkWikiPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to delete glossary",
		})
	}
	if err := dao.DeleteGlossary(c.Params("id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

// ---- Item handlers ----

func GetItemList(c *fiber.Ctx) error {
	list, err := dao.GetItemList()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if list == nil {
		list = make([]model.Item, 0)
	}
	return c.JSON(fiber.Map{
		"items": list,
	})
}

func GetItemById(c *fiber.Ctx) error {
	entry, err := dao.GetItemById(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "item not found",
		})
	}
	return c.JSON(entry)
}

func CreateItem(c *fiber.Ctx) error {
	if !checkWikiPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to create item",
		})
	}
	id := uuid.New().String()
	entry := model.Item{Id: id}
	if err := c.BodyParser(&entry); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}
	entry.Id = id
	if err := dao.CreateItem(entry); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"id": id,
	})
}

func UpdateItem(c *fiber.Ctx) error {
	if !checkWikiPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to update item",
		})
	}
	var entry model.Item
	if err := c.BodyParser(&entry); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}
	entry.Id = c.Params("id")
	if err := dao.UpdateItem(entry); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func DeleteItem(c *fiber.Ctx) error {
	if !checkWikiPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to delete item",
		})
	}
	if err := dao.DeleteItem(c.Params("id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

// ---- File upload for wiki ----

func generateWikiStoredFilename(originalName string) (string, error) {
	return generateStoredFilename(originalName)
}

func UploadWikiFile(c *fiber.Ctx) error {
	if !checkWikiPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden",
		})
	}
	id := c.Params("id")
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	dir := fmt.Sprintf("./contents/wiki/%s", id)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	storedName, err := generateWikiStoredFilename(file.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	contentPath, err := util.SafeContentPath("./contents/wiki", id, storedName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if err := c.SaveFile(file, contentPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"url": fmt.Sprintf("/contents/wiki/%s/%s", id, storedName)})
}

func DeleteWikiFile(c *fiber.Ctx) error {
	if !checkWikiPermission(c) {
		return c.Status(fiber.StatusForbidden).
			JSON(fiber.Map{"error": "Forbidden"})
	}
	id := c.Params("id")
	type Payload struct {
		Filename string `json:"filename"`
	}
	var payload Payload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "Invalid request body"})
	}
	target, err := util.SafeContentPath("./contents/wiki", id, payload.Filename)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "Invalid filename"})
	}
	if err := os.Remove(target); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return c.Status(fiber.StatusNotFound).
				JSON(fiber.Map{"error": "File not found"})
		}
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// generateStoredFilename is shared from article.go in the same package.
