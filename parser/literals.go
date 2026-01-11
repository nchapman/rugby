// Package parser implements the Rugby parser.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/token"
)

// parseIntLiteral parses an integer literal.
func (p *Parser) parseIntLiteral() ast.Expression {
	val, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("invalid integer %s", p.curToken.Literal))
		return nil
	}
	return &ast.IntLit{Value: val}
}

// parseFloatLiteral parses a floating point literal.
func (p *Parser) parseFloatLiteral() ast.Expression {
	val, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("invalid float %s", p.curToken.Literal))
		return nil
	}
	return &ast.FloatLit{Value: val}
}

// parseStringLiteral parses a string literal, handling interpolation.
func (p *Parser) parseStringLiteral() ast.Expression {
	value := p.curToken.Literal

	// Check if string contains interpolation
	if !strings.Contains(value, "#{") {
		return &ast.StringLit{Value: value}
	}

	// Parse interpolated string
	return p.parseInterpolatedString(value)
}

// parseInterpolatedString parses a string containing #{expr} interpolations.
func (p *Parser) parseInterpolatedString(value string) ast.Expression {
	var parts []any
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
			switch value[exprEnd] {
			case '{':
				braceCount++
			case '}':
				braceCount--
			}
			exprEnd++
		}

		if braceCount != 0 {
			p.errorAt(p.curToken.Line, p.curToken.Column, "unterminated string interpolation")
			return &ast.StringLit{Value: value}
		}

		// Extract expression content (excluding the closing })
		exprContent := value[exprStart : exprEnd-1]

		// Parse the expression
		exprLexer := lexer.New(exprContent)
		exprParser := New(exprLexer)
		expr := exprParser.parseExpression(lowest)

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

// parseBoolLiteral parses a boolean literal (true or false).
func (p *Parser) parseBoolLiteral() ast.Expression {
	return &ast.BoolLit{Value: p.curTokenIs(token.TRUE)}
}

// parseNilLiteral parses a nil literal.
func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLit{}
}

// parseSymbolLiteral parses a symbol literal (:name).
func (p *Parser) parseSymbolLiteral() ast.Expression {
	return &ast.SymbolLit{Value: p.curToken.Literal}
}

// parseArrayLiteral parses an array literal [a, b, c].
func (p *Parser) parseArrayLiteral() ast.Expression {
	arr := &ast.ArrayLit{}

	p.nextToken() // consume '['

	// Handle empty array
	if p.curTokenIs(token.RBRACKET) {
		return arr
	}

	// Parse first element
	elem := p.parseExpression(lowest)
	if elem == nil {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected expression in array literal")
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

		elem := p.parseExpression(lowest)
		if elem == nil {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected expression after comma in array literal")
			return nil
		}
		arr.Elements = append(arr.Elements, elem)
	}

	// Move past last element to ']'
	p.nextToken()

	if !p.curTokenIs(token.RBRACKET) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected ']' after array elements")
		return nil
	}

	return arr
}

// parseMapLiteral parses a map literal {a => b, c => d}.
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
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '}' after map entries")
		return nil
	}

	return mapLit
}

// parseMapEntry parses a single key => value entry in a map literal.
func (p *Parser) parseMapEntry() (ast.MapEntry, bool) {
	key := p.parseExpression(lowest)
	if key == nil {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected key in map literal")
		return ast.MapEntry{}, false
	}

	// Expect '=>'
	if !p.peekTokenIs(token.HASHROCKET) {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected '=>' after map key")
		return ast.MapEntry{}, false
	}
	p.nextToken() // move to '=>'
	p.nextToken() // move past '=>' to value

	value := p.parseExpression(lowest)
	if value == nil {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected value after '=>' in map literal")
		return ast.MapEntry{}, false
	}

	return ast.MapEntry{Key: key, Value: value}, true
}
