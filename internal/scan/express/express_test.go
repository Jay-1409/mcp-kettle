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
