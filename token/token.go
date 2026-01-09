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

	// Operators
	PLUS    TokenType = "+"
	MINUS   TokenType = "-"
	STAR    TokenType = "*"
	SLASH   TokenType = "/"
	PERCENT TokenType = "%"

	// Comparison
	EQ TokenType = "=="
	NE TokenType = "!="
	LT TokenType = "<"
	GT TokenType = ">"
	LE TokenType = "<="
	GE TokenType = ">="

	// Assignment
	ASSIGN TokenType = "="

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
	PIPE       TokenType = "|"
	AT         TokenType = "@"

	// Keywords
	IMPORT TokenType = "IMPORT"
	DEF    TokenType = "DEF"
	END    TokenType = "END"
	IF     TokenType = "IF"
	ELSIF  TokenType = "ELSIF"
	ELSE   TokenType = "ELSE"
	WHILE  TokenType = "WHILE"
	RETURN TokenType = "RETURN"
	TRUE   TokenType = "TRUE"
	FALSE  TokenType = "FALSE"
	AND    TokenType = "AND"
	OR     TokenType = "OR"
	NOT    TokenType = "NOT"
	AS     TokenType = "AS"
	DEFER TokenType = "DEFER"
	DO    TokenType = "DO"
	CLASS TokenType = "CLASS"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

var keywords = map[string]TokenType{
	"import": IMPORT,
	"def":    DEF,
	"end":    END,
	"if":     IF,
	"elsif":  ELSIF,
	"else":   ELSE,
	"while":  WHILE,
	"return": RETURN,
	"true":   TRUE,
	"false":  FALSE,
	"and":    AND,
	"or":     OR,
	"not":    NOT,
	"as":     AS,
	"defer": DEFER,
	"do":    DO,
	"class": CLASS,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
