package runtime

import "fmt"

// Times iterates n times, calling the function with the current index (0 to n-1).
func Times(n int, fn func(int)) {
	for i := range n {
		fn(i)
	}
}

// Upto iterates from start to end (inclusive), calling the function with each value.
func Upto(start, end int, fn func(int)) {
	for i := start; i <= end; i++ {
		fn(i)
	}
}

// Downto iterates from start down to end (inclusive), calling the function with each value.
func Downto(start, end int, fn func(int)) {
	for i := start; i >= end; i-- {
		fn(i)
	}
}

// Abs returns the absolute value of an integer.
// Ruby: n.abs
// Note: Abs(math.MinInt) overflows and returns math.MinInt.
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Clamp constrains a value to a range [min, max].
// Ruby: n.clamp(min, max)
// Precondition: min <= max. Behavior is undefined if min > max.
func Clamp(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// AbsInt64 returns the absolute value of an int64.
// Note: AbsInt64(math.MinInt64) overflows and returns math.MinInt64.
func AbsInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// ClampInt64 constrains an int64 value to a range [min, max].
// Precondition: min <= max. Behavior is undefined if min > max.
func ClampInt64(n, min, max int64) int64 {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// Even returns true if n is even.
// Ruby: n.even?
func Even(n int) bool {
	return n%2 == 0
}

// Odd returns true if n is odd.
// Ruby: n.odd?
func Odd(n int) bool {
	return n%2 != 0
}

// Positive returns true if n is positive (> 0).
// Ruby: n.positive?
func Positive(n int) bool {
	return n > 0
}

// Negative returns true if n is negative (< 0).
// Ruby: n.negative?
func Negative(n int) bool {
	return n < 0
}

// Zero returns true if n is zero.
// Ruby: n.zero?
func Zero(n int) bool {
	return n == 0
}

// IntToString converts an int to its string representation.
// Ruby: n.to_s
func IntToString(n int) string {
	return fmt.Sprintf("%d", n)
}

// IntToFloat converts an int to a float64.
// Ruby: n.to_f
func IntToFloat(n int) float64 {
	return float64(n)
}

// ShiftRight performs bitwise right shift (>>).
// Works with int and any types for compatibility with dynamic typing.
func ShiftRight(left, right any) any {
	switch l := left.(type) {
	case int:
		r := toIntForShift(right)
		return l >> r
	case int64:
		r := toIntForShift(right)
		return l >> r
	case uint:
		r := toIntForShift(right)
		return l >> r
	case uint64:
		r := toIntForShift(right)
		return l >> r
	case byte: // uint8
		r := toIntForShift(right)
		return int(l) >> r
	default:
		panic(fmt.Sprintf("ShiftRight: unsupported type %T", left))
	}
}
