package control

import (
	"path/filepath"
	"testing"
)

func TestSpaceMutationRestoresRegistryAfterPersistFailure(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	item, err := s.Create("spaces", Item{Title: "kept"})
	if err != nil {
		t.Fatal(err)
	}
	before := s.Preferences()
	s.path = filepath.Join(t.TempDir(), "missing", "control.json")
	if err := s.SpaceMutation("spaces", item.ID, "active_space", false); err == nil {
		t.Fatal("expected persist failure")
	}
	items, _ := s.List("spaces")
	if len(items) != 1 || items[0].ID != item.ID || len(s.Preferences()) != len(before) {
		t.Fatalf("registry mutated: items=%v preferences=%v", items, s.Preferences())
	}
}

func TestSetPreferencesRestoresStateAfterPersistFailure(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s.path = filepath.Join(t.TempDir(), "missing", "control.json")
	if err := s.SetPreferences(map[string]string{"theme": "dark"}); err == nil {
		t.Fatal("expected persist failure")
	}
	if got := s.Preferences()["theme"]; got != "" {
		t.Fatalf("preference persisted in memory: %q", got)
	}
}
