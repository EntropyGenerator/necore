package necore_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"necore/dao"

	"github.com/gofiber/fiber/v2"
)

// 回归测试：private 文档不可被匿名或越权读取。
//  1. private 文件夹下的 public 子节点：祖先链为 private 时公开路由不可见（404）。
//  2. private 节点的上传文件：匿名 401，非 document_admin 403。
//     成功路径（document_admin 200）不通过 HTTP 断言——fasthttp 在 Windows 上
//     会让 SendFile 的文件句柄存活到上下文池回收，导致 TempDir 清理失败
//     （既有已知限制，见 TestDocumentRoutes 的注释）；改用 DAO 守卫函数直接验证。
func TestSecurityRegression_PrivateDocumentsAreNotPubliclyReadable(t *testing.T) {
	env := setupTestEnv(t)

	// private 文件夹 + 其下 public 子节点
	folderResp := doJSON(t, env, http.MethodPost, "/necore/documents/node", env.adminToken, fiber.Map{
		"parentId": "root", "isFolder": true, "private": true, "name": "Secret Folder",
	})
	assertStatus(t, folderResp, http.StatusOK)
	folderID, _ := decodeBody(t, folderResp)["id"].(string)

	childResp := doJSON(t, env, http.MethodPost, "/necore/documents/node", env.adminToken, fiber.Map{
		"parentId": folderID, "isFolder": false, "private": false, "name": "Child Doc",
	})
	assertStatus(t, childResp, http.StatusOK)
	childID, _ := decodeBody(t, childResp)["id"].(string)

	// 1. 匿名读取 private 文件夹下的子节点 → 404
	assertStatus(t, doJSON(t, env, http.MethodGet, "/necore/documents/"+childID, "", nil), http.StatusNotFound)
	// 1b. 匿名枚举 private 文件夹的子节点 → 空列表
	layerResp := doJSON(t, env, http.MethodGet, "/necore/documents/layer/"+folderID, "", nil)
	assertStatus(t, layerResp, http.StatusOK)
	if strings.Contains(string(layerResp.Body), "Child Doc") {
		t.Fatalf("private folder children leaked in public layer")
	}

	// 2. private 节点上传文件
	docResp := doJSON(t, env, http.MethodPost, "/necore/documents/node", env.adminToken, fiber.Map{
		"parentId": "root", "isFolder": false, "private": true, "name": "Secret PDF Doc",
	})
	assertStatus(t, docResp, http.StatusOK)
	docID, _ := decodeBody(t, docResp)["id"].(string)

	upResp := doMultipartFile(t, env, "/necore/documents/upload/"+docID, env.adminToken, "file", "secret.pdf", "PDF-CONTENT-SECRET")
	assertStatus(t, upResp, http.StatusOK)
	filename := strings.Split(decodeBody(t, upResp)["url"].(string), "/")[3]

	p := filepath.Join(env.tmpDir, "contents", docID, filename)
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("uploaded file missing: %v", err)
	}

	// 2a. 匿名访问 private 文件 → 401
	assertStatus(t, doJSON(t, env, http.MethodGet, "/necore/contents/"+docID+"/"+filename, "", nil), http.StatusUnauthorized)
	// 2b. 普通登录用户 → 403
	assertStatus(t, doJSON(t, env, http.MethodGet, "/necore/contents/"+docID+"/"+filename, env.userToken, nil), http.StatusForbidden)

	// 2c. 守卫决策：private 节点（含祖先）为 true，public 节点与非文档 ID 为 false
	if v, err := dao.IsDocumentNodeEffectivelyPrivate(docID); err != nil || !v {
		t.Fatalf("private node should be effectively private, v=%v err=%v", v, err)
	}
	if v, err := dao.IsDocumentNodeEffectivelyPrivate(childID); err != nil || !v {
		t.Fatalf("child under private folder should be effectively private, v=%v err=%v", v, err)
	}

	// 2d. public 节点的文件仍然可匿名访问（不误伤正常资源）
	pubResp := doJSON(t, env, http.MethodPost, "/necore/documents/node", env.adminToken, fiber.Map{
		"parentId": "root", "isFolder": false, "private": false, "name": "Public Doc",
	})
	assertStatus(t, pubResp, http.StatusOK)
	pubID, _ := decodeBody(t, pubResp)["id"].(string)
	if v, err := dao.IsDocumentNodeEffectivelyPrivate(pubID); err != nil || v {
		t.Fatalf("public node should not be effectively private, v=%v err=%v", v, err)
	}
	// 非文档 ID（文章/服务器等）不触发守卫
	if v, err := dao.IsDocumentNodeEffectivelyPrivate("not-a-document-id"); err != nil || v {
		t.Fatalf("non-document id should be public, v=%v err=%v", v, err)
	}
}
