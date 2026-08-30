package fastapi

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"

	"mcp-kettel/internal/model"
)

var (
	routePattern = regexp.MustCompile(`(?m)^[ \t]*@([A-Za-z_]\w*)\.(get|post|put|patch|delete|options|head)\(\s*["']([^"']+)["']`)
	funcPattern  = regexp.MustCompile(`(?m)^[ \t]*(?:async\s+)?def\s+([A-Za-z_]\w*)\s*\(([^)]*)\)`)
	paramPattern = regexp.MustCompile(`^([A-Za-z_]\w*)\s*(?::\s*([A-Za-z_]\w*))?\s*(?:=\s*(.+))?$`)
)

func ScanFile(root, path string) ([]model.Candidate, error) {
	if filepath.Ext(path) != ".py" {
		return nil, nil
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(string(source), "from fastapi") && !strings.Contains(string(source), "import fastapi") {
		return nil, nil
	}

	routes, err := scanRoutes(root, path, source)
	if err != nil {
		return nil, err
	}
	var candidates []model.Candidate
	for _, found := range routes {
		candidates = append(candidates, found...)
	}
	return candidates, nil
}

func scanRoutes(root, path string, source []byte) (map[string][]model.Candidate, error) {
	parser := tree_sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language())); err != nil {
		return nil, err
	}
	tree := parser.Parse(source, nil)
	if tree == nil {
		return nil, fmt.Errorf("tree-sitter returned no syntax tree")
	}
	defer tree.Close()

	relative, err := filepath.Rel(root, path)
	if err != nil {
		return nil, err
	}
	candidates := make(map[string][]model.Candidate)
	// ponytail: literal decorators and primitive parameters cover the first slice;
	// switch to Tree-sitter queries plus symbol resolution when aliases or models matter.
	walk(tree.RootNode(), func(node *tree_sitter.Node) {
		if node.Kind() != "decorated_definition" {
			return
		}
		text := string(source[node.StartByte():node.EndByte()])
		function := funcPattern.FindStringSubmatch(text)
		if function == nil {
			return
		}
		for _, route := range routePattern.FindAllStringSubmatch(text, -1) {
			parameters, ok := parseParameters(function[2], route[3])
			if !ok {
				continue
			}
			method := strings.ToUpper(route[2])
			line := int(node.StartPosition().Row) + 1
			candidates[route[1]] = append(candidates[route[1]], model.Candidate{
				ID:          fmt.Sprintf("%s %s@%s:%d", method, route[3], filepath.ToSlash(relative), line),
				ToolName:    model.ToolName(method, route[3]),
				Description: fmt.Sprintf("Call %s %s", method, route[3]),
				Method:      method,
				Route:       route[3],
				SourceFile:  filepath.ToSlash(relative),
				SourceLine:  line,
				Parameters:  parameters,
			})
		}
	})
	return candidates, nil
}

func walk(node *tree_sitter.Node, visit func(*tree_sitter.Node)) {
	visit(node)
	for i := uint(0); i < node.ChildCount(); i++ {
		if child := node.Child(i); child != nil {
			walk(child, visit)
		}
	}
}

func parseParameters(raw, route string) ([]model.Parameter, bool) {
	if strings.TrimSpace(raw) == "" {
		return nil, true
	}
	var parameters []model.Parameter
	for _, rawParam := range strings.Split(raw, ",") {
		match := paramPattern.FindStringSubmatch(strings.TrimSpace(rawParam))
		if match == nil {
			return nil, false
		}
		if match[1] == "self" || match[1] == "request" {
			continue
		}
		typeName := match[2]
		if typeName == "" {
			typeName = "str"
		}
		if typeName != "str" && typeName != "int" && typeName != "float" && typeName != "bool" {
			return nil, false
		}
		location := "query"
		if strings.Contains(route, "{"+match[1]+"}") {
			location = "path"
		}
		parameters = append(parameters, model.Parameter{
			Name:     match[1],
			Location: location,
			Type:     typeName,
			Required: match[3] == "",
		})
	}
	return parameters, true
}
