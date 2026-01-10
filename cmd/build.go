package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"rugby/internal/builder"
)

var outputName string

var buildCmd = &cobra.Command{
	Use:   "build <file.rg> [files...]",
	Short: "Build an optimized binary",
	Long: `Compiles Rugby files and produces a standalone binary in the current directory.

Examples:
  rugby build main.rg              # Produces ./main
  rugby build main.rg -o myapp     # Produces ./myapp`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate all files have .rg extension
		for _, f := range args {
			if !strings.HasSuffix(f, ".rg") {
				return fmt.Errorf("expected .rg file, got: %s", f)
			}
		}

		project, err := builder.FindProjectFrom(args[0])
		if err != nil {
			return err
		}

		b := builder.New(project, builder.WithVerbose(verbose))
		return b.Build(args, outputName)
	},
}

func init() {
	buildCmd.Flags().StringVarP(&outputName, "output", "o", "", "output binary name")
	rootCmd.AddCommand(buildCmd)
}
