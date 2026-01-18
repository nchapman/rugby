package runtime

import (
	"sort"
	"testing"
)

func TestSetContains(t *testing.T) {
	s := map[int]struct{}{1: {}, 2: {}, 3: {}}

	if !SetContains(s, 1) {
		t.Error("SetContains(s, 1): expected true, got false")
	}
	if !SetContains(s, 2) {
		t.Error("SetContains(s, 2): expected true, got false")
	}
	if SetContains(s, 4) {
		t.Error("SetContains(s, 4): expected false, got true")
	}

	// Empty set
	empty := map[int]struct{}{}
	if SetContains(empty, 1) {
		t.Error("SetContains(empty, 1): expected false, got true")
	}
}

func TestSetAdd(t *testing.T) {
	s := map[int]struct{}{}

	SetAdd(s, 1)
	if !SetContains(s, 1) {
		t.Error("SetAdd failed to add element")
	}

	// Adding duplicate should not error
	SetAdd(s, 1)
	if len(s) != 1 {
		t.Errorf("SetAdd duplicate: expected size 1, got %d", len(s))
	}

	SetAdd(s, 2)
	if len(s) != 2 {
		t.Errorf("SetAdd second element: expected size 2, got %d", len(s))
	}
}

func TestSetDelete(t *testing.T) {
	s := map[int]struct{}{1: {}, 2: {}, 3: {}}

	SetDelete(s, 2)
	if SetContains(s, 2) {
		t.Error("SetDelete failed to remove element")
	}
	if len(s) != 2 {
		t.Errorf("SetDelete: expected size 2, got %d", len(s))
	}

	// Deleting non-existent element should not error
	SetDelete(s, 99)
	if len(s) != 2 {
		t.Errorf("SetDelete non-existent: expected size 2, got %d", len(s))
	}
}

func TestSetUnion(t *testing.T) {
	s1 := map[int]struct{}{1: {}, 2: {}}
	s2 := map[int]struct{}{2: {}, 3: {}}

	result := SetUnion(s1, s2)

	if len(result) != 3 {
		t.Errorf("SetUnion: expected size 3, got %d", len(result))
	}
	for _, v := range []int{1, 2, 3} {
		if !SetContains(result, v) {
			t.Errorf("SetUnion: missing element %d", v)
		}
	}

	// Original sets should be unchanged
	if len(s1) != 2 || len(s2) != 2 {
		t.Error("SetUnion modified original sets")
	}

	// Empty set union
	empty := map[int]struct{}{}
	result2 := SetUnion(s1, empty)
	if len(result2) != 2 {
		t.Errorf("SetUnion with empty: expected size 2, got %d", len(result2))
	}
}

func TestSetIntersection(t *testing.T) {
	s1 := map[int]struct{}{1: {}, 2: {}, 3: {}}
	s2 := map[int]struct{}{2: {}, 3: {}, 4: {}}

	result := SetIntersection(s1, s2)

	if len(result) != 2 {
		t.Errorf("SetIntersection: expected size 2, got %d", len(result))
	}
	for _, v := range []int{2, 3} {
		if !SetContains(result, v) {
			t.Errorf("SetIntersection: missing element %d", v)
		}
	}
	if SetContains(result, 1) || SetContains(result, 4) {
		t.Error("SetIntersection: contains elements not in both sets")
	}

	// Disjoint sets
	s3 := map[int]struct{}{5: {}, 6: {}}
	result2 := SetIntersection(s1, s3)
	if len(result2) != 0 {
		t.Errorf("SetIntersection disjoint: expected size 0, got %d", len(result2))
	}
}

func TestSetDifference(t *testing.T) {
	s1 := map[int]struct{}{1: {}, 2: {}, 3: {}}
	s2 := map[int]struct{}{2: {}, 3: {}, 4: {}}

	result := SetDifference(s1, s2)

	if len(result) != 1 {
		t.Errorf("SetDifference: expected size 1, got %d", len(result))
	}
	if !SetContains(result, 1) {
		t.Error("SetDifference: missing element 1")
	}
	if SetContains(result, 2) || SetContains(result, 3) {
		t.Error("SetDifference: contains elements from s2")
	}

	// Difference with empty set
	empty := map[int]struct{}{}
	result2 := SetDifference(s1, empty)
	if len(result2) != 3 {
		t.Errorf("SetDifference with empty: expected size 3, got %d", len(result2))
	}
}

func TestSetSize(t *testing.T) {
	s := map[int]struct{}{1: {}, 2: {}, 3: {}}
	if SetSize(s) != 3 {
		t.Errorf("SetSize: expected 3, got %d", SetSize(s))
	}

	empty := map[int]struct{}{}
	if SetSize(empty) != 0 {
		t.Errorf("SetSize empty: expected 0, got %d", SetSize(empty))
	}
}

func TestSetToArray(t *testing.T) {
	s := map[int]struct{}{3: {}, 1: {}, 2: {}}
	result := SetToArray(s)

	if len(result) != 3 {
		t.Errorf("SetToArray: expected length 3, got %d", len(result))
	}

	// Sort for deterministic comparison
	sort.Ints(result)
	expected := []int{1, 2, 3}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("SetToArray[%d]: got %d, want %d", i, v, expected[i])
		}
	}

	// Empty set
	empty := map[int]struct{}{}
	result2 := SetToArray(empty)
	if len(result2) != 0 {
		t.Errorf("SetToArray empty: expected length 0, got %d", len(result2))
	}
}

func TestSetClear(t *testing.T) {
	s := map[int]struct{}{1: {}, 2: {}, 3: {}}
	SetClear(s)

	if len(s) != 0 {
		t.Errorf("SetClear: expected size 0, got %d", len(s))
	}

	// Clear empty set should not error
	SetClear(s)
	if len(s) != 0 {
		t.Errorf("SetClear empty: expected size 0, got %d", len(s))
	}
}

func TestSetEach(t *testing.T) {
	s := map[int]struct{}{1: {}, 2: {}, 3: {}}
	var collected []int

	SetEach(s, func(elem int) {
		collected = append(collected, elem)
	})

	if len(collected) != 3 {
		t.Errorf("SetEach: expected 3 elements, got %d", len(collected))
	}

	// Sort for deterministic comparison
	sort.Ints(collected)
	expected := []int{1, 2, 3}
	for i, v := range collected {
		if v != expected[i] {
			t.Errorf("SetEach collected[%d]: got %d, want %d", i, v, expected[i])
		}
	}

	// Empty set
	var count int
	SetEach(map[int]struct{}{}, func(elem int) {
		count++
	})
	if count != 0 {
		t.Errorf("SetEach empty: expected 0 iterations, got %d", count)
	}
}

func TestSetWithStrings(t *testing.T) {
	s := map[string]struct{}{"a": {}, "b": {}, "c": {}}

	if !SetContains(s, "a") {
		t.Error("SetContains string: expected true")
	}

	SetAdd(s, "d")
	if !SetContains(s, "d") {
		t.Error("SetAdd string failed")
	}

	SetDelete(s, "a")
	if SetContains(s, "a") {
		t.Error("SetDelete string failed")
	}
}
