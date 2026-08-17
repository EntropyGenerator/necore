package necore_test

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestCoord_LowercaseTagsField(t *testing.T) {
	env := setupTestEnv(t)

	resp := doJSON(t, env, http.MethodPatch, "/necore/auth/user", env.adminToken, fiber.Map{
		"username": "alice",
		"group":    []string{},
		"tags": []fiber.Map{
			{"text": "成员", "color": "#ffffff", "tagColor": "#409EFF"},
		},
	})
	assertStatus(t, resp, http.StatusOK)

	info := doJSON(t, env, http.MethodGet, "/necore/auth/user/alice", "", nil)
	t.Logf("alice info: %s", string(info.Body))
}
