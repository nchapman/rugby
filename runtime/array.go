// Package runtime provides Ruby-like stdlib ergonomics for Rugby programs.
package runtime

// Each iterates over a slice, calling the function for each element.
// The callback returns false to break, true to continue.
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
	}
}

// EachWithIndex iterates over a slice with index, calling the function for each element.
// The callback returns false to break, true to continue.
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
