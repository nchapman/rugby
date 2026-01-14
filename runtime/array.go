// Package runtime provides Ruby-like stdlib ergonomics for Rugby programs.
package runtime

import (
	"cmp"
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"unicode"
)

// Each iterates over a slice, calling the function for each element.
// Uses any with type switches to support both typed and untyped slices.
func Each(slice any, fn func(any)) {
	switch s := slice.(type) {
	case []any:
		for _, v := range s {
			fn(v)
		}
	case []int:
		for _, v := range s {
			fn(v)
		}
	case []string:
		for _, v := range s {
			fn(v)
		}
	case []float64:
		for _, v := range s {
			fn(v)
		}
	case []bool:
		for _, v := range s {
			fn(v)
		}
	default:
		// Use reflection as fallback for other slice types
		val := reflect.ValueOf(slice)
		if val.Kind() != reflect.Slice {
			panic(fmt.Sprintf("Each: expected slice, got %T", slice))
		}
		for i := range val.Len() {
			fn(val.Index(i).Interface())
		}
	}
}

// EachWithIndex iterates over a slice with index, calling the function for each element.
// Uses any with type switches to support both typed and untyped slices.
func EachWithIndex(slice any, fn func(any, int)) {
	switch s := slice.(type) {
	case []any:
		for i, v := range s {
			fn(v, i)
		}
	case []int:
		for i, v := range s {
			fn(v, i)
		}
	case []string:
		for i, v := range s {
			fn(v, i)
		}
	case []float64:
		for i, v := range s {
			fn(v, i)
		}
	case []bool:
		for i, v := range s {
			fn(v, i)
		}
	default:
		// Use reflection as fallback for other slice types
		val := reflect.ValueOf(slice)
		if val.Kind() != reflect.Slice {
			panic(fmt.Sprintf("EachWithIndex: expected slice, got %T", slice))
		}
		for i := range val.Len() {
			fn(val.Index(i).Interface(), i)
		}
	}
}

// Select returns elements for which the predicate returns true.
func Select[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reject returns elements for which the predicate returns false.
func Reject[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, v := range slice {
		if !predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Map transforms each element using the mapper function.
func Map[T, R any](slice []T, mapper func(T) R) []R {
	result := make([]R, 0, len(slice))
	for _, v := range slice {
		result = append(result, mapper(v))
	}
	return result
}

// Reduce folds the slice into a single value using the reducer function.
func Reduce[T, R any](slice []T, initial R, reducer func(R, T) R) R {
	acc := initial
	for _, v := range slice {
		acc = reducer(acc, v)
	}
	return acc
}

// Find returns the first element matching the predicate.
func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, v := range slice {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// FindPtr returns a pointer to the first element matching the predicate, or nil.
// Used for optional coalescing in Rugby.
func FindPtr[T any](slice []T, predicate func(T) bool) *T {
	for i, v := range slice {
		if predicate(v) {
			return &slice[i]
		}
	}
	return nil
}

// Any returns true if any element matches the predicate.
func Any[T any](slice []T, predicate func(T) bool) bool {
	return slices.ContainsFunc(slice, predicate)
}

// All returns true if all elements match the predicate.
func All[T any](slice []T, predicate func(T) bool) bool {
	for _, v := range slice {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// None returns true if no elements match the predicate.
func None[T any](slice []T, predicate func(T) bool) bool {
	return !slices.ContainsFunc(slice, predicate)
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

// MinIntPtr returns a pointer to the minimum value, or nil if empty.
// Used for optional coalescing in Rugby.
func MinIntPtr(slice []int) *int {
	if len(slice) == 0 {
		return nil
	}
	v := slices.Min(slice)
	return &v
}

// MaxIntPtr returns a pointer to the maximum value, or nil if empty.
// Used for optional coalescing in Rugby.
func MaxIntPtr(slice []int) *int {
	if len(slice) == 0 {
		return nil
	}
	v := slices.Max(slice)
	return &v
}

// MinFloatPtr returns a pointer to the minimum value, or nil if empty.
// Used for optional coalescing in Rugby.
func MinFloatPtr(slice []float64) *float64 {
	if len(slice) == 0 {
		return nil
	}
	v := slices.Min(slice)
	return &v
}

// MaxFloatPtr returns a pointer to the maximum value, or nil if empty.
// Used for optional coalescing in Rugby.
func MaxFloatPtr(slice []float64) *float64 {
	if len(slice) == 0 {
		return nil
	}
	v := slices.Max(slice)
	return &v
}

// FirstPtr returns a pointer to the first element, or nil if empty.
// Used for optional coalescing in Rugby.
func FirstPtr[T any](slice []T) *T {
	if len(slice) == 0 {
		return nil
	}
	return &slice[0]
}

// LastPtr returns a pointer to the last element, or nil if empty.
// Used for optional coalescing in Rugby.
func LastPtr[T any](slice []T) *T {
	if len(slice) == 0 {
		return nil
	}
	return &slice[len(slice)-1]
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

// Shift removes and returns the first element of the slice.
// Modifies the original slice in place via pointer.
// Ruby: arr.shift
func Shift[T any](slice *[]T) (T, bool) {
	if len(*slice) == 0 {
		var zero T
		return zero, false
	}
	first := (*slice)[0]
	*slice = (*slice)[1:]
	return first, true
}

// ShiftInt removes and returns the first element of an int slice.
func ShiftInt(slice *[]int) int {
	if len(*slice) == 0 {
		return 0
	}
	first := (*slice)[0]
	*slice = (*slice)[1:]
	return first
}

// ShiftString removes and returns the first element of a string slice.
func ShiftString(slice *[]string) string {
	if len(*slice) == 0 {
		return ""
	}
	first := (*slice)[0]
	*slice = (*slice)[1:]
	return first
}

// ShiftFloat removes and returns the first element of a float64 slice.
func ShiftFloat(slice *[]float64) float64 {
	if len(*slice) == 0 {
		return 0
	}
	first := (*slice)[0]
	*slice = (*slice)[1:]
	return first
}

// ShiftAny removes and returns the first element of an any slice.
func ShiftAny(slice *[]any) any {
	if len(*slice) == 0 {
		return nil
	}
	first := (*slice)[0]
	*slice = (*slice)[1:]
	return first
}

// Pop removes and returns the last element of the slice.
// Modifies the original slice in place via pointer.
// Ruby: arr.pop
func Pop[T any](slice *[]T) (T, bool) {
	if len(*slice) == 0 {
		var zero T
		return zero, false
	}
	last := (*slice)[len(*slice)-1]
	*slice = (*slice)[:len(*slice)-1]
	return last, true
}

// PopInt removes and returns the last element of an int slice.
func PopInt(slice *[]int) int {
	if len(*slice) == 0 {
		return 0
	}
	last := (*slice)[len(*slice)-1]
	*slice = (*slice)[:len(*slice)-1]
	return last
}

// PopString removes and returns the last element of a string slice.
func PopString(slice *[]string) string {
	if len(*slice) == 0 {
		return ""
	}
	last := (*slice)[len(*slice)-1]
	*slice = (*slice)[:len(*slice)-1]
	return last
}

// PopFloat removes and returns the last element of a float64 slice.
func PopFloat(slice *[]float64) float64 {
	if len(*slice) == 0 {
		return 0
	}
	last := (*slice)[len(*slice)-1]
	*slice = (*slice)[:len(*slice)-1]
	return last
}

// PopAny removes and returns the last element of an any slice.
func PopAny(slice *[]any) any {
	if len(*slice) == 0 {
		return nil
	}
	last := (*slice)[len(*slice)-1]
	*slice = (*slice)[:len(*slice)-1]
	return last
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

// Slice extracts a portion of a slice or string using a Range.
// Supports negative indices (Ruby-style: -1 is last element).
// For inclusive range (..): includes both start and end
// For exclusive range (...): includes start but not end
func Slice(collection any, r Range) any {
	switch c := collection.(type) {
	case []int:
		return sliceGeneric(c, r)
	case []string:
		return sliceGeneric(c, r)
	case []float64:
		return sliceGeneric(c, r)
	case []bool:
		return sliceGeneric(c, r)
	case []any:
		return sliceGeneric(c, r)
	case string:
		return sliceString(c, r)
	default:
		// Reflection fallback for other slice types
		val := reflect.ValueOf(collection)
		if val.Kind() == reflect.Slice {
			length := val.Len()
			start, end := normalizeRangeIndices(r.Start, r.End, length, r.Exclusive)
			if start >= length || start > end || start < 0 {
				// Return empty slice of same type
				return reflect.MakeSlice(val.Type(), 0, 0).Interface()
			}
			return val.Slice(start, end).Interface()
		}
		panic(fmt.Sprintf("Slice: expected slice or string, got %T", collection))
	}
}

// sliceGeneric extracts a portion of a slice with negative index support.
func sliceGeneric[T any](slice []T, r Range) []T {
	length := len(slice)
	start, end := normalizeRangeIndices(r.Start, r.End, length, r.Exclusive)
	if start >= length || start > end || start < 0 {
		return []T{}
	}
	return slice[start:end]
}

// sliceString extracts a portion of a string with negative index support.
func sliceString(s string, r Range) string {
	runes := []rune(s)
	length := len(runes)
	start, end := normalizeRangeIndices(r.Start, r.End, length, r.Exclusive)
	if start >= length || start > end || start < 0 {
		return ""
	}
	return string(runes[start:end])
}

// normalizeRangeIndices converts Ruby-style indices to Go slice indices.
// Returns start and end for Go slicing (end is exclusive in Go).
func normalizeRangeIndices(start, end, length int, exclusive bool) (int, int) {
	// Handle negative indices
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}

	// For inclusive range, end should include the element
	// Go slices are exclusive, so we add 1 to include the end element
	if !exclusive {
		end = end + 1
	}

	// Clamp end to length
	if end > length {
		end = length
	}

	return start, end
}

// CallMethod calls a method on an object by name using reflection.
// Used by symbol-to-proc (&:method) syntax.
// Returns the method's return value, or panics if the method doesn't exist.
func CallMethod(obj any, methodName string) any {
	// Handle nil receiver
	if obj == nil {
		panic(fmt.Sprintf("CallMethod: cannot call %q on nil", methodName))
	}

	// Handle runtime string methods (Go strings don't have these as methods)
	if s, ok := obj.(string); ok {
		switch methodName {
		case "upcase":
			return Upcase(s)
		case "downcase":
			return Downcase(s)
		case "capitalize":
			return Capitalize(s)
		case "strip":
			return Strip(s)
		case "lstrip":
			return Lstrip(s)
		case "rstrip":
			return Rstrip(s)
		case "reverse":
			return StringReverse(s)
		case "chars":
			return Chars(s)
		case "length":
			return CharLength(s)
		case "empty?":
			return len(s) == 0
		case "present?":
			return len(s) > 0
		}
	}

	// Handle runtime int methods
	if n, ok := obj.(int); ok {
		switch methodName {
		case "even?":
			return Even(n)
		case "odd?":
			return Odd(n)
		case "abs":
			return Abs(n)
		}
	}

	val := reflect.ValueOf(obj)

	// Try to find the method on the value
	method := val.MethodByName(toGoMethodName(methodName))
	if !method.IsValid() {
		// Try pointer receiver
		if val.Kind() != reflect.Ptr && val.CanAddr() {
			method = val.Addr().MethodByName(toGoMethodName(methodName))
		}
	}

	if !method.IsValid() {
		panic(fmt.Sprintf("CallMethod: method %q not found on %T", methodName, obj))
	}

	// Call the method with no arguments
	results := method.Call(nil)

	// Return the first result, or nil if no return values
	if len(results) == 0 {
		return nil
	}
	return results[0].Interface()
}

// toGoMethodName converts a Ruby-style method name to Go-style.
// Examples: upcase -> Upcase, to_s -> ToS, empty? -> Empty_PRED
func toGoMethodName(name string) string {
	if len(name) == 0 {
		return ""
	}

	// Handle predicate methods (ending in ?)
	isPredicate := false
	if name[len(name)-1] == '?' {
		isPredicate = true
		name = name[:len(name)-1]
	}

	// Convert snake_case to PascalCase
	var result strings.Builder
	capitalizeNext := true
	for _, r := range name {
		if r == '_' {
			capitalizeNext = true
		} else if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}
	}

	if isPredicate {
		result.WriteString("_PRED")
	}

	return result.String()
}
