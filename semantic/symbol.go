package semantic

import (
	"github.com/nchapman/rugby/ast"
)

// SymbolKind classifies symbols in the symbol table.
type SymbolKind int

const (
	SymVariable SymbolKind = iota
	SymFunction
	SymMethod
	SymClass
	SymInterface
	SymModule
	SymField
	SymParam
	SymGoPackage // Go package import
)

// Symbol represents a named entity in the program.
type Symbol struct {
	Name string
	Kind SymbolKind
	Type *Type

	// Position information
	Line   int
	Column int

	// For functions/methods
	Params      []*Symbol // parameter symbols
	ReturnTypes []*Type   // return types
	Variadic    bool      // true if last param is variadic (accepts multiple args)

	// For classes/interfaces
	Fields  map[string]*Symbol // field symbols
	Methods map[string]*Symbol // method symbols

	// For classes
	Parent     string   // parent class name (if any)
	Implements []string // implemented interfaces
	Includes   []string // included modules

	// For fields - accessor flags
	HasGetter bool // field has a getter accessor (getter or property)
	HasSetter bool // field has a setter accessor (setter or property)

	// Visibility
	Public bool

	// AST node reference (for later analysis)
	Node ast.Node
}

// NewSymbol creates a new symbol with the given name, kind, and type.
func NewSymbol(name string, kind SymbolKind, typ *Type) *Symbol {
	return &Symbol{
		Name:    name,
		Kind:    kind,
		Type:    typ,
		Fields:  make(map[string]*Symbol),
		Methods: make(map[string]*Symbol),
	}
}

// NewVariable creates a variable symbol.
func NewVariable(name string, typ *Type) *Symbol {
	return NewSymbol(name, SymVariable, typ)
}

// NewParam creates a parameter symbol.
func NewParam(name string, typ *Type) *Symbol {
	return NewSymbol(name, SymParam, typ)
}

// NewFunction creates a function symbol.
func NewFunction(name string, params []*Symbol, returns []*Type) *Symbol {
	paramTypes := make([]*Type, len(params))
	for i, p := range params {
		paramTypes[i] = p.Type
	}
	sym := NewSymbol(name, SymFunction, NewFuncType(paramTypes, returns))
	sym.Params = params
	sym.ReturnTypes = returns
	return sym
}

// NewMethod creates a method symbol.
func NewMethod(name string, params []*Symbol, returns []*Type) *Symbol {
	sym := NewFunction(name, params, returns)
	sym.Kind = SymMethod
	return sym
}

// NewClass creates a class symbol.
func NewClass(name string) *Symbol {
	return NewSymbol(name, SymClass, NewClassType(name))
}

// NewInterface creates an interface symbol.
func NewInterface(name string) *Symbol {
	return NewSymbol(name, SymInterface, NewInterfaceType(name))
}

// NewModule creates a module symbol.
func NewModule(name string) *Symbol {
	return NewSymbol(name, SymModule, nil)
}

// NewField creates a field symbol.
func NewField(name string, typ *Type) *Symbol {
	return NewSymbol(name, SymField, typ)
}

// AddField adds a field to a class or module symbol.
func (s *Symbol) AddField(field *Symbol) {
	s.Fields[field.Name] = field
}

// AddMethod adds a method to a class, interface, or module symbol.
func (s *Symbol) AddMethod(method *Symbol) {
	s.Methods[method.Name] = method
}

// GetField returns a field by name, or nil if not found.
func (s *Symbol) GetField(name string) *Symbol {
	if s.Fields == nil {
		return nil
	}
	return s.Fields[name]
}

// GetMethod returns a method by name, or nil if not found.
func (s *Symbol) GetMethod(name string) *Symbol {
	if s.Methods == nil {
		return nil
	}
	return s.Methods[name]
}

// String returns a human-readable representation of the symbol.
func (s *Symbol) String() string {
	if s == nil {
		return "<nil>"
	}
	if s.Type == nil {
		return s.Name + " : <no type>"
	}
	return s.Name + " : " + s.Type.String()
}
