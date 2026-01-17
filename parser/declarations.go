// Package parser implements the Rugby parser.
package parser

import (
	"fmt"
	"strings"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/token"
)

func (p *Parser) parseImport() *ast.ImportDecl {
	line := p.curToken.Line
	doc := p.leadingComments(line)
	p.nextToken() // consume 'import'

	var path strings.Builder
	expectSeparator := false // start by expecting a component

	for !p.curTokenIs(token.EOF) && !p.curTokenIs(token.NEWLINE) {
		// Check for alias 'as'
		if p.curTokenIs(token.AS) && expectSeparator {
			break
		}

		isSeparator := p.curTokenIs(token.SLASH) || p.curTokenIs(token.DOT)
		isConnector := p.curTokenIs(token.MINUS)

		if isSeparator {
			path.WriteString(p.curToken.Literal)
			expectSeparator = false
			p.nextToken()
			continue
		}

		if isConnector {
			path.WriteString(p.curToken.Literal)
			expectSeparator = false
			p.nextToken()
			continue
		}

		// If we expect a separator but got a component (and it's not 'as'), it's an error
		// e.g., "import a b"
		if expectSeparator {
			p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("unexpected token %s in import path", p.curToken.Literal))
			return nil
		}

		// Consume component (IDENT, INT, KEYWORDS, etc.)
		path.WriteString(p.curToken.Literal)
		expectSeparator = true
		p.nextToken()
	}

	if path.Len() == 0 {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected package path after import")
		return nil
	}

	imp := &ast.ImportDecl{Path: path.String(), Line: line, Doc: doc}

	// Check for optional alias: import foo/bar as baz
	if p.curTokenIs(token.AS) {
		p.nextToken() // consume 'as'
		if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected alias name after 'as'")
			return nil
		}
		imp.Alias = p.curToken.Literal
		p.nextToken()
	}

	return imp
}

// parseTypedParam parses a parameter with optional type annotation: name or name : Type
// Also supports parameter promotion: @name : Type
// Also supports variadic parameters: *name : Type
// Also supports default values: name : Type = value
// Also supports destructuring: {name:, age:} : Type or {name: n, age: a} : Type
func (p *Parser) parseTypedParam(seen map[string]bool) *ast.Param {
	// Check for destructuring pattern: {name:, age:} : Type
	if p.curTokenIs(token.LBRACE) {
		return p.parseDestructuringParam(seen)
	}

	// Check for variadic parameter (*name : Type)
	isVariadic := false
	if p.curTokenIs(token.STAR) {
		isVariadic = true
		p.nextToken() // consume '*'
	}

	// Check for parameter promotion (@name : Type)
	isPromoted := false
	if p.curTokenIs(token.AT) {
		isPromoted = true
		p.nextToken() // consume '@'
	}

	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected parameter name")
		return nil
	}

	name := p.curToken.Literal
	if isPromoted {
		// Prefix with @ to mark as promoted parameter. Codegen strips this
		// to generate both the parameter name and the auto-assignment.
		name = "@" + name
	}

	if seen[name] {
		p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("duplicate parameter name %q", name))
		return nil
	}
	seen[name] = true
	p.nextToken() // consume identifier

	// Require explicit type annotation (Phase 17 strictness)
	if !p.curTokenIs(token.COLON) {
		p.errorWithHint("parameter type required",
			fmt.Sprintf("add type annotation: %s: Type", name))
		return nil
	}
	p.nextToken() // consume ':'
	var paramType string
	if p.curTokenIs(token.LPAREN) {
		// Function type: (Int) -> Int
		paramType = p.parseFunctionType()
	} else if p.curTokenIs(token.IDENT) || p.curTokenIs(token.ANY) {
		paramType = p.parseTypeName()
	} else {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ':'")
		return nil
	}

	// Check for default value: = expression
	var defaultValue ast.Expression
	if p.curTokenIs(token.ASSIGN) {
		p.nextToken() // consume '='
		defaultValue = p.parseExpression(lowest)
		if defaultValue == nil {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected default value after '='")
			return nil
		}
		p.nextToken() // move past expression
	}

	return &ast.Param{
		Name:         name,
		Type:         paramType,
		DefaultValue: defaultValue,
		Variadic:     isVariadic,
	}
}

// parseDestructuringParam parses a destructuring parameter: {name:, age:} : Type
// or with renaming: {name: n, age: a} : Type
func (p *Parser) parseDestructuringParam(seen map[string]bool) *ast.Param {
	p.nextToken() // consume '{'

	var pairs []ast.MapDestructurePair

	// Parse key-variable pairs
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if !p.curTokenIs(token.IDENT) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected key identifier in destructuring pattern")
			return nil
		}
		key := p.curToken.Literal
		p.nextToken() // consume key

		if !p.curTokenIs(token.COLON) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected ':' after key in destructuring pattern")
			return nil
		}
		p.nextToken() // consume ':'

		// Check for rename pattern (key: var) vs shorthand (key:,)
		variable := key // default: shorthand, variable name = key
		if p.curTokenIs(token.IDENT) {
			variable = p.curToken.Literal
			p.nextToken() // consume variable
		}

		// Check for duplicate variable names
		if seen[variable] {
			p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("duplicate parameter name %q", variable))
			return nil
		}
		seen[variable] = true

		pairs = append(pairs, ast.MapDestructurePair{Key: key, Variable: variable})

		// Check for comma or end
		if p.curTokenIs(token.COMMA) {
			p.nextToken() // consume ','
		}
	}

	// Validate that we have at least one pair
	if len(pairs) == 0 {
		p.errorAt(p.curToken.Line, p.curToken.Column, "destructuring pattern requires at least one key")
		return nil
	}

	if !p.curTokenIs(token.RBRACE) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '}' after destructuring pattern")
		return nil
	}
	p.nextToken() // consume '}'

	// Require type annotation for destructuring params
	if !p.curTokenIs(token.COLON) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "type annotation required for destructuring parameter")
		return nil
	}
	p.nextToken() // consume ':'

	var paramType string
	if p.curTokenIs(token.IDENT) || p.curTokenIs(token.ANY) {
		paramType = p.parseTypeName()
	} else {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ':'")
		return nil
	}

	// Generate a synthetic name for the parameter (used in codegen)
	// This will be replaced with something like _arg0 in codegen
	name := "_destructure"

	return &ast.Param{
		Name:             name,
		Type:             paramType,
		DestructurePairs: pairs,
	}
}

// parseFunctionType parses a function type like (): Int or (Int, String): Bool.
// Returns the full function type string including params and return type.
// Assumes the parser is positioned at '(' when called.
func (p *Parser) parseFunctionType() string {
	result := "("
	p.nextToken() // consume '('

	// Parse parameter types
	if !p.curTokenIs(token.RPAREN) {
		result += p.parseTypeName()
		for p.curTokenIs(token.COMMA) {
			result += ", "
			p.nextToken() // consume ','
			result += p.parseTypeName()
		}
	}

	if !p.curTokenIs(token.RPAREN) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected ')' in function type")
		return result
	}
	result += ")"
	p.nextToken() // consume ')'

	if !p.curTokenIs(token.COLON) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected ':' after function params")
		return result
	}
	result += ": "
	p.nextToken() // consume ':'

	// Parse return type (which could itself be a function type)
	if p.curTokenIs(token.LPAREN) {
		result += p.parseFunctionType()
	} else if p.curTokenIs(token.IDENT) || p.curTokenIs(token.ANY) {
		result += p.parseTypeName()
	} else {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected return type")
	}

	return result
}

// parseTypeName parses a type name (e.g., "Int", "String?")
// The ? suffix is already included in the identifier by the lexer for simple types (Int?).
// For generics (Array<Int>), angle brackets are tokens.
func (p *Parser) parseTypeName() string {
	typeName := p.curToken.Literal
	p.nextToken() // consume type name

	// Check for generics: Type<T> or Type<K, V>
	if p.curTokenIs(token.LT) {
		typeName += "<"
		p.nextToken() // consume '<'

		typeName += p.parseTypeName()

		for p.curTokenIs(token.COMMA) {
			typeName += ", "
			p.nextToken() // consume ','
			typeName += p.parseTypeName()
		}

		if p.curTokenIs(token.GT) {
			typeName += ">"
			p.nextToken() // consume '>'
		} else if p.curTokenIs(token.SHOVELRIGHT) {
			// Handle >> in nested generics like Array<Map<String, Int>>
			// Treat the >> as two separate > tokens. Safe to mutate p.curToken
			// directly because we only consume one > here; the synthetic > will
			// be consumed by the outer parseTypeName call (recursive).
			typeName += ">"
			p.curToken = token.Token{
				Type:    token.GT,
				Literal: ">",
				Line:    p.curToken.Line,
				Column:  p.curToken.Column + 1,
			}
		} else {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected '>' after generic types")
			return typeName // Return what we have so far
		}
	}

	// Check for optional suffix '?' (for generic types like Array<Int>?)
	if p.curTokenIs(token.QUESTION) {
		typeName += "?"
		p.nextToken()
	}

	return typeName
}

func (p *Parser) parseFuncDecl() *ast.FuncDecl {
	line := p.curToken.Line
	doc := p.leadingComments(line)
	return p.parseFuncDeclWithDoc(doc)
}

// parseTypeParams parses generic type parameters: <T>, <T, U>, <T : Constraint>
// Returns nil if no type parameters are present (no '<' token)
func (p *Parser) parseTypeParams() []*ast.TypeParam {
	if !p.curTokenIs(token.LT) {
		return nil
	}
	p.nextToken() // consume '<'

	var typeParams []*ast.TypeParam

	// Parse first type parameter
	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected type parameter name")
		return nil
	}

	for {
		tp := &ast.TypeParam{Name: p.curToken.Literal}
		p.nextToken()

		// Check for constraint: T : Constraint
		if p.curTokenIs(token.COLON) {
			p.nextToken() // consume ':'
			if !p.curTokenIs(token.IDENT) {
				p.errorAt(p.curToken.Line, p.curToken.Column, "expected constraint name after ':'")
				return nil
			}
			tp.Constraint = p.curToken.Literal
			p.nextToken()

			// Handle multiple constraints: T : A & B
			for p.curTokenIs(token.AMP) {
				p.nextToken() // consume '&'
				if !p.curTokenIs(token.IDENT) {
					p.errorAt(p.curToken.Line, p.curToken.Column, "expected constraint name after '&'")
					return nil
				}
				tp.Constraint += " & " + p.curToken.Literal
				p.nextToken()
			}
		}

		typeParams = append(typeParams, tp)

		// Check for more type parameters
		if p.curTokenIs(token.COMMA) {
			p.nextToken() // consume ','
			if !p.curTokenIs(token.IDENT) {
				p.errorAt(p.curToken.Line, p.curToken.Column, "expected type parameter name after ','")
				return nil
			}
			continue
		}

		break
	}

	if !p.curTokenIs(token.GT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '>' after type parameters")
		return nil
	}
	p.nextToken() // consume '>'

	return typeParams
}

func (p *Parser) parseFuncDeclWithDoc(doc *ast.CommentGroup) *ast.FuncDecl {
	line := p.curToken.Line
	p.nextToken() // consume 'def'

	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected function name after def")
		return nil
	}

	fn := &ast.FuncDecl{Name: p.curToken.Literal, Line: line, Doc: doc}
	p.nextToken()

	// Parse optional type parameters: def name<T, U>
	fn.TypeParams = p.parseTypeParams()

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
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected ')' after parameters")
			return nil
		}
		p.nextToken() // consume ')'
	}

	// Parse optional return type: : Type, : (Type1, Type2), or : (): T (function type)
	if p.curTokenIs(token.COLON) {
		p.nextToken() // consume ':'

		if p.curTokenIs(token.LPAREN) {
			// Could be tuple return (Type1, Type2) or function type (): T
			// Use lookahead to check if it's a function type
			if p.isFunctionTypeAhead() {
				fn.ReturnTypes = append(fn.ReturnTypes, p.parseFunctionType())
			} else {
				// Multiple return types: (Type1, Type2, ...)
				p.nextToken() // consume '('

				if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
					p.errorAt(p.curToken.Line, p.curToken.Column, "expected type in return type list")
					return nil
				}
				fn.ReturnTypes = append(fn.ReturnTypes, p.parseTypeName())

				for p.curTokenIs(token.COMMA) {
					p.nextToken() // consume ','
					// Allow trailing comma
					if p.curTokenIs(token.RPAREN) {
						break
					}
					if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
						p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after comma")
						return nil
					}
					fn.ReturnTypes = append(fn.ReturnTypes, p.parseTypeName())
				}

				if !p.curTokenIs(token.RPAREN) {
					p.errorAt(p.curToken.Line, p.curToken.Column, "expected ')' after return types")
					return nil
				}
				p.nextToken() // consume ')'
			}
		} else if p.curTokenIs(token.IDENT) || p.curTokenIs(token.ANY) {
			// Single return type
			fn.ReturnTypes = append(fn.ReturnTypes, p.parseTypeName())
		} else {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ':'")
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
		} else {
			p.nextToken() // error recovery: always advance to prevent infinite loop
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAtWithHint(p.curToken.Line, p.curToken.Column,
			"expected 'end' to close function",
			"every 'def' needs a matching 'end'")
		return nil
	}
	p.nextToken() // consume 'end'

	return fn
}

func (p *Parser) parseClassDecl() *ast.ClassDecl {
	line := p.curToken.Line
	doc := p.leadingComments(line)
	return p.parseClassDeclWithDoc(doc)
}

// isGenericTypeParamList uses lookahead to check if current '<' starts a generic type parameter list.
// Returns true if we find a matching '>' (generics), false otherwise (inheritance).
func (p *Parser) isGenericTypeParamList() bool {
	if !p.curTokenIs(token.LT) {
		return false
	}

	// Save state for lookahead
	lexerState := p.l.SaveState()
	savedCurToken := p.curToken
	savedPeekToken := p.peekToken

	p.nextToken() // skip '<'

	// Look for '>' while allowing valid type param syntax
	depth := 1
	result := false
	for depth > 0 && !p.curTokenIs(token.EOF) && !p.curTokenIs(token.NEWLINE) {
		switch p.curToken.Type {
		case token.LT:
			depth++
		case token.GT:
			depth--
			if depth == 0 {
				result = true // Found matching '>'
				goto restore
			}
		case token.SHOVELRIGHT:
			// >> counts as two closing brackets for nested generics
			depth -= 2
			if depth <= 0 {
				result = true
				goto restore
			}
		case token.IDENT, token.COMMA, token.COLON, token.AMP:
			// Valid in type param list
		default:
			goto restore // Invalid token for type params
		}
		p.nextToken()
	}

restore:
	// Restore parser state
	p.l.RestoreState(lexerState)
	p.curToken = savedCurToken
	p.peekToken = savedPeekToken
	return result
}

func (p *Parser) parseClassDeclWithDoc(doc *ast.CommentGroup) *ast.ClassDecl {
	line := p.curToken.Line
	p.nextToken() // consume 'class'

	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected class name after 'class'")
		return nil
	}

	cls := &ast.ClassDecl{Name: p.curToken.Literal, Line: line, Doc: doc}
	p.nextToken()

	// Parse optional type parameters: class Box<T>
	// Use lookahead to distinguish from inheritance: class Child < Parent
	if p.isGenericTypeParamList() {
		cls.TypeParams = p.parseTypeParams()
	}

	// Parse optional embedded types (inheritance): class Service < Logger, Authenticator
	if p.curTokenIs(token.LT) {
		p.nextToken() // consume '<'
		if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected type name after '<'")
			return nil
		}
		cls.Embeds = append(cls.Embeds, p.curToken.Literal)
		p.nextToken()

		// Parse additional embedded types separated by commas
		for p.curTokenIs(token.COMMA) {
			p.nextToken() // consume ','
			if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
				p.errorAt(p.curToken.Line, p.curToken.Column, "expected type name after ','")
				return nil
			}
			cls.Embeds = append(cls.Embeds, p.curToken.Literal)
			p.nextToken()
		}
	}

	// Parse optional implements: class User implements Speaker, Serializable
	if p.curTokenIs(token.IMPLEMENTS) {
		p.nextToken() // consume 'implements'
		for {
			if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
				p.errorAt(p.curToken.Line, p.curToken.Column, "expected interface name after 'implements'")
				break
			}
			cls.Implements = append(cls.Implements, p.curToken.Literal)
			p.nextToken()
			if !p.curTokenIs(token.COMMA) {
				break
			}
			p.nextToken() // consume comma
		}
	}

	p.skipNewlines()

	// Parse field declarations, accessors, includes, and methods until 'end'
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}

		switch p.curToken.Type {
		case token.AT:
			// Field declaration: @name : Type
			if field := p.parseFieldDecl(); field != nil {
				cls.Fields = append(cls.Fields, field)
			}
		case token.ATAT:
			// Class variable declaration: @@name = value
			if classVar := p.parseClassVarDecl(); classVar != nil {
				cls.ClassVars = append(cls.ClassVars, classVar)
			}
		case token.GETTER, token.SETTER, token.PROPERTY:
			// Accessor declaration: getter/setter/property name : Type
			if accessor := p.parseAccessorDecl(); accessor != nil {
				cls.Accessors = append(cls.Accessors, accessor)
			}
		case token.INCLUDE:
			// Include module: include ModuleName
			p.nextToken() // consume 'include'
			if !p.curTokenIs(token.IDENT) {
				p.errorAt(p.curToken.Line, p.curToken.Column, "expected module name after 'include'")
				p.nextToken()
			} else {
				cls.Includes = append(cls.Includes, p.curToken.Literal)
				p.nextToken()
			}
		case token.PUB:
			pubLine := p.curToken.Line
			methodDoc := p.leadingComments(pubLine)
			p.nextToken() // consume 'pub'
			if p.curTokenIs(token.DEF) {
				if method := p.parseMethodDeclWithDoc(methodDoc); method != nil {
					method.Pub = true
					cls.Methods = append(cls.Methods, method)
				}
			} else if p.curTokenIs(token.GETTER) || p.curTokenIs(token.SETTER) || p.curTokenIs(token.PROPERTY) {
				if accessor := p.parseAccessorDecl(); accessor != nil {
					accessor.Pub = true
					cls.Accessors = append(cls.Accessors, accessor)
				}
			} else {
				p.errorAt(p.curToken.Line, p.curToken.Column, "'pub' in class body must be followed by 'def', 'getter', 'setter', or 'property'")
				p.nextToken()
			}
		case token.PRIVATE:
			privateLine := p.curToken.Line
			methodDoc := p.leadingComments(privateLine)
			p.nextToken() // consume 'private'
			if p.curTokenIs(token.DEF) {
				if method := p.parseMethodDeclWithDoc(methodDoc); method != nil {
					method.Private = true
					cls.Methods = append(cls.Methods, method)
				}
			} else {
				p.errorAt(p.curToken.Line, p.curToken.Column, "'private' in class body must be followed by 'def'")
				p.nextToken()
			}
		case token.DEF:
			if method := p.parseMethodDecl(); method != nil {
				cls.Methods = append(cls.Methods, method)
			}
		default:
			p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("unexpected token %s in class body, expected '@field', 'def', 'pub def', 'private def', 'getter', 'setter', 'property', 'include', or 'end'", p.curToken.Type))
			p.nextToken()
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAtWithHint(p.curToken.Line, p.curToken.Column,
			"expected 'end' to close class",
			"every 'class' needs a matching 'end'")
		return nil
	}
	p.nextToken() // consume 'end'

	// Merge explicit field declarations with inferred fields from initialize
	for _, method := range cls.Methods {
		if method.Name == "initialize" {
			inferredFields := extractFields(method)
			// Add inferred fields that weren't explicitly declared or have accessors
			explicitNames := make(map[string]bool)
			for _, field := range cls.Fields {
				explicitNames[field.Name] = true
			}
			for _, acc := range cls.Accessors {
				explicitNames[acc.Name] = true
			}
			for _, inferred := range inferredFields {
				if !explicitNames[inferred.Name] {
					cls.Fields = append(cls.Fields, inferred)
				}
			}
			break
		}
	}

	// Validate: methods other than initialize cannot introduce new instance variables
	knownFields := make(map[string]bool)
	for _, field := range cls.Fields {
		knownFields[field.Name] = true
	}
	// Accessor fields are also known fields
	for _, acc := range cls.Accessors {
		knownFields[acc.Name] = true
	}
	for _, method := range cls.Methods {
		if method.Name != "initialize" {
			p.validateNoNewFields(method.Body, knownFields)
		}
	}

	return cls
}

func (p *Parser) parseInterfaceDecl() *ast.InterfaceDecl {
	line := p.curToken.Line
	doc := p.leadingComments(line)
	return p.parseInterfaceDeclWithDoc(doc)
}

func (p *Parser) parseInterfaceDeclWithDoc(doc *ast.CommentGroup) *ast.InterfaceDecl {
	line := p.curToken.Line
	p.nextToken() // consume 'interface'

	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected interface name after 'interface'")
		return nil
	}

	iface := &ast.InterfaceDecl{Name: p.curToken.Literal, Line: line, Doc: doc}
	p.nextToken()

	// Parse optional type parameters: interface Container<T>
	// Use lookahead to distinguish from parent interfaces: interface IO < Reader
	if p.isGenericTypeParamList() {
		iface.TypeParams = p.parseTypeParams()
	}

	// Parse parent interfaces if present (interface IO < Reader, Writer)
	if p.curTokenIs(token.LT) {
		p.nextToken() // consume '<'
		for {
			if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
				p.errorAt(p.curToken.Line, p.curToken.Column, "expected interface name after '<'")
				break
			}
			iface.Parents = append(iface.Parents, p.curToken.Literal)
			p.nextToken()
			if !p.curTokenIs(token.COMMA) {
				break
			}
			p.nextToken() // consume comma
		}
	}

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
			p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("unexpected token %s in interface body, expected 'def' or 'end'", p.curToken.Type))
			p.nextToken()
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAtWithHint(p.curToken.Line, p.curToken.Column,
			"expected 'end' to close interface",
			"every 'interface' needs a matching 'end'")
		return nil
	}
	p.nextToken() // consume 'end'

	return iface
}

// parseMethodSig parses a method signature (no body) for interfaces
func (p *Parser) parseMethodSig() *ast.MethodSig {
	p.nextToken() // consume 'def'

	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected method name after def")
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
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected ')' after parameters")
			return nil
		}
		p.nextToken() // consume ')'
	}

	// Parse optional return type: : Type or : (Type1, Type2)
	if p.curTokenIs(token.COLON) {
		p.nextToken() // consume ':'

		if p.curTokenIs(token.LPAREN) {
			// Multiple return types: (Type1, Type2, ...)
			p.nextToken() // consume '('

			if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
				p.errorAt(p.curToken.Line, p.curToken.Column, "expected type in return type list")
				return nil
			}
			sig.ReturnTypes = append(sig.ReturnTypes, p.parseTypeName())

			for p.curTokenIs(token.COMMA) {
				p.nextToken() // consume ','
				// Allow trailing comma
				if p.curTokenIs(token.RPAREN) {
					break
				}
				if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
					p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after comma")
					return nil
				}
				sig.ReturnTypes = append(sig.ReturnTypes, p.parseTypeName())
			}

			if !p.curTokenIs(token.RPAREN) {
				p.errorAt(p.curToken.Line, p.curToken.Column, "expected ')' after return types")
				return nil
			}
			p.nextToken() // consume ')'
		} else if p.curTokenIs(token.IDENT) || p.curTokenIs(token.ANY) {
			// Single return type
			sig.ReturnTypes = append(sig.ReturnTypes, p.parseTypeName())
		} else {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ':'")
			return nil
		}
	}

	p.skipNewlines()

	return sig
}

// parseAccessorDecl parses getter/setter/property name : Type
func (p *Parser) parseAccessorDecl() *ast.AccessorDecl {
	kind := p.curToken.Literal // "getter", "setter", or "property"
	line := p.curToken.Line
	p.nextToken() // consume accessor keyword

	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("expected field name after '%s'", kind))
		return nil
	}

	name := p.curToken.Literal
	p.nextToken() // consume field name

	if !p.curTokenIs(token.COLON) {
		p.errorWithHint("accessor type required",
			fmt.Sprintf("add type annotation: %s %s: Type", kind, name))
		return nil
	}
	p.nextToken() // consume ':'

	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ':'")
		return nil
	}

	fieldType := p.parseTypeName()
	return &ast.AccessorDecl{Kind: kind, Name: name, Type: fieldType, Line: line}
}

// parseFieldDecl parses an explicit field declaration: @name : Type
func (p *Parser) parseFieldDecl() *ast.FieldDecl {
	p.nextToken() // consume '@'

	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected field name after '@'")
		return nil
	}

	name := p.curToken.Literal
	p.nextToken() // consume field name

	if !p.curTokenIs(token.COLON) {
		p.errorWithHint("field type required",
			fmt.Sprintf("add type annotation: @%s: Type", name))
		return nil
	}
	p.nextToken() // consume ':'

	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ':'")
		return nil
	}

	fieldType := p.parseTypeName()
	return &ast.FieldDecl{Name: name, Type: fieldType}
}

// parseClassVarDecl parses a class variable declaration: @@name = value
func (p *Parser) parseClassVarDecl() *ast.ClassVarDecl {
	line := p.curToken.Line
	p.nextToken() // consume '@@'

	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected variable name after '@@'")
		return nil
	}

	name := p.curToken.Literal
	p.nextToken() // consume variable name

	if !p.curTokenIs(token.ASSIGN) {
		p.errorWithHint("class variable requires initial value",
			fmt.Sprintf("add initial value: @@%s = value", name))
		return nil
	}
	p.nextToken() // consume '='

	value := p.parseExpression(lowest)
	p.nextToken() // move past the expression
	p.skipNewlines()
	return &ast.ClassVarDecl{Name: name, Value: value, Line: line}
}

// extractFields extracts field declarations from instance variable assignments and
// parameter promotions in the initialize method.
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

	// Handle parameter promotion: def initialize(@name : String)
	// Parameters starting with @ are promoted to instance variables
	for _, param := range method.Params {
		if len(param.Name) > 0 && param.Name[0] == '@' {
			// This is parameter promotion
			fieldName := param.Name[1:] // strip @
			if !seen[fieldName] {
				seen[fieldName] = true
				fields = append(fields, &ast.FieldDecl{Name: fieldName, Type: param.Type})
				// Also add to paramTypes so assignments work
				paramTypes[fieldName] = param.Type
			}
		}
	}

	// Recursively scan initialize body for @field = value assignments
	// This handles fields initialized in nested control flow (if, while, for)
	var scanStatements func(stmts []ast.Statement)
	scanStatements = func(stmts []ast.Statement) {
		for _, stmt := range stmts {
			switch s := stmt.(type) {
			case *ast.InstanceVarAssign:
				if !seen[s.Name] {
					seen[s.Name] = true

					// Infer type from assignment source
					fieldType := ""
					if ident, ok := s.Value.(*ast.Ident); ok {
						// If assigned from a parameter, use its type
						if pt, ok := paramTypes[ident.Name]; ok {
							fieldType = pt
						}
					}

					fields = append(fields, &ast.FieldDecl{Name: s.Name, Type: fieldType})
				}
			case *ast.InstanceVarOrAssign:
				if !seen[s.Name] {
					seen[s.Name] = true
					// Type inference for ||= is harder since value is conditional
					// For now, use empty type (interface{})
					fields = append(fields, &ast.FieldDecl{Name: s.Name, Type: ""})
				}
			case *ast.IfStmt:
				scanStatements(s.Then)
				scanStatements(s.Else)
				for _, elsif := range s.ElseIfs {
					scanStatements(elsif.Body)
				}
			case *ast.WhileStmt:
				scanStatements(s.Body)
			case *ast.ForStmt:
				scanStatements(s.Body)
			}
		}
	}
	scanStatements(method.Body)

	return fields
}

// validateNoNewFields checks that methods don't introduce new instance variables.
// Only initialize (and explicit declarations) can create new fields.
func (p *Parser) validateNoNewFields(stmts []ast.Statement, knownFields map[string]bool) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.InstanceVarAssign:
			if !knownFields[s.Name] {
				p.errorWithHint(
					fmt.Sprintf("cannot introduce new instance variable @%s outside initialize", s.Name),
					fmt.Sprintf("move @%s assignment to initialize method or add explicit field declaration", s.Name))
			}
		case *ast.InstanceVarOrAssign:
			if !knownFields[s.Name] {
				p.errorWithHint(
					fmt.Sprintf("cannot introduce new instance variable @%s outside initialize", s.Name),
					fmt.Sprintf("move @%s assignment to initialize method or add explicit field declaration", s.Name))
			}
		case *ast.IfStmt:
			p.validateNoNewFields(s.Then, knownFields)
			p.validateNoNewFields(s.Else, knownFields)
			for _, elsif := range s.ElseIfs {
				p.validateNoNewFields(elsif.Body, knownFields)
			}
		case *ast.WhileStmt:
			p.validateNoNewFields(s.Body, knownFields)
		case *ast.ForStmt:
			p.validateNoNewFields(s.Body, knownFields)
		}
	}
}

func (p *Parser) parseMethodDecl() *ast.MethodDecl {
	line := p.curToken.Line
	doc := p.leadingComments(line)
	return p.parseMethodDeclWithDoc(doc)
}

func (p *Parser) parseMethodDeclWithDoc(doc *ast.CommentGroup) *ast.MethodDecl {
	line := p.curToken.Line
	p.nextToken() // consume 'def'

	// Check for class method (def self.method_name)
	isClassMethod := false
	if p.curTokenIs(token.SELF) {
		isClassMethod = true
		p.nextToken() // consume 'self'
		if !p.curTokenIs(token.DOT) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected '.' after self in class method")
			return nil
		}
		p.nextToken() // consume '.'
	}

	// Accept regular identifiers or operator tokens (== for custom equality)
	// Also accept test keywords (describe, it, before, after) as method names
	var methodName string
	switch p.curToken.Type {
	case token.IDENT:
		methodName = p.curToken.Literal
		p.nextToken()
		// Check for setter method (name=)
		if p.curTokenIs(token.ASSIGN) {
			methodName += "="
			p.nextToken() // consume '='
		}
	case token.EQ:
		methodName = "=="
		p.nextToken()
	case token.DESCRIBE, token.IT, token.BEFORE, token.AFTER:
		// Allow test keywords as method names
		methodName = p.curToken.Literal
		p.nextToken()
	default:
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected method name after def")
		return nil
	}

	method := &ast.MethodDecl{Name: methodName, Line: line, Doc: doc, IsClassMethod: isClassMethod}

	// Parse optional type parameters: def map<R>(f : (T) -> R) -> Box<R>
	method.TypeParams = p.parseTypeParams()

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
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected ')' after parameters")
			return nil
		}
		p.nextToken() // consume ')'
	}

	// Parse optional return type: : Type or : (Type1, Type2)
	if p.curTokenIs(token.COLON) {
		p.nextToken() // consume ':'

		if p.curTokenIs(token.LPAREN) {
			// Multiple return types: (Type1, Type2, ...)
			p.nextToken() // consume '('

			if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
				p.errorAt(p.curToken.Line, p.curToken.Column, "expected type in return type list")
				return nil
			}
			method.ReturnTypes = append(method.ReturnTypes, p.parseTypeName())

			for p.curTokenIs(token.COMMA) {
				p.nextToken() // consume ','
				// Allow trailing comma
				if p.curTokenIs(token.RPAREN) {
					break
				}
				if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
					p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after comma")
					return nil
				}
				method.ReturnTypes = append(method.ReturnTypes, p.parseTypeName())
			}

			if !p.curTokenIs(token.RPAREN) {
				p.errorAt(p.curToken.Line, p.curToken.Column, "expected ')' after return types")
				return nil
			}
			p.nextToken() // consume ')'
		} else if p.curTokenIs(token.IDENT) || p.curTokenIs(token.ANY) {
			// Single return type
			method.ReturnTypes = append(method.ReturnTypes, p.parseTypeName())
		} else {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ':'")
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
		} else {
			p.nextToken() // error recovery: always advance to prevent infinite loop
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAtWithHint(p.curToken.Line, p.curToken.Column,
			"expected 'end' to close method",
			"every 'def' needs a matching 'end'")
		return nil
	}
	p.nextToken() // consume 'end'

	return method
}

func (p *Parser) parseModuleDecl() *ast.ModuleDecl {
	line := p.curToken.Line
	doc := p.leadingComments(line)
	return p.parseModuleDeclWithDoc(doc)
}

func (p *Parser) parseModuleDeclWithDoc(doc *ast.CommentGroup) *ast.ModuleDecl {
	line := p.curToken.Line
	p.nextToken() // consume 'module'

	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected module name after 'module'")
		return nil
	}

	mod := &ast.ModuleDecl{Name: p.curToken.Literal, Line: line, Doc: doc}
	p.nextToken()
	p.skipNewlines()

	// Parse field declarations, accessors, and methods until 'end'
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}

		switch p.curToken.Type {
		case token.AT:
			// Field declaration: @name : Type
			if field := p.parseFieldDecl(); field != nil {
				mod.Fields = append(mod.Fields, field)
			}
		case token.GETTER, token.SETTER, token.PROPERTY:
			// Accessor declaration: getter/setter/property name : Type
			if accessor := p.parseAccessorDecl(); accessor != nil {
				mod.Accessors = append(mod.Accessors, accessor)
			}
		case token.DEF:
			if method := p.parseMethodDecl(); method != nil {
				mod.Methods = append(mod.Methods, method)
			}
		case token.CLASS:
			// Nested class declaration inside module
			if class := p.parseClassDecl(); class != nil {
				mod.Classes = append(mod.Classes, class)
			}
		default:
			p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("unexpected token %s in module body, expected '@field', 'def', 'class', 'getter', 'setter', 'property', or 'end'", p.curToken.Type))
			p.nextToken()
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAtWithHint(p.curToken.Line, p.curToken.Column,
			"expected 'end' to close module",
			"every 'module' needs a matching 'end'")
		return nil
	}
	p.nextToken() // consume 'end'

	return mod
}

// parseConstDecl parses a constant declaration: const MAX_SIZE = 1024
func (p *Parser) parseConstDecl() *ast.ConstDecl {
	doc := p.leadingComments(p.curToken.Line)
	return p.parseConstDeclWithDoc(doc)
}

// parseConstDeclWithDoc parses a constant declaration with pre-collected doc comment
func (p *Parser) parseConstDeclWithDoc(doc *ast.CommentGroup) *ast.ConstDecl {
	line := p.curToken.Line
	p.nextToken() // consume 'const'

	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected constant name after 'const'")
		return nil
	}

	name := p.curToken.Literal
	p.nextToken() // consume name

	// Parse optional type annotation: const NAME : Type = value
	var constType string
	if p.curTokenIs(token.COLON) {
		p.nextToken() // consume ':'
		if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ':'")
			return nil
		}
		constType = p.parseTypeName()
	}

	if !p.curTokenIs(token.ASSIGN) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '=' after constant name")
		return nil
	}
	p.nextToken() // consume '='

	// Parse the constant value expression
	value := p.parseExpression(lowest)
	if value == nil {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected expression after '='")
		return nil
	}

	// Move past the expression to the next statement
	p.nextToken()
	p.skipNewlines()

	return &ast.ConstDecl{
		Name:  name,
		Type:  constType,
		Value: value,
		Line:  line,
		Doc:   doc,
	}
}

// parseTypeAliasDecl parses a type alias declaration: type UserID = Int64
func (p *Parser) parseTypeAliasDecl() *ast.TypeAliasDecl {
	doc := p.leadingComments(p.curToken.Line)
	return p.parseTypeAliasDeclWithDoc(doc)
}

// parseTypeAliasDeclWithDoc parses a type alias with pre-collected doc comment
func (p *Parser) parseTypeAliasDeclWithDoc(doc *ast.CommentGroup) *ast.TypeAliasDecl {
	line := p.curToken.Line
	p.nextToken() // consume 'type'

	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected type alias name after 'type'")
		return nil
	}

	name := p.curToken.Literal
	p.nextToken() // consume name

	if !p.curTokenIs(token.ASSIGN) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '=' after type alias name")
		return nil
	}
	p.nextToken() // consume '='

	// Parse the underlying type using the full type parser (handles generics like Array<T>)
	// Also handle function types like (Int) -> Int
	var typeName string
	if p.curTokenIs(token.LPAREN) {
		typeName = p.parseFunctionType()
	} else if p.curTokenIs(token.IDENT) || p.curTokenIs(token.ANY) {
		typeName = p.parseTypeName()
	} else {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after '='")
		return nil
	}

	return &ast.TypeAliasDecl{
		Name: name,
		Type: typeName,
		Line: line,
		Doc:  doc,
	}
}

// parseEnumDecl parses an enum declaration: enum Status ... end
func (p *Parser) parseEnumDecl() *ast.EnumDecl {
	doc := p.leadingComments(p.curToken.Line)
	return p.parseEnumDeclWithDoc(doc)
}

// parseEnumDeclWithDoc parses an enum with pre-collected doc comment
func (p *Parser) parseEnumDeclWithDoc(doc *ast.CommentGroup) *ast.EnumDecl {
	line := p.curToken.Line
	p.nextToken() // consume 'enum'

	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected enum name after 'enum'")
		return nil
	}

	name := p.curToken.Literal
	p.nextToken() // consume name

	// Skip newlines before body
	p.skipNewlines()

	// Parse enum values until 'end'
	var values []*ast.EnumValue
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}

		if !p.curTokenIs(token.IDENT) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected enum value name")
			p.nextToken()
			continue
		}

		valueLine := p.curToken.Line
		valueName := p.curToken.Literal
		p.nextToken() // consume value name

		var valueExpr ast.Expression
		// Check for explicit value: Name = 200
		if p.curTokenIs(token.ASSIGN) {
			p.nextToken() // consume '='
			valueExpr = p.parseExpression(lowest)
			// parseExpression leaves cursor on last token of expression
			// Advance past it if we're not already at newline
			if !p.curTokenIs(token.NEWLINE) && !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
				p.nextToken()
			}
		}

		values = append(values, &ast.EnumValue{
			Name:  valueName,
			Value: valueExpr,
			Line:  valueLine,
		})

		// Skip newlines and continue
		p.skipNewlines()
	}

	if !p.curTokenIs(token.END) {
		p.errorAtWithHint(p.curToken.Line, p.curToken.Column,
			"expected 'end' to close enum",
			"every 'enum' needs a matching 'end'")
		return nil
	}
	p.nextToken() // consume 'end'

	return &ast.EnumDecl{
		Name:   name,
		Values: values,
		Line:   line,
		Doc:    doc,
	}
}

// parseStructDecl parses a struct declaration: struct Point ... end
func (p *Parser) parseStructDecl() *ast.StructDecl {
	doc := p.leadingComments(p.curToken.Line)
	return p.parseStructDeclWithDoc(doc)
}

// parseStructDeclWithDoc parses a struct with pre-collected doc comment
func (p *Parser) parseStructDeclWithDoc(doc *ast.CommentGroup) *ast.StructDecl {
	line := p.curToken.Line
	p.nextToken() // consume 'struct'

	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected struct name after 'struct'")
		return nil
	}

	name := p.curToken.Literal
	p.nextToken() // consume name

	// Parse optional type parameters: struct Box<T> ...
	var typeParams []*ast.TypeParam
	if p.curTokenIs(token.LT) {
		typeParams = p.parseTypeParams()
	}

	// Skip newlines before body
	p.skipNewlines()

	// Parse struct body until 'end'
	var fields []*ast.StructField
	var methods []*ast.MethodDecl
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.END) {
			break
		}

		// Check for method definition
		if p.curTokenIs(token.DEF) {
			if method := p.parseMethodDecl(); method != nil {
				methods = append(methods, method)
			}
			continue
		}

		// Parse field: name : Type
		if !p.curTokenIs(token.IDENT) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected field name or 'def'")
			p.nextToken()
			continue
		}

		fieldLine := p.curToken.Line
		fieldName := p.curToken.Literal
		p.nextToken() // consume field name

		if !p.curTokenIs(token.COLON) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected ':' after field name in struct")
			continue
		}
		p.nextToken() // consume ':'

		fieldType := p.parseTypeName()
		if fieldType == "" {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ':'")
			continue
		}

		fields = append(fields, &ast.StructField{
			Name: fieldName,
			Type: fieldType,
			Line: fieldLine,
		})

		p.skipNewlines()
	}

	if !p.curTokenIs(token.END) {
		p.errorAtWithHint(p.curToken.Line, p.curToken.Column,
			"expected 'end' to close struct",
			"every 'struct' needs a matching 'end'")
		return nil
	}
	p.nextToken() // consume 'end'

	return &ast.StructDecl{
		Name:       name,
		TypeParams: typeParams,
		Fields:     fields,
		Methods:    methods,
		Line:       line,
		Doc:        doc,
	}
}
