package routegraph

import (
	"fmt"
	"path"
	"strings"

	"mcp-kettel/internal/model"
)

type Node struct {
	Prefix string
	Routes []model.Candidate
}

type Edge struct {
	From   string
	To     string
	Prefix string
}

type Graph struct {
	Nodes map[string]Node
	Edges []Edge
}

func (g Graph) Resolve(roots []string) []model.Candidate {
	edges := make(map[string][]Edge)
	for _, edge := range g.Edges {
		edges[edge.From] = append(edges[edge.From], edge)
	}
	var candidates []model.Candidate
	mounted := make(map[string]bool)
	var walk func(string, string, map[string]bool)
	walk = func(id, prefix string, stack map[string]bool) {
		if stack[id] {
			return
		}
		node, ok := g.Nodes[id]
		if !ok {
			return
		}
		mounted[id] = true
		stack[id] = true
		base := join(prefix, node.Prefix)
		for _, candidate := range node.Routes {
			candidate.Route = join(base, candidate.Route)
			candidate.ToolName = model.ToolName(candidate.Method, candidate.Route)
			candidate.Description = fmt.Sprintf("Call %s %s", candidate.Method, candidate.Route)
			candidate.ID = fmt.Sprintf("%s %s@%s:%d", candidate.Method, candidate.Route, candidate.SourceFile, candidate.SourceLine)
			candidates = append(candidates, candidate)
		}
		for _, edge := range edges[id] {
			walk(edge.To, join(base, edge.Prefix), stack)
		}
		delete(stack, id)
	}
	for _, root := range roots {
		walk(root, "", make(map[string]bool))
	}
	for id := range g.Nodes {
		if !mounted[id] {
			walk(id, "", make(map[string]bool))
		}
	}
	return unique(candidates)
}

func join(parts ...string) string {
	var joined string
	for _, part := range parts {
		if part == "" || part == "/" {
			continue
		}
		joined = path.Join(joined, part)
	}
	if joined == "" {
		return "/"
	}
	if !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}
	if strings.HasSuffix(parts[len(parts)-1], "/") && joined != "/" {
		joined += "/"
	}
	return joined
}

func unique(candidates []model.Candidate) []model.Candidate {
	seen := make(map[string]bool)
	result := make([]model.Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		if !seen[candidate.ID] {
			seen[candidate.ID] = true
			result = append(result, candidate)
		}
	}
	return result
}
