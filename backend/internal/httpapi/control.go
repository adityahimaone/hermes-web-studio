package httpapi

import (
	"errors"
	"net/http"

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
	writeJSON(w, 200, map[string]any{"preferences": store.Preferences()})
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
	if err := store.SetPreferences(values); err != nil {
		writeError(w, 400, "preferences_invalid", "Preferences could not be saved.")
		return
	}
	writeJSON(w, 200, map[string]any{"preferences": store.Preferences()})
}
