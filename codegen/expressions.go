package codegen

import (
	"fmt"
	"maps"
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
			if g.currentClass() != "" {
				g.buf.WriteString(receiverName(g.currentClass()))
			} else {
				g.addError(fmt.Errorf("line %d: 'self' used outside of class context", e.Line))
				g.buf.WriteString("nil")
			}
		} else if runtimeCall, ok := noParenKernelFuncs[e.Name]; ok {
			// Check for kernel functions that can be used without parens
			g.needsRuntime = true
			g.buf.WriteString(runtimeCall)
		} else if g.currentClass() != "" && g.isModuleMethod(e.Name) {
			// Module method call within class - generate as self.Method() (PascalCase for interface)
			recv := receiverName(g.currentClass())
			methodName := snakeToPascalWithAcronyms(e.Name)
			g.buf.WriteString(fmt.Sprintf("%s.%s()", recv, methodName))
		} else if g.currentClass() != "" && g.instanceMethods[g.currentClass()] != nil && g.instanceMethods[g.currentClass()][e.Name] != "" && g.vars[e.Name] == "" {
			// Implicit method call within class - generate as self.method()
			// Only if not a local variable (e.g., parameters shadow instance methods)
			recv := receiverName(g.currentClass())
			methodName := g.instanceMethods[g.currentClass()][e.Name]
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
		} else if g.currentClass() != "" {
			recv := receiverName(g.currentClass())
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
		if g.currentClass() != "" {
			key := g.currentClass() + "@@" + e.Name
			if varName, ok := g.classVars[key]; ok {
				g.buf.WriteString(varName)
			} else {
				// Fall back to generated name (for forward references)
				g.buf.WriteString(fmt.Sprintf("_%s_%s", g.currentClass(), e.Name))
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
		// Handle type parameter methods (T.zero) directly - not as method calls
		if ident, ok := e.X.(*ast.Ident); ok && g.currentFuncTypeParams() != nil {
			if constraint, isTypeParam := g.currentFuncTypeParams()[ident.Name]; isTypeParam {
				if e.Sel == "zero" && constraint == "Numeric" {
					// T.zero -> *new(T) for Numeric constraint
					g.buf.WriteString("*new(")
					g.buf.WriteString(ident.Name)
					g.buf.WriteString(")")
					break
				}
			}
		}

		// Handle .length and .size universally
		// This must come before selectorKind dispatch to handle Go interop types like []byte
		if e.Sel == "length" || e.Sel == "size" {
			// Range types use runtime.RangeSize, everything else uses len()
			receiverType := g.inferTypeFromExpr(e.X)
			if receiverType == "Range" || strings.HasPrefix(receiverType, "runtime.Range") {
				g.needsRuntime = true
				g.buf.WriteString("runtime.RangeSize(")
				g.genExpr(e.X)
				g.buf.WriteString(")")
			} else {
				g.buf.WriteString("len(")
				g.genExpr(e.X)
				g.buf.WriteString(")")
			}
			break
		}

		// Special case: channel close uses Go builtin close(ch), not ch.close()
		if e.Sel == "close" {
			recvType := g.inferTypeFromExpr(e.X)
			isChannel := strings.HasPrefix(recvType, "Chan<") || strings.HasPrefix(recvType, "chan ")
			// Also check TypeInfo for channel detection
			if !isChannel && g.typeInfo != nil {
				isChannel = g.typeInfo.GetTypeKind(e.X) == TypeChannel
			}
			// Fallback: if type is unknown and not a Go interop var, treat as channel
			// This maintains backward compatibility with tests that don't have type info
			if !isChannel && recvType == "" {
				if ident, ok := e.X.(*ast.Ident); ok && !g.goInteropVars[ident.Name] && !g.imports[ident.Name] {
					isChannel = true
				}
			}
			if isChannel {
				g.buf.WriteString("close(")
				g.genExpr(e.X)
				g.buf.WriteString(")")
				break
			}
		}

		// Use SelectorKind to determine if this is a field access or method call
		selectorKind := g.getSelectorKind(e)
		switch selectorKind {
		case ast.SelectorMethod:
			// Method calls need () - convert to CallExpr
			g.genCallExpr(&ast.CallExpr{Func: e, Args: nil})
		case ast.SelectorGetter:
			// Check if this is a declared accessor (getter/property) with a backing field
			receiverClassName := g.getReceiverClassName(e.X)
			hasAccessor := g.typeInfo != nil && g.typeInfo.HasAccessor(receiverClassName, e.Sel)

			if hasAccessor {
				// Inline getter to direct field access for performance
				// obj.field -> obj._field (same-package access)
				g.genExpr(e.X)
				g.buf.WriteString(".")
				g.buf.WriteString(privateFieldName(e.Sel))
			} else {
				// No declared accessor - treat as method call
				g.genCallExpr(&ast.CallExpr{Func: e, Args: nil})
			}
		case ast.SelectorGoMethod:
			// Go method calls need () - use actual Go name from semantic analysis
			g.genExpr(e.X)
			g.buf.WriteString(".")
			if goName := g.typeInfo.GetGoName(e); goName != "" {
				g.buf.WriteString(goName)
			} else {
				g.buf.WriteString(snakeToPascalWithAcronyms(e.Sel))
			}
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
	if g.currentClass() == "" || g.currentMethod() == "" {
		g.buf.WriteString("/* super outside method */")
		return
	}
	if len(g.currentClassEmbeds()) == 0 {
		g.addError(fmt.Errorf("line %d: super used in class without parent", e.Line))
		g.buf.WriteString("/* super without parent */")
		return
	}

	// Use the first embedded type as the parent class
	parentClass := g.currentClassEmbeds()[0]
	recv := receiverName(g.currentClass())

	// Handle super in constructor (initialize) specially:
	// Generate: recv.ParentClass = *newParentClass(args...)
	if g.currentMethod() == "initialize" {
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
	if g.method != nil && g.method.IsPub {
		methodName = snakeToPascalWithAcronyms(g.currentMethod())
	} else {
		methodName = snakeToCamelWithAcronyms(g.currentMethod())
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

// genLambdaExpr generates code for arrow lambda expressions: -> { |params| body }
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
		// Explicit return type annotation takes precedence
		g.buf.WriteString(" ")
		g.buf.WriteString(mapType(e.ReturnType))
	} else if g.forceAnyLambdaReturn && len(e.Body) > 0 && g.lambdaBodyNeedsReturn(e.Body) {
		// Go interop context: use 'any' for implicit returns since Go func types are not covariant
		g.buf.WriteString(" any")
	} else if inferredReturnType != "" && len(inferredParamTypes) > 0 {
		// Return type inferred from explicit function type context (variable declaration, etc.)
		// Use inferred type when param types are also inferred (the whole function type came from annotation)
		g.buf.WriteString(" ")
		g.buf.WriteString(inferredReturnType)
	} else if inferredReturnType != "" && len(e.Params) == 0 {
		// Zero-param lambda with inferred return type (likely from function return type context)
		g.buf.WriteString(" ")
		g.buf.WriteString(inferredReturnType)
	} else if len(e.Body) > 0 && g.lambdaBodyNeedsReturn(e.Body) {
		// Implicit return without explicit type context - use any for safety
		g.buf.WriteString(" any")
	} else if inferredReturnType != "" {
		// Use inferred return type for lambdas without implicit returns
		g.buf.WriteString(" ")
		g.buf.WriteString(inferredReturnType)
	}
	// Note: if body ends with a non-expression statement (assignment, control flow, etc.),
	// we don't add a return type, producing func() { ... } for Go interop compatibility

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

// lambdaBodyNeedsReturn checks if the lambda body ends with a statement that
// should be implicitly returned (expression statement or compound assignment).
// Returns false for bodies ending with regular assignments, control flow, etc.
// This allows lambdas passed to Go functions expecting func() to work correctly.
func (g *Generator) lambdaBodyNeedsReturn(body []ast.Statement) bool {
	if len(body) == 0 {
		return false
	}
	lastStmt := body[len(body)-1]
	switch lastStmt.(type) {
	case *ast.ExprStmt:
		// Expression statements are returned (e.g., `x + 1`, `call()`)
		return true
	case *ast.CompoundAssignStmt:
		// Compound assignments (x += 1) return the new value
		return true
	default:
		// Regular assignments, control flow, etc. don't produce a return value
		return false
	}
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
		// Check if left operand is a typed array - use generic Append
		if elemType := g.inferArrayElementGoType(e.Left); elemType != "" && elemType != "any" {
			g.needsRuntime = true
			g.buf.WriteString("runtime.Append(")
			g.genExpr(e.Left)
			g.buf.WriteString(", ")
			g.genExpr(e.Right)
			g.buf.WriteString(")")
			return
		}
		// Check if left operand is an integer - use native Go bitwise shift
		if g.isIntegerType(e.Left) {
			g.genExpr(e.Left)
			g.buf.WriteString(" << ")
			g.genExpr(e.Right)
			return
		}
		// Fallback to runtime.ShiftLeft for unknown types
		g.needsRuntime = true
		g.buf.WriteString("runtime.ShiftLeft(")
		g.genExpr(e.Left)
		g.buf.WriteString(", ")
		g.genExpr(e.Right)
		g.buf.WriteString(")")
		return
	}

	// >> is always bitwise right shift
	if e.Op == ">>" {
		// Check if left operand is an integer - use native Go bitwise shift
		if g.isIntegerType(e.Left) {
			g.genExpr(e.Left)
			g.buf.WriteString(" >> ")
			g.genExpr(e.Right)
			return
		}
		// Fallback to runtime.ShiftRight for unknown types
		g.needsRuntime = true
		g.buf.WriteString("runtime.ShiftRight(")
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
// isIntegerType returns true if the expression is known to be an integer type.
// This includes integer literals and variables/expressions typed as Int.
func (g *Generator) isIntegerType(expr ast.Expression) bool {
	// Integer literals are always integers
	if _, ok := expr.(*ast.IntLit); ok {
		return true
	}

	// Check inferred type
	inferredType := g.inferTypeFromExpr(expr)
	if inferredType == "Int" || inferredType == "int" {
		return true
	}

	// Check semantic type info
	if goType := g.typeInfo.GetGoType(expr); goType == "int" {
		return true
	}

	return false
}

func (g *Generator) hasAnyTypedOperand(left, right ast.Expression) bool {
	leftKind := g.typeInfo.GetTypeKind(left)
	rightKind := g.typeInfo.GetTypeKind(right)
	return leftKind == TypeAny && rightKind == TypeAny
}

// canUseDirectComparison returns true if we can use Go's == operator directly
// instead of runtime.Equal. This is possible when:
// 1. Either operand is nil (Go's == works for any nullable type)
// 2. Both operands are primitive literals, OR
// 3. Both operands are primitive types (Int, Float, String, Bool)
func (g *Generator) canUseDirectComparison(left, right ast.Expression) bool {
	// Nil comparisons always work with direct == in Go
	_, leftIsNil := left.(*ast.NilLit)
	_, rightIsNil := right.(*ast.NilLit)
	if leftIsNil || rightIsNil {
		return true
	}

	// Primitive literals can always use direct comparison
	if isPrimitiveLiteral(left) && isPrimitiveLiteral(right) {
		return true
	}

	// Check if both types are primitive, using enhanced inference
	leftKind := g.inferPrimitiveKind(left)
	rightKind := g.inferPrimitiveKind(right)
	return isPrimitiveKind(leftKind) && isPrimitiveKind(rightKind)
}

// inferPrimitiveKind tries to determine the primitive type of an expression
// by checking typeInfo first, then falling back to local variable info and
// structural inference for binary expressions.
func (g *Generator) inferPrimitiveKind(expr ast.Expression) TypeKind {
	// First try typeInfo
	kind := g.typeInfo.GetTypeKind(expr)
	if isPrimitiveKind(kind) {
		return kind
	}

	// Check for primitive literals
	switch expr.(type) {
	case *ast.IntLit:
		return TypeInt
	case *ast.FloatLit:
		return TypeFloat
	case *ast.BoolLit:
		return TypeBool
	case *ast.StringLit:
		return TypeString
	}

	// Check identifiers against vars map
	if ident, ok := expr.(*ast.Ident); ok {
		if goType, exists := g.vars[ident.Name]; exists {
			return goTypeToKind(goType)
		}
	}

	// For arithmetic binary expressions, infer from operands
	if binExpr, ok := expr.(*ast.BinaryExpr); ok {
		switch binExpr.Op {
		case "+", "-", "*", "/", "%":
			// Arithmetic operations preserve numeric type
			leftKind := g.inferPrimitiveKind(binExpr.Left)
			if leftKind == TypeInt || leftKind == TypeFloat || leftKind == TypeInt64 {
				return leftKind
			}
			rightKind := g.inferPrimitiveKind(binExpr.Right)
			if rightKind == TypeInt || rightKind == TypeFloat || rightKind == TypeInt64 {
				return rightKind
			}
		case "==", "!=", "<", ">", "<=", ">=", "and", "or":
			return TypeBool
		}
	}

	return kind
}

// goTypeToKind converts a Go type string to a TypeKind.
func goTypeToKind(goType string) TypeKind {
	switch goType {
	case "int":
		return TypeInt
	case "int64":
		return TypeInt64
	case "float64":
		return TypeFloat
	case "bool":
		return TypeBool
	case "string":
		return TypeString
	default:
		return TypeUnknown
	}
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
	case *ast.IntLit, *ast.FloatLit, *ast.BoolLit, *ast.StringLit, *ast.NilLit:
		return true
	default:
		return false
	}
}

// isZeroValue returns true if the expression is a zero value literal.
// Zero values: 0, 0.0, false, "", nil
// When filling arrays with zero values, we can use make() instead of runtime.Fill
// since Go automatically zeroes allocated memory.
func isZeroValue(e ast.Expression) bool {
	switch v := e.(type) {
	case *ast.IntLit:
		return v.Value == 0
	case *ast.FloatLit:
		return v.Value == 0.0
	case *ast.BoolLit:
		return !v.Value // false is zero value
	case *ast.StringLit:
		return v.Value == ""
	case *ast.NilLit:
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
	rightType := g.inferTypeFromExpr(e.Right)

	// Check if the left side is a safe navigation expression (returns any)
	_, isSafeNav := e.Left.(*ast.SafeNavExpr)

	// Check if right side is optional (for chaining: a ?? b ?? c)
	isRightOptional := strings.HasSuffix(rightType, "?")

	// Handle known value-type optionals with generic Coalesce
	switch leftType {
	case "Int?", "Int64?", "Float?", "Bool?":
		if isRightOptional {
			g.buf.WriteString("runtime.CoalesceOpt(")
		} else {
			g.buf.WriteString("runtime.Coalesce(")
		}
	case "String?":
		// Use CoalesceStringAny for safe navigation (which returns any)
		if isSafeNav {
			g.buf.WriteString("runtime.CoalesceStringAny(")
		} else if isRightOptional {
			g.buf.WriteString("runtime.CoalesceOpt(")
		} else {
			g.buf.WriteString("runtime.Coalesce(")
		}
	default:
		varName := fmt.Sprintf("_nc%d", g.tempVarCounter)
		g.tempVarCounter++
		g.buf.WriteString("func() ")
		// Use rightType already computed above
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
	// Collect the entire chain of safe nav expressions
	// For a&.b&.c, we collect [a, b, c] and the base receiver
	chain := []string{e.Selector}
	current := e.Receiver
	for {
		if sn, ok := current.(*ast.SafeNavExpr); ok {
			chain = append([]string{sn.Selector}, chain...)
			current = sn.Receiver
		} else {
			break
		}
	}
	// Now 'current' is the base receiver (e.g., user.address for user.address&.city&.name)
	// and 'chain' is [city, name]

	// Generate flattened code that preserves types:
	// func() any {
	//   _sn0 := user.address(); if _sn0 == nil { return nil }
	//   _sn1 := (*_sn0).city(); if _sn1 == nil { return nil }
	//   return (*_sn1).name()
	// }()

	g.buf.WriteString("func() any { ")

	// Generate code for the base receiver
	baseVar := fmt.Sprintf("_sn%d", g.tempVarCounter)
	g.tempVarCounter++

	g.buf.WriteString(baseVar)
	g.buf.WriteString(" := ")

	// Generate the base receiver expression
	if sel, ok := current.(*ast.SelectorExpr); ok {
		selectorKind := g.getSelectorKind(sel)
		g.genExpr(current)
		if selectorKind != ast.SelectorMethod && selectorKind != ast.SelectorGetter &&
			selectorKind != ast.SelectorGoMethod && !strings.HasPrefix(sel.Sel, "_") {
			g.buf.WriteString("()")
		}
	} else {
		g.genExpr(current)
	}

	g.buf.WriteString("; if ")
	g.buf.WriteString(baseVar)
	g.buf.WriteString(" == nil { return nil }; ")

	// Generate code for each step in the chain
	prevVar := baseVar
	for i, selector := range chain {
		isLast := i == len(chain)-1
		currentVar := fmt.Sprintf("_sn%d", g.tempVarCounter)
		g.tempVarCounter++

		if isLast {
			// Last step: return the result
			if selector == "length" || selector == "size" {
				// .length and .size use len() for slices/strings/maps
				g.buf.WriteString("return len((*")
				g.buf.WriteString(prevVar)
				g.buf.WriteString("))")
			} else {
				g.buf.WriteString("return (*")
				g.buf.WriteString(prevVar)
				g.buf.WriteString(").")
				g.buf.WriteString(snakeToCamelWithAcronyms(selector))
				g.buf.WriteString("()")
			}
		} else {
			// Intermediate step: assign and check nil
			g.buf.WriteString(currentVar)
			if selector == "length" || selector == "size" {
				// .length and .size use len() - but this shouldn't be intermediate
				// since len() returns int, not a pointer. This is likely an error case.
				g.buf.WriteString(" := len((*")
				g.buf.WriteString(prevVar)
				g.buf.WriteString(")); _ = ")
				g.buf.WriteString(currentVar)
				g.buf.WriteString("; ")
			} else {
				g.buf.WriteString(" := (*")
				g.buf.WriteString(prevVar)
				g.buf.WriteString(").")
				g.buf.WriteString(snakeToCamelWithAcronyms(selector))
				g.buf.WriteString("(); if ")
				g.buf.WriteString(currentVar)
				g.buf.WriteString(" == nil { return nil }; ")
			}
			prevVar = currentVar
		}
	}

	g.buf.WriteString(" }()")
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

	// Extract element type for tuple elements
	// If array type is []struct{...}, we need to generate tuple literals as struct literals
	// Handle both "[]struct{" and "[]struct{ " (with space)
	elemType := ""
	if strings.HasPrefix(arrayType, "[]struct{") || strings.HasPrefix(arrayType, "[]struct ") {
		elemType = arrayType[2:] // Remove "[]" prefix
	}

	g.buf.WriteString(arrayType + "{")
	for i, elem := range arr.Elements {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		// If element is a tuple literal and we have a struct element type, generate as struct
		if tupleLit, ok := elem.(*ast.TupleLit); ok && elemType != "" {
			g.buf.WriteString(elemType)
			g.buf.WriteString("{")
			for j, tupleElem := range tupleLit.Elements {
				if j > 0 {
					g.buf.WriteString(", ")
				}
				g.genExpr(tupleElem)
			}
			g.buf.WriteString("}")
		} else {
			g.genExpr(elem)
		}
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
			// Classes are reference types - need pointer prefix in Go
			if g.typeInfo != nil && g.typeInfo.IsClass(firstType) {
				elemType = "*" + elemType
			}
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

// genTupleLitAsStruct generates a Go anonymous struct literal for a tuple.
// Used when assigning a tuple literal to a variable with a tuple type annotation.
// Example: x: (String, Int) = ("hello", 42) -> var x struct{ _0 string; _1 int } = struct{ _0 string; _1 int }{"hello", 42}
func (g *Generator) genTupleLitAsStruct(t *ast.TupleLit, tupleType string) {
	goType := mapTupleType(tupleType)
	g.buf.WriteString(goType)
	g.buf.WriteString("{")
	for i, elem := range t.Elements {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.genExpr(elem)
	}
	g.buf.WriteString("}")
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
		leftType := g.inferTypeFromExpr(idx.Left)

		// For "any" type receivers, use runtime.GetKey to handle type assertion
		if leftType == "any" {
			g.needsRuntime = true
			g.buf.WriteString("runtime.GetKey(")
			g.genExpr(idx.Left)
			g.buf.WriteString(", ")
			g.genExpr(idx.Index)
			g.buf.WriteString(")")
			return
		}

		// For maps with string keys, use direct Go indexing
		// Class values return nil for missing keys (pointer is nullable)
		// Value types return zero value for missing keys (Go semantics)
		// Use .get(key) method for optional-returning access
		if strings.HasPrefix(leftType, "map[") || g.typeInfo.GetTypeKind(idx.Left) == TypeMap {
			g.genExpr(idx.Left)
			g.buf.WriteString("[")
			g.genExpr(idx.Index)
			g.buf.WriteString("]")
			return
		}

		// For other types (like structs with map fields via method call), use GetKey
		g.needsRuntime = true
		g.buf.WriteString("runtime.GetKey(")
		g.genExpr(idx.Left)
		g.buf.WriteString(", ")
		g.genExpr(idx.Index)
		g.buf.WriteString(")")
		return
	}

	// For strings, use runtime.StringIndex to return characters (not bytes)
	// This preserves type information (returns string, not any)
	leftType := g.inferTypeFromExpr(idx.Left)
	if leftType == "string" || leftType == "String" {
		g.needsRuntime = true
		g.buf.WriteString("runtime.StringIndex(")
		g.genExpr(idx.Left)
		g.buf.WriteString(", ")
		g.genExpr(idx.Index)
		g.buf.WriteString(")")
		return
	}

	// For maps, use direct Go indexing
	// Class values return nil for missing keys (pointer is nullable)
	// Value types return zero value for missing keys (Go semantics)
	// Use .get(key) method for optional-returning access
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

	// For variable indices, we need runtime support for negative index handling.
	g.needsRuntime = true

	// Try to get element type from TypeInfo or from the Rugby type string
	elemType := g.typeInfo.GetElementType(idx.Left)
	if elemType == "" || elemType == "unknown" || strings.Contains(elemType, "unknown") {
		// Fallback: extract from type string like "Array<Int>"
		elemType = extractArrayElementType(leftType)
	}
	if elemType == "" || elemType == "unknown" || strings.Contains(elemType, "unknown") {
		// Additional fallback: for nested index expressions (2D arrays),
		// try to infer from the identifier's type in g.vars
		elemType = g.inferIndexedElementType(idx.Left)
	}

	if elemType != "" && elemType != "any" && elemType != "unknown" && !strings.Contains(elemType, "unknown") {
		goElemType := mapType(elemType)
		// Classes are reference types - need pointer prefix in Go
		if g.typeInfo != nil && g.typeInfo.IsClass(elemType) {
			goElemType = "*" + goElemType
		}
		// Use generic SliceAt[T] for typed slices - much faster than AtIndex
		g.buf.WriteString("runtime.SliceAt[")
		g.buf.WriteString(goElemType)
		g.buf.WriteString("](")
		g.genExpr(idx.Left)
		g.buf.WriteString(", ")
		g.genExpr(idx.Index)
		g.buf.WriteString(")")
	} else {
		// Fallback to AtIndex for unknown types
		g.buf.WriteString("runtime.AtIndex(")
		g.genExpr(idx.Left)
		g.buf.WriteString(", ")
		g.genExpr(idx.Index)
		g.buf.WriteString(")")
	}
}

// inferIndexedElementType attempts to infer the element type for an indexed expression.
// For example, for xloc[i] where xloc has type Array<Array<Float>>,
// this returns "Array<Float>" (the element type after one level of indexing).
func (g *Generator) inferIndexedElementType(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Ident:
		// Get the variable's type from semantic analysis and extract element type
		if rugbyType := g.typeInfo.GetRugbyType(e); rugbyType != "" && rugbyType != "unknown" {
			return extractArrayElementType(rugbyType)
		}
	case *ast.IndexExpr:
		// For nested indexing, first get the collection's type, then extract element type
		collType := g.inferTypeFromExpr(e.Left)
		if collType != "" {
			// Extract element type from the collection type
			elemType := extractArrayElementType(collType)
			if elemType != "" {
				return elemType
			}
		}
	}
	return ""
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
	case *ast.Ident:
		// Check if this variable is known to be non-negative
		return g.nonNegativeVars[e.Name]
	case *ast.BinaryExpr:
		// Binary expressions with non-negative operands and safe operators
		// are also non-negative (e.g., i + 1, i * 2)
		if e.Op == "+" || e.Op == "*" {
			return g.isNonNegativeExpr(e.Left) && g.isNonNegativeExpr(e.Right)
		}
		return false
	default:
		return false
	}
}

// isNonNegativeExpr checks if an expression is known to be non-negative.
func (g *Generator) isNonNegativeExpr(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.IntLit:
		return e.Value >= 0
	case *ast.Ident:
		return g.nonNegativeVars[e.Name]
	case *ast.BinaryExpr:
		if e.Op == "+" || e.Op == "*" {
			return g.isNonNegativeExpr(e.Left) && g.isNonNegativeExpr(e.Right)
		}
		return false
	default:
		return false
	}
}

// extractGenericTypeString recursively extracts a type string from an IndexExpr AST node.
// For Array<Int>, returns "Array<Int>". For Array<Array<Float>>, returns "Array<Array<Float>>".
// This handles nested generic types that the semantic analyzer may not fully resolve.
func extractGenericTypeString(idx *ast.IndexExpr) string {
	baseIdent, ok := idx.Left.(*ast.Ident)
	if !ok {
		return ""
	}

	// Simple case: Array<Int> where Index is an Ident
	if typeParamIdent, ok := idx.Index.(*ast.Ident); ok {
		return baseIdent.Name + "<" + typeParamIdent.Name + ">"
	}

	// Nested case: Array<Array<Float>> where Index is another IndexExpr
	if nestedIdx, ok := idx.Index.(*ast.IndexExpr); ok {
		innerType := extractGenericTypeString(nestedIdx)
		if innerType != "" {
			return baseIdent.Name + "<" + innerType + ">"
		}
	}

	return ""
}

// extractArrayElementType extracts the element type from a Rugby type string.
// For "Array<Int>" returns "Int", for "Array<Body>" returns "Body".
// For "[]byte" returns "byte", for "[]int" returns "int" (Go slice syntax).
// Returns empty string if not an array type.
func extractArrayElementType(rugbyType string) string {
	if strings.HasPrefix(rugbyType, "Array<") && strings.HasSuffix(rugbyType, ">") {
		return rugbyType[6 : len(rugbyType)-1]
	}
	// Handle Go slice syntax: []T
	if strings.HasPrefix(rugbyType, "[]") {
		return rugbyType[2:]
	}
	return ""
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

func (g *Generator) genSpawnExpr(e *ast.SpawnExpr) {
	// spawn { expr } or spawn do ... end
	// Variables from enclosing scope are captured by reference (closure semantics)

	// Check if the block has a void return type (last expression returns nothing)
	isVoid := g.isBlockVoid(e.Block.Body)

	if isVoid {
		// Void block - inline to bare go statement (fire-and-forget)
		// No runtime import needed for this case
		g.buf.WriteString("go func() {\n")
		g.indent++
		for _, stmt := range e.Block.Body {
			g.genStatement(stmt)
		}
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}()")
	} else {
		// Value-returning block - use Spawn and return Task<T>
		g.needsRuntime = true
		returnType := g.inferBlockBodyReturnType(e.Block.Body)

		g.buf.WriteString("runtime.Spawn(func() ")
		g.buf.WriteString(returnType)
		g.buf.WriteString(" {\n")
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
						g.genZeroReturn(returnType)
					}
				} else {
					g.genStatement(stmt)
				}
			}
		} else {
			g.writeIndent()
			g.genZeroReturn(returnType)
		}
		g.indent--
		g.writeIndent()
		g.buf.WriteString("})")
	}
}

// isBlockVoid returns true if the block's last expression returns void (no value).
func (g *Generator) isBlockVoid(body []ast.Statement) bool {
	if len(body) == 0 {
		return true
	}

	lastStmt := body[len(body)-1]
	exprStmt, ok := lastStmt.(*ast.ExprStmt)
	if !ok {
		// Not an expression statement (e.g., assignment, if, while)
		return true
	}

	// Check if the expression is a void function call
	call, ok := exprStmt.Expr.(*ast.CallExpr)
	if !ok {
		return false // Other expressions (literals, etc.) have values
	}

	// Check if this is a user-defined function with no return type
	if ident, ok := call.Func.(*ast.Ident); ok {
		if fn, exists := g.functions[ident.Name]; exists {
			// Function is defined - check if it has a return type
			return len(fn.ReturnTypes) == 0
		}
	}

	// For unknown functions, check type inference
	returnType := g.inferExprGoTypeDeep(call)
	return returnType == ""
}

// genZeroReturn generates a return statement with the zero value for the given type.
func (g *Generator) genZeroReturn(returnType string) {
	if returnType == "any" {
		g.buf.WriteString("return nil\n")
	} else {
		g.buf.WriteString("var _zero ")
		g.buf.WriteString(returnType)
		g.buf.WriteString("\n")
		g.writeIndent()
		g.buf.WriteString("return _zero\n")
	}
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
	g.goInteropVars[scopeVar] = true // Mark as Go interop for PascalCase method calls

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
