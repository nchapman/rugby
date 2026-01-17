package lexer

import (
	"testing"

	"github.com/nchapman/rugby/token"
)

func TestNextToken(t *testing.T) {
	input := `import fmt

def main
  x = 5 + 3
  puts "hello"
end
`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IMPORT, "import"},
		{token.IDENT, "fmt"},
		{token.NEWLINE, "\n"},
		{token.NEWLINE, "\n"},
		{token.DEF, "def"},
		{token.IDENT, "main"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.PLUS, "+"},
		{token.INT, "3"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "puts"},
		{token.STRING, "hello"},
		{token.NEWLINE, "\n"},
		{token.END, "end"},
		{token.NEWLINE, "\n"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestOperators(t *testing.T) {
	input := `+ - * / % == != < > <= >= = ( ) , ->`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.STAR, "*"},
		{token.SLASH, "/"},
		{token.PERCENT, "%"},
		{token.EQ, "=="},
		{token.NE, "!="},
		{token.LT, "<"},
		{token.GT, ">"},
		{token.LE, "<="},
		{token.GE, ">="},
		{token.ASSIGN, "="},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.COMMA, ","},
		{token.ARROW, "->"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestKeywords(t *testing.T) {
	input := `if elsif else case when while until for in break next return true false and or not end def import as defer do class self`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IF, "if"},
		{token.ELSIF, "elsif"},
		{token.ELSE, "else"},
		{token.CASE, "case"},
		{token.WHEN, "when"},
		{token.WHILE, "while"},
		{token.UNTIL, "until"},
		{token.FOR, "for"},
		{token.IN, "in"},
		{token.BREAK, "break"},
		{token.NEXT, "next"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.FALSE, "false"},
		{token.AND, "and"},
		{token.OR, "or"},
		{token.NOT, "not"},
		{token.END, "end"},
		{token.DEF, "def"},
		{token.IMPORT, "import"},
		{token.AS, "as"},
		{token.DEFER, "defer"},
		{token.DO, "do"},
		{token.CLASS, "class"},
		{token.SELF, "self"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestTestingKeywords(t *testing.T) {
	input := `describe it table before after`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.DESCRIBE, "describe"},
		{token.IT, "it"},
		{token.TABLE, "table"},
		{token.BEFORE, "before"},
		{token.AFTER, "after"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNumbers(t *testing.T) {
	input := `42 3.14 100 0.5 1e10 2.5e-3 1E6 3.14e+2`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INT, "42"},
		{token.FLOAT, "3.14"},
		{token.INT, "100"},
		{token.FLOAT, "0.5"},
		{token.FLOAT, "1e10"},
		{token.FLOAT, "2.5e-3"},
		{token.FLOAT, "1E6"},
		{token.FLOAT, "3.14e+2"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestHexBinaryOctalNumbers(t *testing.T) {
	input := `0x80 0xFF 0xABCD 0b1010 0b11111111 0o755 0o17 0XFF 0B1010 0O755`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INT, "0x80"},
		{token.INT, "0xFF"},
		{token.INT, "0xABCD"},
		{token.INT, "0b1010"},
		{token.INT, "0b11111111"},
		{token.INT, "0o755"},
		{token.INT, "0o17"},
		{token.INT, "0XFF"},   // uppercase X
		{token.INT, "0B1010"}, // uppercase B
		{token.INT, "0O755"},  // uppercase O
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestInvalidNumberLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected token.TokenType
	}{
		{"0x", token.ILLEGAL},  // hex with no digits
		{"0b", token.ILLEGAL},  // binary with no digits
		{"0o", token.ILLEGAL},  // octal with no digits
		{"0xG", token.ILLEGAL}, // hex with invalid digit
		{"0b2", token.ILLEGAL}, // binary with invalid digit
		{"0o8", token.ILLEGAL}, // octal with invalid digit
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != tt.expected {
			t.Errorf("input %q: expected token type %q, got %q (literal: %q)",
				tt.input, tt.expected, tok.Type, tok.Literal)
		}
	}
}

func TestComments(t *testing.T) {
	input := `x = 5 # this is a comment
y = 10`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "y"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestDotToken(t *testing.T) {
	input := `http.Get`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "http"},
		{token.DOT, "."},
		{token.IDENT, "Get"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestImportAliasSyntax(t *testing.T) {
	input := `import encoding/json as json`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IMPORT, "import"},
		{token.IDENT, "encoding"},
		{token.SLASH, "/"},
		{token.IDENT, "json"},
		{token.AS, "as"},
		{token.IDENT, "json"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestDeferSyntax(t *testing.T) {
	input := `defer resp.Body.Close`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.DEFER, "defer"},
		{token.IDENT, "resp"},
		{token.DOT, "."},
		{token.IDENT, "Body"},
		{token.DOT, "."},
		{token.IDENT, "Close"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestBrackets(t *testing.T) {
	input := `[1, 2, 3]`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.COMMA, ","},
		{token.INT, "3"},
		{token.RBRACKET, "]"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestMapSyntax(t *testing.T) {
	input := `{"a" => 1}`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LBRACE, "{"},
		{token.STRING, "a"},
		{token.HASHROCKET, "=>"},
		{token.INT, "1"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestBlockSyntax(t *testing.T) {
	// Block syntax: do |x| ... end
	input := `do |x| end`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.DO, "do"},
		{token.PIPE, "|"},
		{token.IDENT, "x"},
		{token.PIPE, "|"},
		{token.END, "end"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestClassKeyword(t *testing.T) {
	input := `class User end`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.CLASS, "class"},
		{token.IDENT, "User"},
		{token.END, "end"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestInstanceVariableToken(t *testing.T) {
	input := `@name = value`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.AT, "@"},
		{token.IDENT, "name"},
		{token.ASSIGN, "="},
		{token.IDENT, "value"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestColonToken(t *testing.T) {
	input := `x : Int = 5`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "x"},
		{token.COLON, ":"},
		{token.IDENT, "Int"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestTypedParameterTokens(t *testing.T) {
	input := `def add(a : Int, b : Int) -> Int`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.DEF, "def"},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.COLON, ":"},
		{token.IDENT, "Int"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
		{token.COLON, ":"},
		{token.IDENT, "Int"},
		{token.RPAREN, ")"},
		{token.ARROW, "->"},
		{token.IDENT, "Int"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestInterfaceKeyword(t *testing.T) {
	input := `interface Speaker end`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INTERFACE, "interface"},
		{token.IDENT, "Speaker"},
		{token.END, "end"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestPubKeyword(t *testing.T) {
	input := `pub def add(a : Int, b : Int) end`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.PUB, "pub"},
		{token.DEF, "def"},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.COLON, ":"},
		{token.IDENT, "Int"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
		{token.COLON, ":"},
		{token.IDENT, "Int"},
		{token.RPAREN, ")"},
		{token.END, "end"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestPubClassAndInterface(t *testing.T) {
	input := `pub class User end
pub interface Speaker end`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.PUB, "pub"},
		{token.CLASS, "class"},
		{token.IDENT, "User"},
		{token.END, "end"},
		{token.NEWLINE, "\n"},
		{token.PUB, "pub"},
		{token.INTERFACE, "interface"},
		{token.IDENT, "Speaker"},
		{token.END, "end"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestGetLine(t *testing.T) {
	input := "line1\nline2\nline3"
	l := New(input)

	tests := []struct {
		lineNum  int
		expected string
	}{
		{1, "line1"},
		{2, "line2"},
		{3, "line3"},
		{0, ""},  // out of range low
		{4, ""},  // out of range high
		{-1, ""}, // negative
	}

	for _, tt := range tests {
		got := l.GetLine(tt.lineNum)
		if got != tt.expected {
			t.Errorf("GetLine(%d) = %q, want %q", tt.lineNum, got, tt.expected)
		}
	}
}

func TestGetLineEmptyInput(t *testing.T) {
	l := New("")
	if got := l.GetLine(1); got != "" {
		t.Errorf("GetLine(1) on empty input = %q, want empty", got)
	}
}

func TestInput(t *testing.T) {
	input := "def main\n  puts 42\nend"
	l := New(input)
	if got := l.Input(); got != input {
		t.Errorf("Input() = %q, want %q", got, input)
	}
}

func TestRangeTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			typ token.TokenType
			lit string
		}
	}{
		{
			name:  "inclusive range",
			input: `1..10`,
			expected: []struct {
				typ token.TokenType
				lit string
			}{
				{token.INT, "1"},
				{token.DOTDOT, ".."},
				{token.INT, "10"},
				{token.EOF, ""},
			},
		},
		{
			name:  "exclusive range",
			input: `1...10`,
			expected: []struct {
				typ token.TokenType
				lit string
			}{
				{token.INT, "1"},
				{token.TRIPLEDOT, "..."},
				{token.INT, "10"},
				{token.EOF, ""},
			},
		},
		{
			name:  "method call still works",
			input: `obj.method`,
			expected: []struct {
				typ token.TokenType
				lit string
			}{
				{token.IDENT, "obj"},
				{token.DOT, "."},
				{token.IDENT, "method"},
				{token.EOF, ""},
			},
		},
		{
			name:  "float literal still works",
			input: `3.14`,
			expected: []struct {
				typ token.TokenType
				lit string
			}{
				{token.FLOAT, "3.14"},
				{token.EOF, ""},
			},
		},
		{
			name:  "range with variables",
			input: `start..finish`,
			expected: []struct {
				typ token.TokenType
				lit string
			}{
				{token.IDENT, "start"},
				{token.DOTDOT, ".."},
				{token.IDENT, "finish"},
				{token.EOF, ""},
			},
		},
		{
			name:  "for loop with range",
			input: `for i in 0..5`,
			expected: []struct {
				typ token.TokenType
				lit string
			}{
				{token.FOR, "for"},
				{token.IDENT, "i"},
				{token.IN, "in"},
				{token.INT, "0"},
				{token.DOTDOT, ".."},
				{token.INT, "5"},
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
				if tok.Literal != exp.lit {
					t.Errorf("token %d: literal = %q, want %q", i, tok.Literal, exp.lit)
				}
			}
		})
	}
}

func TestOrAssignToken(t *testing.T) {
	input := `x ||= 5`
	l := New(input)

	expected := []struct {
		Type    token.TokenType
		Literal string
	}{
		{token.IDENT, "x"},
		{token.ORASSIGN, "||="},
		{token.INT, "5"},
		{token.EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("token %d: type = %q, want %q", i, tok.Type, exp.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("token %d: literal = %q, want %q", i, tok.Literal, exp.Literal)
		}
	}
}

func TestOrAssignVsBlockPipes(t *testing.T) {
	// Ensure ||= is recognized but || without = is still two PIPE tokens
	tests := []struct {
		name     string
		input    string
		expected []token.TokenType
	}{
		{
			name:     "or assign",
			input:    "x ||= y",
			expected: []token.TokenType{token.IDENT, token.ORASSIGN, token.IDENT, token.EOF},
		},
		{
			name:     "plus assign",
			input:    "x += 5",
			expected: []token.TokenType{token.IDENT, token.PLUSASSIGN, token.INT, token.EOF},
		},
		{
			name:     "minus assign",
			input:    "x -= 5",
			expected: []token.TokenType{token.IDENT, token.MINUSASSIGN, token.INT, token.EOF},
		},
		{
			name:     "star assign",
			input:    "x *= 5",
			expected: []token.TokenType{token.IDENT, token.STARASSIGN, token.INT, token.EOF},
		},
		{
			name:     "slash assign",
			input:    "x /= 5",
			expected: []token.TokenType{token.IDENT, token.SLASHASSIGN, token.INT, token.EOF},
		},
		{
			name:     "block without params",
			input:    "do end",
			expected: []token.TokenType{token.DO, token.END, token.EOF},
		},
		{
			name:     "block with params",
			input:    "do |x| end",
			expected: []token.TokenType{token.DO, token.PIPE, token.IDENT, token.PIPE, token.END, token.EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, expType := range tt.expected {
				tok := l.NextToken()
				if tok.Type != expType {
					t.Errorf("token %d: type = %q, want %q", i, tok.Type, expType)
				}
			}
		})
	}
}

func TestQuestionToken(t *testing.T) {
	// ? suffix is consumed as part of identifier (like method names: empty?)
	// This is correct behavior - "Int?" is a single token
	input := `x : Int?`
	l := New(input)

	expected := []struct {
		Type    token.TokenType
		Literal string
	}{
		{token.IDENT, "x"},
		{token.COLON, ":"},
		{token.IDENT, "Int?"}, // Type with ? suffix is a single identifier
		{token.EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("token %d: type = %q, want %q", i, tok.Type, exp.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("token %d: literal = %q, want %q", i, tok.Literal, exp.Literal)
		}
	}
}

func TestStandaloneQuestionToken(t *testing.T) {
	// Standalone ? (not after an identifier) is a QUESTION token
	input := `?`
	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.QUESTION {
		t.Errorf("expected QUESTION token, got %q", tok.Type)
	}
}

func TestSymbols(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			Type    token.TokenType
			Literal string
		}
	}{
		{
			name:  "basic symbol",
			input: `:ok`,
			expected: []struct {
				Type    token.TokenType
				Literal string
			}{
				{token.SYMBOL, "ok"},
				{token.EOF, ""},
			},
		},
		{
			name:  "symbol with underscores",
			input: `:not_found`,
			expected: []struct {
				Type    token.TokenType
				Literal string
			}{
				{token.SYMBOL, "not_found"},
				{token.EOF, ""},
			},
		},
		{
			name:  "multiple symbols",
			input: `:success :error :pending`,
			expected: []struct {
				Type    token.TokenType
				Literal string
			}{
				{token.SYMBOL, "success"},
				{token.SYMBOL, "error"},
				{token.SYMBOL, "pending"},
				{token.EOF, ""},
			},
		},
		{
			name:  "symbol vs type annotation",
			input: `x : Int`,
			expected: []struct {
				Type    token.TokenType
				Literal string
			}{
				{token.IDENT, "x"},
				{token.COLON, ":"},
				{token.IDENT, "Int"},
				{token.EOF, ""},
			},
		},
		{
			name:  "symbol in expression",
			input: `status = :active`,
			expected: []struct {
				Type    token.TokenType
				Literal string
			}{
				{token.IDENT, "status"},
				{token.ASSIGN, "="},
				{token.SYMBOL, "active"},
				{token.EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.Type {
					t.Errorf("token %d: type = %q, want %q", i, tok.Type, exp.Type)
				}
				if tok.Literal != exp.Literal {
					t.Errorf("token %d: literal = %q, want %q", i, tok.Literal, exp.Literal)
				}
			}
		})
	}
}

func TestCommentCollection(t *testing.T) {
	input := `# First comment
x = 5 # trailing comment
# Another comment
y = 10`

	l := New(input)
	// Consume all tokens
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
	}

	// Check collected comments
	if len(l.Comments) != 3 {
		t.Fatalf("expected 3 comments, got %d", len(l.Comments))
	}

	tests := []struct {
		text   string
		line   int
		column int
	}{
		{"# First comment", 1, 1},
		{"# trailing comment", 2, 7},
		{"# Another comment", 3, 1},
	}

	for i, tt := range tests {
		c := l.Comments[i]
		if c.Text != tt.text {
			t.Errorf("comment %d: text = %q, want %q", i, c.Text, tt.text)
		}
		if c.Line != tt.line {
			t.Errorf("comment %d: line = %d, want %d", i, c.Line, tt.line)
		}
		if c.Column != tt.column {
			t.Errorf("comment %d: column = %d, want %d", i, c.Column, tt.column)
		}
	}
}

func TestCollectCommentsGrouping(t *testing.T) {
	input := `# Group 1 line 1
# Group 1 line 2
x = 5

# Group 2 (after blank line)
y = 10
# Group 3`

	l := New(input)
	// Consume all tokens
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
	}

	groups := l.CollectComments()

	if len(groups) != 3 {
		t.Fatalf("expected 3 comment groups, got %d", len(groups))
	}

	// Group 1 should have 2 comments
	if len(groups[0].List) != 2 {
		t.Errorf("group 0: expected 2 comments, got %d", len(groups[0].List))
	}

	// Group 2 should have 1 comment
	if len(groups[1].List) != 1 {
		t.Errorf("group 1: expected 1 comment, got %d", len(groups[1].List))
	}

	// Group 3 should have 1 comment
	if len(groups[2].List) != 1 {
		t.Errorf("group 2: expected 1 comment, got %d", len(groups[2].List))
	}
}

func TestCommentGroupText(t *testing.T) {
	input := `# Line one
# Line two
x = 5`

	l := New(input)
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
	}

	groups := l.CollectComments()
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}

	text := groups[0].Text()
	expected := "Line one\nLine two"
	if text != expected {
		t.Errorf("Text() = %q, want %q", text, expected)
	}
}

func TestOptionalOperators(t *testing.T) {
	input := `x ?? y
user&.name
if let x = y`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "x"},
		{token.QUESTIONQUESTION, "??"},
		{token.IDENT, "y"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "user"},
		{token.AMPDOT, "&."},
		{token.IDENT, "name"},
		{token.NEWLINE, "\n"},
		{token.IF, "if"},
		{token.LET, "let"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.IDENT, "y"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestSpaceBefore(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []bool // SpaceBefore for each token (excluding EOF)
	}{
		{
			name:  "basic spacing",
			input: "foo bar",
			want:  []bool{false, true}, // foo=false, bar=true
		},
		{
			name:  "no spaces",
			input: "foo-1",
			want:  []bool{false, false, false}, // foo=false, -=false, 1=false
		},
		{
			name:  "spaces around operator",
			input: "foo - 1",
			want:  []bool{false, true, true}, // foo=false, -=true, 1=true
		},
		{
			name:  "space before operator only",
			input: "foo -1",
			want:  []bool{false, true, false}, // foo=false, -=true, 1=false
		},
		{
			name:  "method chain",
			input: "a.b.c",
			want:  []bool{false, false, false, false, false}, // a . b . c
		},
		{
			name:  "method chain with arg",
			input: "a.b x",
			want:  []bool{false, false, false, true}, // a . b x
		},
		{
			name:  "multiple args",
			input: "foo a, b",
			want:  []bool{false, true, false, true}, // foo a , b
		},
		{
			name:  "string arg",
			input: `puts "hello"`,
			want:  []bool{false, true}, // puts "hello"
		},
		{
			name:  "negative number",
			input: "x = -1",
			want:  []bool{false, true, true, false}, // x = - 1
		},
		{
			name:  "bang operator",
			input: "foo!",
			want:  []bool{false, false}, // foo ! (two tokens)
		},
		{
			name:  "bang with space",
			input: "foo !x",
			want:  []bool{false, true, false}, // foo ! x
		},
		{
			name:  "array literal",
			input: "[1, 2]",
			want:  []bool{false, false, false, true, false}, // [ 1 , 2 ]
		},
		{
			name:  "symbol",
			input: "x = :ok",
			want:  []bool{false, true, true}, // x = :ok
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			var got []bool
			for {
				tok := l.NextToken()
				if tok.Type == token.EOF {
					break
				}
				got = append(got, tok.SpaceBefore)
			}
			if len(got) != len(tt.want) {
				t.Errorf("token count = %d, want %d\ngot tokens: %v", len(got), len(tt.want), got)
				return
			}
			for i, want := range tt.want {
				if got[i] != want {
					t.Errorf("token %d SpaceBefore = %v, want %v", i, got[i], want)
				}
			}
		})
	}
}

func TestWordArrayLiteral(t *testing.T) {
	tests := []struct {
		input   string
		tokType token.TokenType
		literal string
	}{
		// Basic word arrays with different delimiters
		{`%w{foo bar baz}`, token.WORDARRAY, "foo\x00bar\x00baz"},
		{`%w(one two three)`, token.WORDARRAY, "one\x00two\x00three"},
		{`%w[a b c]`, token.WORDARRAY, "a\x00b\x00c"},
		{`%w<x y z>`, token.WORDARRAY, "x\x00y\x00z"},
		{`%w|hello world|`, token.WORDARRAY, "hello\x00world"},
		{`%w{}`, token.WORDARRAY, ""},
		{`%W{foo bar}`, token.INTERPWARRAY, "foo\x00bar"},

		// Escape sequences
		{`%w{hello\ world}`, token.WORDARRAY, "hello world"},    // escaped space keeps words together
		{`%w{a\nb}`, token.WORDARRAY, "a\nb"},                   // escaped newline in word
		{`%w{a\tb}`, token.WORDARRAY, "a\tb"},                   // escaped tab in word
		{`%w{a\\b}`, token.WORDARRAY, "a\\b"},                   // escaped backslash
		{`%w{a\}b}`, token.WORDARRAY, "a}b"},                    // escaped closing delimiter
		{`%w(a\)b)`, token.WORDARRAY, "a)b"},                    // escaped closing paren
		{`%w[a\]b]`, token.WORDARRAY, "a]b"},                    // escaped closing bracket
		{`%w<a\>b>`, token.WORDARRAY, "a>b"},                    // escaped closing angle
		{`%w|a\|b|`, token.WORDARRAY, "a|b"},                    // escaped pipe delimiter
		{`%w{foo\ bar baz}`, token.WORDARRAY, "foo bar\x00baz"}, // escaped space creates one word

		// Multiline word arrays
		{"%w{foo\nbar\nbaz}", token.WORDARRAY, "foo\x00bar\x00baz"}, // newlines separate words
		{"%w{foo\tbar}", token.WORDARRAY, "foo\x00bar"},             // tabs separate words
		{"%w{  foo   bar  }", token.WORDARRAY, "foo\x00bar"},        // multiple spaces collapse

		// Interpolation in %W arrays
		{`%W{hello #{name} world}`, token.INTERPWARRAY, "hello\x00#{name}\x00world"},  // interpolation preserved
		{`%W{#{a} #{b}}`, token.INTERPWARRAY, "#{a}\x00#{b}"},                         // multiple interpolations
		{`%W{prefix#{x}suffix}`, token.INTERPWARRAY, "prefix#{x}suffix"},              // interpolation mid-word
		{`%W{#{nested.call()} done}`, token.INTERPWARRAY, "#{nested.call()}\x00done"}, // complex expression
		{`%W{#{a + b}}`, token.INTERPWARRAY, "#{a + b}"},                              // expression with spaces
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != tt.tokType {
			t.Errorf("input %q: expected token type %q, got %q", tt.input, tt.tokType, tok.Type)
		}
		if tok.Literal != tt.literal {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.literal, tok.Literal)
		}
	}
}

func TestDoubleStarToken(t *testing.T) {
	input := `** * *= **`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.DOUBLESTAR, "**"},
		{token.STAR, "*"},
		{token.STARASSIGN, "*="},
		{token.DOUBLESTAR, "**"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestAmpersandTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected []struct {
			tokType token.TokenType
			literal string
		}
	}{
		// Symbol-to-proc: &:method
		{
			input: `&:upcase`,
			expected: []struct {
				tokType token.TokenType
				literal string
			}{
				{token.AMP, "&"},
				{token.SYMBOL, "upcase"},
				{token.EOF, ""},
			},
		},
		// Safe navigation: &.method
		{
			input: `x&.foo`,
			expected: []struct {
				tokType token.TokenType
				literal string
			}{
				{token.IDENT, "x"},
				{token.AMPDOT, "&."},
				{token.IDENT, "foo"},
				{token.EOF, ""},
			},
		},
		// Standalone ampersand
		{
			input: `& x`,
			expected: []struct {
				tokType token.TokenType
				literal string
			}{
				{token.AMP, "&"},
				{token.IDENT, "x"},
				{token.EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		l := New(tt.input)

		for i, exp := range tt.expected {
			tok := l.NextToken()

			if tok.Type != exp.tokType {
				t.Errorf("input %q: tests[%d] - tokentype wrong. expected=%q, got=%q",
					tt.input, i, exp.tokType, tok.Type)
			}

			if tok.Literal != exp.literal {
				t.Errorf("input %q: tests[%d] - literal wrong. expected=%q, got=%q",
					tt.input, i, exp.literal, tok.Literal)
			}
		}
	}
}

func TestHeredoc(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		tokenType token.TokenType
	}{
		{
			name: "basic heredoc",
			input: `x = <<END
hello
world
END`,
			expected:  "hello\nworld\n",
			tokenType: token.HEREDOC,
		},
		{
			name: "heredoc with indented content",
			input: `x = <<END
  hello
    world
END`,
			expected:  "  hello\n    world\n",
			tokenType: token.HEREDOC,
		},
		{
			name: "heredoc with dashes allows indented end",
			input: `x = <<-END
hello
world
  END`,
			expected:  "hello\nworld\n",
			tokenType: token.HEREDOC,
		},
		{
			name: "squiggly heredoc strips leading whitespace",
			input: `x = <<~END
    hello
    world
  END`,
			expected:  "hello\nworld\n",
			tokenType: token.HEREDOC,
		},
		{
			name: "squiggly heredoc with varying indent",
			input: `x = <<~END
    first
      second
    third
  END`,
			expected:  "first\n  second\nthird\n",
			tokenType: token.HEREDOC,
		},
		{
			name: "heredoc with empty lines",
			input: `x = <<END
hello

world
END`,
			expected:  "hello\n\nworld\n",
			tokenType: token.HEREDOC,
		},
		{
			name: "heredoc with underscore delimiter",
			input: `x = <<END_TEXT
content
END_TEXT`,
			expected:  "content\n",
			tokenType: token.HEREDOC,
		},
		{
			name: "empty heredoc",
			input: `x = <<END
END`,
			expected:  "",
			tokenType: token.HEREDOC,
		},
		// Literal heredocs (single-quoted delimiter, no interpolation)
		{
			name: "literal heredoc",
			input: `x = <<'END'
Hello #{name}
END`,
			expected:  "Hello #{name}\n",
			tokenType: token.HEREDOCLITERAL,
		},
		{
			name: "literal heredoc with dash",
			input: `x = <<-'END'
content
  END`,
			expected:  "content\n",
			tokenType: token.HEREDOCLITERAL,
		},
		{
			name: "literal squiggly heredoc",
			input: `x = <<~'END'
    Hello #{name}
    World
  END`,
			expected:  "Hello #{name}\nWorld\n",
			tokenType: token.HEREDOCLITERAL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			// Skip to HEREDOC or HEREDOCLITERAL token
			expectedType := tt.tokenType
			if expectedType == "" {
				expectedType = token.HEREDOC
			}
			for {
				tok := l.NextToken()
				if tok.Type == token.HEREDOC || tok.Type == token.HEREDOCLITERAL {
					if tok.Type != expectedType {
						t.Errorf("wrong token type.\nexpected: %s\ngot: %s",
							expectedType, tok.Type)
					}
					if tok.Literal != tt.expected {
						t.Errorf("heredoc literal wrong.\nexpected:\n%q\ngot:\n%q",
							tt.expected, tok.Literal)
					}
					return
				}
				if tok.Type == token.EOF {
					t.Fatalf("did not find heredoc token")
				}
			}
		})
	}
}

func TestUnterminatedHeredoc(t *testing.T) {
	input := `x = <<END
hello world
no closing delimiter`

	l := New(input)

	// Skip to HEREDOC/ILLEGAL token
	for {
		tok := l.NextToken()
		if tok.Type == token.HEREDOC {
			t.Fatalf("expected ILLEGAL token for unterminated heredoc, got HEREDOC")
		}
		if tok.Type == token.ILLEGAL {
			if tok.Literal != "unterminated heredoc: missing END" {
				t.Errorf("wrong error message: %q", tok.Literal)
			}
			return
		}
		if tok.Type == token.EOF {
			t.Fatalf("did not find ILLEGAL token for unterminated heredoc")
		}
	}
}

func TestHeredocVsShovelLeft(t *testing.T) {
	// Ensure << followed by non-identifier is still SHOVELLEFT for array append
	tests := []struct {
		input    string
		expected []token.TokenType
	}{
		{
			input:    `arr << 1`,
			expected: []token.TokenType{token.IDENT, token.SHOVELLEFT, token.INT, token.EOF},
		},
		{
			input:    `arr << "hello"`,
			expected: []token.TokenType{token.IDENT, token.SHOVELLEFT, token.STRING, token.EOF},
		},
		{
			input:    `arr<<1`,
			expected: []token.TokenType{token.IDENT, token.SHOVELLEFT, token.INT, token.EOF},
		},
	}

	for _, tt := range tests {
		l := New(tt.input)
		for i, exp := range tt.expected {
			tok := l.NextToken()
			if tok.Type != exp {
				t.Errorf("input %q: token[%d] wrong. expected=%q, got=%q",
					tt.input, i, exp, tok.Type)
			}
		}
	}
}

func TestShovelRight(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.TokenType
	}{
		{
			input:    `16 >> 2`,
			expected: []token.TokenType{token.INT, token.SHOVELRIGHT, token.INT, token.EOF},
		},
		{
			input:    `x >> y`,
			expected: []token.TokenType{token.IDENT, token.SHOVELRIGHT, token.IDENT, token.EOF},
		},
		{
			input:    `x>>1`,
			expected: []token.TokenType{token.IDENT, token.SHOVELRIGHT, token.INT, token.EOF},
		},
	}

	for _, tt := range tests {
		l := New(tt.input)
		for i, exp := range tt.expected {
			tok := l.NextToken()
			if tok.Type != exp {
				t.Errorf("input %q: token[%d] wrong. expected=%q, got=%q",
					tt.input, i, exp, tok.Type)
			}
		}
	}
}

func TestSingleQuotedStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`'hello'`, "hello"},
		{`'it\'s'`, "it's"},
		{`'back\\slash'`, "back\\slash"},
		{`'line\nstays'`, "line\\nstays"}, // \n is literal, not escape
		{`''`, ""},                        // empty string
		{`'single "double" inside'`, `single "double" inside`},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != token.STRINGLITERAL {
			t.Errorf("input %q: expected STRINGLITERAL, got %q", tt.input, tok.Type)
			continue
		}
		if tok.Literal != tt.expected {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expected, tok.Literal)
		}
	}
}

func TestStringInterpolationWithNestedQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello #{name}"`, "hello #{name}"},
		{`"value: #{data["key"]}"`, `value: #{data["key"]}`},
		{`"nested: #{map["a"]["b"]}"`, `nested: #{map["a"]["b"]}`},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != token.STRING {
			t.Errorf("input %q: expected STRING, got %q", tt.input, tok.Type)
			continue
		}
		if tok.Literal != tt.expected {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expected, tok.Literal)
		}
	}
}

func TestBlankIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.TokenType
	}{
		{"_", []token.TokenType{token.IDENT}},
		{"_, err", []token.TokenType{token.IDENT, token.COMMA, token.IDENT}},
		{"_foo", []token.TokenType{token.IDENT}},
		{"_, _ = x, y", []token.TokenType{token.IDENT, token.COMMA, token.IDENT, token.ASSIGN, token.IDENT, token.COMMA, token.IDENT}},
	}

	for _, tt := range tests {
		l := New(tt.input)
		for i, exp := range tt.expected {
			tok := l.NextToken()
			if tok.Type != exp {
				t.Errorf("input %q: token[%d] wrong type. expected=%q, got=%q", tt.input, i, exp, tok.Type)
			}
		}
	}

	// Verify standalone _ is correctly identified
	l := New("_")
	tok := l.NextToken()
	if tok.Type != token.IDENT {
		t.Errorf("expected IDENT for '_', got %v", tok.Type)
	}
	if tok.Literal != "_" {
		t.Errorf("expected literal '_', got %q", tok.Literal)
	}
}

func TestMatchOperators(t *testing.T) {
	tests := []struct {
		input           string
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{"=~", token.MATCH, "=~"},
		{"!~", token.NOTMATCH, "!~"},
		// Make sure we don't confuse with other operators
		{"==", token.EQ, "=="},
		{"!=", token.NE, "!="},
		{"=", token.ASSIGN, "="},
		{"!", token.BANG, "!"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected type %q, got %q", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestRegexLiteral(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{
			name:            "simple regex",
			input:           "/hello/",
			expectedType:    token.REGEX,
			expectedLiteral: "hello",
		},
		{
			name:            "regex with flags",
			input:           "/hello/i",
			expectedType:    token.REGEX,
			expectedLiteral: "hello\x00i",
		},
		{
			name:            "regex with multiple flags",
			input:           "/hello/im",
			expectedType:    token.REGEX,
			expectedLiteral: "hello\x00im",
		},
		{
			name:            "regex with escaped slash",
			input:           `/a\/b/`,
			expectedType:    token.REGEX,
			expectedLiteral: `a\/b`,
		},
		{
			name:            "regex with special chars",
			input:           `/\d+/`,
			expectedType:    token.REGEX,
			expectedLiteral: `\d+`,
		},
		{
			name:            "empty regex",
			input:           "//",
			expectedType:    token.REGEX,
			expectedLiteral: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			if tok.Type != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, tok.Type)
			}
			if tok.Literal != tt.expectedLiteral {
				t.Errorf("expected literal %q, got %q", tt.expectedLiteral, tok.Literal)
			}
		})
	}
}

func TestRegexVsDivision(t *testing.T) {
	// Test that / is correctly interpreted based on context
	tests := []struct {
		name     string
		input    string
		expected []token.TokenType
	}{
		{
			name:     "regex at start",
			input:    `/hello/`,
			expected: []token.TokenType{token.REGEX},
		},
		{
			name:     "division after number",
			input:    `10 / 2`,
			expected: []token.TokenType{token.INT, token.SLASH, token.INT},
		},
		{
			name:     "division after ident",
			input:    `a / b`,
			expected: []token.TokenType{token.IDENT, token.SLASH, token.IDENT},
		},
		{
			name:     "regex after assign",
			input:    `x = /hello/`,
			expected: []token.TokenType{token.IDENT, token.ASSIGN, token.REGEX},
		},
		{
			name:     "regex after match op",
			input:    `x =~ /hello/`,
			expected: []token.TokenType{token.IDENT, token.MATCH, token.REGEX},
		},
		{
			name:     "regex after lparen",
			input:    `(/hello/)`,
			expected: []token.TokenType{token.LPAREN, token.REGEX, token.RPAREN},
		},
		{
			name:     "division after rparen",
			input:    `(a) / b`,
			expected: []token.TokenType{token.LPAREN, token.IDENT, token.RPAREN, token.SLASH, token.IDENT},
		},
		{
			name:     "regex after if",
			input:    `if /hello/`,
			expected: []token.TokenType{token.IF, token.REGEX},
		},
		{
			name:     "regex after comma",
			input:    `foo(a, /hello/)`,
			expected: []token.TokenType{token.IDENT, token.LPAREN, token.IDENT, token.COMMA, token.REGEX, token.RPAREN},
		},
		{
			name:     "regex after newline",
			input:    "x\n/hello/",
			expected: []token.TokenType{token.IDENT, token.NEWLINE, token.REGEX},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp {
					t.Errorf("token[%d]: expected %q, got %q", i, exp, tok.Type)
				}
			}
		})
	}
}
