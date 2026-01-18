package runtime

import (
	"reflect"
	"strings"
)

// Option is a generic optional type for storing optional values in fields,
// arrays, and maps. For return values, Rugby uses Go's (T, bool) pattern instead.
type Option[T any] struct {
	Value T
	Ok    bool
}

// SomeOpt creates an Option containing a value
func SomeOpt[T any](v T) Option[T] {
	return Option[T]{Value: v, Ok: true}
}

// NoneOpt creates an empty Option
func NoneOpt[T any]() Option[T] {
	return Option[T]{Ok: false}
}

// IsPresent returns true if the Option contains a value
func (o Option[T]) IsPresent() bool {
	return o.Ok
}

// IsAbsent returns true if the Option is empty
func (o Option[T]) IsAbsent() bool {
	return !o.Ok
}

// Unwrap returns the value or panics if empty
func (o Option[T]) Unwrap() T {
	if !o.Ok {
		panic("called Unwrap on empty Option")
	}
	return o.Value
}

// UnwrapOr returns the value or the provided default if empty
func (o Option[T]) UnwrapOr(def T) T {
	if !o.Ok {
		return def
	}
	return o.Value
}

// ToTuple converts the Option to a (value, ok) tuple
func (o Option[T]) ToTuple() (T, bool) {
	return o.Value, o.Ok
}

// FromTuple creates an Option from a (value, ok) tuple
func FromTuple[T any](v T, ok bool) Option[T] {
	return Option[T]{Value: v, Ok: ok}
}

// Generic pointer-based optional helpers
// Optional types are represented as pointers (*T) for storage.

// Some wraps a value in a pointer (creates a present optional).
// Ruby: equivalent to having a value vs nil
func Some[T any](v T) *T {
	return &v
}

// Coalesce returns the dereferenced value if present, otherwise returns the default.
// Ruby: value ?? default
func Coalesce[T any](opt *T, def T) T {
	if opt != nil {
		return *opt
	}
	return def
}

// CoalesceOpt returns the first non-nil optional, for chaining optionals.
// Ruby: a ?? b where both are optional
func CoalesceOpt[T any](opt *T, def *T) *T {
	if opt != nil {
		return opt
	}
	return def
}

// Legacy type-specific helpers (for backward compatibility)
// New code should use the generic Some and Coalesce functions above.

func SomeInt(v int) *int { return Some(v) }
func NoneInt() *int      { return nil }
func ToOptionalInt(v int, ok bool) *int {
	if ok {
		return Some(v)
	}
	return nil
}
func SomeInt64(v int64) *int64 { return Some(v) }
func NoneInt64() *int64        { return nil }
func ToOptionalInt64(v int64, ok bool) *int64 {
	if ok {
		return Some(v)
	}
	return nil
}
func SomeFloat(v float64) *float64 { return Some(v) }
func NoneFloat() *float64          { return nil }
func ToOptionalFloat(v float64, ok bool) *float64 {
	if ok {
		return Some(v)
	}
	return nil
}
func SomeString(v string) *string { return Some(v) }
func NoneString() *string         { return nil }
func ToOptionalString(v string, ok bool) *string {
	if ok {
		return Some(v)
	}
	return nil
}
func SomeBool(v bool) *bool { return Some(v) }
func NoneBool() *bool       { return nil }
func ToOptionalBool(v bool, ok bool) *bool {
	if ok {
		return Some(v)
	}
	return nil
}

// Nil coalescing helpers - delegate to generic Coalesce

func CoalesceInt(opt *int, def int) int                    { return Coalesce(opt, def) }
func CoalesceIntOpt(opt *int, def *int) *int               { return CoalesceOpt(opt, def) }
func CoalesceInt64(opt *int64, def int64) int64            { return Coalesce(opt, def) }
func CoalesceInt64Opt(opt *int64, def *int64) *int64       { return CoalesceOpt(opt, def) }
func CoalesceFloat(opt *float64, def float64) float64      { return Coalesce(opt, def) }
func CoalesceFloatOpt(opt *float64, def *float64) *float64 { return CoalesceOpt(opt, def) }
func CoalesceString(opt *string, def string) string        { return Coalesce(opt, def) }
func CoalesceStringOpt(opt *string, def *string) *string   { return CoalesceOpt(opt, def) }
func CoalesceBool(opt *bool, def bool) bool                { return Coalesce(opt, def) }
func CoalesceBoolOpt(opt *bool, def *bool) *bool           { return CoalesceOpt(opt, def) }

// CoalesceStringAny handles nil coalescing for safe navigation results (which return any).
// If opt is nil or is a nil interface, returns def.
// If opt is a string, returns it directly.
// If opt is a *string, returns the dereferenced value.
func CoalesceStringAny(opt any, def string) string {
	if opt == nil {
		return def
	}
	switch v := opt.(type) {
	case string:
		return v
	case *string:
		if v != nil {
			return *v
		}
		return def
	default:
		return def
	}
}

// OptionalResult wraps a result from optional map operations.
// Use Ok() to check if present, Unwrap() to get the value.
type OptionalResult struct {
	Value   any
	Present bool
}

// Ok returns true if the optional result contains a value.
func (r OptionalResult) Ok() bool {
	return r.Present
}

// Unwrap returns the value or panics if empty.
func (r OptionalResult) Unwrap() any {
	return r.Value
}

// OptionalMap applies a function to an optional value if present.
// For pointer-based optionals (*T).
func OptionalMap[T any, R any](opt *T, fn func(T) R) OptionalResult {
	if opt == nil {
		return OptionalResult{Present: false}
	}
	return OptionalResult{Value: fn(*opt), Present: true}
}

// OptionalMapString is a type-specific version for *string optionals.
func OptionalMapString(opt *string, fn func(string) string) OptionalResult {
	if opt == nil {
		return OptionalResult{Present: false}
	}
	return OptionalResult{Value: fn(*opt), Present: true}
}

// OptionalMapAny handles optionals when the inner type is unknown.
func OptionalMapAny(opt any, fn func(any) any) OptionalResult {
	if opt == nil {
		return OptionalResult{Present: false}
	}
	return OptionalResult{Value: fn(opt), Present: true}
}

// OptionalEach executes a function on an optional value if present.
func OptionalEach[T any](opt *T, fn func(T)) {
	if opt != nil {
		fn(*opt)
	}
}

// OptionalEachString is a type-specific version for *string optionals.
func OptionalEachString(opt *string, fn func(string)) {
	if opt != nil {
		fn(*opt)
	}
}

// OptionalEachAny handles optionals when the inner type is unknown.
func OptionalEachAny(opt any, fn func(any)) {
	if opt != nil {
		fn(opt)
	}
}

// SafeNavField accesses a field/method on an 'any' value for chained safe navigation.
// The value is typically a pointer wrapped in 'any'. We use reflection to call the getter method.
// Since Go reflection only works with exported methods, we need to check the original
// (pointer) value for methods, not the dereferenced struct.
func SafeNavField(obj any, fieldName string) any {
	if obj == nil {
		return nil
	}
	v := reflect.ValueOf(obj)
	// Check for nil pointer
	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
		return nil
	}

	// Try to find and call a method on the pointer value first (methods are on *T)
	// Try lowercase first (Rugby private classes), then uppercase (pub classes)
	method := v.MethodByName(fieldName)
	if !method.IsValid() {
		// Try with uppercase first letter
		upperName := strings.ToUpper(fieldName[:1]) + fieldName[1:]
		method = v.MethodByName(upperName)
	}
	if method.IsValid() {
		results := method.Call(nil)
		if len(results) > 0 {
			return results[0].Interface()
		}
		return nil
	}

	// If it's a pointer, get the element to access fields
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	// Try to access as a field
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		// Try with underscore prefix (Rugby internal field naming)
		field = v.FieldByName("_" + fieldName)
	}
	if !field.IsValid() {
		// Try with uppercase first letter
		upperName := strings.ToUpper(fieldName[:1]) + fieldName[1:]
		field = v.FieldByName(upperName)
	}
	if field.IsValid() {
		return field.Interface()
	}
	return nil
}
