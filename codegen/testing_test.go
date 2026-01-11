package codegen

import (
	"strings"
	"testing"

	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/parser"
)

func TestDescribeCodegen(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "simple describe",
			input: `describe "User" do
  it "has a name" do |t|
    puts "test"
  end
end`,
			contains: []string{
				"func TestUser(t *testing.T)",
				`t.Run("has a name"`,
				"runtime.Puts",
			},
		},
		{
			name: "describe with special chars",
			input: `describe "String#to_i?" do
  it "parses valid ints" do |t|
    puts "ok"
  end
end`,
			contains: []string{
				"func TestStringToI(t *testing.T)",
				`t.Run("parses valid ints"`,
			},
		},
		{
			name: "nested describe",
			input: `describe "User" do
  describe "Validation" do
    it "requires name" do |t|
      puts "test"
    end
  end
end`,
			contains: []string{
				"func TestUser(t *testing.T)",
				`t.Run("Validation"`,
				`t.Run("requires name"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			gen := New()
			output, err := gen.Generate(program)
			if err != nil {
				t.Fatalf("codegen error: %v", err)
			}

			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("expected output to contain %q\nGot:\n%s", substr, output)
				}
			}
		})
	}
}

func TestTestCodegen(t *testing.T) {
	input := `test "math/add" do |t|
  x = 1 + 2
  puts x
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	gen := New()
	output, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	expected := []string{
		"func TestMathAdd(t *testing.T)",
		"x := (1 + 2)",
		"runtime.Puts",
	}

	for _, substr := range expected {
		if !strings.Contains(output, substr) {
			t.Errorf("expected output to contain %q\nGot:\n%s", substr, output)
		}
	}
}

func TestBeforeAfterCodegen(t *testing.T) {
	input := `describe "Database" do
  before do |t|
    puts "setup"
  end

  after do |t|
    puts "teardown"
  end

  it "works" do |t|
    puts "test"
  end
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	gen := New()
	output, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	expected := []string{
		"func TestDatabase(t *testing.T)",
		`t.Run("works"`,
		"t.Cleanup(func()",
	}

	for _, substr := range expected {
		if !strings.Contains(output, substr) {
			t.Errorf("expected output to contain %q\nGot:\n%s", substr, output)
		}
	}
}

func TestAssertCodegen(t *testing.T) {
	// Note: We can't use assert.true because 'true' is a keyword.
	// The test package provides these methods, but we need to call them
	// with names that aren't keywords. In practice, users would write:
	//   assert.equal(t, true, value)  # instead of assert.true(t, value)
	input := `describe "Assertions" do
  it "uses assert" do |t|
    assert.equal(t, 1, 1)
    assert.not_nil(t, "hello")
    require.equal(t, 2, 2)
  end
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	gen := New()
	output, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	expected := []string{
		"test.AssertInstance.Equal(t, 1, 1)",
		"test.AssertInstance.NotNil(t, \"hello\")",
		"test.RequireInstance.Equal(t, 2, 2)",
	}

	for _, substr := range expected {
		if !strings.Contains(output, substr) {
			t.Errorf("expected output to contain %q\nGot:\n%s", substr, output)
		}
	}

	// Should import the test package
	if !strings.Contains(output, "github.com/nchapman/rugby/test") {
		t.Errorf("expected output to import rugby/test\nGot:\n%s", output)
	}
}

func TestSanitizeTestName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"String#to_i?", "StringToI"},
		{"math/add", "MathAdd"},
		{"User Validation", "UserValidation"},
		{"handles special chars!", "HandlesSpecialChars"},
		{"Array[T]", "ArrayT"},
		{"==", "EqEq"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeTestName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeTestName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
