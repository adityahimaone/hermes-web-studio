package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

func TestSessionReadRoutesWaitForSessionLock(t *testing.T) {
	s := NewWithGateway(config.Config{StateDir: t.TempDir()}, gateway.New(gateway.Config{}))
	if _, err := s.sessions.Create("locked", "Locked", nil); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		call func(*httptest.ResponseRecorder)
	}{
		{"get", func(w *httptest.ResponseRecorder) {
			r := httptest.NewRequest(http.MethodGet, "/api/sessions/locked", nil)
			r.SetPathValue("session_id", "locked")
			s.handleSession(w, r)
		}},
		{"duplicate", func(w *httptest.ResponseRecorder) {
			r := httptest.NewRequest(http.MethodPost, "/api/sessions/locked/duplicate", nil)
			r.SetPathValue("session_id", "locked")
			s.handleSessionDuplicate(w, r)
		}},
		{"export", func(w *httptest.ResponseRecorder) {
			r := httptest.NewRequest(http.MethodGet, "/api/sessions/locked/export?format=json", nil)
			r.SetPathValue("session_id", "locked")
			s.handleSessionExport(w, r)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lock := s.sessionLock("locked")
			lock.Lock()
			done := make(chan struct{})
			response := httptest.NewRecorder()
			go func() { tc.call(response); close(done) }()
			select {
			case <-done:
				t.Fatal("session route bypassed lock")
			case <-time.After(30 * time.Millisecond):
			}
			lock.Unlock()
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("session route stayed blocked after unlock")
			}
		})
	}
}

func TestRunTurnStopsOnNonNotFoundInitialLoadError(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(stateDir, "sessions"), []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	s := NewWithGateway(config.Config{StateDir: stateDir}, gateway.New(gateway.Config{}))
	item := &turn{streamID: "stream", sessionID: "broken", cancel: func() {}, subs: make(map[chan replayEvent]struct{})}
	s.runTurn(context.Background(), item, gateway.ChatRequest{SessionID: "broken", Message: "hello"})
	if len(item.events) != 1 || item.events[0].Event.Name != "apperror" {
		t.Fatalf("events=%#v", item.events)
	}
	if item.events[0].Event.Data["code"] != "session_unavailable" {
		t.Fatalf("event=%#v", item.events[0].Event)
	}
}

func TestRunTurnNewSessionPersistsOneUserMessage(t *testing.T) {
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"answer\"}}]}\n\ndata: [DONE]\n\n")
	}))
	defer gw.Close()
	s := NewWithGateway(config.Config{StateDir: t.TempDir()}, gateway.New(gateway.Config{BaseURL: gw.URL, ReadTimeout: time.Second}))
	item := &turn{streamID: "stream", sessionID: "new", cancel: func() {}, subs: make(map[chan replayEvent]struct{})}
	s.runTurn(context.Background(), item, gateway.ChatRequest{SessionID: "new", Message: "hello"})
	if len(item.events) == 0 {
		t.Fatal("no events")
	}
	loaded, err := s.sessions.Load("new")
	if err != nil {
		t.Fatalf("events=%#v load=%v", item.events, err)
	}
	if len(loaded.Messages) != 2 {
		t.Fatalf("messages=%s", mustJSONMessages(loaded.Messages))
	}
	var user, assistant map[string]string
	if err := json.Unmarshal(loaded.Messages[0], &user); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(loaded.Messages[1], &assistant); err != nil {
		t.Fatal(err)
	}
	if user["role"] != "user" || user["content"] != "hello" || assistant["role"] != "assistant" || assistant["content"] != "answer" {
		t.Fatalf("messages=%s", mustJSONMessages(loaded.Messages))
	}
}

func TestRunTurnCancellationRollsBackOwnedUserMessage(t *testing.T) {
	gatewayStarted := make(chan struct{})
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		close(gatewayStarted)
		<-r.Context().Done()
	}))
	defer gw.Close()

	s := NewWithGateway(config.Config{StateDir: t.TempDir()}, gateway.New(gateway.Config{BaseURL: gw.URL, ReadTimeout: time.Second}))
	ctx, cancel := context.WithCancel(context.Background())
	item := &turn{streamID: "stream", sessionID: "cancelled", cancel: cancel, subs: make(map[chan replayEvent]struct{})}
	done := make(chan struct{})
	go func() {
		s.runTurn(ctx, item, gateway.ChatRequest{SessionID: "cancelled", Message: "cancel me"})
		close(done)
	}()
	select {
	case <-gatewayStarted:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("gateway request not observed")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cancelled turn did not finish")
	}
	if _, err := s.sessions.Load("cancelled"); !os.IsNotExist(err) {
		t.Fatalf("cancelled session err=%v, want not exist", err)
	}
	if len(item.events) == 0 || item.events[len(item.events)-1].Event.Name != "cancel" {
		t.Fatalf("events=%#v", item.events)
	}
}

func mustJSONMessages(messages []json.RawMessage) string {
	data, _ := json.Marshal(messages)
	return string(data)
}

func TestRunTurnCreateCollisionDoesNotDeleteSession(t *testing.T) {
	stateDir := t.TempDir()
	first := NewWithGateway(config.Config{StateDir: stateDir}, gateway.New(gateway.Config{}))
	second := NewWithGateway(config.Config{StateDir: stateDir}, gateway.New(gateway.Config{}))
	if _, err := first.sessions.Create("collision", "Existing", []json.RawMessage{mustMessage("user", "keep")}); err != nil {
		t.Fatal(err)
	}
	item := &turn{streamID: "stream", sessionID: "collision", cancel: func() {}, subs: make(map[chan replayEvent]struct{})}
	second.runTurn(context.Background(), item, gateway.ChatRequest{SessionID: "collision", Message: "new"})
	loaded, err := second.sessions.Load("collision")
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Messages) != 1 {
		t.Fatalf("messages=%s", mustJSONMessages(loaded.Messages))
	}
	var message map[string]string
	if err := json.Unmarshal(loaded.Messages[0], &message); err != nil || message["role"] != "user" || message["content"] != "keep" {
		t.Fatalf("messages=%s", mustJSONMessages(loaded.Messages))
	}
}
