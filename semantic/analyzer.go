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
		scope:       global,
		globalScope: global,
		nodeTypes:   make(map[ast.Node]*Type),
		classes:     make(map[string]*Symbol),
		interfaces:  make(map[string]*Symbol),
		modules:     make(map[string]*Symbol),
		functions:   make(map[string]*Symbol),
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
	// First pass: collect all top-level declarations (classes, interfaces, modules, functions)
	for _, decl := range program.Declarations {
		a.collectDeclaration(decl)
	}

	// Second pass: analyze bodies
	for _, decl := range program.Declarations {
		a.analyzeStatement(decl)
	}

	return a.errors
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
	}
}

func (a *Analyzer) analyzeCompoundAssign(s *ast.CompoundAssignStmt) {
	// Variable must exist
	existing := a.scope.Lookup(s.Name)
	if existing == nil {
		a.addError(&UndefinedError{Name: s.Name})
		return
	}

	a.analyzeExpr(s.Value)
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

func (a *Analyzer) analyzeIf(s *ast.IfStmt) {
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
		a.analyzeExpr(s.Cond)

		// Analyze then branch
		a.pushScope()
		for _, stmt := range s.Then {
			a.analyzeStatement(stmt)
		}
		a.popScope()
	}

	// Analyze elsif branches
	for _, elsif := range s.ElseIfs {
		a.analyzeExpr(elsif.Cond)
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
	a.analyzeExpr(s.Cond)

	a.pushLoopScope()
	for _, stmt := range s.Body {
		a.analyzeStatement(stmt)
	}
	a.popScope()
}

func (a *Analyzer) analyzeUntil(s *ast.UntilStmt) {
	a.analyzeExpr(s.Cond)

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

	// Analyze return values
	for _, val := range s.Values {
		a.analyzeExpr(val)
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
	if s.Subject != nil {
		a.analyzeExpr(s.Subject)
	}

	for _, when := range s.WhenClauses {
		for _, val := range when.Values {
			a.analyzeExpr(val)
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

func (a *Analyzer) analyzeGo(s *ast.GoStmt) {
	if s.Call != nil {
		a.analyzeExpr(s.Call)
	}
	if s.Block != nil {
		a.analyzeBlock(s.Block)
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
			a.addError(&UndefinedError{Name: e.Name, Candidates: a.findSimilar(e.Name)})
			typ = TypeUnknownVal
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
			et := a.analyzeExpr(elem)
			if elemType == nil {
				elemType = et
			}
			// TODO: unify types
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
			} else {
				if entry.Key != nil {
					kt := a.analyzeExpr(entry.Key)
					if keyType == nil {
						keyType = kt
					}
				}
				if entry.Value != nil {
					vt := a.analyzeExpr(entry.Value)
					if valueType == nil {
						valueType = vt
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
		a.analyzeExpr(e.X)
		// For now, return unknown for method calls
		// TODO: look up method return type
		if strings.HasSuffix(e.Sel, "?") {
			typ = TypeBoolVal
		} else {
			typ = TypeUnknownVal
		}

	case *ast.IndexExpr:
		leftType := a.analyzeExpr(e.Left)
		a.analyzeExpr(e.Index)

		switch leftType.Kind {
		case TypeArray:
			typ = leftType.Elem
		case TypeMap:
			// Map access returns optional
			typ = NewOptionalType(leftType.ValueType)
		case TypeString:
			typ = TypeStringVal
		default:
			typ = TypeUnknownVal
		}

	case *ast.TernaryExpr:
		a.analyzeExpr(e.Condition)
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
		// Result is the unwrapped left type or right type
		if leftType.Kind == TypeOptional {
			typ = leftType.Elem
		} else {
			typ = rightType
		}

	case *ast.SafeNavExpr:
		recvType := a.analyzeExpr(e.Receiver)
		// Safe navigation returns optional
		if recvType.Kind == TypeOptional {
			// Already optional, stays optional
			typ = recvType
		} else {
			typ = NewOptionalType(TypeUnknownVal)
		}

	case *ast.BangExpr:
		innerType := a.analyzeExpr(e.Expr)
		// Bang unwraps (T, error) to T
		if innerType.Kind == TypeTuple && len(innerType.Elements) >= 2 {
			typ = innerType.Elements[0]
		} else {
			typ = innerType
		}

	case *ast.RescueExpr:
		a.analyzeExpr(e.Expr)
		if e.Default != nil {
			typ = a.analyzeExpr(e.Default)
		} else if e.Block != nil {
			a.analyzeBlock(e.Block)
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
	// Analyze function/receiver
	funcType := a.analyzeExpr(call.Func)

	// Analyze arguments and check for splat (which makes arity checking unreliable)
	hasSplat := false
	for _, arg := range call.Args {
		a.analyzeExpr(arg)
		if _, ok := arg.(*ast.SplatExpr); ok {
			hasSplat = true
		}
	}

	// Analyze block if present
	if call.Block != nil {
		a.analyzeBlock(call.Block)
	}

	// Check argument count for known functions (skip if call uses splat)
	if !hasSplat {
		if ident, ok := call.Func.(*ast.Ident); ok {
			if fn := a.functions[ident.Name]; fn != nil {
				a.checkArity(ident.Name, fn, call)
			}
		}
	}

	// Try to determine return type
	if funcType.Kind == TypeFunc && len(funcType.Returns) > 0 {
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

	// Check for class constructor call: ClassName.new(args)
	if sel, ok := call.Func.(*ast.SelectorExpr); ok && sel.Sel == "new" {
		if classIdent, ok := sel.X.(*ast.Ident); ok {
			if _, isClass := a.classes[classIdent.Name]; isClass {
				return NewClassType(classIdent.Name)
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

func (a *Analyzer) analyzeBlock(block *ast.BlockExpr) {
	blockScope := NewScope(ScopeIterator, a.scope)

	// Add block parameters
	for _, param := range block.Params {
		mustDefine(blockScope, NewParam(param, TypeAnyVal))
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
