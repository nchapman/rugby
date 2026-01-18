package codegen

// Control flow statement code generation.
// This file handles:
//   - Conditionals (if, elsif, else, unless)
//   - Case statements (case/when, case_type)
//   - Loops (while, until, loop, for)

import (
	"fmt"
	"strings"

	"github.com/nchapman/rugby/ast"
)

func (g *Generator) genCondition(cond ast.Expression) error {
	condType := g.inferTypeFromExpr(cond)

	// Allow Bool type explicitly
	if condType == "Bool" {
		g.genExpr(cond)
		return nil
	}

	// Allow nil comparisons (err != nil, x == nil) - these return Bool
	if binary, ok := cond.(*ast.BinaryExpr); ok {
		if binary.Op == "==" || binary.Op == "!=" {
			if _, isNil := binary.Right.(*ast.NilLit); isNil {
				g.genExpr(cond)
				return nil
			}
			if _, isNil := binary.Left.(*ast.NilLit); isNil {
				g.genExpr(cond)
				return nil
			}
		}
	}

	// If type is unknown, generate as-is (may be external function returning bool)
	if condType == "" {
		g.genExpr(cond)
		return nil
	}

	// Optional types require explicit unwrapping
	if isOptionalType(condType) {
		return fmt.Errorf("condition must be Bool, got %s (use 'if let x = ...' or 'x != nil' for optionals)", condType)
	}

	// Non-Bool types are errors
	return fmt.Errorf("condition must be Bool, got %s", condType)
}

func (g *Generator) genIfStmt(s *ast.IfStmt) {
	g.writeIndent()
	g.buf.WriteString("if ")

	// Handle assignment-in-condition pattern: if let n = expr
	// For pointer-based optionals (T?): if _tmp := expr; _tmp != nil { n := *_tmp; ... }
	// For tuple-returning functions: if n, ok := expr; ok { ... }
	if s.AssignName != "" {
		exprType := g.inferTypeFromExpr(s.AssignExpr)

		if isOptionalType(exprType) {
			// Pointer-based optional: generate temp var and nil check
			tempVar := fmt.Sprintf("_iflet%d", g.tempVarCounter)
			g.tempVarCounter++
			g.buf.WriteString(tempVar)
			g.buf.WriteString(" := ")
			g.genExpr(s.AssignExpr)
			g.buf.WriteString("; ")
			g.buf.WriteString(tempVar)
			g.buf.WriteString(" != nil {\n")

			// Unwrap the pointer into the user's variable
			g.indent++
			g.writeIndent()
			g.buf.WriteString(s.AssignName)
			g.buf.WriteString(" := *")
			g.buf.WriteString(tempVar)
			g.buf.WriteString("\n")
			// Silence "declared and not used" error if variable isn't referenced
			g.writeIndent()
			g.buf.WriteString("_ = ")
			g.buf.WriteString(s.AssignName)
			g.buf.WriteString("\n")
			g.indent--

			// Track the variable with its unwrapped type
			unwrappedType := strings.TrimSuffix(exprType, "?")
			g.vars[s.AssignName] = unwrappedType
		} else {
			// Tuple-returning function: if n, ok := expr; ok { ... }
			g.buf.WriteString(s.AssignName)
			g.buf.WriteString(", ok := ")
			g.genExpr(s.AssignExpr)
			g.buf.WriteString("; ok {\n")

			// Track the variable with its inferred type
			// Special handling for .as(Type) calls - extract the target type
			varType := exprType
			if call, ok := s.AssignExpr.(*ast.CallExpr); ok {
				if sel, ok := call.Func.(*ast.SelectorExpr); ok && sel.Sel == "as" && len(call.Args) == 1 {
					if typeArg, ok := call.Args[0].(*ast.Ident); ok {
						varType = typeArg.Name
					}
				}
			}
			g.vars[s.AssignName] = varType
		}
	} else {
		// For unless, negate the condition
		if s.IsUnless {
			g.buf.WriteString("!(")
		}
		if err := g.genCondition(s.Cond); err != nil {
			g.addError(fmt.Errorf("line %d: %w", s.Line, err))
			// Generate a placeholder to keep the output somewhat valid
			g.buf.WriteString("false")
		}
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
	}

	g.indent++
	for _, stmt := range s.Then {
		g.genStatement(stmt)
	}
	g.indent--

	for _, elsif := range s.ElseIfs {
		g.writeIndent()
		g.buf.WriteString("} else if ")
		if err := g.genCondition(elsif.Cond); err != nil {
			g.addError(fmt.Errorf("line %d: %w", s.Line, err))
			g.buf.WriteString("false")
		}
		g.buf.WriteString(" {\n")

		g.indent++
		for _, stmt := range elsif.Body {
			g.genStatement(stmt)
		}
		g.indent--
	}

	if len(s.Else) > 0 {
		g.writeIndent()
		g.buf.WriteString("} else {\n")

		g.indent++
		for _, stmt := range s.Else {
			g.genStatement(stmt)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genCaseStmt(s *ast.CaseStmt) {
	g.writeIndent()

	// Check if any when clause contains a range (requires special handling)
	hasRanges := false
	for _, when := range s.WhenClauses {
		for _, val := range when.Values {
			if _, ok := val.(*ast.RangeLit); ok {
				hasRanges = true
				break
			}
		}
		if hasRanges {
			break
		}
	}

	// Handle case with subject vs case without subject
	if s.Subject != nil && !hasRanges {
		// Simple switch: switch subject { case value: ... }
		g.buf.WriteString("switch ")
		g.genExpr(s.Subject)
		g.buf.WriteString(" {\n")
	} else if s.Subject != nil && hasRanges {
		// Has ranges: need to store subject and use switch true pattern
		g.needsRuntime = true
		g.buf.WriteString("switch _subj := ")
		g.genExpr(s.Subject)
		g.buf.WriteString("; true {\n")
	} else {
		// Case without subject - use switch true
		g.buf.WriteString("switch {\n")
	}

	// Generate when clauses
	for _, whenClause := range s.WhenClauses {
		g.writeIndent()

		if s.Subject != nil && !hasRanges {
			// With subject (no ranges): case value1, value2:
			g.buf.WriteString("case ")
			for i, val := range whenClause.Values {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				g.genExpr(val)
			}
			g.buf.WriteString(":\n")
		} else if s.Subject != nil && hasRanges {
			// With subject and ranges: case _subj == value || runtime.RangeContains(range, _subj):
			g.buf.WriteString("case ")
			for i, val := range whenClause.Values {
				if i > 0 {
					g.buf.WriteString(" || ")
				}
				if rangeLit, ok := val.(*ast.RangeLit); ok {
					// Generate runtime.RangeContains(range, _subj)
					g.buf.WriteString("runtime.RangeContains(")
					g.genExpr(rangeLit)
					g.buf.WriteString(", _subj)")
				} else {
					// Generate _subj == value
					g.buf.WriteString("_subj == ")
					g.genExpr(val)
				}
			}
			g.buf.WriteString(":\n")
		} else {
			// Without subject: case condition1 || condition2:
			g.buf.WriteString("case ")
			for i, val := range whenClause.Values {
				if i > 0 {
					g.buf.WriteString(" || ")
				}
				if err := g.genCondition(val); err != nil {
					g.addError(err)
					g.buf.WriteString("false")
				}
			}
			g.buf.WriteString(":\n")
		}

		g.indent++
		for _, stmt := range whenClause.Body {
			g.genStatement(stmt)
		}
		g.indent--
	}

	// Generate default (else) clause
	if len(s.Else) > 0 {
		g.writeIndent()
		g.buf.WriteString("default:\n")

		g.indent++
		for _, stmt := range s.Else {
			g.genStatement(stmt)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genCaseStmtWithReturns generates a case statement where each branch returns its last value.
// This is used when a case expression is the last statement in a method with a return type.
func (g *Generator) genCaseStmtWithReturns(s *ast.CaseStmt) {
	g.writeIndent()

	// Check if any when clause contains a range (requires special handling)
	hasRanges := false
	for _, when := range s.WhenClauses {
		for _, val := range when.Values {
			if _, ok := val.(*ast.RangeLit); ok {
				hasRanges = true
				break
			}
		}
		if hasRanges {
			break
		}
	}

	// Handle case with subject vs case without subject
	if s.Subject != nil && !hasRanges {
		g.buf.WriteString("switch ")
		g.genExpr(s.Subject)
		g.buf.WriteString(" {\n")
	} else if s.Subject != nil && hasRanges {
		g.needsRuntime = true
		g.buf.WriteString("switch _subj := ")
		g.genExpr(s.Subject)
		g.buf.WriteString("; true {\n")
	} else {
		g.buf.WriteString("switch {\n")
	}

	// Generate when clauses with returns
	for _, whenClause := range s.WhenClauses {
		g.writeIndent()

		if s.Subject != nil && !hasRanges {
			g.buf.WriteString("case ")
			for i, val := range whenClause.Values {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				g.genExpr(val)
			}
			g.buf.WriteString(":\n")
		} else if s.Subject != nil && hasRanges {
			g.buf.WriteString("case ")
			for i, val := range whenClause.Values {
				if i > 0 {
					g.buf.WriteString(" || ")
				}
				if rangeLit, ok := val.(*ast.RangeLit); ok {
					g.buf.WriteString("runtime.RangeContains(")
					g.genExpr(rangeLit)
					g.buf.WriteString(", _subj)")
				} else {
					g.buf.WriteString("_subj == ")
					g.genExpr(val)
				}
			}
			g.buf.WriteString(":\n")
		} else {
			g.buf.WriteString("case ")
			for i, val := range whenClause.Values {
				if i > 0 {
					g.buf.WriteString(" || ")
				}
				if err := g.genCondition(val); err != nil {
					g.addError(err)
					g.buf.WriteString("false")
				}
			}
			g.buf.WriteString(":\n")
		}

		g.indent++
		// Generate all but last statement normally
		for i, stmt := range whenClause.Body {
			if i == len(whenClause.Body)-1 {
				// For the last statement, if it's an expression, add return
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					g.writeIndent()
					g.buf.WriteString("return ")
					g.genExpr(exprStmt.Expr)
					g.buf.WriteString("\n")
					continue
				}
			}
			g.genStatement(stmt)
		}
		g.indent--
	}

	// Generate default (else) clause with returns
	if len(s.Else) > 0 {
		g.writeIndent()
		g.buf.WriteString("default:\n")

		g.indent++
		for i, stmt := range s.Else {
			if i == len(s.Else)-1 {
				// For the last statement, if it's an expression, add return
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					g.writeIndent()
					g.buf.WriteString("return ")
					g.genExpr(exprStmt.Expr)
					g.buf.WriteString("\n")
					continue
				}
			}
			g.genStatement(stmt)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genIfStmtWithReturns generates an if statement where each branch returns its last value.
// This is used when an if expression is the last statement in a method with a return type.
func (g *Generator) genIfStmtWithReturns(s *ast.IfStmt) {
	g.writeIndent()
	g.buf.WriteString("if ")

	// Handle assignment-in-condition pattern: if let n = expr
	if s.AssignName != "" {
		exprType := g.inferTypeFromExpr(s.AssignExpr)

		if isOptionalType(exprType) {
			// Pointer-based optional: generate temp var and nil check
			tempVar := fmt.Sprintf("_iflet%d", g.tempVarCounter)
			g.tempVarCounter++
			g.buf.WriteString(tempVar)
			g.buf.WriteString(" := ")
			g.genExpr(s.AssignExpr)
			g.buf.WriteString("; ")
			g.buf.WriteString(tempVar)
			g.buf.WriteString(" != nil {\n")

			// Unwrap the pointer into the user's variable
			g.indent++
			g.writeIndent()
			g.buf.WriteString(s.AssignName)
			g.buf.WriteString(" := *")
			g.buf.WriteString(tempVar)
			g.buf.WriteString("\n")
			// Silence "declared and not used" error if variable isn't referenced
			g.writeIndent()
			g.buf.WriteString("_ = ")
			g.buf.WriteString(s.AssignName)
			g.buf.WriteString("\n")
			g.indent--

			// Track the variable with its unwrapped type
			unwrappedType := strings.TrimSuffix(exprType, "?")
			g.vars[s.AssignName] = unwrappedType
		} else {
			// Tuple-returning function: if n, ok := expr; ok { ... }
			g.buf.WriteString(s.AssignName)
			g.buf.WriteString(", ok := ")
			g.genExpr(s.AssignExpr)
			g.buf.WriteString("; ok {\n")
			g.vars[s.AssignName] = exprType
		}
	} else {
		// For unless, negate the condition
		if s.IsUnless {
			g.buf.WriteString("!(")
		}
		if err := g.genCondition(s.Cond); err != nil {
			g.addError(err)
		}
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
	}

	g.indent++
	for i, stmt := range s.Then {
		if i == len(s.Then)-1 {
			// For the last statement, if it's an expression, add return
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				g.writeIndent()
				g.buf.WriteString("return ")
				g.genExpr(exprStmt.Expr)
				g.buf.WriteString("\n")
				continue
			}
		}
		g.genStatement(stmt)
	}
	g.indent--

	// Handle elsif clauses
	for _, elsif := range s.ElseIfs {
		g.writeIndent()
		g.buf.WriteString("} else if ")
		if err := g.genCondition(elsif.Cond); err != nil {
			g.addError(err)
		}
		g.buf.WriteString(" {\n")

		g.indent++
		for i, stmt := range elsif.Body {
			if i == len(elsif.Body)-1 {
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					g.writeIndent()
					g.buf.WriteString("return ")
					g.genExpr(exprStmt.Expr)
					g.buf.WriteString("\n")
					continue
				}
			}
			g.genStatement(stmt)
		}
		g.indent--
	}

	// Handle else clause
	if len(s.Else) > 0 {
		g.writeIndent()
		g.buf.WriteString("} else {\n")

		g.indent++
		for i, stmt := range s.Else {
			if i == len(s.Else)-1 {
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					g.writeIndent()
					g.buf.WriteString("return ")
					g.genExpr(exprStmt.Expr)
					g.buf.WriteString("\n")
					continue
				}
			}
			g.genStatement(stmt)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genCaseTypeStmt(s *ast.CaseTypeStmt) {
	g.writeIndent()

	// Check if any when clause has a binding variable
	hasBindings := false
	for _, whenClause := range s.WhenClauses {
		if whenClause.BindingVar != "" {
			hasBindings = true
			break
		}
	}

	// Generate a temp variable for the type switch (only if we have bindings)
	tempVar := "_"
	if hasBindings {
		tempVar = fmt.Sprintf("_ts%d", g.tempVarCounter)
		g.tempVarCounter++
	}

	// If subject is a simple identifier, we can use it directly
	// Otherwise, evaluate and store in temp
	var subjectExpr string
	if ident, ok := s.Subject.(*ast.Ident); ok {
		subjectExpr = ident.Name
	} else {
		// For complex expressions, always generate a temp var
		evalVar := fmt.Sprintf("_ts%d", g.tempVarCounter)
		g.tempVarCounter++
		g.buf.WriteString(fmt.Sprintf("%s := ", evalVar))
		g.genExpr(s.Subject)
		g.buf.WriteString("\n")
		g.writeIndent()
		subjectExpr = evalVar
	}

	// Generate type switch: switch _ts := subject.(type) { or switch subject.(type) {
	if hasBindings {
		g.buf.WriteString(fmt.Sprintf("switch %s := %s.(type) {\n", tempVar, subjectExpr))
	} else {
		g.buf.WriteString(fmt.Sprintf("switch %s.(type) {\n", subjectExpr))
	}

	// Generate when clauses for each type
	for _, whenClause := range s.WhenClauses {
		g.writeIndent()
		g.buf.WriteString("case ")
		// Classes use pointer receivers, so use pointer types in type switch
		goType := mapType(whenClause.Type)
		if g.isClass(whenClause.Type) && !strings.HasPrefix(goType, "*") {
			g.buf.WriteString("*")
		}
		g.buf.WriteString(goType)
		g.buf.WriteString(":\n")

		g.indent++

		// If binding variable is specified, create the binding
		var prevType string
		var wasDefinedBefore bool
		if whenClause.BindingVar != "" {
			// Save previous state for cleanup
			prevType, wasDefinedBefore = g.vars[whenClause.BindingVar]

			g.writeIndent()
			g.buf.WriteString(fmt.Sprintf("%s := %s\n", whenClause.BindingVar, tempVar))
			// Suppress "declared and not used" if variable isn't referenced
			g.writeIndent()
			g.buf.WriteString(fmt.Sprintf("_ = %s\n", whenClause.BindingVar))
			// Track the variable so it's not re-declared
			g.vars[whenClause.BindingVar] = whenClause.Type
		}

		for _, stmt := range whenClause.Body {
			g.genStatement(stmt)
		}

		// Clean up variable scope
		if whenClause.BindingVar != "" {
			if !wasDefinedBefore {
				delete(g.vars, whenClause.BindingVar)
			} else {
				g.vars[whenClause.BindingVar] = prevType
			}
		}

		g.indent--
	}

	// Generate default (else) clause
	if len(s.Else) > 0 {
		g.writeIndent()
		g.buf.WriteString("default:\n")

		g.indent++
		for _, stmt := range s.Else {
			g.genStatement(stmt)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genWhileStmt(s *ast.WhileStmt) {
	g.writeIndent()
	g.buf.WriteString("for ")
	if err := g.genCondition(s.Cond); err != nil {
		g.addError(fmt.Errorf("line %d: %w", s.Line, err))
		g.buf.WriteString("false")
	}
	g.buf.WriteString(" {\n")

	g.enterLoop()
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.exitLoop()

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genUntilStmt(s *ast.UntilStmt) {
	g.writeIndent()
	g.buf.WriteString("for !")
	// Wrap condition in parentheses unless it's a simple expression.
	// Simple expressions that don't need parens: identifiers, booleans, calls, selectors.
	// All other expressions (binary, nil-coalesce, etc.) need parens for correct precedence.
	needsParens := true
	switch s.Cond.(type) {
	case *ast.Ident, *ast.BoolLit, *ast.CallExpr, *ast.SelectorExpr:
		needsParens = false
	}
	if needsParens {
		g.buf.WriteString("(")
	}
	if err := g.genCondition(s.Cond); err != nil {
		g.addError(fmt.Errorf("line %d: %w", s.Line, err))
		g.buf.WriteString("false")
	}
	if needsParens {
		g.buf.WriteString(")")
	}
	g.buf.WriteString(" {\n")

	g.enterLoop()
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.exitLoop()

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genLoopStmt(s *ast.LoopStmt) {
	g.writeIndent()
	g.buf.WriteString("for {\n")

	g.enterLoop()
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.exitLoop()

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genForStmt(s *ast.ForStmt) {
	// Save variable state
	prevType, wasDefinedBefore := g.vars[s.Var]
	var prevType2 string
	var wasDefinedBefore2 bool
	if s.Var2 != "" {
		prevType2, wasDefinedBefore2 = g.vars[s.Var2]
	}

	// Check if iterable is a range literal - optimize to C-style for loop
	if rangeLit, ok := s.Iterable.(*ast.RangeLit); ok {
		g.genForRangeLoop(s.Var, rangeLit, s.Body, wasDefinedBefore, prevType)
		return
	}

	// Check if iterable is a variable of type Range
	if ident, ok := s.Iterable.(*ast.Ident); ok {
		if g.typeInfo.GetRugbyType(ident) == "Range" {
			g.genForRangeVarLoop(s.Var, ident.Name, s.Body, wasDefinedBefore, prevType)
			return
		}
	}

	g.writeIndent()
	// Check if iterating over a channel - channels only allow one iteration variable
	// Note: type might be in Rugby format (Chan<T>) or Go format (chan T)
	iterType := g.inferTypeFromExpr(s.Iterable)
	if strings.HasPrefix(iterType, "chan ") || strings.HasPrefix(iterType, "Chan<") {
		g.buf.WriteString("for ")
		g.buf.WriteString(s.Var)
		g.buf.WriteString(" := range ")
		g.genExpr(s.Iterable)
		g.buf.WriteString(" {\n")
	} else if s.Var2 != "" {
		// Two-variable form: for key, value in map/array
		g.buf.WriteString("for ")
		g.buf.WriteString(s.Var)
		g.buf.WriteString(", ")
		g.buf.WriteString(s.Var2)
		g.buf.WriteString(" := range ")
		g.genExpr(s.Iterable)
		g.buf.WriteString(" {\n")
	} else if strings.HasPrefix(iterType, "map[") || strings.HasPrefix(iterType, "Map<") {
		// Single-variable map iteration: get keys only
		g.buf.WriteString("for ")
		g.buf.WriteString(s.Var)
		g.buf.WriteString(" := range ")
		g.genExpr(s.Iterable)
		g.buf.WriteString(" {\n")
	} else {
		// Single-variable array iteration: skip index, get values
		g.buf.WriteString("for _, ")
		g.buf.WriteString(s.Var)
		g.buf.WriteString(" := range ")
		g.genExpr(s.Iterable)
		g.buf.WriteString(" {\n")
	}

	g.vars[s.Var] = "" // loop variable, type unknown
	if s.Var2 != "" {
		g.vars[s.Var2] = "" // second loop variable
		// In two-variable form over arrays, first var is the index (always non-negative)
		if !strings.HasPrefix(iterType, "map[") && !strings.HasPrefix(iterType, "Map<") {
			g.nonNegativeVars[s.Var] = true
		}
	}

	g.enterLoop()
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.exitLoop()

	// Restore variable state (including non-negative tracking)
	if !wasDefinedBefore {
		delete(g.vars, s.Var)
		delete(g.nonNegativeVars, s.Var)
	} else {
		g.vars[s.Var] = prevType
	}
	if s.Var2 != "" {
		if !wasDefinedBefore2 {
			delete(g.vars, s.Var2)
			delete(g.nonNegativeVars, s.Var2)
		} else {
			g.vars[s.Var2] = prevType2
		}
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genForRangeVarLoop(varName string, rangeVar string, body []ast.Statement, wasDefinedBefore bool, prevType string) {
	g.writeIndent()
	g.buf.WriteString("for ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := ")
	g.buf.WriteString(rangeVar)
	g.buf.WriteString(".Start; ")

	// Condition: (r.Exclusive && i < r.End) || (!r.Exclusive && i <= r.End)
	g.buf.WriteString(fmt.Sprintf("(%s.Exclusive && %s < %s.End) || (!%s.Exclusive && %s <= %s.End)",
		rangeVar, varName, rangeVar, rangeVar, varName, rangeVar))

	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	g.buf.WriteString("++ {\n")

	g.vars[varName] = "Int"

	g.enterLoop()
	g.indent++
	for _, stmt := range body {
		g.genStatement(stmt)
	}
	g.indent--
	g.exitLoop()

	// Restore variable state (including non-negative tracking)
	if !wasDefinedBefore {
		delete(g.vars, varName)
		delete(g.nonNegativeVars, varName)
	} else {
		g.vars[varName] = prevType
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genForRangeLoop(varName string, r *ast.RangeLit, body []ast.Statement, wasDefinedBefore bool, prevType string) {
	// Optimization: use Go 1.22 range-over-int for 0...n loops
	// Check if start is 0
	isStartZero := false
	if intLit, ok := r.Start.(*ast.IntLit); ok && intLit.Value == 0 {
		isStartZero = true
	}

	g.writeIndent()
	if isStartZero && r.Exclusive {
		// Generate: for i := range end { ... }
		g.buf.WriteString("for ")
		g.buf.WriteString(varName)
		g.buf.WriteString(" := range ")
		g.genExpr(r.End)
		g.buf.WriteString(" {\n")
		// Loop variable starting from 0 is always non-negative
		g.nonNegativeVars[varName] = true
	} else {
		// Generate: for i := start; i <= end; i++ { ... }
		// or:       for i := start; i < end; i++ { ... } (exclusive)
		g.buf.WriteString("for ")
		g.buf.WriteString(varName)
		g.buf.WriteString(" := ")
		g.genExpr(r.Start)
		g.buf.WriteString("; ")
		g.buf.WriteString(varName)
		if r.Exclusive {
			g.buf.WriteString(" < ")
		} else {
			g.buf.WriteString(" <= ")
		}
		g.genExpr(r.End)
		g.buf.WriteString("; ")
		g.buf.WriteString(varName)
		g.buf.WriteString("++ {\n")
	}

	g.vars[varName] = "Int" // range loop variable is always Int

	// Track loop variable as non-negative if start is non-negative
	if g.isNonNegativeValue(r.Start) {
		g.nonNegativeVars[varName] = true
	}

	g.enterLoop()
	g.indent++
	for _, stmt := range body {
		g.genStatement(stmt)
	}
	g.indent--
	g.exitLoop()

	// Restore variable state (including non-negative tracking)
	if !wasDefinedBefore {
		delete(g.vars, varName)
		delete(g.nonNegativeVars, varName)
	} else {
		g.vars[varName] = prevType
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}
