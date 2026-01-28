package engine

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"pkt.systems/loc/internal/count"
	"pkt.systems/loc/internal/lang"
	"pkt.systems/loc/internal/model"
	"pkt.systems/loc/internal/scan"
)

// Counter provides line counting capabilities.
type Counter interface {
	Count(ctx context.Context, req model.CountRequest) (model.CountResponse, error)
}

type counter struct {
	registry *lang.Registry
}

// NewCounter builds a Counter with the provided registry.
func NewCounter(registry *lang.Registry) Counter {
	return &counter{registry: registry}
}

// Count executes a count request.
func (c *counter) Count(ctx context.Context, req model.CountRequest) (model.CountResponse, error) {
	if err := ctx.Err(); err != nil {
		return model.CountResponse{}, err
	}
	roots := normalizeRoots(req)

	normalizedExts, err := c.normalizeExtensions(req.Extensions)
	if err != nil {
		return model.CountResponse{}, err
	}

	excluded := scan.DefaultExcludedDirs
	if len(req.ExcludedDirs) > 0 {
		excluded = make(map[string]bool, len(req.ExcludedDirs))
		for _, dir := range req.ExcludedDirs {
			excluded[dir] = true
		}
	}

	var entries []scan.FileEntry
	for _, root := range roots {
		rootEntries, err := scan.Walk(root, c.registry, normalizedExts, excluded)
		if err != nil {
			return model.CountResponse{}, err
		}
		entries = append(entries, rootEntries...)
	}
	return count.Aggregate(entries)
}

func (c *counter) normalizeExtensions(exts []string) ([]string, error) {
	if len(exts) == 0 {
		return nil, nil
	}
	selectedLangs := make(map[string]lang.LanguageSpec)
	unsupported := make([]string, 0)
	for _, ext := range exts {
		norm := lang.NormalizeExtension(ext)
		if norm == "" {
			continue
		}
		langs, ok := c.registry.ByExtension(norm)
		if !ok {
			unsupported = append(unsupported, norm)
			continue
		}
		for _, spec := range langs {
			selectedLangs[spec.ID] = spec
		}
	}
	if len(unsupported) > 0 {
		sort.Strings(unsupported)
		return nil, fmt.Errorf("unsupported extensions: %s", strings.Join(unsupported, ", "))
	}

	seen := make(map[string]bool)
	normalized := make([]string, 0)
	for _, spec := range selectedLangs {
		for _, ext := range spec.Extensions {
			norm := lang.NormalizeExtension(ext)
			if norm == "" || seen[norm] {
				continue
			}
			seen[norm] = true
			normalized = append(normalized, norm)
		}
	}
	sort.Strings(normalized)
	return normalized, nil
}

func normalizeRoots(req model.CountRequest) []string {
	if len(req.Roots) > 0 {
		roots := make([]string, 0, len(req.Roots))
		for _, root := range req.Roots {
			trimmed := strings.TrimSpace(root)
			if trimmed == "" {
				continue
			}
			roots = append(roots, trimmed)
		}
		if len(roots) > 0 {
			return roots
		}
		return []string{"."}
	}
	root := strings.TrimSpace(req.Root)
	if root == "" {
		root = "."
	}
	return []string{root}
}
