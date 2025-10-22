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

func checkDocumentPermission(c *fiber.Ctx) bool {
	// Check if user is admin or document_admin
	token := c.Locals("user").(*jwt.Token)
	isAdmin := dao.IsUserInGroup(token, "admin")
	isDocsAdmin := dao.IsUserInGroup(token, "document_admin")
	if isAdmin || isDocsAdmin {
		return true
	}
	return false
}

func CreateDocumentNode(c *fiber.Ctx) error {
	if !checkDocumentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to create document node",
		})
	}

	type request struct {
		ParentId string `json:"parentId"`
		IsFolder bool   `json:"isFolder"`
		Private  bool   `json:"private"`
		Name     string `json:"name"`
	}
	r := new(request)
	if err := c.BodyParser(r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}
	uuid := uuid.New().String()

	if err := dao.CreateDocumentNode(r.ParentId, r.IsFolder, r.Private, r.Name, uuid); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"id": uuid,
	})
}

func DeleteDocumentNode(c *fiber.Ctx) error {
	if !checkDocumentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to delete document node",
		})
	}

	id := c.Params("id")
	if err := dao.DeleteDocumentNode(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func UpdateDocumentNodeParentId(c *fiber.Ctx) error {
	if !checkDocumentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to update document node parent id",
		})
	}
	id := c.Params("id")
	type request struct {
		ParentId string `json:"parentId"`
	}
	r := new(request)
	if err := c.BodyParser(r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}
	if err := dao.UpdateDocumentNodeParentId(id, r.ParentId); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func UpdateDocumentNodeContent(c *fiber.Ctx) error {
	if !checkDocumentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to update document node content",
		})
	}
	id := c.Params("id")

	token := c.Locals("user").(*jwt.Token)
	username := dao.GetUsernameFromToken(token)

	type contentRequest struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	}
	type request struct {
		Private bool             `json:"private"`
		Content []contentRequest `json:"content"`
	}
	r := new(request)
	if err := c.BodyParser(r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	marshalledContent, err := json.Marshal(r.Content)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := dao.UpdateDocumentNodeContent(id, string(marshalledContent), r.Private, username); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func UpdateDocumentNodeName(c *fiber.Ctx) error {
	if !checkDocumentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to update document node name",
		})
	}
	id := c.Params("id")
	type request struct {
		Name string `json:"name"`
	}
	r := new(request)
	if err := c.BodyParser(r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}
	if err := dao.UpdateDocumentNodeName(id, r.Name); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

type docContent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type docNode struct {
	ParentId string `json:"parentId"`
	Id       string `json:"id"`
	IsFolder bool   `json:"isFolder"`
	Private  bool   `json:"private"`
	Name     string `json:"name"`
}

type docNodeWithContent struct {
	ParentId     string       `json:"parentId"`
	Id           string       `json:"id"`
	IsFolder     bool         `json:"isFolder"`
	Private      bool         `json:"private"`
	Name         string       `json:"name"`
	Contributors []string     `json:"contributors"`
	Content      []docContent `json:"content"`
	UpdateTime   string       `json:"updateTime"`
}

func marshalDocNode(doc *model.DocumentNode) docNode {
	return docNode{
		ParentId: doc.ParentId,
		Id:       doc.Id,
		IsFolder: doc.IsFolder,
		Private:  doc.Private,
		Name:     doc.Name,
	}
}

func marshalDocNodeWithContent(doc *model.DocumentNode) docNodeWithContent {
	var contents []docContent
	if err := json.Unmarshal([]byte(doc.Content), &contents); err != nil {
		contents = make([]docContent, 0)
	}
	var contributors []string
	if err := json.Unmarshal([]byte(doc.Contributors), &contributors); err != nil {
		contributors = make([]string, 0)
	}
	return docNodeWithContent{
		ParentId:     doc.ParentId,
		Id:           doc.Id,
		IsFolder:     doc.IsFolder,
		Private:      doc.Private,
		Name:         doc.Name,
		Contributors: contributors,
		Content:      contents,
		UpdateTime:   doc.UpdateTime,
	}
}

func GetDocumentNodeChildrenPrivate(c *fiber.Ctx) error {
	if !checkDocumentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to update document node name",
		})
	}
	id := c.Params("parentId")

	nodeList, err := dao.GetDocumentNodeChildren(id, true)

	marshalledNodeList := make([]docNode, len(nodeList))
	for i, node := range nodeList {
		marshalledNodeList[i] = marshalDocNode(&node)
	}
	if err != nil {
		return c.JSON(fiber.Map{
			"error":    err.Error(),
			"children": marshalledNodeList,
		})
	}
	return c.JSON(fiber.Map{
		"children": marshalledNodeList,
	})
}

func GetDocumentNodeChildren(c *fiber.Ctx) error {
	id := c.Params("parentId")
	nodeList, err := dao.GetDocumentNodeChildren(id, false)

	marshalledNodeList := make([]docNode, len(nodeList))
	for i, node := range nodeList {
		marshalledNodeList[i] = marshalDocNode(&node)
	}

	if err != nil {
		return c.JSON(fiber.Map{
			"error":    err.Error(),
			"children": marshalledNodeList,
		})
	}
	return c.JSON(fiber.Map{
		"children": marshalledNodeList,
	})
}

func GetDocumentNodeContentPrivate(c *fiber.Ctx) error {
	if !checkDocumentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to update document node name",
		})
	}
	id := c.Params("id")
	node, err := dao.GetDocumentContent(id, true)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(marshalDocNodeWithContent(&node))
}

func GetDocumentNodeContent(c *fiber.Ctx) error {
	id := c.Params("id")
	node, err := dao.GetDocumentContent(id, false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(marshalDocNodeWithContent(&node))
}

func UploadDocumentFile(c *fiber.Ctx) error {
	if !checkDocumentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to update document node name",
		})
	}
	id := c.Params("id")
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := os.MkdirAll(fmt.Sprintf("./contents/%s", id), os.ModePerm); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	if err := c.SaveFile(file, fmt.Sprintf("./contents/%s/%s", id, file.Filename)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	return c.JSON(fiber.Map{"url": fmt.Sprintf("/contents/%s/%s", id, file.Filename)})
}

func DeleteDocumentFile(c *fiber.Ctx) error {
	if !checkDocumentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to update document node name",
		})
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
