// Package errors provides rich error formatting with source context for Rugby.
package errors

import "strings"

// Source holds source code for error rendering.
type Source struct {
	Filename string
	Content  string
	lines    []string // cached split lines
}

// NewSource creates a Source from filename and content.
func NewSource(filename, content string) *Source {
	return &Source{
		Filename: filename,
		Content:  content,
	}
}

// Line returns the 1-indexed line from the source.
// Returns empty string if out of range.
func (s *Source) Line(num int) string {
	if s == nil || num < 1 {
		return ""
	}
	s.ensureLines()
	if num > len(s.lines) {
		return ""
	}
	return s.lines[num-1]
}

// LineCount returns the total number of lines.
func (s *Source) LineCount() int {
	if s == nil {
		return 0
	}
	s.ensureLines()
	return len(s.lines)
}

// ensureLines lazily splits the content into lines.
func (s *Source) ensureLines() {
	if s.lines == nil {
		s.lines = strings.Split(s.Content, "\n")
	}
}

// Span represents a location in source code.
type Span struct {
	Line      int // 1-indexed line number
	Column    int // 1-indexed column number
	EndLine   int // 1-indexed end line (0 means same as Line)
	EndColumn int // 1-indexed end column (0 means end of token)
}

// IsSingleLine returns true if the span is on a single line.
func (s Span) IsSingleLine() bool {
	return s.EndLine == 0 || s.EndLine == s.Line
}

// PositionedError is an error with source location information.
type PositionedError interface {
	error
	Position() (line, column int)
}

// ErrorWithHelp is an error that can provide help suggestions.
type ErrorWithHelp interface {
	error
	Help() string
}

// ErrorWithNotes is an error that can provide additional notes.
type ErrorWithNotes interface {
	error
	Notes() []Note
}

// Note represents an additional piece of information for an error.
type Note struct {
	Message string
	Span    Span
}
