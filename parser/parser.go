package parser

import (
	"fmt"
	"strconv"
	"strings"

	"rugby/ast"
	"rugby/lexer"
	"rugby/token"
)

// Precedence levels
const (
	_ int = iota
	LOWEST
	RANGE       // .., ... (ranges)
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
	token.DOTDOT:    RANGE,
	token.TRIPLEDOT: RANGE,
	token.OR:        OR_PREC,
	token.AND:       AND_PREC,
	token.EQ:        EQUALS,
	token.NE:        EQUALS,
	token.LT:        LESSGREATER,
	token.GT:        LESSGREATER,
	token.LE:        LESSGREATER,
	token.GE:        LESSGREATER,
	token.PLUS:      SUM,
	token.MINUS:     SUM,
	token.STAR:      PRODUCT,
	token.SLASH:     PRODUCT,
	token.PERCENT:   PRODUCT,
	token.LPAREN:    CALL,
	token.LBRACKET:  CALL, // array indexing has same precedence as function calls
	token.DOT:       MEMBER,
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
		case token.PUB:
			p.nextToken() // consume 'pub'
			switch p.curToken.Type {
			case token.DEF:
				if fn := p.parseFuncDecl(); fn != nil {
					fn.Pub = true
					program.Declarations = append(program.Declarations, fn)
				}
			case token.CLASS:
				if cls := p.parseClassDecl(); cls != nil {
					cls.Pub = true
					program.Declarations = append(program.Declarations, cls)
				}
			case token.INTERFACE:
				if iface := p.parseInterfaceDecl(); iface != nil {
					iface.Pub = true
					program.Declarations = append(program.Declarations, iface)
				}
			default:
				p.errors = append(p.errors, fmt.Sprintf("%d:%d: 'pub' must be followed by 'def', 'class', or 'interface'",
					p.curToken.Line, p.curToken.Column))
				p.nextToken()
			}
		case token.DEF:
			if fn := p.parseFuncDecl(); fn != nil {
				program.Declarations = append(program.Declarations, fn)
			}
		case token.CLASS:
			if cls := p.parseClassDecl(); cls != nil {
				// Validate: pub def inside non-pub class is an error (spec 9.2)
				for _, method := range cls.Methods {
					if method.Pub {
						p.errors = append(p.errors, fmt.Sprintf("'pub def %s' not allowed in non-pub class '%s'", method.Name, cls.Name))
					}
				}
				program.Declarations = append(program.Declarations, cls)
			}
		case token.INTERFACE:
			if iface := p.parseInterfaceDecl(); iface != nil {
				program.Declarations = append(program.Declarations, iface)
			}
		default:
			// Parse top-level statements (bare script support)
			if stmt := p.parseStatement(); stmt != nil {
				program.Declarations = append(program.Declarations, stmt)
			} else {
				// Skip token if parseStatement returned nil to avoid infinite loop
				p.errors = append(p.errors, fmt.Sprintf("%d:%d: unexpected token %s at top level",
					p.curToken.Line, p.curToken.Column, p.curToken.Type))
				p.nextToken()
			}
		}
	}

	return program
}

func (p *Parser) parseImport() *ast.ImportDecl {
	p.nextToken() // consume 'import'

	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected package path after import",
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	imp := &ast.ImportDecl{Path: p.curToken.Literal}
	p.nextToken()

	// Check for optional alias: import foo/bar as baz
	if p.curTokenIs(token.AS) {
		p.nextToken() // consume 'as'
		if !p.curTokenIs(token.IDENT) {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected alias name after 'as'",
				p.curToken.Line, p.curToken.Column))
			return nil
		}
		imp.Alias = p.curToken.Literal
		p.nextToken()
	}

	return imp
}

// parseTypedParam parses a parameter with optional type annotation: name or name : Type
func (p *Parser) parseTypedParam(seen map[string]bool) *ast.Param {
	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected parameter name",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	name := p.curToken.Literal
	if seen[name] {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: duplicate parameter name %q",
			p.curToken.Line, p.curToken.Column, name))
		return nil
	}
	seen[name] = true
	p.nextToken() // consume identifier

	// Check for optional type annotation: name : Type or name : Type?
	var paramType string
	if p.curTokenIs(token.COLON) {
		p.nextToken() // consume ':'
		if !p.curTokenIs(token.IDENT) {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected type after ':'",
				p.curToken.Line, p.curToken.Column))
			return nil
		}
		paramType = p.parseTypeName()
	}

	return &ast.Param{Name: name, Type: paramType}
}

// parseTypeName parses a type name (e.g., "Int", "String?")
// The ? suffix is already included in the identifier by the lexer.
// Expects current token to be IDENT. Consumes the type.
func (p *Parser) parseTypeName() string {
	typeName := p.curToken.Literal
	p.nextToken() // consume type
	return typeName
}

func (p *Parser) parseFuncDecl() *ast.FuncDecl {
	p.nextToken() // consume 'def'

	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected function name after def",
			p.curToken.Line, p.curToken.Column))
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
			param := p.parseTypedParam(seen)
			if param == nil {
				return nil
			}
			fn.Params = append(fn.Params, param)

			// Additional parameters
			for p.curTokenIs(token.COMMA) {
				p.nextToken() // consume ','
				// Allow trailing comma
				if p.curTokenIs(token.RPAREN) {
					break
				}
				param := p.parseTypedParam(seen)
				if param == nil {
					return nil
				}
				fn.Params = append(fn.Params, param)
			}
		}

		if !p.curTokenIs(token.RPAREN) {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected ')' after parameters",
				p.curToken.Line, p.curToken.Column))
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
				p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected type in return type list",
					p.curToken.Line, p.curToken.Column))
				return nil
			}
			fn.ReturnTypes = append(fn.ReturnTypes, p.parseTypeName())

			for p.curTokenIs(token.COMMA) {
				p.nextToken() // consume ','
				// Allow trailing comma
				if p.curTokenIs(token.RPAREN) {
					break
				}
				if !p.curTokenIs(token.IDENT) {
					p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected type after comma",
						p.curToken.Line, p.curToken.Column))
					return nil
				}
				fn.ReturnTypes = append(fn.ReturnTypes, p.parseTypeName())
			}

			if !p.curTokenIs(token.RPAREN) {
				p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected ')' after return types",
					p.curToken.Line, p.curToken.Column))
				return nil
			}
			p.nextToken() // consume ')'
		} else if p.curTokenIs(token.IDENT) {
			// Single return type
			fn.ReturnTypes = append(fn.ReturnTypes, p.parseTypeName())
		} else {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected type after ->",
				p.curToken.Line, p.curToken.Column))
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
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected 'end' to close function",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	p.nextToken() // consume 'end'

	return fn
}

func (p *Parser) parseClassDecl() *ast.ClassDecl {
	p.nextToken() // consume 'class'

	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected class name after 'class'",
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	cls := &ast.ClassDecl{Name: p.curToken.Literal}
	p.nextToken()

	// Parse optional embedded types: class Service < Logger, Authenticator
	if p.curTokenIs(token.LT) {
		p.nextToken() // consume '<'
		if !p.curTokenIs(token.IDENT) {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected type name after '<'",
				p.curToken.Line, p.curToken.Column))
			return nil
		}
		cls.Embeds = append(cls.Embeds, p.curToken.Literal)
		p.nextToken()

		// Parse additional embedded types separated by commas
		for p.curTokenIs(token.COMMA) {
			p.nextToken() // consume ','
			if !p.curTokenIs(token.IDENT) {
				p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected type name after ','",
					p.curToken.Line, p.curToken.Column))
				return nil
			}
			cls.Embeds = append(cls.Embeds, p.curToken.Literal)
			p.nextToken()
		}
	}

	p.skipNewlines()

	// Parse methods until 'end'
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}

		switch p.curToken.Type {
		case token.PUB:
			p.nextToken() // consume 'pub'
			if p.curTokenIs(token.DEF) {
				if method := p.parseMethodDecl(); method != nil {
					method.Pub = true
					cls.Methods = append(cls.Methods, method)
				}
			} else {
				p.errors = append(p.errors, fmt.Sprintf("%d:%d: 'pub' in class body must be followed by 'def'",
					p.curToken.Line, p.curToken.Column))
				p.nextToken()
			}
		case token.DEF:
			if method := p.parseMethodDecl(); method != nil {
				cls.Methods = append(cls.Methods, method)
			}
		default:
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: unexpected token %s in class body, expected 'def', 'pub def', or 'end'",
				p.curToken.Line, p.curToken.Column, p.curToken.Type))
			p.nextToken()
		}
	}

	if !p.curTokenIs(token.END) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected 'end' to close class",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	p.nextToken() // consume 'end'

	// Extract fields from initialize method
	for _, method := range cls.Methods {
		if method.Name == "initialize" {
			cls.Fields = extractFields(method)
			break
		}
	}

	return cls
}

func (p *Parser) parseInterfaceDecl() *ast.InterfaceDecl {
	p.nextToken() // consume 'interface'

	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected interface name after 'interface'",
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	iface := &ast.InterfaceDecl{Name: p.curToken.Literal}
	p.nextToken()
	p.skipNewlines()

	// Parse method signatures until 'end'
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}

		if p.curTokenIs(token.DEF) {
			if sig := p.parseMethodSig(); sig != nil {
				iface.Methods = append(iface.Methods, sig)
			}
		} else {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: unexpected token %s in interface body, expected 'def' or 'end'",
				p.curToken.Line, p.curToken.Column, p.curToken.Type))
			p.nextToken()
		}
	}

	if !p.curTokenIs(token.END) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected 'end' to close interface",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	p.nextToken() // consume 'end'

	return iface
}

// parseMethodSig parses a method signature (no body) for interfaces
func (p *Parser) parseMethodSig() *ast.MethodSig {
	p.nextToken() // consume 'def'

	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected method name after def",
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	sig := &ast.MethodSig{Name: p.curToken.Literal}
	p.nextToken()

	// Parse optional parameter list
	if p.curTokenIs(token.LPAREN) {
		p.nextToken() // consume '('
		seen := make(map[string]bool)

		// Parse parameters until ')'
		if !p.curTokenIs(token.RPAREN) {
			// First parameter
			param := p.parseTypedParam(seen)
			if param == nil {
				return nil
			}
			sig.Params = append(sig.Params, param)

			// Additional parameters
			for p.curTokenIs(token.COMMA) {
				p.nextToken() // consume ','
				// Allow trailing comma
				if p.curTokenIs(token.RPAREN) {
					break
				}
				param := p.parseTypedParam(seen)
				if param == nil {
					return nil
				}
				sig.Params = append(sig.Params, param)
			}
		}

		if !p.curTokenIs(token.RPAREN) {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected ')' after parameters",
				p.curToken.Line, p.curToken.Column))
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
				p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected type in return type list",
					p.curToken.Line, p.curToken.Column))
				return nil
			}
			sig.ReturnTypes = append(sig.ReturnTypes, p.parseTypeName())

			for p.curTokenIs(token.COMMA) {
				p.nextToken() // consume ','
				// Allow trailing comma
				if p.curTokenIs(token.RPAREN) {
					break
				}
				if !p.curTokenIs(token.IDENT) {
					p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected type after comma",
						p.curToken.Line, p.curToken.Column))
					return nil
				}
				sig.ReturnTypes = append(sig.ReturnTypes, p.parseTypeName())
			}

			if !p.curTokenIs(token.RPAREN) {
				p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected ')' after return types",
					p.curToken.Line, p.curToken.Column))
				return nil
			}
			p.nextToken() // consume ')'
		} else if p.curTokenIs(token.IDENT) {
			// Single return type
			sig.ReturnTypes = append(sig.ReturnTypes, p.parseTypeName())
		} else {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected type after ->",
				p.curToken.Line, p.curToken.Column))
			return nil
		}
	}

	p.skipNewlines()

	return sig
}

// extractFields extracts field declarations from instance variable assignments in a method.
// It infers field types from typed parameters when @field = param_name.
func extractFields(method *ast.MethodDecl) []*ast.FieldDecl {
	seen := make(map[string]bool)
	var fields []*ast.FieldDecl

	// Build a map of parameter names to their types
	paramTypes := make(map[string]string)
	for _, param := range method.Params {
		if param.Type != "" {
			paramTypes[param.Name] = param.Type
		}
	}

	for _, stmt := range method.Body {
		if assign, ok := stmt.(*ast.InstanceVarAssign); ok {
			if !seen[assign.Name] {
				seen[assign.Name] = true

				// Infer type from assignment source
				fieldType := ""
				if ident, ok := assign.Value.(*ast.Ident); ok {
					// If assigned from a parameter, use its type
					if pt, ok := paramTypes[ident.Name]; ok {
						fieldType = pt
					}
				}

				fields = append(fields, &ast.FieldDecl{Name: assign.Name, Type: fieldType})
			}
		}
		// Also handle ||= assignments
		if assign, ok := stmt.(*ast.InstanceVarOrAssign); ok {
			if !seen[assign.Name] {
				seen[assign.Name] = true
				// Type inference for ||= is harder since value is conditional
				// For now, use empty type (interface{})
				fields = append(fields, &ast.FieldDecl{Name: assign.Name, Type: ""})
			}
		}
	}

	return fields
}

func (p *Parser) parseMethodDecl() *ast.MethodDecl {
	p.nextToken() // consume 'def'

	// Accept regular identifiers or operator tokens (== for custom equality)
	var methodName string
	switch p.curToken.Type {
	case token.IDENT:
		methodName = p.curToken.Literal
	case token.EQ:
		methodName = "=="
	default:
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected method name after def",
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	method := &ast.MethodDecl{Name: methodName}
	p.nextToken()

	// Parse optional parameter list
	if p.curTokenIs(token.LPAREN) {
		p.nextToken() // consume '('
		seen := make(map[string]bool)

		// Parse parameters until ')'
		if !p.curTokenIs(token.RPAREN) {
			// First parameter
			param := p.parseTypedParam(seen)
			if param == nil {
				return nil
			}
			method.Params = append(method.Params, param)

			// Additional parameters
			for p.curTokenIs(token.COMMA) {
				p.nextToken() // consume ','
				// Allow trailing comma
				if p.curTokenIs(token.RPAREN) {
					break
				}
				param := p.parseTypedParam(seen)
				if param == nil {
					return nil
				}
				method.Params = append(method.Params, param)
			}
		}

		if !p.curTokenIs(token.RPAREN) {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected ')' after parameters",
				p.curToken.Line, p.curToken.Column))
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
				p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected type in return type list",
					p.curToken.Line, p.curToken.Column))
				return nil
			}
			method.ReturnTypes = append(method.ReturnTypes, p.parseTypeName())

			for p.curTokenIs(token.COMMA) {
				p.nextToken() // consume ','
				// Allow trailing comma
				if p.curTokenIs(token.RPAREN) {
					break
				}
				if !p.curTokenIs(token.IDENT) {
					p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected type after comma",
						p.curToken.Line, p.curToken.Column))
					return nil
				}
				method.ReturnTypes = append(method.ReturnTypes, p.parseTypeName())
			}

			if !p.curTokenIs(token.RPAREN) {
				p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected ')' after return types",
					p.curToken.Line, p.curToken.Column))
				return nil
			}
			p.nextToken() // consume ')'
		} else if p.curTokenIs(token.IDENT) {
			// Single return type
			method.ReturnTypes = append(method.ReturnTypes, p.parseTypeName())
		} else {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected type after ->",
				p.curToken.Line, p.curToken.Column))
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
			method.Body = append(method.Body, stmt)
		}
	}

	if !p.curTokenIs(token.END) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected 'end' to close method",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	p.nextToken() // consume 'end'

	return method
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.IF:
		return p.parseIfStmt()
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
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: block can only follow a method call",
			p.curToken.Line, p.curToken.Column))
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
				p.errors = append(p.errors, fmt.Sprintf("%d:%d: duplicate block parameter name %q",
					p.curToken.Line, p.curToken.Column, name))
				p.skipTo(terminator)
				return nil
			}
			seen[name] = true
			block.Params = append(block.Params, name)
			p.nextToken()

			for p.curTokenIs(token.COMMA) {
				p.nextToken() // move past ','
				if !p.curTokenIs(token.IDENT) {
					p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected parameter name after comma",
						p.curToken.Line, p.curToken.Column))
					p.skipTo(terminator)
					return nil
				}
				name := p.curToken.Literal
				if seen[name] {
					p.errors = append(p.errors, fmt.Sprintf("%d:%d: duplicate block parameter name %q",
						p.curToken.Line, p.curToken.Column, name))
					p.skipTo(terminator)
					return nil
				}
				seen[name] = true
				block.Params = append(block.Params, name)
				p.nextToken()
			}
		}

		if !p.curTokenIs(token.PIPE) {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected '|' after block parameters",
				p.curToken.Line, p.curToken.Column))
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
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected '%s' to close block",
			p.curToken.Line, p.curToken.Column, closer))
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

	// Check for optional type annotation: x : Type = value or x : Type? = value
	var typeAnnotation string
	if p.curTokenIs(token.COLON) {
		p.nextToken() // consume ':'
		if !p.curTokenIs(token.IDENT) {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected type after ':'",
				p.curToken.Line, p.curToken.Column))
			return nil
		}
		typeAnnotation = p.parseTypeName()
	}

	if !p.curTokenIs(token.ASSIGN) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected '=' in assignment",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	p.nextToken() // consume '='

	value := p.parseExpression(LOWEST)

	// Check for block: expr.method do |x| ... end  OR  expr.method {|x| ... }
	if p.peekTokenIs(token.DO) || p.peekTokenIs(token.LBRACE) {
		value = p.parseBlockCall(value)
	}

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.AssignStmt{Name: name, Type: typeAnnotation, Value: value}
}

func (p *Parser) parseIfStmt() *ast.IfStmt {
	p.nextToken() // consume 'if'

	stmt := &ast.IfStmt{}

	// Check for assignment-in-condition pattern: if (name = expr)
	if p.curTokenIs(token.LPAREN) {
		// Peek ahead to see if it's (ident = expr) pattern
		p.nextToken() // consume '('
		if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.ASSIGN) {
			// Assignment pattern detected
			stmt.AssignName = p.curToken.Literal
			p.nextToken() // consume ident
			p.nextToken() // consume '='
			stmt.AssignExpr = p.parseExpression(LOWEST)
			p.nextToken() // move past expression
			if !p.curTokenIs(token.RPAREN) {
				p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected ')' after assignment expression",
					p.curToken.Line, p.curToken.Column))
				return nil
			}
			p.nextToken() // consume ')'
		} else {
			// Regular parenthesized condition - parse as grouped expression
			// We already consumed '(', so parse the inner expression
			stmt.Cond = p.parseExpression(LOWEST)
			p.nextToken() // move past expression
			if p.curTokenIs(token.RPAREN) {
				p.nextToken() // consume ')'
			}
		}
	} else {
		stmt.Cond = p.parseExpression(LOWEST)
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
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected 'end' to close if",
			p.curToken.Line, p.curToken.Column))
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
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected 'end' to close while",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

func (p *Parser) parseForStmt() *ast.ForStmt {
	p.nextToken() // consume 'for'

	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected variable name after 'for'",
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	stmt := &ast.ForStmt{Var: p.curToken.Literal}
	p.nextToken() // consume variable name

	if !p.curTokenIs(token.IN) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected 'in' after loop variable",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	p.nextToken() // consume 'in'

	stmt.Iterable = p.parseExpression(LOWEST)
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
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected 'end' to close for",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	p.nextToken() // consume 'end'
	p.skipNewlines()

	return stmt
}

func (p *Parser) parseBreakStmt() *ast.BreakStmt {
	p.nextToken() // consume 'break'
	p.skipNewlines()
	return &ast.BreakStmt{}
}

func (p *Parser) parseNextStmt() *ast.NextStmt {
	p.nextToken() // consume 'next'
	p.skipNewlines()
	return &ast.NextStmt{}
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
	case token.NIL:
		left = p.parseNilLiteral()
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
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected ')' after arguments",
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	return call
}

func (p *Parser) parseIntLiteral() ast.Expression {
	val, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: invalid integer %s",
			p.curToken.Line, p.curToken.Column, p.curToken.Literal))
		return nil
	}
	return &ast.IntLit{Value: val}
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	val, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: invalid float %s",
			p.curToken.Line, p.curToken.Column, p.curToken.Literal))
		return nil
	}
	return &ast.FloatLit{Value: val}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	value := p.curToken.Literal

	// Check if string contains interpolation
	if !strings.Contains(value, "#{") {
		return &ast.StringLit{Value: value}
	}

	// Parse interpolated string
	return p.parseInterpolatedString(value)
}

func (p *Parser) parseInterpolatedString(value string) ast.Expression {
	var parts []interface{}
	i := 0

	for i < len(value) {
		// Look for #{
		hashIdx := strings.Index(value[i:], "#{")
		if hashIdx == -1 {
			// No more interpolations, add remaining string
			if i < len(value) {
				parts = append(parts, value[i:])
			}
			break
		}

		// Add string part before #{
		if hashIdx > 0 {
			parts = append(parts, value[i:i+hashIdx])
		}

		// Find matching } using brace counting.
		// NOTE: This doesn't handle braces inside string literals within the expression.
		// e.g., "#{format("{}")}" will fail. Use a variable for complex expressions.
		exprStart := i + hashIdx + 2 // skip #{
		braceCount := 1
		exprEnd := exprStart

		for exprEnd < len(value) && braceCount > 0 {
			if value[exprEnd] == '{' {
				braceCount++
			} else if value[exprEnd] == '}' {
				braceCount--
			}
			exprEnd++
		}

		if braceCount != 0 {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: unterminated string interpolation",
				p.curToken.Line, p.curToken.Column))
			return &ast.StringLit{Value: value}
		}

		// Extract expression content (excluding the closing })
		exprContent := value[exprStart : exprEnd-1]

		// Parse the expression
		exprLexer := lexer.New(exprContent)
		exprParser := New(exprLexer)
		expr := exprParser.parseExpression(LOWEST)

		if len(exprParser.errors) > 0 {
			p.errors = append(p.errors, exprParser.errors...)
			return &ast.StringLit{Value: value}
		}

		if expr != nil {
			parts = append(parts, expr)
		}

		i = exprEnd
	}

	// If only one part and it's a string, return StringLit
	if len(parts) == 1 {
		if s, ok := parts[0].(string); ok {
			return &ast.StringLit{Value: s}
		}
	}

	return &ast.InterpolatedString{Parts: parts}
}

func (p *Parser) parseBoolLiteral() ast.Expression {
	return &ast.BoolLit{Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLit{}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	arr := &ast.ArrayLit{}

	p.nextToken() // consume '['

	// Handle empty array
	if p.curTokenIs(token.RBRACKET) {
		return arr
	}

	// Parse first element
	elem := p.parseExpression(LOWEST)
	if elem == nil {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected expression in array literal",
			p.curToken.Line, p.curToken.Column))
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
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected expression after comma in array literal",
				p.curToken.Line, p.curToken.Column))
			return nil
		}
		arr.Elements = append(arr.Elements, elem)
	}

	// Move past last element to ']'
	p.nextToken()

	if !p.curTokenIs(token.RBRACKET) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected ']' after array elements",
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	return arr
}

func (p *Parser) parseIndexExpr(left ast.Expression) ast.Expression {
	// curToken is '['
	p.nextToken() // move past '[' to the index expression

	index := p.parseExpression(LOWEST)
	if index == nil {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected index expression",
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	// Move past index expression to ']'
	p.nextToken()

	if !p.curTokenIs(token.RBRACKET) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected ']' after index",
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	return &ast.IndexExpr{Left: left, Index: index}
}

func (p *Parser) parseMapLiteral() ast.Expression {
	mapLit := &ast.MapLit{}

	p.nextToken() // consume '{'

	// Handle empty map
	if p.curTokenIs(token.RBRACE) {
		return mapLit
	}

	// Parse first entry
	entry, ok := p.parseMapEntry()
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

		entry, ok := p.parseMapEntry()
		if !ok {
			return nil
		}
		mapLit.Entries = append(mapLit.Entries, entry)
	}

	// Move past last value to '}'
	p.nextToken()

	if !p.curTokenIs(token.RBRACE) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected '}' after map entries",
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	return mapLit
}

func (p *Parser) parseMapEntry() (ast.MapEntry, bool) {
	key := p.parseExpression(LOWEST)
	if key == nil {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected key in map literal",
			p.curToken.Line, p.curToken.Column))
		return ast.MapEntry{}, false
	}

	// Expect '=>'
	if !p.peekTokenIs(token.HASHROCKET) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected '=>' after map key",
			p.curToken.Line, p.curToken.Column))
		return ast.MapEntry{}, false
	}
	p.nextToken() // move to '=>'
	p.nextToken() // move past '=>' to value

	value := p.parseExpression(LOWEST)
	if value == nil {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected value after '=>' in map literal",
			p.curToken.Line, p.curToken.Column))
		return ast.MapEntry{}, false
	}

	return ast.MapEntry{Key: key, Value: value}, true
}

func (p *Parser) parseGroupedExpr() ast.Expression {
	p.nextToken() // consume '('

	expr := p.parseExpression(LOWEST)

	if !p.peekTokenIs(token.RPAREN) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected ')'",
			p.peekToken.Line, p.peekToken.Column))
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

func (p *Parser) parseRangeLit(start ast.Expression) ast.Expression {
	// curToken is '..' or '...'
	exclusive := p.curToken.Type == token.TRIPLEDOT
	precedence := p.curPrecedence()
	p.nextToken()
	end := p.parseExpression(precedence)

	return &ast.RangeLit{Start: start, End: end, Exclusive: exclusive}
}

func (p *Parser) parseSelectorExpr(x ast.Expression) ast.Expression {
	// curToken is '.'
	p.nextToken() // move past '.' to the selector

	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected identifier after '.'",
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	return &ast.SelectorExpr{
		X:   x,
		Sel: p.curToken.Literal,
	}
}

func (p *Parser) parseDeferStmt() *ast.DeferStmt {
	deferLine := p.curToken.Line
	deferCol := p.curToken.Column
	p.nextToken() // consume 'defer'

	// Parse the expression that should be a call
	expr := p.parseExpression(LOWEST)
	if expr == nil {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected expression after defer",
			deferLine, deferCol))
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
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: defer requires a callable expression",
			deferLine, deferCol))
		return nil
	}

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.DeferStmt{Call: call}
}

func (p *Parser) parseInstanceVar() ast.Expression {
	p.nextToken() // consume '@'
	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected identifier after '@'",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	name := p.curToken.Literal
	return &ast.InstanceVar{Name: name}
}

func (p *Parser) parseInstanceVarAssign() *ast.InstanceVarAssign {
	p.nextToken() // consume '@'
	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected identifier after '@'",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	if !p.curTokenIs(token.ASSIGN) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected '=' after instance variable",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	p.nextToken() // consume '='

	value := p.parseExpression(LOWEST)

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
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected '||=' in or-assignment",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	p.nextToken() // consume '||='

	value := p.parseExpression(LOWEST)

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
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected compound assignment operator",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	p.nextToken() // consume operator

	value := p.parseExpression(LOWEST)

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
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected identifier after '@'",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	if !p.curTokenIs(token.ORASSIGN) {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: expected '||=' after instance variable",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	p.nextToken() // consume '||='

	value := p.parseExpression(LOWEST)

	// Check for block after expression
	if p.peekTokenIs(token.DO) || p.peekTokenIs(token.LBRACE) {
		value = p.parseBlockCall(value)
	}

	p.nextToken() // move past expression
	p.skipNewlines()

	return &ast.InstanceVarOrAssign{Name: name, Value: value}
}
