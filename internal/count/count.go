package count

import (
	"bytes"
	"math"
	"os"

	"pkt.systems/loc/internal/lang"
	"pkt.systems/loc/internal/model"
	"pkt.systems/loc/internal/scan"
)

// Aggregate counts lines and tests for the provided file entries.
func Aggregate(entries []scan.FileEntry) (model.CountResponse, error) {
	response := model.CountResponse{
		Languages: make(map[string]model.LanguageStats),
	}

	for _, entry := range entries {
		data, err := os.ReadFile(entry.Path)
		if err != nil {
			return model.CountResponse{}, err
		}
		lines := countLines(data)

		stats := response.Languages[entry.Language.ID]
		initializeTestCounters(&stats, entry.Language.TestCountSupport)
		stats.LOC += lines
		if entry.IsTest {
			stats.TestLOC += lines
		}
		if entry.Language.TestCounter != nil && (entry.Language.CountTestsInAllFiles || entry.IsTest) {
			testCounts := entry.Language.TestCounter(entry.Path, data)
			applyTestCounts(&stats, entry.Language.TestCountSupport, testCounts)
		}
		response.Languages[entry.Language.ID] = stats
	}

	for langID, stats := range response.Languages {
		stats.CodeLOC = stats.LOC - stats.TestLOC
		if stats.LOC > 0 {
			stats.PercentTestLOC = floatPtr(round2(float64(stats.TestLOC) / float64(stats.LOC) * 100))
			stats.PercentCodeLOC = floatPtr(round2(float64(stats.CodeLOC) / float64(stats.LOC) * 100))
		}
		response.LOC += stats.LOC
		response.TestLOC += stats.TestLOC
		response.Languages[langID] = stats
	}
	response.CodeLOC = response.LOC - response.TestLOC
	if response.LOC > 0 {
		response.PercentTestLOC = floatPtr(round2(float64(response.TestLOC) / float64(response.LOC) * 100))
		response.PercentCodeLOC = floatPtr(round2(float64(response.CodeLOC) / float64(response.LOC) * 100))
	}

	return response, nil
}

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func initializeTestCounters(stats *model.LanguageStats, support lang.TestCountSupport) {
	if support.Tests && stats.TestCount == nil {
		stats.TestCount = intPtr(0)
	}
	if support.Examples && stats.ExampleCount == nil {
		stats.ExampleCount = intPtr(0)
	}
	if support.Benchmarks && stats.BenchmarkCount == nil {
		stats.BenchmarkCount = intPtr(0)
	}
	if support.Fuzz && stats.FuzzCount == nil {
		stats.FuzzCount = intPtr(0)
	}
}

func applyTestCounts(stats *model.LanguageStats, support lang.TestCountSupport, counts model.TestCounts) {
	if support.Tests {
		if stats.TestCount == nil {
			stats.TestCount = intPtr(0)
		}
		*stats.TestCount += counts.Tests
	}
	if support.Examples {
		if stats.ExampleCount == nil {
			stats.ExampleCount = intPtr(0)
		}
		*stats.ExampleCount += counts.Examples
	}
	if support.Benchmarks {
		if stats.BenchmarkCount == nil {
			stats.BenchmarkCount = intPtr(0)
		}
		*stats.BenchmarkCount += counts.Benchmarks
	}
	if support.Fuzz {
		if stats.FuzzCount == nil {
			stats.FuzzCount = intPtr(0)
		}
		*stats.FuzzCount += counts.Fuzz
	}
}

func intPtr(value int) *int {
	v := value
	return &v
}

func floatPtr(value float64) *float64 {
	v := value
	return &v
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
