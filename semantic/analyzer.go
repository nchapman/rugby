package semantic

import (
	"fmt"
	"maps"
	"strings"

	"github.com/nchapman/rugby/ast"
)

// typeAliasInfo stores information about a type alias for error reporting.
type typeAliasInfo struct {
	Type string // underlying type
	Line int    // declaration line
}

// Analyzer performs semantic analysis on a Rugby AST.
type Analyzer struct {
	scope       *Scope
	globalScope *Scope
	errors      []error

	// Type information attached to AST nodes
	nodeTypes map[ast.Node]*Type

	// Selector kind resolution (field vs method vs getter vs setter)
	selectorKinds map[*ast.SelectorExpr]ast.SelectorKind

	// Class and interface definitions
	classes    map[string]*Symbol
	interfaces map[string]*Symbol
	modules    map[string]*Symbol
	functions  map[string]*Symbol

	// Type aliases: maps alias name to alias info (type and line)
	typeAliases map[string]typeAliasInfo

	// Enums: maps enum name to enum info
	enums map[string]*ast.EnumDecl

	// Structs: track struct names (for immutability enforcement)
	structs map[string]bool

	// Track which symbols are used (read) vs just declared
	usedSymbols map[*Symbol]bool

	// Track symbols declared at each AST node (for multi-assign unused var detection)
	declaredAt map[ast.Node]map[string]*Symbol

	// Track which assignment nodes declare new variables (for := vs = in codegen)
	declarations map[ast.Node]bool
}

// NewAnalyzer creates a new semantic analyzer.
func NewAnalyzer() *Analyzer {
	global := NewGlobalScope()
	a := &Analyzer{
		scope:         global,
		globalScope:   global,
		nodeTypes:     make(map[ast.Node]*Type),
		selectorKinds: make(map[*ast.SelectorExpr]ast.SelectorKind),
		classes:       make(map[string]*Symbol),
		interfaces:    make(map[string]*Symbol),
		modules:       make(map[string]*Symbol),
		functions:     make(map[string]*Symbol),
		typeAliases:   make(map[string]typeAliasInfo),
		enums:         make(map[string]*ast.EnumDecl),
		structs:       make(map[string]bool),
		usedSymbols:   make(map[*Symbol]bool),
		declaredAt:    make(map[ast.Node]map[string]*Symbol),
		declarations:  make(map[ast.Node]bool),
	}
	a.defineBuiltins()
	return a
}

// defineBuiltins adds built-in functions and types to the global scope.
func (a *Analyzer) defineBuiltins() {
	// Kernel functions
	builtins := []struct {
		name     string
		params   []*Type
		returns  []*Type
		variadic bool
	}{
		{"puts", []*Type{TypeAnyVal}, nil, true},
		{"print", []*Type{TypeAnyVal}, nil, true},
		{"p", []*Type{TypeAnyVal}, nil, true},
		{"gets", nil, []*Type{TypeStringVal}, false},
		{"exit", []*Type{TypeIntVal}, nil, false},
		{"sleep", []*Type{TypeFloatVal}, nil, false},
		{"rand", []*Type{TypeIntVal}, []*Type{TypeAnyVal}, true}, // 0 args -> Float, 1 arg -> Int
		{"error", []*Type{TypeStringVal}, []*Type{TypeErrorVal}, false},
		{"panic", []*Type{TypeStringVal}, nil, false},
	}

	for _, b := range builtins {
		params := make([]*Symbol, len(b.params))
		for i, pt := range b.params {
			params[i] = NewParam("arg"+string(rune('0'+i)), pt)
		}
		fn := NewFunction(b.name, params, b.returns)
		fn.Variadic = b.variadic
		fn.Builtin = true
		mustDefine(a.globalScope, fn)
		a.functions[b.name] = fn
	}
}

// Analyze performs semantic analysis on the program.
// Returns the list of errors found during analysis.
func (a *Analyzer) Analyze(program *ast.Program) []error {
	// Register Go package imports as valid identifiers
	for _, imp := range program.Imports {
		name := imp.Alias
		if name == "" {
			// Use last part of path as implicit alias (e.g., "encoding/json" -> "json")
			parts := strings.Split(imp.Path, "/")
			name = parts[len(parts)-1]
		}
		// Register the import as a Go package symbol in global scope
		sym := &Symbol{
			Name: name,
			Kind: SymGoPackage,
			Type: TypeAnyVal, // Go packages can have any methods
		}
		mustDefine(a.globalScope, sym)
	}

	// Pre-pass: collect type aliases and enums first so they're available when parsing other types
	for _, decl := range program.Declarations {
		if typeAlias, ok := decl.(*ast.TypeAliasDecl); ok {
			a.collectTypeAlias(typeAlias)
		}
		if enumDecl, ok := decl.(*ast.EnumDecl); ok {
			a.collectEnum(enumDecl)
		}
	}

	// First pass: collect all top-level declarations (classes, interfaces, modules, functions)
	for _, decl := range program.Declarations {
		a.collectDeclaration(decl)
	}

	// Propagate inherited fields from parent classes to child classes
	// This must happen after all classes are collected
	a.propagateInheritedFields(program)

	// Resolve types: convert class types to interface types where appropriate
	// This must happen after all declarations are collected so we know what interfaces exist
	a.resolveTypes()

	// Validate interface implementations
	a.validateInterfaceImplementations()

	// Second pass: analyze bodies
	for _, decl := range program.Declarations {
		a.analyzeStatement(decl)
	}

	return a.errors
}

// propagateInheritedFields copies fields from parent classes to child classes.
// This ensures that inherited fields (including getters/setters) are accessible
// through the child class symbol. Uses recursion to handle multi-level inheritance.
// Also detects and reports circular inheritance.
func (a *Analyzer) propagateInheritedFields(program *ast.Program) {
	// Build a map from class name to its AST declaration for easy lookup
	classDecls := make(map[string]*ast.ClassDecl)
	for _, decl := range program.Declarations {
		if cls, ok := decl.(*ast.ClassDecl); ok {
			classDecls[cls.Name] = cls
		}
	}

	// Track which classes have been fully processed
	visited := make(map[string]bool)
	// Track current path for cycle detection
	inPath := make(map[string]bool)

	// Recursive function to propagate fields from all ancestors
	var propagate func(className string, path []string) bool
	propagate = func(className string, path []string) bool {
		// Check for cycle: if className is already in current path, we have a cycle
		if inPath[className] {
			// Build cycle path for error message
			cycle := append(path, className)
			cls := classDecls[path[0]]
			line := 0
			if cls != nil {
				line = cls.Line
			}
			a.addError(&CircularInheritanceError{
				Cycle: cycle,
				Line:  line,
			})
			return false
		}

		// Skip if already fully processed
		if visited[className] {
			return true
		}

		cls := classDecls[className]
		if cls == nil || len(cls.Embeds) == 0 {
			visited[className] = true
			return true
		}

		// Mark as being processed (in current path)
		inPath[className] = true
		newPath := append(path, className)

		childSym := a.classes[className]
		if childSym == nil {
			inPath[className] = false
			visited[className] = true
			return true
		}

		// Copy fields from parent class(es)
		for _, parentName := range cls.Embeds {
			// Recursively ensure parent has all its inherited fields first
			if !propagate(parentName, newPath) {
				// Cycle detected, stop processing
				inPath[className] = false
				return false
			}

			parentSym := a.classes[parentName]
			if parentSym == nil {
				continue
			}

			// Copy all fields from parent that don't exist in child
			for name, field := range parentSym.Fields {
				if childSym.Fields[name] == nil {
					// Create a copy of the field symbol
					inheritedField := NewField(name, field.Type)
					inheritedField.HasGetter = field.HasGetter
					inheritedField.HasSetter = field.HasSetter
					inheritedField.Public = field.Public
					childSym.AddField(inheritedField)
				}
			}
		}

		// Store the parent class name for later reference.
		// Note: Methods are NOT copied here. Instead, getClassMethod() traverses the
		// Parent chain at lookup time. This is more memory-efficient and ensures
		// method override resolution works correctly.
		childSym.Parent = cls.Embeds[0]

		// Mark as fully processed and remove from path
		inPath[className] = false
		visited[className] = true
		return true
	}

	// Process all classes
	for className := range classDecls {
		propagate(className, nil)
	}
}

// validateInterfaceImplementations checks that classes implement their declared interfaces.
func (a *Analyzer) validateInterfaceImplementations() {
	for _, cls := range a.classes {
		for _, ifaceName := range cls.Implements {
			iface := a.interfaces[ifaceName]
			if iface == nil {
				a.addError(&UndefinedError{
					Name: ifaceName,
					Line: cls.Line,
				})
				continue
			}

			// Check each interface method is implemented (including inherited methods)
			for methodName, ifaceMethod := range iface.Methods {
				clsMethod := a.getClassMethod(cls.Name, methodName)
				if clsMethod == nil {
					a.addError(&InterfaceNotImplementedError{
						ClassName:     cls.Name,
						InterfaceName: ifaceName,
						MissingMethod: methodName,
						Line:          cls.Line,
					})
					continue
				}

				// Check parameter count matches
				if len(clsMethod.Params) != len(ifaceMethod.Params) {
					a.addError(&MethodSignatureMismatchError{
						ClassName:     cls.Name,
						InterfaceName: ifaceName,
						MethodName:    methodName,
						Expected:      a.formatMethodSignature(ifaceMethod),
						Got:           a.formatMethodSignature(clsMethod),
						Line:          clsMethod.Line,
					})
					continue
				}

				// Check parameter types match
				for i, ifaceParam := range ifaceMethod.Params {
					clsParam := clsMethod.Params[i]
					if !a.typesMatch(ifaceParam.Type, clsParam.Type) {
						a.addError(&MethodSignatureMismatchError{
							ClassName:     cls.Name,
							InterfaceName: ifaceName,
							MethodName:    methodName,
							Expected:      a.formatMethodSignature(ifaceMethod),
							Got:           a.formatMethodSignature(clsMethod),
							Line:          clsMethod.Line,
						})
						break
					}
				}

				// Check return types match
				if len(clsMethod.ReturnTypes) != len(ifaceMethod.ReturnTypes) {
					a.addError(&MethodSignatureMismatchError{
						ClassName:     cls.Name,
						InterfaceName: ifaceName,
						MethodName:    methodName,
						Expected:      a.formatMethodSignature(ifaceMethod),
						Got:           a.formatMethodSignature(clsMethod),
						Line:          clsMethod.Line,
					})
					continue
				}

				for i, ifaceRet := range ifaceMethod.ReturnTypes {
					clsRet := clsMethod.ReturnTypes[i]
					if !a.typesMatch(ifaceRet, clsRet) {
						a.addError(&MethodSignatureMismatchError{
							ClassName:     cls.Name,
							InterfaceName: ifaceName,
							MethodName:    methodName,
							Expected:      a.formatMethodSignature(ifaceMethod),
							Got:           a.formatMethodSignature(clsMethod),
							Line:          clsMethod.Line,
						})
						break
					}
				}
			}
		}
	}
}

// typesMatch checks if two types are compatible for interface implementation.
func (a *Analyzer) typesMatch(expected, got *Type) bool {
	if expected == nil || got == nil {
		return true // Unknown types are compatible
	}
	if expected.Kind == TypeUnknown || got.Kind == TypeUnknown {
		return true
	}
	if expected.Kind == TypeAny || got.Kind == TypeAny {
		return true
	}
	return expected.Equals(got)
}

// getClassMethod finds a method on a class, including inherited methods from parent classes.
func (a *Analyzer) getClassMethod(className, methodName string) *Symbol {
	for className != "" {
		cls := a.classes[className]
		if cls == nil {
			return nil
		}
		if method := cls.GetMethod(methodName); method != nil {
			return method
		}
		className = cls.Parent
	}
	return nil
}

// getClassField finds a field on a class, including inherited fields from parent classes.
func (a *Analyzer) getClassField(className, fieldName string) *Symbol {
	for className != "" {
		cls := a.classes[className]
		if cls == nil {
			return nil
		}
		if field := cls.GetField(fieldName); field != nil {
			return field
		}
		className = cls.Parent
	}
	return nil
}

// formatMethodSignature returns a human-readable method signature.
func (a *Analyzer) formatMethodSignature(method *Symbol) string {
	var params []string
	for _, p := range method.Params {
		if p.Type != nil {
			params = append(params, p.Type.String())
		} else {
			params = append(params, "unknown")
		}
	}

	var returns []string
	for _, r := range method.ReturnTypes {
		if r != nil {
			returns = append(returns, r.String())
		}
	}

	sig := "(" + strings.Join(params, ", ") + ")"
	if len(returns) > 0 {
		sig += " -> " + strings.Join(returns, ", ")
	}
	return sig
}

// resolveTypes converts TypeClass types to TypeInterface where the name matches a declared interface.
// This is necessary because ParseType doesn't know about interfaces at parse time.
func (a *Analyzer) resolveTypes() {
	// Resolve function parameter and return types
	for _, fn := range a.functions {
		for _, param := range fn.Params {
			a.resolveType(param.Type)
		}
		for i, rt := range fn.ReturnTypes {
			fn.ReturnTypes[i] = a.resolveType(rt)
		}
	}

	// Resolve class method parameter and return types
	for _, cls := range a.classes {
		for _, method := range cls.Methods {
			for _, param := range method.Params {
				a.resolveType(param.Type)
			}
			for i, rt := range method.ReturnTypes {
				method.ReturnTypes[i] = a.resolveType(rt)
			}
		}
	}

	// Resolve module method parameter and return types
	for _, mod := range a.modules {
		for _, method := range mod.Methods {
			for _, param := range method.Params {
				a.resolveType(param.Type)
			}
			for i, rt := range method.ReturnTypes {
				method.ReturnTypes[i] = a.resolveType(rt)
			}
		}
	}

	// Resolve interface method parameter and return types (for consistency)
	for _, iface := range a.interfaces {
		for _, method := range iface.Methods {
			for _, param := range method.Params {
				a.resolveType(param.Type)
			}
			for i, rt := range method.ReturnTypes {
				method.ReturnTypes[i] = a.resolveType(rt)
			}
		}
	}
}

// resolveType checks if a type is a TypeClass that should be TypeInterface or marked as a struct.
// If the type name matches a declared interface, it updates the type's Kind to TypeInterface.
// If the type name matches a declared struct, it sets IsStruct to true.
// Also recursively resolves nested types (optional, array, map, etc.)
func (a *Analyzer) resolveType(t *Type) *Type {
	if t == nil {
		return t
	}

	// If this is a class type, check if it should be an interface or struct
	if t.Kind == TypeClass && t.Name != "" {
		if _, isInterface := a.interfaces[t.Name]; isInterface {
			t.Kind = TypeInterface
		} else if a.structs[t.Name] {
			t.IsStruct = true
		}
	}

	// Recursively resolve nested types
	if t.Elem != nil {
		a.resolveType(t.Elem)
	}
	if t.KeyType != nil {
		a.resolveType(t.KeyType)
	}
	if t.ValueType != nil {
		a.resolveType(t.ValueType)
	}
	for _, elem := range t.Elements {
		a.resolveType(elem)
	}

	return t
}

// collectDeclaration does a quick first pass to register top-level names.
// Note: TypeAliasDecl is handled in a separate pre-pass so aliases are available
// when parsing other types.
func (a *Analyzer) collectDeclaration(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.FuncDecl:
		a.collectFunc(s)
	case *ast.ClassDecl:
		a.collectClass(s)
	case *ast.InterfaceDecl:
		a.collectInterface(s)
	case *ast.ModuleDecl:
		a.collectModule(s)
	case *ast.ConstDecl:
		a.collectConst(s)
	case *ast.StructDecl:
		// Register struct for immutability enforcement
		a.structs[s.Name] = true
		// TypeAliasDecl is collected in pre-pass, skip here
	}
}

func (a *Analyzer) collectFunc(f *ast.FuncDecl) {
	// Build a map of type parameter names to their constraints for type-aware parsing
	var typeParams map[string]string
	if len(f.TypeParams) > 0 {
		typeParams = make(map[string]string, len(f.TypeParams))
		for _, tp := range f.TypeParams {
			typeParams[tp.Name] = tp.Constraint
		}
	}

	params := make([]*Symbol, len(f.Params))
	requiredParams := 0
	isVariadic := false
	for i, p := range f.Params {
		var typ *Type
		if p.Type != "" {
			typ = ParseTypeWithParams(p.Type, typeParams)
		} else {
			typ = TypeUnknownVal
		}
		params[i] = NewParam(p.Name, typ)
		if p.DefaultValue != nil {
			params[i].HasDefault = true
		} else if !p.Variadic {
			// Variadic params don't count as required (they accept 0+ args)
			requiredParams = i + 1 // All params up to this point are required
		}
		if p.Variadic {
			params[i].Variadic = true
			isVariadic = true
		}
	}

	returnTypes := make([]*Type, len(f.ReturnTypes))
	for i, rt := range f.ReturnTypes {
		returnTypes[i] = ParseTypeWithParams(rt, typeParams)
	}

	fn := NewFunction(f.Name, params, returnTypes)
	fn.Line = f.Line
	fn.Public = f.Pub
	fn.Node = f
	fn.RequiredParams = requiredParams
	fn.Variadic = isVariadic

	if err := a.scope.Define(fn); err != nil {
		a.addError(err)
	}
	a.functions[f.Name] = fn
}

func (a *Analyzer) collectClass(c *ast.ClassDecl) {
	cls := NewClass(c.Name)
	cls.Line = c.Line
	cls.Public = c.Pub
	cls.Implements = c.Implements
	cls.Includes = c.Includes
	cls.Node = c

	// Build a map of type parameter names to constraints from the class declaration
	var classTypeParams map[string]string
	if len(c.TypeParams) > 0 {
		classTypeParams = make(map[string]string, len(c.TypeParams))
		for _, tp := range c.TypeParams {
			classTypeParams[tp.Name] = tp.Constraint
		}
	}

	// Collect fields from explicit declarations and accessors
	for _, f := range c.Fields {
		field := NewField(f.Name, ParseTypeWithParams(f.Type, classTypeParams))
		cls.AddField(field)
	}
	for _, acc := range c.Accessors {
		field := NewField(acc.Name, ParseTypeWithParams(acc.Type, classTypeParams))
		// Set accessor flags based on accessor kind
		switch acc.Kind {
		case "getter":
			field.HasGetter = true
		case "setter":
			field.HasSetter = true
		case "property":
			field.HasGetter = true
			field.HasSetter = true
		}
		cls.AddField(field)
	}

	// Collect methods (signatures only for now)
	for _, m := range c.Methods {
		// Build type params for this method: class params + method-specific params
		methodTypeParams := classTypeParams
		if len(m.TypeParams) > 0 {
			methodTypeParams = make(map[string]string)
			maps.Copy(methodTypeParams, classTypeParams)
			for _, tp := range m.TypeParams {
				methodTypeParams[tp.Name] = tp.Constraint
			}
		}

		params := make([]*Symbol, len(m.Params))
		requiredParams := 0
		for i, p := range m.Params {
			var typ *Type
			if p.Type != "" {
				typ = ParseTypeWithParams(p.Type, methodTypeParams)
			} else {
				typ = TypeUnknownVal
			}
			params[i] = NewParam(p.Name, typ)
			if p.DefaultValue != nil {
				params[i].HasDefault = true
			} else {
				requiredParams = i + 1
			}
			if p.Variadic {
				params[i].Variadic = true
			}
		}

		returnTypes := make([]*Type, len(m.ReturnTypes))
		for i, rt := range m.ReturnTypes {
			returnTypes[i] = ParseTypeWithParams(rt, methodTypeParams)
		}

		// Check if method implicitly returns self (for method chaining)
		// Handles both bare `self` as last statement and explicit `return self`
		if len(m.ReturnTypes) == 0 && len(m.Body) > 0 {
			lastStmt := m.Body[len(m.Body)-1]
			returnsSelf := false

			// Check for bare `self` as last expression
			if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
				if ident, ok := exprStmt.Expr.(*ast.Ident); ok && ident.Name == "self" {
					returnsSelf = true
				}
			}
			// Check for explicit `return self`
			if returnStmt, ok := lastStmt.(*ast.ReturnStmt); ok {
				if len(returnStmt.Values) == 1 {
					if ident, ok := returnStmt.Values[0].(*ast.Ident); ok && ident.Name == "self" {
						returnsSelf = true
					}
				}
			}

			if returnsSelf {
				// Method returns self for chaining - add ClassName as return type
				classType := NewClassType(c.Name)
				returnTypes = []*Type{classType}
			}
		}

		method := NewMethod(m.Name, params, returnTypes)
		method.Line = m.Line
		method.Public = m.Pub
		method.Private = m.Private
		method.Node = m
		method.RequiredParams = requiredParams

		// Validate pub and private are mutually exclusive
		if m.Pub && m.Private {
			a.addError(&ConflictingVisibilityError{MethodName: m.Name, Line: m.Line})
		}

		if m.IsClassMethod {
			cls.AddClassMethod(method)
		} else {
			cls.AddMethod(method)
		}
	}

	if err := a.scope.Define(cls); err != nil {
		a.addError(err)
	}
	a.classes[c.Name] = cls
}

func (a *Analyzer) collectInterface(i *ast.InterfaceDecl) {
	iface := NewInterface(i.Name)
	iface.Line = i.Line
	iface.Public = i.Pub
	iface.Node = i

	for _, m := range i.Methods {
		params := make([]*Symbol, len(m.Params))
		requiredParams := 0
		for j, p := range m.Params {
			var typ *Type
			if p.Type != "" {
				typ = ParseType(p.Type)
			} else {
				typ = TypeUnknownVal
			}
			params[j] = NewParam(p.Name, typ)
			if p.DefaultValue != nil {
				params[j].HasDefault = true
			} else {
				requiredParams = j + 1
			}
			if p.Variadic {
				params[j].Variadic = true
			}
		}

		returnTypes := make([]*Type, len(m.ReturnTypes))
		for j, rt := range m.ReturnTypes {
			returnTypes[j] = ParseType(rt)
		}

		method := NewMethod(m.Name, params, returnTypes)
		method.RequiredParams = requiredParams
		iface.AddMethod(method)
	}

	if err := a.scope.Define(iface); err != nil {
		a.addError(err)
	}
	a.interfaces[i.Name] = iface
}

func (a *Analyzer) collectModule(m *ast.ModuleDecl) {
	mod := NewModule(m.Name)
	mod.Line = m.Line
	mod.Public = m.Pub
	mod.Node = m

	// Collect fields
	for _, f := range m.Fields {
		field := NewField(f.Name, ParseType(f.Type))
		mod.AddField(field)
	}

	// Collect accessors (like class accessors)
	for _, acc := range m.Accessors {
		field := NewField(acc.Name, ParseType(acc.Type))
		switch acc.Kind {
		case "getter":
			field.HasGetter = true
		case "setter":
			field.HasSetter = true
		case "property":
			field.HasGetter = true
			field.HasSetter = true
		}
		mod.AddField(field)
	}

	// Collect methods
	for _, method := range m.Methods {
		params := make([]*Symbol, len(method.Params))
		requiredParams := 0
		for i, p := range method.Params {
			var typ *Type
			if p.Type != "" {
				typ = ParseType(p.Type)
			} else {
				typ = TypeUnknownVal
			}
			params[i] = NewParam(p.Name, typ)
			if p.DefaultValue != nil {
				params[i].HasDefault = true
			} else {
				requiredParams = i + 1
			}
			if p.Variadic {
				params[i].Variadic = true
			}
		}

		returnTypes := make([]*Type, len(method.ReturnTypes))
		for i, rt := range method.ReturnTypes {
			returnTypes[i] = ParseType(rt)
		}

		meth := NewMethod(method.Name, params, returnTypes)
		meth.Line = method.Line
		meth.Public = method.Pub
		meth.Node = method
		meth.RequiredParams = requiredParams
		mod.AddMethod(meth)
	}

	if err := a.scope.Define(mod); err != nil {
		a.addError(err)
	}
	a.modules[m.Name] = mod

	// Collect nested classes defined within the module
	for _, cls := range m.Classes {
		a.collectClass(cls)
	}
}

// collectTypeAlias registers a type alias for later resolution.
func (a *Analyzer) collectTypeAlias(t *ast.TypeAliasDecl) {
	// Check for duplicate type alias
	if existing, ok := a.typeAliases[t.Name]; ok {
		a.addError(&DuplicateTypeAliasError{Name: t.Name, Line: existing.Line})
		return
	}

	// Note: Conflict checks with classes/interfaces/modules would need to happen
	// in a second pass since type aliases are collected before other declarations.
	// For now, Go will catch any naming conflicts at compile time.

	a.typeAliases[t.Name] = typeAliasInfo{Type: t.Type, Line: t.Line}
}

// collectEnum registers an enum declaration for later reference.
// Enums are treated as integer types with named constants.
func (a *Analyzer) collectEnum(e *ast.EnumDecl) {
	// Check for duplicate enum
	if existing, ok := a.enums[e.Name]; ok {
		a.addError(&DuplicateTypeAliasError{Name: e.Name, Line: existing.Line})
		return
	}

	// Store the enum for later reference
	a.enums[e.Name] = e

	// Define the enum type as an integer type alias
	// This allows variables to be typed as the enum
	a.typeAliases[e.Name] = typeAliasInfo{Type: "Int", Line: e.Line}

	// Validate enum values
	hasExplicit := false
	hasImplicit := false
	for _, v := range e.Values {
		if v.Value != nil {
			hasExplicit = true
			// Validate that explicit values are integer literals
			if _, ok := v.Value.(*ast.IntLit); !ok {
				a.addError(&Error{
					Message: fmt.Sprintf("enum '%s' value '%s' must be an integer literal", e.Name, v.Name),
					Line:    v.Line,
				})
			}
		} else {
			hasImplicit = true
		}
	}
	// Disallow mixing explicit and implicit values
	if hasExplicit && hasImplicit {
		a.addError(&Error{
			Message: fmt.Sprintf("enum '%s' has mixed explicit and implicit values; all values must be explicit or all must be implicit", e.Name),
			Line:    e.Line,
		})
	}

	// Define each enum value as a constant in the global scope
	for _, v := range e.Values {
		constSym := &Symbol{
			Name:   e.Name + "::" + v.Name, // Status::Active
			Kind:   SymConstant,
			Type:   &Type{Kind: TypeClass, Name: e.Name}, // Treat as the enum type
			Line:   v.Line,
			Public: true,
		}
		// Don't error on duplicates - will be caught elsewhere
		_ = a.globalScope.Define(constSym)
	}
}

// collectConst registers a constant declaration in the global scope.
func (a *Analyzer) collectConst(c *ast.ConstDecl) {
	// Analyze the constant value to determine its type
	valueType := a.analyzeExpr(c.Value)

	// If type annotation provided, use it; otherwise use inferred type
	var typ *Type
	if c.Type != "" {
		typ = ParseType(c.Type)
		// Check that value is assignable to declared type
		if valueType.Kind != TypeUnknown && !a.isAssignable(typ, valueType) {
			a.addError(&TypeMismatchError{
				Expected: typ,
				Got:      valueType,
				Line:     c.Line,
				Context:  "constant declaration",
			})
		}
	} else {
		typ = valueType
	}

	// Create constant symbol and add to global scope
	constSym := NewConstant(c.Name, typ)
	constSym.Line = c.Line
	if err := a.globalScope.Define(constSym); err != nil {
		a.addError(&Error{
			Message: "constant '" + c.Name + "' already defined",
			Line:    c.Line,
		})
	}
}

// resolveTypeAlias returns the underlying type name if the given name is a type alias,
// otherwise returns the original name. Handles chained aliases and reports circular errors.
func (a *Analyzer) resolveTypeAlias(name string) string {
	// Follow alias chain (e.g., type A = B, type B = Int → A resolves to Int)
	visited := make(map[string]bool)
	startName := name
	for {
		if visited[name] {
			// Circular alias - report error and stop
			if info, ok := a.typeAliases[startName]; ok {
				a.addError(&CircularTypeAliasError{Name: startName, Line: info.Line})
			}
			return name
		}
		visited[name] = true

		if info, ok := a.typeAliases[name]; ok {
			name = info.Type
		} else {
			return name
		}
	}
}

// resolveTypeAliasType checks if a Type is a class type that corresponds to a type alias,
// and if so, returns the resolved underlying type. Otherwise returns the type unchanged.
// This function recursively resolves all nested type aliases.
func (a *Analyzer) resolveTypeAliasType(t *Type) *Type {
	if t == nil {
		return t
	}

	// Only resolve TypeClass that might be type aliases
	if t.Kind == TypeClass && t.Name != "" {
		resolvedName := a.resolveTypeAlias(t.Name)
		if resolvedName != t.Name {
			// It was an alias, parse the underlying type and recursively resolve it
			resolved := ParseType(resolvedName)
			return a.resolveTypeAliasType(resolved)
		}
	}

	// Resolve nested types (e.g., Array<UserID> should resolve to Array<Int64>)
	if t.Elem != nil {
		resolved := a.resolveTypeAliasType(t.Elem)
		if resolved != t.Elem {
			switch t.Kind {
			case TypeArray:
				return NewArrayType(resolved)
			case TypeOptional:
				return NewOptionalType(resolved)
			case TypeChan:
				return NewChanType(resolved)
			case TypeTask:
				return NewTaskType(resolved)
			}
		}
	}

	// Resolve Map key and value types
	if t.Kind == TypeMap {
		keyResolved := a.resolveTypeAliasType(t.KeyType)
		valueResolved := a.resolveTypeAliasType(t.ValueType)
		if keyResolved != t.KeyType || valueResolved != t.ValueType {
			return NewMapType(keyResolved, valueResolved)
		}
	}

	return t
}

// analyzeStatement analyzes a statement and returns its type (if applicable).
func (a *Analyzer) analyzeStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.FuncDecl:
		a.analyzeFuncDecl(s)
	case *ast.ClassDecl:
		a.analyzeClassDecl(s)
	case *ast.StructDecl:
		// Struct registration handled in collectDeclaration
		// No additional analysis needed here
	case *ast.AssignStmt:
		a.analyzeAssign(s)
	case *ast.MultiAssignStmt:
		a.analyzeMultiAssign(s)
	case *ast.MapDestructuringStmt:
		a.analyzeMapDestructuring(s)
	case *ast.CompoundAssignStmt:
		a.analyzeCompoundAssign(s)
	case *ast.OrAssignStmt:
		a.analyzeOrAssign(s)
	case *ast.SelectorAssignStmt:
		a.analyzeSelectorAssign(s)
	case *ast.IndexAssignStmt:
		a.analyzeIndexAssign(s)
	case *ast.ExprStmt:
		a.analyzeExpr(s.Expr)
		if s.Condition != nil {
			a.analyzeExpr(s.Condition)
		}
	case *ast.IfStmt:
		a.analyzeIf(s)
	case *ast.WhileStmt:
		a.analyzeWhile(s)
	case *ast.UntilStmt:
		a.analyzeUntil(s)
	case *ast.ForStmt:
		a.analyzeFor(s)
	case *ast.LoopStmt:
		a.analyzeLoop(s)
	case *ast.ReturnStmt:
		a.analyzeReturn(s)
	case *ast.BreakStmt:
		a.analyzeBreak(s)
	case *ast.NextStmt:
		a.analyzeNext(s)
	case *ast.CaseStmt:
		a.analyzeCase(s)
	case *ast.CaseTypeStmt:
		a.analyzeCaseType(s)
	case *ast.InstanceVarAssign:
		a.analyzeInstanceVarAssign(s)
	case *ast.InstanceVarCompoundAssign:
		a.analyzeInstanceVarCompoundAssign(s)
	case *ast.GoStmt:
		a.analyzeGo(s)
	case *ast.SelectStmt:
		a.analyzeSelect(s)
	case *ast.DeferStmt:
		if s.Call != nil {
			a.analyzeExpr(s.Call)
		}
	case *ast.PanicStmt:
		if s.Message != nil {
			a.analyzeExpr(s.Message)
		}
	case *ast.ChanSendStmt:
		a.analyzeChanSend(s)
	// Test statements
	case *ast.DescribeStmt:
		a.analyzeDescribe(s)
	case *ast.ItStmt:
		a.analyzeIt(s)
	case *ast.TestStmt:
		a.analyzeTest(s)
	case *ast.TableStmt:
		a.analyzeTable(s)
	case *ast.BeforeStmt:
		a.analyzeBefore(s)
	case *ast.AfterStmt:
		a.analyzeAfter(s)
	case *ast.TypeAliasDecl:
		// Type aliases are collected in the pre-pass; no body analysis needed
	case *ast.EnumDecl:
		// Enums are collected in the pre-pass; no body analysis needed
	}
}

func (a *Analyzer) analyzeFuncDecl(f *ast.FuncDecl) {
	// Build a map of type parameter names to their constraints
	var typeParams map[string]string
	if len(f.TypeParams) > 0 {
		typeParams = make(map[string]string, len(f.TypeParams))
		for _, tp := range f.TypeParams {
			typeParams[tp.Name] = tp.Constraint
		}
	}

	// Create function scope
	fnScope := NewScope(ScopeFunction, a.scope)
	fnScope.ReturnTypes = make([]*Type, len(f.ReturnTypes))
	for i, rt := range f.ReturnTypes {
		fnScope.ReturnTypes[i] = ParseTypeWithParams(rt, typeParams)
	}

	// Add parameters to scope
	for _, p := range f.Params {
		var typ *Type
		if p.Type != "" {
			typ = ParseTypeWithParams(p.Type, typeParams)
		} else {
			typ = TypeUnknownVal
		}
		param := NewParam(p.Name, typ)
		param.Line = f.Line
		mustDefine(fnScope, param)
	}

	// Analyze body
	prevScope := a.scope
	a.scope = fnScope
	for i, stmt := range f.Body {
		// Special case: if this is the last statement and it's a lambda expression,
		// infer its types from the function's return type (for implicit return)
		if i == len(f.Body)-1 {
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				if lambda, ok := exprStmt.Expr.(*ast.LambdaExpr); ok {
					if len(fnScope.ReturnTypes) > 0 {
						// The function returns a function type - use it to infer lambda types
						expectedType := fnScope.ReturnTypes[0]
						resultType := a.analyzeLambdaWithExpectedType(lambda, expectedType)
						if resultType != nil {
							// Store the inferred type for codegen
							a.setNodeType(lambda, resultType)
						}
						continue
					}
				}
			}
		}
		a.analyzeStatement(stmt)
	}
	a.scope = prevScope
}

func (a *Analyzer) analyzeClassDecl(c *ast.ClassDecl) {
	cls := a.classes[c.Name]
	if cls == nil {
		return // Error already reported
	}

	// Create class scope
	classScope := NewScope(ScopeClass, a.scope)
	classScope.ClassName = c.Name

	// Add 'self' to scope
	selfSym := NewVariable("self", cls.Type)
	mustDefine(classScope, selfSym)

	// Add fields to scope
	for name, field := range cls.Fields {
		mustDefine(classScope, NewVariable(name, field.Type))
	}

	// Add fields and methods from included modules
	// Fields inherit accessor flags (HasGetter/HasSetter) from the module
	// Methods become callable as bare functions within class methods
	// AND as external method calls on class instances
	for _, modName := range cls.Includes {
		if mod := a.modules[modName]; mod != nil {
			// Add module fields (with accessor flags) to class
			for name, field := range mod.Fields {
				// Only add if class doesn't already have a field with this name
				if cls.Fields[name] == nil {
					inheritedField := NewField(name, field.Type)
					inheritedField.HasGetter = field.HasGetter
					inheritedField.HasSetter = field.HasSetter
					cls.AddField(inheritedField)
				} else {
					// If class has the field, inherit accessor flags from module
					cls.Fields[name].HasGetter = cls.Fields[name].HasGetter || field.HasGetter
					cls.Fields[name].HasSetter = cls.Fields[name].HasSetter || field.HasSetter
				}
			}
			// Add module methods (with FromModule tracking)
			// Only add if class doesn't already define the method (class methods take precedence)
			for _, method := range mod.Methods {
				if cls.Methods[method.Name] != nil {
					// Class already defines this method - skip module version
					continue
				}
				// Create a copy with FromModule set (don't modify original)
				methodCopy := *method
				methodCopy.FromModule = modName
				mustDefine(classScope, &methodCopy)
				cls.AddMethod(&methodCopy)
			}
		}
	}

	// Analyze methods
	prevScope := a.scope
	a.scope = classScope
	for _, m := range c.Methods {
		a.analyzeMethodDecl(m)
	}
	a.scope = prevScope
}

func (a *Analyzer) analyzeMethodDecl(m *ast.MethodDecl) {
	// Create method scope
	methodScope := NewScope(ScopeMethod, a.scope)
	methodScope.ClassName = a.scope.ClassName
	methodScope.IsClassMethod = m.IsClassMethod
	methodScope.ReturnTypes = make([]*Type, len(m.ReturnTypes))
	for i, rt := range m.ReturnTypes {
		methodScope.ReturnTypes[i] = ParseType(rt)
	}

	// Add parameters
	for _, p := range m.Params {
		var typ *Type
		if p.Type != "" {
			typ = ParseType(p.Type)
		} else {
			typ = TypeUnknownVal
		}
		param := NewParam(p.Name, typ)
		param.Line = m.Line
		mustDefine(methodScope, param)
	}

	// Analyze body
	prevScope := a.scope
	a.scope = methodScope
	for _, stmt := range m.Body {
		a.analyzeStatement(stmt)
	}
	a.scope = prevScope
}

func (a *Analyzer) analyzeAssign(s *ast.AssignStmt) {
	// Check if this is a new variable with Array<Any> type annotation
	// In this case, we need to analyze the array literal allowing heterogeneous elements
	var valueType *Type
	if s.Type != "" {
		declaredType := ParseType(s.Type)
		// Resolve type alias if applicable
		declaredType = a.resolveTypeAliasType(declaredType)
		// Resolve interface and struct types
		a.resolveType(declaredType)

		if declaredType.Kind == TypeArray && declaredType.Elem != nil && declaredType.Elem.Kind == TypeAny {
			if arr, ok := s.Value.(*ast.ArrayLit); ok {
				// Analyze array literal allowing heterogeneous elements
				valueType = a.analyzeArrayLitWithExpectedType(arr, TypeAnyVal)
			}
		}

		// Check if this is a lambda being assigned to a function type
		// In this case, infer parameter/return types from the expected type
		if declaredType.Kind == TypeFunc {
			if lambda, ok := s.Value.(*ast.LambdaExpr); ok {
				valueType = a.analyzeLambdaWithExpectedType(lambda, declaredType)
			}
		}
	}
	// Default: analyze without expected type context
	if valueType == nil {
		valueType = a.analyzeExpr(s.Value)
	}

	// Handle blank identifier specially - always use = not :=
	if s.Name == "_" {
		a.declarations[s] = false
		return
	}

	// Check if variable already exists in scope
	existing := a.scope.Lookup(s.Name)

	// Determine if this should be a reassignment or a new declaration
	isReassignment := false
	if existing != nil {
		// Allow shadowing of built-in functions (like Ruby allows x = 5 even though x is a method)
		if existing.Builtin {
			// Treat as new declaration, shadowing the built-in
			isReassignment = false
		} else if existing.Kind == SymConstant {
			a.addError(&Error{
				Message: "cannot assign to constant '" + s.Name + "'",
				Line:    s.Line,
			})
			return
		} else if existing.Kind == SymVariable || existing.Kind == SymParam {
			// Variable or parameter exists - treat as reassignment
			isReassignment = true
		} else if existing.Kind == SymFunction {
			// User-defined function - allow shadowing with warning could be added here
			// For now, treat as new declaration (shadows the function)
			isReassignment = false
		}
	}

	if isReassignment {
		// Assignment to existing variable - check type compatibility
		a.declarations[s] = false // not a declaration, it's a reassignment
		if existing.Type.Kind != TypeUnknown && valueType.Kind != TypeUnknown {
			if !a.isAssignable(existing.Type, valueType) {
				a.addError(&TypeMismatchError{
					Expected: existing.Type,
					Got:      valueType,
					Line:     s.Line,
					Context:  "assignment",
				})
			}
		} else if existing.Type.Kind == TypeUnknown && valueType.Kind != TypeUnknown {
			// Update type if we now know it
			a.scope.Update(s.Name, valueType)
		}
	} else {
		// New variable declaration
		a.declarations[s] = true // this is a declaration
		var typ *Type
		if s.Type != "" {
			typ = ParseType(s.Type)
			// Resolve interface and struct types
			a.resolveType(typ)
			// Check that value is assignable to declared type
			if valueType.Kind != TypeUnknown && !a.isAssignable(typ, valueType) {
				a.addError(&TypeMismatchError{
					Expected: typ,
					Got:      valueType,
					Line:     s.Line,
					Context:  "assignment",
				})
			}

			// For empty array/map literals, propagate the expected type to the value node
			// This allows codegen to generate properly typed literals (e.g., []int{} not []any{})
			if arr, ok := s.Value.(*ast.ArrayLit); ok && len(arr.Elements) == 0 {
				a.setNodeType(arr, typ)
			}
			if mp, ok := s.Value.(*ast.MapLit); ok && len(mp.Entries) == 0 {
				a.setNodeType(mp, typ)
			}
		} else {
			typ = valueType
		}

		v := NewVariable(s.Name, typ)
		v.Line = s.Line
		_ = a.scope.DefineOrShadow(v)
	}

	a.setNodeType(s, valueType)
}

func (a *Analyzer) analyzeMultiAssign(s *ast.MultiAssignStmt) {
	valueType := a.analyzeExpr(s.Value)

	// Initialize tracking for symbols declared at this node
	a.declaredAt[s] = make(map[string]*Symbol)
	hasNewVar := false // track if any variable is new

	// For tuple unpacking, we need to match the number of names to tuple elements
	if valueType.Kind == TypeTuple {
		if len(s.Names) != len(valueType.Elements) {
			a.addError(&Error{
				Message: "wrong number of values in multi-assignment",
			})
			return
		}

		for i, name := range s.Names {
			if name == "_" {
				continue
			}
			elemType := valueType.Elements[i]

			existing := a.scope.LookupLocal(name)
			if existing == nil {
				v := NewVariable(name, elemType)
				_ = a.scope.DefineOrShadow(v)
				a.declaredAt[s][name] = v
				hasNewVar = true
			}
		}
	} else if valueType.Kind == TypeOptional {
		// Special case: val, ok = optional_expr
		if len(s.Names) == 2 {
			valName, okName := s.Names[0], s.Names[1]
			innerType := valueType.Elem
			if innerType == nil {
				innerType = TypeAnyVal
			}

			if valName != "_" {
				existing := a.scope.LookupLocal(valName)
				if existing == nil {
					v := NewVariable(valName, innerType)
					_ = a.scope.DefineOrShadow(v)
					a.declaredAt[s][valName] = v
					hasNewVar = true
				}
			}
			if okName != "_" {
				existing := a.scope.LookupLocal(okName)
				if existing == nil {
					v := NewVariable(okName, TypeBoolVal)
					_ = a.scope.DefineOrShadow(v)
					a.declaredAt[s][okName] = v
					hasNewVar = true
				}
			}
		}
	} else {
		// For Go interop or unknown types, define all variables with unknown types.
		// This supports patterns like: result, err = strconv.Atoi("123")
		// where we don't know the exact return types but need to allow multi-value assignment.
		for _, name := range s.Names {
			if name == "_" {
				continue
			}
			existing := a.scope.LookupLocal(name)
			if existing == nil {
				v := NewVariable(name, TypeAnyVal)
				_ = a.scope.DefineOrShadow(v)
				a.declaredAt[s][name] = v
				hasNewVar = true
			}
		}
	}

	// Mark whether this multi-assign declares any new variables (for := vs =)
	a.declarations[s] = hasNewVar
}

func (a *Analyzer) analyzeMapDestructuring(s *ast.MapDestructuringStmt) {
	// Analyze the map expression
	a.analyzeExpr(s.Value)

	// Initialize tracking for symbols declared at this node
	a.declaredAt[s] = make(map[string]*Symbol)
	hasNewVar := false

	// Define variables for each key-variable pair
	for _, pair := range s.Pairs {
		name := pair.Variable
		if name == "_" {
			continue // skip blank identifier
		}
		existing := a.scope.LookupLocal(name)
		if existing == nil {
			// Map values are dynamically typed - use any
			v := NewVariable(name, TypeAnyVal)
			_ = a.scope.DefineOrShadow(v)
			a.declaredAt[s][name] = v
			hasNewVar = true
		}
	}

	a.declarations[s] = hasNewVar
}

func (a *Analyzer) analyzeCompoundAssign(s *ast.CompoundAssignStmt) {
	// Variable must exist
	existing := a.scope.Lookup(s.Name)
	if existing == nil {
		a.addError(&UndefinedError{Name: s.Name})
		return
	}

	// Check if trying to reassign a constant
	if existing.Kind == SymConstant {
		a.addError(&Error{
			Message: "cannot assign to constant '" + s.Name + "'",
			Line:    s.Line,
		})
		return
	}

	valueType := a.analyzeExpr(s.Value)
	varType := existing.Type

	// Skip if either type is unknown
	if varType.Kind == TypeUnknown || varType.Kind == TypeAny ||
		valueType.Kind == TypeUnknown || valueType.Kind == TypeAny {
		return
	}

	// Validate types based on operator
	switch s.Op {
	case "+":
		// += works for numeric types and string concatenation
		if varType.Kind == TypeString && valueType.Kind == TypeString {
			return // string concatenation OK
		}
		if !a.isNumeric(varType) || !a.isNumeric(valueType) {
			a.addError(&OperatorTypeMismatchError{
				Op:        "+=",
				LeftType:  varType,
				RightType: valueType,
			})
		}
	case "-", "*", "/", "%":
		// These only work for numeric types
		if !a.isNumeric(varType) || !a.isNumeric(valueType) {
			a.addError(&OperatorTypeMismatchError{
				Op:        s.Op + "=",
				LeftType:  varType,
				RightType: valueType,
			})
		}
	}
}

func (a *Analyzer) analyzeOrAssign(s *ast.OrAssignStmt) {
	valueType := a.analyzeExpr(s.Value)

	// Variable may or may not exist
	existing := a.scope.Lookup(s.Name)
	if existing == nil {
		// First use - this is a declaration
		a.declarations[s] = true
		v := NewVariable(s.Name, valueType)
		_ = a.scope.DefineOrShadow(v)
	} else {
		// Check if trying to reassign a constant
		if existing.Kind == SymConstant {
			a.addError(&Error{
				Message: "cannot assign to constant '" + s.Name + "'",
				Line:    s.Line,
			})
			return
		}
		// Variable already exists - this is a reassignment (nil check)
		a.declarations[s] = false
	}
}

func (a *Analyzer) analyzeSelectorAssign(s *ast.SelectorAssignStmt) {
	// Analyze the receiver object
	receiverType := a.analyzeExpr(s.Object)

	// Check if trying to assign to a struct field (structs are immutable)
	if receiverType.Kind == TypeClass && receiverType.Name != "" {
		if a.structs[receiverType.Name] {
			a.addError(&Error{
				Line:    s.Line,
				Message: "cannot modify struct field '" + s.Field + "' (structs are immutable)",
			})
			return
		}
	}

	// Analyze the value being assigned
	valueType := a.analyzeExpr(s.Value)

	// Verify the receiver type has a setter for this field
	if receiverType.Kind == TypeClass && receiverType.Name != "" {
		if cls := a.classes[receiverType.Name]; cls != nil {
			// First check for a field with a setter accessor
			if field := cls.GetField(s.Field); field != nil {
				// Check if the field has a setter
				if !field.HasSetter {
					a.addError(&Error{
						Line:    s.Line,
						Message: "cannot assign to field '" + s.Field + "' without setter",
					})
				}
				// Check type compatibility
				if !a.isAssignable(field.Type, valueType) {
					a.addError(&TypeMismatchError{
						Expected: field.Type,
						Got:      valueType,
						Line:     s.Line,
						Context:  "setter assignment",
					})
				}
				return
			}

			// Check for custom setter method: def field=(value)
			setterName := s.Field + "="
			if setter := cls.GetMethod(setterName); setter != nil {
				// Setter methods should take exactly one parameter
				if len(setter.Params) == 1 {
					// Check type compatibility with the parameter
					if !a.isAssignable(setter.Params[0].Type, valueType) {
						a.addError(&TypeMismatchError{
							Expected: setter.Params[0].Type,
							Got:      valueType,
							Line:     s.Line,
							Context:  "setter assignment",
						})
					}
				}
				return
			}

			// Neither field nor custom setter found
			a.addError(&UndefinedError{
				Name: s.Field,
				Line: s.Line,
			})
		}
	}
}

func (a *Analyzer) analyzeIndexAssign(s *ast.IndexAssignStmt) {
	// Analyze the collection being indexed
	collectionType := a.analyzeExpr(s.Left)
	// Analyze the index/key
	a.analyzeExpr(s.Index)
	// Analyze the value being assigned
	a.analyzeExpr(s.Value)

	// Set the type for the statement (for codegen to use)
	a.setNodeType(s, collectionType)
}

func (a *Analyzer) analyzeIf(s *ast.IfStmt) {
	context := "if"
	if s.IsUnless {
		context = "unless"
	}

	// Handle assignment in condition (if let pattern)
	if s.AssignName != "" && s.AssignExpr != nil {
		exprType := a.analyzeExpr(s.AssignExpr)

		// Create scope for the then branch that includes the bound variable
		thenScope := NewScope(ScopeBlock, a.scope)
		var boundType *Type
		if exprType.Kind == TypeOptional {
			boundType = exprType.Elem
		} else {
			boundType = exprType
		}
		if boundType == nil {
			boundType = TypeUnknownVal
		}
		mustDefine(thenScope, NewVariable(s.AssignName, boundType))

		prevScope := a.scope
		a.scope = thenScope
		for _, stmt := range s.Then {
			a.analyzeStatement(stmt)
		}
		a.scope = prevScope
	} else {
		// Check that condition is Bool
		a.checkCondition(s.Cond, context, s.Line)

		// Analyze then branch
		a.pushScope()
		for _, stmt := range s.Then {
			a.analyzeStatement(stmt)
		}
		a.popScope()
	}

	// Analyze elsif branches
	for _, elsif := range s.ElseIfs {
		a.checkCondition(elsif.Cond, "elsif", 0) // ElseIfClause doesn't have Line field
		a.pushScope()
		for _, stmt := range elsif.Body {
			a.analyzeStatement(stmt)
		}
		a.popScope()
	}

	// Analyze else branch
	if len(s.Else) > 0 {
		a.pushScope()
		for _, stmt := range s.Else {
			a.analyzeStatement(stmt)
		}
		a.popScope()
	}
}

func (a *Analyzer) analyzeWhile(s *ast.WhileStmt) {
	a.checkCondition(s.Cond, "while", s.Line)

	a.pushLoopScope()
	for _, stmt := range s.Body {
		a.analyzeStatement(stmt)
	}
	a.popScope()
}

func (a *Analyzer) analyzeUntil(s *ast.UntilStmt) {
	a.checkCondition(s.Cond, "until", s.Line)

	a.pushLoopScope()
	for _, stmt := range s.Body {
		a.analyzeStatement(stmt)
	}
	a.popScope()
}

func (a *Analyzer) analyzeLoop(s *ast.LoopStmt) {
	a.pushLoopScope()
	for _, stmt := range s.Body {
		a.analyzeStatement(stmt)
	}
	a.popScope()
}

func (a *Analyzer) analyzeFor(s *ast.ForStmt) {
	iterableType := a.analyzeExpr(s.Iterable)

	// Determine loop variable type from iterable
	var elemType *Type
	var valueType *Type // for maps with two-variable syntax
	switch iterableType.Kind {
	case TypeArray:
		elemType = iterableType.Elem
	case TypeRange:
		elemType = TypeIntVal
	case TypeChan:
		elemType = iterableType.Elem
	case TypeMap:
		// For maps, the first variable is the key
		elemType = iterableType.KeyType
		valueType = iterableType.Elem
	case TypeString:
		elemType = TypeStringVal // iterating runes as strings
	default:
		elemType = TypeAnyVal
	}
	if elemType == nil {
		elemType = TypeAnyVal
	}

	// Create loop scope with loop variable(s)
	loopScope := NewScope(ScopeLoop, a.scope)
	mustDefine(loopScope, NewVariable(s.Var, elemType))

	// If there's a second variable (for key, value in map), add it too
	if s.Var2 != "" {
		if valueType != nil {
			mustDefine(loopScope, NewVariable(s.Var2, valueType))
		} else {
			mustDefine(loopScope, NewVariable(s.Var2, TypeAnyVal))
		}
	}

	prevScope := a.scope
	a.scope = loopScope
	for _, stmt := range s.Body {
		a.analyzeStatement(stmt)
	}
	a.scope = prevScope
}

func (a *Analyzer) analyzeReturn(s *ast.ReturnStmt) {
	fnScope := a.scope.FunctionScope()
	if fnScope == nil {
		a.addError(&ReturnOutsideFunctionError{Line: s.Line})
		return
	}

	// Disallow return inside iterator blocks (each, map, select, etc.)
	// Use next or break instead - return is a false friend from Ruby
	if a.scope.IsInsideIterator() {
		a.addError(&ReturnInsideBlockError{Line: s.Line})
		return
	}

	// Analyze return values and collect their types
	var returnTypes []*Type
	for _, val := range s.Values {
		typ := a.analyzeExpr(val)
		returnTypes = append(returnTypes, typ)
	}

	// Validate return types against declared return types
	expectedTypes := fnScope.ReturnTypes
	if len(expectedTypes) > 0 {
		// Check each return value against expected type
		for i, got := range returnTypes {
			if i >= len(expectedTypes) {
				break // Extra return values (could be another error)
			}
			expected := expectedTypes[i]
			if !a.isAssignable(expected, got) {
				a.addError(&ReturnTypeMismatchError{
					Expected: expected,
					Got:      got,
					Line:     s.Line,
				})
			}
		}
	}

	// Analyze condition if present
	if s.Condition != nil {
		a.analyzeExpr(s.Condition)
	}
}

func (a *Analyzer) analyzeBreak(s *ast.BreakStmt) {
	// break is not allowed inside iterator blocks (each, map, etc.)
	if a.scope.IsInsideIterator() {
		a.addError(&BreakInsideBlockError{Line: s.Line})
	} else if !a.scope.IsInsideLoop() {
		a.addError(&BreakOutsideLoopError{})
	}

	if s.Condition != nil {
		a.analyzeExpr(s.Condition)
	}
}

func (a *Analyzer) analyzeNext(s *ast.NextStmt) {
	// next is not allowed inside iterator blocks (each, map, etc.)
	if a.scope.IsInsideIterator() {
		a.addError(&NextInsideBlockError{Line: s.Line})
	} else if !a.scope.IsInsideLoop() {
		a.addError(&NextOutsideLoopError{Line: s.Line})
	}

	if s.Condition != nil {
		a.analyzeExpr(s.Condition)
	}
}

func (a *Analyzer) analyzeCase(s *ast.CaseStmt) {
	var subjectType *Type
	if s.Subject != nil {
		subjectType = a.analyzeExpr(s.Subject)
	}

	for _, when := range s.WhenClauses {
		for _, val := range when.Values {
			valType := a.analyzeExpr(val)

			// Validate when value type matches subject type
			if subjectType != nil && subjectType.Kind != TypeUnknown && subjectType.Kind != TypeAny &&
				valType.Kind != TypeUnknown && valType.Kind != TypeAny {
				// Allow Range values when subject is numeric (for "in range" matching)
				if valType.Kind == TypeRange && a.isNumeric(subjectType) {
					continue // Range matching is valid for numeric subjects
				}
				if !a.areTypesComparable(subjectType, valType) {
					a.addError(&TypeMismatchError{
						Expected: subjectType,
						Got:      valType,
						Context:  "case when value",
					})
				}
			}
		}
		a.pushScope()
		for _, stmt := range when.Body {
			a.analyzeStatement(stmt)
		}
		a.popScope()
	}

	if len(s.Else) > 0 {
		a.pushScope()
		for _, stmt := range s.Else {
			a.analyzeStatement(stmt)
		}
		a.popScope()
	}
}

func (a *Analyzer) analyzeCaseType(s *ast.CaseTypeStmt) {
	subjectType := a.analyzeExpr(s.Subject)

	for _, when := range s.WhenClauses {
		a.pushScope()

		narrowedType := ParseType(when.Type)

		// If binding variable is specified, define it with the narrowed type
		if when.BindingVar != "" {
			_ = a.scope.DefineOrShadow(NewVariable(when.BindingVar, narrowedType))
		}

		// Also shadow the subject identifier with the narrowed type (for backwards compat)
		if ident, ok := s.Subject.(*ast.Ident); ok {
			_ = a.scope.DefineOrShadow(NewVariable(ident.Name, narrowedType))
		}

		for _, stmt := range when.Body {
			a.analyzeStatement(stmt)
		}
		a.popScope()
	}

	if len(s.Else) > 0 {
		a.pushScope()
		for _, stmt := range s.Else {
			a.analyzeStatement(stmt)
		}
		a.popScope()
	}

	a.setNodeType(s.Subject, subjectType)
}

func (a *Analyzer) analyzeInstanceVarAssign(s *ast.InstanceVarAssign) {
	if !a.scope.IsInsideClass() {
		a.addError(&InstanceVarOutsideClassError{Name: s.Name})
		return
	}
	if a.scope.IsInsideClassMethod() {
		a.addError(&InstanceStateInClassMethodError{What: "@" + s.Name})
		return
	}

	valueType := a.analyzeExpr(s.Value)

	// If the field exists but has unknown type, infer it from the assigned value
	if classScope := a.scope.ClassScope(); classScope != nil {
		if cls := a.classes[classScope.ClassName]; cls != nil {
			if field := cls.GetField(s.Name); field != nil {
				if field.Type != nil && field.Type.Kind == TypeUnknown && valueType != nil && valueType.Kind != TypeUnknown {
					field.Type = valueType
				}
			}
		}
	}
}

func (a *Analyzer) analyzeInstanceVarCompoundAssign(s *ast.InstanceVarCompoundAssign) {
	if !a.scope.IsInsideClass() {
		a.addError(&InstanceVarOutsideClassError{Name: s.Name})
		return
	}
	if a.scope.IsInsideClassMethod() {
		a.addError(&InstanceStateInClassMethodError{What: "@" + s.Name, Line: s.Line})
		return
	}

	// Look up the field in the current class
	var fieldType *Type
	if classScope := a.scope.ClassScope(); classScope != nil {
		if cls := a.classes[classScope.ClassName]; cls != nil {
			if field := cls.GetField(s.Name); field != nil {
				fieldType = field.Type
			}
		}
	}

	valueType := a.analyzeExpr(s.Value)

	// Skip type checking if field type is unknown or not found
	// (field might be dynamically created in initialize)
	if fieldType == nil || fieldType.Kind == TypeUnknown || fieldType.Kind == TypeAny ||
		valueType.Kind == TypeUnknown || valueType.Kind == TypeAny {
		return
	}

	// Validate types based on operator
	switch s.Op {
	case "+":
		// += works for numeric types and string concatenation
		if fieldType.Kind == TypeString && valueType.Kind == TypeString {
			return // string concatenation OK
		}
		if !a.isNumeric(fieldType) || !a.isNumeric(valueType) {
			a.addError(&OperatorTypeMismatchError{
				Op:        "+=",
				LeftType:  fieldType,
				RightType: valueType,
			})
		}
	case "-", "*", "/":
		// These only work for numeric types
		if !a.isNumeric(fieldType) || !a.isNumeric(valueType) {
			a.addError(&OperatorTypeMismatchError{
				Op:        s.Op + "=",
				LeftType:  fieldType,
				RightType: valueType,
			})
		}
	}
}

func (a *Analyzer) analyzeGo(s *ast.GoStmt) {
	if s.Call != nil {
		a.analyzeExpr(s.Call)
	}
	if s.Block != nil {
		a.analyzeBlock(s.Block)
	}
}

func (a *Analyzer) analyzeChanSend(s *ast.ChanSendStmt) {
	chanType := a.analyzeExpr(s.Chan)
	valueType := a.analyzeExpr(s.Value)

	// Validate channel type
	if chanType.Kind != TypeChan && chanType.Kind != TypeUnknown && chanType.Kind != TypeAny {
		a.addError(&TypeMismatchError{
			Expected: &Type{Kind: TypeChan},
			Got:      chanType,
			Context:  "channel send",
			Line:     s.Line,
		})
		return
	}

	// Validate value type matches channel element type
	if chanType.Kind == TypeChan && chanType.Elem != nil {
		elemType := chanType.Elem
		if elemType.Kind != TypeUnknown && elemType.Kind != TypeAny &&
			valueType.Kind != TypeUnknown && valueType.Kind != TypeAny {
			if !a.isAssignable(elemType, valueType) {
				a.addError(&TypeMismatchError{
					Expected: elemType,
					Got:      valueType,
					Context:  "channel send value",
					Line:     s.Line,
				})
			}
		}
	}
}

func (a *Analyzer) analyzeSelect(s *ast.SelectStmt) {
	for _, c := range s.Cases {
		if c.Chan != nil {
			a.analyzeExpr(c.Chan)
		}
		if c.Value != nil {
			a.analyzeExpr(c.Value)
		}

		// Create scope for case body
		caseScope := NewScope(ScopeBlock, a.scope)
		if c.AssignName != "" {
			// Infer type from channel
			chanType := a.getNodeType(c.Chan)
			var elemType *Type
			if chanType != nil && chanType.Kind == TypeChan {
				elemType = chanType.Elem
			} else {
				elemType = TypeAnyVal
			}
			mustDefine(caseScope, NewVariable(c.AssignName, elemType))
		}

		prevScope := a.scope
		a.scope = caseScope
		for _, stmt := range c.Body {
			a.analyzeStatement(stmt)
		}
		a.scope = prevScope
	}

	if len(s.Else) > 0 {
		a.pushScope()
		for _, stmt := range s.Else {
			a.analyzeStatement(stmt)
		}
		a.popScope()
	}
}

// Test statement analysis functions

func (a *Analyzer) analyzeDescribe(s *ast.DescribeStmt) {
	a.pushScope()
	for _, stmt := range s.Body {
		a.analyzeStatement(stmt)
	}
	a.popScope()
}

func (a *Analyzer) analyzeIt(s *ast.ItStmt) {
	a.pushScope()
	for _, stmt := range s.Body {
		a.analyzeStatement(stmt)
	}
	a.popScope()
}

func (a *Analyzer) analyzeTest(s *ast.TestStmt) {
	a.pushScope()
	for _, stmt := range s.Body {
		a.analyzeStatement(stmt)
	}
	a.popScope()
}

func (a *Analyzer) analyzeTable(s *ast.TableStmt) {
	a.pushScope()
	for _, stmt := range s.Body {
		a.analyzeStatement(stmt)
	}
	a.popScope()
}

func (a *Analyzer) analyzeBefore(s *ast.BeforeStmt) {
	a.pushScope()
	for _, stmt := range s.Body {
		a.analyzeStatement(stmt)
	}
	a.popScope()
}

func (a *Analyzer) analyzeAfter(s *ast.AfterStmt) {
	a.pushScope()
	for _, stmt := range s.Body {
		a.analyzeStatement(stmt)
	}
	a.popScope()
}

// analyzeArrayLitWithExpectedType analyzes an array literal when the expected element type
// is known (e.g., from a type annotation like `Array<Any>`). When expectedElemType is Any,
// heterogeneous elements are allowed without type mismatch errors.
func (a *Analyzer) analyzeArrayLitWithExpectedType(arr *ast.ArrayLit, expectedElemType *Type) *Type {
	for _, elem := range arr.Elements {
		// Analyze each element to check for errors, but don't validate type compatibility
		// when expected element type is Any
		if _, isSplat := elem.(*ast.SplatExpr); isSplat {
			a.analyzeExpr(elem)
			continue
		}
		a.analyzeExpr(elem)
	}
	typ := NewArrayType(expectedElemType)
	a.setNodeType(arr, typ)
	return typ
}

// analyzeExpr analyzes an expression and returns its type.
func (a *Analyzer) analyzeExpr(expr ast.Expression) *Type {
	if expr == nil {
		return TypeUnknownVal
	}

	var typ *Type

	switch e := expr.(type) {
	case *ast.IntLit:
		typ = TypeIntVal
	case *ast.FloatLit:
		typ = TypeFloatVal
	case *ast.StringLit:
		typ = TypeStringVal
	case *ast.BoolLit:
		typ = TypeBoolVal
	case *ast.NilLit:
		typ = TypeNilVal
	case *ast.SymbolLit:
		typ = TypeStringVal

	case *ast.Ident:
		// Check for 'self' access in class methods
		if e.Name == "self" && a.scope.IsInsideClassMethod() {
			a.addError(&InstanceStateInClassMethodError{What: "self", Line: e.Line})
			typ = TypeUnknownVal
			break
		}

		sym := a.scope.Lookup(e.Name)
		// Check top-level functions as fallback (handles forward references)
		if sym == nil {
			if fn := a.functions[e.Name]; fn != nil {
				sym = fn
			}
		}
		// Check methods of the current class (implicit self.method call)
		if sym == nil && a.scope.IsInsideClass() {
			className := a.scope.ClassName
			if cls := a.classes[className]; cls != nil {
				if method := cls.GetMethod(e.Name); method != nil {
					sym = method
				}
			}
		}
		// Check if this is an enum type reference (for static method calls like Color.values)
		if sym == nil {
			if _, isEnum := a.enums[e.Name]; isEnum {
				typ = &Type{Kind: TypeClass, Name: e.Name}
				break
			}
		}
		if sym == nil {
			a.addError(&UndefinedError{Name: e.Name, Line: e.Line, Column: e.Column, Candidates: a.findSimilar(e.Name)})
			typ = TypeUnknownVal
		} else if sym.Kind == SymFunction || sym.Kind == SymMethod {
			// In Rugby, no-arg functions/methods are called implicitly (like Ruby)
			// Functions/methods with required params must use parentheses
			// Variadic functions can be called with 0 args, so they're allowed
			if len(sym.Params) > 0 && !sym.Variadic {
				a.addError(&MethodRequiresArgumentsError{
					MethodName: e.Name,
					ParamCount: len(sym.Params),
				})
				typ = TypeUnknownVal
			} else if len(sym.ReturnTypes) == 1 {
				typ = sym.ReturnTypes[0]
			} else if len(sym.ReturnTypes) > 1 {
				typ = NewTupleType(sym.ReturnTypes...)
			} else {
				typ = TypeUnknownVal
			}
		} else {
			typ = sym.Type
			// Mark variable as used (read)
			if sym.Kind == SymVariable || sym.Kind == SymParam {
				a.usedSymbols[sym] = true
			}
		}

	case *ast.InstanceVar:
		if !a.scope.IsInsideClass() {
			a.addError(&InstanceVarOutsideClassError{Name: e.Name})
			typ = TypeUnknownVal
		} else if a.scope.IsInsideClassMethod() {
			a.addError(&InstanceStateInClassMethodError{What: "@" + e.Name})
			typ = TypeUnknownVal
		} else {
			// Look up field type from class, including inherited fields
			classScope := a.scope.ClassScope()
			if classScope != nil {
				if field := a.getClassField(classScope.ClassName, e.Name); field != nil {
					typ = field.Type
				}
			}
			if typ == nil {
				typ = TypeUnknownVal
			}
		}

	case *ast.ArrayLit:
		var elemType *Type
		for _, elem := range e.Elements {
			// Skip splat expressions - they expand to multiple elements
			if _, isSplat := elem.(*ast.SplatExpr); isSplat {
				a.analyzeExpr(elem)
				continue
			}

			et := a.analyzeExpr(elem)
			if elemType == nil {
				elemType = et
			} else if elemType.Kind != TypeUnknown && elemType.Kind != TypeAny &&
				et.Kind != TypeUnknown && et.Kind != TypeAny {
				// Validate element type matches (allow numeric widening)
				if !a.isAssignable(elemType, et) && !a.isAssignable(et, elemType) {
					a.addError(&TypeMismatchError{
						Expected: elemType,
						Got:      et,
						Context:  "array element",
					})
				}
			}
		}
		if elemType == nil {
			elemType = TypeAnyVal
		}
		typ = NewArrayType(elemType)

	case *ast.MapLit:
		var keyType, valueType *Type
		for _, entry := range e.Entries {
			if entry.Splat != nil {
				a.analyzeExpr(entry.Splat)
				continue
			}

			if entry.Key != nil {
				kt := a.analyzeExpr(entry.Key)
				if keyType == nil {
					keyType = kt
				} else if keyType.Kind != TypeUnknown && keyType.Kind != TypeAny &&
					kt.Kind != TypeUnknown && kt.Kind != TypeAny {
					// Validate key type matches
					if !a.isAssignable(keyType, kt) && !a.isAssignable(kt, keyType) {
						a.addError(&TypeMismatchError{
							Expected: keyType,
							Got:      kt,
							Context:  "map key",
						})
					}
				}
			}
			if entry.Value != nil {
				vt := a.analyzeExpr(entry.Value)
				if valueType == nil {
					valueType = vt
				} else if valueType.Kind != TypeUnknown && valueType.Kind != TypeAny &&
					vt.Kind != TypeUnknown && vt.Kind != TypeAny {
					// If value types don't match, widen to any (heterogeneous map)
					if !a.isAssignable(valueType, vt) && !a.isAssignable(vt, valueType) {
						valueType = TypeAnyVal
					}
				}
			}
		}
		if keyType == nil {
			keyType = TypeAnyVal
		}
		if valueType == nil {
			valueType = TypeAnyVal
		}
		typ = NewMapType(keyType, valueType)

	case *ast.RangeLit:
		a.analyzeExpr(e.Start)
		a.analyzeExpr(e.End)
		typ = TypeRangeVal

	case *ast.TupleLit:
		elemTypes := make([]*Type, len(e.Elements))
		for i, elem := range e.Elements {
			elemTypes[i] = a.analyzeExpr(elem)
		}
		typ = NewTupleType(elemTypes...)

	case *ast.StructLit:
		// Struct literal: Point{x: 10, y: 20}
		// Analyze field values
		for _, field := range e.Fields {
			if field.Value != nil {
				a.analyzeExpr(field.Value)
			}
		}
		// Return type as struct (using TypeClass kind with IsStruct flag)
		typ = &Type{Kind: TypeClass, Name: e.Name, IsStruct: a.structs[e.Name]}

	case *ast.SetLit:
		var elemType *Type
		for _, elem := range e.Elements {
			et := a.analyzeExpr(elem)
			if elemType == nil {
				elemType = et
			} else if elemType.Kind != TypeUnknown && et.Kind != TypeUnknown {
				// Validate element types are compatible
				if !a.isAssignable(elemType, et) && !a.isAssignable(et, elemType) {
					a.addError(&TypeMismatchError{
						Expected: elemType,
						Got:      et,
						Line:     e.Line,
					})
				}
			}
		}
		if elemType == nil {
			elemType = &Type{Kind: TypeAny}
		}
		typ = NewSetType(elemType)

	case *ast.BinaryExpr:
		leftType := a.analyzeExpr(e.Left)
		rightType := a.analyzeExpr(e.Right)
		typ = a.checkBinaryExpr(e.Op, leftType, rightType)

	case *ast.UnaryExpr:
		operandType := a.analyzeExpr(e.Expr)
		typ = a.checkUnaryExpr(e.Op, operandType)

	case *ast.CallExpr:
		typ = a.analyzeCall(e)

	case *ast.SelectorExpr:
		receiverType := a.analyzeExpr(e.X)
		var selectorKind ast.SelectorKind

		// Check if receiver is a Go package (for Go interop)
		isGoPackage := false
		if ident, ok := e.X.(*ast.Ident); ok {
			if sym := a.scope.Lookup(ident.Name); sym != nil && sym.Kind == SymGoPackage {
				isGoPackage = true
			}
		}

		// Special case: ClassName.new is a constructor (returns instance of class)
		if e.Sel == "new" {
			if classIdent, ok := e.X.(*ast.Ident); ok {
				if _, isClass := a.classes[classIdent.Name]; isClass {
					typ = NewClassType(classIdent.Name)
					selectorKind = ast.SelectorMethod // Constructor is a method call
				}
			}
		}

		// Check for class method: ClassName.method_name
		if typ == nil {
			if classIdent, ok := e.X.(*ast.Ident); ok {
				if cls, isClass := a.classes[classIdent.Name]; isClass {
					if classMethod := cls.GetClassMethod(e.Sel); classMethod != nil {
						// Class method with no params can be called implicitly
						if len(classMethod.Params) > 0 && !classMethod.Variadic {
							a.addError(&MethodRequiresArgumentsError{
								MethodName: e.Sel,
								ParamCount: len(classMethod.Params),
							})
							typ = TypeUnknownVal
						} else if len(classMethod.ReturnTypes) == 1 {
							typ = classMethod.ReturnTypes[0]
						} else if len(classMethod.ReturnTypes) > 1 {
							typ = NewTupleType(classMethod.ReturnTypes...)
						} else {
							typ = TypeUnknownVal
						}
						selectorKind = ast.SelectorMethod
					}
				}
			}
		}

		// Check for enum type static methods (Color.values)
		if typ == nil {
			if enumIdent, ok := e.X.(*ast.Ident); ok {
				if _, isEnum := a.enums[enumIdent.Name]; isEnum {
					switch e.Sel {
					case "values":
						// Enum.values returns Array<EnumType>
						typ = NewArrayType(&Type{Kind: TypeClass, Name: enumIdent.Name})
						selectorKind = ast.SelectorMethod
					}
				}
			}
		}

		// Check for enum instance methods (Color::Red.to_s, Color::Red.value)
		if typ == nil && receiverType.Kind == TypeClass && receiverType.Name != "" {
			if _, isEnum := a.enums[receiverType.Name]; isEnum {
				switch e.Sel {
				case "to_s":
					typ = TypeStringVal
					selectorKind = ast.SelectorMethod
				case "value":
					typ = TypeIntVal
					selectorKind = ast.SelectorMethod
				}
			}
		}

		// Check for class field access (with getter)
		if typ == nil && receiverType.Kind == TypeClass && receiverType.Name != "" {
			if cls := a.classes[receiverType.Name]; cls != nil {
				if field := cls.GetField(e.Sel); field != nil {
					typ = field.Type
					// Determine if this is a getter or raw field access
					if field.HasGetter {
						selectorKind = ast.SelectorGetter
					} else {
						selectorKind = ast.SelectorField
					}
				}
			}
		}

		// If not a field, look up method
		if typ == nil {
			if method := a.lookupMethod(receiverType, e.Sel); method != nil {
				// Check for private method access from outside the class
				if method.Private {
					// Private methods can only be called from within the same class
					canAccess := a.scope.IsInsideClass() && a.scope.ClassName == receiverType.Name
					if !canAccess {
						a.addError(&PrivateMethodAccessError{
							MethodName: e.Sel,
							ClassName:  receiverType.Name,
							Line:       method.Line, // Use method's line since SelectorExpr doesn't have one
						})
					}
				}

				// Only allow implicit call for methods with no required params
				// Variadic methods can be called with 0 args, so they're allowed
				if len(method.Params) > 0 && !method.Variadic {
					a.addError(&MethodRequiresArgumentsError{
						MethodName: e.Sel,
						ParamCount: len(method.Params),
					})
					typ = TypeUnknownVal
				} else if len(method.ReturnTypes) == 1 {
					typ = method.ReturnTypes[0]
				} else if len(method.ReturnTypes) > 1 {
					typ = NewTupleType(method.ReturnTypes...)
				}
				// Set selector kind based on whether it's Go interop or Rugby method
				if isGoPackage {
					selectorKind = ast.SelectorGoMethod
				} else {
					selectorKind = ast.SelectorMethod
				}
			}
		}

		// Fallback: predicate methods return Bool and are treated as method calls
		if typ == nil {
			if strings.HasSuffix(e.Sel, "?") {
				typ = TypeBoolVal
				selectorKind = ast.SelectorMethod
			} else {
				typ = TypeUnknownVal
				// For Go packages, don't assume unknown selectors are methods -
				// they could be constants like time.Millisecond
				// Let codegen handle it based on context
			}
		}

		// Store the resolved selector kind
		if selectorKind != ast.SelectorUnknown {
			a.setSelectorKind(e, selectorKind)
		}

	case *ast.ScopeExpr:
		// Scope resolution: Module::Class or Enum::Value
		if modIdent, ok := e.Left.(*ast.Ident); ok {
			// Check if it's an enum value reference: Status::Active
			if enumDecl, isEnum := a.enums[modIdent.Name]; isEnum {
				// Verify the value exists in the enum
				found := false
				for _, v := range enumDecl.Values {
					if v.Name == e.Right {
						found = true
						break
					}
				}
				if found {
					// Return the enum type
					typ = &Type{Kind: TypeClass, Name: modIdent.Name}
				} else {
					a.addError(&UndefinedError{Name: modIdent.Name + "::" + e.Right, Line: modIdent.Line})
					typ = TypeUnknownVal
				}
			} else if _, isModule := a.modules[modIdent.Name]; isModule {
				// Look up the class within the module
				if cls, isClass := a.classes[e.Right]; isClass {
					typ = NewClassType(cls.Name)
				} else {
					typ = TypeUnknownVal
				}
			} else {
				typ = TypeUnknownVal
			}
		} else {
			typ = TypeUnknownVal
		}

	case *ast.IndexExpr:
		// Special case: Chan<Type> is a channel type constructor, not an index expression
		if ident, ok := e.Left.(*ast.Ident); ok && ident.Name == "Chan" {
			// Parse the element type from the index
			elemType := a.parseTypeFromExpr(e.Index)
			typ = NewChanType(elemType)
			a.setNodeType(expr, typ)
			return typ
		}

		leftType := a.analyzeExpr(e.Left)
		indexType := a.analyzeExpr(e.Index)

		switch leftType.Kind {
		case TypeArray:
			if indexType.Kind == TypeRange {
				// Array slice with range - returns a slice of same type
				typ = leftType
			} else {
				// Array index must be Int
				a.checkIndexType(indexType, e.Line)
				typ = leftType.Elem
			}
		case TypeMap:
			// Map key type is checked separately (not requiring Int)
			// Map access returns optional
			typ = NewOptionalType(leftType.ValueType)
		case TypeString:
			if indexType.Kind == TypeRange {
				// String slice with range - returns a string
				typ = TypeStringVal
			} else {
				// String index must be Int
				a.checkIndexType(indexType, e.Line)
				typ = TypeStringVal
			}
		default:
			typ = TypeUnknownVal
		}

	case *ast.TernaryExpr:
		condType := a.analyzeExpr(e.Condition)
		// Validate condition is Bool (consistent with if/while/until)
		if condType.Kind != TypeUnknown && condType.Kind != TypeAny && condType.Kind != TypeBool {
			a.addError(&ConditionTypeMismatchError{
				Got:     condType,
				Context: "ternary",
				Line:    e.Line,
			})
		}
		thenType := a.analyzeExpr(e.Then)
		elseType := a.analyzeExpr(e.Else)
		// Return the more specific type
		if thenType.Kind != TypeUnknown {
			typ = thenType
		} else {
			typ = elseType
		}

	case *ast.NilCoalesceExpr:
		leftType := a.analyzeExpr(e.Left)
		rightType := a.analyzeExpr(e.Right)

		// Validate left side is optional
		if leftType.Kind != TypeOptional && leftType.Kind != TypeUnknown && leftType.Kind != TypeAny {
			a.addError(&TypeMismatchError{
				Expected: &Type{Kind: TypeOptional},
				Got:      leftType,
				Context:  "nil coalesce left side",
				Line:     e.Line,
			})
		}

		// Validate right type is compatible with unwrapped left type
		if leftType.Kind == TypeOptional && leftType.Elem != nil {
			unwrappedType := leftType.Elem
			if unwrappedType.Kind != TypeUnknown && unwrappedType.Kind != TypeAny &&
				rightType.Kind != TypeUnknown && rightType.Kind != TypeAny {
				// Right value must be assignable to the unwrapped type (unidirectional)
				if !a.isAssignable(unwrappedType, rightType) {
					a.addError(&TypeMismatchError{
						Expected: unwrappedType,
						Got:      rightType,
						Context:  "nil coalesce default value",
						Line:     e.Line,
					})
				}
			}
			typ = unwrappedType
		} else {
			typ = rightType
		}

	case *ast.SafeNavExpr:
		recvType := a.analyzeExpr(e.Receiver)

		// Get the underlying type (unwrap if optional)
		innerType := recvType
		if recvType.Kind == TypeOptional && recvType.Elem != nil {
			innerType = recvType.Elem
		}

		// Look up the field or method on the inner type
		var resultType *Type
		if innerType.Kind == TypeClass && innerType.Name != "" {
			if cls := a.classes[innerType.Name]; cls != nil {
				// Check for field first
				if field := cls.GetField(e.Selector); field != nil {
					resultType = field.Type
				} else if method := cls.GetMethod(e.Selector); method != nil {
					// Only allow implicit call for methods with no required params
					// Variadic methods can be called with 0 args, so they're allowed
					if len(method.Params) > 0 && !method.Variadic {
						a.addError(&MethodRequiresArgumentsError{
							MethodName: e.Selector,
							ParamCount: len(method.Params),
						})
						resultType = TypeUnknownVal
					} else if len(method.ReturnTypes) == 1 {
						resultType = method.ReturnTypes[0]
					} else if len(method.ReturnTypes) > 1 {
						resultType = NewTupleType(method.ReturnTypes...)
					}
				}
			}
		}

		// Try builtin methods if no class field/method found
		if resultType == nil {
			if method := a.lookupBuiltinMethod(innerType, e.Selector); method != nil {
				// Only allow implicit call for methods with no required params
				// Variadic methods can be called with 0 args, so they're allowed
				if len(method.Params) > 0 && !method.Variadic {
					a.addError(&MethodRequiresArgumentsError{
						MethodName: e.Selector,
						ParamCount: len(method.Params),
					})
					resultType = TypeUnknownVal
				} else if len(method.ReturnTypes) == 1 {
					resultType = method.ReturnTypes[0]
				} else if len(method.ReturnTypes) > 1 {
					resultType = NewTupleType(method.ReturnTypes...)
				}
			}
		}

		// Default to unknown if we couldn't determine the type
		if resultType == nil {
			resultType = TypeUnknownVal
		}

		// Safe navigation always returns optional (unless result is already optional)
		if resultType.Kind == TypeOptional {
			typ = resultType
		} else {
			typ = NewOptionalType(resultType)
		}

	case *ast.BangExpr:
		innerType := a.analyzeExpr(e.Expr)

		// Validate that inner expression is an error tuple (T, error) or just error
		isErrorTuple := false
		if innerType.Kind == TypeTuple && len(innerType.Elements) >= 1 {
			lastElem := innerType.Elements[len(innerType.Elements)-1]
			if lastElem.Kind == TypeError {
				isErrorTuple = true
			}
		} else if innerType.Kind == TypeError {
			isErrorTuple = true
		}

		if !isErrorTuple && innerType.Kind != TypeUnknown && innerType.Kind != TypeAny {
			a.addError(&BangOnNonErrorError{
				Got:  innerType,
				Line: e.Line,
			})
		}

		// Validate that enclosing function returns error (or we're at top-level/main)
		fnScope := a.scope.FunctionScope()
		if fnScope != nil && len(fnScope.ReturnTypes) > 0 {
			lastReturn := fnScope.ReturnTypes[len(fnScope.ReturnTypes)-1]
			if lastReturn.Kind != TypeError && lastReturn.Kind != TypeUnknown && lastReturn.Kind != TypeAny {
				a.addError(&BangOutsideErrorFunctionError{
					Line: e.Line,
				})
			}
		}
		// If no function scope (top-level), bang is allowed (prints error and exits)

		// Bang unwraps (T, error) to T
		if innerType.Kind == TypeTuple && len(innerType.Elements) >= 2 {
			typ = innerType.Elements[0]
		} else if innerType.Kind == TypeError {
			typ = TypeUnknownVal // error-only return, no value
		} else {
			typ = innerType
		}

	case *ast.RescueExpr:
		a.analyzeExpr(e.Expr)
		if e.Default != nil {
			typ = a.analyzeExpr(e.Default)
		} else if e.Block != nil {
			// Create a new scope for the block and register error binding if present
			blockScope := NewScope(ScopeBlock, a.scope)
			if e.ErrName != "" {
				// Register the error binding variable with error type
				mustDefine(blockScope, NewSymbol(e.ErrName, SymVariable, TypeErrorVal))
			}
			// Analyze the block with the new scope
			prevScope := a.scope
			a.scope = blockScope
			for _, stmt := range e.Block.Body {
				a.analyzeStatement(stmt)
			}
			a.scope = prevScope
			typ = TypeUnknownVal
		} else {
			typ = TypeUnknownVal
		}

	case *ast.SpawnExpr:
		if e.Block != nil {
			a.analyzeBlock(e.Block)
			// Infer the Task element type from the block's return type
			blockReturnType := a.inferBlockReturnType(e.Block)
			typ = NewTaskType(blockReturnType)
		} else {
			typ = NewTaskType(TypeUnknownVal)
		}

	case *ast.AwaitExpr:
		taskType := a.analyzeExpr(e.Task)
		if taskType.Kind == TypeTask && taskType.Elem != nil && taskType.Elem.Kind != TypeUnknown {
			typ = taskType.Elem
		} else {
			// Await on Task<Unknown> returns any (interface{})
			typ = TypeAnyVal
		}

	case *ast.ConcurrentlyExpr:
		// Create block scope for the concurrently block
		blockScope := NewScope(ScopeBlock, a.scope)
		if e.ScopeVar != "" {
			mustDefine(blockScope, NewSymbol(e.ScopeVar, SymVariable, TypeUnknownVal))
		}
		prevScope := a.scope
		a.scope = blockScope
		for _, stmt := range e.Body {
			a.analyzeStatement(stmt)
		}
		a.scope = prevScope
		typ = TypeUnknownVal // Returns any

	case *ast.BlockExpr:
		a.analyzeBlock(e)
		typ = TypeUnknownVal

	case *ast.LambdaExpr:
		// Create a new scope for the lambda
		lambdaScope := NewScope(ScopeBlock, a.scope)
		prevScope := a.scope
		a.scope = lambdaScope

		// Add parameters to scope
		for _, param := range e.Params {
			paramType := TypeUnknownVal
			if param.Type != "" {
				paramType = ParseType(param.Type)
			}
			mustDefine(lambdaScope, NewSymbol(param.Name, SymParam, paramType))
		}

		// Analyze body and track the type of the last expression (implicit return)
		var lastExprType *Type
		for _, stmt := range e.Body {
			a.analyzeStatement(stmt)
			// Track type of last statement if it's an expression (implicit return)
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				lastExprType = a.GetType(exprStmt.Expr)
			} else if returnStmt, ok := stmt.(*ast.ReturnStmt); ok {
				// Explicit return - get type of first return value
				if len(returnStmt.Values) > 0 {
					lastExprType = a.GetType(returnStmt.Values[0])
				}
			}
		}

		a.scope = prevScope

		// Build function type: (ParamTypes) -> ReturnType
		var paramTypes []*Type
		for _, param := range e.Params {
			if param.Type != "" {
				paramTypes = append(paramTypes, ParseType(param.Type))
			} else {
				paramTypes = append(paramTypes, TypeAnyVal)
			}
		}
		returnType := TypeAnyVal
		if e.ReturnType != "" {
			// Explicit return type annotation takes precedence
			returnType = ParseType(e.ReturnType)
		} else if lastExprType != nil && lastExprType.Kind != TypeUnknown {
			// Infer return type from the last expression in the body
			returnType = lastExprType
		}
		typ = NewFuncType(paramTypes, []*Type{returnType})

	case *ast.InterpolatedString:
		for _, part := range e.Parts {
			if partExpr, ok := part.(ast.Expression); ok {
				a.analyzeExpr(partExpr)
			}
		}
		typ = TypeStringVal

	case *ast.SuperExpr:
		for _, arg := range e.Args {
			a.analyzeExpr(arg)
		}
		typ = TypeUnknownVal

	case *ast.SplatExpr:
		a.analyzeExpr(e.Expr)
		typ = TypeUnknownVal

	case *ast.DoubleSplatExpr:
		a.analyzeExpr(e.Expr)
		typ = TypeUnknownVal

	case *ast.SymbolToProcExpr:
		typ = TypeUnknownVal

	default:
		typ = TypeUnknownVal
	}

	if typ == nil {
		typ = TypeUnknownVal
	}

	a.setNodeType(expr, typ)
	return typ
}

// analyzeLambdaWithExpectedType analyzes a lambda expression using the expected
// function type to infer parameter and return types when not explicitly specified.
// This enables `handler : (Int) -> Int = -> (x) { x * 2 }` to work without
// explicit type annotations on the lambda.
func (a *Analyzer) analyzeLambdaWithExpectedType(e *ast.LambdaExpr, expectedType *Type) *Type {
	if expectedType == nil || expectedType.Kind != TypeFunc {
		// Fall back to regular analysis
		return a.analyzeExpr(e)
	}

	// Create a new scope for the lambda
	lambdaScope := NewScope(ScopeBlock, a.scope)
	prevScope := a.scope
	a.scope = lambdaScope

	// Add parameters to scope, using expected types for untyped params
	for i, param := range e.Params {
		var paramType *Type
		if param.Type != "" {
			// Explicit type takes precedence
			paramType = ParseType(param.Type)
		} else if i < len(expectedType.Params) {
			// Infer from expected function type
			paramType = expectedType.Params[i]
		} else {
			paramType = TypeAnyVal
		}
		mustDefine(lambdaScope, NewSymbol(param.Name, SymParam, paramType))
	}

	// Analyze body
	for _, stmt := range e.Body {
		a.analyzeStatement(stmt)
	}

	a.scope = prevScope

	// Build function type using inferred types
	var paramTypes []*Type
	for i, param := range e.Params {
		if param.Type != "" {
			paramTypes = append(paramTypes, ParseType(param.Type))
		} else if i < len(expectedType.Params) {
			paramTypes = append(paramTypes, expectedType.Params[i])
		} else {
			paramTypes = append(paramTypes, TypeAnyVal)
		}
	}

	// Use expected return type if lambda doesn't have explicit return type
	var returnType *Type
	if e.ReturnType != "" {
		returnType = ParseType(e.ReturnType)
	} else if len(expectedType.Returns) > 0 {
		returnType = expectedType.Returns[0]
	} else {
		returnType = TypeAnyVal
	}

	typ := NewFuncType(paramTypes, []*Type{returnType})
	a.setNodeType(e, typ)
	return typ
}

func (a *Analyzer) analyzeCall(call *ast.CallExpr) *Type {
	// First, try to resolve the function to get expected parameter types for lambdas
	var expectedParamTypes []*Type
	if ident, ok := call.Func.(*ast.Ident); ok {
		if fn := a.functions[ident.Name]; fn != nil {
			for _, param := range fn.Params {
				expectedParamTypes = append(expectedParamTypes, param.Type)
			}
		} else if a.scope.IsInsideClass() {
			className := a.scope.ClassName
			if cls := a.classes[className]; cls != nil {
				if method := cls.GetMethod(ident.Name); method != nil {
					for _, param := range method.Params {
						expectedParamTypes = append(expectedParamTypes, param.Type)
					}
				}
			}
		}
	}

	// Analyze arguments and collect types, check for splat
	// Use expected types for lambda arguments to enable type inference
	hasSplat := false
	var argTypes []*Type
	for i, arg := range call.Args {
		var argType *Type
		// If this is a lambda and we have an expected type that's a function,
		// use that to infer the lambda's parameter and return types
		if lambda, ok := arg.(*ast.LambdaExpr); ok && i < len(expectedParamTypes) {
			expectedType := expectedParamTypes[i]
			// Resolve type aliases to get the underlying function type
			if expectedType != nil {
				expectedType = a.resolveTypeAliasType(expectedType)
			}
			if expectedType != nil && expectedType.Kind == TypeFunc {
				argType = a.analyzeLambdaWithExpectedType(lambda, expectedType)
			} else {
				argType = a.analyzeExpr(arg)
			}
		} else {
			argType = a.analyzeExpr(arg)
		}
		argTypes = append(argTypes, argType)
		if _, ok := arg.(*ast.SplatExpr); ok {
			hasSplat = true
		}
	}

	// Handle method calls: receiver.method(args)
	if sel, ok := call.Func.(*ast.SelectorExpr); ok {
		receiverType := a.analyzeExpr(sel.X)

		// Analyze block with inferred parameter types
		if call.Block != nil {
			blockParamTypes := a.inferBlockParamTypes(receiverType, sel.Sel)
			// Special case: reduce with initial value - accumulator type is initial value's type
			if sel.Sel == "reduce" && len(argTypes) > 0 && argTypes[0] != nil && len(blockParamTypes) >= 2 {
				blockParamTypes = []*Type{argTypes[0], blockParamTypes[1]}
			}
			a.analyzeBlockWithTypes(call.Block, blockParamTypes)
		}

		// Check for class constructor call: ClassName.new(args)
		if sel.Sel == "new" {
			if classIdent, ok := sel.X.(*ast.Ident); ok {
				if cls, isClass := a.classes[classIdent.Name]; isClass {
					// Check initialize method arguments
					if initMethod := cls.GetMethod("initialize"); initMethod != nil && !hasSplat {
						a.checkArity("initialize", initMethod, call)
						a.checkArgumentTypes(classIdent.Name+".new", initMethod, argTypes)
					}
					return NewClassType(classIdent.Name)
				}
			}
			// Handle Module::Class.new(args) - scoped constructor call
			if scopeExpr, ok := sel.X.(*ast.ScopeExpr); ok {
				if modIdent, ok := scopeExpr.Left.(*ast.Ident); ok {
					if _, isModule := a.modules[modIdent.Name]; isModule {
						if cls, isClass := a.classes[scopeExpr.Right]; isClass {
							// Check initialize method arguments
							if initMethod := cls.GetMethod("initialize"); initMethod != nil && !hasSplat {
								a.checkArity("initialize", initMethod, call)
								a.checkArgumentTypes(modIdent.Name+"::"+scopeExpr.Right+".new", initMethod, argTypes)
							}
							return NewClassType(cls.Name)
						}
					}
				}
			}
			// Handle Chan<Type>.new(size) constructor
			if receiverType.Kind == TypeChan {
				return receiverType
			}
		}

		// Check for class method call: ClassName.method_name(args)
		if classIdent, ok := sel.X.(*ast.Ident); ok {
			if cls, isClass := a.classes[classIdent.Name]; isClass {
				if classMethod := cls.GetClassMethod(sel.Sel); classMethod != nil {
					// Check argument count and types
					if !hasSplat {
						a.checkArity(classIdent.Name+"."+sel.Sel, classMethod, call)
						a.checkArgumentTypes(classIdent.Name+"."+sel.Sel, classMethod, argTypes)
					}
					// Return class method's return type
					if len(classMethod.ReturnTypes) == 1 {
						return classMethod.ReturnTypes[0]
					} else if len(classMethod.ReturnTypes) > 1 {
						return NewTupleType(classMethod.ReturnTypes...)
					}
					return TypeUnknownVal
				}
			}
		}

		// Check for enum type method call: EnumName.from_string(str)
		if enumIdent, ok := sel.X.(*ast.Ident); ok {
			if _, isEnum := a.enums[enumIdent.Name]; isEnum {
				switch sel.Sel {
				case "from_string":
					// Enum.from_string returns EnumType? (optional)
					return NewOptionalType(&Type{Kind: TypeClass, Name: enumIdent.Name})
				case "values":
					// Enum.values returns Array<EnumType>
					return NewArrayType(&Type{Kind: TypeClass, Name: enumIdent.Name})
				}
			}
		}

		// Check for module method call: ModuleName.method_name(args)
		// Only class methods (def self.method) can be called on the module itself
		// Instance methods are for `include` into classes
		if modIdent, ok := sel.X.(*ast.Ident); ok {
			if mod, isModule := a.modules[modIdent.Name]; isModule {
				if modMethod := mod.Methods[sel.Sel]; modMethod != nil {
					// Verify this is a class method (def self.method)
					methodNode, isMethodDecl := modMethod.Node.(*ast.MethodDecl)
					if isMethodDecl && methodNode.IsClassMethod {
						// Check argument count and types
						if !hasSplat {
							a.checkArity(modIdent.Name+"."+sel.Sel, modMethod, call)
							a.checkArgumentTypes(modIdent.Name+"."+sel.Sel, modMethod, argTypes)
						}
						// Return module method's return type
						if len(modMethod.ReturnTypes) == 1 {
							return modMethod.ReturnTypes[0]
						} else if len(modMethod.ReturnTypes) > 1 {
							return NewTupleType(modMethod.ReturnTypes...)
						}
						return TypeUnknownVal
					}
					// Instance method called as module method - fall through to error handling
				}
			}
		}

		// Look up method on receiver type
		if method := a.lookupMethod(receiverType, sel.Sel); method != nil {
			// Check for private method access from outside the class
			if method.Private {
				// Private methods can only be called from within the same class
				canAccess := a.scope.IsInsideClass() && a.scope.ClassName == receiverType.Name
				if !canAccess {
					a.addError(&PrivateMethodAccessError{
						MethodName: sel.Sel,
						ClassName:  receiverType.Name,
						Line:       method.Line, // Use method's line since SelectorExpr doesn't have one
					})
				}
			}

			// Check argument count and types
			if !hasSplat {
				methodName := sel.Sel
				if receiverType.Kind == TypeClass && receiverType.Name != "" {
					methodName = receiverType.Name + "." + sel.Sel
				}
				a.checkArity(methodName, method, call)
				a.checkArgumentTypes(methodName, method, argTypes)
			}

			// Special handling for optional.map { } - return type is Optional[BlockReturnType]
			if receiverType.Kind == TypeOptional && sel.Sel == "map" && call.Block != nil {
				blockReturnType := a.inferBlockReturnType(call.Block)
				return NewOptionalType(blockReturnType)
			}

			// Special handling for array.map { } - return type is Array[BlockReturnType]
			if receiverType.Kind == TypeArray && sel.Sel == "map" && call.Block != nil {
				blockReturnType := a.inferBlockReturnType(call.Block)
				return NewArrayType(blockReturnType)
			}

			// Return method's return type
			if len(method.ReturnTypes) == 1 {
				return method.ReturnTypes[0]
			} else if len(method.ReturnTypes) > 1 {
				return NewTupleType(method.ReturnTypes...)
			}
			return TypeUnknownVal
		}

		// Fallback for unknown methods
		if strings.HasSuffix(sel.Sel, "?") {
			return TypeBoolVal
		}
		return TypeUnknownVal
	}

	// Handle safe navigation method calls: receiver&.method(args)
	if safeNav, ok := call.Func.(*ast.SafeNavExpr); ok {
		recvType := a.analyzeExpr(safeNav.Receiver)

		// Unwrap optional to get inner type
		innerType := recvType
		if recvType.Kind == TypeOptional && recvType.Elem != nil {
			innerType = recvType.Elem
		}

		// Analyze block with inferred parameter types
		if call.Block != nil {
			blockParamTypes := a.inferBlockParamTypes(innerType, safeNav.Selector)
			a.analyzeBlockWithTypes(call.Block, blockParamTypes)
		}

		// Look up method on inner type and check arguments
		var resultType *Type
		if method := a.lookupMethod(innerType, safeNav.Selector); method != nil {
			if !hasSplat {
				a.checkArity(safeNav.Selector, method, call)
				a.checkArgumentTypes(safeNav.Selector, method, argTypes)
			}

			if len(method.ReturnTypes) == 1 {
				resultType = method.ReturnTypes[0]
			} else if len(method.ReturnTypes) > 1 {
				resultType = NewTupleType(method.ReturnTypes...)
			}
		}

		// Default to unknown if we couldn't determine the type
		if resultType == nil {
			resultType = TypeUnknownVal
		}

		// Safe navigation returns optional (unless already optional)
		if resultType.Kind == TypeOptional {
			return resultType
		}
		return NewOptionalType(resultType)
	}

	// Handle direct function calls: func(args)
	// Note: Don't call analyzeExpr on Ident here - that would trigger the
	// "requires arguments" error even though args ARE being provided via CallExpr.
	// We handle Ident function lookups directly below.
	// For other expressions (e.g., function stored in variable), we still need
	// to analyze them to get type information.
	var funcType *Type
	if _, isIdent := call.Func.(*ast.Ident); !isIdent {
		funcType = a.analyzeExpr(call.Func)
	}

	// Analyze block if present (no type hints for direct function calls)
	if call.Block != nil {
		a.analyzeBlock(call.Block)
	}

	// Check argument count and types for known functions (skip if call uses splat)
	if !hasSplat {
		if ident, ok := call.Func.(*ast.Ident); ok {
			var fn *Symbol
			if fn = a.functions[ident.Name]; fn != nil {
				// Found as top-level function
			} else if a.scope.IsInsideClass() {
				// Check methods of the current class
				className := a.scope.ClassName
				if cls := a.classes[className]; cls != nil {
					fn = cls.GetMethod(ident.Name)
				}
			}
			if fn != nil {
				a.checkArity(ident.Name, fn, call)
				a.checkArgumentTypes(ident.Name, fn, argTypes)
			}
		}
	}

	// Try to determine return type from function type
	if funcType != nil && funcType.Kind == TypeFunc && len(funcType.Returns) > 0 {
		if len(funcType.Returns) == 1 {
			return funcType.Returns[0]
		}
		return NewTupleType(funcType.Returns...)
	}

	// Check for known function and return type
	if ident, ok := call.Func.(*ast.Ident); ok {
		var fn *Symbol
		if fn = a.functions[ident.Name]; fn != nil {
			// Found as top-level function
		} else if a.scope.IsInsideClass() {
			// Check methods of the current class
			className := a.scope.ClassName
			if cls := a.classes[className]; cls != nil {
				fn = cls.GetMethod(ident.Name)
			}
		}
		if fn != nil && len(fn.ReturnTypes) > 0 {
			if len(fn.ReturnTypes) == 1 {
				return fn.ReturnTypes[0]
			}
			return NewTupleType(fn.ReturnTypes...)
		}
	}

	return TypeUnknownVal
}

// checkArity verifies the argument count matches the function's parameter count.
func (a *Analyzer) checkArity(name string, fn *Symbol, call *ast.CallExpr) {
	maxParams := len(fn.Params)
	minParams := fn.RequiredParams // defaults to 0, which means all params required unless set
	if minParams == 0 {
		minParams = maxParams // if not set, all params are required
	}
	got := len(call.Args)

	if fn.Variadic {
		// Variadic functions accept any number of arguments >= minParams
		if got < minParams {
			a.addError(&ArityMismatchError{
				Name:     name,
				Expected: minParams,
				Got:      got,
			})
		}
		return
	}

	// Check if argument count is within valid range
	if got < minParams || got > maxParams {
		// Report expected as minParams if too few, maxParams if too many
		expected := minParams
		if got > maxParams {
			expected = maxParams
		}
		a.addError(&ArityMismatchError{
			Name:     name,
			Expected: expected,
			Got:      got,
		})
	}
}

// checkArgumentTypes verifies each argument type matches the corresponding parameter type.
func (a *Analyzer) checkArgumentTypes(name string, fn *Symbol, argTypes []*Type) {
	if fn.Variadic {
		// Skip type checking for variadic functions (like puts, print)
		return
	}

	for i, param := range fn.Params {
		if i >= len(argTypes) {
			break // Not enough arguments - checkArity will report this
		}
		argType := argTypes[i]
		paramType := param.Type

		// Skip if either type is unknown or any
		if argType.Kind == TypeUnknown || argType.Kind == TypeAny {
			continue
		}
		if paramType == nil || paramType.Kind == TypeUnknown || paramType.Kind == TypeAny {
			continue
		}

		if !a.isAssignable(paramType, argType) {
			a.addError(&ArgumentTypeMismatchError{
				FuncName:  name,
				ParamName: param.Name,
				Expected:  paramType,
				Got:       argType,
				ArgIndex:  i,
			})
		}
	}
}

// checkCondition validates that a condition expression evaluates to Bool.
func (a *Analyzer) checkCondition(cond ast.Expression, context string, line int) {
	condType := a.analyzeExpr(cond)

	// Skip if type is unknown or any
	if condType.Kind == TypeUnknown || condType.Kind == TypeAny {
		return
	}

	// Condition must be Bool
	if condType.Kind != TypeBool {
		a.addError(&ConditionTypeMismatchError{
			Got:     condType,
			Context: context,
			Line:    line,
		})
	}
}

// checkIndexType validates that an array/string index is Int.
func (a *Analyzer) checkIndexType(indexType *Type, line int) {
	// Skip if type is unknown or any
	if indexType.Kind == TypeUnknown || indexType.Kind == TypeAny {
		return
	}

	// Index must be Int (Int64 is also acceptable for indexing)
	if indexType.Kind != TypeInt && indexType.Kind != TypeInt64 {
		a.addError(&IndexTypeMismatchError{
			Got:  indexType,
			Line: line,
		})
	}
}

func (a *Analyzer) analyzeBlock(block *ast.BlockExpr) {
	a.analyzeBlockWithTypes(block, nil)
}

func (a *Analyzer) analyzeBlockWithTypes(block *ast.BlockExpr, paramTypes []*Type) {
	blockScope := NewScope(ScopeIterator, a.scope)

	// Add block parameters with inferred types
	for i, param := range block.Params {
		var paramType *Type
		if i < len(paramTypes) && paramTypes[i] != nil {
			paramType = paramTypes[i]
		} else {
			paramType = TypeAnyVal
		}
		mustDefine(blockScope, NewParam(param, paramType))
	}

	prevScope := a.scope
	a.scope = blockScope
	for _, stmt := range block.Body {
		a.analyzeStatement(stmt)
	}
	a.scope = prevScope
}

// inferBlockReturnType returns the type that a block expression produces.
// This is the type of the last expression in the block body.
func (a *Analyzer) inferBlockReturnType(block *ast.BlockExpr) *Type {
	if block == nil || len(block.Body) == 0 {
		return TypeAnyVal
	}

	// The return value is the last expression in the block
	lastStmt := block.Body[len(block.Body)-1]
	return a.inferStmtReturnType(lastStmt)
}

// inferStmtReturnType infers the return type of a statement when used as an expression.
// This handles control flow statements that may return values from branches.
func (a *Analyzer) inferStmtReturnType(stmt ast.Statement) *Type {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		if typ := a.getNodeType(s.Expr); typ != nil && typ.Kind != TypeUnknown {
			return typ
		}

	case *ast.IfStmt:
		// Get types from both branches and find common type
		var thenType, elseType *Type

		if len(s.Then) > 0 {
			thenType = a.inferStmtReturnType(s.Then[len(s.Then)-1])
		}

		if len(s.Else) > 0 {
			elseType = a.inferStmtReturnType(s.Else[len(s.Else)-1])
		} else if len(s.ElseIfs) > 0 {
			// Check last elsif's body
			lastElsif := s.ElseIfs[len(s.ElseIfs)-1]
			if len(lastElsif.Body) > 0 {
				elseType = a.inferStmtReturnType(lastElsif.Body[len(lastElsif.Body)-1])
			}
		}

		// If both branches have the same type, return it
		if thenType != nil && elseType != nil && thenType.Equals(elseType) {
			return thenType
		}
		// If only one branch has a type, return that (the other is likely nil/void)
		if thenType != nil && thenType.Kind != TypeAny {
			return thenType
		}
		if elseType != nil && elseType.Kind != TypeAny {
			return elseType
		}

	case *ast.ReturnStmt:
		// Explicit return - use its type
		if len(s.Values) == 1 {
			if typ := a.getNodeType(s.Values[0]); typ != nil {
				return typ
			}
		}
	}

	return TypeAnyVal
}

func (a *Analyzer) checkBinaryExpr(op string, left, right *Type) *Type {
	// Skip checking if either type is unknown/any (can't validate)
	if left.Kind == TypeUnknown || right.Kind == TypeUnknown {
		return a.inferBinaryType(op, left, right)
	}
	if left.Kind == TypeAny || right.Kind == TypeAny {
		return a.inferBinaryType(op, left, right)
	}

	switch op {
	case "==", "!=":
		// Equality: types should be comparable (same type or compatible)
		if !a.areTypesComparable(left, right) {
			a.addError(&OperatorTypeMismatchError{Op: op, LeftType: left, RightType: right})
		}
		return TypeBoolVal

	case "<", ">", "<=", ">=":
		// Ordering: only numeric and string types
		if !a.isOrdered(left) || !a.isOrdered(right) {
			a.addError(&OperatorTypeMismatchError{Op: op, LeftType: left, RightType: right})
		} else if !a.areTypesComparable(left, right) {
			a.addError(&OperatorTypeMismatchError{Op: op, LeftType: left, RightType: right})
		}
		return TypeBoolVal

	case "and", "or", "&&", "||":
		// Logical: both should be Bool
		if left.Kind != TypeBool {
			a.addError(&OperatorTypeMismatchError{Op: op, LeftType: left, RightType: right})
		} else if right.Kind != TypeBool {
			a.addError(&OperatorTypeMismatchError{Op: op, LeftType: left, RightType: right})
		}
		return TypeBoolVal

	case "+":
		// String concatenation or numeric addition
		if left.Kind == TypeString && right.Kind == TypeString {
			return TypeStringVal
		}
		if a.isNumeric(left) && a.isNumeric(right) {
			return a.inferBinaryType(op, left, right)
		}
		a.addError(&OperatorTypeMismatchError{Op: op, LeftType: left, RightType: right})
		return TypeUnknownVal

	case "-":
		// Arithmetic subtraction or set difference
		if left.Kind == TypeSet && right.Kind == TypeSet {
			return left // Set difference returns same set type
		}
		if !a.isNumeric(left) || !a.isNumeric(right) {
			a.addError(&OperatorTypeMismatchError{Op: op, LeftType: left, RightType: right})
			return TypeUnknownVal
		}
		return a.inferBinaryType(op, left, right)

	case "*", "/", "%":
		// Arithmetic: both must be numeric
		if !a.isNumeric(left) || !a.isNumeric(right) {
			a.addError(&OperatorTypeMismatchError{Op: op, LeftType: left, RightType: right})
			return TypeUnknownVal
		}
		return a.inferBinaryType(op, left, right)

	case "|":
		// Set union
		if left.Kind == TypeSet && right.Kind == TypeSet {
			return left // Union returns same set type
		}
		a.addError(&OperatorTypeMismatchError{Op: op, LeftType: left, RightType: right})
		return TypeUnknownVal

	case "&":
		// Set intersection
		if left.Kind == TypeSet && right.Kind == TypeSet {
			return left // Intersection returns same set type
		}
		a.addError(&OperatorTypeMismatchError{Op: op, LeftType: left, RightType: right})
		return TypeUnknownVal

	default:
		return TypeUnknownVal
	}
}

func (a *Analyzer) inferBinaryType(op string, left, right *Type) *Type {
	switch op {
	case "==", "!=", "<", ">", "<=", ">=":
		return TypeBoolVal
	case "and", "or", "&&", "||":
		return TypeBoolVal
	case "+":
		if left.Kind == TypeString || right.Kind == TypeString {
			return TypeStringVal
		}
		if left.Kind == TypeFloat || right.Kind == TypeFloat {
			return TypeFloatVal
		}
		if left.Kind == TypeInt {
			return TypeIntVal
		}
		// When both operands are any, result is any (e.g., await results)
		if left.Kind == TypeAny && right.Kind == TypeAny {
			return TypeAnyVal
		}
		return TypeUnknownVal
	case "-", "*", "/", "%":
		if left.Kind == TypeFloat || right.Kind == TypeFloat {
			return TypeFloatVal
		}
		if left.Kind == TypeInt {
			return TypeIntVal
		}
		// When both operands are any, result is any
		if left.Kind == TypeAny && right.Kind == TypeAny {
			return TypeAnyVal
		}
		return TypeUnknownVal
	default:
		return TypeUnknownVal
	}
}

// isNumeric returns true if the type is a numeric type (Int, Int64, Float).
// Also returns true for type parameters with the Numeric or Ordered constraint
// (Ordered implies Numeric for our purposes).
func (a *Analyzer) isNumeric(t *Type) bool {
	if t.Kind == TypeInt || t.Kind == TypeInt64 || t.Kind == TypeFloat {
		return true
	}
	// Type parameters with Numeric or Ordered constraint are considered numeric
	if t.Kind == TypeTypeParam {
		return t.Constraint == "Numeric" || t.Constraint == "Ordered"
	}
	return false
}

// isOrdered returns true if the type supports ordering comparison (<, >, etc.).
// Also returns true for type parameters with the Ordered or Numeric constraint.
func (a *Analyzer) isOrdered(t *Type) bool {
	if t.Kind == TypeInt || t.Kind == TypeInt64 || t.Kind == TypeFloat || t.Kind == TypeString {
		return true
	}
	// Type parameters with Ordered or Numeric constraint support ordering
	// (Numeric types are all ordered)
	if t.Kind == TypeTypeParam {
		return t.Constraint == "Ordered" || t.Constraint == "Numeric"
	}
	return false
}

// areTypesComparable returns true if two types can be compared for equality.
func (a *Analyzer) areTypesComparable(left, right *Type) bool {
	// Same type is always comparable
	if left.Equals(right) {
		return true
	}
	// Numeric types are comparable to each other
	if a.isNumeric(left) && a.isNumeric(right) {
		return true
	}
	// Type parameters are comparable if they have Equatable/Comparable/Ordered constraints
	// Note: Same type parameters (same name + constraint) are already handled by Equals above
	if left.Kind == TypeTypeParam && right.Kind == TypeTypeParam {
		leftComp := left.Constraint == "Equatable" || left.Constraint == "Comparable" || left.Constraint == "Ordered"
		rightComp := right.Constraint == "Equatable" || right.Constraint == "Comparable" || right.Constraint == "Ordered"
		return leftComp && rightComp
	}
	// nil can be compared to optional and error types
	if left.Kind == TypeNil {
		return right.Kind == TypeOptional || right.Kind == TypeError || right.Kind == TypeNil
	}
	if right.Kind == TypeNil {
		return left.Kind == TypeOptional || left.Kind == TypeError || left.Kind == TypeNil
	}
	// Optional can be compared to its underlying type (e.g., String? == String)
	if left.Kind == TypeOptional && left.Elem != nil && left.Elem.Equals(right) {
		return true
	}
	if right.Kind == TypeOptional && right.Elem != nil && right.Elem.Equals(left) {
		return true
	}
	return false
}

func (a *Analyzer) inferUnaryType(op string, operand *Type) *Type {
	switch op {
	case "not", "!":
		return TypeBoolVal
	case "-":
		return operand
	default:
		return TypeUnknownVal
	}
}

// checkUnaryExpr validates the operand type for a unary expression and returns the result type.
func (a *Analyzer) checkUnaryExpr(op string, operand *Type) *Type {
	// Skip checking if operand type is unknown/any (can't validate)
	if operand.Kind == TypeUnknown || operand.Kind == TypeAny {
		return a.inferUnaryType(op, operand)
	}

	switch op {
	case "not", "!":
		// Logical not: operand must be Bool
		if operand.Kind != TypeBool {
			a.addError(&UnaryTypeMismatchError{Op: op, OperandType: operand, Expected: "Bool"})
		}
		return TypeBoolVal

	case "-":
		// Unary minus: operand must be numeric
		if !a.isNumeric(operand) {
			a.addError(&UnaryTypeMismatchError{Op: op, OperandType: operand, Expected: "numeric"})
			return TypeUnknownVal
		}
		return operand

	default:
		return TypeUnknownVal
	}
}

// isAssignable checks if a value of type 'from' can be assigned to a variable of type 'to'.
func (a *Analyzer) isAssignable(to, from *Type) bool {
	if to == nil || from == nil {
		return true
	}

	// Resolve type aliases: if either type is a TypeClass that's actually a type alias,
	// resolve it to the underlying type before comparison
	to = a.resolveTypeAliasType(to)
	from = a.resolveTypeAliasType(from)

	if to.Kind == TypeAny || from.Kind == TypeAny {
		return true
	}

	// Type parameters accept any type (generic type checking)
	// TODO: Implement proper constraint checking - verify that `from` satisfies
	// any constraints on `to` when it's a TypeTypeParam (e.g., T : Ordered)
	if to.Kind == TypeTypeParam || from.Kind == TypeTypeParam {
		return true
	}
	if to.Kind == TypeUnknown || from.Kind == TypeUnknown {
		return true
	}
	if from.Kind == TypeNil {
		// nil can be assigned to optional types and error
		return to.Kind == TypeOptional || to.Kind == TypeError
	}

	// Optional types: concrete type T can be assigned to T?
	// This enables returning String from a String? function, passing Address to Address? param, etc.
	if to.Kind == TypeOptional && to.Elem != nil {
		// If the source type matches the optional's inner type, it's assignable
		if a.isAssignable(to.Elem, from) {
			return true
		}
	}

	// Interface structural typing: a class can be assigned to an interface
	// if it has all the required methods (Go-style structural typing)
	if to.Kind == TypeInterface && to.Name != "" && from.Kind == TypeClass && from.Name != "" {
		if a.classImplementsInterface(from.Name, to.Name) {
			return true
		}
	}

	// Error interface: a class with def error -> String implements error
	if to.Kind == TypeError && from.Kind == TypeClass && from.Name != "" {
		if a.classImplementsError(from.Name) {
			return true
		}
	}

	// Numeric widening: Int can be assigned to Int64 or Float
	if from.Kind == TypeInt {
		if to.Kind == TypeInt64 || to.Kind == TypeFloat {
			return true
		}
	}
	// Int64 can be assigned to Float
	if from.Kind == TypeInt64 && to.Kind == TypeFloat {
		return true
	}

	// Empty array (Array<Any>) is assignable to any typed array
	// This handles cases like: nums : Array<Int> = []
	if to.Kind == TypeArray && from.Kind == TypeArray {
		if from.Elem != nil && from.Elem.Kind == TypeAny {
			return true
		}
		// Note: Array<Int> is NOT assignable to Array<Any> because Go arrays
		// are invariant ([]int cannot be assigned to []any without copying)
	}

	// Empty map (Map[any, any]) is assignable to any typed map
	// This handles cases like: data : Map[String, Int] = {}
	if to.Kind == TypeMap && from.Kind == TypeMap {
		if from.KeyType != nil && from.KeyType.Kind == TypeAny &&
			from.ValueType != nil && from.ValueType.Kind == TypeAny {
			return true
		}
	}

	return to.Equals(from)
}

// classImplementsInterface checks if a class structurally implements an interface.
// This enables Go-style duck typing where explicit 'implements' is not required.
func (a *Analyzer) classImplementsInterface(className, interfaceName string) bool {
	iface := a.interfaces[interfaceName]
	if iface == nil {
		return false
	}

	// Check each interface method is implemented by the class (including inherited methods)
	for methodName, ifaceMethod := range iface.Methods {
		clsMethod := a.getClassMethod(className, methodName)
		if clsMethod == nil {
			return false
		}

		// Check parameter count matches
		if len(clsMethod.Params) != len(ifaceMethod.Params) {
			return false
		}

		// Check parameter types match
		for i, ifaceParam := range ifaceMethod.Params {
			clsParam := clsMethod.Params[i]
			if !a.typesMatch(ifaceParam.Type, clsParam.Type) {
				return false
			}
		}

		// Check return types match
		if len(clsMethod.ReturnTypes) != len(ifaceMethod.ReturnTypes) {
			return false
		}
		for i, ifaceRet := range ifaceMethod.ReturnTypes {
			clsRet := clsMethod.ReturnTypes[i]
			if !a.typesMatch(ifaceRet, clsRet) {
				return false
			}
		}
	}

	return true
}

// classImplementsError checks if a class implements the Go error interface.
// A class implements error if it has a method: def error -> String or def message -> String
func (a *Analyzer) classImplementsError(className string) bool {
	// Check for 'error' method first, then 'message' as fallback
	for _, methodName := range []string{"error", "message"} {
		method := a.getClassMethod(className, methodName)
		if method == nil {
			continue
		}

		// Must have no parameters
		if len(method.Params) != 0 {
			continue
		}

		// Must return exactly one String
		if len(method.ReturnTypes) != 1 {
			continue
		}

		if method.ReturnTypes[0].Kind == TypeString {
			return true
		}
	}
	return false
}

// findSimilar finds similar names for "did you mean?" suggestions.
func (a *Analyzer) findSimilar(name string) []string {
	var candidates []string

	// Check all symbols in scope chain
	scope := a.scope
	for scope != nil {
		for _, sym := range scope.Symbols() {
			if isSimilar(name, sym.Name) {
				candidates = append(candidates, sym.Name)
			}
		}
		scope = scope.Parent
	}

	// Check built-in functions
	for n := range a.functions {
		if isSimilar(name, n) {
			candidates = append(candidates, n)
		}
	}

	return candidates
}

// isSimilar checks if two names are similar (for suggestions).
// Uses Levenshtein distance with a threshold based on name length.
func isSimilar(a, b string) bool {
	if a == b {
		return false
	}

	// Calculate Levenshtein distance
	dist := levenshtein(a, b)

	// Threshold: allow up to ~30% of the longer name to be different
	// But at least 1 edit for short names, at most 3 for long names
	maxLen := max(len(a), len(b))
	threshold := min(max(maxLen/3, 1), 3)

	return dist <= threshold
}

// levenshtein calculates the Levenshtein distance between two strings.
func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create matrix
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	// Fill in the matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

// Helper methods

func (a *Analyzer) pushScope() {
	a.scope = NewScope(ScopeBlock, a.scope)
}

func (a *Analyzer) pushLoopScope() {
	a.scope = NewScope(ScopeLoop, a.scope)
}

func (a *Analyzer) popScope() {
	if a.scope.Parent != nil {
		a.scope = a.scope.Parent
	}
}

// mustDefine adds a symbol to a scope, ignoring redefinition errors.
// This is safe to use when defining in fresh scopes where duplicates are impossible.
func mustDefine(s *Scope, sym *Symbol) {
	_ = s.Define(sym)
}

func (a *Analyzer) addError(err error) {
	a.errors = append(a.errors, err)
}

func (a *Analyzer) setNodeType(node ast.Node, typ *Type) {
	a.nodeTypes[node] = typ
}

func (a *Analyzer) getNodeType(node ast.Node) *Type {
	if t, ok := a.nodeTypes[node]; ok {
		return t
	}
	return TypeUnknownVal
}

// Errors returns all errors collected during analysis.
func (a *Analyzer) Errors() []error {
	return a.errors
}

// GetType returns the inferred type for an AST node.
func (a *Analyzer) GetType(node ast.Node) *Type {
	return a.getNodeType(node)
}

// GetSelectorKind returns the resolved selector kind for a SelectorExpr.
// Returns SelectorUnknown if the node is not a SelectorExpr or was not resolved.
func (a *Analyzer) GetSelectorKind(node ast.Node) ast.SelectorKind {
	if sel, ok := node.(*ast.SelectorExpr); ok {
		// First check the node's own resolved kind
		if sel.ResolvedKind != ast.SelectorUnknown {
			return sel.ResolvedKind
		}
		// Fall back to the map
		if kind, ok := a.selectorKinds[sel]; ok {
			return kind
		}
	}
	return ast.SelectorUnknown
}

// setSelectorKind stores the selector kind for a SelectorExpr.
func (a *Analyzer) setSelectorKind(sel *ast.SelectorExpr, kind ast.SelectorKind) {
	a.selectorKinds[sel] = kind
	sel.ResolvedKind = kind
}

// IsVariableUsedAt checks if a variable declared at a specific AST node is actually used.
// This is used to detect unused variables in multi-value assignments.
func (a *Analyzer) IsVariableUsedAt(node ast.Node, name string) bool {
	if name == "_" {
		return true // blank identifier is always "used"
	}
	declMap, ok := a.declaredAt[node]
	if !ok {
		return true // if not tracked, assume used to be safe
	}
	sym, ok := declMap[name]
	if !ok {
		return true // variable was not declared here, assume used
	}
	return a.usedSymbols[sym]
}

// IsDeclaration returns true if this AST node declares a new variable.
// This is used by codegen to determine whether to use := (declaration) or = (assignment).
func (a *Analyzer) IsDeclaration(node ast.Node) bool {
	// Check the declarations map for assignment statements
	if isDecl, ok := a.declarations[node]; ok {
		return isDecl
	}

	// ForStmt loop variables are always new declarations
	if _, ok := node.(*ast.ForStmt); ok {
		return true
	}

	// Default to true (declaration) for safety - if unknown, use :=
	// This is safer because Go will error on redeclaration, which is more obvious
	// than silently reassigning when we wanted to declare
	return true
}

// GetFieldType returns the Rugby type of a class field by class and field name.
// Returns empty string if field not found.
func (a *Analyzer) GetFieldType(className, fieldName string) string {
	classSym, ok := a.classes[className]
	if !ok {
		return ""
	}
	if classSym.Fields == nil {
		return ""
	}
	if fieldSym, ok := classSym.Fields[fieldName]; ok && fieldSym.Type != nil {
		return fieldSym.Type.String()
	}
	return ""
}

// IsClass returns true if the given type name is a declared class.
func (a *Analyzer) IsClass(typeName string) bool {
	_, ok := a.classes[typeName]
	return ok
}

// IsStruct returns true if the given type name is a declared struct.
func (a *Analyzer) IsStruct(typeName string) bool {
	return a.structs[typeName]
}

// IsInterface returns true if the given type name is a declared interface.
func (a *Analyzer) IsInterface(typeName string) bool {
	_, ok := a.interfaces[typeName]
	return ok
}

// IsNoArgFunction returns true if the given name is a declared function with no parameters.
func (a *Analyzer) IsNoArgFunction(name string) bool {
	fn, ok := a.functions[name]
	if !ok {
		return false
	}
	return len(fn.Params) == 0
}

// IsPublicClass returns true if the given class name is declared as public.
func (a *Analyzer) IsPublicClass(className string) bool {
	cls, ok := a.classes[className]
	if !ok {
		return false
	}
	return cls.Public
}

// HasAccessor returns true if the given class field has a getter or setter accessor.
func (a *Analyzer) HasAccessor(className, fieldName string) bool {
	cls, ok := a.classes[className]
	if !ok {
		return false
	}
	if cls.Fields == nil {
		return false
	}
	field, ok := cls.Fields[fieldName]
	if !ok {
		return false
	}
	return field.HasGetter || field.HasSetter
}

// GetInterfaceMethodNames returns the method names declared in an interface.
func (a *Analyzer) GetInterfaceMethodNames(interfaceName string) []string {
	iface, ok := a.interfaces[interfaceName]
	if !ok {
		return nil
	}
	if iface.Methods == nil {
		return nil
	}
	names := make([]string, 0, len(iface.Methods))
	for name := range iface.Methods {
		names = append(names, name)
	}
	return names
}

// GetAllInterfaceNames returns the names of all declared interfaces.
func (a *Analyzer) GetAllInterfaceNames() []string {
	names := make([]string, 0, len(a.interfaces))
	for name := range a.interfaces {
		names = append(names, name)
	}
	return names
}

// GetAllModuleNames returns the names of all declared modules.
func (a *Analyzer) GetAllModuleNames() []string {
	names := make([]string, 0, len(a.modules))
	for name := range a.modules {
		names = append(names, name)
	}
	return names
}

// GetModuleMethodNames returns the method names declared in a module.
func (a *Analyzer) GetModuleMethodNames(moduleName string) []string {
	mod, ok := a.modules[moduleName]
	if !ok {
		return nil
	}
	if mod.Methods == nil {
		return nil
	}
	names := make([]string, 0, len(mod.Methods))
	for name := range mod.Methods {
		names = append(names, name)
	}
	return names
}

// IsModuleMethod returns true if the method in the given class came from an included module.
func (a *Analyzer) IsModuleMethod(className, methodName string) bool {
	cls, ok := a.classes[className]
	if !ok {
		return false
	}
	if cls.Methods == nil {
		return false
	}
	method, ok := cls.Methods[methodName]
	if !ok {
		return false
	}
	return method.FromModule != ""
}

// GetConstructorParamCount returns the number of constructor parameters for a class.
func (a *Analyzer) GetConstructorParamCount(className string) int {
	cls, ok := a.classes[className]
	if !ok {
		return 0
	}
	if cls.Methods == nil {
		return 0
	}
	init, ok := cls.Methods["initialize"]
	if !ok {
		return 0
	}
	return len(init.Params)
}

// GetConstructorParams returns the constructor parameter names and types for a class.
func (a *Analyzer) GetConstructorParams(className string) [][2]string {
	cls, ok := a.classes[className]
	if !ok {
		return nil
	}
	if cls.Methods == nil {
		return nil
	}
	init, ok := cls.Methods["initialize"]
	if !ok {
		return nil
	}
	result := make([][2]string, len(init.Params))
	for i, param := range init.Params {
		typeName := ""
		if param.Type != nil {
			typeName = param.Type.String()
		}
		result[i] = [2]string{param.Name, typeName}
	}
	return result
}

// GetSymbol looks up a symbol by name in the global scope.
func (a *Analyzer) GetSymbol(name string) *Symbol {
	return a.globalScope.Lookup(name)
}

// GetClass returns the class symbol by name.
func (a *Analyzer) GetClass(name string) *Symbol {
	return a.classes[name]
}

// GetInterface returns the interface symbol by name.
func (a *Analyzer) GetInterface(name string) *Symbol {
	return a.interfaces[name]
}

// GetFunction returns the function symbol by name.
func (a *Analyzer) GetFunction(name string) *Symbol {
	return a.functions[name]
}

// lookupMethod finds a method on the given type.
// Returns nil if not found.
func (a *Analyzer) lookupMethod(receiverType *Type, methodName string) *Symbol {
	if receiverType == nil {
		return nil
	}

	// Check for class methods, including inherited methods
	if receiverType.Kind == TypeClass && receiverType.Name != "" {
		// Walk up the class hierarchy looking for the method
		className := receiverType.Name
		for className != "" {
			cls := a.classes[className]
			if cls == nil {
				break
			}
			if method := cls.GetMethod(methodName); method != nil {
				return method
			}
			className = cls.Parent
		}
	}

	// Check for interface methods
	if receiverType.Kind == TypeInterface && receiverType.Name != "" {
		if iface := a.interfaces[receiverType.Name]; iface != nil {
			if method := iface.GetMethod(methodName); method != nil {
				return method
			}
		}
	}

	// Check for built-in methods on primitive types
	return a.lookupBuiltinMethod(receiverType, methodName)
}

// inferBlockParamTypes returns the expected types for block parameters
// based on the receiver type and method name.
func (a *Analyzer) inferBlockParamTypes(receiverType *Type, methodName string) []*Type {
	if receiverType == nil {
		return nil
	}

	switch receiverType.Kind {
	case TypeArray:
		elemType := receiverType.Elem
		if elemType == nil {
			elemType = TypeAnyVal
		}
		switch methodName {
		case "each", "map", "select", "filter", "reject", "find", "detect", "any?", "all?", "none?":
			return []*Type{elemType}
		case "each_with_index":
			return []*Type{elemType, TypeIntVal}
		case "reduce":
			// reduce takes |accumulator, element|, accumulator type is unknown
			return []*Type{TypeAnyVal, elemType}
		}

	case TypeMap:
		keyType := receiverType.KeyType
		valueType := receiverType.ValueType
		if keyType == nil {
			keyType = TypeAnyVal
		}
		if valueType == nil {
			valueType = TypeAnyVal
		}
		switch methodName {
		case "each", "select", "reject":
			return []*Type{keyType, valueType}
		case "each_key":
			return []*Type{keyType}
		case "each_value":
			return []*Type{valueType}
		}

	case TypeRange:
		switch methodName {
		case "each":
			return []*Type{TypeIntVal}
		}

	case TypeInt, TypeInt64:
		switch methodName {
		case "times", "upto", "downto":
			return []*Type{TypeIntVal}
		}

	case TypeOptional:
		innerType := receiverType.Elem
		if innerType == nil {
			innerType = TypeAnyVal
		}
		switch methodName {
		case "each", "map":
			return []*Type{innerType}
		}
	}

	return nil
}

// lookupBuiltinMethod returns method info for built-in type methods.
func (a *Analyzer) lookupBuiltinMethod(receiverType *Type, methodName string) *Symbol {
	switch receiverType.Kind {
	case TypeString:
		return a.stringMethod(methodName)
	case TypeInt, TypeInt64:
		return a.intMethod(methodName)
	case TypeFloat:
		return a.floatMethod(methodName)
	case TypeArray:
		return a.arrayMethod(methodName, receiverType.Elem)
	case TypeMap:
		return a.mapMethod(methodName, receiverType.KeyType, receiverType.ValueType)
	case TypeRange:
		return a.rangeMethod(methodName)
	case TypeChan:
		return a.chanMethod(methodName, receiverType.Elem)
	case TypeOptional:
		return a.optionalMethod(methodName, receiverType.Elem)
	}
	return nil
}

// String methods
func (a *Analyzer) stringMethod(name string) *Symbol {
	switch name {
	case "length", "size":
		return NewMethod(name, nil, []*Type{TypeIntVal})
	case "char_length":
		return NewMethod(name, nil, []*Type{TypeIntVal})
	case "empty?":
		return NewMethod(name, nil, []*Type{TypeBoolVal})
	case "upcase", "downcase", "strip", "lstrip", "rstrip", "reverse":
		return NewMethod(name, nil, []*Type{TypeStringVal})
	case "include?", "contains?", "start_with?", "end_with?":
		return NewMethod(name, []*Symbol{NewParam("s", TypeStringVal)}, []*Type{TypeBoolVal})
	case "replace":
		return NewMethod(name, []*Symbol{NewParam("old", TypeStringVal), NewParam("new", TypeStringVal)}, []*Type{TypeStringVal})
	case "split":
		return NewMethod(name, []*Symbol{NewParam("sep", TypeStringVal)}, []*Type{NewArrayType(TypeStringVal)})
	case "chars", "lines":
		return NewMethod(name, nil, []*Type{NewArrayType(TypeStringVal)})
	case "bytes":
		return NewMethod(name, nil, []*Type{TypeBytesVal})
	case "to_i":
		return NewMethod(name, nil, []*Type{TypeIntVal, TypeErrorVal})
	case "to_f":
		return NewMethod(name, nil, []*Type{TypeFloatVal, TypeErrorVal})
	}
	return nil
}

// Int methods
func (a *Analyzer) intMethod(name string) *Symbol {
	switch name {
	case "even?", "odd?", "zero?", "positive?", "negative?":
		return NewMethod(name, nil, []*Type{TypeBoolVal})
	case "abs":
		return NewMethod(name, nil, []*Type{TypeIntVal})
	case "clamp":
		return NewMethod(name, []*Symbol{NewParam("min", TypeIntVal), NewParam("max", TypeIntVal)}, []*Type{TypeIntVal})
	case "to_s":
		return NewMethod(name, nil, []*Type{TypeStringVal})
	case "to_f":
		return NewMethod(name, nil, []*Type{TypeFloatVal})
	case "times":
		// times takes an optional lambda argument
		method := NewMethod(name, nil, nil)
		method.Variadic = true // Allow 0 args (block) or 1 arg (lambda)
		return method
	case "upto", "downto":
		// upto/downto take 1 Int argument (the target) + optional lambda
		method := NewMethod(name, []*Symbol{NewParam("n", TypeIntVal)}, nil)
		method.Variadic = true // Allow 1 arg (block) or 2 args (target + lambda)
		return method
	}
	return nil
}

// Float methods
func (a *Analyzer) floatMethod(name string) *Symbol {
	switch name {
	case "floor", "ceil", "round", "truncate":
		return NewMethod(name, nil, []*Type{TypeFloatVal})
	case "zero?", "positive?", "negative?", "nan?", "infinite?":
		return NewMethod(name, nil, []*Type{TypeBoolVal})
	case "abs":
		return NewMethod(name, nil, []*Type{TypeFloatVal})
	case "to_i":
		return NewMethod(name, nil, []*Type{TypeIntVal})
	case "to_s":
		return NewMethod(name, nil, []*Type{TypeStringVal})
	}
	return nil
}

// Array methods
func (a *Analyzer) arrayMethod(name string, elemType *Type) *Symbol {
	if elemType == nil {
		elemType = TypeAnyVal
	}
	switch name {
	case "length", "size":
		return NewMethod(name, nil, []*Type{TypeIntVal})
	case "empty?":
		return NewMethod(name, nil, []*Type{TypeBoolVal})
	case "first", "last":
		return NewMethod(name, nil, []*Type{NewOptionalType(elemType)})
	case "include?", "contains?":
		return NewMethod(name, []*Symbol{NewParam("val", elemType)}, []*Type{TypeBoolVal})
	case "reverse", "sort":
		// In-place mutation, returns self
		return NewMethod(name, nil, nil)
	case "reversed", "sorted":
		// Return new array
		return NewMethod(name, nil, []*Type{NewArrayType(elemType)})
	case "compact":
		return NewMethod(name, nil, []*Type{NewArrayType(elemType)})
	case "sum":
		// Returns element type for numeric arrays
		return NewMethod(name, nil, []*Type{elemType})
	case "min", "max":
		return NewMethod(name, nil, []*Type{NewOptionalType(elemType)})
	case "each", "each_with_index":
		// These take a lambda argument for iteration
		method := NewMethod(name, nil, nil)
		method.Variadic = true // Allow 0 args (block) or 1 arg (lambda)
		return method
	case "map", "select", "filter", "reject", "find", "detect", "any?", "all?", "none?":
		// These can take an optional symbol-to-proc argument (&:method)
		method := NewMethod(name, nil, nil)
		method.Variadic = true // Allow 0 or 1 argument for symbol-to-proc
		switch name {
		case "map":
			method.ReturnTypes = []*Type{NewArrayType(TypeUnknownVal)}
		case "select", "filter", "reject":
			method.ReturnTypes = []*Type{NewArrayType(elemType)}
		case "find", "detect":
			method.ReturnTypes = []*Type{NewOptionalType(elemType)}
		case "any?", "all?", "none?":
			method.ReturnTypes = []*Type{TypeBoolVal}
		}
		return method
	case "reduce":
		// reduce takes 1 argument (initial value) or 2 arguments (initial value + lambda)
		method := NewMethod(name, []*Symbol{NewParam("initial", TypeAnyVal)}, []*Type{TypeUnknownVal})
		method.Variadic = true // Allow 1 or 2 args
		return method
	case "join":
		return NewMethod(name, []*Symbol{NewParam("sep", TypeStringVal)}, []*Type{TypeStringVal})
	case "push", "pop", "shift", "unshift":
		return NewMethod(name, nil, nil)
	}
	return nil
}

// Map methods
func (a *Analyzer) mapMethod(name string, keyType, valueType *Type) *Symbol {
	if keyType == nil {
		keyType = TypeAnyVal
	}
	if valueType == nil {
		valueType = TypeAnyVal
	}
	switch name {
	case "length", "size":
		return NewMethod(name, nil, []*Type{TypeIntVal})
	case "empty?":
		return NewMethod(name, nil, []*Type{TypeBoolVal})
	case "keys":
		return NewMethod(name, nil, []*Type{NewArrayType(keyType)})
	case "values":
		return NewMethod(name, nil, []*Type{NewArrayType(valueType)})
	case "has_key?", "key?":
		return NewMethod(name, []*Symbol{NewParam("key", keyType)}, []*Type{TypeBoolVal})
	case "fetch":
		return NewMethod(name, []*Symbol{NewParam("key", keyType), NewParam("default", valueType)}, []*Type{valueType})
	case "merge":
		return NewMethod(name, []*Symbol{NewParam("other", NewMapType(keyType, valueType))}, []*Type{NewMapType(keyType, valueType)})
	case "each", "each_key", "each_value", "select", "reject":
		switch name {
		case "each", "each_key", "each_value":
			method := NewMethod(name, nil, nil)
			method.Variadic = true // Allow 0 args (block) or 1 arg (lambda)
			return method
		case "select", "reject":
			method := NewMethod(name, nil, []*Type{NewMapType(keyType, valueType)})
			method.Variadic = true // Allow 0 args (block) or 1 arg (lambda)
			return method
		}
	case "delete":
		return NewMethod(name, []*Symbol{NewParam("key", keyType)}, []*Type{NewOptionalType(valueType)})
	}
	return nil
}

// Range methods
func (a *Analyzer) rangeMethod(name string) *Symbol {
	switch name {
	case "include?", "contains?":
		return NewMethod(name, []*Symbol{NewParam("val", TypeIntVal)}, []*Type{TypeBoolVal})
	case "to_a":
		return NewMethod(name, nil, []*Type{NewArrayType(TypeIntVal)})
	case "size", "length":
		return NewMethod(name, nil, []*Type{TypeIntVal})
	case "each":
		return NewMethod(name, nil, nil)
	}
	return nil
}

// Channel methods
func (a *Analyzer) chanMethod(name string, elemType *Type) *Symbol {
	if elemType == nil {
		elemType = TypeAnyVal
	}
	switch name {
	case "receive":
		return NewMethod(name, nil, []*Type{elemType})
	case "try_receive":
		return NewMethod(name, nil, []*Type{NewOptionalType(elemType)})
	case "close":
		return NewMethod(name, nil, nil)
	}
	return nil
}

// Optional methods
func (a *Analyzer) optionalMethod(name string, innerType *Type) *Symbol {
	if innerType == nil {
		innerType = TypeAnyVal
	}
	switch name {
	case "ok?", "present?":
		return NewMethod(name, nil, []*Type{TypeBoolVal})
	case "nil?", "absent?":
		return NewMethod(name, nil, []*Type{TypeBoolVal})
	case "unwrap":
		return NewMethod(name, nil, []*Type{innerType})
	case "map", "each":
		return NewMethod(name, nil, nil) // Block methods
	}
	return nil
}

// parseTypeFromExpr converts an AST expression to a Type.
// This is used for generic type parameters like Chan<Int> where Int is an identifier.
func (a *Analyzer) parseTypeFromExpr(expr ast.Expression) *Type {
	switch e := expr.(type) {
	case *ast.Ident:
		// Map type name identifiers to their types
		switch e.Name {
		case "Int":
			return TypeIntVal
		case "Int64":
			return TypeInt64Val
		case "Float":
			return TypeFloatVal
		case "Bool":
			return TypeBoolVal
		case "String":
			return TypeStringVal
		case "Bytes":
			return TypeBytesVal
		case "any":
			return TypeAnyVal
		case "Error", "error":
			return TypeErrorVal
		default:
			// Check if it's a struct type
			if a.structs[e.Name] {
				return &Type{Kind: TypeClass, Name: e.Name, IsStruct: true}
			}
			// Check if it's a class type
			if cls := a.classes[e.Name]; cls != nil {
				return NewClassType(e.Name)
			}
			// Check if it's an interface type
			if iface := a.interfaces[e.Name]; iface != nil {
				return NewInterfaceType(e.Name)
			}
			return TypeUnknownVal
		}
	case *ast.IndexExpr:
		// Handle nested generics like Array<Int>
		if ident, ok := e.Left.(*ast.Ident); ok {
			elemType := a.parseTypeFromExpr(e.Index)
			switch ident.Name {
			case "Array":
				return NewArrayType(elemType)
			case "Chan":
				return NewChanType(elemType)
			case "Task":
				return NewTaskType(elemType)
			}
		}
		return TypeUnknownVal
	default:
		return TypeUnknownVal
	}
}
