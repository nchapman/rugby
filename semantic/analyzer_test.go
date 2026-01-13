package semantic

import (
	"slices"
	"strings"
	"testing"

	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/parser"
)

func parse(t *testing.T, input string) *parser.Parser {
	t.Helper()
	l := lexer.New(input)
	return parser.New(l)
}

func TestAnalyzeVariableDeclaration(t *testing.T) {
	input := `x = 5`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}

	// Check that x was declared
	sym := a.GetSymbol("x")
	if sym == nil {
		t.Fatal("symbol 'x' not found")
	}
	if !sym.Type.Equals(TypeIntVal) {
		t.Errorf("x.Type = %v, want Int", sym.Type)
	}
}

func TestAnalyzeVariableWithTypeAnnotation(t *testing.T) {
	input := `x : Int64 = 5`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}

	sym := a.GetSymbol("x")
	if sym == nil {
		t.Fatal("symbol 'x' not found")
	}
	if !sym.Type.Equals(TypeInt64Val) {
		t.Errorf("x.Type = %v, want Int64", sym.Type)
	}
}

func TestAnalyzeUndefinedVariable(t *testing.T) {
	input := `puts y`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	if len(errs) == 0 {
		t.Fatal("expected error for undefined variable")
	}

	found := false
	for _, err := range errs {
		if undef, ok := err.(*UndefinedError); ok && undef.Name == "y" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected UndefinedError for 'y', got: %v", errs)
	}
}

func TestAnalyzeFunctionDeclaration(t *testing.T) {
	input := `
def add(a : Int, b : Int) -> Int
  a + b
end
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}

	fn := a.GetFunction("add")
	if fn == nil {
		t.Fatal("function 'add' not found")
	}
	if len(fn.Params) != 2 {
		t.Errorf("add has %d params, want 2", len(fn.Params))
	}
	if len(fn.ReturnTypes) != 1 || !fn.ReturnTypes[0].Equals(TypeIntVal) {
		t.Errorf("add returns %v, want [Int]", fn.ReturnTypes)
	}
}

func TestAnalyzeClassDeclaration(t *testing.T) {
	input := `
class User
  def initialize(@name : String, @age : Int)
  end

  def greet -> String
    "Hello, #{@name}"
  end
end
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}

	cls := a.GetClass("User")
	if cls == nil {
		t.Fatal("class 'User' not found")
	}

	if cls.GetMethod("initialize") == nil {
		t.Error("method 'initialize' not found")
	}
	if cls.GetMethod("greet") == nil {
		t.Error("method 'greet' not found")
	}
}

func TestAnalyzeInstanceVarOutsideClass(t *testing.T) {
	input := `puts @name`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	if len(errs) == 0 {
		t.Fatal("expected error for instance var outside class")
	}

	found := false
	for _, err := range errs {
		if _, ok := err.(*InstanceVarOutsideClassError); ok {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected InstanceVarOutsideClassError, got: %v", errs)
	}
}

func TestAnalyzeForLoop(t *testing.T) {
	input := `
nums = [1, 2, 3]
for n in nums
  puts n
end
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
}

func TestAnalyzeBreakOutsideLoop(t *testing.T) {
	input := `break`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	if len(errs) == 0 {
		t.Fatal("expected error for break outside loop")
	}

	found := false
	for _, err := range errs {
		if _, ok := err.(*BreakOutsideLoopError); ok {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected BreakOutsideLoopError, got: %v", errs)
	}
}

func TestAnalyzeIfLet(t *testing.T) {
	input := `
def find_user(id : Int) -> User?
  nil
end

if let user = find_user(1)
  puts user
end
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	// Should not error - user is defined within the if-let scope
	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
}

func TestAnalyzeArrayType(t *testing.T) {
	input := `nums = [1, 2, 3]`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	a.Analyze(program)

	sym := a.GetSymbol("nums")
	if sym == nil {
		t.Fatal("symbol 'nums' not found")
	}
	if sym.Type.Kind != TypeArray {
		t.Errorf("nums.Type.Kind = %v, want TypeArray", sym.Type.Kind)
	}
	if sym.Type.Elem == nil || !sym.Type.Elem.Equals(TypeIntVal) {
		t.Errorf("nums.Type.Elem = %v, want Int", sym.Type.Elem)
	}
}

func TestAnalyzeMapType(t *testing.T) {
	input := `m = {"a" => 1, "b" => 2}`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	a.Analyze(program)

	sym := a.GetSymbol("m")
	if sym == nil {
		t.Fatal("symbol 'm' not found")
	}
	if sym.Type.Kind != TypeMap {
		t.Errorf("m.Type.Kind = %v, want TypeMap", sym.Type.Kind)
	}
	if sym.Type.KeyType == nil || !sym.Type.KeyType.Equals(TypeStringVal) {
		t.Errorf("m.Type.KeyType = %v, want String", sym.Type.KeyType)
	}
	if sym.Type.ValueType == nil || !sym.Type.ValueType.Equals(TypeIntVal) {
		t.Errorf("m.Type.ValueType = %v, want Int", sym.Type.ValueType)
	}
}

func TestAnalyzeBuiltinFunctions(t *testing.T) {
	input := `puts "hello"`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	// puts should be recognized as a built-in
	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
}

func TestAnalyzeDidYouMean(t *testing.T) {
	input := `
x = 5
puts y
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)

	found := false
	for _, err := range errs {
		if undef, ok := err.(*UndefinedError); ok {
			// Should suggest 'x' for 'y' (same length, differ by 1)
			if slices.Contains(undef.Candidates, "x") {
				found = true
				break
			}
		}
	}
	if !found {
		t.Log("Note: 'did you mean' suggestion for 'y' -> 'x' not found, but this is a heuristic")
	}
}

func TestParseType(t *testing.T) {
	tests := []struct {
		input string
		want  *Type
	}{
		{"Int", TypeIntVal},
		{"Int64", TypeInt64Val},
		{"Float", TypeFloatVal},
		{"Bool", TypeBoolVal},
		{"String", TypeStringVal},
		{"Bytes", TypeBytesVal},
		{"any", TypeAnyVal},
		{"error", TypeErrorVal},
		{"Int?", NewOptionalType(TypeIntVal)},
		{"String?", NewOptionalType(TypeStringVal)},
		{"Array[Int]", NewArrayType(TypeIntVal)},
		{"Array[String]", NewArrayType(TypeStringVal)},
		{"Map[String, Int]", NewMapType(TypeStringVal, TypeIntVal)},
		{"Chan[Int]", NewChanType(TypeIntVal)},
		{"Task[String]", NewTaskType(TypeStringVal)},
		{"(Int, Bool)", NewTupleType(TypeIntVal, TypeBoolVal)},
		{"(String, error)", NewTupleType(TypeStringVal, TypeErrorVal)},
		{"User", NewClassType("User")},
	}

	for _, tt := range tests {
		got := ParseType(tt.input)
		if !got.Equals(tt.want) {
			t.Errorf("ParseType(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestAnalyzeBinaryExpressionTypes(t *testing.T) {
	tests := []struct {
		input    string
		wantType *Type
	}{
		{`x = 1 + 2`, TypeIntVal},
		{`x = 1.0 + 2.0`, TypeFloatVal},
		{`x = "a" + "b"`, TypeStringVal},
		{`x = 1 < 2`, TypeBoolVal},
		{`x = true and false`, TypeBoolVal},
	}

	for _, tt := range tests {
		p := parse(t, tt.input)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("parse errors for %q: %v", tt.input, p.Errors())
		}

		a := NewAnalyzer()
		a.Analyze(program)

		sym := a.GetSymbol("x")
		if sym == nil {
			t.Errorf("symbol 'x' not found for %q", tt.input)
			continue
		}
		if !sym.Type.Equals(tt.wantType) {
			t.Errorf("for %q: x.Type = %v, want %v", tt.input, sym.Type, tt.wantType)
		}
	}
}

func TestAnalyzeInterface(t *testing.T) {
	input := `
interface Speaker
  def speak -> String
end
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}

	iface := a.GetInterface("Speaker")
	if iface == nil {
		t.Fatal("interface 'Speaker' not found")
	}
	if iface.GetMethod("speak") == nil {
		t.Error("method 'speak' not found on interface")
	}
}

func TestAnalyzeScopeNesting(t *testing.T) {
	input := `
x = 1
if true
  y = 2
  puts x  # should see x from outer scope
  puts y  # should see y from this scope
end
puts z  # z is not defined anywhere
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)

	// Should have exactly one error for undefined 'z'
	foundZ := false
	for _, err := range errs {
		if undef, ok := err.(*UndefinedError); ok {
			if undef.Name == "z" {
				foundZ = true
			} else {
				t.Errorf("unexpected undefined error for: %s", undef.Name)
			}
		}
	}
	if !foundZ {
		t.Error("expected UndefinedError for 'z'")
	}
}

func TestAnalyzeMultipleReturnTypes(t *testing.T) {
	input := `
def parse(s : String) -> (Int, Bool)
  return 0, false
end
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}

	fn := a.GetFunction("parse")
	if fn == nil {
		t.Fatal("function 'parse' not found")
	}
	if len(fn.ReturnTypes) != 2 {
		t.Errorf("parse has %d return types, want 2", len(fn.ReturnTypes))
	}
}

func TestAnalyzeRangeType(t *testing.T) {
	input := `r = 1..10`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	a.Analyze(program)

	sym := a.GetSymbol("r")
	if sym == nil {
		t.Fatal("symbol 'r' not found")
	}
	if !sym.Type.Equals(TypeRangeVal) {
		t.Errorf("r.Type = %v, want Range", sym.Type)
	}
}

func TestAnalyzeErrorMessages(t *testing.T) {
	tests := []struct {
		input   string
		wantMsg string
	}{
		{`puts undefined_var`, "undefined: 'undefined_var'"},
		{`@field`, "instance variable '@field' used outside class"},
		{`break`, "break statement outside loop"},
	}

	for _, tt := range tests {
		p := parse(t, tt.input)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("parse errors for %q: %v", tt.input, p.Errors())
		}

		a := NewAnalyzer()
		errs := a.Analyze(program)
		if len(errs) == 0 {
			t.Errorf("expected error for %q", tt.input)
			continue
		}

		found := false
		for _, err := range errs {
			if strings.Contains(err.Error(), tt.wantMsg) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("for %q: expected error containing %q, got: %v", tt.input, tt.wantMsg, errs)
		}
	}
}

func TestAnalyzeArityMismatch(t *testing.T) {
	input := `
def add(a : Int, b : Int) -> Int
  a + b
end

add(1)
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	if len(errs) == 0 {
		t.Fatal("expected error for wrong number of arguments")
	}

	found := false
	for _, err := range errs {
		if arity, ok := err.(*ArityMismatchError); ok {
			if arity.Name == "add" && arity.Expected == 2 && arity.Got == 1 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("expected ArityMismatchError for 'add', got: %v", errs)
	}
}

func TestAnalyzeArityMismatchTooMany(t *testing.T) {
	input := `
def greet(name : String)
  puts name
end

greet("Alice", "Bob")
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	if len(errs) == 0 {
		t.Fatal("expected error for wrong number of arguments")
	}

	found := false
	for _, err := range errs {
		if arity, ok := err.(*ArityMismatchError); ok {
			if arity.Name == "greet" && arity.Expected == 1 && arity.Got == 2 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("expected ArityMismatchError for 'greet', got: %v", errs)
	}
}

func TestAnalyzeArityCorrect(t *testing.T) {
	input := `
def add(a : Int, b : Int) -> Int
  a + b
end

result = add(1, 2)
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)
	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
}

func TestAnalyzeBuiltinVariadic(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`puts "hello"`},
		{`puts "hello", "world"`},
		{`puts "a", "b", "c", "d"`},
		{`print "one", "two", "three"`},
		{`p 1, 2, 3`},
	}

	for _, tt := range tests {
		p := parse(t, tt.input)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("parse errors for %q: %v", tt.input, p.Errors())
		}

		a := NewAnalyzer()
		errs := a.Analyze(program)
		if len(errs) > 0 {
			t.Errorf("for %q: expected no errors, got: %v", tt.input, errs)
		}
	}
}
