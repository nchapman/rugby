package codegen

import (
	"fmt"
	"strings"

	"github.com/nchapman/rugby/ast"
)

// genTestAssertion generates code for assert.* and require.* calls
// assert.equal(t, expected, actual) -> test.AssertInstance.Equal(t, expected, actual)
// require.not_nil(t, value) -> test.RequireInstance.NotNil(t, value)
func (g *Generator) genTestAssertion(call *ast.CallExpr, pkg, method string) {
	g.needsTestImport = true

	// Map snake_case method names to PascalCase
	methodName := snakeToPascalWithAcronyms(method)

	// Determine the instance name
	var instance string
	if pkg == "assert" {
		instance = "test.AssertInstance"
	} else {
		instance = "test.RequireInstance"
	}

	g.buf.WriteString(fmt.Sprintf("%s.%s(", instance, methodName))
	for i, arg := range call.Args {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.genExpr(arg)
	}
	g.buf.WriteString(")")
}

// sanitizeTestName converts a test description to a valid Go test function name
// Examples: "String#to_i?" -> "StringToI", "math/add" -> "MathAdd"
func sanitizeTestName(name string) string {
	// Replace special characters with underscores (to create word boundaries)
	// or remove them entirely
	replacer := strings.NewReplacer(
		"#", "_",
		"?", "",
		"!", "",
		"/", "_",
		" ", "_",
		"-", "_",
		".", "_",
		"::", "_",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"<", "",
		">", "",
		"=", "_Eq_",
		"+", "_Plus_",
		"*", "_Star_",
		"@", "_At_",
	)
	sanitized := replacer.Replace(name)

	// Clean up multiple consecutive underscores
	for strings.Contains(sanitized, "__") {
		sanitized = strings.ReplaceAll(sanitized, "__", "_")
	}
	sanitized = strings.Trim(sanitized, "_")

	// Convert to PascalCase
	return snakeToPascalWithAcronyms(sanitized)
}

// genDescribeStmt generates a test function from a describe block
// describe "String#to_i?" do ... end -> func TestStringToI(t *testing.T) { ... }
func (g *Generator) genDescribeStmt(desc *ast.DescribeStmt) {
	g.needsTestingImport = true
	clear(g.vars)
	g.vars["t"] = "*testing.T"

	funcName := "Test" + sanitizeTestName(desc.Name)
	g.buf.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", funcName))
	g.indent++

	// Collect before and after hooks
	var beforeStmts, afterStmts []ast.Statement
	var testStmts []ast.Statement

	for _, stmt := range desc.Body {
		switch s := stmt.(type) {
		case *ast.BeforeStmt:
			beforeStmts = append(beforeStmts, s.Body...)
		case *ast.AfterStmt:
			afterStmts = append(afterStmts, s.Body...)
		default:
			testStmts = append(testStmts, stmt)
		}
	}

	// Generate test body
	for _, stmt := range testStmts {
		switch s := stmt.(type) {
		case *ast.ItStmt:
			g.genItStmtWithHooks(s, beforeStmts, afterStmts)
		case *ast.DescribeStmt:
			g.genNestedDescribe(s, beforeStmts, afterStmts)
		default:
			g.genStatement(stmt)
		}
	}

	g.indent--
	g.buf.WriteString("}\n")
}

// genNestedDescribe generates a nested t.Run for a describe block
// Note: parentBefore and parentAfter are currently unused per MVP spec (hooks don't cascade),
// but kept for future enhancement.
func (g *Generator) genNestedDescribe(desc *ast.DescribeStmt, _, _ []ast.Statement) {
	g.writeIndent()
	g.buf.WriteString(fmt.Sprintf("t.Run(%q, func(t *testing.T) {\n", desc.Name))
	g.indent++

	// Collect this describe's before and after hooks
	var beforeStmts, afterStmts []ast.Statement
	var testStmts []ast.Statement

	for _, stmt := range desc.Body {
		switch s := stmt.(type) {
		case *ast.BeforeStmt:
			beforeStmts = append(beforeStmts, s.Body...)
		case *ast.AfterStmt:
			afterStmts = append(afterStmts, s.Body...)
		default:
			testStmts = append(testStmts, stmt)
		}
	}

	// Generate test body (note: per MVP spec, hooks don't cascade)
	for _, stmt := range testStmts {
		switch s := stmt.(type) {
		case *ast.ItStmt:
			g.genItStmtWithHooks(s, beforeStmts, afterStmts)
		case *ast.DescribeStmt:
			g.genNestedDescribe(s, beforeStmts, afterStmts)
		default:
			g.genStatement(stmt)
		}
	}

	g.indent--
	g.writeIndent()
	g.buf.WriteString("})")
}

// genItStmtWithHooks generates a t.Run with optional before/after hooks
func (g *Generator) genItStmtWithHooks(it *ast.ItStmt, before, after []ast.Statement) {
	g.writeIndent()
	g.buf.WriteString(fmt.Sprintf("t.Run(%q, func(t *testing.T) {\n", it.Name))
	g.indent++

	// Generate before hooks
	for _, stmt := range before {
		g.genStatement(stmt)
	}

	// Generate after hooks as t.Cleanup if present
	if len(after) > 0 {
		g.writeIndent()
		g.buf.WriteString("t.Cleanup(func() {\n")
		g.indent++
		for _, stmt := range after {
			g.genStatement(stmt)
		}
		g.indent--
		g.writeIndent()
		g.buf.WriteString("})\n")
	}

	// Generate test body
	for _, stmt := range it.Body {
		g.genStatement(stmt)
	}

	g.indent--
	g.writeIndent()
	g.buf.WriteString("})")
}

// genItStmt generates a t.Run statement (for standalone it blocks, though unusual)
func (g *Generator) genItStmt(it *ast.ItStmt) {
	g.genItStmtWithHooks(it, nil, nil)
}

// genTestStmt generates a standalone test function
// test "math/add" do |t| ... end -> func TestMathAdd(t *testing.T) { ... }
func (g *Generator) genTestStmt(test *ast.TestStmt) {
	g.needsTestingImport = true
	clear(g.vars)
	g.vars["t"] = "*testing.T"

	funcName := "Test" + sanitizeTestName(test.Name)
	g.buf.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", funcName))
	g.indent++

	for _, stmt := range test.Body {
		g.genStatement(stmt)
	}

	g.indent--
	g.buf.WriteString("}\n")
}

// genTableStmt generates a table-driven test
// This is more complex and will be implemented in a follow-up for the MVP
func (g *Generator) genTableStmt(table *ast.TableStmt) {
	g.needsTestingImport = true
	clear(g.vars)
	g.vars["t"] = "*testing.T"

	funcName := "Test" + sanitizeTestName(table.Name) + "_Table"
	g.buf.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", funcName))
	g.indent++

	// For now, just generate the body statements
	// Full table test parsing requires more sophisticated analysis of tt.case/tt.run
	for _, stmt := range table.Body {
		g.genStatement(stmt)
	}

	g.indent--
	g.buf.WriteString("}\n")
}
