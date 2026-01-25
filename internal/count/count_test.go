package count

import (
	"os"
	"path/filepath"
	"testing"

	"pkt.systems/loc/internal/lang"
	"pkt.systems/loc/internal/model"
	"pkt.systems/loc/internal/scan"
)

func TestAggregateTotals(t *testing.T) {
	root := makeTempRoot(t)

	files := map[string]string{
		"main.go":           "package main\n\nfunc main() {}\n",
		"main_test.go":      "package main\n\nfunc TestMain(t *T) {}\nfunc ExampleMain() {}\n",
		"src/app.ts":        "export const x = 1;\n",
		"tests/app.test.ts": "describe('x', () => { it('a', () => {}) })\n",
	}

	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	registry := lang.NewRegistry()
	entries, err := scan.Walk(root, registry, nil, nil)
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	response, err := Aggregate(entries)
	if err != nil {
		t.Fatalf("aggregate: %v", err)
	}

	goStats, ok := response.Languages["go"]
	if !ok {
		t.Fatalf("missing go stats")
	}
	if goStats.TestLOC == 0 || goStats.LOC == 0 {
		t.Fatalf("expected go LOC and test LOC")
	}
	if goStats.TestCount == nil || *goStats.TestCount != 1 {
		t.Fatalf("expected go test count 1, got %v", goStats.TestCount)
	}
	if goStats.ExampleCount == nil || *goStats.ExampleCount != 1 {
		t.Fatalf("expected go example count 1, got %v", goStats.ExampleCount)
	}

	tsStats, ok := response.Languages["typescript"]
	if !ok {
		t.Fatalf("missing typescript stats")
	}
	if tsStats.TestLOC == 0 || tsStats.LOC == 0 {
		t.Fatalf("expected typescript LOC and test LOC")
	}
	if tsStats.TestCount == nil || *tsStats.TestCount != 2 {
		t.Fatalf("expected typescript test count 2, got %v", tsStats.TestCount)
	}
	if response.LOC == 0 || response.CodeLOC == 0 {
		t.Fatalf("expected totals to be non-zero")
	}
	if response.PercentTestLOC == nil || response.PercentCodeLOC == nil {
		t.Fatalf("expected percent totals to be set")
	}
}

func TestAggregateMultiLanguageRepo(t *testing.T) {
	root := makeTempRoot(t)

	files := map[string]string{
		"src/main.go":         "package main\nfunc main() {}\n",
		"src/main_test.go":    "package main\nfunc TestMain(t *T) {}\n",
		"src/app.ts":          "export const x = 1;\n",
		"tests/app.test.ts":   "test('x', () => {})\n",
		"src/lib.rs":          "pub fn lib() {}\n",
		"tests/lib.rs":        "#[test]\nfn test_lib() {}\n",
		"src/module.py":       "def func():\n    pass\n",
		"tests/test_mod.py":   "def test_func():\n    pass\n",
		"src/Thing.cs":        "public class Thing {}\n",
		"tests/ThingTests.cs": "[Fact]\npublic void TestA() {}\n",
		".git/ignored.go":     "package ignored\n",
	}

	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	registry := lang.NewRegistry()
	entries, err := scan.Walk(root, registry, nil, nil)
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	response, err := Aggregate(entries)
	if err != nil {
		t.Fatalf("aggregate: %v", err)
	}

	if _, ok := response.Languages["go"]; !ok {
		t.Fatalf("expected go stats")
	}
	if _, ok := response.Languages["typescript"]; !ok {
		t.Fatalf("expected typescript stats")
	}
	if _, ok := response.Languages["rust"]; !ok {
		t.Fatalf("expected rust stats")
	}
	if _, ok := response.Languages["python"]; !ok {
		t.Fatalf("expected python stats")
	}
	if _, ok := response.Languages["csharp"]; !ok {
		t.Fatalf("expected csharp stats")
	}
}

func TestPercentRoundingAndOmit(t *testing.T) {
	root := makeTempRoot(t)

	writeFile(t, filepath.Join(root, "main.go"), "package main\nline\nline\nline\n")
	writeFile(t, filepath.Join(root, "main_test.go"), "package main\n")

	registry := lang.NewRegistry()
	entries, err := scan.Walk(root, registry, nil, nil)
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	response, err := Aggregate(entries)
	if err != nil {
		t.Fatalf("aggregate: %v", err)
	}

	if response.PercentTestLOC == nil || response.PercentCodeLOC == nil {
		t.Fatalf("expected percent totals to be set")
	}
	if *response.PercentTestLOC != 20 || *response.PercentCodeLOC != 80 {
		t.Fatalf("unexpected total percents: %v %v", *response.PercentTestLOC, *response.PercentCodeLOC)
	}

	stats := response.Languages["go"]
	if stats.PercentTestLOC == nil || stats.PercentCodeLOC == nil {
		t.Fatalf("expected percent per-language to be set")
	}
	if *stats.PercentTestLOC != 20 || *stats.PercentCodeLOC != 80 {
		t.Fatalf("unexpected per-language percents: %v %v", *stats.PercentTestLOC, *stats.PercentCodeLOC)
	}

	response = model.CountResponse{
		Languages: map[string]model.LanguageStats{
			"go": {LOC: 0, TestLOC: 0, CodeLOC: 0},
		},
	}
	if response.PercentTestLOC != nil || response.PercentCodeLOC != nil {
		t.Fatalf("expected nil percents when total loc is zero")
	}
	stats = response.Languages["go"]
	if stats.PercentTestLOC != nil || stats.PercentCodeLOC != nil {
		t.Fatalf("expected nil percents when language loc is zero")
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
