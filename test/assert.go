// Package test provides assertion functions for Rugby tests.
// It wraps standard Go testing patterns with a Ruby-like API.
package test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/nchapman/rugby/runtime"
)

// Assert provides assertion methods that call t.Errorf on failure.
// These assertions do not stop test execution.
type Assert struct{}

// Equal asserts that expected and actual are equal.
// Uses runtime.Equal for deep comparison.
func (Assert) Equal(t testing.TB, expected, actual any) {
	t.Helper()
	if !runtime.Equal(expected, actual) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

// NotEqual asserts that expected and actual are not equal.
func (Assert) NotEqual(t testing.TB, expected, actual any) {
	t.Helper()
	if runtime.Equal(expected, actual) {
		t.Errorf("expected values to differ, but both are %v", expected)
	}
}

// True asserts that value is true.
func (Assert) True(t testing.TB, value bool) {
	t.Helper()
	if !value {
		t.Errorf("expected true, got false")
	}
}

// False asserts that value is false.
func (Assert) False(t testing.TB, value bool) {
	t.Helper()
	if value {
		t.Errorf("expected false, got true")
	}
}

// Nil asserts that value is nil.
func (Assert) Nil(t testing.TB, value any) {
	t.Helper()
	if !isNil(value) {
		t.Errorf("expected nil, got %v", value)
	}
}

// NotNil asserts that value is not nil.
func (Assert) NotNil(t testing.TB, value any) {
	t.Helper()
	if isNil(value) {
		t.Errorf("expected non-nil value, got nil")
	}
}

// Error asserts that err is not nil.
func (Assert) Error(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}

// NoError asserts that err is nil.
func (Assert) NoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// Contains asserts that container contains the substring or element.
func (Assert) Contains(t testing.TB, container, element any) {
	t.Helper()
	switch c := container.(type) {
	case string:
		s, ok := element.(string)
		if !ok {
			t.Errorf("expected string element for string container, got %T", element)
			return
		}
		if !strings.Contains(c, s) {
			t.Errorf("expected %q to contain %q", c, s)
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
			t.Errorf("expected slice to contain %v", element)
			return
		}
		t.Errorf("Contains not supported for type %T", container)
	}
}

// isNil checks if a value is nil, handling interface nil correctly.
func isNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}
	return false
}
