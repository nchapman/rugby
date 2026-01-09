package codegen

import (
	"fmt"
	"go/format"
	"strings"
	"unicode"

	"rugby/ast"
)

type Generator struct {
	buf    strings.Builder
	indent int
	vars   map[string]bool // track declared variables
}

func New() *Generator {
	return &Generator{vars: make(map[string]bool)}
}

func (g *Generator) Generate(program *ast.Program) (string, error) {
	// Package declaration
	g.buf.WriteString("package main\n\n")

	// Imports
	if len(program.Imports) > 0 {
		g.buf.WriteString("import (\n")
		for _, imp := range program.Imports {
			if imp.Alias != "" {
				g.buf.WriteString(fmt.Sprintf("\t%s %q\n", imp.Alias, imp.Path))
			} else {
				g.buf.WriteString(fmt.Sprintf("\t%q\n", imp.Path))
			}
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

func (g *Generator) writeIndent() {
	for i := 0; i < g.indent; i++ {
		g.buf.WriteString("\t")
	}
}

func (g *Generator) genStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.FuncDecl:
		g.genFuncDecl(s)
	case *ast.ExprStmt:
		g.writeIndent()
		g.genExpr(s.Expr)
		g.buf.WriteString("\n")
	case *ast.AssignStmt:
		g.genAssignStmt(s)
	case *ast.IfStmt:
		g.genIfStmt(s)
	case *ast.WhileStmt:
		g.genWhileStmt(s)
	case *ast.ReturnStmt:
		g.genReturnStmt(s)
	case *ast.DeferStmt:
		g.genDeferStmt(s)
	}
}

func (g *Generator) genFuncDecl(fn *ast.FuncDecl) {
	g.vars = make(map[string]bool) // reset vars for each function

	// Mark parameters as declared variables
	for _, param := range fn.Params {
		g.vars[param.Name] = true
	}

	// Generate function signature
	g.buf.WriteString(fmt.Sprintf("func %s(", fn.Name))
	for i, param := range fn.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.buf.WriteString(fmt.Sprintf("%s interface{}", param.Name))
	}
	g.buf.WriteString(")")

	// Generate return type(s) if specified
	if len(fn.ReturnTypes) == 1 {
		g.buf.WriteString(" ")
		g.buf.WriteString(mapType(fn.ReturnTypes[0]))
	} else if len(fn.ReturnTypes) > 1 {
		g.buf.WriteString(" (")
		for i, rt := range fn.ReturnTypes {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.buf.WriteString(mapType(rt))
		}
		g.buf.WriteString(")")
	}

	g.buf.WriteString(" {\n")

	g.indent++
	for _, stmt := range fn.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.buf.WriteString("}\n")
}

func (g *Generator) genAssignStmt(s *ast.AssignStmt) {
	g.writeIndent()
	g.buf.WriteString(s.Name)
	if g.vars[s.Name] {
		g.buf.WriteString(" = ")
	} else {
		g.buf.WriteString(" := ")
		g.vars[s.Name] = true
	}
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

func (g *Generator) genIfStmt(s *ast.IfStmt) {
	g.writeIndent()
	g.buf.WriteString("if ")
	g.genExpr(s.Cond)
	g.buf.WriteString(" {\n")

	g.indent++
	for _, stmt := range s.Then {
		g.genStatement(stmt)
	}
	g.indent--

	for _, elsif := range s.ElseIfs {
		g.writeIndent()
		g.buf.WriteString("} else if ")
		g.genExpr(elsif.Cond)
		g.buf.WriteString(" {\n")

		g.indent++
		for _, stmt := range elsif.Body {
			g.genStatement(stmt)
		}
		g.indent--
	}

	if len(s.Else) > 0 {
		g.writeIndent()
		g.buf.WriteString("} else {\n")

		g.indent++
		for _, stmt := range s.Else {
			g.genStatement(stmt)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genWhileStmt(s *ast.WhileStmt) {
	g.writeIndent()
	g.buf.WriteString("for ")
	g.genExpr(s.Cond)
	g.buf.WriteString(" {\n")

	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genReturnStmt(s *ast.ReturnStmt) {
	g.writeIndent()
	g.buf.WriteString("return")
	if len(s.Values) > 0 {
		g.buf.WriteString(" ")
		for i, val := range s.Values {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.genExpr(val)
		}
	}
	g.buf.WriteString("\n")
}

func (g *Generator) genExpr(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		g.genCallExpr(e)
	case *ast.SelectorExpr:
		g.genSelectorExpr(e)
	case *ast.StringLit:
		g.buf.WriteString(fmt.Sprintf("%q", e.Value))
	case *ast.IntLit:
		g.buf.WriteString(fmt.Sprintf("%d", e.Value))
	case *ast.FloatLit:
		g.buf.WriteString(fmt.Sprintf("%g", e.Value))
	case *ast.BoolLit:
		if e.Value {
			g.buf.WriteString("true")
		} else {
			g.buf.WriteString("false")
		}
	case *ast.Ident:
		g.buf.WriteString(e.Name)
	case *ast.BinaryExpr:
		g.genBinaryExpr(e)
	case *ast.UnaryExpr:
		g.genUnaryExpr(e)
	}
}

func (g *Generator) genBinaryExpr(e *ast.BinaryExpr) {
	g.buf.WriteString("(")
	g.genExpr(e.Left)
	g.buf.WriteString(" ")
	g.buf.WriteString(e.Op)
	g.buf.WriteString(" ")
	g.genExpr(e.Right)
	g.buf.WriteString(")")
}

func (g *Generator) genUnaryExpr(e *ast.UnaryExpr) {
	g.buf.WriteString(e.Op)
	g.genExpr(e.Expr)
}

func (g *Generator) genCallExpr(call *ast.CallExpr) {
	// Generate the function expression
	switch fn := call.Func.(type) {
	case *ast.Ident:
		// Simple function call - check for builtins
		funcName := fn.Name
		switch funcName {
		case "puts":
			funcName = "fmt.Println"
		}
		// Don't transform local function names - they stay as defined
		g.buf.WriteString(funcName)
	case *ast.SelectorExpr:
		// Method/package call like http.Get or resp.Body.Close
		// snake_case transformation only applies here (Go interop)
		g.genSelectorExpr(fn)
	default:
		// Fallback for other expressions
		g.genExpr(call.Func)
	}

	g.buf.WriteString("(")
	for i, arg := range call.Args {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.genExpr(arg)
	}
	g.buf.WriteString(")")
}

func (g *Generator) genSelectorExpr(sel *ast.SelectorExpr) {
	g.genExpr(sel.X)
	g.buf.WriteString(".")
	// Apply snake_case to CamelCase mapping
	g.buf.WriteString(snakeToCamel(sel.Sel))
}

func (g *Generator) genDeferStmt(s *ast.DeferStmt) {
	g.writeIndent()
	g.buf.WriteString("defer ")
	g.genCallExpr(s.Call)
	g.buf.WriteString("\n")
}

// snakeToCamel converts snake_case to CamelCase for Go interop
// Examples: read_all -> ReadAll, new_request -> NewRequest
func snakeToCamel(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	capitalizeNext := true

	for _, r := range s {
		if r == '_' {
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// mapType converts Rugby type names to Go type names
func mapType(rubyType string) string {
	switch rubyType {
	case "Int":
		return "int"
	case "Int64":
		return "int64"
	case "Float":
		return "float64"
	case "String":
		return "string"
	case "Bool":
		return "bool"
	case "Bytes":
		return "[]byte"
	default:
		return rubyType // pass through unknown types (e.g., user-defined)
	}
}
