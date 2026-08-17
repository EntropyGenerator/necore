package necore_test

import (
	"os"
	"path/filepath"
	"testing"

	"necore/dao"
	"necore/database"
	"necore/model"

	"gorm.io/gorm"
)

func TestDao_DocumentNodeCreationValidation(t *testing.T) {
	setupTestEnv(t)

	if err := dao.CreateDocumentNode("", false, false, "n", "id-1", "admin"); err == nil {
		t.Fatal("empty parentId should be rejected")
	}

	if err := dao.CreateDocumentNode("id-1", false, false, "n", "id-1", "admin"); err == nil {
		t.Fatal("self-parent should be rejected")
	}

	if err := dao.CreateDocumentNode("missing", false, false, "n", "id-2", "admin"); err == nil {
		t.Fatal("missing parent should be rejected")
	}

	if err := dao.CreateDocumentNode("root", true, false, "Folder", "folder-1", "admin"); err != nil {
		t.Fatal(err)
	}
	if err := dao.CreateDocumentNode("folder-1", false, false, "Child", "child-1", "admin"); err != nil {
		t.Fatal(err)
	}

	if err := dao.CreateDocumentNode("child-1", false, false, "x", "id-3", "admin"); err == nil {
		t.Fatal("non-folder parent should be rejected")
	}

}

func TestDao_DocumentNodeCyclePrevention(t *testing.T) {
	setupTestEnv(t)

	if err := dao.CreateDocumentNode("root", true, false, "A", "cyc-a", "admin"); err != nil {
		t.Fatal(err)
	}
	if err := dao.CreateDocumentNode("cyc-a", true, false, "B", "cyc-b", "admin"); err != nil {
		t.Fatal(err)
	}

	if err := dao.UpdateDocumentNodeParentId("cyc-a", "cyc-b"); err == nil {
		t.Fatal("circular parent move should be rejected")
	}

	if err := dao.UpdateDocumentNodeParentId("cyc-b", "root"); err != nil {
		t.Fatalf("valid move rejected: %v", err)
	}
}

func TestDao_DocumentNodeDeleteRecursive(t *testing.T) {
	env := setupTestEnv(t)

	if err := dao.CreateDocumentNode("root", true, false, "Folder", "del-folder", "admin"); err != nil {
		t.Fatal(err)
	}
	if err := dao.CreateDocumentNode("del-folder", false, false, "C1", "del-c1", "admin"); err != nil {
		t.Fatal(err)
	}
	if err := dao.CreateDocumentNode("del-folder", false, false, "C2", "del-c2", "admin"); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"del-folder", "del-c1", "del-c2"} {
		if err := os.MkdirAll(filepath.Join(env.tmpDir, "contents", id), 0o750); err != nil {
			t.Fatal(err)
		}
	}

	if err := dao.DeleteDocumentNode("del-folder"); err != nil {
		t.Fatal(err)
	}

	var count int64
	if err := database.GetDocumentDatabase().Model(&model.DocumentNode{}).
		Where("id IN ?", []string{"del-folder", "del-c1", "del-c2"}).
		Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("recursive delete left %d nodes", count)
	}
	for _, id := range []string{"del-folder", "del-c1", "del-c2"} {
		if _, err := os.Stat(filepath.Join(env.tmpDir, "contents", id)); !os.IsNotExist(err) {
			t.Fatalf("contents dir %s should be removed, err=%v", id, err)
		}
	}

	if err := dao.DeleteDocumentNode("del-folder"); err != gorm.ErrRecordNotFound {
		t.Fatalf("delete missing node err = %v, want ErrRecordNotFound", err)
	}
}

func TestDao_ArticleListPagination(t *testing.T) {
	setupTestEnv(t)

	if err := dao.CreateArticle("art-1"); err != nil {
		t.Fatal(err)
	}
	if err := dao.CreateArticle("art-2"); err != nil {
		t.Fatal(err)
	}
	if err := dao.CreateArticle("art-3"); err != nil {
		t.Fatal(err)
	}

	upd := func(id, category string, pin bool, date string) {
		t.Helper()
		if err := dao.UpdateArticle(model.Article{
			Id:       id,
			Category: category,
			Pin:      pin,
			Date:     date,
		}); err != nil {
			t.Fatal(err)
		}
	}
	upd("art-1", "notice", true, "2026-01-03")
	upd("art-2", "notice", false, "2026-01-02")
	upd("art-3", "notice", false, "2026-01-01")

	all, err := dao.GetArticleList("notice", 1, 10, false)
	if err != nil || len(all) != 3 {
		t.Fatalf("all list len=%d err=%v", len(all), err)
	}
	if all[0].Id != "art-1" {
		t.Fatalf("date desc order broken: first=%s", all[0].Id)
	}

	pinned, err := dao.GetArticleList("notice", 1, 10, true)
	if err != nil || len(pinned) != 1 || pinned[0].Id != "art-1" {
		t.Fatalf("pin list len=%d first=%s err=%v", len(pinned), firstID(pinned), err)
	}

	page2, err := dao.GetArticleList("notice", 2, 2, false)
	if err != nil || len(page2) != 1 || page2[0].Id != "art-3" {
		t.Fatalf("page2 len=%d first=%s err=%v", len(page2), firstID(page2), err)
	}
}

func firstID(list []model.Article) string {
	if len(list) == 0 {
		return ""
	}
	return list[0].Id
}

func TestDao_BotTokenHashRoundTrip(t *testing.T) {
	setupTestEnv(t)

	record, plain, err := dao.CreateBotToken("unit-bot")
	if err != nil {
		t.Fatal(err)
	}
	if plain == "" || plain == record.TokenHash {
		t.Fatal("plain token must differ from stored hash")
	}

	byPlain, err := dao.GetBotTokenByPlainToken(plain)
	if err != nil || byPlain.Name != "unit-bot" {
		t.Fatalf("plain lookup failed: %v", err)
	}
	if _, err := dao.GetBotTokenByPlainToken(record.TokenHash); err == nil {
		t.Fatal("hash must not be usable as a plain token")
	}

	if _, _, err := dao.CreateBotToken("unit-bot"); err != dao.ErrBotTokenAlreadyExists {
		t.Fatalf("duplicate name err = %v", err)
	}
}

func TestDao_UserPasswordRoundTrip(t *testing.T) {
	setupTestEnv(t)

	if err := dao.AddUserByUsername("pw-user", "s3cret"); err != nil {
		t.Fatal(err)
	}
	user, err := dao.GetUserByUsername("pw-user")
	if err != nil {
		t.Fatal(err)
	}
	if !dao.CheckUserPassword("s3cret", user.Password) {
		t.Fatal("correct password rejected")
	}
	if dao.CheckUserPassword("wrong", user.Password) {
		t.Fatal("wrong password accepted")
	}
	if user.Password == "s3cret" {
		t.Fatal("password stored in plaintext")
	}

	if !dao.ContainsGroup(`["admin","news_admin"]`, "news_admin") {
		t.Fatal("ContainsGroup should match present group")
	}
	if dao.ContainsGroup(`["admin"]`, "bot_admin") {
		t.Fatal("ContainsGroup matched absent group")
	}
}
