package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestModelsPreservesSameIDAcrossProviders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{
			map[string]any{"id": "same", "provider": "openai"},
			map[string]any{"id": "same", "provider": "anthropic"},
		}})
	}))
	defer server.Close()
	models, err := New(Config{BaseURL: server.URL}).Models(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0].Provider != "openai" || models[1].Provider != "anthropic" {
		t.Fatalf("models=%+v", models)
	}
}

func TestModelsRejectsUnavailableGateway(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	if _, err := New(Config{BaseURL: server.URL}).Models(context.Background()); err == nil {
		t.Fatal("expected unavailable gateway error")
	}
}

func TestModelsRejectsCatalogOverMaximumItemCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		items := make([]map[string]string, maxCatalogModels+1)
		for i := range items {
			items[i] = map[string]string{"id": "model-" + string(rune('a'+i%26)) + strings.Repeat("x", i/26)}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": items})
	}))
	defer server.Close()
	if _, err := New(Config{BaseURL: server.URL}).Models(context.Background()); err == nil {
		t.Fatal("expected catalog item limit error")
	}
}

func TestModelsRejectsResponseBodyOverOneMiB(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}` + strings.Repeat(" ", (1<<20))))
	}))
	defer server.Close()
	if _, err := New(Config{BaseURL: server.URL}).Models(context.Background()); err == nil {
		t.Fatal("expected oversized catalog body error")
	}
}
