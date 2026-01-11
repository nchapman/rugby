package runtime

// Times iterates n times, calling the function with the current index (0 to n-1).
// The callback returns false to break, true to continue.
func Times(n int, fn func(int) bool) {
	for i := range n {
		if !fn(i) {
			break
		}
	}
}

// Upto iterates from start to end (inclusive), calling the function with each value.
// The callback returns false to break, true to continue.
func Upto(start, end int, fn func(int) bool) {
	for i := start; i <= end; i++ {
		if !fn(i) {
			break
		}
	}
}

// Downto iterates from start down to end (inclusive), calling the function with each value.
// The callback returns false to break, true to continue.
func Downto(start, end int, fn func(int) bool) {
	for i := start; i >= end; i-- {
		if !fn(i) {
			break
		}
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
