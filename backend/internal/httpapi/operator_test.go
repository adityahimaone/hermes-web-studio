package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/control"
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

func TestOperatorInsightsReportsPersistedSessionFactsAndNoCostClaim(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir()}
	server := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL}))
	if _, err := server.sessions.Create("session-insights", "Recorded work", []json.RawMessage{
		json.RawMessage(`{"role":"user","content":"hello","provider":"local","model":"qwen"}`),
		json.RawMessage(`{"role":"assistant","content":"hi","usage":{"prompt_tokens":4,"completion_tokens":6,"total_tokens":10}}`),
	}); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/operator/insights", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("insights=%d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Summary struct {
			Sessions int `json:"sessions"`
			Messages int `json:"messages"`
		} `json:"summary"`
		Usage struct {
			Input     int64 `json:"input_tokens"`
			Output    int64 `json:"output_tokens"`
			Total     int64 `json:"total_tokens"`
			Available bool  `json:"available"`
		} `json:"usage"`
		Cost struct {
			Available bool `json:"available"`
		} `json:"cost"`
		ProviderHistory []struct {
			Provider string `json:"provider"`
			Model    string `json:"model"`
		} `json:"provider_history"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Summary.Sessions != 1 || body.Summary.Messages != 2 || body.Usage.Input != 4 || body.Usage.Output != 6 || body.Usage.Total != 10 || !body.Usage.Available || body.Cost.Available {
		t.Fatalf("unexpected insights: %#v", body)
	}
	if len(body.ProviderHistory) != 1 || body.ProviderHistory[0].Provider != "local" || body.ProviderHistory[0].Model != "qwen" {
		t.Fatalf("unexpected provider history: %#v", body.ProviderHistory)
	}
}

func TestOperatorDiagnosticsIsReadOnlyAndSanitized(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", GatewayAPIKey: "do-not-leak", StateDir: t.TempDir()}
	server := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL, APIKey: cfg.GatewayAPIKey}))
	if _, err := server.control.Create("todos", control.Item{Title: "Review projection"}); err != nil {
		t.Fatal(err)
	}
	if _, err := server.sessions.Create("diagnostic-session", "Safe session", nil); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/operator/diagnostics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("diagnostics=%d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Status     string `json:"status"`
		ReadOnly   bool   `json:"read_only"`
		Components map[string]struct {
			Available bool `json:"available"`
		} `json:"components"`
		Counts struct {
			Sessions    int            `json:"sessions"`
			Collections map[string]int `json:"collections"`
		} `json:"counts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ready" || !body.ReadOnly || !body.Components["gateway"].Available || !body.Components["workspace"].Available || body.Counts.Sessions != 1 || body.Counts.Collections["todos"] != 1 {
		t.Fatalf("unexpected diagnostics: %#v", body)
	}
	if strings.Contains(rec.Body.String(), "do-not-leak") || (cfg.WorkspaceRoot != "" && strings.Contains(rec.Body.String(), cfg.WorkspaceRoot)) {
		t.Fatalf("diagnostics leaked sensitive data: %s", rec.Body.String())
	}
}

func TestOperatorVersionAndUpdateAreReadOnlyContracts(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir()}
	server := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL}))

	for _, path := range []string{"/api/operator/version", "/api/operator/update"} {
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s=%d %s", path, rec.Code, rec.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s invalid JSON: %v", path, err)
		}
		if body["read_only"] != true {
			t.Fatalf("%s did not declare read_only: %s", path, rec.Body.String())
		}
	}

	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/operator/update", nil))
	if !strings.Contains(rec.Body.String(), `"available":false`) || !strings.Contains(rec.Body.String(), "service supervisor") {
		t.Fatalf("update contract exposed an optimistic action: %s", rec.Body.String())
	}
}
