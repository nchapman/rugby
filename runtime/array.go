// Package runtime provides Ruby-like stdlib ergonomics for Rugby programs.
package runtime

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/exp/constraints"
)

// Each iterates over a slice, calling the function for each element.
// The callback returns false to break, true to continue.
// Uses interface{} with type switches to support both typed and untyped slices.
func Each(slice interface{}, fn func(interface{}) bool) {
	switch s := slice.(type) {
	case []interface{}:
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
		for i := 0; i < val.Len(); i++ {
			if !fn(val.Index(i).Interface()) {
				break
			}
		}
	}
}

// EachWithIndex iterates over a slice with index, calling the function for each element.
// The callback returns false to break, true to continue.
// Uses interface{} with type switches to support both typed and untyped slices.
func EachWithIndex(slice interface{}, fn func(interface{}, int) bool) {
	switch s := slice.(type) {
	case []interface{}:
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
		for i := 0; i < val.Len(); i++ {
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
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
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

// Reverse reverses the slice in-place.
// Ruby: arr.reverse!
func Reverse[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// Reversed returns a new slice with elements in reverse order.
// Ruby: arr.reverse
func Reversed[T any](slice []T) []T {
	result := make([]T, len(slice))
	for i, v := range slice {
		result[len(slice)-1-i] = v
	}
	return result
}

// Sum returns the sum of all elements (for int slices).
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
	min := slice[0]
	for _, v := range slice[1:] {
		if v < min {
			min = v
		}
	}
	return min, true
}

// MaxInt returns the maximum value in an int slice.
func MaxInt(slice []int) (int, bool) {
	if len(slice) == 0 {
		return 0, false
	}
	max := slice[0]
	for _, v := range slice[1:] {
		if v > max {
			max = v
		}
	}
	return max, true
}

// MinFloat returns the minimum value in a float64 slice.
// NaN values use IEEE 754 semantics (NaN comparisons are always false).
func MinFloat(slice []float64) (float64, bool) {
	if len(slice) == 0 {
		return 0, false
	}
	min := slice[0]
	for _, v := range slice[1:] {
		if v < min {
			min = v
		}
	}
	return min, true
}

// MaxFloat returns the maximum value in a float64 slice.
// NaN values use IEEE 754 semantics (NaN comparisons are always false).
func MaxFloat(slice []float64) (float64, bool) {
	if len(slice) == 0 {
		return 0, false
	}
	max := slice[0]
	for _, v := range slice[1:] {
		if v > max {
			max = v
		}
	}
	return max, true
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
// For now, let's just do a simple flattening if it's []interface{}.
// Typed slices []int cannot be flattened further.
// Ruby: arr.flatten
func Flatten(slice interface{}) []interface{} {
	result := make([]interface{}, 0)
	val := reflect.ValueOf(slice)
	if val.Kind() != reflect.Slice {
		return []interface{}{slice}
	}

	for i := 0; i < val.Len(); i++ {
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
func Sort[T constraints.Ordered](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
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
		// Return copy? Or slice? Slice is fine.
		// Ruby returns new array.
		res := make([]T, len(slice))
		copy(res, slice)
		return res
	}
	res := make([]T, n)
	copy(res, slice[:n])
	return res
}

// LastN returns the last n elements.
func LastN[T any](slice []T, n int) []T {
	if n <= 0 {
		return []T{}
	}
	l := len(slice)
	if n >= l {
		res := make([]T, l)
		copy(res, slice)
		return res
	}
	res := make([]T, n)
	copy(res, slice[l-n:])
	return res
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
		res := make([]T, l)
		copy(res, slice)
		return res
	}
	res := make([]T, 0, l)
	res = append(res, slice[n:]...)
	res = append(res, slice[:n]...)
	return res
}
