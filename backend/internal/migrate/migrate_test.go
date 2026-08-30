package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupAndRestoreLatest(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "legacy.json"), []byte("before"), 0600); err != nil {
		t.Fatal(err)
	}
	backup, err := Backup(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(backup); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "legacy.json"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreLatest(root)
	if err != nil || restored != backup {
		t.Fatalf("restored=%q err=%v", restored, err)
	}
	data, err := os.ReadFile(filepath.Join(root, "legacy.json"))
	if err != nil || string(data) != "before" {
		t.Fatalf("data=%q err=%v", data, err)
	}
}
