package formatter

import (
	"strings"
	"testing"
)

func TestFormatBasicFunction(t *testing.T) {
	input := `def hello
puts("Hello")
end`

	expected := `def hello
  puts("Hello")
end
`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if result != expected {
		t.Errorf("Format mismatch:\nGot:\n%s\nExpected:\n%s", result, expected)
	}
}

func TestFormatWithComments(t *testing.T) {
	input := `# This is a comment
def greet(name: String)
  # Say hello
  puts("Hello, #{name}")
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "# This is a comment") {
		t.Error("Leading comment not preserved")
	}
	if !strings.Contains(result, "# Say hello") {
		t.Error("Body comment not preserved")
	}
}

func TestFormatIdempotent(t *testing.T) {
	input := `def add(a: Int, b: Int): Int
  return a + b
end
`

	first, err := Format(input)
	if err != nil {
		t.Fatalf("First format error: %v", err)
	}

	second, err := Format(first)
	if err != nil {
		t.Fatalf("Second format error: %v", err)
	}

	if first != second {
		t.Errorf("Format is not idempotent:\nFirst:\n%s\nSecond:\n%s", first, second)
	}
}

func TestFormatClass(t *testing.T) {
	input := `class User
def initialize(name: String)
@name = name
end
def greet
puts(@name)
end
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// Check proper indentation
	if !strings.Contains(result, "  def initialize") {
		t.Error("Method should be indented inside class")
	}
	if !strings.Contains(result, "    @name = name") {
		t.Error("Body should be double-indented inside method in class")
	}
}

func TestFormatIfStatement(t *testing.T) {
	input := `if x > 0
puts("positive")
elsif x < 0
puts("negative")
else
puts("zero")
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "if x > 0") {
		t.Error("if condition not formatted correctly")
	}
	if !strings.Contains(result, "elsif x < 0") {
		t.Error("elsif not present")
	}
	if !strings.Contains(result, "  puts") {
		t.Error("body should be indented")
	}
}

func TestFormatForLoop(t *testing.T) {
	input := `for i in 1..10
puts(i)
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "for i in 1..10") {
		t.Error("for loop header not formatted correctly")
	}
	if !strings.Contains(result, "  puts") {
		t.Error("body should be indented")
	}
}

func TestFormatBlock(t *testing.T) {
	// Single-statement blocks use {} style
	input := `items.each { |x| puts(x) }`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// Single-line blocks use {} with params
	if !strings.Contains(result, "{|x|") && !strings.Contains(result, "{ |x|") {
		t.Errorf("single-statement block should use {} style, got: %s", result)
	}
}

func TestFormatMultiLineBlock(t *testing.T) {
	// Multi-statement blocks use do...end style
	input := `items.each do |x|
puts(x)
puts(x * 2)
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "do |x|") {
		t.Errorf("multi-statement block should use do...end style, got: %s", result)
	}
}

func TestFormatImport(t *testing.T) {
	input := `import fmt
import net/http as http

def main
puts("hello")
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "import fmt") {
		t.Error("import not preserved")
	}
	if !strings.Contains(result, "import net/http as http") {
		t.Error("aliased import not preserved")
	}
}

func TestFormatOperatorSpacing(t *testing.T) {
	input := `x=1+2*3`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "x = 1 + 2 * 3") {
		t.Errorf("operators should have spaces: got %s", result)
	}
}

func TestFormatPrecedencePreserved(t *testing.T) {
	// Ensure parentheses are added when needed to preserve semantics
	tests := []struct {
		input    string
		expected string
	}{
		{"x = (1 + 2) * 3", "x = (1 + 2) * 3"},
		{"x = 1 + 2 * 3", "x = 1 + 2 * 3"},     // no parens needed
		{"x = 1 * 2 + 3", "x = 1 * 2 + 3"},     // no parens needed
		{"x = (1 + 2) + 3", "x = 1 + 2 + 3"},   // parens not needed (same precedence, left-assoc)
		{"x = 1 + (2 + 3)", "x = 1 + (2 + 3)"}, // parens needed on right (right-assoc differs)
	}

	for _, tc := range tests {
		result, err := Format(tc.input)
		if err != nil {
			t.Fatalf("Format error for %q: %v", tc.input, err)
		}
		result = strings.TrimSpace(result)
		if result != tc.expected {
			t.Errorf("Precedence test failed:\n  input: %s\n  got: %s\n  want: %s", tc.input, result, tc.expected)
		}
	}
}

func TestFormatArrayAndMap(t *testing.T) {
	input := `arr = [1,2,3]
m = {"a"=>1,"b"=>2}`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "[1, 2, 3]") {
		t.Error("array should have spaces after commas")
	}
	if !strings.Contains(result, `"a" => 1`) {
		t.Error("map should have spaces around =>")
	}
}

func TestFormatInterface(t *testing.T) {
	input := `interface Speaker
def speak: String
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "interface Speaker") {
		t.Error("interface declaration not preserved")
	}
	if !strings.Contains(result, "  def speak: String") {
		t.Error("method signature should be indented")
	}
}

func TestFormatMultiLineDocComment(t *testing.T) {
	input := `# This is the first line of documentation
# This is the second line
# And this is the third line
def greet(name: String)
  puts("Hello, " + name)
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// All three comment lines should be preserved
	if !strings.Contains(result, "# This is the first line of documentation") {
		t.Error("First doc comment line not preserved")
	}
	if !strings.Contains(result, "# This is the second line") {
		t.Error("Second doc comment line not preserved")
	}
	if !strings.Contains(result, "# And this is the third line") {
		t.Error("Third doc comment line not preserved")
	}
}

func TestFormatFreeFloatingComment(t *testing.T) {
	input := `# This is a free-floating comment at the top

def hello
  puts("hi")
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// Free-floating comment should be preserved
	if !strings.Contains(result, "# This is a free-floating comment at the top") {
		t.Error("Free-floating comment not preserved")
	}
}

func TestFormatCommentBetweenStatements(t *testing.T) {
	input := `def hello
  x = 1
  # This is a comment between statements
  y = 2
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "# This is a comment between statements") {
		t.Error("Comment between statements not preserved")
	}
}

func TestFormatCommentAtEndOfFile(t *testing.T) {
	input := `def hello
  puts("hi")
end
# Comment at end of file`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "# Comment at end of file") {
		t.Error("End of file comment not preserved")
	}
}

func TestFormatClassWithMethodComments(t *testing.T) {
	input := `# User class represents a user
class User
  # Initialize with a name
  def initialize(name: String)
    @name = name
  end

  # Greet the user
  def greet
    puts("Hello, " + @name)
  end
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "# User class represents a user") {
		t.Error("Class doc comment not preserved")
	}
	if !strings.Contains(result, "# Initialize with a name") {
		t.Error("initialize method doc comment not preserved")
	}
	if !strings.Contains(result, "# Greet the user") {
		t.Error("greet method doc comment not preserved")
	}
}

func TestFormatCommentIdempotent(t *testing.T) {
	input := `# Doc comment
def hello
  # Body comment
  puts("hi")
end
# End comment`

	first, err := Format(input)
	if err != nil {
		t.Fatalf("First format error: %v", err)
	}

	second, err := Format(first)
	if err != nil {
		t.Fatalf("Second format error: %v", err)
	}

	if first != second {
		t.Errorf("Format with comments is not idempotent:\nFirst:\n%s\nSecond:\n%s", first, second)
	}
}

func TestFormatImportWithComment(t *testing.T) {
	input := `# HTTP client library
import net/http as http

def main
  puts("hello")
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "# HTTP client library") {
		t.Error("Import doc comment not preserved")
	}
}

func TestFormatIfWithComment(t *testing.T) {
	input := `# Check the value
if x > 0
  puts("positive")
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "# Check the value") {
		t.Error("If statement doc comment not preserved")
	}
}

func TestFormatForLoopWithComment(t *testing.T) {
	input := `# Iterate over items
for i in 1..10
  puts(i)
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "# Iterate over items") {
		t.Error("For loop doc comment not preserved")
	}
}

func TestFormatWhileWithComment(t *testing.T) {
	input := `# Loop while condition is true
while x > 0
  x = x - 1
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "# Loop while condition is true") {
		t.Error("While loop doc comment not preserved")
	}
}

func TestFormatMultipleFunctionsWithComments(t *testing.T) {
	input := `# First function
def first
  puts("first")
end

# Second function
def second
  puts("second")
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "# First function") {
		t.Error("First function doc comment not preserved")
	}
	if !strings.Contains(result, "# Second function") {
		t.Error("Second function doc comment not preserved")
	}
}

// Phase 1: Core Expressions Tests

func TestFormatLambdaExpr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lambda",
			input:    `x = -> { 1 }`,
			expected: "-> { 1 }",
		},
		{
			name:     "lambda with param",
			input:    `x = -> { |n: Int| n * 2 }`,
			expected: "-> { |n: Int| n * 2 }",
		},
		{
			name:     "lambda with return type",
			input:    `x = -> { |n: Int|: Int n * 2 }`,
			expected: "-> { |n: Int|: Int n * 2 }",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Format(tc.input)
			if err != nil {
				t.Fatalf("Format error: %v", err)
			}
			if !strings.Contains(result, tc.expected) {
				t.Errorf("Expected %q in output, got: %s", tc.expected, result)
			}
		})
	}
}

func TestFormatBangExpr(t *testing.T) {
	input := `x = call()!`
	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !strings.Contains(result, "call()!") {
		t.Errorf("Expected bang expr, got: %s", result)
	}
}

func TestFormatNilCoalesceExpr(t *testing.T) {
	input := `x = maybe() ?? "default"`
	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !strings.Contains(result, `maybe() ?? "default"`) {
		t.Errorf("Expected nil coalesce, got: %s", result)
	}
}

func TestFormatRescueExpr(t *testing.T) {
	input := `x = risky() rescue 0`
	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !strings.Contains(result, "risky() rescue 0") {
		t.Errorf("Expected rescue expr, got: %s", result)
	}
}

func TestFormatSafeNavExpr(t *testing.T) {
	input := `x = obj&.method`
	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !strings.Contains(result, "obj&.method") {
		t.Errorf("Expected safe nav, got: %s", result)
	}
}

func TestFormatTernaryExpr(t *testing.T) {
	input := `x = cond ? "yes" : "no"`
	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !strings.Contains(result, `cond ? "yes" : "no"`) {
		t.Errorf("Expected ternary, got: %s", result)
	}
}

// Phase 2: Type Declarations Tests

func TestFormatStructDecl(t *testing.T) {
	input := `struct Point
x: Int
y: Int
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "struct Point") {
		t.Error("Expected struct declaration")
	}
	if !strings.Contains(result, "  x: Int") {
		t.Error("Expected indented field")
	}
}

func TestFormatEnumDecl(t *testing.T) {
	input := `enum Status
Pending
Active
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "enum Status") {
		t.Error("Expected enum declaration")
	}
	if !strings.Contains(result, "  Pending") {
		t.Error("Expected indented variant")
	}
}

func TestFormatEnumWithValues(t *testing.T) {
	input := `enum HttpStatus
Ok = 200
NotFound = 404
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "Ok = 200") {
		t.Error("Expected enum with explicit value")
	}
}

func TestFormatConstDecl(t *testing.T) {
	input := `const MAX = 100`
	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !strings.Contains(result, "const MAX = 100") {
		t.Errorf("Expected const decl, got: %s", result)
	}
}

func TestFormatTypeAliasDecl(t *testing.T) {
	input := `type UserID = Int64`
	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !strings.Contains(result, "type UserID = Int64") {
		t.Errorf("Expected type alias, got: %s", result)
	}
}

func TestFormatModuleDecl(t *testing.T) {
	input := `module Loggable
def log(msg: String)
puts(msg)
end
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "module Loggable") {
		t.Error("Expected module declaration")
	}
	if !strings.Contains(result, "  def log") {
		t.Error("Expected indented method")
	}
}

// Phase 3: Type Parameters Tests

func TestFormatGenericFunction(t *testing.T) {
	input := `def identity<T>(x: T): T
x
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "def identity<T>(x: T): T") {
		t.Errorf("Expected generic function, got: %s", result)
	}
}

func TestFormatGenericClass(t *testing.T) {
	input := `class Box<T>
def initialize(@value: T)
end
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "class Box<T>") {
		t.Errorf("Expected generic class, got: %s", result)
	}
}

func TestFormatTypeConstraint(t *testing.T) {
	input := `def compare<T: Comparable>(a: T, b: T): Bool
a == b
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "def compare<T: Comparable>") {
		t.Errorf("Expected type constraint, got: %s", result)
	}
}

// Phase 4: Control Flow Tests

func TestFormatIfLet(t *testing.T) {
	input := `if let name = find()
puts(name)
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "if let name = find()") {
		t.Errorf("Expected if-let syntax, got: %s", result)
	}
}

func TestFormatForWithTwoVars(t *testing.T) {
	input := `for key, value in map
puts(key)
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "for key, value in map") {
		t.Errorf("Expected two-var for loop, got: %s", result)
	}
}

func TestFormatMultiAssign(t *testing.T) {
	input := `a, b = pair`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "a, b = pair") {
		t.Errorf("Expected multi-assign, got: %s", result)
	}
}

// Phase 5: Class Features Tests

func TestFormatAccessors(t *testing.T) {
	input := `class Person
getter name: String
setter age: Int
property email: String
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "  getter name: String") {
		t.Error("Expected getter")
	}
	if !strings.Contains(result, "  setter age: Int") {
		t.Error("Expected setter")
	}
	if !strings.Contains(result, "  property email: String") {
		t.Error("Expected property")
	}
}

func TestFormatInclude(t *testing.T) {
	input := `class Worker
include Loggable
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "include Loggable") {
		t.Errorf("Expected include, got: %s", result)
	}
}

func TestFormatClassMethod(t *testing.T) {
	input := `class Counter
def self.count: Int
0
end
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "def self.count: Int") {
		t.Errorf("Expected class method, got: %s", result)
	}
}

// Phase 6: Concurrency Tests

func TestFormatSpawnAwait(t *testing.T) {
	input := `t = spawn { 42 }
result = await t`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "spawn {") {
		t.Error("Expected spawn expression")
	}
	if !strings.Contains(result, "await t") {
		t.Error("Expected await expression")
	}
}

func TestFormatChanSend(t *testing.T) {
	input := `ch << 42`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "ch << 42") {
		t.Errorf("Expected channel send, got: %s", result)
	}
}

// Phase 7: Additional Expressions Tests

func TestFormatStructLit(t *testing.T) {
	input := `p = Point{x: 10, y: 20}`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "Point{x: 10, y: 20}") {
		t.Errorf("Expected struct lit, got: %s", result)
	}
}

func TestFormatSetLit(t *testing.T) {
	input := `s = Set{1, 2, 3}`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "Set{1, 2, 3}") {
		t.Errorf("Expected set lit, got: %s", result)
	}
}

func TestFormatScopeExpr(t *testing.T) {
	input := `x = Status::Active`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "Status::Active") {
		t.Errorf("Expected scope expr, got: %s", result)
	}
}

func TestFormatSuper(t *testing.T) {
	input := `class Child < Parent
def initialize(name: String)
super(name)
end
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "super(name)") {
		t.Errorf("Expected super call, got: %s", result)
	}
}

func TestFormatVariadicParam(t *testing.T) {
	input := `def sum(*nums: Int): Int
0
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "*nums: Int") {
		t.Errorf("Expected variadic param, got: %s", result)
	}
}

func TestFormatDefaultParam(t *testing.T) {
	input := `def greet(name: String = "World")
puts(name)
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, `name: String = "World"`) {
		t.Errorf("Expected default param, got: %s", result)
	}
}

// Spec-Style Formatting Tests
// These test that messy input gets formatted into clean, canonical output.

func TestFormatSpec(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic function formatting
		{
			name: "function with poor indentation",
			input: `def greet(name: String)
puts("Hello")
puts(name)
end`,
			expected: `def greet(name: String)
  puts("Hello")
  puts(name)
end
`,
		},
		{
			name: "function with cramped params",
			input: `def add(a: Int,b: Int): Int
a+b
end`,
			expected: `def add(a: Int, b: Int): Int
  a + b
end
`,
		},

		// Operator spacing
		{
			name:     "operators without spaces",
			input:    `x=1+2*3-4/2`,
			expected: "x = 1 + 2 * 3 - 4 / 2\n",
		},
		{
			name:     "comparison operators",
			input:    `result=a>b && c<=d || e!=f`,
			expected: "result = a > b && c <= d || e != f\n",
		},

		// Array and map literals
		{
			name:     "array without spaces",
			input:    `arr=[1,2,3,4,5]`,
			expected: "arr = [1, 2, 3, 4, 5]\n",
		},
		{
			name:     "map without spaces",
			input:    `m={"a"=>1,"b"=>2}`,
			expected: "m = {\"a\" => 1, \"b\" => 2}\n",
		},

		// Control flow indentation
		{
			name: "if/elsif/else indentation",
			input: `if x>0
puts("positive")
elsif x<0
puts("negative")
else
puts("zero")
end`,
			expected: `if x > 0
  puts("positive")
elsif x < 0
  puts("negative")
else
  puts("zero")
end
`,
		},
		{
			name: "nested if statements",
			input: `if a
if b
puts("both")
end
end`,
			expected: `if a
  if b
    puts("both")
  end
end
`,
		},
		{
			name: "while loop",
			input: `while x>0
x=x-1
end`,
			expected: `while x > 0
  x = x - 1
end
`,
		},
		{
			name: "for loop",
			input: `for i in 1..10
puts(i)
end`,
			expected: `for i in 1..10
  puts(i)
end
`,
		},
		{
			name: "for loop with two vars",
			input: `for k,v in map
puts(k)
end`,
			expected: `for k, v in map
  puts(k)
end
`,
		},

		// Class formatting
		{
			name: "class with methods",
			input: `class User
def initialize(@name: String)
end
def greet
puts(@name)
end
end`,
			expected: `class User
  @name : String

  def initialize(@name: String)
  end
  def greet
    puts(@name)
  end
end
`,
		},
		{
			name: "class with inheritance",
			input: `class Admin < User
def role: String
"admin"
end
end`,
			expected: `class Admin < User
  def role: String
    "admin"
  end
end
`,
		},
		{
			name: "class with accessors",
			input: `class Person
getter name: String
setter age: Int
property email: String
end`,
			expected: `class Person
  getter name: String
  setter age: Int
  property email: String
end
`,
		},

		// Interface formatting
		{
			name: "interface declaration",
			input: `interface Speaker
def speak: String
def volume: Int
end`,
			expected: `interface Speaker
  def speak: String
  def volume: Int
end
`,
		},

		// Struct formatting
		{
			name: "struct declaration",
			input: `struct Point
x: Int
y: Int
end`,
			expected: `struct Point
  x: Int
  y: Int
end
`,
		},

		// Enum formatting
		{
			name: "enum declaration",
			input: `enum Status
Pending
Active
Done
end`,
			expected: `enum Status
  Pending
  Active
  Done
end
`,
		},
		{
			name: "enum with values",
			input: `enum HttpCode
Ok=200
NotFound=404
end`,
			expected: `enum HttpCode
  Ok = 200
  NotFound = 404
end
`,
		},

		// Module formatting
		{
			name: "module declaration",
			input: `module Loggable
def log(msg: String)
puts(msg)
end
end`,
			expected: `module Loggable
  def log(msg: String)
    puts(msg)
  end
end
`,
		},

		// Generics
		{
			name: "generic function",
			input: `def identity<T>(x: T): T
x
end`,
			expected: `def identity<T>(x: T): T
  x
end
`,
		},
		{
			name: "generic with constraint",
			input: `def max<T: Comparable>(a: T, b: T): T
a
end`,
			expected: `def max<T: Comparable>(a: T, b: T): T
  a
end
`,
		},

		// Lambda expressions
		{
			name:     "simple lambda",
			input:    `f=->{ 42 }`,
			expected: "f = -> { 42 }\n",
		},
		{
			name:     "lambda with params",
			input:    `f = -> { |x: Int| x * 2 }`,
			expected: "f = -> { |x: Int| x * 2 }\n",
		},

		// Error handling
		{
			name:     "bang operator",
			input:    `x=call()!`,
			expected: "x = call()!\n",
		},
		{
			name:     "nil coalescing",
			input:    `x=maybe()??"default"`,
			expected: "x = maybe() ?? \"default\"\n",
		},
		{
			name:     "rescue expression",
			input:    `x=risky() rescue 0`,
			expected: "x = risky() rescue 0\n",
		},
		{
			name:     "ternary operator",
			input:    `x = cond ? "yes" : "no"`,
			expected: "x = cond ? \"yes\" : \"no\"\n",
		},

		// Blocks
		{
			name:     "single-line block",
			input:    `items.each{|x|puts(x)}`,
			expected: "items.each {|x| puts(x)}\n",
		},
		{
			name: "multi-line block",
			input: `items.each do |x|
puts(x)
puts(x*2)
end`,
			expected: `items.each do |x|
  puts(x)
  puts(x * 2)
end
`,
		},

		// Case statement
		{
			name: "case statement",
			input: `case x
when 1
puts("one")
when 2,3
puts("two or three")
else
puts("other")
end`,
			expected: `case x
when 1
  puts("one")
when 2, 3
  puts("two or three")
else
  puts("other")
end
`,
		},

		// If-let
		{
			name: "if-let binding",
			input: `if let name=find()
puts(name)
end`,
			expected: `if let name = find()
  puts(name)
end
`,
		},

		// Concurrency
		{
			name:     "spawn expression",
			input:    `t=spawn{42}`,
			expected: "t = spawn {42}\n",
		},
		{
			name:     "await expression",
			input:    `result=await t`,
			expected: "result = await t\n",
		},
		{
			name:     "channel send",
			input:    `ch<<42`,
			expected: "ch << 42\n",
		},

		// Struct literals
		{
			name:     "struct literal",
			input:    `p=Point{x:10,y:20}`,
			expected: "p = Point{x: 10, y: 20}\n",
		},
		{
			name:     "set literal",
			input:    `s=Set{1,2,3}`,
			expected: "s = Set{1, 2, 3}\n",
		},

		// Scope resolution
		{
			name:     "scope operator",
			input:    `x=Status::Active`,
			expected: "x = Status::Active\n",
		},

		// Imports
		{
			name: "multiple imports",
			input: `import fmt
import net/http as http

def main
puts("hello")
end`,
			expected: `import fmt
import net/http as http

def main
  puts("hello")
end
`,
		},

		// Multi-assignment
		{
			name:     "tuple unpacking",
			input:    `a,b=pair`,
			expected: "a, b = pair\n",
		},
		{
			name:     "splat assignment",
			input:    `first,*rest=items`,
			expected: "first, *rest = items\n",
		},

		// Compound assignment
		{
			name:     "compound add",
			input:    `x+=1`,
			expected: "x += 1\n",
		},
		{
			name:     "or-assign",
			input:    `x||=default`,
			expected: "x ||= default\n",
		},

		// Variadic and default params
		{
			name: "variadic param",
			input: `def sum(*nums: Int): Int
0
end`,
			expected: `def sum(*nums: Int): Int
  0
end
`,
		},
		{
			name: "default param",
			input: `def greet(name: String = "World")
puts(name)
end`,
			expected: `def greet(name: String = "World")
  puts(name)
end
`,
		},

		// Range literals
		{
			name:     "inclusive range",
			input:    `r=1..10`,
			expected: "r = 1..10\n",
		},
		{
			name:     "exclusive range",
			input:    `r=0...n`,
			expected: "r = 0...n\n",
		},

		// Safe navigation
		{
			name:     "safe nav operator",
			input:    `x=obj&.method`,
			expected: "x = obj&.method\n",
		},

		// Statement modifiers
		{
			name:     "return if",
			input:    `return 0 if error`,
			expected: "return 0 if error\n",
		},
		{
			name:     "return unless",
			input:    `return unless valid`,
			expected: "return unless valid\n",
		},
		{
			name:     "break if",
			input:    `break if done`,
			expected: "break if done\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Format(tc.input)
			if err != nil {
				t.Fatalf("Format error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Format mismatch:\nInput:\n%s\nExpected:\n%s\nGot:\n%s", tc.input, tc.expected, result)
			}
		})
	}
}

// Deep Nesting Test - ensures proper indentation at many levels
func TestFormatDeepNesting(t *testing.T) {
	input := `def process
for item in items
if item.valid
case item.kind
when "a"
items.each do |d|
if let value = d.get("key")
while value.length > 0
if value.first == "x"
puts("found x")
else
puts("not x")
end
value = value.rest
end
end
end
when "b"
puts("type b")
else
puts("unknown")
end
else
puts("invalid")
end
end
end`

	expected := `def process
  for item in items
    if item.valid
      case item.kind
      when "a"
        items.each do |d|
          if let value = d.get("key")
            while value.length > 0
              if value.first == "x"
                puts("found x")
              else
                puts("not x")
              end
              value = value.rest
            end
          end
        end
      when "b"
        puts("type b")
      else
        puts("unknown")
      end
    else
      puts("invalid")
    end
  end
end
`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if result != expected {
		t.Errorf("Deep nesting format mismatch:\nExpected:\n%s\nGot:\n%s", expected, result)
	}

	// Verify idempotency
	second, err := Format(result)
	if err != nil {
		t.Fatalf("Second format error: %v", err)
	}
	if result != second {
		t.Errorf("Deep nesting not idempotent:\nFirst:\n%s\nSecond:\n%s", result, second)
	}
}

// Kitchen Sink Test - comprehensive example with many language features
func TestFormatKitchenSink(t *testing.T) {
	input := `import fmt
import net/http as http

const MAX_SIZE = 1024

type UserID = Int64

enum Status
Pending
Active
Done = 100
end

struct Point
x: Int
y: Int

def sum: Int
@x + @y
end
end

interface Runnable
def run: Error?
def stop
end

module Loggable
def log(msg: String)
puts("[LOG] #{msg}")
end
end

class Worker < Base
include Loggable

getter name: String
setter status: Status

def initialize(@name: String)
end

def run: Error?
log("Starting #{@name}")
items.each {|x| process(x)}

tasks.map do |task|
result = task.execute
result * 2
end

for key, value in config
puts("#{key}: #{value}")
end

ch = Chan[Int].new(10)
ch << 42

t = spawn { compute() }
result = await t

value = risky() rescue "default"
name = maybe() ?? "anonymous"
label = active? ? "on" : "off"
data = fetch()!

nil
end

private def process(item: Item)
puts(item)
end
end

def identity<T>(x: T): T
x
end

def greet(name: String = "World")
puts("Hello, #{name}!")
end

def main
p = Point{x: 10, y: 20}
s = Set{1, 2, 3}
status = Status::Active

a, b = get_pair()
first, *rest = get_items()

1.upto(10) {|n| puts(n)}
5.times {|i| puts(i)}

nums.reduce(0) {|acc, n| acc + n}
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// Check key features are present and properly formatted
	checks := []string{
		"import fmt",
		"const MAX_SIZE = 1024",
		"type UserID = Int64",
		"enum Status",
		"  Pending",
		"  Done = 100",
		"struct Point",
		"  x: Int",
		"  def sum: Int",
		"interface Runnable",
		"  def run: Error?",
		"module Loggable",
		"  def log(msg: String)",
		"class Worker < Base",
		"  include Loggable",
		"  getter name: String",
		"  def initialize(@name: String)",
		"  def run: Error?",
		"    log(\"Starting #{@name}\")",
		"    items.each {|x| process(x)}",
		"    for key, value in config",
		"    ch << 42",
		"    t = spawn {",
		"    result = await t",
		"    value = risky() rescue \"default\"",
		"    name = maybe() ?? \"anonymous\"",
		"    label = active? ? \"on\" : \"off\"",
		"    data = fetch()!",
		"  private def process(item: Item)",
		"def identity<T>(x: T): T",
		"def greet(name: String = \"World\")",
		"def main",
		"  p = Point{x: 10, y: 20}",
		"  s = Set{1, 2, 3}",
		"  status = Status::Active",
		"  a, b = get_pair()",
		"  first, *rest = get_items()",
		"  1.upto(10) {|n| puts(n)}",
		"  nums.reduce(0) {|acc, n| acc + n}",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("Kitchen sink missing expected content: %q\nFull output:\n%s", check, result)
		}
	}

	// Verify idempotency
	second, err := Format(result)
	if err != nil {
		t.Fatalf("Second format error: %v", err)
	}
	if result != second {
		t.Errorf("Kitchen sink not idempotent:\nFirst:\n%s\nSecond:\n%s", result, second)
	}
}

// Edge Case Tests

func TestFormatEmptyProgram(t *testing.T) {
	result, err := Format("")
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty output, got: %q", result)
	}
}

func TestFormatOnlyComments(t *testing.T) {
	input := `# Just a comment`
	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !strings.Contains(result, "# Just a comment") {
		t.Error("Comment should be preserved")
	}
}

// Idempotency Tests for New Features

func TestFormatNewFeaturesIdempotent(t *testing.T) {
	inputs := []string{
		`double = -> { |x: Int| x * 2 }`,
		`x = maybe() ?? "default"`,
		`x = cond ? "yes" : "no"`,
		`struct Point
  x: Int
  y: Int
end`,
		`enum Status
  Pending
  Active
end`,
		`def identity<T>(x: T): T
  x
end`,
		`if let name = find()
  puts(name)
end`,
		`for key, value in map
  puts(key)
end`,
	}

	for _, input := range inputs {
		first, err := Format(input)
		if err != nil {
			t.Fatalf("First format error for %q: %v", input[:20], err)
		}

		second, err := Format(first)
		if err != nil {
			t.Fatalf("Second format error: %v", err)
		}

		if first != second {
			t.Errorf("Not idempotent for input starting %q:\nFirst:\n%s\nSecond:\n%s", input[:20], first, second)
		}
	}
}
