package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParseSSETranslatesOpenAIAndHermesFrames(t *testing.T) {
	input := strings.Join([]string{
		"data: {\"choices\":[{\"delta\":{\"content\":\"Hello \"}}]}", "",
		"event: message.delta", "data: {\"delta\":\"Hermes\"}", "",
		"event: tool.started", "data: {\"name\":\"terminal\",\"tid\":\"t1\"}", "",
		"event: tool.completed", "data: {\"name\":\"terminal\",\"tid\":\"t1\"}", "",
		"data: [DONE]", "",
	}, "\n")
	var events []Event
	answer, err := parseSSE(strings.NewReader(input), func(event Event) { events = append(events, event) })
	if err != nil {
		t.Fatal(err)
	}
	if answer != "Hello Hermes" {
		t.Fatalf("answer = %q", answer)
	}
	if len(events) != 4 {
		t.Fatalf("events = %#v", events)
	}
	if events[0].Name != "token" || events[2].Name != "tool" || events[3].Name != "tool_complete" {
		t.Fatalf("unexpected translated events: %#v", events)
	}
}

func TestParseSSESurfacesRunFailure(t *testing.T) {
	input := "event: run.failed\ndata: {\"error\":\"provider failed\"}\n\n"
	_, err := parseSSE(strings.NewReader(input), func(Event) {})
	if err == nil || !strings.Contains(err.Error(), "provider failed") {
		t.Fatalf("err = %v", err)
	}
}

func TestParseSSERunCompletedDoesNotRepeatStreamedAnswer(t *testing.T) {
	input := strings.Join([]string{
		"data: {\"choices\":[{\"delta\":{\"content\":\"Hello Hermes\"}}]}", "",
		"event: run.completed", "data: {\"output\":\"Hello Hermes\"}", "",
		"data: [DONE]", "",
	}, "\n")
	var events []Event
	answer, err := parseSSE(strings.NewReader(input), func(event Event) { events = append(events, event) })
	if err != nil {
		t.Fatal(err)
	}
	if answer != "Hello Hermes" {
		t.Fatalf("answer = %q", answer)
	}
	if len(events) != 1 || events[0].Data["text"] != "Hello Hermes" {
		t.Fatalf("events = %#v", events)
	}
}

func TestParseSSETranslatesActivityAndRedactsSecrets(t *testing.T) {
	input := strings.Join([]string{
		"event: subagent.started", "data: {\"id\":\"s1\",\"name\":\"research\"}", "",
		"event: approval.request", "data: {\"run_id\":\"r1\",\"command\":\"echo hi\",\"api_key\":\"do-not-send\"}", "",
		"event: usage", "data: {\"usage\":{\"prompt_tokens\":4,\"completion_tokens\":3,\"total_tokens\":7}}", "",
	}, "\n")
	var events []Event
	_, err := parseSSE(strings.NewReader(input), func(event Event) { events = append(events, event) })
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 || events[0].Name != "subagent" || events[1].Name != "approval" || events[2].Name != "usage" {
		t.Fatalf("events=%#v", events)
	}
	if events[1].Data["id"] != "r1" || events[1].Data["api_key"] != nil {
		t.Fatalf("approval=%#v", events[1].Data)
	}
	usage, ok := events[2].Data["total_tokens"]
	if !ok || usage != float64(7) {
		t.Fatalf("usage=%#v", events[2].Data)
	}
}

func TestHealthRejectsUnauthorizedModelProbe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	client := New(Config{BaseURL: server.URL, APIKey: "wrong", ReadTimeout: time.Second})
	err := client.Health(context.Background())
	if err == nil || !strings.Contains(err.Error(), "rejected") {
		t.Fatalf("expected safe auth error, got %v", err)
	}
	if !strings.Contains(err.Error(), "API key") || !strings.Contains(authErrString(err), "gateway_auth_error") {
		t.Fatalf("unexpected auth error: %v", err)
	}
}

func authErrString(err error) string {
	if typed, ok := err.(*HTTPError); ok {
		return typed.Code
	}
	return ""
}
