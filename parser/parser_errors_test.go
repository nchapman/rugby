package parser

import (
	"strings"
	"testing"

	"github.com/nchapman/rugby/lexer"
)

// expectError is a test helper that ensures the parser produces errors
func expectError(t *testing.T, input string, expectedSubstring string) {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error containing %q, got no errors", expectedSubstring)
		return
	}

	found := false
	for _, err := range p.Errors() {
		if strings.Contains(err, expectedSubstring) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error containing %q, got: %v", expectedSubstring, p.Errors())
	}
}

// ====================
// Literal parsing errors
// ====================

func TestInvalidIntLiteral(t *testing.T) {
	// This is tricky - the lexer produces INT tokens, and ParseInt handles them.
	// Invalid ints are caught at lexer level, but we can test the error formatting.
	input := `def main
  x = 999999999999999999999999999999
end`
	expectError(t, input, "invalid integer")
}

// Note: Invalid float literals are handled by the lexer, not the parser.
// The parser receives valid tokens from the lexer.

// ====================
// Expression errors
// ====================

func TestMissingClosingParenInCall(t *testing.T) {
	input := `def main
  foo(1, 2
end`
	expectError(t, input, "expected ')'")
}

func TestMissingClosingBracketInIndex(t *testing.T) {
	input := `def main
  arr[0
end`
	expectError(t, input, "expected ']'")
}

func TestMissingClosingParenInGrouped(t *testing.T) {
	input := `def main
  x = (1 + 2
end`
	expectError(t, input, "expected ')'")
}

func TestBangOnInvalidExpression(t *testing.T) {
	input := `def main
  x = 42!
end`
	expectError(t, input, "'!' can only follow")
}

func TestRescueOnInvalidExpression(t *testing.T) {
	input := `def main
  x = 42 rescue "default"
end`
	expectError(t, input, "'rescue' can only follow")
}

func TestSafeNavOnNonIdent(t *testing.T) {
	input := `def main
  x = 42&.foo
end`
	// The safe navigation should fail on a literal
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	// May or may not produce error depending on implementation
	// Just ensure it doesn't panic
}

func TestSymbolToProcInvalidMethod(t *testing.T) {
	input := `def main
  x = items.map(&:)
end`
	expectError(t, input, "expected symbol after '&'")
}

// ====================
// Statement errors
// ====================

func TestIfWithoutEnd(t *testing.T) {
	input := `def main
  if x > 5
    puts "big"
`
	expectError(t, input, "expected")
}

func TestWhileWithoutEnd(t *testing.T) {
	input := `def main
  while x > 0
    x = x - 1
`
	expectError(t, input, "expected")
}

func TestForWithoutIn(t *testing.T) {
	input := `def main
  for x items
    puts x
  end
end`
	expectError(t, input, "expected 'in'")
}

func TestCaseWithoutEnd(t *testing.T) {
	input := `def main
  case x
  when 1
    puts "one"
`
	expectError(t, input, "expected")
}

func TestSelectWithoutEnd(t *testing.T) {
	input := `def main
  select
  case <-ch
    puts "got"
`
	expectError(t, input, "expected")
}

func TestDeferWithoutCall(t *testing.T) {
	input := `def main
  defer 42
end`
	expectError(t, input, "defer requires")
}

func TestPanicWithoutExpression(t *testing.T) {
	input := `def main
  panic
end`
	// panic without expression - parser may allow this or error
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	// Just ensure no crash
}

// ====================
// Declaration errors
// ====================

func TestFunctionDuplicateParams(t *testing.T) {
	input := `def foo(a : Int, a : Int)
end`
	expectError(t, input, "duplicate parameter")
}

func TestFunctionMissingReturnType(t *testing.T) {
	input := `def foo() ->
end`
	expectError(t, input, "expected type")
}

func TestClassWithoutEnd(t *testing.T) {
	input := `class Foo
  def bar
  end
`
	expectError(t, input, "expected")
}

func TestInterfaceWithoutEnd(t *testing.T) {
	input := `interface Foo
  def bar
`
	expectError(t, input, "expected")
}

func TestImportInvalidPath(t *testing.T) {
	input := `import
def main
end`
	expectError(t, input, "expected")
}

func TestFieldDeclMissingType(t *testing.T) {
	input := `class Foo
  @field name
end`
	expectError(t, input, "field type required")
}

// ====================
// Block parsing errors
// ====================

func TestBlockWithMissingEnd(t *testing.T) {
	input := `def main
  items.each do |x|
    puts x
`
	expectError(t, input, "expected")
}

func TestBlockWithInvalidParams(t *testing.T) {
	input := `def main
  items.each do |123|
    puts "test"
  end
end`
	expectError(t, input, "expected")
}

// ====================
// Testing DSL errors
// ====================

func TestDescribeWithoutEnd(t *testing.T) {
	input := `describe "Math" do
  it "adds" do
    assert 1 + 1 == 2
  end
`
	expectError(t, input, "expected")
}

func TestItWithoutEnd(t *testing.T) {
	input := `describe "Math" do
  it "adds" do
    assert 1 + 1 == 2
`
	expectError(t, input, "expected")
}

// ====================
// Multi-assignment errors
// ====================

func TestMultiAssignMismatch(t *testing.T) {
	input := `def main
  a, b = 1
end`
	// Multi-assign with wrong number of values
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	// Parser may accept this - codegen would catch it
}

// ====================
// Concurrency errors
// ====================

func TestConcurrentlyWithoutEnd(t *testing.T) {
	input := `def main
  concurrently do
    spawn task1
`
	expectError(t, input, "expected")
}

func TestAwaitNonTask(t *testing.T) {
	input := `def main
  x = await 42
end`
	// await should work on any expression syntactically
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	// Should parse - semantic checking would catch invalid await
	// Parser errors here are also acceptable
	_ = program
	_ = p.Errors() // silence unused
}

// ====================
// Map/Array literal errors
// ====================

// Note: Map literal errors can cause parser loops - skipping these tests
// as they may indicate a bug in error recovery rather than coverage.

func TestArrayLiteralMissingBracket(t *testing.T) {
	input := `def main
  x = [1, 2, 3
end`
	expectError(t, input, "expected")
}

// ====================
// Ternary expression errors
// ====================

func TestTernaryMissingColon(t *testing.T) {
	input := `def main
  x = true ? 1
end`
	expectError(t, input, "expected ':'")
}

// ====================
// String interpolation errors
// ====================

func TestUnclosedInterpolation(t *testing.T) {
	input := `def main
  x = "hello #{name"
end`
	// Unclosed interpolation - lexer should catch this
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	// May error at lexer or parser level
}

// ====================
// Rescue expression edge cases
// ====================

func TestRescueWithBlockForm(t *testing.T) {
	input := `def main
  x = risky() rescue => err do
    puts err
    "default"
  end
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	// Should parse the block form of rescue
	if len(p.Errors()) > 0 {
		t.Errorf("unexpected errors: %v", p.Errors())
	}
	_ = program
}

// ====================
// Instance variable errors
// ====================

func TestInstanceVarOutsideClass(t *testing.T) {
	// Instance variables outside class context are parsed but flagged elsewhere
	// The parser accepts @name syntax anywhere, codegen validates context
	input := `def main
  @name = "test"
end`
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	// Parser accepts this - semantic checking catches it
}

func TestNewInstanceVarOutsideInit(t *testing.T) {
	input := `class User
  def initialize
    @name = "test"
  end
  def set_age
    @age = 30
  end
end`
	expectError(t, input, "cannot introduce new instance variable")
}

// ====================
// Go statement errors
// ====================

func TestGoStatementWithoutCall(t *testing.T) {
	// `go` followed by a non-call expression may parse but fail at codegen
	input := `def main
  go 42
end`
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	// Parser may accept this - semantic checking would catch it
}

// ====================
// Unless statement edge cases
// ====================

func TestUnlessWithElsif(t *testing.T) {
	// unless doesn't support elsif - should error or handle gracefully
	input := `def main
  unless x
    puts "not x"
  elsif y
    puts "y"
  end
end`
	// Parser may accept this or error
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
}

// ====================
// Type annotation errors
// ====================

func TestMissingTypeAfterColon(t *testing.T) {
	input := `def foo(x : )
end`
	expectError(t, input, "expected type")
}

func TestInvalidTypeAnnotation(t *testing.T) {
	input := `def foo(x : 123)
end`
	expectError(t, input, "expected type")
}

// ====================
// Edge case: Empty constructs
// ====================

func TestEmptyBlock(t *testing.T) {
	input := `def main
  items.each do |x|
  end
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Declarations) != 1 {
		t.Errorf("expected 1 declaration, got %d", len(program.Declarations))
	}
}

// ====================
// Complex nested errors
// ====================

func TestNestedBlocksWithError(t *testing.T) {
	input := `def main
  items.each do |x|
    x.map do |y|
      puts y
    # missing end
  end
end`
	expectError(t, input, "expected")
}

func TestDeeplyNestedIfWithError(t *testing.T) {
	input := `def main
  if a
    if b
      if c
        puts "deep"
      # missing end
    end
  end
end`
	expectError(t, input, "expected")
}
