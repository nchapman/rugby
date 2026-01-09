package parser

import (
	"testing"

	"rugby/ast"
	"rugby/lexer"
)

func TestAssignStatement(t *testing.T) {
	input := `def main
  x = 5
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 statement in body, got %d", len(fn.Body))
	}

	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	if assign.Name != "x" {
		t.Errorf("expected name 'x', got %q", assign.Name)
	}

	intLit, ok := assign.Value.(*ast.IntLit)
	if !ok {
		t.Fatalf("expected IntLit, got %T", assign.Value)
	}

	if intLit.Value != 5 {
		t.Errorf("expected value 5, got %d", intLit.Value)
	}
}

func TestBinaryExpression(t *testing.T) {
	tests := []struct {
		input    string
		left     int64
		operator string
		right    int64
	}{
		{"def main\n  x = 5 + 3\nend", 5, "+", 3},
		{"def main\n  x = 10 - 2\nend", 10, "-", 2},
		{"def main\n  x = 4 * 5\nend", 4, "*", 5},
		{"def main\n  x = 20 / 4\nend", 20, "/", 4},
		{"def main\n  x = 10 % 3\nend", 10, "%", 3},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		fn := program.Declarations[0].(*ast.FuncDecl)
		assign := fn.Body[0].(*ast.AssignStmt)
		binExpr, ok := assign.Value.(*ast.BinaryExpr)
		if !ok {
			t.Fatalf("expected BinaryExpr, got %T", assign.Value)
		}

		leftInt := binExpr.Left.(*ast.IntLit)
		if leftInt.Value != tt.left {
			t.Errorf("expected left %d, got %d", tt.left, leftInt.Value)
		}

		if binExpr.Op != tt.operator {
			t.Errorf("expected operator %q, got %q", tt.operator, binExpr.Op)
		}

		rightInt := binExpr.Right.(*ast.IntLit)
		if rightInt.Value != tt.right {
			t.Errorf("expected right %d, got %d", tt.right, rightInt.Value)
		}
	}
}

func TestIfStatement(t *testing.T) {
	input := `def main
  if x > 5
    puts "big"
  elsif x > 2
    puts "medium"
  else
    puts "small"
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	ifStmt, ok := fn.Body[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", fn.Body[0])
	}

	if ifStmt.Cond == nil {
		t.Fatal("expected condition, got nil")
	}

	if len(ifStmt.Then) != 1 {
		t.Errorf("expected 1 then statement, got %d", len(ifStmt.Then))
	}

	if len(ifStmt.ElseIfs) != 1 {
		t.Errorf("expected 1 elsif clause, got %d", len(ifStmt.ElseIfs))
	}

	if len(ifStmt.Else) != 1 {
		t.Errorf("expected 1 else statement, got %d", len(ifStmt.Else))
	}
}

func TestWhileStatement(t *testing.T) {
	input := `def main
  while x > 0
    x = x - 1
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	whileStmt, ok := fn.Body[0].(*ast.WhileStmt)
	if !ok {
		t.Fatalf("expected WhileStmt, got %T", fn.Body[0])
	}

	if whileStmt.Cond == nil {
		t.Fatal("expected condition, got nil")
	}

	if len(whileStmt.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(whileStmt.Body))
	}
}

func TestFunctionCall(t *testing.T) {
	input := `def main
  puts "hello"
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	exprStmt, ok := fn.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", fn.Body[0])
	}

	call, ok := exprStmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
	}

	if call.Func != "puts" {
		t.Errorf("expected func 'puts', got %q", call.Func)
	}

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}

	strLit, ok := call.Args[0].(*ast.StringLit)
	if !ok {
		t.Fatalf("expected StringLit, got %T", call.Args[0])
	}

	if strLit.Value != "hello" {
		t.Errorf("expected 'hello', got %q", strLit.Value)
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"def main\n  x = 1 + 2 * 3\nend", "((1) + ((2) * (3)))"},
		{"def main\n  x = (1 + 2) * 3\nend", "(((1) + (2)) * (3))"},
		{"def main\n  x = 1 < 2 and 3 > 4\nend", "(((1) < (2)) && ((3) > (4)))"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		fn := program.Declarations[0].(*ast.FuncDecl)
		assign := fn.Body[0].(*ast.AssignStmt)
		actual := exprString(assign.Value)

		if actual != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, actual)
		}
	}
}

func TestImport(t *testing.T) {
	input := `import fmt
import net/http

def main
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(program.Imports))
	}

	if program.Imports[0].Path != "fmt" {
		t.Errorf("expected 'fmt', got %q", program.Imports[0].Path)
	}

	if program.Imports[1].Path != "net/http" {
		t.Errorf("expected 'net/http', got %q", program.Imports[1].Path)
	}
}

func TestFunctionParams(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"def foo\nend", nil},
		{"def foo()\nend", nil},
		{"def foo(a)\nend", []string{"a"}},
		{"def foo(a, b)\nend", []string{"a", "b"}},
		{"def foo(a, b, c)\nend", []string{"a", "b", "c"}},
		{"def foo(a, b,)\nend", []string{"a", "b"}}, // trailing comma allowed
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Declarations) != 1 {
			t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
		}

		fn, ok := program.Declarations[0].(*ast.FuncDecl)
		if !ok {
			t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
		}

		if len(fn.Params) != len(tt.expected) {
			t.Errorf("input %q: expected %d params, got %d", tt.input, len(tt.expected), len(fn.Params))
			continue
		}

		for i, name := range tt.expected {
			if fn.Params[i].Name != name {
				t.Errorf("input %q: param %d expected %q, got %q", tt.input, i, name, fn.Params[i].Name)
			}
		}
	}
}

func TestDuplicateParams(t *testing.T) {
	input := `def foo(a, a)
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected error for duplicate parameter name, got none")
	}
}

func TestMissingReturnType(t *testing.T) {
	input := `def foo() ->
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected error for missing type after ->, got none")
	}
}

func TestEmptyReturnTypeParens(t *testing.T) {
	input := `def foo() -> ()
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected error for empty return type list, got none")
	}
}

func TestReturnType(t *testing.T) {
	tests := []struct {
		input       string
		returnTypes []string
	}{
		{"def foo\nend", nil},
		{"def foo()\nend", nil},
		{"def foo() -> Int\nend", []string{"Int"}},
		{"def foo(a, b) -> String\nend", []string{"String"}},
		{"def foo(x) -> Bool\nend", []string{"Bool"}},
		{"def foo -> Int\nend", []string{"Int"}}, // no parens
		{"def foo() -> (Int, Bool)\nend", []string{"Int", "Bool"}},
		{"def foo() -> (String, Int, Bool)\nend", []string{"String", "Int", "Bool"}},
		{"def foo() -> (Int, Bool,)\nend", []string{"Int", "Bool"}}, // trailing comma
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		fn := program.Declarations[0].(*ast.FuncDecl)

		if len(fn.ReturnTypes) != len(tt.returnTypes) {
			t.Errorf("input %q: expected %d return types, got %d",
				tt.input, len(tt.returnTypes), len(fn.ReturnTypes))
			continue
		}

		for i, rt := range tt.returnTypes {
			if fn.ReturnTypes[i] != rt {
				t.Errorf("input %q: return type %d expected %q, got %q",
					tt.input, i, rt, fn.ReturnTypes[i])
			}
		}
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %s", msg)
	}
	t.FailNow()
}

func exprString(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.IntLit:
		return "(" + string(rune('0'+e.Value)) + ")"
	case *ast.BinaryExpr:
		return "(" + exprString(e.Left) + " " + e.Op + " " + exprString(e.Right) + ")"
	default:
		return "?"
	}
}
