package runtime

import (
	"maps"
	"reflect"
)

// MapEach iterates over a map, calling the function for each key-value pair.
// Ruby: hash.each { |k, v| ... }
func MapEach[K comparable, V any](m map[K]V, fn func(K, V)) {
	for k, v := range m {
		fn(k, v)
	}
}

// Keys returns all keys from the map.
// Ruby: hash.keys
func Keys[K comparable, V any](m map[K]V) []K {
	result := make([]K, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// Values returns all values from the map.
// Ruby: hash.values
func Values[K comparable, V any](m map[K]V) []V {
	result := make([]V, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

// Merge returns a new map with entries from both maps.
// Values from the second map override the first.
// Ruby: hash.merge(other)
func Merge[K comparable, V any](m1, m2 map[K]V) map[K]V {
	result := make(map[K]V, len(m1)+len(m2))
	maps.Copy(result, m1)
	maps.Copy(result, m2)
	return result
}

// MapSelect returns a new map with entries for which the predicate returns true.
// Ruby: hash.select { |k, v| ... }
func MapSelect[K comparable, V any](m map[K]V, predicate func(K, V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

// MapReject returns a new map with entries for which the predicate returns false.
// Ruby: hash.reject { |k, v| ... }
func MapReject[K comparable, V any](m map[K]V, predicate func(K, V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if !predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

// Fetch returns the value for the key, or the default if not found.
// Ruby: hash.fetch(key, default)
func Fetch[K comparable, V any](m map[K]V, key K, defaultVal V) V {
	if v, ok := m[key]; ok {
		return v
	}
	return defaultVal
}

// MapDelete removes the key from the map and returns the deleted value.
// Ruby: hash.delete(key)
func MapDelete[K comparable, V any](m map[K]V, key K) (V, bool) {
	val, ok := m[key]
	if ok {
		delete(m, key)
	}
	return val, ok
}

// MapHasKey returns true if the map contains the key.
// Ruby: hash.has_key?(key)
func MapHasKey[K comparable, V any](m map[K]V, key K) bool {
	_, ok := m[key]
	return ok
}

// MapClear removes all entries from the map.
// Ruby: hash.clear
func MapClear[K comparable, V any](m map[K]V) {
	clear(m)
}

// MapInvert returns a new map with keys and values swapped.
// Ruby: hash.invert
func MapInvert[K, V comparable](m map[K]V) map[V]K {
	result := make(map[V]K, len(m))
	for k, v := range m {
		result[v] = k
	}
	return result
}

// GetKey retrieves a value from a map-like value using a string key.
// This handles dynamic access patterns common when working with JSON data
// and supports any map type with string keys via reflection.
func GetKey(m any, key string) any {
	if m == nil {
		return nil
	}

	// Fast path for common types
	switch v := m.(type) {
	case map[string]any:
		return v[key]
	case map[string]string:
		return v[key]
	case map[string]int:
		return v[key]
	case map[string]bool:
		return v[key]
	case map[string]float64:
		return v[key]
	case map[any]any:
		return v[key]
	}

	// Slow path: use reflection for other map types
	rv := reflect.ValueOf(m)
	if rv.Kind() != reflect.Map {
		return nil
	}

	// Check if map has string keys
	if rv.Type().Key().Kind() != reflect.String {
		return nil
	}

	// Look up the value
	keyVal := reflect.ValueOf(key)
	result := rv.MapIndex(keyVal)
	if !result.IsValid() {
		return nil
	}
	return result.Interface()
}
