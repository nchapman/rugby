package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <package> [version]",
	Short: "Add a dependency to rugby.mod",
	Long: `Adds a Go package as a dependency to rugby.mod.

Examples:
  rugby add github.com/gin-gonic/gin
  rugby add github.com/gin-gonic/gin v1.9.0`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg := args[0]
		version := ""
		if len(args) > 1 {
			version = args[1]
		}

		// Find rugby.mod in current or parent directories
		rugbyModPath, err := findRugbyMod()
		if err != nil {
			return err
		}

		// Read existing rugby.mod
		content, err := os.ReadFile(rugbyModPath)
		if err != nil {
			return fmt.Errorf("failed to read rugby.mod: %w", err)
		}

		// Check if exact package already exists (not just substring match)
		if packageExists(string(content), pkg) {
			logger.Warn("Package already in rugby.mod", "package", pkg)
			return nil
		}

		// Parse and update rugby.mod
		newContent := addDependency(string(content), pkg, version)

		if err := os.WriteFile(rugbyModPath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to update rugby.mod: %w", err)
		}

		if version != "" {
			logger.Info("Added dependency", "package", pkg, "version", version)
		} else {
			logger.Info("Added dependency", "package", pkg)
		}

		fmt.Println("\nRun 'rugby run' or 'rugby build' to fetch dependencies.")
		return nil
	},
}

// findRugbyMod walks up from cwd to find rugby.mod.
func findRugbyMod() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		path := filepath.Join(dir, "rugby.mod")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("rugby.mod not found. Run 'rugby init' first")
		}
		dir = parent
	}
}

// packageExists checks if an exact package (not substring) exists in rugby.mod.
func packageExists(content, pkg string) bool {
	for line := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(line)
		fields := strings.Fields(trimmed)

		// Check for package as first field (inside require block)
		if len(fields) >= 1 && fields[0] == pkg {
			return true
		}
		// Check for single-line require: "require github.com/foo/bar v1.0.0"
		if len(fields) >= 2 && fields[0] == "require" && fields[1] == pkg {
			return true
		}
	}
	return false
}

// addDependency adds a package to the rugby.mod content.
func addDependency(content, pkg, version string) string {
	// Normalize version - treat empty or "latest" as no version (let Go resolve)
	if version == "" || version == "latest" {
		version = ""
	} else if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	depLine := pkg
	if version != "" {
		depLine = fmt.Sprintf("%s %s", pkg, version)
	}

	lines := strings.Split(content, "\n")
	var result []string
	inRequireBlock := false
	added := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "require (" {
			inRequireBlock = true
			result = append(result, line)
			continue
		}

		if inRequireBlock && trimmed == ")" {
			// Add new dependency before closing paren
			result = append(result, "\t"+depLine)
			added = true
			inRequireBlock = false
		}

		result = append(result, line)
	}

	// If no require block exists, add one
	if !added {
		// Check if there's a single-line require
		hasRequire := false
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "require ") &&
				!strings.Contains(line, "(") {
				hasRequire = true
				break
			}
		}

		if hasRequire {
			// Convert single-line to block format would be complex,
			// just append another require line
			result = append(result, fmt.Sprintf("require %s", depLine))
		} else {
			// No require at all, add a require block
			result = append(result, "")
			result = append(result, "require (")
			result = append(result, "\t"+depLine)
			result = append(result, ")")
		}
	}

	return strings.Join(result, "\n")
}

func init() {
	rootCmd.AddCommand(addCmd)
}
