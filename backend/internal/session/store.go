package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

var ErrSessionExists = errors.New("session already exists")

func (s *Store) Create(id, title string, messages []json.RawMessage) (Session, error) {
	if err := validateID(id); err != nil {
		return Session{}, err
	}
	if title == "" {
		title = "New session"
	}
	if _, err := os.Stat(s.sessionPath(id)); err == nil {
		return Session{}, ErrSessionExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return Session{}, err
	}
	fields := map[string]json.RawMessage{
		"session_id": mustJSON(id),
		"title":      mustJSON(title),
		"messages":   mustJSON(messagesOrEmpty(messages)),
		"created_at": mustJSON(time.Now().UTC().Format(time.RFC3339Nano)),
		"updated_at": mustJSON(time.Now().UTC().Format(time.RFC3339Nano)),
	}
	if err := s.writeFields(id, fields); err != nil {
		return Session{}, err
	}
	return sessionFromFields(id, fields)
}

func (s *Store) Update(id string, patch map[string]json.RawMessage) (Session, error) {
	if err := validateID(id); err != nil {
		return Session{}, err
	}
	fields, err := readObject(s.sessionPath(id))
	if err != nil {
		return Session{}, err
	}
	for key, value := range patch {
		if key == "session_id" || key == "messages" {
			continue
		}
		if !json.Valid(value) {
			return Session{}, fmt.Errorf("invalid JSON field %q", key)
		}
		fields[key] = append(json.RawMessage(nil), value...)
	}
	fields["updated_at"] = mustJSON(time.Now().UTC().Format(time.RFC3339Nano))
	if err := s.writeFields(id, fields); err != nil {
		return Session{}, err
	}
	return sessionFromFields(id, fields)
}

func (s *Store) Rename(id, title string) (Session, error) {
	if title == "" {
		return Session{}, errors.New("title required")
	}
	return s.Update(id, map[string]json.RawMessage{"title": mustJSON(title)})
}

func (s *Store) SetPinned(id string, pinned bool) (Session, error) {
	return s.Update(id, map[string]json.RawMessage{"pinned": mustJSON(pinned)})
}

func (s *Store) SetArchived(id string, archived bool) (Session, error) {
	return s.Update(id, map[string]json.RawMessage{"archived": mustJSON(archived)})
}

func (s *Store) Delete(id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	if err := os.Remove(s.sessionPath(id)); err != nil {
		return err
	}
	return s.rebuildIndex()
}

func (s *Store) AppendMessages(id string, additions ...json.RawMessage) error {
	item, err := s.Load(id)
	if err != nil {
		return err
	}
	item.Messages = append(item.Messages, additions...)
	fields, err := readObject(s.sessionPath(id))
	if err != nil {
		return err
	}
	fields["messages"] = mustJSON(item.Messages)
	fields["updated_at"] = mustJSON(time.Now().UTC().Format(time.RFC3339Nano))
	return s.writeFields(id, fields)
}

func (s *Store) sessionPath(id string) string {
	return filepath.Join(s.stateDir, "sessions", id+".json")
}

func (s *Store) writeFields(id string, fields map[string]json.RawMessage) error {
	dir := filepath.Dir(s.sessionPath(id))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(dir, ".session-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, s.sessionPath(id)); err != nil {
		return err
	}
	return s.rebuildIndex()
}

func (s *Store) rebuildIndex() error {
	list, err := s.scanSummaries()
	if err != nil {
		return err
	}
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	rows := make([]map[string]json.RawMessage, 0, len(list))
	for _, item := range list {
		fields, readErr := readObject(s.sessionPath(item.ID))
		if readErr != nil {
			continue
		}
		delete(fields, "messages")
		rows = append(rows, fields)
	}
	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	dir := filepath.Join(s.stateDir, "sessions")
	tmp, err := os.CreateTemp(dir, ".index-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, filepath.Join(dir, "_index.json"))
}

func messagesOrEmpty(messages []json.RawMessage) []json.RawMessage {
	if messages == nil {
		return []json.RawMessage{}
	}
	return messages
}

func mustJSON(value any) json.RawMessage {
	data, _ := json.Marshal(value)
	return data
}
