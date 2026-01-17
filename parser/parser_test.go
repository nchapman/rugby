package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/lexer"
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

		fn, ok := program.Declarations[0].(*ast.FuncDecl)
		if !ok {
			t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
		}
		assign, ok := fn.Body[0].(*ast.AssignStmt)
		if !ok {
			t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
		}
		binExpr, ok := assign.Value.(*ast.BinaryExpr)
		if !ok {
			t.Fatalf("expected BinaryExpr, got %T", assign.Value)
		}

		leftInt, ok := binExpr.Left.(*ast.IntLit)
		if !ok {
			t.Fatalf("expected IntLit, got %T", binExpr.Left)
		}
		if leftInt.Value != tt.left {
			t.Errorf("expected left %d, got %d", tt.left, leftInt.Value)
		}

		if binExpr.Op != tt.operator {
			t.Errorf("expected operator %q, got %q", tt.operator, binExpr.Op)
		}

		rightInt, ok := binExpr.Right.(*ast.IntLit)
		if !ok {
			t.Fatalf("expected IntLit, got %T", binExpr.Right)
		}
		if rightInt.Value != tt.right {
			t.Errorf("expected right %d, got %d", tt.right, rightInt.Value)
		}
	}
}

func TestIfStatement(t *testing.T) {
	input := `def main
  if x > 5
    puts("big")
  elsif x > 2
    puts("medium")
  else
    puts("small")
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
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

func TestIfLetStatement(t *testing.T) {
	input := `def main
  if let user = find_user(5)
    puts(user.name)
  else
    puts("not found")
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	ifStmt, ok := fn.Body[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", fn.Body[0])
	}

	// Check that it's using the assignment pattern (if let)
	if ifStmt.AssignName != "user" {
		t.Errorf("expected AssignName 'user', got %q", ifStmt.AssignName)
	}
	if ifStmt.AssignExpr == nil {
		t.Fatal("expected AssignExpr, got nil")
	}
	// The condition should be nil when using assignment pattern
	if ifStmt.Cond != nil {
		t.Errorf("expected Cond to be nil for 'if let', got %T", ifStmt.Cond)
	}

	if len(ifStmt.Then) != 1 {
		t.Errorf("expected 1 then statement, got %d", len(ifStmt.Then))
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

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
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

func TestUntilStatement(t *testing.T) {
	input := `def main
  until queue.empty?
    process(queue.pop)
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	untilStmt, ok := fn.Body[0].(*ast.UntilStmt)
	if !ok {
		t.Fatalf("expected UntilStmt, got %T", fn.Body[0])
	}

	if untilStmt.Cond == nil {
		t.Fatal("expected condition, got nil")
	}

	if len(untilStmt.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(untilStmt.Body))
	}
}

func TestPostfixWhile(t *testing.T) {
	input := `def main
  puts x while x > 0
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	// Postfix while should be parsed as a WhileStmt
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

	// The body should contain the original expression statement
	_, ok = whileStmt.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt in body, got %T", whileStmt.Body[0])
	}
}

func TestPostfixUntil(t *testing.T) {
	input := `def main
  process(item) until done
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	// Postfix until should be parsed as an UntilStmt
	untilStmt, ok := fn.Body[0].(*ast.UntilStmt)
	if !ok {
		t.Fatalf("expected UntilStmt, got %T", fn.Body[0])
	}

	if untilStmt.Cond == nil {
		t.Fatal("expected condition, got nil")
	}

	if len(untilStmt.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(untilStmt.Body))
	}

	// The body should contain the original expression statement
	_, ok = untilStmt.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt in body, got %T", untilStmt.Body[0])
	}
}

func TestFunctionCall(t *testing.T) {
	input := `def main
  puts("hello")
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	exprStmt, ok := fn.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", fn.Body[0])
	}

	call, ok := exprStmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
	}

	ident, ok := call.Func.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident for func, got %T", call.Func)
	}
	if ident.Name != "puts" {
		t.Errorf("expected func 'puts', got %q", ident.Name)
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

		fn, ok := program.Declarations[0].(*ast.FuncDecl)
		if !ok {
			t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
		}
		assign, ok := fn.Body[0].(*ast.AssignStmt)
		if !ok {
			t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
		}
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
		{"def foo(a: any)\nend", []string{"a"}},
		{"def foo(a: any, b: any)\nend", []string{"a", "b"}},
		{"def foo(a: any, b: any, c: any)\nend", []string{"a", "b", "c"}},
		{"def foo(a: any, b: any,)\nend", []string{"a", "b"}}, // trailing comma allowed
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
	input := `def foo(a: any, a: any)
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
	input := `def foo(): ()
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected error for empty return type list, got none")
	}
}

func TestMultipleReturnValues(t *testing.T) {
	input := `def foo()
  return 1, true, "hello"
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	retStmt, ok := fn.Body[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("expected ReturnStmt, got %T", fn.Body[0])
	}

	if len(retStmt.Values) != 3 {
		t.Fatalf("expected 3 return values, got %d", len(retStmt.Values))
	}

	// Check first value is int
	if _, ok := retStmt.Values[0].(*ast.IntLit); !ok {
		t.Errorf("expected IntLit, got %T", retStmt.Values[0])
	}

	// Check second value is bool
	if _, ok := retStmt.Values[1].(*ast.BoolLit); !ok {
		t.Errorf("expected BoolLit, got %T", retStmt.Values[1])
	}

	// Check third value is string
	if _, ok := retStmt.Values[2].(*ast.StringLit); !ok {
		t.Errorf("expected StringLit, got %T", retStmt.Values[2])
	}
}

func TestReturnType(t *testing.T) {
	tests := []struct {
		input       string
		returnTypes []string
	}{
		{"def foo\nend", nil},
		{"def foo()\nend", nil},
		{"def foo(): Int\nend", []string{"Int"}},
		{"def foo(a: any, b: any): String\nend", []string{"String"}},
		{"def foo(x: any): Bool\nend", []string{"Bool"}},
		{"def foo: Int\nend", []string{"Int"}}, // no parens
		{"def foo(): (Int, Bool)\nend", []string{"Int", "Bool"}},
		{"def foo(): (String, Int, Bool)\nend", []string{"String", "Int", "Bool"}},
		{"def foo(): (Int, Bool,)\nend", []string{"Int", "Bool"}}, // trailing comma
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		fn, ok := program.Declarations[0].(*ast.FuncDecl)
		if !ok {
			t.Fatalf("input %q: expected FuncDecl, got %T", tt.input, program.Declarations[0])
		}

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

func TestImportAlias(t *testing.T) {
	input := `import encoding/json as json
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

	// First import has alias
	if program.Imports[0].Path != "encoding/json" {
		t.Errorf("expected path 'encoding/json', got %q", program.Imports[0].Path)
	}
	if program.Imports[0].Alias != "json" {
		t.Errorf("expected alias 'json', got %q", program.Imports[0].Alias)
	}

	// Second import has no alias
	if program.Imports[1].Path != "net/http" {
		t.Errorf("expected path 'net/http', got %q", program.Imports[1].Path)
	}
	if program.Imports[1].Alias != "" {
		t.Errorf("expected no alias, got %q", program.Imports[1].Alias)
	}
}

func TestSelectorExpression(t *testing.T) {
	input := `def main
  x = resp.Body
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	sel, ok := assign.Value.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr, got %T", assign.Value)
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident for X, got %T", sel.X)
	}
	if ident.Name != "resp" {
		t.Errorf("expected X 'resp', got %q", ident.Name)
	}
	if sel.Sel != "Body" {
		t.Errorf("expected Sel 'Body', got %q", sel.Sel)
	}
}

func TestChainedSelector(t *testing.T) {
	input := `def main
  x = resp.Body.Close
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	// resp.Body.Close
	outer, ok := assign.Value.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr, got %T", assign.Value)
	}
	if outer.Sel != "Close" {
		t.Errorf("expected outer Sel 'Close', got %q", outer.Sel)
	}

	inner, ok := outer.X.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected inner SelectorExpr, got %T", outer.X)
	}
	if inner.Sel != "Body" {
		t.Errorf("expected inner Sel 'Body', got %q", inner.Sel)
	}
}

func TestSelectorCall(t *testing.T) {
	input := `def main
  http.Get("http://example.com")
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	exprStmt, ok := fn.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", fn.Body[0])
	}

	call, ok := exprStmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
	}

	sel, ok := call.Func.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr as Func, got %T", call.Func)
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok || ident.Name != "http" {
		t.Errorf("expected 'http', got %v", sel.X)
	}
	if sel.Sel != "Get" {
		t.Errorf("expected 'Get', got %q", sel.Sel)
	}

	if len(call.Args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(call.Args))
	}
}

func TestDeferStatement(t *testing.T) {
	input := `def main
  defer resp.Body.Close
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	deferStmt, ok := fn.Body[0].(*ast.DeferStmt)
	if !ok {
		t.Fatalf("expected DeferStmt, got %T", fn.Body[0])
	}

	if deferStmt.Call == nil {
		t.Fatal("expected Call in DeferStmt, got nil")
	}

	// The call should have a SelectorExpr as its function
	sel, ok := deferStmt.Call.Func.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr in call, got %T", deferStmt.Call.Func)
	}

	if sel.Sel != "Close" {
		t.Errorf("expected 'Close', got %q", sel.Sel)
	}
}

func TestDeferWithParens(t *testing.T) {
	input := `def main
  defer file.Close()
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	deferStmt, ok := fn.Body[0].(*ast.DeferStmt)
	if !ok {
		t.Fatalf("expected DeferStmt, got %T", fn.Body[0])
	}

	if deferStmt.Call == nil {
		t.Fatal("expected Call in DeferStmt")
	}
}

func TestPanicStatement(t *testing.T) {
	input := `def main
  panic "something went wrong"
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	panicStmt, ok := fn.Body[0].(*ast.PanicStmt)
	if !ok {
		t.Fatalf("expected PanicStmt, got %T", fn.Body[0])
	}

	strLit, ok := panicStmt.Message.(*ast.StringLit)
	if !ok {
		t.Fatalf("expected StringLit message, got %T", panicStmt.Message)
	}

	if strLit.Value != "something went wrong" {
		t.Errorf("expected message 'something went wrong', got %q", strLit.Value)
	}
}

func TestBangExpression(t *testing.T) {
	input := `def main
  data = read_file("test.txt")!
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	assignStmt, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	bangExpr, ok := assignStmt.Value.(*ast.BangExpr)
	if !ok {
		t.Fatalf("expected BangExpr as value, got %T", assignStmt.Value)
	}

	callExpr, ok := bangExpr.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr inside BangExpr, got %T", bangExpr.Expr)
	}

	ident, ok := callExpr.Func.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident as func, got %T", callExpr.Func)
	}

	if ident.Name != "read_file" {
		t.Errorf("expected function name 'read_file', got %q", ident.Name)
	}
}

func TestBangExpressionOnNonCall(t *testing.T) {
	// Test that ! on a literal (not a call/selector/ident) produces an error
	input := `def main
  y = 42!
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	// Should have an error because ! can only follow a call, selector, or identifier
	if len(p.Errors()) == 0 {
		t.Fatal("expected parse error for '!' on non-call expression (literal)")
	}
}

func TestBangExpressionOnSelector(t *testing.T) {
	// Test that ! works on selector expressions: resp.json! → resp.json()!
	input := `def main
  data = resp.json!
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	assignStmt, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	bangExpr, ok := assignStmt.Value.(*ast.BangExpr)
	if !ok {
		t.Fatalf("expected BangExpr as value, got %T", assignStmt.Value)
	}

	// The selector should have been converted to a CallExpr
	callExpr, ok := bangExpr.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr inside BangExpr, got %T", bangExpr.Expr)
	}

	// The function should be a SelectorExpr
	selExpr, ok := callExpr.Func.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr as func, got %T", callExpr.Func)
	}

	if selExpr.Sel != "json" {
		t.Errorf("expected selector 'json', got %q", selExpr.Sel)
	}
}

func TestBangExpressionOnIdent(t *testing.T) {
	// Test that ! works on identifiers: foo! → foo()!
	input := `def main
  result = some_func!
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	assignStmt, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	bangExpr, ok := assignStmt.Value.(*ast.BangExpr)
	if !ok {
		t.Fatalf("expected BangExpr as value, got %T", assignStmt.Value)
	}

	// The identifier should have been converted to a CallExpr
	callExpr, ok := bangExpr.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr inside BangExpr, got %T", bangExpr.Expr)
	}

	// The function should be an Ident
	ident, ok := callExpr.Func.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident as func, got %T", callExpr.Func)
	}

	if ident.Name != "some_func" {
		t.Errorf("expected function name 'some_func', got %q", ident.Name)
	}
}

func TestArrayLiteral(t *testing.T) {
	input := `def main
  x = [1, 2, 3]
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	arr, ok := assign.Value.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected ArrayLit, got %T", assign.Value)
	}

	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}

	for i, expected := range []int64{1, 2, 3} {
		intLit, ok := arr.Elements[i].(*ast.IntLit)
		if !ok {
			t.Fatalf("element %d: expected IntLit, got %T", i, arr.Elements[i])
		}
		if intLit.Value != expected {
			t.Errorf("element %d: expected %d, got %d", i, expected, intLit.Value)
		}
	}
}

func TestEmptyArrayLiteral(t *testing.T) {
	input := `def main
  x = []
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	arr, ok := assign.Value.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected ArrayLit, got %T", assign.Value)
	}

	if len(arr.Elements) != 0 {
		t.Errorf("expected 0 elements, got %d", len(arr.Elements))
	}
}

func TestArrayLiteralTrailingComma(t *testing.T) {
	input := `def main
  x = [1, 2,]
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	arr, ok := assign.Value.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected ArrayLit, got %T", assign.Value)
	}

	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr.Elements))
	}
}

func TestArrayLiteralMixedTypes(t *testing.T) {
	input := `def main
  x = [1, "hello", true]
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	arr, ok := assign.Value.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected ArrayLit, got %T", assign.Value)
	}

	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}

	if _, ok := arr.Elements[0].(*ast.IntLit); !ok {
		t.Errorf("element 0: expected IntLit, got %T", arr.Elements[0])
	}
	if _, ok := arr.Elements[1].(*ast.StringLit); !ok {
		t.Errorf("element 1: expected StringLit, got %T", arr.Elements[1])
	}
	if _, ok := arr.Elements[2].(*ast.BoolLit); !ok {
		t.Errorf("element 2: expected BoolLit, got %T", arr.Elements[2])
	}
}

func TestNestedArrayLiteral(t *testing.T) {
	input := `def main
  x = [[1, 2], [3, 4]]
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	arr, ok := assign.Value.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected ArrayLit, got %T", assign.Value)
	}

	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
	}

	// Check first inner array
	inner1, ok := arr.Elements[0].(*ast.ArrayLit)
	if !ok {
		t.Fatalf("element 0: expected ArrayLit, got %T", arr.Elements[0])
	}
	if len(inner1.Elements) != 2 {
		t.Errorf("inner array 0: expected 2 elements, got %d", len(inner1.Elements))
	}

	// Check second inner array
	inner2, ok := arr.Elements[1].(*ast.ArrayLit)
	if !ok {
		t.Fatalf("element 1: expected ArrayLit, got %T", arr.Elements[1])
	}
	if len(inner2.Elements) != 2 {
		t.Errorf("inner array 1: expected 2 elements, got %d", len(inner2.Elements))
	}
}

func TestMultilineArrayLiteral(t *testing.T) {
	input := `def main
  x = [
    1,
    2,
    3,
  ]
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 statement in body, got %d", len(fn.Body))
	}

	assign := fn.Body[0].(*ast.AssignStmt)
	arr, ok := assign.Value.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected ArrayLit, got %T", assign.Value)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Elements))
	}
}

func TestMultilineMapLiteral(t *testing.T) {
	input := `def main
  x = {
    "a" => 1,
    "b" => 2,
    "c" => 3,
  }
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	mapLit, ok := assign.Value.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", assign.Value)
	}
	if len(mapLit.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(mapLit.Entries))
	}
}

func TestMultilineMapWithSymbolKeys(t *testing.T) {
	input := `def main
  x = {
    name: "Alice",
    age: 30,
  }
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	mapLit, ok := assign.Value.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", assign.Value)
	}
	if len(mapLit.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(mapLit.Entries))
	}
}

func TestMultilineMapWithImplicitValues(t *testing.T) {
	input := `def main
  name = "Alice"
  age = 30
  x = {
    name:,
    age:,
  }
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[2].(*ast.AssignStmt)
	mapLit, ok := assign.Value.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", assign.Value)
	}
	if len(mapLit.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(mapLit.Entries))
	}

	// Check that key and value are both present
	entry := mapLit.Entries[0]
	keyStr, ok := entry.Key.(*ast.StringLit)
	if !ok {
		t.Fatalf("expected StringLit key, got %T", entry.Key)
	}
	if keyStr.Value != "name" {
		t.Errorf("expected key 'name', got %q", keyStr.Value)
	}
	valueIdent, ok := entry.Value.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident value (implicit), got %T", entry.Value)
	}
	if valueIdent.Name != "name" {
		t.Errorf("expected value 'name', got %q", valueIdent.Name)
	}
}

func TestNestedMultilineArrays(t *testing.T) {
	input := `def main
  x = [
    [1, 2],
    [3, 4],
  ]
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	arr, ok := assign.Value.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected ArrayLit, got %T", assign.Value)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr.Elements))
	}

	inner1, ok := arr.Elements[0].(*ast.ArrayLit)
	if !ok {
		t.Fatalf("element 0: expected ArrayLit, got %T", arr.Elements[0])
	}
	if len(inner1.Elements) != 2 {
		t.Errorf("inner array 0: expected 2 elements, got %d", len(inner1.Elements))
	}
}

func TestArrayIndexing(t *testing.T) {
	input := `def main
  x = arr[0]
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	idx, ok := assign.Value.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %T", assign.Value)
	}

	ident, ok := idx.Left.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident for Left, got %T", idx.Left)
	}
	if ident.Name != "arr" {
		t.Errorf("expected Left 'arr', got %q", ident.Name)
	}

	intLit, ok := idx.Index.(*ast.IntLit)
	if !ok {
		t.Fatalf("expected IntLit for Index, got %T", idx.Index)
	}
	if intLit.Value != 0 {
		t.Errorf("expected Index 0, got %d", intLit.Value)
	}
}

func TestArrayIndexingWithExpression(t *testing.T) {
	input := `def main
  x = arr[i + 1]
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	idx, ok := assign.Value.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %T", assign.Value)
	}

	_, ok = idx.Index.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr for Index, got %T", idx.Index)
	}
}

func TestChainedArrayIndexing(t *testing.T) {
	input := `def main
  x = matrix[0][1]
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	// matrix[0][1] should be IndexExpr(IndexExpr(matrix, 0), 1)
	outer, ok := assign.Value.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected outer IndexExpr, got %T", assign.Value)
	}

	inner, ok := outer.Left.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected inner IndexExpr, got %T", outer.Left)
	}

	ident, ok := inner.Left.(*ast.Ident)
	if !ok || ident.Name != "matrix" {
		t.Errorf("expected 'matrix', got %v", inner.Left)
	}
}

func TestMapLiteral(t *testing.T) {
	input := `def main
  x = {"a" => 1, "b" => 2}
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	mapLit, ok := assign.Value.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", assign.Value)
	}

	if len(mapLit.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(mapLit.Entries))
	}

	// Check first entry
	key1, ok := mapLit.Entries[0].Key.(*ast.StringLit)
	if !ok || key1.Value != "a" {
		t.Errorf("entry 0 key: expected 'a', got %v", mapLit.Entries[0].Key)
	}
	val1, ok := mapLit.Entries[0].Value.(*ast.IntLit)
	if !ok || val1.Value != 1 {
		t.Errorf("entry 0 value: expected 1, got %v", mapLit.Entries[0].Value)
	}

	// Check second entry
	key2, ok := mapLit.Entries[1].Key.(*ast.StringLit)
	if !ok || key2.Value != "b" {
		t.Errorf("entry 1 key: expected 'b', got %v", mapLit.Entries[1].Key)
	}
	val2, ok := mapLit.Entries[1].Value.(*ast.IntLit)
	if !ok || val2.Value != 2 {
		t.Errorf("entry 1 value: expected 2, got %v", mapLit.Entries[1].Value)
	}
}

func TestEmptyMapLiteral(t *testing.T) {
	input := `def main
  x = {}
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	mapLit, ok := assign.Value.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", assign.Value)
	}

	if len(mapLit.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(mapLit.Entries))
	}
}

func TestMapLiteralTrailingComma(t *testing.T) {
	input := `def main
  x = {"a" => 1,}
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	mapLit, ok := assign.Value.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", assign.Value)
	}

	if len(mapLit.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(mapLit.Entries))
	}
}

func TestMapLiteralJSONStyle(t *testing.T) {
	// JSON-style syntax with colon instead of hashrocket
	input := `def main
  x = {"name": "Alice", "age": 30}
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	mapLit, ok := assign.Value.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", assign.Value)
	}

	if len(mapLit.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(mapLit.Entries))
	}

	// Check first entry: "name": "Alice"
	key1, ok := mapLit.Entries[0].Key.(*ast.StringLit)
	if !ok || key1.Value != "name" {
		t.Errorf("entry 0 key: expected 'name', got %v", mapLit.Entries[0].Key)
	}
	val1, ok := mapLit.Entries[0].Value.(*ast.StringLit)
	if !ok || val1.Value != "Alice" {
		t.Errorf("entry 0 value: expected 'Alice', got %v", mapLit.Entries[0].Value)
	}

	// Check second entry: "age": 30
	key2, ok := mapLit.Entries[1].Key.(*ast.StringLit)
	if !ok || key2.Value != "age" {
		t.Errorf("entry 1 key: expected 'age', got %v", mapLit.Entries[1].Key)
	}
	val2, ok := mapLit.Entries[1].Value.(*ast.IntLit)
	if !ok || val2.Value != 30 {
		t.Errorf("entry 1 value: expected 30, got %v", mapLit.Entries[1].Value)
	}
}

func TestEachBlock(t *testing.T) {
	input := `def main
  arr.each -> { |x|
    puts(x)
  }
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

	sel, ok := call.Func.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr, got %T", call.Func)
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok || ident.Name != "arr" {
		t.Errorf("expected receiver 'arr', got %v", sel.X)
	}

	if sel.Sel != "each" {
		t.Errorf("expected selector 'each', got %q", sel.Sel)
	}

	// Arrow lambda puts the lambda in Args, not Block
	if len(call.Args) != 1 {
		t.Fatal("expected 1 lambda argument, got", len(call.Args))
	}

	lambda, ok := call.Args[0].(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("expected LambdaExpr, got %T", call.Args[0])
	}

	if len(lambda.Params) != 1 || lambda.Params[0].Name != "x" {
		t.Errorf("expected lambda params [x], got %v", lambda.Params)
	}

	if len(lambda.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(lambda.Body))
	}
}

func TestEachWithIndexBlock(t *testing.T) {
	input := `def main
  arr.each_with_index -> { |v, i|
    puts(v)
  }
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

	sel, ok := call.Func.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr, got %T", call.Func)
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok || ident.Name != "arr" {
		t.Errorf("expected receiver 'arr', got %v", sel.X)
	}

	if sel.Sel != "each_with_index" {
		t.Errorf("expected selector 'each_with_index', got %q", sel.Sel)
	}

	// Arrow lambda puts the lambda in Args, not Block
	if len(call.Args) != 1 {
		t.Fatal("expected 1 lambda argument, got", len(call.Args))
	}

	lambda, ok := call.Args[0].(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("expected LambdaExpr, got %T", call.Args[0])
	}

	if len(lambda.Params) != 2 {
		t.Fatalf("expected 2 lambda params, got %d", len(lambda.Params))
	}

	if lambda.Params[0].Name != "v" {
		t.Errorf("expected first param 'v', got %q", lambda.Params[0].Name)
	}

	if lambda.Params[1].Name != "i" {
		t.Errorf("expected second param 'i', got %q", lambda.Params[1].Name)
	}

	if len(lambda.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(lambda.Body))
	}
}

func TestMapBlock(t *testing.T) {
	input := `def main
  result = arr.map -> { |x|
    x * 2
  }
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)

	assignStmt, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	call, ok := assignStmt.Value.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", assignStmt.Value)
	}

	sel, ok := call.Func.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr, got %T", call.Func)
	}

	if sel.Sel != "map" {
		t.Errorf("expected selector 'map', got %q", sel.Sel)
	}

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg (lambda), got %d", len(call.Args))
	}

	lambda, ok := call.Args[0].(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("expected LambdaExpr, got %T", call.Args[0])
	}

	if len(lambda.Params) != 1 || lambda.Params[0].Name != "x" {
		t.Errorf("expected lambda params [x], got %v", lambda.Params)
	}
}

func TestBlockWithNoParams(t *testing.T) {
	input := `def main
  items.each -> {
    puts("hello")
  }
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

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg (lambda), got %d", len(call.Args))
	}

	lambda, ok := call.Args[0].(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("expected LambdaExpr, got %T", call.Args[0])
	}

	if len(lambda.Params) != 0 {
		t.Errorf("expected 0 lambda params, got %v", lambda.Params)
	}
}

func TestGenericBlockOnAnyMethod(t *testing.T) {
	input := `def main
  items.custom_method -> { |x, y, z|
    puts(x)
  }
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

	sel, ok := call.Func.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr, got %T", call.Func)
	}

	if sel.Sel != "custom_method" {
		t.Errorf("expected selector 'custom_method', got %q", sel.Sel)
	}

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg (lambda), got %d", len(call.Args))
	}

	lambda, ok := call.Args[0].(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("expected LambdaExpr, got %T", call.Args[0])
	}

	if len(lambda.Params) != 3 {
		t.Fatalf("expected 3 lambda params, got %d", len(lambda.Params))
	}

	expectedParams := []string{"x", "y", "z"}
	for i, expected := range expectedParams {
		if lambda.Params[i].Name != expected {
			t.Errorf("expected param %d to be %q, got %q", i, expected, lambda.Params[i].Name)
		}
	}
}

func TestDuplicateBlockParameter(t *testing.T) {
	input := `def main
  items.each -> { |x, x|
    puts(x)
  }
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Fatal("expected parser error for duplicate block parameter")
	}

	errMsg := p.Errors()[0]
	if !strings.Contains(errMsg, "duplicate lambda parameter") {
		t.Errorf("expected error about duplicate lambda parameter, got: %s", errMsg)
	}
}

func TestNestedBlocks(t *testing.T) {
	input := `def main
  matrix.each -> { |row|
    row.each -> { |x|
      puts(x)
    }
  }
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

	outerCall, ok := exprStmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
	}

	if len(outerCall.Args) != 1 {
		t.Fatalf("expected 1 arg (lambda), got %d", len(outerCall.Args))
	}

	outerLambda, ok := outerCall.Args[0].(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("expected LambdaExpr, got %T", outerCall.Args[0])
	}

	if len(outerLambda.Params) != 1 || outerLambda.Params[0].Name != "row" {
		t.Errorf("expected outer lambda params [row], got %v", outerLambda.Params)
	}

	// Inner block should be in the body
	if len(outerLambda.Body) != 1 {
		t.Fatalf("expected 1 statement in outer lambda, got %d", len(outerLambda.Body))
	}

	innerExprStmt, ok := outerLambda.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected inner ExprStmt, got %T", outerLambda.Body[0])
	}

	innerCall, ok := innerExprStmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected inner CallExpr, got %T", innerExprStmt.Expr)
	}

	if len(innerCall.Args) != 1 {
		t.Fatalf("expected 1 arg (lambda), got %d", len(innerCall.Args))
	}

	innerLambda, ok := innerCall.Args[0].(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("expected inner LambdaExpr, got %T", innerCall.Args[0])
	}

	if len(innerLambda.Params) != 1 || innerLambda.Params[0].Name != "x" {
		t.Errorf("expected inner lambda params [x], got %v", innerLambda.Params)
	}
}

func TestEmptyClass(t *testing.T) {
	input := `class User
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	cls, ok := program.Declarations[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", program.Declarations[0])
	}

	if cls.Name != "User" {
		t.Errorf("expected class name 'User', got %q", cls.Name)
	}

	if len(cls.Methods) != 0 {
		t.Errorf("expected 0 methods, got %d", len(cls.Methods))
	}
}

func TestClassWithMethod(t *testing.T) {
	input := `class User
  def greet
    puts("hello")
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls := program.Declarations[0].(*ast.ClassDecl)

	if cls.Name != "User" {
		t.Errorf("expected class name 'User', got %q", cls.Name)
	}

	if len(cls.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(cls.Methods))
	}

	method := cls.Methods[0]
	if method.Name != "greet" {
		t.Errorf("expected method name 'greet', got %q", method.Name)
	}

	if len(method.Body) != 1 {
		t.Errorf("expected 1 statement in method body, got %d", len(method.Body))
	}
}

func TestClassWithMethodParams(t *testing.T) {
	input := `class Calculator
  def add(a: any, b: any): Int
    a + b
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls := program.Declarations[0].(*ast.ClassDecl)
	method := cls.Methods[0]

	if method.Name != "add" {
		t.Errorf("expected method name 'add', got %q", method.Name)
	}

	if len(method.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(method.Params))
	}

	if method.Params[0].Name != "a" {
		t.Errorf("expected first param 'a', got %q", method.Params[0].Name)
	}

	if method.Params[1].Name != "b" {
		t.Errorf("expected second param 'b', got %q", method.Params[1].Name)
	}

	if len(method.ReturnTypes) != 1 || method.ReturnTypes[0] != "Int" {
		t.Errorf("expected return type 'Int', got %v", method.ReturnTypes)
	}
}

func TestClassWithMultipleMethods(t *testing.T) {
	input := `class Counter
  def inc
    puts("increment")
  end

  def dec
    puts("decrement")
  end

  def value: Int
    0
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls := program.Declarations[0].(*ast.ClassDecl)

	if len(cls.Methods) != 3 {
		t.Fatalf("expected 3 methods, got %d", len(cls.Methods))
	}

	expectedNames := []string{"inc", "dec", "value"}
	for i, name := range expectedNames {
		if cls.Methods[i].Name != name {
			t.Errorf("expected method %d name %q, got %q", i, name, cls.Methods[i].Name)
		}
	}
}

func TestClassWithEmbed(t *testing.T) {
	input := `class Service < Logger
  def run
    puts("running")
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls := program.Declarations[0].(*ast.ClassDecl)

	if cls.Name != "Service" {
		t.Errorf("expected class name 'Service', got %q", cls.Name)
	}

	if len(cls.Embeds) != 1 || cls.Embeds[0] != "Logger" {
		t.Errorf("expected embeds ['Logger'], got %v", cls.Embeds)
	}
}

func TestClassWithMultipleEmbeds(t *testing.T) {
	input := `class Service < Logger, Authenticator
  def run
    puts("running")
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls := program.Declarations[0].(*ast.ClassDecl)

	if cls.Name != "Service" {
		t.Errorf("expected class name 'Service', got %q", cls.Name)
	}

	expectedEmbeds := []string{"Logger", "Authenticator"}
	if len(cls.Embeds) != len(expectedEmbeds) {
		t.Fatalf("expected %d embeds, got %d", len(expectedEmbeds), len(cls.Embeds))
	}

	for i, expected := range expectedEmbeds {
		if cls.Embeds[i] != expected {
			t.Errorf("embed %d: expected %q, got %q", i, expected, cls.Embeds[i])
		}
	}
}

func TestClassWithThreeEmbeds(t *testing.T) {
	input := `class App < Logger, Authenticator, Database
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls := program.Declarations[0].(*ast.ClassDecl)

	if len(cls.Embeds) != 3 {
		t.Fatalf("expected 3 embeds, got %d", len(cls.Embeds))
	}

	if cls.Embeds[0] != "Logger" || cls.Embeds[1] != "Authenticator" || cls.Embeds[2] != "Database" {
		t.Errorf("unexpected embeds: %v", cls.Embeds)
	}
}

func TestClassEmbedMissingTypeAfterComma(t *testing.T) {
	input := `class Service < Logger,
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected error for missing type after comma, got none")
	}
}

func TestClassMissingName(t *testing.T) {
	input := `class end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected error for missing class name, got none")
	}
}

func TestClassDuplicateMethodParams(t *testing.T) {
	input := `class User
  def add(a: any, a: any)
  end
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected error for duplicate parameter name, got none")
	}
}

func TestClassWithInitialize(t *testing.T) {
	input := `class User
  def initialize(name: any, age: any)
    @name = name
    @age = age
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls := program.Declarations[0].(*ast.ClassDecl)

	if cls.Name != "User" {
		t.Errorf("expected class name 'User', got %q", cls.Name)
	}

	if len(cls.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(cls.Fields))
	}

	if cls.Fields[0].Name != "name" {
		t.Errorf("expected first field 'name', got %q", cls.Fields[0].Name)
	}

	if cls.Fields[1].Name != "age" {
		t.Errorf("expected second field 'age', got %q", cls.Fields[1].Name)
	}
}

func TestInstanceVariableAssignment(t *testing.T) {
	input := `class User
  def initialize(name: any)
    @name = name
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls := program.Declarations[0].(*ast.ClassDecl)
	method := cls.Methods[0]

	if len(method.Body) != 1 {
		t.Fatalf("expected 1 statement in body, got %d", len(method.Body))
	}

	assign, ok := method.Body[0].(*ast.InstanceVarAssign)
	if !ok {
		t.Fatalf("expected InstanceVarAssign, got %T", method.Body[0])
	}

	if assign.Name != "name" {
		t.Errorf("expected var name 'name', got %q", assign.Name)
	}
}

func TestInstanceVariableReference(t *testing.T) {
	input := `class User
  def initialize(name: any)
    @name = name
  end

  def get_name
    @name
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls := program.Declarations[0].(*ast.ClassDecl)

	// get_name method should reference @name
	getNameMethod := cls.Methods[1]
	if len(getNameMethod.Body) != 1 {
		t.Fatalf("expected 1 statement in get_name body, got %d", len(getNameMethod.Body))
	}

	exprStmt, ok := getNameMethod.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", getNameMethod.Body[0])
	}

	ivar, ok := exprStmt.Expr.(*ast.InstanceVar)
	if !ok {
		t.Fatalf("expected InstanceVar, got %T", exprStmt.Expr)
	}

	if ivar.Name != "name" {
		t.Errorf("expected instance var 'name', got %q", ivar.Name)
	}
}

func TestTypedAssignment(t *testing.T) {
	input := `def main
  x: Int = 5
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

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

	if assign.Type != "Int" {
		t.Errorf("expected type 'Int', got %q", assign.Type)
	}

	intLit, ok := assign.Value.(*ast.IntLit)
	if !ok {
		t.Fatalf("expected IntLit, got %T", assign.Value)
	}

	if intLit.Value != 5 {
		t.Errorf("expected value 5, got %d", intLit.Value)
	}
}

func TestTypedFunctionParams(t *testing.T) {
	input := `def add(a: Int, b: Int): Int
  return a
end

def main
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	if fn.Name != "add" {
		t.Errorf("expected function name 'add', got %q", fn.Name)
	}

	if len(fn.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(fn.Params))
	}

	if fn.Params[0].Name != "a" {
		t.Errorf("expected param name 'a', got %q", fn.Params[0].Name)
	}
	if fn.Params[0].Type != "Int" {
		t.Errorf("expected param type 'Int', got %q", fn.Params[0].Type)
	}

	if fn.Params[1].Name != "b" {
		t.Errorf("expected param name 'b', got %q", fn.Params[1].Name)
	}
	if fn.Params[1].Type != "Int" {
		t.Errorf("expected param type 'Int', got %q", fn.Params[1].Type)
	}
}

func TestMixedTypedUntypedParams(t *testing.T) {
	input := `def foo(a: Int, b: any, c: String)
end

def main
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	if len(fn.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(fn.Params))
	}

	// a: Int
	if fn.Params[0].Name != "a" || fn.Params[0].Type != "Int" {
		t.Errorf("param 0: expected a:Int, got %s:%s", fn.Params[0].Name, fn.Params[0].Type)
	}

	// b: any
	if fn.Params[1].Name != "b" || fn.Params[1].Type != "any" {
		t.Errorf("param 1: expected b:any, got %s:%s", fn.Params[1].Name, fn.Params[1].Type)
	}

	// c: String
	if fn.Params[2].Name != "c" || fn.Params[2].Type != "String" {
		t.Errorf("param 2: expected c:String, got %s:%s", fn.Params[2].Name, fn.Params[2].Type)
	}
}

func TestClassFieldTypeInference(t *testing.T) {
	input := `class User
  @name: String
  @age: Int

  def initialize(name: String, age: Int)
    @name = name
    @age = age
  end
end

def main
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls, ok := program.Declarations[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", program.Declarations[0])
	}

	if len(cls.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(cls.Fields))
	}

	// Check field types from explicit declarations
	fieldTypes := make(map[string]string)
	for _, f := range cls.Fields {
		fieldTypes[f.Name] = f.Type
	}

	if fieldTypes["name"] != "String" {
		t.Errorf("expected field 'name' to have type 'String', got %q", fieldTypes["name"])
	}
	if fieldTypes["age"] != "Int" {
		t.Errorf("expected field 'age' to have type 'Int', got %q", fieldTypes["age"])
	}
}

func TestParameterPromotion(t *testing.T) {
	input := `class Point
  def initialize(@x: Int, @y: Int)
  end
end

def main
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls, ok := program.Declarations[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", program.Declarations[0])
	}

	// Should have 2 fields inferred from promoted parameters
	if len(cls.Fields) != 2 {
		t.Fatalf("expected 2 fields from parameter promotion, got %d", len(cls.Fields))
	}

	fieldTypes := make(map[string]string)
	for _, f := range cls.Fields {
		fieldTypes[f.Name] = f.Type
	}

	if fieldTypes["x"] != "Int" {
		t.Errorf("expected field 'x' to have type 'Int', got %q", fieldTypes["x"])
	}

	if fieldTypes["y"] != "Int" {
		t.Errorf("expected field 'y' to have type 'Int', got %q", fieldTypes["y"])
	}

	// Check that initialize has promoted parameters
	initMethod := cls.Methods[0]
	if len(initMethod.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(initMethod.Params))
	}

	if initMethod.Params[0].Name != "@x" {
		t.Errorf("expected promoted param '@x', got %q", initMethod.Params[0].Name)
	}

	if initMethod.Params[1].Name != "@y" {
		t.Errorf("expected promoted param '@y', got %q", initMethod.Params[1].Name)
	}
}

func TestNewFieldOutsideInitializeError(t *testing.T) {
	input := `class User
  def initialize(name: String)
    @name = name
  end

  def birthday
    @age = 0  # ERROR: cannot introduce new field
  end
end

def main
end`

	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()

	// Should have an error about introducing new field outside initialize
	if len(p.Errors()) == 0 {
		t.Fatal("expected error for new field outside initialize, got none")
	}

	errorMsg := p.Errors()[0]
	if !strings.Contains(errorMsg, "cannot introduce new instance variable @age outside initialize") {
		t.Errorf("expected error about new field, got: %s", errorMsg)
	}
}

func TestTypedMethodParams(t *testing.T) {
	input := `class Calculator
  def initialize
  end

  def add(a: Int, b: Int): Int
    return a
  end
end

def main
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls, ok := program.Declarations[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", program.Declarations[0])
	}

	// Find the add method
	var addMethod *ast.MethodDecl
	for _, m := range cls.Methods {
		if m.Name == "add" {
			addMethod = m
			break
		}
	}

	if addMethod == nil {
		t.Fatal("expected to find 'add' method")
	}

	if len(addMethod.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(addMethod.Params))
	}

	if addMethod.Params[0].Name != "a" || addMethod.Params[0].Type != "Int" {
		t.Errorf("param 0: expected a:Int, got %s:%s", addMethod.Params[0].Name, addMethod.Params[0].Type)
	}
	if addMethod.Params[1].Name != "b" || addMethod.Params[1].Type != "Int" {
		t.Errorf("param 1: expected b:Int, got %s:%s", addMethod.Params[1].Name, addMethod.Params[1].Type)
	}
}

func TestUntypedAssignmentStillWorks(t *testing.T) {
	input := `def main
  x = 5
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	if assign.Name != "x" {
		t.Errorf("expected name 'x', got %q", assign.Name)
	}

	if assign.Type != "" {
		t.Errorf("expected empty type for untyped assignment, got %q", assign.Type)
	}
}

func TestForStatement(t *testing.T) {
	input := `def main
  for item in items
    puts(item)
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	forStmt, ok := fn.Body[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", fn.Body[0])
	}

	if forStmt.Var != "item" {
		t.Errorf("expected loop variable 'item', got %q", forStmt.Var)
	}

	ident, ok := forStmt.Iterable.(*ast.Ident)
	if !ok || ident.Name != "items" {
		t.Errorf("expected iterable 'items', got %v", forStmt.Iterable)
	}

	if len(forStmt.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(forStmt.Body))
	}
}

func TestForStatementTwoVariables(t *testing.T) {
	input := `def main
  for key, value in data
    puts(key)
    puts(value)
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	forStmt, ok := fn.Body[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", fn.Body[0])
	}

	if forStmt.Var != "key" {
		t.Errorf("expected loop variable 'key', got %q", forStmt.Var)
	}

	if forStmt.Var2 != "value" {
		t.Errorf("expected second loop variable 'value', got %q", forStmt.Var2)
	}

	ident, ok := forStmt.Iterable.(*ast.Ident)
	if !ok || ident.Name != "data" {
		t.Errorf("expected iterable 'data', got %v", forStmt.Iterable)
	}

	if len(forStmt.Body) != 2 {
		t.Errorf("expected 2 body statements, got %d", len(forStmt.Body))
	}
}

func TestBreakStatement(t *testing.T) {
	input := `def main
  while true
    break
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

	if len(whileStmt.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(whileStmt.Body))
	}

	_, ok = whileStmt.Body[0].(*ast.BreakStmt)
	if !ok {
		t.Fatalf("expected BreakStmt, got %T", whileStmt.Body[0])
	}
}

func TestNextStatement(t *testing.T) {
	input := `def main
  for item in items
    next
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	forStmt, ok := fn.Body[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", fn.Body[0])
	}

	if len(forStmt.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(forStmt.Body))
	}

	_, ok = forStmt.Body[0].(*ast.NextStmt)
	if !ok {
		t.Fatalf("expected NextStmt, got %T", forStmt.Body[0])
	}
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

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	forStmt, ok := fn.Body[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", fn.Body[0])
	}

	if len(forStmt.Body) != 3 {
		t.Fatalf("expected 3 body statements, got %d", len(forStmt.Body))
	}
}

func TestSelfKeyword(t *testing.T) {
	input := `class Builder
  def initialize
    @name = ""
  end

  def with_name(n: any)
    @name = n
    self
  end
end

def main
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls, ok := program.Declarations[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", program.Declarations[0])
	}

	// Find the with_name method
	var method *ast.MethodDecl
	for _, m := range cls.Methods {
		if m.Name == "with_name" {
			method = m
			break
		}
	}
	if method == nil {
		t.Fatalf("expected to find method 'with_name'")
	}

	// Last statement should be an expression statement containing self
	if len(method.Body) < 2 {
		t.Fatalf("expected at least 2 body statements, got %d", len(method.Body))
	}

	exprStmt, ok := method.Body[1].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", method.Body[1])
	}

	ident, ok := exprStmt.Expr.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident, got %T", exprStmt.Expr)
	}

	if ident.Name != "self" {
		t.Errorf("expected ident name 'self', got %q", ident.Name)
	}
}

func TestStringInterpolation(t *testing.T) {
	input := `def main
  name = "world"
  x = "hello #{name}"
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	if len(fn.Body) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(fn.Body))
	}

	// Second statement should assign an interpolated string
	assignStmt, ok := fn.Body[1].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[1])
	}

	interpStr, ok := assignStmt.Value.(*ast.InterpolatedString)
	if !ok {
		t.Fatalf("expected InterpolatedString, got %T", assignStmt.Value)
	}

	if len(interpStr.Parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(interpStr.Parts))
	}

	// First part should be string "hello "
	part1, ok := interpStr.Parts[0].(string)
	if !ok || part1 != "hello " {
		t.Errorf("expected part 1 to be string 'hello ', got %T: %v", interpStr.Parts[0], interpStr.Parts[0])
	}

	// Second part should be Ident "name"
	part2, ok := interpStr.Parts[1].(*ast.Ident)
	if !ok || part2.Name != "name" {
		t.Errorf("expected part 2 to be Ident 'name', got %T: %v", interpStr.Parts[1], interpStr.Parts[1])
	}
}

func TestStringInterpolationWithExpression(t *testing.T) {
	input := `def main
  x = "sum: #{1 + 2}"
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assignStmt := fn.Body[0].(*ast.AssignStmt)
	interpStr, ok := assignStmt.Value.(*ast.InterpolatedString)
	if !ok {
		t.Fatalf("expected InterpolatedString, got %T", assignStmt.Value)
	}

	if len(interpStr.Parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(interpStr.Parts))
	}

	// First part should be string "sum: "
	part1, ok := interpStr.Parts[0].(string)
	if !ok || part1 != "sum: " {
		t.Errorf("expected part 1 to be 'sum: ', got %v", interpStr.Parts[0])
	}

	// Second part should be a BinaryExpr
	_, ok = interpStr.Parts[1].(*ast.BinaryExpr)
	if !ok {
		t.Errorf("expected part 2 to be BinaryExpr, got %T", interpStr.Parts[1])
	}
}

func TestPlainStringNotInterpolated(t *testing.T) {
	input := `def main
  x = "hello world"
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assignStmt := fn.Body[0].(*ast.AssignStmt)

	// Plain string should still be StringLit
	_, ok := assignStmt.Value.(*ast.StringLit)
	if !ok {
		t.Fatalf("expected StringLit for plain string, got %T", assignStmt.Value)
	}
}

func TestCustomEqualityMethod(t *testing.T) {
	input := `class Point
  def initialize(x: Int, y: Int)
    @x = x
    @y = y
  end

  def ==(other: any)
    @x == other.x and @y == other.y
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls := program.Declarations[0].(*ast.ClassDecl)
	if cls.Name != "Point" {
		t.Errorf("expected class name 'Point', got %q", cls.Name)
	}

	// Should have 2 methods: initialize and ==
	if len(cls.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(cls.Methods))
	}

	// Find the == method
	var eqMethod *ast.MethodDecl
	for _, m := range cls.Methods {
		if m.Name == "==" {
			eqMethod = m
			break
		}
	}

	if eqMethod == nil {
		t.Fatal("expected to find == method")
	}

	if len(eqMethod.Params) != 1 {
		t.Errorf("expected 1 parameter for ==, got %d", len(eqMethod.Params))
	}

	if eqMethod.Params[0].Name != "other" {
		t.Errorf("expected parameter name 'other', got %q", eqMethod.Params[0].Name)
	}
}

func TestInterfaceDeclaration(t *testing.T) {
	input := `interface Speaker
  def speak: String
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	iface, ok := program.Declarations[0].(*ast.InterfaceDecl)
	if !ok {
		t.Fatalf("expected InterfaceDecl, got %T", program.Declarations[0])
	}

	if iface.Name != "Speaker" {
		t.Errorf("expected interface name 'Speaker', got %q", iface.Name)
	}

	if len(iface.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(iface.Methods))
	}

	method := iface.Methods[0]
	if method.Name != "speak" {
		t.Errorf("expected method 'speak', got %q", method.Name)
	}

	if len(method.ReturnTypes) != 1 || method.ReturnTypes[0] != "String" {
		t.Errorf("expected return type 'String', got %v", method.ReturnTypes)
	}
}

func TestInterfaceWithMultipleMethods(t *testing.T) {
	input := `interface ReadWriter
  def read(n: Int): String
  def write(data: String): Int
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	iface := program.Declarations[0].(*ast.InterfaceDecl)

	if iface.Name != "ReadWriter" {
		t.Errorf("expected interface name 'ReadWriter', got %q", iface.Name)
	}

	if len(iface.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(iface.Methods))
	}

	// Check first method
	if iface.Methods[0].Name != "read" {
		t.Errorf("expected method 'read', got %q", iface.Methods[0].Name)
	}
	if len(iface.Methods[0].Params) != 1 {
		t.Errorf("expected 1 param for read, got %d", len(iface.Methods[0].Params))
	}

	// Check second method
	if iface.Methods[1].Name != "write" {
		t.Errorf("expected method 'write', got %q", iface.Methods[1].Name)
	}
}

func TestEmptyInterface(t *testing.T) {
	input := `interface Empty
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	iface := program.Declarations[0].(*ast.InterfaceDecl)
	if len(iface.Methods) != 0 {
		t.Errorf("expected 0 methods, got %d", len(iface.Methods))
	}
}

func TestPubFunction(t *testing.T) {
	input := `pub def add(a: Int, b: Int): Int
  a + b
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

	if fn.Name != "add" {
		t.Errorf("expected function name 'add', got %q", fn.Name)
	}

	if !fn.Pub {
		t.Error("expected function to be pub")
	}
}

func TestPubClass(t *testing.T) {
	input := `pub class User
  def initialize(name: String)
    @name = name
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls, ok := program.Declarations[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", program.Declarations[0])
	}

	if !cls.Pub {
		t.Error("expected class to be pub")
	}
}

func TestPubClassWithPubMethods(t *testing.T) {
	input := `pub class User
  def initialize(name: String)
    @name = name
  end

  pub def get_name: String
    @name
  end

  def internal_method
    @name
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls := program.Declarations[0].(*ast.ClassDecl)

	if !cls.Pub {
		t.Error("expected class to be pub")
	}

	if len(cls.Methods) != 3 {
		t.Fatalf("expected 3 methods, got %d", len(cls.Methods))
	}

	// Find methods
	var getName, internalMethod *ast.MethodDecl
	for _, m := range cls.Methods {
		switch m.Name {
		case "get_name":
			getName = m
		case "internal_method":
			internalMethod = m
		}
	}

	if getName == nil || !getName.Pub {
		t.Error("expected get_name to be pub")
	}

	if internalMethod == nil || internalMethod.Pub {
		t.Error("expected internal_method to not be pub")
	}
}

func TestPubInterface(t *testing.T) {
	input := `pub interface Speaker
  def speak: String
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	iface, ok := program.Declarations[0].(*ast.InterfaceDecl)
	if !ok {
		t.Fatalf("expected InterfaceDecl, got %T", program.Declarations[0])
	}

	if !iface.Pub {
		t.Error("expected interface to be pub")
	}
}

func TestNonPubByDefault(t *testing.T) {
	input := `def helper
end

class InternalClass
end

interface InternalInterface
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 3 {
		t.Fatalf("expected 3 declarations, got %d", len(program.Declarations))
	}

	fn := program.Declarations[0].(*ast.FuncDecl)
	if fn.Pub {
		t.Error("expected function to not be pub by default")
	}

	cls := program.Declarations[1].(*ast.ClassDecl)
	if cls.Pub {
		t.Error("expected class to not be pub by default")
	}

	iface := program.Declarations[2].(*ast.InterfaceDecl)
	if iface.Pub {
		t.Error("expected interface to not be pub by default")
	}
}

func TestPubMustBeFollowedByDefClassOrInterface(t *testing.T) {
	input := `pub x = 5`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected error for 'pub' not followed by def/class/interface")
	}

	if !strings.Contains(p.Errors()[0], "pub") {
		t.Errorf("expected error about 'pub', got: %s", p.Errors()[0])
	}
}

func TestPubDefInNonPubClassIsAllowed(t *testing.T) {
	// pub def inside a non-pub class is allowed - Go supports
	// exported methods on unexported types (e.g., for interface implementation)
	input := `class User
  pub def greet
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Errorf("unexpected errors: %v", p.Errors())
	}

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	cls, ok := program.Declarations[0].(*ast.ClassDecl)
	if !ok {
		t.Fatal("expected ClassDecl")
	}

	if len(cls.Methods) != 1 || !cls.Methods[0].Pub {
		t.Error("expected pub method in class")
	}
}

func TestRangeLiteral(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		start     int64
		end       int64
		exclusive bool
	}{
		{
			name:      "inclusive range",
			input:     "def main\n  x = 1..10\nend",
			start:     1,
			end:       10,
			exclusive: false,
		},
		{
			name:      "exclusive range",
			input:     "def main\n  x = 0...5\nend",
			start:     0,
			end:       5,
			exclusive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			fn := program.Declarations[0].(*ast.FuncDecl)
			assign := fn.Body[0].(*ast.AssignStmt)

			rangeLit, ok := assign.Value.(*ast.RangeLit)
			if !ok {
				t.Fatalf("expected RangeLit, got %T", assign.Value)
			}

			startLit, ok := rangeLit.Start.(*ast.IntLit)
			if !ok {
				t.Fatalf("expected Start to be IntLit, got %T", rangeLit.Start)
			}
			if startLit.Value != tt.start {
				t.Errorf("expected start %d, got %d", tt.start, startLit.Value)
			}

			endLit, ok := rangeLit.End.(*ast.IntLit)
			if !ok {
				t.Fatalf("expected End to be IntLit, got %T", rangeLit.End)
			}
			if endLit.Value != tt.end {
				t.Errorf("expected end %d, got %d", tt.end, endLit.Value)
			}

			if rangeLit.Exclusive != tt.exclusive {
				t.Errorf("expected exclusive=%v, got %v", tt.exclusive, rangeLit.Exclusive)
			}
		})
	}
}

func TestRangeInForLoop(t *testing.T) {
	input := `def main
  for i in 0..5
    puts(i)
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	forStmt, ok := fn.Body[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", fn.Body[0])
	}

	if forStmt.Var != "i" {
		t.Errorf("expected loop var 'i', got %q", forStmt.Var)
	}

	rangeLit, ok := forStmt.Iterable.(*ast.RangeLit)
	if !ok {
		t.Fatalf("expected RangeLit as iterable, got %T", forStmt.Iterable)
	}

	startLit := rangeLit.Start.(*ast.IntLit)
	if startLit.Value != 0 {
		t.Errorf("expected range start 0, got %d", startLit.Value)
	}

	endLit := rangeLit.End.(*ast.IntLit)
	if endLit.Value != 5 {
		t.Errorf("expected range end 5, got %d", endLit.Value)
	}

	if rangeLit.Exclusive {
		t.Error("expected inclusive range (..), got exclusive")
	}
}

func TestRangePrecedence(t *testing.T) {
	// Range should have lower precedence than arithmetic
	// so 1+2..3+4 parses as (1+2)..(3+4)
	input := `def main
  x = 1 + 2..3 + 4
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	rangeLit, ok := assign.Value.(*ast.RangeLit)
	if !ok {
		t.Fatalf("expected RangeLit, got %T", assign.Value)
	}

	// Start should be (1 + 2)
	startBin, ok := rangeLit.Start.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected Start to be BinaryExpr, got %T", rangeLit.Start)
	}
	if startBin.Op != "+" {
		t.Errorf("expected start op '+', got %q", startBin.Op)
	}

	// End should be (3 + 4)
	endBin, ok := rangeLit.End.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected End to be BinaryExpr, got %T", rangeLit.End)
	}
	if endBin.Op != "+" {
		t.Errorf("expected end op '+', got %q", endBin.Op)
	}
}

func TestRangeWithVariables(t *testing.T) {
	input := `def main
  start = 1
  finish = 10
  r = start..finish
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[2].(*ast.AssignStmt)

	rangeLit, ok := assign.Value.(*ast.RangeLit)
	if !ok {
		t.Fatalf("expected RangeLit, got %T", assign.Value)
	}

	startIdent, ok := rangeLit.Start.(*ast.Ident)
	if !ok || startIdent.Name != "start" {
		t.Errorf("expected Start to be ident 'start', got %T", rangeLit.Start)
	}

	endIdent, ok := rangeLit.End.(*ast.Ident)
	if !ok || endIdent.Name != "finish" {
		t.Errorf("expected End to be ident 'finish', got %T", rangeLit.End)
	}
}

func TestOrAssignStmt(t *testing.T) {
	input := `def main
  x ||= 5
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(fn.Body))
	}

	orAssign, ok := fn.Body[0].(*ast.OrAssignStmt)
	if !ok {
		t.Fatalf("expected OrAssignStmt, got %T", fn.Body[0])
	}

	if orAssign.Name != "x" {
		t.Errorf("expected name 'x', got %q", orAssign.Name)
	}

	intLit, ok := orAssign.Value.(*ast.IntLit)
	if !ok {
		t.Fatalf("expected IntLit, got %T", orAssign.Value)
	}
	if intLit.Value != 5 {
		t.Errorf("expected value 5, got %d", intLit.Value)
	}
}

func TestOrAssignWithMethodCall(t *testing.T) {
	input := `def main
  user ||= User.new("guest")
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	orAssign, ok := fn.Body[0].(*ast.OrAssignStmt)
	if !ok {
		t.Fatalf("expected OrAssignStmt, got %T", fn.Body[0])
	}

	if orAssign.Name != "user" {
		t.Errorf("expected name 'user', got %q", orAssign.Name)
	}

	// Value should be a CallExpr (User.new("guest"))
	_, ok = orAssign.Value.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", orAssign.Value)
	}
}

func TestInstanceVarOrAssign(t *testing.T) {
	input := `class User
  def initialize
    @cache = nil
  end

  def get_cache
    @cache ||= load_cache()
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls := program.Declarations[0].(*ast.ClassDecl)
	// get_cache is the second method (after initialize)
	getCacheMethod := cls.Methods[1]

	orAssign, ok := getCacheMethod.Body[0].(*ast.InstanceVarOrAssign)
	if !ok {
		t.Fatalf("expected InstanceVarOrAssign, got %T", getCacheMethod.Body[0])
	}

	if orAssign.Name != "cache" {
		t.Errorf("expected name 'cache', got %q", orAssign.Name)
	}
}

func TestCompoundAssignStmt(t *testing.T) {
	tests := []struct {
		input string
		name  string
		op    string
		value int64
	}{
		{"x += 5", "x", "+", 5},
		{"y -= 10", "y", "-", 10},
		{"z *= 2", "z", "*", 2},
		{"w /= 4", "w", "/", 4},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Declarations) != 1 {
				t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
			}

			compoundAssign, ok := program.Declarations[0].(*ast.CompoundAssignStmt)
			if !ok {
				t.Fatalf("expected CompoundAssignStmt, got %T", program.Declarations[0])
			}

			if compoundAssign.Name != tt.name {
				t.Errorf("expected name %q, got %q", tt.name, compoundAssign.Name)
			}

			if compoundAssign.Op != tt.op {
				t.Errorf("expected op %q, got %q", tt.op, compoundAssign.Op)
			}

			intLit, ok := compoundAssign.Value.(*ast.IntLit)
			if !ok {
				t.Fatalf("expected IntLit value, got %T", compoundAssign.Value)
			}

			if intLit.Value != tt.value {
				t.Errorf("expected value %d, got %d", tt.value, intLit.Value)
			}
		})
	}
}

func TestOptionalTypeInParam(t *testing.T) {
	input := `def find(id: Int?): User?
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)

	if len(fn.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(fn.Params))
	}
	if fn.Params[0].Type != "Int?" {
		t.Errorf("expected param type 'Int?', got %q", fn.Params[0].Type)
	}

	if len(fn.ReturnTypes) != 1 {
		t.Fatalf("expected 1 return type, got %d", len(fn.ReturnTypes))
	}
	if fn.ReturnTypes[0] != "User?" {
		t.Errorf("expected return type 'User?', got %q", fn.ReturnTypes[0])
	}
}

func TestOptionalTypeInVariable(t *testing.T) {
	input := `def main
  x: Int? = nil
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	if assign.Type != "Int?" {
		t.Errorf("expected type 'Int?', got %q", assign.Type)
	}
}

// --- Bare Script Tests ---

func TestBareScriptSimple(t *testing.T) {
	input := `puts("hello")`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	exprStmt, ok := program.Declarations[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", program.Declarations[0])
	}

	call, ok := exprStmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
	}

	ident, ok := call.Func.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident, got %T", call.Func)
	}
	if ident.Name != "puts" {
		t.Errorf("expected 'puts', got %q", ident.Name)
	}
}

func TestBareScriptWithFunctions(t *testing.T) {
	input := `def greet(name: String)
  puts("Hello!")
end

greet("World")`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 2 {
		t.Fatalf("expected 2 declarations, got %d", len(program.Declarations))
	}

	// First should be FuncDecl
	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	if fn.Name != "greet" {
		t.Errorf("expected 'greet', got %q", fn.Name)
	}

	// Second should be ExprStmt (function call)
	exprStmt, ok := program.Declarations[1].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", program.Declarations[1])
	}

	call, ok := exprStmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
	}

	ident, ok := call.Func.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident, got %T", call.Func)
	}
	if ident.Name != "greet" {
		t.Errorf("expected 'greet', got %q", ident.Name)
	}
}

func TestBareScriptWithClass(t *testing.T) {
	input := `class Counter
  def initialize(n: Int)
    @n = n
  end
end

c = Counter.new(0)
puts(c)`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 3 {
		t.Fatalf("expected 3 declarations, got %d", len(program.Declarations))
	}

	// First should be ClassDecl
	cls, ok := program.Declarations[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", program.Declarations[0])
	}
	if cls.Name != "Counter" {
		t.Errorf("expected 'Counter', got %q", cls.Name)
	}

	// Second should be AssignStmt
	assign, ok := program.Declarations[1].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", program.Declarations[1])
	}
	if assign.Name != "c" {
		t.Errorf("expected 'c', got %q", assign.Name)
	}

	// Third should be ExprStmt
	_, ok = program.Declarations[2].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", program.Declarations[2])
	}
}

func TestBareScriptMultipleStatements(t *testing.T) {
	input := `x = 1
y = 2
puts(x)`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 3 {
		t.Fatalf("expected 3 declarations, got %d", len(program.Declarations))
	}

	// All should be statements (2 assigns + 1 expr)
	assign1, ok := program.Declarations[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", program.Declarations[0])
	}
	if assign1.Name != "x" {
		t.Errorf("expected 'x', got %q", assign1.Name)
	}

	assign2, ok := program.Declarations[1].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", program.Declarations[1])
	}
	if assign2.Name != "y" {
		t.Errorf("expected 'y', got %q", assign2.Name)
	}

	_, ok = program.Declarations[2].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", program.Declarations[2])
	}
}

func TestBareScriptWithControlFlow(t *testing.T) {
	input := `x = 5
if x > 3
  puts("big")
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 2 {
		t.Fatalf("expected 2 declarations, got %d", len(program.Declarations))
	}

	// First is assignment
	_, ok := program.Declarations[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", program.Declarations[0])
	}

	// Second is if statement
	_, ok = program.Declarations[1].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", program.Declarations[1])
	}
}

func TestBareScriptWithImport(t *testing.T) {
	input := `import fmt

fmt.Println("hello")`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(program.Imports))
	}
	if program.Imports[0].Path != "fmt" {
		t.Errorf("expected import path 'fmt', got %q", program.Imports[0].Path)
	}

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	_, ok := program.Declarations[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", program.Declarations[0])
	}
}

func TestBareScriptWithForLoop(t *testing.T) {
	input := `for i in 1..5
  puts(i)
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	forStmt, ok := program.Declarations[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", program.Declarations[0])
	}

	if forStmt.Var != "i" {
		t.Errorf("expected loop var 'i', got %q", forStmt.Var)
	}
}

func TestBareScriptWithWhileLoop(t *testing.T) {
	input := `x = 0
while x < 3
  x = x + 1
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 2 {
		t.Fatalf("expected 2 declarations, got %d", len(program.Declarations))
	}

	_, ok := program.Declarations[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", program.Declarations[0])
	}

	_, ok = program.Declarations[1].(*ast.WhileStmt)
	if !ok {
		t.Fatalf("expected WhileStmt, got %T", program.Declarations[1])
	}
}

func TestBareScriptWithBlock(t *testing.T) {
	input := `[1, 2, 3].each -> { |x|
  puts(x)
}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	exprStmt, ok := program.Declarations[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", program.Declarations[0])
	}

	call, ok := exprStmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
	}

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg (lambda), got %d", len(call.Args))
	}

	lambda, ok := call.Args[0].(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("expected LambdaExpr, got %T", call.Args[0])
	}

	if len(lambda.Params) != 1 || lambda.Params[0].Name != "x" {
		t.Errorf("expected lambda params [x], got %v", lambda.Params)
	}
}

func TestBareScriptMixedOrdering(t *testing.T) {
	input := `puts("start")

def helper(x: Int): Int
  x * 2
end

result = helper(5)
puts(result)

class Counter
  def initialize(n: Int)
    @n = n
  end
end

c = Counter.new(0)
puts("done")`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	// Should have: ExprStmt, FuncDecl, AssignStmt, ExprStmt, ClassDecl, AssignStmt, ExprStmt
	if len(program.Declarations) != 7 {
		t.Fatalf("expected 7 declarations, got %d", len(program.Declarations))
	}

	// Verify types in order
	expectedTypes := []string{
		"*ast.ExprStmt",
		"*ast.FuncDecl",
		"*ast.AssignStmt",
		"*ast.ExprStmt",
		"*ast.ClassDecl",
		"*ast.AssignStmt",
		"*ast.ExprStmt",
	}

	for i, decl := range program.Declarations {
		got := fmt.Sprintf("%T", decl)
		if got != expectedTypes[i] {
			t.Errorf("declaration %d: expected %s, got %s", i, expectedTypes[i], got)
		}
	}
}

func TestBareScriptEmpty(t *testing.T) {
	input := ``

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Imports) != 0 {
		t.Errorf("expected 0 imports, got %d", len(program.Imports))
	}
	if len(program.Declarations) != 0 {
		t.Errorf("expected 0 declarations, got %d", len(program.Declarations))
	}
}

func TestBareScriptOnlyComments(t *testing.T) {
	input := `# This is a comment
# Another comment`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 0 {
		t.Errorf("expected 0 declarations, got %d", len(program.Declarations))
	}
}

func TestSymbolLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic symbol",
			input:    "def main\n  x = :ok\nend",
			expected: "ok",
		},
		{
			name:     "symbol with underscores",
			input:    "def main\n  status = :not_found\nend",
			expected: "not_found",
		},
		{
			name:     "symbol in array",
			input:    "def main\n  statuses = [:pending, :active, :completed]\nend",
			expected: "pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			if len(fn.Body) == 0 {
				t.Fatalf("expected at least 1 statement in body")
			}

			var symbolLit *ast.SymbolLit
			switch stmt := fn.Body[0].(type) {
			case *ast.AssignStmt:
				if sym, ok := stmt.Value.(*ast.SymbolLit); ok {
					symbolLit = sym
				} else if arr, ok := stmt.Value.(*ast.ArrayLit); ok && len(arr.Elements) > 0 {
					symbolLit, _ = arr.Elements[0].(*ast.SymbolLit)
				}
			}

			if symbolLit == nil {
				t.Fatalf("expected SymbolLit in assignment")
			}

			if symbolLit.Value != tt.expected {
				t.Errorf("expected symbol value %q, got %q", tt.expected, symbolLit.Value)
			}
		})
	}
}

func TestCaseStatement(t *testing.T) {
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

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	caseStmt, ok := fn.Body[1].(*ast.CaseStmt)
	if !ok {
		t.Fatalf("expected CaseStmt, got %T", fn.Body[1])
	}

	if caseStmt.Subject == nil {
		t.Fatal("expected subject, got nil")
	}

	if len(caseStmt.WhenClauses) != 2 {
		t.Errorf("expected 2 when clauses, got %d", len(caseStmt.WhenClauses))
	}

	// Check first when clause has 1 value
	if len(caseStmt.WhenClauses[0].Values) != 1 {
		t.Errorf("expected first when to have 1 value, got %d", len(caseStmt.WhenClauses[0].Values))
	}

	// Check second when clause has 2 values
	if len(caseStmt.WhenClauses[1].Values) != 2 {
		t.Errorf("expected second when to have 2 values, got %d", len(caseStmt.WhenClauses[1].Values))
	}

	if len(caseStmt.Else) != 1 {
		t.Errorf("expected 1 else statement, got %d", len(caseStmt.Else))
	}
}

func TestCaseStatementNoElse(t *testing.T) {
	input := `def main
  case status
  when 200
    puts("ok")
  when 404
    puts("not found")
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	caseStmt, ok := fn.Body[0].(*ast.CaseStmt)
	if !ok {
		t.Fatalf("expected CaseStmt, got %T", fn.Body[0])
	}

	if caseStmt.Subject == nil {
		t.Fatal("expected subject, got nil")
	}

	if len(caseStmt.WhenClauses) != 2 {
		t.Errorf("expected 2 when clauses, got %d", len(caseStmt.WhenClauses))
	}

	if len(caseStmt.Else) != 0 {
		t.Errorf("expected 0 else statements, got %d", len(caseStmt.Else))
	}
}

func TestCaseStatementNoSubject(t *testing.T) {
	input := `def main
  case
  when x > 10
    puts("big")
  when x > 5
    puts("medium")
  else
    puts("small")
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	caseStmt, ok := fn.Body[0].(*ast.CaseStmt)
	if !ok {
		t.Fatalf("expected CaseStmt, got %T", fn.Body[0])
	}

	if caseStmt.Subject != nil {
		t.Fatal("expected no subject (nil), got non-nil")
	}

	if len(caseStmt.WhenClauses) != 2 {
		t.Errorf("expected 2 when clauses, got %d", len(caseStmt.WhenClauses))
	}

	if len(caseStmt.Else) != 1 {
		t.Errorf("expected 1 else statement, got %d", len(caseStmt.Else))
	}
}

func TestDocCommentAttachment(t *testing.T) {
	input := `# This is a doc comment
def hello
  puts("hi")
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

	if fn.Doc == nil {
		t.Fatal("expected Doc comment, got nil")
	}

	if len(fn.Doc.List) != 1 {
		t.Errorf("expected 1 doc comment, got %d", len(fn.Doc.List))
	}

	if fn.Doc.List[0].Text != "# This is a doc comment" {
		t.Errorf("doc comment text = %q, want %q", fn.Doc.List[0].Text, "# This is a doc comment")
	}
}

func TestMultiLineDocComment(t *testing.T) {
	input := `# First line of doc
# Second line of doc
# Third line of doc
def hello
  puts("hi")
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	if fn.Doc == nil {
		t.Fatal("expected Doc comment, got nil")
	}

	if len(fn.Doc.List) != 3 {
		t.Errorf("expected 3 doc comments, got %d", len(fn.Doc.List))
	}

	expectedTexts := []string{
		"# First line of doc",
		"# Second line of doc",
		"# Third line of doc",
	}

	for i, exp := range expectedTexts {
		if fn.Doc.List[i].Text != exp {
			t.Errorf("doc comment %d: text = %q, want %q", i, fn.Doc.List[i].Text, exp)
		}
	}
}

func TestClassDocComment(t *testing.T) {
	input := `# User represents a user in the system
class User
  def initialize(name: String)
    @name = name
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls, ok := program.Declarations[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", program.Declarations[0])
	}

	if cls.Doc == nil {
		t.Fatal("expected Doc comment on class, got nil")
	}

	if len(cls.Doc.List) != 1 {
		t.Errorf("expected 1 doc comment, got %d", len(cls.Doc.List))
	}
}

func TestMethodDocComment(t *testing.T) {
	input := `class User
  # Greets the user by name
  def greet
    puts(@name)
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cls, ok := program.Declarations[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", program.Declarations[0])
	}

	if len(cls.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(cls.Methods))
	}

	method := cls.Methods[0]
	if method.Doc == nil {
		t.Fatal("expected Doc comment on method, got nil")
	}

	if method.Doc.List[0].Text != "# Greets the user by name" {
		t.Errorf("method doc = %q, want %q", method.Doc.List[0].Text, "# Greets the user by name")
	}
}

func TestImportDocComment(t *testing.T) {
	input := `# Standard library import
import fmt

def main
  puts("hello")
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(program.Imports))
	}

	imp := program.Imports[0]
	if imp.Doc == nil {
		t.Fatal("expected Doc comment on import, got nil")
	}

	if imp.Doc.List[0].Text != "# Standard library import" {
		t.Errorf("import doc = %q, want %q", imp.Doc.List[0].Text, "# Standard library import")
	}
}

func TestFreeFloatingCommentNotAttached(t *testing.T) {
	input := `# Free floating comment

# Doc for function
def hello
  puts("hi")
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	// The free floating comment should NOT be attached (it's separated by blank line)
	// Only the doc comment should be attached
	if fn.Doc == nil {
		t.Fatal("expected Doc comment, got nil")
	}

	if len(fn.Doc.List) != 1 {
		t.Errorf("expected 1 doc comment (not free floating), got %d", len(fn.Doc.List))
	}

	if fn.Doc.List[0].Text != "# Doc for function" {
		t.Errorf("doc = %q, want %q", fn.Doc.List[0].Text, "# Doc for function")
	}
}

func TestProgramCommentsCollected(t *testing.T) {
	input := `# Comment 1
# Comment 2
def hello
  # Body comment
  puts("hi")
end
# End comment`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	// Program.Comments should contain all comment groups
	if len(program.Comments) == 0 {
		t.Fatal("expected Program.Comments to be populated")
	}
}

func TestInterfaceInheritance(t *testing.T) {
	tests := []struct {
		input   string
		parents []string
	}{
		{
			"interface IO < Reader\nend",
			[]string{"Reader"},
		},
		{
			"interface IO < Reader, Writer\nend",
			[]string{"Reader", "Writer"},
		},
		{
			"interface Complex < Reader, Writer, Closer\nend",
			[]string{"Reader", "Writer", "Closer"},
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Declarations) != 1 {
			t.Fatalf("Expected 1 declaration, got %d", len(program.Declarations))
		}

		iface, ok := program.Declarations[0].(*ast.InterfaceDecl)
		if !ok {
			t.Fatalf("Expected InterfaceDecl, got %T", program.Declarations[0])
		}

		if len(iface.Parents) != len(tt.parents) {
			t.Errorf("input %q: expected %d parents, got %d", tt.input, len(tt.parents), len(iface.Parents))
			continue
		}

		for i, expected := range tt.parents {
			if iface.Parents[i] != expected {
				t.Errorf("input %q: parent %d expected %q, got %q", tt.input, i, expected, iface.Parents[i])
			}
		}
	}
}

func TestClassImplements(t *testing.T) {
	tests := []struct {
		input      string
		implements []string
	}{
		{
			"class User implements Speaker\nend",
			[]string{"Speaker"},
		},
		{
			"class User implements Speaker, Serializable\nend",
			[]string{"Speaker", "Serializable"},
		},
		{
			"class Complex implements A, B, C\nend",
			[]string{"A", "B", "C"},
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Declarations) != 1 {
			t.Fatalf("Expected 1 declaration, got %d", len(program.Declarations))
		}

		cls, ok := program.Declarations[0].(*ast.ClassDecl)
		if !ok {
			t.Fatalf("Expected ClassDecl, got %T", program.Declarations[0])
		}

		if len(cls.Implements) != len(tt.implements) {
			t.Errorf("input %q: expected %d implements, got %d", tt.input, len(tt.implements), len(cls.Implements))
			continue
		}

		for i, expected := range tt.implements {
			if cls.Implements[i] != expected {
				t.Errorf("input %q: implements %d expected %q, got %q", tt.input, i, expected, cls.Implements[i])
			}
		}
	}
}

func TestAnyKeywordInParameters(t *testing.T) {
	tests := []struct {
		input     string
		paramType string
	}{
		{"def log(thing: any)\nend", "any"},
		{"def process(a: Int, b: any)\nend", "any"},
		{"def identity(x: any): any\nend", "any"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Declarations) != 1 {
			t.Fatalf("Expected 1 declaration, got %d", len(program.Declarations))
		}

		fn, ok := program.Declarations[0].(*ast.FuncDecl)
		if !ok {
			t.Fatalf("Expected FuncDecl, got %T", program.Declarations[0])
		}

		// Check that at least one parameter has type 'any'
		found := false
		for _, param := range fn.Params {
			if param.Type == "any" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("input %q: expected parameter with type 'any'", tt.input)
		}
	}
}

func TestNilCoalesceOperator(t *testing.T) {
	input := `def main
  x = a ?? b
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}
	nilCoalesce, ok := assign.Value.(*ast.NilCoalesceExpr)
	if !ok {
		t.Fatalf("expected NilCoalesceExpr, got %T", assign.Value)
	}
	if left, ok := nilCoalesce.Left.(*ast.Ident); !ok || left.Name != "a" {
		t.Errorf("expected left 'a', got %v", nilCoalesce.Left)
	}
	if right, ok := nilCoalesce.Right.(*ast.Ident); !ok || right.Name != "b" {
		t.Errorf("expected right 'b', got %v", nilCoalesce.Right)
	}
}

func TestSafeNavOperator(t *testing.T) {
	input := `def main
  x = user&.name
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}
	safeNav, ok := assign.Value.(*ast.SafeNavExpr)
	if !ok {
		t.Fatalf("expected SafeNavExpr, got %T", assign.Value)
	}
	if recv, ok := safeNav.Receiver.(*ast.Ident); !ok || recv.Name != "user" {
		t.Errorf("expected receiver 'user', got %v", safeNav.Receiver)
	}
	if safeNav.Selector != "name" {
		t.Errorf("expected selector 'name', got %q", safeNav.Selector)
	}
}

func TestChainedSafeNavAndNilCoalesce(t *testing.T) {
	input := `def main
  x = user&.address&.city ?? "Unknown"
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	// Should be NilCoalesceExpr at top level
	nilCoalesce, ok := assign.Value.(*ast.NilCoalesceExpr)
	if !ok {
		t.Fatalf("expected NilCoalesceExpr at top, got %T", assign.Value)
	}

	// Right side should be the string "Unknown"
	str, strOk := nilCoalesce.Right.(*ast.StringLit)
	if !strOk || str.Value != "Unknown" {
		t.Errorf("expected right 'Unknown', got %v", nilCoalesce.Right)
	}

	// Left side should be SafeNavExpr for &.city
	safeNav1, ok := nilCoalesce.Left.(*ast.SafeNavExpr)
	if !ok {
		t.Fatalf("expected SafeNavExpr, got %T", nilCoalesce.Left)
	}
	if safeNav1.Selector != "city" {
		t.Errorf("expected selector 'city', got %q", safeNav1.Selector)
	}

	// Inner should be SafeNavExpr for &.address
	safeNav2, ok := safeNav1.Receiver.(*ast.SafeNavExpr)
	if !ok {
		t.Fatalf("expected inner SafeNavExpr, got %T", safeNav1.Receiver)
	}
	if safeNav2.Selector != "address" {
		t.Errorf("expected selector 'address', got %q", safeNav2.Selector)
	}
}

func TestMultiAssignment(t *testing.T) {
	input := `def main
  val, ok = get_data()
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	multi, ok := fn.Body[0].(*ast.MultiAssignStmt)
	if !ok {
		t.Fatalf("expected MultiAssignStmt, got %T", fn.Body[0])
	}

	if len(multi.Names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(multi.Names))
	}
	if multi.Names[0] != "val" || multi.Names[1] != "ok" {
		t.Errorf("expected names [val, ok], got %v", multi.Names)
	}

	call, ok := multi.Value.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", multi.Value)
	}
	if ident, ok := call.Func.(*ast.Ident); !ok || ident.Name != "get_data" {
		t.Errorf("expected call to 'get_data', got %v", call.Func)
	}
}

func TestMultiAssignmentThreeNames(t *testing.T) {
	input := `def main
  a, b, c = get_triple()
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	multi, ok := fn.Body[0].(*ast.MultiAssignStmt)
	if !ok {
		t.Fatalf("expected MultiAssignStmt, got %T", fn.Body[0])
	}

	if len(multi.Names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(multi.Names))
	}
	if multi.Names[0] != "a" || multi.Names[1] != "b" || multi.Names[2] != "c" {
		t.Errorf("expected names [a, b, c], got %v", multi.Names)
	}
}

// =============================================================================
// Command Syntax Tests (optional parentheses)
// =============================================================================

func TestCommandSyntaxBasic(t *testing.T) {
	tests := []struct {
		input    string
		funcName string
		argCount int
	}{
		{`puts "hello"`, "puts", 1},
		{`puts x`, "puts", 1},
		{`print "test"`, "print", 1},
		{`foo bar`, "foo", 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Declarations) != 1 {
				t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
			}

			exprStmt, ok := program.Declarations[0].(*ast.ExprStmt)
			if !ok {
				t.Fatalf("expected ExprStmt, got %T", program.Declarations[0])
			}

			call, ok := exprStmt.Expr.(*ast.CallExpr)
			if !ok {
				t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
			}

			ident, ok := call.Func.(*ast.Ident)
			if !ok {
				t.Fatalf("expected Ident as function, got %T", call.Func)
			}

			if ident.Name != tt.funcName {
				t.Errorf("expected function name %q, got %q", tt.funcName, ident.Name)
			}

			if len(call.Args) != tt.argCount {
				t.Errorf("expected %d args, got %d", tt.argCount, len(call.Args))
			}
		})
	}
}

func TestCommandSyntaxMultipleArgs(t *testing.T) {
	input := `foo a, b, c`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	call := exprStmt.Expr.(*ast.CallExpr)

	if len(call.Args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(call.Args))
	}

	// Check each arg is an identifier
	for i, expected := range []string{"a", "b", "c"} {
		ident, ok := call.Args[i].(*ast.Ident)
		if !ok {
			t.Errorf("arg %d: expected Ident, got %T", i, call.Args[i])
			continue
		}
		if ident.Name != expected {
			t.Errorf("arg %d: expected %q, got %q", i, expected, ident.Name)
		}
	}
}

func TestCommandSyntaxWithArithmetic(t *testing.T) {
	// Arithmetic operators should be included in arguments
	input := `puts x + 1`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	call := exprStmt.Expr.(*ast.CallExpr)

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}

	// The argument should be a binary expression (x + 1)
	binary, ok := call.Args[0].(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr arg, got %T", call.Args[0])
	}

	if binary.Op != "+" {
		t.Errorf("expected '+' operator, got %q", binary.Op)
	}
}

func TestCommandSyntaxWithComparison(t *testing.T) {
	// Comparison operators should be included in arguments
	input := `puts x == y`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	call := exprStmt.Expr.(*ast.CallExpr)

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}

	binary, ok := call.Args[0].(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr arg, got %T", call.Args[0])
	}

	if binary.Op != "==" {
		t.Errorf("expected '==' operator, got %q", binary.Op)
	}
}

func TestCommandSyntaxMethodChain(t *testing.T) {
	// Method chain with command syntax
	input := `foo.bar baz`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	call := exprStmt.Expr.(*ast.CallExpr)

	// Function should be a selector expression
	sel, ok := call.Func.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr as function, got %T", call.Func)
	}

	if sel.Sel != "bar" {
		t.Errorf("expected selector 'bar', got %q", sel.Sel)
	}

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}
}

func TestCommandSyntaxInAssignment(t *testing.T) {
	// Command syntax on right side of assignment
	input := `def main
  x = foo bar
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	call, ok := assign.Value.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", assign.Value)
	}

	ident := call.Func.(*ast.Ident)
	if ident.Name != "foo" {
		t.Errorf("expected function 'foo', got %q", ident.Name)
	}

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}
}

func TestCommandSyntaxWithBlock(t *testing.T) {
	// Command syntax followed by do block
	input := `arr.each -> { |x|
  puts x
}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	call := exprStmt.Expr.(*ast.CallExpr)

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg (lambda), got %d", len(call.Args))
	}

	lambda, ok := call.Args[0].(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("expected LambdaExpr, got %T", call.Args[0])
	}

	if len(lambda.Params) != 1 {
		t.Errorf("expected 1 lambda param, got %d", len(lambda.Params))
	}
}

func TestCommandSyntaxUnaryMinus(t *testing.T) {
	// `foo -1` should be command syntax with unary minus arg
	input := `foo -1`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	call, ok := exprStmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
	}

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}

	// The argument should be a unary expression (-1)
	unary, ok := call.Args[0].(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr arg, got %T", call.Args[0])
	}

	if unary.Op != "-" {
		t.Errorf("expected '-' operator, got %q", unary.Op)
	}
}

func TestCommandSyntaxBinaryMinus(t *testing.T) {
	// `foo - 1` should be binary subtraction, not command syntax
	input := `foo - 1`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	binary, ok := exprStmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", exprStmt.Expr)
	}

	if binary.Op != "-" {
		t.Errorf("expected '-' operator, got %q", binary.Op)
	}
}

func TestCommandSyntaxWithLiterals(t *testing.T) {
	tests := []struct {
		input   string
		argType string
	}{
		{`foo 42`, "*ast.IntLit"},
		{`foo 3.14`, "*ast.FloatLit"},
		{`foo "str"`, "*ast.StringLit"},
		{`foo true`, "*ast.BoolLit"},
		{`foo false`, "*ast.BoolLit"},
		{`foo nil`, "*ast.NilLit"},
		{`foo :sym`, "*ast.SymbolLit"},
		{`foo [1, 2]`, "*ast.ArrayLit"},
		// Note: {a => b} is not supported in command syntax - use foo({"a" => 1})
		// because { } after a method call is parsed as a block
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			exprStmt := program.Declarations[0].(*ast.ExprStmt)
			call, ok := exprStmt.Expr.(*ast.CallExpr)
			if !ok {
				t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
			}

			if len(call.Args) != 1 {
				t.Fatalf("expected 1 arg, got %d", len(call.Args))
			}

			argType := fmt.Sprintf("%T", call.Args[0])
			if argType != tt.argType {
				t.Errorf("expected arg type %s, got %s", tt.argType, argType)
			}
		})
	}
}

func TestCommandSyntaxParensStillWork(t *testing.T) {
	// Ensure parenthesized calls still work
	input := `puts("hello")`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	call, ok := exprStmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
	}

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}
}

func TestCommandSyntaxNoSpaceNoCommand(t *testing.T) {
	// `foo"bar"` without space should NOT be command syntax
	// It should just be an identifier (the string is a separate statement)
	input := `x = foo"bar"`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	// This might generate errors, but we're checking the parse behavior
	_ = p.Errors()

	if len(program.Declarations) == 0 {
		t.Fatal("expected at least one declaration")
	}
}

func TestCommandSyntaxArrayVsIndex(t *testing.T) {
	// `foo[1]` without space is index expression
	t.Run("index_no_space", func(t *testing.T) {
		input := `foo[1]`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		exprStmt := program.Declarations[0].(*ast.ExprStmt)
		indexExpr, ok := exprStmt.Expr.(*ast.IndexExpr)
		if !ok {
			t.Fatalf("expected IndexExpr, got %T", exprStmt.Expr)
		}

		ident, ok := indexExpr.Left.(*ast.Ident)
		if !ok || ident.Name != "foo" {
			t.Errorf("expected left to be 'foo', got %v", indexExpr.Left)
		}
	})

	// `foo [1, 2]` with space is command syntax with array arg
	t.Run("array_with_space", func(t *testing.T) {
		input := `foo [1, 2]`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		exprStmt := program.Declarations[0].(*ast.ExprStmt)
		call, ok := exprStmt.Expr.(*ast.CallExpr)
		if !ok {
			t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
		}

		if len(call.Args) != 1 {
			t.Fatalf("expected 1 arg, got %d", len(call.Args))
		}

		_, ok = call.Args[0].(*ast.ArrayLit)
		if !ok {
			t.Errorf("expected ArrayLit arg, got %T", call.Args[0])
		}
	})
}

func TestCommandSyntaxWithAndOr(t *testing.T) {
	// `puts x and y` parses as `puts(x and y)` - and/or are part of the argument
	t.Run("and_in_argument", func(t *testing.T) {
		input := `puts x and y`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		exprStmt := program.Declarations[0].(*ast.ExprStmt)
		call, ok := exprStmt.Expr.(*ast.CallExpr)
		if !ok {
			t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
		}

		if len(call.Args) != 1 {
			t.Errorf("expected 1 arg in call, got %d", len(call.Args))
		}

		// Argument should be BinaryExpr(x && y)
		binary, ok := call.Args[0].(*ast.BinaryExpr)
		if !ok {
			t.Fatalf("expected BinaryExpr arg, got %T", call.Args[0])
		}

		if binary.Op != "&&" {
			t.Errorf("expected '&&' operator, got %q", binary.Op)
		}
	})

	t.Run("or_in_argument", func(t *testing.T) {
		input := `puts x or y`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		exprStmt := program.Declarations[0].(*ast.ExprStmt)
		call, ok := exprStmt.Expr.(*ast.CallExpr)
		if !ok {
			t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
		}

		// Argument should be BinaryExpr(x || y)
		binary, ok := call.Args[0].(*ast.BinaryExpr)
		if !ok {
			t.Fatalf("expected BinaryExpr arg, got %T", call.Args[0])
		}

		if binary.Op != "||" {
			t.Errorf("expected '||' operator, got %q", binary.Op)
		}
	})
}

func TestCommandSyntaxWithTrailingBlock(t *testing.T) {
	// Arrow lambda syntax as separate argument
	// `foo(bar, -> { |x| puts x })`
	input := `foo(bar, -> { |x|
  puts x
})`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	call, ok := exprStmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.Expr)
	}

	// Should have 2 args: bar (ident) and lambda
	if len(call.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(call.Args))
	}

	// Second arg should be a lambda
	lambda, ok := call.Args[1].(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("expected LambdaExpr as second arg, got %T", call.Args[1])
	}

	if len(lambda.Params) != 1 {
		t.Errorf("expected 1 lambda param, got %d", len(lambda.Params))
	}
}

func TestSymbolKeyShorthand(t *testing.T) {
	input := `{name: "Alice", age: 30}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	mapLit, ok := exprStmt.Expr.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", exprStmt.Expr)
	}

	if len(mapLit.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(mapLit.Entries))
	}

	// Check first entry: name: "Alice"
	key1, ok := mapLit.Entries[0].Key.(*ast.StringLit)
	if !ok || key1.Value != "name" {
		t.Errorf("expected key 'name', got %v", mapLit.Entries[0].Key)
	}
	val1, ok := mapLit.Entries[0].Value.(*ast.StringLit)
	if !ok || val1.Value != "Alice" {
		t.Errorf("expected value 'Alice', got %v", mapLit.Entries[0].Value)
	}

	// Check second entry: age: 30
	key2, ok := mapLit.Entries[1].Key.(*ast.StringLit)
	if !ok || key2.Value != "age" {
		t.Errorf("expected key 'age', got %v", mapLit.Entries[1].Key)
	}
	val2, ok := mapLit.Entries[1].Value.(*ast.IntLit)
	if !ok || val2.Value != 30 {
		t.Errorf("expected value 30, got %v", mapLit.Entries[1].Value)
	}
}

func TestImplicitValueShorthand(t *testing.T) {
	input := `{name:, age:}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	mapLit, ok := exprStmt.Expr.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", exprStmt.Expr)
	}

	if len(mapLit.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(mapLit.Entries))
	}

	// First entry: name: name
	key1, ok := mapLit.Entries[0].Key.(*ast.StringLit)
	if !ok || key1.Value != "name" {
		t.Errorf("expected key 'name', got %v", mapLit.Entries[0].Key)
	}
	val1, ok := mapLit.Entries[0].Value.(*ast.Ident)
	if !ok || val1.Name != "name" {
		t.Errorf("expected value ident 'name', got %v", mapLit.Entries[0].Value)
	}

	// Second entry: age: age
	key2, ok := mapLit.Entries[1].Key.(*ast.StringLit)
	if !ok || key2.Value != "age" {
		t.Errorf("expected key 'age', got %v", mapLit.Entries[1].Key)
	}
	val2, ok := mapLit.Entries[1].Value.(*ast.Ident)
	if !ok || val2.Name != "age" {
		t.Errorf("expected value ident 'age', got %v", mapLit.Entries[1].Value)
	}
}

func TestWordArrayLiteral(t *testing.T) {
	input := `%w{foo bar baz}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	arr, ok := exprStmt.Expr.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected ArrayLit, got %T", exprStmt.Expr)
	}

	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}

	expected := []string{"foo", "bar", "baz"}
	for i, elem := range arr.Elements {
		strLit, ok := elem.(*ast.StringLit)
		if !ok {
			t.Errorf("element %d: expected StringLit, got %T", i, elem)
			continue
		}
		if strLit.Value != expected[i] {
			t.Errorf("element %d: expected %q, got %q", i, expected[i], strLit.Value)
		}
	}
}

func TestArraySplat(t *testing.T) {
	input := `[1, *rest, 3]`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	arr, ok := exprStmt.Expr.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected ArrayLit, got %T", exprStmt.Expr)
	}

	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}

	// First element: 1
	_, ok = arr.Elements[0].(*ast.IntLit)
	if !ok {
		t.Errorf("element 0: expected IntLit, got %T", arr.Elements[0])
	}

	// Second element: *rest (SplatExpr)
	splat, ok := arr.Elements[1].(*ast.SplatExpr)
	if !ok {
		t.Fatalf("element 1: expected SplatExpr, got %T", arr.Elements[1])
	}
	ident, ok := splat.Expr.(*ast.Ident)
	if !ok || ident.Name != "rest" {
		t.Errorf("splat expression: expected ident 'rest', got %v", splat.Expr)
	}

	// Third element: 3
	_, ok = arr.Elements[2].(*ast.IntLit)
	if !ok {
		t.Errorf("element 2: expected IntLit, got %T", arr.Elements[2])
	}
}

func TestMapDoubleSplat(t *testing.T) {
	input := `{**defaults, name: "Alice"}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	mapLit, ok := exprStmt.Expr.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", exprStmt.Expr)
	}

	if len(mapLit.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(mapLit.Entries))
	}

	// First entry: **defaults (splat)
	if mapLit.Entries[0].Splat == nil {
		t.Fatal("expected first entry to be a splat")
	}
	ident, ok := mapLit.Entries[0].Splat.(*ast.Ident)
	if !ok || ident.Name != "defaults" {
		t.Errorf("splat expression: expected ident 'defaults', got %v", mapLit.Entries[0].Splat)
	}

	// Second entry: name: "Alice"
	if mapLit.Entries[1].Splat != nil {
		t.Error("expected second entry to NOT be a splat")
	}
	key, ok := mapLit.Entries[1].Key.(*ast.StringLit)
	if !ok || key.Value != "name" {
		t.Errorf("expected key 'name', got %v", mapLit.Entries[1].Key)
	}
}

func TestMixedMapSyntax(t *testing.T) {
	// Test mixing hash rocket and symbol key shorthand
	input := `{"explicit" => 1, short: 2}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	mapLit, ok := exprStmt.Expr.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", exprStmt.Expr)
	}

	if len(mapLit.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(mapLit.Entries))
	}

	// First entry: "explicit" => 1 (hash rocket)
	key1, ok := mapLit.Entries[0].Key.(*ast.StringLit)
	if !ok || key1.Value != "explicit" {
		t.Errorf("expected key 'explicit', got %v", mapLit.Entries[0].Key)
	}

	// Second entry: short: 2 (symbol key shorthand)
	key2, ok := mapLit.Entries[1].Key.(*ast.StringLit)
	if !ok || key2.Value != "short" {
		t.Errorf("expected key 'short', got %v", mapLit.Entries[1].Key)
	}
}

func TestInterpolatedWordArray(t *testing.T) {
	input := `%W{hello #{name} world}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	arr, ok := exprStmt.Expr.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected ArrayLit, got %T", exprStmt.Expr)
	}

	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}

	// First element: "hello" (plain string)
	str1, ok := arr.Elements[0].(*ast.StringLit)
	if !ok || str1.Value != "hello" {
		t.Errorf("element 0: expected StringLit 'hello', got %T %v", arr.Elements[0], arr.Elements[0])
	}

	// Second element: interpolated string with #{name}
	interp, ok := arr.Elements[1].(*ast.InterpolatedString)
	if !ok {
		t.Errorf("element 1: expected InterpolatedString, got %T", arr.Elements[1])
	} else if len(interp.Parts) != 1 {
		t.Errorf("element 1: expected 1 part in interpolation, got %d", len(interp.Parts))
	}

	// Third element: "world" (plain string)
	str3, ok := arr.Elements[2].(*ast.StringLit)
	if !ok || str3.Value != "world" {
		t.Errorf("element 2: expected StringLit 'world', got %T %v", arr.Elements[2], arr.Elements[2])
	}
}

func TestArraySplatAtEdges(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		numElems int
	}{
		{"splat only", "[*items]", 1},
		{"splat at start", "[*first, 2, 3]", 3},
		{"splat at end", "[1, 2, *last]", 3},
		{"multiple splats", "[*a, middle, *b]", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			exprStmt := program.Declarations[0].(*ast.ExprStmt)
			arr, ok := exprStmt.Expr.(*ast.ArrayLit)
			if !ok {
				t.Fatalf("expected ArrayLit, got %T", exprStmt.Expr)
			}

			if len(arr.Elements) != tt.numElems {
				t.Errorf("expected %d elements, got %d", tt.numElems, len(arr.Elements))
			}
		})
	}
}

func TestSplatWithExpression(t *testing.T) {
	// Splat with method call, not just identifier
	input := `[*get_items(), 1]`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	arr, ok := exprStmt.Expr.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected ArrayLit, got %T", exprStmt.Expr)
	}

	splat, ok := arr.Elements[0].(*ast.SplatExpr)
	if !ok {
		t.Fatalf("expected SplatExpr, got %T", arr.Elements[0])
	}

	// The splat should contain a call expression
	_, ok = splat.Expr.(*ast.CallExpr)
	if !ok {
		t.Errorf("expected CallExpr inside splat, got %T", splat.Expr)
	}
}

func TestMultipleDoubleSplats(t *testing.T) {
	input := `{**defaults, **overrides, final: 1}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	mapLit, ok := exprStmt.Expr.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", exprStmt.Expr)
	}

	if len(mapLit.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(mapLit.Entries))
	}

	// First two should be splats
	if mapLit.Entries[0].Splat == nil {
		t.Error("expected entry 0 to be a splat")
	}
	if mapLit.Entries[1].Splat == nil {
		t.Error("expected entry 1 to be a splat")
	}
	// Third should be regular entry
	if mapLit.Entries[2].Splat != nil {
		t.Error("expected entry 2 to NOT be a splat")
	}
}

func TestSymbolShorthandTrailingComma(t *testing.T) {
	input := `{name: "test", age: 30,}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	mapLit, ok := exprStmt.Expr.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", exprStmt.Expr)
	}

	if len(mapLit.Entries) != 2 {
		t.Errorf("expected 2 entries (trailing comma should be allowed), got %d", len(mapLit.Entries))
	}
}

func TestArraySplatTrailingComma(t *testing.T) {
	input := `[1, *rest,]`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Declarations[0].(*ast.ExprStmt)
	arr, ok := exprStmt.Expr.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected ArrayLit, got %T", exprStmt.Expr)
	}

	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 elements (trailing comma should be allowed), got %d", len(arr.Elements))
	}
}

// Module tests

func TestModuleDecl(t *testing.T) {
	input := `module Loggable
  def log(msg: String)
    puts msg
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	mod, ok := program.Declarations[0].(*ast.ModuleDecl)
	if !ok {
		t.Fatalf("expected ModuleDecl, got %T", program.Declarations[0])
	}

	if mod.Name != "Loggable" {
		t.Errorf("expected module name 'Loggable', got %q", mod.Name)
	}

	if len(mod.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(mod.Methods))
	}

	if mod.Methods[0].Name != "log" {
		t.Errorf("expected method name 'log', got %q", mod.Methods[0].Name)
	}
}

func TestModuleWithAccessor(t *testing.T) {
	input := `module Countable
  getter count: Int

  def increment
    puts "incrementing"
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	mod, ok := program.Declarations[0].(*ast.ModuleDecl)
	if !ok {
		t.Fatalf("expected ModuleDecl, got %T", program.Declarations[0])
	}

	if len(mod.Accessors) != 1 {
		t.Fatalf("expected 1 accessor, got %d", len(mod.Accessors))
	}

	if mod.Accessors[0].Kind != "getter" {
		t.Errorf("expected accessor kind 'getter', got %q", mod.Accessors[0].Kind)
	}
	if mod.Accessors[0].Name != "count" {
		t.Errorf("expected accessor name 'count', got %q", mod.Accessors[0].Name)
	}
	if mod.Accessors[0].Type != "Int" {
		t.Errorf("expected accessor type 'Int', got %q", mod.Accessors[0].Type)
	}
}

func TestModuleWithField(t *testing.T) {
	input := `module Cacheable
  @data: Map<String, any>
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	mod, ok := program.Declarations[0].(*ast.ModuleDecl)
	if !ok {
		t.Fatalf("expected ModuleDecl, got %T", program.Declarations[0])
	}

	if len(mod.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(mod.Fields))
	}

	if mod.Fields[0].Name != "data" {
		t.Errorf("expected field name 'data', got %q", mod.Fields[0].Name)
	}
}

// Accessor tests

func TestAccessorDecl_Getter(t *testing.T) {
	input := `class User
  getter name: String

  def initialize(@name: String)
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	class, ok := program.Declarations[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", program.Declarations[0])
	}

	if len(class.Accessors) != 1 {
		t.Fatalf("expected 1 accessor, got %d", len(class.Accessors))
	}

	accessor := class.Accessors[0]
	if accessor.Kind != "getter" {
		t.Errorf("expected kind 'getter', got %q", accessor.Kind)
	}
	if accessor.Name != "name" {
		t.Errorf("expected name 'name', got %q", accessor.Name)
	}
	if accessor.Type != "String" {
		t.Errorf("expected type 'String', got %q", accessor.Type)
	}
}

func TestAccessorDecl_Setter(t *testing.T) {
	input := `class User
  setter name: String

  def initialize(@name: String)
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	class := program.Declarations[0].(*ast.ClassDecl)
	if len(class.Accessors) != 1 {
		t.Fatalf("expected 1 accessor, got %d", len(class.Accessors))
	}

	accessor := class.Accessors[0]
	if accessor.Kind != "setter" {
		t.Errorf("expected kind 'setter', got %q", accessor.Kind)
	}
}

func TestAccessorDecl_Property(t *testing.T) {
	input := `class User
  property email: String

  def initialize(@email: String)
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	class := program.Declarations[0].(*ast.ClassDecl)
	if len(class.Accessors) != 1 {
		t.Fatalf("expected 1 accessor, got %d", len(class.Accessors))
	}

	accessor := class.Accessors[0]
	if accessor.Kind != "property" {
		t.Errorf("expected kind 'property', got %q", accessor.Kind)
	}
}

func TestAccessorDecl_WithAnyType(t *testing.T) {
	input := `class Container
  property data: any

  def initialize
    @data = nil
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	class := program.Declarations[0].(*ast.ClassDecl)
	if len(class.Accessors) != 1 {
		t.Fatalf("expected 1 accessor, got %d", len(class.Accessors))
	}

	if class.Accessors[0].Type != "any" {
		t.Errorf("expected type 'any', got %q", class.Accessors[0].Type)
	}
}

// Super expression tests

func TestSuperExpr_NoArgs(t *testing.T) {
	input := `class Child < Parent
  def greet
    super
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	class := program.Declarations[0].(*ast.ClassDecl)
	method := class.Methods[0]

	exprStmt, ok := method.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", method.Body[0])
	}

	superExpr, ok := exprStmt.Expr.(*ast.SuperExpr)
	if !ok {
		t.Fatalf("expected SuperExpr, got %T", exprStmt.Expr)
	}

	if len(superExpr.Args) != 0 {
		t.Errorf("expected 0 args, got %d", len(superExpr.Args))
	}
}

func TestSuperExpr_WithArgs(t *testing.T) {
	input := `class Child < Parent
  def initialize(name: String, age: Int)
    super(name, age)
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	class := program.Declarations[0].(*ast.ClassDecl)
	method := class.Methods[0]

	exprStmt, ok := method.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", method.Body[0])
	}

	superExpr, ok := exprStmt.Expr.(*ast.SuperExpr)
	if !ok {
		t.Fatalf("expected SuperExpr, got %T", exprStmt.Expr)
	}

	if len(superExpr.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(superExpr.Args))
	}
}

// Rescue expression tests

func TestRescueExpr_InlineDefault(t *testing.T) {
	input := `def main
  port = parse_int("8080") rescue 3000
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", fn.Body[0])
	}

	rescue, ok := assign.Value.(*ast.RescueExpr)
	if !ok {
		t.Fatalf("expected RescueExpr, got %T", assign.Value)
	}

	if rescue.Default == nil {
		t.Fatal("expected Default to be set")
	}
	if rescue.Block != nil {
		t.Error("expected Block to be nil for inline form")
	}

	intLit, ok := rescue.Default.(*ast.IntLit)
	if !ok {
		t.Fatalf("expected IntLit default, got %T", rescue.Default)
	}
	if intLit.Value != 3000 {
		t.Errorf("expected default 3000, got %d", intLit.Value)
	}
}

func TestRescueExpr_BlockForm(t *testing.T) {
	input := `def main
  data = read_file(path) rescue do
    default_data
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	rescue, ok := assign.Value.(*ast.RescueExpr)
	if !ok {
		t.Fatalf("expected RescueExpr, got %T", assign.Value)
	}

	if rescue.Block == nil {
		t.Fatal("expected Block to be set for block form")
	}
	if rescue.Default != nil {
		t.Error("expected Default to be nil for block form")
	}
}

func TestRescueExpr_WithErrorBinding(t *testing.T) {
	input := `def main
  data = read_file(path) rescue => err do
    puts err
    fallback
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	rescue, ok := assign.Value.(*ast.RescueExpr)
	if !ok {
		t.Fatalf("expected RescueExpr, got %T", assign.Value)
	}

	if rescue.ErrName != "err" {
		t.Errorf("expected ErrName 'err', got %q", rescue.ErrName)
	}
	if rescue.Block == nil {
		t.Fatal("expected Block to be set")
	}
}

func TestRescueExpr_MethodCall(t *testing.T) {
	input := `def main
  value = obj.get_value() rescue nil
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	rescue, ok := assign.Value.(*ast.RescueExpr)
	if !ok {
		t.Fatalf("expected RescueExpr, got %T", assign.Value)
	}

	call, ok := rescue.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr in rescue, got %T", rescue.Expr)
	}

	if call.Func == nil {
		t.Fatal("expected Func to be set")
	}
}

func TestRescueExpr_SelectorWithoutParens(t *testing.T) {
	input := `def main
  value = obj.get_value rescue nil
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	rescue, ok := assign.Value.(*ast.RescueExpr)
	if !ok {
		t.Fatalf("expected RescueExpr, got %T", assign.Value)
	}

	call, ok := rescue.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr (converted from selector), got %T", rescue.Expr)
	}

	if call.Func == nil {
		t.Fatal("expected Func to be set")
	}
}

func TestRescueExpr_IdentWithoutParens(t *testing.T) {
	input := `def main
  value = fetch rescue nil
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	rescue, ok := assign.Value.(*ast.RescueExpr)
	if !ok {
		t.Fatalf("expected RescueExpr, got %T", assign.Value)
	}

	call, ok := rescue.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr (converted from ident), got %T", rescue.Expr)
	}

	ident, ok := call.Func.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident in call, got %T", call.Func)
	}
	if ident.Name != "fetch" {
		t.Errorf("expected func name 'fetch', got %q", ident.Name)
	}
}

// Error formatting tests

func TestFormatErrors(t *testing.T) {
	tests := []struct {
		name     string
		errors   []ParseError
		expected string
	}{
		{
			name:     "empty errors",
			errors:   []ParseError{},
			expected: "",
		},
		{
			name: "single error without hint",
			errors: []ParseError{
				{Line: 1, Column: 5, Message: "unexpected token"},
			},
			expected: "1:5: unexpected token",
		},
		{
			name: "single error with hint",
			errors: []ParseError{
				{Line: 2, Column: 10, Message: "missing type", Hint: "add: Type"},
			},
			expected: "2:10: missing type (hint: add: Type)",
		},
		{
			name: "multiple errors",
			errors: []ParseError{
				{Line: 1, Column: 1, Message: "error one"},
				{Line: 2, Column: 2, Message: "error two"},
			},
			expected: "1:1: error one\n2:2: error two",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatErrors(tt.errors)
			if got != tt.expected {
				t.Errorf("FormatErrors() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	input := `def main
  if
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	errs := p.ParseErrors()
	if len(errs) == 0 {
		t.Fatal("expected parse errors, got none")
	}

	err := errs[0]
	if err.Line == 0 {
		t.Error("expected Line to be set")
	}
	if err.Message == "" {
		t.Error("expected Message to be set")
	}
}

func TestParseError_String(t *testing.T) {
	tests := []struct {
		name     string
		err      ParseError
		expected string
	}{
		{
			name:     "without hint",
			err:      ParseError{Line: 5, Column: 10, Message: "test error"},
			expected: "5:10: test error",
		},
		{
			name:     "with hint",
			err:      ParseError{Line: 3, Column: 1, Message: "missing end", Hint: "add 'end' keyword"},
			expected: "3:1: missing end (hint: add 'end' keyword)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.String()
			if got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// Error case tests

func TestModuleDecl_MissingName(t *testing.T) {
	input := `module
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for missing module name")
	}
}

func TestModuleDecl_MissingEnd(t *testing.T) {
	input := `module Foo
  def bar
  end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for missing end")
	}
}

func TestModuleDecl_NestedClass(t *testing.T) {
	input := `module Foo
  class Bar
    getter name: String

    def initialize(@name: String)
    end
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	mod, ok := program.Declarations[0].(*ast.ModuleDecl)
	if !ok {
		t.Fatalf("expected ModuleDecl, got %T", program.Declarations[0])
	}

	if mod.Name != "Foo" {
		t.Errorf("expected module name 'Foo', got '%s'", mod.Name)
	}

	if len(mod.Classes) != 1 {
		t.Fatalf("expected 1 nested class, got %d", len(mod.Classes))
	}

	cls := mod.Classes[0]
	if cls.Name != "Bar" {
		t.Errorf("expected class name 'Bar', got '%s'", cls.Name)
	}

	if len(cls.Accessors) != 1 || cls.Accessors[0].Name != "name" {
		t.Errorf("expected getter 'name', got %v", cls.Accessors)
	}

	if len(cls.Methods) != 1 || cls.Methods[0].Name != "initialize" {
		t.Errorf("expected method 'initialize', got %v", cls.Methods)
	}
}

func TestAccessorDecl_MissingType(t *testing.T) {
	input := `class Foo
  getter name
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for missing accessor type")
	}
}

func TestAccessorDecl_MissingName(t *testing.T) {
	input := `class Foo
  getter: String
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for missing accessor name")
	}
}

func TestSuperExpr_UnclosedParens(t *testing.T) {
	input := `class Child < Parent
  def foo
    super(1, 2
  end
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for unclosed parentheses in super")
	}
}

func TestTypeNameParsing_GenericTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "class Foo\n  @items: Array<String>\nend",
			expected: "Array<String>",
		},
		{
			input:    "class Foo\n  @map: Map<String, Int>\nend",
			expected: "Map<String, Int>",
		},
		{
			input:    "class Foo\n  @nested: Array<Map<String, Int>>\nend",
			expected: "Array<Map<String, Int>>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			class := program.Declarations[0].(*ast.ClassDecl)
			if len(class.Fields) < 1 {
				t.Fatal("expected at least 1 field")
			}
			if class.Fields[0].Type != tt.expected {
				t.Errorf("expected type %q, got %q", tt.expected, class.Fields[0].Type)
			}
		})
	}
}

func TestTypeNameParsing_Optional(t *testing.T) {
	input := `class Foo
  @value: String?
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	class := program.Declarations[0].(*ast.ClassDecl)
	if len(class.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(class.Fields))
	}
	if class.Fields[0].Type != "String?" {
		t.Errorf("expected type 'String?', got %q", class.Fields[0].Type)
	}
}

func TestFieldDecl_MissingType(t *testing.T) {
	input := `class Foo
  @name
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for field without type")
	}
}

func TestMethodSig_InInterface(t *testing.T) {
	input := `interface Speaker
  def speak(msg: String): String
  def volume: Int
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	iface := program.Declarations[0].(*ast.InterfaceDecl)
	if len(iface.Methods) != 2 {
		t.Fatalf("expected 2 method signatures, got %d", len(iface.Methods))
	}

	sig := iface.Methods[0]
	if sig.Name != "speak" {
		t.Errorf("expected method name 'speak', got %q", sig.Name)
	}
	if len(sig.Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(sig.Params))
	}
	if len(sig.ReturnTypes) != 1 || sig.ReturnTypes[0] != "String" {
		t.Errorf("expected return type 'String', got %v", sig.ReturnTypes)
	}
}

func TestMethodSig_MultipleReturnTypes(t *testing.T) {
	input := `interface Parser
  def parse(input: String): (AST, Error)
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	iface := program.Declarations[0].(*ast.InterfaceDecl)
	if len(iface.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(iface.Methods))
	}

	sig := iface.Methods[0]
	if len(sig.ReturnTypes) != 2 {
		t.Fatalf("expected 2 return types, got %d", len(sig.ReturnTypes))
	}
	if sig.ReturnTypes[0] != "AST" {
		t.Errorf("expected first return type 'AST', got %q", sig.ReturnTypes[0])
	}
	if sig.ReturnTypes[1] != "Error" {
		t.Errorf("expected second return type 'Error', got %q", sig.ReturnTypes[1])
	}
}

func TestMethodSig_MultipleParams(t *testing.T) {
	input := `interface Calculator
  def add(a: Int, b: Int, c: Int): Int
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	iface := program.Declarations[0].(*ast.InterfaceDecl)
	sig := iface.Methods[0]
	if len(sig.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(sig.Params))
	}
}

func TestMethodSig_TrailingCommaParams(t *testing.T) {
	input := `interface Foo
  def bar(a: Int, b: String,): Bool
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	iface := program.Declarations[0].(*ast.InterfaceDecl)
	sig := iface.Methods[0]
	if len(sig.Params) != 2 {
		t.Fatalf("expected 2 params (trailing comma), got %d", len(sig.Params))
	}
}

func TestMethodSig_TrailingCommaReturnTypes(t *testing.T) {
	input := `interface Foo
  def bar: (Int, String,)
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	iface := program.Declarations[0].(*ast.InterfaceDecl)
	sig := iface.Methods[0]
	if len(sig.ReturnTypes) != 2 {
		t.Fatalf("expected 2 return types (trailing comma), got %d", len(sig.ReturnTypes))
	}
}

func TestMethodSig_MissingName(t *testing.T) {
	input := `interface Foo
  def
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for missing method name")
	}
}

func TestMethodSig_MissingCloseParen(t *testing.T) {
	input := `interface Foo
  def bar(a: Int
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for unclosed parentheses")
	}
}

func TestMethodSig_MissingTypeAfterArrow(t *testing.T) {
	input := `interface Foo
  def bar ->
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for missing type after ->")
	}
}

// Field validation tests

func TestValidateNoNewFields_InMethod(t *testing.T) {
	input := `class User
  @name: String

  def initialize(@name: String)
  end

  def set_unknown
    @unknown = "test"
  end
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for new field in non-initialize method")
	}
}

func TestValidateNoNewFields_InIfStatement(t *testing.T) {
	input := `class User
  @name: String

  def initialize(@name: String)
  end

  def check
    if true
      @unknown = "test"
    end
  end
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for new field in if statement")
	}
}

func TestValidateNoNewFields_InWhileLoop(t *testing.T) {
	input := `class User
  @name: String

  def initialize(@name: String)
  end

  def loop
    while true
      @unknown = "test"
    end
  end
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for new field in while loop")
	}
}

func TestValidateNoNewFields_InForLoop(t *testing.T) {
	input := `class User
  @name: String

  def initialize(@name: String)
  end

  def iterate
    for i in items
      @unknown = "test"
    end
  end
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for new field in for loop")
	}
}

func TestValidateNoNewFields_OrAssign(t *testing.T) {
	input := `class User
  @name: String

  def initialize(@name: String)
  end

  def lazy
    @unknown ||= "default"
  end
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected parse error for new field via ||= in method")
	}
}

// Index expression tests

func TestIndexExpr_Range(t *testing.T) {
	input := `def main
  arr = items[1..3]
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	indexExpr, ok := assign.Value.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %T", assign.Value)
	}

	rangeLit, ok := indexExpr.Index.(*ast.RangeLit)
	if !ok {
		t.Fatalf("expected RangeLit index, got %T", indexExpr.Index)
	}

	if rangeLit.Exclusive {
		t.Error("expected inclusive range (..), got exclusive (...)")
	}
}

func TestIndexExpr_Negative(t *testing.T) {
	input := `def main
  last = items[-1]
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

	indexExpr, ok := assign.Value.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %T", assign.Value)
	}

	unary, ok := indexExpr.Index.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr for negative index, got %T", indexExpr.Index)
	}
	if unary.Op != "-" {
		t.Errorf("expected '-' operator, got %q", unary.Op)
	}
}

// Interface and class inheritance tests

func TestInterfaceWithParents(t *testing.T) {
	input := `interface ReadWriter < Reader, Writer
  def close
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	iface := program.Declarations[0].(*ast.InterfaceDecl)
	if len(iface.Parents) != 2 {
		t.Fatalf("expected 2 parent interfaces, got %d", len(iface.Parents))
	}
	if iface.Parents[0] != "Reader" || iface.Parents[1] != "Writer" {
		t.Errorf("expected parents [Reader, Writer], got %v", iface.Parents)
	}
}

func TestClassWithEmbeds(t *testing.T) {
	input := `class AdminUser < User
  def admin?
    true
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	class := program.Declarations[0].(*ast.ClassDecl)
	if len(class.Embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(class.Embeds))
	}
	if class.Embeds[0] != "User" {
		t.Errorf("expected embed 'User', got %q", class.Embeds[0])
	}
}

// AST Position Tracking Tests
// These tests verify that AST nodes have correct Line/Column positions set

func TestASTNodePositions_Ident(t *testing.T) {
	input := `def main
  x = foo
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	ident := assign.Value.(*ast.Ident)

	if ident.Line != 2 {
		t.Errorf("expected Line 2, got %d", ident.Line)
	}
	if ident.Column != 7 {
		t.Errorf("expected Column 7, got %d", ident.Column)
	}
}

func TestASTNodePositions_IntLit(t *testing.T) {
	input := `def main
  x = 42
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	intLit := assign.Value.(*ast.IntLit)

	if intLit.Line != 2 {
		t.Errorf("expected Line 2, got %d", intLit.Line)
	}
	if intLit.Column != 7 {
		t.Errorf("expected Column 7, got %d", intLit.Column)
	}
}

func TestASTNodePositions_FloatLit(t *testing.T) {
	input := `def main
  x = 3.14
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	floatLit := assign.Value.(*ast.FloatLit)

	if floatLit.Line != 2 {
		t.Errorf("expected Line 2, got %d", floatLit.Line)
	}
	if floatLit.Column != 7 {
		t.Errorf("expected Column 7, got %d", floatLit.Column)
	}
}

func TestASTNodePositions_BoolLit(t *testing.T) {
	input := `def main
  x = true
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	boolLit := assign.Value.(*ast.BoolLit)

	if boolLit.Line != 2 {
		t.Errorf("expected Line 2, got %d", boolLit.Line)
	}
	if boolLit.Column != 7 {
		t.Errorf("expected Column 7, got %d", boolLit.Column)
	}
}

func TestASTNodePositions_NilLit(t *testing.T) {
	input := `def main
  x = nil
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	nilLit := assign.Value.(*ast.NilLit)

	if nilLit.Line != 2 {
		t.Errorf("expected Line 2, got %d", nilLit.Line)
	}
	if nilLit.Column != 7 {
		t.Errorf("expected Column 7, got %d", nilLit.Column)
	}
}

func TestASTNodePositions_SymbolLit(t *testing.T) {
	input := `def main
  x = :hello
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	symbolLit := assign.Value.(*ast.SymbolLit)

	if symbolLit.Line != 2 {
		t.Errorf("expected Line 2, got %d", symbolLit.Line)
	}
	if symbolLit.Column != 7 {
		t.Errorf("expected Column 7, got %d", symbolLit.Column)
	}
}

func TestASTNodePositions_ArrayLit(t *testing.T) {
	input := `def main
  x = [1, 2, 3]
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)
	arrayLit := assign.Value.(*ast.ArrayLit)

	if arrayLit.Line != 2 {
		t.Errorf("expected Line 2, got %d", arrayLit.Line)
	}
	if arrayLit.Column != 7 {
		t.Errorf("expected Column 7, got %d", arrayLit.Column)
	}
}

func TestASTNodePositions_LineStart(t *testing.T) {
	// Test identifier at very start of line (column 1)
	input := `def main
foo()
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	exprStmt := fn.Body[0].(*ast.ExprStmt)
	call := exprStmt.Expr.(*ast.CallExpr)
	ident := call.Func.(*ast.Ident)

	if ident.Line != 2 {
		t.Errorf("expected Line 2, got %d", ident.Line)
	}
	if ident.Column != 1 {
		t.Errorf("expected Column 1, got %d", ident.Column)
	}
}

func TestConstDeclaration(t *testing.T) {
	tests := []struct {
		input         string
		expectedName  string
		expectedType  string
		expectedValue any
	}{
		{"const MAX_SIZE = 1024", "MAX_SIZE", "", int64(1024)},
		{"const PI = 3.14", "PI", "", 3.14},
		{`const NAME = "hello"`, "NAME", "", "hello"},
		{"const TIMEOUT: Int64 = 30", "TIMEOUT", "Int64", int64(30)},
		{"const BUFFER_SIZE = 1024 * 2", "BUFFER_SIZE", "", nil}, // computed value
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Declarations) != 1 {
			t.Fatalf("%s: expected 1 declaration, got %d", tt.input, len(program.Declarations))
		}

		constDecl, ok := program.Declarations[0].(*ast.ConstDecl)
		if !ok {
			t.Fatalf("%s: expected ConstDecl, got %T", tt.input, program.Declarations[0])
		}

		if constDecl.Name != tt.expectedName {
			t.Errorf("%s: expected name %q, got %q", tt.input, tt.expectedName, constDecl.Name)
		}

		if constDecl.Type != tt.expectedType {
			t.Errorf("%s: expected type %q, got %q", tt.input, tt.expectedType, constDecl.Type)
		}

		// Check value for simple literals
		if tt.expectedValue != nil {
			switch expected := tt.expectedValue.(type) {
			case int64:
				intLit, ok := constDecl.Value.(*ast.IntLit)
				if !ok {
					t.Errorf("%s: expected IntLit, got %T", tt.input, constDecl.Value)
				} else if intLit.Value != expected {
					t.Errorf("%s: expected value %d, got %d", tt.input, expected, intLit.Value)
				}
			case float64:
				floatLit, ok := constDecl.Value.(*ast.FloatLit)
				if !ok {
					t.Errorf("%s: expected FloatLit, got %T", tt.input, constDecl.Value)
				} else if floatLit.Value != expected {
					t.Errorf("%s: expected value %f, got %f", tt.input, expected, floatLit.Value)
				}
			case string:
				strLit, ok := constDecl.Value.(*ast.StringLit)
				if !ok {
					t.Errorf("%s: expected StringLit, got %T", tt.input, constDecl.Value)
				} else if strLit.Value != expected {
					t.Errorf("%s: expected value %q, got %q", tt.input, expected, strLit.Value)
				}
			}
		}
	}
}

func TestConstDeclaration_MultipleConsts(t *testing.T) {
	input := `const A = 1
const B = 2
const C = 3`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Declarations) != 3 {
		t.Fatalf("expected 3 declarations, got %d", len(program.Declarations))
	}

	names := []string{"A", "B", "C"}
	for i, name := range names {
		constDecl, ok := program.Declarations[i].(*ast.ConstDecl)
		if !ok {
			t.Fatalf("declaration %d: expected ConstDecl, got %T", i, program.Declarations[i])
		}
		if constDecl.Name != name {
			t.Errorf("declaration %d: expected name %q, got %q", i, name, constDecl.Name)
		}
	}
}
