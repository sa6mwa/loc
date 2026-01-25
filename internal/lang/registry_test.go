package lang

import (
	"testing"
)

func TestNormalizeExtension(t *testing.T) {
	cases := map[string]string{
		"go":    ".go",
		".GO":   ".go",
		" .ts ": ".ts",
		"":      "",
	}
	for input, want := range cases {
		if got := NormalizeExtension(input); got != want {
			t.Fatalf("NormalizeExtension(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestIsTestFileRules(t *testing.T) {
	registry := NewRegistry()
	cases := []struct {
		ext      string
		langID   string
		path     string
		wantTest bool
	}{
		{ext: ".go", langID: "go", path: "foo_test.go", wantTest: true},
		{ext: ".go", langID: "go", path: "foo.go", wantTest: false},
		{ext: ".java", langID: "java", path: "src/test/FooTest.java", wantTest: true},
		{ext: ".java", langID: "java", path: "FooTests.java", wantTest: true},
		{ext: ".java", langID: "java", path: "Foo.java", wantTest: false},
		{ext: ".kt", langID: "kotlin", path: "src/test/FooTest.kt", wantTest: true},
		{ext: ".kt", langID: "kotlin", path: "FooTests.kts", wantTest: true},
		{ext: ".kt", langID: "kotlin", path: "Foo.kt", wantTest: false},
		{ext: ".js", langID: "javascript", path: "__tests__/thing.test.js", wantTest: true},
		{ext: ".js", langID: "javascript", path: "tests/thing.spec.mjs", wantTest: true},
		{ext: ".js", langID: "javascript", path: "src/thing.js", wantTest: false},
		{ext: ".ts", langID: "typescript", path: "tests/thing.spec.ts", wantTest: true},
		{ext: ".ts", langID: "typescript", path: "src/__tests__/thing.test.tsx", wantTest: true},
		{ext: ".ts", langID: "typescript", path: "src/thing.ts", wantTest: false},
		{ext: ".py", langID: "python", path: "tests/test_thing.py", wantTest: true},
		{ext: ".py", langID: "python", path: "thing_test.py", wantTest: true},
		{ext: ".py", langID: "python", path: "thing.py", wantTest: false},
		{ext: ".php", langID: "php", path: "tests/ThingTest.php", wantTest: true},
		{ext: ".php", langID: "php", path: "Thing.php", wantTest: false},
		{ext: ".rb", langID: "ruby", path: "spec/thing_spec.rb", wantTest: true},
		{ext: ".rb", langID: "ruby", path: "test/test_thing.rb", wantTest: true},
		{ext: ".rb", langID: "ruby", path: "lib/thing.rb", wantTest: false},
		{ext: ".rs", langID: "rust", path: "tests/thing.rs", wantTest: true},
		{ext: ".rs", langID: "rust", path: "thing_test.rs", wantTest: true},
		{ext: ".rs", langID: "rust", path: "src/lib.rs", wantTest: false},
		{ext: ".swift", langID: "swift", path: "Tests/ThingTests.swift", wantTest: true},
		{ext: ".swift", langID: "swift", path: "Sources/Thing.swift", wantTest: false},
		{ext: ".m", langID: "objective-c", path: "Tests/ThingTests.m", wantTest: true},
		{ext: ".m", langID: "matlab", path: "tests/thing_test.m", wantTest: true},
		{ext: ".cs", langID: "csharp", path: "tests/ThingTests.cs", wantTest: true},
		{ext: ".cs", langID: "csharp", path: "ThingTest.cs", wantTest: true},
		{ext: ".cs", langID: "csharp", path: "Thing.cs", wantTest: false},
		{ext: ".scala", langID: "scala", path: "src/test/ThingSpec.scala", wantTest: true},
		{ext: ".scala", langID: "scala", path: "src/main/Thing.scala", wantTest: false},
		{ext: ".groovy", langID: "groovy", path: "src/test/ThingSpec.groovy", wantTest: true},
		{ext: ".groovy", langID: "groovy", path: "src/main/Thing.groovy", wantTest: false},
		{ext: ".pl", langID: "perl", path: "t/thing.t", wantTest: true},
		{ext: ".pm", langID: "perl", path: "lib/Thing.pm", wantTest: false},
		{ext: ".dart", langID: "dart", path: "test/thing_test.dart", wantTest: true},
		{ext: ".dart", langID: "dart", path: "lib/thing.dart", wantTest: false},
		{ext: ".lua", langID: "lua", path: "spec/thing_spec.lua", wantTest: true},
		{ext: ".lua", langID: "lua", path: "thing.lua", wantTest: false},
		{ext: ".R", langID: "r", path: "tests/testthat/test-thing.R", wantTest: true},
		{ext: ".R", langID: "r", path: "thing.R", wantTest: false},
		{ext: ".c", langID: "c", path: "tests/thing_test.c", wantTest: true},
		{ext: ".c", langID: "c", path: "test_thing.c", wantTest: true},
		{ext: ".c", langID: "c", path: "thing.c", wantTest: false},
		{ext: ".cpp", langID: "cpp", path: "tests/thing_test.cpp", wantTest: true},
		{ext: ".cpp", langID: "cpp", path: "test_thing.cc", wantTest: true},
		{ext: ".cpp", langID: "cpp", path: "thing.cpp", wantTest: false},
		{ext: ".sh", langID: "shell", path: "tests/test_thing.sh", wantTest: true},
		{ext: ".sh", langID: "shell", path: "test_thing.sh", wantTest: true},
		{ext: ".sh", langID: "shell", path: "script.sh", wantTest: false},
	}

	for _, tc := range cases {
		langSpec := findLanguageSpec(t, registry, tc.ext, tc.langID)
		if langSpec.IsTestFile == nil {
			t.Fatalf("missing IsTestFile for %s", tc.ext)
		}
		if got := langSpec.IsTestFile(tc.path); got != tc.wantTest {
			t.Fatalf("IsTestFile(%s, %s) = %v, want %v", tc.ext, tc.path, got, tc.wantTest)
		}
	}
}

func TestTestCounters(t *testing.T) {
	registry := NewRegistry()
	cases := []struct {
		ext    string
		langID string
		path   string
		data   string
		want   int
		field  string
	}{
		{
			ext:    ".go",
			langID: "go",
			path:   "thing_test.go",
			data:   "package thing\nfunc TestAlpha(t *T) {}\nfunc ExampleThing() {}\nfunc BenchmarkX(b *B) {}\nfunc FuzzY(f *F) {}",
			want:   1,
			field:  "tests",
		},
		{
			ext:    ".java",
			langID: "java",
			path:   "src/test/FooTest.java",
			data:   "@Test public void testA() {}\n@ParameterizedTest void testB() {}\n@TestFactory void testC() {}\n@RepeatedTest(2) void testD() {}\n@TestTemplate void testE() {}",
			want:   5,
			field:  "tests",
		},
		{
			ext:    ".kt",
			langID: "kotlin",
			path:   "src/test/FooTest.kt",
			data:   "@Test fun testA() {}\n@ParameterizedTest fun testB() {}\n@TestFactory fun testC() {}\n@RepeatedTest(2) fun testD() {}\n@TestTemplate fun testE() {}",
			want:   5,
			field:  "tests",
		},
		{
			ext:    ".js",
			langID: "javascript",
			path:   "__tests__/thing.test.js",
			data:   "describe.only('x', () => { it.each([1,2])('a', () => {}) })\n test.concurrent('b', () => {})\n it.skip('c', () => {})",
			want:   4,
			field:  "tests",
		},
		{
			ext:    ".ts",
			langID: "typescript",
			path:   "tests/thing.spec.ts",
			data:   "describe.each([1,2])('x', () => { it.only('a', () => {}) })\n test('b', () => {})\n it.concurrent('c', () => {})",
			want:   4,
			field:  "tests",
		},
		{
			ext:    ".py",
			langID: "python",
			path:   "tests/test_thing.py",
			data:   "def test_a():\n    pass\n\nclass TestThing:\n    def test_b(self):\n        pass\n",
			want:   2,
			field:  "tests",
		},
		{
			ext:    ".php",
			langID: "php",
			path:   "tests/ThingTest.php",
			data:   "/** @test */\npublic function testA() {}\npublic function testB() {}\n",
			want:   3,
			field:  "tests",
		},
		{
			ext:    ".rb",
			langID: "ruby",
			path:   "spec/thing_spec.rb",
			data:   "describe 'x' do\n  it 'a' do\n  end\nend\n\ndef test_b\nend\n",
			want:   3,
			field:  "tests",
		},
		{
			ext:    ".rs",
			langID: "rust",
			path:   "tests/thing.rs",
			data:   "#[test]\nfn test_a() {}\n\n#[tokio::test]\nasync fn test_b() {}\n\n#[async_std::test]\nasync fn test_c() {}\n\n#[rstest]\nfn test_d() {}",
			want:   4,
			field:  "tests",
		},
		{
			ext:    ".swift",
			langID: "swift",
			path:   "Tests/ThingTests.swift",
			data:   "func testA() {}\nfunc testB() {}\n",
			want:   2,
			field:  "tests",
		},
		{
			ext:    ".m",
			langID: "objective-c",
			path:   "Tests/ThingTests.m",
			data:   "- (void)testA {}\n- (void)testB {}\n",
			want:   2,
			field:  "tests",
		},
		{
			ext:    ".cs",
			langID: "csharp",
			path:   "tests/ThingTests.cs",
			data:   "[Test]\npublic void TestA() {}\n[Fact]\npublic void TestB() {}\n[Theory]\npublic void TestC() {}\n[TestMethod]\npublic void TestD() {}\n[TestCase]\npublic void TestE() {}\n[TestCaseSource]\npublic void TestF() {}",
			want:   6,
			field:  "tests",
		},
		{
			ext:    ".scala",
			langID: "scala",
			path:   "src/test/ThingSpec.scala",
			data:   "test(\"a\")\nit(\"b\")\nshould(\"c\")\n",
			want:   3,
			field:  "tests",
		},
		{
			ext:    ".groovy",
			langID: "groovy",
			path:   "src/test/ThingSpec.groovy",
			data:   "@Test\nvoid testA() {}\n\ndef \"should work\"() {}\n",
			want:   2,
			field:  "tests",
		},
		{
			ext:    ".pl",
			langID: "perl",
			path:   "t/thing.t",
			data:   "ok(1, 'works')\nsubtest(\"x\" => sub { ok(1) })\n",
			want:   3,
			field:  "tests",
		},
		{
			ext:    ".dart",
			langID: "dart",
			path:   "test/thing_test.dart",
			data:   "group('x', () { test('a', () {}); })\n",
			want:   2,
			field:  "tests",
		},
		{
			ext:    ".lua",
			langID: "lua",
			path:   "spec/thing_spec.lua",
			data:   "describe('x', function() it('a', function() end) end)\n",
			want:   2,
			field:  "tests",
		},
		{
			ext:    ".R",
			langID: "r",
			path:   "tests/testthat/test-thing.R",
			data:   "test_that(\"x\", { expect_equal(1, 1) })\n",
			want:   1,
			field:  "tests",
		},
	}

	for _, tc := range cases {
		langSpec := findLanguageSpec(t, registry, tc.ext, tc.langID)
		if langSpec.TestCounter == nil {
			t.Fatalf("missing TestCounter for %s", tc.ext)
		}
		counts := langSpec.TestCounter(tc.path, []byte(tc.data))
		switch tc.field {
		case "tests":
			if counts.Tests != tc.want {
				t.Fatalf("TestCounter(%s) tests=%d, want %d", tc.ext, counts.Tests, tc.want)
			}
		default:
			t.Fatalf("unknown field %s", tc.field)
		}
		if tc.ext == ".go" {
			if counts.Examples != 1 || counts.Benchmarks != 1 || counts.Fuzz != 1 {
				t.Fatalf("Go test counts mismatch: %+v", counts)
			}
		}
	}
}

func findLanguageSpec(t *testing.T, registry *Registry, ext string, id string) LanguageSpec {
	t.Helper()
	specs, ok := registry.ByExtension(ext)
	if !ok {
		t.Fatalf("missing language for %s", ext)
	}
	if id == "" {
		if len(specs) != 1 {
			t.Fatalf("expected single language for %s, got %d", ext, len(specs))
		}
		return specs[0]
	}
	for _, spec := range specs {
		if spec.ID == id {
			return spec
		}
	}
	t.Fatalf("missing language id %s for %s", id, ext)
	return LanguageSpec{}
}
