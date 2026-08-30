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
	if len(candidates) != 3 {
		t.Fatalf("got %d candidates, want 3", len(candidates))
	}
	if got := candidates[0].ToolName; got != "get_users_user_id" {
		t.Fatalf("tool name = %q", got)
	}
	if candidates[0].Parameters[0].Location != "path" || candidates[0].Parameters[1].Required {
		t.Fatalf("unexpected parameters: %#v", candidates[0].Parameters)
	}
	if candidates[2].Ready() {
		t.Fatal("Pydantic route should be discovered but unavailable for generation")
	}
}

func TestScanResolvesRecursiveRouterPrefixes(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"app/main.py": `from fastapi import FastAPI
from app.api.main import api_router
from app.core.config import settings
app = FastAPI()
app.include_router(api_router, prefix=settings.API_V1_STR)
`,
		"app/api/main.py": `from fastapi import APIRouter
from app.api.routes import users
api_router = APIRouter()
api_router.include_router(users.router)
`,
		"app/api/routes/users.py": `from fastapi import APIRouter
router = APIRouter(prefix="/users")
@router.get("/{user_id}")
def user(user_id: int):
    return {}
`,
		"app/core/config.py": `class Settings:
    API_V1_STR: str = "/api/v1"
settings = Settings()
`,
	}
	paths := make([]string, 0, len(files))
	for relative, source := range files {
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	candidates, err := Scan(root, paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 || candidates[0].Route != "/api/v1/users/{user_id}" {
		t.Fatalf("unexpected candidates: %#v", candidates)
	}
}
