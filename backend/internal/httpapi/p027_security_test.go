package httpapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitizeKanbanProjectionDropsNestedMapsAndUnsupportedValues(t *testing.T) {
	got := sanitizeKanbanProjection(map[string]any{
		"id":       "t-1",
		"parents":  []any{"safe", map[string]any{"token": "secret"}, 42, []any{"nested"}},
		"title":    map[string]any{"secret": "nested"},
		"archived": true,
	})
	body, _ := json.Marshal(got)
	text := string(body)
	if strings.Contains(text, "secret") || strings.Contains(text, "nested") {
		t.Fatalf("nested value leaked: %s", text)
	}
	if !strings.Contains(text, `"id":"t-1"`) || !strings.Contains(text, `"archived":true`) {
		t.Fatalf("safe primitives missing: %s", text)
	}
}

func TestValidateKanbanActionReasonRejectsUnboundedAndUnsafeInput(t *testing.T) {
	if _, err := validateKanbanActionReason(strings.Repeat("x", maxKanbanActionReasonLength+1)); err == nil {
		t.Fatal("oversized reason accepted")
	}
	for _, reason := range []string{"line\nbreak", "shell;command", "$(secret)"} {
		if _, err := validateKanbanActionReason(reason); err == nil {
			t.Fatalf("unsafe reason accepted: %q", reason)
		}
	}
	if got, err := validateKanbanActionReason("valid reason: review"); err != nil || got != "valid reason: review" {
		t.Fatalf("safe reason rejected: %q %v", got, err)
	}
}
