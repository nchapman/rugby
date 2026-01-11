// Package path provides path manipulation for Rugby programs.
// These are pure functions that do not perform I/O.
// Rugby: import rugby/path
//
// Example:
//
//	full = path.join("/home", "user", "file.txt")
//	dir = path.dirname(full)
//	ext = path.extname(full)
package path

import (
	"path/filepath"
	"strings"
)

// Join joins path elements into a single path.
// Ruby: path.join(parts...)
func Join(parts ...string) string {
	return filepath.Join(parts...)
}

// Split splits a path into directory and file components.
// Ruby: path.split(path)
func Split(path string) (string, string) {
	return filepath.Split(path)
}

// Dirname returns the directory portion of a path.
// Ruby: path.dirname(path)
func Dirname(path string) string {
	return filepath.Dir(path)
}

// Basename returns the last element of a path.
// Ruby: path.basename(path)
func Basename(path string) string {
	return filepath.Base(path)
}

// Extname returns the file extension including the dot.
// Returns empty string if there is no extension.
// Ruby: path.extname(path)
func Extname(path string) string {
	return filepath.Ext(path)
}

// Absolute reports whether the path is absolute.
// Ruby: path.absolute?(path)
func Absolute(path string) bool {
	return filepath.IsAbs(path)
}

// Relative returns a relative path from base to target.
// Ruby: path.relative(base, target)
func Relative(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

// Clean returns the shortest path equivalent to path.
// Ruby: path.clean(path)
func Clean(path string) string {
	return filepath.Clean(path)
}

// Abs returns the absolute path of a relative path.
// Ruby: path.abs(path)
func Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// Stem returns the file name without extension.
// Ruby: path.stem(path)
func Stem(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// Match reports whether name matches the shell pattern.
// Returns false if the pattern is malformed.
// Ruby: path.match?(pattern, name)
func Match(pattern, name string) bool {
	matched, _ := filepath.Match(pattern, name)
	return matched
}
