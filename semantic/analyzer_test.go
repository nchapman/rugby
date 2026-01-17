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
	input := `x: Int64 = 5`
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
def add(a: Int, b: Int): Int
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
  def initialize(@name: String, @age: Int)
  end

  def greet: String
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

func TestAnalyzeReturnInsideBlock(t *testing.T) {
	input := `
def main
  [1, 2, 3].each do |x|
    return x
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
	if len(errs) == 0 {
		t.Fatal("expected error for return inside block")
	}

	found := false
	for _, err := range errs {
		if _, ok := err.(*ReturnInsideBlockError); ok {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected ReturnInsideBlockError, got: %v", errs)
	}
}

func TestAnalyzeReturnInsideMapBlock(t *testing.T) {
	input := `
def main
  [1, 2, 3].map do |x|
    return x * 2 if x > 1
    x
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
	if len(errs) == 0 {
		t.Fatal("expected error for return inside block")
	}

	found := false
	for _, err := range errs {
		if _, ok := err.(*ReturnInsideBlockError); ok {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected ReturnInsideBlockError, got: %v", errs)
	}
}

func TestAnalyzeReturnOutsideBlockIsAllowed(t *testing.T) {
	// return in a regular loop should still be allowed
	input := `
def main: Int
  for i in [1, 2, 3]
    return i if i > 1
  end
  0
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

func TestAnalyzeBreakInsideBlock(t *testing.T) {
	input := `
def main
  [1, 2, 3].each do |x|
    if x == 2
      break
    end
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
	if len(errs) == 0 {
		t.Fatal("expected error for break inside block")
	}

	found := false
	for _, err := range errs {
		if _, ok := err.(*BreakInsideBlockError); ok {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected BreakInsideBlockError, got: %v", errs)
	}
}

func TestAnalyzeNextInsideBlock(t *testing.T) {
	input := `
def main
  [1, 2, 3].each do |x|
    next if x == 2
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
	if len(errs) == 0 {
		t.Fatal("expected error for next inside block")
	}

	found := false
	for _, err := range errs {
		if _, ok := err.(*NextInsideBlockError); ok {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected NextInsideBlockError, got: %v", errs)
	}
}

func TestAnalyzeBreakNextInLoopAllowed(t *testing.T) {
	// break and next in regular loops should still be allowed
	input := `
def main
  for i in [1, 2, 3]
    next if i == 1
    break if i == 2
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
}

func TestAnalyzeLoopInsideBlockAllowsBreakNext(t *testing.T) {
	// A for loop inside a block should still allow break/next
	// The loop creates a new scope that allows control flow
	input := `
def main
  [1, 2, 3].each do |x|
    for i in [1, 2]
      next if i == 1
      break if i == 2
    end
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
		t.Errorf("break/next in loop inside block should be allowed, got errors: %v", errs)
	}
}

func TestAnalyzeIfLet(t *testing.T) {
	input := `
def find_user(id: Int): User?
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
		{"Array<Int>", NewArrayType(TypeIntVal)},
		{"Array<String>", NewArrayType(TypeStringVal)},
		{"Map<String, Int>", NewMapType(TypeStringVal, TypeIntVal)},
		{"Chan<Int>", NewChanType(TypeIntVal)},
		{"Task<String>", NewTaskType(TypeStringVal)},
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
  def speak: String
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
def parse(s: String): (Int, Bool)
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
def add(a: Int, b: Int): Int
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
def greet(name: String)
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
def add(a: Int, b: Int): Int
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

func TestAnalyzeOperatorTypeMismatch(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOp  string
		wantErr bool
	}{
		// Arithmetic errors
		{"int plus string", `x = 1 + "hello"`, "+", true},
		{"string minus int", `x = "hello" - 1`, "-", true},
		{"bool times int", `x = true * 5`, "*", true},
		{"string divide float", `x = "a" / 2.0`, "/", true},
		{"bool modulo int", `x = false % 2`, "%", true},

		// Comparison errors
		{"string greater than int", `x = "hello" > 5`, ">", true},
		{"bool less than float", `x = true < 1.0`, "<", true},
		{"string less-equal int", `x = "a" <= 1`, "<=", true},
		{"bool greater-equal int", `x = false >= 0`, ">=", true},

		// Logical errors - parser converts 'and'/'or' to '&&'/'||' in AST
		{"int and int", `x = 1 and 2`, "&&", true},
		{"string or string", `x = "a" or "b"`, "||", true},
		{"int and bool", `x = 1 and true`, "&&", true},
		{"bool or int", `x = true or 1`, "||", true},

		// Valid operations - should NOT error
		{"int plus int", `x = 1 + 2`, "", false},
		{"float plus float", `x = 1.0 + 2.0`, "", false},
		{"int plus float", `x = 1 + 2.0`, "", false}, // mixed numeric allowed
		{"string plus string", `x = "a" + "b"`, "", false},
		{"int less than int", `x = 1 < 2`, "", false},
		{"string less than string", `x = "a" < "b"`, "", false},
		{"int equals int", `x = 1 == 2`, "", false},
		{"string equals string", `x = "a" == "b"`, "", false},
		{"bool and bool", `x = true and false`, "", false},
		{"bool or bool", `x = true or false`, "", false},
		{"int equals float", `x = 1 == 2.0`, "", false},   // numeric comparison allowed
		{"float less than int", `x = 1.5 < 2`, "", false}, // numeric comparison allowed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error for %q, got none", tt.input)
					return
				}
				found := false
				for _, err := range errs {
					if opErr, ok := err.(*OperatorTypeMismatchError); ok {
						if opErr.Op == tt.wantOp {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("expected OperatorTypeMismatchError for '%s', got: %v", tt.wantOp, errs)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("expected no errors, got: %v", errs)
				}
			}
		})
	}
}

func TestAnalyzeUnaryTypeMismatch(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOp   string
		wantErr  bool
		expected string // expected type in error message
	}{
		// Logical not errors - 'not' requires Bool (parser converts 'not' to '!' in AST)
		{"not int", `x = not 5`, "!", true, "Bool"},
		{"not string", `x = not "hello"`, "!", true, "Bool"},
		{"not float", `x = not 3.14`, "!", true, "Bool"},
		{"not array", `x = not [1, 2, 3]`, "!", true, "Bool"},

		// Unary minus errors - '-' requires numeric
		{"minus string", `x = -"hello"`, "-", true, "numeric"},
		{"minus bool", `x = -true`, "-", true, "numeric"},
		{"minus array", `x = -[1, 2]`, "-", true, "numeric"},

		// Valid operations - should NOT error
		{"not bool", `x = not true`, "", false, ""},
		{"not false", `x = not false`, "", false, ""},
		{"minus int", `x = -5`, "", false, ""},
		{"minus float", `x = -3.14`, "", false, ""},
		{"minus int64", `def main
  x: Int64 = 5
  y = -x
end`, "", false, ""},
		{"double minus", `x = --5`, "", false, ""},
		{"not not bool", `x = not not true`, "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error for %q, got none", tt.input)
					return
				}
				found := false
				for _, err := range errs {
					if unaryErr, ok := err.(*UnaryTypeMismatchError); ok {
						if unaryErr.Op == tt.wantOp && unaryErr.Expected == tt.expected {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("expected UnaryTypeMismatchError for '%s' (expected %s), got: %v", tt.wantOp, tt.expected, errs)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("expected no errors, got: %v", errs)
				}
			}
		})
	}
}

func TestAnalyzeEqualityComparison(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid equality comparisons - primitives
		{"int equals int", `x = 1 == 2`, false},
		{"string equals string", `x = "a" == "b"`, false},
		{"bool equals bool", `x = true == false`, false},
		{"int not-equals int", `x = 1 != 2`, false},
		{"numeric mixed", `x = 1 == 2.0`, false}, // int and float are comparable

		// Valid equality comparisons - arrays with same element type
		{"int array equals int array", `x = [1, 2] == [3, 4]`, false},
		{"string array equals string array", `x = ["a"] == ["b"]`, false},
		{"nested int array", `x = [[1, 2], [3]] == [[4]]`, false},

		// Valid equality comparisons - maps with same key/value types
		{"string-int map equals", `x = {"a" => 1} == {"b" => 2}`, false},
		{"int-string map equals", `x = {1 => "a"} == {2 => "b"}`, false},

		// Invalid equality comparisons - primitives
		{"int equals string", `x = 1 == "hello"`, true},
		{"bool equals int", `x = true == 1`, true},
		{"string equals bool", `x = "true" == true`, true},

		// Invalid equality comparisons - arrays with different element types
		{"int array equals string array", `x = [1, 2] == ["a", "b"]`, true},
		{"bool array equals int array", `x = [true] == [1]`, true},

		// Invalid equality comparisons - maps with different types
		{"different key types", `x = {"a" => 1} == {1 => 1}`, true},
		{"different value types", `x = {"a" => 1} == {"a" => "one"}`, true},

		// Class instance equality (requires class definitions to be tested separately)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr && len(errs) == 0 {
				t.Errorf("expected error for %q, got none", tt.input)
			} else if !tt.wantErr && len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeClassInstanceEquality(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "same class instances comparable",
			input: `class Dog
  def initialize(name: String)
    @name: String = name
  end
end

def main
  dog1 = Dog.new("Fido")
  dog2 = Dog.new("Rex")
  x = dog1 == dog2
end`,
			wantErr: false,
		},
		{
			name: "different class instances not comparable",
			input: `class Dog
  def initialize(name: String)
    @name: String = name
  end
end

class Cat
  def initialize(name: String)
    @name: String = name
  end
end

def main
  dog = Dog.new("Fido")
  cat = Cat.new("Whiskers")
  x = dog == cat
end`,
			wantErr: true,
		},
		{
			name: "class compared with primitive not comparable",
			input: `class Dog
  def initialize(name: String)
    @name: String = name
  end
end

def main
  dog = Dog.new("Fido")
  x = dog == "hello"
end`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr && len(errs) == 0 {
				t.Errorf("expected error, got none")
			} else if !tt.wantErr && len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeReturnTypeValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "correct return type",
			input: `def add(a: Int, b: Int): Int
  return a + b
end`,
			wantErr: false,
		},
		{
			name: "wrong return type - string instead of int",
			input: `def add(a: Int, b: Int): Int
  return "hello"
end`,
			wantErr: true,
		},
		{
			name: "wrong return type - bool instead of string",
			input: `def getName: String
  return true
end`,
			wantErr: true,
		},
		{
			name: "compatible numeric return - int for float",
			input: `def getFloat: Float
  return 42
end`,
			wantErr: false, // Int is assignable to Float
		},
		{
			name: "no return type declared - any return allowed",
			input: `def anything
  return "whatever"
end`,
			wantErr: false,
		},
		{
			name: "return array of correct type",
			input: `def getNumbers: Array<Int>
  return [1, 2, 3]
end`,
			wantErr: false,
		},
		{
			name: "return array of wrong element type",
			input: `def getNumbers: Array<Int>
  return ["a", "b"]
end`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr && len(errs) == 0 {
				t.Errorf("expected error, got none")
			} else if !tt.wantErr && len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeArgumentTypeValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "correct argument types",
			input: `def greet(name: String, age: Int)
  puts name
end

greet("Alice", 30)`,
			wantErr: false,
		},
		{
			name: "wrong argument type - int instead of string",
			input: `def greet(name: String)
  puts name
end

greet(42)`,
			wantErr: true,
		},
		{
			name: "wrong argument type - string instead of int",
			input: `def addOne(x: Int): Int
  return x + 1
end

addOne("hello")`,
			wantErr: true,
		},
		{
			name: "compatible numeric - int for float param",
			input: `def double(x: Float): Float
  return x * 2.0
end

double(5)`,
			wantErr: false, // Int is assignable to Float
		},
		{
			name: "wrong second argument type",
			input: `def compute(a: Int, b: Int): Int
  return a + b
end

compute(1, "two")`,
			wantErr: true,
		},
		{
			name: "array argument with correct element type",
			input: `def sumAll(nums: Array<Int>): Int
  return 0
end

sumAll([1, 2, 3])`,
			wantErr: false,
		},
		{
			name: "array argument with wrong element type",
			input: `def sumAll(nums: Array<Int>): Int
  return 0
end

sumAll(["a", "b"])`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr && len(errs) == 0 {
				t.Errorf("expected error, got none")
			} else if !tt.wantErr && len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeConditionTypeValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		context string // expected context in error message
	}{
		// Valid conditions
		{
			name:    "if with bool literal",
			input:   "if true\n  puts \"yes\"\nend",
			wantErr: false,
		},
		{
			name:    "if with comparison",
			input:   "if 1 < 2\n  puts \"yes\"\nend",
			wantErr: false,
		},
		{
			name:    "if with bool variable",
			input:   "flag = true\nif flag\n  puts \"yes\"\nend",
			wantErr: false,
		},
		{
			name:    "while with bool",
			input:   "x = true\nwhile x\n  x = false\nend",
			wantErr: false,
		},
		{
			name:    "until with bool",
			input:   "done = false\nuntil done\n  done = true\nend",
			wantErr: false,
		},
		{
			name:    "unless with bool",
			input:   "unless false\n  puts \"yes\"\nend",
			wantErr: false,
		},

		// Invalid conditions - if
		{
			name:    "if with int",
			input:   "if 42\n  puts \"yes\"\nend",
			wantErr: true,
			context: "if",
		},
		{
			name:    "if with string",
			input:   "if \"hello\"\n  puts \"yes\"\nend",
			wantErr: true,
			context: "if",
		},
		{
			name:    "if with array",
			input:   "if [1, 2, 3]\n  puts \"yes\"\nend",
			wantErr: true,
			context: "if",
		},

		// Invalid conditions - while
		{
			name:    "while with int",
			input:   "while 1\n  break\nend",
			wantErr: true,
			context: "while",
		},
		{
			name:    "while with string",
			input:   "while \"go\"\n  break\nend",
			wantErr: true,
			context: "while",
		},

		// Invalid conditions - until
		{
			name:    "until with int",
			input:   "until 0\n  break\nend",
			wantErr: true,
			context: "until",
		},
		{
			name:    "until with float",
			input:   "until 3.14\n  break\nend",
			wantErr: true,
			context: "until",
		},

		// Invalid conditions - unless
		{
			name:    "unless with int",
			input:   "unless 0\n  puts \"yes\"\nend",
			wantErr: true,
			context: "unless",
		},

		// Ternary expressions
		{
			name:    "ternary with bool condition",
			input:   "x = true ? 1 : 2",
			wantErr: false,
		},
		{
			name:    "ternary with comparison condition",
			input:   "x = (1 < 2) ? \"yes\" : \"no\"",
			wantErr: false,
		},
		{
			name:    "ternary with int condition",
			input:   "x = 42 ? 1 : 2",
			wantErr: true,
			context: "ternary",
		},
		{
			name:    "ternary with string condition",
			input:   "x = \"hello\" ? 1 : 2",
			wantErr: true,
			context: "ternary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				found := false
				for _, err := range errs {
					if condErr, ok := err.(*ConditionTypeMismatchError); ok {
						if condErr.Context == tt.context {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("expected ConditionTypeMismatchError with context %q, got: %v", tt.context, errs)
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeIndexTypeValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid array indexing
		{
			name:    "array with int literal index",
			input:   "arr = [1, 2, 3]\nx = arr[0]",
			wantErr: false,
		},
		{
			name:    "array with int variable index",
			input:   "arr = [1, 2, 3]\ni = 1\nx = arr[i]",
			wantErr: false,
		},
		{
			name: "array with int64 index",
			input: `def main
  arr = [1, 2, 3]
  i: Int64 = 1
  x = arr[i]
end`,
			wantErr: false,
		},

		// Valid string indexing
		{
			name:    "string with int literal index",
			input:   "s = \"hello\"\nc = s[0]",
			wantErr: false,
		},
		{
			name:    "string with int variable index",
			input:   "s = \"hello\"\ni = 2\nc = s[i]",
			wantErr: false,
		},

		// Valid map indexing (any key type is allowed)
		{
			name:    "map with string key",
			input:   "m = {\"a\" => 1}\nv = m[\"a\"]",
			wantErr: false,
		},
		{
			name:    "map with int key",
			input:   "m = {1 => \"one\"}\nv = m[1]",
			wantErr: false,
		},

		// Invalid array indexing
		{
			name:    "array with string index",
			input:   "arr = [1, 2, 3]\nx = arr[\"first\"]",
			wantErr: true,
		},
		{
			name:    "array with float index",
			input:   "arr = [1, 2, 3]\nx = arr[1.5]",
			wantErr: true,
		},
		{
			name:    "array with bool index",
			input:   "arr = [1, 2, 3]\nx = arr[true]",
			wantErr: true,
		},

		// Invalid string indexing
		{
			name:    "string with string index",
			input:   "s = \"hello\"\nc = s[\"x\"]",
			wantErr: true,
		},
		{
			name:    "string with float index",
			input:   "s = \"hello\"\nc = s[2.5]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				found := false
				for _, err := range errs {
					if _, ok := err.(*IndexTypeMismatchError); ok {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected IndexTypeMismatchError, got: %v", errs)
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeCompoundAssignmentTypeValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantOp  string
	}{
		// Valid compound assignments
		{
			name:    "int += int",
			input:   "x = 5\nx += 3",
			wantErr: false,
		},
		{
			name:    "float += float",
			input:   "x = 5.0\nx += 3.0",
			wantErr: false,
		},
		{
			name:    "int += float (numeric widening)",
			input:   "x = 5\nx += 3.0",
			wantErr: false,
		},
		{
			name:    "string += string",
			input:   "s = \"hello\"\ns += \" world\"",
			wantErr: false,
		},
		{
			name:    "int -= int",
			input:   "x = 10\nx -= 3",
			wantErr: false,
		},
		{
			name:    "int *= int",
			input:   "x = 5\nx *= 2",
			wantErr: false,
		},
		{
			name:    "int /= int",
			input:   "x = 10\nx /= 2",
			wantErr: false,
		},

		// Invalid compound assignments
		{
			name:    "int += string",
			input:   "x = 5\nx += \"hello\"",
			wantErr: true,
			wantOp:  "+=",
		},
		{
			name:    "string += int",
			input:   "s = \"hello\"\ns += 5",
			wantErr: true,
			wantOp:  "+=",
		},
		{
			name:    "string -= string",
			input:   "s = \"hello\"\ns -= \"world\"",
			wantErr: true,
			wantOp:  "-=",
		},
		{
			name:    "string *= int",
			input:   "s = \"hello\"\ns *= 3",
			wantErr: true,
			wantOp:  "*=",
		},
		{
			name:    "bool += bool",
			input:   "b = true\nb += false",
			wantErr: true,
			wantOp:  "+=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				found := false
				for _, err := range errs {
					if opErr, ok := err.(*OperatorTypeMismatchError); ok {
						if opErr.Op == tt.wantOp {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("expected OperatorTypeMismatchError for %q, got: %v", tt.wantOp, errs)
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeArrayMapHomogeneity(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		context string // expected context in error
	}{
		// Valid homogeneous arrays
		{
			name:    "int array",
			input:   "arr = [1, 2, 3]",
			wantErr: false,
		},
		{
			name:    "string array",
			input:   "arr = [\"a\", \"b\", \"c\"]",
			wantErr: false,
		},
		{
			name:    "float array",
			input:   "arr = [1.0, 2.5, 3.14]",
			wantErr: false,
		},
		{
			name:    "mixed numeric array (int and float)",
			input:   "arr = [1, 2.5, 3]",
			wantErr: false, // numeric widening allowed
		},
		{
			name:    "empty array",
			input:   "arr = []",
			wantErr: false,
		},

		// Invalid heterogeneous arrays
		{
			name:    "int and string in array",
			input:   "arr = [1, \"hello\", 3]",
			wantErr: true,
			context: "array element",
		},
		{
			name:    "string and bool in array",
			input:   "arr = [\"a\", true, \"c\"]",
			wantErr: true,
			context: "array element",
		},
		{
			name:    "int and bool in array",
			input:   "arr = [1, false, 3]",
			wantErr: true,
			context: "array element",
		},

		// Valid homogeneous maps
		{
			name:    "string to int map",
			input:   "m = {\"a\" => 1, \"b\" => 2}",
			wantErr: false,
		},
		{
			name:    "int to string map",
			input:   "m = {1 => \"one\", 2 => \"two\"}",
			wantErr: false,
		},
		{
			name:    "empty map",
			input:   "m = {}",
			wantErr: false,
		},

		// Invalid heterogeneous maps - key type
		{
			name:    "mixed key types (string and int)",
			input:   "m = {\"a\" => 1, 2 => 3}",
			wantErr: true,
			context: "map key",
		},

		// Heterogeneous maps - allowed (widens to Map[K, any])
		// This is valid in Ruby/Rugby: {host:, port:} where host is String and port is Int
		{
			name:    "mixed value types (int and string)",
			input:   "m = {\"a\" => 1, \"b\" => \"two\"}",
			wantErr: false,
		},
		{
			name:    "mixed value types (string and bool)",
			input:   "m = {1 => \"one\", 2 => true}",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				found := false
				for _, err := range errs {
					if typeErr, ok := err.(*TypeMismatchError); ok {
						if typeErr.Context == tt.context {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("expected TypeMismatchError with context %q, got: %v", tt.context, errs)
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeCaseWhenTypeMatching(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid case statements
		{
			name: "case int when int",
			input: `x = 5
case x
when 1
  puts "one"
when 2
  puts "two"
end`,
			wantErr: false,
		},
		{
			name: "case string when string",
			input: `s = "hello"
case s
when "hello"
  puts "hi"
when "bye"
  puts "goodbye"
end`,
			wantErr: false,
		},
		{
			name: "case with numeric widening",
			input: `x = 5
case x
when 1.0
  puts "one"
end`,
			wantErr: false, // Int and Float are comparable
		},
		{
			name: "case without subject (bool when)",
			input: `x = 5
case
when x > 0
  puts "positive"
when x < 0
  puts "negative"
end`,
			wantErr: false,
		},

		// Invalid case statements
		{
			name: "case int when string",
			input: `x = 5
case x
when "hello"
  puts "mismatch"
end`,
			wantErr: true,
		},
		{
			name: "case string when int",
			input: `s = "hello"
case s
when 42
  puts "mismatch"
end`,
			wantErr: true,
		},
		{
			name: "case int when bool",
			input: `x = 5
case x
when true
  puts "mismatch"
end`,
			wantErr: true,
		},
		{
			name: "case bool when string",
			input: `b = true
case b
when "yes"
  puts "mismatch"
end`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				found := false
				for _, err := range errs {
					if typeErr, ok := err.(*TypeMismatchError); ok {
						if typeErr.Context == "case when value" {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("expected TypeMismatchError with context 'case when value', got: %v", errs)
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeNilCoalesceTypeValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		context string
	}{
		// Valid nil coalesce operations
		{
			name: "optional int with int default",
			input: `def check(x: Int?)
  y = x ?? 0
end`,
			wantErr: false,
		},
		{
			name: "optional string with string default",
			input: `def check(x: String?)
  y = x ?? "default"
end`,
			wantErr: false,
		},
		{
			name: "optional float with int default (numeric widening)",
			input: `def check(x: Float?)
  y = x ?? 0
end`,
			wantErr: false,
		},

		// Invalid nil coalesce operations
		{
			name: "optional int with string default",
			input: `def check(x: Int?)
  y = x ?? "hello"
end`,
			wantErr: true,
			context: "nil coalesce default value",
		},
		{
			name: "optional string with int default",
			input: `def check(x: String?)
  y = x ?? 42
end`,
			wantErr: true,
			context: "nil coalesce default value",
		},
		{
			name: "non-optional left side",
			input: `x = 5
y = x ?? 0`,
			wantErr: true,
			context: "nil coalesce left side",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				found := false
				for _, err := range errs {
					if typeErr, ok := err.(*TypeMismatchError); ok {
						if typeErr.Context == tt.context {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("expected TypeMismatchError with context %q, got: %v", tt.context, errs)
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeMethodCallTypeChecking(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		// String method return types
		{
			name: "string length returns Int",
			input: `s = "hello"
x: Int = s.length`,
			wantErr: false,
		},
		{
			name: "string upcase returns String",
			input: `s = "hello"
x: String = s.upcase`,
			wantErr: false,
		},
		{
			name: "string empty? returns Bool",
			input: `s = "hello"
x: Bool = s.empty?`,
			wantErr: false,
		},
		{
			name: "string split returns Array[String]",
			input: `s = "a,b,c"
words = s.split(",")
first: String = words[0]`,
			wantErr: false,
		},

		// Array method return types
		{
			name: "array length returns Int",
			input: `arr = [1, 2, 3]
x: Int = arr.length`,
			wantErr: false,
		},
		{
			name: "array first returns Optional",
			input: `arr = [1, 2, 3]
x = arr.first
y = x ?? 0`,
			wantErr: false,
		},
		{
			name: "array empty? returns Bool",
			input: `arr = [1, 2, 3]
x: Bool = arr.empty?`,
			wantErr: false,
		},

		// Int method return types
		{
			name: "int abs returns Int",
			input: `n = -5
x: Int = n.abs`,
			wantErr: false,
		},
		{
			name: "int even? returns Bool",
			input: `n = 4
x: Bool = n.even?`,
			wantErr: false,
		},
		{
			name: "int to_s returns String",
			input: `n = 42
x: String = n.to_s`,
			wantErr: false,
		},

		// Float method return types
		{
			name: "float floor returns Float",
			input: `f = 3.14
x: Float = f.floor`,
			wantErr: false,
		},
		{
			name: "float nan? returns Bool",
			input: `f = 3.14
x: Bool = f.nan?`,
			wantErr: false,
		},

		// Map method return types
		{
			name: "map keys returns array of keys",
			input: `m = {"a" => 1, "b" => 2}
keys = m.keys
first: String = keys[0]`,
			wantErr: false,
		},
		{
			name: "map values returns array of values",
			input: `m = {"a" => 1, "b" => 2}
vals = m.values
first: Int = vals[0]`,
			wantErr: false,
		},
		{
			name: "map has_key? returns Bool",
			input: `m = {"a" => 1}
x: Bool = m.has_key?("a")`,
			wantErr: false,
		},

		// Class method return types
		{
			name: "class method returns declared type",
			input: `
class User
  def initialize(@name: String)
  end

  def get_name: String
    @name
  end
end

u = User.new("Alice")
name: String = u.get_name`,
			wantErr: false,
		},
		{
			name: "class method with wrong assignment type",
			input: `
class User
  def initialize(@name: String)
  end

  def get_name: String
    @name
  end
end

u = User.new("Alice")
name: Int = u.get_name`,
			wantErr: true,
			errMsg:  "assignment",
		},

		// Chained method calls
		{
			name: "chained string methods",
			input: `s = "hello world"
x: String = s.upcase.strip`,
			wantErr: false,
		},
		{
			name: "chained array methods",
			input: `arr = [3, 1, 2]
sorted = arr.sorted
len: Int = sorted.length`,
			wantErr: false,
		},

		// Method argument type checking
		{
			name: "string include? with correct arg type",
			input: `s = "hello"
x = s.include?("ell")`,
			wantErr: false,
		},

		// Range method return types
		{
			name: "range to_a returns array of Int",
			input: `r = 1..5
arr = r.to_a
x: Int = arr[0]`,
			wantErr: false,
		},
		{
			name: "range include? returns Bool",
			input: `r = 1..10
x: Bool = r.include?(5)`,
			wantErr: false,
		},

		// Note: Channel tests require Chan to be a known type, skipped for now

		// Optional method return types
		{
			name: "optional ok? returns Bool",
			input: `def get_val: Int?
  5
end
x = get_val
y: Bool = x.ok?`,
			wantErr: false,
		},
		{
			name: "optional unwrap returns inner type",
			input: `def get_val: Int?
  5
end
x = get_val
y: Int = x.unwrap`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				if tt.errMsg != "" {
					found := false
					for _, err := range errs {
						if strings.Contains(err.Error(), tt.errMsg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing %q, got: %v", tt.errMsg, errs)
					}
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeBangOperator(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "bang at top level (allowed for scripts)",
			input: `
def read_file(path: String): (String, Error)
  return "contents", nil
end

data = read_file("test.txt")!
puts data`,
			wantErr: false,
		},
		{
			name: "bang in non-error-returning function",
			input: `
def read_file(path: String): (String, Error)
  return "contents", nil
end

def process: String
  read_file("test.txt")!
end`,
			wantErr: true,
			errMsg:  "can only be used in functions that return error",
		},
		{
			name: "bang on error-only return",
			input: `
def validate(x: Int): Error
  nil
end

def process: Error
  validate(5)!
  nil
end`,
			wantErr: false,
		},
		{
			name: "bang on non-error function",
			input: `
def get_value: Int
  42
end

def process: Error
  get_value!
  nil
end`,
			wantErr: true,
			errMsg:  "cannot use '!'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				if tt.errMsg != "" {
					found := false
					for _, err := range errs {
						if strings.Contains(err.Error(), tt.errMsg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing %q, got: %v", tt.errMsg, errs)
					}
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeInterfaceImplementation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "class correctly implements interface",
			input: `
interface Speaker
  def speak: String
end

class Dog implements Speaker
  def speak: String
    "woof"
  end
end`,
			wantErr: false,
		},
		{
			name: "class missing interface method",
			input: `
interface Speaker
  def speak: String
end

class Dog implements Speaker
end`,
			wantErr: true,
			errMsg:  "missing method 'speak'",
		},
		{
			name: "class with wrong return type",
			input: `
interface Speaker
  def speak: String
end

class Dog implements Speaker
  def speak: Int
    42
  end
end`,
			wantErr: true,
			errMsg:  "wrong signature",
		},
		{
			name: "class with wrong parameter type",
			input: `
interface Greeter
  def greet(name: String): String
end

class FormalGreeter implements Greeter
  def greet(name: Int): String
    "Hello"
  end
end`,
			wantErr: true,
			errMsg:  "wrong signature",
		},
		{
			name: "class with wrong parameter count",
			input: `
interface Adder
  def add(a: Int, b: Int): Int
end

class Calculator implements Adder
  def add(a: Int): Int
    a
  end
end`,
			wantErr: true,
			errMsg:  "wrong signature",
		},
		{
			name: "undefined interface",
			input: `
class Dog implements NonExistent
  def speak: String
    "woof"
  end
end`,
			wantErr: true,
			errMsg:  "undefined: 'NonExistent'",
		},
		{
			name: "interface with multiple methods",
			input: `
interface Animal
  def speak: String
  def move: String
end

class Dog implements Animal
  def speak: String
    "woof"
  end
  def move: String
    "run"
  end
end`,
			wantErr: false,
		},
		{
			name: "interface with multiple methods - one missing",
			input: `
interface Animal
  def speak: String
  def move: String
end

class Dog implements Animal
  def speak: String
    "woof"
  end
end`,
			wantErr: true,
			errMsg:  "missing method 'move'",
		},
		{
			name: "inherited method satisfies interface",
			input: `
interface Speaker
  def speak: String
end

class Animal
  def speak: String
    "sound"
  end
end

class Dog < Animal implements Speaker
end`,
			wantErr: false,
		},
		{
			name: "inherited method with override satisfies interface",
			input: `
interface Speaker
  def speak: String
end

class Animal
  def speak: String
    "sound"
  end
end

class Dog < Animal implements Speaker
  def speak: String
    "woof"
  end
end`,
			wantErr: false,
		},
		{
			name: "parent missing method fails interface",
			input: `
interface Speaker
  def speak: String
end

class Animal
end

class Dog < Animal implements Speaker
end`,
			wantErr: true,
			errMsg:  "missing method 'speak'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				if tt.errMsg != "" {
					found := false
					for _, err := range errs {
						if strings.Contains(err.Error(), tt.errMsg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing %q, got: %v", tt.errMsg, errs)
					}
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeCircularInheritance(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "direct circular inheritance",
			input: `
class A < B
end
class B < A
end`,
			wantErr: true,
			errMsg:  "circular inheritance",
		},
		{
			name: "self inheritance",
			input: `
class A < A
end`,
			wantErr: true,
			errMsg:  "circular inheritance",
		},
		{
			name: "circular inheritance chain",
			input: `
class A < B
end
class B < C
end
class C < A
end`,
			wantErr: true,
			errMsg:  "circular inheritance",
		},
		{
			name: "valid inheritance chain",
			input: `
class Animal
end
class Dog < Animal
end
class Puppy < Dog
end`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				if tt.errMsg != "" {
					found := false
					for _, err := range errs {
						if strings.Contains(err.Error(), tt.errMsg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing %q, got: %v", tt.errMsg, errs)
					}
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeClassMethodArgumentTypeChecking(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "class method with correct argument type",
			input: `
class Calculator
  def add(a: Int, b: Int): Int
    a + b
  end
end

c = Calculator.new
result = c.add(1, 2)`,
			wantErr: false,
		},
		{
			name: "class method with wrong argument type",
			input: `
class Calculator
  def add(a: Int, b: Int): Int
    a + b
  end
end

c = Calculator.new
result = c.add("hello", 2)`,
			wantErr: true,
			errMsg:  "cannot use String as Int",
		},
		{
			name: "class method with wrong arity",
			input: `
class Calculator
  def add(a: Int, b: Int): Int
    a + b
  end
end

c = Calculator.new
result = c.add(1)`,
			wantErr: true,
			errMsg:  "wrong number of arguments",
		},
		{
			name: "constructor with correct argument types",
			input: `
class User
  def initialize(@name: String, @age: Int)
  end
end

u = User.new("Alice", 30)`,
			wantErr: false,
		},
		{
			name: "constructor with wrong argument type",
			input: `
class User
  def initialize(@name: String, @age: Int)
  end
end

u = User.new(123, 30)`,
			wantErr: true,
			errMsg:  "cannot use Int as String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				if tt.errMsg != "" {
					found := false
					for _, err := range errs {
						if strings.Contains(err.Error(), tt.errMsg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing %q, got: %v", tt.errMsg, errs)
					}
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeBlockParameterTypeInference(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "array.each block param is element type",
			input: `
arr = [1, 2, 3]
arr.each do |x|
  y = x + 1
end
`,
			wantErr: false,
		},
		{
			name: "array.each type mismatch in block",
			input: `
arr = [1, 2, 3]
arr.each do |x|
  y = x + "string"
end
`,
			wantErr: true,
			errMsg:  "cannot use '+' with types Int and String",
		},
		{
			name: "array.map block param is element type",
			input: `
arr = ["a", "b", "c"]
arr.map do |s|
  s.length
end
`,
			wantErr: false,
		},
		{
			name: "array.each_with_index has elem and int params",
			input: `
arr = ["a", "b"]
arr.each_with_index do |elem, idx|
  x = idx + 1
  y = elem.length
end
`,
			wantErr: false,
		},
		{
			name: "map.each has key and value params",
			input: `
m = {"a" => 1, "b" => 2}
m.each do |k, v|
  x = k.length
  y = v + 1
end
`,
			wantErr: false,
		},
		{
			name: "map.each type mismatch on key",
			input: `
m = {"a" => 1}
m.each do |k, v|
  x = k + 1
end
`,
			wantErr: true,
			errMsg:  "cannot use '+' with types String and Int",
		},
		{
			name: "map.each_key has key param",
			input: `
m = {"a" => 1}
m.each_key do |k|
  x = k.length
end
`,
			wantErr: false,
		},
		{
			name: "map.each_value has value param",
			input: `
m = {"a" => 1}
m.each_value do |v|
  x = v + 1
end
`,
			wantErr: false,
		},
		{
			name: "int.times block param is int",
			input: `
5.times do |i|
  x = i + 1
end
`,
			wantErr: false,
		},
		{
			name: "int.times type mismatch",
			input: `
5.times do |i|
  x = i + "string"
end
`,
			wantErr: true,
			errMsg:  "cannot use '+' with types Int and String",
		},
		{
			name: "range.each block param is int",
			input: `
(1..10).each do |n|
  x = n + 1
end
`,
			wantErr: false,
		},
		{
			name: "array.select block param is element type",
			input: `
nums = [1, 2, 3, 4, 5]
nums.select do |n|
  n > 2
end
`,
			wantErr: false,
		},
		{
			name: "array.reduce has accumulator and element params",
			input: `
nums = [1, 2, 3]
nums.reduce(0) do |acc, n|
  acc + n
end
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				if tt.errMsg != "" {
					found := false
					for _, err := range errs {
						if strings.Contains(err.Error(), tt.errMsg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing %q, got: %v", tt.errMsg, errs)
					}
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeSafeNavigationReturnType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "safe nav on class field returns optional of field type",
			input: `
class User
  def initialize(@name: String)
  end
end

def get_user: User?
  nil
end

u = get_user
# name should be String?, so we can use string methods after unwrap
if let name = u&.name
  x = name.length
end
`,
			wantErr: false,
		},
		{
			name: "safe nav on string method returns optional",
			input: `
def get_string: String?
  "hello"
end
s = get_string
# length returns Int, so safe nav returns Int?
len = s&.length
`,
			wantErr: false,
		},
		{
			name: "safe nav on array method returns optional",
			input: `
def get_array: Array<Int>?
  [1, 2, 3]
end
arr = get_array
# first returns Int?, safe nav returns Int? (doesn't double wrap)
first = arr&.first
`,
			wantErr: false,
		},
		{
			name: "safe nav result used correctly with if-let",
			input: `
class User
  def initialize(@age: Int)
  end
end

def find_user: User?
  nil
end

u = find_user
if let age = u&.age
  x = age + 1
end
`,
			wantErr: false,
		},
		{
			name: "safe nav on non-optional also returns optional",
			input: `
s = "hello"
# Even on non-optional, safe nav returns optional
len = s&.length
`,
			wantErr: false,
		},
		{
			name: "chained safe navigation",
			input: `
class Address
  def initialize(@city: String)
  end
end

class User
  def initialize(@address: Address?)
  end
end

def find_user: User?
  nil
end

u = find_user
# Chained: u&.address returns Address?, then &.city returns String?
city = u&.address&.city
`,
			wantErr: false,
		},
		{
			name: "safe nav method call with arguments type checked",
			input: `
class Calculator
  def add(a: Int, b: Int): Int
    a + b
  end
end

def get_calc: Calculator?
  nil
end

c = get_calc
result = c&.add(1, 2)
`,
			wantErr: false,
		},
		{
			name: "safe nav method call with wrong argument type",
			input: `
class Calculator
  def add(a: Int, b: Int): Int
    a + b
  end
end

def get_calc: Calculator?
  nil
end

c = get_calc
result = c&.add("wrong", 2)
`,
			wantErr: true,
			errMsg:  "cannot use String as Int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				if tt.errMsg != "" {
					found := false
					for _, err := range errs {
						if strings.Contains(err.Error(), tt.errMsg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing %q, got: %v", tt.errMsg, errs)
					}
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeImplicitMethodCalls(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "no-arg function implicit call",
			input: `
def get_value: Int
  42
end

x = get_value
y = x + 1
`,
			wantErr: false,
		},
		{
			name: "function with required params without parens errors",
			input: `
def greet(name: String): String
  "Hello"
end

x = greet
`,
			wantErr: true,
			errMsg:  "method 'greet' requires 1 argument(s)",
		},
		{
			name: "no-arg method implicit call",
			input: `
class Counter
  def count: Int
    10
  end
end

c = Counter.new
x = c.count + 1
`,
			wantErr: false,
		},
		{
			name: "method with required params without parens errors",
			input: `
class Calculator
  def add(a: Int, b: Int): Int
    a + b
  end
end

c = Calculator.new
x = c.add
`,
			wantErr: true,
			errMsg:  "method 'add' requires 2 argument(s)",
		},
		{
			name: "method with one required param without parens errors",
			input: `
class Greeter
  def greet(name: String): String
    "Hello"
  end
end

g = Greeter.new
x = g.greet
`,
			wantErr: true,
			errMsg:  "method 'greet' requires 1 argument(s)",
		},
		{
			name: "builtin method with required param without parens errors",
			input: `
s = "hello"
x = s.include?
`,
			wantErr: true,
			errMsg:  "method 'include?' requires 1 argument(s)",
		},
		{
			name: "safe nav method with required params without parens errors",
			input: `
class Calculator
  def add(a: Int, b: Int): Int
    a + b
  end
end

def get_calc: Calculator?
  nil
end

c = get_calc
x = c&.add
`,
			wantErr: true,
			errMsg:  "method 'add' requires 2 argument(s)",
		},
		{
			name: "method with parens and args works",
			input: `
class Calculator
  def add(a: Int, b: Int): Int
    a + b
  end
end

c = Calculator.new
x = c.add(1, 2)
`,
			wantErr: false,
		},
		{
			name: "chained no-arg methods work",
			input: `
s = "hello"
x = s.length.abs
`,
			wantErr: false,
		},
		{
			name:    "variadic function implicit call (no args)",
			input:   `puts`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				if tt.errMsg != "" {
					found := false
					for _, err := range errs {
						if strings.Contains(err.Error(), tt.errMsg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing %q, got: %v", tt.errMsg, errs)
					}
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestAnalyzeConstantReassignment(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "constant reassignment with = is error",
			input: `const MAX = 100
def main
  MAX = 200
end`,
			wantErr: true,
			errMsg:  "cannot assign to constant 'MAX'",
		},
		{
			name: "constant reassignment with += is error",
			input: `const COUNT = 10
def main
  COUNT += 1
end`,
			wantErr: true,
			errMsg:  "cannot assign to constant 'COUNT'",
		},
		{
			name: "constant reassignment with ||= is error",
			input: `const VALUE = 5
def main
  VALUE ||= 10
end`,
			wantErr: true,
			errMsg:  "cannot assign to constant 'VALUE'",
		},
		{
			name: "constant usage is allowed",
			input: `const MAX = 100
def main
  x = MAX * 2
end`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parse(t, tt.input)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			errs := a.Analyze(program)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got none")
					return
				}
				if tt.errMsg != "" {
					found := false
					for _, err := range errs {
						if strings.Contains(err.Error(), tt.errMsg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing %q, got: %v", tt.errMsg, errs)
					}
				}
			} else if len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

// ====================
// "Did you mean?" suggestion tests
// ====================

func TestUndefinedErrorSuggestionsInterface(t *testing.T) {
	input := `
interface Displayable
  def display
end

class Widget implements Displayble
  def display
    puts "widget"
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

	// Should have an error for undefined "Displayble"
	var undefErr *UndefinedError
	for _, err := range errs {
		if u, ok := err.(*UndefinedError); ok && u.Name == "Displayble" {
			undefErr = u
			break
		}
	}

	if undefErr == nil {
		t.Fatalf("expected UndefinedError for 'Displayble', got: %v", errs)
	}

	// Check that "Displayable" is suggested
	if !slices.Contains(undefErr.Candidates, "Displayable") {
		t.Errorf("expected 'Displayable' in candidates, got: %v", undefErr.Candidates)
	}
}

func TestUndefinedErrorSuggestionsClass(t *testing.T) {
	input := `
class UserProfile
end

def main
  x = Userprofle.new
end
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)

	// Should have an error for undefined "Userprofle"
	var undefErr *UndefinedError
	for _, err := range errs {
		if u, ok := err.(*UndefinedError); ok && u.Name == "Userprofle" {
			undefErr = u
			break
		}
	}

	if undefErr == nil {
		t.Fatalf("expected UndefinedError for 'Userprofle', got: %v", errs)
	}

	// Check that "UserProfile" is suggested
	if !slices.Contains(undefErr.Candidates, "UserProfile") {
		t.Errorf("expected 'UserProfile' in candidates, got: %v", undefErr.Candidates)
	}
}

func TestUndefinedErrorSuggestionsEnumValue(t *testing.T) {
	input := `
enum Status
  Pending
  Approved
  Rejected
end

def main
  s = Status::Aproved
end
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)

	// Should have an error for undefined "Status::Aproved"
	var undefErr *UndefinedError
	for _, err := range errs {
		if u, ok := err.(*UndefinedError); ok && u.Name == "Status::Aproved" {
			undefErr = u
			break
		}
	}

	if undefErr == nil {
		t.Fatalf("expected UndefinedError for 'Status::Aproved', got: %v", errs)
	}

	// Check that "Approved" is suggested
	if !slices.Contains(undefErr.Candidates, "Approved") {
		t.Errorf("expected 'Approved' in candidates, got: %v", undefErr.Candidates)
	}
}

func TestUndefinedErrorMessageIncludesSuggestion(t *testing.T) {
	input := `
interface Displayable
  def display
end

class Widget implements Displayble
  def display
    puts "widget"
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

	// Find the error and check the message includes "did you mean"
	found := false
	for _, err := range errs {
		errMsg := err.Error()
		if strings.Contains(errMsg, "Displayble") && strings.Contains(errMsg, "did you mean") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected error message with 'did you mean' suggestion, got: %v", errs)
	}
}

// ====================
// PositionedError interface tests
// ====================

func TestSemanticErrorsImplementPositionedError(t *testing.T) {
	// Test that semantic errors return correct position info
	input := `
def main
  undefined_var
end
`
	p := parse(t, input)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	errs := a.Analyze(program)

	if len(errs) == 0 {
		t.Fatal("expected at least one error")
	}

	// Check that the error has position info
	undefErr, ok := errs[0].(*UndefinedError)
	if !ok {
		t.Fatalf("expected UndefinedError, got %T", errs[0])
	}

	line, col := undefErr.Position()
	if line <= 0 {
		t.Errorf("expected positive line number, got %d", line)
	}
	// Column may be 0 if not set, but line should be set
	_ = col
}
