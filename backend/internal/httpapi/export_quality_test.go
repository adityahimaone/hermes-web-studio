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
		"apiKey":               "key-secret",
		"private_key":          "private-secret",
		"credential":           "credential-secret",
		"accessKey":            "access-secret",
		"APIKEY":               "key-secret",
		"privateKey":           "private-secret",
		"authorization_header": "Bearer secret",
		"auth_token_value":     "token-secret",
		"secret_value":         "secret-secret",
		"password_hint":        "password-secret",
		"publicKey":            "public-camel-secret",
		"public_key":           "public-underscore-secret",
		"publicKEY":            "public-acronym-secret",
		"clientKey":            "client-camel-secret",
		"client_key":           "client-underscore-secret",
		"serverKey":            "server-camel-secret",
		"server_key":           "server-underscore-secret",
		"encryptionKey":        "encryption-camel-secret",
		"encryption_key":       "encryption-underscore-secret",
		"signingKey":           "signing-camel-secret",
		"signing_key":          "signing-underscore-secret",
		"authKey":              "auth-camel-secret",
		"auth_key":             "auth-underscore-secret",
		"secretKey":            "secret-camel-secret",
		"secret_key":           "secret-underscore-secret",
		"author":               "keep author",
		"tokenizer":            "keep tokenizer",
		"directory_name":       "keep this free-form value",
		"message":              "auth token key secret password credential in prose",
	}
	out := safeExportValue(value).(map[string]any)
	for _, key := range []string{"apiKey", "private_key", "credential", "accessKey", "APIKEY", "privateKey", "authorization_header", "auth_token_value", "secret_value", "password_hint", "publicKey", "public_key", "clientKey", "client_key", "serverKey", "server_key", "encryptionKey", "encryption_key", "signingKey", "signing_key", "authKey", "auth_key", "secretKey", "secret_key"} {
		if _, ok := out[key]; ok {
			t.Fatalf("sensitive key preserved: %q", key)
		}
	}
	for _, key := range []string{"author", "tokenizer", "directory_name", "message"} {
		if out[key] != value[key] {
			t.Fatalf("free-form data changed for %q: %#v", key, out[key])
		}
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

func TestSensitiveExportKeyMatchesNormalizedKeyVariants(t *testing.T) {
	for _, key := range []string{"publickey", "publicKEY", "clientkey", "serverkey", "encryptionkey", "signingkey", "authkey", "secretkey"} {
		if !sensitiveExportKey(key) {
			t.Fatalf("sensitive normalized key not redacted: %q", key)
		}
	}
}

func TestSensitiveExportKeyDoesNotRedactUnrelatedKeyNames(t *testing.T) {
	for _, key := range []string{"keyboard", "directory_name"} {
		if sensitiveExportKey(key) {
			t.Fatalf("unrelated key redacted: %q", key)
		}
	}
}
