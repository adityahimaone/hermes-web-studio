package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

func TestSafeExportValueRedactsCredentialVariantsButPreservesFreeFormData(t *testing.T) {
	value := map[string]any{
		"authorization_header": "Bearer secret",
		"auth_token_value":     "token-secret",
		"api_key":              "key-secret",
		"password_hint":        "password-secret",
		"credential_id":        "credential-secret",
		"directory_name":       "keep this free-form value",
		"message":              "auth token key secret password credential in prose",
	}
	out := safeExportValue(value).(map[string]any)
	for _, key := range []string{"authorization_header", "auth_token_value", "api_key", "password_hint", "credential_id"} {
		if _, ok := out[key]; ok {
			t.Fatalf("sensitive key preserved: %q", key)
		}
	}
	if out["directory_name"] != "keep this free-form value" || out["message"] != value["message"] {
		t.Fatalf("free-form data changed: %#v", out)
	}
}

func TestSafeExportValueRedactsAllPrivatePathForms(t *testing.T) {
	for _, path := range []string{"/tmp/file", "~/file", `C:\\Users\\name\\file`, `\\\\server\\share\\file`} {
		if got := safeExportValue(path); got != "[redacted]" {
			t.Fatalf("path %q got %#v", path, got)
		}
	}
	if got := safeExportValue("ordinary message with / slash"); got != "ordinary message with / slash" {
		t.Fatalf("free-form text changed: %#v", got)
	}
}

func TestJSONExportRejectsMalformedPersistedMessage(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stateDir, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	data := `{"session_id":"broken","title":"Broken","messages":[{"role":"user","content":"kept"}, "malformed message"]}`
	if err := os.WriteFile(filepath.Join(stateDir, "sessions", "broken.json"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: stateDir}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()

	response, err := http.Get(api.URL + "/api/sessions/broken/export?format=json")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status=%d", response.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "export_failed" {
		t.Fatalf("body=%#v", body)
	}
}

func TestSensitiveExportKeyDoesNotRedactUnrelatedKeyNames(t *testing.T) {
	for _, key := range []string{"keyboard", "directory_name"} {
		if sensitiveExportKey(key) {
			t.Fatalf("unrelated key redacted: %q", key)
		}
	}
}
