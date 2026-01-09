// Package runtime provides Ruby-like stdlib ergonomics for Rugby programs.
package runtime

// Select returns elements for which the predicate returns true.
// Ruby: arr.select { |x| x > 5 }
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
// Ruby: arr.reject { |x| x > 5 }
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
// Ruby: arr.map { |x| x * 2 }
func Map[T, R any](slice []T, mapper func(T) R) []R {
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = mapper(v)
	}
	return result
}

// Reduce folds the slice into a single value using the reducer function.
// Ruby: arr.reduce(0) { |acc, x| acc + x }
func Reduce[T, R any](slice []T, initial R, reducer func(R, T) R) R {
	acc := initial
	for _, v := range slice {
		acc = reducer(acc, v)
	}
	return acc
}

// Find returns the first element matching the predicate.
// Ruby: arr.find { |x| x > 5 }
func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, v := range slice {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// Any returns true if any element matches the predicate.
// Ruby: arr.any? { |x| x > 5 }
func Any[T any](slice []T, predicate func(T) bool) bool {
	for _, v := range slice {
		if predicate(v) {
			return true
		}
	}
	return false
}

// All returns true if all elements match the predicate.
// Ruby: arr.all? { |x| x > 0 }
func All[T any](slice []T, predicate func(T) bool) bool {
	for _, v := range slice {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// None returns true if no elements match the predicate.
// Ruby: arr.none? { |x| x < 0 }
func None[T any](slice []T, predicate func(T) bool) bool {
	for _, v := range slice {
		if predicate(v) {
			return false
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
