package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

func TestControlCenterCRUDDoesNotExposeGatewaySecrets(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", GatewayAPIKey: "secret", StateDir: t.TempDir()}
	h := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL, APIKey: cfg.GatewayAPIKey})).Handler()
	req := httptest.NewRequest(http.MethodPost, "/api/control/todos", strings.NewReader(`{"title":"Review workspace","description":"safe"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("create=%d body=%s", rec.Code, rec.Body.String())
	}
	req = httptest.NewRequest(http.MethodGet, "/api/control/todos", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Review workspace") {
		t.Fatalf("list=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestControlListsSerializeEmptyCollectionsAsArrays(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir()}
	h := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/control/tasks", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"items":[]`) {
		t.Fatalf("list=%d body=%s", rec.Code, rec.Body.String())
	}
}
