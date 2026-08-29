package generate

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"mcp-kettel/internal/model"
)

const generatedGoMod = `module generated-mcp-server

go 1.24.0

require github.com/modelcontextprotocol/go-sdk v1.6.1
`

func Write(output string, candidates []model.Candidate) error {
	if len(candidates) == 0 {
		return errors.New("no APIs selected")
	}
	if _, err := os.Stat(output); err == nil {
		return fmt.Errorf("output path already exists: %s", output)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	source, err := serverSource(candidates)
	if err != nil {
		return err
	}
	parent := filepath.Dir(output)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	temporary, err := os.MkdirTemp(parent, ".mcp-kettel-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)

	files := map[string][]byte{
		"go.mod":    []byte(generatedGoMod),
		"main.go":   source,
		"README.md": []byte("# Generated MCP server\n\nSet `MCP_API_BASE_URL` to the FastAPI service URL, then run `go run .`. Optional: set `MCP_API_AUTHORIZATION` to the complete Authorization header value.\n"),
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(temporary, name), contents, 0o644); err != nil {
			return err
		}
	}
	return os.Rename(temporary, output)
}

func serverSource(candidates []model.Candidate) ([]byte, error) {
	seen := map[string]bool{}
	for _, candidate := range candidates {
		if seen[candidate.ToolName] {
			return nil, fmt.Errorf("duplicate MCP tool name %q", candidate.ToolName)
		}
		seen[candidate.ToolName] = true
	}

	var b bytes.Buffer
	b.WriteString(`package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type toolOutput struct {
	Status int ` + "`json:\"status\"`" + `
	Body any ` + "`json:\"body\"`" + `
}

var baseURL = strings.TrimRight(os.Getenv("MCP_API_BASE_URL"), "/")

func callAPI(ctx context.Context, method, route string, query url.Values) (toolOutput, error) {
	if baseURL == "" {
		return toolOutput{}, fmt.Errorf("MCP_API_BASE_URL is required")
	}
	endpoint := baseURL + route
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return toolOutput{}, err
	}
	if authorization := os.Getenv("MCP_API_AUTHORIZATION"); authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return toolOutput{}, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 10<<20))
	if err != nil {
		return toolOutput{}, err
	}
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		decoded = string(body)
	}
	return toolOutput{Status: response.StatusCode, Body: decoded}, nil
}

`)
	for index, candidate := range candidates {
		typeName := fmt.Sprintf("tool%dInput", index+1)
		fmt.Fprintf(&b, "type %s struct {\n", typeName)
		for _, parameter := range candidate.Parameters {
			tag := parameter.Name
			typeName := goType(parameter.Type)
			if !parameter.Required {
				tag += ",omitempty"
				typeName = "*" + typeName
			}
			fmt.Fprintf(&b, "\t%s %s `json:%s jsonschema:%s`\n", goName(parameter.Name), typeName, strconv.Quote(tag), strconv.Quote(parameter.Location+" parameter"))
		}
		b.WriteString("}\n\n")
		fmt.Fprintf(&b, "func tool%d(ctx context.Context, _ *mcp.CallToolRequest, input %s) (*mcp.CallToolResult, toolOutput, error) {\n", index+1, typeName)
		fmt.Fprintf(&b, "\troute := %s\n", strconv.Quote(candidate.Route))
		b.WriteString("\tquery := url.Values{}\n")
		for _, parameter := range candidate.Parameters {
			field := "input." + goName(parameter.Name)
			if parameter.Location == "path" {
				fmt.Fprintf(&b, "\troute = strings.ReplaceAll(route, %s, url.PathEscape(fmt.Sprint(%s)))\n", strconv.Quote("{"+parameter.Name+"}"), field)
			} else {
				if parameter.Required {
					fmt.Fprintf(&b, "\tquery.Set(%s, fmt.Sprint(%s))\n", strconv.Quote(parameter.Name), field)
				} else {
					fmt.Fprintf(&b, "\tif %s != nil { query.Set(%s, fmt.Sprint(*%s)) }\n", field, strconv.Quote(parameter.Name), field)
				}
			}
		}
		fmt.Fprintf(&b, "\toutput, err := callAPI(ctx, %s, route, query)\n", strconv.Quote(candidate.Method))
		b.WriteString("\treturn nil, output, err\n}\n\n")
	}
	b.WriteString("func main() {\n")
	b.WriteString("\tserver := mcp.NewServer(&mcp.Implementation{Name: \"generated-api\", Version: \"0.1.0\"}, nil)\n")
	for index, candidate := range candidates {
		fmt.Fprintf(&b, "\tmcp.AddTool(server, &mcp.Tool{Name: %s, Description: %s}, tool%d)\n", strconv.Quote(candidate.ToolName), strconv.Quote(candidate.Description), index+1)
	}
	b.WriteString("\tif err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {\n\t\tlog.Fatal(err)\n\t}\n}\n")
	formatted, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated server: %w", err)
	}
	return formatted, nil
}

func goName(name string) string {
	var b strings.Builder
	upper := true
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			upper = true
			continue
		}
		if upper {
			r = unicode.ToUpper(r)
			upper = false
		}
		b.WriteRune(r)
	}
	result := b.String()
	runes := []rune(result)
	if result == "" || unicode.IsDigit(runes[0]) {
		return "Value" + result
	}
	return result
}

func goType(sourceType string) string {
	switch sourceType {
	case "int":
		return "int"
	case "float":
		return "float64"
	case "bool":
		return "bool"
	default:
		return "string"
	}
}
