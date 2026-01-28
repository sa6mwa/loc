package main

import (
	"context"
	"flag"
	"os"
	"sort"
	"strings"

	"pkt.systems/loc/internal/engine"
	"pkt.systems/loc/internal/lang"
	"pkt.systems/loc/internal/model"
	"pkt.systems/loc/internal/report"
)

func main() {
	registry := lang.NewRegistry()

	showHelp := flag.Bool("h", false, "show help")
	flag.BoolVar(showHelp, "help", false, "show help")
	var dirs stringList
	flag.Var(&dirs, "d", "directory to scan (repeatable)")
	flag.Parse()

	if *showHelp {
		writeHelp(registry)
		return
	}

	req := model.CountRequest{
		Roots:      []string(dirs),
		Root:       ".",
		Extensions: flag.Args(),
	}

	counter := engine.NewCounter(registry)
	response, err := counter.Count(context.Background(), req)
	if err != nil {
		writeError(err)
		os.Exit(1)
	}

	if err := report.WriteJSON(os.Stdout, response); err != nil {
		writeError(err)
		os.Exit(1)
	}
}

func writeHelp(registry *lang.Registry) {
	extensions := supportedExtensions(registry)
	lines := wrapWords("  Extensions:", extensions, 80)
	output := strings.Join([]string{
		"loc - source code line counter",
		"",
		"Usage:",
		"  loc [flags] [extensions...]",
		"",
		"Flags:",
		"  -d DIR  Scan directory (repeatable). Defaults to current directory.",
		"",
		"Description:",
		"  Count source code lines by language, including test code separation.",
		"",
		"Extension filters:",
		"  Passing extensions expands to all extensions for those languages.",
		"  Example: loc .h includes C and Objective-C sources.",
		"  Example: loc .m includes MATLAB and Objective-C sources.",
		"",
		"Examples:",
		"  loc",
		"  loc .go",
		"  loc .go .c .cpp",
		"  loc -d proj1 -d proj2",
		"",
		"Supported extensions:",
		lines,
		"",
	}, "\n")
	_, _ = os.Stdout.WriteString(output)
}

func supportedExtensions(registry *lang.Registry) []string {
	seen := make(map[string]bool)
	var exts []string
	for _, language := range registry.Languages() {
		for _, ext := range language.Extensions {
			norm := lang.NormalizeExtension(ext)
			if seen[norm] {
				continue
			}
			seen[norm] = true
			exts = append(exts, norm)
		}
	}
	sort.Strings(exts)
	return exts
}

func writeError(err error) {
	detail := strings.TrimSpace(err.Error())
	payload := report.ErrorResponse{
		Error: report.ErrorDetail{
			Message: "loc failed",
			Detail:  detail,
		},
	}
	_ = report.WriteJSON(os.Stderr, payload)
}

type stringList []string

func (s *stringList) String() string {
	if s == nil {
		return ""
	}
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func wrapWords(prefix string, words []string, width int) string {
	if width <= len(prefix)+1 || len(words) == 0 {
		return prefix + " " + strings.Join(words, " ")
	}
	lines := make([]string, 0, 4)
	current := prefix
	lineLen := len(prefix)
	for _, word := range words {
		needed := len(word) + 1
		if lineLen+needed > width {
			lines = append(lines, current)
			current = strings.Repeat(" ", len(prefix)) + word
			lineLen = len(current)
			continue
		}
		current += " " + word
		lineLen += needed
	}
	lines = append(lines, current)
	return strings.Join(lines, "\n")
}
