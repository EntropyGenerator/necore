package service

import (
	"encoding/json"
	"fmt"
	"necore/dao"
	"necore/model"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func CreateArticle(c *fiber.Ctx) error {
	// Check if user is admin or news_admin
	token := c.Locals("user").(*jwt.Token)
	isAdmin := dao.IsUserInGroup(token, "admin")
	isNewsAdmin := dao.IsUserInGroup(token, "news_admin")
	if !isAdmin && isNewsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	// Create new article
	id := uuid.New().String()
	err := dao.CreateArticle(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err,
		})
	}
	return c.JSON(fiber.Map{
		"id": id,
	})
}

func UpdateArticle(c *fiber.Ctx) error {
	// Check if user is admin or news_admin
	token := c.Locals("user").(*jwt.Token)
	isAdmin := dao.IsUserInGroup(token, "admin")
	isNewsAdmin := dao.IsUserInGroup(token, "news_admin")
	if !isAdmin && isNewsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}
	author := dao.GetUsernameFromToken(token)

	id := c.Params("id")
	// Parse
	type PayloadEntity struct {
		Pin     bool   `json:"pin"`
		Title   string `json:"title"`
		Brief   string `json:"brief"`
		Date    string `json:"date"`
		EndDate string `json:"endDate"`
		Image   string `json:"image"`
	}
	type PayloadContent struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	}
	type Payload struct {
		Entity   PayloadEntity    `json:"entity"`
		Content  []PayloadContent `json:"content"`
		Category string           `json:"category"`
	}
	payload := new(Payload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err})
	}

	newContent, err := json.Marshal(payload.Content)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err})
	}

	newArticle := model.Article{
		Id:       id,
		Pin:      payload.Entity.Pin,
		Title:    payload.Entity.Title,
		Brief:    payload.Entity.Brief,
		Date:     payload.Entity.Date,
		EndDate:  payload.Entity.EndDate,
		Image:    payload.Entity.Image,
		Category: payload.Category,
		Content:  string(newContent),
		Author:   author,
	}

	if err := dao.UpdateArticle(newArticle); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}

	return c.SendStatus(fiber.StatusOK)
}

func GetArticleById(c *fiber.Ctx) error {
	id := c.Params("id")

	article, err := dao.GetArticle(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	type PayloadEntity struct {
		Pin     bool   `json:"pin"`
		Title   string `json:"title"`
		Brief   string `json:"brief"`
		Date    string `json:"date"`
		EndDate string `json:"endDate"`
		Image   string `json:"image"`
	}
	type PayloadContent struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	}
	type Payload struct {
		Entity   PayloadEntity    `json:"entity"`
		Content  []PayloadContent `json:"content"`
		Category string           `json:"category"`
		Author   string           `json:"author"`
	}
	payloadEntity := PayloadEntity{
		Pin:     article.Pin,
		Title:   article.Title,
		Brief:   article.Brief,
		Date:    article.Date,
		EndDate: article.EndDate,
		Image:   article.Image,
	}
	var payloadContent []PayloadContent
	json.Unmarshal([]byte(article.Content), &payloadContent)
	payload := Payload{
		Entity:   payloadEntity,
		Content:  payloadContent,
		Category: article.Category,
		Author:   article.Author,
	}
	return c.JSON(payload)
}

func GetArticleCountByCategory(c *fiber.Ctx) error {
	category := c.Params("target")
	// target: "information" | "magazine" | "notice" | "activity" | "document"
	count, err := dao.GetArticleCountByCategory(category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	return c.JSON(fiber.Map{"total": count})
}

func GetArticleList(c *fiber.Ctx) error {
	type Payload struct {
		Target   string `json:"target"`
		Page     int    `json:"page"`
		PageSize int    `json:"page_size"`
		Pin      bool   `json:"pin"`
	}
	payload := new(Payload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err})
	}
	articles, err := dao.GetArticleList(payload.Target, payload.Page, payload.PageSize, payload.Pin)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	type Entity struct {
		Id      string `json:"id"`
		Pin     bool   `json:"pin"`
		Title   string `json:"title"`
		Brief   string `json:"brief"`
		Date    string `json:"date"`
		EndDate string `json:"endDate"`
		Image   string `json:"image"`
	}
	var entities []Entity
	for _, article := range articles {
		entities = append(entities, Entity{
			Id:      article.Id,
			Pin:     article.Pin,
			Title:   article.Title,
			Brief:   article.Brief,
			Date:    article.Date,
			EndDate: article.EndDate,
			Image:   article.Image,
		})
	}
	return c.JSON(fiber.Map{"list": entities})
}

func UploadArticleFile(c *fiber.Ctx) error {
	// Check if user is admin or news_admin
	token := c.Locals("user").(*jwt.Token)
	isAdmin := dao.IsUserInGroup(token, "admin")
	isNewsAdmin := dao.IsUserInGroup(token, "news_admin")
	if !isAdmin && isNewsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	id := c.Params("id")
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err})
	}
	if err := os.MkdirAll(fmt.Sprintf("./contents/%s", id), os.ModePerm); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	if err := c.SaveFile(file, fmt.Sprintf("./contents/%s/%s", id, file.Filename)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	return c.JSON(fiber.Map{"url": fmt.Sprintf("/contents/%s/%s", id, file.Filename)})
}

func DeleteArticleFile(c *fiber.Ctx) error {
	// Check if user is admin or news_admin
	token := c.Locals("user").(*jwt.Token)
	isAdmin := dao.IsUserInGroup(token, "admin")
	isNewsAdmin := dao.IsUserInGroup(token, "news_admin")
	if !isAdmin && isNewsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	// id := c.Params("id") // It is included in the url
	type Payload struct {
		Url string `json:"url"`
	}
	payload := new(Payload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err})
	}
	if err := os.Remove(fmt.Sprintf("./%s", payload.Url)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	return c.SendStatus(fiber.StatusOK)
}

func DeleteArticle(c *fiber.Ctx) error {
	// Check if user is admin or news_admin
	token := c.Locals("user").(*jwt.Token)
	isAdmin := dao.IsUserInGroup(token, "admin")
	isNewsAdmin := dao.IsUserInGroup(token, "news_admin")
	if !isAdmin && isNewsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	id := c.Params("id")
	if err := dao.DeleteArticle(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	if err := os.RemoveAll(fmt.Sprintf("./contents/%s", id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	return c.SendStatus(fiber.StatusOK)
}
