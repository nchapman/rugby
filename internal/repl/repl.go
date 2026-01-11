// Package repl implements the interactive Rugby REPL.
package repl

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"rugby/ast"
	"rugby/codegen"
	"rugby/internal/builder"
	"rugby/lexer"
	"rugby/parser"
	"rugby/token"
)

// Styles
var (
	promptStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	contStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	outputStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	mutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// evalResult is the message sent when async evaluation completes.
type evalResult struct {
	output string
	err    error
	input  string // original input for tracking assignments
}

// Model is the bubbletea model for the REPL.
type Model struct {
	textInput  textinput.Model
	spinner    spinner.Model
	history    []string
	historyIdx int
	output     []string

	// Accumulated Rugby state
	imports    []string
	functions  []string
	statements []string

	// Multi-line input
	inputLines []string
	blockDepth int

	// Async evaluation state
	evaluating   bool
	pendingInput string // input being evaluated

	project  *builder.Project
	counter  int // for unique temp file names
	quitting bool
}

// New creates a new REPL model.
func New() Model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 80
	ti.Prompt = ""

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	project, _ := builder.FindProject()

	return Model{
		textInput:  ti,
		spinner:    s,
		historyIdx: -1,
		project:    project,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case evalResult:
		// Evaluation complete
		m.evaluating = false
		if msg.err != nil {
			m.output = append(m.output, errorStyle.Render(msg.err.Error()))
		} else if msg.output != "" {
			for line := range strings.SplitSeq(strings.TrimRight(msg.output, "\n"), "\n") {
				m.output = append(m.output, outputStyle.Render(line))
			}
		}

		// Track assignment for variable persistence
		if m.isAssignment(msg.input) {
			m.statements = append(m.statements, msg.input)
		}

		m.truncateOutput()
		return m, nil

	case spinner.TickMsg:
		if m.evaluating {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		// Ignore input while evaluating
		if m.evaluating {
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyUp:
			if len(m.history) > 0 {
				if m.historyIdx < 0 {
					m.historyIdx = len(m.history) - 1
				} else if m.historyIdx > 0 {
					m.historyIdx--
				}
				m.textInput.SetValue(m.history[m.historyIdx])
				m.textInput.CursorEnd()
			}
			return m, nil

		case tea.KeyDown:
			if m.historyIdx >= 0 {
				m.historyIdx++
				if m.historyIdx >= len(m.history) {
					m.historyIdx = -1
					m.textInput.SetValue("")
				} else {
					m.textInput.SetValue(m.history[m.historyIdx])
					m.textInput.CursorEnd()
				}
			}
			return m, nil

		case tea.KeyEnter:
			line := m.textInput.Value()
			m.textInput.SetValue("")
			m.historyIdx = -1

			// Handle empty input
			if strings.TrimSpace(line) == "" && len(m.inputLines) == 0 {
				return m, nil
			}

			// Accumulate multi-line input
			m.inputLines = append(m.inputLines, line)
			fullInput := strings.Join(m.inputLines, "\n")

			// Check if input is complete
			m.blockDepth = m.calculateBlockDepth(fullInput)
			if m.blockDepth > 0 {
				// Need more input
				return m, nil
			}

			// Input is complete - evaluate
			m.inputLines = nil
			m.blockDepth = 0

			// Add to history
			if strings.TrimSpace(fullInput) != "" {
				m.history = append(m.history, fullInput)
			}

			// Show the input in output
			lines := strings.Split(fullInput, "\n")
			for i, l := range lines {
				if i == 0 {
					m.output = append(m.output, promptStyle.Render(">> ")+l)
				} else {
					m.output = append(m.output, contStyle.Render(".. ")+l)
				}
			}

			// Start async evaluation with spinner
			m.evaluating = true
			m.pendingInput = fullInput
			return m, tea.Batch(m.spinner.Tick, m.runEval(fullInput))
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// runEval returns a command that runs evaluation asynchronously.
func (m *Model) runEval(input string) tea.Cmd {
	return func() tea.Msg {
		result, err := m.eval(input)
		return evalResult{output: result, err: err, input: input}
	}
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return mutedStyle.Render("Goodbye!") + "\n"
	}

	var s strings.Builder

	// Show output history
	for _, line := range m.output {
		s.WriteString(line + "\n")
	}

	// Show spinner or prompt
	if m.evaluating {
		s.WriteString(m.spinner.View())
	} else if m.blockDepth > 0 {
		s.WriteString(contStyle.Render(".. "))
		s.WriteString(m.textInput.View())
	} else {
		s.WriteString(promptStyle.Render(">> "))
		s.WriteString(m.textInput.View())
	}

	return s.String()
}

// calculateBlockDepth counts unclosed blocks.
func (m *Model) calculateBlockDepth(input string) int {
	depth := 0
	l := lexer.New(input)
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
		switch tok.Type {
		case token.DEF, token.CLASS, token.INTERFACE, token.IF, token.WHILE, token.FOR, token.DO:
			depth++
		case token.END:
			depth--
		case token.LBRACE:
			depth++
		case token.RBRACE:
			depth--
		}
	}
	return depth
}

// wrapForParsing wraps input in a function so the parser can handle statements.
func (m *Model) wrapForParsing(input string) string {
	return "def replwrap\n" + input + "\nend"
}

// parseStatement parses the input as a statement inside a function.
func (m *Model) parseStatement(input string) ast.Statement {
	wrapped := m.wrapForParsing(input)
	l := lexer.New(wrapped)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 || len(program.Declarations) != 1 {
		return nil
	}

	funcDecl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok || len(funcDecl.Body) != 1 {
		return nil
	}

	return funcDecl.Body[0]
}

// isExpression checks if the input is a standalone expression that returns a value.
func (m *Model) isExpression(input string) bool {
	stmt := m.parseStatement(input)
	if stmt == nil {
		return false
	}
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}
	// Don't wrap void function calls (puts, print, p, exit, sleep)
	if call, ok := exprStmt.Expr.(*ast.CallExpr); ok {
		if ident, ok := call.Func.(*ast.Ident); ok {
			switch ident.Name {
			case "puts", "print", "p", "exit", "sleep":
				return false
			}
		}
	}
	return true
}

// isImport checks if the input is an import statement.
func (m *Model) isImport(input string) bool {
	return strings.HasPrefix(strings.TrimSpace(input), "import ")
}

// isFunctionDef checks if the input is a function definition.
func (m *Model) isFunctionDef(input string) bool {
	trimmed := strings.TrimSpace(input)
	return strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "pub def ")
}

// isClassDef checks if the input is a class definition.
func (m *Model) isClassDef(input string) bool {
	trimmed := strings.TrimSpace(input)
	return strings.HasPrefix(trimmed, "class ") || strings.HasPrefix(trimmed, "pub class ")
}

// isInterfaceDef checks if the input is an interface definition.
func (m *Model) isInterfaceDef(input string) bool {
	trimmed := strings.TrimSpace(input)
	return strings.HasPrefix(trimmed, "interface ") || strings.HasPrefix(trimmed, "pub interface ")
}

// truncateOutput keeps output to a maximum of 100 lines.
func (m *Model) truncateOutput() {
	if len(m.output) > 100 {
		m.output = m.output[len(m.output)-100:]
	}
}

// eval compiles and runs the input, returning the output.
func (m *Model) eval(input string) (string, error) {
	// Handle special inputs
	trimmed := strings.TrimSpace(input)
	if trimmed == "exit" || trimmed == "quit" {
		return "", fmt.Errorf("use Ctrl+D to exit")
	}

	// Accumulate state based on input type
	if m.isImport(input) {
		m.imports = append(m.imports, input)
		return successStyle.Render("(imported)"), nil
	}

	if m.isFunctionDef(input) {
		m.functions = append(m.functions, input)
		return successStyle.Render("(defined)"), nil
	}

	if m.isClassDef(input) {
		m.functions = append(m.functions, input)
		return successStyle.Render("(defined)"), nil
	}

	if m.isInterfaceDef(input) {
		m.functions = append(m.functions, input)
		return successStyle.Render("(defined)"), nil
	}

	// Check for valid project
	if m.project == nil {
		return "", fmt.Errorf("no Rugby project found - run from a project directory")
	}

	// Build the program
	program := m.buildProgram(input)

	// Ensure project dirs exist
	if err := m.project.EnsureDirs(); err != nil {
		return "", err
	}

	// Write to temp file
	replDir := filepath.Join(m.project.RugbyDir, "repl")
	if err := os.MkdirAll(replDir, 0755); err != nil {
		return "", err
	}

	m.counter++
	srcFile := filepath.Join(replDir, fmt.Sprintf("repl_%d.rg", m.counter))
	if err := os.WriteFile(srcFile, []byte(program), 0644); err != nil {
		return "", err
	}

	// Parse and generate
	l := lexer.New(program)
	p := parser.New(l)
	ast := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return "", fmt.Errorf("parse error: %s", p.Errors()[0])
	}

	gen := codegen.New()
	goCode, err := gen.Generate(ast)
	if err != nil {
		return "", fmt.Errorf("codegen error: %v", err)
	}

	// Write Go file
	goFile := filepath.Join(replDir, fmt.Sprintf("repl_%d.go", m.counter))
	err = os.WriteFile(goFile, []byte(goCode), 0644)
	if err != nil {
		return "", err
	}

	// Create go.mod for the repl directory
	goModContent := fmt.Sprintf(`module repl

go 1.25

require rugby v0.0.0

replace rugby => %s
`, m.project.Root)

	goModFile := filepath.Join(replDir, "go.mod")
	err = os.WriteFile(goModFile, []byte(goModContent), 0644)
	if err != nil {
		return "", err
	}

	// Run go mod tidy (ignore errors - it may fail if no dependencies needed)
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = replDir
	_ = tidyCmd.Run()

	// Compile
	binFile := filepath.Join(replDir, fmt.Sprintf("repl_%d", m.counter))
	buildCmd := exec.Command("go", "build", "-o", binFile, goFile)
	buildCmd.Dir = replDir

	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build error: %s", strings.TrimSpace(string(buildOut)))
	}

	// Run
	runCmd := exec.Command(binFile)
	var stdout, stderr bytes.Buffer
	runCmd.Stdout = &stdout
	runCmd.Stderr = &stderr

	err = runCmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += stderr.String()
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 0 && output == "" {
				return "", fmt.Errorf("runtime error (exit %d)", exitErr.ExitCode())
			}
		}
	}

	return output, nil
}

// isAssignment checks if the input is a variable assignment.
func (m *Model) isAssignment(input string) bool {
	stmt := m.parseStatement(input)
	if stmt == nil {
		return false
	}
	_, ok := stmt.(*ast.AssignStmt)
	return ok
}

// getAssignmentVar returns the variable name if input is an assignment.
func (m *Model) getAssignmentVar(input string) string {
	stmt := m.parseStatement(input)
	if stmt == nil {
		return ""
	}
	if assign, ok := stmt.(*ast.AssignStmt); ok {
		return assign.Name
	}
	return ""
}

// buildProgram constructs a full Rugby program using bare script support.
// Top-level statements are written directly and codegen wraps them in main().
func (m *Model) buildProgram(input string) string {
	var buf strings.Builder

	// Imports
	for _, imp := range m.imports {
		buf.WriteString(imp + "\n")
	}

	// Function/class definitions
	for _, fn := range m.functions {
		buf.WriteString(fn + "\n")
	}

	// Prior statements (for variable persistence)
	for _, stmt := range m.statements {
		buf.WriteString(stmt + "\n")
	}

	// New input - wrap expressions in p() to print
	if m.isExpression(input) {
		buf.WriteString("p(" + input + ")\n")
	} else if varName := m.getAssignmentVar(input); varName != "" {
		// Assignment - print the value after assigning
		buf.WriteString(input + "\n")
		buf.WriteString("p(" + varName + ")\n")
	} else {
		// Output as-is (multi-line input)
		for line := range strings.SplitSeq(input, "\n") {
			buf.WriteString(line + "\n")
		}
	}

	return buf.String()
}

// Run starts the REPL.
func Run() error {
	fmt.Println(mutedStyle.Render("Rugby REPL v0.1") + " " + mutedStyle.Render("(Ctrl+D to quit)"))
	fmt.Println()

	p := tea.NewProgram(New())
	_, err := p.Run()
	return err
}
