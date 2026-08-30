package httpapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/workspace"
)

func (s *Server) workspaceReady(w http.ResponseWriter) (*workspace.Service, bool) {
	if s.workspaceErr != nil || s.workspace == nil {
		writeError(w, http.StatusServiceUnavailable, "workspace_unavailable", "Workspace is unavailable.")
		return nil, false
	}
	return s.workspace, true
}

func workspacePath(r *http.Request) string { return r.URL.Query().Get("path") }

func workspaceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, workspace.ErrInvalidPath):
		writeError(w, http.StatusBadRequest, "invalid_workspace_path", "The workspace path is invalid.")
	case errors.Is(err, os.ErrNotExist):
		writeError(w, http.StatusNotFound, "workspace_not_found", "The workspace item was not found.")
	case errors.Is(err, os.ErrExist):
		writeError(w, http.StatusConflict, "workspace_exists", "The workspace item already exists.")
	default:
		writeError(w, http.StatusInternalServerError, "workspace_error", "The workspace operation failed.")
	}
}

func (s *Server) handleWorkspaceTree(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.workspaceReady(w)
	if !ok {
		return
	}
	entries, err := ws.List(workspacePath(r))
	if err != nil {
		workspaceError(w, err)
		return
	}
	path := workspacePath(r)
	if strings.TrimSpace(path) == "" {
		path = "."
	}
	writeJSON(w, http.StatusOK, map[string]any{"root": filepath.Base(ws.Root()), "path": filepath.ToSlash(filepath.Clean(path)), "entries": entries})
}

func (s *Server) handleWorkspacePreview(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.workspaceReady(w)
	if !ok {
		return
	}
	preview, err := ws.Preview(workspacePath(r))
	if err != nil {
		workspaceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, preview)
}

func (s *Server) handleWorkspaceDownload(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.workspaceReady(w)
	if !ok {
		return
	}
	data, entry, err := ws.Read(workspacePath(r))
	if err != nil {
		workspaceError(w, err)
		return
	}
	w.Header().Set("Content-Type", entry.MIME)
	if entry.MIME == "" {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+strings.ReplaceAll(entry.Name, `"`, "")+`"`)
	http.ServeContent(w, r, entry.Name, entry.ModifiedAt, strings.NewReader(string(data)))
}

func (s *Server) handleWorkspaceGit(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.workspaceReady(w)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	writeJSON(w, http.StatusOK, ws.Git(ctx, workspacePath(r)))
}

func (s *Server) handleWorkspaceCreate(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.workspaceReady(w)
	if !ok {
		return
	}
	var input struct{ Path, Type, Content string }
	if !decodeBody(w, r, &input) {
		return
	}
	if err := ws.Create(input.Path, input.Type, input.Content); err != nil {
		workspaceError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleWorkspaceWrite(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.workspaceReady(w)
	if !ok {
		return
	}
	var input struct{ Path, Content string }
	if !decodeBody(w, r, &input) {
		return
	}
	preview, err := ws.Write(input.Path, input.Content)
	if err != nil {
		workspaceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, preview)
}

func (s *Server) handleWorkspaceRename(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.workspaceReady(w)
	if !ok {
		return
	}
	var input struct{ Path, Name string }
	if !decodeBody(w, r, &input) {
		return
	}
	if err := ws.Rename(input.Path, input.Name); err != nil {
		workspaceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleWorkspaceDelete(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.workspaceReady(w)
	if !ok {
		return
	}
	if err := ws.Delete(workspacePath(r)); err != nil {
		workspaceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleWorkspaceUpload(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.workspaceReady(w)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(workspace.MaxUpload + 1<<20); err != nil {
		workspaceError(w, err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file_required", "A file is required.")
		return
	}
	defer file.Close()
	entry, err := ws.SaveUpload(r.FormValue("path"), filepath.Base(header.Filename), io.LimitReader(file, workspace.MaxUpload+1))
	if err != nil {
		workspaceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, entry)
}
