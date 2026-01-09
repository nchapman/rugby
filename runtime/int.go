package runtime

// Times iterates n times, calling the function with the current index (0 to n-1).
// Ruby: 5.times { |i| puts i }
func Times(n int, fn func(int)) {
	for i := 0; i < n; i++ {
		fn(i)
	}
}

// Upto iterates from start to end (inclusive), calling the function with each value.
// Ruby: 1.upto(5) { |i| puts i }
func Upto(start, end int, fn func(int)) {
	for i := start; i <= end; i++ {
		fn(i)
	}
}

// Downto iterates from start down to end (inclusive), calling the function with each value.
// Ruby: 5.downto(1) { |i| puts i }
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
