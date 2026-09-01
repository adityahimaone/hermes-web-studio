package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestModelsRejectsUnsafeBaseURLBeforeRequest(t *testing.T) {
	for _, baseURL := range []string{"file:///tmp/gateway", "http://user:pass@127.0.0.1:8642", "http://127.0.0.1:8642/path?secret=1", "http://10.0.0.1:8642"} {
		if _, err := New(Config{BaseURL: baseURL}).Models(context.Background()); err == nil {
			t.Fatalf("expected rejection for %q", baseURL)
		}
	}
}

func TestAllGatewayRequestPathsRejectUnsafeBaseURL(t *testing.T) {
	client := New(Config{BaseURL: "http://10.0.0.1:8642", ReadTimeout: time.Millisecond})
	if err := client.Health(context.Background()); err == nil {
		t.Fatal("Health accepted unsafe base URL")
	}
	if _, err := client.Stream(context.Background(), ChatRequest{Message: "hello"}, func(Event) {}); err == nil {
		t.Fatal("Stream accepted unsafe base URL")
	}
	if _, err := client.RunStream(context.Background(), ChatRequest{Message: "hello"}, func(Event) {}); err == nil {
		t.Fatal("RunStream accepted unsafe base URL")
	}
	if err := client.ResolveApproval(context.Background(), "run-1", "deny"); err == nil {
		t.Fatal("ResolveApproval accepted unsafe base URL")
	}
}

func TestCleanCatalogTextRemovesC1ControlsAndPreservesUTF8(t *testing.T) {
	value := cleanCatalogText(" model-\u0085\u009f-🙂 ", 256)
	if value != "model--🙂" || !utf8.ValidString(value) {
		t.Fatalf("cleaned=%q valid=%v", value, utf8.ValidString(value))
	}
	truncated := cleanCatalogText("🙂🙂🙂", 8)
	if truncated != "🙂🙂" || !utf8.ValidString(truncated) {
		t.Fatalf("truncated=%q valid=%v", truncated, utf8.ValidString(truncated))
	}
}

func TestModelsDoesNotFollowUnsafeRedirect(t *testing.T) {
	internalHit := false
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { internalHit = true; http.NotFound(w, nil) }))
	defer internal.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, internal.URL+"/v1/models", http.StatusFound)
	}))
	defer redirect.Close()
	_, _ = New(Config{BaseURL: redirect.URL}).Models(context.Background())
	if internalHit {
		t.Fatal("model request followed redirect")
	}
}

func TestModelsNormalizesAndDeduplicatesCatalog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{
			map[string]any{"id": " model-a ", "provider": " openai ", "aliases": []string{" fast\n", "fast", "  "}},
			map[string]any{"id": "model-a", "provider": "openai"},
			map[string]any{"id": "model-b\n", "provider": "bad\rprovider", "aliases": []string{"x"}},
		}})
	}))
	defer server.Close()
	models, err := New(Config{BaseURL: server.URL}).Models(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0].ID != "model-a" || len(models[0].Aliases) != 1 || models[1].ID != "model-b" || strings.ContainsAny(models[1].Provider, "\r\n") {
		t.Fatalf("models=%+v", models)
	}
}
