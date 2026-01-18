// Package repl implements the interactive Rugby REPL.
package repl

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/codegen"
	"github.com/nchapman/rugby/internal/builder"
	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/parser"
	"github.com/nchapman/rugby/semantic"
	"github.com/nchapman/rugby/token"
)

// typeInfoAdapter adapts semantic.Analyzer to implement codegen.TypeInfo.
type typeInfoAdapter struct {
	analyzer *semantic.Analyzer
}

func (t *typeInfoAdapter) GetTypeKind(node ast.Node) codegen.TypeKind {
	typ := t.analyzer.GetType(node)
	if typ == nil {
		return codegen.TypeUnknown
	}
	switch typ.Kind {
	case semantic.TypeInt:
		return codegen.TypeInt
	case semantic.TypeInt64:
		return codegen.TypeInt64
	case semantic.TypeFloat:
		return codegen.TypeFloat
	case semantic.TypeBool:
		return codegen.TypeBool
	case semantic.TypeString:
		return codegen.TypeString
	case semantic.TypeNil:
		return codegen.TypeNil
	case semantic.TypeArray:
		return codegen.TypeArray
	case semantic.TypeMap:
		return codegen.TypeMap
	case semantic.TypeClass:
		return codegen.TypeClass
	case semantic.TypeOptional:
		return codegen.TypeOptional
	case semantic.TypeChan:
		return codegen.TypeChannel
	case semantic.TypeAny:
		return codegen.TypeAny
	default:
		return codegen.TypeUnknown
	}
}

func (t *typeInfoAdapter) GetGoType(node ast.Node) string {
	typ := t.analyzer.GetType(node)
	if typ == nil {
		return ""
	}
	return typ.GoType()
}

func (t *typeInfoAdapter) GetRugbyType(node ast.Node) string {
	typ := t.analyzer.GetType(node)
	if typ == nil {
		return ""
	}
	return typ.String()
}

func (t *typeInfoAdapter) GetSelectorKind(node ast.Node) ast.SelectorKind {
	if sel, ok := node.(*ast.SelectorExpr); ok && sel.ResolvedKind != ast.SelectorUnknown {
		return sel.ResolvedKind
	}
	return t.analyzer.GetSelectorKind(node)
}

func (t *typeInfoAdapter) GetElementType(node ast.Node) string {
	typ := t.analyzer.GetType(node)
	if typ == nil || typ.Elem == nil {
		return ""
	}
	return typ.Elem.String()
}

func (t *typeInfoAdapter) GetKeyValueTypes(node ast.Node) (string, string) {
	typ := t.analyzer.GetType(node)
	if typ == nil || typ.Kind != semantic.TypeMap {
		return "", ""
	}
	keyType := ""
	valueType := ""
	if typ.KeyType != nil {
		keyType = typ.KeyType.String()
	}
	if typ.ValueType != nil {
		valueType = typ.ValueType.String()
	}
	return keyType, valueType
}

func (t *typeInfoAdapter) GetTupleTypes(node ast.Node) []string {
	typ := t.analyzer.GetType(node)
	if typ == nil {
		return nil
	}
	if typ.Kind == semantic.TypeTuple && len(typ.Elements) > 0 {
		result := make([]string, len(typ.Elements))
		for i, elem := range typ.Elements {
			result[i] = elem.String()
		}
		return result
	}
	if typ.Kind == semantic.TypeOptional && typ.Elem != nil {
		return []string{typ.Elem.String(), "Bool"}
	}
	return nil
}

func (t *typeInfoAdapter) IsVariableUsedAt(node ast.Node, name string) bool {
	return t.analyzer.IsVariableUsedAt(node, name)
}

func (t *typeInfoAdapter) IsDeclaration(node ast.Node) bool {
	return t.analyzer.IsDeclaration(node)
}

func (t *typeInfoAdapter) GetFieldType(className, fieldName string) string {
	return t.analyzer.GetFieldType(className, fieldName)
}

func (t *typeInfoAdapter) IsClass(typeName string) bool {
	return t.analyzer.IsClass(typeName)
}

func (t *typeInfoAdapter) IsInterface(typeName string) bool {
	return t.analyzer.IsInterface(typeName)
}

func (t *typeInfoAdapter) IsStruct(typeName string) bool {
	return t.analyzer.IsStruct(typeName)
}

func (t *typeInfoAdapter) IsNoArgFunction(name string) bool {
	return t.analyzer.IsNoArgFunction(name)
}

func (t *typeInfoAdapter) IsPublicClass(className string) bool {
	return t.analyzer.IsPublicClass(className)
}

func (t *typeInfoAdapter) HasAccessor(className, fieldName string) bool {
	return t.analyzer.HasAccessor(className, fieldName)
}

func (t *typeInfoAdapter) GetInterfaceMethodNames(interfaceName string) []string {
	return t.analyzer.GetInterfaceMethodNames(interfaceName)
}

func (t *typeInfoAdapter) GetAllInterfaceNames() []string {
	return t.analyzer.GetAllInterfaceNames()
}

func (t *typeInfoAdapter) GetAllModuleNames() []string {
	return t.analyzer.GetAllModuleNames()
}

func (t *typeInfoAdapter) GetModuleMethodNames(moduleName string) []string {
	return t.analyzer.GetModuleMethodNames(moduleName)
}

func (t *typeInfoAdapter) IsModuleMethod(className, methodName string) bool {
	return t.analyzer.IsModuleMethod(className, methodName)
}

func (t *typeInfoAdapter) GetConstructorParamCount(className string) int {
	return t.analyzer.GetConstructorParamCount(className)
}

func (t *typeInfoAdapter) GetConstructorParams(className string) [][2]string {
	return t.analyzer.GetConstructorParams(className)
}

func (t *typeInfoAdapter) GetGoName(node ast.Node) string {
	return t.analyzer.GetGoName(node)
}

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

	// State changes to apply (avoids data races by not mutating during async eval)
	addImport   string // import to add
	addFunction string // function/class/interface to add
	clearState  bool   // reset all state
	clearOutput bool   // clear output (for reset)
	quit        bool   // exit the REPL
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

		// Handle quit request
		if msg.quit {
			m.quitting = true
			return m, tea.Quit
		}

		// Apply state changes from eval (done here to avoid data races)
		if msg.clearState {
			m.imports = nil
			m.functions = nil
			m.statements = nil
		}
		if msg.clearOutput {
			m.output = nil
		}
		if msg.addImport != "" {
			m.imports = append(m.imports, msg.addImport)
		}
		if msg.addFunction != "" {
			m.functions = append(m.functions, msg.addFunction)
		}

		// Handle output
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
			m.counter++ // Increment before spawning to avoid race
			return m, tea.Batch(m.spinner.Tick, m.runEval(fullInput, m.counter))
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// runEval returns a command that runs evaluation asynchronously.
// Counter is passed in to avoid race conditions on m.counter.
func (m *Model) runEval(input string, counter int) tea.Cmd {
	// Capture read-only state needed for eval to avoid races
	ctx := evalContext{
		imports:    append([]string{}, m.imports...),
		functions:  append([]string{}, m.functions...),
		statements: append([]string{}, m.statements...),
		project:    m.project,
		counter:    counter,
	}
	return func() tea.Msg {
		return m.eval(input, ctx)
	}
}

// evalContext holds read-only state captured before async evaluation.
type evalContext struct {
	imports    []string
	functions  []string
	statements []string
	project    *builder.Project
	counter    int
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
	// Don't wrap expressions containing blocks - p(expr { block }) is invalid syntax
	if hasBlock(exprStmt.Expr) {
		return false
	}
	return true
}

// isBlockExpression checks if the input is a value-returning block expression.
// Returns false for void block methods (each, times, upto, downto).
func (m *Model) isBlockExpression(input string) bool {
	stmt := m.parseStatement(input)
	if stmt == nil {
		return false
	}
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}
	if !hasBlock(exprStmt.Expr) {
		return false
	}
	// Check if it's a void block method
	if isVoidBlockMethod(exprStmt.Expr) {
		return false
	}
	return true
}

// isVoidBlockMethod checks if the expression is a call to a void block method.
func isVoidBlockMethod(expr ast.Expression) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Func.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel {
	case "each", "times", "upto", "downto":
		return true
	}
	return false
}

// hasBlock recursively checks if an expression contains a block.
func hasBlock(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.CallExpr:
		if e.Block != nil {
			return true
		}
		// Check receiver for method calls
		if sel, ok := e.Func.(*ast.SelectorExpr); ok {
			if hasBlock(sel.X) {
				return true
			}
		}
		// Check arguments
		if slices.ContainsFunc(e.Args, hasBlock) {
			return true
		}
	case *ast.SelectorExpr:
		return hasBlock(e.X)
	case *ast.BinaryExpr:
		return hasBlock(e.Left) || hasBlock(e.Right)
	case *ast.UnaryExpr:
		return hasBlock(e.Expr)
	case *ast.IndexExpr:
		return hasBlock(e.Left) || hasBlock(e.Index)
	}
	return false
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

// isStructDef checks if the input is a struct definition.
func (m *Model) isStructDef(input string) bool {
	trimmed := strings.TrimSpace(input)
	return strings.HasPrefix(trimmed, "struct ") || strings.HasPrefix(trimmed, "pub struct ")
}

// isEnumDef checks if the input is an enum definition.
func (m *Model) isEnumDef(input string) bool {
	trimmed := strings.TrimSpace(input)
	return strings.HasPrefix(trimmed, "enum ") || strings.HasPrefix(trimmed, "pub enum ")
}

// isModuleDef checks if the input is a module definition.
func (m *Model) isModuleDef(input string) bool {
	trimmed := strings.TrimSpace(input)
	return strings.HasPrefix(trimmed, "module ") || strings.HasPrefix(trimmed, "pub module ")
}

// truncateOutput keeps output to a maximum of 100 lines.
func (m *Model) truncateOutput() {
	if len(m.output) > 100 {
		m.output = m.output[len(m.output)-100:]
	}
}

// eval compiles and runs the input, returning an evalResult.
// Uses ctx for read-only state to avoid data races during async evaluation.
func (m *Model) eval(input string, ctx evalContext) evalResult {
	result := evalResult{input: input}

	// Handle special inputs (with or without leading slash)
	trimmed := strings.TrimSpace(input)
	cmd := strings.TrimPrefix(trimmed, "/")
	if cmd == "exit" || cmd == "quit" {
		result.quit = true
		return result
	}

	// Handle reset command
	if cmd == "reset" || cmd == "clear" {
		result.clearState = true
		result.clearOutput = true
		result.output = successStyle.Render("(state cleared)")
		return result
	}

	// Handle help command
	if cmd == "help" {
		result.output = mutedStyle.Render(`Commands:
  reset, clear  - Clear all accumulated state
  exit, quit    - Exit the REPL (or use Ctrl+D)
  help          - Show this message`)
		return result
	}

	// Accumulate state based on input type
	if m.isImport(input) {
		result.addImport = input
		result.output = successStyle.Render("(imported)")
		return result
	}

	if m.isFunctionDef(input) {
		result.addFunction = input
		result.output = successStyle.Render("(defined)")
		return result
	}

	if m.isClassDef(input) {
		result.addFunction = input
		result.output = successStyle.Render("(defined)")
		return result
	}

	if m.isInterfaceDef(input) {
		result.addFunction = input
		result.output = successStyle.Render("(defined)")
		return result
	}

	if m.isStructDef(input) {
		result.addFunction = input
		result.output = successStyle.Render("(defined)")
		return result
	}

	if m.isEnumDef(input) {
		result.addFunction = input
		result.output = successStyle.Render("(defined)")
		return result
	}

	if m.isModuleDef(input) {
		result.addFunction = input
		result.output = successStyle.Render("(defined)")
		return result
	}

	// Check for valid project
	if ctx.project == nil {
		result.err = fmt.Errorf("no Rugby project found - run from a project directory")
		return result
	}

	// Build the program using captured state
	program := m.buildProgramFromContext(input, ctx)

	// Ensure project dirs exist
	if err := ctx.project.EnsureDirs(); err != nil {
		result.err = err
		return result
	}

	// Write to temp file
	replDir := filepath.Join(ctx.project.RugbyDir, "repl")
	if err := os.MkdirAll(replDir, 0755); err != nil {
		result.err = err
		return result
	}

	srcFile := filepath.Join(replDir, fmt.Sprintf("repl_%d.rg", ctx.counter))
	if err := os.WriteFile(srcFile, []byte(program), 0644); err != nil {
		result.err = err
		return result
	}

	// Parse
	l := lexer.New(program)
	p := parser.New(l)
	parsedAST := p.ParseProgram()

	if len(p.Errors()) > 0 {
		result.err = fmt.Errorf("parse error: %s", p.Errors()[0])
		return result
	}

	// Semantic analysis
	analyzer := semantic.NewAnalyzer()
	semanticErrors := analyzer.Analyze(parsedAST)
	if len(semanticErrors) > 0 {
		result.err = semanticErrors[0]
		return result
	}

	// Generate
	typeInfo := &typeInfoAdapter{analyzer: analyzer}
	gen := codegen.New(codegen.WithTypeInfo(typeInfo))
	goCode, err := gen.Generate(parsedAST)
	if err != nil {
		result.err = fmt.Errorf("codegen error: %v", err)
		return result
	}

	// Write Go file
	goFile := filepath.Join(replDir, fmt.Sprintf("repl_%d.go", ctx.counter))
	err = os.WriteFile(goFile, []byte(goCode), 0644)
	if err != nil {
		result.err = err
		return result
	}

	// Create go.mod for the repl directory
	goModContent := fmt.Sprintf(`module repl

go 1.25

require %s %s
`, builder.RuntimeModule, builder.RuntimeVersion)

	// In dev mode (running from rugby source), add replace directive
	if inRepo, repoPath := builder.IsInRugbyRepo(); inRepo {
		goModContent += fmt.Sprintf("\nreplace %s => %s\n", builder.RuntimeModule, repoPath)
	}

	goModFile := filepath.Join(replDir, "go.mod")
	err = os.WriteFile(goModFile, []byte(goModContent), 0644)
	if err != nil {
		result.err = err
		return result
	}

	// Run go mod tidy (ignore errors - it may fail if no dependencies needed)
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = replDir
	_ = tidyCmd.Run()

	// Compile
	binFile := filepath.Join(replDir, fmt.Sprintf("repl_%d", ctx.counter))
	buildCmd := exec.Command("go", "build", "-o", binFile, goFile)
	buildCmd.Dir = replDir

	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		result.err = fmt.Errorf("build error: %s", strings.TrimSpace(string(buildOut)))
		return result
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
				result.err = fmt.Errorf("runtime error (exit %d)", exitErr.ExitCode())
				return result
			}
		}
	}

	result.output = output
	return result
}

// isAssignment checks if the input is a variable assignment.
func (m *Model) isAssignment(input string) bool {
	stmt := m.parseStatement(input)
	if stmt == nil {
		return false
	}
	switch stmt.(type) {
	case *ast.AssignStmt, *ast.CompoundAssignStmt:
		return true
	}
	return false
}

// getAssignmentVar returns the variable name if input is an assignment.
func (m *Model) getAssignmentVar(input string) string {
	stmt := m.parseStatement(input)
	if stmt == nil {
		return ""
	}
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		return s.Name
	case *ast.CompoundAssignStmt:
		return s.Name
	}
	return ""
}

// buildProgram constructs a full Rugby program using current model state.
// This is a convenience method that creates a context from model state.
func (m *Model) buildProgram(input string) string {
	ctx := evalContext{
		imports:    m.imports,
		functions:  m.functions,
		statements: m.statements,
	}
	return m.buildProgramFromContext(input, ctx)
}

// buildProgramFromContext constructs a full Rugby program using captured state.
// Top-level statements are written directly and codegen wraps them in main().
func (m *Model) buildProgramFromContext(input string, ctx evalContext) string {
	var buf strings.Builder

	// Imports
	for _, imp := range ctx.imports {
		buf.WriteString(imp + "\n")
	}

	// Function/class definitions
	for _, fn := range ctx.functions {
		buf.WriteString(fn + "\n")
	}

	// Prior statements (for variable persistence)
	// Track variables so we can suppress "declared and not used" errors
	var declaredVars []string
	for _, stmt := range ctx.statements {
		buf.WriteString(stmt + "\n")
		if varName := m.getAssignmentVar(stmt); varName != "" {
			declaredVars = append(declaredVars, varName)
		}
	}

	// Suppress unused variable errors for accumulated variables
	for _, v := range declaredVars {
		buf.WriteString("_ = " + v + "\n")
	}

	// New input - wrap expressions in p() to print
	if m.isExpression(input) {
		buf.WriteString("p(" + input + ")\n")
	} else if m.isBlockExpression(input) {
		// Block expressions can't be wrapped in p() directly
		// Assign to temp var first, then print
		buf.WriteString("__repl_result__ = " + input + "\n")
		buf.WriteString("p(__repl_result__)\n")
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

// Run starts the REPL with the given version string.
func Run(version string) error {
	fmt.Println(mutedStyle.Render(version+" REPL") + " " + mutedStyle.Render("(Ctrl+D to quit, 'help' for commands)"))
	fmt.Println()

	p := tea.NewProgram(New())
	_, err := p.Run()
	return err
}
