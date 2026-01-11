package runtime

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
