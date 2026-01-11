// Package token defines the token types for the Rugby lexer.
package token

type TokenType string

const (
	// Special tokens
	EOF     TokenType = "EOF"
	NEWLINE TokenType = "NEWLINE"
	ILLEGAL TokenType = "ILLEGAL"

	// Literals
	IDENT  TokenType = "IDENT"
	STRING TokenType = "STRING"
	INT    TokenType = "INT"
	FLOAT  TokenType = "FLOAT"
	SYMBOL TokenType = "SYMBOL"

	// Operators
	PLUS    TokenType = "+"
	MINUS   TokenType = "-"
	STAR    TokenType = "*"
	SLASH   TokenType = "/"
	PERCENT TokenType = "%"
	BANG    TokenType = "!"

	// Comparison
	EQ TokenType = "=="
	NE TokenType = "!="
	LT TokenType = "<"
	GT TokenType = ">"
	LE TokenType = "<="
	GE TokenType = ">="

	// Assignment
	ASSIGN      TokenType = "="
	ORASSIGN    TokenType = "||="
	PLUSASSIGN  TokenType = "+="
	MINUSASSIGN TokenType = "-="
	STARASSIGN  TokenType = "*="
	SLASHASSIGN TokenType = "/="

	// Type modifiers
	QUESTION TokenType = "?"

	// Delimiters
	LPAREN     TokenType = "("
	RPAREN     TokenType = ")"
	LBRACKET   TokenType = "["
	RBRACKET   TokenType = "]"
	LBRACE     TokenType = "{"
	RBRACE     TokenType = "}"
	COMMA      TokenType = ","
	ARROW      TokenType = "->"
	HASHROCKET TokenType = "=>"
	DOT        TokenType = "."
	DOTDOT     TokenType = ".."
	TRIPLEDOT  TokenType = "..."
	PIPE       TokenType = "|"
	AT         TokenType = "@"
	COLON      TokenType = ":"

	// Keywords
	IMPORT    TokenType = "IMPORT"
	DEF       TokenType = "DEF"
	END       TokenType = "END"
	IF        TokenType = "IF"
	ELSIF     TokenType = "ELSIF"
	ELSE      TokenType = "ELSE"
	UNLESS    TokenType = "UNLESS"
	CASE      TokenType = "CASE"
	WHEN      TokenType = "WHEN"
	WHILE     TokenType = "WHILE"
	FOR       TokenType = "FOR"
	IN        TokenType = "IN"
	BREAK     TokenType = "BREAK"
	NEXT      TokenType = "NEXT"
	RETURN    TokenType = "RETURN"
	PANIC     TokenType = "PANIC"
	RESCUE    TokenType = "RESCUE"
	TRUE      TokenType = "TRUE"
	FALSE     TokenType = "FALSE"
	NIL       TokenType = "NIL"
	AND       TokenType = "AND"
	OR        TokenType = "OR"
	NOT       TokenType = "NOT"
	AS        TokenType = "AS"
	DEFER     TokenType = "DEFER"
	DO         TokenType = "DO"
	CLASS      TokenType = "CLASS"
	SELF       TokenType = "SELF"
	INTERFACE  TokenType = "INTERFACE"
	IMPLEMENTS TokenType = "IMPLEMENTS"
	ANY        TokenType = "ANY"
	PUB        TokenType = "PUB"

	// Testing keywords
	DESCRIBE TokenType = "DESCRIBE"
	IT       TokenType = "IT"
	TEST     TokenType = "TEST"
	TABLE    TokenType = "TABLE"
	BEFORE   TokenType = "BEFORE"
	AFTER    TokenType = "AFTER"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

var keywords = map[string]TokenType{
	"import":    IMPORT,
	"def":       DEF,
	"end":       END,
	"if":        IF,
	"elsif":     ELSIF,
	"else":      ELSE,
	"unless":    UNLESS,
	"case":      CASE,
	"when":      WHEN,
	"while":     WHILE,
	"for":       FOR,
	"in":        IN,
	"break":     BREAK,
	"next":      NEXT,
	"return":    RETURN,
	"panic":     PANIC,
	"rescue":    RESCUE,
	"true":      TRUE,
	"false":     FALSE,
	"nil":       NIL,
	"and":       AND,
	"or":        OR,
	"not":       NOT,
	"as":        AS,
	"defer":     DEFER,
	"do":         DO,
	"class":      CLASS,
	"self":       SELF,
	"interface":  INTERFACE,
	"implements": IMPLEMENTS,
	"any":        ANY,
	"pub":        PUB,
	// Testing keywords
	"describe": DESCRIBE,
	"it":       IT,
	"test":     TEST,
	"table":    TABLE,
	"before":   BEFORE,
	"after":    AFTER,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
