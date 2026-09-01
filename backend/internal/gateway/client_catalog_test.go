package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestModelsReturnsSanitizedCatalog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" || r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("request path=%q auth=%q", r.URL.Path, r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{
			map[string]any{"id": "gpt-test", "owned_by": "openai"},
			map[string]any{"id": "claude-test", "provider": "anthropic"},
		}})
	}))
	defer server.Close()

	models, err := New(Config{BaseURL: server.URL, APIKey: "secret"}).Models(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0].ID != "gpt-test" || models[0].Provider != "openai" {
		t.Fatalf("models=%+v", models)
	}
	if models[1].Provider != "anthropic" {
		t.Fatalf("provider=%q", models[1].Provider)
	}
}

func TestModelsRejectsUnavailableGateway(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	if _, err := New(Config{BaseURL: server.URL}).Models(context.Background()); err == nil {
		t.Fatal("expected unavailable gateway error")
	}
}
