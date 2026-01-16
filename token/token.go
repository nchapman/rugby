// Package token defines the token types for the Rugby lexer.
package token

type TokenType string

const (
	// Special tokens
	EOF     TokenType = "EOF"
	NEWLINE TokenType = "NEWLINE"
	ILLEGAL TokenType = "ILLEGAL"

	// Literals
	IDENT          TokenType = "IDENT"
	STRING         TokenType = "STRING"         // "..." double-quoted string with interpolation
	STRINGLITERAL  TokenType = "STRINGLITERAL"  // '...' single-quoted string (no interpolation)
	HEREDOC        TokenType = "HEREDOC"        // <<END...END multi-line string
	HEREDOCLITERAL TokenType = "HEREDOCLITERAL" // <<'END'...END literal heredoc (no interpolation)
	INT            TokenType = "INT"
	FLOAT          TokenType = "FLOAT"
	SYMBOL         TokenType = "SYMBOL"
	WORDARRAY      TokenType = "WORDARRAY"    // %w{...} word array literal
	INTERPWARRAY   TokenType = "INTERPWARRAY" // %W{...} interpolated word array
	REGEX          TokenType = "REGEX"        // /pattern/flags regex literal

	// Operators
	PLUS       TokenType = "+"
	MINUS      TokenType = "-"
	STAR       TokenType = "*"
	DOUBLESTAR TokenType = "**"
	SLASH      TokenType = "/"
	PERCENT    TokenType = "%"
	BANG       TokenType = "!"

	// Comparison
	EQ       TokenType = "=="
	NE       TokenType = "!="
	LT       TokenType = "<"
	GT       TokenType = ">"
	LE       TokenType = "<="
	GE       TokenType = ">="
	MATCH    TokenType = "=~" // regex match
	NOTMATCH TokenType = "!~" // regex not match

	// Assignment
	ASSIGN      TokenType = "="
	ORASSIGN    TokenType = "||="
	PLUSASSIGN  TokenType = "+="
	MINUSASSIGN TokenType = "-="
	STARASSIGN  TokenType = "*="
	SLASHASSIGN TokenType = "/="

	// Type modifiers and optional operators
	QUESTION         TokenType = "?"
	QUESTIONQUESTION TokenType = "??"
	AMP              TokenType = "&"
	AMPAMP           TokenType = "&&"
	AMPDOT           TokenType = "&."
	PIPEPIPE         TokenType = "||"

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
	ATAT       TokenType = "@@"
	COLON      TokenType = ":"
	COLONCOLON TokenType = "::"
	SHOVELLEFT TokenType = "<<"

	// Keywords
	IMPORT       TokenType = "IMPORT"
	DEF          TokenType = "DEF"
	END          TokenType = "END"
	GETTER       TokenType = "GETTER"
	SETTER       TokenType = "SETTER"
	PROPERTY     TokenType = "PROPERTY"
	SUPER        TokenType = "SUPER"
	GO           TokenType = "GO"
	SPAWN        TokenType = "SPAWN"
	AWAIT        TokenType = "AWAIT"
	CONCURRENTLY TokenType = "CONCURRENTLY"
	MODULE       TokenType = "MODULE"
	INCLUDE      TokenType = "INCLUDE"
	SELECT       TokenType = "SELECT"
	IF           TokenType = "IF"
	ELSIF        TokenType = "ELSIF"
	ELSE         TokenType = "ELSE"
	UNLESS       TokenType = "UNLESS"
	CASE         TokenType = "CASE"
	CASETYPE     TokenType = "CASETYPE"
	WHEN         TokenType = "WHEN"
	WHILE        TokenType = "WHILE"
	UNTIL        TokenType = "UNTIL"
	FOR          TokenType = "FOR"
	LOOP         TokenType = "LOOP"
	IN           TokenType = "IN"
	BREAK        TokenType = "BREAK"
	NEXT         TokenType = "NEXT"
	RETURN       TokenType = "RETURN"
	PANIC        TokenType = "PANIC"
	RESCUE       TokenType = "RESCUE"
	TRUE         TokenType = "TRUE"
	FALSE        TokenType = "FALSE"
	NIL          TokenType = "NIL"
	AND          TokenType = "AND"
	OR           TokenType = "OR"
	NOT          TokenType = "NOT"
	AS           TokenType = "AS"
	LET          TokenType = "LET"
	DEFER        TokenType = "DEFER"
	DO           TokenType = "DO"
	CLASS        TokenType = "CLASS"
	SELF         TokenType = "SELF"
	INTERFACE    TokenType = "INTERFACE"
	IMPLEMENTS   TokenType = "IMPLEMENTS"
	ANY          TokenType = "ANY"
	PUB          TokenType = "PUB"
	PRIVATE      TokenType = "PRIVATE"
	CONST        TokenType = "CONST"
	TYPE         TokenType = "TYPE"
	ENUM         TokenType = "ENUM"
	STRUCT       TokenType = "STRUCT"

	// Testing keywords
	DESCRIBE TokenType = "DESCRIBE"
	IT       TokenType = "IT"
	TEST     TokenType = "TEST"
	TABLE    TokenType = "TABLE"
	BEFORE   TokenType = "BEFORE"
	AFTER    TokenType = "AFTER"
)

type Token struct {
	Type        TokenType
	Literal     string
	Line        int
	Column      int
	SpaceBefore bool // true if whitespace preceded this token
}

var keywords = map[string]TokenType{
	"import":       IMPORT,
	"def":          DEF,
	"end":          END,
	"getter":       GETTER,
	"setter":       SETTER,
	"property":     PROPERTY,
	"super":        SUPER,
	"go":           GO,
	"spawn":        SPAWN,
	"await":        AWAIT,
	"concurrently": CONCURRENTLY,
	"module":       MODULE,
	"include":      INCLUDE,
	"select":       SELECT,
	"if":           IF,
	"elsif":        ELSIF,
	"else":         ELSE,
	"unless":       UNLESS,
	"case":         CASE,
	"case_type":    CASETYPE,
	"when":         WHEN,
	"while":        WHILE,
	"until":        UNTIL,
	"for":          FOR,
	"loop":         LOOP,
	"in":           IN,
	"break":        BREAK,
	"next":         NEXT,
	"return":       RETURN,
	"panic":        PANIC,
	"rescue":       RESCUE,
	"true":         TRUE,
	"false":        FALSE,
	"nil":          NIL,
	"and":          AND,
	"or":           OR,
	"not":          NOT,
	"as":           AS,
	"let":          LET,
	"defer":        DEFER,
	"do":           DO,
	"class":        CLASS,
	"self":         SELF,
	"interface":    INTERFACE,
	"implements":   IMPLEMENTS,
	"any":          ANY,
	"pub":          PUB,
	"private":      PRIVATE,
	"const":        CONST,
	"type":         TYPE,
	"enum":         ENUM,
	"struct":       STRUCT,
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
