package runtime

import (
	"reflect"
	"sort"
	"testing"
)

func TestKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := Keys(m)

	// Sort for deterministic comparison
	sort.Strings(keys)
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("Keys: got %v, want %v", keys, expected)
	}

	// Empty map
	emptyKeys := Keys(map[string]int{})
	if len(emptyKeys) != 0 {
		t.Errorf("Keys empty: got %v, want empty", emptyKeys)
	}
}

func TestValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	values := Values(m)

	// Sort for deterministic comparison
	sort.Ints(values)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Values: got %v, want %v", values, expected)
	}
}

func TestMerge(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"b": 20, "c": 3}

	merged := Merge(m1, m2)

	// Second map overrides
	if merged["b"] != 20 {
		t.Errorf("Merge override: got %d, want 20", merged["b"])
	}

	// All keys present
	if len(merged) != 3 {
		t.Errorf("Merge length: got %d, want 3", len(merged))
	}

	// Original maps unchanged
	if m1["b"] != 2 {
		t.Errorf("Merge modified m1: got %d, want 2", m1["b"])
	}
}

func TestMapSelect(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}

	selected := MapSelect(m, func(k string, v int) (bool, bool) {
		return v > 2, true
	})

	if len(selected) != 2 {
		t.Errorf("MapSelect length: got %d, want 2", len(selected))
	}

	if selected["c"] != 3 || selected["d"] != 4 {
		t.Errorf("MapSelect: got %v", selected)
	}
}

func TestMapReject(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}

	rejected := MapReject(m, func(k string, v int) (bool, bool) {
		return v > 2, true
	})

	if len(rejected) != 2 {
		t.Errorf("MapReject length: got %d, want 2", len(rejected))
	}

	if rejected["a"] != 1 || rejected["b"] != 2 {
		t.Errorf("MapReject: got %v", rejected)
	}
}

func TestFetch(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}

	// Key exists
	if v := Fetch(m, "a", 99); v != 1 {
		t.Errorf("Fetch existing: got %d, want 1", v)
	}

	// Key missing
	if v := Fetch(m, "z", 99); v != 99 {
		t.Errorf("Fetch missing: got %d, want 99", v)
	}
}
