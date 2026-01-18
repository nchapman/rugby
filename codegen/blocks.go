package codegen

// Block and iteration method code generation.
// This file handles:
//   - Block calls (arr.each { |x| ... })
//   - Lambda iteration (arr.map -> { |x| ... })
//   - Optional iteration methods (optional.map, .filter, .each)
//   - Integer iteration (times, upto, downto)
//   - Map iteration (each with key/value)

import (
	"fmt"
	"strings"

	"github.com/nchapman/rugby/ast"
)

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
		case "flat_map":
			g.genOptionalFlatMapBlock(sel.X, block, receiverType)
			return
		case "filter":
			g.genOptionalFilterBlock(sel.X, block, receiverType)
			return
		case "each":
			g.genOptionalEachBlock(sel.X, block, receiverType)
			return
		}
	}

	// Check if receiver is a map type with 2-param block - needs special handling for select/reject
	if goType := g.typeInfo.GetGoType(sel.X); strings.HasPrefix(goType, "map[") && len(block.Params) == 2 {
		switch method {
		case "select", "filter":
			g.genMapSelectBlock(sel.X, block, goType)
			return
		case "reject":
			g.genMapRejectBlock(sel.X, block, goType)
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

// genOptionalFlatMapBlock generates code for optional.flat_map { |x| ... }
// Returns the result directly (flattened) if block returns an optional.
func (g *Generator) genOptionalFlatMapBlock(receiver ast.Expression, block *ast.BlockExpr, optType string) {
	varName := "_optVal"
	blockParam := "_"
	if len(block.Params) > 0 {
		blockParam = block.Params[0]
	}

	// Determine the inner type (without ?)
	innerType := strings.TrimSuffix(optType, "?")
	goType := mapType(innerType)

	// For flat_map, the block returns an optional (*R), and we return that directly
	// Generate: func() *R { if opt := receiver; opt != nil { return block(*opt) }; return nil }()
	g.buf.WriteString("func() ")

	// Infer return type from block - assume it returns an optional (pointer type)
	blockReturnGoType := "*string" // default fallback
	if len(block.Body) > 0 {
		if exprStmt, ok := block.Body[len(block.Body)-1].(*ast.ExprStmt); ok {
			if inferredType := g.inferTypeFromExpr(exprStmt.Expr); inferredType != "" {
				// The block should return an optional T?, which is *T in Go
				inferredGoType := mapType(strings.TrimSuffix(inferredType, "?"))
				blockReturnGoType = "*" + inferredGoType
			}
		}
	}
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
	if blockParam != "_" {
		g.buf.WriteString("; _ = ")
		g.buf.WriteString(blockParam)
	}
	g.buf.WriteString("; return ")

	// Generate block body - last expression is the result (should be an optional)
	if len(block.Body) > 0 {
		lastStmt := block.Body[len(block.Body)-1]
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			g.genExpr(exprStmt.Expr)
		} else {
			g.buf.WriteString("nil")
		}
	} else {
		g.buf.WriteString("nil")
	}

	g.buf.WriteString(" }; return nil }()")
	_ = goType
}

// genOptionalFilterBlock generates code for optional.filter { |x| ... }
// Returns the original value if predicate is true, nil otherwise.
func (g *Generator) genOptionalFilterBlock(receiver ast.Expression, block *ast.BlockExpr, optType string) {
	varName := "_optVal"
	blockParam := "_"
	if len(block.Params) > 0 {
		blockParam = block.Params[0]
	}

	// Determine the inner type (without ?)
	innerType := strings.TrimSuffix(optType, "?")
	goType := mapType(innerType)

	// Generate: func() *T { if opt := receiver; opt != nil && predicate(*opt) { return opt }; return nil }()
	g.buf.WriteString("func() *")
	g.buf.WriteString(goType)
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
	if blockParam != "_" {
		g.buf.WriteString("; _ = ")
		g.buf.WriteString(blockParam)
	}
	g.buf.WriteString("; if ")

	// Generate block body - last expression should be a boolean predicate
	if len(block.Body) > 0 {
		lastStmt := block.Body[len(block.Body)-1]
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			g.genExpr(exprStmt.Expr)
		} else {
			g.buf.WriteString("true")
		}
	} else {
		g.buf.WriteString("true")
	}

	g.buf.WriteString(" { return ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" } }; return nil }()")
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

// genOptionalMapLambda generates code for optional.map -> { |x| ... }
// Returns *ResultType (pointer) to maintain optional semantics for nil coalescing.
func (g *Generator) genOptionalMapLambda(receiver ast.Expression, lambda *ast.LambdaExpr, optType string) {
	varName := "_optVal"
	lambdaParam := "_"
	if len(lambda.Params) > 0 {
		lambdaParam = lambda.Params[0].Name
	}

	// Determine the inner type (without ?)
	innerType := strings.TrimSuffix(optType, "?")
	goType := mapType(innerType)

	// Infer the lambda's return type from its last expression
	lambdaReturnGoType := "string" // default fallback
	if len(lambda.Body) > 0 {
		if exprStmt, ok := lambda.Body[len(lambda.Body)-1].(*ast.ExprStmt); ok {
			if inferredType := g.inferTypeFromExpr(exprStmt.Expr); inferredType != "" {
				lambdaReturnGoType = mapType(inferredType)
			}
		}
	}

	// Generate: func() *ResultType { if opt := receiver; opt != nil { result := lambda(*opt); return &result }; return nil }()
	g.buf.WriteString("func() *")
	g.buf.WriteString(lambdaReturnGoType)
	g.buf.WriteString(" { if ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := ")
	g.genExpr(receiver)
	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" != nil { ")
	g.buf.WriteString(lambdaParam)
	g.buf.WriteString(" := *")
	g.buf.WriteString(varName)
	// Declare _ to avoid unused variable if lambdaParam is used
	if lambdaParam != "_" {
		g.buf.WriteString("; _ = ")
		g.buf.WriteString(lambdaParam)
	}
	g.buf.WriteString("; _result := ")

	// Generate lambda body - last expression is the result
	if len(lambda.Body) > 0 {
		lastStmt := lambda.Body[len(lambda.Body)-1]
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			g.genExpr(exprStmt.Expr)
		} else {
			g.buf.WriteString("\"\"") // fallback
		}
	} else {
		g.buf.WriteString("\"\"")
	}

	g.buf.WriteString("; return &_result }; return nil }()")
	_ = goType // inner type used for lambda param binding
}

// genOptionalFlatMapLambda generates code for optional.flat_map -> { |x| ... }
// Returns the result directly (flattened) if lambda returns an optional.
func (g *Generator) genOptionalFlatMapLambda(receiver ast.Expression, lambda *ast.LambdaExpr, optType string) {
	varName := "_optVal"
	lambdaParam := "_"
	if len(lambda.Params) > 0 {
		lambdaParam = lambda.Params[0].Name
	}

	// Determine the inner type (without ?)
	innerType := strings.TrimSuffix(optType, "?")
	goType := mapType(innerType)

	// For flat_map, the lambda returns an optional (*R), and we return that directly
	// Generate: func() *R { if opt := receiver; opt != nil { return lambda(*opt) }; return nil }()
	g.buf.WriteString("func() ")

	// Infer return type from lambda - assume it returns an optional (pointer type)
	lambdaReturnGoType := "*string" // default fallback
	if len(lambda.Body) > 0 {
		if exprStmt, ok := lambda.Body[len(lambda.Body)-1].(*ast.ExprStmt); ok {
			if inferredType := g.inferTypeFromExpr(exprStmt.Expr); inferredType != "" {
				// The lambda should return an optional T?, which is *T in Go
				inferredGoType := mapType(strings.TrimSuffix(inferredType, "?"))
				lambdaReturnGoType = "*" + inferredGoType
			}
		}
	}
	g.buf.WriteString(lambdaReturnGoType)
	g.buf.WriteString(" { if ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := ")
	g.genExpr(receiver)
	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" != nil { ")
	g.buf.WriteString(lambdaParam)
	g.buf.WriteString(" := *")
	g.buf.WriteString(varName)
	if lambdaParam != "_" {
		g.buf.WriteString("; _ = ")
		g.buf.WriteString(lambdaParam)
	}
	g.buf.WriteString("; return ")

	// Generate lambda body - last expression is the result (should be an optional)
	if len(lambda.Body) > 0 {
		lastStmt := lambda.Body[len(lambda.Body)-1]
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			g.genExpr(exprStmt.Expr)
		} else {
			g.buf.WriteString("nil")
		}
	} else {
		g.buf.WriteString("nil")
	}

	g.buf.WriteString(" }; return nil }()")
	_ = goType
}

// genOptionalFilterLambda generates code for optional.filter -> { |x| ... }
// Returns the original value if predicate is true, nil otherwise.
func (g *Generator) genOptionalFilterLambda(receiver ast.Expression, lambda *ast.LambdaExpr, optType string) {
	varName := "_optVal"
	lambdaParam := "_"
	if len(lambda.Params) > 0 {
		lambdaParam = lambda.Params[0].Name
	}

	// Determine the inner type (without ?)
	innerType := strings.TrimSuffix(optType, "?")
	goType := mapType(innerType)

	// Generate: func() *T { if opt := receiver; opt != nil && predicate(*opt) { return opt }; return nil }()
	g.buf.WriteString("func() *")
	g.buf.WriteString(goType)
	g.buf.WriteString(" { if ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := ")
	g.genExpr(receiver)
	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" != nil { ")
	g.buf.WriteString(lambdaParam)
	g.buf.WriteString(" := *")
	g.buf.WriteString(varName)
	if lambdaParam != "_" {
		g.buf.WriteString("; _ = ")
		g.buf.WriteString(lambdaParam)
	}
	g.buf.WriteString("; if ")

	// Generate lambda body - last expression should be a boolean predicate
	if len(lambda.Body) > 0 {
		lastStmt := lambda.Body[len(lambda.Body)-1]
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			g.genExpr(exprStmt.Expr)
		} else {
			g.buf.WriteString("true")
		}
	} else {
		g.buf.WriteString("true")
	}

	g.buf.WriteString(" { return ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" } }; return nil }()")
}

// genOptionalEachLambda generates code for optional.each -> { |x| ... }
func (g *Generator) genOptionalEachLambda(receiver ast.Expression, lambda *ast.LambdaExpr, optType string) {
	g.needsRuntime = true

	varName := "_"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
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
	for _, stmt := range lambda.Body {
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
// For example: arr.each -> { |x| puts x } becomes a for loop calling the lambda.
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

	// Check for set iteration (single parameter iterating over keys)
	// Sets are map[T]struct{}, so we iterate with `for k := range` not `for _, v := range`
	rugbyType := g.inferTypeFromExpr(iterable)
	isSetIteration := rugbyType == "Set" || strings.HasPrefix(rugbyType, "Set<")

	varName := "_v"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}

	if isSetIteration {
		// Extract element type from Set<T> or use any for plain Set
		elemType := "any"
		if strings.HasPrefix(rugbyType, "Set<") && strings.HasSuffix(rugbyType, ">") {
			inner := rugbyType[4 : len(rugbyType)-1]
			elemType = mapType(inner)
		}

		// For sets, iterate over keys: for k := range set
		g.buf.WriteString("for ")
		g.buf.WriteString(varName)
		g.buf.WriteString(" := range ")
		g.genExpr(iterable)
		g.buf.WriteString(" {\n")

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
		return
	}

	// Generate: for _, v := range iterable { lambda(v) }
	elemType := g.inferArrayElementGoType(iterable)

	// Check for tuple destructuring: array of tuples with multiple lambda params
	// e.g., pairs.each -> { |name, value| ... } where pairs is Array<(String, Int)>
	isTupleDestructure := len(lambda.Params) > 1 && strings.HasPrefix(elemType, "struct{")

	if isTupleDestructure {
		// Generate tuple destructuring:
		// for _, _tuple := range iterable {
		//     name := _tuple._0
		//     value := _tuple._1
		//     ...body...
		// }
		g.buf.WriteString("for _, _tuple := range ")
		g.genExpr(iterable)
		g.buf.WriteString(" {\n")

		prevInInlinedLambda := g.inInlinedLambda
		g.inInlinedLambda = true

		// Register all destructured variables
		var scopeVars []struct{ name, varType string }
		for i, param := range lambda.Params {
			scopeVars = append(scopeVars, struct{ name, varType string }{param.Name, "any"})
			g.indent++
			g.writeIndent()
			g.buf.WriteString(param.Name)
			g.buf.WriteString(" := _tuple._")
			g.buf.WriteString(fmt.Sprintf("%d", i))
			g.buf.WriteString("\n")
			g.indent--
		}

		g.scopedVars(scopeVars, func() {
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

// genLambdaMap generates code for arr.map -> { |x| expr }
func (g *Generator) genLambdaMap(iterable ast.Expression, lambda *ast.LambdaExpr) {
	g.needsRuntime = true
	varName := "_v"
	if len(lambda.Params) > 0 {
		varName = lambda.Params[0].Name
	}

	// Infer element type for proper typing
	elemType := g.inferArrayElementGoType(iterable)

	// Store element type in vars BEFORE inferring return type
	// so expressions referencing the parameter can be typed
	prevType, wasDefined := g.vars[varName]
	if varName != "_" {
		g.vars[varName] = elemType
	}

	// Infer return type from the last expression in the lambda body
	returnType := g.inferLambdaReturnType(lambda, elemType)

	g.buf.WriteString("runtime.Map(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" ")
	g.buf.WriteString(elemType)
	g.buf.WriteString(") ")
	g.buf.WriteString(returnType)
	g.buf.WriteString(" {\n")

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

	// Restore previous variable type
	if varName != "_" {
		if wasDefined {
			g.vars[varName] = prevType
		} else {
			delete(g.vars, varName)
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaSelect generates code for arr.select -> { |x| condition } or map.select -> { |k, v| condition }
func (g *Generator) genLambdaSelect(iterable ast.Expression, lambda *ast.LambdaExpr) {
	g.needsRuntime = true

	// Check if iterating over a map with 2 parameters (key, value)
	if len(lambda.Params) == 2 {
		if goType := g.typeInfo.GetGoType(iterable); strings.HasPrefix(goType, "map[") {
			g.genLambdaMapSelect(iterable, lambda, goType)
			return
		}
	}

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

	// Store element type in vars so getReceiverClassName can detect getters
	prevType, wasDefined := g.vars[varName]
	if varName != "_" {
		g.vars[varName] = elemType
	}

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

	// Restore previous variable type
	if varName != "_" {
		if wasDefined {
			g.vars[varName] = prevType
		} else {
			delete(g.vars, varName)
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaMapSelect generates code for map.select -> (k, v) { condition }
func (g *Generator) genLambdaMapSelect(iterable ast.Expression, lambda *ast.LambdaExpr, mapType string) {
	keyName := lambda.Params[0].Name
	valueName := lambda.Params[1].Name

	// Extract key and value types from map[K]V
	keyType, valueType := g.extractMapTypes(mapType)

	// Track the lambda params with their types
	g.vars[keyName] = keyType
	g.vars[valueName] = valueType

	g.buf.WriteString("runtime.MapSelect(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(keyName)
	g.buf.WriteString(" ")
	g.buf.WriteString(keyType)
	g.buf.WriteString(", ")
	g.buf.WriteString(valueName)
	g.buf.WriteString(" ")
	g.buf.WriteString(valueType)
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

// genLambdaReject generates code for arr.reject -> { |x| condition } or map.reject -> { |k, v| condition }
func (g *Generator) genLambdaReject(iterable ast.Expression, lambda *ast.LambdaExpr) {
	g.needsRuntime = true

	// Check if iterating over a map with 2 parameters (key, value)
	if len(lambda.Params) == 2 {
		if goType := g.typeInfo.GetGoType(iterable); strings.HasPrefix(goType, "map[") {
			g.genLambdaMapReject(iterable, lambda, goType)
			return
		}
	}

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

	// Store element type in vars so getReceiverClassName can detect getters
	prevType, wasDefined := g.vars[varName]
	if varName != "_" {
		g.vars[varName] = elemType
	}

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

	// Restore previous variable type
	if varName != "_" {
		if wasDefined {
			g.vars[varName] = prevType
		} else {
			delete(g.vars, varName)
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaMapReject generates code for map.reject -> (k, v) { condition }
func (g *Generator) genLambdaMapReject(iterable ast.Expression, lambda *ast.LambdaExpr, mapType string) {
	keyName := lambda.Params[0].Name
	valueName := lambda.Params[1].Name

	// Extract key and value types from map[K]V
	keyType, valueType := g.extractMapTypes(mapType)

	// Track the lambda params with their types
	g.vars[keyName] = keyType
	g.vars[valueName] = valueType

	g.buf.WriteString("runtime.MapReject(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(keyName)
	g.buf.WriteString(" ")
	g.buf.WriteString(keyType)
	g.buf.WriteString(", ")
	g.buf.WriteString(valueName)
	g.buf.WriteString(" ")
	g.buf.WriteString(valueType)
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

// genMapSelectBlock generates code for map.select do |k, v| ... end
func (g *Generator) genMapSelectBlock(mapExpr ast.Expression, block *ast.BlockExpr, mapType string) {
	g.needsRuntime = true

	keyName := "_k"
	valueName := "_v"
	if len(block.Params) > 0 {
		keyName = block.Params[0]
	}
	if len(block.Params) > 1 {
		valueName = block.Params[1]
	}

	// Extract key and value types from map[K]V
	keyType, valueType := g.extractMapTypes(mapType)

	// Track block params for variable resolution
	keyPrevType, keyWasDefinedBefore := g.vars[keyName]
	valuePrevType, valueWasDefinedBefore := g.vars[valueName]
	g.vars[keyName] = keyType
	g.vars[valueName] = valueType

	g.buf.WriteString("runtime.MapSelect(")
	g.genExpr(mapExpr)
	g.buf.WriteString(", func(")
	g.buf.WriteString(keyName)
	g.buf.WriteString(" ")
	g.buf.WriteString(keyType)
	g.buf.WriteString(", ")
	g.buf.WriteString(valueName)
	g.buf.WriteString(" ")
	g.buf.WriteString(valueType)
	g.buf.WriteString(") bool {\n")

	g.indent++
	for i, stmt := range block.Body {
		if i == len(block.Body)-1 {
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
	if len(block.Body) == 0 {
		g.writeIndent()
		g.buf.WriteString("return false\n")
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("})")

	// Restore previous var state
	if keyWasDefinedBefore {
		g.vars[keyName] = keyPrevType
	} else {
		delete(g.vars, keyName)
	}
	if valueWasDefinedBefore {
		g.vars[valueName] = valuePrevType
	} else {
		delete(g.vars, valueName)
	}
}

// genMapRejectBlock generates code for map.reject do |k, v| ... end
func (g *Generator) genMapRejectBlock(mapExpr ast.Expression, block *ast.BlockExpr, mapType string) {
	g.needsRuntime = true

	keyName := "_k"
	valueName := "_v"
	if len(block.Params) > 0 {
		keyName = block.Params[0]
	}
	if len(block.Params) > 1 {
		valueName = block.Params[1]
	}

	// Extract key and value types from map[K]V
	keyType, valueType := g.extractMapTypes(mapType)

	// Track block params for variable resolution
	keyPrevType, keyWasDefinedBefore := g.vars[keyName]
	valuePrevType, valueWasDefinedBefore := g.vars[valueName]
	g.vars[keyName] = keyType
	g.vars[valueName] = valueType

	g.buf.WriteString("runtime.MapReject(")
	g.genExpr(mapExpr)
	g.buf.WriteString(", func(")
	g.buf.WriteString(keyName)
	g.buf.WriteString(" ")
	g.buf.WriteString(keyType)
	g.buf.WriteString(", ")
	g.buf.WriteString(valueName)
	g.buf.WriteString(" ")
	g.buf.WriteString(valueType)
	g.buf.WriteString(") bool {\n")

	g.indent++
	for i, stmt := range block.Body {
		if i == len(block.Body)-1 {
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
	if len(block.Body) == 0 {
		g.writeIndent()
		g.buf.WriteString("return false\n")
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("})")

	// Restore previous var state
	if keyWasDefinedBefore {
		g.vars[keyName] = keyPrevType
	} else {
		delete(g.vars, keyName)
	}
	if valueWasDefinedBefore {
		g.vars[valueName] = valuePrevType
	} else {
		delete(g.vars, valueName)
	}
}

// genLambdaFind generates code for arr.find -> { |x| condition }
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

	// Store element type in vars so getReceiverClassName can detect getters
	prevType, wasDefined := g.vars[varName]
	if varName != "_" {
		g.vars[varName] = elemType
	}

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

	// Restore previous variable type
	if varName != "_" {
		if wasDefined {
			g.vars[varName] = prevType
		} else {
			delete(g.vars, varName)
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaAny generates code for arr.any? -> { |x| condition }
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

	// Store element type in vars so getReceiverClassName can detect getters
	prevType, wasDefined := g.vars[varName]
	if varName != "_" {
		g.vars[varName] = elemType
	}

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

	// Restore previous variable type
	if varName != "_" {
		if wasDefined {
			g.vars[varName] = prevType
		} else {
			delete(g.vars, varName)
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaAll generates code for arr.all? -> { |x| condition }
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

	// Store element type in vars so getReceiverClassName can detect getters
	prevType, wasDefined := g.vars[varName]
	if varName != "_" {
		g.vars[varName] = elemType
	}

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

	// Restore previous variable type
	if varName != "_" {
		if wasDefined {
			g.vars[varName] = prevType
		} else {
			delete(g.vars, varName)
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaNone generates code for arr.none? -> { |x| condition }
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

	// Store element type in vars so getReceiverClassName can detect getters
	prevType, wasDefined := g.vars[varName]
	if varName != "_" {
		g.vars[varName] = elemType
	}

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

	// Restore previous variable type
	if varName != "_" {
		if wasDefined {
			g.vars[varName] = prevType
		} else {
			delete(g.vars, varName)
		}
	}

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

	// Store variable types in vars so getReceiverClassName can detect getters
	accPrevType, accWasDefined := g.vars[accName]
	varPrevType, varWasDefined := g.vars[varName]
	if accName != "_" {
		g.vars[accName] = accType
	}
	if varName != "_" {
		g.vars[varName] = elemType
	}

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

	// Restore previous variable types
	if accName != "_" {
		if accWasDefined {
			g.vars[accName] = accPrevType
		} else {
			delete(g.vars, accName)
		}
	}
	if varName != "_" {
		if varWasDefined {
			g.vars[varName] = varPrevType
		} else {
			delete(g.vars, varName)
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genLambdaTimes generates code for n.times -> { |i| body }
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

// genLambdaUpto generates code for start.upto(end) -> { |i| body }
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

// genLambdaDownto generates code for start.downto(end) -> { |i| body }
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
	g.needsRuntime = true

	// Infer return type from block body for generic ScopeSpawn[T]
	returnType := g.inferBlockBodyReturnType(block.Body)

	// Generate: runtime.ScopeSpawn(scope, func() T { ... })
	g.buf.WriteString("runtime.ScopeSpawn(")
	g.genExpr(scope)
	g.buf.WriteString(", func() ")
	g.buf.WriteString(returnType)
	g.buf.WriteString(" {\n")

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

// genScopedSpawnLambda generates code for scope.spawn(lambda) -> runtime.ScopeSpawn(scope, lambda)
func (g *Generator) genScopedSpawnLambda(scope ast.Expression, lambda *ast.LambdaExpr) {
	g.needsRuntime = true

	// Infer return type from lambda body for generic ScopeSpawn[T]
	returnType := g.inferBlockBodyReturnType(lambda.Body)

	// Generate: runtime.ScopeSpawn(scope, func() T { ... })
	g.buf.WriteString("runtime.ScopeSpawn(")
	g.genExpr(scope)
	g.buf.WriteString(", func() ")
	g.buf.WriteString(returnType)
	g.buf.WriteString(" {\n")

	g.indent++
	if len(lambda.Body) > 0 {
		for i, stmt := range lambda.Body {
			if i == len(lambda.Body)-1 {
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

func (g *Generator) genSymbolToProcBlockCall(iterable ast.Expression, stp *ast.SymbolToProcExpr, method blockMethod) {
	g.needsRuntime = true
	elemType := g.inferArrayElementGoType(iterable)

	// Extract base class name from element type (e.g., "*User" -> "User")
	baseClass := strings.TrimPrefix(elemType, "*")

	// Check if this is a user-defined class with the method
	goMethodName := ""
	if g.instanceMethods[baseClass] != nil {
		goMethodName = g.instanceMethods[baseClass][stp.Method]
	}

	// Check for stdLib method on primitives
	stdLibType := goTypeToStdLibType(elemType)
	var stdLibMethod *MethodDef
	if stdLibType != "" {
		if methods, ok := stdLib[stdLibType]; ok {
			if info, ok := methods[stp.Method]; ok {
				stdLibMethod = &info
			}
		}
	}

	// Determine return type based on the method being called
	returnType := method.returnType
	if stdLibMethod != nil && stdLibMethod.ReturnType != "" {
		returnType = mapType(stdLibMethod.ReturnType)
	}

	g.buf.WriteString(method.runtimeFunc)
	g.buf.WriteString("(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(x ")
	g.buf.WriteString(elemType)
	g.buf.WriteString(") ")
	g.buf.WriteString(returnType)

	if goMethodName != "" {
		// User-defined class with known method - generate direct call
		g.buf.WriteString(" { return x.")
		g.buf.WriteString(goMethodName)
		g.buf.WriteString("() })")
	} else if stdLibMethod != nil {
		// Primitive type with known runtime function - generate direct call
		g.buf.WriteString(" { return ")
		g.buf.WriteString(stdLibMethod.RuntimeFunc)
		g.buf.WriteString("(x) })")
	} else {
		// Unknown method - use runtime.CallMethod as fallback
		g.buf.WriteString(" { return runtime.CallMethod(x, \"")
		g.buf.WriteString(stp.Method)
		g.buf.WriteString("\")")
		// Add type assertion for non-any return types
		if returnType != "any" {
			g.buf.WriteString(".(")
			g.buf.WriteString(returnType)
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" })")
	}
}

// goTypeToStdLibType maps Go type names to stdLib type keys.
func goTypeToStdLibType(goType string) string {
	switch goType {
	case "string":
		return "String"
	case "int":
		return "Int"
	case "int64":
		return "Int"
	case "float64":
		return "Float"
	case "bool":
		return "Bool"
	default:
		return ""
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

	// Store element types in vars so getReceiverClassName can detect getters
	if method.hasAccumulator {
		// For reduce: acc has accType, element has elemType
		if param1Name != "_" {
			accType := "any"
			if len(args) > 0 {
				accType = mapType(g.inferTypeFromExpr(args[0]))
			}
			g.vars[param1Name] = accType
		}
		if param2Name != "_" {
			g.vars[param2Name] = elemType
		}
	} else {
		// For map/select/etc: element has elemType
		if param1Name != "_" {
			g.vars[param1Name] = elemType
		}
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
