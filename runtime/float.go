package runtime

import (
	"fmt"
	"math"
)

// Truncate returns the integer part of n, truncating toward zero.
// Ruby: f.truncate or f.to_i
func Truncate(n float64) int {
	return int(math.Trunc(n))
}

// FloatAbs returns the absolute value of a float64.
// Ruby: f.abs
func FloatAbs(n float64) float64 {
	return math.Abs(n)
}

// FloatToString converts a float64 to its string representation.
// Ruby: f.to_s
func FloatToString(n float64) string {
	return fmt.Sprintf("%v", n)
}

// FloatToInt converts a float64 to an int by truncation.
// Ruby: f.to_i
func FloatToInt(n float64) int {
	return int(n)
}

// FloatPositive returns true if n is positive (> 0).
// Ruby: f.positive?
func FloatPositive(n float64) bool {
	return n > 0
}

// FloatNegative returns true if n is negative (< 0).
// Ruby: f.negative?
func FloatNegative(n float64) bool {
	return n < 0
}

// FloatZero returns true if n is zero.
// Ruby: f.zero?
func FloatZero(n float64) bool {
	return n == 0
}
