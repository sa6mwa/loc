package lang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"pkt.systems/loc/internal/model"
)

// TestCountSupport indicates which test counters a language supports.
type TestCountSupport struct {
	Tests      bool
	Examples   bool
	Benchmarks bool
	Fuzz       bool
}

// TestCounter counts test-related constructs in a file.
type TestCounter func(path string, content []byte) model.TestCounts

// LanguageSpec defines language matching and test detection behavior.
type LanguageSpec struct {
	ID                   string
	Name                 string
	Extensions           []string
	IsTestFile           func(path string) bool
	TestCounter          TestCounter
	ContentClassifier    func(path string, content []byte) bool
	CountTestsInAllFiles bool
	TestCountSupport     TestCountSupport
}

// Registry holds language specifications and extension lookups.
type Registry struct {
	languages []LanguageSpec
	byExt     map[string][]LanguageSpec
}

// NewRegistry builds the default language registry.
func NewRegistry() *Registry {
	languages := []LanguageSpec{
		{
			ID:          "go",
			Name:        "Go",
			Extensions:  []string{".go"},
			IsTestFile:  isGoTestFile,
			TestCounter: countGoTests,
			TestCountSupport: TestCountSupport{
				Tests:      true,
				Examples:   true,
				Benchmarks: true,
				Fuzz:       true,
			},
		},
		{
			ID:                "matlab",
			Name:              "MATLAB",
			Extensions:        []string{".m"},
			IsTestFile:        isMatlabTestFile,
			ContentClassifier: isMatlabContent,
		},
		{
			ID:               "kotlin",
			Name:             "Kotlin",
			Extensions:       []string{".kt", ".kts"},
			IsTestFile:       isKotlinTestFile,
			TestCounter:      countJUnitTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:               "java",
			Name:             "Java",
			Extensions:       []string{".java"},
			IsTestFile:       isJavaTestFile,
			TestCounter:      countJUnitTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:               "javascript",
			Name:             "JavaScript",
			Extensions:       []string{".js", ".jsx", ".mjs", ".cjs"},
			IsTestFile:       isJSTestFile,
			TestCounter:      countJSTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:               "typescript",
			Name:             "TypeScript",
			Extensions:       []string{".ts", ".tsx", ".mts", ".cts"},
			IsTestFile:       isTSTestFile,
			TestCounter:      countJSTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:               "python",
			Name:             "Python",
			Extensions:       []string{".py"},
			IsTestFile:       isPythonTestFile,
			TestCounter:      countPythonTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:               "php",
			Name:             "PHP",
			Extensions:       []string{".php"},
			IsTestFile:       isPHPTestFile,
			TestCounter:      countPHPTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:               "ruby",
			Name:             "Ruby",
			Extensions:       []string{".rb"},
			IsTestFile:       isRubyTestFile,
			TestCounter:      countRubyTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:                   "rust",
			Name:                 "Rust",
			Extensions:           []string{".rs"},
			IsTestFile:           isRustTestFile,
			TestCounter:          countRustTests,
			CountTestsInAllFiles: true,
			TestCountSupport:     TestCountSupport{Tests: true},
		},
		{
			ID:               "swift",
			Name:             "Swift",
			Extensions:       []string{".swift"},
			IsTestFile:       isSwiftTestFile,
			TestCounter:      countSwiftTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:         "shell",
			Name:       "Shell",
			Extensions: []string{".sh", ".bash", ".zsh"},
			IsTestFile: isShellTestFile,
		},
		{
			ID:         "c",
			Name:       "C",
			Extensions: []string{".c", ".h"},
			IsTestFile: isCTestFile,
		},
		{
			ID:         "cpp",
			Name:       "C++",
			Extensions: []string{".cpp", ".cc", ".cxx", ".hpp", ".hh", ".hxx"},
			IsTestFile: isCPPTestFile,
		},
		{
			ID:                "objective-c",
			Name:              "Objective-C",
			Extensions:        []string{".m", ".mm", ".h"},
			IsTestFile:        isObjectiveCTestFile,
			TestCounter:       countObjectiveCTests,
			ContentClassifier: isObjectiveCContent,
			TestCountSupport:  TestCountSupport{Tests: true},
		},
		{
			ID:               "csharp",
			Name:             "C#",
			Extensions:       []string{".cs"},
			IsTestFile:       isCSharpTestFile,
			TestCounter:      countCSharpTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:               "scala",
			Name:             "Scala",
			Extensions:       []string{".scala"},
			IsTestFile:       isScalaTestFile,
			TestCounter:      countScalaTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:               "groovy",
			Name:             "Groovy",
			Extensions:       []string{".groovy"},
			IsTestFile:       isGroovyTestFile,
			TestCounter:      countGroovyTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:               "perl",
			Name:             "Perl",
			Extensions:       []string{".pl", ".pm", ".t"},
			IsTestFile:       isPerlTestFile,
			TestCounter:      countPerlTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:               "dart",
			Name:             "Dart",
			Extensions:       []string{".dart"},
			IsTestFile:       isDartTestFile,
			TestCounter:      countDartTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:               "lua",
			Name:             "Lua",
			Extensions:       []string{".lua"},
			IsTestFile:       isLuaTestFile,
			TestCounter:      countLuaTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
		{
			ID:               "r",
			Name:             "R",
			Extensions:       []string{".r", ".R"},
			IsTestFile:       isRTestFile,
			TestCounter:      countRTests,
			TestCountSupport: TestCountSupport{Tests: true},
		},
	}

	byExt := make(map[string][]LanguageSpec)
	for _, lang := range languages {
		for _, ext := range lang.Extensions {
			key := normalizeExt(ext)
			byExt[key] = append(byExt[key], lang)
		}
	}

	return &Registry{
		languages: languages,
		byExt:     byExt,
	}
}

// Languages returns the configured language specs.
func (r *Registry) Languages() []LanguageSpec {
	langs := make([]LanguageSpec, len(r.languages))
	copy(langs, r.languages)
	return langs
}

// ByExtension resolves languages by extension.
func (r *Registry) ByExtension(ext string) ([]LanguageSpec, bool) {
	langs, ok := r.byExt[normalizeExt(ext)]
	if !ok {
		return nil, false
	}
	out := make([]LanguageSpec, len(langs))
	copy(out, langs)
	return out, true
}

// HasExtension reports whether the extension is supported.
func (r *Registry) HasExtension(ext string) bool {
	_, ok := r.byExt[normalizeExt(ext)]
	return ok
}

// NormalizeExtension ensures an extension starts with a dot and lowercases it.
func NormalizeExtension(ext string) string {
	return normalizeExt(ext)
}

func normalizeExt(ext string) string {
	trimmed := strings.TrimSpace(ext)
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, ".") {
		trimmed = "." + trimmed
	}
	return strings.ToLower(trimmed)
}

func hasTestDir(path string) bool {
	slash := filepath.ToSlash(path)
	parts := strings.Split(slash, "/")
	for _, part := range parts[:len(parts)-1] {
		switch strings.ToLower(part) {
		case "test", "tests", "__tests__":
			return true
		}
	}
	return false
}

func hasPathFragment(path string, fragment string) bool {
	slash := filepath.ToSlash(path)
	return strings.Contains(slash, fragment)
}

func hasSuffixFold(name string, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(name), strings.ToLower(suffix))
}

func isGoTestFile(path string) bool {
	return hasSuffixFold(path, "_test.go")
}

func isJavaTestFile(path string) bool {
	base := filepath.Base(path)
	if hasPathFragment(path, "/src/test/") {
		return true
	}
	return hasSuffixFold(base, "Test.java") || hasSuffixFold(base, "Tests.java")
}

func isKotlinTestFile(path string) bool {
	base := filepath.Base(path)
	if hasPathFragment(path, "/src/test/") {
		return true
	}
	return hasSuffixFold(base, "Test.kt") || hasSuffixFold(base, "Tests.kt") || hasSuffixFold(base, "Test.kts") || hasSuffixFold(base, "Tests.kts")
}

func isJSTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) {
		return true
	}
	return strings.HasSuffix(base, ".test.js") || strings.HasSuffix(base, ".spec.js") || strings.HasSuffix(base, ".test.jsx") || strings.HasSuffix(base, ".spec.jsx") || strings.HasSuffix(base, ".test.mjs") || strings.HasSuffix(base, ".spec.mjs") || strings.HasSuffix(base, ".test.cjs") || strings.HasSuffix(base, ".spec.cjs")
}

func isTSTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) {
		return true
	}
	return strings.HasSuffix(base, ".test.ts") || strings.HasSuffix(base, ".spec.ts") || strings.HasSuffix(base, ".test.tsx") || strings.HasSuffix(base, ".spec.tsx") || strings.HasSuffix(base, ".test.mts") || strings.HasSuffix(base, ".spec.mts") || strings.HasSuffix(base, ".test.cts") || strings.HasSuffix(base, ".spec.cts")
}

func isPythonTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) {
		return true
	}
	return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py")
}

func isPHPTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) || hasPathFragment(path, "/tests/") || hasPathFragment(path, "/test/") {
		return true
	}
	return strings.HasSuffix(base, "test.php") || strings.HasSuffix(base, "tests.php")
}

func isRubyTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) || hasPathFragment(path, "/spec/") {
		return true
	}
	return strings.HasSuffix(base, "_test.rb") || strings.HasSuffix(base, "_spec.rb")
}

func isRustTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) {
		return true
	}
	return strings.HasSuffix(base, "_test.rs")
}

func isSwiftTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasPathFragment(path, "/tests/") || hasPathFragment(path, "/test/") {
		return true
	}
	return strings.HasSuffix(base, "tests.swift") || strings.HasSuffix(base, "test.swift")
}

func isShellTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) {
		return true
	}
	return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.sh")
}

func isCTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) {
		return true
	}
	return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.c") || strings.HasSuffix(base, "_test.h")
}

func isCPPTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) {
		return true
	}
	return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.cpp") || strings.HasSuffix(base, "_test.cc") || strings.HasSuffix(base, "_test.cxx") || strings.HasSuffix(base, "_test.hpp") || strings.HasSuffix(base, "_test.hh") || strings.HasSuffix(base, "_test.hxx")
}

func isCSharpTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) {
		return true
	}
	return strings.HasSuffix(base, "test.cs") || strings.HasSuffix(base, "tests.cs") || strings.HasSuffix(base, "spec.cs")
}

func isObjectiveCTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasPathFragment(path, "/tests/") || hasPathFragment(path, "/test/") {
		return true
	}
	return strings.HasSuffix(base, "tests.m") || strings.HasSuffix(base, "test.m") || strings.HasSuffix(base, "tests.mm") || strings.HasSuffix(base, "test.mm")
}

func isScalaTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasPathFragment(path, "/src/test/") || hasPathFragment(path, "/test/") || hasPathFragment(path, "/tests/") {
		return true
	}
	return strings.HasSuffix(base, "test.scala") || strings.HasSuffix(base, "spec.scala")
}

func isGroovyTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasPathFragment(path, "/src/test/") || hasPathFragment(path, "/test/") || hasPathFragment(path, "/tests/") {
		return true
	}
	return strings.HasSuffix(base, "test.groovy") || strings.HasSuffix(base, "spec.groovy")
}

func isPerlTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) || hasPathFragment(path, "/t/") {
		return true
	}
	return strings.HasSuffix(base, ".t") || strings.HasSuffix(base, "test.pm") || strings.HasSuffix(base, "tests.pm")
}

func isDartTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) || hasPathFragment(path, "/test/") || hasPathFragment(path, "/tests/") {
		return true
	}
	return strings.HasSuffix(base, "_test.dart")
}

func isLuaTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) || hasPathFragment(path, "/spec/") {
		return true
	}
	return strings.HasSuffix(base, "_test.lua") || strings.HasSuffix(base, "_spec.lua")
}

func isRTestFile(path string) bool {
	base := filepath.Base(path)
	if hasTestDir(path) || hasPathFragment(path, "/testthat/") {
		return true
	}
	return strings.HasSuffix(base, "_test.R") || strings.HasSuffix(base, "_test.r")
}

func isMatlabTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if hasTestDir(path) || hasPathFragment(path, "/test/") || hasPathFragment(path, "/tests/") {
		return true
	}
	return strings.HasSuffix(base, "_test.m")
}

func countGoTests(path string, content []byte) model.TestCounts {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, content, parser.SkipObjectResolution)
	if err != nil {
		return model.TestCounts{}
	}
	var counts model.TestCounts
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil {
			continue
		}
		name := fn.Name.Name
		switch {
		case isGoTestName(name, "Test"):
			counts.Tests++
		case isGoTestName(name, "Example"):
			counts.Examples++
		case isGoTestName(name, "Benchmark"):
			counts.Benchmarks++
		case isGoTestName(name, "Fuzz"):
			counts.Fuzz++
		}
	}
	return counts
}

func isGoTestName(name string, prefix string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	if len(name) == len(prefix) {
		return true
	}
	r, _ := utf8.DecodeRuneInString(name[len(prefix):])
	return !unicode.IsLower(r)
}

var (
	junitTestRe  = regexp.MustCompile(`@\s*(?:Test|ParameterizedTest|RepeatedTest|TestFactory|TestTemplate)\b`)
	jsTestRe     = regexp.MustCompile(`\b(?:describe|it|test)\s*(?:\.\s*(?:only|skip|each|concurrent))*\s*\(`)
	pythonTestRe = regexp.MustCompile(`(?m)^\s*def\s+test_[A-Za-z0-9_]*\s*\(`)
	rustTestRe   = regexp.MustCompile(`(?m)^\s*#\s*\[\s*(?:test|tokio::test|async_std::test|rstest)\b`)
	csharpTestRe = regexp.MustCompile(`(?m)^\s*\[\s*(?:Test|Fact|Theory|TestCase|TestCaseSource|TestMethod)\b`)
	phpTestRe    = regexp.MustCompile(`(?m)(?:@\s*test\b|^\s*(?:public|protected|private)?\s*function\s+test[A-Za-z0-9_]*\s*\()`)
	rubyTestRe   = regexp.MustCompile(`(?m)(?:^\s*def\s+test_[A-Za-z0-9_]*\b|\b(?:describe|it)\s*(?:\(|['\"]))`)
	swiftTestRe  = regexp.MustCompile(`(?m)^\s*func\s+test[A-Za-z0-9_]*\s*\(`)
	objcTestRe   = regexp.MustCompile(`(?m)^\s*[-+]\s*\(\s*void\s*\)\s*test[A-Za-z0-9_]*`)
	scalaTestRe  = regexp.MustCompile(`\b(?:test|it|should)\s*\(`)
	groovyTestRe = regexp.MustCompile(`(?m)(?:@\s*Test\b|^\s*def\s+\".+\"\s*\()`)
	perlTestRe   = regexp.MustCompile(`\b(?:ok|subtest)\s*\(`)
	dartTestRe   = regexp.MustCompile(`\b(?:test|group)\s*\(`)
	luaTestRe    = regexp.MustCompile(`\b(?:describe|it)\s*\(`)
	rTestRe      = regexp.MustCompile(`\btest_that\s*\(`)
	matlabRe     = regexp.MustCompile(`(?m)^\s*(?:function|classdef)\b`)
	objcRe       = regexp.MustCompile(`(?m)^\s*(?:@interface|@implementation|@protocol|@property)\b|#\s*import\b`)
)

func countJUnitTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(junitTestRe, content)}
}

func countJSTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(jsTestRe, content)}
}

func countPythonTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(pythonTestRe, content)}
}

func countRustTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(rustTestRe, content)}
}

func countCSharpTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(csharpTestRe, content)}
}

func countPHPTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(phpTestRe, content)}
}

func countRubyTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(rubyTestRe, content)}
}

func countSwiftTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(swiftTestRe, content)}
}

func countObjectiveCTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(objcTestRe, content)}
}

func countScalaTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(scalaTestRe, content)}
}

func countGroovyTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(groovyTestRe, content)}
}

func countPerlTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(perlTestRe, content)}
}

func countDartTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(dartTestRe, content)}
}

func countLuaTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(luaTestRe, content)}
}

func countRTests(_ string, content []byte) model.TestCounts {
	return model.TestCounts{Tests: countRegex(rTestRe, content)}
}

func isMatlabContent(_ string, content []byte) bool {
	return matlabRe.Match(content)
}

func isObjectiveCContent(_ string, content []byte) bool {
	return objcRe.Match(content) || objcTestRe.Match(content)
}

func countRegex(re *regexp.Regexp, content []byte) int {
	return len(re.FindAllIndex(content, -1))
}
