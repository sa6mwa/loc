package model

// CountRequest describes a counting request.
type CountRequest struct {
	Roots        []string
	Root         string
	Extensions   []string
	ExcludedDirs []string
}

// CountResponse is the JSON response for counts.
type CountResponse struct {
	LOC            int                      `json:"loc"`
	TestLOC        int                      `json:"test_loc"`
	CodeLOC        int                      `json:"code_loc"`
	PercentTestLOC *float64                 `json:"percent_test_loc,omitempty"`
	PercentCodeLOC *float64                 `json:"percent_code_loc,omitempty"`
	Languages      map[string]LanguageStats `json:"languages"`
}

// LanguageStats holds counts for a single language.
type LanguageStats struct {
	LOC            int      `json:"loc"`
	TestLOC        int      `json:"test_loc"`
	CodeLOC        int      `json:"code_loc"`
	PercentTestLOC *float64 `json:"percent_test_loc,omitempty"`
	PercentCodeLOC *float64 `json:"percent_code_loc,omitempty"`
	TestCount      *int     `json:"test_count,omitempty"`
	ExampleCount   *int     `json:"example_count,omitempty"`
	BenchmarkCount *int     `json:"benchmark_count,omitempty"`
	FuzzCount      *int     `json:"fuzz_count,omitempty"`
}

// TestCounts captures test-related counts for a file.
type TestCounts struct {
	Tests      int
	Examples   int
	Benchmarks int
	Fuzz       int
}
