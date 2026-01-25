package scan

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"pkt.systems/loc/internal/lang"
)

// DefaultExcludedDirs lists directory names to skip by default.
var DefaultExcludedDirs = map[string]bool{
	"vendor":       true,
	"node_modules": true,
	"third_party":  true,
	".git":         true,
	".hg":          true,
	".svn":         true,
	"dist":         true,
	"build":        true,
	"out":          true,
	"bin":          true,
	"target":       true,
}

// FileEntry represents a discovered source file.
type FileEntry struct {
	Path     string
	Language lang.LanguageSpec
	IsTest   bool
}

// Walk finds source files under root for the given registry and extension filter.
func Walk(root string, registry *lang.Registry, extensions []string, excludedDirs map[string]bool) ([]FileEntry, error) {
	allowed := make(map[string]bool)
	if len(extensions) > 0 {
		for _, ext := range extensions {
			allowed[lang.NormalizeExtension(ext)] = true
		}
	}
	if excludedDirs == nil {
		excludedDirs = DefaultExcludedDirs
	}

	entries := make([]FileEntry, 0, 128)
	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := entry.Name()
		if entry.IsDir() {
			if excludedDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ext == "" {
			return nil
		}
		if len(allowed) > 0 && !allowed[lang.NormalizeExtension(ext)] {
			return nil
		}
		languages, ok := registry.ByExtension(ext)
		if !ok {
			return nil
		}
		chosen, err := selectLanguage(path, languages)
		if err != nil {
			return err
		}
		if chosen == nil {
			return nil
		}
		entries = append(entries, FileEntry{
			Path:     path,
			Language: *chosen,
			IsTest:   chosen.IsTestFile != nil && chosen.IsTestFile(path),
		})
		return nil
	})

	return entries, walkErr
}

func selectLanguage(path string, languages []lang.LanguageSpec) (*lang.LanguageSpec, error) {
	if len(languages) == 0 {
		return nil, nil
	}
	if len(languages) == 1 {
		return &languages[0], nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	for _, candidate := range languages {
		if candidate.ContentClassifier != nil && candidate.ContentClassifier(path, content) {
			chosen := candidate
			return &chosen, nil
		}
	}
	chosen := languages[0]
	return &chosen, nil
}
