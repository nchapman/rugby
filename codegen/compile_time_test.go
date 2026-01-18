package codegen

import (
	"strings"
	"testing"

	"github.com/nchapman/rugby/ast"
)

// --- Registry Tests ---

func TestCompileTimeRegistry_LiquidCompile(t *testing.T) {
	// Test that liquid.compile is registered
	call := &ast.CallExpr{
		Func: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "liquid"},
			Sel: "compile",
		},
	}

	gen := New()
	handler, method, ok := gen.getCompileTimeHandler(call)

	if !ok {
		t.Fatal("expected liquid.compile to be registered")
	}
	if method != "compile" {
		t.Errorf("expected method 'compile', got %q", method)
	}
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestCompileTimeRegistry_LiquidCompileFile(t *testing.T) {
	// Test that liquid.compile_file is registered
	call := &ast.CallExpr{
		Func: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "liquid"},
			Sel: "compile_file",
		},
	}

	gen := New()
	handler, method, ok := gen.getCompileTimeHandler(call)

	if !ok {
		t.Fatal("expected liquid.compile_file to be registered")
	}
	if method != "compile_file" {
		t.Errorf("expected method 'compile_file', got %q", method)
	}
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestCompileTimeRegistry_UnknownPackage(t *testing.T) {
	// Test that unknown packages are not matched
	call := &ast.CallExpr{
		Func: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "unknown"},
			Sel: "compile",
		},
	}

	gen := New()
	_, _, ok := gen.getCompileTimeHandler(call)

	if ok {
		t.Fatal("expected unknown.compile to NOT be registered")
	}
}

func TestCompileTimeRegistry_UnknownMethod(t *testing.T) {
	// Test that unknown methods on known packages are not matched
	call := &ast.CallExpr{
		Func: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "liquid"},
			Sel: "parse", // parse is runtime, not compile-time
		},
	}

	gen := New()
	_, _, ok := gen.getCompileTimeHandler(call)

	if ok {
		t.Fatal("expected liquid.parse to NOT be registered as compile-time handler")
	}
}

func TestCompileTimeRegistry_NonSelectorCall(t *testing.T) {
	// Test that non-selector calls (e.g., function calls) don't match
	call := &ast.CallExpr{
		Func: &ast.Ident{Name: "compile"},
	}

	gen := New()
	_, _, ok := gen.getCompileTimeHandler(call)

	if ok {
		t.Fatal("expected bare 'compile' call to NOT match")
	}
}

// --- Liquid Compile Handler Tests ---

func TestLiquidCompile_BasicText(t *testing.T) {
	input := `import "rugby/liquid"
const HELLO = liquid.compile("Hello, World!")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `var HELLO = liquid.CompiledTemplate{`)
	assertContains(t, output, `Render: func(data map[string]any) (string, error)`)
	assertContains(t, output, `buf.WriteString("Hello, World!")`)
	assertContains(t, output, `return buf.String(), nil`)
}

func TestLiquidCompile_VariableInterpolation(t *testing.T) {
	input := `import "rugby/liquid"
const GREETING = liquid.compile("Hello, {{ name }}!")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `var GREETING = liquid.CompiledTemplate{`)
	assertContains(t, output, `ctx := liquid.NewContext(data)`)
	assertContains(t, output, `buf.WriteString("Hello, ")`)
	assertContains(t, output, `buf.WriteString(liquid.ToString(ctx.Get("name")))`)
	assertContains(t, output, `buf.WriteString("!")`)
}

func TestLiquidCompile_Filter(t *testing.T) {
	input := `import "rugby/liquid"
const UPPER = liquid.compile("{{ name | upcase }}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `liquid.FilterUpcase(ctx.Get("name"))`)
}

func TestLiquidCompile_FilterChain(t *testing.T) {
	input := `import "rugby/liquid"
const CHAIN = liquid.compile("{{ name | upcase | downcase }}")
`
	output := compileRelaxed(t, input)

	// Filter chain: downcase(upcase(name))
	assertContains(t, output, `liquid.FilterDowncase(liquid.FilterUpcase(ctx.Get("name")))`)
}

func TestLiquidCompile_IfElse(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{% if show %}Yes{% else %}No{% endif %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `if liquid.ToBool(ctx.Get("show")) {`)
	assertContains(t, output, `buf.WriteString("Yes")`)
	assertContains(t, output, `} else {`)
	assertContains(t, output, `buf.WriteString("No")`)
}

func TestLiquidCompile_ForLoop(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{% for item in items %}{{ item }}{% endfor %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `liquid.ToSlice(ctx.Get("items"))`)
	assertContains(t, output, `for _i, _item := range`)
	assertContains(t, output, `ctx.Set("item", _item)`)
	assertContains(t, output, `ctx.SetForloop(_i,`)
}

func TestLiquidCompile_ForLoopWithElse(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{% for item in items %}{{ item }}{% else %}Empty{% endfor %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `if len(`)
	assertContains(t, output, `) == 0 {`)
	assertContains(t, output, `buf.WriteString("Empty")`)
	assertContains(t, output, `} else {`)
}

func TestLiquidCompile_Unless(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{% unless hidden %}Visible{% endunless %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `if !liquid.ToBool(ctx.Get("hidden")) {`)
	assertContains(t, output, `buf.WriteString("Visible")`)
}

func TestLiquidCompile_CaseWhen(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{% case x %}{% when 1 %}One{% when 2 %}Two{% else %}Other{% endcase %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `liquid.CompareValues(`)
	assertContains(t, output, `"=="`)
	assertContains(t, output, `buf.WriteString("One")`)
	assertContains(t, output, `buf.WriteString("Two")`)
	assertContains(t, output, `buf.WriteString("Other")`)
}

func TestLiquidCompile_Assign(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{% assign x = 42 %}{{ x }}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `ctx.Set("x", 42)`)
	assertContains(t, output, `ctx.Get("x")`)
}

func TestLiquidCompile_Capture(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{% capture greeting %}Hello{% endcapture %}{{ greeting }}")
`
	output := compileRelaxed(t, input)

	// Check for capture pattern: save, reset, capture, restore
	assertContains(t, output, `:= buf.String()`)       // save
	assertContains(t, output, `buf = strings.Builder`) // reset for capture
	assertContains(t, output, `ctx.Set("greeting",`)   // set captured value
}

func TestLiquidCompile_Comment(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("A{% comment %}ignored{% endcomment %}B")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `buf.WriteString("A")`)
	assertContains(t, output, `buf.WriteString("B")`)
	assertNotContains(t, output, `ignored`)
}

func TestLiquidCompile_Raw(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{% raw %}{{ not interpolated }}{% endraw %}")
`
	output := compileRelaxed(t, input)

	// Raw content should be output as literal string
	assertContains(t, output, `buf.WriteString("{{ not interpolated }}")`)
}

func TestLiquidCompile_DotAccess(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{{ user.name }}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `liquid.GetProperty(ctx.Get("user"), "name")`)
}

func TestLiquidCompile_IndexAccess(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{{ items[0] }}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `liquid.GetIndex(ctx.Get("items"), 0)`)
}

func TestLiquidCompile_BinaryExpr(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{% if x == 1 %}yes{% endif %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `liquid.CompareValues(ctx.Get("x"), "==", 1)`)
}

func TestLiquidCompile_AndOr(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{% if a and b %}both{% endif %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `liquid.ToBool(ctx.Get("a")) && liquid.ToBool(ctx.Get("b"))`)
}

func TestLiquidCompile_Range(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{% for i in (1..3) %}{{ i }}{% endfor %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `liquid.MakeRange(1, 3)`)
}

// --- Liquid Compile Handler: All Filters ---

func TestLiquidCompile_AllFilters(t *testing.T) {
	tests := []struct {
		filter   string
		expected string
	}{
		{"upcase", "liquid.FilterUpcase"},
		{"downcase", "liquid.FilterDowncase"},
		{"capitalize", "liquid.FilterCapitalize"},
		{"strip", "liquid.FilterStrip"},
		{"escape", "liquid.FilterEscape"},
		{"first", "liquid.FilterFirst"},
		{"last", "liquid.FilterLast"},
		{"size", "liquid.FilterSize"},
		{"reverse", "liquid.FilterReverse"},
	}

	for _, tt := range tests {
		t.Run(tt.filter, func(t *testing.T) {
			input := `import "rugby/liquid"
const TMPL = liquid.compile("{{ x | ` + tt.filter + ` }}")
`
			output := compileRelaxed(t, input)
			assertContains(t, output, tt.expected)
		})
	}
}

func TestLiquidCompile_FilterWithArg(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{{ items | join: \", \" }}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `liquid.FilterJoin(ctx.Get("items"),`)
}

func TestLiquidCompile_DefaultFilter(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{{ missing | default: \"N/A\" }}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `liquid.FilterDefault(ctx.Get("missing"),`)
}

// --- Error Cases ---

func TestLiquidCompile_ErrorMissingArgument(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile()
`
	errs := compileWithErrors(t, input)

	if len(errs) == 0 {
		t.Fatal("expected error for missing argument")
	}

	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "requires a template string argument") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'requires a template string argument' error, got: %v", errs)
	}
}

func TestLiquidCompile_ErrorInvalidSyntax(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{% if %}")
`
	errs := compileWithErrors(t, input)

	if len(errs) == 0 {
		t.Fatal("expected error for invalid template syntax")
	}

	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "liquid template syntax error") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'liquid template syntax error', got: %v", errs)
	}
}

// --- Import Tests ---

func TestLiquidCompile_ImportsLiquid(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("Hello")
`
	output := compileRelaxed(t, input)

	// Should import the liquid package
	assertContains(t, output, `"github.com/nchapman/rugby/stdlib/liquid"`)
}

func TestLiquidCompile_ImportsStrings(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("Hello")
`
	output := compileRelaxed(t, input)

	// Should import strings for strings.Builder
	assertContains(t, output, `"strings"`)
}

// --- Context Scope Tests ---

func TestLiquidCompile_ForLoopContextScope(t *testing.T) {
	input := `import "rugby/liquid"
const TMPL = liquid.compile("{% for i in items %}{{ i }}{% endfor %}")
`
	output := compileRelaxed(t, input)

	// Should save and restore context around for loop
	assertContains(t, output, `:= ctx`)     // save outer context
	assertContains(t, output, `ctx.Push()`) // push new scope
	assertContains(t, output, `ctx =`)      // restore (appears multiple times)
}
