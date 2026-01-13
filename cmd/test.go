package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nchapman/rugby/internal/builder"
)

var testVerbose bool

var testCmd = &cobra.Command{
	Use:   "test [patterns...] [-- go-test-args]",
	Short: "Run Rugby tests",
	Long: `Compiles *_test.rg files and runs them with go test.

Examples:
  rugby test                    # Run all tests in current directory
  rugby test ./...              # Run all tests recursively
  rugby test -v                 # Run tests with verbose output
  rugby test -- -run TestFoo    # Pass arguments to go test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find go test arguments after --
		var patterns []string
		var goTestArgs []string

		if dashIndex := cmd.ArgsLenAtDash(); dashIndex >= 0 {
			patterns = args[:dashIndex]
			goTestArgs = args[dashIndex:]
		} else {
			patterns = args
		}

		// Default to current directory if no patterns
		if len(patterns) == 0 {
			patterns = []string{"."}
		}

		// Add verbose flag if set
		if testVerbose {
			goTestArgs = append([]string{"-v"}, goTestArgs...)
		}

		// Find project
		project, err := builder.FindProjectFrom(".")
		if err != nil {
			return err
		}

		b := builder.New(project, builder.WithVerbose(verbose), builder.WithColorMode(getColorMode()))
		return b.Test(patterns, goTestArgs)
	},
}

func init() {
	testCmd.Flags().BoolVarP(&testVerbose, "verbose", "V", false, "verbose test output (go test -v)")
	rootCmd.AddCommand(testCmd)
}
