// Package parser implements the Rugby parser.
package parser

import (
	"fmt"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/token"
)

// parseBlockCall parses a block attached to a method call.
// Converts expr.method or expr.method(args) followed by do |params| ... end
// or {|params| ... } into a CallExpr with a Block attached.
func (p *Parser) parseBlockCall(expr ast.Expression) ast.Expression {
	p.nextToken() // move to 'do' or '{'

	var block *ast.BlockExpr
	if p.curTokenIs(token.DO) {
		block = p.parseBlock(token.END)
	} else if p.curTokenIs(token.LBRACE) {
		block = p.parseBlock(token.RBRACE)
	} else {
		return expr // should not happen since we checked peekTokenIs
	}

	if block == nil {
		return expr
	}

	// Convert expression to CallExpr with block
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		// arr.each do |x| -> CallExpr{Func: SelectorExpr{arr, "each"}, Block: ...}
		return &ast.CallExpr{Func: e, Block: block}
	case *ast.CallExpr:
		// arr.method(args) do |x| -> just attach block to existing call
		e.Block = block
		return e
	default:
		p.errorAt(p.curToken.Line, p.curToken.Column, "block can only follow a method call")
		return expr
	}
}

// parseTrailingLambda parses a lambda attached to a method call.
// Converts expr.method or expr.method(args) followed by -> (params) { ... }
// into a CallExpr with the lambda appended as an argument.
func (p *Parser) parseTrailingLambda(expr ast.Expression) ast.Expression {
	// Parse the lambda expression (curToken is before ARROW)
	p.nextToken() // move to ARROW
	lambda := p.parseLambdaExpr()
	if lambda == nil {
		return expr
	}

	// Append lambda as an argument to the CallExpr
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		// arr.each -> (x) { } -> CallExpr{Func: SelectorExpr{arr, "each"}, Args: [lambda]}
		return &ast.CallExpr{Func: e, Args: []ast.Expression{lambda}}
	case *ast.CallExpr:
		// arr.reduce(0) -> (acc, x) { } -> append lambda to existing args
		e.Args = append(e.Args, lambda)
		return e
	default:
		p.errorAt(p.curToken.Line, p.curToken.Column, "lambda can only follow a method call")
		return expr
	}
}

// parseBlock parses a block: do |params| ... end  OR  {|params| ... }
// Assumes curToken is 'do' or '{'. terminator specifies the closing token (END or RBRACE).
func (p *Parser) parseBlock(terminator token.TokenType) *ast.BlockExpr {
	p.nextToken() // move past 'do' or '{'

	block := &ast.BlockExpr{}

	// Parse optional parameters: |var| or |var1, var2|
	if p.curTokenIs(token.PIPE) {
		p.nextToken() // move past first '|'

		seen := make(map[string]bool)

		// Parse parameter list
		if p.curTokenIs(token.IDENT) {
			name := p.curToken.Literal
			if seen[name] {
				p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("duplicate block parameter name %q", name))
				p.skipTo(terminator)
				return nil
			}
			seen[name] = true
			block.Params = append(block.Params, name)
			p.nextToken()

			for p.curTokenIs(token.COMMA) {
				p.nextToken() // move past ','
				if !p.curTokenIs(token.IDENT) {
					p.errorAt(p.curToken.Line, p.curToken.Column, "expected parameter name after comma")
					p.skipTo(terminator)
					return nil
				}
				name := p.curToken.Literal
				if seen[name] {
					p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("duplicate block parameter name %q", name))
					p.skipTo(terminator)
					return nil
				}
				seen[name] = true
				block.Params = append(block.Params, name)
				p.nextToken()
			}
		}

		if !p.curTokenIs(token.PIPE) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected '|' after block parameters")
			p.skipTo(terminator)
			return nil
		}
		p.nextToken() // move past closing '|'
	}

	p.skipNewlines()

	// Parse body until terminator (end or })
	for !p.curTokenIs(terminator) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(terminator) {
			break
		}
		if stmt := p.parseStatement(); stmt != nil {
			block.Body = append(block.Body, stmt)
		} else {
			p.nextToken() // error recovery: always advance to prevent infinite loop
		}
	}

	if !p.curTokenIs(terminator) {
		closer := "end"
		if terminator == token.RBRACE {
			closer = "}"
		}
		p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("expected '%s' to close block", closer))
		return nil
	}
	p.nextToken() // consume terminator

	return block
}

// skipTo skips tokens until the specified terminator is found (for error recovery in blocks)
func (p *Parser) skipTo(terminator token.TokenType) {
	for !p.curTokenIs(terminator) && !p.curTokenIs(token.EOF) {
		p.nextToken()
	}
	if p.curTokenIs(terminator) {
		p.nextToken() // consume terminator
	}
}
