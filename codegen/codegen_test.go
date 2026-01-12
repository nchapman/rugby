package codegen

import (
	"strings"
	"testing"

	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/parser"
)

func TestGenerateHello(t *testing.T) {
	input := `def main
  puts("hello")
end`

	output := compile(t, input)

	assertContains(t, output, `package main`)
	assertContains(t, output, `import`)
	assertContains(t, output, `"github.com/nchapman/rugby/runtime"`)
	assertContains(t, output, `func main()`)
	assertContains(t, output, `runtime.Puts("hello")`)
}

func TestGenerateArithmetic(t *testing.T) {
	input := `def main
  x = 2 + 3 * 4
  puts(x)
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
    puts("big")
  else
    puts("small")
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
    puts(i)
    i = i + 1
  end
end`

	output := compile(t, input)

	assertContains(t, output, `i := 0`)
	assertContains(t, output, `for i < 5 {`)
	assertContains(t, output, `runtime.Puts(i)`)
	assertContains(t, output, `i = (i + 1)`)
}

func TestGenerateUntil(t *testing.T) {
	input := `def main
  i = 5
  until i == 0
    puts(i)
    i = i - 1
  end
end`

	output := compile(t, input)

	assertContains(t, output, `i := 5`)
	// == is compiled to runtime.Equal for type safety
	assertContains(t, output, `for !(runtime.Equal(i, 0)) {`)
	assertContains(t, output, `runtime.Puts(i)`)
	assertContains(t, output, `i = (i - 1)`)
}

func TestGenerateUntilSimpleCondition(t *testing.T) {
	input := `def main
  until done
    work()
  end
end`

	output := compile(t, input)

	// Simple identifier condition doesn't need parentheses
	assertContains(t, output, `for !done {`)
}

func TestGenerateUntilWithLogicalOperator(t *testing.T) {
	input := `def main
  until ready and valid
    wait()
  end
end`

	output := compile(t, input)

	// Logical operators need parentheses for correct precedence
	assertContains(t, output, `for !(ready && valid) {`)
}

func TestGenerateUntilWithBreakAndNext(t *testing.T) {
	input := `def main
  until done
    next if skip
    break if found
  end
end`

	output := compile(t, input)

	// break and next should work correctly inside until loops
	assertContains(t, output, `for !done {`)
	assertContains(t, output, `if skip {`)
	assertContains(t, output, `continue`)
	assertContains(t, output, `if found {`)
	assertContains(t, output, `break`)
}

func TestGeneratePostfixWhile(t *testing.T) {
	input := `def main
  puts x while x > 0
end`

	output := compile(t, input)

	// Postfix while compiles to a regular while loop
	assertContains(t, output, `for x > 0 {`)
	assertContains(t, output, `runtime.Puts(x)`)
}

func TestGeneratePostfixUntil(t *testing.T) {
	input := `def main
  process() until done
end`

	output := compile(t, input)

	// Postfix until compiles to for !condition
	assertContains(t, output, `for !done {`)
	assertContains(t, output, `process()`)
}

func TestGeneratePostfixWhileMethodChain(t *testing.T) {
	input := `def main
  obj.foo.bar while cond
end`

	output := compile(t, input)

	// Selector expressions as statements become method calls
	assertContains(t, output, `for cond {`)
	assertContains(t, output, `obj.foo.bar()`)
}

func TestGeneratePostfixWhileCompoundCondition(t *testing.T) {
	input := `def main
  puts x while x > 0 and y < 10
end`

	output := compile(t, input)

	// Compound conditions are parsed correctly
	assertContains(t, output, `for (x > 0) && (y < 10) {`)
	assertContains(t, output, `runtime.Puts(x)`)
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

func TestGenerateSelectorAsStatement(t *testing.T) {
	input := `def main
  obj.foo
  obj.bar.baz
end`

	output := compile(t, input)

	// Selector expressions as statements become method calls (Ruby behavior)
	assertContains(t, output, `obj.foo()`)
	assertContains(t, output, `obj.bar.baz()`)
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
	input := `def add(a : any, b : any)
  x = a
  a = b
  return a
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func add(a any, b any)`)
	assertContains(t, output, `x := a`) // new var uses :=
	assertContains(t, output, `a = b`)  // param reassignment uses =
	assertContains(t, output, `func main()`)
}

func TestGenerateReturnType(t *testing.T) {
	input := `def add(a : any, b : any) -> Int
  return 42
end

def greet() -> String
  return "hello"
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func add(a any, b any) int`)
	assertContains(t, output, `func greet() string`)
	assertContains(t, output, `func main()`)
}

func TestGenerateMultipleReturnTypes(t *testing.T) {
	input := `def parse(s : any) -> (Int, Bool)
  return 42, true
end

def fetch() -> (String, Int, Bool)
  return "hello", 1, false
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func parse(s any) (int, bool)`)
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

func assertNotContains(t *testing.T, output, substr string) {
	if strings.Contains(output, substr) {
		t.Errorf("expected output NOT to contain %q, got:\n%s", substr, output)
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

func TestRugbyStdlibImport(t *testing.T) {
	input := `import rugby/http

def main
  http.Get("http://example.com")
end`

	output := compile(t, input)

	// Should transform rugby/http to full module path
	assertContains(t, output, `"github.com/nchapman/rugby/stdlib/http"`)
	assertNotContains(t, output, `"rugby/http"`)
}

func TestRugbyStdlibImportWithAlias(t *testing.T) {
	input := `import rugby/http as h

def main
  h.Get("http://example.com")
end`

	output := compile(t, input)

	// Should transform and preserve alias
	assertContains(t, output, `h "github.com/nchapman/rugby/stdlib/http"`)
}

func TestRugbyJsonImport(t *testing.T) {
	input := `import rugby/json

def main
  json.Parse("{}")
end`

	output := compile(t, input)

	assertContains(t, output, `"github.com/nchapman/rugby/stdlib/json"`)
}

func TestRugbyRuntimeImportSpecialCase(t *testing.T) {
	// rugby/runtime should go to root, not stdlib
	input := `import rugby/runtime

def main
  runtime.Puts("test")
end`

	output := compile(t, input)

	// Should NOT be in stdlib
	assertContains(t, output, `"github.com/nchapman/rugby/runtime"`)
	assertNotContains(t, output, `stdlib/runtime`)
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
		{"is_empty?", "isEmpty"},
		// Non-snake_case passes through as-is (supports Go interop on variables)
		{"close", "close"},
		{"get", "get"},
		{"Close", "Close"},
		{"Body", "Body"},
		// Ruby-style ? suffix stripped even without underscore
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

	assertContains(t, output, `x := []int{1, 2, 3}`)
}

func TestGenerateEmptyArray(t *testing.T) {
	input := `def main
  x = []
end`

	output := compile(t, input)

	assertContains(t, output, `x := []any{}`)
}

func TestGenerateArrayWithExpressions(t *testing.T) {
	input := `def main
  x = [1 + 2, 3 * 4]
end`

	output := compile(t, input)

	// Type inference correctly identifies []int from integer expressions
	assertContains(t, output, `[]int{(1 + 2), (3 * 4)}`)
}

func TestGenerateArrayAsArg(t *testing.T) {
	input := `def main
  puts([1, 2, 3])
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Puts([]int{1, 2, 3})`)
}

func TestGenerateNestedArray(t *testing.T) {
	input := `def main
  x = [[1, 2], [3, 4]]
end`

	output := compile(t, input)

	assertContains(t, output, `[]any{[]int{1, 2}, []int{3, 4}}`)
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

	// Variable index uses runtime.AtIndex to support negative indices
	assertContains(t, output, `runtime.AtIndex(arr, (i + 1))`)
}

func TestGenerateNegativeIndexLiteral(t *testing.T) {
	input := `def main
  x = arr[-1]
  y = arr[-2]
end`

	output := compile(t, input)

	// Negative literal indices use runtime.AtIndex
	assertContains(t, output, `runtime.AtIndex(arr, -1)`)
	assertContains(t, output, `runtime.AtIndex(arr, -2)`)
}

func TestGenerateNegativeStringIndex(t *testing.T) {
	input := `def main
  s = "hello"
  x = s[-1]
end`

	output := compile(t, input)

	// String negative indexing also uses runtime.AtIndex
	assertContains(t, output, `runtime.AtIndex(s, -1)`)
}

func TestGenerateVariableIndex(t *testing.T) {
	input := `def main
  x = arr[i]
end`

	output := compile(t, input)

	// Variable indices use runtime.AtIndex to support negative values
	assertContains(t, output, `runtime.AtIndex(arr, i)`)
}

func TestGenerateChainedArrayIndex(t *testing.T) {
	input := `def main
  x = matrix[0][1]
end`

	output := compile(t, input)

	assertContains(t, output, `matrix[0][1]`)
}

func TestGenerateChainedNegativeIndex(t *testing.T) {
	input := `def main
  x = matrix[-1][-1]
end`

	output := compile(t, input)

	// Negative indices chain AtIndex calls
	assertContains(t, output, `runtime.AtIndex(runtime.AtIndex(matrix, -1), -1)`)
}

func TestGenerateMixedChainedIndex(t *testing.T) {
	input := `def main
  x = matrix[0][-1]
end`

	output := compile(t, input)

	// Positive literal uses native, negative uses AtIndex
	assertContains(t, output, `runtime.AtIndex(matrix[0], -1)`)
}

func TestGenerateArrayAppend(t *testing.T) {
	input := `def main
  arr << 5
end`

	output := compile(t, input)

	// Array append uses runtime.ShiftLeft
	assertContains(t, output, `runtime.ShiftLeft(arr, 5)`)
}

func TestGenerateArrayAppendChained(t *testing.T) {
	input := `def main
  arr << 1 << 2 << 3
end`

	output := compile(t, input)

	// Chained appends nest ShiftLeft calls
	assertContains(t, output, `runtime.ShiftLeft(runtime.ShiftLeft(runtime.ShiftLeft(arr, 1), 2), 3)`)
}

func TestGenerateArrayAppendAssignment(t *testing.T) {
	input := `def main
  arr = [1, 2, 3]
  arr = arr << 5
end`

	output := compile(t, input)

	// Reassignment with append
	assertContains(t, output, `arr = runtime.ShiftLeft(arr, 5)`)
}

func TestGenerateChannelSend(t *testing.T) {
	input := `def main
  ch << 42
end`

	output := compile(t, input)

	// Channel send also uses runtime.ShiftLeft
	assertContains(t, output, `runtime.ShiftLeft(ch, 42)`)
}

func TestGenerateTernaryOperator(t *testing.T) {
	input := `def main
  x = a > b ? a : b
end`

	output := compile(t, input)

	// Ternary compiles to IIFE with if-else
	assertContains(t, output, `func() any {`)
	assertContains(t, output, `if a > b {`)
	assertContains(t, output, `return a`)
	assertContains(t, output, `return b`)
}

func TestGenerateTernaryWithStrings(t *testing.T) {
	input := `def main
  status = valid ? "ok" : "error"
end`

	output := compile(t, input)

	// Should infer string type from literals
	assertContains(t, output, `func() string {`)
	assertContains(t, output, `if valid {`)
	assertContains(t, output, `return "ok"`)
	assertContains(t, output, `return "error"`)
}

func TestGenerateTernaryNested(t *testing.T) {
	input := `def main
  x = a ? b ? 1 : 2 : 3
end`

	output := compile(t, input)

	// Nested ternary should be right-associative: a ? (b ? 1 : 2) : 3
	assertContains(t, output, `if a {`)
	assertContains(t, output, `if b {`)
	assertContains(t, output, `return 1`)
	assertContains(t, output, `return 2`)
	assertContains(t, output, `return 3`)
}

func TestGenerateTernaryWithFunctionCall(t *testing.T) {
	input := `def main
  x = foo() ? a : b
end`

	output := compile(t, input)

	// Ternary with function call condition
	assertContains(t, output, `if foo()`)
}

func TestGenerateTernaryWithMethodCalls(t *testing.T) {
	input := `def main
  x = cond ? list.first : list.last
end`

	output := compile(t, input)

	// Ternary with method calls in branches
	assertContains(t, output, `return list.first`)
	assertContains(t, output, `return list.last`)
}

func TestGenerateMapLiteral(t *testing.T) {
	input := `def main
  x = {"a" => 1, "b" => 2}
end`

	output := compile(t, input)

	assertContains(t, output, `map[any]any{"a": 1, "b": 2}`)
}

func TestGenerateEmptyMap(t *testing.T) {
	input := `def main
  x = {}
end`

	output := compile(t, input)

	assertContains(t, output, `map[any]any{}`)
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
    puts(x)
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Each(arr, func(x any) bool {`)
	assertContains(t, output, `runtime.Puts(x)`)
	assertContains(t, output, `return true`)
}

func TestGenerateEachWithIndex(t *testing.T) {
	input := `def main
  arr.each_with_index do |v, i|
    puts(v)
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.EachWithIndex(arr, func(v any, i int) bool {`)
	assertContains(t, output, `runtime.Puts(v)`)
	assertContains(t, output, `return true`)
}

func TestGenerateMapBlock(t *testing.T) {
	input := `def main
  result = arr.map do |x|
    x * 2
  end
end`

	output := compile(t, input)

	// Map generates runtime.Map() call with function literal (three-value return)
	assertContains(t, output, `runtime.Map(arr, func(x any) (any, bool, bool) {`)
	assertContains(t, output, `return (x * 2), true, true`)
}

func TestBlockWithNoParams(t *testing.T) {
	input := `def main
  items.each do ||
    puts("hello")
  end
end`

	output := compile(t, input)

	// Block with no params should use _ for the parameter
	assertContains(t, output, `runtime.Each(items, func(_ any) bool {`)
	assertContains(t, output, `runtime.Puts("hello")`)
	assertContains(t, output, `return true`)
}

func TestBlockOnMethodCall(t *testing.T) {
	input := `def main
  getItems().each do |x|
    puts(x)
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Each(getItems(), func(x any) bool {`)
	assertContains(t, output, `runtime.Puts(x)`)
	assertContains(t, output, `return true`)
}

func TestNestedBlocks(t *testing.T) {
	input := `def main
  matrix.each do |row|
    row.each do |x|
      puts(x)
    end
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Each(matrix, func(row any) bool {`)
	assertContains(t, output, `runtime.Each(row, func(x any) bool {`)
	assertContains(t, output, `runtime.Puts(x)`)
	assertContains(t, output, `return true`)
}

func TestSelectBlock(t *testing.T) {
	input := `def main
  evens = nums.select do |n|
    n % 2 == 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Select(nums, func(n any) (bool, bool) {`)
	assertContains(t, output, `return runtime.Equal((n % 2), 0), true`)
}

func TestRejectBlock(t *testing.T) {
	input := `def main
  odds = nums.reject do |n|
    n % 2 == 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Reject(nums, func(n any) (bool, bool) {`)
	assertContains(t, output, `return runtime.Equal((n % 2), 0), true`)
}

func TestReduceBlock(t *testing.T) {
	input := `def main
  sum = nums.reduce(0) do |acc, n|
    acc + n
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Reduce(nums, 0, func(acc any, n any) (any, bool) {`)
	assertContains(t, output, `return (acc + n), true`)
}

func TestFindBlock(t *testing.T) {
	input := `def main
  first_even = nums.find do |n|
    n % 2 == 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Find(nums, func(n any) (bool, bool) {`)
	assertContains(t, output, `return runtime.Equal((n % 2), 0), true`)
}

func TestAnyBlock(t *testing.T) {
	input := `def main
  has_even = nums.any? do |n|
    n % 2 == 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Any(nums, func(n any) (bool, bool) {`)
	assertContains(t, output, `return runtime.Equal((n % 2), 0), true`)
}

func TestAllBlock(t *testing.T) {
	input := `def main
  all_positive = nums.all? do |n|
    n > 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.All(nums, func(n any) (bool, bool) {`)
	assertContains(t, output, `return (n > 0), true`)
}

func TestNoneBlock(t *testing.T) {
	input := `def main
  no_negatives = nums.none? do |n|
    n < 0
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.None(nums, func(n any) (bool, bool) {`)
	assertContains(t, output, `return (n < 0), true`)
}

func TestKernelFunctions(t *testing.T) {
	input := `def main
  puts("hello")
  print("world")
  p(x)
  name = gets
  exit(1)
  sleep(2)
  n = rand(10)
  f = rand
end`

	output := compile(t, input)

	assertContains(t, output, `"github.com/nchapman/rugby/runtime"`)
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
    puts(i)
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Times(5, func(i int) bool {`)
	assertContains(t, output, `runtime.Puts(i)`)
	assertContains(t, output, `return true`)
}

func TestTimesBlockWithExpression(t *testing.T) {
	input := `def main
  n.times do |i|
    puts(i)
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Times(n, func(i int) bool {`)
	assertContains(t, output, `runtime.Puts(i)`)
	assertContains(t, output, `return true`)
}

func TestUptoBlock(t *testing.T) {
	input := `def main
  1.upto(5) do |i|
    puts(i)
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Upto(1, 5, func(i int) bool {`)
	assertContains(t, output, `runtime.Puts(i)`)
	assertContains(t, output, `return true`)
}

func TestUptoBlockWithVariables(t *testing.T) {
	input := `def main
  start.upto(finish) do |i|
    puts(i)
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Upto(start, finish, func(i int) bool {`)
	assertContains(t, output, `runtime.Puts(i)`)
	assertContains(t, output, `return true`)
}

func TestDowntoBlock(t *testing.T) {
	input := `def main
  5.downto(1) do |i|
    puts(i)
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Downto(5, 1, func(i int) bool {`)
	assertContains(t, output, `runtime.Puts(i)`)
	assertContains(t, output, `return true`)
}

func TestDowntoBlockWithVariables(t *testing.T) {
	input := `def main
  high.downto(low) do |i|
    puts(i)
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Downto(high, low, func(i int) bool {`)
	assertContains(t, output, `runtime.Puts(i)`)
	assertContains(t, output, `return true`)
}

func TestTimesBlockNoParam(t *testing.T) {
	input := `def main
  3.times do ||
    puts("hello")
  end
end`

	output := compile(t, input)

	// Should generate runtime.Times with _ parameter
	assertContains(t, output, `runtime.Times(3, func(_ int) bool {`)
	assertContains(t, output, `runtime.Puts("hello")`)
	assertContains(t, output, `return true`)
}

func TestUptoBlockNoParam(t *testing.T) {
	input := `def main
  1.upto(3) do ||
    puts("hello")
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Upto(1, 3, func(_ int) bool {`)
	assertContains(t, output, `runtime.Puts("hello")`)
	assertContains(t, output, `return true`)
}

func TestDowntoBlockNoParam(t *testing.T) {
	input := `def main
  3.downto(1) do ||
    puts("hello")
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Downto(3, 1, func(_ int) bool {`)
	assertContains(t, output, `runtime.Puts("hello")`)
	assertContains(t, output, `return true`)
}

func TestBraceBlockCodegen(t *testing.T) {
	input := `def main
  arr.each { |x| puts(x) }
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Each(arr, func(x any) bool {`)
	assertContains(t, output, `runtime.Puts(x)`)
	assertContains(t, output, `return true`)
}

func TestBraceBlockMapCodegen(t *testing.T) {
	input := `def main
  result = arr.map { |x| x * 2 }
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Map(arr, func(x any) (any, bool, bool) {`)
	assertContains(t, output, `return (x * 2), true, true`)
}

func TestBraceBlockTimesCodegen(t *testing.T) {
	input := `def main
  5.times { |i| puts(i) }
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Times(5, func(i int) bool {`)
	assertContains(t, output, `runtime.Puts(i)`)
	assertContains(t, output, `return true`)
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
    puts("hello")
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
  def add(a : any, b : any) -> Int
    return a + b
  end
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `type Calculator struct{}`)
	assertContains(t, output, `func (c *Calculator) add(a any, b any) int`)
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
    puts("inc")
  end

  def dec
    puts("dec")
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
    puts("running")
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
    puts("running")
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
  def initialize(name : any, age : any)
    @name = name
    @age = age
  end
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `type User struct {`)
	assertContains(t, output, `name any`)
	assertContains(t, output, `age any`)
	assertContains(t, output, `func newUser(name any, age any) *User`)
	assertContains(t, output, `u := &User{}`)
	assertContains(t, output, `u.name = name`)
	assertContains(t, output, `u.age = age`)
	assertContains(t, output, `return u`)
}

func TestClassMethodAccessingInstanceVar(t *testing.T) {
	input := `class User
  def initialize(name : any)
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
  def initialize(name : any)
    @name = name
  end
end

def main
  user = User.new("Alice")
end`

	output := compile(t, input)

	assertContains(t, output, `func newUser(name any) *User`)
	assertContains(t, output, `user := newUser("Alice")`)
}

func TestClassNewWithMultipleArgs(t *testing.T) {
	input := `class Point
  def initialize(x : any, y : any)
    @x = x
    @y = y
  end
end

def main
  p = Point.new(10, 20)
end`

	output := compile(t, input)

	assertContains(t, output, `func newPoint(x any, y any) *Point`)
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
	input := `def foo(a : Int, b : any, c : String)
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func foo(a int, b any, c string)`)
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
	input := `def add(a : any, b : any)
  return a
end

def main
  x = 5
end`

	output := compile(t, input)

	assertContains(t, output, `func add(a any, b any)`)
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
    puts(item)
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
    puts(item)
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
    puts("incremented")
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
  def valid? -> Bool
    return true
  end
end

def main
end`

	output := compile(t, input)

	// Method name should have ? stripped and be camelCased (lowercase first letter)
	assertContains(t, output, `func (u *User) valid() bool`)
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
    @name = ""
  end

  def with_name(n : any)
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
	assertContains(t, output, `func (b *Builder) withName(n any)`)
	// Implicit return of self
	assertContains(t, output, "return b\n}")
	// Explicit return of self
	assertContains(t, output, "return b")
}

func TestSelfInMethodChain(t *testing.T) {
	input := `class Config
  def initialize
    @value = nil
  end

  def set_value(v : any)
    @value = v
    self
  end
end

def main
  c = Config.new()
  c.setValue(1).setValue(2)
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

  def ==(other : any)
    @x == other.x and @y == other.y
  end
end

def main
end`

	output := compile(t, input)

	// def == should compile to Equal(other any) bool
	assertContains(t, output, `func (p *Point) Equal(other any) bool {`)
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
	assertContains(t, output, `"github.com/nchapman/rugby/runtime"`)
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
    puts(i)
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
    puts(i)
  end
end`

	output := compile(t, input)

	// Should use range over int (Go 1.22) for 0...n
	assertContains(t, output, `for i := range 5 {`)
}

func TestForLoopWithRangeVariables(t *testing.T) {
	input := `def main
  start = 1
  finish = 10
  for i in start..finish
    puts(i)
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
    puts(i)
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
    puts(i)
  end
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.RangeEach(runtime.Range{Start: 1, End: 5, Exclusive: false}, func(i int) bool {`)
	assertContains(t, output, `runtime.Puts(i)`)
	assertContains(t, output, `return true`)
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

func TestRangeRejectsFloatStart(t *testing.T) {
	// Range start must be Int
	input := `def main
  r = 1.5..10
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	gen := New()
	_, err := gen.Generate(program)

	if err == nil {
		t.Fatal("expected error for range with Float start")
	}

	if !strings.Contains(err.Error(), "range start must be Int") {
		t.Errorf("expected error about Int start, got: %s", err.Error())
	}
}

func TestRangeRejectsStringEnd(t *testing.T) {
	// Range end must be Int
	input := `def main
  r = 1.."z"
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	gen := New()
	_, err := gen.Generate(program)

	if err == nil {
		t.Fatal("expected error for range with String end")
	}

	if !strings.Contains(err.Error(), "range end must be Int") {
		t.Errorf("expected error about Int end, got: %s", err.Error())
	}
}

func TestRangeWithIntVariables(t *testing.T) {
	// Ranges with Int variables should work fine
	input := `def main
  a : Int = 1
  b : Int = 10
  r = a..b
end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Range{Start: a, End: b, Exclusive: false}`)
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
  def initialize
    @cache = nil
  end

  def get_cache
    @cache ||= load_cache()
  end
end

def main
end`

	output := compile(t, input)

	// Should generate nil check for instance var
	assertContains(t, output, "if s.cache == nil {")
	assertContains(t, output, "s.cache = loadCache()")
}

func TestCompoundAssignPlusEquals(t *testing.T) {
	input := `def main
  x = 10
  x += 5
end`

	output := compile(t, input)

	// Should generate x = x + 5
	assertContains(t, output, "x = x + 5")
}

func TestCompoundAssignMinusEquals(t *testing.T) {
	input := `def main
  x = 10
  x -= 3
end`

	output := compile(t, input)

	// Should generate x = x - 3
	assertContains(t, output, "x = x - 3")
}

func TestCompoundAssignStarEquals(t *testing.T) {
	input := `def main
  x = 10
  x *= 2
end`

	output := compile(t, input)

	// Should generate x = x * 2
	assertContains(t, output, "x = x * 2")
}

func TestCompoundAssignSlashEquals(t *testing.T) {
	input := `def main
  x = 10
  x /= 5
end`

	output := compile(t, input)

	// Should generate x = x / 5
	assertContains(t, output, "x = x / 5")
}

func TestOptionalValueType(t *testing.T) {
	input := `def find(id : Int?) -> String?
end

def main
end`

	output := compile(t, input)

	// Value type optionals use *T (pointers)
	assertContains(t, output, "func find(id *int) *string")
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
	assertContains(t, output, "var x *int = runtime.NoneInt()")
}

func TestOptionalValueTypeInCondition(t *testing.T) {
	// Rugby requires explicit nil checks for optionals (no implicit truthiness)
	input := `def main
  x : Int? = nil
  if x != nil
    puts("has value")
  end
end`

	output := compile(t, input)

	// Explicit nil check compiles using runtime.Equal
	assertContains(t, output, "if !runtime.Equal(x, nil)")
}

func TestOptionalReferenceTypeInCondition(t *testing.T) {
	// Rugby requires explicit nil checks for optionals
	input := `def process(u : User?)
  if u != nil
    puts("has user")
  end
end

def main
end`

	output := compile(t, input)

	// Explicit nil check compiles using runtime.Equal
	assertContains(t, output, "if !runtime.Equal(u, nil)")
}

func TestOptionalInElsifCondition(t *testing.T) {
	// Rugby requires explicit nil checks for optionals
	input := `def main
  x : Int? = nil
  y : String? = nil
  if x != nil
    puts("x")
  elsif y != nil
    puts("y")
  end
end`

	output := compile(t, input)

	// Both if and elsif use explicit nil checks with runtime.Equal
	assertContains(t, output, "if !runtime.Equal(x, nil)")
	assertContains(t, output, "} else if !runtime.Equal(y, nil)")
}

func TestNonOptionalInCondition(t *testing.T) {
	input := `def main
  x : Bool = true
  if x
    puts("yes")
  end
end`

	output := compile(t, input)

	// Non-optional types should be used as-is
	assertContains(t, output, "if x {")
}

func TestStrictConditionRejectsNonBool(t *testing.T) {
	// Using non-Bool type directly in condition should error
	input := `def main
  x : Int = 5
  if x
    puts("yes")
  end
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	gen := New()
	_, err := gen.Generate(program)

	if err == nil {
		t.Fatal("expected error for non-Bool condition, got none")
	}

	if !strings.Contains(err.Error(), "condition must be Bool, got Int") {
		t.Errorf("expected error about Bool type, got: %s", err.Error())
	}
}

func TestStrictConditionRejectsOptionalWithoutExplicitCheck(t *testing.T) {
	// Using optional type directly in condition should error (suggest if let)
	input := `def main
  x : Int? = nil
  if x
    puts("yes")
  end
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	gen := New()
	_, err := gen.Generate(program)

	if err == nil {
		t.Fatal("expected error for optional in condition without explicit check")
	}

	if !strings.Contains(err.Error(), "use 'if let x = ...' or 'x != nil' for optionals") {
		t.Errorf("expected helpful error message about optionals, got: %s", err.Error())
	}
}

func TestNilLiteral(t *testing.T) {
	input := `def main
  x : User? = nil
end`

	output := compile(t, input)

	// nil literal should be generated as-is
	assertContains(t, output, "var x *User = nil")
}

// --- Strict ? Suffix Tests ---

func TestPredicateMethodReturnsBool(t *testing.T) {
	// Methods ending in ? must return Bool - this should compile
	input := `def valid? -> Bool
  true
end`

	output := compile(t, input)

	assertContains(t, output, "func valid() bool")
	assertContains(t, output, "return true")
}

func TestPredicateMethodRejectsNonBoolReturn(t *testing.T) {
	// Methods ending in ? that don't return Bool should error
	input := `def valid? -> Int
  42
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	gen := New()
	_, err := gen.Generate(program)

	if err == nil {
		t.Fatal("expected error for predicate method not returning Bool")
	}

	if !strings.Contains(err.Error(), "must return Bool") {
		t.Errorf("expected error about Bool return type, got: %s", err.Error())
	}
}

func TestPredicateMethodRejectsNoReturnType(t *testing.T) {
	// Methods ending in ? with no explicit return type should error
	input := `def valid?
  true
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	gen := New()
	_, err := gen.Generate(program)

	if err == nil {
		t.Fatal("expected error for predicate method without explicit Bool return")
	}

	if !strings.Contains(err.Error(), "must return Bool") {
		t.Errorf("expected error about Bool return type, got: %s", err.Error())
	}
}

func TestClassPredicateMethodReturnsBool(t *testing.T) {
	// Class methods ending in ? must return Bool - this should compile
	input := `class Container
  def empty? -> Bool
    true
  end
end`

	output := compile(t, input)

	assertContains(t, output, "func (c *Container) empty() bool")
}

func TestClassPredicateMethodRejectsNonBoolReturn(t *testing.T) {
	// Class methods ending in ? that don't return Bool should error
	input := `class Container
  def count? -> Int
    0
  end
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	gen := New()
	_, err := gen.Generate(program)

	if err == nil {
		t.Fatal("expected error for class predicate method not returning Bool")
	}

	if !strings.Contains(err.Error(), "must return Bool") {
		t.Errorf("expected error about Bool return type, got: %s", err.Error())
	}
}

func TestStringToIntWithBang(t *testing.T) {
	input := `def process(s : String) -> (Int, error)
  n = s.to_i()!
  n
end`

	output := compile(t, input)

	// to_i()! should propagate error
	assertContains(t, output, "runtime.StringToInt(s)")
	assertContains(t, output, "if _err != nil {")
	assertContains(t, output, "return 0, _err")
}

func TestStringToFloatWithBang(t *testing.T) {
	input := `def process(s : String) -> (Float, error)
  n = s.to_f()!
  n
end`

	output := compile(t, input)

	// to_f()! should propagate error
	assertContains(t, output, "runtime.StringToFloat(s)")
	assertContains(t, output, "if _err != nil {")
	assertContains(t, output, "return 0, _err")
}

func TestStringToIntWithRescue(t *testing.T) {
	input := `def main
  s = "abc"
  n = s.to_i() rescue 0
  puts(n)
end`

	output := compile(t, input)

	// to_i() rescue should provide default on error
	assertContains(t, output, "runtime.StringToInt(s)")
	assertContains(t, output, "if _err != nil {")
}

func TestStringToIntExplicit(t *testing.T) {
	input := `def main
  s = "42"
  n, err = s.to_i()
  if err != nil
    puts("error")
  else
    puts(n)
  end
end`

	output := compile(t, input)

	// explicit error handling
	assertContains(t, output, "n, err := runtime.StringToInt(s)")
	// The condition uses runtime.Equal for comparison
	assertContains(t, output, "!runtime.Equal(err, nil)")
}

// --- Error Utility Tests ---

func TestErrorIs(t *testing.T) {
	input := `import "io"

def main
  err : error = nil
  if error_is?(err, io.EOF)
    puts("end of file")
  end
end`

	output := compile(t, input)

	// error_is? should compile to errors.Is
	assertContains(t, output, "errors.Is(err, io.EOF)")
	// Should auto-import errors package
	assertContains(t, output, `"errors"`)
}

func TestErrorIsWithCustomError(t *testing.T) {
	input := `def check(err : error, target : error) -> Bool
  error_is?(err, target)
end`

	output := compile(t, input)

	assertContains(t, output, "errors.Is(err, target)")
}

func TestErrorAs(t *testing.T) {
	input := `import "os"

def main
  err : error = nil
  if let pathErr = error_as(err, os.PathError)
    puts(pathErr.Path)
  end
end`

	output := compile(t, input)

	// error_as should compile to errors.As pattern returning optional
	assertContains(t, output, "errors.As(err, &_target)")
	assertContains(t, output, "*os.PathError")
	// Should auto-import errors package
	assertContains(t, output, `"errors"`)
}

func TestErrorAsWithIfLet(t *testing.T) {
	input := `class MyError
  def initialize(@code : Int)
  end
end

def check(err : error) -> Int
  if let myErr = error_as(err, MyError)
    myErr.code
  else
    -1
  end
end`

	output := compile(t, input)

	// Check key parts of the generated errors.As pattern
	assertContains(t, output, "var _target *MyError")
	assertContains(t, output, "errors.As(err, &_target)")
	assertContains(t, output, "return _target")
}

// --- Bare Script Tests ---

func TestBareScriptSimple(t *testing.T) {
	input := `puts("hello")`

	output := compile(t, input)

	// Should generate implicit main
	assertContains(t, output, "func main()")
	assertContains(t, output, `runtime.Puts("hello")`)
}

func TestBareScriptWithFunction(t *testing.T) {
	input := `def greet(name : String)
  puts(name)
end

greet("World")`

	output := compile(t, input)

	// Should have the function definition
	assertContains(t, output, "func greet(name string)")
	// Should have implicit main with the call
	assertContains(t, output, "func main()")
	assertContains(t, output, `greet("World")`)
}

func TestBareScriptWithClass(t *testing.T) {
	input := `class Counter
  def initialize(n : Int)
    @n = n
  end

  def value -> Int
    @n
  end
end

c = Counter.new(5)
puts(c.value)`

	output := compile(t, input)

	// Should have the class definition
	assertContains(t, output, "type Counter struct")
	assertContains(t, output, "func newCounter(n int)")
	// Should have implicit main
	assertContains(t, output, "func main()")
	assertContains(t, output, "c := newCounter(5)")
}

func TestBareScriptMultipleStatements(t *testing.T) {
	input := `x = 1
y = 2
z = x + y
puts(z)`

	output := compile(t, input)

	// Should have implicit main with all statements
	assertContains(t, output, "func main()")
	assertContains(t, output, "x := 1")
	assertContains(t, output, "y := 2")
	assertContains(t, output, "z := (x + y)")
}

func TestBareScriptWithControlFlow(t *testing.T) {
	input := `x = 5
if x > 3
  puts("big")
end`

	output := compile(t, input)

	// Should have implicit main with control flow
	assertContains(t, output, "func main()")
	assertContains(t, output, "x := 5")
	assertContains(t, output, "if x > 3")
}

func TestBareScriptConflictError(t *testing.T) {
	input := `def main
  puts("in main")
end

puts("top-level")`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Skipf("parser error: %v", p.Errors())
	}

	gen := New()
	_, err := gen.Generate(program)

	if err == nil {
		t.Error("expected error for conflicting main and top-level statements, got none")
	}
	if err != nil && !strings.Contains(err.Error(), "cannot mix") {
		t.Errorf("expected 'cannot mix' error, got: %v", err)
	}
}

func TestExplicitMainStillWorks(t *testing.T) {
	input := `def main
  puts("hello")
end`

	output := compile(t, input)

	// Should have the explicit main
	assertContains(t, output, "func main()")
	assertContains(t, output, `runtime.Puts("hello")`)
}

func TestLibraryModeNoMain(t *testing.T) {
	input := `def helper(x : Int) -> Int
  x * 2
end

class Util
  def initialize
  end
end`

	output := compile(t, input)

	// Should NOT have a main function (library mode)
	if strings.Contains(output, "func main()") {
		t.Error("library mode should not generate main function")
	}

	// Should have the function and class
	assertContains(t, output, "func helper(x int) int")
	assertContains(t, output, "type Util struct")
}

func TestBareScriptWithImport(t *testing.T) {
	input := `import fmt

fmt.Println("hello")`

	output := compile(t, input)

	// Should have import and implicit main
	assertContains(t, output, `"fmt"`)
	assertContains(t, output, "func main()")
	assertContains(t, output, `fmt.Println("hello")`)
}

func TestBareScriptWithForLoop(t *testing.T) {
	input := `for i in 1..3
  puts(i)
end`

	output := compile(t, input)

	// Should have implicit main with for loop
	assertContains(t, output, "func main()")
	assertContains(t, output, "for i := 1; i <= 3; i++")
}

func TestBareScriptWithWhileLoop(t *testing.T) {
	input := `x = 0
while x < 3
  x = x + 1
end`

	output := compile(t, input)

	// Should have implicit main with while loop
	assertContains(t, output, "func main()")
	assertContains(t, output, "x := 0")
	assertContains(t, output, "for x < 3")
}

func TestBareScriptWithBlock(t *testing.T) {
	input := `[1, 2, 3].each do |x|
  puts(x)
end`

	output := compile(t, input)

	// Should have implicit main with each block
	assertContains(t, output, "func main()")
	assertContains(t, output, "runtime.Each")
}

func TestBareScriptMixedOrdering(t *testing.T) {
	input := `puts("start")

def helper(x : Int) -> Int
  return x * 2
end

result = helper(5)
puts(result)`

	output := compile(t, input)

	// Function should be defined at package level, not inside main
	assertContains(t, output, "func helper(x int) int")
	// Statements should be in implicit main
	assertContains(t, output, "func main()")
	assertContains(t, output, `runtime.Puts("start")`)
	assertContains(t, output, "result := helper(5)")

	// Verify function is NOT inside main (check ordering)
	mainIdx := strings.Index(output, "func main()")
	helperIdx := strings.Index(output, "func helper(")
	if helperIdx > mainIdx {
		t.Error("helper function should be defined before main, not inside it")
	}
}

func TestBareScriptEmpty(t *testing.T) {
	input := ``

	output := compile(t, input)

	// Should generate minimal valid Go package with no main
	assertContains(t, output, "package main")
	if strings.Contains(output, "func main()") {
		t.Error("empty program should not generate main function")
	}
}

func TestBareScriptOnlyComments(t *testing.T) {
	input := `# This is a comment
# Another comment`

	output := compile(t, input)

	// Should generate minimal valid Go package with no main
	assertContains(t, output, "package main")
	if strings.Contains(output, "func main()") {
		t.Error("comment-only program should not generate main function")
	}
}

func TestBareScriptVariableScoping(t *testing.T) {
	// Test that variables in functions don't leak into implicit main
	input := `def helper
  x = 10
end

y = 20
puts(y)`

	output := compile(t, input)

	// Both should use := since they're new declarations in their respective scopes
	assertContains(t, output, "x := 10")
	assertContains(t, output, "y := 20")
}

func TestSymbolLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "basic symbol assignment",
			input: `def main
  status = :ok
end`,
			expected: []string{
				`status := "ok"`,
			},
		},
		{
			name: "symbol with underscores",
			input: `def main
  state = :not_found
end`,
			expected: []string{
				`state := "not_found"`,
			},
		},
		{
			name: "symbols in array",
			input: `def main
  statuses = [:pending, :active, :completed]
end`,
			expected: []string{
				`statuses := []string{"pending", "active", "completed"}`,
			},
		},
		{
			name: "symbol as function argument",
			input: `def main
  puts(:hello)
end`,
			expected: []string{
				`runtime.Puts("hello")`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := compile(t, tt.input)
			for _, exp := range tt.expected {
				assertContains(t, output, exp)
			}
		})
	}
}

func TestGenerateCaseStatement(t *testing.T) {
	input := `def main
  x = 2
  case x
  when 1
    puts("one")
  when 2, 3
    puts("two or three")
  else
    puts("other")
  end
end`

	output := compile(t, input)

	assertContains(t, output, `switch x {`)
	assertContains(t, output, `case 1:`)
	assertContains(t, output, `case 2, 3:`)
	assertContains(t, output, `default:`)
	assertContains(t, output, `runtime.Puts("one")`)
	assertContains(t, output, `runtime.Puts("two or three")`)
	assertContains(t, output, `runtime.Puts("other")`)
}

func TestGenerateCaseStatementNoElse(t *testing.T) {
	input := `def main
  status = 200
  case status
  when 200
    puts("ok")
  when 404
    puts("not found")
  end
end`

	output := compile(t, input)

	assertContains(t, output, `switch status {`)
	assertContains(t, output, `case 200:`)
	assertContains(t, output, `case 404:`)
	assertNotContains(t, output, `default:`)
}

func TestGenerateCaseStatementNoSubject(t *testing.T) {
	input := `def main
  x = 15
  case
  when x > 10
    puts("big")
  when x > 5
    puts("medium")
  else
    puts("small")
  end
end`

	output := compile(t, input)

	assertContains(t, output, `switch {`)
	assertContains(t, output, `case (x > 10):`)
	assertContains(t, output, `case (x > 5):`)
	assertContains(t, output, `default:`)
}

func TestGenerateCaseWithStringsAndSymbols(t *testing.T) {
	input := `def main
  status = "active"
  case status
  when "active"
    puts("running")
  when :stopped
    puts("halted")
  else
    puts("unknown")
  end
end`

	output := compile(t, input)

	assertContains(t, output, `switch status {`)
	assertContains(t, output, `case "active":`)
	assertContains(t, output, `case "stopped":`) // symbol compiles to string
	assertContains(t, output, `runtime.Puts("running")`)
	assertContains(t, output, `runtime.Puts("halted")`)
}

// --- case_type (Type Switch) Tests ---

func TestGenerateCaseTypeBasic(t *testing.T) {
	input := `def type_of(x : any) -> String
  case_type x
  when String
    return "it's a string"
  when Int
    return "it's an int"
  else
    return "unknown"
  end
end`

	output := compile(t, input)

	// Should generate Go type switch with shadowing
	assertContains(t, output, `switch x := x.(type) {`)
	assertContains(t, output, `case string:`)
	assertContains(t, output, `case int:`)
	assertContains(t, output, `default:`)
	assertContains(t, output, `return "it's a string"`)
}

func TestGenerateCaseTypeNoElse(t *testing.T) {
	input := `def process(x : any)
  case_type x
  when String
    puts("string")
  when Int
    puts("int")
  end
end`

	output := compile(t, input)

	assertContains(t, output, `switch x := x.(type) {`)
	assertContains(t, output, `case string:`)
	assertContains(t, output, `case int:`)
	assertNotContains(t, output, `default:`)
}

func TestGenerateCaseTypeWithFloat(t *testing.T) {
	input := `def check(x : any)
  case_type x
  when Float
    puts("float")
  when String
    puts("string")
  end
end`

	output := compile(t, input)

	assertContains(t, output, `switch x := x.(type) {`)
	assertContains(t, output, `case float64:`)
	assertContains(t, output, `case string:`)
}

func TestGenerateCaseTypeWithBool(t *testing.T) {
	input := `def check(x : any)
  case_type x
  when Bool
    puts("boolean")
  when Int
    puts("integer")
  end
end`

	output := compile(t, input)

	assertContains(t, output, `switch x := x.(type) {`)
	assertContains(t, output, `case bool:`)
	assertContains(t, output, `case int:`)
}

// Error handling tests (Phase 19)

func TestGenerateErrorReturnType(t *testing.T) {
	input := `def save(path : String) -> error
  nil
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func save(path string) error`)
	assertContains(t, output, `return nil`)
}

func TestGenerateValueAndErrorReturnType(t *testing.T) {
	input := `def read_file(path : String) -> (String, error)
  return "content", nil
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func readFile(path string) (string, error)`)
	assertContains(t, output, `return "content", nil`)
}

func TestGenerateMultipleValuesAndErrorReturnType(t *testing.T) {
	input := `def parse(s : String) -> (Int, Bool, error)
  return 42, true, nil
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func parse(s string) (int, bool, error)`)
	assertContains(t, output, `return 42, true, nil`)
}

func TestGeneratePanicStatement(t *testing.T) {
	input := `def main
  panic "something went wrong"
end`

	output := compile(t, input)

	assertContains(t, output, `panic("something went wrong")`)
}

func TestGeneratePanicWithInterpolation(t *testing.T) {
	input := `def main
  state = "invalid"
  panic "bad state: #{state}"
end`

	output := compile(t, input)

	assertContains(t, output, `panic(fmt.Sprintf("bad state: %v", state))`)
}

func TestGeneratePanicWithVariable(t *testing.T) {
	input := `def main
  msg = "error occurred"
  panic msg
end`

	output := compile(t, input)

	assertContains(t, output, `panic(msg)`)
}

func TestGenerateBangInMain(t *testing.T) {
	input := `def read_file(path : String) -> (String, error)
  return "content", nil
end

def main
  data = read_file("test.txt")!
  puts(data)
end`

	output := compile(t, input)

	// Should use runtime.Fatal in main
	assertContains(t, output, `data, _err := readFile("test.txt")`)
	assertContains(t, output, `if _err != nil {`)
	assertContains(t, output, `runtime.Fatal(_err)`)
}

func TestGenerateBangInErrorFunction(t *testing.T) {
	input := `def read_file(path : String) -> (String, error)
  return "content", nil
end

def load(path : String) -> (String, error)
  data = read_file(path)!
  return data, nil
end

def main
end`

	output := compile(t, input)

	// Should propagate error in error-returning function
	assertContains(t, output, `data, _err := readFile(path)`)
	assertContains(t, output, `if _err != nil {`)
	assertContains(t, output, `return "", _err`)
}

func TestGenerateBangAsStatement(t *testing.T) {
	input := `def do_something() -> error
  return nil
end

def main
  do_something()!
end`

	output := compile(t, input)

	// Should generate inline if for statement bang
	assertContains(t, output, `if _err := doSomething(); _err != nil {`)
	assertContains(t, output, `runtime.Fatal(_err)`)
}

func TestGenerateBangWithMultiReturnError(t *testing.T) {
	// Test error propagation in a function that returns (T, T, error)
	input := `def fetch(url : String) -> (String, error)
  return "data", nil
end

def process(url : String) -> (String, Int, error)
  data = fetch(url)!
  return data, 200, nil
end

def main
end`

	output := compile(t, input)

	// Should generate zero values for all non-error returns
	assertContains(t, output, `data, _err := fetch(url)`)
	assertContains(t, output, `return "", 0, _err`)
}

func TestGenerateBangWithInterfaceReturn(t *testing.T) {
	input := `def parse(s : String) -> (any, error)
  return nil, nil
end

def process(s : String) -> (any, error)
  obj = parse(s)!
  return obj, nil
end

def main
end`

	output := compile(t, input)

	// Should use nil as zero value for 'any' (interface)
	assertContains(t, output, `obj, _err := parse(s)`)
	assertContains(t, output, `return nil, _err`)
}

func TestGenerateRescueInline(t *testing.T) {
	input := `def read_config(path : String) -> (String, error)
  return "config", nil
end

def main
  config = read_config("app.yml") rescue "default"
  puts(config)
end`

	output := compile(t, input)

	// Should generate inline rescue with default value
	assertContains(t, output, `config, _err := readConfig("app.yml")`)
	assertContains(t, output, `if _err != nil {`)
	assertContains(t, output, `config = "default"`)
}

func TestGenerateRescueBlock(t *testing.T) {
	input := `def load_data(path : String) -> (String, error)
  return "data", nil
end

def main
  data = load_data("file.txt") rescue do
    puts("load failed")
    "fallback"
  end
end`

	output := compile(t, input)

	// Should generate block rescue
	assertContains(t, output, `data, _err := loadData("file.txt")`)
	assertContains(t, output, `if _err != nil {`)
	assertContains(t, output, `runtime.Puts("load failed")`)
	assertContains(t, output, `data = "fallback"`)
}

func TestGenerateRescueWithErrorBinding(t *testing.T) {
	input := `def fetch_url(url : String) -> (String, error)
  return "response", nil
end

def main
  result = fetch_url("http://example.com") rescue => err do
    puts("error: #{err}")
    "error"
  end
end`

	output := compile(t, input)

	// Should generate error binding
	assertContains(t, output, `result, _err := fetchURL("http://example.com")`)
	assertContains(t, output, `if _err != nil {`)
	assertContains(t, output, `err := _err`)
	assertContains(t, output, `runtime.Puts(fmt.Sprintf("error: %v", err))`)
	assertContains(t, output, `result = "error"`)
}

func TestGenerateErrorKernelFunc(t *testing.T) {
	input := `def validate(n : Int) -> (Int, error)
  if n < 0
    return 0, error("negative number")
  end
  return n, nil
end

def main
end`

	output := compile(t, input)

	// Should use runtime.Error
	assertContains(t, output, `runtime.Error("negative number")`)
}

func TestGenerateErrorWithInterpolation(t *testing.T) {
	input := `def validate(n : Int) -> (Int, error)
  if n < 0
    return 0, error("negative: #{n}")
  end
  return n, nil
end

def main
end`

	output := compile(t, input)

	// Should use runtime.Error with fmt.Sprintf
	assertContains(t, output, `runtime.Error(fmt.Sprintf("negative: %v", n))`)
}

func TestNilCoalesceIntOptional(t *testing.T) {
	input := `def get_value(opt : Int?) -> Int
  return opt ?? 0
end

def main
end`

	output := compile(t, input)

	// Should use runtime.CoalesceInt for Int?
	assertContains(t, output, `runtime.CoalesceInt(opt, 0)`)
}

func TestNilCoalesceStringOptional(t *testing.T) {
	input := `def get_name(opt : String?) -> String
  return opt ?? "default"
end

def main
end`

	output := compile(t, input)

	// Should use runtime.CoalesceString for String?
	assertContains(t, output, `runtime.CoalesceString(opt, "default")`)
}

func TestNilCoalesceFloatOptional(t *testing.T) {
	input := `def get_price(opt : Float?) -> Float
  return opt ?? 3.14
end

def main
end`

	output := compile(t, input)

	// Should use runtime.CoalesceFloat for Float?
	assertContains(t, output, `runtime.CoalesceFloat(opt, 3.14)`)
}

func TestNilCoalesceBoolOptional(t *testing.T) {
	input := `def get_flag(opt : Bool?) -> Bool
  return opt ?? false
end

def main
end`

	output := compile(t, input)

	// Should use runtime.CoalesceBool for Bool?
	assertContains(t, output, `runtime.CoalesceBool(opt, false)`)
}

func TestSafeNavigation(t *testing.T) {
	input := `def get_length(opt : String?) -> any
  return opt&.length
end

def main
end`

	output := compile(t, input)

	// Should generate safe navigation check with captured variable to avoid double evaluation
	// Uses unique variable name (_sn0, _sn1, etc.) to handle nested expressions
	assertContains(t, output, `_sn0 := opt`)
	assertContains(t, output, `if _sn0 != nil`)
	assertContains(t, output, `(*_sn0).length`)
	assertContains(t, output, `return nil`)
}

func TestChainedSafeNavigationUniqueVars(t *testing.T) {
	input := `def get_city(user : User?) -> any
  return user&.address&.city
end

def main
end`

	output := compile(t, input)

	// Should use unique variable names for each level of chained safe navigation
	// Inner expressions are generated first, so user&.address uses _sn1, and the outer &.city uses _sn0
	assertContains(t, output, `_sn0 :=`)
	assertContains(t, output, `_sn1 := user`)
}

func TestMultiAssignmentNewVariables(t *testing.T) {
	input := `def main
  val, ok = get_data()
end

def get_data() -> (Int, Bool)
  return 42, true
end`

	output := compile(t, input)

	// Should use := for new variables
	assertContains(t, output, `val, ok := getData()`)
}

func TestMultiAssignmentExistingVariables(t *testing.T) {
	input := `def main
  val = 0
  ok = false
  val, ok = get_data()
end

def get_data() -> (Int, Bool)
  return 42, true
end`

	output := compile(t, input)

	// Should use = for existing variables
	assertContains(t, output, `val, ok = getData()`)
}

func TestMultiAssignmentThreeValues(t *testing.T) {
	input := `def main
  a, b, c = get_triple()
end

def get_triple() -> (Int, Int, Int)
  return 1, 2, 3
end`

	output := compile(t, input)

	// Should handle three values
	assertContains(t, output, `a, b, c := getTriple()`)
}

func TestOptionalMethodOk(t *testing.T) {
	input := `def check(opt : Int?) -> Bool
  return opt.ok?
end

def main
end`

	output := compile(t, input)

	// Should generate nil check
	assertContains(t, output, `(opt != nil)`)
}

func TestOptionalMethodPresent(t *testing.T) {
	input := `def check(opt : String?) -> Bool
  return opt.present?
end

def main
end`

	output := compile(t, input)

	// Should generate nil check
	assertContains(t, output, `(opt != nil)`)
}

func TestOptionalMethodNil(t *testing.T) {
	input := `def check(opt : Int?) -> Bool
  return opt.nil?
end

def main
end`

	output := compile(t, input)

	// Should generate nil check
	assertContains(t, output, `(opt == nil)`)
}

func TestOptionalMethodAbsent(t *testing.T) {
	input := `def check(opt : Float?) -> Bool
  return opt.absent?
end

def main
end`

	output := compile(t, input)

	// Should generate nil check
	assertContains(t, output, `(opt == nil)`)
}

func TestOptionalMethodUnwrap(t *testing.T) {
	input := `def get_value(opt : Int?) -> Int
  return opt.unwrap
end

def main
end`

	output := compile(t, input)

	// Should generate dereference
	assertContains(t, output, `return *opt`)
}

func TestAccessorGetter(t *testing.T) {
	input := `pub class User
  getter name : String
  
  def initialize(@name : String)
  end
end

def main
end`

	output := compile(t, input)

	// Should have the field in the struct
	assertContains(t, output, "name string")
	// Should have a getter method
	assertContains(t, output, "func (u *User) Name() string {")
	assertContains(t, output, "return u.name")
}

func TestAccessorSetter(t *testing.T) {
	input := `pub class User
  setter email : String
  
  def initialize(@email : String)
  end
end

def main
end`

	output := compile(t, input)

	// Should have the field in the struct
	assertContains(t, output, "email string")
	// Should have a setter method
	assertContains(t, output, "func (u *User) SetEmail(v string) {")
	assertContains(t, output, "u.email = v")
}

func TestAccessorProperty(t *testing.T) {
	input := `pub class Counter
  property value : Int
  
  def initialize(@value : Int)
  end
end

def main
end`

	output := compile(t, input)

	// Should have the field in the struct
	assertContains(t, output, "value int")
	// Should have both getter and setter methods
	assertContains(t, output, "func (c *Counter) Value() int {")
	assertContains(t, output, "return c.value")
	assertContains(t, output, "func (c *Counter) SetValue(v int) {")
	assertContains(t, output, "c.value = v")
}

func TestAccessorNonPubClass(t *testing.T) {
	input := `class User
  property name : String

  def initialize(@name : String)
  end
end

def main
end`

	output := compile(t, input)

	// Non-pub class should have camelCase method names
	assertContains(t, output, "func (u *User) name() string {")
	assertContains(t, output, "func (u *User) setName(v string) {")
}

func TestSuperKeyword(t *testing.T) {
	input := `class Parent
  def greet
    puts "Hello from Parent"
  end
end

class Child < Parent
  def greet
    super
    puts "Hello from Child"
  end
end

def main
end`

	output := compile(t, input)

	// Child class should embed Parent
	assertContains(t, output, "type Child struct {")
	assertContains(t, output, "Parent")

	// super should call parent's method on the embedded field
	assertContains(t, output, "c.Parent.greet()")
}

func TestSuperWithArgs(t *testing.T) {
	input := `class Parent
  def add(a : Int, b : Int) -> Int
    a + b
  end
end

class Child < Parent
  def add(a : Int, b : Int) -> Int
    super(a, b) + 10
  end
end

def main
end`

	output := compile(t, input)

	// super with args should pass them through
	assertContains(t, output, "c.Parent.add(a, b)")
}

func TestSuperInPubClass(t *testing.T) {
	input := `pub class Base
  pub def process
    puts "Base process"
  end
end

pub class Derived < Base
  pub def process
    super
  end
end

def main
end`

	output := compile(t, input)

	// Pub method names should be PascalCase
	assertContains(t, output, "d.Base.Process()")
}

func TestSuperInNonPubMethod(t *testing.T) {
	input := `class Parent
  def greet
    puts "Hello"
  end
end

class Child < Parent
  def greet
    super
  end
end

def main
end`

	output := compile(t, input)

	// Non-pub method names should be camelCase
	assertContains(t, output, "c.Parent.greet()")
}

func TestModuleWithMethod(t *testing.T) {
	input := `module Callable
  def call
    puts "calling"
  end
end

class Worker
  include Callable
end

def main
end`

	output := compile(t, input)

	// Worker should have the call method from Callable
	assertContains(t, output, "type Worker struct{}")
	assertContains(t, output, "func (w *Worker) call() {")
}

func TestModuleWithProperty(t *testing.T) {
	input := `module Named
  property name : String
end

class User
  include Named
end

def main
end`

	output := compile(t, input)

	// User should have the name field from Named module
	assertContains(t, output, "type User struct {")
	assertContains(t, output, "name string")
	// User should have getter/setter from Named module
	assertContains(t, output, "func (u *User) name() string {")
	assertContains(t, output, "func (u *User) setName(v string) {")
}

func TestModuleMultipleIncludes(t *testing.T) {
	input := `module Callable
  def call
    puts "calling"
  end
end

module Named
  property name : String
end

class Worker
  include Callable
  include Named
end

def main
end`

	output := compile(t, input)

	// Worker should have both module's members
	assertContains(t, output, "name string")
	assertContains(t, output, "func (w *Worker) call() {")
	assertContains(t, output, "func (w *Worker) name() string {")
}

func TestModuleWithClassFields(t *testing.T) {
	input := `module Loggable
  def log(msg : String)
    puts msg
  end
end

class Service
  include Loggable
  @port : Int

  def initialize(@port : Int)
  end
end

def main
end`

	output := compile(t, input)

	// Service should have both its own port field and module's log method
	assertContains(t, output, "port int")
	assertContains(t, output, "func (s *Service) log(msg string) {")
}

func TestSymbolKeyShorthandCodegen(t *testing.T) {
	input := `def main
  m = {name: "Alice", age: 30}
end`

	output := compile(t, input)

	// Keys should be strings
	assertContains(t, output, `"name":`)
	assertContains(t, output, `"Alice"`)
	assertContains(t, output, `"age":`)
	assertContains(t, output, `30`)
}

func TestWordArrayCodegen(t *testing.T) {
	input := `def main
  words = %w{foo bar baz}
end`

	output := compile(t, input)

	// Should be a string array
	assertContains(t, output, `[]string{`)
	assertContains(t, output, `"foo"`)
	assertContains(t, output, `"bar"`)
	assertContains(t, output, `"baz"`)
}

func TestEmptyWordArrayCodegen(t *testing.T) {
	input := `def main
  empty = %w{}
end`

	output := compile(t, input)

	// Empty word array generates []any{} since type can't be inferred
	assertContains(t, output, `[]any{}`)
}

func TestArraySplatCodegen(t *testing.T) {
	input := `def main
  rest = [2, 3]
  all = [1, *rest, 4]
end`

	output := compile(t, input)

	// Should generate IIFE for splat with unique var names
	assertContains(t, output, `func() []any {`)
	assertContains(t, output, `_arr0 := []any{}`)
	assertContains(t, output, `append(_arr0`)
	assertContains(t, output, `for _, _v0 := range rest`)
}

func TestMapDoubleSplatCodegen(t *testing.T) {
	input := `def main
  defaults = {"a" => 1}
  m = {**defaults, b: 2}
end`

	output := compile(t, input)

	// Should generate IIFE for splat with unique var names
	assertContains(t, output, `func() map[any]any {`)
	assertContains(t, output, `_map0 := map[any]any{}`)
	assertContains(t, output, `for _k0, _v0 := range defaults`)
	assertContains(t, output, `_map0[_k0] = _v0`)
	assertContains(t, output, `_map0["b"] = 2`)
}

func TestImplicitValueShorthandCodegen(t *testing.T) {
	input := `def main
  name = "Alice"
  age = 30
  m = {name:, age:}
end`

	output := compile(t, input)

	// Keys should be strings, values should be identifiers
	assertContains(t, output, `"name": name`)
	assertContains(t, output, `"age": age`)
}

func TestMultipleSplatsCodegen(t *testing.T) {
	input := `def main
  a = [1, 2]
  b = [3, 4]
  all = [*a, 0, *b]
end`

	output := compile(t, input)

	// Should generate IIFE with multiple splats
	assertContains(t, output, `func() []any {`)
	assertContains(t, output, `range a`)
	assertContains(t, output, `range b`)
}

func TestMultipleDoubleSplatsCodegen(t *testing.T) {
	input := `def main
  defaults = {"a" => 1}
  overrides = {"b" => 2}
  m = {**defaults, **overrides, c: 3}
end`

	output := compile(t, input)

	// Should generate IIFE with multiple splats
	assertContains(t, output, `func() map[any]any {`)
	assertContains(t, output, `range defaults`)
	assertContains(t, output, `range overrides`)
	assertContains(t, output, `"c"] = 3`)
}

func TestSplatOnlyArrayCodegen(t *testing.T) {
	input := `def main
  items = [1, 2, 3]
  copy = [*items]
end`

	output := compile(t, input)

	// Should generate IIFE even for single splat
	assertContains(t, output, `func() []any {`)
	assertContains(t, output, `range items`)
}

func TestInterpolatedWordArrayCodegen(t *testing.T) {
	input := `def main
  name = "world"
  words = %W{hello #{name}}
end`

	output := compile(t, input)

	// Should contain string concatenation for interpolation
	assertContains(t, output, `"hello"`)
	// The interpolated element should reference the name variable
	assertContains(t, output, `name`)
}

func TestMixedMapSyntaxCodegen(t *testing.T) {
	input := `def main
  m = {"explicit" => 1, short: 2}
end`

	output := compile(t, input)

	// Both syntaxes should produce string keys
	assertContains(t, output, `"explicit": 1`)
	assertContains(t, output, `"short": 2`)
}
