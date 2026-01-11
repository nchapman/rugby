package parser

import (
	"testing"

	"rugby/ast"
	"rugby/lexer"
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

func TestBangMethodEdgeCase(t *testing.T) {
	// save! != y
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

	// Left side should be Ident(save!)
	ident, ok := binExpr.Left.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident, got %T", binExpr.Left)
	}
	if ident.Name != "save!" {
		t.Errorf("expected ident name 'save!', got %q", ident.Name)
	}

	if binExpr.Op != "!=" {
		t.Errorf("expected op '!=', got %q", binExpr.Op)
	}
}
