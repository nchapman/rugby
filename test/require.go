// Package test provides assertion functions for Rugby tests.
package test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/nchapman/rugby/runtime"
)

// Require provides assertion methods that call t.Fatalf on failure.
// These assertions stop test execution immediately on failure.
type Require struct{}

// Equal asserts that expected and actual are equal.
// Stops test execution on failure.
func (Require) Equal(t testing.TB, expected, actual any) {
	t.Helper()
	if !runtime.Equal(expected, actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

// NotEqual asserts that expected and actual are not equal.
// Stops test execution on failure.
func (Require) NotEqual(t testing.TB, expected, actual any) {
	t.Helper()
	if runtime.Equal(expected, actual) {
		t.Fatalf("expected values to differ, but both are %v", expected)
	}
}

// True asserts that value is true.
// Stops test execution on failure.
func (Require) True(t testing.TB, value bool) {
	t.Helper()
	if !value {
		t.Fatalf("expected true, got false")
	}
}

// False asserts that value is false.
// Stops test execution on failure.
func (Require) False(t testing.TB, value bool) {
	t.Helper()
	if value {
		t.Fatalf("expected false, got true")
	}
}

// Nil asserts that value is nil.
// Stops test execution on failure.
func (Require) Nil(t testing.TB, value any) {
	t.Helper()
	if !isNil(value) {
		t.Fatalf("expected nil, got %v", value)
	}
}

// NotNil asserts that value is not nil.
// Stops test execution on failure.
func (Require) NotNil(t testing.TB, value any) {
	t.Helper()
	if isNil(value) {
		t.Fatalf("expected non-nil value, got nil")
	}
}

// Error asserts that err is not nil.
// Stops test execution on failure.
func (Require) Error(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
}

// NoError asserts that err is nil.
// Stops test execution on failure.
func (Require) NoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// Contains asserts that container contains the substring or element.
// Stops test execution on failure.
func (Require) Contains(t testing.TB, container, element any) {
	t.Helper()
	switch c := container.(type) {
	case string:
		s, ok := element.(string)
		if !ok {
			t.Fatalf("expected string element for string container, got %T", element)
		}
		if !strings.Contains(c, s) {
			t.Fatalf("expected %q to contain %q", c, s)
		}
	default:
		// Handle slices using reflection
		rv := reflect.ValueOf(container)
		if rv.Kind() == reflect.Slice {
			for i := range rv.Len() {
				if runtime.Equal(rv.Index(i).Interface(), element) {
					return
				}
			}
			t.Fatalf("expected slice to contain %v", element)
		}
		t.Fatalf("Contains not supported for type %T", container)
	}
}
