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
	// Handle Map<K, V>
	if strings.HasPrefix(rubyType, "Map<") && strings.HasSuffix(rubyType, ">") {
		content := rubyType[4 : len(rubyType)-1]
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
	case "UInt8", "Byte":
		return "uint8"
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
	case "Map":
		return "map[any]any"
	// Special types
	case "any":
		return "any"
	case "error":
		return "error"
	default:
		return rubyType // pass through unknown types (e.g., user-defined)
	}
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
func snakeToPascalWithAcronyms(s string) string {
	if s == "" {
		return s
	}

	// Strip Ruby-style suffixes (! for mutation, ? for predicates)
	s = strings.TrimSuffix(s, "!")
	s = strings.TrimSuffix(s, "?")

	// If no underscore, check for single-word acronym or capitalize first letter
	if !strings.Contains(s, "_") {
		if upper, ok := acronyms[strings.ToLower(s)]; ok {
			return upper
		}
		// Capitalize first letter only
		if len(s) > 0 {
			return strings.ToUpper(s[:1]) + s[1:]
		}
		return s
	}

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

	return result.String()
}

// snakeToCamelWithAcronyms converts snake_case to camelCase with acronym handling
// Examples: user_id -> userID, parse_json -> parseJSON, http_request -> httpRequest
// Note: first-part acronyms stay lowercase in camelCase
func snakeToCamelWithAcronyms(s string) string {
	if s == "" {
		return s
	}

	// Strip Ruby-style ? suffix for predicate methods
	s = strings.TrimSuffix(s, "?")

	// If no underscore, return as-is (already camelCase or single word)
	if !strings.Contains(s, "_") {
		return s
	}

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

	return result.String()
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
