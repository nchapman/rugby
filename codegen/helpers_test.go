package codegen

// Test helpers for codegen package tests.
//
// Available compile helpers:
//   - compile(t, input)              - strict: fails on parser/semantic errors
//   - compileRelaxed(t, input)       - ignores semantic errors (for testing codegen patterns)
//   - compileExpectError(t, input)   - returns Generate() error (for testing codegen errors)
//   - compileWithErrors(t, input)    - returns collected errors from gen.Errors()
//   - compileWithLineDirectives(...) - with source file for line directive tests
//
// Available assertions:
//   - assertContains(t, output, substr)
//   - assertNotContains(t, output, substr)

import (
	"strings"
	"testing"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/parser"
	"github.com/nchapman/rugby/semantic"
)

// compile runs the full compilation pipeline and fails on any errors.
// Use this for tests where the input should be valid Rugby code.
func compile(t *testing.T, input string) string {
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
	errs := analyzer.Analyze(program)
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("semantic error: %s", err)
		}
		t.FailNow()
	}

	typeInfo := &testTypeInfoAdapter{analyzer: analyzer}
	gen := New(WithTypeInfo(typeInfo))
	output, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	return output
}

// compileRelaxed runs codegen but ignores semantic errors.
// Use this for tests that use undefined variables to test codegen output patterns.
func compileRelaxed(t *testing.T, input string) string {
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
	gen := New(WithTypeInfo(typeInfo))
	output, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	return output
}

// compileExpectError runs codegen and returns the Generate() error.
// Use this for tests that expect codegen to fail (e.g., invalid range types).
func compileExpectError(t *testing.T, input string) error {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Skipf("parser error: %v", p.Errors())
	}

	analyzer := semantic.NewAnalyzer()
	_ = analyzer.Analyze(program) // Ignore semantic errors

	typeInfo := &testTypeInfoAdapter{analyzer: analyzer}
	gen := New(WithTypeInfo(typeInfo))
	_, err := gen.Generate(program)
	return err
}

// compileWithErrors runs codegen and returns collected errors from gen.Errors().
// Use this for tests checking codegen error collection (e.g., module method conflicts).
func compileWithErrors(t *testing.T, input string) []error {
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
	gen := New(WithTypeInfo(typeInfo))
	_, _ = gen.Generate(program)
	return gen.Errors()
}

// compileWithLineDirectives runs codegen with line directive emission enabled.
// Use this for tests that verify //line directive generation.
func compileWithLineDirectives(t *testing.T, input string, sourceFile string) string {
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
	errs := analyzer.Analyze(program)
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("semantic error: %s", err)
		}
		t.FailNow()
	}

	typeInfo := &testTypeInfoAdapter{analyzer: analyzer}
	gen := New(WithSourceFile(sourceFile), WithTypeInfo(typeInfo))
	output, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	return output
}

// assertContains fails if output does not contain substr.
func assertContains(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("expected output to contain %q, got:\n%s", substr, output)
	}
}

// assertNotContains fails if output contains substr.
func assertNotContains(t *testing.T, output, substr string) {
	t.Helper()
	if strings.Contains(output, substr) {
		t.Errorf("expected output NOT to contain %q, got:\n%s", substr, output)
	}
}

// testTypeInfoAdapter adapts semantic.Analyzer to implement TypeInfo for tests.
type testTypeInfoAdapter struct {
	analyzer *semantic.Analyzer
}

func (t *testTypeInfoAdapter) GetTypeKind(node ast.Node) TypeKind {
	typ := t.analyzer.GetType(node)
	if typ == nil {
		return TypeUnknown
	}
	switch typ.Kind {
	case semantic.TypeInt:
		return TypeInt
	case semantic.TypeInt64:
		return TypeInt64
	case semantic.TypeFloat:
		return TypeFloat
	case semantic.TypeBool:
		return TypeBool
	case semantic.TypeString:
		return TypeString
	case semantic.TypeNil:
		return TypeNil
	case semantic.TypeArray:
		return TypeArray
	case semantic.TypeMap:
		return TypeMap
	case semantic.TypeClass:
		return TypeClass
	case semantic.TypeOptional:
		return TypeOptional
	case semantic.TypeAny:
		return TypeAny
	default:
		return TypeUnknown
	}
}

func (t *testTypeInfoAdapter) GetGoType(node ast.Node) string {
	typ := t.analyzer.GetType(node)
	if typ == nil {
		return ""
	}
	return typ.GoType()
}

func (t *testTypeInfoAdapter) GetRugbyType(node ast.Node) string {
	typ := t.analyzer.GetType(node)
	if typ == nil {
		return ""
	}
	return typ.String()
}

func (t *testTypeInfoAdapter) GetSelectorKind(node ast.Node) ast.SelectorKind {
	if sel, ok := node.(*ast.SelectorExpr); ok && sel.ResolvedKind != ast.SelectorUnknown {
		return sel.ResolvedKind
	}
	return t.analyzer.GetSelectorKind(node)
}

func (t *testTypeInfoAdapter) GetElementType(node ast.Node) string {
	typ := t.analyzer.GetType(node)
	if typ == nil || typ.Elem == nil {
		return ""
	}
	return typ.Elem.String()
}

func (t *testTypeInfoAdapter) GetKeyValueTypes(node ast.Node) (string, string) {
	typ := t.analyzer.GetType(node)
	if typ == nil || typ.Kind != semantic.TypeMap {
		return "", ""
	}
	keyType := ""
	valueType := ""
	if typ.KeyType != nil {
		keyType = typ.KeyType.String()
	}
	if typ.ValueType != nil {
		valueType = typ.ValueType.String()
	}
	return keyType, valueType
}

func (t *testTypeInfoAdapter) GetTupleTypes(node ast.Node) []string {
	typ := t.analyzer.GetType(node)
	if typ == nil {
		return nil
	}
	if typ.Kind == semantic.TypeTuple && len(typ.Elements) > 0 {
		result := make([]string, len(typ.Elements))
		for i, elem := range typ.Elements {
			result[i] = elem.String()
		}
		return result
	}
	if typ.Kind == semantic.TypeOptional && typ.Elem != nil {
		return []string{typ.Elem.String(), "Bool"}
	}
	return nil
}

func (t *testTypeInfoAdapter) IsVariableUsedAt(node ast.Node, name string) bool {
	return t.analyzer.IsVariableUsedAt(node, name)
}

func (t *testTypeInfoAdapter) IsDeclaration(node ast.Node) bool {
	return t.analyzer.IsDeclaration(node)
}

func (t *testTypeInfoAdapter) GetFieldType(className, fieldName string) string {
	return t.analyzer.GetFieldType(className, fieldName)
}

func (t *testTypeInfoAdapter) IsClass(typeName string) bool {
	return t.analyzer.IsClass(typeName)
}

func (t *testTypeInfoAdapter) IsInterface(typeName string) bool {
	return t.analyzer.IsInterface(typeName)
}

func (t *testTypeInfoAdapter) IsNoArgFunction(name string) bool {
	return t.analyzer.IsNoArgFunction(name)
}

func (t *testTypeInfoAdapter) IsPublicClass(className string) bool {
	return t.analyzer.IsPublicClass(className)
}

func (t *testTypeInfoAdapter) HasAccessor(className, fieldName string) bool {
	return t.analyzer.HasAccessor(className, fieldName)
}

func (t *testTypeInfoAdapter) GetInterfaceMethodNames(interfaceName string) []string {
	return t.analyzer.GetInterfaceMethodNames(interfaceName)
}

func (t *testTypeInfoAdapter) GetAllInterfaceNames() []string {
	return t.analyzer.GetAllInterfaceNames()
}

func (t *testTypeInfoAdapter) GetAllModuleNames() []string {
	return t.analyzer.GetAllModuleNames()
}

func (t *testTypeInfoAdapter) GetModuleMethodNames(moduleName string) []string {
	return t.analyzer.GetModuleMethodNames(moduleName)
}

func (t *testTypeInfoAdapter) IsModuleMethod(className, methodName string) bool {
	return t.analyzer.IsModuleMethod(className, methodName)
}

func (t *testTypeInfoAdapter) GetConstructorParamCount(className string) int {
	return t.analyzer.GetConstructorParamCount(className)
}

func (t *testTypeInfoAdapter) GetConstructorParams(className string) [][2]string {
	return t.analyzer.GetConstructorParams(className)
}
