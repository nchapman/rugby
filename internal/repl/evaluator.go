// Package repl implements the interactive Rugby REPL.
package repl

import (
	"github.com/nchapman/rugby/internal/builder"
)

// Evaluator provides a testable interface to the REPL's evaluation logic.
// It maintains state (imports, functions, statements) across evaluations,
// mimicking an interactive REPL session.
type Evaluator struct {
	model   *Model
	imports []string
	funcs   []string
	stmts   []string
	counter int
	project *builder.Project
}

// NewEvaluator creates a new Evaluator for testing REPL sessions.
func NewEvaluator() (*Evaluator, error) {
	project, err := builder.FindProject()
	if err != nil {
		return nil, err
	}
	m := New()
	return &Evaluator{
		model:   &m,
		project: project,
	}, nil
}

// NewEvaluatorWithProject creates a new Evaluator with a specific project.
func NewEvaluatorWithProject(project *builder.Project) *Evaluator {
	m := New()
	return &Evaluator{
		model:   &m,
		project: project,
	}
}

// EvalResult holds the result of evaluating an input.
type EvalResult struct {
	Output      string // stdout from the evaluation
	Error       error  // error if evaluation failed
	IsQuit      bool   // true if the input was an exit command
	WasCleared  bool   // true if state was cleared (reset command)
	WasDefined  bool   // true if a function/class/interface was defined
	WasImported bool   // true if an import was added
}

// Eval evaluates a single input and returns the result.
// State changes (imports, functions, variable assignments) are tracked
// automatically for subsequent evaluations.
func (e *Evaluator) Eval(input string) EvalResult {
	ctx := evalContext{
		imports:    append([]string{}, e.imports...),
		functions:  append([]string{}, e.funcs...),
		statements: append([]string{}, e.stmts...),
		project:    e.project,
		counter:    e.counter,
	}
	e.counter++

	result := e.model.eval(input, ctx)

	// Apply state changes
	if result.clearState {
		e.imports = nil
		e.funcs = nil
		e.stmts = nil
	}
	if result.addImport != "" {
		e.imports = append(e.imports, result.addImport)
	}
	if result.addFunction != "" {
		e.funcs = append(e.funcs, result.addFunction)
	}
	if e.model.isAssignment(result.input) {
		e.stmts = append(e.stmts, result.input)
	}

	return EvalResult{
		Output:      result.output,
		Error:       result.err,
		IsQuit:      result.quit,
		WasCleared:  result.clearState,
		WasDefined:  result.addFunction != "",
		WasImported: result.addImport != "",
	}
}

// Reset clears all accumulated state.
func (e *Evaluator) Reset() {
	e.imports = nil
	e.funcs = nil
	e.stmts = nil
}

// State returns the current accumulated state for debugging.
func (e *Evaluator) State() (imports, functions, statements []string) {
	return e.imports, e.funcs, e.stmts
}
