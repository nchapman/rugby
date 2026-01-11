// Package parser implements the Rugby parser.
package parser

import (
	"fmt"
	"strings"

	"github.com/nchapman/rugby/token"
)

// ParseError provides structured error information for parser errors.
type ParseError struct {
	Line    int         // Line number (1-based)
	Column  int         // Column number (1-based)
	Message string      // Error message
	Token   token.Token // The offending token
	Hint    string      // Optional suggestion for fixing the error
}

// Error implements the error interface.
func (e ParseError) Error() string {
	loc := fmt.Sprintf("%d:%d", e.Line, e.Column)
	if e.Hint != "" {
		return fmt.Sprintf("%s: %s (hint: %s)", loc, e.Message, e.Hint)
	}
	return fmt.Sprintf("%s: %s", loc, e.Message)
}

// String returns a formatted string representation.
func (e ParseError) String() string {
	return e.Error()
}

// error adds a parse error at the current token position.
func (p *Parser) error(msg string) {
	p.errors = append(p.errors, ParseError{
		Line:    p.curToken.Line,
		Column:  p.curToken.Column,
		Message: msg,
		Token:   p.curToken,
	})
}

// errorAt adds a parse error at a specific position.
func (p *Parser) errorAt(line, col int, msg string) {
	p.errors = append(p.errors, ParseError{
		Line:    line,
		Column:  col,
		Message: msg,
	})
}

// errorWithHint adds a parse error with a suggestion for fixing it.
func (p *Parser) errorWithHint(msg, hint string) {
	p.errors = append(p.errors, ParseError{
		Line:    p.curToken.Line,
		Column:  p.curToken.Column,
		Message: msg,
		Token:   p.curToken,
		Hint:    hint,
	})
}

// expectedError reports that a specific token was expected but not found.
//
//nolint:unused // Will be used when implementing Phase 17 strictness features
func (p *Parser) expectedError(expected string) {
	p.errors = append(p.errors, ParseError{
		Line:    p.curToken.Line,
		Column:  p.curToken.Column,
		Message: fmt.Sprintf("expected %s, got %s", expected, p.curToken.Type),
		Token:   p.curToken,
	})
}

// unexpectedError reports an unexpected token in the current context.
//
//nolint:unused // Will be used when implementing Phase 17 strictness features
func (p *Parser) unexpectedError() {
	p.errors = append(p.errors, ParseError{
		Line:    p.curToken.Line,
		Column:  p.curToken.Column,
		Message: fmt.Sprintf("unexpected %s", p.curToken.Type),
		Token:   p.curToken,
	})
}

// FormatErrors formats a slice of ParseErrors for display.
func FormatErrors(errs []ParseError) string {
	if len(errs) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, e := range errs {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(e.Error())
	}
	return sb.String()
}
