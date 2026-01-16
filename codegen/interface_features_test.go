package codegen_test

import (
	"strings"
	"testing"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/codegen"
	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/parser"
	"github.com/nchapman/rugby/semantic"
)

// testTypeInfo adapts semantic.Analyzer to implement codegen.TypeInfo
type testTypeInfo struct {
	analyzer *semantic.Analyzer
}

func (t *testTypeInfo) GetTypeKind(node ast.Node) codegen.TypeKind {
	typ := t.analyzer.GetType(node)
	if typ == nil {
		return codegen.TypeUnknown
	}
	switch typ.Kind {
	case semantic.TypeInt:
		return codegen.TypeInt
	case semantic.TypeInt64:
		return codegen.TypeInt64
	case semantic.TypeFloat:
		return codegen.TypeFloat
	case semantic.TypeBool:
		return codegen.TypeBool
	case semantic.TypeString:
		return codegen.TypeString
	case semantic.TypeNil:
		return codegen.TypeNil
	case semantic.TypeArray:
		return codegen.TypeArray
	case semantic.TypeMap:
		return codegen.TypeMap
	case semantic.TypeClass:
		return codegen.TypeClass
	case semantic.TypeOptional:
		return codegen.TypeOptional
	case semantic.TypeChan:
		return codegen.TypeChannel
	case semantic.TypeAny:
		return codegen.TypeAny
	default:
		return codegen.TypeUnknown
	}
}

func (t *testTypeInfo) GetGoType(node ast.Node) string {
	typ := t.analyzer.GetType(node)
	if typ == nil {
		return ""
	}
	return typ.GoType()
}

func (t *testTypeInfo) GetRugbyType(node ast.Node) string {
	typ := t.analyzer.GetType(node)
	if typ == nil {
		return ""
	}
	return typ.String()
}

func (t *testTypeInfo) GetSelectorKind(node ast.Node) ast.SelectorKind {
	if sel, ok := node.(*ast.SelectorExpr); ok && sel.ResolvedKind != ast.SelectorUnknown {
		return sel.ResolvedKind
	}
	return t.analyzer.GetSelectorKind(node)
}

func (t *testTypeInfo) GetElementType(node ast.Node) string {
	typ := t.analyzer.GetType(node)
	if typ == nil || typ.Elem == nil {
		return ""
	}
	return typ.Elem.String()
}

func (t *testTypeInfo) GetKeyValueTypes(node ast.Node) (string, string) {
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

func (t *testTypeInfo) GetTupleTypes(node ast.Node) []string {
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

func (t *testTypeInfo) IsVariableUsedAt(node ast.Node, name string) bool {
	return t.analyzer.IsVariableUsedAt(node, name)
}

func (t *testTypeInfo) IsDeclaration(node ast.Node) bool {
	return t.analyzer.IsDeclaration(node)
}

func (t *testTypeInfo) GetFieldType(className, fieldName string) string {
	return t.analyzer.GetFieldType(className, fieldName)
}

func (t *testTypeInfo) IsClass(typeName string) bool {
	return t.analyzer.IsClass(typeName)
}

func (t *testTypeInfo) IsInterface(typeName string) bool {
	return t.analyzer.IsInterface(typeName)
}

func (t *testTypeInfo) IsNoArgFunction(name string) bool {
	return t.analyzer.IsNoArgFunction(name)
}

func (t *testTypeInfo) IsPublicClass(className string) bool {
	return t.analyzer.IsPublicClass(className)
}

func (t *testTypeInfo) HasAccessor(className, fieldName string) bool {
	return t.analyzer.HasAccessor(className, fieldName)
}

func (t *testTypeInfo) GetInterfaceMethodNames(interfaceName string) []string {
	return t.analyzer.GetInterfaceMethodNames(interfaceName)
}

func (t *testTypeInfo) GetAllInterfaceNames() []string {
	return t.analyzer.GetAllInterfaceNames()
}

func (t *testTypeInfo) GetAllModuleNames() []string {
	return t.analyzer.GetAllModuleNames()
}

func (t *testTypeInfo) GetModuleMethodNames(moduleName string) []string {
	return t.analyzer.GetModuleMethodNames(moduleName)
}

func (t *testTypeInfo) IsModuleMethod(className, methodName string) bool {
	return t.analyzer.IsModuleMethod(className, methodName)
}

func (t *testTypeInfo) GetConstructorParamCount(className string) int {
	return t.analyzer.GetConstructorParamCount(className)
}

func (t *testTypeInfo) GetConstructorParams(className string) [][2]string {
	return t.analyzer.GetConstructorParams(className)
}

// compileWithTypeInfo runs the full compilation pipeline.
// It ignores semantic errors since some tests use constructs not known to semantic analyzer.
func compileWithTypeInfo(t *testing.T, input string) string {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Run semantic analysis (required for codegen) but ignore semantic errors
	// since some tests use constructs not known to semantic analyzer
	analyzer := semantic.NewAnalyzer()
	_ = analyzer.Analyze(program)

	typeInfo := &testTypeInfo{analyzer: analyzer}
	g := codegen.New(codegen.WithTypeInfo(typeInfo))
	output, err := g.Generate(program)
	if err != nil {
		t.Fatalf("Codegen error: %v", err)
	}

	return output
}

func TestInterfaceInheritance(t *testing.T) {
	input := `
interface Reader
  def read -> String
end

interface Writer
  def write(data : String)
end

interface IO < Reader, Writer
  def close
end
`

	output := compileWithTypeInfo(t, input)

	// Check for interface embedding
	if !strings.Contains(output, "type IO interface {") {
		t.Error("Expected IO interface declaration")
	}
	if !strings.Contains(output, "\tReader") {
		t.Error("Expected Reader embedded in IO")
	}
	if !strings.Contains(output, "\tWriter") {
		t.Error("Expected Writer embedded in IO")
	}
}

func TestClassImplements(t *testing.T) {
	input := `
interface Speaker
  def speak -> String
end

interface Serializable
  def to_json -> String
end

class User implements Speaker, Serializable
  def initialize(@name : String)
  end

  def speak -> String
    @name
  end

  def to_json -> String
    @name
  end
end
`

	output := compileWithTypeInfo(t, input)

	// Check for compile-time conformance checks
	if !strings.Contains(output, "var _ Speaker = (*User)(nil)") {
		t.Error("Expected Speaker conformance check")
	}
	if !strings.Contains(output, "var _ Serializable = (*User)(nil)") {
		t.Error("Expected Serializable conformance check")
	}
}

func TestAnyKeyword(t *testing.T) {
	input := `
def log(thing : any)
  thing
end
`

	output := compileWithTypeInfo(t, input)

	// Check that 'any' maps to Go's any
	if !strings.Contains(output, "func log(thing any)") {
		t.Error("Expected 'any' to map to Go's any type")
	}
}

func TestIsAMethodCall(t *testing.T) {
	input := `
def check(obj : any) -> Bool
  obj.is_a?(String)
end
`

	output := compileWithTypeInfo(t, input)

	// Check for type assertion
	if !strings.Contains(output, "func() bool") {
		t.Error("Expected is_a? to compile to type assertion")
	}
	if !strings.Contains(output, "_, ok :=") {
		t.Error("Expected comma-ok idiom in is_a?")
	}
}

func TestAsMethodCall(t *testing.T) {
	input := `
def cast(obj : any) -> (String, Bool)
  obj.as(String)
end
`

	output := compileWithTypeInfo(t, input)

	// Check for type assertion with value
	if !strings.Contains(output, "v, ok :=") {
		t.Error("Expected value and ok from as()")
	}
}

func TestMessageMethodMapping(t *testing.T) {
	input := `
class CustomError
  def initialize(@msg : String)
  end

  def message -> String
    @msg
  end
end
`

	output := compileWithTypeInfo(t, input)

	// Check that message -> Error()
	if !strings.Contains(output, "func (c *CustomError) Error() string") {
		t.Error("Expected message to map to Error() method")
	}
}
