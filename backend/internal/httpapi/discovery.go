package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) handleSkills(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"skills": discoverFiles(filepath.Join(s.config.StateDir, "skills"), ".md")})
}
func (s *Server) handleMemory(w http.ResponseWriter, _ *http.Request) {
	root := s.config.StateDir
	notes := discoverFiles(root, ".md")
	writeJSON(w, 200, map[string]any{"notes": notes, "sources": []string{"server-state"}})
}
func (s *Server) handleCapabilities(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"tasks": true, "skills": true, "memory": true, "todos": true, "goals": true, "spaces": true, "preferences": true, "background": false, "voice": false, "terminal": false, "extensions": false, "reason": "Runtime integrations require an explicit Hermes provider contract."})
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
