package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/control"
)

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
	active := store.Preferences()["active_space"]
	for i := range items {
		if items[i].ID == active {
			if items[i].Metadata == nil {
				items[i].Metadata = map[string]string{}
			}
			items[i].Metadata["active"] = "true"
		}
	}
	writeJSON(w, 200, map[string]any{"spaces": items, "active": active})
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
	items, _ := store.List("spaces")
	found := false
	for _, item := range items {
		if item.ID == input.ID {
			found = true
		}
	}
	if !found {
		writeError(w, 404, "space_not_found", "The space was not found.")
		return
	}
	if err := store.SetPreferences(map[string]string{"active_space": input.ID}); err != nil {
		writeError(w, 500, "space_activate_failed", "The space could not be activated.")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "active": input.ID})
}

func (s *Server) handleOperatorHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"workspace": s.workspaceErr == nil && s.workspace != nil, "control": s.controlErr == nil && s.control != nil, "gateway": s.gateway != nil})
}

func (s *Server) handleOperatorLogs(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"entries": []any{}, "source": "hermes-web-studio runtime", "available": true})
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
