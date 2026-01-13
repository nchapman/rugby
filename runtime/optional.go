package runtime

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

// Legacy pointer-based optional helpers (for backward compatibility)
// Optional types are now represented as pointers (*T) for storage.
// Helper functions for creating optionals.

// Int helpers

func SomeInt(v int) *int {
	return &v
}

func NoneInt() *int {
	return nil
}

func ToOptionalInt(v int, ok bool) *int {
	if !ok {
		return nil
	}
	return &v
}

// Int64 helpers

func SomeInt64(v int64) *int64 {
	return &v
}

func NoneInt64() *int64 {
	return nil
}

func ToOptionalInt64(v int64, ok bool) *int64 {
	if !ok {
		return nil
	}
	return &v
}

// Float helpers

func SomeFloat(v float64) *float64 {
	return &v
}

func NoneFloat() *float64 {
	return nil
}

func ToOptionalFloat(v float64, ok bool) *float64 {
	if !ok {
		return nil
	}
	return &v
}

// String helpers

func SomeString(v string) *string {
	return &v
}

func NoneString() *string {
	return nil
}

func ToOptionalString(v string, ok bool) *string {
	if !ok {
		return nil
	}
	return &v
}

// Bool helpers

func SomeBool(v bool) *bool {
	return &v
}

func NoneBool() *bool {
	return nil
}

func ToOptionalBool(v bool, ok bool) *bool {
	if !ok {
		return nil
	}
	return &v
}

// Nil coalescing helpers - return the value if present, otherwise return the default

func CoalesceInt(opt *int, def int) int {
	if opt != nil {
		return *opt
	}
	return def
}

func CoalesceInt64(opt *int64, def int64) int64 {
	if opt != nil {
		return *opt
	}
	return def
}

func CoalesceFloat(opt *float64, def float64) float64 {
	if opt != nil {
		return *opt
	}
	return def
}

func CoalesceString(opt *string, def string) string {
	if opt != nil {
		return *opt
	}
	return def
}

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

func CoalesceBool(opt *bool, def bool) bool {
	if opt != nil {
		return *opt
	}
	return def
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
