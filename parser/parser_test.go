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
		{"def foo(a : any)\nend", []string{"a"}},
		{"def foo(a : any, b : any)\nend", []string{"a", "b"}},
		{"def foo(a : any, b : any, c : any)\nend", []string{"a", "b", "c"}},
		{"def foo(a : any, b : any,)\nend", []string{"a", "b"}}, // trailing comma allowed
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
	input := `def foo(a : any, a : any)
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
		{"def foo() -> Int\nend", []string{"Int"}},
		{"def foo(a : any, b : any) -> String\nend", []string{"String"}},
		{"def foo(x : any) -> Bool\nend", []string{"Bool"}},
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

func TestRaiseStatement(t *testing.T) {
	input := `def main
  raise "something went wrong"
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}

	raiseStmt, ok := fn.Body[0].(*ast.RaiseStmt)
	if !ok {
		t.Fatalf("expected RaiseStmt, got %T", fn.Body[0])
	}

	strLit, ok := raiseStmt.Message.(*ast.StringLit)
	if !ok {
		t.Fatalf("expected StringLit message, got %T", raiseStmt.Message)
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
	// Note: `x!` parses as identifier "x!" (valid method name)
	// We test with a literal followed by ! to trigger the error
	input := `def main
  y = 42!
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	// Should have an error because ! can only follow a call expression
	if len(p.Errors()) == 0 {
		t.Fatal("expected parse error for '!' on non-call expression (literal)")
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

func TestEachBlock(t *testing.T) {
	input := `def main
  arr.each do |x|
    puts(x)
  end
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

	if call.Block == nil {
		t.Fatal("expected block, got nil")
	}

	if len(call.Block.Params) != 1 || call.Block.Params[0] != "x" {
		t.Errorf("expected block params [x], got %v", call.Block.Params)
	}

	if len(call.Block.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(call.Block.Body))
	}
}

func TestEachWithIndexBlock(t *testing.T) {
	input := `def main
  arr.each_with_index do |v, i|
    puts(v)
  end
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

	if call.Block == nil {
		t.Fatal("expected block, got nil")
	}

	if len(call.Block.Params) != 2 {
		t.Fatalf("expected 2 block params, got %d", len(call.Block.Params))
	}

	if call.Block.Params[0] != "v" {
		t.Errorf("expected first param 'v', got %q", call.Block.Params[0])
	}

	if call.Block.Params[1] != "i" {
		t.Errorf("expected second param 'i', got %q", call.Block.Params[1])
	}

	if len(call.Block.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(call.Block.Body))
	}
}

func TestMapBlock(t *testing.T) {
	input := `def main
  result = arr.map do |x|
    x * 2
  end
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

	if call.Block == nil {
		t.Fatal("expected block, got nil")
	}

	if len(call.Block.Params) != 1 || call.Block.Params[0] != "x" {
		t.Errorf("expected block params [x], got %v", call.Block.Params)
	}
}

func TestBlockWithNoParams(t *testing.T) {
	input := `def main
  items.each do ||
    puts("hello")
  end
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

	if call.Block == nil {
		t.Fatal("expected block, got nil")
	}

	if len(call.Block.Params) != 0 {
		t.Errorf("expected 0 block params, got %v", call.Block.Params)
	}
}

func TestGenericBlockOnAnyMethod(t *testing.T) {
	input := `def main
  items.custom_method do |x, y, z|
    puts(x)
  end
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

	if call.Block == nil {
		t.Fatal("expected block, got nil")
	}

	if len(call.Block.Params) != 3 {
		t.Fatalf("expected 3 block params, got %d", len(call.Block.Params))
	}

	expectedParams := []string{"x", "y", "z"}
	for i, expected := range expectedParams {
		if call.Block.Params[i] != expected {
			t.Errorf("expected param %d to be %q, got %q", i, expected, call.Block.Params[i])
		}
	}
}

func TestDuplicateBlockParameter(t *testing.T) {
	input := `def main
  items.each do |x, x|
    puts(x)
  end
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Fatal("expected parser error for duplicate block parameter")
	}

	errMsg := p.Errors()[0]
	if !strings.Contains(errMsg, "duplicate block parameter") {
		t.Errorf("expected error about duplicate block parameter, got: %s", errMsg)
	}
}

func TestNestedBlocks(t *testing.T) {
	input := `def main
  matrix.each do |row|
    row.each do |x|
      puts(x)
    end
  end
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

	if outerCall.Block == nil {
		t.Fatal("expected outer block, got nil")
	}

	if len(outerCall.Block.Params) != 1 || outerCall.Block.Params[0] != "row" {
		t.Errorf("expected outer block params [row], got %v", outerCall.Block.Params)
	}

	// Inner block should be in the body
	if len(outerCall.Block.Body) != 1 {
		t.Fatalf("expected 1 statement in outer block, got %d", len(outerCall.Block.Body))
	}

	innerExprStmt, ok := outerCall.Block.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected inner ExprStmt, got %T", outerCall.Block.Body[0])
	}

	innerCall, ok := innerExprStmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected inner CallExpr, got %T", innerExprStmt.Expr)
	}

	if innerCall.Block == nil {
		t.Fatal("expected inner block, got nil")
	}

	if len(innerCall.Block.Params) != 1 || innerCall.Block.Params[0] != "x" {
		t.Errorf("expected inner block params [x], got %v", innerCall.Block.Params)
	}
}

func TestBraceBlockSyntax(t *testing.T) {
	input := `def main
  arr.each {|x| puts(x) }
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

	if sel.Sel != "each" {
		t.Errorf("expected selector 'each', got %q", sel.Sel)
	}

	if call.Block == nil {
		t.Fatal("expected block, got nil")
	}

	if len(call.Block.Params) != 1 || call.Block.Params[0] != "x" {
		t.Errorf("expected block params [x], got %v", call.Block.Params)
	}

	if len(call.Block.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(call.Block.Body))
	}
}

func TestBraceBlockMultipleParams(t *testing.T) {
	input := `def main
  arr.each_with_index {|v, i| puts(v) }
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	exprStmt := fn.Body[0].(*ast.ExprStmt)
	call := exprStmt.Expr.(*ast.CallExpr)

	if call.Block == nil {
		t.Fatal("expected block, got nil")
	}

	if len(call.Block.Params) != 2 {
		t.Fatalf("expected 2 block params, got %d", len(call.Block.Params))
	}

	if call.Block.Params[0] != "v" || call.Block.Params[1] != "i" {
		t.Errorf("expected block params [v, i], got %v", call.Block.Params)
	}
}

func TestBraceBlockNoParams(t *testing.T) {
	input := `def main
  3.times {|| puts("hello") }
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	exprStmt := fn.Body[0].(*ast.ExprStmt)
	call := exprStmt.Expr.(*ast.CallExpr)

	if call.Block == nil {
		t.Fatal("expected block, got nil")
	}

	if len(call.Block.Params) != 0 {
		t.Errorf("expected 0 block params, got %v", call.Block.Params)
	}
}

func TestBraceBlockWithMethodArgs(t *testing.T) {
	input := `def main
  arr.reduce(0) {|acc, x| acc + x }
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	exprStmt := fn.Body[0].(*ast.ExprStmt)
	call := exprStmt.Expr.(*ast.CallExpr)

	// Should have one argument (the initial value)
	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}

	if call.Block == nil {
		t.Fatal("expected block, got nil")
	}

	if len(call.Block.Params) != 2 {
		t.Fatalf("expected 2 block params, got %d", len(call.Block.Params))
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
  def add(a : any, b : any) -> Int
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

  def value -> Int
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
  def add(a : any, a : any)
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
  def initialize(name : any, age : any)
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
  def initialize(name : any)
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
  def initialize(name : any)
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
  x : Int = 5
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
	input := `def add(a : Int, b : Int) -> Int
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
	input := `def foo(a : Int, b : any, c : String)
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

	// a : Int
	if fn.Params[0].Name != "a" || fn.Params[0].Type != "Int" {
		t.Errorf("param 0: expected a:Int, got %s:%s", fn.Params[0].Name, fn.Params[0].Type)
	}

	// b : any
	if fn.Params[1].Name != "b" || fn.Params[1].Type != "any" {
		t.Errorf("param 1: expected b:any, got %s:%s", fn.Params[1].Name, fn.Params[1].Type)
	}

	// c : String
	if fn.Params[2].Name != "c" || fn.Params[2].Type != "String" {
		t.Errorf("param 2: expected c:String, got %s:%s", fn.Params[2].Name, fn.Params[2].Type)
	}
}

func TestClassFieldTypeInference(t *testing.T) {
	input := `class User
  @name : String
  @age : Int

  def initialize(name : String, age : Int)
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
  def initialize(@x : Int, @y : Int)
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
  def initialize(name : String)
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

  def add(a : Int, b : Int) -> Int
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

  def with_name(n : any)
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
  def initialize(x : Int, y : Int)
    @x = x
    @y = y
  end

  def ==(other : any)
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
  def speak -> String
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
  def read(n : Int) -> String
  def write(data : String) -> Int
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
	input := `pub def add(a : Int, b : Int) -> Int
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
  def initialize(name : String)
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
  def initialize(name : String)
    @name = name
  end

  pub def get_name -> String
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
  def speak -> String
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

func TestPubDefInNonPubClassIsError(t *testing.T) {
	input := `class User
  pub def greet
  end
end`

	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Error("expected error for 'pub def' in non-pub class")
	}

	if !strings.Contains(p.Errors()[0], "pub def") || !strings.Contains(p.Errors()[0], "non-pub class") {
		t.Errorf("expected error about 'pub def' in non-pub class, got: %s", p.Errors()[0])
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
	input := `def find(id : Int?) -> User?
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
  x : Int? = nil
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
	input := `def greet(name : String)
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
  def initialize(n : Int)
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
	input := `[1, 2, 3].each do |x|
  puts(x)
end`

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

	if call.Block == nil {
		t.Fatal("expected block, got nil")
	}

	if len(call.Block.Params) != 1 || call.Block.Params[0] != "x" {
		t.Errorf("expected block params [x], got %v", call.Block.Params)
	}
}

func TestBareScriptMixedOrdering(t *testing.T) {
	input := `puts("start")

def helper(x : Int) -> Int
  x * 2
end

result = helper(5)
puts(result)

class Counter
  def initialize(n : Int)
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
  def initialize(name : String)
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
