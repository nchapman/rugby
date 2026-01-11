package parser

import (
	"testing"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/lexer"
)

func TestStatementModifiers(t *testing.T) {
	tests := []struct {
		input    string
		testFunc func(*ast.Program) bool
	}{
		{
			input: "break if x == 5",
			testFunc: func(p *ast.Program) bool {
				stmt := p.Declarations[0].(*ast.BreakStmt)
				return stmt.Condition != nil && !stmt.IsUnless
			},
		},
		{
			input: "next unless valid?",
			testFunc: func(p *ast.Program) bool {
				stmt := p.Declarations[0].(*ast.NextStmt)
				return stmt.Condition != nil && stmt.IsUnless
			},
		},
		{
			input: "return if error",
			testFunc: func(p *ast.Program) bool {
				stmt := p.Declarations[0].(*ast.ReturnStmt)
				return stmt.Condition != nil && !stmt.IsUnless
			},
		},
		{
			input: "return 42 if x > 0",
			testFunc: func(p *ast.Program) bool {
				stmt := p.Declarations[0].(*ast.ReturnStmt)
				return len(stmt.Values) == 1 && stmt.Condition != nil && !stmt.IsUnless
			},
		},
		{
			input: "puts x unless y == 0",
			testFunc: func(p *ast.Program) bool {
				stmt := p.Declarations[0].(*ast.ExprStmt)
				return stmt.Condition != nil && stmt.IsUnless
			},
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("Parser errors for %q: %v", tt.input, p.Errors())
		}

		if len(program.Declarations) != 1 {
			t.Fatalf("Expected 1 statement, got %d for input %q",
				len(program.Declarations), tt.input)
		}

		if !tt.testFunc(program) {
			t.Errorf("Test failed for input %q", tt.input)
		}
	}
}

func TestUnlessStatement(t *testing.T) {
	input := `
unless x > 10
  puts "small"
else
  puts "large"
end
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	if len(program.Declarations) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Declarations))
	}

	stmt, ok := program.Declarations[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("Expected *ast.IfStmt, got %T", program.Declarations[0])
	}

	if !stmt.IsUnless {
		t.Error("Expected IsUnless to be true")
	}

	if stmt.Cond == nil {
		t.Error("Expected condition to be set")
	}

	if len(stmt.Then) == 0 {
		t.Error("Expected then block to have statements")
	}

	if len(stmt.Else) == 0 {
		t.Error("Expected else block to have statements")
	}
}

func TestUnlessStatementWithoutElse(t *testing.T) {
	input := `
unless error
  process()
end
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	stmt, ok := program.Declarations[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("Expected *ast.IfStmt, got %T", program.Declarations[0])
	}

	if !stmt.IsUnless {
		t.Error("Expected IsUnless to be true")
	}

	if len(stmt.Else) != 0 {
		t.Error("Expected no else block")
	}
}
