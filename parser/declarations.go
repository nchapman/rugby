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
func (p *Parser) parseTypedParam(seen map[string]bool) *ast.Param {
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
			fmt.Sprintf("add type annotation: %s : Type", name))
		return nil
	}
	p.nextToken() // consume ':'
	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ':'")
		return nil
	}
	paramType := p.parseTypeName()

	return &ast.Param{Name: name, Type: paramType}
}

// parseTypeName parses a type name (e.g., "Int", "String?")
// The ? suffix is already included in the identifier by the lexer for simple types (Int?).
// For generics (Array[Int]), brackets are tokens.
func (p *Parser) parseTypeName() string {
	typeName := p.curToken.Literal
	p.nextToken() // consume type name

	// Check for generics: Type[T] or Type[K, V]
	if p.curTokenIs(token.LBRACKET) {
		typeName += "["
		p.nextToken() // consume '['

		typeName += p.parseTypeName()

		for p.curTokenIs(token.COMMA) {
			typeName += ", "
			p.nextToken() // consume ','
			typeName += p.parseTypeName()
		}

		if p.curTokenIs(token.RBRACKET) {
			typeName += "]"
			p.nextToken() // consume ']'
		} else {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected ']' after generic types")
			return typeName // Return what we have so far
		}
	}

	// Check for optional suffix '?' (for generic types like Array[Int]?)
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

func (p *Parser) parseFuncDeclWithDoc(doc *ast.CommentGroup) *ast.FuncDecl {
	line := p.curToken.Line
	p.nextToken() // consume 'def'

	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected function name after def")
		return nil
	}

	fn := &ast.FuncDecl{Name: p.curToken.Literal, Line: line, Doc: doc}
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
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected ')' after parameters")
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
		} else if p.curTokenIs(token.IDENT) || p.curTokenIs(token.ANY) {
			// Single return type
			fn.ReturnTypes = append(fn.ReturnTypes, p.parseTypeName())
		} else {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ->")
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
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close function")
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

func (p *Parser) parseClassDeclWithDoc(doc *ast.CommentGroup) *ast.ClassDecl {
	line := p.curToken.Line
	p.nextToken() // consume 'class'

	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.ANY) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected class name after 'class'")
		return nil
	}

	cls := &ast.ClassDecl{Name: p.curToken.Literal, Line: line, Doc: doc}
	p.nextToken()

	// Parse optional embedded types: class Service < Logger, Authenticator
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

	// Parse field declarations and methods until 'end'
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
		case token.PUB:
			pubLine := p.curToken.Line
			methodDoc := p.leadingComments(pubLine)
			p.nextToken() // consume 'pub'
			if p.curTokenIs(token.DEF) {
				if method := p.parseMethodDeclWithDoc(methodDoc); method != nil {
					method.Pub = true
					cls.Methods = append(cls.Methods, method)
				}
			} else {
				p.errorAt(p.curToken.Line, p.curToken.Column, "'pub' in class body must be followed by 'def'")
				p.nextToken()
			}
		case token.DEF:
			if method := p.parseMethodDecl(); method != nil {
				cls.Methods = append(cls.Methods, method)
			}
		default:
			p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("unexpected token %s in class body, expected '@field', 'def', 'pub def', or 'end'", p.curToken.Type))
			p.nextToken()
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close class")
		return nil
	}
	p.nextToken() // consume 'end'

	// Merge explicit field declarations with inferred fields from initialize
	for _, method := range cls.Methods {
		if method.Name == "initialize" {
			inferredFields := extractFields(method)
			// Add inferred fields that weren't explicitly declared
			explicitNames := make(map[string]bool)
			for _, field := range cls.Fields {
				explicitNames[field.Name] = true
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
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close interface")
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

	// Parse optional return type: -> Type or -> (Type1, Type2)
	if p.curTokenIs(token.ARROW) {
		p.nextToken() // consume '->'

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
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ->")
			return nil
		}
	}

	p.skipNewlines()

	return sig
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
			fmt.Sprintf("add type annotation: @%s : Type", name))
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

	// Accept regular identifiers or operator tokens (== for custom equality)
	var methodName string
	switch p.curToken.Type {
	case token.IDENT:
		methodName = p.curToken.Literal
	case token.EQ:
		methodName = "=="
	default:
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected method name after def")
		return nil
	}

	method := &ast.MethodDecl{Name: methodName, Line: line, Doc: doc}
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
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected ')' after parameters")
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
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected type after ->")
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
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close method")
		return nil
	}
	p.nextToken() // consume 'end'

	return method
}
