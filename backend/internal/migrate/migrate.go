package migrate

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func Backup(stateDir string) (string, error) {
	if stateDir == "" {
		return "", errors.New("state directory required")
	}
	target := filepath.Join(stateDir, ".backups", time.Now().UTC().Format("20060102T150405Z"))
	if err := copyTree(stateDir, target, filepath.Join(stateDir, ".backups")); err != nil {
		return "", err
	}
	return target, nil
}
func RestoreLatest(stateDir string) (string, error) {
	entries, err := os.ReadDir(filepath.Join(stateDir, ".backups"))
	if err != nil {
		return "", err
	}
	names := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	if len(names) == 0 {
		return "", errors.New("no backup found")
	}
	sort.Strings(names)
	source := filepath.Join(stateDir, ".backups", names[len(names)-1])
	if err := copyTree(source, stateDir, ""); err != nil {
		return "", err
	}
	return source, nil
}
func copyTree(source, destination, skip string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if skip != "" && (path == skip || filepath.HasPrefix(path, skip+string(filepath.Separator))) {
			return filepath.SkipDir
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0700)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0600)
	})
}
