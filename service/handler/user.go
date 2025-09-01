package handler

import (
	"necore/service/impl"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func GetUserInfo(c *fiber.Ctx) error {
	userId := c.Params("id")

	// // Check if user is admin or himself
	// token := c.Locals("user").(*jwt.Token)
	// isAdmin := impl.IsUserInGroup(token, "admin")
	// tokenUsername := impl.GetUsernameFromToken(token)
	// if !isAdmin && tokenUsername != userId {
	// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	// }

	userModel, err := impl.GetUserByUsername(userId)
	if err != nil || userModel == nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}
	type UserInfo struct {
		Username string `json:"username"`
		Group    string `json:"group"`
		Tags     string `json:"tags"`
	}
	return c.JSON(fiber.Map{
		"user": UserInfo{
			Username: userModel.Username,
			Group:    userModel.Group,
			Tags:     userModel.Tags,
		}})
}

func GetUserList(c *fiber.Ctx) error {
	type UserInfo struct {
		Username string `json:"username"`
		Group    string `json:"group"`
		Tags     string `json:"Tags"`
	}
	users, err := impl.GetAllUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	userinfos := make([]UserInfo, len(users))
	for i, user := range users {
		userinfos[i] = UserInfo{
			Username: user.Username,
			Group:    user.Group,
			Tags:     user.Tags,
		}
	}
	return c.JSON(fiber.Map{
		"users": userinfos,
	})
}

func DeleteUser(c *fiber.Ctx) error {
	// Must be admin
	token := c.Locals("user").(*jwt.Token)
	if !impl.IsUserInGroup(token, "admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	username := c.Params("id")
	err := impl.DeleteUserByUsername(username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	return c.SendStatus(200)
}

func UpdateUserPassword(c *fiber.Ctx) error {
	userId := c.Params("id")

	// Check if user is admin or himself
	token := c.Locals("user").(*jwt.Token)
	isAdmin := impl.IsUserInGroup(token, "admin")
	tokenUsername := impl.GetUsernameFromToken(token)
	if !isAdmin && tokenUsername != userId {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	type Payload struct {
		Id       string `json:"id"`
		Password string `json:"new_password"`
	}
	payload := new(Payload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := impl.UpdateUserPassword(payload.Id, payload.Password); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.SendStatus(fiber.StatusOK)

}

func UpdateUserInfo(c *fiber.Ctx) error {
	// Must be admin
	token := c.Locals("user").(*jwt.Token)
	if !impl.IsUserInGroup(token, "admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	type Payload struct {
		Username string `json:"username"`
		Group    string `json:"group"`
		Tags     string `json:"Tags"`
	}
	payload := new(Payload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := impl.UpdateUserInfo(payload.Username, payload.Group, payload.Tags); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func Logout(c *fiber.Ctx) error {
	// TODO: expire token
	return c.SendStatus(fiber.StatusOK)
}

func GetUserAvatar(c *fiber.Ctx) error {
	userId := c.Params("id")

	avatar, err := impl.GetUserAvatar(userId)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	return c.JSON(fiber.Map{
		avatar: avatar,
	})
}

func UpdateUserAvatar(c *fiber.Ctx) error {
	type Payload struct {
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	}
	payload := new(Payload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check if user is admin or himself
	token := c.Locals("user").(*jwt.Token)
	isAdmin := impl.IsUserInGroup(token, "admin")
	tokenUsername := impl.GetUsernameFromToken(token)
	if !isAdmin && tokenUsername != payload.Username {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	if err := impl.UpdateUserAvatar(payload.Username, payload.Avatar); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.SendStatus(fiber.StatusOK)
}
