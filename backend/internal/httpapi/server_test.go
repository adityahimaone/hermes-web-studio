package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

func TestPublicGatewayURLRedactsPathQueryAndUserinfo(t *testing.T) {
	got := publicGatewayURL("https://user:secret@example.test/private/path?token=secret#fragment")
	if got != "https://example.test" {
		t.Fatalf("publicGatewayURL = %q", got)
	}
}

func TestHermesHealthResponseUsesSafeGatewayOrigin(t *testing.T) {
	api := newTestServer(t, "https://user:secret@example.test/private?token=secret", "")
	defer api.Close()
	response, err := api.Client().Get(api.URL + "/api/health/hermes")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["base_url"] != "https://example.test" || strings.Contains(string(body), "secret") || strings.Contains(string(body), "/private") {
		t.Fatalf("unsafe health payload: %s", body)
	}
}

func TestChatStreamsHermesGatewayResponse(t *testing.T) {
	requestSeen := make(chan struct{}, 1)
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("missing gateway auth")
		}
		if r.Header.Get("X-Hermes-Session-Id") != "session-1" {
			t.Errorf("missing session header")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["model"] != "selected-model" || body["provider"] != "selected-provider" || body["stream"] != true {
			t.Errorf("body = %#v", body)
		}
		requestSeen <- struct{}{}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello from Hermes\"}}]}\n\ndata: [DONE]\n\n")
	}))
	defer gw.Close()

	api := newTestServer(t, gw.URL, "secret")
	start := postJSON(t, api.URL+"/api/chat/start", map[string]any{"session_id": "session-1", "message": "hello", "model": "selected-model", "provider": "selected-provider"})
	var started map[string]string
	decode(t, start.Body, &started)
	_ = start.Body.Close()

	response, err := http.Get(api.URL + "/api/chat/stream?stream_id=" + started["stream_id"])
	if err != nil {
		t.Fatal(err)
	}
	stream, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	text := string(stream)
	if !strings.Contains(text, "event: token") || !strings.Contains(text, "Hello from Hermes") || !strings.Contains(text, "event: done") {
		t.Fatalf("stream = %s", text)
	}
	select {
	case <-requestSeen:
	default:
		t.Fatal("gateway request not observed")
	}
}

func TestGatewayAuthErrorIsRedacted(t *testing.T) {
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "upstream leaked secret detail", http.StatusUnauthorized)
	}))
	defer gw.Close()
	api := newTestServer(t, gw.URL, "wrong-key")
	start := postJSON(t, api.URL+"/api/chat/start", map[string]any{"message": "hello"})
	var started map[string]string
	decode(t, start.Body, &started)
	_ = start.Body.Close()
	response, _ := http.Get(api.URL + "/api/chat/stream?stream_id=" + started["stream_id"])
	stream, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	text := string(stream)
	if !strings.Contains(text, "gateway_auth_error") || strings.Contains(text, "leaked secret") || strings.Contains(text, "wrong-key") {
		t.Fatalf("unsafe stream = %s", text)
	}
}

func TestFailedGatewayTurnDoesNotPersistNewSessionOrEmitDone(t *testing.T) {
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "event: run.failed\ndata: {\"error\":\"provider secret detail\"}\n\n")
	}))
	defer gw.Close()
	api := newTestServer(t, gw.URL, "")
	defer api.Close()
	start := postJSON(t, api.URL+"/api/chat/start", map[string]any{"session_id": "failed-new", "message": "will fail"})
	var started map[string]string
	decode(t, start.Body, &started)
	_ = start.Body.Close()
	response, err := http.Get(api.URL + "/api/chat/stream?stream_id=" + started["stream_id"])
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	text := string(body)
	if strings.Contains(text, "event: done") || !strings.Contains(text, "event: apperror") || strings.Contains(text, "provider secret detail") {
		t.Fatalf("stream=%s", text)
	}
	response, err = api.Client().Get(api.URL + "/api/sessions")
	if err != nil {
		t.Fatal(err)
	}
	var listed map[string]any
	decode(t, response.Body, &listed)
	_ = response.Body.Close()
	listedBody, _ := json.Marshal(listed)
	if strings.Contains(string(listedBody), "failed-new") {
		t.Fatalf("failed turn persisted: %s", listedBody)
	}
}

func TestReadyReportsLocalDependenciesWithoutContactingGateway(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir()}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()
	response, err := api.Client().Get(api.URL + "/ready")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("ready status=%d", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"ready":true`) || !strings.Contains(string(body), `"workspace":true`) {
		t.Fatalf("unexpected readiness payload: %s", body)
	}
}

func TestSecurityHeadersCoverBrowserBoundary(t *testing.T) {
	api := newTestServer(t, "http://127.0.0.1:1", "")
	defer api.Close()
	response, err := api.Client().Get(api.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	for _, header := range []string{"X-Content-Type-Options", "Referrer-Policy", "X-Frame-Options", "Permissions-Policy", "Content-Security-Policy"} {
		if response.Header.Get(header) == "" {
			t.Errorf("missing security header %s", header)
		}
	}
	if strings.Contains(response.Header.Get("Content-Security-Policy"), "unsafe-eval") {
		t.Error("CSP permits unsafe eval")
	}
}

func TestCancelStopsUpstreamTurn(t *testing.T) {
	requestStarted := make(chan struct{})
	cancelled := make(chan struct{})
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		close(requestStarted)
		<-r.Context().Done()
		close(cancelled)
	}))
	defer gw.Close()
	api := newTestServer(t, gw.URL, "")
	start := postJSON(t, api.URL+"/api/chat/start", map[string]any{"message": "wait"})
	var started map[string]string
	decode(t, start.Body, &started)
	_ = start.Body.Close()
	select {
	case <-requestStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("upstream request did not start")
	}
	cancelResponse, err := http.Post(api.URL+"/api/chat/cancel?stream_id="+started["stream_id"], "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = cancelResponse.Body.Close()
	select {
	case <-cancelled:
	case <-time.After(2 * time.Second):
		t.Fatal("upstream request was not cancelled")
	}
}

func TestAttachmentUploadAndMultimodalGatewayPayload(t *testing.T) {
	seenPayload := make(chan string, 1)
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		seenPayload <- string(body)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\ndata: [DONE]\n\n")
	}))
	defer gw.Close()
	api := newTestServer(t, gw.URL, "")
	defer api.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "note.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("hello attachment"))
	_ = writer.Close()
	request, _ := http.NewRequest(http.MethodPost, api.URL+"/api/attachments", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	upload, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	var uploaded map[string]any
	decode(t, upload.Body, &uploaded)
	_ = upload.Body.Close()
	if upload.StatusCode != http.StatusCreated || uploaded["id"] == "" {
		t.Fatalf("upload status=%d body=%v", upload.StatusCode, uploaded)
	}

	start := postJSON(t, api.URL+"/api/chat/start", map[string]any{"message": "describe", "attachment_ids": []string{uploaded["id"].(string)}})
	var started map[string]string
	decode(t, start.Body, &started)
	_ = start.Body.Close()
	stream, err := http.Get(api.URL + "/api/chat/stream?stream_id=" + started["stream_id"])
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.ReadAll(stream.Body)
	_ = stream.Body.Close()
	select {
	case payload := <-seenPayload:
		if !strings.Contains(payload, `"text":"hello attachment"`) || strings.Contains(payload, "Authorization") {
			t.Fatalf("payload=%s", payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("gateway payload not observed")
	}
}

func TestRunsModeKeepsAttachmentTurnsOnChatCompletions(t *testing.T) {
	paths := make(chan string, 1)
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths <- r.URL.Path
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\ndata: [DONE]\n\n")
	}))
	defer gw.Close()
	cfg := config.Config{GatewayBaseURL: gw.URL, DefaultModel: "default", ReadTimeout: 5 * time.Second, StateDir: t.TempDir(), UseRunsAPI: true}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: gw.URL, ReadTimeout: cfg.ReadTimeout})).Handler())
	defer api.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "note.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("attachment"))
	_ = writer.Close()
	uploadRequest, _ := http.NewRequest(http.MethodPost, api.URL+"/api/attachments", &body)
	uploadRequest.Header.Set("Content-Type", writer.FormDataContentType())
	upload, err := http.DefaultClient.Do(uploadRequest)
	if err != nil {
		t.Fatal(err)
	}
	var uploaded map[string]any
	decode(t, upload.Body, &uploaded)
	_ = upload.Body.Close()

	start := postJSON(t, api.URL+"/api/chat/start", map[string]any{"message": "describe", "attachment_ids": []string{uploaded["id"].(string)}})
	var started map[string]string
	decode(t, start.Body, &started)
	_ = start.Body.Close()
	stream, err := http.Get(api.URL + "/api/chat/stream?stream_id=" + started["stream_id"])
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.ReadAll(stream.Body)
	_ = stream.Body.Close()
	select {
	case path := <-paths:
		if path != "/v1/chat/completions" {
			t.Fatalf("attachment turn used %s", path)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("gateway request not observed")
	}
}

func TestCompletedStreamReplaysAfterLastEventID(t *testing.T) {
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"replay\"}}]}\n\ndata: [DONE]\n\n")
	}))
	defer gw.Close()
	api := newTestServer(t, gw.URL, "")
	defer api.Close()
	start := postJSON(t, api.URL+"/api/chat/start", map[string]any{"message": "hello"})
	var started map[string]string
	decode(t, start.Body, &started)
	_ = start.Body.Close()
	first, _ := http.Get(api.URL + "/api/chat/stream?stream_id=" + started["stream_id"])
	firstBody, _ := io.ReadAll(first.Body)
	_ = first.Body.Close()
	if !strings.Contains(string(firstBody), "id: 1") || !strings.Contains(string(firstBody), "id: 2") {
		t.Fatalf("first stream=%s", firstBody)
	}
	reconnectRequest, _ := http.NewRequest(http.MethodGet, api.URL+"/api/chat/stream?stream_id="+started["stream_id"], nil)
	reconnectRequest.Header.Set("Last-Event-ID", "1")
	reconnected, err := http.DefaultClient.Do(reconnectRequest)
	if err != nil {
		t.Fatal(err)
	}
	reconnectedBody, _ := io.ReadAll(reconnected.Body)
	_ = reconnected.Body.Close()
	if strings.Contains(string(reconnectedBody), "id: 1") || !strings.Contains(string(reconnectedBody), "id: 2") {
		t.Fatalf("replayed stream=%s", reconnectedBody)
	}
}

func TestApprovalRouteForwardsDecision(t *testing.T) {
	seen := make(chan string, 1)
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/runs/run-1/approval" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		seen <- string(body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer gw.Close()
	api := newTestServer(t, gw.URL, "approval-secret")
	defer api.Close()
	response := postJSON(t, api.URL+"/api/runs/run-1/approval", map[string]string{"decision": "approved"})
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", response.StatusCode)
	}
	_ = response.Body.Close()
	select {
	case payload := <-seen:
		if payload != `{"choice":"once"}` {
			t.Fatalf("payload=%s", payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("approval was not forwarded")
	}
}

func TestSessionsAPIListsAndLoadsLegacySession(t *testing.T) {
	stateDir := t.TempDir()
	sessionsDir := filepath.Join(stateDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessionsDir, "session-1.json"), []byte(`{"session_id":"session-1","title":"Hello","messages":[{"role":"user","content":"hello"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessionsDir, "_index.json"), []byte(`[{"session_id":"session-1","title":"Hello"}]`), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: stateDir, DefaultModel: "default"}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()
	response, err := http.Get(api.URL + "/api/sessions")
	if err != nil {
		t.Fatal(err)
	}
	var listing map[string]any
	decode(t, response.Body, &listing)
	_ = response.Body.Close()
	if len(listing["sessions"].([]any)) != 1 {
		t.Fatalf("listing = %#v", listing)
	}

	response, err = http.Get(api.URL + "/api/sessions/session-1")
	if err != nil {
		t.Fatal(err)
	}
	var loaded map[string]any
	decode(t, response.Body, &loaded)
	_ = response.Body.Close()
	if loaded["session_id"] != "session-1" || len(loaded["messages"].([]any)) != 1 {
		t.Fatalf("loaded = %#v", loaded)
	}
}

func TestSessionsAPIFlattensLegacyMetadataForBrowserContract(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stateDir, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	data := `{"session_id":"session-1","title":"Hello","updated_at":"2026-08-30T12:00:00Z","pinned":true,"tags":["work"],"messages":[]}`
	if err := os.WriteFile(filepath.Join(stateDir, "sessions", "session-1.json"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: stateDir}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()
	response, err := http.Get(api.URL + "/api/sessions")
	if err != nil {
		t.Fatal(err)
	}
	var listing struct {
		Sessions []map[string]any `json:"sessions"`
	}
	decode(t, response.Body, &listing)
	_ = response.Body.Close()
	if len(listing.Sessions) != 1 || listing.Sessions[0]["pinned"] != true || listing.Sessions[0]["tags"].([]any)[0] != "work" {
		t.Fatalf("listing=%#v", listing)
	}
}

func TestSessionsAPISearchesTranscriptServerSide(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stateDir, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	data := `{"session_id":"session-search","title":"Untitled","messages":[{"role":"user","content":"needle in transcript"}]}`
	if err := os.WriteFile(filepath.Join(stateDir, "sessions", "session-search.json"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: stateDir}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()
	response, err := http.Get(api.URL + "/api/sessions?q=NEEDLE")
	if err != nil {
		t.Fatal(err)
	}
	var listing struct {
		Sessions []map[string]any `json:"sessions"`
	}
	decode(t, response.Body, &listing)
	_ = response.Body.Close()
	if len(listing.Sessions) != 1 || listing.Sessions[0]["session_id"] != "session-search" {
		t.Fatalf("search listing = %#v", listing)
	}
	if _, ok := listing.Sessions[0]["messages"]; ok {
		t.Fatal("search listing returned transcript messages")
	}
}

func TestSessionDuplicateAndMarkdownExport(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stateDir, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	data := `{"session_id":"session-source","title":"Source","messages":[{"role":"user","content":"hello"},{"role":"assistant","content":"world"}]}`
	if err := os.WriteFile(filepath.Join(stateDir, "sessions", "session-source.json"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: stateDir}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()
	response, err := http.Post(api.URL+"/api/sessions/session-source/duplicate", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("duplicate status=%d", response.StatusCode)
	}
	var duplicate map[string]any
	decode(t, response.Body, &duplicate)
	_ = response.Body.Close()
	if duplicate["title"] != "Copy of Source" || duplicate["session_id"] == "session-source" {
		t.Fatalf("duplicate=%#v", duplicate)
	}
	response, err = http.Get(api.URL + "/api/sessions/session-source/export?format=markdown")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || !strings.Contains(string(body), "# Source") || !strings.Contains(string(body), "## Hermes") {
		t.Fatalf("export status=%d body=%s", response.StatusCode, body)
	}
}

func TestSessionJSONExportRedactsTitleAndSecretKeyVariants(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stateDir, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	data := `{"session_id":"session-json-secrets","title":"/home/private/workspace/session","messages":[{"role":"user","content":"hello","metadata":{"apiKey":"camel-secret","private_key":"underscore-secret","credential":"credential-secret","accessKey":"access-secret","nested":{"APIKEY":"upper-secret","privateKey":"camel-private-secret"},"safe":"kept"}}]}`
	if err := os.WriteFile(filepath.Join(stateDir, "sessions", "session-json-secrets.json"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: stateDir}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()

	response, err := http.Get(api.URL + "/api/sessions/session-json-secrets/export?format=json")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("export status=%d body=%s", response.StatusCode, body)
	}
	var exported map[string]any
	if err := json.Unmarshal(body, &exported); err != nil {
		t.Fatalf("invalid JSON export: %v", err)
	}
	if exported["title"] != "[redacted]" {
		t.Fatalf("title leaked: %#v", exported["title"])
	}
	text := string(body)
	for _, secret := range []string{"camel-secret", "underscore-secret", "credential-secret", "access-secret", "upper-secret", "camel-private-secret", "/home/private/workspace/session"} {
		if strings.Contains(text, secret) {
			t.Fatalf("unsafe JSON export contains %q: %s", secret, body)
		}
	}
	if !strings.Contains(text, `"safe": "kept"`) {
		t.Fatalf("safe metadata missing: %s", body)
	}
}

func TestSessionJSONExportIsSafeAttachment(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stateDir, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	data := `{"session_id":"session-json","title":"JSON export","created_at":"2026-01-02T03:04:05Z","api_key":"do-not-export","workspace_path":"/home/private/project","messages":[{"role":"user","content":"hello","metadata":{"authorization":"Bearer secret","path":"/srv/private/file"}},{"role":"assistant","content":"world"}]}`
	if err := os.WriteFile(filepath.Join(stateDir, "sessions", "session-json.json"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: stateDir}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()

	response, err := http.Get(api.URL + "/api/sessions/session-json/export?format=json")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || response.Header.Get("Content-Type") != "application/json; charset=utf-8" || response.Header.Get("Content-Disposition") != `attachment; filename="hermes-session.json"` {
		t.Fatalf("export status=%d content-type=%q disposition=%q", response.StatusCode, response.Header.Get("Content-Type"), response.Header.Get("Content-Disposition"))
	}
	var exported map[string]any
	if err := json.Unmarshal(body, &exported); err != nil {
		t.Fatalf("invalid JSON export: %v", err)
	}
	if exported["session_id"] != "session-json" || exported["title"] != "JSON export" {
		t.Fatalf("metadata missing: %#v", exported)
	}
	messages, ok := exported["messages"].([]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("messages missing: %#v", exported["messages"])
	}
	text := string(body)
	if strings.Contains(text, "do-not-export") || strings.Contains(text, "/home/private/project") || strings.Contains(text, "/srv/private/file") || strings.Contains(text, "Bearer secret") {
		t.Fatalf("unsafe JSON export: %s", body)
	}

	response, err = http.Get(api.URL + "/api/sessions/session-json/export?format=xml")
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid format status=%d", response.StatusCode)
	}
}

func TestSessionsAPIRejectsUnsafeAndMissingIDs(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir()}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()
	response, err := http.Get(api.URL + "/api/sessions/../x")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("traversal status = %d", response.StatusCode)
	}
	_ = response.Body.Close()
	response, err = http.Get(api.URL + "/api/sessions/missing")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("missing status = %d", response.StatusCode)
	}
	_ = response.Body.Close()
}

func newTestServer(t *testing.T, gatewayURL, key string) *httptest.Server {
	t.Helper()
	cfg := config.Config{GatewayBaseURL: gatewayURL, GatewayAPIKey: key, DefaultModel: "default", ReadTimeout: 5 * time.Second, StateDir: t.TempDir()}
	client := gateway.New(gateway.Config{BaseURL: gatewayURL, APIKey: key, ReadTimeout: 5 * time.Second})
	return httptest.NewServer(NewWithGateway(cfg, client).Handler())
}

func postJSON(t *testing.T, url string, body any) *http.Response {
	t.Helper()
	payload, _ := json.Marshal(body)
	request, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func decode(t *testing.T, reader io.Reader, target any) {
	t.Helper()
	if err := json.NewDecoder(reader).Decode(target); err != nil {
		t.Fatal(err)
	}
}
