// Package parser implements the Rugby parser.
package parser

import (
	"fmt"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/token"
)

func (p *Parser) parseDescribeStmt() *ast.DescribeStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'describe'

	// Expect string for name
	if !p.curTokenIs(token.STRING) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected string after 'describe'")
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume string

	// Expect 'do'
	if !p.curTokenIs(token.DO) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'do' after describe name")
		return nil
	}
	p.nextToken() // consume 'do'
	p.skipNewlines()

	stmt := &ast.DescribeStmt{Name: name, Line: line}

	// Parse body until 'end'
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}

		switch p.curToken.Type {
		case token.IT:
			if it := p.parseItStmt(); it != nil {
				stmt.Body = append(stmt.Body, it)
			}
		case token.DESCRIBE:
			if desc := p.parseDescribeStmt(); desc != nil {
				desc.Parent = stmt
				stmt.Body = append(stmt.Body, desc)
			}
		case token.BEFORE:
			if before := p.parseBeforeStmt(); before != nil {
				stmt.Body = append(stmt.Body, before)
			}
		case token.AFTER:
			if after := p.parseAfterStmt(); after != nil {
				stmt.Body = append(stmt.Body, after)
			}
		default:
			p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("unexpected token %s in describe block, expected 'it', 'describe', 'before', 'after', or 'end'", p.curToken.Type))
			p.nextToken()
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close describe block")
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

// parseItStmt parses a test case: it "description" do |t| ... end
func (p *Parser) parseItStmt() *ast.ItStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'it'

	// Expect string for description
	if !p.curTokenIs(token.STRING) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected string after 'it'")
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume string

	// Expect 'do'
	if !p.curTokenIs(token.DO) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'do' after it description")
		return nil
	}
	p.nextToken() // consume 'do'

	// Parse optional block parameter |t|
	if p.curTokenIs(token.PIPE) {
		p.nextToken() // consume '|'
		if p.curTokenIs(token.IDENT) {
			paramName := p.curToken.Literal
			if paramName != "t" {
				p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("test block parameter must be 't', got %q", paramName))
			}
			p.nextToken()
		}
		// Handle additional parameters or commas
		for !p.curTokenIs(token.PIPE) && !p.curTokenIs(token.NEWLINE) && !p.curTokenIs(token.EOF) {
			p.nextToken()
		}
		if !p.curTokenIs(token.PIPE) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected closing '|' for block parameter")
			return nil
		}
		p.nextToken() // consume closing '|'
	}
	p.skipNewlines()

	stmt := &ast.ItStmt{Name: name, Line: line}

	// Parse body until 'end'
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}
		if s := p.parseStatement(); s != nil {
			stmt.Body = append(stmt.Body, s)
		} else {
			p.nextToken() // error recovery: always advance to prevent infinite loop
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close it block")
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

// parseTestStmt parses a standalone test: test "name" do |t| ... end
func (p *Parser) parseTestStmt() *ast.TestStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'test'

	// Expect string for name
	if !p.curTokenIs(token.STRING) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected string after 'test'")
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume string

	// Expect 'do'
	if !p.curTokenIs(token.DO) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'do' after test name")
		return nil
	}
	p.nextToken() // consume 'do'

	// Parse optional block parameter |t|
	if p.curTokenIs(token.PIPE) {
		p.nextToken() // consume '|'
		if p.curTokenIs(token.IDENT) {
			paramName := p.curToken.Literal
			if paramName != "t" {
				p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("test block parameter must be 't', got %q", paramName))
			}
			p.nextToken()
		}
		for !p.curTokenIs(token.PIPE) && !p.curTokenIs(token.NEWLINE) && !p.curTokenIs(token.EOF) {
			p.nextToken()
		}
		if !p.curTokenIs(token.PIPE) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected closing '|' for block parameter")
			return nil
		}
		p.nextToken() // consume closing '|'
	}
	p.skipNewlines()

	stmt := &ast.TestStmt{Name: name, Line: line}

	// Parse body until 'end'
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}
		if s := p.parseStatement(); s != nil {
			stmt.Body = append(stmt.Body, s)
		} else {
			p.nextToken() // error recovery: always advance to prevent infinite loop
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close test block")
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

// parseTableStmt parses a table-driven test: table "name" do |tt| ... end
func (p *Parser) parseTableStmt() *ast.TableStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'table'

	// Expect string for name
	if !p.curTokenIs(token.STRING) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected string after 'table'")
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume string

	// Expect 'do'
	if !p.curTokenIs(token.DO) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'do' after table name")
		return nil
	}
	p.nextToken() // consume 'do'

	// Parse optional block parameter |tt|
	if p.curTokenIs(token.PIPE) {
		p.nextToken() // consume '|'
		if p.curTokenIs(token.IDENT) {
			paramName := p.curToken.Literal
			if paramName != "tt" {
				p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("table block parameter must be 'tt', got %q", paramName))
			}
			p.nextToken()
		}
		// Handle additional parameters or commas
		for !p.curTokenIs(token.PIPE) && !p.curTokenIs(token.NEWLINE) && !p.curTokenIs(token.EOF) {
			p.nextToken()
		}
		if p.curTokenIs(token.PIPE) {
			p.nextToken() // consume closing '|'
		}
	}
	p.skipNewlines()

	stmt := &ast.TableStmt{Name: name, Line: line}

	// Parse body until 'end'
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}
		if s := p.parseStatement(); s != nil {
			stmt.Body = append(stmt.Body, s)
		} else {
			p.nextToken() // error recovery: always advance to prevent infinite loop
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close table block")
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

// parseBeforeStmt parses a before hook: before do |t| ... end
func (p *Parser) parseBeforeStmt() *ast.BeforeStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'before'

	// Expect 'do'
	if !p.curTokenIs(token.DO) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'do' after 'before'")
		return nil
	}
	p.nextToken() // consume 'do'

	// Parse optional block parameter |t|
	if p.curTokenIs(token.PIPE) {
		p.nextToken() // consume '|'
		if p.curTokenIs(token.IDENT) {
			paramName := p.curToken.Literal
			if paramName != "t" {
				p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("before block parameter must be 't', got %q", paramName))
			}
			p.nextToken()
		}
		// Handle additional parameters or commas
		for !p.curTokenIs(token.PIPE) && !p.curTokenIs(token.NEWLINE) && !p.curTokenIs(token.EOF) {
			p.nextToken()
		}
		if p.curTokenIs(token.PIPE) {
			p.nextToken() // consume closing '|'
		}
	}
	p.skipNewlines()

	stmt := &ast.BeforeStmt{Line: line}

	// Parse body until 'end'
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}
		if s := p.parseStatement(); s != nil {
			stmt.Body = append(stmt.Body, s)
		} else {
			p.nextToken() // error recovery: always advance to prevent infinite loop
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close before block")
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

// parseAfterStmt parses an after hook: after do |t, ctx| ... end
func (p *Parser) parseAfterStmt() *ast.AfterStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'after'

	// Expect 'do'
	if !p.curTokenIs(token.DO) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'do' after 'after'")
		return nil
	}
	p.nextToken() // consume 'do'

	// Parse optional block parameters |t| or |t, ctx|
	if p.curTokenIs(token.PIPE) {
		p.nextToken() // consume '|'
		if p.curTokenIs(token.IDENT) {
			paramName := p.curToken.Literal
			if paramName != "t" {
				p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("after block first parameter must be 't', got %q", paramName))
			}
			p.nextToken()
		}
		// Handle additional parameters (ctx) or commas
		for !p.curTokenIs(token.PIPE) && !p.curTokenIs(token.NEWLINE) && !p.curTokenIs(token.EOF) {
			p.nextToken()
		}
		if p.curTokenIs(token.PIPE) {
			p.nextToken() // consume closing '|'
		}
	}
	p.skipNewlines()

	stmt := &ast.AfterStmt{Line: line}

	// Parse body until 'end'
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}
		if s := p.parseStatement(); s != nil {
			stmt.Body = append(stmt.Body, s)
		} else {
			p.nextToken() // error recovery: always advance to prevent infinite loop
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close after block")
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}
