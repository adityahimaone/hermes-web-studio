package gateway

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
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

const maxCatalogModels = 1000

type Model struct {
	ID       string   `json:"id"`
	Provider string   `json:"provider,omitempty"`
	Aliases  []string `json:"aliases,omitempty"`
}

func (c *Client) Models(ctx context.Context) ([]Model, error) {
	baseURL, err := safeBaseURL(c.config.BaseURL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req, "")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Code: "models_unavailable", Message: "Hermes Gateway model catalog is unavailable."}
	}
	var payload struct {
		Data []struct {
			ID       string   `json:"id"`
			Provider string   `json:"provider"`
			OwnedBy  string   `json:"owned_by"`
			Aliases  []string `json:"aliases"`
		} `json:"data"`
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
	if err != nil || len(body) > 1<<20 || json.NewDecoder(bytes.NewReader(body)).Decode(&payload) != nil {
		return nil, errors.New("Hermes Gateway returned an invalid model catalog")
	}
	if len(payload.Data) > maxCatalogModels {
		return nil, errors.New("Hermes Gateway returned too many models")
	}
	models := make([]Model, 0, len(payload.Data))
	seen := make(map[string]bool)
	for _, item := range payload.Data {
		id := cleanCatalogText(item.ID, 256)
		if id == "" {
			continue
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		provider := cleanCatalogText(item.Provider, 128)
		if provider == "" {
			provider = cleanCatalogText(item.OwnedBy, 128)
		}
		aliases := make([]string, 0, len(item.Aliases))
		aliasSeen := map[string]bool{}
		for _, alias := range item.Aliases {
			alias = cleanCatalogText(alias, 128)
			if alias != "" && !aliasSeen[alias] {
				aliases = append(aliases, alias)
				aliasSeen[alias] = true
			}
		}
		models = append(models, Model{ID: id, Provider: provider, Aliases: aliases})
	}
	return models, nil
}

type ChatRequest struct {
	SessionID   string
	Message     string
	Model       string
	Provider    string
	Attachments []Attachment
}

type Attachment struct {
	Name string
	MIME string
	Data []byte
}

func (c *Client) ResolveApproval(ctx context.Context, runID, choice string) error {
	if strings.TrimSpace(runID) == "" || (choice != "once" && choice != "session" && choice != "always" && choice != "deny") {
		return errors.New("invalid approval decision")
	}
	body, err := json.Marshal(map[string]string{"choice": choice})
	if err != nil {
		return err
	}
	baseURL, err := safeBaseURL(c.config.BaseURL)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/runs/"+url.PathEscape(runID)+"/approval", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.addHeaders(req, "")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{StatusCode: resp.StatusCode, Code: "approval_failed", Message: "Hermes rejected the approval decision."}
	}
	return nil
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
		http:   &http.Client{Transport: transport, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return errors.New("gateway redirect rejected") }},
	}
}

func safeBaseURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", errors.New("invalid Hermes Gateway URL")
	}
	host := u.Hostname()
	ip := net.ParseIP(host)
	if host != "localhost" && ip == nil {
		return "", errors.New("Hermes Gateway URL must target local host")
	}
	if ip != nil && !(ip.IsLoopback()) {
		return "", errors.New("Hermes Gateway URL must target local host")
	}
	return strings.TrimRight(u.String(), "/"), nil
}

func cleanCatalogText(value string, limit int) string {
	value = strings.Map(func(r rune) rune {
		if r < 0x20 || (r >= 0x7f && r <= 0x9f) {
			return -1
		}
		return r
	}, strings.TrimSpace(value))
	value = strings.ToValidUTF8(value, "")
	if len(value) > limit {
		value = value[:limit]
		for !utf8.ValidString(value) {
			value = value[:len(value)-1]
		}
	}
	return value
}

func (c *Client) BaseURL() string  { return c.config.BaseURL }
func (c *Client) Configured() bool { return strings.TrimSpace(c.config.BaseURL) != "" }

func (c *Client) Health(ctx context.Context) error {
	baseURL, err := safeBaseURL(c.config.BaseURL)
	if err != nil {
		return err
	}
	paths := []string{"/health", "/v1/models"}
	if c.config.APIKey != "" {
		paths = []string{"/v1/models", "/health"}
	}
	var last error
	for _, path := range paths {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
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
	baseURL, err := safeBaseURL(c.config.BaseURL)
	if err != nil {
		return "", err
	}
	content := any(input.Message)
	if len(input.Attachments) > 0 {
		parts := []map[string]any{{"type": "text", "text": input.Message}}
		for _, attachment := range input.Attachments {
			encoded := base64.StdEncoding.EncodeToString(attachment.Data)
			switch {
			case strings.HasPrefix(attachment.MIME, "image/"):
				parts = append(parts, map[string]any{"type": "image_url", "image_url": map[string]string{"url": "data:" + attachment.MIME + ";base64," + encoded}})
			case attachment.MIME == "application/pdf":
				parts = append(parts, map[string]any{"type": "file", "file": map[string]string{"filename": attachment.Name, "file_data": "data:application/pdf;base64," + encoded}})
			default:
				parts = append(parts, map[string]any{"type": "text", "text": string(attachment.Data)})
			}
		}
		content = parts
	}
	body := map[string]any{
		"model":    input.Model,
		"stream":   true,
		"messages": []map[string]any{{"role": "user", "content": content}},
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
	req, err := http.NewRequestWithContext(streamCtx, http.MethodPost, baseURL+"/v1/chat/completions", bytes.NewReader(payload))
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

func (c *Client) RunStream(ctx context.Context, input ChatRequest, emit func(Event)) (string, error) {
	baseURL, err := safeBaseURL(c.config.BaseURL)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(map[string]any{"input": input.Message, "session_id": input.SessionID, "model": input.Model, "provider": input.Provider})
	if err != nil {
		return "", err
	}
	requestContext, cancel := context.WithTimeout(ctx, c.config.ReadTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestContext, http.MethodPost, baseURL+"/v1/runs", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	c.addHeaders(req, input.SessionID)
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &HTTPError{StatusCode: resp.StatusCode, Code: "gateway_run_error", Message: fmt.Sprintf("Hermes Gateway returned HTTP %d while starting the run.", resp.StatusCode)}
	}
	var started struct {
		RunID string `json:"run_id"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&started); err != nil || strings.TrimSpace(started.RunID) == "" {
		return "", errors.New("Hermes Gateway returned no run ID")
	}
	eventsURL := baseURL + "/v1/runs/" + url.PathEscape(started.RunID) + "/events"
	eventRequest, err := http.NewRequestWithContext(requestContext, http.MethodGet, eventsURL, nil)
	if err != nil {
		return "", err
	}
	eventRequest.Header.Set("Accept", "text/event-stream")
	c.addHeaders(eventRequest, input.SessionID)
	eventResponse, err := c.http.Do(eventRequest)
	if err != nil {
		return "", err
	}
	defer eventResponse.Body.Close()
	if eventResponse.StatusCode < 200 || eventResponse.StatusCode >= 300 {
		return "", &HTTPError{StatusCode: eventResponse.StatusCode, Code: "gateway_run_stream_error", Message: "Hermes Gateway could not open the run event stream."}
	}
	answer, err := parseSSE(eventResponse.Body, emit)
	return answer, err
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
	reasoningSnapshot := ""

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
		effectiveName := name
		if payloadName := stringValue(payload["event"]); payloadName != "" {
			effectiveName = payloadName
		}
		if payloadName := stringValue(payload["type"]); payloadName != "" {
			effectiveName = payloadName
		}
		if (effectiveName == "run.completed" || effectiveName == "message.complete") && delta != "" {
			if sameSnapshot(answer, delta) {
				delta = ""
			} else {
				delta = missingSuffix(answer, delta)
			}
			if len(translated) > 0 && translated[0].Name == "token" {
				if delta == "" {
					translated = nil
				} else {
					translated[0].Data["text"] = delta
				}
			}
		}
		if len(translated) > 0 && translated[0].Name == "reasoning" {
			reasoningSnapshot = stringValue(translated[0].Data["text"])
		}
		if len(translated) > 0 && translated[0].Name == "token" && reasoningSnapshot != "" && sameSnapshot(delta, reasoningSnapshot) {
			delta = missingSuffix(answer, delta)
			if sameSnapshot(answer, delta) {
				delta = ""
			}
			if delta == "" {
				translated = nil
			} else {
				translated[0].Data["text"] = delta
			}
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

// Gateways may send token deltas and then repeat the complete answer in
// run.completed. Only emit the suffix that was not already streamed.
func missingSuffix(current, completed string) string {
	if completed == current || strings.HasPrefix(current, completed) {
		return ""
	}
	if strings.HasPrefix(completed, current) {
		return strings.TrimPrefix(completed, current)
	}
	return completed
}

func translate(sseName string, payload map[string]any) ([]Event, string, error) {
	name := stringValue(payload["event"])
	if name == "" {
		name = stringValue(payload["type"])
	}
	if name == "" || name == "message" {
		name = sseName
	}
	if name == "run.completed" || name == "message.complete" {
		completed := stringValue(payload["output"])
		if completed == "" {
			completed = stringValue(payload["text"])
		}
		if completed == "" {
			completed = choiceText(payload)
		}
		if completed != "" {
			return []Event{{Name: "token", Data: map[string]any{"text": completed}}}, completed, nil
		}
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
	case "reasoning.available", "reasoning", "reasoning.delta", "thinking.delta":
		text := stringValue(payload["delta"])
		if text == "" {
			text = stringValue(payload["text"])
		}
		return []Event{{Name: "reasoning", Data: map[string]any{"text": text}}}, "", nil
	case "tool.start", "tool.started", "tool.progress", "hermes.tool.progress":
		return []Event{{Name: "tool", Data: toolData(payload, false)}}, "", nil
	case "tool.complete", "tool.completed":
		return []Event{{Name: "tool_complete", Data: toolData(payload, true)}}, "", nil
	case "subagent.spawn_requested", "subagent.start", "subagent.started", "subagent.thinking", "subagent.tool", "subagent.progress", "subagent.complete", "subagent.completed", "subagent.failed":
		return []Event{{Name: "subagent", Data: subagentData(payload, name)}}, "", nil
	case "approval.request", "hermes.approval.request", "approval.requested", "approval.required", "approval":
		return []Event{{Name: "approval", Data: approvalData(payload)}}, "", nil
	case "usage", "run.usage", "usage.updated":
		return []Event{{Name: "usage", Data: usageData(payload)}}, "", nil
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

func choiceText(payload map[string]any) string {
	choices, ok := payload["choices"].([]any)
	if !ok || len(choices) == 0 {
		return ""
	}
	choice, ok := choices[0].(map[string]any)
	if !ok {
		return ""
	}
	if delta, ok := choice["delta"].(map[string]any); ok {
		return stringValue(delta["content"])
	}
	if message, ok := choice["message"].(map[string]any); ok {
		return stringValue(message["content"])
	}
	return ""
}

func sameSnapshot(left, right string) bool {
	return strings.TrimSpace(left) != "" && strings.TrimSpace(left) == strings.TrimSpace(right)
}

func toolData(payload map[string]any, complete bool) map[string]any {
	name := stringValue(payload["name"])
	if name == "" {
		name = stringValue(payload["tool"])
	}
	if name == "" {
		name = stringValue(payload["tool_name"])
	}
	result := map[string]any{"name": name, "is_error": false}
	tid := stringValue(payload["tid"])
	if tid == "" {
		tid = stringValue(payload["tool_id"])
	}
	if tid != "" {
		result["tid"] = tid
	}
	if args, ok := payload["args"].(map[string]any); ok {
		result["args"] = redact(args)
	}
	if complete {
		result["is_error"] = payload["error"] != nil || boolValue(payload["is_error"])
		if summary := stringValue(payload["summary"]); summary != "" {
			result["result"] = summary
		}
	} else if preview := stringValue(payload["preview"]); preview != "" {
		result["result"] = preview
	}
	return result
}

func subagentData(payload map[string]any, status string) map[string]any {
	id := stringValue(payload["id"])
	if id == "" {
		id = stringValue(payload["subagent_id"])
	}
	name := stringValue(payload["name"])
	if name == "" {
		name = stringValue(payload["goal"])
	}
	state := "running"
	if status == "subagent.complete" || status == "subagent.completed" || status == "subagent.failed" {
		state = "complete"
		if status == "subagent.failed" {
			state = "error"
		}
	}
	return map[string]any{"id": id, "name": name, "status": state, "task": stringValue(payload["task"]), "result": redact(payload["result"])}
}

func approvalData(payload map[string]any) map[string]any {
	id := stringValue(payload["id"])
	if id == "" {
		id = stringValue(payload["run_id"])
	}
	return map[string]any{"id": id, "kind": stringValue(payload["kind"]), "command": redact(payload["command"]), "status": stringValue(payload["status"]), "options": redact(payload["options"])}
}

func usageData(payload map[string]any) map[string]any {
	result := map[string]any{}
	for _, key := range []string{"prompt_tokens", "completion_tokens", "total_tokens", "input_tokens", "output_tokens", "context_window"} {
		if value, ok := payload[key]; ok {
			result[key] = value
		}
	}
	if nested, ok := payload["usage"].(map[string]any); ok {
		for key, value := range usageData(nested) {
			result[key] = value
		}
	}
	return result
}

func redact(value any) any {
	switch item := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(item))
		for key, value := range item {
			lower := strings.ToLower(key)
			if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "password") || strings.Contains(lower, "api_key") || lower == "authorization" {
				out[key] = "[redacted]"
				continue
			}
			out[key] = redact(value)
		}
		return out
	case []any:
		out := make([]any, len(item))
		for i, value := range item {
			out[i] = redact(value)
		}
		return out
	default:
		return value
	}
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func boolValue(value any) bool {
	flag, _ := value.(bool)
	return flag
}
