package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

func TestSessionsAPIDoesNotProjectPrivateMetadata(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stateDir, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	data := `{"session_id":"session-safe","title":"Safe","project":"studio","workspace_path":"/srv/private/project","metadata":{"api_key":"secret"},"messages":[]}`
	if err := os.WriteFile(filepath.Join(stateDir, "sessions", "session-safe.json"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: stateDir}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()
	response, err := http.Get(api.URL + "/api/sessions")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(body)
	text := string(encoded)
	if strings.Contains(text, "workspace_path") || strings.Contains(text, "/srv/private/project") || strings.Contains(text, "api_key") || strings.Contains(text, "secret") {
		t.Fatalf("unsafe metadata projected: %s", text)
	}
	if !strings.Contains(text, `"project":"studio"`) {
		t.Fatalf("safe metadata missing: %s", text)
	}
}
