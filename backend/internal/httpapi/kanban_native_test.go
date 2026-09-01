package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

func TestExtractJSONIgnoresCLIStatusText(t *testing.T) {
	value, err := extractJSON([]byte("notice: starting\n[{\"id\":\"t_1\"}]\n"))
	if err != nil {
		t.Fatal(err)
	}
	var items []map[string]string
	encoded, _ := json.Marshal(value)
	if err := json.Unmarshal(encoded, &items); err != nil || items[0]["id"] != "t_1" {
		t.Fatalf("value=%s err=%v", encoded, err)
	}
}

func TestExtractJSONRejectsTrailingData(t *testing.T) {
	if _, err := extractJSON([]byte(`{"ok":true} trailing`)); err == nil {
		t.Fatal("trailing data accepted")
	}
}

func TestNativeKanbanActionRejectsUnknownActionBeforeCLI(t *testing.T) {
	server := nativeCLITestServer(t)
	defer server.Close()
	response, err := http.Post(server.URL+"/api/kanban/tasks/t_1/actions/unknown", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d", response.StatusCode)
	}
}

func nativeCLITestServer(t *testing.T) *httptest.Server {
	t.Helper()
	path := filepath.Join(t.TempDir(), "hermes")
	script := `#!/bin/sh
case "$*" in
  *"stats --json"*) printf '%s' '{"ready":1,"done":2}' ;;
  *"boards list"*) printf '%s' '[{"slug":"default","name":"Default","task_count":3}]' ;;
  *"list --json"*) printf '%s' '[{"id":"t_1","title":"Ship board","status":"ready","priority":2,"comment_count":1,"parents":[],"children":[],"workspace":"/srv/private/task","secret":"secret-token","subprocess":"/usr/bin/hermes"}]' ;;
  *"create"*) printf '%s' '{"id":"t_new","title":"Created","status":"ready"}' ;;
  *"complete"*) printf '%s' 'completed /srv/private/task\nsubprocess: /usr/bin/hermes --api-key secret-token\n{"ok":true}' ;;
  *) printf '%s' '{"ok":true}' ;;
esac
`
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{HermesCLIPath: path, StateDir: t.TempDir(), GatewayBaseURL: "http://127.0.0.1:1"}
	return httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
}

func TestNativeKanbanBoardUsesCLIJSON(t *testing.T) {
	server := nativeCLITestServer(t)
	defer server.Close()
	response, err := http.Get(server.URL + "/api/kanban/board?board=default")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", response.StatusCode)
	}
	body, _ := io.ReadAll(response.Body)
	if !strings.Contains(string(body), "t_1") {
		t.Fatalf("body=%s", body)
	}
}

func TestNativeKanbanBoardSanitizesCLIFields(t *testing.T) {
	server := nativeCLITestServer(t)
	defer server.Close()
	response, err := http.Get(server.URL + "/api/kanban/board?board=default")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	for _, secret := range []string{"/srv/private/task", "secret-token", "subprocess", "environment"} {
		if strings.Contains(string(body), secret) {
			t.Fatalf("private CLI field leaked: %q in %s", secret, body)
		}
	}
	if !strings.Contains(string(body), "t_1") {
		t.Fatalf("safe task missing: %s", body)
	}
}

func TestNativeKanbanCreateDoesNotExposeCLIPath(t *testing.T) {
	server := nativeCLITestServer(t)
	defer server.Close()
	response, err := http.Post(server.URL+"/api/kanban/tasks?board=default", "application/json", strings.NewReader(`{"title":"Created"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("status=%d", response.StatusCode)
	}
	body, _ := io.ReadAll(response.Body)
	if strings.Contains(string(body), "hermes") {
		t.Fatalf("CLI path leaked: %s", body)
	}
}

func TestNativeKanbanActionOmitsPrivateCLIOutput(t *testing.T) {
	server := nativeCLITestServer(t)
	defer server.Close()
	request, err := http.NewRequest(http.MethodPost, server.URL+"/api/kanban/tasks/t_1/actions/complete?board=default", bytes.NewBufferString(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusOK || string(body) != "{\"action\":\"complete\",\"ok\":true,\"task_id\":\"t_1\"}\n" {
		t.Fatalf("status=%d body=%s", response.StatusCode, body)
	}
	for _, secret := range []string{"/srv/private/task", "/usr/bin/hermes", "secret-token", "subprocess"} {
		if strings.Contains(string(body), secret) {
			t.Fatalf("private CLI detail leaked: %q in %s", secret, body)
		}
	}
}

func TestNativeKanbanCapabilitiesGateUnsupportedFeatures(t *testing.T) {
	server := nativeCLITestServer(t)
	defer server.Close()
	response, err := http.Get(server.URL + "/api/kanban/capabilities")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if !strings.Contains(string(body), `"transport":"cli"`) || !strings.Contains(string(body), `"live_updates":false`) || !strings.Contains(string(body), `"arbitrary_edit":false`) {
		t.Fatalf("capabilities=%s", body)
	}
}

func TestKanbanWorkspaceValueRejectsRemoteAndUnsafeReferences(t *testing.T) {
	for _, value := range []string{"dir:Remote workspace unavailable", "dir:../secret", "dir:/tmp/private", "dir:ssh://host/path", "dir:~/.ssh"} {
		cfg := config.Config{WorkspaceRoot: t.TempDir()}
		server := NewWithGateway(cfg, nil)
		if _, ok := server.canonicalKanbanWorkspace(value); ok {
			t.Fatalf("unsafe workspace accepted: %q", value)
		}
	}
}

func TestSanitizeKanbanDetailProjectsSafeActivityFields(t *testing.T) {
	got := sanitizeKanbanProjection(map[string]any{
		"id": "t_1", "workspace": "dir:project", "comments": []any{map[string]any{"id": "c1", "body": "safe", "secret": "no"}},
		"events": []any{map[string]any{"type": "status", "status": "done", "private_path": "/x"}},
		"runs":   []any{map[string]any{"id": "r1", "status": "done", "result": "ok", "subprocess": "bad"}}, "result": "done",
	})
	body, _ := json.Marshal(got)
	text := string(body)
	for _, want := range []string{`"workspace":"dir:project"`, `"comments"`, `"events"`, `"runs"`, `"result":"done"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %s: %s", want, text)
		}
	}
	for _, bad := range []string{"secret", "private_path", "subprocess", "/x"} {
		if strings.Contains(text, bad) {
			t.Fatalf("sensitive field leaked: %s", text)
		}
	}
}

func TestSanitizeKanbanDetailRejectsUnsafeActivityValues(t *testing.T) {
	got := sanitizeKanbanProjection(map[string]any{"workspace": "/srv/private", "result": strings.Repeat("x", 5000), "comments": []any{map[string]any{"body": strings.Repeat("x", 5000)}}})
	body, _ := json.Marshal(got)
	if strings.Contains(string(body), "/srv/private") || len(body) > 4096 {
		t.Fatalf("unsafe or unbounded detail: %d %s", len(body), body)
	}
}
