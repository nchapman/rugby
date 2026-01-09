package parser

import (
	"fmt"
	"strconv"

	"rugby/ast"
	"rugby/lexer"
	"rugby/token"
)

// Precedence levels
const (
	_ int = iota
	LOWEST
	OR_PREC      // or
	AND_PREC     // and
	EQUALS       // ==, !=
	LESSGREATER  // <, >, <=, >=
	SUM          // +, -
	PRODUCT      // *, /, %
	PREFIX       // -x, not x
	CALL         // f(x)
)

var precedences = map[token.TokenType]int{
	token.OR:      OR_PREC,
	token.AND:     AND_PREC,
	token.EQ:      EQUALS,
	token.NE:      EQUALS,
	token.LT:      LESSGREATER,
	token.GT:      LESSGREATER,
	token.LE:      LESSGREATER,
	token.GE:      LESSGREATER,
	token.PLUS:    SUM,
	token.MINUS:   SUM,
	token.STAR:    PRODUCT,
	token.SLASH:   PRODUCT,
	token.PERCENT: PRODUCT,
	token.LPAREN:  CALL,
}

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	errors    []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
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

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
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

	// Parse optional parameter list
	if p.curTokenIs(token.LPAREN) {
		p.nextToken() // consume '('
		seen := make(map[string]bool)

		// Parse parameters until ')'
		if !p.curTokenIs(token.RPAREN) {
			// First parameter
			if !p.curTokenIs(token.IDENT) {
				p.errors = append(p.errors, fmt.Sprintf("line %d: expected parameter name",
					p.curToken.Line))
				return nil
			}
			name := p.curToken.Literal
			if seen[name] {
				p.errors = append(p.errors, fmt.Sprintf("line %d: duplicate parameter name %q",
					p.curToken.Line, name))
				return nil
			}
			seen[name] = true
			fn.Params = append(fn.Params, &ast.Param{Name: name})
			p.nextToken()

			// Additional parameters
			for p.curTokenIs(token.COMMA) {
				p.nextToken() // consume ','
				// Allow trailing comma
				if p.curTokenIs(token.RPAREN) {
					break
				}
				if !p.curTokenIs(token.IDENT) {
					p.errors = append(p.errors, fmt.Sprintf("line %d: expected parameter name after comma",
						p.curToken.Line))
					return nil
				}
				name := p.curToken.Literal
				if seen[name] {
					p.errors = append(p.errors, fmt.Sprintf("line %d: duplicate parameter name %q",
						p.curToken.Line, name))
					return nil
				}
				seen[name] = true
				fn.Params = append(fn.Params, &ast.Param{Name: name})
				p.nextToken()
			}
		}

		if !p.curTokenIs(token.RPAREN) {
			p.errors = append(p.errors, fmt.Sprintf("line %d: expected ')' after parameters",
				p.curToken.Line))
			return nil
		}
		p.nextToken() // consume ')'
	}

	// Parse optional return type: -> Type or -> (Type1, Type2)
	if p.curTokenIs(token.ARROW) {
		p.nextToken() // consume '->'

		if p.curTokenIs(token.LPAREN) {
			// Multiple return types: (Type1, Type2, ...)
			p.nextToken() // consume '('

			if !p.curTokenIs(token.IDENT) {
				p.errors = append(p.errors, fmt.Sprintf("line %d: expected type in return type list",
					p.curToken.Line))
				return nil
			}
			fn.ReturnTypes = append(fn.ReturnTypes, p.curToken.Literal)
			p.nextToken()

			for p.curTokenIs(token.COMMA) {
				p.nextToken() // consume ','
				// Allow trailing comma
				if p.curTokenIs(token.RPAREN) {
					break
				}
				if !p.curTokenIs(token.IDENT) {
					p.errors = append(p.errors, fmt.Sprintf("line %d: expected type after comma",
						p.curToken.Line))
					return nil
				}
				fn.ReturnTypes = append(fn.ReturnTypes, p.curToken.Literal)
				p.nextToken()
			}

			if !p.curTokenIs(token.RPAREN) {
				p.errors = append(p.errors, fmt.Sprintf("line %d: expected ')' after return types",
					p.curToken.Line))
				return nil
			}
			p.nextToken() // consume ')'
		} else if p.curTokenIs(token.IDENT) {
			// Single return type
			fn.ReturnTypes = append(fn.ReturnTypes, p.curToken.Literal)
			p.nextToken()
		} else {
			p.errors = append(p.errors, fmt.Sprintf("line %d: expected type after ->",
				p.curToken.Line))
			return nil
		}
	}

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
	switch p.curToken.Type {
	case token.IF:
		return p.parseIfStmt()
	case token.WHILE:
		return p.parseWhileStmt()
	case token.RETURN:
		return p.parseReturnStmt()
	case token.IDENT:
		// Check for assignment: ident = expr
		if p.peekTokenIs(token.ASSIGN) {
			return p.parseAssignStmt()
		}
	}

	// Default: expression statement
	expr := p.parseExpression(LOWEST)
	if expr != nil {
		p.nextToken() // move past expression
		p.skipNewlines()
		return &ast.ExprStmt{Expr: expr}
	}
	return nil
}

func (p *Parser) parseAssignStmt() *ast.AssignStmt {
	name := p.curToken.Literal
	p.nextToken() // consume ident
	p.nextToken() // consume '='

	value := p.parseExpression(LOWEST)
	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.AssignStmt{Name: name, Value: value}
}

func (p *Parser) parseIfStmt() *ast.IfStmt {
	p.nextToken() // consume 'if'

	stmt := &ast.IfStmt{}
	stmt.Cond = p.parseExpression(LOWEST)
	p.nextToken() // move past condition
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
		clause.Cond = p.parseExpression(LOWEST)
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
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected 'end' to close if",
			p.curToken.Line))
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

func (p *Parser) parseWhileStmt() *ast.WhileStmt {
	p.nextToken() // consume 'while'

	stmt := &ast.WhileStmt{}
	stmt.Cond = p.parseExpression(LOWEST)
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
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected 'end' to close while",
			p.curToken.Line))
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

func (p *Parser) parseReturnStmt() *ast.ReturnStmt {
	p.nextToken() // consume 'return'

	stmt := &ast.ReturnStmt{}

	// Check if there's a return value (not newline or end)
	if !p.curTokenIs(token.NEWLINE) && !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		stmt.Value = p.parseExpression(LOWEST)
		p.nextToken() // move past expression
	}

	p.skipNewlines()
	return stmt
}

// Pratt parser for expressions
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
	case token.LPAREN:
		left = p.parseGroupedExpr()
	case token.MINUS, token.NOT:
		left = p.parsePrefixExpr()
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
		case token.LPAREN:
			p.nextToken()
			left = p.parseCallExprWithParens(left)
		default:
			break infixLoop
		}
	}

	// Handle Ruby-style function call without parens: puts "hi"
	if ident, ok := left.(*ast.Ident); ok {
		if p.peekTokenIs(token.STRING) || p.peekTokenIs(token.INT) ||
			p.peekTokenIs(token.FLOAT) || p.peekTokenIs(token.TRUE) ||
			p.peekTokenIs(token.FALSE) || p.peekTokenIs(token.IDENT) ||
			p.peekTokenIs(token.LPAREN) {
			p.nextToken()
			arg := p.parseExpression(LOWEST)
			return &ast.CallExpr{Func: ident.Name, Args: []ast.Expression{arg}}
		}
	}

	return left
}

func (p *Parser) parseIdent() ast.Expression {
	return &ast.Ident{Name: p.curToken.Literal}
}

func (p *Parser) parseCallExprWithParens(fn ast.Expression) ast.Expression {
	// curToken is '('
	call := &ast.CallExpr{}

	// Get function name from identifier
	if ident, ok := fn.(*ast.Ident); ok {
		call.Func = ident.Name
	} else {
		p.errors = append(p.errors, "expected identifier before (")
		return nil
	}

	p.nextToken() // move past '(' to first arg or ')'

	// Parse arguments
	if !p.curTokenIs(token.RPAREN) {
		call.Args = append(call.Args, p.parseExpression(LOWEST))

		for p.peekTokenIs(token.COMMA) {
			p.nextToken() // move to ','
			p.nextToken() // move past ',' to next arg
			call.Args = append(call.Args, p.parseExpression(LOWEST))
		}

		p.nextToken() // move past last arg to ')'
	}

	if !p.curTokenIs(token.RPAREN) {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected ')' after arguments",
			p.curToken.Line))
		return nil
	}

	return call
}

func (p *Parser) parseIntLiteral() ast.Expression {
	val, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("line %d: invalid integer %s",
			p.curToken.Line, p.curToken.Literal))
		return nil
	}
	return &ast.IntLit{Value: val}
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	val, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("line %d: invalid float %s",
			p.curToken.Line, p.curToken.Literal))
		return nil
	}
	return &ast.FloatLit{Value: val}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLit{Value: p.curToken.Literal}
}

func (p *Parser) parseBoolLiteral() ast.Expression {
	return &ast.BoolLit{Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpr() ast.Expression {
	p.nextToken() // consume '('

	expr := p.parseExpression(LOWEST)

	if !p.peekTokenIs(token.RPAREN) {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected ')'",
			p.peekToken.Line))
		return nil
	}
	p.nextToken() // move to ')'

	return expr
}

func (p *Parser) parsePrefixExpr() ast.Expression {
	op := p.curToken.Literal
	if p.curToken.Type == token.NOT {
		op = "!"
	}
	p.nextToken()

	expr := p.parseExpression(PREFIX)
	return &ast.UnaryExpr{Op: op, Expr: expr}
}

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
