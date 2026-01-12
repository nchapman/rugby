// Package builder orchestrates the Rugby compilation process.
package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/codegen"
	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/parser"
)

// Styles for pretty output
var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	fileStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
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

	// Track which file has top-level statements (for single-entry rule)
	var entryPointFile string

	for _, rgFile := range files {
		genFile, hasTopLevel, err := b.compileFile(rgFile)
		if err != nil {
			return nil, err
		}
		result.GenFiles = append(result.GenFiles, genFile)

		// Enforce single-entry rule: only one file may have top-level statements
		if hasTopLevel {
			if entryPointFile != "" {
				return nil, fmt.Errorf(
					"multiple files with top-level statements: %s and %s\n"+
						"Only one file may contain executable code at the top level (see spec 2.1)",
					entryPointFile, rgFile)
			}
			entryPointFile = rgFile
		}
	}

	// Set up the gen directory for building
	if err := b.SetupGenDir(); err != nil {
		return nil, err
	}

	return result, nil
}

// compileFile transpiles a single .rg file to .go.
// Returns (outputPath, hasTopLevelStmts, error)
func (b *Builder) compileFile(inputPath string) (string, bool, error) {
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		return "", false, fmt.Errorf("invalid path %s: %w", inputPath, err)
	}

	outputPath := b.project.GenPath(absPath)
	metaPath := outputPath + ".meta"

	// Check if we can skip compilation (output newer than source)
	if b.isUpToDate(absPath, outputPath) {
		if b.verbose {
			b.logger.Debug("up to date, skipping", "file", b.project.RelPath(absPath))
		}
		// Read cached metadata to get hasTopLevel
		hasTopLevel := b.readCachedMeta(metaPath)
		return outputPath, hasTopLevel, nil
	}

	source, err := os.ReadFile(absPath)
	if err != nil {
		return "", false, fmt.Errorf("error reading %s: %w", inputPath, err)
	}

	// Lex
	l := lexer.New(string(source))

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return "", false, b.formatParseErrors(inputPath, p.Errors())
	}

	// Check if this file has top-level statements
	hasTopLevel := hasTopLevelStatements(program)

	// Generate with //line directive pointing to original source
	gen := codegen.New(codegen.WithSourceFile(absPath))
	output, err := gen.Generate(program)
	if err != nil {
		return "", false, fmt.Errorf("code generation error in %s: %w", inputPath, err)
	}

	// Write to .rugby/gen/
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", false, fmt.Errorf("error creating directory: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		return "", false, fmt.Errorf("error writing %s: %w", outputPath, err)
	}

	// Save metadata for cache
	b.writeCachedMeta(metaPath, hasTopLevel)

	if b.verbose {
		b.logger.Info("compiled", "file", b.project.RelPath(absPath))
	}

	return outputPath, hasTopLevel, nil
}

// readCachedMeta reads the hasTopLevel metadata from a cache file.
func (b *Builder) readCachedMeta(metaPath string) bool {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return false // conservative default if no cache
	}
	return strings.TrimSpace(string(data)) == "toplevel=true"
}

// writeCachedMeta writes the hasTopLevel metadata to a cache file.
func (b *Builder) writeCachedMeta(metaPath string, hasTopLevel bool) {
	content := "toplevel=false"
	if hasTopLevel {
		content = "toplevel=true"
	}
	// Ignore errors - cache is best-effort
	_ = os.WriteFile(metaPath, []byte(content), 0644)
}

// isUpToDate checks if the output file is newer than the source file.
func (b *Builder) isUpToDate(sourcePath, outputPath string) bool {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return false
	}

	outputInfo, err := os.Stat(outputPath)
	if err != nil {
		return false // output doesn't exist
	}

	return outputInfo.ModTime().After(sourceInfo.ModTime())
}

// hasTopLevelStatements checks if a program contains executable top-level statements
// (anything other than def, class, or interface)
func hasTopLevelStatements(program *ast.Program) bool {
	for _, decl := range program.Declarations {
		switch decl.(type) {
		case *ast.FuncDecl, *ast.ClassDecl, *ast.InterfaceDecl:
			// These are definitions, not executable statements
			continue
		default:
			// Any other statement is a top-level executable statement
			return true
		}
	}
	return false
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

// RuntimeModule is the Go module that contains the Rugby runtime.
const RuntimeModule = "github.com/nchapman/rugby"

// RuntimeVersion is the version of the runtime to require in generated go.mod files.
const RuntimeVersion = "v0.1.1"

// IsInRugbyRepo checks if we're running from within the rugby repository.
// Returns (true, repoPath) if we're in the repo, (false, "") otherwise.
// This enables local development by auto-injecting a replace directive.
func IsInRugbyRepo() (bool, string) {
	cwd, err := os.Getwd()
	if err != nil {
		return false, ""
	}

	// Walk up the directory tree looking for go.mod with our module
	dir := cwd
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(goModPath); err == nil {
			// Check if this is the rugby module
			if strings.Contains(string(data), "module "+RuntimeModule) {
				return true, dir
			}
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return false, ""
}

// SetupGenDir prepares the gen directory for go build.
// Generates .rugby/go.mod from rugby.mod (if present) + injects runtime dep.
func (b *Builder) SetupGenDir() error {
	goModPath := filepath.Join(b.project.GenDir, "go.mod")
	goSumPath := filepath.Join(b.project.GenDir, "go.sum")

	// Check if we need to regenerate go.mod
	needsUpdate := b.needsGoModUpdate(goModPath)
	if !needsUpdate {
		if b.verbose {
			b.logger.Debug("go.mod up to date, skipping regeneration")
		}
		return nil
	}

	goModContent := b.generateGoMod()

	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("error creating go.mod: %w", err)
	}

	if b.verbose {
		b.logger.Info("generated", "file", ".rugby/go.mod")
	}

	// Only run go mod tidy if go.sum doesn't exist or go.mod is newer
	if b.needsModTidy(goModPath, goSumPath) {
		if b.verbose {
			b.logger.Debug("running go mod tidy")
		}
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = b.project.GenDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go mod tidy failed: %s", string(output))
		}
	}

	return nil
}

// needsModTidy checks if go mod tidy needs to run.
func (b *Builder) needsModTidy(goModPath, goSumPath string) bool {
	goSumInfo, err := os.Stat(goSumPath)
	if err != nil {
		return true // go.sum doesn't exist
	}

	goModInfo, err := os.Stat(goModPath)
	if err != nil {
		return true // go.mod doesn't exist (shouldn't happen)
	}

	// Run tidy if go.mod is newer than go.sum
	return goModInfo.ModTime().After(goSumInfo.ModTime())
}

// needsGoModUpdate checks if .rugby/go.mod needs to be regenerated.
// Returns true if rugby.mod is newer than .rugby/go.mod or if .rugby/go.mod doesn't exist.
func (b *Builder) needsGoModUpdate(goModPath string) bool {
	goModInfo, err := os.Stat(goModPath)
	if err != nil {
		// .rugby/go.mod doesn't exist, need to create it
		return true
	}

	// If rugby.mod exists, check if it's newer than .rugby/go.mod
	rugbyModPath := b.project.RugbyModPath()
	rugbyModInfo, err := os.Stat(rugbyModPath)
	if err != nil {
		// No rugby.mod, but we still need go.mod - only regenerate if missing
		return false
	}

	// Regenerate if rugby.mod is newer than .rugby/go.mod
	return rugbyModInfo.ModTime().After(goModInfo.ModTime())
}

// generateGoMod creates the go.mod content for .rugby/go.mod.
// Reads rugby.mod if present and injects the runtime dependency.
func (b *Builder) generateGoMod() string {
	rugbyModPath := b.project.RugbyModPath()

	// Read rugby.mod if it exists
	rugbyModContent, err := os.ReadFile(rugbyModPath)
	if err != nil {
		// No rugby.mod - generate minimal go.mod
		result := fmt.Sprintf(`module main

go 1.25

require %s %s
`, RuntimeModule, RuntimeVersion)
		// Auto-detect local development and inject replace directive
		if inRepo, repoPath := IsInRugbyRepo(); inRepo {
			result += fmt.Sprintf("\nreplace %s => %s\n", RuntimeModule, repoPath)
		}
		return result
	}

	// Parse rugby.mod and augment it
	var out strings.Builder
	lines := strings.Split(string(rugbyModContent), "\n")

	hasRuntimeDep := false
	inRequireBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if runtime is already in rugby.mod (unlikely but possible)
		if strings.Contains(trimmed, RuntimeModule) {
			hasRuntimeDep = true
		}

		// Track require block state
		if trimmed == "require (" {
			inRequireBlock = true
		}
		if inRequireBlock && trimmed == ")" {
			// Inject runtime before closing the require block
			if !hasRuntimeDep {
				out.WriteString(fmt.Sprintf("\t%s %s\n", RuntimeModule, RuntimeVersion))
			}
			inRequireBlock = false
		}

		out.WriteString(line)
		out.WriteString("\n")
	}

	result := out.String()

	// If no require block exists, add one with the runtime
	if !hasRuntimeDep && !strings.Contains(result, "require") {
		result += fmt.Sprintf("\nrequire %s %s\n", RuntimeModule, RuntimeVersion)
	}

	// Handle single-line require statement (require foo v1.0.0)
	if !hasRuntimeDep && strings.Contains(result, "require ") && !strings.Contains(result, "require (") {
		result += fmt.Sprintf("require %s %s\n", RuntimeModule, RuntimeVersion)
	}

	// Auto-detect local development and inject replace directive
	if inRepo, repoPath := IsInRugbyRepo(); inRepo {
		result += fmt.Sprintf("\nreplace %s => %s\n", RuntimeModule, repoPath)
	}

	return result
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
	if err := b.GoBuild(result.GenFiles, outputPath); err != nil {
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

	// Skip build if binary is up to date
	if !b.needsRebuild(result.GenFiles, binPath) {
		if b.verbose {
			b.logger.Debug("binary up to date, skipping build")
		}
	} else {
		if err := b.GoBuild(result.GenFiles, binPath); err != nil {
			return err
		}
	}

	// Execute
	return b.execute(binPath, args)
}

// needsRebuild checks if the binary needs to be rebuilt.
// Returns true if binary doesn't exist or any source file is newer.
func (b *Builder) needsRebuild(genFiles []string, binPath string) bool {
	binInfo, err := os.Stat(binPath)
	if err != nil {
		return true // binary doesn't exist
	}
	binTime := binInfo.ModTime()

	// Check if any generated file is newer than the binary
	for _, genFile := range genFiles {
		genInfo, err := os.Stat(genFile)
		if err != nil {
			return true // can't stat, rebuild to be safe
		}
		if genInfo.ModTime().After(binTime) {
			return true
		}
	}

	// Also check go.mod and go.sum
	goModPath := filepath.Join(b.project.GenDir, "go.mod")
	if info, err := os.Stat(goModPath); err == nil && info.ModTime().After(binTime) {
		return true
	}

	goSumPath := filepath.Join(b.project.GenDir, "go.sum")
	if info, err := os.Stat(goSumPath); err == nil && info.ModTime().After(binTime) {
		return true
	}

	return false
}

// GoBuild runs go build in the gen directory.
func (b *Builder) GoBuild(genFiles []string, outputPath string) error {
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

// Test compiles and runs Rugby tests.
// Patterns can be directories or file glob patterns.
func (b *Builder) Test(patterns []string, goTestArgs []string) error {
	// Find all *_test.rg files matching patterns
	testFiles, err := b.findTestFiles(patterns)
	if err != nil {
		return err
	}

	if len(testFiles) == 0 {
		b.logger.Info("no test files found")
		return nil
	}

	// Also compile non-test files (the code being tested)
	sourceFiles, err := b.findSourceFiles(patterns)
	if err != nil {
		return err
	}

	// Compile all files (tests + sources)
	allFiles := append(sourceFiles, testFiles...)
	if len(allFiles) > 0 {
		_, err = b.Compile(allFiles)
		if err != nil {
			return err
		}
	}

	// Run go test
	return b.GoTest(goTestArgs)
}

// findTestFiles finds all *_test.rg files in the given patterns.
func (b *Builder) findTestFiles(patterns []string) ([]string, error) {
	var result []string

	for _, pattern := range patterns {
		// Handle recursive pattern ./...
		if dir, found := strings.CutSuffix(pattern, "/..."); found {
			if dir == "." || dir == "" {
				dir = b.project.Root
			}
			files, err := b.walkTestFiles(dir)
			if err != nil {
				return nil, err
			}
			result = append(result, files...)
			continue
		}

		// Handle direct file
		if strings.HasSuffix(pattern, "_test.rg") {
			result = append(result, pattern)
			continue
		}

		// Handle directory - find *_test.rg files in it
		info, err := os.Stat(pattern)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			entries, err := os.ReadDir(pattern)
			if err != nil {
				return nil, err
			}
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.rg") {
					result = append(result, filepath.Join(pattern, entry.Name()))
				}
			}
		}
	}

	return result, nil
}

// walkTestFiles recursively finds all *_test.rg files in a directory.
func (b *Builder) walkTestFiles(dir string) ([]string, error) {
	var result []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip .rugby directory
		if info.IsDir() && info.Name() == ".rugby" {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "_test.rg") {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}

// findSourceFiles finds all .rg source files (non-test) in the given patterns.
func (b *Builder) findSourceFiles(patterns []string) ([]string, error) {
	var result []string

	for _, pattern := range patterns {
		// Handle recursive pattern ./...
		if dir, found := strings.CutSuffix(pattern, "/..."); found {
			if dir == "." || dir == "" {
				dir = b.project.Root
			}
			files, err := b.walkSourceFiles(dir)
			if err != nil {
				return nil, err
			}
			result = append(result, files...)
			continue
		}

		// Handle direct file
		if strings.HasSuffix(pattern, ".rg") && !strings.HasSuffix(pattern, "_test.rg") {
			result = append(result, pattern)
			continue
		}

		// Handle directory - find .rg files (non-test) in it
		info, err := os.Stat(pattern)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			entries, err := os.ReadDir(pattern)
			if err != nil {
				return nil, err
			}
			for _, entry := range entries {
				name := entry.Name()
				if !entry.IsDir() && strings.HasSuffix(name, ".rg") && !strings.HasSuffix(name, "_test.rg") {
					result = append(result, filepath.Join(pattern, name))
				}
			}
		}
	}

	return result, nil
}

// walkSourceFiles recursively finds all .rg source files (non-test) in a directory.
func (b *Builder) walkSourceFiles(dir string) ([]string, error) {
	var result []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip .rugby directory
		if info.IsDir() && info.Name() == ".rugby" {
			return filepath.SkipDir
		}
		name := info.Name()
		if !info.IsDir() && strings.HasSuffix(name, ".rg") && !strings.HasSuffix(name, "_test.rg") {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}

// GoTest runs go test in the gen directory.
func (b *Builder) GoTest(args []string) error {
	// Build the command
	testArgs := []string{"test"}
	testArgs = append(testArgs, args...)

	// Only append ./... if no package pattern is specified
	hasPackagePattern := false
	for _, arg := range args {
		// Package patterns start with . or contain /
		if strings.HasPrefix(arg, ".") || strings.Contains(arg, "/") {
			hasPackagePattern = true
			break
		}
	}
	if !hasPackagePattern {
		testArgs = append(testArgs, "./...")
	}

	cmd := exec.Command("go", testArgs...)
	cmd.Dir = b.project.GenDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		// Preserve the exit code from go test
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}
