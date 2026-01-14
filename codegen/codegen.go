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
	TypeOptional
	TypeAny // any/interface{} type, from spawn/await
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

	// GetRugbyType returns the Rugby type string for an AST node.
	// For example: "Int", "String", "String?", "Array[Int]".
	// Returns empty string if no type info is available.
	GetRugbyType(node ast.Node) string

	// GetSelectorKind returns the kind of a selector expression (field, method, getter, etc.)
	// This is used to determine whether to generate field access or method call syntax.
	// Returns SelectorUnknown if not a SelectorExpr or not yet resolved.
	GetSelectorKind(node ast.Node) ast.SelectorKind

	// GetElementType returns the element type for arrays, channels, optionals, etc.
	// For Array[Int], returns "Int". For String?, returns "String".
	// Returns empty string if not a composite type or unknown.
	GetElementType(node ast.Node) string

	// GetKeyValueTypes returns the key and value types for maps.
	// For Map[String, Int], returns ("String", "Int").
	// Returns empty strings if not a map or unknown.
	GetKeyValueTypes(node ast.Node) (keyType, valueType string)

	// GetTupleTypes returns the element types for tuple/multi-value expressions.
	// For a function returning (Int, error), returns ["Int", "error"].
	// Returns nil if not a tuple type or unknown.
	GetTupleTypes(node ast.Node) []string

	// IsVariableUsedAt checks if a variable declared at a specific AST node is used.
	// This is used to replace unused variables with _ in multi-value assignments.
	IsVariableUsedAt(node ast.Node, name string) bool

	// IsDeclaration returns true if this AST node declares a new variable.
	// This is used to determine whether to use := (declaration) or = (assignment).
	// Works for AssignStmt, MultiAssignStmt, ForStmt, if-let statements, etc.
	IsDeclaration(node ast.Node) bool

	// GetFieldType returns the Rugby type of a class field by class and field name.
	// Returns empty string if field not found.
	GetFieldType(className, fieldName string) string

	// IsClass returns true if the given type name is a declared class.
	IsClass(typeName string) bool

	// IsInterface returns true if the given type name is a declared interface.
	IsInterface(typeName string) bool
}

type Generator struct {
	buf                          strings.Builder
	indent                       int
	vars                         map[string]string            // track declared variables and their types (empty string = unknown type)
	imports                      map[string]bool              // track import aliases for Go interop detection
	needsRuntime                 bool                         // track if rugby/runtime import is needed
	needsFmt                     bool                         // track if fmt import is needed (string interpolation)
	needsErrors                  bool                         // track if errors import is needed (error_is?, error_as)
	needsTestingImport           bool                         // track if testing import is needed
	needsTestImport              bool                         // track if rugby/test import is needed
	currentClass                 string                       // current class being generated (for instance vars)
	currentMethod                string                       // current method name being generated (for super)
	currentMethodPub             bool                         // whether current method is pub (for super)
	currentClassEmbeds           []string                     // embedded types (parent classes) of current class
	currentClassModuleMethods    map[string]bool              // methods from included modules (need self.method() call)
	pubClasses                   map[string]bool              // track public classes for constructor naming
	classes                      map[string]bool              // track all class names for pointer type mapping
	accessorFields               map[string]bool              // track which fields have accessor methods (need underscore prefix)
	classAccessorFields          map[string]map[string]bool   // class name -> accessor field names (for subclass access)
	modules                      map[string]*ast.ModuleDecl   // track module definitions for include
	interfaces                   map[string]bool              // track declared interfaces for zero-value generation
	interfaceMethods             map[string]map[string]bool   // interface name -> method names (for interface implementation)
	currentClassInterfaceMethods map[string]bool              // methods that must be exported for current class (to satisfy interfaces)
	classConstructorParams       map[string][]*ast.Param      // track constructor parameters for each class (for subclass delegation)
	noArgFunctions               map[string]bool              // track no-arg top-level functions (for implicit call)
	goInteropVars                map[string]bool              // track variables holding Go interop types
	sourceFile                   string                       // original .rg filename for //line directives
	emitLineDir                  bool                         // whether to emit //line directives
	currentReturnTypes           []string                     // return types of the current function/method
	loopDepth                    int                          // nesting depth of loops (for break/next)
	inMainFunc                   bool                         // true when generating code inside main() function
	errors                       []error                      // collected errors during generation
	tempVarCounter               int                          // counter for generating unique temp variable names
	typeInfo                     TypeInfo                     // optional type info from semantic analysis
	baseClasses                  map[string]bool              // track classes that are extended by other classes
	baseClassMethods             map[string][]string          // base class name -> method names (for lint suppression)
	baseClassAccessors           map[string][]string          // base class name -> accessor names (for lint suppression)
	classMethods                 map[string]map[string]string // class name -> method name -> generated function name
}

// addError records an error during code generation
func (g *Generator) addError(err error) {
	g.errors = append(g.errors, err)
}

// Errors returns all errors collected during code generation
func (g *Generator) Errors() []error {
	return g.errors
}

// returnsError checks if the current function returns error as its last type
func (g *Generator) returnsError() bool {
	if len(g.currentReturnTypes) == 0 {
		return false
	}
	return g.currentReturnTypes[len(g.currentReturnTypes)-1] == "error"
}

func (g *Generator) enterLoop() {
	g.loopDepth++
}

func (g *Generator) exitLoop() {
	if g.loopDepth > 0 {
		g.loopDepth--
	}
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
		vars:                         make(map[string]string),
		imports:                      make(map[string]bool),
		pubClasses:                   make(map[string]bool),
		accessorFields:               make(map[string]bool),
		currentClassModuleMethods:    make(map[string]bool),
		classes:                      make(map[string]bool),
		classAccessorFields:          make(map[string]map[string]bool),
		modules:                      make(map[string]*ast.ModuleDecl),
		interfaces:                   make(map[string]bool),
		interfaceMethods:             make(map[string]map[string]bool),
		currentClassInterfaceMethods: make(map[string]bool),
		classConstructorParams:       make(map[string][]*ast.Param),
		noArgFunctions:               make(map[string]bool),
		goInteropVars:                make(map[string]bool),
		baseClasses:                  make(map[string]bool),
		baseClassMethods:             make(map[string][]string),
		baseClassAccessors:           make(map[string][]string),
		classMethods:                 make(map[string]map[string]string),
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// shouldDeclare returns true if this AST node should use := (declaration)
// and false if it should use = (assignment). Requires typeInfo from semantic analysis.
func (g *Generator) shouldDeclare(node ast.Node) bool {
	if g.typeInfo == nil {
		// Default to declaration if no type info - Go will error on redeclaration
		return true
	}
	return g.typeInfo.IsDeclaration(node)
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

// isAccessorField checks if a field name is an accessor field in the current class
// or any parent class (through embedding). Accessor fields use underscore prefix
// in Go to avoid conflict with getter/setter methods.
func (g *Generator) isAccessorField(fieldName string) bool {
	// Check current class's accessor fields
	if g.accessorFields[fieldName] {
		return true
	}
	// Check parent classes (through embedding)
	for _, embed := range g.currentClassEmbeds {
		if parentAccessors, ok := g.classAccessorFields[embed]; ok {
			if parentAccessors[fieldName] {
				return true
			}
		}
	}
	return false
}

// getFieldType returns the Rugby type of a class field. Requires typeInfo from semantic analysis.
func (g *Generator) getFieldType(fieldName string) string {
	if g.typeInfo == nil {
		return ""
	}
	return g.typeInfo.GetFieldType(g.currentClass, fieldName)
}

// isClass checks if a type name is a declared class.
// Uses TypeInfo when available, falls back to local classes map.
func (g *Generator) isClass(typeName string) bool {
	if g.typeInfo != nil {
		return g.typeInfo.IsClass(typeName)
	}
	return g.classes[typeName]
}

// isInterface checks if a type name is a declared interface.
// Uses TypeInfo when available, falls back to local interfaces map.
func (g *Generator) isInterface(typeName string) bool {
	if g.typeInfo != nil {
		return g.typeInfo.IsInterface(typeName)
	}
	return g.interfaces[typeName]
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

// inferArrayElementGoType returns the Go element type for an array/slice expression.
// For example, if expr evaluates to []int, this returns "int".
// Returns "any" if the element type cannot be determined.
func (g *Generator) inferArrayElementGoType(expr ast.Expression) string {
	if g.typeInfo != nil {
		goType := g.typeInfo.GetGoType(expr)
		if strings.HasPrefix(goType, "[]") {
			return goType[2:] // Remove "[]" prefix
		}
	}
	// Fall back to AST-based inference
	switch e := expr.(type) {
	case *ast.ArrayLit:
		if len(e.Elements) > 0 {
			return mapType(g.inferTypeFromExpr(e.Elements[0]))
		}
	case *ast.Ident:
		// Check if we have the variable type
		if typ, ok := g.vars[e.Name]; ok {
			goType := mapType(typ)
			if strings.HasPrefix(goType, "[]") {
				return goType[2:]
			}
		}
	}
	return "any"
}

// getSelectorKind returns the resolved selector kind for a SelectorExpr.
// First checks the AST annotation, then falls back to TypeInfo, then to heuristics.
func (g *Generator) getSelectorKind(sel *ast.SelectorExpr) ast.SelectorKind {
	// First check the AST annotation
	if sel.ResolvedKind != ast.SelectorUnknown {
		return sel.ResolvedKind
	}

	// Then check TypeInfo
	if g.typeInfo != nil {
		kind := g.typeInfo.GetSelectorKind(sel)
		if kind != ast.SelectorUnknown {
			return kind
		}
	}

	// Fall back to heuristics:
	// - Interface methods need method call syntax
	if g.isInterfaceMethod(sel.Sel) {
		return ast.SelectorMethod
	}
	// - to_s is always a method (maps to String())
	if sel.Sel == "to_s" {
		return ast.SelectorMethod
	}
	// - Go interop variables (like http.Response from http.get)
	//   Property-style access like resp.json should become resp.JSON()
	//   But NOT for direct package member access like io.EOF (which could be a variable)
	if ident, ok := sel.X.(*ast.Ident); ok && g.goInteropVars[ident.Name] {
		return ast.SelectorGoMethod
	}

	return ast.SelectorUnknown
}

// inferTypeFromExpr returns the Rugby type for an expression.
// Primary source is TypeInfo from semantic analysis. Fallback is minimal
// literal type detection for robustness.
func (g *Generator) inferTypeFromExpr(expr ast.Expression) string {
	// Primary: use semantic type info
	if g.typeInfo != nil {
		// GetRugbyType returns the full type including generics (e.g., "Array[Int]", "String?")
		// Note: "any" is a valid type that should be returned
		if rugbyType := g.typeInfo.GetRugbyType(expr); rugbyType != "" && rugbyType != "unknown" {
			return rugbyType
		}
	}

	// Minimal fallback: literal type detection only
	// Semantic analysis should provide types for all other expressions
	switch expr.(type) {
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
	case *ast.NilLit:
		return "nil"
	case *ast.RangeLit:
		return "Range"
	case *ast.ArrayLit:
		return "Array"
	case *ast.MapLit:
		return "Map"
	}
	return "" // unknown - semantic analysis should have provided the type
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

	// Pre-pass: collect interface methods before generating code
	// This ensures classes know which methods need to be exported to satisfy interfaces,
	// even if interfaces are declared after classes in source order
	for _, def := range definitions {
		if iface, ok := def.(*ast.InterfaceDecl); ok {
			if g.interfaceMethods[iface.Name] == nil {
				g.interfaceMethods[iface.Name] = make(map[string]bool)
			}
			for _, method := range iface.Methods {
				g.interfaceMethods[iface.Name][method.Name] = true
			}
		}
	}

	// Pre-pass: collect no-arg functions for implicit call syntax
	// In Rugby, calling a no-arg function without parentheses is allowed (like Ruby)
	for _, def := range definitions {
		if fn, ok := def.(*ast.FuncDecl); ok {
			if len(fn.Params) == 0 {
				g.noArgFunctions[fn.Name] = true
			}
		}
	}

	// Pre-pass: collect class info before generating code
	// This ensures subclasses can access parent constructor params and accessor fields,
	// even if the parent class is declared after the subclass in source order
	for _, def := range definitions {
		if cls, ok := def.(*ast.ClassDecl); ok {
			// Track class name
			g.classes[cls.Name] = true
			g.pubClasses[cls.Name] = cls.Pub

			// Collect constructor parameters
			for _, method := range cls.Methods {
				if method.Name == "initialize" {
					g.classConstructorParams[cls.Name] = method.Params
					break
				}
			}

			// Collect accessor fields
			accessorFields := make(map[string]bool)
			for _, acc := range cls.Accessors {
				accessorFields[acc.Name] = true
			}
			// Also track fields that have methods with the same name
			fieldNames := make(map[string]bool)
			for _, f := range cls.Fields {
				fieldNames[f.Name] = true
			}
			for _, acc := range cls.Accessors {
				fieldNames[acc.Name] = true
			}
			for _, m := range cls.Methods {
				if fieldNames[m.Name] {
					accessorFields[m.Name] = true
				}
			}
			g.classAccessorFields[cls.Name] = accessorFields
		}
	}

	// Propagate accessor fields through inheritance chain
	// This ensures multi-level inheritance works correctly
	for _, def := range definitions {
		if cls, ok := def.(*ast.ClassDecl); ok {
			for _, embed := range cls.Embeds {
				if parentAccessors, ok := g.classAccessorFields[embed]; ok {
					for name := range parentAccessors {
						g.classAccessorFields[cls.Name][name] = true
					}
				}
			}
		}
	}

	// Pre-pass: track base classes (classes that are extended by others)
	// This is used to generate method reference assertions that suppress
	// "unused method" lint warnings for methods in base classes that are
	// shadowed by subclasses.
	for _, def := range definitions {
		if cls, ok := def.(*ast.ClassDecl); ok {
			for _, embed := range cls.Embeds {
				g.baseClasses[embed] = true
			}
		}
	}

	// Pre-pass: collect methods and accessors for base classes
	// These are used to generate var _ = (*BaseClass).methodName assertions
	for _, def := range definitions {
		if cls, ok := def.(*ast.ClassDecl); ok {
			if g.baseClasses[cls.Name] {
				// Collect method names (skip initialize - it's the constructor)
				for _, m := range cls.Methods {
					if m.Name != "initialize" {
						g.baseClassMethods[cls.Name] = append(g.baseClassMethods[cls.Name], m.Name)
					}
				}
				// Collect accessor names (getters generate methods that may be shadowed)
				for _, acc := range cls.Accessors {
					if acc.Kind == "getter" || acc.Kind == "property" {
						g.baseClassAccessors[cls.Name] = append(g.baseClassAccessors[cls.Name], acc.Name)
					}
				}
			}
		}
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
