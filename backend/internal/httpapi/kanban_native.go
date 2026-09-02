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

const (
	maxKanbanActionReasonLength = 512
	maxKanbanProjectionItems    = 100
	maxKanbanCLIOutput          = 1 << 20
	maxKanbanIdentifierLength   = 128
	maxKanbanPriority           = 100
	maxKanbanRuntimeSeconds     = 86400
	maxKanbanRetries            = 10
	maxKanbanGoalTurns          = 1000
)

func validateKanbanCreateNumber(name string, value *int) error {
	if value == nil {
		return nil
	}
	if *value < 0 {
		return errors.New("invalid kanban numeric field")
	}
	max := map[string]int{"priority": maxKanbanPriority, "max_runtime_seconds": maxKanbanRuntimeSeconds, "max_retries": maxKanbanRetries, "goal_max_turns": maxKanbanGoalTurns}[name]
	if name == "max_runtime_seconds" || name == "goal_max_turns" {
		if *value == 0 {
			return errors.New("invalid kanban numeric field")
		}
	}
	if *value > max {
		return errors.New("invalid kanban numeric field")
	}
	return nil
}

func validateKanbanIdentifier(value string, allowEmpty bool) (string, error) {
	if allowEmpty && value == "" {
		return "", nil
	}
	if value == "" || len(value) > maxKanbanIdentifierLength || strings.TrimSpace(value) != value || strings.HasPrefix(value, "-") || strings.ContainsAny(value, "/\\\\;|&$`<>()\"") {
		return "", errors.New("invalid kanban identifier")
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return "", errors.New("invalid kanban identifier")
		}
	}
	return value, nil
}

func validateKanbanMax(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 || n > 100 {
		return "", errors.New("invalid kanban max")
	}
	return strconv.Itoa(n), nil
}

func validateKanbanArgument(value string, allowEmpty bool) error {
	if allowEmpty && value == "" {
		return nil
	}
	if value == "" || len(value) > 2048 || strings.TrimSpace(value) != value || strings.HasPrefix(value, "-") {
		return errors.New("invalid kanban argument")
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return errors.New("invalid kanban argument")
		}
	}
	return nil
}

func validateKanbanArgumentList(values []string) error {
	if len(values) > maxKanbanProjectionItems {
		return errors.New("too many kanban arguments")
	}
	for _, value := range values {
		if err := validateKanbanArgument(value, false); err != nil {
			return err
		}
	}
	return nil
}

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
	if len(output) > maxKanbanCLIOutput {
		return nil, errors.New("hermes kanban output too large")
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
			"assign": false, "comments": false, "links": available,
			"live_updates": false, "edit": available, "arbitrary_edit": false, "bulk": false,
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
	writeKanbanProjection(w, http.StatusOK, out)
}

func (s *Server) handleKanbanBoardNative(w http.ResponseWriter, r *http.Request) {
	board, err := validateKanbanIdentifier(r.URL.Query().Get("board"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "board_invalid", "The Kanban board is invalid.")
		return
	}
	tenant, err := validateKanbanIdentifier(r.URL.Query().Get("tenant"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "tenant_invalid", "The Kanban tenant is invalid.")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	args := []string{"list", "--json"}
	if r.URL.Query().Get("include_archived") == "true" {
		args = append(args, "--archived")
	}
	if tenant != "" {
		args = append(args, "--tenant", tenant)
	}
	out, err := s.kanbanCLI().run(ctx, board, args...)
	if err != nil {
		writeKanbanError(w, err)
		return
	}
	writeKanbanProjection(w, http.StatusOK, out)
}

func (s *Server) handleKanbanTaskNative(w http.ResponseWriter, r *http.Request) {
	id, err := validateKanbanIdentifier(r.PathValue("id"), false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "task_id_invalid", "The Kanban task ID is invalid.")
		return
	}
	board, err := validateKanbanIdentifier(r.URL.Query().Get("board"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "board_invalid", "The Kanban board is invalid.")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	out, err := s.kanbanCLI().run(ctx, board, "show", id, "--json")
	if err != nil {
		writeKanbanError(w, err)
		return
	}
	writeKanbanProjection(w, http.StatusOK, out)
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
	for _, value := range []string{input.Title, input.Body, input.Assignee, input.Tenant, input.IdempotencyKey} {
		if err := validateKanbanArgument(value, true); err != nil {
			writeError(w, http.StatusBadRequest, "task_field_invalid", "A Kanban task field is invalid.")
			return
		}
	}
	if err := validateKanbanArgumentList(input.Parents); err != nil {
		writeError(w, http.StatusBadRequest, "task_field_invalid", "A Kanban task field is invalid.")
		return
	}
	if err := validateKanbanArgumentList(input.Skills); err != nil {
		writeError(w, http.StatusBadRequest, "task_field_invalid", "A Kanban task field is invalid.")
		return
	}
	for name, value := range map[string]*int{"priority": input.Priority, "max_runtime_seconds": input.MaxRuntimeSeconds, "max_retries": input.MaxRetries, "goal_max_turns": input.GoalMaxTurns} {
		if err := validateKanbanCreateNumber(name, value); err != nil {
			writeError(w, http.StatusBadRequest, "task_field_invalid", "A Kanban numeric field is invalid.")
			return
		}
	}
	board, err := validateKanbanIdentifier(r.URL.Query().Get("board"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "board_invalid", "The Kanban board is invalid.")
		return
	}
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
	writeKanbanProjection(w, http.StatusCreated, out)
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
	id, err := validateKanbanIdentifier(r.PathValue("id"), false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "task_id_invalid", "The Kanban task ID is invalid.")
		return
	}
	board, err := validateKanbanIdentifier(r.URL.Query().Get("board"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "board_invalid", "The Kanban board is invalid.")
		return
	}
	var body struct {
		Reason   string         `json:"reason"`
		Result   string         `json:"result"`
		Summary  string         `json:"summary"`
		ChildID  string         `json:"child_id"`
		Metadata map[string]any `json:"metadata"`
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
	if action == "edit" {
		if body.Result == "" && body.Summary == "" {
			writeError(w, http.StatusBadRequest, "edit_required", "Result or summary is required.")
			return
		}
		for _, value := range []string{body.Result, body.Summary} {
			if err := validateKanbanArgument(value, true); err != nil {
				writeError(w, http.StatusBadRequest, "edit_invalid", "The Kanban edit text is invalid.")
				return
			}
		}
		if body.Result != "" {
			args = append(args, "--result", body.Result)
		}
		if body.Summary != "" {
			args = append(args, "--summary", body.Summary)
		}
		if body.Metadata != nil {
			encoded, err := json.Marshal(body.Metadata)
			if err != nil || len(encoded) > maxKanbanActionReasonLength {
				writeError(w, http.StatusBadRequest, "metadata_invalid", "The Kanban metadata is invalid.")
				return
			}
			args = append(args, "--metadata", string(encoded))
		}
	}
	if action == "link" || action == "unlink" {
		child, err := validateKanbanIdentifier(body.ChildID, false)
		if err != nil || child == id {
			writeError(w, http.StatusBadRequest, "link_invalid", "The Kanban linked task ID is invalid.")
			return
		}
		args = append(args, child)
	}
	if body.Reason != "" && (action == "block" || action == "schedule" || action == "promote" || action == "unblock") {
		reason, err := validateKanbanActionReason(body.Reason)
		if err != nil {
			writeError(w, http.StatusBadRequest, "reason_invalid", "The Kanban action reason is invalid.")
			return
		}
		args = append(args, reason)
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
	"edit": true, "link": true, "unlink": true,
}

func (s *Server) handleKanbanDispatchNative(w http.ResponseWriter, r *http.Request) {
	args := []string{"dispatch", "--json"}
	board, err := validateKanbanIdentifier(r.URL.Query().Get("board"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "board_invalid", "The Kanban board is invalid.")
		return
	}
	if r.URL.Query().Get("dry_run") == "true" {
		args = append(args, "--dry-run")
	}
	if max, err := validateKanbanMax(r.URL.Query().Get("max")); err != nil {
		writeError(w, http.StatusBadRequest, "max_invalid", "The Kanban dispatch maximum is invalid.")
		return
	} else if max != "" {
		args = append(args, "--max", max)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	out, err := s.kanbanCLI().run(ctx, board, args...)
	if err != nil {
		writeKanbanError(w, err)
		return
	}
	writeKanbanProjection(w, http.StatusOK, out)
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

func writeKanbanProjection(w http.ResponseWriter, status int, data []byte) {
	if len(data) > maxKanbanCLIOutput {
		writeError(w, http.StatusBadGateway, "kanban_invalid_response", "Hermes returned an invalid Kanban response.")
		return
	}
	value, err := extractJSON(data)
	if err != nil {
		writeError(w, http.StatusBadGateway, "kanban_invalid_response", "Hermes returned an invalid Kanban response.")
		return
	}
	writeJSON(w, status, sanitizeKanbanProjection(value))
}

func sanitizeKanbanProjection(value any) any {
	// Keep browser output limited to fields consumed by kanban-client. In
	// particular, never project arbitrary CLI JSON, workspace paths, or runs.
	switch rows := value.(type) {
	case []any:
		out := make([]map[string]any, 0, min(len(rows), maxKanbanProjectionItems))
		for _, row := range rows[:min(len(rows), maxKanbanProjectionItems)] {
			if item, ok := row.(map[string]any); ok {
				out = append(out, sanitizeKanbanItem(item))
			}
		}
		return out
	case map[string]any:
		out := map[string]any{}
		for _, key := range []string{"tasks", "boards"} {
			if rows[key] != nil {
				out[key] = sanitizeKanbanProjection(rows[key])
			}
		}
		for _, key := range []string{"tenants", "assignees"} {
			if values, ok := rows[key].([]any); ok {
				items := make([]string, 0, min(len(values), maxKanbanProjectionItems))
				for _, value := range values[:min(len(values), maxKanbanProjectionItems)] {
					if text, ok := value.(string); ok && len(text) <= 256 {
						items = append(items, text)
					}
				}
				out[key] = items
			}
		}
		if stats, ok := rows["stats"].(map[string]any); ok {
			safeStats := map[string]any{}
			for _, key := range []string{"triage", "todo", "scheduled", "ready", "running", "blocked", "review", "done", "archived"} {
				if number, ok := stats[key].(float64); ok && number >= 0 && number <= maxKanbanProjectionItems*1000 && number == float64(int64(number)) {
					safeStats[key] = number
				}
			}
			out["stats"] = safeStats
		}
		if len(out) == 0 {
			return sanitizeKanbanItem(rows)
		}
		return out
	default:
		return map[string]any{}
	}
}

func sanitizeKanbanItem(item map[string]any) map[string]any {
	out := map[string]any{}
	for _, key := range []string{"id", "title", "body", "status", "assignee", "tenant", "priority", "comment_count", "created_at", "updated_at", "slug", "name", "task_count", "archived", "parents", "children", "skills"} {
		if value, ok := item[key]; ok {
			if safe, ok := safeBoundedKanbanValue(value); ok {
				out[key] = safe
			}
		}
	}
	if workspace, ok := item["workspace"].(string); ok && safeKanbanWorkspace(workspace) {
		out["workspace"] = workspace
	}
	for _, key := range []string{"comments", "events", "runs"} {
		if values, ok := item[key].([]any); ok {
			out[key] = sanitizeKanbanActivity(values, key)
		}
	}
	if result, ok := item["result"].(string); ok {
		out["result"] = boundedText(result, 2048)
	}
	return out
}

func sanitizeKanbanActivity(values []any, kind string) []map[string]any {
	out := make([]map[string]any, 0, min(len(values), 100))
	for _, raw := range values[:min(len(values), 100)] {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		row := map[string]any{}
		keys := []string{"id", "type", "status", "name", "created_at", "updated_at", "author", "assignee"}
		if kind == "comments" {
			keys = []string{"id", "body", "created_at", "author"}
		}
		if kind == "runs" {
			keys = []string{"id", "status", "created_at", "updated_at", "result"}
		}
		for _, key := range keys {
			if value, ok := item[key]; ok {
				if safe, ok := safeBoundedKanbanValue(value); ok {
					row[key] = safe
				}
			}
		}
		out = append(out, row)
	}
	return out
}

func boundedKanbanValue(value any) any {
	result, _ := safeBoundedKanbanValue(value)
	return result
}

func safeBoundedKanbanValue(value any) (any, bool) {
	switch v := value.(type) {
	case string:
		return boundedText(v, 512), true
	case bool:
		return v, true
	case float64:
		return v, true
	case nil:
		return nil, true
	case []any:
		out := make([]any, 0, min(len(v), 100))
		for _, item := range v[:min(len(v), maxKanbanProjectionItems)] {
			if _, nested := item.([]any); nested {
				return nil, false
			}
			if safe, ok := safeBoundedKanbanValue(item); ok {
				out = append(out, safe)
			} else {
				return nil, false
			}
		}
		return out, true
	default:
		return nil, false
	}
}

func validateKanbanActionReason(reason string) (string, error) {
	if reason == "" || len(reason) > maxKanbanActionReasonLength || strings.TrimSpace(reason) != reason {
		return "", errors.New("invalid action reason")
	}
	if strings.ContainsAny(reason, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f\x7f;|&$`<>\\") {
		return "", errors.New("unsafe action reason")
	}
	return reason, nil
}

func boundedText(value string, limit int) string {
	if len(value) > limit {
		return value[:limit]
	}
	return value
}
func safeKanbanWorkspace(value string) bool {
	return value == "scratch" || (strings.HasPrefix(value, "dir:") || strings.HasPrefix(value, "worktree:")) && !strings.ContainsAny(value, "/\\") && len(value) <= 256
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractJSON accepts the machine-readable document while tolerating harmless
// CLI banners or trailing notices. The subprocess boundary is still strict:
// the first JSON object/array must decode completely, and no arbitrary text is
// ever forwarded to the browser.
func extractJSON(data []byte) (any, error) {
	if len(data) > maxKanbanCLIOutput {
		return nil, errors.New("JSON document too large")
	}
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
