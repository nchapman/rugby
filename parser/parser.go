// Package parser implements the Rugby parser.
package parser

import (
	"fmt"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/token"
)

// Parser parses Rugby source code into an AST.
type Parser struct {
	l            *lexer.Lexer
	curToken     token.Token
	peekToken    token.Token
	errors       []ParseError
	commentIndex int // current position in lexer.Comments slice
}

// New creates a new Parser for the given lexer.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	// Note: comments are collected during lexing (as tokens are skipped)
	// and accessed via p.l.CollectComments() after parsing completes.
	// The p.leadingComments/p.trailingComment methods consume from p.comments,
	// which will be populated when we need it. For now, we defer collecting.
	return p
}

// Errors returns parse errors as strings for backwards compatibility.
func (p *Parser) Errors() []string {
	result := make([]string, len(p.errors))
	for i, e := range p.errors {
		result[i] = e.String()
	}
	return result
}

// ParseErrors returns structured parse errors.
func (p *Parser) ParseErrors() []ParseError {
	return p.errors
}

// nextToken advances the parser to the next token.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs checks if the current token is of the given type.
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs checks if the peek token is of the given type.
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// peekTokenAfterIs checks if the token immediately after peekToken matches the given type.
// Used for three-token lookahead patterns like @name = expr.
func (p *Parser) peekTokenAfterIs(t token.TokenType) bool {
	// Save lexer state and parser tokens
	lexerState := p.l.SaveState()
	savedCur := p.curToken
	savedPeek := p.peekToken

	// Advance to see the token after peekToken
	p.nextToken()
	p.nextToken()
	result := p.curToken.Type == t

	// Restore lexer state and parser tokens
	p.l.RestoreState(lexerState)
	p.curToken = savedCur
	p.peekToken = savedPeek
	return result
}

// skipNewlines skips over consecutive newline tokens.
func (p *Parser) skipNewlines() {
	for p.curTokenIs(token.NEWLINE) {
		p.nextToken()
	}
}

// peekAheadIsDotAfterNewlines checks if there's a DOT or AMPDOT token after skipping newlines.
// This is used for multi-line method chaining: expr\n.method()
func (p *Parser) peekAheadIsDotAfterNewlines() bool {
	// Save lexer state and parser tokens
	lexerState := p.l.SaveState()
	savedCur := p.curToken
	savedPeek := p.peekToken

	// Advance past the current peek (which should be NEWLINE)
	p.nextToken()

	// Skip all newlines
	for p.curTokenIs(token.NEWLINE) {
		p.nextToken()
	}

	// Check if we're now at a DOT or AMPDOT
	result := p.curTokenIs(token.DOT) || p.curTokenIs(token.AMPDOT)

	// Restore lexer state and parser tokens
	p.l.RestoreState(lexerState)
	p.curToken = savedCur
	p.peekToken = savedPeek

	return result
}

// isAssignmentPattern checks if we have a (ident = expr) assignment pattern.
// This is used to distinguish `if (x = getValue())` from `if (1..10).include?(5)`.
// Must be called when curToken is LPAREN.
func (p *Parser) isAssignmentPattern() bool {
	// Save lexer state and parser tokens
	lexerState := p.l.SaveState()
	savedCur := p.curToken
	savedPeek := p.peekToken

	// Look inside the parens: consume '(' then check for IDENT = pattern
	p.nextToken() // move past '('
	result := p.curTokenIs(token.IDENT) && p.peekTokenIs(token.ASSIGN)

	// Restore lexer state and parser tokens
	p.l.RestoreState(lexerState)
	p.curToken = savedCur
	p.peekToken = savedPeek

	return result
}

// isFunctionTypeAhead checks if the current position is at a function type.
// A function type is: () -> T or (T) -> U or (T, U) -> V
// Must be called when curToken is LPAREN.
// This is used to distinguish `-> () -> Int` (function type) from `-> (Int, String)` (tuple).
func (p *Parser) isFunctionTypeAhead() bool {
	// Save lexer state and parser tokens
	lexerState := p.l.SaveState()
	savedCur := p.curToken
	savedPeek := p.peekToken

	// Consume '(' and scan to matching ')'
	p.nextToken() // consume '('
	depth := 1
	for depth > 0 && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.LPAREN) {
			depth++
		} else if p.curTokenIs(token.RPAREN) {
			depth--
		}
		if depth > 0 {
			p.nextToken()
		}
	}

	// Now curToken should be ')'. Check if next is '->'
	var result bool
	if p.curTokenIs(token.RPAREN) {
		p.nextToken() // consume ')'
		result = p.curTokenIs(token.ARROW)
	}

	// Restore lexer state and parser tokens
	p.l.RestoreState(lexerState)
	p.curToken = savedCur
	p.peekToken = savedPeek

	return result
}

// isMultiAssignPattern checks if we have (ident, ident, ... = expr) pattern.
// Also handles splat patterns like (ident, *rest = expr) or (*head, ident = expr).
// This is used to distinguish multi-assignment from tuple literals.
// Must be called when curToken is IDENT and peekToken is COMMA.
func (p *Parser) isMultiAssignPattern() bool {
	// Save lexer state and parser tokens
	lexerState := p.l.SaveState()
	savedCur := p.curToken
	savedPeek := p.peekToken

	// Scan through comma-separated identifiers (and splat patterns) looking for '='
	// ident, ident, ... = expr  -> true
	// ident, *rest = expr       -> true
	// *head, ident = expr       -> true (but handled by different entry point)
	// ident, expr               -> false
	for {
		// Handle splat: *ident
		if p.curTokenIs(token.STAR) {
			p.nextToken() // move past '*'
			if !p.curTokenIs(token.IDENT) {
				break // not a valid pattern
			}
		}

		if !p.curTokenIs(token.IDENT) {
			break // not a valid pattern
		}
		p.nextToken() // move past ident

		// After ident, we should see either COMMA or ASSIGN
		if p.curTokenIs(token.ASSIGN) {
			// Found the = sign, this is a multi-assignment
			p.l.RestoreState(lexerState)
			p.curToken = savedCur
			p.peekToken = savedPeek
			return true
		}
		if p.curTokenIs(token.COMMA) {
			p.nextToken() // move past comma
			continue
		}
		// Something else (not comma or assign) after ident - not a multi-assign
		break
	}

	// Didn't find '=' after all identifiers
	p.l.RestoreState(lexerState)
	p.curToken = savedCur
	p.peekToken = savedPeek
	return false
}

// isSplatMultiAssignPattern checks if we have (*ident, ident, ... = expr) pattern.
// Must be called when curToken is STAR and peekToken is IDENT.
func (p *Parser) isSplatMultiAssignPattern() bool {
	// Save lexer state and parser tokens
	lexerState := p.l.SaveState()
	savedCur := p.curToken
	savedPeek := p.peekToken

	// Start by consuming * and first ident
	p.nextToken() // consume '*'
	if !p.curTokenIs(token.IDENT) {
		p.l.RestoreState(lexerState)
		p.curToken = savedCur
		p.peekToken = savedPeek
		return false
	}

	// Now use the same logic as isMultiAssignPattern
	for {
		if p.curTokenIs(token.STAR) {
			p.nextToken() // move past '*'
			if !p.curTokenIs(token.IDENT) {
				break
			}
		}

		if !p.curTokenIs(token.IDENT) {
			break
		}
		p.nextToken() // move past ident

		if p.curTokenIs(token.ASSIGN) {
			p.l.RestoreState(lexerState)
			p.curToken = savedCur
			p.peekToken = savedPeek
			return true
		}
		if p.curTokenIs(token.COMMA) {
			p.nextToken() // move past comma
			continue
		}
		break
	}

	p.l.RestoreState(lexerState)
	p.curToken = savedCur
	p.peekToken = savedPeek
	return false
}

// parseBlocksAndChaining handles blocks and method chaining after an expression.
// This enables:
//   - arr.select { |x| }.map { |x| } (inline chaining)
//   - Multi-line method chains with newlines before DOT/AMPDOT
//
// Returns the (possibly modified) expression with blocks and chained methods attached.
func (p *Parser) parseBlocksAndChaining(expr ast.Expression) ast.Expression {
	for {
		// Check for block: expr do |params| ... end  OR  expr {|params| ... }
		// Only attach blocks to method calls (CallExpr or SelectorExpr), not plain identifiers
		// This prevents `puts {}` from treating {} as a block - it should be a map literal
		if p.peekTokenIs(token.DO) || p.peekTokenIs(token.LBRACE) {
			switch e := expr.(type) {
			case *ast.CallExpr:
				expr = p.parseBlockCall(expr)
				continue
			case *ast.Ident:
				// Check for struct literal pattern: TypeName{field: value, ...}
				// where TypeName is PascalCase (starts with uppercase)
				if p.peekTokenIs(token.LBRACE) && p.looksLikeStructLiteral(e) {
					expr = p.parseStructLiteral(e)
					continue
				}
				// Plain identifier with {} - fall through to let caller handle as map literal
			case *ast.SelectorExpr:
				// Check for Go struct literal pattern: pkg.Type{} where pkg is lowercase
				// and Type is PascalCase. Empty {} is a Go zero-value initializer, not a block.
				if p.peekTokenIs(token.LBRACE) && p.looksLikeGoStructLiteral(e) {
					expr = p.parseGoStructLiteral(e)
					continue
				}
				expr = p.parseBlockCall(expr)
				continue
			}
			// Not a valid block target - break to let caller handle LBRACE as map literal
			break
		}
		// Check for trailing lambda: expr -> (params) { ... }
		if p.peekTokenIs(token.ARROW) {
			expr = p.parseTrailingLambda(expr)
			continue
		}
		// Check for newline-continued method chain after a block
		if p.peekTokenIs(token.NEWLINE) && p.peekAheadIsDotAfterNewlines() {
			p.nextToken() // curToken is now NEWLINE
			p.skipNewlines()
			// curToken should now be DOT or AMPDOT
			if p.curTokenIs(token.DOT) {
				expr = p.parseSelectorExpr(expr)
				continue
			}
			if p.curTokenIs(token.AMPDOT) {
				expr = p.parseSafeNavExpr(expr)
				continue
			}
			// Defensive: if we get here, something unexpected happened
			break
		}
		// Check for same-line method chain after a block
		if p.peekTokenIs(token.DOT) {
			p.nextToken()
			expr = p.parseSelectorExpr(expr)
			continue
		}
		if p.peekTokenIs(token.AMPDOT) {
			p.nextToken()
			expr = p.parseSafeNavExpr(expr)
			continue
		}
		break
	}
	return expr
}

// looksLikeGoStructLiteral checks if a selector expression looks like a Go type constructor
// (e.g., sync.WaitGroup, bytes.Buffer, etc.) where the package is lowercase and type is PascalCase.
// This is a heuristic to distinguish Go struct literals from Ruby blocks.
func (p *Parser) looksLikeGoStructLiteral(sel *ast.SelectorExpr) bool {
	// Check if we have pkg.Type pattern
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	// Package should be lowercase (Go convention)
	pkgName := ident.Name
	if len(pkgName) == 0 || pkgName[0] < 'a' || pkgName[0] > 'z' {
		return false
	}

	// Type should be PascalCase (Go exported type)
	typeName := sel.Sel
	if len(typeName) == 0 || typeName[0] < 'A' || typeName[0] > 'Z' {
		return false
	}

	// Check if {} is empty by looking ahead
	// Save lexer state for lookahead
	state := p.l.SaveState()
	savedCur := p.curToken
	savedPeek := p.peekToken

	p.nextToken() // move to '{'
	p.nextToken() // move past '{'

	isEmpty := p.curTokenIs(token.RBRACE)

	// Restore state
	p.l.RestoreState(state)
	p.curToken = savedCur
	p.peekToken = savedPeek

	return isEmpty
}

// parseGoStructLiteral parses a Go struct literal like sync.WaitGroup{}
// Returns a GoStructLit AST node.
func (p *Parser) parseGoStructLiteral(sel *ast.SelectorExpr) ast.Expression {
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		// Should never happen since looksLikeGoStructLiteral already checked
		return sel
	}
	line := ident.Line
	col := ident.Column

	p.nextToken() // move to '{'
	p.nextToken() // consume '{'

	if !p.curTokenIs(token.RBRACE) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '}' in Go struct literal")
		return sel
	}

	return &ast.GoStructLit{
		Package: ident.Name,
		Type:    sel.Sel,
		Line:    line,
		Column:  col,
	}
}

// looksLikeStructLiteral checks if an identifier looks like a struct type name
// (PascalCase, starting with uppercase letter).
func (p *Parser) looksLikeStructLiteral(ident *ast.Ident) bool {
	name := ident.Name
	if len(name) == 0 {
		return false
	}
	// Must start with uppercase letter (PascalCase type name)
	return name[0] >= 'A' && name[0] <= 'Z'
}

// parseStructLiteral parses a struct literal like Point{x: 10, y: 20}
// Returns a StructLit AST node.
func (p *Parser) parseStructLiteral(ident *ast.Ident) ast.Expression {
	line := ident.Line

	p.nextToken() // move to '{'
	p.nextToken() // consume '{'

	var fields []*ast.StructLitField

	// Parse field initializers: field: value, field: value, ...
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.RBRACE) {
			break
		}

		// Expect field name
		if !p.curTokenIs(token.IDENT) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected field name in struct literal")
			return ident
		}
		fieldName := p.curToken.Literal
		p.nextToken()

		// Expect ':'
		if !p.curTokenIs(token.COLON) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected ':' after field name in struct literal")
			return ident
		}
		p.nextToken()

		// Parse field value
		value := p.parseExpression(lowest)
		if value == nil {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected value after ':' in struct literal")
			return ident
		}

		fields = append(fields, &ast.StructLitField{
			Name:  fieldName,
			Value: value,
		})

		p.nextToken() // move past value

		// Skip optional comma or newline
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.skipNewlines()
	}

	if !p.curTokenIs(token.RBRACE) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '}' to close struct literal")
		return ident
	}

	return &ast.StructLit{
		Name:   ident.Name,
		Fields: fields,
		Line:   line,
	}
}

// leadingComments returns a group of consecutive comments that end immediately
// before the given line. Multi-line doc comments are collected as a group.
// Returns nil if no leading comments.
func (p *Parser) leadingComments(line int) *ast.CommentGroup {
	comments := p.l.Comments
	if p.commentIndex >= len(comments) {
		return nil
	}

	// Skip comments that are too early (free-floating)
	for p.commentIndex < len(comments) && comments[p.commentIndex].Line < line-1 {
		// Check if this could be the start of a consecutive group ending at line-1
		startIdx := p.commentIndex
		endIdx := startIdx

		// Find the end of consecutive comments
		for endIdx+1 < len(comments) && comments[endIdx+1].Line == comments[endIdx].Line+1 {
			endIdx++
		}

		// If this group ends at line-1, collect it
		if comments[endIdx].Line == line-1 {
			var result []*ast.Comment
			for i := startIdx; i <= endIdx; i++ {
				result = append(result, &ast.Comment{
					Text: comments[i].Text,
					Line: comments[i].Line,
				})
			}
			p.commentIndex = endIdx + 1
			return &ast.CommentGroup{List: result}
		}

		// Not the right group, skip it
		p.commentIndex = endIdx + 1
	}

	// Check for comments at exactly line-1
	if p.commentIndex < len(comments) && comments[p.commentIndex].Line == line-1 {
		// Collect consecutive comments ending at line-1
		startIdx := p.commentIndex
		endIdx := startIdx
		for endIdx+1 < len(comments) && comments[endIdx+1].Line == comments[endIdx].Line+1 &&
			comments[endIdx+1].Line <= line-1 {
			endIdx++
		}

		var result []*ast.Comment
		for i := startIdx; i <= endIdx; i++ {
			result = append(result, &ast.Comment{
				Text: comments[i].Text,
				Line: comments[i].Line,
			})
		}
		p.commentIndex = endIdx + 1
		return &ast.CommentGroup{List: result}
	}

	return nil
}

// ParseProgram parses the entire Rugby program.
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
		case token.PUB:
			pubLine := p.curToken.Line
			doc := p.leadingComments(pubLine)
			p.nextToken() // consume 'pub'
			switch p.curToken.Type {
			case token.DEF:
				if fn := p.parseFuncDeclWithDoc(doc); fn != nil {
					fn.Pub = true
					program.Declarations = append(program.Declarations, fn)
				}
			case token.CLASS:
				if cls := p.parseClassDeclWithDoc(doc); cls != nil {
					cls.Pub = true
					program.Declarations = append(program.Declarations, cls)
				}
			case token.INTERFACE:
				if iface := p.parseInterfaceDeclWithDoc(doc); iface != nil {
					iface.Pub = true
					program.Declarations = append(program.Declarations, iface)
				}
			case token.MODULE:
				if mod := p.parseModuleDeclWithDoc(doc); mod != nil {
					mod.Pub = true
					program.Declarations = append(program.Declarations, mod)
				}
			case token.TYPE:
				if typeAlias := p.parseTypeAliasDeclWithDoc(doc); typeAlias != nil {
					typeAlias.Pub = true
					program.Declarations = append(program.Declarations, typeAlias)
				}
			case token.ENUM:
				if enumDecl := p.parseEnumDeclWithDoc(doc); enumDecl != nil {
					enumDecl.Pub = true
					program.Declarations = append(program.Declarations, enumDecl)
				}
			case token.STRUCT:
				if structDecl := p.parseStructDeclWithDoc(doc); structDecl != nil {
					structDecl.Pub = true
					program.Declarations = append(program.Declarations, structDecl)
				}
			default:
				p.errorAt(p.curToken.Line, p.curToken.Column, "'pub' must be followed by 'def', 'class', 'interface', 'module', 'type', 'enum', or 'struct'")
				p.nextToken()
			}
		case token.DEF:
			if fn := p.parseFuncDecl(); fn != nil {
				program.Declarations = append(program.Declarations, fn)
			}
		case token.CLASS:
			if cls := p.parseClassDecl(); cls != nil {
				program.Declarations = append(program.Declarations, cls)
			}
		case token.INTERFACE:
			if iface := p.parseInterfaceDecl(); iface != nil {
				program.Declarations = append(program.Declarations, iface)
			}
		case token.MODULE:
			if mod := p.parseModuleDecl(); mod != nil {
				program.Declarations = append(program.Declarations, mod)
			}
		case token.TYPE:
			if typeAlias := p.parseTypeAliasDecl(); typeAlias != nil {
				program.Declarations = append(program.Declarations, typeAlias)
			}
		case token.ENUM:
			if enumDecl := p.parseEnumDecl(); enumDecl != nil {
				program.Declarations = append(program.Declarations, enumDecl)
			}
		case token.STRUCT:
			if structDecl := p.parseStructDecl(); structDecl != nil {
				program.Declarations = append(program.Declarations, structDecl)
			}
		case token.CONST:
			if constDecl := p.parseConstDecl(); constDecl != nil {
				program.Declarations = append(program.Declarations, constDecl)
			}
		case token.DESCRIBE:
			if desc := p.parseDescribeStmt(); desc != nil {
				program.Declarations = append(program.Declarations, desc)
			}
		case token.TEST:
			if test := p.parseTestStmt(); test != nil {
				program.Declarations = append(program.Declarations, test)
			}
		case token.TABLE:
			if table := p.parseTableStmt(); table != nil {
				program.Declarations = append(program.Declarations, table)
			}
		default:
			// Top-level statement
			if stmt := p.parseStatement(); stmt != nil {
				program.Declarations = append(program.Declarations, stmt)
			} else {
				p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("unexpected token %s at top level", p.curToken.Type))
				p.nextToken()
			}
		}
	}

	// Collect all comments after parsing is complete
	program.Comments = p.l.CollectComments()

	return program
}
