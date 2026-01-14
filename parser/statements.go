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
	case token.CASETYPE:
		return p.parseCaseTypeStmt()
	case token.WHILE:
		return p.parseWhileStmt()
	case token.UNTIL:
		return p.parseUntilStmt()
	case token.FOR:
		return p.parseForStmt()
	case token.BREAK:
		return p.parseBreakStmt()
	case token.NEXT:
		return p.parseNextStmt()
	case token.RETURN:
		return p.parseReturnStmt()
	case token.PANIC:
		return p.parsePanicStmt()
	case token.DEFER:
		return p.parseDeferStmt()
	case token.GO:
		return p.parseGoStmt()
	case token.SELECT:
		return p.parseSelectStmt()
	case token.CONCURRENTLY:
		return p.parseConcurrentlyStmt()
	case token.IDENT:
		// Check for multi-assignment: ident, ident = expr
		if p.peekTokenIs(token.COMMA) {
			return p.parseMultiAssignStmt()
		}
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
		// Check for compound assignment: @name += expr, @name -= expr, etc.
		// Pattern: AT IDENT PLUSASSIGN/MINUSASSIGN/STARASSIGN/SLASHASSIGN
		if p.peekTokenIs(token.IDENT) && (p.peekTokenAfterIs(token.PLUSASSIGN) ||
			p.peekTokenAfterIs(token.MINUSASSIGN) || p.peekTokenAfterIs(token.STARASSIGN) ||
			p.peekTokenAfterIs(token.SLASHASSIGN)) {
			return p.parseInstanceVarCompoundAssign()
		}
		// Otherwise fall through to expression parsing
	}

	// Default: expression statement
	expr := p.parseExpression(lowest)
	if expr != nil {
		// Handle blocks and method chaining (inline and multi-line)
		expr = p.parseBlocksAndChaining(expr)

		// Check for selector assignment: obj.field = value
		// This must be handled before nextToken() since we need to see the ASSIGN
		if selExpr, ok := expr.(*ast.SelectorExpr); ok && p.peekTokenIs(token.ASSIGN) {
			p.nextToken() // move to '='
			p.nextToken() // move past '=' to value
			value := p.parseExpression(lowest)
			if value == nil {
				p.errorAt(p.curToken.Line, p.curToken.Column, "expected expression after '='")
				return nil
			}
			p.nextToken() // move past value
			p.skipNewlines()
			return &ast.SelectorAssignStmt{
				Object: selExpr.X,
				Field:  selExpr.Sel,
				Value:  value,
				Line:   line,
			}
		}

		p.nextToken() // move past expression

		// Check for loop modifier (e.g., "puts x while cond" or "puts x until done")
		if loopStmt := p.parseLoopModifier(&ast.ExprStmt{Expr: expr, Line: line}); loopStmt != nil {
			return loopStmt
		}

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

	// Handle blocks and method chaining (inline and multi-line)
	value = p.parseBlocksAndChaining(value)

	// Check for inline type annotation: x = value : Type
	// This is an alternative to: x : Type = value
	if typeAnnotation == "" && p.peekTokenIs(token.COLON) {
		p.nextToken() // move to ':'
		p.nextToken() // move past ':'
		if !p.curTokenIs(token.IDENT) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ':'")
			return nil
		}
		typeAnnotation = p.parseTypeName()
	} else {
		p.nextToken() // move past expression
	}
	p.skipNewlines()

	return &ast.AssignStmt{Name: name, Type: typeAnnotation, Value: value, Line: line}
}

// parseMultiAssignStmt parses tuple unpacking: val, ok = expr
func (p *Parser) parseMultiAssignStmt() *ast.MultiAssignStmt {
	line := p.curToken.Line
	names := []string{p.curToken.Literal}

	// Collect all names separated by commas
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume current ident or comma
		p.nextToken() // consume comma, move to next ident
		if !p.curTokenIs(token.IDENT) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected identifier in multi-assignment")
			return nil
		}
		names = append(names, p.curToken.Literal)
	}

	p.nextToken() // move past last ident

	if !p.curTokenIs(token.ASSIGN) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '=' in multi-assignment")
		return nil
	}
	p.nextToken() // consume '='

	value := p.parseExpression(lowest)

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.MultiAssignStmt{Names: names, Value: value, Line: line}
}

func (p *Parser) parseIfStmt() *ast.IfStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'if'

	stmt := &ast.IfStmt{Line: line}

	// Check for 'if let' pattern: if let name = expr
	if p.curTokenIs(token.LET) {
		p.nextToken() // consume 'let'
		if !p.curTokenIs(token.IDENT) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected identifier after 'let'")
			return nil
		}
		stmt.AssignName = p.curToken.Literal
		p.nextToken() // consume ident
		if !p.curTokenIs(token.ASSIGN) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected '=' after identifier in 'if let'")
			return nil
		}
		p.nextToken() // consume '='
		stmt.AssignExpr = p.parseExpression(lowest)
		p.nextToken() // move past expression
	} else if p.curTokenIs(token.LPAREN) && p.isAssignmentPattern() {
		// Assignment-in-condition pattern: if (name = expr)
		p.nextToken() // consume '('
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
		} else {
			p.nextToken() // error recovery: always advance
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
			} else {
				p.nextToken() // error recovery: always advance
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
			} else {
				p.nextToken() // error recovery: always advance
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
		} else {
			p.nextToken() // error recovery: always advance
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
			} else {
				p.nextToken() // error recovery: always advance
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
			} else {
				p.nextToken() // error recovery: always advance
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
			} else {
				p.nextToken() // error recovery: always advance
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

func (p *Parser) parseCaseTypeStmt() *ast.CaseTypeStmt {
	line := p.curToken.Line
	doc := p.leadingComments(line)
	p.nextToken() // consume 'case_type'

	stmt := &ast.CaseTypeStmt{Line: line, Doc: doc}

	// Parse required subject expression
	if p.curTokenIs(token.NEWLINE) || p.curTokenIs(token.WHEN) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "case_type requires a subject expression")
		return nil
	}
	stmt.Subject = p.parseExpression(lowest)
	p.nextToken() // move past expression
	p.skipNewlines()

	// Parse when clauses with type patterns
	for p.curTokenIs(token.WHEN) {
		p.nextToken() // consume 'when'

		clause := ast.TypeWhenClause{}

		// Parse type name (String, Int, etc.)
		if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected type name in 'when' clause")
			return nil
		}
		clause.Type = p.parseTypeName()
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
			} else {
				p.nextToken() // error recovery: always advance
			}
		}
		stmt.WhenClauses = append(stmt.WhenClauses, clause)
	}

	// Validate that case_type has at least one when clause
	if len(stmt.WhenClauses) == 0 {
		p.errorAt(p.curToken.Line, p.curToken.Column, "case_type requires at least one 'when' clause")
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
			} else {
				p.nextToken() // error recovery: always advance
			}
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close case_type")
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
		} else {
			p.nextToken() // error recovery: always advance
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

func (p *Parser) parseUntilStmt() *ast.UntilStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'until'

	stmt := &ast.UntilStmt{Line: line}
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
		} else {
			p.nextToken() // error recovery: always advance
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close until")
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
		} else {
			p.nextToken() // error recovery: always advance
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

// parseLoopModifier checks for trailing while/until modifiers and transforms
// the statement into a WhileStmt or UntilStmt with the original statement as
// the loop body. Returns nil if no loop modifier is present.
// e.g., "puts x while cond" → WhileStmt { Cond: cond, Body: [ExprStmt{puts x}] }
// Note: Loop modifiers cannot be combined with if/unless modifiers.
func (p *Parser) parseLoopModifier(stmt ast.Statement) ast.Statement {
	if p.curTokenIs(token.WHILE) {
		line := p.curToken.Line
		p.nextToken() // consume 'while'
		cond := p.parseExpression(lowest)
		p.nextToken() // move past condition
		p.skipNewlines()
		return &ast.WhileStmt{Cond: cond, Body: []ast.Statement{stmt}, Line: line}
	}
	if p.curTokenIs(token.UNTIL) {
		line := p.curToken.Line
		p.nextToken() // consume 'until'
		cond := p.parseExpression(lowest)
		p.nextToken() // move past condition
		p.skipNewlines()
		return &ast.UntilStmt{Cond: cond, Body: []ast.Statement{stmt}, Line: line}
	}
	return nil
}

func (p *Parser) parseBreakStmt() *ast.BreakStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'break'
	stmt := &ast.BreakStmt{Line: line}
	stmt.Condition, stmt.IsUnless = p.parseStatementModifier()
	p.skipNewlines()
	return stmt
}

func (p *Parser) parseNextStmt() *ast.NextStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'next'
	stmt := &ast.NextStmt{Line: line}
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

func (p *Parser) parsePanicStmt() *ast.PanicStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'panic'

	stmt := &ast.PanicStmt{Line: line}

	// panic requires a message expression
	if p.curTokenIs(token.NEWLINE) || p.curTokenIs(token.END) || p.curTokenIs(token.EOF) {
		p.errorAt(line, p.curToken.Column, "expected expression after panic")
		return stmt
	}

	stmt.Message = p.parseExpression(lowest)
	p.nextToken() // move past expression

	// Check for statement modifier (panic "msg" if cond)
	stmt.Condition, stmt.IsUnless = p.parseStatementModifier()
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

	return &ast.DeferStmt{Call: call, Line: deferLine}
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

	// Handle blocks and method chaining (inline and multi-line)
	value = p.parseBlocksAndChaining(value)

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.InstanceVarAssign{Name: name, Value: value}
}

func (p *Parser) parseOrAssignStmt() *ast.OrAssignStmt {
	line := p.curToken.Line
	name := p.curToken.Literal
	p.nextToken() // consume ident

	if !p.curTokenIs(token.ORASSIGN) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '||=' in or-assignment")
		return nil
	}
	p.nextToken() // consume '||='

	value := p.parseExpression(lowest)

	// Handle blocks and method chaining (inline and multi-line)
	value = p.parseBlocksAndChaining(value)

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.OrAssignStmt{Name: name, Value: value, Line: line}
}

func (p *Parser) parseCompoundAssignStmt() ast.Statement {
	line := p.curToken.Line
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

	// Handle blocks and method chaining (inline and multi-line)
	value = p.parseBlocksAndChaining(value)

	p.nextToken() // move past expression

	stmt := &ast.CompoundAssignStmt{Name: name, Op: op, Value: value, Line: line}

	// Check for loop modifier (e.g., "counter += 1 until counter == 3")
	if loopStmt := p.parseLoopModifier(stmt); loopStmt != nil {
		return loopStmt
	}

	p.skipNewlines()
	return stmt
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

	// Handle blocks and method chaining (inline and multi-line)
	value = p.parseBlocksAndChaining(value)

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.InstanceVarOrAssign{Name: name, Value: value}
}

// parseInstanceVarCompoundAssign parses: @name += expr, @name -= expr, etc.
func (p *Parser) parseInstanceVarCompoundAssign() *ast.InstanceVarCompoundAssign {
	line := p.curToken.Line
	p.nextToken() // consume '@'
	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected identifier after '@'")
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

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

	// Handle blocks and method chaining (inline and multi-line)
	value = p.parseBlocksAndChaining(value)

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.InstanceVarCompoundAssign{Name: name, Op: op, Value: value, Line: line}
}

// parseGoStmt parses: go func_call() or go do ... end
func (p *Parser) parseGoStmt() *ast.GoStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'go'

	stmt := &ast.GoStmt{Line: line}

	// Check for block form: go do ... end
	if p.curTokenIs(token.DO) {
		stmt.Block = p.parseBlock(token.END)
		p.skipNewlines()
		return stmt
	}

	// Otherwise parse the expression to call
	expr := p.parseExpression(lowest)
	if expr == nil {
		p.errorAt(line, p.curToken.Column, "expected expression after 'go'")
		return nil
	}
	stmt.Call = expr

	p.nextToken() // move past expression
	p.skipNewlines()
	return stmt
}

// parseSelectStmt parses: select when ... end
func (p *Parser) parseSelectStmt() *ast.SelectStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'select'
	p.skipNewlines()

	stmt := &ast.SelectStmt{Line: line}

	// Parse when clauses
	for p.curTokenIs(token.WHEN) {
		selectCase := p.parseSelectCase()
		stmt.Cases = append(stmt.Cases, selectCase)
	}

	// Parse optional else clause (default case)
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
			} else {
				p.nextToken() // error recovery: always advance
			}
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close select")
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

// parseSelectCase parses a single when clause in select
func (p *Parser) parseSelectCase() ast.SelectCase {
	p.nextToken() // consume 'when'

	selectCase := ast.SelectCase{}

	// Two forms:
	// 1. Receive: when val = ch.receive  OR  when ch.receive (no assignment)
	// 2. Send: when ch << value

	// Check for assignment pattern: ident = ...
	if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.ASSIGN) {
		selectCase.AssignName = p.curToken.Literal
		p.nextToken() // consume ident
		p.nextToken() // consume '='
	}

	// Parse the channel expression
	expr := p.parseExpression(lowest)

	// Check if this is a send expression (ch << val was parsed as expression)
	// or a receive (ch.receive)
	if binary, ok := expr.(*ast.BinaryExpr); ok && binary.Op == "<<" {
		selectCase.IsSend = true
		selectCase.Chan = binary.Left
		selectCase.Value = binary.Right
	} else {
		// Receive case - the expression should be the channel or channel.receive
		selectCase.IsSend = false
		selectCase.Chan = expr
	}

	p.nextToken() // move past expression
	p.skipNewlines()

	// Parse body until next WHEN, ELSE, or END
	for !p.curTokenIs(token.WHEN) && !p.curTokenIs(token.ELSE) &&
		!p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.WHEN) || p.curTokenIs(token.ELSE) || p.curTokenIs(token.END) {
			break
		}
		if s := p.parseStatement(); s != nil {
			selectCase.Body = append(selectCase.Body, s)
		} else {
			p.nextToken() // error recovery: always advance
		}
	}

	return selectCase
}

// parseConcurrentlyStmt parses: concurrently do |scope| ... end
func (p *Parser) parseConcurrentlyStmt() *ast.ConcurrentlyStmt {
	line := p.curToken.Line
	p.nextToken() // consume 'concurrently'

	stmt := &ast.ConcurrentlyStmt{Line: line}

	// Expect 'do' keyword
	if !p.curTokenIs(token.DO) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'do' after 'concurrently'")
		return nil
	}
	p.nextToken() // consume 'do'

	// Parse optional scope parameter: |scope|
	if p.curTokenIs(token.PIPE) {
		p.nextToken() // consume '|'
		if p.curTokenIs(token.IDENT) {
			stmt.ScopeVar = p.curToken.Literal
			p.nextToken() // consume scope var name
		}
		if !p.curTokenIs(token.PIPE) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected '|' after scope parameter")
			return nil
		}
		p.nextToken() // consume '|'
	}
	p.skipNewlines()

	// Parse body until 'end'
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}
		if s := p.parseStatement(); s != nil {
			stmt.Body = append(stmt.Body, s)
		} else {
			p.nextToken() // error recovery: always advance
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close concurrently block")
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}
