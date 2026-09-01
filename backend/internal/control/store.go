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
var validCollections = map[string]bool{"tasks": true, "todos": true, "goals": true, "spaces": true, "artifacts": true, "boards": true}

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
	TaskHistory []RunRecord       `json:"task_history"`
	Artifacts   []Item            `json:"artifacts"`
	Boards      []Item            `json:"boards"`
	MCPServers  []MCPServer       `json:"mcp_servers"`
	Plugins     map[string]Plugin `json:"plugins"`
}

type MCPTool struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type MCPServer struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Transport  string            `json:"transport"`
	Endpoint   string            `json:"endpoint,omitempty"`
	Command    string            `json:"command,omitempty"`
	Args       []string          `json:"args,omitempty"`
	Enabled    bool              `json:"enabled"`
	Status     string            `json:"status"`
	Tools      []MCPTool         `json:"tools,omitempty"`
	Settings   map[string]string `json:"settings,omitempty"`
	SecretKeys []string          `json:"secret_keys,omitempty"`
}

type Plugin struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Path     string            `json:"path"`
	Enabled  bool              `json:"enabled"`
	Status   string            `json:"status"`
	Settings map[string]string `json:"settings,omitempty"`
}
type RunRecord struct {
	ID         string     `json:"id"`
	TaskID     string     `json:"task_id"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Error      string     `json:"error,omitempty"`
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
	s := &Store{path: filepath.Join(stateDir, "control.json"), state: State{Tasks: []Item{}, Todos: []Item{}, Goals: []Item{}, Spaces: []Item{}, Artifacts: []Item{}, Boards: []Item{}, Preferences: map[string]string{}, TaskHistory: []RunRecord{}, MCPServers: []MCPServer{}, Plugins: map[string]Plugin{}}}
	data, err := os.ReadFile(s.path)
	if err == nil {
		if json.Unmarshal(data, &s.state) != nil {
			return nil, errors.New("control state is invalid")
		}
	}
	if s.state.Preferences == nil {
		s.state.Preferences = map[string]string{}
	}
	if s.state.Tasks == nil {
		s.state.Tasks = []Item{}
	}
	if s.state.Todos == nil {
		s.state.Todos = []Item{}
	}
	if s.state.Goals == nil {
		s.state.Goals = []Item{}
	}
	if s.state.Spaces == nil {
		s.state.Spaces = []Item{}
	}
	if s.state.TaskHistory == nil {
		s.state.TaskHistory = []RunRecord{}
	}
	if s.state.Artifacts == nil {
		s.state.Artifacts = []Item{}
	}
	if s.state.Boards == nil {
		s.state.Boards = []Item{}
	}
	if s.state.MCPServers == nil {
		s.state.MCPServers = []MCPServer{}
	}
	if s.state.Plugins == nil {
		s.state.Plugins = map[string]Plugin{}
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
	result := append([]Item{}, items...)
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
			if patch.Metadata != nil {
				if items[i].Metadata == nil {
					items[i].Metadata = map[string]string{}
				}
				for key, value := range patch.Metadata {
					items[i].Metadata[key] = value
				}
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
func (s *Store) RunTask(id string) (RunRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.state.Tasks {
		if s.state.Tasks[i].ID == id {
			now := time.Now().UTC()
			s.state.Tasks[i].Status = "running"
			s.state.Tasks[i].UpdatedAt = now
			finished := now
			record := RunRecord{ID: newID(), TaskID: id, Status: "completed", StartedAt: now, FinishedAt: &finished}
			s.state.Tasks[i].Status = "completed"
			s.state.Tasks[i].UpdatedAt = finished
			s.state.TaskHistory = append([]RunRecord{record}, s.state.TaskHistory...)
			if len(s.state.TaskHistory) > 100 {
				s.state.TaskHistory = s.state.TaskHistory[:100]
			}
			return record, s.persist()
		}
	}
	return RunRecord{}, ErrNotFound
}
func (s *Store) PauseTask(id string, paused bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.state.Tasks {
		if s.state.Tasks[i].ID == id {
			if paused {
				s.state.Tasks[i].Status = "paused"
			} else {
				s.state.Tasks[i].Status = "open"
			}
			s.state.Tasks[i].UpdatedAt = time.Now().UTC()
			return s.persist()
		}
	}
	return ErrNotFound
}
func (s *Store) History(taskID string) []RunRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]RunRecord, 0)
	for _, item := range s.state.TaskHistory {
		if taskID == "" || item.TaskID == taskID {
			result = append(result, item)
		}
	}
	return result
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
func (s *Store) SpaceMutation(collection, id, activeKey string, deleteItem bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, err := s.items(collection)
	if err != nil {
		return err
	}
	if deleteItem && s.state.Preferences[activeKey] == id {
		return errors.New("active space protected")
	}
	found := false
	for i := range items {
		if items[i].ID == id {
			found = true
			if deleteItem {
				items = append(items[:i], items[i+1:]...)
			}
			break
		}
	}
	if !found {
		return ErrNotFound
	}
	if !deleteItem {
		s.state.Preferences[activeKey] = id
	}
	s.setItems(collection, items)
	return s.persist()
}

func (s *Store) SetPreferences(values map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, value := range values {
		lowerKey := strings.ToLower(strings.TrimSpace(key))
		if strings.HasPrefix(lowerKey, "secret") || strings.Contains(lowerKey, "token") || strings.Contains(lowerKey, "password") || strings.Contains(lowerKey, "api_key") || strings.Contains(lowerKey, "apikey") || strings.Contains(lowerKey, "authorization") {
			continue
		}
		s.state.Preferences[key] = value
	}
	return s.persist()
}

func (s *Store) MCPServers() []MCPServer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]MCPServer, len(s.state.MCPServers))
	copy(result, s.state.MCPServers)
	return result
}

func (s *Store) CreateMCPServer(server MCPServer) (MCPServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	server.ID = newID()
	server.Enabled = true
	server.Status = "configured"
	server.Settings = cloneSafeMap(server.Settings)
	server.SecretKeys = nil
	server.Tools = append([]MCPTool{}, server.Tools...)
	s.state.MCPServers = append(s.state.MCPServers, server)
	return server, s.persist()
}

func (s *Store) UpdateMCPServer(id string, patch MCPServer) (MCPServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.state.MCPServers {
		if s.state.MCPServers[i].ID != id {
			continue
		}
		current := &s.state.MCPServers[i]
		if patch.Name != "" {
			current.Name = patch.Name
		}
		if patch.Transport != "" {
			current.Transport = patch.Transport
		}
		if patch.Endpoint != "" {
			current.Endpoint = patch.Endpoint
		}
		if patch.Command != "" {
			current.Command = patch.Command
		}
		if patch.Args != nil {
			current.Args = append([]string{}, patch.Args...)
		}
		if patch.Settings != nil {
			current.Settings = cloneSafeMap(patch.Settings)
		}
		current.Enabled = patch.Enabled
		if current.Enabled {
			current.Status = "configured"
		} else {
			current.Status = "disabled"
		}
		return *current, s.persist()
	}
	return MCPServer{}, ErrNotFound
}

func (s *Store) DeleteMCPServer(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.state.MCPServers {
		if s.state.MCPServers[i].ID == id {
			s.state.MCPServers = append(s.state.MCPServers[:i], s.state.MCPServers[i+1:]...)
			return s.persist()
		}
	}
	return ErrNotFound
}

func (s *Store) Plugins() map[string]Plugin {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := map[string]Plugin{}
	for id, plugin := range s.state.Plugins {
		plugin.Settings = cloneSafeMap(plugin.Settings)
		result[id] = plugin
	}
	return result
}

func (s *Store) UpdatePlugin(id string, patch Plugin) (Plugin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plugin, ok := s.state.Plugins[id]
	if !ok {
		return Plugin{}, ErrNotFound
	}
	plugin.Enabled = patch.Enabled
	plugin.Status = "disabled"
	if plugin.Enabled {
		plugin.Status = "available"
	}
	if patch.Settings != nil {
		plugin.Settings = cloneSafeMap(patch.Settings)
	}
	s.state.Plugins[id] = plugin
	return plugin, s.persist()
}

func (s *Store) UpsertPlugin(plugin Plugin) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	plugin.Settings = cloneSafeMap(plugin.Settings)
	s.state.Plugins[plugin.ID] = plugin
	return s.persist()
}

func cloneSafeMap(values map[string]string) map[string]string {
	result := map[string]string{}
	for key, value := range values {
		lower := strings.ToLower(strings.TrimSpace(key))
		if strings.Contains(lower, "secret") || strings.Contains(lower, "token") || strings.Contains(lower, "password") || strings.Contains(lower, "api_key") || strings.Contains(lower, "apikey") || strings.Contains(lower, "authorization") || strings.Contains(lower, "credential") || strings.Contains(lower, "private_key") {
			continue
		}
		result[key] = value
	}
	return result
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
	case "spaces":
		return s.state.Spaces, nil
	case "artifacts":
		return s.state.Artifacts, nil
	default:
		return s.state.Boards, nil
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
	case "artifacts":
		s.state.Artifacts = items
	case "boards":
		s.state.Boards = items
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
