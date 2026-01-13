// Package codegen generates Go code from Rugby AST.
package codegen

import (
	"fmt"
	"go/format"
	"strings"

	"github.com/nchapman/rugby/ast"
)

// Rugby module prefix for stdlib imports
const rugbyModulePrefix = "github.com/nchapman/rugby"

// transformImportPath converts Rugby import paths to Go import paths.
// rugby/http -> github.com/nchapman/rugby/stdlib/http
// rugby/json -> github.com/nchapman/rugby/stdlib/json
// Other paths pass through unchanged.
func transformImportPath(path string) string {
	if subpath, found := strings.CutPrefix(path, "rugby/"); found {
		// Special cases: runtime and test are in the root, not stdlib
		if subpath == "runtime" || subpath == "test" {
			return rugbyModulePrefix + "/" + subpath
		}
		// All other rugby/* imports go to stdlib
		return rugbyModulePrefix + "/stdlib/" + subpath
	}
	return path
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

// TypeKind represents the kind of a type for codegen optimization decisions.
// This is intentionally a subset of semantic.TypeKind - only types that affect
// code generation optimizations are included. Unmapped types use TypeUnknown,
// which triggers safe fallback behavior (e.g., runtime.Equal for comparisons).
type TypeKind int

const (
	TypeUnknown TypeKind = iota
	TypeInt
	TypeInt64
	TypeFloat
	TypeBool
	TypeString
	TypeNil
	TypeArray
	TypeMap
	TypeClass
)

// TypeInfo provides type information for AST nodes during code generation.
// This enables optimizations like using direct == instead of runtime.Equal
// when we know both operands are primitive types, and generating typed
// slices/maps instead of []any/map[any]any.
type TypeInfo interface {
	// GetTypeKind returns the type kind for an expression node.
	GetTypeKind(node ast.Node) TypeKind

	// GetGoType returns the Go type string for an AST node.
	// For example: "int", "string", "[]int", "map[string]int".
	// Returns empty string if:
	// - No type info is available for the node
	// - The node's type cannot be converted to a Go type
	GetGoType(node ast.Node) string
}

type Generator struct {
	buf                strings.Builder
	indent             int
	vars               map[string]string          // track declared variables and their types (empty string = unknown type)
	imports            map[string]bool            // track import aliases for Go interop detection
	needsRuntime       bool                       // track if rugby/runtime import is needed
	needsFmt           bool                       // track if fmt import is needed (string interpolation)
	needsErrors        bool                       // track if errors import is needed (error_is?, error_as)
	needsTestingImport bool                       // track if testing import is needed
	needsTestImport    bool                       // track if rugby/test import is needed
	currentClass       string                     // current class being generated (for instance vars)
	currentMethod      string                     // current method name being generated (for super)
	currentMethodPub   bool                       // whether current method is pub (for super)
	currentClassEmbeds []string                   // embedded types (parent classes) of current class
	pubClasses         map[string]bool            // track public classes for constructor naming
	classFields        map[string]string          // track fields of the current class and their types
	modules            map[string]*ast.ModuleDecl // track module definitions for include
	interfaces         map[string]bool            // track declared interfaces for zero-value generation
	sourceFile         string                     // original .rg filename for //line directives
	emitLineDir        bool                       // whether to emit //line directives
	currentReturnTypes []string                   // return types of the current function/method
	contexts           []loopContext              // stack of loop/block contexts
	inMainFunc         bool                       // true when generating code inside main() function
	errors             []error                    // collected errors during generation
	tempVarCounter     int                        // counter for generating unique temp variable names
	typeInfo           TypeInfo                   // optional type info from semantic analysis
}

// addError records an error during code generation
func (g *Generator) addError(err error) {
	g.errors = append(g.errors, err)
}

// returnsError checks if the current function returns error as its last type
func (g *Generator) returnsError() bool {
	if len(g.currentReturnTypes) == 0 {
		return false
	}
	return g.currentReturnTypes[len(g.currentReturnTypes)-1] == "error"
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

// emitLineDirective writes a //line directive for the given source line.
// This maps Go errors/panics back to the original Rugby source location.
func (g *Generator) emitLineDirective(line int) {
	if g.emitLineDir && g.sourceFile != "" && line > 0 {
		g.buf.WriteString(fmt.Sprintf("//line %s:%d\n", g.sourceFile, line))
	}
}

// WithTypeInfo provides type information from semantic analysis for optimization.
// When available, enables direct == comparisons for primitive types instead of
// falling back to runtime.Equal.
func WithTypeInfo(ti TypeInfo) Option {
	return func(g *Generator) {
		g.typeInfo = ti
	}
}

func New(opts ...Option) *Generator {
	g := &Generator{
		vars:        make(map[string]string),
		imports:     make(map[string]bool),
		pubClasses:  make(map[string]bool),
		classFields: make(map[string]string),
		modules:     make(map[string]*ast.ModuleDecl),
		interfaces:  make(map[string]bool),
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

// scopedVar temporarily binds a variable to a type for the duration of fn,
// then restores the previous state. If name is "_", no binding occurs.
func (g *Generator) scopedVar(name, varType string, fn func()) {
	if name == "_" {
		fn()
		return
	}

	prevType, wasDefined := g.vars[name]
	g.vars[name] = varType

	fn()

	if !wasDefined {
		delete(g.vars, name)
	} else {
		g.vars[name] = prevType
	}
}

// scopedVars temporarily binds multiple variables to types for the duration of fn,
// then restores the previous state. Each pair in vars is (name, type).
// Variables named "_" are skipped.
func (g *Generator) scopedVars(vars []struct{ name, varType string }, fn func()) {
	// Save previous state
	type savedVar struct {
		name       string
		prevType   string
		wasDefined bool
	}
	saved := make([]savedVar, 0, len(vars))

	for _, v := range vars {
		if v.name == "_" {
			continue
		}
		prevType, wasDefined := g.vars[v.name]
		saved = append(saved, savedVar{v.name, prevType, wasDefined})
		g.vars[v.name] = v.varType
	}

	fn()

	// Restore previous state
	for _, s := range saved {
		if !s.wasDefined {
			delete(g.vars, s.name)
		} else {
			g.vars[s.name] = s.prevType
		}
	}
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
		if sel, ok := e.Func.(*ast.SelectorExpr); ok {
			// Check for predicate methods ending in ? (must return Bool)
			if strings.HasSuffix(sel.Sel, "?") {
				return "Bool"
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
		// Check for predicate methods ending in ?
		if strings.HasSuffix(e.Sel, "?") {
			return "Bool"
		}
	case *ast.BinaryExpr:
		// Comparison operators return Bool
		switch e.Op {
		case "==", "!=", "<", ">", "<=", ">=":
			return "Bool"
		case "and", "or":
			return "Bool"
		case "+", "-", "*", "/", "%":
			// Arithmetic operators return the type of operands
			leftType := g.inferTypeFromExpr(e.Left)
			if leftType == "Float" || g.inferTypeFromExpr(e.Right) == "Float" {
				return "Float"
			}
			if leftType == "Int" {
				return "Int"
			}
			if leftType == "String" && e.Op == "+" {
				return "String"
			}
		}
	case *ast.UnaryExpr:
		if e.Op == "not" || e.Op == "!" {
			return "Bool"
		}
		// Unary minus returns same type as operand
		if e.Op == "-" {
			return g.inferTypeFromExpr(e.Expr)
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
	case *ast.NilLit:
		return "nil"
	case *ast.TernaryExpr:
		// Return the type of the then branch (or else branch if then is unknown)
		if t := g.inferTypeFromExpr(e.Then); t != "" {
			return t
		}
		return g.inferTypeFromExpr(e.Else)
	}
	return "" // unknown
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
		case *ast.ModuleDecl:
			definitions = append(definitions, d)
		// Test constructs are definitions, not top-level statements
		case *ast.DescribeStmt, *ast.TestStmt, *ast.TableStmt:
			definitions = append(definitions, decl)
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

	// Track user imports to avoid duplicates (use transformed paths)
	userImports := make(map[string]bool)
	for _, imp := range program.Imports {
		userImports[transformImportPath(imp.Path)] = true
	}

	// Collect all imports
	const runtimeImport = "github.com/nchapman/rugby/runtime"
	const testImport = "github.com/nchapman/rugby/test"
	needsRuntimeImport := g.needsRuntime && !userImports[runtimeImport]
	needsFmtImport := g.needsFmt && !userImports["fmt"]
	needsErrorsImport := g.needsErrors && !userImports["errors"]
	needsTestingImport := g.needsTestingImport && !userImports["testing"]
	needsTestImport := g.needsTestImport && !userImports[testImport]
	hasImports := len(program.Imports) > 0 || needsRuntimeImport || needsFmtImport || needsErrorsImport || needsTestingImport || needsTestImport
	if hasImports {
		out.WriteString("import (\n")
		// User-specified imports
		for _, imp := range program.Imports {
			goPath := transformImportPath(imp.Path)
			if imp.Alias != "" {
				out.WriteString(fmt.Sprintf("\t%s %q\n", imp.Alias, goPath))
			} else {
				out.WriteString(fmt.Sprintf("\t%q\n", goPath))
			}
		}
		// Auto-imports (only if not already imported by user)
		if needsErrorsImport {
			out.WriteString("\t\"errors\"\n")
		}
		if needsFmtImport {
			out.WriteString("\t\"fmt\"\n")
		}
		if needsTestingImport {
			out.WriteString("\t\"testing\"\n")
		}
		if needsRuntimeImport {
			out.WriteString(fmt.Sprintf("\t%q\n", runtimeImport))
		}
		if needsTestImport {
			out.WriteString(fmt.Sprintf("\ttest %q\n", testImport))
		}
		out.WriteString(")\n\n")
	}

	out.WriteString(bodyCode)

	// Check for errors collected during generation
	if len(g.errors) > 0 {
		// Combine all errors into one
		var msgs []string
		for _, e := range g.errors {
			msgs = append(msgs, e.Error())
		}
		return "", fmt.Errorf("codegen errors:\n  %s", strings.Join(msgs, "\n  "))
	}

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
