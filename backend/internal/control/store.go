package control

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("control item not found")
var validCollections = map[string]bool{"tasks": true, "todos": true, "goals": true, "spaces": true}

type Item struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Status      string            `json:"status"`
	DueAt       string            `json:"due_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}
type State struct {
	Tasks       []Item            `json:"tasks"`
	Todos       []Item            `json:"todos"`
	Goals       []Item            `json:"goals"`
	Spaces      []Item            `json:"spaces"`
	Preferences map[string]string `json:"preferences"`
}
type Store struct {
	path  string
	mu    sync.RWMutex
	state State
}

func New(stateDir string) (*Store, error) {
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(stateDir, "control.json"), state: State{Preferences: map[string]string{}}}
	data, err := os.ReadFile(s.path)
	if err == nil {
		if json.Unmarshal(data, &s.state) != nil {
			return nil, errors.New("control state is invalid")
		}
	}
	if s.state.Preferences == nil {
		s.state.Preferences = map[string]string{}
	}
	return s, nil
}
func (s *Store) List(collection string) ([]Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items, err := s.items(collection)
	if err != nil {
		return nil, err
	}
	result := append([]Item(nil), items...)
	return result, nil
}
func (s *Store) Create(collection string, item Item) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !validCollections[collection] {
		return Item{}, errors.New("invalid collection")
	}
	now := time.Now().UTC()
	item.ID = newID()
	item.CreatedAt, item.UpdatedAt = now, now
	if item.Status == "" {
		item.Status = "open"
	}
	items, _ := s.items(collection)
	items = append(items, item)
	s.setItems(collection, items)
	return item, s.persist()
}
func (s *Store) Update(collection, id string, patch Item) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, err := s.items(collection)
	if err != nil {
		return Item{}, err
	}
	for i := range items {
		if items[i].ID == id {
			if patch.Title != "" {
				items[i].Title = patch.Title
			}
			if patch.Description != "" {
				items[i].Description = patch.Description
			}
			if patch.Status != "" {
				items[i].Status = patch.Status
			}
			if patch.DueAt != "" {
				items[i].DueAt = patch.DueAt
			}
			items[i].UpdatedAt = time.Now().UTC()
			s.setItems(collection, items)
			return items[i], s.persist()
		}
	}
	return Item{}, ErrNotFound
}
func (s *Store) Delete(collection, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, err := s.items(collection)
	if err != nil {
		return err
	}
	for i := range items {
		if items[i].ID == id {
			s.setItems(collection, append(items[:i], items[i+1:]...))
			return s.persist()
		}
	}
	return ErrNotFound
}
func (s *Store) Preferences() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := map[string]string{}
	for k, v := range s.state.Preferences {
		result[k] = v
	}
	return result
}
func (s *Store) SetPreferences(values map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, value := range values {
		if strings.HasPrefix(key, "secret") || strings.Contains(strings.ToLower(key), "token") {
			continue
		}
		s.state.Preferences[key] = value
	}
	return s.persist()
}
func (s *Store) items(collection string) ([]Item, error) {
	if !validCollections[collection] {
		return nil, errors.New("invalid collection")
	}
	switch collection {
	case "tasks":
		return s.state.Tasks, nil
	case "todos":
		return s.state.Todos, nil
	case "goals":
		return s.state.Goals, nil
	default:
		return s.state.Spaces, nil
	}
}
func (s *Store) setItems(collection string, items []Item) {
	switch collection {
	case "tasks":
		s.state.Tasks = items
	case "todos":
		s.state.Todos = items
	case "goals":
		s.state.Goals = items
	case "spaces":
		s.state.Spaces = items
	}
}
func (s *Store) persist() error {
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
func newID() string { return time.Now().UTC().Format("20060102T150405.000000000Z") }
