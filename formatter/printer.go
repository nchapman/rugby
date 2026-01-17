package formatter

import (
	"fmt"
	"strings"

	"github.com/nchapman/rugby/ast"
)

// formatProgram formats an entire program.
func (f *Formatter) formatProgram(prog *ast.Program) {
	// Pre-mark all attached comments so they don't get emitted as free-floating
	f.preMarkAttachedComments(prog)

	// Format imports
	for i, imp := range prog.Imports {
		f.emitFreeFloatingComments(imp.Line)
		f.formatImport(imp)
		f.lastLine = imp.Line
		// Blank line after imports block
		if i == len(prog.Imports)-1 && len(prog.Declarations) > 0 {
			f.write("\n")
		}
	}

	// Format declarations
	for i, decl := range prog.Declarations {
		line := f.getStatementLine(decl)
		f.emitFreeFloatingComments(line)
		f.formatStatement(decl)
		f.lastLine = line

		// Blank line between top-level declarations
		if i < len(prog.Declarations)-1 {
			f.write("\n")
		}
	}

	// Emit any remaining comments
	f.emitRemainingComments()
}

// preMarkAttachedComments marks all comments attached to AST nodes as emitted,
// so they won't be duplicated by the free-floating comment handler.
func (f *Formatter) preMarkAttachedComments(prog *ast.Program) {
	// Mark import comments
	for _, imp := range prog.Imports {
		f.markCommentGroupEmitted(imp.Doc)
		f.markCommentGroupEmitted(imp.Comment)
	}

	// Mark declaration comments
	for _, decl := range prog.Declarations {
		f.preMarkStatementComments(decl)
	}
}

// preMarkStatementComments marks comments attached to a statement.
func (f *Formatter) preMarkStatementComments(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.FuncDecl:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
		for _, bodyStmt := range s.Body {
			f.preMarkStatementComments(bodyStmt)
		}
	case *ast.ClassDecl:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
		for _, method := range s.Methods {
			f.markCommentGroupEmitted(method.Doc)
			f.markCommentGroupEmitted(method.Comment)
			for _, bodyStmt := range method.Body {
				f.preMarkStatementComments(bodyStmt)
			}
		}
	case *ast.InterfaceDecl:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
	case *ast.IfStmt:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
		for _, bodyStmt := range s.Then {
			f.preMarkStatementComments(bodyStmt)
		}
		for _, elsif := range s.ElseIfs {
			for _, bodyStmt := range elsif.Body {
				f.preMarkStatementComments(bodyStmt)
			}
		}
		for _, bodyStmt := range s.Else {
			f.preMarkStatementComments(bodyStmt)
		}
	case *ast.WhileStmt:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
		for _, bodyStmt := range s.Body {
			f.preMarkStatementComments(bodyStmt)
		}
	case *ast.ForStmt:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
		for _, bodyStmt := range s.Body {
			f.preMarkStatementComments(bodyStmt)
		}
	case *ast.LoopStmt:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
		for _, bodyStmt := range s.Body {
			f.preMarkStatementComments(bodyStmt)
		}
	case *ast.CaseStmt:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
		for _, when := range s.WhenClauses {
			for _, bodyStmt := range when.Body {
				f.preMarkStatementComments(bodyStmt)
			}
		}
		for _, bodyStmt := range s.Else {
			f.preMarkStatementComments(bodyStmt)
		}
	case *ast.AssignStmt:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
	case *ast.ExprStmt:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
	case *ast.ReturnStmt:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
	}
}

// getStatementLine returns the line number of a statement.
func (f *Formatter) getStatementLine(stmt ast.Statement) int {
	switch s := stmt.(type) {
	case *ast.FuncDecl:
		return s.Line
	case *ast.ClassDecl:
		return s.Line
	case *ast.InterfaceDecl:
		return s.Line
	case *ast.IfStmt:
		return s.Line
	case *ast.WhileStmt:
		return s.Line
	case *ast.ForStmt:
		return s.Line
	case *ast.LoopStmt:
		return s.Line
	case *ast.CaseStmt:
		return s.Line
	case *ast.AssignStmt:
		return s.Line
	case *ast.ExprStmt:
		return s.Line
	case *ast.ReturnStmt:
		return s.Line
	default:
		return 0
	}
}

// formatImport formats an import statement.
func (f *Formatter) formatImport(imp *ast.ImportDecl) {
	f.formatCommentGroup(imp.Doc)
	f.write("import ")
	f.write(imp.Path)
	if imp.Alias != "" {
		f.write(" as ")
		f.write(imp.Alias)
	}
	f.formatTrailingComment(imp.Comment)
	f.write("\n")
}

// formatStatement formats a statement.
func (f *Formatter) formatStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.FuncDecl:
		f.formatFuncDecl(s)
	case *ast.ClassDecl:
		f.formatClassDecl(s)
	case *ast.InterfaceDecl:
		f.formatInterfaceDecl(s)
	case *ast.MethodDecl:
		f.formatMethodDecl(s)
	case *ast.IfStmt:
		f.formatIfStmt(s)
	case *ast.WhileStmt:
		f.formatWhileStmt(s)
	case *ast.ForStmt:
		f.formatForStmt(s)
	case *ast.LoopStmt:
		f.formatLoopStmt(s)
	case *ast.CaseStmt:
		f.formatCaseStmt(s)
	case *ast.AssignStmt:
		f.formatAssignStmt(s)
	case *ast.OrAssignStmt:
		f.formatOrAssignStmt(s)
	case *ast.CompoundAssignStmt:
		f.formatCompoundAssignStmt(s)
	case *ast.IndexAssignStmt:
		f.formatIndexAssignStmt(s)
	case *ast.IndexCompoundAssignStmt:
		f.formatIndexCompoundAssignStmt(s)
	case *ast.ExprStmt:
		f.formatExprStmt(s)
	case *ast.ReturnStmt:
		f.formatReturnStmt(s)
	case *ast.BreakStmt:
		f.formatBreakStmt(s)
	case *ast.NextStmt:
		f.formatNextStmt(s)
	case *ast.DeferStmt:
		f.formatDeferStmt(s)
	case *ast.InstanceVarAssign:
		f.formatInstanceVarAssign(s)
	case *ast.InstanceVarOrAssign:
		f.formatInstanceVarOrAssign(s)
	}
}

// formatFuncDecl formats a function declaration.
func (f *Formatter) formatFuncDecl(fn *ast.FuncDecl) {
	f.formatCommentGroup(fn.Doc)
	f.writeIndent()
	if fn.Pub {
		f.write("pub ")
	}
	f.write("def ")
	f.write(fn.Name)

	if len(fn.Params) > 0 {
		f.write("(")
		f.formatParams(fn.Params)
		f.write(")")
	}

	if len(fn.ReturnTypes) > 0 {
		f.write(": ")
		f.write(strings.Join(fn.ReturnTypes, ", "))
	}

	f.formatTrailingComment(fn.Comment)
	f.write("\n")

	f.indent++
	f.formatBody(fn.Body)
	f.indent--

	f.writeIndent()
	f.writeLine("end")
}

// formatClassDecl formats a class declaration.
func (f *Formatter) formatClassDecl(cls *ast.ClassDecl) {
	f.formatCommentGroup(cls.Doc)
	f.writeIndent()
	if cls.Pub {
		f.write("pub ")
	}
	f.write("class ")
	f.write(cls.Name)

	if len(cls.Embeds) > 0 {
		f.write(" < ")
		f.write(strings.Join(cls.Embeds, ", "))
	}

	f.formatTrailingComment(cls.Comment)
	f.write("\n")

	f.indent++

	// Format field declarations
	for _, field := range cls.Fields {
		f.formatFieldDecl(field)
	}

	// Add blank line between fields and methods if both exist
	if len(cls.Fields) > 0 && len(cls.Methods) > 0 {
		f.write("\n")
	}

	for _, method := range cls.Methods {
		f.formatMethodDecl(method)
	}
	f.indent--

	f.writeIndent()
	f.writeLine("end")
}

// formatFieldDecl formats a field declaration.
// Only outputs fields with explicit types (inferred fields without types are not formatted).
func (f *Formatter) formatFieldDecl(field *ast.FieldDecl) {
	// Skip fields with no type (they were inferred and shouldn't be explicitly formatted)
	if field.Type == "" {
		return
	}
	f.writeIndent()
	f.write("@")
	f.write(field.Name)
	f.write(" : ")
	f.write(field.Type)
	f.write("\n")
}

// formatInterfaceDecl formats an interface declaration.
func (f *Formatter) formatInterfaceDecl(iface *ast.InterfaceDecl) {
	f.formatCommentGroup(iface.Doc)
	f.writeIndent()
	if iface.Pub {
		f.write("pub ")
	}
	f.write("interface ")
	f.write(iface.Name)
	f.formatTrailingComment(iface.Comment)
	f.write("\n")

	f.indent++
	for _, sig := range iface.Methods {
		f.formatMethodSig(sig)
	}
	f.indent--

	f.writeIndent()
	f.writeLine("end")
}

// formatMethodDecl formats a method declaration.
func (f *Formatter) formatMethodDecl(m *ast.MethodDecl) {
	f.formatCommentGroup(m.Doc)
	f.writeIndent()
	if m.Pub {
		f.write("pub ")
	}
	f.write("def ")
	f.write(m.Name)

	if len(m.Params) > 0 {
		f.write("(")
		f.formatParams(m.Params)
		f.write(")")
	}

	if len(m.ReturnTypes) > 0 {
		f.write(": ")
		f.write(strings.Join(m.ReturnTypes, ", "))
	}

	f.formatTrailingComment(m.Comment)
	f.write("\n")

	f.indent++
	f.formatBody(m.Body)
	f.indent--

	f.writeIndent()
	f.writeLine("end")
}

// formatMethodSig formats a method signature (in an interface).
func (f *Formatter) formatMethodSig(sig *ast.MethodSig) {
	f.writeIndent()
	f.write("def ")
	f.write(sig.Name)

	if len(sig.Params) > 0 {
		f.write("(")
		f.formatParams(sig.Params)
		f.write(")")
	}

	if len(sig.ReturnTypes) > 0 {
		f.write(": ")
		f.write(strings.Join(sig.ReturnTypes, ", "))
	}
	f.write("\n")
}

// formatParams formats a parameter list.
func (f *Formatter) formatParams(params []*ast.Param) {
	for i, p := range params {
		if i > 0 {
			f.write(", ")
		}
		f.write(p.Name)
		if p.Type != "" {
			f.write(": ")
			f.write(p.Type)
		}
	}
}

// formatBody formats a list of statements as a body.
func (f *Formatter) formatBody(stmts []ast.Statement) {
	for _, stmt := range stmts {
		line := f.getStatementLine(stmt)
		if line > 0 {
			f.emitFreeFloatingComments(line)
		}
		f.formatStatement(stmt)
		if line > 0 {
			f.lastLine = line
		}
	}
}

// formatIfStmt formats an if statement.
func (f *Formatter) formatIfStmt(s *ast.IfStmt) {
	f.formatCommentGroup(s.Doc)
	f.writeIndent()
	if s.IsUnless {
		f.write("unless ")
	} else {
		f.write("if ")
	}

	if s.AssignName != "" {
		f.write("(")
		f.write(s.AssignName)
		f.write(" = ")
		f.formatExpr(s.AssignExpr)
		f.write(")")
	} else {
		f.formatExpr(s.Cond)
	}

	f.formatTrailingComment(s.Comment)
	f.write("\n")

	f.indent++
	f.formatBody(s.Then)
	f.indent--

	for _, elsif := range s.ElseIfs {
		f.writeIndent()
		f.write("elsif ")
		f.formatExpr(elsif.Cond)
		f.write("\n")
		f.indent++
		f.formatBody(elsif.Body)
		f.indent--
	}

	if len(s.Else) > 0 {
		f.writeIndent()
		f.writeLine("else")
		f.indent++
		f.formatBody(s.Else)
		f.indent--
	}

	f.writeIndent()
	f.writeLine("end")
}

// formatWhileStmt formats a while statement.
func (f *Formatter) formatWhileStmt(s *ast.WhileStmt) {
	f.formatCommentGroup(s.Doc)
	f.writeIndent()
	f.write("while ")
	f.formatExpr(s.Cond)
	f.formatTrailingComment(s.Comment)
	f.write("\n")

	f.indent++
	f.formatBody(s.Body)
	f.indent--

	f.writeIndent()
	f.writeLine("end")
}

// formatForStmt formats a for statement.
func (f *Formatter) formatForStmt(s *ast.ForStmt) {
	f.formatCommentGroup(s.Doc)
	f.writeIndent()
	f.write("for ")
	f.write(s.Var)
	f.write(" in ")
	f.formatExpr(s.Iterable)
	f.formatTrailingComment(s.Comment)
	f.write("\n")

	f.indent++
	f.formatBody(s.Body)
	f.indent--

	f.writeIndent()
	f.writeLine("end")
}

// formatLoopStmt formats an infinite loop statement.
func (f *Formatter) formatLoopStmt(s *ast.LoopStmt) {
	f.formatCommentGroup(s.Doc)
	f.writeIndent()
	f.write("loop do")
	f.formatTrailingComment(s.Comment)
	f.write("\n")

	f.indent++
	f.formatBody(s.Body)
	f.indent--

	f.writeIndent()
	f.writeLine("end")
}

// formatCaseStmt formats a case statement.
func (f *Formatter) formatCaseStmt(s *ast.CaseStmt) {
	f.formatCommentGroup(s.Doc)
	f.writeIndent()
	f.write("case")
	if s.Subject != nil {
		f.write(" ")
		f.formatExpr(s.Subject)
	}
	f.formatTrailingComment(s.Comment)
	f.write("\n")

	for _, when := range s.WhenClauses {
		f.writeIndent()
		f.write("when ")
		for i, v := range when.Values {
			if i > 0 {
				f.write(", ")
			}
			f.formatExpr(v)
		}
		f.write("\n")
		f.indent++
		f.formatBody(when.Body)
		f.indent--
	}

	if len(s.Else) > 0 {
		f.writeIndent()
		f.writeLine("else")
		f.indent++
		f.formatBody(s.Else)
		f.indent--
	}

	f.writeIndent()
	f.writeLine("end")
}

// formatAssignStmt formats an assignment statement.
func (f *Formatter) formatAssignStmt(s *ast.AssignStmt) {
	f.formatCommentGroup(s.Doc)
	f.writeIndent()
	f.write(s.Name)
	if s.Type != "" {
		f.write(": ")
		f.write(s.Type)
	}
	f.write(" = ")
	f.formatExpr(s.Value)
	f.formatTrailingComment(s.Comment)
	f.write("\n")
}

// formatOrAssignStmt formats an or-assignment statement.
func (f *Formatter) formatOrAssignStmt(s *ast.OrAssignStmt) {
	f.writeIndent()
	f.write(s.Name)
	f.write(" ||= ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// formatCompoundAssignStmt formats a compound assignment statement.
func (f *Formatter) formatCompoundAssignStmt(s *ast.CompoundAssignStmt) {
	f.writeIndent()
	f.write(s.Name)
	f.write(" ")
	f.write(s.Op)
	f.write("= ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// formatIndexAssignStmt formats an index assignment statement.
func (f *Formatter) formatIndexAssignStmt(s *ast.IndexAssignStmt) {
	f.writeIndent()
	f.formatExpr(s.Left)
	f.write("[")
	f.formatExpr(s.Index)
	f.write("] = ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// formatIndexCompoundAssignStmt formats an index compound assignment statement.
func (f *Formatter) formatIndexCompoundAssignStmt(s *ast.IndexCompoundAssignStmt) {
	f.writeIndent()
	f.formatExpr(s.Left)
	f.write("[")
	f.formatExpr(s.Index)
	f.write("] ")
	f.write(s.Op)
	f.write("= ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// formatExprStmt formats an expression statement.
func (f *Formatter) formatExprStmt(s *ast.ExprStmt) {
	f.formatCommentGroup(s.Doc)
	f.writeIndent()
	f.formatExpr(s.Expr)
	if s.Condition != nil {
		if s.IsUnless {
			f.write(" unless ")
		} else {
			f.write(" if ")
		}
		f.formatExpr(s.Condition)
	}
	f.formatTrailingComment(s.Comment)
	f.write("\n")
}

// formatReturnStmt formats a return statement.
func (f *Formatter) formatReturnStmt(s *ast.ReturnStmt) {
	f.formatCommentGroup(s.Doc)
	f.writeIndent()
	f.write("return")
	if len(s.Values) > 0 {
		f.write(" ")
		for i, v := range s.Values {
			if i > 0 {
				f.write(", ")
			}
			f.formatExpr(v)
		}
	}
	if s.Condition != nil {
		if s.IsUnless {
			f.write(" unless ")
		} else {
			f.write(" if ")
		}
		f.formatExpr(s.Condition)
	}
	f.formatTrailingComment(s.Comment)
	f.write("\n")
}

// formatBreakStmt formats a break statement.
func (f *Formatter) formatBreakStmt(s *ast.BreakStmt) {
	f.writeIndent()
	f.write("break")
	if s.Condition != nil {
		if s.IsUnless {
			f.write(" unless ")
		} else {
			f.write(" if ")
		}
		f.formatExpr(s.Condition)
	}
	f.write("\n")
}

// formatNextStmt formats a next statement.
func (f *Formatter) formatNextStmt(s *ast.NextStmt) {
	f.writeIndent()
	f.write("next")
	if s.Condition != nil {
		if s.IsUnless {
			f.write(" unless ")
		} else {
			f.write(" if ")
		}
		f.formatExpr(s.Condition)
	}
	f.write("\n")
}

// formatDeferStmt formats a defer statement.
func (f *Formatter) formatDeferStmt(s *ast.DeferStmt) {
	f.writeIndent()
	f.write("defer ")
	f.formatExpr(s.Call)
	f.write("\n")
}

// formatInstanceVarAssign formats an instance variable assignment.
func (f *Formatter) formatInstanceVarAssign(s *ast.InstanceVarAssign) {
	f.writeIndent()
	f.write("@")
	f.write(s.Name)
	f.write(" = ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// formatInstanceVarOrAssign formats an instance variable or-assignment.
func (f *Formatter) formatInstanceVarOrAssign(s *ast.InstanceVarOrAssign) {
	f.writeIndent()
	f.write("@")
	f.write(s.Name)
	f.write(" ||= ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// formatExpr formats an expression.
func (f *Formatter) formatExpr(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.Ident:
		f.write(e.Name)
	case *ast.IntLit:
		f.write(fmt.Sprintf("%d", e.Value))
	case *ast.FloatLit:
		f.write(fmt.Sprintf("%g", e.Value))
	case *ast.BoolLit:
		if e.Value {
			f.write("true")
		} else {
			f.write("false")
		}
	case *ast.NilLit:
		f.write("nil")
	case *ast.StringLit:
		f.write(`"`)
		f.write(escapeString(e.Value))
		f.write(`"`)
	case *ast.InterpolatedString:
		f.formatInterpolatedString(e)
	case *ast.SymbolLit:
		f.write(":")
		f.write(e.Value)
	case *ast.ArrayLit:
		f.formatArrayLit(e)
	case *ast.MapLit:
		f.formatMapLit(e)
	case *ast.RangeLit:
		f.formatExpr(e.Start)
		if e.Exclusive {
			f.write("...")
		} else {
			f.write("..")
		}
		f.formatExpr(e.End)
	case *ast.BinaryExpr:
		f.formatBinaryExpr(e)
	case *ast.UnaryExpr:
		f.formatUnaryExpr(e)
	case *ast.CallExpr:
		f.formatCallExpr(e)
	case *ast.SelectorExpr:
		f.formatExpr(e.X)
		f.write(".")
		f.write(e.Sel)
	case *ast.IndexExpr:
		f.formatExpr(e.Left)
		f.write("[")
		f.formatExpr(e.Index)
		f.write("]")
	case *ast.InstanceVar:
		f.write("@")
		f.write(e.Name)
	case *ast.BlockExpr:
		f.formatBlockExpr(e)
	}
}

// formatInterpolatedString formats an interpolated string.
func (f *Formatter) formatInterpolatedString(s *ast.InterpolatedString) {
	f.write(`"`)
	for _, part := range s.Parts {
		switch p := part.(type) {
		case string:
			f.write(escapeString(p))
		case ast.Expression:
			f.write("#{")
			f.formatExpr(p)
			f.write("}")
		}
	}
	f.write(`"`)
}

// formatArrayLit formats an array literal.
func (f *Formatter) formatArrayLit(a *ast.ArrayLit) {
	f.write("[")
	for i, elem := range a.Elements {
		if i > 0 {
			f.write(", ")
		}
		f.formatExpr(elem)
	}
	f.write("]")
}

// formatMapLit formats a map literal.
func (f *Formatter) formatMapLit(m *ast.MapLit) {
	f.write("{")
	for i, entry := range m.Entries {
		if i > 0 {
			f.write(", ")
		}
		f.formatExpr(entry.Key)
		f.write(" => ")
		f.formatExpr(entry.Value)
	}
	f.write("}")
}

// formatBinaryExpr formats a binary expression with proper parentheses for precedence.
func (f *Formatter) formatBinaryExpr(e *ast.BinaryExpr) {
	// Check if left operand needs parentheses
	if left, ok := e.Left.(*ast.BinaryExpr); ok {
		if precedence(left.Op) < precedence(e.Op) {
			f.write("(")
			f.formatExpr(e.Left)
			f.write(")")
		} else {
			f.formatExpr(e.Left)
		}
	} else {
		f.formatExpr(e.Left)
	}

	f.write(" ")
	f.write(e.Op)
	f.write(" ")

	// Check if right operand needs parentheses
	if right, ok := e.Right.(*ast.BinaryExpr); ok {
		if precedence(right.Op) <= precedence(e.Op) {
			f.write("(")
			f.formatExpr(e.Right)
			f.write(")")
		} else {
			f.formatExpr(e.Right)
		}
	} else {
		f.formatExpr(e.Right)
	}
}

// precedence returns the precedence level of an operator.
func precedence(op string) int {
	switch op {
	case "or":
		return 1
	case "and":
		return 2
	case "==", "!=":
		return 3
	case "<", ">", "<=", ">=":
		return 4
	case "+", "-":
		return 5
	case "*", "/", "%":
		return 6
	case "..", "...":
		return 0 // lowest
	default:
		return 7
	}
}

// formatUnaryExpr formats a unary expression.
func (f *Formatter) formatUnaryExpr(e *ast.UnaryExpr) {
	if e.Op == "not" {
		f.write("not ")
	} else {
		f.write(e.Op)
	}
	f.formatExpr(e.Expr)
}

// formatCallExpr formats a call expression.
func (f *Formatter) formatCallExpr(c *ast.CallExpr) {
	f.formatExpr(c.Func)
	if len(c.Args) > 0 || c.Block == nil {
		f.write("(")
		for i, arg := range c.Args {
			if i > 0 {
				f.write(", ")
			}
			f.formatExpr(arg)
		}
		f.write(")")
	}
	if c.Block != nil {
		f.write(" ")
		f.formatBlockExpr(c.Block)
	}
}

// formatBlockExpr formats a block expression.
func (f *Formatter) formatBlockExpr(b *ast.BlockExpr) {
	// Use do...end for multi-line, {} for single-line
	if len(b.Body) <= 1 && !hasNestedBlocks(b.Body) {
		f.write("{")
		if len(b.Params) > 0 {
			f.write("|")
			f.write(strings.Join(b.Params, ", "))
			f.write("| ")
		}
		if len(b.Body) == 1 {
			// Format single statement without indent or newline
			f.formatInlineStatement(b.Body[0])
		}
		f.write("}")
	} else {
		f.write("do")
		if len(b.Params) > 0 {
			f.write(" |")
			f.write(strings.Join(b.Params, ", "))
			f.write("|")
		}
		f.write("\n")
		f.indent++
		f.formatBody(b.Body)
		f.indent--
		f.writeIndent()
		f.write("end")
	}
}

// formatInlineStatement formats a statement for inline use (no indentation).
func (f *Formatter) formatInlineStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		f.formatExpr(s.Expr)
	case *ast.ReturnStmt:
		f.write("return")
		if len(s.Values) > 0 {
			f.write(" ")
			for i, v := range s.Values {
				if i > 0 {
					f.write(", ")
				}
				f.formatExpr(v)
			}
		}
	default:
		// Fall back to regular formatting
		f.formatStatement(stmt)
	}
}

// hasNestedBlocks checks if any statements contain nested blocks.
func hasNestedBlocks(stmts []ast.Statement) bool {
	for _, stmt := range stmts {
		if s, ok := stmt.(*ast.ExprStmt); ok {
			if c, ok := s.Expr.(*ast.CallExpr); ok {
				if c.Block != nil {
					return true
				}
			}
		}
	}
	return false
}

// escapeString escapes special characters in a string.
func escapeString(s string) string {
	var buf strings.Builder
	for _, r := range s {
		switch r {
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
