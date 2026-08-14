package necore_test

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"necore/dao"
	"necore/database"
)

// setupEmptyUserEnv prepares an isolated working directory with an empty user
// database (all other databases are also migrated so ConnectSqlite does not
// fail) and chdir into it. Cleanup restores the previous working directory and
// closes all opened SQLite connections so the temp dir can be removed.
func setupEmptyUserEnv(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	must(t, os.MkdirAll(filepath.Join(tmpDir, "data"), 0o755))
	must(t, os.MkdirAll(filepath.Join(tmpDir, "contents"), 0o755))

	envContent := "SECRET=unit-test-secret\nBOT_LOG_BUFFER_SIZE=100\n"
	must(t, os.WriteFile(filepath.Join(tmpDir, ".env"), []byte(envContent), 0o600))

	oldWd, err := os.Getwd()
	must(t, err)
	must(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	t.Setenv("SECRET", "unit-test-secret")
	t.Setenv("BOT_LOG_BUFFER_SIZE", "100")

	database.ConnectSqlite()

	setGormLoggerSilent(database.GetUserDatabase())
	setGormLoggerSilent(database.GetArticleDatabase())
	setGormLoggerSilent(database.GetServerDatabase())
	setGormLoggerSilent(database.GetDocumentDatabase())
	setGormLoggerSilent(database.GetBotTokenDatabase())

	t.Cleanup(func() {
		closeGormDB(t, database.GetUserDatabase())
		closeGormDB(t, database.GetArticleDatabase())
		closeGormDB(t, database.GetServerDatabase())
		closeGormDB(t, database.GetDocumentDatabase())
		closeGormDB(t, database.GetBotTokenDatabase())
		closeGormDB(t, database.GetWikiDatabase())
	})

	return tmpDir
}

// captureLogOutput redirects the standard logger output to a buffer for the
// duration of fn. Go test functions run sequentially by default (none of the
// tests in this file call t.Parallel), so the global swap is safe.
func captureLogOutput(t *testing.T, fn func()) string {
	t.Helper()

	var buf bytes.Buffer
	oldOutput := log.Writer()
	oldFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(oldOutput)
		log.SetFlags(oldFlags)
	}()

	fn()
	return buf.String()
}

// TestEnsureInitialAdmin_CreatesAdminWhenEmpty verifies that on a freshly
// initialized user database, EnsureInitialAdmin creates exactly one "admin"
// user with the admin group, prints the credentials once, and is a no-op on
// subsequent invocations.
func TestEnsureInitialAdmin_CreatesAdminWhenEmpty(t *testing.T) {
	setupEmptyUserEnv(t)

	count, err := dao.GetUserCount()
	must(t, err)
	if count != 0 {
		t.Fatalf("expected empty user database, got count=%d", count)
	}

	output := captureLogOutput(t, func() {
		must(t, dao.EnsureInitialAdmin())
	})

	if !strings.Contains(output, "admin") {
		t.Fatalf("expected log output to mention admin, got %q", output)
	}
	if !strings.Contains(output, "password") {
		t.Fatalf("expected log output to mention password, got %q", output)
	}

	// Running again should be a no-op: no additional user, no log output.
	output2 := captureLogOutput(t, func() {
		must(t, dao.EnsureInitialAdmin())
	})
	if output2 != "" {
		t.Fatalf("expected no log output on repeated call, got %q", output2)
	}

	users, err := dao.GetAllUsers()
	must(t, err)
	if len(users) != 1 {
		t.Fatalf("expected exactly 1 user, got %d", len(users))
	}
	if users[0].Username != "admin" {
		t.Fatalf("expected admin user, got %q", users[0].Username)
	}
	if !dao.ContainsGroup(users[0].Group, "admin") {
		t.Fatalf("expected admin group, got %q", users[0].Group)
	}
	if users[0].Password == "" {
		t.Fatalf("admin password hash is empty")
	}
}

// TestEnsureInitialAdmin_NoOpWhenPopulated verifies EnsureInitialAdmin does
// nothing when at least one active user already exists.
func TestEnsureInitialAdmin_NoOpWhenPopulated(t *testing.T) {
	setupEmptyUserEnv(t)

	must(t, dao.AddUserByUsername("someone", "some-pass"))

	output := captureLogOutput(t, func() {
		must(t, dao.EnsureInitialAdmin())
	})

	if output != "" {
		t.Fatalf("expected no log output when users exist, got %q", output)
	}

	users, err := dao.GetAllUsers()
	must(t, err)
	if len(users) != 1 {
		t.Fatalf("expected exactly 1 user, got %d", len(users))
	}
	if users[0].Username != "someone" {
		t.Fatalf("expected only the pre-existing user, got %q", users[0].Username)
	}
}

// TestEnsureInitialAdmin_RevivesSoftDeletedAdmin documents that when an admin
// was previously soft-deleted (active count is 0), EnsureInitialAdmin falls
// into the creation path, where AddAdminUser revives the row, increments
// token_version, and restores the admin group.
func TestEnsureInitialAdmin_RevivesSoftDeletedAdmin(t *testing.T) {
	setupEmptyUserEnv(t)

	must(t, dao.AddAdminUser("admin", "first-pass"))
	must(t, dao.DeleteUserByUsername("admin"))

	output := captureLogOutput(t, func() {
		must(t, dao.EnsureInitialAdmin())
	})

	if !strings.Contains(output, "admin") {
		t.Fatalf("expected log output to mention admin, got %q", output)
	}

	users, err := dao.GetAllUsers()
	must(t, err)
	if len(users) != 1 {
		t.Fatalf("expected exactly 1 active user, got %d", len(users))
	}
	if users[0].Username != "admin" {
		t.Fatalf("expected admin user, got %q", users[0].Username)
	}
	if !dao.ContainsGroup(users[0].Group, "admin") {
		t.Fatalf("expected admin group, got %q", users[0].Group)
	}
	if users[0].TokenVersion != 2 {
		t.Fatalf("expected token_version=2 after revival, got %d", users[0].TokenVersion)
	}
}

// TestAddAdminUser_DuplicateActive verifies AddAdminUser rejects creating an
// admin that already exists as an active user.
func TestAddAdminUser_DuplicateActive(t *testing.T) {
	setupEmptyUserEnv(t)

	must(t, dao.AddAdminUser("admin", "first-pass"))
	if err := dao.AddAdminUser("admin", "second-pass"); err == nil {
		t.Fatalf("expected error creating duplicate admin, got nil")
	}
}

// TestAddAdminUser_EmptyInputs verifies basic input validation.
func TestAddAdminUser_EmptyInputs(t *testing.T) {
	setupEmptyUserEnv(t)

	if err := dao.AddAdminUser("", "pass"); err == nil {
		t.Fatalf("expected error for empty username, got nil")
	}
	if err := dao.AddAdminUser("admin", ""); err == nil {
		t.Fatalf("expected error for empty password, got nil")
	}
}

// TestGetUserCount_ExcludesSoftDeleted verifies Count reflects only active
// (non-soft-deleted) users, which is what EnsureInitialAdmin relies on.
func TestGetUserCount_ExcludesSoftDeleted(t *testing.T) {
	setupEmptyUserEnv(t)

	if count, err := dao.GetUserCount(); err != nil || count != 0 {
		t.Fatalf("expected 0 users, got %d (err=%v)", count, err)
	}

	must(t, dao.AddUserByUsername("a", "p1"))
	if count, err := dao.GetUserCount(); err != nil || count != 1 {
		t.Fatalf("expected 1 user, got %d (err=%v)", count, err)
	}

	must(t, dao.AddUserByUsername("b", "p2"))
	if count, err := dao.GetUserCount(); err != nil || count != 2 {
		t.Fatalf("expected 2 users, got %d (err=%v)", count, err)
	}

	must(t, dao.DeleteUserByUsername("a"))
	if count, err := dao.GetUserCount(); err != nil || count != 1 {
		t.Fatalf("expected 1 user after soft delete, got %d (err=%v)", count, err)
	}
}
