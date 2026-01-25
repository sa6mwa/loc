# loc

`loc` scans the current directory tree and reports lines of code by language, separating test code from non-test code and emitting pretty JSON via `prettyx`.

### Usage

```
loc [flags] [extensions...]
```

By default, `loc` scans the current working directory and counts all supported
languages. Providing extensions limits the scan to those languages.

Examples:

```
loc
loc .go
loc .go .c .cpp
```

### Extension filters

When you pass extensions, `loc` expands them to the full set of extensions for
the matching languages. This avoids partial counts for languages that use
multiple extensions.

Examples:

- `loc .h` counts C and Objective-C headers and also includes their related
  extensions (e.g. `.c`, `.m`, `.mm`).
- `loc .m` includes MATLAB and Objective-C sources (disambiguated by content).

Ambiguous extensions like `.h` and `.m` are resolved by inspecting file content.

### Output

The JSON includes:

- `loc`, `test_loc`, `code_loc`
- `percent_test_loc`, `percent_code_loc` (100 == 100%)
- `languages` map with per-language `loc`, `test_loc`, `code_loc`
- Optional test counters (e.g., Go `test_count`, `example_count`, `benchmark_count`, `fuzz_count`)

### Supported languages

- Go (`.go`)
- Kotlin (`.kt`, `.kts`)
- Java (`.java`)
- JavaScript (`.js`, `.jsx`, `.mjs`, `.cjs`)
- TypeScript (`.ts`, `.tsx`, `.mts`, `.cts`)
- Python (`.py`)
- PHP (`.php`)
- Ruby (`.rb`)
- Rust (`.rs`)
- Swift (`.swift`)
- C (`.c`, `.h`)
- C++ (`.cpp`, `.cc`, `.cxx`, `.hpp`, `.hh`, `.hxx`)
- Objective-C (`.m`, `.mm`, `.h`)
- C# (`.cs`)
- Scala (`.scala`)
- Groovy (`.groovy`)
- Perl (`.pl`, `.pm`, `.t`)
- Dart (`.dart`)
- Lua (`.lua`)
- R (`.r`, `.R`)
- MATLAB (`.m`)
- Shell (`.sh`, `.bash`, `.zsh`)

### Test detection (high-level)

Tests are detected using common, deterministic conventions:

- Go: `_test.go` files; counts `Test`, `Example`, `Benchmark`, `Fuzz` functions.
- Java/Kotlin: `src/test` or `*Test.*`; counts `@Test`.
- JS/TS: `__tests__`, `test/`, `tests/`, `*.test.*`, `*.spec.*`; counts `describe`/`it`/`test`.
- Python: `test_*.py`/`*_test.py` or `tests/`; counts `def test_`.
- PHP: `tests/` or `*Test.php`; counts `@test` and `function test*`.
- Ruby: `test/` or `spec/`; counts `def test_`, `describe`, `it`.
- Rust: `tests/` or `*_test.rs` for test LOC; counts `#[test]` anywhere.
- Swift: `Tests/` or `*Tests.swift`; counts `func test*`.
- C/C++/Shell: test LOC via `test(s)/` or filename patterns; no test counts.
- Objective-C: `Tests/` or `*Test(s).m/.mm`; counts `- (void)test*`.
- C#: test LOC via `test(s)/` or `*Test(s).cs`; counts `[Test]`, `[Fact]`, `[Theory]`.
- Scala: `src/test` or `*Test.scala`; counts `test(`/`it(`/`should(`.
- Groovy: `src/test` or `*Test.groovy`; counts `@Test` or `def "..."`.
- Perl: `t/` or `*.t`; counts `ok(` and `subtest(`.
- Dart: `test/` or `*_test.dart`; counts `test(` and `group(`.
- Lua: `spec/` or `*_spec.lua`; counts `describe`/`it`.
- R: `tests/` or `testthat/`; counts `test_that(`.
- MATLAB: `tests/` or `*_test.m`; no test counts.

### Install

```
go build -o bin/loc -trimpath -ldflags="-s -w" .
sudo install bin/loc /usr/local/bin/

# or...
go install pkt.systems/loc@latest
```
