package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nchapman/rugby/formatter"
)

var checkOnly bool

var formatCmd = &cobra.Command{
	Use:   "format <path> [paths...]",
	Short: "Format Rugby source files",
	Long: `Formats .rg files with consistent style (2-space indent, standard spacing).

Accepts files, directories, or patterns:
  .          Format all .rg files in current directory
  ./...      Format all .rg files recursively
  dir/       Format all .rg files in directory (non-recursive)
  file.rg    Format a specific file

Examples:
  rugby format .               # Format all .rg files in current directory
  rugby format ./...           # Format all .rg files recursively
  rugby format src/            # Format .rg files in src/ directory
  rugby format main.rg lib.rg  # Format specific files
  rugby format --check ./...   # Check if all files are formatted`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Collect all files to format
		var files []string
		for _, arg := range args {
			expanded, err := expandPath(arg)
			if err != nil {
				return err
			}
			files = append(files, expanded...)
		}

		// Deduplicate files (e.g., if user runs `rugby format . main.rg`)
		files = deduplicateFiles(files)

		if len(files) == 0 {
			logger.Info("no .rg files found")
			return nil
		}

		hasChanges := false
		for _, file := range files {
			changed, err := formatFile(file)
			if err != nil {
				return err
			}
			if changed {
				hasChanges = true
			}
		}

		if checkOnly && hasChanges {
			return fmt.Errorf("some files are not formatted")
		}

		return nil
	},
}

// deduplicateFiles removes duplicate file paths.
func deduplicateFiles(files []string) []string {
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

// expandPath expands a path argument into a list of .rg files.
func expandPath(path string) ([]string, error) {
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

// findRgFilesRecursive finds all .rg files recursively.
func findRgFilesRecursive(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden directories and .rugby build directory
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

// formatFile formats a single file, returns true if file was changed.
func formatFile(file string) (bool, error) {
	info, err := os.Stat(file)
	if err != nil {
		return false, fmt.Errorf("stat %s: %w", file, err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return false, fmt.Errorf("reading %s: %w", file, err)
	}

	formatted, err := formatter.Format(string(content))
	if err != nil {
		return false, fmt.Errorf("formatting %s: %w", file, err)
	}

	if string(content) == formatted {
		if verbose {
			logger.Info("already formatted", "file", file)
		}
		return false, nil
	}

	if checkOnly {
		logger.Warn("not formatted", "file", file)
		return true, nil
	}

	if err := os.WriteFile(file, []byte(formatted), info.Mode()); err != nil {
		return false, fmt.Errorf("writing %s: %w", file, err)
	}
	logger.Info("formatted", "file", file)
	return true, nil
}

func init() {
	formatCmd.Flags().BoolVar(&checkOnly, "check", false,
		"check if files are formatted (exit 1 if not)")
	rootCmd.AddCommand(formatCmd)
}
