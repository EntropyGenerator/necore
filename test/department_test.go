package necore_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestDepartmentRoutes(t *testing.T) {
	env := setupTestEnv(t)

	assertStatus(t, doJSON(t, env, http.MethodGet, "/necore/department/", "", nil), http.StatusOK)

	createResp := doJSON(t, env, http.MethodPost, "/necore/department/create", env.adminToken, fiber.Map{
		"name":        "运维保障部",
		"description": "负责服务器与网站稳定运行",
		"icon":        "/contents/dept/icon.png",
		"sortOrder":   1,
	})
	assertStatus(t, createResp, http.StatusOK)
	createBody := decodeBody(t, createResp)
	deptID, _ := createBody["id"].(string)
	if deptID == "" {
		t.Fatalf("create department should return id, got %#v", createBody)
	}

	assertStatus(t, doJSON(t, env, http.MethodPost, "/necore/department/"+deptID+"/member", env.adminToken, fiber.Map{
		"username":  "alice",
		"sortOrder": 1,
		"isLeader":  true,
	}), http.StatusOK)

	listResp := doJSON(t, env, http.MethodGet, "/necore/department/", "", nil)
	assertStatus(t, listResp, http.StatusOK)
	listBody := decodeBody(t, listResp)
	departments, ok := listBody["departments"].([]any)
	if !ok || len(departments) != 1 {
		t.Fatalf("department list = %#v", listBody["departments"])
	}
	dept := departments[0].(map[string]any)
	members, ok := dept["members"].([]any)
	if !ok || len(members) != 1 {
		t.Fatalf("members = %#v", dept["members"])
	}
	member := members[0].(map[string]any)
	if member["isLeader"] != true {
		t.Fatalf("expected isLeader true after add, got %#v (%T)", member["isLeader"], member["isLeader"])
	}

	assertStatus(t, doJSON(t, env, http.MethodPatch, "/necore/department/"+deptID+"/member/alice/leader", env.adminToken, fiber.Map{
		"isLeader": false,
	}), http.StatusOK)

	listAfterToggleResp := doJSON(t, env, http.MethodGet, "/necore/department/", "", nil)
	assertStatus(t, listAfterToggleResp, http.StatusOK)
	listAfterToggleBody := decodeBody(t, listAfterToggleResp)
	departmentsAfterToggle, ok := listAfterToggleBody["departments"].([]any)
	if !ok || len(departmentsAfterToggle) != 1 {
		t.Fatalf("department list after toggle = %#v", listAfterToggleBody["departments"])
	}
	deptAfterToggle := departmentsAfterToggle[0].(map[string]any)
	membersAfterToggle, ok := deptAfterToggle["members"].([]any)
	if !ok || len(membersAfterToggle) != 1 {
		t.Fatalf("members after toggle = %#v", deptAfterToggle["members"])
	}
	memberAfterToggle := membersAfterToggle[0].(map[string]any)
	if memberAfterToggle["isLeader"] != false {
		t.Fatalf("expected isLeader false after toggle, got %#v (%T)", memberAfterToggle["isLeader"], memberAfterToggle["isLeader"])
	}

	assertStatus(t, doJSON(t, env, http.MethodPatch, "/necore/department/order", env.adminToken, fiber.Map{
		"orders": []fiber.Map{
			{"id": deptID, "sortOrder": 2},
		},
	}), http.StatusOK)

	assertStatus(t, doJSON(t, env, http.MethodDelete, "/necore/department/"+deptID+"/member/alice", env.adminToken, nil), http.StatusOK)
	assertStatus(t, doJSON(t, env, http.MethodDelete, "/necore/department/"+deptID, env.adminToken, nil), http.StatusOK)
}

func TestDepartment_PermissionsAndNotFound(t *testing.T) {
	env := setupTestEnv(t)

	assertStatus(t, doJSON(t, env, http.MethodGet, "/necore/department/", "", nil), http.StatusOK)

	cases := []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodPost, "/necore/department/create", fiber.Map{"name": "x"}},
		{http.MethodPatch, "/necore/department/", fiber.Map{"id": "x", "name": "y"}},
		{http.MethodPatch, "/necore/department/order", fiber.Map{"orders": []fiber.Map{{"id": "x", "sortOrder": 1}}}},
		{http.MethodDelete, "/necore/department/x", nil},
		{http.MethodPost, "/necore/department/x/member", fiber.Map{"username": "alice"}},
		{http.MethodDelete, "/necore/department/x/member/alice", nil},
		{http.MethodPatch, "/necore/department/x/member/alice/leader", fiber.Map{"isLeader": true}},
		{http.MethodPatch, "/necore/department/x/member/order", fiber.Map{"members": []fiber.Map{{"username": "alice", "sortOrder": 1}}}},
	}
	for _, tc := range cases {
		assertStatus(t, doJSON(t, env, tc.method, tc.path, env.userToken, tc.body), http.StatusForbidden)
	}

	assertStatus(t, doJSON(t, env, http.MethodPost, "/necore/department/create", "", fiber.Map{"name": "x"}), http.StatusUnauthorized)

	assertStatus(t, doJSON(t, env, http.MethodDelete, "/necore/department/missing-id", env.adminToken, nil), http.StatusNotFound)
	assertStatus(t, doJSON(t, env, http.MethodPost, "/necore/department/create", env.adminToken, fiber.Map{"name": "  "}), http.StatusBadRequest)
}

func TestDepartment_MemberDuplicateAndGroupOmission(t *testing.T) {
	env := setupTestEnv(t)

	createResp := doJSON(t, env, http.MethodPost, "/necore/department/create", env.adminToken, fiber.Map{"name": "D"})
	deptID, _ := decodeBody(t, createResp)["id"].(string)

	assertStatus(t, doJSON(t, env, http.MethodPost, "/necore/department/"+deptID+"/member", env.adminToken, fiber.Map{
		"username": "alice",
	}), http.StatusOK)

	assertStatus(t, doJSON(t, env, http.MethodPost, "/necore/department/"+deptID+"/member", env.adminToken, fiber.Map{
		"username": "alice",
	}), http.StatusConflict)

	listResp := doJSON(t, env, http.MethodGet, "/necore/department/", "", nil)
	if strings.Contains(string(listResp.Body), "\"group\"") {
		t.Fatalf("public department list must not expose group, body=%s", string(listResp.Body))
	}
}
