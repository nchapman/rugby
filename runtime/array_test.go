package runtime

import (
	"reflect"
	"testing"
)

func TestSelect(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6}
	evens := Select(nums, func(n int) (bool, bool) { return n%2 == 0, true })
	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(evens, expected) {
		t.Errorf("Select evens: got %v, want %v", evens, expected)
	}

	// Empty result
	none := Select(nums, func(n int) (bool, bool) { return n > 10, true })
	if len(none) != 0 {
		t.Errorf("Select none: got %v, want empty", none)
	}
}

func TestReject(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6}
	odds := Reject(nums, func(n int) (bool, bool) { return n%2 == 0, true })
	expected := []int{1, 3, 5}
	if !reflect.DeepEqual(odds, expected) {
		t.Errorf("Reject evens: got %v, want %v", odds, expected)
	}
}

func TestMap(t *testing.T) {
	nums := []int{1, 2, 3}
	doubled := Map(nums, func(n int) (int, bool, bool) { return n * 2, true, true })
	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(doubled, expected) {
		t.Errorf("Map double: got %v, want %v", doubled, expected)
	}

	// Type conversion
	strs := Map(nums, func(n int) (string, bool, bool) {
		return string(rune('a' + n - 1)), true, true
	})
	expectedStrs := []string{"a", "b", "c"}
	if !reflect.DeepEqual(strs, expectedStrs) {
		t.Errorf("Map to strings: got %v, want %v", strs, expectedStrs)
	}
}

func TestReduce(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	sum := Reduce(nums, 0, func(acc, n int) (int, bool) { return acc + n, true })
	if sum != 15 {
		t.Errorf("Reduce sum: got %d, want 15", sum)
	}

	// With initial value
	product := Reduce(nums, 1, func(acc, n int) (int, bool) { return acc * n, true })
	if product != 120 {
		t.Errorf("Reduce product: got %d, want 120", product)
	}

	// Empty slice
	emptySum := Reduce([]int{}, 0, func(acc, n int) (int, bool) { return acc + n, true })
	if emptySum != 0 {
		t.Errorf("Reduce empty: got %d, want 0", emptySum)
	}
}

func TestFind(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	// Found
	val, ok := Find(nums, func(n int) (bool, bool) { return n > 3, true })
	if !ok || val != 4 {
		t.Errorf("Find >3: got (%d, %v), want (4, true)", val, ok)
	}

	// Not found
	val, ok = Find(nums, func(n int) (bool, bool) { return n > 10, true })
	if ok {
		t.Errorf("Find >10: got (%d, %v), want (0, false)", val, ok)
	}
}

func TestAny(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	if !Any(nums, func(n int) (bool, bool) { return n > 3, true }) {
		t.Error("Any >3: got false, want true")
	}

	if Any(nums, func(n int) (bool, bool) { return n > 10, true }) {
		t.Error("Any >10: got true, want false")
	}
}

func TestAll(t *testing.T) {
	nums := []int{2, 4, 6, 8}

	if !All(nums, func(n int) (bool, bool) { return n%2 == 0, true }) {
		t.Error("All even: got false, want true")
	}

	if All(nums, func(n int) (bool, bool) { return n > 5, true }) {
		t.Error("All >5: got true, want false")
	}

	// Empty slice returns true (vacuous truth)
	if !All([]int{}, func(n int) (bool, bool) { return false, true }) {
		t.Error("All empty: got false, want true")
	}
}

func TestNone(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	if !None(nums, func(n int) (bool, bool) { return n > 10, true }) {
		t.Error("None >10: got false, want true")
	}

	if None(nums, func(n int) (bool, bool) { return n > 3, true }) {
		t.Error("None >3: got true, want false")
	}
}

func TestContains(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	if !Contains(nums, 3) {
		t.Error("Contains 3: got false, want true")
	}

	if Contains(nums, 10) {
		t.Error("Contains 10: got true, want false")
	}

	// Strings
	strs := []string{"a", "b", "c"}
	if !Contains(strs, "b") {
		t.Error("Contains 'b': got false, want true")
	}
}

func TestFirst(t *testing.T) {
	nums := []int{1, 2, 3}

	val, ok := First(nums)
	if !ok || val != 1 {
		t.Errorf("First: got (%d, %v), want (1, true)", val, ok)
	}

	val, ok = First([]int{})
	if ok {
		t.Errorf("First empty: got (%d, %v), want (0, false)", val, ok)
	}
}

func TestLast(t *testing.T) {
	nums := []int{1, 2, 3}

	val, ok := Last(nums)
	if !ok || val != 3 {
		t.Errorf("Last: got (%d, %v), want (3, true)", val, ok)
	}

	val, ok = Last([]int{})
	if ok {
		t.Errorf("Last empty: got (%d, %v), want (0, false)", val, ok)
	}
}

func TestReverse(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	Reverse(nums)
	expected := []int{5, 4, 3, 2, 1}
	if !reflect.DeepEqual(nums, expected) {
		t.Errorf("Reverse: got %v, want %v", nums, expected)
	}

	// Even length
	even := []int{1, 2, 3, 4}
	Reverse(even)
	expectedEven := []int{4, 3, 2, 1}
	if !reflect.DeepEqual(even, expectedEven) {
		t.Errorf("Reverse even: got %v, want %v", even, expectedEven)
	}
}

func TestReversed(t *testing.T) {
	nums := []int{1, 2, 3}
	reversed := Reversed(nums)

	// Original unchanged
	if !reflect.DeepEqual(nums, []int{1, 2, 3}) {
		t.Errorf("Reversed modified original: %v", nums)
	}

	expected := []int{3, 2, 1}
	if !reflect.DeepEqual(reversed, expected) {
		t.Errorf("Reversed: got %v, want %v", reversed, expected)
	}
}

func TestSumInt(t *testing.T) {
	if sum := SumInt([]int{1, 2, 3, 4, 5}); sum != 15 {
		t.Errorf("SumInt: got %d, want 15", sum)
	}

	if sum := SumInt([]int{}); sum != 0 {
		t.Errorf("SumInt empty: got %d, want 0", sum)
	}
}

func TestSumFloat(t *testing.T) {
	if sum := SumFloat([]float64{1.5, 2.5, 3.0}); sum != 7.0 {
		t.Errorf("SumFloat: got %f, want 7.0", sum)
	}
}

func TestMinMaxInt(t *testing.T) {
	nums := []int{3, 1, 4, 1, 5, 9, 2, 6}

	min, ok := MinInt(nums)
	if !ok || min != 1 {
		t.Errorf("MinInt: got (%d, %v), want (1, true)", min, ok)
	}

	max, ok := MaxInt(nums)
	if !ok || max != 9 {
		t.Errorf("MaxInt: got (%d, %v), want (9, true)", max, ok)
	}

	// Empty
	_, ok = MinInt([]int{})
	if ok {
		t.Error("MinInt empty: got true, want false")
	}
}

func TestMinMaxFloat(t *testing.T) {
	nums := []float64{3.14, 1.0, 2.71, 0.5}

	min, ok := MinFloat(nums)
	if !ok || min != 0.5 {
		t.Errorf("MinFloat: got (%f, %v), want (0.5, true)", min, ok)
	}

	max, ok := MaxFloat(nums)
	if !ok || max != 3.14 {
		t.Errorf("MaxFloat: got (%f, %v), want (3.14, true)", max, ok)
	}
}
