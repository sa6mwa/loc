package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: count <ext1> <ext2> ...")
		os.Exit(1)
	}
	exts := os.Args[1:]
	total := 0
	// Common dependency / generated / external directories to skip
	excludedDirs := map[string]bool{
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
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if excludedDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		for _, ext := range exts {
			if strings.HasSuffix(path, ext) {
				lines, err := countLines(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
					return nil
				}
				total += lines
				break
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(total)
}

func countLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}
	return lines, scanner.Err()
}
