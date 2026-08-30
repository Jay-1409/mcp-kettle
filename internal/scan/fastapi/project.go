package fastapi

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"mcp-kettel/internal/model"
	"mcp-kettel/internal/scan/routegraph"
)

var (
	declarationPattern = regexp.MustCompile(`(?m)^[ \t]*([A-Za-z_]\w*)\s*=\s*(FastAPI|APIRouter)\s*\(([^)]*)\)`)
	prefixPattern      = regexp.MustCompile(`\bprefix\s*=\s*["']([^"']*)["']`)
	includePattern     = regexp.MustCompile(`(?ms)^[ \t]*([A-Za-z_]\w*)\.include_router\(\s*([A-Za-z_]\w*(?:\.[A-Za-z_]\w*)?)(.*?)\)`)
	includePrefix      = regexp.MustCompile(`\bprefix\s*=\s*([^,\n)]+)`)
	fromImportPattern  = regexp.MustCompile(`(?m)^[ \t]*from\s+([.A-Za-z_]\w*(?:\.[A-Za-z_]\w*)*)\s+import\s+([^\n]+)`)
	constantPattern    = regexp.MustCompile(`(?m)^[ \t]*([A-Z][A-Z0-9_]*)\s*(?::[^=\n]+)?=\s*["']([^"']*)["']`)
)

type importRef struct {
	module string
	symbol string
}

type moduleInfo struct {
	name    string
	path    string
	source  []byte
	imports map[string]importRef
}

func Scan(root string, files []string) ([]model.Candidate, error) {
	modules := make([]moduleInfo, 0)
	constants := make(map[string]string)
	for _, path := range files {
		if filepath.Ext(path) != ".py" {
			continue
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		for _, match := range constantPattern.FindAllSubmatch(source, -1) {
			constants[string(match[1])] = string(match[2])
		}
		if !strings.Contains(string(source), "fastapi") && !strings.Contains(string(source), "include_router") {
			continue
		}
		modules = append(modules, moduleInfo{
			name:    moduleName(root, path),
			path:    path,
			source:  source,
			imports: parseImports(moduleName(root, path), source),
		})
	}

	graph := routegraph.Graph{Nodes: make(map[string]routegraph.Node)}
	var roots []string
	for _, module := range modules {
		for _, match := range declarationPattern.FindAllSubmatch(module.source, -1) {
			id := symbolID(module.name, string(match[1]))
			node := graph.Nodes[id]
			if prefix := prefixPattern.FindSubmatch(match[3]); prefix != nil {
				node.Prefix = string(prefix[1])
			}
			graph.Nodes[id] = node
			if string(match[2]) == "FastAPI" {
				roots = append(roots, id)
			}
		}
	}
	for index, module := range modules {
		routes, err := scanRoutes(root, module.path, module.source)
		if err != nil {
			return nil, err
		}
		for owner, candidates := range routes {
			id := symbolID(module.name, owner)
			node := graph.Nodes[id]
			node.Routes = append(node.Routes, candidates...)
			graph.Nodes[id] = node
		}
		for _, match := range includePattern.FindAllSubmatch(module.source, -1) {
			from := resolveSymbol(modules[index], string(match[1]), graph.Nodes)
			to := resolveSymbol(modules[index], string(match[2]), graph.Nodes)
			if from == "" || to == "" {
				continue
			}
			prefix := ""
			if value := includePrefix.FindSubmatch(match[3]); value != nil {
				prefix = resolvePrefix(string(value[1]), constants)
			}
			graph.Edges = append(graph.Edges, routegraph.Edge{From: from, To: to, Prefix: prefix})
		}
	}
	return graph.Resolve(roots), nil
}

func moduleName(root, path string) string {
	relative, _ := filepath.Rel(root, path)
	name := strings.TrimSuffix(filepath.ToSlash(relative), filepath.Ext(relative))
	return strings.ReplaceAll(name, "/", ".")
}

func symbolID(module, symbol string) string { return module + ":" + symbol }

func parseImports(module string, source []byte) map[string]importRef {
	imports := make(map[string]importRef)
	for _, match := range fromImportPattern.FindAllSubmatch(source, -1) {
		from := absoluteModule(module, string(match[1]))
		for _, raw := range strings.Split(string(match[2]), ",") {
			fields := strings.Fields(strings.TrimSpace(raw))
			if len(fields) == 0 {
				continue
			}
			name, alias := fields[0], fields[0]
			if len(fields) == 3 && fields[1] == "as" {
				alias = fields[2]
			}
			imports[alias] = importRef{module: from, symbol: name}
		}
	}
	return imports
}

func absoluteModule(current, imported string) string {
	if !strings.HasPrefix(imported, ".") {
		return imported
	}
	dots := len(imported) - len(strings.TrimLeft(imported, "."))
	parts := strings.Split(current, ".")
	parts = parts[:max(0, len(parts)-dots)]
	rest := strings.TrimLeft(imported, ".")
	if rest != "" {
		parts = append(parts, rest)
	}
	return strings.Join(parts, ".")
}

func resolveSymbol(module moduleInfo, expression string, nodes map[string]routegraph.Node) string {
	parts := strings.Split(expression, ".")
	if id := symbolID(module.name, parts[0]); len(parts) == 1 {
		if _, ok := nodes[id]; ok {
			return id
		}
	}
	ref, ok := module.imports[parts[0]]
	if !ok {
		return ""
	}
	if len(parts) == 1 {
		return symbolID(ref.module, ref.symbol)
	}
	return symbolID(ref.module+"."+ref.symbol, parts[1])
}

func resolvePrefix(expression string, constants map[string]string) string {
	expression = strings.TrimSpace(expression)
	if len(expression) >= 2 && (expression[0] == '\'' || expression[0] == '"') {
		return strings.Trim(expression, "\"'")
	}
	parts := strings.Split(expression, ".")
	return constants[parts[len(parts)-1]]
}
