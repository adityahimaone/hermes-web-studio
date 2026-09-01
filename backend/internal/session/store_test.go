package session

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestStoreCRUDPreservesUnknownFields(t *testing.T) {
	stateDir := t.TempDir()
	store := NewStore(stateDir)
	created, err := store.Create("session-1", "Initial", []json.RawMessage{raw(`{"role":"user","content":"hello"}`)})
	if err != nil || created.ID != "session-1" {
		t.Fatalf("create = %#v, err = %v", created, err)
	}
	if _, err := store.Update("session-1", map[string]json.RawMessage{
		"title": raw(`"Renamed"`), "pinned": raw(`true`), "future_field": raw(`{"keep":true}`),
	}); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load("session-1")
	if err != nil || len(loaded.Messages) != 1 {
		t.Fatalf("load = %#v, err = %v", loaded, err)
	}
	if _, ok := loaded.Metadata["future_field"]; !ok {
		t.Fatal("unknown field was not preserved")
	}
	if _, err := store.Rename("session-1", "Final"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetArchived("session-1", true); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete("session-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load("session-1"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("load after delete = %v", err)
	}
	if _, err := store.Create("../unsafe", "bad", nil); !errors.Is(err, ErrInvalidSessionID) {
		t.Fatalf("unsafe create = %v", err)
	}
}

func TestStoreConcurrentDifferentSessionWritesKeepIndexCurrent(t *testing.T) {
	store := NewStore(t.TempDir())
	for _, id := range []string{"session-a", "session-b"} {
		if _, err := store.Create(id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	for attempt := 0; attempt < 100; attempt++ {
		var wg sync.WaitGroup
		for _, id := range []string{"session-a", "session-b"} {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				if _, err := store.Update(id, map[string]json.RawMessage{"title": raw(`"updated"`)}); err != nil {
					t.Errorf("update %s: %v", id, err)
				}
			}(id)
		}
		wg.Wait()

		index, err := os.ReadFile(filepath.Join(store.stateDir, "sessions", "_index.json"))
		if err != nil {
			t.Fatal(err)
		}
		var rows []struct {
			ID string `json:"session_id"`
		}
		if err := json.Unmarshal(index, &rows); err != nil {
			t.Fatalf("index is invalid JSON after attempt %d: %v", attempt, err)
		}
		seen := map[string]bool{}
		for _, row := range rows {
			seen[row.ID] = true
		}
		if !seen["session-a"] || !seen["session-b"] {
			t.Fatalf("index after attempt %d omitted session: %s", attempt, index)
		}
	}
}

func TestStoreAppendMessagesSupportsResumeHistory(t *testing.T) {
	store := NewStore(t.TempDir())
	if _, err := store.Create("resume-1", "Resume", nil); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendMessages("resume-1", raw(`{"role":"user","content":"hello"}`), raw(`{"role":"assistant","content":"world"}`)); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load("resume-1")
	if err != nil {
		t.Fatal(err)
	}
	var first, second map[string]string
	if err := json.Unmarshal(loaded.Messages[0], &first); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(loaded.Messages[1], &second); err != nil {
		t.Fatal(err)
	}
	if first["role"] != "user" || first["content"] != "hello" {
		t.Fatalf("first message = %#v", first)
	}
	if second["role"] != "assistant" || second["content"] != "world" {
		t.Fatalf("second message = %#v", second)
	}
}

func TestStoreWritesLegacyIndexAndSafePermissions(t *testing.T) {
	stateDir := t.TempDir()
	store := NewStore(stateDir)
	if _, err := store.Create("session-1", "One", nil); err != nil {
		t.Fatal(err)
	}
	index, err := os.ReadFile(filepath.Join(stateDir, "sessions", "_index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(index, &rows); err != nil || len(rows) != 1 {
		t.Fatalf("index = %s, err = %v", index, err)
	}
	if _, ok := rows[0]["messages"]; ok {
		t.Fatal("index contains transcript")
	}
	info, err := os.Stat(filepath.Join(stateDir, "sessions", "session-1.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("permissions = %o", info.Mode().Perm())
	}
}

func TestStoreTruncatesMessagesAndPreservesMetadata(t *testing.T) {
	store := NewStore(t.TempDir())
	if _, err := store.Create("session-1", "Draft", []json.RawMessage{raw(`{"role":"user","content":"one"}`), raw(`{"role":"assistant","content":"two"}`)}); err != nil {
		t.Fatal(err)
	}
	if err := store.TruncateMessages("session-1", 1); err != nil {
		t.Fatal(err)
	}
	item, err := store.Load("session-1")
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != "Draft" || len(item.Messages) != 1 {
		t.Fatalf("item=%#v", item)
	}
}

func raw(value string) json.RawMessage { return json.RawMessage(value) }
