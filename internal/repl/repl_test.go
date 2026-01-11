package repl

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Helper to create a test model with no project dependency
func newTestModel() Model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 80
	ti.Prompt = ""

	s := spinner.New()
	s.Spinner = spinner.Dot

	return Model{
		textInput:  ti,
		spinner:    s,
		historyIdx: -1,
	}
}

// =============================================================================
// Block Depth Tests
// =============================================================================

func TestCalculateBlockDepth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		// Simple cases
		{"empty", "", 0},
		{"simple expression", "1 + 2", 0},
		{"variable assignment", "x = 5", 0},

		// Function definitions
		{"def opens block", "def foo", 1},
		{"def closed", "def foo\nend", 0},
		{"nested def", "def foo\n  def bar", 2},
		{"nested def closed", "def foo\n  def bar\n  end\nend", 0},
		{"pub def opens block", "pub def foo", 1},

		// Class definitions
		{"class opens block", "class User", 1},
		{"class closed", "class User\nend", 0},
		{"pub class opens block", "pub class User", 1},

		// Interface definitions
		{"interface opens block", "interface Readable", 1},
		{"interface closed", "interface Readable\nend", 0},

		// Control structures
		{"if opens block", "if x > 0", 1},
		{"if closed", "if x > 0\n  puts(x)\nend", 0},
		{"while opens block", "while true", 1},
		{"while closed", "while true\n  break\nend", 0},
		{"for opens block", "for x in items", 1},
		{"for closed", "for x in items\n  puts(x)\nend", 0},

		// Blocks with do
		{"do opens block", "items.each do |x|", 1},
		{"do closed", "items.each do |x|\n  puts(x)\nend", 0},

		// Braces
		{"open brace", "items.map { |x|", 1},
		{"closed brace", "items.map { |x| x * 2 }", 0},

		// Complex nesting
		{"mixed nesting", "def foo\n  if bar\n    while baz", 3},
		{"partial close", "def foo\n  if bar\n  end", 1},
		{"class with method", "class User\n  def initialize", 2},
		{"class with method closed", "class User\n  def initialize(name : any)\n    @name = name\n  end\nend", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			got := m.calculateBlockDepth(tt.input)
			if got != tt.expected {
				t.Errorf("calculateBlockDepth(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Expression Detection Tests
// =============================================================================

func TestIsExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Expressions that should be wrapped in p()
		{"integer literal", "42", true},
		{"float literal", "3.14", true},
		{"string literal", `"hello"`, true},
		{"variable", "foo", true},
		{"arithmetic", "1 + 2", true},
		{"comparison", "x > 5", true},
		{"array literal", "[1, 2, 3]", true},
		{"method call returning value", "arr.length", true},
		{"chained method", "str.upcase.reverse", true},
		{"function call", "calculate(5)", true},

		// Void function calls - should NOT be wrapped
		{"puts call", "puts(x)", false},
		{"print call", "print(x)", false},
		{"p call", "p(x)", false},
		{"exit call", "exit(0)", false},
		{"sleep call", "sleep(1)", false},

		// Statements - not expressions
		{"assignment", "x = 5", false},
		{"if statement", "if x > 0", false},
		{"while statement", "while true", false},
		{"for statement", "for x in items", false},
		{"return statement", "return 5", false},

		// Edge cases
		{"puts with parens", "puts(x)", false},
		// each with block IS an expression (returns the array)
		{"each with block", "arr.each { |x| puts(x) }", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			got := m.isExpression(tt.input)
			if got != tt.expected {
				t.Errorf("isExpression(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Import Detection Tests
// =============================================================================

func TestIsImport(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"simple import", "import fmt", true},
		{"import with path", `import "net/http"`, true},
		{"import with alias", `import io "io/ioutil"`, true},
		{"import with leading space", "  import fmt", true},
		{"import with tab", "\timport fmt", true},

		// Not imports
		{"function call", "import_data()", false},
		{"variable", "importer", false},
		{"comment", "# import fmt", false},
		{"string containing import", `"import fmt"`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			got := m.isImport(tt.input)
			if got != tt.expected {
				t.Errorf("isImport(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Function Definition Detection Tests
// =============================================================================

func TestIsFunctionDef(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"simple def", "def foo", true},
		{"def with params", "def foo(x : any, y : any)", true},
		{"def with return type", "def foo -> Int", true},
		{"pub def", "pub def foo", true},
		{"pub def with params", "pub def foo(x : Int)", true},

		// With whitespace
		{"def with leading space", "  def foo", true},
		{"pub def with leading space", "  pub def foo", true},

		// Not function definitions
		{"define call", "define_method()", false},
		{"default variable", "default = 5", false},
		{"class def", "class Foo", false},
		{"pub class", "pub class Foo", false},
		{"comment", "# def foo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			got := m.isFunctionDef(tt.input)
			if got != tt.expected {
				t.Errorf("isFunctionDef(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Class Definition Detection Tests
// =============================================================================

func TestIsClassDef(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"simple class", "class User", true},
		{"class with parent", "class Admin < User", true},
		{"class with interface", "class User : Printable", true},
		{"pub class", "pub class User", true},

		// With whitespace
		{"class with leading space", "  class User", true},
		{"pub class with leading space", "  pub class User", true},

		// Not class definitions
		{"classify call", "classify(item)", false},
		{"class_name variable", `class_name = "User"`, false},
		{"def", "def foo", false},
		{"comment", "# class User", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			got := m.isClassDef(tt.input)
			if got != tt.expected {
				t.Errorf("isClassDef(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Interface Definition Detection Tests
// =============================================================================

func TestIsInterfaceDef(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"simple interface", "interface Readable", true},
		{"interface with methods", "interface Printable", true},
		{"pub interface", "pub interface Serializable", true},

		// With whitespace
		{"interface with leading space", "  interface Readable", true},
		{"pub interface with leading space", "  pub interface Serializable", true},

		// Not interface definitions
		{"interfacing call", "interfacing(item)", false},
		{"interface_name variable", `interface_name = "Reader"`, false},
		{"def", "def foo", false},
		{"class", "class User", false},
		{"comment", "# interface Readable", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			got := m.isInterfaceDef(tt.input)
			if got != tt.expected {
				t.Errorf("isInterfaceDef(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Assignment Detection Tests
// =============================================================================

func TestIsAssignment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"simple assignment", "x = 5", true},
		{"typed assignment", "x : Int = 5", true},
		{"string assignment", `name = "alice"`, true},
		{"array assignment", "items = [1, 2, 3]", true},
		{"expression assignment", "result = 1 + 2", true},
		{"method result assignment", "len = arr.length", true},

		// Not assignments
		{"comparison", "x == 5", false},
		{"expression", "1 + 2", false},
		{"function call", "foo()", false},
		{"puts", "puts(x)", false},
		{"or-assign", "x ||= 5", false},
		{"compound-assign", "x += 5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			got := m.isAssignment(tt.input)
			if got != tt.expected {
				t.Errorf("isAssignment(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Get Assignment Variable Tests
// =============================================================================

func TestGetAssignmentVar(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "x = 5", "x"},
		{"typed", "x : Int = 5", "x"},
		{"longer name", `user_name = "alice"`, "user_name"},
		{"expression rhs", "result = a + b", "result"},

		// Not assignments
		{"expression", "1 + 2", ""},
		{"comparison", "x == 5", ""},
		{"function call", "foo()", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			got := m.getAssignmentVar(tt.input)
			if got != tt.expected {
				t.Errorf("getAssignmentVar(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Build Program Tests
// =============================================================================

func TestBuildProgram(t *testing.T) {
	tests := []struct {
		name       string
		imports    []string
		functions  []string
		statements []string
		input      string
		expected   []string // substrings that must appear
		notExpect  []string // substrings that must NOT appear
	}{
		{
			name:     "simple expression",
			input:    "42",
			expected: []string{"p(42)"},
		},
		{
			name:     "expression with variable",
			input:    "x",
			expected: []string{"p(x)"},
		},
		{
			name:     "assignment prints value",
			input:    "x = 5",
			expected: []string{"x = 5", "p(x)"},
		},
		{
			name:      "void call not wrapped",
			input:     "puts(x)",
			expected:  []string{"puts(x)"},
			notExpect: []string{"p(puts"},
		},
		{
			name:     "with prior imports",
			imports:  []string{"import fmt"},
			input:    "42",
			expected: []string{"import fmt", "p(42)"},
		},
		{
			name:      "with function definitions",
			functions: []string{"def greet\n  puts \"hi\"\nend"},
			input:     "greet",
			expected:  []string{"def greet", "puts \"hi\"", "end", "p(greet)"},
		},
		{
			name:       "with prior statements",
			statements: []string{"x = 10"},
			input:      "x + 1",
			expected:   []string{"x = 10", "p(x + 1)"},
		},
		{
			name:       "full context",
			imports:    []string{"import fmt"},
			functions:  []string{"def double(n : any)\n  n * 2\nend"},
			statements: []string{"base = 5"},
			input:      "double(base)",
			expected:   []string{"import fmt", "def double", "base = 5", "p(double(base))"},
		},
		{
			name:      "bare script format",
			input:     "1 + 1",
			expected:  []string{"p(1 + 1)"},
			notExpect: []string{"def main"}, // should use bare script, not def main
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.imports = tt.imports
			m.functions = tt.functions
			m.statements = tt.statements

			got := m.buildProgram(tt.input)

			for _, want := range tt.expected {
				if !strings.Contains(got, want) {
					t.Errorf("buildProgram() missing %q in:\n%s", want, got)
				}
			}

			for _, notWant := range tt.notExpect {
				if strings.Contains(got, notWant) {
					t.Errorf("buildProgram() should not contain %q in:\n%s", notWant, got)
				}
			}
		})
	}
}

// =============================================================================
// State Accumulation Tests
// =============================================================================

func TestStateAccumulation(t *testing.T) {
	m := newTestModel()

	// Test that imports accumulate
	if m.isImport("import fmt") {
		m.imports = append(m.imports, "import fmt")
	}
	if m.isImport("import strings") {
		m.imports = append(m.imports, "import strings")
	}

	if len(m.imports) != 2 {
		t.Errorf("Expected 2 imports, got %d", len(m.imports))
	}

	// Test that functions accumulate
	if m.isFunctionDef("def foo\nend") {
		m.functions = append(m.functions, "def foo\nend")
	}
	if m.isFunctionDef("def bar\nend") {
		m.functions = append(m.functions, "def bar\nend")
	}

	if len(m.functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(m.functions))
	}
}

// =============================================================================
// History Navigation Tests (via Update method)
// =============================================================================

func TestHistoryNavigation(t *testing.T) {
	m := newTestModel()
	m.history = []string{"first", "second", "third"}
	m.historyIdx = -1

	// Press Up - should go to "third" (most recent)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m, _ = newModel.(Model)
	if m.historyIdx != 2 {
		t.Errorf("After first Up, historyIdx = %d, want 2", m.historyIdx)
	}
	if m.textInput.Value() != "third" {
		t.Errorf("After first Up, value = %q, want 'third'", m.textInput.Value())
	}

	// Press Up again - should go to "second"
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m, _ = newModel.(Model)
	if m.historyIdx != 1 {
		t.Errorf("After second Up, historyIdx = %d, want 1", m.historyIdx)
	}
	if m.textInput.Value() != "second" {
		t.Errorf("After second Up, value = %q, want 'second'", m.textInput.Value())
	}

	// Press Down - should go back to "third"
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = newModel.(Model)
	if m.historyIdx != 2 {
		t.Errorf("After Down, historyIdx = %d, want 2", m.historyIdx)
	}
	if m.textInput.Value() != "third" {
		t.Errorf("After Down, value = %q, want 'third'", m.textInput.Value())
	}

	// Press Down again - should clear input (past history)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = newModel.(Model)
	if m.historyIdx != -1 {
		t.Errorf("After going past history, historyIdx = %d, want -1", m.historyIdx)
	}
	if m.textInput.Value() != "" {
		t.Errorf("After going past history, value = %q, want empty", m.textInput.Value())
	}
}

func TestHistoryStaysAtBounds(t *testing.T) {
	m := newTestModel()
	m.history = []string{"only"}
	m.historyIdx = 0

	// Press Up at start of history - should stay at 0
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m, _ = newModel.(Model)
	if m.historyIdx != 0 {
		t.Errorf("historyIdx should stay at 0, got %d", m.historyIdx)
	}
}

func TestHistoryEmptyDoesNothing(t *testing.T) {
	m := newTestModel()
	m.history = []string{}
	m.historyIdx = -1

	// Press Up with empty history - should do nothing
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m, _ = newModel.(Model)
	if m.historyIdx != -1 {
		t.Errorf("historyIdx should stay at -1 with empty history, got %d", m.historyIdx)
	}
}

// =============================================================================
// Multi-line Input Tests
// =============================================================================

func TestMultiLineInputAccumulation(t *testing.T) {
	m := newTestModel()

	// Type "def foo" and press enter
	m.textInput.SetValue("def foo")
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = newModel.(Model)

	// Block should be open (depth 1)
	if m.blockDepth != 1 {
		t.Errorf("After 'def foo', blockDepth = %d, want 1", m.blockDepth)
	}
	if len(m.inputLines) != 1 {
		t.Errorf("After 'def foo', len(inputLines) = %d, want 1", len(m.inputLines))
	}

	// Type 'puts "hi"' and press enter
	m.textInput.SetValue(`  puts "hi"`)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = newModel.(Model)

	// Still open
	if m.blockDepth != 1 {
		t.Errorf("After 'puts', blockDepth = %d, want 1", m.blockDepth)
	}
	if len(m.inputLines) != 2 {
		t.Errorf("After 'puts', len(inputLines) = %d, want 2", len(m.inputLines))
	}

	// Note: We don't test closing blocks here because that triggers eval,
	// which requires a project. Block depth tracking is tested via calculateBlockDepth.
}

func TestNestedBlockDepth(t *testing.T) {
	m := newTestModel()

	// Open class
	m.textInput.SetValue("class User")
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = newModel.(Model)
	if m.blockDepth != 1 {
		t.Errorf("After 'class User', blockDepth = %d, want 1", m.blockDepth)
	}

	// Open method
	m.textInput.SetValue("  def initialize")
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = newModel.(Model)
	if m.blockDepth != 2 {
		t.Errorf("After 'def initialize', blockDepth = %d, want 2", m.blockDepth)
	}

	// Note: We don't test closing blocks here because that triggers eval.
}

// =============================================================================
// View Tests
// =============================================================================

func TestViewShowsPrompt(t *testing.T) {
	m := newTestModel()

	view := m.View()
	if !strings.Contains(view, ">>") {
		t.Error("View should contain >> prompt")
	}
}

func TestViewShowsContinuationPrompt(t *testing.T) {
	m := newTestModel()
	m.blockDepth = 1 // Simulate open block

	view := m.View()
	if !strings.Contains(view, "..") {
		t.Error("View should contain .. continuation prompt when block is open")
	}
}

func TestViewShowsGoodbye(t *testing.T) {
	m := newTestModel()
	m.quitting = true

	view := m.View()
	if !strings.Contains(view, "Goodbye") {
		t.Error("View should contain 'Goodbye' when quitting")
	}
}

func TestViewShowsSpinnerWhenEvaluating(t *testing.T) {
	m := newTestModel()
	m.evaluating = true

	view := m.View()
	// Should NOT show the prompt when evaluating
	if strings.Contains(view, ">>") {
		t.Error("View should not show prompt when evaluating")
	}
}

func TestViewShowsOutput(t *testing.T) {
	m := newTestModel()
	m.output = []string{"line one", "line two"}

	view := m.View()
	if !strings.Contains(view, "line one") {
		t.Error("View should contain output lines")
	}
	if !strings.Contains(view, "line two") {
		t.Error("View should contain output lines")
	}
}

// =============================================================================
// Quit Signal Tests
// =============================================================================

func TestCtrlCQuits(t *testing.T) {
	m := newTestModel()

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m, _ = newModel.(Model)

	if !m.quitting {
		t.Error("Ctrl+C should set quitting to true")
	}
	if cmd == nil {
		t.Error("Ctrl+C should return tea.Quit command")
	}
}

func TestCtrlDQuits(t *testing.T) {
	m := newTestModel()

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m, _ = newModel.(Model)

	if !m.quitting {
		t.Error("Ctrl+D should set quitting to true")
	}
	if cmd == nil {
		t.Error("Ctrl+D should return tea.Quit command")
	}
}

// =============================================================================
// Empty Input Tests
// =============================================================================

func TestEmptyInputDoesNothing(t *testing.T) {
	m := newTestModel()
	m.textInput.SetValue("")

	outputBefore := len(m.output)
	historyBefore := len(m.history)

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = newModel.(Model)

	if len(m.output) != outputBefore {
		t.Error("Empty input should not add to output")
	}
	if len(m.history) != historyBefore {
		t.Error("Empty input should not add to history")
	}
}

func TestWhitespaceOnlyDoesNothing(t *testing.T) {
	m := newTestModel()
	m.textInput.SetValue("   ")

	outputBefore := len(m.output)

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = newModel.(Model)

	if len(m.output) != outputBefore {
		t.Error("Whitespace-only input should not add to output")
	}
}

// =============================================================================
// Wrap for Parsing Tests
// =============================================================================

func TestWrapForParsing(t *testing.T) {
	m := newTestModel()

	tests := []struct {
		input    string
		expected string
	}{
		{"x = 5", "def replwrap\nx = 5\nend"},
		{"puts(x)", "def replwrap\nputs(x)\nend"},
		{"1 + 2", "def replwrap\n1 + 2\nend"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := m.wrapForParsing(tt.input)
			if got != tt.expected {
				t.Errorf("wrapForParsing(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Parse Statement Tests
// =============================================================================

func TestParseStatement(t *testing.T) {
	m := newTestModel()

	// Only test complete statements - incomplete ones cause parser to hang
	tests := []struct {
		input string
		isNil bool
	}{
		{"x = 5", false},
		{"puts(x)", false},
		{"1 + 2", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := m.parseStatement(tt.input)
			if (got == nil) != tt.isNil {
				t.Errorf("parseStatement(%q) nil=%v, want nil=%v", tt.input, got == nil, tt.isNil)
			}
		})
	}
}

// =============================================================================
// Output Truncation Tests
// =============================================================================

func TestTruncateOutput(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		expected int
	}{
		{"under limit", 50, 50},
		{"at limit", 100, 100},
		{"over limit", 150, 100},
		{"way over limit", 500, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.output = make([]string, tt.initial)
			for i := range m.output {
				m.output[i] = "line"
			}

			m.truncateOutput()

			if len(m.output) != tt.expected {
				t.Errorf("truncateOutput() with %d lines = %d lines, want %d",
					tt.initial, len(m.output), tt.expected)
			}
		})
	}
}

func TestTruncateOutputPreservesRecentLines(t *testing.T) {
	m := newTestModel()
	m.output = make([]string, 150)
	for i := range m.output {
		m.output[i] = fmt.Sprintf("line-%d", i)
	}

	m.truncateOutput()

	// Should keep the last 100 lines (50-149)
	if m.output[0] != "line-50" {
		t.Errorf("First line should be 'line-50', got %q", m.output[0])
	}
	if m.output[99] != "line-149" {
		t.Errorf("Last line should be 'line-149', got %q", m.output[99])
	}
}
