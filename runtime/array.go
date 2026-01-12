// Package runtime provides Ruby-like stdlib ergonomics for Rugby programs.
package runtime

import (
	"cmp"
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"strings"
)

// Each iterates over a slice, calling the function for each element.
// The callback returns false to break, true to continue.
// Uses any with type switches to support both typed and untyped slices.
func Each(slice any, fn func(any) bool) {
	switch s := slice.(type) {
	case []any:
		for _, v := range s {
			if !fn(v) {
				break
			}
		}
	case []int:
		for _, v := range s {
			if !fn(v) {
				break
			}
		}
	case []string:
		for _, v := range s {
			if !fn(v) {
				break
			}
		}
	case []float64:
		for _, v := range s {
			if !fn(v) {
				break
			}
		}
	case []bool:
		for _, v := range s {
			if !fn(v) {
				break
			}
		}
	default:
		// Use reflection as fallback for other slice types
		val := reflect.ValueOf(slice)
		if val.Kind() != reflect.Slice {
			panic(fmt.Sprintf("Each: expected slice, got %T", slice))
		}
		for i := range val.Len() {
			if !fn(val.Index(i).Interface()) {
				break
			}
		}
	}
}

// EachWithIndex iterates over a slice with index, calling the function for each element.
// The callback returns false to break, true to continue.
// Uses any with type switches to support both typed and untyped slices.
func EachWithIndex(slice any, fn func(any, int) bool) {
	switch s := slice.(type) {
	case []any:
		for i, v := range s {
			if !fn(v, i) {
				break
			}
		}
	case []int:
		for i, v := range s {
			if !fn(v, i) {
				break
			}
		}
	case []string:
		for i, v := range s {
			if !fn(v, i) {
				break
			}
		}
	case []float64:
		for i, v := range s {
			if !fn(v, i) {
				break
			}
		}
	case []bool:
		for i, v := range s {
			if !fn(v, i) {
				break
			}
		}
	default:
		// Use reflection as fallback for other slice types
		val := reflect.ValueOf(slice)
		if val.Kind() != reflect.Slice {
			panic(fmt.Sprintf("EachWithIndex: expected slice, got %T", slice))
		}
		for i := range val.Len() {
			if !fn(val.Index(i).Interface(), i) {
				break
			}
		}
	}
}

// Select returns elements for which the predicate returns true.
// The predicate returns (match, continue).
func Select[T any](slice []T, predicate func(T) (bool, bool)) []T {
	result := make([]T, 0)
	for _, v := range slice {
		match, cont := predicate(v)
		if match {
			result = append(result, v)
		}
		if !cont {
			break
		}
	}
	return result
}

// Reject returns elements for which the predicate returns false.
// The predicate returns (match, continue).
func Reject[T any](slice []T, predicate func(T) (bool, bool)) []T {
	result := make([]T, 0)
	for _, v := range slice {
		match, cont := predicate(v)
		if !match {
			result = append(result, v)
		}
		if !cont {
			break
		}
	}
	return result
}

// Map transforms each element using the mapper function.
// The mapper returns (result, include, continue).
// include=false means skip this element (next), continue=false means stop (break).
func Map[T, R any](slice []T, mapper func(T) (R, bool, bool)) []R {
	result := make([]R, 0, len(slice))
	for _, v := range slice {
		val, include, cont := mapper(v)
		if include {
			result = append(result, val)
		}
		if !cont {
			break
		}
	}
	return result
}

// Reduce folds the slice into a single value using the reducer function.
// The reducer returns (accumulator, continue).
func Reduce[T, R any](slice []T, initial R, reducer func(R, T) (R, bool)) R {
	acc := initial
	for _, v := range slice {
		newAcc, cont := reducer(acc, v)
		acc = newAcc
		if !cont {
			break
		}
	}
	return acc
}

// Find returns the first element matching the predicate.
// The predicate returns (match, continue).
func Find[T any](slice []T, predicate func(T) (bool, bool)) (T, bool) {
	for _, v := range slice {
		match, cont := predicate(v)
		if match {
			return v, true
		}
		if !cont {
			break
		}
	}
	var zero T
	return zero, false
}

// Any returns true if any element matches the predicate.
func Any[T any](slice []T, predicate func(T) (bool, bool)) bool {
	for _, v := range slice {
		match, cont := predicate(v)
		if match {
			return true
		}
		if !cont {
			break
		}
	}
	return false
}

// All returns true if all elements match the predicate.
func All[T any](slice []T, predicate func(T) (bool, bool)) bool {
	for _, v := range slice {
		match, cont := predicate(v)
		if !match {
			return false
		}
		if !cont {
			break
		}
	}
	return true
}

// None returns true if no elements match the predicate.
func None[T any](slice []T, predicate func(T) (bool, bool)) bool {
	for _, v := range slice {
		match, cont := predicate(v)
		if match {
			return false
		}
		if !cont {
			break
		}
	}
	return true
}

// Contains returns true if the slice contains the value.
// Ruby: arr.include?(5)
func Contains[T comparable](slice []T, value T) bool {
	return slices.Contains(slice, value)
}

// First returns the first element of the slice.
// Ruby: arr.first
func First[T any](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}
	return slice[0], true
}

// Last returns the last element of the slice.
// Ruby: arr.last
func Last[T any](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}
	return slice[len(slice)-1], true
}

// Index returns the element at the given index, supporting negative indices.
// Negative indices count from the end: -1 is last, -2 is second-to-last, etc.
// Ruby: arr[-1], arr[-2]
func Index[T any](slice []T, i int) T {
	if i < 0 {
		i = len(slice) + i
	}
	return slice[i]
}

// IndexOpt returns the element at the given index with bounds checking.
// Returns (value, true) if index is valid, (zero, false) if out of bounds.
// Supports negative indices.
func IndexOpt[T any](slice []T, i int) (T, bool) {
	if i < 0 {
		i = len(slice) + i
	}
	if i < 0 || i >= len(slice) {
		var zero T
		return zero, false
	}
	return slice[i], true
}

// Reverse reverses the slice in-place.
// Ruby: arr.reverse!
func Reverse[T any](slice []T) {
	slices.Reverse(slice)
}

// Reversed returns a new slice with elements in reverse order.
// Ruby: arr.reverse
func Reversed[T any](slice []T) []T {
	result := slices.Clone(slice)
	slices.Reverse(result)
	return result
}

// SumInt returns the sum of all elements (for int slices).
func SumInt(slice []int) int {
	sum := 0
	for _, v := range slice {
		sum += v
	}
	return sum
}

// SumFloat returns the sum of all elements (for float64 slices).
func SumFloat(slice []float64) float64 {
	sum := 0.0
	for _, v := range slice {
		sum += v
	}
	return sum
}

// MinInt returns the minimum value in an int slice.
func MinInt(slice []int) (int, bool) {
	if len(slice) == 0 {
		return 0, false
	}
	return slices.Min(slice), true
}

// MaxInt returns the maximum value in an int slice.
func MaxInt(slice []int) (int, bool) {
	if len(slice) == 0 {
		return 0, false
	}
	return slices.Max(slice), true
}

// MinFloat returns the minimum value in a float64 slice.
// NaN values use IEEE 754 semantics (NaN comparisons are always false).
func MinFloat(slice []float64) (float64, bool) {
	if len(slice) == 0 {
		return 0, false
	}
	return slices.Min(slice), true
}

// MaxFloat returns the maximum value in a float64 slice.
// NaN values use IEEE 754 semantics (NaN comparisons are always false).
func MaxFloat(slice []float64) (float64, bool) {
	if len(slice) == 0 {
		return 0, false
	}
	return slices.Max(slice), true
}

// Join concatenates elements into a string using the separator.
// Ruby: arr.join(sep)
func Join[T any](slice []T, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	var b strings.Builder
	for i, v := range slice {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(fmt.Sprint(v))
	}
	return b.String()
}

// Flatten flattens a slice of slices (one level or deep? Ruby default is deep, but arg is depth).
// For now, let's just do a simple flattening if it's []any.
// Typed slices []int cannot be flattened further.
// Ruby: arr.flatten
func Flatten(slice any) []any {
	result := make([]any, 0)
	val := reflect.ValueOf(slice)
	if val.Kind() != reflect.Slice {
		return []any{slice}
	}

	for i := range val.Len() {
		elem := val.Index(i).Interface()
		// If elem is slice, recurse
		elemVal := reflect.ValueOf(elem)
		if elemVal.Kind() == reflect.Slice {
			result = append(result, Flatten(elem)...)
		} else {
			result = append(result, elem)
		}
	}
	return result
}

// Uniq returns a new slice with unique elements.
// Ruby: arr.uniq
func Uniq[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Sort returns a new sorted slice.
// Ruby: arr.sort
func Sort[T cmp.Ordered](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	slices.Sort(result)
	return result
}

// Shuffle returns a new shuffled slice.
// Ruby: arr.shuffle
func Shuffle[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return result
}

// Sample returns a random element.
// Ruby: arr.sample
func Sample[T any](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}
	return slice[rand.Intn(len(slice))], true
}

// FirstN returns the first n elements.
func FirstN[T any](slice []T, n int) []T {
	if n <= 0 {
		return []T{}
	}
	if n >= len(slice) {
		return slices.Clone(slice)
	}
	return slices.Clone(slice[:n])
}

// LastN returns the last n elements.
func LastN[T any](slice []T, n int) []T {
	if n <= 0 {
		return []T{}
	}
	l := len(slice)
	if n >= l {
		return slices.Clone(slice)
	}
	return slices.Clone(slice[l-n:])
}

// Rotate returns a new slice rotated by n.
// Ruby: arr.rotate(n)
func Rotate[T any](slice []T, n int) []T {
	l := len(slice)
	if l == 0 {
		return []T{}
	}
	n = n % l
	if n < 0 {
		n += l
	}
	if n == 0 {
		return slices.Clone(slice)
	}
	return slices.Concat(slice[n:], slice[:n])
}

// AtIndex returns the element at the given index, supporting negative indices.
// This is a universal function that handles both strings and slices.
// For strings, returns the character (as string) at the index.
// For slices, returns the element at the index.
// Negative indices count from the end: -1 is last, -2 is second-to-last, etc.
// Panics if index is out of bounds.
// Ruby: arr[-1], str[-1]
func AtIndex(collection any, i int) any {
	// Fast paths for common types to avoid reflection overhead
	switch s := collection.(type) {
	case string:
		return StringIndex(s, i)
	case []int:
		if i < 0 {
			i = len(s) + i
		}
		return s[i]
	case []string:
		if i < 0 {
			i = len(s) + i
		}
		return s[i]
	case []float64:
		if i < 0 {
			i = len(s) + i
		}
		return s[i]
	case []bool:
		if i < 0 {
			i = len(s) + i
		}
		return s[i]
	case []any:
		if i < 0 {
			i = len(s) + i
		}
		return s[i]
	}

	// Reflection fallback for other slice types
	val := reflect.ValueOf(collection)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		panic(fmt.Sprintf("AtIndex: expected string, slice or array, got %T", collection))
	}
	length := val.Len()
	if i < 0 {
		i = length + i
	}
	return val.Index(i).Interface()
}

// AtIndexOpt returns the element at the given index with bounds checking.
// Returns (value, true) if index is valid, (nil, false) if out of bounds.
// Supports negative indices and works for both strings and slices.
func AtIndexOpt(collection any, i int) (any, bool) {
	// Fast paths for common types to avoid reflection overhead
	switch s := collection.(type) {
	case string:
		return StringIndexOpt(s, i)
	case []int:
		if i < 0 {
			i = len(s) + i
		}
		if i < 0 || i >= len(s) {
			return nil, false
		}
		return s[i], true
	case []string:
		if i < 0 {
			i = len(s) + i
		}
		if i < 0 || i >= len(s) {
			return nil, false
		}
		return s[i], true
	case []float64:
		if i < 0 {
			i = len(s) + i
		}
		if i < 0 || i >= len(s) {
			return nil, false
		}
		return s[i], true
	case []bool:
		if i < 0 {
			i = len(s) + i
		}
		if i < 0 || i >= len(s) {
			return nil, false
		}
		return s[i], true
	case []any:
		if i < 0 {
			i = len(s) + i
		}
		if i < 0 || i >= len(s) {
			return nil, false
		}
		return s[i], true
	}

	// Reflection fallback for other slice types
	val := reflect.ValueOf(collection)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, false
	}
	length := val.Len()
	if i < 0 {
		i = length + i
	}
	if i < 0 || i >= length {
		return nil, false
	}
	return val.Index(i).Interface(), true
}

// ShiftLeft implements the << operator for both slices and channels.
// For slices: appends the value and returns the new slice.
// For channels: sends the value and returns the channel.
// This allows chaining: arr << a << b or ch << val.
// Ruby: arr << item, ch << val
func ShiftLeft(collection any, value any) any {
	// Fast paths for common slice types
	switch s := collection.(type) {
	case []int:
		if v, ok := value.(int); ok {
			return append(s, v)
		}
	case []string:
		if v, ok := value.(string); ok {
			return append(s, v)
		}
	case []float64:
		if v, ok := value.(float64); ok {
			return append(s, v)
		}
	case []bool:
		if v, ok := value.(bool); ok {
			return append(s, v)
		}
	case []any:
		return append(s, value)
	}

	// Use reflection for channels and other types
	val := reflect.ValueOf(collection)

	switch val.Kind() {
	case reflect.Chan:
		// Channel send
		val.Send(reflect.ValueOf(value))
		return collection

	case reflect.Slice:
		// Slice append via reflection
		return appendToSlice(val, value)

	default:
		panic(fmt.Sprintf("ShiftLeft: expected slice or channel, got %T", collection))
	}
}

// appendToSlice appends a value to a slice using reflection.
func appendToSlice(sliceVal reflect.Value, value any) any {
	// Convert value to the slice's element type
	elemType := sliceVal.Type().Elem()
	valueVal := reflect.ValueOf(value)

	// Handle type conversion if needed
	if valueVal.Type() != elemType && valueVal.Type().ConvertibleTo(elemType) {
		valueVal = valueVal.Convert(elemType)
	}

	// Append and return the new slice
	newSlice := reflect.Append(sliceVal, valueVal)
	return newSlice.Interface()
}
