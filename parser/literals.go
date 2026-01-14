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
	line, col := p.curToken.Line, p.curToken.Column
	val, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		p.errorAt(line, col, fmt.Sprintf("invalid integer %s", p.curToken.Literal))
		return nil
	}
	return &ast.IntLit{Value: val, Line: line, Column: col}
}

// parseFloatLiteral parses a floating point literal.
func (p *Parser) parseFloatLiteral() ast.Expression {
	line, col := p.curToken.Line, p.curToken.Column
	val, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.errorAt(line, col, fmt.Sprintf("invalid float %s", p.curToken.Literal))
		return nil
	}
	return &ast.FloatLit{Value: val, Line: line, Column: col}
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

		// Check for block after expression (e.g., nums.map { |n| n * 2 })
		if expr != nil && (exprParser.peekTokenIs(token.DO) || exprParser.peekTokenIs(token.LBRACE)) {
			expr = exprParser.parseBlockCall(expr)
		}

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
	return &ast.BoolLit{
		Value:  p.curTokenIs(token.TRUE),
		Line:   p.curToken.Line,
		Column: p.curToken.Column,
	}
}

// parseNilLiteral parses a nil literal.
func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLit{Line: p.curToken.Line, Column: p.curToken.Column}
}

// parseSymbolLiteral parses a symbol literal (:name).
func (p *Parser) parseSymbolLiteral() ast.Expression {
	return &ast.SymbolLit{
		Value:  p.curToken.Literal,
		Line:   p.curToken.Line,
		Column: p.curToken.Column,
	}
}

// parseWordArray parses a %w{...} or %W{...} word array literal.
// The lexer stores words separated by \x00 in the token literal.
// If isInterpolated is true, each word may contain #{} interpolation.
func (p *Parser) parseWordArray(isInterpolated bool) ast.Expression {
	arr := &ast.ArrayLit{Line: p.curToken.Line, Column: p.curToken.Column}
	literal := p.curToken.Literal

	// Empty word array
	if literal == "" {
		return arr
	}

	// Split words by null byte separator
	for word := range strings.SplitSeq(literal, "\x00") {
		if word == "" {
			continue
		}

		if isInterpolated && strings.Contains(word, "#{") {
			// Parse as interpolated string
			elem := p.parseInterpolatedString(word)
			arr.Elements = append(arr.Elements, elem)
		} else {
			arr.Elements = append(arr.Elements, &ast.StringLit{Value: word})
		}
	}

	return arr
}

// parseArrayLiteral parses an array literal [a, b, c].
// Supports splat operator: [1, *rest, 3]
func (p *Parser) parseArrayLiteral() ast.Expression {
	arr := &ast.ArrayLit{Line: p.curToken.Line, Column: p.curToken.Column}

	p.nextToken() // consume '['

	// Skip newlines after opening bracket (allows multiline arrays)
	p.skipNewlines()

	// Handle empty array
	if p.curTokenIs(token.RBRACKET) {
		return arr
	}

	// Parse first element (may be splat)
	elem := p.parseArrayElement()
	if elem == nil {
		p.errorAt(p.curToken.Line, p.curToken.Column, "expected expression in array literal")
		return nil
	}
	arr.Elements = append(arr.Elements, elem)

	// Parse remaining elements - handle both commas and newlines between elements
	// After parsing an element, current is at the element, peek is what comes next
	for {
		// Skip any newlines after the current element
		for p.peekTokenIs(token.NEWLINE) {
			p.nextToken()
		}

		// Check for end of array
		if p.peekTokenIs(token.RBRACKET) {
			p.nextToken() // move to ']'
			return arr
		}

		// Expect comma between elements
		if !p.peekTokenIs(token.COMMA) {
			break // exit loop, will hit error below
		}

		p.nextToken() // move to ','
		p.nextToken() // move past ',' to next element or newline

		// Skip newlines after comma
		p.skipNewlines()

		// Allow trailing comma
		if p.curTokenIs(token.RBRACKET) {
			return arr
		}

		elem := p.parseArrayElement()
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

// parseArrayElement parses a single array element, which may be a splat.
func (p *Parser) parseArrayElement() ast.Expression {
	// Check for splat: *expr
	if p.curTokenIs(token.STAR) {
		p.nextToken() // consume '*'
		expr := p.parseExpression(lowest)
		if expr == nil {
			return nil
		}
		return &ast.SplatExpr{Expr: expr}
	}

	return p.parseExpression(lowest)
}

// parseMapLiteral parses a map literal {a => b, c => d}.
// Supports double splat: {**defaults, key: value}
func (p *Parser) parseMapLiteral() ast.Expression {
	mapLit := &ast.MapLit{}

	p.nextToken() // consume '{'

	// Skip newlines after opening brace (allows multiline maps)
	p.skipNewlines()

	// Handle empty map
	if p.curTokenIs(token.RBRACE) {
		return mapLit
	}

	// Parse first entry (may be double splat)
	entry, ok := p.parseMapEntry()
	if !ok {
		return nil
	}
	mapLit.Entries = append(mapLit.Entries, entry)

	// Parse remaining entries - handle both commas and newlines between entries
	// After parsing an entry, current is at the value, peek is what comes next
	for {
		// Skip any newlines after the current entry
		for p.peekTokenIs(token.NEWLINE) {
			p.nextToken()
		}

		// Check for end of map
		if p.peekTokenIs(token.RBRACE) {
			p.nextToken() // move to '}'
			return mapLit
		}

		// Expect comma between entries
		if !p.peekTokenIs(token.COMMA) {
			break // exit loop, will hit error below
		}

		p.nextToken() // move to ','
		p.nextToken() // move past ',' to next key or newline

		// Skip newlines after comma
		p.skipNewlines()

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

// parseMapEntry parses a single key => value or key: value entry in a map literal.
// Supports five forms:
//   - Hash rocket: "key" => value
//   - Symbol key shorthand: key: value (key becomes string)
//   - String key shorthand: "key": value (JSON-style)
//   - Implicit value shorthand: key: (key becomes both string key and variable value)
//   - Double splat: **expr (spreads a map into the literal)
func (p *Parser) parseMapEntry() (ast.MapEntry, bool) {
	// Check for double splat: **expr
	if p.curTokenIs(token.DOUBLESTAR) {
		p.nextToken() // consume '**'
		expr := p.parseExpression(lowest)
		if expr == nil {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected expression after '**'")
			return ast.MapEntry{}, false
		}
		return ast.MapEntry{Splat: expr}, true
	}

	// Check for symbol key shorthand: identifier followed by colon
	// e.g., {name: "Alice", age: 30} or {name:, age:} (implicit value)
	if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.COLON) {
		keyName := p.curToken.Literal
		keyLine := p.curToken.Line
		keyCol := p.curToken.Column
		key := &ast.StringLit{Value: keyName}
		p.nextToken() // move to ':'

		// Check for implicit value shorthand: {x:, y:} or {x:}
		// If peek is comma, rbrace, or newline (for multiline), use identifier as value
		if p.peekTokenIs(token.COMMA) || p.peekTokenIs(token.RBRACE) || p.peekTokenIs(token.NEWLINE) {
			value := &ast.Ident{Name: keyName, Line: keyLine, Column: keyCol}
			return ast.MapEntry{Key: key, Value: value}, true
		}

		p.nextToken() // move past ':' to value

		// Parse the value expression
		value := p.parseExpression(lowest)
		if value == nil {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected value after ':' in map literal")
			return ast.MapEntry{}, false
		}
		return ast.MapEntry{Key: key, Value: value}, true
	}

	// Check for string key with colon (JSON-style): "key": value
	if p.curTokenIs(token.STRING) && p.peekTokenIs(token.COLON) {
		key := &ast.StringLit{Value: p.curToken.Literal}
		p.nextToken() // move to ':'
		p.nextToken() // move past ':' to value

		// Parse the value expression
		value := p.parseExpression(lowest)
		if value == nil {
			p.errorAt(p.curToken.Line, p.curToken.Column, "expected value after ':' in map literal")
			return ast.MapEntry{}, false
		}
		return ast.MapEntry{Key: key, Value: value}, true
	}

	// Standard hash rocket form: key => value
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
