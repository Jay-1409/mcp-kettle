package fastapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFileFindsLiteralPrimitiveRoutes(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "app.py")
	source := `from fastapi import FastAPI
app = FastAPI()

@app.get("/users/{user_id}")
async def user(user_id: int, verbose: bool = False):
    return {}

@app.delete('/users/{user_id}')
def remove(user_id: int):
    return {}

@app.post("/users")
def create(user: User):
    return user
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	candidates, err := ScanFile(root, path)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 {
		t.Fatalf("got %d candidates, want 2", len(candidates))
	}
	if got := candidates[0].ToolName; got != "get_users_user_id" {
		t.Fatalf("tool name = %q", got)
	}
	if candidates[0].Parameters[0].Location != "path" || candidates[0].Parameters[1].Required {
		t.Fatalf("unexpected parameters: %#v", candidates[0].Parameters)
	}
}
