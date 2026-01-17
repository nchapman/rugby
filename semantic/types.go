// Package semantic provides type checking and semantic analysis for Rugby programs.
package semantic

import (
	"fmt"
	"strings"
)

// TypeKind classifies types.
type TypeKind int

const (
	TypeUnknown TypeKind = iota
	TypeInt
	TypeInt64
	TypeFloat
	TypeBool
	TypeString
	TypeBytes
	TypeRune
	TypeNil
	TypeAny
	TypeError
	TypeArray
	TypeSet
	TypeMap
	TypeRange
	TypeChan
	TypeTask
	TypeOptional
	TypeTuple     // for multiple return values
	TypeClass     // user-defined class
	TypeInterface // user-defined interface
	TypeFunc      // function type
	TypeTypeParam // generic type parameter (e.g., T in def identity<T>)
)

// Type represents a type in the Rugby type system.
type Type struct {
	Kind TypeKind

	// For composite types
	Elem      *Type // element type for Array, Chan, Task, Optional
	KeyType   *Type // key type for Map
	ValueType *Type // value type for Map

	// For Tuple (multiple return values)
	Elements []*Type

	// For Class, Interface, Func
	Name string

	// For Class - true if this is a struct (value type) rather than a class (reference type)
	IsStruct bool

	// For Func
	Params  []*Type
	Returns []*Type

	// For TypeTypeParam - constraint name (e.g., "Ordered", "Numeric"), empty if unconstrained
	Constraint string
}

// Primitive type singletons for common types.
var (
	TypeUnknownVal = &Type{Kind: TypeUnknown}
	TypeIntVal     = &Type{Kind: TypeInt}
	TypeInt64Val   = &Type{Kind: TypeInt64}
	TypeFloatVal   = &Type{Kind: TypeFloat}
	TypeBoolVal    = &Type{Kind: TypeBool}
	TypeStringVal  = &Type{Kind: TypeString}
	TypeBytesVal   = &Type{Kind: TypeBytes}
	TypeRuneVal    = &Type{Kind: TypeRune}
	TypeNilVal     = &Type{Kind: TypeNil}
	TypeAnyVal     = &Type{Kind: TypeAny}
	TypeErrorVal   = &Type{Kind: TypeError}
	TypeRangeVal   = &Type{Kind: TypeRange}
)

// NewArrayType creates an array type with the given element type.
func NewArrayType(elem *Type) *Type {
	return &Type{Kind: TypeArray, Elem: elem}
}

// NewSetType creates a set type with the given element type.
func NewSetType(elem *Type) *Type {
	return &Type{Kind: TypeSet, Elem: elem}
}

// NewMapType creates a map type with the given key and value types.
func NewMapType(key, value *Type) *Type {
	return &Type{Kind: TypeMap, KeyType: key, ValueType: value}
}

// NewChanType creates a channel type with the given element type.
func NewChanType(elem *Type) *Type {
	return &Type{Kind: TypeChan, Elem: elem}
}

// NewTaskType creates a task type with the given result type.
func NewTaskType(result *Type) *Type {
	return &Type{Kind: TypeTask, Elem: result}
}

// NewOptionalType creates an optional type wrapping the given type.
func NewOptionalType(inner *Type) *Type {
	return &Type{Kind: TypeOptional, Elem: inner}
}

// NewTupleType creates a tuple type for multiple return values.
func NewTupleType(elems ...*Type) *Type {
	return &Type{Kind: TypeTuple, Elements: elems}
}

// NewClassType creates a reference to a class type.
func NewClassType(name string) *Type {
	return &Type{Kind: TypeClass, Name: name}
}

// NewInterfaceType creates a reference to an interface type.
func NewInterfaceType(name string) *Type {
	return &Type{Kind: TypeInterface, Name: name}
}

// NewFuncType creates a function type.
func NewFuncType(params, returns []*Type) *Type {
	return &Type{Kind: TypeFunc, Params: params, Returns: returns}
}

// NewTypeParamType creates a type parameter type (e.g., T in generics).
func NewTypeParamType(name string) *Type {
	return &Type{Kind: TypeTypeParam, Name: name}
}

// NewConstrainedTypeParamType creates a type parameter with a constraint.
func NewConstrainedTypeParamType(name, constraint string) *Type {
	return &Type{Kind: TypeTypeParam, Name: name, Constraint: constraint}
}

// String returns a human-readable representation of the type.
func (t *Type) String() string {
	if t == nil {
		return "unknown"
	}
	switch t.Kind {
	case TypeUnknown:
		return "unknown"
	case TypeInt:
		return "Int"
	case TypeInt64:
		return "Int64"
	case TypeFloat:
		return "Float"
	case TypeBool:
		return "Bool"
	case TypeString:
		return "String"
	case TypeBytes:
		return "Bytes"
	case TypeRune:
		return "Rune"
	case TypeNil:
		return "nil"
	case TypeAny:
		return "any"
	case TypeError:
		return "error"
	case TypeArray:
		if t.Elem != nil {
			return fmt.Sprintf("Array<%s>", t.Elem)
		}
		return "Array<any>"
	case TypeSet:
		if t.Elem != nil {
			return fmt.Sprintf("Set<%s>", t.Elem)
		}
		return "Set<any>"
	case TypeMap:
		if t.KeyType != nil && t.ValueType != nil {
			return fmt.Sprintf("Map<%s, %s>", t.KeyType, t.ValueType)
		}
		return "Map<any, any>"
	case TypeRange:
		return "Range"
	case TypeChan:
		if t.Elem != nil {
			return fmt.Sprintf("Chan<%s>", t.Elem)
		}
		return "Chan<any>"
	case TypeTask:
		if t.Elem != nil {
			return fmt.Sprintf("Task<%s>", t.Elem)
		}
		return "Task<any>"
	case TypeOptional:
		if t.Elem != nil {
			return fmt.Sprintf("%s?", t.Elem)
		}
		return "any?"
	case TypeTuple:
		if len(t.Elements) == 0 {
			return "()"
		}
		parts := make([]string, len(t.Elements))
		for i, e := range t.Elements {
			parts[i] = e.String()
		}
		return fmt.Sprintf("(%s)", strings.Join(parts, ", "))
	case TypeClass:
		return t.Name
	case TypeInterface:
		return t.Name
	case TypeFunc:
		params := make([]string, len(t.Params))
		for i, p := range t.Params {
			params[i] = p.String()
		}
		returns := make([]string, len(t.Returns))
		for i, r := range t.Returns {
			returns[i] = r.String()
		}
		if len(returns) == 0 {
			return fmt.Sprintf("(%s): ()", strings.Join(params, ", "))
		}
		if len(returns) == 1 {
			return fmt.Sprintf("(%s): %s", strings.Join(params, ", "), returns[0])
		}
		return fmt.Sprintf("(%s): (%s)", strings.Join(params, ", "), strings.Join(returns, ", "))
	case TypeTypeParam:
		return t.Name
	default:
		return "unknown"
	}
}

// Equals checks if two types are equal.
func (t *Type) Equals(other *Type) bool {
	if t == nil || other == nil {
		return t == other
	}
	if t.Kind != other.Kind {
		return false
	}
	switch t.Kind {
	case TypeArray, TypeSet, TypeChan, TypeTask, TypeOptional:
		if t.Elem == nil || other.Elem == nil {
			return t.Elem == other.Elem
		}
		return t.Elem.Equals(other.Elem)
	case TypeMap:
		keyEq := (t.KeyType == nil && other.KeyType == nil) ||
			(t.KeyType != nil && other.KeyType != nil && t.KeyType.Equals(other.KeyType))
		valEq := (t.ValueType == nil && other.ValueType == nil) ||
			(t.ValueType != nil && other.ValueType != nil && t.ValueType.Equals(other.ValueType))
		return keyEq && valEq
	case TypeTuple:
		if len(t.Elements) != len(other.Elements) {
			return false
		}
		for i := range t.Elements {
			if !t.Elements[i].Equals(other.Elements[i]) {
				return false
			}
		}
		return true
	case TypeClass, TypeInterface:
		return t.Name == other.Name
	case TypeTypeParam:
		return t.Name == other.Name && t.Constraint == other.Constraint
	case TypeFunc:
		if len(t.Params) != len(other.Params) || len(t.Returns) != len(other.Returns) {
			return false
		}
		for i := range t.Params {
			if !t.Params[i].Equals(other.Params[i]) {
				return false
			}
		}
		for i := range t.Returns {
			if !t.Returns[i].Equals(other.Returns[i]) {
				return false
			}
		}
		return true
	default:
		return true
	}
}

// IsOptional returns true if the type is optional.
func (t *Type) IsOptional() bool {
	return t != nil && t.Kind == TypeOptional
}

// IsError returns true if the type is the error type.
func (t *Type) IsError() bool {
	return t != nil && t.Kind == TypeError
}

// Unwrap returns the inner type for optional types, or the type itself.
func (t *Type) Unwrap() *Type {
	if t != nil && t.Kind == TypeOptional {
		return t.Elem
	}
	return t
}

// GoType returns the Go type string for this type.
// For example: "int", "string", "[]int", "map[string]int".
// Returns empty string if type is unknown or cannot be represented.
func (t *Type) GoType() string {
	if t == nil {
		return ""
	}
	switch t.Kind {
	case TypeInt:
		return "int"
	case TypeInt64:
		return "int64"
	case TypeFloat:
		return "float64"
	case TypeBool:
		return "bool"
	case TypeString:
		return "string"
	case TypeBytes:
		return "[]byte"
	case TypeRune:
		return "rune"
	case TypeNil:
		return "any" // nil can be any type
	case TypeAny, TypeUnknown:
		return "any"
	case TypeError:
		return "error"
	case TypeArray:
		if t.Elem != nil {
			elemType := t.Elem.GoType()
			if elemType != "" {
				return "[]" + elemType
			}
		}
		return "[]any"
	case TypeSet:
		if t.Elem != nil {
			elemType := t.Elem.GoType()
			if elemType != "" {
				return "map[" + elemType + "]struct{}"
			}
		}
		return "map[any]struct{}"
	case TypeMap:
		keyType := "any"
		valType := "any"
		if t.KeyType != nil {
			if kt := t.KeyType.GoType(); kt != "" {
				keyType = kt
			}
		}
		if t.ValueType != nil {
			if vt := t.ValueType.GoType(); vt != "" {
				valType = vt
			}
		}
		return "map[" + keyType + "]" + valType
	case TypeClass:
		if t.Name != "" {
			// Structs are value types - no pointer prefix
			if t.IsStruct {
				return t.Name
			}
			return "*" + t.Name
		}
		return "any"
	case TypeInterface:
		if t.Name != "" {
			return t.Name
		}
		return "any"
	case TypeTypeParam:
		// Return the type parameter name as-is for generic code generation
		if t.Name != "" {
			return t.Name
		}
		return "any"
	case TypeFunc:
		// Generate Go function type: func(params) returns
		params := make([]string, len(t.Params))
		for i, p := range t.Params {
			params[i] = p.GoType()
		}
		returns := make([]string, len(t.Returns))
		for i, r := range t.Returns {
			returns[i] = r.GoType()
		}
		sig := "func(" + strings.Join(params, ", ") + ")"
		if len(returns) == 1 {
			sig += " " + returns[0]
		} else if len(returns) > 1 {
			sig += " (" + strings.Join(returns, ", ") + ")"
		}
		return sig
	default:
		return ""
	}
}
