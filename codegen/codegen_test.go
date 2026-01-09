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
		{"read_all", "ReadAll"},
		{"new_request", "NewRequest"},
		{"close", "Close"},
		{"get", "Get"},
		{"read_file", "ReadFile"},
		{"http_server_error", "HttpServerError"},
	}

	for _, tt := range tests {
		got := snakeToCamel(tt.input)
		if got != tt.expected {
			t.Errorf("snakeToCamel(%q) = %q, want %q", tt.input, got, tt.expected)
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

	assertContains(t, output, `for _, x := range arr {`)
	assertContains(t, output, `runtime.Puts(x)`)
}

func TestGenerateEachWithIndex(t *testing.T) {
	input := `def main
  arr.each_with_index do |v, i|
    puts v
  end
end`

	output := compile(t, input)

	// Rugby |value, index| is swapped to Go's index, value order
	assertContains(t, output, `for i, v := range arr {`)
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

	// Block with no params should use _ for range variable
	assertContains(t, output, `for _, _ := range items`)
	assertContains(t, output, `runtime.Puts("hello")`)
}

func TestBlockOnMethodCall(t *testing.T) {
	input := `def main
  get_items().each do |x|
    puts x
  end
end`

	output := compile(t, input)

	assertContains(t, output, `for _, x := range get_items()`)
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

	assertContains(t, output, `for _, row := range matrix`)
	assertContains(t, output, `for _, x := range row`)
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
	assertContains(t, output, `((n % 2) == 0)`)
}

func TestRejectBlock(t *testing.T) {
	input := `def main
  odds = nums.reject do |n|
    n % 2 == 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Reject(nums, func(n interface{}) bool`)
	assertContains(t, output, `((n % 2) == 0)`)
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
	assertContains(t, output, `((n % 2) == 0)`)
}

func TestAnyBlock(t *testing.T) {
	input := `def main
  has_even = nums.any? do |n|
    n % 2 == 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Any(nums, func(n interface{}) bool`)
	assertContains(t, output, `((n % 2) == 0)`)
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

	assertContains(t, output, `for i := 0; i < 5; i++`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestTimesBlockWithExpression(t *testing.T) {
	input := `def main
  n.times do |i|
    puts i
  end
end`

	output := compile(t, input)

	assertContains(t, output, `for i := 0; i < n; i++`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestUptoBlock(t *testing.T) {
	input := `def main
  1.upto(5) do |i|
    puts i
  end
end`

	output := compile(t, input)

	assertContains(t, output, `for i := 1; i <= 5; i++`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestUptoBlockWithVariables(t *testing.T) {
	input := `def main
  start.upto(finish) do |i|
    puts i
  end
end`

	output := compile(t, input)

	assertContains(t, output, `for i := start; i <= finish; i++`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestDowntoBlock(t *testing.T) {
	input := `def main
  5.downto(1) do |i|
    puts i
  end
end`

	output := compile(t, input)

	assertContains(t, output, `for i := 5; i >= 1; i--`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestDowntoBlockWithVariables(t *testing.T) {
	input := `def main
  high.downto(low) do |i|
    puts i
  end
end`

	output := compile(t, input)

	assertContains(t, output, `for i := high; i >= low; i--`)
	assertContains(t, output, `runtime.Puts(i)`)
}

func TestTimesBlockNoParam(t *testing.T) {
	input := `def main
  3.times do ||
    puts "hello"
  end
end`

	output := compile(t, input)

	// Should generate valid Go with synthetic variable _i
	assertContains(t, output, `for _i := 0; _i < 3; _i++`)
	assertContains(t, output, `runtime.Puts("hello")`)
}

func TestUptoBlockNoParam(t *testing.T) {
	input := `def main
  1.upto(3) do ||
    puts "hello"
  end
end`

	output := compile(t, input)

	assertContains(t, output, `for _i := 1; _i <= 3; _i++`)
	assertContains(t, output, `runtime.Puts("hello")`)
}

func TestDowntoBlockNoParam(t *testing.T) {
	input := `def main
  3.downto(1) do ||
    puts "hello"
  end
end`

	output := compile(t, input)

	assertContains(t, output, `for _i := 3; _i >= 1; _i--`)
	assertContains(t, output, `runtime.Puts("hello")`)
}

func TestBraceBlockCodegen(t *testing.T) {
	input := `def main
  arr.each {|x| puts x }
end`

	output := compile(t, input)

	assertContains(t, output, `for _, x := range arr {`)
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

	assertContains(t, output, `for i := 0; i < 5; i++`)
	assertContains(t, output, `runtime.Puts(i)`)
}
