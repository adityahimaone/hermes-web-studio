package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

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
		if body["model"] != "default" || body["stream"] != true {
			t.Errorf("body = %#v", body)
		}
		requestSeen <- struct{}{}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello from Hermes\"}}]}\n\ndata: [DONE]\n\n")
	}))
	defer gw.Close()

	api := newTestServer(gw.URL, "secret")
	start := postJSON(t, api.URL+"/api/chat/start", map[string]any{"session_id": "session-1", "message": "hello"})
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
	api := newTestServer(gw.URL, "wrong-key")
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

func TestCancelStopsUpstreamTurn(t *testing.T) {
	cancelled := make(chan struct{})
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		<-r.Context().Done()
		close(cancelled)
	}))
	defer gw.Close()
	api := newTestServer(gw.URL, "")
	start := postJSON(t, api.URL+"/api/chat/start", map[string]any{"message": "wait"})
	var started map[string]string
	decode(t, start.Body, &started)
	_ = start.Body.Close()
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

func newTestServer(gatewayURL, key string) *httptest.Server {
	cfg := config.Config{GatewayBaseURL: gatewayURL, GatewayAPIKey: key, DefaultModel: "default", ReadTimeout: 5 * time.Second}
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
