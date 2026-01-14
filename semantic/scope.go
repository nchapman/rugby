package semantic

// ScopeKind identifies the type of scope.
type ScopeKind int

const (
	ScopeGlobal ScopeKind = iota
	ScopeFunction
	ScopeMethod
	ScopeClass
	ScopeModule
	ScopeBlock    // for if-then, case branches, etc.
	ScopeLoop     // for while, until, for loops
	ScopeIterator // for block parameters in each/map/etc.
)

// Scope represents a lexical scope in the program.
type Scope struct {
	Kind    ScopeKind
	Parent  *Scope
	symbols map[string]*Symbol

	// For method/function scopes
	ReturnTypes []*Type // expected return types

	// For class/module scopes
	ClassName  string // name of enclosing class (if any)
	ModuleName string // name of enclosing module (if any)
}

// NewScope creates a new scope with the given kind and parent.
func NewScope(kind ScopeKind, parent *Scope) *Scope {
	s := &Scope{
		Kind:    kind,
		Parent:  parent,
		symbols: make(map[string]*Symbol),
	}
	// Inherit class/module context from parent
	if parent != nil {
		s.ClassName = parent.ClassName
		s.ModuleName = parent.ModuleName
	}
	return s
}

// NewGlobalScope creates the root global scope.
func NewGlobalScope() *Scope {
	return NewScope(ScopeGlobal, nil)
}

// Define adds a symbol to this scope.
// Returns an error if the symbol is already defined in this exact scope.
func (s *Scope) Define(sym *Symbol) error {
	if existing := s.symbols[sym.Name]; existing != nil {
		return &RedefinitionError{
			Name:       sym.Name,
			PrevLine:   existing.Line,
			PrevColumn: existing.Column,
			NewLine:    sym.Line,
			NewColumn:  sym.Column,
		}
	}
	s.symbols[sym.Name] = sym
	return nil
}

// DefineOrShadow adds a symbol to this scope, allowing shadowing of parent scopes.
// Returns an error only if the symbol is already defined in THIS scope.
func (s *Scope) DefineOrShadow(sym *Symbol) error {
	if existing := s.symbols[sym.Name]; existing != nil {
		return &RedefinitionError{
			Name:       sym.Name,
			PrevLine:   existing.Line,
			PrevColumn: existing.Column,
			NewLine:    sym.Line,
			NewColumn:  sym.Column,
		}
	}
	s.symbols[sym.Name] = sym
	return nil
}

// Lookup searches for a symbol by name in this scope and parent scopes.
// Returns nil if not found.
func (s *Scope) Lookup(name string) *Symbol {
	if sym := s.symbols[name]; sym != nil {
		return sym
	}
	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}
	return nil
}

// LookupLocal searches for a symbol only in this scope, not parent scopes.
func (s *Scope) LookupLocal(name string) *Symbol {
	return s.symbols[name]
}

// IsDefinedLocally returns true if a symbol with the given name exists in this scope.
func (s *Scope) IsDefinedLocally(name string) bool {
	return s.symbols[name] != nil
}

// IsDefined returns true if a symbol with the given name exists in this or any parent scope.
func (s *Scope) IsDefined(name string) bool {
	return s.Lookup(name) != nil
}

// Update updates an existing symbol's type in the scope where it's defined.
// Returns false if the symbol doesn't exist.
func (s *Scope) Update(name string, typ *Type) bool {
	if sym := s.symbols[name]; sym != nil {
		sym.Type = typ
		return true
	}
	if s.Parent != nil {
		return s.Parent.Update(name, typ)
	}
	return false
}

// Symbols returns all symbols defined in this scope (not including parents).
func (s *Scope) Symbols() []*Symbol {
	result := make([]*Symbol, 0, len(s.symbols))
	for _, sym := range s.symbols {
		result = append(result, sym)
	}
	return result
}

// IsInsideClass returns true if this scope is inside a class.
func (s *Scope) IsInsideClass() bool {
	return s.ClassName != ""
}

// IsInsideModule returns true if this scope is inside a module.
func (s *Scope) IsInsideModule() bool {
	return s.ModuleName != ""
}

// IsInsideLoop returns true if there's a loop or iterator block in the scope chain.
func (s *Scope) IsInsideLoop() bool {
	if s.Kind == ScopeLoop || s.Kind == ScopeIterator {
		return true
	}
	if s.Parent != nil && s.Kind != ScopeFunction && s.Kind != ScopeMethod {
		return s.Parent.IsInsideLoop()
	}
	return false
}

// IsInsideIterator returns true if there's an iterator block (each/map/etc.) in the scope chain
// without an intervening loop. A loop inside an iterator block creates a new context that allows
// break/next, so we stop traversal at loop scopes.
func (s *Scope) IsInsideIterator() bool {
	if s.Kind == ScopeIterator {
		return true
	}
	// Stop at function/method/loop boundaries - loops inside iterator blocks allow break/next
	if s.Parent != nil && s.Kind != ScopeFunction && s.Kind != ScopeMethod && s.Kind != ScopeLoop {
		return s.Parent.IsInsideIterator()
	}
	return false
}

// FunctionScope returns the nearest enclosing function/method scope, or nil.
func (s *Scope) FunctionScope() *Scope {
	if s.Kind == ScopeFunction || s.Kind == ScopeMethod {
		return s
	}
	if s.Parent != nil {
		return s.Parent.FunctionScope()
	}
	return nil
}

// ClassScope returns the nearest enclosing class scope, or nil.
func (s *Scope) ClassScope() *Scope {
	if s.Kind == ScopeClass {
		return s
	}
	if s.Parent != nil {
		return s.Parent.ClassScope()
	}
	return nil
}

// RedefinitionError is returned when a symbol is redefined in the same scope.
type RedefinitionError struct {
	Name       string
	PrevLine   int
	PrevColumn int
	NewLine    int
	NewColumn  int
}

func (e *RedefinitionError) Error() string {
	return "'" + e.Name + "' already defined in this scope"
}
