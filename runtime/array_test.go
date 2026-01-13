package runtime

import (
	"reflect"
	"strings"
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

func TestIndex(t *testing.T) {
	arr := []int{10, 20, 30, 40, 50}

	// Positive indices
	if got := Index(arr, 0); got != 10 {
		t.Errorf("Index(arr, 0) = %d; want 10", got)
	}
	if got := Index(arr, 2); got != 30 {
		t.Errorf("Index(arr, 2) = %d; want 30", got)
	}

	// Negative indices
	if got := Index(arr, -1); got != 50 {
		t.Errorf("Index(arr, -1) = %d; want 50", got)
	}
	if got := Index(arr, -2); got != 40 {
		t.Errorf("Index(arr, -2) = %d; want 40", got)
	}
	if got := Index(arr, -5); got != 10 {
		t.Errorf("Index(arr, -5) = %d; want 10", got)
	}
}

func TestIndexOpt(t *testing.T) {
	arr := []int{10, 20, 30}

	// Valid indices
	if val, ok := IndexOpt(arr, 0); !ok || val != 10 {
		t.Errorf("IndexOpt(arr, 0) = %d, %v; want 10, true", val, ok)
	}
	if val, ok := IndexOpt(arr, -1); !ok || val != 30 {
		t.Errorf("IndexOpt(arr, -1) = %d, %v; want 30, true", val, ok)
	}

	// Out of bounds
	if _, ok := IndexOpt(arr, 3); ok {
		t.Error("IndexOpt(arr, 3) should return false for out of bounds")
	}
	if _, ok := IndexOpt(arr, -4); ok {
		t.Error("IndexOpt(arr, -4) should return false for out of bounds")
	}
}

func TestAtIndex(t *testing.T) {
	// Test with int slice
	intArr := []int{10, 20, 30, 40, 50}
	if got := AtIndex(intArr, 0).(int); got != 10 {
		t.Errorf("AtIndex(intArr, 0) = %d; want 10", got)
	}
	if got := AtIndex(intArr, -1).(int); got != 50 {
		t.Errorf("AtIndex(intArr, -1) = %d; want 50", got)
	}
	if got := AtIndex(intArr, -2).(int); got != 40 {
		t.Errorf("AtIndex(intArr, -2) = %d; want 40", got)
	}

	// Test with string slice
	strArr := []string{"a", "b", "c"}
	if got := AtIndex(strArr, 1).(string); got != "b" {
		t.Errorf("AtIndex(strArr, 1) = %s; want b", got)
	}
	if got := AtIndex(strArr, -1).(string); got != "c" {
		t.Errorf("AtIndex(strArr, -1) = %s; want c", got)
	}

	// Test with string (character indexing)
	str := "hello"
	if got := AtIndex(str, 0).(string); got != "h" {
		t.Errorf("AtIndex(str, 0) = %s; want h", got)
	}
	if got := AtIndex(str, -1).(string); got != "o" {
		t.Errorf("AtIndex(str, -1) = %s; want o", got)
	}
	if got := AtIndex(str, -2).(string); got != "l" {
		t.Errorf("AtIndex(str, -2) = %s; want l", got)
	}

	// Test with unicode string
	unicode := "héllo"
	if got := AtIndex(unicode, 1).(string); got != "é" {
		t.Errorf("AtIndex(unicode, 1) = %s; want é", got)
	}
	if got := AtIndex(unicode, -1).(string); got != "o" {
		t.Errorf("AtIndex(unicode, -1) = %s; want o", got)
	}
}

func TestAtIndexOpt(t *testing.T) {
	// Test with slice
	arr := []int{10, 20, 30}
	if val, ok := AtIndexOpt(arr, 0); !ok || val.(int) != 10 {
		t.Errorf("AtIndexOpt(arr, 0) = %v, %v; want 10, true", val, ok)
	}
	if val, ok := AtIndexOpt(arr, -1); !ok || val.(int) != 30 {
		t.Errorf("AtIndexOpt(arr, -1) = %v, %v; want 30, true", val, ok)
	}
	if _, ok := AtIndexOpt(arr, 3); ok {
		t.Error("AtIndexOpt(arr, 3) should return false")
	}
	if _, ok := AtIndexOpt(arr, -4); ok {
		t.Error("AtIndexOpt(arr, -4) should return false")
	}

	// Test with string
	str := "abc"
	if val, ok := AtIndexOpt(str, 0); !ok || val.(string) != "a" {
		t.Errorf("AtIndexOpt(str, 0) = %v, %v; want a, true", val, ok)
	}
	if val, ok := AtIndexOpt(str, -1); !ok || val.(string) != "c" {
		t.Errorf("AtIndexOpt(str, -1) = %v, %v; want c, true", val, ok)
	}
	if _, ok := AtIndexOpt(str, 3); ok {
		t.Error("AtIndexOpt(str, 3) should return false")
	}
	if _, ok := AtIndexOpt(str, -4); ok {
		t.Error("AtIndexOpt(str, -4) should return false")
	}
}

func TestIndexPanicsOutOfBounds(t *testing.T) {
	arr := []int{1, 2, 3}

	// Test positive out-of-bounds
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Index with positive out-of-bounds did not panic")
			}
		}()
		Index(arr, 10)
	}()

	// Test negative out-of-bounds
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Index with negative out-of-bounds did not panic")
			}
		}()
		Index(arr, -10)
	}()
}

func TestAtIndexPanicsOutOfBounds(t *testing.T) {
	arr := []int{1, 2, 3}

	// Test positive out-of-bounds
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("AtIndex with positive out-of-bounds did not panic")
			}
		}()
		AtIndex(arr, 10)
	}()

	// Test negative out-of-bounds
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("AtIndex with negative out-of-bounds did not panic")
			}
		}()
		AtIndex(arr, -10)
	}()
}

func TestShiftLeftSlice(t *testing.T) {
	// Test with int slice
	arr := []int{1, 2, 3}
	result := ShiftLeft(arr, 4).([]int)
	if len(result) != 4 || result[3] != 4 {
		t.Errorf("ShiftLeft([]int, 4) = %v; want [1,2,3,4]", result)
	}

	// Original slice unchanged
	if len(arr) != 3 {
		t.Errorf("Original slice modified: %v", arr)
	}

	// Test with string slice
	strs := []string{"a", "b"}
	result2 := ShiftLeft(strs, "c").([]string)
	if len(result2) != 3 || result2[2] != "c" {
		t.Errorf("ShiftLeft([]string, c) = %v; want [a,b,c]", result2)
	}

	// Test with []any
	anys := []any{1, "two"}
	result3 := ShiftLeft(anys, 3.0).([]any)
	if len(result3) != 3 {
		t.Errorf("ShiftLeft([]any, 3.0) = %v; want [1,two,3.0]", result3)
	}
}

func TestShiftLeftChaining(t *testing.T) {
	// Test chaining: arr << 1 << 2 << 3
	arr := []int{}
	result := ShiftLeft(ShiftLeft(ShiftLeft(arr, 1), 2), 3).([]int)
	if len(result) != 3 || result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("Chained ShiftLeft = %v; want [1,2,3]", result)
	}
}

func TestShiftLeftChannel(t *testing.T) {
	// Test with channel
	ch := make(chan int, 1)
	result := ShiftLeft(ch, 42)

	// Should return the same channel
	if result != ch {
		t.Error("ShiftLeft on channel should return the same channel")
	}

	// Value should have been sent
	val := <-ch
	if val != 42 {
		t.Errorf("Channel received %d; want 42", val)
	}
}

func TestSliceInclusiveRange(t *testing.T) {
	arr := []int{0, 1, 2, 3, 4}

	// arr[1..3] should return [1, 2, 3]
	result := Slice(arr, Range{Start: 1, End: 3, Exclusive: false}).([]int)
	if len(result) != 3 || result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("Slice([0,1,2,3,4], 1..3) = %v; want [1,2,3]", result)
	}
}

func TestSliceExclusiveRange(t *testing.T) {
	arr := []int{0, 1, 2, 3, 4}

	// arr[1...3] should return [1, 2] (exclusive end)
	result := Slice(arr, Range{Start: 1, End: 3, Exclusive: true}).([]int)
	if len(result) != 2 || result[0] != 1 || result[1] != 2 {
		t.Errorf("Slice([0,1,2,3,4], 1...3) = %v; want [1,2]", result)
	}
}

func TestSliceNegativeIndices(t *testing.T) {
	arr := []int{0, 1, 2, 3, 4}

	// arr[1..-2] should return [1, 2, 3] (index -2 is 3)
	result := Slice(arr, Range{Start: 1, End: -2, Exclusive: false}).([]int)
	if len(result) != 3 || result[0] != 1 || result[2] != 3 {
		t.Errorf("Slice([0,1,2,3,4], 1..-2) = %v; want [1,2,3]", result)
	}

	// arr[-3..-1] should return [2, 3, 4]
	result2 := Slice(arr, Range{Start: -3, End: -1, Exclusive: false}).([]int)
	if len(result2) != 3 || result2[0] != 2 || result2[2] != 4 {
		t.Errorf("Slice([0,1,2,3,4], -3..-1) = %v; want [2,3,4]", result2)
	}
}

func TestSliceString(t *testing.T) {
	str := "hello"

	// str[1..3] should return "ell"
	result := Slice(str, Range{Start: 1, End: 3, Exclusive: false}).(string)
	if result != "ell" {
		t.Errorf("Slice(\"hello\", 1..3) = %q; want \"ell\"", result)
	}

	// str[0..-1] should return "hello" (entire string)
	result2 := Slice(str, Range{Start: 0, End: -1, Exclusive: false}).(string)
	if result2 != "hello" {
		t.Errorf("Slice(\"hello\", 0..-1) = %q; want \"hello\"", result2)
	}
}

func TestSliceEmptyResult(t *testing.T) {
	arr := []int{0, 1, 2, 3, 4}

	// arr[3..1] should return [] (start > end after normalization)
	result := Slice(arr, Range{Start: 3, End: 1, Exclusive: false}).([]int)
	if len(result) != 0 {
		t.Errorf("Slice([0,1,2,3,4], 3..1) = %v; want []", result)
	}

	// arr[10..20] should return [] (out of bounds)
	result2 := Slice(arr, Range{Start: 10, End: 20, Exclusive: false}).([]int)
	if len(result2) != 0 {
		t.Errorf("Slice([0,1,2,3,4], 10..20) = %v; want []", result2)
	}
}

func TestSliceNegativeOutOfBounds(t *testing.T) {
	arr := []int{0, 1, 2}

	// Start index -10 resolves to negative, should return empty slice
	result := Slice(arr, Range{Start: -10, End: -1, Exclusive: false}).([]int)
	if len(result) != 0 {
		t.Errorf("Slice with out-of-bounds negative start = %v; want []", result)
	}
}

func TestSliceEndBeyondLength(t *testing.T) {
	arr := []int{0, 1, 2}

	// arr[1..100] should return [1, 2] (end clamped to length)
	result := Slice(arr, Range{Start: 1, End: 100, Exclusive: false}).([]int)
	if len(result) != 2 || result[0] != 1 || result[1] != 2 {
		t.Errorf("Slice([0,1,2], 1..100) = %v; want [1,2]", result)
	}
}

func TestSliceUnicodeString(t *testing.T) {
	str := "héllo" // Contains multi-byte character

	// Should slice by rune, not byte
	result := Slice(str, Range{Start: 0, End: 2, Exclusive: false}).(string)
	if result != "hél" {
		t.Errorf("Slice Unicode string = %q; want \"hél\"", result)
	}

	// Slice the accented e
	result2 := Slice(str, Range{Start: 1, End: 1, Exclusive: false}).(string)
	if result2 != "é" {
		t.Errorf("Slice single Unicode char = %q; want \"é\"", result2)
	}
}

// Helper type for testing CallMethod
type testStringer struct {
	value string
}

func (t testStringer) Upcase() string {
	return strings.ToUpper(t.value)
}

func (t testStringer) Length() int {
	return len(t.value)
}

func TestCallMethodBasic(t *testing.T) {
	obj := testStringer{value: "hello"}

	// Call Upcase method
	result := CallMethod(obj, "upcase").(string)
	if result != "HELLO" {
		t.Errorf("CallMethod(obj, 'upcase') = %q; want \"HELLO\"", result)
	}

	// Call Length method
	length := CallMethod(obj, "length").(int)
	if length != 5 {
		t.Errorf("CallMethod(obj, 'length') = %d; want 5", length)
	}
}

func TestCallMethodSnakeCase(t *testing.T) {
	// toGoMethodName should convert snake_case to PascalCase
	name := toGoMethodName("to_upper")
	if name != "ToUpper" {
		t.Errorf("toGoMethodName('to_upper') = %q; want \"ToUpper\"", name)
	}

	name2 := toGoMethodName("get_user_id")
	if name2 != "GetUserId" {
		t.Errorf("toGoMethodName('get_user_id') = %q; want \"GetUserId\"", name2)
	}
}

func TestCallMethodPredicates(t *testing.T) {
	// Test predicate method naming (empty? -> Empty_PRED)
	name := toGoMethodName("empty?")
	if name != "Empty_PRED" {
		t.Errorf("toGoMethodName('empty?') = %q; want \"Empty_PRED\"", name)
	}
}

func TestCallMethodWithMapIntegration(t *testing.T) {
	// Test symbol-to-proc with Map - simulates names.map(&:upcase)
	names := []any{
		testStringer{value: "alice"},
		testStringer{value: "bob"},
	}

	result := Map(names, func(x any) (any, bool, bool) {
		return CallMethod(x, "upcase"), true, true
	})

	if result[0] != "ALICE" || result[1] != "BOB" {
		t.Errorf("Map with CallMethod = %v; want [ALICE, BOB]", result)
	}
}

func TestEach_AllTypes(t *testing.T) {
	// Test []any
	var anyResults []any
	Each([]any{1, "two", 3.0}, func(v any) bool {
		anyResults = append(anyResults, v)
		return true
	})
	if len(anyResults) != 3 {
		t.Errorf("Each []any: got %d elements, want 3", len(anyResults))
	}

	// Test []float64
	var floatResults []float64
	Each([]float64{1.1, 2.2, 3.3}, func(v any) bool {
		floatResults = append(floatResults, v.(float64))
		return true
	})
	if !reflect.DeepEqual(floatResults, []float64{1.1, 2.2, 3.3}) {
		t.Errorf("Each []float64: got %v", floatResults)
	}

	// Test []bool
	var boolResults []bool
	Each([]bool{true, false, true}, func(v any) bool {
		boolResults = append(boolResults, v.(bool))
		return true
	})
	if !reflect.DeepEqual(boolResults, []bool{true, false, true}) {
		t.Errorf("Each []bool: got %v", boolResults)
	}

	// Test early break with []float64
	var partialFloats []float64
	Each([]float64{1.0, 2.0, 3.0, 4.0}, func(v any) bool {
		partialFloats = append(partialFloats, v.(float64))
		return v.(float64) < 2.5
	})
	if len(partialFloats) != 3 {
		t.Errorf("Each []float64 with break: got %d elements, want 3", len(partialFloats))
	}

	// Test early break with []bool
	var partialBools []bool
	Each([]bool{true, true, false, true}, func(v any) bool {
		partialBools = append(partialBools, v.(bool))
		return v.(bool)
	})
	if len(partialBools) != 3 {
		t.Errorf("Each []bool with break: got %d elements, want 3", len(partialBools))
	}

	// Test early break with []any
	var partialAny []any
	Each([]any{1, 2, 3, 4}, func(v any) bool {
		partialAny = append(partialAny, v)
		return v.(int) < 3
	})
	if len(partialAny) != 3 {
		t.Errorf("Each []any with break: got %d elements, want 3", len(partialAny))
	}

	// Test reflection fallback with custom slice type
	type customInt int
	customSlice := []customInt{1, 2, 3}
	var customResults []customInt
	Each(customSlice, func(v any) bool {
		customResults = append(customResults, v.(customInt))
		return true
	})
	if len(customResults) != 3 {
		t.Errorf("Each custom slice: got %d elements, want 3", len(customResults))
	}
}

func TestEachWithIndex_AllTypes(t *testing.T) {
	// Test []string
	var strPairs []struct {
		val string
		idx int
	}
	EachWithIndex([]string{"a", "b", "c"}, func(v any, i int) bool {
		strPairs = append(strPairs, struct {
			val string
			idx int
		}{v.(string), i})
		return true
	})
	if len(strPairs) != 3 || strPairs[2].val != "c" || strPairs[2].idx != 2 {
		t.Errorf("EachWithIndex []string: got %v", strPairs)
	}

	// Test []float64
	var floatPairs []struct {
		val float64
		idx int
	}
	EachWithIndex([]float64{1.1, 2.2}, func(v any, i int) bool {
		floatPairs = append(floatPairs, struct {
			val float64
			idx int
		}{v.(float64), i})
		return true
	})
	if len(floatPairs) != 2 {
		t.Errorf("EachWithIndex []float64: got %d elements", len(floatPairs))
	}

	// Test []bool
	var boolPairs []struct {
		val bool
		idx int
	}
	EachWithIndex([]bool{true, false}, func(v any, i int) bool {
		boolPairs = append(boolPairs, struct {
			val bool
			idx int
		}{v.(bool), i})
		return true
	})
	if len(boolPairs) != 2 {
		t.Errorf("EachWithIndex []bool: got %d elements", len(boolPairs))
	}

	// Test []any
	var anyPairs []struct {
		val any
		idx int
	}
	EachWithIndex([]any{"x", "y"}, func(v any, i int) bool {
		anyPairs = append(anyPairs, struct {
			val any
			idx int
		}{v, i})
		return true
	})
	if len(anyPairs) != 2 {
		t.Errorf("EachWithIndex []any: got %d elements", len(anyPairs))
	}

	// Test early break
	var partial []int
	EachWithIndex([]int{1, 2, 3, 4}, func(v any, i int) bool {
		partial = append(partial, v.(int))
		return i < 2
	})
	if len(partial) != 3 {
		t.Errorf("EachWithIndex with break: got %d elements, want 3", len(partial))
	}

	// Test reflection fallback
	type customStr string
	customSlice := []customStr{"a", "b"}
	var customPairs []struct {
		val customStr
		idx int
	}
	EachWithIndex(customSlice, func(v any, i int) bool {
		customPairs = append(customPairs, struct {
			val customStr
			idx int
		}{v.(customStr), i})
		return true
	})
	if len(customPairs) != 2 {
		t.Errorf("EachWithIndex custom slice: got %d elements", len(customPairs))
	}
}

func TestAtIndex_AllTypes(t *testing.T) {
	// Test []int
	if AtIndex([]int{10, 20, 30}, 1).(int) != 20 {
		t.Error("AtIndex []int positive failed")
	}
	if AtIndex([]int{10, 20, 30}, -1).(int) != 30 {
		t.Error("AtIndex []int negative failed")
	}

	// Test []string
	if AtIndex([]string{"a", "b", "c"}, 0).(string) != "a" {
		t.Error("AtIndex []string positive failed")
	}
	if AtIndex([]string{"a", "b", "c"}, -2).(string) != "b" {
		t.Error("AtIndex []string negative failed")
	}

	// Test []float64
	if AtIndex([]float64{1.1, 2.2, 3.3}, 2).(float64) != 3.3 {
		t.Error("AtIndex []float64 positive failed")
	}
	if AtIndex([]float64{1.1, 2.2, 3.3}, -3).(float64) != 1.1 {
		t.Error("AtIndex []float64 negative failed")
	}

	// Test []bool
	if AtIndex([]bool{true, false, true}, 1).(bool) != false {
		t.Error("AtIndex []bool positive failed")
	}
	if AtIndex([]bool{true, false, true}, -1).(bool) != true {
		t.Error("AtIndex []bool negative failed")
	}

	// Test []any
	if AtIndex([]any{1, "two", 3.0}, 1) != "two" {
		t.Error("AtIndex []any positive failed")
	}
	if AtIndex([]any{1, "two", 3.0}, -1).(float64) != 3.0 {
		t.Error("AtIndex []any negative failed")
	}

	// Test string
	if AtIndex("hello", 1).(string) != "e" {
		t.Error("AtIndex string positive failed")
	}
	if AtIndex("hello", -1).(string) != "o" {
		t.Error("AtIndex string negative failed")
	}

	// Test reflection fallback
	type customInt int
	customSlice := []customInt{1, 2, 3}
	if AtIndex(customSlice, -1).(customInt) != 3 {
		t.Error("AtIndex custom slice negative failed")
	}
}

func TestAtIndexOpt_AllTypes(t *testing.T) {
	// Test []int - valid indices
	v, ok := AtIndexOpt([]int{10, 20, 30}, 1)
	if !ok || v.(int) != 20 {
		t.Error("AtIndexOpt []int valid positive failed")
	}
	v, ok = AtIndexOpt([]int{10, 20, 30}, -1)
	if !ok || v.(int) != 30 {
		t.Error("AtIndexOpt []int valid negative failed")
	}

	// Test []int - invalid indices
	_, ok = AtIndexOpt([]int{1, 2}, 5)
	if ok {
		t.Error("AtIndexOpt []int out of bounds should return false")
	}
	_, ok = AtIndexOpt([]int{1, 2}, -5)
	if ok {
		t.Error("AtIndexOpt []int negative out of bounds should return false")
	}

	// Test []string - valid
	v, ok = AtIndexOpt([]string{"a", "b"}, 0)
	if !ok || v.(string) != "a" {
		t.Error("AtIndexOpt []string valid failed")
	}
	// Test []string - invalid
	_, ok = AtIndexOpt([]string{"a"}, 10)
	if ok {
		t.Error("AtIndexOpt []string out of bounds should return false")
	}

	// Test []float64
	v, ok = AtIndexOpt([]float64{1.1, 2.2}, -1)
	if !ok || v.(float64) != 2.2 {
		t.Error("AtIndexOpt []float64 valid negative failed")
	}
	_, ok = AtIndexOpt([]float64{1.1}, 5)
	if ok {
		t.Error("AtIndexOpt []float64 out of bounds should return false")
	}

	// Test []bool
	v, ok = AtIndexOpt([]bool{true, false}, 0)
	if !ok || v.(bool) != true {
		t.Error("AtIndexOpt []bool valid failed")
	}
	_, ok = AtIndexOpt([]bool{true}, -5)
	if ok {
		t.Error("AtIndexOpt []bool negative out of bounds should return false")
	}

	// Test []any
	v, ok = AtIndexOpt([]any{1, "two"}, 1)
	if !ok || v.(string) != "two" {
		t.Error("AtIndexOpt []any valid failed")
	}
	_, ok = AtIndexOpt([]any{}, 0)
	if ok {
		t.Error("AtIndexOpt empty []any should return false")
	}

	// Test string
	v, ok = AtIndexOpt("hello", 1)
	if !ok || v.(string) != "e" {
		t.Error("AtIndexOpt string valid failed")
	}
	_, ok = AtIndexOpt("hi", 10)
	if ok {
		t.Error("AtIndexOpt string out of bounds should return false")
	}

	// Test reflection fallback
	type customInt int
	customSlice := []customInt{1, 2}
	v, ok = AtIndexOpt(customSlice, -1)
	if !ok || v.(customInt) != 2 {
		t.Error("AtIndexOpt custom slice valid negative failed")
	}
	_, ok = AtIndexOpt(customSlice, 10)
	if ok {
		t.Error("AtIndexOpt custom slice out of bounds should return false")
	}
}

func TestSlice_AllTypes(t *testing.T) {
	// Test []int with inclusive range
	result := Slice([]int{1, 2, 3, 4, 5}, Range{Start: 1, End: 3, Exclusive: false}).([]int)
	if !reflect.DeepEqual(result, []int{2, 3, 4}) {
		t.Errorf("Slice []int inclusive: got %v, want [2 3 4]", result)
	}

	// Test []int with exclusive range
	result = Slice([]int{1, 2, 3, 4, 5}, Range{Start: 1, End: 3, Exclusive: true}).([]int)
	if !reflect.DeepEqual(result, []int{2, 3}) {
		t.Errorf("Slice []int exclusive: got %v, want [2 3]", result)
	}

	// Test []int with negative indices
	result = Slice([]int{1, 2, 3, 4, 5}, Range{Start: -3, End: -1, Exclusive: false}).([]int)
	if !reflect.DeepEqual(result, []int{3, 4, 5}) {
		t.Errorf("Slice []int negative: got %v, want [3 4 5]", result)
	}

	// Test []string
	strResult := Slice([]string{"a", "b", "c", "d"}, Range{Start: 0, End: 2, Exclusive: false}).([]string)
	if !reflect.DeepEqual(strResult, []string{"a", "b", "c"}) {
		t.Errorf("Slice []string: got %v", strResult)
	}

	// Test []float64
	floatResult := Slice([]float64{1.1, 2.2, 3.3}, Range{Start: 0, End: 1, Exclusive: false}).([]float64)
	if !reflect.DeepEqual(floatResult, []float64{1.1, 2.2}) {
		t.Errorf("Slice []float64: got %v", floatResult)
	}

	// Test []bool
	boolResult := Slice([]bool{true, false, true}, Range{Start: 1, End: 2, Exclusive: false}).([]bool)
	if !reflect.DeepEqual(boolResult, []bool{false, true}) {
		t.Errorf("Slice []bool: got %v", boolResult)
	}

	// Test []any
	anyResult := Slice([]any{1, "two", 3.0}, Range{Start: 0, End: 1, Exclusive: false}).([]any)
	if len(anyResult) != 2 {
		t.Errorf("Slice []any: got %d elements, want 2", len(anyResult))
	}

	// Test string
	strSlice := Slice("hello world", Range{Start: 0, End: 4, Exclusive: false}).(string)
	if strSlice != "hello" {
		t.Errorf("Slice string: got %q, want \"hello\"", strSlice)
	}

	// Test out of bounds (should return empty)
	emptyResult := Slice([]int{1, 2}, Range{Start: 10, End: 20, Exclusive: false}).([]int)
	if len(emptyResult) != 0 {
		t.Errorf("Slice out of bounds: got %v, want empty", emptyResult)
	}

	// Test reflection fallback
	type customInt int
	customSlice := []customInt{1, 2, 3, 4}
	customResult := Slice(customSlice, Range{Start: 1, End: 2, Exclusive: false}).([]customInt)
	if len(customResult) != 2 {
		t.Errorf("Slice custom: got %d elements, want 2", len(customResult))
	}
}

func TestShiftLeft_Chan(t *testing.T) {
	ch := make(chan int, 5)
	result := ShiftLeft(ch, 42)
	if result != ch {
		t.Error("ShiftLeft chan should return the same channel")
	}
	received := <-ch
	if received != 42 {
		t.Errorf("ShiftLeft chan: got %d, want 42", received)
	}
}

func TestMaxMinFloat_Empty(t *testing.T) {
	_, ok := MinFloat([]float64{})
	if ok {
		t.Error("MinFloat empty should return false")
	}
	_, ok = MaxFloat([]float64{})
	if ok {
		t.Error("MaxFloat empty should return false")
	}
}

func TestMaxMinInt_Empty(t *testing.T) {
	_, ok := MinInt([]int{})
	if ok {
		t.Error("MinInt empty should return false")
	}
	_, ok = MaxInt([]int{})
	if ok {
		t.Error("MaxInt empty should return false")
	}
}

func TestCallMethod_NotFound(t *testing.T) {
	obj := struct{ Name string }{Name: "test"}
	defer func() {
		if r := recover(); r == nil {
			t.Error("CallMethod should panic for non-existent method")
		}
	}()
	CallMethod(obj, "nonexistent_method")
}

func TestShift(t *testing.T) {
	// Generic Shift
	arr := []int{1, 2, 3}
	val, ok := Shift(&arr)
	if !ok || val != 1 {
		t.Errorf("Shift: got %d, ok=%v, want 1, ok=true", val, ok)
	}
	if !reflect.DeepEqual(arr, []int{2, 3}) {
		t.Errorf("Shift: remaining %v, want [2 3]", arr)
	}

	// Empty slice
	empty := []int{}
	val, ok = Shift(&empty)
	if ok {
		t.Errorf("Shift empty: ok=%v, want false", ok)
	}

	// Single element
	single := []string{"only"}
	strVal, strOk := Shift(&single)
	if !strOk || strVal != "only" || len(single) != 0 {
		t.Errorf("Shift single: got %q, remaining %v", strVal, single)
	}
}

func TestShiftInt(t *testing.T) {
	arr := []int{10, 20, 30}
	val := ShiftInt(&arr)
	if val != 10 {
		t.Errorf("ShiftInt: got %d, want 10", val)
	}
	if !reflect.DeepEqual(arr, []int{20, 30}) {
		t.Errorf("ShiftInt: remaining %v, want [20 30]", arr)
	}

	// Empty returns zero
	empty := []int{}
	val = ShiftInt(&empty)
	if val != 0 {
		t.Errorf("ShiftInt empty: got %d, want 0", val)
	}
}

func TestShiftString(t *testing.T) {
	arr := []string{"a", "b", "c"}
	val := ShiftString(&arr)
	if val != "a" {
		t.Errorf("ShiftString: got %q, want 'a'", val)
	}
	if !reflect.DeepEqual(arr, []string{"b", "c"}) {
		t.Errorf("ShiftString: remaining %v, want ['b' 'c']", arr)
	}

	// Empty returns empty string
	empty := []string{}
	val = ShiftString(&empty)
	if val != "" {
		t.Errorf("ShiftString empty: got %q, want ''", val)
	}
}

func TestShiftFloat(t *testing.T) {
	arr := []float64{1.1, 2.2, 3.3}
	val := ShiftFloat(&arr)
	if val != 1.1 {
		t.Errorf("ShiftFloat: got %f, want 1.1", val)
	}
	if !reflect.DeepEqual(arr, []float64{2.2, 3.3}) {
		t.Errorf("ShiftFloat: remaining %v, want [2.2 3.3]", arr)
	}
}

func TestShiftAny(t *testing.T) {
	arr := []any{1, "two", 3.0}
	val := ShiftAny(&arr)
	if val != 1 {
		t.Errorf("ShiftAny: got %v, want 1", val)
	}
	if len(arr) != 2 {
		t.Errorf("ShiftAny: remaining length %d, want 2", len(arr))
	}
}

func TestPop(t *testing.T) {
	// Generic Pop
	arr := []int{1, 2, 3}
	val, ok := Pop(&arr)
	if !ok || val != 3 {
		t.Errorf("Pop: got %d, ok=%v, want 3, ok=true", val, ok)
	}
	if !reflect.DeepEqual(arr, []int{1, 2}) {
		t.Errorf("Pop: remaining %v, want [1 2]", arr)
	}

	// Empty slice
	empty := []int{}
	val, ok = Pop(&empty)
	if ok {
		t.Errorf("Pop empty: ok=%v, want false", ok)
	}

	// Single element
	single := []string{"only"}
	strVal, strOk := Pop(&single)
	if !strOk || strVal != "only" || len(single) != 0 {
		t.Errorf("Pop single: got %q, remaining %v", strVal, single)
	}
}

func TestPopInt(t *testing.T) {
	arr := []int{10, 20, 30}
	val := PopInt(&arr)
	if val != 30 {
		t.Errorf("PopInt: got %d, want 30", val)
	}
	if !reflect.DeepEqual(arr, []int{10, 20}) {
		t.Errorf("PopInt: remaining %v, want [10 20]", arr)
	}

	// Empty returns zero
	empty := []int{}
	val = PopInt(&empty)
	if val != 0 {
		t.Errorf("PopInt empty: got %d, want 0", val)
	}
}

func TestPopString(t *testing.T) {
	arr := []string{"a", "b", "c"}
	val := PopString(&arr)
	if val != "c" {
		t.Errorf("PopString: got %q, want 'c'", val)
	}
	if !reflect.DeepEqual(arr, []string{"a", "b"}) {
		t.Errorf("PopString: remaining %v, want ['a' 'b']", arr)
	}

	// Empty returns empty string
	empty := []string{}
	val = PopString(&empty)
	if val != "" {
		t.Errorf("PopString empty: got %q, want ''", val)
	}
}

func TestPopFloat(t *testing.T) {
	arr := []float64{1.1, 2.2, 3.3}
	val := PopFloat(&arr)
	if val != 3.3 {
		t.Errorf("PopFloat: got %f, want 3.3", val)
	}
	if !reflect.DeepEqual(arr, []float64{1.1, 2.2}) {
		t.Errorf("PopFloat: remaining %v, want [1.1 2.2]", arr)
	}
}

func TestPopAny(t *testing.T) {
	arr := []any{1, "two", 3.0}
	val := PopAny(&arr)
	if val != 3.0 {
		t.Errorf("PopAny: got %v, want 3.0", val)
	}
	if len(arr) != 2 {
		t.Errorf("PopAny: remaining length %d, want 2", len(arr))
	}
}
