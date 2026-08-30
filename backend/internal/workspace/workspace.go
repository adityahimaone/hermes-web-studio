package workspace

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	MaxTextPreview = 1 << 20
	MaxUpload      = 10 << 20
)

var ErrInvalidPath = errors.New("workspace path is invalid")

type Entry struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Type       string    `json:"type"`
	Size       int64     `json:"size,omitempty"`
	ModifiedAt time.Time `json:"modified_at,omitempty"`
	MIME       string    `json:"mime,omitempty"`
}

type Preview struct {
	Entry
	Content  string `json:"content,omitempty"`
	Editable bool   `json:"editable"`
	Binary   bool   `json:"binary"`
}

type GitStatus struct {
	Available bool       `json:"available"`
	Root      string     `json:"root,omitempty"`
	Branch    string     `json:"branch,omitempty"`
	Entries   []GitEntry `json:"entries,omitempty"`
	Error     string     `json:"error,omitempty"`
}

type GitEntry struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

type Service struct {
	root     string
	rootReal string
}

func New(root string) (*Service, error) {
	if strings.TrimSpace(root) == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		root = filepath.Join(home, ".hermes", "webui", "workspace")
	}
	abs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0700); err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("workspace root is not a directory")
	}
	realRoot, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, err
	}
	return &Service{root: abs, rootReal: realRoot}, nil
}

func (s *Service) Root() string { return s.root }

func (s *Service) List(rel string) ([]Entry, error) {
	path, err := s.resolve(rel, false)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("workspace path is not a directory")
	}
	items, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, 0, len(items))
	for _, item := range items {
		itemPath := filepath.Join(path, item.Name())
		itemInfo, infoErr := os.Lstat(itemPath)
		if infoErr != nil {
			continue
		}
		typeName := "file"
		if itemInfo.IsDir() {
			typeName = "directory"
		} else if itemInfo.Mode()&os.ModeSymlink != 0 {
			typeName = "symlink"
		}
		entry := Entry{Name: item.Name(), Path: relative(s.root, itemPath), Type: typeName, Size: itemInfo.Size(), ModifiedAt: itemInfo.ModTime()}
		if typeName == "file" {
			entry.MIME = mime.TypeByExtension(filepath.Ext(item.Name()))
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Type != entries[j].Type {
			return entries[i].Type == "directory"
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	return entries, nil
}

func (s *Service) Preview(rel string) (Preview, error) {
	path, err := s.resolve(rel, false)
	if err != nil {
		return Preview{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return Preview{}, err
	}
	if info.IsDir() {
		return Preview{}, errors.New("workspace path is a directory")
	}
	entry := Entry{Name: filepath.Base(path), Path: relative(s.root, path), Type: "file", Size: info.Size(), ModifiedAt: info.ModTime(), MIME: mime.TypeByExtension(filepath.Ext(path))}
	if entry.MIME == "" {
		entry.MIME = "application/octet-stream"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Preview{}, err
	}
	if isText(entry.MIME, path, data) {
		if len(data) > MaxTextPreview {
			return Preview{}, errors.New("text preview is too large")
		}
		return Preview{Entry: entry, Content: string(data), Editable: true}, nil
	}
	return Preview{Entry: entry, Binary: true}, nil
}

func (s *Service) Read(rel string) ([]byte, Entry, error) {
	path, err := s.resolve(rel, false)
	if err != nil {
		return nil, Entry{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, Entry{}, err
	}
	if info.IsDir() {
		return nil, Entry{}, errors.New("workspace path is a directory")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, Entry{}, err
	}
	return data, Entry{Name: filepath.Base(path), Path: relative(s.root, path), Type: "file", Size: info.Size(), ModifiedAt: info.ModTime(), MIME: mime.TypeByExtension(filepath.Ext(path))}, nil
}

func (s *Service) Write(rel, content string) (Preview, error) {
	path, err := s.resolve(rel, true)
	if err != nil {
		return Preview{}, err
	}
	if len(content) > MaxTextPreview {
		return Preview{}, errors.New("text file is too large")
	}
	if info, statErr := os.Stat(path); statErr == nil && info.IsDir() {
		return Preview{}, errors.New("workspace path is a directory")
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return Preview{}, err
	}
	return s.Preview(rel)
}

func (s *Service) Create(rel, kind, content string) error {
	path, err := s.resolve(rel, true)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(path); err == nil {
		return os.ErrExist
	}
	if kind == "directory" {
		return os.Mkdir(path, 0700)
	}
	if kind != "file" {
		return errors.New("workspace item type is invalid")
	}
	if len(content) > MaxTextPreview {
		return errors.New("text file is too large")
	}
	return os.WriteFile(path, []byte(content), 0600)
}

func (s *Service) Rename(rel, name string) error {
	path, err := s.resolve(rel, false)
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\\`) {
		return ErrInvalidPath
	}
	destination := filepath.Join(filepath.Dir(path), name)
	if _, err := s.resolve(relative(s.root, destination), true); err != nil {
		return err
	}
	return os.Rename(path, destination)
}

func (s *Service) Delete(rel string) error {
	path, err := s.resolve(rel, false)
	if err != nil {
		return err
	}
	if filepath.Clean(path) == filepath.Clean(s.root) {
		return errors.New("workspace root cannot be deleted")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}

func (s *Service) SaveUpload(relDir, name string, reader io.Reader) (Entry, error) {
	dir, err := s.resolve(relDir, false)
	if err != nil {
		return Entry{}, err
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return Entry{}, errors.New("upload destination is not a directory")
	}
	name = filepath.Base(strings.TrimSpace(name))
	if name == "." || name == ".." || name == "" || strings.ContainsAny(name, `/\\`) {
		return Entry{}, ErrInvalidPath
	}
	destination := filepath.Join(dir, name)
	if _, err := s.resolve(relative(s.root, destination), true); err != nil {
		return Entry{}, err
	}
	data, err := io.ReadAll(io.LimitReader(reader, MaxUpload+1))
	if err != nil {
		return Entry{}, err
	}
	if len(data) > MaxUpload {
		return Entry{}, errors.New("upload is too large")
	}
	if err := os.WriteFile(destination, data, 0600); err != nil {
		return Entry{}, err
	}
	info, err := os.Stat(destination)
	if err != nil {
		return Entry{}, err
	}
	return Entry{Name: name, Path: relative(s.root, destination), Type: "file", Size: info.Size(), ModifiedAt: info.ModTime(), MIME: mime.TypeByExtension(filepath.Ext(name))}, nil
}

func (s *Service) Git(ctx context.Context, rel string) GitStatus {
	path, err := s.resolve(rel, false)
	if err != nil {
		return GitStatus{Error: "workspace path is invalid"}
	}
	rootCmd := exec.CommandContext(ctx, "git", "-C", path, "rev-parse", "--show-toplevel")
	rootOutput, err := rootCmd.Output()
	if err != nil {
		return GitStatus{}
	}
	gitRoot := strings.TrimSpace(string(rootOutput))
	status := GitStatus{Available: true, Root: gitRoot}
	branchCmd := exec.CommandContext(ctx, "git", "-C", gitRoot, "branch", "--show-current")
	if branchOutput, branchErr := branchCmd.Output(); branchErr == nil {
		status.Branch = strings.TrimSpace(string(branchOutput))
	}
	statusCmd := exec.CommandContext(ctx, "git", "-C", gitRoot, "status", "--porcelain")
	if output, statusErr := statusCmd.Output(); statusErr == nil {
		for _, line := range strings.Split(strings.TrimRight(string(output), "\n"), "\n") {
			if len(line) < 4 {
				continue
			}
			status.Entries = append(status.Entries, GitEntry{Status: strings.TrimSpace(line[:2]), Path: strings.TrimSpace(line[3:])})
		}
	}
	return status
}

func (s *Service) resolve(rel string, allowMissing bool) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		rel = "."
	}
	if filepath.IsAbs(rel) || strings.Contains(rel, `\`) {
		return "", ErrInvalidPath
	}
	clean := filepath.Clean(filepath.FromSlash(rel))
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", ErrInvalidPath
	}
	candidate := filepath.Join(s.root, clean)
	if !isWithin(s.root, candidate) {
		return "", ErrInvalidPath
	}
	check := candidate
	if allowMissing {
		for {
			if _, err := os.Lstat(check); err == nil {
				break
			} else if !errors.Is(err, os.ErrNotExist) {
				return "", err
			}
			parent := filepath.Dir(check)
			if parent == check {
				return "", ErrInvalidPath
			}
			check = parent
		}
	}
	real, err := filepath.EvalSymlinks(check)
	if err != nil {
		if allowMissing && errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}
		return "", err
	}
	if !isWithin(s.rootReal, real) {
		return "", ErrInvalidPath
	}
	if allowMissing && real != check {
		return filepath.Join(real, strings.TrimPrefix(candidate, check)), nil
	}
	return candidate, nil
}

func isWithin(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != string(filepath.Separator)
}

func relative(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return "."
	}
	return filepath.ToSlash(rel)
}

func isText(contentType, path string, data []byte) bool {
	if strings.HasPrefix(contentType, "text/") || contentType == "application/json" || contentType == "application/javascript" || contentType == "application/xml" {
		return true
	}
	detected := http.DetectContentType(data)
	return strings.HasPrefix(detected, "text/") || filepath.Ext(path) == ".md" || filepath.Ext(path) == ".go" || filepath.Ext(path) == ".tsx" || filepath.Ext(path) == ".ts" || filepath.Ext(path) == ".css"
}
