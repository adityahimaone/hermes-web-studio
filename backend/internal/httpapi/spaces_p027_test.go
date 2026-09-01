package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/control"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

func TestSpacesP027RegistryLifecycleAndRemoteState(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), WorkspaceRoot: t.TempDir()}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()
	create := func(body string) (int, map[string]any) {
		response, err := http.Post(api.URL+"/api/spaces", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var payload map[string]any
		if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		return response.StatusCode, payload
	}
	if status, _ := create(`{"name":"Remote Mac","location_kind":"remote","workspace_ref":"macbook-project","order":2}`); status != http.StatusCreated {
		t.Fatalf("remote create status=%d", status)
	}
	status, local := create(`{"name":"Local","location_kind":"local","workspace_ref":"project","order":1}`)
	if status != http.StatusCreated {
		t.Fatalf("local create status=%d", status)
	}
	if local["id"] == "" || local["health"] != "ready" || local["profile_id"] != "default" {
		t.Fatalf("local=%#v", local)
	}
	response, err := http.Get(api.URL + "/api/spaces")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var listed struct {
		Active string           `json:"active"`
		Spaces []map[string]any `json:"spaces"`
	}
	if err := json.NewDecoder(response.Body).Decode(&listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Spaces) != 2 || listed.Spaces[0]["name"] != "Local" {
		t.Fatalf("listed=%#v", listed)
	}
	if status, _ := create(`{"name":"bad","location_kind":"remote","workspace_ref":""}`); status != http.StatusBadRequest {
		t.Fatalf("invalid remote status=%d", status)
	}
	request, _ := http.NewRequest(http.MethodPost, api.URL+"/api/spaces/active", bytes.NewBufferString(`{"id":"missing"}`))
	request.Header.Set("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("activate missing status=%d", response.StatusCode)
	}
}

func TestSpacesP027ProtectsActiveDeletion(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), WorkspaceRoot: t.TempDir()}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()
	request, _ := http.NewRequest(http.MethodPost, api.URL+"/api/spaces", bytes.NewBufferString(`{"name":"Local","location_kind":"local","workspace_ref":"project"}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	var item map[string]any
	json.NewDecoder(response.Body).Decode(&item)
	response.Body.Close()
	request, _ = http.NewRequest(http.MethodDelete, api.URL+"/api/spaces/"+item["id"].(string), nil)
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("delete active status=%d", response.StatusCode)
	}
}

func TestSpacesP027RejectsUnsafeAndUnauthorizedAssignments(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), WorkspaceRoot: t.TempDir(), ProfilesJSON: `[{"id":"default","name":"Default"},{"id":"other","name":"Other"}]`}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()
	for _, body := range []string{
		`{"name":"bad","location_kind":"local","workspace_ref":"../secret"}`, `{"name":"bad","location_kind":"local","workspace_ref":"/etc"}`, `{"name":"bad","location_kind":"local","workspace_ref":"file://tmp"}`, `{"name":"bad","location_kind":"local","workspace_ref":"ssh://host/path"}`, `{"name":"bad","location_kind":"local","workspace_ref":"~/.ssh"}`, `{"name":"bad","location_kind":"local","workspace_ref":"safe:thing"}`, `{"name":"bad","location_kind":"remote","workspace_ref":"ssh://host/path"}`, `{"name":"bad","location_kind":"local","workspace_ref":"safe","profile_id":"unknown"}`, `{"name":"bad","location_kind":"local","workspace_ref":"safe","profile_id":"other"}`,
	} {
		response, err := http.Post(api.URL+"/api/spaces", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("body=%s status=%d", body, response.StatusCode)
		}
	}
}

func TestSpacesP027ResolvesSymlinkBeforeContainment(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), WorkspaceRoot: root}
	server := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL}))
	if safeSpaceRef(server, "local", "link") {
		t.Fatal("symlink escaping workspace accepted")
	}
	if !safeSpaceRef(server, "local", "missing/nested") {
		t.Fatal("safe relative reference rejected")
	}
}

func TestSpacesP027RejectsMissingDescendantBehindSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), WorkspaceRoot: root}
	server := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL}))
	if safeSpaceRef(server, "local", "link/missing/nested") {
		t.Fatal("missing descendant behind escaping symlink accepted")
	}
}

func TestSpacesP027RejectsUnsafeRemoteWorkspaceRefs(t *testing.T) {
	for _, ref := range []string{"-oProxyCommand=bad", "user@host:22", "host/path", "host;id", "host\ncommand", "host port"} {
		if validateRemoteWorkspaceRef(ref) {
			t.Fatalf("unsafe remote ref accepted: %q", ref)
		}
	}
	for _, ref := range []string{"macbook-project", "host_2", "host-2", "host.example"} {
		if !validateRemoteWorkspaceRef(ref) {
			t.Fatalf("safe remote ref rejected: %q", ref)
		}
	}
}

func TestSpacesP027SanitizesLegacyReferences(t *testing.T) {
	item := control.Item{ID: "space-1", Metadata: map[string]string{"location_kind": "local", "workspace_ref": "file:///secret"}}
	if got := sanitizeSpaceRef(item); got == "file:///secret" || got == "" {
		t.Fatalf("unsafe ref leaked: %q", got)
	}
	remote := control.Item{ID: "space-1", Metadata: map[string]string{"location_kind": "remote", "workspace_ref": "ssh://host/private"}}
	if got := sanitizeSpaceRef(remote); got != "remote:space-1" {
		t.Fatalf("remote ref=%q", got)
	}
}

func TestSpacesGETDoesNotPersistActiveMetadata(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), WorkspaceRoot: t.TempDir()}
	server := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL}))
	item, err := server.control.Create("spaces", control.Item{Title: "space", Metadata: map[string]string{"profile_id": "default"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := server.control.SetPreferences(map[string]string{"active_space:default": item.ID}); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/spaces", nil)
	rec := httptest.NewRecorder()
	server.handleSpaces(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	items, err := server.control.List("spaces")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := items[0].Metadata["active"]; ok {
		t.Fatal("GET persisted active metadata")
	}
}

func TestSpacesP027UsesNumericStableOrder(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), WorkspaceRoot: t.TempDir()}
	server := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL}))
	one, _ := server.control.Create("spaces", control.Item{Title: "ten", Metadata: map[string]string{"location_kind": "local", "workspace_ref": "ten", "order": "10", "profile_id": "default"}})
	two, _ := server.control.Create("spaces", control.Item{Title: "two", Metadata: map[string]string{"location_kind": "local", "workspace_ref": "two", "order": "2", "profile_id": "default"}})
	if got := spacePayloads([]control.Item{one, two}, "")[0]["id"]; got != two.ID {
		t.Fatalf("numeric order got=%v", got)
	}
}
