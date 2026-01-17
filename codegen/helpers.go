package codegen

import (
	"fmt"
	"strings"
)

// receiverName returns the Go receiver variable name for a class.
// Uses the first letter lowercased, or "x" for empty/invalid input.
func receiverName(className string) string {
	if len(className) == 0 {
		return "x"
	}
	return strings.ToLower(className[:1])
}

// mapType converts Rugby type names to Go type names
func mapType(rubyType string) string {
	// Handle function types: (): Int, (Int): String, etc.
	if strings.HasPrefix(rubyType, "(") && strings.Contains(rubyType, "): ") {
		return mapFunctionType(rubyType)
	}

	// Handle tuple types: (Type1, Type2, ...)
	if strings.HasPrefix(rubyType, "(") && strings.HasSuffix(rubyType, ")") && !strings.Contains(rubyType, "): ") {
		return mapTupleType(rubyType)
	}

	// Handle Array<T>
	if strings.HasPrefix(rubyType, "Array<") && strings.HasSuffix(rubyType, ">") {
		inner := rubyType[6 : len(rubyType)-1]
		return "[]" + mapType(inner)
	}
	// Handle Chan<T>
	if strings.HasPrefix(rubyType, "Chan<") && strings.HasSuffix(rubyType, ">") {
		inner := rubyType[5 : len(rubyType)-1]
		return "chan " + mapType(inner)
	}
	// Handle Map<K, V> and Hash<K, V> (Hash is an alias for Map, Ruby compatibility)
	if (strings.HasPrefix(rubyType, "Map<") || strings.HasPrefix(rubyType, "Hash<")) && strings.HasSuffix(rubyType, ">") {
		// Find the prefix length: "Map<" is 4, "Hash<" is 5
		prefixLen := 4
		if strings.HasPrefix(rubyType, "Hash<") {
			prefixLen = 5
		}
		content := rubyType[prefixLen : len(rubyType)-1]
		// Simple comma split for now (assuming no nested generic commas for MVP)
		parts := strings.Split(content, ",")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			return fmt.Sprintf("map[%s]%s", mapType(key), mapType(val))
		}
	}

	// Handle optional types (T?)
	if baseType, found := strings.CutSuffix(rubyType, "?"); found {
		if valueTypes[baseType] {
			// Value type optional -> *T
			return "*" + mapType(baseType)
		}
		// Reference types: slices/maps are already nullable, classes use pointer
		goBase := mapType(baseType)
		if strings.HasPrefix(goBase, "[]") || strings.HasPrefix(goBase, "map[") {
			return goBase // slices and maps are already nullable
		}
		return "*" + goBase // class types become pointers
	}

	switch rubyType {
	// Integer types
	case "Int":
		return "int"
	case "Int8":
		return "int8"
	case "Int16":
		return "int16"
	case "Int32":
		return "int32"
	case "Int64":
		return "int64"
	case "UInt":
		return "uint"
	case "UInt8", "Byte", "byte":
		return "byte"
	case "UInt16":
		return "uint16"
	case "UInt32":
		return "uint32"
	case "UInt64":
		return "uint64"
	// Float types
	case "Float":
		return "float64"
	case "Float32":
		return "float32"
	// String types
	case "String":
		return "string"
	case "Rune":
		return "rune"
	// Other primitives
	case "Bool":
		return "bool"
	case "Bytes":
		return "[]byte"
	// Container types
	case "Array":
		return "[]any"
	case "Map", "Hash":
		return "map[any]any"
	// Special types
	case "Any", "any":
		return "any"
	case "Error", "error":
		return "error"
	default:
		return rubyType // pass through unknown types (e.g., user-defined)
	}
}

// mapFunctionType converts a Rugby function type to Go func type.
// Examples:
//   - "(): Int" → "func() int"
//   - "(Int): String" → "func(int) string"
//   - "(Int, String): Bool" → "func(int, string) bool"
func mapFunctionType(rubyType string) string {
	// Find the "): " separator
	colonIdx := strings.Index(rubyType, "): ")
	if colonIdx == -1 {
		return rubyType // Not a valid function type
	}

	// Extract params part: "(T1, T2)" and return type
	paramsStr := rubyType[:colonIdx+1] // include the closing paren
	returnStr := rubyType[colonIdx+3:]

	// Parse params (strip outer parens and split by comma)
	if !strings.HasPrefix(paramsStr, "(") || !strings.HasSuffix(paramsStr, ")") {
		return rubyType // Invalid
	}
	paramsInner := paramsStr[1 : len(paramsStr)-1]

	var goParams []string
	if paramsInner != "" {
		// Split by comma, handling nested types
		params := splitTopLevelCommas(paramsInner)
		for _, param := range params {
			goParams = append(goParams, mapType(strings.TrimSpace(param)))
		}
	}

	// Map return type (could be another function type)
	goReturn := mapType(returnStr)

	return fmt.Sprintf("func(%s) %s", strings.Join(goParams, ", "), goReturn)
}

// mapTupleType converts a Rugby tuple type to a Go anonymous struct type.
// Examples:
//   - "(String, Int)" → "struct{ _0 string; _1 int }"
//   - "(String, Float, Bool)" → "struct{ _0 string; _1 float64; _2 bool }"
func mapTupleType(rubyType string) string {
	// Strip outer parentheses
	inner := rubyType[1 : len(rubyType)-1]
	if inner == "" {
		return "struct{}"
	}

	// Split by comma at top level (handling nested types)
	parts := splitTopLevelCommas(inner)
	fields := make([]string, len(parts))
	for i, part := range parts {
		goType := mapType(strings.TrimSpace(part))
		fields[i] = fmt.Sprintf("_%d %s", i, goType)
	}

	return "struct{ " + strings.Join(fields, "; ") + " }"
}

// isTupleType checks if a type annotation represents a tuple type.
// Tuple types look like "(Type1, Type2, ...)" but NOT function types like "(T1): T2".
func isTupleType(typeName string) bool {
	return strings.HasPrefix(typeName, "(") &&
		strings.HasSuffix(typeName, ")") &&
		!strings.Contains(typeName, "): ")
}

// splitTopLevelCommas splits a string by commas, but only at the top level
// (not inside parentheses or angle brackets).
func splitTopLevelCommas(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0

	for _, r := range s {
		switch r {
		case '(', '<', '[':
			depth++
			current.WriteRune(r)
		case ')', '>', ']':
			depth--
			current.WriteRune(r)
		case ',':
			if depth == 0 {
				result = append(result, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// Common acronyms that should be uppercased in Go identifiers
// Per spec section 9.5
var acronyms = map[string]string{
	"id":    "ID",
	"url":   "URL",
	"uri":   "URI",
	"http":  "HTTP",
	"https": "HTTPS",
	"json":  "JSON",
	"xml":   "XML",
	"api":   "API",
	"uuid":  "UUID",
	"ip":    "IP",
	"tcp":   "TCP",
	"udp":   "UDP",
	"sql":   "SQL",
	"tls":   "TLS",
	"ssh":   "SSH",
	"cpu":   "CPU",
	"gpu":   "GPU",
	"html":  "HTML",
	"css":   "CSS",
}

// snakeToPascalWithAcronyms converts snake_case to PascalCase with acronym handling
// Examples: user_id -> UserID, parse_json -> ParseJSON, read_http_url -> ReadHTTPURL
// Predicate methods (ending in ?) get "Is" prefix: active? -> IsActive
// For Go interop (goInterop=true), predicate suffix is just stripped without adding "Is"
func snakeToPascalWithAcronyms(s string, goInterop ...bool) string {
	if s == "" {
		return s
	}

	// Check if this is a Go interop call (don't add "Is" prefix)
	isGoInterop := len(goInterop) > 0 && goInterop[0]

	// Strip Ruby-style ! suffix for mutation methods
	s = strings.TrimSuffix(s, "!")

	// Handle predicate methods (?) with "Is" prefix (unless Go interop)
	isPredicate := strings.HasSuffix(s, "?") && !isGoInterop
	s = strings.TrimSuffix(s, "?")

	// Compute the base result
	var baseResult string

	// If no underscore, check for single-word acronym or capitalize first letter
	if !strings.Contains(s, "_") {
		if upper, ok := acronyms[strings.ToLower(s)]; ok {
			baseResult = upper
		} else if len(s) > 0 {
			// Capitalize first letter only
			baseResult = strings.ToUpper(s[:1]) + s[1:]
		} else {
			baseResult = s
		}
	} else {
		// Split by underscore and process each part
		parts := strings.Split(s, "_")
		var result strings.Builder

		for _, part := range parts {
			if part == "" {
				continue
			}
			if upper, ok := acronyms[strings.ToLower(part)]; ok {
				result.WriteString(upper)
			} else {
				// Capitalize first letter
				result.WriteString(strings.ToUpper(part[:1]))
				if len(part) > 1 {
					result.WriteString(part[1:])
				}
			}
		}
		baseResult = result.String()
	}

	// Add "Is" prefix for predicate methods to avoid collision with property getters
	if isPredicate {
		return "Is" + baseResult
	}
	return baseResult
}

// snakeToCamelWithAcronyms converts snake_case to camelCase with acronym handling
// Examples: user_id -> userID, parse_json -> parseJSON, http_request -> httpRequest
// Predicate methods (ending in ?) get "is" prefix: active? -> isActive
// Note: first-part acronyms stay lowercase in camelCase
func snakeToCamelWithAcronyms(s string) string {
	if s == "" {
		return s
	}

	// Handle predicate methods (?) with "is" prefix
	isPredicate := strings.HasSuffix(s, "?")
	s = strings.TrimSuffix(s, "?")

	// Compute the base result
	var baseResult string

	// If no underscore, return as-is (already camelCase or single word)
	if !strings.Contains(s, "_") {
		baseResult = s
	} else {
		// Split by underscore and process each part
		parts := strings.Split(s, "_")
		var result strings.Builder

		for i, part := range parts {
			if part == "" {
				continue
			}
			if i == 0 {
				// First part: lowercase, even if it's an acronym
				result.WriteString(strings.ToLower(part))
			} else {
				// Subsequent parts: uppercase acronyms, capitalize others
				if upper, ok := acronyms[strings.ToLower(part)]; ok {
					result.WriteString(upper)
				} else {
					result.WriteString(strings.ToUpper(part[:1]))
					if len(part) > 1 {
						result.WriteString(part[1:])
					}
				}
			}
		}
		baseResult = result.String()
	}

	// Add "is" prefix for predicate methods to avoid collision with property getters
	if isPredicate {
		// Capitalize the first letter of baseResult to follow camelCase after "is"
		if len(baseResult) > 0 {
			return "is" + strings.ToUpper(baseResult[:1]) + baseResult[1:]
		}
		return "is"
	}
	return baseResult
}

// zeroValue returns the Go zero value for a Rugby type.
// It checks the generator's tracked interfaces to properly return nil for interface types.
func (g *Generator) zeroValue(rubyType string) string {
	goType := mapType(rubyType)
	switch goType {
	case "int", "int64", "float64":
		return "0"
	case "bool":
		return "false"
	case "string":
		return "\"\""
	case "error", "any":
		return "nil" // interfaces
	default:
		if strings.HasPrefix(goType, "[]") || strings.HasPrefix(goType, "map[") {
			return "nil"
		}
		if strings.HasPrefix(goType, "*") {
			return "nil"
		}
		// Check if this is a declared interface type
		if g.isInterface(rubyType) || g.isInterface(goType) {
			return "nil"
		}
		// Qualified Go types (e.g., io.Reader, http.Handler) are likely interfaces.
		// Return nil as a safe default since interfaces can't use Type{} syntax.
		if strings.Contains(goType, ".") {
			return "nil"
		}
		// For unknown types, assume struct
		return goType + "{}"
	}
}
