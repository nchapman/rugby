package semantic

import (
	"strings"
)

// ParseTypeWithParams parses a Rugby type string into a Type, recognizing type parameters.
// typeParams maps type parameter names to their constraints (empty string if unconstrained).
// e.g., {"T": "Ordered", "K": "", "V": ""}
func ParseTypeWithParams(s string, typeParams map[string]string) *Type {
	s = strings.TrimSpace(s)
	if s == "" {
		return TypeUnknownVal
	}

	// Check if it's a type parameter (must check before other parsing)
	if typeParams != nil {
		if constraint, ok := typeParams[s]; ok {
			if constraint != "" {
				return NewConstrainedTypeParamType(s, constraint)
			}
			return NewTypeParamType(s)
		}
	}

	// Check for optional suffix
	if strings.HasSuffix(s, "?") {
		inner := ParseTypeWithParams(s[:len(s)-1], typeParams)
		return NewOptionalType(inner)
	}

	// Check for function type: (Params): Return
	if parenIdx := findTopLevelColon(s); parenIdx != -1 {
		paramsPart := strings.TrimSpace(s[:parenIdx+1]) // include closing paren
		returnPart := strings.TrimSpace(s[parenIdx+3:]) // skip "): "

		var paramTypes []*Type
		if strings.HasPrefix(paramsPart, "(") && strings.HasSuffix(paramsPart, ")") {
			inner := paramsPart[1 : len(paramsPart)-1]
			if inner != "" {
				parts := splitTopLevel(inner)
				for _, p := range parts {
					paramTypes = append(paramTypes, ParseTypeWithParams(strings.TrimSpace(p), typeParams))
				}
			}
		} else if paramsPart != "" {
			paramTypes = append(paramTypes, ParseTypeWithParams(paramsPart, typeParams))
		}

		returnType := ParseTypeWithParams(returnPart, typeParams)
		return NewFuncType(paramTypes, []*Type{returnType})
	}

	// Check for tuple: (T1, T2, ...)
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		inner := s[1 : len(s)-1]
		if inner == "" {
			return NewTupleType()
		}
		parts := splitTopLevel(inner)
		elems := make([]*Type, len(parts))
		for i, p := range parts {
			elems[i] = ParseTypeWithParams(strings.TrimSpace(p), typeParams)
		}
		return NewTupleType(elems...)
	}

	// Check for generic types: Array<T>, Map<K, V>, Chan<T>, Task<T>
	if idx := strings.Index(s, "<"); idx != -1 && strings.HasSuffix(s, ">") {
		base := s[:idx]
		inner := s[idx+1 : len(s)-1]

		switch base {
		case "Array":
			return NewArrayType(ParseTypeWithParams(inner, typeParams))
		case "Chan":
			return NewChanType(ParseTypeWithParams(inner, typeParams))
		case "Task":
			return NewTaskType(ParseTypeWithParams(inner, typeParams))
		case "Map", "Hash":
			parts := splitTopLevel(inner)
			if len(parts) == 2 {
				return NewMapType(
					ParseTypeWithParams(strings.TrimSpace(parts[0]), typeParams),
					ParseTypeWithParams(strings.TrimSpace(parts[1]), typeParams),
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
	case "Rune":
		return TypeRuneVal
	case "Any", "any":
		return TypeAnyVal
	case "Error", "error":
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

// ParseType parses a Rugby type string into a Type.
// Examples: "Int", "String?", "Array<Int>", "Map<String, Int>", "(Int, Bool)"
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

	// Check for function type: (Params): Return
	if parenIdx := findTopLevelColon(s); parenIdx != -1 {
		paramsPart := strings.TrimSpace(s[:parenIdx+1]) // include closing paren
		returnPart := strings.TrimSpace(s[parenIdx+3:]) // skip "): "

		var paramTypes []*Type
		// Handle (Params) or single Param
		if strings.HasPrefix(paramsPart, "(") && strings.HasSuffix(paramsPart, ")") {
			inner := paramsPart[1 : len(paramsPart)-1]
			if inner != "" {
				parts := splitTopLevel(inner)
				for _, p := range parts {
					paramTypes = append(paramTypes, ParseType(strings.TrimSpace(p)))
				}
			}
		} else if paramsPart != "" {
			// Single parameter without parens
			paramTypes = append(paramTypes, ParseType(paramsPart))
		}

		returnType := ParseType(returnPart)
		return NewFuncType(paramTypes, []*Type{returnType})
	}

	// Check for tuple: (T1, T2, ...)
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		inner := s[1 : len(s)-1]
		if inner == "" {
			return NewTupleType()
		}
		parts := splitTopLevel(inner)
		elems := make([]*Type, len(parts))
		for i, p := range parts {
			elems[i] = ParseType(strings.TrimSpace(p))
		}
		return NewTupleType(elems...)
	}

	// Check for generic types: Array<T>, Map<K, V>, Chan<T>, Task<T>, Hash<K, V>
	if idx := strings.Index(s, "<"); idx != -1 && strings.HasSuffix(s, ">") {
		base := s[:idx]
		inner := s[idx+1 : len(s)-1]

		switch base {
		case "Array":
			return NewArrayType(ParseType(inner))
		case "Chan":
			return NewChanType(ParseType(inner))
		case "Task":
			return NewTaskType(ParseType(inner))
		case "Map", "Hash":
			// Hash is an alias for Map (Ruby compatibility)
			// Split on comma at top level (not inside nested brackets)
			parts := splitTopLevel(inner)
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
	case "Rune":
		return TypeRuneVal
	case "Any", "any":
		return TypeAnyVal
	case "Error", "error":
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

// splitTopLevel splits a string on commas, but only at the top level
// (not inside brackets). Used for parsing generic type parameters.
func splitTopLevel(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0

	for _, c := range s {
		switch c {
		case '<', '(':
			depth++
			current.WriteRune(c)
		case '>', ')':
			depth--
			current.WriteRune(c)
		case ',':
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

// findTopLevelColon finds the index of "): " at the top level (function type separator).
// Returns the index of the closing paren, or -1 if not found.
func findTopLevelColon(s string) int {
	depth := 0
	for i, c := range s {
		switch c {
		case '<', '(':
			depth++
		case '>':
			depth--
		case ')':
			depth--
			// Check if this is "): " pattern at top level
			if depth == 0 && i+2 < len(s) && s[i+1] == ':' && s[i+2] == ' ' {
				return i
			}
		}
	}
	return -1
}
