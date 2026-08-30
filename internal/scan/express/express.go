package express

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"mcp-kettel/internal/model"
)

var (
	routePattern = regexp.MustCompile(`(?m)^[ \t]*([A-Za-z_$][\w$]*)\.(get|post|put|patch|delete|options|head)\(\s*["']([^"']+)["']`)
	paramPattern = regexp.MustCompile(`:([A-Za-z_]\w*)`)
)

// ScanFile finds literal Express route calls without importing or executing JavaScript.
func ScanFile(root, path string) ([]model.Candidate, error) {
	extension := filepath.Ext(path)
	if extension != ".js" && extension != ".jsx" && extension != ".ts" && extension != ".tsx" && extension != ".mjs" && extension != ".cjs" {
		return nil, nil
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := string(source)
	if !strings.Contains(text, "express") {
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
	text := string(source)
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return nil, err
	}

	routes := routePattern.FindAllStringSubmatchIndex(text, -1)
	candidates := make(map[string][]model.Candidate)
	for _, match := range routes {
		owner := text[match[2]:match[3]]
		method := strings.ToUpper(text[match[4]:match[5]])
		route := text[match[6]:match[7]]
		if strings.Contains(route, "*") || strings.Contains(route, "?") {
			continue
		}
		normalized := normalizePath(route)
		parameters := make([]model.Parameter, 0)
		for _, parameter := range paramPattern.FindAllStringSubmatch(route, -1) {
			parameters = append(parameters, model.Parameter{Name: parameter[1], Location: "path", Type: "str", Required: true})
		}
		line := 1 + strings.Count(text[:match[0]], "\n")
		candidates[owner] = append(candidates[owner], model.Candidate{
			ID:          fmt.Sprintf("%s %s@%s:%d", method, normalized, filepath.ToSlash(relative), line),
			ToolName:    model.ToolName(method, normalized),
			Description: fmt.Sprintf("Call %s %s", method, normalized),
			Method:      method,
			Route:       normalized,
			SourceFile:  filepath.ToSlash(relative),
			SourceLine:  line,
			Parameters:  parameters,
		})
	}
	return candidates, nil
}

func normalizePath(route string) string { return paramPattern.ReplaceAllString(route, "{$1}") }
