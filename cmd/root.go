package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/nchapman/rugby/internal/builder"
)

var (
	verbose     bool
	showVersion bool
	logger      = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: false,
	})
)

var rootCmd = &cobra.Command{
	Use:   "rugby",
	Short: "Rugby - A Ruby-like language that compiles to Go",
	Long: `Rugby is a compiler that transforms Ruby-like syntax into idiomatic Go code.
It provides a joyful developer experience with zero configuration.

Commands:
  init    Initialize a new Rugby project
  run     Compile and run a Rugby file
  build   Produce an optimized binary
  add     Add a dependency to rugby.mod
  clean   Remove build artifacts
  repl    Interactive Rugby shell`,
	// Handle bare 'rugby file.rg' for backward compatibility
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle --version flag
		if showVersion {
			fmt.Println(getVersion())
			return nil
		}
		// If first arg looks like a .rg file, run it
		if len(args) > 0 && strings.HasSuffix(args[0], ".rg") {
			return runFile(args[0], nil)
		}
		return cmd.Help()
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initLogger)

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "", false, "verbose output")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "V", false, "print version")
}

func initLogger() {
	if verbose {
		logger.SetLevel(log.DebugLevel)
	}
}

// runFile is a helper used by both root (backward compat) and run commands.
func runFile(file string, args []string) error {
	project, err := builder.FindProjectFrom(file)
	if err != nil {
		return err
	}

	b := builder.New(project, builder.WithVerbose(verbose))
	return b.Run(file, args)
}
