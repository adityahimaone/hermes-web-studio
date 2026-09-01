package httpapi

import (
	"context"
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
