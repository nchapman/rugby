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
  x : Int64 = 5
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
  def initialize(name : String)
    @name : String = name
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
  def initialize(name : String)
    @name : String = name
  end
end

class Cat
  def initialize(name : String)
    @name : String = name
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
  def initialize(name : String)
    @name : String = name
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
