package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/control"
)

type insightsUsage struct {
	Input     int64 `json:"input_tokens"`
	Output    int64 `json:"output_tokens"`
	Total     int64 `json:"total_tokens"`
	Available bool  `json:"available"`
}

type insightsProvider struct {
	Provider string        `json:"provider"`
	Model    string        `json:"model,omitempty"`
	Sessions int           `json:"sessions"`
	Messages int           `json:"messages"`
	Usage    insightsUsage `json:"usage"`
}

func (s *Server) handleOperatorInsights(w http.ResponseWriter, _ *http.Request) {
	if s.sessions == nil {
		writeError(w, http.StatusServiceUnavailable, "insights_unavailable", "Insights state is unavailable.")
		return
	}
	summaries, err := s.sessions.List()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "insights_unavailable", "Insights state could not be read.")
		return
	}
	providers := map[string]*insightsProvider{}
	var usage insightsUsage
	var userMessages, assistantMessages, messageCount, readableSessions int
	var lastActivity time.Time
	for _, summary := range summaries {
		item, loadErr := s.sessions.Load(summary.ID)
		if loadErr != nil {
			continue
		}
		readableSessions++
		if updated := insightTime(summary.Metadata, "updated_at"); updated.After(lastActivity) {
			lastActivity = updated
		}
		provider := insightString(summary.Metadata, "provider", "provider_id")
		model := insightString(summary.Metadata, "model")
		for _, raw := range item.Messages {
			var message map[string]json.RawMessage
			if json.Unmarshal(raw, &message) != nil {
				continue
			}
			messageCount++
			switch insightString(message, "role") {
			case "user":
				userMessages++
			case "assistant":
				assistantMessages++
			}
			if provider == "" {
				provider = insightString(message, "provider", "provider_id")
			}
			if model == "" {
				model = insightString(message, "model")
			}
			insightAddUsage(&usage, message)
		}
		if provider != "" {
			key := provider + "\x00" + model
			row := providers[key]
			if row == nil {
				row = &insightsProvider{Provider: provider, Model: model}
				providers[key] = row
			}
			row.Sessions++
			row.Messages += len(item.Messages)
			insightAddUsage(&row.Usage, summary.Metadata)
			for _, raw := range item.Messages {
				var message map[string]json.RawMessage
				if json.Unmarshal(raw, &message) == nil {
					insightAddUsage(&row.Usage, message)
				}
			}
		}
	}
	providerRows := make([]insightsProvider, 0, len(providers))
	for _, row := range providers {
		providerRows = append(providerRows, *row)
	}
	if usage.Total == 0 && (usage.Input > 0 || usage.Output > 0) {
		usage.Total = usage.Input + usage.Output
	}
	usage.Available = usage.Input > 0 || usage.Output > 0 || usage.Total > 0
	for i := range providerRows {
		if providerRows[i].Usage.Total == 0 && (providerRows[i].Usage.Input > 0 || providerRows[i].Usage.Output > 0) {
			providerRows[i].Usage.Total = providerRows[i].Usage.Input + providerRows[i].Usage.Output
		}
		providerRows[i].Usage.Available = providerRows[i].Usage.Input > 0 || providerRows[i].Usage.Output > 0 || providerRows[i].Usage.Total > 0
	}
	lastActivityValue := any(nil)
	if !lastActivity.IsZero() {
		lastActivityValue = lastActivity.Format(time.RFC3339Nano)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"generated_at":     time.Now().UTC().Format(time.RFC3339Nano),
		"source":           "server-owned Hermes session state",
		"synchronization":  map[string]any{"status": "synchronized", "sessions_scanned": len(summaries), "sessions_read": readableSessions, "last_activity_at": lastActivityValue},
		"summary":          map[string]any{"sessions": readableSessions, "messages": messageCount, "user_messages": userMessages, "assistant_messages": assistantMessages},
		"usage":            usage,
		"provider_history": providerRows,
		"cost":             map[string]any{"available": false, "reason": "No persisted billing data is available from Hermes session state."},
	})
}

func insightString(values map[string]json.RawMessage, keys ...string) string {
	for _, key := range keys {
		var value string
		if json.Unmarshal(values[key], &value) == nil && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func insightNumber(values map[string]json.RawMessage, keys ...string) int64 {
	for _, key := range keys {
		var value float64
		if json.Unmarshal(values[key], &value) == nil && value >= 0 {
			return int64(value)
		}
	}
	return 0
}

func insightAddUsage(usage *insightsUsage, values map[string]json.RawMessage) {
	usage.Input += insightNumber(values, "input_tokens", "prompt_tokens", "input")
	usage.Output += insightNumber(values, "output_tokens", "completion_tokens", "output")
	usage.Total += insightNumber(values, "total_tokens", "total")
	var nested map[string]json.RawMessage
	if json.Unmarshal(values["usage"], &nested) == nil {
		insightAddUsage(usage, nested)
	}
}

func insightTime(values map[string]json.RawMessage, key string) time.Time {
	value := insightString(values, key)
	parsed, _ := time.Parse(time.RFC3339Nano, value)
	return parsed
}

func (s *Server) handleSpaces(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	items, err := store.List("spaces")
	if err != nil {
		writeError(w, 500, "spaces_unavailable", "Spaces are unavailable.")
		return
	}
	s.stateMu.RLock()
	profileID := s.activeProfile
	s.stateMu.RUnlock()
	visible := items[:0]
	for _, item := range items {
		if item.Metadata["profile_id"] == "" || item.Metadata["profile_id"] == profileID {
			visible = append(visible, item)
		}
	}
	items = visible
	active := store.Preferences()["active_space:"+profileID]
	if active == "" {
		active = store.Preferences()["active_space"]
	}
	for i := range items {
		if items[i].ID == active {
			if items[i].Metadata == nil {
				items[i].Metadata = map[string]string{}
			}
			items[i].Metadata["active"] = "true"
		}
	}
	writeJSON(w, 200, map[string]any{"spaces": spacePayloads(items, active), "active": active})
}

func spacePayloads(items []control.Item, active string) []map[string]any {
	sort.SliceStable(items, func(i, j int) bool { return spaceOrder(items[i]) < spaceOrder(items[j]) })
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		metadata := item.Metadata
		if metadata == nil {
			metadata = map[string]string{}
		}
		row := map[string]any{"id": item.ID, "name": item.Title, "title": item.Title, "location_kind": metadata["location_kind"], "workspace_ref": sanitizeSpaceRef(item), "order": spaceOrder(item), "health": metadata["health"], "profile_id": metadata["profile_id"], "active": item.ID == active}
		if row["location_kind"] == "" {
			row["location_kind"] = "local"
		}
		if row["health"] == "" {
			row["health"] = item.Status
		}
		result = append(result, row)
	}
	return result
}

func spaceOrder(item control.Item) int {
	n, err := strconv.Atoi(item.Metadata["order"])
	if err != nil {
		return 0
	}
	return n
}

func (s *Server) handleSpaceCreate(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	s.stateMu.RLock()
	activeProfile := s.activeProfile
	s.stateMu.RUnlock()
	var input struct {
		Name         string `json:"name"`
		Title        string `json:"title"`
		LocationKind string `json:"location_kind"`
		WorkspaceRef string `json:"workspace_ref"`
		Order        int    `json:"order"`
		ProfileID    string `json:"profile_id"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = strings.TrimSpace(input.Title)
	}
	kind := strings.TrimSpace(input.LocationKind)
	if kind == "" {
		kind = "local"
	}
	ref := strings.TrimSpace(input.WorkspaceRef)
	if name == "" || (kind != "local" && kind != "remote") || !safeSpaceRef(s, kind, ref) {
		writeError(w, http.StatusBadRequest, "space_invalid", "Space name, location kind, and workspace reference are required.")
		return
	}
	profileOK := input.ProfileID == "" || input.ProfileID == activeProfile
	if input.ProfileID == "" {
		input.ProfileID = activeProfile
	}
	if !profileOK {
		writeError(w, http.StatusBadRequest, "profile_invalid", "The profile was not found.")
		return
	}
	metadata := map[string]string{"location_kind": kind, "workspace_ref": ref, "order": fmt.Sprint(input.Order), "profile_id": input.ProfileID}
	if kind == "remote" {
		metadata["health"] = "unavailable"
	} else {
		metadata["health"] = "ready"
	}
	activeKey := "active_space:" + activeProfile
	active := store.Preferences()[activeKey]
	created, err := store.CreateSpace(control.Item{Title: name, Status: metadata["health"], Metadata: metadata}, activeKey, active == "")
	if err != nil {
		writeError(w, 500, "space_persist_failed", "The space could not be registered.")
		return
	}
	if active == "" {
		active = created.ID
	}
	writeJSON(w, http.StatusCreated, spacePayloads([]control.Item{created}, active)[0])
}

func safeSpaceRef(s *Server, kind, ref string) bool {
	if ref == "" || strings.ContainsAny(ref, "\\\\\x00:") || strings.HasPrefix(ref, "/") || strings.HasPrefix(ref, "~") || strings.Contains(ref, "..") || strings.Contains(ref, "://") {
		return false
	}
	if strings.Contains(ref, ".ssh") || strings.Contains(ref, "id_rsa") || strings.Contains(ref, "private") || strings.Contains(ref, "secret") {
		return false
	}
	if kind == "remote" {
		return validateRemoteWorkspaceRef(ref)
	}
	clean := filepath.Clean(filepath.FromSlash(ref))
	if clean == "." || s.workspace == nil {
		return false
	}
	root, err := filepath.Abs(s.workspace.Root())
	if err != nil {
		return false
	}
	candidate, err := filepath.Abs(filepath.Join(root, clean))
	if err != nil {
		return false
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return false
	}
	realCandidate, err := filepath.EvalSymlinks(candidate)
	if err == nil {
		rel, relErr := filepath.Rel(realRoot, realCandidate)
		return relErr == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
	}
	if !os.IsNotExist(err) {
		return false
	}
	nearest := filepath.Dir(candidate)
	for {
		realNearest, resolveErr := filepath.EvalSymlinks(nearest)
		if resolveErr == nil {
			rel, relErr := filepath.Rel(realRoot, realNearest)
			return relErr == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
		}
		if !os.IsNotExist(resolveErr) || nearest == root {
			return false
		}
		nearest = filepath.Dir(nearest)
	}
}

func validateRemoteWorkspaceRef(ref string) bool {
	if ref == "" || len(ref) > 128 || strings.HasPrefix(ref, "-") {
		return false
	}
	for _, r := range ref {
		if !(r == '-' || r == '_' || r == '.' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

func sanitizeSpaceRef(item control.Item) string {
	if item.Metadata["location_kind"] == "remote" {
		return "remote:" + item.ID
	}
	ref := item.Metadata["workspace_ref"]
	if ref == "" || strings.HasPrefix(ref, "/") || strings.HasPrefix(ref, "~") || strings.Contains(ref, "..") || strings.ContainsAny(ref, "\\\\\x00:") || strings.Contains(ref, "://") || strings.Contains(ref, ".ssh") || strings.Contains(ref, "private") || strings.Contains(ref, "secret") {
		return "unavailable"
	}
	return ref
}

func (s *Server) handleSpaceDelete(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeError(w, 400, "space_invalid", "The space ID is required.")
		return
	}
	profileID := s.capturedActiveProfile()
	if !s.spaceOwnedByProfile(store, id, profileID) {
		writeError(w, http.StatusNotFound, "space_not_found", "The space was not found.")
		return
	}
	if err := store.SpaceMutation("spaces", id, "active_space:"+profileID, true); errors.Is(err, control.ErrNotFound) {
		writeError(w, 404, "space_not_found", "The space was not found.")
		return
	} else if err != nil {
		if strings.Contains(err.Error(), "active space protected") {
			writeError(w, http.StatusConflict, "active_space_protected", "The active space cannot be deleted.")
		} else {
			writeError(w, 500, "space_delete_failed", "The space could not be deleted.")
		}
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) handleSpaceActivate(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	var input struct {
		ID string `json:"id"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	if strings.TrimSpace(input.ID) == "" {
		writeError(w, http.StatusBadRequest, "space_invalid", "The space ID is required.")
		return
	}
	profileID := s.capturedActiveProfile()
	if !s.spaceOwnedByProfile(store, input.ID, profileID) {
		writeError(w, http.StatusNotFound, "space_not_found", "The space was not found.")
		return
	}
	if err := store.SpaceMutation("spaces", input.ID, "active_space:"+profileID, false); err != nil {
		if errors.Is(err, control.ErrNotFound) {
			writeError(w, 404, "space_not_found", "The space was not found.")
		} else {
			writeError(w, 500, "space_activate_failed", "The space could not be activated.")
		}
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "active": input.ID})
}

func (s *Server) capturedActiveProfile() string {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	return s.activeProfile
}

func (s *Server) spaceOwnedByProfile(store *control.Store, id, profileID string) bool {
	items, err := store.List("spaces")
	if err != nil {
		return false
	}
	for _, item := range items {
		if item.ID == id {
			return item.Metadata["profile_id"] == "" || item.Metadata["profile_id"] == profileID
		}
	}
	return false
}

func (s *Server) handleOperatorHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"workspace": s.workspaceErr == nil && s.workspace != nil, "control": s.controlErr == nil && s.control != nil, "gateway": s.gateway != nil})
}

func (s *Server) handleOperatorLogs(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"entries": []any{}, "source": "hermes-web-studio runtime", "available": true})
}

// handleOperatorDiagnostics exposes a sanitized, read-only snapshot for operators.
// It deliberately reports capabilities and counts only: credentials, workspace
// paths, and mutation controls do not belong in a browser diagnostic payload.
func (s *Server) handleOperatorDiagnostics(w http.ResponseWriter, _ *http.Request) {
	components := map[string]any{
		"gateway":   map[string]any{"available": s.gateway != nil},
		"workspace": map[string]any{"available": s.workspaceErr == nil && s.workspace != nil},
		"control":   map[string]any{"available": s.controlErr == nil && s.control != nil},
		"sessions":  map[string]any{"available": s.sessions != nil},
	}

	collections := map[string]int{}
	if s.control != nil && s.controlErr == nil {
		for _, name := range []string{"tasks", "todos", "goals", "spaces", "artifacts", "boards"} {
			items, err := s.control.List(name)
			if err == nil {
				collections[name] = len(items)
			}
		}
	}

	sessionCount := 0
	if s.sessions != nil {
		if sessions, err := s.sessions.List(); err == nil {
			sessionCount = len(sessions)
		}
	}

	status := "ready"
	if s.controlErr != nil || s.workspaceErr != nil || s.sessions == nil {
		status = "degraded"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":       status,
		"read_only":    true,
		"generated_at": time.Now().UTC().Format(time.RFC3339Nano),
		"components":   components,
		"counts": map[string]any{
			"sessions":    sessionCount,
			"collections": collections,
		},
		"contracts": map[string]string{
			"extensions": "/api/extensions",
			"updates":    "/api/operator/update",
			"version":    "/api/operator/version",
		},
	})
}

func (s *Server) handleOperatorVersion(w http.ResponseWriter, _ *http.Request) {
	version := strings.TrimSpace(os.Getenv("HERMES_WEB_STUDIO_VERSION"))
	if version == "" {
		version = "development"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"read_only":  true,
		"version":    version,
		"go_version": runtime.Version(),
		"service":    "hermes-web-studio",
		"source":     "server runtime",
	})
}

func (s *Server) handleOperatorUpdate(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"read_only": true,
		"state":     "not_configured",
		"reason":    "Update metadata is not configured for this server.",
		"version":   "/api/operator/version",
		"health": map[string]string{
			"liveness":    "/health",
			"readiness":   "/ready",
			"gateway":     "/api/health/hermes",
			"diagnostics": "/api/operator/diagnostics",
		},
		"actions": map[string]any{
			"check":         map[string]any{"available": false, "state": "unavailable", "reason": "No signed update source is configured."},
			"apply":         map[string]any{"available": false, "state": "unavailable", "reason": "Updates are never applied from the browser."},
			"shutdown":      map[string]any{"available": false, "state": "unavailable", "reason": "Process lifecycle is owned by the service supervisor."},
			"restart":       map[string]any{"available": false, "state": "unavailable", "reason": "Process lifecycle is owned by the service supervisor."},
			"lock_recovery": map[string]any{"available": false, "state": "unavailable", "reason": "No update lock is managed by the BFF."},
		},
		"release": map[string]any{
			"status":    "repository-checks-only",
			"artifacts": []string{"embedded-frontend", "docker", "installer", "nix", "multi-architecture"},
			"verified":  false,
		},
	})
}

func (s *Server) handleKanban(w http.ResponseWriter, _ *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	boards, _ := store.List("boards")
	cards, _ := store.List("artifacts")
	writeJSON(w, 200, map[string]any{"boards": boards, "cards": cards})
}

func (s *Server) handleKanbanBoardCreate(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	var item control.Item
	if !decodeBody(w, r, &item) {
		return
	}
	if strings.TrimSpace(item.Title) == "" {
		writeError(w, 400, "title_required", "A board title is required.")
		return
	}
	created, err := store.Create("boards", item)
	if err != nil {
		writeError(w, 400, "kanban_invalid", "The board is invalid.")
		return
	}
	writeJSON(w, 201, created)
}

func (s *Server) handleKanbanCardCreate(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	var item control.Item
	if !decodeBody(w, r, &item) {
		return
	}
	if strings.TrimSpace(item.Title) == "" {
		writeError(w, 400, "title_required", "A card title is required.")
		return
	}
	if item.Metadata == nil {
		item.Metadata = map[string]string{}
	}
	item.Metadata["board_id"] = strings.TrimSpace(item.Metadata["board_id"])
	created, err := store.Create("artifacts", item)
	if err != nil {
		writeError(w, 400, "kanban_invalid", "The card is invalid.")
		return
	}
	writeJSON(w, 201, created)
}

func (s *Server) handleKanbanCardUpdate(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	var patch control.Item
	if !decodeBody(w, r, &patch) {
		return
	}
	item, err := store.Update("artifacts", r.PathValue("id"), patch)
	if errors.Is(err, control.ErrNotFound) {
		writeError(w, 404, "card_not_found", "The card was not found.")
		return
	}
	if err != nil {
		writeError(w, 400, "kanban_invalid", "The card is invalid.")
		return
	}
	writeJSON(w, 200, item)
}

func (s *Server) handleKanbanCardDelete(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	err := store.Delete("artifacts", r.PathValue("id"))
	if errors.Is(err, control.ErrNotFound) {
		writeError(w, 404, "card_not_found", "The card was not found.")
		return
	}
	if err != nil {
		writeError(w, 400, "kanban_invalid", "The card is invalid.")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}
