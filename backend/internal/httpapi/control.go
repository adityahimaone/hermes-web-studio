package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/control"
)

func (s *Server) controlReady(w http.ResponseWriter) (*control.Store, bool) {
	if s.controlErr != nil || s.control == nil {
		writeError(w, 503, "control_unavailable", "Control center state is unavailable.")
		return nil, false
	}
	return s.control, true
}
func (s *Server) handleControlList(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	items, err := store.List(r.PathValue("collection"))
	if err != nil {
		writeError(w, 400, "invalid_collection", "The control collection is invalid.")
		return
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
func (s *Server) handleControlCreate(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	var item control.Item
	if !decodeBody(w, r, &item) {
		return
	}
	if item.Title == "" {
		writeError(w, 400, "title_required", "A title is required.")
		return
	}
	created, err := store.Create(r.PathValue("collection"), item)
	if err != nil {
		writeError(w, 400, "invalid_collection", "The control collection is invalid.")
		return
	}
	writeJSON(w, 201, created)
}
func (s *Server) handleControlUpdate(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	var patch control.Item
	if !decodeBody(w, r, &patch) {
		return
	}
	item, err := store.Update(r.PathValue("collection"), r.PathValue("id"), patch)
	if errors.Is(err, control.ErrNotFound) {
		writeError(w, 404, "control_item_not_found", "The control item was not found.")
		return
	}
	if err != nil {
		writeError(w, 400, "invalid_collection", "The control collection is invalid.")
		return
	}
	writeJSON(w, 200, item)
}
func (s *Server) handleControlDelete(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	err := store.Delete(r.PathValue("collection"), r.PathValue("id"))
	if errors.Is(err, control.ErrNotFound) {
		writeError(w, 404, "control_item_not_found", "The control item was not found.")
		return
	}
	if err != nil {
		writeError(w, 400, "invalid_collection", "The control collection is invalid.")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}
func (s *Server) handlePreferences(w http.ResponseWriter, _ *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	writeJSON(w, 200, map[string]any{"preferences": profilePreferences(store.Preferences(), s.activeProfileID())})
}
func profilePreferences(values map[string]string, profileID string) map[string]string {
	prefix := "profile:" + profileID + ":"
	result := map[string]string{}
	for key, value := range values {
		if strings.HasPrefix(key, prefix) {
			result[strings.TrimPrefix(key, prefix)] = value
		}
	}
	return result
}

func (s *Server) handlePreferencesUpdate(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	var values map[string]string
	if !decodeBody(w, r, &values) {
		return
	}
	profileID := s.activeProfileID()
	scoped := map[string]string{}
	for key, value := range values {
		scoped["profile:"+profileID+":"+key] = value
	}
	if err := store.SetPreferences(scoped); err != nil {
		writeError(w, 400, "preferences_invalid", "Preferences could not be saved.")
		return
	}
	writeJSON(w, 200, map[string]any{"preferences": profilePreferences(store.Preferences(), profileID)})
}

func (s *Server) handleCrons(w http.ResponseWriter, _ *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	items, err := store.List("tasks")
	if err != nil {
		writeError(w, 400, "cron_unavailable", "Scheduled tasks are unavailable.")
		return
	}
	writeJSON(w, 200, map[string]any{"jobs": items, "available": true})
}
func (s *Server) handleCronHistory(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	writeJSON(w, 200, map[string]any{"history": store.History(r.URL.Query().Get("job_id"))})
}
func (s *Server) handleCronCreate(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	var input struct {
		control.Item
		Name     string `json:"name"`
		Schedule string `json:"schedule"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	item := input.Item
	if item.Title == "" {
		item.Title = input.Name
	}
	if item.Title == "" {
		item.Title = item.Description
	}
	if item.Title == "" {
		writeError(w, 400, "name_required", "A scheduled task name is required.")
		return
	}
	if item.Metadata == nil {
		item.Metadata = map[string]string{}
	}
	if input.Schedule != "" {
		item.Metadata["schedule"] = input.Schedule
	} else if schedule := r.URL.Query().Get("schedule"); schedule != "" {
		item.Metadata["schedule"] = schedule
	}
	created, err := store.Create("tasks", item)
	if err != nil {
		writeError(w, 400, "cron_invalid", "The scheduled task is invalid.")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "job": created})
}
func (s *Server) handleCronRun(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	var input struct {
		JobID string `json:"job_id"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	record, err := store.RunTask(input.JobID)
	if errors.Is(err, control.ErrNotFound) {
		writeError(w, 404, "cron_not_found", "The scheduled task was not found.")
		return
	}
	if err != nil {
		writeError(w, 500, "cron_run_failed", "The scheduled task could not run.")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "run": record})
}
func (s *Server) handleCronPause(w http.ResponseWriter, r *http.Request) {
	s.handleCronPauseResume(w, r, true)
}
func (s *Server) handleCronResume(w http.ResponseWriter, r *http.Request) {
	s.handleCronPauseResume(w, r, false)
}
func (s *Server) handleCronPauseResume(w http.ResponseWriter, r *http.Request, paused bool) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	var input struct {
		JobID string `json:"job_id"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	if err := store.PauseTask(input.JobID, paused); err != nil {
		if errors.Is(err, control.ErrNotFound) {
			writeError(w, 404, "cron_not_found", "The scheduled task was not found.")
			return
		}
		writeError(w, 500, "cron_update_failed", "The scheduled task could not be updated.")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "paused": paused})
}
