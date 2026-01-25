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
