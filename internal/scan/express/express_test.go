package express

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFileFindsLiteralExpressRoutes(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "app.js")
	source := `const express = require("express");
const app = express();

app.get("/users/:userId", handler);
router.post('/users', handler);
app.get("/search/*", handler);
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
	if candidates[0].Route != "/users/{userId}" || candidates[0].Parameters[0].Type != "str" {
		t.Fatalf("unexpected candidate: %#v", candidates[0])
	}
}

func TestScanResolvesNestedExpressRouterMounts(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"app.js": `const express = require("express");
const apiRouter = require("./routes/api");
const app = express();
app.use("/api/:version", apiRouter);
`,
		"routes/api.js": `const express = require("express");
const usersRouter = require("./users");
const router = express.Router();
router.use("/users", usersRouter);
module.exports = router;
`,
		"routes/users.js": `import { Router } from "express";
const router = Router();
router.get("/:userId", handler);
export default router;
`,
	}
	paths := make([]string, 0, len(files))
	for relative, source := range files {
		file := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, file)
	}
	candidates, err := Scan(root, paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 || candidates[0].Route != "/api/{version}/users/{userId}" {
		t.Fatalf("unexpected candidates: %#v", candidates)
	}
	if len(candidates[0].Parameters) != 2 {
		t.Fatalf("unexpected parameters: %#v", candidates[0].Parameters)
	}
}
