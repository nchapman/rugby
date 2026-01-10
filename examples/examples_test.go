package examples

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestExamplesCompile ensures all .rg examples compile successfully
func TestExamplesCompile(t *testing.T) {
	examples := []struct {
		file           string
		expectsRuntime bool
	}{
		{"hello.rg", true},
		{"arith.rg", false},
		{"simple.rg", false},
		{"tiny.rg", false},
		{"fizzbuzz.rg", false},
		{"ranges.rg", true},
		{"symbols.rg", true},
		{"case.rg", true},
	}

	// Get the rugby binary path
	rugbyBin := filepath.Join("..", "rugby")
	if _, err := os.Stat(rugbyBin); os.IsNotExist(err) {
		t.Skip("rugby binary not built, run 'go build' first")
	}

	for _, ex := range examples {
		t.Run(ex.file, func(t *testing.T) {
			// Compile the example
			cmd := exec.Command(rugbyBin, "run", ex.file)
			cmd.Dir = "."
			output, err := cmd.CombinedOutput()

			// We don't care if it runs successfully, just that it compiles
			// Some examples might need runtime setup or produce errors
			if err != nil {
				// Check if it's a compile error vs runtime error
				outputStr := string(output)
				if strings.Contains(outputStr, "syntax error") ||
					strings.Contains(outputStr, "undefined:") ||
					strings.Contains(outputStr, "cannot use") {
					t.Errorf("%s failed to compile:\n%s", ex.file, outputStr)
				}
				// Runtime errors are OK for this test
			}
		})
	}
}

// TestHelloExample specifically tests the hello.rg output
func TestHelloExample(t *testing.T) {
	rugbyBin := filepath.Join("..", "rugby")
	if _, err := os.Stat(rugbyBin); os.IsNotExist(err) {
		t.Skip("rugby binary not built, run 'go build' first")
	}

	cmd := exec.Command(rugbyBin, "run", "hello.rg")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("hello.rg failed: %v\n%s", err, output)
	}

	expected := "hi\n"
	if string(output) != expected {
		t.Errorf("hello.rg output = %q, want %q", output, expected)
	}
}
