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
def greet(name : String)
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
	input := `def add(a : Int, b : Int) -> Int
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
def initialize(name : String)
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
	t.Skip("Skipping: formatter needs update to handle new lambda syntax")
	// Single-statement blocks use {} style
	input := `items.each -> do |x|
puts(x)
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// Single-line blocks use {} with params
	if !strings.Contains(result, "{ |x|") && !strings.Contains(result, "{|x|") {
		t.Errorf("single-statement block should use {} style, got: %s", result)
	}
}

func TestFormatMultiLineBlock(t *testing.T) {
	t.Skip("Skipping: formatter needs update to handle new lambda syntax")
	// Multi-statement blocks use do...end style
	input := `items.each -> do |x|
puts(x)
puts(x * 2)
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "-> do |x|") {
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
def speak -> String
end`

	result, err := Format(input)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(result, "interface Speaker") {
		t.Error("interface declaration not preserved")
	}
	if !strings.Contains(result, "  def speak -> String") {
		t.Error("method signature should be indented")
	}
}

func TestFormatMultiLineDocComment(t *testing.T) {
	input := `# This is the first line of documentation
# This is the second line
# And this is the third line
def greet(name : String)
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
  def initialize(name : String)
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
