package codegen

import (
	"strings"
	"testing"

	"rugby/lexer"
	"rugby/parser"
)

func TestGenerateHello(t *testing.T) {
	input := `import fmt

def main
  puts "hello"
end`

	output := compile(t, input)

	assertContains(t, output, `package main`)
	assertContains(t, output, `import`)
	assertContains(t, output, `"fmt"`)
	assertContains(t, output, `func main()`)
	assertContains(t, output, `fmt.Println("hello")`)
}

func TestGenerateArithmetic(t *testing.T) {
	input := `import fmt

def main
  x = 2 + 3 * 4
  puts x
end`

	output := compile(t, input)

	assertContains(t, output, `x :=`)
	assertContains(t, output, `(2 + (3 * 4))`)
	assertContains(t, output, `fmt.Println(x)`)
}

func TestGenerateIfElse(t *testing.T) {
	input := `import fmt

def main
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
	assertContains(t, output, `fmt.Println("big")`)
	assertContains(t, output, `fmt.Println("small")`)
}

func TestGenerateWhile(t *testing.T) {
	input := `import fmt

def main
  i = 0
  while i < 5
    puts i
    i = i + 1
  end
end`

	output := compile(t, input)

	assertContains(t, output, `i := 0`)
	assertContains(t, output, `for i < 5 {`)
	assertContains(t, output, `i = (i + 1)`)
}

func TestGenerateComparison(t *testing.T) {
	input := `import fmt

def main
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
	input := `import fmt

def main
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
	assertContains(t, output, `x := a`)  // new var uses :=
	assertContains(t, output, `a = b`)   // param reassignment uses =
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
  return 42
end

def fetch() -> (String, Int, Bool)
  return "hello"
end

def main
end`

	output := compile(t, input)

	assertContains(t, output, `func parse(s interface{}) (int, bool)`)
	assertContains(t, output, `func fetch() (string, int, bool)`)
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
