package parser

import (
	"testing"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/lexer"
)

func TestComplexImports(t *testing.T) {
	input := `import gopkg.in/yaml.v2
import github.com/user/project`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(program.Imports))
	}

	expected := []string{
		"gopkg.in/yaml.v2",
		"github.com/user/project",
	}

	for i, path := range expected {
		if program.Imports[i].Path != path {
			t.Errorf("import %d: expected %q, got %q", i, path, program.Imports[i].Path)
		}
	}
}

func TestDivisionEdgeCase(t *testing.T) {
	// x/y should be parsed as binary expression x / y
	input := `def main
  x / y
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	stmt, ok := decl.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", decl.Body[0])
	}
	binExpr, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", stmt.Expr)
	}

	if binExpr.Op != "/" {
		t.Errorf("expected op '/', got %q", binExpr.Op)
	}
}

func TestNotEqualEdgeCase(t *testing.T) {
	// x!=y should be parsed as binary expression x != y
	input := `def main
  x != y
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	stmt, ok := decl.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", decl.Body[0])
	}
	binExpr, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", stmt.Expr)
	}

	if binExpr.Op != "!=" {
		t.Errorf("expected op '!=', got %q", binExpr.Op)
	}
}

func TestBangOperatorWithComparison(t *testing.T) {
	// save! != y → (save()!) != y
	// The ! is now the error unwrap operator, not part of the method name
	input := `def main
  save! != y
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	stmt, ok := decl.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", decl.Body[0])
	}
	binExpr, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", stmt.Expr)
	}

	// Left side should be BangExpr(CallExpr(Ident("save")))
	bangExpr, ok := binExpr.Left.(*ast.BangExpr)
	if !ok {
		t.Fatalf("expected BangExpr, got %T", binExpr.Left)
	}
	callExpr, ok := bangExpr.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr inside BangExpr, got %T", bangExpr.Expr)
	}
	ident, ok := callExpr.Func.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident, got %T", callExpr.Func)
	}
	if ident.Name != "save" {
		t.Errorf("expected ident name 'save', got %q", ident.Name)
	}

	if binExpr.Op != "!=" {
		t.Errorf("expected op '!=', got %q", binExpr.Op)
	}
}

func TestEmptyFunctionBody(t *testing.T) {
	input := `def empty_func
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	if len(decl.Body) != 0 {
		t.Errorf("expected empty body, got %d statements", len(decl.Body))
	}
}

func TestEmptyClassBody(t *testing.T) {
	input := `class Empty
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", program.Declarations[0])
	}
	if len(decl.Methods) != 0 {
		t.Errorf("expected no methods, got %d", len(decl.Methods))
	}
}

func TestEmptyInterfaceBody(t *testing.T) {
	input := `interface Empty
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.InterfaceDecl)
	if !ok {
		t.Fatalf("expected InterfaceDecl, got %T", program.Declarations[0])
	}
	if len(decl.Methods) != 0 {
		t.Errorf("expected no methods, got %d", len(decl.Methods))
	}
}

func TestDeeplyNestedExpressions(t *testing.T) {
	input := `def main
  ((((1 + 2) * 3) - 4) / 5)
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	if len(decl.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(decl.Body))
	}
	stmt, ok := decl.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", decl.Body[0])
	}
	// Just verify it parsed without error - the expression should be a BinaryExpr
	_, ok = stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", stmt.Expr)
	}
}

func TestChainedMethodCalls(t *testing.T) {
	input := `def main
  a.b().c().d()
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	stmt, ok := decl.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", decl.Body[0])
	}
	// Verify it's a call expression
	call, ok := stmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", stmt.Expr)
	}
	// The outermost call should be .d()
	sel, ok := call.Func.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr for d, got %T", call.Func)
	}
	if sel.Sel != "d" {
		t.Errorf("expected selector 'd', got %q", sel.Sel)
	}
}

func TestIfLetPattern(t *testing.T) {
	input := `def main
  if let user = find_user(5)
    puts user.name
  end
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	ifStmt, ok := decl.Body[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", decl.Body[0])
	}
	if ifStmt.AssignName == "" {
		t.Fatal("expected AssignName, got empty")
	}
	if ifStmt.AssignName != "user" {
		t.Errorf("expected assign name 'user', got %q", ifStmt.AssignName)
	}
}

func TestNilCoalesceChain(t *testing.T) {
	input := `def main
  a ?? b ?? c
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	stmt, ok := decl.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", decl.Body[0])
	}
	// Should parse as (a ?? b) ?? c - NilCoalesceExpr
	nilCoal, ok := stmt.Expr.(*ast.NilCoalesceExpr)
	if !ok {
		t.Fatalf("expected NilCoalesceExpr, got %T", stmt.Expr)
	}
	if nilCoal.Right == nil {
		t.Fatal("expected Right, got nil")
	}
}

func TestSafeNavigationChain(t *testing.T) {
	input := `def main
  user&.address&.city
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	stmt, ok := decl.Body[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", decl.Body[0])
	}
	// Outer expression should be SafeNavExpr
	safeNav, ok := stmt.Expr.(*ast.SafeNavExpr)
	if !ok {
		t.Fatalf("expected SafeNavExpr, got %T", stmt.Expr)
	}
	if safeNav.Selector != "city" {
		t.Errorf("expected selector 'city', got %q", safeNav.Selector)
	}
}

func TestTupleUnpacking(t *testing.T) {
	input := `def main
  user, ok = find_user(5)
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	assign, ok := decl.Body[0].(*ast.MultiAssignStmt)
	if !ok {
		t.Fatalf("expected MultiAssignStmt, got %T", decl.Body[0])
	}
	if len(assign.Names) != 2 {
		t.Errorf("expected 2 targets, got %d", len(assign.Names))
	}
}

func TestForInRange(t *testing.T) {
	input := `def main
  for i in 0..10
    puts i
  end
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	forStmt, ok := decl.Body[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", decl.Body[0])
	}
	rangeLit, ok := forStmt.Iterable.(*ast.RangeLit)
	if !ok {
		t.Fatalf("expected RangeLit, got %T", forStmt.Iterable)
	}
	if rangeLit.Exclusive {
		t.Error("expected inclusive range (..)")
	}
}

func TestExclusiveRange(t *testing.T) {
	input := `def main
  for i in 0...10
    puts i
  end
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	forStmt, ok := decl.Body[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", decl.Body[0])
	}
	rangeLit, ok := forStmt.Iterable.(*ast.RangeLit)
	if !ok {
		t.Fatalf("expected RangeLit, got %T", forStmt.Iterable)
	}
	if !rangeLit.Exclusive {
		t.Error("expected exclusive range (...)")
	}
}

func TestCaseTypeSwitch(t *testing.T) {
	input := `def main
  case_type x
  when String
    puts "string"
  when Int
    puts "int"
  end
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	caseStmt, ok := decl.Body[0].(*ast.CaseTypeStmt)
	if !ok {
		t.Fatalf("expected CaseTypeStmt, got %T", decl.Body[0])
	}
	if len(caseStmt.WhenClauses) != 2 {
		t.Errorf("expected 2 when clauses, got %d", len(caseStmt.WhenClauses))
	}
}

func TestSymbolLiteral(t *testing.T) {
	input := `def main
  x = :ok
  y = :not_found
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	if len(decl.Body) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(decl.Body))
	}

	assign, ok := decl.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", decl.Body[0])
	}
	sym, ok := assign.Value.(*ast.SymbolLit)
	if !ok {
		t.Fatalf("expected SymbolLit, got %T", assign.Value)
	}
	if sym.Value != "ok" {
		t.Errorf("expected symbol 'ok', got %q", sym.Value)
	}
}

func TestDeferStatementParsing(t *testing.T) {
	input := `def main
  defer f.close()
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	deferStmt, ok := decl.Body[0].(*ast.DeferStmt)
	if !ok {
		t.Fatalf("expected DeferStmt, got %T", decl.Body[0])
	}
	if deferStmt.Call == nil {
		t.Fatal("expected Call in defer, got nil")
	}
}

func TestPanicStatementParsing(t *testing.T) {
	input := `def main
  panic "something went wrong"
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	panicStmt, ok := decl.Body[0].(*ast.PanicStmt)
	if !ok {
		t.Fatalf("expected PanicStmt, got %T", decl.Body[0])
	}
	strLit, ok := panicStmt.Message.(*ast.StringLit)
	if !ok {
		t.Fatalf("expected StringLit in panic, got %T", panicStmt.Message)
	}
	if strLit.Value != "something went wrong" {
		t.Errorf("expected panic message, got %q", strLit.Value)
	}
}

func TestTernaryExpressionParsing(t *testing.T) {
	input := `def main
  x = a > b ? a : b
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	assignStmt, ok := decl.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", decl.Body[0])
	}
	ternary, ok := assignStmt.Value.(*ast.TernaryExpr)
	if !ok {
		t.Fatalf("expected TernaryExpr, got %T", assignStmt.Value)
	}

	// Check condition is a > b
	cond, ok := ternary.Condition.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr condition, got %T", ternary.Condition)
	}
	if cond.Op != ">" {
		t.Errorf("expected condition op '>', got %q", cond.Op)
	}

	// Check then is identifier 'a'
	thenIdent, ok := ternary.Then.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident for then, got %T", ternary.Then)
	}
	if thenIdent.Name != "a" {
		t.Errorf("expected then 'a', got %q", thenIdent.Name)
	}

	// Check else is identifier 'b'
	elseIdent, ok := ternary.Else.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident for else, got %T", ternary.Else)
	}
	if elseIdent.Name != "b" {
		t.Errorf("expected else 'b', got %q", elseIdent.Name)
	}
}

func TestTernaryNestedRightAssociativity(t *testing.T) {
	// a ? b ? 1 : 2 : 3 should parse as a ? (b ? 1 : 2) : 3
	input := `def main
  x = a ? b ? 1 : 2 : 3
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	assignStmt, ok := decl.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", decl.Body[0])
	}
	outer, ok := assignStmt.Value.(*ast.TernaryExpr)
	if !ok {
		t.Fatalf("expected TernaryExpr, got %T", assignStmt.Value)
	}

	// Outer condition should be 'a'
	outerCond, ok := outer.Condition.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident for outer condition, got %T", outer.Condition)
	}
	if outerCond.Name != "a" {
		t.Errorf("expected outer condition 'a', got %q", outerCond.Name)
	}

	// Outer then should be another ternary (b ? 1 : 2)
	inner, ok := outer.Then.(*ast.TernaryExpr)
	if !ok {
		t.Fatalf("expected nested TernaryExpr for outer then, got %T", outer.Then)
	}

	// Inner condition should be 'b'
	innerCond, ok := inner.Condition.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident for inner condition, got %T", inner.Condition)
	}
	if innerCond.Name != "b" {
		t.Errorf("expected inner condition 'b', got %q", innerCond.Name)
	}

	// Inner then should be 1
	innerThen, ok := inner.Then.(*ast.IntLit)
	if !ok {
		t.Fatalf("expected IntLit for inner then, got %T", inner.Then)
	}
	if innerThen.Value != 1 {
		t.Errorf("expected inner then 1, got %d", innerThen.Value)
	}

	// Inner else should be 2
	innerElse, ok := inner.Else.(*ast.IntLit)
	if !ok {
		t.Fatalf("expected IntLit for inner else, got %T", inner.Else)
	}
	if innerElse.Value != 2 {
		t.Errorf("expected inner else 2, got %d", innerElse.Value)
	}

	// Outer else should be 3
	outerElse, ok := outer.Else.(*ast.IntLit)
	if !ok {
		t.Fatalf("expected IntLit for outer else, got %T", outer.Else)
	}
	if outerElse.Value != 3 {
		t.Errorf("expected outer else 3, got %d", outerElse.Value)
	}
}

func TestTernaryWithPredicateMethod(t *testing.T) {
	// empty? ? a : b - the ? suffix on method name shouldn't confuse ternary parsing
	input := `def main
  x = empty? ? "none" : "some"
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	assignStmt, ok := decl.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", decl.Body[0])
	}
	ternary, ok := assignStmt.Value.(*ast.TernaryExpr)
	if !ok {
		t.Fatalf("expected TernaryExpr, got %T", assignStmt.Value)
	}

	// Condition should be identifier 'empty?'
	cond, ok := ternary.Condition.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident for condition, got %T", ternary.Condition)
	}
	if cond.Name != "empty?" {
		t.Errorf("expected condition 'empty?', got %q", cond.Name)
	}
}

func TestSymbolToProcParsing(t *testing.T) {
	input := `def main
  x = &:upcase
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	assignStmt, ok := decl.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", decl.Body[0])
	}
	stp, ok := assignStmt.Value.(*ast.SymbolToProcExpr)
	if !ok {
		t.Fatalf("expected SymbolToProcExpr, got %T", assignStmt.Value)
	}

	if stp.Method != "upcase" {
		t.Errorf("expected method 'upcase', got %q", stp.Method)
	}
}

func TestSymbolToProcAsArgument(t *testing.T) {
	input := `def main
  result = names.map(&:upcase)
end`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	decl, ok := program.Declarations[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Declarations[0])
	}
	assignStmt, ok := decl.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", decl.Body[0])
	}

	call, ok := assignStmt.Value.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", assignStmt.Value)
	}

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(call.Args))
	}

	stp, ok := call.Args[0].(*ast.SymbolToProcExpr)
	if !ok {
		t.Fatalf("expected SymbolToProcExpr argument, got %T", call.Args[0])
	}

	if stp.Method != "upcase" {
		t.Errorf("expected method 'upcase', got %q", stp.Method)
	}
}
