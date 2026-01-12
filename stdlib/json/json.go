// Package json provides JSON parsing and generation for Rugby programs.
// Rugby: import rugby/json
//
// Example:
//
//	data = json.parse('{"name": "Alice"}')!
//	str = json.generate(data)!
package json

import (
	"encoding/json"
	"fmt"
)

// Parse parses a JSON string into a map.
// Ruby: json.parse(str)
func Parse(s string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ParseArray parses a JSON string into an array.
// Ruby: json.parse_array(str)
func ParseArray(s string) ([]any, error) {
	var result []any
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ParseInto parses a JSON string into the provided struct.
// Ruby: json.parse_into(str, target)
func ParseInto(s string, v any) error {
	return json.Unmarshal([]byte(s), v)
}

// ParseBytes parses JSON bytes into a map.
// Ruby: json.parse_bytes(bytes)
func ParseBytes(b []byte) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Generate converts a value to a JSON string.
// Ruby: json.generate(data)
func Generate(v any) (string, error) {
	// Convert map[any]any to map[string]any for JSON compatibility
	converted := convertForJSON(v)
	b, err := json.Marshal(converted)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GenerateBytes converts a value to JSON bytes.
// Ruby: json.generate_bytes(data)
func GenerateBytes(v any) ([]byte, error) {
	converted := convertForJSON(v)
	return json.Marshal(converted)
}

// Pretty converts a value to a pretty-printed JSON string.
// Ruby: json.pretty(data)
func Pretty(v any) (string, error) {
	converted := convertForJSON(v)
	b, err := json.MarshalIndent(converted, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// PrettyBytes converts a value to pretty-printed JSON bytes.
// Ruby: json.pretty_bytes(data)
func PrettyBytes(v any) ([]byte, error) {
	converted := convertForJSON(v)
	return json.MarshalIndent(converted, "", "  ")
}

// convertForJSON recursively converts map[any]any to map[string]any
// so that Go's encoding/json can marshal it properly.
func convertForJSON(v any) any {
	switch val := v.(type) {
	case map[any]any:
		result := make(map[string]any, len(val))
		for k, v := range val {
			// Convert key to string
			var keyStr string
			switch key := k.(type) {
			case string:
				keyStr = key
			default:
				// Use fmt.Sprint for other types
				keyStr = fmt.Sprint(key)
			}
			result[keyStr] = convertForJSON(v)
		}
		return result
	case map[string]any:
		result := make(map[string]any, len(val))
		for k, v := range val {
			result[k] = convertForJSON(v)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, v := range val {
			result[i] = convertForJSON(v)
		}
		return result
	default:
		return v
	}
}

// Valid reports whether s is valid JSON.
// Ruby: json.valid?(str)
func Valid(s string) bool {
	return json.Valid([]byte(s))
}

// ValidBytes reports whether b is valid JSON.
// Ruby: json.valid_bytes?(bytes)
func ValidBytes(b []byte) bool {
	return json.Valid(b)
}
