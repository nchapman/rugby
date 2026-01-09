package lexer

import (
	"rugby/token"
)

type Lexer struct {
	input   string
	pos     int  // current position in input
	readPos int  // next position to read
	ch      byte // current char
	line    int
	column  int
}

// LexerState holds lexer state for save/restore
type LexerState struct {
	pos     int
	readPos int
	ch      byte
	line    int
	column  int
}

// SaveState returns the current lexer state
func (l *Lexer) SaveState() LexerState {
	return LexerState{
		pos:     l.pos,
		readPos: l.readPos,
		ch:      l.ch,
		line:    l.line,
		column:  l.column,
	}
}

// RestoreState restores a previously saved lexer state
func (l *Lexer) RestoreState(s LexerState) {
	l.pos = s.pos
	l.readPos = s.readPos
	l.ch = s.ch
	l.line = s.line
	l.column = s.column
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	l.column++
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '\n':
		tok = l.newToken(token.NEWLINE, string(l.ch))
		l.line++
		l.column = 0
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.Line = l.line
		tok.Column = l.column
		return tok
	case '#':
		l.skipComment()
		return l.NextToken()
	case '+':
		tok = l.newToken(token.PLUS, "+")
	case '-':
		if l.peekChar() == '>' {
			l.readChar()
			tok = l.newToken(token.ARROW, "->")
		} else {
			tok = l.newToken(token.MINUS, "-")
		}
	case '*':
		tok = l.newToken(token.STAR, "*")
	case '/':
		tok = l.newToken(token.SLASH, "/")
	case '%':
		tok = l.newToken(token.PERCENT, "%")
	case '(':
		tok = l.newToken(token.LPAREN, "(")
	case ')':
		tok = l.newToken(token.RPAREN, ")")
	case '[':
		tok = l.newToken(token.LBRACKET, "[")
	case ']':
		tok = l.newToken(token.RBRACKET, "]")
	case ',':
		tok = l.newToken(token.COMMA, ",")
	case '.':
		tok = l.newToken(token.DOT, ".")
	case '{':
		tok = l.newToken(token.LBRACE, "{")
	case '}':
		tok = l.newToken(token.RBRACE, "}")
	case '|':
		tok = l.newToken(token.PIPE, "|")
	case '@':
		tok = l.newToken(token.AT, "@")
	case ':':
		tok = l.newToken(token.COLON, ":")
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.EQ, "==")
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = l.newToken(token.HASHROCKET, "=>")
		} else {
			tok = l.newToken(token.ASSIGN, "=")
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.NE, "!=")
		} else {
			tok = l.newToken(token.NOT, "!")
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.LE, "<=")
		} else {
			tok = l.newToken(token.LT, "<")
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.GE, ">=")
		} else {
			tok = l.newToken(token.GT, ">")
		}
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		}
		if isDigit(l.ch) {
			return l.readNumber()
		}
		tok = l.newToken(token.ILLEGAL, string(l.ch))
	}

	l.readChar()
	return tok
}

func (l *Lexer) newToken(tokenType token.TokenType, literal string) token.Token {
	return token.Token{Type: tokenType, Literal: literal, Line: l.line, Column: l.column}
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '/' {
		l.readChar()
	}
	// Ruby-style method suffixes: ? for predicates, ! for mutating methods
	if l.ch == '?' || l.ch == '!' {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readString() string {
	l.readChar() // skip opening quote
	pos := l.pos
	for l.ch != '"' && l.ch != 0 {
		l.readChar()
	}
	str := l.input[pos:l.pos]
	l.readChar() // skip closing quote
	return str
}

func (l *Lexer) readNumber() token.Token {
	pos := l.pos
	isFloat := false

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	literal := l.input[pos:l.pos]
	if isFloat {
		return token.Token{Type: token.FLOAT, Literal: literal, Line: l.line, Column: l.column}
	}
	return token.Token{Type: token.INT, Literal: literal, Line: l.line, Column: l.column}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
