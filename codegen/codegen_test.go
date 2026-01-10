package codegen

import (
	"strings"
	"testing"

	"rugby/lexer"
	"rugby/parser"
)

func TestGenerateHello(t *testing.T) {
	input := `def main
  puts "hello"
end`

	output := compile(t, input)

	assertContains(t, output, `package main`)
	assertContains(t, output, `import`)
	assertContains(t, output, `"rugby/runtime"`)
	assertContains(t, output, `func main()`)
	assertContains(t, output, `runtime.Puts("hello")`)
}

func TestGenerateArithmetic(t *testing.T) {
	input := `def main
  x = 2 + 3 * 4
  puts x
end`

	output := compile(t, input)

	assertContains(t, output, `x :=`)
	assertContains(t, output, `(2 + (3 * 4))`)
	assertContains(t, output, `runtime.Puts(x)`)
}

func TestGenerateIfElse(t *testing.T) {
	input := `def main
  x = 5
  if x > 3
    puts "big"
  else
    puts "small"
  end
end`

	output := compile(t, input)

	assertContains(t, output, `if x > 3 {`)
	assertContains(t, output, `} else {`)
	assertContains(t, output, `runtime.Puts("big")`)
	assertContains(t, output, `runtime.Puts("small")`)
}

func TestGenerateWhile(t *testing.T) {
	input := `def main
  i = 0
  while i < 5
    puts i
    i = i + 1
  end
end`

	output := compile(t, input)

	assertContains(t, output, `i := 0`)
	assertContains(t, output, `for i < 5 {`)
	assertContains(t, output, `runtime.Puts(i)`)
	assertContains(t, output, `i = (i + 1)`)
}

func TestGenerateComparison(t *testing.T) {
	input := `def main
  x = 5 == 5
  y = 3 != 4
  z = 1 < 2 and 3 > 2
end`

	output := compile(t, input)

	assertContains(t, output, `x := (5 == 5)`)
	assertContains(t, output, `y := (3 != 4)`)
	assertContains(t, output, `((1 < 2) && (3 > 2))`)
}

func TestGenerateBoolean(t *testing.T) {
	input := `def main
  x = true
  y = false
  z = not x
end`

	output := compile(t, input)

	assertContains(t, output, `x := true`)
	assertContains(t, output, `y := false`)
	assertContains(t, output, `z := !x`)
}

func TestGenerateReturn(t *testing.T) {
	input := `def main
  return 42
end`

	output := compile(t, input)

	assertContains(t, output, `return 42`)
}

func TestGenerateFunctionParams(t *testing.T) {
	input := `def add(a, b)
  x = a
  a = b
  return a
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func add(a interface{}, b interface{})`)
	assertContains(t, output, `x := a`) // new var uses :=
	assertContains(t, output, `a = b`)  // param reassignment uses =
	assertContains(t, output, `func main()`)
}

func TestGenerateReturnType(t *testing.T) {
	input := `def add(a, b) -> Int
  return 42
end

def greet() -> String
  return "hello"
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func add(a interface{}, b interface{}) int`)
	assertContains(t, output, `func greet() string`)
	assertContains(t, output, `func main()`)
}

func TestGenerateMultipleReturnTypes(t *testing.T) {
	input := `def parse(s) -> (Int, Bool)
  return 42, true
end

def fetch() -> (String, Int, Bool)
  return "hello", 1, false
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func parse(s interface{}) (int, bool)`)
	assertContains(t, output, `return 42, true`)
	assertContains(t, output, `func fetch() (string, int, bool)`)
	assertContains(t, output, `return "hello", 1, false`)
}

func TestVariableReassignment(t *testing.T) {
	input := `def main
  x = 1
  x = 2
  x = 3
end`

	output := compile(t, input)

	// First assignment uses :=, subsequent use =
	if strings.Count(output, "x :=") != 1 {
		t.Errorf("expected exactly 1 ':=' for x, got output:\n%s", output)
	}
	if strings.Count(output, "x = ") != 2 {
		t.Errorf("expected exactly 2 '=' for x reassignment, got output:\n%s", output)
	}
}

func compile(t *testing.T, input string) string {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	gen := New()
	output, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	return output
}

func assertContains(t *testing.T, output, substr string) {
	if !strings.Contains(output, substr) {
		t.Errorf("expected output to contain %q, got:\n%s", substr, output)
	}
}

func TestGenerateImportAlias(t *testing.T) {
	input := `import encoding/json as json

def main
end`

	output := compile(t, input)

	assertContains(t, output, `json "encoding/json"`)
}

func TestGenerateSelectorExpr(t *testing.T) {
	input := `import net/http

def main
  http.Get("http://example.com")
end`

	output := compile(t, input)

	assertContains(t, output, `http.Get("http://example.com")`)
}

func TestGenerateChainedSelector(t *testing.T) {
	input := `def main
  x = resp.Body
end`

	output := compile(t, input)

	assertContains(t, output, `resp.Body`)
}

func TestGenerateSnakeCaseMapping(t *testing.T) {
	input := `import io

def main
  io.read_all(r)
end`

	output := compile(t, input)

	assertContains(t, output, `io.ReadAll(r)`)
}

func TestGenerateDefer(t *testing.T) {
	input := `def main
  defer resp.Body.Close
end`

	output := compile(t, input)

	assertContains(t, output, `defer resp.Body.Close()`)
}

func TestGenerateDeferWithParens(t *testing.T) {
	input := `def main
  defer file.Close()
end`

	output := compile(t, input)

	assertContains(t, output, `defer file.Close()`)
}

func TestGenerateDeferSimple(t *testing.T) {
	input := `def main
  defer cleanup
end`

	output := compile(t, input)

	// Local function names are not transformed (no pub support yet)
	assertContains(t, output, `defer cleanup()`)
}

func TestSnakeToCamelConversion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// snake_case transforms to camelCase
		{"read_all", "readAll"},
		{"new_request", "newRequest"},
		{"read_file", "readFile"},
		{"http_server_error", "httpServerError"},
		{"do_something!", "doSomething"},
		{"is_empty?", "isEmpty"},
		// Non-snake_case passes through as-is (supports Go interop on variables)
		{"close", "close"},
		{"get", "get"},
		{"Close", "Close"},
		{"Body", "Body"},
		// Ruby-style suffixes stripped even without underscore
		{"inc!", "inc"},
		{"save!", "save"},
		{"empty?", "empty"},
		{"valid?", "valid"},
	}

	for _, tt := range tests {
		got := snakeToCamel(tt.input)
		if got != tt.expected {
			t.Errorf("snakeToCamel(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSnakeToPascalConversion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// snake_case transforms to PascalCase
		{"read_all", "ReadAll"},
		{"new_request", "NewRequest"},
		{"read_file", "ReadFile"},
		{"http_server_error", "HttpServerError"},
		// Non-snake_case passes through as-is
		{"close", "close"},
		{"get", "get"},
		{"Close", "Close"},
		{"Get", "Get"},
	}

	for _, tt := range tests {
		got := snakeToPascal(tt.input)
		if got != tt.expected {
			t.Errorf("snakeToPascal(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestGenerateArrayLiteral(t *testing.T) {
	input := `def main
  x = [1, 2, 3]
end`

	output := compile(t, input)

	assertContains(t, output, `x := []interface{}{1, 2, 3}`)
}

func TestGenerateEmptyArray(t *testing.T) {
	input := `def main
  x = []
end`

	output := compile(t, input)

	assertContains(t, output, `x := []interface{}{}`)
}

func TestGenerateArrayWithExpressions(t *testing.T) {
	input := `def main
  x = [1 + 2, 3 * 4]
end`

	output := compile(t, input)

	assertContains(t, output, `[]interface{}{(1 + 2), (3 * 4)}`)
}

func TestGenerateArrayAsArg(t *testing.T) {
	input := `def main
  puts([1, 2, 3])
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Puts([]interface{}{1, 2, 3})`)
}

func TestGenerateNestedArray(t *testing.T) {
	input := `def main
  x = [[1, 2], [3, 4]]
end`

	output := compile(t, input)

	assertContains(t, output, `[]interface{}{[]interface{}{1, 2}, []interface{}{3, 4}}`)
}

func TestGenerateArrayIndex(t *testing.T) {
	input := `def main
  x = arr[0]
end`

	output := compile(t, input)

	assertContains(t, output, `x := arr[0]`)
}

func TestGenerateArrayIndexWithExpression(t *testing.T) {
	input := `def main
  x = arr[i + 1]
end`

	output := compile(t, input)

	assertContains(t, output, `arr[(i + 1)]`)
}

func TestGenerateChainedArrayIndex(t *testing.T) {
	input := `def main
  x = matrix[0][1]
end`

	output := compile(t, input)

	assertContains(t, output, `matrix[0][1]`)
}

func TestGenerateMapLiteral(t *testing.T) {
	input := `def main
  x = {"a" => 1, "b" => 2}
end`

	output := compile(t, input)

	assertContains(t, output, `map[interface{}]interface{}{"a": 1, "b": 2}`)
}

func TestGenerateEmptyMap(t *testing.T) {
	input := `def main
  x = {}
end`

	output := compile(t, input)

	assertContains(t, output, `map[interface{}]interface{}{}`)
}

func TestGenerateMapAccess(t *testing.T) {
	input := `def main
  x = m["key"]
end`

	output := compile(t, input)

	assertContains(t, output, `m["key"]`)
}

func TestGenerateEachBlock(t *testing.T) {
	input := `def main
  arr.each do |x|
    puts x
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Each(arr, func(x interface{}) {`)
	assertContains(t, output, `runtime.Puts(x)`)
}

func TestGenerateEachWithIndex(t *testing.T) {
	input := `def main
  arr.each_with_index do |v, i|
    puts v
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.EachWithIndex(arr, func(v interface{}, i int) {`)
	assertContains(t, output, `runtime.Puts(v)`)
}

func TestGenerateMapBlock(t *testing.T) {
	input := `def main
  result = arr.map do |x|
    x * 2
  end
end`

	output := compile(t, input)

	// Map generates runtime.Map() call with function literal
	assertContains(t, output, `runtime.Map(arr, func(x interface{}) interface{}`)
	assertContains(t, output, `return (x * 2)`)
}

func TestBlockWithNoParams(t *testing.T) {
	input := `def main
  items.each do ||
    puts "hello"
  end
end`

	output := compile(t, input)

	// Block with no params should use _ for the parameter
	assertContains(t, output, `runtime.Each(items, func(_ interface{}) {`)
	assertContains(t, output, `runtime.Puts("hello")`)
}

func TestBlockOnMethodCall(t *testing.T) {
	input := `def main
  get_items().each do |x|
    puts x
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Each(get_items(), func(x interface{}) {`)
	assertContains(t, output, `runtime.Puts(x)`)
}

func TestNestedBlocks(t *testing.T) {
	input := `def main
  matrix.each do |row|
    row.each do |x|
      puts x
    end
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Each(matrix, func(row interface{}) {`)
	assertContains(t, output, `runtime.Each(row, func(x interface{}) {`)
	assertContains(t, output, `runtime.Puts(x)`)
}

func TestSelectBlock(t *testing.T) {
	input := `def main
  evens = nums.select do |n|
    n % 2 == 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Select(nums, func(n interface{}) bool`)
	assertContains(t, output, `runtime.Equal((n % 2), 0)`)
}

func TestRejectBlock(t *testing.T) {
	input := `def main
  odds = nums.reject do |n|
    n % 2 == 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Reject(nums, func(n interface{}) bool`)
	assertContains(t, output, `runtime.Equal((n % 2), 0)`)
}

func TestReduceBlock(t *testing.T) {
	input := `def main
  sum = nums.reduce(0) do |acc, n|
    acc + n
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Reduce(nums, 0, func(acc interface{}, n interface{}) interface{}`)
	assertContains(t, output, `return (acc + n)`)
}

func TestFindBlock(t *testing.T) {
	input := `def main
  first_even = nums.find do |n|
    n % 2 == 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Find(nums, func(n interface{}) bool`)
	assertContains(t, output, `runtime.Equal((n % 2), 0)`)
}

func TestAnyBlock(t *testing.T) {
	input := `def main
  has_even = nums.any? do |n|
    n % 2 == 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Any(nums, func(n interface{}) bool`)
	assertContains(t, output, `runtime.Equal((n % 2), 0)`)
}

func TestAllBlock(t *testing.T) {
	input := `def main
  all_positive = nums.all? do |n|
    n > 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.All(nums, func(n interface{}) bool`)
	assertContains(t, output, `(n > 0)`)
}

func TestNoneBlock(t *testing.T) {
	input := `def main
  no_negatives = nums.none? do |n|
    n < 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.None(nums, func(n interface{}) bool`)
	assertContains(t, output, `(n < 0)`)
}

func TestKernelFunctions(t *testing.T) {
	input := `def main
  puts "hello"
  print "world"
  p x
  name = gets
  exit(1)
  sleep(2)
  n = rand(10)
  f = rand
end`

	output := compile(t, input)

	assertContains(t, output, `"rugby/runtime"`)
	assertContains(t, output, `runtime.Puts("hello")`)
	assertContains(t, output, `runtime.Print("world")`)
	assertContains(t, output, `runtime.P(x)`)
	assertContains(t, output, `name := runtime.Gets()`)
	assertContains(t, output, `runtime.Exit(1)`)
	assertContains(t, output, `runtime.Sleep(2)`)
	assertContains(t, output, `n := runtime.RandInt(10)`)
	assertContains(t, output, `f := runtime.RandFloat()`)
}

func TestTimesBlock(t *testing.T) {
	input := `def main
  5.times do |i|
    puts i
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Times(5, func(i int) {`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestTimesBlockWithExpression(t *testing.T) {
	input := `def main
  n.times do |i|
    puts i
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Times(n, func(i int) {`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestUptoBlock(t *testing.T) {
	input := `def main
  1.upto(5) do |i|
    puts i
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Upto(1, 5, func(i int) {`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestUptoBlockWithVariables(t *testing.T) {
	input := `def main
  start.upto(finish) do |i|
    puts i
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Upto(start, finish, func(i int) {`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestDowntoBlock(t *testing.T) {
	input := `def main
  5.downto(1) do |i|
    puts i
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Downto(5, 1, func(i int) {`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestDowntoBlockWithVariables(t *testing.T) {
	input := `def main
  high.downto(low) do |i|
    puts i
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Downto(high, low, func(i int) {`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestTimesBlockNoParam(t *testing.T) {
	input := `def main
  3.times do ||
    puts "hello"
  end
end`

	output := compile(t, input)

	// Should generate runtime.Times with _ parameter
	assertContains(t, output, `runtime.Times(3, func(_ int) {`)
	assertContains(t, output, `runtime.Puts("hello")`)
}

func TestUptoBlockNoParam(t *testing.T) {
	input := `def main
  1.upto(3) do ||
    puts "hello"
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Upto(1, 3, func(_ int) {`)
	assertContains(t, output, `runtime.Puts("hello")`)
}

func TestDowntoBlockNoParam(t *testing.T) {
	input := `def main
  3.downto(1) do ||
    puts "hello"
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Downto(3, 1, func(_ int) {`)
	assertContains(t, output, `runtime.Puts("hello")`)
}

func TestBraceBlockCodegen(t *testing.T) {
	input := `def main
  arr.each {|x| puts x }
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Each(arr, func(x interface{}) {`)
	assertContains(t, output, `runtime.Puts(x)`)
}

func TestBraceBlockMapCodegen(t *testing.T) {
	input := `def main
  result = arr.map {|x| x * 2 }
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Map(arr, func(x interface{}) interface{}`)
	assertContains(t, output, `return (x * 2)`)
}

func TestBraceBlockTimesCodegen(t *testing.T) {
	input := `def main
  5.times {|i| puts i }
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Times(5, func(i int) {`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestEmptyClassCodegen(t *testing.T) {
	input := `class User
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `type User struct{}`)
}

func TestClassWithMethodCodegen(t *testing.T) {
	input := `class User
  def greet
    puts "hello"
  end
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `type User struct{}`)
	assertContains(t, output, `func (u *User) greet()`)
	assertContains(t, output, `runtime.Puts("hello")`)
}

func TestClassMethodWithParams(t *testing.T) {
	input := `class Calculator
  def add(a, b) -> Int
    return a + b
  end
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `type Calculator struct{}`)
	assertContains(t, output, `func (c *Calculator) add(a interface{}, b interface{}) int`)
	assertContains(t, output, `return (a + b)`)
}

func TestClassMethodSnakeCaseToCamelCase(t *testing.T) {
	input := `class User
  def get_name -> String
    "test"
  end
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func (u *User) getName() string`)
}

func TestClassWithMultipleMethods(t *testing.T) {
	input := `class Counter
  def inc
    puts "inc"
  end

  def dec
    puts "dec"
  end
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `type Counter struct{}`)
	assertContains(t, output, `func (c *Counter) inc()`)
	assertContains(t, output, `func (c *Counter) dec()`)
}

func TestClassWithEmbedding(t *testing.T) {
	input := `class Service < Logger
  def run
    puts "running"
  end
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `type Service struct {`)
	assertContains(t, output, "Logger")
	assertContains(t, output, `func (s *Service) run()`)
}

func TestClassWithMultipleEmbedding(t *testing.T) {
	input := `class Service < Logger, Authenticator
  def run
    puts "running"
  end
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `type Service struct {`)
	assertContains(t, output, "Logger")
	assertContains(t, output, "Authenticator")
	assertContains(t, output, `func (s *Service) run()`)
}

func TestClassWithEmbeddingAndFields(t *testing.T) {
	input := `class Service < Logger
  def initialize(name : String)
    @name = name
  end
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `type Service struct {`)
	assertContains(t, output, "Logger")
	assertContains(t, output, "name string")
}

func TestClassWithInstanceVariables(t *testing.T) {
	input := `class User
  def initialize(name, age)
    @name = name
    @age = age
  end
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `type User struct {`)
	assertContains(t, output, `name interface{}`)
	assertContains(t, output, `age interface{}`)
	assertContains(t, output, `func newUser(name interface{}, age interface{}) *User`)
	assertContains(t, output, `u := &User{}`)
	assertContains(t, output, `u.name = name`)
	assertContains(t, output, `u.age = age`)
	assertContains(t, output, `return u`)
}

func TestClassMethodAccessingInstanceVar(t *testing.T) {
	input := `class User
  def initialize(name)
    @name = name
  end

  def get_name -> String
    return @name
  end
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func (u *User) getName() string`)
	assertContains(t, output, `return u.name`)
}

func TestClassNewSyntax(t *testing.T) {
	input := `class User
  def initialize(name)
    @name = name
  end
end

def main
  user = User.new("Alice")
end`

	output := compile(t, input)

	assertContains(t, output, `func newUser(name interface{}) *User`)
	assertContains(t, output, `user := newUser("Alice")`)
}

func TestClassNewWithMultipleArgs(t *testing.T) {
	input := `class Point
  def initialize(x, y)
    @x = x
    @y = y
  end
end

def main
  p = Point.new(10, 20)
end`

	output := compile(t, input)

	assertContains(t, output, `func newPoint(x interface{}, y interface{}) *Point`)
	assertContains(t, output, `p := newPoint(10, 20)`)
}

func TestGenerateTypedVariable(t *testing.T) {
	input := `def main
  x : Int = 5
end`

	output := compile(t, input)

	assertContains(t, output, `var x int = 5`)
}

func TestGenerateTypedParams(t *testing.T) {
	input := `def add(a : Int, b : Int) -> Int
  return a
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func add(a int, b int) int`)
}

func TestGenerateMixedParams(t *testing.T) {
	input := `def foo(a : Int, b, c : String)
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func foo(a int, b interface{}, c string)`)
}

func TestGenerateTypedInstanceVars(t *testing.T) {
	input := `class User
  def initialize(name : String, age : Int)
    @name = name
    @age = age
  end
end

def main
end`

	output := compile(t, input)

	// Check struct fields have inferred types
	assertContains(t, output, `name string`)
	assertContains(t, output, `age int`)
	// Check constructor has typed params (non-pub class uses camelCase)
	assertContains(t, output, `func newUser(name string, age int) *User`)
}

func TestGenerateTypedMethodParams(t *testing.T) {
	input := `class Calculator
  def initialize
  end

  def add(a : Int, b : Int) -> Int
    return a
  end
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func (c *Calculator) add(a int, b int) int`)
}

func TestGenerateUntypedStillWorks(t *testing.T) {
	input := `def add(a, b)
  return a
end

def main
  x = 5
end`

	output := compile(t, input)

	assertContains(t, output, `func add(a interface{}, b interface{})`)
	assertContains(t, output, `x := 5`)
}

func TestGenerateTypedReassignment(t *testing.T) {
	input := `def main
  x : Int = 5
  x = 10
end`

	output := compile(t, input)

	assertContains(t, output, `var x int = 5`)
	// Second assignment should use = not :=
	if strings.Count(output, "var x int") != 1 {
		t.Errorf("expected exactly 1 'var x int', got output:\n%s", output)
	}
	assertContains(t, output, `x = 10`)
}

func TestGenerateForLoop(t *testing.T) {
	input := `def main
  for item in items
    puts item
  end
end`

	output := compile(t, input)

	assertContains(t, output, `for _, item := range items {`)
	assertContains(t, output, `runtime.Puts(item)`)
}

func TestGenerateBreak(t *testing.T) {
	input := `def main
  while true
    break
  end
end`

	output := compile(t, input)

	assertContains(t, output, `for true {`)
	assertContains(t, output, `break`)
}

func TestGenerateNext(t *testing.T) {
	input := `def main
  for item in items
    next
  end
end`

	output := compile(t, input)

	assertContains(t, output, `for _, item := range items {`)
	assertContains(t, output, `continue`)
}

func TestForLoopWithControlFlow(t *testing.T) {
	input := `def main
  for item in items
    if item == 5
      break
    end
    if item == 3
      next
    end
    puts item
  end
end`

	output := compile(t, input)

	assertContains(t, output, `for _, item := range items {`)
	assertContains(t, output, `if runtime.Equal(item, 5) {`)
	assertContains(t, output, `break`)
	assertContains(t, output, `if runtime.Equal(item, 3) {`)
	assertContains(t, output, `continue`)
	assertContains(t, output, `runtime.Puts(item)`)
}

func TestClassMethodWithBangSuffix(t *testing.T) {
	input := `class Counter
  def inc!
    puts "incremented"
  end
end

def main
end`

	output := compile(t, input)

	// Method name should have ! stripped and be camelCased (lowercase first letter)
	assertContains(t, output, `func (c *Counter) inc()`)
}

func TestClassMethodWithPredicateSuffix(t *testing.T) {
	input := `class User
  def valid?
    return true
  end
end

def main
end`

	output := compile(t, input)

	// Method name should have ? stripped and be camelCased (lowercase first letter)
	assertContains(t, output, `func (u *User) valid()`)
}

func TestRugbyMethodCallCasing(t *testing.T) {
	input := `class User
  def get_name -> String
    return "test"
  end
end

def main
  user = User.new()
  name = user.get_name()
end`

	output := compile(t, input)

	// Method definition should use camelCase
	assertContains(t, output, `func (u *User) getName() string`)
	// Method call should also use camelCase (matching the definition)
	assertContains(t, output, `user.getName()`)
}

func TestGoInteropVsRugbyMethodCasing(t *testing.T) {
	input := `import io

class Reader
  def read_all -> String
    return "data"
  end
end

def main
  io.read_all(r)
  reader = Reader.new()
  reader.read_all()
end`

	output := compile(t, input)

	// Go import uses PascalCase
	assertContains(t, output, `io.ReadAll(r)`)
	// Rugby method definition uses camelCase
	assertContains(t, output, `func (r *Reader) readAll() string`)
	// Rugby method call uses camelCase
	assertContains(t, output, `reader.readAll()`)
}

func TestSelfKeyword(t *testing.T) {
	input := `class Builder
  def initialize
  end

  def with_name(n)
    @name = n
    self
  end

  def build
    return self
  end
end

def main
end`

	output := compile(t, input)

	// 'self' should compile to receiver variable 'b' (first letter of Builder)
	assertContains(t, output, `func (b *Builder) withName(n interface{})`)
	// Implicit return of self
	assertContains(t, output, "return b\n}")
	// Explicit return of self
	assertContains(t, output, "return b")
}

func TestSelfInMethodChain(t *testing.T) {
	input := `class Config
  def initialize
  end

  def set_value(v)
    @value = v
    self
  end
end

def main
  c = Config.new()
  c.set_value(1).set_value(2)
end`

	output := compile(t, input)

	// Method should return self (receiver 'c')
	assertContains(t, output, "return c\n}")
	// Method chaining should work
	assertContains(t, output, `c.setValue(1).setValue(2)`)
}

func TestSelfOutsideClass(t *testing.T) {
	input := `def main
  x = self
end`

	output := compile(t, input)

	// self outside class should generate a comment indicating the error
	assertContains(t, output, `/* self outside class */`)
}

func TestToSMethod(t *testing.T) {
	input := `class User
  def initialize(name : String)
    @name = name
  end

  def to_s
    return @name
  end
end

def main
end`

	output := compile(t, input)

	// to_s should compile to String() string (satisfies fmt.Stringer)
	assertContains(t, output, `func (u *User) String() string {`)
	assertContains(t, output, `return u.name`)
}

func TestToSWithInterpolation(t *testing.T) {
	input := `class Point
  def initialize(x : Int, y : Int)
    @x = x
    @y = y
  end

  def to_s
    return "point"
  end
end

def main
end`

	output := compile(t, input)

	// to_s should compile to String() string
	assertContains(t, output, `func (p *Point) String() string {`)
}

func TestToSWithParamsFallsBack(t *testing.T) {
	// to_s with parameters doesn't satisfy fmt.Stringer, so it's treated as a normal method
	input := `class User
  def initialize
  end

  def to_s(format : String)
    return format
  end
end

def main
end`

	output := compile(t, input)

	// With parameters, to_s becomes toS (normal snake_case conversion)
	assertContains(t, output, `func (u *User) toS(format string) {`)
}

func TestStringInterpolation(t *testing.T) {
	input := `def main
  name = "world"
  x = "hello #{name}"
end`

	output := compile(t, input)

	// Should generate fmt.Sprintf
	assertContains(t, output, `fmt.Sprintf("hello %v", name)`)
	// Should import fmt
	assertContains(t, output, `"fmt"`)
}

func TestStringInterpolationWithExpression(t *testing.T) {
	input := `def main
  x = "sum: #{1 + 2}"
end`

	output := compile(t, input)

	// Should generate fmt.Sprintf with expression
	assertContains(t, output, `fmt.Sprintf("sum: %v", (1 + 2))`)
}

func TestStringInterpolationMultiple(t *testing.T) {
	input := `def main
  a = "foo"
  b = "bar"
  x = "#{a} and #{b}"
end`

	output := compile(t, input)

	// Should generate fmt.Sprintf with multiple args
	assertContains(t, output, `fmt.Sprintf("%v and %v", a, b)`)
}

func TestStringInterpolationWithInstanceVar(t *testing.T) {
	input := `class User
  def initialize(name : String)
    @name = name
  end

  def to_s
    return "User: #{@name}"
  end
end

def main
end`

	output := compile(t, input)

	// Should generate fmt.Sprintf with instance var
	assertContains(t, output, `fmt.Sprintf("User: %v", u.name)`)
}

func TestPlainStringNoInterpolation(t *testing.T) {
	input := `def main
  x = "hello world"
end`

	output := compile(t, input)

	// Plain string should remain as literal, no fmt.Sprintf
	assertContains(t, output, `x := "hello world"`)
	// Should NOT import fmt for plain strings (unless other code needs it)
}

func TestStringInterpolationWithPercent(t *testing.T) {
	input := `def main
  name = "test"
  x = "100% of #{name}"
end`

	output := compile(t, input)

	// Percent should be escaped as %% for fmt.Sprintf
	assertContains(t, output, `fmt.Sprintf("100%% of %v", name)`)
}

func TestCustomEqualityMethod(t *testing.T) {
	input := `class Point
  def initialize(x : Int, y : Int)
    @x = x
    @y = y
  end

  def ==(other)
    @x == other.x and @y == other.y
  end
end

def main
end`

	output := compile(t, input)

	// def == should compile to Equal(other interface{}) bool
	assertContains(t, output, `func (p *Point) Equal(other interface{}) bool {`)
	// Should include type assertion
	assertContains(t, output, `other, ok := other.(*Point)`)
	assertContains(t, output, `if !ok {`)
	assertContains(t, output, `return false`)
	// Should have implicit return for the last expression
	assertContains(t, output, `return (runtime.Equal(p.x, other.x) && runtime.Equal(p.y, other.y))`)
}

func TestEqualityWithVariables(t *testing.T) {
	input := `def main
  a = User.new("alice")
  b = User.new("bob")
  x = a == b
end`

	output := compile(t, input)

	// Variable equality should use runtime.Equal
	assertContains(t, output, `runtime.Equal(a, b)`)
	// Should import runtime
	assertContains(t, output, `"rugby/runtime"`)
}

func TestEqualityWithLiteralsStaysDirect(t *testing.T) {
	input := `def main
  x = 5 == 5
  y = "a" == "b"
end`

	output := compile(t, input)

	// Literal-to-literal comparisons should use direct ==
	assertContains(t, output, `x := (5 == 5)`)
	assertContains(t, output, `y := ("a" == "b")`)
}

func TestNotEqualWithVariables(t *testing.T) {
	input := `def main
  a = User.new("alice")
  b = User.new("bob")
  x = a != b
end`

	output := compile(t, input)

	// Variable inequality should use !runtime.Equal
	assertContains(t, output, `!runtime.Equal(a, b)`)
}

func TestMixedLiteralVariableComparison(t *testing.T) {
	input := `def main
  x = 5
  y = x == 5
  z = "hello" == name
end`

	output := compile(t, input)

	// Mixed comparisons (variable vs literal) should use runtime.Equal
	assertContains(t, output, `runtime.Equal(x, 5)`)
	assertContains(t, output, `runtime.Equal("hello", name)`)
}

func TestInterfaceDeclaration(t *testing.T) {
	input := `interface Speaker
  def speak -> String
end

def main
end`

	output := compile(t, input)

	// Interface should be generated with PascalCase method names
	assertContains(t, output, `type Speaker interface {`)
	assertContains(t, output, `Speak() string`)
	assertContains(t, output, `}`)
}

func TestInterfaceWithMultipleMethods(t *testing.T) {
	input := `interface ReadWriter
  def read(n : Int) -> String
  def write(data : String) -> Int
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `type ReadWriter interface {`)
	assertContains(t, output, `Read(int) string`)
	assertContains(t, output, `Write(string) int`)
}

func TestInterfaceMethodWithSnakeCase(t *testing.T) {
	input := `interface Handler
  def handle_request(req : String) -> String
end

def main
end`

	output := compile(t, input)

	// Interface methods should be PascalCase
	assertContains(t, output, `HandleRequest(string) string`)
}

func TestPubFunctionExported(t *testing.T) {
	input := `pub def parse_json(s : String) -> String
  s
end

def main
end`

	output := compile(t, input)

	// pub function should be PascalCase with acronym handling
	assertContains(t, output, `func ParseJSON(s string) string`)
}

func TestNonPubFunctionCamelCase(t *testing.T) {
	input := `def parse_json(s : String) -> String
  s
end

def main
end`

	output := compile(t, input)

	// non-pub function should be camelCase with acronym handling
	assertContains(t, output, `func parseJSON(s string) string`)
}

func TestPubClassWithPubMethod(t *testing.T) {
	input := `pub class User
  def initialize(name : String)
    @name = name
  end

  pub def get_name -> String
    @name
  end

  def internal_helper -> String
    @name
  end
end

def main
end`

	output := compile(t, input)

	// pub class should have NewClassName constructor (uppercase)
	assertContains(t, output, `func NewUser(name string) *User`)
	// pub method should be PascalCase
	assertContains(t, output, `func (u *User) GetName() string`)
	// non-pub method should be camelCase
	assertContains(t, output, `func (u *User) internalHelper() string`)
}

func TestNonPubClassWithMethods(t *testing.T) {
	input := `class User
  def initialize(name : String)
    @name = name
  end

  def get_name -> String
    @name
  end
end

def main
end`

	output := compile(t, input)

	// non-pub class should have newClassName constructor (lowercase)
	assertContains(t, output, `func newUser(name string) *User`)
	// methods should be camelCase
	assertContains(t, output, `func (u *User) getName() string`)
}

func TestAcronymHandling(t *testing.T) {
	input := `pub def get_user_id -> Int
  0
end

pub def parse_http_url -> String
  ""
end

def get_api_json -> String
  ""
end

def main
end`

	output := compile(t, input)

	// Pub functions with acronyms should have proper casing
	assertContains(t, output, `func GetUserID() int`)
	assertContains(t, output, `func ParseHTTPURL() string`)
	// Non-pub function with acronyms
	assertContains(t, output, `func getAPIJSON() string`)
}

func TestMethodWithBangSuffixStripped(t *testing.T) {
	input := `pub class Counter
  def initialize
    @n = 0
  end

  pub def inc!
    @n = @n + 1
  end

  def reset!
    @n = 0
  end
end

def main
end`

	output := compile(t, input)

	// ! suffix should be stripped
	assertContains(t, output, `func (c *Counter) Inc()`)
	assertContains(t, output, `func (c *Counter) reset()`)
}

func TestInterfaceWithSnakeCaseAcronyms(t *testing.T) {
	input := `interface HttpHandler
  def get_user_id -> Int
  def parse_json -> String
end

def main
end`

	output := compile(t, input)

	// Interface methods should be PascalCase with acronyms
	assertContains(t, output, `GetUserID() int`)
	assertContains(t, output, `ParseJSON() string`)
}

func TestRangeLiteral(t *testing.T) {
	input := `def main
  r = 1..10
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Range{Start: 1, End: 10, Exclusive: false}`)
}

func TestExclusiveRangeLiteral(t *testing.T) {
	input := `def main
  r = 0...5
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Range{Start: 0, End: 5, Exclusive: true}`)
}

func TestForLoopWithInclusiveRange(t *testing.T) {
	input := `def main
  for i in 0..5
    puts i
  end
end`

	output := compile(t, input)

	// Should generate C-style for loop, not range loop
	assertContains(t, output, `for i := 0; i <= 5; i++ {`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestForLoopWithExclusiveRange(t *testing.T) {
	input := `def main
  for i in 0...5
    puts i
  end
end`

	output := compile(t, input)

	// Should use < for exclusive range
	assertContains(t, output, `for i := 0; i < 5; i++ {`)
}

func TestForLoopWithRangeVariables(t *testing.T) {
	input := `def main
  start = 1
  finish = 10
  for i in start..finish
    puts i
  end
end`

	output := compile(t, input)

	// Should use variables in the for loop
	assertContains(t, output, `for i := start; i <= finish; i++ {`)
}

func TestForLoopWithRangeObjectVariable(t *testing.T) {
	input := `def main
  r = 1..10
  for i in r
    puts i
  end
end`

	output := compile(t, input)

	// Should generate loop using r.Start, r.End and checking r.Exclusive
	assertContains(t, output, `for i := r.Start; (r.Exclusive && i < r.End) || (!r.Exclusive && i <= r.End); i++ {`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestRangeToArray(t *testing.T) {
	input := `def main
  nums = (1..5).to_a
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.RangeToArray(runtime.Range{Start: 1, End: 5, Exclusive: false})`)
}

func TestRangeSize(t *testing.T) {
	input := `def main
  n = (1..10).size
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.RangeSize(runtime.Range{Start: 1, End: 10, Exclusive: false})`)
}

func TestRangeContains(t *testing.T) {
	input := `def main
  x = (1..10).include?(5)
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.RangeContains(runtime.Range{Start: 1, End: 10, Exclusive: false}, 5)`)
}

func TestRangeEachBlock(t *testing.T) {
	input := `def main
  (1..5).each do |i|
    puts i
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.RangeEach(runtime.Range{Start: 1, End: 5, Exclusive: false}, func(i int) {`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestEmptyRange(t *testing.T) {
	// When start > end, the loop condition fails immediately (empty range)
	input := `def main
  for i in 5..4
    puts i
  end
end`

	output := compile(t, input)

	// Verify the loop generates correct condition (will iterate 0 times)
	assertContains(t, output, `for i := 5; i <= 4; i++ {`)
}

func TestOrAssignFirstDeclaration(t *testing.T) {
	input := `def main
  x ||= 5
end`

	output := compile(t, input)

	// First use should declare with :=
	assertContains(t, output, "x := 5")
}

func TestOrAssignSecondUse(t *testing.T) {
	input := `def main
  x = nil
  x ||= 5
end`

	output := compile(t, input)

	// Second use should generate nil check
	assertContains(t, output, "if x == nil {")
	assertContains(t, output, "x = 5")
}

func TestInstanceVarOrAssign(t *testing.T) {
	input := `class Service
  def get_cache
    @cache ||= load_cache()
  end
end

def main
end`

	output := compile(t, input)

	// Should generate nil check for instance var
	assertContains(t, output, "if s.cache == nil {")
	assertContains(t, output, "s.cache = load_cache()")
}

func TestOptionalValueType(t *testing.T) {
	input := `def find(id : Int?) -> String?
end

def main
end`

	output := compile(t, input)

	// Value type optionals use runtime.OptionalT
	assertContains(t, output, "func find(id runtime.OptionalInt) runtime.OptionalString")
}

func TestOptionalReferenceType(t *testing.T) {
	input := `def find(id : Int) -> User?
end

def main
end`

	output := compile(t, input)

	// Reference type optionals use pointer
	assertContains(t, output, "func find(id int) *User")
}

func TestOptionalVariable(t *testing.T) {
	input := `def main
  x : Int? = nil
end`

	output := compile(t, input)

	// Variable with optional type should use NoneInt()
	assertContains(t, output, "var x runtime.OptionalInt = runtime.NoneInt()")
}

func TestOptionalValueTypeInCondition(t *testing.T) {
	input := `def main
  x : Int? = nil
  if x
    puts "has value"
  end
end`

	output := compile(t, input)

	// Value type optional in condition checks .Valid
	assertContains(t, output, "if x.Valid {")
}

func TestOptionalReferenceTypeInCondition(t *testing.T) {
	input := `def process(u : User?)
  if u
    puts "has user"
  end
end

def main
end`

	output := compile(t, input)

	// Reference type optional in condition checks != nil
	assertContains(t, output, "if u != nil {")
}

func TestOptionalInElsifCondition(t *testing.T) {
	input := `def main
  x : Int? = nil
  y : String? = nil
  if x
    puts "x"
  elsif y
    puts "y"
  end
end`

	output := compile(t, input)

	// Both if and elsif should use .Valid for value type optionals
	assertContains(t, output, "if x.Valid {")
	assertContains(t, output, "} else if y.Valid {")
}

func TestNonOptionalInCondition(t *testing.T) {
	input := `def main
  x : Bool = true
  if x
    puts "yes"
  end
end`

	output := compile(t, input)

	// Non-optional types should be used as-is
	assertContains(t, output, "if x {")
}

func TestNilLiteral(t *testing.T) {
	input := `def main
  x : User? = nil
end`

	output := compile(t, input)

	// nil literal should be generated as-is
	assertContains(t, output, "var x *User = nil")
}

func TestOptionalMethodToInt(t *testing.T) {
	input := `def main
  s = "42"
  n = s.to_i?()
end`

	output := compile(t, input)

	// to_i? should map to runtime.StringToInt, discarding the boolean
	assertContains(t, output, "n, _ := runtime.StringToInt(s)")
}

func TestOptionalMethodToFloat(t *testing.T) {
	input := `def main
  s = "3.14"
  n = s.to_f?()
end`

	output := compile(t, input)

	// to_f? should map to runtime.StringToFloat, discarding the boolean
	assertContains(t, output, "n, _ := runtime.StringToFloat(s)")
}

func TestAssignmentInCondition(t *testing.T) {
	input := `def main
  s = "42"
  if (n = s.to_i?())
    puts n
  end
end`

	output := compile(t, input)

	// should generate if n, ok := runtime.StringToInt(s); ok {
	assertContains(t, output, "if n, ok := runtime.StringToInt(s); ok {")
}
