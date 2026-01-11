package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/nchapman/rugby/internal/builder"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Download dependencies",
	Long: `Downloads all dependencies specified in rugby.mod.

This is a wrapper around 'go mod download' that operates on the
project's generated go.mod in .rugby/gen/.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInstall()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall() error {
	// Find project root
	project, err := builder.FindProject()
	if err != nil {
		return err
	}

	// Ensure .rugby directories exist
	if err := project.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create .rugby directory: %w", err)
	}

	// We need a go.mod in .rugby/gen/ for go mod download to work.
	// Create a minimal builder just to set up the gen directory.
	b := builder.New(project, builder.WithVerbose(verbose))
	if err := b.SetupGenDir(); err != nil {
		return fmt.Errorf("failed to set up gen directory: %w", err)
	}

	logger.Info("Downloading dependencies...")

	// Run go mod download
	goCmd := exec.Command("go", "mod", "download")
	goCmd.Dir = project.GenDir
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr

	if err := goCmd.Run(); err != nil {
		return fmt.Errorf("go mod download failed: %w", err)
	}

	logger.Info("Dependencies installed")
	return nil
}
