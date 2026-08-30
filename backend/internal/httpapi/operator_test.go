package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

func TestSpacesAndKanbanUsePersistedControlState(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir()}
	h := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()
	create := httptest.NewRequest(http.MethodPost, "/api/control/spaces", strings.NewReader(`{"title":"Project","metadata":{"path":"/tmp/project"}}`))
	create.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, create)
	if rec.Code != http.StatusCreated {
		t.Fatalf("space=%d %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() == "" {
		t.Fatal("space response empty")
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/spaces", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Project") {
		t.Fatalf("spaces=%d %s", rec.Code, rec.Body.String())
	}
	board := httptest.NewRequest(http.MethodPost, "/api/kanban/boards", strings.NewReader(`{"title":"Delivery"}`))
	board.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, board)
	if rec.Code != http.StatusCreated {
		t.Fatalf("board=%d %s", rec.Code, rec.Body.String())
	}
	card := httptest.NewRequest(http.MethodPost, "/api/kanban/cards", strings.NewReader(`{"title":"Ship slice","metadata":{"board_id":"board"}}`))
	card.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, card)
	if rec.Code != http.StatusCreated {
		t.Fatalf("card=%d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/kanban", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Ship slice") {
		t.Fatalf("kanban=%d %s", rec.Code, rec.Body.String())
	}
}
