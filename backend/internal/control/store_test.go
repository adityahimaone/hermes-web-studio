package control

import "testing"

func TestStorePersistsCollectionsAndFiltersSecrets(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	item, err := s.Create("todos", Item{Title: "Ship M4"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update("todos", item.ID, Item{Status: "done"}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetPreferences(map[string]string{"theme": "dark", "api_token": "hidden"}); err != nil {
		t.Fatal(err)
	}
	if s.Preferences()["api_token"] != "" || s.Preferences()["theme"] != "dark" {
		t.Fatalf("preferences=%v", s.Preferences())
	}
	s2, err := New(s.path[:len(s.path)-len("control.json")])
	if err != nil {
		t.Fatal(err)
	}
	items, err := s2.List("todos")
	if err != nil || len(items) != 1 || items[0].Status != "done" {
		t.Fatalf("items=%v err=%v", items, err)
	}
}
