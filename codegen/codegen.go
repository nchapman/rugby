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

	// GetGoName returns the actual Go name for a Go interop method/function call.
	// For example, if Rugby code calls `hasher.sum(nil)` and the Go method is `Sum`,
	// this returns "Sum". Returns empty string if not a Go interop call or not found.
	GetGoName(node ast.Node) string
}

// ClassContext holds state for the current class being generated.
// This is set when entering genClassDecl and cleared when exiting.
type ClassContext struct {
	Name             string          // class name (e.g., "User")
	OriginalName     string          // original name for module-scoped classes (for semantic lookups)
	Embeds           []string        // parent classes (embedded types)
	TypeParamClause  string          // type parameter clause (e.g., "[T any]")
	TypeParamNames   string          // type parameter names (e.g., "[T]")
	InterfaceMethods map[string]bool // methods that must be exported to satisfy interfaces
	AccessorFields   map[string]bool // fields with accessor methods (need underscore prefix)
}

// MethodContext holds state for the current method/function being generated.
// This is set when entering genMethodDecl/genFuncDecl and cleared when exiting.
type MethodContext struct {
	Name        string            // method name (e.g., "initialize")
	IsPub       bool              // whether method is declared with 'pub'
	ReturnTypes []string          // return types of the function/method
	TypeParams  map[string]string // type parameters (name -> constraint)
}

type Generator struct {
	buf    strings.Builder
	indent int

	// Context: tracks current class and method being generated
	class  *ClassContext  // nil when not in a class
	method *MethodContext // nil when not in a method/function

	// Variable tracking
	vars            map[string]string // declared variables and their types
	constTypes      map[string]string // module-level constant types (persists across functions)
	nonNegativeVars map[string]bool   // variables known to be >= 0 (for native indexing)
	goInteropVars   map[string]bool   // variables holding Go interop types

	// Import tracking
	imports            map[string]bool // import aliases for Go interop detection
	needsRuntime       bool            // needs rugby/runtime import
	needsFmt           bool            // needs fmt import (string interpolation)
	needsErrors        bool            // needs errors import (error_is?, error_as)
	needsRegexp        bool            // needs regexp import (regex literals)
	needsConstraints   bool            // needs constraints import (generics)
	needsTestingImport bool            // needs testing import
	needsTestImport    bool            // needs rugby/test import
	needsTemplate      bool            // needs rugby/template import (compile-time templates)
	needsStrings       bool            // needs strings import (template codegen)

	// Module state
	currentModule            string                     // current module being generated
	pendingOriginalClassName string                     // original class name for module-scoped classes (set before genClassDecl)
	modules                  map[string]*ast.ModuleDecl // module definitions for include

	// Loop/lambda state
	loopDepth       int  // nesting depth of loops (for break/next)
	inInlinedLambda bool // inside inlined lambda (each/times); return -> continue
	inMainFunc      bool // inside main() function

	// Source mapping
	sourceFile  string // original .rg filename for //line directives
	sourceDir   string // directory containing the source file (for relative paths)
	emitLineDir bool   // whether to emit //line directives

	// Type info from semantic analysis
	typeInfo TypeInfo

	// Struct state
	currentStruct *ast.StructDecl            // current struct being generated
	structs       map[string]*ast.StructDecl // struct definitions

	// Enum state
	enums map[string]*ast.EnumDecl // enum definitions

	// Function tracking
	functions            map[string]*ast.FuncDecl // function declarations
	pubFuncs             map[string]bool          // functions declared with 'pub'
	forceAnyLambdaReturn bool                     // lambda args to Go interop use 'any' return

	// Class tracking (global, not per-class)
	baseClasses        map[string]bool              // classes extended by other classes
	baseClassMethods   map[string][]string          // base class -> method names (lint suppression)
	baseClassAccessors map[string][]string          // base class -> accessor names (lint suppression)
	classMethods       map[string]map[string]string // class -> method -> generated function name
	classVars          map[string]string            // "Class@@var" -> "_Class_var"
	pubAccessors       map[string]map[string]bool   // class -> accessor -> isPub
	privateMethods     map[string]map[string]bool   // class -> method -> isPrivate
	instanceMethods    map[string]map[string]string // class -> method -> goName
	pubMethods         map[string]map[string]bool   // class -> method -> isPub
	classParent        map[string]string            // child -> parent class

	// Error collection
	errors         []error
	tempVarCounter int
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
	if len(g.currentReturnTypes()) == 0 {
		return false
	}
	lastType := g.currentReturnTypes()[len(g.currentReturnTypes())-1]
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
		// Extract directory for resolving relative paths in template.compile_file
		if idx := strings.LastIndex(path, "/"); idx != -1 {
			g.sourceDir = path[:idx]
		} else if idx := strings.LastIndex(path, "\\"); idx != -1 {
			g.sourceDir = path[:idx]
		} else {
			g.sourceDir = "."
		}
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
		vars:               make(map[string]string),
		constTypes:         make(map[string]string),
		imports:            make(map[string]bool),
		modules:            make(map[string]*ast.ModuleDecl),
		goInteropVars:      make(map[string]bool),
		baseClasses:        make(map[string]bool),
		baseClassMethods:   make(map[string][]string),
		baseClassAccessors: make(map[string][]string),
		classMethods:       make(map[string]map[string]string),
		classVars:          make(map[string]string),
		pubFuncs:           make(map[string]bool),
		pubAccessors:       make(map[string]map[string]bool),
		privateMethods:     make(map[string]map[string]bool),
		instanceMethods:    make(map[string]map[string]string),
		pubMethods:         make(map[string]map[string]bool),
		classParent:        make(map[string]string),
		functions:          make(map[string]*ast.FuncDecl),
		nonNegativeVars:    make(map[string]bool),
		structs:            make(map[string]*ast.StructDecl),
		enums:              make(map[string]*ast.EnumDecl),
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// --- Class context helpers ---

// currentClass returns the current class name, or empty string if not in a class.
func (g *Generator) currentClass() string {
	if g.class == nil {
		return ""
	}
	return g.class.Name
}

// currentOriginalClass returns the original class name for module-scoped classes.
func (g *Generator) currentOriginalClass() string {
	if g.class == nil {
		return ""
	}
	return g.class.OriginalName
}

// currentClassEmbeds returns the embedded types (parent classes) of the current class.
func (g *Generator) currentClassEmbeds() []string {
	if g.class == nil {
		return nil
	}
	return g.class.Embeds
}

// accessorFields returns fields with accessor methods (need underscore prefix).
func (g *Generator) accessorFields() map[string]bool {
	if g.class == nil {
		return nil
	}
	return g.class.AccessorFields
}

// --- Method context helpers ---

// currentMethod returns the current method name, or empty string if not in a method.
func (g *Generator) currentMethod() string {
	if g.method == nil {
		return ""
	}
	return g.method.Name
}

// currentReturnTypes returns the return types of the current function/method.
func (g *Generator) currentReturnTypes() []string {
	if g.method == nil {
		return nil
	}
	return g.method.ReturnTypes
}

// currentFuncTypeParams returns the type parameters of the current function.
func (g *Generator) currentFuncTypeParams() map[string]string {
	if g.method == nil {
		return nil
	}
	return g.method.TypeParams
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
	if g.accessorFields()[fieldName] {
		return true
	}
	// Check current class in semantic analysis
	// For module-scoped classes, use the original class name
	classToCheck := g.currentClass()
	if g.currentOriginalClass() != "" {
		classToCheck = g.currentOriginalClass()
	}
	if g.typeInfo.HasAccessor(classToCheck, fieldName) {
		return true
	}
	// Check parent classes (through embedding)
	for _, embed := range g.currentClassEmbeds() {
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
	if !g.isParentAccessor(g.currentClassEmbeds(), fieldName) {
		return "" // Field is local, no path needed
	}

	// Find which parent class has this accessor and build the path
	return g.findFieldPath(g.currentClassEmbeds(), fieldName)
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
			// Found it in this parent class - return path with private field prefix
			return parentClass + "." + privateFieldName(fieldName)
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
	return g.typeInfo.GetFieldType(g.currentClass(), fieldName)
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

// privateFieldName returns the internal field name for a declared accessor.
// Rugby uses underscore prefix to avoid name conflicts with getter/setter methods.
// Example: "name" -> "_name"
func privateFieldName(name string) string {
	return "_" + name
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
	if g.typeInfo == nil || g.currentClass() == "" {
		return false
	}
	return g.typeInfo.IsModuleMethod(g.currentClass(), methodName)
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

// extractMapTypes extracts the key and value types from a Go map type string.
// For example, "map[string]int" returns ("string", "int").
// Returns ("any", "any") if parsing fails.
func (g *Generator) extractMapTypes(mapType string) (keyType, valueType string) {
	keyType = "any"
	valueType = "any"
	if !strings.HasPrefix(mapType, "map[") {
		return
	}
	rest := mapType[4:]
	depth := 0
	for i, ch := range rest {
		switch ch {
		case '[':
			depth++
		case ']':
			if depth == 0 {
				keyType = rest[:i]
				valueType = rest[i+1:]
				return
			}
			depth--
		}
	}
	return
}

// mapTypeClassAware converts Rugby type names to Go type names, with awareness
// of class types. Classes in generic containers like Array<Body> need pointer
// prefixes, resulting in []*Body instead of []Body.
func (g *Generator) mapTypeClassAware(rubyType string) string {
	// Handle Array<T> specially to check if T is a class
	if strings.HasPrefix(rubyType, "Array<") && strings.HasSuffix(rubyType, ">") {
		inner := rubyType[6 : len(rubyType)-1]
		innerMapped := g.mapTypeClassAware(inner)
		// If the inner type is a class, add pointer prefix
		if g.typeInfo != nil && g.typeInfo.IsClass(inner) {
			return "[]*" + innerMapped
		}
		return "[]" + innerMapped
	}
	// Handle Map<K, V> and Hash<K, V> specially to check if V is a class
	if (strings.HasPrefix(rubyType, "Map<") || strings.HasPrefix(rubyType, "Hash<")) && strings.HasSuffix(rubyType, ">") {
		prefixLen := 4
		if strings.HasPrefix(rubyType, "Hash<") {
			prefixLen = 5
		}
		content := rubyType[prefixLen : len(rubyType)-1]
		parts := strings.Split(content, ",")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			keyMapped := g.mapTypeClassAware(key)
			valMapped := g.mapTypeClassAware(val)
			// If the value type is a class, add pointer prefix
			if g.typeInfo != nil && g.typeInfo.IsClass(val) {
				return fmt.Sprintf("map[%s]*%s", keyMapped, valMapped)
			}
			return fmt.Sprintf("map[%s]%s", keyMapped, valMapped)
		}
	}
	// For other types, use the regular mapType
	return mapType(rubyType)
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

// inferLambdaReturnType infers the Go return type for a lambda expression.
// It looks at the last expression in the lambda body and infers its type.
// Falls back to "any" if the type cannot be determined.
func (g *Generator) inferLambdaReturnType(lambda *ast.LambdaExpr, _ string) string {
	return g.inferBlockBodyReturnType(lambda.Body)
}

// inferBlockBodyReturnType infers the Go return type for a block body.
// It looks at the last statement and infers its type.
// Falls back to "any" if the type cannot be determined.
func (g *Generator) inferBlockBodyReturnType(body []ast.Statement) string {
	if len(body) == 0 {
		return "any"
	}

	// Get the last statement
	lastStmt := body[len(body)-1]

	// If it's an expression statement, get the expression's type
	if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
		return g.inferExprGoTypeDeep(exprStmt.Expr)
	}

	// If it's a return statement with a value
	if retStmt, ok := lastStmt.(*ast.ReturnStmt); ok && len(retStmt.Values) > 0 {
		return g.inferExprGoTypeDeep(retStmt.Values[0])
	}

	return "any"
}

// inferExprGoTypeDeep infers the Go type for an expression with deep analysis.
// It first tries standard inference, then recursively analyzes binary expressions
// and checks the vars map for identifiers.
func (g *Generator) inferExprGoTypeDeep(expr ast.Expression) string {
	// First try the standard inference
	if goType := g.inferExprGoType(expr); goType != "" && goType != "any" {
		return goType
	}

	// For binary expressions, infer from operands
	if binExpr, ok := expr.(*ast.BinaryExpr); ok {
		switch binExpr.Op {
		case "+", "-", "*", "/", "%":
			// Arithmetic operations preserve type
			leftType := g.inferExprGoTypeDeep(binExpr.Left)
			if leftType != "" && leftType != "any" {
				return leftType
			}
			rightType := g.inferExprGoTypeDeep(binExpr.Right)
			if rightType != "" && rightType != "any" {
				return rightType
			}
			// If operand is an identifier, check vars
			if ident, ok := binExpr.Left.(*ast.Ident); ok {
				if varType, exists := g.vars[ident.Name]; exists && varType != "" {
					return varType
				}
			}
		case "==", "!=", "<", ">", "<=", ">=", "and", "or":
			return "bool"
		}
	}

	// For identifiers, check vars
	if ident, ok := expr.(*ast.Ident); ok {
		if varType, exists := g.vars[ident.Name]; exists && varType != "" {
			return varType
		}
	}

	// For method calls on known types, infer return type
	if call, ok := expr.(*ast.CallExpr); ok {
		if sel, ok := call.Func.(*ast.SelectorExpr); ok {
			// String methods that return string
			if sel.Sel == "upcase" || sel.Sel == "downcase" || sel.Sel == "strip" ||
				sel.Sel == "reverse" || sel.Sel == "capitalize" {
				return "string"
			}
			// Methods that return int
			if sel.Sel == "length" || sel.Sel == "size" || sel.Sel == "abs" {
				return "int"
			}
		}
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
	// - Check codegen's vars map for variables from if-let bindings
	//   The semantic analyzer may not have type info for these
	if ident, ok := sel.X.(*ast.Ident); ok {
		if varType, ok := g.vars[ident.Name]; ok && varType != "" {
			// Strip pointer/optional suffixes
			varType = strings.TrimPrefix(varType, "*")
			varType = strings.TrimSuffix(varType, "?")
			if g.isClassInstanceMethod(varType, sel.Sel) != "" {
				return ast.SelectorGetter
			}
		}
	}

	return ast.SelectorUnknown
}

// inferTypeFromExpr returns the Rugby type for an expression.
// Primary source is TypeInfo from semantic analysis. Fallback is minimal
// literal type detection for robustness.
func (g *Generator) inferTypeFromExpr(expr ast.Expression) string {
	// First, handle special cases that can provide more accurate types than semantic analysis
	switch e := expr.(type) {
	case *ast.CallExpr:
		// Constructor call: ClassName.new(...) -> ClassName
		// Handles both non-generic (Foo.new) and generic (Array<T>.new) constructors
		if sel, ok := e.Func.(*ast.SelectorExpr); ok && sel.Sel == "new" {
			if ident, ok := sel.X.(*ast.Ident); ok {
				return ident.Name
			}
			// Handle generic types: Array<T>.new, Hash<K,V>.new, etc.
			// sel.X is an IndexExpr where Left is the base type and Index is the type param
			if idx, ok := sel.X.(*ast.IndexExpr); ok {
				if typeStr := extractGenericTypeString(idx); typeStr != "" {
					return typeStr
				}
			}
		}
	}

	// Primary: use semantic type info
	// GetRugbyType returns the full type including generics (e.g., "Array<Int>", "String?")
	// Note: "any" is a valid type that should be returned
	rugbyType := g.typeInfo.GetRugbyType(expr)
	if rugbyType != "" && rugbyType != "unknown" && !strings.Contains(rugbyType, "unknown") {
		return rugbyType
	}

	// For identifiers, check codegen's variable tracking as fallback
	// This handles cases where:
	// - The semantic analyzer doesn't have type info for the specific node
	// - The semantic analyzer has incomplete type info (e.g., Array<unknown>)
	if ident, ok := expr.(*ast.Ident); ok {
		if varType, ok := g.vars[ident.Name]; ok && varType != "" && !strings.Contains(varType, "unknown") {
			return varType
		}
		// Check module-level const types (set by compile-time handlers, persists across functions)
		if constType, ok := g.constTypes[ident.Name]; ok && constType != "" {
			return constType
		}
	}

	// If semantic type is partially known (contains "unknown"), still use it as fallback
	if rugbyType != "" && rugbyType != "unknown" {
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
		// Note: Constructor calls (ClassName.new) are handled by early check at top of function
		// Here we only handle struct method calls for return type inference
		if sel, ok := e.Func.(*ast.SelectorExpr); ok && sel.Sel != "new" {
			receiverType := g.inferTypeFromExpr(sel.X)

			// Check stdLib for method return type (e.g., String.bytes -> []byte)
			lookupType := receiverType
			switch receiverType {
			case "string":
				lookupType = "String"
			case "int":
				lookupType = "Int"
			case "float64":
				lookupType = "Float"
			case "bool":
				lookupType = "Bool"
			}
			if methods, ok := stdLib[lookupType]; ok {
				if method, ok := methods[sel.Sel]; ok && method.ReturnType != "" {
					return method.ReturnType
				}
			}

			// Struct method call: p.moved(...) -> look up method return type
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
	case *ast.IndexExpr:
		// For array/slice indexing, get element type from collection type
		collType := g.inferTypeFromExpr(e.Left)
		if collType != "" {
			return extractArrayElementType(collType)
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

	// Pre-pass 1: Collect metadata from all declarations in a single traversal.
	// This pass collects:
	// - Base class relationships (for inheritance and lint suppression)
	// - Public functions (for correct casing at call sites)
	// - Class metadata: public accessors, private methods, instance methods
	// - Module nested class methods (for type lookups)
	// - Struct methods (always public)
	for _, def := range definitions {
		switch d := def.(type) {
		case *ast.FuncDecl:
			if d.Pub {
				g.pubFuncs[d.Name] = true
			}

		case *ast.ClassDecl:
			// Track base classes and inheritance hierarchy
			for _, embed := range d.Embeds {
				g.baseClasses[embed] = true
				if len(d.Embeds) == 1 {
					g.classParent[d.Name] = embed
				}
			}

			// Collect public accessors
			for _, acc := range d.Accessors {
				if acc.Pub || d.Pub {
					if g.pubAccessors[d.Name] == nil {
						g.pubAccessors[d.Name] = make(map[string]bool)
					}
					g.pubAccessors[d.Name][acc.Name] = true
				}
				// Collect accessor getters as instance methods
				if acc.Kind == "getter" || acc.Kind == "property" {
					if g.instanceMethods[d.Name] == nil {
						g.instanceMethods[d.Name] = make(map[string]string)
					}
					g.instanceMethods[d.Name][acc.Name] = goFuncName(acc.Name, acc.Pub || d.Pub)
				}
			}

			// Collect methods: private, public, and instance methods
			for _, method := range d.Methods {
				if method.Private {
					if g.privateMethods[d.Name] == nil {
						g.privateMethods[d.Name] = make(map[string]bool)
					}
					g.privateMethods[d.Name][method.Name] = true
				}
				if method.Pub {
					if g.pubMethods[d.Name] == nil {
						g.pubMethods[d.Name] = make(map[string]bool)
					}
					g.pubMethods[d.Name][method.Name] = true
				}
				if !method.IsClassMethod {
					if g.instanceMethods[d.Name] == nil {
						g.instanceMethods[d.Name] = make(map[string]string)
					}
					var goName string
					if baseName, isSetter := strings.CutSuffix(method.Name, "="); isSetter {
						goName = goSetterName(baseName, method.Pub, method.Private)
					} else {
						goName = goMethodName(method.Name, method.Pub, method.Private, false)
					}
					g.instanceMethods[d.Name][method.Name] = goName
				}
			}

		case *ast.ModuleDecl:
			// Collect instance methods for nested classes
			for _, cls := range d.Classes {
				for _, method := range cls.Methods {
					if !method.IsClassMethod {
						if g.instanceMethods[cls.Name] == nil {
							g.instanceMethods[cls.Name] = make(map[string]string)
						}
						var goName string
						if baseName, isSetter := strings.CutSuffix(method.Name, "="); isSetter {
							goName = goSetterName(baseName, method.Pub, method.Private)
						} else {
							goName = goMethodName(method.Name, method.Pub, method.Private, false)
						}
						g.instanceMethods[cls.Name][method.Name] = goName
					}
				}
				for _, acc := range cls.Accessors {
					if acc.Kind == "getter" || acc.Kind == "property" {
						if g.instanceMethods[cls.Name] == nil {
							g.instanceMethods[cls.Name] = make(map[string]string)
						}
						g.instanceMethods[cls.Name][acc.Name] = goFuncName(acc.Name, acc.Pub || cls.Pub)
					}
				}
			}

		case *ast.StructDecl:
			// Struct methods are always public (PascalCase)
			for _, method := range d.Methods {
				if g.instanceMethods[d.Name] == nil {
					g.instanceMethods[d.Name] = make(map[string]string)
				}
				g.instanceMethods[d.Name][method.Name] = goFuncName(method.Name, true)
			}
		}
	}

	// Pre-pass 2: Collect base class methods and accessors.
	// This must be a separate pass because it depends on g.baseClasses populated above.
	// Used to generate var _ = (*BaseClass).methodName assertions for lint suppression.
	for _, def := range definitions {
		if cls, ok := def.(*ast.ClassDecl); ok && g.baseClasses[cls.Name] {
			for _, m := range cls.Methods {
				if m.Name != "initialize" {
					g.baseClassMethods[cls.Name] = append(g.baseClassMethods[cls.Name], m.Name)
				}
			}
			for _, acc := range cls.Accessors {
				if acc.Kind == "getter" || acc.Kind == "property" {
					g.baseClassAccessors[cls.Name] = append(g.baseClassAccessors[cls.Name], acc.Name)
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
		g.inMainFunc = true // Enable main-specific features like bang operator
		clear(g.vars)       // Reset variable tracking for implicit main scope
		for _, stmt := range topLevelStmts {
			g.genStatement(stmt)
		}
		g.inMainFunc = false
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
	const templateImport = "github.com/nchapman/rugby/stdlib/template"
	const constraintsImport = "golang.org/x/exp/constraints"
	needsRuntimeImport := g.needsRuntime && !userImports[runtimeImport]
	needsFmtImport := g.needsFmt && !userImports["fmt"]
	needsErrorsImport := g.needsErrors && !userImports["errors"]
	needsRegexpImport := g.needsRegexp && !userImports["regexp"]
	needsStringsImport := g.needsStrings && !userImports["strings"]
	needsTestingImport := g.needsTestingImport && !userImports["testing"]
	needsTestImport := g.needsTestImport && !userImports[testImport]
	needsTemplateImport := g.needsTemplate && !userImports[templateImport]
	needsConstraintsImport := g.needsConstraints && !userImports[constraintsImport]
	hasImports := len(program.Imports) > 0 || needsRuntimeImport || needsFmtImport || needsErrorsImport || needsRegexpImport || needsStringsImport || needsTestingImport || needsTestImport || needsTemplateImport || needsConstraintsImport
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
		if needsStringsImport {
			out.WriteString("\t\"strings\"\n")
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
		if needsTemplateImport {
			out.WriteString(fmt.Sprintf("\t%q\n", templateImport))
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
