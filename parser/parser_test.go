package parser

import (
	"strings"
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

func TestMultipleReturnValues(t *testing.T) {
	input := `def foo()
  return 1, true, "hello"
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Declarations[0].(*ast.FuncDecl)
	retStmt := fn.Body[0].(*ast.ReturnStmt)

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

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

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

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

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

	fn := program.Declarations[0].(*ast.FuncDecl)
	exprStmt := fn.Body[0].(*ast.ExprStmt)

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

	fn := program.Declarations[0].(*ast.FuncDecl)

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

	fn := program.Declarations[0].(*ast.FuncDecl)
	deferStmt := fn.Body[0].(*ast.DeferStmt)

	if deferStmt.Call == nil {
		t.Fatal("expected Call in DeferStmt")
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

	fn := program.Declarations[0].(*ast.FuncDecl)
	assign := fn.Body[0].(*ast.AssignStmt)

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
    puts x
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
    puts v
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
    puts "hello"
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
    puts x
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
    puts x
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
      puts x
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
  arr.each {|x| puts x }
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
  arr.each_with_index {|v, i| puts v }
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
  3.times {|| puts "hello" }
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
