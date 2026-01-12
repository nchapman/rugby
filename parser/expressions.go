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
	case token.MINUS, token.NOT, token.BANG:
		left = p.parsePrefixExpr()
	case token.AT:
		left = p.parseInstanceVar()
	case token.SELF:
		left = &ast.Ident{Name: "self"}
	case token.SUPER:
		left = p.parseSuperExpr()
	case token.SPAWN:
		left = p.parseSpawnExpr()
	case token.AWAIT:
		left = p.parseAwaitExpr()
	default:
		return nil
	}

	// Infix parsing
infixLoop:
	for !p.peekTokenIs(token.NEWLINE) && !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
		switch p.peekToken.Type {
		case token.PLUS, token.MINUS, token.STAR, token.SLASH, token.PERCENT,
			token.EQ, token.NE, token.LT, token.GT, token.LE, token.GE,
			token.AND, token.OR, token.SHOVELLEFT:
			p.nextToken()
			left = p.parseInfixExpr(left)
		case token.QUESTIONQUESTION:
			p.nextToken()
			left = p.parseNilCoalesceExpr(left)
		case token.AMPDOT:
			p.nextToken()
			left = p.parseSafeNavExpr(left)
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
		case token.BANG:
			p.nextToken()
			left = p.parseBangExpr(left)
		case token.RESCUE:
			p.nextToken()
			left = p.parseRescueExpr(left)
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
	if p.curToken.Type == token.NOT || p.curToken.Type == token.BANG {
		op = "!"
	}
	p.nextToken()

	expr := p.parseExpression(prefix)
	return &ast.UnaryExpr{Op: op, Expr: expr}
}

// parseBangExpr parses postfix error propagation: call()!
func (p *Parser) parseBangExpr(left ast.Expression) ast.Expression {
	// Validate that left is a CallExpr
	if _, ok := left.(*ast.CallExpr); !ok {
		p.errorWithHint("'!' can only follow a call expression",
			"add parentheses: foo()!")
		return left
	}
	return &ast.BangExpr{Expr: left}
}

// parseRescueExpr parses error recovery: call() rescue default
// curToken is RESCUE when this is called
func (p *Parser) parseRescueExpr(left ast.Expression) ast.Expression {
	// Validate that left is a CallExpr
	if _, ok := left.(*ast.CallExpr); !ok {
		p.errorWithHint("'rescue' can only follow a call expression",
			"use 'if err != nil' for explicit handling")
		return left
	}

	rescue := &ast.RescueExpr{Expr: left}

	// Move past 'rescue' to see what comes next
	p.nextToken()

	// Check for error binding: rescue => err do
	if p.curTokenIs(token.HASHROCKET) {
		p.nextToken() // consume '=>'
		if !p.curTokenIs(token.IDENT) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected identifier after '=>'")
			return left
		}
		rescue.ErrName = p.curToken.Literal
		p.nextToken() // consume identifier

		// Must be followed by 'do'
		if !p.curTokenIs(token.DO) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'do' after error binding")
			return left
		}
		rescue.Block = p.parseBlock(token.END)
	} else if p.curTokenIs(token.DO) {
		// Block form: rescue do ... end
		rescue.Block = p.parseBlock(token.END)
	} else {
		// Inline form: rescue default_expr
		rescue.Default = p.parseExpression(rescuePrec + 1) // parse with higher precedence to get just the default
	}

	return rescue
}

// parseNilCoalesceExpr parses the nil coalescing operator: expr ?? default
// Returns the value if present, otherwise the default
func (p *Parser) parseNilCoalesceExpr(left ast.Expression) ast.Expression {
	// curToken is '??'
	precedence := p.curPrecedence()
	p.nextToken()
	right := p.parseExpression(precedence)

	return &ast.NilCoalesceExpr{Left: left, Right: right}
}

// parseSafeNavExpr parses the safe navigation operator: expr&.method
// Calls method only if expr is present, returns optional
func (p *Parser) parseSafeNavExpr(left ast.Expression) ast.Expression {
	// curToken is '&.'
	p.nextToken() // move past '&.' to the selector

	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected identifier after '&.'")
		return nil
	}

	return &ast.SafeNavExpr{Receiver: left, Selector: p.curToken.Literal}
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
	line := p.curToken.Line
	exclusive := p.curToken.Type == token.TRIPLEDOT
	precedence := p.curPrecedence()
	p.nextToken()
	end := p.parseExpression(precedence)

	return &ast.RangeLit{Start: start, End: end, Exclusive: exclusive, Line: line}
}

// parseSelectorExpr parses field/method access: x.y.
func (p *Parser) parseSelectorExpr(x ast.Expression) ast.Expression {
	// curToken is '.'
	p.nextToken() // move past '.' to the selector

	// Allow certain keywords as method names (obj.as(Type), arr.select { |x| ... }, scope.spawn { })
	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.AS) && !p.curTokenIs(token.SELECT) &&
		!p.curTokenIs(token.SPAWN) && !p.curTokenIs(token.AWAIT) {
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

// parseSuperExpr parses a super call: super or super(args).
func (p *Parser) parseSuperExpr() ast.Expression {
	line := p.curToken.Line
	// curToken is 'super'

	// Check if followed by '(' for explicit args
	var args []ast.Expression
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken() // move to '('
		p.nextToken() // move past '(' to first arg or ')'

		// Parse arguments
		if !p.curTokenIs(token.RPAREN) {
			args = append(args, p.parseExpression(lowest))

			for p.peekTokenIs(token.COMMA) {
				p.nextToken() // move to ','
				p.nextToken() // move past ',' to next arg
				args = append(args, p.parseExpression(lowest))
			}

			p.nextToken() // move past last arg to ')'
		}

		if !p.curTokenIs(token.RPAREN) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected ')' after super arguments")
			return nil
		}
	}

	return &ast.SuperExpr{Args: args, Line: line}
}

// parseSpawnExpr parses: spawn { expr } or spawn do ... end
func (p *Parser) parseSpawnExpr() ast.Expression {
	line := p.curToken.Line
	p.nextToken() // consume 'spawn'

	// Expect block: spawn { expr } or spawn do ... end
	var block *ast.BlockExpr
	if p.curTokenIs(token.DO) {
		block = p.parseBlock(token.END)
	} else if p.curTokenIs(token.LBRACE) {
		block = p.parseBlock(token.RBRACE)
	} else {
		p.errorAt(line, p.curToken.Column, "expected block after 'spawn'")
		return nil
	}

	return &ast.SpawnExpr{Block: block, Line: line}
}

// parseAwaitExpr parses: await task or await(task)
func (p *Parser) parseAwaitExpr() ast.Expression {
	line := p.curToken.Line
	p.nextToken() // consume 'await'

	// Handle optional parentheses: await(task) or await task
	var task ast.Expression
	if p.curTokenIs(token.LPAREN) {
		p.nextToken() // consume '('
		task = p.parseExpression(lowest)
		p.nextToken() // move past expression
		if !p.curTokenIs(token.RPAREN) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected ')' after await expression")
			return nil
		}
		// Don't advance past ')' - let infix loop handle it
	} else {
		// Without parens, parse as the following expression
		task = p.parseExpression(lowest)
		// Don't advance - the main parseExpression will handle token advancement
	}

	return &ast.AwaitExpr{Task: task, Line: line}
}
