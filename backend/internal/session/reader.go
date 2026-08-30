package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidSessionID = errors.New("invalid session id")

type LegacySessionReader struct {
	stateDir string
}

type Summary struct {
	ID       string                     `json:"session_id"`
	Title    string                     `json:"title"`
	Metadata map[string]json.RawMessage `json:"metadata"`
}

type Session struct {
	Summary
	Messages []json.RawMessage `json:"messages"`
}

func NewLegacySessionReader(stateDir string) *LegacySessionReader {
	return &LegacySessionReader{stateDir: filepath.Clean(stateDir)}
}

func ResolveStateDir(environ func(string) string, homeDir func() (string, error)) (string, error) {
	if value := strings.TrimSpace(environ("HERMES_WEBUI_STATE_DIR")); value != "" {
		return filepath.Clean(value), nil
	}
	if value := strings.TrimSpace(environ("HERMES_HOME")); value != "" {
		return filepath.Join(value, "webui"), nil
	}
	home, err := homeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".hermes", "webui"), nil
}

func (r *LegacySessionReader) List() ([]Summary, error) {
	indexPath := filepath.Join(r.stateDir, "sessions", "_index.json")
	if summaries, err := readIndex(indexPath); err == nil && len(summaries) > 0 {
		return summaries, nil
	}
	return r.scanSummaries()
}

func (r *LegacySessionReader) Load(id string) (Session, error) {
	if err := validateID(id); err != nil {
		return Session{}, err
	}
	path := filepath.Join(r.stateDir, "sessions", id+".json")
	contents, err := os.ReadFile(path)
	if err != nil {
		return Session{}, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(contents, &fields); err != nil {
		return Session{}, fmt.Errorf("decode session %q: %w", id, err)
	}
	return sessionFromFields(id, fields)
}

func (r *LegacySessionReader) scanSummaries() ([]Summary, error) {
	entries, err := os.ReadDir(filepath.Join(r.stateDir, "sessions"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Summary{}, nil
		}
		return nil, err
	}
	result := make([]Summary, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "_index.json" || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		if validateID(id) != nil {
			continue
		}
		fields, err := readObject(filepath.Join(r.stateDir, "sessions", entry.Name()))
		if err != nil {
			continue
		}
		summary, err := summaryFromFields(id, fields)
		if err == nil {
			result = append(result, summary)
		}
	}
	return result, nil
}

func readIndex(path string) ([]Summary, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(contents, &rows); err != nil {
		return nil, err
	}
	result := make([]Summary, 0, len(rows))
	for _, fields := range rows {
		id := stringField(fields, "session_id")
		if validateID(id) != nil {
			continue
		}
		summary, err := summaryFromFields(id, fields)
		if err == nil {
			result = append(result, summary)
		}
	}
	return result, nil
}

func readObject(path string) (map[string]json.RawMessage, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(contents, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

func summaryFromFields(id string, fields map[string]json.RawMessage) (Summary, error) {
	if id == "" {
		return Summary{}, ErrInvalidSessionID
	}
	metadata := make(map[string]json.RawMessage, len(fields))
	for key, value := range fields {
		if key != "messages" {
			metadata[key] = value
		}
	}
	return Summary{ID: id, Title: stringField(fields, "title"), Metadata: metadata}, nil
}

func sessionFromFields(id string, fields map[string]json.RawMessage) (Session, error) {
	summary, err := summaryFromFields(id, fields)
	if err != nil {
		return Session{}, err
	}
	var messages []json.RawMessage
	if raw, ok := fields["messages"]; ok {
		if err := json.Unmarshal(raw, &messages); err != nil {
			return Session{}, fmt.Errorf("decode messages for %q: %w", id, err)
		}
	}
	return Session{Summary: summary, Messages: messages}, nil
}

func stringField(fields map[string]json.RawMessage, key string) string {
	var value string
	_ = json.Unmarshal(fields[key], &value)
	return value
}

func validateID(id string) error {
	if id == "" || id == "." || id == ".." || filepath.Base(id) != id || strings.ContainsAny(id, `/\\`) {
		return ErrInvalidSessionID
	}
	return nil
}
