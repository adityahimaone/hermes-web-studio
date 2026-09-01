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
	input := "event: run.failed\ndata: {\"error\":\"provider secret detail\"}\n\n"
	_, err := parseSSE(strings.NewReader(input), func(Event) {})
	if err == nil || err.Error() != "Hermes completion failed." || strings.Contains(err.Error(), "provider secret detail") {
		t.Fatalf("err = %v", err)
	}
}

func TestParseSSERejectsFailureInAnyChoice(t *testing.T) {
	input := "data: {\"choices\":[{\"delta\":{\"content\":\"safe\"}},{\"finish_reason\":\"error\",\"message\":{\"content\":\"secret provider detail\"}}]}\n\n"
	var events []Event
	_, err := parseSSE(strings.NewReader(input), func(event Event) { events = append(events, event) })
	if err == nil || !strings.Contains(err.Error(), "Hermes completion failed") || strings.Contains(err.Error(), "secret provider detail") {
		t.Fatalf("err=%v events=%#v", err, events)
	}
	if len(events) != 0 {
		t.Fatalf("events=%#v", events)
	}
}

func TestParseSSERejectsHermesFailedPayloadShapes(t *testing.T) {
	for _, payload := range []string{
		`{"hermes":{"failed":true}}`,
		`{"hermes.failed":true}`,
		`{"hermes_failed":true}`,
	} {
		_, err := parseSSE(strings.NewReader("data: "+payload+"\n\n"), func(Event) {})
		if err == nil || !strings.Contains(err.Error(), "Hermes completion failed") {
			t.Fatalf("payload=%s err=%v", payload, err)
		}
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

func TestParseSSERunCompletedChoicesDoesNotRepeatStreamedAnswer(t *testing.T) {
	input := strings.Join([]string{
		"data: {\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}", "",
		"event: run.completed", "data: {\"choices\":[{\"delta\":{\"content\":\"Hello world\"}}]}", "",
		"data: [DONE]", "",
	}, "\n")
	var events []Event
	answer, err := parseSSE(strings.NewReader(input), func(event Event) { events = append(events, event) })
	if err != nil || answer != "Hello world" || len(events) != 2 || events[1].Data["text"] != " world" {
		t.Fatalf("answer=%q events=%#v err=%v", answer, events, err)
	}
}

func TestParseSSERunCompletedPayloadEventDoesNotRepeatStreamedAnswer(t *testing.T) {
	input := strings.Join([]string{
		"data: {\"event\":\"message.delta\",\"delta\":\"RAW_RUNS\"}", "",
		"data: {\"event\":\"run.completed\",\"output\":\"RAW_RUNS\"}", "",
		": stream closed", "",
	}, "\n")
	var events []Event
	answer, err := parseSSE(strings.NewReader(input), func(event Event) { events = append(events, event) })
	if err != nil || answer != "RAW_RUNS" || len(events) != 1 || events[0].Data["text"] != "RAW_RUNS" {
		t.Fatalf("answer=%q events=%#v err=%v", answer, events, err)
	}
}

func TestParseSSEReasoningSnapshotDoesNotRepeatTokenAnswer(t *testing.T) {
	input := strings.Join([]string{
		"data: {\"event\":\"message.delta\",\"delta\":\"\\n\\nanswer\"}", "",
		"data: {\"event\":\"reasoning.available\",\"text\":\"answer\"}", "",
		"data: {\"event\":\"message.delta\",\"delta\":\"answer\"}", "",
		"data: {\"event\":\"run.completed\",\"output\":\"answer\"}", "",
	}, "\n")
	var events []Event
	answer, err := parseSSE(strings.NewReader(input), func(event Event) { events = append(events, event) })
	if err != nil || answer != "\n\nanswer" || len(events) != 2 {
		t.Fatalf("answer=%q events=%#v err=%v", answer, events, err)
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

func TestParseSSETranslatesCanonicalRunsActivityEvents(t *testing.T) {
	input := strings.Join([]string{
		"event: tool.start", "data: {\"tool_id\":\"t1\",\"tool_name\":\"terminal\"}", "",
		"event: tool.progress", "data: {\"tool_id\":\"t1\",\"tool_name\":\"terminal\",\"preview\":\"working\"}", "",
		"event: tool.complete", "data: {\"tool_id\":\"t1\",\"tool_name\":\"terminal\",\"summary\":\"finished\"}", "",
		"event: subagent.start", "data: {\"subagent_id\":\"s1\",\"goal\":\"research\"}", "",
		"event: subagent.complete", "data: {\"subagent_id\":\"s1\",\"summary\":\"done\"}", "",
		"event: reasoning.delta", "data: {\"text\":\"thinking\"}", "",
	}, "\n")
	var events []Event
	_, err := parseSSE(strings.NewReader(input), func(event Event) { events = append(events, event) })
	if err != nil || len(events) != 6 {
		t.Fatalf("events=%#v err=%v", events, err)
	}
	if events[0].Name != "tool" || events[1].Name != "tool" || events[2].Name != "tool_complete" || events[3].Name != "subagent" || events[4].Data["status"] != "complete" || events[5].Name != "reasoning" {
		t.Fatalf("events=%#v", events)
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

func TestRunStreamStartsRunAndReadsEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/runs" {
			if r.Method != http.MethodPost || r.Header.Get("X-Hermes-Session-Id") != "session-1" {
				t.Fatalf("run request method=%s session=%q", r.Method, r.Header.Get("X-Hermes-Session-Id"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"run_id":"run-1"}`))
			return
		}
		if r.URL.Path == "/v1/runs/run-1/events" {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("event: message.delta\ndata: {\"delta\":\"runs\"}\n\ndata: [DONE]\n\n"))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	client := New(Config{BaseURL: server.URL, ReadTimeout: time.Second})
	var events []Event
	answer, err := client.RunStream(context.Background(), ChatRequest{SessionID: "session-1", Message: "hello", Model: "default"}, func(event Event) { events = append(events, event) })
	if err != nil || answer != "runs" || len(events) != 1 || events[0].Name != "token" {
		t.Fatalf("answer=%q events=%#v err=%v", answer, events, err)
	}
}

func authErrString(err error) string {
	if typed, ok := err.(*HTTPError); ok {
		return typed.Code
	}
	return ""
}
