package codegen

import (
	"strings"
	"testing"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/parser"
	"github.com/nchapman/rugby/semantic"
)

// --- Registry Tests ---

func TestCompileTimeRegistry_TemplateCompile(t *testing.T) {
	// Test that template.compile is registered
	call := &ast.CallExpr{
		Func: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "template"},
			Sel: "compile",
		},
	}

	gen := New()
	handler, method, ok := gen.getCompileTimeHandler(call)

	if !ok {
		t.Fatal("expected template.compile to be registered")
	}
	if method != "compile" {
		t.Errorf("expected method 'compile', got %q", method)
	}
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestCompileTimeRegistry_TemplateCompileFile(t *testing.T) {
	// Test that template.compile_file is registered
	call := &ast.CallExpr{
		Func: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "template"},
			Sel: "compile_file",
		},
	}

	gen := New()
	handler, method, ok := gen.getCompileTimeHandler(call)

	if !ok {
		t.Fatal("expected template.compile_file to be registered")
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
			X:   &ast.Ident{Name: "template"},
			Sel: "parse", // parse is runtime, not compile-time
		},
	}

	gen := New()
	_, _, ok := gen.getCompileTimeHandler(call)

	if ok {
		t.Fatal("expected template.parse to NOT be registered as compile-time handler")
	}
}

func TestCompileTimeRegistry_TemplateCompileDir(t *testing.T) {
	call := &ast.CallExpr{
		Func: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "template"},
			Sel: "compile_dir",
		},
	}

	gen := New()
	handler, method, ok := gen.getCompileTimeHandler(call)

	if !ok {
		t.Fatal("expected template.compile_dir to be registered")
	}
	if method != "compile_dir" {
		t.Errorf("expected method 'compile_dir', got %q", method)
	}
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestCompileTimeRegistry_TemplateCompileGlob(t *testing.T) {
	call := &ast.CallExpr{
		Func: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "template"},
			Sel: "compile_glob",
		},
	}

	gen := New()
	handler, method, ok := gen.getCompileTimeHandler(call)

	if !ok {
		t.Fatal("expected template.compile_glob to be registered")
	}
	if method != "compile_glob" {
		t.Errorf("expected method 'compile_glob', got %q", method)
	}
	if handler == nil {
		t.Fatal("expected non-nil handler")
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

// --- Template Compile Handler Tests ---

func TestTemplateCompile_BasicText(t *testing.T) {
	input := `import "rugby/template"
const HELLO = template.compile("Hello, World!")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `var HELLO = template.CompiledTemplate{`)
	assertContains(t, output, `Render: func(data map[string]any) (string, error)`)
	assertContains(t, output, `buf.WriteString("Hello, World!")`)
	assertContains(t, output, `return buf.String(), nil`)
}

func TestTemplateCompile_VariableInterpolation(t *testing.T) {
	input := `import "rugby/template"
const GREETING = template.compile("Hello, {{ name }}!")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `var GREETING = template.CompiledTemplate{`)
	assertContains(t, output, `ctx := template.NewContext(data)`)
	assertContains(t, output, `buf.WriteString("Hello, ")`)
	assertContains(t, output, `buf.WriteString(template.ToString(ctx.Get("name")))`)
	assertContains(t, output, `buf.WriteString("!")`)
}

func TestTemplateCompile_Filter(t *testing.T) {
	input := `import "rugby/template"
const UPPER = template.compile("{{ name | upcase }}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `template.FilterUpcase(ctx.Get("name"))`)
}

func TestTemplateCompile_FilterChain(t *testing.T) {
	input := `import "rugby/template"
const CHAIN = template.compile("{{ name | upcase | downcase }}")
`
	output := compileRelaxed(t, input)

	// Filter chain: downcase(upcase(name))
	assertContains(t, output, `template.FilterDowncase(template.FilterUpcase(ctx.Get("name")))`)
}

func TestTemplateCompile_IfElse(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{% if show %}Yes{% else %}No{% endif %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `if template.ToBool(ctx.Get("show")) {`)
	assertContains(t, output, `buf.WriteString("Yes")`)
	assertContains(t, output, `} else {`)
	assertContains(t, output, `buf.WriteString("No")`)
}

func TestTemplateCompile_ForLoop(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{% for item in items %}{{ item }}{% endfor %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `template.ToSlice(ctx.Get("items"))`)
	assertContains(t, output, `for _i, _item := range`)
	assertContains(t, output, `ctx.Set("item", _item)`)
	assertContains(t, output, `ctx.SetForloop(_i,`)
}

func TestTemplateCompile_ForLoopWithElse(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{% for item in items %}{{ item }}{% else %}Empty{% endfor %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `if len(`)
	assertContains(t, output, `) == 0 {`)
	assertContains(t, output, `buf.WriteString("Empty")`)
	assertContains(t, output, `} else {`)
}

func TestTemplateCompile_Unless(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{% unless hidden %}Visible{% endunless %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `if !template.ToBool(ctx.Get("hidden")) {`)
	assertContains(t, output, `buf.WriteString("Visible")`)
}

func TestTemplateCompile_CaseWhen(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{% case x %}{% when 1 %}One{% when 2 %}Two{% else %}Other{% endcase %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `template.CompareValues(`)
	assertContains(t, output, `"=="`)
	assertContains(t, output, `buf.WriteString("One")`)
	assertContains(t, output, `buf.WriteString("Two")`)
	assertContains(t, output, `buf.WriteString("Other")`)
}

func TestTemplateCompile_Assign(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{% assign x = 42 %}{{ x }}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `ctx.Set("x", 42)`)
	assertContains(t, output, `ctx.Get("x")`)
}

func TestTemplateCompile_Capture(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{% capture greeting %}Hello{% endcapture %}{{ greeting }}")
`
	output := compileRelaxed(t, input)

	// Check for capture pattern: save, reset, capture, restore
	assertContains(t, output, `:= buf.String()`)       // save
	assertContains(t, output, `buf = strings.Builder`) // reset for capture
	assertContains(t, output, `ctx.Set("greeting",`)   // set captured value
}

func TestTemplateCompile_Comment(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("A{% comment %}ignored{% endcomment %}B")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `buf.WriteString("A")`)
	assertContains(t, output, `buf.WriteString("B")`)
	assertNotContains(t, output, `ignored`)
}

func TestTemplateCompile_Raw(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{% raw %}{{ not interpolated }}{% endraw %}")
`
	output := compileRelaxed(t, input)

	// Raw content should be output as literal string
	assertContains(t, output, `buf.WriteString("{{ not interpolated }}")`)
}

func TestTemplateCompile_DotAccess(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{{ user.name }}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `template.GetProperty(ctx.Get("user"), "name")`)
}

func TestTemplateCompile_IndexAccess(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{{ items[0] }}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `template.GetIndex(ctx.Get("items"), 0)`)
}

func TestTemplateCompile_BinaryExpr(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{% if x == 1 %}yes{% endif %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `template.CompareValues(ctx.Get("x"), "==", 1)`)
}

func TestTemplateCompile_AndOr(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{% if a and b %}both{% endif %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `template.ToBool(ctx.Get("a")) && template.ToBool(ctx.Get("b"))`)
}

func TestTemplateCompile_Range(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{% for i in (1..3) %}{{ i }}{% endfor %}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `template.MakeRange(1, 3)`)
}

// --- Template Compile Handler: All Filters ---

func TestTemplateCompile_AllFilters(t *testing.T) {
	tests := []struct {
		filter   string
		expected string
	}{
		{"upcase", "template.FilterUpcase"},
		{"downcase", "template.FilterDowncase"},
		{"capitalize", "template.FilterCapitalize"},
		{"strip", "template.FilterStrip"},
		{"escape", "template.FilterEscape"},
		{"first", "template.FilterFirst"},
		{"last", "template.FilterLast"},
		{"size", "template.FilterSize"},
		{"reverse", "template.FilterReverse"},
	}

	for _, tt := range tests {
		t.Run(tt.filter, func(t *testing.T) {
			input := `import "rugby/template"
const TMPL = template.compile("{{ x | ` + tt.filter + ` }}")
`
			output := compileRelaxed(t, input)
			assertContains(t, output, tt.expected)
		})
	}
}

func TestTemplateCompile_FilterWithArg(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{{ items | join: \", \" }}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `template.FilterJoin(ctx.Get("items"),`)
}

func TestTemplateCompile_DefaultFilter(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{{ missing | default: \"N/A\" }}")
`
	output := compileRelaxed(t, input)

	assertContains(t, output, `template.FilterDefault(ctx.Get("missing"),`)
}

// --- Error Cases ---

func TestTemplateCompile_ErrorMissingArgument(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile()
`
	errs := compileWithErrors(t, input)

	if len(errs) == 0 {
		t.Fatal("expected error for missing argument")
	}

	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "requires a string argument") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'requires a string argument' error, got: %v", errs)
	}
}

func TestTemplateCompile_ErrorInvalidSyntax(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{% if %}")
`
	errs := compileWithErrors(t, input)

	if len(errs) == 0 {
		t.Fatal("expected error for invalid template syntax")
	}

	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "template syntax error") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'template syntax error', got: %v", errs)
	}
}

// --- Import Tests ---

func TestTemplateCompile_ImportsTemplate(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("Hello")
`
	output := compileRelaxed(t, input)

	// Should import the template package
	assertContains(t, output, `"github.com/nchapman/rugby/stdlib/template"`)
}

func TestTemplateCompile_ImportsStrings(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("Hello")
`
	output := compileRelaxed(t, input)

	// Should import strings for strings.Builder
	assertContains(t, output, `"strings"`)
}

// --- Context Scope Tests ---

func TestTemplateCompile_ForLoopContextScope(t *testing.T) {
	input := `import "rugby/template"
const TMPL = template.compile("{% for i in items %}{{ i }}{% endfor %}")
`
	output := compileRelaxed(t, input)

	// Should save and restore context around for loop
	assertContains(t, output, `:= ctx`)     // save outer context
	assertContains(t, output, `ctx.Push()`) // push new scope
	assertContains(t, output, `ctx =`)      // restore (appears multiple times)
}

// --- Template compile_dir Tests ---

func TestTemplateCompileDir_GeneratesMap(t *testing.T) {
	// Uses testdata/templates/ which contains greeting.tmpl and i18n/*.tmpl
	// Tests run from the codegen package directory, so use "." as source dir
	input := `import "rugby/template"
const TEMPLATES = template.compile_dir("testdata/templates/i18n/")
`
	output := compileRelaxedWithSourceDir(t, input, ".")

	// Should generate map type with pointer values (so MustRender pointer receiver works)
	assertContains(t, output, `var TEMPLATES = map[string]*template.CompiledTemplate{`)

	// Should have keys for both templates (sorted alphabetically)
	assertContains(t, output, `"en.tmpl": {`)
	assertContains(t, output, `"fr.tmpl": {`)

	// Each entry should have a Render function
	assertContains(t, output, `Render: func(data map[string]any) (string, error)`)
}

func TestTemplateCompileDir_TemplateContent(t *testing.T) {
	input := `import "rugby/template"
const TEMPLATES = template.compile_dir("testdata/templates/i18n/")
`
	output := compileRelaxedWithSourceDir(t, input, ".")

	// en.tmpl contains: Hello, {{ name }}!
	assertContains(t, output, `buf.WriteString("Hello, ")`)
	assertContains(t, output, `ctx.Get("name")`)

	// fr.tmpl contains: Bonjour, {{ name }}!
	assertContains(t, output, `buf.WriteString("Bonjour, ")`)
}

func TestTemplateCompileDir_ErrorNotDirectory(t *testing.T) {
	input := `import "rugby/template"
const TEMPLATES = template.compile_dir("testdata/templates/greeting.tmpl")
`
	errs := compileWithErrorsAndSourceDir(t, input, ".")

	if len(errs) == 0 {
		t.Fatal("expected error for non-directory path")
	}

	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "is not a directory") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'is not a directory' error, got: %v", errs)
	}
}

func TestTemplateCompileDir_ErrorNoFiles(t *testing.T) {
	// Use testdata/empty_dir which has no template files
	input := `import "rugby/template"
const TEMPLATES = template.compile_dir("testdata/empty_dir/")
`
	errs := compileWithErrorsAndSourceDir(t, input, ".")

	if len(errs) == 0 {
		t.Fatal("expected error for directory with no template files")
	}

	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "no template files") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'no template files' error, got: %v", errs)
	}
}

// --- Template compile_glob Tests ---

func TestTemplateCompileGlob_GeneratesMap(t *testing.T) {
	input := `import "rugby/template"
const TEMPLATES = template.compile_glob("testdata/templates/i18n/*.tmpl")
`
	output := compileRelaxedWithSourceDir(t, input, ".")

	// Should generate map type with pointer values
	assertContains(t, output, `var TEMPLATES = map[string]*template.CompiledTemplate{`)

	// Should have keys for matched templates
	assertContains(t, output, `"en.tmpl": {`)
	assertContains(t, output, `"fr.tmpl": {`)
}

func TestTemplateCompileGlob_NestedPattern(t *testing.T) {
	// Note: Go's filepath.Glob doesn't support ** for recursive matching.
	// ** matches exactly one directory level, so "templates/**/*.tmpl"
	// matches "templates/SUBDIR/*.tmpl" but not "templates/*.tmpl"
	input := `import "rugby/template"
const TEMPLATES = template.compile_glob("testdata/templates/*/*.tmpl")
`
	output := compileRelaxedWithSourceDir(t, input, ".")

	// Should match .tmpl files one level deep (with pointer values)
	assertContains(t, output, `var TEMPLATES = map[string]*template.CompiledTemplate{`)

	// Should include files from subdirectories with relative paths
	assertContains(t, output, `"i18n/en.tmpl": {`)
	assertContains(t, output, `"i18n/fr.tmpl": {`)
}

func TestTemplateCompileGlob_ErrorNoMatches(t *testing.T) {
	input := `import "rugby/template"
const TEMPLATES = template.compile_glob("testdata/nonexistent/*.tmpl")
`
	errs := compileWithErrorsAndSourceDir(t, input, ".")

	if len(errs) == 0 {
		t.Fatal("expected error for no matching files")
	}

	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "no files matched pattern") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'no files matched pattern' error, got: %v", errs)
	}
}

// --- Test helpers for source directory ---

// compileRelaxedWithSourceDir compiles code with a source directory set.
// The sourceDir parameter is preserved for flexibility, though current tests use ".".
func compileRelaxedWithSourceDir(t *testing.T, input string, sourceDir string) string { //nolint:unparam
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := semantic.NewAnalyzer()
	_ = analyzer.Analyze(program) // Ignore semantic errors

	typeInfo := &testTypeInfoAdapter{analyzer: analyzer}
	// Use sourceDir + "/dummy.rg" to set the source directory
	gen := New(WithSourceFile(sourceDir+"/dummy.rg"), WithTypeInfo(typeInfo))
	output, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	return output
}

// compileWithErrorsAndSourceDir compiles code and returns errors, with a source directory set.
func compileWithErrorsAndSourceDir(t *testing.T, input string, sourceDir string) []error { //nolint:unparam
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	analyzer := semantic.NewAnalyzer()
	_ = analyzer.Analyze(program) // Ignore semantic errors

	typeInfo := &testTypeInfoAdapter{analyzer: analyzer}
	gen := New(WithSourceFile(sourceDir+"/dummy.rg"), WithTypeInfo(typeInfo))
	_, _ = gen.Generate(program)
	return gen.Errors()
}
