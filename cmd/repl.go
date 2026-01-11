package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nchapman/rugby/internal/repl"
)

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start an interactive Rugby shell",
	Long: `Start an interactive REPL (Read-Eval-Print Loop) for Rugby.

Features:
  - Expression results are automatically printed
  - Variables persist across inputs
  - Functions and classes can be defined
  - Multi-line input supported (def...end, if...end, etc.)
  - Up/Down arrows navigate history

Press Ctrl+D to exit.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return repl.Run()
	},
}

func init() {
	rootCmd.AddCommand(replCmd)
}
