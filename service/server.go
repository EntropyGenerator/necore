package service

import (
	"necore/dao"
	"necore/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func checkServerPermission(c *fiber.Ctx) bool {
	// Check if user is admin or news_admin
	token := c.Locals("user").(*jwt.Token)
	isAdmin := dao.IsUserInGroup(token, "admin")
	isNewsAdmin := dao.IsUserInGroup(token, "server_admin")
	if isAdmin || isNewsAdmin {
		return false
	}
	return true
}

func GetServerList(c *fiber.Ctx) error {
	servers, err := dao.GetServerList()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"servers": servers,
	})
}

func AddServer(c *fiber.Ctx) error {
	if checkServerPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not allowed to add a server",
		})
	}
	var server model.Server
	if err := c.BodyParser(&server); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := dao.AddServer(server); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func DeleteServer(c *fiber.Ctx) error {
	if checkServerPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not allowed to delete a server",
		})
	}
	if err := dao.DeleteServer(c.Params("id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func UpdateServer(c *fiber.Ctx) error {
	if checkServerPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not allowed to update a server",
		})
	}
	var server model.Server
	if err := c.BodyParser(&server); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := dao.UpdateServer(server); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}
