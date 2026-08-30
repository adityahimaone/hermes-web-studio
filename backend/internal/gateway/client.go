package gateway

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	BaseURL     string
	APIKey      string
	ReadTimeout time.Duration
}

type Client struct {
	config Config
	http   *http.Client
}

type ChatRequest struct {
	SessionID string
	Message   string
	Model     string
	Provider  string
}

type Event struct {
	Name string
	Data map[string]any
}

type HTTPError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *HTTPError) Error() string { return e.Message }

func New(config Config) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 30 * time.Second
	return &Client{
		config: config,
		http:   &http.Client{Transport: transport},
	}
}

func (c *Client) BaseURL() string  { return c.config.BaseURL }
func (c *Client) Configured() bool { return strings.TrimSpace(c.config.BaseURL) != "" }

func (c *Client) Health(ctx context.Context) error {
	paths := []string{"/health", "/v1/models"}
	if c.config.APIKey != "" {
		paths = []string{"/v1/models", "/health"}
	}
	var last error
	for _, path := range paths {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.BaseURL+path, nil)
		if err != nil {
			return err
		}
		c.addHeaders(req, "")
		resp, err := c.http.Do(req)
		if err != nil {
			last = err
			continue
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized {
			return &HTTPError{StatusCode: resp.StatusCode, Code: "gateway_auth_error", Message: "Hermes Gateway rejected the configured API key."}
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		last = fmt.Errorf("gateway health returned %d", resp.StatusCode)
	}
	return last
}

func (c *Client) Stream(ctx context.Context, input ChatRequest, emit func(Event)) (string, error) {
	body := map[string]any{
		"model":    input.Model,
		"stream":   true,
		"messages": []map[string]any{{"role": "user", "content": input.Message}},
	}
	if input.Provider != "" {
		body["provider"] = input.Provider
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	streamCtx, cancel := context.WithTimeout(ctx, c.config.ReadTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(streamCtx, http.MethodPost, c.config.BaseURL+"/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	c.addHeaders(req, input.SessionID)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		if resp.StatusCode == http.StatusUnauthorized {
			return "", &HTTPError{StatusCode: resp.StatusCode, Code: "gateway_auth_error", Message: "Hermes Gateway rejected the configured API key."}
		}
		return "", &HTTPError{StatusCode: resp.StatusCode, Code: "gateway_http_error", Message: fmt.Sprintf("Hermes Gateway returned HTTP %d.", resp.StatusCode)}
	}

	return parseSSE(resp.Body, emit)
}

func (c *Client) addHeaders(req *http.Request, sessionID string) {
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
	if sessionID != "" {
		req.Header.Set("X-Hermes-Session-Id", sessionID)
		req.Header.Set("X-Hermes-Session-Key", "webui:"+sessionID)
	}
}

func parseSSE(reader io.Reader, emit func(Event)) (string, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64<<10), 2<<20)
	eventName := "message"
	var dataLines []string
	answer := ""

	flush := func() error {
		if len(dataLines) == 0 {
			eventName = "message"
			return nil
		}
		data := strings.Join(dataLines, "\n")
		dataLines = dataLines[:0]
		name := eventName
		eventName = "message"
		if data == "[DONE]" {
			return io.EOF
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			return nil
		}
		translated, delta, terminalErr := translate(name, payload)
		if terminalErr != nil {
			return terminalErr
		}
		if delta != "" {
			answer += delta
		}
		for _, item := range translated {
			emit(item)
		}
		return nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if err := flush(); errors.Is(err, io.EOF) {
				return answer, nil
			} else if err != nil {
				return answer, err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		if strings.HasPrefix(line, "event:") {
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	if err := scanner.Err(); err != nil {
		return answer, err
	}
	if err := flush(); err != nil && !errors.Is(err, io.EOF) {
		return answer, err
	}
	return answer, nil
}

func translate(sseName string, payload map[string]any) ([]Event, string, error) {
	name := stringValue(payload["event"])
	if name == "" {
		name = stringValue(payload["type"])
	}
	if name == "" || name == "message" {
		name = sseName
	}

	if choices, ok := payload["choices"].([]any); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]any); ok {
			if delta, ok := choice["delta"].(map[string]any); ok {
				if text := stringValue(delta["content"]); text != "" {
					return []Event{{Name: "token", Data: map[string]any{"text": text}}}, text, nil
				}
			}
		}
	}

	switch name {
	case "message.delta", "token":
		text := stringValue(payload["delta"])
		if text == "" {
			text = stringValue(payload["text"])
		}
		if text == "" {
			return nil, "", nil
		}
		return []Event{{Name: "token", Data: map[string]any{"text": text}}}, text, nil
	case "reasoning.available", "reasoning":
		text := stringValue(payload["delta"])
		if text == "" {
			text = stringValue(payload["text"])
		}
		return []Event{{Name: "reasoning", Data: map[string]any{"text": text}}}, "", nil
	case "tool.started", "hermes.tool.progress":
		return []Event{{Name: "tool", Data: toolData(payload, false)}}, "", nil
	case "tool.completed":
		return []Event{{Name: "tool_complete", Data: toolData(payload, true)}}, "", nil
	case "run.completed":
		output := stringValue(payload["output"])
		if output != "" {
			return []Event{{Name: "token", Data: map[string]any{"text": output}}}, output, nil
		}
		return nil, "", nil
	case "run.failed", "error":
		message := stringValue(payload["error"])
		if message == "" {
			message = stringValue(payload["message"])
		}
		if message == "" {
			message = "Hermes run failed."
		}
		return nil, "", errors.New(message)
	}
	return nil, "", nil
}

func toolData(payload map[string]any, complete bool) map[string]any {
	name := stringValue(payload["name"])
	if name == "" {
		name = stringValue(payload["tool"])
	}
	result := map[string]any{"name": name, "is_error": false}
	if tid := stringValue(payload["tid"]); tid != "" {
		result["tid"] = tid
	}
	if args, ok := payload["args"].(map[string]any); ok {
		result["args"] = args
	}
	if complete {
		result["is_error"] = payload["error"] != nil
	}
	return result
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}
