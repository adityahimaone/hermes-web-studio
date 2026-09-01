package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSanitizeKanbanProjectionBoundsEveryCollection(t *testing.T) {
	many := make([]any, 150)
	for i := range many {
		many[i] = map[string]any{"id": "t"}
	}
	got := sanitizeKanbanProjection(map[string]any{
		"tasks": many, "boards": many, "tenants": stringsToAny(150), "assignees": stringsToAny(150),
		"parents": many, "children": many, "skills": many,
	})
	encoded, _ := json.Marshal(got)
	var value map[string]any
	if err := json.Unmarshal(encoded, &value); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"tasks", "boards", "tenants", "assignees"} {
		if len(value[key].([]any)) != maxKanbanProjectionItems {
			t.Fatalf("%s unbounded: %d", key, len(value[key].([]any)))
		}
	}
	item := sanitizeKanbanItem(map[string]any{"parents": stringsToAny(150), "children": stringsToAny(150), "skills": stringsToAny(150)})
	for _, key := range []string{"parents", "children", "skills"} {
		if len(item[key].([]any)) != maxKanbanProjectionItems {
			t.Fatalf("%s unbounded", key)
		}
	}
}

func TestSanitizeKanbanProjectionRejectsNestedMaps(t *testing.T) {
	got := sanitizeKanbanItem(map[string]any{"parents": []any{map[string]any{"secret": "x"}}, "id": "t_1"})
	if _, ok := got["parents"]; ok {
		t.Fatalf("nested map projected: %#v", got)
	}
}

func TestKanbanCLIRejectsOversizedOutputBeforeJSONParse(t *testing.T) {
	data := append([]byte("{"), make([]byte, maxKanbanCLIOutput+1)...)
	if _, err := extractJSON(data); err == nil {
		t.Fatal("oversized raw output accepted")
	}
}

func TestNativeKanbanRoutesRejectHostileBoundaryValues(t *testing.T) {
	server := nativeCLITestServer(t)
	defer server.Close()
	cases := []string{
		"/api/kanban/board?board=-bad", "/api/kanban/board?board=%00bad", "/api/kanban/board?tenant=bad%3Bwhoami",
		"/api/kanban/tasks/-bad", "/api/kanban/tasks/t_1/actions/complete?board=-bad",
		"/api/kanban/dispatch?max=0", "/api/kanban/dispatch?max=999999999", "/api/kanban/dispatch?max=bad",
	}
	for _, path := range cases {
		method := http.MethodGet
		if strings.Contains(path, "/actions/") || strings.Contains(path, "/dispatch") {
			method = http.MethodPost
		}
		request, err := http.NewRequest(method, server.URL+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, response.Body)
		response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("%s status=%d", path, response.StatusCode)
		}
	}
}

func stringsToAny(n int) []any {
	out := make([]any, n)
	for i := range out {
		out[i] = "value"
	}
	return out
}
