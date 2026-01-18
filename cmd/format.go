package cmd

import (
	"fmt"
	"os"

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
		files, err := ExpandRgPaths(args)
		if err != nil {
			return err
		}

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
