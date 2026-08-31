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
	if err := s.SetPreferences(map[string]string{
		"theme": "dark", "api_token": "hidden", "API_KEY": "hidden", "password": "hidden",
		"Authorization": "Bearer hidden", "safe_key": "kept",
	}); err != nil {
		t.Fatal(err)
	}
	preferences := s.Preferences()
	if preferences["api_token"] != "" || preferences["API_KEY"] != "" || preferences["password"] != "" || preferences["Authorization"] != "" || preferences["theme"] != "dark" || preferences["safe_key"] != "kept" {
		t.Fatalf("preferences=%v", preferences)
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

func TestTaskRunPauseAndHistory(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	task, err := s.Create("tasks", Item{Title: "Daily check"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.PauseTask(task.ID, true); err != nil {
		t.Fatal(err)
	}
	record, err := s.RunTask(task.ID)
	if err != nil || record.Status != "completed" {
		t.Fatalf("record=%v err=%v", record, err)
	}
	if len(s.History(task.ID)) != 1 {
		t.Fatal("run history missing")
	}
}
