// Package parser implements the Rugby parser.
package parser

import (
	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/token"
)

func (p *Parser) parseStatement() ast.Statement {
	line := p.curToken.Line
	switch p.curToken.Type {
	case token.IF:
		return p.parseIfStmt()
	case token.UNLESS:
		return p.parseUnlessStmt()
	case token.CASE:
		return p.parseCaseStmt()
	case token.WHILE:
		return p.parseWhileStmt()
	case token.FOR:
		return p.parseForStmt()
	case token.BREAK:
		return p.parseBreakStmt()
	case token.NEXT:
		return p.parseNextStmt()
	case token.RETURN:
		return p.parseReturnStmt()
	case token.RAISE:
		return p.parseRaiseStmt()
	case token.DEFER:
		return p.parseDeferStmt()
	case token.IDENT:
		// Check for assignment: ident = expr or ident : Type = expr
		if p.peekTokenIs(token.ASSIGN) || p.peekTokenIs(token.COLON) {
			return p.parseAssignStmt()
		}
		// Check for or-assignment: ident ||= expr
		if p.peekTokenIs(token.ORASSIGN) {
			return p.parseOrAssignStmt()
		}
		// Check for compound assignment: ident += expr, ident -= expr, etc.
		if p.peekTokenIs(token.PLUSASSIGN) || p.peekTokenIs(token.MINUSASSIGN) ||
			p.peekTokenIs(token.STARASSIGN) || p.peekTokenIs(token.SLASHASSIGN) {
			return p.parseCompoundAssignStmt()
		}
	case token.AT:
		// Check if this is assignment: @name = expr
		// Pattern: AT IDENT ASSIGN
		if p.peekTokenIs(token.IDENT) && p.peekTokenAfterIs(token.ASSIGN) {
			return p.parseInstanceVarAssign()
		}
		// Check if this is or-assignment: @name ||= expr
		// Pattern: AT IDENT ORASSIGN
		if p.peekTokenIs(token.IDENT) && p.peekTokenAfterIs(token.ORASSIGN) {
			return p.parseInstanceVarOrAssign()
		}
		// Otherwise fall through to expression parsing
	}

	// Default: expression statement
	expr := p.parseExpression(lowest)
	if expr != nil {
		// Check for block: expr do |params| ... end  OR  expr {|params| ... }
		if p.peekTokenIs(token.DO) || p.peekTokenIs(token.LBRACE) {
			expr = p.parseBlockCall(expr)
		}
		p.nextToken() // move past expression
		// Check for statement modifier (e.g., "puts x unless valid?")
		cond, isUnless := p.parseStatementModifier()
		p.skipNewlines()
		return &ast.ExprStmt{Expr: expr, Condition: cond, IsUnless: isUnless, Line: line}
	}
	return nil
}

func (p *Parser) parseAssignStmt() *ast.AssignStmt {
	line := p.curToken.Line
	name := p.curToken.Literal
	p.nextToken() // consume ident

	// Check for optional type annotation: x : Type = value or x : Type? = value
	var typeAnnotation string
	if p.curTokenIs(token.COLON) {
		p.nextToken() // consume ':'
		if !p.curTokenIs(token.IDENT) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ':'")
			return nil
		}
		typeAnnotation = p.parseTypeName()
	}

	if !p.curTokenIs(token.ASSIGN) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '=' in assignment")
		return nil
	}
	p.nextToken() // consume '='

	value := p.parseExpression(lowest)

	// Check for block: expr.method do |x| ... end  OR  expr.method {|x| ... }
	if p.peekTokenIs(token.DO) || p.peekTokenIs(token.LBRACE) {
		value = p.parseBlockCall(value)
	}

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.AssignStmt{Name: name, Type: typeAnnotation, Value: value, Line: line}
}

func (p *Parser) parseIfStmt() *ast.IfStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'if'

	stmt := &ast.IfStmt{Line: line}

	// Check for assignment-in-condition pattern: if (name = expr)
	if p.curTokenIs(token.LPAREN) {
		// Peek ahead to see if it's (ident = expr) pattern
		p.nextToken() // consume '('
		if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.ASSIGN) {
			// Assignment pattern detected
			stmt.AssignName = p.curToken.Literal
			p.nextToken() // consume ident
			p.nextToken() // consume '='
			stmt.AssignExpr = p.parseExpression(lowest)
			p.nextToken() // move past expression
			if !p.curTokenIs(token.RPAREN) {
				p.errorAt(p.curToken.Line, p.curToken.Column, "expected ')' after assignment expression")
				return nil
			}
			p.nextToken() // consume ')'
		} else {
			// Regular parenthesized condition - parse as grouped expression
			// We already consumed '(', so parse the inner expression
			stmt.Cond = p.parseExpression(lowest)
			p.nextToken() // move past expression
			if p.curTokenIs(token.RPAREN) {
				p.nextToken() // consume ')'
			}
		}
	} else {
		stmt.Cond = p.parseExpression(lowest)
		p.nextToken() // move past condition
	}
	p.skipNewlines()

	// Parse 'then' body
	for !p.curTokenIs(token.ELSIF) && !p.curTokenIs(token.ELSE) &&
		!p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.ELSIF) || p.curTokenIs(token.ELSE) || p.curTokenIs(token.END) {
			break
		}
		if s := p.parseStatement(); s != nil {
			stmt.Then = append(stmt.Then, s)
		}
	}

	// Parse elsif clauses
	for p.curTokenIs(token.ELSIF) {
		p.nextToken() // consume 'elsif'
		clause := ast.ElseIfClause{}
		clause.Cond = p.parseExpression(lowest)
		p.nextToken() // move past condition
		p.skipNewlines()

		for !p.curTokenIs(token.ELSIF) && !p.curTokenIs(token.ELSE) &&
			!p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
			p.skipNewlines()
			if p.curTokenIs(token.ELSIF) || p.curTokenIs(token.ELSE) || p.curTokenIs(token.END) {
				break
			}
			if s := p.parseStatement(); s != nil {
				clause.Body = append(clause.Body, s)
			}
		}
		stmt.ElseIfs = append(stmt.ElseIfs, clause)
	}

	// Parse else clause
	if p.curTokenIs(token.ELSE) {
		p.nextToken() // consume 'else'
		p.skipNewlines()

		for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
			p.skipNewlines()
			if p.curTokenIs(token.END) {
				break
			}
			if s := p.parseStatement(); s != nil {
				stmt.Else = append(stmt.Else, s)
			}
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close if")
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

func (p *Parser) parseUnlessStmt() *ast.IfStmt {
	// Note: elsif is not supported with unless (matches Ruby behavior)
	line := p.curToken.Line
	p.nextToken() // consume 'unless'

	stmt := &ast.IfStmt{IsUnless: true, Line: line}

	// Parse condition (unless doesn't support assignment pattern)
	stmt.Cond = p.parseExpression(lowest)
	p.nextToken() // move past condition
	p.skipNewlines()

	// Parse 'then' body (executed when condition is false)
	// Note: elsif should not appear here, but if it does, treat it as an error
	for !p.curTokenIs(token.ELSE) && !p.curTokenIs(token.END) &&
		!p.curTokenIs(token.EOF) && !p.curTokenIs(token.ELSIF) {
		p.skipNewlines()
		if p.curTokenIs(token.ELSE) || p.curTokenIs(token.END) || p.curTokenIs(token.ELSIF) {
			break
		}
		if s := p.parseStatement(); s != nil {
			stmt.Then = append(stmt.Then, s)
		}
	}

	// If we encountered elsif, that's an error
	if p.curTokenIs(token.ELSIF) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "elsif not supported with unless")
		// Skip to end to recover
		for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
			p.nextToken()
		}
	}

	// Parse else clause (optional)
	if p.curTokenIs(token.ELSE) {
		p.nextToken() // consume 'else'
		p.skipNewlines()
		for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
			p.skipNewlines()
			if p.curTokenIs(token.END) {
				break
			}
			if s := p.parseStatement(); s != nil {
				stmt.Else = append(stmt.Else, s)
			}
		}
	}

	// Expect 'end'
	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' after unless block")
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

func (p *Parser) parseCaseStmt() *ast.CaseStmt {
	line := p.curToken.Line
	doc := p.leadingComments(line)
	p.nextToken() // consume 'case'

	stmt := &ast.CaseStmt{Line: line, Doc: doc}

	// Parse optional subject expression (before newline/when)
	// If next token is NEWLINE or WHEN, it's case without subject
	if !p.curTokenIs(token.NEWLINE) && !p.curTokenIs(token.WHEN) {
		stmt.Subject = p.parseExpression(lowest)
		p.nextToken() // move past expression
	}
	p.skipNewlines()

	// Parse when clauses
	for p.curTokenIs(token.WHEN) {
		p.nextToken() // consume 'when'

		clause := ast.WhenClause{}

		// Parse one or more comma-separated values
		clause.Values = append(clause.Values, p.parseExpression(lowest))
		p.nextToken() // move past first value

		// Parse additional values separated by commas
		for p.curTokenIs(token.COMMA) {
			p.nextToken() // consume ','
			clause.Values = append(clause.Values, p.parseExpression(lowest))
			p.nextToken() // move past value
		}

		// Validate that when clause has at least one value
		if len(clause.Values) == 0 {
			p.errorAt(p.curToken.Line, p.curToken.Column, "'when' requires at least one value")
			return nil
		}
		p.skipNewlines()

		// Parse body until next WHEN, ELSE, or END
		for !p.curTokenIs(token.WHEN) && !p.curTokenIs(token.ELSE) &&
			!p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
			p.skipNewlines()
			if p.curTokenIs(token.WHEN) || p.curTokenIs(token.ELSE) || p.curTokenIs(token.END) {
				break
			}
			if s := p.parseStatement(); s != nil {
				clause.Body = append(clause.Body, s)
			}
		}
		stmt.WhenClauses = append(stmt.WhenClauses, clause)
	}

	// Validate that case has at least one when clause
	if len(stmt.WhenClauses) == 0 {
		p.errorAt(p.curToken.Line, p.curToken.Column, "case statement requires at least one 'when' clause")
		return nil
	}

	// Parse optional else clause
	if p.curTokenIs(token.ELSE) {
		p.nextToken() // consume 'else'
		p.skipNewlines()

		for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
			p.skipNewlines()
			if p.curTokenIs(token.END) {
				break
			}
			if s := p.parseStatement(); s != nil {
				stmt.Else = append(stmt.Else, s)
			}
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close case")
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

func (p *Parser) parseWhileStmt() *ast.WhileStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'while'

	stmt := &ast.WhileStmt{Line: line}
	stmt.Cond = p.parseExpression(lowest)
	p.nextToken() // move past condition
	p.skipNewlines()

	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}
		if s := p.parseStatement(); s != nil {
			stmt.Body = append(stmt.Body, s)
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close while")
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

func (p *Parser) parseForStmt() *ast.ForStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'for'

	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected variable name after 'for'")
		return nil
	}

	stmt := &ast.ForStmt{Var: p.curToken.Literal, Line: line}
	p.nextToken() // consume variable name

	if !p.curTokenIs(token.IN) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'in' after loop variable")
		return nil
	}
	p.nextToken() // consume 'in'

	stmt.Iterable = p.parseExpression(lowest)
	p.nextToken() // move past iterable expression
	p.skipNewlines()

	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}
		if s := p.parseStatement(); s != nil {
			stmt.Body = append(stmt.Body, s)
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close for")
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

// parseStatementModifier checks for and parses trailing if/unless modifiers
// Returns (condition, isUnless) - condition is nil if no modifier present
func (p *Parser) parseStatementModifier() (ast.Expression, bool) {
	// Check for trailing if/unless before newline
	if p.curTokenIs(token.IF) || p.curTokenIs(token.UNLESS) {
		isUnless := p.curTokenIs(token.UNLESS)
		p.nextToken() // consume 'if' or 'unless'
		cond := p.parseExpression(lowest)
		p.nextToken() // move past condition
		return cond, isUnless
	}
	return nil, false
}

func (p *Parser) parseBreakStmt() *ast.BreakStmt {
	p.nextToken() // consume 'break'
	stmt := &ast.BreakStmt{}
	stmt.Condition, stmt.IsUnless = p.parseStatementModifier()
	p.skipNewlines()
	return stmt
}

func (p *Parser) parseNextStmt() *ast.NextStmt {
	p.nextToken() // consume 'next'
	stmt := &ast.NextStmt{}
	stmt.Condition, stmt.IsUnless = p.parseStatementModifier()
	p.skipNewlines()
	return stmt
}

func (p *Parser) parseReturnStmt() *ast.ReturnStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'return'

	stmt := &ast.ReturnStmt{Line: line}

	// Check if there's a return value (not newline, end, or modifier keyword)
	if !p.curTokenIs(token.NEWLINE) && !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) &&
		!p.curTokenIs(token.IF) && !p.curTokenIs(token.UNLESS) {
		stmt.Values = append(stmt.Values, p.parseExpression(lowest))
		p.nextToken() // move past expression

		// Parse additional return values
		for p.curTokenIs(token.COMMA) {
			p.nextToken() // consume ','
			// Allow trailing comma
			if p.curTokenIs(token.NEWLINE) || p.curTokenIs(token.END) || p.curTokenIs(token.EOF) {
				break
			}
			stmt.Values = append(stmt.Values, p.parseExpression(lowest))
			p.nextToken() // move past expression
		}
	}

	// Check for statement modifier
	stmt.Condition, stmt.IsUnless = p.parseStatementModifier()
	p.skipNewlines()
	return stmt
}

func (p *Parser) parseRaiseStmt() *ast.RaiseStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'raise'

	stmt := &ast.RaiseStmt{Line: line}

	// raise requires a message expression
	if p.curTokenIs(token.NEWLINE) || p.curTokenIs(token.END) || p.curTokenIs(token.EOF) {
		p.errorAt(line, p.curToken.Column, "expected expression after raise")
		return stmt
	}

	stmt.Message = p.parseExpression(lowest)
	p.nextToken() // move past expression
	p.skipNewlines()
	return stmt
}

func (p *Parser) parseDeferStmt() *ast.DeferStmt {
	deferLine := p.curToken.Line
	deferCol := p.curToken.Column
	p.nextToken() // consume 'defer'

	// Parse the expression that should be a call
	expr := p.parseExpression(lowest)
	if expr == nil {
		p.errorAt(deferLine, deferCol, "expected expression after defer")
		return nil
	}

	// The expression must be a call or a selector (which we'll convert to a call)
	var call *ast.CallExpr

	switch e := expr.(type) {
	case *ast.CallExpr:
		call = e
	case *ast.SelectorExpr:
		// defer resp.Body.Close becomes defer resp.Body.Close()
		call = &ast.CallExpr{Func: e, Args: nil}
	case *ast.Ident:
		// defer close becomes defer close()
		call = &ast.CallExpr{Func: e, Args: nil}
	default:
		p.errorAt(deferLine, deferCol, "defer requires a callable expression")
		return nil
	}

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.DeferStmt{Call: call}
}

func (p *Parser) parseInstanceVarAssign() *ast.InstanceVarAssign {
	p.nextToken() // consume '@'
	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected identifier after '@'")
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	if !p.curTokenIs(token.ASSIGN) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '=' after instance variable")
		return nil
	}
	p.nextToken() // consume '='

	value := p.parseExpression(lowest)

	// Check for block after expression
	if p.peekTokenIs(token.DO) || p.peekTokenIs(token.LBRACE) {
		value = p.parseBlockCall(value)
	}

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.InstanceVarAssign{Name: name, Value: value}
}

func (p *Parser) parseOrAssignStmt() *ast.OrAssignStmt {
	name := p.curToken.Literal
	p.nextToken() // consume ident

	if !p.curTokenIs(token.ORASSIGN) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '||=' in or-assignment")
		return nil
	}
	p.nextToken() // consume '||='

	value := p.parseExpression(lowest)

	// Check for block after expression
	if p.peekTokenIs(token.DO) || p.peekTokenIs(token.LBRACE) {
		value = p.parseBlockCall(value)
	}

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.OrAssignStmt{Name: name, Value: value}
}

func (p *Parser) parseCompoundAssignStmt() *ast.CompoundAssignStmt {
	name := p.curToken.Literal
	p.nextToken() // consume ident

	// Determine the operator
	var op string
	switch p.curToken.Type {
	case token.PLUSASSIGN:
		op = "+"
	case token.MINUSASSIGN:
		op = "-"
	case token.STARASSIGN:
		op = "*"
	case token.SLASHASSIGN:
		op = "/"
	default:
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected compound assignment operator")
		return nil
	}
	p.nextToken() // consume operator

	value := p.parseExpression(lowest)

	// Check for block after expression
	if p.peekTokenIs(token.DO) || p.peekTokenIs(token.LBRACE) {
		value = p.parseBlockCall(value)
	}

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.CompoundAssignStmt{Name: name, Op: op, Value: value}
}

func (p *Parser) parseInstanceVarOrAssign() *ast.InstanceVarOrAssign {
	p.nextToken() // consume '@'
	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected identifier after '@'")
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	if !p.curTokenIs(token.ORASSIGN) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '||=' after instance variable")
		return nil
	}
	p.nextToken() // consume '||='

	value := p.parseExpression(lowest)

	// Check for block after expression
	if p.peekTokenIs(token.DO) || p.peekTokenIs(token.LBRACE) {
		value = p.parseBlockCall(value)
	}

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.InstanceVarOrAssign{Name: name, Value: value}
}
