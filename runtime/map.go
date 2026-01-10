package runtime

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
	for k, v := range m1 {
		result[k] = v
	}
	for k, v := range m2 {
		result[k] = v
	}
	return result
}

// MapSelect returns a new map with entries for which the predicate returns true.
// The predicate returns (match, continue).
func MapSelect[K comparable, V any](m map[K]V, predicate func(K, V) (bool, bool)) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		match, cont := predicate(k, v)
		if match {
			result[k] = v
		}
		if !cont {
			break
		}
	}
	return result
}

// MapReject returns a new map with entries for which the predicate returns false.
// The predicate returns (match, continue).
func MapReject[K comparable, V any](m map[K]V, predicate func(K, V) (bool, bool)) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		match, cont := predicate(k, v)
		if !match {
			result[k] = v
		}
		if !cont {
			break
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
