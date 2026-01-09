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
	OR_PREC     // or
	AND_PREC    // and
	EQUALS      // ==, !=
	LESSGREATER // <, >, <=, >=
	SUM         // +, -
	PRODUCT     // *, /, %
	PREFIX      // -x, not x
	CALL        // f(x)
	MEMBER      // x.y (highest)
)

var precedences = map[token.TokenType]int{
	token.OR:       OR_PREC,
	token.AND:      AND_PREC,
	token.EQ:       EQUALS,
	token.NE:       EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LE:       LESSGREATER,
	token.GE:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.STAR:     PRODUCT,
	token.SLASH:    PRODUCT,
	token.PERCENT:  PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: CALL, // array indexing has same precedence as function calls
	token.DOT:      MEMBER,
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

	// Check for optional alias: import foo/bar as baz
	if p.curTokenIs(token.AS) {
		p.nextToken() // consume 'as'
		if !p.curTokenIs(token.IDENT) {
			p.errors = append(p.errors, fmt.Sprintf("line %d: expected alias name after 'as'",
				p.curToken.Line))
			return nil
		}
		imp.Alias = p.curToken.Literal
		p.nextToken()
	}

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
	case token.DEFER:
		return p.parseDeferStmt()
	case token.IDENT:
		// Check for assignment: ident = expr
		if p.peekTokenIs(token.ASSIGN) {
			return p.parseAssignStmt()
		}
	}

	// Default: expression statement
	expr := p.parseExpression(LOWEST)
	if expr != nil {
		// Check for block: expr do |params| ... end  OR  expr {|params| ... }
		if p.peekTokenIs(token.DO) || p.peekTokenIs(token.LBRACE) {
			expr = p.parseBlockCall(expr)
		}
		p.nextToken() // move past expression
		p.skipNewlines()
		return &ast.ExprStmt{Expr: expr}
	}
	return nil
}

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
		p.errors = append(p.errors, fmt.Sprintf("line %d: block can only follow a method call",
			p.curToken.Line))
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
				p.errors = append(p.errors, fmt.Sprintf("line %d: duplicate block parameter name %q",
					p.curToken.Line, name))
				p.skipTo(terminator)
				return nil
			}
			seen[name] = true
			block.Params = append(block.Params, name)
			p.nextToken()

			for p.curTokenIs(token.COMMA) {
				p.nextToken() // move past ','
				if !p.curTokenIs(token.IDENT) {
					p.errors = append(p.errors, fmt.Sprintf("line %d: expected parameter name after comma",
						p.curToken.Line))
					p.skipTo(terminator)
					return nil
				}
				name := p.curToken.Literal
				if seen[name] {
					p.errors = append(p.errors, fmt.Sprintf("line %d: duplicate block parameter name %q",
						p.curToken.Line, name))
					p.skipTo(terminator)
					return nil
				}
				seen[name] = true
				block.Params = append(block.Params, name)
				p.nextToken()
			}
		}

		if !p.curTokenIs(token.PIPE) {
			p.errors = append(p.errors, fmt.Sprintf("line %d: expected '|' after block parameters",
				p.curToken.Line))
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
		}
	}

	if !p.curTokenIs(terminator) {
		closer := "end"
		if terminator == token.RBRACE {
			closer = "}"
		}
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected '%s' to close block",
			p.curToken.Line, closer))
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

func (p *Parser) parseAssignStmt() *ast.AssignStmt {
	name := p.curToken.Literal
	p.nextToken() // consume ident
	p.nextToken() // consume '='

	value := p.parseExpression(LOWEST)

	// Check for block: expr.method do |x| ... end  OR  expr.method {|x| ... }
	if p.peekTokenIs(token.DO) || p.peekTokenIs(token.LBRACE) {
		value = p.parseBlockCall(value)
	}

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
		stmt.Values = append(stmt.Values, p.parseExpression(LOWEST))
		p.nextToken() // move past expression

		// Parse additional return values
		for p.curTokenIs(token.COMMA) {
			p.nextToken() // consume ','
			// Allow trailing comma
			if p.curTokenIs(token.NEWLINE) || p.curTokenIs(token.END) || p.curTokenIs(token.EOF) {
				break
			}
			stmt.Values = append(stmt.Values, p.parseExpression(LOWEST))
			p.nextToken() // move past expression
		}
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
	case token.LBRACKET:
		left = p.parseArrayLiteral()
	case token.LBRACE:
		left = p.parseMapLiteral()
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

	// Handle Ruby-style function call without parens: puts "hi"
	// Note: LBRACKET not included here - use parens for array args: puts([1,2,3])
	// This avoids ambiguity with array indexing (arr[0])
	if ident, ok := left.(*ast.Ident); ok {
		if p.peekTokenIs(token.STRING) || p.peekTokenIs(token.INT) ||
			p.peekTokenIs(token.FLOAT) || p.peekTokenIs(token.TRUE) ||
			p.peekTokenIs(token.FALSE) || p.peekTokenIs(token.IDENT) ||
			p.peekTokenIs(token.LPAREN) {
			p.nextToken()
			arg := p.parseExpression(LOWEST)
			return &ast.CallExpr{Func: ident, Args: []ast.Expression{arg}}
		}
	}

	return left
}

func (p *Parser) parseIdent() ast.Expression {
	return &ast.Ident{Name: p.curToken.Literal}
}

func (p *Parser) parseCallExprWithParens(fn ast.Expression) ast.Expression {
	// curToken is '('
	call := &ast.CallExpr{Func: fn}

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

func (p *Parser) parseArrayLiteral() ast.Expression {
	arr := &ast.ArrayLit{}
	startLine := p.curToken.Line

	p.nextToken() // consume '['

	// Handle empty array
	if p.curTokenIs(token.RBRACKET) {
		return arr
	}

	// Parse first element
	elem := p.parseExpression(LOWEST)
	if elem == nil {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected expression in array literal",
			p.curToken.Line))
		return nil
	}
	arr.Elements = append(arr.Elements, elem)

	// Parse remaining elements
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // move to ','
		p.nextToken() // move past ',' to next element

		// Allow trailing comma
		if p.curTokenIs(token.RBRACKET) {
			return arr
		}

		elem := p.parseExpression(LOWEST)
		if elem == nil {
			p.errors = append(p.errors, fmt.Sprintf("line %d: expected expression after comma in array literal",
				startLine))
			return nil
		}
		arr.Elements = append(arr.Elements, elem)
	}

	// Move past last element to ']'
	p.nextToken()

	if !p.curTokenIs(token.RBRACKET) {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected ']' after array elements",
			p.curToken.Line))
		return nil
	}

	return arr
}

func (p *Parser) parseIndexExpr(left ast.Expression) ast.Expression {
	// curToken is '['
	p.nextToken() // move past '[' to the index expression

	index := p.parseExpression(LOWEST)
	if index == nil {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected index expression",
			p.curToken.Line))
		return nil
	}

	// Move past index expression to ']'
	p.nextToken()

	if !p.curTokenIs(token.RBRACKET) {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected ']' after index",
			p.curToken.Line))
		return nil
	}

	return &ast.IndexExpr{Left: left, Index: index}
}

func (p *Parser) parseMapLiteral() ast.Expression {
	mapLit := &ast.MapLit{}
	startLine := p.curToken.Line

	p.nextToken() // consume '{'

	// Handle empty map
	if p.curTokenIs(token.RBRACE) {
		return mapLit
	}

	// Parse first entry
	entry, ok := p.parseMapEntry(startLine)
	if !ok {
		return nil
	}
	mapLit.Entries = append(mapLit.Entries, entry)

	// Parse remaining entries
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // move to ','
		p.nextToken() // move past ',' to next key

		// Allow trailing comma
		if p.curTokenIs(token.RBRACE) {
			return mapLit
		}

		entry, ok := p.parseMapEntry(startLine)
		if !ok {
			return nil
		}
		mapLit.Entries = append(mapLit.Entries, entry)
	}

	// Move past last value to '}'
	p.nextToken()

	if !p.curTokenIs(token.RBRACE) {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected '}' after map entries",
			p.curToken.Line))
		return nil
	}

	return mapLit
}

func (p *Parser) parseMapEntry(startLine int) (ast.MapEntry, bool) {
	key := p.parseExpression(LOWEST)
	if key == nil {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected key in map literal",
			p.curToken.Line))
		return ast.MapEntry{}, false
	}

	// Expect '=>'
	if !p.peekTokenIs(token.HASHROCKET) {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected '=>' after map key",
			startLine))
		return ast.MapEntry{}, false
	}
	p.nextToken() // move to '=>'
	p.nextToken() // move past '=>' to value

	value := p.parseExpression(LOWEST)
	if value == nil {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected value after '=>' in map literal",
			startLine))
		return ast.MapEntry{}, false
	}

	return ast.MapEntry{Key: key, Value: value}, true
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

func (p *Parser) parseSelectorExpr(x ast.Expression) ast.Expression {
	// curToken is '.'
	p.nextToken() // move past '.' to the selector

	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected identifier after '.'",
			p.curToken.Line))
		return nil
	}

	return &ast.SelectorExpr{
		X:   x,
		Sel: p.curToken.Literal,
	}
}

func (p *Parser) parseDeferStmt() *ast.DeferStmt {
	deferLine := p.curToken.Line
	p.nextToken() // consume 'defer'

	// Parse the expression that should be a call
	expr := p.parseExpression(LOWEST)
	if expr == nil {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected expression after defer",
			deferLine))
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
		p.errors = append(p.errors, fmt.Sprintf("line %d: defer requires a callable expression",
			deferLine))
		return nil
	}

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.DeferStmt{Call: call}
}
