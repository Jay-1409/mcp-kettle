package scan

import (
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"mcp-kettel/internal/model"
)

type Scanner func(root string, files []string) ([]model.Candidate, error)

type FileScanner func(root, path string) ([]model.Candidate, error)

func Files(scanner FileScanner) Scanner {
	return func(root string, files []string) ([]model.Candidate, error) {
		var candidates []model.Candidate
		for _, path := range files {
			found, err := scanner(root, path)
			if err != nil {
				return nil, fmt.Errorf("scan %s: %w", path, err)
			}
			candidates = append(candidates, found...)
		}
		return candidates, nil
	}
}

func Directory(root string, scanners ...Scanner) ([]model.Candidate, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	files, err := sourceFiles(root)
	if err != nil {
		return nil, err
	}

	var candidates []model.Candidate
	for _, scanner := range scanners {
		found, err := scanner(root, files)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, found...)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
	return candidates, nil
}

func sourceFiles(root string) ([]string, error) {
	if files, err := gitFiles(root); err == nil {
		return files, nil
	}

	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && path != root && skippedDir(entry.Name()) {
			return filepath.SkipDir
		}
		if !entry.IsDir() && supportedExtension(filepath.Ext(path)) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func gitFiles(root string) ([]string, error) {
	output, err := exec.Command("git", "-C", root, "ls-files", "-co", "--exclude-standard", "-z", "--", "*.py", "*.js", "*.jsx", "*.ts", "*.tsx", "*.mjs", "*.cjs").Output()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, relative := range strings.Split(string(output), "\x00") {
		if relative != "" {
			files = append(files, filepath.Join(root, filepath.FromSlash(relative)))
		}
	}
	return files, nil
}

func supportedExtension(extension string) bool {
	switch extension {
	case ".py", ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs":
		return true
	default:
		return false
	}
}

func skippedDir(name string) bool {
	switch name {
	case ".git", ".venv", "venv", "node_modules", "dist", "build", "__pycache__":
		return true
	default:
		return false
	}
}
