// Package lexer implements the tokenizer for Rugby source code.
package lexer

import (
	"strings"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/token"
)

// Comment represents a comment in the source code.
type Comment struct {
	Text   string // the comment text including the leading #
	Line   int    // 1-indexed line number
	Column int    // 0-indexed column position
}

type Lexer struct {
	input         string
	pos           int  // current position in input
	readPos       int  // next position to read
	ch            byte // current char
	line          int
	column        int
	Comments      []Comment       // collected comments during lexing
	prevTokenType token.TokenType // previous token type for context-sensitive lexing
}

// LexerState holds lexer state for save/restore
type LexerState struct {
	pos           int
	readPos       int
	ch            byte
	line          int
	column        int
	commentsLen   int
	prevTokenType token.TokenType
}

// SaveState returns the current lexer state
func (l *Lexer) SaveState() LexerState {
	return LexerState{
		pos:           l.pos,
		readPos:       l.readPos,
		ch:            l.ch,
		line:          l.line,
		column:        l.column,
		commentsLen:   len(l.Comments),
		prevTokenType: l.prevTokenType,
	}
}

// RestoreState restores a previously saved lexer state
func (l *Lexer) RestoreState(s LexerState) {
	l.pos = s.pos
	l.readPos = s.readPos
	l.ch = s.ch
	l.line = s.line
	l.column = s.column
	l.prevTokenType = s.prevTokenType
	// Truncate comments that might have been added during the speculative execution
	if len(l.Comments) > s.commentsLen {
		l.Comments = l.Comments[:s.commentsLen]
	}
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

// GetLine returns the source line at the given line number (1-indexed).
// Returns empty string if line is out of range.
func (l *Lexer) GetLine(lineNum int) string {
	if lineNum < 1 {
		return ""
	}

	currentLine := 1
	start := 0

	for i := range len(l.input) {
		if l.input[i] == '\n' {
			if currentLine == lineNum {
				return l.input[start:i]
			}
			currentLine++
			start = i + 1
		}
	}

	// Handle last line (no trailing newline)
	if currentLine == lineNum {
		return l.input[start:]
	}

	return ""
}

// Input returns the full source input.
func (l *Lexer) Input() string {
	return l.input
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

	spaceBefore := l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '\n':
		tok = l.newToken(token.NEWLINE, string(l.ch))
		tok.SpaceBefore = spaceBefore
		l.line++
		l.column = 0
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.Line = l.line
		tok.Column = l.column
		tok.SpaceBefore = spaceBefore
		l.prevTokenType = tok.Type
		return tok
	case '\'':
		tok.Type = token.STRINGLITERAL
		tok.Literal = l.readSingleQuoteString()
		tok.Line = l.line
		tok.Column = l.column
		tok.SpaceBefore = spaceBefore
		l.prevTokenType = tok.Type
		return tok
	case '#':
		l.readComment()
		return l.NextToken()
	case '+':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.PLUSASSIGN, "+=")
		} else {
			tok = l.newToken(token.PLUS, "+")
		}
	case '-':
		if l.peekChar() == '>' {
			l.readChar()
			tok = l.newToken(token.ARROW, "->")
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.MINUSASSIGN, "-=")
		} else {
			tok = l.newToken(token.MINUS, "-")
		}
	case '*':
		if l.peekChar() == '*' {
			l.readChar()
			tok = l.newToken(token.DOUBLESTAR, "**")
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.STARASSIGN, "*=")
		} else {
			tok = l.newToken(token.STAR, "*")
		}
	case '/':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.SLASHASSIGN, "/=")
		} else if l.isRegexContext() && l.looksLikeRegex() {
			tok = l.readRegex()
			tok.SpaceBefore = spaceBefore
			l.prevTokenType = tok.Type
			return tok
		} else {
			tok = l.newToken(token.SLASH, "/")
		}
	case '%':
		// Check for word array literals: %w{...} or %W{...}
		if l.peekChar() == 'w' || l.peekChar() == 'W' {
			isInterpolated := l.peekChar() == 'W'
			l.readChar() // consume 'w' or 'W'
			l.readChar() // consume delimiter opening
			tok = l.readWordArray(isInterpolated)
			tok.SpaceBefore = spaceBefore
			l.prevTokenType = tok.Type
			return tok
		}
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
		if l.peekChar() == '.' {
			l.readChar()
			if l.peekChar() == '.' {
				l.readChar()
				tok = l.newToken(token.TRIPLEDOT, "...")
			} else {
				tok = l.newToken(token.DOTDOT, "..")
			}
		} else {
			tok = l.newToken(token.DOT, ".")
		}
	case '{':
		tok = l.newToken(token.LBRACE, "{")
	case '}':
		tok = l.newToken(token.RBRACE, "}")
	case '|':
		// Check for ||= (or-assignment) or || (logical or)
		if l.peekChar() == '|' {
			l.readChar() // consume second |
			// Check for ||=
			if l.peekChar() == '=' {
				l.readChar() // consume =
				tok = l.newToken(token.ORASSIGN, "||=")
			} else {
				// It's || (logical or)
				tok = l.newToken(token.PIPEPIPE, "||")
			}
		} else {
			tok = l.newToken(token.PIPE, "|")
		}
	case '@':
		if l.peekChar() == '@' {
			l.readChar()
			tok = l.newToken(token.ATAT, "@@")
		} else {
			tok = l.newToken(token.AT, "@")
		}
	case ':':
		// Check for :: (scope resolution)
		if l.peekChar() == ':' {
			l.readChar()
			tok = l.newToken(token.COLONCOLON, "::")
		} else if isLetter(l.peekChar()) {
			// Check if this is a symbol (:identifier)
			l.readChar() // consume the ':'
			tok.Type = token.SYMBOL
			tok.Literal = l.readIdentifier()
			tok.SpaceBefore = spaceBefore
			l.prevTokenType = tok.Type
			return tok
		} else {
			tok = l.newToken(token.COLON, ":")
		}
	case '?':
		if l.peekChar() == '?' {
			l.readChar()
			tok = l.newToken(token.QUESTIONQUESTION, "??")
		} else {
			tok = l.newToken(token.QUESTION, "?")
		}
	case '&':
		if l.peekChar() == '.' {
			l.readChar()
			tok = l.newToken(token.AMPDOT, "&.")
		} else if l.peekChar() == '&' {
			l.readChar()
			tok = l.newToken(token.AMPAMP, "&&")
		} else {
			tok = l.newToken(token.AMP, "&")
		}
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.EQ, "==")
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = l.newToken(token.HASHROCKET, "=>")
		} else if l.peekChar() == '~' {
			l.readChar()
			tok = l.newToken(token.MATCH, "=~")
		} else {
			tok = l.newToken(token.ASSIGN, "=")
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.NE, "!=")
		} else if l.peekChar() == '~' {
			l.readChar()
			tok = l.newToken(token.NOTMATCH, "!~")
		} else {
			tok = l.newToken(token.BANG, "!")
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.LE, "<=")
		} else if l.peekChar() == '<' {
			l.readChar()
			// Check for heredoc: <<IDENT, <<-IDENT, <<~IDENT, or <<'IDENT'
			nextChar := l.peekChar()
			if isLetter(nextChar) || nextChar == '_' {
				tok = l.readHeredoc(false, false, false)
				l.prevTokenType = tok.Type
				return tok
			} else if nextChar == '\'' {
				// Single-quoted heredoc: <<'DELIM' - no interpolation
				tok = l.readHeredoc(false, false, true)
				l.prevTokenType = tok.Type
				return tok
			} else if nextChar == '-' || nextChar == '~' {
				stripIndent := nextChar == '~'
				l.readChar() // consume - or ~
				if isLetter(l.peekChar()) || l.peekChar() == '_' {
					tok = l.readHeredoc(true, stripIndent, false)
					l.prevTokenType = tok.Type
					return tok
				} else if l.peekChar() == '\'' {
					// <<-'DELIM' or <<~'DELIM'
					tok = l.readHeredoc(true, stripIndent, true)
					l.prevTokenType = tok.Type
					return tok
				}
				// Not a heredoc, was just <<- or <<~ without identifier
				// This is an error case, but we'll let parser handle it
				tok = l.newToken(token.SHOVELLEFT, "<<")
			} else {
				tok = l.newToken(token.SHOVELLEFT, "<<")
			}
		} else {
			tok = l.newToken(token.LT, "<")
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.GE, ">=")
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = l.newToken(token.SHOVELRIGHT, ">>")
		} else {
			tok = l.newToken(token.GT, ">")
		}
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) || l.ch == '_' {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.SpaceBefore = spaceBefore
			l.prevTokenType = tok.Type
			return tok
		}
		if isDigit(l.ch) {
			tok = l.readNumber()
			tok.SpaceBefore = spaceBefore
			l.prevTokenType = tok.Type
			return tok
		}
		tok = l.newToken(token.ILLEGAL, string(l.ch))
	}

	l.readChar()
	tok.SpaceBefore = spaceBefore
	l.prevTokenType = tok.Type
	return tok
}

func (l *Lexer) newToken(tokenType token.TokenType, literal string) token.Token {
	return token.Token{Type: tokenType, Literal: literal, Line: l.line, Column: l.column}
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	// Ruby-style ? suffix for predicate methods (e.g., empty?, valid?)
	// Note: ! is NOT included - it's reserved for the error unwrap operator
	if l.ch == '?' {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readString() string {
	l.readChar() // skip opening quote
	var out []byte
	braceDepth := 0 // track #{...} nesting

	for l.ch != 0 {
		// Only treat " as string terminator when not inside #{}
		if l.ch == '"' && braceDepth == 0 {
			break
		}

		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				out = append(out, '\n')
			case 'r':
				out = append(out, '\r')
			case 't':
				out = append(out, '\t')
			case '"':
				out = append(out, '"')
			case '\\':
				out = append(out, '\\')
			default:
				out = append(out, l.ch)
			}
		} else if l.ch == '#' && l.peekChar() == '{' {
			// Start of interpolation
			out = append(out, l.ch)
			l.readChar()
			out = append(out, l.ch)
			braceDepth++
		} else if l.ch == '{' && braceDepth > 0 {
			// Nested brace inside interpolation
			out = append(out, l.ch)
			braceDepth++
		} else if l.ch == '}' && braceDepth > 0 {
			// End of interpolation or nested brace
			out = append(out, l.ch)
			braceDepth--
		} else {
			out = append(out, l.ch)
		}
		l.readChar()
	}
	l.readChar() // skip closing quote
	return string(out)
}

// readSingleQuoteString reads a single-quoted string literal.
// Single-quoted strings don't support interpolation (like Ruby).
// Only \' and \\ are recognized as escape sequences.
func (l *Lexer) readSingleQuoteString() string {
	l.readChar() // skip opening quote
	var out []byte
	for l.ch != '\'' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case '\'':
				out = append(out, '\'')
			case '\\':
				out = append(out, '\\')
			default:
				// In single-quoted strings, other escapes are literal
				out = append(out, '\\', l.ch)
			}
		} else {
			out = append(out, l.ch)
		}
		l.readChar()
	}
	l.readChar() // skip closing quote
	return string(out)
}

// readHeredoc reads a heredoc string literal.
// allowIndentedEnd: true for <<- and <<~ syntax (closing delimiter can be indented)
// stripIndent: true for <<~ syntax (strips common leading whitespace from content)
// literal: true for <<'DELIM' syntax (no interpolation)
func (l *Lexer) readHeredoc(allowIndentedEnd bool, stripIndent bool, literal bool) token.Token {
	line := l.line
	col := l.column

	// For literal heredocs, skip the opening quote
	if literal {
		l.readChar() // skip opening '
	}

	// Read the delimiter identifier
	l.readChar() // move to start of identifier
	delimStart := l.pos
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	delimiter := l.input[delimStart:l.pos]

	// For literal heredocs, skip the closing quote
	if literal && l.ch == '\'' {
		l.readChar()
	}

	// Skip to end of current line (heredoc content starts on next line)
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}

	// Read heredoc content
	var content strings.Builder
	foundDelimiter := false
	for l.ch != 0 {
		l.readChar() // move past newline to start of new line
		l.line++
		l.column = 1

		// Check if this line starts with (optionally indented) delimiter
		lineStart := l.pos

		// Skip leading whitespace for indented end check
		if allowIndentedEnd {
			for l.ch == ' ' || l.ch == '\t' {
				l.readChar()
			}
		}

		// Check if we've reached the delimiter
		delimMatch := true
		for i := 0; i < len(delimiter) && l.ch != 0; i++ {
			if l.ch != delimiter[i] {
				delimMatch = false
				break
			}
			l.readChar()
		}

		// Verify delimiter is followed by newline or EOF
		if delimMatch && (l.ch == '\n' || l.ch == 0 || l.ch == '\r') {
			foundDelimiter = true
			break
		}

		// Not the delimiter, rewind and read the line as content
		l.pos = lineStart
		l.readPos = lineStart + 1
		if lineStart < len(l.input) {
			l.ch = l.input[lineStart]
		}

		// Read the entire line
		for l.ch != '\n' && l.ch != 0 {
			content.WriteByte(l.ch)
			l.readChar()
		}
		content.WriteByte('\n')
	}

	// Check for unterminated heredoc
	if !foundDelimiter {
		return token.Token{
			Type:    token.ILLEGAL,
			Literal: "unterminated heredoc: missing " + delimiter,
			Line:    line,
			Column:  col,
		}
	}

	result := content.String()

	// Heredocs include trailing newlines per Ruby behavior
	// Each line in the heredoc includes its newline character
	// The line with the closing delimiter does not contribute content

	// Strip common leading whitespace for <<~ heredocs
	if stripIndent && len(result) > 0 {
		result = stripLeadingWhitespace(result)
	}

	// Use HEREDOCLITERAL for single-quoted heredocs (no interpolation)
	tokType := token.HEREDOC
	if literal {
		tokType = token.HEREDOCLITERAL
	}

	return token.Token{
		Type:    tokType,
		Literal: result,
		Line:    line,
		Column:  col,
	}
}

// stripLeadingWhitespace removes common leading whitespace from heredoc content.
func stripLeadingWhitespace(s string) string {
	lines := strings.Split(s, "\n")

	// Find minimum indentation (ignoring empty lines)
	minIndent := -1
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		indent := 0
		for _, ch := range line {
			if ch == ' ' || ch == '\t' {
				indent++
			} else {
				break
			}
		}
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return s
	}

	// Strip the common indentation
	var result strings.Builder
	for i, line := range lines {
		if i > 0 {
			result.WriteByte('\n')
		}
		if len(line) >= minIndent {
			result.WriteString(line[minIndent:])
		} else {
			result.WriteString(line)
		}
	}
	return result.String()
}

func (l *Lexer) readNumber() token.Token {
	pos := l.pos
	startLine, startCol := l.line, l.column
	isFloat := false

	// Handle hex (0x), binary (0b), and octal (0o) literals
	if l.ch == '0' {
		nextCh := l.peekChar()
		if nextCh == 'x' || nextCh == 'X' {
			l.readChar() // consume '0'
			l.readChar() // consume 'x' or 'X'
			if !isHexDigit(l.ch) {
				return token.Token{Type: token.ILLEGAL, Literal: "invalid hex literal: expected digit after 0x", Line: startLine, Column: startCol}
			}
			for isHexDigit(l.ch) {
				l.readChar()
			}
			literal := l.input[pos:l.pos]
			return token.Token{Type: token.INT, Literal: literal, Line: startLine, Column: startCol}
		}
		if nextCh == 'b' || nextCh == 'B' {
			l.readChar() // consume '0'
			l.readChar() // consume 'b' or 'B'
			if l.ch != '0' && l.ch != '1' {
				return token.Token{Type: token.ILLEGAL, Literal: "invalid binary literal: expected 0 or 1 after 0b", Line: startLine, Column: startCol}
			}
			for l.ch == '0' || l.ch == '1' {
				l.readChar()
			}
			literal := l.input[pos:l.pos]
			return token.Token{Type: token.INT, Literal: literal, Line: startLine, Column: startCol}
		}
		if nextCh == 'o' || nextCh == 'O' {
			l.readChar() // consume '0'
			l.readChar() // consume 'o' or 'O'
			if l.ch < '0' || l.ch > '7' {
				return token.Token{Type: token.ILLEGAL, Literal: "invalid octal literal: expected 0-7 after 0o", Line: startLine, Column: startCol}
			}
			for l.ch >= '0' && l.ch <= '7' {
				l.readChar()
			}
			literal := l.input[pos:l.pos]
			return token.Token{Type: token.INT, Literal: literal, Line: startLine, Column: startCol}
		}
	}

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

	// Handle scientific notation: e+00, e-03, E10, etc.
	if l.ch == 'e' || l.ch == 'E' {
		isFloat = true
		l.readChar() // consume 'e' or 'E'

		// Optional sign
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}

		// Exponent digits (required)
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	literal := l.input[pos:l.pos]
	if isFloat {
		return token.Token{Type: token.FLOAT, Literal: literal, Line: startLine, Column: startCol}
	}
	return token.Token{Type: token.INT, Literal: literal, Line: startLine, Column: startCol}
}

// skipWhitespace skips whitespace characters and returns true if any were skipped.
func (l *Lexer) skipWhitespace() bool {
	hadWhitespace := false
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		hadWhitespace = true
		l.readChar()
	}
	return hadWhitespace
}

// readComment reads a comment and stores it in the Comments slice.
// Comments start with # and extend to end of line.
func (l *Lexer) readComment() {
	startCol := l.column
	startLine := l.line
	pos := l.pos
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	l.Comments = append(l.Comments, Comment{
		Text:   l.input[pos:l.pos],
		Line:   startLine,
		Column: startCol,
	})
}

// CollectComments groups consecutive comments into CommentGroups.
// Comments on consecutive lines (no blank lines between) form a single group.
func (l *Lexer) CollectComments() []*ast.CommentGroup {
	if len(l.Comments) == 0 {
		return nil
	}

	var groups []*ast.CommentGroup
	var current []*ast.Comment

	for i, c := range l.Comments {
		astComment := &ast.Comment{
			Text: c.Text,
			Line: c.Line,
			Col:  c.Column,
		}

		if i == 0 {
			current = []*ast.Comment{astComment}
			continue
		}

		// Check if this comment is on the next line after the previous one
		prevLine := l.Comments[i-1].Line
		if c.Line == prevLine+1 {
			// Consecutive - add to current group
			current = append(current, astComment)
		} else {
			// Not consecutive - finish current group and start new one
			groups = append(groups, &ast.CommentGroup{List: current})
			current = []*ast.Comment{astComment}
		}
	}

	// Don't forget the last group
	if len(current) > 0 {
		groups = append(groups, &ast.CommentGroup{List: current})
	}

	return groups
}

// looksLikeRegex returns true if the character after '/' looks like
// the start of a regex pattern (not whitespace, newline, or EOF).
func (l *Lexer) looksLikeRegex() bool {
	next := l.peekChar()
	// If followed by whitespace, newline, or EOF, it's not a regex
	return next != ' ' && next != '\t' && next != '\n' && next != '\r' && next != 0
}

// isRegexContext returns true if the previous token indicates that
// a following '/' should be interpreted as a regex delimiter rather than division.
func (l *Lexer) isRegexContext() bool {
	switch l.prevTokenType {
	// After operators, '/' starts a regex
	case token.PLUS, token.MINUS, token.STAR, token.SLASH, token.PERCENT,
		token.DOUBLESTAR, token.ASSIGN, token.PLUSASSIGN, token.MINUSASSIGN,
		token.STARASSIGN, token.SLASHASSIGN, token.ORASSIGN,
		token.EQ, token.NE, token.LT, token.GT, token.LE, token.GE,
		token.MATCH, token.NOTMATCH,
		token.BANG, token.QUESTION, token.QUESTIONQUESTION,
		token.COMMA, token.COLON, token.HASHROCKET,
		token.ARROW, token.DOTDOT, token.TRIPLEDOT:
		return true
	// After open delimiters, '/' starts a regex
	case token.LPAREN, token.LBRACKET, token.LBRACE, token.PIPE:
		return true
	// After certain keywords, '/' starts a regex
	// Note: IN is deliberately excluded because import paths like "gopkg.in/yaml"
	// would incorrectly trigger regex parsing.
	case token.IF, token.ELSIF, token.ELSE, token.UNLESS,
		token.WHILE, token.UNTIL, token.WHEN, token.RETURN,
		token.CASE, token.AND, token.OR, token.NOT,
		token.DO:
		return true
	// At the start of a statement (after newline or at beginning)
	case token.NEWLINE, "":
		return true
	default:
		return false
	}
}

// readRegex reads a regex literal: /pattern/flags
// Returns the token (REGEX on success, or falls back to SLASH).
func (l *Lexer) readRegex() token.Token {
	startLine := l.line
	startCol := l.column

	l.readChar() // consume opening '/'

	var pattern strings.Builder
	for l.ch != '/' && l.ch != 0 && l.ch != '\n' {
		if l.ch == '\\' {
			// Escape sequence - include both backslash and next char
			pattern.WriteByte(l.ch)
			l.readChar()
			if l.ch != 0 && l.ch != '\n' {
				pattern.WriteByte(l.ch)
				l.readChar()
			}
		} else {
			pattern.WriteByte(l.ch)
			l.readChar()
		}
	}

	// Must end with '/'
	if l.ch != '/' {
		// Not a valid regex, but we've consumed characters.
		// Return as ILLEGAL or handle differently.
		return token.Token{
			Type:    token.ILLEGAL,
			Literal: "unterminated regex literal",
			Line:    startLine,
			Column:  startCol,
		}
	}

	l.readChar() // consume closing '/'

	// Read optional flags: i (case-insensitive), m (multiline), s (single-line), x (extended)
	var flags strings.Builder
	for l.ch == 'i' || l.ch == 'm' || l.ch == 's' || l.ch == 'x' {
		flags.WriteByte(l.ch)
		l.readChar()
	}

	// Combine pattern and flags in the literal
	// Format: pattern + "\x00" + flags (null separator)
	literal := pattern.String()
	if flags.Len() > 0 {
		literal += "\x00" + flags.String()
	}

	return token.Token{
		Type:    token.REGEX,
		Literal: literal,
		Line:    startLine,
		Column:  startCol,
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}

// getClosingDelimiter returns the closing delimiter for paired delimiters,
// or the same character for non-paired delimiters.
func getClosingDelimiter(opening byte) byte {
	switch opening {
	case '(':
		return ')'
	case '[':
		return ']'
	case '{':
		return '}'
	case '<':
		return '>'
	default:
		return opening // for |, /, !, etc.
	}
}

// readWordArray reads a %w{...} or %W{...} word array literal.
// Words are separated by whitespace. The literal is stored with \x00 as separator.
// isInterpolated indicates %W (true) vs %w (false).
func (l *Lexer) readWordArray(isInterpolated bool) token.Token {
	opening := l.ch
	closing := getClosingDelimiter(opening)
	line := l.line
	col := l.column

	var words []string
	var currentWord []byte

	l.readChar() // move past opening delimiter

	for l.ch != closing && l.ch != 0 {
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}

		// Handle escape sequences
		if l.ch == '\\' {
			l.readChar()
			// Handle closing delimiter escape first (dynamic value)
			if l.ch == closing {
				currentWord = append(currentWord, closing)
			} else {
				// Handle standard escape sequences
				switch l.ch {
				case ' ':
					currentWord = append(currentWord, ' ')
				case 'n':
					currentWord = append(currentWord, '\n')
				case 't':
					currentWord = append(currentWord, '\t')
				case '\\':
					currentWord = append(currentWord, '\\')
				default:
					// Keep the backslash and the character
					currentWord = append(currentWord, '\\', l.ch)
				}
			}
			l.readChar()
			continue
		}

		// Handle #{...} interpolation in %W arrays - don't treat inner } as closing
		if isInterpolated && l.ch == '#' && l.peekChar() == '{' {
			currentWord = append(currentWord, '#')
			l.readChar() // consume #
			currentWord = append(currentWord, '{')
			l.readChar() // consume {
			braceCount := 1
			for braceCount > 0 && l.ch != 0 {
				switch l.ch {
				case '{':
					braceCount++
				case '}':
					braceCount--
				}
				currentWord = append(currentWord, l.ch)
				l.readChar()
			}
			continue
		}

		// Whitespace separates words
		if l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			if len(currentWord) > 0 {
				words = append(words, string(currentWord))
				currentWord = nil
			}
			l.readChar()
			continue
		}

		currentWord = append(currentWord, l.ch)
		l.readChar()
	}

	// Don't forget the last word
	if len(currentWord) > 0 {
		words = append(words, string(currentWord))
	}

	// Consume closing delimiter
	if l.ch == closing {
		l.readChar()
	}

	// Join words with null byte separator
	literal := strings.Join(words, "\x00")

	tokType := token.WORDARRAY
	if isInterpolated {
		tokType = token.INTERPWARRAY
	}

	return token.Token{
		Type:    tokType,
		Literal: literal,
		Line:    line,
		Column:  col,
	}
}
