// Package cmd implements the Rugby CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nchapman/rugby/internal/builder"
)

var outputName string

var buildCmd = &cobra.Command{
	Use:   "build <path> [paths...]",
	Short: "Build an optimized binary",
	Long: `Compiles Rugby files and produces a standalone binary in the current directory.

Accepts files, directories, or patterns:
  file.rg    Build a specific file
  dir/       Build all .rg files in directory (non-recursive)
  ./...      Build all .rg files recursively

Examples:
  rugby build main.rg              # Produces ./main
  rugby build main.rg -o myapp     # Produces ./myapp
  rugby build src/                 # Build all .rg files in src/
  rugby build ./...                # Build all .rg files recursively`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		files, err := ExpandRgPaths(args)
		if err != nil {
			return err
		}

		if len(files) == 0 {
			return fmt.Errorf("no .rg files found")
		}

		project, err := builder.FindProjectFrom(files[0])
		if err != nil {
			return err
		}

		b := builder.New(project, builder.WithVerbose(verbose), builder.WithColorMode(getColorMode()))
		return b.Build(files, outputName)
	},
}

func init() {
	buildCmd.Flags().StringVarP(&outputName, "output", "o", "", "output binary name")
	rootCmd.AddCommand(buildCmd)
}
