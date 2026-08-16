package necore_test

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// 前端发送小写 tags，后端字段标签为大写 Tags：
// 验证 encoding/json 的大小写不敏感匹配是否真的生效（API.md 曾警告必须发送大写）。
func TestCoord_LowercaseTagsField(t *testing.T) {
	env := setupTestEnv(t)

	// admin 用小写 tags 更新 alice 的标签
	resp := doJSON(t, env, http.MethodPatch, "/necore/auth/user", env.adminToken, fiber.Map{
		"username": "alice",
		"group":    []string{},
		"tags": []fiber.Map{
			{"text": "成员", "color": "#ffffff", "tagColor": "#409EFF"},
		},
	})
	assertStatus(t, resp, http.StatusOK)

	// 读取 alice 的公开信息确认标签是否保存
	info := doJSON(t, env, http.MethodGet, "/necore/auth/user/alice", "", nil)
	t.Logf("alice info: %s", string(info.Body))
}
