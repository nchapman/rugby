package lexer

import (
	"testing"

	"rugby/token"
)

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			typ token.TokenType
			lit string
		}
	}{
		{
			name:  "division without spaces",
			input: `x/y`,
			expected: []struct {
				typ token.TokenType
				lit string
			}{
				{token.IDENT, "x"},
				{token.SLASH, "/"},
				{token.IDENT, "y"},
				{token.EOF, ""},
			},
		},
		{
			name:  "not equal without spaces",
			input: `x!=y`,
			expected: []struct {
				typ token.TokenType
				lit string
			}{
				{token.IDENT, "x"},
				{token.NE, "!="},
				{token.IDENT, "y"},
				{token.EOF, ""},
			},
		},
		{
			name:  "method ending in bang assign",
			input: `save!=y`, // interpreted as save != y
			expected: []struct {
				typ token.TokenType
				lit string
			}{
				{token.IDENT, "save"},
				{token.NE, "!="},
				{token.IDENT, "y"},
				{token.EOF, ""},
			},
		},
		{
			name:  "string escaping",
			input: `"hello \"world\""`,
			expected: []struct {
				typ token.TokenType
				lit string
			}{
				{token.STRING, `hello "world"`}, // Lexer should unescape? Or just return raw?
				// Usually lexer returns the string content. If it supports escaping, it should handle it.
				// Based on readString implementation, it stops at first quote.
				{token.EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.typ {
					t.Errorf("token %d: type = %q, want %q", i, tok.Type, exp.typ)
				}
				// For string escaping, we might need to adjust expectation depending on implementation
				// But for division and NE, checking types is enough to fail.
				if tok.Literal != exp.lit {
					t.Errorf("token %d: literal = %q, want %q", i, tok.Literal, exp.lit)
				}
			}
		})
	}
}
