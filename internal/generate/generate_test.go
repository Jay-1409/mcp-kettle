package generate

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mcp-kettel/internal/model"
)

func TestWriteGeneratesSyntacticallyValidSelectedTool(t *testing.T) {
	output := filepath.Join(t.TempDir(), "server")
	candidate := model.Candidate{
		ToolName:    "get_users_user_id",
		Description: "Call GET /users/{user_id}",
		Method:      "GET",
		Route:       "/users/{user_id}",
		Parameters:  []model.Parameter{{Name: "user_id", Location: "path", Type: "int", Required: true}},
	}
	if err := Write(output, []model.Candidate{candidate}); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join(output, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "main.go", source, parser.AllErrors); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), `Name: "get_users_user_id"`) {
		t.Fatal("generated tool is missing")
	}
}
