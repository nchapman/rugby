// Package file provides file I/O operations for Rugby programs.
// Rugby: import rugby/file
//
// Example:
//
//	content = file.read("config.txt")!
//	file.write("output.txt", content)!
package file

import (
	"bufio"
	"os"
	"strings"
)

// Read reads a file and returns its contents as a string.
// Ruby: file.read(path)
func Read(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadBytes reads a file and returns its contents as bytes.
// Ruby: file.read_bytes(path)
func ReadBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ReadLines reads a file and returns its lines as an array.
// Ruby: file.read_lines(path)
func ReadLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var lines []string
	scanner := bufio.NewScanner(f)
	// Increase buffer to handle lines up to 1MB
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// Write writes content to a file, creating it if it doesn't exist.
// Ruby: file.write(path, content)
func Write(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// WriteBytes writes bytes to a file, creating it if it doesn't exist.
// Ruby: file.write_bytes(path, bytes)
func WriteBytes(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// WriteLines writes lines to a file, joining them with newlines.
// Ruby: file.write_lines(path, lines)
func WriteLines(path string, lines []string) error {
	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// Append appends content to a file, creating it if it doesn't exist.
// Ruby: file.append(path, content)
func Append(path string, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(content)
	return err
}

// AppendBytes appends bytes to a file, creating it if it doesn't exist.
// Ruby: file.append_bytes(path, bytes)
func AppendBytes(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.Write(data)
	return err
}

// Exists reports whether a file or directory exists at the path.
// Ruby: file.exists?(path)
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Size returns the size of a file in bytes.
// Ruby: file.size(path)
func Size(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// Directory reports whether the path is a directory.
// Ruby: file.directory?(path)
func Directory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// File reports whether the path is a regular file.
// Ruby: file.file?(path)
func File(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

// Remove deletes a file or empty directory.
// Ruby: file.remove(path)
func Remove(path string) error {
	return os.Remove(path)
}

// RemoveAll deletes a path and any children it contains.
// Ruby: file.remove_all(path)
func RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Rename renames (moves) a file or directory.
// Ruby: file.rename(old, new)
func Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// Copy copies a file from src to dst, preserving permissions.
// Ruby: file.copy(src, dst)
func Copy(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode())
}

// Mkdir creates a directory.
// Ruby: file.mkdir(path)
func Mkdir(path string) error {
	return os.Mkdir(path, 0755)
}

// MkdirAll creates a directory and all parent directories.
// Ruby: file.mkdir_all(path)
func MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}
