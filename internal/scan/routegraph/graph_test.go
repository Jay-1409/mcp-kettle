package routegraph

import (
	"testing"

	"mcp-kettel/internal/model"
)

func TestResolveAccumulatesPrefixesAndStopsCycles(t *testing.T) {
	g := Graph{
		Nodes: map[string]Node{
			"app":   {},
			"api":   {Prefix: "/users"},
			"items": {Prefix: "/items", Routes: []model.Candidate{{Method: "GET", Route: "/{id}", SourceFile: "items.py", SourceLine: 1}}},
		},
		Edges: []Edge{
			{From: "app", To: "api", Prefix: "/api/v1"},
			{From: "api", To: "items"},
			{From: "items", To: "api"},
		},
	}
	candidates := g.Resolve([]string{"app"})
	if len(candidates) != 1 || candidates[0].Route != "/api/v1/users/items/{id}" {
		t.Fatalf("unexpected candidates: %#v", candidates)
	}
}
