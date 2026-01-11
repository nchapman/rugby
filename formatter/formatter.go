// Package formatter implements code formatting for Rugby source files.
package formatter

import (
	"errors"
	"strings"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/parser"
)

// Formatter formats Rugby source code.
type Formatter struct {
	buf      strings.Builder
	indent   int
	lastLine int

	// Comment tracking for free-floating comments
	emittedLines   map[int]bool    // tracks which comment lines have been emitted
	freeCommentIdx int             // index into raw lexer comments for free-floating
	rawComments    []lexer.Comment // all comments from lexer for position-based fallback
}

// Format parses and formats Rugby source code.
func Format(source string) (string, error) {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return "", errors.New(p.Errors()[0])
	}

	f := &Formatter{
		lastLine:     1,
		emittedLines: make(map[int]bool),
		rawComments:  l.Comments,
	}
	f.formatProgram(program)
	return f.buf.String(), nil
}

// write appends a string to the output buffer.
func (f *Formatter) write(s string) {
	f.buf.WriteString(s)
}

// writeLine writes a string followed by a newline.
func (f *Formatter) writeLine(s string) {
	f.write(s)
	f.write("\n")
}

// writeIndent writes the current indentation.
func (f *Formatter) writeIndent() {
	for range f.indent {
		f.write("  ")
	}
}

// formatCommentGroup formats a comment group (Doc or trailing Comment).
func (f *Formatter) formatCommentGroup(g *ast.CommentGroup) {
	if g == nil || len(g.List) == 0 {
		return
	}
	for _, c := range g.List {
		f.writeIndent()
		f.writeLine(c.Text)
		f.emittedLines[c.Line] = true
	}
}

// formatTrailingComment formats an inline trailing comment.
func (f *Formatter) formatTrailingComment(g *ast.CommentGroup) {
	if g == nil || len(g.List) == 0 {
		return
	}
	// Only emit the first comment for trailing (they should be on same line)
	c := g.List[0]
	f.write("  ")
	f.write(c.Text)
	f.emittedLines[c.Line] = true
}

// emitFreeFloatingComments emits any comments before the given line that haven't been attached to nodes.
func (f *Formatter) emitFreeFloatingComments(beforeLine int) {
	for f.freeCommentIdx < len(f.rawComments) {
		c := f.rawComments[f.freeCommentIdx]
		if c.Line >= beforeLine {
			break
		}
		if f.emittedLines[c.Line] {
			f.freeCommentIdx++
			continue
		}
		// Emit blank line if there's a gap
		if c.Line > f.lastLine+1 {
			f.write("\n")
		}
		f.writeIndent()
		f.writeLine(c.Text)
		f.emittedLines[c.Line] = true
		f.lastLine = c.Line
		f.freeCommentIdx++
	}
}

// emitRemainingComments emits any comments not yet emitted at the end of the file.
func (f *Formatter) emitRemainingComments() {
	for f.freeCommentIdx < len(f.rawComments) {
		c := f.rawComments[f.freeCommentIdx]
		if f.emittedLines[c.Line] {
			f.freeCommentIdx++
			continue
		}
		if c.Line > f.lastLine+1 {
			f.write("\n")
		}
		f.writeIndent()
		f.writeLine(c.Text)
		f.emittedLines[c.Line] = true
		f.lastLine = c.Line
		f.freeCommentIdx++
	}
}

// markCommentGroupEmitted pre-marks all lines in a comment group as emitted.
// This prevents free-floating comment handling from emitting attached comments.
func (f *Formatter) markCommentGroupEmitted(g *ast.CommentGroup) {
	if g == nil {
		return
	}
	for _, c := range g.List {
		f.emittedLines[c.Line] = true
	}
}
