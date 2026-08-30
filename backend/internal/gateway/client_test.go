package gateway

import (
	"strings"
	"testing"
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
