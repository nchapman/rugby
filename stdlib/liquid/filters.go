package liquid

import (
	"fmt"
	"html"
	"reflect"
	"strings"
)

// FilterFunc is the signature for filter functions.
// Input is the value being filtered, args are any additional arguments.
type FilterFunc func(input any, args ...any) any

// filters is the global filter registry.
var filters = map[string]FilterFunc{
	// String filters
	"upcase":   filterUpcase,
	"downcase": filterDowncase,
	"strip":    filterStrip,
	"escape":   filterEscape,

	// Array filters
	"first":   filterFirst,
	"last":    filterLast,
	"size":    filterSize,
	"join":    filterJoin,
	"reverse": filterReverse,

	// Utility filters
	"default": filterDefault,

	// Math filters
	"plus":  filterPlus,
	"minus": filterMinus,
}

// String filters

func filterUpcase(input any, args ...any) any {
	return strings.ToUpper(toString(input))
}

func filterDowncase(input any, args ...any) any {
	return strings.ToLower(toString(input))
}

func filterStrip(input any, args ...any) any {
	return strings.TrimSpace(toString(input))
}

func filterEscape(input any, args ...any) any {
	return html.EscapeString(toString(input))
}

// Array filters

func filterFirst(input any, args ...any) any {
	slice := toSlice(input)
	if len(slice) == 0 {
		return nil
	}
	return slice[0]
}

func filterLast(input any, args ...any) any {
	slice := toSlice(input)
	if len(slice) == 0 {
		return nil
	}
	return slice[len(slice)-1]
}

func filterSize(input any, args ...any) any {
	if input == nil {
		return 0
	}
	switch v := input.(type) {
	case string:
		return len(v)
	case []any:
		return len(v)
	default:
		rv := reflect.ValueOf(input)
		switch rv.Kind() {
		case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
			return rv.Len()
		}
	}
	return 0
}

func filterJoin(input any, args ...any) any {
	slice := toSlice(input)
	sep := " " // default separator
	if len(args) > 0 {
		sep = toString(args[0])
	}

	strs := make([]string, len(slice))
	for i, v := range slice {
		strs[i] = toString(v)
	}
	return strings.Join(strs, sep)
}

func filterReverse(input any, args ...any) any {
	slice := toSlice(input)
	result := make([]any, len(slice))
	for i, v := range slice {
		result[len(slice)-1-i] = v
	}
	return result
}

// Utility filters

func filterDefault(input any, args ...any) any {
	if isFalsy(input) {
		if len(args) > 0 {
			return args[0]
		}
		return ""
	}
	return input
}

// Math filters

func filterPlus(input any, args ...any) any {
	if len(args) == 0 {
		return input
	}

	a := toNumber(input)
	b := toNumber(args[0])

	// If either is a float, return float
	if _, ok := a.(float64); ok {
		return toFloat(a) + toFloat(b)
	}
	if _, ok := b.(float64); ok {
		return toFloat(a) + toFloat(b)
	}
	return toInt(a) + toInt(b)
}

func filterMinus(input any, args ...any) any {
	if len(args) == 0 {
		return input
	}

	a := toNumber(input)
	b := toNumber(args[0])

	// If either is a float, return float
	if _, ok := a.(float64); ok {
		return toFloat(a) - toFloat(b)
	}
	if _, ok := b.(float64); ok {
		return toFloat(a) - toFloat(b)
	}
	return toInt(a) - toInt(b)
}

// Type conversion utilities

// toString converts any value to a string.
func toString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		// Format without trailing zeros
		s := fmt.Sprintf("%g", val)
		return s
	default:
		return fmt.Sprintf("%v", val)
	}
}

// toSlice converts any value to a slice.
func toSlice(v any) []any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []any:
		return val
	case []string:
		result := make([]any, len(val))
		for i, s := range val {
			result[i] = s
		}
		return result
	case []int:
		result := make([]any, len(val))
		for i, n := range val {
			result[i] = n
		}
		return result
	case string:
		// Split string into characters
		result := make([]any, len(val))
		for i, c := range val {
			result[i] = string(c)
		}
		return result
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			result := make([]any, rv.Len())
			for i := range rv.Len() {
				result[i] = rv.Index(i).Interface()
			}
			return result
		}
	}
	return nil
}

// toNumber converts any value to a number (int64 or float64).
func toNumber(v any) any {
	if v == nil {
		return int64(0)
	}
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return val
	case string:
		// Try parsing as int first, then float
		var i int64
		if _, err := fmt.Sscanf(val, "%d", &i); err == nil {
			return i
		}
		var f float64
		if _, err := fmt.Sscanf(val, "%f", &f); err == nil {
			return f
		}
		return int64(0)
	case bool:
		if val {
			return int64(1)
		}
		return int64(0)
	default:
		return int64(0)
	}
}

// toInt converts any value to int64.
func toInt(v any) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	default:
		return 0
	}
}

// toFloat converts any value to float64.
func toFloat(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	default:
		return 0
	}
}

// toBool converts any value to a boolean.
func toBool(v any) bool {
	return !isFalsy(v)
}

// isFalsy returns true if the value is considered falsy in Liquid.
// Falsy: nil, false, empty string
func isFalsy(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case bool:
		return !val
	case string:
		return val == ""
	case emptyValue:
		return true
	case blankValue:
		return true
	}
	return false
}

// isEmpty checks if a value matches the "empty" special value.
// Empty: nil, empty string, empty array, empty map
func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case []any:
		return len(val) == 0
	case map[string]any:
		return len(val) == 0
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
			return rv.Len() == 0
		}
	}
	return false
}

// isBlank checks if a value matches the "blank" special value.
// Blank: nil, false, empty string, whitespace-only string, empty array, empty map
func isBlank(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case bool:
		return !val
	case string:
		return strings.TrimSpace(val) == ""
	case []any:
		return len(val) == 0
	case map[string]any:
		return len(val) == 0
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Array, reflect.Slice, reflect.Map:
			return rv.Len() == 0
		case reflect.String:
			return strings.TrimSpace(rv.String()) == ""
		}
	}
	return false
}

// getFilter returns a filter function by name.
func getFilter(name string) (FilterFunc, bool) {
	f, ok := filters[name]
	return f, ok
}
