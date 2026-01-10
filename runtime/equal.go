// Package runtime provides Ruby-like stdlib ergonomics for Rugby programs.
package runtime

import "reflect"

// Equaler is the interface for types that implement custom equality.
// Rugby classes with def ==(other) compile to this interface.
type Equaler interface {
	Equal(other interface{}) bool
}

// Equal compares two values with dispatch logic:
// 1. If a implements Equaler, call a.Equal(b)
// 2. If both are slices, perform deep comparison
// 3. If both are maps, perform deep comparison
// 4. Otherwise, use Go's == (via reflect.DeepEqual for interface{} safety)
func Equal(a, b interface{}) bool {
	// Check if a implements custom equality
	if eq, ok := a.(Equaler); ok {
		return eq.Equal(b)
	}

	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Get reflect values
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// Handle slices
	if va.Kind() == reflect.Slice && vb.Kind() == reflect.Slice {
		return sliceEqual(va, vb)
	}

	// Handle maps
	if va.Kind() == reflect.Map && vb.Kind() == reflect.Map {
		return mapEqual(va, vb)
	}

	// For primitives and structs without Equaler, use DeepEqual
	// This handles int, string, bool, float64, and struct comparisons safely
	return reflect.DeepEqual(a, b)
}

// sliceEqual compares two slices element by element.
func sliceEqual(a, b reflect.Value) bool {
	if a.Len() != b.Len() {
		return false
	}
	for i := 0; i < a.Len(); i++ {
		if !Equal(a.Index(i).Interface(), b.Index(i).Interface()) {
			return false
		}
	}
	return true
}

// mapEqual compares two maps key by key.
func mapEqual(a, b reflect.Value) bool {
	if a.Len() != b.Len() {
		return false
	}
	for _, key := range a.MapKeys() {
		aVal := a.MapIndex(key)
		bVal := b.MapIndex(key)
		if !bVal.IsValid() {
			return false
		}
		if !Equal(aVal.Interface(), bVal.Interface()) {
			return false
		}
	}
	return true
}
