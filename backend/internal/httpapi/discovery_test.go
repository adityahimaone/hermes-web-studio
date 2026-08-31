package httpapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

func TestDiscoveryReadsHermesSkillsAndMemory(t *testing.T) {
	hermesHome := t.TempDir()
	skillDir := filepath.Join(hermesHome, "skills", "sample")
	if err := os.MkdirAll(skillDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: sample\ndescription: A sample skill\n---\ncontent\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(hermesHome, "memories"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hermesHome, "memories", "MEMORY.md"), []byte("memory"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), HermesHome: hermesHome}
	h := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()
	for path, want := range map[string]string{"/api/skills": "sample", "/api/memory": "MEMORY.md"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("%s=%d body=%s", path, rec.Code, rec.Body.String())
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/skills?name=sample%2FSKILL.md", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "content") {
		t.Fatalf("skill content=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSkillsSearchAndCRUDStayInsideHermesHome(t *testing.T) {
	hermesHome := t.TempDir()
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), HermesHome: hermesHome}
	h := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()
	create := httptest.NewRequest(http.MethodPost, "/api/skills", strings.NewReader(`{"name":"release-notes","content":"# Release notes"}`))
	create.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, create)
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), "release-notes/SKILL.md") {
		t.Fatalf("create=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/skills?q=release", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "release-notes") {
		t.Fatalf("search=%d body=%s", rec.Code, rec.Body.String())
	}
	update := httptest.NewRequest(http.MethodPut, "/api/skills", strings.NewReader(`{"name":"release-notes/SKILL.md","content":"# Updated"}`))
	update.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, update)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Updated") {
		t.Fatalf("update=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/skills?name=release-notes%2FSKILL.md", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("delete=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/skills", strings.NewReader(`{"name":"../escape/SKILL.md","content":"unsafe"}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unsafe update=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMemoryCRUDStaysInsideMemoryRoot(t *testing.T) {
	hermesHome := t.TempDir()
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), HermesHome: hermesHome}
	h := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()
	request := httptest.NewRequest(http.MethodPost, "/api/memory", strings.NewReader(`{"name":"USER.md","content":"prefers concise replies"}`))
	request.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, request)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "concise") {
		t.Fatalf("create=%d body=%s", rec.Code, rec.Body.String())
	}
	request = httptest.NewRequest(http.MethodPatch, "/api/memory", strings.NewReader(`{"name":"USER.md","content":"updated"}`))
	request.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, request)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "updated") {
		t.Fatalf("update=%d body=%s", rec.Code, rec.Body.String())
	}
	request = httptest.NewRequest(http.MethodPost, "/api/memory", strings.NewReader(`{"name":"../escape.md","content":"unsafe"}`))
	request.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, request)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unsafe=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTerminalCapabilityIsExplicitlyUnavailableWithoutProcessRoutes(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), HermesHome: t.TempDir()}
	h := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/terminal", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"available":false`) || !strings.Contains(rec.Body.String(), `"reason":"sandbox_required"`) {
		t.Fatalf("capability=%d body=%s", rec.Code, rec.Body.String())
	}

	for _, path := range []string{"/api/terminal/start", "/api/terminal/input", "/api/terminal/output", "/api/terminal/resize", "/api/terminal/close"} {
		rec = httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`)))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("unsafe process route %s returned %d body=%s", path, rec.Code, rec.Body.String())
		}
	}
}

func TestExtensionAndPluginContractsAreReadOnlyAndSanitized(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stateDir, "extensions", "safe-extension"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(stateDir, "plugins"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "plugins", "plugin.json"), []byte(`{"api_key":"must-not-be-read"}`), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: stateDir, HermesHome: t.TempDir()}
	h := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()

	for _, path := range []string{"/api/extensions", "/api/plugins"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"read_only":true`) || strings.Contains(rec.Body.String(), "must-not-be-read") {
			t.Fatalf("%s=%d body=%s", path, rec.Code, rec.Body.String())
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/extensions", nil))
	if !strings.Contains(rec.Body.String(), "safe-extension") || !strings.Contains(rec.Body.String(), "metadata-only") {
		t.Fatalf("extension registry=%s", rec.Body.String())
	}
}
