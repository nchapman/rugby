package semantic

import (
	"strings"

	"github.com/nchapman/rugby/ast"
)

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
		{"rand", []*Type{TypeIntVal}, []*Type{TypeIntVal}, false},
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

	// First pass: collect all top-level declarations (classes, interfaces, modules, functions)
	for _, decl := range program.Declarations {
		a.collectDeclaration(decl)
	}

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

			// Check each interface method is implemented
			for methodName, ifaceMethod := range iface.Methods {
				clsMethod := cls.GetMethod(methodName)
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

// resolveType checks if a type is a TypeClass that should be TypeInterface.
// If the type name matches a declared interface, it updates the type's Kind to TypeInterface.
// Also recursively resolves nested types (optional, array, map, etc.)
func (a *Analyzer) resolveType(t *Type) *Type {
	if t == nil {
		return t
	}

	// If this is a class type, check if it should be an interface
	if t.Kind == TypeClass && t.Name != "" {
		if _, isInterface := a.interfaces[t.Name]; isInterface {
			t.Kind = TypeInterface
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
	}
}

func (a *Analyzer) collectFunc(f *ast.FuncDecl) {
	params := make([]*Symbol, len(f.Params))
	for i, p := range f.Params {
		var typ *Type
		if p.Type != "" {
			typ = ParseType(p.Type)
		} else {
			typ = TypeUnknownVal
		}
		params[i] = NewParam(p.Name, typ)
	}

	returnTypes := make([]*Type, len(f.ReturnTypes))
	for i, rt := range f.ReturnTypes {
		returnTypes[i] = ParseType(rt)
	}

	fn := NewFunction(f.Name, params, returnTypes)
	fn.Line = f.Line
	fn.Public = f.Pub
	fn.Node = f

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

	// Collect fields from explicit declarations and accessors
	for _, f := range c.Fields {
		field := NewField(f.Name, ParseType(f.Type))
		cls.AddField(field)
	}
	for _, acc := range c.Accessors {
		field := NewField(acc.Name, ParseType(acc.Type))
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
		params := make([]*Symbol, len(m.Params))
		for i, p := range m.Params {
			var typ *Type
			if p.Type != "" {
				typ = ParseType(p.Type)
			} else {
				typ = TypeUnknownVal
			}
			params[i] = NewParam(p.Name, typ)
		}

		returnTypes := make([]*Type, len(m.ReturnTypes))
		for i, rt := range m.ReturnTypes {
			returnTypes[i] = ParseType(rt)
		}

		method := NewMethod(m.Name, params, returnTypes)
		method.Line = m.Line
		method.Public = m.Pub
		method.Node = m
		cls.AddMethod(method)
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
		for j, p := range m.Params {
			var typ *Type
			if p.Type != "" {
				typ = ParseType(p.Type)
			} else {
				typ = TypeUnknownVal
			}
			params[j] = NewParam(p.Name, typ)
		}

		returnTypes := make([]*Type, len(m.ReturnTypes))
		for j, rt := range m.ReturnTypes {
			returnTypes[j] = ParseType(rt)
		}

		method := NewMethod(m.Name, params, returnTypes)
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

	// Collect methods
	for _, method := range m.Methods {
		params := make([]*Symbol, len(method.Params))
		for i, p := range method.Params {
			var typ *Type
			if p.Type != "" {
				typ = ParseType(p.Type)
			} else {
				typ = TypeUnknownVal
			}
			params[i] = NewParam(p.Name, typ)
		}

		returnTypes := make([]*Type, len(method.ReturnTypes))
		for i, rt := range method.ReturnTypes {
			returnTypes[i] = ParseType(rt)
		}

		meth := NewMethod(method.Name, params, returnTypes)
		meth.Line = method.Line
		meth.Public = method.Pub
		meth.Node = method
		mod.AddMethod(meth)
	}

	if err := a.scope.Define(mod); err != nil {
		a.addError(err)
	}
	a.modules[m.Name] = mod
}

// analyzeStatement analyzes a statement and returns its type (if applicable).
func (a *Analyzer) analyzeStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.FuncDecl:
		a.analyzeFuncDecl(s)
	case *ast.ClassDecl:
		a.analyzeClassDecl(s)
	case *ast.AssignStmt:
		a.analyzeAssign(s)
	case *ast.MultiAssignStmt:
		a.analyzeMultiAssign(s)
	case *ast.CompoundAssignStmt:
		a.analyzeCompoundAssign(s)
	case *ast.OrAssignStmt:
		a.analyzeOrAssign(s)
	case *ast.SelectorAssignStmt:
		a.analyzeSelectorAssign(s)
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
	}
}

func (a *Analyzer) analyzeFuncDecl(f *ast.FuncDecl) {
	// Create function scope
	fnScope := NewScope(ScopeFunction, a.scope)
	fnScope.ReturnTypes = make([]*Type, len(f.ReturnTypes))
	for i, rt := range f.ReturnTypes {
		fnScope.ReturnTypes[i] = ParseType(rt)
	}

	// Add parameters to scope
	for _, p := range f.Params {
		var typ *Type
		if p.Type != "" {
			typ = ParseType(p.Type)
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
	for _, stmt := range f.Body {
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

	// Add methods from included modules to scope and to cls.Methods
	// These become callable as bare functions within class methods
	// AND as external method calls on class instances
	for _, modName := range cls.Includes {
		if mod := a.modules[modName]; mod != nil {
			for _, method := range mod.Methods {
				mustDefine(classScope, method)
				// Also add to cls.Methods so GetMethod can find them
				cls.AddMethod(method)
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
	valueType := a.analyzeExpr(s.Value)

	// Check if variable already exists
	existing := a.scope.Lookup(s.Name)
	if existing != nil {
		// Assignment to existing variable - check type compatibility
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
		var typ *Type
		if s.Type != "" {
			typ = ParseType(s.Type)
			// Check that value is assignable to declared type
			if valueType.Kind != TypeUnknown && !a.isAssignable(typ, valueType) {
				a.addError(&TypeMismatchError{
					Expected: typ,
					Got:      valueType,
					Line:     s.Line,
					Context:  "assignment",
				})
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
					_ = a.scope.DefineOrShadow(NewVariable(valName, innerType))
				}
			}
			if okName != "_" {
				existing := a.scope.LookupLocal(okName)
				if existing == nil {
					_ = a.scope.DefineOrShadow(NewVariable(okName, TypeBoolVal))
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
				_ = a.scope.DefineOrShadow(NewVariable(name, TypeAnyVal))
			}
		}
	}
}

func (a *Analyzer) analyzeCompoundAssign(s *ast.CompoundAssignStmt) {
	// Variable must exist
	existing := a.scope.Lookup(s.Name)
	if existing == nil {
		a.addError(&UndefinedError{Name: s.Name})
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
		v := NewVariable(s.Name, valueType)
		_ = a.scope.DefineOrShadow(v)
	}
}

func (a *Analyzer) analyzeSelectorAssign(s *ast.SelectorAssignStmt) {
	// Analyze the receiver object
	receiverType := a.analyzeExpr(s.Object)

	// Analyze the value being assigned
	valueType := a.analyzeExpr(s.Value)

	// Verify the receiver type has a setter for this field
	if receiverType.Kind == TypeClass && receiverType.Name != "" {
		if cls := a.classes[receiverType.Name]; cls != nil {
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
			} else {
				a.addError(&UndefinedError{
					Name: s.Field,
					Line: s.Line,
				})
			}
		}
	}
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

func (a *Analyzer) analyzeFor(s *ast.ForStmt) {
	iterableType := a.analyzeExpr(s.Iterable)

	// Determine loop variable type from iterable
	var elemType *Type
	switch iterableType.Kind {
	case TypeArray:
		elemType = iterableType.Elem
	case TypeRange:
		elemType = TypeIntVal
	case TypeChan:
		elemType = iterableType.Elem
	case TypeMap:
		// For maps, the loop variable is the key
		elemType = iterableType.KeyType
	case TypeString:
		elemType = TypeStringVal // iterating runes as strings
	default:
		elemType = TypeAnyVal
	}
	if elemType == nil {
		elemType = TypeAnyVal
	}

	// Create loop scope with loop variable
	loopScope := NewScope(ScopeLoop, a.scope)
	mustDefine(loopScope, NewVariable(s.Var, elemType))

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
	if !a.scope.IsInsideLoop() {
		a.addError(&BreakOutsideLoopError{})
	}

	if s.Condition != nil {
		a.analyzeExpr(s.Condition)
	}
}

func (a *Analyzer) analyzeNext(s *ast.NextStmt) {
	if !a.scope.IsInsideLoop() {
		a.addError(&Error{Message: "next statement outside loop"})
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

		// In each branch, the subject has the narrowed type
		// We use a special variable to track this
		if ident, ok := s.Subject.(*ast.Ident); ok {
			narrowedType := ParseType(when.Type)
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

	a.analyzeExpr(s.Value)
}

func (a *Analyzer) analyzeInstanceVarCompoundAssign(s *ast.InstanceVarCompoundAssign) {
	if !a.scope.IsInsideClass() {
		a.addError(&InstanceVarOutsideClassError{Name: s.Name})
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
		sym := a.scope.Lookup(e.Name)
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
		}

	case *ast.InstanceVar:
		if !a.scope.IsInsideClass() {
			a.addError(&InstanceVarOutsideClassError{Name: e.Name})
			typ = TypeUnknownVal
		} else {
			// Look up field type from class
			classScope := a.scope.ClassScope()
			if classScope != nil {
				if cls := a.classes[classScope.ClassName]; cls != nil {
					if field := cls.GetField(e.Name); field != nil {
						typ = field.Type
					}
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
					// Validate value type matches
					if !a.isAssignable(valueType, vt) && !a.isAssignable(vt, valueType) {
						a.addError(&TypeMismatchError{
							Expected: valueType,
							Got:      vt,
							Context:  "map value",
						})
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
				// If we couldn't resolve, assume it's a method for Go interop
				if isGoPackage {
					selectorKind = ast.SelectorGoMethod
				}
			}
		}

		// Store the resolved selector kind
		if selectorKind != ast.SelectorUnknown {
			a.setSelectorKind(e, selectorKind)
		}

	case *ast.IndexExpr:
		leftType := a.analyzeExpr(e.Left)
		indexType := a.analyzeExpr(e.Index)

		switch leftType.Kind {
		case TypeArray:
			if indexType.Kind == TypeRange {
				// Array slice with range - returns a slice of same type
				typ = leftType
			} else {
				// Array index must be Int
				a.checkIndexType(indexType)
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
				a.checkIndexType(indexType)
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
				Line:    0, // TernaryExpr doesn't have Line field
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
				Line: 0, // BangExpr doesn't have Line field
			})
		}

		// Validate that enclosing function returns error (or we're at top-level/main)
		fnScope := a.scope.FunctionScope()
		if fnScope != nil && len(fnScope.ReturnTypes) > 0 {
			lastReturn := fnScope.ReturnTypes[len(fnScope.ReturnTypes)-1]
			if lastReturn.Kind != TypeError && lastReturn.Kind != TypeUnknown && lastReturn.Kind != TypeAny {
				a.addError(&BangOutsideErrorFunctionError{
					Line: 0, // BangExpr doesn't have Line field
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
		}
		typ = NewTaskType(TypeUnknownVal) // TODO: infer from block return type

	case *ast.AwaitExpr:
		taskType := a.analyzeExpr(e.Task)
		if taskType.Kind == TypeTask {
			typ = taskType.Elem
		} else {
			typ = TypeUnknownVal
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

func (a *Analyzer) analyzeCall(call *ast.CallExpr) *Type {
	// Analyze arguments and collect types, check for splat
	hasSplat := false
	var argTypes []*Type
	for _, arg := range call.Args {
		argType := a.analyzeExpr(arg)
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
		}

		// Look up method on receiver type
		if method := a.lookupMethod(receiverType, sel.Sel); method != nil {
			// Check argument count and types
			if !hasSplat {
				methodName := sel.Sel
				if receiverType.Kind == TypeClass && receiverType.Name != "" {
					methodName = receiverType.Name + "." + sel.Sel
				}
				a.checkArity(methodName, method, call)
				a.checkArgumentTypes(methodName, method, argTypes)
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
			if fn := a.functions[ident.Name]; fn != nil {
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
		if fn := a.functions[ident.Name]; fn != nil {
			if len(fn.ReturnTypes) > 0 {
				if len(fn.ReturnTypes) == 1 {
					return fn.ReturnTypes[0]
				}
				return NewTupleType(fn.ReturnTypes...)
			}
		}
	}

	return TypeUnknownVal
}

// checkArity verifies the argument count matches the function's parameter count.
func (a *Analyzer) checkArity(name string, fn *Symbol, call *ast.CallExpr) {
	expected := len(fn.Params)
	got := len(call.Args)

	if fn.Variadic {
		// Variadic functions accept any number of arguments (including zero)
		// The params slice contains the base parameter type for documentation
		return
	}

	if got != expected {
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
func (a *Analyzer) checkIndexType(indexType *Type) {
	// Skip if type is unknown or any
	if indexType.Kind == TypeUnknown || indexType.Kind == TypeAny {
		return
	}

	// Index must be Int (Int64 is also acceptable for indexing)
	if indexType.Kind != TypeInt && indexType.Kind != TypeInt64 {
		a.addError(&IndexTypeMismatchError{
			Got:  indexType,
			Line: 0, // IndexExpr doesn't have Line field
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

	case "-", "*", "/", "%":
		// Arithmetic: both must be numeric
		if !a.isNumeric(left) || !a.isNumeric(right) {
			a.addError(&OperatorTypeMismatchError{Op: op, LeftType: left, RightType: right})
			return TypeUnknownVal
		}
		return a.inferBinaryType(op, left, right)

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
		return TypeUnknownVal
	case "-", "*", "/", "%":
		if left.Kind == TypeFloat || right.Kind == TypeFloat {
			return TypeFloatVal
		}
		if left.Kind == TypeInt {
			return TypeIntVal
		}
		return TypeUnknownVal
	default:
		return TypeUnknownVal
	}
}

// isNumeric returns true if the type is a numeric type (Int, Int64, Float).
func (a *Analyzer) isNumeric(t *Type) bool {
	return t.Kind == TypeInt || t.Kind == TypeInt64 || t.Kind == TypeFloat
}

// isOrdered returns true if the type supports ordering comparison (<, >, etc.).
func (a *Analyzer) isOrdered(t *Type) bool {
	return t.Kind == TypeInt || t.Kind == TypeInt64 || t.Kind == TypeFloat || t.Kind == TypeString
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
	// nil can be compared to optional and error types
	if left.Kind == TypeNil {
		return right.Kind == TypeOptional || right.Kind == TypeError || right.Kind == TypeNil
	}
	if right.Kind == TypeNil {
		return left.Kind == TypeOptional || left.Kind == TypeError || left.Kind == TypeNil
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
	if to.Kind == TypeAny || from.Kind == TypeAny {
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

	return to.Equals(from)
}

// classImplementsInterface checks if a class structurally implements an interface.
// This enables Go-style duck typing where explicit 'implements' is not required.
func (a *Analyzer) classImplementsInterface(className, interfaceName string) bool {
	cls := a.classes[className]
	iface := a.interfaces[interfaceName]
	if cls == nil || iface == nil {
		return false
	}

	// Check each interface method is implemented by the class
	for methodName, ifaceMethod := range iface.Methods {
		clsMethod := cls.GetMethod(methodName)
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
func isSimilar(a, b string) bool {
	if a == b {
		return false
	}
	// Simple heuristic: same length and differ by 1-2 characters
	if len(a) == len(b) {
		diff := 0
		for i := range a {
			if a[i] != b[i] {
				diff++
			}
		}
		return diff <= 2
	}
	// Or one is a prefix of the other
	if strings.HasPrefix(a, b) || strings.HasPrefix(b, a) {
		return true
	}
	return false
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

	// Check for class methods
	if receiverType.Kind == TypeClass && receiverType.Name != "" {
		if cls := a.classes[receiverType.Name]; cls != nil {
			if method := cls.GetMethod(methodName); method != nil {
				return method
			}
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
		// times takes no arguments, just a block
		return NewMethod(name, nil, nil)
	case "upto", "downto":
		// upto/downto take 1 Int argument (the target)
		return NewMethod(name, []*Symbol{NewParam("n", TypeIntVal)}, nil)
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
		return NewMethod(name, nil, nil)
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
		// reduce takes 1 argument (initial value)
		return NewMethod(name, []*Symbol{NewParam("initial", TypeAnyVal)}, []*Type{TypeUnknownVal})
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
			return NewMethod(name, nil, nil)
		case "select", "reject":
			return NewMethod(name, nil, []*Type{NewMapType(keyType, valueType)})
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
