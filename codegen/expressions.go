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
		} else if g.currentClass != "" && g.currentClassModuleMethods[e.Name] {
			// Module method call within class - generate as self.Method() (PascalCase for interface)
			recv := receiverName(g.currentClass)
			methodName := snakeToPascalWithAcronyms(e.Name)
			g.buf.WriteString(fmt.Sprintf("%s.%s()", recv, methodName))
		} else if g.noArgFunctions[e.Name] {
			// No-arg function - call it implicitly (Ruby-style)
			goName := snakeToCamelWithAcronyms(e.Name)
			g.buf.WriteString(fmt.Sprintf("%s()", goName))
		} else {
			g.buf.WriteString(e.Name)
		}
	case *ast.InstanceVar:
		if g.currentClass != "" {
			recv := receiverName(g.currentClass)
			// Use underscore prefix for accessor fields to match struct definition
			// Check both current class and parent classes for accessor fields
			goFieldName := e.Name
			if g.isAccessorField(e.Name) {
				goFieldName = "_" + e.Name
			}
			g.buf.WriteString(fmt.Sprintf("%s.%s", recv, goFieldName))
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
		// Use SelectorKind to determine if this is a field access or method call
		selectorKind := g.getSelectorKind(e)
		switch selectorKind {
		case ast.SelectorMethod, ast.SelectorGetter:
			// Method calls need () - convert to CallExpr
			g.genCallExpr(&ast.CallExpr{Func: e, Args: nil})
		case ast.SelectorGoMethod:
			// Go method calls need () with PascalCase naming
			g.genExpr(e.X)
			g.buf.WriteString(".")
			g.buf.WriteString(snakeToPascalWithAcronyms(e.Sel))
			g.buf.WriteString("()")
		default:
			// Field access or unknown - use existing logic
			g.genSelectorExpr(e)
		}
	case *ast.IndexExpr:
		g.genIndexExpr(e)
	case *ast.SafeNavExpr:
		g.genSafeNavExpr(e)

	// Concurrency
	case *ast.SpawnExpr:
		g.genSpawnExpr(e)
	case *ast.AwaitExpr:
		g.genAwaitExpr(e)
	case *ast.ConcurrentlyExpr:
		g.genConcurrentlyExpr(e)

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

	// Handle super in constructor (initialize) specially:
	// Generate: recv.ParentClass = *newParentClass(args...)
	if g.currentMethod == "initialize" {
		// Determine parent constructor name based on visibility
		// For now, assume parent constructor has same visibility as parent class
		constructorName := "new" + parentClass
		if g.pubClasses[parentClass] {
			constructorName = "New" + parentClass
		}
		g.buf.WriteString(fmt.Sprintf("%s.%s = *%s(", recv, parentClass, constructorName))
		for i, arg := range e.Args {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.genExpr(arg)
		}
		g.buf.WriteString(")")
		return
	}

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
	// Check if we can use direct comparison instead of runtime.Equal
	canUseDirectEqual := g.canUseDirectComparison(e.Left, e.Right)

	if e.Op == "==" && !canUseDirectEqual {
		g.needsRuntime = true
		g.buf.WriteString("runtime.Equal(")
		g.genExpr(e.Left)
		g.buf.WriteString(", ")
		g.genExpr(e.Right)
		g.buf.WriteString(")")
		return
	}
	if e.Op == "!=" && !canUseDirectEqual {
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

	// Handle arithmetic on any-typed values (e.g., from await)
	if e.Op == "+" && g.hasAnyTypedOperand(e.Left, e.Right) {
		g.needsRuntime = true
		g.buf.WriteString("runtime.Add(")
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

// hasAnyTypedOperand returns true if BOTH operands have type any (interface{})
// We require both because if only one is any, it's likely due to generic typing
// (like reduce's accumulator) where codegen infers concrete types.
func (g *Generator) hasAnyTypedOperand(left, right ast.Expression) bool {
	if g.typeInfo == nil {
		return false
	}
	leftKind := g.typeInfo.GetTypeKind(left)
	rightKind := g.typeInfo.GetTypeKind(right)
	return leftKind == TypeAny && rightKind == TypeAny
}

// canUseDirectComparison returns true if we can use Go's == operator directly
// instead of runtime.Equal. This is possible when:
// 1. Both operands are primitive literals, OR
// 2. We have type info and both operands are primitive types (Int, Float, String, Bool)
func (g *Generator) canUseDirectComparison(left, right ast.Expression) bool {
	// Primitive literals can always use direct comparison
	if isPrimitiveLiteral(left) && isPrimitiveLiteral(right) {
		return true
	}

	// If we have type info, check if both types are primitive
	if g.typeInfo != nil {
		leftKind := g.typeInfo.GetTypeKind(left)
		rightKind := g.typeInfo.GetTypeKind(right)
		return isPrimitiveKind(leftKind) && isPrimitiveKind(rightKind)
	}

	return false
}

// isPrimitiveKind returns true for types that support direct == comparison in Go.
func isPrimitiveKind(k TypeKind) bool {
	switch k {
	case TypeInt, TypeInt64, TypeFloat, TypeBool, TypeString:
		return true
	default:
		return false
	}
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

	// Check if the left side is a safe navigation expression (returns any)
	_, isSafeNav := e.Left.(*ast.SafeNavExpr)

	switch leftType {
	case "Int?":
		g.buf.WriteString("runtime.CoalesceInt(")
	case "Int64?":
		g.buf.WriteString("runtime.CoalesceInt64(")
	case "Float?":
		g.buf.WriteString("runtime.CoalesceFloat(")
	case "String?":
		// Use CoalesceStringAny for safe navigation (which returns any)
		if isSafeNav {
			g.buf.WriteString("runtime.CoalesceStringAny(")
		} else {
			g.buf.WriteString("runtime.CoalesceString(")
		}
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
	// obj&.method -> func() any { _snN := obj; if _snN != nil { return (*_snN).method() } else { return nil } }()
	varName := fmt.Sprintf("_sn%d", g.tempVarCounter)
	g.tempVarCounter++

	g.buf.WriteString("func() any { ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := ")

	// If the receiver is a selector expression, we need to call it as a method.
	// Safe navigation is used with optional-returning getters/methods, so the receiver
	// must be called (not just referenced as a field).
	if sel, ok := e.Receiver.(*ast.SelectorExpr); ok {
		// Check if genExpr will already add () for this selector (getter/method)
		selectorKind := g.getSelectorKind(sel)
		g.genExpr(e.Receiver)
		// Only add () if genExpr didn't already handle it.
		// genExpr adds () for SelectorMethod, SelectorGetter, and SelectorGoMethod.
		// Underscore-prefixed selectors (e.g., _field) are raw field accesses that
		// bypass the getter mechanism and don't need ().
		if selectorKind != ast.SelectorMethod && selectorKind != ast.SelectorGetter &&
			selectorKind != ast.SelectorGoMethod && !strings.HasPrefix(sel.Sel, "_") {
			g.buf.WriteString("()")
		}
	} else {
		g.genExpr(e.Receiver)
	}

	g.buf.WriteString("; if ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" != nil { return ")

	// Use (*var).selector() - call it as a method
	g.buf.WriteString("(*")
	g.buf.WriteString(varName)
	g.buf.WriteString(").")
	g.buf.WriteString(snakeToCamelWithAcronyms(e.Selector))
	g.buf.WriteString("()")

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

	// Determine array type using semantic type info or fallback inference
	arrayType := g.getArrayType(arr)

	g.buf.WriteString(arrayType + "{")
	for i, elem := range arr.Elements {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.genExpr(elem)
	}
	g.buf.WriteString("}")
}

// getArrayType returns the Go type string for an array literal (e.g., "[]int").
// Uses semantic type info when available, falls back to expression inference.
func (g *Generator) getArrayType(arr *ast.ArrayLit) string {
	// Try semantic type info first (most reliable)
	if g.typeInfo != nil {
		if goType := g.typeInfo.GetGoType(arr); goType != "" {
			return goType
		}
	}

	// Fall back to expression-based inference
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
	return "[]" + elemType
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
		// For "any" type receivers, use runtime.GetKey to handle type assertion
		leftType := g.inferTypeFromExpr(idx.Left)
		if leftType == "any" {
			g.needsRuntime = true
			g.buf.WriteString("runtime.GetKey(")
			g.genExpr(idx.Left)
			g.buf.WriteString(", ")
			g.genExpr(idx.Index)
			g.buf.WriteString(")")
			return
		}

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

	// For strings, always use runtime.AtIndex to return characters (not bytes)
	leftType := g.inferTypeFromExpr(idx.Left)
	if leftType == "string" || leftType == "String" {
		g.needsRuntime = true
		g.buf.WriteString("runtime.AtIndex(")
		g.genExpr(idx.Left)
		g.buf.WriteString(", ")
		g.genExpr(idx.Index)
		g.buf.WriteString(")")
		return
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

	// Determine map type using semantic type info or fallback inference
	goMapType := g.getMapType(m)

	g.buf.WriteString(goMapType + "{")
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

// getMapType returns the Go type string for a map literal (e.g., "map[string]int").
// Uses semantic type info when available, falls back to expression inference.
func (g *Generator) getMapType(m *ast.MapLit) string {
	// Try semantic type info first (most reliable)
	if g.typeInfo != nil {
		if goType := g.typeInfo.GetGoType(m); goType != "" {
			return goType
		}
	}

	// Fall back to expression-based inference
	if len(m.Entries) > 0 {
		keyType := g.inferTypeFromExpr(m.Entries[0].Key)
		valType := g.inferTypeFromExpr(m.Entries[0].Value)
		keysConsistent := true
		valsConsistent := true

		for _, e := range m.Entries[1:] {
			if g.inferTypeFromExpr(e.Key) != keyType {
				keysConsistent = false
			}
			if g.inferTypeFromExpr(e.Value) != valType {
				valsConsistent = false
			}
		}

		goKeyType := "any"
		goValType := "any"
		if keysConsistent && keyType != "" {
			goKeyType = mapType(keyType)
		}
		if valsConsistent && valType != "" {
			goValType = mapType(valType)
		}
		return "map[" + goKeyType + "]" + goValType
	}
	return "map[any]any"
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
					// Use mapParamType to properly convert class types to pointers
					goType := g.mapParamType(chanType)

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
			// Only treat as Rugby class constructor if NOT a Go import
			// (errors.new should become errors.New, not newerrors)
			if ident, ok := fn.X.(*ast.Ident); ok && !g.imports[ident.Name] {
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
			// Handle Go package type .new calls like sync.WaitGroup.new
			// Generate zero-value initialization: &pkg.Type{}
			if sel, ok := fn.X.(*ast.SelectorExpr); ok {
				if pkgIdent, ok := sel.X.(*ast.Ident); ok && g.imports[pkgIdent.Name] {
					g.buf.WriteString("&")
					g.genExpr(fn.X)
					g.buf.WriteString("{}")
					return
				}
			}
		}

		// Handle class method calls (ClassName.method_name)
		if ident, ok := fn.X.(*ast.Ident); ok && g.classes[ident.Name] {
			if methodMap, hasClass := g.classMethods[ident.Name]; hasClass {
				if funcName, hasMethod := methodMap[fn.Sel]; hasMethod {
					g.buf.WriteString(funcName)
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

		// Handle array/slice method calls FIRST, before stdLib lookup
		recvTypeForArray := g.inferTypeFromExpr(fn.X)
		if strings.HasPrefix(recvTypeForArray, "[]") || recvTypeForArray == "Array" || strings.HasPrefix(recvTypeForArray, "Array[") {
			elemType := g.inferArrayElementGoType(fn.X)
			switch fn.Sel {
			case "any?":
				// any? without a block returns true if array has any elements
				g.buf.WriteString("(len(")
				g.genExpr(fn.X)
				g.buf.WriteString(") > 0)")
				return
			case "empty?":
				// empty? returns true if array has no elements
				g.buf.WriteString("(len(")
				g.genExpr(fn.X)
				g.buf.WriteString(") == 0)")
				return
			case "shift":
				// shift removes and returns the first element, needs pointer
				g.needsRuntime = true
				switch elemType {
				case "int":
					g.buf.WriteString("runtime.ShiftInt(&")
				case "string":
					g.buf.WriteString("runtime.ShiftString(&")
				case "float64":
					g.buf.WriteString("runtime.ShiftFloat(&")
				default:
					g.buf.WriteString("runtime.ShiftAny(&")
				}
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "pop":
				// pop removes and returns the last element, needs pointer
				g.needsRuntime = true
				switch elemType {
				case "int":
					g.buf.WriteString("runtime.PopInt(&")
				case "string":
					g.buf.WriteString("runtime.PopString(&")
				case "float64":
					g.buf.WriteString("runtime.PopFloat(&")
				default:
					g.buf.WriteString("runtime.PopAny(&")
				}
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "sum":
				g.needsRuntime = true
				if elemType == "float64" {
					g.buf.WriteString("runtime.SumFloat(")
				} else {
					g.buf.WriteString("runtime.SumInt(")
				}
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "min":
				g.needsRuntime = true
				if elemType == "float64" {
					g.buf.WriteString("runtime.MinFloatPtr(")
				} else {
					g.buf.WriteString("runtime.MinIntPtr(")
				}
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "max":
				g.needsRuntime = true
				if elemType == "float64" {
					g.buf.WriteString("runtime.MaxFloatPtr(")
				} else {
					g.buf.WriteString("runtime.MaxIntPtr(")
				}
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "first":
				g.needsRuntime = true
				g.buf.WriteString("runtime.FirstPtr(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "last":
				g.needsRuntime = true
				g.buf.WriteString("runtime.LastPtr(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "sorted":
				g.needsRuntime = true
				g.buf.WriteString("runtime.Sort(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "include?", "contains?":
				g.needsRuntime = true
				g.buf.WriteString("runtime.Contains(")
				g.genExpr(fn.X)
				if len(call.Args) > 0 {
					g.buf.WriteString(", ")
					g.genExpr(call.Args[0])
				}
				g.buf.WriteString(")")
				return
			case "length", "size":
				// Array length uses Go's built-in len()
				g.buf.WriteString("len(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			}
		}

		recvType := g.inferTypeFromExpr(fn.X)
		// Normalize Go types to Rugby type names for stdLib lookup
		lookupType := recvType
		switch recvType {
		case "int", "int64":
			lookupType = "Int"
		case "float64":
			lookupType = "Float"
		case "string":
			lookupType = "String"
		}

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

		if lookupType != "" {
			if methods, ok := stdLib[lookupType]; ok {
				if def, ok := methods[fn.Sel]; ok {
					methodDef = def
					found = true
				}
			}
		}

		// Only check uniqueMethods if this isn't a Go interop call
		// Otherwise strings.split() would incorrectly match the String.split method
		if !found && !g.isGoInterop(fn.X) {
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
			// Cast to any first to allow type assertion on concrete types
			g.buf.WriteString("func() bool { _, ok := any(")
			g.genExpr(fn.X)
			g.buf.WriteString(").(")
			g.genExpr(call.Args[0])
			g.buf.WriteString("); return ok }()")
			return
		}

		if fn.Sel == "as" {
			if len(call.Args) != 1 {
				g.buf.WriteString("func() (any, bool) { return nil, false }() /* ERROR: as requires exactly one type argument */")
				return
			}
			// Cast to any first to allow type assertion on concrete types
			g.buf.WriteString("func() (")
			g.genExpr(call.Args[0])
			g.buf.WriteString(", bool) { v, ok := any(")
			g.genExpr(fn.X)
			g.buf.WriteString(").(")
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

		// Handle range method calls - check both literal ranges and variables of type Range
		isRangeExpr := false
		if _, ok := fn.X.(*ast.RangeLit); ok {
			isRangeExpr = true
		} else if recvType := g.inferTypeFromExpr(fn.X); recvType == "Range" {
			isRangeExpr = true
		}
		if isRangeExpr {
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

	// Check if receiver is an optional type - needs special handling
	receiverType := g.inferTypeFromExpr(sel.X)
	if isOptionalType(receiverType) {
		switch method {
		case "map":
			g.genOptionalMapBlock(sel.X, block, receiverType)
			return
		case "each":
			g.genOptionalEachBlock(sel.X, block, receiverType)
			return
		}
	}

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

// genOptionalMapBlock generates code for optional.map { |x| ... }
// Returns *ResultType (pointer) to maintain optional semantics for nil coalescing.
func (g *Generator) genOptionalMapBlock(receiver ast.Expression, block *ast.BlockExpr, optType string) {
	varName := "_optVal"
	blockParam := "_"
	if len(block.Params) > 0 {
		blockParam = block.Params[0]
	}

	// Determine the inner type (without ?)
	innerType := strings.TrimSuffix(optType, "?")
	goType := mapType(innerType)

	// Infer the block's return type from its last expression
	blockReturnGoType := "string" // default fallback
	if len(block.Body) > 0 {
		if exprStmt, ok := block.Body[len(block.Body)-1].(*ast.ExprStmt); ok {
			if inferredType := g.inferTypeFromExpr(exprStmt.Expr); inferredType != "" {
				blockReturnGoType = mapType(inferredType)
			}
		}
	}

	// Generate: func() *ResultType { if opt := receiver; opt != nil { result := block(*opt); return &result }; return nil }()
	g.buf.WriteString("func() *")
	g.buf.WriteString(blockReturnGoType)
	g.buf.WriteString(" { if ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := ")
	g.genExpr(receiver)
	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" != nil { ")
	g.buf.WriteString(blockParam)
	g.buf.WriteString(" := *")
	g.buf.WriteString(varName)
	// Declare _ to avoid unused variable if blockParam is used
	if blockParam != "_" {
		g.buf.WriteString("; _ = ")
		g.buf.WriteString(blockParam)
	}
	g.buf.WriteString("; _result := ")

	// Generate block body - last expression is the result
	if len(block.Body) > 0 {
		lastStmt := block.Body[len(block.Body)-1]
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			g.genExpr(exprStmt.Expr)
		} else {
			g.buf.WriteString("\"\"") // fallback
		}
	} else {
		g.buf.WriteString("\"\"")
	}

	g.buf.WriteString("; return &_result }; return nil }()")
	_ = goType // inner type used for block param binding
}

// genOptionalEachBlock generates code for optional.each { |x| ... }
func (g *Generator) genOptionalEachBlock(receiver ast.Expression, block *ast.BlockExpr, optType string) {
	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	// Determine the inner type (without ?)
	innerType := strings.TrimSuffix(optType, "?")
	goType := mapType(innerType)

	g.buf.WriteString("runtime.OptionalEachString(")
	g.genExpr(receiver)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" ")
	g.buf.WriteString(goType)
	g.buf.WriteString(") {\n")

	g.indent++
	for _, stmt := range block.Body {
		g.genStatement(stmt)
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("})")
}

func (g *Generator) genEachBlock(iterable ast.Expression, block *ast.BlockExpr) {
	if _, ok := iterable.(*ast.RangeLit); ok {
		g.genRangeEachBlock(iterable, block)
		return
	}

	// Check if iterating over a map with two parameters (key, value)
	// Use semantic analysis for Go type if available
	var goType string
	if g.typeInfo != nil {
		goType = g.typeInfo.GetGoType(iterable)
	}
	if goType == "" {
		goType = g.inferTypeFromExpr(iterable)
	}

	// Check for map iteration with two parameters
	isMapIteration := len(block.Params) >= 2 && (strings.HasPrefix(goType, "map[") || goType == "Map")

	if isMapIteration {
		g.needsRuntime = true
		// If we have semantic type info with specific map type, use it; otherwise infer
		if strings.HasPrefix(goType, "map[") {
			g.genMapEachBlock(iterable, block, goType)
		} else {
			// Fall back to any types for untyped maps
			g.genMapEachBlock(iterable, block, "map[any]any")
		}
		return
	}

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	// Infer element type for proper type inference in the loop
	elemType := g.inferArrayElementGoType(iterable)

	// Generate a native for-range loop instead of runtime.Each
	// This allows proper type inference for block parameters
	g.buf.WriteString("for _, ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := range ")
	g.genExpr(iterable)
	g.buf.WriteString(" {\n")

	g.scopedVar(varName, elemType, func() {
		g.indent++
		for _, stmt := range block.Body {
			g.genStatement(stmt)
		}
		g.indent--
	})

	g.writeIndent()
	g.buf.WriteString("}")
}

func (g *Generator) genMapEachBlock(iterable ast.Expression, block *ast.BlockExpr, mapType string) {
	keyName := block.Params[0]
	valueName := block.Params[1]

	// Extract key and value types from map[K]V
	keyType, valueType := g.parseMapType(mapType)

	g.buf.WriteString("runtime.MapEach(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(keyName)
	g.buf.WriteString(" ")
	g.buf.WriteString(keyType)
	g.buf.WriteString(", ")
	g.buf.WriteString(valueName)
	g.buf.WriteString(" ")
	g.buf.WriteString(valueType)
	g.buf.WriteString(") {\n")

	prevKeyType, keyWasDefinedBefore := g.vars[keyName]
	prevValueType, valueWasDefinedBefore := g.vars[valueName]
	g.vars[keyName] = keyType
	g.vars[valueName] = valueType

	g.indent++
	for _, stmt := range block.Body {
		g.genStatement(stmt)
	}
	g.indent--

	if keyWasDefinedBefore {
		g.vars[keyName] = prevKeyType
	} else {
		delete(g.vars, keyName)
	}
	if valueWasDefinedBefore {
		g.vars[valueName] = prevValueType
	} else {
		delete(g.vars, valueName)
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// parseMapType extracts key and value types from a Go map type string like "map[string]int"
func (g *Generator) parseMapType(mapType string) (keyType, valueType string) {
	// map[string]int -> keyType="string", valueType="int"
	if !strings.HasPrefix(mapType, "map[") {
		return "any", "any"
	}
	rest := mapType[4:] // Skip "map["
	bracketCount := 1
	keyEnd := 0
	for i, c := range rest {
		if c == '[' {
			bracketCount++
		} else if c == ']' {
			bracketCount--
			if bracketCount == 0 {
				keyEnd = i
				break
			}
		}
	}
	if keyEnd == 0 {
		return "any", "any"
	}
	keyType = rest[:keyEnd]
	valueType = rest[keyEnd+1:]
	return keyType, valueType
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
	g.buf.WriteString(" int) {\n")

	g.scopedVar(varName, "Int", func() {
		g.indent++
		for _, stmt := range block.Body {
			g.genStatement(stmt)
		}
		g.indent--
	})

	g.writeIndent()
	g.buf.WriteString("})")
}

func (g *Generator) genEachWithIndexBlock(iterable ast.Expression, block *ast.BlockExpr) {
	varName := "_"
	indexName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}
	if len(block.Params) > 1 {
		indexName = block.Params[1]
	}

	// Infer element type for proper type inference in the loop
	elemType := g.inferArrayElementGoType(iterable)

	// Generate a native for-range loop instead of runtime.EachWithIndex
	// This allows proper type inference for block parameters
	g.buf.WriteString("for ")
	g.buf.WriteString(indexName)
	g.buf.WriteString(", ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := range ")
	g.genExpr(iterable)
	g.buf.WriteString(" {\n")

	g.scopedVars([]struct{ name, varType string }{
		{varName, elemType},
		{indexName, "int"},
	}, func() {
		g.indent++
		for _, stmt := range block.Body {
			g.genStatement(stmt)
		}
		g.indent--
	})

	g.writeIndent()
	g.buf.WriteString("}")
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
	g.buf.WriteString(" int) {\n")

	g.scopedVar(varName, "Int", func() {
		g.indent++
		for _, stmt := range block.Body {
			g.genStatement(stmt)
		}
		g.indent--
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
	g.buf.WriteString(" int) {\n")

	g.scopedVar(varName, "Int", func() {
		g.indent++
		for _, stmt := range block.Body {
			g.genStatement(stmt)
		}
		g.indent--
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
	g.buf.WriteString(" int) {\n")

	g.scopedVar(varName, "Int", func() {
		g.indent++
		for _, stmt := range block.Body {
			g.genStatement(stmt)
		}
		g.indent--
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
	elemType := g.inferArrayElementGoType(iterable)
	g.buf.WriteString(method.runtimeFunc)
	g.buf.WriteString("(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(x ")
	g.buf.WriteString(elemType)
	g.buf.WriteString(") ")
	g.buf.WriteString(method.returnType)
	g.buf.WriteString(" { return runtime.CallMethod(x, \"")
	g.buf.WriteString(stp.Method)
	g.buf.WriteString("\")")
	// Add type assertion for non-any return types
	if method.returnType != "any" {
		g.buf.WriteString(".(")
		g.buf.WriteString(method.returnType)
		g.buf.WriteString(")")
	}
	g.buf.WriteString(" })")
}

func (g *Generator) genRuntimeBlock(iterable ast.Expression, block *ast.BlockExpr, args []ast.Expression, method blockMethod) {
	g.needsRuntime = true

	// Infer element type for the block parameter
	elemType := g.inferArrayElementGoType(iterable)

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
	returnType := method.returnType
	if method.hasAccumulator {
		// For reduce: infer acc type from initial value, element is inferred from iterable
		accType := "any"
		if len(args) > 0 {
			accType = mapType(g.inferTypeFromExpr(args[0]))
		}
		// Return type matches accumulator type
		returnType = accType
		g.buf.WriteString(param1Name)
		g.buf.WriteString(" ")
		g.buf.WriteString(accType)
		g.buf.WriteString(", ")
		g.buf.WriteString(param2Name)
		g.buf.WriteString(" ")
		g.buf.WriteString(elemType)
	} else {
		// For map/select/etc: element is inferred type
		g.buf.WriteString(param1Name)
		g.buf.WriteString(" ")
		g.buf.WriteString(elemType)
	}
	g.buf.WriteString(") ")
	g.buf.WriteString(returnType)
	g.buf.WriteString(" {\n")

	if param1Name != "_" {
		g.vars[param1Name] = ""
	}
	if param2Name != "_" {
		g.vars[param2Name] = ""
	}

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
			g.buf.WriteString("\n")
		} else {
			g.genStatement(lastStmt)
			g.writeIndent()
			if method.returnType == "bool" {
				g.buf.WriteString("return false\n")
			} else {
				g.buf.WriteString("return nil\n")
			}
		}
	} else {
		g.writeIndent()
		if method.returnType == "bool" {
			g.buf.WriteString("return false\n")
		} else {
			g.buf.WriteString("return nil\n")
		}
	}

	g.indent--

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
				// Use mapParamType to properly convert class types to pointers
				goType := g.mapParamType(chanType)
				g.buf.WriteString("make(chan ")
				g.buf.WriteString(goType)
				g.buf.WriteString(")")
				return
			}
		}
		// Handle Go package type .new calls like sync.WaitGroup.new
		// Generate zero-value initialization: &pkg.Type{}
		if pkgSel, ok := sel.X.(*ast.SelectorExpr); ok {
			if pkgIdent, ok := pkgSel.X.(*ast.Ident); ok && g.imports[pkgIdent.Name] {
				g.buf.WriteString("&")
				g.buf.WriteString(pkgIdent.Name)
				g.buf.WriteString(".")
				g.buf.WriteString(pkgSel.Sel)
				g.buf.WriteString("{}")
				return
			}
		}
	}

	// Handle class method calls (ClassName.method_name) - implicit call without parens
	if ident, ok := sel.X.(*ast.Ident); ok && g.classes[ident.Name] {
		if methodMap, hasClass := g.classMethods[ident.Name]; hasClass {
			if funcName, hasMethod := methodMap[sel.Sel]; hasMethod {
				// Generate function call: ClassName_methodName()
				g.buf.WriteString(funcName)
				g.buf.WriteString("()")
				return
			}
		}
	}

	// Check for range methods used without parentheses
	// Handle both literal ranges and variables of type Range
	isRangeExpr := false
	if _, ok := sel.X.(*ast.RangeLit); ok {
		isRangeExpr = true
	} else if recvType := g.inferTypeFromExpr(sel.X); recvType == "Range" {
		isRangeExpr = true
	}
	if isRangeExpr {
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

	// Fallback for OptionalResult methods (from optional.map operations)
	// These use PascalCase methods: Ok(), Unwrap()
	// Note: ok? is typically wrapped in CallExpr (predicate methods), so don't add ()
	// But unwrap is used as a property and needs () to call the method
	if sel.Sel == "ok?" {
		g.genExpr(sel.X)
		g.buf.WriteString(".Ok")
		return
	}
	if sel.Sel == "unwrap" {
		g.genExpr(sel.X)
		g.buf.WriteString(".Unwrap()")
		return
	}

	// Handle property-style array/map method calls
	if g.genArrayPropertyMethod(sel) {
		return
	}

	// Handle property-style stdLib method calls (Int.even?, Float.floor, etc.)
	if g.genStdLibPropertyMethod(sel) {
		return
	}

	if g.isGoInterop(sel.X) {
		g.genExpr(sel.X)
		g.buf.WriteString(".")
		g.buf.WriteString(snakeToPascalWithAcronyms(sel.Sel))
		return
	}

	// Check if this method name matches any interface method
	// If so, use PascalCase for Go interface compatibility
	isInterfaceMethod := g.isInterfaceMethod(sel.Sel)

	// Check if receiver is an instance of a pub class
	// Pub class methods use PascalCase
	receiverClassName := g.getReceiverClassName(sel.X)
	isPubClass := g.pubClasses[receiverClassName]

	g.genExpr(sel.X)
	g.buf.WriteString(".")

	// Special case: to_s maps to String (Go's Stringer interface convention)
	if sel.Sel == "to_s" {
		g.buf.WriteString("String")
		return
	}

	if isInterfaceMethod || isPubClass {
		g.buf.WriteString(snakeToPascalWithAcronyms(sel.Sel))
	} else {
		g.buf.WriteString(snakeToCamelWithAcronyms(sel.Sel))
	}
}

func (g *Generator) isGoInterop(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		// Check both imports and variables holding Go interop types
		return g.imports[e.Name] || g.goInteropVars[e.Name]
	case *ast.SelectorExpr:
		return g.isGoInterop(e.X)
	default:
		return false
	}
}

// isGoInteropTypeConstructor checks if an expression is a Go type constructor like sync.WaitGroup.new
// Returns true if the expression creates an instance of a Go type.
func (g *Generator) isGoInteropTypeConstructor(expr ast.Expression) bool {
	// First check for CallExpr (sync.WaitGroup.new())
	if call, ok := expr.(*ast.CallExpr); ok {
		sel, ok := call.Func.(*ast.SelectorExpr)
		if !ok || sel.Sel != "new" {
			return false
		}
		expr = sel // Continue checking the selector
	}

	// Check for SelectorExpr (sync.WaitGroup.new - property style)
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel != "new" {
		return false
	}
	// Check if it's pkg.Type.new pattern
	innerSel, ok := sel.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	// Check if the package is a Go import
	pkgIdent, ok := innerSel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return g.imports[pkgIdent.Name]
}

// isGoInteropCall checks if an expression is a function call from a Go package.
// Examples: http.get(url), json.parse(data), etc.
// Also handles BangExpr (!) wrapping: http.get(url)!
func (g *Generator) isGoInteropCall(expr ast.Expression) bool {
	// Unwrap BangExpr (!)
	if bang, ok := expr.(*ast.BangExpr); ok {
		expr = bang.Expr
	}

	// Check for CallExpr
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	// Check if function is pkg.func pattern
	sel, ok := call.Func.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Check if the receiver is a Go import
	return g.isGoInterop(sel.X)
}

// isInterfaceMethod checks if a method name matches any interface method in the program.
// This is used to determine whether to use PascalCase (for interface methods) or camelCase.
func (g *Generator) isInterfaceMethod(methodName string) bool {
	for _, methods := range g.interfaceMethods {
		if methods[methodName] {
			return true
		}
	}
	return false
}

// getReceiverClassName extracts the class name from a receiver expression.
// For example, if expr is a variable of type *User, returns "User".
// Returns empty string if the class name cannot be determined.
func (g *Generator) getReceiverClassName(expr ast.Expression) string {
	// Use type info if available
	if g.typeInfo != nil {
		if typ := g.typeInfo.GetRugbyType(expr); typ != "" {
			// Strip pointer/optional prefixes if present
			typ = strings.TrimPrefix(typ, "*")
			typ = strings.TrimSuffix(typ, "?")
			return typ
		}
	}

	// Fall back to inferring from expression
	switch e := expr.(type) {
	case *ast.Ident:
		// Look up variable type
		if varType, ok := g.vars[e.Name]; ok {
			varType = strings.TrimPrefix(varType, "*")
			varType = strings.TrimSuffix(varType, "?")
			return varType
		}
	case *ast.CallExpr:
		// Constructor call like User.new() -> "User"
		if sel, ok := e.Func.(*ast.SelectorExpr); ok && sel.Sel == "new" {
			if ident, ok := sel.X.(*ast.Ident); ok {
				return ident.Name
			}
		}
	}
	return ""
}

// genArrayPropertyMethod handles property-style array and map method calls
// like nums.sum, nums.first, m.keys, etc.
// Returns true if the method was handled, false otherwise.
func (g *Generator) genArrayPropertyMethod(sel *ast.SelectorExpr) bool {
	receiverType := g.inferTypeFromExpr(sel.X)
	elemType := g.inferArrayElementGoType(sel.X)

	// Handle array property methods
	if strings.HasPrefix(receiverType, "[]") || receiverType == "Array" || strings.HasPrefix(receiverType, "Array[") {
		switch sel.Sel {
		case "sum":
			g.needsRuntime = true
			if elemType == "float64" {
				g.buf.WriteString("runtime.SumFloat(")
			} else {
				g.buf.WriteString("runtime.SumInt(")
			}
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "min":
			g.needsRuntime = true
			// Use Ptr versions that return *T for optional coalescing
			if elemType == "float64" {
				g.buf.WriteString("runtime.MinFloatPtr(")
			} else {
				g.buf.WriteString("runtime.MinIntPtr(")
			}
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "max":
			g.needsRuntime = true
			if elemType == "float64" {
				g.buf.WriteString("runtime.MaxFloatPtr(")
			} else {
				g.buf.WriteString("runtime.MaxIntPtr(")
			}
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "first":
			g.needsRuntime = true
			g.buf.WriteString("runtime.FirstPtr(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "last":
			g.needsRuntime = true
			g.buf.WriteString("runtime.LastPtr(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "sorted":
			g.needsRuntime = true
			g.buf.WriteString("runtime.Sort(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "any?":
			// any? without a block returns true if array has any elements
			g.buf.WriteString("(len(")
			g.genExpr(sel.X)
			g.buf.WriteString(") > 0)")
			return true
		case "empty?":
			// empty? returns true if array has no elements
			g.buf.WriteString("(len(")
			g.genExpr(sel.X)
			g.buf.WriteString(") == 0)")
			return true
		case "shift":
			// shift removes and returns the first element, needs pointer
			g.needsRuntime = true
			switch elemType {
			case "int":
				g.buf.WriteString("runtime.ShiftInt(&")
			case "string":
				g.buf.WriteString("runtime.ShiftString(&")
			case "float64":
				g.buf.WriteString("runtime.ShiftFloat(&")
			default:
				g.buf.WriteString("runtime.ShiftAny(&")
			}
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "pop":
			// pop removes and returns the last element, needs pointer
			g.needsRuntime = true
			switch elemType {
			case "int":
				g.buf.WriteString("runtime.PopInt(&")
			case "string":
				g.buf.WriteString("runtime.PopString(&")
			case "float64":
				g.buf.WriteString("runtime.PopFloat(&")
			default:
				g.buf.WriteString("runtime.PopAny(&")
			}
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		}
	}

	// Handle map property methods
	if strings.HasPrefix(receiverType, "map[") || receiverType == "Map" {
		g.needsRuntime = true
		switch sel.Sel {
		case "keys":
			g.buf.WriteString("runtime.Keys(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "values":
			g.buf.WriteString("runtime.Values(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		}
	}

	return false
}

// genStdLibPropertyMethod handles property-style method calls on Int, Float, String types
// like x.even?, x.abs, f.floor, s.upcase, etc.
// Returns true if the method was handled, false otherwise.
func (g *Generator) genStdLibPropertyMethod(sel *ast.SelectorExpr) bool {
	// Don't intercept Go interop calls - they should use native Go methods
	if g.isGoInterop(sel.X) {
		return false
	}

	receiverType := g.inferTypeFromExpr(sel.X)

	// Map common Go types to Rugby types for lookup
	lookupType := receiverType
	switch receiverType {
	case "int", "int64":
		lookupType = "Int"
	case "float64":
		lookupType = "Float"
	case "string":
		lookupType = "String"
	}

	if methods, ok := stdLib[lookupType]; ok {
		if def, ok := methods[sel.Sel]; ok {
			g.needsRuntime = true
			g.buf.WriteString(def.RuntimeFunc)
			g.buf.WriteString("(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		}
	}

	// Also check uniqueMethods for methods with unique names
	if def, ok := uniqueMethods[sel.Sel]; ok {
		g.needsRuntime = true
		g.buf.WriteString(def.RuntimeFunc)
		g.buf.WriteString("(")
		g.genExpr(sel.X)
		g.buf.WriteString(")")
		return true
	}

	return false
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

func (g *Generator) genConcurrentlyExpr(e *ast.ConcurrentlyExpr) {
	g.needsRuntime = true
	g.buf.WriteString("func() any {\n")
	g.indent++

	// Create scope variable
	scopeVar := e.ScopeVar
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

	// Generate body, returning the last expression's value
	if len(e.Body) > 0 {
		for i, stmt := range e.Body {
			if i == len(e.Body)-1 {
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
	g.buf.WriteString("}()")
}
