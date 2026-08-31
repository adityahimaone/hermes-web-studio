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

func TestMCPServerCRUDValidatesEndpointsAndFiltersSecrets(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir()}
	h := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()
	request := httptest.NewRequest(http.MethodPost, "/api/mcp/servers", strings.NewReader(`{"name":"Docs","transport":"http","endpoint":"https://mcp.example.test/tools","settings":{"region":"us","api_token":"never-store"},"tools":[{"name":"search"}]}`))
	request.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, request)
	if rec.Code != http.StatusCreated || strings.Contains(rec.Body.String(), "never-store") || !strings.Contains(rec.Body.String(), `"status":"configured"`) {
		t.Fatalf("create=%d body=%s", rec.Code, rec.Body.String())
	}
	var created struct {
		ID       string            `json:"id"`
		Settings map[string]string `json:"settings"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.ID == "" || created.Settings["api_token"] != "" || created.Settings["region"] != "us" {
		t.Fatalf("created=%#v err=%v", created, err)
	}
	bad := httptest.NewRequest(http.MethodPost, "/api/mcp/servers", strings.NewReader(`{"name":"Unsafe","transport":"http","endpoint":"file:///tmp/socket"}`))
	bad.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, bad)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid endpoint=%d body=%s", rec.Code, rec.Body.String())
	}
	get := httptest.NewRecorder()
	h.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/mcp/servers", nil))
	if get.Code != http.StatusOK || strings.Contains(get.Body.String(), "never-store") || !strings.Contains(get.Body.String(), "Docs") {
		t.Fatalf("list=%d body=%s", get.Code, get.Body.String())
	}
	tools := httptest.NewRecorder()
	h.ServeHTTP(tools, httptest.NewRequest(http.MethodGet, "/api/mcp/servers/"+created.ID+"/tools", nil))
	if tools.Code != http.StatusOK || !strings.Contains(tools.Body.String(), "search") || !strings.Contains(tools.Body.String(), `"read_only":true`) {
		t.Fatalf("tools=%d body=%s", tools.Code, tools.Body.String())
	}
	del := httptest.NewRecorder()
	h.ServeHTTP(del, httptest.NewRequest(http.MethodDelete, "/api/mcp/servers/"+created.ID, nil))
	if del.Code != http.StatusOK {
		t.Fatalf("delete=%d body=%s", del.Code, del.Body.String())
	}
}

func TestPluginSettingsAreServerOwnedAndSecretFiltered(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir()}
	h := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()
	// A plugin must be discovered by the server before its settings can be changed.
	if err := os.MkdirAll(filepath.Join(cfg.StateDir, "plugins"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.StateDir, "plugins", "demo.plugin"), []byte("plugin"), 0600); err != nil {
		t.Fatal(err)
	}
	get := httptest.NewRecorder()
	h.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/plugins", nil))
	if get.Code != http.StatusOK || !strings.Contains(get.Body.String(), "demo.plugin") {
		t.Fatalf("list=%d body=%s", get.Code, get.Body.String())
	}
	update := httptest.NewRequest(http.MethodPatch, "/api/plugins/demo.plugin", strings.NewReader(`{"enabled":false,"settings":{"theme":"dark","password":"hidden"}}`))
	update.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, update)
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "hidden") || !strings.Contains(rec.Body.String(), "dark") {
		t.Fatalf("plugin update=%d body=%s", rec.Code, rec.Body.String())
	}
}
