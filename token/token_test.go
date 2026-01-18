package token

import "testing"

func TestLookupIdent_Keywords(t *testing.T) {
	tests := []struct {
		ident    string
		expected TokenType
	}{
		// Core keywords
		{"import", IMPORT},
		{"def", DEF},
		{"end", END},
		{"class", CLASS},
		{"module", MODULE},
		{"interface", INTERFACE},
		{"implements", IMPLEMENTS},

		// Control flow
		{"if", IF},
		{"elsif", ELSIF},
		{"else", ELSE},
		{"unless", UNLESS},
		{"case", CASE},
		{"case_type", CASETYPE},
		{"when", WHEN},
		{"while", WHILE},
		{"until", UNTIL},
		{"for", FOR},
		{"in", IN},

		// Jump statements
		{"break", BREAK},
		{"next", NEXT},
		{"continue", NEXT}, // alias for next
		{"return", RETURN},
		{"panic", PANIC},
		{"rescue", RESCUE},

		// Literals
		{"true", TRUE},
		{"false", FALSE},
		{"nil", NIL},

		// Logical operators
		{"and", AND},
		{"or", OR},
		{"not", NOT},

		// Type system
		{"as", AS},
		{"let", LET},
		{"any", ANY},

		// Accessors
		{"getter", GETTER},
		{"setter", SETTER},
		{"property", PROPERTY},

		// OOP
		{"self", SELF},
		{"super", SUPER},
		{"include", INCLUDE},
		{"pub", PUB},

		// Concurrency
		{"go", GO},
		{"spawn", SPAWN},
		{"await", AWAIT},
		{"concurrently", CONCURRENTLY},
		{"select", SELECT},
		{"defer", DEFER},

		// Blocks
		{"do", DO},

		// Testing keywords
		{"describe", DESCRIBE},
		{"it", IT},
		{"test", TEST},
		{"table", TABLE},
		{"before", BEFORE},
		{"after", AFTER},
	}

	for _, tt := range tests {
		t.Run(tt.ident, func(t *testing.T) {
			got := LookupIdent(tt.ident)
			if got != tt.expected {
				t.Errorf("LookupIdent(%q) = %v, want %v", tt.ident, got, tt.expected)
			}
		})
	}
}

func TestLookupIdent_Identifiers(t *testing.T) {
	tests := []string{
		"foo",
		"bar",
		"my_variable",
		"CamelCase",
		"snake_case",
		"user123",
		"valid?",
		"process!",
		"x",
		"_private",
		"IF",      // uppercase - not a keyword
		"Class",   // capitalized - not a keyword
		"RETURN",  // uppercase - not a keyword
		"defn",    // similar to keyword but not
		"iffy",    // contains keyword but not
		"endgame", // contains keyword but not
	}

	for _, ident := range tests {
		t.Run(ident, func(t *testing.T) {
			got := LookupIdent(ident)
			if got != IDENT {
				t.Errorf("LookupIdent(%q) = %v, want IDENT", ident, got)
			}
		})
	}
}

func TestLookupIdent_EmptyString(t *testing.T) {
	got := LookupIdent("")
	if got != IDENT {
		t.Errorf("LookupIdent(\"\") = %v, want IDENT", got)
	}
}

func TestToken_Fields(t *testing.T) {
	tok := Token{
		Type:        DEF,
		Literal:     "def",
		Line:        10,
		Column:      5,
		SpaceBefore: true,
	}

	if tok.Type != DEF {
		t.Errorf("tok.Type = %v, want DEF", tok.Type)
	}
	if tok.Literal != "def" {
		t.Errorf("tok.Literal = %v, want \"def\"", tok.Literal)
	}
	if tok.Line != 10 {
		t.Errorf("tok.Line = %v, want 10", tok.Line)
	}
	if tok.Column != 5 {
		t.Errorf("tok.Column = %v, want 5", tok.Column)
	}
	if !tok.SpaceBefore {
		t.Error("tok.SpaceBefore = false, want true")
	}
}

func TestToken_ZeroValue(t *testing.T) {
	var tok Token

	if tok.Type != "" {
		t.Errorf("zero Token.Type = %v, want empty string", tok.Type)
	}
	if tok.Literal != "" {
		t.Errorf("zero Token.Literal = %v, want empty string", tok.Literal)
	}
	if tok.Line != 0 {
		t.Errorf("zero Token.Line = %v, want 0", tok.Line)
	}
	if tok.Column != 0 {
		t.Errorf("zero Token.Column = %v, want 0", tok.Column)
	}
	if tok.SpaceBefore {
		t.Error("zero Token.SpaceBefore = true, want false")
	}
}

func TestTokenType_StringValues(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  string
	}{
		// Special tokens
		{EOF, "EOF"},
		{NEWLINE, "NEWLINE"},
		{ILLEGAL, "ILLEGAL"},

		// Literals
		{IDENT, "IDENT"},
		{STRING, "STRING"},
		{HEREDOC, "HEREDOC"},
		{INT, "INT"},
		{FLOAT, "FLOAT"},
		{SYMBOL, "SYMBOL"},
		{WORDARRAY, "WORDARRAY"},
		{INTERPWARRAY, "INTERPWARRAY"},

		// Operators
		{PLUS, "+"},
		{MINUS, "-"},
		{STAR, "*"},
		{DOUBLESTAR, "**"},
		{SLASH, "/"},
		{PERCENT, "%"},
		{BANG, "!"},

		// Comparison
		{EQ, "=="},
		{NE, "!="},
		{LT, "<"},
		{GT, ">"},
		{LE, "<="},
		{GE, ">="},

		// Assignment
		{ASSIGN, "="},
		{ORASSIGN, "||="},
		{PLUSASSIGN, "+="},
		{MINUSASSIGN, "-="},
		{STARASSIGN, "*="},
		{SLASHASSIGN, "/="},

		// Optional operators
		{QUESTION, "?"},
		{QUESTIONQUESTION, "??"},
		{AMP, "&"},
		{AMPDOT, "&."},
		{CARET, "^"},

		// Delimiters
		{LPAREN, "("},
		{RPAREN, ")"},
		{LBRACKET, "["},
		{RBRACKET, "]"},
		{LBRACE, "{"},
		{RBRACE, "}"},
		{COMMA, ","},
		{ARROW, "->"},
		{HASHROCKET, "=>"},
		{DOT, "."},
		{DOTDOT, ".."},
		{TRIPLEDOT, "..."},
		{PIPE, "|"},
		{AT, "@"},
		{COLON, ":"},
		{SHOVELLEFT, "<<"},
	}

	for _, tt := range tests {
		t.Run(string(tt.tokenType), func(t *testing.T) {
			if string(tt.tokenType) != tt.expected {
				t.Errorf("TokenType %v string = %q, want %q", tt.tokenType, string(tt.tokenType), tt.expected)
			}
		})
	}
}

func TestTokenType_UniqueValues(t *testing.T) {
	allTypes := []TokenType{
		EOF, NEWLINE, ILLEGAL,
		IDENT, STRING, HEREDOC, INT, FLOAT, SYMBOL, WORDARRAY, INTERPWARRAY,
		PLUS, MINUS, STAR, DOUBLESTAR, SLASH, PERCENT, BANG,
		EQ, NE, LT, GT, LE, GE,
		ASSIGN, ORASSIGN, PLUSASSIGN, MINUSASSIGN, STARASSIGN, SLASHASSIGN,
		QUESTION, QUESTIONQUESTION, AMP, AMPDOT,
		LPAREN, RPAREN, LBRACKET, RBRACKET, LBRACE, RBRACE,
		COMMA, ARROW, HASHROCKET, DOT, DOTDOT, TRIPLEDOT, PIPE, AT, COLON, SHOVELLEFT,
		IMPORT, DEF, END, GETTER, SETTER, PROPERTY, SUPER,
		GO, SPAWN, AWAIT, CONCURRENTLY, MODULE, INCLUDE, SELECT,
		IF, ELSIF, ELSE, UNLESS, CASE, CASETYPE, WHEN,
		WHILE, UNTIL, FOR, IN, BREAK, NEXT, RETURN, PANIC, RESCUE,
		TRUE, FALSE, NIL, AND, OR, NOT, AS, LET, DEFER, DO,
		CLASS, SELF, INTERFACE, IMPLEMENTS, ANY, PUB,
		DESCRIBE, IT, TEST, TABLE, BEFORE, AFTER,
	}

	seen := make(map[TokenType]bool)
	for _, tt := range allTypes {
		if seen[tt] {
			t.Errorf("duplicate TokenType: %v", tt)
		}
		seen[tt] = true
	}
}

func TestKeywordsMap_Complete(t *testing.T) {
	// Test all keywords via LookupIdent - this ensures every keyword
	// in the package's keywords map is verified without duplication
	expectedKeywords := map[string]TokenType{
		// Core definitions
		"import":    IMPORT,
		"def":       DEF,
		"end":       END,
		"class":     CLASS,
		"module":    MODULE,
		"interface": INTERFACE,
		"struct":    STRUCT,
		"enum":      ENUM,
		"type":      TYPE,
		"const":     CONST,

		// Accessors
		"getter":   GETTER,
		"setter":   SETTER,
		"property": PROPERTY,

		// Visibility
		"pub":     PUB,
		"private": PRIVATE,

		// OOP
		"self":       SELF,
		"super":      SUPER,
		"implements": IMPLEMENTS,
		"include":    INCLUDE,
		"any":        ANY,

		// Control flow
		"if":        IF,
		"elsif":     ELSIF,
		"else":      ELSE,
		"unless":    UNLESS,
		"case":      CASE,
		"case_type": CASETYPE,
		"when":      WHEN,
		"while":     WHILE,
		"until":     UNTIL,
		"for":       FOR,
		"loop":      LOOP,
		"in":        IN,

		// Jump statements
		"break":    BREAK,
		"next":     NEXT,
		"continue": NEXT, // alias for next
		"return":   RETURN,
		"panic":    PANIC,
		"rescue":   RESCUE,

		// Literals
		"true":  TRUE,
		"false": FALSE,
		"nil":   NIL,

		// Logical operators
		"and": AND,
		"or":  OR,
		"not": NOT,

		// Type system
		"as":  AS,
		"let": LET,

		// Concurrency
		"go":           GO,
		"spawn":        SPAWN,
		"await":        AWAIT,
		"concurrently": CONCURRENTLY,
		"select":       SELECT,
		"defer":        DEFER,

		// Blocks
		"do": DO,

		// Testing keywords
		"describe": DESCRIBE,
		"it":       IT,
		"test":     TEST,
		"table":    TABLE,
		"before":   BEFORE,
		"after":    AFTER,
	}

	// Check all expected keywords are in the map
	for kw, expected := range expectedKeywords {
		got := LookupIdent(kw)
		if got != expected {
			t.Errorf("keyword %q: got %v, want %v", kw, got, expected)
		}
	}
}
