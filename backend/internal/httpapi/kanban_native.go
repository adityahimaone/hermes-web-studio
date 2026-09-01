package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type kanbanCLI struct{ path string }

func (s *Server) kanbanCLI() kanbanCLI { return kanbanCLI{path: s.config.HermesCLIPath} }

func (c kanbanCLI) run(ctx context.Context, board string, args ...string) ([]byte, error) {
	if strings.TrimSpace(c.path) == "" {
		return nil, errors.New("Hermes CLI is not configured")
	}
	all := []string{"kanban"}
	if board != "" {
		all = append(all, "--board", board)
	}
	all = append(all, args...)
	cmd := exec.CommandContext(ctx, c.path, all...)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("hermes kanban failed: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, err
	}
	return output, nil
}

func (s *Server) handleKanbanCapabilities(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	_, err := s.kanbanCLI().run(ctx, "", "stats", "--json")
	available := err == nil
	writeJSON(w, http.StatusOK, map[string]any{
		"transport": "cli", "available": available,
		"dashboard": map[string]any{"configured": s.config.KanbanDashboardURL != "", "available": false},
		"features": map[string]any{
			"read": available, "create": available, "dispatch": available,
			"assign": available, "comments": available, "links": available,
			"live_updates": false, "arbitrary_edit": false, "bulk": false,
			"orchestration": false, "board_metadata": true,
		},
		"statuses": []string{"triage", "todo", "scheduled", "ready", "running", "blocked", "review", "done", "archived"},
	})
}

func (s *Server) handleKanbanBoards(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	out, err := s.kanbanCLI().run(ctx, "", "boards", "list", "--json", "--all")
	if err != nil {
		writeKanbanError(w, err)
		return
	}
	writeRawJSON(w, out)
}

func (s *Server) handleKanbanBoardNative(w http.ResponseWriter, r *http.Request) {
	board := strings.TrimSpace(r.URL.Query().Get("board"))
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	args := []string{"list", "--json"}
	if r.URL.Query().Get("include_archived") == "true" {
		args = append(args, "--archived")
	}
	if tenant := strings.TrimSpace(r.URL.Query().Get("tenant")); tenant != "" {
		args = append(args, "--tenant", tenant)
	}
	out, err := s.kanbanCLI().run(ctx, board, args...)
	if err != nil {
		writeKanbanError(w, err)
		return
	}
	writeRawJSON(w, out)
}

func (s *Server) handleKanbanTaskNative(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	out, err := s.kanbanCLI().run(ctx, "", "show", id, "--json")
	if err != nil {
		writeKanbanError(w, err)
		return
	}
	writeRawJSON(w, out)
}

type nativeKanbanCreate struct {
	Title             string   `json:"title"`
	Body              string   `json:"body"`
	Assignee          string   `json:"assignee"`
	Tenant            string   `json:"tenant"`
	Priority          *int     `json:"priority"`
	Workspace         string   `json:"workspace"`
	Parents           []string `json:"parents"`
	Triage            bool     `json:"triage"`
	Skills            []string `json:"skills"`
	MaxRuntimeSeconds *int     `json:"max_runtime_seconds"`
	MaxRetries        *int     `json:"max_retries"`
	GoalMode          bool     `json:"goal_mode"`
	GoalMaxTurns      *int     `json:"goal_max_turns"`
	IdempotencyKey    string   `json:"idempotency_key"`
}

func (s *Server) handleKanbanTaskCreateNative(w http.ResponseWriter, r *http.Request) {
	var input nativeKanbanCreate
	if !decodeBody(w, r, &input) {
		return
	}
	if strings.TrimSpace(input.Title) == "" {
		writeError(w, 400, "title_required", "A task title is required.")
		return
	}
	board := strings.TrimSpace(r.URL.Query().Get("board"))
	workspace, ok := s.canonicalKanbanWorkspace(input.Workspace)
	if !ok {
		writeError(w, http.StatusBadRequest, "workspace_invalid", "The task workspace reference is invalid or unavailable.")
		return
	}
	// Revalidate immediately before subprocess construction to fail closed on symlink replacement.
	workspace, ok = s.canonicalKanbanWorkspace(workspace)
	if !ok {
		writeError(w, http.StatusBadRequest, "workspace_invalid", "The task workspace reference is invalid or unavailable.")
		return
	}
	args := []string{"create", input.Title, "--json", "--workspace", workspace}
	if input.Body != "" {
		args = append(args, "--body", input.Body)
	}
	if input.Assignee != "" {
		args = append(args, "--assignee", input.Assignee)
	}
	if input.Tenant != "" {
		args = append(args, "--tenant", input.Tenant)
	}
	if input.Priority != nil {
		args = append(args, "--priority", strconv.Itoa(*input.Priority))
	}
	if input.Triage {
		args = append(args, "--triage")
	}
	if input.IdempotencyKey != "" {
		args = append(args, "--idempotency-key", input.IdempotencyKey)
	}
	for _, parent := range input.Parents {
		if strings.TrimSpace(parent) != "" {
			args = append(args, "--parent", strings.TrimSpace(parent))
		}
	}
	for _, skill := range input.Skills {
		if strings.TrimSpace(skill) != "" {
			args = append(args, "--skill", strings.TrimSpace(skill))
		}
	}
	if input.MaxRuntimeSeconds != nil {
		args = append(args, "--max-runtime", strconv.Itoa(*input.MaxRuntimeSeconds))
	}
	if input.MaxRetries != nil {
		args = append(args, "--max-retries", strconv.Itoa(*input.MaxRetries))
	}
	if input.GoalMode {
		args = append(args, "--goal")
	}
	if input.GoalMaxTurns != nil {
		args = append(args, "--goal-max-turns", strconv.Itoa(*input.GoalMaxTurns))
	}
	// Final containment check sits directly before CLI execution.
	workspace, ok = s.canonicalKanbanWorkspace(input.Workspace)
	if !ok {
		writeError(w, http.StatusBadRequest, "workspace_invalid", "The task workspace reference is invalid or unavailable.")
		return
	}
	args[4] = workspace
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	out, err := s.kanbanCLI().run(ctx, board, args...)
	if err != nil {
		writeKanbanError(w, err)
		return
	}
	writeRawJSONStatus(w, http.StatusCreated, out)
}

func (s *Server) canonicalKanbanWorkspace(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || value == "scratch" {
		return "scratch", true
	}
	prefix, ref, ok := strings.Cut(value, ":")
	if !ok || (prefix != "dir" && prefix != "worktree") || ref == "Remote workspace unavailable" || !safeSpaceRef(s, "local", ref) {
		return "", false
	}
	return prefix + ":" + ref, true
}

func (s *Server) handleKanbanTaskActionNative(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	board := strings.TrimSpace(r.URL.Query().Get("board"))
	var body struct {
		Reason string `json:"reason"`
	}
	if r.Body != nil && r.ContentLength != 0 {
		if !decodeBody(w, r, &body) {
			return
		}
	}
	action := r.PathValue("action")
	if !supportedKanbanActions[action] {
		writeError(w, http.StatusBadRequest, "action_invalid", "The Kanban action is not supported.")
		return
	}
	args := []string{action, id}
	if body.Reason != "" && (action == "block" || action == "schedule" || action == "promote" || action == "unblock") {
		args = append(args, body.Reason)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	out, err := s.kanbanCLI().run(ctx, board, args...)
	if err != nil {
		writeKanbanError(w, err)
		return
	}
	_ = out // Native action output stays server-side; browser receives stable status only.
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "action": action, "task_id": id})
}

var supportedKanbanActions = map[string]bool{
	"complete": true, "archive": true, "block": true, "schedule": true,
	"promote": true, "unblock": true, "assign": true, "comment": true,
}

func (s *Server) handleKanbanDispatchNative(w http.ResponseWriter, r *http.Request) {
	args := []string{"dispatch", "--json"}
	if r.URL.Query().Get("dry_run") == "true" {
		args = append(args, "--dry-run")
	}
	if max := strings.TrimSpace(r.URL.Query().Get("max")); max != "" {
		args = append(args, "--max", max)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	out, err := s.kanbanCLI().run(ctx, strings.TrimSpace(r.URL.Query().Get("board")), args...)
	if err != nil {
		writeKanbanError(w, err)
		return
	}
	writeRawJSON(w, out)
}

func (s *Server) handleKanbanEventsNative(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, 500, "events_unavailable", "Live Kanban updates are unavailable.")
		return
	}
	_, _ = fmt.Fprint(w, "event: ready\ndata: {\"transport\":\"cli\",\"live\":false}\n\n")
	flusher.Flush()
	<-r.Context().Done()
}

func writeKanbanError(w http.ResponseWriter, err error) {
	message := "Hermes Kanban is unavailable. Start Hermes or check the CLI configuration."
	if err != nil && strings.Contains(err.Error(), "failed:") {
		message = "Hermes Kanban rejected the request."
	}
	writeError(w, http.StatusBadGateway, "kanban_unavailable", message)
}

func writeRawJSON(w http.ResponseWriter, data []byte) {
	value, err := extractJSON(data)
	if err != nil {
		writeError(w, 502, "kanban_invalid_response", "Hermes returned an invalid Kanban response.")
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func writeRawJSONStatus(w http.ResponseWriter, status int, data []byte) {
	value, err := extractJSON(data)
	if err != nil {
		writeError(w, 502, "kanban_invalid_response", "Hermes returned an invalid Kanban response.")
		return
	}
	writeJSON(w, status, value)
}

// extractJSON accepts the machine-readable document while tolerating harmless
// CLI banners or trailing notices. The subprocess boundary is still strict:
// the first JSON object/array must decode completely, and no arbitrary text is
// ever forwarded to the browser.
func extractJSON(data []byte) (any, error) {
	start := bytes.IndexAny(data, "{[")
	if start < 0 {
		return nil, errors.New("no JSON document")
	}
	decoder := json.NewDecoder(bytes.NewReader(data[start:]))
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, errors.New("trailing JSON data")
		}
		return nil, err
	}
	return value, nil
}
