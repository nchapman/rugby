package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version is set via ldflags at build time, or falls back to build info.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Rugby version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(getVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// getVersion returns the version string with build info if available.
func getVersion() string {
	version := Version

	// If built with go install or go build, try to get version from build info
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
			version = info.Main.Version
		}
	}

	return "rugby " + version
}
