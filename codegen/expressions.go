package codegen

import (
	"fmt"
	"maps"
	"slices"
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
		// Use %g format but ensure it looks like a float (has decimal or exponent)
		s := fmt.Sprintf("%g", e.Value)
		// Check if result needs decimal point to be recognized as float
		hasDecimalOrExp := false
		for _, c := range s {
			if c == '.' || c == 'e' || c == 'E' {
				hasDecimalOrExp = true
				break
			}
		}
		if !hasDecimalOrExp {
			s += ".0"
		}
		g.buf.WriteString(s)
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
	case *ast.RegexLit:
		g.genRegexLit(e)

	// Literals: composite values
	case *ast.ArrayLit:
		g.genArrayLit(e)
	case *ast.MapLit:
		g.genMapLit(e)
	case *ast.RangeLit:
		g.genRangeLit(e)
	case *ast.TupleLit:
		g.genTupleLit(e)
	case *ast.SetLit:
		g.genSetLit(e)
	case *ast.GoStructLit:
		g.genGoStructLit(e)
	case *ast.StructLit:
		g.genStructLit(e)
	case *ast.SplatExpr:
		// Splat expression: *arr -> arr...
		g.genExpr(e.Expr)
		g.buf.WriteString("...")

	// Identifiers
	case *ast.Ident:
		// Handle 'self' keyword - compiles to receiver variable
		if e.Name == "self" {
			if g.currentClass != "" {
				g.buf.WriteString(receiverName(g.currentClass))
			} else {
				g.addError(fmt.Errorf("line %d: 'self' used outside of class context", e.Line))
				g.buf.WriteString("nil")
			}
		} else if runtimeCall, ok := noParenKernelFuncs[e.Name]; ok {
			// Check for kernel functions that can be used without parens
			g.needsRuntime = true
			g.buf.WriteString(runtimeCall)
		} else if g.currentClass != "" && g.isModuleMethod(e.Name) {
			// Module method call within class - generate as self.Method() (PascalCase for interface)
			recv := receiverName(g.currentClass)
			methodName := snakeToPascalWithAcronyms(e.Name)
			g.buf.WriteString(fmt.Sprintf("%s.%s()", recv, methodName))
		} else if g.currentClass != "" && g.instanceMethods[g.currentClass] != nil && g.instanceMethods[g.currentClass][e.Name] != "" && g.vars[e.Name] == "" {
			// Implicit method call within class - generate as self.method()
			// Only if not a local variable (e.g., parameters shadow instance methods)
			recv := receiverName(g.currentClass)
			methodName := g.instanceMethods[g.currentClass][e.Name]
			g.buf.WriteString(fmt.Sprintf("%s.%s()", recv, methodName))
		} else if g.isNoArgFunction(e.Name) {
			// No-arg function - call it implicitly (Ruby-style)
			goName := snakeToCamelWithAcronyms(e.Name)
			g.buf.WriteString(fmt.Sprintf("%s()", goName))
		} else {
			g.buf.WriteString(e.Name)
		}
	case *ast.InstanceVar:
		if g.currentStruct != nil {
			// In a struct method: @field -> s.Field
			goFieldName := snakeToPascalWithAcronyms(e.Name)
			g.buf.WriteString(fmt.Sprintf("s.%s", goFieldName))
		} else if g.currentClass != "" {
			recv := receiverName(g.currentClass)
			// Use underscore prefix for accessor fields to match struct definition
			// Check both current class and parent classes for accessor fields
			goFieldName := e.Name
			if g.isAccessorField(e.Name) {
				goFieldName = "_" + e.Name
			}
			g.buf.WriteString(fmt.Sprintf("%s.%s", recv, goFieldName))
		} else {
			g.addError(fmt.Errorf("instance variable '@%s' used outside of class context", e.Name))
			g.buf.WriteString("nil")
		}
	case *ast.ClassVar:
		// Class variables are package-level vars named _ClassName_varname
		if g.currentClass != "" {
			key := g.currentClass + "@@" + e.Name
			if varName, ok := g.classVars[key]; ok {
				g.buf.WriteString(varName)
			} else {
				// Fall back to generated name (for forward references)
				g.buf.WriteString(fmt.Sprintf("_%s_%s", g.currentClass, e.Name))
			}
		} else {
			g.addError(fmt.Errorf("class variable '@@%s' used outside of class context", e.Name))
			g.buf.WriteString("nil")
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
	case *ast.ScopeExpr:
		g.genScopeExpr(e)

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
	case *ast.LambdaExpr:
		g.genLambdaExpr(e)
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
		if g.isPublicClass(parentClass) {
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

// genLambdaExpr generates code for arrow lambda expressions: -> (params) { body }
// Lambdas compile to Go anonymous functions (closures).
func (g *Generator) genLambdaExpr(e *ast.LambdaExpr) {
	// Try to get inferred types from type info (for lambdas assigned to typed variables)
	inferredParamTypes, inferredReturnType := g.getLambdaInferredTypes(e)

	g.buf.WriteString("func(")

	// Generate parameters
	for i, param := range e.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.buf.WriteString(param.Name)
		g.buf.WriteString(" ")

		if param.Type != "" {
			// Explicit type annotation takes precedence
			g.buf.WriteString(mapType(param.Type))
		} else if i < len(inferredParamTypes) && inferredParamTypes[i] != "" {
			// Use inferred type from expected function type
			g.buf.WriteString(inferredParamTypes[i])
		} else {
			// Default to any if no type annotation or inference
			g.buf.WriteString("any")
		}
	}

	g.buf.WriteString(")")

	// Generate return type
	if e.ReturnType != "" {
		g.buf.WriteString(" ")
		g.buf.WriteString(mapType(e.ReturnType))
	} else if inferredReturnType != "" {
		// Use inferred return type
		g.buf.WriteString(" ")
		g.buf.WriteString(inferredReturnType)
	} else if len(e.Body) > 0 {
		// If no return type but body has statements, use 'any' as return type
		// to allow implicit returns
		g.buf.WriteString(" any")
	}

	g.buf.WriteString(" {\n")
	g.indent++

	// Save variable scope
	savedVars := maps.Clone(g.vars)

	// Add parameters to scope with inferred types
	for i, param := range e.Params {
		var goType string
		if param.Type != "" {
			goType = mapType(param.Type)
		} else if i < len(inferredParamTypes) && inferredParamTypes[i] != "" {
			goType = inferredParamTypes[i]
		} else {
			goType = "any"
		}
		g.vars[param.Name] = goType
	}

	// Generate body
	if len(e.Body) > 0 {
		// Check if last statement is an expression that should be returned
		lastIdx := len(e.Body) - 1
		for i, stmt := range e.Body {
			if i == lastIdx {
				// For the last statement, if it's an expression statement, add return
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					g.writeIndent()
					g.buf.WriteString("return ")
					g.genExpr(exprStmt.Expr)
					g.buf.WriteString("\n")
				} else if compound, ok := stmt.(*ast.CompoundAssignStmt); ok {
					// For compound assignments (x += y), generate assignment and return the new value
					// This makes lambdas like `-> { count += 1 }` return the new count value
					g.genStatement(stmt)
					g.writeIndent()
					g.buf.WriteString("return ")
					g.buf.WriteString(compound.Name)
					g.buf.WriteString("\n")
				} else {
					g.genStatement(stmt)
				}
			} else {
				g.genStatement(stmt)
			}
		}
	}

	// Restore variable scope
	g.vars = savedVars

	g.indent--
	g.writeIndent()
	g.buf.WriteString("}")
}

// getLambdaInferredTypes extracts parameter and return types from the lambda's
// inferred function type. Returns nil slices if no type info is available.
func (g *Generator) getLambdaInferredTypes(e *ast.LambdaExpr) (paramTypes []string, returnType string) {
	goType := g.typeInfo.GetGoType(e)
	if goType == "" || !strings.HasPrefix(goType, "func(") {
		return nil, ""
	}

	// Parse: func(type1, type2) returnType
	// Find the closing paren for params
	depth := 0
	paramsEnd := -1
	for i := 5; i < len(goType); i++ { // Start after "func("
		switch goType[i] {
		case '(':
			depth++
		case ')':
			if depth == 0 {
				paramsEnd = i
				break
			}
			depth--
		}
		if paramsEnd != -1 {
			break
		}
	}

	if paramsEnd == -1 {
		return nil, ""
	}

	// Extract params
	paramsStr := goType[5:paramsEnd]
	if paramsStr != "" {
		// Split on comma, handling nested types
		paramTypes = splitGoTypeParams(paramsStr)
	}

	// Extract return type (everything after ") ")
	rest := strings.TrimSpace(goType[paramsEnd+1:])
	if rest != "" {
		returnType = rest
	}

	return paramTypes, returnType
}

// splitGoTypeParams splits a comma-separated list of Go types, respecting nested brackets.
func splitGoTypeParams(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0

	for _, c := range s {
		switch c {
		case '(', '[', '<':
			depth++
			current.WriteRune(c)
		case ')', ']', '>':
			depth--
			current.WriteRune(c)
		case ',':
			if depth == 0 {
				result = append(result, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(c)
			}
		default:
			current.WriteRune(c)
		}
	}

	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
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
	// Handle regex match operators
	if e.Op == "=~" {
		g.genRegexMatch(e.Left, e.Right, false)
		return
	}
	if e.Op == "!~" {
		g.genRegexMatch(e.Left, e.Right, true)
		return
	}

	// Handle set operators: | (union), & (intersection), - (difference)
	leftType := g.inferTypeFromExpr(e.Left)
	if leftType == "Set" || strings.HasPrefix(leftType, "Set<") {
		switch e.Op {
		case "|":
			g.needsRuntime = true
			g.buf.WriteString("runtime.SetUnion(")
			g.genExpr(e.Left)
			g.buf.WriteString(", ")
			g.genExpr(e.Right)
			g.buf.WriteString(")")
			return
		case "&":
			g.needsRuntime = true
			g.buf.WriteString("runtime.SetIntersection(")
			g.genExpr(e.Left)
			g.buf.WriteString(", ")
			g.genExpr(e.Right)
			g.buf.WriteString(")")
			return
		case "-":
			g.needsRuntime = true
			g.buf.WriteString("runtime.SetDifference(")
			g.genExpr(e.Left)
			g.buf.WriteString(", ")
			g.genExpr(e.Right)
			g.buf.WriteString(")")
			return
		}
	}

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
		// Check if left operand is a channel - use native Go channel send
		if g.typeInfo.GetTypeKind(e.Left) == TypeChannel {
			g.genExpr(e.Left)
			g.buf.WriteString(" <- ")
			g.genExpr(e.Right)
			return
		}
		// For arrays, use runtime.ShiftLeft
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

	// Translate Rugby operators to Go operators
	op := e.Op
	switch op {
	case "and":
		op = "&&"
	case "or":
		op = "||"
	}

	g.buf.WriteString("(")
	g.genExpr(e.Left)
	g.buf.WriteString(" ")
	g.buf.WriteString(op)
	g.buf.WriteString(" ")
	g.genExpr(e.Right)
	g.buf.WriteString(")")
}

// hasAnyTypedOperand returns true if BOTH operands have type any (interface{})
// We require both because if only one is any, it's likely due to generic typing
// (like reduce's accumulator) where codegen infers concrete types.
func (g *Generator) hasAnyTypedOperand(left, right ast.Expression) bool {
	leftKind := g.typeInfo.GetTypeKind(left)
	rightKind := g.typeInfo.GetTypeKind(right)
	return leftKind == TypeAny && rightKind == TypeAny
}

// canUseDirectComparison returns true if we can use Go's == operator directly
// instead of runtime.Equal. This is possible when:
// 1. Both operands are primitive literals, OR
// 2. Both operands are primitive types (Int, Float, String, Bool)
func (g *Generator) canUseDirectComparison(left, right ast.Expression) bool {
	// Primitive literals can always use direct comparison
	if isPrimitiveLiteral(left) && isPrimitiveLiteral(right) {
		return true
	}

	// Check if both types are primitive
	leftKind := g.typeInfo.GetTypeKind(left)
	rightKind := g.typeInfo.GetTypeKind(right)
	return isPrimitiveKind(leftKind) && isPrimitiveKind(rightKind)
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

// genScopeExpr generates code for scope resolution expressions (Module::Class or Enum::Value)
// For enums: Status::Active -> StatusActive
// For modules: Http::Response -> Http_Response
func (g *Generator) genScopeExpr(e *ast.ScopeExpr) {
	// Build the full namespaced name
	var prefix string
	switch left := e.Left.(type) {
	case *ast.Ident:
		prefix = left.Name
	case *ast.ScopeExpr:
		// Recursively handle nested scope (A::B::C)
		// For now, just flatten it
		g.genScopeExpr(left)
		g.buf.WriteString("_")
		g.buf.WriteString(e.Right)
		return
	default:
		// Unexpected left expression type - just generate both parts
		g.genExpr(e.Left)
		g.buf.WriteString("_")
		g.buf.WriteString(e.Right)
		return
	}

	// Check if this is an enum reference
	if g.enums != nil {
		if _, isEnum := g.enums[prefix]; isEnum {
			// For enums, generate: EnumName + ValueName (no separator)
			// Status::Active -> StatusActive
			g.buf.WriteString(prefix)
			g.buf.WriteString(e.Right)
			return
		}
	}

	// Generate the namespaced name for modules: Module_Class
	g.buf.WriteString(prefix)
	g.buf.WriteString("_")
	g.buf.WriteString(e.Right)
}

// scopeExprToClassName converts a scope expression to a namespaced class name string.
// For example, Http::Response becomes "Http_Response"
func (g *Generator) scopeExprToClassName(e *ast.ScopeExpr) string {
	var parts []string
	current := e
	for {
		parts = append([]string{current.Right}, parts...)
		switch left := current.Left.(type) {
		case *ast.Ident:
			parts = append([]string{left.Name}, parts...)
			return strings.Join(parts, "_")
		case *ast.ScopeExpr:
			current = left
		default:
			// Unexpected - just return what we have
			return strings.Join(parts, "_")
		}
	}
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

// genRegexMatch generates code for =~ and !~ operators.
// For `string =~ regex`, generates `regex.MatchString(string)`.
// For `string !~ regex`, generates `!regex.MatchString(string)`.
func (g *Generator) genRegexMatch(left, right ast.Expression, negate bool) {
	if negate {
		g.buf.WriteString("!")
	}
	g.buf.WriteString("(")
	g.genExpr(right)
	g.buf.WriteString(").MatchString(")
	g.genExpr(left)
	g.buf.WriteString(")")
}

// genRegexLit generates a compiled regex: regexp.MustCompile("pattern")
// Flags are prepended to the pattern as (?flags).
func (g *Generator) genRegexLit(r *ast.RegexLit) {
	g.needsRegexp = true

	pattern := r.Pattern
	if r.Flags != "" {
		// Prepend flags as (?flags) to the pattern
		// Go supports: i (case insensitive), m (multiline), s (. matches \n)
		// Map Rugby flags: i -> i, m -> m, s -> s, x is not supported in Go
		goFlags := ""
		for _, f := range r.Flags {
			switch f {
			case 'i', 'm', 's':
				goFlags += string(f)
				// 'x' (extended/verbose) is not supported in Go regexp, skip it
			}
		}
		if goFlags != "" {
			pattern = "(?" + goFlags + ")" + pattern
		}
	}

	g.buf.WriteString(fmt.Sprintf("regexp.MustCompile(%q)", pattern))
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
	if goType := g.typeInfo.GetGoType(arr); goType != "" {
		return goType
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

// genTupleLit generates comma-separated values for tuple literals.
// Tuples in Rugby are used for multi-value returns and compile to Go's
// multiple return value syntax when used in a return context.
func (g *Generator) genTupleLit(t *ast.TupleLit) {
	for i, elem := range t.Elements {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.genExpr(elem)
	}
}

// genSetLit generates Go code for a set literal: Set{1, 2, 3} -> map[int]struct{}{1: {}, 2: {}, 3: {}}
func (g *Generator) genSetLit(s *ast.SetLit) {
	// Determine element type from type hint or first element
	elemType := "any"
	if s.TypeHint != "" {
		elemType = mapType(s.TypeHint)
	} else if len(s.Elements) > 0 {
		// Infer type from first element
		if inferred := g.inferTypeFromExpr(s.Elements[0]); inferred != "" {
			elemType = mapType(inferred)
		}
	}

	g.buf.WriteString("map[")
	g.buf.WriteString(elemType)
	g.buf.WriteString("]struct{}{")

	for i, elem := range s.Elements {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.genExpr(elem)
		g.buf.WriteString(": {}")
	}

	g.buf.WriteString("}")
}

// genGoStructLit generates Go code for a Go struct literal: sync.WaitGroup{} -> &sync.WaitGroup{}
func (g *Generator) genGoStructLit(s *ast.GoStructLit) {
	// Mark this import as used
	g.imports[s.Package] = true
	// Generate pointer to struct for consistent behavior with Go types
	g.buf.WriteString("&")
	g.buf.WriteString(s.Package)
	g.buf.WriteString(".")
	g.buf.WriteString(s.Type)
	g.buf.WriteString("{}")
}

// genStructLit generates Go code for a Rugby struct literal: Point{x: 10, y: 20} -> Point{X: 10, Y: 20}
func (g *Generator) genStructLit(s *ast.StructLit) {
	g.buf.WriteString(s.Name)
	g.buf.WriteString("{")
	for i, field := range s.Fields {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		// Convert snake_case field name to PascalCase
		goFieldName := snakeToPascalWithAcronyms(field.Name)
		g.buf.WriteString(goFieldName)
		g.buf.WriteString(": ")
		g.genExpr(field.Value)
	}
	g.buf.WriteString("}")
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

	// For maps, always use native Go map indexing
	if strings.HasPrefix(leftType, "map[") || g.typeInfo.GetTypeKind(idx.Left) == TypeMap {
		g.genExpr(idx.Left)
		g.buf.WriteString("[")
		g.genExpr(idx.Index)
		g.buf.WriteString("]")
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

	// Determine the collection type to add proper type assertion
	collectionType := g.inferTypeFromExpr(collection)
	goType := g.rugbyToGoSliceType(collectionType)

	g.buf.WriteString("runtime.Slice(")
	g.genExpr(collection)
	g.buf.WriteString(", runtime.Range{Start: ")
	g.genExpr(r.Start)
	g.buf.WriteString(", End: ")
	if r.End != nil {
		g.genExpr(r.End)
	} else {
		// Open-ended range (arr[start..]) - End is ignored when OpenEnd is true
		g.buf.WriteString("0")
	}
	g.buf.WriteString(", Exclusive: ")
	if r.Exclusive {
		g.buf.WriteString("true")
	} else {
		g.buf.WriteString("false")
	}
	if r.End == nil {
		g.buf.WriteString(", OpenEnd: true")
	}
	g.buf.WriteString("})")

	// Add type assertion if we know the type
	if goType != "" {
		g.buf.WriteString(".(")
		g.buf.WriteString(goType)
		g.buf.WriteString(")")
	}
}

// rugbyToGoSliceType converts a Rugby type to the Go slice type that would result from slicing.
// Returns empty string if type is unknown or not sliceable.
func (g *Generator) rugbyToGoSliceType(rugbyType string) string {
	switch rugbyType {
	case "String":
		return "string"
	case "Array<Int>", "Array<int>":
		return "[]int"
	case "Array<String>", "Array<string>":
		return "[]string"
	case "Array<Float>", "Array<float>", "Array<Float64>":
		return "[]float64"
	case "Array<Bool>", "Array<bool>":
		return "[]bool"
	case "Array":
		return "[]any"
	default:
		// Handle generic Array<T> patterns
		if strings.HasPrefix(rugbyType, "Array<") && strings.HasSuffix(rugbyType, ">") {
			inner := rugbyType[6 : len(rugbyType)-1]
			return "[]" + mapType(inner)
		}
		return ""
	}
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
	if goType := g.typeInfo.GetGoType(m); goType != "" {
		return goType
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
				// Handle lambda argument for iteration methods
				if lambda, ok := call.Args[0].(*ast.LambdaExpr); ok {
					g.genLambdaIterationCall(sel.X, sel.Sel, lambda, call.Args[1:])
					return
				}
				// Handle reduce with initial value + lambda
				if bm.hasAccumulator && len(call.Args) >= 2 {
					if lambda, ok := call.Args[1].(*ast.LambdaExpr); ok {
						g.genLambdaIterationCall(sel.X, sel.Sel, lambda, call.Args[:1])
						return
					}
				}
			}
		}
		// Handle each/each_with_index with lambda argument
		if sel.Sel == "each" || sel.Sel == "each_with_index" {
			if len(call.Args) >= 1 {
				if lambda, ok := call.Args[0].(*ast.LambdaExpr); ok {
					g.genLambdaIterationCall(sel.X, sel.Sel, lambda, call.Args[1:])
					return
				}
			}
		}
		// Handle times with lambda argument (n.times -> (i) { ... })
		if sel.Sel == "times" {
			if len(call.Args) >= 1 {
				if lambda, ok := call.Args[0].(*ast.LambdaExpr); ok {
					g.genLambdaIterationCall(sel.X, sel.Sel, lambda, call.Args[1:])
					return
				}
			}
		}
		// Handle upto/downto with lambda argument (n.upto(m) -> (i) { ... })
		if sel.Sel == "upto" || sel.Sel == "downto" {
			if len(call.Args) >= 2 {
				if lambda, ok := call.Args[1].(*ast.LambdaExpr); ok {
					g.genLambdaIterationCall(sel.X, sel.Sel, lambda, call.Args[:1])
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
			// Handle enum type methods: Color.from_string("Red"), Color.values
			if g.enums != nil {
				if _, isEnum := g.enums[ident.Name]; isEnum {
					switch sel.Sel {
					case "from_string":
						// Color.from_string("Red") -> ColorFromString("Red")
						g.buf.WriteString(ident.Name)
						g.buf.WriteString("FromString(")
						for i, arg := range call.Args {
							if i > 0 {
								g.buf.WriteString(", ")
							}
							g.genExpr(arg)
						}
						g.buf.WriteString(")")
						return
					case "values":
						// Color.values -> ColorValues()
						g.buf.WriteString(ident.Name)
						g.buf.WriteString("Values()")
						return
					}
				}
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

		// Check for module method call within class context
		// (e.g., touch(now) in a class that includes a module with touch method)
		if g.currentClass != "" && g.isModuleMethod(funcName) {
			recv := receiverName(g.currentClass)
			methodName := snakeToPascalWithAcronyms(funcName)
			g.buf.WriteString(fmt.Sprintf("%s.%s(", recv, methodName))
			for i, arg := range call.Args {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				g.genExpr(arg)
			}
			g.buf.WriteString(")")
			return
		}

		// Check for private method call within class context
		// (e.g., encrypt_password() in a class that has private def encrypt_password)
		if g.currentClass != "" && g.privateMethods[g.currentClass] != nil && g.privateMethods[g.currentClass][funcName] {
			recv := receiverName(g.currentClass)
			methodName := "_" + snakeToCamelWithAcronyms(funcName)
			g.buf.WriteString(fmt.Sprintf("%s.%s(", recv, methodName))
			for i, arg := range call.Args {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				g.genExpr(arg)
			}
			g.buf.WriteString(")")
			return
		}

		// Check for keyword arguments: func(name: value, ...)
		if fnDecl, ok := g.functions[funcName]; ok && g.hasKeywordArgs(call.Args) {
			if g.pubFuncs[funcName] {
				funcName = snakeToPascalWithAcronyms(funcName)
			} else {
				funcName = snakeToCamelWithAcronyms(funcName)
			}
			g.buf.WriteString(funcName)
			g.genKeywordCallArgs(fnDecl, call.Args)
			return
		}

		// Check for user-defined function with default parameters
		if fnDecl, ok := g.functions[funcName]; ok && len(call.Args) < len(fnDecl.Params) {
			// We need to fill in default values for missing arguments
			if g.pubFuncs[funcName] {
				funcName = snakeToPascalWithAcronyms(funcName)
			} else {
				funcName = snakeToCamelWithAcronyms(funcName)
			}
			g.buf.WriteString(funcName)
			g.buf.WriteString("(")
			for i, param := range fnDecl.Params {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				if i < len(call.Args) {
					// Use provided argument
					g.genExpr(call.Args[i])
				} else if param.DefaultValue != nil {
					// Use default value
					g.genExpr(param.DefaultValue)
				} else {
					// No default - this would be an error, but semantic analysis should catch it
					g.buf.WriteString("nil")
				}
			}
			g.buf.WriteString(")")
			return
		}

		if kf, ok := kernelFuncs[funcName]; ok {
			g.needsRuntime = true
			if kf.transform != nil {
				funcName = kf.transform(call.Args)
			} else {
				funcName = kf.runtimeFunc
			}
		} else if g.pubFuncs[funcName] {
			// Function declared with 'pub' -> PascalCase
			funcName = snakeToPascalWithAcronyms(funcName)
		} else {
			// Regular function -> camelCase
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
				// Resolve module-scoped class name (e.g., Response -> Http_Response when inside module Http)
				className := g.resolveClassName(ident.Name)
				var ctorName string
				if g.isPublicClass(ident.Name) {
					ctorName = fmt.Sprintf("New%s", className)
				} else {
					ctorName = fmt.Sprintf("new%s", className)
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
			// Handle scoped class constructors like Http::Response.new(...)
			// Generate: newHttp_Response(...)
			if scope, ok := fn.X.(*ast.ScopeExpr); ok {
				className := g.scopeExprToClassName(scope)
				var ctorName string
				// For scoped classes, check if the module is public
				if g.isPublicClass(className) {
					ctorName = "New" + className
				} else {
					ctorName = "new" + className
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
		if ident, ok := fn.X.(*ast.Ident); ok && g.isClass(ident.Name) {
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

		// Handle module method calls (ModuleName.method_name -> ModuleName_Method)
		if ident, ok := fn.X.(*ast.Ident); ok {
			if mod, hasModule := g.modules[ident.Name]; hasModule {
				// Check if the module has this class method
				for _, method := range mod.Methods {
					if method.IsClassMethod && method.Name == fn.Sel {
						// Generate: ModuleName_MethodName(args)
						funcName := ident.Name + "_" + snakeToPascalWithAcronyms(fn.Sel)
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
		if strings.HasPrefix(recvTypeForArray, "[]") || recvTypeForArray == "Array" || strings.HasPrefix(recvTypeForArray, "Array<") {
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

		// Handle set method calls: s.include?(x), s.add(x), s.delete(x), s.size, s.to_a, s.clear
		recvTypeForSet := g.inferTypeFromExpr(fn.X)
		if recvTypeForSet == "Set" || strings.HasPrefix(recvTypeForSet, "Set<") {
			switch fn.Sel {
			case "include?":
				if len(call.Args) != 1 {
					g.addError(fmt.Errorf("include? requires exactly 1 argument"))
					g.buf.WriteString("false")
					return
				}
				g.needsRuntime = true
				g.buf.WriteString("runtime.SetContains(")
				g.genExpr(fn.X)
				g.buf.WriteString(", ")
				g.genExpr(call.Args[0])
				g.buf.WriteString(")")
				return
			case "add":
				if len(call.Args) != 1 {
					g.addError(fmt.Errorf("add requires exactly 1 argument"))
					return
				}
				g.needsRuntime = true
				g.buf.WriteString("runtime.SetAdd(")
				g.genExpr(fn.X)
				g.buf.WriteString(", ")
				g.genExpr(call.Args[0])
				g.buf.WriteString(")")
				return
			case "delete":
				if len(call.Args) != 1 {
					g.addError(fmt.Errorf("delete requires exactly 1 argument"))
					return
				}
				g.needsRuntime = true
				g.buf.WriteString("runtime.SetDelete(")
				g.genExpr(fn.X)
				g.buf.WriteString(", ")
				g.genExpr(call.Args[0])
				g.buf.WriteString(")")
				return
			case "size", "length", "count":
				g.needsRuntime = true
				g.buf.WriteString("runtime.SetSize(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "to_a":
				g.needsRuntime = true
				g.buf.WriteString("runtime.SetToArray(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "clear":
				g.needsRuntime = true
				g.buf.WriteString("runtime.SetClear(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "empty?":
				g.buf.WriteString("(len(")
				g.genExpr(fn.X)
				g.buf.WriteString(") == 0)")
				return
			case "any?":
				g.buf.WriteString("(len(")
				g.genExpr(fn.X)
				g.buf.WriteString(") > 0)")
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
		// Also normalize array/map types
		if strings.HasPrefix(recvType, "[]") || strings.HasPrefix(recvType, "Array") {
			lookupType = "Array"
		}
		if strings.HasPrefix(recvType, "map[") || strings.HasPrefix(recvType, "Map") {
			lookupType = "Map"
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
		// Also skip for user-defined class receivers (class methods take precedence)
		if !found && !g.isGoInterop(fn.X) {
			receiverClassName := g.getReceiverClassName(fn.X)
			baseClassName := stripTypeParams(receiverClassName)
			isClassReceiver := baseClassName != "" && (g.isClass(baseClassName) || g.isModuleScopedClass(baseClassName))

			if !isClassReceiver {
				if def, ok := uniqueMethods[fn.Sel]; ok {
					methodDef = def
					found = true
				}
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
				g.addError(fmt.Errorf("is_a? requires exactly one type argument, got %d", len(call.Args)))
				g.buf.WriteString("false")
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
				g.addError(fmt.Errorf("as requires exactly one type argument, got %d", len(call.Args)))
				g.buf.WriteString("func() (any, bool) { return nil, false }()")
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
			case "unwrap_or":
				if len(call.Args) == 0 {
					g.buf.WriteString("/* unwrap_or requires default argument */")
					return
				}
				g.needsRuntime = true
				// Check if the receiver is a safe navigation expression (returns any)
				_, isSafeNav := fn.X.(*ast.SafeNavExpr)
				// Generate coalescing code based on the optional type
				switch receiverType {
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
					// Generic optional type - use IIFE
					varName := fmt.Sprintf("_uo%d", g.tempVarCounter)
					g.tempVarCounter++
					g.buf.WriteString("func() ")
					defaultType := g.inferTypeFromExpr(call.Args[0])
					if defaultType != "" {
						g.buf.WriteString(mapType(defaultType))
					} else {
						g.buf.WriteString("any")
					}
					g.buf.WriteString(" { ")
					g.buf.WriteString(varName)
					g.buf.WriteString(" := ")
					g.genExpr(fn.X)
					g.buf.WriteString("; if ")
					g.buf.WriteString(varName)
					g.buf.WriteString(" != nil { return *")
					g.buf.WriteString(varName)
					g.buf.WriteString(" }; return ")
					g.genExpr(call.Args[0])
					g.buf.WriteString(" }()")
					return
				}
				g.genExpr(fn.X)
				g.buf.WriteString(", ")
				g.genExpr(call.Args[0])
				g.buf.WriteString(")")
				return
			}
		}

		// Handle enum instance method calls (Color::Red.to_s, enum_var.value)
		if g.enums != nil {
			if _, isEnum := g.enums[receiverType]; isEnum {
				switch fn.Sel {
				case "to_s":
					// enum.to_s -> enum.String()
					g.genExpr(fn.X)
					g.buf.WriteString(".String()")
					return
				case "value":
					// enum.value -> int(enum)
					g.buf.WriteString("int(")
					g.genExpr(fn.X)
					g.buf.WriteString(")")
					return
				}
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

// hasKeywordArgs checks if any of the arguments are keyword arguments.
func (g *Generator) hasKeywordArgs(args []ast.Expression) bool {
	for _, arg := range args {
		if _, ok := arg.(*ast.KeywordArg); ok {
			return true
		}
	}
	return false
}

// genKeywordCallArgs generates function call arguments with keyword args reordered.
// It matches keyword arguments to parameter names and fills in defaults for missing args.
func (g *Generator) genKeywordCallArgs(fnDecl *ast.FuncDecl, args []ast.Expression) {
	// Separate positional and keyword arguments
	var positionalArgs []ast.Expression
	keywordArgs := make(map[string]ast.Expression)
	for _, arg := range args {
		if kwArg, ok := arg.(*ast.KeywordArg); ok {
			keywordArgs[kwArg.Name] = kwArg.Value
		} else {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	// Generate arguments in parameter order
	g.buf.WriteString("(")
	for i, param := range fnDecl.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}

		// Check if we have a positional arg for this position
		if i < len(positionalArgs) {
			g.genExpr(positionalArgs[i])
			continue
		}

		// Check if we have a keyword arg for this parameter
		if kwValue, ok := keywordArgs[param.Name]; ok {
			g.genExpr(kwValue)
			continue
		}

		// Use default value if available
		if param.DefaultValue != nil {
			g.genExpr(param.DefaultValue)
			continue
		}

		// No value provided - this should be caught by semantic analysis
		g.buf.WriteString("nil")
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
	// Use semantic analysis for Go type
	goType := g.typeInfo.GetGoType(iterable)
	if goType == "" {
		goType = g.inferTypeFromExpr(iterable)
	}

	// Check for set iteration (single parameter iterating over keys)
	// Sets are map[T]struct{}, so we iterate with `for k := range` not `for _, v := range`
	rugbyType := g.inferTypeFromExpr(iterable)
	isSetIteration := rugbyType == "Set" || strings.HasPrefix(rugbyType, "Set<")

	if isSetIteration {
		varName := "_"
		if len(block.Params) > 0 {
			varName = block.Params[0]
		}

		// For sets, iterate over keys: for k := range set
		g.buf.WriteString("for ")
		g.buf.WriteString(varName)
		g.buf.WriteString(" := range ")
		g.genExpr(iterable)
		g.buf.WriteString(" {\n")

		// Infer element type from the set's Go type (map[T]struct{} -> T)
		elemType := "any"
		if strings.HasPrefix(goType, "map[") {
			// Extract key type from map[T]struct{}
			keyEnd := strings.Index(goType, "]")
			if keyEnd > 4 {
				elemType = goType[4:keyEnd]
			}
		}

		g.scopedVar(varName, elemType, func() {
			g.indent++
			for _, stmt := range block.Body {
				g.genStatement(stmt)
			}
			g.indent--
		})

		g.writeIndent()
		g.buf.WriteString("}")
		return
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

// genLambdaIterationCall generates code for iteration methods with lambda arguments.
// For example: arr.each -> (x) { puts x } becomes a for loop calling the lambda.
func (g *Generator) genLambdaIterationCall(iterable ast.Expression, method string, lambda *ast.LambdaExpr, extraArgs []ast.Expression) {
	switch method {
	case "each":
		g.genLambdaEach(iterable, lambda)
	case "each_with_index":
		g.genLambdaEachWithIndex(iterable, lambda)
	case "map":
		g.genLambdaMap(iterable, lambda)
	case "select", "filter":
		g.genLambdaSelect(iterable, lambda)
	case "reject":
		g.genLambdaReject(iterable, lambda)
	case "find", "detect":
		g.genLambdaFind(iterable, lambda)
	case "any?":
		g.genLambdaAny(iterable, lambda)
	case "all?":
		g.genLambdaAll(iterable, lambda)
	case "none?":
		g.genLambdaNone(iterable, lambda)
	case "reduce":
		if len(extraArgs) > 0 {
			g.genLambdaReduce(iterable, lambda, extraArgs[0])
		} else {
			g.buf.WriteString("/* reduce requires initial value */")
		}
	case "times":
		g.genLambdaTimes(iterable, lambda)
	case "upto":
		if len(extraArgs) > 0 {
			g.genLambdaUpto(iterable, lambda, extraArgs[0])
		} else {
			g.buf.WriteString("/* upto requires target value */")
		}
	case "downto":
		if len(extraArgs) > 0 {
			g.genLambdaDownto(iterable, lambda, extraArgs[0])
		} else {
			g.buf.WriteString("/* downto requires target value */")
		}
	default:
		g.buf.WriteString(fmt.Sprintf("/* unsupported lambda iteration: %s */", method))
	}
}

// genLambdaEach generates a for loop that calls the lambda for each element.
func (g *Generator) genLambdaEach(iterable ast.Expression, lambda *ast.LambdaExpr) {
	// Check if iterating over a map with 2 parameters (key, value)
	if len(lambda.Params) == 2 {
		// Check if the iterable type is a map
		if goType := g.typeInfo.GetGoType(iterable); strings.HasPrefix(goType, "map[") {
			keyName := lambda.Params[0].Name
			valueName := lambda.Params[1].Name
			// Extract key and value types from map[K]V
			mapType := goType
			keyType := "any"
			valueType := "any"
			if strings.HasPrefix(mapType, "map[") {
				rest := mapType[4:]
				depth := 0
				for i, ch := range rest {
					if ch == '[' {
						depth++
					} else if ch == ']' {
						if depth == 0 {
							keyType = rest[:i]
							valueType = rest[i+1:]
							break
						}
						depth--
					}
				}
			}

			g.buf.WriteString("for ")
			g.buf.WriteString(keyName)
			g.buf.WriteString(", ")
			g.buf.WriteString(valueName)
			g.buf.WriteString(" := range ")
			g.genExpr(iterable)
			g.buf.WriteString(" {\n")

			// Use scopedVars to register both key and value variables
			// Set inInlinedLambda so that bare return becomes continue
			prevInInlinedLambda := g.inInlinedLambda
			g.inInlinedLambda = true
			g.scopedVars([]struct{ name, varType string }{{keyName, keyType}, {valueName, valueType}}, func() {
				g.indent++
				for _, stmt := range lambda.Body {
					g.genStatement(stmt)
				}
				g.indent--
			})
			g.inInlinedLambda = prevInInlinedLambda

			g.writeIndent()
			g.buf.WriteString("}")
			return
		}
	}

	// Generate: for _, v := range iterable { lambda(v) }
	varName := "_v"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}

	elemType := g.inferArrayElementGoType(iterable)

	g.buf.WriteString("for _, ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := range ")
	g.genExpr(iterable)
	g.buf.WriteString(" {\n")

	// Use scopedVar to register the loop variable with its type
	// Set inInlinedLambda so that bare return becomes continue
	prevInInlinedLambda := g.inInlinedLambda
	g.inInlinedLambda = true
	g.scopedVar(varName, elemType, func() {
		g.indent++
		for _, stmt := range lambda.Body {
			g.genStatement(stmt)
		}
		g.indent--
	})
	g.inInlinedLambda = prevInInlinedLambda

	g.writeIndent()
	g.buf.WriteString("}")
}

// genLambdaEachWithIndex generates a for loop with index.
func (g *Generator) genLambdaEachWithIndex(iterable ast.Expression, lambda *ast.LambdaExpr) {
	varName := "_v"
	idxName := "_i"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}
	if len(lambda.Params) > 1 {
		idxName = lambda.Params[1].Name
	}

	elemType := g.inferArrayElementGoType(iterable)

	g.buf.WriteString("for ")
	g.buf.WriteString(idxName)
	g.buf.WriteString(", ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := range ")
	g.genExpr(iterable)
	g.buf.WriteString(" {\n")

	// Use scopedVars to register both index and element variables
	// Set inInlinedLambda so that bare return becomes continue
	prevInInlinedLambda := g.inInlinedLambda
	g.inInlinedLambda = true
	g.scopedVars([]struct{ name, varType string }{{idxName, "int"}, {varName, elemType}}, func() {
		g.indent++
		for _, stmt := range lambda.Body {
			g.genStatement(stmt)
		}
		g.indent--
	})
	g.inInlinedLambda = prevInInlinedLambda

	g.writeIndent()
	g.buf.WriteString("}")
}

// genLambdaMap generates code for arr.map -> (x) { expr }
func (g *Generator) genLambdaMap(iterable ast.Expression, lambda *ast.LambdaExpr) {
	g.needsRuntime = true
	varName := "_v"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}

	// Infer element type for proper typing
	elemType := g.inferArrayElementGoType(iterable)

	g.buf.WriteString("runtime.Map(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" ")
	g.buf.WriteString(elemType)
	g.buf.WriteString(") any {\n")

	g.indent++
	// Generate body, last expression is return value
	for i, stmt := range lambda.Body {
		if i == len(lambda.Body)-1 {
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
	if len(lambda.Body) == 0 {
		g.writeIndent()
		g.buf.WriteString("return nil\n")
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaSelect generates code for arr.select -> (x) { condition }
func (g *Generator) genLambdaSelect(iterable ast.Expression, lambda *ast.LambdaExpr) {
	g.needsRuntime = true
	varName := "_v"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}

	elemType := g.inferArrayElementGoType(iterable)

	g.buf.WriteString("runtime.Select(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" ")
	g.buf.WriteString(elemType)
	g.buf.WriteString(") bool {\n")

	g.indent++
	for i, stmt := range lambda.Body {
		if i == len(lambda.Body)-1 {
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
	if len(lambda.Body) == 0 {
		g.writeIndent()
		g.buf.WriteString("return false\n")
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaReject generates code for arr.reject -> (x) { condition }
func (g *Generator) genLambdaReject(iterable ast.Expression, lambda *ast.LambdaExpr) {
	g.needsRuntime = true
	varName := "_v"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}

	elemType := g.inferArrayElementGoType(iterable)

	g.buf.WriteString("runtime.Reject(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" ")
	g.buf.WriteString(elemType)
	g.buf.WriteString(") bool {\n")

	g.indent++
	for i, stmt := range lambda.Body {
		if i == len(lambda.Body)-1 {
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
	if len(lambda.Body) == 0 {
		g.writeIndent()
		g.buf.WriteString("return false\n")
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaFind generates code for arr.find -> (x) { condition }
func (g *Generator) genLambdaFind(iterable ast.Expression, lambda *ast.LambdaExpr) {
	g.needsRuntime = true
	varName := "_v"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}

	elemType := g.inferArrayElementGoType(iterable)

	g.buf.WriteString("runtime.FindPtr(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" ")
	g.buf.WriteString(elemType)
	g.buf.WriteString(") bool {\n")

	g.indent++
	for i, stmt := range lambda.Body {
		if i == len(lambda.Body)-1 {
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
	if len(lambda.Body) == 0 {
		g.writeIndent()
		g.buf.WriteString("return false\n")
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaAny generates code for arr.any? -> (x) { condition }
func (g *Generator) genLambdaAny(iterable ast.Expression, lambda *ast.LambdaExpr) {
	g.needsRuntime = true
	varName := "_v"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}

	elemType := g.inferArrayElementGoType(iterable)

	g.buf.WriteString("runtime.Any(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" ")
	g.buf.WriteString(elemType)
	g.buf.WriteString(") bool {\n")

	g.indent++
	for i, stmt := range lambda.Body {
		if i == len(lambda.Body)-1 {
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
	if len(lambda.Body) == 0 {
		g.writeIndent()
		g.buf.WriteString("return false\n")
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaAll generates code for arr.all? -> (x) { condition }
func (g *Generator) genLambdaAll(iterable ast.Expression, lambda *ast.LambdaExpr) {
	g.needsRuntime = true
	varName := "_v"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}

	elemType := g.inferArrayElementGoType(iterable)

	g.buf.WriteString("runtime.All(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" ")
	g.buf.WriteString(elemType)
	g.buf.WriteString(") bool {\n")

	g.indent++
	for i, stmt := range lambda.Body {
		if i == len(lambda.Body)-1 {
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
	if len(lambda.Body) == 0 {
		g.writeIndent()
		g.buf.WriteString("return true\n")
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaNone generates code for arr.none? -> (x) { condition }
func (g *Generator) genLambdaNone(iterable ast.Expression, lambda *ast.LambdaExpr) {
	g.needsRuntime = true
	varName := "_v"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}

	elemType := g.inferArrayElementGoType(iterable)

	g.buf.WriteString("runtime.None(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" ")
	g.buf.WriteString(elemType)
	g.buf.WriteString(") bool {\n")

	g.indent++
	for i, stmt := range lambda.Body {
		if i == len(lambda.Body)-1 {
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
	if len(lambda.Body) == 0 {
		g.writeIndent()
		g.buf.WriteString("return false\n")
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaReduce generates code for arr.reduce(init) -> (acc, x) { expr }
func (g *Generator) genLambdaReduce(iterable ast.Expression, lambda *ast.LambdaExpr, initial ast.Expression) {
	g.needsRuntime = true
	accName := "_acc"
	varName := "_v"
	if len(lambda.Params) > 0 {
		accName = lambda.Params[0].Name
	}
	if len(lambda.Params) > 1 {
		varName = lambda.Params[1].Name
	}

	elemType := g.inferArrayElementGoType(iterable)
	// For reduce, the accumulator type is inferred from the initial value
	accType := g.inferExprGoType(initial)

	g.buf.WriteString("runtime.Reduce(")
	g.genExpr(iterable)
	g.buf.WriteString(", ")
	g.genExpr(initial)
	g.buf.WriteString(", func(")
	g.buf.WriteString(accName)
	g.buf.WriteString(" ")
	g.buf.WriteString(accType)
	g.buf.WriteString(", ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" ")
	g.buf.WriteString(elemType)
	g.buf.WriteString(") ")
	g.buf.WriteString(accType)
	g.buf.WriteString(" {\n")

	g.indent++
	for i, stmt := range lambda.Body {
		if i == len(lambda.Body)-1 {
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
	if len(lambda.Body) == 0 {
		g.writeIndent()
		g.buf.WriteString("return ")
		g.buf.WriteString(accName)
		g.buf.WriteString("\n")
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaTimes generates code for n.times -> (i) { body }
func (g *Generator) genLambdaTimes(count ast.Expression, lambda *ast.LambdaExpr) {
	varName := "_i"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}

	g.buf.WriteString("for ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := 0; ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" < ")
	g.genExpr(count)
	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	g.buf.WriteString("++ {\n")

	// Set inInlinedLambda so that bare return becomes continue
	prevInInlinedLambda := g.inInlinedLambda
	g.inInlinedLambda = true
	g.scopedVar(varName, "int", func() {
		g.indent++
		for _, stmt := range lambda.Body {
			g.genStatement(stmt)
		}
		g.indent--
	})
	g.inInlinedLambda = prevInInlinedLambda

	g.writeIndent()
	g.buf.WriteString("}")
}

// genLambdaUpto generates code for start.upto(end) -> (i) { body }
func (g *Generator) genLambdaUpto(start ast.Expression, lambda *ast.LambdaExpr, end ast.Expression) {
	varName := "_i"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}

	g.buf.WriteString("for ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := ")
	g.genExpr(start)
	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" <= ")
	g.genExpr(end)
	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	g.buf.WriteString("++ {\n")

	// Set inInlinedLambda so that bare return becomes continue
	prevInInlinedLambda := g.inInlinedLambda
	g.inInlinedLambda = true
	g.scopedVar(varName, "int", func() {
		g.indent++
		for _, stmt := range lambda.Body {
			g.genStatement(stmt)
		}
		g.indent--
	})
	g.inInlinedLambda = prevInInlinedLambda

	g.writeIndent()
	g.buf.WriteString("}")
}

// genLambdaDownto generates code for start.downto(end) -> (i) { body }
func (g *Generator) genLambdaDownto(start ast.Expression, lambda *ast.LambdaExpr, end ast.Expression) {
	varName := "_i"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}

	g.buf.WriteString("for ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := ")
	g.genExpr(start)
	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" >= ")
	g.genExpr(end)
	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	g.buf.WriteString("-- {\n")

	// Set inInlinedLambda so that bare return becomes continue
	prevInInlinedLambda := g.inInlinedLambda
	g.inInlinedLambda = true
	g.scopedVar(varName, "int", func() {
		g.indent++
		for _, stmt := range lambda.Body {
			g.genStatement(stmt)
		}
		g.indent--
	})
	g.inInlinedLambda = prevInInlinedLambda

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

	// Extract base class name from element type (e.g., "*User" -> "User")
	baseClass := strings.TrimPrefix(elemType, "*")

	// Check if this is a user-defined class with the method
	// If so, we can generate a direct method call (required for unexported methods)
	// Otherwise, use runtime.CallMethod (required for primitives)
	goMethodName := ""
	if g.instanceMethods[baseClass] != nil {
		goMethodName = g.instanceMethods[baseClass][stp.Method]
	}

	g.buf.WriteString(method.runtimeFunc)
	g.buf.WriteString("(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(x ")
	g.buf.WriteString(elemType)
	g.buf.WriteString(") ")
	g.buf.WriteString(method.returnType)

	if goMethodName != "" {
		// User-defined class with known method - generate direct call
		g.buf.WriteString(" { return x.")
		g.buf.WriteString(goMethodName)
		g.buf.WriteString("() })")
	} else {
		// Primitive type or unknown method - use runtime.CallMethod
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
	// Handle special selector expressions: error interface, channel operations, etc.
	switch sel.Sel {
	case "error":
		// .error on error types -> .Error() (Go's error interface)
		g.genExpr(sel.X)
		g.buf.WriteString(".Error()")
		return
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
		// Chan<T>.new -> make(chan T)
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
	if ident, ok := sel.X.(*ast.Ident); ok && g.isClass(ident.Name) {
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

	// Handle enum methods (.value, .to_s, .values)
	receiverType := g.inferTypeFromExpr(sel.X)
	if g.enums != nil {
		if _, isEnum := g.enums[receiverType]; isEnum {
			switch sel.Sel {
			case "value":
				// enum.value -> int(enum)
				g.buf.WriteString("int(")
				g.genExpr(sel.X)
				g.buf.WriteString(")")
				return
			case "to_s":
				// enum.to_s -> enum.String()
				g.genExpr(sel.X)
				g.buf.WriteString(".String()")
				return
			}
		}
		// Handle enum type methods (Color.values)
		if ident, ok := sel.X.(*ast.Ident); ok {
			if _, isEnum := g.enums[ident.Name]; isEnum {
				switch sel.Sel {
				case "values":
					// Color.values -> ColorValues()
					g.buf.WriteString(ident.Name)
					g.buf.WriteString("Values()")
					return
				}
			}
		}
	}

	// Handle struct field access - convert snake_case to PascalCase
	if g.structs != nil {
		if _, isStruct := g.structs[receiverType]; isStruct {
			g.genExpr(sel.X)
			g.buf.WriteString(".")
			g.buf.WriteString(snakeToPascalWithAcronyms(sel.Sel))
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
		// For Go interop, don't add "Is" prefix to predicate methods
		g.buf.WriteString(snakeToPascalWithAcronyms(sel.Sel, true))
		return
	}

	// Check if this method name matches any interface method
	// If so, use PascalCase for Go interface compatibility
	isInterfaceMethod := g.isInterfaceMethod(sel.Sel)

	// Check if receiver is an instance of a pub class
	// Pub class methods use PascalCase
	receiverClassName := g.getReceiverClassName(sel.X)
	isPubClass := g.isPublicClass(receiverClassName)

	// Check if this specific accessor is pub (for non-pub classes with pub accessors)
	isPubAccessor := false
	if classAccessors := g.pubAccessors[receiverClassName]; classAccessors != nil {
		isPubAccessor = classAccessors[sel.Sel]
	}

	// Check if this specific method is pub (for non-pub classes with pub methods)
	isPubMethod := false
	if classMethods := g.pubMethods[receiverClassName]; classMethods != nil {
		isPubMethod = classMethods[sel.Sel]
	}

	// Check if this is a private method (for non-pub classes with private methods)
	isPrivateMethod := false
	if classMethods := g.privateMethods[receiverClassName]; classMethods != nil {
		isPrivateMethod = classMethods[sel.Sel]
	}

	g.genExpr(sel.X)
	g.buf.WriteString(".")

	// Special case: to_s maps to String (Go's Stringer interface convention)
	if sel.Sel == "to_s" {
		g.buf.WriteString("String")
		return
	}

	if isInterfaceMethod || isPubClass || isPubAccessor || isPubMethod {
		g.buf.WriteString(snakeToPascalWithAcronyms(sel.Sel))
	} else if isPrivateMethod {
		g.buf.WriteString("_" + snakeToCamelWithAcronyms(sel.Sel))
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
// or a Go struct literal like sync.WaitGroup{}.
// Returns true if the expression creates an instance of a Go type.
func (g *Generator) isGoInteropTypeConstructor(expr ast.Expression) bool {
	// Check for GoStructLit (sync.WaitGroup{})
	if goStruct, ok := expr.(*ast.GoStructLit); ok {
		return g.imports[goStruct.Package]
	}

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
	// Check explicit interfaces
	for _, ifaceName := range g.getAllInterfaceNames() {
		if slices.Contains(g.getInterfaceMethodNames(ifaceName), methodName) {
			return true
		}
	}
	// Check module-generated interfaces
	for _, modName := range g.getAllModuleNames() {
		if slices.Contains(g.getModuleMethodNames(modName), methodName) {
			return true
		}
	}
	return false
}

// getReceiverClassName extracts the class name from a receiver expression.
// For example, if expr is a variable of type *User, returns "User".
// Returns empty string if the class name cannot be determined.
func (g *Generator) getReceiverClassName(expr ast.Expression) string {
	// Use type info from semantic analysis
	if typ := g.typeInfo.GetRugbyType(expr); typ != "" {
		// Strip pointer/optional prefixes if present
		typ = strings.TrimPrefix(typ, "*")
		typ = strings.TrimSuffix(typ, "?")
		return typ
	}

	// Fall back to inferring from expression
	switch e := expr.(type) {
	case *ast.Ident:
		// Fall back to codegen's vars map (variable name -> type)
		if varType, ok := g.vars[e.Name]; ok && varType != "" {
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

// isClassInstanceMethod checks if the given method name is an instance method
// (including getters) on the given class or any of its parent classes.
// Returns the Go method name if found, empty string otherwise.
func (g *Generator) isClassInstanceMethod(className, methodName string) string {
	// Walk up the inheritance chain
	for className != "" {
		if methods := g.instanceMethods[className]; methods != nil {
			if goName, ok := methods[methodName]; ok {
				return goName
			}
		}
		// Move to parent class
		className = g.classParent[className]
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
	if strings.HasPrefix(receiverType, "[]") || receiverType == "Array" || strings.HasPrefix(receiverType, "Array<") {
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

	// Skip stdLib methods for user-defined class receivers (class methods take precedence)
	receiverClassName := g.getReceiverClassName(sel.X)
	if receiverClassName != "" {
		baseClassName := stripTypeParams(receiverClassName)
		if g.isClass(baseClassName) || g.isModuleScopedClass(baseClassName) {
			return false
		}
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
	// But skip if the receiver is a user-defined class (class methods take precedence)
	baseType := stripTypeParams(receiverType)
	if !g.isClass(baseType) && !g.isModuleScopedClass(baseType) {
		if def, ok := uniqueMethods[sel.Sel]; ok {
			g.needsRuntime = true
			g.buf.WriteString(def.RuntimeFunc)
			g.buf.WriteString("(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		}
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
