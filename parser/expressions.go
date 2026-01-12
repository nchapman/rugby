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
	case token.STRING, token.HEREDOC:
		left = p.parseStringLiteral()
	case token.TRUE, token.FALSE:
		left = p.parseBoolLiteral()
	case token.NIL:
		left = p.parseNilLiteral()
	case token.SYMBOL:
		left = p.parseSymbolLiteral()
	case token.WORDARRAY:
		left = p.parseWordArray(false)
	case token.INTERPWARRAY:
		left = p.parseWordArray(true)
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
	case token.AMP:
		left = p.parseSymbolToProc()
	default:
		return nil
	}

	// Infix parsing
infixLoop:
	for !p.peekTokenIs(token.NEWLINE) && !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
		// Check for command syntax before processing infix operators
		// This handles cases like `foo -1` (command) vs `foo - 1` (binary)
		if p.isCallableExpr(left) && p.looksLikeCommandArg() {
			break infixLoop
		}

		switch p.peekToken.Type {
		case token.PLUS, token.MINUS, token.STAR, token.SLASH, token.PERCENT,
			token.EQ, token.NE, token.LT, token.GT, token.LE, token.GE,
			token.AND, token.OR, token.SHOVELLEFT:
			p.nextToken()
			left = p.parseInfixExpr(left)
		case token.QUESTION:
			p.nextToken()
			left = p.parseTernaryExpr(left)
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

	// Check for command syntax after the infix loop
	// This enables `puts "hello"` instead of `puts("hello")`
	// We check here because command arguments (like STRING) have no operator precedence,
	// so the infix loop exits before we can handle them.
	if p.isCallableExpr(left) && p.canStartCommandArg() {
		left = p.parseCommandCall(left)

		// After command syntax, continue checking for low-precedence operators
		// like `and`/`or`. This allows `puts x and y` → `(puts(x)) and y`
		for !p.peekTokenIs(token.NEWLINE) && !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
			switch p.peekToken.Type {
			case token.AND, token.OR:
				p.nextToken()
				left = p.parseInfixExpr(left)
			default:
				return left
			}
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

// parseBangExpr parses postfix error propagation: call()! or selector!
// For selector/ident expressions (like resp.json! or foo!), converts to a zero-arg call.
func (p *Parser) parseBangExpr(left ast.Expression) ast.Expression {
	switch expr := left.(type) {
	case *ast.CallExpr:
		return &ast.BangExpr{Expr: expr}
	case *ast.SelectorExpr:
		// Convert selector to zero-arg call: resp.json! → resp.json()!
		call := &ast.CallExpr{Func: expr}
		return &ast.BangExpr{Expr: call}
	case *ast.Ident:
		// Convert identifier to zero-arg call: foo! → foo()!
		call := &ast.CallExpr{Func: expr}
		return &ast.BangExpr{Expr: call}
	default:
		p.errorWithHint("'!' can only follow a call or method expression",
			"use: method()! or obj.method!")
		return left
	}
}

// parseRescueExpr parses error recovery: call() rescue default
// curToken is RESCUE when this is called
// For selector/ident expressions, converts to a zero-arg call.
func (p *Parser) parseRescueExpr(left ast.Expression) ast.Expression {
	// Convert to CallExpr if needed
	var callExpr *ast.CallExpr
	switch expr := left.(type) {
	case *ast.CallExpr:
		callExpr = expr
	case *ast.SelectorExpr:
		// Convert selector to zero-arg call: resp.json rescue x → resp.json() rescue x
		callExpr = &ast.CallExpr{Func: expr}
	case *ast.Ident:
		// Convert identifier to zero-arg call: foo rescue x → foo() rescue x
		callExpr = &ast.CallExpr{Func: expr}
	default:
		p.errorWithHint("'rescue' can only follow a call or method expression",
			"use: method() rescue default or obj.method rescue default")
		return left
	}

	rescue := &ast.RescueExpr{Expr: callExpr}

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

// parseTernaryExpr parses the ternary conditional: cond ? then : else
// Ruby: status = valid? ? "ok" : "error"
func (p *Parser) parseTernaryExpr(condition ast.Expression) ast.Expression {
	// curToken is '?'
	// Parse the "then" expression with ternary precedence - 1 for right-associativity
	// This means a ? b ? c : d : e parses as a ? (b ? c : d) : e
	p.nextToken()
	thenExpr := p.parseExpression(ternaryPrec - 1)

	// Expect ':'
	if !p.peekTokenIs(token.COLON) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected ':' in ternary expression")
		return nil
	}
	p.nextToken() // consume ':'

	// Parse the "else" expression (also ternaryPrec - 1 for right-associativity)
	p.nextToken()
	elseExpr := p.parseExpression(ternaryPrec - 1)

	return &ast.TernaryExpr{
		Condition: condition,
		Then:      thenExpr,
		Else:      elseExpr,
	}
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

// parseSymbolToProc parses the &:method syntax for symbol-to-proc conversion.
// Example: &:upcase → creates a block/lambda that calls upcase on its argument
func (p *Parser) parseSymbolToProc() ast.Expression {
	// curToken is '&'
	p.nextToken() // move to the symbol token

	if !p.curTokenIs(token.SYMBOL) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected symbol after '&' (e.g., &:method)")
		return nil
	}

	// Symbol token literal is just the method name (lexer strips the colon)
	return &ast.SymbolToProcExpr{Method: p.curToken.Literal}
}

// isCallableExpr returns true if the expression can be called with command syntax.
// Only identifiers and selector expressions (method calls) support command syntax.
func (p *Parser) isCallableExpr(expr ast.Expression) bool {
	switch expr.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return true
	}
	return false
}

// canStartCommandArg returns true if the peek token can begin a command argument.
// This is used to detect command syntax (calls without parentheses).
func (p *Parser) canStartCommandArg() bool {
	t := p.peekToken.Type

	// Must have space before the argument
	if !p.peekToken.SpaceBefore {
		return false
	}

	switch t {
	// Literals and identifiers
	case token.IDENT, token.INT, token.FLOAT, token.STRING, token.HEREDOC, token.SYMBOL,
		token.TRUE, token.FALSE, token.NIL, token.SELF:
		return true
	// Grouping and collection literals (not LBRACE - that's handled as a block)
	case token.LPAREN, token.LBRACKET:
		return true
	// Instance variable
	case token.AT:
		return true
	// Prefix operators - only if followed immediately by their operand (no space after)
	// This handles `foo -1` (command with unary minus) vs `foo - 1` (binary)
	case token.MINUS, token.NOT:
		// For command syntax, treat prefix operators as arg starters
		// The SpaceBefore check above ensures there's space before the operator
		return true
	}
	return false
}

// looksLikeCommandArg checks if the peek token looks like a command argument
// in contexts where it could also be an infix operator.
// This handles ambiguous cases like:
//   - `foo -1` (command with unary minus) vs `foo - 1` (binary subtraction)
//   - `foo [1,2]` (command with array) vs `foo[1]` (index expression)
func (p *Parser) looksLikeCommandArg() bool {
	// Must have space before
	if !p.peekToken.SpaceBefore {
		return false
	}

	switch p.peekToken.Type {
	case token.MINUS, token.NOT:
		// For `-` or `not`: command syntax if space before but no space after operand
		// `foo -1` → command (space before -, no space after)
		// `foo - 1` → binary (space on both sides)
		// Use lexer lookahead to check the token after the operator
		state := p.l.SaveState()
		savedCur := p.curToken
		savedPeek := p.peekToken
		p.nextToken() // move to the operator
		nextHasSpace := p.peekToken.SpaceBefore
		p.l.RestoreState(state)
		p.curToken = savedCur
		p.peekToken = savedPeek
		return !nextHasSpace

	case token.LBRACKET:
		// `foo [1,2]` with space → array literal argument
		// `foo[1]` without space → index expression (handled by SpaceBefore check)
		return true

		// Note: LBRACE is NOT handled here - `foo { }` is a block, not a map argument
		// Use `foo({"a" => 1})` to pass a map explicitly
	}

	return false
}

// parseCommandCall parses a function call using command syntax (no parentheses).
// Example: `puts "hello"` becomes CallExpr{Func: puts, Args: ["hello"]}
func (p *Parser) parseCommandCall(fn ast.Expression) *ast.CallExpr {
	call := &ast.CallExpr{Func: fn}

	p.nextToken() // move to first argument token

	// Parse first argument at andPrec level
	// This includes ==, +, etc. but excludes and/or which terminate the argument list
	arg := p.parseExpression(andPrec)
	if arg == nil {
		return call
	}
	call.Args = append(call.Args, arg)

	// Parse additional comma-separated arguments
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next arg
		arg := p.parseExpression(andPrec)
		if arg != nil {
			call.Args = append(call.Args, arg)
		}
	}

	return call
}
