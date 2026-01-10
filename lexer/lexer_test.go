package lexer

import (
	"testing"

	"rugby/token"
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
	input := `if elsif else while for in break next return true false and or not end def import as defer do class self`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IF, "if"},
		{token.ELSIF, "elsif"},
		{token.ELSE, "else"},
		{token.WHILE, "while"},
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

func TestNumbers(t *testing.T) {
	input := `42 3.14 100 0.5`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INT, "42"},
		{token.FLOAT, "3.14"},
		{token.INT, "100"},
		{token.FLOAT, "0.5"},
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
		{token.IDENT, "encoding/json"},
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
	input := `pub def add(a, b) end`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.PUB, "pub"},
		{token.DEF, "def"},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
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
