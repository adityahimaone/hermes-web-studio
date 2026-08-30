package httpapi

import (
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

func TestWorkspaceAPIRejectsTraversalAndSupportsPreviewAndWrite(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.md"), []byte("# hello"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), WorkspaceRoot: root}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()
	response, err := http.Get(api.URL + "/api/workspace/preview?path=hello.md")
	if err != nil {
		t.Fatal(err)
	}
	var preview map[string]any
	decode(t, response.Body, &preview)
	response.Body.Close()
	if response.StatusCode != http.StatusOK || preview["content"] != "# hello" {
		t.Fatalf("preview=%#v status=%d", preview, response.StatusCode)
	}
	response, err = http.Get(api.URL + "/api/workspace/tree?path=../")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("traversal status=%d", response.StatusCode)
	}
	request := httptest.NewRequest(http.MethodPut, api.URL+"/api/workspace/file", nil)
	request.Body = io.NopCloser(strings.NewReader(`{"path":"notes.txt","content":"safe"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	api.Config.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("write status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if data, err := os.ReadFile(filepath.Join(root, "notes.txt")); err != nil || string(data) != "safe" {
		t.Fatalf("written=%q err=%v", data, err)
	}
}
