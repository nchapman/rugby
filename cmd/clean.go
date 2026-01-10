package cmd

import (
	"github.com/spf13/cobra"

	"rugby/internal/builder"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove build artifacts",
	Long:  `Wipes the .rugby/ directory to force a fresh build.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := builder.FindProject()
		if err != nil {
			return err
		}

		logger.Info("Cleaning build artifacts", "dir", project.RugbyDir)
		return project.Clean()
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
