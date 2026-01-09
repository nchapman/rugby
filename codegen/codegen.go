package codegen

import (
	"fmt"
	"go/format"
	"strings"
	"unicode"

	"rugby/ast"
)

type Generator struct {
	buf          strings.Builder
	indent       int
	vars         map[string]bool // track declared variables
	needsRuntime bool            // track if rugby/runtime import is needed
}

func New() *Generator {
	return &Generator{vars: make(map[string]bool)}
}

func (g *Generator) Generate(program *ast.Program) (string, error) {
	// First pass: generate declarations to determine what imports we need
	var bodyBuf strings.Builder
	g.buf = bodyBuf
	for _, decl := range program.Declarations {
		g.genStatement(decl)
		g.buf.WriteString("\n")
	}
	bodyCode := g.buf.String()

	// Second pass: build final output with package and imports
	var out strings.Builder
	out.WriteString("package main\n\n")

	// Track user imports to avoid duplicates
	userImports := make(map[string]bool)
	for _, imp := range program.Imports {
		userImports[imp.Path] = true
	}

	// Collect all imports
	needsRuntimeImport := g.needsRuntime && !userImports["rugby/runtime"]
	hasImports := len(program.Imports) > 0 || needsRuntimeImport
	if hasImports {
		out.WriteString("import (\n")
		// User-specified imports
		for _, imp := range program.Imports {
			if imp.Alias != "" {
				out.WriteString(fmt.Sprintf("\t%s %q\n", imp.Alias, imp.Path))
			} else {
				out.WriteString(fmt.Sprintf("\t%q\n", imp.Path))
			}
		}
		// Auto-imports (only if not already imported by user)
		if needsRuntimeImport {
			out.WriteString("\t\"rugby/runtime\"\n")
		}
		out.WriteString(")\n\n")
	}

	out.WriteString(bodyCode)

	// Format the output
	source := out.String()
	formatted, err := format.Source([]byte(source))
	if err != nil {
		return source, nil // Return unformatted if gofmt fails
	}
	return string(formatted), nil
}

func (g *Generator) writeIndent() {
	for i := 0; i < g.indent; i++ {
		g.buf.WriteString("\t")
	}
}

func (g *Generator) genStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.FuncDecl:
		g.genFuncDecl(s)
	case *ast.ExprStmt:
		g.writeIndent()
		g.genExpr(s.Expr)
		g.buf.WriteString("\n")
	case *ast.AssignStmt:
		g.genAssignStmt(s)
	case *ast.IfStmt:
		g.genIfStmt(s)
	case *ast.WhileStmt:
		g.genWhileStmt(s)
	case *ast.ReturnStmt:
		g.genReturnStmt(s)
	case *ast.DeferStmt:
		g.genDeferStmt(s)
	}
}

func (g *Generator) genFuncDecl(fn *ast.FuncDecl) {
	g.vars = make(map[string]bool) // reset vars for each function

	// Mark parameters as declared variables
	for _, param := range fn.Params {
		g.vars[param.Name] = true
	}

	// Generate function signature
	g.buf.WriteString(fmt.Sprintf("func %s(", fn.Name))
	for i, param := range fn.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.buf.WriteString(fmt.Sprintf("%s interface{}", param.Name))
	}
	g.buf.WriteString(")")

	// Generate return type(s) if specified
	if len(fn.ReturnTypes) == 1 {
		g.buf.WriteString(" ")
		g.buf.WriteString(mapType(fn.ReturnTypes[0]))
	} else if len(fn.ReturnTypes) > 1 {
		g.buf.WriteString(" (")
		for i, rt := range fn.ReturnTypes {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.buf.WriteString(mapType(rt))
		}
		g.buf.WriteString(")")
	}

	g.buf.WriteString(" {\n")

	g.indent++
	for _, stmt := range fn.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.buf.WriteString("}\n")
}

func (g *Generator) genAssignStmt(s *ast.AssignStmt) {
	g.writeIndent()
	g.buf.WriteString(s.Name)
	if g.vars[s.Name] {
		g.buf.WriteString(" = ")
	} else {
		g.buf.WriteString(" := ")
		g.vars[s.Name] = true
	}
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

func (g *Generator) genIfStmt(s *ast.IfStmt) {
	g.writeIndent()
	g.buf.WriteString("if ")
	g.genExpr(s.Cond)
	g.buf.WriteString(" {\n")

	g.indent++
	for _, stmt := range s.Then {
		g.genStatement(stmt)
	}
	g.indent--

	for _, elsif := range s.ElseIfs {
		g.writeIndent()
		g.buf.WriteString("} else if ")
		g.genExpr(elsif.Cond)
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

func (g *Generator) genWhileStmt(s *ast.WhileStmt) {
	g.writeIndent()
	g.buf.WriteString("for ")
	g.genExpr(s.Cond)
	g.buf.WriteString(" {\n")

	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genReturnStmt(s *ast.ReturnStmt) {
	g.writeIndent()
	g.buf.WriteString("return")
	if len(s.Values) > 0 {
		g.buf.WriteString(" ")
		for i, val := range s.Values {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.genExpr(val)
		}
	}
	g.buf.WriteString("\n")
}

func (g *Generator) genExpr(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		g.genCallExpr(e)
	case *ast.SelectorExpr:
		g.genSelectorExpr(e)
	case *ast.StringLit:
		g.buf.WriteString(fmt.Sprintf("%q", e.Value))
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
	case *ast.ArrayLit:
		g.genArrayLit(e)
	case *ast.IndexExpr:
		g.genIndexExpr(e)
	case *ast.MapLit:
		g.genMapLit(e)
	case *ast.Ident:
		// Check for kernel functions that can be used without parens
		switch e.Name {
		case "gets":
			g.needsRuntime = true
			g.buf.WriteString("runtime.Gets()")
		case "rand":
			g.needsRuntime = true
			g.buf.WriteString("runtime.RandFloat()")
		default:
			g.buf.WriteString(e.Name)
		}
	case *ast.BinaryExpr:
		g.genBinaryExpr(e)
	case *ast.UnaryExpr:
		g.genUnaryExpr(e)
	}
}

func (g *Generator) genBinaryExpr(e *ast.BinaryExpr) {
	g.buf.WriteString("(")
	g.genExpr(e.Left)
	g.buf.WriteString(" ")
	g.buf.WriteString(e.Op)
	g.buf.WriteString(" ")
	g.genExpr(e.Right)
	g.buf.WriteString(")")
}

func (g *Generator) genUnaryExpr(e *ast.UnaryExpr) {
	g.buf.WriteString(e.Op)
	g.genExpr(e.Expr)
}

func (g *Generator) genArrayLit(arr *ast.ArrayLit) {
	// Emit untyped array; type inference will be added in Phase 7
	g.buf.WriteString("[]interface{}{")
	for i, elem := range arr.Elements {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.genExpr(elem)
	}
	g.buf.WriteString("}")
}

func (g *Generator) genIndexExpr(idx *ast.IndexExpr) {
	g.genExpr(idx.Left)
	g.buf.WriteString("[")
	g.genExpr(idx.Index)
	g.buf.WriteString("]")
}

func (g *Generator) genMapLit(m *ast.MapLit) {
	// Emit untyped map; type inference will be added in Phase 7
	g.buf.WriteString("map[interface{}]interface{}{")
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

func (g *Generator) genCallExpr(call *ast.CallExpr) {
	// If the call has a block, handle it specially
	if call.Block != nil {
		g.genBlockCall(call)
		return
	}

	// Generate the function expression
	switch fn := call.Func.(type) {
	case *ast.Ident:
		// Simple function call - check for kernel/builtin functions
		funcName := fn.Name
		switch funcName {
		case "puts":
			g.needsRuntime = true
			funcName = "runtime.Puts"
		case "print":
			g.needsRuntime = true
			funcName = "runtime.Print"
		case "p":
			g.needsRuntime = true
			funcName = "runtime.P"
		case "gets":
			g.needsRuntime = true
			funcName = "runtime.Gets"
		case "exit":
			g.needsRuntime = true
			funcName = "runtime.Exit"
		case "sleep":
			g.needsRuntime = true
			funcName = "runtime.Sleep"
		case "rand":
			g.needsRuntime = true
			// rand with args -> RandInt, rand without -> RandFloat
			if len(call.Args) > 0 {
				funcName = "runtime.RandInt"
			} else {
				funcName = "runtime.RandFloat"
			}
		}
		// Don't transform local function names - they stay as defined
		g.buf.WriteString(funcName)
	case *ast.SelectorExpr:
		// Method/package call like http.Get or resp.Body.Close
		// snake_case transformation only applies here (Go interop)
		g.genSelectorExpr(fn)
	default:
		// Fallback for other expressions
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

// genBlockCall generates code for method calls with blocks.
// Recognizes patterns like each, each_with_index, map and generates appropriate Go code.
func (g *Generator) genBlockCall(call *ast.CallExpr) {
	sel, ok := call.Func.(*ast.SelectorExpr)
	if !ok {
		// Block on non-selector call - future: support generic block-taking functions
		g.buf.WriteString("/* unsupported block call */")
		return
	}

	method := sel.Sel
	block := call.Block

	switch method {
	case "each":
		g.genEachBlock(sel.X, block)
	case "each_with_index":
		g.genEachWithIndexBlock(sel.X, block)
	case "map":
		g.genMapBlock(sel.X, block)
	case "select", "filter":
		g.genSelectBlock(sel.X, block)
	case "reject":
		g.genRejectBlock(sel.X, block)
	case "reduce":
		g.genReduceBlock(sel.X, block, call.Args)
	case "find", "detect":
		g.genFindBlock(sel.X, block)
	default:
		// Unknown method with block - future: support user-defined block methods
		g.buf.WriteString(fmt.Sprintf("/* unsupported block method: %s */", method))
	}
}

// genEachBlock generates: for _, v := range iterable { body }
func (g *Generator) genEachBlock(iterable ast.Expression, block *ast.BlockExpr) {
	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	// Save previous variable state for proper scope restoration
	wasDefinedBefore := g.vars[varName]

	g.buf.WriteString("for _, ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := range ")
	g.genExpr(iterable)
	g.buf.WriteString(" {\n")

	if varName != "_" {
		g.vars[varName] = true
	}

	g.indent++
	for _, stmt := range block.Body {
		g.genStatement(stmt)
	}
	g.indent--

	// Restore variable state after block
	if !wasDefinedBefore && varName != "_" {
		delete(g.vars, varName)
	}

	g.writeIndent()
	g.buf.WriteString("}")
}

// genEachWithIndexBlock generates: for i, v := range iterable { body }
// Note: Rugby uses |value, index| order but Go uses index, value
func (g *Generator) genEachWithIndexBlock(iterable ast.Expression, block *ast.BlockExpr) {
	varName := "_"
	indexName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}
	if len(block.Params) > 1 {
		indexName = block.Params[1]
	}

	// Save previous variable state for proper scope restoration
	varWasDefinedBefore := g.vars[varName]
	indexWasDefinedBefore := g.vars[indexName]

	// Swap order: Rugby |value, index| -> Go index, value
	g.buf.WriteString("for ")
	g.buf.WriteString(indexName)
	g.buf.WriteString(", ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := range ")
	g.genExpr(iterable)
	g.buf.WriteString(" {\n")

	if varName != "_" {
		g.vars[varName] = true
	}
	if indexName != "_" {
		g.vars[indexName] = true
	}

	g.indent++
	for _, stmt := range block.Body {
		g.genStatement(stmt)
	}
	g.indent--

	// Restore variable state after block
	if !varWasDefinedBefore && varName != "_" {
		delete(g.vars, varName)
	}
	if !indexWasDefinedBefore && indexName != "_" {
		delete(g.vars, indexName)
	}

	g.writeIndent()
	g.buf.WriteString("}")
}

// genMapBlock generates: runtime.Map(iterable, func(x T) R { return expr })
func (g *Generator) genMapBlock(iterable ast.Expression, block *ast.BlockExpr) {
	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	// Save previous variable state for proper scope restoration
	wasDefinedBefore := g.vars[varName]

	g.buf.WriteString("runtime.Map(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" interface{}) interface{} {\n")

	if varName != "_" {
		g.vars[varName] = true
	}

	g.indent++

	// The last expression in the block is the return value
	if len(block.Body) > 0 {
		// Generate all but last statement normally
		for i := 0; i < len(block.Body)-1; i++ {
			g.genStatement(block.Body[i])
		}
		// Last statement should be an expression - return it
		lastStmt := block.Body[len(block.Body)-1]
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			g.writeIndent()
			g.buf.WriteString("return ")
			g.genExpr(exprStmt.Expr)
			g.buf.WriteString("\n")
		} else {
			// Not an expression, just generate it
			g.genStatement(lastStmt)
		}
	}

	g.indent--

	// Restore variable state after block
	if !wasDefinedBefore && varName != "_" {
		delete(g.vars, varName)
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genSelectBlock generates: runtime.Select(iterable, func(x T) bool { return expr })
func (g *Generator) genSelectBlock(iterable ast.Expression, block *ast.BlockExpr) {
	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	wasDefinedBefore := g.vars[varName]

	g.buf.WriteString("runtime.Select(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" interface{}) bool {\n")

	if varName != "_" {
		g.vars[varName] = true
	}

	g.indent++

	// Last expression is the predicate
	if len(block.Body) > 0 {
		for i := 0; i < len(block.Body)-1; i++ {
			g.genStatement(block.Body[i])
		}
		lastStmt := block.Body[len(block.Body)-1]
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			g.writeIndent()
			g.buf.WriteString("return ")
			g.genExpr(exprStmt.Expr)
			g.buf.WriteString("\n")
		} else {
			g.genStatement(lastStmt)
		}
	}

	g.indent--

	if !wasDefinedBefore && varName != "_" {
		delete(g.vars, varName)
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genRejectBlock generates: runtime.Reject(iterable, func(x T) bool { return expr })
func (g *Generator) genRejectBlock(iterable ast.Expression, block *ast.BlockExpr) {
	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	wasDefinedBefore := g.vars[varName]

	g.buf.WriteString("runtime.Reject(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" interface{}) bool {\n")

	if varName != "_" {
		g.vars[varName] = true
	}

	g.indent++

	if len(block.Body) > 0 {
		for i := 0; i < len(block.Body)-1; i++ {
			g.genStatement(block.Body[i])
		}
		lastStmt := block.Body[len(block.Body)-1]
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			g.writeIndent()
			g.buf.WriteString("return ")
			g.genExpr(exprStmt.Expr)
			g.buf.WriteString("\n")
		} else {
			g.genStatement(lastStmt)
		}
	}

	g.indent--

	if !wasDefinedBefore && varName != "_" {
		delete(g.vars, varName)
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genReduceBlock generates: runtime.Reduce(iterable, initial, func(acc R, x T) R { return expr })
func (g *Generator) genReduceBlock(iterable ast.Expression, block *ast.BlockExpr, args []ast.Expression) {
	g.needsRuntime = true

	accName := "acc"
	varName := "_"
	if len(block.Params) > 0 {
		accName = block.Params[0]
	}
	if len(block.Params) > 1 {
		varName = block.Params[1]
	}

	accWasDefinedBefore := g.vars[accName]
	varWasDefinedBefore := g.vars[varName]

	g.buf.WriteString("runtime.Reduce(")
	g.genExpr(iterable)
	g.buf.WriteString(", ")

	// Initial value (first arg to reduce)
	if len(args) > 0 {
		g.genExpr(args[0])
	} else {
		g.buf.WriteString("nil")
	}

	g.buf.WriteString(", func(")
	g.buf.WriteString(accName)
	g.buf.WriteString(" interface{}, ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" interface{}) interface{} {\n")

	if accName != "_" {
		g.vars[accName] = true
	}
	if varName != "_" {
		g.vars[varName] = true
	}

	g.indent++

	if len(block.Body) > 0 {
		for i := 0; i < len(block.Body)-1; i++ {
			g.genStatement(block.Body[i])
		}
		lastStmt := block.Body[len(block.Body)-1]
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			g.writeIndent()
			g.buf.WriteString("return ")
			g.genExpr(exprStmt.Expr)
			g.buf.WriteString("\n")
		} else {
			g.genStatement(lastStmt)
		}
	}

	g.indent--

	if !accWasDefinedBefore && accName != "_" {
		delete(g.vars, accName)
	}
	if !varWasDefinedBefore && varName != "_" {
		delete(g.vars, varName)
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genFindBlock generates: runtime.Find(iterable, func(x T) bool { return expr })
func (g *Generator) genFindBlock(iterable ast.Expression, block *ast.BlockExpr) {
	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	wasDefinedBefore := g.vars[varName]

	g.buf.WriteString("runtime.Find(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" interface{}) bool {\n")

	if varName != "_" {
		g.vars[varName] = true
	}

	g.indent++

	if len(block.Body) > 0 {
		for i := 0; i < len(block.Body)-1; i++ {
			g.genStatement(block.Body[i])
		}
		lastStmt := block.Body[len(block.Body)-1]
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			g.writeIndent()
			g.buf.WriteString("return ")
			g.genExpr(exprStmt.Expr)
			g.buf.WriteString("\n")
		} else {
			g.genStatement(lastStmt)
		}
	}

	g.indent--

	if !wasDefinedBefore && varName != "_" {
		delete(g.vars, varName)
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

func (g *Generator) genSelectorExpr(sel *ast.SelectorExpr) {
	g.genExpr(sel.X)
	g.buf.WriteString(".")
	// Apply snake_case to CamelCase mapping
	g.buf.WriteString(snakeToCamel(sel.Sel))
}

func (g *Generator) genDeferStmt(s *ast.DeferStmt) {
	g.writeIndent()
	g.buf.WriteString("defer ")
	g.genCallExpr(s.Call)
	g.buf.WriteString("\n")
}

// snakeToCamel converts snake_case to CamelCase for Go interop
// Examples: read_all -> ReadAll, new_request -> NewRequest
func snakeToCamel(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	capitalizeNext := true

	for _, r := range s {
		if r == '_' {
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// mapType converts Rugby type names to Go type names
func mapType(rubyType string) string {
	switch rubyType {
	case "Int":
		return "int"
	case "Int64":
		return "int64"
	case "Float":
		return "float64"
	case "String":
		return "string"
	case "Bool":
		return "bool"
	case "Bytes":
		return "[]byte"
	default:
		return rubyType // pass through unknown types (e.g., user-defined)
	}
}
