package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExpandRgPaths expands path arguments into a list of .rg files.
// Supports: single files, directories (non-recursive), and ./... patterns (recursive).
func ExpandRgPaths(paths []string) ([]string, error) {
	var files []string
	for _, path := range paths {
		expanded, err := expandRgPath(path)
		if err != nil {
			return nil, err
		}
		files = append(files, expanded...)
	}
	return DeduplicateFiles(files), nil
}

// expandRgPath expands a single path argument into a list of .rg files.
func expandRgPath(path string) ([]string, error) {
	// Handle ./... recursive pattern
	if path == "./..." || strings.HasSuffix(path, "/...") {
		baseDir := "."
		if path != "./..." {
			baseDir = strings.TrimSuffix(path, "/...")
		}
		return findRgFilesRecursive(baseDir)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}

	// If it's a directory, find .rg files in it (non-recursive)
	if info.IsDir() {
		return findRgFilesInDir(path)
	}

	// Single file - must be .rg
	if !strings.HasSuffix(path, ".rg") {
		return nil, fmt.Errorf("expected .rg file, got: %s", path)
	}

	return []string{path}, nil
}

// findRgFilesInDir finds .rg files in a directory (non-recursive).
func findRgFilesInDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".rg") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files, nil
}

// findRgFilesRecursive finds all .rg files recursively, skipping hidden dirs and build artifacts.
func findRgFilesRecursive(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden directories and common build/vendor directories
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".rg") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// DeduplicateFiles removes duplicate file paths, normalizing to absolute paths.
func DeduplicateFiles(files []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, f := range files {
		abs, err := filepath.Abs(f)
		if err != nil {
			abs = f // fallback to original if Abs fails
		}
		if !seen[abs] {
			seen[abs] = true
			result = append(result, f)
		}
	}
	return result
}

// ValidateRgFile checks that a path is a .rg file and exists.
func ValidateRgFile(path string) error {
	if !strings.HasSuffix(path, ".rg") {
		return fmt.Errorf("expected .rg file, got: %s", path)
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("cannot access %s: %w", path, err)
	}
	return nil
}
