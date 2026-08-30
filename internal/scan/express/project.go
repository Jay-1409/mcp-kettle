package express

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"mcp-kettel/internal/model"
	"mcp-kettel/internal/scan/routegraph"
)

var (
	declarationPattern  = regexp.MustCompile(`(?m)^[ \t]*(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*(express\s*\(|(?:express|require\(\s*["']express["']\s*\))\.Router\s*\(|Router\s*\()`)
	usePattern          = regexp.MustCompile(`(?m)^[ \t]*([A-Za-z_$][\w$]*)\.use\(\s*(?:["']([^"']*)["']\s*,\s*)?([A-Za-z_$][\w$]*(?:\.[A-Za-z_$][\w$]*)?)`)
	defaultImport       = regexp.MustCompile(`(?m)^[ \t]*import\s+([A-Za-z_$][\w$]*)\s+from\s+["']([^"']+)["']`)
	namedImport         = regexp.MustCompile(`(?m)^[ \t]*import\s*\{([^}]+)\}\s*from\s*["']([^"']+)["']`)
	namespaceImport     = regexp.MustCompile(`(?m)^[ \t]*import\s*\*\s*as\s*([A-Za-z_$][\w$]*)\s*from\s*["']([^"']+)["']`)
	defaultRequire      = regexp.MustCompile(`(?m)^[ \t]*(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*require\(\s*["']([^"']+)["']\s*\)`)
	namedRequire        = regexp.MustCompile(`(?m)^[ \t]*(?:const|let|var)\s*\{([^}]+)\}\s*=\s*require\(\s*["']([^"']+)["']\s*\)`)
	defaultExport       = regexp.MustCompile(`(?m)^[ \t]*export\s+default\s+([A-Za-z_$][\w$]*)`)
	moduleExport        = regexp.MustCompile(`(?m)^[ \t]*module\.exports\s*=\s*([A-Za-z_$][\w$]*)`)
	namedExport         = regexp.MustCompile(`(?m)^[ \t]*exports\.([A-Za-z_$][\w$]*)\s*=\s*([A-Za-z_$][\w$]*)`)
	exportList          = regexp.MustCompile(`(?m)^[ \t]*export\s*\{([^}]+)\}`)
	normalizedParameter = regexp.MustCompile(`\{([A-Za-z_]\w*)\}`)
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
	exports map[string]string
}

func Scan(root string, files []string) ([]model.Candidate, error) {
	// ponytail: literal relative imports and .use() mounts cover the common case;
	// switch to a JavaScript AST/module resolver when computed mounts matter.
	known := make(map[string]bool)
	var modules []moduleInfo
	for _, file := range files {
		if !supportedFile(file) {
			continue
		}
		source, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		if !strings.Contains(string(source), "express") {
			continue
		}
		name := moduleName(root, file)
		known[name] = true
		modules = append(modules, moduleInfo{name: name, path: file, source: source})
	}
	for index := range modules {
		modules[index].imports = parseImports(modules[index].name, modules[index].source, known)
		modules[index].exports = parseExports(modules[index].source)
	}
	moduleByName := make(map[string]moduleInfo, len(modules))
	for _, module := range modules {
		moduleByName[module.name] = module
	}

	graph := routegraph.Graph{Nodes: make(map[string]routegraph.Node)}
	var roots []string
	for _, module := range modules {
		for _, match := range declarationPattern.FindAllSubmatch(module.source, -1) {
			id := symbolID(module.name, string(match[1]))
			graph.Nodes[id] = routegraph.Node{}
			if strings.HasPrefix(strings.ReplaceAll(string(match[2]), " ", ""), "express(") {
				roots = append(roots, id)
			}
		}
	}
	for _, module := range modules {
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
		for _, match := range usePattern.FindAllSubmatch(module.source, -1) {
			from := resolveSymbol(module, string(match[1]), graph.Nodes, moduleByName)
			to := resolveSymbol(module, string(match[3]), graph.Nodes, moduleByName)
			if from != "" && to != "" {
				graph.Edges = append(graph.Edges, routegraph.Edge{From: from, To: to, Prefix: normalizePath(string(match[2]))})
			}
		}
	}
	candidates := graph.Resolve(roots)
	for index := range candidates {
		addPathParameters(&candidates[index])
	}
	return candidates, nil
}

func supportedFile(file string) bool {
	switch filepath.Ext(file) {
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs":
		return true
	default:
		return false
	}
}

func moduleName(root, file string) string {
	relative, _ := filepath.Rel(root, file)
	return strings.TrimSuffix(filepath.ToSlash(relative), filepath.Ext(relative))
}

func symbolID(module, symbol string) string { return module + ":" + symbol }

func parseImports(module string, source []byte, known map[string]bool) map[string]importRef {
	imports := make(map[string]importRef)
	add := func(alias, specifier, symbol string) {
		if target := importedModule(module, specifier, known); target != "" {
			imports[alias] = importRef{module: target, symbol: symbol}
		}
	}
	for _, match := range defaultImport.FindAllSubmatch(source, -1) {
		add(string(match[1]), string(match[2]), "default")
	}
	for _, match := range namespaceImport.FindAllSubmatch(source, -1) {
		add(string(match[1]), string(match[2]), "*")
	}
	for _, pattern := range []struct {
		regexp  *regexp.Regexp
		require bool
	}{{namedImport, false}, {namedRequire, true}} {
		for _, match := range pattern.regexp.FindAllSubmatch(source, -1) {
			specifier := string(match[2])
			for _, item := range strings.Split(string(match[1]), ",") {
				fields := strings.Fields(strings.ReplaceAll(strings.TrimSpace(item), ":", " "))
				if len(fields) == 0 {
					continue
				}
				alias := fields[0]
				if len(fields) == 3 && fields[1] == "as" || pattern.require && len(fields) == 2 {
					alias = fields[len(fields)-1]
				}
				add(alias, specifier, fields[0])
			}
		}
	}
	for _, match := range defaultRequire.FindAllSubmatch(source, -1) {
		add(string(match[1]), string(match[2]), "default")
	}
	return imports
}

func parseExports(source []byte) map[string]string {
	exports := make(map[string]string)
	for _, pattern := range []*regexp.Regexp{defaultExport, moduleExport} {
		for _, match := range pattern.FindAllSubmatch(source, -1) {
			exports["default"] = string(match[1])
		}
	}
	for _, match := range namedExport.FindAllSubmatch(source, -1) {
		exports[string(match[1])] = string(match[2])
	}
	for _, match := range exportList.FindAllSubmatch(source, -1) {
		for _, item := range strings.Split(string(match[1]), ",") {
			fields := strings.Fields(strings.TrimSpace(item))
			if len(fields) == 1 {
				exports[fields[0]] = fields[0]
			} else if len(fields) == 3 && fields[1] == "as" {
				exports[fields[2]] = fields[0]
			}
		}
	}
	return exports
}

func importedModule(current, specifier string, known map[string]bool) string {
	if !strings.HasPrefix(specifier, ".") {
		return ""
	}
	name := path.Clean(path.Join(path.Dir(current), specifier))
	if extension := path.Ext(name); supportedFile("x" + extension) {
		name = strings.TrimSuffix(name, extension)
	}
	if known[name] {
		return name
	}
	if known[name+"/index"] {
		return name + "/index"
	}
	return ""
}

func resolveSymbol(module moduleInfo, expression string, nodes map[string]routegraph.Node, modules map[string]moduleInfo) string {
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
	exported := ref.symbol
	if exported == "*" && len(parts) == 2 {
		exported = parts[1]
	}
	target := modules[ref.module].exports[exported]
	if target == "" && exported != "default" {
		target = exported
	}
	id := symbolID(ref.module, target)
	if _, ok := nodes[id]; ok {
		return id
	}
	return ""
}

func addPathParameters(candidate *model.Candidate) {
	known := make(map[string]bool)
	for _, parameter := range candidate.Parameters {
		known[parameter.Name] = true
	}
	for _, match := range normalizedParameter.FindAllStringSubmatch(candidate.Route, -1) {
		if !known[match[1]] {
			candidate.Parameters = append(candidate.Parameters, model.Parameter{Name: match[1], Location: "path", Type: "str", Required: true})
			known[match[1]] = true
		}
	}
}
