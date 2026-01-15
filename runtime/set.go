package runtime

// SetContains returns true if the set contains the element.
// Ruby: set.include?(elem)
func SetContains[T comparable](s map[T]struct{}, elem T) bool {
	_, ok := s[elem]
	return ok
}

// SetAdd adds an element to the set.
// Ruby: set.add(elem)
func SetAdd[T comparable](s map[T]struct{}, elem T) {
	s[elem] = struct{}{}
}

// SetDelete removes an element from the set.
// Ruby: set.delete(elem)
func SetDelete[T comparable](s map[T]struct{}, elem T) {
	delete(s, elem)
}

// SetUnion returns a new set with all elements from both sets.
// Ruby: set | other
func SetUnion[T comparable](s1, s2 map[T]struct{}) map[T]struct{} {
	result := make(map[T]struct{}, len(s1)+len(s2))
	for k := range s1 {
		result[k] = struct{}{}
	}
	for k := range s2 {
		result[k] = struct{}{}
	}
	return result
}

// SetIntersection returns a new set with elements present in both sets.
// Ruby: set & other
func SetIntersection[T comparable](s1, s2 map[T]struct{}) map[T]struct{} {
	result := make(map[T]struct{})
	// Iterate over the smaller set for efficiency
	if len(s1) > len(s2) {
		s1, s2 = s2, s1
	}
	for k := range s1 {
		if _, ok := s2[k]; ok {
			result[k] = struct{}{}
		}
	}
	return result
}

// SetDifference returns a new set with elements in s1 but not in s2.
// Ruby: set - other
func SetDifference[T comparable](s1, s2 map[T]struct{}) map[T]struct{} {
	result := make(map[T]struct{})
	for k := range s1 {
		if _, ok := s2[k]; !ok {
			result[k] = struct{}{}
		}
	}
	return result
}

// SetSize returns the number of elements in the set.
// Ruby: set.size
func SetSize[T comparable](s map[T]struct{}) int {
	return len(s)
}

// SetToArray converts a set to an array.
// Ruby: set.to_a
func SetToArray[T comparable](s map[T]struct{}) []T {
	result := make([]T, 0, len(s))
	for k := range s {
		result = append(result, k)
	}
	return result
}

// SetClear removes all elements from the set.
// Ruby: set.clear
func SetClear[T comparable](s map[T]struct{}) {
	clear(s)
}

// SetEach iterates over all elements in the set.
// Ruby: set.each { |elem| ... }
func SetEach[T comparable](s map[T]struct{}, fn func(T)) {
	for k := range s {
		fn(k)
	}
}
