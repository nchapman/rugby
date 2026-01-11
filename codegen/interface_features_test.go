package codegen_test

import (
	"strings"
	"testing"

	"github.com/nchapman/rugby/codegen"
	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/parser"
)

func TestInterfaceInheritance(t *testing.T) {
	input := `
interface Reader
  def read -> String
end

interface Writer
  def write(data : String)
end

interface IO < Reader, Writer
  def close
end
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	g := codegen.New()
	output, err := g.Generate(program)
	if err != nil {
		t.Fatalf("Codegen error: %v", err)
	}

	// Check for interface embedding
	if !strings.Contains(output, "type IO interface {") {
		t.Error("Expected IO interface declaration")
	}
	if !strings.Contains(output, "\tReader") {
		t.Error("Expected Reader embedded in IO")
	}
	if !strings.Contains(output, "\tWriter") {
		t.Error("Expected Writer embedded in IO")
	}
}

func TestClassImplements(t *testing.T) {
	input := `
interface Speaker
  def speak -> String
end

interface Serializable
  def to_json -> String
end

class User implements Speaker, Serializable
  def initialize(@name : String)
  end
  
  def speak -> String
    @name
  end
  
  def to_json -> String
    @name
  end
end
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	g := codegen.New()
	output, err := g.Generate(program)
	if err != nil {
		t.Fatalf("Codegen error: %v", err)
	}

	// Check for compile-time conformance checks
	if !strings.Contains(output, "var _ Speaker = (*User)(nil)") {
		t.Error("Expected Speaker conformance check")
	}
	if !strings.Contains(output, "var _ Serializable = (*User)(nil)") {
		t.Error("Expected Serializable conformance check")
	}
}

func TestAnyKeyword(t *testing.T) {
	input := `
def log(thing : any)
  thing
end
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	g := codegen.New()
	output, err := g.Generate(program)
	if err != nil {
		t.Fatalf("Codegen error: %v", err)
	}

	// Check that 'any' maps to Go's any
	if !strings.Contains(output, "func log(thing any)") {
		t.Error("Expected 'any' to map to Go's any type")
	}
}

func TestIsAMethodCall(t *testing.T) {
	input := `
def check(obj : any) -> Bool
  obj.is_a?(String)
end
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	g := codegen.New()
	output, err := g.Generate(program)
	if err != nil {
		t.Fatalf("Codegen error: %v", err)
	}

	// Check for type assertion
	if !strings.Contains(output, "func() bool") {
		t.Error("Expected is_a? to compile to type assertion")
	}
	if !strings.Contains(output, "_, ok :=") {
		t.Error("Expected comma-ok idiom in is_a?")
	}
}

func TestAsMethodCall(t *testing.T) {
	input := `
def cast(obj : any) -> (String, Bool)
  obj.as(String)
end
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	g := codegen.New()
	output, err := g.Generate(program)
	if err != nil {
		t.Fatalf("Codegen error: %v", err)
	}

	// Check for type assertion with value
	if !strings.Contains(output, "v, ok :=") {
		t.Error("Expected value and ok from as()")
	}
}

func TestMessageMethodMapping(t *testing.T) {
	input := `
class CustomError
  def initialize(@msg : String)
  end
  
  def message -> String
    @msg
  end
end
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	g := codegen.New()
	output, err := g.Generate(program)
	if err != nil {
		t.Fatalf("Codegen error: %v", err)
	}

	// Check that message -> Error()
	if !strings.Contains(output, "func (c *CustomError) Error() string") {
		t.Error("Expected message to map to Error() method")
	}
}
