// Package tests provides the spec-driven test framework for Rugby.
//
// Spec tests are self-contained .rg files with directives that specify
// expected behavior. This approach tests language features end-to-end,
// catching integration bugs that unit tests miss.
//
// Directives (in comments at top of file):
//
//	#@ run-pass          - Must compile and run without error
//	#@ run-fail          - Must compile but fail at runtime
//	#@ compile-fail      - Must fail to compile
//	#@ check-output      - Verify stdout matches expected output
//	#@ skip: reason      - Skip test with reason
//
// Inline error expectations:
//
//	#~ ERROR: pattern    - Expected error matching regex pattern
//
// Expected output (for check-output tests):
//
//	#@ expect:
//	# line 1
//	# line 2
//
// Or use a .stdout golden file alongside the .rg file.
//
// Note: #@ expect: blocks end at the first non-comment line. For expected
// output containing blank lines, use a .stdout golden file instead.
package tests

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var bless = flag.Bool("bless", false, "Update golden files with actual output")

// TestMode specifies how a spec test should be run
type TestMode int

const (
	ModeRunPass TestMode = iota
	ModeRunFail
	ModeCompileFail
)

// Directives holds parsed test directives from a spec file
type Directives struct {
	Mode           TestMode
	CheckOutput    bool
	ExpectedOutput string
	ErrorPatterns  []*regexp.Regexp // compiled regex patterns from #~ ERROR: comments
	Skip           string           // skip reason, empty if not skipped
}

// parseDirectives reads a spec file and extracts test directives
func parseDirectives(path string) (*Directives, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	d := &Directives{
		Mode: ModeRunPass, // default
	}

	scanner := bufio.NewScanner(file)
	inExpect := false
	var expectLines []string
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for inline error expectations: #~ ERROR: pattern
		if idx := strings.Index(line, "#~"); idx >= 0 {
			rest := strings.TrimSpace(line[idx+2:])
			if strings.HasPrefix(rest, "ERROR:") {
				pattern := strings.TrimSpace(rest[6:])
				re, compileErr := regexp.Compile(pattern)
				if compileErr != nil {
					return nil, fmt.Errorf("invalid regex pattern %q on line %d: %w", pattern, lineNum, compileErr)
				}
				d.ErrorPatterns = append(d.ErrorPatterns, re)
			}
			continue
		}

		// Check for directives: #@ directive
		if !strings.HasPrefix(line, "#@") {
			// If we were in expect block, non-# line ends it
			if inExpect && !strings.HasPrefix(line, "#") {
				inExpect = false
			}
			// Collect expect lines (lines starting with # after #@ expect:)
			if inExpect && strings.HasPrefix(line, "#") {
				// Strip leading "# " or "#"
				expectLine := strings.TrimPrefix(line, "# ")
				if expectLine == line {
					expectLine = strings.TrimPrefix(line, "#")
				}
				expectLines = append(expectLines, expectLine)
			}
			continue
		}

		directive := strings.TrimSpace(line[2:])

		switch {
		case directive == "run-pass":
			d.Mode = ModeRunPass
		case directive == "run-fail":
			d.Mode = ModeRunFail
		case directive == "compile-fail":
			d.Mode = ModeCompileFail
		case directive == "check-output":
			d.CheckOutput = true
		case strings.HasPrefix(directive, "skip:"):
			d.Skip = strings.TrimSpace(directive[5:])
		case directive == "expect:":
			inExpect = true
		case directive == "" || strings.HasPrefix(directive, " "):
			// Empty directive or continuation, ignore
		default:
			return nil, fmt.Errorf("unknown directive %q on line %d", directive, lineNum)
		}
	}

	if len(expectLines) > 0 {
		d.ExpectedOutput = strings.Join(expectLines, "\n") + "\n"
	}

	return d, scanner.Err()
}

// loadGoldenFile attempts to load a .stdout golden file for a spec
func loadGoldenFile(specPath string) (string, bool) {
	goldenPath := strings.TrimSuffix(specPath, ".rg") + ".stdout"
	content, err := os.ReadFile(goldenPath)
	if err != nil {
		return "", false
	}
	return string(content), true
}

// saveGoldenFile writes a .stdout golden file for a spec
func saveGoldenFile(specPath, content string) error {
	goldenPath := strings.TrimSuffix(specPath, ".rg") + ".stdout"
	return os.WriteFile(goldenPath, []byte(content), 0644)
}

// TestSpecs runs all spec tests in the spec/ directory
func TestSpecs(t *testing.T) {
	// Find the rugby binary
	rugbyBin, err := findRugbyBinary()
	if err != nil {
		t.Fatalf("failed to find rugby binary: %v", err)
	}

	// Find all spec files
	specDir := "spec"
	if _, statErr := os.Stat(specDir); os.IsNotExist(statErr) {
		t.Skip("spec directory not found")
	}

	var specs []string
	err = filepath.Walk(specDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() && strings.HasSuffix(path, ".rg") {
			specs = append(specs, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk spec directory: %v", err)
	}

	if len(specs) == 0 {
		t.Skip("no spec files found")
	}

	for _, spec := range specs {
		// Use relative path for test name
		testName := strings.TrimPrefix(spec, "spec/")
		testName = strings.TrimSuffix(testName, ".rg")

		t.Run(testName, func(t *testing.T) {
			runSpec(t, rugbyBin, spec)
		})
	}
}

// runSpec executes a single spec test
func runSpec(t *testing.T, rugbyBin, specPath string) {
	directives, err := parseDirectives(specPath)
	if err != nil {
		t.Fatalf("failed to parse directives: %v", err)
	}

	// Handle skip
	if directives.Skip != "" {
		t.Skipf("skipped: %s", directives.Skip)
	}

	switch directives.Mode {
	case ModeRunPass:
		runPassTest(t, rugbyBin, specPath, directives)
	case ModeRunFail:
		runFailTest(t, rugbyBin, specPath, directives)
	case ModeCompileFail:
		compileFailTest(t, rugbyBin, specPath, directives)
	}
}

// runPassTest verifies code compiles and runs successfully
func runPassTest(t *testing.T, rugbyBin, specPath string, d *Directives) {
	cmd := exec.Command(rugbyBin, "run", specPath)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		t.Fatalf("expected to pass but failed:\n%s", outputStr)
	}

	// Check output if requested
	if d.CheckOutput {
		expected := d.ExpectedOutput

		// Try golden file if no inline expect
		if expected == "" {
			var ok bool
			expected, ok = loadGoldenFile(specPath)
			if !ok && *bless {
				// Create golden file with actual output
				if saveErr := saveGoldenFile(specPath, outputStr); saveErr != nil {
					t.Errorf("failed to save golden file: %v", saveErr)
				}
				t.Logf("created golden file for %s", specPath)
				return
			}
			if !ok {
				t.Fatal("check-output specified but no expected output (use #@ expect: or create .stdout file, or run with -bless)")
			}
		}

		if outputStr != expected {
			if *bless {
				// Update golden file
				if saveErr := saveGoldenFile(specPath, outputStr); saveErr != nil {
					t.Errorf("failed to update golden file: %v", saveErr)
				}
				t.Logf("updated golden file for %s", specPath)
				return
			}
			t.Errorf("output mismatch:\n--- expected ---\n%s\n--- actual ---\n%s", expected, outputStr)
		}
	}
}

// runFailTest verifies code compiles but fails at runtime
func runFailTest(t *testing.T, rugbyBin, specPath string, d *Directives) {
	// First, verify it compiles successfully
	buildCmd := exec.Command(rugbyBin, "build", specPath)
	buildOutput, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		t.Fatalf("expected to compile but failed:\n%s", string(buildOutput))
	}

	// Now run it and expect failure
	runCmd := exec.Command(rugbyBin, "run", specPath)
	output, err := runCmd.CombinedOutput()
	outputStr := string(output)

	if err == nil {
		t.Fatalf("expected runtime failure but succeeded:\n%s", outputStr)
	}

	// Check expected error patterns
	for _, re := range d.ErrorPatterns {
		if !re.MatchString(outputStr) {
			t.Errorf("expected error matching %q but got:\n%s", re.String(), outputStr)
		}
	}
}

// compileFailTest verifies code fails to compile with expected errors
func compileFailTest(t *testing.T, rugbyBin, specPath string, d *Directives) {
	cmd := exec.Command(rugbyBin, "build", specPath)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err == nil {
		t.Fatalf("expected compile failure but succeeded")
	}

	// Verify expected error patterns
	if len(d.ErrorPatterns) == 0 {
		// No specific patterns required, just needs to fail
		return
	}

	for _, re := range d.ErrorPatterns {
		if !re.MatchString(outputStr) {
			t.Errorf("expected error matching %q but got:\n%s", re.String(), outputStr)
		}
	}
}

// findRugbyBinary locates or builds the rugby compiler
func findRugbyBinary() (string, error) {
	// Try relative path from tests directory
	candidates := []string{
		"../rugby",
		"rugby",
		"./rugby",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			abs, absErr := filepath.Abs(path)
			if absErr != nil {
				return "", fmt.Errorf("failed to get absolute path for %s: %w", path, absErr)
			}
			return abs, nil
		}
	}

	// Try to build it
	cmd := exec.Command("go", "build", "-o", "../rugby", "..")
	if err := cmd.Run(); err != nil {
		return "", err
	}

	abs, absErr := filepath.Abs("../rugby")
	if absErr != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", absErr)
	}
	return abs, nil
}
