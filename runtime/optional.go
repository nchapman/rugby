package runtime

// Optional types for value types that need (T, valid) semantics.
// Reference types (classes, slices, maps) use nil instead.

// OptionalInt represents an optional int value
type OptionalInt struct {
	Value int
	Valid bool
}

// OptionalInt64 represents an optional int64 value
type OptionalInt64 struct {
	Value int64
	Valid bool
}

// OptionalFloat represents an optional float64 value
type OptionalFloat struct {
	Value float64
	Valid bool
}

// OptionalString represents an optional string value
type OptionalString struct {
	Value string
	Valid bool
}

// OptionalBool represents an optional bool value
type OptionalBool struct {
	Value bool
	Valid bool
}

// Helper functions for creating optionals

func SomeInt(v int) OptionalInt {
	return OptionalInt{Value: v, Valid: true}
}

func NoneInt() OptionalInt {
	return OptionalInt{Valid: false}
}

func SomeInt64(v int64) OptionalInt64 {
	return OptionalInt64{Value: v, Valid: true}
}

func NoneInt64() OptionalInt64 {
	return OptionalInt64{Valid: false}
}

func SomeFloat(v float64) OptionalFloat {
	return OptionalFloat{Value: v, Valid: true}
}

func NoneFloat() OptionalFloat {
	return OptionalFloat{Valid: false}
}

func SomeString(v string) OptionalString {
	return OptionalString{Value: v, Valid: true}
}

func NoneString() OptionalString {
	return OptionalString{Valid: false}
}

func SomeBool(v bool) OptionalBool {
	return OptionalBool{Value: v, Valid: true}
}

func NoneBool() OptionalBool {
	return OptionalBool{Valid: false}
}
