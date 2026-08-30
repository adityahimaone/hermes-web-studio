package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReaderListsIndexMetadataAndLoadsMessages(t *testing.T) {
	stateDir := fixtureState(t)
	reader := NewLegacySessionReader(stateDir)

	summaries, err := reader.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 || summaries[0].ID != "session-1" || summaries[0].Title != "First chat" {
		t.Fatalf("summaries = %#v", summaries)
	}
	if _, ok := summaries[0].Metadata["future_field"]; !ok {
		t.Fatal("unknown metadata field was discarded")
	}

	loaded, err := reader.Load("session-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Messages) != 2 {
		t.Fatalf("messages = %#v", loaded.Messages)
	}
	var firstMessage map[string]string
	if err := json.Unmarshal(loaded.Messages[0], &firstMessage); err != nil {
		t.Fatal(err)
	}
	if firstMessage["role"] != "user" || firstMessage["content"] != "hello" {
		t.Fatalf("first message changed: %s", loaded.Messages[0])
	}
}

func TestReaderFallsBackWhenIndexIsCorrupt(t *testing.T) {
	stateDir := fixtureState(t)
	index := filepath.Join(stateDir, "sessions", "_index.json")
	if err := os.WriteFile(index, []byte("not-json"), 0600); err != nil {
		t.Fatal(err)
	}

	summaries, err := NewLegacySessionReader(stateDir).List()
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 || summaries[0].ID != "session-1" {
		t.Fatalf("fallback summaries = %#v", summaries)
	}
}

func TestReaderSearchesTranscriptWithoutReturningMessages(t *testing.T) {
	summaries, err := NewLegacySessionReader(fixtureState(t)).Search("WORLD")
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 || summaries[0].ID != "session-1" {
		t.Fatalf("search summaries = %#v", summaries)
	}
	if _, ok := summaries[0].Metadata["messages"]; ok {
		t.Fatal("search summary returned transcript messages")
	}
}

func TestReaderRejectsTraversalAndDoesNotWrite(t *testing.T) {
	stateDir := fixtureState(t)
	before, err := os.ReadFile(filepath.Join(stateDir, "sessions", "session-1.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"../x", `a/b`, `..\\x`, ".."} {
		if _, err := NewLegacySessionReader(stateDir).Load(id); err != ErrInvalidSessionID {
			t.Fatalf("id %q error = %v", id, err)
		}
	}
	after, err := os.ReadFile(filepath.Join(stateDir, "sessions", "session-1.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("read-only reader changed the legacy session")
	}
}

func TestResolveStateDirHonorsOverrides(t *testing.T) {
	environ := func(key string) string {
		if key == "HERMES_WEBUI_STATE_DIR" {
			return "/tmp/custom-state"
		}
		return ""
	}
	got, err := ResolveStateDir(environ, func() (string, error) { return "/home/tester", nil })
	if err != nil || got != "/tmp/custom-state" {
		t.Fatalf("state dir = %q, err = %v", got, err)
	}
}

func fixtureState(t *testing.T) string {
	t.Helper()
	stateDir := t.TempDir()
	sessionsDir := filepath.Join(stateDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0700); err != nil {
		t.Fatal(err)
	}
	session := map[string]any{
		"session_id":   "session-1",
		"title":        "First chat",
		"workspace":    "/workspace",
		"future_field": map[string]any{"keep": true},
		"messages": []map[string]string{
			{"role": "user", "content": "hello"},
			{"role": "assistant", "content": "world"},
		},
	}
	contents, err := json.Marshal(session)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessionsDir, "session-1.json"), contents, 0600); err != nil {
		t.Fatal(err)
	}
	index, err := json.Marshal([]map[string]any{{"session_id": "session-1", "title": "First chat", "future_field": true}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessionsDir, "_index.json"), index, 0600); err != nil {
		t.Fatal(err)
	}
	return stateDir
}
