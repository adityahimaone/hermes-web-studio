package httpapi

import (
	"encoding/json"
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
