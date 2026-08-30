package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

func TestPasswordAuthProtectsAPIAndSetsHttpOnlyCookie(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir()}
	handler := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()
	request := httptest.NewRequest(http.MethodPost, "/api/onboarding/password", strings.NewReader(`{"password":"correct horse battery"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("setup=%d %s", recorder.Code, recorder.Body.String())
	}
	request = httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("protected=%d", recorder.Code)
	}
	request = httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"password":"correct horse battery"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Header().Get("Set-Cookie"), "HttpOnly") {
		t.Fatalf("login=%d cookie=%q", recorder.Code, recorder.Header().Get("Set-Cookie"))
	}
	var payload map[string]any
	_ = json.Unmarshal(recorder.Body.Bytes(), &payload)
}

func TestProfileAndProviderCRUDRedactsCredentials(t *testing.T) {
	api := newTestServer(t, "http://127.0.0.1:1", "")
	defer api.Close()
	profile := postJSON(t, api.URL+"/api/profiles", map[string]any{"id": "writing", "name": "Writing", "model": "q4", "provider_id": "custom"})
	if profile.StatusCode != http.StatusCreated {
		t.Fatalf("profile create=%d", profile.StatusCode)
	}
	_ = profile.Body.Close()
	provider := postJSON(t, api.URL+"/api/providers", map[string]any{"id": "custom", "name": "Local", "kind": "openai-compatible", "base_url": "http://127.0.0.1:9000/v1", "api_key": "do-not-return"})
	if provider.StatusCode != http.StatusCreated {
		t.Fatalf("provider create=%d", provider.StatusCode)
	}
	body, _ := io.ReadAll(provider.Body)
	_ = provider.Body.Close()
	if strings.Contains(string(body), "do-not-return") || !strings.Contains(string(body), `"has_key":true`) {
		t.Fatalf("provider leaked credential: %s", body)
	}
	response, err := api.Client().Get(api.URL + "/api/settings/capabilities")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("capabilities=%d", response.StatusCode)
	}
}
