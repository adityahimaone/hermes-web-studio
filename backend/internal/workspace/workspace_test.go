package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceOperationsStayWithinResolvedRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "src"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\n"), 0600); err != nil {
		t.Fatal(err)
	}
	service, err := New(root)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := service.List(".")
	if err != nil || len(entries) != 1 || entries[0].Path != "src" || entries[0].Type != "directory" {
		t.Fatalf("entries=%#v err=%v", entries, err)
	}
	preview, err := service.Preview("src/main.go")
	if err != nil || preview.Content != "package main\n" || !preview.Editable {
		t.Fatalf("preview=%#v err=%v", preview, err)
	}

	if _, err := service.Preview("../outside"); !errorsIsInvalid(err) {
		t.Fatalf("traversal error=%v", err)
	}
	if _, err := service.Preview(filepath.Join(root, "src", "main.go")); !errorsIsInvalid(err) {
		t.Fatalf("absolute path error=%v", err)
	}
	if _, err := service.Preview(`src\main.go`); !errorsIsInvalid(err) {
		t.Fatalf("backslash path error=%v", err)
	}

	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "outside-link")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := service.Preview("outside-link"); !errorsIsInvalid(err) {
		t.Fatalf("external symlink error=%v", err)
	}
}

func TestWorkspaceMutationsAndUpload(t *testing.T) {
	service, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Create("notes", "directory", ""); err != nil {
		t.Fatal(err)
	}
	if err := service.Create("notes/draft.md", "file", "draft"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Write("notes/draft.md", "updated"); err != nil {
		t.Fatal(err)
	}
	if err := service.Rename("notes/draft.md", "final.md"); err != nil {
		t.Fatal(err)
	}
	entry, err := service.SaveUpload("notes", "upload.txt", strings.NewReader("uploaded"))
	if err != nil || entry.Path != "notes/upload.txt" {
		t.Fatalf("entry=%#v err=%v", entry, err)
	}
	info, err := os.Stat(filepath.Join(service.Root(), "notes", "upload.txt"))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("upload mode=%v err=%v", info.Mode().Perm(), err)
	}
	if err := service.Delete("notes/final.md"); err != nil {
		t.Fatal(err)
	}
	if err := service.Delete("."); err == nil {
		t.Fatal("root deletion unexpectedly succeeded")
	}
}

func errorsIsInvalid(err error) bool {
	return err == ErrInvalidPath
}
