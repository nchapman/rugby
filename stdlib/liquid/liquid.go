// Package liquid provides a Liquid template engine for Rugby programs.
// Rugby: import rugby/liquid
//
// Liquid is a template language created by Shopify, widely used in static
// site generators and web applications. This package provides a pure Go
// implementation of the Liquid template language.
//
// Example:
//
//	result = liquid.render("Hello, {{ name }}!", {name: "World"})!
//	# => "Hello, World!"
//
//	tmpl = liquid.parse("{% for item in items %}{{ item }}{% endfor %}")!
//	result = tmpl.render({items: [1, 2, 3]})!
//	# => "123"
package liquid

import "reflect"

// Template represents a parsed Liquid template.
type Template struct {
	ast *templateAST
}

// Parse parses a Liquid template string and returns a Template.
// Returns an error if the template contains syntax errors.
// Ruby: liquid.parse(source)
func Parse(source string) (*Template, error) {
	p := newParser(source)
	ast, err := p.parse()
	if err != nil {
		return nil, err
	}
	return &Template{ast: ast}, nil
}

// MustParse parses a Liquid template string and panics on error.
// Ruby: liquid.must_parse(source)
func MustParse(source string) *Template {
	tmpl, err := Parse(source)
	if err != nil {
		panic(err)
	}
	return tmpl
}

// Render executes the template with the given data and returns the result.
// Data can be a map or struct containing template variables.
// Ruby: tmpl.render(data)
func (t *Template) Render(data any) (string, error) {
	converted := convertToStringAny(data)
	eval := newEvaluator(converted)
	return eval.evaluate(t.ast)
}

// Render parses and renders a template in one step.
// This is a convenience function for simple one-off templates.
// Ruby: liquid.render(source, data)
func Render(source string, data any) (string, error) {
	tmpl, err := Parse(source)
	if err != nil {
		return "", err
	}
	return tmpl.Render(data)
}

// MustRender parses and renders a template, panicking on any error.
// Ruby: liquid.must_render(source, data)
func MustRender(source string, data any) string {
	result, err := Render(source, data)
	if err != nil {
		panic(err)
	}
	return result
}

// convertToStringAny converts various map types to map[string]any.
func convertToStringAny(data any) map[string]any {
	if data == nil {
		return make(map[string]any)
	}

	// If it's already map[string]any, return as-is
	if m, ok := data.(map[string]any); ok {
		return m
	}

	// Use reflection to convert other map types
	rv := reflect.ValueOf(data)
	if rv.Kind() == reflect.Map {
		result := make(map[string]any)
		iter := rv.MapRange()
		for iter.Next() {
			key := toString(iter.Key().Interface())
			val := iter.Value().Interface()
			// Recursively convert nested maps
			if reflect.TypeOf(val) != nil && reflect.ValueOf(val).Kind() == reflect.Map {
				result[key] = convertToStringAny(val)
			} else {
				result[key] = val
			}
		}
		return result
	}

	return make(map[string]any)
}
