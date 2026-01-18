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
	// Phase 2: Type Declarations
	case *ast.StructDecl:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
		for _, method := range s.Methods {
			f.markCommentGroupEmitted(method.Doc)
			f.markCommentGroupEmitted(method.Comment)
			for _, bodyStmt := range method.Body {
				f.preMarkStatementComments(bodyStmt)
			}
		}
	case *ast.EnumDecl:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
	case *ast.ConstDecl:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
	case *ast.TypeAliasDecl:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
	case *ast.ModuleDecl:
		f.markCommentGroupEmitted(s.Doc)
		for _, method := range s.Methods {
			f.markCommentGroupEmitted(method.Doc)
			f.markCommentGroupEmitted(method.Comment)
			for _, bodyStmt := range method.Body {
				f.preMarkStatementComments(bodyStmt)
			}
		}
		for _, cls := range s.Classes {
			f.preMarkStatementComments(cls)
		}
	// Phase 4: Control Flow
	case *ast.UntilStmt:
		f.markCommentGroupEmitted(s.Doc)
		f.markCommentGroupEmitted(s.Comment)
		for _, bodyStmt := range s.Body {
			f.preMarkStatementComments(bodyStmt)
		}
	case *ast.CaseTypeStmt:
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
	case *ast.SelectStmt:
		for _, c := range s.Cases {
			for _, bodyStmt := range c.Body {
				f.preMarkStatementComments(bodyStmt)
			}
		}
		for _, bodyStmt := range s.Else {
			f.preMarkStatementComments(bodyStmt)
		}
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
	// Phase 2: Type Declarations
	case *ast.StructDecl:
		return s.Line
	case *ast.EnumDecl:
		return s.Line
	case *ast.ConstDecl:
		return s.Line
	case *ast.TypeAliasDecl:
		return s.Line
	case *ast.ModuleDecl:
		return s.Line
	// Phase 4: Control Flow
	case *ast.UntilStmt:
		return s.Line
	case *ast.CaseTypeStmt:
		return s.Line
	case *ast.MultiAssignStmt:
		return s.Line
	case *ast.MapDestructuringStmt:
		return s.Line
	case *ast.SelectStmt:
		return s.Line
	// Phase 5: Class Features
	case *ast.AccessorDecl:
		return s.Line
	case *ast.ClassVarAssign:
		return s.Line
	case *ast.ClassVarCompoundAssign:
		return s.Line
	case *ast.IncludeStmt:
		return s.Line
	case *ast.InstanceVarAssign:
		return s.Line
	case *ast.InstanceVarOrAssign:
		return s.Line
	case *ast.InstanceVarCompoundAssign:
		return s.Line
	case *ast.SelectorAssignStmt:
		return s.Line
	case *ast.SelectorCompoundAssign:
		return s.Line
	// Phase 6: Concurrency
	case *ast.GoStmt:
		return s.Line
	case *ast.ChanSendStmt:
		return s.Line
	// Additional statements
	case *ast.OrAssignStmt:
		return s.Line
	case *ast.CompoundAssignStmt:
		return s.Line
	case *ast.IndexAssignStmt:
		return s.Line
	case *ast.IndexCompoundAssignStmt:
		return s.Line
	case *ast.BreakStmt:
		return s.Line
	case *ast.NextStmt:
		return s.Line
	case *ast.DeferStmt:
		return s.Line
	case *ast.PanicStmt:
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
	// Phase 2: Type Declarations
	case *ast.StructDecl:
		f.formatStructDecl(s)
	case *ast.EnumDecl:
		f.formatEnumDecl(s)
	case *ast.ConstDecl:
		f.formatConstDecl(s)
	case *ast.TypeAliasDecl:
		f.formatTypeAliasDecl(s)
	case *ast.ModuleDecl:
		f.formatModuleDecl(s)
	// Phase 4: Control Flow
	case *ast.UntilStmt:
		f.formatUntilStmt(s)
	case *ast.CaseTypeStmt:
		f.formatCaseTypeStmt(s)
	case *ast.MultiAssignStmt:
		f.formatMultiAssignStmt(s)
	case *ast.MapDestructuringStmt:
		f.formatMapDestructuringStmt(s)
	case *ast.SelectStmt:
		f.formatSelectStmt(s)
	// Phase 5: Class Features
	case *ast.AccessorDecl:
		f.formatAccessorDecl(s)
	case *ast.ClassVarAssign:
		f.formatClassVarAssign(s)
	case *ast.ClassVarCompoundAssign:
		f.formatClassVarCompoundAssign(s)
	case *ast.IncludeStmt:
		f.formatIncludeStmt(s)
	case *ast.InstanceVarCompoundAssign:
		f.formatInstanceVarCompoundAssign(s)
	case *ast.SelectorAssignStmt:
		f.formatSelectorAssignStmt(s)
	case *ast.SelectorCompoundAssign:
		f.formatSelectorCompoundAssign(s)
	// Phase 6: Concurrency
	case *ast.GoStmt:
		f.formatGoStmt(s)
	case *ast.ChanSendStmt:
		f.formatChanSendStmt(s)
	// Additional statements
	case *ast.PanicStmt:
		f.formatPanicStmt(s)
	default:
		// Unhandled statement type - write a comment to make issue visible
		f.writeIndent()
		f.write(fmt.Sprintf("# unformatted: %T\n", stmt))
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
	f.formatTypeParams(fn.TypeParams)

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
	f.formatTypeParams(cls.TypeParams)

	if len(cls.Embeds) > 0 {
		f.write(" < ")
		f.write(strings.Join(cls.Embeds, ", "))
	}

	if len(cls.Implements) > 0 {
		f.write(" implements ")
		f.write(strings.Join(cls.Implements, ", "))
	}

	f.formatTrailingComment(cls.Comment)
	f.write("\n")

	f.indent++

	// Format includes
	for _, inc := range cls.Includes {
		f.writeIndent()
		f.write("include ")
		f.write(inc)
		f.write("\n")
	}

	// Format accessors
	for _, acc := range cls.Accessors {
		f.formatAccessorDecl(acc)
	}

	// Format class variables
	for _, cv := range cls.ClassVars {
		f.writeIndent()
		f.write("@@")
		f.write(cv.Name)
		f.write(" = ")
		f.formatExpr(cv.Value)
		f.write("\n")
	}

	// Format field declarations
	for _, field := range cls.Fields {
		f.formatFieldDecl(field)
	}

	// Add blank line between fields/accessors and methods if both exist
	hasPreContent := len(cls.Includes) > 0 || len(cls.Accessors) > 0 || len(cls.ClassVars) > 0 || len(cls.Fields) > 0
	if hasPreContent && len(cls.Methods) > 0 {
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
	f.formatTypeParams(iface.TypeParams)

	// Format parent interfaces for embedding
	if len(iface.Parents) > 0 {
		f.write(" < ")
		f.write(strings.Join(iface.Parents, ", "))
	}

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
	if m.Private {
		f.write("private ")
	} else if m.Pub {
		f.write("pub ")
	}
	f.write("def ")
	if m.IsClassMethod {
		f.write("self.")
	}
	f.write(m.Name)
	f.formatTypeParams(m.TypeParams)

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
		// Handle pattern destructuring: {name:, age:} : Type
		if len(p.DestructurePairs) > 0 {
			f.write("{")
			for j, pair := range p.DestructurePairs {
				if j > 0 {
					f.write(", ")
				}
				f.write(pair.Key)
				if pair.Variable != "" && pair.Variable != pair.Key {
					f.write(": ")
					f.write(pair.Variable)
				} else {
					f.write(":")
				}
			}
			f.write("}")
			if p.Type != "" {
				f.write(": ")
				f.write(p.Type)
			}
			continue
		}
		if p.Variadic {
			f.write("*")
		}
		f.write(p.Name)
		if p.Type != "" {
			f.write(": ")
			f.write(p.Type)
		}
		if p.DefaultValue != nil {
			f.write(" = ")
			f.formatExpr(p.DefaultValue)
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
		// if-let syntax: if let name = expr
		f.write("let ")
		f.write(s.AssignName)
		f.write(" = ")
		f.formatExpr(s.AssignExpr)
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
	if s.Var2 != "" {
		f.write(", ")
		f.write(s.Var2)
	}
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
	// Phase 1: Core Expressions
	case *ast.LambdaExpr:
		f.formatLambdaExpr(e)
	case *ast.BangExpr:
		f.formatExpr(e.Expr)
		f.write("!")
	case *ast.NilCoalesceExpr:
		f.formatExpr(e.Left)
		f.write(" ?? ")
		f.formatExpr(e.Right)
	case *ast.RescueExpr:
		f.formatRescueExpr(e)
	case *ast.SafeNavExpr:
		f.formatExpr(e.Receiver)
		f.write("&.")
		f.write(e.Selector)
	case *ast.TernaryExpr:
		f.formatExpr(e.Condition)
		f.write(" ? ")
		f.formatExpr(e.Then)
		f.write(" : ")
		f.formatExpr(e.Else)
	// Phase 6: Concurrency expressions
	case *ast.SpawnExpr:
		f.write("spawn ")
		f.formatBlockExpr(e.Block)
	case *ast.AwaitExpr:
		f.write("await ")
		f.formatExpr(e.Task)
	// Phase 7: Additional expressions
	case *ast.StructLit:
		f.formatStructLit(e)
	case *ast.SetLit:
		f.formatSetLit(e)
	case *ast.RegexLit:
		f.write("/")
		f.write(e.Pattern)
		f.write("/")
		f.write(e.Flags)
	case *ast.SplatExpr:
		f.write("*")
		f.formatExpr(e.Expr)
	case *ast.DoubleSplatExpr:
		f.write("**")
		f.formatExpr(e.Expr)
	case *ast.KeywordArg:
		f.write(e.Name)
		f.write(": ")
		f.formatExpr(e.Value)
	case *ast.ScopeExpr:
		f.formatExpr(e.Left)
		f.write("::")
		f.write(e.Right)
	case *ast.SuperExpr:
		f.write("super")
		if len(e.Args) > 0 {
			f.write("(")
			for i, arg := range e.Args {
				if i > 0 {
					f.write(", ")
				}
				f.formatExpr(arg)
			}
			f.write(")")
		}
	case *ast.TupleLit:
		for i, elem := range e.Elements {
			if i > 0 {
				f.write(", ")
			}
			f.formatExpr(elem)
		}
	case *ast.ClassVar:
		f.write("@@")
		f.write(e.Name)
	case *ast.SymbolToProcExpr:
		f.write("&:")
		f.write(e.Method)
	case *ast.GoStructLit:
		if e.Package != "" {
			f.write(e.Package)
			f.write(".")
		}
		f.write(e.Type)
		f.write("{}")
	default:
		// Unhandled expression type - write a placeholder
		f.write(fmt.Sprintf("/* unformatted: %T */", expr))
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
	case "or", "||":
		return 1
	case "and", "&&":
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
	case *ast.BreakStmt:
		f.write("break")
	case *ast.NextStmt:
		f.write("next")
	case *ast.AssignStmt:
		f.write(s.Name)
		if s.Type != "" {
			f.write(": ")
			f.write(s.Type)
		}
		f.write(" = ")
		f.formatExpr(s.Value)
	default:
		// For inline blocks, unsupported statements get a placeholder
		// This shouldn't happen in well-formed code
		f.write(fmt.Sprintf("/* inline: %T */", stmt))
	}
}

// hasNestedBlocks checks if any statements contain nested blocks.
func hasNestedBlocks(stmts []ast.Statement) bool {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.IfStmt, *ast.WhileStmt, *ast.ForStmt, *ast.CaseStmt,
			*ast.LoopStmt, *ast.UntilStmt, *ast.CaseTypeStmt, *ast.SelectStmt,
			*ast.ConcurrentlyStmt:
			// These statements always require multi-line block format
			return true
		case *ast.ExprStmt:
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

// formatLambdaExpr formats a lambda expression.
func (f *Formatter) formatLambdaExpr(l *ast.LambdaExpr) {
	// Decide between {} (single expression) and do...end (multi-line)
	isMultiLine := len(l.Body) > 1 || (len(l.Body) == 1 && hasNestedBlocks(l.Body))

	if isMultiLine {
		f.write("-> do")
		if len(l.Params) > 0 {
			f.write(" |")
			f.formatLambdaParams(l.Params)
			f.write("|")
		}
		f.write("\n")
		f.indent++
		f.formatBody(l.Body)
		f.indent--
		f.writeIndent()
		f.write("end")
	} else {
		f.write("-> {")
		if len(l.Params) > 0 {
			f.write(" |")
			f.formatLambdaParams(l.Params)
			f.write("|")
		}
		if l.ReturnType != "" {
			f.write(": ")
			f.write(l.ReturnType)
		}
		if len(l.Body) == 1 {
			f.write(" ")
			f.formatInlineStatement(l.Body[0])
		}
		f.write(" }")
	}
}

// formatLambdaParams formats lambda parameters with types.
func (f *Formatter) formatLambdaParams(params []*ast.Param) {
	for i, p := range params {
		if i > 0 {
			f.write(", ")
		}
		if p.Variadic {
			f.write("*")
		}
		f.write(p.Name)
		if p.Type != "" {
			f.write(": ")
			f.write(p.Type)
		}
		if p.DefaultValue != nil {
			f.write(" = ")
			f.formatExpr(p.DefaultValue)
		}
	}
}

// formatRescueExpr formats a rescue expression.
func (f *Formatter) formatRescueExpr(r *ast.RescueExpr) {
	f.formatExpr(r.Expr)
	f.write(" rescue ")
	if r.Block != nil {
		// Block form: rescue => err do ... end
		if r.ErrName != "" {
			f.write("=> ")
			f.write(r.ErrName)
			f.write(" ")
		}
		f.formatBlockExpr(r.Block)
	} else {
		// Inline form: rescue default
		f.formatExpr(r.Default)
	}
}

// formatStructLit formats a struct literal.
func (f *Formatter) formatStructLit(s *ast.StructLit) {
	f.write(s.Name)
	f.write("{")
	for i, field := range s.Fields {
		if i > 0 {
			f.write(", ")
		}
		f.write(field.Name)
		f.write(": ")
		f.formatExpr(field.Value)
	}
	f.write("}")
}

// formatSetLit formats a set literal.
func (f *Formatter) formatSetLit(s *ast.SetLit) {
	f.write("Set")
	if s.TypeHint != "" {
		f.write("<")
		f.write(s.TypeHint)
		f.write(">")
	}
	f.write("{")
	for i, elem := range s.Elements {
		if i > 0 {
			f.write(", ")
		}
		f.formatExpr(elem)
	}
	f.write("}")
}

// formatTypeParams formats type parameters for generics.
func (f *Formatter) formatTypeParams(typeParams []*ast.TypeParam) {
	if len(typeParams) == 0 {
		return
	}
	f.write("<")
	for i, tp := range typeParams {
		if i > 0 {
			f.write(", ")
		}
		f.write(tp.Name)
		if tp.Constraint != "" {
			f.write(": ")
			f.write(tp.Constraint)
		}
	}
	f.write(">")
}

// Phase 2: Type Declaration formatters

// formatStructDecl formats a struct declaration.
func (f *Formatter) formatStructDecl(s *ast.StructDecl) {
	f.formatCommentGroup(s.Doc)
	f.writeIndent()
	if s.Pub {
		f.write("pub ")
	}
	f.write("struct ")
	f.write(s.Name)
	f.formatTypeParams(s.TypeParams)
	f.formatTrailingComment(s.Comment)
	f.write("\n")

	f.indent++
	for _, field := range s.Fields {
		f.writeIndent()
		f.write(field.Name)
		f.write(": ")
		f.write(field.Type)
		f.write("\n")
	}

	// Add blank line between fields and methods if both exist
	if len(s.Fields) > 0 && len(s.Methods) > 0 {
		f.write("\n")
	}

	for _, method := range s.Methods {
		f.formatMethodDecl(method)
	}
	f.indent--

	f.writeIndent()
	f.writeLine("end")
}

// formatEnumDecl formats an enum declaration.
func (f *Formatter) formatEnumDecl(e *ast.EnumDecl) {
	f.formatCommentGroup(e.Doc)
	f.writeIndent()
	if e.Pub {
		f.write("pub ")
	}
	f.write("enum ")
	f.write(e.Name)
	f.formatTrailingComment(e.Comment)
	f.write("\n")

	f.indent++
	for _, val := range e.Values {
		f.writeIndent()
		f.write(val.Name)
		if val.Value != nil {
			f.write(" = ")
			f.formatExpr(val.Value)
		}
		f.write("\n")
	}
	f.indent--

	f.writeIndent()
	f.writeLine("end")
}

// formatConstDecl formats a constant declaration.
func (f *Formatter) formatConstDecl(c *ast.ConstDecl) {
	f.formatCommentGroup(c.Doc)
	f.writeIndent()
	f.write("const ")
	f.write(c.Name)
	if c.Type != "" {
		f.write(": ")
		f.write(c.Type)
	}
	f.write(" = ")
	f.formatExpr(c.Value)
	f.formatTrailingComment(c.Comment)
	f.write("\n")
}

// formatTypeAliasDecl formats a type alias declaration.
func (f *Formatter) formatTypeAliasDecl(t *ast.TypeAliasDecl) {
	f.formatCommentGroup(t.Doc)
	f.writeIndent()
	if t.Pub {
		f.write("pub ")
	}
	f.write("type ")
	f.write(t.Name)
	f.write(" = ")
	f.write(t.Type)
	f.formatTrailingComment(t.Comment)
	f.write("\n")
}

// formatModuleDecl formats a module declaration.
func (f *Formatter) formatModuleDecl(m *ast.ModuleDecl) {
	f.formatCommentGroup(m.Doc)
	f.writeIndent()
	if m.Pub {
		f.write("pub ")
	}
	f.write("module ")
	f.write(m.Name)
	f.write("\n")

	f.indent++

	// Format accessors
	for _, acc := range m.Accessors {
		f.formatAccessorDecl(acc)
	}

	// Format methods
	for _, method := range m.Methods {
		f.formatMethodDecl(method)
	}

	// Format nested classes
	for _, cls := range m.Classes {
		f.formatClassDecl(cls)
	}
	f.indent--

	f.writeIndent()
	f.writeLine("end")
}

// Phase 4: Control Flow formatters

// formatUntilStmt formats an until statement.
func (f *Formatter) formatUntilStmt(s *ast.UntilStmt) {
	f.formatCommentGroup(s.Doc)
	f.writeIndent()
	f.write("until ")
	f.formatExpr(s.Cond)
	f.formatTrailingComment(s.Comment)
	f.write("\n")

	f.indent++
	f.formatBody(s.Body)
	f.indent--

	f.writeIndent()
	f.writeLine("end")
}

// formatCaseTypeStmt formats a case_type statement.
func (f *Formatter) formatCaseTypeStmt(s *ast.CaseTypeStmt) {
	f.formatCommentGroup(s.Doc)
	f.writeIndent()
	f.write("case_type ")
	f.formatExpr(s.Subject)
	f.formatTrailingComment(s.Comment)
	f.write("\n")

	for _, when := range s.WhenClauses {
		f.writeIndent()
		f.write("when ")
		if when.BindingVar != "" {
			f.write(when.BindingVar)
			f.write(": ")
		}
		f.write(when.Type)
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

// formatMultiAssignStmt formats a multi-assignment statement.
func (f *Formatter) formatMultiAssignStmt(s *ast.MultiAssignStmt) {
	f.writeIndent()
	for i, name := range s.Names {
		if i > 0 {
			f.write(", ")
		}
		if s.SplatIndex >= 0 && i == s.SplatIndex {
			f.write("*")
		}
		f.write(name)
	}
	f.write(" = ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// formatMapDestructuringStmt formats a map destructuring statement.
func (f *Formatter) formatMapDestructuringStmt(s *ast.MapDestructuringStmt) {
	f.writeIndent()
	f.write("{")
	for i, pair := range s.Pairs {
		if i > 0 {
			f.write(", ")
		}
		f.write(pair.Key)
		if pair.Variable != "" && pair.Variable != pair.Key {
			f.write(": ")
			f.write(pair.Variable)
		} else {
			f.write(":")
		}
	}
	f.write("} = ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// formatSelectStmt formats a select statement.
func (f *Formatter) formatSelectStmt(s *ast.SelectStmt) {
	f.writeIndent()
	f.writeLine("select")

	for _, c := range s.Cases {
		f.writeIndent()
		f.write("when ")
		if c.IsSend {
			// Send case: ch << value
			f.formatExpr(c.Chan)
			f.write(" << ")
			f.formatExpr(c.Value)
		} else {
			// Receive case: val = ch.receive
			if c.AssignName != "" {
				f.write(c.AssignName)
				f.write(" = ")
			}
			f.formatExpr(c.Chan)
			f.write(".receive")
		}
		f.write("\n")
		f.indent++
		f.formatBody(c.Body)
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

// Phase 5: Class Features formatters

// formatAccessorDecl formats an accessor declaration.
func (f *Formatter) formatAccessorDecl(a *ast.AccessorDecl) {
	f.writeIndent()
	if a.Pub {
		f.write("pub ")
	}
	f.write(a.Kind)
	f.write(" ")
	f.write(a.Name)
	f.write(": ")
	f.write(a.Type)
	f.write("\n")
}

// formatClassVarAssign formats a class variable assignment.
func (f *Formatter) formatClassVarAssign(s *ast.ClassVarAssign) {
	f.writeIndent()
	f.write("@@")
	f.write(s.Name)
	f.write(" = ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// formatClassVarCompoundAssign formats a class variable compound assignment.
func (f *Formatter) formatClassVarCompoundAssign(s *ast.ClassVarCompoundAssign) {
	f.writeIndent()
	f.write("@@")
	f.write(s.Name)
	f.write(" ")
	f.write(s.Op)
	f.write("= ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// formatIncludeStmt formats an include statement.
func (f *Formatter) formatIncludeStmt(s *ast.IncludeStmt) {
	f.writeIndent()
	f.write("include ")
	f.write(s.Module)
	f.write("\n")
}

// formatInstanceVarCompoundAssign formats an instance variable compound assignment.
func (f *Formatter) formatInstanceVarCompoundAssign(s *ast.InstanceVarCompoundAssign) {
	f.writeIndent()
	f.write("@")
	f.write(s.Name)
	f.write(" ")
	f.write(s.Op)
	f.write("= ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// formatSelectorAssignStmt formats a selector assignment statement.
func (f *Formatter) formatSelectorAssignStmt(s *ast.SelectorAssignStmt) {
	f.writeIndent()
	f.formatExpr(s.Object)
	f.write(".")
	f.write(s.Field)
	f.write(" = ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// formatSelectorCompoundAssign formats a selector compound assignment statement.
func (f *Formatter) formatSelectorCompoundAssign(s *ast.SelectorCompoundAssign) {
	f.writeIndent()
	f.formatExpr(s.Object)
	f.write(".")
	f.write(s.Field)
	f.write(" ")
	f.write(s.Op)
	f.write("= ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// Phase 6: Concurrency formatters

// formatGoStmt formats a go statement.
func (f *Formatter) formatGoStmt(s *ast.GoStmt) {
	f.writeIndent()
	f.write("go ")
	if s.Block != nil {
		f.formatBlockExpr(s.Block)
	} else {
		f.formatExpr(s.Call)
	}
	f.write("\n")
}

// formatChanSendStmt formats a channel send statement.
func (f *Formatter) formatChanSendStmt(s *ast.ChanSendStmt) {
	f.writeIndent()
	f.formatExpr(s.Chan)
	f.write(" << ")
	f.formatExpr(s.Value)
	f.write("\n")
}

// Additional statement formatters

// formatPanicStmt formats a panic statement.
func (f *Formatter) formatPanicStmt(s *ast.PanicStmt) {
	f.writeIndent()
	f.write("panic ")
	f.formatExpr(s.Message)
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
