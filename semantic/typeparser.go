package semantic

import (
	"strings"
)

// ParseType parses a Rugby type string into a Type.
// Examples: "Int", "String?", "Array[Int]", "Map[String, Int]", "(Int, Bool)"
func ParseType(s string) *Type {
	s = strings.TrimSpace(s)
	if s == "" {
		return TypeUnknownVal
	}

	// Check for optional suffix
	if strings.HasSuffix(s, "?") {
		inner := ParseType(s[:len(s)-1])
		return NewOptionalType(inner)
	}

	// Check for tuple: (T1, T2, ...)
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		inner := s[1 : len(s)-1]
		if inner == "" {
			return NewTupleType()
		}
		parts := splitTopLevel(inner, ',')
		elems := make([]*Type, len(parts))
		for i, p := range parts {
			elems[i] = ParseType(strings.TrimSpace(p))
		}
		return NewTupleType(elems...)
	}

	// Check for generic types: Array[T], Map[K, V], Chan[T], Task[T]
	if idx := strings.Index(s, "["); idx != -1 && strings.HasSuffix(s, "]") {
		base := s[:idx]
		inner := s[idx+1 : len(s)-1]

		switch base {
		case "Array":
			return NewArrayType(ParseType(inner))
		case "Chan":
			return NewChanType(ParseType(inner))
		case "Task":
			return NewTaskType(ParseType(inner))
		case "Map":
			// Split on comma at top level (not inside nested brackets)
			parts := splitTopLevel(inner, ',')
			if len(parts) == 2 {
				return NewMapType(
					ParseType(strings.TrimSpace(parts[0])),
					ParseType(strings.TrimSpace(parts[1])),
				)
			}
			return NewMapType(TypeAnyVal, TypeAnyVal)
		}
	}

	// Primitive types
	switch s {
	case "Int":
		return TypeIntVal
	case "Int64":
		return TypeInt64Val
	case "Float":
		return TypeFloatVal
	case "Bool":
		return TypeBoolVal
	case "String":
		return TypeStringVal
	case "Bytes":
		return TypeBytesVal
	case "any":
		return TypeAnyVal
	case "error":
		return TypeErrorVal
	case "nil":
		return TypeNilVal
	case "Range":
		return TypeRangeVal
	case "":
		return TypeUnknownVal
	}

	// Assume it's a class or interface name
	return NewClassType(s)
}

// splitTopLevel splits a string on a delimiter, but only at the top level
// (not inside brackets).
func splitTopLevel(s string, delim rune) []string {
	var result []string
	var current strings.Builder
	depth := 0

	for _, c := range s {
		switch c {
		case '[', '(':
			depth++
			current.WriteRune(c)
		case ']', ')':
			depth--
			current.WriteRune(c)
		case delim:
			if depth == 0 {
				result = append(result, current.String())
				current.Reset()
			} else {
				current.WriteRune(c)
			}
		default:
			current.WriteRune(c)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
