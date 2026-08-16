package service

import (
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strings"

	"necore/dao"
	"necore/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type departmentTagEntity struct {
	Text     string `json:"text"`
	Color    string `json:"color"`
	TagColor string `json:"tagColor"`
}

type departmentMemberEntity struct {
	Username string                `json:"username"`
	Avatar   string                `json:"avatar,omitempty"`
	Group    []string              `json:"-"` // 权限组不对外暴露（公开部门列表）
	Tags     []departmentTagEntity `json:"tags"`
	IsLeader bool                  `json:"isLeader"`
}

type departmentMemberSortable struct {
	Member    departmentMemberEntity
	SortOrder int
}

func departmentNotFound(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "Department not found",
	})
}

func checkDepartmentPermission(c *fiber.Ctx) bool {
	user := c.Locals("currentUser").(model.User)
	isAdmin := dao.ContainsGroup(user.Group, "admin")
	if isAdmin {
		return false
	}
	return true
}

func parseDepartmentMember(user model.User, isLeader bool) departmentMemberEntity {
	var groups []string
	if err := json.Unmarshal([]byte(user.Group), &groups); err != nil {
		groups = []string{}
	}
	var tags []departmentTagEntity
	if err := json.Unmarshal([]byte(user.Tags), &tags); err != nil {
		tags = []departmentTagEntity{}
	}
	return departmentMemberEntity{
		Username: user.Username,
		Avatar:   user.Avatar,
		Group:    groups,
		Tags:     tags,
		IsLeader: isLeader,
	}
}

func buildDepartmentMembers(departmentID string) ([]departmentMemberEntity, error) {
	memberships, err := dao.GetDepartmentMembers(departmentID)
	if err != nil {
		return nil, err
	}

	sortable := make([]departmentMemberSortable, 0, len(memberships))
	for _, membership := range memberships {
		user, err := dao.GetUserByUsername(membership.Username)
		if err != nil {
			return nil, err
		}
		if user == nil {
			continue
		}
		sortable = append(sortable, departmentMemberSortable{
			Member:    parseDepartmentMember(*user, membership.IsLeader),
			SortOrder: membership.SortOrder,
		})
	}

	sort.SliceStable(sortable, func(i, j int) bool {
		if sortable[i].Member.IsLeader != sortable[j].Member.IsLeader {
			return sortable[i].Member.IsLeader
		}
		if sortable[i].SortOrder != sortable[j].SortOrder {
			return sortable[i].SortOrder < sortable[j].SortOrder
		}
		return sortable[i].Member.Username < sortable[j].Member.Username
	})

	result := make([]departmentMemberEntity, len(sortable))
	for i, item := range sortable {
		result[i] = item.Member
	}
	return result, nil
}

func GetDepartmentList(c *fiber.Ctx) error {
	departments, err := dao.GetDepartmentList()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	type Response struct {
		Id          string                   `json:"id"`
		Name        string                   `json:"name"`
		Description string                   `json:"description"`
		Icon        string                   `json:"icon"`
		SortOrder   int                      `json:"sortOrder"`
		Members     []departmentMemberEntity `json:"members"`
	}

	res := make([]Response, 0, len(departments))
	for _, department := range departments {
		members, err := buildDepartmentMembers(department.Id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		res = append(res, Response{
			Id:          department.Id,
			Name:        department.Name,
			Description: department.Description,
			Icon:        department.Icon,
			SortOrder:   department.SortOrder,
			Members:     members,
		})
	}

	return c.JSON(fiber.Map{
		"departments": res,
	})
}

func CreateDepartment(c *fiber.Ctx) error {
	if checkDepartmentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	type Payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		SortOrder   int    `json:"sortOrder"`
	}

	payload := new(Payload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if strings.TrimSpace(payload.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Department name is required",
		})
	}

	id := uuid.New().String()
	department := model.Department{
		Id:          id,
		Name:        payload.Name,
		Description: payload.Description,
		Icon:        payload.Icon,
		SortOrder:   payload.SortOrder,
	}
	if err := dao.CreateDepartment(department); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"id": id,
	})
}

func UpdateDepartment(c *fiber.Ctx) error {
	if checkDepartmentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	// 指针字段：只在请求中出现时才更新，省略字段保持不变
	type request struct {
		Id          string  `json:"id"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Icon        *string `json:"icon"`
		SortOrder   *int    `json:"sortOrder"`
	}

	payload := new(request)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if strings.TrimSpace(payload.Id) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Department id is required",
		})
	}

	fields := map[string]any{}
	if payload.Name != nil {
		name := strings.TrimSpace(*payload.Name)
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Department name is required",
			})
		}
		fields["name"] = name
	}
	if payload.Description != nil {
		fields["description"] = *payload.Description
	}
	if payload.Icon != nil {
		fields["icon"] = *payload.Icon
	}
	if payload.SortOrder != nil {
		fields["sort_order"] = *payload.SortOrder
	}
	if len(fields) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No fields to update",
		})
	}

	if err := dao.UpdateDepartment(payload.Id, fields); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return departmentNotFound(c)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func UpdateDepartmentOrder(c *fiber.Ctx) error {
	if checkDepartmentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	type OrderItem struct {
		Id        string `json:"id"`
		SortOrder int    `json:"sortOrder"`
	}
	type Payload struct {
		Orders []OrderItem `json:"orders"`
	}

	payload := new(Payload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if len(payload.Orders) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Orders are required",
		})
	}

	orders := make([]model.Department, len(payload.Orders))
	for i, item := range payload.Orders {
		orders[i] = model.Department{
			Id:        item.Id,
			SortOrder: item.SortOrder,
		}
	}

	if err := dao.UpdateDepartmentOrders(orders); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return departmentNotFound(c)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func DeleteDepartment(c *fiber.Ctx) error {
	if checkDepartmentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	if err := dao.DeleteDepartment(c.Params("id")); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return departmentNotFound(c)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func AddDepartmentMember(c *fiber.Ctx) error {
	if checkDepartmentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	departmentID := c.Params("id")
	type Payload struct {
		Username  string `json:"username"`
		SortOrder int    `json:"sortOrder"`
		IsLeader  bool   `json:"isLeader"`
	}
	payload := new(Payload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if strings.TrimSpace(payload.Username) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username is required",
		})
	}

	if _, err := dao.GetDepartmentByID(departmentID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Department not found",
		})
	}

	user, err := dao.GetUserByUsername(payload.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	exists, err := dao.DepartmentMemberExists(departmentID, payload.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "该成员已在部门中",
		})
	}

	member := model.DepartmentMember{
		DepartmentId: departmentID,
		Username:     payload.Username,
		SortOrder:    payload.SortOrder,
		IsLeader:     payload.IsLeader,
	}
	if err := dao.AddDepartmentMember(member); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func RemoveDepartmentMember(c *fiber.Ctx) error {
	if checkDepartmentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	username, err := url.PathUnescape(c.Params("username"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid username",
		})
	}

	if err := dao.RemoveDepartmentMember(c.Params("id"), username); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Department member not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func UpdateDepartmentMemberOrder(c *fiber.Ctx) error {
	if checkDepartmentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	departmentID := c.Params("id")
	type MemberOrder struct {
		Username  string `json:"username"`
		SortOrder int    `json:"sortOrder"`
		IsLeader  bool   `json:"isLeader"`
	}
	type Payload struct {
		Members []MemberOrder `json:"members"`
	}

	payload := new(Payload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if len(payload.Members) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Members are required",
		})
	}

	members := make([]model.DepartmentMember, len(payload.Members))
	for i, item := range payload.Members {
		members[i] = model.DepartmentMember{
			Username:  item.Username,
			SortOrder: item.SortOrder,
			IsLeader:  item.IsLeader,
		}
	}

	if err := dao.UpdateDepartmentMemberOrders(departmentID, members); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Department member not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func UpdateDepartmentMemberLeaderStatus(c *fiber.Ctx) error {
	if checkDepartmentPermission(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	departmentID := c.Params("id")
	username, err := url.PathUnescape(c.Params("username"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid username",
		})
	}

	type Payload struct {
		IsLeader bool `json:"isLeader"`
	}
	payload := new(Payload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := dao.UpdateDepartmentMemberLeader(departmentID, username, payload.IsLeader); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Department member not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}
