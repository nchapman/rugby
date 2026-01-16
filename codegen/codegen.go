// Package codegen generates Go code from Rugby AST.
package codegen

import (
	"fmt"
	"go/format"
	"slices"
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
	TypeChannel // channel type for send/receive
	TypeAny     // any/interface{} type, from spawn/await
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
	// For example: "Int", "String", "String?", "Array<Int>".
	// Returns empty string if no type info is available.
	GetRugbyType(node ast.Node) string

	// GetSelectorKind returns the kind of a selector expression (field, method, getter, etc.)
	// This is used to determine whether to generate field access or method call syntax.
	// Returns SelectorUnknown if not a SelectorExpr or not yet resolved.
	GetSelectorKind(node ast.Node) ast.SelectorKind

	// GetElementType returns the element type for arrays, channels, optionals, etc.
	// For Array<Int>, returns "Int". For String?, returns "String".
	// Returns empty string if not a composite type or unknown.
	GetElementType(node ast.Node) string

	// GetKeyValueTypes returns the key and value types for maps.
	// For Map<String, Int>, returns ("String", "Int").
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

	// IsStruct returns true if the given type name is a declared struct.
	IsStruct(typeName string) bool

	// IsInterface returns true if the given type name is a declared interface.
	IsInterface(typeName string) bool

	// IsNoArgFunction returns true if the given name is a declared function with no parameters.
	// This is used for implicit function calls (calling functions without parentheses).
	IsNoArgFunction(name string) bool

	// IsPublicClass returns true if the given class name is declared as public (pub class).
	IsPublicClass(className string) bool

	// HasAccessor returns true if the given class field has a getter or setter accessor.
	// This is used to determine if a field needs underscore prefix in the struct.
	HasAccessor(className, fieldName string) bool

	// GetInterfaceMethodNames returns the method names declared in an interface.
	// Returns nil if the interface doesn't exist.
	GetInterfaceMethodNames(interfaceName string) []string

	// GetAllInterfaceNames returns the names of all declared interfaces.
	GetAllInterfaceNames() []string

	// GetAllModuleNames returns the names of all declared modules.
	GetAllModuleNames() []string

	// GetModuleMethodNames returns the method names declared in a module.
	// Returns nil if the module doesn't exist.
	GetModuleMethodNames(moduleName string) []string

	// IsModuleMethod returns true if the method in the given class came from an included module.
	// This is used to determine whether to call self.method() vs method() in class bodies.
	IsModuleMethod(className, methodName string) bool

	// GetConstructorParamCount returns the number of constructor parameters for a class.
	// Returns 0 if the class has no constructor or doesn't exist.
	GetConstructorParamCount(className string) int

	// GetConstructorParams returns the constructor parameter names and types for a class.
	// Returns nil if the class has no constructor or doesn't exist.
	// Each element is a [2]string{name, type}.
	GetConstructorParams(className string) [][2]string
}

type Generator struct {
	buf                          strings.Builder
	indent                       int
	vars                         map[string]string            // track declared variables and their types (empty string = unknown type)
	imports                      map[string]bool              // track import aliases for Go interop detection
	needsRuntime                 bool                         // track if rugby/runtime import is needed
	needsFmt                     bool                         // track if fmt import is needed (string interpolation)
	needsErrors                  bool                         // track if errors import is needed (error_is?, error_as)
	needsRegexp                  bool                         // track if regexp import is needed (regex literals)
	needsConstraints             bool                         // track if constraints import is needed (generics)
	needsTestingImport           bool                         // track if testing import is needed
	needsTestImport              bool                         // track if rugby/test import is needed
	currentClass                 string                       // current class being generated (for instance vars)
	currentOriginalClass         string                       // original class name for module-scoped classes (for semantic analysis lookup)
	currentModule                string                       // current module being generated (for nested class naming)
	currentMethod                string                       // current method name being generated (for super)
	currentMethodPub             bool                         // whether current method is pub (for super)
	currentClassEmbeds           []string                     // embedded types (parent classes) of current class
	currentClassTypeParamClause  string                       // type parameter clause for current class (e.g., "[T any]")
	currentClassTypeParamNames   string                       // type parameter names for current class (e.g., "[T]")
	accessorFields               map[string]bool              // track which fields have accessor methods (need underscore prefix)
	modules                      map[string]*ast.ModuleDecl   // track module definitions for include
	currentClassInterfaceMethods map[string]bool              // methods that must be exported for current class (to satisfy interfaces)
	goInteropVars                map[string]bool              // track variables holding Go interop types
	sourceFile                   string                       // original .rg filename for //line directives
	emitLineDir                  bool                         // whether to emit //line directives
	currentReturnTypes           []string                     // return types of the current function/method
	loopDepth                    int                          // nesting depth of loops (for break/next)
	inInlinedLambda              bool                         // true when inside an inlined lambda (each/times); return -> continue
	inMainFunc                   bool                         // true when generating code inside main() function
	errors                       []error                      // collected errors during generation
	tempVarCounter               int                          // counter for generating unique temp variable names
	typeInfo                     TypeInfo                     // type info from semantic analysis (required)
	baseClasses                  map[string]bool              // track classes that are extended by other classes
	baseClassMethods             map[string][]string          // base class name -> method names (for lint suppression)
	baseClassAccessors           map[string][]string          // base class name -> accessor names (for lint suppression)
	classMethods                 map[string]map[string]string // class name -> method name -> generated function name
	classVars                    map[string]string            // "ClassName@@varname" -> "_ClassName_varname" (package-level var)
	pubFuncs                     map[string]bool              // track functions declared with 'pub' for correct casing
	pubAccessors                 map[string]map[string]bool   // track pub accessors per class: className -> accessorName -> isPub
	privateMethods               map[string]map[string]bool   // track private methods per class: className -> methodName -> isPrivate
	instanceMethods              map[string]map[string]string // track instance methods per class: className -> methodName -> goName
	pubMethods                   map[string]map[string]bool   // track pub methods per class: className -> methodName -> isPub
	classParent                  map[string]string            // track class inheritance: childClass -> parentClass
	enums                        map[string]*ast.EnumDecl     // track enum definitions for expression translation
	structs                      map[string]*ast.StructDecl   // track struct definitions
	currentStruct                *ast.StructDecl              // current struct being generated (for @field translation)
	functions                    map[string]*ast.FuncDecl     // track function declarations for default parameter lookup
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
	lastType := g.currentReturnTypes[len(g.currentReturnTypes)-1]
	return lastType == "error" || lastType == "Error"
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
		accessorFields:               make(map[string]bool),
		modules:                      make(map[string]*ast.ModuleDecl),
		currentClassInterfaceMethods: make(map[string]bool),
		goInteropVars:                make(map[string]bool),
		baseClasses:                  make(map[string]bool),
		baseClassMethods:             make(map[string][]string),
		baseClassAccessors:           make(map[string][]string),
		classMethods:                 make(map[string]map[string]string),
		classVars:                    make(map[string]string),
		pubFuncs:                     make(map[string]bool),
		pubAccessors:                 make(map[string]map[string]bool),
		privateMethods:               make(map[string]map[string]bool),
		instanceMethods:              make(map[string]map[string]string),
		pubMethods:                   make(map[string]map[string]bool),
		classParent:                  make(map[string]string),
		functions:                    make(map[string]*ast.FuncDecl),
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
		panic("codegen: typeInfo is required - run semantic analysis before code generation")
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
// in Go to avoid conflict with getter/setter methods. Requires typeInfo from semantic analysis.
func (g *Generator) isAccessorField(fieldName string) bool {
	if g.typeInfo == nil {
		panic("codegen: typeInfo is required - run semantic analysis before code generation")
	}
	// Check if this field is in the current class's accessorFields map
	// (this is populated during class generation and handles module-scoped classes correctly)
	if g.accessorFields[fieldName] {
		return true
	}
	// Check current class in semantic analysis
	// For module-scoped classes, use the original class name
	classToCheck := g.currentClass
	if g.currentOriginalClass != "" {
		classToCheck = g.currentOriginalClass
	}
	if g.typeInfo.HasAccessor(classToCheck, fieldName) {
		return true
	}
	// Check parent classes (through embedding)
	for _, embed := range g.currentClassEmbeds {
		if g.typeInfo.HasAccessor(embed, fieldName) {
			return true
		}
	}
	return false
}

// isParentAccessor checks if a field name is an accessor in any parent class
// (but NOT in the current class). Used to skip field generation for inherited accessors.
func (g *Generator) isParentAccessor(embeds []string, fieldName string) bool {
	// Walk up the inheritance chain
	for _, parentClass := range embeds {
		if g.isAccessorInClassHierarchy(parentClass, fieldName) {
			return true
		}
	}
	return false
}

// isAccessorInClassHierarchy checks if a field is an accessor in the given class
// or any of its parent classes.
func (g *Generator) isAccessorInClassHierarchy(className, fieldName string) bool {
	// Check current class
	if g.typeInfo != nil && g.typeInfo.HasAccessor(className, fieldName) {
		return true
	}
	// Also check baseClassAccessors (populated during pre-pass)
	if slices.Contains(g.baseClassAccessors[className], fieldName) {
		return true
	}
	// Check parent class
	if parent := g.classParent[className]; parent != "" {
		return g.isAccessorInClassHierarchy(parent, fieldName)
	}
	return false
}

// getInheritedFieldPath returns the path to an inherited field (e.g., "Parent._name")
// if the field is from a parent class. Returns empty string if the field is local.
func (g *Generator) getInheritedFieldPath(fieldName string) string {
	// Check if this field is an accessor in a parent class
	if !g.isParentAccessor(g.currentClassEmbeds, fieldName) {
		return "" // Field is local, no path needed
	}

	// Find which parent class has this accessor and build the path
	return g.findFieldPath(g.currentClassEmbeds, fieldName)
}

// findFieldPath recursively finds the path to a field in the class hierarchy
func (g *Generator) findFieldPath(embeds []string, fieldName string) string {
	for _, parentClass := range embeds {
		// Check if this parent class has the accessor
		hasAccessor := false
		if g.typeInfo != nil && g.typeInfo.HasAccessor(parentClass, fieldName) {
			hasAccessor = true
		}
		if slices.Contains(g.baseClassAccessors[parentClass], fieldName) {
			hasAccessor = true
		}

		if hasAccessor {
			// Found it in this parent class - return path with underscore prefix
			return parentClass + "._" + fieldName
		}

		// Check this parent's parents (recursive)
		if grandparent := g.classParent[parentClass]; grandparent != "" {
			if path := g.findFieldPath([]string{grandparent}, fieldName); path != "" {
				// Prepend the parent class name
				return parentClass + "." + path
			}
		}
	}
	return ""
}

// getFieldType returns the Rugby type of a class field. Requires typeInfo from semantic analysis.
func (g *Generator) getFieldType(fieldName string) string {
	if g.typeInfo == nil {
		panic("codegen: typeInfo is required - run semantic analysis before code generation")
	}
	return g.typeInfo.GetFieldType(g.currentClass, fieldName)
}

// isClass checks if a type name is a declared class. Requires typeInfo from semantic analysis.
func (g *Generator) isClass(typeName string) bool {
	if g.typeInfo == nil {
		panic("codegen: typeInfo is required - run semantic analysis before code generation")
	}
	return g.typeInfo.IsClass(typeName)
}

// isStruct checks if a type name is a declared struct. Requires typeInfo from semantic analysis.
func (g *Generator) isStruct(typeName string) bool {
	if g.typeInfo == nil {
		panic("codegen: typeInfo is required - run semantic analysis before code generation")
	}
	return g.typeInfo.IsStruct(typeName)
}

// stripTypeParams removes generic type parameters from a type name.
// For example: "Box<Int>" -> "Box", "User" -> "User"
func stripTypeParams(typeName string) string {
	if idx := strings.Index(typeName, "<"); idx != -1 {
		return typeName[:idx]
	}
	return typeName
}

// isModuleScopedClass checks if a type name is a module-scoped class name (e.g., "Http_Response")
// by checking if it matches the pattern Module_Class where Module exists and has Class defined.
func (g *Generator) isModuleScopedClass(typeName string) bool {
	// Check for underscore separator
	idx := strings.Index(typeName, "_")
	if idx == -1 {
		return false
	}

	moduleName := typeName[:idx]
	className := typeName[idx+1:]

	mod := g.modules[moduleName]
	if mod == nil {
		return false
	}

	for _, cls := range mod.Classes {
		if cls.Name == className {
			return true
		}
	}
	return false
}

// isInterface checks if a type name is a declared interface. Requires typeInfo from semantic analysis.
func (g *Generator) isInterface(typeName string) bool {
	if g.typeInfo == nil {
		panic("codegen: typeInfo is required - run semantic analysis before code generation")
	}
	return g.typeInfo.IsInterface(typeName)
}

// isNoArgFunction checks if a name is a declared function with no parameters.
// Requires typeInfo from semantic analysis.
func (g *Generator) isNoArgFunction(name string) bool {
	if g.typeInfo == nil {
		panic("codegen: typeInfo is required - run semantic analysis before code generation")
	}
	return g.typeInfo.IsNoArgFunction(name)
}

// isPublicClass checks if a class name is declared as public. Requires typeInfo from semantic analysis.
func (g *Generator) isPublicClass(className string) bool {
	if g.typeInfo == nil {
		panic("codegen: typeInfo is required - run semantic analysis before code generation")
	}
	return g.typeInfo.IsPublicClass(className)
}

// getInterfaceMethodNames returns the method names declared in an interface.
// Requires typeInfo from semantic analysis.
func (g *Generator) getInterfaceMethodNames(interfaceName string) []string {
	if g.typeInfo == nil {
		panic("codegen: typeInfo is required - run semantic analysis before code generation")
	}
	return g.typeInfo.GetInterfaceMethodNames(interfaceName)
}

// getConstructorParams returns the constructor parameter names and types for a class.
// Requires typeInfo from semantic analysis.
func (g *Generator) getConstructorParams(className string) [][2]string {
	if g.typeInfo == nil {
		panic("codegen: typeInfo is required - run semantic analysis before code generation")
	}
	return g.typeInfo.GetConstructorParams(className)
}

// getAllInterfaceNames returns the names of all declared interfaces.
// Requires typeInfo from semantic analysis.
func (g *Generator) getAllInterfaceNames() []string {
	if g.typeInfo == nil {
		panic("codegen: typeInfo is required - run semantic analysis before code generation")
	}
	return g.typeInfo.GetAllInterfaceNames()
}

// getAllModuleNames returns the names of all declared modules.
// Requires typeInfo from semantic analysis.
func (g *Generator) getAllModuleNames() []string {
	if g.typeInfo == nil {
		panic("codegen: typeInfo is required - run semantic analysis before code generation")
	}
	return g.typeInfo.GetAllModuleNames()
}

// getModuleMethodNames returns the method names declared in a module.
// Requires typeInfo from semantic analysis.
func (g *Generator) getModuleMethodNames(moduleName string) []string {
	if g.typeInfo == nil {
		panic("codegen: typeInfo is required - run semantic analysis before code generation")
	}
	return g.typeInfo.GetModuleMethodNames(moduleName)
}

// isModuleMethod checks if a method in the current class came from an included module.
// Returns false if not in a class context or if typeInfo is not available.
func (g *Generator) isModuleMethod(methodName string) bool {
	if g.typeInfo == nil || g.currentClass == "" {
		return false
	}
	return g.typeInfo.IsModuleMethod(g.currentClass, methodName)
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
	if goType := g.typeInfo.GetGoType(expr); strings.HasPrefix(goType, "[]") {
		return goType[2:] // Remove "[]" prefix
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

// inferExprGoType returns the Go type for an expression.
// Uses semantic type info with mapType conversion, falls back to "any".
func (g *Generator) inferExprGoType(expr ast.Expression) string {
	if goType := g.typeInfo.GetGoType(expr); goType != "" && goType != "unknown" {
		return goType
	}
	// Fall back to Rugby type inference and mapping
	if rugbyType := g.inferTypeFromExpr(expr); rugbyType != "" {
		return mapType(rugbyType)
	}
	return "any"
}

// getSelectorKind returns the resolved selector kind for a SelectorExpr.
// First checks the AST annotation, then falls back to TypeInfo, then to heuristics.
func (g *Generator) getSelectorKind(sel *ast.SelectorExpr) ast.SelectorKind {
	// First check the AST annotation - trust Getter/Method but verify Field
	// SelectorField may be incorrect for inherited getters that semantic doesn't detect
	if sel.ResolvedKind == ast.SelectorGetter || sel.ResolvedKind == ast.SelectorMethod || sel.ResolvedKind == ast.SelectorGoMethod {
		return sel.ResolvedKind
	}

	// Then check TypeInfo, but only trust it for Getter/Method kinds
	if kind := g.typeInfo.GetSelectorKind(sel); kind == ast.SelectorGetter || kind == ast.SelectorMethod || kind == ast.SelectorGoMethod {
		return kind
	}

	// Fall back to heuristics (also verify SelectorField annotations):
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
	// - Instance methods/getters on known classes (including inherited)
	if receiverClassName := g.getReceiverClassName(sel.X); receiverClassName != "" {
		if g.isClassInstanceMethod(receiverClassName, sel.Sel) != "" {
			return ast.SelectorGetter
		}
	}

	return ast.SelectorUnknown
}

// inferTypeFromExpr returns the Rugby type for an expression.
// Primary source is TypeInfo from semantic analysis. Fallback is minimal
// literal type detection for robustness.
func (g *Generator) inferTypeFromExpr(expr ast.Expression) string {
	// Primary: use semantic type info
	// GetRugbyType returns the full type including generics (e.g., "Array<Int>", "String?")
	// Note: "any" is a valid type that should be returned
	if rugbyType := g.typeInfo.GetRugbyType(expr); rugbyType != "" && rugbyType != "unknown" {
		return rugbyType
	}

	// Minimal fallback: literal type detection only
	// Semantic analysis should provide types for all other expressions
	switch e := expr.(type) {
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
	case *ast.SetLit:
		return "Set"
	case *ast.CallExpr:
		// Constructor call: ClassName.new(...) -> ClassName
		// Note: Doesn't handle chained calls or explicit type params yet
		if sel, ok := e.Func.(*ast.SelectorExpr); ok {
			if sel.Sel == "new" {
				if ident, ok := sel.X.(*ast.Ident); ok {
					return ident.Name
				}
			}
			// Struct method call: p.moved(...) -> look up method return type
			receiverType := g.inferTypeFromExpr(sel.X)
			if g.structs != nil {
				if structDecl, ok := g.structs[receiverType]; ok {
					for _, method := range structDecl.Methods {
						if method.Name == sel.Sel && len(method.ReturnTypes) > 0 {
							return method.ReturnTypes[0]
						}
					}
				}
			}
		}
	case *ast.ScopeExpr:
		// Enum value reference: Status::Active -> "Status"
		if ident, ok := e.Left.(*ast.Ident); ok {
			if g.enums != nil {
				if _, isEnum := g.enums[ident.Name]; isEnum {
					return ident.Name
				}
			}
		}
	case *ast.StructLit:
		// Struct literal: Point{x: 10, y: 20} -> "Point"
		return e.Name
	case *ast.Ident:
		// Variable lookup - check vars map
		if varType, ok := g.vars[e.Name]; ok && varType != "" {
			return varType
		}
	case *ast.BinaryExpr:
		// For arithmetic expressions, infer numeric type
		leftType := g.inferTypeFromExpr(e.Left)
		rightType := g.inferTypeFromExpr(e.Right)
		switch e.Op {
		case "+", "-", "*", "/", "%":
			// If either is Float, result is Float
			if leftType == "Float" || rightType == "Float" {
				return "Float"
			}
			// If both are Int (or can be Int), result is Int
			if leftType == "Int" || rightType == "Int" {
				return "Int"
			}
		case "<", ">", "<=", ">=", "==", "!=", "&&", "||":
			return "Bool"
		}
	case *ast.SelectorExpr:
		// For struct field access, look up the field type
		receiverType := g.inferTypeFromExpr(e.X)
		if g.structs != nil {
			if structDecl, ok := g.structs[receiverType]; ok {
				for _, field := range structDecl.Fields {
					if field.Name == e.Sel {
						return field.Type
					}
				}
			}
		}
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
		case *ast.TypeAliasDecl:
			definitions = append(definitions, d)
		case *ast.ConstDecl:
			definitions = append(definitions, d)
		case *ast.EnumDecl:
			definitions = append(definitions, d)
		case *ast.StructDecl:
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

	// Pre-pass: track base classes (classes that are extended by others)
	// This is used to generate method reference assertions that suppress
	// "unused method" lint warnings for methods in base classes that are
	// shadowed by subclasses.
	// Also track the inheritance hierarchy for method lookup.
	for _, def := range definitions {
		if cls, ok := def.(*ast.ClassDecl); ok {
			for _, embed := range cls.Embeds {
				g.baseClasses[embed] = true
				// Track child -> parent relationship (only single inheritance)
				if len(cls.Embeds) == 1 {
					g.classParent[cls.Name] = embed
				}
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

	// Pre-pass: collect public functions (declared with 'pub')
	// This is needed so function calls can use the correct casing
	for _, def := range definitions {
		if fn, ok := def.(*ast.FuncDecl); ok && fn.Pub {
			g.pubFuncs[fn.Name] = true
		}
	}

	// Pre-pass: collect public accessors per class
	// This is needed so accessor calls can use the correct casing
	for _, def := range definitions {
		if cls, ok := def.(*ast.ClassDecl); ok {
			for _, acc := range cls.Accessors {
				if acc.Pub || cls.Pub {
					// Accessor is exported if it has 'pub' or the class is 'pub'
					if g.pubAccessors[cls.Name] == nil {
						g.pubAccessors[cls.Name] = make(map[string]bool)
					}
					g.pubAccessors[cls.Name][acc.Name] = true
				}
			}
		}
	}

	// Pre-pass: collect private methods per class
	// This is needed so method calls can use the correct name (underscore prefix)
	for _, def := range definitions {
		if cls, ok := def.(*ast.ClassDecl); ok {
			for _, method := range cls.Methods {
				if method.Private {
					if g.privateMethods[cls.Name] == nil {
						g.privateMethods[cls.Name] = make(map[string]bool)
					}
					g.privateMethods[cls.Name][method.Name] = true
				}
			}
		}
	}

	// Pre-pass: collect all instance methods per class
	// This is needed to resolve implicit method calls (validate -> u.validate())
	// Note: We can't check currentClassInterfaceMethods here since it's populated per-class during generation.
	// Interface methods will be handled separately via structural typing during genClassDecl.
	for _, def := range definitions {
		if cls, ok := def.(*ast.ClassDecl); ok {
			for _, method := range cls.Methods {
				if method.IsClassMethod {
					continue // Skip class methods (def self.method_name)
				}
				if g.instanceMethods[cls.Name] == nil {
					g.instanceMethods[cls.Name] = make(map[string]string)
				}
				// Determine the Go method name based on visibility
				// Note: pub takes precedence over private (invalid combo is caught by semantic analyzer)
				// Special: setter methods (name=) -> SetName or setName
				var goName string
				isSetter := strings.HasSuffix(method.Name, "=")
				baseName := strings.TrimSuffix(method.Name, "=")

				if method.Pub {
					if isSetter {
						goName = "Set" + snakeToPascalWithAcronyms(baseName)
					} else {
						goName = snakeToPascalWithAcronyms(method.Name)
					}
				} else if method.Private {
					if isSetter {
						goName = "_set" + snakeToPascalWithAcronyms(baseName)
					} else {
						goName = "_" + snakeToCamelWithAcronyms(method.Name)
					}
				} else {
					if isSetter {
						goName = "set" + snakeToPascalWithAcronyms(baseName)
					} else {
						goName = snakeToCamelWithAcronyms(method.Name)
					}
				}
				g.instanceMethods[cls.Name][method.Name] = goName

				// Track pub methods for correct casing at call sites
				if method.Pub {
					if g.pubMethods[cls.Name] == nil {
						g.pubMethods[cls.Name] = make(map[string]bool)
					}
					g.pubMethods[cls.Name][method.Name] = true
				}
			}

			// Also collect accessor getters as instance methods
			// This allows genSelectorExpr to know when to add () for getter calls
			for _, acc := range cls.Accessors {
				if acc.Kind == "getter" || acc.Kind == "property" {
					if g.instanceMethods[cls.Name] == nil {
						g.instanceMethods[cls.Name] = make(map[string]string)
					}
					// Determine Go name based on visibility
					var goName string
					if acc.Pub || cls.Pub {
						goName = snakeToPascalWithAcronyms(acc.Name)
					} else {
						goName = snakeToCamelWithAcronyms(acc.Name)
					}
					g.instanceMethods[cls.Name][acc.Name] = goName
				}
			}
		}
	}

	// Pre-pass: collect instance methods for nested classes inside modules
	// The semantic analyzer uses the original class name (e.g., "Response"),
	// so we register methods under that name for type lookups to work correctly.
	for _, def := range definitions {
		if mod, ok := def.(*ast.ModuleDecl); ok {
			for _, cls := range mod.Classes {
				// Register instance methods under the original class name
				for _, method := range cls.Methods {
					if method.IsClassMethod {
						continue
					}
					if g.instanceMethods[cls.Name] == nil {
						g.instanceMethods[cls.Name] = make(map[string]string)
					}
					var goName string
					isSetter := strings.HasSuffix(method.Name, "=")
					baseName := strings.TrimSuffix(method.Name, "=")

					if method.Pub {
						if isSetter {
							goName = "Set" + snakeToPascalWithAcronyms(baseName)
						} else {
							goName = snakeToPascalWithAcronyms(method.Name)
						}
					} else if method.Private {
						if isSetter {
							goName = "_set" + snakeToPascalWithAcronyms(baseName)
						} else {
							goName = "_" + snakeToCamelWithAcronyms(method.Name)
						}
					} else {
						if isSetter {
							goName = "set" + snakeToPascalWithAcronyms(baseName)
						} else {
							goName = snakeToCamelWithAcronyms(method.Name)
						}
					}
					g.instanceMethods[cls.Name][method.Name] = goName
				}

				// Register accessor getters under the original class name
				for _, acc := range cls.Accessors {
					if acc.Kind == "getter" || acc.Kind == "property" {
						if g.instanceMethods[cls.Name] == nil {
							g.instanceMethods[cls.Name] = make(map[string]string)
						}
						var goName string
						if acc.Pub || cls.Pub {
							goName = snakeToPascalWithAcronyms(acc.Name)
						} else {
							goName = snakeToCamelWithAcronyms(acc.Name)
						}
						g.instanceMethods[cls.Name][acc.Name] = goName
					}
				}
			}
		}
	}

	// Pre-pass: collect instance methods for structs
	// Structs are value types with all methods always public (PascalCase)
	for _, def := range definitions {
		if structDecl, ok := def.(*ast.StructDecl); ok {
			for _, method := range structDecl.Methods {
				if g.instanceMethods[structDecl.Name] == nil {
					g.instanceMethods[structDecl.Name] = make(map[string]string)
				}
				// Struct methods are always public (PascalCase)
				goName := snakeToPascalWithAcronyms(method.Name)
				g.instanceMethods[structDecl.Name][method.Name] = goName
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
	const constraintsImport = "golang.org/x/exp/constraints"
	needsRuntimeImport := g.needsRuntime && !userImports[runtimeImport]
	needsFmtImport := g.needsFmt && !userImports["fmt"]
	needsErrorsImport := g.needsErrors && !userImports["errors"]
	needsRegexpImport := g.needsRegexp && !userImports["regexp"]
	needsTestingImport := g.needsTestingImport && !userImports["testing"]
	needsTestImport := g.needsTestImport && !userImports[testImport]
	needsConstraintsImport := g.needsConstraints && !userImports[constraintsImport]
	hasImports := len(program.Imports) > 0 || needsRuntimeImport || needsFmtImport || needsErrorsImport || needsRegexpImport || needsTestingImport || needsTestImport || needsConstraintsImport
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
		if needsRegexpImport {
			out.WriteString("\t\"regexp\"\n")
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
		if needsConstraintsImport {
			out.WriteString(fmt.Sprintf("\t%q\n", constraintsImport))
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
