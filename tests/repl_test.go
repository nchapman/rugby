// Package tests provides the test framework for Rugby.
//
// REPL tests use a session-based .repl file format designed for testing
// interactive REPL sessions:
//
//	>>> 42
//	=> 42
//
//	>>> x = 5
//	=> 5
//
//	>>> def greet(name: String)
//	...   "Hello, #{name}!"
//	... end
//	=> (defined)
//
//	>>> puts "test"
//	test
//
//	>>> invalid_syntax(
//	#! parse error
//
// Format elements:
//   - >>> : Primary prompt (new input)
//   - ... : Continuation line (multi-line input)
//   - =>  : Expected auto-printed result
//   - Plain text : Expected stdout (without => prefix)
//   - #!  : Expected error (case-insensitive partial match)
//   - #@  : Directives (session name, description, skip)
//   - #   : Comments (ignored)
//   - Blank lines : Section separators
package tests

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nchapman/rugby/internal/builder"
	"github.com/nchapman/rugby/internal/repl"
)

// Exchange represents a single input/output exchange in a REPL test.
type Exchange struct {
	Input        string // The input to evaluate (may be multi-line)
	ExpectResult string // Expected result (from => line)
	ExpectStdout string // Expected stdout (plain text lines)
	ExpectError  string // Expected error substring (from #! line)
	LineNumber   int    // Line number in the .repl file for error reporting
}

// REPLTest represents a parsed .repl test file.
type REPLTest struct {
	Name        string
	Description string
	Skip        string
	Exchanges   []Exchange
}

// parseREPLTest parses a .repl file into a REPLTest structure.
func parseREPLTest(path string) (*REPLTest, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	test := &REPLTest{}
	scanner := bufio.NewScanner(file)
	lineNum := 0

	var currentExchange *Exchange
	var inputLines []string
	var outputLines []string
	expectingResult := false
	expectingError := false

	finishExchange := func() {
		if currentExchange != nil {
			currentExchange.Input = strings.Join(inputLines, "\n")
			if expectingResult {
				currentExchange.ExpectResult = strings.TrimSpace(strings.Join(outputLines, "\n"))
			} else if len(outputLines) > 0 {
				currentExchange.ExpectStdout = strings.Join(outputLines, "\n")
			}
			test.Exchanges = append(test.Exchanges, *currentExchange)
		}
		currentExchange = nil
		inputLines = nil
		outputLines = nil
		expectingResult = false
		expectingError = false
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Handle directives
		if strings.HasPrefix(line, "#@") {
			directive := strings.TrimSpace(line[2:])
			switch {
			case strings.HasPrefix(directive, "session:"):
				test.Name = strings.TrimSpace(directive[8:])
			case strings.HasPrefix(directive, "description:"):
				test.Description = strings.TrimSpace(directive[12:])
			case strings.HasPrefix(directive, "skip:"):
				test.Skip = strings.TrimSpace(directive[5:])
			}
			continue
		}

		// Handle comments (not error expectations)
		if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "#!") {
			continue
		}

		// Handle error expectations
		if strings.HasPrefix(line, "#!") {
			if currentExchange != nil {
				currentExchange.ExpectError = strings.TrimSpace(line[2:])
				expectingError = true
			}
			continue
		}

		// Handle primary prompt (new input)
		if strings.HasPrefix(line, ">>> ") || line == ">>>" {
			finishExchange()
			currentExchange = &Exchange{LineNumber: lineNum}
			input := strings.TrimPrefix(line, ">>> ")
			if input != line {
				inputLines = append(inputLines, input)
			}
			continue
		}

		// Handle continuation
		if strings.HasPrefix(line, "... ") || line == "..." {
			cont := strings.TrimPrefix(line, "... ")
			if cont == line {
				cont = ""
			}
			inputLines = append(inputLines, cont)
			continue
		}

		// Handle expected result
		if strings.HasPrefix(line, "=> ") {
			expectingResult = true
			result := strings.TrimPrefix(line, "=> ")
			outputLines = append(outputLines, result)
			continue
		}

		// Blank line finishes an exchange
		if strings.TrimSpace(line) == "" {
			finishExchange()
			continue
		}

		// Plain text is expected stdout (only if we have an active exchange)
		if currentExchange != nil && !expectingResult && !expectingError {
			outputLines = append(outputLines, line)
		}
	}

	// Finish any remaining exchange
	finishExchange()

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return test, nil
}

// TestREPLSessions runs all .repl test files.
func TestREPLSessions(t *testing.T) {
	replDir := "repl"
	if _, err := os.Stat(replDir); os.IsNotExist(err) {
		t.Skip("repl directory not found")
	}

	// Find all .repl files
	var tests []string
	err := filepath.Walk(replDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".repl") {
			tests = append(tests, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk repl directory: %v", err)
	}

	if len(tests) == 0 {
		t.Skip("no .repl test files found")
	}

	// Find project root by looking for go.mod
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}

	for _, testPath := range tests {
		testName := strings.TrimPrefix(testPath, "repl/")
		testName = strings.TrimSuffix(testName, ".repl")

		t.Run(testName, func(t *testing.T) {
			runREPLTest(t, testPath, projectRoot)
		})
	}
}

// runREPLTest executes a single .repl test file.
func runREPLTest(t *testing.T, testPath, projectRoot string) {
	test, err := parseREPLTest(testPath)
	if err != nil {
		t.Fatalf("failed to parse test file: %v", err)
	}

	if test.Skip != "" {
		t.Skipf("skipped: %s", test.Skip)
	}

	// Create evaluator with project using proper initialization
	project, err := builder.FindProjectFrom(projectRoot)
	if err != nil {
		t.Fatalf("failed to find project: %v", err)
	}
	evaluator := repl.NewEvaluatorWithProject(project)

	for i, exchange := range test.Exchanges {
		result := evaluator.Eval(exchange.Input)

		// Check for expected error
		if exchange.ExpectError != "" {
			if result.Error == nil {
				t.Errorf("exchange %d (line %d): expected error containing %q but got success with output: %s",
					i+1, exchange.LineNumber, exchange.ExpectError, result.Output)
				continue
			}
			errStr := result.Error.Error()
			if !strings.Contains(strings.ToLower(errStr), strings.ToLower(exchange.ExpectError)) {
				t.Errorf("exchange %d (line %d): expected error containing %q but got: %s",
					i+1, exchange.LineNumber, exchange.ExpectError, errStr)
			}
			continue
		}

		// Check for unexpected error
		if result.Error != nil {
			t.Errorf("exchange %d (line %d): unexpected error: %v\ninput: %s",
				i+1, exchange.LineNumber, result.Error, exchange.Input)
			continue
		}

		// Check expected result (from => lines)
		if exchange.ExpectResult != "" {
			actual := strings.TrimSpace(result.Output)
			expected := exchange.ExpectResult

			// Handle special cases
			if expected == "(defined)" {
				if !result.WasDefined {
					t.Errorf("exchange %d (line %d): expected (defined) but got: %s",
						i+1, exchange.LineNumber, actual)
				}
				continue
			}
			if expected == "(imported)" {
				if !result.WasImported {
					t.Errorf("exchange %d (line %d): expected (imported) but got: %s",
						i+1, exchange.LineNumber, actual)
				}
				continue
			}
			if expected == "(state cleared)" {
				if !result.WasCleared {
					t.Errorf("exchange %d (line %d): expected (state cleared) but got: %s",
						i+1, exchange.LineNumber, actual)
				}
				continue
			}

			if actual != expected {
				t.Errorf("exchange %d (line %d): result mismatch\n  input:    %s\n  expected: %s\n  actual:   %s",
					i+1, exchange.LineNumber, exchange.Input, expected, actual)
			}
			continue
		}

		// Check expected stdout (plain text lines)
		if exchange.ExpectStdout != "" {
			actual := normalizeOutput(result.Output)
			expected := normalizeOutput(exchange.ExpectStdout)
			if actual != expected {
				t.Errorf("exchange %d (line %d): stdout mismatch\n  input:    %s\n  expected:\n%s\n  actual:\n%s",
					i+1, exchange.LineNumber, exchange.Input, expected, actual)
			}
		}
	}
}

// normalizeOutput trims trailing whitespace from each line and trailing newlines.
func normalizeOutput(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

// findProjectRoot walks up from the current directory to find the project root
// (identified by go.mod). This makes tests work from any working directory.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
