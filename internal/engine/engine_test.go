package engine

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pkt.systems/loc/internal/lang"
	"pkt.systems/loc/internal/model"
)

func TestCountUnsupportedExtension(t *testing.T) {
	registry := lang.NewRegistry()
	counter := NewCounter(registry)

	_, err := counter.Count(context.Background(), model.CountRequest{
		Root:       ".",
		Extensions: []string{".go", ".nope"},
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported extensions") {
		t.Fatalf("expected unsupported extension error, got %v", err)
	}
}

func TestCountExtensionExpansion(t *testing.T) {
	root := makeTempRoot(t)
	writeFile(t, filepath.Join(root, "Thing.h"), "#import <Foundation/Foundation.h>\n")
	writeFile(t, filepath.Join(root, "Thing.m"), "@implementation Thing\n@end\n")
	writeFile(t, filepath.Join(root, "Thing.mm"), "@implementation Thing\n@end\n")
	writeFile(t, filepath.Join(root, "Thing.c"), "int main() { return 0; }\n")

	registry := lang.NewRegistry()
	counter := NewCounter(registry)

	resp, err := counter.Count(context.Background(), model.CountRequest{
		Root:       root,
		Extensions: []string{".h"},
	})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if _, ok := resp.Languages["c"]; !ok {
		t.Fatalf("expected c stats")
	}
	if _, ok := resp.Languages["objective-c"]; !ok {
		t.Fatalf("expected objective-c stats")
	}
}

func TestCountExtensionExpansionMatlabObjectiveC(t *testing.T) {
	root := makeTempRoot(t)
	writeFile(t, filepath.Join(root, "script.m"), "function y = f(x)\nend\n")
	writeFile(t, filepath.Join(root, "Thing.m"), "@interface Thing\n@end\n")

	registry := lang.NewRegistry()
	counter := NewCounter(registry)

	resp, err := counter.Count(context.Background(), model.CountRequest{
		Root:       root,
		Extensions: []string{".m"},
	})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if _, ok := resp.Languages["matlab"]; !ok {
		t.Fatalf("expected matlab stats")
	}
	if _, ok := resp.Languages["objective-c"]; !ok {
		t.Fatalf("expected objective-c stats")
	}
}

func TestExtensionExpansionLanguageSets(t *testing.T) {
	cases := []struct {
		name    string
		filter  string
		langID  string
		files   map[string]string
		wantLOC int
	}{
		{
			name:   "javascript",
			filter: ".js",
			langID: "javascript",
			files: map[string]string{
				"app.js":  "x\n",
				"app.jsx": "x\n",
				"app.mjs": "x\n",
				"app.cjs": "x\n",
			},
			wantLOC: 4,
		},
		{
			name:   "typescript",
			filter: ".ts",
			langID: "typescript",
			files: map[string]string{
				"app.ts":  "x\n",
				"app.tsx": "x\n",
				"app.mts": "x\n",
				"app.cts": "x\n",
			},
			wantLOC: 4,
		},
		{
			name:   "kotlin",
			filter: ".kt",
			langID: "kotlin",
			files: map[string]string{
				"app.kt":  "x\n",
				"app.kts": "x\n",
			},
			wantLOC: 2,
		},
		{
			name:   "perl",
			filter: ".pl",
			langID: "perl",
			files: map[string]string{
				"app.pl":  "x\n",
				"app.pm":  "x\n",
				"t/app.t": "x\n",
			},
			wantLOC: 3,
		},
		{
			name:   "r",
			filter: ".R",
			langID: "r",
			files: map[string]string{
				"app.R": "x\n",
				"app.r": "x\n",
			},
			wantLOC: 2,
		},
		{
			name:   "cpp",
			filter: ".cpp",
			langID: "cpp",
			files: map[string]string{
				"app.cpp": "x\n",
				"app.cc":  "x\n",
				"app.cxx": "x\n",
				"app.hpp": "x\n",
				"app.hh":  "x\n",
				"app.hxx": "x\n",
			},
			wantLOC: 6,
		},
		{
			name:   "shell",
			filter: ".sh",
			langID: "shell",
			files: map[string]string{
				"run.sh":   "x\n",
				"run.bash": "x\n",
				"run.zsh":  "x\n",
			},
			wantLOC: 3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := makeTempRoot(t)
			for name, content := range tc.files {
				writeFile(t, filepath.Join(root, filepath.FromSlash(name)), content)
			}
			registry := lang.NewRegistry()
			counter := NewCounter(registry)

			resp, err := counter.Count(context.Background(), model.CountRequest{
				Root:       root,
				Extensions: []string{tc.filter},
			})
			if err != nil {
				t.Fatalf("count: %v", err)
			}
			stats, ok := resp.Languages[tc.langID]
			if !ok {
				t.Fatalf("expected %s stats", tc.langID)
			}
			if stats.LOC != tc.wantLOC {
				t.Fatalf("expected %s loc %d, got %d", tc.langID, tc.wantLOC, stats.LOC)
			}
		})
	}
}

func TestExtensionExpansionAmbiguousHeaders(t *testing.T) {
	root := makeTempRoot(t)
	writeFile(t, filepath.Join(root, "CThing.h"), "#define THING_H\n")
	writeFile(t, filepath.Join(root, "ObjcThing.h"), "@interface Thing : NSObject\n")
	writeFile(t, filepath.Join(root, "Thing.c"), "int main() { return 0; }\n")
	writeFile(t, filepath.Join(root, "Thing.mm"), "@implementation Thing\n")
	writeFile(t, filepath.Join(root, "Thing.m"), "@implementation Thing\n")

	registry := lang.NewRegistry()
	counter := NewCounter(registry)

	resp, err := counter.Count(context.Background(), model.CountRequest{
		Root:       root,
		Extensions: []string{".h"},
	})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	cStats, ok := resp.Languages["c"]
	if !ok || cStats.LOC != 2 {
		t.Fatalf("expected c loc 2, got %+v", cStats)
	}
	objcStats, ok := resp.Languages["objective-c"]
	if !ok || objcStats.LOC != 3 {
		t.Fatalf("expected objective-c loc 3, got %+v", objcStats)
	}
}

func TestExtensionExpansionAmbiguousM(t *testing.T) {
	root := makeTempRoot(t)
	writeFile(t, filepath.Join(root, "script.m"), "function y = f(x)\n")
	writeFile(t, filepath.Join(root, "Thing.m"), "@interface Thing\n")
	writeFile(t, filepath.Join(root, "Thing.mm"), "@implementation Thing\n")
	writeFile(t, filepath.Join(root, "Thing.h"), "@interface Thing\n")

	registry := lang.NewRegistry()
	counter := NewCounter(registry)

	resp, err := counter.Count(context.Background(), model.CountRequest{
		Root:       root,
		Extensions: []string{".m"},
	})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	matlabStats, ok := resp.Languages["matlab"]
	if !ok || matlabStats.LOC != 1 {
		t.Fatalf("expected matlab loc 1, got %+v", matlabStats)
	}
	objcStats, ok := resp.Languages["objective-c"]
	if !ok || objcStats.LOC != 3 {
		t.Fatalf("expected objective-c loc 3, got %+v", objcStats)
	}
}

func TestFullLanguageCoverageRepo(t *testing.T) {
	root := makeTempRoot(t)

	type fixture struct {
		path   string
		langID string
		isTest bool
		data   string
	}
	fixtures := []fixture{
		// Go
		{path: "src/main.go", langID: "go", isTest: false, data: "package main\nfunc main() {}\n"},
		{path: "src/main_test.go", langID: "go", isTest: true, data: "package main\nfunc TestMain(t *T) {}\n"},
		// Java/Kotlin
		{path: "src/main/java/App.java", langID: "java", isTest: false, data: "class App {}\n"},
		{path: "src/test/java/AppTest.java", langID: "java", isTest: true, data: "@Test void testA() {}\n"},
		{path: "src/main/kotlin/App.kt", langID: "kotlin", isTest: false, data: "class App\n"},
		{path: "src/test/kotlin/AppTest.kt", langID: "kotlin", isTest: true, data: "@Test fun testA() {}\n"},
		// JS/TS
		{path: "web/app.js", langID: "javascript", isTest: false, data: "const x = 1;\n"},
		{path: "web/__tests__/app.test.js", langID: "javascript", isTest: true, data: "describe('x', () => { it('a', () => {}) })\n"},
		{path: "web/app.ts", langID: "typescript", isTest: false, data: "export const x = 1;\n"},
		{path: "web/tests/app.spec.ts", langID: "typescript", isTest: true, data: "test('a', () => {})\n"},
		// Python
		{path: "py/app.py", langID: "python", isTest: false, data: "def f():\n    pass\n"},
		{path: "py/tests/test_app.py", langID: "python", isTest: true, data: "def test_a():\n    pass\n"},
		// PHP
		{path: "php/App.php", langID: "php", isTest: false, data: "<?php class App {}\n"},
		{path: "php/tests/AppTest.php", langID: "php", isTest: true, data: "<?php\n/** @test */\npublic function testA() {}\n"},
		// Ruby
		{path: "rb/lib/app.rb", langID: "ruby", isTest: false, data: "class App\nend\n"},
		{path: "rb/spec/app_spec.rb", langID: "ruby", isTest: true, data: "describe 'x' do\n  it 'a' do\n  end\nend\n\ndef test_b\nend\n"},
		// Rust
		{path: "rs/src/lib.rs", langID: "rust", isTest: false, data: "#[test]\nfn test_lib_inner() {}\npub fn lib() {}\n"},
		{path: "rs/tests/lib.rs", langID: "rust", isTest: true, data: "#[test]\nfn test_lib() {}\n"},
		// Swift
		{path: "swift/Sources/App.swift", langID: "swift", isTest: false, data: "struct App {}\n"},
		{path: "swift/Tests/AppTests.swift", langID: "swift", isTest: true, data: "func testA() {}\n"},
		// C/C++
		{path: "c/app.c", langID: "c", isTest: false, data: "int main() { return 0; }\n"},
		{path: "c/tests/app_test.c", langID: "c", isTest: true, data: "int test() { return 0; }\n"},
		{path: "cpp/app.cpp", langID: "cpp", isTest: false, data: "int main() { return 0; }\n"},
		{path: "cpp/tests/app_test.cpp", langID: "cpp", isTest: true, data: "int test() { return 0; }\n"},
		// Objective-C + MATLAB
		{path: "objc/Thing.m", langID: "objective-c", isTest: false, data: "@implementation Thing\n@end\n"},
		{path: "objc/Thing.h", langID: "objective-c", isTest: false, data: "@interface Thing\n@end\n"},
		{path: "objc/Tests/ThingTests.m", langID: "objective-c", isTest: true, data: "- (void)testA {}\n"},
		{path: "matlab/script.m", langID: "matlab", isTest: false, data: "function y = f(x)\n"},
		{path: "matlab/tests/script_test.m", langID: "matlab", isTest: true, data: "function y = f(x)\n"},
		// C#
		{path: "cs/Thing.cs", langID: "csharp", isTest: false, data: "public class Thing {}\n"},
		{path: "cs/tests/ThingTests.cs", langID: "csharp", isTest: true, data: "[Fact]\npublic void TestA() {}\n"},
		// Scala
		{path: "scala/src/main/Thing.scala", langID: "scala", isTest: false, data: "object Thing {}\n"},
		{path: "scala/src/test/ThingSpec.scala", langID: "scala", isTest: true, data: "test(\"a\")\n"},
		// Groovy
		{path: "groovy/src/main/Thing.groovy", langID: "groovy", isTest: false, data: "class Thing {}\n"},
		{path: "groovy/src/test/ThingSpec.groovy", langID: "groovy", isTest: true, data: "@Test\nvoid testA() {}\n"},
		// Perl
		{path: "perl/lib/Thing.pm", langID: "perl", isTest: false, data: "package Thing;\n"},
		{path: "perl/t/thing.t", langID: "perl", isTest: true, data: "ok(1, 'works')\nsubtest(\"x\" => sub { ok(1) })\n"},
		// Dart
		{path: "dart/lib/app.dart", langID: "dart", isTest: false, data: "void main() {}\n"},
		{path: "dart/test/app_test.dart", langID: "dart", isTest: true, data: "group('x', () { test('a', () {}); });\n"},
		// Lua
		{path: "lua/app.lua", langID: "lua", isTest: false, data: "local x = 1\n"},
		{path: "lua/spec/app_spec.lua", langID: "lua", isTest: true, data: "describe('x', function() it('a', function() end) end)\n"},
		// R
		{path: "r/app.R", langID: "r", isTest: false, data: "x <- 1\n"},
		{path: "r/tests/testthat/test-app.R", langID: "r", isTest: true, data: "test_that(\"x\", { expect_equal(1, 1) })\n"},
		// Shell
		{path: "sh/script.sh", langID: "shell", isTest: false, data: "echo hi\n"},
		{path: "sh/tests/test_script.sh", langID: "shell", isTest: true, data: "echo test\n"},
	}

	for _, fx := range fixtures {
		writeFile(t, filepath.Join(root, filepath.FromSlash(fx.path)), fx.data)
	}

	registry := lang.NewRegistry()
	counter := NewCounter(registry)

	resp, err := counter.Count(context.Background(), model.CountRequest{
		Root: root,
	})
	if err != nil {
		t.Fatalf("count: %v", err)
	}

	expectedLangs := []string{
		"go", "java", "kotlin", "javascript", "typescript", "python", "php", "ruby",
		"rust", "swift", "c", "cpp", "objective-c", "matlab", "csharp", "scala",
		"groovy", "perl", "dart", "lua", "r", "shell",
	}
	for _, id := range expectedLangs {
		if _, ok := resp.Languages[id]; !ok {
			t.Fatalf("expected language %s", id)
		}
	}

	expected := make(map[string]model.LanguageStats)
	for _, fx := range fixtures {
		stats := expected[fx.langID]
		lines := countLinesString(fx.data)
		stats.LOC += lines
		if fx.isTest {
			stats.TestLOC += lines
		}
		expected[fx.langID] = stats
	}
	var totalLOC int
	var totalTestLOC int
	for id, stats := range expected {
		stats.CodeLOC = stats.LOC - stats.TestLOC
		if stats.LOC > 0 {
			stats.PercentTestLOC = floatPtr(round2(float64(stats.TestLOC) / float64(stats.LOC) * 100))
			stats.PercentCodeLOC = floatPtr(round2(float64(stats.CodeLOC) / float64(stats.LOC) * 100))
		}
		expected[id] = stats
		totalLOC += stats.LOC
		totalTestLOC += stats.TestLOC
	}
	totalCodeLOC := totalLOC - totalTestLOC
	var expectedPercentTest *float64
	var expectedPercentCode *float64
	if totalLOC > 0 {
		expectedPercentTest = floatPtr(round2(float64(totalTestLOC) / float64(totalLOC) * 100))
		expectedPercentCode = floatPtr(round2(float64(totalCodeLOC) / float64(totalLOC) * 100))
	}

	expectCounts := map[string]int{
		"go":          1,
		"java":        1,
		"kotlin":      1,
		"javascript":  2,
		"typescript":  1,
		"python":      1,
		"php":         2,
		"ruby":        3,
		"rust":        2,
		"swift":       1,
		"objective-c": 1,
		"csharp":      1,
		"scala":       1,
		"groovy":      1,
		"perl":        3,
		"dart":        2,
		"lua":         2,
		"r":           1,
	}
	for id, want := range expectCounts {
		stats := resp.Languages[id]
		if stats.TestCount == nil || *stats.TestCount != want {
			t.Fatalf("expected %s test count %d, got %v", id, want, stats.TestCount)
		}
	}

	for _, id := range expectedLangs {
		stats := resp.Languages[id]
		exp := expected[id]
		if stats.LOC != exp.LOC || stats.TestLOC != exp.TestLOC || stats.CodeLOC != exp.CodeLOC {
			t.Fatalf("unexpected %s loc stats: got %+v want %+v", id, stats, exp)
		}
		if stats.PercentTestLOC == nil || stats.PercentCodeLOC == nil {
			t.Fatalf("expected %s percent locs", id)
		}
		if !floatEqual(*stats.PercentTestLOC, *exp.PercentTestLOC) || !floatEqual(*stats.PercentCodeLOC, *exp.PercentCodeLOC) {
			t.Fatalf("unexpected %s percents: got %v/%v want %v/%v", id, *stats.PercentTestLOC, *stats.PercentCodeLOC, *exp.PercentTestLOC, *exp.PercentCodeLOC)
		}
		if stats.TestLOC == 0 {
			t.Fatalf("expected %s test loc", id)
		}
	}

	if resp.LOC != totalLOC || resp.TestLOC != totalTestLOC || resp.CodeLOC != totalCodeLOC {
		t.Fatalf("unexpected total loc stats: got loc=%d test=%d code=%d want loc=%d test=%d code=%d", resp.LOC, resp.TestLOC, resp.CodeLOC, totalLOC, totalTestLOC, totalCodeLOC)
	}
	if resp.PercentTestLOC == nil || resp.PercentCodeLOC == nil || expectedPercentTest == nil || expectedPercentCode == nil {
		t.Fatalf("expected total percent locs")
	}
	if !floatEqual(*resp.PercentTestLOC, *expectedPercentTest) || !floatEqual(*resp.PercentCodeLOC, *expectedPercentCode) {
		t.Fatalf("unexpected total percents: got %v/%v want %v/%v", *resp.PercentTestLOC, *resp.PercentCodeLOC, *expectedPercentTest, *expectedPercentCode)
	}
}

func countLinesString(data string) int {
	if len(data) == 0 {
		return 0
	}
	lines := strings.Count(data, "\n")
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func floatPtr(value float64) *float64 {
	v := value
	return &v
}

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.0001
}
func TestCountCustomExcludedDirs(t *testing.T) {
	root := makeTempRoot(t)
	writeFile(t, filepath.Join(root, "main.go"), "package main\n")
	writeFile(t, filepath.Join(root, "vendor", "vendored.go"), "package vendor\n")

	registry := lang.NewRegistry()
	counter := NewCounter(registry)

	resp, err := counter.Count(context.Background(), model.CountRequest{
		Root:         root,
		ExcludedDirs: []string{"node_modules"},
	})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if resp.LOC != 2 {
		t.Fatalf("expected 2 total LOC, got %d", resp.LOC)
	}
}

func TestCountContextCanceled(t *testing.T) {
	registry := lang.NewRegistry()
	counter := NewCounter(registry)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := counter.Count(ctx, model.CountRequest{Root: "."})
	if err == nil {
		t.Fatalf("expected context error, got nil")
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
