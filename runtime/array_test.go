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

func TestEach(t *testing.T) {
	// Test with []int
	var collected []int
	Each([]int{1, 2, 3}, func(v any) bool {
		collected = append(collected, v.(int))
		return true
	})
	if !reflect.DeepEqual(collected, []int{1, 2, 3}) {
		t.Errorf("Each []int: got %v, want [1 2 3]", collected)
	}

	// Test with []string
	var strs []string
	Each([]string{"a", "b", "c"}, func(v any) bool {
		strs = append(strs, v.(string))
		return true
	})
	if !reflect.DeepEqual(strs, []string{"a", "b", "c"}) {
		t.Errorf("Each []string: got %v, want [a b c]", strs)
	}

	// Test early break
	var partial []int
	Each([]int{1, 2, 3, 4, 5}, func(v any) bool {
		partial = append(partial, v.(int))
		return v.(int) < 3
	})
	if !reflect.DeepEqual(partial, []int{1, 2, 3}) {
		t.Errorf("Each with break: got %v, want [1 2 3]", partial)
	}
}

func TestEachWithIndex(t *testing.T) {
	var pairs [][2]int
	EachWithIndex([]int{10, 20, 30}, func(v any, i int) bool {
		pairs = append(pairs, [2]int{v.(int), i})
		return true
	})
	expected := [][2]int{{10, 0}, {20, 1}, {30, 2}}
	if !reflect.DeepEqual(pairs, expected) {
		t.Errorf("EachWithIndex: got %v, want %v", pairs, expected)
	}
}

func TestJoin(t *testing.T) {
	// Join strings
	result := Join([]string{"a", "b", "c"}, ", ")
	if result != "a, b, c" {
		t.Errorf("Join strings: got %q, want %q", result, "a, b, c")
	}

	// Join ints
	result = Join([]int{1, 2, 3}, "-")
	if result != "1-2-3" {
		t.Errorf("Join ints: got %q, want %q", result, "1-2-3")
	}

	// Empty slice
	result = Join([]string{}, ", ")
	if result != "" {
		t.Errorf("Join empty: got %q, want empty", result)
	}

	// Single element
	result = Join([]string{"only"}, ", ")
	if result != "only" {
		t.Errorf("Join single: got %q, want %q", result, "only")
	}
}

func TestFlatten(t *testing.T) {
	// Simple nested slice
	nested := []any{1, []any{2, 3}, []any{4, []any{5, 6}}}
	result := Flatten(nested)
	expected := []any{1, 2, 3, 4, 5, 6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Flatten: got %v, want %v", result, expected)
	}

	// Already flat
	flat := []any{1, 2, 3}
	result = Flatten(flat)
	if !reflect.DeepEqual(result, []any{1, 2, 3}) {
		t.Errorf("Flatten flat: got %v, want [1 2 3]", result)
	}

	// Non-slice returns wrapped
	result = Flatten(42)
	if !reflect.DeepEqual(result, []any{42}) {
		t.Errorf("Flatten non-slice: got %v, want [42]", result)
	}
}

func TestUniq(t *testing.T) {
	// With duplicates
	result := Uniq([]int{1, 2, 2, 3, 1, 4, 3})
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Uniq: got %v, want %v", result, expected)
	}

	// Already unique
	result = Uniq([]int{1, 2, 3})
	if !reflect.DeepEqual(result, []int{1, 2, 3}) {
		t.Errorf("Uniq already unique: got %v, want [1 2 3]", result)
	}

	// Empty
	result = Uniq([]int{})
	if len(result) != 0 {
		t.Errorf("Uniq empty: got %v, want empty", result)
	}

	// Strings
	strResult := Uniq([]string{"a", "b", "a", "c", "b"})
	if !reflect.DeepEqual(strResult, []string{"a", "b", "c"}) {
		t.Errorf("Uniq strings: got %v, want [a b c]", strResult)
	}
}

func TestSort(t *testing.T) {
	nums := []int{3, 1, 4, 1, 5, 9, 2, 6}
	sorted := Sort(nums)

	// Original unchanged
	if !reflect.DeepEqual(nums, []int{3, 1, 4, 1, 5, 9, 2, 6}) {
		t.Errorf("Sort modified original: %v", nums)
	}

	expected := []int{1, 1, 2, 3, 4, 5, 6, 9}
	if !reflect.DeepEqual(sorted, expected) {
		t.Errorf("Sort: got %v, want %v", sorted, expected)
	}

	// Strings
	strs := []string{"banana", "apple", "cherry"}
	sortedStrs := Sort(strs)
	if !reflect.DeepEqual(sortedStrs, []string{"apple", "banana", "cherry"}) {
		t.Errorf("Sort strings: got %v, want [apple banana cherry]", sortedStrs)
	}
}

func TestShuffle(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	shuffled := Shuffle(nums)

	// Original unchanged
	if !reflect.DeepEqual(nums, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
		t.Errorf("Shuffle modified original: %v", nums)
	}

	// Same length
	if len(shuffled) != len(nums) {
		t.Errorf("Shuffle length: got %d, want %d", len(shuffled), len(nums))
	}

	// Contains same elements (sorted should match)
	sortedShuffled := Sort(shuffled)
	if !reflect.DeepEqual(sortedShuffled, nums) {
		t.Errorf("Shuffle elements changed: %v", shuffled)
	}
}

func TestSample(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	// Get a sample (just test it returns a valid element)
	val, ok := Sample(nums)
	if !ok {
		t.Error("Sample: got false, want true")
	}
	if !Contains(nums, val) {
		t.Errorf("Sample: %d not in %v", val, nums)
	}

	// Empty slice
	_, ok = Sample([]int{})
	if ok {
		t.Error("Sample empty: got true, want false")
	}
}

func TestFirstN(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	result := FirstN(nums, 3)
	if !reflect.DeepEqual(result, []int{1, 2, 3}) {
		t.Errorf("FirstN(3): got %v, want [1 2 3]", result)
	}

	// More than length
	result = FirstN(nums, 10)
	if !reflect.DeepEqual(result, []int{1, 2, 3, 4, 5}) {
		t.Errorf("FirstN(10): got %v, want [1 2 3 4 5]", result)
	}

	// Zero
	result = FirstN(nums, 0)
	if len(result) != 0 {
		t.Errorf("FirstN(0): got %v, want empty", result)
	}

	// Negative
	result = FirstN(nums, -1)
	if len(result) != 0 {
		t.Errorf("FirstN(-1): got %v, want empty", result)
	}
}

func TestLastN(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	result := LastN(nums, 3)
	if !reflect.DeepEqual(result, []int{3, 4, 5}) {
		t.Errorf("LastN(3): got %v, want [3 4 5]", result)
	}

	// More than length
	result = LastN(nums, 10)
	if !reflect.DeepEqual(result, []int{1, 2, 3, 4, 5}) {
		t.Errorf("LastN(10): got %v, want [1 2 3 4 5]", result)
	}

	// Zero
	result = LastN(nums, 0)
	if len(result) != 0 {
		t.Errorf("LastN(0): got %v, want empty", result)
	}
}

func TestRotate(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	// Positive rotation
	result := Rotate(nums, 2)
	if !reflect.DeepEqual(result, []int{3, 4, 5, 1, 2}) {
		t.Errorf("Rotate(2): got %v, want [3 4 5 1 2]", result)
	}

	// Negative rotation
	result = Rotate(nums, -2)
	if !reflect.DeepEqual(result, []int{4, 5, 1, 2, 3}) {
		t.Errorf("Rotate(-2): got %v, want [4 5 1 2 3]", result)
	}

	// Zero rotation
	result = Rotate(nums, 0)
	if !reflect.DeepEqual(result, []int{1, 2, 3, 4, 5}) {
		t.Errorf("Rotate(0): got %v, want [1 2 3 4 5]", result)
	}

	// Full rotation
	result = Rotate(nums, 5)
	if !reflect.DeepEqual(result, []int{1, 2, 3, 4, 5}) {
		t.Errorf("Rotate(5): got %v, want [1 2 3 4 5]", result)
	}

	// Empty
	result = Rotate([]int{}, 2)
	if len(result) != 0 {
		t.Errorf("Rotate empty: got %v, want empty", result)
	}
}

func TestSelectWithBreak(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Select with early break
	result := Select(nums, func(n int) (bool, bool) {
		if n > 5 {
			return false, false // stop
		}
		return n%2 == 0, true // continue
	})
	expected := []int{2, 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Select with break: got %v, want %v", result, expected)
	}
}

func TestMapWithSkip(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	// Map with skip (include=false)
	result := Map(nums, func(n int) (int, bool, bool) {
		if n%2 == 0 {
			return 0, false, true // skip evens
		}
		return n * 10, true, true
	})
	expected := []int{10, 30, 50}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Map with skip: got %v, want %v", result, expected)
	}
}

func TestReduceWithBreak(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Sum until we exceed 10
	result := Reduce(nums, 0, func(acc, n int) (int, bool) {
		newAcc := acc + n
		if newAcc > 10 {
			return acc, false // stop, keep old acc
		}
		return newAcc, true
	})
	if result != 10 { // 1+2+3+4 = 10
		t.Errorf("Reduce with break: got %d, want 10", result)
	}
}
