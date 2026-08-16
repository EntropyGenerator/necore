package necore_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"necore/util"
)

func TestUtil_SafeFilename(t *testing.T) {
	valid := []string{"a.png", "file name.txt", "ümlaut.pdf", "a.b.c", "UPPER.JPG"}
	for _, in := range valid {
		out, err := util.SafeFilename(in)
		if err != nil || out != in {
			t.Fatalf("SafeFilename(%q) = %q, %v; want %q", in, out, err, in)
		}
	}

	invalid := []string{
		"", "   ", "..", ".", "../x", "a/../b", "a\\b", "/etc/passwd",
		"a/b", "x\x00y",
	}
	for _, in := range invalid {
		if out, err := util.SafeFilename(in); err == nil {
			t.Fatalf("SafeFilename(%q) = %q, nil; want error", in, out)
		}
	}
}

func TestUtil_SafeContentPath(t *testing.T) {
	// 合法：目标位于 root/objectID 下
	p, err := util.SafeContentPath("contents", "obj", "a.png")
	if err != nil {
		t.Fatalf("valid path rejected: %v", err)
	}
	abs, _ := filepath.Abs(filepath.Join("contents", "obj", "a.png"))
	if p != abs {
		t.Fatalf("path = %q, want %q", p, abs)
	}

	// 文件名穿越
	for _, bad := range []string{"../x", "a/../../x", "..\\x", "/abs/path", ".."} {
		if _, err := util.SafeContentPath("contents", "obj", bad); err == nil {
			t.Fatalf("traversal filename %q should be rejected", bad)
		}
	}

	// objectID 逃逸（".." 会把目录推到 root 之外）
	rejectedIDs := []string{"..", "../x"}
	// 反斜杠仅在 Windows 上是路径分隔符，Linux 上是普通文件名，不能作为逃逸用例
	if runtime.GOOS == "windows" {
		rejectedIDs = append(rejectedIDs, "..\\..")
	}
	for _, badID := range rejectedIDs {
		if _, err := util.SafeContentPath("contents", badID, "x.png"); err == nil {
			t.Fatalf("escaping objectID %q should be rejected", badID)
		}
	}

	// "a/.." 规约后仍落在 root 内（等价于 "."），不构成逃逸
	if _, err := util.SafeContentPath("contents", "a/..", "x.png"); err != nil {
		t.Fatalf("objectID %q should normalize into root: %v", "a/..", err)
	}

	// root 本身作为 objectID（等价于直接落在 root）允许
	if _, err := util.SafeContentPath("contents", ".", "x.png"); err != nil {
		t.Fatalf("objectID \".\" should be allowed: %v", err)
	}
}

func TestUtil_GenerateSecureToken(t *testing.T) {
	prefix := "bot"
	a, err := util.GenerateSecureToken(prefix, 32)
	if err != nil {
		t.Fatal(err)
	}
	b, err := util.GenerateSecureToken(prefix, 32)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(a, prefix+"_") || !strings.HasPrefix(b, prefix+"_") {
		t.Fatalf("token should carry prefix: %q / %q", a, b)
	}
	if len(a) != len(prefix)+1+43 { // 32 bytes -> base64url 43 chars
		t.Fatalf("unexpected token length %d", len(a))
	}
	if a == b {
		t.Fatal("two generated tokens must differ")
	}
}
