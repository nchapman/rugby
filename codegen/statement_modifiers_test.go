package codegen

import (
	"strings"
	"testing"

	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/parser"
)

func TestStatementModifierCodegen(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "break if",
			input: "for i in 0..10\n  break if i == 5\nend",
			contains: []string{
				"if",
				"break",
			},
		},
		{
			name:  "next unless",
			input: "for i in 0..10\n  next unless i % 2 == 0\n  puts i\nend",
			contains: []string{
				"if !",
				"continue",
			},
		},
		{
			name: "return if",
			input: `
def test(x : Int) -> String
  return "negative" if x < 0
  "positive"
end`,
			contains: []string{
				"if x < 0",
				"return \"negative\"",
			},
		},
		{
			name:  "puts unless",
			input: "puts \"hello\" unless false",
			contains: []string{
				"if !(false)",
				"runtime.Puts",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			gen := New()
			output, err := gen.Generate(program)
			if err != nil {
				t.Fatalf("Codegen error: %v", err)
			}

			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("Expected output to contain %q\nGot:\n%s", substr, output)
				}
			}
		})
	}
}

func TestUnlessStatementCodegen(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "unless without else",
			input: `
unless x == 0
  puts "not zero"
end`,
			contains: []string{
				"if !runtime.Equal(x, 0)",
				"runtime.Puts(\"not zero\")",
			},
		},
		{
			name: "unless with else",
			input: `
unless valid
  puts "invalid"
else
  puts "valid"
end`,
			contains: []string{
				"if !valid",
				"runtime.Puts(\"invalid\")",
				"} else {",
				"runtime.Puts(\"valid\")",
			},
		},
		{
			name: "unless with complex condition",
			input: `
unless x > 10 and y < 5
  puts "condition false"
end`,
			contains: []string{
				"if !((x > 10) && (y < 5))",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			gen := New()
			output, err := gen.Generate(program)
			if err != nil {
				t.Fatalf("Codegen error: %v", err)
			}

			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("Expected output to contain %q\nGot:\n%s", substr, output)
				}
			}
		})
	}
}
