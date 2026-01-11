package runtime

// Range represents a range of integers (start..end or start...end)
type Range struct {
	Start, End int
	Exclusive  bool // true for ... (exclusive), false for .. (inclusive)
}

// RangeEach iterates over each value in the range
func RangeEach(r Range, fn func(int) bool) {
	if r.Exclusive {
		for i := r.Start; i < r.End; i++ {
			if !fn(i) {
				break
			}
		}
	} else {
		for i := r.Start; i <= r.End; i++ {
			if !fn(i) {
				break
			}
		}
	}
}

// RangeContains checks if a value is within the range
func RangeContains(r Range, n int) bool {
	if r.Exclusive {
		return n >= r.Start && n < r.End
	}
	return n >= r.Start && n <= r.End
}

// RangeToArray materializes the range into a slice of ints
func RangeToArray(r Range) []int {
	size := RangeSize(r)
	if size <= 0 {
		return []int{}
	}
	result := make([]int, size)
	for i := range size {
		result[i] = r.Start + i
	}
	return result
}

// RangeSize returns the number of elements in the range
func RangeSize(r Range) int {
	if r.Exclusive {
		if r.End <= r.Start {
			return 0
		}
		return r.End - r.Start
	}
	if r.End < r.Start {
		return 0
	}
	return r.End - r.Start + 1
}
