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
