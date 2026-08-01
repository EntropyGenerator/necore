package service

import (
	"necore/dao"
	"necore/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetWikiTags(c *fiber.Ctx) error {
	category := c.Query("category", "")
	if category != "glossary" && category != "item" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "category must be 'glossary' or 'item'",
		})
	}
	tags, err := dao.GetWikiTagsByCategory(category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if tags == nil {
		tags = make([]model.WikiTag, 0)
	}
	return c.JSON(fiber.Map{
		"tags": tags,
	})
}

func CreateWikiTag(c *fiber.Ctx) error {
	if !checkWikiPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden",
		})
	}
	var tag model.WikiTag
	if err := c.BodyParser(&tag); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}
	tag.Id = uuid.New().String()
	if err := dao.CreateWikiTag(tag); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"id": tag.Id,
	})
}

func DeleteWikiTag(c *fiber.Ctx) error {
	if !checkWikiPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden",
		})
	}
	if err := dao.DeleteWikiTag(c.Params("id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}