package cmd

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <file.rg> [-- args...]",
	Short: "Compile and run a Rugby file",
	Long: `Transpiles the file to Go, builds it, and executes immediately.

Pass arguments to the program after --:
  rugby run main.rg -- --port 8080`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]

		if err := ValidateRgFile(file); err != nil {
			return err
		}

		// Arguments after -- go to the program
		var programArgs []string
		if dashIndex := cmd.ArgsLenAtDash(); dashIndex >= 0 {
			programArgs = args[dashIndex:]
		}

		return runFile(file, programArgs)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
