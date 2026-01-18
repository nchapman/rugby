package liquid

import (
	"fmt"
	"maps"
	"strings"
)

// CompiledTemplate holds a pre-compiled render function for compile-time templates.
// This type is used by the Rugby codegen to embed templates directly into the binary
// as optimized Go code, eliminating runtime parsing overhead.
type CompiledTemplate struct {
	Render func(data map[string]any) (string, error)
}

// MustRender executes the template with the given data and panics on error.
func (t *CompiledTemplate) MustRender(data any) string {
	converted := convertToStringAny(data)
	result, err := t.Render(converted)
	if err != nil {
		panic(err)
	}
	return result
}

// RenderAny executes the template with data that may not be a map[string]any.
// It converts the data to the appropriate format first.
func (t *CompiledTemplate) RenderAny(data any) (string, error) {
	converted := convertToStringAny(data)
	return t.Render(converted)
}

// --- Exported types and functions for codegen ---

// Context manages variable scopes during template rendering.
// This is exported for use by compiled templates.
type Context struct {
	vars   map[string]any
	parent *Context
}

// NewContext creates a new Context with the given initial data.
func NewContext(data map[string]any) *Context {
	vars := make(map[string]any)
	maps.Copy(vars, data)
	return &Context{vars: vars}
}

// Get retrieves a variable from the context.
func (c *Context) Get(name string) any {
	if val, ok := c.vars[name]; ok {
		return val
	}
	if c.parent != nil {
		return c.parent.Get(name)
	}
	return nil
}

// Set sets a variable in the current scope.
func (c *Context) Set(name string, value any) {
	c.vars[name] = value
}

// Push creates a new child scope.
func (c *Context) Push() *Context {
	return &Context{
		vars:   make(map[string]any),
		parent: c,
	}
}

// SetForloop sets the forloop variable with loop metadata.
func (c *Context) SetForloop(index0, length int) {
	c.vars["forloop"] = map[string]any{
		"index":   index0 + 1,
		"index0":  index0,
		"rindex":  length - index0,
		"rindex0": length - index0 - 1,
		"first":   index0 == 0,
		"last":    index0 == length-1,
		"length":  length,
	}
}

// --- Exported filter functions for codegen ---

// FilterUpcase converts a string to uppercase.
func FilterUpcase(input any) any {
	return strings.ToUpper(ToString(input))
}

// FilterDowncase converts a string to lowercase.
func FilterDowncase(input any) any {
	return strings.ToLower(ToString(input))
}

// FilterCapitalize capitalizes the first character of a string.
func FilterCapitalize(input any) any {
	return filterCapitalize(input)
}

// FilterStrip removes leading and trailing whitespace.
func FilterStrip(input any) any {
	return strings.TrimSpace(ToString(input))
}

// FilterEscape escapes HTML entities.
func FilterEscape(input any) any {
	return filterEscape(input)
}

// FilterFirst returns the first element of an array.
func FilterFirst(input any) any {
	return filterFirst(input)
}

// FilterLast returns the last element of an array.
func FilterLast(input any) any {
	return filterLast(input)
}

// FilterSize returns the size of a string or array.
func FilterSize(input any) any {
	return filterSize(input)
}

// FilterJoin joins array elements with a separator.
func FilterJoin(input any, separator string) any {
	return filterJoin(input, separator)
}

// FilterReverse reverses an array.
func FilterReverse(input any) any {
	return filterReverse(input)
}

// FilterDefault returns the default value if input is falsy.
func FilterDefault(input any, defaultValue any) any {
	return filterDefault(input, defaultValue)
}

// FilterPlus adds two numbers.
func FilterPlus(input any, addend any) any {
	return filterPlus(input, addend)
}

// FilterMinus subtracts a number from another.
func FilterMinus(input any, subtrahend any) any {
	return filterMinus(input, subtrahend)
}

// --- Exported type conversion functions for codegen ---

// ToString converts any value to a string.
func ToString(v any) string {
	return toString(v)
}

// ToBool converts any value to a boolean (Liquid semantics).
func ToBool(v any) bool {
	return toBool(v)
}

// ToSlice converts any value to a slice.
func ToSlice(v any) []any {
	return toSlice(v)
}

// IsFalsy returns true if the value is considered falsy in Liquid.
func IsFalsy(v any) bool {
	return isFalsy(v)
}

// IsEmpty checks if a value matches the "empty" special value.
func IsEmpty(v any) bool {
	return isEmpty(v)
}

// IsBlank checks if a value matches the "blank" special value.
func IsBlank(v any) bool {
	return isBlank(v)
}

// GetProperty retrieves a property from an object (map or struct).
func GetProperty(obj any, property string) any {
	return getProperty(obj, property)
}

// GetIndex retrieves an element from an array or map by index.
func GetIndex(obj any, index any) any {
	return getIndex(obj, index)
}

// CompareValues compares two values for Liquid binary operations.
func CompareValues(left any, op string, right any) bool {
	switch op {
	case "==":
		return equal(left, right)
	case "!=":
		return !equal(left, right)
	case "<":
		return compare(left, right) < 0
	case ">":
		return compare(left, right) > 0
	case "<=":
		return compare(left, right) <= 0
	case ">=":
		return compare(left, right) >= 0
	case "contains":
		return contains(left, right)
	default:
		return false
	}
}

// MakeRange creates a range slice from start to end (inclusive).
func MakeRange(start, end any) []any {
	startInt := toInt(toNumber(start))
	endInt := toInt(toNumber(end))

	if startInt > endInt {
		return nil
	}

	result := make([]any, endInt-startInt+1)
	for i := startInt; i <= endInt; i++ {
		result[i-startInt] = int(i)
	}
	return result
}

// ConvertData converts various map types to map[string]any.
func ConvertData(data any) map[string]any {
	return convertToStringAny(data)
}

// ToIntValue safely converts any value to an int for use in limit/offset.
// Returns 0 if the value cannot be converted to an integer.
func ToIntValue(v any) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		var i int64
		if _, err := fmt.Sscanf(val, "%d", &i); err == nil {
			return int(i)
		}
	}
	return 0
}
