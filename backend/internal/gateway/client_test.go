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
