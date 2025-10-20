package service

import (
	"encoding/json"
	"necore/dao"
	"necore/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// type docBrief struct {
// 	Id          string `json:"id"`
// 	Title       string `json:"title"`
// 	Description string `json:"description"`
// }

// func patchDocumentList(docs []model.Document) []docBrief {
// 	docNum := len(docs)
// 	docBriefs := make([]docBrief, docNum)
// 	for i := 0; i < docNum; i++ {
// 		docBriefs[i] = docBrief{
// 			Id:          docs[i].Id,
// 			Title:       docs[i].Title,
// 			Description: docs[i].Description,
// 		}
// 	}
// 	return docBriefs
// }

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

	Name         string       `json:"name"`
	Contributors []string     `json:"contributors"`
	Content      []docContent `json:"content"`
	UpdateTime   string       `json:"updateTime"`
}

func marshalDocNode(doc *model.DocumentNode) docNode {
	var contents []docContent
	if err := json.Unmarshal([]byte(doc.Content), &contents); err != nil {
		contents = make([]docContent, 0)
	}
	var contributors []string
	if err := json.Unmarshal([]byte(doc.Contributors), &contributors); err != nil {
		contributors = make([]string, 0)
	}
	return docNode{
		ParentId: doc.ParentId,
		Id:       doc.Id,
		IsFolder: doc.IsFolder,
		Private:  doc.Private,

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
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	marshalledNodeList := make([]docNode, len(nodeList))
	for i, node := range nodeList {
		marshalledNodeList[i] = marshalDocNode(&node)
	}
	return c.JSON(fiber.Map{
		"children": marshalledNodeList,
	})
}

func GetDocumentNodeChildren(c *fiber.Ctx) error {
	id := c.Params("parentId")
	nodeList, err := dao.GetDocumentNodeChildren(id, false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	marshalledNodeList := make([]docNode, len(nodeList))
	for i, node := range nodeList {
		marshalledNodeList[i] = marshalDocNode(&node)
	}
	return c.JSON(fiber.Map{
		"children": marshalledNodeList,
	})
}

// func CreateDocumentCategory(c *fiber.Ctx) error {
// 	if checkDocumentPermission(c) {
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"error": "You don't have permission to create document category",
// 		})
// 	}

// 	type Payload struct {
// 		Category string `json:"category"`
// 	}
// 	payload := new(Payload)
// 	if err := c.BodyParser(payload); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}

// 	if payload.Category == "" {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Category name cannot be empty",
// 		})
// 	}

// 	if err := dao.CreateDocumentCategory(payload.Category); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}

// 	return c.SendStatus(fiber.StatusOK)
// }

// func DeleteDocumentCategory(c *fiber.Ctx) error {
// 	if checkDocumentPermission(c) {
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"error": "You don't have permission to delete document category",
// 		})
// 	}

// 	type Payload struct {
// 		Category string `json:"category"`
// 	}
// 	payload := new(Payload)
// 	if err := c.BodyParser(payload); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}

// 	if err := dao.DeleteDocumentCategory(payload.Category); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}

// 	return c.SendStatus(fiber.StatusOK)
// }

// func GetDocumentCategories(c *fiber.Ctx) error {
// 	categories, err := dao.GetDocumentAllCategory()
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	} else {
// 		patchedCategories := make([]string, len(categories))
// 		for i, category := range categories {
// 			patchedCategories[i] = category.Category
// 		}
// 		return c.Status(fiber.StatusOK).JSON(fiber.Map{"categories": patchedCategories})
// 	}
// }

// func CreateDocumentTab(c *fiber.Ctx) error {
// 	if checkDocumentPermission(c) {
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"error": "You don't have permission to create document tab",
// 		})
// 	}
// 	type payload struct {
// 		Category string `json:"category"`
// 		Tab      string `json:"tab"`
// 	}
// 	payloads := new(payload)
// 	if err := c.BodyParser(payloads); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}

// 	if payloads.Category == "" || payloads.Tab == "" {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}

// 	if err := dao.CreateDocumentTab(payloads.Category, payloads.Tab); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}

// 	return c.SendStatus(fiber.StatusOK)
// }

// func DeleteDocumentTab(c *fiber.Ctx) error {
// 	if checkDocumentPermission(c) {
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"error": "You don't have permission to delete document tab",
// 		})
// 	}

// 	type payload struct {
// 		Category string `json:"category"`
// 		Tab      string `json:"tab"`
// 	}
// 	payloads := new(payload)
// 	if err := c.BodyParser(payloads); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}
// 	if err := dao.DeleteDocumentTab(payloads.Category, payloads.Tab); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}

// 	return c.SendStatus(fiber.StatusOK)
// }

// func GetDocumentTabs(c *fiber.Ctx) error {
// 	type payload struct {
// 		Category string `json:"category"`
// 	}
// 	payloads := new(payload)
// 	if err := c.BodyParser(payloads); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}
// 	tabs, err := dao.GetDocumentAllTab(payloads.Category)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	} else {
// 		patchedTabs := make([]string, len(tabs))
// 		for i, tab := range tabs {
// 			patchedTabs[i] = tab.Tab
// 		}
// 		return c.Status(fiber.StatusOK).JSON(fiber.Map{"tabs": patchedTabs})
// 	}
// }

// func GetDocumentById(c *fiber.Ctx) error {
// 	doc, err := dao.GetDocument(c.Params("id"), false)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}

// 	type PayloadContent struct {
// 		Type    string `json:"type"`
// 		Content string `json:"content"`
// 	}
// 	type Payload struct {
// 		Id           string           `json:"id"`
// 		Title        string           `json:"title"`
// 		Description  string           `json:"description"`
// 		Category     string           `json:"category"`
// 		Tab          string           `json:"tab"`
// 		Contributors []string         `json:"contributors"`
// 		Priority     int              `json:"priority"`
// 		Content      []PayloadContent `json:"content"`
// 		CreateTime   string           `json:"createTime"`
// 		UpdateTime   string           `json:"updateTime"`
// 	}
// 	var payloadContent []PayloadContent
// 	json.Unmarshal([]byte(doc.Content), &payloadContent)
// 	var contributors []string
// 	json.Unmarshal([]byte(doc.Contributors), &contributors)
// 	p := Payload{
// 		Id:           doc.Id,
// 		Title:        doc.Title,
// 		Description:  doc.Description,
// 		Category:     doc.Category,
// 		Tab:          doc.Tab,
// 		Contributors: contributors,
// 		Priority:     doc.Priority,
// 		Content:      payloadContent,
// 		CreateTime:   doc.CreateTime,
// 		UpdateTime:   doc.UpdateTime,
// 	}
// 	return c.Status(fiber.StatusOK).JSON(p)
// }

// func GetDocumentPrivateById(c *fiber.Ctx) error {
// 	doc, err := dao.GetDocument(c.Params("id"), true)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}

// 	type PayloadContent struct {
// 		Type    string `json:"type"`
// 		Content string `json:"content"`
// 	}
// 	type Payload struct {
// 		Id           string           `json:"id"`
// 		Title        string           `json:"title"`
// 		Description  string           `json:"description"`
// 		Category     string           `json:"category"`
// 		Tab          string           `json:"tab"`
// 		Contributors []string         `json:"contributors"`
// 		Priority     int              `json:"priority"`
// 		Content      []PayloadContent `json:"content"`
// 		CreateTime   string           `json:"createTime"`
// 		UpdateTime   string           `json:"updateTime"`
// 	}
// 	var payloadContent []PayloadContent
// 	json.Unmarshal([]byte(doc.Content), &payloadContent)
// 	var contributors []string
// 	json.Unmarshal([]byte(doc.Contributors), &contributors)
// 	p := Payload{
// 		Id:           doc.Id,
// 		Title:        doc.Title,
// 		Description:  doc.Description,
// 		Category:     doc.Category,
// 		Tab:          doc.Tab,
// 		Contributors: contributors,
// 		Priority:     doc.Priority,
// 		Content:      payloadContent,
// 		CreateTime:   doc.CreateTime,
// 		UpdateTime:   doc.UpdateTime,
// 	}
// 	return c.Status(fiber.StatusOK).JSON(p)
// }

// func GetDocumentList(c *fiber.Ctx) error {
// 	type request struct {
// 		Category string `json:"category"`
// 		Tab      string `json:"tab"`
// 	}
// 	requests := new(request)
// 	if err := c.BodyParser(requests); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}

// 	documents, err := dao.GetDocumentListByClass(requests.Category, requests.Tab, false)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	patchedDocBriefs := make([]docBrief, len(documents))
// 	for i, doc := range documents {
// 		patchedDocBriefs[i] = docBrief{
// 			Id:          doc.Id,
// 			Title:       doc.Title,
// 			Description: doc.Description,
// 		}
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{"documents": patchedDocBriefs})
// }

// func GetDocumentPrivateList(c *fiber.Ctx) error {
// 	type request struct {
// 		Category string `json:"category"`
// 		Tab      string `json:"tab"`
// 	}
// 	requests := new(request)
// 	if err := c.BodyParser(requests); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}

// 	documents, err := dao.GetDocumentListByClass(requests.Category, requests.Tab, true)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	patchedDocBriefs := make([]docBrief, len(documents))
// 	for i, doc := range documents {
// 		patchedDocBriefs[i] = docBrief{
// 			Id:          doc.Id,
// 			Title:       doc.Title,
// 			Description: doc.Description,
// 		}
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{"documents": patchedDocBriefs})
// }

// func CreateDocument(c *fiber.Ctx) error {
// 	if checkDocumentPermission(c) {
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"error": "You don't have permission to create document",
// 		})
// 	}
// 	uuid := uuid.New().String()
// 	err := dao.CreateDocument(uuid)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{"id": uuid})
// }

// func UploadDocumentFile(c *fiber.Ctx) error {
// 	if checkDocumentPermission(c) {
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"error": "You don't have permission to upload document",
// 		})
// 	}

// 	id := c.Params("id")
// 	file, err := c.FormFile("file")
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}
// 	if err := os.MkdirAll(fmt.Sprintf("./contents/%s", id), os.ModePerm); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	if err := c.SaveFile(file, fmt.Sprintf("./contents/%s/%s", id, file.Filename)); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{"url": fmt.Sprintf("/contents/%s/%s", id, file.Filename)})
// }

// func DeleteDocumentFile(c *fiber.Ctx) error {
// 	if checkDocumentPermission(c) {
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"error": "You don't have permission to delete document",
// 		})
// 	}
// 	id := c.Params("id")
// 	type request struct {
// 		Url string `json:"url"`
// 	}
// 	requests := new(request)
// 	if err := c.BodyParser(requests); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}
// 	if err := os.Remove(fmt.Sprintf("./contents/%s/%s", id, requests.Url)); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{"url": fmt.Sprintf("/contents/%s/%s", c.Params("id"), c.Params("filename"))})
// }

// func UpdateDocument(c *fiber.Ctx) error {
// 	if checkDocumentPermission(c) {
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"error": "You don't have permission to update document",
// 		})
// 	}
// 	type PayloadContent struct {
// 		Type    string `json:"type"`
// 		Content string `json:"content"`
// 	}
// 	type Payload struct {
// 		Id          string           `json:"id"`
// 		Title       string           `json:"title"`
// 		Description string           `json:"description"`
// 		Category    string           `json:"category"`
// 		Tab         string           `json:"tab"`
// 		Priority    int              `json:"priority"`
// 		Content     []PayloadContent `json:"content"`
// 		Contributor string           `json:"contributor"`
// 		Private     bool             `json:"private"`
// 	}
// 	payload := new(Payload)
// 	if err := c.BodyParser(payload); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}
// 	doc, err := dao.GetDocument(payload.Id, true)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	var contributors []string
// 	json.Unmarshal([]byte(doc.Contributors), &contributors)
// 	contributors = append(contributors, payload.Contributor)
// 	newContent, err := json.Marshal(payload.Content)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}

// 	deduplicatedContributors := make(map[string]bool)
// 	for _, contributor := range contributors {
// 		deduplicatedContributors[contributor] = true
// 	}
// 	contributorsList := make([]string, 0, len(deduplicatedContributors))
// 	for contributor := range deduplicatedContributors {
// 		contributorsList = append(contributorsList, contributor)
// 	}
// 	newContributors, err := json.Marshal(contributorsList)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	newDoc := model.Document{
// 		Id:           payload.Id,
// 		Title:        payload.Title,
// 		Description:  payload.Description,
// 		Category:     payload.Category,
// 		Tab:          payload.Tab,
// 		Priority:     payload.Priority,
// 		Content:      string(newContent),
// 		Contributors: string(newContributors),
// 		Private:      payload.Private,
// 	}
// 	if err := dao.UpdateDocument(newDoc); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	return c.SendStatus(fiber.StatusOK)
// }

// func DeleteDocument(c *fiber.Ctx) error {
// 	if checkDocumentPermission(c) {
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"error": "You don't have permission to delete document",
// 		})
// 	}
// 	id := c.Params("id")
// 	if err := dao.DeleteDocument(id); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	if err := os.RemoveAll(fmt.Sprintf("./contents/%s", id)); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	return c.SendStatus(fiber.StatusOK)
// }

// func GetDocumentByNum(c *fiber.Ctx) error {
// 	isAdmin := checkDocumentPermission(c)
// 	type request struct {
// 		PageSize int `json:"pageSize"`
// 		Page     int `json:"page"`
// 	}
// 	requests := new(request)
// 	if err := c.BodyParser(requests); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}
// 	documents, err := dao.GetDocumentListByNum((requests.Page-1)*requests.PageSize, requests.PageSize, isAdmin)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	docBriefs := make([]docBrief, len(documents))
// 	for i, doc := range documents {
// 		docBriefs[i] = docBrief{
// 			Id:          doc.Id,
// 			Title:       doc.Title,
// 			Description: doc.Description,
// 		}
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"documents": docBriefs,
// 	})
// }

// func SearchDocument(c *fiber.Ctx) error {
// 	type request struct {
// 		Keyword  string `json:"keyword"`
// 		PageSize int    `json:"pageSize"`
// 		Page     int    `json:"page"`
// 	}
// 	requests := new(request)
// 	if err := c.BodyParser(requests); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}
// 	documents, err := dao.SearchDocument(requests.Keyword, (requests.Page-1)*requests.PageSize, requests.PageSize, false)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	docBriefs := make([]docBrief, len(documents))
// 	for i, doc := range documents {
// 		docBriefs[i] = docBrief{
// 			Id:          doc.Id,
// 			Title:       doc.Title,
// 			Description: doc.Description,
// 		}
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"documents": docBriefs,
// 	})
// }

// func SearchDocumentPrivate(c *fiber.Ctx) error {
// 	type request struct {
// 		Keyword  string `json:"keyword"`
// 		PageSize int    `json:"pageSize"`
// 		Page     int    `json:"page"`
// 	}
// 	requests := new(request)
// 	if err := c.BodyParser(requests); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request",
// 		})
// 	}
// 	documents, err := dao.SearchDocument(requests.Keyword, (requests.Page-1)*requests.PageSize, requests.PageSize, true)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	docBriefs := make([]docBrief, len(documents))
// 	for i, doc := range documents {
// 		docBriefs[i] = docBrief{
// 			Id:          doc.Id,
// 			Title:       doc.Title,
// 			Description: doc.Description,
// 		}
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"documents": docBriefs,
// 	})
// }
