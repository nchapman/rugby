package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"rugby/codegen"
	"rugby/lexer"
	"rugby/parser"
)

// Styles for pretty output
var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	fileStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	mutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// Builder orchestrates the Rugby compilation process.
type Builder struct {
	project *Project
	verbose bool
	logger  *log.Logger
}

// BuilderOption configures a Builder.
type BuilderOption func(*Builder)

// WithVerbose enables verbose output.
func WithVerbose(v bool) BuilderOption {
	return func(b *Builder) {
		b.verbose = v
	}
}

// New creates a new Builder for the given project.
func New(project *Project, opts ...BuilderOption) *Builder {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: false,
	})

	b := &Builder{
		project: project,
		logger:  logger,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// CompileResult holds the result of a compilation.
type CompileResult struct {
	GenFiles   []string // Generated .go files
	BinaryPath string   // Path to compiled binary (if built)
}

// Compile transpiles .rg files to .go files in .rugby/gen/.
func (b *Builder) Compile(files []string) (*CompileResult, error) {
	if err := b.project.EnsureDirs(); err != nil {
		return nil, fmt.Errorf("failed to create .rugby directory: %w", err)
	}

	result := &CompileResult{}

	for _, rgFile := range files {
		genFile, err := b.compileFile(rgFile)
		if err != nil {
			return nil, err
		}
		result.GenFiles = append(result.GenFiles, genFile)
	}

	// Set up the gen directory for building
	if err := b.setupGenDir(); err != nil {
		return nil, err
	}

	return result, nil
}

// compileFile transpiles a single .rg file to .go.
func (b *Builder) compileFile(inputPath string) (string, error) {
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		return "", fmt.Errorf("invalid path %s: %w", inputPath, err)
	}

	source, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("error reading %s: %w", inputPath, err)
	}

	// Lex
	l := lexer.New(string(source))

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return "", b.formatParseErrors(inputPath, p.Errors())
	}

	// Generate with //line directive pointing to original source
	gen := codegen.New(codegen.WithSourceFile(absPath))
	output, err := gen.Generate(program)
	if err != nil {
		return "", fmt.Errorf("code generation error in %s: %w", inputPath, err)
	}

	// Write to .rugby/gen/
	outputPath := b.project.GenPath(absPath)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		return "", fmt.Errorf("error writing %s: %w", outputPath, err)
	}

	if b.verbose {
		b.logger.Info("compiled", "file", b.project.RelPath(absPath))
	}

	return outputPath, nil
}

// formatParseErrors formats parser errors for display.
func (b *Builder) formatParseErrors(file string, errors []string) error {
	var msg strings.Builder
	msg.WriteString(errorStyle.Render("Parse errors") + " in " + fileStyle.Render(file) + ":\n")
	for _, e := range errors {
		msg.WriteString(fmt.Sprintf("  %s:%s\n", file, e))
	}
	return fmt.Errorf("%s", msg.String())
}

// setupGenDir prepares the gen directory for go build.
func (b *Builder) setupGenDir() error {
	// Create go.mod in gen directory that references the parent
	goModContent := fmt.Sprintf(`module main

go 1.25

require rugby v0.0.0

replace rugby => %s
`, b.project.Root)

	goModPath := filepath.Join(b.project.GenDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("error creating go.mod: %w", err)
	}

	// Run go mod tidy to resolve dependencies
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = b.project.GenDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %s", string(output))
	}

	return nil
}

// Build compiles and links the program, producing a binary.
func (b *Builder) Build(files []string, outputName string) error {
	result, err := b.Compile(files)
	if err != nil {
		return err
	}

	// Determine output name
	if outputName == "" {
		// Default to first file's basename without extension
		base := filepath.Base(files[0])
		outputName = strings.TrimSuffix(base, filepath.Ext(base))
	}

	// Get absolute output path (in cwd)
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	outputPath := filepath.Join(cwd, outputName)

	// Run go build
	if err := b.goBuild(result.GenFiles, outputPath); err != nil {
		return err
	}

	b.logger.Info(successStyle.Render("Built"), "binary", outputName)
	return nil
}

// Run compiles and executes the program.
func (b *Builder) Run(file string, args []string) error {
	result, err := b.Compile([]string{file})
	if err != nil {
		return err
	}

	// Build binary to .rugby/bin/
	base := filepath.Base(file)
	binName := strings.TrimSuffix(base, filepath.Ext(base))
	binPath := b.project.BinPath(binName)

	if err := b.goBuild(result.GenFiles, binPath); err != nil {
		return err
	}

	// Execute
	return b.execute(binPath, args)
}

// goBuild runs go build in the gen directory.
func (b *Builder) goBuild(genFiles []string, outputPath string) error {
	// Build args
	args := []string{"build", "-o", outputPath}

	// Add all generated files
	for _, f := range genFiles {
		rel, err := filepath.Rel(b.project.GenDir, f)
		if err != nil {
			rel = f
		}
		args = append(args, rel)
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = b.project.GenDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return b.formatGoError(string(output))
	}

	return nil
}

// formatGoError attempts to make Go compiler errors more user-friendly.
func (b *Builder) formatGoError(output string) error {
	// The //line directives should already map errors to .rg files
	// We just clean up the output a bit
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var result []string
	for _, line := range lines {
		// Skip internal Go paths
		if strings.Contains(line, "/go/src/") {
			continue
		}
		result = append(result, line)
	}

	return fmt.Errorf("%s\n%s", errorStyle.Render("Build failed:"), strings.Join(result, "\n"))
}

// execute runs the compiled binary.
func (b *Builder) execute(binPath string, args []string) error {
	cmd := exec.Command(binPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Preserve the exit code from the executed program
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}
