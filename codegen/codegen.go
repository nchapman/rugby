package codegen

import (
	"fmt"
	"go/format"
	"rugby/ast"
	"strings"
)

type Generator struct {
	buf strings.Builder
}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(program *ast.Program) (string, error) {
	// Package declaration
	g.buf.WriteString("package main\n\n")

	// Imports
	if len(program.Imports) > 0 {
		g.buf.WriteString("import (\n")
		for _, imp := range program.Imports {
			g.buf.WriteString(fmt.Sprintf("\t%q\n", imp.Path))
		}
		g.buf.WriteString(")\n\n")
	}

	// Declarations
	for _, decl := range program.Declarations {
		g.genStatement(decl)
		g.buf.WriteString("\n")
	}

	// Format the output
	source := g.buf.String()
	formatted, err := format.Source([]byte(source))
	if err != nil {
		return source, nil // Return unformatted if gofmt fails
	}
	return string(formatted), nil
}

func (g *Generator) genStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.FuncDecl:
		g.genFuncDecl(s)
	case *ast.ExprStmt:
		g.genExprStmt(s)
	}
}

func (g *Generator) genFuncDecl(fn *ast.FuncDecl) {
	g.buf.WriteString(fmt.Sprintf("func %s() {\n", fn.Name))
	for _, stmt := range fn.Body {
		g.buf.WriteString("\t")
		g.genStatement(stmt)
		g.buf.WriteString("\n")
	}
	g.buf.WriteString("}\n")
}

func (g *Generator) genExprStmt(stmt *ast.ExprStmt) {
	g.genExpr(stmt.Expr)
}

func (g *Generator) genExpr(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		g.genCallExpr(e)
	case *ast.StringLit:
		g.buf.WriteString(fmt.Sprintf("%q", e.Value))
	case *ast.Ident:
		g.buf.WriteString(e.Name)
	}
}

func (g *Generator) genCallExpr(call *ast.CallExpr) {
	// Map Rugby builtins to Go
	funcName := call.Func
	switch funcName {
	case "puts":
		funcName = "fmt.Println"
	}

	g.buf.WriteString(funcName)
	g.buf.WriteString("(")
	for i, arg := range call.Args {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.genExpr(arg)
	}
	g.buf.WriteString(")")
}
