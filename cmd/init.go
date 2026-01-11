package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new Rugby project",
	Long: `Creates a new Rugby project with rugby.mod, a starter main.rg, and .gitignore.

If name is provided, creates a new directory with that name.
Otherwise, initializes in the current directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectDir string
		var moduleName string

		if len(args) > 0 {
			// Create new directory
			projectDir = args[0]
			moduleName = filepath.Base(projectDir)
			if err := os.MkdirAll(projectDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			logger.Info("Created directory", "name", projectDir)
		} else {
			// Use current directory
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			projectDir = cwd
			moduleName = filepath.Base(cwd)
		}

		// Check if rugby.mod already exists
		rugbyModPath := filepath.Join(projectDir, "rugby.mod")
		if _, err := os.Stat(rugbyModPath); err == nil {
			return fmt.Errorf("rugby.mod already exists in %s", projectDir)
		}

		// Create rugby.mod
		rugbyModContent := fmt.Sprintf(`module %s

go 1.25
`, moduleName)

		if err := os.WriteFile(rugbyModPath, []byte(rugbyModContent), 0644); err != nil {
			return fmt.Errorf("failed to create rugby.mod: %w", err)
		}
		logger.Info("Created", "file", "rugby.mod")

		// Create main.rg
		mainRgPath := filepath.Join(projectDir, "main.rg")
		mainRgContent := `# Welcome to Rugby!
puts "Hello, Rugby!"
`
		if err := os.WriteFile(mainRgPath, []byte(mainRgContent), 0644); err != nil {
			return fmt.Errorf("failed to create main.rg: %w", err)
		}
		logger.Info("Created", "file", "main.rg")

		// Create or update .gitignore
		gitignorePath := filepath.Join(projectDir, ".gitignore")
		var gitignoreContent string

		existing, err := os.ReadFile(gitignorePath)
		if err == nil {
			// .gitignore exists, check if .rugby/ is already there
			if strings.Contains(string(existing), ".rugby/") {
				gitignoreContent = "" // Already has it, skip
			} else {
				gitignoreContent = string(existing) + "\n# Rugby build artifacts\n.rugby/\n"
			}
		} else {
			// Create new .gitignore
			gitignoreContent = "# Rugby build artifacts\n.rugby/\n"
		}

		if gitignoreContent != "" {
			if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
				return fmt.Errorf("failed to create .gitignore: %w", err)
			}
			logger.Info("Created", "file", ".gitignore")
		}

		fmt.Printf("\nProject initialized! Run:\n")
		if len(args) > 0 {
			fmt.Printf("  cd %s\n", args[0])
		}
		fmt.Printf("  rugby run main.rg\n")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
