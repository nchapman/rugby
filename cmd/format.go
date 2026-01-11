package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nchapman/rugby/formatter"
)

var checkOnly bool

var formatCmd = &cobra.Command{
	Use:   "format <file.rg> [files...]",
	Short: "Format Rugby source files",
	Long: `Formats .rg files with consistent style (2-space indent, standard spacing).

Examples:
  rugby format main.rg           # Format single file
  rugby format *.rg              # Format multiple files
  rugby format --check main.rg   # Check if file is formatted (exit 1 if not)`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hasChanges := false

		for _, file := range args {
			if !strings.HasSuffix(file, ".rg") {
				return fmt.Errorf("expected .rg file, got: %s", file)
			}

			info, err := os.Stat(file)
			if err != nil {
				return fmt.Errorf("stat %s: %w", file, err)
			}

			content, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("reading %s: %w", file, err)
			}

			formatted, err := formatter.Format(string(content))
			if err != nil {
				return fmt.Errorf("formatting %s: %w", file, err)
			}

			if string(content) == formatted {
				if verbose {
					logger.Info("already formatted", "file", file)
				}
				continue
			}

			hasChanges = true

			if checkOnly {
				logger.Warn("not formatted", "file", file)
				continue
			}

			if err := os.WriteFile(file, []byte(formatted), info.Mode()); err != nil {
				return fmt.Errorf("writing %s: %w", file, err)
			}
			logger.Info("formatted", "file", file)
		}

		if checkOnly && hasChanges {
			return fmt.Errorf("some files are not formatted")
		}

		return nil
	},
}

func init() {
	formatCmd.Flags().BoolVar(&checkOnly, "check", false,
		"check if files are formatted (exit 1 if not)")
	rootCmd.AddCommand(formatCmd)
}
