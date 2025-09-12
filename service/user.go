package service

import (
	"encoding/json"
	"necore/dao"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func GetUserInfo(c *fiber.Ctx) error {
	userId := c.Params("id")

	// // Check if user is admin or himself
	// token := c.Locals("user").(*jwt.Token)
	// isAdmin := dao.IsUserInGroup(token, "admin")
	// tokenUsername := dao.GetUsernameFromToken(token)
	// if !isAdmin && tokenUsername != userId {
	// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	// }

	userModel, err := dao.GetUserByUsername(userId)
	if err != nil || userModel == nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	type TagEntity struct {
		Text     string `json:"text"`
		Color    string `json:"color"`
		TagColor string `json:"tagColor"`
	}
	type UserInfo struct {
		Username string      `json:"username"`
		Group    []string    `json:"group"`
		Tags     []TagEntity `json:"tags"`
	}
	var groups []string
	err = json.Unmarshal([]byte(userModel.Group), &groups)
	if err != nil {
		// log.Println(err, " User info groups")
		groups = []string{}
	}
	var tags []TagEntity
	err = json.Unmarshal([]byte(userModel.Tags), &tags)
	if err != nil {
		// log.Println(err, " User info tags")
		tags = []TagEntity{}
	}
	// gjsonResult := gjson.Get(userModel.Group, "@this")
	// var groups []string
	// for _, value := range gjsonResult.Array() {
	// 	groups = append(groups, value.String())
	// }
	// gjsonResult = gjson.Get(userModel.Tags, "@this")
	// var tags []TagEntity
	// for _, value := range gjsonResult.Array() {
	// 	tags = append(tags, TagEntity{
	// 		Text:     value.Get("text").String(),
	// 		Color:    value.Get("color").String(),
	// 		TagColor: value.Get("tagColor").String(),
	// 	})
	// }
	return c.JSON(fiber.Map{
		"user": UserInfo{
			Username: userModel.Username,
			Group:    groups,
			Tags:     tags,
		}})
}

func GetUserList(c *fiber.Ctx) error {
	type TagEntity struct {
		Text     string `json:"text"`
		Color    string `json:"color"`
		TagColor string `json:"tagColor"`
	}
	type UserInfo struct {
		Username string      `json:"username"`
		Group    []string    `json:"group"`
		Tags     []TagEntity `json:"tags"`
	}
	users, err := dao.GetAllUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	userinfos := make([]UserInfo, len(users))

	for i, user := range users {
		var groups []string
		err = json.Unmarshal([]byte(user.Group), &groups)
		if err != nil {
			// log.Println(err, " User info groups")
			groups = []string{}
		}
		var tags []TagEntity
		err = json.Unmarshal([]byte(user.Tags), &tags)
		if err != nil {
			// log.Println(err, " User info Tags")
			tags = []TagEntity{}
		}
		// gjsonResult := gjson.Get(user.Group, "@this")
		// var groups []string
		// for _, value := range gjsonResult.Array() {
		// 	groups = append(groups, value.String())
		// }
		// gjsonResult = gjson.Get(user.Tags, "@this")
		// var tags []TagEntity
		// for _, value := range gjsonResult.Array() {
		// 	tags = append(tags, TagEntity{
		// 		Text:     value.Get("text").String(),
		// 		Color:    value.Get("color").String(),
		// 		TagColor: value.Get("tagColor").String(),
		// 	})
		// }
		userinfos[i] = UserInfo{
			Username: user.Username,
			Group:    groups,
			Tags:     tags,
		}
	}
	return c.JSON(fiber.Map{
		"users": userinfos,
	})
}

func DeleteUser(c *fiber.Ctx) error {
	// Must be admin
	token := c.Locals("user").(*jwt.Token)
	if !dao.IsUserInGroup(token, "admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	username := c.Params("id")
	err := dao.DeleteUserByUsername(username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	return c.SendStatus(200)
}

func UpdateUserPassword(c *fiber.Ctx) error {
	userId := c.Params("id")

	// Check if user is admin or himself
	token := c.Locals("user").(*jwt.Token)
	isAdmin := dao.IsUserInGroup(token, "admin")
	tokenUsername := dao.GetUsernameFromToken(token)
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

	if err := dao.UpdateUserPassword(payload.Id, payload.Password); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.SendStatus(fiber.StatusOK)

}

func UpdateUserInfo(c *fiber.Ctx) error {
	// Must be admin
	token := c.Locals("user").(*jwt.Token)
	if !dao.IsUserInGroup(token, "admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}
	type PayloadTags struct {
		Text     string `json:"text"`
		Color    string `json:"color"`
		TagColor string `json:"tagColor"`
	}
	type Payload struct {
		Username string        `json:"username"`
		Group    []string      `json:"group"`
		Tags     []PayloadTags `json:"Tags"`
	}
	payload := new(Payload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	groups, err := json.Marshal(payload.Group)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	tags, err := json.Marshal(payload.Tags)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := dao.UpdateUserInfo(payload.Username, string(groups), string(tags)); err != nil {
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

	avatar, err := dao.GetUserAvatar(userId)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	return c.JSON(fiber.Map{
		"avatar": avatar,
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
	isAdmin := dao.IsUserInGroup(token, "admin")
	tokenUsername := dao.GetUsernameFromToken(token)
	if !isAdmin && tokenUsername != payload.Username {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	if err := dao.UpdateUserAvatar(payload.Username, payload.Avatar); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.SendStatus(fiber.StatusOK)
}
