package codegen

import (
	"fmt"
	"strings"

	"github.com/nchapman/rugby/ast"
)

// isSimpleExpr returns true if the expression is safe to evaluate multiple times
// without side effects (identifiers, instance variables, self).
func isSimpleExpr(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		// All identifiers (including 'self') are simple
		return true
	case *ast.InstanceVar:
		return true
	case *ast.SelectorExpr:
		// selector is simple if receiver is simple (e.g., self.user.name)
		return isSimpleExpr(e.X)
	default:
		return false
	}
}

// getStatementLine returns the source line number for a statement, or 0 if unknown.
func getStatementLine(stmt ast.Statement) int {
	switch s := stmt.(type) {
	// Declarations
	case *ast.FuncDecl:
		return s.Line
	case *ast.ClassDecl:
		return s.Line
	case *ast.InterfaceDecl:
		return s.Line
	case *ast.ModuleDecl:
		return s.Line
	case *ast.TypeAliasDecl:
		return s.Line
	case *ast.ConstDecl:
		return s.Line
	case *ast.EnumDecl:
		return s.Line
	case *ast.StructDecl:
		return s.Line
	case *ast.MethodDecl:
		return s.Line
	case *ast.AccessorDecl:
		return s.Line
	case *ast.IncludeStmt:
		return s.Line

	// Assignments
	case *ast.AssignStmt:
		return s.Line
	case *ast.OrAssignStmt:
		return s.Line
	case *ast.CompoundAssignStmt:
		return s.Line
	case *ast.MultiAssignStmt:
		return s.Line
	case *ast.MapDestructuringStmt:
		return s.Line
	case *ast.SelectorAssignStmt:
		return s.Line
	case *ast.SelectorCompoundAssign:
		return s.Line
	case *ast.IndexAssignStmt:
		return s.Line
	case *ast.IndexCompoundAssignStmt:
		return s.Line
	case *ast.InstanceVarCompoundAssign:
		return s.Line
	case *ast.ClassVarAssign:
		return s.Line
	case *ast.ClassVarCompoundAssign:
		return s.Line

	// Expressions
	case *ast.ExprStmt:
		return s.Line

	// Control flow
	case *ast.IfStmt:
		return s.Line
	case *ast.CaseStmt:
		return s.Line
	case *ast.CaseTypeStmt:
		return s.Line
	case *ast.WhileStmt:
		return s.Line
	case *ast.UntilStmt:
		return s.Line
	case *ast.ForStmt:
		return s.Line

	// Jump statements
	case *ast.ReturnStmt:
		return s.Line
	case *ast.BreakStmt:
		return s.Line
	case *ast.NextStmt:
		return s.Line
	case *ast.PanicStmt:
		return s.Line
	case *ast.DeferStmt:
		return s.Line

	// Concurrency
	case *ast.GoStmt:
		return s.Line
	case *ast.ChanSendStmt:
		return s.Line
	case *ast.SelectStmt:
		return s.Line
	case *ast.ConcurrentlyStmt:
		return s.Line

	// Testing
	case *ast.DescribeStmt:
		return s.Line
	case *ast.ItStmt:
		return s.Line
	case *ast.TestStmt:
		return s.Line
	case *ast.TableStmt:
		return s.Line
	case *ast.BeforeStmt:
		return s.Line
	case *ast.AfterStmt:
		return s.Line

	default:
		return 0
	}
}

// genStatement dispatches to the appropriate generator for each statement type.
// Statement types are organized into logical groups:
//   - Declarations: functions, classes, interfaces, modules
//   - Assignments: simple, compound, multi-value, instance variables
//   - Control flow: conditionals, loops
//   - Jump statements: break, next, return, panic, defer
//   - Concurrency: go, select, channels, structured concurrency
//   - Testing: describe, it, test, table
func (g *Generator) genStatement(stmt ast.Statement) {
	// Emit line directive to map errors back to Rugby source
	if line := getStatementLine(stmt); line > 0 {
		g.emitLineDirective(line)
	}

	switch s := stmt.(type) {
	// Declarations
	case *ast.FuncDecl:
		g.genFuncDecl(s)
	case *ast.ClassDecl:
		g.genClassDecl(s)
	case *ast.InterfaceDecl:
		g.genInterfaceDecl(s)
	case *ast.ModuleDecl:
		g.genModuleDecl(s)
	case *ast.TypeAliasDecl:
		g.genTypeAliasDecl(s)
	case *ast.ConstDecl:
		g.genConstDecl(s)
	case *ast.EnumDecl:
		g.genEnumDecl(s)
	case *ast.StructDecl:
		g.genStructDecl(s)

	// Assignments
	case *ast.AssignStmt:
		g.genAssignStmt(s)
	case *ast.OrAssignStmt:
		g.genOrAssignStmt(s)
	case *ast.CompoundAssignStmt:
		g.genCompoundAssignStmt(s)
	case *ast.MultiAssignStmt:
		g.genMultiAssignStmt(s)
	case *ast.MapDestructuringStmt:
		g.genMapDestructuringStmt(s)
	case *ast.InstanceVarAssign:
		g.genInstanceVarAssign(s)
	case *ast.InstanceVarOrAssign:
		g.genInstanceVarOrAssign(s)
	case *ast.InstanceVarCompoundAssign:
		g.genInstanceVarCompoundAssign(s)
	case *ast.ClassVarAssign:
		g.genClassVarAssign(s)
	case *ast.ClassVarCompoundAssign:
		g.genClassVarCompoundAssign(s)
	case *ast.SelectorAssignStmt:
		g.genSelectorAssignStmt(s)
	case *ast.SelectorCompoundAssign:
		g.genSelectorCompoundAssign(s)
	case *ast.IndexAssignStmt:
		g.genIndexAssignStmt(s)
	case *ast.IndexCompoundAssignStmt:
		g.genIndexCompoundAssignStmt(s)

	// Expression statements
	case *ast.ExprStmt:
		g.genExprStmt(s)

	// Control flow: conditionals
	case *ast.IfStmt:
		g.genIfStmt(s)
	case *ast.CaseStmt:
		g.genCaseStmt(s)
	case *ast.CaseTypeStmt:
		g.genCaseTypeStmt(s)

	// Control flow: loops
	case *ast.WhileStmt:
		g.genWhileStmt(s)
	case *ast.UntilStmt:
		g.genUntilStmt(s)
	case *ast.ForStmt:
		g.genForStmt(s)
	case *ast.LoopStmt:
		g.genLoopStmt(s)

	// Jump statements
	case *ast.BreakStmt:
		g.genBreakStmt(s)
	case *ast.NextStmt:
		g.genNextStmt(s)
	case *ast.ReturnStmt:
		g.genReturnStmt(s)
	case *ast.PanicStmt:
		g.genPanicStmt(s)
	case *ast.DeferStmt:
		g.genDeferStmt(s)

	// Concurrency
	case *ast.GoStmt:
		g.genGoStmt(s)
	case *ast.SelectStmt:
		g.genSelectStmt(s)
	case *ast.ChanSendStmt:
		g.genChanSendStmt(s)
	case *ast.ConcurrentlyStmt:
		g.genConcurrentlyStmt(s)

	// Testing
	case *ast.DescribeStmt:
		g.genDescribeStmt(s)
	case *ast.ItStmt:
		g.genItStmt(s)
	case *ast.TestStmt:
		g.genTestStmt(s)
	case *ast.TableStmt:
		g.genTableStmt(s)

	default:
		g.addError(fmt.Errorf("unhandled statement type: %T", stmt))
	}
}

func (g *Generator) genExprStmt(s *ast.ExprStmt) {
	// Handle BangExpr: call()! as a statement
	if bangExpr, ok := s.Expr.(*ast.BangExpr); ok {
		g.genBangStmt(bangExpr)
		return
	}

	// Handle RescueExpr: call() rescue do ... end as a statement
	if rescueExpr, ok := s.Expr.(*ast.RescueExpr); ok {
		g.genRescueStmt(rescueExpr)
		return
	}

	// Handle BinaryExpr with << on Ident: arr << value becomes arr = arr << value
	// This provides Ruby-like mutation semantics for array append
	// For channels, just generate: ch <- value (no assignment wrapper needed)
	if binExpr, ok := s.Expr.(*ast.BinaryExpr); ok {
		if binExpr.Op == "<<" {
			// Check if left operand is a channel - generate native send without assignment
			if g.typeInfo.GetTypeKind(binExpr.Left) == TypeChannel {
				g.writeIndent()
				g.genExpr(s.Expr) // genExpr generates "ch <- value" for channels
				g.buf.WriteString("\n")
				return
			}
			if ident, ok := binExpr.Left.(*ast.Ident); ok {
				// Generate: arr = runtime.ShiftLeft(arr, value).([]T)
				g.writeIndent()
				g.buf.WriteString(ident.Name)
				g.buf.WriteString(" = ")
				g.genExpr(s.Expr)
				// Add type assertion if needed
				if goType := g.getShiftLeftTargetType(binExpr); goType != "" {
					g.buf.WriteString(".(")
					g.buf.WriteString(goType)
					g.buf.WriteString(")")
				}
				g.buf.WriteString("\n")
				return
			}
		}
	}

	// Optimize Map.delete(key) to Go's delete() builtin when return value is discarded
	// In Rugby, .delete(key) with one arg is only defined for Map/Hash types in stdLib
	if call, ok := s.Expr.(*ast.CallExpr); ok {
		if sel, ok := call.Func.(*ast.SelectorExpr); ok {
			if sel.Sel == "delete" && len(call.Args) == 1 {
				g.writeIndent()
				g.buf.WriteString("delete(")
				g.genExpr(sel.X)
				g.buf.WriteString(", ")
				g.genExpr(call.Args[0])
				g.buf.WriteString(")\n")
				return
			}
		}
	}

	// Convert SelectorExpr to CallExpr when used as statement (Ruby-style method call)
	// e.g., "obj.foo" as statement becomes "obj.foo()"
	expr := s.Expr
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		expr = &ast.CallExpr{Func: sel, Args: nil}
	}

	// Convert Ident to CallExpr when it's a no-arg function (Ruby-style method call)
	// e.g., "helper" as statement becomes "helper()"
	if ident, ok := expr.(*ast.Ident); ok {
		if g.isNoArgFunction(ident.Name) {
			expr = &ast.CallExpr{Func: ident, Args: nil}
		}
	}

	if s.Condition != nil {
		// Wrap in if/unless block
		g.writeIndent()
		g.buf.WriteString("if ")
		if s.IsUnless {
			g.buf.WriteString("!(")
		}
		if err := g.genCondition(s.Condition); err != nil {
			g.addError(err)
			g.buf.WriteString("false")
		}
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
		g.indent++
		g.writeIndent()
		g.genExpr(expr)
		g.buf.WriteString("\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	} else {
		g.writeIndent()
		g.genExpr(expr)
		g.buf.WriteString("\n")
	}
}

func (g *Generator) genBreakStmt(s *ast.BreakStmt) {
	if s.Condition != nil {
		// Wrap in if/unless block
		g.writeIndent()
		g.buf.WriteString("if ")
		if s.IsUnless {
			g.buf.WriteString("!(")
		}
		if err := g.genCondition(s.Condition); err != nil {
			g.addError(err)
			g.buf.WriteString("false")
		}
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
		g.indent++
	}

	g.writeIndent()
	g.buf.WriteString("break\n")

	if s.Condition != nil {
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	}
}

func (g *Generator) genNextStmt(s *ast.NextStmt) {
	if s.Condition != nil {
		// Wrap in if/unless block
		g.writeIndent()
		g.buf.WriteString("if ")
		if s.IsUnless {
			g.buf.WriteString("!(")
		}
		if err := g.genCondition(s.Condition); err != nil {
			g.addError(err)
			g.buf.WriteString("false")
		}
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
		g.indent++
	}

	g.writeIndent()
	g.buf.WriteString("continue\n")

	if s.Condition != nil {
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	}
}

func (g *Generator) genReturnStmt(s *ast.ReturnStmt) {
	if s.Condition != nil {
		// Wrap in if/unless block
		g.writeIndent()
		g.buf.WriteString("if ")
		if s.IsUnless {
			g.buf.WriteString("!(")
		}
		if err := g.genCondition(s.Condition); err != nil {
			g.addError(err)
			g.buf.WriteString("false")
		}
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
		g.indent++
	}

	g.writeIndent()
	// In inlined lambdas (each/times loops), return acts like continue
	// This matches Ruby semantics where return in a block skips to the next iteration
	// The return value (if any) is discarded since each/times ignore block return values
	if g.inInlinedLambda {
		g.buf.WriteString("continue")
		// Close the condition block if present
		if s.Condition != nil {
			g.buf.WriteString("\n")
			g.indent--
			g.writeIndent()
			g.buf.WriteString("}\n")
		} else {
			g.buf.WriteString("\n")
		}
		return
	}
	g.buf.WriteString("return")
	if len(s.Values) > 0 {
		g.buf.WriteString(" ")
		for i, val := range s.Values {
			if i > 0 {
				g.buf.WriteString(", ")
			}

			// Check if we need to wrap for Optional types
			needsWrap := false
			baseType := ""
			handled := false

			if i < len(g.currentReturnTypes()) {
				targetType := g.currentReturnTypes()[i]
				if isValueTypeOptional(targetType) {
					baseType = strings.TrimSuffix(targetType, "?")

					// If returning nil, use NoneT()
					if _, isNil := val.(*ast.NilLit); isNil {
						g.buf.WriteString(fmt.Sprintf("runtime.None%s()", baseType))
						handled = true
					} else {
						// Infer type of val to see if it needs wrapping
						inferred := g.inferTypeFromExpr(val)
						if inferred == baseType {
							needsWrap = true
						}
					}
				}
			}

			if !handled {
				if needsWrap {
					g.buf.WriteString(fmt.Sprintf("runtime.Some%s(", baseType))
					g.genExpr(val)
					g.buf.WriteString(")")
				} else {
					g.genExpr(val)
				}
			}
		}
	}
	g.buf.WriteString("\n")

	if s.Condition != nil {
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	}
}

func (g *Generator) genPanicStmt(s *ast.PanicStmt) {
	if s.Condition != nil {
		// panic "msg" if cond  ->  if cond { panic("msg") }
		// panic "msg" unless cond  ->  if !cond { panic("msg") }
		g.writeIndent()
		g.buf.WriteString("if ")
		if s.IsUnless {
			g.buf.WriteString("!(")
			g.genExpr(s.Condition)
			g.buf.WriteString(")")
		} else {
			g.genExpr(s.Condition)
		}
		g.buf.WriteString(" {\n")
		g.indent++
		g.writeIndent()
		g.buf.WriteString("panic(")
		g.genExpr(s.Message)
		g.buf.WriteString(")\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	} else {
		g.writeIndent()
		g.buf.WriteString("panic(")
		g.genExpr(s.Message)
		g.buf.WriteString(")\n")
	}
}

func (g *Generator) genDeferStmt(s *ast.DeferStmt) {
	g.writeIndent()
	g.buf.WriteString("defer ")
	g.genCallExpr(s.Call)
	g.buf.WriteString("\n")
}

func (g *Generator) genGoStmt(s *ast.GoStmt) {
	g.writeIndent()
	// Check if it's a block (go do ... end) or a simple call (go func())
	if s.Block != nil {
		// go func() { ... }()
		g.buf.WriteString("go func() {\n")
		g.indent++
		for _, stmt := range s.Block.Body {
			g.genStatement(stmt)
		}
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}()\n")
	} else if s.Call != nil {
		// go func()
		g.buf.WriteString("go ")
		if call, ok := s.Call.(*ast.CallExpr); ok {
			g.genCallExpr(call)
		} else {
			g.addError(fmt.Errorf("line %d: go statement must be a function call", s.Line))
			g.buf.WriteString("func(){}()")
		}
		g.buf.WriteString("\n")
	}
}

func (g *Generator) genSelectStmt(s *ast.SelectStmt) {
	g.writeIndent()
	g.buf.WriteString("select {\n")

	for _, when := range s.Cases {
		g.writeIndent()
		g.buf.WriteString("case ")

		if !when.IsSend {
			// case val := <-ch:
			if when.AssignName != "" {
				g.buf.WriteString(when.AssignName)
				g.buf.WriteString(" := ")
				// Infer type of received value (not implemented fully without type info)
				// For now assume "any" or rely on runtime
				g.vars[when.AssignName] = "any"
			}

			// Unwrap .receive or .try_receive to get the underlying channel
			expr := when.Chan
			if sel, ok := expr.(*ast.SelectorExpr); ok {
				if sel.Sel == "receive" || sel.Sel == "try_receive" {
					expr = sel.X
				}
			}

			g.buf.WriteString("<-")
			g.genExpr(expr)
		} else {
			// case ch <- val:
			g.genExpr(when.Chan)
			g.buf.WriteString(" <- ")
			g.genExpr(when.Value)
		}

		g.buf.WriteString(":\n")
		g.indent++
		for _, stmt := range when.Body {
			g.genStatement(stmt)
		}
		g.indent--
	}

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

func (g *Generator) genChanSendStmt(s *ast.ChanSendStmt) {
	g.writeIndent()
	g.genExpr(s.Chan)
	g.buf.WriteString(" <- ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

func (g *Generator) genConcurrentlyStmt(s *ast.ConcurrentlyStmt) {
	// Generate:
	// func() (T, T, error) {
	//   ctx, cancel := context.WithCancel(context.Background())
	//   defer cancel()
	//   var wg sync.WaitGroup
	//   scope := &runtime.Scope{Ctx: ctx, Wg: &wg}
	//   ...
	//   wg.Wait()
	//   return ...
	// }()

	g.needsRuntime = true
	g.buf.WriteString("func() ")

	// Return types are currently just "any" because we can't easily infer block return type
	// This might need refinement for typed returns
	g.buf.WriteString("any")

	g.buf.WriteString(" {\n")
	g.indent++

	// Create scope variable (use fallback if not specified)
	scopeVar := s.ScopeVar
	if scopeVar == "" {
		scopeVar = "_scope"
	}
	g.vars[scopeVar] = "Scope"

	// Create scope object
	g.writeIndent()
	g.buf.WriteString(fmt.Sprintf("%s := runtime.NewScope()\n", scopeVar))

	// Generate: defer scope.Wait()
	g.writeIndent()
	g.buf.WriteString(fmt.Sprintf("defer %s.Wait()\n", scopeVar))

	// Generate body
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}

	g.indent--
	g.writeIndent()
	g.buf.WriteString("}()\n") // Call the closure immediately
}
