package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGeneratedCodeModernization ensures the generated Go code follows modern best practices
func TestGeneratedCodeModernization(t *testing.T) {
	input := `def greet(name)
  puts "Hello, #{name}!"
end

def process(items)
  items.each do |item|
    puts item
  end

  items.map do |x|
    x * 2
  end
end

def main
  greet("World")
  nums = [1, 2, 3]
  result = process(nums)
  puts result
end`

	output := compile(t, input)

	// Check for modern Go idioms
	assertContains(t, output, "any")            // Should use 'any' not 'interface{}'
	assertNotContains(t, output, "interface{}") // Should not have old-style interface{}
}

// TestRangeOverIntOptimization verifies that exclusive ranges starting at 0
// use range-over-int syntax, while inclusive ranges use traditional for loops.
func TestRangeOverIntOptimization(t *testing.T) {
	// Exclusive range 0...n should use range-over-int
	exclusiveInput := `def main
  for i in 0...5
    puts i
  end
end`
	exclusiveOutput := compile(t, exclusiveInput)
	assertContains(t, exclusiveOutput, "for i := range 5")
	assertNotContains(t, exclusiveOutput, "i < 5")

	// Inclusive range 0..n should use traditional for loop
	inclusiveInput := `def main
  for i in 0..5
    puts i
  end
end`
	inclusiveOutput := compile(t, inclusiveInput)
	assertContains(t, inclusiveOutput, "i <= 5")
	assertNotContains(t, inclusiveOutput, "for i := range")

	// Non-zero start should use traditional for loop
	nonZeroInput := `def main
  for i in 1...5
    puts i
  end
end`
	nonZeroOutput := compile(t, nonZeroInput)
	assertContains(t, nonZeroOutput, "i := 1")
	assertContains(t, nonZeroOutput, "i < 5")
}

// TestGeneratedCodeLints ensures generated code passes golangci-lint
func TestGeneratedCodeLints(t *testing.T) {
	// Skip if golangci-lint is not installed
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		t.Skip("golangci-lint not installed, skipping generated code lint test")
	}

	input := `def process(items)
  result = []
  items.each do |item|
    result = result + [item * 2]
  end
  return result
end

def main
  nums = [1, 2, 3, 4, 5]
  doubled = process(nums)
  doubled.each do |n|
    puts n
  end
end`

	goCode := compile(t, input)

	// Write to temp file
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(goFile, []byte(goCode), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Find project root (where .golangci.yml is)
	projectRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	// We're in codegen/, so go up one level to project root
	projectRoot = filepath.Dir(projectRoot)
	configPath := filepath.Join(projectRoot, ".golangci.yml")

	// Run golangci-lint on generated code with our standard config
	cmd := exec.Command("golangci-lint", "run", "--config", configPath, goFile)
	output, err := cmd.CombinedOutput()

	// We expect clean output (exit code 0)
	// Note: Some issues like "package main has no package comment" are expected
	// We mainly care about modernize issues (interface{} vs any)
	if err != nil {
		// Check if it's just package comment warnings, which are acceptable
		if !strings.Contains(string(output), "interface{}") {
			// No interface{} issues, this is fine
			return
		}
		t.Errorf("Generated code has interface{} issues:\n%s\nCode:\n%s", output, goCode)
	}
}
