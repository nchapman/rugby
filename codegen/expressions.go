package codegen

import (
	"fmt"
	"strings"

	"github.com/nchapman/rugby/ast"
)

// genExpr dispatches to the appropriate generator for each expression type.
// Expression types are organized into logical groups:
//   - Literals: strings, numbers, bools, nil, symbols, arrays, maps, ranges
//   - Identifiers: variables, instance variables, self
//   - Operators: binary, unary, ternary, nil-coalescing
//   - Access: calls, selectors, indexing, safe navigation
//   - Concurrency: spawn, await
//   - Special: super, symbol-to-proc
func (g *Generator) genExpr(expr ast.Expression) {
	switch e := expr.(type) {
	// Literals: simple values
	case *ast.StringLit:
		g.buf.WriteString(fmt.Sprintf("%q", e.Value))
	case *ast.InterpolatedString:
		g.genInterpolatedString(e)
	case *ast.IntLit:
		g.buf.WriteString(fmt.Sprintf("%d", e.Value))
	case *ast.FloatLit:
		g.buf.WriteString(fmt.Sprintf("%g", e.Value))
	case *ast.BoolLit:
		if e.Value {
			g.buf.WriteString("true")
		} else {
			g.buf.WriteString("false")
		}
	case *ast.NilLit:
		g.buf.WriteString("nil")
	case *ast.SymbolLit:
		g.buf.WriteString(fmt.Sprintf("%q", e.Value))

	// Literals: composite values
	case *ast.ArrayLit:
		g.genArrayLit(e)
	case *ast.MapLit:
		g.genMapLit(e)
	case *ast.RangeLit:
		g.genRangeLit(e)

	// Identifiers
	case *ast.Ident:
		// Handle 'self' keyword - compiles to receiver variable
		if e.Name == "self" {
			if g.currentClass != "" {
				g.buf.WriteString(receiverName(g.currentClass))
			} else {
				g.buf.WriteString("/* self outside class */")
			}
		} else if runtimeCall, ok := noParenKernelFuncs[e.Name]; ok {
			// Check for kernel functions that can be used without parens
			g.needsRuntime = true
			g.buf.WriteString(runtimeCall)
		} else {
			g.buf.WriteString(e.Name)
		}
	case *ast.InstanceVar:
		if g.currentClass != "" {
			recv := receiverName(g.currentClass)
			g.buf.WriteString(fmt.Sprintf("%s.%s", recv, e.Name))
		} else {
			g.buf.WriteString(fmt.Sprintf("/* @%s outside class */", e.Name))
		}

	// Operators
	case *ast.BinaryExpr:
		g.genBinaryExpr(e)
	case *ast.UnaryExpr:
		g.genUnaryExpr(e)
	case *ast.TernaryExpr:
		g.genTernaryExpr(e)
	case *ast.NilCoalesceExpr:
		g.genNilCoalesceExpr(e)

	// Access: calls, member access, indexing
	case *ast.CallExpr:
		g.genCallExpr(e)
	case *ast.SelectorExpr:
		g.genSelectorExpr(e)
	case *ast.IndexExpr:
		g.genIndexExpr(e)
	case *ast.SafeNavExpr:
		g.genSafeNavExpr(e)

	// Concurrency
	case *ast.SpawnExpr:
		g.genSpawnExpr(e)
	case *ast.AwaitExpr:
		g.genAwaitExpr(e)

	// Special
	case *ast.SuperExpr:
		g.genSuperExpr(e)
	case *ast.SymbolToProcExpr:
		g.genSymbolToProcExpr(e)
	}
}

func (g *Generator) genSuperExpr(e *ast.SuperExpr) {
	if g.currentClass == "" || g.currentMethod == "" {
		g.buf.WriteString("/* super outside method */")
		return
	}
	if len(g.currentClassEmbeds) == 0 {
		g.addError(fmt.Errorf("line %d: super used in class without parent", e.Line))
		g.buf.WriteString("/* super without parent */")
		return
	}

	// Use the first embedded type as the parent class
	parentClass := g.currentClassEmbeds[0]
	recv := receiverName(g.currentClass)

	// Convert the current method name to Go's naming convention
	var methodName string
	if g.currentMethodPub {
		methodName = snakeToPascalWithAcronyms(g.currentMethod)
	} else {
		methodName = snakeToCamelWithAcronyms(g.currentMethod)
	}

	// Generate: recv.ParentName.MethodName(args...)
	g.buf.WriteString(fmt.Sprintf("%s.%s.%s(", recv, parentClass, methodName))
	for i, arg := range e.Args {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.genExpr(arg)
	}
	g.buf.WriteString(")")
}

func (g *Generator) genRangeLit(r *ast.RangeLit) {
	startType := g.inferTypeFromExpr(r.Start)
	endType := g.inferTypeFromExpr(r.End)

	if startType != "" && startType != "Int" {
		g.addError(fmt.Errorf("line %d: range start must be Int, got %s", r.Line, startType))
	}
	if endType != "" && endType != "Int" {
		g.addError(fmt.Errorf("line %d: range end must be Int, got %s", r.Line, endType))
	}

	g.needsRuntime = true
	g.buf.WriteString("runtime.Range{Start: ")
	g.genExpr(r.Start)
	g.buf.WriteString(", End: ")
	g.genExpr(r.End)
	g.buf.WriteString(", Exclusive: ")
	if r.Exclusive {
		g.buf.WriteString("true")
	} else {
		g.buf.WriteString("false")
	}
	g.buf.WriteString("}")
}

func (g *Generator) genRangeMethodCall(rangeExpr ast.Expression, method string, args []ast.Expression) bool {
	g.needsRuntime = true

	switch method {
	case "to_a":
		g.buf.WriteString("runtime.RangeToArray(")
		g.genExpr(rangeExpr)
		g.buf.WriteString(")")
		return true
	case "size", "length":
		g.buf.WriteString("runtime.RangeSize(")
		g.genExpr(rangeExpr)
		g.buf.WriteString(")")
		return true
	case "include?", "contains?":
		if len(args) != 1 {
			return false
		}
		g.buf.WriteString("runtime.RangeContains(")
		g.genExpr(rangeExpr)
		g.buf.WriteString(", ")
		g.genExpr(args[0])
		g.buf.WriteString(")")
		return true
	default:
		return false
	}
}

func (g *Generator) genBinaryExpr(e *ast.BinaryExpr) {
	leftLit := isPrimitiveLiteral(e.Left)
	rightLit := isPrimitiveLiteral(e.Right)
	needsRuntimeEqual := !leftLit || !rightLit

	if e.Op == "==" && needsRuntimeEqual {
		g.needsRuntime = true
		g.buf.WriteString("runtime.Equal(")
		g.genExpr(e.Left)
		g.buf.WriteString(", ")
		g.genExpr(e.Right)
		g.buf.WriteString(")")
		return
	}
	if e.Op == "!=" && needsRuntimeEqual {
		g.needsRuntime = true
		g.buf.WriteString("!runtime.Equal(")
		g.genExpr(e.Left)
		g.buf.WriteString(", ")
		g.genExpr(e.Right)
		g.buf.WriteString(")")
		return
	}

	if e.Op == "<<" {
		g.needsRuntime = true
		g.buf.WriteString("runtime.ShiftLeft(")
		g.genExpr(e.Left)
		g.buf.WriteString(", ")
		g.genExpr(e.Right)
		g.buf.WriteString(")")
		return
	}

	g.buf.WriteString("(")
	g.genExpr(e.Left)
	g.buf.WriteString(" ")
	g.buf.WriteString(e.Op)
	g.buf.WriteString(" ")
	g.genExpr(e.Right)
	g.buf.WriteString(")")
}

func isPrimitiveLiteral(e ast.Expression) bool {
	switch e.(type) {
	case *ast.IntLit, *ast.FloatLit, *ast.BoolLit, *ast.StringLit:
		return true
	default:
		return false
	}
}

func (g *Generator) genUnaryExpr(e *ast.UnaryExpr) {
	g.buf.WriteString(e.Op)
	g.genExpr(e.Expr)
}

func (g *Generator) genNilCoalesceExpr(e *ast.NilCoalesceExpr) {
	g.needsRuntime = true
	leftType := g.inferTypeFromExpr(e.Left)

	switch leftType {
	case "Int?":
		g.buf.WriteString("runtime.CoalesceInt(")
	case "Int64?":
		g.buf.WriteString("runtime.CoalesceInt64(")
	case "Float?":
		g.buf.WriteString("runtime.CoalesceFloat(")
	case "String?":
		g.buf.WriteString("runtime.CoalesceString(")
	case "Bool?":
		g.buf.WriteString("runtime.CoalesceBool(")
	default:
		varName := fmt.Sprintf("_nc%d", g.tempVarCounter)
		g.tempVarCounter++
		g.buf.WriteString("func() ")
		rightType := g.inferTypeFromExpr(e.Right)
		if rightType != "" {
			g.buf.WriteString(mapType(rightType))
		} else {
			g.buf.WriteString("any")
		}
		g.buf.WriteString(" { ")
		g.buf.WriteString(varName)
		g.buf.WriteString(" := ")
		g.genExpr(e.Left)
		g.buf.WriteString("; if ")
		g.buf.WriteString(varName)
		g.buf.WriteString(" != nil { return *")
		g.buf.WriteString(varName)
		g.buf.WriteString(" }; return ")
		g.genExpr(e.Right)
		g.buf.WriteString(" }()")
		return
	}

	g.genExpr(e.Left)
	g.buf.WriteString(", ")
	g.genExpr(e.Right)
	g.buf.WriteString(")")
}

func (g *Generator) genTernaryExpr(e *ast.TernaryExpr) {
	// Go doesn't have ternary operator, use IIFE
	// func() T { if cond { return trueVal } else { return falseVal } }()

	// Determine return type
	trueType := g.inferTypeFromExpr(e.Then)
	falseType := g.inferTypeFromExpr(e.Else)

	var retType string
	if trueType == falseType && trueType != "" {
		retType = mapType(trueType)
	} else {
		retType = "any"
	}

	g.buf.WriteString("func() ")
	g.buf.WriteString(retType)
	g.buf.WriteString(" { if ")
	g.genExpr(e.Condition)
	g.buf.WriteString(" { return ")
	g.genExpr(e.Then)
	g.buf.WriteString(" } else { return ")
	g.genExpr(e.Else)
	g.buf.WriteString(" } }()")
}

func (g *Generator) genSymbolToProcExpr(e *ast.SymbolToProcExpr) {
	// Generate a generic function wrapper
	// func(x any) any { return runtime.CallMethod(x, "method") }
	g.needsRuntime = true
	g.buf.WriteString("func(x any) any { return runtime.CallMethod(x, \"")
	g.buf.WriteString(e.Method)
	g.buf.WriteString("\") }")
}

func (g *Generator) genSafeNavExpr(e *ast.SafeNavExpr) {
	// obj&.method -> func() any { _snN := obj; if _snN != nil { return (*_snN).method } else { return nil } }()
	varName := fmt.Sprintf("_sn%d", g.tempVarCounter)
	g.tempVarCounter++

	g.buf.WriteString("func() any { ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := ")
	g.genExpr(e.Receiver)
	g.buf.WriteString("; if ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" != nil { return ")

	// Use (*var).field
	g.buf.WriteString("(*")
	g.buf.WriteString(varName)
	g.buf.WriteString(").")
	g.buf.WriteString(e.Selector)

	g.buf.WriteString(" } else { return nil } }()")
}

func (g *Generator) genInterpolatedString(s *ast.InterpolatedString) {
	g.needsFmt = true
	g.buf.WriteString("fmt.Sprintf(\"")

	// Format string with %v placeholders
	for _, part := range s.Parts {
		if str, ok := part.(*ast.StringLit); ok {
			// Escape " and \ inside the format string
			escaped := strings.ReplaceAll(str.Value, "\"", "\\\"")
			escaped = strings.ReplaceAll(escaped, "\n", "\\n")
			escaped = strings.ReplaceAll(escaped, "%", "%%")
			g.buf.WriteString(escaped)
		} else if strVal, ok := part.(string); ok {
			// Handle raw string parts (from parser)
			escaped := strings.ReplaceAll(strVal, "\"", "\\\"")
			escaped = strings.ReplaceAll(escaped, "\n", "\\n")
			escaped = strings.ReplaceAll(escaped, "%", "%%")
			g.buf.WriteString(escaped)
		} else {
			g.buf.WriteString("%v")
		}
	}
	g.buf.WriteString("\"")

	// Arguments
	for _, part := range s.Parts {
		if _, ok := part.(*ast.StringLit); ok {
			continue
		}
		if _, ok := part.(string); ok {
			continue
		}
		g.buf.WriteString(", ")
		if expr, ok := part.(ast.Expression); ok {
			g.genExpr(expr)
		} else {
			g.buf.WriteString("nil /* error: invalid part */")
		}
	}
	g.buf.WriteString(")")
}

func (g *Generator) genArrayLit(arr *ast.ArrayLit) {
	// Check for splat expressions
	hasSplat := false
	for _, elem := range arr.Elements {
		if _, ok := elem.(*ast.SplatExpr); ok {
			hasSplat = true
			break
		}
	}

	if hasSplat {
		g.genArrayLitWithSplat(arr)
		return
	}

	// Try to infer element type
	elemType := "any"
	if len(arr.Elements) > 0 {
		firstType := g.inferTypeFromExpr(arr.Elements[0])
		consistent := true
		for _, e := range arr.Elements[1:] {
			if g.inferTypeFromExpr(e) != firstType {
				consistent = false
				break
			}
		}
		if consistent && firstType != "" {
			elemType = mapType(firstType)
		}
	}

	g.buf.WriteString("[]" + elemType + "{")
	for i, elem := range arr.Elements {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.genExpr(elem)
	}
	g.buf.WriteString("}")
}

func (g *Generator) genArrayLitWithSplat(arr *ast.ArrayLit) {
	// Use unique variable names
	arrVar := fmt.Sprintf("_arr%d", g.tempVarCounter)
	elemVar := fmt.Sprintf("_v%d", g.tempVarCounter)
	g.tempVarCounter++

	// Generate: func() []any { _arrN := []any{}; ... ; return _arrN }()
	g.buf.WriteString("func() []any {\n")
	g.indent++

	g.writeIndent()
	g.buf.WriteString(fmt.Sprintf("%s := []any{}\n", arrVar))

	for _, elem := range arr.Elements {
		if splat, ok := elem.(*ast.SplatExpr); ok {
			// Splat: append the splatted array
			g.writeIndent()
			g.buf.WriteString(fmt.Sprintf("for _, %s := range ", elemVar))
			g.genExpr(splat.Expr)
			g.buf.WriteString(" {\n")
			g.indent++
			g.writeIndent()
			g.buf.WriteString(fmt.Sprintf("%s = append(%s, %s)\n", arrVar, arrVar, elemVar))
			g.indent--
			g.writeIndent()
			g.buf.WriteString("}\n")
		} else {
			// Regular element: append single value
			g.writeIndent()
			g.buf.WriteString(fmt.Sprintf("%s = append(%s, ", arrVar, arrVar))
			g.genExpr(elem)
			g.buf.WriteString(")\n")
		}
	}

	g.writeIndent()
	g.buf.WriteString(fmt.Sprintf("return %s\n", arrVar))
	g.indent--
	g.writeIndent()
	g.buf.WriteString("}()")
}

func (g *Generator) genIndexExpr(idx *ast.IndexExpr) {
	if r, ok := idx.Index.(*ast.RangeLit); ok {
		g.genRangeSlice(idx.Left, r)
		return
	}

	if _, isStringKey := idx.Index.(*ast.StringLit); isStringKey {
		switch idx.Left.(type) {
		case *ast.Ident, *ast.MapLit:
			g.genExpr(idx.Left)
			g.buf.WriteString("[")
			g.genExpr(idx.Index)
			g.buf.WriteString("]")
			return
		default:
			g.needsRuntime = true
			g.buf.WriteString("runtime.GetKey(")
			g.genExpr(idx.Left)
			g.buf.WriteString(", ")
			g.genExpr(idx.Index)
			g.buf.WriteString(")")
			return
		}
	}

	if g.shouldUseNativeIndex(idx.Index) {
		g.genExpr(idx.Left)
		g.buf.WriteString("[")
		g.genExpr(idx.Index)
		g.buf.WriteString("]")
		return
	}

	g.needsRuntime = true
	g.buf.WriteString("runtime.AtIndex(")
	g.genExpr(idx.Left)
	g.buf.WriteString(", ")
	g.genExpr(idx.Index)
	g.buf.WriteString(")")
}

func (g *Generator) genRangeSlice(collection ast.Expression, r *ast.RangeLit) {
	g.needsRuntime = true
	g.buf.WriteString("runtime.Slice(")
	g.genExpr(collection)
	g.buf.WriteString(", runtime.Range{Start: ")
	g.genExpr(r.Start)
	g.buf.WriteString(", End: ")
	g.genExpr(r.End)
	g.buf.WriteString(", Exclusive: ")
	if r.Exclusive {
		g.buf.WriteString("true")
	} else {
		g.buf.WriteString("false")
	}
	g.buf.WriteString("})")
}

func (g *Generator) shouldUseNativeIndex(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.IntLit:
		return e.Value >= 0
	case *ast.StringLit:
		return true
	default:
		return false
	}
}

func (g *Generator) genMapLit(m *ast.MapLit) {
	hasSplat := false
	for _, entry := range m.Entries {
		if entry.Splat != nil {
			hasSplat = true
			break
		}
	}

	if hasSplat {
		g.genMapLitWithSplat(m)
		return
	}

	g.buf.WriteString("map[any]any{")
	for i, entry := range m.Entries {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.genExpr(entry.Key)
		g.buf.WriteString(": ")
		g.genExpr(entry.Value)
	}
	g.buf.WriteString("}")
}

func (g *Generator) genMapLitWithSplat(m *ast.MapLit) {
	mapVar := fmt.Sprintf("_map%d", g.tempVarCounter)
	keyVar := fmt.Sprintf("_k%d", g.tempVarCounter)
	valVar := fmt.Sprintf("_v%d", g.tempVarCounter)
	g.tempVarCounter++

	g.buf.WriteString("func() map[any]any {\n")
	g.indent++

	g.writeIndent()
	g.buf.WriteString(fmt.Sprintf("%s := map[any]any{}\n", mapVar))

	for _, entry := range m.Entries {
		if entry.Splat != nil {
			g.writeIndent()
			g.buf.WriteString(fmt.Sprintf("for %s, %s := range ", keyVar, valVar))
			g.genExpr(entry.Splat)
			g.buf.WriteString(" {\n")
			g.indent++
			g.writeIndent()
			g.buf.WriteString(fmt.Sprintf("%s[%s] = %s\n", mapVar, keyVar, valVar))
			g.indent--
			g.writeIndent()
			g.buf.WriteString("}\n")
		} else {
			g.writeIndent()
			g.buf.WriteString(fmt.Sprintf("%s[", mapVar))
			g.genExpr(entry.Key)
			g.buf.WriteString("] = ")
			g.genExpr(entry.Value)
			g.buf.WriteString("\n")
		}
	}

	g.writeIndent()
	g.buf.WriteString(fmt.Sprintf("return %s\n", mapVar))
	g.indent--
	g.writeIndent()
	g.buf.WriteString("}()")
}

func (g *Generator) genCallExpr(call *ast.CallExpr) {
	if call.Block != nil {
		g.genBlockCall(call)
		return
	}

	if sel, ok := call.Func.(*ast.SelectorExpr); ok {
		if bm, isBlockMethod := blockMethods[sel.Sel]; isBlockMethod {
			if len(call.Args) >= 1 {
				if stp, ok := call.Args[0].(*ast.SymbolToProcExpr); ok {
					g.genSymbolToProcBlockCall(sel.X, stp, bm)
					return
				}
			}
		}
	}

	if sel, ok := call.Func.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			if ident.Name == "assert" || ident.Name == "require" {
				g.genTestAssertion(call, ident.Name, sel.Sel)
				return
			}
		}
	}

	switch fn := call.Func.(type) {
	case *ast.Ident:
		funcName := fn.Name

		if funcName == "error_is?" {
			if len(call.Args) != 2 {
				g.addError(fmt.Errorf("error_is? requires 2 arguments (err, target), got %d", len(call.Args)))
				g.buf.WriteString("false")
				return
			}
			g.needsErrors = true
			g.buf.WriteString("errors.Is(")
			g.genExpr(call.Args[0])
			g.buf.WriteString(", ")
			g.genExpr(call.Args[1])
			g.buf.WriteString(")")
			return
		}

		if funcName == "error_as" {
			if len(call.Args) != 2 {
				g.addError(fmt.Errorf("error_as requires 2 arguments (err, Type), got %d", len(call.Args)))
				g.buf.WriteString("nil")
				return
			}
			g.needsErrors = true
			g.buf.WriteString("func() *")
			g.genExpr(call.Args[1])
			g.buf.WriteString(" { var _target *")
			g.genExpr(call.Args[1])
			g.buf.WriteString("; if errors.As(")
			g.genExpr(call.Args[0])
			g.buf.WriteString(", &_target) { return _target }; return nil }()")
			return
		}

		if kf, ok := kernelFuncs[funcName]; ok {
			g.needsRuntime = true
			if kf.transform != nil {
				funcName = kf.transform(call.Args)
			} else {
				funcName = kf.runtimeFunc
			}
		} else {
			funcName = snakeToCamelWithAcronyms(funcName)
		}
		g.buf.WriteString(funcName)
	case *ast.SelectorExpr:
		if fn.Sel == "new" {
			if indexExpr, ok := fn.X.(*ast.IndexExpr); ok {
				if chanIdent, ok := indexExpr.Left.(*ast.Ident); ok && chanIdent.Name == "Chan" {
					var chanType string
					if typeIdent, ok := indexExpr.Index.(*ast.Ident); ok {
						chanType = typeIdent.Name
					} else {
						chanType = "any"
					}
					goType := mapType(chanType)

					g.buf.WriteString("make(chan ")
					g.buf.WriteString(goType)
					if len(call.Args) > 0 {
						g.buf.WriteString(", ")
						g.genExpr(call.Args[0])
					}
					g.buf.WriteString(")")
					return
				}
			}
		}

		if fn.Sel == "new" {
			if ident, ok := fn.X.(*ast.Ident); ok {
				var ctorName string
				if g.pubClasses[ident.Name] {
					ctorName = fmt.Sprintf("New%s", ident.Name)
				} else {
					ctorName = fmt.Sprintf("new%s", ident.Name)
				}
				g.buf.WriteString(ctorName)
				g.buf.WriteString("(")
				for i, arg := range call.Args {
					if i > 0 {
						g.buf.WriteString(", ")
					}
					g.genExpr(arg)
				}
				g.buf.WriteString(")")
				return
			}
		}

		if fn.Sel == "close" {
			g.buf.WriteString("close(")
			g.genExpr(fn.X)
			g.buf.WriteString(")")
			return
		}
		if fn.Sel == "receive" {
			g.buf.WriteString("<- ")
			g.genExpr(fn.X)
			return
		}
		if fn.Sel == "try_receive" {
			g.needsRuntime = true
			g.buf.WriteString("runtime.TryReceive(")
			g.genExpr(fn.X)
			g.buf.WriteString(")")
			return
		}

		recvType := g.inferTypeFromExpr(fn.X)
		var methodDef MethodDef
		found := false

		if ident, ok := fn.X.(*ast.Ident); ok && ident.Name == "Math" {
			if methods, ok := stdLib["Math"]; ok {
				if def, ok := methods[fn.Sel]; ok {
					g.needsRuntime = true
					g.buf.WriteString(def.RuntimeFunc)
					g.buf.WriteString("(")
					for i, arg := range call.Args {
						if i > 0 {
							g.buf.WriteString(", ")
						}
						g.genExpr(arg)
					}
					g.buf.WriteString(")")
					return
				}
			}
		}

		if recvType != "" {
			if methods, ok := stdLib[recvType]; ok {
				if def, ok := methods[fn.Sel]; ok {
					methodDef = def
					found = true
				}
			}
		}

		if !found {
			if def, ok := uniqueMethods[fn.Sel]; ok {
				methodDef = def
				found = true
			}
		}

		if found {
			g.needsRuntime = true
			g.buf.WriteString(methodDef.RuntimeFunc)
			g.buf.WriteString("(")
			g.genExpr(fn.X)
			if len(call.Args) > 0 {
				g.buf.WriteString(", ")
				for i, arg := range call.Args {
					if i > 0 {
						g.buf.WriteString(", ")
					}
					g.genExpr(arg)
				}
			}
			g.buf.WriteString(")")
			return
		}

		if fn.Sel == "is_a?" {
			if len(call.Args) != 1 {
				g.buf.WriteString("false /* ERROR: is_a? requires exactly one type argument */")
				return
			}
			g.buf.WriteString("func() bool { _, ok := ")
			g.genExpr(fn.X)
			g.buf.WriteString(".(")
			g.genExpr(call.Args[0])
			g.buf.WriteString("); return ok }()")
			return
		}

		if fn.Sel == "as" {
			if len(call.Args) != 1 {
				g.buf.WriteString("func() (any, bool) { return nil, false }() /* ERROR: as requires exactly one type argument */")
				return
			}
			g.buf.WriteString("func() (")
			g.genExpr(call.Args[0])
			g.buf.WriteString(", bool) { v, ok := ")
			g.genExpr(fn.X)
			g.buf.WriteString(".(")
			g.genExpr(call.Args[0])
			g.buf.WriteString("); return v, ok }()")
			return
		}

		receiverType := g.inferTypeFromExpr(fn.X)
		if isOptionalType(receiverType) {
			switch fn.Sel {
			case "ok?", "present?":
				g.buf.WriteString("(")
				g.genExpr(fn.X)
				g.buf.WriteString(" != nil)")
				return
			case "nil?", "absent?":
				g.buf.WriteString("(")
				g.genExpr(fn.X)
				g.buf.WriteString(" == nil)")
				return
			case "unwrap":
				g.buf.WriteString("*")
				g.genExpr(fn.X)
				return
			}
		}

		if _, ok := fn.X.(*ast.RangeLit); ok {
			if g.genRangeMethodCall(fn.X, fn.Sel, call.Args) {
				return
			}
		}
		g.genSelectorExpr(fn)
	default:
		g.genExpr(call.Func)
	}

	g.buf.WriteString("(")
	for i, arg := range call.Args {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.genExpr(arg)
	}
	g.buf.WriteString(")")
}

func (g *Generator) genBlockCall(call *ast.CallExpr) {
	sel, ok := call.Func.(*ast.SelectorExpr)
	if !ok {
		g.buf.WriteString("/* unsupported block call */")
		return
	}

	method := sel.Sel
	block := call.Block

	switch method {
	case "each":
		g.genEachBlock(sel.X, block)
		return
	case "each_with_index":
		g.genEachWithIndexBlock(sel.X, block)
		return
	case "times":
		g.genTimesBlock(sel.X, block)
		return
	case "upto":
		g.genUptoBlock(sel.X, block, call.Args)
		return
	case "downto":
		g.genDowntoBlock(sel.X, block, call.Args)
		return
	case "spawn":
		g.genScopedSpawnBlock(sel.X, block)
		return
	}

	if bm, ok := blockMethods[method]; ok {
		g.genRuntimeBlock(sel.X, block, call.Args, bm)
		return
	}

	g.buf.WriteString(fmt.Sprintf("/* unsupported block method: %s */", method))
}

func (g *Generator) genEachBlock(iterable ast.Expression, block *ast.BlockExpr) {
	if _, ok := iterable.(*ast.RangeLit); ok {
		g.genRangeEachBlock(iterable, block)
		return
	}

	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	g.buf.WriteString("runtime.Each(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" any) bool {\n")

	g.scopedVar(varName, "", func() {
		g.pushContext(ctxIterBlock)
		g.indent++
		for _, stmt := range block.Body {
			g.genStatement(stmt)
		}
		g.writeIndent()
		g.buf.WriteString("return true\n")
		g.indent--
		g.popContext()
	})

	g.writeIndent()
	g.buf.WriteString("})")
}

func (g *Generator) genRangeEachBlock(rangeExpr ast.Expression, block *ast.BlockExpr) {
	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	g.buf.WriteString("runtime.RangeEach(")
	g.genExpr(rangeExpr)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" int) bool {\n")

	g.scopedVar(varName, "Int", func() {
		g.pushContext(ctxIterBlock)
		g.indent++
		for _, stmt := range block.Body {
			g.genStatement(stmt)
		}
		g.writeIndent()
		g.buf.WriteString("return true\n")
		g.indent--
		g.popContext()
	})

	g.writeIndent()
	g.buf.WriteString("})")
}

func (g *Generator) genEachWithIndexBlock(iterable ast.Expression, block *ast.BlockExpr) {
	g.needsRuntime = true

	varName := "_"
	indexName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}
	if len(block.Params) > 1 {
		indexName = block.Params[1]
	}

	g.buf.WriteString("runtime.EachWithIndex(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" any, ")
	g.buf.WriteString(indexName)
	g.buf.WriteString(" int) bool {\n")

	g.scopedVars([]struct{ name, varType string }{
		{varName, ""},
		{indexName, "Int"},
	}, func() {
		g.pushContext(ctxIterBlock)
		g.indent++
		for _, stmt := range block.Body {
			g.genStatement(stmt)
		}
		g.writeIndent()
		g.buf.WriteString("return true\n")
		g.indent--
		g.popContext()
	})

	g.writeIndent()
	g.buf.WriteString("})")
}

func (g *Generator) genTimesBlock(times ast.Expression, block *ast.BlockExpr) {
	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	g.buf.WriteString("runtime.Times(")
	g.genExpr(times)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" int) bool {\n")

	g.scopedVar(varName, "Int", func() {
		g.pushContext(ctxIterBlock)
		g.indent++
		for _, stmt := range block.Body {
			g.genStatement(stmt)
		}
		g.writeIndent()
		g.buf.WriteString("return true\n")
		g.indent--
		g.popContext()
	})

	g.writeIndent()
	g.buf.WriteString("})")
}

func (g *Generator) genUptoBlock(start ast.Expression, block *ast.BlockExpr, args []ast.Expression) {
	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	g.buf.WriteString("runtime.Upto(")
	g.genExpr(start)
	g.buf.WriteString(", ")
	if len(args) > 0 {
		g.genExpr(args[0])
	} else {
		g.buf.WriteString("0")
	}
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" int) bool {\n")

	g.scopedVar(varName, "Int", func() {
		g.pushContext(ctxIterBlock)
		g.indent++
		for _, stmt := range block.Body {
			g.genStatement(stmt)
		}
		g.writeIndent()
		g.buf.WriteString("return true\n")
		g.indent--
		g.popContext()
	})

	g.writeIndent()
	g.buf.WriteString("})")
}

func (g *Generator) genDowntoBlock(start ast.Expression, block *ast.BlockExpr, args []ast.Expression) {
	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	g.buf.WriteString("runtime.Downto(")
	g.genExpr(start)
	g.buf.WriteString(", ")
	if len(args) > 0 {
		g.genExpr(args[0])
	} else {
		g.buf.WriteString("0")
	}
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" int) bool {\n")

	g.scopedVar(varName, "Int", func() {
		g.pushContext(ctxIterBlock)
		g.indent++
		for _, stmt := range block.Body {
			g.genStatement(stmt)
		}
		g.writeIndent()
		g.buf.WriteString("return true\n")
		g.indent--
		g.popContext()
	})

	g.writeIndent()
	g.buf.WriteString("})")
}

func (g *Generator) genScopedSpawnBlock(scope ast.Expression, block *ast.BlockExpr) {
	g.genExpr(scope)
	g.buf.WriteString(".Spawn(func() any {\n")

	g.indent++
	if len(block.Body) > 0 {
		for i, stmt := range block.Body {
			if i == len(block.Body)-1 {
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					g.writeIndent()
					g.buf.WriteString("return ")
					g.genExpr(exprStmt.Expr)
					g.buf.WriteString("\n")
				} else {
					g.genStatement(stmt)
					g.writeIndent()
					g.buf.WriteString("return nil\n")
				}
			} else {
				g.genStatement(stmt)
			}
		}
	} else {
		g.writeIndent()
		g.buf.WriteString("return nil\n")
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("})")
}

func (g *Generator) genSymbolToProcBlockCall(iterable ast.Expression, stp *ast.SymbolToProcExpr, method blockMethod) {
	g.needsRuntime = true
	g.buf.WriteString(method.runtimeFunc)
	g.buf.WriteString("(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(x any) (")
	g.buf.WriteString(method.returnType)
	g.buf.WriteString(", bool, bool) { return runtime.CallMethod(x, \"")
	g.buf.WriteString(stp.Method)
	g.buf.WriteString("\").(")
	g.buf.WriteString(method.returnType)
	g.buf.WriteString("), true, true })")
}

func (g *Generator) genRuntimeBlock(iterable ast.Expression, block *ast.BlockExpr, args []ast.Expression, method blockMethod) {
	g.needsRuntime = true

	var param1Name, param2Name string
	if method.hasAccumulator {
		param1Name = "acc"
		param2Name = "_"
		if len(block.Params) > 0 {
			param1Name = block.Params[0]
		}
		if len(block.Params) > 1 {
			param2Name = block.Params[1]
		}
	} else {
		param1Name = "_"
		if len(block.Params) > 0 {
			param1Name = block.Params[0]
		}
	}

	param1PrevType, param1WasDefinedBefore := g.vars[param1Name]
	param2PrevType, param2WasDefinedBefore := g.vars[param2Name]

	g.buf.WriteString(method.runtimeFunc)
	g.buf.WriteString("(")
	g.genExpr(iterable)

	if method.hasAccumulator {
		g.buf.WriteString(", ")
		if len(args) > 0 {
			g.genExpr(args[0])
		} else {
			g.buf.WriteString("nil")
		}
	}

	g.buf.WriteString(", func(")
	if method.hasAccumulator {
		g.buf.WriteString(param1Name)
		g.buf.WriteString(" any, ")
		g.buf.WriteString(param2Name)
		g.buf.WriteString(" any")
	} else {
		g.buf.WriteString(param1Name)
		g.buf.WriteString(" any")
	}
	g.buf.WriteString(") (")
	g.buf.WriteString(method.returnType)
	if method.usesIncludeFlag {
		g.buf.WriteString(", bool, bool) {\n")
	} else {
		g.buf.WriteString(", bool) {\n")
	}

	if param1Name != "_" {
		g.vars[param1Name] = ""
	}
	if param2Name != "_" {
		g.vars[param2Name] = ""
	}

	g.pushContextWithInclude(ctxTransformBlock, method.returnType, method.usesIncludeFlag)
	g.indent++

	if len(block.Body) > 0 {
		for _, stmt := range block.Body[:len(block.Body)-1] {
			g.genStatement(stmt)
		}
		lastStmt := block.Body[len(block.Body)-1]
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			g.writeIndent()
			g.buf.WriteString("return ")
			g.genExpr(exprStmt.Expr)
			if method.usesIncludeFlag {
				g.buf.WriteString(", true, true\n")
			} else {
				g.buf.WriteString(", true\n")
			}
		} else {
			g.genStatement(lastStmt)
			g.writeIndent()
			if method.usesIncludeFlag {
				g.buf.WriteString("return nil, false, true\n")
			} else {
				if method.returnType == "bool" {
					g.buf.WriteString("return false, true\n")
				} else {
					g.buf.WriteString("return nil, true\n")
				}
			}
		}
	} else {
		g.writeIndent()
		if method.usesIncludeFlag {
			g.buf.WriteString("return nil, false, true\n")
		} else {
			if method.returnType == "bool" {
				g.buf.WriteString("return false, true\n")
			} else {
				g.buf.WriteString("return nil, true\n")
			}
		}
	}

	g.indent--
	g.popContext()

	if param1Name != "_" {
		if !param1WasDefinedBefore {
			delete(g.vars, param1Name)
		} else {
			g.vars[param1Name] = param1PrevType
		}
	}
	if param2Name != "_" {
		if !param2WasDefinedBefore {
			delete(g.vars, param2Name)
		} else {
			g.vars[param2Name] = param2PrevType
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

func (g *Generator) genSelectorExpr(sel *ast.SelectorExpr) {
	// Check for channel methods: .receive, .try_receive, .close
	switch sel.Sel {
	case "receive":
		// ch.receive -> <-ch
		g.buf.WriteString("<-")
		g.genExpr(sel.X)
		return
	case "try_receive":
		// ch.try_receive -> runtime.TryReceive(ch)
		g.needsRuntime = true
		g.buf.WriteString("runtime.TryReceive(")
		g.genExpr(sel.X)
		g.buf.WriteString(")")
		return
	case "close":
		// ch.close -> close(ch)
		g.buf.WriteString("close(")
		g.genExpr(sel.X)
		g.buf.WriteString(")")
		return
	case "await":
		// task.await -> runtime.Await(task)
		g.needsRuntime = true
		g.buf.WriteString("runtime.Await(")
		g.genExpr(sel.X)
		g.buf.WriteString(")")
		return
	case "new":
		// Chan[T].new -> make(chan T)
		if indexExpr, ok := sel.X.(*ast.IndexExpr); ok {
			if chanIdent, ok := indexExpr.Left.(*ast.Ident); ok && chanIdent.Name == "Chan" {
				var chanType string
				if typeIdent, ok := indexExpr.Index.(*ast.Ident); ok {
					chanType = typeIdent.Name
				} else {
					chanType = "any"
				}
				goType := mapType(chanType)
				g.buf.WriteString("make(chan ")
				g.buf.WriteString(goType)
				g.buf.WriteString(")")
				return
			}
		}
	}

	// Check for range methods used without parentheses
	if _, ok := sel.X.(*ast.RangeLit); ok {
		switch sel.Sel {
		case "to_a":
			g.needsRuntime = true
			g.buf.WriteString("runtime.RangeToArray(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return
		case "size", "length":
			g.needsRuntime = true
			g.buf.WriteString("runtime.RangeSize(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return
		}
	}

	// Handle .length and .size -> len()
	if sel.Sel == "length" || sel.Sel == "size" {
		g.buf.WriteString("len(")
		g.genExpr(sel.X)
		g.buf.WriteString(")")
		return
	}

	// Check for methods on optional types (ok?, nil?, present?, absent?, unwrap)
	receiverType := g.inferTypeFromExpr(sel.X)
	if isOptionalType(receiverType) {
		switch sel.Sel {
		case "ok?", "present?":
			g.buf.WriteString("(")
			g.genExpr(sel.X)
			g.buf.WriteString(" != nil)")
			return
		case "nil?", "absent?":
			g.buf.WriteString("(")
			g.genExpr(sel.X)
			g.buf.WriteString(" == nil)")
			return
		case "unwrap":
			g.buf.WriteString("*")
			g.genExpr(sel.X)
			return
		}
	}

	if g.isGoInterop(sel.X) {
		g.genExpr(sel.X)
		g.buf.WriteString(".")
		g.buf.WriteString(snakeToPascalWithAcronyms(sel.Sel))
		return
	}

	g.genExpr(sel.X)
	g.buf.WriteString(".")
	g.buf.WriteString(snakeToCamelWithAcronyms(sel.Sel))
}

func (g *Generator) isGoInterop(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return g.imports[e.Name]
	case *ast.SelectorExpr:
		return g.isGoInterop(e.X)
	default:
		return false
	}
}

func (g *Generator) genSpawnExpr(e *ast.SpawnExpr) {
	g.needsRuntime = true
	g.buf.WriteString("runtime.Spawn(func() any {\n")
	g.indent++
	if len(e.Block.Body) > 0 {
		for i, stmt := range e.Block.Body {
			if i == len(e.Block.Body)-1 {
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					g.writeIndent()
					g.buf.WriteString("return ")
					g.genExpr(exprStmt.Expr)
					g.buf.WriteString("\n")
				} else {
					g.genStatement(stmt)
					g.writeIndent()
					g.buf.WriteString("return nil\n")
				}
			} else {
				g.genStatement(stmt)
			}
		}
	} else {
		g.writeIndent()
		g.buf.WriteString("return nil\n")
	}
	g.indent--
	g.writeIndent()
	g.buf.WriteString("})")
}

func (g *Generator) genAwaitExpr(e *ast.AwaitExpr) {
	g.needsRuntime = true
	g.buf.WriteString("runtime.Await(")
	g.genExpr(e.Task)
	g.buf.WriteString(")")
}
