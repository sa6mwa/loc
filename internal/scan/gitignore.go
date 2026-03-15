package scan

import (
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v6/osfs"
	"github.com/go-git/go-git/v6/plumbing/format/gitignore"
)

func newGitIgnoreMatcher(root string) (gitignore.Matcher, error) {
	patterns, err := gitignore.ReadPatterns(osfs.New(root), nil)
	if err != nil {
		return nil, err
	}
	if len(patterns) == 0 {
		return nil, nil
	}
	return gitignore.NewMatcher(patterns), nil
}

func matchesGitIgnore(root string, path string, isDir bool, matcher gitignore.Matcher) (bool, error) {
	if matcher == nil {
		return false, nil
	}
	parts, err := gitIgnoreParts(root, path)
	if err != nil {
		return false, err
	}
	if len(parts) == 0 {
		return false, nil
	}
	return matcher.Match(parts, isDir), nil
}

func gitIgnoreParts(root string, path string) ([]string, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return nil, err
	}
	if rel == "." {
		return nil, nil
	}
	return strings.Split(filepath.ToSlash(rel), "/"), nil
}
