// Package codegen generates Go code from Rugby AST.
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
	runtimeFunc     string
	returnType      string // "any" or "bool"
	hasAccumulator  bool   // for reduce - takes 2 params (acc, elem) instead of 1
	usesIncludeFlag bool   // for map - uses (value, include, continue) instead of (value, continue)
}

// Block method mappings - single source of truth
var blockMethods = map[string]blockMethod{
	"map":    {runtimeFunc: "runtime.Map", returnType: "any", usesIncludeFlag: true},
	"select": {runtimeFunc: "runtime.Select", returnType: "bool"},
	"filter": {runtimeFunc: "runtime.Select", returnType: "bool"},
	"reject": {runtimeFunc: "runtime.Reject", returnType: "bool"},
	"find":   {runtimeFunc: "runtime.Find", returnType: "bool"},
	"detect": {runtimeFunc: "runtime.Find", returnType: "bool"},
	"reduce": {runtimeFunc: "runtime.Reduce", returnType: "any", hasAccumulator: true},
	"any?":   {runtimeFunc: "runtime.Any", returnType: "bool"},
	"all?":   {runtimeFunc: "runtime.All", returnType: "bool"},
	"none?":  {runtimeFunc: "runtime.None", returnType: "bool"},
}

// optionalMethod describes a method that returns (value, bool) for optional results
type optionalMethod struct {
	runtimeFunc string // e.g., "runtime.StringToInt"
	resultType  string // the type of the value (e.g., "Int")
}

// Optional method mappings - methods that return (T, bool)
var optionalMethods = map[string]optionalMethod{
	"to_i?": {runtimeFunc: "runtime.StringToInt", resultType: "Int"},
	"to_f?": {runtimeFunc: "runtime.StringToFloat", resultType: "Float"},
}

// MethodDef describes a standard library method
type MethodDef struct {
	RuntimeFunc string
	ReturnType  string // optional, used for type inference of the result
}

// stdLib defines the standard library methods
// ReceiverType -> MethodName -> MethodDef
var stdLib = map[string]map[string]MethodDef{
	"Array": {
		"join":     {RuntimeFunc: "runtime.Join", ReturnType: "String"},
		"flatten":  {RuntimeFunc: "runtime.Flatten", ReturnType: "Array"},
		"uniq":     {RuntimeFunc: "runtime.Uniq", ReturnType: "Array"},
		"sort":     {RuntimeFunc: "runtime.Sort", ReturnType: "Array"},
		"shuffle":  {RuntimeFunc: "runtime.Shuffle", ReturnType: "Array"},
		"sample":   {RuntimeFunc: "runtime.Sample", ReturnType: ""}, // T
		"first":    {RuntimeFunc: "runtime.First", ReturnType: ""},  // T
		"last":     {RuntimeFunc: "runtime.Last", ReturnType: ""},   // T
		"reverse":  {RuntimeFunc: "runtime.Reversed", ReturnType: "Array"},
		"rotate":   {RuntimeFunc: "runtime.Rotate", ReturnType: "Array"},
		"include?": {RuntimeFunc: "runtime.Contains", ReturnType: "Bool"},
	},
	"Map": {
		"keys":     {RuntimeFunc: "runtime.Keys", ReturnType: "Array"},
		"values":   {RuntimeFunc: "runtime.Values", ReturnType: "Array"},
		"has_key?": {RuntimeFunc: "runtime.MapHasKey", ReturnType: "Bool"},
		"delete":   {RuntimeFunc: "runtime.MapDelete", ReturnType: ""}, // V
		"clear":    {RuntimeFunc: "runtime.MapClear", ReturnType: ""},
		"invert":   {RuntimeFunc: "runtime.MapInvert", ReturnType: "Map"},
		"merge":    {RuntimeFunc: "runtime.Merge", ReturnType: "Map"},
	},
	"String": {
		"split":      {RuntimeFunc: "runtime.Split", ReturnType: "Array"},
		"strip":      {RuntimeFunc: "runtime.Strip", ReturnType: "String"},
		"lstrip":     {RuntimeFunc: "runtime.Lstrip", ReturnType: "String"},
		"rstrip":     {RuntimeFunc: "runtime.Rstrip", ReturnType: "String"},
		"upcase":     {RuntimeFunc: "runtime.Upcase", ReturnType: "String"},
		"downcase":   {RuntimeFunc: "runtime.Downcase", ReturnType: "String"},
		"capitalize": {RuntimeFunc: "runtime.Capitalize", ReturnType: "String"},
		"include?":   {RuntimeFunc: "runtime.StringContains", ReturnType: "Bool"},
		"gsub":       {RuntimeFunc: "runtime.Replace", ReturnType: "String"},
		"reverse":    {RuntimeFunc: "runtime.StringReverse", ReturnType: "String"},
		"chars":      {RuntimeFunc: "runtime.Chars", ReturnType: "Array"},
		"length":     {RuntimeFunc: "runtime.CharLength", ReturnType: "Int"},
		"to_i":       {RuntimeFunc: "runtime.MustAtoi", ReturnType: "Int"},
	},
	"Int": {
		"even?": {RuntimeFunc: "runtime.Even", ReturnType: "Bool"},
		"odd?":  {RuntimeFunc: "runtime.Odd", ReturnType: "Bool"},
		"abs":   {RuntimeFunc: "runtime.Abs", ReturnType: "Int"},
		"clamp": {RuntimeFunc: "runtime.Clamp", ReturnType: "Int"},
	},
	"Math": { // Global Math module calls (Math.sqrt -> runtime.Sqrt)
		"sqrt":  {RuntimeFunc: "runtime.Sqrt", ReturnType: "Float"},
		"pow":   {RuntimeFunc: "runtime.Pow", ReturnType: "Float"},
		"ceil":  {RuntimeFunc: "runtime.Ceil", ReturnType: "Int"},
		"floor": {RuntimeFunc: "runtime.Floor", ReturnType: "Int"},
		"round": {RuntimeFunc: "runtime.Round", ReturnType: "Int"},
	},
}

// uniqueMethods maps method names to their runtime function if they are unique across the stdlib
// This helps when type inference fails.
var uniqueMethods = make(map[string]MethodDef)

func init() {
	// Populate uniqueMethods
	counts := make(map[string]int)
	for _, methods := range stdLib {
		for name := range methods {
			counts[name]++
		}
	}
	for _, methods := range stdLib {
		for name, def := range methods {
			if counts[name] == 1 {
				uniqueMethods[name] = def
			}
		}
	}
}

type contextType int

const (
	ctxLoop contextType = iota
	ctxIterBlock
	ctxTransformBlock
)

type loopContext struct {
	kind            contextType
	returnType      string // "any", "bool", or empty for iterative
	usesIncludeFlag bool   // for map - uses three-value returns
}

type Generator struct {
	buf                strings.Builder
	indent             int
	vars               map[string]string // track declared variables and their types (empty string = unknown type)
	imports            map[string]bool   // track import aliases for Go interop detection
	needsRuntime       bool              // track if rugby/runtime import is needed
	needsFmt           bool              // track if fmt import is needed (string interpolation)
	currentClass       string            // current class being generated (for instance vars)
	pubClasses         map[string]bool   // track public classes for constructor naming
	classFields        map[string]string // track fields of the current class and their types
	sourceFile         string            // original .rg filename for //line directives
	emitLineDir        bool              // whether to emit //line directives
	currentReturnTypes []string          // return types of the current function/method
	contexts           []loopContext     // stack of loop/block contexts
}

func (g *Generator) pushContext(kind contextType) {
	g.contexts = append(g.contexts, loopContext{kind: kind, returnType: "", usesIncludeFlag: false})
}

func (g *Generator) pushContextWithInclude(kind contextType, returnType string, usesInclude bool) {
	g.contexts = append(g.contexts, loopContext{kind: kind, returnType: returnType, usesIncludeFlag: usesInclude})
}

func (g *Generator) popContext() {
	if len(g.contexts) > 0 {
		g.contexts = g.contexts[:len(g.contexts)-1]
	}
}

func (g *Generator) currentContext() (loopContext, bool) {
	if len(g.contexts) == 0 {
		return loopContext{}, false
	}
	return g.contexts[len(g.contexts)-1], true
}

// Option configures a Generator.
type Option func(*Generator)

// WithSourceFile enables //line directive emission for the given source file.
func WithSourceFile(path string) Option {
	return func(g *Generator) {
		g.sourceFile = path
		g.emitLineDir = true
	}
}

func New(opts ...Option) *Generator {
	g := &Generator{
		vars:        make(map[string]string),
		imports:     make(map[string]bool),
		pubClasses:  make(map[string]bool),
		classFields: make(map[string]string),
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// isDeclared checks if a variable has been declared in the current scope
func (g *Generator) isDeclared(name string) bool {
	_, exists := g.vars[name]
	return exists
}

// isOptionalType checks if a type is an optional type (ends with ?)
func isOptionalType(t string) bool {
	return strings.HasSuffix(t, "?")
}

// isValueTypeOptional checks if a type is a value type optional (Int?, String?, etc.)
func isValueTypeOptional(t string) bool {
	if !isOptionalType(t) {
		return false
	}
	base := strings.TrimSuffix(t, "?")
	return valueTypes[base]
}

// inferTypeFromExpr attempts to infer the type from an expression
// Returns the inferred type or empty string if unknown
func (g *Generator) inferTypeFromExpr(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.CallExpr:
		// Check if it's a call to an optional method like s.to_i?
		if sel, ok := e.Func.(*ast.SelectorExpr); ok {
			if om, ok := optionalMethods[sel.Sel]; ok {
				return om.resultType
			}
		}
	case *ast.SelectorExpr:
		// Infer return type from method calls for chaining
		recvType := g.inferTypeFromExpr(e.X)
		if recvType != "" {
			if methods, ok := stdLib[recvType]; ok {
				if def, ok := methods[e.Sel]; ok && def.ReturnType != "" {
					return def.ReturnType
				}
			}
		}
	case *ast.IntLit:
		return "Int"
	case *ast.FloatLit:
		return "Float"
	case *ast.StringLit:
		return "String"
	case *ast.SymbolLit:
		return "String"
	case *ast.BoolLit:
		return "Bool"
	case *ast.Ident:
		if t, ok := g.vars[e.Name]; ok {
			return t
		}
	case *ast.InstanceVar:
		if t, ok := g.classFields[e.Name]; ok {
			return t
		}
	case *ast.RangeLit:
		return "Range"
	case *ast.ArrayLit:
		return "Array"
	case *ast.MapLit:
		return "Map"
	}
	return "" // unknown
}

// isOptionalMethodCall checks if an expression is a call to an optional method (to_i?, to_f?)
func isOptionalMethodCall(expr ast.Expression) bool {
	if call, ok := expr.(*ast.CallExpr); ok {
		if sel, ok := call.Func.(*ast.SelectorExpr); ok {
			_, isOptional := optionalMethods[sel.Sel]
			return isOptional
		}
	}
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		_, isOptional := optionalMethods[sel.Sel]
		return isOptional
	}
	return false
}

func (g *Generator) Generate(program *ast.Program) (string, error) {
	// Collect import aliases for Go interop detection
	for _, imp := range program.Imports {
		if imp.Alias != "" {
			g.imports[imp.Alias] = true
		} else {
			// Use last part of path as implicit alias (e.g., "encoding/json" -> "json")
			parts := strings.Split(imp.Path, "/")
			g.imports[parts[len(parts)-1]] = true
		}
	}

	// Separate declarations from top-level statements
	var definitions []ast.Statement   // def, class, interface
	var topLevelStmts []ast.Statement // executable statements
	hasMainFunc := false

	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			definitions = append(definitions, d)
			if d.Name == "main" {
				hasMainFunc = true
			}
		case *ast.ClassDecl:
			definitions = append(definitions, d)
		case *ast.InterfaceDecl:
			definitions = append(definitions, d)
		default:
			topLevelStmts = append(topLevelStmts, decl)
		}
	}

	// Check for conflict: both def main and top-level statements
	if hasMainFunc && len(topLevelStmts) > 0 {
		return "", fmt.Errorf("cannot mix top-level statements with 'def main'; use one or the other")
	}

	// First pass: generate definitions to determine what imports we need
	var bodyBuf strings.Builder
	g.buf = bodyBuf
	for _, def := range definitions {
		g.genStatement(def)
		g.buf.WriteString("\n")
	}

	// Generate implicit main if there are top-level statements and no explicit main
	if len(topLevelStmts) > 0 && !hasMainFunc {
		g.buf.WriteString("func main() {\n")
		g.indent++
		clear(g.vars) // Reset variable tracking for implicit main scope
		for _, stmt := range topLevelStmts {
			g.genStatement(stmt)
		}
		g.indent--
		g.buf.WriteString("}\n")
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
	needsFmtImport := g.needsFmt && !userImports["fmt"]
	hasImports := len(program.Imports) > 0 || needsRuntimeImport || needsFmtImport
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
		if needsFmtImport {
			out.WriteString("\t\"fmt\"\n")
		}
		if needsRuntimeImport {
			out.WriteString("\t\"rugby/runtime\"\n")
		}
		out.WriteString(")\n\n")
	}

	// Emit //line directive to map errors back to Rugby source
	if g.emitLineDir && g.sourceFile != "" {
		out.WriteString(fmt.Sprintf("//line %s:1\n", g.sourceFile))
	}

	out.WriteString(bodyCode)

	// Format the output
	source := out.String()
	formatted, err := format.Source([]byte(source))
	if err != nil {
		// Return unformatted source for debugging - intentionally no error
		// so caller can inspect what was generated (e.g., error comments like /* self outside class */)
		return source, nil //nolint:nilerr // intentional: return source for debugging
	}
	return string(formatted), nil
}

func (g *Generator) writeIndent() {
	for range g.indent {
		g.buf.WriteString("\t")
	}
}

func (g *Generator) genStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.FuncDecl:
		g.genFuncDecl(s)
	case *ast.ClassDecl:
		g.genClassDecl(s)
	case *ast.InterfaceDecl:
		g.genInterfaceDecl(s)
	case *ast.ExprStmt:
		g.genExprStmt(s)
	case *ast.AssignStmt:
		g.genAssignStmt(s)
	case *ast.OrAssignStmt:
		g.genOrAssignStmt(s)
	case *ast.CompoundAssignStmt:
		g.genCompoundAssignStmt(s)
	case *ast.InstanceVarAssign:
		g.genInstanceVarAssign(s)
	case *ast.InstanceVarOrAssign:
		g.genInstanceVarOrAssign(s)
	case *ast.IfStmt:
		g.genIfStmt(s)
	case *ast.CaseStmt:
		g.genCaseStmt(s)
	case *ast.WhileStmt:
		g.genWhileStmt(s)
	case *ast.ForStmt:
		g.genForStmt(s)
	case *ast.BreakStmt:
		g.genBreakStmt(s)
	case *ast.NextStmt:
		g.genNextStmt(s)
	case *ast.ReturnStmt:
		g.genReturnStmt(s)
	case *ast.DeferStmt:
		g.genDeferStmt(s)
	}
}

func (g *Generator) genFuncDecl(fn *ast.FuncDecl) {
	clear(g.vars) // reset vars for each function
	g.currentReturnTypes = fn.ReturnTypes

	// Mark parameters as declared variables with their types
	for _, param := range fn.Params {
		g.vars[param.Name] = param.Type
	}

	// Generate function name with proper casing
	// pub def parse_json -> ParseJSON (exported)
	// def parse_json -> parseJSON (unexported)
	var funcName string
	if fn.Pub {
		funcName = snakeToPascalWithAcronyms(fn.Name)
	} else {
		funcName = snakeToCamelWithAcronyms(fn.Name)
	}

	// Generate function signature
	g.buf.WriteString(fmt.Sprintf("func %s(", funcName))
	for i, param := range fn.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		if param.Type != "" {
			g.buf.WriteString(fmt.Sprintf("%s %s", param.Name, mapType(param.Type)))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s any", param.Name))
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
	for i, stmt := range fn.Body {
		isLast := i == len(fn.Body)-1
		if isLast && len(fn.ReturnTypes) > 0 {
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				// Implicit return: treat last expression as return value
				retStmt := &ast.ReturnStmt{Values: []ast.Expression{exprStmt.Expr}}
				g.genReturnStmt(retStmt)
				continue
			}
		}
		g.genStatement(stmt)
	}
	g.indent--
	g.buf.WriteString("}\n")
	g.currentReturnTypes = nil
}

func (g *Generator) genClassDecl(cls *ast.ClassDecl) {
	className := cls.Name
	g.currentClass = className
	g.pubClasses[className] = cls.Pub
	clear(g.classFields)

	// Emit struct definition
	// Class names are already PascalCase by convention; pub affects field/method visibility
	hasContent := len(cls.Embeds) > 0 || len(cls.Fields) > 0
	if hasContent {
		g.buf.WriteString(fmt.Sprintf("type %s struct {\n", className))
		for _, embed := range cls.Embeds {
			g.buf.WriteString("\t")
			g.buf.WriteString(embed)
			g.buf.WriteString("\n")
		}
		for _, field := range cls.Fields {
			g.classFields[field.Name] = field.Type // Store field type
			if field.Type != "" {
				g.buf.WriteString(fmt.Sprintf("\t%s %s\n", field.Name, mapType(field.Type)))
			} else {
				g.buf.WriteString(fmt.Sprintf("\t%s any\n", field.Name))
			}
		}
		g.buf.WriteString("}\n\n")
	} else {
		g.buf.WriteString(fmt.Sprintf("type %s struct{}\n\n", className))
	}

	// Emit constructor if initialize exists
	for _, method := range cls.Methods {
		if method.Name == "initialize" {
			g.genConstructor(className, method, cls.Pub)
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
	clear(g.classFields)
}

func (g *Generator) genInterfaceDecl(iface *ast.InterfaceDecl) {
	// Generate: type InterfaceName interface { ... }
	g.buf.WriteString(fmt.Sprintf("type %s interface {\n", iface.Name))

	for _, method := range iface.Methods {
		g.buf.WriteString("\t")
		// Interface methods are always exported (PascalCase) per Go interface rules
		methodName := snakeToPascalWithAcronyms(method.Name)
		g.buf.WriteString(methodName)
		g.buf.WriteString("(")

		// Parameters (just types, no names in Go interface definitions)
		for i, param := range method.Params {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			if param.Type != "" {
				g.buf.WriteString(mapType(param.Type))
			} else {
				g.buf.WriteString("any")
			}
		}
		g.buf.WriteString(")")

		// Return types
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

		g.buf.WriteString("\n")
	}

	g.buf.WriteString("}\n")
}

func (g *Generator) genConstructor(className string, method *ast.MethodDecl, pub bool) {
	clear(g.vars)

	// Mark parameters as declared variables
	for _, param := range method.Params {
		g.vars[param.Name] = param.Type
	}

	// Receiver name for field assignments
	recv := receiverName(className)

	// Constructor: func NewClassName(params) *ClassName (pub) or newClassName (non-pub)
	constructorName := "new" + className
	if pub {
		constructorName = "New" + className
	}
	g.buf.WriteString(fmt.Sprintf("func %s(", constructorName))
	for i, param := range method.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		if param.Type != "" {
			g.buf.WriteString(fmt.Sprintf("%s %s", param.Name, mapType(param.Type)))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s any", param.Name))
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
	clear(g.vars) // reset vars for each method
	g.currentReturnTypes = method.ReturnTypes

	// Mark parameters as declared variables with their types
	for _, param := range method.Params {
		g.vars[param.Name] = param.Type
	}

	// Receiver name: first letter of class, lowercase
	recv := receiverName(className)

	// Special method handling
	// to_s -> String() string (satisfies fmt.Stringer)
	// Only applies when to_s has no parameters
	if method.Name == "to_s" && len(method.Params) == 0 {
		g.buf.WriteString(fmt.Sprintf("func (%s *%s) String() string {\n", recv, className))
		g.indent++
		for _, stmt := range method.Body {
			g.genStatement(stmt)
		}
		g.indent--
		g.buf.WriteString("}\n\n")
		g.currentReturnTypes = nil
		return
	}

	// == -> Equal(other any) bool (satisfies runtime.Equaler)
	// Only applies when == has exactly one parameter
	if method.Name == "==" && len(method.Params) == 1 {
		param := method.Params[0]
		g.buf.WriteString(fmt.Sprintf("func (%s *%s) Equal(other any) bool {\n", recv, className))
		g.indent++
		// Create type assertion for the parameter
		g.writeIndent()
		g.buf.WriteString(fmt.Sprintf("%s, ok := other.(*%s)\n", param.Name, className))
		g.writeIndent()
		g.buf.WriteString("if !ok {\n")
		g.indent++
		g.writeIndent()
		g.buf.WriteString("return false\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
		// Generate body statements, with implicit return for last expression
		for i, stmt := range method.Body {
			isLast := i == len(method.Body)-1
			if isLast {
				// If last statement is an expression, add return
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
		g.indent--
		g.buf.WriteString("}\n\n")
		g.currentReturnTypes = nil
		return
	}

	// Method name: convert snake_case to proper casing
	// pub def in a pub class -> PascalCase (exported)
	// def in any class -> camelCase (unexported)
	var methodName string
	if method.Pub {
		methodName = snakeToPascalWithAcronyms(method.Name)
	} else {
		methodName = snakeToCamelWithAcronyms(method.Name)
	}

	// Generate method signature with pointer receiver: func (r *ClassName) MethodName(params) returns
	g.buf.WriteString(fmt.Sprintf("func (%s *%s) %s(", recv, className, methodName))
	for i, param := range method.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		if param.Type != "" {
			g.buf.WriteString(fmt.Sprintf("%s %s", param.Name, mapType(param.Type)))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s any", param.Name))
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
	for i, stmt := range method.Body {
		isLast := i == len(method.Body)-1
		if isLast && len(method.ReturnTypes) > 0 {
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				// Implicit return
				retStmt := &ast.ReturnStmt{Values: []ast.Expression{exprStmt.Expr}}
				g.genReturnStmt(retStmt)
				continue
			}
		}
		g.genStatement(stmt)
	}
	g.indent--
	g.buf.WriteString("}\n\n")
	g.currentReturnTypes = nil
}

func (g *Generator) genInstanceVarAssign(s *ast.InstanceVarAssign) {
	g.writeIndent()
	if g.currentClass != "" {
		recv := receiverName(g.currentClass)
		g.buf.WriteString(fmt.Sprintf("%s.%s = ", recv, s.Name))

		// Check for optional wrapping
		fieldType := g.classFields[s.Name]
		sourceType := g.inferTypeFromExpr(s.Value)
		needsWrap := false

		if isValueTypeOptional(fieldType) {
			baseType := strings.TrimSuffix(fieldType, "?")
			if sourceType == baseType {
				needsWrap = true
			}
		}

		if needsWrap {
			baseType := strings.TrimSuffix(fieldType, "?")
			g.buf.WriteString(fmt.Sprintf("runtime.Some%s(", baseType))
			g.genExpr(s.Value)
			g.buf.WriteString(")")
		} else {
			g.genExpr(s.Value)
		}
	} else {
		g.buf.WriteString(fmt.Sprintf("/* @%s outside class */ ", s.Name))
		g.genExpr(s.Value)
	}
	g.buf.WriteString("\n")
}

func (g *Generator) genOrAssignStmt(s *ast.OrAssignStmt) {
	if !g.isDeclared(s.Name) {
		// First declaration: just use :=
		g.writeIndent()
		g.buf.WriteString(s.Name)
		g.buf.WriteString(" := ")
		g.genExpr(s.Value)
		g.buf.WriteString("\n")
		g.vars[s.Name] = "" // type unknown
	} else {
		// Variable exists: generate check
		g.writeIndent()
		g.buf.WriteString("if ")

		declaredType := g.vars[s.Name]
		// Check if nil
		g.buf.WriteString(s.Name)
		g.buf.WriteString(" == nil")

		g.buf.WriteString(" {\n")
		g.indent++
		g.writeIndent()
		g.buf.WriteString(s.Name)
		g.buf.WriteString(" = ")

		// Check for optional wrapping
		sourceType := g.inferTypeFromExpr(s.Value)
		needsWrap := false
		if isValueTypeOptional(declaredType) {
			baseType := strings.TrimSuffix(declaredType, "?")
			if sourceType == baseType {
				needsWrap = true
			}
		}

		if needsWrap {
			baseType := strings.TrimSuffix(declaredType, "?")
			g.buf.WriteString(fmt.Sprintf("runtime.Some%s(", baseType))
			g.genExpr(s.Value)
			g.buf.WriteString(")")
		} else {
			g.genExpr(s.Value)
		}

		g.buf.WriteString("\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	}
}

func (g *Generator) genCompoundAssignStmt(s *ast.CompoundAssignStmt) {
	g.writeIndent()
	g.buf.WriteString(s.Name)
	g.buf.WriteString(" = ")
	g.buf.WriteString(s.Name)
	g.buf.WriteString(" ")
	g.buf.WriteString(s.Op)
	g.buf.WriteString(" ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

func (g *Generator) genInstanceVarOrAssign(s *ast.InstanceVarOrAssign) {
	if g.currentClass == "" {
		g.writeIndent()
		g.buf.WriteString(fmt.Sprintf("/* @%s ||= outside class */\n", s.Name))
		return
	}

	recv := receiverName(g.currentClass)
	field := fmt.Sprintf("%s.%s", recv, s.Name)
	fieldType := g.classFields[s.Name]

	// Generate check
	g.writeIndent()
	g.buf.WriteString("if ")

	// Check if nil
	g.buf.WriteString(field)
	g.buf.WriteString(" == nil")

	g.buf.WriteString(" {\n")
	g.indent++
	g.writeIndent()
	g.buf.WriteString(field)
	g.buf.WriteString(" = ")

	// Check for optional wrapping
	sourceType := g.inferTypeFromExpr(s.Value)
	needsWrap := false
	if isValueTypeOptional(fieldType) {
		baseType := strings.TrimSuffix(fieldType, "?")
		if sourceType == baseType {
			needsWrap = true
		}
	}

	if needsWrap {
		baseType := strings.TrimSuffix(fieldType, "?")
		g.buf.WriteString(fmt.Sprintf("runtime.Some%s(", baseType))
		g.genExpr(s.Value)
		g.buf.WriteString(")")
	} else {
		g.genExpr(s.Value)
	}

	g.buf.WriteString("\n")
	g.indent--
	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genAssignStmt(s *ast.AssignStmt) {
	g.writeIndent()

	// Handle optional method calls like s.to_i?() which return (T, bool)
	// Generate: n, _ := runtime.StringToInt(s)
	if isOptionalMethodCall(s.Value) {
		var receiver ast.Expression
		var method string

		if call, ok := s.Value.(*ast.CallExpr); ok {
			if sel, ok := call.Func.(*ast.SelectorExpr); ok {
				receiver = sel.X
				method = sel.Sel
			}
		} else if sel, ok := s.Value.(*ast.SelectorExpr); ok {
			receiver = sel.X
			method = sel.Sel
		}

		if receiver != nil {
			om := optionalMethods[method]
			// runtime.ToOptionalInt(runtime.StringToInt(...))
			baseType := om.resultType // "Int", "Float"

			if s.Type != "" {
				// Typed declaration
				g.buf.WriteString(fmt.Sprintf("var %s %s = ", s.Name, mapType(s.Type)))
				g.vars[s.Name] = s.Type
			} else if !g.isDeclared(s.Name) {
				// New variable - infer as OptionalT
				g.buf.WriteString(s.Name)
				g.buf.WriteString(" := ")
				g.vars[s.Name] = baseType + "?"
			} else {
				// Assignment
				g.buf.WriteString(s.Name)
				g.buf.WriteString(" = ")
			}

			g.buf.WriteString(fmt.Sprintf("runtime.ToOptional%s(", baseType))
			g.buf.WriteString(om.runtimeFunc)
			g.buf.WriteString("(")
			g.genExpr(receiver)
			g.buf.WriteString("))\n")
			return
		}
	}

	// Determine target type
	var targetType string
	if s.Type != "" {
		targetType = s.Type
	} else if declaredType, ok := g.vars[s.Name]; ok {
		targetType = declaredType
	}

	// Check if we are assigning nil to a value type optional
	isNilAssignment := false
	if _, ok := s.Value.(*ast.NilLit); ok {
		isNilAssignment = true
	}

	if isNilAssignment && isValueTypeOptional(targetType) {
		baseType := strings.TrimSuffix(targetType, "?")

		if s.Type != "" && !g.isDeclared(s.Name) {
			// Typed declaration: var x Int? = nil
			g.buf.WriteString(fmt.Sprintf("var %s %s = ", s.Name, mapType(s.Type)))
			g.vars[s.Name] = s.Type
		} else if g.isDeclared(s.Name) {
			// Reassignment: x = nil
			g.buf.WriteString(s.Name)
			g.buf.WriteString(" = ")
		} else {
			// Untyped declaration (x = nil)
			g.buf.WriteString(s.Name)
			g.buf.WriteString(" := ")
			g.vars[s.Name] = ""
		}

		g.buf.WriteString(fmt.Sprintf("runtime.None%s()", baseType))
		g.buf.WriteString("\n")
		return
	}

	// Check if we need to wrap value type -> Optional (e.g. x : Int? = 5)
	sourceType := g.inferTypeFromExpr(s.Value)
	needsWrap := false
	if isValueTypeOptional(targetType) {
		baseType := strings.TrimSuffix(targetType, "?")
		if sourceType == baseType {
			needsWrap = true
		}
	}

	if s.Type != "" && !g.isDeclared(s.Name) {
		// Typed declaration: var x int = value
		g.buf.WriteString(fmt.Sprintf("var %s %s = ", s.Name, mapType(s.Type)))
		g.vars[s.Name] = s.Type // store the declared type
	} else if g.isDeclared(s.Name) {
		// Reassignment: x = value
		g.buf.WriteString(s.Name)
		g.buf.WriteString(" = ")
	} else {
		// Untyped declaration: x := value
		g.buf.WriteString(s.Name)
		g.buf.WriteString(" := ")
		g.vars[s.Name] = sourceType // infer type from value if variable is new
	}

	if needsWrap {
		baseType := strings.TrimSuffix(targetType, "?")
		g.buf.WriteString(fmt.Sprintf("runtime.Some%s(", baseType))
		g.genExpr(s.Value)
		g.buf.WriteString(")")
	} else {
		g.genExpr(s.Value)
	}
	g.buf.WriteString("\n")
}

// genCondition generates a condition expression, handling optional types.
// For optional types: if x generates x.Valid (value types) or x != nil (reference types)
func (g *Generator) genCondition(cond ast.Expression) {
	// Check if condition is a simple identifier
	if ident, ok := cond.(*ast.Ident); ok {
		if varType, exists := g.vars[ident.Name]; exists && isOptionalType(varType) {
			// Optional type: check != nil
			g.buf.WriteString(ident.Name + " != nil")
			return
		}
	}
	// Default: generate expression as-is
	g.genExpr(cond)
}

func (g *Generator) genIfStmt(s *ast.IfStmt) {
	g.writeIndent()
	g.buf.WriteString("if ")

	// Handle assignment-in-condition pattern: if (n = s.to_i?)
	// Generates: if n, ok := runtime.StringToInt(s); ok {
	if s.AssignName != "" {
		g.buf.WriteString(s.AssignName)
		g.buf.WriteString(", ok := ")
		g.genExpr(s.AssignExpr)
		g.buf.WriteString("; ok {\n")
		// Track the variable with its inferred type
		g.vars[s.AssignName] = g.inferTypeFromExpr(s.AssignExpr)
	} else {
		// For unless, negate the condition
		if s.IsUnless {
			g.buf.WriteString("!")
		}
		g.genCondition(s.Cond)
		g.buf.WriteString(" {\n")
	}

	g.indent++
	for _, stmt := range s.Then {
		g.genStatement(stmt)
	}
	g.indent--

	for _, elsif := range s.ElseIfs {
		g.writeIndent()
		g.buf.WriteString("} else if ")
		g.genCondition(elsif.Cond)
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

func (g *Generator) genCaseStmt(s *ast.CaseStmt) {
	g.writeIndent()

	// Handle case with subject vs case without subject
	if s.Subject != nil {
		g.buf.WriteString("switch ")
		g.genExpr(s.Subject)
		g.buf.WriteString(" {\n")
	} else {
		// Case without subject - use switch true
		g.buf.WriteString("switch {\n")
	}

	// Generate when clauses
	for _, whenClause := range s.WhenClauses {
		g.writeIndent()

		if s.Subject != nil {
			// With subject: case value1, value2:
			g.buf.WriteString("case ")
			for i, val := range whenClause.Values {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				g.genExpr(val)
			}
			g.buf.WriteString(":\n")
		} else {
			// Without subject: case condition1 || condition2:
			g.buf.WriteString("case ")
			for i, val := range whenClause.Values {
				if i > 0 {
					g.buf.WriteString(" || ")
				}
				g.genCondition(val)
			}
			g.buf.WriteString(":\n")
		}

		g.indent++
		for _, stmt := range whenClause.Body {
			g.genStatement(stmt)
		}
		g.indent--
	}

	// Generate default (else) clause
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

func (g *Generator) genWhileStmt(s *ast.WhileStmt) {
	g.writeIndent()
	g.buf.WriteString("for ")
	g.genCondition(s.Cond)
	g.buf.WriteString(" {\n")

	g.pushContext(ctxLoop)
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.popContext()

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genForStmt(s *ast.ForStmt) {
	// Save variable state
	prevType, wasDefinedBefore := g.vars[s.Var]

	// Check if iterable is a range literal - optimize to C-style for loop
	if rangeLit, ok := s.Iterable.(*ast.RangeLit); ok {
		g.genForRangeLoop(s.Var, rangeLit, s.Body, wasDefinedBefore, prevType)
		return
	}

	// Check if iterable is a variable of type Range
	if ident, ok := s.Iterable.(*ast.Ident); ok {
		if g.vars[ident.Name] == "Range" {
			g.genForRangeVarLoop(s.Var, ident.Name, s.Body, wasDefinedBefore, prevType)
			return
		}
	}

	g.writeIndent()
	g.buf.WriteString("for _, ")
	g.buf.WriteString(s.Var)
	g.buf.WriteString(" := range ")
	g.genExpr(s.Iterable)
	g.buf.WriteString(" {\n")

	g.vars[s.Var] = "" // loop variable, type unknown

	g.pushContext(ctxLoop)
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.popContext()

	// Restore variable state
	if !wasDefinedBefore {
		delete(g.vars, s.Var)
	} else {
		g.vars[s.Var] = prevType
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genForRangeVarLoop(varName string, rangeVar string, body []ast.Statement, wasDefinedBefore bool, prevType string) {
	g.writeIndent()
	g.buf.WriteString("for ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := ")
	g.buf.WriteString(rangeVar)
	g.buf.WriteString(".Start; ")

	// Condition: (r.Exclusive && i < r.End) || (!r.Exclusive && i <= r.End)
	g.buf.WriteString(fmt.Sprintf("(%s.Exclusive && %s < %s.End) || (!%s.Exclusive && %s <= %s.End)",
		rangeVar, varName, rangeVar, rangeVar, varName, rangeVar))

	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	g.buf.WriteString("++ {\n")

	g.vars[varName] = "Int"

	g.pushContext(ctxLoop)
	g.indent++
	for _, stmt := range body {
		g.genStatement(stmt)
	}
	g.indent--
	g.popContext()

	if !wasDefinedBefore {
		delete(g.vars, varName)
	} else {
		g.vars[varName] = prevType
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genForRangeLoop(varName string, r *ast.RangeLit, body []ast.Statement, wasDefinedBefore bool, prevType string) {
	// Generate: for i := start; i <= end; i++ { ... }
	// or:       for i := start; i < end; i++ { ... } (exclusive)
	g.writeIndent()
	g.buf.WriteString("for ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := ")
	g.genExpr(r.Start)
	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	if r.Exclusive {
		g.buf.WriteString(" < ")
	} else {
		g.buf.WriteString(" <= ")
	}
	g.genExpr(r.End)
	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	g.buf.WriteString("++ {\n")

	g.vars[varName] = "Int" // range loop variable is always Int

	g.pushContext(ctxLoop)
	g.indent++
	for _, stmt := range body {
		g.genStatement(stmt)
	}
	g.indent--
	g.popContext()

	// Restore variable state
	if !wasDefinedBefore {
		delete(g.vars, varName)
	} else {
		g.vars[varName] = prevType
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genExprStmt(s *ast.ExprStmt) {
	if s.Condition != nil {
		// Wrap in if/unless block
		g.writeIndent()
		g.buf.WriteString("if ")
		if s.IsUnless {
			g.buf.WriteString("!(")
		}
		g.genCondition(s.Condition)
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
		g.indent++
		g.writeIndent()
		g.genExpr(s.Expr)
		g.buf.WriteString("\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	} else {
		g.writeIndent()
		g.genExpr(s.Expr)
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
		g.genCondition(s.Condition)
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
		g.indent++
	}

	g.writeIndent()
	ctx, ok := g.currentContext()
	if ok {
		switch ctx.kind {
		case ctxIterBlock:
			g.buf.WriteString("return false\n")
		case ctxTransformBlock:
			// Transform blocks with three values (map): return (value, include, continue)
			if ctx.usesIncludeFlag {
				g.buf.WriteString("return nil, false, false\n")
			} else {
				// Transform blocks with two values (select, reject, find): return (value, continue)
				if ctx.returnType == "bool" {
					g.buf.WriteString("return false, false\n")
				} else {
					g.buf.WriteString("return nil, false\n")
				}
			}
		default:
			// Regular loop context
			g.buf.WriteString("break\n")
		}
	} else {
		// No context - regular loop
		g.buf.WriteString("break\n")
	}

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
		g.genCondition(s.Condition)
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
		g.indent++
	}

	g.writeIndent()
	ctx, ok := g.currentContext()
	if ok {
		switch ctx.kind {
		case ctxIterBlock:
			g.buf.WriteString("return true\n")
		case ctxTransformBlock:
			// next in transform blocks with three values (map): skip element, continue iteration
			if ctx.usesIncludeFlag {
				g.buf.WriteString("return nil, false, true\n")
			} else {
				// next in two-value transform blocks: return zero value and true to continue
				if ctx.returnType == "bool" {
					g.buf.WriteString("return false, true\n")
				} else {
					g.buf.WriteString("return nil, true\n")
				}
			}
		default:
			// Regular loop context
			g.buf.WriteString("continue\n")
		}
	} else {
		// No context - regular loop
		g.buf.WriteString("continue\n")
	}

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
		g.genCondition(s.Condition)
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
		g.indent++
	}

	g.writeIndent()
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

			if i < len(g.currentReturnTypes) {
				targetType := g.currentReturnTypes[i]
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

func (g *Generator) genExpr(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		g.genCallExpr(e)
	case *ast.SelectorExpr:
		g.genSelectorExpr(e)
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
	case *ast.ArrayLit:
		g.genArrayLit(e)
	case *ast.IndexExpr:
		g.genIndexExpr(e)
	case *ast.MapLit:
		g.genMapLit(e)
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
	case *ast.BinaryExpr:
		g.genBinaryExpr(e)
	case *ast.UnaryExpr:
		g.genUnaryExpr(e)
	case *ast.RangeLit:
		g.genRangeLit(e)
	}
}

func (g *Generator) genRangeLit(r *ast.RangeLit) {
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

// genRangeMethodCall handles method calls on range literals
// Returns true if it handled the call, false otherwise
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
	// For == and != use runtime.Equal when at least one operand is not a primitive literal
	// This handles cases like: user1 == user2, x == 5, arr1 == arr2
	// Only use direct == for literal-to-literal comparisons like 5 == 5 or "a" == "b"
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

	g.buf.WriteString("(")
	g.genExpr(e.Left)
	g.buf.WriteString(" ")
	g.buf.WriteString(e.Op)
	g.buf.WriteString(" ")
	g.genExpr(e.Right)
	g.buf.WriteString(")")
}

// isPrimitiveLiteral returns true if the expression is a primitive literal
// (int, float, bool, string) that can be safely compared with Go's ==
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

func (g *Generator) genInterpolatedString(s *ast.InterpolatedString) {
	g.needsFmt = true

	// Build format string and collect expressions
	var formatParts []string
	var exprs []ast.Expression

	for _, part := range s.Parts {
		switch p := part.(type) {
		case string:
			// Escape % for fmt.Sprintf
			escaped := strings.ReplaceAll(p, "%", "%%")
			formatParts = append(formatParts, escaped)
		case ast.Expression:
			formatParts = append(formatParts, "%v")
			exprs = append(exprs, p)
		}
	}

	formatStr := strings.Join(formatParts, "")

	// Generate fmt.Sprintf call
	g.buf.WriteString("fmt.Sprintf(")
	g.buf.WriteString(fmt.Sprintf("%q", formatStr))
	for _, expr := range exprs {
		g.buf.WriteString(", ")
		g.genExpr(expr)
	}
	g.buf.WriteString(")")
}

func (g *Generator) genArrayLit(arr *ast.ArrayLit) {
	// Try to infer element type from elements
	elemType := ""
	consistent := true

	if len(arr.Elements) > 0 {
		for _, elem := range arr.Elements {
			t := g.inferTypeFromExpr(elem)
			if t == "" {
				consistent = false
				break
			}
			if elemType == "" {
				elemType = t
			} else if elemType != t {
				consistent = false
				break
			}
		}
	} else {
		// Empty array - ambiguous without context
		consistent = false
	}

	if consistent && elemType != "" {
		goType := mapType(elemType)
		g.buf.WriteString("[]" + goType + "{")
	} else {
		g.buf.WriteString("[]any{")
	}

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
		} else {
			// Transform local function names to match their definition (snake_case -> camelCase)
			funcName = snakeToCamelWithAcronyms(funcName)
		}
		g.buf.WriteString(funcName)
	case *ast.SelectorExpr:
		// Check for ClassName.new() constructor call
		if fn.Sel == "new" {
			if ident, ok := fn.X.(*ast.Ident); ok {
				// Rewrite User.new(args) to NewUser(args) or newUser(args)
				// based on class visibility
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

		// Check standard library registry
		recvType := g.inferTypeFromExpr(fn.X)
		var methodDef MethodDef
		found := false

		// 1. Special case for Math module (global) - check before type lookup
		if ident, ok := fn.X.(*ast.Ident); ok && ident.Name == "Math" {
			if methods, ok := stdLib["Math"]; ok {
				if def, ok := methods[fn.Sel]; ok {
					g.needsRuntime = true
					g.buf.WriteString(def.RuntimeFunc)
					g.buf.WriteString("(")
					// Math methods don't take receiver as first arg
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

		// 2. Try lookup by type
		if recvType != "" {
			if methods, ok := stdLib[recvType]; ok {
				if def, ok := methods[fn.Sel]; ok {
					methodDef = def
					found = true
				}
			}
		}

		// 3. Try unique method lookup if type unknown or method not found on type
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
			g.genExpr(fn.X) // Receiver is first arg
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

		// Check for range method calls when receiver is a RangeLit
		if _, ok := fn.X.(*ast.RangeLit); ok {
			if g.genRangeMethodCall(fn.X, fn.Sel, call.Args) {
				return
			}
		}
		// Check for optional methods like to_i?, to_f?
		if om, ok := optionalMethods[fn.Sel]; ok {
			g.needsRuntime = true
			g.buf.WriteString(om.runtimeFunc)
			g.buf.WriteString("(")
			g.genExpr(fn.X) // receiver becomes first arg
			g.buf.WriteString(")")
			return
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

// genEachBlock generates: runtime.Each(iterable, func(v any) { body })
// Uses runtime call so that return inside block exits only the block, not enclosing function
func (g *Generator) genEachBlock(iterable ast.Expression, block *ast.BlockExpr) {
	// Check if iterable is a Range - use RangeEach for type-safe iteration
	if _, ok := iterable.(*ast.RangeLit); ok {
		g.genRangeEachBlock(iterable, block)
		return
	}

	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	// Save previous variable state for proper scope restoration
	prevType, wasDefinedBefore := g.vars[varName]

	g.buf.WriteString("runtime.Each(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" any) bool {\n")

	if varName != "_" {
		g.vars[varName] = "" // block param, type unknown
	}

	g.pushContext(ctxIterBlock)
	g.indent++
	for _, stmt := range block.Body {
		g.genStatement(stmt)
	}
	g.writeIndent()
	g.buf.WriteString("return true\n")
	g.indent--
	g.popContext()

	// Restore variable state after block
	if varName != "_" {
		if !wasDefinedBefore {
			delete(g.vars, varName)
		} else {
			g.vars[varName] = prevType
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genRangeEachBlock generates: runtime.RangeEach(range, func(i int) { body })
func (g *Generator) genRangeEachBlock(rangeExpr ast.Expression, block *ast.BlockExpr) {
	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	prevType, wasDefinedBefore := g.vars[varName]

	g.buf.WriteString("runtime.RangeEach(")
	g.genExpr(rangeExpr)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" int) bool {\n")

	if varName != "_" {
		g.vars[varName] = "Int" // range iteration is always int
	}

	g.pushContext(ctxIterBlock)
	g.indent++
	for _, stmt := range block.Body {
		g.genStatement(stmt)
	}
	g.writeIndent()
	g.buf.WriteString("return true\n")
	g.indent--
	g.popContext()

	if varName != "_" {
		if !wasDefinedBefore {
			delete(g.vars, varName)
		} else {
			g.vars[varName] = prevType
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genEachWithIndexBlock generates: runtime.EachWithIndex(iterable, func(v any, i int) { body })
// Note: Rugby uses |value, index| order which matches runtime.EachWithIndex signature
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

	// Save previous variable state for proper scope restoration
	varPrevType, varWasDefinedBefore := g.vars[varName]
	indexPrevType, indexWasDefinedBefore := g.vars[indexName]

	g.buf.WriteString("runtime.EachWithIndex(")
	g.genExpr(iterable)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" any, ")
	g.buf.WriteString(indexName)
	g.buf.WriteString(" int) bool {\n")

	if varName != "_" {
		g.vars[varName] = "" // block param, type unknown
	}
	if indexName != "_" {
		g.vars[indexName] = "Int" // index is always int
	}

	g.pushContext(ctxIterBlock)
	g.indent++
	for _, stmt := range block.Body {
		g.genStatement(stmt)
	}
	g.writeIndent()
	g.buf.WriteString("return true\n")
	g.indent--
	g.popContext()

	// Restore variable state after block
	if varName != "_" {
		if !varWasDefinedBefore {
			delete(g.vars, varName)
		} else {
			g.vars[varName] = varPrevType
		}
	}
	if indexName != "_" {
		if !indexWasDefinedBefore {
			delete(g.vars, indexName)
		} else {
			g.vars[indexName] = indexPrevType
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genTimesBlock generates: runtime.Times(n, func(i int) { body })
func (g *Generator) genTimesBlock(times ast.Expression, block *ast.BlockExpr) {
	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	prevType, wasDefinedBefore := g.vars[varName]

	g.buf.WriteString("runtime.Times(")
	g.genExpr(times)
	g.buf.WriteString(", func(")
	g.buf.WriteString(varName)
	g.buf.WriteString(" int) bool {\n")

	if varName != "_" {
		g.vars[varName] = "Int" // times iteration is always int
	}

	g.pushContext(ctxIterBlock)
	g.indent++
	for _, stmt := range block.Body {
		g.genStatement(stmt)
	}
	g.writeIndent()
	g.buf.WriteString("return true\n")
	g.indent--
	g.popContext()

	if varName != "_" {
		if !wasDefinedBefore {
			delete(g.vars, varName)
		} else {
			g.vars[varName] = prevType
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genUptoBlock generates: runtime.Upto(start, end, func(i int) { body })
func (g *Generator) genUptoBlock(start ast.Expression, block *ast.BlockExpr, args []ast.Expression) {
	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	prevType, wasDefinedBefore := g.vars[varName]

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

	if varName != "_" {
		g.vars[varName] = "Int" // upto iteration is always int
	}

	g.pushContext(ctxIterBlock)
	g.indent++
	for _, stmt := range block.Body {
		g.genStatement(stmt)
	}
	g.writeIndent()
	g.buf.WriteString("return true\n")
	g.indent--
	g.popContext()

	if varName != "_" {
		if !wasDefinedBefore {
			delete(g.vars, varName)
		} else {
			g.vars[varName] = prevType
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
}

// genDowntoBlock generates: runtime.Downto(start, end, func(i int) { body })
func (g *Generator) genDowntoBlock(start ast.Expression, block *ast.BlockExpr, args []ast.Expression) {
	g.needsRuntime = true

	varName := "_"
	if len(block.Params) > 0 {
		varName = block.Params[0]
	}

	prevType, wasDefinedBefore := g.vars[varName]

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

	if varName != "_" {
		g.vars[varName] = "Int" // downto iteration is always int
	}

	g.pushContext(ctxIterBlock)
	g.indent++
	for _, stmt := range block.Body {
		g.genStatement(stmt)
	}
	g.writeIndent()
	g.buf.WriteString("return true\n")
	g.indent--
	g.popContext()

	if varName != "_" {
		if !wasDefinedBefore {
			delete(g.vars, varName)
		} else {
			g.vars[varName] = prevType
		}
	}

	g.writeIndent()
	g.buf.WriteString("})")
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
	param1PrevType, param1WasDefinedBefore := g.vars[param1Name]
	param2PrevType, param2WasDefinedBefore := g.vars[param2Name]

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
		g.buf.WriteString(" any, ")
		g.buf.WriteString(param2Name)
		g.buf.WriteString(" any")
	} else {
		g.buf.WriteString(param1Name)
		g.buf.WriteString(" any")
	}
	g.buf.WriteString(") (")
	g.buf.WriteString(method.returnType)
	// Map uses three-value return (value, include, continue)
	if method.usesIncludeFlag {
		g.buf.WriteString(", bool, bool) {\n")
	} else {
		g.buf.WriteString(", bool) {\n")
	}

	// Mark variables as declared (block params, type unknown)
	if param1Name != "_" {
		g.vars[param1Name] = ""
	}
	if param2Name != "_" {
		g.vars[param2Name] = ""
	}

	g.pushContextWithInclude(ctxTransformBlock, method.returnType, method.usesIncludeFlag)
	g.indent++

	// Generate block body with last statement as return
	if len(block.Body) > 0 {
		for _, stmt := range block.Body[:len(block.Body)-1] {
			g.genStatement(stmt)
		}
		// Last statement should be an expression - return it
		lastStmt := block.Body[len(block.Body)-1]
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			g.writeIndent()
			g.buf.WriteString("return ")
			g.genExpr(exprStmt.Expr)
			if method.usesIncludeFlag {
				g.buf.WriteString(", true, true\n") // value, include=true, continue=true
			} else {
				g.buf.WriteString(", true\n")
			}
		} else {
			g.genStatement(lastStmt)
			g.writeIndent()
			// Default return if last stmt was not an expression
			if method.usesIncludeFlag {
				g.buf.WriteString("return nil, false, true\n") // skip element, continue
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
			g.buf.WriteString("return nil, false, true\n") // skip element, continue
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

	// Restore variable state
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
	// Check for range methods used without parentheses (property-style access)
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

	// Check for optional methods like to_i?, to_f?
	if om, ok := optionalMethods[sel.Sel]; ok {
		g.needsRuntime = true
		g.buf.WriteString(om.runtimeFunc)
		g.buf.WriteString("(")
		g.genExpr(sel.X) // receiver becomes first arg
		g.buf.WriteString(")")
		return
	}

	g.genExpr(sel.X)
	g.buf.WriteString(".")

	// Check if receiver is a Go import (use PascalCase) or Rugby object (use camelCase)
	if g.isGoInterop(sel.X) {
		g.buf.WriteString(snakeToPascal(sel.Sel))
	} else {
		g.buf.WriteString(snakeToCamel(sel.Sel))
	}
}

// isGoInterop checks if an expression refers to a Go import.
// NOTE: This can produce false positives if a local variable shadows an import name.
// For example, if you `import io` and also have a local `io := something()`,
// method calls on that variable will incorrectly use PascalCase.
func (g *Generator) isGoInterop(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return g.imports[e.Name]
	case *ast.SelectorExpr:
		// For chained selectors like resp.Body.Close, check the root
		return g.isGoInterop(e.X)
	default:
		return false
	}
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

// snakeToCamel converts snake_case to camelCase (lowercase first letter)
// Used for private method names in Rugby classes
// Examples: read_all -> readAll, do_something -> doSomething
// Also strips Ruby-style suffixes: inc! -> inc, empty? -> empty
// Non-snake_case identifiers pass through as-is to support Go interop on variables
func snakeToCamel(s string) string {
	if s == "" {
		return s
	}

	// Strip Ruby-style suffixes (! for mutation, ? for predicates)
	s = strings.TrimSuffix(s, "!")
	s = strings.TrimSuffix(s, "?")

	// If no underscore, pass through as-is (preserves Go PascalCase on variables)
	if !strings.Contains(s, "_") {
		return s
	}

	var result strings.Builder
	first := true
	capitalizeNext := false

	for _, r := range s {
		if r == '_' {
			capitalizeNext = true
			continue
		}
		if first {
			result.WriteRune(unicode.ToLower(r))
			first = false
		} else if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// snakeToPascal converts snake_case to PascalCase (uppercase first letter)
// Used for Go interop where exports are always capitalized
// Examples: read_all -> ReadAll, new_request -> NewRequest
// Non-snake_case identifiers are passed through as-is (already PascalCase for Go)
func snakeToPascal(s string) string {
	if s == "" {
		return s
	}

	// If no underscore, pass through as-is (Go identifiers are already correct case)
	if !strings.Contains(s, "_") {
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

// Value types that need runtime.OptionalT wrapper for optional types
var valueTypes = map[string]bool{
	"Int": true, "Int64": true, "Float": true,
	"Bool": true, "String": true,
}

// mapType converts Rugby type names to Go type names
func mapType(rubyType string) string {
	// Handle Array[T]
	if strings.HasPrefix(rubyType, "Array[") && strings.HasSuffix(rubyType, "]") {
		inner := rubyType[6 : len(rubyType)-1]
		return "[]" + mapType(inner)
	}
	// Handle Map[K, V]
	if strings.HasPrefix(rubyType, "Map[") && strings.HasSuffix(rubyType, "]") {
		content := rubyType[4 : len(rubyType)-1]
		// Simple comma split for now (assuming no nested generic commas for MVP)
		parts := strings.Split(content, ",")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			return fmt.Sprintf("map[%s]%s", mapType(key), mapType(val))
		}
	}

	// Handle optional types (T?)
	if baseType, found := strings.CutSuffix(rubyType, "?"); found {
		if valueTypes[baseType] {
			// Value type optional -> *T
			return "*" + mapType(baseType)
		}
		// Reference types: slices/maps are already nullable, classes use pointer
		goBase := mapType(baseType)
		if strings.HasPrefix(goBase, "[]") || strings.HasPrefix(goBase, "map[") {
			return goBase // slices and maps are already nullable
		}
		return "*" + goBase // class types become pointers
	}

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
	case "Array":
		return "[]any"
	case "Map":
		return "map[any]any"
	default:
		return rubyType // pass through unknown types (e.g., user-defined)
	}
}

// Common acronyms that should be uppercased in Go identifiers
// Per spec section 9.5
var acronyms = map[string]string{
	"id":    "ID",
	"url":   "URL",
	"uri":   "URI",
	"http":  "HTTP",
	"https": "HTTPS",
	"json":  "JSON",
	"xml":   "XML",
	"api":   "API",
	"uuid":  "UUID",
	"ip":    "IP",
	"tcp":   "TCP",
	"udp":   "UDP",
	"sql":   "SQL",
	"tls":   "TLS",
	"ssh":   "SSH",
	"cpu":   "CPU",
	"gpu":   "GPU",
	"html":  "HTML",
	"css":   "CSS",
}

// snakeToPascalWithAcronyms converts snake_case to PascalCase with acronym handling
// Examples: user_id -> UserID, parse_json -> ParseJSON, read_http_url -> ReadHTTPURL
func snakeToPascalWithAcronyms(s string) string {
	if s == "" {
		return s
	}

	// Strip Ruby-style suffixes (! for mutation, ? for predicates)
	s = strings.TrimSuffix(s, "!")
	s = strings.TrimSuffix(s, "?")

	// If no underscore, check for single-word acronym or capitalize first letter
	if !strings.Contains(s, "_") {
		if upper, ok := acronyms[strings.ToLower(s)]; ok {
			return upper
		}
		// Capitalize first letter only
		if len(s) > 0 {
			return strings.ToUpper(s[:1]) + s[1:]
		}
		return s
	}

	// Split by underscore and process each part
	parts := strings.Split(s, "_")
	var result strings.Builder

	for _, part := range parts {
		if part == "" {
			continue
		}
		if upper, ok := acronyms[strings.ToLower(part)]; ok {
			result.WriteString(upper)
		} else {
			// Capitalize first letter
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(part[1:])
			}
		}
	}

	return result.String()
}

// snakeToCamelWithAcronyms converts snake_case to camelCase with acronym handling
// Examples: user_id -> userID, parse_json -> parseJSON, http_request -> httpRequest
// Note: first-part acronyms stay lowercase in camelCase
func snakeToCamelWithAcronyms(s string) string {
	if s == "" {
		return s
	}

	// Strip Ruby-style suffixes (! for mutation, ? for predicates)
	s = strings.TrimSuffix(s, "!")
	s = strings.TrimSuffix(s, "?")

	// If no underscore, return as-is (already camelCase or single word)
	if !strings.Contains(s, "_") {
		return s
	}

	// Split by underscore and process each part
	parts := strings.Split(s, "_")
	var result strings.Builder

	for i, part := range parts {
		if part == "" {
			continue
		}
		if i == 0 {
			// First part: lowercase, even if it's an acronym
			result.WriteString(strings.ToLower(part))
		} else {
			// Subsequent parts: uppercase acronyms, capitalize others
			if upper, ok := acronyms[strings.ToLower(part)]; ok {
				result.WriteString(upper)
			} else {
				result.WriteString(strings.ToUpper(part[:1]))
				if len(part) > 1 {
					result.WriteString(part[1:])
				}
			}
		}
	}

	return result.String()
}
