package parser

import (
	"fmt"
	"rugby/ast"
	"rugby/lexer"
	"rugby/token"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	errors    []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("line %d: expected %s, got %s",
		p.peekToken.Line, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) skipNewlines() {
	for p.curTokenIs(token.NEWLINE) {
		p.nextToken()
	}
}

// ParseProgram parses the entire program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}

	for !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.EOF) {
			break
		}

		switch p.curToken.Type {
		case token.IMPORT:
			if imp := p.parseImport(); imp != nil {
				program.Imports = append(program.Imports, imp)
			}
		case token.DEF:
			if fn := p.parseFuncDecl(); fn != nil {
				program.Declarations = append(program.Declarations, fn)
			}
		default:
			p.errors = append(p.errors, fmt.Sprintf("line %d: unexpected token %s",
				p.curToken.Line, p.curToken.Type))
			p.nextToken()
		}
	}

	return program
}

func (p *Parser) parseImport() *ast.ImportDecl {
	p.nextToken() // consume 'import'

	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected package path after import",
			p.curToken.Line))
		return nil
	}

	imp := &ast.ImportDecl{Path: p.curToken.Literal}
	p.nextToken()

	return imp
}

func (p *Parser) parseFuncDecl() *ast.FuncDecl {
	p.nextToken() // consume 'def'

	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected function name after def",
			p.curToken.Line))
		return nil
	}

	fn := &ast.FuncDecl{Name: p.curToken.Literal}
	p.nextToken()
	p.skipNewlines()

	// Parse body until 'end'
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}
		if stmt := p.parseStatement(); stmt != nil {
			fn.Body = append(fn.Body, stmt)
		}
	}

	if !p.curTokenIs(token.END) {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected 'end' to close function",
			p.curToken.Line))
		return nil
	}
	p.nextToken() // consume 'end'

	return fn
}

func (p *Parser) parseStatement() ast.Statement {
	expr := p.parseExpression()
	if expr != nil {
		stmt := &ast.ExprStmt{Expr: expr}
		p.skipNewlines()
		return stmt
	}
	return nil
}

func (p *Parser) parseExpression() ast.Expression {
	switch p.curToken.Type {
	case token.IDENT:
		return p.parseIdentOrCall()
	case token.STRING:
		return p.parseStringLiteral()
	default:
		return nil
	}
}

func (p *Parser) parseIdentOrCall() ast.Expression {
	name := p.curToken.Literal
	p.nextToken()

	// Check if this is a function call (followed by arguments)
	if p.curTokenIs(token.STRING) || p.curTokenIs(token.IDENT) {
		call := &ast.CallExpr{Func: name}
		// Parse arguments (simple: just one for now)
		if arg := p.parseExpression(); arg != nil {
			call.Args = append(call.Args, arg)
		}
		return call
	}

	return &ast.Ident{Name: name}
}

func (p *Parser) parseStringLiteral() *ast.StringLit {
	lit := &ast.StringLit{Value: p.curToken.Literal}
	p.nextToken()
	return lit
}
