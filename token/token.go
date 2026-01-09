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

	// Keywords
	IMPORT TokenType = "IMPORT"
	DEF    TokenType = "DEF"
	END    TokenType = "END"
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
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
