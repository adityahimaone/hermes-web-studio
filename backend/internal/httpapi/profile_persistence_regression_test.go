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

func TestMalformedPersistedProfilesFailClosedToSafeDefault(t *testing.T) {
	dir := t.TempDir()
	data := `{"profiles":[{"id":"bad/id","name":"Bad"}],"active":"bad/id"}`
	if err := os.WriteFile(filepath.Join(dir, "profiles.json"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	s := NewWithGateway(config.Config{StateDir: dir}, gateway.New(gateway.Config{}))
	if len(s.profiles) != 1 || s.profiles[0].ID != "default" || s.activeProfile != "default" {
		t.Fatalf("not safe default: %+v %q", s.profiles, s.activeProfile)
	}
}

func TestProfileUpdateRestoresStateWhenPersistenceFails(t *testing.T) {
	dir := t.TempDir()
	s := NewWithGateway(config.Config{StateDir: dir}, gateway.New(gateway.Config{}))
	s.profilePath = filepath.Join(dir, "missing", "profiles.json")
	req := httptest.NewRequest(http.MethodPatch, "/api/profiles/default", strings.NewReader(`{"id":"default","name":"Changed"}`))
	req.SetPathValue("id", "default")
	rec := httptest.NewRecorder()
	s.handleProfileUpdate(rec, req)
	if rec.Code != http.StatusInternalServerError || s.profiles[0].Name != "Default" {
		t.Fatalf("rollback failed: code=%d profiles=%+v", rec.Code, s.profiles)
	}
}

func TestSpacesAndPreferencesStayProfileLocalAcrossRecreation(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{StateDir: dir, WorkspaceRoot: dir}
	first := NewWithGateway(cfg, gateway.New(gateway.Config{}))
	create := httptest.NewRecorder()
	first.handleSpaceCreate(create, httptest.NewRequest(http.MethodPost, "/api/spaces", strings.NewReader(`{"name":"A","location_kind":"local","workspace_ref":"."}`)))
	if create.Code != http.StatusCreated {
		t.Fatalf("create A=%d %s", create.Code, create.Body.String())
	}
	pref := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/preferences", strings.NewReader(`{"theme":"dark"}`))
	req.Header.Set("Content-Type", "application/json")
	first.handlePreferencesUpdate(pref, req)
	if pref.Code != http.StatusOK {
		t.Fatalf("pref A=%d", pref.Code)
	}
	other := httptest.NewRecorder()
	first.handleProfileCreate(other, httptest.NewRequest(http.MethodPost, "/api/profiles", strings.NewReader(`{"id":"other","name":"Other"}`)))
	if other.Code != http.StatusCreated {
		t.Fatalf("profile=%d", other.Code)
	}
	first.activeProfile = "other"
	spaces := httptest.NewRecorder()
	first.handleSpaces(spaces, httptest.NewRequest(http.MethodGet, "/api/spaces", nil))
	if strings.Contains(spaces.Body.String(), "A") {
		t.Fatalf("profile leak spaces: %s", spaces.Body.String())
	}
	preferences := httptest.NewRecorder()
	first.handlePreferences(preferences, httptest.NewRequest(http.MethodGet, "/api/preferences", nil))
	if strings.Contains(preferences.Body.String(), "dark") {
		t.Fatalf("profile leak preferences: %s", preferences.Body.String())
	}
	second := NewWithGateway(cfg, gateway.New(gateway.Config{}))
	second.activeProfile = "default"
	got := httptest.NewRecorder()
	second.handlePreferences(got, httptest.NewRequest(http.MethodGet, "/api/preferences", nil))
	var payload map[string]map[string]string
	_ = json.Unmarshal(got.Body.Bytes(), &payload)
	if payload["preferences"]["theme"] != "dark" {
		t.Fatalf("default preference missing: %s", got.Body.String())
	}
}
