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

func TestMapDelete(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}

	// Delete existing key
	val, ok := MapDelete(m, "b")
	if !ok || val != 2 {
		t.Errorf("MapDelete existing: got (%d, %v), want (2, true)", val, ok)
	}
	if _, exists := m["b"]; exists {
		t.Error("MapDelete: key still exists")
	}
	if len(m) != 2 {
		t.Errorf("MapDelete: length is %d, want 2", len(m))
	}

	// Delete non-existing key
	val, ok = MapDelete(m, "z")
	if ok {
		t.Errorf("MapDelete non-existing: got (%d, %v), want (0, false)", val, ok)
	}
}

func TestMapHasKey(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}

	if !MapHasKey(m, "a") {
		t.Error("MapHasKey existing: got false, want true")
	}

	if MapHasKey(m, "z") {
		t.Error("MapHasKey non-existing: got true, want false")
	}

	// Empty map
	empty := map[string]int{}
	if MapHasKey(empty, "a") {
		t.Error("MapHasKey empty map: got true, want false")
	}
}

func TestMapClear(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}

	MapClear(m)

	if len(m) != 0 {
		t.Errorf("MapClear: length is %d, want 0", len(m))
	}

	// Clear already empty map
	empty := map[string]int{}
	MapClear(empty) // should not panic
	if len(empty) != 0 {
		t.Error("MapClear empty: unexpected entries")
	}
}

func TestMapInvert(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}

	inverted := MapInvert(m)

	if inverted[1] != "a" || inverted[2] != "b" || inverted[3] != "c" {
		t.Errorf("MapInvert: got %v", inverted)
	}

	if len(inverted) != 3 {
		t.Errorf("MapInvert length: got %d, want 3", len(inverted))
	}

	// Original unchanged
	if m["a"] != 1 {
		t.Error("MapInvert modified original")
	}
}

func TestGetKey(t *testing.T) {
	// Test with map[string]any (common JSON pattern)
	stringMap := map[string]any{"name": "Alice", "age": 30}
	if v := GetKey(stringMap, "name"); v != "Alice" {
		t.Errorf("GetKey map[string]any: got %v, want Alice", v)
	}
	if v := GetKey(stringMap, "missing"); v != nil {
		t.Errorf("GetKey missing key: got %v, want nil", v)
	}

	// Test with map[any]any (Rugby's default map type)
	anyMap := map[any]any{"key": "value", 42: "number key"}
	if v := GetKey(anyMap, "key"); v != "value" {
		t.Errorf("GetKey map[any]any: got %v, want value", v)
	}

	// Test with unsupported type (returns nil)
	if v := GetKey("not a map", "key"); v != nil {
		t.Errorf("GetKey unsupported type: got %v, want nil", v)
	}
	if v := GetKey(nil, "key"); v != nil {
		t.Errorf("GetKey nil: got %v, want nil", v)
	}

	// Test with slice (unsupported, returns nil)
	if v := GetKey([]string{"a", "b"}, "0"); v != nil {
		t.Errorf("GetKey slice: got %v, want nil", v)
	}
}
