package scan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pkt.systems/loc/internal/lang"
)

func TestWalkFiltersAndExcludes(t *testing.T) {
	root := makeTempRoot(t)

	paths := []string{
		filepath.Join(root, "main.go"),
		filepath.Join(root, "main_test.go"),
		filepath.Join(root, "src", "app.ts"),
		filepath.Join(root, "tests", "app.spec.ts"),
		filepath.Join(root, "vendor", "skip.go"),
		filepath.Join(root, ".git", "ignored.go"),
		filepath.Join(root, "matlab", "script.m"),
		filepath.Join(root, "objc", "Thing.m"),
	}
	for _, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		content := []byte("line\n")
		if strings.HasSuffix(path, "script.m") {
			content = []byte("function y = f(x)\nend\n")
		}
		if strings.HasSuffix(path, "Thing.m") {
			content = []byte("@interface Thing : NSObject\n@end\n")
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	registry := lang.NewRegistry()
	entries, err := Walk(root, registry, nil, nil)
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if len(entries) != 6 {
		t.Fatalf("expected 6 entries, got %d", len(entries))
	}

	var sawMatlab bool
	var sawObjC bool
	for _, entry := range entries {
		switch entry.Language.ID {
		case "matlab":
			sawMatlab = true
		case "objective-c":
			sawObjC = true
		}
	}
	if !sawMatlab || !sawObjC {
		t.Fatalf("expected both matlab and objective-c entries")
	}

	entries, err = Walk(root, registry, []string{".go"}, nil)
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries for .go, got %d", len(entries))
	}
}

func TestWalkHonorsGitIgnore(t *testing.T) {
	root := makeTempRoot(t)

	writeFile(t, filepath.Join(root, ".gitignore"), ".cache/\nignored.go\nnested/skip.py\n")
	writeFile(t, filepath.Join(root, "main.go"), "package main\n")
	writeFile(t, filepath.Join(root, "ignored.go"), "package ignored\n")
	writeFile(t, filepath.Join(root, ".cache", "generated.py"), "def f():\n    pass\n")
	writeFile(t, filepath.Join(root, "nested", ".gitignore"), "ignored.ts\n")
	writeFile(t, filepath.Join(root, "nested", "main.py"), "def f():\n    pass\n")
	writeFile(t, filepath.Join(root, "nested", "skip.py"), "def f():\n    pass\n")
	writeFile(t, filepath.Join(root, "nested", "ignored.ts"), "export const x = 1\n")

	registry := lang.NewRegistry()
	entries, err := Walk(root, registry, nil, nil)
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	got := make(map[string]bool, len(entries))
	for _, entry := range entries {
		rel, err := filepath.Rel(root, entry.Path)
		if err != nil {
			t.Fatalf("rel: %v", err)
		}
		got[filepath.ToSlash(rel)] = true
	}

	wantPresent := []string{
		"main.go",
		"nested/main.py",
	}
	for _, path := range wantPresent {
		if !got[path] {
			t.Fatalf("expected %s to be scanned", path)
		}
	}

	wantAbsent := []string{
		"ignored.go",
		".cache/generated.py",
		"nested/skip.py",
		"nested/ignored.ts",
	}
	for _, path := range wantAbsent {
		if got[path] {
			t.Fatalf("expected %s to be excluded by .gitignore", path)
		}
	}
}

func makeTempRoot(t *testing.T) string {
	t.Helper()
	base := filepath.Join(os.TempDir(), "test.loc.pkt.systems")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir base: %v", err)
	}
	root, err := os.MkdirTemp(base, "case-")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(root)
	})
	return root
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
