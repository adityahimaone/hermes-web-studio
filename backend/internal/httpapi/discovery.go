package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	root := filepath.Join(s.config.HermesHome, "skills")
	if name := strings.TrimSpace(r.URL.Query().Get("name")); name != "" {
		content, err := readSkill(root, name)
		if err != nil {
			writeError(w, http.StatusNotFound, "skill_not_found", "The requested skill was not found.")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"name": name, "content": content})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"skills": discoverSkills(root)})
}
func (s *Server) handleMemory(w http.ResponseWriter, r *http.Request) {
	root := filepath.Join(s.config.HermesHome, "memories")
	if name := strings.TrimSpace(r.URL.Query().Get("name")); name != "" {
		path := filepath.Join(root, filepath.Base(name))
		if name == "SOUL.md" {
			path = filepath.Join(s.config.HermesHome, name)
		}
		data, err := os.ReadFile(path)
		if err != nil || len(data) > 1<<20 {
			writeError(w, http.StatusNotFound, "memory_not_found", "The requested note was not found.")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"name": name, "content": string(data)})
		return
	}
	notes := make([]map[string]string, 0)
	for _, name := range []string{"MEMORY.md", "USER.md"} {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			notes = append(notes, map[string]string{"name": name, "path": name})
		}
	}
	if _, err := os.Stat(filepath.Join(s.config.HermesHome, "SOUL.md")); err == nil {
		notes = append(notes, map[string]string{"name": "SOUL.md", "path": "../SOUL.md"})
	}
	writeJSON(w, http.StatusOK, map[string]any{"notes": notes, "sources": []string{"hermes-state"}})
}
func (s *Server) handleCapabilities(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"tasks": true, "skills": true, "memory": true, "todos": true, "goals": true, "spaces": true, "preferences": true, "background": false, "voice": false, "terminal": false, "extensions": false, "reason": "Runtime integrations require an explicit Hermes provider contract."})
}
func (s *Server) handlePlugins(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"plugins": discoverFiles(filepath.Join(s.config.StateDir, "plugins"), ""), "empty": true, "supported_hooks": []string{"pre_tool_call", "post_tool_call", "pre_llm_call", "post_llm_call"}})
}
func discoverFiles(root, extension string) []map[string]string {
	items, err := os.ReadDir(root)
	if err != nil {
		return []map[string]string{}
	}
	result := make([]map[string]string, 0)
	for _, item := range items {
		if item.IsDir() || (extension != "" && !strings.HasSuffix(item.Name(), extension)) {
			continue
		}
		result = append(result, map[string]string{"name": item.Name(), "path": item.Name()})
	}
	return result
}

func discoverSkills(root string) []map[string]string {
	items := make([]map[string]string, 0)
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || entry.Name() != "SKILL.md" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		name, description := skillMetadata(string(data), filepath.Base(filepath.Dir(path)))
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		items = append(items, map[string]string{"name": name, "description": description, "path": filepath.ToSlash(rel)})
		return nil
	})
	return items
}

func skillMetadata(content, fallback string) (string, string) {
	name, description := fallback, ""
	if strings.HasPrefix(strings.TrimSpace(content), "---") {
		parts := strings.SplitN(strings.TrimSpace(content), "---", 3)
		if len(parts) >= 2 {
			for _, line := range strings.Split(parts[1], "\n") {
				key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
				if !ok {
					continue
				}
				value = strings.Trim(strings.TrimSpace(value), "\"'")
				switch key {
				case "name":
					name = value
				case "description":
					description = value
				}
			}
		}
	}
	return name, description
}

func readSkill(root, name string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || filepath.IsAbs(clean) || filepath.Base(clean) != "SKILL.md" {
		return "", os.ErrNotExist
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	resolvedPath, err := filepath.EvalSymlinks(filepath.Join(root, clean))
	if err != nil {
		return "", err
	}
	if resolvedPath != resolvedRoot && !strings.HasPrefix(resolvedPath, resolvedRoot+string(filepath.Separator)) {
		return "", os.ErrNotExist
	}
	data, err := os.ReadFile(resolvedPath)
	if err != nil || len(data) > 1<<20 {
		return "", os.ErrNotExist
	}
	return string(data), nil
}
