// Package parser implements the Rugby parser.
package parser

import (
	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/token"
)

// parseExpression is the Pratt parser core for expressions.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	var left ast.Expression

	// Prefix parsing
	switch p.curToken.Type {
	case token.IDENT:
		left = p.parseIdent()
	case token.INT:
		left = p.parseIntLiteral()
	case token.FLOAT:
		left = p.parseFloatLiteral()
	case token.STRING:
		left = p.parseStringLiteral()
	case token.TRUE, token.FALSE:
		left = p.parseBoolLiteral()
	case token.NIL:
		left = p.parseNilLiteral()
	case token.SYMBOL:
		left = p.parseSymbolLiteral()
	case token.LPAREN:
		left = p.parseGroupedExpr()
	case token.LBRACKET:
		left = p.parseArrayLiteral()
	case token.LBRACE:
		left = p.parseMapLiteral()
	case token.MINUS, token.NOT:
		left = p.parsePrefixExpr()
	case token.AT:
		left = p.parseInstanceVar()
	case token.SELF:
		left = &ast.Ident{Name: "self"}
	default:
		return nil
	}

	// Infix parsing
infixLoop:
	for !p.peekTokenIs(token.NEWLINE) && !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
		switch p.peekToken.Type {
		case token.PLUS, token.MINUS, token.STAR, token.SLASH, token.PERCENT,
			token.EQ, token.NE, token.LT, token.GT, token.LE, token.GE,
			token.AND, token.OR:
			p.nextToken()
			left = p.parseInfixExpr(left)
		case token.DOTDOT, token.TRIPLEDOT:
			p.nextToken()
			left = p.parseRangeLit(left)
		case token.DOT:
			p.nextToken()
			left = p.parseSelectorExpr(left)
		case token.LPAREN:
			p.nextToken()
			left = p.parseCallExprWithParens(left)
		case token.LBRACKET:
			p.nextToken()
			left = p.parseIndexExpr(left)
		default:
			break infixLoop
		}
	}

	return left
}

// parseIdent parses an identifier.
func (p *Parser) parseIdent() ast.Expression {
	return &ast.Ident{Name: p.curToken.Literal}
}

// parseCallExprWithParens parses a function call with parentheses: f(a, b, c).
func (p *Parser) parseCallExprWithParens(fn ast.Expression) ast.Expression {
	// curToken is '('
	call := &ast.CallExpr{Func: fn}

	p.nextToken() // move past '(' to first arg or ')'

	// Parse arguments
	if !p.curTokenIs(token.RPAREN) {
		call.Args = append(call.Args, p.parseExpression(lowest))

		for p.peekTokenIs(token.COMMA) {
			p.nextToken() // move to ','
			p.nextToken() // move past ',' to next arg
			call.Args = append(call.Args, p.parseExpression(lowest))
		}

		p.nextToken() // move past last arg to ')'
	}

	if !p.curTokenIs(token.RPAREN) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected ')' after arguments")
		return nil
	}

	return call
}

// parseIndexExpr parses array/map indexing: arr[i].
func (p *Parser) parseIndexExpr(left ast.Expression) ast.Expression {
	// curToken is '['
	p.nextToken() // move past '[' to the index expression

	index := p.parseExpression(lowest)
	if index == nil {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected index expression")
		return nil
	}

	// Move past index expression to ']'
	p.nextToken()

	if !p.curTokenIs(token.RBRACKET) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected ']' after index")
		return nil
	}

	return &ast.IndexExpr{Left: left, Index: index}
}

// parseGroupedExpr parses a parenthesized expression: (expr).
func (p *Parser) parseGroupedExpr() ast.Expression {
	p.nextToken() // consume '('

	expr := p.parseExpression(lowest)

	if !p.peekTokenIs(token.RPAREN) {
		p.errorAt(p.peekToken.Line, p.peekToken.Column, "expected ')'")
		return nil
	}
	p.nextToken() // move to ')'

	return expr
}

// parsePrefixExpr parses a prefix expression: -x, !x.
func (p *Parser) parsePrefixExpr() ast.Expression {
	op := p.curToken.Literal
	if p.curToken.Type == token.NOT {
		op = "!"
	}
	p.nextToken()

	expr := p.parseExpression(prefix)
	return &ast.UnaryExpr{Op: op, Expr: expr}
}

// parseInfixExpr parses a binary expression: a + b, a && b.
func (p *Parser) parseInfixExpr(left ast.Expression) ast.Expression {
	op := p.curToken.Literal
	// Map 'and'/'or' to Go operators
	switch p.curToken.Type {
	case token.AND:
		op = "&&"
	case token.OR:
		op = "||"
	}

	precedence := p.curPrecedence()
	p.nextToken()
	right := p.parseExpression(precedence)

	return &ast.BinaryExpr{Left: left, Op: op, Right: right}
}

// parseRangeLit parses a range literal: start..end or start...end.
func (p *Parser) parseRangeLit(start ast.Expression) ast.Expression {
	// curToken is '..' or '...'
	exclusive := p.curToken.Type == token.TRIPLEDOT
	precedence := p.curPrecedence()
	p.nextToken()
	end := p.parseExpression(precedence)

	return &ast.RangeLit{Start: start, End: end, Exclusive: exclusive}
}

// parseSelectorExpr parses field/method access: x.y.
func (p *Parser) parseSelectorExpr(x ast.Expression) ast.Expression {
	// curToken is '.'
	p.nextToken() // move past '.' to the selector

	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected identifier after '.'")
		return nil
	}

	return &ast.SelectorExpr{
		X:   x,
		Sel: p.curToken.Literal,
	}
}

// parseInstanceVar parses an instance variable read: @name.
func (p *Parser) parseInstanceVar() ast.Expression {
	p.nextToken() // consume '@'
	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected identifier after '@'")
		return nil
	}
	name := p.curToken.Literal
	return &ast.InstanceVar{Name: name}
}
