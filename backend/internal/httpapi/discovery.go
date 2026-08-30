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
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) > 200 {
		writeError(w, http.StatusBadRequest, "search_query_too_long", "The skill search query is too long.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"skills": filterSkills(discoverSkills(root), query)})
}

func (s *Server) handleSkillCreate(w http.ResponseWriter, r *http.Request) {
	var input struct{ Name, Content string }
	if !decodeBody(w, r, &input) {
		return
	}
	name := strings.TrimSpace(input.Name)
	if !validSkillDirectory(name) {
		writeError(w, http.StatusBadRequest, "invalid_skill_name", "Skill names may contain letters, numbers, dots, hyphens, and underscores.")
		return
	}
	if len(input.Content) > 1<<20 {
		writeError(w, http.StatusRequestEntityTooLarge, "skill_too_large", "The skill content is too large.")
		return
	}
	path := filepath.Join(s.config.HermesHome, "skills", name, "SKILL.md")
	if _, err := os.Stat(path); err == nil {
		writeError(w, http.StatusConflict, "skill_exists", "The skill already exists.")
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil || os.WriteFile(path, []byte(input.Content), 0600) != nil {
		writeError(w, http.StatusInternalServerError, "skill_write_failed", "The skill could not be created.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"name": name, "path": filepath.ToSlash(filepath.Join(name, "SKILL.md")), "content": input.Content})
}

func (s *Server) handleSkillUpdate(w http.ResponseWriter, r *http.Request) {
	var input struct{ Name, Content string }
	if !decodeBody(w, r, &input) {
		return
	}
	path, ok := safeSkillPath(filepath.Join(s.config.HermesHome, "skills"), input.Name)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_skill_path", "The skill path is invalid.")
		return
	}
	if len(input.Content) > 1<<20 {
		writeError(w, http.StatusRequestEntityTooLarge, "skill_too_large", "The skill content is too large.")
		return
	}
	if _, err := os.Stat(path); err != nil {
		writeError(w, http.StatusNotFound, "skill_not_found", "The requested skill was not found.")
		return
	}
	if err := os.WriteFile(path, []byte(input.Content), 0600); err != nil {
		writeError(w, http.StatusInternalServerError, "skill_write_failed", "The skill could not be saved.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"name": input.Name, "content": input.Content})
}

func (s *Server) handleSkillDelete(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	path, ok := safeSkillPath(filepath.Join(s.config.HermesHome, "skills"), name)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_skill_path", "The skill path is invalid.")
		return
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "skill_not_found", "The requested skill was not found.")
			return
		}
		writeError(w, http.StatusInternalServerError, "skill_delete_failed", "The skill could not be deleted.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "name": name})
}

func filterSkills(items []map[string]string, query string) []map[string]string {
	if query == "" {
		return items
	}
	query = strings.ToLower(query)
	result := make([]map[string]string, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item["name"]+" "+item["description"]+" "+item["path"]), query) {
			result = append(result, item)
		}
	}
	return result
}

func validSkillDirectory(name string) bool {
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name {
		return false
	}
	for _, char := range name {
		if !(char == '-' || char == '_' || char == '.' || char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9') {
			return false
		}
	}
	return true
}

func safeSkillPath(root, name string) (string, bool) {
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || filepath.IsAbs(clean) || filepath.Base(clean) != "SKILL.md" || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", false
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", false
	}
	resolved, err := filepath.EvalSymlinks(filepath.Join(root, clean))
	if err != nil {
		return "", false
	}
	return resolved, resolved != resolvedRoot && strings.HasPrefix(resolved, resolvedRoot+string(filepath.Separator))
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
		info, _ := os.Stat(path)
		writeJSON(w, http.StatusOK, map[string]any{"name": name, "content": string(data), "updated_at": info.ModTime().UTC()})
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
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query != "" {
		filtered := notes[:0]
		for _, note := range notes {
			if strings.Contains(strings.ToLower(note["name"]), strings.ToLower(query)) {
				filtered = append(filtered, note)
			}
		}
		notes = filtered
	}
	writeJSON(w, http.StatusOK, map[string]any{"notes": notes, "sources": []string{"hermes-state"}})
}

func (s *Server) memoryPath(name string) (string, bool) {
	clean := filepath.Clean(filepath.FromSlash(strings.TrimSpace(name)))
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || filepath.Base(clean) != clean || !strings.HasSuffix(clean, ".md") {
		return "", false
	}
	root := filepath.Join(s.config.HermesHome, "memories")
	if clean == "SOUL.md" {
		root = s.config.HermesHome
	}
	return filepath.Join(root, clean), true
}

func (s *Server) handleMemoryWrite(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	path, ok := s.memoryPath(input.Name)
	if !ok || strings.TrimSpace(input.Name) == "SOUL.md" || len(input.Content) > 1<<20 {
		writeError(w, http.StatusBadRequest, "invalid_memory_path", "The memory note path or content is invalid.")
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil || os.WriteFile(path, []byte(input.Content), 0600) != nil {
		writeError(w, http.StatusInternalServerError, "memory_write_failed", "The memory note could not be saved.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"name": filepath.Base(path), "content": input.Content})
}

func (s *Server) handleMemoryDelete(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	path, ok := s.memoryPath(name)
	if !ok || name == "SOUL.md" {
		writeError(w, http.StatusBadRequest, "invalid_memory_path", "The memory note path is invalid.")
		return
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "memory_not_found", "The memory note was not found.")
			return
		}
		writeError(w, http.StatusInternalServerError, "memory_delete_failed", "The memory note could not be deleted.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
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
