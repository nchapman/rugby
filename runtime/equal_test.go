package runtime

import "testing"

// Test type with custom Equal method
type testPoint struct {
	x, y int
}

func (p *testPoint) Equal(other any) bool {
	o, ok := other.(*testPoint)
	if !ok {
		return false
	}
	return p.x == o.x && p.y == o.y
}

func TestEqualWithCustomType(t *testing.T) {
	p1 := &testPoint{1, 2}
	p2 := &testPoint{1, 2}
	p3 := &testPoint{3, 4}

	if !Equal(p1, p2) {
		t.Error("Equal(p1, p2) should be true for same coordinates")
	}

	if Equal(p1, p3) {
		t.Error("Equal(p1, p3) should be false for different coordinates")
	}
}

func TestEqualWithDifferentTypes(t *testing.T) {
	p := &testPoint{1, 2}

	if Equal(p, "not a point") {
		t.Error("Equal(point, string) should be false")
	}

	if Equal(p, 42) {
		t.Error("Equal(point, int) should be false")
	}
}

func TestEqualWithNil(t *testing.T) {
	if !Equal(nil, nil) {
		t.Error("Equal(nil, nil) should be true")
	}

	p := &testPoint{1, 2}
	if Equal(p, nil) {
		t.Error("Equal(point, nil) should be false")
	}

	if Equal(nil, p) {
		t.Error("Equal(nil, point) should be false")
	}
}

func TestEqualWithPrimitives(t *testing.T) {
	// Integers
	if !Equal(42, 42) {
		t.Error("Equal(42, 42) should be true")
	}
	if Equal(42, 43) {
		t.Error("Equal(42, 43) should be false")
	}

	// Strings
	if !Equal("hello", "hello") {
		t.Error("Equal(hello, hello) should be true")
	}
	if Equal("hello", "world") {
		t.Error("Equal(hello, world) should be false")
	}

	// Bools
	if !Equal(true, true) {
		t.Error("Equal(true, true) should be true")
	}
	if Equal(true, false) {
		t.Error("Equal(true, false) should be false")
	}

	// Floats
	if !Equal(3.14, 3.14) {
		t.Error("Equal(3.14, 3.14) should be true")
	}
	if Equal(3.14, 2.71) {
		t.Error("Equal(3.14, 2.71) should be false")
	}
}

func TestEqualWithSlices(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{1, 2, 3}
	c := []int{1, 2, 4}
	d := []int{1, 2}

	if !Equal(a, b) {
		t.Error("Equal([1,2,3], [1,2,3]) should be true")
	}

	if Equal(a, c) {
		t.Error("Equal([1,2,3], [1,2,4]) should be false")
	}

	if Equal(a, d) {
		t.Error("Equal([1,2,3], [1,2]) should be false (different lengths)")
	}
}

func TestEqualWithMaps(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2}
	b := map[string]int{"a": 1, "b": 2}
	c := map[string]int{"a": 1, "b": 3}
	d := map[string]int{"a": 1}

	if !Equal(a, b) {
		t.Error("Equal(map1, map2) should be true for same content")
	}

	if Equal(a, c) {
		t.Error("Equal(map1, map3) should be false for different values")
	}

	if Equal(a, d) {
		t.Error("Equal(map1, map4) should be false for different sizes")
	}
}

func TestEqualWithNestedSlices(t *testing.T) {
	a := [][]int{{1, 2}, {3, 4}}
	b := [][]int{{1, 2}, {3, 4}}
	c := [][]int{{1, 2}, {3, 5}}

	if !Equal(a, b) {
		t.Error("Equal for nested slices with same content should be true")
	}

	if Equal(a, c) {
		t.Error("Equal for nested slices with different content should be false")
	}
}

func TestEqualWithSlicesContainingCustomTypes(t *testing.T) {
	a := []*testPoint{{1, 2}, {3, 4}}
	b := []*testPoint{{1, 2}, {3, 4}}
	c := []*testPoint{{1, 2}, {5, 6}}

	if !Equal(a, b) {
		t.Error("Equal for slices of custom types with same content should be true")
	}

	if Equal(a, c) {
		t.Error("Equal for slices of custom types with different content should be false")
	}
}
