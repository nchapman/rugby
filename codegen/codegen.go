package codegen

import (
	"fmt"
	"go/format"
	"strings"
	"unicode"

	"rugby/ast"
)

// kernelFunc describes a Rugby kernel function mapping to runtime
type kernelFunc struct {
	runtimeFunc string
	// transform customizes function name based on args (e.g., rand -> RandInt/RandFloat)
	transform func(args []ast.Expression) string
}

// Kernel function mappings - single source of truth
var kernelFuncs = map[string]kernelFunc{
	"puts":  {runtimeFunc: "runtime.Puts"},
	"print": {runtimeFunc: "runtime.Print"},
	"p":     {runtimeFunc: "runtime.P"},
	"gets":  {runtimeFunc: "runtime.Gets"},
	"exit":  {runtimeFunc: "runtime.Exit"},
	"sleep": {runtimeFunc: "runtime.Sleep"},
	"rand": {
		transform: func(args []ast.Expression) string {
			if len(args) > 0 {
				return "runtime.RandInt"
			}
			return "runtime.RandFloat"
		},
	},
}

// Kernel functions that can be used as identifiers (without parens)
var noParenKernelFuncs = map[string]string{
	"gets": "runtime.Gets()",
	"rand": "runtime.RandFloat()",
}

// blockMethod describes a block method mapping
type blockMethod struct {
	runtimeFunc    string
	returnType     string // "interface{}" or "bool"
	hasAccumulator bool   // for reduce - takes 2 params (acc, elem) instead of 1
}

// Block method mappings - single source of truth
var blockMethods = map[string]blockMethod{
	"map":    {runtimeFunc: "runtime.Map", returnType: "interface{}"},
	"select": {runtimeFunc: "runtime.Select", returnType: "bool"},
	"filter": {runtimeFunc: "runtime.Select", returnType: "bool"},
	"reject": {runtimeFunc: "runtime.Reject", returnType: "bool"},
	"find":   {runtimeFunc: "runtime.Find", returnType: "bool"},
	"detect": {runtimeFunc: "runtime.Find", returnType: "bool"},
	"reduce": {runtimeFunc: "runtime.Reduce", returnType: "interface{}", hasAccumulator: true},
	"any?":   {runtimeFunc: "runtime.Any", returnType: "bool"},
	"all?":   {runtimeFunc: "runtime.All", returnType: "bool"},
	"none?":  {runtimeFunc: "runtime.None", returnType: "bool"},
}

type Generator struct {
	buf          strings.Builder
	indent       int
	vars         map[string]bool // track declared variables
	needsRuntime bool            // track if rugby/runtime import is needed
	currentClass string          // current class being generated (for instance vars)
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
	case *ast.ClassDecl:
		g.genClassDecl(s)
	case *ast.ExprStmt:
		g.writeIndent()
		g.genExpr(s.Expr)
		g.buf.WriteString("\n")
	case *ast.AssignStmt:
		g.genAssignStmt(s)
	case *ast.InstanceVarAssign:
		g.genInstanceVarAssign(s)
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
		if param.Type != "" {
			g.buf.WriteString(fmt.Sprintf("%s %s", param.Name, mapType(param.Type)))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s interface{}", param.Name))
		}
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

func (g *Generator) genClassDecl(cls *ast.ClassDecl) {
	className := cls.Name
	g.currentClass = className

	// Emit struct definition
	hasContent := cls.Parent != "" || len(cls.Fields) > 0
	if hasContent {
		g.buf.WriteString(fmt.Sprintf("type %s struct {\n", className))
		if cls.Parent != "" {
			g.buf.WriteString("\t")
			g.buf.WriteString(cls.Parent)
			g.buf.WriteString("\n")
		}
		for _, field := range cls.Fields {
			if field.Type != "" {
				g.buf.WriteString(fmt.Sprintf("\t%s %s\n", field.Name, mapType(field.Type)))
			} else {
				g.buf.WriteString(fmt.Sprintf("\t%s interface{}\n", field.Name))
			}
		}
		g.buf.WriteString("}\n\n")
	} else {
		g.buf.WriteString(fmt.Sprintf("type %s struct{}\n\n", className))
	}

	// Emit constructor if initialize exists
	for _, method := range cls.Methods {
		if method.Name == "initialize" {
			g.genConstructor(className, method)
			break
		}
	}

	// Emit methods (skip initialize - it's the constructor)
	for _, method := range cls.Methods {
		if method.Name != "initialize" {
			g.genMethodDecl(className, method)
		}
	}

	g.currentClass = ""
}

func (g *Generator) genConstructor(className string, method *ast.MethodDecl) {
	g.vars = make(map[string]bool)

	// Mark parameters as declared variables
	for _, param := range method.Params {
		g.vars[param.Name] = true
	}

	// Receiver name for field assignments
	recv := receiverName(className)

	// Constructor: func NewClassName(params) *ClassName
	g.buf.WriteString(fmt.Sprintf("func New%s(", className))
	for i, param := range method.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		if param.Type != "" {
			g.buf.WriteString(fmt.Sprintf("%s %s", param.Name, mapType(param.Type)))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s interface{}", param.Name))
		}
	}
	g.buf.WriteString(fmt.Sprintf(") *%s {\n", className))

	// Create instance
	g.buf.WriteString(fmt.Sprintf("\t%s := &%s{}\n", recv, className))

	// Generate body (instance var assignments become field sets)
	g.indent++
	for _, stmt := range method.Body {
		g.genStatement(stmt)
	}
	g.indent--

	// Return instance
	g.buf.WriteString(fmt.Sprintf("\treturn %s\n", recv))
	g.buf.WriteString("}\n\n")
}

func (g *Generator) genMethodDecl(className string, method *ast.MethodDecl) {
	g.vars = make(map[string]bool) // reset vars for each method

	// Mark parameters as declared variables
	for _, param := range method.Params {
		g.vars[param.Name] = true
	}

	// Receiver name: first letter of class, lowercase
	recv := receiverName(className)

	// Method name: convert snake_case to CamelCase
	methodName := snakeToCamel(method.Name)

	// Generate method signature with pointer receiver: func (r *ClassName) MethodName(params) returns
	g.buf.WriteString(fmt.Sprintf("func (%s *%s) %s(", recv, className, methodName))
	for i, param := range method.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		if param.Type != "" {
			g.buf.WriteString(fmt.Sprintf("%s %s", param.Name, mapType(param.Type)))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s interface{}", param.Name))
		}
	}
	g.buf.WriteString(")")

	// Generate return type(s) if specified
	if len(method.ReturnTypes) == 1 {
		g.buf.WriteString(" ")
		g.buf.WriteString(mapType(method.ReturnTypes[0]))
	} else if len(method.ReturnTypes) > 1 {
		g.buf.WriteString(" (")
		for i, rt := range method.ReturnTypes {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.buf.WriteString(mapType(rt))
		}
		g.buf.WriteString(")")
	}

	g.buf.WriteString(" {\n")

	g.indent++
	for _, stmt := range method.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.buf.WriteString("}\n\n")
}

func (g *Generator) genInstanceVarAssign(s *ast.InstanceVarAssign) {
	g.writeIndent()
	if g.currentClass != "" {
		recv := receiverName(g.currentClass)
		g.buf.WriteString(fmt.Sprintf("%s.%s = ", recv, s.Name))
	} else {
		g.buf.WriteString(fmt.Sprintf("/* @%s outside class */ ", s.Name))
	}
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

func (g *Generator) genAssignStmt(s *ast.AssignStmt) {
	g.writeIndent()
	if s.Type != "" && !g.vars[s.Name] {
		// Typed declaration: var x int = value
		g.buf.WriteString(fmt.Sprintf("var %s %s = ", s.Name, mapType(s.Type)))
		g.vars[s.Name] = true
	} else if g.vars[s.Name] {
		// Reassignment: x = value
		g.buf.WriteString(s.Name)
		g.buf.WriteString(" = ")
	} else {
		// Untyped declaration: x := value
		g.buf.WriteString(s.Name)
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
		if runtimeCall, ok := noParenKernelFuncs[e.Name]; ok {
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
		if kf, ok := kernelFuncs[funcName]; ok {
			g.needsRuntime = true
			if kf.transform != nil {
				funcName = kf.transform(call.Args)
			} else {
				funcName = kf.runtimeFunc
			}
		}
		// Don't transform local function names - they stay as defined
		g.buf.WriteString(funcName)
	case *ast.SelectorExpr:
		// Check for ClassName.new() constructor call
		if fn.Sel == "new" {
			if ident, ok := fn.X.(*ast.Ident); ok {
				// Rewrite User.new(args) to NewUser(args)
				g.buf.WriteString(fmt.Sprintf("New%s", ident.Name))
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

	// Special cases: inline iteration (each, each_with_index, times, upto, downto)
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
	}

	// Check if this is a runtime block method
	if bm, ok := blockMethods[method]; ok {
		g.genRuntimeBlock(sel.X, block, call.Args, bm)
		return
	}

	// Unknown method with block
	g.buf.WriteString(fmt.Sprintf("/* unsupported block method: %s */", method))
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

// genTimesBlock generates: for i := 0; i < n; i++ { body }
func (g *Generator) genTimesBlock(times ast.Expression, block *ast.BlockExpr) {
	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	wasDefinedBefore := g.vars[varName]

	// Use synthetic variable if no block param (can't use _ in for _ := 0; _ < n; _++)
	loopVar := varName
	if loopVar == "_" {
		loopVar = "_i"
	}

	g.buf.WriteString("for ")
	g.buf.WriteString(loopVar)
	g.buf.WriteString(" := 0; ")
	g.buf.WriteString(loopVar)
	g.buf.WriteString(" < ")
	g.genExpr(times)
	g.buf.WriteString("; ")
	g.buf.WriteString(loopVar)
	g.buf.WriteString("++ {\n")

	if varName != "_" {
		g.vars[varName] = true
	}

	g.indent++
	for _, stmt := range block.Body {
		g.genStatement(stmt)
	}
	g.indent--

	if !wasDefinedBefore && varName != "_" {
		delete(g.vars, varName)
	}

	g.writeIndent()
	g.buf.WriteString("}")
}

// genUptoBlock generates: for i := start; i <= end; i++ { body }
func (g *Generator) genUptoBlock(start ast.Expression, block *ast.BlockExpr, args []ast.Expression) {
	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	wasDefinedBefore := g.vars[varName]

	// Use synthetic variable if no block param (can't use _ in for _ := start; _ <= end; _++)
	loopVar := varName
	if loopVar == "_" {
		loopVar = "_i"
	}

	g.buf.WriteString("for ")
	g.buf.WriteString(loopVar)
	g.buf.WriteString(" := ")
	g.genExpr(start)
	g.buf.WriteString("; ")
	g.buf.WriteString(loopVar)
	g.buf.WriteString(" <= ")
	if len(args) > 0 {
		g.genExpr(args[0])
	} else {
		g.buf.WriteString("0")
	}
	g.buf.WriteString("; ")
	g.buf.WriteString(loopVar)
	g.buf.WriteString("++ {\n")

	if varName != "_" {
		g.vars[varName] = true
	}

	g.indent++
	for _, stmt := range block.Body {
		g.genStatement(stmt)
	}
	g.indent--

	if !wasDefinedBefore && varName != "_" {
		delete(g.vars, varName)
	}

	g.writeIndent()
	g.buf.WriteString("}")
}

// genDowntoBlock generates: for i := start; i >= end; i-- { body }
func (g *Generator) genDowntoBlock(start ast.Expression, block *ast.BlockExpr, args []ast.Expression) {
	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	wasDefinedBefore := g.vars[varName]

	// Use synthetic variable if no block param (can't use _ in for _ := start; _ >= end; _--)
	loopVar := varName
	if loopVar == "_" {
		loopVar = "_i"
	}

	g.buf.WriteString("for ")
	g.buf.WriteString(loopVar)
	g.buf.WriteString(" := ")
	g.genExpr(start)
	g.buf.WriteString("; ")
	g.buf.WriteString(loopVar)
	g.buf.WriteString(" >= ")
	if len(args) > 0 {
		g.genExpr(args[0])
	} else {
		g.buf.WriteString("0")
	}
	g.buf.WriteString("; ")
	g.buf.WriteString(loopVar)
	g.buf.WriteString("-- {\n")

	if varName != "_" {
		g.vars[varName] = true
	}

	g.indent++
	for _, stmt := range block.Body {
		g.genStatement(stmt)
	}
	g.indent--

	if !wasDefinedBefore && varName != "_" {
		delete(g.vars, varName)
	}

	g.writeIndent()
	g.buf.WriteString("}")
}

// genRuntimeBlock generates runtime block calls (map, select, reject, reduce, find)
// Unified handler for all runtime block methods
func (g *Generator) genRuntimeBlock(iterable ast.Expression, block *ast.BlockExpr, args []ast.Expression, method blockMethod) {
	g.needsRuntime = true

	// Extract parameter names
	var param1Name, param2Name string
	if method.hasAccumulator {
		// reduce: |acc, elem|
		param1Name = "acc"
		param2Name = "_"
		if len(block.Params) > 0 {
			param1Name = block.Params[0]
		}
		if len(block.Params) > 1 {
			param2Name = block.Params[1]
		}
	} else {
		// map/select/reject/find: |elem|
		param1Name = "_"
		if len(block.Params) > 0 {
			param1Name = block.Params[0]
		}
	}

	// Save variable state
	param1WasDefinedBefore := g.vars[param1Name]
	param2WasDefinedBefore := g.vars[param2Name]

	// Generate: runtime.Method(iterable, [initial,] func(...) returnType {
	g.buf.WriteString(method.runtimeFunc)
	g.buf.WriteString("(")
	g.genExpr(iterable)

	// For reduce, add initial value
	if method.hasAccumulator {
		g.buf.WriteString(", ")
		if len(args) > 0 {
			g.genExpr(args[0])
		} else {
			g.buf.WriteString("nil")
		}
	}

	// Generate function literal
	g.buf.WriteString(", func(")
	if method.hasAccumulator {
		g.buf.WriteString(param1Name)
		g.buf.WriteString(" interface{}, ")
		g.buf.WriteString(param2Name)
		g.buf.WriteString(" interface{}")
	} else {
		g.buf.WriteString(param1Name)
		g.buf.WriteString(" interface{}")
	}
	g.buf.WriteString(") ")
	g.buf.WriteString(method.returnType)
	g.buf.WriteString(" {\n")

	// Mark variables as declared
	if param1Name != "_" {
		g.vars[param1Name] = true
	}
	if param2Name != "_" {
		g.vars[param2Name] = true
	}

	g.indent++

	// Generate block body with last statement as return
	if len(block.Body) > 0 {
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
			g.genStatement(lastStmt)
		}
	}

	g.indent--

	// Restore variable state
	if !param1WasDefinedBefore && param1Name != "_" {
		delete(g.vars, param1Name)
	}
	if !param2WasDefinedBefore && param2Name != "_" {
		delete(g.vars, param2Name)
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

// receiverName returns the Go receiver variable name for a class.
// Uses the first letter lowercased, or "x" for empty/invalid input.
func receiverName(className string) string {
	if len(className) == 0 {
		return "x"
	}
	return strings.ToLower(className[:1])
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
