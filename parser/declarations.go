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
		if !p.curTokenIs(token.IDENT) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected alias name after 'as'")
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
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected parameter name")
		return nil
	}
	name := p.curToken.Literal
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
	if !p.curTokenIs(token.IDENT) {
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

	if !p.curTokenIs(token.IDENT) {
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

			if !p.curTokenIs(token.IDENT) {
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
				if !p.curTokenIs(token.IDENT) {
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
		} else if p.curTokenIs(token.IDENT) {
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

	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected class name after 'class'")
		return nil
	}

	cls := &ast.ClassDecl{Name: p.curToken.Literal, Line: line, Doc: doc}
	p.nextToken()

	// Parse optional embedded types: class Service < Logger, Authenticator
	if p.curTokenIs(token.LT) {
		p.nextToken() // consume '<'
		if !p.curTokenIs(token.IDENT) {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected type name after '<'")
			return nil
		}
		cls.Embeds = append(cls.Embeds, p.curToken.Literal)
		p.nextToken()

		// Parse additional embedded types separated by commas
		for p.curTokenIs(token.COMMA) {
			p.nextToken() // consume ','
			if !p.curTokenIs(token.IDENT) {
				p.errorAt(p.curToken.Line, p.curToken.Column, "expected type name after ','")
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
			p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("unexpected token %s in class body, expected 'def', 'pub def', or 'end'", p.curToken.Type))
			p.nextToken()
		}
	}

	if !p.curTokenIs(token.END) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected 'end' to close class")
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
	line := p.curToken.Line
	doc := p.leadingComments(line)
	return p.parseInterfaceDeclWithDoc(doc)
}

func (p *Parser) parseInterfaceDeclWithDoc(doc *ast.CommentGroup) *ast.InterfaceDecl {
	line := p.curToken.Line
	p.nextToken() // consume 'interface'

	if !p.curTokenIs(token.IDENT) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected interface name after 'interface'")
		return nil
	}

	iface := &ast.InterfaceDecl{Name: p.curToken.Literal, Line: line, Doc: doc}
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

	if !p.curTokenIs(token.IDENT) {
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

			if !p.curTokenIs(token.IDENT) {
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
				if !p.curTokenIs(token.IDENT) {
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
		} else if p.curTokenIs(token.IDENT) {
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

			if !p.curTokenIs(token.IDENT) {
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
				if !p.curTokenIs(token.IDENT) {
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
		} else if p.curTokenIs(token.IDENT) {
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
