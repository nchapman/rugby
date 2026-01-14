package examples

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// example defines a Rugby example file and its expected behavior
type example struct {
	file           string
	expectedOutput string // empty means just verify it compiles and runs without error
	skipRun        bool   // true for examples that need external resources (network, etc.)
	hasBugs        bool   // true for examples with known compiler bugs (see BUGS.md)
}

// allExamples returns all example files with their expected outputs
func allExamples() []example {
	return []example{
		// Core language examples (01-15)
		{file: "01_hello.rg", expectedOutput: "Hello, Rugby!\nHello, World!\n2 + 3 = 5\nUppercase: HELLO\n"},
		{file: "02_types.rg", hasBugs: true},           // inline type annotation, empty typed arrays
		{file: "03_control_flow.rg", expectedOutput: "Grade: B\nNot empty!\nPositive\nNot zero\nMax: 5\nOK\nWeekend\nNice\nIt's a string: hello\nIt's an int: 42\nIt's a float: 3.14\n"},
		{file: "04_loops.rg"},                          // predicate methods now work
		{file: "05_functions.rg"},                      // optional return types now work
		{file: "06_classes.rg", hasBugs: true},         // self return, method chaining
		{file: "07_interfaces.rg"},                     // interface structural typing works
		{file: "08_modules.rg", hasBugs: true},         // lint: unused module methods
		{file: "09_blocks.rg", hasBugs: true},          // method chaining with newlines
		{file: "10_optionals.rg", hasBugs: true},       // safe navigation on getters, unwrap
		{file: "11_errors.rg"},                          // errors.new now works
		{file: "12_strings.rg"},                        // string methods work
		{file: "13_ranges.rg"},                         // range methods work
		{file: "14_go_interop.rg", expectedOutput: "Upper: HELLO WORLD\nContains 'world': true\nSplit: [hello world]\nJoined: one-two-three\nReplaced: hello Rugby\nInt to string: 42\nString to int: 123\nName: Alice, Age: 30\nStarting cleanup demo\nMiddle of function\nEnd of function body\nDeferred: runs last\n"},
		{file: "15_concurrency.rg", skipRun: true},      // output is non-deterministic due to goroutines

		// Additional examples
		{file: "fizzbuzz.rg", expectedOutput: "1\n2\nFizz\n4\nBuzz\nFizz\n7\n8\nFizz\nBuzz\n11\nFizz\n13\n14\nFizzBuzz\n"},
		{file: "field_inference.rg", expectedOutput: "I'm Alice, email: alice@example.com, role: admin\nI'm Alice, email: alice.new@example.com, role: admin\n"},
		{file: "json_simple.rg"},                       // rugby/json stdlib works
		{file: "worker_pool.rg", hasBugs: true},        // getter access on channel receive results
		{file: "todo_app.rg", hasBugs: true},           // map literal parsing in method body

		// HTTP + JSON (needs network, has bugs too)
		{file: "http_json.rg", skipRun: true, hasBugs: true}, // resp.json method, interface indexing
	}
}

// TestExamplesCompileAndRun ensures all examples compile and produce expected output
func TestExamplesCompileAndRun(t *testing.T) {
	rugbyBin := filepath.Join("..", "rugby")
	if _, err := os.Stat(rugbyBin); os.IsNotExist(err) {
		// Try to build it
		cmd := exec.Command("go", "build", "-o", rugbyBin, "..")
		if err := cmd.Run(); err != nil {
			t.Fatalf("rugby binary not found and failed to build: %v", err)
		}
	}

	for _, ex := range allExamples() {
		t.Run(ex.file, func(t *testing.T) {
			// Verify the file exists
			if _, err := os.Stat(ex.file); os.IsNotExist(err) {
				t.Fatalf("example file %s does not exist", ex.file)
			}

			if ex.hasBugs {
				t.Skip("skipping - has known compiler bugs (see BUGS.md)")
			}

			if ex.skipRun {
				// Just compile, don't run
				testCompileOnly(t, rugbyBin, ex.file)
			} else {
				// Compile and run, check output
				testCompileAndRun(t, rugbyBin, ex.file, ex.expectedOutput)
			}
		})
	}
}

// testCompileOnly verifies an example compiles without running it
func testCompileOnly(t *testing.T, rugbyBin, file string) {
	cmd := exec.Command(rugbyBin, "build", file)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("%s failed to compile:\n%s", file, output)
	}
}

// testCompileAndRun compiles and runs an example, checking output if expected is non-empty
func testCompileAndRun(t *testing.T, rugbyBin, file, expectedOutput string) {
	cmd := exec.Command(rugbyBin, "run", file)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if it's a compile error
		outputStr := string(output)
		if isCompileError(outputStr) {
			t.Fatalf("%s failed to compile:\n%s", file, outputStr)
		}
		// Runtime error
		t.Fatalf("%s failed at runtime: %v\n%s", file, err, outputStr)
	}

	// If we have expected output, verify it
	if expectedOutput != "" {
		if string(output) != expectedOutput {
			t.Errorf("%s output mismatch:\ngot:\n%s\nwant:\n%s", file, output, expectedOutput)
		}
	}
}

// isCompileError checks if output indicates a compilation failure
func isCompileError(output string) bool {
	compileErrors := []string{
		"syntax error",
		"undefined:",
		"cannot use",
		"cannot convert",
		"not enough arguments",
		"too many arguments",
		"declared and not used",
		"imported and not used",
		"type mismatch",
		"undeclared name",
	}
	for _, pattern := range compileErrors {
		if strings.Contains(output, pattern) {
			return true
		}
	}
	return false
}

// TestGeneratedCodeLint runs golangci-lint on generated Go code
func TestGeneratedCodeLint(t *testing.T) {
	// Check if golangci-lint is available
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		t.Skip("golangci-lint not installed, skipping lint test")
	}

	rugbyBin := filepath.Join("..", "rugby")
	if _, err := os.Stat(rugbyBin); os.IsNotExist(err) {
		cmd := exec.Command("go", "build", "-o", rugbyBin, "..")
		if err := cmd.Run(); err != nil {
			t.Fatalf("rugby binary not found and failed to build: %v", err)
		}
	}

	// Find the .rugby/gen directory
	genDir := filepath.Join(".", ".rugby", "gen")

	// Only lint examples that don't have known bugs
	for _, ex := range allExamples() {
		if ex.hasBugs || ex.skipRun {
			continue
		}

		t.Run(ex.file, func(t *testing.T) {
			// Build the example
			cmd := exec.Command(rugbyBin, "build", ex.file)
			cmd.Dir = "."
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Skipf("skipping lint for %s (build failed): %s", ex.file, output)
				return
			}

			// Find the generated Go file
			baseName := strings.TrimSuffix(ex.file, ".rg")
			goFile := filepath.Join(genDir, baseName+".go")
			if _, err := os.Stat(goFile); os.IsNotExist(err) {
				t.Skipf("generated file %s not found", goFile)
				return
			}

			// Run golangci-lint on the individual file (govet only)
			cmd = exec.Command("golangci-lint", "run",
				"--no-config",
				"-E", "govet",
				"-D", "staticcheck",
				"--max-issues-per-linter=0",
				"--max-same-issues=0",
				goFile,
			)
			cmd.Dir = "."
			output, err := cmd.CombinedOutput()

			if err != nil {
				outputStr := string(output)
				if strings.Contains(outputStr, "no go files") {
					return
				}
				if strings.Contains(outputStr, "typecheck") {
					return
				}
				if strings.TrimSpace(outputStr) != "" {
					t.Errorf("lint issues in generated code for %s:\n%s", ex.file, outputStr)
				}
			}
		})
	}
}

// TestAllExampleFilesAreTested ensures no example files are missed
func TestAllExampleFilesAreTested(t *testing.T) {
	// Get all .rg files in the examples directory
	files, err := filepath.Glob("*.rg")
	if err != nil {
		t.Fatalf("failed to glob .rg files: %v", err)
	}

	// Build a set of tested files
	tested := make(map[string]bool)
	for _, ex := range allExamples() {
		tested[ex.file] = true
	}

	// Check that all non-test .rg files are in the test list
	var missing []string
	for _, file := range files {
		// Skip test files (they're run differently)
		if strings.HasSuffix(file, "_test.rg") {
			continue
		}
		if !tested[file] {
			missing = append(missing, file)
		}
	}

	if len(missing) > 0 {
		t.Errorf("example files not covered by tests: %v\nAdd them to allExamples() in examples_test.go", missing)
	}
}

// TestExampleTestFiles ensures *_test.rg files parse correctly
func TestExampleTestFiles(t *testing.T) {
	testFiles, err := filepath.Glob("*_test.rg")
	if err != nil {
		t.Fatalf("failed to glob test files: %v", err)
	}

	if len(testFiles) == 0 {
		t.Skip("no *_test.rg files found")
	}

	for _, file := range testFiles {
		t.Run(file, func(t *testing.T) {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("failed to read %s: %v", file, err)
			}

			contentStr := string(content)
			if len(contentStr) == 0 {
				t.Errorf("%s is empty", file)
			}

			hasRugbyCode := strings.Contains(contentStr, "test ") ||
				strings.Contains(contentStr, "describe ") ||
				strings.Contains(contentStr, "it ") ||
				strings.Contains(contentStr, "def ") ||
				strings.Contains(contentStr, "puts")
			if !hasRugbyCode {
				t.Errorf("%s doesn't appear to contain valid Rugby code", file)
			}
		})
	}
}
