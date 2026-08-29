package tui

import (
	"testing"

	"mcp-kettel/internal/model"
)

func TestGroupedItemsToggleByRoutePrefix(t *testing.T) {
	candidates := []model.Candidate{
		{ID: "1", Route: "/users/{id}", SourceFile: "users.py", Method: "GET"},
		{ID: "2", Route: "/users", SourceFile: "users.py", Method: "POST"},
		{ID: "3", Route: "/health", SourceFile: "health.py", Method: "GET"},
	}
	m := selectionModel{candidates: candidates, selected: map[string]bool{"1": true, "2": true, "3": true}, grouping: groupPath, expanded: make(map[string]bool)}
	items := m.items()
	if len(items) != 2 {
		t.Fatalf("collapsed grouped items = %d, want 2", len(items))
	}
	if got := m.groupKey(candidates[0]); got != "/users/*" {
		t.Fatalf("group key = %q, want /users/*", got)
	}
	m.expanded["/users/*"] = true
	if got := len(m.items()); got != 4 {
		t.Fatalf("expanded grouped items = %d, want 4", got)
	}
}

func TestFilterValueMatchesVisibleItemLayout(t *testing.T) {
	candidate := model.Candidate{Method: "GET", Route: "/health", SourceFile: "app.py", SourceLine: 6}
	i := item{candidate: candidate}
	if got, want := i.FilterValue(), "○  GET    /health                      app.py:6"; got != want {
		t.Fatalf("filter value = %q, want %q", got, want)
	}
}
