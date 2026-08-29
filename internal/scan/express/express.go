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
	routePattern = regexp.MustCompile(`(?m)^[ \t]*(?:[A-Za-z_$][\w$]*|express\.Router\(\))\.(get|post|put|patch|delete|options|head)\(\s*["']([^"']+)["']`)
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
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return nil, err
	}

	routes := routePattern.FindAllStringSubmatchIndex(text, -1)
	candidates := make([]model.Candidate, 0, len(routes))
	for _, match := range routes {
		method := strings.ToUpper(text[match[2]:match[3]])
		route := text[match[4]:match[5]]
		if strings.Contains(route, "*") || strings.Contains(route, "?") {
			continue
		}
		normalized := paramPattern.ReplaceAllString(route, "{$1}")
		parameters := make([]model.Parameter, 0)
		for _, parameter := range paramPattern.FindAllStringSubmatch(route, -1) {
			parameters = append(parameters, model.Parameter{Name: parameter[1], Location: "path", Type: "str", Required: true})
		}
		line := 1 + strings.Count(text[:match[0]], "\n")
		candidates = append(candidates, model.Candidate{
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
